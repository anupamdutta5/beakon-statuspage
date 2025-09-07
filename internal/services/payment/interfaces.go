package payment

import (
	"context"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
)

// PaymentRequest represents a payment request
type PaymentRequest struct {
	TenantID     uint                 `json:"tenant_id"`
	Amount       float64              `json:"amount"`
	Currency     string               `json:"currency"`
	Method       string               `json:"method"` // card, upi, netbanking, wallet
	Description  string               `json:"description"`
	Customer     CustomerInfo         `json:"customer"`
	Metadata     map[string]string    `json:"metadata"`
	ReturnURL    string               `json:"return_url"`
	WebhookURL   string               `json:"webhook_url"`
	Subscription *SubscriptionRequest `json:"subscription,omitempty"`
}

// CustomerInfo represents customer information
type CustomerInfo struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Phone   string   `json:"phone"`
	Address *Address `json:"address,omitempty"`
}

// Address represents customer address
type Address struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// SubscriptionRequest represents subscription details
type SubscriptionRequest struct {
	PlanID          uint   `json:"plan_id"`
	PlanSlug        string `json:"plan_slug"`
	BillingInterval string `json:"billing_interval"` // monthly, yearly
	TrialDays       int    `json:"trial_days"`
}

// PaymentResponse represents a payment response
type PaymentResponse struct {
	ID              string                 `json:"id"`
	Status          string                 `json:"status"` // pending, processing, completed, failed, cancelled
	Amount          float64                `json:"amount"`
	Currency        string                 `json:"currency"`
	Method          string                 `json:"method"`
	Gateway         string                 `json:"gateway"`
	GatewayID       string                 `json:"gateway_id"`
	GatewayResponse map[string]interface{} `json:"gateway_response"`
	RedirectURL     string                 `json:"redirect_url,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// RefundRequest represents a refund request
type RefundRequest struct {
	PaymentID string            `json:"payment_id"`
	Amount    float64           `json:"amount"` // 0 means full refund
	Reason    string            `json:"reason"`
	Metadata  map[string]string `json:"metadata"`
}

// RefundResponse represents a refund response
type RefundResponse struct {
	ID              string                 `json:"id"`
	PaymentID       string                 `json:"payment_id"`
	Amount          float64                `json:"amount"`
	Status          string                 `json:"status"` // pending, processed, failed
	GatewayID       string                 `json:"gateway_id"`
	GatewayResponse map[string]interface{} `json:"gateway_response"`
	CreatedAt       time.Time              `json:"created_at"`
}

// WebhookEvent represents a webhook event
type WebhookEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Gateway   string                 `json:"gateway"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
}

// PaymentGateway defines the interface for payment gateways
type PaymentGateway interface {
	// Initialize initializes the gateway with configuration
	Initialize(config map[string]string) error

	// CreatePayment creates a new payment
	CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)

	// GetPayment retrieves payment details
	GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error)

	// UpdatePayment updates payment details
	UpdatePayment(ctx context.Context, paymentID string, updates map[string]interface{}) (*PaymentResponse, error)

	// CancelPayment cancels a payment
	CancelPayment(ctx context.Context, paymentID string) (*PaymentResponse, error)

	// RefundPayment processes a refund
	RefundPayment(ctx context.Context, req *RefundRequest) (*RefundResponse, error)

	// CreateSubscription creates a subscription
	CreateSubscription(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)

	// GetSubscription retrieves subscription details
	GetSubscription(ctx context.Context, subscriptionID string) (*PaymentResponse, error)

	// CancelSubscription cancels a subscription
	CancelSubscription(ctx context.Context, subscriptionID string) (*PaymentResponse, error)

	// ProcessWebhook processes webhook events
	ProcessWebhook(ctx context.Context, payload []byte, signature string) (*WebhookEvent, error)

	// VerifyWebhook verifies webhook signature
	VerifyWebhook(payload []byte, signature string) error

	// GetSupportedMethods returns supported payment methods
	GetSupportedMethods() []string

	// GetGatewayName returns the gateway name
	GetGatewayName() string

	// IsHealthy checks if the gateway is healthy
	IsHealthy(ctx context.Context) error
}

// PaymentService defines the main payment service interface
type PaymentService interface {
	// CreatePayment creates a payment using the appropriate gateway
	CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)

	// GetPayment retrieves payment details
	GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error)

	// UpdatePayment updates payment details
	UpdatePayment(ctx context.Context, paymentID string, updates map[string]interface{}) (*PaymentResponse, error)

	// CancelPayment cancels a payment
	CancelPayment(ctx context.Context, paymentID string) (*PaymentResponse, error)

	// RefundPayment processes a refund
	RefundPayment(ctx context.Context, req *RefundRequest) (*RefundResponse, error)

	// CreateSubscription creates a subscription
	CreateSubscription(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)

	// GetSubscription retrieves subscription details
	GetSubscription(ctx context.Context, subscriptionID string) (*PaymentResponse, error)

	// CancelSubscription cancels a subscription
	CancelSubscription(ctx context.Context, subscriptionID string) (*PaymentResponse, error)

	// ProcessWebhook processes webhook events
	ProcessWebhook(ctx context.Context, gateway string, payload []byte, signature string) (*WebhookEvent, error)

	// GetSupportedMethods returns all supported payment methods
	GetSupportedMethods() []string

	// GetGateways returns available gateways
	GetGateways() []string

	// GetGatewayForMethod returns the best gateway for a payment method
	GetGatewayForMethod(method string) (PaymentGateway, error)

	// RecordPayment records a payment in the database
	RecordPayment(ctx context.Context, response *PaymentResponse) error

	// RecordRefund records a refund in the database
	RecordRefund(ctx context.Context, response *RefundResponse) error

	// GetPaymentHistory returns payment history for a tenant
	GetPaymentHistory(ctx context.Context, tenantID uint, limit, offset int) ([]*models.BillingPayment, error)

	// GetInvoiceHistory returns invoice history for a tenant
	GetInvoiceHistory(ctx context.Context, tenantID uint, limit, offset int) ([]*models.BillingInvoice, error)

	// GetPaymentMetrics returns payment metrics for a tenant
	GetPaymentMetrics(ctx context.Context, tenantID uint, period string) (*PaymentMetrics, error)

	// GetConversionFunnel returns conversion funnel data for a tenant
	GetConversionFunnel(ctx context.Context, tenantID uint, period string) (*ConversionFunnel, error)
}

// PaymentRetryService defines the interface for payment retry logic
type PaymentRetryService interface {
	// ScheduleRetry schedules a payment retry
	ScheduleRetry(ctx context.Context, paymentID string, delay time.Duration) error

	// ProcessRetries processes scheduled retries
	ProcessRetries(ctx context.Context) error

	// GetRetryCount returns the retry count for a payment
	GetRetryCount(ctx context.Context, paymentID string) (int, error)

	// ShouldRetry determines if a payment should be retried
	ShouldRetry(ctx context.Context, paymentID string, errorType string) (bool, time.Duration, error)
}

// PaymentNotificationService defines the interface for payment notifications
type PaymentNotificationService interface {
	// SendPaymentSuccessNotification sends payment success notification
	SendPaymentSuccessNotification(ctx context.Context, payment *PaymentResponse) error

	// SendPaymentFailureNotification sends payment failure notification
	SendPaymentFailureNotification(ctx context.Context, payment *PaymentResponse, reason string) error

	// SendRefundNotification sends refund notification
	SendRefundNotification(ctx context.Context, refund *RefundResponse) error

	// SendSubscriptionNotification sends subscription notification
	SendSubscriptionNotification(ctx context.Context, subscription *PaymentResponse, eventType string) error
}

// PaymentAnalyticsService defines the interface for payment analytics
type PaymentAnalyticsService interface {
	// RecordPaymentEvent records a payment event for analytics
	RecordPaymentEvent(ctx context.Context, event *PaymentEvent) error

	// GetPaymentMetrics returns payment metrics
	GetPaymentMetrics(ctx context.Context, tenantID uint, period string) (*PaymentMetrics, error)

	// GetConversionFunnel returns conversion funnel data
	GetConversionFunnel(ctx context.Context, tenantID uint, period string) (*ConversionFunnel, error)
}

// PaymentEvent represents a payment event for analytics
type PaymentEvent struct {
	TenantID  uint                   `json:"tenant_id"`
	EventType string                 `json:"event_type"` // created, completed, failed, refunded
	PaymentID string                 `json:"payment_id"`
	Amount    float64                `json:"amount"`
	Currency  string                 `json:"currency"`
	Method    string                 `json:"method"`
	Gateway   string                 `json:"gateway"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// PaymentMetrics represents payment metrics
type PaymentMetrics struct {
	TotalPayments      int     `json:"total_payments"`
	SuccessfulPayments int     `json:"successful_payments"`
	FailedPayments     int     `json:"failed_payments"`
	TotalAmount        float64 `json:"total_amount"`
	SuccessfulAmount   float64 `json:"successful_amount"`
	FailedAmount       float64 `json:"failed_amount"`
	ConversionRate     float64 `json:"conversion_rate"`
	AverageAmount      float64 `json:"average_amount"`
	RefundCount        int     `json:"refund_count"`
	RefundAmount       float64 `json:"refund_amount"`
}

// ConversionFunnel represents conversion funnel data
type ConversionFunnel struct {
	Steps []ConversionStep `json:"steps"`
}

// ConversionStep represents a step in the conversion funnel
type ConversionStep struct {
	Name        string  `json:"name"`
	Count       int     `json:"count"`
	Percentage  float64 `json:"percentage"`
	DropOffRate float64 `json:"drop_off_rate"`
}
