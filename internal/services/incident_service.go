package services

import (
	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
)

// IncidentService provides methods for interacting with incidents.
type IncidentService struct{}

// NewIncidentService creates a new IncidentService.
func NewIncidentService() *IncidentService {
	return &IncidentService{}
}

// GetAllIncidents retrieves all incidents from the database.
func (s *IncidentService) GetAllIncidents() ([]models.Incident, error) {
	var incidents []models.Incident
	if err := database.DB.Preload("Updates").Preload("Services").Order("created_at desc").Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}

// CreateIncident creates a new incident and associates it with services.
func (s *IncidentService) CreateIncident(incident *models.Incident, serviceIDs []uint) error {
	// Start a transaction
	tx := database.DB.Begin()

	// Create the incident
	if err := tx.Create(incident).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Associate services with the incident
	if len(serviceIDs) > 0 {
		var services []*models.Service
		if err := tx.Where("id IN ?", serviceIDs).Find(&services).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Model(incident).Association("Services").Append(services); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetIncidentByID retrieves an incident by its ID
func (s *IncidentService) GetIncidentByID(id uint) (*models.Incident, error) {
	var incident models.Incident
	if err := database.DB.Preload("Updates").Preload("Services").First(&incident, id).Error; err != nil {
		return nil, err
	}
	return &incident, nil
}

// GetIncidentUpdates retrieves all updates for a specific incident
func (s *IncidentService) GetIncidentUpdates(incidentID uint) ([]models.StatusUpdate, error) {
	var updates []models.StatusUpdate
	if err := database.DB.Where("incident_id = ?", incidentID).Order("created_at desc").Find(&updates).Error; err != nil {
		return nil, err
	}
	return updates, nil
}

// GetActiveIncidents retrieves all non-resolved incidents
func (s *IncidentService) GetActiveIncidents() ([]models.Incident, error) {
	var incidents []models.Incident
	if err := database.DB.Preload("Updates").Preload("Services").Where("status != ?", models.IncidentStatusResolved).Order("created_at desc").Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}

// GetResolvedIncidents retrieves all resolved incidents
func (s *IncidentService) GetResolvedIncidents() ([]models.Incident, error) {
	var incidents []models.Incident
	if err := database.DB.Preload("Updates").Preload("Services").Where("status = ?", models.IncidentStatusResolved).Order("created_at desc").Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}
