// Package models provides component and container monitoring data models.
package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// MonitoredComponent represents a system component being monitored.
type MonitoredComponent struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	Name        string         `gorm:"not null;size:255" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Type        string         `gorm:"not null;size:50" json:"type"`          // web_service, api, mobile_app, database, cache, queue
	Status      string         `gorm:"default:unknown;size:50" json:"status"` // operational, degraded, partial_outage, major_outage, maintenance
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data

	// Related entities
	Containers []MonitoredContainer `gorm:"foreignKey:ComponentID;constraint:OnDelete:CASCADE" json:"containers,omitempty"`
	Metrics    []ComponentMetric    `gorm:"foreignKey:ComponentID;constraint:OnDelete:CASCADE" json:"metrics,omitempty"`
}

// MonitoredContainer represents a sub-component or container within a component.
type MonitoredContainer struct {
	ID          uint               `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	DeletedAt   gorm.DeletedAt     `gorm:"index" json:"deleted_at,omitempty"`
	ComponentID uint               `gorm:"not null;index" json:"component_id"`
	Component   MonitoredComponent `gorm:"foreignKey:ComponentID" json:"component"`
	Name        string             `gorm:"not null;size:255" json:"name"`
	Description string             `gorm:"type:text" json:"description"`
	Type        string             `gorm:"not null;size:50" json:"type"`          // docker_container, kubernetes_pod, service, location
	Location    string             `gorm:"size:100" json:"location"`              // region, datacenter, availability_zone
	Status      string             `gorm:"default:unknown;size:50" json:"status"` // healthy, unhealthy, unknown, maintenance
	IsActive    bool               `gorm:"default:true" json:"is_active"`
	Metadata    string             `gorm:"type:text" json:"metadata"` // JSON string for additional data

	// Related entities
	HealthChecks []ContainerHealthCheck `gorm:"foreignKey:ContainerID;constraint:OnDelete:CASCADE" json:"health_checks,omitempty"`
}

// ContainerHealthCheck represents a health check for a container.
type ContainerHealthCheck struct {
	ID          uint               `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	DeletedAt   gorm.DeletedAt     `gorm:"index" json:"deleted_at,omitempty"`
	ContainerID uint               `gorm:"not null;index" json:"container_id"`
	Container   MonitoredContainer `gorm:"foreignKey:ContainerID" json:"container"`
	Name        string             `gorm:"not null;size:255" json:"name"`
	Type        string             `gorm:"not null;size:50" json:"type"`          // readiness, liveness, startup
	Status      string             `gorm:"default:unknown;size:50" json:"status"` // success, failure, unknown
	LastChecked *time.Time         `json:"last_checked"`
	Message     string             `gorm:"type:text" json:"message"`
	Metadata    string             `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// ComponentMetric represents a metric for a component.
type ComponentMetric struct {
	ID          uint               `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	DeletedAt   gorm.DeletedAt     `gorm:"index" json:"deleted_at,omitempty"`
	ComponentID uint               `gorm:"not null;index" json:"component_id"`
	Component   MonitoredComponent `gorm:"foreignKey:ComponentID" json:"component"`
	Name        string             `gorm:"not null;size:255" json:"name"`
	Value       float64            `gorm:"not null" json:"value"`
	Unit        string             `gorm:"size:50" json:"unit"` // percentage, milliseconds, bytes, count, etc.
	Timestamp   time.Time          `gorm:"not null;index" json:"timestamp"`
	Labels      string             `gorm:"type:text" json:"labels"`   // JSON string for metric labels
	Source      string             `gorm:"size:100" json:"source"`    // agent, api, webhook, etc.
	Metadata    string             `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for MonitoredComponent.
func (MonitoredComponent) TableName() string {
	return "monitored_components"
}

// TableName returns the table name for MonitoredContainer.
func (MonitoredContainer) TableName() string {
	return "monitored_containers"
}

// TableName returns the table name for ContainerHealthCheck.
func (ContainerHealthCheck) TableName() string {
	return "container_health_checks"
}

// TableName returns the table name for ComponentMetric.
func (ComponentMetric) TableName() string {
	return "component_metrics"
}

// Validate performs validation on MonitoredComponent.
func (c *MonitoredComponent) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("component name is required")
	}
	if c.Type == "" {
		return fmt.Errorf("component type is required")
	}
	if c.TenantID == 0 {
		return fmt.Errorf("tenant ID is required")
	}
	return nil
}

// Validate performs validation on MonitoredContainer.
func (c *MonitoredContainer) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("container name is required")
	}
	if c.Type == "" {
		return fmt.Errorf("container type is required")
	}
	if c.ComponentID == 0 {
		return fmt.Errorf("component ID is required")
	}
	return nil
}

// Validate performs validation on ContainerHealthCheck.
func (c *ContainerHealthCheck) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("health check name is required")
	}
	if c.Type == "" {
		return fmt.Errorf("health check type is required")
	}
	if c.ContainerID == 0 {
		return fmt.Errorf("container ID is required")
	}
	return nil
}

// Validate performs validation on ComponentMetric.
func (c *ComponentMetric) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("metric name is required")
	}
	if c.ComponentID == 0 {
		return fmt.Errorf("component ID is required")
	}
	if c.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	return nil
}
