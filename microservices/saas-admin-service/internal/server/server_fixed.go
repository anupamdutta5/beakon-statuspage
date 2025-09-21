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
	"github.com/anupamdutta5/saas-admin-service/internal/middleware"
	"github.com/anupamdutta5/saas-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server represents the SaaS Admin Service server with proper database management.
type Server struct {
	config    *config.Config
	logger    *zap.Logger
	router    *gin.Engine
	server    *http.Server
	service   *services.SaaSAdminService
	dbManager *database.Manager
}

// New creates a new SaaS Admin Service server with proper database lifecycle management.
func New(cfg *config.Config, logger *zap.Logger) (*Server, error) {
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

	// Create database configuration from service config
	dbConfig := database.Config{
		Host:                cfg.Database.Host,
		Port:                cfg.Database.Port,
		User:                cfg.Database.User,
		Password:            cfg.Database.Password,
		Name:                cfg.Database.Name,
		SSLMode:             cfg.Database.SSLMode,
		MaxOpenConns:        cfg.Database.MaxConns,
		MaxIdleConns:        cfg.Database.MaxIdle,
		MaxLifetime:         cfg.Database.MaxLifetime,
		ConnectTimeout:      10,                // Default timeout
		HealthCheckInterval: 30,               // Default health check interval
	}

	// Initialize database manager with proper lifecycle management
	dbManager, err := database.NewManager(dbConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database manager: %w", err)
	}

	// Create service with database manager
	service, err := services.NewSaaSAdminService(cfg, logger, dbManager.GetDB())
	if err != nil {
		dbManager.Close() // Clean up database connection on service creation failure
		return nil, fmt.Errorf("failed to initialize service: %w", err)
	}

	// Initialize handlers
	adminHandler := handlers.NewSaaSAdminHandler(service, logger)

	// Create server address from host and port
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	// Create server instance
	s := &Server{
		config:    cfg,
		logger:    logger,
		router:    router,
		server: &http.Server{
			Addr:         serverAddr,
			Handler:      router,
			ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
			IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
		},
		service:   service,
		dbManager: dbManager,
	}

	// Register routes
	s.setupRoutes(router, adminHandler)

	logger.Info("SaaS Admin Service server initialized successfully",
		zap.String("address", serverAddr),
		zap.String("environment", cfg.Environment),
	)

	return s, nil
}

// Start starts the SaaS Admin Service server with proper shutdown handling.
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting SaaS Admin Service server",
		zap.String("address", s.server.Addr))

	// Create a channel to capture server errors
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Server error", zap.Error(err))
			serverErrors <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		s.logger.Info("Shutdown signal received, shutting down SaaS Admin Service server...")
	}

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Server forced shutdown", zap.Error(err))
		return err
	}

	// Close database connections
	if err := s.dbManager.Close(); err != nil {
		s.logger.Error("Failed to close database connections", zap.Error(err))
		return err
	}

	s.logger.Info("SaaS Admin Service server stopped gracefully")
	return nil
}

// GetDatabaseStats returns database connection statistics for monitoring
func (s *Server) GetDatabaseStats() interface{} {
	if s.dbManager == nil {
		return nil
	}
	return s.dbManager.GetStats()
}

// GetDatabaseShutdownHandler returns a shutdown handler for database connections
func (s *Server) GetDatabaseShutdownHandler() func(context.Context) error {
	return func(ctx context.Context) error {
		if s.dbManager != nil {
			s.logger.Info("Closing database connections...")
			return s.dbManager.Close()
		}
		return nil
	}
}

// CloseDatabaseConnections closes database connections directly
func (s *Server) CloseDatabaseConnections() error {
	if s.dbManager != nil {
		return s.dbManager.Close()
	}
	return nil
}

// HealthCheck performs a comprehensive health check
func (s *Server) HealthCheck(ctx context.Context) map[string]interface{} {
	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "saas-admin-service",
		"timestamp": time.Now(),
	}

	// Check database health
	if s.dbManager != nil {
		if err := s.dbManager.HealthCheck(ctx); err != nil {
			health["status"] = "unhealthy"
			health["database_error"] = err.Error()
		} else {
			health["database"] = "healthy"
			health["database_stats"] = s.dbManager.GetStats()
		}
	}

	return health
}

// setupRoutes configures all the routes for the server (unchanged from original)
func (s *Server) setupRoutes(router *gin.Engine, adminHandler *handlers.SaaSAdminHandler) {
	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// Health check with enhanced health information
		v1.GET("/health", func(c *gin.Context) {
			health := s.HealthCheck(c.Request.Context())
			statusCode := http.StatusOK
			if status, ok := health["status"].(string); ok && status != "healthy" {
				statusCode = http.StatusServiceUnavailable
			}
			c.JSON(statusCode, health)
		})

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
	}
}