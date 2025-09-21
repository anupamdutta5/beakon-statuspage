// Package main is the entry point for the Analytics Service.
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
	"github.com/anupamdutta5/analytics-service/internal/handlers"
	"github.com/anupamdutta5/analytics-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load configuration from environment variables
	config := resilience.LoadConfigFromEnv()

	// Create startup manager for proper error handling
	startupMgr, err := resilience.NewStartupManager("analytics-service", config)
	if err != nil {
		// This is the only acceptable use of fatal - when we can't even initialize logging
		fmt.Fprintf(os.Stderr, "Failed to create startup manager: %v\n", err)
		os.Exit(1)
	}

	// Set up panic recovery
	defer startupMgr.RecoverFromPanic()

	logger := startupMgr.Logger
	defer logger.Sync()

	// Validate configuration
	if err := startupMgr.ValidateConfiguration(); err != nil {
		startupMgr.HandleStartupError(err)
		return
	}

	// Set Gin mode based on environment
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	logger.Info("Starting Analytics Service",
		zap.String("service", "analytics-service"),
		zap.String("version", "1.0.0"),
		zap.String("environment", config.Environment),
		zap.Int("port", config.Server.Port),
	)

	// Initialize database manager with connection pooling and health checks
	dbManager, err := resilience.NewDatabaseManager(config.Database, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database manager", zap.Error(err))
	}
	defer dbManager.Close()

	// Test database connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dbManager.HealthCheck(ctx); err != nil {
		logger.Fatal("Database health check failed", zap.Error(err))
	}

	logger.Info("Database connection established successfully")

	// Initialize circuit breakers for external dependencies
	var circuitBreakers = make(map[string]*resilience.CircuitBreaker)

	if config.CircuitBreaker.Database.Enabled {
		circuitBreakers["database"] = resilience.NewCircuitBreaker(
			config.CircuitBreaker.Database.Name,
			config.CircuitBreaker.Database,
			logger,
		)
	}

	if config.CircuitBreaker.External.Enabled {
		circuitBreakers["external"] = resilience.NewCircuitBreaker(
			config.CircuitBreaker.External.Name,
			config.CircuitBreaker.External,
			logger,
		)
	}

	// Initialize cache if enabled
	var cache resilience.Cache
	if config.Cache.Enabled {
		if config.Cache.Type == "redis" {
			cache = resilience.NewRedisCache(config.Redis, config.Cache, logger)
		} else {
			cache = resilience.NewInMemoryCache(config.Cache, logger)
		}
		logger.Info("Cache initialized", zap.String("type", config.Cache.Type))
	}

	// Initialize rate limiter if enabled
	var rateLimiter resilience.RateLimiter
	if config.RateLimit.Enabled {
		if config.Cache.Type == "redis" {
			rateLimiter = resilience.NewRedisRateLimiter(config.Redis, config.RateLimit, logger)
		} else {
			rateLimiter = resilience.NewInMemoryRateLimiter(config.RateLimit, logger)
		}
		logger.Info("Rate limiter initialized")
	}

	// Initialize business services with modernized dependencies
	analyticsService := services.NewAnalyticsService(dbManager.GetDB(), logger)
	slaService := services.NewSLAService(dbManager.GetDB(), logger)

	// Create Gin router
	router := gin.New()

	// Add comprehensive middleware stack
	middleware := resilience.DefaultMiddlewareStack(config, logger)
	for _, mw := range middleware {
		router.Use(mw)
	}

	// Add rate limiting middleware if enabled
	if rateLimiter != nil {
		router.Use(resilience.RateLimitMiddleware(config.RateLimit.RequestsPerMinute, time.Minute))
	}

	// Initialize modernized handlers with shared error handling
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService, logger)
	slaHandler := handlers.NewSLAHandler(slaService, logger)

	// Setup routes with improved structure
	setupModernizedRoutes(router, analyticsHandler, slaHandler, config)

	// Create HTTP server with proper timeouts and configuration
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port),
		Handler:      router,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		IdleTimeout:  config.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Analytics Service server starting",
			zap.String("addr", server.Addr),
			zap.Duration("read_timeout", config.Server.ReadTimeout),
			zap.Duration("write_timeout", config.Server.WriteTimeout),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Setup health check monitoring
	if config.Monitoring.Enabled {
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

	logger.Info("Shutting down Analytics Service server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), config.Server.GracefulStop)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	// Close database connections
	if err := dbManager.Close(); err != nil {
		logger.Error("Failed to close database connections", zap.Error(err))
	}

	// Close cache if initialized
	if cache != nil {
		cache.Close()
	}

	// Close rate limiter if initialized
	if rateLimiter != nil {
		rateLimiter.Close()
	}

	logger.Info("Analytics Service server exited gracefully")
}

// setupModernizedRoutes configures basic routes with existing handlers
func setupModernizedRoutes(router *gin.Engine, handler *handlers.AnalyticsHandler, slaHandler *handlers.SLAHandler, config *resilience.Config) {
	// Health check endpoints
	health := router.Group("/health")
	{
		health.GET("", handler.Health)
	}

	// API routes with versioning
	api := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		public := api.Group("/public")
		{
			public.GET("/metrics", handler.GetPublicMetrics)
			public.GET("/reports", handler.GetPublicReports)
		}

		// Protected routes (authentication required)
		protected := api.Group("/")

		// Add JWT authentication middleware
		if config.JWT.Secret != "" {
			protected.Use(resilience.AuthMiddleware(config.JWT))
		}

		// Add tenant middleware for multi-tenancy
		protected.Use(resilience.TenantMiddleware())

		{
			// Core analytics and metrics management
			analytics := protected.Group("/analytics")
			{
				analytics.GET("/overview", handler.GetAnalyticsOverview)

				// Metrics management
				metrics := analytics.Group("/metrics")
				{
					metrics.GET("", handler.GetMetrics)
					metrics.POST("", handler.CreateMetric)
					metrics.GET("/:id", handler.GetMetric)
					metrics.PUT("/:id", handler.UpdateMetric)
					metrics.DELETE("/:id", handler.DeleteMetric)
					metrics.GET("/:id/data", handler.GetMetricData)
					metrics.POST("/:id/data", handler.AddMetricData)
				}

				// Reports management
				reports := analytics.Group("/reports")
				{
					reports.GET("", handler.GetReports)
					reports.POST("", handler.CreateReport)
					reports.GET("/:id", handler.GetReport)
					reports.PUT("/:id", handler.UpdateReport)
					reports.DELETE("/:id", handler.DeleteReport)
					reports.POST("/:id/generate", handler.GenerateReport)
					reports.GET("/:id/download", handler.DownloadReport)
				}

				// Dashboard management
				dashboards := analytics.Group("/dashboards")
				{
					dashboards.GET("", handler.GetDashboards)
					dashboards.POST("", handler.CreateDashboard)
					dashboards.GET("/:id", handler.GetDashboard)
					dashboards.PUT("/:id", handler.UpdateDashboard)
					dashboards.DELETE("/:id", handler.DeleteDashboard)
					dashboards.GET("/:id/widgets", handler.GetDashboardWidgets)
					dashboards.POST("/:id/widgets", handler.AddDashboardWidget)
					dashboards.PUT("/:id/widgets/:widget_id", handler.UpdateDashboardWidget)
					dashboards.DELETE("/:id/widgets/:widget_id", handler.DeleteDashboardWidget)
				}

				// Data exports
				exports := analytics.Group("/exports")
				{
					exports.GET("/csv", handler.ExportCSV)
					exports.GET("/json", handler.ExportJSON)
					exports.GET("/excel", handler.ExportExcel)
				}
			}

			// SLA Management and Reporting
			sla := protected.Group("/sla")
			{
				// SLA definitions
				sla.POST("", slaHandler.CreateSLA)
				sla.GET("", slaHandler.GetSLAs)
				sla.GET("/:id", slaHandler.GetSLA)
				sla.PUT("/:id", slaHandler.UpdateSLA)
				sla.DELETE("/:id", slaHandler.DeleteSLA)

				// SLA measurements and calculations
				sla.POST("/:id/calculate", slaHandler.CalculateSLAMeasurement)
				sla.GET("/:id/measurements", slaHandler.GetSLAMeasurements)

				// SLA breaches
				sla.GET("/breaches", slaHandler.GetSLABreaches)

				// SLA reports
				sla.POST("/reports", slaHandler.GenerateSLAReport)
				sla.GET("/reports", slaHandler.GetSLAReports)

				// SLA statistics
				sla.POST("/statistics", slaHandler.GetSLAStatistics)

				// SLA targets (templates)
				sla.POST("/targets", slaHandler.CreateSLATarget)
				sla.GET("/targets", slaHandler.GetSLATargets)

				// Uptime calculations
				sla.POST("/uptime/calculate", slaHandler.CalculateUptime)

				// Response time recording
				sla.POST("/response-time", slaHandler.RecordResponseTime)
			}
		}
	}
}