// Package models provides data models for the Analytics Service.
package models

import (
	"time"

	"gorm.io/gorm"
)

// Metric represents a metric definition in the system.
type Metric struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID        uint           `gorm:"not null;index" json:"tenant_id"`
	Name            string         `gorm:"not null" json:"name"`
	Description     string         `json:"description"`
	Type            string         `gorm:"not null" json:"type"` // counter, gauge, histogram, summary
	Unit            string         `json:"unit"`                 // seconds, bytes, requests, etc.
	Category        string         `json:"category"`             // performance, business, system, user
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	IsPublic        bool           `gorm:"default:false" json:"is_public"`
	AggregationType string         `gorm:"default:sum" json:"aggregation_type"` // sum, avg, min, max, count
	RetentionDays   int            `gorm:"default:365" json:"retention_days"`
	Metadata        string         `gorm:"type:text" json:"metadata"` // JSON string for additional data

	// Related entities
	DataPoints []MetricDataPoint `gorm:"foreignKey:MetricID" json:"data_points,omitempty"`
}

// MetricDataPoint represents a single data point for a metric.
type MetricDataPoint struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	MetricID  uint           `gorm:"not null;index" json:"metric_id"`
	Metric    Metric         `gorm:"foreignKey:MetricID" json:"metric"`
	Value     float64        `gorm:"not null" json:"value"`
	Timestamp time.Time      `gorm:"not null;index" json:"timestamp"`
	Labels    string         `gorm:"type:text" json:"labels"`   // JSON string for metric labels
	Source    string         `json:"source"`                    // api, webhook, manual, etc.
	Metadata  string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// Report represents a report definition in the system.
type Report struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID       uint           `gorm:"not null;index" json:"tenant_id"`
	UserID         uint           `gorm:"not null;index" json:"user_id"`
	Name           string         `gorm:"not null" json:"name"`
	Description    string         `json:"description"`
	Type           string         `gorm:"not null" json:"type"`        // scheduled, on_demand, real_time
	Status         string         `gorm:"default:draft" json:"status"` // draft, active, paused, archived
	Schedule       string         `json:"schedule"`                    // cron expression for scheduled reports
	Format         string         `gorm:"default:pdf" json:"format"`   // pdf, csv, excel, json
	IsPublic       bool           `gorm:"default:false" json:"is_public"`
	LastGenerated  *time.Time     `json:"last_generated"`
	NextGeneration *time.Time     `json:"next_generation"`
	Config         string         `gorm:"type:text" json:"config"`   // JSON string for report configuration
	Metadata       string         `gorm:"type:text" json:"metadata"` // JSON string for additional data

	// Related entities
	Generations []ReportGeneration `gorm:"foreignKey:ReportID" json:"generations,omitempty"`
}

// ReportGeneration represents a generated report instance.
type ReportGeneration struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	ReportID    uint           `gorm:"not null;index" json:"report_id"`
	Report      Report         `gorm:"foreignKey:ReportID" json:"report"`
	Status      string         `gorm:"default:generating" json:"status"` // generating, completed, failed
	StartedAt   time.Time      `gorm:"not null" json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	FileSize    int64          `json:"file_size"`
	FilePath    string         `json:"file_path"`
	Error       string         `json:"error"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// Dashboard represents a dashboard in the system.
type Dashboard struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	IsPublic    bool           `gorm:"default:false" json:"is_public"`
	IsDefault   bool           `gorm:"default:false" json:"is_default"`
	Layout      string         `gorm:"type:text" json:"layout"`   // JSON string for dashboard layout
	Settings    string         `gorm:"type:text" json:"settings"` // JSON string for dashboard settings
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data

	// Related entities
	Widgets []DashboardWidget `gorm:"foreignKey:DashboardID" json:"widgets,omitempty"`
}

// DashboardWidget represents a widget in a dashboard.
type DashboardWidget struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	DashboardID uint           `gorm:"not null;index" json:"dashboard_id"`
	Dashboard   Dashboard      `gorm:"foreignKey:DashboardID" json:"dashboard"`
	Type        string         `gorm:"not null" json:"type"` // chart, table, metric, text
	Title       string         `gorm:"not null" json:"title"`
	Description string         `json:"description"`
	Position    int            `gorm:"default:0" json:"position"`
	Size        string         `gorm:"default:medium" json:"size"` // small, medium, large, xlarge
	Config      string         `gorm:"type:text" json:"config"`    // JSON string for widget configuration
	Metadata    string         `gorm:"type:text" json:"metadata"`  // JSON string for additional data
}

// AnalyticsEvent represents an analytics event in the system.
type AnalyticsEvent struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID   uint           `gorm:"not null;index" json:"tenant_id"`
	UserID     *uint          `gorm:"index" json:"user_id"`
	EventType  string         `gorm:"not null;index" json:"event_type"`
	EventName  string         `gorm:"not null" json:"event_name"`
	Properties string         `gorm:"type:text" json:"properties"` // JSON string for event properties
	SessionID  string         `gorm:"index" json:"session_id"`
	IPAddress  string         `json:"ip_address"`
	UserAgent  string         `json:"user_agent"`
	Referrer   string         `json:"referrer"`
	Metadata   string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// DataExport represents a data export request.
type DataExport struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	Type        string         `gorm:"not null" json:"type"`          // metrics, events, reports
	Format      string         `gorm:"not null" json:"format"`        // csv, json, excel
	Status      string         `gorm:"default:pending" json:"status"` // pending, processing, completed, failed
	Filters     string         `gorm:"type:text" json:"filters"`      // JSON string for export filters
	DateRange   string         `gorm:"type:text" json:"date_range"`   // JSON string for date range
	FileSize    int64          `json:"file_size"`
	FilePath    string         `json:"file_path"`
	DownloadURL string         `json:"download_url"`
	ExpiresAt   *time.Time     `json:"expires_at"`
	Error       string         `json:"error"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for Metric.
func (Metric) TableName() string {
	return "metrics"
}

// TableName returns the table name for MetricDataPoint.
func (MetricDataPoint) TableName() string {
	return "metric_data_points"
}

// TableName returns the table name for Report.
func (Report) TableName() string {
	return "reports"
}

// TableName returns the table name for ReportGeneration.
func (ReportGeneration) TableName() string {
	return "report_generations"
}

// TableName returns the table name for Dashboard.
func (Dashboard) TableName() string {
	return "dashboards"
}

// TableName returns the table name for DashboardWidget.
func (DashboardWidget) TableName() string {
	return "dashboard_widgets"
}

// TableName returns the table name for AnalyticsEvent.
func (AnalyticsEvent) TableName() string {
	return "analytics_events"
}

// TableName returns the table name for DataExport.
func (DataExport) TableName() string {
	return "data_exports"
}
