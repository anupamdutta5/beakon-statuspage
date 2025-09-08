package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PrivatePageService struct {
	db *gorm.DB
}

func NewPrivatePageService() *PrivatePageService {
	return &PrivatePageService{
		db: database.GetDB(),
	}
}

// CreatePrivatePage creates a new private status page
func (s *PrivatePageService) CreatePrivatePage(name, description string, serviceIDs []uint) (*models.PrivatePage, error) {
	// Generate a unique access key
	accessKey, err := s.generateAccessKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate access key: %w", err)
	}

	privatePage := &models.PrivatePage{
		Name:        name,
		Description: description,
		AccessKey:   accessKey,
		IsActive:    true,
	}

	if err := s.db.Create(privatePage).Error; err != nil {
		logger.Error("Failed to create private page", zap.Error(err))
		return nil, err
	}

	// Associate with services if provided
	if len(serviceIDs) > 0 {
		var services []models.Service
		if err := s.db.Where("id IN ?", serviceIDs).Find(&services).Error; err != nil {
			logger.Error("Failed to find services for private page", zap.Error(err))
			return nil, err
		}

		if err := s.db.Model(privatePage).Association("Services").Append(services); err != nil {
			logger.Error("Failed to associate services with private page", zap.Error(err))
			return nil, err
		}
	}

	logger.Info("Private page created successfully",
		zap.String("name", name),
		zap.String("access_key", accessKey))

	return privatePage, nil
}

// GetPrivatePageByAccessKey retrieves a private page by its access key
func (s *PrivatePageService) GetPrivatePageByAccessKey(accessKey string) (*models.PrivatePage, error) {
	var privatePage models.PrivatePage
	if err := s.db.Preload("Services").Where("access_key = ? AND is_active = ?", accessKey, true).First(&privatePage).Error; err != nil {
		return nil, err
	}
	return &privatePage, nil
}

// GetAllPrivatePages retrieves all private pages
func (s *PrivatePageService) GetAllPrivatePages() ([]models.PrivatePage, error) {
	var pages []models.PrivatePage
	if err := s.db.Preload("Services").Find(&pages).Error; err != nil {
		return nil, err
	}
	return pages, nil
}

// GetPrivatePageByID retrieves a private page by ID
func (s *PrivatePageService) GetPrivatePageByID(id uint) (*models.PrivatePage, error) {
	var page models.PrivatePage
	if err := s.db.Preload("Services").First(&page, id).Error; err != nil {
		return nil, err
	}
	return &page, nil
}

// UpdatePrivatePage updates a private page
func (s *PrivatePageService) UpdatePrivatePage(page *models.PrivatePage, serviceIDs []uint) error {
	if err := s.db.Save(page).Error; err != nil {
		logger.Error("Failed to update private page", zap.Error(err))
		return err
	}

	// Update service associations if provided
	if serviceIDs != nil {
		// Clear existing associations
		if err := s.db.Model(page).Association("Services").Clear(); err != nil {
			logger.Error("Failed to clear service associations", zap.Error(err))
			return err
		}

		// Add new associations
		if len(serviceIDs) > 0 {
			var services []models.Service
			if err := s.db.Where("id IN ?", serviceIDs).Find(&services).Error; err != nil {
				logger.Error("Failed to find services for private page", zap.Error(err))
				return err
			}

			if err := s.db.Model(page).Association("Services").Append(services); err != nil {
				logger.Error("Failed to associate services with private page", zap.Error(err))
				return err
			}
		}
	}

	logger.Info("Private page updated successfully", zap.String("name", page.Name))
	return nil
}

// DeletePrivatePage deletes a private page
func (s *PrivatePageService) DeletePrivatePage(id uint) error {
	if err := s.db.Delete(&models.PrivatePage{}, id).Error; err != nil {
		logger.Error("Failed to delete private page", zap.Error(err))
		return err
	}

	logger.Info("Private page deleted successfully", zap.Uint("id", id))
	return nil
}

// RegenerateAccessKey generates a new access key for a private page
func (s *PrivatePageService) RegenerateAccessKey(id uint) (string, error) {
	accessKey, err := s.generateAccessKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate access key: %w", err)
	}

	if err := s.db.Model(&models.PrivatePage{}).Where("id = ?", id).Update("access_key", accessKey).Error; err != nil {
		logger.Error("Failed to regenerate access key", zap.Error(err))
		return "", err
	}

	logger.Info("Access key regenerated successfully", zap.Uint("id", id))
	return accessKey, nil
}

// TogglePrivatePageStatus toggles the active status of a private page
func (s *PrivatePageService) TogglePrivatePageStatus(id uint) error {
	var page models.PrivatePage
	if err := s.db.First(&page, id).Error; err != nil {
		return err
	}

	newStatus := !page.IsActive
	if err := s.db.Model(&page).Update("is_active", newStatus).Error; err != nil {
		logger.Error("Failed to toggle private page status", zap.Error(err))
		return err
	}

	logger.Info("Private page status toggled",
		zap.Uint("id", id),
		zap.Bool("new_status", newStatus))

	return nil
}

// GetPrivatePageData returns the data needed to display a private page
func (s *PrivatePageService) GetPrivatePageData(accessKey string) (map[string]interface{}, error) {
	page, err := s.GetPrivatePageByAccessKey(accessKey)
	if err != nil {
		return nil, err
	}

	// Get incidents for the associated services
	var serviceIDs []uint
	for _, service := range page.Services {
		serviceIDs = append(serviceIDs, service.ID)
	}

	var incidents []models.Incident
	if len(serviceIDs) > 0 {
		if err := s.db.Preload("Services").Where("id IN (SELECT incident_id FROM incident_services WHERE service_id IN ?)", serviceIDs).Find(&incidents).Error; err != nil {
			logger.Error("Failed to get incidents for private page", zap.Error(err))
			return nil, err
		}
	}

	// Get maintenance events for the associated services
	var maintenance []models.Maintenance
	if len(serviceIDs) > 0 {
		if err := s.db.Preload("Services").Where("id IN (SELECT maintenance_id FROM maintenance_services WHERE service_id IN ?)", serviceIDs).Find(&maintenance).Error; err != nil {
			logger.Error("Failed to get maintenance for private page", zap.Error(err))
			return nil, err
		}
	}

	// Determine overall status
	overallStatus := "operational"
	for _, service := range page.Services {
		if service.Status != "operational" {
			overallStatus = "degraded"
			break
		}
	}

	return map[string]interface{}{
		"page":           page,
		"services":       page.Services,
		"incidents":      incidents,
		"maintenance":    maintenance,
		"overall_status": overallStatus,
	}, nil
}

// generateAccessKey generates a secure random access key
func (s *PrivatePageService) generateAccessKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateAccessKey validates if an access key exists and is active
func (s *PrivatePageService) ValidateAccessKey(accessKey string) bool {
	var count int64
	s.db.Model(&models.PrivatePage{}).Where("access_key = ? AND is_active = ?", accessKey, true).Count(&count)
	return count > 0
}
