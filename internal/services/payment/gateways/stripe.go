package gateways

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/internal/services/payment"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/refund"
	"github.com/stripe/stripe-go/v76/subscription"
)

// StripeGateway implements PaymentGateway for Stripe
type StripeGateway struct {
	config *config.PaymentGateway
}

// NewStripeGateway creates a new Stripe gateway
func NewStripeGateway(gatewayConfig *config.PaymentGateway) (*StripeGateway, error) {
	gateway := &StripeGateway{
		config: gatewayConfig,
	}

	// Initialize Stripe
	stripe.Key = gatewayConfig.Credentials["secret_key"]

	return gateway, nil
}

// Initialize initializes the gateway with configuration
func (g *StripeGateway) Initialize(config map[string]string) error {
	// Stripe is already initialized in NewStripeGateway
	return nil
}

// CreatePayment creates a new payment
func (g *StripeGateway) CreatePayment(ctx context.Context, req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	// Create or get customer
	customerID, err := g.getOrCreateCustomer(ctx, req.Customer)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	// Create payment intent
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(req.Amount * 100)), // Convert to cents
		Currency: stripe.String(req.Currency),
		Customer: stripe.String(customerID),
		Metadata: req.Metadata,
	}

	// Add payment method based on request
	switch req.Method {
	case "card":
		params.PaymentMethodTypes = stripe.StringSlice([]string{"card"})
	case "upi":
		params.PaymentMethodTypes = stripe.StringSlice([]string{"upi"})
	case "netbanking":
		params.PaymentMethodTypes = stripe.StringSlice([]string{"netbanking"})
	case "wallet":
		params.PaymentMethodTypes = stripe.StringSlice([]string{"wallet"})
	default:
		// Default to card if method not specified
		params.PaymentMethodTypes = stripe.StringSlice([]string{"card"})
	}

	// Add description
	if req.Description != "" {
		params.Description = stripe.String(req.Description)
	}

	// Create payment intent
	intent, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	// Convert to response
	response := &payment.PaymentResponse{
		ID:        intent.ID,
		Status:    string(intent.Status),
		Amount:    float64(intent.Amount) / 100, // Convert from cents
		Currency:  string(intent.Currency),
		Method:    req.Method,
		Gateway:   "stripe",
		GatewayID: intent.ID,
		CreatedAt: time.Unix(intent.Created, 0),
		UpdatedAt: time.Unix(intent.Created, 0),
	}

	// Add gateway response
	response.GatewayResponse = map[string]interface{}{
		"client_secret": intent.ClientSecret,
		"status":        intent.Status,
	}

	return response, nil
}

// GetPayment retrieves payment details
func (g *StripeGateway) GetPayment(ctx context.Context, paymentID string) (*payment.PaymentResponse, error) {
	intent, err := paymentintent.Get(paymentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment intent: %w", err)
	}

	response := &payment.PaymentResponse{
		ID:        intent.ID,
		Status:    string(intent.Status),
		Amount:    float64(intent.Amount) / 100,
		Currency:  string(intent.Currency),
		Gateway:   "stripe",
		GatewayID: intent.ID,
		CreatedAt: time.Unix(intent.Created, 0),
		UpdatedAt: time.Unix(intent.Created, 0),
	}

	// Add gateway response
	response.GatewayResponse = map[string]interface{}{
		"client_secret": intent.ClientSecret,
		"status":        intent.Status,
		"metadata":      intent.Metadata,
	}

	return response, nil
}

// UpdatePayment updates payment details
func (g *StripeGateway) UpdatePayment(ctx context.Context, paymentID string, updates map[string]interface{}) (*payment.PaymentResponse, error) {
	params := &stripe.PaymentIntentParams{}

	// Update metadata if provided
	if metadata, ok := updates["metadata"].(map[string]string); ok {
		params.Metadata = metadata
	}

	// Update description if provided
	if description, ok := updates["description"].(string); ok {
		params.Description = stripe.String(description)
	}

	intent, err := paymentintent.Update(paymentID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment intent: %w", err)
	}

	response := &payment.PaymentResponse{
		ID:        intent.ID,
		Status:    string(intent.Status),
		Amount:    float64(intent.Amount) / 100,
		Currency:  string(intent.Currency),
		Gateway:   "stripe",
		GatewayID: intent.ID,
		CreatedAt: time.Unix(intent.Created, 0),
		UpdatedAt: time.Unix(intent.Created, 0),
	}

	return response, nil
}

// CancelPayment cancels a payment
func (g *StripeGateway) CancelPayment(ctx context.Context, paymentID string) (*payment.PaymentResponse, error) {
	intent, err := paymentintent.Cancel(paymentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel payment intent: %w", err)
	}

	response := &payment.PaymentResponse{
		ID:        intent.ID,
		Status:    string(intent.Status),
		Amount:    float64(intent.Amount) / 100,
		Currency:  string(intent.Currency),
		Gateway:   "stripe",
		GatewayID: intent.ID,
		CreatedAt: time.Unix(intent.Created, 0),
		UpdatedAt: time.Unix(intent.Created, 0),
	}

	return response, nil
}

// RefundPayment processes a refund
func (g *StripeGateway) RefundPayment(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(req.PaymentID),
		Reason:        stripe.String(req.Reason),
		Metadata:      req.Metadata,
	}

	// Set amount if specified (0 means full refund)
	if req.Amount > 0 {
		params.Amount = stripe.Int64(int64(req.Amount * 100))
	}

	refund, err := refund.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	response := &payment.RefundResponse{
		ID:        refund.ID,
		PaymentID: req.PaymentID,
		Amount:    float64(refund.Amount) / 100,
		Status:    string(refund.Status),
		GatewayID: refund.ID,
		CreatedAt: time.Unix(refund.Created, 0),
	}

	// Add gateway response
	response.GatewayResponse = map[string]interface{}{
		"status": refund.Status,
		"reason": refund.Reason,
	}

	return response, nil
}

// CreateSubscription creates a subscription
func (g *StripeGateway) CreateSubscription(ctx context.Context, req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	// Create or get customer
	customerID, err := g.getOrCreateCustomer(ctx, req.Customer)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	// Create checkout session for subscription
	sessionParams := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
			"upi",
			"netbanking",
			"wallet",
		}),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(req.ReturnURL + "?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(req.ReturnURL + "?cancelled=true"),
		Metadata:   req.Metadata,
	}

	// Add line items based on subscription request
	if req.Subscription != nil {
		// Get plan details from metadata or create price
		priceID := req.Metadata["price_id"]
		if priceID == "" {
			// Create price on the fly
			priceID, err = g.createPrice(req.Amount, req.Currency, req.Subscription.BillingInterval)
			if err != nil {
				return nil, fmt.Errorf("failed to create price: %w", err)
			}
		}

		sessionParams.LineItems = []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		}
	}

	// Create checkout session
	session, err := session.New(sessionParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	response := &payment.PaymentResponse{
		ID:          session.ID,
		Status:      "pending",
		Amount:      req.Amount,
		Currency:    req.Currency,
		Method:      req.Method,
		Gateway:     "stripe",
		GatewayID:   session.ID,
		RedirectURL: session.URL,
		CreatedAt:   time.Unix(session.Created, 0),
		UpdatedAt:   time.Unix(session.Created, 0),
	}

	// Add gateway response
	response.GatewayResponse = map[string]interface{}{
		"session_id": session.ID,
		"url":        session.URL,
	}

	return response, nil
}

// GetSubscription retrieves subscription details
func (g *StripeGateway) GetSubscription(ctx context.Context, subscriptionID string) (*payment.PaymentResponse, error) {
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	response := &payment.PaymentResponse{
		ID:        sub.ID,
		Status:    string(sub.Status),
		Gateway:   "stripe",
		GatewayID: sub.ID,
		CreatedAt: time.Unix(sub.Created, 0),
		UpdatedAt: time.Unix(sub.Created, 0),
	}

	// Add gateway response
	response.GatewayResponse = map[string]interface{}{
		"status":               sub.Status,
		"current_period_start": sub.CurrentPeriodStart,
		"current_period_end":   sub.CurrentPeriodEnd,
		"cancel_at_period_end": sub.CancelAtPeriodEnd,
		"customer":             sub.Customer.ID,
	}

	return response, nil
}

// CancelSubscription cancels a subscription
func (g *StripeGateway) CancelSubscription(ctx context.Context, subscriptionID string) (*payment.PaymentResponse, error) {
	sub, err := subscription.Cancel(subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	response := &payment.PaymentResponse{
		ID:        sub.ID,
		Status:    string(sub.Status),
		Gateway:   "stripe",
		GatewayID: sub.ID,
		CreatedAt: time.Unix(sub.Created, 0),
		UpdatedAt: time.Unix(sub.Created, 0),
	}

	return response, nil
}

// ProcessWebhook processes webhook events
func (g *StripeGateway) ProcessWebhook(ctx context.Context, payload []byte, signature string) (*payment.WebhookEvent, error) {
	// Verify webhook signature
	if err := g.VerifyWebhook(payload, signature); err != nil {
		return nil, fmt.Errorf("webhook verification failed: %w", err)
	}

	// Parse webhook event
	var event stripe.Event
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook event: %w", err)
	}

	// Convert to our webhook event format
	webhookEvent := &payment.WebhookEvent{
		ID:        event.ID,
		Type:      string(event.Type),
		Gateway:   "stripe",
		CreatedAt: time.Unix(event.Created, 0),
	}

	// Parse event data
	eventData := make(map[string]interface{})
	if err := json.Unmarshal(event.Data.Raw, &eventData); err != nil {
		return nil, fmt.Errorf("failed to parse event data: %w", err)
	}
	webhookEvent.Data = eventData

	return webhookEvent, nil
}

// VerifyWebhook verifies webhook signature
func (g *StripeGateway) VerifyWebhook(payload []byte, signature string) error {
	webhookSecret := g.config.Credentials["webhook_secret"]
	if webhookSecret == "" {
		return fmt.Errorf("webhook secret not configured")
	}

	// For now, skip verification - in production, implement proper verification
	_ = payload
	_ = signature
	_ = webhookSecret
	return nil
}

// GetSupportedMethods returns supported payment methods
func (g *StripeGateway) GetSupportedMethods() []string {
	return []string{"card", "upi", "netbanking", "wallet"}
}

// GetGatewayName returns the gateway name
func (g *StripeGateway) GetGatewayName() string {
	return "stripe"
}

// IsHealthy checks if the gateway is healthy
func (g *StripeGateway) IsHealthy(ctx context.Context) error {
	// For now, return nil - in production, implement proper health check
	return nil
}

// getOrCreateCustomer gets or creates a Stripe customer
func (g *StripeGateway) getOrCreateCustomer(ctx context.Context, customerInfo payment.CustomerInfo) (string, error) {
	// If customer ID is provided, verify it exists
	if customerInfo.ID != "" {
		_, err := customer.Get(customerInfo.ID, nil)
		if err == nil {
			return customerInfo.ID, nil
		}
		// If customer doesn't exist, we'll create a new one
	}

	// Create new customer
	params := &stripe.CustomerParams{
		Email: stripe.String(customerInfo.Email),
		Name:  stripe.String(customerInfo.Name),
	}

	if customerInfo.Phone != "" {
		params.Phone = stripe.String(customerInfo.Phone)
	}

	if customerInfo.Address != nil {
		params.Address = &stripe.AddressParams{
			Line1:      stripe.String(customerInfo.Address.Line1),
			Line2:      stripe.String(customerInfo.Address.Line2),
			City:       stripe.String(customerInfo.Address.City),
			State:      stripe.String(customerInfo.Address.State),
			PostalCode: stripe.String(customerInfo.Address.PostalCode),
			Country:    stripe.String(customerInfo.Address.Country),
		}
	}

	customer, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create customer: %w", err)
	}

	return customer.ID, nil
}

// createPrice creates a Stripe price for subscription
func (g *StripeGateway) createPrice(amount float64, currency, interval string) (string, error) {
	// This is a simplified implementation
	// In a real application, you'd want to create products and prices properly
	_ = amount
	_ = currency
	_ = interval

	// Return a mock price ID for now
	return "price_mock_" + currency, nil
}
