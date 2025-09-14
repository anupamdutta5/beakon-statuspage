// Package models provides data models for the Payment Service.
package models

import (
	"time"

	"gorm.io/gorm"
)

// Payment represents a payment in the system.
type Payment struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID        uint           `gorm:"not null;index" json:"tenant_id"`
	UserID          uint           `gorm:"not null;index" json:"user_id"`
	SubscriptionID  *uint          `gorm:"index" json:"subscription_id,omitempty"`
	Amount          float64        `gorm:"not null" json:"amount"`
	Currency        string         `gorm:"not null;default:USD" json:"currency"`
	Status          string         `gorm:"default:pending" json:"status"`     // pending, processing, completed, failed, cancelled, refunded
	PaymentMethod   string         `gorm:"not null" json:"payment_method"`    // stripe, paypal, razorpay, bank_transfer
	Gateway         string         `gorm:"not null" json:"gateway"`           // stripe, paypal, razorpay
	GatewayID       string         `gorm:"index" json:"gateway_id"`           // External gateway payment ID
	GatewayResponse string         `gorm:"type:text" json:"gateway_response"` // JSON response from gateway
	Description     string         `json:"description"`
	Metadata        string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
	ProcessedAt     *time.Time     `json:"processed_at"`
	FailedAt        *time.Time     `json:"failed_at"`
	RefundedAt      *time.Time     `json:"refunded_at"`
	RefundAmount    float64        `gorm:"default:0" json:"refund_amount"`
	RefundReason    string         `json:"refund_reason"`

	// Related entities
	Transactions []PaymentTransaction `gorm:"foreignKey:PaymentID" json:"transactions,omitempty"`
}

// PaymentTransaction represents individual transactions within a payment.
type PaymentTransaction struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	PaymentID   uint           `gorm:"not null;index" json:"payment_id"`
	Payment     Payment        `gorm:"foreignKey:PaymentID" json:"payment"`
	Type        string         `gorm:"not null" json:"type"` // charge, refund, partial_refund, fee
	Amount      float64        `gorm:"not null" json:"amount"`
	Currency    string         `gorm:"not null" json:"currency"`
	Status      string         `gorm:"default:pending" json:"status"` // pending, completed, failed
	GatewayID   string         `gorm:"index" json:"gateway_id"`       // External gateway transaction ID
	Description string         `json:"description"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
	ProcessedAt *time.Time     `json:"processed_at"`
}

// Subscription represents a subscription in the system.
type Subscription struct {
	ID                 uint           `gorm:"primarykey" json:"id"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID           uint           `gorm:"not null;index" json:"tenant_id"`
	UserID             uint           `gorm:"not null;index" json:"user_id"`
	PlanID             uint           `gorm:"not null;index" json:"plan_id"`
	Plan               Plan           `gorm:"foreignKey:PlanID" json:"plan"`
	Status             string         `gorm:"default:active" json:"status"`  // active, cancelled, expired, suspended
	BillingCycle       string         `gorm:"not null" json:"billing_cycle"` // monthly, yearly
	Amount             float64        `gorm:"not null" json:"amount"`
	Currency           string         `gorm:"not null;default:USD" json:"currency"`
	StartedAt          time.Time      `gorm:"not null" json:"started_at"`
	CurrentPeriodStart time.Time      `gorm:"not null" json:"current_period_start"`
	CurrentPeriodEnd   time.Time      `gorm:"not null" json:"current_period_end"`
	NextBillingDate    time.Time      `gorm:"not null" json:"next_billing_date"`
	CancelledAt        *time.Time     `json:"cancelled_at"`
	CancelReason       string         `json:"cancel_reason"`
	Gateway            string         `gorm:"not null" json:"gateway"`           // stripe, paypal, razorpay
	GatewayID          string         `gorm:"index" json:"gateway_id"`           // External gateway subscription ID
	GatewayResponse    string         `gorm:"type:text" json:"gateway_response"` // JSON response from gateway
	Metadata           string         `gorm:"type:text" json:"metadata"`         // JSON string for additional data

	// Related entities
	Payments []Payment `gorm:"foreignKey:SubscriptionID" json:"payments,omitempty"`
}

// Plan represents a subscription plan.
type Plan struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID     uint           `gorm:"not null;index" json:"tenant_id"`
	Name         string         `gorm:"not null" json:"name"`
	Description  string         `json:"description"`
	Price        float64        `gorm:"not null" json:"price"`
	Currency     string         `gorm:"not null;default:USD" json:"currency"`
	BillingCycle string         `gorm:"not null" json:"billing_cycle"` // monthly, yearly
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	IsPublic     bool           `gorm:"default:true" json:"is_public"`
	Features     string         `gorm:"type:text" json:"features"` // JSON string for plan features
	Limits       string         `gorm:"type:text" json:"limits"`   // JSON string for plan limits
	TrialDays    int            `gorm:"default:0" json:"trial_days"`
	SortOrder    int            `gorm:"default:0" json:"sort_order"`
	Metadata     string         `gorm:"type:text" json:"metadata"` // JSON string for additional data

	// Related entities
	Subscriptions []Subscription `gorm:"foreignKey:PlanID" json:"subscriptions,omitempty"`
}

// Invoice represents an invoice in the system.
type Invoice struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID       uint           `gorm:"not null;index" json:"tenant_id"`
	UserID         uint           `gorm:"not null;index" json:"user_id"`
	SubscriptionID *uint          `gorm:"index" json:"subscription_id"`
	Subscription   *Subscription  `gorm:"foreignKey:SubscriptionID" json:"subscription,omitempty"`
	InvoiceNumber  string         `gorm:"uniqueIndex;not null" json:"invoice_number"`
	Amount         float64        `gorm:"not null" json:"amount"`
	Currency       string         `gorm:"not null;default:USD" json:"currency"`
	Status         string         `gorm:"default:draft" json:"status"` // draft, sent, paid, overdue, cancelled
	DueDate        time.Time      `gorm:"not null" json:"due_date"`
	PaidAt         *time.Time     `json:"paid_at"`
	PaymentID      *uint          `gorm:"index" json:"payment_id"`
	Payment        *Payment       `gorm:"foreignKey:PaymentID" json:"payment,omitempty"`
	Description    string         `json:"description"`
	Items          string         `gorm:"type:text" json:"items"` // JSON string for invoice items
	TaxAmount      float64        `gorm:"default:0" json:"tax_amount"`
	DiscountAmount float64        `gorm:"default:0" json:"discount_amount"`
	TotalAmount    float64        `gorm:"not null" json:"total_amount"`
	Metadata       string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// BillingUsage represents usage tracking for billing.
type BillingUsage struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID       uint           `gorm:"not null;index" json:"tenant_id"`
	UserID         uint           `gorm:"not null;index" json:"user_id"`
	SubscriptionID uint           `gorm:"not null;index" json:"subscription_id"`
	Subscription   Subscription   `gorm:"foreignKey:SubscriptionID" json:"subscription"`
	MetricType     string         `gorm:"not null" json:"metric_type"` // api_calls, storage, users, etc.
	Usage          float64        `gorm:"not null" json:"usage"`
	Limit          float64        `gorm:"not null" json:"limit"`
	Overage        float64        `gorm:"default:0" json:"overage"`
	BillingPeriod  time.Time      `gorm:"not null;index" json:"billing_period"`
	Metadata       string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// PaymentWebhook represents webhook events from payment gateways.
type PaymentWebhook struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Gateway     string         `gorm:"not null" json:"gateway"` // stripe, paypal, razorpay
	EventType   string         `gorm:"not null" json:"event_type"`
	GatewayID   string         `gorm:"index" json:"gateway_id"`  // External gateway event ID
	Payload     string         `gorm:"type:text" json:"payload"` // Raw webhook payload
	Processed   bool           `gorm:"default:false" json:"processed"`
	ProcessedAt *time.Time     `json:"processed_at"`
	Error       string         `json:"error"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for Payment.
func (Payment) TableName() string {
	return "payments"
}

// TableName returns the table name for PaymentTransaction.
func (PaymentTransaction) TableName() string {
	return "payment_transactions"
}

// TableName returns the table name for Subscription.
func (Subscription) TableName() string {
	return "subscriptions"
}

// TableName returns the table name for Plan.
func (Plan) TableName() string {
	return "plans"
}

// TableName returns the table name for Invoice.
func (Invoice) TableName() string {
	return "invoices"
}

// TableName returns the table name for BillingUsage.
func (BillingUsage) TableName() string {
	return "billing_usage"
}

// TableName returns the table name for PaymentWebhook.
func (PaymentWebhook) TableName() string {
	return "payment_webhooks"
}
