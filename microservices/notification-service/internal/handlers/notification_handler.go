// Package handlers provides HTTP handlers for the Notification Service.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/anupamdutta5/notification-service/internal/models"
	"github.com/anupamdutta5/notification-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NotificationHandler handles notification-related HTTP requests.
type NotificationHandler struct {
	notificationService *services.NotificationService
	logger              *zap.Logger
}

// NewNotificationHandler creates a new notification handler.
func NewNotificationHandler(notificationService *services.NotificationService, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		logger:              logger,
	}
}

// Health returns the health status of the Notification Service.
func (h *NotificationHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "notification-service",
		"version":   "1.0.0",
		"timestamp": "2024-01-01T00:00:00Z",
	})
}

// HandleWebhook handles incoming webhook events.
func (h *NotificationHandler) HandleWebhook(c *gin.Context) {
	// For now, use tenant ID 1 as default
	// In production, this would be determined from the request context
	tenantID := uint(1)

	var req struct {
		EventType string `json:"event_type" binding:"required"`
		EventData string `json:"event_data" binding:"required"`
		Source    string `json:"source"`
		Metadata  string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid webhook request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	webhookEvent := &models.WebhookEvent{
		TenantID:  tenantID,
		EventType: req.EventType,
		EventData: req.EventData,
		Source:    req.Source,
		Metadata:  req.Metadata,
	}

	if err := h.notificationService.HandleWebhook(webhookEvent); err != nil {
		h.logger.Error("Failed to handle webhook", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to handle webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Webhook processed successfully",
		"event_id": webhookEvent.ID,
	})
}

// Notification Management Handlers

// GetNotifications handles getting a list of notifications.
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
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

	notifications, total, err := h.notificationService.GetNotifications(tenantID.(uint), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get notifications", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"total":         total,
		"limit":         limit,
		"offset":        offset,
	})
}

// GetNotification handles getting a specific notification.
func (h *NotificationHandler) GetNotification(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	notification, err := h.notificationService.GetNotification(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notification": notification})
}

// CreateNotification handles creating a new notification.
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		UserID      *uint  `json:"user_id"`
		TemplateID  *uint  `json:"template_id"`
		Type        string `json:"type" binding:"required"`
		Priority    string `json:"priority"`
		Subject     string `json:"subject"`
		Content     string `json:"content" binding:"required"`
		Recipients  string `json:"recipients" binding:"required"`
		Metadata    string `json:"metadata"`
		ScheduledAt string `json:"scheduled_at"`
		MaxRetries  int    `json:"max_retries"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create notification request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	notification := &models.Notification{
		TenantID:   tenantID.(uint),
		UserID:     req.UserID,
		TemplateID: req.TemplateID,
		Type:       req.Type,
		Priority:   req.Priority,
		Subject:    req.Subject,
		Content:    req.Content,
		Recipients: req.Recipients,
		Metadata:   req.Metadata,
		MaxRetries: req.MaxRetries,
	}

	if req.ScheduledAt != "" {
		// Parse scheduled time
		// For now, we'll leave it as nil
	}

	if err := h.notificationService.CreateNotification(notification); err != nil {
		h.logger.Error("Failed to create notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Notification created successfully",
		"notification": notification,
	})
}

// UpdateNotification handles updating a notification.
func (h *NotificationHandler) UpdateNotification(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	notification, err := h.notificationService.GetNotification(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	var req struct {
		Type       string `json:"type"`
		Priority   string `json:"priority"`
		Status     string `json:"status"`
		Subject    string `json:"subject"`
		Content    string `json:"content"`
		Recipients string `json:"recipients"`
		Metadata   string `json:"metadata"`
		MaxRetries *int   `json:"max_retries"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update notification request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update fields
	if req.Type != "" {
		notification.Type = req.Type
	}
	if req.Priority != "" {
		notification.Priority = req.Priority
	}
	if req.Status != "" {
		notification.Status = req.Status
	}
	if req.Subject != "" {
		notification.Subject = req.Subject
	}
	if req.Content != "" {
		notification.Content = req.Content
	}
	if req.Recipients != "" {
		notification.Recipients = req.Recipients
	}
	if req.Metadata != "" {
		notification.Metadata = req.Metadata
	}
	if req.MaxRetries != nil {
		notification.MaxRetries = *req.MaxRetries
	}

	if err := h.notificationService.UpdateNotification(notification); err != nil {
		h.logger.Error("Failed to update notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Notification updated successfully",
		"notification": notification,
	})
}

// DeleteNotification handles deleting a notification.
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	if err := h.notificationService.DeleteNotification(uint(id)); err != nil {
		h.logger.Error("Failed to delete notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

// SendNotification handles sending a notification.
func (h *NotificationHandler) SendNotification(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	if err := h.notificationService.SendNotification(uint(id)); err != nil {
		h.logger.Error("Failed to send notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification sent successfully"})
}

// GetNotificationStatus handles getting notification status.
func (h *NotificationHandler) GetNotificationStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	status, err := h.notificationService.GetNotificationStatus(uint(id))
	if err != nil {
		h.logger.Error("Failed to get notification status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notification status"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// Template Management Handlers

// GetTemplates handles getting a list of templates.
func (h *NotificationHandler) GetTemplates(c *gin.Context) {
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

	templates, total, err := h.notificationService.GetTemplates(tenantID.(uint), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get templates", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetTemplate handles getting a specific template.
func (h *NotificationHandler) GetTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	template, err := h.notificationService.GetTemplate(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"template": template})
}

// CreateTemplate handles creating a new template.
func (h *NotificationHandler) CreateTemplate(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type" binding:"required"`
		Category    string `json:"category"`
		Subject     string `json:"subject"`
		Content     string `json:"content" binding:"required"`
		Variables   string `json:"variables"`
		IsActive    *bool  `json:"is_active"`
		IsDefault   *bool  `json:"is_default"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create template request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	template := &models.Template{
		TenantID:    tenantID.(uint),
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Category:    req.Category,
		Subject:     req.Subject,
		Content:     req.Content,
		Variables:   req.Variables,
		Metadata:    req.Metadata,
	}

	if req.IsActive != nil {
		template.IsActive = *req.IsActive
	}
	if req.IsDefault != nil {
		template.IsDefault = *req.IsDefault
	}

	if err := h.notificationService.CreateTemplate(template); err != nil {
		h.logger.Error("Failed to create template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Template created successfully",
		"template": template,
	})
}

// UpdateTemplate handles updating a template.
func (h *NotificationHandler) UpdateTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	template, err := h.notificationService.GetTemplate(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Category    string `json:"category"`
		Subject     string `json:"subject"`
		Content     string `json:"content"`
		Variables   string `json:"variables"`
		IsActive    *bool  `json:"is_active"`
		IsDefault   *bool  `json:"is_default"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update template request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update fields
	if req.Name != "" {
		template.Name = req.Name
	}
	if req.Description != "" {
		template.Description = req.Description
	}
	if req.Type != "" {
		template.Type = req.Type
	}
	if req.Category != "" {
		template.Category = req.Category
	}
	if req.Subject != "" {
		template.Subject = req.Subject
	}
	if req.Content != "" {
		template.Content = req.Content
	}
	if req.Variables != "" {
		template.Variables = req.Variables
	}
	if req.IsActive != nil {
		template.IsActive = *req.IsActive
	}
	if req.IsDefault != nil {
		template.IsDefault = *req.IsDefault
	}
	if req.Metadata != "" {
		template.Metadata = req.Metadata
	}

	if err := h.notificationService.UpdateTemplate(template); err != nil {
		h.logger.Error("Failed to update template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Template updated successfully",
		"template": template,
	})
}

// DeleteTemplate handles deleting a template.
func (h *NotificationHandler) DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	if err := h.notificationService.DeleteTemplate(uint(id)); err != nil {
		h.logger.Error("Failed to delete template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully"})
}

// TestTemplate handles testing a template.
func (h *NotificationHandler) TestTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var req struct {
		TestRecipient string `json:"test_recipient" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid test template request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err := h.notificationService.TestTemplate(uint(id), req.TestRecipient); err != nil {
		h.logger.Error("Failed to test template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to test template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template test completed successfully"})
}

// Channel Management Handlers

// GetChannels handles getting a list of channels.
func (h *NotificationHandler) GetChannels(c *gin.Context) {
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

	channels, total, err := h.notificationService.GetChannels(tenantID.(uint), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get channels", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get channels"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"channels": channels,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetChannel handles getting a specific channel.
func (h *NotificationHandler) GetChannel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	channel, err := h.notificationService.GetChannel(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"channel": channel})
}

// CreateChannel handles creating a new channel.
func (h *NotificationHandler) CreateChannel(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type" binding:"required"`
		Provider    string `json:"provider"`
		Config      string `json:"config" binding:"required"`
		IsActive    *bool  `json:"is_active"`
		IsDefault   *bool  `json:"is_default"`
		Priority    int    `json:"priority"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create channel request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	channel := &models.Channel{
		TenantID:    tenantID.(uint),
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Provider:    req.Provider,
		Config:      req.Config,
		Priority:    req.Priority,
		Metadata:    req.Metadata,
	}

	if req.IsActive != nil {
		channel.IsActive = *req.IsActive
	}
	if req.IsDefault != nil {
		channel.IsDefault = *req.IsDefault
	}

	if err := h.notificationService.CreateChannel(channel); err != nil {
		h.logger.Error("Failed to create channel", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create channel"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Channel created successfully",
		"channel": channel,
	})
}

// UpdateChannel handles updating a channel.
func (h *NotificationHandler) UpdateChannel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	channel, err := h.notificationService.GetChannel(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Provider    string `json:"provider"`
		Config      string `json:"config"`
		IsActive    *bool  `json:"is_active"`
		IsDefault   *bool  `json:"is_default"`
		Priority    *int   `json:"priority"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update channel request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update fields
	if req.Name != "" {
		channel.Name = req.Name
	}
	if req.Description != "" {
		channel.Description = req.Description
	}
	if req.Type != "" {
		channel.Type = req.Type
	}
	if req.Provider != "" {
		channel.Provider = req.Provider
	}
	if req.Config != "" {
		channel.Config = req.Config
	}
	if req.IsActive != nil {
		channel.IsActive = *req.IsActive
	}
	if req.IsDefault != nil {
		channel.IsDefault = *req.IsDefault
	}
	if req.Priority != nil {
		channel.Priority = *req.Priority
	}
	if req.Metadata != "" {
		channel.Metadata = req.Metadata
	}

	if err := h.notificationService.UpdateChannel(channel); err != nil {
		h.logger.Error("Failed to update channel", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update channel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Channel updated successfully",
		"channel": channel,
	})
}

// DeleteChannel handles deleting a channel.
func (h *NotificationHandler) DeleteChannel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	if err := h.notificationService.DeleteChannel(uint(id)); err != nil {
		h.logger.Error("Failed to delete channel", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete channel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Channel deleted successfully"})
}

// TestChannel handles testing a channel.
func (h *NotificationHandler) TestChannel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	var req struct {
		TestRecipient string `json:"test_recipient" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid test channel request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err := h.notificationService.TestChannel(uint(id), req.TestRecipient); err != nil {
		h.logger.Error("Failed to test channel", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to test channel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Channel test completed successfully"})
}

// Subscription Management Handlers

// GetSubscriptions handles getting a list of subscriptions.
func (h *NotificationHandler) GetSubscriptions(c *gin.Context) {
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

	subscriptions, total, err := h.notificationService.GetSubscriptions(tenantID.(uint), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get subscriptions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscriptions": subscriptions,
		"total":         total,
		"limit":         limit,
		"offset":        offset,
	})
}

// GetSubscription handles getting a specific subscription.
func (h *NotificationHandler) GetSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	subscription, err := h.notificationService.GetSubscription(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscription": subscription})
}

// CreateSubscription handles creating a new subscription.
func (h *NotificationHandler) CreateSubscription(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		UserID      uint   `json:"user_id" binding:"required"`
		ChannelID   uint   `json:"channel_id" binding:"required"`
		EventTypes  string `json:"event_types"`
		IsActive    *bool  `json:"is_active"`
		Preferences string `json:"preferences"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create subscription request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	subscription := &models.Subscription{
		TenantID:    tenantID.(uint),
		UserID:      req.UserID,
		ChannelID:   req.ChannelID,
		EventTypes:  req.EventTypes,
		Preferences: req.Preferences,
		Metadata:    req.Metadata,
	}

	if req.IsActive != nil {
		subscription.IsActive = *req.IsActive
	}

	if err := h.notificationService.CreateSubscription(subscription); err != nil {
		h.logger.Error("Failed to create subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Subscription created successfully",
		"subscription": subscription,
	})
}

// UpdateSubscription handles updating a subscription.
func (h *NotificationHandler) UpdateSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	subscription, err := h.notificationService.GetSubscription(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	var req struct {
		UserID      *uint  `json:"user_id"`
		ChannelID   *uint  `json:"channel_id"`
		EventTypes  string `json:"event_types"`
		IsActive    *bool  `json:"is_active"`
		Preferences string `json:"preferences"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update subscription request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update fields
	if req.UserID != nil {
		subscription.UserID = *req.UserID
	}
	if req.ChannelID != nil {
		subscription.ChannelID = *req.ChannelID
	}
	if req.EventTypes != "" {
		subscription.EventTypes = req.EventTypes
	}
	if req.IsActive != nil {
		subscription.IsActive = *req.IsActive
	}
	if req.Preferences != "" {
		subscription.Preferences = req.Preferences
	}
	if req.Metadata != "" {
		subscription.Metadata = req.Metadata
	}

	if err := h.notificationService.UpdateSubscription(subscription); err != nil {
		h.logger.Error("Failed to update subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Subscription updated successfully",
		"subscription": subscription,
	})
}

// DeleteSubscription handles deleting a subscription.
func (h *NotificationHandler) DeleteSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	if err := h.notificationService.DeleteSubscription(uint(id)); err != nil {
		h.logger.Error("Failed to delete subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription deleted successfully"})
}

// Delivery Management Handlers

// GetDeliveries handles getting a list of deliveries.
func (h *NotificationHandler) GetDeliveries(c *gin.Context) {
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

	deliveries, total, err := h.notificationService.GetDeliveries(tenantID.(uint), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get deliveries", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get deliveries"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deliveries": deliveries,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}

// GetDelivery handles getting a specific delivery.
func (h *NotificationHandler) GetDelivery(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid delivery ID"})
		return
	}

	delivery, err := h.notificationService.GetDelivery(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Delivery not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"delivery": delivery})
}

// RetryDelivery handles retrying a failed delivery.
func (h *NotificationHandler) RetryDelivery(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid delivery ID"})
		return
	}

	if err := h.notificationService.RetryDelivery(uint(id)); err != nil {
		h.logger.Error("Failed to retry delivery", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retry delivery"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delivery retry initiated successfully"})
}
