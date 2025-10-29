// Package main is the entry point for the Tenant Admin Service.
// This is the modernized version using the shared-resilience module.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	resilience "github.com/anupamdutta5/shared-resilience"
	"github.com/anupamdutta5/tenant-admin-service/internal/config"
	"github.com/anupamdutta5/tenant-admin-service/internal/events"
	"github.com/anupamdutta5/tenant-admin-service/internal/handlers"
	"github.com/anupamdutta5/tenant-admin-service/internal/middleware"
	// "github.com/anupamdutta5/tenant-admin-service/internal/models" // Unused after disabling AutoMigrate
	"github.com/anupamdutta5/tenant-admin-service/internal/sessions"
	"github.com/anupamdutta5/tenant-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // PostgreSQL driver
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

	// Ensure database exists before connecting
	if err := ensureDatabaseExists(resilienceConfig.Database, logger); err != nil {
		logger.Fatal("Failed to ensure database exists", zap.Error(err))
	}

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

	// Database migrations are managed by Atlas (atlas.hcl)
	// DO NOT use GORM AutoMigrate - it conflicts with Atlas schema management
	// To apply migrations: atlas migrate apply --env dev
	// To create new migrations: atlas migrate diff <name> --env dev
	db := dbManager.GetDB()

	logger.Info("Tenant Admin Service using Atlas for migrations - AutoMigrate disabled")

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

	// Initialize session storage with Redis as primary and database as fallback
	var sessionManager *sessions.SessionManager
	var primarySessionStore sessions.SessionStore
	var fallbackSessionStore sessions.SessionStore

	// Try to initialize Redis session store as primary
	// Re-enabled after configuration standardization
	if resilienceConfig.Redis.Host != "" {
		redisStore, err := sessions.NewRedisSessionStore(resilienceConfig.Redis, logger)
		if err != nil {
			logger.Warn("Failed to initialize Redis session store, using database only",
				zap.Error(err))
		} else {
			primarySessionStore = redisStore
			logger.Info("Redis session store initialized as primary")
		}
	}

	// Always initialize database session store as fallback
	dbSessionStore := sessions.NewDBSessionStore(dbManager.GetDB(), logger)

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

	// Initialize business services with modernized dependencies
	tenantAdminService, err := services.NewTenantAdminService(localConfig, logger)
	if err != nil {
		logger.Fatal("Failed to initialize tenant admin service", zap.Error(err))
	}

	domainService := services.NewDomainService(logger)

	// Initialize RBAC service with session manager
	rbacService := services.NewRBACService(dbManager.GetDB(), logger, sessionManager)

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
	middlewareStack := resilience.DefaultMiddlewareStack(resilienceConfig, logger)
	for _, mw := range middlewareStack {
		router.Use(mw)
	}

	// Add custom Phase 4 middleware - Enhanced Error Handling & Logging
	router.Use(middleware.CorrelationIDMiddleware())      // Correlation ID for request tracing
	router.Use(middleware.RecoveryMiddleware(logger))     // Panic recovery with stack traces
	router.Use(middleware.LoggingMiddleware(logger))      // Enhanced structured logging

	// Add rate limiting middleware if enabled
	if rateLimiter != nil {
		router.Use(resilience.RateLimitMiddleware(resilienceConfig.RateLimit.RequestsPerMinute, time.Minute))
	}

	// Add tenant context middleware for subdomain routing
	baseDomain := os.Getenv("BASE_DOMAIN")
	if baseDomain == "" {
		baseDomain = "localhost" // Default for development
	}
	router.Use(middleware.TenantContextMiddleware(dbManager.GetDB(), logger, baseDomain))

	// Initialize Redis cache (production-ready with circuit breaker)
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := 6379
	if portStr := os.Getenv("REDIS_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			redisPort = p
		}
	}

	// Redis is enabled by default unless explicitly set to "false"
	redisEnabled := os.Getenv("REDIS_ENABLED") != "false"

	// Configure Redis connection settings
	redisConfig := resilience.RedisConfig{
		Host:         redisHost,
		Port:         redisPort,
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           0,
		MaxRetries:   3,
		PoolSize:     10,
		MinIdleConns: 5,
		KeyPrefix:    "tenant-admin:",
	}

	// Configure cache behavior
	cacheConfig := resilience.CacheConfig{
		Enabled:         redisEnabled,
		DefaultTTL:      5 * time.Minute,
		MaxSize:         1000,
		CleanupInterval: 10 * time.Minute,
		Type:            "redis",
	}

	// Create Redis cache client using shared-resilience
	redisClient := resilience.NewRedisCache(redisConfig, cacheConfig, logger)

	// Initialize services with Redis caching
	componentService := services.NewComponentService(db, redisClient, logger)
	incidentService := services.NewIncidentService(db, redisClient, logger)
	subscriberService := services.NewSubscriberService(db, redisClient, logger)
	dependencyService := services.NewDependencyService(db)

	// Initialize session service for refresh token management
	sessionService := services.NewSessionService(db, logger)

	// Initialize modernized handlers with shared error handling
	tenantAdminHandler := handlers.NewTenantAdminHandler(tenantAdminService, statusPageService, rbacService, sessionService, logger)
	domainHandler := handlers.NewDomainHandler(domainService, logger)
	rbacHandler := handlers.NewRBACHandler(rbacService, logger)
	userHandler := handlers.NewUserHandler(tenantAdminService, logger)
	componentHandler := handlers.NewComponentHandler(componentService, logger)
	incidentHandler := handlers.NewIncidentHandler(incidentService, logger)
	subscriberHandler := handlers.NewSubscriberHandler(subscriberService, logger)
	dependencyHandler := handlers.NewDependencyHandler(dependencyService)

	// Initialize SAML/SSO service for enterprise authentication
	samlBaseURL := os.Getenv("SAML_BASE_URL")
	if samlBaseURL == "" {
		samlBaseURL = fmt.Sprintf("http://localhost:%d", resilienceConfig.Server.Port)
	}

	samlService := services.NewSAMLService(db, tenantAdminService, logger, &services.SAMLConfig{
		BaseURL: samlBaseURL,
	})

	samlHandler := handlers.NewSAMLHandler(samlService, tenantAdminService, logger)

	logger.Info("SAML/SSO service initialized",
		zap.String("base_url", samlBaseURL))

	// Initialize RabbitMQ event consumer for tenant sync
	// Use GetRabbitMQURL() from config instead of hardcoded credentials
	rabbitmqURL := localConfig.GetRabbitMQURL()

	// Create event handler with proper service injection (best practice)
	eventHandler := events.NewRabbitMQTenantEventHandler(dbManager.GetDB(), tenantAdminService, logger)

	// Create consumer
	eventConsumer, err := events.NewConsumer(events.ConsumerConfig{
		URL:     rabbitmqURL,
		Handler: eventHandler,
		Logger:  logger,
	})
	if err != nil {
		logger.Warn("Failed to initialize RabbitMQ consumer, tenant events will not be synchronized",
			zap.Error(err),
			zap.String("url", rabbitmqURL))
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

	// Setup routes with improved structure
	setupModernizedRoutes(router, tenantAdminHandler, domainHandler, rbacHandler, userHandler, componentHandler, incidentHandler, subscriberHandler, dependencyHandler, samlHandler, rbacService, logger, resilienceConfig)

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

	// Start periodic session cleanup
	if sessionManager != nil {
		go func() {
			ticker := time.NewTicker(30 * time.Minute) // Run cleanup every 30 minutes
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

	// Close rate limiter if initialized
	if rateLimiter != nil {
		rateLimiter.Close()
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

// setupModernizedRoutes configures all the routes with improved structure and security
func setupModernizedRoutes(router *gin.Engine, tenantHandler *handlers.TenantAdminHandler, domainHandler *handlers.DomainHandler, rbacHandler *handlers.RBACHandler, userHandler *handlers.UserHandler, componentHandler *handlers.ComponentHandler, incidentHandler *handlers.IncidentHandler, subscriberHandler *handlers.SubscriberHandler, dependencyHandler *handlers.DependencyHandler, samlHandler *handlers.SAMLHandler, rbacService *services.RBACService, logger *zap.Logger, config *resilience.Config) {
	// Health check endpoints (excluded from auth and rate limiting)
	router.GET("/health", tenantHandler.HealthCheck)
	router.GET("/health/ready", tenantHandler.HealthCheck)
	router.GET("/health/live", tenantHandler.HealthCheck)

	// API routes with versioning
	api := router.Group("/api/v1")
	{
		// SAML/SSO public endpoints (no authentication required)
		saml := api.Group("/saml")
		{
			saml.POST("/login", samlHandler.InitiateLogin)                // Initiate SAML login
			saml.POST("/acs", samlHandler.AssertionConsumerService)       // Assertion Consumer Service
			saml.GET("/metadata", samlHandler.GetMetadata)                // SP metadata for IdP configuration
			saml.POST("/logout", samlHandler.SingleLogout)                // Single Logout
		}

		// Auth routes (public, no authentication required)
		auth := api.Group("/auth")
		{
			auth.POST("/login", tenantHandler.Login)
			auth.POST("/refresh", tenantHandler.RefreshToken) // Refresh access token using refresh token
		}

		// Public routes (no authentication required)
		public := api.Group("/public")
		{
			public.POST("/login", tenantHandler.Login) // Keep for backward compatibility

			// Tenant management (for SaaS-Admin service calls)
			public.GET("/tenants", tenantHandler.GetTenants)
			public.POST("/tenants", tenantHandler.CreateTenant)
			public.POST("/tenants/sync", tenantHandler.SyncTenant) // Idempotent sync endpoint for SaaS Admin
			public.GET("/tenants/:id", tenantHandler.GetTenant)
			public.GET("/tenants/slug/:slug", tenantHandler.GetTenantBySlug)
			public.PUT("/tenants/:id", tenantHandler.UpdateTenant)
			public.DELETE("/tenants/:id", tenantHandler.DeleteTenant)
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

		// Add JWT authentication middleware
		if config.JWT.Secret != "" {
			protected.Use(resilience.AuthMiddleware(config.JWT))
		}

		// Add tenant middleware for multi-tenancy
		protected.Use(resilience.TenantMiddleware())

		// Initialize RBAC middleware
		rbacMiddleware := middleware.NewRBACMiddleware(rbacService, logger)

		// REMOVED: Session validation middleware (redundant with JWT authentication)
		// JWT authentication middleware (line 526) already provides user_id, tenant_id, and email from token claims
		// protected.Use(rbacMiddleware.SessionValidation())

		// Add audit logging middleware
		protected.Use(rbacMiddleware.AuditLogging())

		{
			// Auth routes
			auth := protected.Group("/auth")
			{
				auth.POST("/logout", tenantHandler.Logout)
				auth.POST("/verify", tenantHandler.VerifyToken)
			}

			// Dashboard routes (aggregated statistics)
			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/stats", tenantHandler.GetDashboardStats)
			}

			// Tenant management routes (main CRUD operations)
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

			// Tenant branding routes - separate group to avoid route conflicts
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

			// User management routes (tenant context from middleware)
			users := protected.Group("/users")
			{
				users.GET("/stats", userHandler.GetUserStats)
				users.GET("", userHandler.GetUsers)
				users.POST("", userHandler.CreateUser)
				users.GET("/:id", userHandler.GetUser)
				users.PUT("/:id", userHandler.UpdateUser)
				users.DELETE("/:id", userHandler.DeleteUser)
			}

			// Component management routes (tenant context from middleware)
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

			// Incident management routes (tenant context from middleware)
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

			// Subscriber management routes (tenant context from middleware)
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

			// Dependency mapping routes (tenant context from middleware)
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
				sso.GET("/providers", samlHandler.ListProviders)          // List SSO providers for tenant
				sso.GET("/providers/:id", samlHandler.GetProvider)        // Get provider details
				sso.POST("/providers", samlHandler.CreateProvider)        // Create new SSO provider
				sso.PUT("/providers/:id", samlHandler.UpdateProvider)     // Update SSO provider
				sso.DELETE("/providers/:id", samlHandler.DeleteProvider)  // Delete SSO provider
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
				}

				// Team member management - separate endpoint to avoid route conflicts
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

	// Frontend now served independently on port 3002 (tenant-admin-frontend service)
	// All static file routes removed - backend is API-only
	// Subdomain routing middleware remains active for tenant isolation
}
// TODO: CLEANUP - Update auth middleware usage
// Replace local auth with: auth.NewMiddleware(authConfig, logger)
// Import: github.com/anupamdutta5/shared-resilience/auth
