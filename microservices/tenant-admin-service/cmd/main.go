// Package main is the entry point for the Tenant Admin Service.
// This is the modernized version using the shared-resilience module.
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

	"github.com/anupamdutta5/shared-resilience"
	"github.com/anupamdutta5/tenant-admin-service/internal/config"
	"github.com/anupamdutta5/tenant-admin-service/internal/handlers"
	"github.com/anupamdutta5/tenant-admin-service/internal/middleware"
	"github.com/anupamdutta5/tenant-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load resilience configuration from environment variables
	resilienceConfig := resilience.LoadConfigFromEnv()
	if err := resilienceConfig.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Load local service configuration
	localConfig, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load local configuration: %v", err)
	}

	// Initialize logger based on environment
	var logger *zap.Logger
	if resilienceConfig.Environment == "production" {
		logger, err = zap.NewProduction()
		gin.SetMode(gin.ReleaseMode)
	} else {
		logger, err = zap.NewDevelopment()
		gin.SetMode(gin.DebugMode)
	}

	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Tenant Admin Service",
		zap.String("service", "tenant-admin-service"),
		zap.String("version", "1.0.0"),
		zap.String("environment", resilienceConfig.Environment),
		zap.Int("port", resilienceConfig.Server.Port),
	)

	// Initialize database manager with connection pooling and health checks
	dbManager, err := resilience.NewDatabaseManager(resilienceConfig.Database, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database manager", zap.Error(err))
	}
	defer dbManager.Close()

	// Test database connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dbManager.HealthCheck(ctx); err != nil {
		logger.Fatal("Database health check failed", zap.Error(err))
	}

	logger.Info("Database connection established successfully")

	// Initialize circuit breakers for external dependencies
	var circuitBreakers = make(map[string]*resilience.CircuitBreaker)

	if resilienceConfig.CircuitBreaker.Database.Enabled {
		circuitBreakers["database"] = resilience.NewCircuitBreaker(
			resilienceConfig.CircuitBreaker.Database.Name,
			resilienceConfig.CircuitBreaker.Database,
			logger,
		)
	}

	if resilienceConfig.CircuitBreaker.External.Enabled {
		circuitBreakers["external"] = resilience.NewCircuitBreaker(
			resilienceConfig.CircuitBreaker.External.Name,
			resilienceConfig.CircuitBreaker.External,
			logger,
		)
	}

	// Initialize cache if enabled
	var cache resilience.Cache
	if resilienceConfig.Cache.Enabled {
		if resilienceConfig.Cache.Type == "redis" {
			cache = resilience.NewRedisCache(resilienceConfig.Redis, resilienceConfig.Cache, logger)
		} else {
			cache = resilience.NewInMemoryCache(resilienceConfig.Cache, logger)
		}
		logger.Info("Cache initialized", zap.String("type", resilienceConfig.Cache.Type))
	}

	// Initialize rate limiter if enabled
	var rateLimiter resilience.RateLimiter
	if resilienceConfig.RateLimit.Enabled {
		if resilienceConfig.Cache.Type == "redis" {
			rateLimiter = resilience.NewRedisRateLimiter(resilienceConfig.Redis, resilienceConfig.RateLimit, logger)
		} else {
			rateLimiter = resilience.NewInMemoryRateLimiter(resilienceConfig.RateLimit, logger)
		}
		logger.Info("Rate limiter initialized")
	}

	// Initialize business services with modernized dependencies
	tenantAdminService, err := services.NewTenantAdminService(localConfig, logger)
	if err != nil {
		logger.Fatal("Failed to initialize tenant admin service", zap.Error(err))
	}

	domainService := services.NewDomainService(logger)

	// Initialize RBAC service
	rbacService := services.NewRBACService(dbManager.GetDB(), logger)

	// Initialize status page service
	statusPageService := services.NewStatusPageManagementService(dbManager.GetDB(), logger, &services.StatusPageConfig{
		ComponentServiceURL:    "http://localhost:8001",
		IncidentServiceURL:     "http://localhost:8002",
		MonitoringServiceURL:   "http://localhost:8003",
		NotificationServiceURL: "http://localhost:8004",
		BrandingServiceURL:     "http://localhost:8005",
	})

	// Create Gin router
	router := gin.New()

	// Add comprehensive middleware stack
	middleware := resilience.DefaultMiddlewareStack(resilienceConfig, logger)
	for _, mw := range middleware {
		router.Use(mw)
	}

	// Add rate limiting middleware if enabled
	if rateLimiter != nil {
		router.Use(resilience.RateLimitMiddleware(resilienceConfig.RateLimit.RequestsPerMinute, time.Minute))
	}

	// Initialize modernized handlers with shared error handling
	tenantAdminHandler := handlers.NewTenantAdminHandler(tenantAdminService, statusPageService, logger)
	domainHandler := handlers.NewDomainHandler(domainService, logger)
	rbacHandler := handlers.NewRBACHandler(rbacService, logger)

	// Setup routes with improved structure
	setupModernizedRoutes(router, tenantAdminHandler, domainHandler, rbacHandler, rbacService, logger, resilienceConfig)

	// Create HTTP server with proper timeouts and configuration
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", resilienceConfig.Server.Host, resilienceConfig.Server.Port),
		Handler:      router,
		ReadTimeout:  resilienceConfig.Server.ReadTimeout,
		WriteTimeout: resilienceConfig.Server.WriteTimeout,
		IdleTimeout:  resilienceConfig.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Tenant Admin Service server starting",
			zap.String("addr", server.Addr),
			zap.Duration("read_timeout", resilienceConfig.Server.ReadTimeout),
			zap.Duration("write_timeout", resilienceConfig.Server.WriteTimeout),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Setup health check monitoring
	if resilienceConfig.Monitoring.Enabled {
		dbHealthChecker := resilience.NewDatabaseHealthChecker(dbManager)

		// Periodically log health status
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					healthCtx, healthCancel := context.WithTimeout(context.Background(), 5*time.Second)
					health := dbHealthChecker.Check(healthCtx)
					healthCancel()

					if status, ok := health["status"].(string); ok && status != "healthy" {
						logger.Warn("Database health check failed", zap.Any("health", health))
					}
				}
			}
		}()
	}

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Tenant Admin Service server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), resilienceConfig.Server.GracefulStop)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	// Close database connections
	if err := dbManager.Close(); err != nil {
		logger.Error("Failed to close database connections", zap.Error(err))
	}

	// Close cache if initialized
	if cache != nil {
		cache.Close()
	}

	// Close rate limiter if initialized
	if rateLimiter != nil {
		rateLimiter.Close()
	}

	logger.Info("Tenant Admin Service server exited gracefully")
}

// setupModernizedRoutes configures all the routes with improved structure and security
func setupModernizedRoutes(router *gin.Engine, tenantHandler *handlers.TenantAdminHandler, domainHandler *handlers.DomainHandler, rbacHandler *handlers.RBACHandler, rbacService *services.RBACService, logger *zap.Logger, config *resilience.Config) {
	// Health check endpoints (excluded from auth and rate limiting)
	router.GET("/health", tenantHandler.HealthCheck)
	router.GET("/health/ready", tenantHandler.HealthCheck)
	router.GET("/health/live", tenantHandler.HealthCheck)

	// API routes with versioning
	api := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		public := api.Group("/public")
		{
			public.POST("/login", tenantHandler.Login)
		}

		// Protected routes (authentication required)
		protected := api.Group("/")

		// Add JWT authentication middleware
		if config.JWT.Secret != "" {
			protected.Use(resilience.AuthMiddleware(config.JWT))
		}

		// Add tenant middleware for multi-tenancy
		protected.Use(resilience.TenantMiddleware())

		// Initialize RBAC middleware
		rbacMiddleware := middleware.NewRBACMiddleware(rbacService, logger)

		// Add session validation middleware
		protected.Use(rbacMiddleware.SessionValidation())

		// Add audit logging middleware
		protected.Use(rbacMiddleware.AuditLogging())

		{
			// Auth routes
			auth := protected.Group("/auth")
			{
				auth.POST("/logout", tenantHandler.Logout)
				auth.POST("/verify", tenantHandler.VerifyToken)
			}

			// Tenant admin routes
			admins := protected.Group("/admins")
			{
				admins.GET("", tenantHandler.ListTenantAdmins)
				admins.POST("", tenantHandler.CreateTenantAdmin)
				admins.GET("/:id", tenantHandler.GetTenantAdmin)
				admins.PUT("/:id", tenantHandler.UpdateTenantAdmin)
				admins.DELETE("/:id", tenantHandler.DeleteTenantAdmin)
			}

			// Status page management routes
			statusPages := protected.Group("/status-pages")
			{
				statusPages.GET("", tenantHandler.GetStatusPages)
				statusPages.POST("", tenantHandler.CreateStatusPage)
				statusPages.GET("/:id", tenantHandler.GetStatusPageData)
				statusPages.PUT("/:id", tenantHandler.UpdateStatusPage)
				statusPages.DELETE("/:id", tenantHandler.DeleteStatusPage)
				statusPages.PUT("/:id/config", tenantHandler.UpdateStatusPageConfig)
			}

			// Domain management routes
			domains := protected.Group("/domains")
			{
				domains.POST("", domainHandler.AddDomain)
				domains.GET("/:domain", domainHandler.GetDomain)
				domains.DELETE("/:domain", domainHandler.DeleteDomain)
				domains.POST("/:domain/verify", domainHandler.VerifyDomain)
			}

			// Settings routes
			settings := protected.Group("/settings")
			{
				settings.GET("", tenantHandler.GetTenantSettings)
				settings.PUT("", tenantHandler.UpdateTenantSettings)
			}

			// Feature flags routes
			features := protected.Group("/feature-flags")
			{
				features.GET("", tenantHandler.ListTenantFeatureFlags)
				features.POST("", tenantHandler.CreateTenantFeatureFlag)
				features.GET("/:id", tenantHandler.GetTenantFeatureFlag)
				features.PUT("/:id", tenantHandler.UpdateTenantFeatureFlag)
				features.DELETE("/:id", tenantHandler.DeleteTenantFeatureFlag)
			}

			// Usage and billing routes
			usage := protected.Group("/usage")
			{
				usage.GET("", tenantHandler.GetTenantUsage)
				usage.POST("", tenantHandler.RecordTenantUsage)
			}

			billing := protected.Group("/billing")
			{
				billing.GET("", tenantHandler.GetTenantBilling)
				billing.PUT("", tenantHandler.UpdateTenantBilling)
			}

			// Notification routes
			notifications := protected.Group("/notifications")
			{
				notifications.GET("", tenantHandler.ListTenantNotifications)
				notifications.POST("", tenantHandler.CreateTenantNotification)
				notifications.GET("/:id", tenantHandler.GetTenantNotification)
				notifications.PUT("/:id", tenantHandler.UpdateTenantNotification)
				notifications.DELETE("/:id", tenantHandler.DeleteTenantNotification)
			}

			// Activity and backup routes
			protected.GET("/activities", tenantHandler.ListTenantActivities)
			protected.GET("/stats", tenantHandler.GetTenantStats)

			backups := protected.Group("/backups")
			{
				backups.GET("", tenantHandler.ListTenantBackups)
				backups.POST("", tenantHandler.CreateTenantBackup)
				backups.GET("/:id", tenantHandler.GetTenantBackup)
				backups.DELETE("/:id", tenantHandler.DeleteTenantBackup)
			}

			// RBAC Management Routes
			rbac := protected.Group("/rbac")
			{
				// Role management
				roles := rbac.Group("/roles")
				roles.Use(rbacMiddleware.RequirePermission("users.read")) // Basic permission check for role management
				{
					roles.GET("", rbacHandler.GetRoles)
					roles.POST("", rbacMiddleware.RequirePermission("users.create"), rbacHandler.CreateRole)
					roles.GET("/:id", rbacHandler.GetRole)
					roles.PUT("/:id", rbacMiddleware.RequirePermission("users.update"), rbacHandler.UpdateRole)
					roles.DELETE("/:id", rbacMiddleware.RequirePermission("users.delete"), rbacHandler.DeleteRole)
					roles.POST("/:id/permissions", rbacMiddleware.RequirePermission("users.update"), rbacHandler.AssignPermissionsToRole)
				}

				// Permission management
				permissions := rbac.Group("/permissions")
				permissions.Use(rbacMiddleware.RequirePermission("users.read"))
				{
					permissions.GET("", rbacHandler.GetPermissions)
				}

				// User role assignments
				userRoles := rbac.Group("/user-roles")
				userRoles.Use(rbacMiddleware.RequirePermission("users.read"))
				{
					userRoles.POST("", rbacMiddleware.RequirePermission("users.update"), rbacHandler.AssignRoleToUser)
					userRoles.DELETE("/:user_id/:role_id", rbacMiddleware.RequirePermission("users.update"), rbacHandler.RemoveRoleFromUser)
					userRoles.GET("/:user_id", rbacHandler.GetUserRoles)
					userRoles.GET("/:user_id/permissions", rbacHandler.GetUserPermissions)
					userRoles.GET("/:user_id/check", rbacHandler.CheckPermission)
				}

				// Team management
				teams := rbac.Group("/teams")
				teams.Use(rbacMiddleware.RequirePermission("users.read"))
				{
					teams.GET("", rbacHandler.GetTeams)
					teams.POST("", rbacMiddleware.RequirePermission("users.create"), rbacHandler.CreateTeam)
					teams.GET("/:id", rbacHandler.GetTeam)
					teams.PUT("/:id", rbacMiddleware.RequirePermission("users.update"), rbacHandler.UpdateTeam)
					teams.DELETE("/:id", rbacMiddleware.RequirePermission("users.delete"), rbacHandler.DeleteTeam)
					teams.POST("/:team_id/members", rbacMiddleware.RequirePermission("users.update"), rbacHandler.AddUserToTeam)
					teams.DELETE("/:team_id/members/:user_id", rbacMiddleware.RequirePermission("users.update"), rbacHandler.RemoveUserFromTeam)
				}

				// Administrative functions
				admin := rbac.Group("/admin")
				admin.Use(rbacMiddleware.RequireRole("admin"))
				{
					admin.POST("/initialize-roles", rbacHandler.InitializeTenantRoles)
					admin.GET("/audit-logs", rbacHandler.GetAuditLogs)
					admin.GET("/sessions", rbacHandler.GetActiveSessions)
				}
			}
		}
	}

	// Static file serving for admin pages
	router.Static("/static", "./web/static")
	router.GET("/", tenantHandler.GetLoginPage)
	router.GET("/login", tenantHandler.GetLoginPage)
	router.GET("/admin", tenantHandler.GetAdminDashboard)
}