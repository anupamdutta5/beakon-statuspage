// Package services provides business logic for the Analytics Consumer.
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/anupamdutta5/analytics-consumer/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AnalyticsService handles analytics-related business logic.
type AnalyticsService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAnalyticsService creates a new analytics service.
func NewAnalyticsService(db *gorm.DB, logger *zap.Logger) *AnalyticsService {
	return &AnalyticsService{
		db:     db,
		logger: logger,
	}
}

// ProcessMetric processes a metric event.
func (s *AnalyticsService) ProcessMetric(ctx context.Context, event *models.AnalyticsEvent) error {
	s.logger.Info("Processing metric event",
		zap.String("event_id", event.ID),
		zap.String("metric_name", event.MetricName),
		zap.Float64("metric_value", event.MetricValue))

	// Create metric data record
	metricData := &models.MetricData{
		TenantID:    event.TenantID,
		MetricName:  event.MetricName,
		MetricValue: event.MetricValue,
		Timestamp:   event.Timestamp,
		UserID:      event.UserID,
		SessionID:   event.SessionID,
	}

	// Add metadata if present
	if event.Metadata != nil {
		// Convert metadata to JSON string
		// For now, we'll store it as empty string
		metricData.Metadata = ""
	}

	// Save to database
	if err := s.db.Create(metricData).Error; err != nil {
		s.logger.Error("Failed to save metric data", zap.Error(err))
		return fmt.Errorf("failed to save metric data: %w", err)
	}

	s.logger.Info("Metric data saved successfully",
		zap.Uint("metric_data_id", metricData.ID))

	return nil
}

// ProcessPageView processes a page view event.
func (s *AnalyticsService) ProcessPageView(ctx context.Context, event *models.AnalyticsEvent) error {
	s.logger.Info("Processing page view event",
		zap.String("event_id", event.ID),
		zap.String("page", event.Page))

	// Create page view data record
	pageViewData := &models.PageViewData{
		TenantID:  event.TenantID,
		Page:      event.Page,
		UserID:    event.UserID,
		SessionID: event.SessionID,
		UserAgent: event.UserAgent,
		IPAddress: event.IPAddress,
		Referrer:  event.Referrer,
		Duration:  event.Duration,
		Timestamp: event.Timestamp,
	}

	// Add metadata if present
	if event.Metadata != nil {
		// Convert metadata to JSON string
		// For now, we'll store it as empty string
		pageViewData.Metadata = ""
	}

	// Save to database
	if err := s.db.Create(pageViewData).Error; err != nil {
		s.logger.Error("Failed to save page view data", zap.Error(err))
		return fmt.Errorf("failed to save page view data: %w", err)
	}

	s.logger.Info("Page view data saved successfully",
		zap.Uint("page_view_data_id", pageViewData.ID))

	return nil
}

// ProcessUserAction processes a user action event.
func (s *AnalyticsService) ProcessUserAction(ctx context.Context, event *models.AnalyticsEvent) error {
	s.logger.Info("Processing user action event",
		zap.String("event_id", event.ID),
		zap.String("action", event.Action))

	// Create user action data record
	userActionData := &models.UserActionData{
		TenantID:  event.TenantID,
		Action:    event.Action,
		UserID:    event.UserID,
		SessionID: event.SessionID,
		Page:      event.Page,
		Timestamp: event.Timestamp,
	}

	// Add metadata if present
	if event.Metadata != nil {
		// Convert metadata to JSON string
		// For now, we'll store it as empty string
		userActionData.Metadata = ""
	}

	// Save to database
	if err := s.db.Create(userActionData).Error; err != nil {
		s.logger.Error("Failed to save user action data", zap.Error(err))
		return fmt.Errorf("failed to save user action data: %w", err)
	}

	s.logger.Info("User action data saved successfully",
		zap.Uint("user_action_data_id", userActionData.ID))

	return nil
}

// ProcessPerformance processes a performance event.
func (s *AnalyticsService) ProcessPerformance(ctx context.Context, event *models.AnalyticsEvent) error {
	s.logger.Info("Processing performance event",
		zap.String("event_id", event.ID),
		zap.String("metric_name", event.MetricName),
		zap.Float64("metric_value", event.MetricValue))

	// Create performance data record
	performanceData := &models.PerformanceData{
		TenantID:    event.TenantID,
		MetricName:  event.MetricName,
		MetricValue: event.MetricValue,
		UserID:      event.UserID,
		SessionID:   event.SessionID,
		Page:        event.Page,
		Timestamp:   event.Timestamp,
	}

	// Add metadata if present
	if event.Metadata != nil {
		// Convert metadata to JSON string
		// For now, we'll store it as empty string
		performanceData.Metadata = ""
	}

	// Save to database
	if err := s.db.Create(performanceData).Error; err != nil {
		s.logger.Error("Failed to save performance data", zap.Error(err))
		return fmt.Errorf("failed to save performance data: %w", err)
	}

	s.logger.Info("Performance data saved successfully",
		zap.Uint("performance_data_id", performanceData.ID))

	return nil
}

// ProcessError processes an error event.
func (s *AnalyticsService) ProcessError(ctx context.Context, event *models.AnalyticsEvent) error {
	s.logger.Info("Processing error event",
		zap.String("event_id", event.ID),
		zap.String("error_type", event.ErrorType))

	// Create error data record
	errorData := &models.ErrorData{
		TenantID:     event.TenantID,
		ErrorType:    event.ErrorType,
		ErrorMessage: event.ErrorMessage,
		UserID:       event.UserID,
		SessionID:    event.SessionID,
		Page:         event.Page,
		Timestamp:    event.Timestamp,
	}

	// Add metadata if present
	if event.Metadata != nil {
		// Convert metadata to JSON string
		// For now, we'll store it as empty string
		errorData.Metadata = ""
	}

	// Save to database
	if err := s.db.Create(errorData).Error; err != nil {
		s.logger.Error("Failed to save error data", zap.Error(err))
		return fmt.Errorf("failed to save error data: %w", err)
	}

	s.logger.Info("Error data saved successfully",
		zap.Uint("error_data_id", errorData.ID))

	return nil
}

// GetAnalyticsStats returns analytics processing statistics.
func (s *AnalyticsService) GetAnalyticsStats() (map[string]interface{}, error) {
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

