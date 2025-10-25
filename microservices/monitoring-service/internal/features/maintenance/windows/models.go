// Package models provides maintenance management data models for monitoring service.
package windows

import (
	"fmt"
	"time"

	"github.com/anupamdutta5/monitoring-service/internal/core/validation"
	"gorm.io/gorm"
)

// MaintenanceWindow represents a scheduled maintenance window.
type MaintenanceWindow struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// Note: No DeletedAt field - table doesn't support soft deletes
	TenantID    string    `gorm:"type:uuid;not null;index;column:tenant_id" json:"tenant_id"`
	Name        string    `gorm:"not null;size:255;column:name" json:"name"`
	Description string    `gorm:"type:text;column:description" json:"description"`
	StartsAt    time.Time `gorm:"not null;column:starts_at" json:"starts_at"`
	EndsAt      time.Time `gorm:"not null;column:ends_at" json:"ends_at"`
	Status      string    `gorm:"default:scheduled;size:50;column:status" json:"status"` // scheduled, in_progress, completed, cancelled
	IsActive    bool      `gorm:"default:true;column:is_active" json:"is_active"`
	CreatedBy   string    `gorm:"type:uuid;column:created_by" json:"created_by"`

	// Automation fields
	ReminderSent     bool       `gorm:"default:false;column:reminder_sent" json:"reminder_sent"`
	AutoStarted      bool       `gorm:"default:false;column:auto_started" json:"auto_started"`
	AutoCompleted    bool       `gorm:"default:false;column:auto_completed" json:"auto_completed"`
	ActualStartTime  *time.Time `gorm:"column:actual_start_time" json:"actual_start_time,omitempty"`
	ActualEndTime    *time.Time `gorm:"column:actual_end_time" json:"actual_end_time,omitempty"`

	// Related entities
	Components []MaintenanceComponent `gorm:"foreignKey:MaintenanceID;constraint:OnDelete:CASCADE" json:"components,omitempty"`
	Updates    []MaintenanceUpdate    `gorm:"foreignKey:MaintenanceID;constraint:OnDelete:CASCADE" json:"updates,omitempty"`
}

// MaintenanceComponent represents a component affected by maintenance.
type MaintenanceComponent struct {
	ID            uint               `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	DeletedAt     gorm.DeletedAt     `gorm:"index" json:"deleted_at,omitempty"`
	MaintenanceID uint               `gorm:"not null;index" json:"maintenance_id"`
	Maintenance   MaintenanceWindow  `gorm:"foreignKey:MaintenanceID" json:"maintenance"`
	ComponentID   uint               `gorm:"not null;index" json:"component_id"`
	Component     MonitoredComponent `gorm:"foreignKey:ComponentID" json:"component"`
	Status        string             `gorm:"default:operational;size:50" json:"status"` // operational, degraded, partial_outage, major_outage, maintenance
	Metadata      string             `gorm:"type:text" json:"metadata"`                 // JSON string for additional data
}

// MaintenanceUpdate represents an update to a maintenance window.
type MaintenanceUpdate struct {
	ID            uint              `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	DeletedAt     gorm.DeletedAt    `gorm:"index" json:"deleted_at,omitempty"`
	MaintenanceID uint              `gorm:"not null;index" json:"maintenance_id"`
	Maintenance   MaintenanceWindow `gorm:"foreignKey:MaintenanceID" json:"maintenance"`
	Status        string            `gorm:"not null;size:50" json:"status"` // scheduled, in_progress, completed, cancelled
	Message       string            `gorm:"type:text" json:"message"`
	IsPublic      bool              `gorm:"default:true" json:"is_public"`
	CreatedBy     uint              `gorm:"not null" json:"created_by"`
	Metadata      string            `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// MaintenanceTemplate represents a template for creating maintenance windows.
type MaintenanceTemplate struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	Name        string         `gorm:"not null;size:255" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Title       string         `gorm:"not null;size:255" json:"title"`
	Message     string         `gorm:"type:text" json:"message"`
	Type        string         `gorm:"not null;size:50" json:"type"`
	Impact      string         `gorm:"not null;size:50" json:"impact"`
	Duration    int            `gorm:"default:60" json:"duration"` // in minutes
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for MaintenanceWindow.
func (MaintenanceWindow) TableName() string {
	return "maintenance_windows"
}

// TableName returns the table name for MaintenanceComponent.
func (MaintenanceComponent) TableName() string {
	return "maintenance_components"
}

// TableName returns the table name for MaintenanceUpdate.
func (MaintenanceUpdate) TableName() string {
	return "maintenance_updates"
}

// TableName returns the table name for MaintenanceTemplate.
func (MaintenanceTemplate) TableName() string {
	return "maintenance_templates"
}

// Validate performs validation on MaintenanceWindow.
func (m *MaintenanceWindow) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("maintenance name is required")
	}
	if m.TenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if m.StartsAt.IsZero() {
		return fmt.Errorf("start time is required")
	}
	if m.EndsAt.IsZero() {
		return fmt.Errorf("end time is required")
	}
	if m.EndsAt.Before(m.StartsAt) {
		return fmt.Errorf("end time must be after start time")
	}

	// Validate status
	validStatuses := []string{"scheduled", "in_progress", "completed", "cancelled"}
	if m.Status != "" && !utils.ContainsString(validStatuses, m.Status) {
		return fmt.Errorf("invalid maintenance status: %s", m.Status)
	}

	return nil
}

// Validate performs validation on MaintenanceComponent.
func (m *MaintenanceComponent) Validate() error {
	if m.MaintenanceID == 0 {
		return fmt.Errorf("maintenance ID is required")
	}
	if m.ComponentID == 0 {
		return fmt.Errorf("component ID is required")
	}

	// Validate status
	validStatuses := []string{"operational", "degraded", "partial_outage", "major_outage", "maintenance"}
	if !utils.ContainsString(validStatuses, m.Status) {
		return fmt.Errorf("invalid component status: %s", m.Status)
	}

	return nil
}

// Validate performs validation on MaintenanceUpdate.
func (m *MaintenanceUpdate) Validate() error {
	if m.MaintenanceID == 0 {
		return fmt.Errorf("maintenance ID is required")
	}
	if m.Message == "" {
		return fmt.Errorf("update message is required")
	}
	if m.CreatedBy == 0 {
		return fmt.Errorf("created by user ID is required")
	}

	// Validate status
	validStatuses := []string{"scheduled", "in_progress", "completed", "cancelled"}
	if !utils.ContainsString(validStatuses, m.Status) {
		return fmt.Errorf("invalid update status: %s", m.Status)
	}

	return nil
}

// Validate performs validation on MaintenanceTemplate.
func (m *MaintenanceTemplate) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if m.Title == "" {
		return fmt.Errorf("template title is required")
	}
	if m.TenantID == 0 {
		return fmt.Errorf("tenant ID is required")
	}
	if m.Duration <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}

	// Validate type
	validTypes := []string{"planned", "emergency", "routine"}
	if !utils.ContainsString(validTypes, m.Type) {
		return fmt.Errorf("invalid template type: %s", m.Type)
	}

	// Validate impact
	validImpacts := []string{"none", "minor", "major", "critical"}
	if !utils.ContainsString(validImpacts, m.Impact) {
		return fmt.Errorf("invalid template impact: %s", m.Impact)
	}

	return nil
}
