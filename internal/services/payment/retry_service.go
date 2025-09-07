package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RetryService implements the RetryService interface
type RetryService struct {
	db            *gorm.DB
	retrySettings config.RetrySettings
}

// NewRetryService creates a new retry service
func NewRetryService(db *gorm.DB, retrySettings config.RetrySettings) *RetryService {
	return &RetryService{
		db:            db,
		retrySettings: retrySettings,
	}
}

// ScheduleRetry schedules a payment retry
func (s *RetryService) ScheduleRetry(ctx context.Context, paymentID string, delay time.Duration) error {
	// Create retry record
	retry := &models.PaymentRetry{
		PaymentID:   paymentID,
		ScheduledAt: time.Now().Add(delay),
		Status:      "scheduled",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(retry).Error; err != nil {
		logger.Error("Failed to schedule payment retry",
			zap.String("payment_id", paymentID),
			zap.Error(err))
		return err
	}

	logger.Info("Payment retry scheduled",
		zap.String("payment_id", paymentID),
		zap.Duration("delay", delay))

	return nil
}

// ProcessRetries processes scheduled retries
func (s *RetryService) ProcessRetries(ctx context.Context) error {
	// Get scheduled retries that are due
	var retries []models.PaymentRetry
	if err := s.db.Where("status = ? AND scheduled_at <= ?", "scheduled", time.Now()).Find(&retries).Error; err != nil {
		return fmt.Errorf("failed to get scheduled retries: %w", err)
	}

	for _, retry := range retries {
		if err := s.processRetry(ctx, &retry); err != nil {
			logger.Error("Failed to process retry",
				zap.String("payment_id", retry.PaymentID),
				zap.Error(err))
			continue
		}
	}

	return nil
}

// processRetry processes a single retry
func (s *RetryService) processRetry(ctx context.Context, retry *models.PaymentRetry) error {
	// Update retry status to processing
	retry.Status = "processing"
	retry.UpdatedAt = time.Now()
	if err := s.db.Save(retry).Error; err != nil {
		return err
	}

	// Get payment details
	var payment models.BillingPayment
	if err := s.db.Where("external_id = ?", retry.PaymentID).First(&payment).Error; err != nil {
		retry.Status = "failed"
		retry.ErrorMessage = fmt.Sprintf("Payment not found: %v", err)
		retry.UpdatedAt = time.Now()
		s.db.Save(retry)
		return err
	}

	// Check if payment should be retried
	shouldRetry, _, err := s.ShouldRetry(ctx, retry.PaymentID, "payment_failed")
	if err != nil {
		retry.Status = "failed"
		retry.ErrorMessage = fmt.Sprintf("Retry check failed: %v", err)
		retry.UpdatedAt = time.Now()
		s.db.Save(retry)
		return err
	}

	if !shouldRetry {
		retry.Status = "cancelled"
		retry.ErrorMessage = "Maximum retry attempts reached"
		retry.UpdatedAt = time.Now()
		s.db.Save(retry)
		return nil
	}

	// Here you would implement the actual retry logic
	// For now, we'll just update the retry count
	retry.Attempts++
	retry.Status = "completed"
	retry.UpdatedAt = time.Now()
	s.db.Save(retry)

	logger.Info("Payment retry processed",
		zap.String("payment_id", retry.PaymentID),
		zap.Int("attempts", retry.Attempts))

	return nil
}

// GetRetryCount returns the retry count for a payment
func (s *RetryService) GetRetryCount(ctx context.Context, paymentID string) (int, error) {
	var count int64
	if err := s.db.Model(&models.PaymentRetry{}).Where("payment_id = ?", paymentID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

// ShouldRetry determines if a payment should be retried
func (s *RetryService) ShouldRetry(ctx context.Context, paymentID string, errorType string) (bool, time.Duration, error) {
	// Get retry count
	retryCount, err := s.GetRetryCount(ctx, paymentID)
	if err != nil {
		return false, 0, err
	}

	// Check if max retries reached
	if retryCount >= s.retrySettings.MaxRetries {
		return false, 0, nil
	}

	// Check if max retry days reached
	var firstRetry models.PaymentRetry
	if err := s.db.Where("payment_id = ?", paymentID).Order("created_at ASC").First(&firstRetry).Error; err == nil {
		if time.Since(firstRetry.CreatedAt) > time.Duration(s.retrySettings.MaxRetryDays)*24*time.Hour {
			return false, 0, nil
		}
	}

	// Calculate delay based on retry count and backoff factor
	baseDelay := time.Duration(s.retrySettings.RetryInterval) * time.Hour
	delay := time.Duration(float64(baseDelay) * pow(s.retrySettings.BackoffFactor, float64(retryCount)))

	// Determine if error type is retryable
	if !s.isRetryableError(errorType) {
		return false, 0, nil
	}

	return true, delay, nil
}

// isRetryableError determines if an error type is retryable
func (s *RetryService) isRetryableError(errorType string) bool {
	retryableErrors := map[string]bool{
		"payment_failed":     true,
		"insufficient_funds": true,
		"network_error":      true,
		"timeout":            true,
		"temporary_failure":  true,
		"card_declined":      false, // Hard decline, not retryable
		"expired_card":       false, // Hard decline, not retryable
		"invalid_card":       false, // Hard decline, not retryable
		"fraud_detected":     false, // Hard decline, not retryable
	}

	retryable, exists := retryableErrors[errorType]
	return exists && retryable
}

// pow calculates x^y
func pow(x, y float64) float64 {
	result := 1.0
	for i := 0; i < int(y); i++ {
		result *= x
	}
	return result
}
