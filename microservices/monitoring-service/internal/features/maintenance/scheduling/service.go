// Package services provides maintenance window management for alert suppression.
package scheduling

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MaintenanceWindow represents a scheduled maintenance window.
type MaintenanceWindow struct {
	ID                   uint      `gorm:"primarykey" json:"id"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	TenantID             uuid.UUID `gorm:"type:uuid;not null;index:idx_maintenance_tenant" json:"tenant_id"`
	Name                 string    `gorm:"size:255;not null" json:"name"`
	Description          string    `gorm:"type:text" json:"description,omitempty"`
	StartsAt             time.Time `gorm:"not null;index:idx_maintenance_schedule" json:"starts_at"`
	EndsAt               time.Time `gorm:"not null;index:idx_maintenance_schedule" json:"ends_at"`
	AffectedMonitors     string    `gorm:"type:text" json:"affected_monitors,omitempty"` // JSON array: [1, 2, 3]
	IsActive             bool      `gorm:"default:false;index:idx_maintenance_active" json:"is_active"`
	SuppressNotifications bool     `gorm:"default:true" json:"suppress_notifications"`
	AutoUpdateStatusPage  bool     `gorm:"default:true" json:"auto_update_status_page"`
	CreatedBy            uuid.UUID `gorm:"type:uuid" json:"created_by,omitempty"`
}

// TableName specifies the table name for GORM.
func (MaintenanceWindow) TableName() string {
	return "maintenance_windows"
}

// IsCurrentlyActive checks if the maintenance window is currently active.
func (m *MaintenanceWindow) IsCurrentlyActive() bool {
	now := time.Now()
	return m.IsActive && now.After(m.StartsAt) && now.Before(m.EndsAt)
}

// GetAffectedMonitorIDs returns the list of affected monitor IDs.
func (m *MaintenanceWindow) GetAffectedMonitorIDs() ([]uint, error) {
	if m.AffectedMonitors == "" || m.AffectedMonitors == "null" {
		return nil, nil // nil means all monitors
	}

	var monitorIDs []uint
	if err := json.Unmarshal([]byte(m.AffectedMonitors), &monitorIDs); err != nil {
		return nil, fmt.Errorf("failed to parse affected monitors: %w", err)
	}

	return monitorIDs, nil
}

// MaintenanceService handles maintenance window operations.
type MaintenanceService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMaintenanceService creates a new maintenance service.
func NewMaintenanceService(db *gorm.DB, logger *zap.Logger) *MaintenanceService {
	// Auto-migrate maintenance windows table
	if err := db.AutoMigrate(&MaintenanceWindow{}); err != nil {
		logger.Error("Failed to migrate maintenance_windows table", zap.Error(err))
	}

	return &MaintenanceService{
		db:     db,
		logger: logger,
	}
}

// CreateMaintenanceWindow creates a new maintenance window.
func (s *MaintenanceService) CreateMaintenanceWindow(tenantID, createdBy uuid.UUID, window *MaintenanceWindow) error {
	// Validate dates
	if window.EndsAt.Before(window.StartsAt) {
		return fmt.Errorf("end time must be after start time")
	}

	window.TenantID = tenantID
	window.CreatedBy = createdBy

	// Auto-activate if window starts immediately
	if window.StartsAt.Before(time.Now()) && window.EndsAt.After(time.Now()) {
		window.IsActive = true
	}

	if err := s.db.Create(window).Error; err != nil {
		s.logger.Error("Failed to create maintenance window", zap.Error(err))
		return fmt.Errorf("failed to create maintenance window: %w", err)
	}

	s.logger.Info("Maintenance window created",
		zap.Uint("window_id", window.ID),
		zap.String("name", window.Name),
		zap.Time("starts_at", window.StartsAt),
		zap.Time("ends_at", window.EndsAt),
	)

	return nil
}

// GetMaintenanceWindows retrieves all maintenance windows for a tenant.
func (s *MaintenanceService) GetMaintenanceWindows(tenantID uuid.UUID, includeInactive bool) ([]MaintenanceWindow, error) {
	var windows []MaintenanceWindow
	query := s.db.Where("tenant_id = ?", tenantID)

	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Order("starts_at DESC").Find(&windows).Error; err != nil {
		s.logger.Error("Failed to get maintenance windows", zap.Error(err))
		return nil, fmt.Errorf("failed to get maintenance windows: %w", err)
	}

	return windows, nil
}

// GetMaintenanceWindow retrieves a specific maintenance window by ID.
func (s *MaintenanceService) GetMaintenanceWindow(id uint, tenantID uuid.UUID) (*MaintenanceWindow, error) {
	var window MaintenanceWindow
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&window).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("maintenance window not found")
		}
		s.logger.Error("Failed to get maintenance window", zap.Error(err))
		return nil, fmt.Errorf("failed to get maintenance window: %w", err)
	}

	return &window, nil
}

// UpdateMaintenanceWindow updates a maintenance window.
func (s *MaintenanceService) UpdateMaintenanceWindow(id uint, tenantID uuid.UUID, updates map[string]interface{}) error {
	result := s.db.Model(&MaintenanceWindow{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(updates)

	if result.Error != nil {
		s.logger.Error("Failed to update maintenance window", zap.Error(result.Error))
		return fmt.Errorf("failed to update maintenance window: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("maintenance window not found")
	}

	s.logger.Info("Maintenance window updated", zap.Uint("window_id", id))
	return nil
}

// DeleteMaintenanceWindow deletes a maintenance window.
func (s *MaintenanceService) DeleteMaintenanceWindow(id uint, tenantID uuid.UUID) error {
	result := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&MaintenanceWindow{})

	if result.Error != nil {
		s.logger.Error("Failed to delete maintenance window", zap.Error(result.Error))
		return fmt.Errorf("failed to delete maintenance window: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("maintenance window not found")
	}

	s.logger.Info("Maintenance window deleted", zap.Uint("window_id", id))
	return nil
}

// GetActiveMaintenanceWindows retrieves all currently active maintenance windows.
func (s *MaintenanceService) GetActiveMaintenanceWindows(tenantID uuid.UUID) ([]MaintenanceWindow, error) {
	var windows []MaintenanceWindow
	now := time.Now()

	if err := s.db.Where("tenant_id = ? AND is_active = ? AND starts_at <= ? AND ends_at >= ?",
		tenantID, true, now, now).Find(&windows).Error; err != nil {
		s.logger.Error("Failed to get active maintenance windows", zap.Error(err))
		return nil, fmt.Errorf("failed to get active maintenance windows: %w", err)
	}

	return windows, nil
}

// IsMonitorInMaintenance checks if a specific monitor is currently in a maintenance window.
func (s *MaintenanceService) IsMonitorInMaintenance(tenantID uuid.UUID, monitorID uint) (bool, error) {
	windows, err := s.GetActiveMaintenanceWindows(tenantID)
	if err != nil {
		return false, err
	}

	for _, window := range windows {
		if !window.SuppressNotifications {
			continue
		}

		// If no affected monitors specified, all monitors are affected
		if window.AffectedMonitors == "" || window.AffectedMonitors == "null" {
			s.logger.Info("Monitor is in maintenance (all monitors)",
				zap.Uint("monitor_id", monitorID),
				zap.Uint("window_id", window.ID),
			)
			return true, nil
		}

		// Check if this monitor is in the affected list
		affectedIDs, err := window.GetAffectedMonitorIDs()
		if err != nil {
			s.logger.Error("Failed to parse affected monitors", zap.Error(err))
			continue
		}

		for _, affectedID := range affectedIDs {
			if affectedID == monitorID {
				s.logger.Info("Monitor is in maintenance",
					zap.Uint("monitor_id", monitorID),
					zap.Uint("window_id", window.ID),
				)
				return true, nil
			}
		}
	}

	return false, nil
}

// ActivateMaintenanceWindow manually activates a maintenance window.
func (s *MaintenanceService) ActivateMaintenanceWindow(id uint, tenantID uuid.UUID) error {
	return s.UpdateMaintenanceWindow(id, tenantID, map[string]interface{}{
		"is_active": true,
	})
}

// DeactivateMaintenanceWindow manually deactivates a maintenance window.
func (s *MaintenanceService) DeactivateMaintenanceWindow(id uint, tenantID uuid.UUID) error {
	return s.UpdateMaintenanceWindow(id, tenantID, map[string]interface{}{
		"is_active": false,
	})
}

// AutoActivateExpiredWindows automatically activates/deactivates maintenance windows based on schedule.
// This should be run periodically (e.g., every minute) via a background job.
func (s *MaintenanceService) AutoActivateExpiredWindows() error {
	now := time.Now()

	// Activate windows that should be active now
	result := s.db.Model(&MaintenanceWindow{}).
		Where("is_active = ? AND starts_at <= ? AND ends_at >= ?", false, now, now).
		Update("is_active", true)

	if result.Error != nil {
		s.logger.Error("Failed to auto-activate maintenance windows", zap.Error(result.Error))
		return result.Error
	}

	if result.RowsAffected > 0 {
		s.logger.Info("Auto-activated maintenance windows", zap.Int64("count", result.RowsAffected))
	}

	// Deactivate windows that have expired
	result = s.db.Model(&MaintenanceWindow{}).
		Where("is_active = ? AND ends_at < ?", true, now).
		Update("is_active", false)

	if result.Error != nil {
		s.logger.Error("Failed to auto-deactivate maintenance windows", zap.Error(result.Error))
		return result.Error
	}

	if result.RowsAffected > 0 {
		s.logger.Info("Auto-deactivated maintenance windows", zap.Int64("count", result.RowsAffected))
	}

	return nil
}

// GetUpcomingMaintenanceWindows retrieves maintenance windows scheduled to start within the next N hours.
func (s *MaintenanceService) GetUpcomingMaintenanceWindows(tenantID uuid.UUID, hoursAhead int) ([]MaintenanceWindow, error) {
	var windows []MaintenanceWindow
	now := time.Now()
	futureTime := now.Add(time.Duration(hoursAhead) * time.Hour)

	if err := s.db.Where("tenant_id = ? AND starts_at >= ? AND starts_at <= ?",
		tenantID, now, futureTime).Order("starts_at ASC").Find(&windows).Error; err != nil {
		s.logger.Error("Failed to get upcoming maintenance windows", zap.Error(err))
		return nil, fmt.Errorf("failed to get upcoming maintenance windows: %w", err)
	}

	return windows, nil
}
