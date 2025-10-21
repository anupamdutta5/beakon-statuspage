// Package server provides the Branding Service server implementation.
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/anupamdutta5/branding-service/internal/config"
	"github.com/anupamdutta5/branding-service/internal/handlers"
	"github.com/anupamdutta5/branding-service/internal/middleware"
	"github.com/anupamdutta5/branding-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server represents the Branding Service server.
type Server struct {
	config  *config.Config
	logger  *zap.Logger
	router  *gin.Engine
	server  *http.Server
	service *services.BrandingService
}

// New creates a new Branding Service server.
func New(cfg *config.Config, logger *zap.Logger) (*Server, error) {
	// Initialize branding service
	service, err := services.NewBrandingService(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize branding service: %w", err)
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(cors.Default())
	router.Use(gin.Logger())

	// Initialize handlers
	handler := handlers.NewBrandingHandler(service, logger)

	// Setup routes
	setupRoutes(router, handler)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	return &Server{
		config:  cfg,
		logger:  logger,
		router:  router,
		server:  server,
		service: service,
	}, nil
}

// Start starts the Branding Service server.
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting Branding Service server",
		zap.String("address", s.server.Addr))

	// Start server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	s.logger.Info("Shutting down Branding Service server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Server shutdown error", zap.Error(err))
		return err
	}

	s.logger.Info("Branding Service server stopped")
	return nil
}

// setupRoutes sets up the API routes.
func setupRoutes(router *gin.Engine, handler *handlers.BrandingHandler) {
	// Health check
	router.GET("/health", handler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Brand operations
		v1.GET("/brands", handler.ListBrands)
		v1.POST("/brands", handler.CreateBrand)
		v1.GET("/brands/:id", handler.GetBrand)
		v1.PUT("/brands/:id", handler.UpdateBrand)
		v1.DELETE("/brands/:id", handler.DeleteBrand)
		v1.GET("/brands/slug/:slug", handler.GetBrandBySlug)

		// Theme operations
		v1.GET("/brands/:brand_id/themes", handler.ListThemes)
		v1.POST("/brands/:brand_id/themes", handler.CreateTheme)
		v1.GET("/themes/:id", handler.GetTheme)
		v1.PUT("/themes/:id", handler.UpdateTheme)
		v1.DELETE("/themes/:id", handler.DeleteTheme)

		// Color scheme operations
		v1.GET("/themes/:theme_id/color-schemes", handler.ListColorSchemes)
		v1.POST("/themes/:theme_id/color-schemes", handler.CreateColorScheme)
		v1.GET("/color-schemes/:id", handler.GetColorScheme)
		v1.PUT("/color-schemes/:id", handler.UpdateColorScheme)
		v1.DELETE("/color-schemes/:id", handler.DeleteColorScheme)

		// Typography operations
		v1.GET("/themes/:theme_id/typographies", handler.ListTypographies)
		v1.POST("/themes/:theme_id/typographies", handler.CreateTypography)
		v1.GET("/typographies/:id", handler.GetTypography)
		v1.PUT("/typographies/:id", handler.UpdateTypography)
		v1.DELETE("/typographies/:id", handler.DeleteTypography)

		// Asset operations
		v1.GET("/brands/:brand_id/assets", handler.ListAssets)
		v1.POST("/brands/:brand_id/assets", handler.CreateAsset)
		v1.GET("/assets/:id", handler.GetAsset)
		v1.PUT("/assets/:id", handler.UpdateAsset)
		v1.DELETE("/assets/:id", handler.DeleteAsset)
		v1.POST("/brands/:brand_id/assets/upload", handler.UploadAsset)

		// Custom CSS operations
		v1.GET("/brands/:brand_id/custom-css", handler.ListCustomCSS)
		v1.POST("/brands/:brand_id/custom-css", handler.CreateCustomCSS)
		v1.GET("/custom-css/:id", handler.GetCustomCSS)
		v1.PUT("/custom-css/:id", handler.UpdateCustomCSS)
		v1.DELETE("/custom-css/:id", handler.DeleteCustomCSS)

		// Custom JS operations
		v1.GET("/brands/:brand_id/custom-js", handler.ListCustomJS)
		v1.POST("/brands/:brand_id/custom-js", handler.CreateCustomJS)
		v1.GET("/custom-js/:id", handler.GetCustomJS)
		v1.PUT("/custom-js/:id", handler.UpdateCustomJS)
		v1.DELETE("/custom-js/:id", handler.DeleteCustomJS)

		// Layout operations
		v1.GET("/themes/:theme_id/layouts", handler.ListLayouts)
		v1.POST("/themes/:theme_id/layouts", handler.CreateLayout)
		v1.GET("/layouts/:id", handler.GetLayout)
		v1.PUT("/layouts/:id", handler.UpdateLayout)
		v1.DELETE("/layouts/:id", handler.DeleteLayout)

		// Component operations
		v1.GET("/themes/:theme_id/components", handler.ListComponents)
		v1.POST("/themes/:theme_id/components", handler.CreateComponent)
		v1.GET("/components/:id", handler.GetComponent)
		v1.PUT("/components/:id", handler.UpdateComponent)
		v1.DELETE("/components/:id", handler.DeleteComponent)

		// Statistics
		v1.GET("/stats", handler.GetStats)
		v1.GET("/brands/:brand_id/stats", handler.GetBrandStats)
		v1.GET("/themes/:theme_id/stats", handler.GetThemeStats)

		// Public endpoints
		v1.GET("/public/brands/:slug", handler.GetPublicBrand)
		v1.GET("/public/themes/:id", handler.GetPublicTheme)
		v1.GET("/public/assets/:id", handler.GetPublicAsset)
	}
}

