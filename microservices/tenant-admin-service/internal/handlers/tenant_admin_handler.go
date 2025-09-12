// Package handlers provides HTTP handlers for the Tenant Admin Service.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/enterprise-status/statuspage-tenant-admin-service/internal/models"
	"github.com/enterprise-status/statuspage-tenant-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TenantAdminHandler handles tenant admin-related HTTP requests.
type TenantAdminHandler struct {
	service *services.TenantAdminService
	logger  *zap.Logger
}

// NewTenantAdminHandler creates a new tenant admin handler.
func NewTenantAdminHandler(service *services.TenantAdminService, logger *zap.Logger) *TenantAdminHandler {
	return &TenantAdminHandler{
		service: service,
		logger:  logger,
	}
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

