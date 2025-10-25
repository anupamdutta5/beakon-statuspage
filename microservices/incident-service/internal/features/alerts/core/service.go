// Package core provides alert management business logic.
package core

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AlertService handles alert business logic including auto-resolution and deduplication.
type AlertService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAlertService creates a new alert service.
func NewAlertService(db *gorm.DB, logger *zap.Logger) *AlertService {
	return &AlertService{
		db:     db,
		logger: logger,
	}
}

// CreateAlert creates a new alert with deduplication check.
// P1 Feature: Alert Deduplication
func (s *AlertService) CreateAlert(ctx context.Context, alert *Alert) error {
	errorType := ""
	if alert.ErrorType != nil {
		errorType = *alert.ErrorType
	}
	// Generate deduplication key
	dedupKey := s.generateDedupKey(alert.ServiceID, errorType)
	alert.DedupKey = &dedupKey
	alert.ErrorType = &errorType

	// Check if alert already exists (within last 15 minutes)
	dedupWindow := 15 * time.Minute
	var existingAlert Alert
	err := s.db.Where("dedup_key = ?", dedupKey).
		Where("created_at > ?", time.Now().Add(-dedupWindow)).
		Where("status IN ?", []string{"active", "acknowledged"}).
		First(&existingAlert).Error

	if err == nil {
		// Alert already exists, skip creation (deduplicated)
		s.logger.Info("Skipping duplicate alert",
			zap.String("dedup_key", dedupKey),
			zap.Uint("existing_id", existingAlert.ID),
			zap.String("error_type", errorType))
		return nil
	}

	// Create new alert
	if err := s.db.Create(alert).Error; err != nil {
		s.logger.Error("Failed to create alert", zap.Error(err))
		return fmt.Errorf("failed to create alert: %w", err)
	}

	s.logger.Info("Alert created successfully",
		zap.Uint("alert_id", alert.ID),
		zap.String("dedup_key", dedupKey),
		zap.String("severity", alert.Severity))

	return nil
}

// generateDedupKey generates a deduplication key for an alert.
func (s *AlertService) generateDedupKey(serviceID *uint, errorType string) string {
	if serviceID == nil {
		return fmt.Sprintf("service-unknown-%s", errorType)
	}
	return fmt.Sprintf("service-%d-%s", *serviceID, errorType)
}

// AutoResolveAlerts automatically resolves alerts when service is operational.
// P1 Feature: Alert Auto-Resolution
// This should be called after 3 consecutive successful health checks.
func (s *AlertService) AutoResolveAlerts(serviceID uint) error {
	// Find active alerts for this service
	var alerts []Alert
	err := s.db.Where("service_id = ?", serviceID).
		Where("status IN ?", []string{"active", "acknowledged"}).
		Find(&alerts).Error

	if err != nil {
		return fmt.Errorf("failed to fetch alerts for auto-resolution: %w", err)
	}

	if len(alerts) == 0 {
		s.logger.Debug("No active alerts to auto-resolve", zap.Uint("service_id", serviceID))
		return nil
	}

	// Check if service is now operational (3 consecutive successes)
	if !s.isServiceOperational(serviceID) {
		s.logger.Debug("Service not yet operational, skipping auto-resolution",
			zap.Uint("service_id", serviceID))
		return nil
	}

	// Auto-resolve all active alerts
	resolutionType := "auto"
	resolutionNote := "Service returned to operational status after 3 consecutive successful health checks"
	resolvedAt := time.Now()

	for _, alert := range alerts {
		alert.Status = "resolved"
		alert.ResolvedAt = &resolvedAt
		alert.ResolutionType = &resolutionType
		alert.ResolutionNote = &resolutionNote

		if err := s.db.Save(&alert).Error; err != nil {
			s.logger.Error("Failed to auto-resolve alert",
				zap.Error(err),
				zap.Uint("alert_id", alert.ID))
			continue
		}

		s.logger.Info("Auto-resolved alert",
			zap.Uint("alert_id", alert.ID),
			zap.String("title", alert.Title),
			zap.Uint("service_id", serviceID))
	}

	return nil
}

// isServiceOperational checks if a service has 3 consecutive successful health checks.
func (s *AlertService) isServiceOperational(serviceID uint) bool {
	// Query last 3 uptime checks for this service
	// Note: This assumes uptime_checks table has service_id and status columns
	// You may need to adjust based on your actual schema
	var results []struct {
		Status string
	}

	err := s.db.Table("uptime_results").
		Select("status").
		Where("service_id = ?", serviceID).
		Order("checked_at DESC").
		Limit(3).
		Scan(&results).Error

	if err != nil || len(results) < 3 {
		s.logger.Debug("Not enough uptime results to determine operational status",
			zap.Uint("service_id", serviceID),
			zap.Int("results_count", len(results)))
		return false
	}

	// Check all 3 are operational
	for _, r := range results {
		if r.Status != "operational" && r.Status != "up" {
			return false
		}
	}

	return true
}

// GetAlerts retrieves alerts with filters.
func (s *AlertService) GetAlerts(tenantID uint, status string, limit, offset int) ([]*Alert, int64, error) {
	var alerts []*Alert
	var total int64

	query := s.db.Where("tenant_id = ?", tenantID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Get total count
	if err := query.Model(&Alert{}).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count alerts", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count alerts: %w", err)
	}

	// Get alerts with pagination
	if err := query.
		Preload("Service").
		Order("triggered_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&alerts).Error; err != nil {
		s.logger.Error("Failed to get alerts", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get alerts: %w", err)
	}

	return alerts, total, nil
}

// GetAlert retrieves a single alert by ID.
func (s *AlertService) GetAlert(id uint) (*Alert, error) {
	var alert Alert
	if err := s.db.Preload("Service").First(&alert, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("alert not found")
		}
		s.logger.Error("Failed to get alert", zap.Error(err))
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	return &alert, nil
}

// AcknowledgeAlert acknowledges an alert.
func (s *AlertService) AcknowledgeAlert(id uint, acknowledgedBy uint) error {
	now := time.Now()
	err := s.db.Model(&Alert{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":           "acknowledged",
			"acknowledged_at":  now,
			"acknowledged_by":  acknowledgedBy,
		}).Error

	if err != nil {
		s.logger.Error("Failed to acknowledge alert", zap.Error(err))
		return fmt.Errorf("failed to acknowledge alert: %w", err)
	}

	s.logger.Info("Alert acknowledged", zap.Uint("alert_id", id), zap.Uint("by_user", acknowledgedBy))
	return nil
}

// ResolveAlert manually resolves an alert.
func (s *AlertService) ResolveAlert(id uint, resolvedBy uint, note string) error {
	now := time.Now()
	resolutionType := "manual"

	err := s.db.Model(&Alert{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":          "resolved",
			"resolved_at":     now,
			"resolved_by":     resolvedBy,
			"resolution_type": resolutionType,
			"resolution_note": note,
		}).Error

	if err != nil {
		s.logger.Error("Failed to resolve alert", zap.Error(err))
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	s.logger.Info("Alert manually resolved",
		zap.Uint("alert_id", id),
		zap.Uint("by_user", resolvedBy),
		zap.String("note", note))
	return nil
}

// DeleteAlert deletes an alert.
func (s *AlertService) DeleteAlert(id uint) error {
	if err := s.db.Delete(&Alert{}, id).Error; err != nil {
		s.logger.Error("Failed to delete alert", zap.Error(err))
		return fmt.Errorf("failed to delete alert: %w", err)
	}

	s.logger.Info("Alert deleted", zap.Uint("alert_id", id))
	return nil
}

// GetAlertStatistics returns statistics for alerts.
func (s *AlertService) GetAlertStatistics(tenantID uint, startDate, endDate time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := s.db.Model(&Alert{}).Where("tenant_id = ?", tenantID)
	if !startDate.IsZero() {
		query = query.Where("triggered_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("triggered_at <= ?", endDate)
	}

	// Get total alerts
	var totalAlerts int64
	if err := query.Count(&totalAlerts).Error; err != nil {
		s.logger.Error("Failed to count alerts", zap.Error(err))
		return nil, fmt.Errorf("failed to count alerts: %w", err)
	}

	// Get alerts by status
	var statusCounts []struct {
		Status string
		Count  int64
	}
	if err := query.Select("status, COUNT(*) as count").Group("status").Scan(&statusCounts).Error; err != nil {
		s.logger.Error("Failed to get status counts", zap.Error(err))
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}

	// Get alerts by severity
	var severityCounts []struct {
		Severity string
		Count    int64
	}
	if err := query.Select("severity, COUNT(*) as count").Group("severity").Scan(&severityCounts).Error; err != nil {
		s.logger.Error("Failed to get severity counts", zap.Error(err))
		return nil, fmt.Errorf("failed to get severity counts: %w", err)
	}

	// Get auto-resolution statistics
	var autoResolved int64
	if err := s.db.Model(&Alert{}).
		Where("tenant_id = ? AND resolution_type = ?", tenantID, "auto").
		Count(&autoResolved).Error; err != nil {
		s.logger.Warn("Failed to count auto-resolved alerts", zap.Error(err))
		autoResolved = 0
	}

	stats["total_alerts"] = totalAlerts
	stats["status_breakdown"] = statusCounts
	stats["severity_breakdown"] = severityCounts
	stats["auto_resolved_count"] = autoResolved
	stats["date_range"] = map[string]interface{}{
		"start": startDate,
		"end":   endDate,
	}
	stats["timestamp"] = time.Now()

	return stats, nil
}
