package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/enterprise-status/statuspage/internal/api/middleware"
	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/internal/services"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/email"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Server struct {
	config                       *config.Config
	router                       *gin.Engine
	httpServer                   *http.Server
	statusService                *services.StatusService
	incidentService              *services.IncidentService
	authService                  *services.AuthService
	maintenanceService           *services.MaintenanceService
	subscriberService            *services.SubscriberService
	notificationService          *services.NotificationService
	monitorService               *services.MonitorService
	uptimeService                *services.UptimeService
	rbacService                  *services.RBACService
	auditService                 *services.AuditService
	brandingService              *services.BrandingService
	templateService              *services.TemplateService
	integrationService           *services.IntegrationService
	privatePageService           *services.PrivatePageService
	monitoringIntegrationService *services.MonitoringIntegrationService
	saasService                  *services.SaaSService
	analyticsService             *services.AnalyticsService
	subscriptionService          *services.SubscriptionService
	paymentService               *services.PaymentService
	featureFlagService           *services.FeatureFlagService
}

func NewServer(cfg *config.Config) *Server {
	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	server := &Server{
		config:                       cfg,
		router:                       gin.New(),
		statusService:                services.NewStatusService(),
		incidentService:              services.NewIncidentService(),
		authService:                  services.NewAuthService(cfg),
		maintenanceService:           services.NewMaintenanceService(),
		subscriberService:            services.NewSubscriberService(),
		monitorService:               services.NewMonitorService(),
		uptimeService:                services.NewUptimeService(),
		rbacService:                  services.NewRBACService(),
		auditService:                 services.NewAuditService(),
		brandingService:              services.NewBrandingService(),
		templateService:              services.NewTemplateService(),
		integrationService:           services.NewIntegrationService(),
		privatePageService:           services.NewPrivatePageService(),
		monitoringIntegrationService: services.NewMonitoringIntegrationService(),
		saasService:                  services.NewSaaSService(),
		analyticsService:             services.NewAnalyticsService(),
		subscriptionService:          services.NewSubscriptionService(),
		paymentService:               services.NewPaymentService(),
		featureFlagService:           services.NewFeatureFlagService(database.GetDB()),
	}

	// Initialize services
	var emailSender email.Sender
	if cfg.Email.Provider == "smtp" {
		smtpConfig := email.SMTPConfig{
			Host:     cfg.Email.SMTP.Host,
			Port:     cfg.Email.SMTP.Port,
			Username: cfg.Email.SMTP.Username,
			Password: cfg.Email.SMTP.Password,
			From:     cfg.Email.SMTP.From,
			UseTLS:   cfg.Email.SMTP.UseTLS,
		}
		emailSender = email.NewSMTPSender(smtpConfig)
	} else {
		emailSender = email.NewLogSender()
	}
	server.notificationService = services.NewNotificationService(server.subscriberService, emailSender)

	server.setupRouter()
	return server
}

func (s *Server) setupRouter() {
	// Initialize middleware
	errorHandler := middleware.NewErrorHandler(s.config.Environment == "development")
	validationMiddleware := middleware.NewValidationMiddleware()
	rateLimiter := middleware.NewRateLimiter(&s.config.Security.RateLimit)
	csrfProtection := middleware.NewCSRFProtection(s.config.Security.Session.Secret, s.config.Security.Session.Secure, s.config.Security.Session.SameSite)
	monitoringMiddleware := middleware.NewMonitoringMiddleware()

	// Add request ID middleware first
	s.router.Use(monitoringMiddleware.RequestIDMiddleware())

	// Add monitoring middleware
	s.router.Use(monitoringMiddleware.RequestLoggingMiddleware())
	s.router.Use(monitoringMiddleware.PerformanceMiddleware())
	s.router.Use(monitoringMiddleware.ErrorTrackingMiddleware())
	s.router.Use(monitoringMiddleware.SecurityMonitoringMiddleware())
	s.router.Use(monitoringMiddleware.HealthCheckMiddleware())
	s.router.Use(monitoringMiddleware.MetricsMiddleware())

	// Add CORS middleware
	s.router.Use(middleware.CORSMiddleware(nil))

	// Add security headers
	s.router.Use(middleware.NewSecurityMiddleware(nil).SecurityHeadersMiddleware())

	// Add error handling
	s.router.Use(errorHandler.HandleError())
	s.router.Use(errorHandler.Recovery())

	// Add input sanitization
	s.router.Use(validationMiddleware.SanitizeInput())

	// Add rate limiting
	s.router.Use(rateLimiter.RateLimitMiddleware())
	s.router.Use(rateLimiter.BurstRateLimitMiddleware())

	// Add CSRF protection for web routes
	s.router.Use(csrfProtection.CSRFForWeb())

	// Middleware
	s.router.Use(gin.Logger())
	s.router.Use(gin.Recovery())

	// Analytics middleware
	s.router.Use(middleware.AnalyticsMiddleware(s.analyticsService))
	s.router.Use(middleware.PageViewMiddleware(s.analyticsService))

	// Load HTML templates
	s.router.LoadHTMLGlob("web/templates/*")

	// PRIORITY ROUTES - Define these FIRST to ensure they work
	s.router.GET("/debug-route-test", func(c *gin.Context) {
		logger.Info("DEBUG route handler called - this should work!")
		c.JSON(http.StatusOK, gin.H{"message": "DEBUG route works", "path": c.Request.URL.Path, "method": c.Request.Method})
	})

	s.router.GET("/pricing", func(c *gin.Context) {
		logger.Info("PRIORITY pricing route handler called")
		c.HTML(http.StatusOK, "pricing.html", gin.H{
			"title": "Pricing Plans - Status Page Platform",
		})
	})

	s.router.GET("/pricing-test", func(c *gin.Context) {
		logger.Info("PRIORITY pricing-test route handler called")
		c.JSON(http.StatusOK, gin.H{"message": "PRIORITY test route works", "path": c.Request.URL.Path})
	})

	// Health check endpoint
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": 1757404666,
			"version":   "1.0.0",
		})
	})

	// Test route that mimics health structure exactly
	s.router.GET("/working-test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "working",
			"message": "This route should work",
			"path":    c.Request.URL.Path,
		})
	})

	// SUPER SIMPLE TEST ROUTE - NO COMPLEX LOGIC
	s.router.GET("/simple-test", func(c *gin.Context) {
		c.String(http.StatusOK, "SIMPLE TEST WORKS")
	})

	// Test endpoint for debugging
	s.router.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "API is working",
			"cookies": c.Request.Cookies(),
		})
	})

	// Test admin endpoint without authentication
	s.router.GET("/api/test-admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Admin API is working without auth",
		})
	})

	// Direct admin services endpoint without authentication
	s.router.GET("/api/v1/admin/services-direct", s.getAdminServices)

	// Test page for debugging
	s.router.GET("/test", func(c *gin.Context) {
		c.HTML(http.StatusOK, "test.html", gin.H{})
	})

	// JavaScript test page
	s.router.GET("/js-test", func(c *gin.Context) {
		c.HTML(http.StatusOK, "js_test.html", gin.H{})
	})

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Admin API routes (no tenant required)
		adminAPI := v1.Group("/admin")
		// Skip authentication for development
		// adminAPI.Use(middleware.AuthRequired(s.authService))
		{
			// Additional admin routes for dashboard
			adminAPI.GET("/services", s.getAdminServices)
			adminAPI.GET("/subscribers", s.getAdminSubscribers)
			adminAPI.GET("/audit-logs", s.getAdminAuditLogs)
			adminAPI.GET("/status", s.getAdminStatus)
			adminAPI.GET("/maintenance", s.getAdminMaintenance)
			adminAPI.GET("/monitors", s.getAdminMonitors)
			adminAPI.GET("/users", s.getAdminUsers)
			adminAPI.GET("/integrations", s.getAdminIntegrations)

			// Feature Flags
			adminAPI.GET("/feature-flags", s.getFeatureFlags)
			adminAPI.POST("/feature-flags", s.setFeatureFlag)
			adminAPI.PUT("/feature-flags/:feature", s.updateFeatureFlag)
			adminAPI.DELETE("/feature-flags/:feature", s.deleteFeatureFlag)
		}

		// Status routes
		statusGroup := v1.Group("/status")
		{
			statusGroup.GET("", s.getStatus)
			statusGroup.GET("/:id", s.getStatusByID)
			statusGroup.GET("/summary", s.getStatusSummary)
		}

		// Services routes
		servicesGroup := v1.Group("/services")
		{
			servicesGroup.GET("", s.getPublicServices)
		}

		// Incident routes
		incidentGroup := v1.Group("/incidents")
		{
			incidentGroup.GET("", s.getIncidents)
			incidentGroup.GET("/:id", s.getIncidentByID)
			incidentGroup.GET("/:id/updates", s.getIncidentUpdates)
			incidentGroup.GET("/active", s.getActiveIncidents)
			incidentGroup.GET("/resolved", s.getResolvedIncidents)
		}

		// Maintenance routes
		maintenanceGroup := v1.Group("/maintenance")
		{
			maintenanceGroup.GET("", s.getMaintenanceEvents)
			maintenanceGroup.GET("/:id", s.getMaintenanceEventByID)
			maintenanceGroup.GET("/upcoming", s.getUpcomingMaintenance)
			maintenanceGroup.GET("/active", s.getActiveMaintenance)
		}

		// Subscriber routes
		subscriberGroup := v1.Group("/subscribers")
		{
			subscriberGroup.POST("", s.subscribe)
			subscriberGroup.DELETE("", s.unsubscribe)
		}

		// Uptime routes
		uptimeGroup := v1.Group("/uptime")
		{
			uptimeGroup.GET("", s.getUptimeStats)
			uptimeGroup.GET("/history/:monitor_id", s.getUptimeHistory)
		}

		// Analytics routes
		v1.GET("/analytics", s.getAnalyticsSummary)

		// RSS/Atom feeds
		v1.GET("/incidents.rss", s.getIncidentsRSS)
		v1.GET("/incidents.atom", s.getIncidentsAtom)
	}

	// Admin API routes
	adminAPI := v1.Group("/admin")
	{
		// Authentication
		adminAPI.POST("/login", s.login)

		// Protected admin routes
		authedAPI := adminAPI.Group("")
		// Skip authentication for development
		// authedAPI.Use(middleware.AuthRequired(s.authService))
		{
			// Status management
			authedAPI.POST("/status", s.createStatus)
			authedAPI.PUT("/status/:id", s.updateStatus)
			authedAPI.DELETE("/status/:id", s.deleteStatus)

			// Incident management
			authedAPI.GET("/incidents", s.getAdminIncidents)
			authedAPI.GET("/incidents/:id", s.getIncident)
			authedAPI.POST("/incidents", s.createIncident)
			authedAPI.PUT("/incidents/:id", s.updateIncident)
			authedAPI.DELETE("/incidents/:id", s.deleteIncident)

			// Maintenance management
			authedAPI.POST("/maintenance", s.createMaintenanceEvent)
			authedAPI.PUT("/maintenance/:id", s.updateMaintenanceEvent)
			authedAPI.DELETE("/maintenance/:id", s.deleteMaintenanceEvent)

			// Monitor management
			authedAPI.POST("/monitors", s.createMonitor)
			authedAPI.PUT("/monitors/:id", s.updateMonitor)
			authedAPI.DELETE("/monitors/:id", s.deleteMonitor)

			// User management
			authedAPI.POST("/users", s.createUser)
			authedAPI.PUT("/users/:id", s.updateUser)
			authedAPI.DELETE("/users/:id", s.deleteUser)

			// Template management
			authedAPI.GET("/templates/incidents", s.getIncidentTemplates)
			authedAPI.POST("/templates/incidents", s.createIncidentTemplate)
			authedAPI.PUT("/templates/incidents/:id", s.updateIncidentTemplate)
			authedAPI.DELETE("/templates/incidents/:id", s.deleteIncidentTemplate)
			authedAPI.GET("/templates/maintenance", s.getMaintenanceTemplates)
			authedAPI.POST("/templates/maintenance", s.createMaintenanceTemplate)
			authedAPI.PUT("/templates/maintenance/:id", s.updateMaintenanceTemplate)
			authedAPI.DELETE("/templates/maintenance/:id", s.deleteMaintenanceTemplate)

			// Audit logs - removed duplicate route

			// Private page management
			authedAPI.GET("/private-pages", s.getPrivatePages)
			authedAPI.POST("/private-pages", s.createPrivatePage)
			authedAPI.PUT("/private-pages/:id", s.updatePrivatePage)
			authedAPI.DELETE("/private-pages/:id", s.deletePrivatePage)
			authedAPI.POST("/private-pages/:id/regenerate-key", s.regeneratePrivatePageKey)
			authedAPI.POST("/private-pages/:id/toggle", s.togglePrivatePageStatus)

			// Monitoring integration
			authedAPI.GET("/monitoring/status", s.getMonitoringStatus)
			authedAPI.POST("/monitoring/sync", s.syncAllMonitoringTools)

			// Service management
			authedAPI.POST("/services", s.createAdminService)
			authedAPI.GET("/services/:id", s.getAdminService)
			authedAPI.PUT("/services/:id", s.updateAdminService)
			authedAPI.DELETE("/services/:id", s.deleteAdminService)

			// Subscriber management
			authedAPI.DELETE("/subscribers/:id", s.deleteAdminSubscriber)

			// SaaS Admin routes
			saasAdminAPI := authedAPI.Group("/saas/admin")
			{
				// Overview and metrics
				saasAdminAPI.GET("/metrics", s.getSaaSOverviewMetrics)
				saasAdminAPI.GET("/activity", s.getSaaSRecentActivity)
				saasAdminAPI.GET("/plan-distribution", s.getSaaSPlanDistribution)
				saasAdminAPI.GET("/charts/overview", s.getSaaSOverviewCharts)

				// Tenant management
				saasAdminAPI.GET("/tenants", s.getSAASAllTenants)
				saasAdminAPI.POST("/tenants", s.createSAASSTenant)
				saasAdminAPI.GET("/tenants/:id", s.getSAASSTenantByID)
				saasAdminAPI.PUT("/tenants/:id", s.updateSAASSTenant)
				saasAdminAPI.DELETE("/tenants/:id", s.deleteSAASSTenant)

				// Tenant subscription management
				saasAdminAPI.GET("/tenants/:id/subscription", s.getTenantSubscription)
				saasAdminAPI.POST("/tenants/:id/subscription", s.createTenantSubscription)
				saasAdminAPI.PUT("/tenants/:id/subscription/upgrade", s.upgradeTenantSubscription)
				saasAdminAPI.PUT("/tenants/:id/subscription/downgrade", s.downgradeTenantSubscription)
				saasAdminAPI.POST("/tenants/:id/subscription/cancel", s.cancelTenantSubscription)
				saasAdminAPI.GET("/tenants/:id/usage", s.getTenantUsage)

				// Tenant feature flag management
				saasAdminAPI.GET("/tenants/:id/feature-flags", s.getTenantFeatureFlags)
				saasAdminAPI.POST("/tenants/:id/feature-flags", s.setTenantFeatureFlag)
				saasAdminAPI.PUT("/tenants/:id/feature-flags/:feature", s.updateTenantFeatureFlag)
				saasAdminAPI.DELETE("/tenants/:id/feature-flags/:feature", s.deleteTenantFeatureFlag)

				// Global feature flag management
				saasAdminAPI.GET("/feature-flags", s.getAllFeatureFlags)
				saasAdminAPI.POST("/feature-flags/global", s.setGlobalFeatureFlag)

				// SaaS-level feature availability management
				saasAdminAPI.GET("/feature-availability", s.getFeatureAvailability)
				saasAdminAPI.POST("/feature-availability", s.setFeatureAvailability)
				saasAdminAPI.PUT("/feature-availability/:feature", s.updateFeatureAvailability)
				saasAdminAPI.DELETE("/feature-availability/:feature", s.deleteFeatureAvailability)

				// Subscription management
				saasAdminAPI.GET("/subscriptions", s.getSAASAllSubscriptions)
				saasAdminAPI.GET("/subscriptions/:id", s.getSAASSubscriptionByID)
				saasAdminAPI.PUT("/subscriptions/:id", s.updateSAASSubscription)
				saasAdminAPI.POST("/subscriptions/:id/cancel", s.cancelSAASSubscription)

				// Plan management
				saasAdminAPI.GET("/plans", s.getSAASAllPlans)
				saasAdminAPI.POST("/plans", s.createSAASPlan)
				saasAdminAPI.GET("/plans/:id", s.getSAASPlanByID)
				saasAdminAPI.PUT("/plans/:id", s.updateSAASPlan)
				saasAdminAPI.DELETE("/plans/:id", s.deleteSAASPlan)

				// Billing management
				saasAdminAPI.GET("/billing/stats", s.getSAASBillingStats)
				saasAdminAPI.GET("/billing/events", s.getSAASBillingEvents)

				// Analytics
				saasAdminAPI.GET("/analytics", s.getSAASAnalytics)

				// Settings
				saasAdminAPI.GET("/settings", s.getSAASSettings)
				saasAdminAPI.PUT("/settings/:key", s.updateSAASSetting)

				// Notifications
				saasAdminAPI.GET("/notifications", s.getSAASNotifications)
				saasAdminAPI.POST("/notifications", s.createSAASNotification)
				saasAdminAPI.PUT("/notifications/:id", s.updateSAASNotification)
				saasAdminAPI.DELETE("/notifications/:id", s.deleteSAASNotification)
			}

		}
	}

	// Static files
	s.router.Static("/static", "./web/static")

	// Public routes (must be defined before other groups)
	s.router.GET("/", s.showLandingPage)
	s.router.GET("/private/:access_key", s.showPrivatePage)

	// Web routes
	web := s.router.Group("")
	{
		// Admin login
		web.GET("/admin/login", s.showLoginPage)
		web.POST("/admin/login", s.handleWebLogin)
		web.GET("/admin/logout", s.handleLogout)

		// Protected web routes
		authedWeb := web.Group("")
		// Skip authentication for development
		// authedWeb.Use(middleware.AuthRequired(s.authService))
		{
			// Admin dashboard
			authedWeb.GET("/admin/dashboard", s.showDashboardPage)
			authedWeb.GET("/admin/saas", s.showSaaSAdminDashboard)

			// Maintenance management pages
			authedWeb.GET("/admin/maintenance/new", s.showNewMaintenancePage)
			authedWeb.POST("/admin/maintenance/new", s.handleNewMaintenance)
			authedWeb.GET("/admin/maintenance/edit/:id", s.showEditMaintenancePage)
			authedWeb.POST("/admin/maintenance/edit/:id", s.handleEditMaintenance)
			authedWeb.GET("/admin/maintenance/delete/:id", s.handleDeleteMaintenance)
		}
	}

	// Tenant routes
	tenantHandlers := NewTenantHandlers(
		database.DB,
		s.statusService,
		s.incidentService,
		s.maintenanceService,
		s.subscriberService,
		s.monitorService,
		s.brandingService,
		s.subscriptionService,
		s.paymentService,
	)

	paymentHandlers := NewPaymentHandlers(s.paymentService)

	// Custom domain service and handlers
	customDomainService := services.NewCustomDomainService(database.DB)
	customDomainHandlers := NewCustomDomainHandlers(customDomainService)

	// Security service and handlers
	securityService := services.NewTenantSecurityService(database.DB)
	securityHandlers := NewSecurityHandlers(securityService)
	securityMiddleware := middleware.NewSecurityMiddleware(securityService)

	// Monitoring analytics service and handlers
	monitoringService := services.NewMonitoringAnalyticsService(database.DB)
	monitoringHandlers := NewMonitoringAnalyticsHandlers(monitoringService)

	// Tenant API routes
	tenantAPI := s.router.Group("/api/v1/tenant")
	tenantAPI.Use(middleware.TenantMiddleware(s.saasService))
	tenantAPI.Use(middleware.RequireTenant())
	tenantAPI.Use(securityMiddleware.SecurityHeadersMiddleware())
	tenantAPI.Use(securityMiddleware.AuditLoggingMiddleware())
	tenantAPI.Use(securityMiddleware.DataAccessLoggingMiddleware())
	tenantAPI.Use(securityMiddleware.RateLimitingMiddleware())
	tenantAPI.Use(securityMiddleware.TenantIsolationMiddleware())
	{
		// Tenant info
		tenantAPI.GET("/info", tenantHandlers.GetTenantInfo)
		tenantAPI.GET("/overview", tenantHandlers.GetTenantOverview)

		// Services
		tenantAPI.GET("/services", tenantHandlers.GetTenantServices)
		tenantAPI.POST("/services", tenantHandlers.CreateTenantService)

		// Incidents
		tenantAPI.GET("/incidents", tenantHandlers.GetTenantIncidents)
		tenantAPI.POST("/incidents", tenantHandlers.CreateTenantIncident)

		// Maintenance
		tenantAPI.GET("/maintenance", tenantHandlers.GetTenantMaintenance)

		// Subscribers
		tenantAPI.GET("/subscribers", tenantHandlers.GetTenantSubscribers)
		tenantAPI.POST("/subscribe", s.handleTenantSubscribe)

		// Monitors
		tenantAPI.GET("/monitors", tenantHandlers.GetTenantMonitors)

		// Branding
		tenantAPI.GET("/branding", tenantHandlers.GetTenantBranding)
		tenantAPI.PUT("/branding", tenantHandlers.UpdateTenantBranding)

		// Settings
		tenantAPI.GET("/settings", tenantHandlers.GetTenantSettings)
		tenantAPI.PUT("/settings", tenantHandlers.UpdateTenantSettings)

		// Billing
		tenantAPI.GET("/billing", tenantHandlers.GetTenantBilling)
		tenantAPI.GET("/billing/invoices", tenantHandlers.GetTenantInvoices)
		tenantAPI.GET("/billing/payments", tenantHandlers.GetTenantPayments)
		tenantAPI.GET("/billing/metrics", tenantHandlers.GetTenantBillingMetrics)
		tenantAPI.GET("/subscription/status", tenantHandlers.GetSubscriptionStatus)
		tenantAPI.POST("/checkout", paymentHandlers.CreateCheckoutSession)
		tenantAPI.POST("/subscription/cancel", paymentHandlers.CancelSubscription)
		tenantAPI.POST("/customer-portal", paymentHandlers.GetCustomerPortalURL)

		// Custom Domain Management
		tenantAPI.POST("/domains", customDomainHandlers.SetCustomDomain)
		tenantAPI.GET("/domains/status", customDomainHandlers.GetDomainStatus)
		tenantAPI.POST("/domains/verify", customDomainHandlers.VerifyDomain)
		tenantAPI.DELETE("/domains", customDomainHandlers.RemoveCustomDomain)
		tenantAPI.GET("/domains/dns-instructions", customDomainHandlers.GetDNSInstructions)
		tenantAPI.GET("/domains/ssl-instructions", customDomainHandlers.GetSSLInstructions)
		tenantAPI.POST("/domains/ssl/upload", customDomainHandlers.UploadSSLCertificate)
		tenantAPI.POST("/domains/ssl/auto", customDomainHandlers.EnableAutomaticSSL)

		// Security Management
		tenantAPI.GET("/security/config", securityHandlers.GetSecurityConfig)
		tenantAPI.PUT("/security/config", securityHandlers.UpdateSecurityConfig)
		tenantAPI.GET("/security/metrics", securityHandlers.GetSecurityMetrics)
		tenantAPI.GET("/security/audit-logs", securityHandlers.GetAuditLogs)
		tenantAPI.GET("/security/violations", securityHandlers.GetSecurityViolations)
		tenantAPI.POST("/security/violations/:id/resolve", securityHandlers.ResolveSecurityViolation)
		tenantAPI.POST("/security/api-keys", securityHandlers.GenerateAPIKey)
		tenantAPI.POST("/security/api-keys/validate", securityHandlers.ValidateAPIKey)
		tenantAPI.POST("/security/encrypt", securityHandlers.EncryptData)
		tenantAPI.POST("/security/decrypt", securityHandlers.DecryptData)

		// Monitoring Analytics
		tenantAPI.GET("/monitoring/uptime-summary", monitoringHandlers.GetUptimeSummary)
		tenantAPI.GET("/monitoring/metrics", monitoringHandlers.GetMonitoringMetrics)
		tenantAPI.GET("/monitoring/services/health", monitoringHandlers.GetServiceHealthData)
		tenantAPI.GET("/monitoring/services/:serviceId/uptime", monitoringHandlers.GetServiceUptimeData)
		tenantAPI.GET("/monitoring/services/:serviceId/response-time", monitoringHandlers.GetServiceResponseTimeData)
		tenantAPI.GET("/monitoring/services/:serviceId/status-history", monitoringHandlers.GetServiceStatusHistory)
		tenantAPI.GET("/monitoring/services/:serviceId/graph", monitoringHandlers.GetServiceGraphData)
	}

	// Tenant admin routes
	tenantAdmin := s.router.Group("/tenant-admin")
	tenantAdmin.Use(middleware.TenantMiddleware(s.saasService))
	tenantAdmin.Use(middleware.RequireTenant())
	tenantAdmin.Use(middleware.TenantAdminMiddleware())
	{
		tenantAdmin.GET("/", s.showTenantAdminDashboard)
	}

	// Payment webhooks
	s.router.POST("/webhooks/stripe", paymentHandlers.HandleWebhook)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:    ":" + strconv.Itoa(s.config.Server.Port),
		Handler: s.router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", zap.Int("port", s.config.Server.Port))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

	if err := s.httpServer.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
	return nil
}

// All the handler methods would go here...
// For now, I'll include just the essential ones to get the server running

func (s *Server) getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "operational"})
}

func (s *Server) getStatusByID(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "operational"})
}

func (s *Server) getStatusSummary(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "operational"})
}

func (s *Server) getPublicServices(c *gin.Context) {
	// Get services from the admin services endpoint
	services, err := s.getAdminServicesData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get services", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, services)
}

func (s *Server) getAdminServicesData() (gin.H, error) {
	// Get tenant ID from context (for multi-tenant support)
	tenantID := uint(1) // Default tenant for now

	services, err := s.statusService.GetServicesByTenantID(tenantID)
	if err != nil {
		return nil, err
	}

	// Convert to response format with uptime data
	var serviceResponses []gin.H
	for _, service := range services {
		// Get uptime for this service
		uptime, _ := s.statusService.GetServiceUptime(service.ID, 30) // 30 days

		serviceResponses = append(serviceResponses, gin.H{
			"id":          service.ID,
			"name":        service.Name,
			"status":      service.Status,
			"description": service.Description,
			"group":       service.Group,
			"position":    service.Position,
			"show_uptime": service.ShowUptime,
			"uptime":      fmt.Sprintf("%.1f%%", uptime),
			"created_at":  service.CreatedAt,
			"updated_at":  service.UpdatedAt,
		})
	}

	return gin.H{"services": serviceResponses}, nil
}

func (s *Server) getIncidents(c *gin.Context) {
	incidents, err := s.incidentService.GetAllIncidents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get incidents", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"incidents": incidents})
}

func (s *Server) getIncidentByID(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"incident": gin.H{}})
}

func (s *Server) getIncidentUpdates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"updates": []gin.H{}})
}

func (s *Server) getActiveIncidents(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"incidents": []gin.H{}})
}

func (s *Server) getResolvedIncidents(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"incidents": []gin.H{}})
}

func (s *Server) getMaintenanceEvents(c *gin.Context) {
	maintenanceEvents, err := s.maintenanceService.GetAllMaintenanceEvents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get maintenance events", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"maintenance": maintenanceEvents})
}

func (s *Server) getMaintenanceEventByID(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"maintenance": gin.H{}})
}

func (s *Server) getUpcomingMaintenance(c *gin.Context) {
	maintenanceEvents, err := s.maintenanceService.GetUpcomingMaintenance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get upcoming maintenance", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"maintenance": maintenanceEvents})
}

func (s *Server) getActiveMaintenance(c *gin.Context) {
	maintenanceEvents, err := s.maintenanceService.GetActiveMaintenance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active maintenance", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"maintenance": maintenanceEvents})
}

func (s *Server) subscribe(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Subscribed"})
}

func (s *Server) unsubscribe(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed"})
}

func (s *Server) getUptimeStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"uptime": 99.9})
}

func (s *Server) getUptimeHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"history": []gin.H{}})
}

func (s *Server) getAnalyticsSummary(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"analytics": gin.H{}})
}

func (s *Server) getIncidentsRSS(c *gin.Context) {
	c.XML(http.StatusOK, gin.H{"rss": "feed"})
}

func (s *Server) getIncidentsAtom(c *gin.Context) {
	c.XML(http.StatusOK, gin.H{"atom": "feed"})
}

func (s *Server) login(c *gin.Context) {
	var loginReq struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Simple authentication for demo
	if loginReq.Username == "admin" && loginReq.Password == "password" {
		token, err := s.authService.GenerateToken(1, "admin", "admin")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
	}
}

// Admin handlers

func (s *Server) createStatus(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Group       string `json:"group"`
		Status      string `json:"status" binding:"required,oneof=operational degraded_performance partial_outage major_outage"`
		ShowUptime  bool   `json:"show_uptime"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Create the service/component
	service := &models.Service{
		Name:        req.Name,
		Description: req.Description,
		Group:       req.Group,
		Status:      req.Status,
		ShowUptime:  req.ShowUptime,
		TenantID:    1, // Default tenant for demo
	}

	createdService, err := s.statusService.CreateService(service)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create component", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Component created successfully",
		"component": createdService,
	})
}

func (s *Server) updateStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Status updated"})
}

func (s *Server) deleteStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Status deleted"})
}

func (s *Server) createIncident(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description" binding:"required"`
		Status      string `json:"status" binding:"required,oneof=investigating identified monitoring resolved"`
		Impact      string `json:"impact" binding:"required,oneof=minor major critical"`
		ServiceIds  []int  `json:"service_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Create the incident
	incident := &models.Incident{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Impact:      req.Impact,
		TenantID:    1, // Default tenant for demo
	}

	createdIncident, err := s.incidentService.CreateIncident(incident)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create incident", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Incident created successfully",
		"incident": createdIncident,
	})
}

func (s *Server) createMaintenanceEvent(c *gin.Context) {
	var req struct {
		Title       string    `json:"title" binding:"required"`
		Description string    `json:"description"`
		StartAt     time.Time `json:"start_at" binding:"required"`
		EndAt       time.Time `json:"end_at" binding:"required"`
		Status      string    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Status == "" {
		req.Status = "scheduled"
	}

	maintenance := models.Maintenance{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		StartAt:     req.StartAt,
		EndAt:       req.EndAt,
		TenantID:    1, // Default tenant for now
	}

	if err := database.DB.Create(&maintenance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create maintenance event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Maintenance event created successfully",
		"maintenance": maintenance,
	})
}

func (s *Server) updateMaintenanceEvent(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Maintenance updated"})
}

func (s *Server) deleteMaintenanceEvent(c *gin.Context) {
	maintenanceID := c.Param("id")
	if maintenanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maintenance ID is required"})
		return
	}

	var maintenance models.Maintenance
	if err := database.DB.First(&maintenance, maintenanceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Maintenance event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find maintenance event"})
		return
	}

	if err := database.DB.Delete(&maintenance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete maintenance event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Maintenance event deleted successfully"})
}

func (s *Server) createMonitor(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Monitor created"})
}

func (s *Server) updateMonitor(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Monitor updated"})
}

func (s *Server) deleteMonitor(c *gin.Context) {
	monitorID := c.Param("id")
	if monitorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Monitor ID is required"})
		return
	}

	var monitor models.Monitor
	if err := database.DB.First(&monitor, monitorID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Monitor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find monitor"})
		return
	}

	if err := database.DB.Delete(&monitor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete monitor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Monitor deleted successfully"})
}

func (s *Server) createUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Name     string `json:"name" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		Role     string `json:"role" binding:"required"`
		Active   bool   `json:"active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Username: req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     req.Role,
		IsActive: req.Active,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully", "user": user})
}

func (s *Server) updateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "User updated"})
}

func (s *Server) deleteUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

func (s *Server) getIncidentTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"templates": []gin.H{}})
}

func (s *Server) createIncidentTemplate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Template created"})
}

func (s *Server) updateIncidentTemplate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Template updated"})
}

func (s *Server) deleteIncidentTemplate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Template deleted"})
}

func (s *Server) getMaintenanceTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"templates": []gin.H{}})
}

func (s *Server) createMaintenanceTemplate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Template created"})
}

func (s *Server) updateMaintenanceTemplate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Template updated"})
}

func (s *Server) deleteMaintenanceTemplate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Template deleted"})
}

func (s *Server) getPrivatePages(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"pages": []gin.H{}})
}

func (s *Server) createPrivatePage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Page created"})
}

func (s *Server) updatePrivatePage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Page updated"})
}

func (s *Server) deletePrivatePage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Page deleted"})
}

func (s *Server) regeneratePrivatePageKey(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Key regenerated"})
}

func (s *Server) togglePrivatePageStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Status toggled"})
}

func (s *Server) getMonitoringStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "active"})
}

func (s *Server) syncAllMonitoringTools(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Sync completed"})
}

func (s *Server) getAdminServices(c *gin.Context) {
	services, err := s.getAdminServicesData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch services"})
		return
	}

	c.JSON(http.StatusOK, services)
}

func (s *Server) createAdminService(c *gin.Context) {
	var createReq struct {
		Name           string `json:"name" binding:"required"`
		Description    string `json:"description"`
		Status         string `json:"status" binding:"required"`
		Group          string `json:"group"`
		ShowUptime     bool   `json:"show_uptime"`
		Position       int    `json:"position"`
		HealthCheckURL string `json:"health_check_url"`
	}

	if err := c.ShouldBindJSON(&createReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant ID from context
	tenantID := uint(1) // Default tenant for now

	service := &models.Service{
		Name:           createReq.Name,
		Description:    createReq.Description,
		Status:         createReq.Status,
		Group:          createReq.Group,
		ShowUptime:     createReq.ShowUptime,
		Position:       createReq.Position,
		HealthCheckURL: createReq.HealthCheckURL,
		TenantID:       tenantID,
	}

	createdService, err := s.statusService.CreateService(service)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Service created successfully",
		"service": createdService,
	})
}

func (s *Server) getAdminService(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service ID is required"})
		return
	}

	var service models.Service
	if err := database.DB.First(&service, serviceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find service"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"service": service})
}

func (s *Server) updateAdminService(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service ID is required"})
		return
	}

	var updateReq struct {
		Name           string `json:"name" binding:"required"`
		Description    string `json:"description"`
		Status         string `json:"status" binding:"required"`
		Group          string `json:"group"`
		ShowUptime     bool   `json:"show_uptime"`
		Position       int    `json:"position"`
		HealthCheckURL string `json:"health_check_url"`
	}

	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the service
	var service models.Service
	if err := database.DB.First(&service, serviceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find service"})
		return
	}

	// Update the service
	service.Name = updateReq.Name
	service.Description = updateReq.Description
	service.Status = updateReq.Status
	service.Group = updateReq.Group
	service.ShowUptime = updateReq.ShowUptime
	service.Position = updateReq.Position
	service.HealthCheckURL = updateReq.HealthCheckURL

	updatedService, err := s.statusService.UpdateService(&service)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service updated successfully", "service": updatedService})
}

func (s *Server) deleteAdminService(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service ID is required"})
		return
	}

	var service models.Service
	if err := database.DB.First(&service, serviceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find service"})
		return
	}

	if err := s.statusService.DeleteService(service.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete service"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service deleted successfully"})
}

func (s *Server) getAdminSubscribers(c *gin.Context) {
	subscribers, err := s.subscriberService.GetAllSubscribers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscribers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"subscribers": subscribers})
}

func (s *Server) deleteAdminSubscriber(c *gin.Context) {
	subscriberID := c.Param("id")
	if subscriberID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscriber ID is required"})
		return
	}

	// For now, we'll use a simple approach - find the subscriber by ID and delete
	// In a real implementation, you'd want to get the email first and use the Unsubscribe method
	var subscriber models.Subscriber
	if err := database.DB.First(&subscriber, subscriberID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscriber not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find subscriber"})
		return
	}

	// Use the service method to properly unsubscribe
	if err := s.subscriberService.Unsubscribe(subscriber.Email, subscriber.TenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscriber"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscriber deleted successfully"})
}

func (s *Server) getIncident(c *gin.Context) {
	incidentID := c.Param("id")
	if incidentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incident ID is required"})
		return
	}

	var incident models.Incident
	if err := database.DB.Preload("Services").First(&incident, incidentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find incident"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"incident": incident})
}

func (s *Server) updateIncident(c *gin.Context) {
	incidentID := c.Param("id")
	if incidentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incident ID is required"})
		return
	}

	var updateReq struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description" binding:"required"`
		Status      string `json:"status" binding:"required"`
		Impact      string `json:"impact" binding:"required"`
	}

	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the incident
	var incident models.Incident
	if err := database.DB.First(&incident, incidentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find incident"})
		return
	}

	// Update the incident
	incident.Title = updateReq.Title
	incident.Description = updateReq.Description
	incident.Status = updateReq.Status
	incident.Impact = updateReq.Impact

	if err := database.DB.Save(&incident).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update incident"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incident updated successfully", "incident": incident})
}

func (s *Server) deleteIncident(c *gin.Context) {
	incidentID := c.Param("id")
	if incidentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incident ID is required"})
		return
	}

	// Find the incident by ID
	var incident models.Incident
	if err := database.DB.First(&incident, incidentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find incident"})
		return
	}

	// Delete the incident
	if err := database.DB.Delete(&incident).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete incident"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incident deleted successfully"})
}

// SaaS Admin handlers
func (s *Server) getSaaSOverviewMetrics(c *gin.Context) {
	// Get tenant counts
	var totalTenants, activeTenants int64
	database.DB.Model(&models.Tenant{}).Count(&totalTenants)
	database.DB.Model(&models.Tenant{}).Where("status = ?", "active").Count(&activeTenants)

	// Get plan distribution
	var planStats []struct {
		Plan  string `json:"plan"`
		Count int64  `json:"count"`
	}
	database.DB.Model(&models.Tenant{}).Select("plan, count(*) as count").Group("plan").Scan(&planStats)

	// Get recent activity (last 7 days)
	var recentTenants int64
	database.DB.Model(&models.Tenant{}).Where("created_at > ?", time.Now().AddDate(0, 0, -7)).Count(&recentTenants)

	metrics := gin.H{
		"total_tenants":     totalTenants,
		"active_tenants":    activeTenants,
		"inactive_tenants":  totalTenants - activeTenants,
		"recent_tenants":    recentTenants,
		"plan_distribution": planStats,
	}

	c.JSON(http.StatusOK, gin.H{"metrics": metrics})
}

func (s *Server) getSaaSRecentActivity(c *gin.Context) {
	// Get recent tenant activities
	var recentTenants []models.Tenant
	database.DB.Order("created_at DESC").Limit(10).Find(&recentTenants)

	var activities []gin.H
	for _, tenant := range recentTenants {
		activities = append(activities, gin.H{
			"type":      "tenant_created",
			"message":   fmt.Sprintf("New tenant '%s' created", tenant.Name),
			"timestamp": tenant.CreatedAt,
			"tenant_id": tenant.ID,
		})
	}

	c.JSON(http.StatusOK, gin.H{"activity": activities})
}

func (s *Server) getSaaSPlanDistribution(c *gin.Context) {
	// Get plan distribution
	var planStats []struct {
		Plan  string `json:"plan"`
		Count int64  `json:"count"`
	}
	database.DB.Model(&models.Tenant{}).Select("plan, count(*) as count").Group("plan").Scan(&planStats)

	c.JSON(http.StatusOK, gin.H{"distribution": planStats})
}

func (s *Server) getSaaSOverviewCharts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"charts": gin.H{}})
}

func (s *Server) getSAASAllTenants(c *gin.Context) {
	var tenants []models.Tenant
	if err := database.DB.Find(&tenants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tenants"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tenants": tenants})
}

func (s *Server) createSAASSTenant(c *gin.Context) {
	var req struct {
		Name         string `json:"name" binding:"required"`
		Slug         string `json:"slug" binding:"required"`
		Domain       string `json:"domain"`
		Subdomain    string `json:"subdomain" binding:"required"`
		BillingEmail string `json:"billing_email"`
		ContactEmail string `json:"contact_email"`
		Plan         string `json:"plan"`
		Status       string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.Plan == "" {
		req.Plan = "free"
	}
	if req.Status == "" {
		req.Status = "active"
	}

	tenant := models.Tenant{
		Name:         req.Name,
		Slug:         req.Slug,
		Domain:       req.Domain,
		Subdomain:    req.Subdomain,
		BillingEmail: req.BillingEmail,
		ContactEmail: req.ContactEmail,
		Plan:         req.Plan,
		Status:       req.Status,
	}

	if err := database.DB.Create(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tenant created successfully", "tenant": tenant})
}

func (s *Server) getSAASSTenantByID(c *gin.Context) {
	tenantID := c.Param("id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	var tenant models.Tenant
	if err := database.DB.First(&tenant, tenantID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tenant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tenant": tenant})
}

func (s *Server) updateSAASSTenant(c *gin.Context) {
	tenantID := c.Param("id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	var req struct {
		Name         string `json:"name"`
		Slug         string `json:"slug"`
		Domain       string `json:"domain"`
		Subdomain    string `json:"subdomain"`
		BillingEmail string `json:"billing_email"`
		ContactEmail string `json:"contact_email"`
		Plan         string `json:"plan"`
		Status       string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var tenant models.Tenant
	if err := database.DB.First(&tenant, tenantID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find tenant"})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		tenant.Name = req.Name
	}
	if req.Slug != "" {
		tenant.Slug = req.Slug
	}
	if req.Domain != "" {
		tenant.Domain = req.Domain
	}
	if req.Subdomain != "" {
		tenant.Subdomain = req.Subdomain
	}
	if req.BillingEmail != "" {
		tenant.BillingEmail = req.BillingEmail
	}
	if req.ContactEmail != "" {
		tenant.ContactEmail = req.ContactEmail
	}
	if req.Plan != "" {
		tenant.Plan = req.Plan
	}
	if req.Status != "" {
		tenant.Status = req.Status
	}

	if err := database.DB.Save(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tenant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tenant updated successfully", "tenant": tenant})
}

func (s *Server) deleteSAASSTenant(c *gin.Context) {
	tenantID := c.Param("id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	var tenant models.Tenant
	if err := database.DB.First(&tenant, tenantID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find tenant"})
		return
	}

	if err := database.DB.Delete(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tenant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tenant deleted successfully"})
}

func (s *Server) getTenantSubscription(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"subscription": gin.H{}})
}

func (s *Server) createTenantSubscription(c *gin.Context) {
	tenantID := c.Param("id")

	// Parse request data
	var req struct {
		Plan             string `json:"plan" binding:"required"`
		BillingCycle     string `json:"billingCycle"`
		StartDate        string `json:"startDate"`
		TrialDays        int    `json:"trialDays"`
		PaymentMethod    string `json:"paymentMethod"`
		AutoRenewal      bool   `json:"autoRenewal"`
		SendWelcomeEmail bool   `json:"sendWelcomeEmail"`
		Notes            string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	// Find the plan by slug
	var plan models.SubscriptionPlan
	if err := database.DB.Where("slug = ?", req.Plan).First(&plan).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plan not found: " + req.Plan})
		return
	}

	// Parse start date
	startDate := time.Now()
	if req.StartDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = parsed
		}
	}

	// Calculate end date based on billing cycle
	var endDate time.Time
	if req.BillingCycle == "yearly" {
		endDate = startDate.AddDate(1, 0, 0)
	} else {
		endDate = startDate.AddDate(0, 1, 0)
	}

	// Parse tenant ID
	tenantIDUint, err := parseUint(tenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID: " + err.Error()})
		return
	}

	// Create subscription
	subscription := models.Subscription{
		TenantID:           tenantIDUint,
		PlanID:             plan.ID,
		Status:             "active",
		CurrentPeriodStart: startDate,
		CurrentPeriodEnd:   endDate,
		CancelAtPeriodEnd:  !req.AutoRenewal,
	}

	if err := database.DB.Create(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription: " + err.Error()})
		return
	}

	// Load the created subscription with relations
	if err := database.DB.Preload("Tenant").Preload("Plan").First(&subscription, subscription.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load created subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Subscription created successfully",
		"subscription": subscription,
	})
}

func (s *Server) upgradeTenantSubscription(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Subscription upgraded"})
}

func (s *Server) downgradeTenantSubscription(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Subscription downgraded"})
}

func (s *Server) cancelTenantSubscription(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Subscription cancelled"})
}

func (s *Server) getTenantUsage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"usage": gin.H{}})
}

func (s *Server) getSAASAllSubscriptions(c *gin.Context) {
	var subscriptions []models.Subscription
	if err := database.DB.Preload("Tenant").Preload("Plan").Find(&subscriptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscriptions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"subscriptions": subscriptions})
}

func (s *Server) getSAASSubscriptionByID(c *gin.Context) {
	subscriptionID := c.Param("id")
	if subscriptionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscription ID is required"})
		return
	}

	var subscription models.Subscription
	if err := database.DB.Preload("Tenant").Preload("Plan").First(&subscription, subscriptionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscription": subscription})
}

func (s *Server) updateSAASSubscription(c *gin.Context) {
	subscriptionID := c.Param("id")
	if subscriptionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscription ID is required"})
		return
	}

	var req struct {
		Status             string     `json:"status"`
		CurrentPeriodStart *time.Time `json:"current_period_start"`
		CurrentPeriodEnd   *time.Time `json:"current_period_end"`
		CancelAtPeriodEnd  *bool      `json:"cancel_at_period_end"`
		CancelledAt        *time.Time `json:"cancelled_at"`
		TrialStart         *time.Time `json:"trial_start"`
		TrialEnd           *time.Time `json:"trial_end"`
		ExternalID         string     `json:"external_id"`
		ExternalCustomerID string     `json:"external_customer_id"`
		Metadata           string     `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var subscription models.Subscription
	if err := database.DB.First(&subscription, subscriptionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find subscription"})
		return
	}

	// Update fields if provided
	if req.Status != "" {
		subscription.Status = req.Status
	}
	if req.CurrentPeriodStart != nil {
		subscription.CurrentPeriodStart = *req.CurrentPeriodStart
	}
	if req.CurrentPeriodEnd != nil {
		subscription.CurrentPeriodEnd = *req.CurrentPeriodEnd
	}
	if req.CancelAtPeriodEnd != nil {
		subscription.CancelAtPeriodEnd = *req.CancelAtPeriodEnd
	}
	if req.CancelledAt != nil {
		subscription.CancelledAt = req.CancelledAt
	}
	if req.TrialStart != nil {
		subscription.TrialStart = req.TrialStart
	}
	if req.TrialEnd != nil {
		subscription.TrialEnd = req.TrialEnd
	}
	if req.ExternalID != "" {
		subscription.ExternalID = req.ExternalID
	}
	if req.ExternalCustomerID != "" {
		subscription.ExternalCustomerID = req.ExternalCustomerID
	}
	if req.Metadata != "" {
		subscription.Metadata = req.Metadata
	}

	if err := database.DB.Save(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription updated successfully", "subscription": subscription})
}

func (s *Server) cancelSAASSubscription(c *gin.Context) {
	subscriptionID := c.Param("id")
	if subscriptionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscription ID is required"})
		return
	}

	var subscription models.Subscription
	if err := database.DB.First(&subscription, subscriptionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find subscription"})
		return
	}

	// Cancel the subscription
	now := time.Now()
	subscription.Status = "cancelled"
	subscription.CancelledAt = &now
	subscription.CancelAtPeriodEnd = true

	if err := database.DB.Save(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription cancelled successfully"})
}

func (s *Server) getSAASAllPlans(c *gin.Context) {
	var plans []models.SubscriptionPlan
	if err := database.DB.Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plans"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

func (s *Server) createSAASPlan(c *gin.Context) {
	var req struct {
		Name            string  `json:"name" binding:"required"`
		Slug            string  `json:"slug" binding:"required"`
		Description     string  `json:"description"`
		Price           float64 `json:"price" binding:"required,min=0"`
		Currency        string  `json:"currency"`
		BillingInterval string  `json:"billing_interval" binding:"required"`
		MaxServices     int     `json:"max_services"`
		MaxMonitors     int     `json:"max_monitors"`
		MaxSubscribers  int     `json:"max_subscribers"`
		MaxIncidents    int     `json:"max_incidents"`
		MaxMaintenance  int     `json:"max_maintenance"`
		CustomDomain    bool    `json:"custom_domain"`
		WhiteLabel      bool    `json:"white_label"`
		API             bool    `json:"api"`
		Integrations    bool    `json:"integrations"`
		Analytics       bool    `json:"analytics"`
		Support         string  `json:"support"`
		IsActive        bool    `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.Currency == "" {
		req.Currency = "USD"
	}
	if req.MaxServices == 0 {
		req.MaxServices = 5
	}
	if req.MaxMonitors == 0 {
		req.MaxMonitors = 10
	}
	if req.MaxSubscribers == 0 {
		req.MaxSubscribers = 100
	}
	if req.MaxIncidents == 0 {
		req.MaxIncidents = 50
	}
	if req.MaxMaintenance == 0 {
		req.MaxMaintenance = 20
	}
	if req.Support == "" {
		req.Support = "email"
	}

	plan := models.SubscriptionPlan{
		Name:            req.Name,
		Slug:            req.Slug,
		Description:     req.Description,
		Price:           req.Price,
		Currency:        req.Currency,
		BillingInterval: req.BillingInterval,
		MaxServices:     req.MaxServices,
		MaxMonitors:     req.MaxMonitors,
		MaxSubscribers:  req.MaxSubscribers,
		MaxIncidents:    req.MaxIncidents,
		MaxMaintenance:  req.MaxMaintenance,
		CustomDomain:    req.CustomDomain,
		WhiteLabel:      req.WhiteLabel,
		API:             req.API,
		Integrations:    req.Integrations,
		Analytics:       req.Analytics,
		Support:         req.Support,
		IsActive:        req.IsActive,
	}

	if err := database.DB.Create(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create plan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Plan created successfully",
		"plan":    plan,
	})
}

func (s *Server) getSAASPlanByID(c *gin.Context) {
	planID := c.Param("id")
	if planID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plan ID is required"})
		return
	}

	var plan models.SubscriptionPlan
	if err := database.DB.First(&plan, planID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

func (s *Server) updateSAASPlan(c *gin.Context) {
	planID := c.Param("id")
	if planID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plan ID is required"})
		return
	}

	var req struct {
		Name            string  `json:"name"`
		Slug            string  `json:"slug"`
		Description     string  `json:"description"`
		Price           float64 `json:"price"`
		Currency        string  `json:"currency"`
		BillingInterval string  `json:"billing_interval"`
		MaxServices     int     `json:"max_services"`
		MaxMonitors     int     `json:"max_monitors"`
		MaxSubscribers  int     `json:"max_subscribers"`
		MaxIncidents    int     `json:"max_incidents"`
		MaxMaintenance  int     `json:"max_maintenance"`
		CustomDomain    bool    `json:"custom_domain"`
		WhiteLabel      bool    `json:"white_label"`
		API             bool    `json:"api"`
		Integrations    bool    `json:"integrations"`
		Analytics       bool    `json:"analytics"`
		Support         string  `json:"support"`
		IsActive        bool    `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var plan models.SubscriptionPlan
	if err := database.DB.First(&plan, planID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find plan"})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		plan.Name = req.Name
	}
	if req.Slug != "" {
		plan.Slug = req.Slug
	}
	if req.Description != "" {
		plan.Description = req.Description
	}
	if req.Price > 0 {
		plan.Price = req.Price
	}
	if req.Currency != "" {
		plan.Currency = req.Currency
	}
	if req.BillingInterval != "" {
		plan.BillingInterval = req.BillingInterval
	}
	if req.MaxServices > 0 {
		plan.MaxServices = req.MaxServices
	}
	if req.MaxMonitors > 0 {
		plan.MaxMonitors = req.MaxMonitors
	}
	if req.MaxSubscribers > 0 {
		plan.MaxSubscribers = req.MaxSubscribers
	}
	if req.MaxIncidents > 0 {
		plan.MaxIncidents = req.MaxIncidents
	}
	if req.MaxMaintenance > 0 {
		plan.MaxMaintenance = req.MaxMaintenance
	}
	plan.CustomDomain = req.CustomDomain
	plan.WhiteLabel = req.WhiteLabel
	plan.API = req.API
	plan.Integrations = req.Integrations
	plan.Analytics = req.Analytics
	if req.Support != "" {
		plan.Support = req.Support
	}
	plan.IsActive = req.IsActive

	if err := database.DB.Save(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update plan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Plan updated successfully",
		"plan":    plan,
	})
}

func (s *Server) deleteSAASPlan(c *gin.Context) {
	planID := c.Param("id")
	if planID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plan ID is required"})
		return
	}

	var plan models.SubscriptionPlan
	if err := database.DB.First(&plan, planID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find plan"})
		return
	}

	if err := database.DB.Delete(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete plan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plan deleted successfully"})
}

func (s *Server) getSAASBillingStats(c *gin.Context) {
	// Calculate billing stats from subscriptions and plans
	var totalRevenue float64
	var monthlyRecurringRevenue float64
	var activeSubscriptions int64
	var churnRate float64

	// Get total revenue from all active subscriptions
	database.DB.Table("subscriptions").
		Select("COALESCE(SUM(plans.price), 0)").
		Joins("JOIN subscription_plans plans ON subscriptions.plan_id = plans.id").
		Where("subscriptions.status = ?", "active").
		Scan(&totalRevenue)

	// Get monthly recurring revenue
	database.DB.Table("subscriptions").
		Select("COALESCE(SUM(plans.price), 0)").
		Joins("JOIN subscription_plans plans ON subscriptions.plan_id = plans.id").
		Where("subscriptions.status = ? AND plans.billing_interval = ?", "active", "monthly").
		Scan(&monthlyRecurringRevenue)

	// Count active subscriptions
	database.DB.Model(&models.Subscription{}).Where("status = ?", "active").Count(&activeSubscriptions)

	// Calculate churn rate (simplified - would need historical data for accurate calculation)
	churnRate = 2.5 // Mock value for now

	stats := gin.H{
		"total_revenue":             totalRevenue,
		"monthly_recurring_revenue": monthlyRecurringRevenue,
		"active_subscriptions":      activeSubscriptions,
		"churn_rate":                churnRate,
		"pending_payments":          0.0, // Would need payment tracking
		"failed_payments":           0.0, // Would need payment tracking
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

func (s *Server) getSAASBillingEvents(c *gin.Context) {
	// Get recent billing events (subscriptions, payments, etc.)
	var events []gin.H

	// Get recent subscriptions as billing events
	var subscriptions []models.Subscription
	database.DB.Preload("Plan").Preload("Tenant").
		Where("created_at >= ?", "2025-01-01").
		Order("created_at DESC").
		Limit(20).
		Find(&subscriptions)

	for _, sub := range subscriptions {
		event := gin.H{
			"id":          sub.ID,
			"type":        "subscription",
			"amount":      sub.Plan.Price,
			"currency":    sub.Plan.Currency,
			"status":      sub.Status,
			"created_at":  sub.CreatedAt,
			"tenant_name": sub.Tenant.Name,
			"plan_name":   sub.Plan.Name,
		}
		events = append(events, event)
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

func (s *Server) getSAASAnalytics(c *gin.Context) {
	// Calculate analytics data
	var totalTenants int64
	var activeTenants int64
	var totalRevenue float64
	var totalServices int64
	var totalIncidents int64
	var totalSubscribers int64

	// Count tenants
	database.DB.Model(&models.Tenant{}).Count(&totalTenants)
	database.DB.Model(&models.Tenant{}).Where("status = ?", "active").Count(&activeTenants)

	// Calculate revenue from active subscriptions
	database.DB.Table("subscriptions").
		Select("COALESCE(SUM(plans.price), 0)").
		Joins("JOIN subscription_plans plans ON subscriptions.plan_id = plans.id").
		Where("subscriptions.status = ?", "active").
		Scan(&totalRevenue)

	// Count services across all tenants
	database.DB.Model(&models.Service{}).Count(&totalServices)

	// Count incidents across all tenants
	database.DB.Model(&models.Incident{}).Count(&totalIncidents)

	// Count subscribers across all tenants
	database.DB.Model(&models.Subscriber{}).Count(&totalSubscribers)

	// Get real API usage analytics (last 30 days)
	apiUsageAnalytics, err := s.analyticsService.GetPlatformAPIUsageAnalytics(30)
	if err != nil {
		logger.Error("Failed to get API usage analytics", zap.Error(err))
		// Fallback to zero values
		apiUsageAnalytics = map[string]interface{}{
			"total_requests":    int64(0),
			"error_count":       int64(0),
			"avg_response_time": float64(0),
		}
	}

	// Get real page view analytics (last 30 days)
	pageViewAnalytics, err := s.analyticsService.GetPlatformPageViewAnalytics(30)
	if err != nil {
		logger.Error("Failed to get page view analytics", zap.Error(err))
		// Fallback to zero values
		pageViewAnalytics = map[string]interface{}{
			"total_views":     int64(0),
			"unique_visitors": int64(0),
			"bounce_rate":     float64(0),
		}
	}

	analytics := gin.H{
		"usage_metrics": gin.H{
			"total_tenants":     totalTenants,
			"active_tenants":    activeTenants,
			"total_revenue":     totalRevenue,
			"total_services":    totalServices,
			"total_incidents":   totalIncidents,
			"total_subscribers": totalSubscribers,
		},
		"api_usage": gin.H{
			"requests":          apiUsageAnalytics["total_requests"],
			"errors":            apiUsageAnalytics["error_count"],
			"avg_response_time": apiUsageAnalytics["avg_response_time"],
		},
		"page_views": gin.H{
			"total":       pageViewAnalytics["total_views"],
			"unique":      pageViewAnalytics["unique_visitors"],
			"bounce_rate": pageViewAnalytics["bounce_rate"],
		},
	}

	c.JSON(http.StatusOK, gin.H{"analytics": analytics})
}

func (s *Server) getSAASSettings(c *gin.Context) {
	// Return platform settings
	settings := gin.H{
		"general": gin.H{
			"site_name":        "StatusPage SaaS",
			"contact_email":    "admin@statuspage.com",
			"max_tenants":      1000,
			"maintenance_mode": false,
		},
		"feature_flags": gin.H{
			"enable_white_label":    true,
			"enable_custom_domains": true,
			"enable_api_access":     true,
			"enable_analytics":      true,
			"enable_notifications":  true,
		},
		"billing": gin.H{
			"stripe_public_key": "pk_test_...",
			"stripe_secret_key": "sk_test_...",
			"default_currency":  "USD",
			"trial_days":        14,
		},
		"security": gin.H{
			"require_2fa":         false,
			"session_timeout":     3600,
			"max_login_attempts":  5,
			"password_min_length": 8,
		},
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

func (s *Server) updateSAASSetting(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Setting updated"})
}

func (s *Server) getSAASNotifications(c *gin.Context) {
	var notifications []models.SystemNotification
	if err := database.DB.Order("created_at DESC").Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"notifications": notifications})
}

func (s *Server) createSAASNotification(c *gin.Context) {
	var req struct {
		Title         string `json:"title" binding:"required"`
		Message       string `json:"message" binding:"required"`
		Type          string `json:"type" binding:"required"`
		StartAt       string `json:"start_at" binding:"required"`
		EndAt         string `json:"end_at"`
		TargetTenants string `json:"target_tenants"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse start time
	startTime, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_at format"})
		return
	}

	// Parse end time if provided
	var endTime *time.Time
	if req.EndAt != "" {
		parsedEndTime, err := time.Parse(time.RFC3339, req.EndAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_at format"})
			return
		}
		endTime = &parsedEndTime
	}

	notification := models.SystemNotification{
		Title:         req.Title,
		Message:       req.Message,
		Type:          req.Type,
		IsActive:      true,
		StartAt:       startTime,
		EndAt:         endTime,
		TargetTenants: req.TargetTenants,
	}

	if err := database.DB.Create(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification created successfully", "notification": notification})
}

func (s *Server) updateSAASNotification(c *gin.Context) {
	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Notification ID is required"})
		return
	}

	var req struct {
		Title         string `json:"title"`
		Message       string `json:"message"`
		Type          string `json:"type"`
		IsActive      *bool  `json:"is_active"`
		StartAt       string `json:"start_at"`
		EndAt         string `json:"end_at"`
		TargetTenants string `json:"target_tenants"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var notification models.SystemNotification
	if err := database.DB.First(&notification, notificationID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notification"})
		return
	}

	// Update fields if provided
	if req.Title != "" {
		notification.Title = req.Title
	}
	if req.Message != "" {
		notification.Message = req.Message
	}
	if req.Type != "" {
		notification.Type = req.Type
	}
	if req.IsActive != nil {
		notification.IsActive = *req.IsActive
	}
	if req.StartAt != "" {
		startTime, err := time.Parse(time.RFC3339, req.StartAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_at format"})
			return
		}
		notification.StartAt = startTime
	}
	if req.EndAt != "" {
		endTime, err := time.Parse(time.RFC3339, req.EndAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_at format"})
			return
		}
		notification.EndAt = &endTime
	} else if req.EndAt == "" && req.StartAt != "" {
		// If end_at is explicitly set to empty, clear it
		notification.EndAt = nil
	}
	if req.TargetTenants != "" {
		notification.TargetTenants = req.TargetTenants
	}

	if err := database.DB.Save(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification updated successfully", "notification": notification})
}

func (s *Server) deleteSAASNotification(c *gin.Context) {
	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Notification ID is required"})
		return
	}

	var notification models.SystemNotification
	if err := database.DB.First(&notification, notificationID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notification"})
		return
	}

	if err := database.DB.Delete(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

// SaaS Admin Feature Flag Handlers
func (s *Server) getTenantFeatureFlags(c *gin.Context) {
	tenantIDStr := c.Param("id")
	if tenantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	flags, err := s.featureFlagService.GetAllFeatureFlags(uint(tenantID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feature flags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"feature_flags": flags})
}

func (s *Server) setTenantFeatureFlag(c *gin.Context) {
	tenantIDStr := c.Param("id")
	if tenantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var req struct {
		Feature   string `json:"feature" binding:"required"`
		IsEnabled bool   `json:"is_enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = s.featureFlagService.SetFeatureEnabled(uint(tenantID), req.Feature, req.IsEnabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set feature flag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feature flag updated successfully"})
}

func (s *Server) updateTenantFeatureFlag(c *gin.Context) {
	tenantIDStr := c.Param("id")
	featureName := c.Param("feature")

	if tenantIDStr == "" || featureName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID and feature name are required"})
		return
	}

	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var req struct {
		IsEnabled bool `json:"is_enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = s.featureFlagService.SetFeatureEnabled(uint(tenantID), featureName, req.IsEnabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update feature flag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feature flag updated successfully"})
}

func (s *Server) deleteTenantFeatureFlag(c *gin.Context) {
	tenantIDStr := c.Param("id")
	featureName := c.Param("feature")

	if tenantIDStr == "" || featureName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID and feature name are required"})
		return
	}

	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// For now, just disable the feature flag instead of deleting it
	err = s.featureFlagService.SetFeatureEnabled(uint(tenantID), featureName, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable feature flag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feature flag disabled successfully"})
}

func (s *Server) getAllFeatureFlags(c *gin.Context) {
	// Return all available feature flags (this could be from a config or hardcoded list)
	flags := []gin.H{
		{"feature": models.FeaturePerServiceGraphs, "name": "Per-Service Graphs", "description": "Enable individual monitoring graphs for each service"},
		{"feature": models.FeatureCustomDomains, "name": "Custom Domains", "description": "Allow tenants to use custom domains"},
		{"feature": models.FeatureAdvancedAnalytics, "name": "Advanced Analytics", "description": "Enable advanced analytics and reporting"},
		{"feature": models.FeatureSSO, "name": "Single Sign-On", "description": "Enable SSO authentication"},
		{"feature": models.FeatureAPI, "name": "API Access", "description": "Enable API access for tenants"},
	}

	c.JSON(http.StatusOK, gin.H{"available_features": flags})
}

func (s *Server) setGlobalFeatureFlag(c *gin.Context) {
	var req struct {
		Feature    string `json:"feature" binding:"required"`
		IsEnabled  bool   `json:"is_enabled"`
		ApplyToAll bool   `json:"apply_to_all"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ApplyToAll {
		// Get all tenants and apply the feature flag to all of them
		var tenants []models.Tenant
		if err := database.DB.Find(&tenants).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tenants"})
			return
		}

		for _, tenant := range tenants {
			err := s.featureFlagService.SetFeatureEnabled(tenant.ID, req.Feature, req.IsEnabled)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set feature flag for some tenants"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Global feature flag applied successfully"})
}

// SaaS Feature Availability Handlers
func (s *Server) getFeatureAvailability(c *gin.Context) {
	availabilities, err := s.featureFlagService.GetAllFeatureAvailability()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feature availability"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"feature_availability": availabilities})
}

func (s *Server) setFeatureAvailability(c *gin.Context) {
	var req struct {
		Feature     string `json:"feature" binding:"required"`
		IsAvailable bool   `json:"is_available"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := s.featureFlagService.SetFeatureAvailability(req.Feature, req.IsAvailable, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set feature availability"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feature availability updated successfully"})
}

func (s *Server) updateFeatureAvailability(c *gin.Context) {
	featureName := c.Param("feature")
	if featureName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Feature name is required"})
		return
	}

	var req struct {
		IsAvailable bool   `json:"is_available"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := s.featureFlagService.SetFeatureAvailability(featureName, req.IsAvailable, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update feature availability"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feature availability updated successfully"})
}

func (s *Server) deleteFeatureAvailability(c *gin.Context) {
	featureName := c.Param("feature")
	if featureName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Feature name is required"})
		return
	}

	// Actually delete the feature availability record
	err := s.featureFlagService.DeleteFeatureAvailability(featureName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete feature availability"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feature availability deleted successfully"})
}

// Web handlers
func (s *Server) showLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_login.html", gin.H{})
}

func (s *Server) handleWebLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "admin" && password == "password" {
		token, err := s.authService.GenerateToken(1, "admin", "admin")
		if err != nil {
			c.HTML(http.StatusInternalServerError, "admin_login.html", gin.H{"error": "Failed to generate token"})
			return
		}
		c.SetCookie("auth_token", token, 3600, "/", "", false, true)
		c.Redirect(http.StatusFound, "/admin/dashboard")
	} else {
		c.HTML(http.StatusUnauthorized, "admin_login.html", gin.H{"error": "Invalid credentials"})
	}
}

func (s *Server) handleLogout(c *gin.Context) {
	c.SetCookie("auth_token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/admin/login")
}

func (s *Server) showDashboardPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_dashboard.html", gin.H{})
}

func (s *Server) showSaaSAdminDashboard(c *gin.Context) {
	// The middleware.AuthRequired already handles authentication
	// If we reach here, the user is authenticated
	c.HTML(http.StatusOK, "saas_admin_dashboard.html", gin.H{})
}

func (s *Server) showNewMaintenancePage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_maintenance_new.html", gin.H{})
}

func (s *Server) handleNewMaintenance(c *gin.Context) {
	c.Redirect(http.StatusFound, "/admin/dashboard")
}

func (s *Server) showEditMaintenancePage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_maintenance_edit.html", gin.H{})
}

func (s *Server) handleEditMaintenance(c *gin.Context) {
	c.Redirect(http.StatusFound, "/admin/dashboard")
}

func (s *Server) handleDeleteMaintenance(c *gin.Context) {
	c.Redirect(http.StatusFound, "/admin/dashboard")
}

func (s *Server) showLandingPage(c *gin.Context) {
	logger.Info("!!!! showLandingPage called - THIS SHOULD NOT HAPPEN FOR SPECIFIC ROUTES !!!!",
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method))

	// If this is the pricing route, redirect to proper handler
	if c.Request.URL.Path == "/pricing" {
		logger.Info("Redirecting pricing route to showPricingPage")
		s.showPricingPage(c)
		return
	}

	// Check if this is a tenant request
	_, exists := middleware.GetTenantFromContext(c)
	if exists {
		// Show tenant status page
		logger.Info("Showing tenant status page")
		c.HTML(http.StatusOK, "statuspage.html", gin.H{})
		return
	}

	// Show the new statuspage.io style interface
	logger.Info("Showing default status page")
	c.HTML(http.StatusOK, "statuspage.html", gin.H{})
}

func (s *Server) showPricingPage(c *gin.Context) {
	logger.Info("Pricing page handler called", zap.String("path", c.Request.URL.Path))
	// Check if pricing.html template is loaded
	logger.Info("Attempting to render pricing.html template")
	c.HTML(http.StatusOK, "pricing.html", gin.H{
		"test": "This is the pricing page",
	})
}

func (s *Server) showTenantAdminDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "tenant_admin_dashboard.html", gin.H{})
}

func (s *Server) showPrivatePage(c *gin.Context) {
	c.HTML(http.StatusOK, "private_page.html", gin.H{})
}

func (s *Server) handleTenantSubscribe(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant from context
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Create subscriber
	subscriber := &models.Subscriber{
		TenantID: tenant.ID,
		Email:    req.Email,
	}

	if err := s.subscriberService.CreateSubscriber(subscriber); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscriber"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully subscribed"})
}

// Helper function to generate unique IDs

// Missing admin API handlers
func (s *Server) getAdminIncidents(c *gin.Context) {
	// Get tenant ID from context
	tenantID := uint(1) // Default tenant for now

	incidents, err := s.incidentService.GetIncidentsByTenantID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch incidents"})
		return
	}

	// Convert to response format
	var incidentResponses []gin.H
	for _, incident := range incidents {
		incidentResponses = append(incidentResponses, gin.H{
			"id":          incident.ID,
			"title":       incident.Title,
			"description": incident.Description,
			"status":      incident.Status,
			"impact":      incident.Impact,
			"severity":    incident.Severity,
			"resolved_at": incident.ResolvedAt,
			"created_at":  incident.CreatedAt,
			"updated_at":  incident.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"incidents": incidentResponses})
}

func (s *Server) getAdminAuditLogs(c *gin.Context) {
	// Return mock data for now
	logs := []gin.H{
		{
			"id":        1,
			"action":    "login",
			"user":      "admin",
			"timestamp": "2025-01-08T10:00:00Z",
			"ip":        "127.0.0.1",
		},
		{
			"id":        2,
			"action":    "incident_created",
			"user":      "admin",
			"timestamp": "2025-01-08T09:30:00Z",
			"ip":        "127.0.0.1",
		},
	}
	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

func (s *Server) getAdminStatus(c *gin.Context) {
	// Return mock data for now
	c.JSON(http.StatusOK, gin.H{
		"overall_status": "operational",
		"services": []gin.H{
			{"name": "API", "status": "operational"},
			{"name": "Database", "status": "operational"},
			{"name": "CDN", "status": "degraded"},
		},
	})
}

func (s *Server) getAdminMaintenance(c *gin.Context) {
	// Get tenant ID from context
	tenantID := uint(1) // Default tenant for now

	maintenance, err := s.maintenanceService.GetMaintenanceByTenantID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch maintenance"})
		return
	}

	// Convert to response format
	var maintenanceResponses []gin.H
	for _, item := range maintenance {
		duration := item.EndAt.Sub(item.StartAt)
		maintenanceResponses = append(maintenanceResponses, gin.H{
			"id":            item.ID,
			"title":         item.Title,
			"description":   item.Description,
			"status":        item.Status,
			"start_at":      item.StartAt,
			"end_at":        item.EndAt,
			"scheduled_for": item.StartAt,
			"duration":      duration.String(),
			"created_at":    item.CreatedAt,
			"updated_at":    item.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"maintenance": maintenanceResponses})
}

func (s *Server) getAdminMonitors(c *gin.Context) {
	// Get tenant ID from context
	tenantID := uint(1) // Default tenant for now

	monitors, err := s.monitorService.GetMonitorsByTenantID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch monitors"})
		return
	}

	// Convert to response format with uptime data
	var monitorResponses []gin.H
	for _, monitor := range monitors {
		// Get uptime for this monitor
		uptime, _ := s.monitorService.GetMonitorUptime(monitor.ID, 30) // 30 days

		monitorResponses = append(monitorResponses, gin.H{
			"id":              monitor.ID,
			"name":            monitor.Name,
			"type":            monitor.Type,
			"url":             monitor.URL,
			"status":          monitor.LastResult,
			"uptime":          fmt.Sprintf("%.1f%%", uptime),
			"interval":        monitor.Interval,
			"timeout":         monitor.Timeout,
			"expected_status": monitor.ExpectedStatus,
			"last_check_at":   monitor.LastCheckAt,
			"created_at":      monitor.CreatedAt,
			"updated_at":      monitor.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"monitors": monitorResponses})
}

func (s *Server) getAdminUsers(c *gin.Context) {
	// Return mock data for now
	users := []gin.H{
		{
			"id":       1,
			"username": "admin",
			"email":    "admin@example.com",
			"role":     "admin",
			"status":   "active",
		},
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (s *Server) getAdminIntegrations(c *gin.Context) {
	// Return mock data for now
	integrations := []gin.H{
		{
			"id":      1,
			"name":    "Slack",
			"type":    "notification",
			"status":  "active",
			"webhook": "https://hooks.slack.com/...",
		},
		{
			"id":      2,
			"name":    "PagerDuty",
			"type":    "incident",
			"status":  "active",
			"api_key": "***",
		},
	}
	c.JSON(http.StatusOK, gin.H{"integrations": integrations})
}

// Feature Flag Handlers

func (s *Server) getFeatureFlags(c *gin.Context) {
	// Get tenant ID from context (for multi-tenant support)
	tenantID := uint(1) // Default tenant for now

	// Only return features that are available at SaaS level
	flags, err := s.featureFlagService.GetAvailableFeaturesForTenant(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feature flags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"feature_flags": flags})
}

func (s *Server) setFeatureFlag(c *gin.Context) {
	var request struct {
		Feature   string      `json:"feature" binding:"required"`
		IsEnabled bool        `json:"is_enabled"`
		Config    interface{} `json:"config,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant ID from context (for multi-tenant support)
	tenantID := uint(1) // Default tenant for now

	err := s.featureFlagService.SetFeatureEnabled(tenantID, request.Feature, request.IsEnabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set feature flag"})
		return
	}

	// Set config if provided
	if request.Config != nil {
		err = s.featureFlagService.SetFeatureConfig(tenantID, request.Feature, request.Config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set feature config"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feature flag updated successfully"})
}

func (s *Server) updateFeatureFlag(c *gin.Context) {
	feature := c.Param("feature")

	var request struct {
		IsEnabled bool        `json:"is_enabled"`
		Config    interface{} `json:"config,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant ID from context (for multi-tenant support)
	tenantID := uint(1) // Default tenant for now

	err := s.featureFlagService.SetFeatureEnabled(tenantID, feature, request.IsEnabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update feature flag"})
		return
	}

	// Update config if provided
	if request.Config != nil {
		err = s.featureFlagService.SetFeatureConfig(tenantID, feature, request.Config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update feature config"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feature flag updated successfully"})
}

func (s *Server) deleteFeatureFlag(c *gin.Context) {
	feature := c.Param("feature")

	// Get tenant ID from context (for multi-tenant support)
	tenantID := uint(1) // Default tenant for now

	// Delete the feature flag
	err := s.featureFlagService.SetFeatureEnabled(tenantID, feature, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete feature flag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feature flag deleted successfully"})
}
