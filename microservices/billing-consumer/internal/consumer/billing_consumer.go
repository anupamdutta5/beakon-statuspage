// Package consumer provides billing consumer implementation.
package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anupamdutta5/billing-consumer/internal/config"
	"github.com/anupamdutta5/billing-consumer/internal/models"
	"github.com/anupamdutta5/billing-consumer/internal/services"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// BillingConsumer handles consuming billing events.
type BillingConsumer struct {
	config         *config.Config
	logger         *zap.Logger
	db             *gorm.DB
	billingService *services.BillingService
	invoiceService *services.InvoiceService
	paymentService *services.PaymentService
	queueService   *services.QueueService
}

// NewBillingConsumer creates a new billing consumer.
func NewBillingConsumer(cfg *config.Config, logger *zap.Logger) (*BillingConsumer, error) {
	// Initialize database
	db, err := initDatabase(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize services
	billingService := services.NewBillingService(db, logger)
	invoiceService := services.NewInvoiceService(cfg, logger)
	paymentService := services.NewPaymentService(cfg, logger)
	queueService := services.NewQueueService(cfg, logger)

	return &BillingConsumer{
		config:         cfg,
		logger:         logger,
		db:             db,
		billingService: billingService,
		invoiceService: invoiceService,
		paymentService: paymentService,
		queueService:   queueService,
	}, nil
}

// Start starts the billing consumer.
func (c *BillingConsumer) Start(ctx context.Context) error {
	c.logger.Info("Starting billing consumer",
		zap.String("queue", c.config.Queue.QueueName),
		zap.String("provider", c.config.Queue.Provider))

	// Start consuming messages
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Billing consumer context cancelled")
			return ctx.Err()
		default:
			if err := c.consumeMessages(ctx); err != nil {
				c.logger.Error("Failed to consume messages", zap.Error(err))
				time.Sleep(time.Duration(c.config.Billing.RetryBackoff) * time.Second)
			}
		}
	}
}

// consumeMessages consumes messages from the queue.
func (c *BillingConsumer) consumeMessages(ctx context.Context) error {
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

	c.logger.Info("Processing billing messages",
		zap.Int("count", len(messages)))

	// Process messages concurrently
	semaphore := make(chan struct{}, c.config.Billing.MaxConcurrency)
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
	for i := 0; i < c.config.Billing.MaxConcurrency; i++ {
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

// processMessage processes a single billing message.
func (c *BillingConsumer) processMessage(ctx context.Context, message *models.QueueMessage) error {
	c.logger.Info("Processing billing message",
		zap.String("message_id", message.ID),
		zap.String("event_type", message.EventType))

	// Parse billing event
	var billingEvent models.BillingEvent
	if err := json.Unmarshal([]byte(message.Data), &billingEvent); err != nil {
		return fmt.Errorf("failed to unmarshal billing event: %w", err)
	}

	// Process based on event type
	switch billingEvent.Type {
	case "subscription":
		return c.processSubscriptionEvent(ctx, &billingEvent)
	case "usage":
		return c.processUsageEvent(ctx, &billingEvent)
	case "payment":
		return c.processPaymentEvent(ctx, &billingEvent)
	case "invoice":
		return c.processInvoiceEvent(ctx, &billingEvent)
	default:
		return fmt.Errorf("unsupported billing event type: %s", billingEvent.Type)
	}
}

// processSubscriptionEvent processes a subscription event.
func (c *BillingConsumer) processSubscriptionEvent(ctx context.Context, event *models.BillingEvent) error {
	if !c.config.Billing.ProcessingEnabled {
		c.logger.Warn("Billing processing disabled, skipping",
			zap.String("event_id", event.ID))
		return nil
	}

	c.logger.Info("Processing subscription event",
		zap.String("event_id", event.ID),
		zap.String("subscription_id", event.SubscriptionID))

	// Process subscription
	if err := c.billingService.ProcessSubscription(ctx, event); err != nil {
		return fmt.Errorf("failed to process subscription: %w", err)
	}

	c.logger.Info("Subscription event processed successfully",
		zap.String("event_id", event.ID))

	return nil
}

// processUsageEvent processes a usage event.
func (c *BillingConsumer) processUsageEvent(ctx context.Context, event *models.BillingEvent) error {
	if !c.config.Billing.ProcessingEnabled {
		c.logger.Warn("Billing processing disabled, skipping",
			zap.String("event_id", event.ID))
		return nil
	}

	c.logger.Info("Processing usage event",
		zap.String("event_id", event.ID),
		zap.String("subscription_id", event.SubscriptionID))

	// Process usage
	if err := c.billingService.ProcessUsage(ctx, event); err != nil {
		return fmt.Errorf("failed to process usage: %w", err)
	}

	c.logger.Info("Usage event processed successfully",
		zap.String("event_id", event.ID))

	return nil
}

// processPaymentEvent processes a payment event.
func (c *BillingConsumer) processPaymentEvent(ctx context.Context, event *models.BillingEvent) error {
	if !c.config.Billing.PaymentEnabled {
		c.logger.Warn("Payment processing disabled, skipping",
			zap.String("event_id", event.ID))
		return nil
	}

	c.logger.Info("Processing payment event",
		zap.String("event_id", event.ID),
		zap.String("payment_id", event.PaymentID))

	// Process payment
	if err := c.paymentService.ProcessPayment(ctx, event); err != nil {
		return fmt.Errorf("failed to process payment: %w", err)
	}

	c.logger.Info("Payment event processed successfully",
		zap.String("event_id", event.ID))

	return nil
}

// processInvoiceEvent processes an invoice event.
func (c *BillingConsumer) processInvoiceEvent(ctx context.Context, event *models.BillingEvent) error {
	if !c.config.Billing.InvoiceEnabled {
		c.logger.Warn("Invoice processing disabled, skipping",
			zap.String("event_id", event.ID))
		return nil
	}

	c.logger.Info("Processing invoice event",
		zap.String("event_id", event.ID),
		zap.String("invoice_id", event.InvoiceID))

	// Process invoice
	if err := c.invoiceService.ProcessInvoice(ctx, event); err != nil {
		return fmt.Errorf("failed to process invoice: %w", err)
	}

	c.logger.Info("Invoice event processed successfully",
		zap.String("event_id", event.ID))

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

	// NOTE: Database migrations are managed by Atlas (see migrations/ directory and atlas.hcl)
	// Run migrations before starting the service:
	//   cd microservices/billing-consumer
	//   atlas migrate apply --env dev
	//
	// AutoMigrate is NOT used in this project as per best practices documented in CLAUDE.md
	// All schema changes must be tracked in version-controlled migration files

	return db, nil
}

