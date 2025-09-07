package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/enterprise-status/statuspage/internal/api/middleware"
	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TenantHandlers handles tenant-specific API endpoints
type TenantHandlers struct {
	db                  *gorm.DB
	statusService       *services.StatusService
	incidentService     *services.IncidentService
	maintenanceService  *services.MaintenanceService
	subscriberService   *services.SubscriberService
	monitorService      *services.MonitorService
	brandingService     *services.BrandingService
	subscriptionService *services.SubscriptionService
	paymentService      *services.PaymentService
}

func NewTenantHandlers(
	db *gorm.DB,
	statusService *services.StatusService,
	incidentService *services.IncidentService,
	maintenanceService *services.MaintenanceService,
	subscriberService *services.SubscriberService,
	monitorService *services.MonitorService,
	brandingService *services.BrandingService,
	subscriptionService *services.SubscriptionService,
	paymentService *services.PaymentService,
) *TenantHandlers {
	return &TenantHandlers{
		db:                  db,
		statusService:       statusService,
		incidentService:     incidentService,
		maintenanceService:  maintenanceService,
		subscriberService:   subscriberService,
		monitorService:      monitorService,
		brandingService:     brandingService,
		subscriptionService: subscriptionService,
		paymentService:      paymentService,
	}
}

// GetTenantInfo returns basic tenant information
func (h *TenantHandlers) GetTenantInfo(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tenant": tenant})
}

// GetTenantOverview returns overview statistics for the tenant
func (h *TenantHandlers) GetTenantOverview(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get services
	services, err := h.statusService.GetServicesByTenantID(tenant.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch services"})
		return
	}

	// Get active incidents
	incidents, err := h.incidentService.GetActiveIncidentsByTenantID(tenant.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch incidents"})
		return
	}

	// Get subscribers
	subscribers, err := h.subscriberService.GetSubscribersByTenantID(tenant.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscribers"})
		return
	}

	// Calculate stats
	stats := gin.H{
		"total_services":       len(services),
		"active_incidents":     len(incidents),
		"total_subscribers":    len(subscribers),
		"uptime_percentage":    100.0, // TODO: Calculate actual uptime
		"operational_services": 0,
		"degraded_services":    0,
		"outage_services":      0,
	}

	// Count services by status
	for _, service := range services {
		switch service.Status {
		case "operational":
			stats["operational_services"] = stats["operational_services"].(int) + 1
		case "degraded":
			stats["degraded_services"] = stats["degraded_services"].(int) + 1
		case "outage":
			stats["outage_services"] = stats["outage_services"].(int) + 1
		}
	}

	// Get incidents timeline (last 7 days)
	incidentsTimeline := gin.H{
		"labels": []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
		"data":   []int{0, 0, 0, 0, 0, 0, 0}, // TODO: Calculate actual timeline
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":              stats,
		"incidents_timeline": incidentsTimeline,
	})
}

// GetTenantServices returns all services for the tenant
func (h *TenantHandlers) GetTenantServices(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	services, err := h.statusService.GetServicesByTenantID(tenant.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch services"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"services": services})
}

// CreateTenantService creates a new service for the tenant
func (h *TenantHandlers) CreateTenantService(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Status      string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service := &models.Service{
		TenantID:    tenant.ID,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
	}

	if _, err := h.statusService.CreateService(service); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"service": service})
}

// GetTenantIncidents returns all incidents for the tenant
func (h *TenantHandlers) GetTenantIncidents(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	incidents, err := h.incidentService.GetIncidentsByTenantID(tenant.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch incidents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"incidents": incidents})
}

// CreateTenantIncident creates a new incident for the tenant
func (h *TenantHandlers) CreateTenantIncident(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description" binding:"required"`
		Status      string `json:"status" binding:"required"`
		Severity    string `json:"severity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	incident := &models.Incident{
		TenantID:    tenant.ID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Severity:    req.Severity,
	}

	if _, err := h.incidentService.CreateIncident(incident); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create incident"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"incident": incident})
}

// GetTenantMaintenance returns all maintenance windows for the tenant
func (h *TenantHandlers) GetTenantMaintenance(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	maintenance, err := h.maintenanceService.GetMaintenanceByTenantID(tenant.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch maintenance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"maintenance": maintenance})
}

// GetTenantSubscribers returns all subscribers for the tenant
func (h *TenantHandlers) GetTenantSubscribers(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	subscribers, err := h.subscriberService.GetSubscribersByTenantID(tenant.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscribers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscribers": subscribers})
}

// GetTenantMonitors returns all monitors for the tenant
func (h *TenantHandlers) GetTenantMonitors(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	monitors, err := h.monitorService.GetMonitorsByTenantID(tenant.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch monitors"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"monitors": monitors})
}

// GetTenantBranding returns branding information for the tenant
func (h *TenantHandlers) GetTenantBranding(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	branding, err := h.brandingService.GetBrandingByTenantID(tenant.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch branding"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"branding": branding})
}

// UpdateTenantBranding updates branding information for the tenant
func (h *TenantHandlers) UpdateTenantBranding(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	var req struct {
		CompanyName    string `json:"company_name"`
		PrimaryColor   string `json:"primary_color"`
		SecondaryColor string `json:"secondary_color"`
		LogoURL        string `json:"logo_url"`
		FooterText     string `json:"footer_text"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branding := &models.Branding{
		TenantID:       tenant.ID,
		CompanyName:    req.CompanyName,
		PrimaryColor:   req.PrimaryColor,
		SecondaryColor: req.SecondaryColor,
		LogoURL:        req.LogoURL,
		FooterText:     req.FooterText,
	}

	if err := h.brandingService.UpdateBranding(branding); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update branding"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"branding": branding})
}

// GetTenantSettings returns settings for the tenant
func (h *TenantHandlers) GetTenantSettings(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	settings := gin.H{
		"contact_email": tenant.ContactEmail,
		"timezone":      "UTC", // TODO: Add timezone to tenant model
		"custom_domain": tenant.Domain,
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

// UpdateTenantSettings updates settings for the tenant
func (h *TenantHandlers) UpdateTenantSettings(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	var req struct {
		ContactEmail string `json:"contact_email"`
		Timezone     string `json:"timezone"`
		CustomDomain string `json:"custom_domain"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update tenant
	tenant.ContactEmail = req.ContactEmail
	tenant.Domain = req.CustomDomain

	// TODO: Update tenant in database
	// if err := h.saasService.UpdateTenant(tenant); err != nil {
	//     c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
	//     return
	// }

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}

// GetTenantBilling returns billing information for the tenant
func (h *TenantHandlers) GetTenantBilling(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get current subscription
	subscription, err := h.subscriptionService.GetSubscriptionByTenantID(tenant.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscription"})
		return
	}

	// Get available plans
	plans, err := h.subscriptionService.GetAllPlans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plans"})
		return
	}

	// Get billing history (invoices and payments)
	var invoices []models.BillingInvoice
	if err := h.db.Where("tenant_id = ?", tenant.ID).Order("created_at DESC").Limit(10).Find(&invoices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch billing history"})
		return
	}

	// Get payment history
	var payments []models.BillingPayment
	if err := h.db.Where("tenant_id = ?", tenant.ID).Order("created_at DESC").Limit(10).Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payment history"})
		return
	}

	// Get billing metrics
	var totalInvoices int64
	var totalPaid float64
	var totalOutstanding float64

	h.db.Model(&models.BillingInvoice{}).Where("tenant_id = ?", tenant.ID).Count(&totalInvoices)
	h.db.Model(&models.BillingInvoice{}).Where("tenant_id = ? AND status = ?", tenant.ID, "paid").Select("COALESCE(SUM(total_amount), 0)").Scan(&totalPaid)
	h.db.Model(&models.BillingInvoice{}).Where("tenant_id = ? AND status IN ?", tenant.ID, []string{"pending", "overdue"}).Select("COALESCE(SUM(total_amount), 0)").Scan(&totalOutstanding)

	response := gin.H{
		"current_plan":    nil,
		"available_plans": plans,
		"billing_history": invoices,
		"payment_history": payments,
		"billing_metrics": gin.H{
			"total_invoices":    totalInvoices,
			"total_paid":        totalPaid,
			"total_outstanding": totalOutstanding,
		},
	}

	if subscription != nil {
		response["current_plan"] = subscription.Plan
	}

	c.JSON(http.StatusOK, response)
}

// GetTenantInvoices returns invoice history for the tenant
func (h *TenantHandlers) GetTenantInvoices(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get pagination parameters
	limit := c.DefaultQuery("limit", "20")
	offset := c.DefaultQuery("offset", "0")

	var invoices []models.BillingInvoice
	query := h.db.Where("tenant_id = ?", tenant.ID).Order("created_at DESC")

	if limit != "" {
		query = query.Limit(parseInt(limit))
	}
	if offset != "" {
		query = query.Offset(parseInt(offset))
	}

	if err := query.Find(&invoices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch invoices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invoices": invoices})
}

// GetTenantPayments returns payment history for the tenant
func (h *TenantHandlers) GetTenantPayments(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get pagination parameters
	limit := c.DefaultQuery("limit", "20")
	offset := c.DefaultQuery("offset", "0")

	var payments []models.BillingPayment
	query := h.db.Where("tenant_id = ?", tenant.ID).Order("created_at DESC")

	if limit != "" {
		query = query.Limit(parseInt(limit))
	}
	if offset != "" {
		query = query.Offset(parseInt(offset))
	}

	if err := query.Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payments": payments})
}

// GetTenantBillingMetrics returns billing metrics for the tenant
func (h *TenantHandlers) GetTenantBillingMetrics(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get period parameter (default to last 30 days)
	period := c.DefaultQuery("period", "30d")

	// Calculate metrics based on period
	var metrics gin.H

	switch period {
	case "7d":
		metrics = h.calculateBillingMetrics(tenant.ID, 7)
	case "30d":
		metrics = h.calculateBillingMetrics(tenant.ID, 30)
	case "90d":
		metrics = h.calculateBillingMetrics(tenant.ID, 90)
	case "1y":
		metrics = h.calculateBillingMetrics(tenant.ID, 365)
	default:
		metrics = h.calculateBillingMetrics(tenant.ID, 30)
	}

	c.JSON(http.StatusOK, metrics)
}

// calculateBillingMetrics calculates billing metrics for a given period
func (h *TenantHandlers) calculateBillingMetrics(tenantID uint, days int) gin.H {
	var totalInvoices int64
	var totalPaid float64
	var totalOutstanding float64
	var totalRefunded float64

	// Calculate date range
	startDate := time.Now().AddDate(0, 0, -days)

	// Total invoices in period
	h.db.Model(&models.BillingInvoice{}).Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).Count(&totalInvoices)

	// Total paid in period
	h.db.Model(&models.BillingInvoice{}).Where("tenant_id = ? AND status = ? AND created_at >= ?", tenantID, "paid", startDate).Select("COALESCE(SUM(total_amount), 0)").Scan(&totalPaid)

	// Total outstanding
	h.db.Model(&models.BillingInvoice{}).Where("tenant_id = ? AND status IN ?", tenantID, []string{"pending", "overdue"}).Select("COALESCE(SUM(total_amount), 0)").Scan(&totalOutstanding)

	// Total refunded in period
	h.db.Model(&models.BillingRefund{}).Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).Select("COALESCE(SUM(amount), 0)").Scan(&totalRefunded)

	return gin.H{
		"period":            fmt.Sprintf("%dd", days),
		"total_invoices":    totalInvoices,
		"total_paid":        totalPaid,
		"total_outstanding": totalOutstanding,
		"total_refunded":    totalRefunded,
		"net_revenue":       totalPaid - totalRefunded,
	}
}

// parseInt safely parses a string to int
func parseInt(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return 0
}

// CreateCheckoutSession creates a Stripe checkout session for plan upgrade
func (h *TenantHandlers) CreateCheckoutSession(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	var req struct {
		PlanSlug   string `json:"plan_slug" binding:"required"`
		SuccessURL string `json:"success_url" binding:"required"`
		CancelURL  string `json:"cancel_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.paymentService.CreateCheckoutSession(
		tenant.ID,
		req.PlanSlug,
		req.SuccessURL,
		req.CancelURL,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"checkout_url": session.URL,
		"session_id":   session.ID,
	})
}

// GetSubscriptionStatus returns the current subscription status
func (h *TenantHandlers) GetSubscriptionStatus(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	subscription, err := h.subscriptionService.GetSubscriptionByTenantID(tenant.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscription"})
		return
	}

	if subscription == nil {
		c.JSON(http.StatusOK, gin.H{
			"has_subscription": false,
			"plan":             "free",
			"status":           "active",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"has_subscription":     true,
		"plan":                 subscription.Plan.Slug,
		"status":               subscription.Status,
		"current_period_end":   subscription.CurrentPeriodEnd,
		"cancel_at_period_end": subscription.CancelAtPeriodEnd,
	})
}
