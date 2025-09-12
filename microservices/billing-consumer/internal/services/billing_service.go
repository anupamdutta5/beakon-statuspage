// Package services provides business logic for the Billing Consumer.
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage-billing-consumer/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BillingService handles billing-related business logic.
type BillingService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewBillingService creates a new billing service.
func NewBillingService(db *gorm.DB, logger *zap.Logger) *BillingService {
	return &BillingService{
		db:     db,
		logger: logger,
	}
}

// ProcessSubscription processes a subscription event.
func (s *BillingService) ProcessSubscription(ctx context.Context, event *models.BillingEvent) error {
	s.logger.Info("Processing subscription event",
		zap.String("event_id", event.ID),
		zap.String("subscription_id", event.SubscriptionID),
		zap.Float64("amount", event.Amount))

	// Create billing record
	billingRecord := &models.BillingRecord{
		TenantID:       event.TenantID,
		UserID:         event.UserID,
		Type:           event.Type,
		SubscriptionID: event.SubscriptionID,
		Amount:         event.Amount,
		Currency:       event.Currency,
		Description:    event.Description,
		Status:         event.Status,
		Timestamp:      event.Timestamp,
	}

	// Add metadata if present
	if event.Metadata != nil {
		// Convert metadata to JSON string
		// For now, we'll store it as empty string
		billingRecord.Metadata = ""
	}

	// Save to database
	if err := s.db.Create(billingRecord).Error; err != nil {
		s.logger.Error("Failed to save billing record", zap.Error(err))
		return fmt.Errorf("failed to save billing record: %w", err)
	}

	s.logger.Info("Billing record saved successfully",
		zap.Uint("billing_record_id", billingRecord.ID))

	return nil
}

// ProcessUsage processes a usage event.
func (s *BillingService) ProcessUsage(ctx context.Context, event *models.BillingEvent) error {
	s.logger.Info("Processing usage event",
		zap.String("event_id", event.ID),
		zap.String("subscription_id", event.SubscriptionID),
		zap.Float64("amount", event.Amount))

	// Create billing record
	billingRecord := &models.BillingRecord{
		TenantID:       event.TenantID,
		UserID:         event.UserID,
		Type:           event.Type,
		SubscriptionID: event.SubscriptionID,
		Amount:         event.Amount,
		Currency:       event.Currency,
		Description:    event.Description,
		Status:         event.Status,
		Timestamp:      event.Timestamp,
	}

	// Add metadata if present
	if event.Metadata != nil {
		// Convert metadata to JSON string
		// For now, we'll store it as empty string
		billingRecord.Metadata = ""
	}

	// Save to database
	if err := s.db.Create(billingRecord).Error; err != nil {
		s.logger.Error("Failed to save billing record", zap.Error(err))
		return fmt.Errorf("failed to save billing record: %w", err)
	}

	s.logger.Info("Billing record saved successfully",
		zap.Uint("billing_record_id", billingRecord.ID))

	return nil
}

// GetBillingStats returns billing processing statistics.
func (s *BillingService) GetBillingStats() (map[string]interface{}, error) {
	var stats map[string]interface{} = make(map[string]interface{})

	// Get total processed count
	var totalProcessed int64
	if err := s.db.Model(&models.ProcessingLog{}).Where("status = ?", "completed").Count(&totalProcessed).Error; err != nil {
		s.logger.Error("Failed to count processed logs", zap.Error(err))
		return nil, fmt.Errorf("failed to count processed logs: %w", err)
	}

	// Get total failed count
	var totalFailed int64
	if err := s.db.Model(&models.ProcessingLog{}).Where("status = ?", "failed").Count(&totalFailed).Error; err != nil {
		s.logger.Error("Failed to count failed logs", zap.Error(err))
		return nil, fmt.Errorf("failed to count failed logs: %w", err)
	}

	// Get average processing time
	var avgProcessingTime float64
	if err := s.db.Model(&models.ProcessingLog{}).Where("status = ? AND processing_time > 0", "completed").Select("AVG(processing_time)").Scan(&avgProcessingTime).Error; err != nil {
		s.logger.Error("Failed to calculate average processing time", zap.Error(err))
		return nil, fmt.Errorf("failed to calculate average processing time: %w", err)
	}

	// Get recent processing logs (last 24 hours)
	var recentProcessed int64
	last24Hours := time.Now().Add(-24 * time.Hour)
	if err := s.db.Model(&models.ProcessingLog{}).Where("status = ? AND created_at >= ?", "completed", last24Hours).Count(&recentProcessed).Error; err != nil {
		s.logger.Error("Failed to count recent processed logs", zap.Error(err))
		return nil, fmt.Errorf("failed to count recent processed logs: %w", err)
	}

	stats["total_processed"] = totalProcessed
	stats["total_failed"] = totalFailed
	stats["average_processing_time_ms"] = avgProcessingTime
	stats["recent_processed_24h"] = recentProcessed
	stats["success_rate"] = float64(0)
	if totalProcessed+totalFailed > 0 {
		stats["success_rate"] = float64(totalProcessed) / float64(totalProcessed+totalFailed) * 100
	}
	stats["last_updated"] = time.Now().UTC()

	return stats, nil
}

