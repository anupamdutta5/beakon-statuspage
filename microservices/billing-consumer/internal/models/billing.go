// Package models provides data models for the Billing Consumer.
package models

import (
	"time"

	"gorm.io/gorm"
)

// BillingEvent represents a billing event.
type BillingEvent struct {
	ID             string                 `json:"id"`
	TenantID       uint                   `json:"tenant_id"`
	UserID         *uint                  `json:"user_id"`
	Type           string                 `json:"type"` // subscription, usage, payment, invoice
	SubscriptionID string                 `json:"subscription_id,omitempty"`
	PaymentID      string                 `json:"payment_id,omitempty"`
	InvoiceID      string                 `json:"invoice_id,omitempty"`
	Amount         float64                `json:"amount"`
	Currency       string                 `json:"currency"`
	Description    string                 `json:"description"`
	Status         string                 `json:"status"` // pending, completed, failed, cancelled
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

// BillingRecord represents a billing record.
type BillingRecord struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID       uint           `gorm:"not null;index" json:"tenant_id"`
	UserID         *uint          `gorm:"index" json:"user_id"`
	Type           string         `gorm:"not null;index" json:"type"`
	SubscriptionID string         `gorm:"index" json:"subscription_id"`
	PaymentID      string         `gorm:"index" json:"payment_id"`
	InvoiceID      string         `gorm:"index" json:"invoice_id"`
	Amount         float64        `gorm:"not null" json:"amount"`
	Currency       string         `gorm:"not null" json:"currency"`
	Description    string         `json:"description"`
	Status         string         `gorm:"not null;index" json:"status"`
	Metadata       string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
	Timestamp      time.Time      `gorm:"not null;index" json:"timestamp"`
}

// InvoiceRecord represents an invoice record.
type InvoiceRecord struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID       uint           `gorm:"not null;index" json:"tenant_id"`
	UserID         *uint          `gorm:"index" json:"user_id"`
	InvoiceID      string         `gorm:"not null;index" json:"invoice_id"`
	SubscriptionID string         `gorm:"index" json:"subscription_id"`
	Amount         float64        `gorm:"not null" json:"amount"`
	TaxAmount      float64        `json:"tax_amount"`
	TotalAmount    float64        `gorm:"not null" json:"total_amount"`
	Currency       string         `gorm:"not null" json:"currency"`
	Status         string         `gorm:"not null;index" json:"status"` // draft, sent, paid, overdue, cancelled
	DueDate        time.Time      `gorm:"index" json:"due_date"`
	PaidAt         *time.Time     `json:"paid_at"`
	Metadata       string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
	Timestamp      time.Time      `gorm:"not null;index" json:"timestamp"`
}

// PaymentRecord represents a payment record.
type PaymentRecord struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID        uint           `gorm:"not null;index" json:"tenant_id"`
	UserID          *uint          `gorm:"index" json:"user_id"`
	PaymentID       string         `gorm:"not null;index" json:"payment_id"`
	InvoiceID       string         `gorm:"index" json:"invoice_id"`
	SubscriptionID  string         `gorm:"index" json:"subscription_id"`
	Amount          float64        `gorm:"not null" json:"amount"`
	Currency        string         `gorm:"not null" json:"currency"`
	PaymentMethod   string         `gorm:"not null;index" json:"payment_method"` // card, bank_transfer, paypal, etc.
	Status          string         `gorm:"not null;index" json:"status"`         // pending, completed, failed, refunded
	TransactionID   string         `gorm:"index" json:"transaction_id"`
	GatewayResponse string         `gorm:"type:text" json:"gateway_response"`
	Metadata        string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
	Timestamp       time.Time      `gorm:"not null;index" json:"timestamp"`
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

// TableName returns the table name for BillingRecord.
func (BillingRecord) TableName() string {
	return "billing_records"
}

// TableName returns the table name for InvoiceRecord.
func (InvoiceRecord) TableName() string {
	return "invoice_records"
}

// TableName returns the table name for PaymentRecord.
func (PaymentRecord) TableName() string {
	return "payment_records"
}

// TableName returns the table name for ProcessingLog.
func (ProcessingLog) TableName() string {
	return "processing_logs"
}

// TableName returns the table name for QueueMessage.
func (QueueMessage) TableName() string {
	return "queue_messages"
}

