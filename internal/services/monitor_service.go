package services

import (
	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
)

type MonitorService struct{}

func NewMonitorService() *MonitorService {
	return &MonitorService{}
}

func (s *MonitorService) GetAllMonitors() ([]models.Monitor, error) {
	var monitors []models.Monitor
	err := database.DB.Find(&monitors).Error
	return monitors, err
}

func (s *MonitorService) GetMonitorByID(id uint) (*models.Monitor, error) {
	var monitor models.Monitor
	err := database.DB.First(&monitor, id).Error
	return &monitor, err
}

func (s *MonitorService) CreateMonitor(monitor *models.Monitor) error {
	return database.DB.Create(monitor).Error
}

func (s *MonitorService) UpdateMonitor(monitor *models.Monitor) error {
	return database.DB.Save(monitor).Error
}

func (s *MonitorService) DeleteMonitor(id uint) error {
	return database.DB.Delete(&models.Monitor{}, id).Error
}

// GetMonitorsByTenantID retrieves all monitors for a specific tenant
func (s *MonitorService) GetMonitorsByTenantID(tenantID uint) ([]models.Monitor, error) {
	var monitors []models.Monitor
	if err := database.DB.Where("tenant_id = ?", tenantID).Find(&monitors).Error; err != nil {
		return nil, err
	}
	return monitors, nil
}
