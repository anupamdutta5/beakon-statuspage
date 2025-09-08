package services

import (
	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TemplateService struct {
	db *gorm.DB
}

func NewTemplateService() *TemplateService {
	return &TemplateService{
		db: database.GetDB(),
	}
}

// Incident Template Methods

// CreateIncidentTemplate creates a new incident template
func (s *TemplateService) CreateIncidentTemplate(template *models.IncidentTemplate, serviceIDs []uint) error {
	if err := s.db.Create(template).Error; err != nil {
		logger.Error("Failed to create incident template", zap.Error(err))
		return err
	}

	// Associate with services if provided
	if len(serviceIDs) > 0 {
		var services []models.Service
		if err := s.db.Where("id IN ?", serviceIDs).Find(&services).Error; err != nil {
			logger.Error("Failed to find services for template", zap.Error(err))
			return err
		}

		if err := s.db.Model(template).Association("Services").Append(services); err != nil {
			logger.Error("Failed to associate services with template", zap.Error(err))
			return err
		}
	}

	logger.Info("Incident template created successfully", zap.String("name", template.Name))
	return nil
}

// GetAllIncidentTemplates retrieves all incident templates
func (s *TemplateService) GetAllIncidentTemplates() ([]models.IncidentTemplate, error) {
	var templates []models.IncidentTemplate
	if err := s.db.Preload("Services").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

// GetIncidentTemplateByID retrieves an incident template by ID
func (s *TemplateService) GetIncidentTemplateByID(id uint) (*models.IncidentTemplate, error) {
	var template models.IncidentTemplate
	if err := s.db.Preload("Services").First(&template, id).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

// UpdateIncidentTemplate updates an incident template
func (s *TemplateService) UpdateIncidentTemplate(template *models.IncidentTemplate, serviceIDs []uint) error {
	if err := s.db.Save(template).Error; err != nil {
		logger.Error("Failed to update incident template", zap.Error(err))
		return err
	}

	// Update service associations if provided
	if serviceIDs != nil {
		// Clear existing associations
		if err := s.db.Model(template).Association("Services").Clear(); err != nil {
			logger.Error("Failed to clear service associations", zap.Error(err))
			return err
		}

		// Add new associations
		if len(serviceIDs) > 0 {
			var services []models.Service
			if err := s.db.Where("id IN ?", serviceIDs).Find(&services).Error; err != nil {
				logger.Error("Failed to find services for template", zap.Error(err))
				return err
			}

			if err := s.db.Model(template).Association("Services").Append(services); err != nil {
				logger.Error("Failed to associate services with template", zap.Error(err))
				return err
			}
		}
	}

	logger.Info("Incident template updated successfully", zap.String("name", template.Name))
	return nil
}

// DeleteIncidentTemplate deletes an incident template
func (s *TemplateService) DeleteIncidentTemplate(id uint) error {
	if err := s.db.Delete(&models.IncidentTemplate{}, id).Error; err != nil {
		logger.Error("Failed to delete incident template", zap.Error(err))
		return err
	}

	logger.Info("Incident template deleted successfully", zap.Uint("id", id))
	return nil
}

// Maintenance Template Methods

// CreateMaintenanceTemplate creates a new maintenance template
func (s *TemplateService) CreateMaintenanceTemplate(template *models.MaintenanceTemplate, serviceIDs []uint) error {
	if err := s.db.Create(template).Error; err != nil {
		logger.Error("Failed to create maintenance template", zap.Error(err))
		return err
	}

	// Associate with services if provided
	if len(serviceIDs) > 0 {
		var services []models.Service
		if err := s.db.Where("id IN ?", serviceIDs).Find(&services).Error; err != nil {
			logger.Error("Failed to find services for template", zap.Error(err))
			return err
		}

		if err := s.db.Model(template).Association("Services").Append(services); err != nil {
			logger.Error("Failed to associate services with template", zap.Error(err))
			return err
		}
	}

	logger.Info("Maintenance template created successfully", zap.String("name", template.Name))
	return nil
}

// GetAllMaintenanceTemplates retrieves all maintenance templates
func (s *TemplateService) GetAllMaintenanceTemplates() ([]models.MaintenanceTemplate, error) {
	var templates []models.MaintenanceTemplate
	if err := s.db.Preload("Services").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

// GetMaintenanceTemplateByID retrieves a maintenance template by ID
func (s *TemplateService) GetMaintenanceTemplateByID(id uint) (*models.MaintenanceTemplate, error) {
	var template models.MaintenanceTemplate
	if err := s.db.Preload("Services").First(&template, id).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

// UpdateMaintenanceTemplate updates a maintenance template
func (s *TemplateService) UpdateMaintenanceTemplate(template *models.MaintenanceTemplate, serviceIDs []uint) error {
	if err := s.db.Save(template).Error; err != nil {
		logger.Error("Failed to update maintenance template", zap.Error(err))
		return err
	}

	// Update service associations if provided
	if serviceIDs != nil {
		// Clear existing associations
		if err := s.db.Model(template).Association("Services").Clear(); err != nil {
			logger.Error("Failed to clear service associations", zap.Error(err))
			return err
		}

		// Add new associations
		if len(serviceIDs) > 0 {
			var services []models.Service
			if err := s.db.Where("id IN ?", serviceIDs).Find(&services).Error; err != nil {
				logger.Error("Failed to find services for template", zap.Error(err))
				return err
			}

			if err := s.db.Model(template).Association("Services").Append(services); err != nil {
				logger.Error("Failed to associate services with template", zap.Error(err))
				return err
			}
		}
	}

	logger.Info("Maintenance template updated successfully", zap.String("name", template.Name))
	return nil
}

// DeleteMaintenanceTemplate deletes a maintenance template
func (s *TemplateService) DeleteMaintenanceTemplate(id uint) error {
	if err := s.db.Delete(&models.MaintenanceTemplate{}, id).Error; err != nil {
		logger.Error("Failed to delete maintenance template", zap.Error(err))
		return err
	}

	logger.Info("Maintenance template deleted successfully", zap.Uint("id", id))
	return nil
}
