package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/anupamdutta5/monitoring-service/internal/models"
)

// DiscordIntegrationService manages Discord webhook integrations
type DiscordIntegrationService struct {
	db         *gorm.DB
	logger     *zap.Logger
	httpClient *http.Client
}

// NewDiscordIntegrationService creates a new Discord integration service
func NewDiscordIntegrationService(db *gorm.DB, logger *zap.Logger) *DiscordIntegrationService {
	return &DiscordIntegrationService{
		db:     db,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateIntegration creates a new Discord integration
func (s *DiscordIntegrationService) CreateIntegration(integration *models.DiscordIntegration) error {
	// Validate integration
	if err := integration.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Test webhook URL by sending a test message
	err := s.TestWebhook(integration.WebhookURL, integration.WebhookName)
	if err != nil {
		return fmt.Errorf("webhook test failed: %w", err)
	}

	// Create integration in database
	err = s.db.Create(integration).Error
	if err != nil {
		s.logger.Error("Failed to create Discord integration",
			zap.Error(err),
			zap.String("webhook_name", integration.WebhookName),
			zap.String("tenant_id", integration.TenantID.String()))
		return fmt.Errorf("failed to create integration: %w", err)
	}

	s.logger.Info("Discord integration created",
		zap.Uint("id", integration.ID),
		zap.String("webhook_name", integration.WebhookName),
		zap.String("tenant_id", integration.TenantID.String()))

	return nil
}

// UpdateIntegration updates an existing Discord integration
func (s *DiscordIntegrationService) UpdateIntegration(integration *models.DiscordIntegration) error {
	// Validate integration
	if err := integration.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	err := s.db.Save(integration).Error
	if err != nil {
		s.logger.Error("Failed to update Discord integration",
			zap.Error(err),
			zap.Uint("id", integration.ID))
		return fmt.Errorf("failed to update integration: %w", err)
	}

	s.logger.Info("Discord integration updated",
		zap.Uint("id", integration.ID),
		zap.String("webhook_name", integration.WebhookName))

	return nil
}

// GetIntegration retrieves a Discord integration by ID
func (s *DiscordIntegrationService) GetIntegration(id uint, tenantID uuid.UUID) (*models.DiscordIntegration, error) {
	var integration models.DiscordIntegration
	err := s.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).First(&integration).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("integration not found")
		}
		return nil, fmt.Errorf("failed to get integration: %w", err)
	}
	return &integration, nil
}

// GetIntegrationsByTenant retrieves all active integrations for a tenant
func (s *DiscordIntegrationService) GetIntegrationsByTenant(tenantID uuid.UUID) ([]models.DiscordIntegration, error) {
	var integrations []models.DiscordIntegration
	err := s.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("created_at DESC").
		Find(&integrations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}
	return integrations, nil
}

// GetActiveIntegrationsByTenant retrieves all active and enabled integrations for a tenant
func (s *DiscordIntegrationService) GetActiveIntegrationsByTenant(tenantID uuid.UUID) ([]models.DiscordIntegration, error) {
	var integrations []models.DiscordIntegration
	err := s.db.Where("tenant_id = ? AND is_active = ? AND deleted_at IS NULL", tenantID, true).
		Order("created_at DESC").
		Find(&integrations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get active integrations: %w", err)
	}
	return integrations, nil
}

// DeleteIntegration soft deletes a Discord integration
func (s *DiscordIntegrationService) DeleteIntegration(id uint, tenantID uuid.UUID) error {
	result := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.DiscordIntegration{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete integration: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("integration not found")
	}

	s.logger.Info("Discord integration deleted",
		zap.Uint("id", id),
		zap.String("tenant_id", tenantID.String()))

	return nil
}

// SubscribeChannel subscribes a channel to monitor notifications
func (s *DiscordIntegrationService) SubscribeChannel(subscription *models.DiscordChannelSubscription) error {
	// Validate subscription
	if err := subscription.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	err := s.db.Create(subscription).Error
	if err != nil {
		s.logger.Error("Failed to create channel subscription",
			zap.Error(err),
			zap.Uint("integration_id", subscription.IntegrationID))
		return fmt.Errorf("failed to subscribe channel: %w", err)
	}

	s.logger.Info("Discord channel subscribed to monitor",
		zap.Uint("integration_id", subscription.IntegrationID),
		zap.Uint("monitor_id", *subscription.MonitorID),
		zap.String("channel_id", subscription.ChannelID))

	return nil
}

// UnsubscribeChannel unsubscribes a channel from monitor notifications
func (s *DiscordIntegrationService) UnsubscribeChannel(id uint) error {
	result := s.db.Delete(&models.DiscordChannelSubscription{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to unsubscribe channel: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription not found")
	}

	s.logger.Info("Discord channel unsubscribed", zap.Uint("id", id))
	return nil
}

// GetChannelSubscriptions retrieves all active subscriptions for a monitor
func (s *DiscordIntegrationService) GetChannelSubscriptions(monitorID uint) ([]models.DiscordChannelSubscription, error) {
	var subscriptions []models.DiscordChannelSubscription
	err := s.db.Where("monitor_id = ? AND is_active = ?", monitorID, true).
		Preload("Integration").
		Find(&subscriptions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}
	return subscriptions, nil
}

// SendMonitorAlert sends a monitor alert to Discord
func (s *DiscordIntegrationService) SendMonitorAlert(
	ctx context.Context,
	monitorID uint,
	tenantID uuid.UUID,
	eventType string,
	monitorName string,
	monitorURL string,
	details map[string]interface{},
) error {
	// Get active integrations for the tenant
	integrations, err := s.GetActiveIntegrationsByTenant(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get integrations: %w", err)
	}

	if len(integrations) == 0 {
		s.logger.Debug("No active Discord integrations found for tenant",
			zap.String("tenant_id", tenantID.String()))
		return nil
	}

	// Filter integrations based on event type notification preferences
	var notifyIntegrations []models.DiscordIntegration
	for _, integration := range integrations {
		if integration.ShouldNotify(eventType) {
			notifyIntegrations = append(notifyIntegrations, integration)
		}
	}

	if len(notifyIntegrations) == 0 {
		s.logger.Debug("No integrations configured to notify for event type",
			zap.String("event_type", eventType),
			zap.String("tenant_id", tenantID.String()))
		return nil
	}

	// Send notifications to each integration
	for _, integration := range notifyIntegrations {
		err := s.sendNotification(ctx, &integration, monitorID, eventType, monitorName, monitorURL, details)
		if err != nil {
			s.logger.Error("Failed to send Discord notification",
				zap.Error(err),
				zap.Uint("integration_id", integration.ID),
				zap.String("event_type", eventType))
			// Continue to other integrations even if one fails
		}
	}

	return nil
}

// sendNotification sends a single notification to a Discord integration
func (s *DiscordIntegrationService) sendNotification(
	ctx context.Context,
	integration *models.DiscordIntegration,
	monitorID uint,
	eventType string,
	monitorName string,
	monitorURL string,
	details map[string]interface{},
) error {
	// Build Discord message
	payload := s.buildMessage(integration, eventType, monitorName, monitorURL, details)

	// Send webhook request
	response, err := s.sendWebhook(integration.WebhookURL, payload)

	// Create notification log
	notification := &models.DiscordNotification{
		IntegrationID: integration.ID,
		MonitorID:     &monitorID,
		EventType:     eventType,
		MonitorName:   monitorName,
		MonitorURL:    monitorURL,
		MessageContent: payload.Content,
		Status:        models.DiscordStatusPending,
	}

	// Serialize embed data
	if len(payload.Embeds) > 0 {
		embedJSON, _ := json.Marshal(payload.Embeds)
		notification.EmbedData = embedJSON
	}

	if err != nil {
		// Mark as failed
		notification.Status = models.DiscordStatusFailed
		notification.ErrorMessage = err.Error()
		now := time.Now()
		notification.FailedAt = &now
	} else {
		// Mark as sent
		notification.Status = models.DiscordStatusSent
		now := time.Now()
		notification.SentAt = &now
		notification.DeliveredAt = &now

		// Parse Discord response
		if response != nil {
			notification.DiscordMessageID = response.ID
			notification.DiscordChannelID = response.ChannelID
			notification.HTTPStatusCode = new(int)
			*notification.HTTPStatusCode = 200
		}
	}

	// Save notification log
	if err := s.db.Create(notification).Error; err != nil {
		s.logger.Error("Failed to save notification log",
			zap.Error(err),
			zap.Uint("integration_id", integration.ID))
	}

	// Update integration last used time
	now := time.Now()
	integration.LastUsedAt = &now
	s.db.Model(integration).Update("last_used_at", now)

	return err
}

// buildMessage builds a Discord webhook payload for a monitor alert
func (s *DiscordIntegrationService) buildMessage(
	integration *models.DiscordIntegration,
	eventType string,
	monitorName string,
	monitorURL string,
	details map[string]interface{},
) *models.DiscordWebhookPayload {
	// Build embed
	embed := models.DiscordEmbed{
		Title:       s.getEventTitle(eventType, monitorName),
		Description: s.getEventDescription(eventType, details),
		Color:       integration.GetCustomColorDecimal(eventType),
		Fields:      s.buildEmbedFields(eventType, details),
	}

	// Add monitor URL if enabled
	if integration.IncludeMonitorURL && monitorURL != "" {
		embed.URL = monitorURL
	}

	// Add timestamp if enabled
	if integration.IncludeTimestamp {
		embed.Timestamp = time.Now().Format(time.RFC3339)
	}

	// Add footer
	embed.Footer = &models.DiscordEmbedFooter{
		Text: "Beakon Status Monitor",
	}

	// Build payload
	payload := &models.DiscordWebhookPayload{
		Embeds: []models.DiscordEmbed{embed},
	}

	// Add username override if set
	if integration.WebhookName != "" {
		payload.Username = integration.WebhookName
	}

	// Add avatar URL if set
	if integration.AvatarURL != "" {
		payload.AvatarURL = integration.AvatarURL
	}

	// Add mentions
	if len(integration.MentionUsers) > 0 || len(integration.MentionRoles) > 0 || integration.MentionEveryone {
		content := s.buildMentions(integration)
		payload.Content = content
	}

	return payload
}

// buildMentions builds the mention string for Discord
func (s *DiscordIntegrationService) buildMentions(integration *models.DiscordIntegration) string {
	var mentions []string

	// Add @everyone if enabled
	if integration.MentionEveryone {
		mentions = append(mentions, "@everyone")
	}

	// Add role mentions
	for _, roleID := range integration.MentionRoles {
		if roleID != "" {
			mentions = append(mentions, fmt.Sprintf("<@&%s>", roleID))
		}
	}

	// Add user mentions
	for _, userID := range integration.MentionUsers {
		if userID != "" {
			mentions = append(mentions, fmt.Sprintf("<@%s>", userID))
		}
	}

	// Join all mentions
	content := ""
	for i, mention := range mentions {
		if i > 0 {
			content += " "
		}
		content += mention
	}

	return content
}

// getEventTitle returns the embed title for an event type
func (s *DiscordIntegrationService) getEventTitle(eventType string, monitorName string) string {
	switch eventType {
	case models.DiscordEventDown:
		return fmt.Sprintf("🔴 Monitor Down: %s", monitorName)
	case models.DiscordEventUp:
		return fmt.Sprintf("🟢 Monitor Recovered: %s", monitorName)
	case models.DiscordEventDegraded:
		return fmt.Sprintf("🟡 Monitor Degraded: %s", monitorName)
	case models.DiscordEventMaintenance:
		return fmt.Sprintf("🔵 Maintenance Scheduled: %s", monitorName)
	default:
		return fmt.Sprintf("Monitor Alert: %s", monitorName)
	}
}

// getEventDescription returns the embed description for an event type
func (s *DiscordIntegrationService) getEventDescription(eventType string, details map[string]interface{}) string {
	switch eventType {
	case models.DiscordEventDown:
		if msg, ok := details["error_message"].(string); ok {
			return fmt.Sprintf("The monitor is currently down. Error: %s", msg)
		}
		return "The monitor is currently experiencing issues and is down."
	case models.DiscordEventUp:
		return "The monitor has recovered and is now operational."
	case models.DiscordEventDegraded:
		return "The monitor is experiencing degraded performance."
	case models.DiscordEventMaintenance:
		return "Scheduled maintenance is in progress."
	default:
		return "Monitor status has changed."
	}
}

// buildEmbedFields builds embed fields from details
func (s *DiscordIntegrationService) buildEmbedFields(eventType string, details map[string]interface{}) []models.DiscordEmbedField {
	var fields []models.DiscordEmbedField

	// Add status field
	fields = append(fields, models.DiscordEmbedField{
		Name:   "Status",
		Value:  s.formatEventType(eventType),
		Inline: true,
	})

	// Add timestamp
	fields = append(fields, models.DiscordEmbedField{
		Name:   "Time",
		Value:  time.Now().Format("2006-01-02 15:04:05 MST"),
		Inline: true,
	})

	// Add response time if available
	if responseTime, ok := details["response_time"].(float64); ok {
		fields = append(fields, models.DiscordEmbedField{
			Name:   "Response Time",
			Value:  fmt.Sprintf("%.0f ms", responseTime),
			Inline: true,
		})
	}

	// Add HTTP status code if available
	if statusCode, ok := details["status_code"].(int); ok {
		fields = append(fields, models.DiscordEmbedField{
			Name:   "HTTP Status",
			Value:  fmt.Sprintf("%d", statusCode),
			Inline: true,
		})
	}

	// Add location if available
	if location, ok := details["location"].(string); ok {
		fields = append(fields, models.DiscordEmbedField{
			Name:   "Location",
			Value:  location,
			Inline: true,
		})
	}

	return fields
}

// formatEventType formats event type for display
func (s *DiscordIntegrationService) formatEventType(eventType string) string {
	switch eventType {
	case models.DiscordEventDown:
		return "Down"
	case models.DiscordEventUp:
		return "Up"
	case models.DiscordEventDegraded:
		return "Degraded"
	case models.DiscordEventMaintenance:
		return "Maintenance"
	default:
		return eventType
	}
}

// sendWebhook sends a webhook request to Discord
func (s *DiscordIntegrationService) sendWebhook(webhookURL string, payload *models.DiscordWebhookPayload) (*models.DiscordWebhookResponse, error) {
	// Marshal payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Beakon-Monitor/1.0")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("webhook request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response if available
	if len(body) > 0 {
		var discordResponse models.DiscordWebhookResponse
		if err := json.Unmarshal(body, &discordResponse); err != nil {
			s.logger.Warn("Failed to parse Discord response", zap.Error(err))
			// Continue anyway as the message was sent
		} else {
			return &discordResponse, nil
		}
	}

	return nil, nil
}

// TestWebhook tests a Discord webhook URL by sending a test message
func (s *DiscordIntegrationService) TestWebhook(webhookURL string, webhookName string) error {
	// Build test message
	payload := &models.DiscordWebhookPayload{
		Content: "✅ Beakon Discord integration test successful!",
		Embeds: []models.DiscordEmbed{
			{
				Title:       "Integration Test",
				Description: "This is a test message to verify your Discord webhook integration is working correctly.",
				Color:       models.DiscordColorGreen,
				Timestamp:   time.Now().Format(time.RFC3339),
				Footer: &models.DiscordEmbedFooter{
					Text: "Beakon Status Monitor",
				},
			},
		},
	}

	if webhookName != "" {
		payload.Username = webhookName
	}

	// Send test webhook
	_, err := s.sendWebhook(webhookURL, payload)
	if err != nil {
		return fmt.Errorf("webhook test failed: %w", err)
	}

	s.logger.Info("Discord webhook test successful", zap.String("webhook_url", webhookURL[:50]+"..."))
	return nil
}

// GetNotificationHistory retrieves notification history for an integration
func (s *DiscordIntegrationService) GetNotificationHistory(integrationID uint, limit int) ([]models.DiscordNotification, error) {
	var notifications []models.DiscordNotification
	query := s.db.Where("integration_id = ?", integrationID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&notifications).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get notification history: %w", err)
	}

	return notifications, nil
}

// GetNotificationStats retrieves notification statistics for an integration
func (s *DiscordIntegrationService) GetNotificationStats(integrationID uint) (map[string]interface{}, error) {
	var stats struct {
		TotalNotifications      int64
		SuccessfulNotifications int64
		FailedNotifications     int64
		DownNotifications       int64
		UpNotifications         int64
	}

	// Count total notifications
	s.db.Model(&models.DiscordNotification{}).
		Where("integration_id = ?", integrationID).
		Count(&stats.TotalNotifications)

	// Count successful notifications
	s.db.Model(&models.DiscordNotification{}).
		Where("integration_id = ? AND status = ?", integrationID, models.DiscordStatusSent).
		Count(&stats.SuccessfulNotifications)

	// Count failed notifications
	s.db.Model(&models.DiscordNotification{}).
		Where("integration_id = ? AND status = ?", integrationID, models.DiscordStatusFailed).
		Count(&stats.FailedNotifications)

	// Count down notifications
	s.db.Model(&models.DiscordNotification{}).
		Where("integration_id = ? AND event_type = ?", integrationID, models.DiscordEventDown).
		Count(&stats.DownNotifications)

	// Count up notifications
	s.db.Model(&models.DiscordNotification{}).
		Where("integration_id = ? AND event_type = ?", integrationID, models.DiscordEventUp).
		Count(&stats.UpNotifications)

	return map[string]interface{}{
		"total_notifications":      stats.TotalNotifications,
		"successful_notifications": stats.SuccessfulNotifications,
		"failed_notifications":     stats.FailedNotifications,
		"down_notifications":       stats.DownNotifications,
		"up_notifications":         stats.UpNotifications,
		"success_rate":             float64(stats.SuccessfulNotifications) / float64(stats.TotalNotifications) * 100,
	}, nil
}
