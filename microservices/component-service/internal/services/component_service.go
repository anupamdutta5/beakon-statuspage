// Package services provides business logic for the Component Service.
package services

import (
	"fmt"
	"time"

	"github.com/anupamdutta5/component-service/internal/config"
	"github.com/anupamdutta5/component-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ComponentService handles component-related business logic.
type ComponentService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewComponentService creates a new component service.
func NewComponentService(db *gorm.DB, logger *zap.Logger) *ComponentService {
	// Auto-migrate models for component service
	if err := db.AutoMigrate(
		&models.Component{},
		&models.ComponentGroup{},
		&models.ComponentStatus{},
		&models.ComponentHistory{},
		&models.ComponentMetric{},
		&models.ComponentAlert{},
		&models.ComponentWebhook{},
	); err != nil {
		logger.Error("Failed to migrate component service database", zap.Error(err))
	} else {
		logger.Info("Component service database migration completed successfully")
	}

	return &ComponentService{
		db:     db,
		logger: logger,
	}
}

// CreateComponent creates a new component.
func (s *ComponentService) CreateComponent(component *models.Component) error {
	// Set default values
	if component.Status == "" {
		component.Status = "operational"
	}
	if component.IsVisible == false && component.IsVisible != true {
		component.IsVisible = true
	}

	// Create component
	if err := s.db.Create(component).Error; err != nil {
		s.logger.Error("Failed to create component", zap.Error(err))
		return fmt.Errorf("failed to create component: %w", err)
	}

	// Create initial status record
	status := &models.ComponentStatus{
		ComponentID: component.ID,
		Status:      component.Status,
		Message:     "Component created",
		UpdatedBy:   0, // System
	}

	if err := s.db.Create(status).Error; err != nil {
		s.logger.Error("Failed to create component status", zap.Error(err))
		// Don't fail component creation if status creation fails
	}

	s.logger.Info("Component created successfully", zap.Uint("component_id", component.ID))
	return nil
}

// GetComponent retrieves a component by ID.
func (s *ComponentService) GetComponent(id uint) (*models.Component, error) {
	var component models.Component
	if err := s.db.Preload("Group").First(&component, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("component not found")
		}
		s.logger.Error("Failed to get component", zap.Error(err))
		return nil, fmt.Errorf("failed to get component: %w", err)
	}

	return &component, nil
}

// GetComponents retrieves a list of components with pagination.
func (s *ComponentService) GetComponents(tenantID uint, limit, offset int) ([]*models.Component, int64, error) {
	var components []*models.Component
	var total int64

	// Get total count
	if err := s.db.Model(&models.Component{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count components", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count components: %w", err)
	}

	// Get components with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).Preload("Group").Limit(limit).Offset(offset).Order("position ASC").Find(&components).Error; err != nil {
		s.logger.Error("Failed to get components", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get components: %w", err)
	}

	return components, total, nil
}

// GetPublicComponents retrieves public components for a tenant.
func (s *ComponentService) GetPublicComponents(tenantID uint) ([]*models.Component, error) {
	var components []*models.Component
	if err := s.db.Where("tenant_id = ? AND is_visible = ?", tenantID, true).Preload("Group").Order("position ASC").Find(&components).Error; err != nil {
		s.logger.Error("Failed to get public components", zap.Error(err))
		return nil, fmt.Errorf("failed to get public components: %w", err)
	}

	return components, nil
}

// UpdateComponent updates a component.
func (s *ComponentService) UpdateComponent(component *models.Component) error {
	if err := s.db.Save(component).Error; err != nil {
		s.logger.Error("Failed to update component", zap.Error(err))
		return fmt.Errorf("failed to update component: %w", err)
	}

	s.logger.Info("Component updated successfully", zap.Uint("component_id", component.ID))
	return nil
}

// UpdateComponentStatus updates a component's status.
func (s *ComponentService) UpdateComponentStatus(componentID uint, status, message string, updatedBy uint) error {
	// Get current component
	var component models.Component
	if err := s.db.First(&component, componentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("component not found")
		}
		s.logger.Error("Failed to get component", zap.Error(err))
		return fmt.Errorf("failed to get component: %w", err)
	}

	// Get current status
	var currentStatus models.ComponentStatus
	if err := s.db.Where("component_id = ?", componentID).Order("created_at DESC").First(&currentStatus).Error; err != nil {
		s.logger.Error("Failed to get current status", zap.Error(err))
		// Continue with status update even if current status not found
	}

	// Update component status
	component.Status = status
	if err := s.db.Save(&component).Error; err != nil {
		s.logger.Error("Failed to update component status", zap.Error(err))
		return fmt.Errorf("failed to update component status: %w", err)
	}

	// Create new status record
	newStatus := &models.ComponentStatus{
		ComponentID: componentID,
		Status:      status,
		Message:     message,
		UpdatedBy:   updatedBy,
	}

	if err := s.db.Create(newStatus).Error; err != nil {
		s.logger.Error("Failed to create status record", zap.Error(err))
		// Don't fail the update if status record creation fails
	}

	// Create history record if status changed
	if currentStatus.Status != "" && currentStatus.Status != status {
		history := &models.ComponentHistory{
			ComponentID: componentID,
			OldStatus:   currentStatus.Status,
			NewStatus:   status,
			Message:     message,
			UpdatedBy:   updatedBy,
		}

		if err := s.db.Create(history).Error; err != nil {
			s.logger.Error("Failed to create history record", zap.Error(err))
			// Don't fail the update if history creation fails
		}
	}

	s.logger.Info("Component status updated successfully",
		zap.Uint("component_id", componentID),
		zap.String("old_status", currentStatus.Status),
		zap.String("new_status", status))
	return nil
}

// DeleteComponent soft deletes a component.
func (s *ComponentService) DeleteComponent(id uint) error {
	if err := s.db.Delete(&models.Component{}, id).Error; err != nil {
		s.logger.Error("Failed to delete component", zap.Error(err))
		return fmt.Errorf("failed to delete component: %w", err)
	}

	s.logger.Info("Component deleted successfully", zap.Uint("component_id", id))
	return nil
}

// GetComponentHistory retrieves component status history.
func (s *ComponentService) GetComponentHistory(componentID uint, limit, offset int) ([]*models.ComponentHistory, int64, error) {
	var history []*models.ComponentHistory
	var total int64

	// Get total count
	if err := s.db.Model(&models.ComponentHistory{}).Where("component_id = ?", componentID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count component history", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count component history: %w", err)
	}

	// Get history with pagination
	if err := s.db.Where("component_id = ?", componentID).Preload("Component").Limit(limit).Offset(offset).Order("created_at DESC").Find(&history).Error; err != nil {
		s.logger.Error("Failed to get component history", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get component history: %w", err)
	}

	return history, total, nil
}

// CreateComponentGroup creates a new component group.
func (s *ComponentService) CreateComponentGroup(group *models.ComponentGroup) error {
	// Set default values
	if group.IsVisible == false && group.IsVisible != true {
		group.IsVisible = true
	}

	// Create group
	if err := s.db.Create(group).Error; err != nil {
		s.logger.Error("Failed to create component group", zap.Error(err))
		return fmt.Errorf("failed to create component group: %w", err)
	}

	s.logger.Info("Component group created successfully", zap.Uint("group_id", group.ID))
	return nil
}

// GetComponentGroup retrieves a component group by ID.
func (s *ComponentService) GetComponentGroup(id uint) (*models.ComponentGroup, error) {
	var group models.ComponentGroup
	if err := s.db.Preload("Components").First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("component group not found")
		}
		s.logger.Error("Failed to get component group", zap.Error(err))
		return nil, fmt.Errorf("failed to get component group: %w", err)
	}

	return &group, nil
}

// GetComponentGroups retrieves a list of component groups.
func (s *ComponentService) GetComponentGroups(tenantID uint, limit, offset int) ([]*models.ComponentGroup, int64, error) {
	var groups []*models.ComponentGroup
	var total int64

	// Get total count
	if err := s.db.Model(&models.ComponentGroup{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count component groups", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count component groups: %w", err)
	}

	// Get groups with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).Preload("Components").Limit(limit).Offset(offset).Order("position ASC").Find(&groups).Error; err != nil {
		s.logger.Error("Failed to get component groups", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get component groups: %w", err)
	}

	return groups, total, nil
}

// UpdateComponentGroup updates a component group.
func (s *ComponentService) UpdateComponentGroup(group *models.ComponentGroup) error {
	if err := s.db.Save(group).Error; err != nil {
		s.logger.Error("Failed to update component group", zap.Error(err))
		return fmt.Errorf("failed to update component group: %w", err)
	}

	s.logger.Info("Component group updated successfully", zap.Uint("group_id", group.ID))
	return nil
}

// DeleteComponentGroup soft deletes a component group.
func (s *ComponentService) DeleteComponentGroup(id uint) error {
	// First, ungroup all components in this group
	if err := s.db.Model(&models.Component{}).Where("group_id = ?", id).Update("group_id", nil).Error; err != nil {
		s.logger.Error("Failed to ungroup components", zap.Error(err))
		return fmt.Errorf("failed to ungroup components: %w", err)
	}

	// Then delete the group
	if err := s.db.Delete(&models.ComponentGroup{}, id).Error; err != nil {
		s.logger.Error("Failed to delete component group", zap.Error(err))
		return fmt.Errorf("failed to delete component group: %w", err)
	}

	s.logger.Info("Component group deleted successfully", zap.Uint("group_id", id))
	return nil
}

// GetPublicStatus returns the overall status for a tenant.
func (s *ComponentService) GetPublicStatus(tenantID uint) (map[string]interface{}, error) {
	var components []models.Component
	if err := s.db.Where("tenant_id = ? AND is_visible = ?", tenantID, true).Find(&components).Error; err != nil {
		s.logger.Error("Failed to get components for status", zap.Error(err))
		return nil, fmt.Errorf("failed to get components for status: %w", err)
	}

	// Calculate overall status
	statusCounts := make(map[string]int)
	for _, component := range components {
		statusCounts[component.Status]++
	}

	// Determine overall status based on component statuses
	overallStatus := "operational"
	if statusCounts["major_outage"] > 0 {
		overallStatus = "major_outage"
	} else if statusCounts["partial_outage"] > 0 {
		overallStatus = "partial_outage"
	} else if statusCounts["degraded_performance"] > 0 {
		overallStatus = "degraded_performance"
	} else if statusCounts["maintenance"] > 0 {
		overallStatus = "maintenance"
	}

	status := map[string]interface{}{
		"overall_status": overallStatus,
		"components":     components,
		"status_counts":  statusCounts,
		"last_updated":   time.Now().UTC(),
	}

	return status, nil
}

// InitDatabase initializes the database connection and runs migrations.
func InitDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.Component{},
		&models.ComponentGroup{},
		&models.ComponentStatus{},
		&models.ComponentHistory{},
		&models.ComponentMetric{},
		&models.ComponentAlert{},
		&models.ComponentWebhook{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
