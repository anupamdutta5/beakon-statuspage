// Package services provides external service monitoring business logic.
package services

import (
	"fmt"
	"time"

	"github.com/anupamdutta5/statuspage-monitoring-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ExternalMonitoringService handles external service monitoring business logic.
type ExternalMonitoringService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewExternalMonitoringService creates a new external monitoring service.
func NewExternalMonitoringService(db *gorm.DB, logger *zap.Logger) *ExternalMonitoringService {
	return &ExternalMonitoringService{
		db:     db,
		logger: logger,
	}
}

// CreateExternalService creates a new external service.
func (s *ExternalMonitoringService) CreateExternalService(service *models.ExternalService) error {
	if err := service.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values
	if service.Status == "" {
		service.Status = "unknown"
	}
	if service.CheckInterval == 0 {
		service.CheckInterval = 300 // 5 minutes default
	}

	if err := s.db.Create(service).Error; err != nil {
		s.logger.Error("Failed to create external service", zap.Error(err))
		return fmt.Errorf("failed to create external service: %w", err)
	}

	s.logger.Info("External service created successfully", zap.Uint("service_id", service.ID))
	return nil
}

// GetExternalService retrieves an external service by ID.
func (s *ExternalMonitoringService) GetExternalService(id uint) (*models.ExternalService, error) {
	var service models.ExternalService
	if err := s.db.Preload("HealthChecks").First(&service, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("external service not found")
		}
		s.logger.Error("Failed to get external service", zap.Error(err))
		return nil, fmt.Errorf("failed to get external service: %w", err)
	}

	return &service, nil
}

// GetExternalServices retrieves a list of external services with pagination.
func (s *ExternalMonitoringService) GetExternalServices(tenantID uint, limit, offset int) ([]*models.ExternalService, int64, error) {
	var services []*models.ExternalService
	var total int64

	// Get total count
	if err := s.db.Model(&models.ExternalService{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count external services", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count external services: %w", err)
	}

	// Get services with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).
		Preload("HealthChecks").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&services).Error; err != nil {
		s.logger.Error("Failed to get external services", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get external services: %w", err)
	}

	return services, total, nil
}

// UpdateExternalService updates an external service.
func (s *ExternalMonitoringService) UpdateExternalService(service *models.ExternalService) error {
	if err := service.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Save(service).Error; err != nil {
		s.logger.Error("Failed to update external service", zap.Error(err))
		return fmt.Errorf("failed to update external service: %w", err)
	}

	s.logger.Info("External service updated successfully", zap.Uint("service_id", service.ID))
	return nil
}

// DeleteExternalService soft deletes an external service.
func (s *ExternalMonitoringService) DeleteExternalService(id uint) error {
	if err := s.db.Delete(&models.ExternalService{}, id).Error; err != nil {
		s.logger.Error("Failed to delete external service", zap.Error(err))
		return fmt.Errorf("failed to delete external service: %w", err)
	}

	s.logger.Info("External service deleted successfully", zap.Uint("service_id", id))
	return nil
}

// CreateExternalServiceHealthCheck creates a new external service health check.
func (s *ExternalMonitoringService) CreateExternalServiceHealthCheck(healthCheck *models.ExternalServiceHealthCheck) error {
	if err := healthCheck.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values
	if healthCheck.Method == "" {
		healthCheck.Method = "GET"
	}
	if healthCheck.ExpectedStatus == 0 {
		healthCheck.ExpectedStatus = 200
	}
	if healthCheck.Timeout == 0 {
		healthCheck.Timeout = 30
	}
	if healthCheck.LastResult == "" {
		healthCheck.LastResult = "unknown"
	}

	if err := s.db.Create(healthCheck).Error; err != nil {
		s.logger.Error("Failed to create external service health check", zap.Error(err))
		return fmt.Errorf("failed to create external service health check: %w", err)
	}

	s.logger.Info("External service health check created successfully", zap.Uint("health_check_id", healthCheck.ID))
	return nil
}

// UpdateExternalServiceHealthCheck updates an external service health check.
func (s *ExternalMonitoringService) UpdateExternalServiceHealthCheck(healthCheck *models.ExternalServiceHealthCheck) error {
	if err := healthCheck.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.Save(healthCheck).Error; err != nil {
		s.logger.Error("Failed to update external service health check", zap.Error(err))
		return fmt.Errorf("failed to update external service health check: %w", err)
	}

	s.logger.Info("External service health check updated successfully", zap.Uint("health_check_id", healthCheck.ID))
	return nil
}

// DeleteExternalServiceHealthCheck deletes an external service health check.
func (s *ExternalMonitoringService) DeleteExternalServiceHealthCheck(id uint) error {
	if err := s.db.Delete(&models.ExternalServiceHealthCheck{}, id).Error; err != nil {
		s.logger.Error("Failed to delete external service health check", zap.Error(err))
		return fmt.Errorf("failed to delete external service health check: %w", err)
	}

	s.logger.Info("External service health check deleted successfully", zap.Uint("health_check_id", id))
	return nil
}

// GetExternalServiceHealth returns the health status of an external service.
func (s *ExternalMonitoringService) GetExternalServiceHealth(serviceID uint) (map[string]interface{}, error) {
	var service models.ExternalService
	if err := s.db.Preload("HealthChecks").First(&service, serviceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("external service not found")
		}
		s.logger.Error("Failed to get external service", zap.Error(err))
		return nil, fmt.Errorf("failed to get external service: %w", err)
	}

	// Calculate health metrics
	health := map[string]interface{}{
		"service_id":     service.ID,
		"name":           service.Name,
		"type":           service.Type,
		"provider":       service.Provider,
		"status":         service.Status,
		"uptime":         service.Uptime,
		"response_time":  service.ResponseTime,
		"last_checked":   service.LastChecked,
		"last_healthy":   service.LastHealthy,
		"last_unhealthy": service.LastUnhealthy,
		"health_checks":  service.HealthChecks,
		"timestamp":      time.Now().UTC(),
	}

	return health, nil
}

// GetExternalServiceMetrics returns metrics for an external service.
func (s *ExternalMonitoringService) GetExternalServiceMetrics(serviceID uint, startDate, endDate time.Time) (map[string]interface{}, error) {
	var metrics map[string]interface{} = make(map[string]interface{})

	// Get service info
	var service models.ExternalService
	if err := s.db.First(&service, serviceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("external service not found")
		}
		s.logger.Error("Failed to get external service", zap.Error(err))
		return nil, fmt.Errorf("failed to get external service: %w", err)
	}

	// Calculate basic metrics
	metrics["service_id"] = service.ID
	metrics["service_name"] = service.Name
	metrics["uptime"] = service.Uptime
	metrics["response_time"] = service.ResponseTime
	metrics["status"] = service.Status
	metrics["last_checked"] = service.LastChecked
	metrics["date_range"] = map[string]interface{}{
		"start": startDate,
		"end":   endDate,
	}

	return metrics, nil
}

// UpdateExternalServiceStatus updates the status of an external service.
func (s *ExternalMonitoringService) UpdateExternalServiceStatus(serviceID uint, status string, responseTime float64) error {
	var service models.ExternalService
	if err := s.db.First(&service, serviceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("external service not found")
		}
		s.logger.Error("Failed to get external service", zap.Error(err))
		return fmt.Errorf("failed to get external service: %w", err)
	}

	now := time.Now()
	service.Status = status
	service.ResponseTime = responseTime
	service.LastChecked = &now

	if status == "operational" {
		service.LastHealthy = &now
	} else {
		service.LastUnhealthy = &now
	}

	if err := s.db.Save(&service).Error; err != nil {
		s.logger.Error("Failed to update external service status", zap.Error(err))
		return fmt.Errorf("failed to update external service status: %w", err)
	}

	s.logger.Info("External service status updated",
		zap.Uint("service_id", serviceID),
		zap.String("status", status),
		zap.Float64("response_time", responseTime))

	return nil
}

