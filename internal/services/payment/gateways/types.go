package gateways

import (
	"context"
	"time"
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
