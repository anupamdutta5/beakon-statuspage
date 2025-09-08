package services

import (
	"fmt"
	"net/http"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
)

type MonitorService struct{}

func NewMonitorService() *MonitorService {
	return &MonitorService{}
}

func (s *MonitorService) GetAllMonitors() ([]models.Monitor, error) {
	var monitors []models.Monitor
	err := database.DB.Preload("Service").Find(&monitors).Error
	return monitors, err
}

func (s *MonitorService) GetMonitorByID(id uint) (*models.Monitor, error) {
	var monitor models.Monitor
	err := database.DB.Preload("Service").First(&monitor, id).Error
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
	if err := database.DB.Where("tenant_id = ?", tenantID).Preload("Service").Find(&monitors).Error; err != nil {
		return nil, err
	}
	return monitors, nil
}

// GetMonitorsByServiceID retrieves all monitors for a specific service
func (s *MonitorService) GetMonitorsByServiceID(serviceID uint) ([]models.Monitor, error) {
	var monitors []models.Monitor
	if err := database.DB.Where("service_id = ?", serviceID).Preload("Service").Find(&monitors).Error; err != nil {
		return nil, err
	}
	return monitors, nil
}

// GetMonitorsByType retrieves monitors by type for a tenant
func (s *MonitorService) GetMonitorsByType(tenantID uint, monitorType string) ([]models.Monitor, error) {
	var monitors []models.Monitor
	if err := database.DB.Where("tenant_id = ? AND type = ?", tenantID, monitorType).Preload("Service").Find(&monitors).Error; err != nil {
		return nil, err
	}
	return monitors, nil
}

// UpdateMonitorStatus updates the last check result for a monitor
func (s *MonitorService) UpdateMonitorStatus(id uint, status string, latency int64, message string) error {
	now := time.Now()
	return database.DB.Model(&models.Monitor{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_check_at": now,
			"last_result":   status,
		}).Error
}

// RecordHeartbeat records a heartbeat result for a monitor
func (s *MonitorService) RecordHeartbeat(monitorID uint, status string, latency int64, message string) error {
	heartbeat := &models.Heartbeat{
		MonitorID: monitorID,
		Timestamp: time.Now(),
		Status:    status,
		Latency:   latency,
		Message:   message,
	}

	if err := database.DB.Create(heartbeat).Error; err != nil {
		return err
	}

	// Update monitor's last check
	return s.UpdateMonitorStatus(monitorID, status, latency, message)
}

// GetMonitorHeartbeats retrieves heartbeat history for a monitor
func (s *MonitorService) GetMonitorHeartbeats(monitorID uint, hours int) ([]models.Heartbeat, error) {
	var heartbeats []models.Heartbeat
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	if err := database.DB.Where("monitor_id = ? AND timestamp >= ?", monitorID, since).
		Order("timestamp DESC").
		Find(&heartbeats).Error; err != nil {
		return nil, err
	}
	return heartbeats, nil
}

// GetMonitorUptime calculates uptime percentage for a monitor
func (s *MonitorService) GetMonitorUptime(monitorID uint, hours int) (float64, error) {
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	var totalChecks int64
	var upChecks int64

	if err := database.DB.Model(&models.Heartbeat{}).
		Where("monitor_id = ? AND timestamp >= ?", monitorID, since).
		Count(&totalChecks).Error; err != nil {
		return 0, err
	}

	if err := database.DB.Model(&models.Heartbeat{}).
		Where("monitor_id = ? AND timestamp >= ? AND status = ?", monitorID, since, "up").
		Count(&upChecks).Error; err != nil {
		return 0, err
	}

	if totalChecks == 0 {
		return 100.0, nil // No checks, assume operational
	}

	return float64(upChecks) / float64(totalChecks) * 100.0, nil
}

// GetMonitorStatistics returns monitoring statistics for a tenant
func (s *MonitorService) GetMonitorStatistics(tenantID uint) (*MonitorStatistics, error) {
	var stats MonitorStatistics

	// Total monitors
	if err := database.DB.Model(&models.Monitor{}).
		Where("tenant_id = ?", tenantID).
		Count(&stats.TotalMonitors).Error; err != nil {
		return nil, err
	}

	// Active monitors (checked in last hour)
	since := time.Now().Add(-1 * time.Hour)
	if err := database.DB.Model(&models.Monitor{}).
		Where("tenant_id = ? AND last_check_at >= ?", tenantID, since).
		Count(&stats.ActiveMonitors).Error; err != nil {
		return nil, err
	}

	// Monitors by status
	var statusStats []struct {
		Status string
		Count  int64
	}
	if err := database.DB.Model(&models.Monitor{}).
		Select("last_result as status, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("last_result").
		Scan(&statusStats).Error; err != nil {
		return nil, err
	}

	stats.MonitorsByStatus = make(map[string]int64)
	for _, stat := range statusStats {
		stats.MonitorsByStatus[stat.Status] = stat.Count
	}

	// Monitors by type
	var typeStats []struct {
		Type  string
		Count int64
	}
	if err := database.DB.Model(&models.Monitor{}).
		Select("type, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("type").
		Scan(&typeStats).Error; err != nil {
		return nil, err
	}

	stats.MonitorsByType = make(map[string]int64)
	for _, stat := range typeStats {
		stats.MonitorsByType[stat.Type] = stat.Count
	}

	return &stats, nil
}

// PerformHealthCheck performs a health check for a monitor
func (s *MonitorService) PerformHealthCheck(monitor *models.Monitor) (*HealthCheckResult, error) {
	start := time.Now()

	client := &http.Client{
		Timeout: time.Duration(monitor.Timeout) * time.Second,
	}

	resp, err := client.Get(monitor.URL)
	latency := time.Since(start).Milliseconds()

	result := &HealthCheckResult{
		MonitorID: monitor.ID,
		Timestamp: time.Now(),
		Latency:   latency,
	}

	if err != nil {
		result.Status = "down"
		result.Message = err.Error()
	} else {
		defer resp.Body.Close()

		if resp.StatusCode == monitor.ExpectedStatus {
			result.Status = "up"
			result.Message = "OK"
		} else {
			result.Status = "down"
			result.Message = fmt.Sprintf("Expected status %d, got %d", monitor.ExpectedStatus, resp.StatusCode)
		}
	}

	// Record the heartbeat
	if err := s.RecordHeartbeat(monitor.ID, result.Status, result.Latency, result.Message); err != nil {
		return result, err
	}

	return result, nil
}

// GetServiceHealthStatus returns the overall health status for a service
func (s *MonitorService) GetServiceHealthStatus(serviceID uint) (*ServiceHealthStatus, error) {
	var monitors []models.Monitor
	if err := database.DB.Where("service_id = ?", serviceID).Find(&monitors).Error; err != nil {
		return nil, err
	}

	if len(monitors) == 0 {
		return &ServiceHealthStatus{
			ServiceID: serviceID,
			Status:    "unknown",
			Message:   "No monitors configured",
		}, nil
	}

	// Check if any monitor is down
	for _, monitor := range monitors {
		if monitor.LastResult == "down" {
			return &ServiceHealthStatus{
				ServiceID: serviceID,
				Status:    "down",
				Message:   fmt.Sprintf("Monitor %s is down", monitor.Name),
			}, nil
		}
	}

	// Check if any monitor hasn't been checked recently
	recentThreshold := time.Now().Add(-5 * time.Minute)
	for _, monitor := range monitors {
		if monitor.LastCheckAt.Before(recentThreshold) {
			return &ServiceHealthStatus{
				ServiceID: serviceID,
				Status:    "degraded",
				Message:   fmt.Sprintf("Monitor %s hasn't been checked recently", monitor.Name),
			}, nil
		}
	}

	return &ServiceHealthStatus{
		ServiceID: serviceID,
		Status:    "up",
		Message:   "All monitors are operational",
	}, nil
}

// GetUptimeHistory returns uptime history for a monitor
func (s *MonitorService) GetUptimeHistory(monitorID uint, days int) ([]UptimeDataPoint, error) {
	var dataPoints []UptimeDataPoint

	// Get hourly uptime data
	for i := 0; i < days*24; i++ {
		hourStart := time.Now().Add(-time.Duration(i+1) * time.Hour)
		hourEnd := time.Now().Add(-time.Duration(i) * time.Hour)

		var totalChecks int64
		var upChecks int64

		database.DB.Model(&models.Heartbeat{}).
			Where("monitor_id = ? AND timestamp >= ? AND timestamp < ?", monitorID, hourStart, hourEnd).
			Count(&totalChecks)

		database.DB.Model(&models.Heartbeat{}).
			Where("monitor_id = ? AND timestamp >= ? AND timestamp < ? AND status = ?", monitorID, hourStart, hourEnd, "up").
			Count(&upChecks)

		uptime := 100.0
		if totalChecks > 0 {
			uptime = float64(upChecks) / float64(totalChecks) * 100.0
		}

		dataPoints = append(dataPoints, UptimeDataPoint{
			Timestamp: hourStart,
			Uptime:    uptime,
		})
	}

	return dataPoints, nil
}

// MonitorStatistics represents monitoring statistics
type MonitorStatistics struct {
	TotalMonitors    int64            `json:"total_monitors"`
	ActiveMonitors   int64            `json:"active_monitors"`
	MonitorsByStatus map[string]int64 `json:"monitors_by_status"`
	MonitorsByType   map[string]int64 `json:"monitors_by_type"`
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	MonitorID uint      `json:"monitor_id"`
	Status    string    `json:"status"`
	Latency   int64     `json:"latency"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// ServiceHealthStatus represents the health status of a service
type ServiceHealthStatus struct {
	ServiceID uint   `json:"service_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}
