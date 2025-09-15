// Package models provides data models for status page management in Tenant Admin Service.
package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// StatusPage represents a status page configuration managed by tenant admin.
type StatusPage struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	Name        string         `gorm:"not null" json:"name"`
	Slug        string         `gorm:"not null;uniqueIndex" json:"slug"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Domain      string         `json:"domain"`
	IsPublic    bool           `gorm:"default:true" json:"is_public"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// StatusPageConfig represents status page configuration settings.
type StatusPageConfig struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	StatusPageID  uint           `gorm:"not null;index" json:"status_page_id"`
	StatusPage    StatusPage     `gorm:"foreignKey:StatusPageID" json:"status_page"`
	SiteName      string         `gorm:"not null" json:"site_name"`
	SiteDescription string       `json:"site_description"`
	LogoURL       string         `json:"logo_url"`
	Favicon       string         `json:"favicon"`
	OGImage       string         `json:"og_image"`
	SiteURL       string         `json:"site_url"`
	SubscribeURL  string         `json:"subscribe_url"`
	HistoryURL    string         `json:"history_url"`
	APIURL        string         `json:"api_url"`
	WebSocketURL  string         `json:"websocket_url"`
	PoweredBy     string         `json:"powered_by"`
	PoweredByURL  string         `json:"powered_by_url"`
	Theme         string         `gorm:"default:default" json:"theme"`
	CustomCSS     string         `gorm:"type:text" json:"custom_css"`
	Metadata      string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// StatusPageData represents the data structure for rendering the status page.
type StatusPageData struct {
	// Site Configuration
	SiteName        string `json:"site_name"`
	SiteDescription string `json:"site_description"`
	LogoURL         string `json:"logo_url"`
	Favicon         string `json:"favicon"`
	OGImage         string `json:"og_image"`
	SiteURL         string `json:"site_url"`
	SubscribeURL    string `json:"subscribe_url"`
	HistoryURL      string `json:"history_url"`
	APIURL          string `json:"api_url"`
	WebSocketURL    string `json:"websocket_url"`
	PoweredBy       string `json:"powered_by"`
	PoweredByURL    string `json:"powered_by_url"`
	LastUpdated     string `json:"last_updated"`

	// Overall Status
	OverallStatus     string `json:"overall_status"`
	StatusTitle       string `json:"status_title"`
	StatusDescription string `json:"status_description"`

	// Components
	Components []ComponentStatus `json:"components"`

	// Incidents
	RecentIncidents []IncidentStatus `json:"recent_incidents"`

	// Maintenance
	ScheduledMaintenance []MaintenanceStatus `json:"scheduled_maintenance"`
}

// ComponentStatus represents component status for the status page.
type ComponentStatus struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	StatusText  string  `json:"status_text"`
	Uptime      float64 `json:"uptime"`
}

// IncidentStatus represents incident status for the status page.
type IncidentStatus struct {
	ID          uint              `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
	StatusText  string            `json:"status_text"`
	CreatedAt   string            `json:"created_at"`
	Updates     []IncidentUpdate  `json:"updates"`
}

// IncidentUpdate represents an incident update for the status page.
type IncidentUpdate struct {
	ID        uint   `json:"id"`
	Message   string `json:"message"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// MaintenanceStatus represents maintenance status for the status page.
type MaintenanceStatus struct {
	ID            uint   `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Status        string `json:"status"`
	Impact        string `json:"impact"`
	ImpactText    string `json:"impact_text"`
	ScheduledTime string `json:"scheduled_time"`
}

// TableName returns the table name for StatusPage.
func (StatusPage) TableName() string {
	return "status_pages"
}

// TableName returns the table name for StatusPageConfig.
func (StatusPageConfig) TableName() string {
	return "status_page_configs"
}

// Validate performs validation on StatusPage.
func (s *StatusPage) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("status page name is required")
	}
	if s.Slug == "" {
		return fmt.Errorf("status page slug is required")
	}
	if s.Title == "" {
		return fmt.Errorf("status page title is required")
	}
	if s.TenantID == 0 {
		return fmt.Errorf("tenant ID is required")
	}
	return nil
}

// Validate performs validation on StatusPageConfig.
func (c *StatusPageConfig) Validate() error {
	if c.SiteName == "" {
		return fmt.Errorf("site name is required")
	}
	if c.StatusPageID == 0 {
		return fmt.Errorf("status page ID is required")
	}
	return nil
}
