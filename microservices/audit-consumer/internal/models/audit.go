// Package models provides data models for the Audit Consumer.
package models

import (
	"time"

	"gorm.io/gorm"
)

// AuditEvent represents an audit event.
type AuditEvent struct {
	ID           string                 `json:"id"`
	TenantID     uint                   `json:"tenant_id"`
	UserID       *uint                  `json:"user_id"`
	SessionID    string                 `json:"session_id"`
	Action       string                 `json:"action"`   // create, read, update, delete, login, logout
	Resource     string                 `json:"resource"` // user, tenant, component, incident, etc.
	ResourceID   string                 `json:"resource_id"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	RequestID    string                 `json:"request_id"`
	Status       string                 `json:"status"` // success, failure, error
	ErrorMessage string                 `json:"error_message,omitempty"`
	Changes      map[string]interface{} `json:"changes,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// AuditLog represents an audit log entry.
type AuditLog struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID     uint           `gorm:"not null;index" json:"tenant_id"`
	UserID       *uint          `gorm:"index" json:"user_id"`
	SessionID    string         `gorm:"index" json:"session_id"`
	Action       string         `gorm:"not null;index" json:"action"`
	Resource     string         `gorm:"not null;index" json:"resource"`
	ResourceID   string         `gorm:"index" json:"resource_id"`
	IPAddress    string         `json:"ip_address"`
	UserAgent    string         `json:"user_agent"`
	RequestID    string         `gorm:"index" json:"request_id"`
	Status       string         `gorm:"not null;index" json:"status"`
	ErrorMessage string         `json:"error_message"`
	Changes      string         `gorm:"type:text" json:"changes"`  // JSON string for changes
	Metadata     string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
	Timestamp    time.Time      `gorm:"not null;index" json:"timestamp"`
}

// ComplianceLog represents a compliance log entry.
type ComplianceLog struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID       uint           `gorm:"not null;index" json:"tenant_id"`
	AuditLogID     uint           `gorm:"not null;index" json:"audit_log_id"`
	AuditLog       AuditLog       `gorm:"foreignKey:AuditLogID" json:"audit_log"`
	ComplianceType string         `gorm:"not null;index" json:"compliance_type"` // gdpr, sox, pci, hipaa
	Rule           string         `gorm:"not null;index" json:"rule"`
	Status         string         `gorm:"not null;index" json:"status"`   // compliant, non_compliant, warning
	Severity       string         `gorm:"not null;index" json:"severity"` // low, medium, high, critical
	Message        string         `gorm:"type:text" json:"message"`
	Timestamp      time.Time      `gorm:"not null;index" json:"timestamp"`
	Metadata       string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
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

// TableName returns the table name for AuditLog.
func (AuditLog) TableName() string {
	return "audit_logs"
}

// TableName returns the table name for ComplianceLog.
func (ComplianceLog) TableName() string {
	return "compliance_logs"
}

// TableName returns the table name for ProcessingLog.
func (ProcessingLog) TableName() string {
	return "processing_logs"
}

// TableName returns the table name for QueueMessage.
func (QueueMessage) TableName() string {
	return "queue_messages"
}

