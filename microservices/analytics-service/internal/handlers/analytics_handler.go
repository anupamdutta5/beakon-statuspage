// Package handlers provides HTTP handlers for the Analytics Service.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/anupamdutta5/statuspage-analytics-service/internal/models"
	"github.com/anupamdutta5/statuspage-analytics-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AnalyticsHandler handles analytics-related HTTP requests.
type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
	logger           *zap.Logger
}

// NewAnalyticsHandler creates a new analytics handler.
func NewAnalyticsHandler(analyticsService *services.AnalyticsService, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		logger:           logger,
	}
}

// Health returns the health status of the Analytics Service.
func (h *AnalyticsHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "analytics-service",
		"version":   "1.0.0",
		"timestamp": "2024-01-01T00:00:00Z",
	})
}

// GetPublicMetrics returns public metrics for a tenant.
func (h *AnalyticsHandler) GetPublicMetrics(c *gin.Context) {
	// For now, use tenant ID 1 as default
	// In production, this would be determined from the request context
	tenantID := uint(1)

	metrics, err := h.analyticsService.GetPublicMetrics(tenantID)
	if err != nil {
		h.logger.Error("Failed to get public metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"metrics": metrics})
}

// GetPublicReports returns public reports for a tenant.
func (h *AnalyticsHandler) GetPublicReports(c *gin.Context) {
	// For now, use tenant ID 1 as default
	// In production, this would be determined from the request context
	tenantID := uint(1)

	reports, err := h.analyticsService.GetPublicReports(tenantID)
	if err != nil {
		h.logger.Error("Failed to get public reports", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reports"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reports": reports})
}

// GetAnalyticsOverview returns an overview of analytics data.
func (h *AnalyticsHandler) GetAnalyticsOverview(c *gin.Context) {
	// Get tenant ID from context (set by auth middleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	overview, err := h.analyticsService.GetAnalyticsOverview(tenantID.(uint))
	if err != nil {
		h.logger.Error("Failed to get analytics overview", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get analytics overview"})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// Metric Management Handlers

// GetMetrics handles getting a list of metrics.
func (h *AnalyticsHandler) GetMetrics(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	metrics, total, err := h.analyticsService.GetMetrics(tenantID.(uint), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetMetric handles getting a specific metric.
func (h *AnalyticsHandler) GetMetric(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metric ID"})
		return
	}

	metric, err := h.analyticsService.GetMetric(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Metric not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"metric": metric})
}

// CreateMetric handles creating a new metric.
func (h *AnalyticsHandler) CreateMetric(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		Name            string `json:"name" binding:"required"`
		Description     string `json:"description"`
		Type            string `json:"type"`
		Unit            string `json:"unit"`
		Category        string `json:"category"`
		IsActive        *bool  `json:"is_active"`
		IsPublic        *bool  `json:"is_public"`
		AggregationType string `json:"aggregation_type"`
		RetentionDays   int    `json:"retention_days"`
		Metadata        string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create metric request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	metric := &models.Metric{
		TenantID:        tenantID.(uint),
		Name:            req.Name,
		Description:     req.Description,
		Type:            req.Type,
		Unit:            req.Unit,
		Category:        req.Category,
		AggregationType: req.AggregationType,
		RetentionDays:   req.RetentionDays,
		Metadata:        req.Metadata,
	}

	if req.IsActive != nil {
		metric.IsActive = *req.IsActive
	}
	if req.IsPublic != nil {
		metric.IsPublic = *req.IsPublic
	}

	if err := h.analyticsService.CreateMetric(metric); err != nil {
		h.logger.Error("Failed to create metric", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create metric"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Metric created successfully",
		"metric":  metric,
	})
}

// UpdateMetric handles updating a metric.
func (h *AnalyticsHandler) UpdateMetric(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metric ID"})
		return
	}

	metric, err := h.analyticsService.GetMetric(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Metric not found"})
		return
	}

	var req struct {
		Name            string `json:"name"`
		Description     string `json:"description"`
		Type            string `json:"type"`
		Unit            string `json:"unit"`
		Category        string `json:"category"`
		IsActive        *bool  `json:"is_active"`
		IsPublic        *bool  `json:"is_public"`
		AggregationType string `json:"aggregation_type"`
		RetentionDays   *int   `json:"retention_days"`
		Metadata        string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update metric request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update fields
	if req.Name != "" {
		metric.Name = req.Name
	}
	if req.Description != "" {
		metric.Description = req.Description
	}
	if req.Type != "" {
		metric.Type = req.Type
	}
	if req.Unit != "" {
		metric.Unit = req.Unit
	}
	if req.Category != "" {
		metric.Category = req.Category
	}
	if req.IsActive != nil {
		metric.IsActive = *req.IsActive
	}
	if req.IsPublic != nil {
		metric.IsPublic = *req.IsPublic
	}
	if req.AggregationType != "" {
		metric.AggregationType = req.AggregationType
	}
	if req.RetentionDays != nil {
		metric.RetentionDays = *req.RetentionDays
	}
	if req.Metadata != "" {
		metric.Metadata = req.Metadata
	}

	if err := h.analyticsService.UpdateMetric(metric); err != nil {
		h.logger.Error("Failed to update metric", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update metric"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Metric updated successfully",
		"metric":  metric,
	})
}

// DeleteMetric handles deleting a metric.
func (h *AnalyticsHandler) DeleteMetric(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metric ID"})
		return
	}

	if err := h.analyticsService.DeleteMetric(uint(id)); err != nil {
		h.logger.Error("Failed to delete metric", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete metric"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Metric deleted successfully"})
}

// GetMetricData handles getting data points for a metric.
func (h *AnalyticsHandler) GetMetricData(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metric ID"})
		return
	}

	// Parse date range parameters
	var startDate, endDate time.Time
	if startStr := c.Query("start_date"); startStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startStr); err == nil {
			startDate = parsed
		}
	}
	if endStr := c.Query("end_date"); endStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endStr); err == nil {
			endDate = parsed
		}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 1000 {
		limit = 1000
	}

	dataPoints, total, err := h.analyticsService.GetMetricData(uint(id), startDate, endDate, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get metric data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metric data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data_points": dataPoints,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

// AddMetricData handles adding a data point to a metric.
func (h *AnalyticsHandler) AddMetricData(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metric ID"})
		return
	}

	var req struct {
		Value     float64   `json:"value" binding:"required"`
		Timestamp time.Time `json:"timestamp"`
		Labels    string    `json:"labels"`
		Source    string    `json:"source"`
		Metadata  string    `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid add metric data request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	dataPoint := &models.MetricDataPoint{
		MetricID:  uint(id),
		Value:     req.Value,
		Timestamp: req.Timestamp,
		Labels:    req.Labels,
		Source:    req.Source,
		Metadata:  req.Metadata,
	}

	if err := h.analyticsService.AddMetricData(dataPoint); err != nil {
		h.logger.Error("Failed to add metric data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add metric data"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Metric data added successfully",
		"data_point": dataPoint,
	})
}

// Report Management Handlers

// GetReports handles getting a list of reports.
func (h *AnalyticsHandler) GetReports(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	reports, total, err := h.analyticsService.GetReports(tenantID.(uint), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get reports", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reports"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reports": reports,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetReport handles getting a specific report.
func (h *AnalyticsHandler) GetReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	report, err := h.analyticsService.GetReport(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"report": report})
}

// CreateReport handles creating a new report.
func (h *AnalyticsHandler) CreateReport(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Status      string `json:"status"`
		Schedule    string `json:"schedule"`
		Format      string `json:"format"`
		IsPublic    *bool  `json:"is_public"`
		Config      string `json:"config"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create report request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	report := &models.Report{
		TenantID:    tenantID.(uint),
		UserID:      userID.(uint),
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Status:      req.Status,
		Schedule:    req.Schedule,
		Format:      req.Format,
		Config:      req.Config,
		Metadata:    req.Metadata,
	}

	if req.IsPublic != nil {
		report.IsPublic = *req.IsPublic
	}

	if err := h.analyticsService.CreateReport(report); err != nil {
		h.logger.Error("Failed to create report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create report"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Report created successfully",
		"report":  report,
	})
}

// UpdateReport handles updating a report.
func (h *AnalyticsHandler) UpdateReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	report, err := h.analyticsService.GetReport(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Status      string `json:"status"`
		Schedule    string `json:"schedule"`
		Format      string `json:"format"`
		IsPublic    *bool  `json:"is_public"`
		Config      string `json:"config"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update report request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update fields
	if req.Name != "" {
		report.Name = req.Name
	}
	if req.Description != "" {
		report.Description = req.Description
	}
	if req.Type != "" {
		report.Type = req.Type
	}
	if req.Status != "" {
		report.Status = req.Status
	}
	if req.Schedule != "" {
		report.Schedule = req.Schedule
	}
	if req.Format != "" {
		report.Format = req.Format
	}
	if req.IsPublic != nil {
		report.IsPublic = *req.IsPublic
	}
	if req.Config != "" {
		report.Config = req.Config
	}
	if req.Metadata != "" {
		report.Metadata = req.Metadata
	}

	if err := h.analyticsService.UpdateReport(report); err != nil {
		h.logger.Error("Failed to update report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Report updated successfully",
		"report":  report,
	})
}

// DeleteReport handles deleting a report.
func (h *AnalyticsHandler) DeleteReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	if err := h.analyticsService.DeleteReport(uint(id)); err != nil {
		h.logger.Error("Failed to delete report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Report deleted successfully"})
}

// GenerateReport handles generating a report.
func (h *AnalyticsHandler) GenerateReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	generation, err := h.analyticsService.GenerateReport(uint(id))
	if err != nil {
		h.logger.Error("Failed to generate report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Report generation started",
		"generation": generation,
	})
}

// DownloadReport handles downloading a generated report.
func (h *AnalyticsHandler) DownloadReport(c *gin.Context) {
	// Implementation for downloading reports
	c.JSON(http.StatusOK, gin.H{"message": "Report download functionality not implemented yet"})
}

// Dashboard Management Handlers

// GetDashboards handles getting a list of dashboards.
func (h *AnalyticsHandler) GetDashboards(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	dashboards, total, err := h.analyticsService.GetDashboards(tenantID.(uint), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get dashboards", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dashboards"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dashboards": dashboards,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}

// GetDashboard handles getting a specific dashboard.
func (h *AnalyticsHandler) GetDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard ID"})
		return
	}

	dashboard, err := h.analyticsService.GetDashboard(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dashboard not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"dashboard": dashboard})
}

// CreateDashboard handles creating a new dashboard.
func (h *AnalyticsHandler) CreateDashboard(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		IsPublic    *bool  `json:"is_public"`
		IsDefault   *bool  `json:"is_default"`
		Layout      string `json:"layout"`
		Settings    string `json:"settings"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create dashboard request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	dashboard := &models.Dashboard{
		TenantID:    tenantID.(uint),
		UserID:      userID.(uint),
		Name:        req.Name,
		Description: req.Description,
		Layout:      req.Layout,
		Settings:    req.Settings,
		Metadata:    req.Metadata,
	}

	if req.IsPublic != nil {
		dashboard.IsPublic = *req.IsPublic
	}
	if req.IsDefault != nil {
		dashboard.IsDefault = *req.IsDefault
	}

	if err := h.analyticsService.CreateDashboard(dashboard); err != nil {
		h.logger.Error("Failed to create dashboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create dashboard"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Dashboard created successfully",
		"dashboard": dashboard,
	})
}

// UpdateDashboard handles updating a dashboard.
func (h *AnalyticsHandler) UpdateDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard ID"})
		return
	}

	dashboard, err := h.analyticsService.GetDashboard(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dashboard not found"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsPublic    *bool  `json:"is_public"`
		IsDefault   *bool  `json:"is_default"`
		Layout      string `json:"layout"`
		Settings    string `json:"settings"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update dashboard request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update fields
	if req.Name != "" {
		dashboard.Name = req.Name
	}
	if req.Description != "" {
		dashboard.Description = req.Description
	}
	if req.IsPublic != nil {
		dashboard.IsPublic = *req.IsPublic
	}
	if req.IsDefault != nil {
		dashboard.IsDefault = *req.IsDefault
	}
	if req.Layout != "" {
		dashboard.Layout = req.Layout
	}
	if req.Settings != "" {
		dashboard.Settings = req.Settings
	}
	if req.Metadata != "" {
		dashboard.Metadata = req.Metadata
	}

	if err := h.analyticsService.UpdateDashboard(dashboard); err != nil {
		h.logger.Error("Failed to update dashboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update dashboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Dashboard updated successfully",
		"dashboard": dashboard,
	})
}

// DeleteDashboard handles deleting a dashboard.
func (h *AnalyticsHandler) DeleteDashboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard ID"})
		return
	}

	if err := h.analyticsService.DeleteDashboard(uint(id)); err != nil {
		h.logger.Error("Failed to delete dashboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete dashboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dashboard deleted successfully"})
}

// Widget Management Handlers

// GetDashboardWidgets handles getting widgets for a dashboard.
func (h *AnalyticsHandler) GetDashboardWidgets(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard ID"})
		return
	}

	widgets, err := h.analyticsService.GetDashboardWidgets(uint(id))
	if err != nil {
		h.logger.Error("Failed to get dashboard widgets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dashboard widgets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"widgets": widgets})
}

// AddDashboardWidget handles adding a widget to a dashboard.
func (h *AnalyticsHandler) AddDashboardWidget(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard ID"})
		return
	}

	var req struct {
		Type        string `json:"type" binding:"required"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Position    int    `json:"position"`
		Size        string `json:"size"`
		Config      string `json:"config"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid add widget request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	widget := &models.DashboardWidget{
		DashboardID: uint(id),
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Position:    req.Position,
		Size:        req.Size,
		Config:      req.Config,
		Metadata:    req.Metadata,
	}

	if err := h.analyticsService.AddDashboardWidget(widget); err != nil {
		h.logger.Error("Failed to add dashboard widget", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add dashboard widget"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Dashboard widget added successfully",
		"widget":  widget,
	})
}

// UpdateDashboardWidget handles updating a dashboard widget.
func (h *AnalyticsHandler) UpdateDashboardWidget(c *gin.Context) {
	widgetID, err := strconv.ParseUint(c.Param("widget_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid widget ID"})
		return
	}

	var req struct {
		Type        string `json:"type"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Position    *int   `json:"position"`
		Size        string `json:"size"`
		Config      string `json:"config"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update widget request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Get existing widget - we need to add this method to the service
	// For now, we'll create a simple widget object
	widget := &models.DashboardWidget{
		ID: uint(widgetID),
	}

	// Update fields
	if req.Type != "" {
		widget.Type = req.Type
	}
	if req.Title != "" {
		widget.Title = req.Title
	}
	if req.Description != "" {
		widget.Description = req.Description
	}
	if req.Position != nil {
		widget.Position = *req.Position
	}
	if req.Size != "" {
		widget.Size = req.Size
	}
	if req.Config != "" {
		widget.Config = req.Config
	}
	if req.Metadata != "" {
		widget.Metadata = req.Metadata
	}

	if err := h.analyticsService.UpdateDashboardWidget(widget); err != nil {
		h.logger.Error("Failed to update dashboard widget", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update dashboard widget"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Dashboard widget updated successfully",
		"widget":  widget,
	})
}

// DeleteDashboardWidget handles deleting a dashboard widget.
func (h *AnalyticsHandler) DeleteDashboardWidget(c *gin.Context) {
	widgetID, err := strconv.ParseUint(c.Param("widget_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid widget ID"})
		return
	}

	if err := h.analyticsService.DeleteDashboardWidget(uint(widgetID)); err != nil {
		h.logger.Error("Failed to delete dashboard widget", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete dashboard widget"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dashboard widget deleted successfully"})
}

// Data Export Handlers

// ExportCSV handles exporting data as CSV.
func (h *AnalyticsHandler) ExportCSV(c *gin.Context) {
	// Implementation for CSV export
	c.JSON(http.StatusOK, gin.H{"message": "CSV export functionality not implemented yet"})
}

// ExportJSON handles exporting data as JSON.
func (h *AnalyticsHandler) ExportJSON(c *gin.Context) {
	// Implementation for JSON export
	c.JSON(http.StatusOK, gin.H{"message": "JSON export functionality not implemented yet"})
}

// ExportExcel handles exporting data as Excel.
func (h *AnalyticsHandler) ExportExcel(c *gin.Context) {
	// Implementation for Excel export
	c.JSON(http.StatusOK, gin.H{"message": "Excel export functionality not implemented yet"})
}
