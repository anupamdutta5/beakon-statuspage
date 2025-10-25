package discord

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/anupamdutta5/notification-service/internal/models"
	"github.com/anupamdutta5/notification-service/internal/services"
)

// DiscordHandler handles Discord integration endpoints
type DiscordHandler struct {
	db             *gorm.DB
	logger         *zap.Logger
	discordService *services.DiscordIntegrationService
}

// NewDiscordHandler creates a new Discord handler
func NewDiscordHandler(db *gorm.DB, logger *zap.Logger, discordService *services.DiscordIntegrationService) *DiscordHandler {
	return &DiscordHandler{
		db:             db,
		logger:         logger,
		discordService: discordService,
	}
}

// CreateIntegrationRequest represents the request body for creating a Discord integration
type CreateIntegrationRequest struct {
	WebhookURL        string   `json:"webhook_url" binding:"required"`
	WebhookName       string   `json:"webhook_name"`
	AvatarURL         string   `json:"avatar_url"`
	DefaultChannelID  string   `json:"default_channel_id"`
	DefaultChannelName string  `json:"default_channel_name"`
	NotifyOnDown      *bool    `json:"notify_on_down"`
	NotifyOnUp        *bool    `json:"notify_on_up"`
	NotifyOnDegraded  *bool    `json:"notify_on_degraded"`
	NotifyOnMaintenance *bool  `json:"notify_on_maintenance"`
	MentionUsers      []string `json:"mention_users"`
	MentionRoles      []string `json:"mention_roles"`
	MentionEveryone   *bool    `json:"mention_everyone"`
	CustomColor       string   `json:"custom_color"`
	IncludeMonitorURL *bool    `json:"include_monitor_url"`
	IncludeTimestamp  *bool    `json:"include_timestamp"`
	RetryCount        *int     `json:"retry_count"`
	RetryIntervalSeconds *int  `json:"retry_interval_seconds"`
}

// UpdateIntegrationRequest represents the request body for updating a Discord integration
type UpdateIntegrationRequest struct {
	WebhookURL        *string  `json:"webhook_url"`
	WebhookName       *string  `json:"webhook_name"`
	AvatarURL         *string  `json:"avatar_url"`
	DefaultChannelID  *string  `json:"default_channel_id"`
	DefaultChannelName *string `json:"default_channel_name"`
	IsActive          *bool    `json:"is_active"`
	NotifyOnDown      *bool    `json:"notify_on_down"`
	NotifyOnUp        *bool    `json:"notify_on_up"`
	NotifyOnDegraded  *bool    `json:"notify_on_degraded"`
	NotifyOnMaintenance *bool  `json:"notify_on_maintenance"`
	MentionUsers      []string `json:"mention_users"`
	MentionRoles      []string `json:"mention_roles"`
	MentionEveryone   *bool    `json:"mention_everyone"`
	CustomColor       *string  `json:"custom_color"`
	IncludeMonitorURL *bool    `json:"include_monitor_url"`
	IncludeTimestamp  *bool    `json:"include_timestamp"`
	RetryCount        *int     `json:"retry_count"`
	RetryIntervalSeconds *int  `json:"retry_interval_seconds"`
}

// SubscribeChannelRequest represents the request to subscribe a channel to a monitor
type SubscribeChannelRequest struct {
	MonitorID        uint     `json:"monitor_id" binding:"required"`
	ChannelID        string   `json:"channel_id"`
	ChannelName      string   `json:"channel_name"`
	NotifyOnDown     *bool    `json:"notify_on_down"`
	NotifyOnUp       *bool    `json:"notify_on_up"`
	NotifyOnDegraded *bool    `json:"notify_on_degraded"`
	NotifyOnMaintenance *bool `json:"notify_on_maintenance"`
}

// CreateIntegration creates a new Discord integration
// POST /api/v1/integrations/discord
func (h *DiscordHandler) CreateIntegration(c *gin.Context) {
	var req CreateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	// Get tenant_id from context (set by authentication middleware)
	tenantIDStr := c.GetString("tenant_id")
	if tenantIDStr == "" {
		// Fallback to query parameter for testing
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
	integration := &models.DiscordIntegration{
		TenantID:            tenantID,
		WebhookURL:          req.WebhookURL,
		WebhookName:         req.WebhookName,
		AvatarURL:           req.AvatarURL,
		DefaultChannelID:    req.DefaultChannelID,
		DefaultChannelName:  req.DefaultChannelName,
		IsActive:            true,
		NotifyOnDown:        boolValue(req.NotifyOnDown, true),
		NotifyOnUp:          boolValue(req.NotifyOnUp, true),
		NotifyOnDegraded:    boolValue(req.NotifyOnDegraded, true),
		NotifyOnMaintenance: boolValue(req.NotifyOnMaintenance, false),
		MentionUsers:        req.MentionUsers,
		MentionRoles:        req.MentionRoles,
		MentionEveryone:     boolValue(req.MentionEveryone, false),
		CustomColor:         req.CustomColor,
		IncludeMonitorURL:   boolValue(req.IncludeMonitorURL, true),
		IncludeTimestamp:    boolValue(req.IncludeTimestamp, true),
		RetryCount:          intValue(req.RetryCount, 3),
		RetryIntervalSeconds: intValue(req.RetryIntervalSeconds, 5),
	}

	// Create integration (includes webhook validation)
	if err := h.discordService.CreateIntegration(integration); err != nil {
		h.logger.Error("Failed to create Discord integration",
			zap.String("tenant_id", tenantID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create integration", "details": err.Error()})
		return
	}

	h.logger.Info("Discord integration created successfully",
		zap.Uint("id", integration.ID),
		zap.String("tenant_id", tenantID.String()),
		zap.String("webhook_name", integration.WebhookName),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "integration created successfully",
		"integration": integration,
	})
}

// UpdateIntegration updates an existing Discord integration
// PUT /api/v1/integrations/discord/:id
func (h *DiscordHandler) UpdateIntegration(c *gin.Context) {
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

	var req UpdateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	// Get existing integration
	integration, err := h.discordService.GetIntegration(id, tenantID)
	if err != nil || integration == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	// Update fields if provided
	if req.WebhookURL != nil {
		integration.WebhookURL = *req.WebhookURL
	}
	if req.WebhookName != nil {
		integration.WebhookName = *req.WebhookName
	}
	if req.AvatarURL != nil {
		integration.AvatarURL = *req.AvatarURL
	}
	if req.DefaultChannelID != nil {
		integration.DefaultChannelID = *req.DefaultChannelID
	}
	if req.DefaultChannelName != nil {
		integration.DefaultChannelName = *req.DefaultChannelName
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
	if req.MentionUsers != nil {
		integration.MentionUsers = req.MentionUsers
	}
	if req.MentionRoles != nil {
		integration.MentionRoles = req.MentionRoles
	}
	if req.MentionEveryone != nil {
		integration.MentionEveryone = *req.MentionEveryone
	}
	if req.CustomColor != nil {
		integration.CustomColor = *req.CustomColor
	}
	if req.IncludeMonitorURL != nil {
		integration.IncludeMonitorURL = *req.IncludeMonitorURL
	}
	if req.IncludeTimestamp != nil {
		integration.IncludeTimestamp = *req.IncludeTimestamp
	}
	if req.RetryCount != nil {
		integration.RetryCount = *req.RetryCount
	}
	if req.RetryIntervalSeconds != nil {
		integration.RetryIntervalSeconds = *req.RetryIntervalSeconds
	}

	// Update integration
	if err := h.discordService.UpdateIntegration(integration); err != nil {
		h.logger.Error("Failed to update Discord integration",
			zap.Uint("id", id),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update integration", "details": err.Error()})
		return
	}

	h.logger.Info("Discord integration updated successfully",
		zap.Uint("id", id),
		zap.String("tenant_id", tenantID.String()),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "integration updated successfully",
		"integration": integration,
	})
}

// GetIntegration retrieves a Discord integration by ID
// GET /api/v1/integrations/discord/:id
func (h *DiscordHandler) GetIntegration(c *gin.Context) {
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

	integration, err := h.discordService.GetIntegration(id, tenantID)
	if err != nil || integration == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"integration": integration})
}

// GetIntegrations retrieves all Discord integrations for a tenant
// GET /api/v1/integrations/discord
func (h *DiscordHandler) GetIntegrations(c *gin.Context) {
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

	integrations, err := h.discordService.GetIntegrationsByTenant(tenantID)
	if err != nil {
		h.logger.Error("Failed to get Discord integrations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve integrations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"integrations": integrations,
		"count": len(integrations),
	})
}

// DeleteIntegration soft deletes a Discord integration
// DELETE /api/v1/integrations/discord/:id
func (h *DiscordHandler) DeleteIntegration(c *gin.Context) {
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

	if err := h.discordService.DeleteIntegration(id, tenantID); err != nil {
		h.logger.Error("Failed to delete Discord integration",
			zap.Uint("id", id),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete integration"})
		return
	}

	h.logger.Info("Discord integration deleted",
		zap.Uint("id", id),
		zap.String("tenant_id", tenantID.String()),
	)

	c.JSON(http.StatusOK, gin.H{"message": "integration deleted successfully"})
}

// TestIntegration sends a test notification to Discord
// POST /api/v1/integrations/discord/:id/test
func (h *DiscordHandler) TestIntegration(c *gin.Context) {
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
	integration, err := h.discordService.GetIntegration(id, tenantID)
	if err != nil || integration == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	// Send test notification using SendMonitorAlert
	testData := map[string]interface{}{
		"status": "This is a test notification from Beakon Status Page",
		"response_time": "125ms",
		"status_code": "200",
	}

	if err := h.discordService.SendMonitorAlert(
		c.Request.Context(),
		0,
		tenantID,
		"up",
		"Test Monitor",
		"https://status.example.com",
		testData,
	); err != nil {
		h.logger.Error("Failed to send test notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send test notification", "details": err.Error()})
		return
	}

	h.logger.Info("Test Discord notification sent",
		zap.Uint("id", id),
		zap.String("tenant_id", tenantID.String()),
	)

	c.JSON(http.StatusOK, gin.H{"message": "test notification sent successfully"})
}

// SubscribeChannel subscribes a Discord channel to a monitor
// POST /api/v1/integrations/discord/:id/subscribe
func (h *DiscordHandler) SubscribeChannel(c *gin.Context) {
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

	var req SubscribeChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	// Verify integration exists and belongs to tenant
	integration, err := h.discordService.GetIntegration(integrationID, tenantID)
	if err != nil || integration == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	// Create subscription
	subscription := &models.DiscordChannelSubscription{
		IntegrationID:       integrationID,
		MonitorID:           &req.MonitorID,
		ChannelID:           req.ChannelID,
		ChannelName:         req.ChannelName,
		NotifyOnDown:        boolValue(req.NotifyOnDown, true),
		NotifyOnUp:          boolValue(req.NotifyOnUp, true),
		NotifyOnDegraded:    boolValue(req.NotifyOnDegraded, true),
		NotifyOnMaintenance: boolValue(req.NotifyOnMaintenance, false),
		IsActive:            true,
	}

	if err := h.discordService.SubscribeChannel(subscription); err != nil {
		h.logger.Error("Failed to subscribe channel",
			zap.Uint("integration_id", integrationID),
			zap.Uint("monitor_id", req.MonitorID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to subscribe channel", "details": err.Error()})
		return
	}

	h.logger.Info("Discord channel subscribed to monitor",
		zap.Uint("integration_id", integrationID),
		zap.Uint("monitor_id", req.MonitorID),
		zap.String("channel_id", req.ChannelID),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "channel subscribed successfully",
		"subscription": subscription,
	})
}

// UnsubscribeChannel removes a channel subscription
// DELETE /api/v1/integrations/discord/subscriptions/:id
func (h *DiscordHandler) UnsubscribeChannel(c *gin.Context) {
	idStr := c.Param("id")
	subscriptionID, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	if err := h.discordService.UnsubscribeChannel(subscriptionID); err != nil {
		h.logger.Error("Failed to unsubscribe channel",
			zap.Uint("subscription_id", subscriptionID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unsubscribe channel"})
		return
	}

	h.logger.Info("Discord channel unsubscribed",
		zap.Uint("subscription_id", subscriptionID),
	)

	c.JSON(http.StatusOK, gin.H{"message": "channel unsubscribed successfully"})
}

// GetChannelSubscriptions retrieves all channel subscriptions for an integration
// GET /api/v1/integrations/discord/:id/subscriptions
func (h *DiscordHandler) GetChannelSubscriptions(c *gin.Context) {
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
	integration, err := h.discordService.GetIntegration(integrationID, tenantID)
	if err != nil || integration == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	subscriptions, err := h.discordService.GetChannelSubscriptions(integrationID)
	if err != nil {
		h.logger.Error("Failed to get channel subscriptions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscriptions": subscriptions,
		"count": len(subscriptions),
	})
}

// GetNotificationHistory retrieves notification history for an integration
// GET /api/v1/integrations/discord/:id/notifications
func (h *DiscordHandler) GetNotificationHistory(c *gin.Context) {
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
	integration, err := h.discordService.GetIntegration(integrationID, tenantID)
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

	notifications, err := h.discordService.GetNotificationHistory(integrationID, limit)
	if err != nil {
		h.logger.Error("Failed to get notification history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"count": len(notifications),
	})
}

// GetNotificationStats retrieves statistics for an integration
// GET /api/v1/integrations/discord/:id/stats
func (h *DiscordHandler) GetNotificationStats(c *gin.Context) {
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
	integration, err := h.discordService.GetIntegration(integrationID, tenantID)
	if err != nil || integration == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	stats, err := h.discordService.GetNotificationStats(integrationID)
	if err != nil {
		h.logger.Error("Failed to get notification stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// Helper function to get bool value with default
func boolValue(ptr *bool, defaultVal bool) bool {
	if ptr != nil {
		return *ptr
	}
	return defaultVal
}

// Helper function to get int value with default
func intValue(ptr *int, defaultVal int) int {
	if ptr != nil {
		return *ptr
	}
	return defaultVal
}
