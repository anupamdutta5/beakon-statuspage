// Package services provides business logic for the Analytics Service.
package services

import (
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage-analytics-service/internal/config"
	"github.com/enterprise-status/statuspage-analytics-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// AnalyticsService handles analytics-related business logic.
type AnalyticsService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAnalyticsService creates a new analytics service.
func NewAnalyticsService(db *gorm.DB, logger *zap.Logger) *AnalyticsService {
	return &AnalyticsService{
		db:     db,
		logger: logger,
	}
}

// CreateMetric creates a new metric.
func (s *AnalyticsService) CreateMetric(metric *models.Metric) error {
	// Set default values
	if metric.Type == "" {
		metric.Type = "gauge"
	}
	if metric.AggregationType == "" {
		metric.AggregationType = "sum"
	}
	if metric.RetentionDays == 0 {
		metric.RetentionDays = 365
	}
	if metric.IsActive == false && metric.IsActive != true {
		metric.IsActive = true
	}

	// Create metric
	if err := s.db.Create(metric).Error; err != nil {
		s.logger.Error("Failed to create metric", zap.Error(err))
		return fmt.Errorf("failed to create metric: %w", err)
	}

	s.logger.Info("Metric created successfully", zap.Uint("metric_id", metric.ID))
	return nil
}

// GetMetric retrieves a metric by ID.
func (s *AnalyticsService) GetMetric(id uint) (*models.Metric, error) {
	var metric models.Metric
	if err := s.db.First(&metric, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("metric not found")
		}
		s.logger.Error("Failed to get metric", zap.Error(err))
		return nil, fmt.Errorf("failed to get metric: %w", err)
	}

	return &metric, nil
}

// GetMetrics retrieves a list of metrics with pagination.
func (s *AnalyticsService) GetMetrics(tenantID uint, limit, offset int) ([]*models.Metric, int64, error) {
	var metrics []*models.Metric
	var total int64

	// Get total count
	if err := s.db.Model(&models.Metric{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count metrics", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count metrics: %w", err)
	}

	// Get metrics with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).Limit(limit).Offset(offset).Order("created_at DESC").Find(&metrics).Error; err != nil {
		s.logger.Error("Failed to get metrics", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get metrics: %w", err)
	}

	return metrics, total, nil
}

// UpdateMetric updates a metric.
func (s *AnalyticsService) UpdateMetric(metric *models.Metric) error {
	if err := s.db.Save(metric).Error; err != nil {
		s.logger.Error("Failed to update metric", zap.Error(err))
		return fmt.Errorf("failed to update metric: %w", err)
	}

	s.logger.Info("Metric updated successfully", zap.Uint("metric_id", metric.ID))
	return nil
}

// DeleteMetric soft deletes a metric.
func (s *AnalyticsService) DeleteMetric(id uint) error {
	if err := s.db.Delete(&models.Metric{}, id).Error; err != nil {
		s.logger.Error("Failed to delete metric", zap.Error(err))
		return fmt.Errorf("failed to delete metric: %w", err)
	}

	s.logger.Info("Metric deleted successfully", zap.Uint("metric_id", id))
	return nil
}

// AddMetricData adds a data point to a metric.
func (s *AnalyticsService) AddMetricData(dataPoint *models.MetricDataPoint) error {
	// Set default timestamp if not provided
	if dataPoint.Timestamp.IsZero() {
		dataPoint.Timestamp = time.Now()
	}

	// Create data point
	if err := s.db.Create(dataPoint).Error; err != nil {
		s.logger.Error("Failed to add metric data", zap.Error(err))
		return fmt.Errorf("failed to add metric data: %w", err)
	}

	s.logger.Info("Metric data added successfully", zap.Uint("data_point_id", dataPoint.ID))
	return nil
}

// GetMetricData retrieves data points for a metric.
func (s *AnalyticsService) GetMetricData(metricID uint, startDate, endDate time.Time, limit, offset int) ([]*models.MetricDataPoint, int64, error) {
	var dataPoints []*models.MetricDataPoint
	var total int64

	// Get total count
	query := s.db.Model(&models.MetricDataPoint{}).Where("metric_id = ?", metricID)
	if !startDate.IsZero() {
		query = query.Where("timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("timestamp <= ?", endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count metric data", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count metric data: %w", err)
	}

	// Get data points with pagination
	query = s.db.Where("metric_id = ?", metricID)
	if !startDate.IsZero() {
		query = query.Where("timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("timestamp <= ?", endDate)
	}

	if err := query.Preload("Metric").Limit(limit).Offset(offset).Order("timestamp DESC").Find(&dataPoints).Error; err != nil {
		s.logger.Error("Failed to get metric data", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get metric data: %w", err)
	}

	return dataPoints, total, nil
}

// GetAnalyticsOverview returns an overview of analytics data.
func (s *AnalyticsService) GetAnalyticsOverview(tenantID uint) (map[string]interface{}, error) {
	var overview map[string]interface{} = make(map[string]interface{})

	// Get total metrics count
	var totalMetrics int64
	if err := s.db.Model(&models.Metric{}).Where("tenant_id = ?", tenantID).Count(&totalMetrics).Error; err != nil {
		s.logger.Error("Failed to count metrics", zap.Error(err))
		return nil, fmt.Errorf("failed to count metrics: %w", err)
	}

	// Get total data points count
	var totalDataPoints int64
	if err := s.db.Model(&models.MetricDataPoint{}).Joins("JOIN metrics ON metric_data_points.metric_id = metrics.id").Where("metrics.tenant_id = ?", tenantID).Count(&totalDataPoints).Error; err != nil {
		s.logger.Error("Failed to count data points", zap.Error(err))
		return nil, fmt.Errorf("failed to count data points: %w", err)
	}

	// Get total reports count
	var totalReports int64
	if err := s.db.Model(&models.Report{}).Where("tenant_id = ?", tenantID).Count(&totalReports).Error; err != nil {
		s.logger.Error("Failed to count reports", zap.Error(err))
		return nil, fmt.Errorf("failed to count reports: %w", err)
	}

	// Get total dashboards count
	var totalDashboards int64
	if err := s.db.Model(&models.Dashboard{}).Where("tenant_id = ?", tenantID).Count(&totalDashboards).Error; err != nil {
		s.logger.Error("Failed to count dashboards", zap.Error(err))
		return nil, fmt.Errorf("failed to count dashboards: %w", err)
	}

	// Get recent data points (last 24 hours)
	var recentDataPoints int64
	last24Hours := time.Now().Add(-24 * time.Hour)
	if err := s.db.Model(&models.MetricDataPoint{}).Joins("JOIN metrics ON metric_data_points.metric_id = metrics.id").Where("metrics.tenant_id = ? AND metric_data_points.timestamp >= ?", tenantID, last24Hours).Count(&recentDataPoints).Error; err != nil {
		s.logger.Error("Failed to count recent data points", zap.Error(err))
		return nil, fmt.Errorf("failed to count recent data points: %w", err)
	}

	overview["total_metrics"] = totalMetrics
	overview["total_data_points"] = totalDataPoints
	overview["total_reports"] = totalReports
	overview["total_dashboards"] = totalDashboards
	overview["recent_data_points_24h"] = recentDataPoints
	overview["last_updated"] = time.Now().UTC()

	return overview, nil
}

// Report Management

// CreateReport creates a new report.
func (s *AnalyticsService) CreateReport(report *models.Report) error {
	// Set default values
	if report.Type == "" {
		report.Type = "on_demand"
	}
	if report.Format == "" {
		report.Format = "pdf"
	}
	if report.Status == "" {
		report.Status = "draft"
	}

	// Create report
	if err := s.db.Create(report).Error; err != nil {
		s.logger.Error("Failed to create report", zap.Error(err))
		return fmt.Errorf("failed to create report: %w", err)
	}

	s.logger.Info("Report created successfully", zap.Uint("report_id", report.ID))
	return nil
}

// GetReport retrieves a report by ID.
func (s *AnalyticsService) GetReport(id uint) (*models.Report, error) {
	var report models.Report
	if err := s.db.Preload("Generations").First(&report, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("report not found")
		}
		s.logger.Error("Failed to get report", zap.Error(err))
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	return &report, nil
}

// GetReports retrieves a list of reports with pagination.
func (s *AnalyticsService) GetReports(tenantID uint, limit, offset int) ([]*models.Report, int64, error) {
	var reports []*models.Report
	var total int64

	// Get total count
	if err := s.db.Model(&models.Report{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count reports", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count reports: %w", err)
	}

	// Get reports with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).Preload("Generations").Limit(limit).Offset(offset).Order("created_at DESC").Find(&reports).Error; err != nil {
		s.logger.Error("Failed to get reports", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get reports: %w", err)
	}

	return reports, total, nil
}

// UpdateReport updates a report.
func (s *AnalyticsService) UpdateReport(report *models.Report) error {
	if err := s.db.Save(report).Error; err != nil {
		s.logger.Error("Failed to update report", zap.Error(err))
		return fmt.Errorf("failed to update report: %w", err)
	}

	s.logger.Info("Report updated successfully", zap.Uint("report_id", report.ID))
	return nil
}

// DeleteReport soft deletes a report.
func (s *AnalyticsService) DeleteReport(id uint) error {
	if err := s.db.Delete(&models.Report{}, id).Error; err != nil {
		s.logger.Error("Failed to delete report", zap.Error(err))
		return fmt.Errorf("failed to delete report: %w", err)
	}

	s.logger.Info("Report deleted successfully", zap.Uint("report_id", id))
	return nil
}

// GenerateReport generates a report.
func (s *AnalyticsService) GenerateReport(reportID uint) (*models.ReportGeneration, error) {
	// Get report
	var report models.Report
	if err := s.db.First(&report, reportID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("report not found")
		}
		s.logger.Error("Failed to get report", zap.Error(err))
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	// Create report generation
	generation := &models.ReportGeneration{
		ReportID:  reportID,
		Status:    "generating",
		StartedAt: time.Now(),
	}

	if err := s.db.Create(generation).Error; err != nil {
		s.logger.Error("Failed to create report generation", zap.Error(err))
		return nil, fmt.Errorf("failed to create report generation: %w", err)
	}

	// Update report's last generated time
	now := time.Now()
	report.LastGenerated = &now
	if err := s.db.Save(&report).Error; err != nil {
		s.logger.Error("Failed to update report last generated time", zap.Error(err))
		// Don't fail the generation if this fails
	}

	s.logger.Info("Report generation started", zap.Uint("generation_id", generation.ID))
	return generation, nil
}

// Dashboard Management

// CreateDashboard creates a new dashboard.
func (s *AnalyticsService) CreateDashboard(dashboard *models.Dashboard) error {
	// Set default values
	if dashboard.IsPublic == false && dashboard.IsPublic != true {
		dashboard.IsPublic = false
	}
	if dashboard.IsDefault == false && dashboard.IsDefault != true {
		dashboard.IsDefault = false
	}

	// Create dashboard
	if err := s.db.Create(dashboard).Error; err != nil {
		s.logger.Error("Failed to create dashboard", zap.Error(err))
		return fmt.Errorf("failed to create dashboard: %w", err)
	}

	s.logger.Info("Dashboard created successfully", zap.Uint("dashboard_id", dashboard.ID))
	return nil
}

// GetDashboard retrieves a dashboard by ID.
func (s *AnalyticsService) GetDashboard(id uint) (*models.Dashboard, error) {
	var dashboard models.Dashboard
	if err := s.db.Preload("Widgets").First(&dashboard, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("dashboard not found")
		}
		s.logger.Error("Failed to get dashboard", zap.Error(err))
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}

	return &dashboard, nil
}

// GetDashboards retrieves a list of dashboards with pagination.
func (s *AnalyticsService) GetDashboards(tenantID uint, limit, offset int) ([]*models.Dashboard, int64, error) {
	var dashboards []*models.Dashboard
	var total int64

	// Get total count
	if err := s.db.Model(&models.Dashboard{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count dashboards", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count dashboards: %w", err)
	}

	// Get dashboards with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).Preload("Widgets").Limit(limit).Offset(offset).Order("created_at DESC").Find(&dashboards).Error; err != nil {
		s.logger.Error("Failed to get dashboards", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get dashboards: %w", err)
	}

	return dashboards, total, nil
}

// UpdateDashboard updates a dashboard.
func (s *AnalyticsService) UpdateDashboard(dashboard *models.Dashboard) error {
	if err := s.db.Save(dashboard).Error; err != nil {
		s.logger.Error("Failed to update dashboard", zap.Error(err))
		return fmt.Errorf("failed to update dashboard: %w", err)
	}

	s.logger.Info("Dashboard updated successfully", zap.Uint("dashboard_id", dashboard.ID))
	return nil
}

// DeleteDashboard soft deletes a dashboard.
func (s *AnalyticsService) DeleteDashboard(id uint) error {
	if err := s.db.Delete(&models.Dashboard{}, id).Error; err != nil {
		s.logger.Error("Failed to delete dashboard", zap.Error(err))
		return fmt.Errorf("failed to delete dashboard: %w", err)
	}

	s.logger.Info("Dashboard deleted successfully", zap.Uint("dashboard_id", id))
	return nil
}

// Widget Management

// AddDashboardWidget adds a widget to a dashboard.
func (s *AnalyticsService) AddDashboardWidget(widget *models.DashboardWidget) error {
	// Set default values
	if widget.Type == "" {
		widget.Type = "metric"
	}
	if widget.Size == "" {
		widget.Size = "medium"
	}

	// Create widget
	if err := s.db.Create(widget).Error; err != nil {
		s.logger.Error("Failed to add dashboard widget", zap.Error(err))
		return fmt.Errorf("failed to add dashboard widget: %w", err)
	}

	s.logger.Info("Dashboard widget added successfully", zap.Uint("widget_id", widget.ID))
	return nil
}

// UpdateDashboardWidget updates a dashboard widget.
func (s *AnalyticsService) UpdateDashboardWidget(widget *models.DashboardWidget) error {
	if err := s.db.Save(widget).Error; err != nil {
		s.logger.Error("Failed to update dashboard widget", zap.Error(err))
		return fmt.Errorf("failed to update dashboard widget: %w", err)
	}

	s.logger.Info("Dashboard widget updated successfully", zap.Uint("widget_id", widget.ID))
	return nil
}

// DeleteDashboardWidget deletes a dashboard widget.
func (s *AnalyticsService) DeleteDashboardWidget(id uint) error {
	if err := s.db.Delete(&models.DashboardWidget{}, id).Error; err != nil {
		s.logger.Error("Failed to delete dashboard widget", zap.Error(err))
		return fmt.Errorf("failed to delete dashboard widget: %w", err)
	}

	s.logger.Info("Dashboard widget deleted successfully", zap.Uint("widget_id", id))
	return nil
}

// GetDashboardWidgets retrieves widgets for a dashboard.
func (s *AnalyticsService) GetDashboardWidgets(dashboardID uint) ([]*models.DashboardWidget, error) {
	var widgets []*models.DashboardWidget
	if err := s.db.Where("dashboard_id = ?", dashboardID).Order("position ASC").Find(&widgets).Error; err != nil {
		s.logger.Error("Failed to get dashboard widgets", zap.Error(err))
		return nil, fmt.Errorf("failed to get dashboard widgets: %w", err)
	}

	return widgets, nil
}

// GetPublicMetrics returns public metrics for a tenant.
func (s *AnalyticsService) GetPublicMetrics(tenantID uint) ([]*models.Metric, error) {
	var metrics []*models.Metric
	if err := s.db.Where("tenant_id = ? AND is_active = ? AND is_public = ?", tenantID, true, true).Order("created_at ASC").Find(&metrics).Error; err != nil {
		s.logger.Error("Failed to get public metrics", zap.Error(err))
		return nil, fmt.Errorf("failed to get public metrics: %w", err)
	}

	return metrics, nil
}

// GetPublicReports returns public reports for a tenant.
func (s *AnalyticsService) GetPublicReports(tenantID uint) ([]*models.Report, error) {
	var reports []*models.Report
	if err := s.db.Where("tenant_id = ? AND is_public = ?", tenantID, true).Preload("Generations").Order("created_at DESC").Find(&reports).Error; err != nil {
		s.logger.Error("Failed to get public reports", zap.Error(err))
		return nil, fmt.Errorf("failed to get public reports: %w", err)
	}

	return reports, nil
}

// InitDatabase initializes the database connection and runs migrations.
func InitDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.Metric{},
		&models.MetricDataPoint{},
		&models.Report{},
		&models.ReportGeneration{},
		&models.Dashboard{},
		&models.DashboardWidget{},
		&models.AnalyticsEvent{},
		&models.DataExport{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
