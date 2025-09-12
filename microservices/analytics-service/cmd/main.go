// Package main is the entry point for the Analytics Service.
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

	"github.com/enterprise-status/statuspage-analytics-service/internal/config"
	"github.com/enterprise-status/statuspage-analytics-service/internal/handlers"
	"github.com/enterprise-status/statuspage-analytics-service/internal/middleware"
	"github.com/enterprise-status/statuspage-analytics-service/internal/services"
	"github.com/enterprise-status/statuspage-analytics-service/pkg/logger"
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

	logger.Info("Starting Analytics Service",
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
	analyticsService := services.NewAnalyticsService(db, logger.Logger)

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger(logger.Logger))
	router.Use(middleware.Recovery(logger.Logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// Initialize handlers
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService, logger.Logger)

	// Setup routes
	setupRoutes(router, analyticsHandler, cfg)

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
		logger.Info("Analytics Service server starting", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Analytics Service server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Analytics Service server exited")
}

// setupRoutes configures all the routes for the Analytics Service.
func setupRoutes(router *gin.Engine, handler *handlers.AnalyticsHandler, cfg *config.Config) {
	// Health check endpoint
	router.GET("/health", handler.Health)

	// API routes
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
		protected.Use(middleware.Auth(cfg.JWT.Secret))
		{
			// Analytics and metrics routes
			analytics := protected.Group("/analytics")
			{
				analytics.GET("/overview", handler.GetAnalyticsOverview)
				analytics.GET("/metrics", handler.GetMetrics)
				analytics.GET("/metrics/:id", handler.GetMetric)
				analytics.POST("/metrics", handler.CreateMetric)
				analytics.PUT("/metrics/:id", handler.UpdateMetric)
				analytics.DELETE("/metrics/:id", handler.DeleteMetric)
				analytics.GET("/metrics/:id/data", handler.GetMetricData)
				analytics.POST("/metrics/:id/data", handler.AddMetricData)
			}

			// Reports management
			reports := protected.Group("/reports")
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
			dashboards := protected.Group("/dashboards")
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

			// Data export
			export := protected.Group("/export")
			{
				export.GET("/csv", handler.ExportCSV)
				export.GET("/json", handler.ExportJSON)
				export.GET("/excel", handler.ExportExcel)
			}
		}
	}
}
