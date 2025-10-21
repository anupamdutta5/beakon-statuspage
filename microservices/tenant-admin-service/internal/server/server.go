// Package server provides the Tenant Admin Service server implementation.
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/anupamdutta5/tenant-admin-service/internal/config"
	"github.com/anupamdutta5/tenant-admin-service/internal/handlers"
	"github.com/anupamdutta5/tenant-admin-service/internal/middleware"
	"github.com/anupamdutta5/tenant-admin-service/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server represents the Tenant Admin Service server.
type Server struct {
	config            *config.Config
	logger            *zap.Logger
	router            *gin.Engine
	server            *http.Server
	service           *services.TenantAdminService
	statusPageService *services.StatusPageManagementService
}

// New creates a new Tenant Admin Service server.
func New(cfg *config.Config, logger *zap.Logger) (*Server, error) {
	// Initialize tenant admin service
	service, err := services.NewTenantAdminService(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tenant admin service: %w", err)
	}

	// Initialize status page management service
	statusPageService := services.NewStatusPageManagementService(service.GetDB(), logger, &services.StatusPageConfig{
		ComponentServiceURL:    cfg.Services.ComponentServiceURL,
		IncidentServiceURL:     cfg.Services.IncidentServiceURL,
		MonitoringServiceURL:   cfg.Services.MonitoringServiceURL,
		NotificationServiceURL: cfg.Services.NotificationServiceURL,
		BrandingServiceURL:     cfg.Services.BrandingServiceURL,
	})

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Load HTML templates
	router.LoadHTMLGlob("web/templates/*")

	// Add middleware
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(cors.Default())
	router.Use(gin.Logger())

	// Add tenant context middleware for subdomain routing
	router.Use(middleware.TenantContextMiddleware(service.GetDB(), logger, "localhost"))

	// Initialize handlers
	handler := handlers.NewTenantAdminHandler(service, statusPageService, logger)

	// Initialize auth middleware
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-jwt-secret-change-in-production"
	}
	authMiddleware := middleware.NewAuthMiddleware(jwtSecret, logger)

	// Setup routes
	setupRoutes(router, handler, authMiddleware)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	return &Server{
		config:            cfg,
		logger:            logger,
		router:            router,
		server:            server,
		service:           service,
		statusPageService: statusPageService,
	}, nil
}

// Start starts the Tenant Admin Service server.
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting Tenant Admin Service server",
		zap.String("address", s.server.Addr))

	// Start server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	s.logger.Info("Shutting down Tenant Admin Service server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Server shutdown error", zap.Error(err))
		return err
	}

	s.logger.Info("Tenant Admin Service server stopped")
	return nil
}

// setupRoutes sets up the API routes.
func setupRoutes(router *gin.Engine, handler *handlers.TenantAdminHandler, authMiddleware *middleware.AuthMiddleware) {
	// Health check
	router.GET("/health", handler.HealthCheck)

	// Authentication routes (public)
	router.GET("/login", handler.GetLoginPage)
	router.POST("/api/v1/auth/login", handler.Login)
	router.POST("/api/v1/auth/logout", handler.Logout)
	router.GET("/api/v1/auth/verify", handler.VerifyToken)

	// Public API routes for service-to-service communication
	publicAPI := router.Group("/api/v1/public")
	{
		// Tenant management (for SaaS-Admin service calls)
		publicAPI.GET("/tenants", handler.GetTenants)
		publicAPI.POST("/tenants", handler.CreateTenant)
		publicAPI.GET("/tenants/:id", handler.GetTenant)
		publicAPI.PUT("/tenants/:id", handler.UpdateTenant)
		publicAPI.DELETE("/tenants/:id", handler.DeleteTenant)
	}

	// Admin dashboard (protected)
	router.GET("/admin", authMiddleware.RequireAuth(), handler.GetAdminDashboard)
	router.GET("/", handler.GetLoginPage) // Redirect to login instead of dashboard

	// API v1 routes (protected)
	v1 := router.Group("/api/v1")
	v1.Use(authMiddleware.RequireAuth())
	{
		// Tenant admin management
		v1.GET("/tenants/:tenant_id/admins", handler.ListTenantAdmins)
		v1.POST("/tenants/:tenant_id/admins", handler.CreateTenantAdmin)
		v1.GET("/tenant-admins/:id", handler.GetTenantAdmin)
		v1.PUT("/tenant-admins/:id", handler.UpdateTenantAdmin)
		v1.DELETE("/tenant-admins/:id", handler.DeleteTenantAdmin)

		// Tenant settings management
		v1.GET("/tenants/:tenant_id/settings", handler.GetTenantSettings)
		v1.PUT("/tenants/:tenant_id/settings", handler.UpdateTenantSettings)

		// Tenant feature flag management
		v1.GET("/tenants/:tenant_id/feature-flags", handler.ListTenantFeatureFlags)
		v1.POST("/tenants/:tenant_id/feature-flags", handler.CreateTenantFeatureFlag)
		v1.GET("/tenant-feature-flags/:id", handler.GetTenantFeatureFlag)
		v1.PUT("/tenant-feature-flags/:id", handler.UpdateTenantFeatureFlag)
		v1.DELETE("/tenant-feature-flags/:id", handler.DeleteTenantFeatureFlag)

		// Tenant usage management
		v1.GET("/tenants/:tenant_id/usage", handler.GetTenantUsage)
		v1.POST("/tenants/:tenant_id/usage", handler.RecordTenantUsage)

		// Tenant statistics
		v1.GET("/tenants/:tenant_id/stats", handler.GetTenantStats)

		// Tenant billing management
		v1.GET("/tenants/:tenant_id/billing", handler.GetTenantBilling)
		v1.PUT("/tenants/:tenant_id/billing", handler.UpdateTenantBilling)

		// Tenant notification management
		v1.GET("/tenants/:tenant_id/notifications", handler.ListTenantNotifications)
		v1.POST("/tenants/:tenant_id/notifications", handler.CreateTenantNotification)
		v1.GET("/tenant-notifications/:id", handler.GetTenantNotification)
		v1.PUT("/tenant-notifications/:id", handler.UpdateTenantNotification)
		v1.DELETE("/tenant-notifications/:id", handler.DeleteTenantNotification)

		// Tenant activity logs
		v1.GET("/tenants/:tenant_id/activities", handler.ListTenantActivities)

		// Tenant backup management
		v1.GET("/tenants/:tenant_id/backups", handler.ListTenantBackups)
		v1.POST("/tenants/:tenant_id/backups", handler.CreateTenantBackup)
		v1.GET("/tenant-backups/:id", handler.GetTenantBackup)
		v1.DELETE("/tenant-backups/:id", handler.DeleteTenantBackup)

		// Status page management
		v1.GET("/status-pages/:slug/data", handler.GetStatusPageData)
		v1.GET("/status-pages", handler.GetStatusPages)
		v1.POST("/status-pages", handler.CreateStatusPage)
		v1.PUT("/status-pages/:id", handler.UpdateStatusPage)
		v1.DELETE("/status-pages/:id", handler.DeleteStatusPage)
		v1.PUT("/status-pages/:id/config", handler.UpdateStatusPageConfig)
	}
}
