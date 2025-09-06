package services

import (
	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
)

// MaintenanceService provides methods for managing maintenance events.
type MaintenanceService struct{}

// NewMaintenanceService creates a new MaintenanceService.
func NewMaintenanceService() *MaintenanceService {
	return &MaintenanceService{}
}

// GetAllMaintenanceEvents returns all maintenance events.
func (s *MaintenanceService) GetAllMaintenanceEvents() ([]models.Maintenance, error) {
	var events []models.Maintenance
	if err := database.DB.Preload("Services").Order("start_at desc").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// GetMaintenanceEventByID retrieves a single maintenance event by its ID.
func (s *MaintenanceService) GetMaintenanceEventByID(id uint) (*models.Maintenance, error) {
	var event models.Maintenance
	if err := database.DB.Preload("Services").First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

// CreateMaintenanceEvent creates a new maintenance event and associates it with services.
func (s *MaintenanceService) CreateMaintenanceEvent(event *models.Maintenance, serviceIDs []uint) error {
	tx := database.DB.Begin()

	if err := tx.Create(event).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(serviceIDs) > 0 {
		var services []*models.Service
		if err := tx.Where("id IN ?", serviceIDs).Find(&services).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Model(event).Association("Services").Append(services); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// UpdateMaintenanceEvent updates an existing maintenance event.
func (s *MaintenanceService) UpdateMaintenanceEvent(event *models.Maintenance) error {
	return database.DB.Save(event).Error
}

// DeleteMaintenanceEvent deletes a maintenance event.
func (s *MaintenanceService) DeleteMaintenanceEvent(id uint) error {
	tx := database.DB.Begin()

	var event models.Maintenance
	if err := tx.First(&event, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&event).Association("Services").Clear(); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&event).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetUpcomingMaintenance retrieves all scheduled maintenance events
func (s *MaintenanceService) GetUpcomingMaintenance() ([]models.Maintenance, error) {
	var events []models.Maintenance
	if err := database.DB.Preload("Services").Where("status = ?", models.MaintenanceStatusScheduled).Order("start_at asc").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// GetActiveMaintenance retrieves all maintenance events currently in progress
func (s *MaintenanceService) GetActiveMaintenance() ([]models.Maintenance, error) {
	var events []models.Maintenance
	if err := database.DB.Preload("Services").Where("status = ?", models.MaintenanceStatusInProgress).Order("start_at desc").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}
