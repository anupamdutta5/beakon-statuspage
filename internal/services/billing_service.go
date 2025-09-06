package services

import (
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BillingService struct {
	db *gorm.DB
}

func NewBillingService() *BillingService {
	return &BillingService{
		db: database.DB,
	}
}

// CreateCustomer creates a new customer in the billing system
func (s *BillingService) CreateCustomer(tenantID uint, email, name string) (*models.BillingCustomer, error) {
	customer := &models.BillingCustomer{
		TenantID:  tenantID,
		Email:     email,
		Name:      name,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.db.Create(customer).Error; err != nil {
		logger.Error("Failed to create billing customer", zap.Error(err))
		return nil, err
	}

	logger.Info("Billing customer created successfully",
		zap.Uint("tenant_id", tenantID),
		zap.String("email", email))

	return customer, nil
}

// CreateSubscription creates a new subscription in the billing system
func (s *BillingService) CreateSubscription(tenantID uint, planID uint, customerID string) (*models.BillingSubscription, error) {
	// Get the plan
	var plan models.SubscriptionPlan
	if err := s.db.First(&plan, planID).Error; err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	// Create subscription
	subscription := &models.BillingSubscription{
		TenantID:           tenantID,
		PlanID:             planID,
		CustomerID:         customerID,
		Status:             "active",
		CurrentPeriodStart: time.Now(),
		CurrentPeriodEnd:   time.Now().AddDate(0, 1, 0), // 1 month from now
		Price:              plan.Price,
		Currency:           plan.Currency,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.db.Create(subscription).Error; err != nil {
		logger.Error("Failed to create billing subscription", zap.Error(err))
		return nil, err
	}

	// Create initial invoice
	if err := s.createInvoice(subscription.ID, plan.Price, plan.Currency); err != nil {
		logger.Error("Failed to create initial invoice", zap.Error(err))
		// Don't fail the subscription creation, just log the error
	}

	logger.Info("Billing subscription created successfully",
		zap.Uint("tenant_id", tenantID),
		zap.Uint("plan_id", planID))

	return subscription, nil
}

// UpdateSubscription updates a subscription (plan change, etc.)
func (s *BillingService) UpdateSubscription(subscriptionID uint, newPlanID uint) error {
	// Get current subscription
	var subscription models.BillingSubscription
	if err := s.db.Preload("Plan").First(&subscription, subscriptionID).Error; err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Get new plan
	var newPlan models.SubscriptionPlan
	if err := s.db.First(&newPlan, newPlanID).Error; err != nil {
		return fmt.Errorf("new plan not found: %w", err)
	}

	// Update subscription
	subscription.PlanID = newPlanID
	subscription.Price = newPlan.Price
	subscription.Currency = newPlan.Currency
	subscription.UpdatedAt = time.Now()

	if err := s.db.Save(&subscription).Error; err != nil {
		logger.Error("Failed to update billing subscription", zap.Error(err))
		return err
	}

	// Create prorated invoice for plan change
	proratedAmount := s.calculateProratedAmount(subscription.Price, newPlan.Price, subscription.CurrentPeriodStart, subscription.CurrentPeriodEnd)
	if proratedAmount != 0 {
		if err := s.createInvoice(subscription.ID, proratedAmount, newPlan.Currency); err != nil {
			logger.Error("Failed to create prorated invoice", zap.Error(err))
		}
	}

	logger.Info("Billing subscription updated successfully",
		zap.Uint("subscription_id", subscriptionID),
		zap.Uint("new_plan_id", newPlanID))

	return nil
}

// CancelSubscription cancels a subscription
func (s *BillingService) CancelSubscription(subscriptionID uint) error {
	var subscription models.BillingSubscription
	if err := s.db.First(&subscription, subscriptionID).Error; err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	subscription.Status = "cancelled"
	subscription.CancelledAt = &[]time.Time{time.Now()}[0]
	subscription.UpdatedAt = time.Now()

	if err := s.db.Save(&subscription).Error; err != nil {
		logger.Error("Failed to cancel billing subscription", zap.Error(err))
		return err
	}

	logger.Info("Billing subscription cancelled successfully",
		zap.Uint("subscription_id", subscriptionID))

	return nil
}

// CreateInvoice creates a new invoice
func (s *BillingService) createInvoice(subscriptionID uint, amount float64, currency string) error {
	invoice := &models.BillingInvoice{
		SubscriptionID: subscriptionID,
		Amount:         amount,
		Currency:       currency,
		Status:         "pending",
		DueDate:        time.Now().AddDate(0, 0, 7), // 7 days from now
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.db.Create(invoice).Error; err != nil {
		logger.Error("Failed to create billing invoice", zap.Error(err))
		return err
	}

	return nil
}

// RecordPayment records a payment for an invoice
func (s *BillingService) RecordPayment(invoiceID uint, amount float64, paymentMethod string, externalID string) error {
	payment := &models.BillingPayment{
		InvoiceID:     invoiceID,
		Amount:        amount,
		PaymentMethod: paymentMethod,
		ExternalID:    externalID,
		Status:        "completed",
		ProcessedAt:   time.Now(),
		CreatedAt:     time.Now(),
	}

	if err := s.db.Create(payment).Error; err != nil {
		logger.Error("Failed to record billing payment", zap.Error(err))
		return err
	}

	// Update invoice status
	var invoice models.BillingInvoice
	if err := s.db.First(&invoice, invoiceID).Error; err != nil {
		return err
	}

	invoice.Status = "paid"
	invoice.PaidAt = &[]time.Time{time.Now()}[0]
	invoice.UpdatedAt = time.Now()

	if err := s.db.Save(&invoice).Error; err != nil {
		logger.Error("Failed to update invoice status", zap.Error(err))
		return err
	}

	logger.Info("Payment recorded successfully",
		zap.Uint("invoice_id", invoiceID),
		zap.Float64("amount", amount))

	return nil
}

// GetBillingStats gets billing statistics
func (s *BillingService) GetBillingStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total revenue
	var totalRevenue float64
	if err := s.db.Model(&models.BillingPayment{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("status = ?", "completed").
		Scan(&totalRevenue).Error; err != nil {
		return nil, err
	}
	stats["total_revenue"] = totalRevenue

	// Monthly recurring revenue
	var mrr float64
	if err := s.db.Model(&models.BillingSubscription{}).
		Select("COALESCE(SUM(price), 0)").
		Where("status = ?", "active").
		Scan(&mrr).Error; err != nil {
		return nil, err
	}
	stats["monthly_recurring_revenue"] = mrr

	// Active subscriptions count
	var activeSubscriptions int64
	if err := s.db.Model(&models.BillingSubscription{}).
		Where("status = ?", "active").
		Count(&activeSubscriptions).Error; err != nil {
		return nil, err
	}
	stats["active_subscriptions"] = activeSubscriptions

	// Pending invoices count
	var pendingInvoices int64
	if err := s.db.Model(&models.BillingInvoice{}).
		Where("status = ?", "pending").
		Count(&pendingInvoices).Error; err != nil {
		return nil, err
	}
	stats["pending_invoices"] = pendingInvoices

	// Overdue invoices count
	var overdueInvoices int64
	if err := s.db.Model(&models.BillingInvoice{}).
		Where("status = ? AND due_date < ?", "pending", time.Now()).
		Count(&overdueInvoices).Error; err != nil {
		return nil, err
	}
	stats["overdue_invoices"] = overdueInvoices

	return stats, nil
}

// GetRevenueChart gets revenue data for charts
func (s *BillingService) GetRevenueChart(months int) (map[string]interface{}, error) {
	// Get revenue data for the last N months
	var results []struct {
		Month  string  `json:"month"`
		Amount float64 `json:"amount"`
	}

	query := `
		SELECT 
			DATE_TRUNC('month', processed_at) as month,
			COALESCE(SUM(amount), 0) as amount
		FROM billing_payments 
		WHERE status = 'completed' 
			AND processed_at >= NOW() - INTERVAL '%d months'
		GROUP BY DATE_TRUNC('month', processed_at)
		ORDER BY month
	`

	if err := s.db.Raw(fmt.Sprintf(query, months)).Scan(&results).Error; err != nil {
		return nil, err
	}

	// Format the data for charts
	labels := make([]string, 0, len(results))
	data := make([]float64, 0, len(results))

	for _, result := range results {
		labels = append(labels, result.Month)
		data = append(data, result.Amount)
	}

	return map[string]interface{}{
		"labels": labels,
		"data":   data,
	}, nil
}

// ProcessWebhook processes webhooks from payment providers
func (s *BillingService) ProcessWebhook(webhookData map[string]interface{}) error {
	eventType, ok := webhookData["type"].(string)
	if !ok {
		return fmt.Errorf("invalid webhook data: missing type")
	}

	switch eventType {
	case "payment.succeeded":
		return s.handlePaymentSucceeded(webhookData)
	case "payment.failed":
		return s.handlePaymentFailed(webhookData)
	case "subscription.updated":
		return s.handleSubscriptionUpdated(webhookData)
	case "subscription.cancelled":
		return s.handleSubscriptionCancelled(webhookData)
	default:
		logger.Info("Unhandled webhook event type", zap.String("type", eventType))
		return nil
	}
}

// Private helper methods

func (s *BillingService) calculateProratedAmount(oldPrice, newPrice float64, startDate, endDate time.Time) float64 {
	// Calculate prorated amount for plan changes
	daysRemaining := time.Until(endDate).Hours() / 24
	totalDays := endDate.Sub(startDate).Hours() / 24

	if daysRemaining <= 0 {
		return 0
	}

	proratedOld := (oldPrice * daysRemaining) / totalDays
	proratedNew := (newPrice * daysRemaining) / totalDays

	return proratedNew - proratedOld
}

func (s *BillingService) handlePaymentSucceeded(webhookData map[string]interface{}) error {
	// Extract payment information from webhook
	_, ok := webhookData["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid webhook data structure")
	}

	// Find the invoice and record the payment
	// This would typically involve looking up the invoice by external ID
	// and recording the payment

	logger.Info("Payment succeeded webhook processed")
	return nil
}

func (s *BillingService) handlePaymentFailed(webhookData map[string]interface{}) error {
	// Handle failed payment
	logger.Info("Payment failed webhook processed")
	return nil
}

func (s *BillingService) handleSubscriptionUpdated(webhookData map[string]interface{}) error {
	// Handle subscription update
	logger.Info("Subscription updated webhook processed")
	return nil
}

func (s *BillingService) handleSubscriptionCancelled(webhookData map[string]interface{}) error {
	// Handle subscription cancellation
	logger.Info("Subscription cancelled webhook processed")
	return nil
}
