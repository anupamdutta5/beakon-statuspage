// Package handlers provides HTTP handlers for the Tenant Admin Service.
package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/anupamdutta5/tenant-admin-service/internal/models"
	"github.com/anupamdutta5/tenant-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
	// Check if tenant context exists - only allow login for active tenants
	_, hasID := c.Get("tenant_id")
	_, hasName := c.Get("tenant_name")
	_, hasSlug := c.Get("tenant_slug")

	// Extract subdomain to check if user is trying to access a specific tenant
	host := c.Request.Host
	subdomain := extractSubdomainFromHost(host)

	// If accessing via subdomain but tenant doesn't exist, show error
	if subdomain != "" && (!hasID || !hasName || !hasSlug) {
		h.logger.Warn("Tenant not found - login access denied",
			zap.String("host", host),
			zap.String("subdomain", subdomain),
			zap.String("path", c.Request.URL.Path))

		c.HTML(http.StatusNotFound, "login.html", gin.H{
			"title": "Tenant Not Found",
			"error": "Tenant not found",
		})
		return
	}

	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Status Page Admin - Login",
	})
}

// extractSubdomainFromHost extracts subdomain from host (helper function)
func extractSubdomainFromHost(host string) string {
	// Remove port if present
	if colonIdx := len(host) - 1; colonIdx >= 0 {
		for i := len(host) - 1; i >= 0; i-- {
			if host[i] == ':' {
				host = host[:i]
				break
			}
		}
	}

	// Check if it's a subdomain of localhost
	baseDomain := "localhost"
	if host == baseDomain {
		return ""
	}

	// Check if host ends with .localhost
	if len(host) > len(baseDomain)+1 && host[len(host)-len(baseDomain)-1:] == "."+baseDomain {
		subdomain := host[:len(host)-len(baseDomain)-1]
		// Handle nested subdomains - only take the first part
		for i := 0; i < len(subdomain); i++ {
			if subdomain[i] == '.' {
				return subdomain[:i]
			}
		}
		return subdomain
	}

	return ""
}

// GetAdminDashboard renders the admin dashboard with tenant-specific information.
func (h *TenantAdminHandler) GetAdminDashboard(c *gin.Context) {
	// Extract tenant information from context (set by TenantContextMiddleware)
	tenantID, hasID := c.Get("tenant_id")
	tenantName, hasName := c.Get("tenant_name")
	tenantSlug, hasSlug := c.Get("tenant_slug")

	// Require valid tenant context - reject if tenant doesn't exist in database
	if !hasID || !hasName || !hasSlug {
		h.logger.Warn("Tenant not found - dashboard access denied",
			zap.String("host", c.Request.Host),
			zap.String("path", c.Request.URL.Path))

		c.HTML(http.StatusNotFound, "login.html", gin.H{
			"title": "Tenant Not Found",
			"error": "Tenant not found",
		})
		return
	}

	// Convert to strings
	tenantIDStr := ""
	if id, ok := tenantID.(string); ok {
		tenantIDStr = id
	}
	tenantNameStr := ""
	if name, ok := tenantName.(string); ok {
		tenantNameStr = name
	}
	tenantSlugStr := ""
	if slug, ok := tenantSlug.(string); ok {
		tenantSlugStr = slug
	}

	h.logger.Info("Rendering tenant admin dashboard",
		zap.String("tenant_id", tenantIDStr),
		zap.String("tenant_slug", tenantSlugStr),
		zap.String("tenant_name", tenantNameStr))

	// Generate tenant-specific data
	dashboardData := h.generateTenantDashboardData(tenantIDStr, tenantNameStr)

	c.HTML(http.StatusOK, "admin_dashboard.html", dashboardData)
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
	tenantID, err := uuid.Parse(tenantIDStr)
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

	h.logger.Info("Listing tenant admins", zap.String("tenant_id", tenantID.String()))

	admins, err := h.service.ListTenantAdmins(c.Request.Context(), tenantID, limit, offset)
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
	tenantID, err := uuid.Parse(tenantIDStr)
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

	admin.TenantID = tenantID

	h.logger.Info("Creating tenant admin", zap.String("tenant_id", tenantID.String()))

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
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	h.logger.Info("Getting tenant settings", zap.String("tenant_id", tenantID.String()))

	settings, err := h.service.GetTenantSettings(c.Request.Context(), tenantID)
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
	tenantID, err := uuid.Parse(tenantIDStr)
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

	h.logger.Info("Updating tenant settings", zap.String("tenant_id", tenantID.String()))

	if err := h.service.UpdateTenantSettings(c.Request.Context(), tenantID, &updates); err != nil {
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
	tenantID, err := uuid.Parse(tenantIDStr)
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

	h.logger.Info("Listing tenant feature flags", zap.String("tenant_id", tenantID.String()))

	flags, err := h.service.ListTenantFeatureFlags(c.Request.Context(), tenantID, limit, offset)
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
	tenantID, err := uuid.Parse(tenantIDStr)
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

	flag.TenantID = tenantID

	h.logger.Info("Creating tenant feature flag", zap.String("tenant_id", tenantID.String()))

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
	tenantID, err := uuid.Parse(tenantIDStr)
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

	h.logger.Info("Getting tenant usage", zap.String("tenant_id", tenantID.String()))

	usage, err := h.service.GetTenantUsage(c.Request.Context(), tenantID, startDate, endDate)
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
	tenantID, err := uuid.Parse(tenantIDStr)
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

	usage.TenantID = tenantID

	h.logger.Info("Recording tenant usage", zap.String("tenant_id", tenantID.String()))

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
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	h.logger.Info("Getting tenant statistics", zap.String("tenant_id", tenantID.String()))

	stats, err := h.service.GetTenantStats(c.Request.Context(), tenantID)
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

	// Get tenant context
	tenantID, hasTenantID := c.Get("tenant_id")
	if !hasTenantID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant context not found"})
		return
	}

	// Authenticate user with database
	user, err := h.service.AuthenticateUser(loginRequest.Email, loginRequest.Password)
	if err != nil {
		h.logger.Warn("Authentication failed",
			zap.String("email", loginRequest.Email),
			zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	tokenExpiry := time.Hour * 24 // Default 24 hours
	if loginRequest.RememberMe {
		tokenExpiry = time.Hour * 24 * 30 // 30 days if remember me
	}

	claims := jwt.MapClaims{
		"user_id":   user.ID,
		"email":     user.Email,
		"tenant_id": tenantID,
		"exp":       time.Now().Add(tokenExpiry).Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("your-jwt-secret-key")) // TODO: Use env variable
	if err != nil {
		h.logger.Error("Failed to generate JWT token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	h.logger.Info("User logged in successfully",
		zap.String("email", user.Email),
		zap.Uint("user_id", user.ID))

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"tenant_id": tenantID,
		},
	})
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
func (h *TenantAdminHandler) getTenantID(c *gin.Context) (uuid.UUID, error) {
	// Try to get from query parameter first
	if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return uuid.Nil, err
		}
		return tenantID, nil
	}

	// Try to get from header
	if tenantIDStr := c.GetHeader("X-Tenant-ID"); tenantIDStr != "" {
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return uuid.Nil, err
		}
		return tenantID, nil
	}

	// Return empty UUID for development (should be retrieved from tenant context middleware)
	return uuid.Nil, fmt.Errorf("tenant ID not found")
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
// CreateTenantRequest represents the request to create a tenant with admin credentials
type CreateTenantRequest struct {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant data: " + err.Error()})
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

	if err := h.service.CreateTenant(tenant); err != nil {
		h.logger.Error("Failed to create tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
		return
	}

	// Create admin user with hashed password
	if err := h.service.CreateAdminUser(tenant.ID, req.AdminEmail, req.AdminPassword); err != nil {
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

// generateTenantDashboardData creates tenant-specific dashboard data.
func (h *TenantAdminHandler) generateTenantDashboardData(tenantID, tenantName string) gin.H {
	// Base dashboard data
	dashboardData := gin.H{
		"title":           fmt.Sprintf("%s - Admin Dashboard", tenantName),
		"tenant_id":       tenantID,
		"tenant_name":     tenantName,
		"dashboard_title": fmt.Sprintf("%s Dashboard", tenantName),
		"last_updated":    time.Now().Format("2006-01-02 15:04:05"),
	}

	// Fetch real data from database
	ctx := context.Background()

	// Parse tenant ID to uint for database queries
	var tenantIDUint uint
	if id, err := strconv.ParseUint(tenantID, 10, 32); err == nil {
		tenantIDUint = uint(id)
	}

	// Count status pages for this tenant
	var statusPagesCount int64
	h.service.GetDB().Table("status_pages").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantIDUint).
		Count(&statusPagesCount)
	dashboardData["status_pages_count"] = statusPagesCount

	// Count components for this tenant
	var componentsCount int64
	h.service.GetDB().Table("components").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantIDUint).
		Count(&componentsCount)
	dashboardData["components_count"] = componentsCount

	// Count active incidents
	var activeIncidentsCount int64
	h.service.GetDB().Table("incidents").
		Where("tenant_id = ? AND status IN (?) AND deleted_at IS NULL",
			tenantIDUint,
			[]string{"investigating", "identified", "monitoring"}).
		Count(&activeIncidentsCount)
	dashboardData["active_incidents"] = activeIncidentsCount

	// Count subscribers
	var subscribersCount int64
	h.service.GetDB().Table("subscribers").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantIDUint).
		Count(&subscribersCount)
	dashboardData["subscribers_count"] = subscribersCount

	// Determine overall status based on active incidents
	if activeIncidentsCount > 0 {
		// Check severity of incidents
		var criticalCount int64
		h.service.GetDB().Table("incidents").
			Where("tenant_id = ? AND severity = ? AND status IN (?) AND deleted_at IS NULL",
				tenantIDUint,
				"critical",
				[]string{"investigating", "identified", "monitoring"}).
			Count(&criticalCount)

		if criticalCount > 0 {
			dashboardData["overall_status"] = "Major Service Disruption"
			dashboardData["overall_status_class"] = "critical"
			dashboardData["status_icon"] = "fas fa-exclamation-circle"
			dashboardData["status_color"] = "text-danger"
		} else {
			dashboardData["overall_status"] = "Minor Service Disruption"
			dashboardData["overall_status_class"] = "degraded"
			dashboardData["status_icon"] = "fas fa-exclamation-triangle"
			dashboardData["status_color"] = "text-warning"
		}
	} else {
		dashboardData["overall_status"] = "All Systems Operational"
		dashboardData["overall_status_class"] = "operational"
		dashboardData["status_icon"] = "fas fa-check-circle"
		dashboardData["status_color"] = "text-success"
	}

	h.logger.Debug("Generated dashboard data from database",
		zap.String("tenant_id", tenantID),
		zap.Int64("status_pages", statusPagesCount),
		zap.Int64("components", componentsCount),
		zap.Int64("incidents", activeIncidentsCount),
		zap.Int64("subscribers", subscribersCount))

	// Suppress unused variable warning
	_ = ctx

	return dashboardData
}
