package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/anupamdutta5/monitoring-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// PublicMetricsHandler handles public metrics API endpoints
type PublicMetricsHandler struct {
	performanceService   *services.PerformanceMetricsService
	multiLocationService *services.MultiLocationChecker
	logger               *zap.Logger
}

// NewPublicMetricsHandler creates a new public metrics handler
func NewPublicMetricsHandler(
	performanceService *services.PerformanceMetricsService,
	multiLocationService *services.MultiLocationChecker,
	logger *zap.Logger,
) *PublicMetricsHandler {
	return &PublicMetricsHandler{
		performanceService:   performanceService,
		multiLocationService: multiLocationService,
		logger:               logger,
	}
}

// GetMonitorMetrics returns performance metrics for a specific monitor
// GET /api/v1/public/monitors/:id/metrics?range=24h
func (h *PublicMetricsHandler) GetMonitorMetrics(c *gin.Context) {
	monitorIDStr := c.Param("id")
	monitorID, err := strconv.ParseUint(monitorIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid monitor ID",
		})
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	// Get time range from query param (default: 24h)
	timeRange := c.DefaultQuery("range", "24h")

	// Calculate metrics
	metrics, err := h.performanceService.CalculateMetrics(uint(monitorID), tenantID, timeRange)
	if err != nil {
		h.logger.Error("Failed to calculate metrics",
			zap.Error(err),
			zap.Uint64("monitor_id", monitorID))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to calculate metrics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"metrics": metrics,
	})
}

// GetMonitorMetricsSummary returns metrics summary across multiple time ranges
// GET /api/v1/public/monitors/:id/metrics/summary
func (h *PublicMetricsHandler) GetMonitorMetricsSummary(c *gin.Context) {
	monitorIDStr := c.Param("id")
	monitorID, err := strconv.ParseUint(monitorIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid monitor ID",
		})
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	// Get metrics summary
	summary, err := h.performanceService.GetMetricsSummary(uint(monitorID), tenantID)
	if err != nil {
		h.logger.Error("Failed to get metrics summary",
			zap.Error(err),
			zap.Uint64("monitor_id", monitorID))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get metrics summary",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"summary": summary,
	})
}

// GetMonitorTrend returns performance trend analysis
// GET /api/v1/public/monitors/:id/trend
func (h *PublicMetricsHandler) GetMonitorTrend(c *gin.Context) {
	monitorIDStr := c.Param("id")
	monitorID, err := strconv.ParseUint(monitorIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid monitor ID",
		})
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	// Get trend analysis
	trend, err := h.performanceService.GetPerformanceTrend(uint(monitorID), tenantID)
	if err != nil {
		h.logger.Error("Failed to get performance trend",
			zap.Error(err),
			zap.Uint64("monitor_id", monitorID))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get performance trend",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"trend":  trend,
	})
}

// GetUptimeStatus returns current uptime status for a monitor
// GET /api/v1/public/monitors/:id/uptime
func (h *PublicMetricsHandler) GetUptimeStatus(c *gin.Context) {
	monitorIDStr := c.Param("id")
	monitorID, err := strconv.ParseUint(monitorIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid monitor ID",
		})
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	// Get metrics for different time ranges
	ranges := []string{"24h", "7d", "30d", "90d"}
	uptime := make(map[string]interface{})

	for _, timeRange := range ranges {
		metrics, err := h.performanceService.CalculateMetrics(uint(monitorID), tenantID, timeRange)
		if err != nil {
			continue
		}

		uptime[timeRange] = map[string]interface{}{
			"uptime_percentage":  metrics.UptimePercentage,
			"total_checks":       metrics.TotalChecks,
			"successful_checks":  metrics.SuccessfulChecks,
			"failed_checks":      metrics.FailedChecks,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"monitor_id":   monitorID,
		"uptime_stats": uptime,
		"timestamp":    time.Now(),
	})
}

// GetResponseTimeStats returns response time statistics
// GET /api/v1/public/monitors/:id/response-time?range=24h
func (h *PublicMetricsHandler) GetResponseTimeStats(c *gin.Context) {
	monitorIDStr := c.Param("id")
	monitorID, err := strconv.ParseUint(monitorIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid monitor ID",
		})
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	timeRange := c.DefaultQuery("range", "24h")

	metrics, err := h.performanceService.CalculateMetrics(uint(monitorID), tenantID, timeRange)
	if err != nil {
		h.logger.Error("Failed to calculate response time stats",
			zap.Error(err),
			zap.Uint64("monitor_id", monitorID))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to calculate response time stats",
		})
		return
	}

	responseTimeStats := map[string]interface{}{
		"time_range":         timeRange,
		"avg_response_time":  metrics.AvgResponseTime,
		"min_response_time":  metrics.MinResponseTime,
		"max_response_time":  metrics.MaxResponseTime,
		"p50_response_time":  metrics.P50ResponseTime,
		"p90_response_time":  metrics.P90ResponseTime,
		"p95_response_time":  metrics.P95ResponseTime,
		"p99_response_time":  metrics.P99ResponseTime,
		"total_checks":       metrics.TotalChecks,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":             "success",
		"monitor_id":         monitorID,
		"response_time_stats": responseTimeStats,
		"timestamp":          time.Now(),
	})
}

// GetLocationMetrics returns metrics by location
// GET /api/v1/public/monitors/:id/locations
func (h *PublicMetricsHandler) GetLocationMetrics(c *gin.Context) {
	monitorIDStr := c.Param("id")
	monitorID, err := strconv.ParseUint(monitorIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid monitor ID",
		})
		return
	}

	// Get active locations
	locations, err := h.multiLocationService.GetActiveLocations()
	if err != nil {
		h.logger.Error("Failed to get locations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get locations",
		})
		return
	}

	// Get recent results for each location
	locationMetrics := make([]map[string]interface{}, 0)

	for _, location := range locations {
		results, err := h.multiLocationService.GetLocationResults(uint(monitorID), location.ID, 10)
		if err != nil || len(results) == 0 {
			continue
		}

		// Calculate metrics for this location
		aggregateMetrics := h.multiLocationService.GetAggregateMetrics(results)

		locationMetrics = append(locationMetrics, map[string]interface{}{
			"location_id":   location.ID,
			"location_name": location.Name,
			"city":          location.City,
			"country":       location.Country,
			"region":        location.Region,
			"metrics":       aggregateMetrics,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           "success",
		"monitor_id":       monitorID,
		"location_metrics": locationMetrics,
		"timestamp":        time.Now(),
	})
}

// GetAggregateMetrics returns aggregate metrics across multiple monitors
// POST /api/v1/public/metrics/aggregate
// Body: {"monitor_ids": [1, 2, 3], "time_range": "24h"}
func (h *PublicMetricsHandler) GetAggregateMetrics(c *gin.Context) {
	var request struct {
		MonitorIDs []uint `json:"monitor_ids" binding:"required"`
		TimeRange  string `json:"time_range"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	if request.TimeRange == "" {
		request.TimeRange = "24h"
	}

	// Calculate aggregated metrics
	metrics, err := h.performanceService.GetAggregatedMetrics(request.MonitorIDs, tenantID, request.TimeRange)
	if err != nil {
		h.logger.Error("Failed to calculate aggregated metrics",
			zap.Error(err),
			zap.Any("monitor_ids", request.MonitorIDs))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to calculate aggregated metrics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"metrics": metrics,
	})
}

// GetHistoricalUptime returns historical uptime data for charts
// GET /api/v1/public/monitors/:id/history?range=30d&interval=1d
func (h *PublicMetricsHandler) GetHistoricalUptime(c *gin.Context) {
	monitorIDStr := c.Param("id")
	monitorID, err := strconv.ParseUint(monitorIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid monitor ID",
		})
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	timeRange := c.DefaultQuery("range", "30d")

	// Get metrics with time-series data
	metrics, err := h.performanceService.CalculateMetrics(uint(monitorID), tenantID, timeRange)
	if err != nil {
		h.logger.Error("Failed to get historical uptime",
			zap.Error(err),
			zap.Uint64("monitor_id", monitorID))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get historical uptime",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":                 "success",
		"monitor_id":             monitorID,
		"time_range":             timeRange,
		"uptime_percentage":      metrics.UptimePercentage,
		"response_time_series":   metrics.ResponseTimeSeries,
		"uptime_series":          metrics.UptimeSeries,
		"timestamp":              time.Now(),
	})
}

// GetPublicStatusSummary returns public status summary for all monitors
// GET /api/v1/public/status
func (h *PublicMetricsHandler) GetPublicStatusSummary(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	// This would typically query all active monitors for the tenant
	// For now, return a placeholder structure

	summary := map[string]interface{}{
		"overall_status": "operational",
		"message":        "All systems operational",
		"last_updated":   time.Now(),
		"tenant_id":      tenantID,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"summary": summary,
	})
}

// GetPublicIncidents returns recent public incidents
// GET /api/v1/public/incidents?limit=10
func (h *PublicMetricsHandler) GetPublicIncidents(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	// Placeholder for incidents
	incidents := []map[string]interface{}{}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"incidents":  incidents,
		"tenant_id":  tenantID,
		"limit":      limit,
		"timestamp":  time.Now(),
	})
}
