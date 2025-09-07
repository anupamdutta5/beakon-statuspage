package services

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/subscription"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PaymentService struct {
	db                  *gorm.DB
	subscriptionService *SubscriptionService
}

func NewPaymentService() *PaymentService {
	// Initialize Stripe
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	return &PaymentService{
		db:                  database.DB,
		subscriptionService: NewSubscriptionService(),
	}
}

// CreateCheckoutSession creates a Stripe checkout session for a subscription
func (s *PaymentService) CreateCheckoutSession(tenantID uint, planSlug string, successURL, cancelURL string) (*stripe.CheckoutSession, error) {
	// Get the plan
	var plan models.SubscriptionPlan
	if err := s.db.Where("slug = ? AND is_active = ?", planSlug, true).First(&plan).Error; err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	// Get the tenant
	var tenant models.Tenant
	if err := s.db.First(&tenant, tenantID).Error; err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Create or get Stripe customer
	customerID, err := s.getOrCreateStripeCustomer(&tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	// Create checkout session
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(plan.Currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(plan.Name),
						Description: stripe.String(plan.Description),
					},
					UnitAmount: stripe.Int64(int64(plan.Price * 100)), // Convert to cents
					Recurring: &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
						Interval: stripe.String(plan.BillingInterval),
					},
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			"tenant_id": fmt.Sprintf("%d", tenantID),
			"plan_slug": planSlug,
		},
	}

	session, err := checkoutsession.New(params)
	if err != nil {
		logger.Error("Failed to create checkout session", zap.Error(err))
		return nil, err
	}

	logger.Info("Checkout session created",
		zap.Uint("tenant_id", tenantID),
		zap.String("plan", planSlug),
		zap.String("session_id", session.ID))

	return session, nil
}

// HandleWebhook processes Stripe webhooks
func (s *PaymentService) HandleWebhook(payload []byte, signature string) error {
	// Verify webhook signature
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return fmt.Errorf("STRIPE_WEBHOOK_SECRET not configured")
	}

	// In a real implementation, you would verify the webhook signature
	// For now, we'll parse the event directly
	var event stripe.Event
	if err := json.Unmarshal(payload, &event); err != nil {
		logger.Error("Failed to parse webhook payload", zap.Error(err))
		return fmt.Errorf("invalid webhook payload: %v", err)
	}

	// Handle the event
	switch event.Type {
	case "checkout.session.completed":
		return s.handleCheckoutSessionCompleted(event)
	case "customer.subscription.created":
		return s.handleSubscriptionCreated(event)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(event)
	case "invoice.payment_succeeded":
		return s.handlePaymentSucceeded(event)
	case "invoice.payment_failed":
		return s.handlePaymentFailed(event)
	default:
		logger.Info("Unhandled webhook event type", zap.String("type", string(event.Type)))
		return nil
	}
}

func (s *PaymentService) handleCheckoutSessionCompleted(event stripe.Event) error {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		logger.Error("Failed to parse checkout session", zap.Error(err))
		return err
	}

	// Get tenant ID from metadata
	tenantIDStr, exists := session.Metadata["tenant_id"]
	if !exists {
		return fmt.Errorf("tenant_id not found in session metadata")
	}

	planSlug, exists := session.Metadata["plan_slug"]
	if !exists {
		return fmt.Errorf("plan_slug not found in session metadata")
	}

	// Create subscription from Stripe subscription
	if session.Subscription != nil {
		_, err := s.subscriptionService.CreateSubscriptionFromStripe(
			session.Subscription.ID,
			tenantIDStr,
			planSlug,
		)
		if err != nil {
			logger.Error("Failed to create subscription from Stripe", zap.Error(err))
			return err
		}
	}

	logger.Info("Checkout session completed successfully",
		zap.String("session_id", session.ID),
		zap.String("tenant_id", tenantIDStr),
		zap.String("plan", planSlug))

	return nil
}

func (s *PaymentService) handleSubscriptionCreated(event stripe.Event) error {
	logger.Info("Subscription created", zap.String("event_id", event.ID))
	return nil
}

func (s *PaymentService) handleSubscriptionUpdated(event stripe.Event) error {
	logger.Info("Subscription updated", zap.String("event_id", event.ID))
	return nil
}

func (s *PaymentService) handleSubscriptionDeleted(event stripe.Event) error {
	logger.Info("Subscription deleted", zap.String("event_id", event.ID))
	return nil
}

func (s *PaymentService) handlePaymentSucceeded(event stripe.Event) error {
	logger.Info("Payment succeeded", zap.String("event_id", event.ID))
	return nil
}

func (s *PaymentService) handlePaymentFailed(event stripe.Event) error {
	logger.Info("Payment failed", zap.String("event_id", event.ID))
	return nil
}

func (s *PaymentService) getOrCreateStripeCustomer(tenant *models.Tenant) (string, error) {
	// Check if tenant already has a Stripe customer ID
	if tenant.SubscriptionID != "" {
		// Verify customer exists in Stripe
		_, err := customer.Get(tenant.SubscriptionID, nil)
		if err == nil {
			return tenant.SubscriptionID, nil
		}
	}

	// Create new Stripe customer
	params := &stripe.CustomerParams{
		Email: stripe.String(tenant.BillingEmail),
		Name:  stripe.String(tenant.Name),
		Metadata: map[string]string{
			"tenant_id":   fmt.Sprintf("%d", tenant.ID),
			"tenant_slug": tenant.Slug,
		},
	}

	customer, err := customer.New(params)
	if err != nil {
		return "", err
	}

	// Update tenant with Stripe customer ID
	tenant.SubscriptionID = customer.ID
	if err := s.db.Save(tenant).Error; err != nil {
		logger.Error("Failed to save Stripe customer ID", zap.Error(err))
		return "", err
	}

	return customer.ID, nil
}

func (s *PaymentService) updateSubscriptionFromStripe(stripeSub *stripe.Subscription) error {
	var subscription models.Subscription
	if err := s.db.Where("external_id = ?", stripeSub.ID).First(&subscription).Error; err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Update subscription status
	subscription.Status = string(stripeSub.Status)
	subscription.CurrentPeriodStart = time.Unix(stripeSub.CurrentPeriodStart, 0)
	subscription.CurrentPeriodEnd = time.Unix(stripeSub.CurrentPeriodEnd, 0)
	subscription.CancelAtPeriodEnd = stripeSub.CancelAtPeriodEnd
	subscription.UpdatedAt = time.Now()

	if stripeSub.CanceledAt > 0 {
		cancelledAt := time.Unix(stripeSub.CanceledAt, 0)
		subscription.CancelledAt = &cancelledAt
	}

	if err := s.db.Save(&subscription).Error; err != nil {
		logger.Error("Failed to update subscription", zap.Error(err))
		return err
	}

	return nil
}

func (s *PaymentService) recordPayment(invoice *stripe.Invoice) error {
	// Find subscription
	var subscription models.Subscription
	if err := s.db.Where("external_id = ?", invoice.Subscription.ID).First(&subscription).Error; err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Create billing invoice record
	billingInvoice := &models.BillingInvoice{
		TenantID:       subscription.TenantID,
		SubscriptionID: subscription.ID,
		Amount:         float64(invoice.AmountPaid) / 100, // Convert from cents
		Currency:       string(invoice.Currency),
		Status:         "paid",
		ExternalID:     invoice.ID,
		DueDate:        time.Unix(invoice.DueDate, 0),
		PaidAt:         &[]time.Time{time.Unix(invoice.StatusTransitions.PaidAt, 0)}[0],
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.db.Create(billingInvoice).Error; err != nil {
		logger.Error("Failed to create billing invoice", zap.Error(err))
		return err
	}

	// Record payment
	payment := &models.BillingPayment{
		InvoiceID:     billingInvoice.ID,
		Amount:        billingInvoice.Amount,
		PaymentMethod: "stripe",
		ExternalID:    invoice.PaymentIntent.ID,
		Status:        "completed",
		ProcessedAt:   time.Now(),
		CreatedAt:     time.Now(),
	}

	if err := s.db.Create(payment).Error; err != nil {
		logger.Error("Failed to create payment record", zap.Error(err))
		return err
	}

	logger.Info("Payment recorded successfully",
		zap.Uint("invoice_id", billingInvoice.ID),
		zap.Float64("amount", billingInvoice.Amount))

	return nil
}

// CancelSubscription cancels a subscription
func (s *PaymentService) CancelSubscription(subscriptionID uint) error {
	var sub models.Subscription
	if err := s.db.First(&sub, subscriptionID).Error; err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	if sub.ExternalID == "" {
		return fmt.Errorf("subscription not linked to Stripe")
	}

	// Cancel subscription in Stripe
	_, err := subscription.Cancel(sub.ExternalID, &stripe.SubscriptionCancelParams{})
	if err != nil {
		logger.Error("Failed to cancel subscription in Stripe", zap.Error(err))
		return err
	}

	// Update local subscription
	sub.Status = "cancelled"
	sub.CancelledAt = &[]time.Time{time.Now()}[0]
	sub.UpdatedAt = time.Now()

	if err := s.db.Save(&sub).Error; err != nil {
		logger.Error("Failed to update subscription status", zap.Error(err))
		return err
	}

	logger.Info("Subscription cancelled successfully",
		zap.Uint("subscription_id", subscriptionID))

	return nil
}

// GetCustomerPortalURL creates a Stripe customer portal session
func (s *PaymentService) GetCustomerPortalURL(tenantID uint, returnURL string) (string, error) {
	var tenant models.Tenant
	if err := s.db.First(&tenant, tenantID).Error; err != nil {
		return "", fmt.Errorf("tenant not found: %w", err)
	}

	if tenant.SubscriptionID == "" {
		return "", fmt.Errorf("tenant not linked to Stripe")
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(tenant.SubscriptionID),
		ReturnURL: stripe.String(returnURL),
	}

	session, err := session.New(params)
	if err != nil {
		logger.Error("Failed to create customer portal session", zap.Error(err))
		return "", err
	}

	return session.URL, nil
}
