// Package main is the entry point for the Branding Service.
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
	"github.com/anupamdutta5/statuspage-branding-service/internal/handlers"
	"github.com/anupamdutta5/statuspage-branding-service/internal/services"
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

	logger.Info("Starting Branding Service",
		zap.String("service", "branding-service"),
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
	brandingService, err := services.NewBrandingService(nil, logger)
	if err != nil {
		logger.Fatal("Failed to initialize branding service", zap.Error(err))
	}

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
	brandingHandler := handlers.NewBrandingHandler(brandingService, logger)

	// Setup basic routes for now
	setupBasicRoutes(router, brandingHandler, config)

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
		logger.Info("Branding Service server starting",
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

	logger.Info("Shutting down Branding Service server...")

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

	logger.Info("Branding Service server exited gracefully")
}


// setupBasicRoutes configures basic routes for the branding service
func setupBasicRoutes(router *gin.Engine, brandingHandler *handlers.BrandingHandler, config *resilience.Config) {
	// Health check endpoints
	health := router.Group("/health")
	{
		health.GET("", brandingHandler.HealthCheck)
	}

	// Basic API routes for brands
	api := router.Group("/api/v1")
	{
		// Protected routes (authentication required)
		protected := api.Group("/")

		// Add JWT authentication middleware
		if config.JWT.Secret != "" {
			protected.Use(resilience.AuthMiddleware(config.JWT))
		}

		// Add tenant middleware for multi-tenancy
		protected.Use(resilience.TenantMiddleware())

		{
			// Basic Brand Management - only implement methods that exist
			brands := protected.Group("/brands")
			{
				// TODO: Add brand management routes when handlers are implemented
				_ = brands // Prevent unused variable error
			}
		}
	}
}
