// Package handlers provides enhanced HTTP handlers for the Notification Service.
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/anupamdutta5/statuspage-notification-service/internal/models"
	"github.com/anupamdutta5/statuspage-notification-service/internal/providers"
	"github.com/anupamdutta5/statuspage-notification-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EnhancedNotificationHandler handles notification-related HTTP requests with provider support.
type EnhancedNotificationHandler struct {
	enhancedService *services.EnhancedNotificationService
	logger          *zap.Logger
}

// NewEnhancedNotificationHandler creates a new enhanced notification handler.
func NewEnhancedNotificationHandler(enhancedService *services.EnhancedNotificationService, logger *zap.Logger) *EnhancedNotificationHandler {
	return &EnhancedNotificationHandler{
		enhancedService: enhancedService,
		logger:          logger,
	}
}

// Health returns the health status of the Notification Service.
func (h *EnhancedNotificationHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "notification-service",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// SendNotification handles sending a notification.
func (h *EnhancedNotificationHandler) SendNotification(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		Type       string                 `json:"type" binding:"required"`
		Priority   string                 `json:"priority"`
		Subject    string                 `json:"subject" binding:"required"`
		Content    string                 `json:"content" binding:"required"`
		Recipients []string               `json:"recipients" binding:"required"`
		Metadata   map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid send notification request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Convert recipients to JSON
	recipientsJSON, err := json.Marshal(req.Recipients)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipients format"})
		return
	}

	// Convert metadata to JSON
	var metadataJSON string
	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metadata format"})
			return
		}
		metadataJSON = string(metadataBytes)
	}

	// Set default priority if not provided
	if req.Priority == "" {
		req.Priority = "normal"
	}

	// Create notification
	notification := &models.Notification{
		TenantID:   tenantID.(uint),
		Type:       req.Type,
		Status:     "pending",
		Priority:   req.Priority,
		Subject:    req.Subject,
		Content:    req.Content,
		Recipients: string(recipientsJSON),
		Metadata:   metadataJSON,
	}

	if err := h.enhancedService.SendNotification(c.Request.Context(), notification); err != nil {
		h.logger.Error("Failed to send notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send notification"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Notification sent successfully",
		"notification": notification,
	})
}

// SendMaintenanceNotification handles sending maintenance notifications.
func (h *EnhancedNotificationHandler) SendMaintenanceNotification(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		Type             string                 `json:"type" binding:"required"` // maintenance.scheduled, maintenance.started, maintenance.completed
		MaintenanceID    uint                   `json:"maintenance_id" binding:"required"`
		Title            string                 `json:"title" binding:"required"`
		Description      string                 `json:"description"`
		StartTime        string                 `json:"start_time"`
		EndTime          string                 `json:"end_time"`
		Impact           string                 `json:"impact"`
		AffectedServices []string               `json:"affected_services"`
		Metadata         map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid maintenance notification request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Prepare maintenance data
	maintenanceData := map[string]interface{}{
		"maintenance_id":     req.MaintenanceID,
		"title":              req.Title,
		"description":        req.Description,
		"start_time":         req.StartTime,
		"end_time":           req.EndTime,
		"impact":             req.Impact,
		"affected_services":  req.AffectedServices,
	}

	// Add any additional metadata
	if req.Metadata != nil {
		for key, value := range req.Metadata {
			maintenanceData[key] = value
		}
	}

	if err := h.enhancedService.SendMaintenanceNotification(c.Request.Context(), tenantID.(uint), maintenanceData, req.Type); err != nil {
		h.logger.Error("Failed to send maintenance notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send maintenance notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Maintenance notification sent successfully"})
}

// SendIncidentNotification handles sending incident notifications.
func (h *EnhancedNotificationHandler) SendIncidentNotification(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		Type             string                 `json:"type" binding:"required"` // incident.created, incident.updated, incident.resolved
		IncidentID       uint                   `json:"incident_id" binding:"required"`
		Title            string                 `json:"title" binding:"required"`
		Description      string                 `json:"description"`
		Status           string                 `json:"status"`
		Severity         string                 `json:"severity"`
		AffectedServices []string               `json:"affected_services"`
		Metadata         map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid incident notification request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Prepare incident data
	incidentData := map[string]interface{}{
		"incident_id":        req.IncidentID,
		"title":              req.Title,
		"description":        req.Description,
		"status":             req.Status,
		"severity":           req.Severity,
		"affected_services":  req.AffectedServices,
	}

	// Add any additional metadata
	if req.Metadata != nil {
		for key, value := range req.Metadata {
			incidentData[key] = value
		}
	}

	if err := h.enhancedService.SendIncidentNotification(c.Request.Context(), tenantID.(uint), incidentData, req.Type); err != nil {
		h.logger.Error("Failed to send incident notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send incident notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incident notification sent successfully"})
}

// GetProviders returns all registered notification providers.
func (h *EnhancedNotificationHandler) GetProviders(c *gin.Context) {
	providerManager := h.enhancedService.GetProviderManager()
	stats := providerManager.GetProviderStats()

	c.JSON(http.StatusOK, gin.H{
		"providers": stats,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ConfigureProvider handles provider configuration.
func (h *EnhancedNotificationHandler) ConfigureProvider(c *gin.Context) {
	var config providers.ProviderConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		h.logger.Error("Invalid provider configuration request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err := h.enhancedService.ConfigureProvider(&config); err != nil {
		h.logger.Error("Failed to configure provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to configure provider"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Provider configured successfully"})
}

// TestProvider handles testing a provider configuration.
func (h *EnhancedNotificationHandler) TestProvider(c *gin.Context) {
	var req struct {
		Config        providers.ProviderConfig `json:"config" binding:"required"`
		TestRecipient string                   `json:"test_recipient" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid test provider request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	response, err := h.enhancedService.TestProvider(c.Request.Context(), &req.Config, req.TestRecipient)
	if err != nil {
		h.logger.Error("Failed to test provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to test provider"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Provider test completed",
		"response": response,
	})
}

// GetChannels returns notification channels for a tenant.
func (h *EnhancedNotificationHandler) GetChannels(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	// This would typically call a method to get channels from the database
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"channels": []map[string]interface{}{
			{
				"id":          1,
				"name":        "Email",
				"type":        "email",
				"provider":    "smtp",
				"is_active":   true,
				"is_default":  true,
			},
		},
		"tenant_id": tenantID,
	})
}

// CreateChannel handles creating a new notification channel.
func (h *EnhancedNotificationHandler) CreateChannel(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		Name        string                 `json:"name" binding:"required"`
		Description string                 `json:"description"`
		Type        string                 `json:"type" binding:"required"`
		Provider    string                 `json:"provider" binding:"required"`
		Config      map[string]interface{} `json:"config" binding:"required"`
		IsActive    *bool                  `json:"is_active"`
		IsDefault   *bool                  `json:"is_default"`
		Priority    int                    `json:"priority"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create channel request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Convert config to JSON
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config format"})
		return
	}

	// Create channel (this would typically use a database)
	channel := map[string]interface{}{
		"tenant_id":   tenantID,
		"name":        req.Name,
		"description": req.Description,
		"type":        req.Type,
		"provider":    req.Provider,
		"config":      string(configJSON),
		"is_active":   req.IsActive != nil && *req.IsActive,
		"is_default":  req.IsDefault != nil && *req.IsDefault,
		"priority":    req.Priority,
		"created_at":  time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Channel created successfully",
		"channel": channel,
	})
}

// GetTemplates returns notification templates for a tenant.
func (h *EnhancedNotificationHandler) GetTemplates(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	// This would typically call a method to get templates from the database
	// For now, return placeholder templates
	templates := []map[string]interface{}{
		{
			"id":          1,
			"name":        "Incident Alert",
			"type":        "incident",
			"category":    "alert",
			"subject":     "Incident: {{.Title}}",
			"content":     "An incident has been reported: {{.Description}}",
			"is_active":   true,
			"is_default":  true,
		},
		{
			"id":          2,
			"name":        "Maintenance Notice",
			"type":        "maintenance",
			"category":    "maintenance",
			"subject":     "Scheduled Maintenance: {{.Title}}",
			"content":     "Maintenance scheduled: {{.Description}} from {{.StartTime}} to {{.EndTime}}",
			"is_active":   true,
			"is_default":  true,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"tenant_id": tenantID,
	})
}

// GetSubscriptions returns notification subscriptions for a tenant.
func (h *EnhancedNotificationHandler) GetSubscriptions(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	// This would typically call a method to get subscriptions from the database
	// For now, return placeholder data
	subscriptions := []map[string]interface{}{
		{
			"id":          1,
			"user_id":     1,
			"channel_id":  1,
			"event_types": []string{"incident", "maintenance"},
			"is_active":   true,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"subscriptions": subscriptions,
		"total":         1,
		"limit":         limit,
		"offset":        offset,
		"tenant_id":     tenantID,
	})
}