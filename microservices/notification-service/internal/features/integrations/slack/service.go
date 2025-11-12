package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	resilience "github.com/anupamdutta5/shared-resilience"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SlackIntegration represents a Slack workspace integration configuration
type SlackIntegration struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	TenantID            uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	WorkspaceName       string    `gorm:"size:255;not null" json:"workspace_name"`
	WebhookURL          string    `gorm:"type:text;not null" json:"webhook_url"` // Incoming webhook URL
	DefaultChannel      string    `gorm:"size:255" json:"default_channel"`
	IsActive            bool      `gorm:"default:true" json:"is_active"`
	NotifyOnDown        bool      `gorm:"default:true" json:"notify_on_down"`
	NotifyOnUp          bool      `gorm:"default:true" json:"notify_on_up"`
	NotifyOnDegraded    bool      `gorm:"default:true" json:"notify_on_degraded"`
	NotifyOnMaintenance bool      `gorm:"default:false" json:"notify_on_maintenance"`
	MentionUsers        string    `gorm:"type:text" json:"mention_users"` // Comma-separated user IDs
	MentionChannel      bool      `gorm:"default:false" json:"mention_channel"` // @channel mention
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// TableName specifies the table name for SlackIntegration
func (SlackIntegration) TableName() string {
	return "slack_integrations"
}

// SlackChannelSubscription represents a channel-specific subscription
type SlackChannelSubscription struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	IntegrationID  uint      `gorm:"not null;index" json:"integration_id"`
	MonitorID      uint      `gorm:"not null;index" json:"monitor_id"`
	ChannelName    string    `gorm:"size:255;not null" json:"channel_name"`
	ChannelID      string    `gorm:"size:255" json:"channel_id"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName specifies the table name for SlackChannelSubscription
func (SlackChannelSubscription) TableName() string {
	return "slack_channel_subscriptions"
}

// SlackNotification represents a sent Slack notification
type SlackNotification struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	IntegrationID  uint      `gorm:"not null;index" json:"integration_id"`
	MonitorID      uint      `gorm:"not null;index" json:"monitor_id"`
	ChannelName    string    `gorm:"size:255" json:"channel_name"`
	EventType      string    `gorm:"size:50;not null" json:"event_type"` // down, up, degraded, maintenance
	MessageText    string    `gorm:"type:text" json:"message_text"`
	Status         string    `gorm:"size:50" json:"status"` // sent, failed, pending
	ResponseCode   int       `json:"response_code"`
	ErrorMessage   string    `gorm:"type:text" json:"error_message"`
	SentAt         time.Time `json:"sent_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// TableName specifies the table name for SlackNotification
func (SlackNotification) TableName() string {
	return "slack_notifications"
}

// SlackMessage represents a Slack message payload
type SlackMessage struct {
	Text        string            `json:"text,omitempty"`
	Channel     string            `json:"channel,omitempty"`
	Username    string            `json:"username,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
	Blocks      []SlackBlock      `json:"blocks,omitempty"`
}

// SlackAttachment represents a Slack message attachment
type SlackAttachment struct {
	Color      string              `json:"color,omitempty"`
	Title      string              `json:"title,omitempty"`
	TitleLink  string              `json:"title_link,omitempty"`
	Text       string              `json:"text,omitempty"`
	Fields     []SlackAttachmentField `json:"fields,omitempty"`
	Footer     string              `json:"footer,omitempty"`
	FooterIcon string              `json:"footer_icon,omitempty"`
	Timestamp  int64               `json:"ts,omitempty"`
}

// SlackAttachmentField represents a field in a Slack attachment
type SlackAttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// SlackBlock represents a Slack block (for Block Kit)
type SlackBlock struct {
	Type string                 `json:"type"`
	Text map[string]interface{} `json:"text,omitempty"`
}

// SlackIntegrationService manages Slack integrations
type SlackIntegrationService struct {
	db            *gorm.DB
	logger        *zap.Logger
	serviceClient *resilience.ServiceClient
}

// NewSlackIntegrationService creates a new Slack integration service
func NewSlackIntegrationService(db *gorm.DB, serviceClient *resilience.ServiceClient, logger *zap.Logger) *SlackIntegrationService {
	return &SlackIntegrationService{
		db:            db,
		logger:        logger,
		serviceClient: serviceClient,
	}
}

// CreateIntegration creates a new Slack integration
func (s *SlackIntegrationService) CreateIntegration(integration *SlackIntegration) error {
	// Validate webhook URL by sending a test message
	err := s.TestWebhook(integration.WebhookURL)
	if err != nil {
		return fmt.Errorf("invalid webhook URL: %w", err)
	}

	err = s.db.Create(integration).Error
	if err != nil {
		s.logger.Error("Failed to create Slack integration",
			zap.Error(err),
			zap.String("workspace", integration.WorkspaceName))
		return fmt.Errorf("failed to create integration: %w", err)
	}

	s.logger.Info("Slack integration created",
		zap.Uint("id", integration.ID),
		zap.String("workspace", integration.WorkspaceName))

	return nil
}

// UpdateIntegration updates an existing Slack integration
func (s *SlackIntegrationService) UpdateIntegration(integration *SlackIntegration) error {
	err := s.db.Save(integration).Error
	if err != nil {
		s.logger.Error("Failed to update Slack integration", zap.Error(err))
		return fmt.Errorf("failed to update integration: %w", err)
	}

	s.logger.Info("Slack integration updated", zap.Uint("id", integration.ID))
	return nil
}

// GetIntegration retrieves a Slack integration by ID
func (s *SlackIntegrationService) GetIntegration(id uint, tenantID uuid.UUID) (*SlackIntegration, error) {
	var integration SlackIntegration
	err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&integration).Error
	if err != nil {
		return nil, fmt.Errorf("integration not found: %w", err)
	}
	return &integration, nil
}

// GetIntegrationsByTenant retrieves all integrations for a tenant
func (s *SlackIntegrationService) GetIntegrationsByTenant(tenantID uuid.UUID) ([]SlackIntegration, error) {
	var integrations []SlackIntegration
	err := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&integrations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}
	return integrations, nil
}

// DeleteIntegration deletes a Slack integration
func (s *SlackIntegrationService) DeleteIntegration(id uint, tenantID uuid.UUID) error {
	result := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&SlackIntegration{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete integration: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("integration not found")
	}

	s.logger.Info("Slack integration deleted", zap.Uint("id", id))
	return nil
}

// SubscribeChannel subscribes a channel to monitor notifications
func (s *SlackIntegrationService) SubscribeChannel(subscription *SlackChannelSubscription) error {
	err := s.db.Create(subscription).Error
	if err != nil {
		s.logger.Error("Failed to create channel subscription", zap.Error(err))
		return fmt.Errorf("failed to subscribe channel: %w", err)
	}

	s.logger.Info("Channel subscribed to monitor",
		zap.Uint("integration_id", subscription.IntegrationID),
		zap.Uint("monitor_id", subscription.MonitorID),
		zap.String("channel", subscription.ChannelName))

	return nil
}

// UnsubscribeChannel unsubscribes a channel from monitor notifications
func (s *SlackIntegrationService) UnsubscribeChannel(id uint) error {
	result := s.db.Delete(&SlackChannelSubscription{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to unsubscribe channel: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription not found")
	}

	s.logger.Info("Channel unsubscribed", zap.Uint("id", id))
	return nil
}

// GetChannelSubscriptions retrieves all subscriptions for a monitor
func (s *SlackIntegrationService) GetChannelSubscriptions(monitorID uint) ([]SlackChannelSubscription, error) {
	var subscriptions []SlackChannelSubscription
	err := s.db.Where("monitor_id = ? AND is_active = ?", monitorID, true).Find(&subscriptions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}
	return subscriptions, nil
}

// SendMonitorAlert sends a monitor alert to Slack
func (s *SlackIntegrationService) SendMonitorAlert(ctx context.Context, monitorID uint, tenantID uuid.UUID, eventType string, monitorName string, monitorURL string, details map[string]interface{}) error {
	// Get active integrations for the tenant
	integrations, err := s.GetIntegrationsByTenant(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get integrations: %w", err)
	}

	// Filter active integrations based on event type
	var activeIntegrations []SlackIntegration
	for _, integration := range integrations {
		if !integration.IsActive {
			continue
		}

		shouldNotify := false
		switch eventType {
		case "down":
			shouldNotify = integration.NotifyOnDown
		case "up":
			shouldNotify = integration.NotifyOnUp
		case "degraded":
			shouldNotify = integration.NotifyOnDegraded
		case "maintenance":
			shouldNotify = integration.NotifyOnMaintenance
		}

		if shouldNotify {
			activeIntegrations = append(activeIntegrations, integration)
		}
	}

	if len(activeIntegrations) == 0 {
		s.logger.Info("No active integrations for event type",
			zap.String("event_type", eventType))
		return nil
	}

	// Send notifications to each active integration
	for _, integration := range activeIntegrations {
		// Build message
		message := s.buildMonitorAlertMessage(eventType, monitorName, monitorURL, details, &integration)

		// Get channel subscriptions
		subscriptions, err := s.GetChannelSubscriptions(monitorID)
		if err != nil {
			s.logger.Error("Failed to get channel subscriptions", zap.Error(err))
			continue
		}

		// If no subscriptions, send to default channel
		if len(subscriptions) == 0 {
			if integration.DefaultChannel != "" {
				message.Channel = integration.DefaultChannel
			}
			s.sendSlackMessage(ctx, &integration, message, monitorID, eventType, "default")
		} else {
			// Send to each subscribed channel
			for _, subscription := range subscriptions {
				if !subscription.IsActive {
					continue
				}

				message.Channel = subscription.ChannelName
				s.sendSlackMessage(ctx, &integration, message, monitorID, eventType, subscription.ChannelName)
			}
		}
	}

	return nil
}

// buildMonitorAlertMessage builds a formatted Slack message for monitor alerts
func (s *SlackIntegrationService) buildMonitorAlertMessage(eventType string, monitorName string, monitorURL string, details map[string]interface{}, integration *SlackIntegration) *SlackMessage {
	// Determine color and emoji based on event type
	color := "#36a64f" // green
	emoji := ":white_check_mark:"
	statusText := "Operational"

	switch eventType {
	case "down":
		color = "#ff0000" // red
		emoji = ":x:"
		statusText = "Down"
	case "degraded":
		color = "#FFA500" // orange
		emoji = ":warning:"
		statusText = "Degraded"
	case "maintenance":
		color = "#808080" // gray
		emoji = ":construction:"
		statusText = "Maintenance"
	case "up":
		color = "#36a64f" // green
		emoji = ":white_check_mark:"
		statusText = "Recovered"
	}

	// Build message text
	messageText := fmt.Sprintf("%s *%s* is now *%s*", emoji, monitorName, statusText)

	// Add mentions if configured
	if integration.MentionChannel {
		messageText = "<!channel> " + messageText
	} else if integration.MentionUsers != "" {
		// Mention specific users
		messageText = fmt.Sprintf("<@%s> %s", integration.MentionUsers, messageText)
	}

	// Build attachment fields
	fields := []SlackAttachmentField{
		{
			Title: "Status",
			Value: statusText,
			Short: true,
		},
		{
			Title: "Monitor",
			Value: monitorName,
			Short: true,
		},
	}

	// Add additional details
	if responseTime, ok := details["response_time"]; ok {
		fields = append(fields, SlackAttachmentField{
			Title: "Response Time",
			Value: fmt.Sprintf("%v ms", responseTime),
			Short: true,
		})
	}

	if statusCode, ok := details["status_code"]; ok {
		fields = append(fields, SlackAttachmentField{
			Title: "Status Code",
			Value: fmt.Sprintf("%v", statusCode),
			Short: true,
		})
	}

	if errorMsg, ok := details["error"]; ok {
		fields = append(fields, SlackAttachmentField{
			Title: "Error",
			Value: fmt.Sprintf("%v", errorMsg),
			Short: false,
		})
	}

	if location, ok := details["location"]; ok {
		fields = append(fields, SlackAttachmentField{
			Title: "Location",
			Value: fmt.Sprintf("%v", location),
			Short: true,
		})
	}

	// Create attachment
	attachment := SlackAttachment{
		Color:     color,
		Title:     fmt.Sprintf("Monitor: %s", monitorName),
		TitleLink: monitorURL,
		Text:      fmt.Sprintf("Status changed to *%s*", statusText),
		Fields:    fields,
		Footer:    "Beakon Status Page",
		Timestamp: time.Now().Unix(),
	}

	return &SlackMessage{
		Text:        messageText,
		Username:    "Beakon Status Page",
		IconEmoji:   ":chart_with_upwards_trend:",
		Attachments: []SlackAttachment{attachment},
	}
}

// sendSlackMessage sends a message to Slack
func (s *SlackIntegrationService) sendSlackMessage(ctx context.Context, integration *SlackIntegration, message *SlackMessage, monitorID uint, eventType string, channelName string) {
	// Prepare notification record
	notification := SlackNotification{
		IntegrationID: integration.ID,
		MonitorID:     monitorID,
		ChannelName:   channelName,
		EventType:     eventType,
		MessageText:   message.Text,
		Status:        "pending",
	}

	// Marshal message to JSON
	payload, err := json.Marshal(message)
	if err != nil {
		s.logger.Error("Failed to marshal Slack message", zap.Error(err))
		notification.Status = "failed"
		notification.ErrorMessage = err.Error()
		s.db.Create(&notification)
		return
	}

	// Send via ServiceClient with circuit breaker
	resp, err := s.serviceClient.Call(ctx, resilience.ServiceRequest{
		ServiceName: "slack-api",
		Method:      "POST",
		URL:         integration.WebhookURL,
		Body:        payload,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	})

	if err != nil {
		s.logger.Error("Failed to send Slack message", zap.Error(err))
		notification.Status = "failed"
		notification.ErrorMessage = err.Error()
		s.db.Create(&notification)
		return
	}

	body := resp.Body
	notification.ResponseCode = resp.StatusCode
	notification.SentAt = time.Now()

	if resp.StatusCode == http.StatusOK {
		notification.Status = "sent"
		s.logger.Info("Slack message sent successfully",
			zap.Uint("integration_id", integration.ID),
			zap.Uint("monitor_id", monitorID),
			zap.String("channel", channelName),
			zap.String("event_type", eventType))
	} else {
		notification.Status = "failed"
		notification.ErrorMessage = string(body)
		s.logger.Error("Slack message failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)))
	}

	// Save notification record
	s.db.Create(&notification)
}

// TestWebhook tests a Slack webhook URL
func (s *SlackIntegrationService) TestWebhook(webhookURL string) error {
	message := SlackMessage{
		Text:      "✅ Beakon Status Page integration test successful!",
		Username:  "Beakon Status Page",
		IconEmoji: ":chart_with_upwards_trend:",
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal test message: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Send via ServiceClient with circuit breaker
	resp, err := s.serviceClient.Call(ctx, resilience.ServiceRequest{
		ServiceName: "slack-api",
		Method:      "POST",
		URL:         webhookURL,
		Body:        payload,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	})

	if err != nil {
		return fmt.Errorf("failed to send test message: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("test message failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	return nil
}

// GetNotificationHistory retrieves notification history for a monitor
func (s *SlackIntegrationService) GetNotificationHistory(monitorID uint, limit int) ([]SlackNotification, error) {
	var notifications []SlackNotification
	err := s.db.Where("monitor_id = ?", monitorID).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get notification history: %w", err)
	}

	return notifications, nil
}

// GetNotificationStats retrieves notification statistics
func (s *SlackIntegrationService) GetNotificationStats(tenantID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	var integrations []SlackIntegration
	err := s.db.Where("tenant_id = ?", tenantID).Find(&integrations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}

	if len(integrations) == 0 {
		return map[string]interface{}{
			"total_sent":     0,
			"total_failed":   0,
			"success_rate":   0.0,
			"by_event_type":  map[string]int{},
		}, nil
	}

	integrationIDs := make([]uint, len(integrations))
	for i, integration := range integrations {
		integrationIDs[i] = integration.ID
	}

	// Count total sent
	var totalSent int64
	s.db.Model(&SlackNotification{}).
		Where("integration_id IN ? AND status = ? AND sent_at BETWEEN ? AND ?", integrationIDs, "sent", startDate, endDate).
		Count(&totalSent)

	// Count total failed
	var totalFailed int64
	s.db.Model(&SlackNotification{}).
		Where("integration_id IN ? AND status = ? AND created_at BETWEEN ? AND ?", integrationIDs, "failed", startDate, endDate).
		Count(&totalFailed)

	// Count by event type
	type EventTypeCount struct {
		EventType string
		Count     int64
	}
	var eventTypeCounts []EventTypeCount
	s.db.Model(&SlackNotification{}).
		Select("event_type, COUNT(*) as count").
		Where("integration_id IN ? AND sent_at BETWEEN ? AND ?", integrationIDs, startDate, endDate).
		Group("event_type").
		Scan(&eventTypeCounts)

	byEventType := make(map[string]int64)
	for _, etc := range eventTypeCounts {
		byEventType[etc.EventType] = etc.Count
	}

	successRate := 0.0
	totalNotifications := totalSent + totalFailed
	if totalNotifications > 0 {
		successRate = float64(totalSent) / float64(totalNotifications) * 100
	}

	return map[string]interface{}{
		"total_sent":     totalSent,
		"total_failed":   totalFailed,
		"success_rate":   successRate,
		"by_event_type":  byEventType,
	}, nil
}
