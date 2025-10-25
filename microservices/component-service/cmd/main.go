// Package main is the entry point for the Component Service.
// This is the modernized version using the shared-resilience module.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anupamdutta5/shared-resilience"
	"github.com/anupamdutta5/component-service/internal/handlers"
	"github.com/anupamdutta5/component-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load configuration from environment variables
	config := resilience.LoadConfigFromEnv()
	if err := config.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Initialize logger based on environment
	var logger *zap.Logger
	var err error

	if config.Environment == "production" {
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

	logger.Info("Starting Component Service",
		zap.String("service", "component-service"),
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
	componentService := services.NewComponentService(dbManager.GetDB(), logger)

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

	// Initialize modernized handlers
	componentHandler := handlers.NewComponentHandler(componentService, logger)

	// Setup routes with improved structure
	setupModernizedRoutes(router, componentHandler, config)

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
		logger.Info("Component Service server starting",
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

	logger.Info("Shutting down Component Service server...")

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

	logger.Info("Component Service server exited gracefully")
}

// setupModernizedRoutes configures basic routes with existing handlers
func setupModernizedRoutes(router *gin.Engine, handler *handlers.ComponentHandler, config *resilience.Config) {
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
			public.GET("/components", handler.GetPublicComponents)
			public.GET("/components/:id", handler.GetPublicComponent)
			public.GET("/components/:id/status", handler.GetPublicComponentStatus)
			public.GET("/component-groups", handler.GetPublicComponentGroups)
			public.GET("/overall-status", handler.GetOverallStatus)
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
			// Component management routes (only implemented methods)
			components := protected.Group("/components")
			{
				components.GET("", handler.GetComponents)
				components.POST("", handler.CreateComponent)
				components.GET("/:id", handler.GetComponent)
				components.PUT("/:id", handler.UpdateComponent)
				components.DELETE("/:id", handler.DeleteComponent)
				components.PUT("/:id/status", handler.UpdateComponentStatus)
			}

			// Component group management (only implemented methods)
			groups := protected.Group("/component-groups")
			{
				groups.GET("", handler.GetComponentGroups)
				groups.POST("", handler.CreateComponentGroup)
				groups.GET("/:id", handler.GetComponentGroup)
				groups.PUT("/:id", handler.UpdateComponentGroup)
				groups.DELETE("/:id", handler.DeleteComponentGroup)
			}
		}
	}
}
// TODO: CLEANUP - Update auth middleware usage
// Replace local auth with: auth.NewMiddleware(authConfig, logger)
// Import: github.com/anupamdutta5/shared-resilience/auth
