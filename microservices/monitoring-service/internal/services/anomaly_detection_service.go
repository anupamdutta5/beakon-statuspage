// Package services provides business logic for anomaly detection.
package services

import (
	"fmt"
	"math"
	"time"

	"github.com/anupamdutta5/monitoring-service/internal/models"
	"gorm.io/gorm"
)

// AnomalyDetectionService handles anomaly detection operations.
type AnomalyDetectionService struct {
	db                 *gorm.DB
	baselineCalculator *BaselineCalculator
}

// NewAnomalyDetectionService creates a new AnomalyDetectionService instance.
func NewAnomalyDetectionService(db *gorm.DB, baselineCalculator *BaselineCalculator) *AnomalyDetectionService {
	return &AnomalyDetectionService{
		db:                 db,
		baselineCalculator: baselineCalculator,
	}
}

// DetectAnomaly performs anomaly detection on a new metric value.
func (ads *AnomalyDetectionService) DetectAnomaly(tenantID uint, monitorID uint, metricType string, metricValue float64, timestamp time.Time) (*models.DetectionResult, error) {
	// Get detection configuration
	config, err := ads.getConfig(tenantID, monitorID, metricType)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	// Check if detection is enabled
	if !config.Enabled {
		return &models.DetectionResult{
			IsAnomaly:      false,
			ActualValue:    metricValue,
			ExpectedValue:  metricValue,
			DeviationScore: 0,
		}, nil
	}

	// Get baseline
	baseline, err := ads.baselineCalculator.GetBaseline(tenantID, monitorID, metricType, timestamp)
	if err != nil {
		// No baseline available yet
		return &models.DetectionResult{
			IsAnomaly:      false,
			ActualValue:    metricValue,
			ExpectedValue:  metricValue,
			DeviationScore: 0,
		}, nil
	}

	// Check minimum sample requirement
	if baseline.SampleCount < config.MinBaselineSamples {
		return &models.DetectionResult{
			IsAnomaly:      false,
			ActualValue:    metricValue,
			ExpectedValue:  baseline.MeanValue,
			DeviationScore: 0,
		}, nil
	}

	// Perform Z-score detection
	result := ads.detectWithZScore(metricValue, baseline, config)

	// If anomaly detected, check consecutive requirement
	if result.IsAnomaly && config.RequireConsecutiveAnomalies > 1 {
		consecutiveCount, err := ads.countConsecutiveAnomalies(tenantID, monitorID, metricType)
		if err != nil {
			return nil, fmt.Errorf("failed to count consecutive anomalies: %w", err)
		}

		// Require N consecutive anomalies before alerting
		if consecutiveCount < config.RequireConsecutiveAnomalies-1 {
			result.IsAnomaly = false // Don't trigger yet
		}
	}

	// Store the anomaly if detected
	if result.IsAnomaly {
		if err := ads.storeAnomaly(tenantID, monitorID, metricType, result, timestamp); err != nil {
			return nil, fmt.Errorf("failed to store anomaly: %w", err)
		}
	}

	return result, nil
}

// detectWithZScore performs Z-score based anomaly detection.
func (ads *AnomalyDetectionService) detectWithZScore(value float64, baseline *models.AnomalyBaseline, config *models.AnomalyDetectionConfig) *models.DetectionResult {
	// Calculate Z-score
	zScore := models.CalculateZScore(value, baseline.MeanValue, baseline.StdDev)

	result := &models.DetectionResult{
		ActualValue:     value,
		ExpectedValue:   baseline.MeanValue,
		DeviationScore:  math.Abs(zScore),
		DetectionMethod: models.DetectionMethodZScore,
	}

	// Determine if it's an anomaly and severity
	if models.IsAnomaly(zScore, config.ZScoreThresholdMinor) {
		result.IsAnomaly = true
		result.Severity = models.DetermineSeverity(zScore, config)
	}

	return result
}

// detectWithPercentile performs percentile-based anomaly detection.
func (ads *AnomalyDetectionService) detectWithPercentile(value float64, baseline *models.AnomalyBaseline) *models.DetectionResult {
	result := &models.DetectionResult{
		ActualValue:     value,
		ExpectedValue:   baseline.MeanValue,
		DetectionMethod: models.DetectionMethodPercentile,
	}

	// Check if value exceeds P99
	if baseline.P99 != nil && value > *baseline.P99 {
		result.IsAnomaly = true
		result.Severity = models.SeverityCritical
		result.DeviationScore = (value - *baseline.P99) / baseline.MeanValue * 100
	} else if baseline.P95 != nil && value > *baseline.P95 {
		result.IsAnomaly = true
		result.Severity = models.SeverityMajor
		result.DeviationScore = (value - *baseline.P95) / baseline.MeanValue * 100
	}

	return result
}

// storeAnomaly stores a detected anomaly in the database.
func (ads *AnomalyDetectionService) storeAnomaly(tenantID uint, monitorID uint, metricType string, result *models.DetectionResult, timestamp time.Time) error {
	anomaly := &models.DetectedAnomaly{
		TenantID:        tenantID,
		MonitorID:       &monitorID,
		MetricType:      metricType,
		DetectedAt:      timestamp,
		ActualValue:     result.ActualValue,
		ExpectedValue:   result.ExpectedValue,
		DeviationScore:  result.DeviationScore,
		Severity:        result.Severity,
		DetectionMethod: result.DetectionMethod,
		Status:          models.StatusOpen,
	}

	if err := anomaly.Validate(); err != nil {
		return fmt.Errorf("anomaly validation failed: %w", err)
	}

	if err := ads.db.Create(anomaly).Error; err != nil {
		return fmt.Errorf("failed to create anomaly record: %w", err)
	}

	return nil
}

// GetAnomalies retrieves anomalies with optional filters.
func (ads *AnomalyDetectionService) GetAnomalies(tenantID uint, filters map[string]interface{}, limit int) ([]models.DetectedAnomaly, error) {
	var anomalies []models.DetectedAnomaly

	query := ads.db.Where("tenant_id = ?", tenantID)

	// Apply filters
	if monitorID, ok := filters["monitor_id"]; ok {
		query = query.Where("monitor_id = ?", monitorID)
	}

	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}

	if severity, ok := filters["severity"]; ok {
		query = query.Where("severity = ?", severity)
	}

	if from, ok := filters["from"]; ok {
		query = query.Where("detected_at >= ?", from)
	}

	if to, ok := filters["to"]; ok {
		query = query.Where("detected_at <= ?", to)
	}

	// Order by most recent
	query = query.Order("detected_at DESC").Limit(limit)

	// Preload monitor details
	query = query.Preload("Monitor")

	if err := query.Find(&anomalies).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch anomalies: %w", err)
	}

	return anomalies, nil
}

// GetAnomalyByID retrieves a specific anomaly with context.
func (ads *AnomalyDetectionService) GetAnomalyByID(tenantID uint, anomalyID uint) (*models.AnomalyContext, error) {
	var anomaly models.DetectedAnomaly

	if err := ads.db.Where("id = ? AND tenant_id = ?", anomalyID, tenantID).
		Preload("Monitor").
		First(&anomaly).Error; err != nil {
		return nil, fmt.Errorf("anomaly not found: %w", err)
	}

	// Get baseline context
	baseline7Day, _ := ads.baselineCalculator.GetBaseline(tenantID, *anomaly.MonitorID, anomaly.MetricType, anomaly.DetectedAt)
	baseline30Day, _ := ads.baselineCalculator.GetBaseline(tenantID, *anomaly.MonitorID, anomaly.MetricType, anomaly.DetectedAt.Add(-23*24*time.Hour))

	// Get historical anomaly count for this monitor
	var historicalCount int64
	ads.db.Model(&models.DetectedAnomaly{}).
		Where("tenant_id = ? AND monitor_id = ? AND id < ?", tenantID, anomaly.MonitorID, anomalyID).
		Count(&historicalCount)

	// Get last anomaly date
	var lastAnomaly models.DetectedAnomaly
	err := ads.db.Where("tenant_id = ? AND monitor_id = ? AND id < ?", tenantID, anomaly.MonitorID, anomalyID).
		Order("detected_at DESC").
		First(&lastAnomaly).Error

	var lastAnomalyDate *time.Time
	if err == nil {
		lastAnomalyDate = &lastAnomaly.DetectedAt
	}

	// Get recent metrics (last 100 points)
	var recentMetrics []models.MetricSnapshot
	ads.db.Where("tenant_id = ? AND monitor_id = ? AND metric_type = ? AND timestamp <= ?",
		tenantID, anomaly.MonitorID, anomaly.MetricType, anomaly.DetectedAt).
		Order("timestamp DESC").
		Limit(100).
		Find(&recentMetrics)

	context := &models.AnomalyContext{
		AnomalyID:              anomaly.ID,
		MonitorID:              *anomaly.MonitorID,
		MonitorName:            "", // TODO: Fetch from UptimeCheck if needed
		HistoricalAnomalyCount: int(historicalCount),
		LastAnomalyDate:        lastAnomalyDate,
		RecentMetrics:          recentMetrics,
	}

	if baseline7Day != nil {
		context.Baseline7Day = baseline7Day.MeanValue
	}

	if baseline30Day != nil {
		context.Baseline30Day = baseline30Day.MeanValue
	}

	return context, nil
}

// AcknowledgeAnomaly marks an anomaly as acknowledged.
func (ads *AnomalyDetectionService) AcknowledgeAnomaly(tenantID uint, anomalyID uint, acknowledgedBy uint, notes string) error {
	now := time.Now()

	result := ads.db.Model(&models.DetectedAnomaly{}).
		Where("id = ? AND tenant_id = ? AND status = ?", anomalyID, tenantID, models.StatusOpen).
		Updates(map[string]interface{}{
			"status":          models.StatusAcknowledged,
			"acknowledged_at": now,
			"acknowledged_by": acknowledgedBy,
			"resolution_notes": notes,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to acknowledge anomaly: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("anomaly not found or already acknowledged")
	}

	return nil
}

// ResolveAnomaly marks an anomaly as resolved or false positive.
func (ads *AnomalyDetectionService) ResolveAnomaly(tenantID uint, anomalyID uint, resolution string, notes string) error {
	validResolutions := map[string]bool{
		models.StatusResolved:      true,
		models.StatusFalsePositive: true,
	}

	if !validResolutions[resolution] {
		return fmt.Errorf("invalid resolution: %s", resolution)
	}

	now := time.Now()

	result := ads.db.Model(&models.DetectedAnomaly{}).
		Where("id = ? AND tenant_id = ?", anomalyID, tenantID).
		Updates(map[string]interface{}{
			"status":           resolution,
			"resolved_at":      now,
			"resolution_notes": notes,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to resolve anomaly: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("anomaly not found")
	}

	return nil
}

// GetStatistics retrieves aggregated anomaly statistics.
func (ads *AnomalyDetectionService) GetStatistics(tenantID uint, period string) (*models.AnomalyStatistics, error) {
	// Calculate time range based on period
	var since time.Time
	switch period {
	case "24h":
		since = time.Now().Add(-24 * time.Hour)
	case "7d":
		since = time.Now().Add(-7 * 24 * time.Hour)
	case "30d":
		since = time.Now().Add(-30 * 24 * time.Hour)
	default:
		return nil, fmt.Errorf("invalid period: %s", period)
	}

	// Get all anomalies in period
	var anomalies []models.DetectedAnomaly
	if err := ads.db.Where("tenant_id = ? AND detected_at >= ?", tenantID, since).
		Find(&anomalies).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch anomalies: %w", err)
	}

	// Aggregate by severity
	bySeverity := make(map[string]int)
	for _, a := range anomalies {
		bySeverity[a.Severity]++
	}

	// Aggregate by status
	byStatus := make(map[string]int)
	for _, a := range anomalies {
		byStatus[a.Status]++
	}

	// Calculate false positive rate
	totalAnomalies := len(anomalies)
	falsePositives := byStatus[models.StatusFalsePositive]
	falsePositiveRate := 0.0
	if totalAnomalies > 0 {
		falsePositiveRate = float64(falsePositives) / float64(totalAnomalies) * 100
	}

	// Get top affected monitors
	topMonitors, err := ads.getTopAffectedMonitors(tenantID, since, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to get top monitors: %w", err)
	}

	return &models.AnomalyStatistics{
		TenantID:          tenantID,
		Period:            period,
		TotalAnomalies:    totalAnomalies,
		BySeverity:        bySeverity,
		ByStatus:          byStatus,
		FalsePositiveRate: falsePositiveRate,
		TopAffectedMonitors: topMonitors,
	}, nil
}

// getTopAffectedMonitors gets monitors with highest anomaly count.
func (ads *AnomalyDetectionService) getTopAffectedMonitors(tenantID uint, since time.Time, limit int) ([]models.TopAffectedMonitor, error) {
	type MonitorCount struct {
		MonitorID     uint
		AnomalyCount  int
		LastAnomalyAt time.Time
	}

	var results []MonitorCount

	err := ads.db.Model(&models.DetectedAnomaly{}).
		Select("monitor_id, COUNT(*) as anomaly_count, MAX(detected_at) as last_anomaly_at").
		Where("tenant_id = ? AND detected_at >= ? AND monitor_id IS NOT NULL", tenantID, since).
		Group("monitor_id").
		Order("anomaly_count DESC").
		Limit(limit).
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	// Get monitor names
	var topMonitors []models.TopAffectedMonitor
	for _, result := range results {
		var monitor models.Monitor
		if err := ads.db.Where("id = ?", result.MonitorID).First(&monitor).Error; err != nil {
			continue
		}

		topMonitors = append(topMonitors, models.TopAffectedMonitor{
			MonitorID:     result.MonitorID,
			MonitorName:   monitor.Name,
			AnomalyCount:  result.AnomalyCount,
			LastAnomalyAt: result.LastAnomalyAt,
		})
	}

	return topMonitors, nil
}

// GetConfig retrieves or creates detection configuration.
func (ads *AnomalyDetectionService) getConfig(tenantID uint, monitorID uint, metricType string) (*models.AnomalyDetectionConfig, error) {
	var config models.AnomalyDetectionConfig

	// Try specific monitor + metric config
	err := ads.db.Where("tenant_id = ? AND monitor_id = ? AND metric_type = ?",
		tenantID, monitorID, metricType).First(&config).Error

	if err == nil {
		return &config, nil
	}

	// Try monitor-wide config
	err = ads.db.Where("tenant_id = ? AND monitor_id = ? AND metric_type IS NULL",
		tenantID, monitorID).First(&config).Error

	if err == nil {
		return &config, nil
	}

	// Try tenant-wide config
	err = ads.db.Where("tenant_id = ? AND monitor_id IS NULL AND metric_type IS NULL",
		tenantID).First(&config).Error

	if err == nil {
		return &config, nil
	}

	// Create default config
	minor, major, critical := models.GetThresholdsForSensitivity(models.SensitivityMedium)
	defaultConfig := &models.AnomalyDetectionConfig{
		TenantID:                    tenantID,
		Enabled:                     true,
		Sensitivity:                 models.SensitivityMedium,
		MinBaselineSamples:          50,
		ZScoreThresholdMinor:        minor,
		ZScoreThresholdMajor:        major,
		ZScoreThresholdCritical:     critical,
		NotificationEnabled:         true,
		NotificationCooldownMinutes: 30,
		RequireConsecutiveAnomalies: 1,
	}

	if err := ads.db.Create(defaultConfig).Error; err != nil {
		return nil, fmt.Errorf("failed to create default config: %w", err)
	}

	return defaultConfig, nil
}

// UpdateConfig updates anomaly detection configuration.
func (ads *AnomalyDetectionService) UpdateConfig(tenantID uint, config *models.AnomalyDetectionConfig) error {
	config.TenantID = tenantID

	if err := config.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Find existing config
	existing := &models.AnomalyDetectionConfig{}
	query := ads.db.Where("tenant_id = ?", tenantID)

	if config.MonitorID != nil {
		query = query.Where("monitor_id = ?", *config.MonitorID)
	} else {
		query = query.Where("monitor_id IS NULL")
	}

	if config.MetricType != nil {
		query = query.Where("metric_type = ?", *config.MetricType)
	} else {
		query = query.Where("metric_type IS NULL")
	}

	err := query.First(existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new config
		return ads.db.Create(config).Error
	} else if err != nil {
		return err
	}

	// Update existing
	config.ID = existing.ID
	return ads.db.Save(config).Error
}

// countConsecutiveAnomalies counts recent consecutive anomalies.
func (ads *AnomalyDetectionService) countConsecutiveAnomalies(tenantID uint, monitorID uint, metricType string) (int, error) {
	// Get last N anomalies
	var recentAnomalies []models.DetectedAnomaly
	err := ads.db.Where("tenant_id = ? AND monitor_id = ? AND metric_type = ? AND status = ?",
		tenantID, monitorID, metricType, models.StatusOpen).
		Order("detected_at DESC").
		Limit(10).
		Find(&recentAnomalies).Error

	if err != nil {
		return 0, err
	}

	// Count consecutive from most recent
	consecutiveCount := 0
	lastTime := time.Now()

	for _, anomaly := range recentAnomalies {
		// Check if within 5 minutes of previous
		if lastTime.Sub(anomaly.DetectedAt) <= 5*time.Minute {
			consecutiveCount++
			lastTime = anomaly.DetectedAt
		} else {
			break
		}
	}

	return consecutiveCount, nil
}

// ShouldNotify checks if notification should be sent (respects cooldown).
func (ads *AnomalyDetectionService) ShouldNotify(tenantID uint, monitorID uint, metricType string, cooldownMinutes int) (bool, error) {
	// Check last notification time
	var lastAnomaly models.DetectedAnomaly
	err := ads.db.Where("tenant_id = ? AND monitor_id = ? AND metric_type = ? AND status = ?",
		tenantID, monitorID, metricType, models.StatusOpen).
		Order("detected_at DESC").
		First(&lastAnomaly).Error

	if err == gorm.ErrRecordNotFound {
		return true, nil // No recent anomaly, OK to notify
	} else if err != nil {
		return false, err
	}

	// Check if cooldown period has passed
	cooldownDuration := time.Duration(cooldownMinutes) * time.Minute
	timeSinceLastAnomaly := time.Since(lastAnomaly.DetectedAt)

	return timeSinceLastAnomaly >= cooldownDuration, nil
}

// GetDB returns the database instance (for handlers).
func (ads *AnomalyDetectionService) GetDB() *gorm.DB {
	return ads.db
}
