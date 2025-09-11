package services

import (
	"context"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MonitoringAnalyticsService handles monitoring analytics and graph data
type MonitoringAnalyticsService struct {
	db *gorm.DB
}

// NewMonitoringAnalyticsService creates a new monitoring analytics service
func NewMonitoringAnalyticsService(db *gorm.DB) *MonitoringAnalyticsService {
	return &MonitoringAnalyticsService{
		db: db,
	}
}

// UptimeData represents uptime data for a service
type UptimeData struct {
	ServiceID    uint      `json:"service_id"`
	ServiceName  string    `json:"service_name"`
	Timestamp    time.Time `json:"timestamp"`
	Uptime       float64   `json:"uptime"`        // Percentage (0-100)
	Downtime     float64   `json:"downtime"`      // Percentage (0-100)
	Status       string    `json:"status"`        // "up", "down", "degraded"
	ResponseTime float64   `json:"response_time"` // Milliseconds
	ErrorRate    float64   `json:"error_rate"`    // Percentage (0-100)
}

// ResponseTimeData represents response time data
type ResponseTimeData struct {
	ServiceID    uint      `json:"service_id"`
	Timestamp    time.Time `json:"timestamp"`
	ResponseTime float64   `json:"response_time"` // Milliseconds
	MinTime      float64   `json:"min_time"`
	MaxTime      float64   `json:"max_time"`
	AvgTime      float64   `json:"avg_time"`
	P95Time      float64   `json:"p95_time"`
	P99Time      float64   `json:"p99_time"`
}

// ServiceHealthData represents overall service health
type ServiceHealthData struct {
	ServiceID        uint      `json:"service_id"`
	ServiceName      string    `json:"service_name"`
	Status           string    `json:"status"`
	Uptime           float64   `json:"uptime"`
	ResponseTime     float64   `json:"response_time"`
	ErrorRate        float64   `json:"error_rate"`
	LastChecked      time.Time `json:"last_checked"`
	IncidentCount    int       `json:"incident_count"`
	MaintenanceCount int       `json:"maintenance_count"`
}

// MonitoringMetrics represents comprehensive monitoring metrics
type MonitoringMetrics struct {
	Period              string               `json:"period"`
	StartTime           time.Time            `json:"start_time"`
	EndTime             time.Time            `json:"end_time"`
	OverallUptime       float64              `json:"overall_uptime"`
	Services            []ServiceHealthData  `json:"services"`
	UptimeHistory       []UptimeData         `json:"uptime_history"`
	ResponseTimeHistory []ResponseTimeData   `json:"response_time_history"`
	Incidents           []models.Incident    `json:"incidents"`
	Maintenances        []models.Maintenance `json:"maintenances"`
}

// GetServiceUptimeData retrieves uptime data for a service
func (s *MonitoringAnalyticsService) GetServiceUptimeData(ctx context.Context, tenantID uint, serviceID uint, period string) ([]UptimeData, error) {
	// Calculate time range based on period
	endTime := time.Now()
	var startTime time.Time

	switch period {
	case "1h":
		startTime = endTime.Add(-1 * time.Hour)
	case "24h":
		startTime = endTime.Add(-24 * time.Hour)
	case "7d":
		startTime = endTime.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = endTime.Add(-30 * 24 * time.Hour)
	case "90d":
		startTime = endTime.Add(-90 * 24 * time.Hour)
	default:
		startTime = endTime.Add(-24 * time.Hour)
	}

	// In a real implementation, you'd query actual monitoring data
	// For now, we'll generate realistic mock data
	var uptimeData []UptimeData

	// Get service info
	var service models.Service
	if err := s.db.Where("id = ? AND tenant_id = ?", serviceID, tenantID).First(&service).Error; err != nil {
		return nil, fmt.Errorf("service not found: %w", err)
	}

	// Generate mock uptime data
	interval := time.Minute * 5 // 5-minute intervals
	for t := startTime; t.Before(endTime); t = t.Add(interval) {
		// Simulate realistic uptime data
		uptime := 99.5 + (float64(t.Unix()%100)-50)*0.01        // 99.0-100.0%
		responseTime := 150.0 + (float64(t.Unix()%200)-100)*0.5 // 100-200ms
		errorRate := 0.1 + (float64(t.Unix()%50)-25)*0.002      // 0.05-0.15%

		status := "up"
		if uptime < 99.0 {
			status = "degraded"
		}
		if uptime < 95.0 {
			status = "down"
		}

		uptimeData = append(uptimeData, UptimeData{
			ServiceID:    serviceID,
			ServiceName:  service.Name,
			Timestamp:    t,
			Uptime:       uptime,
			Downtime:     100.0 - uptime,
			Status:       status,
			ResponseTime: responseTime,
			ErrorRate:    errorRate,
		})
	}

	return uptimeData, nil
}

// GetServiceResponseTimeData retrieves response time data for a service
func (s *MonitoringAnalyticsService) GetServiceResponseTimeData(ctx context.Context, tenantID uint, serviceID uint, period string) ([]ResponseTimeData, error) {
	// Calculate time range based on period
	endTime := time.Now()
	var startTime time.Time

	switch period {
	case "1h":
		startTime = endTime.Add(-1 * time.Hour)
	case "24h":
		startTime = endTime.Add(-24 * time.Hour)
	case "7d":
		startTime = endTime.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = endTime.Add(-30 * 24 * time.Hour)
	case "90d":
		startTime = endTime.Add(-90 * 24 * time.Hour)
	default:
		startTime = endTime.Add(-24 * time.Hour)
	}

	// Generate mock response time data
	var responseTimeData []ResponseTimeData

	interval := time.Minute * 5 // 5-minute intervals
	for t := startTime; t.Before(endTime); t = t.Add(interval) {
		// Simulate realistic response time data
		baseTime := 150.0
		variation := (float64(t.Unix()%200) - 100) * 0.5
		responseTime := baseTime + variation

		minTime := responseTime * 0.8
		maxTime := responseTime * 1.5
		avgTime := responseTime
		p95Time := responseTime * 1.2
		p99Time := responseTime * 1.4

		responseTimeData = append(responseTimeData, ResponseTimeData{
			ServiceID:    serviceID,
			Timestamp:    t,
			ResponseTime: responseTime,
			MinTime:      minTime,
			MaxTime:      maxTime,
			AvgTime:      avgTime,
			P95Time:      p95Time,
			P99Time:      p99Time,
		})
	}

	return responseTimeData, nil
}

// GetServiceHealthData retrieves health data for all services
func (s *MonitoringAnalyticsService) GetServiceHealthData(ctx context.Context, tenantID uint) ([]ServiceHealthData, error) {
	// Get all services for the tenant
	var services []models.Service
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&services).Error; err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	var healthData []ServiceHealthData

	for _, service := range services {
		// Get recent incidents
		var incidentCount int64
		s.db.Model(&models.Incident{}).Where("service_id = ? AND tenant_id = ? AND created_at > ?",
			service.ID, tenantID, time.Now().Add(-30*24*time.Hour)).Count(&incidentCount)

		// Get recent maintenances
		var maintenanceCount int64
		s.db.Model(&models.Maintenance{}).Where("service_id = ? AND tenant_id = ? AND created_at > ?",
			service.ID, tenantID, time.Now().Add(-30*24*time.Hour)).Count(&maintenanceCount)

		// Calculate current health metrics
		uptime := 99.5 + (float64(service.ID%100)-50)*0.01
		responseTime := 150.0 + (float64(service.ID%200)-100)*0.5
		errorRate := 0.1 + (float64(service.ID%50)-25)*0.002

		status := "up"
		if uptime < 99.0 {
			status = "degraded"
		}
		if uptime < 95.0 {
			status = "down"
		}

		healthData = append(healthData, ServiceHealthData{
			ServiceID:        service.ID,
			ServiceName:      service.Name,
			Status:           status,
			Uptime:           uptime,
			ResponseTime:     responseTime,
			ErrorRate:        errorRate,
			LastChecked:      time.Now().Add(-time.Minute * 2),
			IncidentCount:    int(incidentCount),
			MaintenanceCount: int(maintenanceCount),
		})
	}

	return healthData, nil
}

// GetMonitoringMetrics retrieves comprehensive monitoring metrics
func (s *MonitoringAnalyticsService) GetMonitoringMetrics(ctx context.Context, tenantID uint, period string) (*MonitoringMetrics, error) {
	// Calculate time range
	endTime := time.Now()
	var startTime time.Time

	switch period {
	case "1h":
		startTime = endTime.Add(-1 * time.Hour)
	case "24h":
		startTime = endTime.Add(-24 * time.Hour)
	case "7d":
		startTime = endTime.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = endTime.Add(-30 * 24 * time.Hour)
	case "90d":
		startTime = endTime.Add(-90 * 24 * time.Hour)
	default:
		startTime = endTime.Add(-24 * time.Hour)
	}

	// Get service health data
	services, err := s.GetServiceHealthData(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get service health data: %w", err)
	}

	// Calculate overall uptime
	var totalUptime float64
	for _, service := range services {
		totalUptime += service.Uptime
	}
	overallUptime := totalUptime / float64(len(services))

	// Get incidents in the period
	var incidents []models.Incident
	s.db.Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, startTime, endTime).Find(&incidents)

	// Get maintenances in the period
	var maintenances []models.Maintenance
	s.db.Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, startTime, endTime).Find(&maintenances)

	// Generate uptime history for all services
	var uptimeHistory []UptimeData
	for _, service := range services {
		serviceUptime, err := s.GetServiceUptimeData(ctx, tenantID, service.ServiceID, period)
		if err != nil {
			logger.Warn("Failed to get uptime data for service",
				zap.Uint("service_id", service.ServiceID),
				zap.Error(err))
			continue
		}
		uptimeHistory = append(uptimeHistory, serviceUptime...)
	}

	// Generate response time history for all services
	var responseTimeHistory []ResponseTimeData
	for _, service := range services {
		serviceResponseTime, err := s.GetServiceResponseTimeData(ctx, tenantID, service.ServiceID, period)
		if err != nil {
			logger.Warn("Failed to get response time data for service",
				zap.Uint("service_id", service.ServiceID),
				zap.Error(err))
			continue
		}
		responseTimeHistory = append(responseTimeHistory, serviceResponseTime...)
	}

	metrics := &MonitoringMetrics{
		Period:              period,
		StartTime:           startTime,
		EndTime:             endTime,
		OverallUptime:       overallUptime,
		Services:            services,
		UptimeHistory:       uptimeHistory,
		ResponseTimeHistory: responseTimeHistory,
		Incidents:           incidents,
		Maintenances:        maintenances,
	}

	return metrics, nil
}

// GetServiceStatusHistory retrieves status history for a service
func (s *MonitoringAnalyticsService) GetServiceStatusHistory(ctx context.Context, tenantID uint, serviceID uint, period string) ([]map[string]interface{}, error) {
	// Get uptime data
	uptimeData, err := s.GetServiceUptimeData(ctx, tenantID, serviceID, period)
	if err != nil {
		return nil, err
	}

	var statusHistory []map[string]interface{}

	for _, data := range uptimeData {
		statusHistory = append(statusHistory, map[string]interface{}{
			"timestamp":     data.Timestamp,
			"status":        data.Status,
			"uptime":        data.Uptime,
			"response_time": data.ResponseTime,
			"error_rate":    data.ErrorRate,
		})
	}

	return statusHistory, nil
}

// GetUptimeSummary retrieves uptime summary for all services
func (s *MonitoringAnalyticsService) GetUptimeSummary(ctx context.Context, tenantID uint) (map[string]interface{}, error) {
	// Get service health data
	services, err := s.GetServiceHealthData(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Calculate summary statistics
	var totalServices int
	var upServices int
	var degradedServices int
	var downServices int
	var totalUptime float64

	for _, service := range services {
		totalServices++
		totalUptime += service.Uptime

		switch service.Status {
		case "up":
			upServices++
		case "degraded":
			degradedServices++
		case "down":
			downServices++
		}
	}

	avgUptime := totalUptime / float64(totalServices)

	summary := map[string]interface{}{
		"total_services":    totalServices,
		"up_services":       upServices,
		"degraded_services": degradedServices,
		"down_services":     downServices,
		"average_uptime":    avgUptime,
		"overall_status":    s.calculateOverallStatus(upServices, degradedServices, downServices),
		"last_updated":      time.Now(),
	}

	return summary, nil
}

// Helper function to calculate overall status
func (s *MonitoringAnalyticsService) calculateOverallStatus(up, degraded, down int) string {
	_ = up // TODO: Use up count for more detailed status calculation
	if down > 0 {
		return "down"
	}
	if degraded > 0 {
		return "degraded"
	}
	return "up"
}
