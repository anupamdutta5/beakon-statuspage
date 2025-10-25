package telegram

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/anupamdutta5/monitoring-service/internal/models"
	"github.com/anupamdutta5/monitoring-service/internal/services"
)

// TelegramHandler handles Telegram integration endpoints
type TelegramHandler struct {
	db              *gorm.DB
	logger          *zap.Logger
	telegramService *services.TelegramIntegrationService
}

// NewTelegramHandler creates a new Telegram handler
func NewTelegramHandler(db *gorm.DB, logger *zap.Logger, telegramService *services.TelegramIntegrationService) *TelegramHandler {
	return &TelegramHandler{
		db:              db,
		logger:          logger,
		telegramService: telegramService,
	}
}

// CreateIntegrationRequest represents the request body for creating a Telegram integration
type CreateTelegramIntegrationRequest struct {
	BotToken            string `json:"bot_token" binding:"required"`
	DefaultChatID       string `json:"default_chat_id"`
	NotifyOnDown        *bool  `json:"notify_on_down"`
	NotifyOnUp          *bool  `json:"notify_on_up"`
	NotifyOnDegraded    *bool  `json:"notify_on_degraded"`
	NotifyOnMaintenance *bool  `json:"notify_on_maintenance"`
	UseMarkdown         *bool  `json:"use_markdown"`
	IncludeMonitorURL   *bool  `json:"include_monitor_url"`
	IncludeTimestamp    *bool  `json:"include_timestamp"`
	SilentNotifications *bool  `json:"silent_notifications"`
	DisablePreview      *bool  `json:"disable_preview"`
	RetryCount          *int   `json:"retry_count"`
	RetryIntervalSeconds *int  `json:"retry_interval_seconds"`
}

// UpdateTelegramIntegrationRequest represents the request body for updating a Telegram integration
type UpdateTelegramIntegrationRequest struct {
	BotToken            *string `json:"bot_token"`
	DefaultChatID       *string `json:"default_chat_id"`
	IsActive            *bool   `json:"is_active"`
	NotifyOnDown        *bool   `json:"notify_on_down"`
	NotifyOnUp          *bool   `json:"notify_on_up"`
	NotifyOnDegraded    *bool   `json:"notify_on_degraded"`
	NotifyOnMaintenance *bool   `json:"notify_on_maintenance"`
	UseMarkdown         *bool   `json:"use_markdown"`
	IncludeMonitorURL   *bool   `json:"include_monitor_url"`
	IncludeTimestamp    *bool   `json:"include_timestamp"`
	SilentNotifications *bool   `json:"silent_notifications"`
	DisablePreview      *bool   `json:"disable_preview"`
	RetryCount          *int    `json:"retry_count"`
	RetryIntervalSeconds *int   `json:"retry_interval_seconds"`
}

// SubscribeChatRequest represents the request to subscribe a chat to a monitor
type SubscribeTelegramChatRequest struct {
	MonitorID           uint   `json:"monitor_id" binding:"required"`
	ChatID              string `json:"chat_id" binding:"required"`
	ChatName            string `json:"chat_name"`
	NotifyOnDown        *bool  `json:"notify_on_down"`
	NotifyOnUp          *bool  `json:"notify_on_up"`
	NotifyOnDegraded    *bool  `json:"notify_on_degraded"`
	NotifyOnMaintenance *bool  `json:"notify_on_maintenance"`
	MessageThreadID     *int   `json:"message_thread_id"`
	CustomMessagePrefix string `json:"custom_message_prefix"`
}

// CreateIntegration creates a new Telegram integration
// POST /api/v1/integrations/telegram
func (h *TelegramHandler) CreateIntegration(c *gin.Context) {
	var req CreateTelegramIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	// Get tenant_id from context
	tenantIDStr := c.GetString("tenant_id")
	if tenantIDStr == "" {
		tenantIDStr = c.Query("tenant_id")
		if tenantIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	// Create integration model
	integration := &models.TelegramIntegration{
		TenantID:            tenantID,
		BotToken:            req.BotToken,
		DefaultChatID:       req.DefaultChatID,
		IsActive:            true,
		NotifyOnDown:        ptrBoolOrDefault(req.NotifyOnDown, true),
		NotifyOnUp:          ptrBoolOrDefault(req.NotifyOnUp, true),
		NotifyOnDegraded:    ptrBoolOrDefault(req.NotifyOnDegraded, true),
		NotifyOnMaintenance: ptrBoolOrDefault(req.NotifyOnMaintenance, false),
		UseMarkdown:         ptrBoolOrDefault(req.UseMarkdown, true),
		IncludeMonitorURL:   ptrBoolOrDefault(req.IncludeMonitorURL, true),
		IncludeTimestamp:    ptrBoolOrDefault(req.IncludeTimestamp, true),
		SilentNotifications: ptrBoolOrDefault(req.SilentNotifications, false),
		DisablePreview:      ptrBoolOrDefault(req.DisablePreview, false),
		RetryCount:          ptrIntOrDefault(req.RetryCount, 3),
		RetryIntervalSeconds: ptrIntOrDefault(req.RetryIntervalSeconds, 5),
	}

	// Create integration (validates bot token via Telegram API)
	if err := h.telegramService.CreateIntegration(integration); err != nil {
		h.logger.Error("Failed to create Telegram integration",
			zap.String("tenant_id", tenantID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create integration", "details": err.Error()})
		return
	}

	h.logger.Info("Telegram integration created successfully",
		zap.Uint("id", integration.ID),
		zap.String("tenant_id", tenantID.String()),
		zap.String("bot_username", integration.BotUsername),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message":     "integration created successfully",
		"integration": integration,
	})
}

// UpdateIntegration updates an existing Telegram integration
// PUT /api/v1/integrations/telegram/:id
func (h *TelegramHandler) UpdateIntegration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration ID"})
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	if tenantIDStr == "" {
		tenantIDStr = c.Query("tenant_id")
		if tenantIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	var req UpdateTelegramIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	// Get existing integration
	integration, err := h.telegramService.GetIntegration(id, tenantID)
	if err != nil || integration == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	// Update fields if provided
	if req.BotToken != nil {
		integration.BotToken = *req.BotToken
	}
	if req.DefaultChatID != nil {
		integration.DefaultChatID = *req.DefaultChatID
	}
	if req.IsActive != nil {
		integration.IsActive = *req.IsActive
	}
	if req.NotifyOnDown != nil {
		integration.NotifyOnDown = *req.NotifyOnDown
	}
	if req.NotifyOnUp != nil {
		integration.NotifyOnUp = *req.NotifyOnUp
	}
	if req.NotifyOnDegraded != nil {
		integration.NotifyOnDegraded = *req.NotifyOnDegraded
	}
	if req.NotifyOnMaintenance != nil {
		integration.NotifyOnMaintenance = *req.NotifyOnMaintenance
	}
	if req.UseMarkdown != nil {
		integration.UseMarkdown = *req.UseMarkdown
	}
	if req.IncludeMonitorURL != nil {
		integration.IncludeMonitorURL = *req.IncludeMonitorURL
	}
	if req.IncludeTimestamp != nil {
		integration.IncludeTimestamp = *req.IncludeTimestamp
	}
	if req.SilentNotifications != nil {
		integration.SilentNotifications = *req.SilentNotifications
	}
	if req.DisablePreview != nil {
		integration.DisablePreview = *req.DisablePreview
	}
	if req.RetryCount != nil {
		integration.RetryCount = *req.RetryCount
	}
	if req.RetryIntervalSeconds != nil {
		integration.RetryIntervalSeconds = *req.RetryIntervalSeconds
	}

	// Update integration
	if err := h.telegramService.UpdateIntegration(integration); err != nil {
		h.logger.Error("Failed to update Telegram integration",
			zap.Uint("id", id),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update integration", "details": err.Error()})
		return
	}

	h.logger.Info("Telegram integration updated successfully",
		zap.Uint("id", id),
		zap.String("tenant_id", tenantID.String()),
	)

	c.JSON(http.StatusOK, gin.H{
		"message":     "integration updated successfully",
		"integration": integration,
	})
}

// GetIntegration retrieves a Telegram integration by ID
// GET /api/v1/integrations/telegram/:id
func (h *TelegramHandler) GetIntegration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration ID"})
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	if tenantIDStr == "" {
		tenantIDStr = c.Query("tenant_id")
		if tenantIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	integration, err := h.telegramService.GetIntegration(id, tenantID)
	if err != nil || integration == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"integration": integration})
}

// GetIntegrations retrieves all Telegram integrations for a tenant
// GET /api/v1/integrations/telegram
func (h *TelegramHandler) GetIntegrations(c *gin.Context) {
	tenantIDStr := c.GetString("tenant_id")
	if tenantIDStr == "" {
		tenantIDStr = c.Query("tenant_id")
		if tenantIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	integrations, err := h.telegramService.GetIntegrationsByTenant(tenantID)
	if err != nil {
		h.logger.Error("Failed to get Telegram integrations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve integrations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"integrations": integrations,
		"count":        len(integrations),
	})
}

// DeleteIntegration soft deletes a Telegram integration
// DELETE /api/v1/integrations/telegram/:id
func (h *TelegramHandler) DeleteIntegration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration ID"})
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	if tenantIDStr == "" {
		tenantIDStr = c.Query("tenant_id")
		if tenantIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	if err := h.telegramService.DeleteIntegration(id, tenantID); err != nil {
		h.logger.Error("Failed to delete Telegram integration",
			zap.Uint("id", id),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete integration"})
		return
	}

	h.logger.Info("Telegram integration deleted",
		zap.Uint("id", id),
		zap.String("tenant_id", tenantID.String()),
	)

	c.JSON(http.StatusOK, gin.H{"message": "integration deleted successfully"})
}

// TestIntegration sends a test message via Telegram
// POST /api/v1/integrations/telegram/:id/test
func (h *TelegramHandler) TestIntegration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration ID"})
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	if tenantIDStr == "" {
		tenantIDStr = c.Query("tenant_id")
		if tenantIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	// Get integration
	integration, err := h.telegramService.GetIntegration(id, tenantID)
	if err != nil || integration == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	// Ensure default chat ID is set
	if integration.DefaultChatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "default_chat_id must be set before testing"})
		return
	}

	// Send test message
	if err := h.telegramService.TestBot(integration.BotToken, integration.DefaultChatID); err != nil {
		h.logger.Error("Failed to send test message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send test message", "details": err.Error()})
		return
	}

	h.logger.Info("Test Telegram message sent",
		zap.Uint("id", id),
		zap.String("tenant_id", tenantID.String()),
		zap.String("chat_id", integration.DefaultChatID),
	)

	c.JSON(http.StatusOK, gin.H{"message": "test message sent successfully"})
}

// SubscribeChat subscribes a Telegram chat to a monitor
// POST /api/v1/integrations/telegram/:id/subscribe
func (h *TelegramHandler) SubscribeChat(c *gin.Context) {
	idStr := c.Param("id")
	integrationID, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration ID"})
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	if tenantIDStr == "" {
		tenantIDStr = c.Query("tenant_id")
		if tenantIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	var req SubscribeTelegramChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	// Verify integration exists and belongs to tenant
	integration, err := h.telegramService.GetIntegration(integrationID, tenantID)
	if err != nil || integration == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	// Create subscription
	subscription := &models.TelegramChatSubscription{
		IntegrationID:       integrationID,
		MonitorID:           &req.MonitorID,
		ChatID:              req.ChatID,
		ChatName:            req.ChatName,
		NotifyOnDown:        ptrBoolOrDefault(req.NotifyOnDown, true),
		NotifyOnUp:          ptrBoolOrDefault(req.NotifyOnUp, true),
		NotifyOnDegraded:    ptrBoolOrDefault(req.NotifyOnDegraded, true),
		NotifyOnMaintenance: ptrBoolOrDefault(req.NotifyOnMaintenance, false),
		MessageThreadID:     req.MessageThreadID,
		CustomMessagePrefix: req.CustomMessagePrefix,
		IsActive:            true,
	}

	if err := h.telegramService.SubscribeChat(subscription); err != nil {
		h.logger.Error("Failed to subscribe chat",
			zap.Uint("integration_id", integrationID),
			zap.Uint("monitor_id", req.MonitorID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to subscribe chat", "details": err.Error()})
		return
	}

	h.logger.Info("Telegram chat subscribed to monitor",
		zap.Uint("integration_id", integrationID),
		zap.Uint("monitor_id", req.MonitorID),
		zap.String("chat_id", req.ChatID),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message":      "chat subscribed successfully",
		"subscription": subscription,
	})
}

// UnsubscribeChat removes a chat subscription
// DELETE /api/v1/integrations/telegram/subscriptions/:id
func (h *TelegramHandler) UnsubscribeChat(c *gin.Context) {
	idStr := c.Param("id")
	subscriptionID, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	if err := h.telegramService.UnsubscribeChat(subscriptionID); err != nil {
		h.logger.Error("Failed to unsubscribe chat",
			zap.Uint("subscription_id", subscriptionID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unsubscribe chat"})
		return
	}

	h.logger.Info("Telegram chat unsubscribed",
		zap.Uint("subscription_id", subscriptionID),
	)

	c.JSON(http.StatusOK, gin.H{"message": "chat unsubscribed successfully"})
}

// GetChatSubscriptions retrieves all chat subscriptions for an integration
// GET /api/v1/integrations/telegram/:id/subscriptions
func (h *TelegramHandler) GetChatSubscriptions(c *gin.Context) {
	idStr := c.Param("id")
	integrationID, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration ID"})
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	if tenantIDStr == "" {
		tenantIDStr = c.Query("tenant_id")
		if tenantIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	// Verify integration exists and belongs to tenant
	integration, err := h.telegramService.GetIntegration(integrationID, tenantID)
	if err != nil || integration == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	subscriptions, err := h.telegramService.GetChatSubscriptions(integrationID)
	if err != nil {
		h.logger.Error("Failed to get chat subscriptions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscriptions": subscriptions,
		"count":         len(subscriptions),
	})
}

// GetNotificationHistory retrieves notification history for an integration
// GET /api/v1/integrations/telegram/:id/notifications
func (h *TelegramHandler) GetNotificationHistory(c *gin.Context) {
	idStr := c.Param("id")
	integrationID, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration ID"})
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	if tenantIDStr == "" {
		tenantIDStr = c.Query("tenant_id")
		if tenantIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	// Verify integration exists and belongs to tenant
	integration, err := h.telegramService.GetIntegration(integrationID, tenantID)
	if err != nil || integration == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	// Parse limit from query params
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 500 {
		limit = 50
	}

	notifications, err := h.telegramService.GetNotificationHistory(integrationID, limit)
	if err != nil {
		h.logger.Error("Failed to get notification history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"count":         len(notifications),
	})
}

// GetNotificationStats retrieves statistics for an integration
// GET /api/v1/integrations/telegram/:id/stats
func (h *TelegramHandler) GetNotificationStats(c *gin.Context) {
	idStr := c.Param("id")
	integrationID, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration ID"})
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	if tenantIDStr == "" {
		tenantIDStr = c.Query("tenant_id")
		if tenantIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	// Verify integration exists and belongs to tenant
	integration, err := h.telegramService.GetIntegration(integrationID, tenantID)
	if err != nil || integration == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	stats, err := h.telegramService.GetNotificationStats(integrationID)
	if err != nil {
		h.logger.Error("Failed to get notification stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// Helper functions

// ptrBoolOrDefault returns value from pointer or default if nil
func ptrBoolOrDefault(ptr *bool, defaultVal bool) bool {
	if ptr != nil {
		return *ptr
	}
	return defaultVal
}

// ptrIntOrDefault returns value from pointer or default if nil
func ptrIntOrDefault(ptr *int, defaultVal int) int {
	if ptr != nil {
		return *ptr
	}
	return defaultVal
}
