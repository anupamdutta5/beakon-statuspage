package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SLATarget represents an SLA target configuration for a monitor
type SLATarget struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	TenantID            uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	MonitorID           uint      `gorm:"not null;index" json:"monitor_id"`
	Name                string    `gorm:"size:255;not null" json:"name"`
	TargetUptimePercent float64   `gorm:"not null" json:"target_uptime_percent"` // e.g., 99.9, 99.95, 99.99
	PeriodType          string    `gorm:"size:50;not null" json:"period_type"` // daily, weekly, monthly, quarterly, yearly
	IsActive            bool      `gorm:"default:true" json:"is_active"`
	NotifyOnBreach      bool      `gorm:"default:true" json:"notify_on_breach"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// SLAReport represents a calculated SLA report for a period
type SLAReport struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	TenantID             uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	MonitorID            uint      `gorm:"not null;index" json:"monitor_id"`
	SLATargetID          uint      `gorm:"index" json:"sla_target_id"`
	PeriodStart          time.Time `gorm:"not null;index" json:"period_start"`
	PeriodEnd            time.Time `gorm:"not null;index" json:"period_end"`
	PeriodType           string    `gorm:"size:50;not null" json:"period_type"`

	// Uptime Metrics
	TotalChecks          int       `json:"total_checks"`
	SuccessfulChecks     int       `json:"successful_checks"`
	FailedChecks         int       `json:"failed_checks"`
	UptimePercent        float64   `json:"uptime_percent"`
	DowntimeMinutes      float64   `json:"downtime_minutes"`

	// SLA Compliance
	TargetUptimePercent  float64   `json:"target_uptime_percent"`
	UptimeDelta          float64   `json:"uptime_delta"` // Actual - Target
	IsSLAMet             bool      `json:"is_sla_met"`
	SLACreditsMinutes    float64   `json:"sla_credits_minutes"` // Downtime beyond SLA

	// Response Time Metrics
	AvgResponseTime      float64   `json:"avg_response_time"`
	P50ResponseTime      int       `json:"p50_response_time"`
	P95ResponseTime      int       `json:"p95_response_time"`
	P99ResponseTime      int       `json:"p99_response_time"`

	// Incident Metrics
	TotalIncidents       int       `json:"total_incidents"`
	MTTR                 float64   `json:"mttr"` // Mean Time To Resolve (minutes)
	MTTD                 float64   `json:"mttd"` // Mean Time To Detect (minutes)

	GeneratedAt          time.Time `json:"generated_at"`
	CreatedAt            time.Time `json:"created_at"`
}

// SLABreach represents a breach of SLA target
type SLABreach struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TenantID        uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	MonitorID       uint      `gorm:"not null;index" json:"monitor_id"`
	SLATargetID     uint      `gorm:"not null;index" json:"sla_target_id"`
	SLAReportID     uint      `gorm:"index" json:"sla_report_id"`
	PeriodStart     time.Time `gorm:"not null" json:"period_start"`
	PeriodEnd       time.Time `gorm:"not null" json:"period_end"`
	TargetPercent   float64   `json:"target_percent"`
	ActualPercent   float64   `json:"actual_percent"`
	DeficitPercent  float64   `json:"deficit_percent"` // Target - Actual
	DowntimeMinutes float64   `json:"downtime_minutes"`
	NotificationSent bool     `gorm:"default:false" json:"notification_sent"`
	AcknowledgedAt  *time.Time `json:"acknowledged_at"`
	CreatedAt       time.Time `json:"created_at"`
}

// SLAReportingService manages SLA targets and generates SLA reports
type SLAReportingService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewSLAReportingService creates a new SLA reporting service
func NewSLAReportingService(db *gorm.DB, logger *zap.Logger) *SLAReportingService {
	return &SLAReportingService{
		db:     db,
		logger: logger,
	}
}

// CreateSLATarget creates a new SLA target
func (s *SLAReportingService) CreateSLATarget(target *SLATarget) error {
	// Validate target uptime percent
	if target.TargetUptimePercent < 0 || target.TargetUptimePercent > 100 {
		return fmt.Errorf("target uptime percent must be between 0 and 100")
	}

	// Validate period type
	validPeriods := map[string]bool{
		"daily": true, "weekly": true, "monthly": true, "quarterly": true, "yearly": true,
	}
	if !validPeriods[target.PeriodType] {
		return fmt.Errorf("invalid period type: %s", target.PeriodType)
	}

	err := s.db.Create(target).Error
	if err != nil {
		s.logger.Error("Failed to create SLA target", zap.Error(err))
		return fmt.Errorf("failed to create SLA target: %w", err)
	}

	s.logger.Info("SLA target created",
		zap.Uint("id", target.ID),
		zap.Uint("monitor_id", target.MonitorID),
		zap.Float64("target_percent", target.TargetUptimePercent))

	return nil
}

// UpdateSLATarget updates an existing SLA target
func (s *SLAReportingService) UpdateSLATarget(target *SLATarget) error {
	err := s.db.Save(target).Error
	if err != nil {
		s.logger.Error("Failed to update SLA target", zap.Error(err))
		return fmt.Errorf("failed to update SLA target: %w", err)
	}

	s.logger.Info("SLA target updated", zap.Uint("id", target.ID))
	return nil
}

// GetSLATarget retrieves an SLA target by ID
func (s *SLAReportingService) GetSLATarget(id uint, tenantID uuid.UUID) (*SLATarget, error) {
	var target SLATarget
	err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&target).Error
	if err != nil {
		return nil, fmt.Errorf("SLA target not found: %w", err)
	}
	return &target, nil
}

// GetSLATargetsByMonitor retrieves all SLA targets for a monitor
func (s *SLAReportingService) GetSLATargetsByMonitor(monitorID uint) ([]SLATarget, error) {
	var targets []SLATarget
	err := s.db.Where("monitor_id = ? AND is_active = ?", monitorID, true).Find(&targets).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get SLA targets: %w", err)
	}
	return targets, nil
}

// DeleteSLATarget deletes an SLA target
func (s *SLAReportingService) DeleteSLATarget(id uint, tenantID uuid.UUID) error {
	result := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&SLATarget{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete SLA target: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("SLA target not found")
	}

	s.logger.Info("SLA target deleted", zap.Uint("id", id))
	return nil
}

// GenerateSLAReport generates an SLA report for a specific period
func (s *SLAReportingService) GenerateSLAReport(monitorID uint, tenantID uuid.UUID, periodStart, periodEnd time.Time, periodType string) (*SLAReport, error) {
	// Fetch monitoring results for the period
	var results []MonitoringResult
	err := s.db.Where("monitor_id = ? AND checked_at BETWEEN ? AND ?", monitorID, periodStart, periodEnd).
		Order("checked_at ASC").
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch monitoring results: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no monitoring data available for the period")
	}

	// Calculate basic metrics
	totalChecks := len(results)
	successfulChecks := 0
	failedChecks := 0
	var totalResponseTime int64

	for _, result := range results {
		if result.Status == "operational" {
			successfulChecks++
		} else {
			failedChecks++
		}
		totalResponseTime += int64(result.ResponseTimeMS)
	}

	uptimePercent := float64(successfulChecks) / float64(totalChecks) * 100
	avgResponseTime := float64(totalResponseTime) / float64(totalChecks)

	// Calculate downtime in minutes
	totalPeriodMinutes := periodEnd.Sub(periodStart).Minutes()
	downtimeMinutes := totalPeriodMinutes * (100 - uptimePercent) / 100

	// Calculate percentiles
	responseTimes := make([]int, len(results))
	for i, result := range results {
		responseTimes[i] = result.ResponseTimeMS
	}
	p50 := calculatePercentile(responseTimes, 50)
	p95 := calculatePercentile(responseTimes, 95)
	p99 := calculatePercentile(responseTimes, 99)

	// Get SLA target
	var target SLATarget
	err = s.db.Where("monitor_id = ? AND period_type = ? AND is_active = ?", monitorID, periodType, true).
		First(&target).Error

	targetUptimePercent := 0.0
	slaTargetID := uint(0)
	if err == nil {
		targetUptimePercent = target.TargetUptimePercent
		slaTargetID = target.ID
	}

	// Calculate SLA compliance
	uptimeDelta := uptimePercent - targetUptimePercent
	isSLAMet := uptimePercent >= targetUptimePercent
	slaCreditsMinutes := 0.0
	if !isSLAMet {
		// Calculate downtime beyond SLA
		allowedDowntimePercent := 100 - targetUptimePercent
		actualDowntimePercent := 100 - uptimePercent
		excessDowntimePercent := actualDowntimePercent - allowedDowntimePercent
		slaCreditsMinutes = totalPeriodMinutes * excessDowntimePercent / 100
	}

	// Calculate incident metrics (MTTR, MTTD)
	totalIncidents, mttr, mttd := s.calculateIncidentMetrics(monitorID, periodStart, periodEnd)

	// Create SLA report
	report := &SLAReport{
		TenantID:            tenantID,
		MonitorID:           monitorID,
		SLATargetID:         slaTargetID,
		PeriodStart:         periodStart,
		PeriodEnd:           periodEnd,
		PeriodType:          periodType,
		TotalChecks:         totalChecks,
		SuccessfulChecks:    successfulChecks,
		FailedChecks:        failedChecks,
		UptimePercent:       uptimePercent,
		DowntimeMinutes:     downtimeMinutes,
		TargetUptimePercent: targetUptimePercent,
		UptimeDelta:         uptimeDelta,
		IsSLAMet:            isSLAMet,
		SLACreditsMinutes:   slaCreditsMinutes,
		AvgResponseTime:     avgResponseTime,
		P50ResponseTime:     p50,
		P95ResponseTime:     p95,
		P99ResponseTime:     p99,
		TotalIncidents:      totalIncidents,
		MTTR:                mttr,
		MTTD:                mttd,
		GeneratedAt:         time.Now(),
	}

	// Save report
	err = s.db.Create(report).Error
	if err != nil {
		return nil, fmt.Errorf("failed to save SLA report: %w", err)
	}

	s.logger.Info("SLA report generated",
		zap.Uint("monitor_id", monitorID),
		zap.String("period", periodType),
		zap.Float64("uptime", uptimePercent),
		zap.Bool("sla_met", isSLAMet))

	// Check for SLA breach
	if !isSLAMet && slaTargetID > 0 {
		s.recordSLABreach(report, &target)
	}

	return report, nil
}

// calculateIncidentMetrics calculates MTTR and MTTD from monitoring results
func (s *SLAReportingService) calculateIncidentMetrics(monitorID uint, periodStart, periodEnd time.Time) (int, float64, float64) {
	var results []MonitoringResult
	s.db.Where("monitor_id = ? AND checked_at BETWEEN ? AND ?", monitorID, periodStart, periodEnd).
		Order("checked_at ASC").
		Find(&results)

	if len(results) == 0 {
		return 0, 0.0, 0.0
	}

	// Detect incidents (sequences of failures)
	type incident struct {
		startTime time.Time
		endTime   time.Time
		detected  bool
	}

	var incidents []incident
	var currentIncident *incident

	for i, result := range results {
		if result.Status != "operational" {
			if currentIncident == nil {
				// Start new incident
				currentIncident = &incident{
					startTime: result.CheckedAt,
					detected:  false,
				}
			}
		} else {
			if currentIncident != nil {
				// End current incident
				currentIncident.endTime = result.CheckedAt
				// MTTD: Time from incident start to first detection (first failed check)
				if !currentIncident.detected {
					currentIncident.detected = true
				}
				incidents = append(incidents, *currentIncident)
				currentIncident = nil
			}
		}

		// If it's the last result and we have an ongoing incident
		if i == len(results)-1 && currentIncident != nil {
			currentIncident.endTime = result.CheckedAt
			incidents = append(incidents, *currentIncident)
		}
	}

	totalIncidents := len(incidents)
	if totalIncidents == 0 {
		return 0, 0.0, 0.0
	}

	// Calculate MTTR (Mean Time To Resolve)
	var totalResolutionTime float64
	for _, inc := range incidents {
		resolutionTime := inc.endTime.Sub(inc.startTime).Minutes()
		totalResolutionTime += resolutionTime
	}
	mttr := totalResolutionTime / float64(totalIncidents)

	// Calculate MTTD (Mean Time To Detect)
	// For simplicity, we assume detection happens at the first failed check
	// In a real system, this might be tracked separately
	var checkInterval float64
	if len(results) > 1 {
		checkInterval = results[1].CheckedAt.Sub(results[0].CheckedAt).Minutes()
	} else {
		checkInterval = 1.0 // Default to 1 minute
	}
	mttd := checkInterval // Typically one check interval

	return totalIncidents, mttr, mttd
}

// recordSLABreach records an SLA breach
func (s *SLAReportingService) recordSLABreach(report *SLAReport, target *SLATarget) {
	breach := &SLABreach{
		TenantID:        report.TenantID,
		MonitorID:       report.MonitorID,
		SLATargetID:     report.SLATargetID,
		SLAReportID:     report.ID,
		PeriodStart:     report.PeriodStart,
		PeriodEnd:       report.PeriodEnd,
		TargetPercent:   report.TargetUptimePercent,
		ActualPercent:   report.UptimePercent,
		DeficitPercent:  report.TargetUptimePercent - report.UptimePercent,
		DowntimeMinutes: report.SLACreditsMinutes,
	}

	err := s.db.Create(breach).Error
	if err != nil {
		s.logger.Error("Failed to record SLA breach", zap.Error(err))
		return
	}

	s.logger.Warn("SLA breach recorded",
		zap.Uint("monitor_id", report.MonitorID),
		zap.Float64("target", report.TargetUptimePercent),
		zap.Float64("actual", report.UptimePercent),
		zap.Float64("deficit", breach.DeficitPercent))

	// TODO: Send notification if configured
	if target.NotifyOnBreach {
		s.logger.Info("SLA breach notification should be sent",
			zap.Uint("breach_id", breach.ID))
	}
}

// GetSLAReport retrieves an SLA report by ID
func (s *SLAReportingService) GetSLAReport(id uint, tenantID uuid.UUID) (*SLAReport, error) {
	var report SLAReport
	err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&report).Error
	if err != nil {
		return nil, fmt.Errorf("SLA report not found: %w", err)
	}
	return &report, nil
}

// GetSLAReportsByMonitor retrieves SLA reports for a monitor
func (s *SLAReportingService) GetSLAReportsByMonitor(monitorID uint, limit int) ([]SLAReport, error) {
	var reports []SLAReport
	err := s.db.Where("monitor_id = ?", monitorID).
		Order("period_start DESC").
		Limit(limit).
		Find(&reports).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get SLA reports: %w", err)
	}

	return reports, nil
}

// GetSLAReportsByPeriod retrieves SLA reports for a specific period type
func (s *SLAReportingService) GetSLAReportsByPeriod(monitorID uint, periodType string, limit int) ([]SLAReport, error) {
	var reports []SLAReport
	err := s.db.Where("monitor_id = ? AND period_type = ?", monitorID, periodType).
		Order("period_start DESC").
		Limit(limit).
		Find(&reports).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get SLA reports: %w", err)
	}

	return reports, nil
}

// GetSLABreaches retrieves SLA breaches for a monitor
func (s *SLAReportingService) GetSLABreaches(monitorID uint, limit int) ([]SLABreach, error) {
	var breaches []SLABreach
	err := s.db.Where("monitor_id = ?", monitorID).
		Order("created_at DESC").
		Limit(limit).
		Find(&breaches).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get SLA breaches: %w", err)
	}

	return breaches, nil
}

// AcknowledgeBreach acknowledges an SLA breach
func (s *SLAReportingService) AcknowledgeBreach(id uint) error {
	now := time.Now()
	result := s.db.Model(&SLABreach{}).Where("id = ?", id).Updates(map[string]interface{}{
		"acknowledged_at": now,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to acknowledge breach: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("breach not found")
	}

	s.logger.Info("SLA breach acknowledged", zap.Uint("id", id))
	return nil
}

// CalculateSLACredits calculates SLA credits for a tenant
func (s *SLAReportingService) CalculateSLACredits(tenantID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	var totalCredits float64

	var breaches []SLABreach
	err := s.db.Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, startDate, endDate).
		Find(&breaches).Error

	if err != nil {
		return 0, fmt.Errorf("failed to calculate SLA credits: %w", err)
	}

	for _, breach := range breaches {
		totalCredits += breach.DowntimeMinutes
	}

	return totalCredits, nil
}

// GetCurrentMonthSLA gets the current month's SLA for a monitor
func (s *SLAReportingService) GetCurrentMonthSLA(monitorID uint, tenantID uuid.UUID) (*SLAReport, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	// Try to get existing report
	var report SLAReport
	err := s.db.Where("monitor_id = ? AND period_type = ? AND period_start = ?", monitorID, "monthly", monthStart).
		First(&report).Error

	if err == nil {
		return &report, nil
	}

	// Generate new report
	return s.GenerateSLAReport(monitorID, tenantID, monthStart, monthEnd, "monthly")
}

// Helper function to calculate percentiles (copied from performance metrics)
func calculatePercentile(sortedData []int, percentile int) int {
	if len(sortedData) == 0 {
		return 0
	}

	// Sort data
	// Note: In production, this should use a proper sorting function
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
	return sortedData[lowerIndex] + int(fraction*float64(sortedData[upperIndex]-sortedData[lowerIndex]))
}
