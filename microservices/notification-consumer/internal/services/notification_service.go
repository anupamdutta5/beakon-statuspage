// Package services provides business logic for the Notification Consumer.
package services

import (
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage-notification-consumer/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NotificationService handles notification-related business logic.
type NotificationService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewNotificationService creates a new notification service.
func NewNotificationService(db *gorm.DB, logger *zap.Logger) *NotificationService {
	return &NotificationService{
		db:     db,
		logger: logger,
	}
}

// CreateProcessingLog creates a processing log entry.
func (s *NotificationService) CreateProcessingLog(log *models.ProcessingLog) error {
	if err := s.db.Create(log).Error; err != nil {
		s.logger.Error("Failed to create processing log", zap.Error(err))
		return fmt.Errorf("failed to create processing log: %w", err)
	}

	s.logger.Info("Processing log created successfully", zap.Uint("log_id", log.ID))
	return nil
}

// UpdateProcessingLog updates a processing log entry.
func (s *NotificationService) UpdateProcessingLog(log *models.ProcessingLog) error {
	if err := s.db.Save(log).Error; err != nil {
		s.logger.Error("Failed to update processing log", zap.Error(err))
		return fmt.Errorf("failed to update processing log: %w", err)
	}

	s.logger.Info("Processing log updated successfully", zap.Uint("log_id", log.ID))
	return nil
}

// GetProcessingLogs retrieves processing logs with pagination.
func (s *NotificationService) GetProcessingLogs(limit, offset int) ([]*models.ProcessingLog, int64, error) {
	var logs []*models.ProcessingLog
	var total int64

	// Get total count
	if err := s.db.Model(&models.ProcessingLog{}).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count processing logs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count processing logs: %w", err)
	}

	// Get logs with pagination
	if err := s.db.Limit(limit).Offset(offset).Order("created_at DESC").Find(&logs).Error; err != nil {
		s.logger.Error("Failed to get processing logs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get processing logs: %w", err)
	}

	return logs, total, nil
}

// GetProcessingLogsByEventType retrieves processing logs by event type.
func (s *NotificationService) GetProcessingLogsByEventType(eventType string, limit, offset int) ([]*models.ProcessingLog, int64, error) {
	var logs []*models.ProcessingLog
	var total int64

	// Get total count
	if err := s.db.Model(&models.ProcessingLog{}).Where("event_type = ?", eventType).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count processing logs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count processing logs: %w", err)
	}

	// Get logs with pagination
	if err := s.db.Where("event_type = ?", eventType).Limit(limit).Offset(offset).Order("created_at DESC").Find(&logs).Error; err != nil {
		s.logger.Error("Failed to get processing logs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get processing logs: %w", err)
	}

	return logs, total, nil
}

// GetProcessingStats returns processing statistics.
func (s *NotificationService) GetProcessingStats() (map[string]interface{}, error) {
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

	// Get total processing count
	var totalProcessing int64
	if err := s.db.Model(&models.ProcessingLog{}).Where("status = ?", "processing").Count(&totalProcessing).Error; err != nil {
		s.logger.Error("Failed to count processing logs", zap.Error(err))
		return nil, fmt.Errorf("failed to count processing logs: %w", err)
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
	stats["total_processing"] = totalProcessing
	stats["average_processing_time_ms"] = avgProcessingTime
	stats["recent_processed_24h"] = recentProcessed
	stats["success_rate"] = float64(0)
	if totalProcessed+totalFailed > 0 {
		stats["success_rate"] = float64(totalProcessed) / float64(totalProcessed+totalFailed) * 100
	}
	stats["last_updated"] = time.Now().UTC()

	return stats, nil
}

// GetProcessingStatsByEventType returns processing statistics by event type.
func (s *NotificationService) GetProcessingStatsByEventType(eventType string) (map[string]interface{}, error) {
	var stats map[string]interface{} = make(map[string]interface{})

	// Get total processed count for event type
	var totalProcessed int64
	if err := s.db.Model(&models.ProcessingLog{}).Where("event_type = ? AND status = ?", eventType, "completed").Count(&totalProcessed).Error; err != nil {
		s.logger.Error("Failed to count processed logs", zap.Error(err))
		return nil, fmt.Errorf("failed to count processed logs: %w", err)
	}

	// Get total failed count for event type
	var totalFailed int64
	if err := s.db.Model(&models.ProcessingLog{}).Where("event_type = ? AND status = ?", eventType, "failed").Count(&totalFailed).Error; err != nil {
		s.logger.Error("Failed to count failed logs", zap.Error(err))
		return nil, fmt.Errorf("failed to count failed logs: %w", err)
	}

	// Get average processing time for event type
	var avgProcessingTime float64
	if err := s.db.Model(&models.ProcessingLog{}).Where("event_type = ? AND status = ? AND processing_time > 0", eventType, "completed").Select("AVG(processing_time)").Scan(&avgProcessingTime).Error; err != nil {
		s.logger.Error("Failed to calculate average processing time", zap.Error(err))
		return nil, fmt.Errorf("failed to calculate average processing time: %w", err)
	}

	// Get recent processing logs for event type (last 24 hours)
	var recentProcessed int64
	last24Hours := time.Now().Add(-24 * time.Hour)
	if err := s.db.Model(&models.ProcessingLog{}).Where("event_type = ? AND status = ? AND created_at >= ?", eventType, "completed", last24Hours).Count(&recentProcessed).Error; err != nil {
		s.logger.Error("Failed to count recent processed logs", zap.Error(err))
		return nil, fmt.Errorf("failed to count recent processed logs: %w", err)
	}

	stats["event_type"] = eventType
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

// GetFailedProcessingLogs retrieves failed processing logs.
func (s *NotificationService) GetFailedProcessingLogs(limit, offset int) ([]*models.ProcessingLog, int64, error) {
	var logs []*models.ProcessingLog
	var total int64

	// Get total count
	if err := s.db.Model(&models.ProcessingLog{}).Where("status = ?", "failed").Count(&total).Error; err != nil {
		s.logger.Error("Failed to count failed processing logs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count failed processing logs: %w", err)
	}

	// Get logs with pagination
	if err := s.db.Where("status = ?", "failed").Limit(limit).Offset(offset).Order("created_at DESC").Find(&logs).Error; err != nil {
		s.logger.Error("Failed to get failed processing logs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get failed processing logs: %w", err)
	}

	return logs, total, nil
}

// RetryFailedProcessing retries a failed processing log.
func (s *NotificationService) RetryFailedProcessing(logID uint) error {
	var log models.ProcessingLog
	if err := s.db.First(&log, logID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("processing log not found")
		}
		s.logger.Error("Failed to get processing log", zap.Error(err))
		return fmt.Errorf("failed to get processing log: %w", err)
	}

	if log.Status != "failed" {
		return fmt.Errorf("processing log is not in failed status")
	}

	// Reset status and increment retry count
	log.Status = "processing"
	log.RetryCount++
	log.Error = ""
	if err := s.db.Save(&log).Error; err != nil {
		s.logger.Error("Failed to update processing log", zap.Error(err))
		return fmt.Errorf("failed to update processing log: %w", err)
	}

	s.logger.Info("Processing log retry initiated", zap.Uint("log_id", logID))
	return nil
}

// CleanupOldLogs cleans up old processing logs.
func (s *NotificationService) CleanupOldLogs(retentionDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	result := s.db.Where("created_at < ?", cutoffDate).Delete(&models.ProcessingLog{})
	if result.Error != nil {
		s.logger.Error("Failed to cleanup old processing logs", zap.Error(result.Error))
		return fmt.Errorf("failed to cleanup old processing logs: %w", result.Error)
	}

	s.logger.Info("Old processing logs cleaned up",
		zap.Int("retention_days", retentionDays),
		zap.Int64("deleted_count", result.RowsAffected))

	return nil
}
