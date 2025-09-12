// Package main is the entry point for the Notification Service.
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

	"github.com/enterprise-status/statuspage-notification-service/internal/config"
	"github.com/enterprise-status/statuspage-notification-service/internal/handlers"
	"github.com/enterprise-status/statuspage-notification-service/internal/middleware"
	"github.com/enterprise-status/statuspage-notification-service/internal/services"
	"github.com/enterprise-status/statuspage-notification-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := logger.New(cfg.Server.Environment)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Notification Service",
		zap.String("service", cfg.Service.Name),
		zap.String("version", cfg.Service.Version),
		zap.Int("port", cfg.Server.Port),
		zap.String("environment", cfg.Server.Environment))

	// Set Gin mode
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db, err := services.InitDatabase(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Initialize services
	notificationService := services.NewNotificationService(db, logger.Logger)

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger(logger.Logger))
	router.Use(middleware.Recovery(logger.Logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// Initialize handlers
	notificationHandler := handlers.NewNotificationHandler(notificationService, logger.Logger)

	// Setup routes
	setupRoutes(router, notificationHandler, cfg)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Notification Service server starting", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Notification Service server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Notification Service server exited")
}

// setupRoutes configures all the routes for the Notification Service.
func setupRoutes(router *gin.Engine, handler *handlers.NotificationHandler, cfg *config.Config) {
	// Health check endpoint
	router.GET("/health", handler.Health)

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		public := api.Group("/public")
		{
			public.POST("/webhook", handler.HandleWebhook)
		}

		// Protected routes (authentication required)
		protected := api.Group("/")
		protected.Use(middleware.Auth(cfg.JWT.Secret))
		{
			// Notification management
			notifications := protected.Group("/notifications")
			{
				notifications.GET("", handler.GetNotifications)
				notifications.POST("", handler.CreateNotification)
				notifications.GET("/:id", handler.GetNotification)
				notifications.PUT("/:id", handler.UpdateNotification)
				notifications.DELETE("/:id", handler.DeleteNotification)
				notifications.POST("/:id/send", handler.SendNotification)
				notifications.GET("/:id/status", handler.GetNotificationStatus)
			}

			// Template management
			templates := protected.Group("/templates")
			{
				templates.GET("", handler.GetTemplates)
				templates.POST("", handler.CreateTemplate)
				templates.GET("/:id", handler.GetTemplate)
				templates.PUT("/:id", handler.UpdateTemplate)
				templates.DELETE("/:id", handler.DeleteTemplate)
				templates.POST("/:id/test", handler.TestTemplate)
			}

			// Channel management
			channels := protected.Group("/channels")
			{
				channels.GET("", handler.GetChannels)
				channels.POST("", handler.CreateChannel)
				channels.GET("/:id", handler.GetChannel)
				channels.PUT("/:id", handler.UpdateChannel)
				channels.DELETE("/:id", handler.DeleteChannel)
				channels.POST("/:id/test", handler.TestChannel)
			}

			// Subscription management
			subscriptions := protected.Group("/subscriptions")
			{
				subscriptions.GET("", handler.GetSubscriptions)
				subscriptions.POST("", handler.CreateSubscription)
				subscriptions.GET("/:id", handler.GetSubscription)
				subscriptions.PUT("/:id", handler.UpdateSubscription)
				subscriptions.DELETE("/:id", handler.DeleteSubscription)
			}

			// Delivery tracking
			deliveries := protected.Group("/deliveries")
			{
				deliveries.GET("", handler.GetDeliveries)
				deliveries.GET("/:id", handler.GetDelivery)
				deliveries.POST("/:id/retry", handler.RetryDelivery)
			}
		}
	}
}
