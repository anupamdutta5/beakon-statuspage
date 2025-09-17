// Package main is the entry point for the Monitoring Service.
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
	"github.com/anupamdutta5/statuspage-monitoring-service/internal/handlers"
	"github.com/anupamdutta5/statuspage-monitoring-service/internal/services"
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

	logger.Info("Starting Monitoring Service",
		zap.String("service", "monitoring-service"),
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

	// Initialize modernized handlers
	monitoringService := services.NewMonitoringService(dbManager.GetDB(), logger)
	maintenanceManagementService := services.NewMaintenanceManagementService(dbManager.GetDB(), logger)
	webhookService := services.NewWebhookService(dbManager.GetDB(), logger)
	integrationService := services.NewIntegrationService(dbManager.GetDB(), logger)
	monitoringHandler := handlers.NewMonitoringHandler(monitoringService, maintenanceManagementService, logger)
	webhookHandler := handlers.NewWebhookHandler(webhookService, logger)
	integrationHandler := handlers.NewIntegrationHandler(integrationService, logger)

	// Setup basic routes
	router.GET("/health", monitoringHandler.Health)
	router.GET("/metrics", monitoringHandler.Metrics)

	// Public API routes (no authentication required)
	public := router.Group("/api/v1/public")
	{
		public.GET("/status", monitoringHandler.GetPublicStatus)
		public.GET("/health", monitoringHandler.GetPublicHealth)
	}

	// Protected API routes (authentication required)
	api := router.Group("/api/v1")
	// TODO: Add authentication middleware here
	{
		// Monitoring overview
		api.GET("/monitoring/overview", monitoringHandler.GetMonitoringOverview)

		// Service management
		services := api.Group("/services")
		{
			services.GET("", monitoringHandler.GetServices)
			services.POST("", monitoringHandler.CreateService)
			services.GET("/:id", monitoringHandler.GetService)
			services.PUT("/:id", monitoringHandler.UpdateService)
			services.DELETE("/:id", monitoringHandler.DeleteService)
			services.GET("/:id/health", monitoringHandler.GetServiceHealth)
			services.GET("/:id/metrics", monitoringHandler.GetServiceMetrics)

			// Health checks for services
			services.POST("/:id/health-checks", monitoringHandler.CreateHealthCheck)
			services.PUT("/health-checks/:check_id", monitoringHandler.UpdateHealthCheck)
			services.DELETE("/health-checks/:check_id", monitoringHandler.DeleteHealthCheck)
		}

		// Alert management
		alerts := api.Group("/alerts")
		{
			alerts.GET("", monitoringHandler.GetAlerts)
			alerts.POST("", monitoringHandler.CreateAlert)
			alerts.GET("/:id", monitoringHandler.GetAlert)
			alerts.PUT("/:id", monitoringHandler.UpdateAlert)
			alerts.DELETE("/:id", monitoringHandler.DeleteAlert)
			alerts.POST("/:id/acknowledge", monitoringHandler.AcknowledgeAlert)
			alerts.POST("/:id/resolve", monitoringHandler.ResolveAlert)
		}

		// Maintenance management
		maintenance := api.Group("/maintenance")
		{
			// Maintenance windows
			maintenance.GET("/windows", monitoringHandler.GetMaintenanceWindows)
			maintenance.POST("/windows", monitoringHandler.CreateMaintenanceWindow)
			maintenance.GET("/windows/:id", monitoringHandler.GetMaintenanceWindow)
			maintenance.PUT("/windows/:id", monitoringHandler.UpdateMaintenanceWindow)
			maintenance.DELETE("/windows/:id", monitoringHandler.DeleteMaintenanceWindow)

			// Maintenance window actions
			maintenance.POST("/windows/:id/start", monitoringHandler.StartMaintenance)
			maintenance.POST("/windows/:id/complete", monitoringHandler.CompleteMaintenance)
			maintenance.POST("/windows/:id/cancel", monitoringHandler.CancelMaintenance)

			// Maintenance updates
			maintenance.POST("/windows/:id/updates", monitoringHandler.CreateMaintenanceUpdate)
			maintenance.GET("/windows/:id/updates", monitoringHandler.GetMaintenanceUpdates)

			// Maintenance components
			maintenance.POST("/windows/:id/components", monitoringHandler.AddMaintenanceComponent)
			maintenance.DELETE("/windows/:id/components/:component_id", monitoringHandler.RemoveMaintenanceComponent)

			// Convenience endpoints
			maintenance.GET("/upcoming", monitoringHandler.GetUpcomingMaintenance)
			maintenance.GET("/active", monitoringHandler.GetActiveMaintenance)
			maintenance.GET("/statistics", monitoringHandler.GetMaintenanceStatistics)

			// Maintenance templates
			maintenance.GET("/templates", monitoringHandler.GetMaintenanceTemplates)
			maintenance.POST("/templates", monitoringHandler.CreateMaintenanceTemplate)
			maintenance.POST("/templates/:template_id/create-maintenance", monitoringHandler.CreateMaintenanceFromTemplate)
		}

		// Placeholder routes for future features
		uptime := api.Group("/uptime")
		{
			uptime.GET("/checks", monitoringHandler.GetUptimeChecks)
			uptime.POST("/checks", monitoringHandler.CreateUptimeCheck)
			uptime.GET("/checks/:id", monitoringHandler.GetUptimeCheck)
			uptime.PUT("/checks/:id", monitoringHandler.UpdateUptimeCheck)
			uptime.DELETE("/checks/:id", monitoringHandler.DeleteUptimeCheck)
			uptime.GET("/results", monitoringHandler.GetUptimeResults)
			uptime.GET("/statistics", monitoringHandler.GetUptimeStatistics)
		}

		performance := api.Group("/performance")
		{
			performance.GET("/metrics", monitoringHandler.GetPerformanceMetrics)
			performance.POST("/metrics", monitoringHandler.CreatePerformanceMetric)
			performance.GET("/metrics/:id", monitoringHandler.GetPerformanceMetric)
			performance.PUT("/metrics/:id", monitoringHandler.UpdatePerformanceMetric)
			performance.DELETE("/metrics/:id", monitoringHandler.DeletePerformanceMetric)
			performance.GET("/data", monitoringHandler.GetPerformanceData)
			performance.POST("/data", monitoringHandler.AddPerformanceData)
		}

		logs := api.Group("/logs")
		{
			logs.GET("", monitoringHandler.GetLogs)
			logs.GET("/search", monitoringHandler.SearchLogs)
			logs.GET("/aggregate", monitoringHandler.AggregateLogs)
			logs.GET("/stream", monitoringHandler.StreamLogs)
		}

		// Webhook management
		webhooks := api.Group("/webhooks")
		{
			// Webhook endpoints management
			webhooks.GET("", webhookHandler.GetWebhookEndpoints)
			webhooks.POST("", webhookHandler.CreateWebhookEndpoint)
			webhooks.GET("/:id", webhookHandler.GetWebhookEndpoint)
			webhooks.PUT("/:id", webhookHandler.UpdateWebhookEndpoint)
			webhooks.DELETE("/:id", webhookHandler.DeleteWebhookEndpoint)
			webhooks.POST("/:id/test", webhookHandler.TestWebhookEndpoint)

			// Webhook delivery management
			webhooks.GET("/deliveries", webhookHandler.GetWebhookDeliveries)
			webhooks.POST("/deliveries/:id/retry", webhookHandler.RetryWebhookDelivery)

			// Webhook statistics and utilities
			webhooks.GET("/statistics", webhookHandler.GetWebhookStatistics)
			webhooks.GET("/event-types", webhookHandler.GetSupportedEventTypes)
		}

		// Third-party integrations
		integrations := api.Group("/integrations")
		{
			// Integration management
			integrations.GET("", integrationHandler.GetIntegrations)
			integrations.POST("", integrationHandler.CreateIntegration)
			integrations.GET("/:id", integrationHandler.GetIntegration)
			integrations.PUT("/:id", integrationHandler.UpdateIntegration)
			integrations.DELETE("/:id", integrationHandler.DeleteIntegration)

			// Integration operations
			integrations.POST("/:id/sync", integrationHandler.SyncIntegration)
			integrations.POST("/:id/test", integrationHandler.TestIntegration)

			// Component mappings
			integrations.POST("/:id/mappings", integrationHandler.CreateComponentMapping)
			integrations.GET("/:id/mappings", integrationHandler.GetComponentMappings)

			// Integration logs and monitoring
			integrations.GET("/:id/logs", integrationHandler.GetIntegrationSyncLogs)

			// Supported integrations
			integrations.GET("/supported", integrationHandler.GetSupportedIntegrations)
		}
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
		logger.Info("Monitoring Service server starting",
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

	logger.Info("Shutting down Monitoring Service server...")

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

	logger.Info("Monitoring Service server exited gracefully")
}