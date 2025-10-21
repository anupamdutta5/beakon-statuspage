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
	TenantExchange    = "tenant.events"
	TenantDLXExchange = "tenant.events.dlx"
)

// Publisher handles publishing events to RabbitMQ
type Publisher struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	confirms chan amqp.Confirmation
	logger   *zap.Logger
}

// PublisherConfig contains configuration for the publisher
type PublisherConfig struct {
	URL    string
	Logger *zap.Logger
}

// NewPublisher creates a new event publisher
func NewPublisher(config PublisherConfig) (*Publisher, error) {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Enable publisher confirms
	if err := channel.Confirm(false); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to enable publisher confirms: %w", err)
	}

	// Set up confirmation channel
	confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	publisher := &Publisher{
		conn:     conn,
		channel:  channel,
		confirms: confirms,
		logger:   config.Logger,
	}

	// Ensure exchange exists (should already be declared via definitions.json)
	if err := publisher.declareExchange(); err != nil {
		publisher.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	config.Logger.Info("RabbitMQ publisher initialized successfully")
	return publisher, nil
}

// declareExchange ensures the tenant events exchange exists
func (p *Publisher) declareExchange() error {
	return p.channel.ExchangeDeclare(
		TenantExchange, // name
		"topic",        // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
}

// PublishTenantEvent publishes a tenant event to RabbitMQ
func (p *Publisher) PublishTenantEvent(ctx context.Context, event TenantEvent) error {
	// Marshal event to JSON
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish with confirmation
	publishing := amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent, // Make message persistent
		Timestamp:    event.Timestamp,
		MessageId:    event.EventID,
		Type:         string(event.EventType),
		Headers: amqp.Table{
			"x-event-type":    string(event.EventType),
			"x-tenant-id":     event.TenantID.String(),
			"x-correlation-id": event.Metadata.CorrelationID,
			"x-source":        event.Metadata.Source,
		},
	}

	// Get routing key based on event type
	routingKey := string(event.EventType)

	// Publish message
	if err := p.channel.PublishWithContext(
		ctx,
		TenantExchange, // exchange
		routingKey,     // routing key
		true,           // mandatory
		false,          // immediate
		publishing,
	); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	// Wait for confirmation
	select {
	case confirm := <-p.confirms:
		if !confirm.Ack {
			return fmt.Errorf("message not acknowledged by broker")
		}
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for publish confirmation")
	case <-ctx.Done():
		return ctx.Err()
	}

	p.logger.Info("Successfully published tenant event",
		zap.String("event_type", string(event.EventType)),
		zap.String("tenant_id", event.TenantID.String()),
		zap.String("event_id", event.EventID),
	)

	return nil
}

// Close closes the publisher connection
func (p *Publisher) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			p.logger.Error("Error closing channel", zap.Error(err))
		}
	}
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			p.logger.Error("Error closing connection", zap.Error(err))
			return err
		}
	}
	p.logger.Info("RabbitMQ publisher closed successfully")
	return nil
}

// IsConnected checks if the publisher is still connected
func (p *Publisher) IsConnected() bool {
	return p.conn != nil && !p.conn.IsClosed()
}

// Reconnect attempts to reconnect to RabbitMQ
func (p *Publisher) Reconnect(url string) error {
	p.logger.Info("Attempting to reconnect to RabbitMQ")

	// Close existing connections
	p.Close()

	// Create new connection
	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to reconnect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Enable publisher confirms
	if err := channel.Confirm(false); err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("failed to enable publisher confirms: %w", err)
	}

	p.conn = conn
	p.channel = channel

	// Declare exchange
	if err := p.declareExchange(); err != nil {
		p.Close()
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	p.logger.Info("Successfully reconnected to RabbitMQ")
	return nil
}
