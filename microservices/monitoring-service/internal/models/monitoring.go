// Package models provides data models for the Monitoring Service.
package models

import (
	"time"

	"gorm.io/gorm"
)

// MonitoredService represents a service being monitored.
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

	// Related entities
	HealthChecks []HealthCheck `gorm:"foreignKey:ServiceID" json:"health_checks,omitempty"`
	Alerts       []Alert       `gorm:"foreignKey:ServiceID" json:"alerts,omitempty"`
}

// HealthCheck represents a health check for a service.
type HealthCheck struct {
	ID             uint             `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	DeletedAt      gorm.DeletedAt   `gorm:"index" json:"deleted_at,omitempty"`
	ServiceID      uint             `gorm:"not null;index:idx_service_active" json:"service_id"`
	Service        MonitoredService `gorm:"foreignKey:ServiceID" json:"service"`
	Name           string           `gorm:"not null" json:"name"`
	Type           string           `gorm:"not null" json:"type"` // http, tcp, ping, custom
	URL            string           `json:"url"`
	Method         string           `gorm:"default:GET" json:"method"`
	Headers        string           `gorm:"type:text" json:"headers"` // JSON string
	Body           string           `gorm:"type:text" json:"body"`
	ExpectedStatus int              `gorm:"default:200" json:"expected_status"`
	ExpectedBody   string           `json:"expected_body"`
	Timeout        int              `gorm:"default:30" json:"timeout"`
	IsActive       bool             `gorm:"default:true;index:idx_service_active" json:"is_active"`
	LastChecked    *time.Time       `gorm:"index:idx_last_checked" json:"last_checked"`
	LastResult     string           `gorm:"default:unknown;index:idx_result_checked" json:"last_result"` // success, failure, timeout
	Metadata       string           `gorm:"type:text" json:"metadata"`          // JSON string for additional data
}

// HealthCheckResult represents the result of a health check.
type HealthCheckResult struct {
	ID            uint             `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	DeletedAt     gorm.DeletedAt   `gorm:"index" json:"deleted_at,omitempty"`
	ServiceID     uint             `gorm:"not null;index" json:"service_id"`
	Service       MonitoredService `gorm:"foreignKey:ServiceID" json:"service"`
	HealthCheckID uint             `gorm:"not null;index" json:"health_check_id"`
	HealthCheck   HealthCheck      `gorm:"foreignKey:HealthCheckID" json:"health_check"`
	Status        string           `gorm:"not null" json:"status"` // success, failure, timeout
	ResponseTime  float64          `json:"response_time"`          // in milliseconds
	StatusCode    int              `json:"status_code"`
	ResponseBody  string           `gorm:"type:text" json:"response_body"`
	ErrorMessage  string           `json:"error_message"`
	Metadata      string           `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

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
	ResolvedAt     *time.Time        `json:"resolved_at"`
	ResolvedBy     *uint             `json:"resolved_by"`
	Metadata       string            `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// UptimeCheck represents an uptime check for a service.
type UptimeCheck struct {
	ID                  uint           `gorm:"primarykey" json:"id"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID            uint           `gorm:"not null;index" json:"tenant_id"`
	Name                string         `gorm:"not null" json:"name"`
	Description         string         `json:"description"`
	URL                 string         `gorm:"not null" json:"url"`
	Method              string         `gorm:"default:GET" json:"method"`
	Headers             string         `gorm:"type:text" json:"headers"` // JSON string
	Body                string         `gorm:"type:text" json:"body"`
	ExpectedStatus      int            `gorm:"default:200" json:"expected_status"`
	ExpectedBody        string         `json:"expected_body"`
	CheckInterval       int            `gorm:"default:60" json:"check_interval"` // in seconds
	Timeout             int            `gorm:"default:30" json:"timeout"`        // in seconds
	Retries             int            `gorm:"default:3" json:"retries"`
	IsActive            bool           `gorm:"default:true" json:"is_active"`
	LastChecked         *time.Time     `json:"last_checked"`
	LastSuccess         *time.Time     `json:"last_success"`
	LastFailure         *time.Time     `json:"last_failure"`
	Uptime              float64        `gorm:"default:0" json:"uptime"`   // percentage
	AverageResponseTime float64        `json:"average_response_time"`     // in milliseconds
	Metadata            string         `gorm:"type:text" json:"metadata"` // JSON string for additional data

	// Related entities
	Results []UptimeResult `gorm:"foreignKey:UptimeCheckID" json:"results,omitempty"`
}

// UptimeResult represents the result of an uptime check.
type UptimeResult struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UptimeCheckID uint           `gorm:"not null;index" json:"uptime_check_id"`
	UptimeCheck   UptimeCheck    `gorm:"foreignKey:UptimeCheckID" json:"uptime_check"`
	Status        string         `gorm:"not null" json:"status"` // success, failure, timeout
	ResponseTime  float64        `json:"response_time"`          // in milliseconds
	StatusCode    int            `json:"status_code"`
	ResponseBody  string         `gorm:"type:text" json:"response_body"`
	ErrorMessage  string         `json:"error_message"`
	Metadata      string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// PerformanceMetric represents a performance metric.
type PerformanceMetric struct {
	ID            uint              `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	DeletedAt     gorm.DeletedAt    `gorm:"index" json:"deleted_at,omitempty"`
	TenantID      uint              `gorm:"not null;index" json:"tenant_id"`
	ServiceID     *uint             `gorm:"index" json:"service_id"`
	Service       *MonitoredService `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Name          string            `gorm:"not null" json:"name"`
	Description   string            `json:"description"`
	Type          string            `gorm:"not null" json:"type"` // response_time, cpu_usage, memory_usage, disk_usage, custom
	Unit          string            `json:"unit"`                 // milliseconds, percentage, bytes, etc.
	Category      string            `json:"category"`             // system, application, business
	IsActive      bool              `gorm:"default:true" json:"is_active"`
	RetentionDays int               `gorm:"default:30" json:"retention_days"`
	Metadata      string            `gorm:"type:text" json:"metadata"` // JSON string for additional data

	// Related entities
	DataPoints []PerformanceDataPoint `gorm:"foreignKey:MetricID" json:"data_points,omitempty"`
}

// PerformanceDataPoint represents a performance metric data point.
type PerformanceDataPoint struct {
	ID        uint              `gorm:"primarykey" json:"id"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	DeletedAt gorm.DeletedAt    `gorm:"index" json:"deleted_at,omitempty"`
	MetricID  uint              `gorm:"not null;index" json:"metric_id"`
	Metric    PerformanceMetric `gorm:"foreignKey:MetricID" json:"metric"`
	Value     float64           `gorm:"not null" json:"value"`
	Timestamp time.Time         `gorm:"not null;index" json:"timestamp"`
	Labels    string            `gorm:"type:text" json:"labels"`   // JSON string for metric labels
	Source    string            `json:"source"`                    // agent, api, webhook, etc.
	Metadata  string            `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// LogEntry represents a log entry in the system.
type LogEntry struct {
	ID        uint              `gorm:"primarykey" json:"id"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	DeletedAt gorm.DeletedAt    `gorm:"index" json:"deleted_at,omitempty"`
	TenantID  uint              `gorm:"not null;index" json:"tenant_id"`
	ServiceID *uint             `gorm:"index" json:"service_id"`
	Service   *MonitoredService `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Level     string            `gorm:"not null;index" json:"level"` // debug, info, warn, error, fatal
	Message   string            `gorm:"type:text;not null" json:"message"`
	Source    string            `json:"source"` // application, system, access, etc.
	Timestamp time.Time         `gorm:"not null;index" json:"timestamp"`
	Fields    string            `gorm:"type:text" json:"fields"`   // JSON string for structured fields
	Metadata  string            `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for MonitoredService.
func (MonitoredService) TableName() string {
	return "monitored_services"
}

// TableName returns the table name for HealthCheck.
func (HealthCheck) TableName() string {
	return "health_checks"
}

// TableName returns the table name for HealthCheckResult.
func (HealthCheckResult) TableName() string {
	return "health_check_results"
}

// TableName returns the table name for Alert.
func (Alert) TableName() string {
	return "alerts"
}

// TableName returns the table name for UptimeCheck.
func (UptimeCheck) TableName() string {
	return "uptime_checks"
}

// TableName returns the table name for UptimeResult.
func (UptimeResult) TableName() string {
	return "uptime_results"
}

// TableName returns the table name for PerformanceMetric.
func (PerformanceMetric) TableName() string {
	return "performance_metrics"
}

// TableName returns the table name for PerformanceDataPoint.
func (PerformanceDataPoint) TableName() string {
	return "performance_data_points"
}

// TableName returns the table name for LogEntry.
func (LogEntry) TableName() string {
	return "log_entries"
}
