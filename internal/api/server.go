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

	"github.com/gin-gonic/gin"
	"github.com/enterprise-status/statuspage/internal/api/middleware"
	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/internal/services"
	"github.com/enterprise-status/statuspage/pkg/email"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Server struct {
	config          *config.Config
	router          *gin.Engine
	httpServer      *http.Server
	statusService   *services.StatusService
	incidentService *services.IncidentService
	authService     *services.AuthService
	maintenanceService *services.MaintenanceService
	subscriberService  *services.SubscriberService
	notificationService *services.NotificationService
	monitorService *services.MonitorService
	uptimeService *services.UptimeService
	rbacService    *services.RBACService
	auditService   *services.AuditService
	brandingService *services.BrandingService
	templateService *services.TemplateService
	integrationService *services.IntegrationService
	privatePageService *services.PrivatePageService
	monitoringIntegrationService *services.MonitoringIntegrationService
}

func NewServer(cfg *config.Config) *Server {
	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	server := &Server{
		config:          cfg,
		router:          gin.New(),
		statusService:   services.NewStatusService(),
		incidentService: services.NewIncidentService(),
		authService:     services.NewAuthService(cfg),
		maintenanceService: services.NewMaintenanceService(),
		subscriberService:  services.NewSubscriberService(),
		monitorService: services.NewMonitorService(),
		uptimeService:      services.NewUptimeService(),
		rbacService:    services.NewRBACService(),
		auditService:   services.NewAuditService(),
		brandingService: services.NewBrandingService(),
		templateService: services.NewTemplateService(),
		integrationService: services.NewIntegrationService(),
		privatePageService: services.NewPrivatePageService(),
		monitoringIntegrationService: services.NewMonitoringIntegrationService(),
	}

	// Initialize services
	var emailSender email.Sender
	if cfg.Email.UseSMTP {
		smtpConfig := email.SMTPConfig{
			Host:     cfg.Email.SMTPHost,
			Port:     cfg.Email.SMTPPort,
			Username: cfg.Email.SMTPUsername,
			Password: cfg.Email.SMTPPassword,
			From:     cfg.Email.SMTPFrom,
			UseTLS:   cfg.Email.SMTPUseTLS,
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
	// Middleware
	s.router.Use(gin.Logger())
	s.router.Use(gin.Recovery())

	// Load HTML templates
	s.router.LoadHTMLGlob("web/templates/*")

	// Health check endpoint
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"environment": s.config.Environment,
		})
	})

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Status routes
		statusGroup := v1.Group("/status")
		{
			statusGroup.GET("", s.getStatus)
			statusGroup.GET("/:id", s.getStatusByID)
			statusGroup.GET("/summary", s.getStatusSummary)
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
			subscriberGroup.DELETE("", s.unsubscribe) // Using email in body
		}

		// Uptime routes
		v1.GET("/uptime", s.getUptimeStats)
		v1.GET("/uptime/history/:monitor_id", s.getUptimeHistory)
		v1.GET("/analytics", s.getAnalyticsSummary)

		// Public branding endpoint
		v1.GET("/branding", s.getPublicBranding)

		// RSS/Atom feeds
		v1.GET("/incidents.rss", s.getIncidentsRSS)
		v1.GET("/incidents.atom", s.getIncidentsAtom)

		// Admin API routes
		adminAPI := v1.Group("/admin")
		{
			adminAPI.POST("/login", s.login)

			authedAPI := adminAPI.Group("/")
			authedAPI.Use(middleware.AuthMiddleware(s.config))
			{
				authedAPI.POST("/status", s.createStatus)
				authedAPI.PUT("/status/:id", s.updateStatus)
				authedAPI.DELETE("/status/:id", s.deleteStatus)
				authedAPI.POST("/incidents", s.createIncident)
				authedAPI.POST("/maintenance", s.createMaintenanceEvent)
				authedAPI.PUT("/maintenance/:id", s.updateMaintenanceEvent)
				authedAPI.DELETE("/maintenance/:id", s.deleteMaintenanceEvent)

				// Monitor routes
				authedAPI.GET("/monitors", s.getMonitors)
				authedAPI.POST("/monitors", s.createMonitor)
				authedAPI.PUT("/monitors/:id", s.updateMonitor)
				authedAPI.DELETE("/monitors/:id", s.deleteMonitor)

				// User management routes
				authedAPI.GET("/users", s.getUsers)
				authedAPI.POST("/users", s.createUser)
				authedAPI.PUT("/users/:id", s.updateUser)
				authedAPI.DELETE("/users/:id", s.deleteUser)

				// Template routes
				authedAPI.GET("/templates/incidents", s.getIncidentTemplates)
				authedAPI.POST("/templates/incidents", s.createIncidentTemplate)
				authedAPI.PUT("/templates/incidents/:id", s.updateIncidentTemplate)
				authedAPI.DELETE("/templates/incidents/:id", s.deleteIncidentTemplate)
				authedAPI.GET("/templates/maintenance", s.getMaintenanceTemplates)
				authedAPI.POST("/templates/maintenance", s.createMaintenanceTemplate)
				authedAPI.PUT("/templates/maintenance/:id", s.updateMaintenanceTemplate)
				authedAPI.DELETE("/templates/maintenance/:id", s.deleteMaintenanceTemplate)

				// Branding routes
				authedAPI.GET("/branding", s.getBranding)
				authedAPI.PUT("/branding", s.updateBranding)
				authedAPI.POST("/branding/reset", s.resetBranding)

				// Integration routes
				authedAPI.GET("/integrations", s.getIntegrations)
				authedAPI.POST("/integrations", s.createIntegration)
				authedAPI.PUT("/integrations/:id", s.updateIntegration)
				authedAPI.DELETE("/integrations/:id", s.deleteIntegration)
				authedAPI.POST("/integrations/:id/test", s.testIntegration)

				// Audit log routes
				authedAPI.GET("/audit-logs", s.getAuditLogs)

				// Private page routes
				authedAPI.GET("/private-pages", s.getPrivatePages)
				authedAPI.POST("/private-pages", s.createPrivatePage)
				authedAPI.PUT("/private-pages/:id", s.updatePrivatePage)
				authedAPI.DELETE("/private-pages/:id", s.deletePrivatePage)
				authedAPI.POST("/private-pages/:id/regenerate-key", s.regeneratePrivatePageKey)
				authedAPI.POST("/private-pages/:id/toggle", s.togglePrivatePageStatus)

				// Monitoring integration routes
				authedAPI.GET("/monitoring/status", s.getMonitoringStatus)
				authedAPI.POST("/monitoring/sync", s.syncAllMonitoringTools)
				authedAPI.POST("/monitoring/sync/:id", s.syncMonitoringTool)
			}
		}
	}

	// Serve static files
	s.router.Static("/static", "./web/static")

	// Web routes
	web := s.router.Group("")
	{
		adminWeb := web.Group("/admin")
		{
			adminWeb.GET("/login", s.showLoginPage)
			adminWeb.POST("/login", s.handleWebLogin)
			adminWeb.GET("/logout", s.handleLogout)

			authedWeb := adminWeb.Group("/")
			authedWeb.Use(middleware.AuthMiddleware(s.config))
			{
				authedWeb.GET("/dashboard", s.showDashboardPage)
				authedWeb.GET("/maintenance/new", s.showNewMaintenancePage)
				authedWeb.POST("/maintenance/new", s.handleNewMaintenance)
				authedWeb.GET("/maintenance/edit/:id", s.showEditMaintenancePage)
				authedWeb.POST("/maintenance/edit/:id", s.handleEditMaintenance)
				authedWeb.GET("/maintenance/delete/:id", s.handleDeleteMaintenance)
			}
		}

		web.GET("/", s.showIndexPage)
		
		// Private page access route
		web.GET("/private/:access_key", s.showPrivatePage)
	}
}

func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         ":" + s.config.Server.Port,
		Handler:      s.router,
		ReadTimeout:  time.Duration(s.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.config.Server.IdleTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", zap.String("port", s.config.Server.Port))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Doesn't block if no connections, otherwise will wait until the timeout
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	logger.Info("Server gracefully stopped")
	return nil
}

// Handler functions
func (s *Server) getStatus(c *gin.Context) {
	services, err := s.statusService.GetAllServices()
	if err != nil {
		logger.Error("Failed to get services", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve services"})
		return
	}

	// Determine overall status
	overallStatus := "operational"
	for _, service := range services {
		if service.Status != "operational" {
			overallStatus = "degraded" // or "outage" based on more complex logic
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   overallStatus,
		"services": services,
	})
}

func (s *Server) getStatusByID(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"name":    "Example Service",
		"status":  "operational",
		"updated": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) getStatusSummary(c *gin.Context) {
	services, err := s.statusService.GetAllServices()
	if err != nil {
		logger.Error("Failed to get services", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve services"})
		return
	}

	// Count services by status
	statusCounts := make(map[string]int)
	for _, service := range services {
		statusCounts[service.Status]++
	}

	// Determine overall status
	overallStatus := "operational"
	for _, service := range services {
		if service.Status != "operational" {
			overallStatus = "degraded"
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"overall_status": overallStatus,
		"status_counts":  statusCounts,
		"total_services": len(services),
		"last_updated":   time.Now().Format(time.RFC3339),
	})
}

func (s *Server) getIncidents(c *gin.Context) {
	incidents, err := s.incidentService.GetAllIncidents()
	if err != nil {
		logger.Error("Failed to get incidents", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve incidents"})
		return
	}
	c.JSON(http.StatusOK, incidents)
}

func (s *Server) getIncidentByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	incident, err := s.incidentService.GetIncidentByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
		return
	}

	c.JSON(http.StatusOK, incident)
}

func (s *Server) getIncidentUpdates(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	updates, err := s.incidentService.GetIncidentUpdates(uint(id))
	if err != nil {
		logger.Error("Failed to get incident updates", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve incident updates"})
		return
	}

	c.JSON(http.StatusOK, updates)
}

func (s *Server) getActiveIncidents(c *gin.Context) {
	incidents, err := s.incidentService.GetActiveIncidents()
	if err != nil {
		logger.Error("Failed to get active incidents", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve active incidents"})
		return
	}
	c.JSON(http.StatusOK, incidents)
}

func (s *Server) getResolvedIncidents(c *gin.Context) {
	incidents, err := s.incidentService.GetResolvedIncidents()
	if err != nil {
		logger.Error("Failed to get resolved incidents", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve resolved incidents"})
		return
	}
	c.JSON(http.StatusOK, incidents)
}

func (s *Server) getMaintenanceEvents(c *gin.Context) {
	events, err := s.maintenanceService.GetAllMaintenanceEvents()
	if err != nil {
		logger.Error("Failed to get maintenance events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve maintenance events"})
		return
	}
	c.JSON(http.StatusOK, events)
}

func (s *Server) getMaintenanceEventByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid maintenance event ID"})
		return
	}

	event, err := s.maintenanceService.GetMaintenanceEventByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Maintenance event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

func (s *Server) getUpcomingMaintenance(c *gin.Context) {
	events, err := s.maintenanceService.GetUpcomingMaintenance()
	if err != nil {
		logger.Error("Failed to get upcoming maintenance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve upcoming maintenance"})
		return
	}
	c.JSON(http.StatusOK, events)
}

func (s *Server) getActiveMaintenance(c *gin.Context) {
	events, err := s.maintenanceService.GetActiveMaintenance()
	if err != nil {
		logger.Error("Failed to get active maintenance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve active maintenance"})
		return
	}
	c.JSON(http.StatusOK, events)
}

func (s *Server) getPublicBranding(c *gin.Context) {
	branding, err := s.brandingService.GetBranding()
	if err != nil {
		logger.Error("Failed to get branding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve branding"})
		return
	}
	c.JSON(http.StatusOK, branding)
}

func (s *Server) getIncidentsRSS(c *gin.Context) {
	incidents, err := s.incidentService.GetAllIncidents()
	if err != nil {
		logger.Error("Failed to get incidents for RSS", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve incidents"})
		return
	}

	// Generate RSS feed
	rss := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
<title>Status Page Incidents</title>
<description>Latest incidents from our status page</description>
<link>http://localhost:8080</link>
<lastBuildDate>` + time.Now().Format(time.RFC1123Z) + `</lastBuildDate>`

	for _, incident := range incidents {
		rss += `
<item>
<title>` + incident.Title + `</title>
<description>` + incident.Description + `</description>
<pubDate>` + incident.CreatedAt.Format(time.RFC1123Z) + `</pubDate>
<guid>` + fmt.Sprintf("%d", incident.ID) + `</guid>
</item>`
	}

	rss += `
</channel>
</rss>`

	c.Header("Content-Type", "application/rss+xml")
	c.String(http.StatusOK, rss)
}

func (s *Server) getIncidentsAtom(c *gin.Context) {
	incidents, err := s.incidentService.GetAllIncidents()
	if err != nil {
		logger.Error("Failed to get incidents for Atom", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve incidents"})
		return
	}

	// Generate Atom feed
	atom := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
<title>Status Page Incidents</title>
<subtitle>Latest incidents from our status page</subtitle>
<link href="http://localhost:8080"/>
<updated>` + time.Now().Format(time.RFC3339) + `</updated>
<id>http://localhost:8080/api/v1/incidents.atom</id>`

	for _, incident := range incidents {
		atom += `
<entry>
<title>` + incident.Title + `</title>
<summary>` + incident.Description + `</summary>
<published>` + incident.CreatedAt.Format(time.RFC3339) + `</published>
<updated>` + incident.UpdatedAt.Format(time.RFC3339) + `</updated>
<id>` + fmt.Sprintf("incident-%d", incident.ID) + `</id>
</entry>`
	}

	atom += `
</feed>`

	c.Header("Content-Type", "application/atom+xml")
	c.String(http.StatusOK, atom)
}

func (s *Server) createStatus(c *gin.Context) {
	// TODO: Implement status creation
	c.JSON(http.StatusCreated, gin.H{
		"message": "Status created successfully",
	})
}

func (s *Server) updateStatus(c *gin.Context) {
	// TODO: Implement status update
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Status updated successfully",
		"id":      id,
	})
}

func (s *Server) deleteStatus(c *gin.Context) {
	// TODO: Implement status deletion
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Status deleted successfully",
		"id":      id,
	})
}


func (s *Server) login(c *gin.Context) {
	type loginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	token, err := s.authService.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (s *Server) createIncident(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Status      string `json:"status"`
		ServiceIDs  []uint `json:"service_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	incident := &models.Incident{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	}

	if err := s.incidentService.CreateIncident(incident, req.ServiceIDs); err != nil {
		logger.Error("Failed to create incident", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create incident"})
		return
	}

	// Send notification
	go func() {
		if err := s.notificationService.NotifyNewIncident(incident); err != nil {
			logger.Error("Failed to send new incident notification", zap.Error(err))
		}
	}()

	c.JSON(http.StatusCreated, incident)
}

// Maintenance handlers

func (s *Server) createMaintenanceEvent(c *gin.Context) {
	var req struct {
		Title       string    `json:"title" binding:"required"`
		Description string    `json:"description"`
		Status      string    `json:"status"`
		StartAt     time.Time `json:"start_at" binding:"required"`
		EndAt       time.Time `json:"end_at" binding:"required"`
		ServiceIDs  []uint    `json:"service_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event := &models.Maintenance{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		StartAt:     req.StartAt,
		EndAt:       req.EndAt,
	}

	if err := s.maintenanceService.CreateMaintenanceEvent(event, req.ServiceIDs); err != nil {
		logger.Error("Failed to create maintenance event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create maintenance event"})
		return
	}

	// Send notification
	go func() {
		if err := s.notificationService.NotifyNewMaintenance(event); err != nil {
			logger.Error("Failed to send new maintenance notification", zap.Error(err))
		}
	}()

	c.JSON(http.StatusCreated, event)
}

func (s *Server) updateMaintenanceEvent(c *gin.Context) {
	var event models.Maintenance
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.maintenanceService.UpdateMaintenanceEvent(&event); err != nil {
		logger.Error("Failed to update maintenance event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update maintenance event"})
		return
	}

	c.JSON(http.StatusOK, event)
}

func (s *Server) deleteMaintenanceEvent(c *gin.Context) {
	id := c.Param("id")
	var maintenanceID uint
	if _, err := fmt.Sscanf(id, "%d", &maintenanceID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid maintenance ID"})
		return
	}

	if err := s.maintenanceService.DeleteMaintenanceEvent(maintenanceID); err != nil {
		logger.Error("Failed to delete maintenance event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete maintenance event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Maintenance event deleted successfully"})
}

// Subscriber handlers
func (s *Server) subscribe(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Services []uint `json:"services"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	subscriber, err := s.subscriberService.Subscribe(req.Email, req.Services)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "One or more service IDs are invalid"})
			return
		}
		logger.Error("Failed to subscribe email", zap.Error(err), zap.String("email", req.Email))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to subscribe. Please try again later."})
		return
	}

	c.JSON(http.StatusCreated, subscriber)
}

func (s *Server) unsubscribe(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address"})
		return
	}

	if err := s.subscriberService.Unsubscribe(req.Email); err != nil {
		logger.Error("Failed to unsubscribe email", zap.Error(err), zap.String("email", req.Email))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unsubscribe. Please try again later."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully unsubscribed"})
}

// Web Handlers
func (s *Server) showLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func (s *Server) handleWebLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	token, err := s.authService.AuthenticateUser(username, password)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Invalid credentials"})
		return
	}

	// Set cookie (secure should be true in production)
	c.SetCookie("auth_token", token, 3600, "/", "localhost", false, true)

	c.Redirect(http.StatusFound, "/admin/dashboard")
}

func (s *Server) handleLogout(c *gin.Context) {
	c.SetCookie("auth_token", "", -1, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/admin/login")
}

func (s *Server) getUptimeStats(c *gin.Context) {
	stats, err := s.uptimeService.GetUptimeStatsForMonitors()
	if err != nil {
		logger.Error("Failed to get uptime stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve uptime statistics"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (s *Server) getUptimeHistory(c *gin.Context) {
	monitorIDStr := c.Param("monitor_id")
	monitorID, err := strconv.ParseUint(monitorIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID"})
		return
	}

	period := c.DefaultQuery("period", "24h")
	if period != "24h" && period != "7d" && period != "30d" && period != "90d" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period. Must be one of: 24h, 7d, 30d, 90d"})
		return
	}

	history, err := s.uptimeService.GetUptimeHistory(uint(monitorID), period)
	if err != nil {
		logger.Error("Failed to get uptime history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve uptime history"})
		return
	}
	c.JSON(http.StatusOK, history)
}

func (s *Server) getAnalyticsSummary(c *gin.Context) {
	summary, err := s.uptimeService.GetAnalyticsSummary()
	if err != nil {
		logger.Error("Failed to get analytics summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve analytics summary"})
		return
	}
	c.JSON(http.StatusOK, summary)
}

// Monitor handlers
func (s *Server) getMonitors(c *gin.Context) {
	monitors, err := s.monitorService.GetAllMonitors()
	if err != nil {
		logger.Error("Failed to get monitors", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve monitors"})
		return
	}
	c.JSON(http.StatusOK, monitors)
}

func (s *Server) createMonitor(c *gin.Context) {
	var req struct {
		Name           string `json:"name" binding:"required"`
		URL            string `json:"url" binding:"required"`
		Type           string `json:"type"`
		Interval       int    `json:"interval"`
		ExpectedStatus int    `json:"expected_status"`
		Timeout        int    `json:"timeout"`
		ServiceID      uint   `json:"service_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	monitor := &models.Monitor{
		Name:           req.Name,
		ServiceID:      req.ServiceID,
		Type:           req.Type,
		URL:            req.URL,
		Interval:       uint(req.Interval),
		ExpectedStatus: req.ExpectedStatus,
		Timeout:        req.Timeout,
	}

	if err := s.monitorService.CreateMonitor(monitor); err != nil {
		logger.Error("Failed to create monitor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create monitor", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, monitor)
}

func (s *Server) updateMonitor(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	monitor, err := s.monitorService.GetMonitorByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Monitor not found"})
		return
	}

	if err := c.ShouldBindJSON(monitor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.monitorService.UpdateMonitor(monitor); err != nil {
		logger.Error("Failed to update monitor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update monitor"})
		return
	}

	c.JSON(http.StatusOK, monitor)
}

func (s *Server) deleteMonitor(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	if err := s.monitorService.DeleteMonitor(uint(id)); err != nil {
		logger.Error("Failed to delete monitor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete monitor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Monitor deleted successfully"})
}

func (s *Server) showDashboardPage(c *gin.Context) {
	maintenanceEvents, err := s.maintenanceService.GetAllMaintenanceEvents()
	if err != nil {
		logger.Error("Failed to get maintenance events for dashboard", zap.Error(err))
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Could not load maintenance events"})
		return
	}

	subscribers, err := s.subscriberService.GetAllSubscribers()
	if err != nil {
		logger.Error("Failed to get subscribers for dashboard", zap.Error(err))
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Could not load subscribers"})
		return
	}

	c.HTML(http.StatusOK, "admin_dashboard.html", gin.H{
		"MaintenanceEvents": maintenanceEvents,
		"Subscribers":       subscribers,
	})
}

func (s *Server) showNewMaintenancePage(c *gin.Context) {
	c.HTML(http.StatusOK, "maintenance_form.html", gin.H{
		"Title":  "New Maintenance Event",
		"Action": "/admin/maintenance/new",
	})
}

func (s *Server) handleNewMaintenance(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	startAtStr := c.PostForm("start_at")
	endAtStr := c.PostForm("end_at")

	startAt, err := time.Parse("2006-01-02T15:04", startAtStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid Start At date format"})
		return
	}

	endAt, err := time.Parse("2006-01-02T15:04", endAtStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid End At date format"})
		return
	}

	maintenance := &models.Maintenance{
		Title:       title,
		Description: description,
		StartAt:     startAt,
		EndAt:       endAt,
	}

	if err := s.maintenanceService.CreateMaintenanceEvent(maintenance, nil); err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to create maintenance event"})
		return
	}

	c.Redirect(http.StatusFound, "/admin/dashboard")
}

func (s *Server) showEditMaintenancePage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid maintenance event ID"})
		return
	}

	maintenance, err := s.maintenanceService.GetMaintenanceEventByID(uint(id))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Maintenance event not found"})
		return
	}

	c.HTML(http.StatusOK, "maintenance_form.html", gin.H{
		"Title":       "Edit Maintenance Event",
		"Action":      "/admin/maintenance/edit/" + idStr,
		"Maintenance": maintenance,
	})
}

func (s *Server) handleEditMaintenance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid maintenance event ID"})
		return
	}

	maintenance, err := s.maintenanceService.GetMaintenanceEventByID(uint(id))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Maintenance event not found"})
		return
	}

	maintenance.Title = c.PostForm("title")
	maintenance.Description = c.PostForm("description")
	startAtStr := c.PostForm("start_at")
	endAtStr := c.PostForm("end_at")

	startAt, err := time.Parse("2006-01-02T15:04", startAtStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid Start At date format"})
		return
	}

	endAt, err := time.Parse("2006-01-02T15:04", endAtStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid End At date format"})
		return
	}
	maintenance.StartAt = startAt
	maintenance.EndAt = endAt

	if err := s.maintenanceService.UpdateMaintenanceEvent(maintenance); err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to update maintenance event"})
		return
	}

	c.Redirect(http.StatusFound, "/admin/dashboard")
}

func (s *Server) handleDeleteMaintenance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid maintenance event ID"})
		return
	}

	if err := s.maintenanceService.DeleteMaintenanceEvent(uint(id)); err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to delete maintenance event"})
		return
	}
	c.Redirect(http.StatusFound, "/admin/dashboard")
}

func (s *Server) showIndexPage(c *gin.Context) {
	services, err := s.statusService.GetAllServices()
	if err != nil {
		logger.Error("Failed to get services for index page", zap.Error(err))
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Could not load services"})
		return
	}

	incidents, err := s.incidentService.GetAllIncidents()
	if err != nil {
		logger.Error("Failed to get incidents for index page", zap.Error(err))
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Could not load incidents"})
		return
	}

	maintenance, err := s.maintenanceService.GetAllMaintenanceEvents()
	if err != nil {
		logger.Error("Failed to get maintenance for index page", zap.Error(err))
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Could not load maintenance events"})
		return
	}

	overallStatus := "operational"
	for _, service := range services {
		if service.Status != "operational" {
			overallStatus = "degraded"
			break
		}
	}

	c.HTML(http.StatusOK, "public_status.html", gin.H{
		"Services":      services,
		"Incidents":     incidents,
		"Maintenance":   maintenance,
		"OverallStatus": overallStatus,
	})
}

// User Management Handlers

func (s *Server) getUsers(c *gin.Context) {
	users, err := s.rbacService.GetAllUsers()
	if err != nil {
		logger.Error("Failed to get users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (s *Server) createUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.rbacService.CreateUser(req.Username, req.Password, req.Role)
	if err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (s *Server) updateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Role string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.rbacService.UpdateUserRole(uint(id), req.Role); err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func (s *Server) deleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := s.rbacService.DeleteUser(uint(id)); err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// Template Handlers

func (s *Server) getIncidentTemplates(c *gin.Context) {
	templates, err := s.templateService.GetAllIncidentTemplates()
	if err != nil {
		logger.Error("Failed to get incident templates", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve incident templates"})
		return
	}
	c.JSON(http.StatusOK, templates)
}

func (s *Server) createIncidentTemplate(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Impact      string `json:"impact"`
		ServiceIDs  []uint `json:"service_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template := &models.IncidentTemplate{
		Name:        req.Name,
		Title:       req.Title,
		Description: req.Description,
		Impact:      req.Impact,
	}

	if err := s.templateService.CreateIncidentTemplate(template, req.ServiceIDs); err != nil {
		logger.Error("Failed to create incident template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create incident template"})
		return
	}

	c.JSON(http.StatusCreated, template)
}

func (s *Server) updateIncidentTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	template, err := s.templateService.GetIncidentTemplateByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Impact      string `json:"impact"`
		ServiceIDs  []uint `json:"service_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template.Name = req.Name
	template.Title = req.Title
	template.Description = req.Description
	template.Impact = req.Impact

	if err := s.templateService.UpdateIncidentTemplate(template, req.ServiceIDs); err != nil {
		logger.Error("Failed to update incident template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update incident template"})
		return
	}

	c.JSON(http.StatusOK, template)
}

func (s *Server) deleteIncidentTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	if err := s.templateService.DeleteIncidentTemplate(uint(id)); err != nil {
		logger.Error("Failed to delete incident template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete incident template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully"})
}

func (s *Server) getMaintenanceTemplates(c *gin.Context) {
	templates, err := s.templateService.GetAllMaintenanceTemplates()
	if err != nil {
		logger.Error("Failed to get maintenance templates", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve maintenance templates"})
		return
	}
	c.JSON(http.StatusOK, templates)
}

func (s *Server) createMaintenanceTemplate(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Duration    uint   `json:"duration"`
		ServiceIDs  []uint `json:"service_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template := &models.MaintenanceTemplate{
		Name:        req.Name,
		Title:       req.Title,
		Description: req.Description,
		Duration:    req.Duration,
	}

	if err := s.templateService.CreateMaintenanceTemplate(template, req.ServiceIDs); err != nil {
		logger.Error("Failed to create maintenance template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create maintenance template"})
		return
	}

	c.JSON(http.StatusCreated, template)
}

func (s *Server) updateMaintenanceTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	template, err := s.templateService.GetMaintenanceTemplateByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Duration    uint   `json:"duration"`
		ServiceIDs  []uint `json:"service_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template.Name = req.Name
	template.Title = req.Title
	template.Description = req.Description
	template.Duration = req.Duration

	if err := s.templateService.UpdateMaintenanceTemplate(template, req.ServiceIDs); err != nil {
		logger.Error("Failed to update maintenance template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update maintenance template"})
		return
	}

	c.JSON(http.StatusOK, template)
}

func (s *Server) deleteMaintenanceTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	if err := s.templateService.DeleteMaintenanceTemplate(uint(id)); err != nil {
		logger.Error("Failed to delete maintenance template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete maintenance template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully"})
}

// Branding Handlers

func (s *Server) getBranding(c *gin.Context) {
	branding, err := s.brandingService.GetBranding()
	if err != nil {
		logger.Error("Failed to get branding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve branding"})
		return
	}
	c.JSON(http.StatusOK, branding)
}

func (s *Server) updateBranding(c *gin.Context) {
	var branding models.Branding
	if err := c.ShouldBindJSON(&branding); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.brandingService.ValidateBranding(&branding); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branding configuration"})
		return
	}

	if err := s.brandingService.UpdateBranding(&branding); err != nil {
		logger.Error("Failed to update branding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update branding"})
		return
	}

	c.JSON(http.StatusOK, branding)
}

func (s *Server) resetBranding(c *gin.Context) {
	if err := s.brandingService.ResetBranding(); err != nil {
		logger.Error("Failed to reset branding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset branding"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Branding reset successfully"})
}

// Integration Handlers

func (s *Server) getIntegrations(c *gin.Context) {
	integrations, err := s.integrationService.GetAllIntegrations()
	if err != nil {
		logger.Error("Failed to get integrations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve integrations"})
		return
	}
	c.JSON(http.StatusOK, integrations)
}

func (s *Server) createIntegration(c *gin.Context) {
	var integration models.Integration
	if err := c.ShouldBindJSON(&integration); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.integrationService.CreateIntegration(&integration); err != nil {
		logger.Error("Failed to create integration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create integration"})
		return
	}

	c.JSON(http.StatusCreated, integration)
}

func (s *Server) updateIntegration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	integration, err := s.integrationService.GetIntegrationByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	if err := c.ShouldBindJSON(integration); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.integrationService.UpdateIntegration(integration); err != nil {
		logger.Error("Failed to update integration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update integration"})
		return
	}

	c.JSON(http.StatusOK, integration)
}

func (s *Server) deleteIntegration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	if err := s.integrationService.DeleteIntegration(uint(id)); err != nil {
		logger.Error("Failed to delete integration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete integration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Integration deleted successfully"})
}

func (s *Server) testIntegration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	integration, err := s.integrationService.GetIntegrationByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	if err := s.integrationService.TestIntegration(integration); err != nil {
		logger.Error("Failed to test integration", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Integration test successful"})
}

// Audit Log Handlers

func (s *Server) getAuditLogs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	userIDStr := c.Query("user_id")
	action := c.Query("action")
	resource := c.Query("resource")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	var logs []models.AuditLog
	var err error

	if userIDStr != "" {
		userID, parseErr := strconv.ParseUint(userIDStr, 10, 32)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		userIDUint := uint(userID)
		logs, err = s.auditService.GetAuditLogs(&userIDUint, action, resource, limit, offset)
	} else {
		logs, err = s.auditService.GetAuditLogs(nil, action, resource, limit, offset)
	}

	if err != nil {
		logger.Error("Failed to get audit logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit logs"})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// Private Page Handlers

func (s *Server) getPrivatePages(c *gin.Context) {
	pages, err := s.privatePageService.GetAllPrivatePages()
	if err != nil {
		logger.Error("Failed to get private pages", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve private pages"})
		return
	}
	c.JSON(http.StatusOK, pages)
}

func (s *Server) createPrivatePage(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		ServiceIDs  []uint `json:"service_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	page, err := s.privatePageService.CreatePrivatePage(req.Name, req.Description, req.ServiceIDs)
	if err != nil {
		logger.Error("Failed to create private page", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create private page"})
		return
	}

	c.JSON(http.StatusCreated, page)
}

func (s *Server) updatePrivatePage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid private page ID"})
		return
	}

	page, err := s.privatePageService.GetPrivatePageByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Private page not found"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		ServiceIDs  []uint `json:"service_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	page.Name = req.Name
	page.Description = req.Description

	if err := s.privatePageService.UpdatePrivatePage(page, req.ServiceIDs); err != nil {
		logger.Error("Failed to update private page", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update private page"})
		return
	}

	c.JSON(http.StatusOK, page)
}

func (s *Server) deletePrivatePage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid private page ID"})
		return
	}

	if err := s.privatePageService.DeletePrivatePage(uint(id)); err != nil {
		logger.Error("Failed to delete private page", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete private page"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Private page deleted successfully"})
}

func (s *Server) regeneratePrivatePageKey(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid private page ID"})
		return
	}

	accessKey, err := s.privatePageService.RegenerateAccessKey(uint(id))
	if err != nil {
		logger.Error("Failed to regenerate access key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to regenerate access key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_key": accessKey})
}

func (s *Server) togglePrivatePageStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid private page ID"})
		return
	}

	if err := s.privatePageService.TogglePrivatePageStatus(uint(id)); err != nil {
		logger.Error("Failed to toggle private page status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle private page status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Private page status toggled successfully"})
}

func (s *Server) showPrivatePage(c *gin.Context) {
	accessKey := c.Param("access_key")
	
	// Validate access key
	if !s.privatePageService.ValidateAccessKey(accessKey) {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Private page not found or access denied"})
		return
	}

	// Get private page data
	data, err := s.privatePageService.GetPrivatePageData(accessKey)
	if err != nil {
		logger.Error("Failed to get private page data", zap.Error(err))
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Could not load private page data"})
		return
	}

	c.HTML(http.StatusOK, "private_page.html", data)
}

// Monitoring Integration Handlers

func (s *Server) getMonitoringStatus(c *gin.Context) {
	statuses, err := s.monitoringIntegrationService.GetMonitoringStatus()
	if err != nil {
		logger.Error("Failed to get monitoring status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve monitoring status"})
		return
	}
	c.JSON(http.StatusOK, statuses)
}

func (s *Server) syncAllMonitoringTools(c *gin.Context) {
	if err := s.monitoringIntegrationService.SyncAllMonitoringTools(); err != nil {
		logger.Error("Failed to sync all monitoring tools", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync monitoring tools"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "All monitoring tools synced successfully"})
}

func (s *Server) syncMonitoringTool(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	integration, err := s.integrationService.GetIntegrationByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	if err := s.monitoringIntegrationService.SyncWithMonitoringTool(integration); err != nil {
		logger.Error("Failed to sync monitoring tool", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync monitoring tool"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Monitoring tool synced successfully"})
}
