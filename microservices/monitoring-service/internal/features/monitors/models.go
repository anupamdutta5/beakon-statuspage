// Package models provides data models for the Monitoring Service.
package monitors

import (
	"time"

	"github.com/google/uuid"
)

// Monitor represents a health check/monitoring configuration for a component.
type Monitor struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	TenantID    uuid.UUID `gorm:"type:uuid;not null;index:idx_monitors_tenant" json:"tenant_id"`
	ComponentID uuid.UUID `gorm:"type:uuid;not null;index:idx_monitors_component" json:"component_id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	MonitorType string    `gorm:"size:50;not null" json:"monitor_type"` // http, ping, tcp, ssl, heartbeat

	// HTTP/URL monitoring
	CheckURL        string `gorm:"size:1000" json:"check_url,omitempty"`
	CheckIntervalSeconds int `gorm:"not null;default:60" json:"check_interval_seconds"`
	TimeoutSeconds  int    `gorm:"not null;default:30" json:"timeout_seconds"`

	// HTTP-specific settings
	HTTPMethod           string `gorm:"size:10;default:GET" json:"http_method,omitempty"`
	HTTPHeaders          string `gorm:"type:text" json:"http_headers,omitempty"`          // JSON
	HTTPBody             string `gorm:"type:text" json:"http_body,omitempty"`
	ExpectedStatusCodes  string `gorm:"default:200,201,204" json:"expected_status_codes"` // Comma-separated

	// Advanced settings
	FollowRedirects bool `gorm:"default:true" json:"follow_redirects"`
	VerifySSL       bool `gorm:"default:true" json:"verify_ssl"`

	// Multi-location monitoring
	EnabledLocations string `gorm:"type:text" json:"enabled_locations,omitempty"` // JSON array: [1,2,3]

	// Auto-incident creation settings
	AutoCreateIncidents bool  `gorm:"default:true" json:"auto_create_incidents"`
	FailureThreshold    int   `gorm:"default:3" json:"failure_threshold"`
	ConsecutiveFailures int   `gorm:"default:0;index:idx_monitors_failures" json:"consecutive_failures"`
	LastIncidentID      *uuid.UUID `gorm:"type:uuid" json:"last_incident_id,omitempty"`

	// Status tracking
	IsActive         bool       `gorm:"default:true;index:idx_monitors_active" json:"is_active"`
	CurrentStatus    string     `gorm:"size:20;default:unknown;index:idx_monitors_status" json:"current_status"` // operational, degraded, down, unknown
	LastCheckAt      *time.Time `json:"last_check_at,omitempty"`
	LastSuccessAt    *time.Time `json:"last_success_at,omitempty"`
	LastFailureAt    *time.Time `json:"last_failure_at,omitempty"`
	UptimePercentage float64    `gorm:"type:decimal(5,2);default:100.00" json:"uptime_percentage"`

	// Maintenance windows
	InMaintenance    bool       `gorm:"default:false" json:"in_maintenance"`
	MaintenanceUntil *time.Time `json:"maintenance_until,omitempty"`
}

// TableName specifies the table name for GORM.
func (Monitor) TableName() string {
	return "monitors"
}

// ShouldCreateIncident returns true if an incident should be auto-created.
func (m *Monitor) ShouldCreateIncident() bool {
	return m.AutoCreateIncidents &&
		m.ConsecutiveFailures >= m.FailureThreshold &&
		m.LastIncidentID == nil && // No active incident
		!m.InMaintenance
}

// ShouldResolveIncident returns true if the active incident should be auto-resolved.
func (m *Monitor) ShouldResolveIncident() bool {
	return m.AutoCreateIncidents &&
		m.LastIncidentID != nil && // Has active incident
		m.CurrentStatus == "operational" &&
		!m.InMaintenance
}

// IncrementFailures increments the consecutive failure counter.
func (m *Monitor) IncrementFailures() {
	m.ConsecutiveFailures++
	m.CurrentStatus = "down"
	now := time.Now()
	m.LastFailureAt = &now
	m.LastCheckAt = &now
}

// ResetFailures resets the consecutive failure counter.
func (m *Monitor) ResetFailures() {
	m.ConsecutiveFailures = 0
	m.CurrentStatus = "operational"
	now := time.Now()
	m.LastSuccessAt = &now
	m.LastCheckAt = &now
}

// SetDegraded marks the monitor as degraded (partial failure).
func (m *Monitor) SetDegraded() {
	m.CurrentStatus = "degraded"
	now := time.Now()
	m.LastCheckAt = &now
}

// AutoIncident represents an auto-created incident tracking record.
type AutoIncident struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	TenantID    uuid.UUID `gorm:"type:uuid;not null;index:idx_auto_incidents_tenant" json:"tenant_id"`
	MonitorID   uint      `gorm:"not null;index:idx_auto_incidents_monitor" json:"monitor_id"`
	IncidentID  uuid.UUID `gorm:"type:uuid;not null;index:idx_auto_incidents_incident" json:"incident_id"`
	ComponentID uuid.UUID `gorm:"type:uuid;not null" json:"component_id"`

	// Creation details
	TriggeredByResultID *uint  `json:"triggered_by_result_id,omitempty"`
	FailureCount        int    `gorm:"not null" json:"failure_count"`

	// Resolution tracking
	Resolved           bool       `gorm:"default:false;index:idx_auto_incidents_unresolved" json:"resolved"`
	ResolvedAt         *time.Time `json:"resolved_at,omitempty"`
	ResolvedByResultID *uint      `json:"resolved_by_result_id,omitempty"`
	AutoResolved       bool       `gorm:"default:false" json:"auto_resolved"`

	// Metadata
	ErrorMessage      string `gorm:"type:text" json:"error_message,omitempty"`
	AffectedLocations string `gorm:"type:text" json:"affected_locations,omitempty"` // JSON array
}

// TableName specifies the table name for GORM.
func (AutoIncident) TableName() string {
	return "auto_incidents"
}

// MonitorNotification represents notification preferences for a monitor.
type MonitorNotification struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	MonitorID uint `gorm:"not null;index:idx_monitor_notifications_monitor" json:"monitor_id"`

	// Notification channels
	NotifyEmail   bool `gorm:"default:true" json:"notify_email"`
	NotifySMS     bool `gorm:"default:false" json:"notify_sms"`
	NotifySlack   bool `gorm:"default:false" json:"notify_slack"`
	NotifyWebhook bool `gorm:"default:false" json:"notify_webhook"`

	// Notification triggers
	NotifyOnFailure  bool `gorm:"default:true" json:"notify_on_failure"`
	NotifyOnRecovery bool `gorm:"default:true" json:"notify_on_recovery"`
	NotifyOnDegraded bool `gorm:"default:false" json:"notify_on_degraded"`

	// Escalation
	EscalationPolicyID *uint `json:"escalation_policy_id,omitempty"`
	OnCallScheduleID   *uint `json:"on_call_schedule_id,omitempty"`

	// Custom recipients
	CustomEmailRecipients string `gorm:"type:text" json:"custom_email_recipients,omitempty"` // JSON array
	CustomWebhookURL      string `gorm:"size:1000" json:"custom_webhook_url,omitempty"`
}

// TableName specifies the table name for GORM.
func (MonitorNotification) TableName() string {
	return "monitor_notifications"
}

// MonitorStatusHistory tracks historical status changes for monitors.
type MonitorStatusHistory struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ChangedAt time.Time `gorm:"not null;index:idx_status_history_monitor,priority:2;index:idx_status_history_tenant,priority:2" json:"changed_at"`

	MonitorID uint      `gorm:"not null;index:idx_status_history_monitor,priority:1" json:"monitor_id"`
	TenantID  uuid.UUID `gorm:"type:uuid;not null;index:idx_status_history_tenant,priority:1" json:"tenant_id"`

	// Status change
	PreviousStatus string `gorm:"size:20" json:"previous_status,omitempty"`
	NewStatus      string `gorm:"size:20;not null" json:"new_status"`

	// Timing
	DurationSeconds *int64 `json:"duration_seconds,omitempty"` // How long previous status lasted

	// Context
	TriggeredBy  string `gorm:"size:50" json:"triggered_by"` // health_check, manual, maintenance, auto_recovery
	ErrorMessage string `gorm:"type:text" json:"error_message,omitempty"`

	// Metadata
	Metadata string `gorm:"type:text" json:"metadata,omitempty"` // JSON
}

// TableName specifies the table name for GORM.
func (MonitorStatusHistory) TableName() string {
	return "monitor_status_history"
}
