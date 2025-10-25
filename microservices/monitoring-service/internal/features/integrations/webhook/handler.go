// Package handlers provides HTTP handlers for webhook management in the Monitoring Service.
package webhook

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/anupamdutta5/monitoring-service/internal/models"
	"github.com/anupamdutta5/monitoring-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// WebhookHandler handles webhook-related HTTP requests.
type WebhookHandler struct {
	webhookService *services.WebhookService
	logger         *zap.Logger
}

// NewWebhookHandler creates a new webhook handler.
func NewWebhookHandler(webhookService *services.WebhookService, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
		logger:         logger,
	}
}

// CreateWebhookEndpoint handles creating a new webhook endpoint.
func (h *WebhookHandler) CreateWebhookEndpoint(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		Name        string            `json:"name" binding:"required"`
		Description string            `json:"description"`
		URL         string            `json:"url" binding:"required,url"`
		SecretKey   string            `json:"secret_key"`
		Events      []string          `json:"events" binding:"required"`
		Headers     map[string]string `json:"headers"`
		Timeout     int               `json:"timeout"`
		RetryCount  int               `json:"retry_count"`
		IsActive    *bool             `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create webhook endpoint request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Convert events to JSON
	eventsJSON, err := json.Marshal(req.Events)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid events format"})
		return
	}

	// Convert headers to JSON
	var headersJSON string
	if req.Headers != nil {
		headersBytes, err := json.Marshal(req.Headers)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid headers format"})
			return
		}
		headersJSON = string(headersBytes)
	}

	// Set defaults
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	if req.Timeout == 0 {
		req.Timeout = 30
	}
	if req.RetryCount == 0 {
		req.RetryCount = 3
	}

	endpoint := &models.WebhookEndpoint{
		TenantID:    tenantID.(uint),
		Name:        req.Name,
		Description: req.Description,
		URL:         req.URL,
		SecretKey:   req.SecretKey,
		Events:      string(eventsJSON),
		Headers:     headersJSON,
		Timeout:     req.Timeout,
		RetryCount:  req.RetryCount,
		IsActive:    isActive,
	}

	if err := h.webhookService.CreateWebhookEndpoint(c.Request.Context(), endpoint); err != nil {
		h.logger.Error("Failed to create webhook endpoint", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create webhook endpoint"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Webhook endpoint created successfully",
		"webhook": endpoint,
	})
}

// GetWebhookEndpoints handles retrieving webhook endpoints.
func (h *WebhookHandler) GetWebhookEndpoints(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	activeOnly := c.Query("active_only") == "true"

	endpoints, err := h.webhookService.GetWebhookEndpoints(c.Request.Context(), tenantID.(uint), activeOnly)
	if err != nil {
		h.logger.Error("Failed to get webhook endpoints", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve webhook endpoints"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"webhooks": endpoints,
		"count":    len(endpoints),
	})
}

// GetWebhookEndpoint handles retrieving a specific webhook endpoint.
func (h *WebhookHandler) GetWebhookEndpoint(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	webhookIDStr := c.Param("id")
	webhookID, err := strconv.ParseUint(webhookIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	endpoints, err := h.webhookService.GetWebhookEndpoints(c.Request.Context(), tenantID.(uint), false)
	if err != nil {
		h.logger.Error("Failed to get webhook endpoints", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve webhook endpoint"})
		return
	}

	// Find the specific webhook
	for _, endpoint := range endpoints {
		if endpoint.ID == uint(webhookID) {
			c.JSON(http.StatusOK, gin.H{
				"webhook": endpoint,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Webhook endpoint not found"})
}

// UpdateWebhookEndpoint handles updating a webhook endpoint.
func (h *WebhookHandler) UpdateWebhookEndpoint(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	webhookIDStr := c.Param("id")
	webhookID, err := strconv.ParseUint(webhookIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	var req struct {
		Name        *string            `json:"name"`
		Description *string            `json:"description"`
		URL         *string            `json:"url"`
		SecretKey   *string            `json:"secret_key"`
		Events      *[]string          `json:"events"`
		Headers     *map[string]string `json:"headers"`
		Timeout     *int               `json:"timeout"`
		RetryCount  *int               `json:"retry_count"`
		IsActive    *bool              `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update webhook endpoint request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.URL != nil {
		updates["url"] = *req.URL
	}
	if req.SecretKey != nil {
		updates["secret_key"] = *req.SecretKey
	}
	if req.Events != nil {
		eventsJSON, err := json.Marshal(*req.Events)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid events format"})
			return
		}
		updates["events"] = string(eventsJSON)
	}
	if req.Headers != nil {
		headersJSON, err := json.Marshal(*req.Headers)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid headers format"})
			return
		}
		updates["headers"] = string(headersJSON)
	}
	if req.Timeout != nil {
		updates["timeout"] = *req.Timeout
	}
	if req.RetryCount != nil {
		updates["retry_count"] = *req.RetryCount
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := h.webhookService.UpdateWebhookEndpoint(c.Request.Context(), uint(webhookID), tenantID.(uint), updates); err != nil {
		h.logger.Error("Failed to update webhook endpoint", zap.Error(err))
		if err.Error() == "webhook endpoint not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Webhook endpoint not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update webhook endpoint"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook endpoint updated successfully"})
}

// DeleteWebhookEndpoint handles deleting a webhook endpoint.
func (h *WebhookHandler) DeleteWebhookEndpoint(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	webhookIDStr := c.Param("id")
	webhookID, err := strconv.ParseUint(webhookIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	if err := h.webhookService.DeleteWebhookEndpoint(c.Request.Context(), uint(webhookID), tenantID.(uint)); err != nil {
		h.logger.Error("Failed to delete webhook endpoint", zap.Error(err))
		if err.Error() == "webhook endpoint not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Webhook endpoint not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete webhook endpoint"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook endpoint deleted successfully"})
}

// TestWebhookEndpoint handles testing a webhook endpoint.
func (h *WebhookHandler) TestWebhookEndpoint(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	webhookIDStr := c.Param("id")
	webhookID, err := strconv.ParseUint(webhookIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	if err := h.webhookService.TestWebhookEndpoint(c.Request.Context(), uint(webhookID), tenantID.(uint)); err != nil {
		h.logger.Error("Failed to test webhook endpoint", zap.Error(err))
		if err.Error() == "webhook endpoint not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Webhook endpoint not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to test webhook endpoint"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test webhook sent successfully"})
}

// GetWebhookDeliveries handles retrieving webhook delivery history.
func (h *WebhookHandler) GetWebhookDeliveries(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var webhookID *uint
	if webhookIDStr := c.Query("webhook_id"); webhookIDStr != "" {
		if id, err := strconv.ParseUint(webhookIDStr, 10, 32); err == nil {
			webhookIDVal := uint(id)
			webhookID = &webhookIDVal
		}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	deliveries, total, err := h.webhookService.GetWebhookDeliveries(c.Request.Context(), tenantID.(uint), webhookID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get webhook deliveries", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve webhook deliveries"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deliveries": deliveries,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}

// RetryWebhookDelivery handles retrying a failed webhook delivery.
func (h *WebhookHandler) RetryWebhookDelivery(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	deliveryIDStr := c.Param("id")
	deliveryID, err := strconv.ParseUint(deliveryIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid delivery ID"})
		return
	}

	if err := h.webhookService.RetryWebhookDelivery(c.Request.Context(), uint(deliveryID), tenantID.(uint)); err != nil {
		h.logger.Error("Failed to retry webhook delivery", zap.Error(err))
		if err.Error() == "webhook delivery not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Webhook delivery not found"})
		} else if err.Error() == "webhook delivery already successful" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook delivery already successful"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retry webhook delivery"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook delivery retry initiated"})
}

// GetWebhookStatistics handles retrieving webhook statistics.
func (h *WebhookHandler) GetWebhookStatistics(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var webhookID *uint
	if webhookIDStr := c.Query("webhook_id"); webhookIDStr != "" {
		if id, err := strconv.ParseUint(webhookIDStr, 10, 32); err == nil {
			webhookIDVal := uint(id)
			webhookID = &webhookIDVal
		}
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days > 90 {
		days = 90 // Limit to 90 days
	}

	stats, err := h.webhookService.GetWebhookStatistics(c.Request.Context(), tenantID.(uint), webhookID, days)
	if err != nil {
		h.logger.Error("Failed to get webhook statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve webhook statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statistics": stats,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

// GetSupportedEventTypes returns the list of supported webhook event types.
func (h *WebhookHandler) GetSupportedEventTypes(c *gin.Context) {
	eventTypes := []map[string]interface{}{
		{
			"type":        models.EventTypeComponentStatusChanged,
			"description": "Component status changed (up, down, degraded)",
			"category":    "component",
		},
		{
			"type":        models.EventTypeComponentCreated,
			"description": "New component created",
			"category":    "component",
		},
		{
			"type":        models.EventTypeComponentUpdated,
			"description": "Component configuration updated",
			"category":    "component",
		},
		{
			"type":        models.EventTypeComponentDeleted,
			"description": "Component deleted",
			"category":    "component",
		},
		{
			"type":        models.EventTypeIncidentCreated,
			"description": "New incident created",
			"category":    "incident",
		},
		{
			"type":        models.EventTypeIncidentUpdated,
			"description": "Incident updated with new information",
			"category":    "incident",
		},
		{
			"type":        models.EventTypeIncidentResolved,
			"description": "Incident resolved",
			"category":    "incident",
		},
		{
			"type":        models.EventTypeMaintenanceScheduled,
			"description": "Maintenance window scheduled",
			"category":    "maintenance",
		},
		{
			"type":        models.EventTypeMaintenanceStarted,
			"description": "Maintenance window started",
			"category":    "maintenance",
		},
		{
			"type":        models.EventTypeMaintenanceCompleted,
			"description": "Maintenance window completed",
			"category":    "maintenance",
		},
		{
			"type":        models.EventTypeMaintenanceCancelled,
			"description": "Maintenance window cancelled",
			"category":    "maintenance",
		},
		{
			"type":        models.EventTypeMetricThresholdExceeded,
			"description": "Metric threshold exceeded",
			"category":    "metric",
		},
		{
			"type":        models.EventTypeMonitorCheckFailed,
			"description": "Monitor check failed",
			"category":    "monitor",
		},
		{
			"type":        models.EventTypeMonitorCheckPassed,
			"description": "Monitor check passed",
			"category":    "monitor",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"event_types": eventTypes,
		"count":       len(eventTypes),
	})
}