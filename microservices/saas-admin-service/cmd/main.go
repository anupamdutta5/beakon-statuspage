// Package main is the entry point for the SaaS Admin Service.
// This is the modernized version using the shared-resilience module.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	resilience "github.com/anupamdutta5/shared-resilience"
	"github.com/anupamdutta5/saas-admin-service/internal/cache"
	"github.com/anupamdutta5/saas-admin-service/internal/config"
	"github.com/anupamdutta5/saas-admin-service/internal/events"
	"github.com/anupamdutta5/saas-admin-service/internal/handlers"
	// "github.com/anupamdutta5/saas-admin-service/internal/models" // Unused after disabling AutoMigrate
	"github.com/anupamdutta5/saas-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // PostgreSQL driver
	"go.uber.org/zap"
)

// setupRoutes configures all the routes for the server
func setupRoutes(router *gin.Engine, adminHandler *handlers.SaaSAdminHandler) {
	// Frontend now served independently on port 3001 (saas-admin-frontend service)
	// All static file routes removed - backend is API-only

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", adminHandler.HealthCheck)

		// Authentication routes (no middleware required)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", adminHandler.Login)
			auth.POST("/logout", adminHandler.Logout)
			auth.GET("/check", adminHandler.CheckAuth)
		}

		// Platform configuration
		v1.GET("/platform", adminHandler.GetPlatform)
		v1.PUT("/platform", adminHandler.UpdatePlatform)

		// Plan management
		plans := v1.Group("/plans")
		{
			plans.GET("", adminHandler.ListPlans)
			plans.POST("", adminHandler.CreatePlan)
			plans.GET("/:id", adminHandler.GetPlan)
			plans.PUT("/:id", adminHandler.UpdatePlan)
			plans.DELETE("/:id", adminHandler.DeletePlan)
			plans.GET("/slug/:slug", adminHandler.GetPlanBySlug)
		}

		// Feature management
		features := v1.Group("/features")
		{
			features.GET("", adminHandler.ListFeatures)
			features.POST("", adminHandler.CreateFeature)
			features.GET("/:id", adminHandler.GetFeature)
			features.PUT("/:id", adminHandler.UpdateFeature)
			features.DELETE("/:id", adminHandler.DeleteFeature)
		}

		// Feature flag management
		featureFlags := v1.Group("/feature-flags")
		{
			featureFlags.GET("", adminHandler.ListFeatureFlags)
			featureFlags.POST("", adminHandler.CreateFeatureFlag)
			featureFlags.GET("/:id", adminHandler.GetFeatureFlag)
			featureFlags.PUT("/:id", adminHandler.UpdateFeatureFlag)
			featureFlags.DELETE("/:id", adminHandler.DeleteFeatureFlag)
		}

		// Statistics
		v1.GET("/stats", adminHandler.GetStats)

		// Admin user management
		adminUsers := v1.Group("/admin-users")
		{
			adminUsers.GET("", adminHandler.ListAdminUsers)
			adminUsers.POST("", adminHandler.CreateAdminUser)
			adminUsers.GET("/:id", adminHandler.GetAdminUser)
			adminUsers.PUT("/:id", adminHandler.UpdateAdminUser)
			adminUsers.DELETE("/:id", adminHandler.DeleteAdminUser)
			adminUsers.PUT("/:id/password", adminHandler.ResetAdminPassword)
		}

		// Notification management
		notifications := v1.Group("/notifications")
		{
			notifications.GET("", adminHandler.ListNotifications)
			notifications.POST("", adminHandler.CreateNotification)
			notifications.GET("/:id", adminHandler.GetNotification)
			notifications.PUT("/:id", adminHandler.UpdateNotification)
			notifications.DELETE("/:id", adminHandler.DeleteNotification)
		}

		// Activity logs
		activities := v1.Group("/activities")
		{
			activities.GET("", adminHandler.ListActivities)
		}

		// Backup management
		backups := v1.Group("/backups")
		{
			backups.GET("", adminHandler.ListBackups)
			backups.POST("", adminHandler.CreateBackup)
			backups.GET("/:id", adminHandler.GetBackup)
			backups.DELETE("/:id", adminHandler.DeleteBackup)
		}

		// Pricing management
		pricingFeatures := v1.Group("/pricing/features")
		{
			pricingFeatures.GET("", adminHandler.GetPricingFeatures)
			pricingFeatures.POST("", adminHandler.CreatePricingFeature)
			pricingFeatures.PUT("/:id", adminHandler.UpdatePricingFeature)
			pricingFeatures.DELETE("/:id", adminHandler.DeletePricingFeature)
		}

		pricingTiers := v1.Group("/pricing/plans/:planId/tiers")
		{
			pricingTiers.GET("", adminHandler.GetPricingTiers)
		}

		tiers := v1.Group("/pricing/tiers")
		{
			tiers.POST("", adminHandler.CreatePricingTier)
			tiers.PUT("/:id", adminHandler.UpdatePricingTier)
			tiers.DELETE("/:id", adminHandler.DeletePricingTier)
		}

		planFeatures := v1.Group("/pricing/plans/:planId/features")
		{
			planFeatures.GET("", adminHandler.GetPlanFeatures)
		}

		planFeaturesActions := v1.Group("/pricing/plans/features")
		{
			planFeaturesActions.POST("/assign", adminHandler.AssignFeatureToPlan)
			planFeaturesActions.POST("/remove", adminHandler.RemoveFeatureFromPlan)
		}

		publicPricing := v1.Group("/pricing/plans")
		{
			publicPricing.GET("/public", adminHandler.GetPublicPricingPlans)
		}

		pricingSync := v1.Group("/pricing")
		{
			pricingSync.POST("/sync", adminHandler.SyncPricingToLandingPage)
		}

		// Tenant management - proxy to tenant-admin service
		tenants := v1.Group("/tenants")
		{
			tenants.GET("", adminHandler.GetTenants)
			tenants.POST("", adminHandler.CreateTenant)
			tenants.GET("/archived", adminHandler.GetArchivedTenants)           // List archived tenants
			tenants.POST("/:id/restore", adminHandler.RestoreTenant)            // Restore archived tenant
			tenants.GET("/verify-sync", adminHandler.VerifySync)                // Verify tenant sync status
			tenants.POST("/reconcile", adminHandler.ReconcileTenants)           // Reconcile out-of-sync tenants
			tenants.GET("/:id", adminHandler.GetTenant)
			tenants.PUT("/:id", adminHandler.UpdateTenant)
			tenants.DELETE("/:id", adminHandler.DeleteTenant)
		}

		// Analytics management
		analytics := v1.Group("/analytics")
		{
			analytics.GET("/overview", adminHandler.GetAnalyticsOverview)
			analytics.GET("/metrics", adminHandler.GetAnalyticsMetrics)
			analytics.POST("/metrics", adminHandler.CreateAnalyticsMetric)
			analytics.GET("/health", adminHandler.CheckAnalyticsHealth)
		}

		// Real-time monitoring
		monitoring := v1.Group("/monitoring")
		{
			monitoring.GET("/monitors", adminHandler.GetMonitors)
			monitoring.POST("/monitors", adminHandler.CreateMonitor)
			monitoring.PUT("/monitors/:id", adminHandler.UpdateMonitor)
			monitoring.DELETE("/monitors/:id", adminHandler.DeleteMonitor)
			monitoring.POST("/monitors/:id/pause", adminHandler.PauseMonitor)
			monitoring.POST("/monitors/:id/resume", adminHandler.ResumeMonitor)
			monitoring.POST("/monitors/pause-all", adminHandler.PauseAllMonitors)
			monitoring.POST("/monitors/resume-all", adminHandler.ResumeAllMonitors)
			monitoring.GET("/stats", adminHandler.GetRealTimeStats)
		}

		// WebSocket endpoint for real-time updates
		v1.GET("/ws/monitoring", adminHandler.WebSocketMonitoring)

		// Incident management
		incidents := v1.Group("/incidents")
		{
			incidents.GET("", adminHandler.GetIncidents)
			incidents.POST("", adminHandler.CreateIncident)
			incidents.GET("/:id", adminHandler.GetIncident)
			incidents.PUT("/:id", adminHandler.UpdateIncident)
			incidents.DELETE("/:id", adminHandler.DeleteIncident)
			incidents.POST("/:id/updates", adminHandler.CreateIncidentUpdate)
			incidents.GET("/:id/updates", adminHandler.GetIncidentUpdates)
			incidents.GET("/stats", adminHandler.GetIncidentStats)
		}

		// Integration management
		integrations := v1.Group("/integrations")
		{
			integrations.GET("", adminHandler.GetIntegrations)
			integrations.POST("", adminHandler.CreateIntegration)
			integrations.PUT("/:id", adminHandler.UpdateIntegration)
			integrations.DELETE("/:id", adminHandler.DeleteIntegration)
			integrations.POST("/:id/test", adminHandler.TestIntegration)
		}

		// Webhook management
		webhooks := v1.Group("/webhooks")
		{
			webhooks.GET("", adminHandler.GetWebhooks)
			webhooks.POST("", adminHandler.CreateWebhook)
			webhooks.PUT("/:id", adminHandler.UpdateWebhook)
			webhooks.DELETE("/:id", adminHandler.DeleteWebhook)
			webhooks.POST("/:id/test", adminHandler.TestWebhook)
		}

		// Subscriber management
		subscribers := v1.Group("/subscribers")
		{
			subscribers.GET("", adminHandler.GetSubscribers)
			subscribers.POST("", adminHandler.CreateSubscriber)
			subscribers.DELETE("/:id", adminHandler.DeleteSubscriber)
		}

		// Component management
		components := v1.Group("/components")
		{
			components.GET("", adminHandler.GetComponents)
			components.POST("", adminHandler.CreateComponent)
			components.GET("/:id", adminHandler.GetComponent)
			components.PUT("/:id", adminHandler.UpdateComponent)
			components.DELETE("/:id", adminHandler.DeleteComponent)
			components.POST("/:id/status", adminHandler.UpdateComponentStatus)
			components.GET("/:id/uptime", adminHandler.GetComponentUptimeStats)
		}

		// Component groups
		v1.GET("/component-groups", adminHandler.GetComponentGroups)
	}
}

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

	logger.Info("Starting SaaS Admin Service",
		zap.String("service", "saas-admin-service"),
		zap.String("version", "1.0.0"),
		zap.String("environment", resilienceConfig.Environment),
		zap.Int("port", resilienceConfig.Server.Port),
	)

	// Ensure primary database (saas_admin) exists before connecting
	primaryDBConfig := resilienceConfig.Database
	primaryDBConfig.Name = "saas_admin" // Force correct database name
	if err := ensureDatabaseExists(primaryDBConfig, logger); err != nil {
		logger.Fatal("Failed to ensure saas_admin database exists", zap.Error(err))
	}

	// Initialize PRIMARY database manager for saas_admin
	dbManager, err := resilience.NewDatabaseManager(primaryDBConfig, logger)
	if err != nil {
		logger.Fatal("Failed to initialize saas_admin database manager", zap.Error(err))
	}
	defer dbManager.Close()

	logger.Info("Primary database connection established", zap.String("database", "saas_admin"))

	// Initialize SECONDARY database connection for tenant_admin_db (credential sync)
	tenantAdminDBConfig := resilienceConfig.Database
	tenantAdminDBConfig.Name = "tenant_admin_db"
	if err := ensureDatabaseExists(tenantAdminDBConfig, logger); err != nil {
		logger.Fatal("Failed to ensure tenant_admin_db exists", zap.Error(err))
	}

	tenantAdminDBManager, err := resilience.NewDatabaseManager(tenantAdminDBConfig, logger)
	if err != nil {
		logger.Fatal("Failed to initialize tenant_admin_db database manager", zap.Error(err))
	}
	defer tenantAdminDBManager.Close()

	logger.Info("Secondary database connection established",
		zap.String("database", "tenant_admin_db"),
		zap.String("purpose", "credential sync for tenant owners"))

	// Auto-migrate SaaS Admin Service models
	// db := dbManager.GetDB() // Unused after disabling AutoMigrate

	// AutoMigrate DISABLED - Using Atlas for database migrations
	// Atlas migrations are managed via atlas.hcl and migrations/ directory
	// Run migrations with: atlas migrate apply --env dev
	/*
	if err := db.AutoMigrate(
		&models.Platform{},
		&models.SaaSPlan{},
		&models.PricingTier{},
		&models.PricingFeature{},
		&models.PlanFeature{},
		&models.SaaSFeature{},
		&models.SaaSFeatureFlag{},
		&models.SaaSAdminUser{},
		&models.SaaSNotification{},
		&models.SaaSActivity{},
		&models.SaaSBackup{},
		&models.SaaSStats{},
		&models.SaaSIncident{},
		&models.SaaSIncidentUpdate{},
		&models.SaaSMaintenanceWindow{},
		&models.SaaSService{},
		&models.SaaSIntegration{},
		&models.SaaSWebhook{},
		&models.SaaSSubscriber{},
		&models.SaaSComponent{},
		&models.SaaSTenant{},
	); err != nil {
		logger.Error("Failed to migrate database", zap.Error(err))
	}
	*/

	logger.Info("SaaS Admin Service using Atlas for migrations - AutoMigrate disabled")

	// Create Gin router
	router := gin.New()

	// CORS middleware for saas-admin-frontend (port 3001)
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Load allowed origins from environment variable
		allowedOriginsStr := os.Getenv("CORS_ALLOWED_ORIGINS")
		if allowedOriginsStr == "" {
			// Default to localhost:3001 for development if not set
			allowedOriginsStr = "http://localhost:3001"
		}
		allowedOrigins := strings.Split(allowedOriginsStr, ",")

		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
				c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
				c.Header("Access-Control-Max-Age", "86400") // 24 hours
				break
			}
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Add comprehensive middleware stack with custom CSP for admin dashboard
	middleware := resilience.DefaultMiddlewareStack(resilienceConfig, logger)

	// Apply all middleware (authentication is now required in ALL environments)
	for _, mw := range middleware {
		router.Use(mw)
	}
	logger.Info("All middleware applied including authentication")

	// CSP header no longer needed - frontend is separate service

	// Initialize service URLs from environment variables
	serviceURLs := config.GetServiceURLsFromEnv()

	// Initialize Redis cache
	redisEnabled := os.Getenv("REDIS_ENABLED") == "true"
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := 6379
	if redisPortStr := os.Getenv("REDIS_PORT"); redisPortStr != "" {
		if port, err := strconv.Atoi(redisPortStr); err == nil {
			redisPort = port
		}
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisConfig := cache.RedisConfig{
		Host:     redisHost,
		Port:     redisPort,
		Password: redisPassword,
		DB:       0,
		Enabled:  redisEnabled,
	}

	redisCache := cache.NewRedisClient(redisConfig, logger)
	logger.Info("Redis cache initialized",
		zap.Bool("enabled", redisEnabled),
		zap.String("host", redisHost),
		zap.Int("port", redisPort))

	// Initialize RabbitMQ event publisher from environment
	rabbitmqHost := os.Getenv("RABBITMQ_HOST")
	if rabbitmqHost == "" {
		rabbitmqHost = "localhost"
	}
	rabbitmqUser := os.Getenv("RABBITMQ_USER")
	if rabbitmqUser == "" {
		rabbitmqUser = "admin"
	}
	rabbitmqPassword := os.Getenv("RABBITMQ_PASSWORD")
	rabbitmqPort := os.Getenv("RABBITMQ_PORT")
	if rabbitmqPort == "" {
		rabbitmqPort = "5672"
	}
	rabbitmqVHost := os.Getenv("RABBITMQ_VHOST")
	if rabbitmqVHost == "" || rabbitmqVHost == "/" {
		rabbitmqVHost = ""
	} else {
		rabbitmqVHost = "/" + rabbitmqVHost
	}

	rabbitmqURL := fmt.Sprintf("amqp://%s:%s@%s:%s%s",
		rabbitmqUser, rabbitmqPassword, rabbitmqHost, rabbitmqPort, rabbitmqVHost)

	eventPublisher, err := events.NewPublisher(events.PublisherConfig{
		URL:    rabbitmqURL,
		Logger: logger,
	})
	if err != nil {
		logger.Warn("Failed to initialize RabbitMQ publisher, event publishing will be disabled",
			zap.Error(err),
			zap.String("url", rabbitmqURL))
		eventPublisher = nil // Will fallback to HTTP sync
	} else {
		logger.Info("RabbitMQ event publisher initialized successfully",
			zap.String("url", rabbitmqURL))
	}

	// Initialize saas admin service and handlers
	saasAdminService, err := services.NewSaaSAdminService(nil, logger, dbManager.GetDB(), redisCache)
	if err != nil {
		logger.Fatal("Failed to create SaaS admin service", zap.Error(err))
	}

	saasAdminHandler := handlers.NewSaaSAdminHandler(saasAdminService, tenantAdminDBManager.GetDB(), eventPublisher, serviceURLs, logger)

	// Setup all enterprise routes
	setupRoutes(router, saasAdminHandler)

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
		logger.Info("SaaS Admin Service server starting",
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

	logger.Info("Shutting down SaaS Admin Service server...")

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

	logger.Info("SaaS Admin Service server exited gracefully")
}

// ensureDatabaseExists creates the database if it doesn't exist
func ensureDatabaseExists(config resilience.DatabaseConfig, logger *zap.Logger) error {
	// Connect to postgres database to create the target database
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.SSLMode)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	defer db.Close()

	// Check if database exists
	var exists bool
	checkQuery := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", config.Name)
	err = db.QueryRow(checkQuery).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Create database if it doesn't exist
	if !exists {
		createQuery := fmt.Sprintf("CREATE DATABASE %s", config.Name)
		_, err = db.Exec(createQuery)
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		logger.Info("Database created successfully", zap.String("database", config.Name))
	} else {
		logger.Info("Database already exists", zap.String("database", config.Name))
	}

	return nil
}