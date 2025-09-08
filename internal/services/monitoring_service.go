package services

import (
	"fmt"
	"net/http"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MonitoringService struct {
	db         *gorm.DB
	httpClient *http.Client
}

func NewMonitoringService() *MonitoringService {
	return &MonitoringService{
		db:         database.DB,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// CheckServiceHealth performs health checks on services
func (s *MonitoringService) CheckServiceHealth(service *models.Service) (*models.HealthCheck, error) {
	startTime := time.Now()

	// Perform HTTP health check
	resp, err := s.httpClient.Get(service.HealthCheckURL)
	duration := time.Since(startTime)

	healthCheck := &models.HealthCheck{
		ServiceID:    service.ID,
		TenantID:     service.TenantID,
		Status:       "down",
		ResponseTime: duration.Milliseconds(),
		CheckedAt:    time.Now(),
		Error:        "",
	}

	if err != nil {
		healthCheck.Error = err.Error()
		healthCheck.Status = "down"
	} else {
		defer resp.Body.Close()
		healthCheck.Status = "up"
		healthCheck.StatusCode = resp.StatusCode
	}

	// Save health check result
	if err := s.db.Create(healthCheck).Error; err != nil {
		logger.Error("Failed to save health check", zap.Error(err))
		return nil, err
	}

	// Update service status based on health check
	if err := s.updateServiceStatus(service, healthCheck); err != nil {
		logger.Error("Failed to update service status", zap.Error(err))
	}

	return healthCheck, nil
}

// GetServiceMetrics retrieves metrics for a service
func (s *MonitoringService) GetServiceMetrics(serviceID uint, timeRange string) (map[string]interface{}, error) {
	// Return mock metrics for now
	return s.getMockMetrics(), nil
}

// CreateAlert creates a new alert rule
func (s *MonitoringService) CreateAlert(alert *models.Alert) error {
	if err := s.db.Create(alert).Error; err != nil {
		logger.Error("Failed to create alert", zap.Error(err))
		return err
	}

	// Create alert rule (mock implementation)
	if err := s.createPrometheusAlertRule(alert); err != nil {
		logger.Error("Failed to create alert rule", zap.Error(err))
		// Don't fail the alert creation, just log the error
	}

	logger.Info("Alert created successfully", zap.Uint("alert_id", alert.ID))
	return nil
}

// CheckAlerts evaluates all active alerts
func (s *MonitoringService) CheckAlerts() error {
	var alerts []models.Alert
	if err := s.db.Where("is_active = ?", true).Find(&alerts).Error; err != nil {
		return err
	}

	for _, alert := range alerts {
		if err := s.evaluateAlert(&alert); err != nil {
			logger.Error("Failed to evaluate alert", zap.Error(err), zap.Uint("alert_id", alert.ID))
		}
	}

	return nil
}

// GetUptimeStats calculates uptime statistics for a service
func (s *MonitoringService) GetUptimeStats(serviceID uint, days int) (*models.UptimeStats, error) {
	// Calculate time range
	end := time.Now()
	start := end.AddDate(0, 0, -days)

	// Get health checks for the time period
	var healthChecks []models.HealthCheck
	if err := s.db.Where("service_id = ? AND checked_at BETWEEN ? AND ?",
		serviceID, start, end).Find(&healthChecks).Error; err != nil {
		return nil, err
	}

	if len(healthChecks) == 0 {
		return &models.UptimeStats{
			ServiceID: serviceID,
			Period:    days,
			Uptime:    0,
			Downtime:  0,
			Checks:    0,
		}, nil
	}

	// Calculate uptime
	upCount := 0
	totalChecks := len(healthChecks)

	for _, check := range healthChecks {
		if check.Status == "up" {
			upCount++
		}
	}

	uptime := float64(upCount) / float64(totalChecks) * 100
	downtime := 100 - uptime

	return &models.UptimeStats{
		ServiceID: serviceID,
		Period:    days,
		Uptime:    uptime,
		Downtime:  downtime,
		Checks:    totalChecks,
		StartDate: start,
		EndDate:   end,
	}, nil
}

// Private helper methods

func (s *MonitoringService) updateServiceStatus(service *models.Service, healthCheck *models.HealthCheck) error {
	// Determine new status based on health check
	newStatus := "operational"
	if healthCheck.Status == "down" {
		newStatus = "degraded"
	}

	// Only update if status changed
	if service.Status != newStatus {
		service.Status = newStatus
		service.UpdatedAt = time.Now()

		if err := s.db.Save(service).Error; err != nil {
			return err
		}

		// Create incident if service goes down
		if newStatus == "degraded" {
			if err := s.createIncidentForService(service, healthCheck); err != nil {
				logger.Error("Failed to create incident for service", zap.Error(err))
			}
		}
	}

	return nil
}

func (s *MonitoringService) createIncidentForService(service *models.Service, healthCheck *models.HealthCheck) error {
	// Check if there's already an active incident for this service
	var existingIncident models.Incident
	if err := s.db.Where("service_id = ? AND status IN ?",
		service.ID, []string{"investigating", "identified", "monitoring"}).First(&existingIncident).Error; err == nil {
		// Incident already exists, don't create a new one
		return nil
	}

	// Create new incident
	incident := &models.Incident{
		TenantID:    service.TenantID,
		Title:       fmt.Sprintf("%s is experiencing issues", service.Name),
		Description: fmt.Sprintf("Health check failed: %s", healthCheck.Error),
		Status:      "investigating",
		Impact:      "major",
	}

	if err := s.db.Create(incident).Error; err != nil {
		return err
	}

	logger.Info("Incident created for service",
		zap.Uint("service_id", service.ID),
		zap.Uint("incident_id", incident.ID))

	return nil
}

func (s *MonitoringService) createPrometheusAlertRule(alert *models.Alert) error {
	// This would typically create an alert rule in Prometheus
	// For now, we'll just log that we would create the rule
	logger.Info("Would create Prometheus alert rule",
		zap.String("name", alert.Name),
		zap.String("query", alert.Query),
		zap.Float64("threshold", alert.Threshold))

	return nil
}

func (s *MonitoringService) evaluateAlert(alert *models.Alert) error {
	// Mock alert evaluation
	return s.evaluateMockAlert(alert)
}

func (s *MonitoringService) evaluateMockAlert(alert *models.Alert) error {
	// Mock alert evaluation for testing
	// Randomly fire alerts for demonstration
	shouldFire := time.Now().Unix()%10 == 0 // 10% chance of firing

	if shouldFire && !alert.IsFiring {
		alert.IsFiring = true
		alert.LastFiredAt = &[]time.Time{time.Now()}[0]
	} else if !shouldFire && alert.IsFiring {
		alert.IsFiring = false
		alert.LastResolvedAt = &[]time.Time{time.Now()}[0]
	}

	return s.db.Save(alert).Error
}

func (s *MonitoringService) getMockMetrics() map[string]interface{} {
	// Return mock metrics when Prometheus is not available
	return map[string]interface{}{
		"uptime": []map[string]interface{}{
			{"timestamp": time.Now().Unix(), "value": 1.0},
		},
		"response_time": []map[string]interface{}{
			{"timestamp": time.Now().Unix(), "value": 0.5},
		},
		"error_rate": []map[string]interface{}{
			{"timestamp": time.Now().Unix(), "value": 0.01},
		},
		"request_rate": []map[string]interface{}{
			{"timestamp": time.Now().Unix(), "value": 100.0},
		},
	}
}
