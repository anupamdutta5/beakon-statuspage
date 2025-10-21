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

	"github.com/anupamdutta5/shared-resilience"
	"github.com/anupamdutta5/monitoring-service/internal/events"
	"github.com/anupamdutta5/monitoring-service/internal/handlers"
	"github.com/anupamdutta5/monitoring-service/internal/jobs"
	"github.com/anupamdutta5/monitoring-service/internal/services"
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

	// Initialize services
	monitoringService := services.NewMonitoringService(dbManager.GetDB(), logger)
	maintenanceManagementService := services.NewMaintenanceManagementService(dbManager.GetDB(), logger)
	webhookService := services.NewWebhookService(dbManager.GetDB(), logger)
	integrationService := services.NewIntegrationService(dbManager.GetDB(), logger)

	// Initialize Week 3 & Week 4 services (will be used later in background jobs)
	twilioSID := getEnv("TWILIO_ACCOUNT_SID", "")
	twilioToken := getEnv("TWILIO_AUTH_TOKEN", "")
	twilioFrom := getEnv("TWILIO_FROM_NUMBER", "")
	smsService := services.NewSMSService(dbManager.GetDB(), logger, twilioSID, twilioToken, twilioFrom)
	onCallService := services.NewOnCallService(dbManager.GetDB(), logger)
	escalationService := services.NewEscalationService(dbManager.GetDB(), logger, smsService, onCallService)

	// Initialize handlers
	monitoringHandler := handlers.NewMonitoringHandler(monitoringService, maintenanceManagementService, logger)
	webhookHandler := handlers.NewWebhookHandler(webhookService, logger)
	integrationHandler := handlers.NewIntegrationHandler(integrationService, logger)
	onCallHandler := handlers.NewOnCallHandler(onCallService, logger)
	escalationHandler := handlers.NewEscalationHandler(escalationService, logger)
	sslHandler := handlers.NewSSLHandler(dbManager.GetDB())

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

		// SSL Certificate Monitoring
		ssl := api.Group("/ssl")
		{
			ssl.POST("/scan", sslHandler.ScanDomain)
			ssl.GET("/certificates", sslHandler.GetCertificates)
			ssl.GET("/certificates/:id", sslHandler.GetCertificate)
			ssl.GET("/expiring", sslHandler.GetExpiringCertificates)
			ssl.DELETE("/certificates/:id", sslHandler.DeleteCertificate)
			ssl.POST("/certificates/:id/rescan", sslHandler.RescanCertificate)
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

		// On-call schedules (Week 4)
		oncall := api.Group("/oncall")
		{
			// Schedule management
			oncall.GET("/schedules", onCallHandler.GetSchedules)
			oncall.POST("/schedules", onCallHandler.CreateSchedule)
			oncall.GET("/schedules/:id", onCallHandler.GetSchedule)
			oncall.PUT("/schedules/:id", onCallHandler.UpdateSchedule)
			oncall.DELETE("/schedules/:id", onCallHandler.DeleteSchedule)

			// Current on-call
			oncall.GET("/schedules/:id/current", onCallHandler.GetCurrentOnCall)

			// Participant management
			oncall.POST("/schedules/:id/participants", onCallHandler.AddParticipant)
			oncall.DELETE("/schedules/:id/participants/:user_id", onCallHandler.RemoveParticipant)
		}

		// Escalation policies (Week 4)
		escalations := api.Group("/escalations")
		{
			// Policy management
			escalations.GET("/policies", escalationHandler.GetPolicies)
			escalations.POST("/policies", escalationHandler.CreatePolicy)
			escalations.GET("/policies/:id", escalationHandler.GetPolicy)
			escalations.PUT("/policies/:id", escalationHandler.UpdatePolicy)
			escalations.DELETE("/policies/:id", escalationHandler.DeletePolicy)

			// Escalation operations
			escalations.POST("/start", escalationHandler.StartEscalation)
			escalations.POST("/incidents/:incident_id/resolve", escalationHandler.ResolveEscalation)

			// Escalation tracking
			escalations.GET("/active", escalationHandler.GetActiveEscalations)
		}
	}

	// Initialize RabbitMQ event publisher for Week 1 features
	rabbitmqURL := getEnv("RABBITMQ_URL", "amqp://admin:password@localhost:5672/")
	eventPublisher, err := events.NewEventPublisher(rabbitmqURL)
	if err != nil {
		logger.Warn("Failed to initialize RabbitMQ event publisher (will continue without events)", zap.Error(err))
	} else {
		defer eventPublisher.Close()
		logger.Info("RabbitMQ event publisher initialized successfully")
	}

	// Start all background jobs
	logger.Info("Starting background jobs...")

	// Week 1: SSL Certificate Jobs
	if eventPublisher != nil {
		sslExpirationJob := jobs.NewSSLExpirationChecker(dbManager.GetDB(), eventPublisher, 24*time.Hour)
		go sslExpirationJob.Start()
		defer sslExpirationJob.Stop()
		logger.Info("SSL Expiration Checker job started (interval: 24 hours)")
	}

	sslRescanJob := jobs.NewCertificateRescanJob(dbManager.GetDB(), 6*time.Hour)
	go sslRescanJob.Start()
	defer sslRescanJob.Stop()
	logger.Info("Certificate Rescan job started (interval: 6 hours)")

	// Week 3: Heartbeat and Maintenance Jobs
	heartbeatJob := jobs.NewHeartbeatCheckerJob(dbManager.GetDB(), logger, 5*time.Minute)
	go heartbeatJob.Start()
	defer heartbeatJob.Stop()
	logger.Info("Heartbeat Checker job started (interval: 5 minutes)")

	maintenanceJob := jobs.NewMaintenanceWindowJob(dbManager.GetDB(), logger, 1*time.Minute)
	go maintenanceJob.Start()
	defer maintenanceJob.Stop()
	logger.Info("Maintenance Window job started (interval: 1 minute)")

	// Week 4: Escalation and Webhook Jobs
	escalationJob := jobs.NewEscalationProcessorJob(dbManager.GetDB(), logger, smsService, onCallService, 1*time.Minute)
	go escalationJob.Start()
	defer escalationJob.Stop()
	logger.Info("Escalation Processor job started (interval: 1 minute)")

	webhookRetryJob := jobs.NewWebhookRetryJob(dbManager.GetDB(), logger, 5*time.Minute)
	go webhookRetryJob.Start()
	defer webhookRetryJob.Stop()
	logger.Info("Webhook Retry job started (interval: 5 minutes)")

	logger.Info("All background jobs started successfully")

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

// getEnv retrieves an environment variable or returns a default value if not set.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}