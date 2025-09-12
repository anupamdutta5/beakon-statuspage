// Package models provides data models for the Component Service.
package models

import (
	"time"

	"gorm.io/gorm"
)

// Component represents a status page component.
type Component struct {
	ID          uint            `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint            `gorm:"not null;index" json:"tenant_id"`
	Name        string          `gorm:"not null" json:"name"`
	Description string          `json:"description"`
	Status      string          `gorm:"default:operational" json:"status"` // operational, degraded_performance, partial_outage, major_outage, maintenance
	Position    int             `gorm:"default:0" json:"position"`
	IsVisible   bool            `gorm:"default:true" json:"is_visible"`
	GroupID     *uint           `gorm:"index" json:"group_id"`
	Group       *ComponentGroup `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	Metadata    string          `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// ComponentGroup represents a group of components.
type ComponentGroup struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	Position    int            `gorm:"default:0" json:"position"`
	IsVisible   bool           `gorm:"default:true" json:"is_visible"`
	Components  []Component    `gorm:"foreignKey:GroupID" json:"components,omitempty"`
}

// ComponentStatus represents the status of a component.
type ComponentStatus struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	ComponentID uint           `gorm:"not null;index" json:"component_id"`
	Component   Component      `gorm:"foreignKey:ComponentID" json:"component"`
	Status      string         `gorm:"not null" json:"status"`
	Message     string         `json:"message"`
	UpdatedBy   uint           `json:"updated_by"`                // User ID who updated the status
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// ComponentHistory represents the history of component status changes.
type ComponentHistory struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	ComponentID uint           `gorm:"not null;index" json:"component_id"`
	Component   Component      `gorm:"foreignKey:ComponentID" json:"component"`
	OldStatus   string         `json:"old_status"`
	NewStatus   string         `gorm:"not null" json:"new_status"`
	Message     string         `json:"message"`
	UpdatedBy   uint           `json:"updated_by"`                // User ID who made the change
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// ComponentMetric represents metrics for a component.
type ComponentMetric struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	ComponentID uint           `gorm:"not null;index" json:"component_id"`
	Component   Component      `gorm:"foreignKey:ComponentID" json:"component"`
	MetricType  string         `gorm:"not null" json:"metric_type"` // uptime, response_time, error_rate, etc.
	Value       float64        `gorm:"not null" json:"value"`
	Unit        string         `json:"unit"` // percentage, milliseconds, count, etc.
	Timestamp   time.Time      `gorm:"not null;index" json:"timestamp"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// ComponentAlert represents alerts for component status changes.
type ComponentAlert struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	ComponentID  uint           `gorm:"not null;index" json:"component_id"`
	Component    Component      `gorm:"foreignKey:ComponentID" json:"component"`
	AlertType    string         `gorm:"not null" json:"alert_type"` // status_change, threshold_breach, etc.
	Status       string         `gorm:"not null" json:"status"`     // active, resolved, dismissed
	Message      string         `json:"message"`
	Threshold    float64        `json:"threshold"`
	CurrentValue float64        `json:"current_value"`
	ResolvedAt   *time.Time     `json:"resolved_at"`
	ResolvedBy   *uint          `json:"resolved_by"`
	Metadata     string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// ComponentWebhook represents webhook configurations for components.
type ComponentWebhook struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	ComponentID   uint           `gorm:"not null;index" json:"component_id"`
	Component     Component      `gorm:"foreignKey:ComponentID" json:"component"`
	URL           string         `gorm:"not null" json:"url"`
	Events        string         `gorm:"type:text" json:"events"` // JSON array of events to trigger on
	Secret        string         `json:"secret"`                  // Webhook secret for verification
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	LastTriggered *time.Time     `json:"last_triggered"`
	Metadata      string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for Component.
func (Component) TableName() string {
	return "components"
}

// TableName returns the table name for ComponentGroup.
func (ComponentGroup) TableName() string {
	return "component_groups"
}

// TableName returns the table name for ComponentStatus.
func (ComponentStatus) TableName() string {
	return "component_statuses"
}

// TableName returns the table name for ComponentHistory.
func (ComponentHistory) TableName() string {
	return "component_history"
}

// TableName returns the table name for ComponentMetric.
func (ComponentMetric) TableName() string {
	return "component_metrics"
}

// TableName returns the table name for ComponentAlert.
func (ComponentAlert) TableName() string {
	return "component_alerts"
}

// TableName returns the table name for ComponentWebhook.
func (ComponentWebhook) TableName() string {
	return "component_webhooks"
}
