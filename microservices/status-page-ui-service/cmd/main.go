// Package main provides the main entry point for the Status Page UI Service.
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

	"github.com/enterprise-status/statuspage-status-ui-service/internal/config"
	"github.com/enterprise-status/statuspage-status-ui-service/internal/handlers"
	"github.com/enterprise-status/statuspage-status-ui-service/internal/services"
	"github.com/enterprise-status/statuspage-status-ui-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Status Page UI Service")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize services (no database needed - pure frontend service)
	statusPageService := services.NewStatusPageService(logger, &services.Config{
		TenantAdminServiceURL: cfg.Services.TenantAdminService.URL,
	})

	// Initialize handlers
	statusPageHandler := handlers.NewStatusPageHandler(statusPageService, logger)

	// Setup router
	router := setupRouter(statusPageHandler, logger)

	// Start server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// setupRouter configures the HTTP router.
func setupRouter(statusPageHandler *handlers.StatusPageHandler, _ *zap.Logger) *gin.Engine {
	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "status-page-ui-service",
			"timestamp": time.Now().UTC(),
		})
	})

	// Static files
	router.Static("/static", "./web/static")

	// Status page routes
	router.GET("/", statusPageHandler.GetStatusPage)
	router.GET("/status", statusPageHandler.GetStatusPage)
	router.GET("/:slug", statusPageHandler.GetStatusPageBySlug)

	// API routes
	api := router.Group("/api/v1")
	{
		api.GET("/status", statusPageHandler.GetStatusData)
		api.GET("/status/:slug", statusPageHandler.GetStatusDataBySlug)
	}

	return router
}

// corsMiddleware adds CORS headers.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
