// Package main is the entry point for the Monitoring Service.
// This is the v2.0 version using shared-resilience v2.0 primitives.
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

	"github.com/anupamdutta5/monitoring-service/internal/core/events"
	"github.com/anupamdutta5/monitoring-service/internal/handlers"
	"github.com/anupamdutta5/monitoring-service/internal/jobs"
	"github.com/anupamdutta5/monitoring-service/internal/services"
	resilience "github.com/anupamdutta5/shared-resilience"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// ServiceInfo contains basic service metadata
type ServiceInfo struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
}

// ServiceClientConfig contains HTTP client configuration for downstream services
type ServiceClientConfig struct {
	Timeout        time.Duration                   `yaml:"timeout"`
	CircuitBreaker resilience.CircuitBreakerConfig `yaml:"circuit_breaker"`
}

// MonitoringServiceConfig is the complete configuration for this service
type MonitoringServiceConfig struct {
	// Shared configuration (database, server, JWT, CORS, etc.)
	SharedConfig resilience.Config `yaml:",inline"`

	// Service-specific configuration
	Service       ServiceInfo         `yaml:"service"`
	ServiceClient ServiceClientConfig `yaml:"service_client"`
	Retry         resilience.RetryConfig `yaml:"retry"`
}

func main() {
	// ========================================
	// STEP 1: Load Configuration from YAML
	// ========================================
	loader := resilience.NewConfigLoader("configs")
	var cfg MonitoringServiceConfig
	if err := loader.Load(&cfg); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration (fail fast)
	if err := cfg.SharedConfig.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// ========================================
	// STEP 2: Initialize Logger
	// ========================================
	var logger *zap.Logger
	var err error

	if cfg.Service.Environment == "production" {
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

	logger.Info("Starting Monitoring Service (v2.0)",
		zap.String("service", cfg.Service.Name),
		zap.String("version", cfg.Service.Version),
		zap.String("environment", cfg.Service.Environment),
		zap.Int("port", cfg.SharedConfig.Server.Port),
		zap.String("shared_resilience", resilience.Version),
	)

	// ========================================
	// STEP 3: Initialize Prometheus Registry
	// ========================================
	registry := prometheus.NewRegistry()
	logger.Info("Prometheus registry initialized (injected, not global)")

	// ========================================
	// STEP 4: Initialize Metrics (with injected registry)
	// ========================================
	metrics, err := resilience.NewMetrics(resilience.MetricsConfig{
		ServiceName: cfg.Service.Name,
		Namespace:   "beakon",
		Subsystem:   "monitoring",
		Enabled:     cfg.SharedConfig.Monitoring.MetricsEnabled,
		Registry:    registry,
	})
	if err != nil {
		logger.Fatal("Failed to initialize metrics", zap.Error(err))
	}
	logger.Info("Metrics initialized with custom registry")

	// ========================================
	// STEP 5: Initialize Database Manager
	// ========================================
	dbManager, err := resilience.NewDatabaseManager(cfg.SharedConfig.Database, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database manager", zap.Error(err))
	}
	defer dbManager.Close()
	logger.Info("Database manager initialized",
		zap.String("database", cfg.SharedConfig.Database.Name),
		zap.Int("max_open_conns", cfg.SharedConfig.Database.MaxOpenConns),
	)

	// ========================================
	// STEP 6: Initialize ServiceClient (with config)
	// ========================================
	serviceClientCfg := resilience.ServiceClientConfig{
		Timeout:        cfg.ServiceClient.Timeout,
		CircuitBreaker: cfg.ServiceClient.CircuitBreaker,
	}

	if err := serviceClientCfg.Validate(); err != nil {
		logger.Fatal("ServiceClient configuration validation failed", zap.Error(err))
	}

	// Load service endpoints from YAML (if exists)
	endpoints, err := loader.LoadServiceEndpoints()
	if err != nil {
		logger.Warn("Failed to load service endpoints", zap.Error(err))
		endpoints = make(map[string]resilience.ServiceEndpoint) // Empty for now
	}

	serviceClient, err := resilience.NewServiceClient(endpoints, serviceClientCfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize ServiceClient", zap.Error(err))
	}
	logger.Info("ServiceClient initialized with circuit breakers",
		zap.Duration("timeout", cfg.ServiceClient.Timeout),
	)

	// ========================================
	// STEP 7: Initialize Retry Manager
	// ========================================
	retryManager, err := resilience.NewRetryManager(cfg.Retry, logger)
	if err != nil {
		logger.Fatal("Failed to initialize retry manager", zap.Error(err))
	}
	logger.Info("Retry manager initialized",
		zap.Int("max_retries", cfg.Retry.MaxRetries),
		zap.Duration("initial_delay", cfg.Retry.InitialDelay),
	)

	// ========================================
	// STEP 8: Initialize Gin Router
	// ========================================
	router := gin.New()

	// Add middleware stack
	middleware := resilience.DefaultMiddlewareStack(&cfg.SharedConfig, logger)
	for _, mw := range middleware {
		router.Use(mw)
	}
	logger.Info("Middleware stack applied (CORS, rate limiting, security headers)")

	// Log metrics and retry manager for debugging (can be used later if needed)
	_ = metrics
	_ = retryManager

	// ========================================
	// STEP 9: Initialize Services
	// ========================================
	monitoringService := services.NewMonitoringService(dbManager.GetDB(), serviceClient, logger)
	maintenanceManagementService := services.NewMaintenanceManagementService(dbManager.GetDB(), logger)
	webhookService := services.NewWebhookService(dbManager.GetDB(), serviceClient, logger)
	integrationService := services.NewIntegrationService(dbManager.GetDB(), serviceClient, logger)

	// Initialize anomaly detection services (Week 13)
	baselineCalculator := services.NewBaselineCalculator(dbManager.GetDB())
	anomalyDetectionService := services.NewAnomalyDetectionService(dbManager.GetDB(), baselineCalculator)

	// Initialize Week 3 & Week 4 services (will be used later in background jobs)
	twilioSID := getEnv("TWILIO_ACCOUNT_SID", "")
	twilioToken := getEnv("TWILIO_AUTH_TOKEN", "")
	twilioFrom := getEnv("TWILIO_FROM_NUMBER", "")
	smsService := services.NewSMSService(dbManager.GetDB(), serviceClient, logger, twilioSID, twilioToken, twilioFrom)
	onCallService := services.NewOnCallService(dbManager.GetDB(), logger)
	escalationService := services.NewEscalationService(dbManager.GetDB(), logger, smsService, onCallService)

	// Initialize integration services (Slack, PagerDuty, Discord, Telegram)
	slackService := services.NewSlackIntegrationService(dbManager.GetDB(), serviceClient, logger)
	pagerdutyService := services.NewPagerDutyIntegrationService(dbManager.GetDB(), serviceClient, logger)
	discordService := services.NewDiscordIntegrationService(dbManager.GetDB(), serviceClient, logger)
	telegramService := services.NewTelegramIntegrationService(dbManager.GetDB(), serviceClient, logger)

	logger.Info("All services initialized successfully")

	// ========================================
	// STEP 10: Initialize Handlers
	// ========================================
	monitoringHandler := handlers.NewMonitoringHandler(monitoringService, maintenanceManagementService, logger)
	webhookHandler := handlers.NewWebhookHandler(webhookService, logger)
	integrationHandler := handlers.NewIntegrationHandler(integrationService, logger)
	onCallHandler := handlers.NewOnCallHandler(onCallService, logger)
	escalationHandler := handlers.NewEscalationHandler(escalationService, logger)
	sslHandler := handlers.NewSSLHandler(dbManager.GetDB())
	slackHandler := handlers.NewSlackHandler(dbManager.GetDB(), logger, slackService)
	pagerdutyHandler := handlers.NewPagerDutyHandler(dbManager.GetDB(), logger, pagerdutyService)
	discordHandler := handlers.NewDiscordHandler(dbManager.GetDB(), logger, discordService)
	telegramHandler := handlers.NewTelegramHandler(dbManager.GetDB(), logger, telegramService)
	anomalyHandler := handlers.NewAnomalyHandler(anomalyDetectionService)

	logger.Info("All handlers initialized successfully")

	// ========================================
	// STEP 11: Setup Routes
	// ========================================

	// Health and metrics endpoints
	router.GET("/health", monitoringHandler.Health)
	router.GET("/health/live", monitoringHandler.Health)
	router.GET("/health/ready", monitoringHandler.Health)
	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))

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

			// Maintenance templates - DISABLED until template tables are implemented
			// maintenance.GET("/templates", monitoringHandler.GetMaintenanceTemplates)
			// maintenance.POST("/templates", monitoringHandler.CreateMaintenanceTemplate)
			// maintenance.POST("/templates/:template_id/create-maintenance", monitoringHandler.CreateMaintenanceFromTemplate)
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

			// Slack Integration
			slack := integrations.Group("/slack")
			{
				slack.GET("/install", slackHandler.InstallSlack)           // OAuth install redirect
				slack.GET("/callback", slackHandler.OAuthCallback)         // OAuth callback
				slack.GET("", slackHandler.GetIntegrations)                // Get all Slack integrations
				slack.DELETE("/:id", slackHandler.DeleteIntegration)       // Delete integration
				slack.POST("/:id/test", slackHandler.TestIntegration)      // Send test notification
			}

			// PagerDuty Integration
			pagerduty := integrations.Group("/pagerduty")
			{
				pagerduty.POST("", pagerdutyHandler.CreateIntegration)         // Create integration
				pagerduty.GET("", pagerdutyHandler.GetIntegrations)            // Get all integrations
				pagerduty.GET("/:id", pagerdutyHandler.GetIntegration)         // Get specific integration
				pagerduty.PUT("/:id", pagerdutyHandler.UpdateIntegration)      // Update integration
				pagerduty.DELETE("/:id", pagerdutyHandler.DeleteIntegration)   // Delete integration
				pagerduty.POST("/:id/test", pagerdutyHandler.TestIntegration)  // Send test event
				pagerduty.POST("/:id/monitors", pagerdutyHandler.MapMonitor)   // Map monitor to integration
				pagerduty.DELETE("/:id/monitors/:mapping_id", pagerdutyHandler.UnmapMonitor) // Unmap monitor
				pagerduty.GET("/incidents", pagerdutyHandler.GetIncidentHistory) // Get incident history
				pagerduty.POST("/webhook", pagerdutyHandler.WebhookReceiver)   // Webhook receiver
			}

			// Discord Integration
			discord := integrations.Group("/discord")
			{
				discord.POST("", discordHandler.CreateIntegration)                  // Create integration
				discord.GET("", discordHandler.GetIntegrations)                     // Get all integrations
				discord.GET("/:id", discordHandler.GetIntegration)                  // Get specific integration
				discord.PUT("/:id", discordHandler.UpdateIntegration)               // Update integration
				discord.DELETE("/:id", discordHandler.DeleteIntegration)            // Delete integration
				discord.POST("/:id/test", discordHandler.TestIntegration)           // Send test notification
				discord.POST("/:id/subscribe", discordHandler.SubscribeChannel)     // Subscribe channel to monitor
				discord.GET("/:id/subscriptions", discordHandler.GetChannelSubscriptions) // Get subscriptions
				discord.DELETE("/subscriptions/:id", discordHandler.UnsubscribeChannel)   // Unsubscribe channel
				discord.GET("/:id/notifications", discordHandler.GetNotificationHistory)  // Get notification history
				discord.GET("/:id/stats", discordHandler.GetNotificationStats)      // Get statistics
			}

			// Telegram Integration
			telegram := integrations.Group("/telegram")
			{
				telegram.POST("", telegramHandler.CreateIntegration)                  // Create integration
				telegram.GET("", telegramHandler.GetIntegrations)                     // Get all integrations
				telegram.GET("/:id", telegramHandler.GetIntegration)                  // Get specific integration
				telegram.PUT("/:id", telegramHandler.UpdateIntegration)               // Update integration
				telegram.DELETE("/:id", telegramHandler.DeleteIntegration)            // Delete integration
				telegram.POST("/:id/test", telegramHandler.TestIntegration)           // Send test message
				telegram.POST("/:id/subscribe", telegramHandler.SubscribeChat)        // Subscribe chat to monitor
				telegram.GET("/:id/subscriptions", telegramHandler.GetChatSubscriptions) // Get subscriptions
				telegram.DELETE("/subscriptions/:id", telegramHandler.UnsubscribeChat)   // Unsubscribe chat
				telegram.GET("/:id/notifications", telegramHandler.GetNotificationHistory)  // Get notification history
				telegram.GET("/:id/stats", telegramHandler.GetNotificationStats)      // Get statistics
			}
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

		// Anomaly detection (Week 13)
		anomalies := api.Group("/anomalies")
		{
			// Anomaly management
			anomalies.GET("", anomalyHandler.GetAnomalies)
			anomalies.GET("/:id", anomalyHandler.GetAnomalyByID)
			anomalies.POST("/:id/acknowledge", anomalyHandler.AcknowledgeAnomaly)
			anomalies.POST("/:id/resolve", anomalyHandler.ResolveAnomaly)

			// Statistics and insights
			anomalies.GET("/statistics", anomalyHandler.GetStatistics)

			// Baseline management
			anomalies.GET("/baselines", anomalyHandler.GetBaselines)

			// Configuration
			anomalies.GET("/config", anomalyHandler.GetConfig)
			anomalies.PUT("/config", anomalyHandler.UpdateConfig)
		}
	}

	logger.Info("All routes registered successfully")

	// ========================================
	// STEP 12: Initialize RabbitMQ Event Publisher
	// ========================================
	rabbitmqURL := getEnv("RABBITMQ_URL", "amqp://admin:password@localhost:5672/")
	eventPublisher, err := events.NewEventPublisher(rabbitmqURL)
	if err != nil {
		logger.Warn("Failed to initialize RabbitMQ event publisher (will continue without events)", zap.Error(err))
	} else {
		defer eventPublisher.Close()
		logger.Info("RabbitMQ event publisher initialized successfully")
	}

	// ========================================
	// STEP 13: Start Background Jobs
	// ========================================
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

	// P1 Feature: Maintenance Automation (auto-reminder, auto-start, auto-complete)
	maintenanceWindowJob := jobs.NewMaintenanceWindowJob(dbManager.GetDB(), logger, 1*time.Minute)
	go maintenanceWindowJob.Start()
	defer maintenanceWindowJob.Stop()
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

	// Week 13: Anomaly Detection Jobs
	metricCollectionJob := jobs.NewMetricCollectionJob(dbManager.GetDB(), anomalyDetectionService, 1*time.Minute)
	go metricCollectionJob.Start()
	logger.Info("Metric Collection job started (interval: 1 minute)")

	baselineUpdateJob := jobs.NewBaselineUpdateJob(dbManager.GetDB(), baselineCalculator, 6*time.Hour)
	go baselineUpdateJob.Start()
	logger.Info("Baseline Update job started (interval: 6 hours)")

	cleanupJob := jobs.NewCleanupJob(dbManager.GetDB(), baselineCalculator, 24*time.Hour, 90, 30)
	go cleanupJob.Start()
	logger.Info("Cleanup job started (interval: 24 hours, metric retention: 90 days, anomaly retention: 30 days)")

	logger.Info("All background jobs started successfully")

	// ========================================
	// STEP 14: Create HTTP Server
	// ========================================
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.SharedConfig.Server.Host, cfg.SharedConfig.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.SharedConfig.Server.ReadTimeout,
		WriteTimeout: cfg.SharedConfig.Server.WriteTimeout,
		IdleTimeout:  cfg.SharedConfig.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Monitoring Service server starting",
			zap.String("addr", server.Addr),
			zap.Duration("read_timeout", cfg.SharedConfig.Server.ReadTimeout),
			zap.Duration("write_timeout", cfg.SharedConfig.Server.WriteTimeout),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// ========================================
	// STEP 15: Setup Health Check Monitoring
	// ========================================
	if cfg.SharedConfig.Monitoring.Enabled {
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

	// ========================================
	// STEP 16: Graceful Shutdown
	// ========================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Monitoring Service server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.SharedConfig.Server.GracefulStop)
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
