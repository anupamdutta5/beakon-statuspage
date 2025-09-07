package services

import (
	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
)

// StatusService provides methods for interacting with status-related data.
type StatusService struct{}

// NewStatusService creates a new StatusService.
func NewStatusService() *StatusService {
	return &StatusService{}
}

// GetAllServices retrieves all services from the database.
func (s *StatusService) GetAllServices() ([]models.Service, error) {
	var services []models.Service
	if err := database.DB.Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}

// GetServiceByID retrieves a service by its ID.
func (s *StatusService) GetServiceByID(id uint) (*models.Service, error) {
	var service models.Service
	if err := database.DB.First(&service, id).Error; err != nil {
		return nil, err
	}
	return &service, nil
}

// CreateService creates a new service.
func (s *StatusService) CreateService(service *models.Service) (*models.Service, error) {
	if err := database.DB.Create(service).Error; err != nil {
		return nil, err
	}
	return service, nil
}

// UpdateService updates an existing service.
func (s *StatusService) UpdateService(service *models.Service) (*models.Service, error) {
	if err := database.DB.Save(service).Error; err != nil {
		return nil, err
	}
	return service, nil
}

// DeleteService deletes a service by its ID.
func (s *StatusService) DeleteService(id uint) error {
	return database.DB.Delete(&models.Service{}, id).Error
}

// GetServicesByTenantID retrieves all services for a specific tenant.
func (s *StatusService) GetServicesByTenantID(tenantID uint) ([]models.Service, error) {
	var services []models.Service
	if err := database.DB.Where("tenant_id = ?", tenantID).Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}
