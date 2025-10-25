// Package events handles RabbitMQ event publishing for the Monitoring Service.
// Package events handles RabbitMQ event publishing for the Monitoring Service.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// EventPublisher handles publishing monitoring events to RabbitMQ.
type EventPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// MonitoringCheckEvent represents a monitoring check result event.
type MonitoringCheckEvent struct {
	EventType    string    `json:"event_type"`    // monitoring.check.passed, monitoring.check.failed, monitoring.check.degraded
	TenantID     uuid.UUID `json:"tenant_id"`
	MonitorID    uint      `json:"monitor_id"`
	ComponentID  *uint     `json:"component_id,omitempty"`
	Location     string    `json:"location"`
	Status       string    `json:"status"`         // operational, degraded, down
	ResponseTime *int      `json:"response_time_ms,omitempty"`
	Error        string    `json:"error,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// SSLExpiringEvent represents an SSL certificate expiration warning event.
type SSLExpiringEvent struct {
	EventType      string    `json:"event_type"` // monitoring.ssl.expiring
	TenantID       uuid.UUID `json:"tenant_id"`
	CertificateID  uint      `json:"certificate_id"`
	Domain         string    `json:"domain"`
	DaysUntilExpiry int      `json:"days_until_expiry"`
	ValidUntil     time.Time `json:"valid_until"`
	WarningType    string    `json:"warning_type"` // 30d, 14d, 7d
	Timestamp      time.Time `json:"timestamp"`
}

// AutoIncidentEvent represents an auto-created incident from monitoring failure.
type AutoIncidentEvent struct {
	EventType       string    `json:"event_type"` // monitoring.incident.auto_created
	TenantID        uuid.UUID `json:"tenant_id"`
	MonitorID       uint      `json:"monitor_id"`
	ComponentID     *uint     `json:"component_id,omitempty"`
	IncidentTitle   string    `json:"incident_title"`
	IncidentSeverity string   `json:"incident_severity"` // critical, major, minor
	FailureCount    int       `json:"failure_count"`
	Timestamp       time.Time `json:"timestamp"`
}

// NewEventPublisher creates a new event publisher connected to RabbitMQ.
func NewEventPublisher(rabbitmqURL string) (*EventPublisher, error) {
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare the monitoring exchange (topic exchange) - passive declaration to verify it exists
	err = channel.ExchangeDeclarePassive(
		"monitoring.events", // exchange name (matches existing RabbitMQ definition)
		"topic",             // exchange type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to verify exchange: %w", err)
	}

	log.Println("✅ RabbitMQ event publisher initialized successfully")

	return &EventPublisher{
		conn:    conn,
		channel: channel,
	}, nil
}

// PublishCheckPassed publishes a successful monitoring check event.
func (p *EventPublisher) PublishCheckPassed(event MonitoringCheckEvent) error {
	event.EventType = "monitoring.check.passed"
	event.Timestamp = time.Now()
	return p.publish("monitoring.check.passed", event)
}

// PublishCheckFailed publishes a failed monitoring check event.
func (p *EventPublisher) PublishCheckFailed(event MonitoringCheckEvent) error {
	event.EventType = "monitoring.check.failed"
	event.Timestamp = time.Now()
	return p.publish("monitoring.check.failed", event)
}

// PublishCheckDegraded publishes a degraded monitoring check event.
func (p *EventPublisher) PublishCheckDegraded(event MonitoringCheckEvent) error {
	event.EventType = "monitoring.check.degraded"
	event.Timestamp = time.Now()
	return p.publish("monitoring.check.degraded", event)
}

// PublishSSLExpiring publishes an SSL certificate expiration warning event.
func (p *EventPublisher) PublishSSLExpiring(event SSLExpiringEvent) error {
	event.EventType = "monitoring.ssl.expiring"
	event.Timestamp = time.Now()
	return p.publish("monitoring.ssl.expiring", event)
}

// PublishAutoIncident publishes an auto-created incident event.
func (p *EventPublisher) PublishAutoIncident(event AutoIncidentEvent) error {
	event.EventType = "monitoring.incident.auto_created"
	event.Timestamp = time.Now()
	return p.publish("monitoring.incident.auto_created", event)
}

// publish is the internal method that publishes events to RabbitMQ.
func (p *EventPublisher) publish(routingKey string, event interface{}) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = p.channel.PublishWithContext(
		ctx,
		"monitoring.events", // exchange (matches existing RabbitMQ definition)
		routingKey,          // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // Make messages persistent
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("📤 Published event: %s", routingKey)
	return nil
}

// Close closes the RabbitMQ connection and channel.
func (p *EventPublisher) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			return err
		}
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

// HealthCheck checks if the RabbitMQ connection is still alive.
func (p *EventPublisher) HealthCheck() error {
	if p.conn == nil || p.conn.IsClosed() {
		return fmt.Errorf("RabbitMQ connection is closed")
	}
	if p.channel == nil {
		return fmt.Errorf("RabbitMQ channel is nil")
	}
	return nil
}
