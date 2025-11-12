package teams

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

// TeamsIntegration represents a Microsoft Teams integration configuration
type TeamsIntegration struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	TenantID            uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name                string    `gorm:"size:255;not null" json:"name"`
	WebhookURL          string    `gorm:"type:text;not null" json:"webhook_url"` // Incoming webhook URL
	DefaultChannelName  string    `gorm:"size:255" json:"default_channel_name"`
	IsActive            bool      `gorm:"default:true" json:"is_active"`
	NotifyOnDown        bool      `gorm:"default:true" json:"notify_on_down"`
	NotifyOnUp          bool      `gorm:"default:true" json:"notify_on_up"`
	NotifyOnDegraded    bool      `gorm:"default:true" json:"notify_on_degraded"`
	NotifyOnMaintenance bool      `gorm:"default:false" json:"notify_on_maintenance"`
	MentionUsers        string    `gorm:"type:text" json:"mention_users"` // Comma-separated user principal names
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// TeamsChannelSubscription represents a channel-specific subscription
type TeamsChannelSubscription struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	IntegrationID  uint      `gorm:"not null;index" json:"integration_id"`
	MonitorID      uint      `gorm:"not null;index" json:"monitor_id"`
	WebhookURL     string    `gorm:"type:text;not null" json:"webhook_url"`
	ChannelName    string    `gorm:"size:255" json:"channel_name"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TeamsNotification represents a sent Teams notification
type TeamsNotification struct {
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

// TeamsAdaptiveCard represents a Microsoft Teams Adaptive Card
type TeamsAdaptiveCard struct {
	Type    string                   `json:"type"`
	Schema  string                   `json:"$schema,omitempty"`
	Version string                   `json:"version"`
	Body    []TeamsAdaptiveCardElement `json:"body"`
	Actions []TeamsAdaptiveCardAction  `json:"actions,omitempty"`
}

// TeamsAdaptiveCardElement represents an element in an Adaptive Card
type TeamsAdaptiveCardElement struct {
	Type   string      `json:"type"`
	Text   string      `json:"text,omitempty"`
	Weight string      `json:"weight,omitempty"`
	Size   string      `json:"size,omitempty"`
	Color  string      `json:"color,omitempty"`
	Wrap   bool        `json:"wrap,omitempty"`
	Spacing string     `json:"spacing,omitempty"`
	Items  []TeamsAdaptiveCardElement `json:"items,omitempty"`
	Columns []TeamsAdaptiveCardElement `json:"columns,omitempty"`
	Width  string      `json:"width,omitempty"`
	Facts  []TeamsFact `json:"facts,omitempty"`
}

// TeamsFact represents a fact in a FactSet
type TeamsFact struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

// TeamsAdaptiveCardAction represents an action in an Adaptive Card
type TeamsAdaptiveCardAction struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// TeamsMessage represents a Microsoft Teams message payload
type TeamsMessage struct {
	Type        string              `json:"@type,omitempty"`
	Context     string              `json:"@context,omitempty"`
	Summary     string              `json:"summary,omitempty"`
	ThemeColor  string              `json:"themeColor,omitempty"`
	Title       string              `json:"title,omitempty"`
	Text        string              `json:"text,omitempty"`
	Sections    []TeamsSection      `json:"sections,omitempty"`
	Attachments []TeamsAttachment   `json:"attachments,omitempty"`
}

// TeamsSection represents a section in a Teams message
type TeamsSection struct {
	ActivityTitle    string       `json:"activityTitle,omitempty"`
	ActivitySubtitle string       `json:"activitySubtitle,omitempty"`
	ActivityImage    string       `json:"activityImage,omitempty"`
	Facts            []TeamsFact  `json:"facts,omitempty"`
	Text             string       `json:"text,omitempty"`
}

// TeamsAttachment represents an attachment in a Teams message
type TeamsAttachment struct {
	ContentType string             `json:"contentType"`
	Content     TeamsAdaptiveCard  `json:"content"`
}

// TeamsIntegrationService manages Microsoft Teams integrations
type TeamsIntegrationService struct {
	db            *gorm.DB
	logger        *zap.Logger
	serviceClient *resilience.ServiceClient
}

// NewTeamsIntegrationService creates a new Teams integration service
func NewTeamsIntegrationService(db *gorm.DB, serviceClient *resilience.ServiceClient, logger *zap.Logger) *TeamsIntegrationService {
	return &TeamsIntegrationService{
		db:            db,
		logger:        logger,
		serviceClient: serviceClient,
	}
}

// CreateIntegration creates a new Teams integration
func (s *TeamsIntegrationService) CreateIntegration(integration *TeamsIntegration) error {
	// Validate webhook URL by sending a test message
	err := s.TestWebhook(integration.WebhookURL)
	if err != nil {
		return fmt.Errorf("invalid webhook URL: %w", err)
	}

	err = s.db.Create(integration).Error
	if err != nil {
		s.logger.Error("Failed to create Teams integration",
			zap.Error(err),
			zap.String("name", integration.Name))
		return fmt.Errorf("failed to create integration: %w", err)
	}

	s.logger.Info("Teams integration created",
		zap.Uint("id", integration.ID),
		zap.String("name", integration.Name))

	return nil
}

// UpdateIntegration updates an existing Teams integration
func (s *TeamsIntegrationService) UpdateIntegration(integration *TeamsIntegration) error {
	err := s.db.Save(integration).Error
	if err != nil {
		s.logger.Error("Failed to update Teams integration", zap.Error(err))
		return fmt.Errorf("failed to update integration: %w", err)
	}

	s.logger.Info("Teams integration updated", zap.Uint("id", integration.ID))
	return nil
}

// GetIntegration retrieves a Teams integration by ID
func (s *TeamsIntegrationService) GetIntegration(id uint, tenantID uuid.UUID) (*TeamsIntegration, error) {
	var integration TeamsIntegration
	err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&integration).Error
	if err != nil {
		return nil, fmt.Errorf("integration not found: %w", err)
	}
	return &integration, nil
}

// GetIntegrationsByTenant retrieves all integrations for a tenant
func (s *TeamsIntegrationService) GetIntegrationsByTenant(tenantID uuid.UUID) ([]TeamsIntegration, error) {
	var integrations []TeamsIntegration
	err := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&integrations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}
	return integrations, nil
}

// DeleteIntegration deletes a Teams integration
func (s *TeamsIntegrationService) DeleteIntegration(id uint, tenantID uuid.UUID) error {
	result := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&TeamsIntegration{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete integration: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("integration not found")
	}

	s.logger.Info("Teams integration deleted", zap.Uint("id", id))
	return nil
}

// SubscribeChannel subscribes a channel to monitor notifications
func (s *TeamsIntegrationService) SubscribeChannel(subscription *TeamsChannelSubscription) error {
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
func (s *TeamsIntegrationService) UnsubscribeChannel(id uint) error {
	result := s.db.Delete(&TeamsChannelSubscription{}, id)
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
func (s *TeamsIntegrationService) GetChannelSubscriptions(monitorID uint) ([]TeamsChannelSubscription, error) {
	var subscriptions []TeamsChannelSubscription
	err := s.db.Where("monitor_id = ? AND is_active = ?", monitorID, true).Find(&subscriptions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}
	return subscriptions, nil
}

// SendMonitorAlert sends a monitor alert to Teams
func (s *TeamsIntegrationService) SendMonitorAlert(ctx context.Context, monitorID uint, tenantID uuid.UUID, eventType string, monitorName string, monitorURL string, details map[string]interface{}) error {
	// Get active integrations for the tenant
	integrations, err := s.GetIntegrationsByTenant(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get integrations: %w", err)
	}

	// Filter active integrations based on event type
	var activeIntegrations []TeamsIntegration
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
		s.logger.Info("No active Teams integrations for event type",
			zap.String("event_type", eventType))
		return nil
	}

	// Send notifications to each active integration
	for _, integration := range activeIntegrations {
		// Build message
		message := s.buildAdaptiveCardMessage(eventType, monitorName, monitorURL, details, &integration)

		// Get channel subscriptions
		subscriptions, err := s.GetChannelSubscriptions(monitorID)
		if err != nil {
			s.logger.Error("Failed to get channel subscriptions", zap.Error(err))
			continue
		}

		// If no subscriptions, send to default webhook
		if len(subscriptions) == 0 {
			s.sendTeamsMessage(ctx, &integration, message, monitorID, eventType, integration.DefaultChannelName, integration.WebhookURL)
		} else {
			// Send to each subscribed channel
			for _, subscription := range subscriptions {
				if !subscription.IsActive {
					continue
				}

				s.sendTeamsMessage(ctx, &integration, message, monitorID, eventType, subscription.ChannelName, subscription.WebhookURL)
			}
		}
	}

	return nil
}

// buildAdaptiveCardMessage builds an Adaptive Card message for Teams
func (s *TeamsIntegrationService) buildAdaptiveCardMessage(eventType string, monitorName string, monitorURL string, details map[string]interface{}, integration *TeamsIntegration) *TeamsMessage {
	// Determine theme color based on event type
	themeColor := "00FF00" // green
	statusEmoji := "✅"
	statusText := "Operational"

	switch eventType {
	case "down":
		themeColor = "FF0000" // red
		statusEmoji = "🔴"
		statusText = "DOWN"
	case "degraded":
		themeColor = "FFA500" // orange
		statusEmoji = "⚠️"
		statusText = "Degraded"
	case "maintenance":
		themeColor = "808080" // gray
		statusEmoji = "🔧"
		statusText = "Maintenance"
	case "up":
		themeColor = "00FF00" // green
		statusEmoji = "✅"
		statusText = "Recovered"
	}

	// Build facts
	facts := []TeamsFact{
		{Title: "Status", Value: statusText},
		{Title: "Monitor", Value: monitorName},
		{Title: "Time", Value: time.Now().Format("January 2, 2006 at 3:04 PM MST")},
	}

	// Add additional facts from details
	if responseTime, ok := details["response_time"].(int); ok {
		facts = append(facts, TeamsFact{
			Title: "Response Time",
			Value: fmt.Sprintf("%d ms", responseTime),
		})
	}

	if statusCode, ok := details["status_code"].(int); ok {
		facts = append(facts, TeamsFact{
			Title: "Status Code",
			Value: fmt.Sprintf("%d", statusCode),
		})
	}

	if errorMsg, ok := details["error"].(string); ok && errorMsg != "" {
		facts = append(facts, TeamsFact{
			Title: "Error",
			Value: errorMsg,
		})
	}

	if location, ok := details["location"].(string); ok {
		facts = append(facts, TeamsFact{
			Title: "Location",
			Value: location,
		})
	}

	// Create Adaptive Card
	card := TeamsAdaptiveCard{
		Type:    "AdaptiveCard",
		Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
		Version: "1.2",
		Body: []TeamsAdaptiveCardElement{
			{
				Type:   "TextBlock",
				Text:   fmt.Sprintf("%s %s", statusEmoji, statusText),
				Weight: "Bolder",
				Size:   "Large",
				Color:  "Default",
			},
			{
				Type: "TextBlock",
				Text: fmt.Sprintf("**Monitor:** %s", monitorName),
				Wrap: true,
			},
			{
				Type:    "FactSet",
				Facts:   facts,
				Spacing: "Medium",
			},
		},
	}

	// Add action button if monitor URL is provided
	if monitorURL != "" {
		card.Actions = []TeamsAdaptiveCardAction{
			{
				Type:  "Action.OpenUrl",
				Title: "View Monitor Details",
				URL:   monitorURL,
			},
		}
	}

	// Build message with Adaptive Card
	message := &TeamsMessage{
		Type:       "MessageCard",
		Context:    "https://schema.org/extensions",
		Summary:    fmt.Sprintf("Monitor Alert: %s", monitorName),
		ThemeColor: themeColor,
		Attachments: []TeamsAttachment{
			{
				ContentType: "application/vnd.microsoft.card.adaptive",
				Content:     card,
			},
		},
	}

	// Add mentions if configured
	if integration.MentionUsers != "" {
		// Note: Mentions require specific user IDs and are complex in Adaptive Cards
		// For now, we'll add it to the summary
		message.Summary = fmt.Sprintf("@%s %s", integration.MentionUsers, message.Summary)
	}

	return message
}

// sendTeamsMessage sends a message to Microsoft Teams
func (s *TeamsIntegrationService) sendTeamsMessage(ctx context.Context, integration *TeamsIntegration, message *TeamsMessage, monitorID uint, eventType string, channelName string, webhookURL string) {
	// Prepare notification record
	notification := TeamsNotification{
		IntegrationID: integration.ID,
		MonitorID:     monitorID,
		ChannelName:   channelName,
		EventType:     eventType,
		MessageText:   message.Summary,
		Status:        "pending",
	}

	// Marshal message to JSON
	payload, err := json.Marshal(message)
	if err != nil {
		s.logger.Error("Failed to marshal Teams message", zap.Error(err))
		notification.Status = "failed"
		notification.ErrorMessage = err.Error()
		s.db.Create(&notification)
		return
	}

	// Send via ServiceClient with circuit breaker
	resp, err := s.serviceClient.Call(ctx, resilience.ServiceRequest{
		ServiceName: "teams-webhook",
		Method:      "POST",
		URL:         webhookURL,
		Body:        payload,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	})

	if err != nil {
		s.logger.Error("Failed to send Teams message", zap.Error(err))
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
		s.logger.Info("Teams message sent successfully",
			zap.Uint("integration_id", integration.ID),
			zap.Uint("monitor_id", monitorID),
			zap.String("channel", channelName),
			zap.String("event_type", eventType))
	} else {
		notification.Status = "failed"
		notification.ErrorMessage = string(body)
		s.logger.Error("Teams message failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)))
	}

	// Save notification record
	s.db.Create(&notification)
}

// TestWebhook tests a Teams webhook URL
func (s *TeamsIntegrationService) TestWebhook(webhookURL string) error {
	message := TeamsMessage{
		Type:       "MessageCard",
		Context:    "https://schema.org/extensions",
		Summary:    "Beakon Status Page integration test successful!",
		ThemeColor: "00FF00",
		Title:      "✅ Test Notification",
		Text:       "Your Microsoft Teams integration has been successfully configured.",
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal test message: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Send via ServiceClient with circuit breaker
	resp, err := s.serviceClient.Call(ctx, resilience.ServiceRequest{
		ServiceName: "teams-webhook",
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
func (s *TeamsIntegrationService) GetNotificationHistory(monitorID uint, limit int) ([]TeamsNotification, error) {
	var notifications []TeamsNotification
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
func (s *TeamsIntegrationService) GetNotificationStats(tenantID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	var integrations []TeamsIntegration
	err := s.db.Where("tenant_id = ?", tenantID).Find(&integrations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}

	if len(integrations) == 0 {
		return map[string]interface{}{
			"total_sent":     0,
			"total_failed":   0,
			"success_rate":   0.0,
			"by_event_type":  map[string]int64{},
		}, nil
	}

	integrationIDs := make([]uint, len(integrations))
	for i, integration := range integrations {
		integrationIDs[i] = integration.ID
	}

	// Count total sent
	var totalSent int64
	s.db.Model(&TeamsNotification{}).
		Where("integration_id IN ? AND status = ? AND sent_at BETWEEN ? AND ?", integrationIDs, "sent", startDate, endDate).
		Count(&totalSent)

	// Count total failed
	var totalFailed int64
	s.db.Model(&TeamsNotification{}).
		Where("integration_id IN ? AND status = ? AND created_at BETWEEN ? AND ?", integrationIDs, "failed", startDate, endDate).
		Count(&totalFailed)

	// Count by event type
	type EventTypeCount struct {
		EventType string
		Count     int64
	}
	var eventTypeCounts []EventTypeCount
	s.db.Model(&TeamsNotification{}).
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
