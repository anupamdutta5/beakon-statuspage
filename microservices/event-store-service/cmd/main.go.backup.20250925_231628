// Package main is the entry point for the Event Store Service.
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
	"github.com/anupamdutta5/event-store-service/internal/config"
	"github.com/anupamdutta5/event-store-service/internal/handlers"
	"github.com/anupamdutta5/event-store-service/internal/services"
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

	logger.Info("Starting Event Store Service",
		zap.String("service", "event-store-service"),
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
	// Create config using internal config package
	serviceConfig, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load service config", zap.Error(err))
	}
	eventStoreService, err := services.NewEventStoreService(serviceConfig, logger)
	if err != nil {
		logger.Fatal("Failed to initialize event store service", zap.Error(err))
	}
	// Set database connection from resilience manager
	eventStoreService.SetDB(dbManager.GetDB())

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
	eventStoreHandler := handlers.NewEventStoreHandler(
		eventStoreService,
		logger,
	)

	// Setup routes with improved structure
	setupModernizedRoutes(router, eventStoreHandler, resilienceConfig)

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
		logger.Info("Event Store Service server starting",
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

	logger.Info("Shutting down Event Store Service server...")

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

	logger.Info("Event Store Service server exited gracefully")
}

// setupModernizedRoutes configures all the routes with improved structure and security
func setupModernizedRoutes(router *gin.Engine, handler *handlers.EventStoreHandler, resilienceConfig *resilience.Config) {
	// Health check endpoints (excluded from auth and rate limiting)
	health := router.Group("/health")
	{
		health.GET("", handler.HealthCheck)
	}

	// Metrics endpoint (excluded from auth)
	if resilienceConfig.Monitoring.MetricsEnabled {
	}

	// API routes with versioning
	api := router.Group("/api/v1")
	{
		// Protected routes (authentication required)
		protected := api.Group("/")

		// Add JWT authentication middleware
		if resilienceConfig.JWT.Secret != "" {
			protected.Use(resilience.AuthMiddleware(resilienceConfig.JWT))
		}

		// Add tenant middleware for multi-tenancy
		protected.Use(resilience.TenantMiddleware())

		{
			// Event streams management (only implementing existing handlers)
			streams := protected.Group("/streams")
			{
				streams.GET("", handler.ListStreams)
				streams.POST("", handler.CreateStream)
				streams.GET("/:id", handler.GetStream)
			}
		}
	}
}