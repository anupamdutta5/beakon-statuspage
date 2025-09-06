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

type AnalyticsService struct {
	db *gorm.DB
}

func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{
		db: database.DB,
	}
}

// Analytics Models

type PageView struct {
	ID        uint   `gorm:"primaryKey"`
	TenantID  uint   `gorm:"not null"`
	Page      string `gorm:"not null"`
	UserAgent string
	IPAddress string
	Referrer  string
	SessionID string
	CreatedAt time.Time
}

type UptimeReport struct {
	ID          uint      `gorm:"primaryKey"`
	TenantID    uint      `gorm:"not null"`
	ServiceID   uint      `gorm:"not null"`
	Period      string    `gorm:"not null"` // daily, weekly, monthly
	Uptime      float64   `gorm:"not null"`
	Downtime    float64   `gorm:"not null"`
	TotalChecks int64     `gorm:"not null"`
	StartDate   time.Time `gorm:"not null"`
	EndDate     time.Time `gorm:"not null"`
	CreatedAt   time.Time
}

type IncidentAnalytics struct {
	ID                uint   `gorm:"primaryKey"`
	TenantID          uint   `gorm:"not null"`
	IncidentID        uint   `gorm:"not null"`
	Severity          string `gorm:"not null"`
	Duration          int64  `gorm:"not null"` // in minutes
	AffectedUsers     int64  `gorm:"default:0"`
	NotificationsSent int64  `gorm:"default:0"`
	ResolvedAt        time.Time
	CreatedAt         time.Time
}

type PerformanceMetrics struct {
	ID           uint      `gorm:"primaryKey"`
	TenantID     uint      `gorm:"not null"`
	ServiceID    uint      `gorm:"not null"`
	ResponseTime float64   `gorm:"not null"` // in milliseconds
	Throughput   float64   `gorm:"not null"` // requests per second
	ErrorRate    float64   `gorm:"not null"` // percentage
	Availability float64   `gorm:"not null"` // percentage
	Timestamp    time.Time `gorm:"not null"`
	CreatedAt    time.Time
}

// Page Analytics

func (s *AnalyticsService) RecordPageView(tenantID uint, page, userAgent, ipAddress, referrer, sessionID string) error {
	pageView := &PageView{
		TenantID:  tenantID,
		Page:      page,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		Referrer:  referrer,
		SessionID: sessionID,
	}

	return s.db.Create(pageView).Error
}

func (s *AnalyticsService) GetPageAnalytics(tenantID uint, days int) (map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	analytics := make(map[string]interface{})

	// Total page views
	var totalViews int64
	if err := s.db.Model(&PageView{}).Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).Count(&totalViews).Error; err != nil {
		return nil, err
	}
	analytics["total_views"] = totalViews

	// Unique visitors (based on IP address)
	var uniqueVisitors int64
	if err := s.db.Model(&PageView{}).Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Distinct("ip_address").Count(&uniqueVisitors).Error; err != nil {
		return nil, err
	}
	analytics["unique_visitors"] = uniqueVisitors

	// Top pages
	var topPages []struct {
		Page  string `json:"page"`
		Views int64  `json:"views"`
	}
	if err := s.db.Model(&PageView{}).
		Select("page, COUNT(*) as views").
		Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Group("page").
		Order("views DESC").
		Limit(10).
		Scan(&topPages).Error; err != nil {
		return nil, err
	}
	analytics["top_pages"] = topPages

	// Daily page views
	var dailyViews []struct {
		Date  string `json:"date"`
		Views int64  `json:"views"`
	}
	if err := s.db.Model(&PageView{}).
		Select("DATE(created_at) as date, COUNT(*) as views").
		Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&dailyViews).Error; err != nil {
		return nil, err
	}
	analytics["daily_views"] = dailyViews

	return analytics, nil
}

// Uptime Analytics

func (s *AnalyticsService) GenerateUptimeReport(tenantID uint, serviceID uint, period string) (*UptimeReport, error) {
	var startDate, endDate time.Time
	now := time.Now()

	switch period {
	case "daily":
		startDate = now.AddDate(0, 0, -1)
		endDate = now
	case "weekly":
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case "monthly":
		startDate = now.AddDate(0, -1, 0)
		endDate = now
	default:
		return nil, fmt.Errorf("invalid period: %s", period)
	}

	// Get health checks for the period
	var healthChecks []models.HealthCheck
	if err := s.db.Where("service_id = ? AND checked_at BETWEEN ? AND ?",
		serviceID, startDate, endDate).Find(&healthChecks).Error; err != nil {
		return nil, err
	}

	if len(healthChecks) == 0 {
		return &UptimeReport{
			TenantID:    tenantID,
			ServiceID:   serviceID,
			Period:      period,
			Uptime:      100.0,
			Downtime:    0.0,
			TotalChecks: 0,
			StartDate:   startDate,
			EndDate:     endDate,
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

	report := &UptimeReport{
		TenantID:    tenantID,
		ServiceID:   serviceID,
		Period:      period,
		Uptime:      uptime,
		Downtime:    downtime,
		TotalChecks: int64(totalChecks),
		StartDate:   startDate,
		EndDate:     endDate,
	}

	// Save report
	if err := s.db.Create(report).Error; err != nil {
		logger.Error("Failed to create uptime report", zap.Error(err))
		return nil, err
	}

	return report, nil
}

func (s *AnalyticsService) GetUptimeAnalytics(tenantID uint, days int) (map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	analytics := make(map[string]interface{})

	// Overall uptime
	var avgUptime float64
	if err := s.db.Model(&UptimeReport{}).
		Where("tenant_id = ? AND start_date >= ?", tenantID, startDate).
		Select("AVG(uptime)").Scan(&avgUptime).Error; err != nil {
		return nil, err
	}
	analytics["overall_uptime"] = avgUptime

	// Service uptime breakdown
	var serviceUptime []struct {
		ServiceID   uint    `json:"service_id"`
		ServiceName string  `json:"service_name"`
		Uptime      float64 `json:"uptime"`
	}
	if err := s.db.Table("uptime_reports").
		Select("service_id, services.name as service_name, AVG(uptime_reports.uptime) as uptime").
		Joins("JOIN services ON uptime_reports.service_id = services.id").
		Where("uptime_reports.tenant_id = ? AND uptime_reports.start_date >= ?", tenantID, startDate).
		Group("service_id, services.name").
		Scan(&serviceUptime).Error; err != nil {
		return nil, err
	}
	analytics["service_uptime"] = serviceUptime

	// Daily uptime trend
	var dailyUptime []struct {
		Date   string  `json:"date"`
		Uptime float64 `json:"uptime"`
	}
	if err := s.db.Model(&UptimeReport{}).
		Select("DATE(start_date) as date, AVG(uptime) as uptime").
		Where("tenant_id = ? AND start_date >= ?", tenantID, startDate).
		Group("DATE(start_date)").
		Order("date ASC").
		Scan(&dailyUptime).Error; err != nil {
		return nil, err
	}
	analytics["daily_uptime"] = dailyUptime

	return analytics, nil
}

// Incident Analytics

func (s *AnalyticsService) RecordIncidentAnalytics(tenantID uint, incident *models.Incident) error {
	// Calculate incident duration
	var duration int64
	if incident.ResolvedAt != nil {
		duration = int64(incident.ResolvedAt.Sub(incident.CreatedAt).Minutes())
	}

	// Count affected subscribers
	var affectedUsers int64
	if err := s.db.Model(&models.Subscriber{}).Where("tenant_id = ? AND is_active = ?", tenantID, true).Count(&affectedUsers).Error; err != nil {
		affectedUsers = 0
	}

	// Count notifications sent (this would be tracked separately)
	notificationsSent := affectedUsers // Simplified for now

	analytics := &IncidentAnalytics{
		TenantID:          tenantID,
		IncidentID:        incident.ID,
		Severity:          incident.Impact,
		Duration:          duration,
		AffectedUsers:     affectedUsers,
		NotificationsSent: notificationsSent,
		ResolvedAt:        time.Now(),
	}

	return s.db.Create(analytics).Error
}

func (s *AnalyticsService) GetIncidentAnalytics(tenantID uint, days int) (map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	analytics := make(map[string]interface{})

	// Total incidents
	var totalIncidents int64
	if err := s.db.Model(&IncidentAnalytics{}).Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).Count(&totalIncidents).Error; err != nil {
		return nil, err
	}
	analytics["total_incidents"] = totalIncidents

	// Average incident duration
	var avgDuration float64
	if err := s.db.Model(&IncidentAnalytics{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Select("AVG(duration)").Scan(&avgDuration).Error; err != nil {
		return nil, err
	}
	analytics["avg_duration"] = avgDuration

	// Incidents by severity
	var incidentsBySeverity []struct {
		Severity string `json:"severity"`
		Count    int64  `json:"count"`
	}
	if err := s.db.Model(&IncidentAnalytics{}).
		Select("severity, COUNT(*) as count").
		Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Group("severity").
		Scan(&incidentsBySeverity).Error; err != nil {
		return nil, err
	}
	analytics["incidents_by_severity"] = incidentsBySeverity

	// Monthly incident trend
	var monthlyIncidents []struct {
		Month string `json:"month"`
		Count int64  `json:"count"`
	}
	if err := s.db.Model(&IncidentAnalytics{}).
		Select("DATE_FORMAT(created_at, '%Y-%m') as month, COUNT(*) as count").
		Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Group("DATE_FORMAT(created_at, '%Y-%m')").
		Order("month ASC").
		Scan(&monthlyIncidents).Error; err != nil {
		return nil, err
	}
	analytics["monthly_incidents"] = monthlyIncidents

	return analytics, nil
}

// Performance Analytics

func (s *AnalyticsService) RecordPerformanceMetrics(tenantID uint, serviceID uint, responseTime, throughput, errorRate, availability float64) error {
	metrics := &PerformanceMetrics{
		TenantID:     tenantID,
		ServiceID:    serviceID,
		ResponseTime: responseTime,
		Throughput:   throughput,
		ErrorRate:    errorRate,
		Availability: availability,
		Timestamp:    time.Now(),
	}

	return s.db.Create(metrics).Error
}

func (s *AnalyticsService) GetPerformanceAnalytics(tenantID uint, days int) (map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	analytics := make(map[string]interface{})

	// Average response time
	var avgResponseTime float64
	if err := s.db.Model(&PerformanceMetrics{}).
		Where("tenant_id = ? AND timestamp >= ?", tenantID, startDate).
		Select("AVG(response_time)").Scan(&avgResponseTime).Error; err != nil {
		return nil, err
	}
	analytics["avg_response_time"] = avgResponseTime

	// Average throughput
	var avgThroughput float64
	if err := s.db.Model(&PerformanceMetrics{}).
		Where("tenant_id = ? AND timestamp >= ?", tenantID, startDate).
		Select("AVG(throughput)").Scan(&avgThroughput).Error; err != nil {
		return nil, err
	}
	analytics["avg_throughput"] = avgThroughput

	// Average error rate
	var avgErrorRate float64
	if err := s.db.Model(&PerformanceMetrics{}).
		Where("tenant_id = ? AND timestamp >= ?", tenantID, startDate).
		Select("AVG(error_rate)").Scan(&avgErrorRate).Error; err != nil {
		return nil, err
	}
	analytics["avg_error_rate"] = avgErrorRate

	// Average availability
	var avgAvailability float64
	if err := s.db.Model(&PerformanceMetrics{}).
		Where("tenant_id = ? AND timestamp >= ?", tenantID, startDate).
		Select("AVG(availability)").Scan(&avgAvailability).Error; err != nil {
		return nil, err
	}
	analytics["avg_availability"] = avgAvailability

	// Performance trends
	var performanceTrends []struct {
		Date         string  `json:"date"`
		ResponseTime float64 `json:"response_time"`
		Throughput   float64 `json:"throughput"`
		ErrorRate    float64 `json:"error_rate"`
		Availability float64 `json:"availability"`
	}
	if err := s.db.Model(&PerformanceMetrics{}).
		Select("DATE(timestamp) as date, AVG(response_time) as response_time, AVG(throughput) as throughput, AVG(error_rate) as error_rate, AVG(availability) as availability").
		Where("tenant_id = ? AND timestamp >= ?", tenantID, startDate).
		Group("DATE(timestamp)").
		Order("date ASC").
		Scan(&performanceTrends).Error; err != nil {
		return nil, err
	}
	analytics["performance_trends"] = performanceTrends

	return analytics, nil
}

// Comprehensive Analytics Dashboard

func (s *AnalyticsService) GetDashboardAnalytics(tenantID uint, days int) (map[string]interface{}, error) {
	dashboard := make(map[string]interface{})

	// Page analytics
	pageAnalytics, err := s.GetPageAnalytics(tenantID, days)
	if err != nil {
		logger.Error("Failed to get page analytics", zap.Error(err))
	} else {
		dashboard["page_analytics"] = pageAnalytics
	}

	// Uptime analytics
	uptimeAnalytics, err := s.GetUptimeAnalytics(tenantID, days)
	if err != nil {
		logger.Error("Failed to get uptime analytics", zap.Error(err))
	} else {
		dashboard["uptime_analytics"] = uptimeAnalytics
	}

	// Incident analytics
	incidentAnalytics, err := s.GetIncidentAnalytics(tenantID, days)
	if err != nil {
		logger.Error("Failed to get incident analytics", zap.Error(err))
	} else {
		dashboard["incident_analytics"] = incidentAnalytics
	}

	// Performance analytics
	performanceAnalytics, err := s.GetPerformanceAnalytics(tenantID, days)
	if err != nil {
		logger.Error("Failed to get performance analytics", zap.Error(err))
	} else {
		dashboard["performance_analytics"] = performanceAnalytics
	}

	// Summary metrics
	summary := s.calculateSummaryMetrics(dashboard)
	dashboard["summary"] = summary

	return dashboard, nil
}

func (s *AnalyticsService) calculateSummaryMetrics(dashboard map[string]interface{}) map[string]interface{} {
	summary := make(map[string]interface{})

	// Extract key metrics
	if pageAnalytics, ok := dashboard["page_analytics"].(map[string]interface{}); ok {
		if totalViews, ok := pageAnalytics["total_views"].(int64); ok {
			summary["total_page_views"] = totalViews
		}
		if uniqueVisitors, ok := pageAnalytics["unique_visitors"].(int64); ok {
			summary["unique_visitors"] = uniqueVisitors
		}
	}

	if uptimeAnalytics, ok := dashboard["uptime_analytics"].(map[string]interface{}); ok {
		if overallUptime, ok := uptimeAnalytics["overall_uptime"].(float64); ok {
			summary["overall_uptime"] = overallUptime
		}
	}

	if incidentAnalytics, ok := dashboard["incident_analytics"].(map[string]interface{}); ok {
		if totalIncidents, ok := incidentAnalytics["total_incidents"].(int64); ok {
			summary["total_incidents"] = totalIncidents
		}
		if avgDuration, ok := incidentAnalytics["avg_duration"].(float64); ok {
			summary["avg_incident_duration"] = avgDuration
		}
	}

	if performanceAnalytics, ok := dashboard["performance_analytics"].(map[string]interface{}); ok {
		if avgResponseTime, ok := performanceAnalytics["avg_response_time"].(float64); ok {
			summary["avg_response_time"] = avgResponseTime
		}
		if avgErrorRate, ok := performanceAnalytics["avg_error_rate"].(float64); ok {
			summary["avg_error_rate"] = avgErrorRate
		}
	}

	return summary
}
