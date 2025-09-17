// Package services provides custom metrics monitoring business logic.
package services

import (
	"fmt"
	"time"

	"github.com/anupamdutta5/statuspage-monitoring-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CustomMetricsService handles custom metrics monitoring business logic.
type CustomMetricsService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewCustomMetricsService creates a new custom metrics service.
func NewCustomMetricsService(db *gorm.DB, logger *zap.Logger) *CustomMetricsService {
	return &CustomMetricsService{
		db:     db,
		logger: logger,
	}
}

// CreateCustomMetric creates a new custom metric.
func (s *CustomMetricsService) CreateCustomMetric(metric *models.CustomMetric) error {
	if err := metric.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values
	if metric.RetentionDays == 0 {
		metric.RetentionDays = 30
	}

	if err := s.db.Create(metric).Error; err != nil {
		s.logger.Error("Failed to create custom metric", zap.Error(err))
		return fmt.Errorf("failed to create custom metric: %w", err)
	}

	s.logger.Info("Custom metric created successfully", zap.Uint("metric_id", metric.ID))
	return nil
}

// GetCustomMetric retrieves a custom metric by ID.
func (s *CustomMetricsService) GetCustomMetric(id uint) (*models.CustomMetric, error) {
	var metric models.CustomMetric
	if err := s.db.First(&metric, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("custom metric not found")
		}
		s.logger.Error("Failed to get custom metric", zap.Error(err))
		return nil, fmt.Errorf("failed to get custom metric: %w", err)
	}

	return &metric, nil
}

// GetCustomMetrics retrieves a list of custom metrics with pagination.
func (s *CustomMetricsService) GetCustomMetrics(tenantID uint, limit, offset int) ([]*models.CustomMetric, int64, error) {
	var metrics []*models.CustomMetric
	var total int64

	// Get total count
	if err := s.db.Model(&models.CustomMetric{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count custom metrics", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count custom metrics: %w", err)
	}

	// Get metrics with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&metrics).Error; err != nil {
		s.logger.Error("Failed to get custom metrics", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get custom metrics: %w", err)
	}

	return metrics, total, nil
}

// UpdateCustomMetric updates a custom metric.
func (s *CustomMetricsService) UpdateCustomMetric(metric *models.CustomMetric) error {
	if err := metric.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Save(metric).Error; err != nil {
		s.logger.Error("Failed to update custom metric", zap.Error(err))
		return fmt.Errorf("failed to update custom metric: %w", err)
	}

	s.logger.Info("Custom metric updated successfully", zap.Uint("metric_id", metric.ID))
	return nil
}

// DeleteCustomMetric soft deletes a custom metric.
func (s *CustomMetricsService) DeleteCustomMetric(id uint) error {
	if err := s.db.Delete(&models.CustomMetric{}, id).Error; err != nil {
		s.logger.Error("Failed to delete custom metric", zap.Error(err))
		return fmt.Errorf("failed to delete custom metric: %w", err)
	}

	s.logger.Info("Custom metric deleted successfully", zap.Uint("metric_id", id))
	return nil
}

// CreateCustomMetricDataPoint creates a new custom metric data point.
func (s *CustomMetricsService) CreateCustomMetricDataPoint(dataPoint *models.CustomMetricDataPoint) error {
	if err := dataPoint.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default timestamp if not provided
	if dataPoint.Timestamp.IsZero() {
		dataPoint.Timestamp = time.Now()
	}

	if err := s.db.Create(dataPoint).Error; err != nil {
		s.logger.Error("Failed to create custom metric data point", zap.Error(err))
		return fmt.Errorf("failed to create custom metric data point: %w", err)
	}

	s.logger.Info("Custom metric data point created successfully", zap.Uint("data_point_id", dataPoint.ID))
	return nil
}

// CreateBulkCustomMetricDataPoints creates multiple custom metric data points in a single transaction.
func (s *CustomMetricsService) CreateBulkCustomMetricDataPoints(dataPoints []*models.CustomMetricDataPoint) error {
	if len(dataPoints) == 0 {
		return fmt.Errorf("no data points provided")
	}

	// Validate all data points
	for i, dp := range dataPoints {
		if err := dp.Validate(); err != nil {
			return fmt.Errorf("validation failed for data point %d: %w", i, err)
		}
		if dp.Timestamp.IsZero() {
			dp.Timestamp = time.Now()
		}
	}

	// Create in batch
	if err := s.db.CreateInBatches(dataPoints, 100).Error; err != nil {
		s.logger.Error("Failed to create bulk custom metric data points", zap.Error(err))
		return fmt.Errorf("failed to create bulk custom metric data points: %w", err)
	}

	s.logger.Info("Bulk custom metric data points created successfully", zap.Int("count", len(dataPoints)))
	return nil
}

// GetCustomMetricDataPoints retrieves data points for a custom metric within a time range.
func (s *CustomMetricsService) GetCustomMetricDataPoints(metricID uint, startDate, endDate time.Time, limit, offset int) ([]*models.CustomMetricDataPoint, int64, error) {
	var dataPoints []*models.CustomMetricDataPoint
	var total int64

	// Build query
	query := s.db.Model(&models.CustomMetricDataPoint{}).Where("metric_id = ?", metricID)
	if !startDate.IsZero() {
		query = query.Where("timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("timestamp <= ?", endDate)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count custom metric data points", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count custom metric data points: %w", err)
	}

	// Get data points with pagination
	if err := query.Preload("Metric").
		Limit(limit).
		Offset(offset).
		Order("timestamp DESC").
		Find(&dataPoints).Error; err != nil {
		s.logger.Error("Failed to get custom metric data points", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get custom metric data points: %w", err)
	}

	return dataPoints, total, nil
}

// GetCustomMetricStatistics returns statistics for a custom metric.
func (s *CustomMetricsService) GetCustomMetricStatistics(metricID uint, startDate, endDate time.Time) (map[string]interface{}, error) {
	var stats map[string]interface{} = make(map[string]interface{})

	// Build query
	query := s.db.Model(&models.CustomMetricDataPoint{}).Where("metric_id = ?", metricID)
	if !startDate.IsZero() {
		query = query.Where("timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("timestamp <= ?", endDate)
	}

	// Get basic statistics
	var count int64
	var sum, avg, min, max float64

	if err := query.Count(&count).Error; err != nil {
		s.logger.Error("Failed to count data points", zap.Error(err))
		return nil, fmt.Errorf("failed to count data points: %w", err)
	}

	if count > 0 {
		if err := query.Select("SUM(value) as sum, AVG(value) as avg, MIN(value) as min, MAX(value) as max").
			Row().Scan(&sum, &avg, &min, &max); err != nil {
			s.logger.Error("Failed to calculate statistics", zap.Error(err))
			return nil, fmt.Errorf("failed to calculate statistics: %w", err)
		}
	}

	stats["metric_id"] = metricID
	stats["count"] = count
	stats["sum"] = sum
	stats["average"] = avg
	stats["minimum"] = min
	stats["maximum"] = max
	stats["date_range"] = map[string]interface{}{
		"start": startDate,
		"end":   endDate,
	}
	stats["timestamp"] = time.Now()

	return stats, nil
}

// CleanupOldDataPoints removes data points older than the retention period.
func (s *CustomMetricsService) CleanupOldDataPoints() error {
	// Get all active metrics with their retention periods
	var metrics []models.CustomMetric
	if err := s.db.Where("is_active = ?", true).Find(&metrics).Error; err != nil {
		s.logger.Error("Failed to get active metrics", zap.Error(err))
		return fmt.Errorf("failed to get active metrics: %w", err)
	}

	totalDeleted := 0
	for _, metric := range metrics {
		cutoffDate := time.Now().AddDate(0, 0, -metric.RetentionDays)

		result := s.db.Where("metric_id = ? AND timestamp < ?", metric.ID, cutoffDate).
			Delete(&models.CustomMetricDataPoint{})

		if result.Error != nil {
			s.logger.Error("Failed to cleanup data points for metric",
				zap.Uint("metric_id", metric.ID),
				zap.Error(result.Error))
			continue
		}

		if result.RowsAffected > 0 {
			totalDeleted += int(result.RowsAffected)
			s.logger.Info("Cleaned up old data points for metric",
				zap.Uint("metric_id", metric.ID),
				zap.Int64("deleted_count", result.RowsAffected))
		}
	}

	s.logger.Info("Cleanup completed", zap.Int("total_deleted", totalDeleted))
	return nil
}

// GetCustomMetricByName retrieves a custom metric by name and tenant ID.
func (s *CustomMetricsService) GetCustomMetricByName(tenantID uint, name string) (*models.CustomMetric, error) {
	var metric models.CustomMetric
	if err := s.db.Where("tenant_id = ? AND name = ?", tenantID, name).First(&metric).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("custom metric not found")
		}
		s.logger.Error("Failed to get custom metric by name", zap.Error(err))
		return nil, fmt.Errorf("failed to get custom metric by name: %w", err)
	}

	return &metric, nil
}

