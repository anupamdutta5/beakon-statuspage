// Package main is the entry point for the Component Service.
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

	"github.com/enterprise-status/statuspage-component-service/internal/config"
	"github.com/enterprise-status/statuspage-component-service/internal/handlers"
	"github.com/enterprise-status/statuspage-component-service/internal/middleware"
	"github.com/enterprise-status/statuspage-component-service/internal/services"
	"github.com/enterprise-status/statuspage-component-service/pkg/logger"
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

	logger.Info("Starting Component Service",
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
	componentService := services.NewComponentService(db, logger.Logger)

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger(logger.Logger))
	router.Use(middleware.Recovery(logger.Logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// Initialize handlers
	componentHandler := handlers.NewComponentHandler(componentService, logger.Logger)

	// Setup routes
	setupRoutes(router, componentHandler, cfg)

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
		logger.Info("Component Service server starting", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Component Service server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Component Service server exited")
}

// setupRoutes configures all the routes for the Component Service.
func setupRoutes(router *gin.Engine, handler *handlers.ComponentHandler, cfg *config.Config) {
	// Health check endpoint
	router.GET("/health", handler.Health)

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		public := api.Group("/public")
		{
			public.GET("/status", handler.GetPublicStatus)
			public.GET("/components", handler.GetPublicComponents)
			public.GET("/components/:id", handler.GetPublicComponent)
		}

		// Protected routes (authentication required)
		protected := api.Group("/")
		protected.Use(middleware.Auth(cfg.JWT.Secret))
		{
			// Component management routes
			components := protected.Group("/components")
			{
				components.GET("", handler.GetComponents)
				components.POST("", handler.CreateComponent)
				components.GET("/:id", handler.GetComponent)
				components.PUT("/:id", handler.UpdateComponent)
				components.DELETE("/:id", handler.DeleteComponent)
				components.PUT("/:id/status", handler.UpdateComponentStatus)
				components.GET("/:id/history", handler.GetComponentHistory)
			}

			// Component group management
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
