// Package models provides webhook notification data models for the Monitoring Service.
package webhook

import (
	"time"

	"gorm.io/gorm"
)

// WebhookEndpoint represents a webhook endpoint configuration.
type WebhookEndpoint struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	URL         string         `gorm:"not null" json:"url"`
	SecretKey   string         `json:"secret_key"`          // For HMAC signing
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	Events      string         `gorm:"type:text" json:"events"` // JSON array of event types
	Headers     string         `gorm:"type:text" json:"headers"` // JSON object of custom headers
	Timeout     int            `gorm:"default:30" json:"timeout"` // Request timeout in seconds
	RetryCount  int            `gorm:"default:3" json:"retry_count"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data

	// Related entities
	Deliveries []WebhookDelivery `gorm:"foreignKey:WebhookID" json:"deliveries,omitempty"`
}

// WebhookDelivery represents a webhook delivery attempt.
type WebhookDelivery struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	WebhookID     uint           `gorm:"not null;index" json:"webhook_id"`
	Webhook       WebhookEndpoint `gorm:"foreignKey:WebhookID" json:"webhook"`
	EventType     string         `gorm:"not null" json:"event_type"`
	EventID       string         `gorm:"not null" json:"event_id"`        // Reference to the event source
	Payload       string         `gorm:"type:text" json:"payload"`        // JSON payload sent
	RequestHeaders string        `gorm:"type:text" json:"request_headers"` // JSON object of sent headers
	ResponseStatus int           `json:"response_status"`
	ResponseHeaders string       `gorm:"type:text" json:"response_headers"` // JSON object of received headers
	ResponseBody   string        `gorm:"type:text" json:"response_body"`
	Duration       int64         `json:"duration"`                     // Request duration in milliseconds
	Success        bool          `json:"success"`
	ErrorMessage   string        `json:"error_message"`
	AttemptCount   int           `gorm:"default:1" json:"attempt_count"`
	NextRetryAt    *time.Time    `json:"next_retry_at"`
	DeliveredAt    *time.Time    `json:"delivered_at"`
	Metadata       string        `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// WebhookEvent represents the payload structure for webhook events.
type WebhookEvent struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Timestamp  time.Time              `json:"timestamp"`
	TenantID   uint                   `json:"tenant_id"`
	Source     string                 `json:"source"`     // monitoring-service, incident-service, etc.
	Action     string                 `json:"action"`     // created, updated, deleted, status_changed
	Data       map[string]interface{} `json:"data"`       // Event-specific data
	Previous   map[string]interface{} `json:"previous,omitempty"` // Previous state for updates
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ComponentStatusEvent represents component status change events.
type ComponentStatusEvent struct {
	ComponentID   uint      `json:"component_id"`
	ComponentName string    `json:"component_name"`
	OldStatus     string    `json:"old_status"`
	NewStatus     string    `json:"new_status"`
	Timestamp     time.Time `json:"timestamp"`
	CheckType     string    `json:"check_type"`
	Message       string    `json:"message"`
	ResponseTime  float64   `json:"response_time,omitempty"`
	LastCheck     time.Time `json:"last_check"`
}

// IncidentEvent represents incident-related events.
type IncidentEvent struct {
	IncidentID    uint      `json:"incident_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	Severity      string    `json:"severity"`
	Components    []string  `json:"components"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty"`
	Duration      int64     `json:"duration,omitempty"` // in seconds
}

// MaintenanceEvent represents maintenance-related events.
type MaintenanceEvent struct {
	MaintenanceID uint      `json:"maintenance_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	Impact        string    `json:"impact"`
	Components    []string  `json:"components"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// MetricEvent represents metric-related events.
type MetricEvent struct {
	MetricID    uint                   `json:"metric_id"`
	MetricName  string                 `json:"metric_name"`
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit"`
	Timestamp   time.Time              `json:"timestamp"`
	Tags        map[string]string      `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	Threshold   *float64               `json:"threshold,omitempty"`
	IsAlert     bool                   `json:"is_alert"`
}

// WebhookEventType constants for supported event types.
const (
	EventTypeComponentStatusChanged = "component.status_changed"
	EventTypeComponentCreated       = "component.created"
	EventTypeComponentUpdated       = "component.updated"
	EventTypeComponentDeleted       = "component.deleted"

	EventTypeIncidentCreated  = "incident.created"
	EventTypeIncidentUpdated  = "incident.updated"
	EventTypeIncidentResolved = "incident.resolved"

	EventTypeMaintenanceScheduled = "maintenance.scheduled"
	EventTypeMaintenanceStarted   = "maintenance.started"
	EventTypeMaintenanceCompleted = "maintenance.completed"
	EventTypeMaintenanceCancelled = "maintenance.cancelled"

	EventTypeMetricThresholdExceeded = "metric.threshold_exceeded"
	EventTypeMetricDataReceived      = "metric.data_received"

	EventTypeMonitorCheckFailed  = "monitor.check_failed"
	EventTypeMonitorCheckPassed  = "monitor.check_passed"
	EventTypeMonitorCreated      = "monitor.created"
	EventTypeMonitorUpdated      = "monitor.updated"
	EventTypeMonitorDeleted      = "monitor.deleted"
)

// Table name functions
func (WebhookEndpoint) TableName() string {
	return "webhook_endpoints"
}

func (WebhookDelivery) TableName() string {
	return "webhook_deliveries"
}

// Validation methods
func (w *WebhookEndpoint) IsEventSubscribed(eventType string) bool {
	// This would parse the Events JSON field and check if eventType is included
	// Implementation would unmarshal the JSON and check the array
	return true // Simplified for now
}

// GetActiveEndpoints returns active webhook endpoints for a tenant that subscribe to the given event type.
func GetActiveWebhookEndpoints(db *gorm.DB, tenantID uint, eventType string) ([]WebhookEndpoint, error) {
	var endpoints []WebhookEndpoint

	// For now, return all active endpoints for the tenant
	// In a full implementation, this would filter by event subscription
	err := db.Where("tenant_id = ? AND is_active = ?", tenantID, true).Find(&endpoints).Error

	return endpoints, err
}