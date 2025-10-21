// Package services provides SLA reporting and analytics business logic.
package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/anupamdutta5/analytics-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SLAService handles SLA-related business logic.
type SLAService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewSLAService creates a new SLA service.
func NewSLAService(db *gorm.DB, logger *zap.Logger) *SLAService {
	return &SLAService{
		db:     db,
		logger: logger,
	}
}

// CreateSLA creates a new SLA definition.
func (s *SLAService) CreateSLA(sla *models.SLA) error {
	if err := sla.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Create(sla).Error; err != nil {
		s.logger.Error("Failed to create SLA", zap.Error(err))
		return fmt.Errorf("failed to create SLA: %w", err)
	}

	s.logger.Info("SLA created successfully", zap.Uint("sla_id", sla.ID))
	return nil
}

// GetSLA retrieves an SLA by ID.
func (s *SLAService) GetSLA(id uint) (*models.SLA, error) {
	var sla models.SLA
	if err := s.db.Preload("Measurements").Preload("Breaches").First(&sla, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("SLA not found")
		}
		s.logger.Error("Failed to get SLA", zap.Error(err))
		return nil, fmt.Errorf("failed to get SLA: %w", err)
	}

	return &sla, nil
}

// GetSLAs retrieves SLAs for a tenant with pagination.
func (s *SLAService) GetSLAs(tenantID uint, limit, offset int) ([]*models.SLA, int64, error) {
	var slas []*models.SLA
	var total int64

	// Get total count
	if err := s.db.Model(&models.SLA{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count SLAs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count SLAs: %w", err)
	}

	// Get SLAs with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).
		Preload("Measurements").
		Preload("Breaches").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&slas).Error; err != nil {
		s.logger.Error("Failed to get SLAs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get SLAs: %w", err)
	}

	return slas, total, nil
}

// CalculateUptimeSLA calculates uptime SLA for a component over a period.
func (s *SLAService) CalculateUptimeSLA(componentID uint, startTime, endTime time.Time) (*models.UptimeCalculation, error) {
	totalMinutes := int64(endTime.Sub(startTime).Minutes())

	// Get incident data for the period (this would typically query incident service)
	// For now, we'll create a placeholder calculation
	uptimeCalc := &models.UptimeCalculation{
		ComponentID:     componentID,
		PeriodStart:     startTime,
		PeriodEnd:       endTime,
		TotalMinutes:    totalMinutes,
		UptimeMinutes:   totalMinutes, // Placeholder - would be calculated from actual incidents
		DowntimeMinutes: 0,
		IncidentCount:   0,
		CalculationType: "custom",
		Status:          "active",
	}

	uptimeCalc.CalculateUptimePercent()

	if err := s.db.Create(uptimeCalc).Error; err != nil {
		s.logger.Error("Failed to save uptime calculation", zap.Error(err))
		return nil, fmt.Errorf("failed to save uptime calculation: %w", err)
	}

	return uptimeCalc, nil
}

// CalculateSLAMeasurement calculates SLA measurement for a specific period.
func (s *SLAService) CalculateSLAMeasurement(slaID uint, startTime, endTime time.Time) (*models.SLAMeasurement, error) {
	sla, err := s.GetSLA(slaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SLA: %w", err)
	}

	measurement := &models.SLAMeasurement{
		SLAID:       slaID,
		PeriodStart: startTime,
		PeriodEnd:   endTime,
		TargetValue: sla.TargetValue,
		Status:      "active",
	}

	// Calculate actual value based on SLA type
	switch sla.Type {
	case "uptime":
		actualValue, err := s.calculateUptimeValue(sla.ComponentID, startTime, endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate uptime: %w", err)
		}
		measurement.ActualValue = actualValue
		measurement.TotalSamples = int64(endTime.Sub(startTime).Minutes())

	case "response_time":
		actualValue, samples, err := s.calculateResponseTimeValue(sla.ComponentID, startTime, endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate response time: %w", err)
		}
		measurement.ActualValue = actualValue
		measurement.TotalSamples = samples

	case "error_rate":
		actualValue, samples, err := s.calculateErrorRateValue(sla.ComponentID, startTime, endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate error rate: %w", err)
		}
		measurement.ActualValue = actualValue
		measurement.TotalSamples = samples

	default:
		return nil, fmt.Errorf("unsupported SLA type: %s", sla.Type)
	}

	// Calculate compliance
	measurement.SLA = *sla
	measurement.CalculateCompliance()

	// Determine status
	if !measurement.IsCompliant {
		measurement.Status = "breached"
	} else if measurement.ComplianceRate < 95.0 {
		measurement.Status = "warning"
	}

	// Save measurement
	if err := s.db.Create(measurement).Error; err != nil {
		s.logger.Error("Failed to save SLA measurement", zap.Error(err))
		return nil, fmt.Errorf("failed to save SLA measurement: %w", err)
	}

	// Check for SLA breach
	if !measurement.IsCompliant {
		if err := s.createSLABreach(measurement); err != nil {
			s.logger.Error("Failed to create SLA breach", zap.Error(err))
		}
	}

	return measurement, nil
}

// calculateUptimeValue calculates uptime percentage for a component.
func (s *SLAService) calculateUptimeValue(componentID *uint, startTime, endTime time.Time) (float64, error) {
	if componentID == nil {
		return 100.0, nil // Default uptime for overall SLA
	}

	// Get uptime calculation for the period
	var uptimeCalc models.UptimeCalculation
	err := s.db.Where("component_id = ? AND period_start = ? AND period_end = ?",
		*componentID, startTime, endTime).First(&uptimeCalc).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Calculate uptime if not already calculated
			calc, err := s.CalculateUptimeSLA(*componentID, startTime, endTime)
			if err != nil {
				return 0, err
			}
			return calc.UptimePercent, nil
		}
		return 0, err
	}

	return uptimeCalc.UptimePercent, nil
}

// calculateResponseTimeValue calculates average response time for a component.
func (s *SLAService) calculateResponseTimeValue(componentID *uint, startTime, endTime time.Time) (float64, int64, error) {
	if componentID == nil {
		return 0, 0, fmt.Errorf("component ID required for response time calculation")
	}

	var result struct {
		AvgResponseTime float64
		Count          int64
	}

	err := s.db.Model(&models.ResponseTimeMetric{}).
		Where("component_id = ? AND timestamp BETWEEN ? AND ? AND is_successful = ?",
			*componentID, startTime, endTime, true).
		Select("AVG(response_time) as avg_response_time, COUNT(*) as count").
		Scan(&result).Error

	if err != nil {
		return 0, 0, err
	}

	return result.AvgResponseTime, result.Count, nil
}

// calculateErrorRateValue calculates error rate for a component.
func (s *SLAService) calculateErrorRateValue(componentID *uint, startTime, endTime time.Time) (float64, int64, error) {
	if componentID == nil {
		return 0, 0, fmt.Errorf("component ID required for error rate calculation")
	}

	var totalCount, errorCount int64

	// Get total requests
	if err := s.db.Model(&models.ResponseTimeMetric{}).
		Where("component_id = ? AND timestamp BETWEEN ? AND ?", *componentID, startTime, endTime).
		Count(&totalCount).Error; err != nil {
		return 0, 0, err
	}

	// Get error requests
	if err := s.db.Model(&models.ResponseTimeMetric{}).
		Where("component_id = ? AND timestamp BETWEEN ? AND ? AND is_successful = ?",
			*componentID, startTime, endTime, false).
		Count(&errorCount).Error; err != nil {
		return 0, 0, err
	}

	if totalCount == 0 {
		return 0, 0, nil
	}

	errorRate := (float64(errorCount) / float64(totalCount)) * 100
	return errorRate, totalCount, nil
}

// createSLABreach creates an SLA breach record.
func (s *SLAService) createSLABreach(measurement *models.SLAMeasurement) error {
	impactValue := measurement.TargetValue - measurement.ActualValue
	if measurement.SLA.Type == "response_time" || measurement.SLA.Type == "error_rate" {
		impactValue = measurement.ActualValue - measurement.TargetValue
	}

	// Determine severity based on impact
	severity := "minor"
	if impactValue > measurement.TargetValue*0.1 {
		severity = "major"
	}
	if impactValue > measurement.TargetValue*0.2 {
		severity = "critical"
	}

	breach := &models.SLABreach{
		SLAID:         measurement.SLAID,
		MeasurementID: measurement.ID,
		Severity:      severity,
		Status:        "open",
		TriggeredAt:   measurement.PeriodEnd,
		DetectedAt:    time.Now(),
		ImpactValue:   impactValue,
		Description:   fmt.Sprintf("SLA breach detected for %s", measurement.SLA.Name),
	}

	if err := s.db.Create(breach).Error; err != nil {
		return fmt.Errorf("failed to create SLA breach: %w", err)
	}

	s.logger.Warn("SLA breach created",
		zap.Uint("sla_id", measurement.SLAID),
		zap.String("severity", severity),
		zap.Float64("impact", impactValue))

	return nil
}

// GetSLABreaches retrieves SLA breaches for a tenant.
func (s *SLAService) GetSLABreaches(tenantID uint, limit, offset int) ([]*models.SLABreach, int64, error) {
	var breaches []*models.SLABreach
	var total int64

	// Get total count
	query := s.db.Model(&models.SLABreach{}).
		Joins("JOIN slas ON sla_breaches.sla_id = slas.id").
		Where("slas.tenant_id = ?", tenantID)

	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count SLA breaches", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count SLA breaches: %w", err)
	}

	// Get breaches with pagination
	if err := s.db.Preload("SLA").Preload("Measurement").
		Joins("JOIN slas ON sla_breaches.sla_id = slas.id").
		Where("slas.tenant_id = ?", tenantID).
		Limit(limit).
		Offset(offset).
		Order("sla_breaches.triggered_at DESC").
		Find(&breaches).Error; err != nil {
		s.logger.Error("Failed to get SLA breaches", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get SLA breaches: %w", err)
	}

	return breaches, total, nil
}

// GenerateSLAReport generates a comprehensive SLA report.
func (s *SLAService) GenerateSLAReport(tenantID uint, slaID *uint, reportType, period string, startTime, endTime time.Time) (*models.SLAReport, error) {
	report := &models.SLAReport{
		TenantID:    tenantID,
		SLAID:       slaID,
		Name:        fmt.Sprintf("SLA Report - %s", period),
		Type:        reportType,
		Period:      period,
		PeriodStart: startTime,
		PeriodEnd:   endTime,
		Status:      "generating",
		Format:      "json",
	}

	if err := s.db.Create(report).Error; err != nil {
		return nil, fmt.Errorf("failed to create SLA report: %w", err)
	}

	// Generate report data
	reportData, err := s.generateReportData(tenantID, slaID, reportType, startTime, endTime)
	if err != nil {
		report.Status = "failed"
		report.Error = err.Error()
		s.db.Save(report)
		return nil, fmt.Errorf("failed to generate report data: %w", err)
	}

	// Save report summary
	summaryJSON, _ := json.Marshal(reportData)
	report.Summary = string(summaryJSON)
	report.Status = "completed"
	report.GeneratedAt = &[]time.Time{time.Now()}[0]

	if err := s.db.Save(report).Error; err != nil {
		return nil, fmt.Errorf("failed to update report: %w", err)
	}

	return report, nil
}

// generateReportData generates the actual report data.
func (s *SLAService) generateReportData(tenantID uint, slaID *uint, reportType string, startTime, endTime time.Time) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"tenant_id":    tenantID,
		"report_type":  reportType,
		"period_start": startTime,
		"period_end":   endTime,
		"generated_at": time.Now(),
	}

	if slaID != nil {
		// Single SLA report
		sla, err := s.GetSLA(*slaID)
		if err != nil {
			return nil, err
		}

		// Get measurements for the period
		var measurements []models.SLAMeasurement
		if err := s.db.Where("sla_id = ? AND period_start >= ? AND period_end <= ?",
			*slaID, startTime, endTime).Find(&measurements).Error; err != nil {
			return nil, err
		}

		// Get breaches for the period
		var breaches []models.SLABreach
		if err := s.db.Where("sla_id = ? AND triggered_at BETWEEN ? AND ?",
			*slaID, startTime, endTime).Find(&breaches).Error; err != nil {
			return nil, err
		}

		// Calculate summary statistics
		summary := s.calculateSLASummary(sla, measurements, breaches)

		data["sla"] = sla
		data["measurements"] = measurements
		data["breaches"] = breaches
		data["summary"] = summary

	} else {
		// Multi-SLA report
		slas, _, err := s.GetSLAs(tenantID, 1000, 0) // Get all SLAs
		if err != nil {
			return nil, err
		}

		var allMeasurements []models.SLAMeasurement
		var allBreaches []models.SLABreach

		slasSummary := make([]map[string]interface{}, 0)

		for _, sla := range slas {
			// Get measurements for each SLA
			var measurements []models.SLAMeasurement
			if err := s.db.Where("sla_id = ? AND period_start >= ? AND period_end <= ?",
				sla.ID, startTime, endTime).Find(&measurements).Error; err != nil {
				continue
			}

			// Get breaches for each SLA
			var breaches []models.SLABreach
			if err := s.db.Where("sla_id = ? AND triggered_at BETWEEN ? AND ?",
				sla.ID, startTime, endTime).Find(&breaches).Error; err != nil {
				continue
			}

			allMeasurements = append(allMeasurements, measurements...)
			allBreaches = append(allBreaches, breaches...)

			// Calculate summary for this SLA
			summary := s.calculateSLASummary(sla, measurements, breaches)
			summary["sla_id"] = sla.ID
			summary["sla_name"] = sla.Name
			slasSummary = append(slasSummary, summary)
		}

		data["slas"] = slas
		data["measurements"] = allMeasurements
		data["breaches"] = allBreaches
		data["slas_summary"] = slasSummary
		data["overall_summary"] = s.calculateOverallSummary(allMeasurements, allBreaches)
	}

	return data, nil
}

// calculateSLASummary calculates summary statistics for an SLA.
func (s *SLAService) calculateSLASummary(sla *models.SLA, measurements []models.SLAMeasurement, breaches []models.SLABreach) map[string]interface{} {
	summary := map[string]interface{}{
		"total_measurements": len(measurements),
		"total_breaches":     len(breaches),
		"compliance_rate":    0.0,
		"avg_actual_value":   0.0,
		"target_value":       sla.TargetValue,
		"sla_type":          sla.Type,
		"period_type":       sla.PeriodType,
	}

	if len(measurements) > 0 {
		var totalCompliance, totalActual float64
		compliantCount := 0

		for _, m := range measurements {
			totalCompliance += m.ComplianceRate
			totalActual += m.ActualValue
			if m.IsCompliant {
				compliantCount++
			}
		}

		summary["compliance_rate"] = totalCompliance / float64(len(measurements))
		summary["avg_actual_value"] = totalActual / float64(len(measurements))
		summary["compliant_periods"] = compliantCount
		summary["breach_periods"] = len(measurements) - compliantCount
	}

	// Breach analysis
	if len(breaches) > 0 {
		severityCounts := map[string]int{
			"warning":  0,
			"minor":    0,
			"major":    0,
			"critical": 0,
		}

		var totalImpact float64
		for _, b := range breaches {
			severityCounts[b.Severity]++
			totalImpact += b.ImpactValue
		}

		summary["severity_breakdown"] = severityCounts
		summary["avg_impact"] = totalImpact / float64(len(breaches))
	}

	return summary
}

// calculateOverallSummary calculates overall summary across all SLAs.
func (s *SLAService) calculateOverallSummary(measurements []models.SLAMeasurement, breaches []models.SLABreach) map[string]interface{} {
	summary := map[string]interface{}{
		"total_measurements": len(measurements),
		"total_breaches":     len(breaches),
		"overall_compliance": 0.0,
	}

	if len(measurements) > 0 {
		var totalCompliance float64
		compliantCount := 0

		for _, m := range measurements {
			totalCompliance += m.ComplianceRate
			if m.IsCompliant {
				compliantCount++
			}
		}

		summary["overall_compliance"] = totalCompliance / float64(len(measurements))
		summary["compliant_measurements"] = compliantCount
		summary["breach_rate"] = float64(len(breaches)) / float64(len(measurements)) * 100
	}

	return summary
}

// GetSLAStatistics returns comprehensive SLA statistics for a tenant.
func (s *SLAService) GetSLAStatistics(tenantID uint, startTime, endTime time.Time) (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"period_start": startTime,
		"period_end":   endTime,
		"generated_at": time.Now(),
	}

	// Get total SLAs
	var totalSLAs int64
	if err := s.db.Model(&models.SLA{}).Where("tenant_id = ? AND is_active = ?", tenantID, true).Count(&totalSLAs).Error; err != nil {
		return nil, err
	}
	stats["total_slas"] = totalSLAs

	// Get SLA measurements in period
	var measurements []models.SLAMeasurement
	if err := s.db.Joins("JOIN slas ON sla_measurements.sla_id = slas.id").
		Where("slas.tenant_id = ? AND sla_measurements.period_start >= ? AND sla_measurements.period_end <= ?",
			tenantID, startTime, endTime).
		Find(&measurements).Error; err != nil {
		return nil, err
	}

	// Get SLA breaches in period
	var breaches []models.SLABreach
	if err := s.db.Joins("JOIN slas ON sla_breaches.sla_id = slas.id").
		Where("slas.tenant_id = ? AND sla_breaches.triggered_at BETWEEN ? AND ?",
			tenantID, startTime, endTime).
		Find(&breaches).Error; err != nil {
		return nil, err
	}

	// Calculate statistics
	if len(measurements) > 0 {
		var totalCompliance float64
		compliantCount := 0
		typeStats := make(map[string]map[string]interface{})

		for _, m := range measurements {
			totalCompliance += m.ComplianceRate
			if m.IsCompliant {
				compliantCount++
			}

			// Group by SLA type
			if _, exists := typeStats[m.SLA.Type]; !exists {
				typeStats[m.SLA.Type] = map[string]interface{}{
					"count":           0,
					"compliant":       0,
					"avg_compliance":  0.0,
					"total_compliance": 0.0,
				}
			}

			typeStats[m.SLA.Type]["count"] = typeStats[m.SLA.Type]["count"].(int) + 1
			if m.IsCompliant {
				typeStats[m.SLA.Type]["compliant"] = typeStats[m.SLA.Type]["compliant"].(int) + 1
			}
			typeStats[m.SLA.Type]["total_compliance"] = typeStats[m.SLA.Type]["total_compliance"].(float64) + m.ComplianceRate
		}

		// Calculate averages for each type
		for slaType, data := range typeStats {
			count := data["count"].(int)
			if count > 0 {
				typeStats[slaType]["avg_compliance"] = data["total_compliance"].(float64) / float64(count)
			}
		}

		stats["total_measurements"] = len(measurements)
		stats["compliant_measurements"] = compliantCount
		stats["overall_compliance"] = totalCompliance / float64(len(measurements))
		stats["compliance_by_type"] = typeStats
	}

	if len(breaches) > 0 {
		severityStats := make(map[string]int)
		for _, b := range breaches {
			severityStats[b.Severity]++
		}
		stats["total_breaches"] = len(breaches)
		stats["breaches_by_severity"] = severityStats
	}

	return stats, nil
}