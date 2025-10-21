// Package handlers provides HTTP handlers for heartbeat monitoring endpoints.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/anupamdutta5/monitoring-service/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// HeartbeatHandler handles heartbeat monitoring HTTP requests.
type HeartbeatHandler struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewHeartbeatHandler creates a new heartbeat handler.
func NewHeartbeatHandler(db *gorm.DB, logger *zap.Logger) *HeartbeatHandler {
	return &HeartbeatHandler{
		db:     db,
		logger: logger,
	}
}

// CreateHeartbeat creates a new heartbeat monitor.
// POST /api/v1/heartbeat
func (h *HeartbeatHandler) CreateHeartbeat(c *gin.Context) {
	// Get tenant_id from context (set by auth middleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	var req struct {
		Name                    string `json:"name" binding:"required"`
		Description             string `json:"description"`
		ExpectedIntervalSeconds int    `json:"expected_interval_seconds" binding:"required,min=60"`
		GracePeriodSeconds      int    `json:"grace_period_seconds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate unique key for ping URL
	uniqueKey := uuid.New().String()

	heartbeat := models.HeartbeatMonitor{
		TenantID:                tenantID.(uuid.UUID),
		Name:                    req.Name,
		Description:             req.Description,
		UniqueKey:               uniqueKey,
		ExpectedIntervalSeconds: req.ExpectedIntervalSeconds,
		GracePeriodSeconds:      req.GracePeriodSeconds,
		IsAlive:                 false,
	}

	if heartbeat.GracePeriodSeconds == 0 {
		heartbeat.GracePeriodSeconds = 300 // Default 5 minutes
	}

	if err := h.db.Create(&heartbeat).Error; err != nil {
		h.logger.Error("Failed to create heartbeat monitor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create heartbeat monitor"})
		return
	}

	h.logger.Info("Heartbeat monitor created",
		zap.Uint("heartbeat_id", heartbeat.ID),
		zap.String("unique_key", uniqueKey),
	)

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   heartbeat,
		"ping_url": "/api/v1/heartbeat/ping/" + uniqueKey,
	})
}

// GetHeartbeats retrieves all heartbeat monitors for a tenant.
// GET /api/v1/heartbeat
func (h *HeartbeatHandler) GetHeartbeats(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	var heartbeats []models.HeartbeatMonitor
	if err := h.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&heartbeats).Error; err != nil {
		h.logger.Error("Failed to get heartbeat monitors", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get heartbeat monitors"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(heartbeats),
		"data":   heartbeats,
	})
}

// GetHeartbeat retrieves a specific heartbeat monitor by ID.
// GET /api/v1/heartbeat/:id
func (h *HeartbeatHandler) GetHeartbeat(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid heartbeat ID"})
		return
	}

	var heartbeat models.HeartbeatMonitor
	if err := h.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&heartbeat).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Heartbeat monitor not found"})
			return
		}
		h.logger.Error("Failed to get heartbeat monitor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get heartbeat monitor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   heartbeat,
		"ping_url": "/api/v1/heartbeat/ping/" + heartbeat.UniqueKey,
		"is_overdue": heartbeat.IsOverdue(),
	})
}

// UpdateHeartbeat updates a heartbeat monitor.
// PUT /api/v1/heartbeat/:id
func (h *HeartbeatHandler) UpdateHeartbeat(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid heartbeat ID"})
		return
	}

	var req struct {
		Name                    string `json:"name"`
		Description             string `json:"description"`
		ExpectedIntervalSeconds int    `json:"expected_interval_seconds"`
		GracePeriodSeconds      int    `json:"grace_period_seconds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var heartbeat models.HeartbeatMonitor
	if err := h.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&heartbeat).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Heartbeat monitor not found"})
			return
		}
		h.logger.Error("Failed to get heartbeat monitor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get heartbeat monitor"})
		return
	}

	// Update fields
	if req.Name != "" {
		heartbeat.Name = req.Name
	}
	if req.Description != "" {
		heartbeat.Description = req.Description
	}
	if req.ExpectedIntervalSeconds > 0 {
		heartbeat.ExpectedIntervalSeconds = req.ExpectedIntervalSeconds
	}
	if req.GracePeriodSeconds > 0 {
		heartbeat.GracePeriodSeconds = req.GracePeriodSeconds
	}

	if err := h.db.Save(&heartbeat).Error; err != nil {
		h.logger.Error("Failed to update heartbeat monitor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update heartbeat monitor"})
		return
	}

	h.logger.Info("Heartbeat monitor updated", zap.Uint("heartbeat_id", heartbeat.ID))

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   heartbeat,
	})
}

// DeleteHeartbeat deletes a heartbeat monitor.
// DELETE /api/v1/heartbeat/:id
func (h *HeartbeatHandler) DeleteHeartbeat(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid heartbeat ID"})
		return
	}

	result := h.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.HeartbeatMonitor{})
	if result.Error != nil {
		h.logger.Error("Failed to delete heartbeat monitor", zap.Error(result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete heartbeat monitor"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Heartbeat monitor not found"})
		return
	}

	h.logger.Info("Heartbeat monitor deleted", zap.Uint("heartbeat_id", uint(id)))

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Heartbeat monitor deleted successfully",
	})
}

// Ping handles heartbeat ping requests (public endpoint, no auth required).
// GET /api/v1/heartbeat/ping/:unique_key
func (h *HeartbeatHandler) Ping(c *gin.Context) {
	uniqueKey := c.Param("unique_key")

	var heartbeat models.HeartbeatMonitor
	if err := h.db.Where("unique_key = ?", uniqueKey).First(&heartbeat).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Heartbeat monitor not found"})
			return
		}
		h.logger.Error("Failed to get heartbeat monitor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process ping"})
		return
	}

	// Update last ping time and status
	heartbeat.RecordPing()

	if err := h.db.Save(&heartbeat).Error; err != nil {
		h.logger.Error("Failed to update heartbeat monitor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record ping"})
		return
	}

	h.logger.Info("Heartbeat ping recorded",
		zap.Uint("heartbeat_id", heartbeat.ID),
		zap.String("name", heartbeat.Name),
	)

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"message":   "Ping recorded successfully",
		"heartbeat": heartbeat.Name,
		"is_alive":  heartbeat.IsAlive,
	})
}

// GetOverdueHeartbeats retrieves all overdue heartbeat monitors.
// GET /api/v1/heartbeat/overdue
func (h *HeartbeatHandler) GetOverdueHeartbeats(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	var heartbeats []models.HeartbeatMonitor
	if err := h.db.Where("tenant_id = ?", tenantID).Find(&heartbeats).Error; err != nil {
		h.logger.Error("Failed to get heartbeat monitors", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get heartbeat monitors"})
		return
	}

	// Filter overdue heartbeats
	var overdueHeartbeats []models.HeartbeatMonitor
	for _, hb := range heartbeats {
		if hb.IsOverdue() {
			overdueHeartbeats = append(overdueHeartbeats, hb)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(overdueHeartbeats),
		"data":   overdueHeartbeats,
	})
}

// GetHeartbeatStats retrieves statistics for heartbeat monitors.
// GET /api/v1/heartbeat/stats
func (h *HeartbeatHandler) GetHeartbeatStats(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	var total, alive, overdue int64

	h.db.Model(&models.HeartbeatMonitor{}).Where("tenant_id = ?", tenantID).Count(&total)
	h.db.Model(&models.HeartbeatMonitor{}).Where("tenant_id = ? AND is_alive = ?", tenantID, true).Count(&alive)

	// Count overdue (requires application-level logic)
	var heartbeats []models.HeartbeatMonitor
	h.db.Where("tenant_id = ?", tenantID).Find(&heartbeats)
	for _, hb := range heartbeats {
		if hb.IsOverdue() {
			overdue++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"stats": gin.H{
			"total":   total,
			"alive":   alive,
			"overdue": overdue,
		},
	})
}
