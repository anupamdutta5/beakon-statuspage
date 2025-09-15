// Package models provides external service monitoring data models.
package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ExternalService represents an external service being monitored.
type ExternalService struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID      uint           `gorm:"not null;index" json:"tenant_id"`
	Name          string         `gorm:"not null;size:255" json:"name"`
	Description   string         `gorm:"type:text" json:"description"`
	Type          string         `gorm:"not null;size:50" json:"type"` // aws, stripe, mailgun, third_party_api, webhook
	Provider      string         `gorm:"size:100" json:"provider"`     // AWS, Stripe, Mailgun, etc.
	URL           string         `gorm:"size:500" json:"url"`
	APIKey        string         `gorm:"size:500" json:"api_key"`               // encrypted
	Status        string         `gorm:"default:unknown;size:50" json:"status"` // operational, degraded, outage, maintenance
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	CheckInterval int            `gorm:"default:300" json:"check_interval"` // in seconds
	LastChecked   *time.Time     `json:"last_checked"`
	LastHealthy   *time.Time     `json:"last_healthy"`
	LastUnhealthy *time.Time     `json:"last_unhealthy"`
	Uptime        float64        `gorm:"default:0" json:"uptime"`   // percentage
	ResponseTime  float64        `json:"response_time"`             // in milliseconds
	Metadata      string         `gorm:"type:text" json:"metadata"` // JSON string for additional data

	// Related entities
	HealthChecks []ExternalServiceHealthCheck `gorm:"foreignKey:ExternalServiceID;constraint:OnDelete:CASCADE" json:"health_checks,omitempty"`
}

// ExternalServiceHealthCheck represents a health check for an external service.
type ExternalServiceHealthCheck struct {
	ID                uint            `gorm:"primarykey" json:"id"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	DeletedAt         gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`
	ExternalServiceID uint            `gorm:"not null;index" json:"external_service_id"`
	ExternalService   ExternalService `gorm:"foreignKey:ExternalServiceID" json:"external_service"`
	Name              string          `gorm:"not null;size:255" json:"name"`
	Type              string          `gorm:"not null;size:50" json:"type"` // api_check, webhook_check, status_page_check, email_trigger
	URL               string          `gorm:"size:500" json:"url"`
	Method            string          `gorm:"default:GET;size:10" json:"method"`
	Headers           string          `gorm:"type:text" json:"headers"` // JSON string
	Body              string          `gorm:"type:text" json:"body"`
	ExpectedStatus    int             `gorm:"default:200" json:"expected_status"`
	ExpectedBody      string          `gorm:"type:text" json:"expected_body"`
	Timeout           int             `gorm:"default:30" json:"timeout"`
	IsActive          bool            `gorm:"default:true" json:"is_active"`
	LastChecked       *time.Time      `json:"last_checked"`
	LastResult        string          `gorm:"default:unknown;size:50" json:"last_result"` // success, failure, timeout
	Metadata          string          `gorm:"type:text" json:"metadata"`                  // JSON string for additional data
}

// TableName returns the table name for ExternalService.
func (ExternalService) TableName() string {
	return "external_services"
}

// TableName returns the table name for ExternalServiceHealthCheck.
func (ExternalServiceHealthCheck) TableName() string {
	return "external_service_health_checks"
}

// Validate performs validation on ExternalService.
func (e *ExternalService) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("external service name is required")
	}
	if e.Type == "" {
		return fmt.Errorf("external service type is required")
	}
	if e.TenantID == 0 {
		return fmt.Errorf("tenant ID is required")
	}
	if e.CheckInterval <= 0 {
		return fmt.Errorf("check interval must be greater than 0")
	}
	return nil
}

// Validate performs validation on ExternalServiceHealthCheck.
func (e *ExternalServiceHealthCheck) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("health check name is required")
	}
	if e.Type == "" {
		return fmt.Errorf("health check type is required")
	}
	if e.ExternalServiceID == 0 {
		return fmt.Errorf("external service ID is required")
	}
	if e.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}
	return nil
}
