package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AnalyticsService implements the AnalyticsService interface
type AnalyticsService struct {
	db *gorm.DB
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{
		db: db,
	}
}

// RecordPaymentEvent records a payment event for analytics
func (s *AnalyticsService) RecordPaymentEvent(ctx context.Context, event *PaymentEvent) error {
	// Create payment event record
	paymentEvent := &models.PaymentEvent{
		TenantID:  event.TenantID,
		EventType: event.EventType,
		PaymentID: event.PaymentID,
		Amount:    event.Amount,
		Currency:  event.Currency,
		Method:    event.Method,
		Gateway:   event.Gateway,
		Timestamp: event.Timestamp,
		Metadata:  `{}`, // Convert to JSON string
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(paymentEvent).Error; err != nil {
		logger.Error("Failed to record payment event",
			zap.String("payment_id", event.PaymentID),
			zap.String("event_type", event.EventType),
			zap.Error(err))
		return err
	}

	logger.Info("Payment event recorded",
		zap.String("payment_id", event.PaymentID),
		zap.String("event_type", event.EventType),
		zap.Uint("tenant_id", event.TenantID))

	return nil
}

// GetPaymentMetrics returns payment metrics
func (s *AnalyticsService) GetPaymentMetrics(ctx context.Context, tenantID uint, period string) (*PaymentMetrics, error) {
	// Calculate date range based on period
	startDate, endDate := s.getDateRange(period)

	// Get total payments
	var totalPayments int64
	if err := s.db.Model(&models.PaymentEvent{}).
		Where("tenant_id = ? AND timestamp BETWEEN ? AND ?", tenantID, startDate, endDate).
		Count(&totalPayments).Error; err != nil {
		return nil, fmt.Errorf("failed to get total payments: %w", err)
	}

	// Get successful payments
	var successfulPayments int64
	if err := s.db.Model(&models.PaymentEvent{}).
		Where("tenant_id = ? AND event_type = ? AND timestamp BETWEEN ? AND ?",
			tenantID, "completed", startDate, endDate).
		Count(&successfulPayments).Error; err != nil {
		return nil, fmt.Errorf("failed to get successful payments: %w", err)
	}

	// Get failed payments
	var failedPayments int64
	if err := s.db.Model(&models.PaymentEvent{}).
		Where("tenant_id = ? AND event_type = ? AND timestamp BETWEEN ? AND ?",
			tenantID, "failed", startDate, endDate).
		Count(&failedPayments).Error; err != nil {
		return nil, fmt.Errorf("failed to get failed payments: %w", err)
	}

	// Get total amount
	var totalAmount float64
	if err := s.db.Model(&models.PaymentEvent{}).
		Where("tenant_id = ? AND timestamp BETWEEN ? AND ?", tenantID, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalAmount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total amount: %w", err)
	}

	// Get successful amount
	var successfulAmount float64
	if err := s.db.Model(&models.PaymentEvent{}).
		Where("tenant_id = ? AND event_type = ? AND timestamp BETWEEN ? AND ?",
			tenantID, "completed", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&successfulAmount).Error; err != nil {
		return nil, fmt.Errorf("failed to get successful amount: %w", err)
	}

	// Get failed amount
	var failedAmount float64
	if err := s.db.Model(&models.PaymentEvent{}).
		Where("tenant_id = ? AND event_type = ? AND timestamp BETWEEN ? AND ?",
			tenantID, "failed", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&failedAmount).Error; err != nil {
		return nil, fmt.Errorf("failed to get failed amount: %w", err)
	}

	// Get refund count
	var refundCount int64
	if err := s.db.Model(&models.PaymentEvent{}).
		Where("tenant_id = ? AND event_type = ? AND timestamp BETWEEN ? AND ?",
			tenantID, "refunded", startDate, endDate).
		Count(&refundCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get refund count: %w", err)
	}

	// Get refund amount
	var refundAmount float64
	if err := s.db.Model(&models.PaymentEvent{}).
		Where("tenant_id = ? AND event_type = ? AND timestamp BETWEEN ? AND ?",
			tenantID, "refunded", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&refundAmount).Error; err != nil {
		return nil, fmt.Errorf("failed to get refund amount: %w", err)
	}

	// Calculate conversion rate
	var conversionRate float64
	if totalPayments > 0 {
		conversionRate = float64(successfulPayments) / float64(totalPayments) * 100
	}

	// Calculate average amount
	var averageAmount float64
	if successfulPayments > 0 {
		averageAmount = successfulAmount / float64(successfulPayments)
	}

	metrics := &PaymentMetrics{
		TotalPayments:      int(totalPayments),
		SuccessfulPayments: int(successfulPayments),
		FailedPayments:     int(failedPayments),
		TotalAmount:        totalAmount,
		SuccessfulAmount:   successfulAmount,
		FailedAmount:       failedAmount,
		ConversionRate:     conversionRate,
		AverageAmount:      averageAmount,
		RefundCount:        int(refundCount),
		RefundAmount:       refundAmount,
	}

	return metrics, nil
}

// GetConversionFunnel returns conversion funnel data
func (s *AnalyticsService) GetConversionFunnel(ctx context.Context, tenantID uint, period string) (*ConversionFunnel, error) {
	// Calculate date range based on period
	startDate, endDate := s.getDateRange(period)

	// Get funnel steps
	steps := []ConversionStep{
		{Name: "Payment Created"},
		{Name: "Payment Processing"},
		{Name: "Payment Completed"},
		{Name: "Payment Successful"},
	}

	// Get counts for each step
	var createdCount int64
	if err := s.db.Model(&models.PaymentEvent{}).
		Where("tenant_id = ? AND event_type = ? AND timestamp BETWEEN ? AND ?",
			tenantID, "created", startDate, endDate).
		Count(&createdCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get created count: %w", err)
	}

	var processingCount int64
	if err := s.db.Model(&models.PaymentEvent{}).
		Where("tenant_id = ? AND event_type = ? AND timestamp BETWEEN ? AND ?",
			tenantID, "processing", startDate, endDate).
		Count(&processingCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get processing count: %w", err)
	}

	var completedCount int64
	if err := s.db.Model(&models.PaymentEvent{}).
		Where("tenant_id = ? AND event_type = ? AND timestamp BETWEEN ? AND ?",
			tenantID, "completed", startDate, endDate).
		Count(&completedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get completed count: %w", err)
	}

	var successfulCount int64
	if err := s.db.Model(&models.PaymentEvent{}).
		Where("tenant_id = ? AND event_type = ? AND timestamp BETWEEN ? AND ?",
			tenantID, "completed", startDate, endDate).
		Count(&successfulCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get successful count: %w", err)
	}

	// Set counts and calculate percentages
	steps[0].Count = int(createdCount)
	steps[0].Percentage = 100.0

	steps[1].Count = int(processingCount)
	if createdCount > 0 {
		steps[1].Percentage = float64(processingCount) / float64(createdCount) * 100
		steps[1].DropOffRate = 100 - steps[1].Percentage
	}

	steps[2].Count = int(completedCount)
	if processingCount > 0 {
		steps[2].Percentage = float64(completedCount) / float64(processingCount) * 100
		steps[2].DropOffRate = 100 - steps[2].Percentage
	}

	steps[3].Count = int(successfulCount)
	if completedCount > 0 {
		steps[3].Percentage = float64(successfulCount) / float64(completedCount) * 100
		steps[3].DropOffRate = 100 - steps[3].Percentage
	}

	funnel := &ConversionFunnel{
		Steps: steps,
	}

	return funnel, nil
}

// getDateRange calculates date range based on period
func (s *AnalyticsService) getDateRange(period string) (time.Time, time.Time) {
	endDate := time.Now()
	var startDate time.Time

	switch period {
	case "today":
		startDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, endDate.Location())
	case "yesterday":
		yesterday := endDate.AddDate(0, 0, -1)
		startDate = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
		endDate = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 999999999, yesterday.Location())
	case "week":
		startDate = endDate.AddDate(0, 0, -7)
	case "month":
		startDate = endDate.AddDate(0, -1, 0)
	case "quarter":
		startDate = endDate.AddDate(0, -3, 0)
	case "year":
		startDate = endDate.AddDate(-1, 0, 0)
	default:
		// Default to last 30 days
		startDate = endDate.AddDate(0, 0, -30)
	}

	return startDate, endDate
}
