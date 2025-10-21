// Package handlers provides HTTP handlers for on-call schedule endpoints.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/anupamdutta5/monitoring-service/internal/models"
	"github.com/anupamdutta5/monitoring-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// OnCallHandler handles on-call schedule HTTP requests.
type OnCallHandler struct {
	service *services.OnCallService
	logger  *zap.Logger
}

// NewOnCallHandler creates a new on-call handler.
func NewOnCallHandler(service *services.OnCallService, logger *zap.Logger) *OnCallHandler {
	return &OnCallHandler{
		service: service,
		logger:  logger,
	}
}

// CreateSchedule creates a new on-call schedule.
// POST /api/v1/oncall/schedules
func (h *OnCallHandler) CreateSchedule(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var schedule models.OnCallSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateSchedule(tenantID, &schedule); err != nil {
		h.logger.Error("Failed to create on-call schedule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create schedule"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":   "success",
		"schedule": schedule,
	})
}

// GetSchedules retrieves all on-call schedules for a tenant.
// GET /api/v1/oncall/schedules
func (h *OnCallHandler) GetSchedules(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	includeInactive := c.DefaultQuery("include_inactive", "false") == "true"
	schedules, err := h.service.GetSchedules(tenantID, includeInactive)
	if err != nil {
		h.logger.Error("Failed to fetch on-call schedules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch schedules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"count":     len(schedules),
		"schedules": schedules,
	})
}

// GetSchedule retrieves a specific on-call schedule.
// GET /api/v1/oncall/schedules/:id
func (h *OnCallHandler) GetSchedule(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	scheduleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}

	schedule, err := h.service.GetSchedule(uint(scheduleID), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"schedule": schedule,
	})
}

// UpdateSchedule updates an on-call schedule.
// PUT /api/v1/oncall/schedules/:id
func (h *OnCallHandler) UpdateSchedule(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	scheduleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateSchedule(uint(scheduleID), tenantID, updates); err != nil {
		h.logger.Error("Failed to update schedule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update schedule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Schedule updated successfully",
	})
}

// DeleteSchedule deletes an on-call schedule.
// DELETE /api/v1/oncall/schedules/:id
func (h *OnCallHandler) DeleteSchedule(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	scheduleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}

	if err := h.service.DeleteSchedule(uint(scheduleID), tenantID); err != nil {
		h.logger.Error("Failed to delete schedule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete schedule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Schedule deleted successfully",
	})
}

// GetCurrentOnCall retrieves the current on-call person for a schedule.
// GET /api/v1/oncall/schedules/:id/current
func (h *OnCallHandler) GetCurrentOnCall(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	scheduleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}

	currentOnCall, err := h.service.GetCurrentOnCall(uint(scheduleID), tenantID)
	if err != nil {
		h.logger.Error("Failed to get current on-call", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current on-call"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"current_oncall": currentOnCall,
	})
}

// AddParticipant adds a participant to an on-call schedule.
// POST /api/v1/oncall/schedules/:id/participants
func (h *OnCallHandler) AddParticipant(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	scheduleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}

	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Email       string `json:"email" binding:"required"`
		PhoneNumber string `json:"phone_number"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse UserID to UUID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	participant := services.OnCallParticipant{
		UserID:      userID,
		Name:        req.Name,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
	}

	if err := h.service.AddParticipant(uint(scheduleID), tenantID, participant); err != nil {
		h.logger.Error("Failed to add participant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add participant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Participant added successfully",
	})
}

// RemoveParticipant removes a participant from an on-call schedule.
// DELETE /api/v1/oncall/schedules/:id/participants/:user_id
func (h *OnCallHandler) RemoveParticipant(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	scheduleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}

	userIDStr := c.Param("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	if err := h.service.RemoveParticipant(uint(scheduleID), tenantID, userID); err != nil {
		h.logger.Error("Failed to remove participant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove participant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Participant removed successfully",
	})
}
