// Package models provides data models for the Incident Service.
package models

import (
	"time"

	"gorm.io/gorm"
)

// Incident represents an incident in the system.
type Incident struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Status      string         `gorm:"default:investigating" json:"status"` // investigating, identified, monitoring, resolved
	Impact      string         `gorm:"default:minor" json:"impact"`         // minor, major, critical
	Severity    string         `gorm:"default:low" json:"severity"`         // low, medium, high, critical
	IsVisible   bool           `gorm:"default:true" json:"is_visible"`
	StartedAt   time.Time      `gorm:"not null" json:"started_at"`
	ResolvedAt  *time.Time     `json:"resolved_at"`
	CreatedBy   uint           `json:"created_by"`                // User ID who created the incident
	UpdatedBy   uint           `json:"updated_by"`                // User ID who last updated the incident
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data

	// Related components
	Components []IncidentComponent `gorm:"foreignKey:IncidentID" json:"components,omitempty"`

	// Incident updates
	Updates []IncidentUpdate `gorm:"foreignKey:IncidentID" json:"updates,omitempty"`
}

// IncidentComponent represents the relationship between incidents and components.
type IncidentComponent struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	IncidentID    uint           `gorm:"not null;index" json:"incident_id"`
	Incident      Incident       `gorm:"foreignKey:IncidentID" json:"incident"`
	ComponentID   uint           `gorm:"not null;index" json:"component_id"`
	ComponentName string         `gorm:"not null" json:"component_name"` // Denormalized for performance
	Status        string         `gorm:"not null" json:"status"`         // operational, degraded_performance, partial_outage, major_outage
	CreatedBy     uint           `json:"created_by"`
	Metadata      string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// IncidentUpdate represents updates to an incident.
type IncidentUpdate struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	IncidentID uint           `gorm:"not null;index" json:"incident_id"`
	Incident   Incident       `gorm:"foreignKey:IncidentID" json:"incident"`
	Status     string         `gorm:"not null" json:"status"` // investigating, identified, monitoring, resolved
	Message    string         `gorm:"type:text;not null" json:"message"`
	IsVisible  bool           `gorm:"default:true" json:"is_visible"`
	CreatedBy  uint           `json:"created_by"`                // User ID who created the update
	Metadata   string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// IncidentTemplate represents templates for creating incidents.
type IncidentTemplate struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	Title       string         `gorm:"not null" json:"title"`
	Message     string         `gorm:"type:text" json:"message"`
	Impact      string         `gorm:"default:minor" json:"impact"`
	Severity    string         `gorm:"default:low" json:"severity"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedBy   uint           `json:"created_by"`
	UpdatedBy   uint           `json:"updated_by"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// IncidentNotification represents notifications sent for incidents.
type IncidentNotification struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	IncidentID  uint           `gorm:"not null;index" json:"incident_id"`
	Incident    Incident       `gorm:"foreignKey:IncidentID" json:"incident"`
	Type        string         `gorm:"not null" json:"type"` // email, sms, webhook, slack, etc.
	Recipient   string         `gorm:"not null" json:"recipient"`
	Subject     string         `json:"subject"`
	Message     string         `gorm:"type:text" json:"message"`
	Status      string         `gorm:"default:pending" json:"status"` // pending, sent, failed, delivered
	SentAt      *time.Time     `json:"sent_at"`
	DeliveredAt *time.Time     `json:"delivered_at"`
	Error       string         `json:"error"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// IncidentMetric represents metrics for incidents.
type IncidentMetric struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	IncidentID uint           `gorm:"not null;index" json:"incident_id"`
	Incident   Incident       `gorm:"foreignKey:IncidentID" json:"incident"`
	MetricType string         `gorm:"not null" json:"metric_type"` // mttr, mtbf, severity_distribution, etc.
	Value      float64        `gorm:"not null" json:"value"`
	Unit       string         `json:"unit"` // minutes, hours, count, percentage, etc.
	Timestamp  time.Time      `gorm:"not null;index" json:"timestamp"`
	Metadata   string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// IncidentSubscriber represents subscribers to incident notifications.
type IncidentSubscriber struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	Email       string         `gorm:"not null" json:"email"`
	Phone       string         `json:"phone"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	Preferences string         `gorm:"type:text" json:"preferences"` // JSON string for notification preferences
	Metadata    string         `gorm:"type:text" json:"metadata"`    // JSON string for additional data
}

// TableName returns the table name for Incident.
func (Incident) TableName() string {
	return "incidents"
}

// TableName returns the table name for IncidentComponent.
func (IncidentComponent) TableName() string {
	return "incident_components"
}

// TableName returns the table name for IncidentUpdate.
func (IncidentUpdate) TableName() string {
	return "incident_updates"
}

// TableName returns the table name for IncidentTemplate.
func (IncidentTemplate) TableName() string {
	return "incident_templates"
}

// TableName returns the table name for IncidentNotification.
func (IncidentNotification) TableName() string {
	return "incident_notifications"
}

// TableName returns the table name for IncidentMetric.
func (IncidentMetric) TableName() string {
	return "incident_metrics"
}

// TableName returns the table name for IncidentSubscriber.
func (IncidentSubscriber) TableName() string {
	return "incident_subscribers"
}
