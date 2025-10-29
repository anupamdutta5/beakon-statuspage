// Package services provides business logic for the Monitoring Service.
package http

import (
	"fmt"
	"time"

	"github.com/anupamdutta5/monitoring-service/internal/core/config"
	"github.com/anupamdutta5/monitoring-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MonitoringService handles monitoring-related business logic.
type MonitoringService struct {
	db     *gorm.DB
	logger *zap.Logger

	// Specialized monitoring services
	ComponentMonitoring   *ComponentMonitoringService
	ExternalMonitoring    *ExternalMonitoringService
	CustomMetrics         *CustomMetricsService
	StatusAutomation      *StatusAutomationService
	KubernetesMonitoring  *KubernetesMonitoringService
	DockerMonitoring      *DockerMonitoringService
	HealthCheck           *HealthCheckService
	MaintenanceManagement *MaintenanceManagementService
}

// NewMonitoringService creates a new monitoring service.
func NewMonitoringService(db *gorm.DB, logger *zap.Logger) *MonitoringService {
	// NOTE: Database migrations managed by Atlas (see ../../../../migrations/ and atlas.hcl)
	// AutoMigrate is NOT used - all schema changes via version-controlled migrations

	return &MonitoringService{
		db:                    db,
		logger:                logger,
		ComponentMonitoring:   NewComponentMonitoringService(db, logger),
		ExternalMonitoring:    NewExternalMonitoringService(db, logger),
		CustomMetrics:         NewCustomMetricsService(db, logger),
		StatusAutomation:      NewStatusAutomationService(db, logger),
		KubernetesMonitoring:  NewKubernetesMonitoringService(db, logger),
		DockerMonitoring:      NewDockerMonitoringService(db, logger),
		HealthCheck:           NewHealthCheckService(logger),
		MaintenanceManagement: NewMaintenanceManagementService(db, logger),
	}
}

// CreateService creates a new monitored service.
func (s *MonitoringService) CreateService(service *models.MonitoredService) error {
	// Set default values
	if service.Type == "" {
		service.Type = "http"
	}
	if service.Status == "" {
		service.Status = "unknown"
	}
	if service.CheckInterval == 0 {
		service.CheckInterval = 60
	}
	if service.Timeout == 0 {
		service.Timeout = 30
	}
	if service.Retries == 0 {
		service.Retries = 3
	}
	if service.IsActive == false && service.IsActive != true {
		service.IsActive = true
	}

	// Create service
	if err := s.db.Create(service).Error; err != nil {
		s.logger.Error("Failed to create monitored service", zap.Error(err))
		return fmt.Errorf("failed to create monitored service: %w", err)
	}

	s.logger.Info("Monitored service created successfully", zap.Uint("service_id", service.ID))
	return nil
}

// GetService retrieves a monitored service by ID.
func (s *MonitoringService) GetService(id uint) (*models.MonitoredService, error) {
	var service models.MonitoredService
	if err := s.db.Preload("HealthChecks").Preload("Alerts").First(&service, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("monitored service not found")
		}
		s.logger.Error("Failed to get monitored service", zap.Error(err))
		return nil, fmt.Errorf("failed to get monitored service: %w", err)
	}

	return &service, nil
}

// GetServices retrieves a list of monitored services with pagination.
func (s *MonitoringService) GetServices(tenantID uint, limit, offset int) ([]*models.MonitoredService, int64, error) {
	var services []*models.MonitoredService
	var total int64

	// Get total count
	if err := s.db.Model(&models.MonitoredService{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count monitored services", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count monitored services: %w", err)
	}

	// Get services with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).Preload("HealthChecks").Preload("Alerts").Limit(limit).Offset(offset).Order("created_at DESC").Find(&services).Error; err != nil {
		s.logger.Error("Failed to get monitored services", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get monitored services: %w", err)
	}

	return services, total, nil
}

// UpdateService updates a monitored service.
func (s *MonitoringService) UpdateService(service *models.MonitoredService) error {
	if err := s.db.Save(service).Error; err != nil {
		s.logger.Error("Failed to update monitored service", zap.Error(err))
		return fmt.Errorf("failed to update monitored service: %w", err)
	}

	s.logger.Info("Monitored service updated successfully", zap.Uint("service_id", service.ID))
	return nil
}

// DeleteService soft deletes a monitored service.
func (s *MonitoringService) DeleteService(id uint) error {
	if err := s.db.Delete(&models.MonitoredService{}, id).Error; err != nil {
		s.logger.Error("Failed to delete monitored service", zap.Error(err))
		return fmt.Errorf("failed to delete monitored service: %w", err)
	}

	s.logger.Info("Monitored service deleted successfully", zap.Uint("service_id", id))
	return nil
}

// GetServiceHealth returns the health status of a service.
func (s *MonitoringService) GetServiceHealth(serviceID uint) (map[string]interface{}, error) {
	var service models.MonitoredService
	if err := s.db.First(&service, serviceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("service not found")
		}
		s.logger.Error("Failed to get service", zap.Error(err))
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	// Get recent health check results
	var recentResults []models.HealthCheckResult
	if err := s.db.Where("service_id = ?", serviceID).Order("created_at DESC").Limit(10).Find(&recentResults).Error; err != nil {
		s.logger.Error("Failed to get recent health check results", zap.Error(err))
		return nil, fmt.Errorf("failed to get recent health check results: %w", err)
	}

	// Calculate health metrics
	health := map[string]interface{}{
		"service_id":     service.ID,
		"status":         service.Status,
		"uptime":         service.Uptime,
		"response_time":  service.ResponseTime,
		"last_checked":   service.LastChecked,
		"last_healthy":   service.LastHealthy,
		"last_unhealthy": service.LastUnhealthy,
		"recent_results": recentResults,
		"timestamp":      time.Now().UTC(),
	}

	return health, nil
}

// GetServiceMetrics returns metrics for a service.
func (s *MonitoringService) GetServiceMetrics(serviceID uint, startDate, endDate time.Time) (map[string]interface{}, error) {
	var metrics map[string]interface{} = make(map[string]interface{})

	// Get health check results in date range
	var results []models.HealthCheckResult
	query := s.db.Where("service_id = ?", serviceID)
	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}

	if err := query.Find(&results).Error; err != nil {
		s.logger.Error("Failed to get health check results", zap.Error(err))
		return nil, fmt.Errorf("failed to get health check results: %w", err)
	}

	// Calculate metrics
	totalChecks := len(results)
	successfulChecks := 0
	var totalResponseTime float64
	var maxResponseTime float64
	var minResponseTime float64 = 999999

	for _, result := range results {
		if result.Status == "success" {
			successfulChecks++
		}
		if result.ResponseTime > 0 {
			totalResponseTime += result.ResponseTime
			if result.ResponseTime > maxResponseTime {
				maxResponseTime = result.ResponseTime
			}
			if result.ResponseTime < minResponseTime {
				minResponseTime = result.ResponseTime
			}
		}
	}

	uptime := float64(0)
	if totalChecks > 0 {
		uptime = float64(successfulChecks) / float64(totalChecks) * 100
	}

	avgResponseTime := float64(0)
	if successfulChecks > 0 {
		avgResponseTime = totalResponseTime / float64(successfulChecks)
	}

	metrics["total_checks"] = totalChecks
	metrics["successful_checks"] = successfulChecks
	metrics["failed_checks"] = totalChecks - successfulChecks
	metrics["uptime_percentage"] = uptime
	metrics["average_response_time"] = avgResponseTime
	metrics["max_response_time"] = maxResponseTime
	metrics["min_response_time"] = minResponseTime
	metrics["date_range"] = map[string]interface{}{
		"start": startDate,
		"end":   endDate,
	}

	return metrics, nil
}

// GetMonitoringOverview returns an overview of monitoring data.
func (s *MonitoringService) GetMonitoringOverview(tenantID uint) (map[string]interface{}, error) {
	var overview map[string]interface{} = make(map[string]interface{})

	// Get total services count
	var totalServices int64
	if err := s.db.Model(&models.MonitoredService{}).Where("tenant_id = ?", tenantID).Count(&totalServices).Error; err != nil {
		s.logger.Error("Failed to count services", zap.Error(err))
		return nil, fmt.Errorf("failed to count services: %w", err)
	}

	// Get healthy services count
	var healthyServices int64
	if err := s.db.Model(&models.MonitoredService{}).Where("tenant_id = ? AND status = ?", tenantID, "healthy").Count(&healthyServices).Error; err != nil {
		s.logger.Error("Failed to count healthy services", zap.Error(err))
		return nil, fmt.Errorf("failed to count healthy services: %w", err)
	}

	// Get unhealthy services count
	var unhealthyServices int64
	if err := s.db.Model(&models.MonitoredService{}).Where("tenant_id = ? AND status = ?", tenantID, "unhealthy").Count(&unhealthyServices).Error; err != nil {
		s.logger.Error("Failed to count unhealthy services", zap.Error(err))
		return nil, fmt.Errorf("failed to count unhealthy services: %w", err)
	}

	// Get active alerts count
	var activeAlerts int64
	if err := s.db.Model(&models.Alert{}).Where("tenant_id = ? AND status = ?", tenantID, "active").Count(&activeAlerts).Error; err != nil {
		s.logger.Error("Failed to count active alerts", zap.Error(err))
		return nil, fmt.Errorf("failed to count active alerts: %w", err)
	}

	// Get total uptime checks count
	var totalUptimeChecks int64
	if err := s.db.Model(&models.UptimeCheck{}).Where("tenant_id = ?", tenantID).Count(&totalUptimeChecks).Error; err != nil {
		s.logger.Error("Failed to count uptime checks", zap.Error(err))
		return nil, fmt.Errorf("failed to count uptime checks: %w", err)
	}

	// Get recent health check results (last 24 hours)
	var recentResults int64
	last24Hours := time.Now().Add(-24 * time.Hour)
	if err := s.db.Model(&models.HealthCheckResult{}).Joins("JOIN monitored_services ON health_check_results.service_id = monitored_services.id").Where("monitored_services.tenant_id = ? AND health_check_results.created_at >= ?", tenantID, last24Hours).Count(&recentResults).Error; err != nil {
		s.logger.Error("Failed to count recent health check results", zap.Error(err))
		return nil, fmt.Errorf("failed to count recent health check results: %w", err)
	}

	overview["total_services"] = totalServices
	overview["healthy_services"] = healthyServices
	overview["unhealthy_services"] = unhealthyServices
	overview["active_alerts"] = activeAlerts
	overview["total_uptime_checks"] = totalUptimeChecks
	overview["recent_health_checks_24h"] = recentResults
	overview["last_updated"] = time.Now().UTC()

	return overview, nil
}

// Health Check Management

// CreateHealthCheck creates a new health check.
func (s *MonitoringService) CreateHealthCheck(healthCheck *models.HealthCheck) error {
	// Set default values
	if healthCheck.Type == "" {
		healthCheck.Type = "http"
	}
	if healthCheck.Method == "" {
		healthCheck.Method = "GET"
	}
	if healthCheck.ExpectedStatus == 0 {
		healthCheck.ExpectedStatus = 200
	}
	if healthCheck.Timeout == 0 {
		healthCheck.Timeout = 30
	}
	if healthCheck.IsActive == false && healthCheck.IsActive != true {
		healthCheck.IsActive = true
	}
	if healthCheck.LastResult == "" {
		healthCheck.LastResult = "unknown"
	}

	// Create health check
	if err := s.db.Create(healthCheck).Error; err != nil {
		s.logger.Error("Failed to create health check", zap.Error(err))
		return fmt.Errorf("failed to create health check: %w", err)
	}

	s.logger.Info("Health check created successfully", zap.Uint("health_check_id", healthCheck.ID))
	return nil
}

// UpdateHealthCheck updates a health check.
func (s *MonitoringService) UpdateHealthCheck(healthCheck *models.HealthCheck) error {
	if err := s.db.Save(healthCheck).Error; err != nil {
		s.logger.Error("Failed to update health check", zap.Error(err))
		return fmt.Errorf("failed to update health check: %w", err)
	}

	s.logger.Info("Health check updated successfully", zap.Uint("health_check_id", healthCheck.ID))
	return nil
}

// DeleteHealthCheck deletes a health check.
func (s *MonitoringService) DeleteHealthCheck(id uint) error {
	if err := s.db.Delete(&models.HealthCheck{}, id).Error; err != nil {
		s.logger.Error("Failed to delete health check", zap.Error(err))
		return fmt.Errorf("failed to delete health check: %w", err)
	}

	s.logger.Info("Health check deleted successfully", zap.Uint("health_check_id", id))
	return nil
}

// Alert Management

// CreateAlert creates a new alert.
func (s *MonitoringService) CreateAlert(alert *models.Alert) error {
	// Set default values
	if alert.Type == "" {
		alert.Type = "custom"
	}
	if alert.Severity == "" {
		alert.Severity = "info"
	}
	if alert.Status == "" {
		alert.Status = "active"
	}
	if alert.TriggeredAt.IsZero() {
		alert.TriggeredAt = time.Now()
	}

	// Create alert
	if err := s.db.Create(alert).Error; err != nil {
		s.logger.Error("Failed to create alert", zap.Error(err))
		return fmt.Errorf("failed to create alert: %w", err)
	}

	s.logger.Info("Alert created successfully", zap.Uint("alert_id", alert.ID))
	return nil
}

// GetAlert retrieves an alert by ID.
func (s *MonitoringService) GetAlert(id uint) (*models.Alert, error) {
	var alert models.Alert
	if err := s.db.Preload("Service").First(&alert, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("alert not found")
		}
		s.logger.Error("Failed to get alert", zap.Error(err))
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	return &alert, nil
}

// GetAlerts retrieves a list of alerts with pagination.
func (s *MonitoringService) GetAlerts(tenantID uint, limit, offset int) ([]*models.Alert, int64, error) {
	var alerts []*models.Alert
	var total int64

	// Get total count
	if err := s.db.Model(&models.Alert{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count alerts", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count alerts: %w", err)
	}

	// Get alerts with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).Preload("Service").Limit(limit).Offset(offset).Order("created_at DESC").Find(&alerts).Error; err != nil {
		s.logger.Error("Failed to get alerts", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get alerts: %w", err)
	}

	return alerts, total, nil
}

// UpdateAlert updates an alert.
func (s *MonitoringService) UpdateAlert(alert *models.Alert) error {
	if err := s.db.Save(alert).Error; err != nil {
		s.logger.Error("Failed to update alert", zap.Error(err))
		return fmt.Errorf("failed to update alert: %w", err)
	}

	s.logger.Info("Alert updated successfully", zap.Uint("alert_id", alert.ID))
	return nil
}

// DeleteAlert soft deletes an alert.
func (s *MonitoringService) DeleteAlert(id uint) error {
	if err := s.db.Delete(&models.Alert{}, id).Error; err != nil {
		s.logger.Error("Failed to delete alert", zap.Error(err))
		return fmt.Errorf("failed to delete alert: %w", err)
	}

	s.logger.Info("Alert deleted successfully", zap.Uint("alert_id", id))
	return nil
}

// AcknowledgeAlert acknowledges an alert.
func (s *MonitoringService) AcknowledgeAlert(alertID uint, userID uint) error {
	var alert models.Alert
	if err := s.db.First(&alert, alertID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("alert not found")
		}
		s.logger.Error("Failed to get alert", zap.Error(err))
		return fmt.Errorf("failed to get alert: %w", err)
	}

	alert.Status = "acknowledged"
	alert.AcknowledgedBy = &userID
	now := time.Now()
	alert.AcknowledgedAt = &now

	if err := s.db.Save(&alert).Error; err != nil {
		s.logger.Error("Failed to acknowledge alert", zap.Error(err))
		return fmt.Errorf("failed to acknowledge alert: %w", err)
	}

	s.logger.Info("Alert acknowledged successfully", zap.Uint("alert_id", alertID))
	return nil
}

// ResolveAlert resolves an alert.
func (s *MonitoringService) ResolveAlert(alertID uint, userID uint) error {
	var alert models.Alert
	if err := s.db.First(&alert, alertID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("alert not found")
		}
		s.logger.Error("Failed to get alert", zap.Error(err))
		return fmt.Errorf("failed to get alert: %w", err)
	}

	alert.Status = "resolved"
	alert.ResolvedBy = &userID
	now := time.Now()
	alert.ResolvedAt = &now

	if err := s.db.Save(&alert).Error; err != nil {
		s.logger.Error("Failed to resolve alert", zap.Error(err))
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	s.logger.Info("Alert resolved successfully", zap.Uint("alert_id", alertID))
	return nil
}

// GetPublicStatus returns public status information.
func (s *MonitoringService) GetPublicStatus(tenantID uint) (map[string]interface{}, error) {
	var services []models.MonitoredService
	if err := s.db.Where("tenant_id = ? AND is_active = ?", tenantID, true).Find(&services).Error; err != nil {
		s.logger.Error("Failed to get public services", zap.Error(err))
		return nil, fmt.Errorf("failed to get public services: %w", err)
	}

	// Calculate overall status
	overallStatus := "operational"
	unhealthyCount := 0
	for _, service := range services {
		if service.Status == "unhealthy" {
			unhealthyCount++
		}
	}

	if unhealthyCount > 0 {
		if unhealthyCount == len(services) {
			overallStatus = "major_outage"
		} else {
			overallStatus = "partial_outage"
		}
	}

	status := map[string]interface{}{
		"overall_status":  overallStatus,
		"services":        services,
		"unhealthy_count": unhealthyCount,
		"total_services":  len(services),
		"last_updated":    time.Now().UTC(),
	}

	return status, nil
}

// GetPublicHealth returns public health information.
func (s *MonitoringService) GetPublicHealth(tenantID uint) (map[string]interface{}, error) {
	// Get monitoring overview
	overview, err := s.GetMonitoringOverview(tenantID)
	if err != nil {
		return nil, err
	}

	// Add health-specific information
	health := map[string]interface{}{
		"status":    "healthy",
		"overview":  overview,
		"timestamp": time.Now().UTC(),
	}

	return health, nil
}

// InitDatabase initializes the database connection and runs migrations.
func InitDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// NOTE: Database migrations managed by Atlas (see ../../../../migrations/ and atlas.hcl)
	// AutoMigrate is NOT used - all schema changes via version-controlled migrations

	return db, nil
}
