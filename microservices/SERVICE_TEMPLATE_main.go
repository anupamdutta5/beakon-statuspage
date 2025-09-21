// SERVICE_TEMPLATE_main.go - Template for microservice main.go with proper error handling
// Replace SERVICE_NAME with your actual service name

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
	"github.com/anupamdutta5/SERVICE_NAME/internal/handlers"
	"github.com/anupamdutta5/SERVICE_NAME/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	ServiceName    = "SERVICE_NAME"
	ServiceVersion = "1.0.0"
)

func main() {
	// Load configuration from environment variables
	config := resilience.LoadConfigFromEnv()

	// Create startup manager for proper error handling
	startupMgr, err := resilience.NewStartupManager(ServiceName, config)
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

	logger.Info("Starting service",
		zap.String("service", ServiceName),
		zap.String("version", ServiceVersion),
		zap.String("environment", config.Environment),
		zap.Int("port", config.Server.Port),
	)

	// Initialize database with retry logic
	var dbManager *resilience.DatabaseManager
	err = startupMgr.InitializeComponent(
		"database",
		func() error {
			var err error
			dbManager, err = resilience.NewDatabaseManager(config.Database, logger)
			return err
		},
		true,  // Critical component
		3,     // Max retries
	)
	if err != nil {
		startupMgr.HandleStartupError(err)
		return
	}
	defer dbManager.Close()

	// Initialize cache (non-critical component)
	var cache *resilience.RedisCache
	if config.Cache.Enabled {
		err = startupMgr.InitializeComponent(
			"cache",
			func() error {
				var err error
				cache, err = resilience.NewRedisCache(config.Cache, logger)
				return err
			},
			false, // Non-critical component
			2,     // Max retries
		)
		if err != nil {
			// Cache is non-critical, so we log and continue
			logger.Warn("Cache initialization failed, continuing without cache",
				zap.Error(err),
			)
		}
		if cache != nil {
			defer cache.Close()
		}
	}

	// Initialize rate limiter (non-critical)
	var rateLimiter *resilience.RateLimiter
	if config.RateLimit.Enabled {
		err = startupMgr.InitializeComponent(
			"rate-limiter",
			func() error {
				var err error
				rateLimiter, err = resilience.NewRateLimiter(config.RateLimit, logger)
				return err
			},
			false, // Non-critical
			2,     // Max retries
		)
		if err != nil {
			logger.Warn("Rate limiter initialization failed, continuing without rate limiting",
				zap.Error(err),
			)
		}
		if rateLimiter != nil {
			defer rateLimiter.Close()
		}
	}

	// Set Gin mode based on environment
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin router
	router := gin.New()

	// Add comprehensive middleware stack
	middleware := resilience.DefaultMiddlewareStack(config, logger)
	for _, mw := range middleware {
		router.Use(mw)
	}

	// Add error handling middleware
	router.Use(resilience.ErrorMiddleware(logger))

	// Initialize services and handlers
	// TODO: Replace with your actual services
	service := services.NewService(dbManager.GetDB(), cache, logger)
	handler := handlers.NewHandler(service, logger)

	// Setup routes
	setupRoutes(router, handler, config)

	// Create HTTP server with proper timeouts
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port),
		Handler:      router,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		IdleTimeout:  config.Server.IdleTimeout,
	}

	// Health check components
	healthChecks := map[string]func() error{
		"database": func() error {
			return dbManager.GetDB().Exec("SELECT 1").Error
		},
	}

	if cache != nil {
		healthChecks["cache"] = func() error {
			return cache.Ping(context.Background())
		}
	}

	// Verify startup health
	health := startupMgr.CheckStartupHealth(healthChecks)
	if health.Status != "healthy" {
		logger.Warn("Service starting with unhealthy components",
			zap.Any("health", health),
		)
	}

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("HTTP server starting",
			zap.String("addr", server.Addr),
			zap.Duration("read_timeout", config.Server.ReadTimeout),
			zap.Duration("write_timeout", config.Server.WriteTimeout),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Setup monitoring (non-critical)
	if config.Monitoring.Enabled {
		startMonitoring(config, dbManager, cache, logger)
	}

	// Setup signal handling for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErrors:
		logger.Error("Server failed to start", zap.Error(err))
		startupMgr.HandleStartupError(err)
		return

	case <-quit:
		logger.Info("Shutdown signal received")
	}

	// Graceful shutdown
	logger.Info("Starting graceful shutdown",
		zap.String("service", ServiceName),
	)

	// Create shutdown context
	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		config.Server.GracefulStop,
	)
	defer shutdownCancel()

	// Shutdown sequence
	startupMgr.GracefulShutdown(
		// Shutdown HTTP server
		func() error {
			return server.Shutdown(shutdownCtx)
		},
		// Close database connections
		func() error {
			return dbManager.Close()
		},
		// Close cache if initialized
		func() error {
			if cache != nil {
				return cache.Close()
			}
			return nil
		},
		// Close rate limiter if initialized
		func() error {
			if rateLimiter != nil {
				return rateLimiter.Close()
			}
			return nil
		},
	)

	logger.Info("Service shutdown completed",
		zap.String("service", ServiceName),
	)
}

func setupRoutes(router *gin.Engine, handler interface{}, config *resilience.Config) {
	// Health check endpoints (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": ServiceName,
			"version": ServiceVersion,
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		// TODO: Add readiness check logic
		c.JSON(http.StatusOK, gin.H{"ready": true})
	})

	// API routes
	api := router.Group("/api/v1")

	// Add authentication middleware if needed
	// api.Use(authMiddleware)

	// TODO: Add your service-specific routes here
	// Example:
	// api.GET("/resources", handler.ListResources)
	// api.POST("/resources", handler.CreateResource)
	// api.GET("/resources/:id", handler.GetResource)
	// api.PUT("/resources/:id", handler.UpdateResource)
	// api.DELETE("/resources/:id", handler.DeleteResource)
}

func startMonitoring(config *resilience.Config, dbManager *resilience.DatabaseManager, cache *resilience.RedisCache, logger *zap.Logger) {
	// Create monitoring goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		dbHealthChecker := resilience.NewDatabaseHealthChecker(dbManager)

		for {
			select {
			case <-ticker.C:
				// Check database health
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				dbHealth := dbHealthChecker.Check(ctx)
				cancel()

				if status, ok := dbHealth["status"].(string); ok && status != "healthy" {
					logger.Warn("Database health check failed", zap.Any("health", dbHealth))
				}

				// Check cache health if enabled
				if cache != nil {
					ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					if err := cache.Ping(ctx); err != nil {
						logger.Warn("Cache health check failed", zap.Error(err))
					}
					cancel()
				}

				// Log service metrics
				logger.Debug("Service health check completed",
					zap.String("service", ServiceName),
					zap.Any("database", dbHealth),
				)
			}
		}
	}()
}