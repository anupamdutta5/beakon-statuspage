// Package handlers provides HTTP handlers for the Tenant Service.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/enterprise-status/statuspage-tenant-service/internal/models"
	"github.com/enterprise-status/statuspage-tenant-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TenantHandler handles tenant-related HTTP requests.
type TenantHandler struct {
	tenantService *services.TenantService
	logger        *zap.Logger
}

// NewTenantHandler creates a new tenant handler.
func NewTenantHandler(tenantService *services.TenantService, logger *zap.Logger) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
		logger:        logger,
	}
}

// Health returns the health status of the Tenant Service.
func (h *TenantHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "tenant-service",
		"version":   "1.0.0",
		"timestamp": "2024-01-01T00:00:00Z",
	})
}

// GetTenants handles getting a list of tenants.
func (h *TenantHandler) GetTenants(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	tenants, total, err := h.tenantService.GetTenants(limit, offset)
	if err != nil {
		h.logger.Error("Failed to get tenants", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tenants"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenants": tenants,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetTenant handles getting a specific tenant.
func (h *TenantHandler) GetTenant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	tenant, err := h.tenantService.GetTenant(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tenant": tenant})
}

// GetTenantBySlug handles getting a tenant by slug (public endpoint).
func (h *TenantHandler) GetTenantBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug is required"})
		return
	}

	tenant, err := h.tenantService.GetTenantBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tenant": tenant})
}

// CreateTenant handles creating a new tenant.
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req struct {
		Name         string `json:"name" binding:"required"`
		Slug         string `json:"slug"`
		Domain       string `json:"domain"`
		Subdomain    string `json:"subdomain"`
		ContactEmail string `json:"contact_email" binding:"required,email"`
		BillingEmail string `json:"billing_email"`
		Plan         string `json:"plan"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create tenant request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Set default plan
	if req.Plan == "" {
		req.Plan = "free"
	}

	tenant := &models.Tenant{
		Name:         req.Name,
		Slug:         req.Slug,
		Domain:       req.Domain,
		Subdomain:    req.Subdomain,
		ContactEmail: req.ContactEmail,
		BillingEmail: req.BillingEmail,
		Plan:         req.Plan,
		Status:       "active",
		IsActive:     true,
	}

	if err := h.tenantService.CreateTenant(tenant); err != nil {
		h.logger.Error("Failed to create tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
		return
	}

	// Log tenant activity
	activity := &models.TenantActivity{
		TenantID:   tenant.ID,
		Action:     "tenant_created",
		Resource:   "tenant",
		ResourceID: strconv.Itoa(int(tenant.ID)),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}
	h.tenantService.LogTenantActivity(activity)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tenant created successfully",
		"tenant":  tenant,
	})
}

// UpdateTenant handles updating a tenant.
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	tenant, err := h.tenantService.GetTenant(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	var req struct {
		Name         string `json:"name"`
		Domain       string `json:"domain"`
		Subdomain    string `json:"subdomain"`
		ContactEmail string `json:"contact_email"`
		BillingEmail string `json:"billing_email"`
		Plan         string `json:"plan"`
		Status       string `json:"status"`
		IsActive     *bool  `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update tenant request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update fields
	if req.Name != "" {
		tenant.Name = req.Name
	}
	if req.Domain != "" {
		tenant.Domain = req.Domain
	}
	if req.Subdomain != "" {
		tenant.Subdomain = req.Subdomain
	}
	if req.ContactEmail != "" {
		tenant.ContactEmail = req.ContactEmail
	}
	if req.BillingEmail != "" {
		tenant.BillingEmail = req.BillingEmail
	}
	if req.Plan != "" {
		tenant.Plan = req.Plan
	}
	if req.Status != "" {
		tenant.Status = req.Status
	}
	if req.IsActive != nil {
		tenant.IsActive = *req.IsActive
	}

	if err := h.tenantService.UpdateTenant(tenant); err != nil {
		h.logger.Error("Failed to update tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tenant"})
		return
	}

	// Log tenant activity
	activity := &models.TenantActivity{
		TenantID:   tenant.ID,
		Action:     "tenant_updated",
		Resource:   "tenant",
		ResourceID: strconv.Itoa(int(tenant.ID)),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}
	h.tenantService.LogTenantActivity(activity)

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant updated successfully",
		"tenant":  tenant,
	})
}

// DeleteTenant handles deleting a tenant.
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	if err := h.tenantService.DeleteTenant(uint(id)); err != nil {
		h.logger.Error("Failed to delete tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tenant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tenant deleted successfully"})
}

// GetTenantSettings handles getting tenant settings.
func (h *TenantHandler) GetTenantSettings(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	settings, err := h.tenantService.GetTenantSettings(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant settings not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

// UpdateTenantSettings handles updating tenant settings.
func (h *TenantHandler) UpdateTenantSettings(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var req struct {
		Timezone            string `json:"timezone"`
		Language            string `json:"language"`
		DateFormat          string `json:"date_format"`
		TimeFormat          string `json:"time_format"`
		EmailNotifications  *bool  `json:"email_notifications"`
		SMSNotifications    *bool  `json:"sms_notifications"`
		WebhookURL          string `json:"webhook_url"`
		ShowIncidentHistory *bool  `json:"show_incident_history"`
		ShowMaintenanceMode *bool  `json:"show_maintenance_mode"`
		CustomCSS           string `json:"custom_css"`
		CustomJS            string `json:"custom_js"`
		RequireAuth         *bool  `json:"require_auth"`
		AllowPublicAccess   *bool  `json:"allow_public_access"`
		SessionTimeout      *int   `json:"session_timeout"`
		APIRateLimit        *int   `json:"api_rate_limit"`
		APIKeyRequired      *bool  `json:"api_key_required"`
		AllowedOrigins      string `json:"allowed_origins"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update settings request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	settings, err := h.tenantService.GetTenantSettings(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant settings not found"})
		return
	}

	// Update fields
	if req.Timezone != "" {
		settings.Timezone = req.Timezone
	}
	if req.Language != "" {
		settings.Language = req.Language
	}
	if req.DateFormat != "" {
		settings.DateFormat = req.DateFormat
	}
	if req.TimeFormat != "" {
		settings.TimeFormat = req.TimeFormat
	}
	if req.EmailNotifications != nil {
		settings.EmailNotifications = *req.EmailNotifications
	}
	if req.SMSNotifications != nil {
		settings.SMSNotifications = *req.SMSNotifications
	}
	if req.WebhookURL != "" {
		settings.WebhookURL = req.WebhookURL
	}
	if req.ShowIncidentHistory != nil {
		settings.ShowIncidentHistory = *req.ShowIncidentHistory
	}
	if req.ShowMaintenanceMode != nil {
		settings.ShowMaintenanceMode = *req.ShowMaintenanceMode
	}
	if req.CustomCSS != "" {
		settings.CustomCSS = req.CustomCSS
	}
	if req.CustomJS != "" {
		settings.CustomJS = req.CustomJS
	}
	if req.RequireAuth != nil {
		settings.RequireAuth = *req.RequireAuth
	}
	if req.AllowPublicAccess != nil {
		settings.AllowPublicAccess = *req.AllowPublicAccess
	}
	if req.SessionTimeout != nil {
		settings.SessionTimeout = *req.SessionTimeout
	}
	if req.APIRateLimit != nil {
		settings.APIRateLimit = *req.APIRateLimit
	}
	if req.APIKeyRequired != nil {
		settings.APIKeyRequired = *req.APIKeyRequired
	}
	if req.AllowedOrigins != "" {
		settings.AllowedOrigins = req.AllowedOrigins
	}

	if err := h.tenantService.UpdateTenantSettings(settings); err != nil {
		h.logger.Error("Failed to update tenant settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Settings updated successfully",
		"settings": settings,
	})
}

// GetTenantBilling handles getting tenant billing information.
func (h *TenantHandler) GetTenantBilling(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	billing, err := h.tenantService.GetTenantBilling(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant billing not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"billing": billing})
}

// UpdateTenantBilling handles updating tenant billing information.
func (h *TenantHandler) UpdateTenantBilling(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var req struct {
		Plan                string  `json:"plan"`
		BillingCycle        string  `json:"billing_cycle"`
		Amount              float64 `json:"amount"`
		Currency            string  `json:"currency"`
		PaymentMethod       string  `json:"payment_method"`
		StripeCustomerID    string  `json:"stripe_customer_id"`
		StripeSubscriptionID string `json:"stripe_subscription_id"`
		MaxUsers            *int    `json:"max_users"`
		MaxComponents       *int    `json:"max_components"`
		MaxIncidents        *int    `json:"max_incidents"`
		MaxAPIRequests      *int    `json:"max_api_requests"`
		Status              string  `json:"status"`
		IsTrial             *bool   `json:"is_trial"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update billing request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	billing, err := h.tenantService.GetTenantBilling(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant billing not found"})
		return
	}

	// Update fields
	if req.Plan != "" {
		billing.Plan = req.Plan
	}
	if req.BillingCycle != "" {
		billing.BillingCycle = req.BillingCycle
	}
	if req.Amount > 0 {
		billing.Amount = req.Amount
	}
	if req.Currency != "" {
		billing.Currency = req.Currency
	}
	if req.PaymentMethod != "" {
		billing.PaymentMethod = req.PaymentMethod
	}
	if req.StripeCustomerID != "" {
		billing.StripeCustomerID = req.StripeCustomerID
	}
	if req.StripeSubscriptionID != "" {
		billing.StripeSubscriptionID = req.StripeSubscriptionID
	}
	if req.MaxUsers != nil {
		billing.MaxUsers = *req.MaxUsers
	}
	if req.MaxComponents != nil {
		billing.MaxComponents = *req.MaxComponents
	}
	if req.MaxIncidents != nil {
		billing.MaxIncidents = *req.MaxIncidents
	}
	if req.MaxAPIRequests != nil {
		billing.MaxAPIRequests = *req.MaxAPIRequests
	}
	if req.Status != "" {
		billing.Status = req.Status
	}
	if req.IsTrial != nil {
		billing.IsTrial = *req.IsTrial
	}

	if err := h.tenantService.UpdateTenantBilling(billing); err != nil {
		h.logger.Error("Failed to update tenant billing", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update billing"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Billing updated successfully",
		"billing": billing,
	})
}
