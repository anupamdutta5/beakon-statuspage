// Package services provides status automation integration business logic.
package services

import (
	"fmt"
	"time"

	"github.com/anupamdutta5/monitoring-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// StatusAutomationService handles status automation integration business logic.
type StatusAutomationService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewStatusAutomationService creates a new status automation service.
func NewStatusAutomationService(db *gorm.DB, logger *zap.Logger) *StatusAutomationService {
	return &StatusAutomationService{
		db:     db,
		logger: logger,
	}
}

// CreateStatusAutomation creates a new status automation.
func (s *StatusAutomationService) CreateStatusAutomation(automation *models.StatusAutomation) error {
	if err := automation.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Create(automation).Error; err != nil {
		s.logger.Error("Failed to create status automation", zap.Error(err))
		return fmt.Errorf("failed to create status automation: %w", err)
	}

	s.logger.Info("Status automation created successfully", zap.Uint("automation_id", automation.ID))
	return nil
}

// GetStatusAutomation retrieves a status automation by ID.
func (s *StatusAutomationService) GetStatusAutomation(id uint) (*models.StatusAutomation, error) {
	var automation models.StatusAutomation
	if err := s.db.Preload("Integrations").First(&automation, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("status automation not found")
		}
		s.logger.Error("Failed to get status automation", zap.Error(err))
		return nil, fmt.Errorf("failed to get status automation: %w", err)
	}

	return &automation, nil
}

// GetStatusAutomations retrieves a list of status automations with pagination.
func (s *StatusAutomationService) GetStatusAutomations(tenantID uint, limit, offset int) ([]*models.StatusAutomation, int64, error) {
	var automations []*models.StatusAutomation
	var total int64

	// Get total count
	if err := s.db.Model(&models.StatusAutomation{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count status automations", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count status automations: %w", err)
	}

	// Get automations with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).
		Preload("Integrations").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&automations).Error; err != nil {
		s.logger.Error("Failed to get status automations", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get status automations: %w", err)
	}

	return automations, total, nil
}

// UpdateStatusAutomation updates a status automation.
func (s *StatusAutomationService) UpdateStatusAutomation(automation *models.StatusAutomation) error {
	if err := automation.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Save(automation).Error; err != nil {
		s.logger.Error("Failed to update status automation", zap.Error(err))
		return fmt.Errorf("failed to update status automation: %w", err)
	}

	s.logger.Info("Status automation updated successfully", zap.Uint("automation_id", automation.ID))
	return nil
}

// DeleteStatusAutomation soft deletes a status automation.
func (s *StatusAutomationService) DeleteStatusAutomation(id uint) error {
	if err := s.db.Delete(&models.StatusAutomation{}, id).Error; err != nil {
		s.logger.Error("Failed to delete status automation", zap.Error(err))
		return fmt.Errorf("failed to delete status automation: %w", err)
	}

	s.logger.Info("Status automation deleted successfully", zap.Uint("automation_id", id))
	return nil
}

// CreateStatusAutomationIntegration creates a new status automation integration.
func (s *StatusAutomationService) CreateStatusAutomationIntegration(integration *models.StatusAutomationIntegration) error {
	if err := integration.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Create(integration).Error; err != nil {
		s.logger.Error("Failed to create status automation integration", zap.Error(err))
		return fmt.Errorf("failed to create status automation integration: %w", err)
	}

	s.logger.Info("Status automation integration created successfully", zap.Uint("integration_id", integration.ID))
	return nil
}

// GetStatusAutomationIntegration retrieves a status automation integration by ID.
func (s *StatusAutomationService) GetStatusAutomationIntegration(id uint) (*models.StatusAutomationIntegration, error) {
	var integration models.StatusAutomationIntegration
	if err := s.db.Preload("Automation").Preload("Service").First(&integration, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("status automation integration not found")
		}
		s.logger.Error("Failed to get status automation integration", zap.Error(err))
		return nil, fmt.Errorf("failed to get status automation integration: %w", err)
	}

	return &integration, nil
}

// UpdateStatusAutomationIntegration updates a status automation integration.
func (s *StatusAutomationService) UpdateStatusAutomationIntegration(integration *models.StatusAutomationIntegration) error {
	if err := integration.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Save(integration).Error; err != nil {
		s.logger.Error("Failed to update status automation integration", zap.Error(err))
		return fmt.Errorf("failed to update status automation integration: %w", err)
	}

	s.logger.Info("Status automation integration updated successfully", zap.Uint("integration_id", integration.ID))
	return nil
}

// DeleteStatusAutomationIntegration deletes a status automation integration.
func (s *StatusAutomationService) DeleteStatusAutomationIntegration(id uint) error {
	if err := s.db.Delete(&models.StatusAutomationIntegration{}, id).Error; err != nil {
		s.logger.Error("Failed to delete status automation integration", zap.Error(err))
		return fmt.Errorf("failed to delete status automation integration: %w", err)
	}

	s.logger.Info("Status automation integration deleted successfully", zap.Uint("integration_id", id))
	return nil
}

// GetStatusAutomationIntegrationsByService retrieves all integrations for a specific service.
func (s *StatusAutomationService) GetStatusAutomationIntegrationsByService(serviceID uint) ([]*models.StatusAutomationIntegration, error) {
	var integrations []*models.StatusAutomationIntegration

	if err := s.db.Where("service_id = ?", serviceID).
		Preload("Automation").
		Preload("Service").
		Find(&integrations).Error; err != nil {
		s.logger.Error("Failed to get status automation integrations by service", zap.Error(err))
		return nil, fmt.Errorf("failed to get status automation integrations by service: %w", err)
	}

	return integrations, nil
}

// GetStatusAutomationIntegrationsByType retrieves all integrations for a specific automation type.
func (s *StatusAutomationService) GetStatusAutomationIntegrationsByType(tenantID uint, automationType string) ([]*models.StatusAutomationIntegration, error) {
	var integrations []*models.StatusAutomationIntegration

	if err := s.db.Joins("JOIN status_automations ON status_automation_integrations.automation_id = status_automations.id").
		Where("status_automations.tenant_id = ? AND status_automations.type = ?", tenantID, automationType).
		Preload("Automation").
		Preload("Service").
		Find(&integrations).Error; err != nil {
		s.logger.Error("Failed to get status automation integrations by type", zap.Error(err))
		return nil, fmt.Errorf("failed to get status automation integrations by type: %w", err)
	}

	return integrations, nil
}

// UpdateLastSync updates the last sync time for a status automation.
func (s *StatusAutomationService) UpdateLastSync(automationID uint) error {
	now := time.Now()
	if err := s.db.Model(&models.StatusAutomation{}).
		Where("id = ?", automationID).
		Update("last_sync", now).Error; err != nil {
		s.logger.Error("Failed to update last sync", zap.Error(err))
		return fmt.Errorf("failed to update last sync: %w", err)
	}

	s.logger.Info("Last sync updated", zap.Uint("automation_id", automationID), zap.Time("last_sync", now))
	return nil
}

// UpdateIntegrationLastSync updates the last sync time for a status automation integration.
func (s *StatusAutomationService) UpdateIntegrationLastSync(integrationID uint) error {
	now := time.Now()
	if err := s.db.Model(&models.StatusAutomationIntegration{}).
		Where("id = ?", integrationID).
		Update("last_sync", now).Error; err != nil {
		s.logger.Error("Failed to update integration last sync", zap.Error(err))
		return fmt.Errorf("failed to update integration last sync: %w", err)
	}

	s.logger.Info("Integration last sync updated", zap.Uint("integration_id", integrationID), zap.Time("last_sync", now))
	return nil
}

// GetActiveStatusAutomations retrieves all active status automations for a tenant.
func (s *StatusAutomationService) GetActiveStatusAutomations(tenantID uint) ([]*models.StatusAutomation, error) {
	var automations []*models.StatusAutomation

	if err := s.db.Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Preload("Integrations").
		Find(&automations).Error; err != nil {
		s.logger.Error("Failed to get active status automations", zap.Error(err))
		return nil, fmt.Errorf("failed to get active status automations: %w", err)
	}

	return automations, nil
}

