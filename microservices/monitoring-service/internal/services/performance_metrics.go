package services

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PerformanceMetrics represents detailed performance metrics for a monitor
type PerformanceMetrics struct {
	MonitorID         uint      `json:"monitor_id"`
	TenantID          uuid.UUID `json:"tenant_id"`
	TimeRange         string    `json:"time_range"` // "1h", "24h", "7d", "30d"
	TotalChecks       int       `json:"total_checks"`
	SuccessfulChecks  int       `json:"successful_checks"`
	FailedChecks      int       `json:"failed_checks"`
	UptimePercentage  float64   `json:"uptime_percentage"`

	// Response Time Metrics (milliseconds)
	AvgResponseTime   float64   `json:"avg_response_time_ms"`
	MinResponseTime   int       `json:"min_response_time_ms"`
	MaxResponseTime   int       `json:"max_response_time_ms"`
	MedianResponseTime int      `json:"median_response_time_ms"` // P50
	P50ResponseTime   int       `json:"p50_response_time_ms"`
	P90ResponseTime   int       `json:"p90_response_time_ms"`
	P95ResponseTime   int       `json:"p95_response_time_ms"`
	P99ResponseTime   int       `json:"p99_response_time_ms"`

	// Advanced Performance Metrics
	AvgTTFB           float64   `json:"avg_ttfb_ms"`
	AvgDNSTime        float64   `json:"avg_dns_time_ms"`
	AvgConnectionTime float64   `json:"avg_connection_time_ms"`

	// Time-series data for charts
	ResponseTimeSeries []TimeSeriesPoint `json:"response_time_series,omitempty"`
	UptimeSeries       []TimeSeriesPoint `json:"uptime_series,omitempty"`

	CalculatedAt time.Time `json:"calculated_at"`
}

// TimeSeriesPoint represents a single point in time-series data
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Status    string    `json:"status,omitempty"`
}

// PerformanceMetricsService calculates performance metrics and percentiles
type PerformanceMetricsService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPerformanceMetricsService creates a new performance metrics service
func NewPerformanceMetricsService(db *gorm.DB, logger *zap.Logger) *PerformanceMetricsService {
	return &PerformanceMetricsService{
		db:     db,
		logger: logger,
	}
}

// CalculateMetrics calculates comprehensive performance metrics for a monitor
func (s *PerformanceMetricsService) CalculateMetrics(monitorID uint, tenantID uuid.UUID, timeRange string) (*PerformanceMetrics, error) {
	// Parse time range
	startTime, err := s.parseTimeRange(timeRange)
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	// Fetch monitoring results from database
	var results []MonitoringResult
	err = s.db.Where("monitor_id = ? AND checked_at >= ?", monitorID, startTime).
		Order("checked_at ASC").
		Find(&results).Error

	if err != nil {
		s.logger.Error("Failed to fetch monitoring results", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch results: %w", err)
	}

	if len(results) == 0 {
		return &PerformanceMetrics{
			MonitorID:    monitorID,
			TenantID:     tenantID,
			TimeRange:    timeRange,
			CalculatedAt: time.Now(),
		}, nil
	}

	metrics := &PerformanceMetrics{
		MonitorID:    monitorID,
		TenantID:     tenantID,
		TimeRange:    timeRange,
		TotalChecks:  len(results),
		CalculatedAt: time.Now(),
	}

	// Calculate basic metrics
	s.calculateBasicMetrics(results, metrics)

	// Calculate percentiles
	s.calculatePercentiles(results, metrics)

	// Calculate advanced metrics
	s.calculateAdvancedMetrics(results, metrics)

	// Generate time-series data
	metrics.ResponseTimeSeries = s.generateResponseTimeSeries(results)
	metrics.UptimeSeries = s.generateUptimeSeries(results)

	s.logger.Info("Performance metrics calculated",
		zap.Uint("monitor_id", monitorID),
		zap.String("time_range", timeRange),
		zap.Int("total_checks", metrics.TotalChecks),
		zap.Float64("uptime", metrics.UptimePercentage))

	return metrics, nil
}

// calculateBasicMetrics calculates basic performance metrics
func (s *PerformanceMetricsService) calculateBasicMetrics(results []MonitoringResult, metrics *PerformanceMetrics) {
	var totalResponseTime int64
	minResponseTime := 999999
	maxResponseTime := 0
	successfulChecks := 0

	for _, result := range results {
		if result.Status == "operational" {
			successfulChecks++
		}

		totalResponseTime += int64(result.ResponseTimeMS)

		if result.ResponseTimeMS < minResponseTime {
			minResponseTime = result.ResponseTimeMS
		}
		if result.ResponseTimeMS > maxResponseTime {
			maxResponseTime = result.ResponseTimeMS
		}
	}

	metrics.SuccessfulChecks = successfulChecks
	metrics.FailedChecks = metrics.TotalChecks - successfulChecks
	metrics.UptimePercentage = float64(successfulChecks) / float64(metrics.TotalChecks) * 100
	metrics.AvgResponseTime = float64(totalResponseTime) / float64(metrics.TotalChecks)
	metrics.MinResponseTime = minResponseTime
	metrics.MaxResponseTime = maxResponseTime
}

// calculatePercentiles calculates response time percentiles (P50, P90, P95, P99)
func (s *PerformanceMetricsService) calculatePercentiles(results []MonitoringResult, metrics *PerformanceMetrics) {
	// Extract response times and sort them
	responseTimes := make([]int, len(results))
	for i, result := range results {
		responseTimes[i] = result.ResponseTimeMS
	}
	sort.Ints(responseTimes)

	// Calculate percentiles
	metrics.P50ResponseTime = s.calculatePercentile(responseTimes, 50)
	metrics.MedianResponseTime = metrics.P50ResponseTime // P50 = median
	metrics.P90ResponseTime = s.calculatePercentile(responseTimes, 90)
	metrics.P95ResponseTime = s.calculatePercentile(responseTimes, 95)
	metrics.P99ResponseTime = s.calculatePercentile(responseTimes, 99)
}

// calculatePercentile calculates a specific percentile from sorted data
func (s *PerformanceMetricsService) calculatePercentile(sortedData []int, percentile int) int {
	if len(sortedData) == 0 {
		return 0
	}

	// Calculate index using percentile formula
	index := float64(percentile) / 100.0 * float64(len(sortedData)-1)
	lowerIndex := int(index)
	upperIndex := lowerIndex + 1

	// Handle edge cases
	if upperIndex >= len(sortedData) {
		return sortedData[len(sortedData)-1]
	}

	// Linear interpolation between two nearest values
	fraction := index - float64(lowerIndex)
	return sortedData[lowerIndex] + int(fraction*float64(sortedData[upperIndex]-sortedData[lowerIndex]))
}

// calculateAdvancedMetrics calculates advanced performance metrics (TTFB, DNS, Connection time)
func (s *PerformanceMetricsService) calculateAdvancedMetrics(results []MonitoringResult, metrics *PerformanceMetrics) {
	var totalTTFB int64
	var totalDNS int64
	var totalConnection int64
	validCount := 0

	for _, result := range results {
		if result.TTFBMS > 0 {
			totalTTFB += int64(result.TTFBMS)
			totalDNS += int64(result.DNSTimeMS)
			totalConnection += int64(result.ConnectionTimeMS)
			validCount++
		}
	}

	if validCount > 0 {
		metrics.AvgTTFB = float64(totalTTFB) / float64(validCount)
		metrics.AvgDNSTime = float64(totalDNS) / float64(validCount)
		metrics.AvgConnectionTime = float64(totalConnection) / float64(validCount)
	}
}

// generateResponseTimeSeries generates time-series data for response times
func (s *PerformanceMetricsService) generateResponseTimeSeries(results []MonitoringResult) []TimeSeriesPoint {
	series := make([]TimeSeriesPoint, len(results))

	for i, result := range results {
		series[i] = TimeSeriesPoint{
			Timestamp: result.CheckedAt,
			Value:     float64(result.ResponseTimeMS),
			Status:    result.Status,
		}
	}

	return series
}

// generateUptimeSeries generates time-series data for uptime (1 = up, 0 = down)
func (s *PerformanceMetricsService) generateUptimeSeries(results []MonitoringResult) []TimeSeriesPoint {
	series := make([]TimeSeriesPoint, len(results))

	for i, result := range results {
		value := 0.0
		if result.Status == "operational" {
			value = 1.0
		} else if result.Status == "degraded" {
			value = 0.5
		}

		series[i] = TimeSeriesPoint{
			Timestamp: result.CheckedAt,
			Value:     value,
			Status:    result.Status,
		}
	}

	return series
}

// GetMetricsSummary returns a summary of metrics across multiple time ranges
func (s *PerformanceMetricsService) GetMetricsSummary(monitorID uint, tenantID uuid.UUID) (map[string]*PerformanceMetrics, error) {
	timeRanges := []string{"1h", "24h", "7d", "30d"}
	summary := make(map[string]*PerformanceMetrics)

	for _, timeRange := range timeRanges {
		metrics, err := s.CalculateMetrics(monitorID, tenantID, timeRange)
		if err != nil {
			s.logger.Error("Failed to calculate metrics",
				zap.String("time_range", timeRange),
				zap.Error(err))
			continue
		}

		// Don't include time-series data in summary to keep response small
		metrics.ResponseTimeSeries = nil
		metrics.UptimeSeries = nil

		summary[timeRange] = metrics
	}

	return summary, nil
}

// GetPerformanceTrend analyzes performance trend (improving, degrading, stable)
func (s *PerformanceMetricsService) GetPerformanceTrend(monitorID uint, tenantID uuid.UUID) (map[string]interface{}, error) {
	// Get metrics for different time ranges
	summary, err := s.GetMetricsSummary(monitorID, tenantID)
	if err != nil {
		return nil, err
	}

	// Compare current performance with previous period
	trend := make(map[string]interface{})

	if metrics24h, ok := summary["24h"]; ok {
		if metrics7d, ok := summary["7d"]; ok {
			// Compare 24h vs 7d averages
			uptimeDelta := metrics24h.UptimePercentage - metrics7d.UptimePercentage
			responseTimeDelta := metrics24h.AvgResponseTime - metrics7d.AvgResponseTime

			trend["uptime_delta"] = uptimeDelta
			trend["response_time_delta"] = responseTimeDelta

			// Determine overall trend
			if uptimeDelta > 1 && responseTimeDelta < -50 {
				trend["overall_trend"] = "improving"
			} else if uptimeDelta < -1 || responseTimeDelta > 50 {
				trend["overall_trend"] = "degrading"
			} else {
				trend["overall_trend"] = "stable"
			}

			trend["uptime_trend"] = s.determineTrend(uptimeDelta, 1)
			trend["response_time_trend"] = s.determineTrend(-responseTimeDelta, 50) // Negative because lower is better
		}
	}

	trend["summary"] = summary

	return trend, nil
}

// determineTrend determines if a metric is improving, degrading, or stable
func (s *PerformanceMetricsService) determineTrend(delta, threshold float64) string {
	if delta > threshold {
		return "improving"
	} else if delta < -threshold {
		return "degrading"
	}
	return "stable"
}

// parseTimeRange converts time range string to start time
func (s *PerformanceMetricsService) parseTimeRange(timeRange string) (time.Time, error) {
	now := time.Now()

	switch timeRange {
	case "1h":
		return now.Add(-1 * time.Hour), nil
	case "24h":
		return now.Add(-24 * time.Hour), nil
	case "7d":
		return now.Add(-7 * 24 * time.Hour), nil
	case "30d":
		return now.Add(-30 * 24 * time.Hour), nil
	case "90d":
		return now.Add(-90 * 24 * time.Hour), nil
	default:
		return time.Time{}, fmt.Errorf("invalid time range: %s (valid: 1h, 24h, 7d, 30d, 90d)", timeRange)
	}
}

// GetAggregatedMetrics calculates metrics aggregated across multiple monitors
func (s *PerformanceMetricsService) GetAggregatedMetrics(monitorIDs []uint, tenantID uuid.UUID, timeRange string) (*PerformanceMetrics, error) {
	startTime, err := s.parseTimeRange(timeRange)
	if err != nil {
		return nil, err
	}

	// Fetch results for all monitors
	var results []MonitoringResult
	err = s.db.Where("monitor_id IN ? AND checked_at >= ?", monitorIDs, startTime).
		Order("checked_at ASC").
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch aggregated results: %w", err)
	}

	if len(results) == 0 {
		return &PerformanceMetrics{
			TenantID:     tenantID,
			TimeRange:    timeRange,
			CalculatedAt: time.Now(),
		}, nil
	}

	metrics := &PerformanceMetrics{
		TenantID:     tenantID,
		TimeRange:    timeRange,
		TotalChecks:  len(results),
		CalculatedAt: time.Now(),
	}

	s.calculateBasicMetrics(results, metrics)
	s.calculatePercentiles(results, metrics)
	s.calculateAdvancedMetrics(results, metrics)

	return metrics, nil
}
