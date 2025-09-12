// Package models provides data models for the Notification Consumer.
package models

import (
	"time"

	"gorm.io/gorm"
)

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

// NotificationEvent represents a notification event.
type NotificationEvent struct {
	ID             string                 `json:"id"`
	NotificationID string                 `json:"notification_id"`
	TenantID       uint                   `json:"tenant_id"`
	UserID         *uint                  `json:"user_id"`
	Type           string                 `json:"type"` // email, sms, webhook, push
	Priority       string                 `json:"priority"`
	Subject        string                 `json:"subject"`
	Content        string                 `json:"content"`
	Recipient      string                 `json:"recipient"`
	DeviceToken    string                 `json:"device_token,omitempty"`
	WebhookURL     string                 `json:"webhook_url,omitempty"`
	TemplateID     *uint                  `json:"template_id,omitempty"`
	Variables      map[string]interface{} `json:"variables,omitempty"`
	ScheduledAt    *time.Time             `json:"scheduled_at,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
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

// EmailNotification represents an email notification.
type EmailNotification struct {
	ID             string                 `json:"id"`
	NotificationID string                 `json:"notification_id"`
	TenantID       uint                   `json:"tenant_id"`
	UserID         *uint                  `json:"user_id"`
	To             []string               `json:"to"`
	CC             []string               `json:"cc,omitempty"`
	BCC            []string               `json:"bcc,omitempty"`
	Subject        string                 `json:"subject"`
	Body           string                 `json:"body"`
	HTMLBody       string                 `json:"html_body,omitempty"`
	Attachments    []EmailAttachment      `json:"attachments,omitempty"`
	TemplateID     *uint                  `json:"template_id,omitempty"`
	Variables      map[string]interface{} `json:"variables,omitempty"`
	Priority       string                 `json:"priority"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// EmailAttachment represents an email attachment.
type EmailAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Data        string `json:"data"` // base64 encoded
	Inline      bool   `json:"inline"`
}

// SMSNotification represents an SMS notification.
type SMSNotification struct {
	ID             string                 `json:"id"`
	NotificationID string                 `json:"notification_id"`
	TenantID       uint                   `json:"tenant_id"`
	UserID         *uint                  `json:"user_id"`
	To             string                 `json:"to"`
	Message        string                 `json:"message"`
	TemplateID     *uint                  `json:"template_id,omitempty"`
	Variables      map[string]interface{} `json:"variables,omitempty"`
	Priority       string                 `json:"priority"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// WebhookNotification represents a webhook notification.
type WebhookNotification struct {
	ID             string                 `json:"id"`
	NotificationID string                 `json:"notification_id"`
	TenantID       uint                   `json:"tenant_id"`
	UserID         *uint                  `json:"user_id"`
	URL            string                 `json:"url"`
	Method         string                 `json:"method"`
	Headers        map[string]string      `json:"headers,omitempty"`
	Body           string                 `json:"body"`
	Timeout        int                    `json:"timeout"` // in seconds
	RetryCount     int                    `json:"retry_count"`
	MaxRetries     int                    `json:"max_retries"`
	Priority       string                 `json:"priority"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// PushNotification represents a push notification.
type PushNotification struct {
	ID             string                 `json:"id"`
	NotificationID string                 `json:"notification_id"`
	TenantID       uint                   `json:"tenant_id"`
	UserID         *uint                  `json:"user_id"`
	DeviceToken    string                 `json:"device_token"`
	Title          string                 `json:"title"`
	Body           string                 `json:"body"`
	Data           map[string]interface{} `json:"data,omitempty"`
	Badge          *int                   `json:"badge,omitempty"`
	Sound          string                 `json:"sound,omitempty"`
	Priority       string                 `json:"priority"`
	TTL            int                    `json:"ttl"` // time to live in seconds
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// TableName returns the table name for QueueMessage.
func (QueueMessage) TableName() string {
	return "queue_messages"
}

// TableName returns the table name for ProcessingLog.
func (ProcessingLog) TableName() string {
	return "processing_logs"
}

