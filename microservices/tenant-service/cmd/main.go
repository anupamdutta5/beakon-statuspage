// Package main is the entry point for the Tenant Service.
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

	"github.com/enterprise-status/statuspage-tenant-service/internal/config"
	"github.com/enterprise-status/statuspage-tenant-service/internal/handlers"
	"github.com/enterprise-status/statuspage-tenant-service/internal/middleware"
	"github.com/enterprise-status/statuspage-tenant-service/internal/services"
	"github.com/enterprise-status/statuspage-tenant-service/pkg/logger"
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

	logger.Info("Starting Tenant Service",
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
	tenantService := services.NewTenantService(db, logger.Logger)

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger(logger.Logger))
	router.Use(middleware.Recovery(logger.Logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// Initialize handlers
	tenantHandler := handlers.NewTenantHandler(tenantService, logger.Logger)

	// Setup routes
	setupRoutes(router, tenantHandler, cfg)

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
		logger.Info("Tenant Service server starting", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Tenant Service server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Tenant Service server exited")
}

// setupRoutes configures all the routes for the Tenant Service.
func setupRoutes(router *gin.Engine, handler *handlers.TenantHandler, cfg *config.Config) {
	// Health check endpoint
	router.GET("/health", handler.Health)

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		public := api.Group("/public")
		{
			public.GET("/tenants/:slug", handler.GetTenantBySlug)
		}

		// Protected routes (authentication required)
		protected := api.Group("/")
		protected.Use(middleware.Auth(cfg.JWT.Secret))
		{
			// Tenant management routes
			tenants := protected.Group("/tenants")
			{
				tenants.GET("", handler.GetTenants)
				tenants.POST("", handler.CreateTenant)
				tenants.GET("/:id", handler.GetTenant)
				tenants.PUT("/:id", handler.UpdateTenant)
				tenants.DELETE("/:id", handler.DeleteTenant)
				tenants.GET("/:id/settings", handler.GetTenantSettings)
				tenants.PUT("/:id/settings", handler.UpdateTenantSettings)
				tenants.GET("/:id/billing", handler.GetTenantBilling)
				tenants.PUT("/:id/billing", handler.UpdateTenantBilling)
			}
		}
	}
}
