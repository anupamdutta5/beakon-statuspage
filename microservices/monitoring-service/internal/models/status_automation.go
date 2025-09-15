// Package models provides status automation integration data models.
package models

import (
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage-monitoring-service/internal/utils"
	"gorm.io/gorm"
)

// StatusAutomation represents a status automation integration.
type StatusAutomation struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	Name        string         `gorm:"not null;size:255" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Type        string         `gorm:"not null;size:50" json:"type"` // nagios, opsgenie, pagerduty, pingdom, uptimerobot, site24x7, statuscake
	Provider    string         `gorm:"size:100" json:"provider"`     // Nagios, OpsGenie, PagerDuty, etc.
	Config      string         `gorm:"type:text" json:"config"`      // JSON string for provider-specific configuration
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	LastSync    *time.Time     `json:"last_sync"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data

	// Related entities
	Integrations []StatusAutomationIntegration `gorm:"foreignKey:AutomationID;constraint:OnDelete:CASCADE" json:"integrations,omitempty"`
}

// StatusAutomationIntegration represents an integration with a status automation provider.
type StatusAutomationIntegration struct {
	ID           uint              `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	DeletedAt    gorm.DeletedAt    `gorm:"index" json:"deleted_at,omitempty"`
	AutomationID uint              `gorm:"not null;index" json:"automation_id"`
	Automation   StatusAutomation  `gorm:"foreignKey:AutomationID" json:"automation"`
	ServiceID    *uint             `gorm:"index" json:"service_id"`
	Service      *MonitoredService `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	ExternalID   string            `gorm:"size:255" json:"external_id"` // ID in the external system
	Config       string            `gorm:"type:text" json:"config"`     // JSON string for integration-specific configuration
	IsActive     bool              `gorm:"default:true" json:"is_active"`
	LastSync     *time.Time        `json:"last_sync"`
	Metadata     string            `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for StatusAutomation.
func (StatusAutomation) TableName() string {
	return "status_automations"
}

// TableName returns the table name for StatusAutomationIntegration.
func (StatusAutomationIntegration) TableName() string {
	return "status_automation_integrations"
}

// Validate performs validation on StatusAutomation.
func (s *StatusAutomation) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("automation name is required")
	}
	if s.Type == "" {
		return fmt.Errorf("automation type is required")
	}
	if s.TenantID == 0 {
		return fmt.Errorf("tenant ID is required")
	}

	// Validate automation type
	validTypes := []string{
		"nagios", "opsgenie", "pagerduty", "pingdom",
		"uptimerobot", "site24x7", "statuscake", "datadog",
		"dynatrace", "logicmonitor", "zabbix",
	}
	if !utils.ContainsString(validTypes, s.Type) {
		return fmt.Errorf("invalid automation type: %s", s.Type)
	}

	return nil
}

// Validate performs validation on StatusAutomationIntegration.
func (s *StatusAutomationIntegration) Validate() error {
	if s.AutomationID == 0 {
		return fmt.Errorf("automation ID is required")
	}
	return nil
}
