// Package consumer provides notification consumer implementation.
package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anupamdutta5/statuspage-notification-consumer/internal/config"
	"github.com/anupamdutta5/statuspage-notification-consumer/internal/models"
	"github.com/anupamdutta5/statuspage-notification-consumer/internal/services"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NotificationConsumer handles consuming notification events.
type NotificationConsumer struct {
	config              *config.Config
	logger              *zap.Logger
	db                  *gorm.DB
	notificationService *services.NotificationService
	emailService        *services.EmailService
	smsService          *services.SMSService
	webhookService      *services.WebhookService
	pushService         *services.PushService
	queueService        *services.QueueService
}

// NewNotificationConsumer creates a new notification consumer.
func NewNotificationConsumer(cfg *config.Config, logger *zap.Logger) (*NotificationConsumer, error) {
	// Initialize database
	db, err := initDatabase(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize services
	notificationService := services.NewNotificationService(db, logger)
	emailService := services.NewEmailService(cfg, logger)
	smsService := services.NewSMSService(cfg, logger)
	webhookService := services.NewWebhookService(cfg, logger)
	pushService := services.NewPushService(cfg, logger)
	queueService := services.NewQueueService(cfg, logger)

	return &NotificationConsumer{
		config:              cfg,
		logger:              logger,
		db:                  db,
		notificationService: notificationService,
		emailService:        emailService,
		smsService:          smsService,
		webhookService:      webhookService,
		pushService:         pushService,
		queueService:        queueService,
	}, nil
}

// Start starts the notification consumer.
func (c *NotificationConsumer) Start(ctx context.Context) error {
	c.logger.Info("Starting notification consumer",
		zap.String("queue", c.config.Queue.QueueName),
		zap.String("provider", c.config.Queue.Provider))

	// Start consuming messages
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Notification consumer context cancelled")
			return ctx.Err()
		default:
			if err := c.consumeMessages(ctx); err != nil {
				c.logger.Error("Failed to consume messages", zap.Error(err))
				time.Sleep(time.Duration(c.config.Notification.RetryBackoff) * time.Second)
			}
		}
	}
}

// consumeMessages consumes messages from the queue.
func (c *NotificationConsumer) consumeMessages(ctx context.Context) error {
	// Get messages from queue
	messages, err := c.queueService.GetMessages(ctx, c.config.Queue.QueueName, c.config.Queue.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to get messages: %w", err)
	}

	if len(messages) == 0 {
		// No messages, wait before polling again
		time.Sleep(time.Duration(c.config.Queue.PollTimeout) * time.Second)
		return nil
	}

	c.logger.Info("Processing notification messages",
		zap.Int("count", len(messages)))

	// Process messages concurrently
	semaphore := make(chan struct{}, c.config.Notification.MaxConcurrency)
	errors := make(chan error, len(messages))

	for _, message := range messages {
		semaphore <- struct{}{}
		go func(msg *models.QueueMessage) {
			defer func() { <-semaphore }()
			if err := c.processMessage(ctx, msg); err != nil {
				errors <- fmt.Errorf("failed to process message %s: %w", msg.ID, err)
			}
		}(message)
	}

	// Wait for all messages to be processed
	for i := 0; i < c.config.Notification.MaxConcurrency; i++ {
		semaphore <- struct{}{}
	}

	// Check for errors
	close(errors)
	var hasErrors bool
	for err := range errors {
		if err != nil {
			c.logger.Error("Message processing error", zap.Error(err))
			hasErrors = true
		}
	}

	if hasErrors {
		return fmt.Errorf("some messages failed to process")
	}

	return nil
}

// processMessage processes a single notification message.
func (c *NotificationConsumer) processMessage(ctx context.Context, message *models.QueueMessage) error {
	c.logger.Info("Processing notification message",
		zap.String("message_id", message.ID),
		zap.String("event_type", message.EventType))

	// Parse notification event
	var notificationEvent models.NotificationEvent
	if err := json.Unmarshal([]byte(message.Data), &notificationEvent); err != nil {
		return fmt.Errorf("failed to unmarshal notification event: %w", err)
	}

	// Process based on event type
	switch notificationEvent.Type {
	case "email":
		return c.processEmailNotification(ctx, &notificationEvent)
	case "sms":
		return c.processSMSNotification(ctx, &notificationEvent)
	case "webhook":
		return c.processWebhookNotification(ctx, &notificationEvent)
	case "push":
		return c.processPushNotification(ctx, &notificationEvent)
	default:
		return fmt.Errorf("unsupported notification type: %s", notificationEvent.Type)
	}
}

// processEmailNotification processes an email notification.
func (c *NotificationConsumer) processEmailNotification(ctx context.Context, event *models.NotificationEvent) error {
	if !c.config.Notification.EmailEnabled {
		c.logger.Warn("Email notifications disabled, skipping",
			zap.String("event_id", event.ID))
		return nil
	}

	c.logger.Info("Processing email notification",
		zap.String("event_id", event.ID),
		zap.String("recipient", event.Recipient))

	// Send email
	if err := c.emailService.SendEmail(ctx, event); err != nil {
		// Update notification status to failed
		if updateErr := c.updateNotificationStatus(event.NotificationID, "failed", err.Error()); updateErr != nil {
			c.logger.Error("Failed to update notification status", zap.Error(updateErr))
		}
		return fmt.Errorf("failed to send email: %w", err)
	}

	// Update notification status to sent
	if err := c.updateNotificationStatus(event.NotificationID, "sent", ""); err != nil {
		c.logger.Error("Failed to update notification status", zap.Error(err))
	}

	c.logger.Info("Email notification sent successfully",
		zap.String("event_id", event.ID))

	return nil
}

// processSMSNotification processes an SMS notification.
func (c *NotificationConsumer) processSMSNotification(ctx context.Context, event *models.NotificationEvent) error {
	if !c.config.Notification.SMSEnabled {
		c.logger.Warn("SMS notifications disabled, skipping",
			zap.String("event_id", event.ID))
		return nil
	}

	c.logger.Info("Processing SMS notification",
		zap.String("event_id", event.ID),
		zap.String("recipient", event.Recipient))

	// Send SMS
	if err := c.smsService.SendSMS(ctx, event); err != nil {
		// Update notification status to failed
		if updateErr := c.updateNotificationStatus(event.NotificationID, "failed", err.Error()); updateErr != nil {
			c.logger.Error("Failed to update notification status", zap.Error(updateErr))
		}
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	// Update notification status to sent
	if err := c.updateNotificationStatus(event.NotificationID, "sent", ""); err != nil {
		c.logger.Error("Failed to update notification status", zap.Error(err))
	}

	c.logger.Info("SMS notification sent successfully",
		zap.String("event_id", event.ID))

	return nil
}

// processWebhookNotification processes a webhook notification.
func (c *NotificationConsumer) processWebhookNotification(ctx context.Context, event *models.NotificationEvent) error {
	if !c.config.Notification.WebhookEnabled {
		c.logger.Warn("Webhook notifications disabled, skipping",
			zap.String("event_id", event.ID))
		return nil
	}

	c.logger.Info("Processing webhook notification",
		zap.String("event_id", event.ID),
		zap.String("url", event.WebhookURL))

	// Send webhook
	if err := c.webhookService.SendWebhook(ctx, event); err != nil {
		// Update notification status to failed
		if updateErr := c.updateNotificationStatus(event.NotificationID, "failed", err.Error()); updateErr != nil {
			c.logger.Error("Failed to update notification status", zap.Error(updateErr))
		}
		return fmt.Errorf("failed to send webhook: %w", err)
	}

	// Update notification status to sent
	if err := c.updateNotificationStatus(event.NotificationID, "sent", ""); err != nil {
		c.logger.Error("Failed to update notification status", zap.Error(err))
	}

	c.logger.Info("Webhook notification sent successfully",
		zap.String("event_id", event.ID))

	return nil
}

// processPushNotification processes a push notification.
func (c *NotificationConsumer) processPushNotification(ctx context.Context, event *models.NotificationEvent) error {
	if !c.config.Notification.PushEnabled {
		c.logger.Warn("Push notifications disabled, skipping",
			zap.String("event_id", event.ID))
		return nil
	}

	c.logger.Info("Processing push notification",
		zap.String("event_id", event.ID),
		zap.String("device_token", event.DeviceToken))

	// Send push notification
	if err := c.pushService.SendPush(ctx, event); err != nil {
		// Update notification status to failed
		if updateErr := c.updateNotificationStatus(event.NotificationID, "failed", err.Error()); updateErr != nil {
			c.logger.Error("Failed to update notification status", zap.Error(updateErr))
		}
		return fmt.Errorf("failed to send push notification: %w", err)
	}

	// Update notification status to sent
	if err := c.updateNotificationStatus(event.NotificationID, "sent", ""); err != nil {
		c.logger.Error("Failed to update notification status", zap.Error(err))
	}

	c.logger.Info("Push notification sent successfully",
		zap.String("event_id", event.ID))

	return nil
}

// updateNotificationStatus updates the status of a notification.
func (c *NotificationConsumer) updateNotificationStatus(notificationID string, status, error string) error {
	// This would typically update the notification in the database
	// For now, we'll just log the status update
	c.logger.Info("Updating notification status",
		zap.String("notification_id", notificationID),
		zap.String("status", status),
		zap.String("error", error))

	return nil
}

// initDatabase initializes the database connection.
func initDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.QueueMessage{},
		&models.NotificationEvent{},
		&models.ProcessingLog{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

