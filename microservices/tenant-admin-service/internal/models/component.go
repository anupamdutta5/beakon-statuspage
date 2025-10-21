// Package models provides data models for component management in Tenant Admin Service.
package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Component represents a service component monitored by the status page.
type Component struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Status      string         `gorm:"default:operational" json:"status"` // operational, degraded_performance, partial_outage, major_outage
	GroupName   string         `gorm:"column:group_name" json:"group_name,omitempty"`  // Group name from migration
	Order       int64          `gorm:"column:order;default:0" json:"order"` // Display order (using int64 to match bigint)
	Visible     bool           `gorm:"column:visible;default:true" json:"visible"`      // Visible from migration
	ShowUptime  bool           `gorm:"column:show_uptime;default:true" json:"show_uptime"` // Show uptime badge
	Link        string         `gorm:"column:link" json:"link,omitempty"`              // Optional external link
}

// ComponentGroup represents a logical group of components.
type ComponentGroup struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name       string         `gorm:"not null" json:"name"`
	Position   int            `gorm:"default:0" json:"position"`
	IsVisible  bool           `gorm:"default:true" json:"is_visible"`
	Components []Component    `gorm:"foreignKey:GroupID" json:"components,omitempty"`
}

// TableName returns the table name for Component.
func (Component) TableName() string {
	return "saas_components"
}

// TableName returns the table name for ComponentGroup.
func (ComponentGroup) TableName() string {
	return "component_groups"
}

// Validate performs validation on Component.
func (c *Component) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("component name is required")
	}
	if c.TenantID == uuid.Nil {
		return fmt.Errorf("tenant ID is required")
	}

	// Validate status
	validStatuses := map[string]bool{
		"operational":          true,
		"degraded_performance": true,
		"partial_outage":       true,
		"major_outage":         true,
		"maintenance":          true,
	}
	if c.Status != "" && !validStatuses[c.Status] {
		return fmt.Errorf("invalid component status: %s", c.Status)
	}

	return nil
}

// Validate performs validation on ComponentGroup.
func (cg *ComponentGroup) Validate() error {
	if cg.Name == "" {
		return fmt.Errorf("component group name is required")
	}
	if cg.TenantID == uuid.Nil {
		return fmt.Errorf("tenant ID is required")
	}
	return nil
}

// GetStatusText returns human-readable status text.
func (c *Component) GetStatusText() string {
	statusMap := map[string]string{
		"operational":          "Operational",
		"degraded_performance": "Degraded Performance",
		"partial_outage":       "Partial Outage",
		"major_outage":         "Major Outage",
		"maintenance":          "Under Maintenance",
	}

	if text, exists := statusMap[c.Status]; exists {
		return text
	}
	return c.Status
}
