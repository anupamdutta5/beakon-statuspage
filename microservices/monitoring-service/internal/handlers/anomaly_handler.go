// Package handlers provides HTTP handlers for anomaly detection.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/anupamdutta5/monitoring-service/internal/models"
	"github.com/anupamdutta5/monitoring-service/internal/services"
	"github.com/gin-gonic/gin"
)

// AnomalyHandler handles HTTP requests for anomaly detection operations.
type AnomalyHandler struct {
	anomalyService *services.AnomalyDetectionService
}

// NewAnomalyHandler creates a new AnomalyHandler instance.
func NewAnomalyHandler(anomalyService *services.AnomalyDetectionService) *AnomalyHandler {
	return &AnomalyHandler{
		anomalyService: anomalyService,
	}
}

// GetAnomalies retrieves anomalies with optional filters.
// GET /api/v1/anomalies
// Query params: monitor_id, status, severity, from, to, limit
func (h *AnomalyHandler) GetAnomalies(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// Parse filters
	filters := make(map[string]interface{})

	if monitorIDStr := c.Query("monitor_id"); monitorIDStr != "" {
		monitorID, err := strconv.ParseUint(monitorIDStr, 10, 32)
		if err == nil {
			filters["monitor_id"] = uint(monitorID)
		}
	}

	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	if severity := c.Query("severity"); severity != "" {
		filters["severity"] = severity
	}

	if from := c.Query("from"); from != "" {
		if fromTime, err := time.Parse(time.RFC3339, from); err == nil {
			filters["from"] = fromTime
		}
	}

	if to := c.Query("to"); to != "" {
		if toTime, err := time.Parse(time.RFC3339, to); err == nil {
			filters["to"] = toTime
		}
	}

	// Parse limit
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get anomalies
	anomalies, err := h.anomalyService.GetAnomalies(uint(tenantID), filters, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch anomalies",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"anomalies": anomalies,
		"total":     len(anomalies),
	})
}

// GetAnomalyByID retrieves a specific anomaly with context.
// GET /api/v1/anomalies/:id
func (h *AnomalyHandler) GetAnomalyByID(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	anomalyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid anomaly ID"})
		return
	}

	context, err := h.anomalyService.GetAnomalyByID(uint(tenantID), uint(anomalyID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Anomaly not found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, context)
}

// AcknowledgeAnomaly marks an anomaly as acknowledged.
// POST /api/v1/anomalies/:id/acknowledge
// Request body: { "notes": "string" }
func (h *AnomalyHandler) AcknowledgeAnomaly(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	userIDStr := c.GetString("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	anomalyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid anomaly ID"})
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	if err := h.anomalyService.AcknowledgeAnomaly(uint(tenantID), uint(anomalyID), uint(userID), req.Notes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to acknowledge anomaly",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Anomaly acknowledged successfully",
		"acknowledged_at": time.Now(),
	})
}

// ResolveAnomaly marks an anomaly as resolved or false positive.
// POST /api/v1/anomalies/:id/resolve
// Request body: { "resolution": "resolved" | "false_positive", "notes": "string" }
func (h *AnomalyHandler) ResolveAnomaly(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	anomalyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid anomaly ID"})
		return
	}

	var req struct {
		Resolution string `json:"resolution" binding:"required,oneof=resolved false_positive"`
		Notes      string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	if err := h.anomalyService.ResolveAnomaly(uint(tenantID), uint(anomalyID), req.Resolution, req.Notes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to resolve anomaly",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Anomaly resolved successfully",
		"resolved_at": time.Now(),
	})
}

// GetBaselines retrieves baselines for a monitor.
// GET /api/v1/anomalies/baselines/:monitor_id
func (h *AnomalyHandler) GetBaselines(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	monitorID, err := strconv.ParseUint(c.Param("monitor_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID"})
		return
	}

	// Get all baselines for the monitor
	var baselines []models.AnomalyBaseline
	if err := h.anomalyService.GetDB().Where("tenant_id = ? AND monitor_id = ?", tenantID, uint(monitorID)).
		Find(&baselines).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch baselines",
			"message": err.Error(),
		})
		return
	}

	// Get monitor name
	var monitor models.Monitor
	if err := h.anomalyService.GetDB().Where("id = ?", uint(monitorID)).First(&monitor).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Monitor not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"monitor_id":   monitor.ID,
		"monitor_name": monitor.Name,
		"baselines":    baselines,
	})
}

// GetStatistics retrieves aggregated anomaly statistics.
// GET /api/v1/anomalies/stats
// Query param: period (24h, 7d, 30d)
func (h *AnomalyHandler) GetStatistics(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	period := c.DefaultQuery("period", "7d")

	stats, err := h.anomalyService.GetStatistics(uint(tenantID), period)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to fetch statistics",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetConfig retrieves anomaly detection configuration.
// GET /api/v1/anomalies/config
// Query params: monitor_id (optional), metric_type (optional)
func (h *AnomalyHandler) GetConfig(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var monitorID uint
	if monitorIDStr := c.Query("monitor_id"); monitorIDStr != "" {
		mid, err := strconv.ParseUint(monitorIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID"})
			return
		}
		monitorID = uint(mid)
	}

	metricType := c.Query("metric_type")

	// Get config
	var config models.AnomalyDetectionConfig
	query := h.anomalyService.GetDB().Where("tenant_id = ?", tenantID)

	if monitorID > 0 {
		query = query.Where("monitor_id = ?", monitorID)
	} else {
		query = query.Where("monitor_id IS NULL")
	}

	if metricType != "" {
		query = query.Where("metric_type = ?", metricType)
	} else {
		query = query.Where("metric_type IS NULL")
	}

	if err := query.First(&config).Error; err != nil {
		// Return default config if not found
		minor, major, critical := models.GetThresholdsForSensitivity(models.SensitivityMedium)
		config = models.AnomalyDetectionConfig{
			TenantID:                    uint(tenantID),
			Enabled:                     true,
			Sensitivity:                 models.SensitivityMedium,
			MinBaselineSamples:          50,
			ZScoreThresholdMinor:        minor,
			ZScoreThresholdMajor:        major,
			ZScoreThresholdCritical:     critical,
			NotificationEnabled:         true,
			NotificationCooldownMinutes: 30,
			RequireConsecutiveAnomalies: 1,
		}
	}

	c.JSON(http.StatusOK, config)
}

// UpdateConfig updates anomaly detection configuration.
// PUT /api/v1/anomalies/config
// Request body: AnomalyDetectionConfig
func (h *AnomalyHandler) UpdateConfig(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var config models.AnomalyDetectionConfig

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	config.TenantID = uint(tenantID)

	if err := h.anomalyService.UpdateConfig(uint(tenantID), &config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to update configuration",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration updated successfully",
		"config":  config,
	})
}
