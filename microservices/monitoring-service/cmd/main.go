// Package main is the entry point for the Monitoring Service.
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

	"github.com/enterprise-status/statuspage-monitoring-service/internal/config"
	"github.com/enterprise-status/statuspage-monitoring-service/internal/handlers"
	"github.com/enterprise-status/statuspage-monitoring-service/internal/middleware"
	"github.com/enterprise-status/statuspage-monitoring-service/internal/services"
	"github.com/enterprise-status/statuspage-monitoring-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := logger.New(cfg.Server.Environment)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Monitoring Service",
		zap.String("service", cfg.Service.Name),
		zap.String("version", cfg.Service.Version),
		zap.Int("port", cfg.Server.Port),
		zap.String("environment", cfg.Server.Environment))

	// Set Gin mode
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db, err := services.InitDatabase(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Initialize services
	monitoringService := services.NewMonitoringService(db, logger.Logger)

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger(logger.Logger))
	router.Use(middleware.Recovery(logger.Logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// Initialize handlers
	monitoringHandler := handlers.NewMonitoringHandler(monitoringService, logger.Logger)

	// Setup routes
	setupRoutes(router, monitoringHandler, cfg)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Monitoring Service server starting", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Monitoring Service server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Monitoring Service server exited")
}

// setupRoutes configures all the routes for the Monitoring Service.
func setupRoutes(router *gin.Engine, handler *handlers.MonitoringHandler, cfg *config.Config) {
	// Health check endpoint
	router.GET("/health", handler.Health)

	// Metrics endpoint for Prometheus
	router.GET("/metrics", handler.Metrics)

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		public := api.Group("/public")
		{
			public.GET("/status", handler.GetPublicStatus)
			public.GET("/health", handler.GetPublicHealth)
		}

		// Protected routes (authentication required)
		protected := api.Group("/")
		protected.Use(middleware.Auth(cfg.JWT.Secret))
		{
			// Monitoring and health routes
			monitoring := protected.Group("/monitoring")
			{
				monitoring.GET("/overview", handler.GetMonitoringOverview)
				monitoring.GET("/services", handler.GetServices)
				monitoring.GET("/services/:id", handler.GetService)
				monitoring.POST("/services", handler.CreateService)
				monitoring.PUT("/services/:id", handler.UpdateService)
				monitoring.DELETE("/services/:id", handler.DeleteService)
				monitoring.GET("/services/:id/health", handler.GetServiceHealth)
				monitoring.GET("/services/:id/metrics", handler.GetServiceMetrics)
				monitoring.POST("/services/:id/checks", handler.CreateHealthCheck)
				monitoring.PUT("/services/:id/checks/:check_id", handler.UpdateHealthCheck)
				monitoring.DELETE("/services/:id/checks/:check_id", handler.DeleteHealthCheck)
			}

			// Alert management
			alerts := protected.Group("/alerts")
			{
				alerts.GET("", handler.GetAlerts)
				alerts.POST("", handler.CreateAlert)
				alerts.GET("/:id", handler.GetAlert)
				alerts.PUT("/:id", handler.UpdateAlert)
				alerts.DELETE("/:id", handler.DeleteAlert)
				alerts.POST("/:id/acknowledge", handler.AcknowledgeAlert)
				alerts.POST("/:id/resolve", handler.ResolveAlert)
			}

			// Uptime monitoring
			uptime := protected.Group("/uptime")
			{
				uptime.GET("/checks", handler.GetUptimeChecks)
				uptime.POST("/checks", handler.CreateUptimeCheck)
				uptime.GET("/checks/:id", handler.GetUptimeCheck)
				uptime.PUT("/checks/:id", handler.UpdateUptimeCheck)
				uptime.DELETE("/checks/:id", handler.DeleteUptimeCheck)
				uptime.GET("/checks/:id/results", handler.GetUptimeResults)
				uptime.GET("/checks/:id/statistics", handler.GetUptimeStatistics)
			}

			// Performance monitoring
			performance := protected.Group("/performance")
			{
				performance.GET("/metrics", handler.GetPerformanceMetrics)
				performance.GET("/metrics/:id", handler.GetPerformanceMetric)
				performance.POST("/metrics", handler.CreatePerformanceMetric)
				performance.PUT("/metrics/:id", handler.UpdatePerformanceMetric)
				performance.DELETE("/metrics/:id", handler.DeletePerformanceMetric)
				performance.GET("/metrics/:id/data", handler.GetPerformanceData)
				performance.POST("/metrics/:id/data", handler.AddPerformanceData)
			}

			// Log aggregation
			logs := protected.Group("/logs")
			{
				logs.GET("", handler.GetLogs)
				logs.GET("/search", handler.SearchLogs)
				logs.GET("/aggregate", handler.AggregateLogs)
				logs.POST("/stream", handler.StreamLogs)
			}
		}
	}
}
