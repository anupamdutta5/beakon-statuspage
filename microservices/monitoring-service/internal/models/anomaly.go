// Package models provides data models for anomaly detection in Monitoring Service.
package models

import (
	"fmt"
	"time"
)

// MetricSnapshot represents a time-series metric data point.
type MetricSnapshot struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	TenantID    uint      `gorm:"not null;index" json:"tenant_id"`
	MonitorID   *uint     `gorm:"index" json:"monitor_id,omitempty"`
	MetricType  string    `gorm:"type:varchar(50);not null;index" json:"metric_type"`
	MetricValue float64   `gorm:"not null" json:"metric_value"`
	Timestamp   time.Time `gorm:"not null;index:,sort:desc" json:"timestamp"`
	CreatedAt   time.Time `json:"created_at"`
}

// TableName returns the table name for MetricSnapshot.
func (MetricSnapshot) TableName() string {
	return "metric_snapshots"
}

// Validate performs validation on MetricSnapshot.
func (m *MetricSnapshot) Validate() error {
	if m.TenantID == 0 {
		return fmt.Errorf("tenant_id is required")
	}

	validMetricTypes := map[string]bool{
		"response_time":   true,
		"error_rate":      true,
		"request_volume":  true,
		"status_changes":  true,
	}

	if !validMetricTypes[m.MetricType] {
		return fmt.Errorf("invalid metric_type: %s", m.MetricType)
	}

	if m.MetricValue < 0 {
		return fmt.Errorf("metric_value cannot be negative")
	}

	return nil
}

// AnomalyBaseline represents calculated baseline statistics for a metric.
type AnomalyBaseline struct {
	ID           uint   `gorm:"primarykey" json:"id"`
	TenantID     uint   `gorm:"not null;index" json:"tenant_id"`
	MonitorID    *uint  `gorm:"index" json:"monitor_id,omitempty"`
	MetricType   string `gorm:"type:varchar(50);not null;index" json:"metric_type"`
	BaselineType string `gorm:"type:varchar(20);not null" json:"baseline_type"` // '7day', '30day', 'hourly', 'daily', 'weekly'
	HourOfDay    *int   `json:"hour_of_day,omitempty"`                          // 0-23 for hourly baselines
	DayOfWeek    *int   `json:"day_of_week,omitempty"`                          // 0-6 for weekly baselines

	// Statistical measures
	MeanValue   float64  `gorm:"not null" json:"mean_value"`
	StdDev      float64  `gorm:"not null" json:"std_dev"`
	MinValue    *float64 `json:"min_value,omitempty"`
	MaxValue    *float64 `json:"max_value,omitempty"`
	P50         *float64 `json:"p50,omitempty"`  // Median
	P95         *float64 `json:"p95,omitempty"`
	P99         *float64 `json:"p99,omitempty"`

	// Metadata
	SampleCount  int        `gorm:"not null" json:"sample_count"`
	CalculatedAt time.Time  `json:"calculated_at"`
	ExpiresAt    *time.Time `gorm:"index" json:"expires_at,omitempty"`
}

// TableName returns the table name for AnomalyBaseline.
func (AnomalyBaseline) TableName() string {
	return "anomaly_baselines"
}

// Validate performs validation on AnomalyBaseline.
func (a *AnomalyBaseline) Validate() error {
	if a.TenantID == 0 {
		return fmt.Errorf("tenant_id is required")
	}

	validBaselineTypes := map[string]bool{
		"7day":   true,
		"30day":  true,
		"hourly": true,
		"daily":  true,
		"weekly": true,
	}

	if !validBaselineTypes[a.BaselineType] {
		return fmt.Errorf("invalid baseline_type: %s", a.BaselineType)
	}

	if a.HourOfDay != nil && (*a.HourOfDay < 0 || *a.HourOfDay > 23) {
		return fmt.Errorf("hour_of_day must be between 0 and 23")
	}

	if a.DayOfWeek != nil && (*a.DayOfWeek < 0 || *a.DayOfWeek > 6) {
		return fmt.Errorf("day_of_week must be between 0 and 6")
	}

	if a.SampleCount < 1 {
		return fmt.Errorf("sample_count must be at least 1")
	}

	if a.StdDev < 0 {
		return fmt.Errorf("std_dev cannot be negative")
	}

	return nil
}

// IsExpired checks if the baseline has expired and needs recalculation.
func (a *AnomalyBaseline) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

// DetectedAnomaly represents a detected anomaly record.
type DetectedAnomaly struct {
	ID         uint   `gorm:"primarykey" json:"id"`
	TenantID   uint   `gorm:"not null;index" json:"tenant_id"`
	MonitorID  *uint  `gorm:"index" json:"monitor_id,omitempty"`
	MetricType string `gorm:"type:varchar(50);not null" json:"metric_type"`

	// Timestamp and values
	DetectedAt     time.Time `gorm:"not null;index:,sort:desc" json:"detected_at"`
	ActualValue    float64   `gorm:"not null" json:"actual_value"`
	ExpectedValue  float64   `gorm:"not null" json:"expected_value"`
	DeviationScore float64   `gorm:"not null" json:"deviation_score"` // Z-score

	// Classification
	Severity         string `gorm:"type:varchar(20);not null;index" json:"severity"` // 'minor', 'major', 'critical'
	DetectionMethod  string `gorm:"type:varchar(50);default:z_score" json:"detection_method"`

	// Status tracking
	Status          string     `gorm:"type:varchar(20);not null;default:open;index" json:"status"` // 'open', 'acknowledged', 'resolved', 'false_positive'
	AcknowledgedAt  *time.Time `json:"acknowledged_at,omitempty"`
	AcknowledgedBy  *uint      `json:"acknowledged_by,omitempty"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	ResolutionNotes string     `gorm:"type:text" json:"resolution_notes,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for DetectedAnomaly.
func (DetectedAnomaly) TableName() string {
	return "detected_anomalies"
}

// Validate performs validation on DetectedAnomaly.
func (d *DetectedAnomaly) Validate() error {
	if d.TenantID == 0 {
		return fmt.Errorf("tenant_id is required")
	}

	validSeverities := map[string]bool{
		"minor":    true,
		"major":    true,
		"critical": true,
	}

	if !validSeverities[d.Severity] {
		return fmt.Errorf("invalid severity: %s", d.Severity)
	}

	validStatuses := map[string]bool{
		"open":           true,
		"acknowledged":   true,
		"resolved":       true,
		"false_positive": true,
	}

	if !validStatuses[d.Status] {
		return fmt.Errorf("invalid status: %s", d.Status)
	}

	return nil
}

// IsOpen checks if the anomaly is still open.
func (d *DetectedAnomaly) IsOpen() bool {
	return d.Status == "open"
}

// IsResolved checks if the anomaly has been resolved.
func (d *DetectedAnomaly) IsResolved() bool {
	return d.Status == "resolved" || d.Status == "false_positive"
}

// AnomalyDetectionConfig represents configuration for anomaly detection.
type AnomalyDetectionConfig struct {
	ID         uint    `gorm:"primarykey" json:"id"`
	TenantID   uint    `gorm:"not null;index" json:"tenant_id"`
	MonitorID  *uint   `gorm:"index" json:"monitor_id,omitempty"`
	MetricType *string `gorm:"type:varchar(50)" json:"metric_type,omitempty"`

	// Enable/disable
	Enabled bool `gorm:"default:true;index" json:"enabled"`

	// Sensitivity settings
	Sensitivity         string `gorm:"type:varchar(20);default:medium" json:"sensitivity"` // 'low', 'medium', 'high'
	MinBaselineSamples  int    `gorm:"default:50" json:"min_baseline_samples"`

	// Z-score thresholds
	ZScoreThresholdMinor    float64 `gorm:"default:2.0" json:"z_score_threshold_minor"`
	ZScoreThresholdMajor    float64 `gorm:"default:3.0" json:"z_score_threshold_major"`
	ZScoreThresholdCritical float64 `gorm:"default:4.0" json:"z_score_threshold_critical"`

	// Notification settings
	NotificationEnabled           bool `gorm:"default:true" json:"notification_enabled"`
	NotificationCooldownMinutes   int  `gorm:"default:30" json:"notification_cooldown_minutes"`
	RequireConsecutiveAnomalies   int  `gorm:"default:1" json:"require_consecutive_anomalies"`

	// Integration
	EscalationPolicyID *uint `json:"escalation_policy_id,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	EscalationPolicy *EscalationPolicy `gorm:"foreignKey:EscalationPolicyID" json:"escalation_policy,omitempty"`
}

// TableName returns the table name for AnomalyDetectionConfig.
func (AnomalyDetectionConfig) TableName() string {
	return "anomaly_detection_config"
}

// Validate performs validation on AnomalyDetectionConfig.
func (a *AnomalyDetectionConfig) Validate() error {
	if a.TenantID == 0 {
		return fmt.Errorf("tenant_id is required")
	}

	validSensitivities := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
	}

	if !validSensitivities[a.Sensitivity] {
		return fmt.Errorf("invalid sensitivity: %s", a.Sensitivity)
	}

	if a.MinBaselineSamples < 10 {
		return fmt.Errorf("min_baseline_samples must be at least 10")
	}

	if a.ZScoreThresholdMinor < 1.0 || a.ZScoreThresholdMinor > 5.0 {
		return fmt.Errorf("z_score_threshold_minor must be between 1.0 and 5.0")
	}

	if a.ZScoreThresholdMajor <= a.ZScoreThresholdMinor {
		return fmt.Errorf("z_score_threshold_major must be greater than minor")
	}

	if a.ZScoreThresholdCritical <= a.ZScoreThresholdMajor {
		return fmt.Errorf("z_score_threshold_critical must be greater than major")
	}

	return nil
}

// GetThresholdsForSensitivity returns Z-score thresholds based on sensitivity level.
func GetThresholdsForSensitivity(sensitivity string) (minor, major, critical float64) {
	switch sensitivity {
	case "low":
		return 3.0, 4.0, 5.0
	case "medium":
		return 2.0, 3.0, 4.0
	case "high":
		return 1.5, 2.5, 3.5
	default:
		return 2.0, 3.0, 4.0
	}
}

// AnomalyStatistics represents aggregated anomaly statistics.
type AnomalyStatistics struct {
	TenantID            uint                 `json:"tenant_id"`
	Period              string               `json:"period"` // '24h', '7d', '30d'
	TotalAnomalies      int                  `json:"total_anomalies"`
	BySeverity          map[string]int       `json:"by_severity"`
	ByStatus            map[string]int       `json:"by_status"`
	FalsePositiveRate   float64              `json:"false_positive_rate"`
	TopAffectedMonitors []TopAffectedMonitor `json:"top_affected_monitors"`
}

// TopAffectedMonitor represents a monitor with high anomaly count.
type TopAffectedMonitor struct {
	MonitorID   uint   `json:"monitor_id"`
	MonitorName string `json:"monitor_name"`
	AnomalyCount int   `json:"anomaly_count"`
	LastAnomalyAt time.Time `json:"last_anomaly_at"`
}

// AnomalyContext represents contextual information about an anomaly.
type AnomalyContext struct {
	AnomalyID             uint      `json:"anomaly_id"`
	MonitorID             uint      `json:"monitor_id"`
	MonitorName           string    `json:"monitor_name"`
	Baseline7Day          float64   `json:"baseline_7day_mean"`
	Baseline30Day         float64   `json:"baseline_30day_mean"`
	HistoricalAnomalyCount int      `json:"historical_anomaly_count"`
	LastAnomalyDate       *time.Time `json:"last_anomaly_date,omitempty"`
	RecentMetrics         []MetricSnapshot `json:"recent_metrics"`
}

// DetectionResult represents the result of anomaly detection.
type DetectionResult struct {
	IsAnomaly       bool    `json:"is_anomaly"`
	Severity        string  `json:"severity,omitempty"`
	DeviationScore  float64 `json:"deviation_score"`
	ActualValue     float64 `json:"actual_value"`
	ExpectedValue   float64 `json:"expected_value"`
	DetectionMethod string  `json:"detection_method"`
}

// BaselineStats represents statistical measures for baseline calculation.
type BaselineStats struct {
	Mean        float64   `json:"mean"`
	StdDev      float64   `json:"std_dev"`
	Min         float64   `json:"min"`
	Max         float64   `json:"max"`
	Median      float64   `json:"median"`
	P95         float64   `json:"p95"`
	P99         float64   `json:"p99"`
	SampleCount int       `json:"sample_count"`
	Values      []float64 `json:"-"` // Raw values for percentile calculation
}

// CalculateZScore calculates the Z-score for a value given mean and standard deviation.
func CalculateZScore(value, mean, stdDev float64) float64 {
	if stdDev == 0 {
		return 0
	}
	return (value - mean) / stdDev
}

// DetermineSeverity determines severity based on Z-score and thresholds.
func DetermineSeverity(zScore float64, config *AnomalyDetectionConfig) string {
	absZScore := absFloat(zScore)

	if absZScore >= config.ZScoreThresholdCritical {
		return "critical"
	} else if absZScore >= config.ZScoreThresholdMajor {
		return "major"
	} else if absZScore >= config.ZScoreThresholdMinor {
		return "minor"
	}

	return ""
}

// IsAnomaly checks if a Z-score indicates an anomaly.
func IsAnomaly(zScore float64, minorThreshold float64) bool {
	return absFloat(zScore) >= minorThreshold
}

// Helper function for absolute value
func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// MetricType constants
const (
	MetricTypeResponseTime  = "response_time"
	MetricTypeErrorRate     = "error_rate"
	MetricTypeRequestVolume = "request_volume"
	MetricTypeStatusChanges = "status_changes"
)

// BaselineType constants
const (
	BaselineType7Day   = "7day"
	BaselineType30Day  = "30day"
	BaselineTypeHourly = "hourly"
	BaselineTypeDaily  = "daily"
	BaselineTypeWeekly = "weekly"
)

// Severity constants
const (
	SeverityMinor    = "minor"
	SeverityMajor    = "major"
	SeverityCritical = "critical"
)

// Status constants
const (
	StatusOpen          = "open"
	StatusAcknowledged  = "acknowledged"
	StatusResolved      = "resolved"
	StatusFalsePositive = "false_positive"
)

// Sensitivity constants
const (
	SensitivityLow    = "low"
	SensitivityMedium = "medium"
	SensitivityHigh   = "high"
)

// DetectionMethod constants
const (
	DetectionMethodZScore     = "z_score"
	DetectionMethodEWMA       = "ewma"
	DetectionMethodPercentile = "percentile"
	DetectionMethodSeasonal   = "seasonal"
)
