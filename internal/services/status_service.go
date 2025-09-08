package services

import (
	"fmt"
	"time"

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

// GetServicesByGroup retrieves all services for a specific group.
func (s *StatusService) GetServicesByGroup(tenantID uint, group string) ([]models.Service, error) {
	var services []models.Service
	query := database.DB.Where("tenant_id = ?", tenantID)
	if group != "" {
		query = query.Where("group = ?", group)
	}
	if err := query.Order("position ASC, name ASC").Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}

// GetServiceGroups retrieves all unique service groups for a tenant.
func (s *StatusService) GetServiceGroups(tenantID uint) ([]string, error) {
	var groups []string
	if err := database.DB.Model(&models.Service{}).
		Where("tenant_id = ? AND group != ''", tenantID).
		Distinct("group").
		Pluck("group", &groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

// UpdateServiceStatus updates the status of a service.
func (s *StatusService) UpdateServiceStatus(id uint, status string) error {
	return database.DB.Model(&models.Service{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// BulkUpdateServiceStatus updates the status of multiple services.
func (s *StatusService) BulkUpdateServiceStatus(serviceIDs []uint, status string) error {
	return database.DB.Model(&models.Service{}).
		Where("id IN ?", serviceIDs).
		Update("status", status).Error
}

// GetServiceStatusSummary returns a summary of service statuses for a tenant.
func (s *StatusService) GetServiceStatusSummary(tenantID uint) (map[string]int, error) {
	var results []struct {
		Status string
		Count  int
	}

	if err := database.DB.Model(&models.Service{}).
		Select("status, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("status").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	summary := make(map[string]int)
	for _, result := range results {
		summary[result.Status] = result.Count
	}

	return summary, nil
}

// GetOverallStatus determines the overall system status based on service statuses.
func (s *StatusService) GetOverallStatus(tenantID uint) (string, error) {
	summary, err := s.GetServiceStatusSummary(tenantID)
	if err != nil {
		return "", err
	}

	// Priority order: major_outage > partial_outage > degraded_performance > under_maintenance > operational
	if summary[models.StatusMajorOutage] > 0 {
		return models.StatusMajorOutage, nil
	}
	if summary[models.StatusPartialOutage] > 0 {
		return models.StatusPartialOutage, nil
	}
	if summary[models.StatusDegradedPerformance] > 0 {
		return models.StatusDegradedPerformance, nil
	}
	if summary[models.StatusUnderMaintenance] > 0 {
		return models.StatusUnderMaintenance, nil
	}

	return models.StatusOperational, nil
}

// ReorderServices updates the position of services.
func (s *StatusService) ReorderServices(serviceOrders map[uint]int) error {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for serviceID, position := range serviceOrders {
		if err := tx.Model(&models.Service{}).
			Where("id = ?", serviceID).
			Update("position", position).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetServiceUptime calculates uptime percentage for a service over a given period.
func (s *StatusService) GetServiceUptime(serviceID uint, days int) (float64, error) {
	var uptimeStats models.UptimeStats
	if err := database.DB.Where("service_id = ? AND period = ?", serviceID, days).
		Order("created_at DESC").
		First(&uptimeStats).Error; err != nil {
		// If no stats found, return 100% (assume operational)
		return 100.0, nil
	}

	return uptimeStats.Uptime, nil
}

// GetServiceHealthHistory retrieves health check history for a service.
func (s *StatusService) GetServiceHealthHistory(serviceID uint, hours int) ([]models.HealthCheck, error) {
	var healthChecks []models.HealthCheck
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	if err := database.DB.Where("service_id = ? AND checked_at >= ?", serviceID, since).
		Order("checked_at DESC").
		Find(&healthChecks).Error; err != nil {
		return nil, err
	}

	return healthChecks, nil
}

// CreateServiceGroup creates a new service group.
func (s *StatusService) CreateServiceGroup(tenantID uint, groupName string) error {
	// This is a simple implementation - in a real system, you might want a separate Group model
	// For now, we'll just ensure the group exists by checking if any services use it
	var count int64
	database.DB.Model(&models.Service{}).
		Where("tenant_id = ? AND group = ?", tenantID, groupName).
		Count(&count)

	if count == 0 {
		// Create a placeholder service to establish the group
		service := &models.Service{
			Name:     fmt.Sprintf("Group: %s", groupName),
			Group:    groupName,
			Status:   models.StatusOperational,
			TenantID: tenantID,
			Position: 0,
		}
		return database.DB.Create(service).Error
	}

	return nil
}

// DeleteServiceGroup deletes a service group and moves all services to ungrouped.
func (s *StatusService) DeleteServiceGroup(tenantID uint, groupName string) error {
	return database.DB.Model(&models.Service{}).
		Where("tenant_id = ? AND group = ?", tenantID, groupName).
		Update("group", "").Error
}

// GetServiceMetrics retrieves performance metrics for a service.
func (s *StatusService) GetServiceMetrics(serviceID uint, hours int) ([]models.SystemMetric, error) {
	var metrics []models.SystemMetric
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	if err := database.DB.Where("service_id = ? AND timestamp >= ?", serviceID, since).
		Order("timestamp DESC").
		Find(&metrics).Error; err != nil {
		return nil, err
	}

	return metrics, nil
}

// GetServicesWithUptime retrieves services with their uptime information.
func (s *StatusService) GetServicesWithUptime(tenantID uint) ([]ServiceWithUptime, error) {
	var services []models.Service
	if err := database.DB.Where("tenant_id = ?", tenantID).
		Order("position ASC, name ASC").
		Find(&services).Error; err != nil {
		return nil, err
	}

	var result []ServiceWithUptime
	for _, service := range services {
		uptime, _ := s.GetServiceUptime(service.ID, 30) // 30-day uptime
		result = append(result, ServiceWithUptime{
			Service: service,
			Uptime:  uptime,
		})
	}

	return result, nil
}

// ServiceWithUptime represents a service with its uptime information.
type ServiceWithUptime struct {
	models.Service
	Uptime float64 `json:"uptime"`
}
