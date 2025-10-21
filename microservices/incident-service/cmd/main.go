// Package main is the entry point for the Incident Service.
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
	"github.com/anupamdutta5/incident-service/internal/handlers"
	"github.com/anupamdutta5/incident-service/internal/services"
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

	logger.Info("Starting Incident Service",
		zap.String("service", "incident-service"),
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

	// Test database connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dbManager.HealthCheck(ctx); err != nil {
		logger.Fatal("Database health check failed", zap.Error(err))
	}

	logger.Info("Database connection established successfully")

	// Initialize circuit breakers for external dependencies
	var circuitBreakers = make(map[string]*resilience.CircuitBreaker)

	if resilienceConfig.CircuitBreaker.Database.Enabled {
		circuitBreakers["database"] = resilience.NewCircuitBreaker(
			resilienceConfig.CircuitBreaker.Database.Name,
			resilienceConfig.CircuitBreaker.Database,
			logger,
		)
	}

	if resilienceConfig.CircuitBreaker.External.Enabled {
		circuitBreakers["external"] = resilience.NewCircuitBreaker(
			resilienceConfig.CircuitBreaker.External.Name,
			resilienceConfig.CircuitBreaker.External,
			logger,
		)
	}

	// Initialize cache if enabled
	var cache resilience.Cache
	if resilienceConfig.Cache.Enabled {
		if resilienceConfig.Cache.Type == "redis" {
			cache = resilience.NewRedisCache(resilienceConfig.Redis, resilienceConfig.Cache, logger)
		} else {
			cache = resilience.NewInMemoryCache(resilienceConfig.Cache, logger)
		}
		logger.Info("Cache initialized", zap.String("type", resilienceConfig.Cache.Type))
	}

	// Initialize rate limiter if enabled
	var rateLimiter resilience.RateLimiter
	if resilienceConfig.RateLimit.Enabled {
		if resilienceConfig.Cache.Type == "redis" {
			rateLimiter = resilience.NewRedisRateLimiter(resilienceConfig.Redis, resilienceConfig.RateLimit, logger)
		} else {
			rateLimiter = resilience.NewInMemoryRateLimiter(resilienceConfig.RateLimit, logger)
		}
		logger.Info("Rate limiter initialized")
	}

	// Initialize business services with modernized dependencies
	incidentService := services.NewIncidentService(dbManager.GetDB(), logger)
	// templateService := services.NewIncidentTemplateService(dbManager.GetDB(), logger) // Temporarily disabled due to model conflicts

	// Create Gin router
	router := gin.New()

	// Add comprehensive middleware stack
	middleware := resilience.DefaultMiddlewareStack(resilienceConfig, logger)
	for _, mw := range middleware {
		router.Use(mw)
	}

	// Add rate limiting middleware if enabled
	if rateLimiter != nil {
		router.Use(resilience.RateLimitMiddleware(resilienceConfig.RateLimit.RequestsPerMinute, time.Minute))
	}

	// Initialize modernized handlers
	incidentHandler := handlers.NewIncidentHandler(incidentService, logger)
	// templateHandler := handlers.NewIncidentTemplateHandler(templateService, logger) // Temporarily disabled

	// Setup routes with improved structure
	setupModernizedRoutes(router, incidentHandler, nil, resilienceConfig) // Pass nil for templateHandler

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
		logger.Info("Incident Service server starting",
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

	logger.Info("Shutting down Incident Service server...")

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

	// Close cache if initialized
	if cache != nil {
		cache.Close()
	}

	// Close rate limiter if initialized
	if rateLimiter != nil {
		rateLimiter.Close()
	}

	logger.Info("Incident Service server exited gracefully")
}

// setupModernizedRoutes configures all the routes with improved structure and security
func setupModernizedRoutes(router *gin.Engine, handler *handlers.IncidentHandler, templateHandler interface{}, resilienceConfig *resilience.Config) {
	// Health check endpoints (excluded from auth and rate limiting)
	health := router.Group("/health")
	{
		health.GET("", handler.Health)
	}

	// Metrics endpoint (excluded from auth)
	if resilienceConfig.Monitoring.MetricsEnabled {
	}

	// API routes with versioning
	api := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		public := api.Group("/public")
		{
			public.GET("/incidents", handler.GetPublicIncidents)
			public.GET("/incidents/:id", handler.GetPublicIncident)
		}

		// Protected routes (authentication required)
		protected := api.Group("/")

		// Add JWT authentication middleware
		if resilienceConfig.JWT.Secret != "" {
			protected.Use(resilience.AuthMiddleware(resilienceConfig.JWT))
		}

		// Add tenant middleware for multi-tenancy
		protected.Use(resilience.TenantMiddleware())

		{
			// Incident management routes (only including existing handlers)
			incidents := protected.Group("/incidents")
			{
				incidents.GET("", handler.GetIncidents)
				incidents.POST("", handler.CreateIncident)
				incidents.GET("/:id", handler.GetIncident)
				incidents.PUT("/:id", handler.UpdateIncident)
				incidents.DELETE("/:id", handler.DeleteIncident)

				// Incident updates (using existing handlers)
				incidents.POST("/:id/updates", handler.AddIncidentUpdate)
				incidents.PUT("/:id/updates/:update_id", handler.UpdateIncidentUpdate)
				incidents.DELETE("/:id/updates/:update_id", handler.DeleteIncidentUpdate)
			}

			// Incident templates management
			templates := protected.Group("/templates")
			{
				templates.GET("", templateHandler.GetTemplates)
				templates.POST("", templateHandler.CreateTemplate)
				templates.GET("/:id", templateHandler.GetTemplate)
				templates.PUT("/:id", templateHandler.UpdateTemplate)
				templates.DELETE("/:id", templateHandler.DeleteTemplate)
				templates.POST("/:id/clone", templateHandler.CloneTemplate)
				templates.POST("/:id/create-incident", templateHandler.CreateIncidentFromTemplate)

				// Workflow step management
				templates.POST("/:template_id/steps", templateHandler.CreateWorkflowStep)
				templates.PUT("/steps/:step_id", templateHandler.UpdateWorkflowStep)
				templates.DELETE("/steps/:step_id", templateHandler.DeleteWorkflowStep)
				templates.PUT("/:template_id/steps/reorder", templateHandler.ReorderWorkflowSteps)
			}

			// Workflow execution management
			workflows := protected.Group("/workflows")
			{
				workflows.GET("/incidents/:incident_id/executions", templateHandler.GetStepExecutions)
				workflows.POST("/executions/:execution_id/execute", templateHandler.ExecuteManualStep)
			}

			// Administrative template functions
			admin := protected.Group("/admin")
			{
				admin.POST("/initialize-templates", templateHandler.InitializeDefaultTemplates)
			}
		}
	}

	// Admin routes with additional security
	admin := router.Group("/admin/v1")
	admin.Use(resilience.AuthMiddleware(resilienceConfig.JWT))
	admin.Use(resilience.TenantMiddleware())
	// admin.Use(middleware.RequireRole("admin")) // Would be implemented
	{
		// Admin routes will be implemented as needed
		// For now, only include basic health endpoint
	}

	// Webhook endpoints for external integrations (to be implemented)
	// webhooks := router.Group("/webhooks")
	// {
	//     // Webhook handlers will be implemented as needed
	// }
}
// TODO: CLEANUP - Update auth middleware usage
// Replace local auth with: auth.NewMiddleware(authConfig, logger)
// Import: github.com/anupamdutta5/shared-resilience/auth
