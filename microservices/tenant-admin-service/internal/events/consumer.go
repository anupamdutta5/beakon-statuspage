package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	TenantExchange      = "tenant.events"
	TenantCreatedQueue  = "tenant.created.queue"
	TenantUpdatedQueue  = "tenant.updated.queue"
	TenantDeletedQueue  = "tenant.deleted.queue"
	TenantRestoredQueue = "tenant.restored.queue"
)

// EventHandler is the interface for handling tenant events
type RabbitMQEventHandler interface {
	HandleTenantCreated(ctx context.Context, event RabbitMQTenantEvent) error
	HandleTenantUpdated(ctx context.Context, event RabbitMQTenantEvent) error
	HandleTenantDeleted(ctx context.Context, event RabbitMQTenantEvent) error
	HandleTenantRestored(ctx context.Context, event RabbitMQTenantEvent) error
}

// Consumer handles consuming events from RabbitMQ
type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	handler RabbitMQEventHandler
	logger  *zap.Logger
	done    chan struct{}
}

// ConsumerConfig contains configuration for the consumer
type ConsumerConfig struct {
	URL     string
	Handler RabbitMQEventHandler
	Logger  *zap.Logger
}

// NewConsumer creates a new event consumer
func NewConsumer(config ConsumerConfig) (*Consumer, error) {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS - prefetch count of 1 for fair dispatch
	if err := channel.Qos(1, 0, false); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	consumer := &Consumer{
		conn:    conn,
		channel: channel,
		handler: config.Handler,
		logger:  config.Logger,
		done:    make(chan struct{}),
	}

	config.Logger.Info("RabbitMQ consumer initialized successfully")
	return consumer, nil
}

// Start begins consuming messages from all tenant queues
func (c *Consumer) Start(ctx context.Context) error {
	// Start consuming from each queue in separate goroutines
	queues := map[string]RabbitMQEventType{
		TenantCreatedQueue:  RabbitMQTenantCreated,
		TenantUpdatedQueue:  RabbitMQTenantUpdated,
		TenantDeletedQueue:  RabbitMQTenantDeleted,
		TenantRestoredQueue: RabbitMQTenantRestored,
	}

	for queueName, eventType := range queues {
		go func(queue string, evtType RabbitMQEventType) {
			if err := c.consumeQueue(ctx, queue, evtType); err != nil {
				c.logger.Error("Error consuming queue",
					zap.String("queue", queue),
					zap.Error(err))
			}
		}(queueName, eventType)
	}

	c.logger.Info("Started consuming tenant events from all queues")

	// Wait for context cancellation or done signal
	select {
	case <-ctx.Done():
		c.logger.Info("Context cancelled, stopping consumer")
		return ctx.Err()
	case <-c.done:
		c.logger.Info("Consumer stopped")
		return nil
	}
}

// consumeQueue consumes messages from a specific queue
func (c *Consumer) consumeQueue(ctx context.Context, queueName string, eventType RabbitMQEventType) error {
	// Start consuming messages
	msgs, err := c.channel.Consume(
		queueName, // queue
		"",        // consumer tag (empty = auto-generated)
		false,     // auto-ack (we'll manually ack)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer for queue %s: %w", queueName, err)
	}

	c.logger.Info("Consuming messages from queue",
		zap.String("queue", queueName),
		zap.String("event_type", string(eventType)))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				c.logger.Warn("Message channel closed", zap.String("queue", queueName))
				return fmt.Errorf("message channel closed for queue %s", queueName)
			}

			// Process message
			if err := c.processMessage(ctx, msg, eventType); err != nil {
				c.logger.Error("Failed to process message",
					zap.String("queue", queueName),
					zap.String("message_id", msg.MessageId),
					zap.Error(err))

				// Negative acknowledgment with requeue
				// On repeated failures, message will go to DLX
				if nackErr := msg.Nack(false, true); nackErr != nil {
					c.logger.Error("Failed to nack message", zap.Error(nackErr))
				}
			} else {
				// Acknowledge successful processing
				if ackErr := msg.Ack(false); ackErr != nil {
					c.logger.Error("Failed to ack message", zap.Error(ackErr))
				}
			}
		}
	}
}

// processMessage processes a single message
func (c *Consumer) processMessage(ctx context.Context, msg amqp.Delivery, eventType RabbitMQEventType) error {
	// Add timeout to message processing
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Parse message
	var event RabbitMQTenantEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		c.logger.Error("Failed to unmarshal event",
			zap.String("message_id", msg.MessageId),
			zap.Error(err))
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	c.logger.Info("Processing tenant event",
		zap.String("event_id", event.EventID),
		zap.String("event_type", string(event.EventType)),
		zap.String("tenant_id", event.TenantID.String()),
		zap.String("correlation_id", event.Metadata.CorrelationID))

	// Route to appropriate handler based on event type
	var err error
	switch event.EventType {
	case RabbitMQTenantCreated:
		err = c.handler.HandleTenantCreated(ctx, event)
	case RabbitMQTenantUpdated:
		err = c.handler.HandleTenantUpdated(ctx, event)
	case RabbitMQTenantDeleted:
		err = c.handler.HandleTenantDeleted(ctx, event)
	case RabbitMQTenantRestored:
		err = c.handler.HandleTenantRestored(ctx, event)
	default:
		c.logger.Warn("Unknown event type",
			zap.String("event_type", string(event.EventType)),
			zap.String("event_id", event.EventID))
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}

	if err != nil {
		c.logger.Error("Handler failed to process event",
			zap.String("event_id", event.EventID),
			zap.String("event_type", string(event.EventType)),
			zap.Error(err))
		return err
	}

	c.logger.Info("Successfully processed tenant event",
		zap.String("event_id", event.EventID),
		zap.String("event_type", string(event.EventType)),
		zap.String("tenant_id", event.TenantID.String()))

	return nil
}

// Close closes the consumer connection
func (c *Consumer) Close() error {
	close(c.done)

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			c.logger.Error("Error closing channel", zap.Error(err))
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			c.logger.Error("Error closing connection", zap.Error(err))
			return err
		}
	}
	c.logger.Info("RabbitMQ consumer closed successfully")
	return nil
}

// IsConnected checks if the consumer is still connected
func (c *Consumer) IsConnected() bool {
	return c.conn != nil && !c.conn.IsClosed()
}
