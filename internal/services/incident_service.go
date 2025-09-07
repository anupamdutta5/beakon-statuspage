package services

import (
	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
)

// IncidentService provides methods for incident management
type IncidentService struct{}

// NewIncidentService creates a new IncidentService
func NewIncidentService() *IncidentService {
	return &IncidentService{}
}

// CreateIncident creates a new incident
func (s *IncidentService) CreateIncident(incident *models.Incident) (*models.Incident, error) {
	if err := database.DB.Create(incident).Error; err != nil {
		return nil, err
	}
	return incident, nil
}

// GetAllIncidents retrieves all incidents
func (s *IncidentService) GetAllIncidents() ([]models.Incident, error) {
	var incidents []models.Incident
	if err := database.DB.Preload("Services").Order("created_at DESC").Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}

// GetIncidentByID retrieves an incident by ID
func (s *IncidentService) GetIncidentByID(id uint) (*models.Incident, error) {
	var incident models.Incident
	if err := database.DB.Preload("Services").Preload("Updates").First(&incident, id).Error; err != nil {
		return nil, err
	}
	return &incident, nil
}

// GetIncidentsByTenantID retrieves all incidents for a specific tenant
func (s *IncidentService) GetIncidentsByTenantID(tenantID uint) ([]models.Incident, error) {
	var incidents []models.Incident
	if err := database.DB.Where("tenant_id = ?", tenantID).Preload("Services").Order("created_at DESC").Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}

// GetActiveIncidentsByTenantID retrieves active incidents for a specific tenant
func (s *IncidentService) GetActiveIncidentsByTenantID(tenantID uint) ([]models.Incident, error) {
	var incidents []models.Incident
	if err := database.DB.Where("tenant_id = ? AND status IN ?", tenantID, []string{"investigating", "identified", "monitoring"}).Preload("Services").Order("created_at DESC").Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}
