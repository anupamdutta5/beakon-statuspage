// Package handlers provides HTTP handlers for the Status Page UI Service.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/enterprise-status/statuspage-status-ui-service/internal/services"
	"go.uber.org/zap"
)

// StatusPageHandler handles status page HTTP requests.
type StatusPageHandler struct {
	statusPageService *services.StatusPageService
	logger            *zap.Logger
}

// NewStatusPageHandler creates a new status page handler.
func NewStatusPageHandler(statusPageService *services.StatusPageService, logger *zap.Logger) *StatusPageHandler {
	return &StatusPageHandler{
		statusPageService: statusPageService,
		logger:            logger,
	}
}

// GetStatusPage renders the main status page.
func (h *StatusPageHandler) GetStatusPage(c *gin.Context) {
	// Get tenant ID from query parameter or header
	tenantID, err := h.getTenantID(c)
	if err != nil {
		h.logger.Error("Failed to get tenant ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// Get status page data
	data, err := h.statusPageService.GetStatusPageData(tenantID, "default")
	if err != nil {
		h.logger.Error("Failed to get status page data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load status page"})
		return
	}

	// Render HTML template
	c.HTML(http.StatusOK, "status.html", data)
}

// GetStatusPageBySlug renders a status page by slug.
func (h *StatusPageHandler) GetStatusPageBySlug(c *gin.Context) {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load status page"})
		return
	}

	// Render HTML template
	c.HTML(http.StatusOK, "status.html", data)
}

// GetStatusData returns status page data as JSON.
func (h *StatusPageHandler) GetStatusData(c *gin.Context) {
	// Get tenant ID from query parameter or header
	tenantID, err := h.getTenantID(c)
	if err != nil {
		h.logger.Error("Failed to get tenant ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// Get status page data
	data, err := h.statusPageService.GetStatusPageData(tenantID, "default")
	if err != nil {
		h.logger.Error("Failed to get status page data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load status data"})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetStatusDataBySlug returns status page data by slug as JSON.
func (h *StatusPageHandler) GetStatusDataBySlug(c *gin.Context) {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load status data"})
		return
	}

	c.JSON(http.StatusOK, data)
}

// getTenantID extracts tenant ID from request.
func (h *StatusPageHandler) getTenantID(c *gin.Context) (uint, error) {
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

