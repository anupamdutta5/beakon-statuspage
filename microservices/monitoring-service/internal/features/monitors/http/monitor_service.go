// Package services provides business logic for the Monitoring Service.
package http

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/anupamdutta5/monitoring-service/internal/core/events"
	"github.com/anupamdutta5/monitoring-service/internal/models"
)

// MonitorService handles monitor-related business logic including auto-incident creation.
type MonitorService struct {
	db             *gorm.DB
	eventPublisher *events.EventPublisher
}

// NewMonitorService creates a new monitor service instance.
func NewMonitorService(db *gorm.DB, eventPublisher *events.EventPublisher) *MonitorService {
	return &MonitorService{
		db:             db,
		eventPublisher: eventPublisher,
	}
}

// CreateMonitor creates a new monitor.
func (s *MonitorService) CreateMonitor(monitor *models.Monitor) error {
	// Validate required fields
	if monitor.TenantID == uuid.Nil {
		return fmt.Errorf("tenant_id is required")
	}
	if monitor.ComponentID == uuid.Nil {
		return fmt.Errorf("component_id is required")
	}
	if monitor.MonitorType == "" {
		return fmt.Errorf("monitor_type is required")
	}

	// Set defaults
	if monitor.CheckIntervalSeconds == 0 {
		monitor.CheckIntervalSeconds = 60
	}
	if monitor.TimeoutSeconds == 0 {
		monitor.TimeoutSeconds = 30
	}
	if monitor.FailureThreshold == 0 {
		monitor.FailureThreshold = 3
	}

	// Create in database
	if err := s.db.Create(monitor).Error; err != nil {
		return fmt.Errorf("failed to create monitor: %w", err)
	}

	log.Printf("✅ Created monitor: %s (ID: %d) for component %s", monitor.Name, monitor.ID, monitor.ComponentID)
	return nil
}

// GetMonitorByID retrieves a monitor by ID.
func (s *MonitorService) GetMonitorByID(id uint) (*models.Monitor, error) {
	var monitor models.Monitor
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&monitor).Error; err != nil {
		return nil, err
	}
	return &monitor, nil
}

// GetMonitorsByTenant retrieves all monitors for a tenant.
func (s *MonitorService) GetMonitorsByTenant(tenantID uuid.UUID) ([]*models.Monitor, error) {
	var monitors []*models.Monitor
	if err := s.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("name ASC").
		Find(&monitors).Error; err != nil {
		return nil, err
	}
	return monitors, nil
}

// GetMonitorByComponent retrieves the monitor for a specific component.
func (s *MonitorService) GetMonitorByComponent(componentID uuid.UUID) (*models.Monitor, error) {
	var monitor models.Monitor
	if err := s.db.Where("component_id = ? AND deleted_at IS NULL", componentID).
		First(&monitor).Error; err != nil {
		return nil, err
	}
	return &monitor, nil
}

// GetActiveMonitors retrieves all active monitors.
func (s *MonitorService) GetActiveMonitors() ([]*models.Monitor, error) {
	var monitors []*models.Monitor
	if err := s.db.Where("is_active = ? AND deleted_at IS NULL", true).
		Find(&monitors).Error; err != nil {
		return nil, err
	}
	return monitors, nil
}

// RecordCheckSuccess records a successful health check for a monitor.
func (s *MonitorService) RecordCheckSuccess(monitorID uint) error {
	var monitor models.Monitor
	if err := s.db.First(&monitor, monitorID).Error; err != nil {
		return fmt.Errorf("monitor not found: %w", err)
	}

	// Check if there's an active incident before resetting (need to check before status changes)
	hasActiveIncident := monitor.LastIncidentID != nil

	// Update monitor status
	monitor.ResetFailures()

	if err := s.db.Save(&monitor).Error; err != nil {
		return fmt.Errorf("failed to update monitor: %w", err)
	}

	// Record status change in history
	s.recordStatusHistory(&monitor, "operational", "health_check", "")

	// Auto-resolve incident if there was one active
	if hasActiveIncident && monitor.AutoCreateIncidents && !monitor.InMaintenance {
		if err := s.autoResolveIncident(&monitor); err != nil {
			log.Printf("❌ Failed to auto-resolve incident: %v", err)
		}
	}

	return nil
}

// RecordCheckFailure records a failed health check for a monitor.
func (s *MonitorService) RecordCheckFailure(monitorID uint, errorMessage string) error {
	var monitor models.Monitor
	if err := s.db.First(&monitor, monitorID).Error; err != nil {
		return fmt.Errorf("monitor not found: %w", err)
	}

	// Skip if in maintenance
	if monitor.InMaintenance {
		log.Printf("⚠️  Monitor %d is in maintenance, skipping failure tracking", monitorID)
		return nil
	}

	// Increment failures
	monitor.IncrementFailures()

	if err := s.db.Save(&monitor).Error; err != nil {
		return fmt.Errorf("failed to update monitor: %w", err)
	}

	// Record status change in history
	s.recordStatusHistory(&monitor, "down", "health_check", errorMessage)

	// Check if we should auto-create an incident
	if monitor.ShouldCreateIncident() {
		if err := s.autoCreateIncident(&monitor, errorMessage); err != nil {
			log.Printf("❌ Failed to auto-create incident: %v", err)
		}
	}

	log.Printf("⚠️  Monitor %d failure count: %d/%d", monitorID, monitor.ConsecutiveFailures, monitor.FailureThreshold)
	return nil
}

// RecordCheckDegraded records a degraded health check for a monitor.
func (s *MonitorService) RecordCheckDegraded(monitorID uint, errorMessage string) error {
	var monitor models.Monitor
	if err := s.db.First(&monitor, monitorID).Error; err != nil {
		return fmt.Errorf("monitor not found: %w", err)
	}

	monitor.SetDegraded()

	if err := s.db.Save(&monitor).Error; err != nil {
		return fmt.Errorf("failed to update monitor: %w", err)
	}

	// Record status change in history
	s.recordStatusHistory(&monitor, "degraded", "health_check", errorMessage)

	return nil
}

// autoCreateIncident creates an incident automatically when failure threshold is reached.
func (s *MonitorService) autoCreateIncident(monitor *models.Monitor, errorMessage string) error {
	log.Printf("🚨 Auto-creating incident for monitor %d (component: %s) after %d consecutive failures",
		monitor.ID, monitor.ComponentID, monitor.ConsecutiveFailures)

	// Generate incident ID (UUID)
	incidentID := uuid.New()

	// Create auto_incident tracking record
	autoIncident := &models.AutoIncident{
		TenantID:     monitor.TenantID,
		MonitorID:    monitor.ID,
		IncidentID:   incidentID,
		ComponentID:  monitor.ComponentID,
		FailureCount: monitor.ConsecutiveFailures,
		ErrorMessage: errorMessage,
		Resolved:     false,
	}

	if err := s.db.Create(autoIncident).Error; err != nil {
		return fmt.Errorf("failed to create auto_incident record: %w", err)
	}

	// Update monitor with last incident ID
	monitor.LastIncidentID = &incidentID
	if err := s.db.Save(monitor).Error; err != nil {
		return fmt.Errorf("failed to update monitor with incident ID: %w", err)
	}

	// Publish auto-incident creation event to RabbitMQ
	componentIDUint := uint(0) // Will be set by incident-service based on component_id UUID
	event := events.AutoIncidentEvent{
		TenantID:         monitor.TenantID,
		MonitorID:        monitor.ID,
		ComponentID:      &componentIDUint,
		IncidentTitle:    fmt.Sprintf("%s is down", monitor.Name),
		IncidentSeverity: "major",
		FailureCount:     monitor.ConsecutiveFailures,
	}

	if err := s.eventPublisher.PublishAutoIncident(event); err != nil {
		log.Printf("❌ Failed to publish auto-incident event: %v", err)
		// Don't return error - incident was still created in tracking table
	} else {
		log.Printf("✅ Published auto-incident creation event for incident %s", incidentID)
	}

	return nil
}

// autoResolveIncident resolves an auto-created incident when checks pass again.
func (s *MonitorService) autoResolveIncident(monitor *models.Monitor) error {
	if monitor.LastIncidentID == nil {
		return nil
	}

	log.Printf("✅ Auto-resolving incident %s for monitor %d (checks passing)", *monitor.LastIncidentID, monitor.ID)

	// Find the auto_incident record
	var autoIncident models.AutoIncident
	if err := s.db.Where("incident_id = ? AND resolved = ?", *monitor.LastIncidentID, false).
		First(&autoIncident).Error; err != nil {
		return fmt.Errorf("auto_incident record not found: %w", err)
	}

	// Mark as resolved
	now := time.Now()
	autoIncident.Resolved = true
	autoIncident.ResolvedAt = &now
	autoIncident.AutoResolved = true

	if err := s.db.Save(&autoIncident).Error; err != nil {
		return fmt.Errorf("failed to update auto_incident: %w", err)
	}

	// Clear monitor's last incident ID
	monitor.LastIncidentID = nil
	if err := s.db.Save(monitor).Error; err != nil {
		return fmt.Errorf("failed to clear monitor incident ID: %w", err)
	}

	// Publish auto-resolve event to RabbitMQ is handled by incident-service
	log.Printf("✅ Auto-resolved incident %s for monitor %d", autoIncident.IncidentID, monitor.ID)

	return nil
}

// recordStatusHistory records a status change in the history table.
func (s *MonitorService) recordStatusHistory(monitor *models.Monitor, newStatus, triggeredBy, errorMessage string) {
	// Get the previous status from the last history entry
	var lastHistory models.MonitorStatusHistory
	err := s.db.Where("monitor_id = ?", monitor.ID).
		Order("changed_at DESC").
		First(&lastHistory).Error

	previousStatus := "unknown"
	var durationSeconds *int64

	if err == nil {
		previousStatus = lastHistory.NewStatus
		duration := time.Since(lastHistory.ChangedAt).Seconds()
		durationInt := int64(duration)
		durationSeconds = &durationInt
	}

	history := &models.MonitorStatusHistory{
		MonitorID:       monitor.ID,
		TenantID:        monitor.TenantID,
		PreviousStatus:  previousStatus,
		NewStatus:       newStatus,
		ChangedAt:       time.Now(),
		DurationSeconds: durationSeconds,
		TriggeredBy:     triggeredBy,
		ErrorMessage:    errorMessage,
	}

	if err := s.db.Create(history).Error; err != nil {
		log.Printf("❌ Failed to record status history: %v", err)
	}
}

// UpdateMonitor updates an existing monitor.
func (s *MonitorService) UpdateMonitor(monitor *models.Monitor) error {
	return s.db.Save(monitor).Error
}

// DeleteMonitor soft deletes a monitor.
func (s *MonitorService) DeleteMonitor(id uint) error {
	now := time.Now()
	return s.db.Model(&models.Monitor{}).Where("id = ?", id).Update("deleted_at", now).Error
}

// GetMonitorsNeedingCheck returns monitors that are due for a health check.
func (s *MonitorService) GetMonitorsNeedingCheck() ([]*models.Monitor, error) {
	var monitors []*models.Monitor

	// Get monitors where:
	// - is_active = true
	// - deleted_at IS NULL
	// - last_check_at IS NULL OR (NOW() - last_check_at) >= check_interval_seconds
	err := s.db.Where(`
		is_active = ? AND deleted_at IS NULL AND in_maintenance = ? AND (
			last_check_at IS NULL OR
			EXTRACT(EPOCH FROM (NOW() - last_check_at)) >= check_interval_seconds
		)
	`, true, false).Find(&monitors).Error

	if err != nil {
		return nil, err
	}

	return monitors, nil
}

// StartMaintenanceWindow activates a maintenance window and updates affected monitors.
func (s *MonitorService) StartMaintenanceWindow(windowID uint) error {
	var window models.MaintenanceWindow
	if err := s.db.First(&window, windowID).Error; err != nil {
		return err
	}

	window.IsActive = true
	if err := s.db.Save(&window).Error; err != nil {
		return err
	}

	// Update affected monitors
	// Parse affected_monitors JSON array and set in_maintenance = true
	// For now, simplified implementation - set all tenant monitors to maintenance
	return s.db.Model(&models.Monitor{}).
		Where("tenant_id = ? AND deleted_at IS NULL", window.TenantID).
		Updates(map[string]interface{}{
			"in_maintenance":    true,
			"maintenance_until": window.EndsAt,
		}).Error
}

// EndMaintenanceWindow deactivates a maintenance window and restores affected monitors.
func (s *MonitorService) EndMaintenanceWindow(windowID uint) error {
	var window models.MaintenanceWindow
	if err := s.db.First(&window, windowID).Error; err != nil {
		return err
	}

	window.IsActive = false
	if err := s.db.Save(&window).Error; err != nil {
		return err
	}

	// Restore affected monitors
	return s.db.Model(&models.Monitor{}).
		Where("tenant_id = ? AND deleted_at IS NULL AND in_maintenance = ?", window.TenantID, true).
		Updates(map[string]interface{}{
			"in_maintenance":    false,
			"maintenance_until": nil,
		}).Error
}
