package services

import (
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
)

// AnalyticsService provides methods for analytics and reporting.
type AnalyticsService struct{}

// NewAnalyticsService creates a new AnalyticsService.
func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{}
}

// GetUptimeReport returns uptime statistics for services
func (s *AnalyticsService) GetUptimeReport(tenantID uint, days int) (*UptimeReport, error) {
	report := &UptimeReport{
		TenantID: tenantID,
		Period:   days,
		Services: make(map[uint]*ServiceUptime),
	}

	// Get all services for the tenant
	var services []models.Service
	if err := database.DB.Where("tenant_id = ?", tenantID).Find(&services).Error; err != nil {
		return nil, err
	}

	since := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	for _, service := range services {
		serviceUptime := &ServiceUptime{
			ServiceID:   service.ID,
			ServiceName: service.Name,
		}

		// Get monitors for this service
		var monitors []models.Monitor
		if err := database.DB.Where("service_id = ?", service.ID).Find(&monitors).Error; err != nil {
			continue
		}

		if len(monitors) == 0 {
			serviceUptime.Uptime = 100.0 // No monitors, assume operational
			report.Services[service.ID] = serviceUptime
			continue
		}

		// Calculate uptime based on monitor heartbeats
		var totalChecks int64
		var upChecks int64

		for _, monitor := range monitors {
			database.DB.Model(&models.Heartbeat{}).
				Where("monitor_id = ? AND timestamp >= ?", monitor.ID, since).
				Count(&totalChecks)

			database.DB.Model(&models.Heartbeat{}).
				Where("monitor_id = ? AND timestamp >= ? AND status = ?", monitor.ID, since, "up").
				Count(&upChecks)
		}

		if totalChecks > 0 {
			serviceUptime.Uptime = float64(upChecks) / float64(totalChecks) * 100.0
		} else {
			serviceUptime.Uptime = 100.0
		}

		report.Services[service.ID] = serviceUptime
	}

	// Calculate overall uptime
	var totalUptime float64
	var serviceCount int
	for _, serviceUptime := range report.Services {
		totalUptime += serviceUptime.Uptime
		serviceCount++
	}

	if serviceCount > 0 {
		report.OverallUptime = totalUptime / float64(serviceCount)
	}

	return report, nil
}

// GetIncidentAnalytics returns incident analytics for a tenant
func (s *AnalyticsService) GetIncidentAnalytics(tenantID uint, days int) (*IncidentAnalytics, error) {
	analytics := &IncidentAnalytics{
		TenantID: tenantID,
		Period:   days,
	}

	since := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	// Total incidents
	if err := database.DB.Model(&models.Incident{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Count(&analytics.TotalIncidents).Error; err != nil {
		return nil, err
	}

	// Incidents by status
	var statusStats []struct {
		Status string
		Count  int64
	}
	if err := database.DB.Model(&models.Incident{}).
		Select("status, COUNT(*) as count").
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Group("status").
		Scan(&statusStats).Error; err != nil {
		return nil, err
	}

	analytics.IncidentsByStatus = make(map[string]int64)
	for _, stat := range statusStats {
		analytics.IncidentsByStatus[stat.Status] = stat.Count
	}

	// Incidents by impact
	var impactStats []struct {
		Impact string
		Count  int64
	}
	if err := database.DB.Model(&models.Incident{}).
		Select("impact, COUNT(*) as count").
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Group("impact").
		Scan(&impactStats).Error; err != nil {
		return nil, err
	}

	analytics.IncidentsByImpact = make(map[string]int64)
	for _, stat := range impactStats {
		analytics.IncidentsByImpact[stat.Impact] = stat.Count
	}

	// Average resolution time
	var avgResolution struct {
		AvgResolution float64
	}
	if err := database.DB.Model(&models.Incident{}).
		Select("AVG(EXTRACT(EPOCH FROM (resolved_at - created_at))/3600) as avg_resolution").
		Where("tenant_id = ? AND created_at >= ? AND resolved_at IS NOT NULL", tenantID, since).
		Scan(&avgResolution).Error; err != nil {
		return nil, err
	}

	analytics.AverageResolutionTime = avgResolution.AvgResolution

	// Incidents by service
	var serviceStats []struct {
		ServiceID uint
		Count     int64
	}
	if err := database.DB.Table("incident_services").
		Select("service_id, COUNT(*) as count").
		Joins("JOIN incidents ON incident_services.incident_id = incidents.id").
		Where("incidents.tenant_id = ? AND incidents.created_at >= ?", tenantID, since).
		Group("service_id").
		Scan(&serviceStats).Error; err != nil {
		return nil, err
	}

	analytics.IncidentsByService = make(map[uint]int64)
	for _, stat := range serviceStats {
		analytics.IncidentsByService[stat.ServiceID] = stat.Count
	}

	return analytics, nil
}

// GetMaintenanceAnalytics returns maintenance analytics for a tenant
func (s *AnalyticsService) GetMaintenanceAnalytics(tenantID uint, days int) (*MaintenanceAnalytics, error) {
	analytics := &MaintenanceAnalytics{
		TenantID: tenantID,
		Period:   days,
	}

	since := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	// Total maintenance events
	if err := database.DB.Model(&models.Maintenance{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Count(&analytics.TotalMaintenance).Error; err != nil {
		return nil, err
	}

	// Maintenance by status
	var statusStats []struct {
		Status string
		Count  int64
	}
	if err := database.DB.Model(&models.Maintenance{}).
		Select("status, COUNT(*) as count").
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Group("status").
		Scan(&statusStats).Error; err != nil {
		return nil, err
	}

	analytics.MaintenanceByStatus = make(map[string]int64)
	for _, stat := range statusStats {
		analytics.MaintenanceByStatus[stat.Status] = stat.Count
	}

	// Average duration
	var avgDuration struct {
		AvgDuration float64
	}
	if err := database.DB.Model(&models.Maintenance{}).
		Select("AVG(EXTRACT(EPOCH FROM (end_at - start_at))/3600) as avg_duration").
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Scan(&avgDuration).Error; err != nil {
		return nil, err
	}

	analytics.AverageDuration = avgDuration.AvgDuration

	return analytics, nil
}

// GetSubscriberAnalytics returns subscriber analytics for a tenant
func (s *AnalyticsService) GetSubscriberAnalytics(tenantID uint, days int) (*SubscriberAnalytics, error) {
	analytics := &SubscriberAnalytics{
		TenantID: tenantID,
		Period:   days,
	}

	// Total subscribers
	if err := database.DB.Model(&models.Subscriber{}).
		Where("tenant_id = ?", tenantID).
		Count(&analytics.TotalSubscribers).Error; err != nil {
		return nil, err
	}

	// Recent subscribers
	since := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	if err := database.DB.Model(&models.Subscriber{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Count(&analytics.RecentSubscribers).Error; err != nil {
		return nil, err
	}

	// SMS capable subscribers
	if err := database.DB.Model(&models.Subscriber{}).
		Where("tenant_id = ? AND phone != ''", tenantID).
		Count(&analytics.SMSSubscribers).Error; err != nil {
		return nil, err
	}

	// Subscribers by service
	var serviceStats []struct {
		ServiceID uint
		Count     int64
	}
	if err := database.DB.Table("subscriber_services").
		Select("service_id, COUNT(*) as count").
		Joins("JOIN subscribers ON subscriber_services.subscriber_id = subscribers.id").
		Where("subscribers.tenant_id = ?", tenantID).
		Group("service_id").
		Scan(&serviceStats).Error; err != nil {
		return nil, err
	}

	analytics.SubscribersByService = make(map[uint]int64)
	for _, stat := range serviceStats {
		analytics.SubscribersByService[stat.ServiceID] = stat.Count
	}

	return analytics, nil
}

// GetPerformanceMetrics returns performance metrics for a tenant
func (s *AnalyticsService) GetPerformanceMetrics(tenantID uint, days int) (*PerformanceMetrics, error) {
	metrics := &PerformanceMetrics{
		TenantID: tenantID,
		Period:   days,
	}

	since := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	// Get all monitors for the tenant
	var monitors []models.Monitor
	if err := database.DB.Where("tenant_id = ?", tenantID).Find(&monitors).Error; err != nil {
		return nil, err
	}

	var totalLatency int64
	var latencyCount int64
	var totalChecks int64
	var upChecks int64

	for _, monitor := range monitors {
		// Get heartbeats for this monitor
		var heartbeats []models.Heartbeat
		if err := database.DB.Where("monitor_id = ? AND timestamp >= ?", monitor.ID, since).
			Find(&heartbeats).Error; err != nil {
			continue
		}

		for _, heartbeat := range heartbeats {
			totalChecks++
			if heartbeat.Status == "up" {
				upChecks++
			}
			totalLatency += heartbeat.Latency
			latencyCount++
		}
	}

	// Calculate metrics
	if totalChecks > 0 {
		metrics.Uptime = float64(upChecks) / float64(totalChecks) * 100.0
	}

	if latencyCount > 0 {
		metrics.AverageLatency = float64(totalLatency) / float64(latencyCount)
	}

	// Get response time percentiles
	var latencies []int64
	if err := database.DB.Model(&models.Heartbeat{}).
		Joins("JOIN monitors ON heartbeats.monitor_id = monitors.id").
		Where("monitors.tenant_id = ? AND heartbeats.timestamp >= ?", tenantID, since).
		Pluck("heartbeats.latency", &latencies).Error; err != nil {
		return metrics, nil
	}

	if len(latencies) > 0 {
		// Sort latencies (simplified - in production, use proper sorting)
		metrics.P95Latency = calculatePercentile(latencies, 95)
		metrics.P99Latency = calculatePercentile(latencies, 99)
	}

	return metrics, nil
}

// GetDashboardSummary returns a summary for the dashboard
func (s *AnalyticsService) GetDashboardSummary(tenantID uint) (*DashboardSummary, error) {
	summary := &DashboardSummary{
		TenantID: tenantID,
	}

	// Get service count
	if err := database.DB.Model(&models.Service{}).
		Where("tenant_id = ?", tenantID).
		Count(&summary.TotalServices).Error; err != nil {
		return nil, err
	}

	// Get active incidents
	if err := database.DB.Model(&models.Incident{}).
		Where("tenant_id = ? AND status IN ?", tenantID, []string{"investigating", "identified", "monitoring"}).
		Count(&summary.ActiveIncidents).Error; err != nil {
		return nil, err
	}

	// Get upcoming maintenance
	tomorrow := time.Now().Add(24 * time.Hour)
	if err := database.DB.Model(&models.Maintenance{}).
		Where("tenant_id = ? AND start_at <= ? AND status = ?", tenantID, tomorrow, "scheduled").
		Count(&summary.UpcomingMaintenance).Error; err != nil {
		return nil, err
	}

	// Get total subscribers
	if err := database.DB.Model(&models.Subscriber{}).
		Where("tenant_id = ?", tenantID).
		Count(&summary.TotalSubscribers).Error; err != nil {
		return nil, err
	}

	// Get overall system status
	statusService := NewStatusService()
	overallStatus, err := statusService.GetOverallStatus(tenantID)
	if err == nil {
		summary.OverallStatus = overallStatus
	}

	return summary, nil
}

// RecordAPIUsage records API usage for analytics
func (s *AnalyticsService) RecordAPIUsage(tenantID uint, endpoint string, method string, responseTime float64, statusCode int, userAgent string, ipAddress string) error {
	// In a real implementation, you would store this in a database
	// For now, we'll just log it or store in memory
	return nil
}

// RecordPageView records page view for analytics
func (s *AnalyticsService) RecordPageView(tenantID uint, page string, userAgent string, ipAddress string, referrer string, sessionID string) error {
	// In a real implementation, you would store this in a database
	// For now, we'll just log it or store in memory
	return nil
}

// GetPlatformAPIUsageAnalytics returns platform-wide API usage analytics
func (s *AnalyticsService) GetPlatformAPIUsageAnalytics(days int) (map[string]interface{}, error) {
	// In a real implementation, you would query the database for platform-wide analytics
	// For now, return mock data
	return map[string]interface{}{
		"total_requests":    1000,
		"unique_users":      100,
		"avg_response_time": 150.5,
		"error_rate":        0.02,
	}, nil
}

// GetPlatformPageViewAnalytics returns platform-wide page view analytics
func (s *AnalyticsService) GetPlatformPageViewAnalytics(days int) (map[string]interface{}, error) {
	// In a real implementation, you would query the database for platform-wide analytics
	// For now, return mock data
	return map[string]interface{}{
		"total_page_views":     5000,
		"unique_visitors":      500,
		"avg_session_duration": 180.0,
		"bounce_rate":          0.3,
	}, nil
}

// Helper function to calculate percentile (simplified implementation)
func calculatePercentile(latencies []int64, percentile int) float64 {
	if len(latencies) == 0 {
		return 0
	}

	// Simple implementation - in production, use proper percentile calculation
	index := int(float64(len(latencies)) * float64(percentile) / 100.0)
	if index >= len(latencies) {
		index = len(latencies) - 1
	}

	return float64(latencies[index])
}

// UptimeReport represents uptime statistics
type UptimeReport struct {
	TenantID      uint                    `json:"tenant_id"`
	Period        int                     `json:"period_days"`
	OverallUptime float64                 `json:"overall_uptime"`
	Services      map[uint]*ServiceUptime `json:"services"`
}

// ServiceUptime represents uptime for a single service
type ServiceUptime struct {
	ServiceID   uint    `json:"service_id"`
	ServiceName string  `json:"service_name"`
	Uptime      float64 `json:"uptime"`
}

// IncidentAnalytics represents incident analytics
type IncidentAnalytics struct {
	TenantID              uint             `json:"tenant_id"`
	Period                int              `json:"period_days"`
	TotalIncidents        int64            `json:"total_incidents"`
	IncidentsByStatus     map[string]int64 `json:"incidents_by_status"`
	IncidentsByImpact     map[string]int64 `json:"incidents_by_impact"`
	IncidentsByService    map[uint]int64   `json:"incidents_by_service"`
	AverageResolutionTime float64          `json:"average_resolution_time_hours"`
}

// MaintenanceAnalytics represents maintenance analytics
type MaintenanceAnalytics struct {
	TenantID            uint             `json:"tenant_id"`
	Period              int              `json:"period_days"`
	TotalMaintenance    int64            `json:"total_maintenance"`
	MaintenanceByStatus map[string]int64 `json:"maintenance_by_status"`
	AverageDuration     float64          `json:"average_duration_hours"`
}

// SubscriberAnalytics represents subscriber analytics
type SubscriberAnalytics struct {
	TenantID             uint           `json:"tenant_id"`
	Period               int            `json:"period_days"`
	TotalSubscribers     int64          `json:"total_subscribers"`
	RecentSubscribers    int64          `json:"recent_subscribers"`
	SMSSubscribers       int64          `json:"sms_subscribers"`
	SubscribersByService map[uint]int64 `json:"subscribers_by_service"`
}

// PerformanceMetrics represents performance metrics
type PerformanceMetrics struct {
	TenantID       uint    `json:"tenant_id"`
	Period         int     `json:"period_days"`
	Uptime         float64 `json:"uptime_percentage"`
	AverageLatency float64 `json:"average_latency_ms"`
	P95Latency     float64 `json:"p95_latency_ms"`
	P99Latency     float64 `json:"p99_latency_ms"`
}

// DashboardSummary represents dashboard summary data
type DashboardSummary struct {
	TenantID            uint   `json:"tenant_id"`
	TotalServices       int64  `json:"total_services"`
	ActiveIncidents     int64  `json:"active_incidents"`
	UpcomingMaintenance int64  `json:"upcoming_maintenance"`
	TotalSubscribers    int64  `json:"total_subscribers"`
	OverallStatus       string `json:"overall_status"`
}
