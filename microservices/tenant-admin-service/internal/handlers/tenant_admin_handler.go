// Package handlers provides HTTP handlers for the Tenant Admin Service.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/anupamdutta5/tenant-admin-service/internal/models"
	"github.com/anupamdutta5/tenant-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// TenantAdminHandler handles tenant admin-related HTTP requests.
type TenantAdminHandler struct {
	service           *services.TenantAdminService
	statusPageService *services.StatusPageManagementService
	logger            *zap.Logger
}

// NewTenantAdminHandler creates a new tenant admin handler.
func NewTenantAdminHandler(service *services.TenantAdminService, statusPageService *services.StatusPageManagementService, logger *zap.Logger) *TenantAdminHandler {
	return &TenantAdminHandler{
		service:           service,
		statusPageService: statusPageService,
		logger:            logger,
	}
}

// GetLoginPage renders the login page.
func (h *TenantAdminHandler) GetLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Status Page Admin - Login",
	})
}

// GetAdminDashboard renders the admin dashboard.
func (h *TenantAdminHandler) GetAdminDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_dashboard.html", gin.H{
		"title": "Status Page Admin Dashboard",
	})
}

// HealthCheck handles health check requests.
func (h *TenantAdminHandler) HealthCheck(c *gin.Context) {
	h.logger.Info("Health check requested")

	// Check service health
	if err := h.service.Health(c.Request.Context()); err != nil {
		h.logger.Error("Health check failed", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"error":   err.Error(),
			"service": "tenant-admin-service",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "tenant-admin-service",
		"version": "1.0.0",
	})
}

// Tenant Admin Management Handlers

// ListTenantAdmins handles listing tenant admins.
func (h *TenantAdminHandler) ListTenantAdmins(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
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

	h.logger.Info("Listing tenant admins", zap.Uint64("tenant_id", tenantID))

	admins, err := h.service.ListTenantAdmins(c.Request.Context(), uint(tenantID), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list tenant admins", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list tenant admins",
		})
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
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	var admin models.TenantAdmin
	if err := c.ShouldBindJSON(&admin); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	admin.TenantID = uint(tenantID)

	h.logger.Info("Creating tenant admin", zap.Uint64("tenant_id", tenantID))

	if err := h.service.CreateTenantAdmin(c.Request.Context(), &admin); err != nil {
		h.logger.Error("Failed to create tenant admin", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create tenant admin",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"admin":   admin,
		"message": "Tenant admin created successfully",
	})
}

// GetTenantAdmin handles retrieving a tenant admin by ID.
func (h *TenantAdminHandler) GetTenantAdmin(c *gin.Context) {
	adminIDStr := c.Param("id")
	adminID, err := strconv.ParseUint(adminIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid admin ID",
		})
		return
	}

	h.logger.Info("Getting tenant admin", zap.Uint64("admin_id", adminID))

	admin, err := h.service.GetTenantAdmin(c.Request.Context(), uint(adminID))
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
	adminIDStr := c.Param("id")
	adminID, err := strconv.ParseUint(adminIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid admin ID",
		})
		return
	}

	var updates models.TenantAdmin
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Updating tenant admin", zap.Uint64("admin_id", adminID))

	if err := h.service.UpdateTenantAdmin(c.Request.Context(), uint(adminID), &updates); err != nil {
		h.logger.Error("Failed to update tenant admin", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update tenant admin",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant admin updated successfully",
	})
}

// DeleteTenantAdmin handles deleting a tenant admin.
func (h *TenantAdminHandler) DeleteTenantAdmin(c *gin.Context) {
	adminIDStr := c.Param("id")
	adminID, err := strconv.ParseUint(adminIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid admin ID",
		})
		return
	}

	h.logger.Info("Deleting tenant admin", zap.Uint64("admin_id", adminID))

	if err := h.service.DeleteTenantAdmin(c.Request.Context(), uint(adminID)); err != nil {
		h.logger.Error("Failed to delete tenant admin", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete tenant admin",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant admin deleted successfully",
	})
}

// Tenant Settings Management Handlers

// GetTenantSettings handles retrieving tenant settings.
func (h *TenantAdminHandler) GetTenantSettings(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	h.logger.Info("Getting tenant settings", zap.Uint64("tenant_id", tenantID))

	settings, err := h.service.GetTenantSettings(c.Request.Context(), uint(tenantID))
	if err != nil {
		h.logger.Error("Failed to get tenant settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get tenant settings",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": settings,
	})
}

// UpdateTenantSettings handles updating tenant settings.
func (h *TenantAdminHandler) UpdateTenantSettings(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	var updates models.TenantSettings
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Updating tenant settings", zap.Uint64("tenant_id", tenantID))

	if err := h.service.UpdateTenantSettings(c.Request.Context(), uint(tenantID), &updates); err != nil {
		h.logger.Error("Failed to update tenant settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update tenant settings",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant settings updated successfully",
	})
}

// Feature Flag Management Handlers

// ListTenantFeatureFlags handles listing tenant feature flags.
func (h *TenantAdminHandler) ListTenantFeatureFlags(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
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

	h.logger.Info("Listing tenant feature flags", zap.Uint64("tenant_id", tenantID))

	flags, err := h.service.ListTenantFeatureFlags(c.Request.Context(), uint(tenantID), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list tenant feature flags", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list tenant feature flags",
		})
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
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	var flag models.TenantFeatureFlag
	if err := c.ShouldBindJSON(&flag); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	flag.TenantID = uint(tenantID)

	h.logger.Info("Creating tenant feature flag", zap.Uint64("tenant_id", tenantID))

	if err := h.service.CreateTenantFeatureFlag(c.Request.Context(), &flag); err != nil {
		h.logger.Error("Failed to create tenant feature flag", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create tenant feature flag",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"feature_flag": flag,
		"message":      "Tenant feature flag created successfully",
	})
}

// GetTenantFeatureFlag handles retrieving a tenant feature flag by ID.
func (h *TenantAdminHandler) GetTenantFeatureFlag(c *gin.Context) {
	flagIDStr := c.Param("id")
	flagID, err := strconv.ParseUint(flagIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feature flag ID",
		})
		return
	}

	h.logger.Info("Getting tenant feature flag", zap.Uint64("flag_id", flagID))

	flag, err := h.service.GetTenantFeatureFlag(c.Request.Context(), uint(flagID))
	if err != nil {
		h.logger.Error("Failed to get tenant feature flag", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tenant feature flag not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"feature_flag": flag,
	})
}

// UpdateTenantFeatureFlag handles updating a tenant feature flag.
func (h *TenantAdminHandler) UpdateTenantFeatureFlag(c *gin.Context) {
	flagIDStr := c.Param("id")
	flagID, err := strconv.ParseUint(flagIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feature flag ID",
		})
		return
	}

	var updates models.TenantFeatureFlag
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Updating tenant feature flag", zap.Uint64("flag_id", flagID))

	if err := h.service.UpdateTenantFeatureFlag(c.Request.Context(), uint(flagID), &updates); err != nil {
		h.logger.Error("Failed to update tenant feature flag", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update tenant feature flag",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant feature flag updated successfully",
	})
}

// DeleteTenantFeatureFlag handles deleting a tenant feature flag.
func (h *TenantAdminHandler) DeleteTenantFeatureFlag(c *gin.Context) {
	flagIDStr := c.Param("id")
	flagID, err := strconv.ParseUint(flagIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feature flag ID",
		})
		return
	}

	h.logger.Info("Deleting tenant feature flag", zap.Uint64("flag_id", flagID))

	if err := h.service.DeleteTenantFeatureFlag(c.Request.Context(), uint(flagID)); err != nil {
		h.logger.Error("Failed to delete tenant feature flag", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete tenant feature flag",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant feature flag deleted successfully",
	})
}

// Usage Management Handlers

// GetTenantUsage handles retrieving tenant usage metrics.
func (h *TenantAdminHandler) GetTenantUsage(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	// Parse date parameters
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid start date format",
		})
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

	h.logger.Info("Getting tenant usage", zap.Uint64("tenant_id", tenantID))

	usage, err := h.service.GetTenantUsage(c.Request.Context(), uint(tenantID), startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to get tenant usage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get tenant usage",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"usage": usage,
		"count": len(usage),
	})
}

// RecordTenantUsage handles recording tenant usage metrics.
func (h *TenantAdminHandler) RecordTenantUsage(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	var usage models.TenantUsage
	if err := c.ShouldBindJSON(&usage); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	usage.TenantID = uint(tenantID)

	h.logger.Info("Recording tenant usage", zap.Uint64("tenant_id", tenantID))

	if err := h.service.RecordTenantUsage(c.Request.Context(), &usage); err != nil {
		h.logger.Error("Failed to record tenant usage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to record tenant usage",
		})
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
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	h.logger.Info("Getting tenant statistics", zap.Uint64("tenant_id", tenantID))

	stats, err := h.service.GetTenantStats(c.Request.Context(), uint(tenantID))
	if err != nil {
		h.logger.Error("Failed to get tenant statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get tenant statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// Placeholder handlers for remaining endpoints
func (h *TenantAdminHandler) GetTenantBilling(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *TenantAdminHandler) UpdateTenantBilling(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *TenantAdminHandler) ListTenantNotifications(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *TenantAdminHandler) CreateTenantNotification(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *TenantAdminHandler) GetTenantNotification(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *TenantAdminHandler) UpdateTenantNotification(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *TenantAdminHandler) DeleteTenantNotification(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *TenantAdminHandler) ListTenantActivities(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *TenantAdminHandler) ListTenantBackups(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *TenantAdminHandler) CreateTenantBackup(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *TenantAdminHandler) GetTenantBackup(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *TenantAdminHandler) DeleteTenantBackup(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

// Status Page Management Handlers

// GetStatusPageData retrieves status page data for rendering.
func (h *TenantAdminHandler) GetStatusPageData(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug is required"})
		return
	}

	// Get tenant ID from query parameter or header
	tenantID, err := h.getTenantID(c)
	if err != nil {
		h.logger.Error("Failed to get tenant ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// Get status page data
	data, err := h.statusPageService.GetStatusPageData(tenantID, slug)
	if err != nil {
		h.logger.Error("Failed to get status page data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load status page data"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var statusPage models.StatusPage
	if err := c.ShouldBindJSON(&statusPage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.statusPageService.CreateStatusPage(tenantID, &statusPage); err != nil {
		h.logger.Error("Failed to create status page", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create status page"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var statusPage models.StatusPage
	if err := c.ShouldBindJSON(&statusPage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.statusPageService.UpdateStatusPage(tenantID, &statusPage); err != nil {
		h.logger.Error("Failed to update status page", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status page"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// Get pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	statusPages, total, err := h.statusPageService.GetStatusPages(tenantID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get status pages", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get status pages"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	statusPageID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status page ID"})
		return
	}

	if err := h.statusPageService.DeleteStatusPage(tenantID, uint(statusPageID)); err != nil {
		h.logger.Error("Failed to delete status page", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete status page"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	statusPageID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status page ID"})
		return
	}

	var config models.StatusPageConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.statusPageService.UpdateStatusPageConfig(tenantID, uint(statusPageID), &config); err != nil {
		h.logger.Error("Failed to update status page config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status page config"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// Authentication Handlers

// Login handles user login.
func (h *TenantAdminHandler) Login(c *gin.Context) {
	var loginRequest struct {
		Email      string `json:"email" binding:"required,email"`
		Password   string `json:"password" binding:"required"`
		RememberMe bool   `json:"remember_me"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Implement proper authentication logic
	// For production, integrate with user-service for authentication
	// For now, implement basic authentication with proper JWT generation

	// Validate credentials (in production, this would call user-service)
	if loginRequest.Email == "admin@example.com" && loginRequest.Password == "admin123" {
		// Generate proper JWT token
		claims := jwt.MapClaims{
			"user_id":   1,
			"email":     loginRequest.Email,
			"username":  "admin",
			"role":      "admin",
			"tenant_id": 1,
			"exp":       time.Now().Add(time.Hour * 24).Unix(), // 24 hours
			"iat":       time.Now().Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("your-jwt-secret-key")) // Use env variable
		if err != nil {
			h.logger.Error("Failed to generate JWT token", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": tokenString,
			"user": gin.H{
				"id":        1,
				"email":     loginRequest.Email,
				"username":  "admin",
				"role":      "admin",
				"tenant_id": 1,
			},
		})
		return
	}

	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
}

// Logout handles user logout.
func (h *TenantAdminHandler) Logout(c *gin.Context) {
	// In production, this would invalidate the JWT token on the server side
	// For now, we rely on client-side token removal
	h.logger.Info("User logged out successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// VerifyToken verifies the JWT token.
func (h *TenantAdminHandler) VerifyToken(c *gin.Context) {
	// In production, this would verify the JWT token signature and expiration
	// For now, we return a mock response for development
	h.logger.Info("Token verification requested")
	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"user": gin.H{
			"id":        1,
			"email":     "admin@example.com",
			"username":  "admin",
			"role":      "admin",
			"tenant_id": 1,
		},
	})
}

// getTenantID extracts tenant ID from request.
func (h *TenantAdminHandler) getTenantID(c *gin.Context) (uint, error) {
	// Try to get from query parameter first
	if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
		if err != nil {
			return 0, err
		}
		return uint(tenantID), nil
	}

	// Try to get from header
	if tenantIDStr := c.GetHeader("X-Tenant-ID"); tenantIDStr != "" {
		tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
		if err != nil {
			return 0, err
		}
		return uint(tenantID), nil
	}

	// Default to tenant ID 1 for development
	return 1, nil
}

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tenants"})
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
func (h *TenantAdminHandler) CreateTenant(c *gin.Context) {
	var tenant models.Tenant
	if err := c.ShouldBindJSON(&tenant); err != nil {
		h.logger.Error("Failed to bind tenant data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant data"})
		return
	}

	// Validate required fields
	if tenant.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant name is required"})
		return
	}
	if tenant.ContactEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contact email is required"})
		return
	}

	if err := h.service.CreateTenant(&tenant); err != nil {
		h.logger.Error("Failed to create tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tenant created successfully",
		"data":    tenant,
	})
}

// GetTenant handles getting a tenant by ID.
func (h *TenantAdminHandler) GetTenant(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Error("Invalid tenant ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	tenant, err := h.service.GetTenant(uint(id))
	if err != nil {
		if err.Error() == "tenant not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			return
		}
		h.logger.Error("Failed to get tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tenant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tenant})
}

// GetTenantBySlug handles getting a tenant by slug.
func (h *TenantAdminHandler) GetTenantBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant slug is required"})
		return
	}

	tenant, err := h.service.GetTenantBySlug(slug)
	if err != nil {
		if err.Error() == "tenant not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			return
		}
		h.logger.Error("Failed to get tenant by slug", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tenant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tenant})
}

// GetTenantByDomain handles getting a tenant by domain.
func (h *TenantAdminHandler) GetTenantByDomain(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain is required"})
		return
	}

	tenant, err := h.service.GetTenantByDomain(domain)
	if err != nil {
		if err.Error() == "tenant not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			return
		}
		h.logger.Error("Failed to get tenant by domain", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tenant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tenant})
}

// UpdateTenant handles updating a tenant.
func (h *TenantAdminHandler) UpdateTenant(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Error("Invalid tenant ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// Get existing tenant
	tenant, err := h.service.GetTenant(uint(id))
	if err != nil {
		if err.Error() == "tenant not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			return
		}
		h.logger.Error("Failed to get tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tenant"})
		return
	}

	// Bind updates
	if err := c.ShouldBindJSON(tenant); err != nil {
		h.logger.Error("Failed to bind tenant data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant data"})
		return
	}

	if err := h.service.UpdateTenant(tenant); err != nil {
		h.logger.Error("Failed to update tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tenant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant updated successfully",
		"data":    tenant,
	})
}

// DeleteTenant handles deleting a tenant.
func (h *TenantAdminHandler) DeleteTenant(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Error("Invalid tenant ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	if err := h.service.DeleteTenant(uint(id)); err != nil {
		h.logger.Error("Failed to delete tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tenant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tenant deleted successfully"})
}

// GetTenantBranding handles getting tenant branding information.
func (h *TenantAdminHandler) GetTenantBranding(c *gin.Context) {
	tenantID, err := h.getTenantID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	branding, err := h.service.GetTenantBranding(tenantID)
	if err != nil {
		if err.Error() == "tenant branding not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant branding not found"})
			return
		}
		h.logger.Error("Failed to get tenant branding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tenant branding"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": branding})
}

// UpdateTenantBranding handles updating tenant branding information.
func (h *TenantAdminHandler) UpdateTenantBranding(c *gin.Context) {
	tenantID, err := h.getTenantID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// Get existing branding
	branding, err := h.service.GetTenantBranding(tenantID)
	if err != nil {
		if err.Error() == "tenant branding not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant branding not found"})
			return
		}
		h.logger.Error("Failed to get tenant branding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tenant branding"})
		return
	}

	// Bind updates
	if err := c.ShouldBindJSON(branding); err != nil {
		h.logger.Error("Failed to bind branding data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branding data"})
		return
	}

	if err := h.service.UpdateTenantBranding(branding); err != nil {
		h.logger.Error("Failed to update tenant branding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tenant branding"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant branding updated successfully",
		"data":    branding,
	})
}
