// Package models provides data models for the Analytics Consumer.
package models

import (
	"time"

	"gorm.io/gorm"
)

// AnalyticsEvent represents an analytics event.
type AnalyticsEvent struct {
	ID           string                 `json:"id"`
	TenantID     uint                   `json:"tenant_id"`
	UserID       *uint                  `json:"user_id"`
	SessionID    string                 `json:"session_id"`
	Type         string                 `json:"type"` // metric, page_view, user_action, performance, error
	MetricName   string                 `json:"metric_name,omitempty"`
	MetricValue  float64                `json:"metric_value,omitempty"`
	Page         string                 `json:"page,omitempty"`
	Action       string                 `json:"action,omitempty"`
	ErrorType    string                 `json:"error_type,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	Referrer     string                 `json:"referrer,omitempty"`
	Duration     int64                  `json:"duration,omitempty"` // in milliseconds
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// MetricData represents processed metric data.
type MetricData struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	MetricName  string         `gorm:"not null;index" json:"metric_name"`
	MetricValue float64        `gorm:"not null" json:"metric_value"`
	Timestamp   time.Time      `gorm:"not null;index" json:"timestamp"`
	UserID      *uint          `gorm:"index" json:"user_id"`
	SessionID   string         `gorm:"index" json:"session_id"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// AggregatedData represents aggregated analytics data.
type AggregatedData struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID        uint           `gorm:"not null;index" json:"tenant_id"`
	MetricName      string         `gorm:"not null;index" json:"metric_name"`
	AggregationType string         `gorm:"not null;index" json:"aggregation_type"` // sum, avg, min, max, count
	Value           float64        `gorm:"not null" json:"value"`
	Count           int64          `gorm:"not null" json:"count"`
	Period          string         `gorm:"not null;index" json:"period"` // hour, day, week, month
	StartTime       time.Time      `gorm:"not null;index" json:"start_time"`
	EndTime         time.Time      `gorm:"not null;index" json:"end_time"`
	Metadata        string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// PageViewData represents page view analytics data.
type PageViewData struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID  uint           `gorm:"not null;index" json:"tenant_id"`
	Page      string         `gorm:"not null;index" json:"page"`
	UserID    *uint          `gorm:"index" json:"user_id"`
	SessionID string         `gorm:"index" json:"session_id"`
	UserAgent string         `json:"user_agent"`
	IPAddress string         `json:"ip_address"`
	Referrer  string         `json:"referrer"`
	Duration  int64          `json:"duration"` // in milliseconds
	Timestamp time.Time      `gorm:"not null;index" json:"timestamp"`
	Metadata  string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// UserActionData represents user action analytics data.
type UserActionData struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID  uint           `gorm:"not null;index" json:"tenant_id"`
	Action    string         `gorm:"not null;index" json:"action"`
	UserID    *uint          `gorm:"index" json:"user_id"`
	SessionID string         `gorm:"index" json:"session_id"`
	Page      string         `gorm:"index" json:"page"`
	Element   string         `json:"element"`
	Value     string         `json:"value"`
	Timestamp time.Time      `gorm:"not null;index" json:"timestamp"`
	Metadata  string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// PerformanceData represents performance analytics data.
type PerformanceData struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	MetricName  string         `gorm:"not null;index" json:"metric_name"`
	MetricValue float64        `gorm:"not null" json:"metric_value"`
	UserID      *uint          `gorm:"index" json:"user_id"`
	SessionID   string         `gorm:"index" json:"session_id"`
	Page        string         `gorm:"index" json:"page"`
	Timestamp   time.Time      `gorm:"not null;index" json:"timestamp"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// ErrorData represents error analytics data.
type ErrorData struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID     uint           `gorm:"not null;index" json:"tenant_id"`
	ErrorType    string         `gorm:"not null;index" json:"error_type"`
	ErrorMessage string         `gorm:"type:text" json:"error_message"`
	UserID       *uint          `gorm:"index" json:"user_id"`
	SessionID    string         `gorm:"index" json:"session_id"`
	Page         string         `gorm:"index" json:"page"`
	Stack        string         `gorm:"type:text" json:"stack"`
	Timestamp    time.Time      `gorm:"not null;index" json:"timestamp"`
	Metadata     string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// ProcessingLog represents a processing log entry.
type ProcessingLog struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	MessageID      string         `gorm:"not null;index" json:"message_id"`
	EventID        string         `gorm:"not null;index" json:"event_id"`
	EventType      string         `gorm:"not null;index" json:"event_type"`
	Status         string         `gorm:"not null" json:"status"` // processing, completed, failed
	ProcessingTime int64          `json:"processing_time"`        // in milliseconds
	Error          string         `json:"error"`
	RetryCount     int            `gorm:"default:0" json:"retry_count"`
	Metadata       string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// QueueMessage represents a message in the queue.
type QueueMessage struct {
	ID          string         `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	QueueName   string         `gorm:"not null;index" json:"queue_name"`
	EventType   string         `gorm:"not null;index" json:"event_type"`
	Data        string         `gorm:"type:text;not null" json:"data"`
	Priority    int            `gorm:"default:0" json:"priority"`
	Status      string         `gorm:"default:pending" json:"status"` // pending, processing, completed, failed
	RetryCount  int            `gorm:"default:0" json:"retry_count"`
	MaxRetries  int            `gorm:"default:3" json:"max_retries"`
	ProcessedAt *time.Time     `json:"processed_at"`
	FailedAt    *time.Time     `json:"failed_at"`
	Error       string         `json:"error"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for MetricData.
func (MetricData) TableName() string {
	return "metric_data"
}

// TableName returns the table name for AggregatedData.
func (AggregatedData) TableName() string {
	return "aggregated_data"
}

// TableName returns the table name for PageViewData.
func (PageViewData) TableName() string {
	return "page_view_data"
}

// TableName returns the table name for UserActionData.
func (UserActionData) TableName() string {
	return "user_action_data"
}

// TableName returns the table name for PerformanceData.
func (PerformanceData) TableName() string {
	return "performance_data"
}

// TableName returns the table name for ErrorData.
func (ErrorData) TableName() string {
	return "error_data"
}

// TableName returns the table name for ProcessingLog.
func (ProcessingLog) TableName() string {
	return "processing_logs"
}

// TableName returns the table name for QueueMessage.
func (QueueMessage) TableName() string {
	return "queue_messages"
}

