// Package services provides Docker monitoring business logic.
package services

import (
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage-monitoring-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DockerMonitoringService handles Docker monitoring business logic.
type DockerMonitoringService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewDockerMonitoringService creates a new Docker monitoring service.
func NewDockerMonitoringService(db *gorm.DB, logger *zap.Logger) *DockerMonitoringService {
	return &DockerMonitoringService{
		db:     db,
		logger: logger,
	}
}

// CreateDockerContainer creates a new Docker container.
func (s *DockerMonitoringService) CreateDockerContainer(container *models.DockerContainer) error {
	if err := container.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values
	if container.Status == "" {
		container.Status = "unknown"
	}

	if err := s.db.Create(container).Error; err != nil {
		s.logger.Error("Failed to create Docker container", zap.Error(err))
		return fmt.Errorf("failed to create Docker container: %w", err)
	}

	s.logger.Info("Docker container created successfully", zap.Uint("container_id", container.ID))
	return nil
}

// GetDockerContainer retrieves a Docker container by ID.
func (s *DockerMonitoringService) GetDockerContainer(id uint) (*models.DockerContainer, error) {
	var container models.DockerContainer
	if err := s.db.Preload("HealthChecks").First(&container, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("Docker container not found")
		}
		s.logger.Error("Failed to get Docker container", zap.Error(err))
		return nil, fmt.Errorf("failed to get Docker container: %w", err)
	}

	return &container, nil
}

// GetDockerContainers retrieves a list of Docker containers with pagination.
func (s *DockerMonitoringService) GetDockerContainers(tenantID uint, limit, offset int) ([]*models.DockerContainer, int64, error) {
	var containers []*models.DockerContainer
	var total int64

	// Get total count
	if err := s.db.Model(&models.DockerContainer{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count Docker containers", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count Docker containers: %w", err)
	}

	// Get containers with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).
		Preload("HealthChecks").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&containers).Error; err != nil {
		s.logger.Error("Failed to get Docker containers", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get Docker containers: %w", err)
	}

	return containers, total, nil
}

// UpdateDockerContainer updates a Docker container.
func (s *DockerMonitoringService) UpdateDockerContainer(container *models.DockerContainer) error {
	if err := container.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Save(container).Error; err != nil {
		s.logger.Error("Failed to update Docker container", zap.Error(err))
		return fmt.Errorf("failed to update Docker container: %w", err)
	}

	s.logger.Info("Docker container updated successfully", zap.Uint("container_id", container.ID))
	return nil
}

// DeleteDockerContainer soft deletes a Docker container.
func (s *DockerMonitoringService) DeleteDockerContainer(id uint) error {
	if err := s.db.Delete(&models.DockerContainer{}, id).Error; err != nil {
		s.logger.Error("Failed to delete Docker container", zap.Error(err))
		return fmt.Errorf("failed to delete Docker container: %w", err)
	}

	s.logger.Info("Docker container deleted successfully", zap.Uint("container_id", id))
	return nil
}

// CreateDockerContainerHealthCheck creates a new Docker container health check.
func (s *DockerMonitoringService) CreateDockerContainerHealthCheck(healthCheck *models.DockerContainerHealthCheck) error {
	if err := healthCheck.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values
	if healthCheck.Status == "" {
		healthCheck.Status = "unknown"
	}

	if err := s.db.Create(healthCheck).Error; err != nil {
		s.logger.Error("Failed to create Docker container health check", zap.Error(err))
		return fmt.Errorf("failed to create Docker container health check: %w", err)
	}

	s.logger.Info("Docker container health check created successfully", zap.Uint("health_check_id", healthCheck.ID))
	return nil
}

// UpdateDockerContainerHealthCheck updates a Docker container health check.
func (s *DockerMonitoringService) UpdateDockerContainerHealthCheck(healthCheck *models.DockerContainerHealthCheck) error {
	if err := healthCheck.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Save(healthCheck).Error; err != nil {
		s.logger.Error("Failed to update Docker container health check", zap.Error(err))
		return fmt.Errorf("failed to update Docker container health check: %w", err)
	}

	s.logger.Info("Docker container health check updated successfully", zap.Uint("health_check_id", healthCheck.ID))
	return nil
}

// DeleteDockerContainerHealthCheck deletes a Docker container health check.
func (s *DockerMonitoringService) DeleteDockerContainerHealthCheck(id uint) error {
	if err := s.db.Delete(&models.DockerContainerHealthCheck{}, id).Error; err != nil {
		s.logger.Error("Failed to delete Docker container health check", zap.Error(err))
		return fmt.Errorf("failed to delete Docker container health check: %w", err)
	}

	s.logger.Info("Docker container health check deleted successfully", zap.Uint("health_check_id", id))
	return nil
}

// GetDockerContainerHealth returns the health status of a Docker container.
func (s *DockerMonitoringService) GetDockerContainerHealth(containerID uint) (map[string]interface{}, error) {
	var container models.DockerContainer
	if err := s.db.Preload("HealthChecks").First(&container, containerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("Docker container not found")
		}
		s.logger.Error("Failed to get Docker container", zap.Error(err))
		return nil, fmt.Errorf("failed to get Docker container: %w", err)
	}

	// Analyze health checks
	healthyChecks := 0
	unhealthyChecks := 0
	startingChecks := 0
	unknownChecks := 0

	for _, healthCheck := range container.HealthChecks {
		switch healthCheck.Status {
		case "healthy":
			healthyChecks++
		case "unhealthy":
			unhealthyChecks++
		case "starting":
			startingChecks++
		default:
			unknownChecks++
		}
	}

	// Determine overall health
	healthStatus := "unknown"
	if len(container.HealthChecks) > 0 {
		if unhealthyChecks > 0 {
			healthStatus = "unhealthy"
		} else if startingChecks > 0 {
			healthStatus = "starting"
		} else if healthyChecks > 0 {
			healthStatus = "healthy"
		}
	} else {
		// If no health checks, use container status
		if container.Status == "running" {
			healthStatus = "healthy"
		} else if container.Status == "stopped" {
			healthStatus = "unhealthy"
		}
	}

	health := map[string]interface{}{
		"container_id":     container.ID,
		"name":             container.Name,
		"image":            container.Image,
		"status":           container.Status,
		"state":            container.State,
		"health_status":    healthStatus,
		"healthy_checks":   healthyChecks,
		"unhealthy_checks": unhealthyChecks,
		"starting_checks":  startingChecks,
		"unknown_checks":   unknownChecks,
		"total_checks":     len(container.HealthChecks),
		"health_checks":    container.HealthChecks,
		"last_checked":     container.LastChecked,
		"timestamp":        time.Now().UTC(),
	}

	return health, nil
}

// GetDockerHostHealth returns the overall health of a Docker host.
func (s *DockerMonitoringService) GetDockerHostHealth(tenantID uint) (map[string]interface{}, error) {
	var containers []models.DockerContainer
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&containers).Error; err != nil {
		s.logger.Error("Failed to get Docker containers", zap.Error(err))
		return nil, fmt.Errorf("failed to get Docker containers: %w", err)
	}

	// Calculate host health metrics
	totalContainers := len(containers)
	runningContainers := 0
	stoppedContainers := 0
	healthyContainers := 0
	unhealthyContainers := 0

	imageCounts := make(map[string]int)
	statusCounts := make(map[string]int)

	for _, container := range containers {
		// Count by status
		statusCounts[container.Status]++
		if container.Status == "running" {
			runningContainers++
		} else if container.Status == "stopped" {
			stoppedContainers++
		}

		// Count by image
		imageCounts[container.Image]++

		// Determine health based on health checks
		hasHealthyCheck := false
		hasUnhealthyCheck := false
		for _, healthCheck := range container.HealthChecks {
			if healthCheck.Status == "healthy" {
				hasHealthyCheck = true
			} else if healthCheck.Status == "unhealthy" {
				hasUnhealthyCheck = true
			}
		}

		if hasHealthyCheck && !hasUnhealthyCheck {
			healthyContainers++
		} else if hasUnhealthyCheck {
			unhealthyContainers++
		}
	}

	// Calculate health percentage
	healthPercentage := float64(0)
	if totalContainers > 0 {
		healthPercentage = float64(healthyContainers) / float64(totalContainers) * 100
	}

	hostHealth := map[string]interface{}{
		"tenant_id":            tenantID,
		"total_containers":     totalContainers,
		"running_containers":   runningContainers,
		"stopped_containers":   stoppedContainers,
		"healthy_containers":   healthyContainers,
		"unhealthy_containers": unhealthyContainers,
		"health_percentage":    healthPercentage,
		"status_breakdown":     statusCounts,
		"image_breakdown":      imageCounts,
		"timestamp":            time.Now().UTC(),
	}

	return hostHealth, nil
}

// GetDockerContainerByName retrieves a Docker container by name.
func (s *DockerMonitoringService) GetDockerContainerByName(tenantID uint, name string) (*models.DockerContainer, error) {
	var container models.DockerContainer
	if err := s.db.Where("tenant_id = ? AND name = ?", tenantID, name).
		Preload("HealthChecks").
		First(&container).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("Docker container not found")
		}
		s.logger.Error("Failed to get Docker container by name", zap.Error(err))
		return nil, fmt.Errorf("failed to get Docker container by name: %w", err)
	}

	return &container, nil
}

// UpdateDockerContainerStatus updates the status of a Docker container.
func (s *DockerMonitoringService) UpdateDockerContainerStatus(containerID uint, status, state string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       status,
		"last_checked": now,
		"updated_at":   now,
	}

	if state != "" {
		updates["state"] = state
	}

	if err := s.db.Model(&models.DockerContainer{}).
		Where("id = ?", containerID).
		Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update Docker container status", zap.Error(err))
		return fmt.Errorf("failed to update Docker container status: %w", err)
	}

	s.logger.Info("Docker container status updated",
		zap.Uint("container_id", containerID),
		zap.String("status", status),
		zap.String("state", state))

	return nil
}
