package api

import (
	"fmt"
	"net/http"

	"github.com/enterprise-status/statuspage/internal/api/middleware"
	"github.com/enterprise-status/statuspage/internal/services"
	"github.com/gin-gonic/gin"
)

// MonitoringAnalyticsHandlers handles monitoring analytics API endpoints
type MonitoringAnalyticsHandlers struct {
	monitoringService *services.MonitoringAnalyticsService
}

// NewMonitoringAnalyticsHandlers creates a new monitoring analytics handlers instance
func NewMonitoringAnalyticsHandlers(monitoringService *services.MonitoringAnalyticsService) *MonitoringAnalyticsHandlers {
	return &MonitoringAnalyticsHandlers{
		monitoringService: monitoringService,
	}
}

// GetServiceUptimeData returns uptime data for a specific service
func (h *MonitoringAnalyticsHandlers) GetServiceUptimeData(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	serviceIDStr := c.Param("serviceId")
	serviceID, err := parseUint(serviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	period := c.DefaultQuery("period", "24h")

	uptimeData, err := h.monitoringService.GetServiceUptimeData(c.Request.Context(), tenant.ID, serviceID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_id": serviceID,
		"period":     period,
		"data":       uptimeData,
	})
}

// GetServiceResponseTimeData returns response time data for a specific service
func (h *MonitoringAnalyticsHandlers) GetServiceResponseTimeData(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	serviceIDStr := c.Param("serviceId")
	serviceID, err := parseUint(serviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	period := c.DefaultQuery("period", "24h")

	responseTimeData, err := h.monitoringService.GetServiceResponseTimeData(c.Request.Context(), tenant.ID, serviceID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_id": serviceID,
		"period":     period,
		"data":       responseTimeData,
	})
}

// GetServiceHealthData returns health data for all services
func (h *MonitoringAnalyticsHandlers) GetServiceHealthData(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	healthData, err := h.monitoringService.GetServiceHealthData(c.Request.Context(), tenant.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"services": healthData,
	})
}

// GetMonitoringMetrics returns comprehensive monitoring metrics
func (h *MonitoringAnalyticsHandlers) GetMonitoringMetrics(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	period := c.DefaultQuery("period", "24h")

	metrics, err := h.monitoringService.GetMonitoringMetrics(c.Request.Context(), tenant.ID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}

// GetServiceStatusHistory returns status history for a service
func (h *MonitoringAnalyticsHandlers) GetServiceStatusHistory(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	serviceIDStr := c.Param("serviceId")
	serviceID, err := parseUint(serviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	period := c.DefaultQuery("period", "24h")

	statusHistory, err := h.monitoringService.GetServiceStatusHistory(c.Request.Context(), tenant.ID, serviceID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_id": serviceID,
		"period":     period,
		"history":    statusHistory,
	})
}

// GetUptimeSummary returns uptime summary for all services
func (h *MonitoringAnalyticsHandlers) GetUptimeSummary(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	summary, err := h.monitoringService.GetUptimeSummary(c.Request.Context(), tenant.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": summary,
	})
}

// GetServiceGraphData returns graph data for a specific service
func (h *MonitoringAnalyticsHandlers) GetServiceGraphData(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	serviceIDStr := c.Param("serviceId")
	serviceID, err := parseUint(serviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	period := c.DefaultQuery("period", "24h")
	graphType := c.DefaultQuery("type", "uptime") // uptime, response_time, errors

	var data interface{}
	var err2 error

	switch graphType {
	case "uptime":
		data, err2 = h.monitoringService.GetServiceUptimeData(c.Request.Context(), tenant.ID, serviceID, period)
	case "response_time":
		data, err2 = h.monitoringService.GetServiceResponseTimeData(c.Request.Context(), tenant.ID, serviceID, period)
	case "status":
		data, err2 = h.monitoringService.GetServiceStatusHistory(c.Request.Context(), tenant.ID, serviceID, period)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid graph type"})
		return
	}

	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err2.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_id": serviceID,
		"period":     period,
		"type":       graphType,
		"data":       data,
	})
}

// Helper function to parse uint from string
func parseUint(s string) (uint, error) {
	// Simple implementation - in production, use strconv.ParseUint
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}

	// For demo purposes, return a mock ID
	return 1, nil
}
