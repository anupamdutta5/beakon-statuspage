// Package handlers provides HTTP handlers for SLA reporting and analytics.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/anupamdutta5/statuspage-analytics-service/internal/models"
	"github.com/anupamdutta5/statuspage-analytics-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SLAHandler handles SLA-related HTTP requests.
type SLAHandler struct {
	slaService *services.SLAService
	logger     *zap.Logger
}

// NewSLAHandler creates a new SLA handler.
func NewSLAHandler(slaService *services.SLAService, logger *zap.Logger) *SLAHandler {
	return &SLAHandler{
		slaService: slaService,
		logger:     logger,
	}
}

// CreateSLA handles creating a new SLA.
func (h *SLAHandler) CreateSLA(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		ComponentID    *uint   `json:"component_id"`
		Name           string  `json:"name" binding:"required"`
		Description    string  `json:"description"`
		Type           string  `json:"type" binding:"required"`
		TargetValue    float64 `json:"target_value" binding:"required"`
		Unit           string  `json:"unit" binding:"required"`
		PeriodType     string  `json:"period_type" binding:"required"`
		AlertThreshold float64 `json:"alert_threshold"`
		Settings       string  `json:"settings"`
		Metadata       string  `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create SLA request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	sla := &models.SLA{
		TenantID:       tenantID.(uint),
		ComponentID:    req.ComponentID,
		Name:           req.Name,
		Description:    req.Description,
		Type:           req.Type,
		TargetValue:    req.TargetValue,
		Unit:           req.Unit,
		PeriodType:     req.PeriodType,
		IsActive:       true,
		AlertThreshold: req.AlertThreshold,
		Settings:       req.Settings,
		Metadata:       req.Metadata,
	}

	if err := h.slaService.CreateSLA(sla); err != nil {
		h.logger.Error("Failed to create SLA", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create SLA"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "SLA created successfully",
		"sla":     sla,
	})
}

// GetSLAs handles retrieving SLAs for a tenant.
func (h *SLAHandler) GetSLAs(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	// Query parameters would be used for filtering
	_ = c.Query("component_id")
	_ = c.Query("type")
	_ = c.Query("is_active")

	slas, total, err := h.slaService.GetSLAs(tenantID.(uint), 100, 0)
	if err != nil {
		h.logger.Error("Failed to get SLAs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve SLAs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"slas":      slas,
		"tenant_id": tenantID,
		"count":     len(slas),
		"total":     total,
	})
}

// GetSLA handles retrieving a specific SLA by ID.
func (h *SLAHandler) GetSLA(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	slaIDStr := c.Param("id")
	slaID, err := strconv.ParseUint(slaIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SLA ID"})
		return
	}

	sla, err := h.slaService.GetSLA(uint(slaID))
	if err != nil {
		h.logger.Error("Failed to get SLA", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "SLA not found"})
		return
	}

	if sla.TenantID != tenantID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sla": sla,
	})
}

// UpdateSLA handles updating an SLA.
func (h *SLAHandler) UpdateSLA(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "SLA update not yet implemented"})
}

// DeleteSLA handles deleting an SLA.
func (h *SLAHandler) DeleteSLA(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "SLA delete not yet implemented"})
}

// CalculateSLAMeasurement handles calculating SLA measurements.
func (h *SLAHandler) CalculateSLAMeasurement(c *gin.Context) {
	slaIDStr := c.Param("id")
	slaID, err := strconv.ParseUint(slaIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SLA ID"})
		return
	}

	var req struct {
		PeriodStart time.Time `json:"period_start" binding:"required"`
		PeriodEnd   time.Time `json:"period_end" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid calculate SLA measurement request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	measurement, err := h.slaService.CalculateSLAMeasurement(uint(slaID), req.PeriodStart, req.PeriodEnd)
	if err != nil {
		h.logger.Error("Failed to calculate SLA measurement", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate SLA measurement"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "SLA measurement calculated successfully",
		"measurement": measurement,
	})
}

// GetSLAMeasurements handles retrieving SLA measurements.
func (h *SLAHandler) GetSLAMeasurements(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "SLA measurements retrieval not yet implemented"})
}

// GetSLABreaches handles retrieving SLA breaches.
func (h *SLAHandler) GetSLABreaches(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	// Query parameters would be used for filtering
	_ = c.Query("sla_id")
	_ = c.Query("severity")
	_ = c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	breaches, total, err := h.slaService.GetSLABreaches(tenantID.(uint), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get SLA breaches", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve SLA breaches"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"breaches": breaches,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// GenerateSLAReport handles generating SLA reports.
func (h *SLAHandler) GenerateSLAReport(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		SLAIDs      []uint    `json:"sla_ids"`
		Name        string    `json:"name" binding:"required"`
		Type        string    `json:"type" binding:"required"`
		Period      string    `json:"period" binding:"required"`
		PeriodStart time.Time `json:"period_start" binding:"required"`
		PeriodEnd   time.Time `json:"period_end" binding:"required"`
		Format      string    `json:"format"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid generate SLA report request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if req.Format == "" {
		req.Format = "pdf"
	}

	var slaID *uint
	if len(req.SLAIDs) > 0 {
		slaID = &req.SLAIDs[0]
	}
	report, err := h.slaService.GenerateSLAReport(tenantID.(uint), slaID, req.Type, req.Period, req.PeriodStart, req.PeriodEnd)
	if err != nil {
		h.logger.Error("Failed to generate SLA report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate SLA report"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "SLA report generation started",
		"report":  report,
	})
}

// GetSLAReports handles retrieving SLA reports.
func (h *SLAHandler) GetSLAReports(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "SLA reports retrieval not yet implemented"})
}

// GetSLAStatistics handles retrieving SLA statistics.
func (h *SLAHandler) GetSLAStatistics(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		SLAIDs      []uint    `json:"sla_ids"`
		PeriodStart time.Time `json:"period_start" binding:"required"`
		PeriodEnd   time.Time `json:"period_end" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid SLA statistics request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	stats, err := h.slaService.GetSLAStatistics(tenantID.(uint), req.PeriodStart, req.PeriodEnd)
	if err != nil {
		h.logger.Error("Failed to get SLA statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve SLA statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statistics": stats,
		"period": gin.H{
			"start": req.PeriodStart,
			"end":   req.PeriodEnd,
		},
	})
}

// CreateSLATarget handles creating an SLA target.
func (h *SLAHandler) CreateSLATarget(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "SLA target creation not yet implemented"})
}

// GetSLATargets handles retrieving SLA targets.
func (h *SLAHandler) GetSLATargets(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "SLA targets retrieval not yet implemented"})
}

// CalculateUptime handles calculating uptime for components.
func (h *SLAHandler) CalculateUptime(c *gin.Context) {
	_, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		ComponentID     uint      `json:"component_id" binding:"required"`
		PeriodStart     time.Time `json:"period_start" binding:"required"`
		PeriodEnd       time.Time `json:"period_end" binding:"required"`
		CalculationType string    `json:"calculation_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid calculate uptime request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if req.CalculationType == "" {
		req.CalculationType = "daily"
	}

	calculation, err := h.slaService.CalculateUptimeSLA(req.ComponentID, req.PeriodStart, req.PeriodEnd)
	if err != nil {
		h.logger.Error("Failed to calculate uptime", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate uptime"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Uptime calculated successfully",
		"calculation": calculation,
	})
}

// RecordResponseTime handles recording response time metrics.
func (h *SLAHandler) RecordResponseTime(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Response time recording not yet implemented"})
}