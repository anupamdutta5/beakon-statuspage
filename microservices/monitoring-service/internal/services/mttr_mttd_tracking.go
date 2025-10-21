package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IncidentTracking represents a tracked incident for MTTR/MTTD calculation
type IncidentTracking struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	TenantID            uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	MonitorID           uint      `gorm:"not null;index" json:"monitor_id"`
	IncidentKey         string    `gorm:"size:255;not null;uniqueIndex" json:"incident_key"` // Unique identifier

	// Incident Lifecycle Timestamps
	IncidentStartTime   time.Time  `gorm:"not null;index" json:"incident_start_time"` // When monitor first failed
	FirstDetectionTime  *time.Time `json:"first_detection_time"` // When system detected the failure
	FirstAlertTime      *time.Time `json:"first_alert_time"` // When first alert was sent
	AcknowledgedTime    *time.Time `json:"acknowledged_time"` // When someone acknowledged the incident
	InvestigationStart  *time.Time `json:"investigation_start_time"` // When investigation began
	ResolutionStartTime *time.Time `json:"resolution_start_time"` // When fix was started
	IncidentEndTime     *time.Time `gorm:"index" json:"incident_end_time"` // When monitor recovered
	VerifiedTime        *time.Time `json:"verified_time"` // When recovery was verified

	// Status
	Status              string    `gorm:"size:50;not null;index" json:"status"` // open, acknowledged, investigating, resolving, resolved, closed
	Severity            string    `gorm:"size:50" json:"severity"` // critical, high, medium, low

	// Calculated Metrics (in minutes)
	MTTD                float64   `json:"mttd"` // Mean Time To Detect
	MTTA                float64   `json:"mtta"` // Mean Time To Acknowledge
	MTTI                float64   `json:"mtti"` // Mean Time To Investigate
	MTTR                float64   `json:"mttr"` // Mean Time To Resolve
	MTTV                float64   `json:"mttv"` // Mean Time To Verify
	TotalDowntime       float64   `json:"total_downtime"` // Total downtime in minutes

	// Metadata
	ImpactedLocations   string    `gorm:"type:text" json:"impacted_locations"` // JSON array of location IDs
	FailureType         string    `gorm:"size:255" json:"failure_type"` // timeout, connection_error, status_code, etc.
	RootCause           string    `gorm:"type:text" json:"root_cause"`
	Resolution          string    `gorm:"type:text" json:"resolution"`
	Notes               string    `gorm:"type:text" json:"notes"`

	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// MetricsSnapshot represents a snapshot of MTTR/MTTD metrics for a period
type MetricsSnapshot struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	TenantID             uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	MonitorID            uint      `gorm:"index" json:"monitor_id"` // 0 for tenant-wide
	PeriodStart          time.Time `gorm:"not null;index" json:"period_start"`
	PeriodEnd            time.Time `gorm:"not null;index" json:"period_end"`
	PeriodType           string    `gorm:"size:50" json:"period_type"` // hourly, daily, weekly, monthly

	// Incident Counts
	TotalIncidents       int       `json:"total_incidents"`
	CriticalIncidents    int       `json:"critical_incidents"`
	HighIncidents        int       `json:"high_incidents"`
	MediumIncidents      int       `json:"medium_incidents"`
	LowIncidents         int       `json:"low_incidents"`

	// Average Metrics (in minutes)
	AvgMTTD              float64   `json:"avg_mttd"`
	AvgMTTA              float64   `json:"avg_mtta"`
	AvgMTTI              float64   `json:"avg_mtti"`
	AvgMTTR              float64   `json:"avg_mttr"`
	AvgMTTV              float64   `json:"avg_mttv"`
	AvgTotalDowntime     float64   `json:"avg_total_downtime"`

	// Percentiles (in minutes)
	P50MTTR              float64   `json:"p50_mttr"`
	P90MTTR              float64   `json:"p90_mttr"`
	P95MTTR              float64   `json:"p95_mttr"`
	P99MTTR              float64   `json:"p99_mttr"`

	// Trend Indicators
	TrendDirection       string    `gorm:"size:50" json:"trend_direction"` // improving, degrading, stable
	TrendPercent         float64   `json:"trend_percent"` // Compared to previous period

	GeneratedAt          time.Time `json:"generated_at"`
	CreatedAt            time.Time `json:"created_at"`
}

// MTTRMTTDTrackingService manages incident tracking and MTTR/MTTD metrics
type MTTRMTTDTrackingService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMTTRMTTDTrackingService creates a new MTTR/MTTD tracking service
func NewMTTRMTTDTrackingService(db *gorm.DB, logger *zap.Logger) *MTTRMTTDTrackingService {
	return &MTTRMTTDTrackingService{
		db:     db,
		logger: logger,
	}
}

// StartIncident starts tracking a new incident
func (s *MTTRMTTDTrackingService) StartIncident(monitorID uint, tenantID uuid.UUID, severity string, failureType string) (*IncidentTracking, error) {
	now := time.Now()
	incidentKey := fmt.Sprintf("monitor-%d-%d", monitorID, now.Unix())

	incident := &IncidentTracking{
		TenantID:           tenantID,
		MonitorID:          monitorID,
		IncidentKey:        incidentKey,
		IncidentStartTime:  now,
		FirstDetectionTime: &now, // Detected immediately
		Status:             "open",
		Severity:           severity,
		FailureType:        failureType,
	}

	// Calculate MTTD (detection time - start time)
	incident.MTTD = incident.FirstDetectionTime.Sub(incident.IncidentStartTime).Minutes()

	err := s.db.Create(incident).Error
	if err != nil {
		s.logger.Error("Failed to start incident tracking", zap.Error(err))
		return nil, fmt.Errorf("failed to start incident: %w", err)
	}

	s.logger.Info("Incident tracking started",
		zap.Uint("id", incident.ID),
		zap.Uint("monitor_id", monitorID),
		zap.String("severity", severity))

	return incident, nil
}

// SendAlert records when an alert was sent
func (s *MTTRMTTDTrackingService) SendAlert(incidentID uint) error {
	now := time.Now()

	result := s.db.Model(&IncidentTracking{}).Where("id = ?", incidentID).Updates(map[string]interface{}{
		"first_alert_time": now,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to update alert time: %w", result.Error)
	}

	s.logger.Info("Alert time recorded", zap.Uint("incident_id", incidentID))
	return nil
}

// AcknowledgeIncident records when an incident was acknowledged
func (s *MTTRMTTDTrackingService) AcknowledgeIncident(incidentID uint) error {
	now := time.Now()

	var incident IncidentTracking
	err := s.db.First(&incident, incidentID).Error
	if err != nil {
		return fmt.Errorf("incident not found: %w", err)
	}

	// Calculate MTTA (acknowledge time - detection time)
	mtta := 0.0
	if incident.FirstDetectionTime != nil {
		mtta = now.Sub(*incident.FirstDetectionTime).Minutes()
	}

	result := s.db.Model(&IncidentTracking{}).Where("id = ?", incidentID).Updates(map[string]interface{}{
		"acknowledged_time": now,
		"status":            "acknowledged",
		"mtta":              mtta,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to acknowledge incident: %w", result.Error)
	}

	s.logger.Info("Incident acknowledged",
		zap.Uint("incident_id", incidentID),
		zap.Float64("mtta_minutes", mtta))

	return nil
}

// StartInvestigation records when investigation started
func (s *MTTRMTTDTrackingService) StartInvestigation(incidentID uint) error {
	now := time.Now()

	var incident IncidentTracking
	err := s.db.First(&incident, incidentID).Error
	if err != nil {
		return fmt.Errorf("incident not found: %w", err)
	}

	// Calculate MTTI (investigation start - detection time)
	mtti := 0.0
	if incident.FirstDetectionTime != nil {
		mtti = now.Sub(*incident.FirstDetectionTime).Minutes()
	}

	result := s.db.Model(&IncidentTracking{}).Where("id = ?", incidentID).Updates(map[string]interface{}{
		"investigation_start": now,
		"status":              "investigating",
		"mtti":                mtti,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to start investigation: %w", result.Error)
	}

	s.logger.Info("Investigation started",
		zap.Uint("incident_id", incidentID),
		zap.Float64("mtti_minutes", mtti))

	return nil
}

// StartResolution records when resolution work started
func (s *MTTRMTTDTrackingService) StartResolution(incidentID uint) error {
	now := time.Now()

	result := s.db.Model(&IncidentTracking{}).Where("id = ?", incidentID).Updates(map[string]interface{}{
		"resolution_start_time": now,
		"status":                "resolving",
	})

	if result.Error != nil {
		return fmt.Errorf("failed to start resolution: %w", result.Error)
	}

	s.logger.Info("Resolution started", zap.Uint("incident_id", incidentID))
	return nil
}

// ResolveIncident records when an incident was resolved
func (s *MTTRMTTDTrackingService) ResolveIncident(incidentID uint, rootCause, resolution string) error {
	now := time.Now()

	var incident IncidentTracking
	err := s.db.First(&incident, incidentID).Error
	if err != nil {
		return fmt.Errorf("incident not found: %w", err)
	}

	// Calculate MTTR (resolve time - start time)
	mttr := now.Sub(incident.IncidentStartTime).Minutes()

	// Calculate total downtime
	totalDowntime := now.Sub(incident.IncidentStartTime).Minutes()

	result := s.db.Model(&IncidentTracking{}).Where("id = ?", incidentID).Updates(map[string]interface{}{
		"incident_end_time": now,
		"status":            "resolved",
		"mttr":              mttr,
		"total_downtime":    totalDowntime,
		"root_cause":        rootCause,
		"resolution":        resolution,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to resolve incident: %w", result.Error)
	}

	s.logger.Info("Incident resolved",
		zap.Uint("incident_id", incidentID),
		zap.Float64("mttr_minutes", mttr),
		zap.Float64("downtime_minutes", totalDowntime))

	return nil
}

// VerifyRecovery records when recovery was verified
func (s *MTTRMTTDTrackingService) VerifyRecovery(incidentID uint) error {
	now := time.Now()

	var incident IncidentTracking
	err := s.db.First(&incident, incidentID).Error
	if err != nil {
		return fmt.Errorf("incident not found: %w", err)
	}

	// Calculate MTTV (verify time - resolve time)
	mttv := 0.0
	if incident.IncidentEndTime != nil {
		mttv = now.Sub(*incident.IncidentEndTime).Minutes()
	}

	result := s.db.Model(&IncidentTracking{}).Where("id = ?", incidentID).Updates(map[string]interface{}{
		"verified_time": now,
		"status":        "closed",
		"mttv":          mttv,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to verify recovery: %w", result.Error)
	}

	s.logger.Info("Recovery verified",
		zap.Uint("incident_id", incidentID),
		zap.Float64("mttv_minutes", mttv))

	return nil
}

// GetIncident retrieves an incident by ID
func (s *MTTRMTTDTrackingService) GetIncident(id uint) (*IncidentTracking, error) {
	var incident IncidentTracking
	err := s.db.First(&incident, id).Error
	if err != nil {
		return nil, fmt.Errorf("incident not found: %w", err)
	}
	return &incident, nil
}

// GetIncidentsByMonitor retrieves incidents for a monitor
func (s *MTTRMTTDTrackingService) GetIncidentsByMonitor(monitorID uint, limit int) ([]IncidentTracking, error) {
	var incidents []IncidentTracking
	err := s.db.Where("monitor_id = ?", monitorID).
		Order("incident_start_time DESC").
		Limit(limit).
		Find(&incidents).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get incidents: %w", err)
	}

	return incidents, nil
}

// GetOpenIncidents retrieves all open incidents
func (s *MTTRMTTDTrackingService) GetOpenIncidents(tenantID uuid.UUID) ([]IncidentTracking, error) {
	var incidents []IncidentTracking
	err := s.db.Where("tenant_id = ? AND status IN ?", tenantID, []string{"open", "acknowledged", "investigating", "resolving"}).
		Order("incident_start_time DESC").
		Find(&incidents).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get open incidents: %w", err)
	}

	return incidents, nil
}

// CalculateMetrics calculates MTTR/MTTD metrics for a period
func (s *MTTRMTTDTrackingService) CalculateMetrics(monitorID uint, tenantID uuid.UUID, periodStart, periodEnd time.Time, periodType string) (*MetricsSnapshot, error) {
	// Fetch incidents for the period
	var incidents []IncidentTracking
	query := s.db.Where("tenant_id = ? AND incident_start_time BETWEEN ? AND ?", tenantID, periodStart, periodEnd)

	if monitorID > 0 {
		query = query.Where("monitor_id = ?", monitorID)
	}

	err := query.Find(&incidents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch incidents: %w", err)
	}

	if len(incidents) == 0 {
		return &MetricsSnapshot{
			TenantID:    tenantID,
			MonitorID:   monitorID,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			PeriodType:  periodType,
			GeneratedAt: time.Now(),
		}, nil
	}

	// Calculate metrics
	snapshot := &MetricsSnapshot{
		TenantID:    tenantID,
		MonitorID:   monitorID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		PeriodType:  periodType,
	}

	var totalMTTD, totalMTTA, totalMTTI, totalMTTR, totalMTTV, totalDowntime float64
	var mttrValues []float64

	for _, incident := range incidents {
		snapshot.TotalIncidents++

		// Count by severity
		switch incident.Severity {
		case "critical":
			snapshot.CriticalIncidents++
		case "high":
			snapshot.HighIncidents++
		case "medium":
			snapshot.MediumIncidents++
		case "low":
			snapshot.LowIncidents++
		}

		// Sum metrics
		totalMTTD += incident.MTTD
		totalMTTA += incident.MTTA
		totalMTTI += incident.MTTI
		totalMTTR += incident.MTTR
		totalMTTV += incident.MTTV
		totalDowntime += incident.TotalDowntime

		if incident.MTTR > 0 {
			mttrValues = append(mttrValues, incident.MTTR)
		}
	}

	// Calculate averages
	count := float64(snapshot.TotalIncidents)
	snapshot.AvgMTTD = totalMTTD / count
	snapshot.AvgMTTA = totalMTTA / count
	snapshot.AvgMTTI = totalMTTI / count
	snapshot.AvgMTTR = totalMTTR / count
	snapshot.AvgMTTV = totalMTTV / count
	snapshot.AvgTotalDowntime = totalDowntime / count

	// Calculate MTTR percentiles
	if len(mttrValues) > 0 {
		snapshot.P50MTTR = calculatePercentileFloat(mttrValues, 50)
		snapshot.P90MTTR = calculatePercentileFloat(mttrValues, 90)
		snapshot.P95MTTR = calculatePercentileFloat(mttrValues, 95)
		snapshot.P99MTTR = calculatePercentileFloat(mttrValues, 99)
	}

	// Calculate trend
	previousPeriodStart := periodStart.Add(-(periodEnd.Sub(periodStart)))
	previousPeriodEnd := periodStart

	previousSnapshot, err := s.GetMetricsSnapshot(monitorID, tenantID, previousPeriodStart, previousPeriodEnd, periodType)
	if err == nil && previousSnapshot != nil && previousSnapshot.AvgMTTR > 0 {
		delta := snapshot.AvgMTTR - previousSnapshot.AvgMTTR
		snapshot.TrendPercent = (delta / previousSnapshot.AvgMTTR) * 100

		if delta < -5 {
			snapshot.TrendDirection = "improving"
		} else if delta > 5 {
			snapshot.TrendDirection = "degrading"
		} else {
			snapshot.TrendDirection = "stable"
		}
	} else {
		snapshot.TrendDirection = "stable"
	}

	snapshot.GeneratedAt = time.Now()

	// Save snapshot
	err = s.db.Create(snapshot).Error
	if err != nil {
		return nil, fmt.Errorf("failed to save metrics snapshot: %w", err)
	}

	s.logger.Info("Metrics snapshot generated",
		zap.Uint("monitor_id", monitorID),
		zap.Int("incidents", snapshot.TotalIncidents),
		zap.Float64("avg_mttr", snapshot.AvgMTTR),
		zap.Float64("avg_mttd", snapshot.AvgMTTD))

	return snapshot, nil
}

// GetMetricsSnapshot retrieves a metrics snapshot
func (s *MTTRMTTDTrackingService) GetMetricsSnapshot(monitorID uint, tenantID uuid.UUID, periodStart, periodEnd time.Time, periodType string) (*MetricsSnapshot, error) {
	var snapshot MetricsSnapshot
	query := s.db.Where("tenant_id = ? AND period_start = ? AND period_end = ? AND period_type = ?",
		tenantID, periodStart, periodEnd, periodType)

	if monitorID > 0 {
		query = query.Where("monitor_id = ?", monitorID)
	} else {
		query = query.Where("monitor_id = 0")
	}

	err := query.First(&snapshot).Error
	if err != nil {
		return nil, fmt.Errorf("metrics snapshot not found: %w", err)
	}

	return &snapshot, nil
}

// GetMetricsHistory retrieves historical metrics snapshots
func (s *MTTRMTTDTrackingService) GetMetricsHistory(monitorID uint, tenantID uuid.UUID, periodType string, limit int) ([]MetricsSnapshot, error) {
	var snapshots []MetricsSnapshot
	query := s.db.Where("tenant_id = ? AND period_type = ?", tenantID, periodType)

	if monitorID > 0 {
		query = query.Where("monitor_id = ?", monitorID)
	}

	err := query.Order("period_start DESC").
		Limit(limit).
		Find(&snapshots).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get metrics history: %w", err)
	}

	return snapshots, nil
}

// GetCurrentMonthMetrics gets metrics for the current month
func (s *MTTRMTTDTrackingService) GetCurrentMonthMetrics(monitorID uint, tenantID uuid.UUID) (*MetricsSnapshot, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	// Try to get existing snapshot
	snapshot, err := s.GetMetricsSnapshot(monitorID, tenantID, monthStart, monthEnd, "monthly")
	if err == nil {
		return snapshot, nil
	}

	// Generate new snapshot
	return s.CalculateMetrics(monitorID, tenantID, monthStart, monthEnd, "monthly")
}

// Helper function to calculate percentiles for float64 values
func calculatePercentileFloat(data []float64, percentile int) float64 {
	if len(data) == 0 {
		return 0
	}

	// Simple bubble sort for small datasets
	sortedData := make([]float64, len(data))
	copy(sortedData, data)

	for i := 0; i < len(sortedData); i++ {
		for j := i + 1; j < len(sortedData); j++ {
			if sortedData[i] > sortedData[j] {
				sortedData[i], sortedData[j] = sortedData[j], sortedData[i]
			}
		}
	}

	index := float64(percentile) / 100.0 * float64(len(sortedData)-1)
	lowerIndex := int(index)
	upperIndex := lowerIndex + 1

	if upperIndex >= len(sortedData) {
		return sortedData[len(sortedData)-1]
	}

	fraction := index - float64(lowerIndex)
	return sortedData[lowerIndex] + fraction*(sortedData[upperIndex]-sortedData[lowerIndex])
}
