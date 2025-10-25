// Package jobs provides background job implementations for anomaly detection.
package jobs

import (
	"fmt"
	"time"

	"github.com/anupamdutta5/monitoring-service/internal/models"
	"github.com/anupamdutta5/monitoring-service/internal/services"
	"gorm.io/gorm"
)

// CleanupJob handles cleanup of old metrics and resolved anomalies.
type CleanupJob struct {
	db                 *gorm.DB
	baselineCalculator *services.BaselineCalculator
	interval           time.Duration
	metricRetentionDays int
	anomalyRetentionDays int
}

// NewCleanupJob creates a new CleanupJob instance.
func NewCleanupJob(db *gorm.DB, baselineCalculator *services.BaselineCalculator, interval time.Duration, metricRetentionDays, anomalyRetentionDays int) *CleanupJob {
	return &CleanupJob{
		db:                   db,
		baselineCalculator:   baselineCalculator,
		interval:             interval,
		metricRetentionDays:  metricRetentionDays,
		anomalyRetentionDays: anomalyRetentionDays,
	}
}

// Start begins the cleanup job loop.
func (cj *CleanupJob) Start() {
	ticker := time.NewTicker(cj.interval)
	defer ticker.Stop()

	fmt.Printf("[CleanupJob] Started with interval: %v\n", cj.interval)
	fmt.Printf("[CleanupJob] Metric retention: %d days, Anomaly retention: %d days\n", cj.metricRetentionDays, cj.anomalyRetentionDays)

	// Run immediately on start
	cj.cleanup()

	for range ticker.C {
		cj.cleanup()
	}
}

// cleanup performs all cleanup operations.
func (cj *CleanupJob) cleanup() {
	fmt.Printf("[CleanupJob] Starting cleanup at %v\n", time.Now())

	// Cleanup old metric snapshots
	if err := cj.cleanupOldMetrics(); err != nil {
		fmt.Printf("[CleanupJob] Error cleaning up metrics: %v\n", err)
	}

	// Cleanup expired baselines
	if err := cj.cleanupExpiredBaselines(); err != nil {
		fmt.Printf("[CleanupJob] Error cleaning up baselines: %v\n", err)
	}

	// Cleanup old resolved anomalies
	if err := cj.cleanupOldAnomalies(); err != nil {
		fmt.Printf("[CleanupJob] Error cleaning up anomalies: %v\n", err)
	}

	fmt.Printf("[CleanupJob] Cleanup completed at %v\n", time.Now())
}

// cleanupOldMetrics removes metric snapshots older than retention period.
func (cj *CleanupJob) cleanupOldMetrics() error {
	cutoff := time.Now().Add(-time.Duration(cj.metricRetentionDays) * 24 * time.Hour)

	result := cj.db.Where("timestamp < ?", cutoff).Delete(&models.MetricSnapshot{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old metrics: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		fmt.Printf("[CleanupJob] Deleted %d old metric snapshots (older than %v)\n", result.RowsAffected, cutoff)
	}

	return nil
}

// cleanupExpiredBaselines removes expired baseline records.
func (cj *CleanupJob) cleanupExpiredBaselines() error {
	if err := cj.baselineCalculator.CleanupExpiredBaselines(); err != nil {
		return fmt.Errorf("failed to cleanup expired baselines: %w", err)
	}

	return nil
}

// cleanupOldAnomalies archives or removes old resolved anomalies.
func (cj *CleanupJob) cleanupOldAnomalies() error {
	cutoff := time.Now().Add(-time.Duration(cj.anomalyRetentionDays) * 24 * time.Hour)

	// Only delete resolved and false_positive anomalies older than retention period
	result := cj.db.Where("resolved_at < ? AND status IN ?", cutoff, []string{models.StatusResolved, models.StatusFalsePositive}).
		Delete(&models.DetectedAnomaly{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old anomalies: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		fmt.Printf("[CleanupJob] Deleted %d old anomalies (resolved before %v)\n", result.RowsAffected, cutoff)
	}

	return nil
}

// CleanupOrphanedData removes orphaned records (e.g., metrics for deleted monitors).
func (cj *CleanupJob) CleanupOrphanedData() error {
	// Find metric snapshots with non-existent monitor IDs
	var orphanedCount int64

	// Subquery to find monitor IDs that don't exist
	subQuery := cj.db.Model(&models.Monitor{}).Select("id")

	result := cj.db.Where("monitor_id IS NOT NULL AND monitor_id NOT IN (?)", subQuery).
		Delete(&models.MetricSnapshot{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup orphaned metrics: %w", result.Error)
	}

	orphanedCount = result.RowsAffected

	// Find anomalies with non-existent monitor IDs
	result = cj.db.Where("monitor_id IS NOT NULL AND monitor_id NOT IN (?)", subQuery).
		Delete(&models.DetectedAnomaly{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup orphaned anomalies: %w", result.Error)
	}

	orphanedCount += result.RowsAffected

	// Find baselines with non-existent monitor IDs
	result = cj.db.Where("monitor_id IS NOT NULL AND monitor_id NOT IN (?)", subQuery).
		Delete(&models.AnomalyBaseline{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup orphaned baselines: %w", result.Error)
	}

	orphanedCount += result.RowsAffected

	if orphanedCount > 0 {
		fmt.Printf("[CleanupJob] Cleaned up %d orphaned records\n", orphanedCount)
	}

	return nil
}

// GetStatistics returns cleanup statistics.
func (cj *CleanupJob) GetStatistics() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count total metrics
	var totalMetrics int64
	if err := cj.db.Model(&models.MetricSnapshot{}).Count(&totalMetrics).Error; err != nil {
		return nil, err
	}
	stats["total_metrics"] = totalMetrics

	// Count total baselines
	var totalBaselines int64
	if err := cj.db.Model(&models.AnomalyBaseline{}).Count(&totalBaselines).Error; err != nil {
		return nil, err
	}
	stats["total_baselines"] = totalBaselines

	// Count total anomalies
	var totalAnomalies int64
	if err := cj.db.Model(&models.DetectedAnomaly{}).Count(&totalAnomalies).Error; err != nil {
		return nil, err
	}
	stats["total_anomalies"] = totalAnomalies

	// Count open anomalies
	var openAnomalies int64
	if err := cj.db.Model(&models.DetectedAnomaly{}).
		Where("status = ?", models.StatusOpen).
		Count(&openAnomalies).Error; err != nil {
		return nil, err
	}
	stats["open_anomalies"] = openAnomalies

	// Count resolved anomalies
	var resolvedAnomalies int64
	if err := cj.db.Model(&models.DetectedAnomaly{}).
		Where("status IN ?", []string{models.StatusResolved, models.StatusFalsePositive}).
		Count(&resolvedAnomalies).Error; err != nil {
		return nil, err
	}
	stats["resolved_anomalies"] = resolvedAnomalies

	return stats, nil
}
