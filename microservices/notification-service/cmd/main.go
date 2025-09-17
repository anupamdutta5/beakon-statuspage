// Package main is the entry point for the Notification Service.
// This is the modernized version using the shared-resilience module.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anupamdutta5/statuspage-shared-resilience"
	"github.com/anupamdutta5/statuspage-notification-service/internal/handlers"
	"github.com/anupamdutta5/statuspage-notification-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load configuration from environment variables
	resilienceConfig := resilience.LoadConfigFromEnv()
	if err := resilienceConfig.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Initialize logger based on environment
	var logger *zap.Logger
	var err error

	if resilienceConfig.Environment == "production" {
		logger, err = zap.NewProduction()
		gin.SetMode(gin.ReleaseMode)
	} else {
		logger, err = zap.NewDevelopment()
		gin.SetMode(gin.DebugMode)
	}

	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Notification Service",
		zap.String("service", "notification-service"),
		zap.String("version", "1.0.0"),
		zap.String("environment", resilienceConfig.Environment),
		zap.Int("port", resilienceConfig.Server.Port),
	)

	// Initialize database manager with connection pooling and health checks
	dbManager, err := resilience.NewDatabaseManager(resilienceConfig.Database, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database manager", zap.Error(err))
	}
	defer dbManager.Close()

	// Create Gin router
	router := gin.New()

	// Add comprehensive middleware stack
	middleware := resilience.DefaultMiddlewareStack(resilienceConfig, logger)
	for _, mw := range middleware {
		router.Use(mw)
	}

	// Initialize notification service and handlers
	notificationService := services.NewNotificationService(dbManager.GetDB(), logger)
	enhancedNotificationService := services.NewEnhancedNotificationService(dbManager.GetDB(), logger)
	notificationHandler := handlers.NewNotificationHandler(notificationService, logger)
	enhancedNotificationHandler := handlers.NewEnhancedNotificationHandler(enhancedNotificationService, logger)

	// Setup basic routes
	router.GET("/health", enhancedNotificationHandler.Health)

	// API routes
	api := router.Group("/api/v1")
	// TODO: Add authentication middleware here
	{
		// Legacy notification endpoints
		api.GET("/notifications", notificationHandler.GetNotifications)

		// Enhanced notification endpoints
		api.POST("/notifications/send", enhancedNotificationHandler.SendNotification)
		api.POST("/notifications/maintenance", enhancedNotificationHandler.SendMaintenanceNotification)
		api.POST("/notifications/incident", enhancedNotificationHandler.SendIncidentNotification)

		// Provider management
		api.GET("/providers", enhancedNotificationHandler.GetProviders)
		api.POST("/providers/configure", enhancedNotificationHandler.ConfigureProvider)
		api.POST("/providers/test", enhancedNotificationHandler.TestProvider)

		// Channel management
		api.GET("/channels", enhancedNotificationHandler.GetChannels)
		api.POST("/channels", enhancedNotificationHandler.CreateChannel)

		// Template management
		api.GET("/templates", enhancedNotificationHandler.GetTemplates)

		// Subscription management
		api.GET("/subscriptions", enhancedNotificationHandler.GetSubscriptions)

		// Legacy webhook endpoint
		api.POST("/webhook", notificationHandler.HandleWebhook)
	}

	// Create HTTP server with proper timeouts and configuration
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", resilienceConfig.Server.Host, resilienceConfig.Server.Port),
		Handler:      router,
		ReadTimeout:  resilienceConfig.Server.ReadTimeout,
		WriteTimeout: resilienceConfig.Server.WriteTimeout,
		IdleTimeout:  resilienceConfig.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Notification Service server starting",
			zap.String("addr", server.Addr),
			zap.Duration("read_timeout", resilienceConfig.Server.ReadTimeout),
			zap.Duration("write_timeout", resilienceConfig.Server.WriteTimeout),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Setup health check monitoring
	if resilienceConfig.Monitoring.Enabled {
		dbHealthChecker := resilience.NewDatabaseHealthChecker(dbManager)

		// Periodically log health status
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					healthCtx, healthCancel := context.WithTimeout(context.Background(), 5*time.Second)
					health := dbHealthChecker.Check(healthCtx)
					healthCancel()

					if status, ok := health["status"].(string); ok && status != "healthy" {
						logger.Warn("Database health check failed", zap.Any("health", health))
					}
				}
			}
		}()
	}

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Notification Service server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), resilienceConfig.Server.GracefulStop)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	// Close database connections
	if err := dbManager.Close(); err != nil {
		logger.Error("Failed to close database connections", zap.Error(err))
	}

	logger.Info("Notification Service server exited gracefully")
}