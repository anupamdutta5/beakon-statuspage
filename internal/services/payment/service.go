package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service implements the PaymentService interface
type Service struct {
	db                  *gorm.DB
	config              *config.PaymentConfig
	gateways            map[string]PaymentGateway
	defaultGateway      PaymentGateway
	retryService        PaymentRetryService
	notificationService PaymentNotificationService
	analyticsService    PaymentAnalyticsService
}

// NewService creates a new payment service
func NewService(paymentConfig *config.PaymentConfig) (*Service, error) {
	service := &Service{
		db:       database.DB,
		config:   paymentConfig,
		gateways: make(map[string]PaymentGateway),
	}

	// Initialize gateways
	if err := service.initializeGateways(); err != nil {
		return nil, fmt.Errorf("failed to initialize gateways: %w", err)
	}

	// Set default gateway
	defaultGateway, err := paymentConfig.GetDefaultGateway()
	if err != nil {
		return nil, fmt.Errorf("failed to get default gateway: %w", err)
	}

	service.defaultGateway = service.gateways[defaultGateway.Name]
	if service.defaultGateway == nil {
		return nil, fmt.Errorf("default gateway %s not initialized", defaultGateway.Name)
	}

	// Initialize retry service
	service.retryService = NewRetryService(service.db, paymentConfig.RetrySettings)

	// Initialize notification service
	service.notificationService = NewNotificationService(service.db)

	// Initialize analytics service
	service.analyticsService = NewAnalyticsService(service.db)

	return service, nil
}

// initializeGateways initializes all configured payment gateways
func (s *Service) initializeGateways() error {
	for _, gatewayConfig := range s.config.Gateways {
		if !gatewayConfig.IsEnabled {
			continue
		}

		var gateway PaymentGateway
		var err error

		switch gatewayConfig.Type {
		case "stripe":
			gateway, err = s.createStripeGateway(&gatewayConfig)
		case "razorpay":
			gateway, err = s.createRazorpayGateway(&gatewayConfig)
		case "payu":
			gateway, err = s.createPayUGateway(&gatewayConfig)
		case "paypal":
			gateway, err = s.createPayPalGateway(&gatewayConfig)
		default:
			logger.Warn("Unsupported gateway type", zap.String("type", gatewayConfig.Type))
			continue
		}

		if err != nil {
			logger.Error("Failed to initialize gateway",
				zap.String("gateway", gatewayConfig.Name),
				zap.String("type", gatewayConfig.Type),
				zap.Error(err))
			continue
		}

		s.gateways[gatewayConfig.Name] = gateway
		logger.Info("Gateway initialized successfully",
			zap.String("gateway", gatewayConfig.Name),
			zap.String("type", gatewayConfig.Type))
	}

	if len(s.gateways) == 0 {
		return fmt.Errorf("no payment gateways could be initialized")
	}

	return nil
}

// CreatePayment creates a payment using the appropriate gateway
func (s *Service) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// Get the appropriate gateway for the payment method
	gateway, err := s.GetGatewayForMethod(req.Method)
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway for method %s: %w", req.Method, err)
	}

	// Create payment with the gateway
	response, err := gateway.CreatePayment(ctx, req)
	if err != nil {
		logger.Error("Failed to create payment",
			zap.String("gateway", gateway.GetGatewayName()),
			zap.String("method", req.Method),
			zap.Error(err))
		return nil, err
	}

	// Record payment in database
	if err := s.RecordPayment(ctx, response); err != nil {
		logger.Error("Failed to record payment", zap.Error(err))
		// Don't fail the payment creation if recording fails
	}

	// Record analytics event
	event := &PaymentEvent{
		TenantID:  req.TenantID,
		EventType: "created",
		PaymentID: response.ID,
		Amount:    response.Amount,
		Currency:  response.Currency,
		Method:    response.Method,
		Gateway:   response.Gateway,
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{},
	}
	if err := s.analyticsService.RecordPaymentEvent(ctx, event); err != nil {
		logger.Error("Failed to record payment event", zap.Error(err))
	}

	logger.Info("Payment created successfully",
		zap.String("payment_id", response.ID),
		zap.String("gateway", response.Gateway),
		zap.String("method", response.Method),
		zap.Float64("amount", response.Amount))

	return response, nil
}

// GetPayment retrieves payment details
func (s *Service) GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	// First try to get from database
	var payment models.BillingPayment
	if err := s.db.Where("external_id = ?", paymentID).First(&payment).Error; err == nil {
		// Get gateway and fetch latest status
		gateway, exists := s.gateways[payment.PaymentMethod]
		if exists {
			response, err := gateway.GetPayment(ctx, paymentID)
			if err == nil {
				// Update database with latest status
				payment.Status = response.Status
				payment.ProcessedAt = response.UpdatedAt
				s.db.Save(&payment)
				return response, nil
			}
		}
	}

	// If not found in database or gateway fetch failed, try all gateways
	for _, gateway := range s.gateways {
		response, err := gateway.GetPayment(ctx, paymentID)
		if err == nil {
			return response, nil
		}
	}

	return nil, fmt.Errorf("payment %s not found", paymentID)
}

// UpdatePayment updates payment details
func (s *Service) UpdatePayment(ctx context.Context, paymentID string, updates map[string]interface{}) (*PaymentResponse, error) {
	// Get payment to determine gateway
	payment, err := s.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	gateway, exists := s.gateways[payment.Gateway]
	if !exists {
		return nil, fmt.Errorf("gateway %s not found", payment.Gateway)
	}

	response, err := gateway.UpdatePayment(ctx, paymentID, updates)
	if err != nil {
		return nil, err
	}

	// Update database
	if err := s.RecordPayment(ctx, response); err != nil {
		logger.Error("Failed to update payment record", zap.Error(err))
	}

	return response, nil
}

// CancelPayment cancels a payment
func (s *Service) CancelPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	// Get payment to determine gateway
	payment, err := s.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	gateway, exists := s.gateways[payment.Gateway]
	if !exists {
		return nil, fmt.Errorf("gateway %s not found", payment.Gateway)
	}

	response, err := gateway.CancelPayment(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	// Update database
	if err := s.RecordPayment(ctx, response); err != nil {
		logger.Error("Failed to update payment record", zap.Error(err))
	}

	// Record analytics event
	paymentEvent := &PaymentEvent{
		TenantID:  0, // Will be extracted from payment data
		EventType: "cancelled",
		PaymentID: response.ID,
		Amount:    response.Amount,
		Currency:  response.Currency,
		Method:    response.Method,
		Gateway:   response.Gateway,
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{},
	}
	if err := s.analyticsService.RecordPaymentEvent(ctx, paymentEvent); err != nil {
		logger.Error("Failed to record payment event", zap.Error(err))
	}

	return response, nil
}

// RefundPayment processes a refund
func (s *Service) RefundPayment(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	// Get payment to determine gateway
	payment, err := s.GetPayment(ctx, req.PaymentID)
	if err != nil {
		return nil, err
	}

	gateway, exists := s.gateways[payment.Gateway]
	if !exists {
		return nil, fmt.Errorf("gateway %s not found", payment.Gateway)
	}

	response, err := gateway.RefundPayment(ctx, req)
	if err != nil {
		return nil, err
	}

	// Record refund in database
	if err := s.RecordRefund(ctx, response); err != nil {
		logger.Error("Failed to record refund", zap.Error(err))
	}

	// Send notification
	if err := s.notificationService.SendRefundNotification(ctx, response); err != nil {
		logger.Error("Failed to send refund notification", zap.Error(err))
	}

	logger.Info("Refund processed successfully",
		zap.String("refund_id", response.ID),
		zap.String("payment_id", response.PaymentID),
		zap.Float64("amount", response.Amount))

	return response, nil
}

// CreateSubscription creates a subscription
func (s *Service) CreateSubscription(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// Use default gateway for subscriptions
	response, err := s.defaultGateway.CreateSubscription(ctx, req)
	if err != nil {
		logger.Error("Failed to create subscription",
			zap.String("gateway", s.defaultGateway.GetGatewayName()),
			zap.Error(err))
		return nil, err
	}

	// Record payment in database
	if err := s.RecordPayment(ctx, response); err != nil {
		logger.Error("Failed to record subscription", zap.Error(err))
	}

	// Send notification
	if err := s.notificationService.SendSubscriptionNotification(ctx, response, "created"); err != nil {
		logger.Error("Failed to send subscription notification", zap.Error(err))
	}

	logger.Info("Subscription created successfully",
		zap.String("subscription_id", response.ID),
		zap.String("gateway", response.Gateway))

	return response, nil
}

// GetSubscription retrieves subscription details
func (s *Service) GetSubscription(ctx context.Context, subscriptionID string) (*PaymentResponse, error) {
	return s.defaultGateway.GetSubscription(ctx, subscriptionID)
}

// CancelSubscription cancels a subscription
func (s *Service) CancelSubscription(ctx context.Context, subscriptionID string) (*PaymentResponse, error) {
	response, err := s.defaultGateway.CancelSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}

	// Update database
	if err := s.RecordPayment(ctx, response); err != nil {
		logger.Error("Failed to update subscription record", zap.Error(err))
	}

	// Send notification
	if err := s.notificationService.SendSubscriptionNotification(ctx, response, "cancelled"); err != nil {
		logger.Error("Failed to send subscription notification", zap.Error(err))
	}

	return response, nil
}

// ProcessWebhook processes webhook events
func (s *Service) ProcessWebhook(ctx context.Context, gateway string, payload []byte, signature string) (*WebhookEvent, error) {
	gatewayImpl, exists := s.gateways[gateway]
	if !exists {
		return nil, fmt.Errorf("gateway %s not found", gateway)
	}

	event, err := gatewayImpl.ProcessWebhook(ctx, payload, signature)
	if err != nil {
		return nil, err
	}

	// Process the webhook event
	if err := s.processWebhookEvent(ctx, event); err != nil {
		logger.Error("Failed to process webhook event",
			zap.String("event_id", event.ID),
			zap.String("event_type", event.Type),
			zap.Error(err))
		return nil, err
	}

	return event, nil
}

// processWebhookEvent processes a webhook event
func (s *Service) processWebhookEvent(ctx context.Context, event *WebhookEvent) error {
	switch event.Type {
	case "payment.completed":
		return s.handlePaymentCompleted(ctx, event)
	case "payment.failed":
		return s.handlePaymentFailed(ctx, event)
	case "subscription.created":
		return s.handleSubscriptionCreated(ctx, event)
	case "subscription.updated":
		return s.handleSubscriptionUpdated(ctx, event)
	case "subscription.cancelled":
		return s.handleSubscriptionCancelled(ctx, event)
	case "refund.processed":
		return s.handleRefundProcessed(ctx, event)
	default:
		logger.Info("Unhandled webhook event type", zap.String("type", event.Type))
		return nil
	}
}

// handlePaymentCompleted handles payment completed webhook
func (s *Service) handlePaymentCompleted(ctx context.Context, event *WebhookEvent) error {
	// Extract payment information from event data
	paymentID, ok := event.Data["id"].(string)
	if !ok {
		return fmt.Errorf("payment ID not found in webhook data")
	}

	// Get payment details
	payment, err := s.GetPayment(ctx, paymentID)
	if err != nil {
		return err
	}

	// Update payment status
	payment.Status = "completed"
	if err := s.RecordPayment(ctx, payment); err != nil {
		return err
	}

	// Send success notification
	if err := s.notificationService.SendPaymentSuccessNotification(ctx, payment); err != nil {
		logger.Error("Failed to send payment success notification", zap.Error(err))
	}

	// Record analytics event
	paymentEvent := &PaymentEvent{
		TenantID:  0, // Will be extracted from payment data
		EventType: "completed",
		PaymentID: payment.ID,
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		Method:    payment.Method,
		Gateway:   payment.Gateway,
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{},
	}
	if err := s.analyticsService.RecordPaymentEvent(ctx, paymentEvent); err != nil {
		logger.Error("Failed to record payment event", zap.Error(err))
	}

	return nil
}

// handlePaymentFailed handles payment failed webhook
func (s *Service) handlePaymentFailed(ctx context.Context, event *WebhookEvent) error {
	// Extract payment information from event data
	paymentID, ok := event.Data["id"].(string)
	if !ok {
		return fmt.Errorf("payment ID not found in webhook data")
	}

	// Get payment details
	payment, err := s.GetPayment(ctx, paymentID)
	if err != nil {
		return err
	}

	// Update payment status
	payment.Status = "failed"
	if err := s.RecordPayment(ctx, payment); err != nil {
		return err
	}

	// Extract failure reason
	reason := "Payment failed"
	if failureReason, ok := event.Data["failure_reason"].(string); ok {
		reason = failureReason
	}

	// Send failure notification
	if err := s.notificationService.SendPaymentFailureNotification(ctx, payment, reason); err != nil {
		logger.Error("Failed to send payment failure notification", zap.Error(err))
	}

	// Check if payment should be retried
	shouldRetry, delay, err := s.retryService.ShouldRetry(ctx, paymentID, "payment_failed")
	if err != nil {
		logger.Error("Failed to check retry eligibility", zap.Error(err))
	} else if shouldRetry {
		if err := s.retryService.ScheduleRetry(ctx, paymentID, delay); err != nil {
			logger.Error("Failed to schedule payment retry", zap.Error(err))
		}
	}

	// Record analytics event
	paymentEvent := &PaymentEvent{
		TenantID:  0, // Will be extracted from payment data
		EventType: "failed",
		PaymentID: payment.ID,
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		Method:    payment.Method,
		Gateway:   payment.Gateway,
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{},
	}
	if err := s.analyticsService.RecordPaymentEvent(ctx, paymentEvent); err != nil {
		logger.Error("Failed to record payment event", zap.Error(err))
	}

	return nil
}

// handleSubscriptionCreated handles subscription created webhook
func (s *Service) handleSubscriptionCreated(ctx context.Context, event *WebhookEvent) error {
	// Extract subscription information from event data
	subscriptionID, ok := event.Data["id"].(string)
	if !ok {
		return fmt.Errorf("subscription ID not found in webhook data")
	}

	// Get subscription details
	subscription, err := s.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	// Record subscription
	if err := s.RecordPayment(ctx, subscription); err != nil {
		return err
	}

	// Send notification
	if err := s.notificationService.SendSubscriptionNotification(ctx, subscription, "created"); err != nil {
		logger.Error("Failed to send subscription notification", zap.Error(err))
	}

	return nil
}

// handleSubscriptionUpdated handles subscription updated webhook
func (s *Service) handleSubscriptionUpdated(ctx context.Context, event *WebhookEvent) error {
	// Extract subscription information from event data
	subscriptionID, ok := event.Data["id"].(string)
	if !ok {
		return fmt.Errorf("subscription ID not found in webhook data")
	}

	// Get subscription details
	subscription, err := s.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	// Update subscription
	if err := s.RecordPayment(ctx, subscription); err != nil {
		return err
	}

	// Send notification
	if err := s.notificationService.SendSubscriptionNotification(ctx, subscription, "updated"); err != nil {
		logger.Error("Failed to send subscription notification", zap.Error(err))
	}

	return nil
}

// handleSubscriptionCancelled handles subscription cancelled webhook
func (s *Service) handleSubscriptionCancelled(ctx context.Context, event *WebhookEvent) error {
	// Extract subscription information from event data
	subscriptionID, ok := event.Data["id"].(string)
	if !ok {
		return fmt.Errorf("subscription ID not found in webhook data")
	}

	// Get subscription details
	subscription, err := s.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	// Update subscription
	subscription.Status = "cancelled"
	if err := s.RecordPayment(ctx, subscription); err != nil {
		return err
	}

	// Send notification
	if err := s.notificationService.SendSubscriptionNotification(ctx, subscription, "cancelled"); err != nil {
		logger.Error("Failed to send subscription notification", zap.Error(err))
	}

	return nil
}

// handleRefundProcessed handles refund processed webhook
func (s *Service) handleRefundProcessed(ctx context.Context, event *WebhookEvent) error {
	// Extract refund information from event data
	refundID, ok := event.Data["id"].(string)
	if !ok {
		return fmt.Errorf("refund ID not found in webhook data")
	}

	// Create refund response from event data
	refund := &RefundResponse{
		ID:        refundID,
		Status:    "processed",
		CreatedAt: time.Now(),
	}

	// Record refund
	if err := s.RecordRefund(ctx, refund); err != nil {
		return err
	}

	// Send notification
	if err := s.notificationService.SendRefundNotification(ctx, refund); err != nil {
		logger.Error("Failed to send refund notification", zap.Error(err))
	}

	return nil
}

// GetSupportedMethods returns all supported payment methods
func (s *Service) GetSupportedMethods() []string {
	return s.config.GetSupportedMethods()
}

// GetGateways returns available gateways
func (s *Service) GetGateways() []string {
	var gateways []string
	for name := range s.gateways {
		gateways = append(gateways, name)
	}
	return gateways
}

// GetGatewayForMethod returns the best gateway for a payment method
func (s *Service) GetGatewayForMethod(method string) (PaymentGateway, error) {
	gatewayConfig, err := s.config.GetGatewayForMethod(method)
	if err != nil {
		return nil, err
	}

	gateway, exists := s.gateways[gatewayConfig.Name]
	if !exists {
		return nil, fmt.Errorf("gateway %s not initialized", gatewayConfig.Name)
	}

	return gateway, nil
}

// RecordPayment records a payment in the database
func (s *Service) RecordPayment(ctx context.Context, response *PaymentResponse) error {
	// Check if payment already exists
	var existingPayment models.BillingPayment
	if err := s.db.Where("external_id = ?", response.ID).First(&existingPayment).Error; err == nil {
		// Update existing payment
		existingPayment.Status = response.Status
		existingPayment.ProcessedAt = response.UpdatedAt
		existingPayment.Amount = response.Amount
		existingPayment.PaymentMethod = response.Gateway
		return s.db.Save(&existingPayment).Error
	}

	// Create new payment record
	payment := &models.BillingPayment{
		ExternalID:    response.ID,
		Amount:        response.Amount,
		PaymentMethod: response.Gateway,
		Status:        response.Status,
		ProcessedAt:   response.UpdatedAt,
		CreatedAt:     response.CreatedAt,
	}

	// Extract tenant ID from metadata if available
	// This would need proper conversion logic
	_ = response.GatewayResponse["tenant_id"]

	return s.db.Create(payment).Error
}

// RecordRefund records a refund in the database
func (s *Service) RecordRefund(ctx context.Context, response *RefundResponse) error {
	// Find the original payment
	var payment models.BillingPayment
	if err := s.db.Where("external_id = ?", response.PaymentID).First(&payment).Error; err != nil {
		return fmt.Errorf("original payment not found: %w", err)
	}

	// Create refund record
	refund := &models.BillingRefund{
		PaymentID:   payment.ID,
		Amount:      response.Amount,
		Status:      response.Status,
		ExternalID:  response.ID,
		ProcessedAt: response.CreatedAt,
		CreatedAt:   response.CreatedAt,
	}

	return s.db.Create(refund).Error
}

// GetPaymentHistory returns payment history for a tenant
func (s *Service) GetPaymentHistory(ctx context.Context, tenantID uint, limit, offset int) ([]*models.BillingPayment, error) {
	var payments []*models.BillingPayment
	query := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&payments).Error; err != nil {
		return nil, err
	}

	return payments, nil
}

// GetInvoiceHistory returns invoice history for a tenant
func (s *Service) GetInvoiceHistory(ctx context.Context, tenantID uint, limit, offset int) ([]*models.BillingInvoice, error) {
	var invoices []*models.BillingInvoice
	query := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&invoices).Error; err != nil {
		return nil, err
	}

	return invoices, nil
}

// GetPaymentMetrics returns payment metrics for a tenant
func (s *Service) GetPaymentMetrics(ctx context.Context, tenantID uint, period string) (*PaymentMetrics, error) {
	return s.analyticsService.GetPaymentMetrics(ctx, tenantID, period)
}

// GetConversionFunnel returns conversion funnel data for a tenant
func (s *Service) GetConversionFunnel(ctx context.Context, tenantID uint, period string) (*ConversionFunnel, error) {
	return s.analyticsService.GetConversionFunnel(ctx, tenantID, period)
}

// Gateway factory methods (to avoid import cycles)

// createStripeGateway creates a new Stripe gateway
func (s *Service) createStripeGateway(gatewayConfig *config.PaymentGateway) (PaymentGateway, error) {
	// This is a placeholder implementation
	// The actual implementation would be in gateways/stripe.go
	return &MockGateway{name: "stripe"}, nil
}

// createRazorpayGateway creates a new Razorpay gateway
func (s *Service) createRazorpayGateway(gatewayConfig *config.PaymentGateway) (PaymentGateway, error) {
	// This is a placeholder implementation
	// The actual implementation would be in gateways/razorpay.go
	return &MockGateway{name: "razorpay"}, nil
}

// createPayUGateway creates a new PayU gateway
func (s *Service) createPayUGateway(gatewayConfig *config.PaymentGateway) (PaymentGateway, error) {
	// This is a placeholder implementation
	// The actual implementation would be in gateways/payu.go
	return &MockGateway{name: "payu"}, nil
}

// createPayPalGateway creates a new PayPal gateway
func (s *Service) createPayPalGateway(gatewayConfig *config.PaymentGateway) (PaymentGateway, error) {
	// This is a placeholder implementation
	// The actual implementation would be in gateways/paypal.go
	return &MockGateway{name: "paypal"}, nil
}

// MockGateway is a mock implementation for testing
type MockGateway struct {
	name string
}

func (g *MockGateway) Initialize(config map[string]string) error {
	return nil
}

func (g *MockGateway) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	return &PaymentResponse{
		ID:        "mock_payment_id",
		Status:    "pending",
		Amount:    req.Amount,
		Currency:  req.Currency,
		Method:    req.Method,
		Gateway:   g.name,
		GatewayID: "mock_payment_id",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (g *MockGateway) GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	return &PaymentResponse{
		ID:        paymentID,
		Status:    "completed",
		Gateway:   g.name,
		GatewayID: paymentID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (g *MockGateway) UpdatePayment(ctx context.Context, paymentID string, updates map[string]interface{}) (*PaymentResponse, error) {
	return g.GetPayment(ctx, paymentID)
}

func (g *MockGateway) CancelPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	return &PaymentResponse{
		ID:        paymentID,
		Status:    "cancelled",
		Gateway:   g.name,
		GatewayID: paymentID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (g *MockGateway) RefundPayment(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	return &RefundResponse{
		ID:        "mock_refund_id",
		PaymentID: req.PaymentID,
		Amount:    req.Amount,
		Status:    "processed",
		GatewayID: "mock_refund_id",
		CreatedAt: time.Now(),
	}, nil
}

func (g *MockGateway) CreateSubscription(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	return g.CreatePayment(ctx, req)
}

func (g *MockGateway) GetSubscription(ctx context.Context, subscriptionID string) (*PaymentResponse, error) {
	return g.GetPayment(ctx, subscriptionID)
}

func (g *MockGateway) CancelSubscription(ctx context.Context, subscriptionID string) (*PaymentResponse, error) {
	return g.CancelPayment(ctx, subscriptionID)
}

func (g *MockGateway) ProcessWebhook(ctx context.Context, payload []byte, signature string) (*WebhookEvent, error) {
	return &WebhookEvent{
		ID:        "mock_webhook_id",
		Type:      "payment.completed",
		Gateway:   g.name,
		Data:      map[string]interface{}{"payment_id": "mock_payment_id"},
		CreatedAt: time.Now(),
	}, nil
}

func (g *MockGateway) VerifyWebhook(payload []byte, signature string) error {
	return nil
}

func (g *MockGateway) GetSupportedMethods() []string {
	return []string{"card", "upi", "netbanking", "wallet"}
}

func (g *MockGateway) GetGatewayName() string {
	return g.name
}

func (g *MockGateway) IsHealthy(ctx context.Context) error {
	return nil
}
