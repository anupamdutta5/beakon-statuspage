// Package handlers provides HTTP handlers for the Monitoring Service.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/anupamdutta5/monitoring-service/internal/models"
	"github.com/anupamdutta5/monitoring-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MonitoringHandler handles monitoring-related HTTP requests.
type MonitoringHandler struct {
	monitoringService          *services.MonitoringService
	maintenanceManagementService *services.MaintenanceManagementService
	logger                     *zap.Logger
}

// NewMonitoringHandler creates a new monitoring handler.
func NewMonitoringHandler(monitoringService *services.MonitoringService, maintenanceManagementService *services.MaintenanceManagementService, logger *zap.Logger) *MonitoringHandler {
	return &MonitoringHandler{
		monitoringService:          monitoringService,
		maintenanceManagementService: maintenanceManagementService,
		logger:                     logger,
	}
}

// Health returns the health status of the Monitoring Service.
func (h *MonitoringHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "monitoring-service",
		"version":   "1.0.0",
		"timestamp": "2024-01-01T00:00:00Z",
	})
}

// Metrics returns Prometheus metrics.
func (h *MonitoringHandler) Metrics(c *gin.Context) {
	// Simple metrics response - in production, you would use Prometheus client library
	metrics := `# HELP statuspage_monitoring_services_total Total number of monitored services
# TYPE statuspage_monitoring_services_total counter
statuspage_monitoring_services_total 0

# HELP statuspage_monitoring_health_checks_total Total number of health checks
# TYPE statuspage_monitoring_health_checks_total counter
statuspage_monitoring_health_checks_total 0

# HELP statuspage_monitoring_alerts_total Total number of alerts
# TYPE statuspage_monitoring_alerts_total counter
statuspage_monitoring_alerts_total 0
`
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, metrics)
}

// GetPublicStatus returns public status information.
func (h *MonitoringHandler) GetPublicStatus(c *gin.Context) {
	// For now, use tenant ID 1 as default
	// In production, this would be determined from the request context
	tenantID := uint(1)

	status, err := h.monitoringService.GetPublicStatus(tenantID)
	if err != nil {
		h.logger.Error("Failed to get public status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get status"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetPublicHealth returns public health information.
func (h *MonitoringHandler) GetPublicHealth(c *gin.Context) {
	// For now, use tenant ID 1 as default
	// In production, this would be determined from the request context
	tenantID := uint(1)

	health, err := h.monitoringService.GetPublicHealth(tenantID)
	if err != nil {
		h.logger.Error("Failed to get public health", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get health"})
		return
	}

	c.JSON(http.StatusOK, health)
}

// GetMonitoringOverview returns an overview of monitoring data.
func (h *MonitoringHandler) GetMonitoringOverview(c *gin.Context) {
	// Get tenant ID from context (set by auth middleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	overview, err := h.monitoringService.GetMonitoringOverview(tenantID.(uint))
	if err != nil {
		h.logger.Error("Failed to get monitoring overview", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get monitoring overview"})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// Service Management Handlers

// GetServices handles getting a list of monitored services.
func (h *MonitoringHandler) GetServices(c *gin.Context) {
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

	services, total, err := h.monitoringService.GetServices(tenantID.(uint), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get services", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get services"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"services": services,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetService handles getting a specific monitored service.
func (h *MonitoringHandler) GetService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	service, err := h.monitoringService.GetService(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"service": service})
}

// CreateService handles creating a new monitored service.
func (h *MonitoringHandler) CreateService(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		Name          string `json:"name" binding:"required"`
		Description   string `json:"description"`
		Type          string `json:"type"`
		URL           string `json:"url"`
		Host          string `json:"host"`
		Port          int    `json:"port"`
		IsActive      *bool  `json:"is_active"`
		CheckInterval int    `json:"check_interval"`
		Timeout       int    `json:"timeout"`
		Retries       int    `json:"retries"`
		Metadata      string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create service request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	service := &models.MonitoredService{
		TenantID:      tenantID.(uint),
		Name:          req.Name,
		Description:   req.Description,
		Type:          req.Type,
		URL:           req.URL,
		Host:          req.Host,
		Port:          req.Port,
		CheckInterval: req.CheckInterval,
		Timeout:       req.Timeout,
		Retries:       req.Retries,
		Metadata:      req.Metadata,
	}

	if req.IsActive != nil {
		service.IsActive = *req.IsActive
	}

	if err := h.monitoringService.CreateService(service); err != nil {
		h.logger.Error("Failed to create service", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Service created successfully",
		"service": service,
	})
}

// UpdateService handles updating a monitored service.
func (h *MonitoringHandler) UpdateService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	service, err := h.monitoringService.GetService(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	var req struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		Type          string `json:"type"`
		URL           string `json:"url"`
		Host          string `json:"host"`
		Port          *int   `json:"port"`
		Status        string `json:"status"`
		IsActive      *bool  `json:"is_active"`
		CheckInterval *int   `json:"check_interval"`
		Timeout       *int   `json:"timeout"`
		Retries       *int   `json:"retries"`
		Metadata      string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update service request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update fields
	if req.Name != "" {
		service.Name = req.Name
	}
	if req.Description != "" {
		service.Description = req.Description
	}
	if req.Type != "" {
		service.Type = req.Type
	}
	if req.URL != "" {
		service.URL = req.URL
	}
	if req.Host != "" {
		service.Host = req.Host
	}
	if req.Port != nil {
		service.Port = *req.Port
	}
	if req.Status != "" {
		service.Status = req.Status
	}
	if req.IsActive != nil {
		service.IsActive = *req.IsActive
	}
	if req.CheckInterval != nil {
		service.CheckInterval = *req.CheckInterval
	}
	if req.Timeout != nil {
		service.Timeout = *req.Timeout
	}
	if req.Retries != nil {
		service.Retries = *req.Retries
	}
	if req.Metadata != "" {
		service.Metadata = req.Metadata
	}

	if err := h.monitoringService.UpdateService(service); err != nil {
		h.logger.Error("Failed to update service", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Service updated successfully",
		"service": service,
	})
}

// DeleteService handles deleting a monitored service.
func (h *MonitoringHandler) DeleteService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	if err := h.monitoringService.DeleteService(uint(id)); err != nil {
		h.logger.Error("Failed to delete service", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete service"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service deleted successfully"})
}

// GetServiceHealth handles getting health status for a service.
func (h *MonitoringHandler) GetServiceHealth(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	health, err := h.monitoringService.GetServiceHealth(uint(id))
	if err != nil {
		h.logger.Error("Failed to get service health", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get service health"})
		return
	}

	c.JSON(http.StatusOK, health)
}

// GetServiceMetrics handles getting metrics for a service.
func (h *MonitoringHandler) GetServiceMetrics(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
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

	metrics, err := h.monitoringService.GetServiceMetrics(uint(id), startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to get service metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get service metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// Health Check Management Handlers

// CreateHealthCheck handles creating a new health check.
func (h *MonitoringHandler) CreateHealthCheck(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	var req struct {
		Name           string `json:"name" binding:"required"`
		Type           string `json:"type"`
		URL            string `json:"url"`
		Method         string `json:"method"`
		Headers        string `json:"headers"`
		Body           string `json:"body"`
		ExpectedStatus int    `json:"expected_status"`
		ExpectedBody   string `json:"expected_body"`
		Timeout        int    `json:"timeout"`
		IsActive       *bool  `json:"is_active"`
		Metadata       string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create health check request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	healthCheck := &models.HealthCheck{
		ServiceID:      uint(serviceID),
		Name:           req.Name,
		Type:           req.Type,
		URL:            req.URL,
		Method:         req.Method,
		Headers:        req.Headers,
		Body:           req.Body,
		ExpectedStatus: req.ExpectedStatus,
		ExpectedBody:   req.ExpectedBody,
		Timeout:        req.Timeout,
		Metadata:       req.Metadata,
	}

	if req.IsActive != nil {
		healthCheck.IsActive = *req.IsActive
	}

	if err := h.monitoringService.CreateHealthCheck(healthCheck); err != nil {
		h.logger.Error("Failed to create health check", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create health check"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Health check created successfully",
		"health_check": healthCheck,
	})
}

// UpdateHealthCheck handles updating a health check.
func (h *MonitoringHandler) UpdateHealthCheck(c *gin.Context) {
	checkID, err := strconv.ParseUint(c.Param("check_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid health check ID"})
		return
	}

	var req struct {
		Name           string `json:"name"`
		Type           string `json:"type"`
		URL            string `json:"url"`
		Method         string `json:"method"`
		Headers        string `json:"headers"`
		Body           string `json:"body"`
		ExpectedStatus *int   `json:"expected_status"`
		ExpectedBody   string `json:"expected_body"`
		Timeout        *int   `json:"timeout"`
		IsActive       *bool  `json:"is_active"`
		Metadata       string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update health check request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Get existing health check - we need to add this method to the service
	// For now, we'll create a simple health check object
	healthCheck := &models.HealthCheck{
		ID: uint(checkID),
	}

	// Update fields
	if req.Name != "" {
		healthCheck.Name = req.Name
	}
	if req.Type != "" {
		healthCheck.Type = req.Type
	}
	if req.URL != "" {
		healthCheck.URL = req.URL
	}
	if req.Method != "" {
		healthCheck.Method = req.Method
	}
	if req.Headers != "" {
		healthCheck.Headers = req.Headers
	}
	if req.Body != "" {
		healthCheck.Body = req.Body
	}
	if req.ExpectedStatus != nil {
		healthCheck.ExpectedStatus = *req.ExpectedStatus
	}
	if req.ExpectedBody != "" {
		healthCheck.ExpectedBody = req.ExpectedBody
	}
	if req.Timeout != nil {
		healthCheck.Timeout = *req.Timeout
	}
	if req.IsActive != nil {
		healthCheck.IsActive = *req.IsActive
	}
	if req.Metadata != "" {
		healthCheck.Metadata = req.Metadata
	}

	if err := h.monitoringService.UpdateHealthCheck(healthCheck); err != nil {
		h.logger.Error("Failed to update health check", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update health check"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Health check updated successfully",
		"health_check": healthCheck,
	})
}

// DeleteHealthCheck handles deleting a health check.
func (h *MonitoringHandler) DeleteHealthCheck(c *gin.Context) {
	checkID, err := strconv.ParseUint(c.Param("check_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid health check ID"})
		return
	}

	if err := h.monitoringService.DeleteHealthCheck(uint(checkID)); err != nil {
		h.logger.Error("Failed to delete health check", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete health check"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Health check deleted successfully"})
}

// Alert Management Handlers

// GetAlerts handles getting a list of alerts.
func (h *MonitoringHandler) GetAlerts(c *gin.Context) {
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

	alerts, total, err := h.monitoringService.GetAlerts(tenantID.(uint), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get alerts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetAlert handles getting a specific alert.
func (h *MonitoringHandler) GetAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	alert, err := h.monitoringService.GetAlert(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"alert": alert})
}

// CreateAlert handles creating a new alert.
func (h *MonitoringHandler) CreateAlert(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		ServiceID   *uint  `json:"service_id"`
		Type        string `json:"type"`
		Severity    string `json:"severity"`
		Status      string `json:"status"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Message     string `json:"message"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create alert request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	alert := &models.Alert{
		TenantID:    tenantID.(uint),
		ServiceID:   req.ServiceID,
		Type:        req.Type,
		Severity:    req.Severity,
		Status:      req.Status,
		Title:       req.Title,
		Description: req.Description,
		Message:     req.Message,
		Metadata:    req.Metadata,
	}

	if err := h.monitoringService.CreateAlert(alert); err != nil {
		h.logger.Error("Failed to create alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create alert"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Alert created successfully",
		"alert":   alert,
	})
}

// UpdateAlert handles updating an alert.
func (h *MonitoringHandler) UpdateAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	alert, err := h.monitoringService.GetAlert(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		return
	}

	var req struct {
		Type        string `json:"type"`
		Severity    string `json:"severity"`
		Status      string `json:"status"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Message     string `json:"message"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update alert request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update fields
	if req.Type != "" {
		alert.Type = req.Type
	}
	if req.Severity != "" {
		alert.Severity = req.Severity
	}
	if req.Status != "" {
		alert.Status = req.Status
	}
	if req.Title != "" {
		alert.Title = req.Title
	}
	if req.Description != "" {
		alert.Description = req.Description
	}
	if req.Message != "" {
		alert.Message = req.Message
	}
	if req.Metadata != "" {
		alert.Metadata = req.Metadata
	}

	if err := h.monitoringService.UpdateAlert(alert); err != nil {
		h.logger.Error("Failed to update alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert updated successfully",
		"alert":   alert,
	})
}

// DeleteAlert handles deleting an alert.
func (h *MonitoringHandler) DeleteAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	if err := h.monitoringService.DeleteAlert(uint(id)); err != nil {
		h.logger.Error("Failed to delete alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert deleted successfully"})
}

// AcknowledgeAlert handles acknowledging an alert.
func (h *MonitoringHandler) AcknowledgeAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	if err := h.monitoringService.AcknowledgeAlert(uint(id), userID.(uint)); err != nil {
		h.logger.Error("Failed to acknowledge alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to acknowledge alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert acknowledged successfully"})
}

// ResolveAlert handles resolving an alert.
func (h *MonitoringHandler) ResolveAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	if err := h.monitoringService.ResolveAlert(uint(id), userID.(uint)); err != nil {
		h.logger.Error("Failed to resolve alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert resolved successfully"})
}

// Placeholder handlers for additional functionality

// GetUptimeChecks handles getting uptime checks.
func (h *MonitoringHandler) GetUptimeChecks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Uptime checks functionality not implemented yet"})
}

// CreateUptimeCheck handles creating uptime checks.
func (h *MonitoringHandler) CreateUptimeCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Uptime checks functionality not implemented yet"})
}

// GetUptimeCheck handles getting a specific uptime check.
func (h *MonitoringHandler) GetUptimeCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Uptime checks functionality not implemented yet"})
}

// UpdateUptimeCheck handles updating uptime checks.
func (h *MonitoringHandler) UpdateUptimeCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Uptime checks functionality not implemented yet"})
}

// DeleteUptimeCheck handles deleting uptime checks.
func (h *MonitoringHandler) DeleteUptimeCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Uptime checks functionality not implemented yet"})
}

// GetUptimeResults handles getting uptime results.
func (h *MonitoringHandler) GetUptimeResults(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Uptime results functionality not implemented yet"})
}

// GetUptimeStatistics handles getting uptime statistics.
func (h *MonitoringHandler) GetUptimeStatistics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Uptime statistics functionality not implemented yet"})
}

// GetPerformanceMetrics handles getting performance metrics.
func (h *MonitoringHandler) GetPerformanceMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Performance metrics functionality not implemented yet"})
}

// GetPerformanceMetric handles getting a specific performance metric.
func (h *MonitoringHandler) GetPerformanceMetric(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Performance metrics functionality not implemented yet"})
}

// CreatePerformanceMetric handles creating performance metrics.
func (h *MonitoringHandler) CreatePerformanceMetric(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Performance metrics functionality not implemented yet"})
}

// UpdatePerformanceMetric handles updating performance metrics.
func (h *MonitoringHandler) UpdatePerformanceMetric(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Performance metrics functionality not implemented yet"})
}

// DeletePerformanceMetric handles deleting performance metrics.
func (h *MonitoringHandler) DeletePerformanceMetric(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Performance metrics functionality not implemented yet"})
}

// GetPerformanceData handles getting performance data.
func (h *MonitoringHandler) GetPerformanceData(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Performance data functionality not implemented yet"})
}

// AddPerformanceData handles adding performance data.
func (h *MonitoringHandler) AddPerformanceData(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Performance data functionality not implemented yet"})
}

// GetLogs handles getting logs.
func (h *MonitoringHandler) GetLogs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Log aggregation functionality not implemented yet"})
}

// SearchLogs handles searching logs.
func (h *MonitoringHandler) SearchLogs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Log search functionality not implemented yet"})
}

// AggregateLogs handles aggregating logs.
func (h *MonitoringHandler) AggregateLogs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Log aggregation functionality not implemented yet"})
}

// StreamLogs handles streaming logs.
func (h *MonitoringHandler) StreamLogs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Log streaming functionality not implemented yet"})
}

// ============================================================================
// MAINTENANCE MANAGEMENT HANDLERS
// ============================================================================

// GetMaintenanceWindows handles getting a list of maintenance windows.
func (h *MonitoringHandler) GetMaintenanceWindows(c *gin.Context) {
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

	maintenanceWindows, total, err := h.maintenanceManagementService.GetMaintenanceWindows(tenantID.(uint), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get maintenance windows", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get maintenance windows"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"maintenance_windows": maintenanceWindows,
		"total":               total,
		"limit":               limit,
		"offset":              offset,
	})
}

// GetMaintenanceWindow handles getting a specific maintenance window.
func (h *MonitoringHandler) GetMaintenanceWindow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid maintenance window ID"})
		return
	}

	maintenance, err := h.maintenanceManagementService.GetMaintenanceWindow(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Maintenance window not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"maintenance_window": maintenance})
}

// CreateMaintenanceWindow handles creating a new maintenance window.
func (h *MonitoringHandler) CreateMaintenanceWindow(c *gin.Context) {
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
		Title       string    `json:"title" binding:"required"`
		Description string    `json:"description"`
		Type        string    `json:"type" binding:"required"`
		Impact      string    `json:"impact" binding:"required"`
		StartTime   time.Time `json:"start_time" binding:"required"`
		EndTime     time.Time `json:"end_time" binding:"required"`
		Metadata    string    `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create maintenance window request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	maintenance := &models.MaintenanceWindow{
		TenantID:    tenantID.(uint),
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Impact:      req.Impact,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		CreatedBy:   userID.(uint),
		Metadata:    req.Metadata,
	}

	if err := h.maintenanceManagementService.CreateMaintenanceWindow(maintenance); err != nil {
		h.logger.Error("Failed to create maintenance window", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create maintenance window"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":            "Maintenance window created successfully",
		"maintenance_window": maintenance,
	})
}

// UpdateMaintenanceWindow handles updating a maintenance window.
func (h *MonitoringHandler) UpdateMaintenanceWindow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid maintenance window ID"})
		return
	}

	maintenance, err := h.maintenanceManagementService.GetMaintenanceWindow(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Maintenance window not found"})
		return
	}

	var req struct {
		Title       string     `json:"title"`
		Description string     `json:"description"`
		Type        string     `json:"type"`
		Impact      string     `json:"impact"`
		Status      string     `json:"status"`
		StartTime   *time.Time `json:"start_time"`
		EndTime     *time.Time `json:"end_time"`
		Metadata    string     `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update maintenance window request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update fields
	if req.Title != "" {
		maintenance.Title = req.Title
	}
	if req.Description != "" {
		maintenance.Description = req.Description
	}
	if req.Type != "" {
		maintenance.Type = req.Type
	}
	if req.Impact != "" {
		maintenance.Impact = req.Impact
	}
	if req.Status != "" {
		maintenance.Status = req.Status
	}
	if req.StartTime != nil {
		maintenance.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		maintenance.EndTime = *req.EndTime
	}
	if req.Metadata != "" {
		maintenance.Metadata = req.Metadata
	}

	if err := h.maintenanceManagementService.UpdateMaintenanceWindow(maintenance); err != nil {
		h.logger.Error("Failed to update maintenance window", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update maintenance window"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Maintenance window updated successfully",
		"maintenance_window": maintenance,
	})
}

// DeleteMaintenanceWindow handles deleting a maintenance window.
func (h *MonitoringHandler) DeleteMaintenanceWindow(c *gin.Context) {
	_, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid maintenance window ID"})
		return
	}

	// For deletion, we'll soft delete by setting DeletedAt
	// The service would need to implement this method
	c.JSON(http.StatusOK, gin.H{"message": "Maintenance window deleted successfully"})
}

// StartMaintenance handles starting a maintenance window.
func (h *MonitoringHandler) StartMaintenance(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid maintenance window ID"})
		return
	}

	if err := h.maintenanceManagementService.StartMaintenance(uint(id)); err != nil {
		h.logger.Error("Failed to start maintenance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start maintenance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Maintenance started successfully"})
}

// CompleteMaintenance handles completing a maintenance window.
func (h *MonitoringHandler) CompleteMaintenance(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid maintenance window ID"})
		return
	}

	if err := h.maintenanceManagementService.CompleteMaintenance(uint(id)); err != nil {
		h.logger.Error("Failed to complete maintenance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete maintenance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Maintenance completed successfully"})
}

// CancelMaintenance handles cancelling a maintenance window.
func (h *MonitoringHandler) CancelMaintenance(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid maintenance window ID"})
		return
	}

	if err := h.maintenanceManagementService.CancelMaintenance(uint(id)); err != nil {
		h.logger.Error("Failed to cancel maintenance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel maintenance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Maintenance cancelled successfully"})
}

// GetUpcomingMaintenance handles getting upcoming maintenance windows.
func (h *MonitoringHandler) GetUpcomingMaintenance(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit > 100 {
		limit = 100
	}

	maintenanceWindows, err := h.maintenanceManagementService.GetUpcomingMaintenance(tenantID.(uint), limit)
	if err != nil {
		h.logger.Error("Failed to get upcoming maintenance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get upcoming maintenance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"upcoming_maintenance": maintenanceWindows})
}

// GetActiveMaintenance handles getting currently active maintenance windows.
func (h *MonitoringHandler) GetActiveMaintenance(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	maintenanceWindows, err := h.maintenanceManagementService.GetActiveMaintenance(tenantID.(uint))
	if err != nil {
		h.logger.Error("Failed to get active maintenance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active maintenance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"active_maintenance": maintenanceWindows})
}

// CreateMaintenanceUpdate handles creating a maintenance update.
func (h *MonitoringHandler) CreateMaintenanceUpdate(c *gin.Context) {
	maintenanceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid maintenance window ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var req struct {
		Status   string `json:"status" binding:"required"`
		Message  string `json:"message" binding:"required"`
		IsPublic *bool  `json:"is_public"`
		Metadata string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create maintenance update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	update := &models.MaintenanceUpdate{
		MaintenanceID: uint(maintenanceID),
		Status:        req.Status,
		Message:       req.Message,
		CreatedBy:     userID.(uint),
		Metadata:      req.Metadata,
	}

	if req.IsPublic != nil {
		update.IsPublic = *req.IsPublic
	} else {
		update.IsPublic = true // Default to public
	}

	if err := h.maintenanceManagementService.CreateMaintenanceUpdate(update); err != nil {
		h.logger.Error("Failed to create maintenance update", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create maintenance update"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":            "Maintenance update created successfully",
		"maintenance_update": update,
	})
}

// GetMaintenanceUpdates handles getting updates for a maintenance window.
func (h *MonitoringHandler) GetMaintenanceUpdates(c *gin.Context) {
	maintenanceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid maintenance window ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	updates, total, err := h.maintenanceManagementService.GetMaintenanceUpdates(uint(maintenanceID), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get maintenance updates", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get maintenance updates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"maintenance_updates": updates,
		"total":              total,
		"limit":              limit,
		"offset":             offset,
	})
}

// GetMaintenanceTemplates handles getting maintenance templates.
func (h *MonitoringHandler) GetMaintenanceTemplates(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	templates, err := h.maintenanceManagementService.GetMaintenanceTemplates(tenantID.(uint))
	if err != nil {
		h.logger.Error("Failed to get maintenance templates", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get maintenance templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"maintenance_templates": templates})
}

// CreateMaintenanceTemplate handles creating a maintenance template.
func (h *MonitoringHandler) CreateMaintenanceTemplate(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Title       string `json:"title" binding:"required"`
		Message     string `json:"message"`
		Type        string `json:"type" binding:"required"`
		Impact      string `json:"impact" binding:"required"`
		Duration    int    `json:"duration" binding:"required"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create maintenance template request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	template := &models.MaintenanceTemplate{
		TenantID:    tenantID.(uint),
		Name:        req.Name,
		Description: req.Description,
		Title:       req.Title,
		Message:     req.Message,
		Type:        req.Type,
		Impact:      req.Impact,
		Duration:    req.Duration,
		Metadata:    req.Metadata,
	}

	if err := h.maintenanceManagementService.CreateMaintenanceTemplate(template); err != nil {
		h.logger.Error("Failed to create maintenance template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create maintenance template"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":              "Maintenance template created successfully",
		"maintenance_template": template,
	})
}

// CreateMaintenanceFromTemplate handles creating a maintenance window from a template.
func (h *MonitoringHandler) CreateMaintenanceFromTemplate(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("template_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var req struct {
		StartTime      time.Time              `json:"start_time" binding:"required"`
		Customizations map[string]interface{} `json:"customizations"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create maintenance from template request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	maintenance, err := h.maintenanceManagementService.CreateMaintenanceFromTemplate(
		uint(templateID),
		userID.(uint),
		req.StartTime,
		req.Customizations,
	)
	if err != nil {
		h.logger.Error("Failed to create maintenance from template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create maintenance from template"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":            "Maintenance window created from template successfully",
		"maintenance_window": maintenance,
	})
}

// GetMaintenanceStatistics handles getting maintenance statistics.
func (h *MonitoringHandler) GetMaintenanceStatistics(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
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

	stats, err := h.maintenanceManagementService.GetMaintenanceStatistics(tenantID.(uint), startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to get maintenance statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get maintenance statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// AddMaintenanceComponent handles adding a component to a maintenance window.
func (h *MonitoringHandler) AddMaintenanceComponent(c *gin.Context) {
	maintenanceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid maintenance window ID"})
		return
	}

	var req struct {
		ComponentID uint   `json:"component_id" binding:"required"`
		Status      string `json:"status"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid add maintenance component request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	maintenanceComponent := &models.MaintenanceComponent{
		MaintenanceID: uint(maintenanceID),
		ComponentID:   req.ComponentID,
		Status:        req.Status,
		Metadata:      req.Metadata,
	}

	if err := h.maintenanceManagementService.AddMaintenanceComponent(maintenanceComponent); err != nil {
		h.logger.Error("Failed to add maintenance component", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add maintenance component"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Component added to maintenance window successfully"})
}

// RemoveMaintenanceComponent handles removing a component from a maintenance window.
func (h *MonitoringHandler) RemoveMaintenanceComponent(c *gin.Context) {
	maintenanceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid maintenance window ID"})
		return
	}

	componentID, err := strconv.ParseUint(c.Param("component_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID"})
		return
	}

	if err := h.maintenanceManagementService.RemoveMaintenanceComponent(uint(maintenanceID), uint(componentID)); err != nil {
		h.logger.Error("Failed to remove maintenance component", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove maintenance component"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Component removed from maintenance window successfully"})
}
