// Package services provides maintenance management business logic.
package services

import (
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage-monitoring-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MaintenanceManagementService handles maintenance management business logic.
type MaintenanceManagementService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMaintenanceManagementService creates a new maintenance management service.
func NewMaintenanceManagementService(db *gorm.DB, logger *zap.Logger) *MaintenanceManagementService {
	return &MaintenanceManagementService{
		db:     db,
		logger: logger,
	}
}

// CreateMaintenanceWindow creates a new maintenance window.
func (s *MaintenanceManagementService) CreateMaintenanceWindow(maintenance *models.MaintenanceWindow) error {
	if err := maintenance.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values
	if maintenance.Status == "" {
		maintenance.Status = "scheduled"
	}

	if err := s.db.Create(maintenance).Error; err != nil {
		s.logger.Error("Failed to create maintenance window", zap.Error(err))
		return fmt.Errorf("failed to create maintenance window: %w", err)
	}

	s.logger.Info("Maintenance window created successfully", zap.Uint("maintenance_id", maintenance.ID))
	return nil
}

// GetMaintenanceWindow retrieves a maintenance window by ID.
func (s *MaintenanceManagementService) GetMaintenanceWindow(id uint) (*models.MaintenanceWindow, error) {
	var maintenance models.MaintenanceWindow
	if err := s.db.Preload("Components").Preload("Components.Component").Preload("Updates").First(&maintenance, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("maintenance window not found")
		}
		s.logger.Error("Failed to get maintenance window", zap.Error(err))
		return nil, fmt.Errorf("failed to get maintenance window: %w", err)
	}

	return &maintenance, nil
}

// GetMaintenanceWindows retrieves a list of maintenance windows with pagination.
func (s *MaintenanceManagementService) GetMaintenanceWindows(tenantID uint, limit, offset int) ([]*models.MaintenanceWindow, int64, error) {
	var maintenanceWindows []*models.MaintenanceWindow
	var total int64

	// Get total count
	if err := s.db.Model(&models.MaintenanceWindow{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count maintenance windows", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count maintenance windows: %w", err)
	}

	// Get maintenance windows with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).
		Preload("Components").
		Preload("Components.Component").
		Preload("Updates").
		Limit(limit).
		Offset(offset).
		Order("start_time DESC").
		Find(&maintenanceWindows).Error; err != nil {
		s.logger.Error("Failed to get maintenance windows", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get maintenance windows: %w", err)
	}

	return maintenanceWindows, total, nil
}

// GetUpcomingMaintenance retrieves upcoming maintenance windows.
func (s *MaintenanceManagementService) GetUpcomingMaintenance(tenantID uint, limit int) ([]*models.MaintenanceWindow, error) {
	var maintenanceWindows []*models.MaintenanceWindow

	now := time.Now()
	if err := s.db.Where("tenant_id = ? AND start_time > ? AND status = ?", tenantID, now, "scheduled").
		Preload("Components").
		Preload("Components.Component").
		Preload("Updates").
		Order("start_time ASC").
		Limit(limit).
		Find(&maintenanceWindows).Error; err != nil {
		s.logger.Error("Failed to get upcoming maintenance", zap.Error(err))
		return nil, fmt.Errorf("failed to get upcoming maintenance: %w", err)
	}

	return maintenanceWindows, nil
}

// GetActiveMaintenance retrieves currently active maintenance windows.
func (s *MaintenanceManagementService) GetActiveMaintenance(tenantID uint) ([]*models.MaintenanceWindow, error) {
	var maintenanceWindows []*models.MaintenanceWindow

	now := time.Now()
	if err := s.db.Where("tenant_id = ? AND start_time <= ? AND end_time >= ? AND status = ?",
		tenantID, now, now, "in_progress").
		Preload("Components").
		Preload("Components.Component").
		Preload("Updates").
		Order("start_time ASC").
		Find(&maintenanceWindows).Error; err != nil {
		s.logger.Error("Failed to get active maintenance", zap.Error(err))
		return nil, fmt.Errorf("failed to get active maintenance: %w", err)
	}

	return maintenanceWindows, nil
}

// UpdateMaintenanceWindow updates a maintenance window.
func (s *MaintenanceManagementService) UpdateMaintenanceWindow(maintenance *models.MaintenanceWindow) error {
	if err := maintenance.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Save(maintenance).Error; err != nil {
		s.logger.Error("Failed to update maintenance window", zap.Error(err))
		return fmt.Errorf("failed to update maintenance window: %w", err)
	}

	s.logger.Info("Maintenance window updated successfully", zap.Uint("maintenance_id", maintenance.ID))
	return nil
}

// StartMaintenance starts a maintenance window.
func (s *MaintenanceManagementService) StartMaintenance(maintenanceID uint) error {
	now := time.Now()
	if err := s.db.Model(&models.MaintenanceWindow{}).
		Where("id = ?", maintenanceID).
		Updates(map[string]interface{}{
			"status":     "in_progress",
			"updated_at": now,
		}).Error; err != nil {
		s.logger.Error("Failed to start maintenance", zap.Error(err))
		return fmt.Errorf("failed to start maintenance: %w", err)
	}

	s.logger.Info("Maintenance started successfully", zap.Uint("maintenance_id", maintenanceID))
	return nil
}

// CompleteMaintenance completes a maintenance window.
func (s *MaintenanceManagementService) CompleteMaintenance(maintenanceID uint) error {
	now := time.Now()
	if err := s.db.Model(&models.MaintenanceWindow{}).
		Where("id = ?", maintenanceID).
		Updates(map[string]interface{}{
			"status":     "completed",
			"updated_at": now,
		}).Error; err != nil {
		s.logger.Error("Failed to complete maintenance", zap.Error(err))
		return fmt.Errorf("failed to complete maintenance: %w", err)
	}

	s.logger.Info("Maintenance completed successfully", zap.Uint("maintenance_id", maintenanceID))
	return nil
}

// CancelMaintenance cancels a maintenance window.
func (s *MaintenanceManagementService) CancelMaintenance(maintenanceID uint) error {
	now := time.Now()
	if err := s.db.Model(&models.MaintenanceWindow{}).
		Where("id = ?", maintenanceID).
		Updates(map[string]interface{}{
			"status":     "cancelled",
			"updated_at": now,
		}).Error; err != nil {
		s.logger.Error("Failed to cancel maintenance", zap.Error(err))
		return fmt.Errorf("failed to cancel maintenance: %w", err)
	}

	s.logger.Info("Maintenance cancelled successfully", zap.Uint("maintenance_id", maintenanceID))
	return nil
}

// CreateMaintenanceUpdate creates a new maintenance update.
func (s *MaintenanceManagementService) CreateMaintenanceUpdate(update *models.MaintenanceUpdate) error {
	if err := update.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Create(update).Error; err != nil {
		s.logger.Error("Failed to create maintenance update", zap.Error(err))
		return fmt.Errorf("failed to create maintenance update: %w", err)
	}

	// Update the maintenance status if the update status is different
	var maintenance models.MaintenanceWindow
	if err := s.db.First(&maintenance, update.MaintenanceID).Error; err != nil {
		s.logger.Error("Failed to get maintenance for status update", zap.Error(err))
		return fmt.Errorf("failed to get maintenance for status update: %w", err)
	}

	if maintenance.Status != update.Status {
		maintenance.Status = update.Status
		if err := s.db.Save(&maintenance).Error; err != nil {
			s.logger.Error("Failed to update maintenance status", zap.Error(err))
			return fmt.Errorf("failed to update maintenance status: %w", err)
		}
	}

	s.logger.Info("Maintenance update created successfully", zap.Uint("update_id", update.ID))
	return nil
}

// GetMaintenanceUpdates retrieves updates for a maintenance window.
func (s *MaintenanceManagementService) GetMaintenanceUpdates(maintenanceID uint, limit, offset int) ([]*models.MaintenanceUpdate, int64, error) {
	var updates []*models.MaintenanceUpdate
	var total int64

	// Get total count
	if err := s.db.Model(&models.MaintenanceUpdate{}).Where("maintenance_id = ?", maintenanceID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count maintenance updates", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count maintenance updates: %w", err)
	}

	// Get updates with pagination
	if err := s.db.Where("maintenance_id = ?", maintenanceID).
		Preload("Maintenance").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&updates).Error; err != nil {
		s.logger.Error("Failed to get maintenance updates", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get maintenance updates: %w", err)
	}

	return updates, total, nil
}

// AddMaintenanceComponent adds a component to a maintenance window.
func (s *MaintenanceManagementService) AddMaintenanceComponent(maintenanceComponent *models.MaintenanceComponent) error {
	if err := maintenanceComponent.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Create(maintenanceComponent).Error; err != nil {
		s.logger.Error("Failed to add maintenance component", zap.Error(err))
		return fmt.Errorf("failed to add maintenance component: %w", err)
	}

	s.logger.Info("Maintenance component added successfully", zap.Uint("maintenance_component_id", maintenanceComponent.ID))
	return nil
}

// RemoveMaintenanceComponent removes a component from a maintenance window.
func (s *MaintenanceManagementService) RemoveMaintenanceComponent(maintenanceID, componentID uint) error {
	if err := s.db.Where("maintenance_id = ? AND component_id = ?", maintenanceID, componentID).
		Delete(&models.MaintenanceComponent{}).Error; err != nil {
		s.logger.Error("Failed to remove maintenance component", zap.Error(err))
		return fmt.Errorf("failed to remove maintenance component: %w", err)
	}

	s.logger.Info("Maintenance component removed successfully", zap.Uint("maintenance_id", maintenanceID), zap.Uint("component_id", componentID))
	return nil
}

// CreateMaintenanceTemplate creates a new maintenance template.
func (s *MaintenanceManagementService) CreateMaintenanceTemplate(template *models.MaintenanceTemplate) error {
	if err := template.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Create(template).Error; err != nil {
		s.logger.Error("Failed to create maintenance template", zap.Error(err))
		return fmt.Errorf("failed to create maintenance template: %w", err)
	}

	s.logger.Info("Maintenance template created successfully", zap.Uint("template_id", template.ID))
	return nil
}

// GetMaintenanceTemplates retrieves maintenance templates for a tenant.
func (s *MaintenanceManagementService) GetMaintenanceTemplates(tenantID uint) ([]*models.MaintenanceTemplate, error) {
	var templates []*models.MaintenanceTemplate

	if err := s.db.Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Order("name ASC").
		Find(&templates).Error; err != nil {
		s.logger.Error("Failed to get maintenance templates", zap.Error(err))
		return nil, fmt.Errorf("failed to get maintenance templates: %w", err)
	}

	return templates, nil
}

// CreateMaintenanceFromTemplate creates a maintenance window from a template.
func (s *MaintenanceManagementService) CreateMaintenanceFromTemplate(templateID uint, createdBy uint, startTime time.Time, customizations map[string]interface{}) (*models.MaintenanceWindow, error) {
	var template models.MaintenanceTemplate
	if err := s.db.First(&template, templateID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("maintenance template not found")
		}
		s.logger.Error("Failed to get maintenance template", zap.Error(err))
		return nil, fmt.Errorf("failed to get maintenance template: %w", err)
	}

	endTime := startTime.Add(time.Duration(template.Duration) * time.Minute)
	maintenance := &models.MaintenanceWindow{
		TenantID:    template.TenantID,
		Title:       template.Title,
		Description: template.Description,
		Status:      "scheduled",
		Type:        template.Type,
		Impact:      template.Impact,
		StartTime:   startTime,
		EndTime:     endTime,
		IsActive:    true,
		CreatedBy:   createdBy,
	}

	// Apply customizations if provided
	if title, ok := customizations["title"].(string); ok && title != "" {
		maintenance.Title = title
	}
	if description, ok := customizations["description"].(string); ok && description != "" {
		maintenance.Description = description
	}
	if impact, ok := customizations["impact"].(string); ok && impact != "" {
		maintenance.Impact = impact
	}
	if start, ok := customizations["start_time"].(time.Time); ok && !start.IsZero() {
		maintenance.StartTime = start
		maintenance.EndTime = start.Add(time.Duration(template.Duration) * time.Minute)
	}

	if err := s.CreateMaintenanceWindow(maintenance); err != nil {
		return nil, fmt.Errorf("failed to create maintenance from template: %w", err)
	}

	// Create initial update with template message
	update := &models.MaintenanceUpdate{
		MaintenanceID: maintenance.ID,
		Status:        "scheduled",
		Message:       template.Message,
		IsPublic:      true,
		CreatedBy:     createdBy,
	}

	if err := s.CreateMaintenanceUpdate(update); err != nil {
		s.logger.Warn("Failed to create initial maintenance update", zap.Error(err))
		// Don't fail the maintenance creation if update fails
	}

	return maintenance, nil
}

// GetMaintenanceStatistics returns statistics for maintenance windows.
func (s *MaintenanceManagementService) GetMaintenanceStatistics(tenantID uint, startDate, endDate time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Build query
	query := s.db.Model(&models.MaintenanceWindow{}).Where("tenant_id = ?", tenantID)
	if !startDate.IsZero() {
		query = query.Where("start_time >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("start_time <= ?", endDate)
	}

	// Get total maintenance windows
	var totalMaintenance int64
	if err := query.Count(&totalMaintenance).Error; err != nil {
		s.logger.Error("Failed to count maintenance windows", zap.Error(err))
		return nil, fmt.Errorf("failed to count maintenance windows: %w", err)
	}

	// Get maintenance by status
	var statusCounts []struct {
		Status string
		Count  int64
	}
	if err := query.Select("status, COUNT(*) as count").Group("status").Scan(&statusCounts).Error; err != nil {
		s.logger.Error("Failed to get status counts", zap.Error(err))
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}

	// Get maintenance by type
	var typeCounts []struct {
		Type  string
		Count int64
	}
	if err := query.Select("type, COUNT(*) as count").Group("type").Scan(&typeCounts).Error; err != nil {
		s.logger.Error("Failed to get type counts", zap.Error(err))
		return nil, fmt.Errorf("failed to get type counts: %w", err)
	}

	// Get average duration
	var avgDuration float64
	if err := query.Where("status = ?", "completed").
		Select("AVG(EXTRACT(EPOCH FROM (end_time - start_time))/60) as avg_duration_minutes").
		Row().Scan(&avgDuration); err != nil {
		s.logger.Warn("Failed to calculate average duration", zap.Error(err))
		avgDuration = 0
	}

	stats["total_maintenance"] = totalMaintenance
	stats["status_breakdown"] = statusCounts
	stats["type_breakdown"] = typeCounts
	stats["average_duration_minutes"] = avgDuration
	stats["date_range"] = map[string]interface{}{
		"start": startDate,
		"end":   endDate,
	}
	stats["timestamp"] = time.Now()

	return stats, nil
}

