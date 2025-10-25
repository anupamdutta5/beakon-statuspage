package handlers

import (
	"net/http"
	"strconv"

	"github.com/anupamdutta5/incident-service/internal/features/alerts/core"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AlertsHandler handles HTTP requests for alerts functionality
type AlertsHandler struct {
	alertService *core.AlertService
	logger       *zap.Logger
}

// NewAlertsHandler creates a new alerts handler
func NewAlertsHandler(alertService *core.AlertService, logger *zap.Logger) *AlertsHandler {
	return &AlertsHandler{
		alertService: alertService,
		logger:       logger,
	}
}

// CreateAlert godoc
// @Summary Create a new alert
// @Description Create a new alert configuration
// @Tags alerts
// @Accept json
// @Produce json
// @Param alert body core.Alert true "Alert configuration"
// @Success 201 {object} core.Alert
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/alerts [post]
func (h *AlertsHandler) CreateAlert(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	if tenantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id"})
		return
	}

	var alert core.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		h.logger.Error("Failed to bind alert JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alert.TenantID = uint(tenantID)

	if err := h.alertService.CreateAlert(c.Request.Context(), &alert); err != nil {
		h.logger.Error("Failed to create alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create alert"})
		return
	}

	c.JSON(http.StatusCreated, alert)
}

// GetAlert godoc
// @Summary Get alert by ID
// @Description Get a specific alert by its ID
// @Tags alerts
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} core.Alert
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/alerts/{id} [get]
func (h *AlertsHandler) GetAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	alert, err := h.alertService.GetAlert(uint(id))
	if err != nil {
		h.logger.Error("Failed to get alert", zap.Error(err), zap.Uint("id", uint(id)))
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// ListAlerts godoc
// @Summary List all alerts
// @Description Get all alerts for the tenant
// @Tags alerts
// @Produce json
// @Success 200 {array} core.Alert
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/alerts [get]
func (h *AlertsHandler) ListAlerts(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id"})
		return
	}

	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	alerts, total, err := h.alertService.GetAlerts(uint(tenantID), status, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list alerts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// DeleteAlert godoc
// @Summary Delete alert
// @Description Delete an alert
// @Tags alerts
// @Param id path int true "Alert ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/alerts/{id} [delete]
func (h *AlertsHandler) DeleteAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	if err := h.alertService.DeleteAlert(uint(id)); err != nil {
		h.logger.Error("Failed to delete alert", zap.Error(err), zap.Uint("id", uint(id)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete alert"})
		return
	}

	c.Status(http.StatusNoContent)
}

// AcknowledgeAlert godoc
// @Summary Acknowledge an alert
// @Description Acknowledge an alert
// @Tags alerts
// @Param id path int true "Alert ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/alerts/{id}/acknowledge [post]
func (h *AlertsHandler) AcknowledgeAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	userID := uint(1) // TODO: Get from authenticated user context

	if err := h.alertService.AcknowledgeAlert(uint(id), userID); err != nil {
		h.logger.Error("Failed to acknowledge alert", zap.Error(err), zap.Uint("id", uint(id)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to acknowledge alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert acknowledged successfully"})
}
