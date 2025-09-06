package services

import (
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UptimeService struct {
	db *gorm.DB
}

func NewUptimeService() *UptimeService {
	return &UptimeService{
		db: database.GetDB(),
	}
}

// UptimeStats holds the calculated uptime statistics for a monitor.
type UptimeStats struct {
	MonitorID    uint    `json:"monitor_id"`
	ServiceName  string  `json:"service_name"`
	Uptime24H    float64 `json:"uptime_24h"`
	Uptime7D     float64 `json:"uptime_7d"`
	Uptime30D    float64 `json:"uptime_30d"`
	Uptime90D    float64 `json:"uptime_90d"`
	LatestStatus string  `json:"latest_status"`
	LastCheckAt  time.Time `json:"last_check_at"`
	TotalChecks  int64   `json:"total_checks"`
	FailedChecks int64   `json:"failed_checks"`
}

// UptimeDataPoint represents a single data point in uptime history
type UptimeDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Uptime    float64   `json:"uptime"`
	Status    string    `json:"status"`
}

// UptimeHistory represents historical uptime data
type UptimeHistory struct {
	MonitorID uint              `json:"monitor_id"`
	ServiceName string          `json:"service_name"`
	Period    string            `json:"period"` // 24h, 7d, 30d, 90d
	DataPoints []UptimeDataPoint `json:"data_points"`
	AverageUptime float64       `json:"average_uptime"`
}

// AnalyticsSummary provides overall analytics for the status page
type AnalyticsSummary struct {
	OverallUptime24H float64 `json:"overall_uptime_24h"`
	OverallUptime7D  float64 `json:"overall_uptime_7d"`
	OverallUptime30D float64 `json:"overall_uptime_30d"`
	OverallUptime90D float64 `json:"overall_uptime_90d"`
	TotalIncidents   int64   `json:"total_incidents"`
	ResolvedIncidents int64  `json:"resolved_incidents"`
	ActiveIncidents  int64   `json:"active_incidents"`
	MTTR            float64  `json:"mttr_hours"` // Mean Time To Resolution in hours
	MTTA            float64  `json:"mtta_hours"` // Mean Time To Acknowledgment in hours
	ServiceCount    int64    `json:"service_count"`
	MonitorCount    int64    `json:"monitor_count"`
}

// GetUptimeStatsForMonitors calculates uptime stats for all monitors.
func (s *UptimeService) GetUptimeStatsForMonitors() ([]UptimeStats, error) {
	var monitors []models.Monitor
	if err := s.db.Preload("Service").Find(&monitors).Error; err != nil {
		return nil, err
	}

	stats := make([]UptimeStats, len(monitors))
	for i, monitor := range monitors {
		uptime24h, _ := s.calculateUptime(monitor.ID, 24*time.Hour)
		uptime7d, _ := s.calculateUptime(monitor.ID, 7*24*time.Hour)
		uptime30d, _ := s.calculateUptime(monitor.ID, 30*24*time.Hour)
		uptime90d, _ := s.calculateUptime(monitor.ID, 90*24*time.Hour)

		// Get check counts
		var totalChecks, failedChecks int64
		s.db.Model(&models.Heartbeat{}).Where("monitor_id = ? AND created_at >= ?", monitor.ID, time.Now().Add(-30*24*time.Hour)).Count(&totalChecks)
		s.db.Model(&models.Heartbeat{}).Where("monitor_id = ? AND created_at >= ? AND status = ?", monitor.ID, time.Now().Add(-30*24*time.Hour), "down").Count(&failedChecks)

		stats[i] = UptimeStats{
			MonitorID:    monitor.ID,
			ServiceName:  monitor.Service.Name,
			Uptime24H:    uptime24h,
			Uptime7D:     uptime7d,
			Uptime30D:    uptime30d,
			Uptime90D:    uptime90d,
			LatestStatus: monitor.LastResult,
			LastCheckAt:  monitor.LastCheckAt,
			TotalChecks:  totalChecks,
			FailedChecks: failedChecks,
		}
	}

	return stats, nil
}

// GetUptimeHistory returns historical uptime data for a specific monitor
func (s *UptimeService) GetUptimeHistory(monitorID uint, period string) (*UptimeHistory, error) {
	var monitor models.Monitor
	if err := s.db.Preload("Service").First(&monitor, monitorID).Error; err != nil {
		return nil, err
	}

	var duration time.Duration
	var interval time.Duration
	
	switch period {
	case "24h":
		duration = 24 * time.Hour
		interval = 1 * time.Hour
	case "7d":
		duration = 7 * 24 * time.Hour
		interval = 6 * time.Hour
	case "30d":
		duration = 30 * 24 * time.Hour
		interval = 24 * time.Hour
	case "90d":
		duration = 90 * 24 * time.Hour
		interval = 24 * time.Hour
	default:
		return nil, fmt.Errorf("invalid period: %s", period)
	}

	var dataPoints []UptimeDataPoint
	startTime := time.Now().Add(-duration)
	
	for currentTime := startTime; currentTime.Before(time.Now()); currentTime = currentTime.Add(interval) {
		endTime := currentTime.Add(interval)
		
		uptime, err := s.calculateUptimeForPeriod(monitorID, currentTime, endTime)
		if err != nil {
			logger.Error("Failed to calculate uptime for period", zap.Error(err))
			continue
		}

		// Get latest status for this period
		var latestHeartbeat models.Heartbeat
		s.db.Where("monitor_id = ? AND timestamp >= ? AND timestamp < ?", monitorID, currentTime, endTime).
			Order("timestamp DESC").First(&latestHeartbeat)

		status := "unknown"
		if latestHeartbeat.ID != 0 {
			status = latestHeartbeat.Status
		}

		dataPoints = append(dataPoints, UptimeDataPoint{
			Timestamp: currentTime,
			Uptime:    uptime,
			Status:    status,
		})
	}

	// Calculate average uptime
	var totalUptime float64
	for _, point := range dataPoints {
		totalUptime += point.Uptime
	}
	averageUptime := float64(0)
	if len(dataPoints) > 0 {
		averageUptime = totalUptime / float64(len(dataPoints))
	}

	return &UptimeHistory{
		MonitorID:     monitorID,
		ServiceName:   monitor.Service.Name,
		Period:        period,
		DataPoints:    dataPoints,
		AverageUptime: averageUptime,
	}, nil
}

// GetAnalyticsSummary returns overall analytics for the status page
func (s *UptimeService) GetAnalyticsSummary() (*AnalyticsSummary, error) {
	// Get overall uptime stats
	overallUptime24H, _ := s.calculateOverallUptime(24 * time.Hour)
	overallUptime7D, _ := s.calculateOverallUptime(7 * 24 * time.Hour)
	overallUptime30D, _ := s.calculateOverallUptime(30 * 24 * time.Hour)
	overallUptime90D, _ := s.calculateOverallUptime(90 * 24 * time.Hour)

	// Get incident stats
	var totalIncidents, resolvedIncidents, activeIncidents int64
	s.db.Model(&models.Incident{}).Count(&totalIncidents)
	s.db.Model(&models.Incident{}).Where("status = ?", models.IncidentStatusResolved).Count(&resolvedIncidents)
	s.db.Model(&models.Incident{}).Where("status != ?", models.IncidentStatusResolved).Count(&activeIncidents)

	// Calculate MTTR and MTTA
	mttr, _ := s.calculateMTTR()
	mtta, _ := s.calculateMTTA()

	// Get service and monitor counts
	var serviceCount, monitorCount int64
	s.db.Model(&models.Service{}).Count(&serviceCount)
	s.db.Model(&models.Monitor{}).Count(&monitorCount)

	return &AnalyticsSummary{
		OverallUptime24H: overallUptime24H,
		OverallUptime7D:  overallUptime7D,
		OverallUptime30D: overallUptime30D,
		OverallUptime90D: overallUptime90D,
		TotalIncidents:   totalIncidents,
		ResolvedIncidents: resolvedIncidents,
		ActiveIncidents:  activeIncidents,
		MTTR:            mttr,
		MTTA:            mtta,
		ServiceCount:    serviceCount,
		MonitorCount:    monitorCount,
	}, nil
}

func (s *UptimeService) calculateUptime(monitorID uint, duration time.Duration) (float64, error) {
	var upCount, totalCount int64
	since := time.Now().Add(-duration)

	s.db.Model(&models.Heartbeat{}).
		Where("monitor_id = ? AND created_at >= ? AND status = ?", monitorID, since, "up").
		Count(&upCount)

	s.db.Model(&models.Heartbeat{}).
		Where("monitor_id = ? AND created_at >= ?", monitorID, since).
		Count(&totalCount)

	if totalCount == 0 {
		return 100.0, nil // Default to 100% if no data is available
	}

	return (float64(upCount) / float64(totalCount)) * 100, nil
}

func (s *UptimeService) calculateUptimeForPeriod(monitorID uint, startTime, endTime time.Time) (float64, error) {
	var upCount, totalCount int64

	s.db.Model(&models.Heartbeat{}).
		Where("monitor_id = ? AND timestamp >= ? AND timestamp < ? AND status = ?", monitorID, startTime, endTime, "up").
		Count(&upCount)

	s.db.Model(&models.Heartbeat{}).
		Where("monitor_id = ? AND timestamp >= ? AND timestamp < ?", monitorID, startTime, endTime).
		Count(&totalCount)

	if totalCount == 0 {
		return 100.0, nil // Default to 100% if no data is available
	}

	return (float64(upCount) / float64(totalCount)) * 100, nil
}

func (s *UptimeService) calculateOverallUptime(duration time.Duration) (float64, error) {
	var monitors []models.Monitor
	if err := s.db.Find(&monitors).Error; err != nil {
		return 0, err
	}

	if len(monitors) == 0 {
		return 100.0, nil
	}

	var totalUptime float64
	for _, monitor := range monitors {
		uptime, _ := s.calculateUptime(monitor.ID, duration)
		totalUptime += uptime
	}

	return totalUptime / float64(len(monitors)), nil
}

func (s *UptimeService) calculateMTTR() (float64, error) {
	var incidents []models.Incident
	if err := s.db.Where("resolved_at IS NOT NULL").Find(&incidents).Error; err != nil {
		return 0, err
	}

	if len(incidents) == 0 {
		return 0, nil
	}

	var totalDuration time.Duration
	for _, incident := range incidents {
		duration := incident.ResolvedAt.Sub(incident.CreatedAt)
		totalDuration += duration
	}

	return totalDuration.Hours() / float64(len(incidents)), nil
}

func (s *UptimeService) calculateMTTA() (float64, error) {
	var incidents []models.Incident
	if err := s.db.Where("status != ?", models.IncidentStatusInvestigating).Find(&incidents).Error; err != nil {
		return 0, err
	}

	if len(incidents) == 0 {
		return 0, nil
	}

	var totalDuration time.Duration
	for _, incident := range incidents {
		// Find the first status update that's not "investigating"
		var firstUpdate models.StatusUpdate
		if err := s.db.Where("incident_id = ? AND status != ?", incident.ID, models.IncidentStatusInvestigating).
			Order("created_at ASC").First(&firstUpdate).Error; err == nil {
			duration := firstUpdate.CreatedAt.Sub(incident.CreatedAt)
			totalDuration += duration
		}
	}

	return totalDuration.Hours() / float64(len(incidents)), nil
}
