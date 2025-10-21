// Package handlers provides HTTP handlers for the Tenant Admin Service.
package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/anupamdutta5/tenant-admin-service/internal/models"
	"github.com/anupamdutta5/tenant-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Tenant Admin Management Handlers

// ListTenantAdmins handles listing tenant admins.
func (h *TenantAdminHandler) ListTenantAdmins(c *gin.Context) {
	tenantID, ok := ParseUUIDParam(c, "tenant_id")
	if !ok {
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	h.logger.Info("Listing tenant admins", zap.String("tenant_id", tenantID.String()))

	admins, err := h.service.ListTenantAdmins(c.Request.Context(), tenantID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list tenant admins", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to list tenant admins")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"admins": admins,
		"count":  len(admins),
		"limit":  limit,
		"offset": offset,
	})
}

// CreateTenantAdmin handles creating a new tenant admin.
func (h *TenantAdminHandler) CreateTenantAdmin(c *gin.Context) {
	tenantID, ok := ParseUUIDParam(c, "tenant_id")
	if !ok {
		return
	}

	var admin models.TenantAdmin
	if err := c.ShouldBindJSON(&admin); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	admin.TenantID = tenantID

	h.logger.Info("Creating tenant admin", zap.String("tenant_id", tenantID.String()))

	if err := h.service.CreateTenantAdmin(c.Request.Context(), &admin); err != nil {
		h.logger.Error("Failed to create tenant admin", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to create tenant admin")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"admin":   admin,
		"message": "Tenant admin created successfully",
	})
}

// GetTenantAdmin handles retrieving a tenant admin by ID.
func (h *TenantAdminHandler) GetTenantAdmin(c *gin.Context) {
	adminID, ok := ParseUintParam(c, "id")
	if !ok {
		return
	}

	h.logger.Info("Getting tenant admin", zap.Uint("admin_id", adminID))

	admin, err := h.service.GetTenantAdmin(c.Request.Context(), adminID)
	if err != nil {
		h.logger.Error("Failed to get tenant admin", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tenant admin not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"admin": admin,
	})
}

// UpdateTenantAdmin handles updating a tenant admin.
func (h *TenantAdminHandler) UpdateTenantAdmin(c *gin.Context) {
	adminID, ok := ParseUintParam(c, "id")
	if !ok {
		return
	}

	var updates models.TenantAdmin
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	h.logger.Info("Updating tenant admin", zap.Uint("admin_id", adminID))

	if err := h.service.UpdateTenantAdmin(c.Request.Context(), adminID, &updates); err != nil {
		h.logger.Error("Failed to update tenant admin", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to update tenant admin")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant admin updated successfully",
	})
}

// DeleteTenantAdmin handles deleting a tenant admin.
func (h *TenantAdminHandler) DeleteTenantAdmin(c *gin.Context) {
	adminID, ok := ParseUintParam(c, "id")
	if !ok {
		return
	}

	h.logger.Info("Deleting tenant admin", zap.Uint("admin_id", adminID))

	if err := h.service.DeleteTenantAdmin(c.Request.Context(), adminID); err != nil {
		h.logger.Error("Failed to delete tenant admin", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to delete tenant admin")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant admin deleted successfully",
	})
}

// Tenant Settings Management Handlers

// GetTenantSettings handles retrieving tenant settings.
func (h *TenantAdminHandler) GetTenantSettings(c *gin.Context) {
	tenantID, ok := ParseUUIDParam(c, "tenant_id")
	if !ok {
		return
	}

	h.logger.Info("Getting tenant settings", zap.String("tenant_id", tenantID.String()))

	settings, err := h.service.GetTenantSettings(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get tenant settings", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to get tenant settings")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": settings,
	})
}

// UpdateTenantSettings handles updating tenant settings.
func (h *TenantAdminHandler) UpdateTenantSettings(c *gin.Context) {
	tenantID, ok := ParseUUIDParam(c, "tenant_id")
	if !ok {
		return
	}

	var updates models.TenantSettings
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	h.logger.Info("Updating tenant settings", zap.String("tenant_id", tenantID.String()))

	if err := h.service.UpdateTenantSettings(c.Request.Context(), tenantID, &updates); err != nil {
		h.logger.Error("Failed to update tenant settings", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to update tenant settings")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant settings updated successfully",
	})
}

// Feature Flag Management Handlers

// ListTenantFeatureFlags handles listing tenant feature flags.
func (h *TenantAdminHandler) ListTenantFeatureFlags(c *gin.Context) {
	tenantID, ok := ParseUUIDParam(c, "tenant_id")
	if !ok {
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	h.logger.Info("Listing tenant feature flags", zap.String("tenant_id", tenantID.String()))

	flags, err := h.service.ListTenantFeatureFlags(c.Request.Context(), tenantID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list tenant feature flags", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to list tenant feature flags")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"feature_flags": flags,
		"count":         len(flags),
		"limit":         limit,
		"offset":        offset,
	})
}

// CreateTenantFeatureFlag handles creating a new tenant feature flag.
func (h *TenantAdminHandler) CreateTenantFeatureFlag(c *gin.Context) {
	tenantID, ok := ParseUUIDParam(c, "tenant_id")
	if !ok {
		return
	}

	var flag models.TenantFeatureFlag
	if err := c.ShouldBindJSON(&flag); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	flag.TenantID = tenantID

	h.logger.Info("Creating tenant feature flag", zap.String("tenant_id", tenantID.String()))

	if err := h.service.CreateTenantFeatureFlag(c.Request.Context(), &flag); err != nil {
		h.logger.Error("Failed to create tenant feature flag", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to create tenant feature flag")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"feature_flag": flag,
		"message":      "Tenant feature flag created successfully",
	})
}

// GetTenantFeatureFlag handles retrieving a tenant feature flag by ID.
func (h *TenantAdminHandler) GetTenantFeatureFlag(c *gin.Context) {
	flagID, ok := ParseUintParam(c, "id")
	if !ok {
		return
	}

	h.logger.Info("Getting tenant feature flag", zap.Uint("flag_id", flagID))

	flag, err := h.service.GetTenantFeatureFlag(c.Request.Context(), flagID)
	if err != nil {
		h.logger.Error("Failed to get tenant feature flag", zap.Error(err))
		RespondWithError(c, http.StatusNotFound, "Tenant feature flag not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"feature_flag": flag,
	})
}

// UpdateTenantFeatureFlag handles updating a tenant feature flag.
func (h *TenantAdminHandler) UpdateTenantFeatureFlag(c *gin.Context) {
	flagID, ok := ParseUintParam(c, "id")
	if !ok {
		return
	}

	var updates models.TenantFeatureFlag
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	h.logger.Info("Updating tenant feature flag", zap.Uint("flag_id", flagID))

	if err := h.service.UpdateTenantFeatureFlag(c.Request.Context(), flagID, &updates); err != nil {
		h.logger.Error("Failed to update tenant feature flag", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to update tenant feature flag")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant feature flag updated successfully",
	})
}

// DeleteTenantFeatureFlag handles deleting a tenant feature flag.
func (h *TenantAdminHandler) DeleteTenantFeatureFlag(c *gin.Context) {
	flagID, ok := ParseUintParam(c, "id")
	if !ok {
		return
	}

	h.logger.Info("Deleting tenant feature flag", zap.Uint("flag_id", flagID))

	if err := h.service.DeleteTenantFeatureFlag(c.Request.Context(), flagID); err != nil {
		h.logger.Error("Failed to delete tenant feature flag", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to delete tenant feature flag")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant feature flag deleted successfully",
	})
}

// Usage Management Handlers

// GetTenantUsage handles retrieving tenant usage metrics.
func (h *TenantAdminHandler) GetTenantUsage(c *gin.Context) {
	tenantID, ok := ParseUUIDParam(c, "tenant_id")
	if !ok {
		return
	}

	// Parse date parameters
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid start date format")
		return
	}

	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid end date format",
		})
		return
	}

	h.logger.Info("Getting tenant usage", zap.String("tenant_id", tenantID.String()))

	usage, err := h.service.GetTenantUsage(c.Request.Context(), tenantID, startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to get tenant usage", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to get tenant usage")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"usage": usage,
		"count": len(usage),
	})
}

// RecordTenantUsage handles recording tenant usage metrics.
func (h *TenantAdminHandler) RecordTenantUsage(c *gin.Context) {
	tenantID, ok := ParseUUIDParam(c, "tenant_id")
	if !ok {
		return
	}

	var usage models.TenantUsage
	if err := c.ShouldBindJSON(&usage); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	usage.TenantID = tenantID

	h.logger.Info("Recording tenant usage", zap.String("tenant_id", tenantID.String()))

	if err := h.service.RecordTenantUsage(c.Request.Context(), &usage); err != nil {
		h.logger.Error("Failed to record tenant usage", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to record tenant usage")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"usage":   usage,
		"message": "Tenant usage recorded successfully",
	})
}

// Statistics Handlers

// GetTenantStats handles getting tenant statistics.
func (h *TenantAdminHandler) GetTenantStats(c *gin.Context) {
	tenantID, ok := ParseUUIDParam(c, "tenant_id")
	if !ok {
		return
	}

	h.logger.Info("Getting tenant statistics", zap.String("tenant_id", tenantID.String()))

	stats, err := h.service.GetTenantStats(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get tenant statistics", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to get tenant statistics")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// GetTenantUserStats handles getting tenant user statistics including current count and limits.
func (h *TenantAdminHandler) GetTenantUserStats(c *gin.Context) {
	tenantID, ok := ParseUUIDParam(c, "tenant_id")
	if !ok {
		return
	}

	h.logger.Info("Getting tenant user statistics", zap.String("tenant_id", tenantID.String()))

	stats, err := h.service.GetTenantUserStats(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get tenant user statistics", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to get tenant user statistics")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

// Placeholder handlers for remaining endpoints
func (h *TenantAdminHandler) GetTenantBilling(c *gin.Context) {
	RespondWithError(c, http.StatusNotImplemented, "Not implemented")
}

func (h *TenantAdminHandler) UpdateTenantBilling(c *gin.Context) {
	RespondWithError(c, http.StatusNotImplemented, "Not implemented")
}

func (h *TenantAdminHandler) ListTenantNotifications(c *gin.Context) {
	RespondWithError(c, http.StatusNotImplemented, "Not implemented")
}

func (h *TenantAdminHandler) CreateTenantNotification(c *gin.Context) {
	RespondWithError(c, http.StatusNotImplemented, "Not implemented")
}

func (h *TenantAdminHandler) GetTenantNotification(c *gin.Context) {
	RespondWithError(c, http.StatusNotImplemented, "Not implemented")
}

func (h *TenantAdminHandler) UpdateTenantNotification(c *gin.Context) {
	RespondWithError(c, http.StatusNotImplemented, "Not implemented")
}

func (h *TenantAdminHandler) DeleteTenantNotification(c *gin.Context) {
	RespondWithError(c, http.StatusNotImplemented, "Not implemented")
}

func (h *TenantAdminHandler) ListTenantActivities(c *gin.Context) {
	RespondWithError(c, http.StatusNotImplemented, "Not implemented")
}

func (h *TenantAdminHandler) ListTenantBackups(c *gin.Context) {
	RespondWithError(c, http.StatusNotImplemented, "Not implemented")
}

func (h *TenantAdminHandler) CreateTenantBackup(c *gin.Context) {
	RespondWithError(c, http.StatusNotImplemented, "Not implemented")
}

func (h *TenantAdminHandler) GetTenantBackup(c *gin.Context) {
	RespondWithError(c, http.StatusNotImplemented, "Not implemented")
}

func (h *TenantAdminHandler) DeleteTenantBackup(c *gin.Context) {
	RespondWithError(c, http.StatusNotImplemented, "Not implemented")
}

// Status Page Management Handlers

// GetStatusPageData retrieves status page data for rendering.
func (h *TenantAdminHandler) GetStatusPageData(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		RespondWithError(c, http.StatusBadRequest, "Slug is required")
		return
	}

	// Get tenant ID from query parameter or header
	tenantID, err := h.getTenantID(c)
	if err != nil {
		h.logger.Error("Failed to get tenant ID", zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	// Get status page data
	data, err := h.statusPageService.GetStatusPageData(tenantID, slug)
	if err != nil {
		h.logger.Error("Failed to get status page data", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to load status page data")
		return
	}

	c.JSON(http.StatusOK, data)
}

// CreateStatusPage creates a new status page.
func (h *TenantAdminHandler) CreateStatusPage(c *gin.Context) {
	// Get tenant ID from query parameter or header
	tenantID, err := h.getTenantID(c)
	if err != nil {
		h.logger.Error("Failed to get tenant ID", zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	var statusPage models.StatusPage
	if err := c.ShouldBindJSON(&statusPage); err != nil {
		RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.statusPageService.CreateStatusPage(tenantID, &statusPage); err != nil {
		h.logger.Error("Failed to create status page", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to create status page")
		return
	}

	c.JSON(http.StatusCreated, statusPage)
}

// UpdateStatusPage updates an existing status page.
func (h *TenantAdminHandler) UpdateStatusPage(c *gin.Context) {
	// Get tenant ID from query parameter or header
	tenantID, err := h.getTenantID(c)
	if err != nil {
		h.logger.Error("Failed to get tenant ID", zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	var statusPage models.StatusPage
	if err := c.ShouldBindJSON(&statusPage); err != nil {
		RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.statusPageService.UpdateStatusPage(tenantID, &statusPage); err != nil {
		h.logger.Error("Failed to update status page", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to update status page")
		return
	}

	c.JSON(http.StatusOK, statusPage)
}

// GetStatusPages retrieves status pages for a tenant.
func (h *TenantAdminHandler) GetStatusPages(c *gin.Context) {
	// Get tenant ID from query parameter or header
	tenantID, err := h.getTenantID(c)
	if err != nil {
		h.logger.Error("Failed to get tenant ID", zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	// Get pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	statusPages, total, err := h.statusPageService.GetStatusPages(tenantID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get status pages", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to get status pages")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status_pages": statusPages,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	})
}

// DeleteStatusPage deletes a status page.
func (h *TenantAdminHandler) DeleteStatusPage(c *gin.Context) {
	// Get tenant ID from query parameter or header
	tenantID, err := h.getTenantID(c)
	if err != nil {
		h.logger.Error("Failed to get tenant ID", zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	statusPageID, ok := ParseUintParam(c, "id")
	if !ok {
		return
	}

	if err := h.statusPageService.DeleteStatusPage(tenantID, statusPageID); err != nil {
		h.logger.Error("Failed to delete status page", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to delete status page")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status page deleted successfully"})
}

// UpdateStatusPageConfig updates status page configuration.
func (h *TenantAdminHandler) UpdateStatusPageConfig(c *gin.Context) {
	// Get tenant ID from query parameter or header
	tenantID, err := h.getTenantID(c)
	if err != nil {
		h.logger.Error("Failed to get tenant ID", zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	statusPageID, ok := ParseUintParam(c, "id")
	if !ok {
		return
	}

	var config models.StatusPageConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.statusPageService.UpdateStatusPageConfig(tenantID, statusPageID, &config); err != nil {
		h.logger.Error("Failed to update status page config", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to update status page config")
		return
	}

	c.JSON(http.StatusOK, config)
}

// Authentication Handlers

// ============================================================================
// CORE TENANT CRUD HANDLERS (migrated from tenant-service)
// ============================================================================

// GetTenants handles getting a list of tenants.
func (h *TenantAdminHandler) GetTenants(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	tenants, total, err := h.service.GetTenants(limit, offset)
	if err != nil {
		h.logger.Error("Failed to get tenants", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to get tenants")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   tenants,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// CreateTenant handles creating a new tenant.
// CreateTenantRequest represents the request to create a tenant with admin credentials
type CreateTenantRequest struct {
	ID            string `json:"id"` // Optional: Use provided UUID from SaaS Admin
	Name          string `json:"name" binding:"required"`
	Slug          string `json:"slug" binding:"required"`
	ContactEmail  string `json:"contact_email" binding:"required"`
	Domain        string `json:"domain"`
	Subdomain     string `json:"subdomain"`
	AdminEmail    string `json:"admin_email" binding:"required,email"`
	AdminPassword string `json:"admin_password" binding:"required,min=8"`
}

func (h *TenantAdminHandler) CreateTenant(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind tenant data", zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid tenant data: "+err.Error())
		return
	}

	// Create tenant model
	tenant := &models.Tenant{
		Name:         req.Name,
		Slug:         req.Slug,
		ContactEmail: req.ContactEmail,
		Domain:       req.Domain,
		Subdomain:    req.Subdomain,
		Status:       "active",
		IsActive:     true,
	}

	// If ID is provided (from SaaS Admin), use it to maintain UUID consistency
	if req.ID != "" {
		tenantUUID, err := uuid.Parse(req.ID)
		if err != nil {
			h.logger.Error("Invalid tenant ID provided", zap.Error(err), zap.String("id", req.ID))
			RespondWithError(c, http.StatusBadRequest, "Invalid tenant ID format")
			return
		}
		tenant.ID = tenantUUID
	}

	h.logger.Info("Creating tenant",
		zap.String("name", req.Name),
		zap.String("slug", req.Slug))

	if err := h.service.CreateTenant(tenant); err != nil {
		h.logger.Error("Failed to create tenant", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to create tenant")
		return
	}

	// Create admin user with hashed password
	if err := h.service.CreateAdminUser(c.Request.Context(), tenant.ID, req.AdminEmail, req.AdminPassword); err != nil {
		h.logger.Error("Failed to create admin user", zap.Error(err))
		// Tenant is created but admin user failed - log but don't rollback
		c.JSON(http.StatusCreated, gin.H{
			"message": "Tenant created but failed to create admin user",
			"data":    tenant,
			"warning": "Admin user creation failed",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tenant created successfully",
		"data":    tenant,
	})
}

// SyncTenantRequest represents tenant sync data from SaaS Admin
type SyncTenantRequest struct {
	ID           string `json:"id" binding:"required"`          // UUID from SaaS Admin
	Name         string `json:"name" binding:"required"`
	Slug         string `json:"slug" binding:"required"`
	ContactEmail string `json:"contact_email" binding:"required"`
	Domain       string `json:"domain"`
	Subdomain    string `json:"subdomain"`
	PlanID       string `json:"plan_id"`
	MaxUsers     *int64 `json:"max_users"`
}

// SyncTenant handles idempotent tenant synchronization from SaaS Admin
// This endpoint is designed to be called by SaaS Admin service when a tenant is created/updated
// It ensures eventual consistency between saas_admin and tenant_admin_db databases
func (h *TenantAdminHandler) SyncTenant(c *gin.Context) {
	var req SyncTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind sync tenant data", zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid sync data: "+err.Error())
		return
	}

	// Parse UUID
	tenantUUID, err := uuid.Parse(req.ID)
	if err != nil {
		h.logger.Error("Invalid tenant UUID", zap.Error(err), zap.String("id", req.ID))
		RespondWithError(c, http.StatusBadRequest, "Invalid tenant ID format")
		return
	}

	h.logger.Info("Syncing tenant from SaaS Admin",
		zap.String("tenant_id", req.ID),
		zap.String("name", req.Name),
		zap.String("slug", req.Slug))

	// Check if tenant already exists (idempotent)
	existingTenant, err := h.service.GetTenantByUUID(tenantUUID)
	if err == nil && existingTenant != nil {
		// Tenant exists - update it
		h.logger.Info("Tenant already exists, updating",
			zap.String("tenant_id", req.ID),
			zap.String("slug", req.Slug))

		existingTenant.Name = req.Name
		existingTenant.Slug = req.Slug
		existingTenant.ContactEmail = req.ContactEmail
		existingTenant.Domain = req.Domain
		existingTenant.Subdomain = req.Subdomain

		if err := h.service.UpdateTenant(existingTenant); err != nil {
			h.logger.Error("Failed to update existing tenant during sync", zap.Error(err))
			RespondWithError(c, http.StatusInternalServerError, "Failed to sync tenant (update failed)")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Tenant synced successfully (updated)",
			"data":    existingTenant,
			"action":  "updated",
		})
		return
	}

	// Tenant doesn't exist - create it with complete data
	var maxUsers *int
	if req.MaxUsers != nil {
		val := int(*req.MaxUsers)
		maxUsers = &val
	}

	tenant := &models.Tenant{
		ID:           tenantUUID,
		Name:         req.Name,
		Slug:         req.Slug,
		ContactEmail: req.ContactEmail,
		Domain:       req.Domain,
		Subdomain:    req.Subdomain,
		Status:       "active",
		IsActive:     true,
		MaxUsers:     maxUsers,
	}

	if err := h.service.SyncTenantFromSaaS(tenant); err != nil {
		h.logger.Error("Failed to create tenant during sync", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to sync tenant (create failed)")
		return
	}

	// Fetch the created tenant to get all fields populated
	syncedTenant, err := h.service.GetTenantByUUID(tenantUUID)
	if err != nil {
		h.logger.Warn("Tenant created but failed to fetch", zap.Error(err))
		// Use the tenant we created as fallback
		syncedTenant = tenant
	}

	h.logger.Info("Tenant synced successfully (created)",
		zap.String("tenant_id", req.ID),
		zap.String("slug", req.Slug))

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tenant synced successfully (created)",
		"data":    syncedTenant,
		"action":  "created",
	})
}

// GetTenant handles getting a tenant by ID.
func (h *TenantAdminHandler) GetTenant(c *gin.Context) {
	id, ok := ParseUintParam(c, "id")
	if !ok {
		return
	}

	tenant, err := h.service.GetTenant(id)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			RespondWithError(c, http.StatusNotFound, "Tenant not found")
			return
		}
		h.logger.Error("Failed to get tenant", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to get tenant")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tenant})
}

// GetTenantBySlug handles getting a tenant by slug.
func (h *TenantAdminHandler) GetTenantBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		RespondWithError(c, http.StatusBadRequest, "Tenant slug is required")
		return
	}

	tenant, err := h.service.GetTenantBySlug(slug)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			RespondWithError(c, http.StatusNotFound, "Tenant not found")
			return
		}
		h.logger.Error("Failed to get tenant by slug", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to get tenant")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tenant})
}

// GetTenantByDomain handles getting a tenant by domain.
func (h *TenantAdminHandler) GetTenantByDomain(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		RespondWithError(c, http.StatusBadRequest, "Domain is required")
		return
	}

	tenant, err := h.service.GetTenantByDomain(domain)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			RespondWithError(c, http.StatusNotFound, "Tenant not found")
			return
		}
		h.logger.Error("Failed to get tenant by domain", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to get tenant")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tenant})
}

// UpdateTenant handles updating a tenant.
func (h *TenantAdminHandler) UpdateTenant(c *gin.Context) {
	id, ok := ParseUintParam(c, "id")
	if !ok {
		return
	}

	// Get existing tenant
	tenant, err := h.service.GetTenant(id)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			RespondWithError(c, http.StatusNotFound, "Tenant not found")
			return
		}
		h.logger.Error("Failed to get tenant", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to get tenant")
		return
	}

	// Bind updates
	if err := c.ShouldBindJSON(tenant); err != nil {
		h.logger.Error("Failed to bind tenant data", zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid tenant data")
		return
	}

	if err := h.service.UpdateTenant(tenant); err != nil{
		h.logger.Error("Failed to update tenant", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to update tenant")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant updated successfully",
		"data":    tenant,
	})
}

// DeleteTenant handles deleting a tenant.
func (h *TenantAdminHandler) DeleteTenant(c *gin.Context) {
	tenantID, ok := ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteTenantByUUID(c.Request.Context(), tenantID); err != nil {
		h.logger.Error("Failed to delete tenant", zap.Error(err), zap.String("tenant_id", tenantID.String()))
		RespondWithError(c, http.StatusInternalServerError, "Failed to delete tenant")
		return
	}

	h.logger.Info("Tenant deleted successfully", zap.String("tenant_id", tenantID.String()))
	c.JSON(http.StatusOK, gin.H{"message": "Tenant deleted successfully"})
}

// GetTenantBranding handles getting tenant branding information.
func (h *TenantAdminHandler) GetTenantBranding(c *gin.Context) {
	tenantID, err := h.getTenantID(c)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	branding, err := h.service.GetTenantBranding(tenantID)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			RespondWithError(c, http.StatusNotFound, "Tenant branding not found")
			return
		}
		h.logger.Error("Failed to get tenant branding", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to get tenant branding")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": branding})
}

// UpdateTenantBranding handles updating tenant branding information.
func (h *TenantAdminHandler) UpdateTenantBranding(c *gin.Context) {
	tenantID, err := h.getTenantID(c)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	// Get existing branding
	branding, err := h.service.GetTenantBranding(tenantID)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			RespondWithError(c, http.StatusNotFound, "Tenant branding not found")
			return
		}
		h.logger.Error("Failed to get tenant branding", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to get tenant branding")
		return
	}

	// Bind updates
	if err := c.ShouldBindJSON(branding); err != nil {
		h.logger.Error("Failed to bind branding data", zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid branding data")
		return
	}

	if err := h.service.UpdateTenantBranding(branding); err != nil {
		h.logger.Error("Failed to update tenant branding", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to update tenant branding")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant branding updated successfully",
		"data":    branding,
	})
}

