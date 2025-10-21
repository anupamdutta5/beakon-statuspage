// Package services provides enhanced notification service with provider support.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/anupamdutta5/notification-service/internal/models"
	"github.com/anupamdutta5/notification-service/internal/providers"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// EnhancedNotificationService provides notification functionality with multiple providers.
type EnhancedNotificationService struct {
	db              *gorm.DB
	logger          *zap.Logger
	providerManager *providers.Manager
}

// NewEnhancedNotificationService creates a new enhanced notification service.
func NewEnhancedNotificationService(db *gorm.DB, logger *zap.Logger) *EnhancedNotificationService {
	service := &EnhancedNotificationService{
		db:              db,
		logger:          logger,
		providerManager: providers.NewManager(logger),
	}

	// Initialize default providers (this would typically come from configuration)
	service.initializeDefaultProviders()

	return service
}

// initializeDefaultProviders initializes default notification providers.
func (s *EnhancedNotificationService) initializeDefaultProviders() {
	// This would typically load from configuration or database
	// For now, we'll set up example configurations

	defaultConfigs := []providers.ProviderConfig{
		{
			Type:    "sms",
			Name:    "twilio",
			Enabled: false, // Disabled by default, requires configuration
			Settings: map[string]interface{}{
				"account_sid": "",
				"auth_token":  "",
				"from_number": "",
				"enabled":     false,
			},
		},
		{
			Type:    "slack",
			Name:    "slack",
			Enabled: false, // Disabled by default, requires configuration
			Settings: map[string]interface{}{
				"webhook_url": "",
				"channel":     "#alerts",
				"username":    "StatusPage",
				"icon_emoji":  ":warning:",
				"enabled":     false,
			},
		},
		{
			Type:    "teams",
			Name:    "teams",
			Enabled: false, // Disabled by default, requires configuration
			Settings: map[string]interface{}{
				"webhook_url": "",
				"enabled":     false,
			},
		},
		{
			Type:    "webhook",
			Name:    "webhook",
			Enabled: false, // Disabled by default, requires configuration
			Settings: map[string]interface{}{
				"url":     "",
				"method":  "POST",
				"headers": map[string]string{},
				"secret":  "",
				"timeout": 30,
				"enabled": false,
			},
		},
	}

	if err := s.providerManager.LoadProvidersFromConfig(defaultConfigs); err != nil {
		s.logger.Error("Failed to load default provider configurations", zap.Error(err))
	}
}

// SendNotification sends a notification using the appropriate provider.
func (s *EnhancedNotificationService) SendNotification(ctx context.Context, notification *models.Notification) error {
	// Create notification record first
	if err := s.db.Create(notification).Error; err != nil {
		s.logger.Error("Failed to create notification record", zap.Error(err))
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Parse recipients
	var recipients []string
	if err := json.Unmarshal([]byte(notification.Recipients), &recipients); err != nil {
		s.logger.Error("Failed to parse recipients", zap.Error(err))
		return fmt.Errorf("failed to parse recipients: %w", err)
	}

	// Parse metadata
	var metadata map[string]interface{}
	if notification.Metadata != "" {
		if err := json.Unmarshal([]byte(notification.Metadata), &metadata); err != nil {
			s.logger.Warn("Failed to parse notification metadata", zap.Error(err))
			metadata = make(map[string]interface{})
		}
	} else {
		metadata = make(map[string]interface{})
	}

	// Send to each recipient
	allSuccessful := true
	for _, recipient := range recipients {
		if err := s.sendToRecipient(ctx, notification, recipient, metadata); err != nil {
			s.logger.Error("Failed to send notification to recipient",
				zap.Uint("notification_id", notification.ID),
				zap.String("recipient", recipient),
				zap.Error(err))
			allSuccessful = false
		}
	}

	// Update notification status
	if allSuccessful {
		notification.Status = "sent"
		notification.SentAt = &[]time.Time{time.Now()}[0]
	} else {
		notification.Status = "failed"
		notification.FailedAt = &[]time.Time{time.Now()}[0]
	}

	if err := s.db.Save(notification).Error; err != nil {
		s.logger.Error("Failed to update notification status", zap.Error(err))
	}

	return nil
}

// sendToRecipient sends a notification to a specific recipient.
func (s *EnhancedNotificationService) sendToRecipient(ctx context.Context, notification *models.Notification, recipient string, metadata map[string]interface{}) error {
	// Get appropriate channel for the notification type
	channel, err := s.getChannelForType(notification.TenantID, notification.Type)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	if channel == nil {
		return fmt.Errorf("no channel configured for type: %s", notification.Type)
	}

	// Create delivery record
	delivery := &models.Delivery{
		NotificationID:  notification.ID,
		ChannelID:       channel.ID,
		Recipient:       recipient,
		Status:          "pending",
		Attempt:         1,
		MaxAttempts:     3,
	}

	if err := s.db.Create(delivery).Error; err != nil {
		return fmt.Errorf("failed to create delivery record: %w", err)
	}

	// Create provider request
	request := &providers.NotificationRequest{
		Recipient: recipient,
		Subject:   notification.Subject,
		Content:   notification.Content,
		Type:      notification.Type,
		Priority:  notification.Priority,
		Metadata:  metadata,
	}

	// Send via provider
	response, err := s.providerManager.Send(ctx, channel.Provider, request)
	if err != nil {
		delivery.Status = "failed"
		delivery.FailedAt = &[]time.Time{time.Now()}[0]
		delivery.Error = err.Error()
	} else {
		if response.Success {
			delivery.Status = "sent"
			delivery.SentAt = &[]time.Time{time.Now()}[0]
		} else {
			delivery.Status = "failed"
			delivery.FailedAt = &[]time.Time{time.Now()}[0]
			delivery.Error = response.Error
		}
		delivery.ResponseCode = response.ResponseCode
		delivery.ResponseMessage = response.ResponseBody

		// Store response metadata
		if len(response.Metadata) > 0 {
			metadataBytes, _ := json.Marshal(response.Metadata)
			delivery.Metadata = string(metadataBytes)
		}
	}

	// Update delivery record
	if err := s.db.Save(delivery).Error; err != nil {
		s.logger.Error("Failed to update delivery record", zap.Error(err))
	}

	return err
}

// getChannelForType retrieves the appropriate channel for a notification type.
func (s *EnhancedNotificationService) getChannelForType(tenantID uint, notificationType string) (*models.Channel, error) {
	var channel models.Channel
	err := s.db.Where("tenant_id = ? AND type = ? AND is_active = ?", tenantID, notificationType, true).
		Order("priority DESC, id ASC").
		First(&channel).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No channel found, not an error
		}
		return nil, err
	}

	return &channel, nil
}

// SendMaintenanceNotification sends a maintenance-related notification.
func (s *EnhancedNotificationService) SendMaintenanceNotification(ctx context.Context, tenantID uint, maintenanceData map[string]interface{}, notificationType string) error {
	// Get subscribers for maintenance notifications
	subscribers, err := s.getMaintenanceSubscribers(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get maintenance subscribers: %w", err)
	}

	if len(subscribers) == 0 {
		s.logger.Info("No subscribers found for maintenance notifications", zap.Uint("tenant_id", tenantID))
		return nil
	}

	// Create notification content
	subject, content := s.createMaintenanceContent(maintenanceData, notificationType)

	// Prepare recipients
	recipients := make([]string, len(subscribers))
	for i, sub := range subscribers {
		recipients[i] = sub.Recipient
	}

	recipientsJSON, _ := json.Marshal(recipients)
	metadataJSON, _ := json.Marshal(maintenanceData)

	// Create notification
	notification := &models.Notification{
		TenantID:   tenantID,
		Type:       "maintenance",
		Status:     "pending",
		Priority:   "normal",
		Subject:    subject,
		Content:    content,
		Recipients: string(recipientsJSON),
		Metadata:   string(metadataJSON),
	}

	return s.SendNotification(ctx, notification)
}

// SendIncidentNotification sends an incident-related notification.
func (s *EnhancedNotificationService) SendIncidentNotification(ctx context.Context, tenantID uint, incidentData map[string]interface{}, notificationType string) error {
	// Get subscribers for incident notifications
	subscribers, err := s.getIncidentSubscribers(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get incident subscribers: %w", err)
	}

	if len(subscribers) == 0 {
		s.logger.Info("No subscribers found for incident notifications", zap.Uint("tenant_id", tenantID))
		return nil
	}

	// Create notification content
	subject, content := s.createIncidentContent(incidentData, notificationType)

	// Determine priority based on incident severity
	priority := "normal"
	if severity, ok := incidentData["severity"].(string); ok {
		switch strings.ToLower(severity) {
		case "critical":
			priority = "urgent"
		case "major":
			priority = "high"
		case "minor":
			priority = "normal"
		default:
			priority = "low"
		}
	}

	// Prepare recipients
	recipients := make([]string, len(subscribers))
	for i, sub := range subscribers {
		recipients[i] = sub.Recipient
	}

	recipientsJSON, _ := json.Marshal(recipients)
	metadataJSON, _ := json.Marshal(incidentData)

	// Create notification
	notification := &models.Notification{
		TenantID:   tenantID,
		Type:       "incident",
		Status:     "pending",
		Priority:   priority,
		Subject:    subject,
		Content:    content,
		Recipients: string(recipientsJSON),
		Metadata:   string(metadataJSON),
	}

	return s.SendNotification(ctx, notification)
}

// Helper functions for content creation
func (s *EnhancedNotificationService) createMaintenanceContent(data map[string]interface{}, notificationType string) (string, string) {
	title, _ := data["title"].(string)
	description, _ := data["description"].(string)
	startTime, _ := data["start_time"].(string)
	endTime, _ := data["end_time"].(string)

	var subject, content string

	switch notificationType {
	case "maintenance.scheduled":
		subject = fmt.Sprintf("Scheduled Maintenance: %s", title)
		content = fmt.Sprintf("A maintenance window has been scheduled.\n\nTitle: %s\nDescription: %s\nStart Time: %s\nEnd Time: %s",
			title, description, startTime, endTime)
	case "maintenance.started":
		subject = fmt.Sprintf("Maintenance Started: %s", title)
		content = fmt.Sprintf("Scheduled maintenance has started.\n\nTitle: %s\nDescription: %s\nEnd Time: %s",
			title, description, endTime)
	case "maintenance.completed":
		subject = fmt.Sprintf("Maintenance Completed: %s", title)
		content = fmt.Sprintf("Scheduled maintenance has been completed.\n\nTitle: %s\nDescription: %s",
			title, description)
	default:
		subject = fmt.Sprintf("Maintenance Update: %s", title)
		content = fmt.Sprintf("Maintenance update.\n\nTitle: %s\nDescription: %s",
			title, description)
	}

	return subject, content
}

func (s *EnhancedNotificationService) createIncidentContent(data map[string]interface{}, notificationType string) (string, string) {
	title, _ := data["title"].(string)
	description, _ := data["description"].(string)
	status, _ := data["status"].(string)
	severity, _ := data["severity"].(string)

	var subject, content string

	switch notificationType {
	case "incident.created":
		subject = fmt.Sprintf("Incident Reported: %s", title)
		content = fmt.Sprintf("A new incident has been reported.\n\nTitle: %s\nDescription: %s\nSeverity: %s\nStatus: %s",
			title, description, severity, status)
	case "incident.updated":
		subject = fmt.Sprintf("Incident Update: %s", title)
		content = fmt.Sprintf("An incident has been updated.\n\nTitle: %s\nDescription: %s\nSeverity: %s\nStatus: %s",
			title, description, severity, status)
	case "incident.resolved":
		subject = fmt.Sprintf("Incident Resolved: %s", title)
		content = fmt.Sprintf("An incident has been resolved.\n\nTitle: %s\nDescription: %s",
			title, description)
	default:
		subject = fmt.Sprintf("Incident: %s", title)
		content = fmt.Sprintf("Incident update.\n\nTitle: %s\nDescription: %s\nStatus: %s",
			title, description, status)
	}

	return subject, content
}

// Subscriber helper functions
func (s *EnhancedNotificationService) getMaintenanceSubscribers(tenantID uint) ([]SubscriberInfo, error) {
	return s.getSubscribersForEventType(tenantID, "maintenance")
}

func (s *EnhancedNotificationService) getIncidentSubscribers(tenantID uint) ([]SubscriberInfo, error) {
	return s.getSubscribersForEventType(tenantID, "incident")
}

type SubscriberInfo struct {
	Recipient string
	Channel   string
}

func (s *EnhancedNotificationService) getSubscribersForEventType(tenantID uint, eventType string) ([]SubscriberInfo, error) {
	var subscriptions []models.Subscription
	err := s.db.Preload("Channel").
		Where("tenant_id = ? AND is_active = ? AND event_types LIKE ?", tenantID, true, "%"+eventType+"%").
		Find(&subscriptions).Error

	if err != nil {
		return nil, err
	}

	var subscribers []SubscriberInfo
	for _, sub := range subscriptions {
		// For now, we'll use a placeholder recipient format
		// In a real implementation, this would get the actual user email/phone/etc.
		recipient := fmt.Sprintf("user_%d", sub.UserID)
		subscribers = append(subscribers, SubscriberInfo{
			Recipient: recipient,
			Channel:   sub.Channel.Type,
		})
	}

	return subscribers, nil
}

// Provider management methods
func (s *EnhancedNotificationService) GetProviderManager() *providers.Manager {
	return s.providerManager
}

func (s *EnhancedNotificationService) ConfigureProvider(config *providers.ProviderConfig) error {
	return s.providerManager.UpdateProviderConfig(config.Name, config)
}

func (s *EnhancedNotificationService) TestProvider(ctx context.Context, config *providers.ProviderConfig, testRecipient string) (*providers.NotificationResponse, error) {
	return s.providerManager.TestProvider(ctx, config, testRecipient)
}