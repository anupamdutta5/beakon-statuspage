// Package models provides data models for the Notification Service.
package models

import (
	"time"

	"gorm.io/gorm"
)

// Notification represents a notification in the system.
type Notification struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	UserID      *uint          `gorm:"index" json:"user_id"`
	TemplateID  *uint          `gorm:"index" json:"template_id"`
	Template    *Template      `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	Type        string         `gorm:"not null" json:"type"`           // email, sms, webhook, push
	Status      string         `gorm:"default:pending" json:"status"`  // pending, sent, failed, cancelled
	Priority    string         `gorm:"default:normal" json:"priority"` // low, normal, high, urgent
	Subject     string         `json:"subject"`
	Content     string         `gorm:"type:text" json:"content"`
	Recipients  string         `gorm:"type:text" json:"recipients"` // JSON array of recipients
	Metadata    string         `gorm:"type:text" json:"metadata"`   // JSON string for additional data
	ScheduledAt *time.Time     `json:"scheduled_at"`
	SentAt      *time.Time     `json:"sent_at"`
	FailedAt    *time.Time     `json:"failed_at"`
	RetryCount  int            `gorm:"default:0" json:"retry_count"`
	MaxRetries  int            `gorm:"default:3" json:"max_retries"`
	Error       string         `json:"error"`

	// Related entities
	Deliveries []Delivery `gorm:"foreignKey:NotificationID" json:"deliveries,omitempty"`
}

// Template represents a notification template.
type Template struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	Type        string         `gorm:"not null" json:"type"` // email, sms, webhook, push
	Category    string         `json:"category"`             // incident, maintenance, general, etc.
	Subject     string         `json:"subject"`
	Content     string         `gorm:"type:text" json:"content"`
	Variables   string         `gorm:"type:text" json:"variables"` // JSON array of available variables
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	IsDefault   bool           `gorm:"default:false" json:"is_default"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data

	// Related entities
	Notifications []Notification `gorm:"foreignKey:TemplateID" json:"notifications,omitempty"`
}

// Channel represents a notification channel.
type Channel struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	Type        string         `gorm:"not null" json:"type"`    // email, sms, webhook, push
	Provider    string         `json:"provider"`                // smtp, twilio, slack, etc.
	Config      string         `gorm:"type:text" json:"config"` // JSON string for channel configuration
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	IsDefault   bool           `gorm:"default:false" json:"is_default"`
	Priority    int            `gorm:"default:0" json:"priority"` // for channel selection priority
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// Subscription represents a notification subscription.
type Subscription struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	ChannelID   uint           `gorm:"not null;index" json:"channel_id"`
	Channel     Channel        `gorm:"foreignKey:ChannelID" json:"channel"`
	EventTypes  string         `gorm:"type:text" json:"event_types"` // JSON array of event types
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	Preferences string         `gorm:"type:text" json:"preferences"` // JSON string for user preferences
	Metadata    string         `gorm:"type:text" json:"metadata"`    // JSON string for additional data
}

// Delivery represents a notification delivery attempt.
type Delivery struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	NotificationID  uint           `gorm:"not null;index" json:"notification_id"`
	Notification    Notification   `gorm:"foreignKey:NotificationID" json:"notification"`
	ChannelID       uint           `gorm:"not null;index" json:"channel_id"`
	Channel         Channel        `gorm:"foreignKey:ChannelID" json:"channel"`
	Recipient       string         `gorm:"not null" json:"recipient"`
	Status          string         `gorm:"default:pending" json:"status"` // pending, sent, failed, bounced
	Attempt         int            `gorm:"default:1" json:"attempt"`
	MaxAttempts     int            `gorm:"default:3" json:"max_attempts"`
	SentAt          *time.Time     `json:"sent_at"`
	FailedAt        *time.Time     `json:"failed_at"`
	ResponseCode    int            `json:"response_code"`
	ResponseMessage string         `json:"response_message"`
	Error           string         `json:"error"`
	Metadata        string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// WebhookEvent represents a webhook event.
type WebhookEvent struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	EventType   string         `gorm:"not null;index" json:"event_type"`
	EventData   string         `gorm:"type:text" json:"event_data"`   // JSON string for event data
	Source      string         `json:"source"`                        // service that generated the event
	Status      string         `gorm:"default:pending" json:"status"` // pending, processed, failed
	ProcessedAt *time.Time     `json:"processed_at"`
	FailedAt    *time.Time     `json:"failed_at"`
	RetryCount  int            `gorm:"default:0" json:"retry_count"`
	MaxRetries  int            `gorm:"default:3" json:"max_retries"`
	Error       string         `json:"error"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// NotificationLog represents a notification log entry.
type NotificationLog struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID       uint           `gorm:"not null;index" json:"tenant_id"`
	NotificationID *uint          `gorm:"index" json:"notification_id"`
	Notification   *Notification  `gorm:"foreignKey:NotificationID" json:"notification,omitempty"`
	Level          string         `gorm:"not null" json:"level"` // info, warn, error, debug
	Message        string         `gorm:"type:text;not null" json:"message"`
	Details        string         `gorm:"type:text" json:"details"`  // JSON string for additional details
	Metadata       string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for Notification.
func (Notification) TableName() string {
	return "notifications"
}

// TableName returns the table name for Template.
func (Template) TableName() string {
	return "templates"
}

// TableName returns the table name for Channel.
func (Channel) TableName() string {
	return "channels"
}

// TableName returns the table name for Subscription.
func (Subscription) TableName() string {
	return "subscriptions"
}

// TableName returns the table name for Delivery.
func (Delivery) TableName() string {
	return "deliveries"
}

// TableName returns the table name for WebhookEvent.
func (WebhookEvent) TableName() string {
	return "webhook_events"
}

// TableName returns the table name for NotificationLog.
func (NotificationLog) TableName() string {
	return "notification_logs"
}
