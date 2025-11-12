package services

import (
	resilience "github.com/anupamdutta5/shared-resilience"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/anupamdutta5/monitoring-service/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TelegramIntegrationService handles Telegram bot integrations and notifications
type TelegramIntegrationService struct {
	db         *gorm.DB
	logger     *zap.Logger
	serviceClient *resilience.ServiceClient
}

// NewTelegramIntegrationService creates a new Telegram integration service
func NewTelegramIntegrationService(db *gorm.DB, serviceClient *resilience.ServiceClient, logger *zap.Logger) *TelegramIntegrationService {
	return &TelegramIntegrationService{
		db:     db,
		logger: logger,
		serviceClient: serviceClient,
	}
}

// CreateIntegration creates a new Telegram integration
func (s *TelegramIntegrationService) CreateIntegration(integration *models.TelegramIntegration) error {
	// Validate integration
	if err := integration.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Test bot token by calling getMe API
	botInfo, err := s.GetBotInfo(integration.BotToken)
	if err != nil {
		return fmt.Errorf("bot token validation failed: %w", err)
	}

	// Auto-populate bot info
	integration.BotUsername = botInfo.Username
	integration.BotName = fmt.Sprintf("%s %s", botInfo.FirstName, botInfo.LastName)

	// Create integration in database
	if err := s.db.Create(integration).Error; err != nil {
		return fmt.Errorf("failed to create integration: %w", err)
	}

	s.logger.Info("Telegram integration created",
		zap.Uint("id", integration.ID),
		zap.String("tenant_id", integration.TenantID.String()),
		zap.String("bot_username", integration.BotUsername),
	)

	return nil
}

// UpdateIntegration updates an existing Telegram integration
func (s *TelegramIntegrationService) UpdateIntegration(integration *models.TelegramIntegration) error {
	// Validate integration
	if err := integration.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// If bot token changed, validate it
	var existing models.TelegramIntegration
	if err := s.db.First(&existing, integration.ID).Error; err != nil {
		return fmt.Errorf("integration not found: %w", err)
	}

	if existing.BotToken != integration.BotToken {
		botInfo, err := s.GetBotInfo(integration.BotToken)
		if err != nil {
			return fmt.Errorf("bot token validation failed: %w", err)
		}
		integration.BotUsername = botInfo.Username
		integration.BotName = fmt.Sprintf("%s %s", botInfo.FirstName, botInfo.LastName)
	}

	// Update integration
	if err := s.db.Save(integration).Error; err != nil {
		return fmt.Errorf("failed to update integration: %w", err)
	}

	s.logger.Info("Telegram integration updated",
		zap.Uint("id", integration.ID),
		zap.String("tenant_id", integration.TenantID.String()),
	)

	return nil
}

// GetIntegration retrieves a Telegram integration by ID and tenant
func (s *TelegramIntegrationService) GetIntegration(id uint, tenantID uuid.UUID) (*models.TelegramIntegration, error) {
	var integration models.TelegramIntegration
	err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&integration).Error
	if err != nil {
		return nil, err
	}
	return &integration, nil
}

// GetIntegrationsByTenant retrieves all Telegram integrations for a tenant
func (s *TelegramIntegrationService) GetIntegrationsByTenant(tenantID uuid.UUID) ([]models.TelegramIntegration, error) {
	var integrations []models.TelegramIntegration
	err := s.db.Where("tenant_id = ?", tenantID).Find(&integrations).Error
	return integrations, err
}

// GetActiveIntegrationsByTenant retrieves all active Telegram integrations for a tenant
func (s *TelegramIntegrationService) GetActiveIntegrationsByTenant(tenantID uuid.UUID) ([]models.TelegramIntegration, error) {
	var integrations []models.TelegramIntegration
	err := s.db.Where("tenant_id = ? AND is_active = ?", tenantID, true).Find(&integrations).Error
	return integrations, err
}

// DeleteIntegration soft deletes a Telegram integration
func (s *TelegramIntegrationService) DeleteIntegration(id uint, tenantID uuid.UUID) error {
	result := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.TelegramIntegration{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("integration not found")
	}

	s.logger.Info("Telegram integration deleted",
		zap.Uint("id", id),
		zap.String("tenant_id", tenantID.String()),
	)

	return nil
}

// SubscribeChat subscribes a Telegram chat to a monitor
func (s *TelegramIntegrationService) SubscribeChat(subscription *models.TelegramChatSubscription) error {
	// Verify integration exists
	var integration models.TelegramIntegration
	if err := s.db.First(&integration, subscription.IntegrationID).Error; err != nil {
		return fmt.Errorf("integration not found: %w", err)
	}

	// Create subscription
	if err := s.db.Create(subscription).Error; err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	s.logger.Info("Telegram chat subscribed",
		zap.Uint("integration_id", subscription.IntegrationID),
		zap.Uint("monitor_id", *subscription.MonitorID),
		zap.String("chat_id", subscription.ChatID),
	)

	return nil
}

// UnsubscribeChat removes a Telegram chat subscription
func (s *TelegramIntegrationService) UnsubscribeChat(subscriptionID uint) error {
	result := s.db.Delete(&models.TelegramChatSubscription{}, subscriptionID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription not found")
	}

	s.logger.Info("Telegram chat unsubscribed", zap.Uint("subscription_id", subscriptionID))
	return nil
}

// GetChatSubscriptions retrieves all chat subscriptions for an integration
func (s *TelegramIntegrationService) GetChatSubscriptions(integrationID uint) ([]models.TelegramChatSubscription, error) {
	var subscriptions []models.TelegramChatSubscription
	err := s.db.Where("integration_id = ?", integrationID).Find(&subscriptions).Error
	return subscriptions, err
}

// SendMonitorAlert sends a monitor alert to Telegram
func (s *TelegramIntegrationService) SendMonitorAlert(
	ctx context.Context,
	monitorID uint,
	tenantID uuid.UUID,
	eventType string,
	monitorName string,
	monitorURL string,
) error {
	// Get active integrations for tenant
	integrations, err := s.GetActiveIntegrationsByTenant(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get integrations: %w", err)
	}

	if len(integrations) == 0 {
		s.logger.Debug("No active Telegram integrations found for tenant",
			zap.String("tenant_id", tenantID.String()),
		)
		return nil
	}

	// Send alert to each integration
	for _, integration := range integrations {
		// Check if integration wants this event type
		if !s.shouldNotifyForEvent(&integration, eventType) {
			continue
		}

		// Check for chat subscriptions
		var targetChatID string
		var customPrefix string
		var threadID *int

		subscriptions, _ := s.GetChatSubscriptions(integration.ID)
		hasSubscription := false

		for _, sub := range subscriptions {
			// Check if subscription matches monitor
			if sub.MonitorID != nil && *sub.MonitorID == monitorID && sub.IsActive {
				// Check if subscription wants this event type
				if s.shouldNotifyForEventSubscription(&sub, eventType) {
					targetChatID = sub.ChatID
					customPrefix = sub.CustomMessagePrefix
					threadID = sub.MessageThreadID
					hasSubscription = true
					break
				}
			}
		}

		// If no specific subscription, use default chat
		if !hasSubscription && integration.DefaultChatID != "" {
			targetChatID = integration.DefaultChatID
		}

		if targetChatID == "" {
			s.logger.Warn("No target chat ID for Telegram notification",
				zap.Uint("integration_id", integration.ID),
				zap.Uint("monitor_id", monitorID),
			)
			continue
		}

		// Send notification
		err := s.sendNotification(
			ctx,
			&integration,
			targetChatID,
			eventType,
			monitorName,
			monitorURL,
			customPrefix,
			threadID,
			monitorID,
		)

		if err != nil {
			s.logger.Error("Failed to send Telegram notification",
				zap.Uint("integration_id", integration.ID),
				zap.String("chat_id", targetChatID),
				zap.Error(err),
			)
		}
	}

	return nil
}

// sendNotification sends a single Telegram notification
func (s *TelegramIntegrationService) sendNotification(
	ctx context.Context,
	integration *models.TelegramIntegration,
	chatID string,
	eventType string,
	monitorName string,
	monitorURL string,
	customPrefix string,
	threadID *int,
	monitorID uint,
) error {
	// Build message text
	message := models.BuildTelegramMessage(
		eventType,
		monitorName,
		monitorURL,
		nil, // details map
		integration.UseMarkdown,
		integration.IncludeMonitorURL,
		integration.IncludeTimestamp,
	)

	// Prepend custom prefix if provided
	if customPrefix != "" {
		message = customPrefix + "\n\n" + message
	}

	// Create notification record
	notification := &models.TelegramNotification{
		IntegrationID: integration.ID,
		MonitorID:     &monitorID,
		EventType:     eventType,
		MonitorName:   monitorName,
		MonitorURL:    monitorURL,
		Status:        "pending",
		MessageText:   message,
		TelegramChatID: chatID,
	}

	if integration.UseMarkdown {
		notification.ParseMode = "MarkdownV2"
	}

	// Save notification record
	if err := s.db.Create(notification).Error; err != nil {
		return fmt.Errorf("failed to create notification record: %w", err)
	}

	// Send message via Telegram Bot API
	messageID, statusCode, err := s.SendMessage(
		integration.BotToken,
		chatID,
		message,
		integration.UseMarkdown,
		integration.DisablePreview,
		integration.SilentNotifications,
		threadID,
	)

	// Update notification record with result
	now := time.Now()
	notification.HTTPStatusCode = &statusCode

	if err != nil {
		notification.Status = "failed"
		notification.FailedAt = &now
		notification.ErrorMessage = err.Error()
	} else {
		notification.Status = "sent"
		notification.SentAt = &now
		notification.DeliveredAt = &now
		notification.TelegramMessageID = messageID
	}

	// Save updated notification
	if updateErr := s.db.Save(notification).Error; updateErr != nil {
		s.logger.Error("Failed to update notification record", zap.Error(updateErr))
	}

	// Update integration last_used_at
	integration.LastUsedAt = &now
	if updateErr := s.db.Model(&models.TelegramIntegration{}).Where("id = ?", integration.ID).Update("last_used_at", now).Error; updateErr != nil {
		s.logger.Error("Failed to update integration last_used_at", zap.Error(updateErr))
	}

	return err
}

// shouldNotifyForEvent checks if integration wants notifications for event type
func (s *TelegramIntegrationService) shouldNotifyForEvent(integration *models.TelegramIntegration, eventType string) bool {
	switch eventType {
	case "down":
		return integration.NotifyOnDown
	case "up":
		return integration.NotifyOnUp
	case "degraded":
		return integration.NotifyOnDegraded
	case "maintenance":
		return integration.NotifyOnMaintenance
	default:
		return false
	}
}

// shouldNotifyForEventSubscription checks if subscription wants notifications for event type
func (s *TelegramIntegrationService) shouldNotifyForEventSubscription(subscription *models.TelegramChatSubscription, eventType string) bool {
	switch eventType {
	case "down":
		return subscription.NotifyOnDown
	case "up":
		return subscription.NotifyOnUp
	case "degraded":
		return subscription.NotifyOnDegraded
	case "maintenance":
		return subscription.NotifyOnMaintenance
	default:
		return false
	}
}

// SendMessage sends a message via Telegram Bot API
func (s *TelegramIntegrationService) SendMessage(
	botToken string,
	chatID string,
	text string,
	useMarkdown bool,
	disablePreview bool,
	silent bool,
	threadID *int,
) (int64, int, error) {
	// Build API URL
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	// Build request payload
	payload := models.TelegramSendMessageRequest{
		ChatID:                chatID,
		Text:                  text,
		DisableWebPagePreview: disablePreview,
		DisableNotification:   silent,
		MessageThreadID:       threadID,
	}

	if useMarkdown {
		payload.ParseMode = "MarkdownV2"
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request

	// Send request using HTTP client
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var telegramResp models.TelegramSendMessageResponse
	if err := json.Unmarshal(body, &telegramResp); err != nil {
		return 0, resp.StatusCode, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if successful
	if !telegramResp.OK {
		return 0, telegramResp.ErrorCode, fmt.Errorf("telegram API error: %s (code: %d)", telegramResp.Description, telegramResp.ErrorCode)
	}

	if telegramResp.Result == nil {
		return 0, resp.StatusCode, fmt.Errorf("telegram API returned no result")
	}

	s.logger.Info("Telegram message sent successfully",
		zap.String("chat_id", chatID),
		zap.Int64("message_id", telegramResp.Result.MessageID),
	)

	return telegramResp.Result.MessageID, resp.StatusCode, nil
}

// GetBotInfo retrieves bot information using the getMe API
func (s *TelegramIntegrationService) GetBotInfo(botToken string) (*models.TelegramUser, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", botToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get bot info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var telegramResp models.TelegramGetMeResponse
	if err := json.Unmarshal(body, &telegramResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !telegramResp.OK {
		return nil, fmt.Errorf("telegram API error: %s (code: %d)", telegramResp.Description, telegramResp.ErrorCode)
	}

	if telegramResp.Result == nil {
		return nil, fmt.Errorf("telegram API returned no result")
	}

	return telegramResp.Result, nil
}

// TestBot tests a bot token by sending a test message
func (s *TelegramIntegrationService) TestBot(botToken string, chatID string) error {
	message := "✅ *Beakon Telegram Integration Test*\n\n"
	message += "Your Telegram bot is configured correctly and can send notifications\\.\n\n"
	message += "You will receive monitor alerts in this chat\\."

	_, _, err := s.SendMessage(botToken, chatID, message, true, false, false, nil)
	return err
}

// GetNotificationHistory retrieves notification history for an integration
func (s *TelegramIntegrationService) GetNotificationHistory(integrationID uint, limit int) ([]models.TelegramNotification, error) {
	var notifications []models.TelegramNotification
	err := s.db.Where("integration_id = ?", integrationID).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

// GetNotificationStats retrieves statistics for an integration
func (s *TelegramIntegrationService) GetNotificationStats(integrationID uint) (map[string]interface{}, error) {
	var stats struct {
		Total      int64
		Sent       int64
		Failed     int64
		Pending    int64
		Retrying   int64
		Down       int64
		Up         int64
		Degraded   int64
		Maintenance int64
	}

	// Total notifications
	s.db.Model(&models.TelegramNotification{}).Where("integration_id = ?", integrationID).Count(&stats.Total)

	// By status
	s.db.Model(&models.TelegramNotification{}).Where("integration_id = ? AND status = ?", integrationID, "sent").Count(&stats.Sent)
	s.db.Model(&models.TelegramNotification{}).Where("integration_id = ? AND status = ?", integrationID, "failed").Count(&stats.Failed)
	s.db.Model(&models.TelegramNotification{}).Where("integration_id = ? AND status = ?", integrationID, "pending").Count(&stats.Pending)
	s.db.Model(&models.TelegramNotification{}).Where("integration_id = ? AND status = ?", integrationID, "retrying").Count(&stats.Retrying)

	// By event type
	s.db.Model(&models.TelegramNotification{}).Where("integration_id = ? AND event_type = ?", integrationID, "down").Count(&stats.Down)
	s.db.Model(&models.TelegramNotification{}).Where("integration_id = ? AND event_type = ?", integrationID, "up").Count(&stats.Up)
	s.db.Model(&models.TelegramNotification{}).Where("integration_id = ? AND event_type = ?", integrationID, "degraded").Count(&stats.Degraded)
	s.db.Model(&models.TelegramNotification{}).Where("integration_id = ? AND event_type = ?", integrationID, "maintenance").Count(&stats.Maintenance)

	// Calculate success rate
	successRate := 0.0
	if stats.Total > 0 {
		successRate = (float64(stats.Sent) / float64(stats.Total)) * 100
	}

	// Get last notification time
	var lastNotification models.TelegramNotification
	s.db.Where("integration_id = ?", integrationID).Order("sent_at DESC").First(&lastNotification)

	return map[string]interface{}{
		"total_notifications":       stats.Total,
		"successful_notifications":  stats.Sent,
		"failed_notifications":      stats.Failed,
		"pending_notifications":     stats.Pending,
		"retrying_notifications":    stats.Retrying,
		"success_rate":              successRate,
		"down_notifications":        stats.Down,
		"up_notifications":          stats.Up,
		"degraded_notifications":    stats.Degraded,
		"maintenance_notifications": stats.Maintenance,
		"last_notification_at":      lastNotification.SentAt,
	}, nil
}
