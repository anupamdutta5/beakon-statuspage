// Package services provides business logic for the Payment Service.
package services

import (
	"fmt"
	"time"

	"github.com/anupamdutta5/payment-service/internal/config"
	"github.com/anupamdutta5/payment-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PaymentService handles payment-related business logic.
type PaymentService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPaymentService creates a new payment service.
func NewPaymentService(db *gorm.DB, logger *zap.Logger) *PaymentService {
	return &PaymentService{
		db:     db,
		logger: logger,
	}
}

// CreatePayment creates a new payment.
func (s *PaymentService) CreatePayment(payment *models.Payment) error {
	// Set default values
	if payment.Currency == "" {
		payment.Currency = "USD"
	}
	if payment.Status == "" {
		payment.Status = "pending"
	}

	// Create payment
	if err := s.db.Create(payment).Error; err != nil {
		s.logger.Error("Failed to create payment", zap.Error(err))
		return fmt.Errorf("failed to create payment: %w", err)
	}

	// Create initial transaction
	transaction := &models.PaymentTransaction{
		PaymentID:   payment.ID,
		Type:        "charge",
		Amount:      payment.Amount,
		Currency:    payment.Currency,
		Status:      "pending",
		Description: payment.Description,
	}

	if err := s.db.Create(transaction).Error; err != nil {
		s.logger.Error("Failed to create payment transaction", zap.Error(err))
		// Don't fail payment creation if transaction creation fails
	}

	s.logger.Info("Payment created successfully", zap.Uint("payment_id", payment.ID))
	return nil
}

// GetPayment retrieves a payment by ID.
func (s *PaymentService) GetPayment(id uint) (*models.Payment, error) {
	var payment models.Payment
	if err := s.db.Preload("Transactions").First(&payment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment not found")
		}
		s.logger.Error("Failed to get payment", zap.Error(err))
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return &payment, nil
}

// GetPayments retrieves a list of payments with pagination.
func (s *PaymentService) GetPayments(tenantID uint, limit, offset int) ([]*models.Payment, int64, error) {
	var payments []*models.Payment
	var total int64

	// Get total count
	if err := s.db.Model(&models.Payment{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count payments", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count payments: %w", err)
	}

	// Get payments with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).Preload("Transactions").Limit(limit).Offset(offset).Order("created_at DESC").Find(&payments).Error; err != nil {
		s.logger.Error("Failed to get payments", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get payments: %w", err)
	}

	return payments, total, nil
}

// UpdatePayment updates a payment.
func (s *PaymentService) UpdatePayment(payment *models.Payment) error {
	if err := s.db.Save(payment).Error; err != nil {
		s.logger.Error("Failed to update payment", zap.Error(err))
		return fmt.Errorf("failed to update payment: %w", err)
	}

	s.logger.Info("Payment updated successfully", zap.Uint("payment_id", payment.ID))
	return nil
}

// RefundPayment processes a refund for a payment.
func (s *PaymentService) RefundPayment(paymentID uint, amount float64, reason string) error {
	// Get payment
	var payment models.Payment
	if err := s.db.First(&payment, paymentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("payment not found")
		}
		s.logger.Error("Failed to get payment", zap.Error(err))
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Check if payment can be refunded
	if payment.Status != "completed" {
		return fmt.Errorf("payment cannot be refunded, current status: %s", payment.Status)
	}

	// Calculate refund amount
	refundAmount := amount
	if refundAmount == 0 {
		refundAmount = payment.Amount - payment.RefundAmount
	}

	if refundAmount <= 0 {
		return fmt.Errorf("no amount available for refund")
	}

	if payment.RefundAmount+refundAmount > payment.Amount {
		return fmt.Errorf("refund amount exceeds available amount")
	}

	// Create refund transaction
	transaction := &models.PaymentTransaction{
		PaymentID:   paymentID,
		Type:        "refund",
		Amount:      refundAmount,
		Currency:    payment.Currency,
		Status:      "pending",
		Description: fmt.Sprintf("Refund: %s", reason),
	}

	if err := s.db.Create(transaction).Error; err != nil {
		s.logger.Error("Failed to create refund transaction", zap.Error(err))
		return fmt.Errorf("failed to create refund transaction: %w", err)
	}

	// Update payment
	payment.RefundAmount += refundAmount
	payment.RefundReason = reason
	if payment.RefundAmount >= payment.Amount {
		payment.Status = "refunded"
		now := time.Now()
		payment.RefundedAt = &now
	}

	if err := s.db.Save(&payment).Error; err != nil {
		s.logger.Error("Failed to update payment for refund", zap.Error(err))
		return fmt.Errorf("failed to update payment for refund: %w", err)
	}

	s.logger.Info("Payment refunded successfully",
		zap.Uint("payment_id", paymentID),
		zap.Float64("refund_amount", refundAmount))
	return nil
}

// GetPaymentTransactions retrieves transactions for a payment.
func (s *PaymentService) GetPaymentTransactions(paymentID uint) ([]*models.PaymentTransaction, error) {
	var transactions []*models.PaymentTransaction
	if err := s.db.Where("payment_id = ?", paymentID).Order("created_at DESC").Find(&transactions).Error; err != nil {
		s.logger.Error("Failed to get payment transactions", zap.Error(err))
		return nil, fmt.Errorf("failed to get payment transactions: %w", err)
	}

	return transactions, nil
}

// Subscription Management

// CreateSubscription creates a new subscription.
func (s *PaymentService) CreateSubscription(subscription *models.Subscription) error {
	// Set default values
	if subscription.Currency == "" {
		subscription.Currency = "USD"
	}
	if subscription.Status == "" {
		subscription.Status = "active"
	}
	if subscription.StartedAt.IsZero() {
		subscription.StartedAt = time.Now()
	}

	// Calculate billing dates
	now := time.Now()
	subscription.CurrentPeriodStart = now
	subscription.NextBillingDate = now.AddDate(0, 1, 0) // Default to monthly
	if subscription.BillingCycle == "yearly" {
		subscription.NextBillingDate = now.AddDate(1, 0, 0)
	}
	subscription.CurrentPeriodEnd = subscription.NextBillingDate

	// Create subscription
	if err := s.db.Create(subscription).Error; err != nil {
		s.logger.Error("Failed to create subscription", zap.Error(err))
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	s.logger.Info("Subscription created successfully", zap.Uint("subscription_id", subscription.ID))
	return nil
}

// GetSubscription retrieves a subscription by ID.
func (s *PaymentService) GetSubscription(id uint) (*models.Subscription, error) {
	var subscription models.Subscription
	if err := s.db.Preload("Plan").Preload("Payments").First(&subscription, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription not found")
		}
		s.logger.Error("Failed to get subscription", zap.Error(err))
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return &subscription, nil
}

// GetSubscriptions retrieves a list of subscriptions with pagination.
func (s *PaymentService) GetSubscriptions(tenantID uint, limit, offset int) ([]*models.Subscription, int64, error) {
	var subscriptions []*models.Subscription
	var total int64

	// Get total count
	if err := s.db.Model(&models.Subscription{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count subscriptions", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	// Get subscriptions with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).Preload("Plan").Preload("Payments").Limit(limit).Offset(offset).Order("created_at DESC").Find(&subscriptions).Error; err != nil {
		s.logger.Error("Failed to get subscriptions", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	return subscriptions, total, nil
}

// UpdateSubscription updates a subscription.
func (s *PaymentService) UpdateSubscription(subscription *models.Subscription) error {
	if err := s.db.Save(subscription).Error; err != nil {
		s.logger.Error("Failed to update subscription", zap.Error(err))
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	s.logger.Info("Subscription updated successfully", zap.Uint("subscription_id", subscription.ID))
	return nil
}

// CancelSubscription cancels a subscription.
func (s *PaymentService) CancelSubscription(subscriptionID uint, reason string) error {
	var subscription models.Subscription
	if err := s.db.First(&subscription, subscriptionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("subscription not found")
		}
		s.logger.Error("Failed to get subscription", zap.Error(err))
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	subscription.Status = "cancelled"
	subscription.CancelReason = reason
	now := time.Now()
	subscription.CancelledAt = &now

	if err := s.db.Save(&subscription).Error; err != nil {
		s.logger.Error("Failed to cancel subscription", zap.Error(err))
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	s.logger.Info("Subscription cancelled successfully", zap.Uint("subscription_id", subscriptionID))
	return nil
}

// Plan Management

// CreatePlan creates a new plan.
func (s *PaymentService) CreatePlan(plan *models.Plan) error {
	// Set default values
	if plan.Currency == "" {
		plan.Currency = "USD"
	}
	if plan.IsActive == false && plan.IsActive != true {
		plan.IsActive = true
	}
	if plan.IsPublic == false && plan.IsPublic != true {
		plan.IsPublic = true
	}

	// Create plan
	if err := s.db.Create(plan).Error; err != nil {
		s.logger.Error("Failed to create plan", zap.Error(err))
		return fmt.Errorf("failed to create plan: %w", err)
	}

	s.logger.Info("Plan created successfully", zap.Uint("plan_id", plan.ID))
	return nil
}

// GetPlan retrieves a plan by ID.
func (s *PaymentService) GetPlan(id uint) (*models.Plan, error) {
	var plan models.Plan
	if err := s.db.First(&plan, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("plan not found")
		}
		s.logger.Error("Failed to get plan", zap.Error(err))
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	return &plan, nil
}

// GetPlans retrieves a list of plans.
func (s *PaymentService) GetPlans(tenantID uint, limit, offset int) ([]*models.Plan, int64, error) {
	var plans []*models.Plan
	var total int64

	// Get total count
	if err := s.db.Model(&models.Plan{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count plans", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count plans: %w", err)
	}

	// Get plans with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).Limit(limit).Offset(offset).Order("sort_order ASC, created_at ASC").Find(&plans).Error; err != nil {
		s.logger.Error("Failed to get plans", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get plans: %w", err)
	}

	return plans, total, nil
}

// GetPublicPlans retrieves public plans for a tenant.
func (s *PaymentService) GetPublicPlans(tenantID uint) ([]*models.Plan, error) {
	var plans []*models.Plan
	if err := s.db.Where("tenant_id = ? AND is_active = ? AND is_public = ?", tenantID, true, true).Order("sort_order ASC, created_at ASC").Find(&plans).Error; err != nil {
		s.logger.Error("Failed to get public plans", zap.Error(err))
		return nil, fmt.Errorf("failed to get public plans: %w", err)
	}

	return plans, nil
}

// UpdatePlan updates a plan.
func (s *PaymentService) UpdatePlan(plan *models.Plan) error {
	if err := s.db.Save(plan).Error; err != nil {
		s.logger.Error("Failed to update plan", zap.Error(err))
		return fmt.Errorf("failed to update plan: %w", err)
	}

	s.logger.Info("Plan updated successfully", zap.Uint("plan_id", plan.ID))
	return nil
}

// DeletePlan soft deletes a plan.
func (s *PaymentService) DeletePlan(id uint) error {
	if err := s.db.Delete(&models.Plan{}, id).Error; err != nil {
		s.logger.Error("Failed to delete plan", zap.Error(err))
		return fmt.Errorf("failed to delete plan: %w", err)
	}

	s.logger.Info("Plan deleted successfully", zap.Uint("plan_id", id))
	return nil
}

// Invoice Management

// CreateInvoice creates a new invoice.
func (s *PaymentService) CreateInvoice(invoice *models.Invoice) error {
	// Set default values
	if invoice.Currency == "" {
		invoice.Currency = "USD"
	}
	if invoice.Status == "" {
		invoice.Status = "draft"
	}
	if invoice.TotalAmount == 0 {
		invoice.TotalAmount = invoice.Amount + invoice.TaxAmount - invoice.DiscountAmount
	}

	// Generate invoice number if not provided
	if invoice.InvoiceNumber == "" {
		invoice.InvoiceNumber = s.generateInvoiceNumber()
	}

	// Create invoice
	if err := s.db.Create(invoice).Error; err != nil {
		s.logger.Error("Failed to create invoice", zap.Error(err))
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	s.logger.Info("Invoice created successfully", zap.Uint("invoice_id", invoice.ID))
	return nil
}

// GetInvoice retrieves an invoice by ID.
func (s *PaymentService) GetInvoice(id uint) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := s.db.Preload("Subscription").Preload("Payment").First(&invoice, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice not found")
		}
		s.logger.Error("Failed to get invoice", zap.Error(err))
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return &invoice, nil
}

// GetInvoices retrieves a list of invoices with pagination.
func (s *PaymentService) GetInvoices(tenantID uint, limit, offset int) ([]*models.Invoice, int64, error) {
	var invoices []*models.Invoice
	var total int64

	// Get total count
	if err := s.db.Model(&models.Invoice{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count invoices", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	// Get invoices with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).Preload("Subscription").Preload("Payment").Limit(limit).Offset(offset).Order("created_at DESC").Find(&invoices).Error; err != nil {
		s.logger.Error("Failed to get invoices", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get invoices: %w", err)
	}

	return invoices, total, nil
}

// PayInvoice marks an invoice as paid.
func (s *PaymentService) PayInvoice(invoiceID uint, paymentID uint) error {
	var invoice models.Invoice
	if err := s.db.First(&invoice, invoiceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("invoice not found")
		}
		s.logger.Error("Failed to get invoice", zap.Error(err))
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	invoice.Status = "paid"
	invoice.PaymentID = &paymentID
	now := time.Now()
	invoice.PaidAt = &now

	if err := s.db.Save(&invoice).Error; err != nil {
		s.logger.Error("Failed to pay invoice", zap.Error(err))
		return fmt.Errorf("failed to pay invoice: %w", err)
	}

	s.logger.Info("Invoice paid successfully", zap.Uint("invoice_id", invoiceID))
	return nil
}

// generateInvoiceNumber generates a unique invoice number.
func (s *PaymentService) generateInvoiceNumber() string {
	// Simple implementation - in production, you might want a more sophisticated approach
	return fmt.Sprintf("INV-%d", time.Now().Unix())
}

// InitDatabase initializes the database connection and runs migrations.
func InitDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.Payment{},
		&models.PaymentTransaction{},
		&models.Subscription{},
		&models.Plan{},
		&models.Invoice{},
		&models.BillingUsage{},
		&models.PaymentWebhook{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
