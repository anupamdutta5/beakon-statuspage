// Package main is the entry point for the uranding Service.
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

	resilience "github.com/anupamdutta5/shared-resilience"
	"github.com/anupamdutta5/branding-service/internal/handlers"
	"github.com/anupamdutta5/branding-service/internal/services"
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

// urandingServiceConfig is the complete configuration for this service
type urandingServiceConfig struct {
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
	var cfg urandingServiceConfig
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

	logger.Info("Starting uranding Service (v2.0)",
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
		Subsystem:   "branding",
		Enabled:     cfg.SharedConfig.Monitoring.MetricsEnabled,
		Registry:    registry,
	})
	if err != nil {
		logger.Fatal("Failed to initialize metrics", zap.Error(err))
	}
	logger.Info("Metrics initialized with custom registry")

	// Log metrics for debugging (can be used later if needed)
	_ = metrics

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

	// Log retry manager for debugging (can be used later if needed)
	_ = retryManager
	_ = serviceClient

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

	// ========================================
	// STEP 9: Initialize Services
	// ========================================
	db := dbManager.GetDB()
	brandingService := services.NewBrandingService(db, logger)

	logger.Info("Branding service initialized")

	// ========================================
	// STEP 10: Initialize Handlers
	// ========================================
	brandingHandler := handlers.NewBrandingHandler(brandingService, logger)

	logger.Info("Handlers initialized")

	// ========================================
	// STEP 11: Setup Routes
	// ========================================

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		dbHealthChecker := resilience.NewDatabaseHealthChecker(dbManager)
		health := dbHealthChecker.Check(c.Request.Context())
		c.JSON(http.StatusOK, health)
	})

	// Prometheus metrics endpoint (with custom registry)
	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))

	// API routes
	api := router.Group("/api/v1")
	{
		// Public endpoints (no authentication required)
		public := api.Group("/public")
		{
			public.GET("/brands/:slug", brandingHandler.GetPublicBrand)
			public.GET("/themes/:id", brandingHandler.GetPublicTheme)
			public.GET("/assets/:id", brandingHandler.GetPublicAsset)
		}

		// Protected routes (require authentication)
		protected := api.Group("")
		protected.Use(resilience.AuthMiddleware(cfg.SharedConfig.JWT))
		protected.Use(resilience.TenantMiddleware())
		{
			// Brand management
			protected.GET("/brands", brandingHandler.ListBrands)
			protected.POST("/brands", brandingHandler.CreateBrand)
			protected.GET("/brands/:id", brandingHandler.GetBrand)
			protected.GET("/brands/slug/:slug", brandingHandler.GetBrandBySlug)
			protected.PUT("/brands/:id", brandingHandler.UpdateBrand)
			protected.DELETE("/brands/:id", brandingHandler.DeleteBrand)
			protected.GET("/brands/:id/stats", brandingHandler.GetBrandStats)

			// Theme management
			protected.GET("/themes", brandingHandler.ListThemes)
			protected.POST("/themes", brandingHandler.CreateTheme)
			protected.GET("/themes/:id", brandingHandler.GetTheme)
			protected.PUT("/themes/:id", brandingHandler.UpdateTheme)
			protected.DELETE("/themes/:id", brandingHandler.DeleteTheme)
			protected.GET("/themes/:id/stats", brandingHandler.GetThemeStats)
			protected.POST("/themes/:id/compile", brandingHandler.CompileTheme)

			// Asset management
			protected.GET("/assets", brandingHandler.ListAssets)
			protected.POST("/assets", brandingHandler.CreateAsset)
			protected.POST("/assets/upload", brandingHandler.UploadAsset)
			protected.GET("/assets/:id", brandingHandler.GetAsset)
			protected.PUT("/assets/:id", brandingHandler.UpdateAsset)
			protected.DELETE("/assets/:id", brandingHandler.DeleteAsset)

			// Custom CSS management
			protected.GET("/custom-css", brandingHandler.ListCustomCSS)
			protected.POST("/custom-css", brandingHandler.CreateCustomCSS)
			protected.GET("/custom-css/:id", brandingHandler.GetCustomCSS)
			protected.PUT("/custom-css/:id", brandingHandler.UpdateCustomCSS)
			protected.DELETE("/custom-css/:id", brandingHandler.DeleteCustomCSS)

			// Color scheme management
			protected.GET("/color-schemes", brandingHandler.ListColorSchemes)
			protected.POST("/color-schemes", brandingHandler.CreateColorScheme)
			protected.GET("/color-schemes/:id", brandingHandler.GetColorScheme)
			protected.PUT("/color-schemes/:id", brandingHandler.UpdateColorScheme)
			protected.DELETE("/color-schemes/:id", brandingHandler.DeleteColorScheme)

			// Typography management
			protected.GET("/typographies", brandingHandler.ListTypographies)
			protected.POST("/typographies", brandingHandler.CreateTypography)
			protected.GET("/typographies/:id", brandingHandler.GetTypography)
			protected.PUT("/typographies/:id", brandingHandler.UpdateTypography)
			protected.DELETE("/typographies/:id", brandingHandler.DeleteTypography)

			// Custom JS management
			protected.GET("/custom-js", brandingHandler.ListCustomJS)
			protected.POST("/custom-js", brandingHandler.CreateCustomJS)
			protected.GET("/custom-js/:id", brandingHandler.GetCustomJS)
			protected.PUT("/custom-js/:id", brandingHandler.UpdateCustomJS)
			protected.DELETE("/custom-js/:id", brandingHandler.DeleteCustomJS)

			// Layout management
			protected.GET("/layouts", brandingHandler.ListLayouts)
			protected.POST("/layouts", brandingHandler.CreateLayout)
			protected.GET("/layouts/:id", brandingHandler.GetLayout)
			protected.PUT("/layouts/:id", brandingHandler.UpdateLayout)
			protected.DELETE("/layouts/:id", brandingHandler.DeleteLayout)

			// Component management
			protected.GET("/components", brandingHandler.ListComponents)
			protected.POST("/components", brandingHandler.CreateComponent)
			protected.GET("/components/:id", brandingHandler.GetComponent)
			protected.PUT("/components/:id", brandingHandler.UpdateComponent)
			protected.DELETE("/components/:id", brandingHandler.DeleteComponent)

			// Stats
			protected.GET("/stats", brandingHandler.GetStats)
			protected.GET("/tenant-stats", brandingHandler.GetTenantStats)
		}
	}

	logger.Info("Routes registered successfully")

	// ========================================
	// STEP 12: Create HTTP Server
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
		logger.Info("uranding Service server starting",
			zap.String("addr", server.Addr),
			zap.Duration("read_timeout", cfg.SharedConfig.Server.ReadTimeout),
			zap.Duration("write_timeout", cfg.SharedConfig.Server.WriteTimeout),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// ========================================
	// STEP 13: Setup Health Check Monitoring
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

		logger.Info("Health check monitoring started")
	}

	// ========================================
	// STEP 14: Graceful Shutdown
	// ========================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down uranding Service server...")

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

	logger.Info("uranding Service server exited gracefully")
}
