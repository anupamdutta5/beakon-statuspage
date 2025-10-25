// Package core provides alert core functionality for the Monitoring Service.
package core

import (
	"time"

	"gorm.io/gorm"
)

// Alert represents an alert in the system.
type Alert struct {
	ID             uint              `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	DeletedAt      gorm.DeletedAt    `gorm:"index" json:"deleted_at,omitempty"`
	TenantID       uint              `gorm:"not null;index" json:"tenant_id"`
	ServiceID      *uint             `gorm:"index" json:"service_id"`
	Service        *MonitoredService `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Type           string            `gorm:"not null" json:"type"`         // service_down, high_response_time, custom
	Severity       string            `gorm:"not null" json:"severity"`     // critical, warning, info
	Status         string            `gorm:"default:active" json:"status"` // active, acknowledged, resolved
	Title          string            `gorm:"not null" json:"title"`
	Description    string            `json:"description"`
	Message        string            `gorm:"type:text" json:"message"`
	TriggeredAt    time.Time         `gorm:"not null" json:"triggered_at"`
	AcknowledgedAt *time.Time        `json:"acknowledged_at"`
	AcknowledgedBy *uint             `json:"acknowledged_by"`
	ResolvedAt      *time.Time        `json:"resolved_at"`
	ResolvedBy      *uint             `json:"resolved_by"`
	ResolutionType  *string           `json:"resolution_type"`  // auto, manual
	ResolutionNote  *string           `gorm:"type:text" json:"resolution_note"`
	DedupKey        *string           `json:"dedup_key"`         // Deduplication key
	ErrorType       *string           `json:"error_type"`        // Type of error (timeout, 5xx, etc)
	Metadata        string            `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// MonitoredService represents a service being monitored (reference for FK).
type MonitoredService struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID      uint           `gorm:"not null;index:idx_tenant_status" json:"tenant_id"`
	Name          string         `gorm:"not null" json:"name"`
	Description   string         `json:"description"`
	Type          string         `gorm:"not null;index:idx_type_status" json:"type"` // http, tcp, ping, custom
	URL           string         `json:"url"`
	Host          string         `json:"host"`
	Port          int            `json:"port"`
	Status        string         `gorm:"default:unknown;index:idx_tenant_status;index:idx_type_status" json:"status"` // healthy, unhealthy, unknown, maintenance
	IsActive      bool           `gorm:"default:true;index:idx_active_status" json:"is_active"`
	CheckInterval int            `gorm:"default:60" json:"check_interval"` // in seconds
	Timeout       int            `gorm:"default:30" json:"timeout"`        // in seconds
	Retries       int            `gorm:"default:3" json:"retries"`
	LastChecked   *time.Time     `json:"last_checked"`
	LastHealthy   *time.Time     `json:"last_healthy"`
	LastUnhealthy *time.Time     `json:"last_unhealthy"`
	Uptime        float64        `gorm:"default:0" json:"uptime"`   // percentage
	ResponseTime  float64        `json:"response_time"`             // in milliseconds
	Metadata      string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}
