// Package consumer provides audit consumer implementation.
package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anupamdutta5/audit-consumer/internal/config"
	"github.com/anupamdutta5/audit-consumer/internal/models"
	"github.com/anupamdutta5/audit-consumer/internal/services"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// AuditConsumer handles consuming audit events.
type AuditConsumer struct {
	config            *config.Config
	logger            *zap.Logger
	db                *gorm.DB
	auditService      *services.AuditService
	complianceService *services.ComplianceService
	queueService      *services.QueueService
}

// NewAuditConsumer creates a new audit consumer.
func NewAuditConsumer(cfg *config.Config, logger *zap.Logger) (*AuditConsumer, error) {
	// Initialize database
	db, err := initDatabase(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize services
	auditService := services.NewAuditService(db, logger)
	complianceService := services.NewComplianceService(cfg, logger)
	queueService := services.NewQueueService(cfg, logger)

	return &AuditConsumer{
		config:            cfg,
		logger:            logger,
		db:                db,
		auditService:      auditService,
		complianceService: complianceService,
		queueService:      queueService,
	}, nil
}

// Start starts the audit consumer.
func (c *AuditConsumer) Start(ctx context.Context) error {
	c.logger.Info("Starting audit consumer",
		zap.String("queue", c.config.Queue.QueueName),
		zap.String("provider", c.config.Queue.Provider))

	// Start consuming messages
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Audit consumer context cancelled")
			return ctx.Err()
		default:
			if err := c.consumeMessages(ctx); err != nil {
				c.logger.Error("Failed to consume messages", zap.Error(err))
				time.Sleep(time.Duration(c.config.Audit.RetryBackoff) * time.Second)
			}
		}
	}
}

// consumeMessages consumes messages from the queue.
func (c *AuditConsumer) consumeMessages(ctx context.Context) error {
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

	c.logger.Info("Processing audit messages",
		zap.Int("count", len(messages)))

	// Process messages concurrently
	semaphore := make(chan struct{}, c.config.Audit.MaxConcurrency)
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
	for i := 0; i < c.config.Audit.MaxConcurrency; i++ {
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

// processMessage processes a single audit message.
func (c *AuditConsumer) processMessage(ctx context.Context, message *models.QueueMessage) error {
	c.logger.Info("Processing audit message",
		zap.String("message_id", message.ID),
		zap.String("event_type", message.EventType))

	// Parse audit event
	var auditEvent models.AuditEvent
	if err := json.Unmarshal([]byte(message.Data), &auditEvent); err != nil {
		return fmt.Errorf("failed to unmarshal audit event: %w", err)
	}

	// Process audit event
	if err := c.processAuditEvent(ctx, &auditEvent); err != nil {
		return fmt.Errorf("failed to process audit event: %w", err)
	}

	// Trigger compliance checks if enabled
	if c.config.Audit.ComplianceEnabled {
		if err := c.complianceService.CheckCompliance(ctx, &auditEvent); err != nil {
			c.logger.Error("Failed to check compliance", zap.Error(err))
		}
	}

	c.logger.Info("Audit event processed successfully",
		zap.String("event_id", auditEvent.ID))

	return nil
}

// processAuditEvent processes an audit event.
func (c *AuditConsumer) processAuditEvent(ctx context.Context, event *models.AuditEvent) error {
	if !c.config.Audit.ProcessingEnabled {
		c.logger.Warn("Audit processing disabled, skipping",
			zap.String("event_id", event.ID))
		return nil
	}

	c.logger.Info("Processing audit event",
		zap.String("event_id", event.ID),
		zap.String("action", event.Action),
		zap.String("resource", event.Resource))

	// Process audit event
	if err := c.auditService.ProcessAuditEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to process audit event: %w", err)
	}

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
	//   cd microservices/audit-consumer
	//   atlas migrate apply --env dev
	//
	// AutoMigrate is NOT used in this project as per best practices documented in CLAUDE.md
	// All schema changes must be tracked in version-controlled migration files

	return db, nil
}

