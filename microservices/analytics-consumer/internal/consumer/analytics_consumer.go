// Package consumer provides analytics consumer implementation.
package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage-analytics-consumer/internal/config"
	"github.com/enterprise-status/statuspage-analytics-consumer/internal/models"
	"github.com/enterprise-status/statuspage-analytics-consumer/internal/services"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// AnalyticsConsumer handles consuming analytics events.
type AnalyticsConsumer struct {
	config             *config.Config
	logger             *zap.Logger
	db                 *gorm.DB
	analyticsService   *services.AnalyticsService
	aggregationService *services.AggregationService
	reportingService   *services.ReportingService
	queueService       *services.QueueService
}

// NewAnalyticsConsumer creates a new analytics consumer.
func NewAnalyticsConsumer(cfg *config.Config, logger *zap.Logger) (*AnalyticsConsumer, error) {
	// Initialize database
	db, err := initDatabase(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize services
	analyticsService := services.NewAnalyticsService(db, logger)
	aggregationService := services.NewAggregationService(cfg, logger)
	reportingService := services.NewReportingService(cfg, logger)
	queueService := services.NewQueueService(cfg, logger)

	return &AnalyticsConsumer{
		config:             cfg,
		logger:             logger,
		db:                 db,
		analyticsService:   analyticsService,
		aggregationService: aggregationService,
		reportingService:   reportingService,
		queueService:       queueService,
	}, nil
}

// Start starts the analytics consumer.
func (c *AnalyticsConsumer) Start(ctx context.Context) error {
	c.logger.Info("Starting analytics consumer",
		zap.String("queue", c.config.Queue.QueueName),
		zap.String("provider", c.config.Queue.Provider))

	// Start consuming messages
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Analytics consumer context cancelled")
			return ctx.Err()
		default:
			if err := c.consumeMessages(ctx); err != nil {
				c.logger.Error("Failed to consume messages", zap.Error(err))
				time.Sleep(time.Duration(c.config.Analytics.RetryBackoff) * time.Second)
			}
		}
	}
}

// consumeMessages consumes messages from the queue.
func (c *AnalyticsConsumer) consumeMessages(ctx context.Context) error {
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

	c.logger.Info("Processing analytics messages",
		zap.Int("count", len(messages)))

	// Process messages concurrently
	semaphore := make(chan struct{}, c.config.Analytics.MaxConcurrency)
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
	for i := 0; i < c.config.Analytics.MaxConcurrency; i++ {
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

// processMessage processes a single analytics message.
func (c *AnalyticsConsumer) processMessage(ctx context.Context, message *models.QueueMessage) error {
	c.logger.Info("Processing analytics message",
		zap.String("message_id", message.ID),
		zap.String("event_type", message.EventType))

	// Parse analytics event
	var analyticsEvent models.AnalyticsEvent
	if err := json.Unmarshal([]byte(message.Data), &analyticsEvent); err != nil {
		return fmt.Errorf("failed to unmarshal analytics event: %w", err)
	}

	// Process based on event type
	switch analyticsEvent.Type {
	case "metric":
		return c.processMetricEvent(ctx, &analyticsEvent)
	case "page_view":
		return c.processPageViewEvent(ctx, &analyticsEvent)
	case "user_action":
		return c.processUserActionEvent(ctx, &analyticsEvent)
	case "performance":
		return c.processPerformanceEvent(ctx, &analyticsEvent)
	case "error":
		return c.processErrorEvent(ctx, &analyticsEvent)
	default:
		return fmt.Errorf("unsupported analytics event type: %s", analyticsEvent.Type)
	}
}

// processMetricEvent processes a metric event.
func (c *AnalyticsConsumer) processMetricEvent(ctx context.Context, event *models.AnalyticsEvent) error {
	if !c.config.Analytics.ProcessingEnabled {
		c.logger.Warn("Analytics processing disabled, skipping",
			zap.String("event_id", event.ID))
		return nil
	}

	c.logger.Info("Processing metric event",
		zap.String("event_id", event.ID),
		zap.String("metric_name", event.MetricName))

	// Process metric
	if err := c.analyticsService.ProcessMetric(ctx, event); err != nil {
		return fmt.Errorf("failed to process metric: %w", err)
	}

	// Trigger aggregation if enabled
	if c.config.Analytics.AggregationEnabled {
		if err := c.aggregationService.AggregateMetric(ctx, event); err != nil {
			c.logger.Error("Failed to aggregate metric", zap.Error(err))
		}
	}

	c.logger.Info("Metric event processed successfully",
		zap.String("event_id", event.ID))

	return nil
}

// processPageViewEvent processes a page view event.
func (c *AnalyticsConsumer) processPageViewEvent(ctx context.Context, event *models.AnalyticsEvent) error {
	if !c.config.Analytics.ProcessingEnabled {
		c.logger.Warn("Analytics processing disabled, skipping",
			zap.String("event_id", event.ID))
		return nil
	}

	c.logger.Info("Processing page view event",
		zap.String("event_id", event.ID),
		zap.String("page", event.Page))

	// Process page view
	if err := c.analyticsService.ProcessPageView(ctx, event); err != nil {
		return fmt.Errorf("failed to process page view: %w", err)
	}

	// Trigger aggregation if enabled
	if c.config.Analytics.AggregationEnabled {
		if err := c.aggregationService.AggregatePageView(ctx, event); err != nil {
			c.logger.Error("Failed to aggregate page view", zap.Error(err))
		}
	}

	c.logger.Info("Page view event processed successfully",
		zap.String("event_id", event.ID))

	return nil
}

// processUserActionEvent processes a user action event.
func (c *AnalyticsConsumer) processUserActionEvent(ctx context.Context, event *models.AnalyticsEvent) error {
	if !c.config.Analytics.ProcessingEnabled {
		c.logger.Warn("Analytics processing disabled, skipping",
			zap.String("event_id", event.ID))
		return nil
	}

	c.logger.Info("Processing user action event",
		zap.String("event_id", event.ID),
		zap.String("action", event.Action))

	// Process user action
	if err := c.analyticsService.ProcessUserAction(ctx, event); err != nil {
		return fmt.Errorf("failed to process user action: %w", err)
	}

	// Trigger aggregation if enabled
	if c.config.Analytics.AggregationEnabled {
		if err := c.aggregationService.AggregateUserAction(ctx, event); err != nil {
			c.logger.Error("Failed to aggregate user action", zap.Error(err))
		}
	}

	c.logger.Info("User action event processed successfully",
		zap.String("event_id", event.ID))

	return nil
}

// processPerformanceEvent processes a performance event.
func (c *AnalyticsConsumer) processPerformanceEvent(ctx context.Context, event *models.AnalyticsEvent) error {
	if !c.config.Analytics.ProcessingEnabled {
		c.logger.Warn("Analytics processing disabled, skipping",
			zap.String("event_id", event.ID))
		return nil
	}

	c.logger.Info("Processing performance event",
		zap.String("event_id", event.ID),
		zap.String("metric_name", event.MetricName))

	// Process performance metric
	if err := c.analyticsService.ProcessPerformance(ctx, event); err != nil {
		return fmt.Errorf("failed to process performance: %w", err)
	}

	// Trigger aggregation if enabled
	if c.config.Analytics.AggregationEnabled {
		if err := c.aggregationService.AggregatePerformance(ctx, event); err != nil {
			c.logger.Error("Failed to aggregate performance", zap.Error(err))
		}
	}

	c.logger.Info("Performance event processed successfully",
		zap.String("event_id", event.ID))

	return nil
}

// processErrorEvent processes an error event.
func (c *AnalyticsConsumer) processErrorEvent(ctx context.Context, event *models.AnalyticsEvent) error {
	if !c.config.Analytics.ProcessingEnabled {
		c.logger.Warn("Analytics processing disabled, skipping",
			zap.String("event_id", event.ID))
		return nil
	}

	c.logger.Info("Processing error event",
		zap.String("event_id", event.ID),
		zap.String("error_type", event.ErrorType))

	// Process error
	if err := c.analyticsService.ProcessError(ctx, event); err != nil {
		return fmt.Errorf("failed to process error: %w", err)
	}

	// Trigger aggregation if enabled
	if c.config.Analytics.AggregationEnabled {
		if err := c.aggregationService.AggregateError(ctx, event); err != nil {
			c.logger.Error("Failed to aggregate error", zap.Error(err))
		}
	}

	c.logger.Info("Error event processed successfully",
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

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.QueueMessage{},
		&models.AnalyticsEvent{},
		&models.ProcessingLog{},
		&models.MetricData{},
		&models.AggregatedData{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

