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
