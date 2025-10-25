package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/anupamdutta5/monitoring-service/internal/services"
)

// PagerDutyHandler handles PagerDuty integration endpoints
type PagerDutyHandler struct {
	db                *gorm.DB
	logger            *zap.Logger
	pagerdutyService  *services.PagerDutyIntegrationService
	webhookSecret     string // Optional webhook signature verification
}

// NewPagerDutyHandler creates a new PagerDuty handler
func NewPagerDutyHandler(db *gorm.DB, logger *zap.Logger, pagerdutyService *services.PagerDutyIntegrationService) *PagerDutyHandler {
	return &PagerDutyHandler{
		db:               db,
		logger:           logger,
		pagerdutyService: pagerdutyService,
	}
}

// CreateIntegration creates a new PagerDuty integration
// POST /api/v1/integrations/pagerduty
func (h *PagerDutyHandler) CreateIntegration(c *gin.Context) {
	var req struct {
		TenantID            string `json:"tenant_id" binding:"required"`
		IntegrationName     string `json:"integration_name" binding:"required"`
		IntegrationKey      string `json:"integration_key" binding:"required"`
		APIKey              string `json:"api_key"`
		ServiceID           string `json:"service_id"`
		ServiceName         string `json:"service_name"`
		Severity            string `json:"severity"` // critical, error, warning, info
		AutoResolve         *bool  `json:"auto_resolve"`
		NotifyOnDown        *bool  `json:"notify_on_down"`
		NotifyOnDegraded    *bool  `json:"notify_on_degraded"`
		NotifyOnMaintenance *bool  `json:"notify_on_maintenance"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	// Set defaults
	autoResolve := true
	if req.AutoResolve != nil {
		autoResolve = *req.AutoResolve
	}

	notifyOnDown := true
	if req.NotifyOnDown != nil {
		notifyOnDown = *req.NotifyOnDown
	}

	notifyOnDegraded := true
	if req.NotifyOnDegraded != nil {
		notifyOnDegraded = *req.NotifyOnDegraded
	}

	notifyOnMaintenance := false
	if req.NotifyOnMaintenance != nil {
		notifyOnMaintenance = *req.NotifyOnMaintenance
	}

	severity := "error"
	if req.Severity != "" {
		severity = req.Severity
	}

	integration := &services.PagerDutyIntegration{
		TenantID:            tenantID,
		IntegrationName:     req.IntegrationName,
		IntegrationKey:      req.IntegrationKey,
		APIKey:              req.APIKey,
		ServiceID:           req.ServiceID,
		ServiceName:         req.ServiceName,
		IsActive:            true,
		AutoResolve:         autoResolve,
		Severity:            severity,
		NotifyOnDown:        notifyOnDown,
		NotifyOnDegraded:    notifyOnDegraded,
		NotifyOnMaintenance: notifyOnMaintenance,
	}

	if err := h.pagerdutyService.CreateIntegration(integration); err != nil {
		h.logger.Error("Failed to create PagerDuty integration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create integration"})
		return
	}

	h.logger.Info("PagerDuty integration created",
		zap.Uint("id", integration.ID),
		zap.String("tenant_id", tenantID.String()),
		zap.String("name", integration.IntegrationName),
	)

	c.JSON(http.StatusCreated, gin.H{"integration": integration})
}

// GetIntegrations returns all PagerDuty integrations for a tenant
// GET /api/v1/integrations/pagerduty
func (h *PagerDutyHandler) GetIntegrations(c *gin.Context) {
	tenantIDStr := c.Query("tenant_id")
	if tenantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	integrations, err := h.pagerdutyService.GetIntegrationsByTenant(tenantID)
	if err != nil {
		h.logger.Error("Failed to get PagerDuty integrations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve integrations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"integrations": integrations})
}

// GetIntegration returns a specific PagerDuty integration
// GET /api/v1/integrations/pagerduty/:id
func (h *PagerDutyHandler) GetIntegration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration ID"})
		return
	}

	tenantIDStr := c.Query("tenant_id")
	if tenantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	integration, err := h.pagerdutyService.GetIntegration(id, tenantID)
	if err != nil {
		h.logger.Error("Failed to get PagerDuty integration", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"integration": integration})
}

// UpdateIntegration updates a PagerDuty integration
// PUT /api/v1/integrations/pagerduty/:id
func (h *PagerDutyHandler) UpdateIntegration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration ID"})
		return
	}

	tenantIDStr := c.Query("tenant_id")
	if tenantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	// Get existing integration
	integration, err := h.pagerdutyService.GetIntegration(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	var req struct {
		IntegrationName     string `json:"integration_name"`
		IntegrationKey      string `json:"integration_key"`
		APIKey              string `json:"api_key"`
		ServiceID           string `json:"service_id"`
		ServiceName         string `json:"service_name"`
		IsActive            *bool  `json:"is_active"`
		AutoResolve         *bool  `json:"auto_resolve"`
		Severity            string `json:"severity"`
		NotifyOnDown        *bool  `json:"notify_on_down"`
		NotifyOnDegraded    *bool  `json:"notify_on_degraded"`
		NotifyOnMaintenance *bool  `json:"notify_on_maintenance"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	if req.IntegrationName != "" {
		integration.IntegrationName = req.IntegrationName
	}
	if req.IntegrationKey != "" {
		integration.IntegrationKey = req.IntegrationKey
	}
	if req.APIKey != "" {
		integration.APIKey = req.APIKey
	}
	if req.ServiceID != "" {
		integration.ServiceID = req.ServiceID
	}
	if req.ServiceName != "" {
		integration.ServiceName = req.ServiceName
	}
	if req.Severity != "" {
		integration.Severity = req.Severity
	}
	if req.IsActive != nil {
		integration.IsActive = *req.IsActive
	}
	if req.AutoResolve != nil {
		integration.AutoResolve = *req.AutoResolve
	}
	if req.NotifyOnDown != nil {
		integration.NotifyOnDown = *req.NotifyOnDown
	}
	if req.NotifyOnDegraded != nil {
		integration.NotifyOnDegraded = *req.NotifyOnDegraded
	}
	if req.NotifyOnMaintenance != nil {
		integration.NotifyOnMaintenance = *req.NotifyOnMaintenance
	}

	if err := h.pagerdutyService.UpdateIntegration(integration); err != nil {
		h.logger.Error("Failed to update PagerDuty integration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update integration"})
		return
	}

	h.logger.Info("PagerDuty integration updated",
		zap.Uint("id", integration.ID),
		zap.String("tenant_id", tenantID.String()),
	)

	c.JSON(http.StatusOK, gin.H{"integration": integration})
}

// DeleteIntegration deletes a PagerDuty integration
// DELETE /api/v1/integrations/pagerduty/:id
func (h *PagerDutyHandler) DeleteIntegration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration ID"})
		return
	}

	tenantIDStr := c.Query("tenant_id")
	if tenantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	if err := h.pagerdutyService.DeleteIntegration(id, tenantID); err != nil {
		h.logger.Error("Failed to delete PagerDuty integration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete integration"})
		return
	}

	h.logger.Info("PagerDuty integration deleted",
		zap.Uint("id", id),
		zap.String("tenant_id", tenantID.String()),
	)

	c.JSON(http.StatusOK, gin.H{"message": "integration deleted successfully"})
}

// TestIntegration sends a test event to PagerDuty
// POST /api/v1/integrations/pagerduty/:id/test
func (h *PagerDutyHandler) TestIntegration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration ID"})
		return
	}

	tenantIDStr := c.Query("tenant_id")
	if tenantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	// Get integration
	integration, err := h.pagerdutyService.GetIntegration(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	// Send test event
	if err := h.pagerdutyService.TestIntegration(integration.IntegrationKey); err != nil {
		h.logger.Error("Failed to send test event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send test event"})
		return
	}

	h.logger.Info("Test PagerDuty event sent",
		zap.Uint("id", id),
		zap.String("tenant_id", tenantID.String()),
	)

	c.JSON(http.StatusOK, gin.H{"message": "test event sent successfully"})
}

// MapMonitor maps a monitor to a PagerDuty integration
// POST /api/v1/integrations/pagerduty/:id/monitors
func (h *PagerDutyHandler) MapMonitor(c *gin.Context) {
	idStr := c.Param("id")
	integrationID, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration ID"})
		return
	}

	var req struct {
		MonitorID uint   `json:"monitor_id" binding:"required"`
		Severity  string `json:"severity"` // Optional: override severity for this monitor
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mapping := &services.PagerDutyMonitorMapping{
		IntegrationID: integrationID,
		MonitorID:     req.MonitorID,
		Severity:      req.Severity,
		IsActive:      true,
	}

	if err := h.pagerdutyService.MapMonitor(mapping); err != nil {
		h.logger.Error("Failed to map monitor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to map monitor"})
		return
	}

	h.logger.Info("Monitor mapped to PagerDuty",
		zap.Uint("integration_id", integrationID),
		zap.Uint("monitor_id", req.MonitorID),
	)

	c.JSON(http.StatusCreated, gin.H{"mapping": mapping})
}

// UnmapMonitor removes a monitor mapping
// DELETE /api/v1/integrations/pagerduty/:id/monitors/:mapping_id
func (h *PagerDutyHandler) UnmapMonitor(c *gin.Context) {
	mappingIDStr := c.Param("mapping_id")
	mappingID, err := parseUint(mappingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mapping ID"})
		return
	}

	if err := h.pagerdutyService.UnmapMonitor(mappingID); err != nil {
		h.logger.Error("Failed to unmap monitor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unmap monitor"})
		return
	}

	h.logger.Info("Monitor unmapped", zap.Uint("mapping_id", mappingID))

	c.JSON(http.StatusOK, gin.H{"message": "monitor unmapped successfully"})
}

// GetIncidentHistory returns incident history for a monitor
// GET /api/v1/integrations/pagerduty/incidents
func (h *PagerDutyHandler) GetIncidentHistory(c *gin.Context) {
	monitorIDStr := c.Query("monitor_id")
	if monitorIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "monitor_id is required"})
		return
	}

	monitorID, err := parseUint(monitorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid monitor_id"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	incidents, err := h.pagerdutyService.GetIncidentHistory(monitorID, limit)
	if err != nil {
		h.logger.Error("Failed to get incident history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get incident history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"incidents": incidents})
}

// WebhookReceiver handles incoming webhooks from PagerDuty
// POST /api/v1/integrations/pagerduty/webhook
func (h *PagerDutyHandler) WebhookReceiver(c *gin.Context) {
	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read webhook body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Verify webhook signature if secret is configured
	if h.webhookSecret != "" {
		signature := c.GetHeader("X-PagerDuty-Signature")
		if !h.verifyWebhookSignature(body, signature) {
			h.logger.Warn("Invalid webhook signature")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
	}

	// Parse webhook payload
	var webhook PagerDutyWebhook
	if err := json.Unmarshal(body, &webhook); err != nil {
		h.logger.Error("Failed to parse webhook", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook payload"})
		return
	}

	h.logger.Info("PagerDuty webhook received",
		zap.Int("event_count", len(webhook.Messages)),
	)

	// Process webhook messages
	for _, message := range webhook.Messages {
		h.processWebhookMessage(message)
	}

	c.JSON(http.StatusOK, gin.H{"message": "webhook processed successfully"})
}

// PagerDutyWebhook represents the webhook payload from PagerDuty
type PagerDutyWebhook struct {
	Messages []PagerDutyWebhookMessage `json:"messages"`
}

// PagerDutyWebhookMessage represents a single webhook message
type PagerDutyWebhookMessage struct {
	ID        string                     `json:"id"`
	Event     string                     `json:"event"` // incident.triggered, incident.acknowledged, incident.resolved
	CreatedOn string                     `json:"created_on"`
	Incident  PagerDutyWebhookIncident   `json:"incident"`
	LogEntries []interface{}             `json:"log_entries,omitempty"`
}

// PagerDutyWebhookIncident represents incident data in webhook
type PagerDutyWebhookIncident struct {
	ID             string `json:"id"`
	IncidentNumber int    `json:"incident_number"`
	Status         string `json:"status"`
	Title          string `json:"title"`
	IncidentKey    string `json:"incident_key"`
	Service        struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"service"`
	Assignments []struct {
		Assignee struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"assignee"`
	} `json:"assignments"`
}

// processWebhookMessage processes a single webhook message
func (h *PagerDutyHandler) processWebhookMessage(message PagerDutyWebhookMessage) {
	h.logger.Info("Processing webhook message",
		zap.String("event", message.Event),
		zap.String("incident_id", message.Incident.ID),
		zap.String("incident_key", message.Incident.IncidentKey),
	)

	// Find incident by incident_key
	var incident services.PagerDutyIncident
	err := h.db.Where("incident_key = ?", message.Incident.IncidentKey).First(&incident).Error
	if err != nil {
		h.logger.Warn("Incident not found for webhook",
			zap.String("incident_key", message.Incident.IncidentKey),
			zap.Error(err),
		)
		return
	}

	// Update incident based on webhook event
	switch message.Event {
	case "incident.acknowledged":
		h.handleIncidentAcknowledged(&incident, message)
	case "incident.resolved":
		h.handleIncidentResolved(&incident, message)
	case "incident.triggered":
		h.handleIncidentTriggered(&incident, message)
	default:
		h.logger.Info("Unhandled webhook event", zap.String("event", message.Event))
	}
}

// handleIncidentAcknowledged handles acknowledged webhook events
func (h *PagerDutyHandler) handleIncidentAcknowledged(incident *services.PagerDutyIncident, message PagerDutyWebhookMessage) {
	// Update incident status
	incident.Status = "acknowledged"
	incident.PDIncidentID = message.Incident.ID

	if err := h.db.Save(incident).Error; err != nil {
		h.logger.Error("Failed to update incident", zap.Error(err))
		return
	}

	h.logger.Info("Incident acknowledged",
		zap.String("incident_key", incident.IncidentKey),
		zap.String("pd_incident_id", message.Incident.ID),
	)
}

// handleIncidentResolved handles resolved webhook events
func (h *PagerDutyHandler) handleIncidentResolved(incident *services.PagerDutyIncident, message PagerDutyWebhookMessage) {
	// Update incident status
	incident.Status = "resolved"
	incident.PDIncidentID = message.Incident.ID

	if err := h.db.Save(incident).Error; err != nil {
		h.logger.Error("Failed to update incident", zap.Error(err))
		return
	}

	h.logger.Info("Incident resolved",
		zap.String("incident_key", incident.IncidentKey),
		zap.String("pd_incident_id", message.Incident.ID),
	)
}

// handleIncidentTriggered handles triggered webhook events
func (h *PagerDutyHandler) handleIncidentTriggered(incident *services.PagerDutyIncident, message PagerDutyWebhookMessage) {
	// Update incident with PagerDuty incident ID
	incident.PDIncidentID = message.Incident.ID

	if err := h.db.Save(incident).Error; err != nil {
		h.logger.Error("Failed to update incident", zap.Error(err))
		return
	}

	h.logger.Info("Incident triggered",
		zap.String("incident_key", incident.IncidentKey),
		zap.String("pd_incident_id", message.Incident.ID),
	)
}

// verifyWebhookSignature verifies PagerDuty webhook signature
func (h *PagerDutyHandler) verifyWebhookSignature(body []byte, signature string) bool {
	if signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

// Helper function to parse uint
func parseUint(s string) (uint, error) {
	var n uint64
	_, err := fmt.Sscanf(s, "%d", &n)
	return uint(n), err
}
