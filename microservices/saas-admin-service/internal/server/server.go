// Package server provides the SaaS Admin Service server implementation.
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/enterprise-status/statuspage-saas-admin-service/internal/config"
	"github.com/enterprise-status/statuspage-saas-admin-service/internal/handlers"
	"github.com/enterprise-status/statuspage-saas-admin-service/internal/middleware"
	"github.com/enterprise-status/statuspage-saas-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server represents the SaaS Admin Service server.
type Server struct {
	config  *config.Config
	logger  *zap.Logger
	router  *gin.Engine
	server  *http.Server
	service *services.SaaSAdminService
}

// New creates a new SaaS Admin Service server.
func New(cfg *config.Config, logger *zap.Logger) (*Server, error) {
	// Initialize SaaS admin service
	service, err := services.NewSaaSAdminService(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SaaS admin service: %w", err)
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
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// Initialize handlers
	handler := handlers.NewSaaSAdminHandler(service, logger)

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

// Start starts the SaaS Admin Service server.
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting SaaS Admin Service server",
		zap.String("address", s.server.Addr))

	// Start server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	s.logger.Info("Shutting down SaaS Admin Service server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Server shutdown error", zap.Error(err))
		return err
	}

	s.logger.Info("SaaS Admin Service server stopped")
	return nil
}

// setupRoutes sets up the API routes.
func setupRoutes(router *gin.Engine, handler *handlers.SaaSAdminHandler) {
	// Health check
	router.GET("/health", handler.HealthCheck)

	// Admin interface
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")
	router.GET("/admin", func(c *gin.Context) {
		c.HTML(200, "admin.html", gin.H{})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Platform management
		v1.GET("/platform", handler.GetPlatform)
		v1.PUT("/platform", handler.UpdatePlatform)

		// Plan management
		v1.GET("/plans", handler.ListPlans)
		v1.POST("/plans", handler.CreatePlan)
		v1.GET("/plans/:id", handler.GetPlan)
		v1.PUT("/plans/:id", handler.UpdatePlan)
		v1.DELETE("/plans/:id", handler.DeletePlan)
		v1.GET("/plans/slug/:slug", handler.GetPlanBySlug)

		// Feature management
		v1.GET("/features", handler.ListFeatures)
		v1.POST("/features", handler.CreateFeature)
		v1.GET("/features/:id", handler.GetFeature)
		v1.PUT("/features/:id", handler.UpdateFeature)
		v1.DELETE("/features/:id", handler.DeleteFeature)

		// Feature flag management
		v1.GET("/feature-flags", handler.ListFeatureFlags)
		v1.POST("/feature-flags", handler.CreateFeatureFlag)
		v1.GET("/feature-flags/:id", handler.GetFeatureFlag)
		v1.PUT("/feature-flags/:id", handler.UpdateFeatureFlag)
		v1.DELETE("/feature-flags/:id", handler.DeleteFeatureFlag)

		// Statistics
		v1.GET("/stats", handler.GetStats)

		// Admin user management
		v1.GET("/admin-users", handler.ListAdminUsers)
		v1.POST("/admin-users", handler.CreateAdminUser)
		v1.GET("/admin-users/:id", handler.GetAdminUser)
		v1.PUT("/admin-users/:id", handler.UpdateAdminUser)
		v1.DELETE("/admin-users/:id", handler.DeleteAdminUser)

		// Notification management
		v1.GET("/notifications", handler.ListNotifications)
		v1.POST("/notifications", handler.CreateNotification)
		v1.GET("/notifications/:id", handler.GetNotification)
		v1.PUT("/notifications/:id", handler.UpdateNotification)
		v1.DELETE("/notifications/:id", handler.DeleteNotification)

		// Activity logs
		v1.GET("/activities", handler.ListActivities)

		// Backup management
		v1.GET("/backups", handler.ListBackups)
		v1.POST("/backups", handler.CreateBackup)
		v1.GET("/backups/:id", handler.GetBackup)
		v1.DELETE("/backups/:id", handler.DeleteBackup)

		// Pricing management
		v1.GET("/pricing/features", handler.GetPricingFeatures)
		v1.POST("/pricing/features", handler.CreatePricingFeature)
		v1.PUT("/pricing/features/:id", handler.UpdatePricingFeature)
		v1.DELETE("/pricing/features/:id", handler.DeletePricingFeature)

		v1.GET("/pricing/plans/:planId/tiers", handler.GetPricingTiers)
		v1.POST("/pricing/tiers", handler.CreatePricingTier)
		v1.PUT("/pricing/tiers/:id", handler.UpdatePricingTier)
		v1.DELETE("/pricing/tiers/:id", handler.DeletePricingTier)

		v1.GET("/pricing/plans/:planId/features", handler.GetPlanFeatures)
		v1.POST("/pricing/plans/features/assign", handler.AssignFeatureToPlan)
		v1.POST("/pricing/plans/features/remove", handler.RemoveFeatureFromPlan)

		v1.GET("/pricing/plans/public", handler.GetPublicPricingPlans)
		v1.POST("/pricing/sync", handler.SyncPricingToLandingPage)
	}
}
