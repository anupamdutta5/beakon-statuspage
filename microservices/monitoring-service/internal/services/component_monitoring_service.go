// Package services provides component and container monitoring business logic.
package services

import (
	"fmt"
	"time"

	"github.com/anupamdutta5/statuspage-monitoring-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ComponentMonitoringService handles component and container monitoring business logic.
type ComponentMonitoringService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewComponentMonitoringService creates a new component monitoring service.
func NewComponentMonitoringService(db *gorm.DB, logger *zap.Logger) *ComponentMonitoringService {
	return &ComponentMonitoringService{
		db:     db,
		logger: logger,
	}
}

// CreateComponent creates a new component.
func (s *ComponentMonitoringService) CreateComponent(component *models.MonitoredComponent) error {
	if err := component.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values
	if component.Status == "" {
		component.Status = "unknown"
	}

	if err := s.db.Create(component).Error; err != nil {
		s.logger.Error("Failed to create component", zap.Error(err))
		return fmt.Errorf("failed to create component: %w", err)
	}

	s.logger.Info("Component created successfully", zap.Uint("component_id", component.ID))
	return nil
}

// GetComponent retrieves a component by ID.
func (s *ComponentMonitoringService) GetComponent(id uint) (*models.MonitoredComponent, error) {
	var component models.MonitoredComponent
	if err := s.db.Preload("Containers").Preload("Containers.HealthChecks").Preload("Metrics").First(&component, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("component not found")
		}
		s.logger.Error("Failed to get component", zap.Error(err))
		return nil, fmt.Errorf("failed to get component: %w", err)
	}

	return &component, nil
}

// GetComponents retrieves a list of components with pagination.
func (s *ComponentMonitoringService) GetComponents(tenantID uint, limit, offset int) ([]*models.MonitoredComponent, int64, error) {
	var components []*models.MonitoredComponent
	var total int64

	// Get total count
	if err := s.db.Model(&models.MonitoredComponent{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count components", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count components: %w", err)
	}

	// Get components with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).
		Preload("Containers").
		Preload("Containers.HealthChecks").
		Preload("Metrics").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&components).Error; err != nil {
		s.logger.Error("Failed to get components", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get components: %w", err)
	}

	return components, total, nil
}

// UpdateComponent updates a component.
func (s *ComponentMonitoringService) UpdateComponent(component *models.MonitoredComponent) error {
	if err := component.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Save(component).Error; err != nil {
		s.logger.Error("Failed to update component", zap.Error(err))
		return fmt.Errorf("failed to update component: %w", err)
	}

	s.logger.Info("Component updated successfully", zap.Uint("component_id", component.ID))
	return nil
}

// DeleteComponent soft deletes a component.
func (s *ComponentMonitoringService) DeleteComponent(id uint) error {
	if err := s.db.Delete(&models.MonitoredComponent{}, id).Error; err != nil {
		s.logger.Error("Failed to delete component", zap.Error(err))
		return fmt.Errorf("failed to delete component: %w", err)
	}

	s.logger.Info("Component deleted successfully", zap.Uint("component_id", id))
	return nil
}

// CreateContainer creates a new container.
func (s *ComponentMonitoringService) CreateContainer(container *models.MonitoredContainer) error {
	if err := container.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values
	if container.Status == "" {
		container.Status = "unknown"
	}

	if err := s.db.Create(container).Error; err != nil {
		s.logger.Error("Failed to create container", zap.Error(err))
		return fmt.Errorf("failed to create container: %w", err)
	}

	s.logger.Info("Container created successfully", zap.Uint("container_id", container.ID))
	return nil
}

// GetContainer retrieves a container by ID.
func (s *ComponentMonitoringService) GetContainer(id uint) (*models.MonitoredContainer, error) {
	var container models.MonitoredContainer
	if err := s.db.Preload("Component").Preload("HealthChecks").First(&container, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("container not found")
		}
		s.logger.Error("Failed to get container", zap.Error(err))
		return nil, fmt.Errorf("failed to get container: %w", err)
	}

	return &container, nil
}

// UpdateContainer updates a container.
func (s *ComponentMonitoringService) UpdateContainer(container *models.MonitoredContainer) error {
	if err := container.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Save(container).Error; err != nil {
		s.logger.Error("Failed to update container", zap.Error(err))
		return fmt.Errorf("failed to update container: %w", err)
	}

	s.logger.Info("Container updated successfully", zap.Uint("container_id", container.ID))
	return nil
}

// DeleteContainer soft deletes a container.
func (s *ComponentMonitoringService) DeleteContainer(id uint) error {
	if err := s.db.Delete(&models.MonitoredContainer{}, id).Error; err != nil {
		s.logger.Error("Failed to delete container", zap.Error(err))
		return fmt.Errorf("failed to delete container: %w", err)
	}

	s.logger.Info("Container deleted successfully", zap.Uint("container_id", id))
	return nil
}

// CreateContainerHealthCheck creates a new container health check.
func (s *ComponentMonitoringService) CreateContainerHealthCheck(healthCheck *models.ContainerHealthCheck) error {
	if err := healthCheck.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values
	if healthCheck.Status == "" {
		healthCheck.Status = "unknown"
	}

	if err := s.db.Create(healthCheck).Error; err != nil {
		s.logger.Error("Failed to create container health check", zap.Error(err))
		return fmt.Errorf("failed to create container health check: %w", err)
	}

	s.logger.Info("Container health check created successfully", zap.Uint("health_check_id", healthCheck.ID))
	return nil
}

// UpdateContainerHealthCheck updates a container health check.
func (s *ComponentMonitoringService) UpdateContainerHealthCheck(healthCheck *models.ContainerHealthCheck) error {
	if err := healthCheck.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Save(healthCheck).Error; err != nil {
		s.logger.Error("Failed to update container health check", zap.Error(err))
		return fmt.Errorf("failed to update container health check: %w", err)
	}

	s.logger.Info("Container health check updated successfully", zap.Uint("health_check_id", healthCheck.ID))
	return nil
}

// CreateComponentMetric creates a new component metric.
func (s *ComponentMonitoringService) CreateComponentMetric(metric *models.ComponentMetric) error {
	if err := metric.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Create(metric).Error; err != nil {
		s.logger.Error("Failed to create component metric", zap.Error(err))
		return fmt.Errorf("failed to create component metric: %w", err)
	}

	s.logger.Info("Component metric created successfully", zap.Uint("metric_id", metric.ID))
	return nil
}

// GetComponentMetrics retrieves metrics for a component within a time range.
func (s *ComponentMonitoringService) GetComponentMetrics(componentID uint, startDate, endDate time.Time) ([]*models.ComponentMetric, error) {
	var metrics []*models.ComponentMetric

	query := s.db.Where("component_id = ?", componentID)
	if !startDate.IsZero() {
		query = query.Where("timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("timestamp <= ?", endDate)
	}

	if err := query.Order("timestamp DESC").Find(&metrics).Error; err != nil {
		s.logger.Error("Failed to get component metrics", zap.Error(err))
		return nil, fmt.Errorf("failed to get component metrics: %w", err)
	}

	return metrics, nil
}

// GetComponentHealth returns the health status of a component.
func (s *ComponentMonitoringService) GetComponentHealth(componentID uint) (map[string]interface{}, error) {
	var component models.MonitoredComponent
	if err := s.db.Preload("Containers.HealthChecks").Preload("Metrics").First(&component, componentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("component not found")
		}
		s.logger.Error("Failed to get component", zap.Error(err))
		return nil, fmt.Errorf("failed to get component: %w", err)
	}

	// Calculate health metrics
	healthyContainers := 0
	totalContainers := len(component.Containers)

	for _, container := range component.Containers {
		if container.Status == "healthy" {
			healthyContainers++
		}
	}

	healthPercentage := float64(0)
	if totalContainers > 0 {
		healthPercentage = float64(healthyContainers) / float64(totalContainers) * 100
	}

	health := map[string]interface{}{
		"component_id":       component.ID,
		"status":             component.Status,
		"healthy_containers": healthyContainers,
		"total_containers":   totalContainers,
		"health_percentage":  healthPercentage,
		"containers":         component.Containers,
		"timestamp":          time.Now().UTC(),
	}

	return health, nil
}
