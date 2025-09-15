// Package services provides Kubernetes monitoring business logic.
package services

import (
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage-monitoring-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// KubernetesMonitoringService handles Kubernetes monitoring business logic.
type KubernetesMonitoringService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewKubernetesMonitoringService creates a new Kubernetes monitoring service.
func NewKubernetesMonitoringService(db *gorm.DB, logger *zap.Logger) *KubernetesMonitoringService {
	return &KubernetesMonitoringService{
		db:     db,
		logger: logger,
	}
}

// CreateKubernetesResource creates a new Kubernetes resource.
func (s *KubernetesMonitoringService) CreateKubernetesResource(resource *models.KubernetesResource) error {
	if err := resource.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values
	if resource.Status == "" {
		resource.Status = "unknown"
	}

	if err := s.db.Create(resource).Error; err != nil {
		s.logger.Error("Failed to create Kubernetes resource", zap.Error(err))
		return fmt.Errorf("failed to create Kubernetes resource: %w", err)
	}

	s.logger.Info("Kubernetes resource created successfully", zap.Uint("resource_id", resource.ID))
	return nil
}

// GetKubernetesResource retrieves a Kubernetes resource by ID.
func (s *KubernetesMonitoringService) GetKubernetesResource(id uint) (*models.KubernetesResource, error) {
	var resource models.KubernetesResource
	if err := s.db.Preload("Events").First(&resource, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("Kubernetes resource not found")
		}
		s.logger.Error("Failed to get Kubernetes resource", zap.Error(err))
		return nil, fmt.Errorf("failed to get Kubernetes resource: %w", err)
	}

	return &resource, nil
}

// GetKubernetesResources retrieves a list of Kubernetes resources with pagination.
func (s *KubernetesMonitoringService) GetKubernetesResources(tenantID uint, limit, offset int) ([]*models.KubernetesResource, int64, error) {
	var resources []*models.KubernetesResource
	var total int64

	// Get total count
	if err := s.db.Model(&models.KubernetesResource{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count Kubernetes resources", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count Kubernetes resources: %w", err)
	}

	// Get resources with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).
		Preload("Events").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&resources).Error; err != nil {
		s.logger.Error("Failed to get Kubernetes resources", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get Kubernetes resources: %w", err)
	}

	return resources, total, nil
}

// UpdateKubernetesResource updates a Kubernetes resource.
func (s *KubernetesMonitoringService) UpdateKubernetesResource(resource *models.KubernetesResource) error {
	if err := resource.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Save(resource).Error; err != nil {
		s.logger.Error("Failed to update Kubernetes resource", zap.Error(err))
		return fmt.Errorf("failed to update Kubernetes resource: %w", err)
	}

	s.logger.Info("Kubernetes resource updated successfully", zap.Uint("resource_id", resource.ID))
	return nil
}

// DeleteKubernetesResource soft deletes a Kubernetes resource.
func (s *KubernetesMonitoringService) DeleteKubernetesResource(id uint) error {
	if err := s.db.Delete(&models.KubernetesResource{}, id).Error; err != nil {
		s.logger.Error("Failed to delete Kubernetes resource", zap.Error(err))
		return fmt.Errorf("failed to delete Kubernetes resource: %w", err)
	}

	s.logger.Info("Kubernetes resource deleted successfully", zap.Uint("resource_id", id))
	return nil
}

// CreateKubernetesEvent creates a new Kubernetes event.
func (s *KubernetesMonitoringService) CreateKubernetesEvent(event *models.KubernetesEvent) error {
	if err := event.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Create(event).Error; err != nil {
		s.logger.Error("Failed to create Kubernetes event", zap.Error(err))
		return fmt.Errorf("failed to create Kubernetes event: %w", err)
	}

	s.logger.Info("Kubernetes event created successfully", zap.Uint("event_id", event.ID))
	return nil
}

// GetKubernetesEvents retrieves events for a Kubernetes resource.
func (s *KubernetesMonitoringService) GetKubernetesEvents(resourceID uint, limit, offset int) ([]*models.KubernetesEvent, int64, error) {
	var events []*models.KubernetesEvent
	var total int64

	// Get total count
	if err := s.db.Model(&models.KubernetesEvent{}).Where("resource_id = ?", resourceID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count Kubernetes events", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count Kubernetes events: %w", err)
	}

	// Get events with pagination
	if err := s.db.Where("resource_id = ?", resourceID).
		Preload("Resource").
		Limit(limit).
		Offset(offset).
		Order("last_seen DESC").
		Find(&events).Error; err != nil {
		s.logger.Error("Failed to get Kubernetes events", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get Kubernetes events: %w", err)
	}

	return events, total, nil
}

// GetKubernetesResourceHealth returns the health status of a Kubernetes resource.
func (s *KubernetesMonitoringService) GetKubernetesResourceHealth(resourceID uint) (map[string]interface{}, error) {
	var resource models.KubernetesResource
	if err := s.db.Preload("Events").First(&resource, resourceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("Kubernetes resource not found")
		}
		s.logger.Error("Failed to get Kubernetes resource", zap.Error(err))
		return nil, fmt.Errorf("failed to get Kubernetes resource: %w", err)
	}

	// Analyze events to determine health
	errorEvents := 0
	warningEvents := 0
	normalEvents := 0

	for _, event := range resource.Events {
		switch event.Type {
		case "Error":
			errorEvents++
		case "Warning":
			warningEvents++
		case "Normal":
			normalEvents++
		}
	}

	// Determine overall health based on events
	healthStatus := "healthy"
	if errorEvents > 0 {
		healthStatus = "unhealthy"
	} else if warningEvents > 0 {
		healthStatus = "degraded"
	}

	health := map[string]interface{}{
		"resource_id":    resource.ID,
		"name":           resource.Name,
		"namespace":      resource.Namespace,
		"type":           resource.Type,
		"kind":           resource.Kind,
		"status":         resource.Status,
		"health_status":  healthStatus,
		"error_events":   errorEvents,
		"warning_events": warningEvents,
		"normal_events":  normalEvents,
		"total_events":   len(resource.Events),
		"last_checked":   resource.LastChecked,
		"timestamp":      time.Now().UTC(),
	}

	return health, nil
}

// GetKubernetesClusterHealth returns the overall health of a Kubernetes cluster.
func (s *KubernetesMonitoringService) GetKubernetesClusterHealth(tenantID uint) (map[string]interface{}, error) {
	var resources []models.KubernetesResource
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&resources).Error; err != nil {
		s.logger.Error("Failed to get Kubernetes resources", zap.Error(err))
		return nil, fmt.Errorf("failed to get Kubernetes resources: %w", err)
	}

	// Calculate cluster health metrics
	totalResources := len(resources)
	healthyResources := 0
	unhealthyResources := 0
	degradedResources := 0
	unknownResources := 0

	resourceTypeCounts := make(map[string]int)
	namespaceCounts := make(map[string]int)

	for _, resource := range resources {
		// Count by status
		switch resource.Status {
		case "running":
			healthyResources++
		case "failed", "error":
			unhealthyResources++
		case "pending", "warning":
			degradedResources++
		default:
			unknownResources++
		}

		// Count by type
		resourceTypeCounts[resource.Type]++
		namespaceCounts[resource.Namespace]++
	}

	// Calculate health percentage
	healthPercentage := float64(0)
	if totalResources > 0 {
		healthPercentage = float64(healthyResources) / float64(totalResources) * 100
	}

	clusterHealth := map[string]interface{}{
		"tenant_id":           tenantID,
		"total_resources":     totalResources,
		"healthy_resources":   healthyResources,
		"unhealthy_resources": unhealthyResources,
		"degraded_resources":  degradedResources,
		"unknown_resources":   unknownResources,
		"health_percentage":   healthPercentage,
		"resource_types":      resourceTypeCounts,
		"namespaces":          namespaceCounts,
		"timestamp":           time.Now().UTC(),
	}

	return clusterHealth, nil
}

// GetKubernetesResourceByName retrieves a Kubernetes resource by name and namespace.
func (s *KubernetesMonitoringService) GetKubernetesResourceByName(tenantID uint, name, namespace string) (*models.KubernetesResource, error) {
	var resource models.KubernetesResource
	if err := s.db.Where("tenant_id = ? AND name = ? AND namespace = ?", tenantID, name, namespace).
		Preload("Events").
		First(&resource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("Kubernetes resource not found")
		}
		s.logger.Error("Failed to get Kubernetes resource by name", zap.Error(err))
		return nil, fmt.Errorf("failed to get Kubernetes resource by name: %w", err)
	}

	return &resource, nil
}

// UpdateKubernetesResourceStatus updates the status of a Kubernetes resource.
func (s *KubernetesMonitoringService) UpdateKubernetesResourceStatus(resourceID uint, status string) error {
	now := time.Now()
	if err := s.db.Model(&models.KubernetesResource{}).
		Where("id = ?", resourceID).
		Updates(map[string]interface{}{
			"status":       status,
			"last_checked": now,
			"updated_at":   now,
		}).Error; err != nil {
		s.logger.Error("Failed to update Kubernetes resource status", zap.Error(err))
		return fmt.Errorf("failed to update Kubernetes resource status: %w", err)
	}

	s.logger.Info("Kubernetes resource status updated",
		zap.Uint("resource_id", resourceID),
		zap.String("status", status))

	return nil
}

