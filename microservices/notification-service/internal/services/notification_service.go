// Package services provides business logic for the Notification Service.
package services

import (
	"fmt"
	"time"

	"github.com/anupamdutta5/notification-service/internal/config"
	"github.com/anupamdutta5/notification-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NotificationService handles notification-related business logic.
type NotificationService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewNotificationService creates a new notification service.
func NewNotificationService(db *gorm.DB, logger *zap.Logger) *NotificationService {
	// NOTE: Database migrations are managed by Atlas (see migrations/ directory and atlas.hcl)
	// Run migrations before starting the service:
	//   cd microservices/notification-service
	//   atlas migrate apply --env dev
	//
	// AutoMigrate is NOT used in this project as per best practices documented in CLAUDE.md
	// All schema changes must be tracked in version-controlled migration files

	return &NotificationService{
		db:     db,
		logger: logger,
	}
}

// CreateNotification creates a new notification.
func (s *NotificationService) CreateNotification(notification *models.Notification) error {
	// Set default values
	if notification.Type == "" {
		notification.Type = "email"
	}
	if notification.Status == "" {
		notification.Status = "pending"
	}
	if notification.Priority == "" {
		notification.Priority = "normal"
	}
	if notification.MaxRetries == 0 {
		notification.MaxRetries = 3
	}
	if notification.RetryCount == 0 {
		notification.RetryCount = 0
	}

	// Create notification
	if err := s.db.Create(notification).Error; err != nil {
		s.logger.Error("Failed to create notification", zap.Error(err))
		return fmt.Errorf("failed to create notification: %w", err)
	}

	s.logger.Info("Notification created successfully", zap.Uint("notification_id", notification.ID))
	return nil
}

// GetNotification retrieves a notification by ID.
func (s *NotificationService) GetNotification(id uint) (*models.Notification, error) {
	var notification models.Notification
	if err := s.db.Preload("Template").Preload("Deliveries").First(&notification, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("notification not found")
		}
		s.logger.Error("Failed to get notification", zap.Error(err))
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	return &notification, nil
}

// GetNotifications retrieves a list of notifications with pagination.
func (s *NotificationService) GetNotifications(tenantID uint, limit, offset int) ([]*models.Notification, int64, error) {
	var notifications []*models.Notification
	var total int64

	// Get total count
	if err := s.db.Model(&models.Notification{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count notifications", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Get notifications with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).Preload("Template").Preload("Deliveries").Limit(limit).Offset(offset).Order("created_at DESC").Find(&notifications).Error; err != nil {
		s.logger.Error("Failed to get notifications", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get notifications: %w", err)
	}

	return notifications, total, nil
}

// UpdateNotification updates a notification.
func (s *NotificationService) UpdateNotification(notification *models.Notification) error {
	if err := s.db.Save(notification).Error; err != nil {
		s.logger.Error("Failed to update notification", zap.Error(err))
		return fmt.Errorf("failed to update notification: %w", err)
	}

	s.logger.Info("Notification updated successfully", zap.Uint("notification_id", notification.ID))
	return nil
}

// DeleteNotification soft deletes a notification.
func (s *NotificationService) DeleteNotification(id uint) error {
	if err := s.db.Delete(&models.Notification{}, id).Error; err != nil {
		s.logger.Error("Failed to delete notification", zap.Error(err))
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	s.logger.Info("Notification deleted successfully", zap.Uint("notification_id", id))
	return nil
}

// SendNotification sends a notification.
func (s *NotificationService) SendNotification(notificationID uint) error {
	// Get notification
	var notification models.Notification
	if err := s.db.First(&notification, notificationID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("notification not found")
		}
		s.logger.Error("Failed to get notification", zap.Error(err))
		return fmt.Errorf("failed to get notification: %w", err)
	}

	// Update status to sending
	notification.Status = "sending"
	if err := s.db.Save(&notification).Error; err != nil {
		s.logger.Error("Failed to update notification status", zap.Error(err))
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	// Here you would implement the actual sending logic
	// For now, we'll just simulate sending
	time.Sleep(100 * time.Millisecond)

	// Update status to sent
	notification.Status = "sent"
	now := time.Now()
	notification.SentAt = &now
	if err := s.db.Save(&notification).Error; err != nil {
		s.logger.Error("Failed to update notification status", zap.Error(err))
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	s.logger.Info("Notification sent successfully", zap.Uint("notification_id", notificationID))
	return nil
}

// GetNotificationStatus returns the status of a notification.
func (s *NotificationService) GetNotificationStatus(notificationID uint) (map[string]interface{}, error) {
	var notification models.Notification
	if err := s.db.Preload("Deliveries").First(&notification, notificationID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("notification not found")
		}
		s.logger.Error("Failed to get notification", zap.Error(err))
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	status := map[string]interface{}{
		"notification_id": notification.ID,
		"status":          notification.Status,
		"priority":        notification.Priority,
		"type":            notification.Type,
		"retry_count":     notification.RetryCount,
		"max_retries":     notification.MaxRetries,
		"created_at":      notification.CreatedAt,
		"scheduled_at":    notification.ScheduledAt,
		"sent_at":         notification.SentAt,
		"failed_at":       notification.FailedAt,
		"error":           notification.Error,
		"deliveries":      notification.Deliveries,
		"timestamp":       time.Now().UTC(),
	}

	return status, nil
}

// Template Management

// CreateTemplate creates a new notification template.
func (s *NotificationService) CreateTemplate(template *models.Template) error {
	// Set default values
	if template.Type == "" {
		template.Type = "email"
	}
	if template.Category == "" {
		template.Category = "general"
	}
	if template.IsActive == false && template.IsActive != true {
		template.IsActive = true
	}
	if template.IsDefault == false && template.IsDefault != true {
		template.IsDefault = false
	}

	// Create template
	if err := s.db.Create(template).Error; err != nil {
		s.logger.Error("Failed to create template", zap.Error(err))
		return fmt.Errorf("failed to create template: %w", err)
	}

	s.logger.Info("Template created successfully", zap.Uint("template_id", template.ID))
	return nil
}

// GetTemplate retrieves a template by ID.
func (s *NotificationService) GetTemplate(id uint) (*models.Template, error) {
	var template models.Template
	if err := s.db.First(&template, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("template not found")
		}
		s.logger.Error("Failed to get template", zap.Error(err))
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return &template, nil
}

// GetTemplates retrieves a list of templates with pagination.
func (s *NotificationService) GetTemplates(tenantID uint, limit, offset int) ([]*models.Template, int64, error) {
	var templates []*models.Template
	var total int64

	// Get total count
	if err := s.db.Model(&models.Template{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count templates", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count templates: %w", err)
	}

	// Get templates with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).Limit(limit).Offset(offset).Order("created_at DESC").Find(&templates).Error; err != nil {
		s.logger.Error("Failed to get templates", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get templates: %w", err)
	}

	return templates, total, nil
}

// UpdateTemplate updates a template.
func (s *NotificationService) UpdateTemplate(template *models.Template) error {
	if err := s.db.Save(template).Error; err != nil {
		s.logger.Error("Failed to update template", zap.Error(err))
		return fmt.Errorf("failed to update template: %w", err)
	}

	s.logger.Info("Template updated successfully", zap.Uint("template_id", template.ID))
	return nil
}

// DeleteTemplate soft deletes a template.
func (s *NotificationService) DeleteTemplate(id uint) error {
	if err := s.db.Delete(&models.Template{}, id).Error; err != nil {
		s.logger.Error("Failed to delete template", zap.Error(err))
		return fmt.Errorf("failed to delete template: %w", err)
	}

	s.logger.Info("Template deleted successfully", zap.Uint("template_id", id))
	return nil
}

// TestTemplate tests a template by sending a test notification.
func (s *NotificationService) TestTemplate(templateID uint, testRecipient string) error {
	// Get template
	var template models.Template
	if err := s.db.First(&template, templateID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("template not found")
		}
		s.logger.Error("Failed to get template", zap.Error(err))
		return fmt.Errorf("failed to get template: %w", err)
	}

	// Create test notification
	testNotification := &models.Notification{
		TenantID:   template.TenantID,
		TemplateID: &template.ID,
		Type:       template.Type,
		Status:     "pending",
		Priority:   "normal",
		Subject:    "[TEST] " + template.Subject,
		Content:    template.Content,
		Recipients: fmt.Sprintf(`["%s"]`, testRecipient),
		Metadata:   `{"test": true}`,
	}

	if err := s.CreateNotification(testNotification); err != nil {
		return fmt.Errorf("failed to create test notification: %w", err)
	}

	// Send test notification
	if err := s.SendNotification(testNotification.ID); err != nil {
		return fmt.Errorf("failed to send test notification: %w", err)
	}

	s.logger.Info("Template test completed successfully", zap.Uint("template_id", templateID))
	return nil
}

// Channel Management

// CreateChannel creates a new notification channel.
func (s *NotificationService) CreateChannel(channel *models.Channel) error {
	// Set default values
	if channel.Type == "" {
		channel.Type = "email"
	}
	if channel.Provider == "" {
		channel.Provider = "smtp"
	}
	if channel.IsActive == false && channel.IsActive != true {
		channel.IsActive = true
	}
	if channel.IsDefault == false && channel.IsDefault != true {
		channel.IsDefault = false
	}
	if channel.Priority == 0 {
		channel.Priority = 0
	}

	// Create channel
	if err := s.db.Create(channel).Error; err != nil {
		s.logger.Error("Failed to create channel", zap.Error(err))
		return fmt.Errorf("failed to create channel: %w", err)
	}

	s.logger.Info("Channel created successfully", zap.Uint("channel_id", channel.ID))
	return nil
}

// GetChannel retrieves a channel by ID.
func (s *NotificationService) GetChannel(id uint) (*models.Channel, error) {
	var channel models.Channel
	if err := s.db.First(&channel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("channel not found")
		}
		s.logger.Error("Failed to get channel", zap.Error(err))
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	return &channel, nil
}

// GetChannels retrieves a list of channels with pagination.
func (s *NotificationService) GetChannels(tenantID uint, limit, offset int) ([]*models.Channel, int64, error) {
	var channels []*models.Channel
	var total int64

	// Get total count
	if err := s.db.Model(&models.Channel{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count channels", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count channels: %w", err)
	}

	// Get channels with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).Limit(limit).Offset(offset).Order("priority DESC, created_at DESC").Find(&channels).Error; err != nil {
		s.logger.Error("Failed to get channels", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get channels: %w", err)
	}

	return channels, total, nil
}

// UpdateChannel updates a channel.
func (s *NotificationService) UpdateChannel(channel *models.Channel) error {
	if err := s.db.Save(channel).Error; err != nil {
		s.logger.Error("Failed to update channel", zap.Error(err))
		return fmt.Errorf("failed to update channel: %w", err)
	}

	s.logger.Info("Channel updated successfully", zap.Uint("channel_id", channel.ID))
	return nil
}

// DeleteChannel soft deletes a channel.
func (s *NotificationService) DeleteChannel(id uint) error {
	if err := s.db.Delete(&models.Channel{}, id).Error; err != nil {
		s.logger.Error("Failed to delete channel", zap.Error(err))
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	s.logger.Info("Channel deleted successfully", zap.Uint("channel_id", id))
	return nil
}

// TestChannel tests a channel by sending a test notification.
func (s *NotificationService) TestChannel(channelID uint, testRecipient string) error {
	// Get channel
	var channel models.Channel
	if err := s.db.First(&channel, channelID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("channel not found")
		}
		s.logger.Error("Failed to get channel", zap.Error(err))
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Create test notification
	testNotification := &models.Notification{
		TenantID:   channel.TenantID,
		Type:       channel.Type,
		Status:     "pending",
		Priority:   "normal",
		Subject:    "[TEST] Channel Test",
		Content:    "This is a test notification to verify channel configuration.",
		Recipients: fmt.Sprintf(`["%s"]`, testRecipient),
		Metadata:   `{"test": true, "channel_id": ` + fmt.Sprintf("%d", channelID) + `}`,
	}

	if err := s.CreateNotification(testNotification); err != nil {
		return fmt.Errorf("failed to create test notification: %w", err)
	}

	// Send test notification
	if err := s.SendNotification(testNotification.ID); err != nil {
		return fmt.Errorf("failed to send test notification: %w", err)
	}

	s.logger.Info("Channel test completed successfully", zap.Uint("channel_id", channelID))
	return nil
}

// Subscription Management

// CreateSubscription creates a new notification subscription.
func (s *NotificationService) CreateSubscription(subscription *models.Subscription) error {
	// Set default values
	if subscription.EventTypes == "" {
		subscription.EventTypes = `["all"]`
	}
	if subscription.IsActive == false && subscription.IsActive != true {
		subscription.IsActive = true
	}

	// Create subscription
	if err := s.db.Create(subscription).Error; err != nil {
		s.logger.Error("Failed to create subscription", zap.Error(err))
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	s.logger.Info("Subscription created successfully", zap.Uint("subscription_id", subscription.ID))
	return nil
}

// GetSubscription retrieves a subscription by ID.
func (s *NotificationService) GetSubscription(id uint) (*models.Subscription, error) {
	var subscription models.Subscription
	if err := s.db.Preload("Channel").First(&subscription, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription not found")
		}
		s.logger.Error("Failed to get subscription", zap.Error(err))
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return &subscription, nil
}

// GetSubscriptions retrieves a list of subscriptions with pagination.
func (s *NotificationService) GetSubscriptions(tenantID uint, limit, offset int) ([]*models.Subscription, int64, error) {
	var subscriptions []*models.Subscription
	var total int64

	// Get total count
	if err := s.db.Model(&models.Subscription{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count subscriptions", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	// Get subscriptions with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).Preload("Channel").Limit(limit).Offset(offset).Order("created_at DESC").Find(&subscriptions).Error; err != nil {
		s.logger.Error("Failed to get subscriptions", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	return subscriptions, total, nil
}

// UpdateSubscription updates a subscription.
func (s *NotificationService) UpdateSubscription(subscription *models.Subscription) error {
	if err := s.db.Save(subscription).Error; err != nil {
		s.logger.Error("Failed to update subscription", zap.Error(err))
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	s.logger.Info("Subscription updated successfully", zap.Uint("subscription_id", subscription.ID))
	return nil
}

// DeleteSubscription soft deletes a subscription.
func (s *NotificationService) DeleteSubscription(id uint) error {
	if err := s.db.Delete(&models.Subscription{}, id).Error; err != nil {
		s.logger.Error("Failed to delete subscription", zap.Error(err))
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	s.logger.Info("Subscription deleted successfully", zap.Uint("subscription_id", id))
	return nil
}

// Delivery Management

// GetDeliveries retrieves a list of deliveries with pagination.
func (s *NotificationService) GetDeliveries(tenantID uint, limit, offset int) ([]*models.Delivery, int64, error) {
	var deliveries []*models.Delivery
	var total int64

	// Get total count
	if err := s.db.Model(&models.Delivery{}).Joins("JOIN notifications ON deliveries.notification_id = notifications.id").Where("notifications.tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count deliveries", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count deliveries: %w", err)
	}

	// Get deliveries with pagination
	if err := s.db.Joins("JOIN notifications ON deliveries.notification_id = notifications.id").Where("notifications.tenant_id = ?", tenantID).Preload("Notification").Preload("Channel").Limit(limit).Offset(offset).Order("deliveries.created_at DESC").Find(&deliveries).Error; err != nil {
		s.logger.Error("Failed to get deliveries", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get deliveries: %w", err)
	}

	return deliveries, total, nil
}

// GetDelivery retrieves a delivery by ID.
func (s *NotificationService) GetDelivery(id uint) (*models.Delivery, error) {
	var delivery models.Delivery
	if err := s.db.Preload("Notification").Preload("Channel").First(&delivery, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("delivery not found")
		}
		s.logger.Error("Failed to get delivery", zap.Error(err))
		return nil, fmt.Errorf("failed to get delivery: %w", err)
	}

	return &delivery, nil
}

// RetryDelivery retries a failed delivery.
func (s *NotificationService) RetryDelivery(deliveryID uint) error {
	var delivery models.Delivery
	if err := s.db.First(&delivery, deliveryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("delivery not found")
		}
		s.logger.Error("Failed to get delivery", zap.Error(err))
		return fmt.Errorf("failed to get delivery: %w", err)
	}

	// Check if delivery can be retried
	if delivery.Attempt >= delivery.MaxAttempts {
		return fmt.Errorf("delivery has reached maximum retry attempts")
	}

	// Update delivery status
	delivery.Status = "pending"
	delivery.Attempt++
	delivery.Error = ""
	if err := s.db.Save(&delivery).Error; err != nil {
		s.logger.Error("Failed to update delivery", zap.Error(err))
		return fmt.Errorf("failed to update delivery: %w", err)
	}

	// Here you would implement the actual retry logic
	// For now, we'll just simulate retry
	time.Sleep(100 * time.Millisecond)

	// Update delivery status to sent
	delivery.Status = "sent"
	now := time.Now()
	delivery.SentAt = &now
	if err := s.db.Save(&delivery).Error; err != nil {
		s.logger.Error("Failed to update delivery status", zap.Error(err))
		return fmt.Errorf("failed to update delivery status: %w", err)
	}

	s.logger.Info("Delivery retry completed successfully", zap.Uint("delivery_id", deliveryID))
	return nil
}

// HandleWebhook handles incoming webhook events.
func (s *NotificationService) HandleWebhook(webhookEvent *models.WebhookEvent) error {
	// Set default values
	if webhookEvent.Status == "" {
		webhookEvent.Status = "pending"
	}
	if webhookEvent.MaxRetries == 0 {
		webhookEvent.MaxRetries = 3
	}
	if webhookEvent.RetryCount == 0 {
		webhookEvent.RetryCount = 0
	}

	// Create webhook event
	if err := s.db.Create(webhookEvent).Error; err != nil {
		s.logger.Error("Failed to create webhook event", zap.Error(err))
		return fmt.Errorf("failed to create webhook event: %w", err)
	}

	// Here you would implement the actual webhook processing logic
	// For now, we'll just mark it as processed
	webhookEvent.Status = "processed"
	now := time.Now()
	webhookEvent.ProcessedAt = &now
	if err := s.db.Save(webhookEvent).Error; err != nil {
		s.logger.Error("Failed to update webhook event", zap.Error(err))
		return fmt.Errorf("failed to update webhook event: %w", err)
	}

	s.logger.Info("Webhook event processed successfully", zap.Uint("webhook_event_id", webhookEvent.ID))
	return nil
}

// InitDatabase initializes the database connection and runs migrations.
func InitDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// NOTE: Database migrations are managed by Atlas (see migrations/ directory and atlas.hcl)
	// Run migrations before starting the service:
	//   cd microservices/notification-service
	//   atlas migrate apply --env dev
	//
	// AutoMigrate is NOT used in this project as per best practices documented in CLAUDE.md
	// All schema changes must be tracked in version-controlled migration files

	return db, nil
}
