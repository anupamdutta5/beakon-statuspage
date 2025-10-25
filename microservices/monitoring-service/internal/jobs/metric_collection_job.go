// Package jobs provides background job implementations for anomaly detection.
package jobs

import (
	"fmt"
	"time"

	"github.com/anupamdutta5/monitoring-service/internal/models"
	"github.com/anupamdutta5/monitoring-service/internal/services"
	"gorm.io/gorm"
)

// MetricCollectionJob handles periodic collection of metrics from monitor checks.
type MetricCollectionJob struct {
	db                     *gorm.DB
	anomalyDetectionService *services.AnomalyDetectionService
	interval               time.Duration
}

// NewMetricCollectionJob creates a new MetricCollectionJob instance.
func NewMetricCollectionJob(db *gorm.DB, anomalyDetectionService *services.AnomalyDetectionService, interval time.Duration) *MetricCollectionJob {
	return &MetricCollectionJob{
		db:                     db,
		anomalyDetectionService: anomalyDetectionService,
		interval:               interval,
	}
}

// Start begins the metric collection job loop.
func (mcj *MetricCollectionJob) Start() {
	ticker := time.NewTicker(mcj.interval)
	defer ticker.Stop()

	fmt.Printf("[MetricCollectionJob] Started with interval: %v\n", mcj.interval)

	// Run immediately on start
	mcj.collectMetrics()

	for range ticker.C {
		mcj.collectMetrics()
	}
}

// collectMetrics collects metrics from recent monitor checks.
func (mcj *MetricCollectionJob) collectMetrics() {
	fmt.Printf("[MetricCollectionJob] Starting metric collection at %v\n", time.Now())

	// Get all active uptime checks
	var monitors []models.UptimeCheck
	if err := mcj.db.Where("is_active = ?", true).Find(&monitors).Error; err != nil {
		fmt.Printf("[MetricCollectionJob] Error fetching uptime checks: %v\n", err)
		return
	}

	collected := 0
	detected := 0

	for _, monitor := range monitors {
		// Collect metrics for this monitor
		metricsCollected, anomaliesDetected := mcj.collectMonitorMetrics(monitor)
		collected += metricsCollected
		detected += anomaliesDetected
	}

	fmt.Printf("[MetricCollectionJob] Completed: %d metrics collected, %d anomalies detected\n", collected, detected)
}

// collectMonitorMetrics collects metrics for a specific monitor.
func (mcj *MetricCollectionJob) collectMonitorMetrics(monitor models.UptimeCheck) (int, int) {
	metricsCollected := 0
	anomaliesDetected := 0

	// Get the most recent check result for this monitor
	var lastCheck models.UptimeResult
	err := mcj.db.Where("uptime_check_id = ?", monitor.ID).
		Order("created_at DESC").
		First(&lastCheck).Error

	if err != nil {
		// No check results yet, skip
		return 0, 0
	}

	// Only collect metrics from checks in the last 2 minutes (to avoid duplicates)
	if time.Since(lastCheck.CreatedAt) > 2*time.Minute {
		return 0, 0
	}

	// Collect response time metric
	if lastCheck.ResponseTime > 0 {
		snapshot := &models.MetricSnapshot{
			TenantID:    monitor.TenantID,
			MonitorID:   &monitor.ID,
			MetricType:  models.MetricTypeResponseTime,
			MetricValue: float64(lastCheck.ResponseTime),
			Timestamp:   lastCheck.CreatedAt,
		}

		if err := mcj.storeMetricAndDetect(snapshot); err != nil {
			fmt.Printf("[MetricCollectionJob] Error storing response_time metric: %v\n", err)
		} else {
			metricsCollected++
			// Check if anomaly was detected
			if mcj.wasAnomalyDetected(monitor.TenantID, monitor.ID, models.MetricTypeResponseTime, lastCheck.CreatedAt) {
				anomaliesDetected++
			}
		}
	}

	// Collect error rate metric (derived from status)
	errorRate := 0.0
	if lastCheck.Status == "failure" || lastCheck.Status == "timeout" {
		errorRate = 100.0 // 100% error rate if down
	}

	snapshot := &models.MetricSnapshot{
		TenantID:    monitor.TenantID,
		MonitorID:   &monitor.ID,
		MetricType:  models.MetricTypeErrorRate,
		MetricValue: errorRate,
		Timestamp:   lastCheck.CreatedAt,
	}

	if err := mcj.storeMetricAndDetect(snapshot); err != nil {
		fmt.Printf("[MetricCollectionJob] Error storing error_rate metric: %v\n", err)
	} else {
		metricsCollected++
		if mcj.wasAnomalyDetected(monitor.TenantID, monitor.ID, models.MetricTypeErrorRate, lastCheck.CreatedAt) {
			anomaliesDetected++
		}
	}

	return metricsCollected, anomaliesDetected
}

// storeMetricAndDetect stores a metric and performs anomaly detection.
func (mcj *MetricCollectionJob) storeMetricAndDetect(snapshot *models.MetricSnapshot) error {
	// Validate snapshot
	if err := snapshot.Validate(); err != nil {
		return fmt.Errorf("snapshot validation failed: %w", err)
	}

	// Store metric snapshot
	if err := mcj.db.Create(snapshot).Error; err != nil {
		return fmt.Errorf("failed to store metric snapshot: %w", err)
	}

	// Perform anomaly detection
	_, err := mcj.anomalyDetectionService.DetectAnomaly(
		snapshot.TenantID,
		*snapshot.MonitorID,
		snapshot.MetricType,
		snapshot.MetricValue,
		snapshot.Timestamp,
	)

	if err != nil {
		// Log error but don't fail the job
		fmt.Printf("[MetricCollectionJob] Anomaly detection error: %v\n", err)
	}

	return nil
}

// wasAnomalyDetected checks if an anomaly was detected for recent metrics.
func (mcj *MetricCollectionJob) wasAnomalyDetected(tenantID uint, monitorID uint, metricType string, timestamp time.Time) bool {
	var count int64

	mcj.db.Model(&models.DetectedAnomaly{}).
		Where("tenant_id = ? AND monitor_id = ? AND metric_type = ? AND detected_at >= ? AND detected_at <= ?",
			tenantID, monitorID, metricType, timestamp.Add(-1*time.Minute), timestamp.Add(1*time.Minute)).
		Count(&count)

	return count > 0
}
