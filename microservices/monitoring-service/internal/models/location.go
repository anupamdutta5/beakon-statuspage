// Package models provides data models for the Monitoring Service.
package models

import (
	"time"

	"gorm.io/gorm"
)

// MonitoringLocation represents a global monitoring node location
// for performing health checks from multiple geographic regions.
type MonitoringLocation struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Name      string  `gorm:"size:100;not null" json:"name"`               // e.g., "US East (N. Virginia)"
	City      string  `gorm:"size:100" json:"city"`                        // e.g., "Ashburn"
	Country   string  `gorm:"size:100;not null" json:"country"`            // e.g., "United States"
	Region    string  `gorm:"size:50;not null;index" json:"region"`        // e.g., "us-east-1"
	Latitude  float64 `gorm:"type:decimal(9,6)" json:"latitude"`           // Geographic coordinates
	Longitude float64 `gorm:"type:decimal(9,6)" json:"longitude"`          // Geographic coordinates
	IsActive  bool    `gorm:"default:true;index" json:"is_active"`         // Can be used for monitoring
	Provider  string  `gorm:"size:50" json:"provider,omitempty"`           // e.g., "aws", "gcp", "azure"
}

// TableName specifies the table name for GORM.
func (MonitoringLocation) TableName() string {
	return "monitoring_locations"
}

// MonitoringResult represents a single health check result from a specific location.
// This table is partitioned by checked_at for efficient time-series queries.
type MonitoringResult struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	MonitorID  uint      `gorm:"not null;index:idx_monitoring_results_monitor_time" json:"monitor_id"`
	LocationID *uint     `gorm:"index:idx_monitoring_results_location" json:"location_id,omitempty"`
	CheckedAt  time.Time `gorm:"not null;index:idx_monitoring_results_monitor_time" json:"checked_at"`

	Status          string  `gorm:"size:20;not null;index:idx_monitoring_results_status" json:"status"` // operational, degraded, down
	ResponseTimeMs  *int    `json:"response_time_ms,omitempty"`                                         // Total response time
	TTFBMs          *int    `json:"ttfb_ms,omitempty"`                                                  // Time to first byte
	DNSTimeMs       *int    `json:"dns_time_ms,omitempty"`                                              // DNS resolution time
	ConnectionTimeMs *int   `json:"connection_time_ms,omitempty"`                                       // TCP connection time
	SSLHandshakeMs  *int    `json:"ssl_handshake_ms,omitempty"`                                         // SSL/TLS handshake time
	StatusCode      *int    `json:"status_code,omitempty"`                                              // HTTP status code
	ErrorMessage    string  `gorm:"type:text" json:"error_message,omitempty"`                           // Error details if check failed

	CreatedAt time.Time `json:"created_at"`

	// Relationships
	Location *MonitoringLocation `gorm:"foreignKey:LocationID" json:"location,omitempty"`
}

// TableName specifies the table name for GORM.
func (MonitoringResult) TableName() string {
	return "monitoring_results"
}

// MonitoringResultHourly represents hourly aggregated monitoring metrics.
// Used for historical analysis (warm data: 8-90 days).
type MonitoringResultHourly struct {
	MonitorID  uint      `gorm:"primarykey" json:"monitor_id"`
	LocationID uint      `gorm:"primarykey" json:"location_id"`
	Hour       time.Time `gorm:"primarykey;index:idx_monitoring_hourly_time" json:"hour"`

	AvgResponseTimeMs int     `json:"avg_response_time_ms"`
	MinResponseTimeMs int     `json:"min_response_time_ms"`
	MaxResponseTimeMs int     `json:"max_response_time_ms"`
	P50ResponseTimeMs int     `json:"p50_response_time_ms"` // Median
	P95ResponseTimeMs int     `json:"p95_response_time_ms"` // 95th percentile
	P99ResponseTimeMs int     `json:"p99_response_time_ms"` // 99th percentile
	UptimePercentage  float64 `gorm:"type:decimal(5,2)" json:"uptime_percentage"`
	CheckCount        int     `json:"check_count"`
	FailureCount      int     `json:"failure_count"`

	CreatedAt time.Time `json:"created_at"`

	// Relationships
	Location *MonitoringLocation `gorm:"foreignKey:LocationID" json:"location,omitempty"`
}

// TableName specifies the table name for GORM.
func (MonitoringResultHourly) TableName() string {
	return "monitoring_results_hourly"
}

// MonitoringResultDaily represents daily aggregated monitoring metrics.
// Used for long-term storage (cold data: 90+ days).
type MonitoringResultDaily struct {
	MonitorID uint      `gorm:"primarykey" json:"monitor_id"`
	Day       time.Time `gorm:"primarykey;index:idx_monitoring_daily_day;type:date" json:"day"`

	AvgResponseTimeMs int     `json:"avg_response_time_ms"`
	P95ResponseTimeMs int     `json:"p95_response_time_ms"`
	P99ResponseTimeMs int     `json:"p99_response_time_ms"`
	UptimePercentage  float64 `gorm:"type:decimal(5,2)" json:"uptime_percentage"`
	TotalChecks       int     `json:"total_checks"`
	TotalFailures     int     `json:"total_failures"`

	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for GORM.
func (MonitoringResultDaily) TableName() string {
	return "monitoring_results_daily"
}
