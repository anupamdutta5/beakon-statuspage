// Package services provides business logic for baseline calculation.
package services

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/anupamdutta5/monitoring-service/internal/models"
	"gorm.io/gorm"
)

// BaselineCalculator handles calculation of statistical baselines for anomaly detection.
type BaselineCalculator struct {
	db *gorm.DB
}

// NewBaselineCalculator creates a new BaselineCalculator instance.
func NewBaselineCalculator(db *gorm.DB) *BaselineCalculator {
	return &BaselineCalculator{db: db}
}

// CalculateBaselines calculates all baselines for a tenant's monitors.
func (bc *BaselineCalculator) CalculateBaselines(tenantID uint) error {
	// Get all monitors for tenant
	var monitors []models.Monitor
	if err := bc.db.Where("tenant_id = ?", tenantID).Find(&monitors).Error; err != nil {
		return fmt.Errorf("failed to fetch monitors: %w", err)
	}

	// Calculate baselines for each monitor
	for _, monitor := range monitors {
		if err := bc.CalculateMonitorBaselines(tenantID, monitor.ID); err != nil {
			// Log error but continue with other monitors
			fmt.Printf("Error calculating baselines for monitor %d: %v\n", monitor.ID, err)
			continue
		}
	}

	return nil
}

// CalculateMonitorBaselines calculates all baseline types for a specific monitor.
func (bc *BaselineCalculator) CalculateMonitorBaselines(tenantID uint, monitorID uint) error {
	metricTypes := []string{
		models.MetricTypeResponseTime,
		models.MetricTypeErrorRate,
		models.MetricTypeRequestVolume,
	}

	for _, metricType := range metricTypes {
		// Calculate 7-day baseline
		if err := bc.Calculate7DayBaseline(tenantID, monitorID, metricType); err != nil {
			return fmt.Errorf("failed to calculate 7-day baseline: %w", err)
		}

		// Calculate 30-day baseline
		if err := bc.Calculate30DayBaseline(tenantID, monitorID, metricType); err != nil {
			return fmt.Errorf("failed to calculate 30-day baseline: %w", err)
		}

		// Calculate hourly baselines (24 baselines, one per hour)
		if err := bc.CalculateHourlyBaselines(tenantID, monitorID, metricType); err != nil {
			return fmt.Errorf("failed to calculate hourly baselines: %w", err)
		}
	}

	return nil
}

// Calculate7DayBaseline calculates the 7-day rolling baseline.
func (bc *BaselineCalculator) Calculate7DayBaseline(tenantID uint, monitorID uint, metricType string) error {
	// Fetch last 7 days of metrics
	since := time.Now().Add(-7 * 24 * time.Hour)
	var snapshots []models.MetricSnapshot

	query := bc.db.Where("tenant_id = ? AND metric_type = ? AND timestamp >= ?", tenantID, metricType, since)
	if monitorID > 0 {
		query = query.Where("monitor_id = ?", monitorID)
	}

	if err := query.Order("timestamp ASC").Find(&snapshots).Error; err != nil {
		return fmt.Errorf("failed to fetch metric snapshots: %w", err)
	}

	// Need minimum samples
	if len(snapshots) < 50 {
		return fmt.Errorf("insufficient data: only %d samples (need 50+)", len(snapshots))
	}

	// Calculate statistics
	stats := bc.calculateStatistics(snapshots)

	// Create or update baseline
	baseline := &models.AnomalyBaseline{
		TenantID:     tenantID,
		MonitorID:    &monitorID,
		MetricType:   metricType,
		BaselineType: models.BaselineType7Day,
		MeanValue:    stats.Mean,
		StdDev:       stats.StdDev,
		MinValue:     &stats.Min,
		MaxValue:     &stats.Max,
		P50:          &stats.Median,
		P95:          &stats.P95,
		P99:          &stats.P99,
		SampleCount:  stats.SampleCount,
		CalculatedAt: time.Now(),
	}

	// Set expiration to 6 hours from now
	expiresAt := time.Now().Add(6 * time.Hour)
	baseline.ExpiresAt = &expiresAt

	// Upsert baseline
	if err := bc.upsertBaseline(baseline); err != nil {
		return fmt.Errorf("failed to save baseline: %w", err)
	}

	return nil
}

// Calculate30DayBaseline calculates the 30-day rolling baseline.
func (bc *BaselineCalculator) Calculate30DayBaseline(tenantID uint, monitorID uint, metricType string) error {
	since := time.Now().Add(-30 * 24 * time.Hour)
	var snapshots []models.MetricSnapshot

	query := bc.db.Where("tenant_id = ? AND metric_type = ? AND timestamp >= ?", tenantID, metricType, since)
	if monitorID > 0 {
		query = query.Where("monitor_id = ?", monitorID)
	}

	if err := query.Order("timestamp ASC").Find(&snapshots).Error; err != nil {
		return fmt.Errorf("failed to fetch metric snapshots: %w", err)
	}

	if len(snapshots) < 50 {
		return fmt.Errorf("insufficient data: only %d samples", len(snapshots))
	}

	stats := bc.calculateStatistics(snapshots)

	baseline := &models.AnomalyBaseline{
		TenantID:     tenantID,
		MonitorID:    &monitorID,
		MetricType:   metricType,
		BaselineType: models.BaselineType30Day,
		MeanValue:    stats.Mean,
		StdDev:       stats.StdDev,
		MinValue:     &stats.Min,
		MaxValue:     &stats.Max,
		P50:          &stats.Median,
		P95:          &stats.P95,
		P99:          &stats.P99,
		SampleCount:  stats.SampleCount,
		CalculatedAt: time.Now(),
	}

	expiresAt := time.Now().Add(12 * time.Hour)
	baseline.ExpiresAt = &expiresAt

	if err := bc.upsertBaseline(baseline); err != nil {
		return fmt.Errorf("failed to save baseline: %w", err)
	}

	return nil
}

// CalculateHourlyBaselines calculates hour-of-day baselines (0-23).
func (bc *BaselineCalculator) CalculateHourlyBaselines(tenantID uint, monitorID uint, metricType string) error {
	// Fetch last 7 days of metrics
	since := time.Now().Add(-7 * 24 * time.Hour)
	var snapshots []models.MetricSnapshot

	query := bc.db.Where("tenant_id = ? AND metric_type = ? AND timestamp >= ?", tenantID, metricType, since)
	if monitorID > 0 {
		query = query.Where("monitor_id = ?", monitorID)
	}

	if err := query.Order("timestamp ASC").Find(&snapshots).Error; err != nil {
		return fmt.Errorf("failed to fetch metric snapshots: %w", err)
	}

	// Group snapshots by hour of day
	hourlySnapshots := make(map[int][]models.MetricSnapshot)
	for _, snapshot := range snapshots {
		hour := snapshot.Timestamp.Hour()
		hourlySnapshots[hour] = append(hourlySnapshots[hour], snapshot)
	}

	// Calculate baseline for each hour
	for hour := 0; hour < 24; hour++ {
		hoursnapshots := hourlySnapshots[hour]

		// Need minimum samples per hour
		if len(hoursnapshots) < 5 {
			continue // Skip this hour, not enough data
		}

		stats := bc.calculateStatistics(hoursnapshots)

		baseline := &models.AnomalyBaseline{
			TenantID:     tenantID,
			MonitorID:    &monitorID,
			MetricType:   metricType,
			BaselineType: models.BaselineTypeHourly,
			HourOfDay:    &hour,
			MeanValue:    stats.Mean,
			StdDev:       stats.StdDev,
			MinValue:     &stats.Min,
			MaxValue:     &stats.Max,
			P50:          &stats.Median,
			P95:          &stats.P95,
			P99:          &stats.P99,
			SampleCount:  stats.SampleCount,
			CalculatedAt: time.Now(),
		}

		expiresAt := time.Now().Add(12 * time.Hour)
		baseline.ExpiresAt = &expiresAt

		if err := bc.upsertBaseline(baseline); err != nil {
			return fmt.Errorf("failed to save hourly baseline for hour %d: %w", hour, err)
		}
	}

	return nil
}

// GetBaseline retrieves the most appropriate baseline for detection.
func (bc *BaselineCalculator) GetBaseline(tenantID uint, monitorID uint, metricType string, timestamp time.Time) (*models.AnomalyBaseline, error) {
	// Try hourly baseline first (most specific)
	hour := timestamp.Hour()
	hourlyBaseline := &models.AnomalyBaseline{}
	err := bc.db.Where(
		"tenant_id = ? AND monitor_id = ? AND metric_type = ? AND baseline_type = ? AND hour_of_day = ?",
		tenantID, monitorID, metricType, models.BaselineTypeHourly, hour,
	).First(hourlyBaseline).Error

	if err == nil && !hourlyBaseline.IsExpired() {
		return hourlyBaseline, nil
	}

	// Fall back to 7-day baseline
	sevenDayBaseline := &models.AnomalyBaseline{}
	err = bc.db.Where(
		"tenant_id = ? AND monitor_id = ? AND metric_type = ? AND baseline_type = ?",
		tenantID, monitorID, metricType, models.BaselineType7Day,
	).First(sevenDayBaseline).Error

	if err == nil && !sevenDayBaseline.IsExpired() {
		return sevenDayBaseline, nil
	}

	// Fall back to 30-day baseline
	thirtyDayBaseline := &models.AnomalyBaseline{}
	err = bc.db.Where(
		"tenant_id = ? AND monitor_id = ? AND metric_type = ? AND baseline_type = ?",
		tenantID, monitorID, metricType, models.BaselineType30Day,
	).First(thirtyDayBaseline).Error

	if err == nil && !thirtyDayBaseline.IsExpired() {
		return thirtyDayBaseline, nil
	}

	return nil, fmt.Errorf("no valid baseline found")
}

// calculateStatistics calculates statistical measures from metric snapshots.
func (bc *BaselineCalculator) calculateStatistics(snapshots []models.MetricSnapshot) *models.BaselineStats {
	if len(snapshots) == 0 {
		return &models.BaselineStats{}
	}

	// Extract values
	values := make([]float64, len(snapshots))
	sum := 0.0
	min := math.MaxFloat64
	max := -math.MaxFloat64

	for i, snapshot := range snapshots {
		val := snapshot.MetricValue
		values[i] = val
		sum += val

		if val < min {
			min = val
		}
		if val > max {
			max = val
		}
	}

	// Calculate mean
	mean := sum / float64(len(values))

	// Calculate standard deviation
	varianceSum := 0.0
	for _, val := range values {
		diff := val - mean
		varianceSum += diff * diff
	}
	variance := varianceSum / float64(len(values))
	stdDev := math.Sqrt(variance)

	// Calculate percentiles
	sortedValues := make([]float64, len(values))
	copy(sortedValues, values)
	sort.Float64s(sortedValues)

	median := bc.percentile(sortedValues, 50)
	p95 := bc.percentile(sortedValues, 95)
	p99 := bc.percentile(sortedValues, 99)

	return &models.BaselineStats{
		Mean:        mean,
		StdDev:      stdDev,
		Min:         min,
		Max:         max,
		Median:      median,
		P95:         p95,
		P99:         p99,
		SampleCount: len(values),
		Values:      values,
	}
}

// percentile calculates the percentile value from sorted data.
func (bc *BaselineCalculator) percentile(sortedValues []float64, p float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}

	if len(sortedValues) == 1 {
		return sortedValues[0]
	}

	// Linear interpolation
	rank := (p / 100.0) * float64(len(sortedValues)-1)
	lowerIndex := int(math.Floor(rank))
	upperIndex := int(math.Ceil(rank))

	if lowerIndex == upperIndex {
		return sortedValues[lowerIndex]
	}

	// Interpolate
	lowerValue := sortedValues[lowerIndex]
	upperValue := sortedValues[upperIndex]
	fraction := rank - float64(lowerIndex)

	return lowerValue + fraction*(upperValue-lowerValue)
}

// upsertBaseline creates or updates a baseline record.
func (bc *BaselineCalculator) upsertBaseline(baseline *models.AnomalyBaseline) error {
	// Try to find existing baseline
	existing := &models.AnomalyBaseline{}
	query := bc.db.Where(
		"tenant_id = ? AND metric_type = ? AND baseline_type = ?",
		baseline.TenantID, baseline.MetricType, baseline.BaselineType,
	)

	if baseline.MonitorID != nil {
		query = query.Where("monitor_id = ?", *baseline.MonitorID)
	} else {
		query = query.Where("monitor_id IS NULL")
	}

	if baseline.HourOfDay != nil {
		query = query.Where("hour_of_day = ?", *baseline.HourOfDay)
	} else {
		query = query.Where("hour_of_day IS NULL")
	}

	if baseline.DayOfWeek != nil {
		query = query.Where("day_of_week = ?", *baseline.DayOfWeek)
	} else {
		query = query.Where("day_of_week IS NULL")
	}

	err := query.First(existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new baseline
		return bc.db.Create(baseline).Error
	} else if err != nil {
		return err
	}

	// Update existing baseline
	baseline.ID = existing.ID
	return bc.db.Save(baseline).Error
}

// CleanupOldMetrics removes metric snapshots older than the retention period.
func (bc *BaselineCalculator) CleanupOldMetrics(retentionDays int) error {
	cutoff := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)

	result := bc.db.Where("timestamp < ?", cutoff).Delete(&models.MetricSnapshot{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old metrics: %w", result.Error)
	}

	fmt.Printf("Cleaned up %d old metric snapshots\n", result.RowsAffected)
	return nil
}

// CleanupExpiredBaselines removes expired baseline records.
func (bc *BaselineCalculator) CleanupExpiredBaselines() error {
	now := time.Now()

	result := bc.db.Where("expires_at IS NOT NULL AND expires_at < ?", now).Delete(&models.AnomalyBaseline{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired baselines: %w", result.Error)
	}

	fmt.Printf("Cleaned up %d expired baselines\n", result.RowsAffected)
	return nil
}
