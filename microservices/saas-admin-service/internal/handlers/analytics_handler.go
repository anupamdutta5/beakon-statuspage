// Package handlers provides HTTP handlers for analytics and statistics.
package handlers

import (
	"net/http"

	"github.com/anupamdutta5/saas-admin-service/internal/clients"
	"github.com/anupamdutta5/saas-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AnalyticsHandler handles analytics and statistics HTTP requests.
type AnalyticsHandler struct {
	service *services.SaaSAdminService
	logger  *zap.Logger
}

// NewAnalyticsHandler creates a new analytics handler.
func NewAnalyticsHandler(service *services.SaaSAdminService, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		service: service,
		logger:  logger,
	}
}

// GetStats handles getting platform statistics.
func (h *AnalyticsHandler) GetStats(c *gin.Context) {
	h.logger.Info("Getting platform statistics")
	// TODO: implement GetStats method
	stats := map[string]interface{}{
		"placeholder": "GetStats method not yet implemented",
	}
	var err error
	if err != nil {
		h.logger.Error("Failed to get statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get platform statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// GetAnalyticsOverview handles getting analytics overview data.
func (h *AnalyticsHandler) GetAnalyticsOverview(c *gin.Context) {
	h.logger.Info("Getting analytics overview")

	// Get tenant ID from query parameter or default
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		tenantID = "tenant-1" // Default tenant
	}

	overview, err := h.service.GetAnalyticsOverview(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get analytics overview", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get analytics overview",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    overview,
	})
}

// GetAnalyticsMetrics handles getting analytics metrics data.
func (h *AnalyticsHandler) GetAnalyticsMetrics(c *gin.Context) {
	h.logger.Info("Getting analytics metrics")

	// Get parameters from query
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		tenantID = "tenant-1" // Default tenant
	}

	timeRange := c.Query("range")
	if timeRange == "" {
		timeRange = "24h" // Default range
	}

	metrics, err := h.service.GetAnalyticsMetrics(c.Request.Context(), tenantID, timeRange)
	if err != nil {
		h.logger.Error("Failed to get analytics metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get analytics metrics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// CreateAnalyticsMetric handles creating a new analytics metric.
func (h *AnalyticsHandler) CreateAnalyticsMetric(c *gin.Context) {
	h.logger.Info("Creating analytics metric")

	// Get tenant ID from query parameter or default
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		tenantID = "tenant-1" // Default tenant
	}

	var metric clients.MetricData
	if err := c.ShouldBindJSON(&metric); err != nil {
		h.logger.Error("Failed to bind metric data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid metric data",
		})
		return
	}

	createdMetric, err := h.service.CreateAnalyticsMetric(c.Request.Context(), tenantID, metric)
	if err != nil {
		h.logger.Error("Failed to create analytics metric", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create analytics metric",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    createdMetric,
	})
}

// CheckAnalyticsHealth handles checking the health of the analytics service.
func (h *AnalyticsHandler) CheckAnalyticsHealth(c *gin.Context) {
	h.logger.Info("Checking analytics service health")

	err := h.service.CheckAnalyticsHealth(c.Request.Context())
	if err != nil {
		h.logger.Error("Analytics service health check failed", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"healthy": false,
			"error":   "Analytics service is unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"healthy": true,
		"service": "analytics",
	})
}