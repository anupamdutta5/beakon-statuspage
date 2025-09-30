// Package server provides the SaaS Admin Service server implementation.
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/anupamdutta5/saas-admin-service/internal/config"
	"github.com/anupamdutta5/saas-admin-service/internal/database"
	"github.com/anupamdutta5/saas-admin-service/internal/handlers"
	"github.com/anupamdutta5/saas-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	// Service will be initialized after database connection is established
	var service *services.SaaSAdminService

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Add middleware
	// router.Use(middleware.Logger(logger))
	// router.Use(middleware.Recovery(logger))
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Initialize database connection
	sqlDB, err := database.InitDatabase(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Defer closing the database connection
	// Database connection will be managed by the Manager lifecycle

	// Create GORM database connection
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create GORM DB: %w", err)
	}

	// Create service with database connection
	service, err = services.NewSaaSAdminService(cfg, logger, db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize service: %w", err)
	}

	// Initialize handlers with service URLs
	serviceURLs := config.ServiceURLs{
		TenantAdminService: fmt.Sprintf("http://localhost:%d", 8099),
		ComponentService:   fmt.Sprintf("http://localhost:%d", 8084),
		IncidentService:    fmt.Sprintf("http://localhost:%d", 8086),
	}
	adminHandler := handlers.NewSaaSAdminHandler(service, serviceURLs, logger)

	// Create server address from host and port
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	// Create server instance
	s := &Server{
		config: cfg,
		logger: logger,
		router: router,
		server: &http.Server{
			Addr:         serverAddr,
			Handler:      router,
			ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
			IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
		},
		service: service,
	}

	// Register routes
	s.setupRoutes(router, adminHandler)

	return s, nil
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

// setupRoutes configures all the routes for the server
func (s *Server) setupRoutes(router *gin.Engine, adminHandler *handlers.SaaSAdminHandler) {
	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", adminHandler.HealthCheck)

		// Platform configuration
		v1.GET("/platform", adminHandler.GetPlatform)
		v1.PUT("/platform", adminHandler.UpdatePlatform)

		// Plan management
		plans := v1.Group("/plans")
		{
			plans.GET("", adminHandler.ListPlans)
			plans.POST("", adminHandler.CreatePlan)
			plans.GET("/:id", adminHandler.GetPlan)
			plans.PUT("/:id", adminHandler.UpdatePlan)
			plans.DELETE("/:id", adminHandler.DeletePlan)
			plans.GET("/slug/:slug", adminHandler.GetPlanBySlug)
		}

		// Feature management
		features := v1.Group("/features")
		{
			features.GET("", adminHandler.ListFeatures)
			features.POST("", adminHandler.CreateFeature)
			features.GET("/:id", adminHandler.GetFeature)
			features.PUT("/:id", adminHandler.UpdateFeature)
			features.DELETE("/:id", adminHandler.DeleteFeature)
		}

		// Feature flag management
		featureFlags := v1.Group("/feature-flags")
		{
			featureFlags.GET("", adminHandler.ListFeatureFlags)
			featureFlags.POST("", adminHandler.CreateFeatureFlag)
			featureFlags.GET("/:id", adminHandler.GetFeatureFlag)
			featureFlags.PUT("/:id", adminHandler.UpdateFeatureFlag)
			featureFlags.DELETE("/:id", adminHandler.DeleteFeatureFlag)
		}

		// Statistics
		v1.GET("/stats", adminHandler.GetStats)

		// Admin user management
		adminUsers := v1.Group("/admin-users")
		{
			adminUsers.GET("", adminHandler.ListAdminUsers)
			adminUsers.POST("", adminHandler.CreateAdminUser)
			adminUsers.GET("/:id", adminHandler.GetAdminUser)
			adminUsers.PUT("/:id", adminHandler.UpdateAdminUser)
			adminUsers.DELETE("/:id", adminHandler.DeleteAdminUser)
		}

		// Notification management
		notifications := v1.Group("/notifications")
		{
			notifications.GET("", adminHandler.ListNotifications)
			notifications.POST("", adminHandler.CreateNotification)
			notifications.GET("/:id", adminHandler.GetNotification)
			notifications.PUT("/:id", adminHandler.UpdateNotification)
			notifications.DELETE("/:id", adminHandler.DeleteNotification)
		}

		// Activity logs
		activities := v1.Group("/activities")
		{
			activities.GET("", adminHandler.ListActivities)
		}

		// Backup management
		backups := v1.Group("/backups")
		{
			backups.GET("", adminHandler.ListBackups)
			backups.POST("", adminHandler.CreateBackup)
			backups.GET("/:id", adminHandler.GetBackup)
			backups.DELETE("/:id", adminHandler.DeleteBackup)
		}

		// Pricing management
		pricingFeatures := v1.Group("/pricing/features")
		{
			pricingFeatures.GET("", adminHandler.GetPricingFeatures)
			pricingFeatures.POST("", adminHandler.CreatePricingFeature)
			pricingFeatures.PUT("/:id", adminHandler.UpdatePricingFeature)
			pricingFeatures.DELETE("/:id", adminHandler.DeletePricingFeature)
		}

		pricingTiers := v1.Group("/pricing/plans/:planId/tiers")
		{
			pricingTiers.GET("", adminHandler.GetPricingTiers)
		}

		tiers := v1.Group("/pricing/tiers")
		{
			tiers.POST("", adminHandler.CreatePricingTier)
				tiers.PUT("/:id", adminHandler.UpdatePricingTier)
			tiers.DELETE("/:id", adminHandler.DeletePricingTier)
		}

		planFeatures := v1.Group("/pricing/plans/:planId/features")
		{
			planFeatures.GET("", adminHandler.GetPlanFeatures)
		}

		planFeaturesActions := v1.Group("/pricing/plans/features")
		{
			planFeaturesActions.POST("/assign", adminHandler.AssignFeatureToPlan)
			planFeaturesActions.POST("/remove", adminHandler.RemoveFeatureFromPlan)
		}

		publicPricing := v1.Group("/pricing/plans")
		{
			publicPricing.GET("/public", adminHandler.GetPublicPricingPlans)
		}

		pricingSync := v1.Group("/pricing")
		{
			pricingSync.POST("/sync", adminHandler.SyncPricingToLandingPage)
		}

		// Tenant management
		tenants := v1.Group("/tenants")
		{
			tenants.GET("", adminHandler.GetTenants)
			tenants.POST("", adminHandler.CreateTenant)
			tenants.GET("/:id", adminHandler.GetTenant)
			tenants.PUT("/:id", adminHandler.UpdateTenant)
			tenants.DELETE("/:id", adminHandler.DeleteTenant)
		}
	}
}
