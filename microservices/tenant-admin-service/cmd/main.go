// Package main is the entry point for the Tenant Admin Service.
// This is the v2.0 version using shared-resilience v2.0 primitives.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anupamdutta5/tenant-admin-service/internal/config"
	"github.com/anupamdutta5/tenant-admin-service/internal/events"
	"github.com/anupamdutta5/tenant-admin-service/internal/handlers"
	"github.com/anupamdutta5/tenant-admin-service/internal/middleware"
	"github.com/anupamdutta5/tenant-admin-service/internal/sessions"
	"github.com/anupamdutta5/tenant-admin-service/internal/services"
	resilience "github.com/anupamdutta5/shared-resilience"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	// ========================================
	// STEP 1: Load Configuration from YAML
	// ========================================
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// ========================================
	// STEP 2: Initialize Logger
	// ========================================
	var logger *zap.Logger

	if cfg.Service.Environment == "production" {
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

	logger.Info("Starting Tenant Admin Service (v2.0)",
		zap.String("service", cfg.Service.Name),
		zap.String("version", cfg.Service.Version),
		zap.String("environment", cfg.Service.Environment),
		zap.Int("port", cfg.Server.Port),
		zap.String("shared_resilience", resilience.Version),
	)

	// ========================================
	// STEP 3: Initialize Prometheus Registry
	// ========================================
	registry := prometheus.NewRegistry()
	logger.Info("Prometheus registry initialized (injected, not global)")

	// ========================================
	// STEP 4: Initialize Metrics (with injected registry)
	// ========================================
	metrics, err := resilience.NewMetrics(resilience.MetricsConfig{
		ServiceName: cfg.Service.Name,
		Namespace:   "beakon",
		Subsystem:   "tenant_admin",
		Enabled:     cfg.Monitoring.Enabled,
		Registry:    registry,
	})
	if err != nil {
		logger.Fatal("Failed to initialize metrics", zap.Error(err))
	}
	logger.Info("Metrics initialized with custom registry")

	// ========================================
	// STEP 5: Ensure Database Exists
	// ========================================
	dbConfig := resilience.DatabaseConfig{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Name:            cfg.Database.Name,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxConns,
		MaxIdleConns:    cfg.Database.MinConns,
		ConnMaxLifetime: cfg.Database.GetConnMaxLifetimeDuration(),
		ConnMaxIdleTime: cfg.Database.GetConnMaxIdleTimeDuration(),
	}

	if err := ensureDatabaseExists(dbConfig, logger); err != nil {
		logger.Fatal("Failed to ensure database exists", zap.Error(err))
	}

	// ========================================
	// STEP 6: Initialize Database Manager
	// ========================================
	dbManager, err := resilience.NewDatabaseManager(dbConfig, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database manager", zap.Error(err))
	}
	defer dbManager.Close()
	logger.Info("Database manager initialized",
		zap.String("database", cfg.Database.Name),
		zap.Int("max_open_conns", cfg.Database.MaxConns),
	)

	db := dbManager.GetDB()

	// ========================================
	// STEP 7: Initialize Session Storage
	// ========================================
	var sessionManager *sessions.SessionManager
	var primarySessionStore sessions.SessionStore
	var fallbackSessionStore sessions.SessionStore

	// Try to initialize Redis session store as primary (if enabled)
	if cfg.Redis.Enabled {
		redisConfig := resilience.RedisConfig{
			Host:         cfg.Redis.Host,
			Port:         cfg.Redis.Port,
			Password:     cfg.Redis.Password,
			DB:           cfg.Redis.DB,
			MaxRetries:   cfg.Redis.MaxRetries,
			PoolSize:     cfg.Redis.PoolSize,
			MinIdleConns: cfg.Redis.MinIdleConns,
			KeyPrefix:    "tenant-admin:",
		}

		redisStore, err := sessions.NewRedisSessionStore(redisConfig, logger)
		if err != nil {
			logger.Warn("Failed to initialize Redis session store, falling back to DB only",
				zap.Error(err))
		} else {
			primarySessionStore = redisStore
			logger.Info("Redis session store initialized as primary")
		}
	}

	// Always initialize database session store as fallback
	dbSessionStore := sessions.NewDBSessionStore(db, logger)

	if primarySessionStore != nil {
		// Redis as primary, DB as fallback
		fallbackSessionStore = dbSessionStore
		sessionManager = sessions.NewSessionManager(primarySessionStore, fallbackSessionStore, logger)
		logger.Info("Session manager initialized with Redis primary and database fallback")
	} else {
		// DB as primary, no fallback
		sessionManager = sessions.NewSessionManager(dbSessionStore, nil, logger)
		logger.Info("Session manager initialized with database only")
	}

	// ========================================
	// STEP 8: Initialize RabbitMQ Consumer
	// ========================================
	rabbitmqURL := cfg.GetRabbitMQURL()

	// Initialize tenant admin service (required for event handler)
	tenantAdminService, err := services.NewTenantAdminService(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize tenant admin service", zap.Error(err))
	}

	// Create event handler with proper service injection
	eventHandler := events.NewRabbitMQTenantEventHandler(db, tenantAdminService, logger)

	// Create consumer
	eventConsumer, err := events.NewConsumer(events.ConsumerConfig{
		URL:     rabbitmqURL,
		Handler: eventHandler,
		Logger:  logger,
	})
	if err != nil {
		logger.Warn("Failed to initialize RabbitMQ consumer, tenant events will not be synchronized",
			zap.Error(err))
		eventConsumer = nil
	} else {
		logger.Info("RabbitMQ event consumer initialized successfully",
			zap.String("url", rabbitmqURL))

		// Start consuming events in background
		go func() {
			ctx := context.Background()
			if err := eventConsumer.Start(ctx); err != nil {
				logger.Error("Event consumer stopped with error", zap.Error(err))
			}
		}()
	}

	// ========================================
	// STEP 9: Initialize Redis Cache
	// ========================================
	var cache resilience.Cache
	if cfg.Redis.Enabled {
		redisConfig := resilience.RedisConfig{
			Host:         cfg.Redis.Host,
			Port:         cfg.Redis.Port,
			Password:     cfg.Redis.Password,
			DB:           cfg.Redis.DB,
			MaxRetries:   cfg.Redis.MaxRetries,
			PoolSize:     cfg.Redis.PoolSize,
			MinIdleConns: cfg.Redis.MinIdleConns,
			KeyPrefix:    "tenant-admin:",
		}

		cacheConfig := resilience.CacheConfig{
			Enabled:         true,
			DefaultTTL:      5 * time.Minute,
			MaxSize:         1000,
			CleanupInterval: 10 * time.Minute,
			Type:            "redis",
		}

		cache = resilience.NewRedisCache(redisConfig, cacheConfig, logger)
		logger.Info("Redis cache initialized", zap.String("type", "redis"))
	}

	// ========================================
	// STEP 10: Initialize Services
	// ========================================
	domainService := services.NewDomainService(logger)
	rbacService := services.NewRBACService(db, logger, sessionManager)

	// Status page service
	statusPageService := services.NewStatusPageManagementService(db, logger, &services.StatusPageConfig{
		ComponentServiceURL:    "http://localhost:8001",
		IncidentServiceURL:     "http://localhost:8002",
		MonitoringServiceURL:   "http://localhost:8003",
		NotificationServiceURL: "http://localhost:8004",
		BrandingServiceURL:     "http://localhost:8005",
	})

	// Component, incident, subscriber services with cache
	componentService := services.NewComponentService(db, cache, logger)
	incidentService := services.NewIncidentService(db, cache, logger)
	subscriberService := services.NewSubscriberService(db, cache, logger)
	dependencyService := services.NewDependencyService(db)

	// Session service for refresh token management
	sessionService := services.NewSessionService(db, logger)

	// SAML/SSO service for enterprise authentication
	samlBaseURL := os.Getenv("SAML_BASE_URL")
	if samlBaseURL == "" {
		samlBaseURL = fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
	}

	samlService := services.NewSAMLService(db, tenantAdminService, logger, &services.SAMLConfig{
		BaseURL: samlBaseURL,
	})

	logger.Info("All services initialized successfully")

	// ========================================
	// STEP 11: Initialize Handlers
	// ========================================
	tenantAdminHandler := handlers.NewTenantAdminHandler(tenantAdminService, statusPageService, rbacService, sessionService, logger)
	domainHandler := handlers.NewDomainHandler(domainService, logger)
	rbacHandler := handlers.NewRBACHandler(rbacService, logger)
	userHandler := handlers.NewUserHandler(tenantAdminService, logger)
	componentHandler := handlers.NewComponentHandler(componentService, logger)
	incidentHandler := handlers.NewIncidentHandler(incidentService, logger)
	subscriberHandler := handlers.NewSubscriberHandler(subscriberService, logger)
	dependencyHandler := handlers.NewDependencyHandler(dependencyService)
	samlHandler := handlers.NewSAMLHandler(samlService, tenantAdminService, logger)

	logger.Info("All handlers initialized successfully")

	// ========================================
	// STEP 12: Initialize Gin Router
	// ========================================
	router := gin.New()

	// Add recovery middleware
	router.Use(gin.Recovery())

	// Add custom middleware
	router.Use(middleware.CorrelationIDMiddleware())
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.LoggingMiddleware(logger))

	// CORS middleware from shared-resilience
	corsConfig := resilience.CORSConfig{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}
	router.Use(resilience.CORSMiddleware(corsConfig))

	// Security headers middleware from shared-resilience
	router.Use(resilience.SecurityHeadersMiddleware())

	// Subdomain-based tenant isolation
	baseDomain := os.Getenv("BASE_DOMAIN")
	if baseDomain == "" {
		baseDomain = "localhost"
	}
	router.Use(middleware.TenantContextMiddleware(db, logger, baseDomain))

	logger.Info("Middleware stack applied")

	// ========================================
	// STEP 13: Setup Routes
	// ========================================
	setupRoutes(router, tenantAdminHandler, domainHandler, rbacHandler, userHandler, componentHandler, incidentHandler, subscriberHandler, dependencyHandler, samlHandler, rbacService, logger, cfg, registry)

	logger.Info("Routes registered successfully")

	// ========================================
	// STEP 14: Create HTTP Server
	// ========================================
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.GetReadTimeout(),
		WriteTimeout: cfg.GetWriteTimeout(),
		IdleTimeout:  cfg.GetIdleTimeout(),
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Tenant Admin Service server starting",
			zap.String("addr", server.Addr),
			zap.Duration("read_timeout", cfg.GetReadTimeout()),
			zap.Duration("write_timeout", cfg.GetWriteTimeout()),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// ========================================
	// STEP 15: Setup Health Check Monitoring
	// ========================================
	if cfg.Monitoring.Enabled {
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

		logger.Info("Health check monitoring started")
	}

	// ========================================
	// STEP 16: Start Periodic Session Cleanup
	// ========================================
	if sessionManager != nil {
		go func() {
			ticker := time.NewTicker(30 * time.Minute)
			defer ticker.Stop()

			for range ticker.C {
				ctx := context.Background()
				if err := sessionManager.CleanupExpiredSessions(ctx); err != nil {
					logger.Error("Failed to cleanup expired sessions", zap.Error(err))
				} else {
					logger.Debug("Expired sessions cleanup completed")
				}
			}
		}()
		logger.Info("Started periodic session cleanup")
	}

	// Log metrics usage (for debugging)
	_ = metrics

	// ========================================
	// STEP 17: Graceful Shutdown
	// ========================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Tenant Admin Service server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	// Close event consumer if initialized
	if eventConsumer != nil {
		if err := eventConsumer.Close(); err != nil {
			logger.Error("Failed to close event consumer", zap.Error(err))
		}
	}

	// Close database connections
	if err := dbManager.Close(); err != nil {
		logger.Error("Failed to close database connections", zap.Error(err))
	}

	// Close cache if initialized
	if cache != nil {
		cache.Close()
	}

	logger.Info("Tenant Admin Service server exited gracefully")
}

// ensureDatabaseExists creates the database if it doesn't exist
func ensureDatabaseExists(config resilience.DatabaseConfig, logger *zap.Logger) error {
	// Connect to postgres database to create the target database
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.SSLMode)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	defer db.Close()

	// Check if database exists
	var exists bool
	checkQuery := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", config.Name)
	err = db.QueryRow(checkQuery).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Create database if it doesn't exist
	if !exists {
		createQuery := fmt.Sprintf("CREATE DATABASE %s", config.Name)
		_, err = db.Exec(createQuery)
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		logger.Info("Database created successfully", zap.String("database", config.Name))
	} else {
		logger.Info("Database already exists", zap.String("database", config.Name))
	}

	return nil
}

// setupRoutes configures all the routes with improved structure and security
func setupRoutes(router *gin.Engine, tenantHandler *handlers.TenantAdminHandler, domainHandler *handlers.DomainHandler, rbacHandler *handlers.RBACHandler, userHandler *handlers.UserHandler, componentHandler *handlers.ComponentHandler, incidentHandler *handlers.IncidentHandler, subscriberHandler *handlers.SubscriberHandler, dependencyHandler *handlers.DependencyHandler, samlHandler *handlers.SAMLHandler, rbacService *services.RBACService, logger *zap.Logger, cfg *config.Config, registry *prometheus.Registry) {
	// Health check endpoints (excluded from auth and rate limiting)
	router.GET("/health", tenantHandler.HealthCheck)
	router.GET("/health/ready", tenantHandler.HealthCheck)
	router.GET("/health/live", tenantHandler.HealthCheck)

	// Prometheus metrics endpoint (with custom registry)
	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))

	// API routes with versioning
	api := router.Group("/api/v1")
	{
		// SAML/SSO public endpoints (no authentication required)
		saml := api.Group("/saml")
		{
			saml.POST("/login", samlHandler.InitiateLogin)
			saml.POST("/acs", samlHandler.AssertionConsumerService)
			saml.GET("/metadata", samlHandler.GetMetadata)
			saml.POST("/logout", samlHandler.SingleLogout)
		}

		// Auth routes (public, no authentication required)
		auth := api.Group("/auth")
		{
			auth.POST("/login", tenantHandler.Login)
			auth.POST("/refresh", tenantHandler.RefreshToken)
		}

		// Public routes (no authentication required)
		public := api.Group("/public")
		{
			public.POST("/login", tenantHandler.Login) // Backward compatibility

			// Tenant management (for SaaS-Admin service calls)
			public.GET("/tenants", tenantHandler.GetTenants)
			public.POST("/tenants", tenantHandler.CreateTenant)
			public.POST("/tenants/sync", tenantHandler.SyncTenant)
			public.GET("/tenants/:id", tenantHandler.GetTenant)
			public.GET("/tenants/slug/:slug", tenantHandler.GetTenantBySlug)
			public.PUT("/tenants/:id", tenantHandler.UpdateTenant)
			public.DELETE("/tenants/:id", tenantHandler.DeleteTenant)
			public.GET("/tenants/validate", tenantHandler.ValidateSubdomain) // Subdomain validation
		}

		// Session management routes (for service-to-service communication)
		sessions := api.Group("/sessions")
		{
			sessions.POST("", rbacHandler.CreateSession)
			sessions.GET("/:sessionId", rbacHandler.ValidateSession)
			sessions.DELETE("/:sessionId", rbacHandler.DeleteSession)
		}

		// Admin routes (for SaaS Admin service to update tenant credentials)
		admin := api.Group("/admin")
		{
			admin.PUT("/update-credentials", userHandler.UpdateCredentials)
		}

		// Public subscriber unsubscribe endpoint (no authentication required)
		api.POST("/subscribers/unsubscribe/:token", subscriberHandler.UnsubscribeByToken)

		// Protected routes (authentication required)
		protected := api.Group("/")

		// Create JWT config from local config
		jwtConfig := resilience.JWTConfig{
			Secret:     cfg.JWT.Secret,
			Expiration: cfg.GetJWTExpiration(),
			Issuer:     cfg.JWT.Issuer,
		}

		// Add JWT authentication middleware
		protected.Use(resilience.AuthMiddleware(jwtConfig))

		// Add tenant middleware for multi-tenancy
		protected.Use(resilience.TenantMiddleware())

		// Initialize RBAC middleware
		rbacMiddleware := middleware.NewRBACMiddleware(rbacService, logger)

		// Add audit logging middleware
		protected.Use(rbacMiddleware.AuditLogging())

		{
			// Auth routes
			auth := protected.Group("/auth")
			{
				auth.POST("/logout", tenantHandler.Logout)
				auth.POST("/verify", tenantHandler.VerifyToken)
			}

			// Dashboard routes
			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/stats", tenantHandler.GetDashboardStats)
			}

			// Tenant management routes
			tenants := protected.Group("/tenants")
			{
				tenants.GET("", tenantHandler.GetTenants)
				tenants.POST("", tenantHandler.CreateTenant)
				tenants.GET("/:id", tenantHandler.GetTenant)
				tenants.PUT("/:id", tenantHandler.UpdateTenant)
				tenants.DELETE("/:id", tenantHandler.DeleteTenant)
				tenants.GET("/slug/:slug", tenantHandler.GetTenantBySlug)
				tenants.GET("/domain/:domain", tenantHandler.GetTenantByDomain)
			}

			// Tenant branding routes
			branding := protected.Group("/branding")
			{
				branding.GET("/:tenant_id", tenantHandler.GetTenantBranding)
				branding.PUT("/:tenant_id", tenantHandler.UpdateTenantBranding)
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

			// User management routes
			users := protected.Group("/users")
			{
				users.GET("/stats", userHandler.GetUserStats)
				users.GET("", userHandler.GetUsers)
				users.POST("", userHandler.CreateUser)
				users.GET("/:id", userHandler.GetUser)
				users.PUT("/:id", userHandler.UpdateUser)
				users.DELETE("/:id", userHandler.DeleteUser)
			}

			// Component management routes
			components := protected.Group("/components")
			{
				components.GET("/stats", componentHandler.GetComponentStats)
				components.POST("/reorder", componentHandler.ReorderComponents)
				components.GET("", componentHandler.GetComponents)
				components.POST("", componentHandler.CreateComponent)
				components.GET("/:id", componentHandler.GetComponent)
				components.PUT("/:id", componentHandler.UpdateComponent)
				components.PUT("/:id/status", componentHandler.UpdateComponentStatus)
				components.DELETE("/:id", componentHandler.DeleteComponent)
			}

			// Incident management routes
			incidents := protected.Group("/incidents")
			{
				incidents.GET("/stats", incidentHandler.GetIncidentStats)
				incidents.GET("", incidentHandler.GetIncidents)
				incidents.POST("", incidentHandler.CreateIncident)
				incidents.GET("/:id", incidentHandler.GetIncident)
				incidents.PUT("/:id", incidentHandler.UpdateIncident)
				incidents.POST("/:id/resolve", incidentHandler.ResolveIncident)
				incidents.DELETE("/:id", incidentHandler.DeleteIncident)
			}

			// Subscriber management routes
			subscribers := protected.Group("/subscribers")
			{
				subscribers.GET("/stats", subscriberHandler.GetSubscriberStats)
				subscribers.GET("", subscriberHandler.GetSubscribers)
				subscribers.POST("", subscriberHandler.CreateSubscriber)
				subscribers.GET("/:id", subscriberHandler.GetSubscriber)
				subscribers.PUT("/:id", subscriberHandler.UpdateSubscriber)
				subscribers.POST("/:id/verify", subscriberHandler.VerifySubscriber)
				subscribers.DELETE("/:id", subscriberHandler.DeleteSubscriber)
			}

			// Dependency mapping routes
			dependencies := protected.Group("/dependencies")
			{
				dependencies.GET("/graph", dependencyHandler.GetDependencyGraph)
				dependencies.POST("", dependencyHandler.AddDependency)
				dependencies.DELETE("/:from_id/:to_id", dependencyHandler.RemoveDependency)
				dependencies.GET("/impact/:component_id", dependencyHandler.AnalyzeImpact)
				dependencies.GET("/health/:component_id", dependencyHandler.GetDependencyHealth)
				dependencies.POST("/validate", dependencyHandler.ValidateDependency)
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

			// SSO/SAML admin routes (protected - admin only)
			sso := protected.Group("/sso")
			{
				sso.GET("/providers", samlHandler.ListProviders)
				sso.GET("/providers/:id", samlHandler.GetProvider)
				sso.POST("/providers", samlHandler.CreateProvider)
				sso.PUT("/providers/:id", samlHandler.UpdateProvider)
				sso.DELETE("/providers/:id", samlHandler.DeleteProvider)
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
				roles.Use(rbacMiddleware.RequirePermission("users.read"))
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
				}

				// Team member management
				teamMembers := rbac.Group("/team-members")
				teamMembers.Use(rbacMiddleware.RequirePermission("users.read"))
				{
					teamMembers.POST("/:team_id", rbacMiddleware.RequirePermission("users.update"), rbacHandler.AddUserToTeam)
					teamMembers.DELETE("/:team_id/:user_id", rbacMiddleware.RequirePermission("users.update"), rbacHandler.RemoveUserFromTeam)
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
}
