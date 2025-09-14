// Package unit provides unit tests for the Monitoring Service.
package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/enterprise-status/statuspage-monitoring-service/internal/handlers"
	"github.com/enterprise-status/statuspage-monitoring-service/internal/models"
	"github.com/enterprise-status/statuspage-monitoring-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMonitoringHandler_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	monitoringService := services.NewMonitoringService(db, logger)
	handler := handlers.NewMonitoringHandler(monitoringService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/health", handler.Health)

	// Test
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "monitoring-service", response["service"])
}

func TestMonitoringHandler_CreateService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	monitoringService := services.NewMonitoringService(db, logger)
	handler := handlers.NewMonitoringHandler(monitoringService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.POST("/services", func(c *gin.Context) {
		// Set required context values for authentication
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		handler.CreateService(c)
	})

	// Test data
	serviceRequest := map[string]interface{}{
		"name":           "Test Service",
		"description":    "Test service description",
		"type":           "http",
		"url":            "https://example.com",
		"is_active":      true,
		"check_interval": 60,
		"timeout":        30,
		"retries":        3,
	}

	jsonData, _ := json.Marshal(serviceRequest)
	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response struct {
		Message string                  `json:"message"`
		Service models.MonitoredService `json:"service"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Service created successfully", response.Message)
	assert.Equal(t, "Test Service", response.Service.Name)
	assert.Equal(t, "Test service description", response.Service.Description)
	assert.Equal(t, "http", response.Service.Type)
	assert.Equal(t, "https://example.com", response.Service.URL)
	assert.NotZero(t, response.Service.ID)
}

func TestMonitoringHandler_GetService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	monitoringService := services.NewMonitoringService(db, logger)
	handler := handlers.NewMonitoringHandler(monitoringService, logger)

	// Create a service first
	service := models.MonitoredService{
		Name:          "Test Service",
		Description:   "Test service description",
		Type:          "http",
		URL:           "https://example.com",
		Status:        "unknown",
		IsActive:      true,
		CheckInterval: 60,
		Timeout:       30,
		Retries:       3,
		TenantID:      1,
	}
	err := monitoringService.CreateService(&service)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/services/:id", handler.GetService)

	// Test
	req, _ := http.NewRequest("GET", fmt.Sprintf("/services/%d", service.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseWrapper struct {
		Service models.MonitoredService `json:"service"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, service.Name, responseWrapper.Service.Name)
	assert.Equal(t, service.Description, responseWrapper.Service.Description)
	assert.Equal(t, service.Type, responseWrapper.Service.Type)
	assert.Equal(t, service.URL, responseWrapper.Service.URL)
}

func TestMonitoringHandler_UpdateService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	monitoringService := services.NewMonitoringService(db, logger)
	handler := handlers.NewMonitoringHandler(monitoringService, logger)

	// Create a service first
	service := models.MonitoredService{
		Name:          "Original Service",
		Description:   "Original description",
		Type:          "http",
		URL:           "https://example.com",
		Status:        "unknown",
		IsActive:      true,
		CheckInterval: 60,
		Timeout:       30,
		Retries:       3,
		TenantID:      1,
	}
	err := monitoringService.CreateService(&service)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.PUT("/services/:id", handler.UpdateService)

	// Update data
	updateData := models.MonitoredService{
		Name:          "Updated Service",
		Description:   "Updated description",
		Type:          "tcp",
		URL:           "https://updated.com",
		CheckInterval: 120,
		Timeout:       60,
		Retries:       5,
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/services/%d", service.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseWrapper struct {
		Message string                  `json:"message"`
		Service models.MonitoredService `json:"service"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, "Updated Service", responseWrapper.Service.Name)
	assert.Equal(t, "Updated description", responseWrapper.Service.Description)
	assert.Equal(t, "tcp", responseWrapper.Service.Type)
	assert.Equal(t, "https://updated.com", responseWrapper.Service.URL)
	assert.Equal(t, 120, responseWrapper.Service.CheckInterval)
	assert.Equal(t, 60, responseWrapper.Service.Timeout)
	assert.Equal(t, 5, responseWrapper.Service.Retries)
}

func TestMonitoringHandler_DeleteService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	monitoringService := services.NewMonitoringService(db, logger)
	handler := handlers.NewMonitoringHandler(monitoringService, logger)

	// Create a service first
	service := models.MonitoredService{
		Name:          "Test Service",
		Description:   "Test service description",
		Type:          "http",
		URL:           "https://example.com",
		Status:        "unknown",
		IsActive:      true,
		CheckInterval: 60,
		Timeout:       30,
		Retries:       3,
		TenantID:      1,
	}
	err := monitoringService.CreateService(&service)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.DELETE("/services/:id", handler.DeleteService)

	// Test
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/services/%d", service.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify deletion
	_, err = monitoringService.GetService(service.ID)
	assert.Error(t, err)
}

func TestMonitoringHandler_GetServices(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	monitoringService := services.NewMonitoringService(db, logger)
	handler := handlers.NewMonitoringHandler(monitoringService, logger)

	// Create multiple services
	services := []models.MonitoredService{
		{Name: "Service 1", Description: "Description 1", Type: "http", URL: "https://service1.com", TenantID: 1},
		{Name: "Service 2", Description: "Description 2", Type: "tcp", URL: "https://service2.com", TenantID: 1},
		{Name: "Service 3", Description: "Description 3", Type: "ping", URL: "https://service3.com", TenantID: 1},
	}

	for _, service := range services {
		err := monitoringService.CreateService(&service)
		require.NoError(t, err)
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/services", handler.GetServices)

	// Test
	req, _ := http.NewRequest("GET", "/services?tenant_id=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	servicesList, ok := response["services"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(servicesList))
}

func TestMonitoringHandler_CreateHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	monitoringService := services.NewMonitoringService(db, logger)
	handler := handlers.NewMonitoringHandler(monitoringService, logger)

	// Create a service first
	service := models.MonitoredService{
		Name:          "Test Service",
		Description:   "Test service description",
		Type:          "http",
		URL:           "https://example.com",
		Status:        "unknown",
		IsActive:      true,
		CheckInterval: 60,
		Timeout:       30,
		Retries:       3,
		TenantID:      1,
	}
	err := monitoringService.CreateService(&service)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.POST("/services/:id/health-checks", handler.CreateHealthCheck)

	// Test data
	healthCheck := models.HealthCheck{
		ServiceID:      service.ID,
		Name:           "Test Health Check",
		Type:           "http",
		URL:            "https://example.com/health",
		Method:         "GET",
		ExpectedStatus: 200,
		Timeout:        30,
		IsActive:       true,
	}

	jsonData, _ := json.Marshal(healthCheck)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/services/%d/health-checks", service.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var responseWrapper struct {
		HealthCheck models.HealthCheck `json:"health_check"`
		Message     string             `json:"message"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, "Health check created successfully", responseWrapper.Message)
	assert.Equal(t, healthCheck.Name, responseWrapper.HealthCheck.Name)
	assert.Equal(t, healthCheck.Type, responseWrapper.HealthCheck.Type)
	assert.Equal(t, healthCheck.URL, responseWrapper.HealthCheck.URL)
	assert.Equal(t, healthCheck.Method, responseWrapper.HealthCheck.Method)
	assert.Equal(t, healthCheck.ExpectedStatus, responseWrapper.HealthCheck.ExpectedStatus)
	assert.Equal(t, healthCheck.ServiceID, responseWrapper.HealthCheck.ServiceID)
}

func TestMonitoringHandler_CreateAlert(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	monitoringService := services.NewMonitoringService(db, logger)
	handler := handlers.NewMonitoringHandler(monitoringService, logger)

	// Create a service first
	service := models.MonitoredService{
		Name:          "Test Service",
		Description:   "Test service description",
		Type:          "http",
		URL:           "https://example.com",
		Status:        "unknown",
		IsActive:      true,
		CheckInterval: 60,
		Timeout:       30,
		Retries:       3,
		TenantID:      1,
	}
	err := monitoringService.CreateService(&service)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.POST("/alerts", handler.CreateAlert)

	// Test data
	alert := models.Alert{
		TenantID:    service.TenantID,
		ServiceID:   &service.ID,
		Type:        "service_down",
		Severity:    "critical",
		Status:      "active",
		Title:       "Service Down Alert",
		Description: "The service is down",
		Message:     "Service has been down for more than 5 minutes",
		TriggeredAt: time.Now(),
	}

	jsonData, _ := json.Marshal(alert)
	req, _ := http.NewRequest("POST", "/alerts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var responseWrapper struct {
		Message string       `json:"message"`
		Alert   models.Alert `json:"alert"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, alert.Type, responseWrapper.Alert.Type)
	assert.Equal(t, alert.Severity, responseWrapper.Alert.Severity)
	assert.Equal(t, alert.Status, responseWrapper.Alert.Status)
	assert.Equal(t, alert.Title, responseWrapper.Alert.Title)
	assert.Equal(t, alert.Description, responseWrapper.Alert.Description)
	assert.Equal(t, alert.Message, responseWrapper.Alert.Message)
	assert.Equal(t, alert.TenantID, responseWrapper.Alert.TenantID)
}

func TestMonitoringHandler_GetAlerts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	monitoringService := services.NewMonitoringService(db, logger)
	handler := handlers.NewMonitoringHandler(monitoringService, logger)

	// Create a service first
	service := models.MonitoredService{
		Name:          "Test Service",
		Description:   "Test service description",
		Type:          "http",
		URL:           "https://example.com",
		Status:        "unknown",
		IsActive:      true,
		CheckInterval: 60,
		Timeout:       30,
		Retries:       3,
		TenantID:      1,
	}
	err := monitoringService.CreateService(&service)
	require.NoError(t, err)

	// Create multiple alerts
	alerts := []models.Alert{
		{TenantID: 1, ServiceID: &service.ID, Type: "service_down", Severity: "critical", Status: "active", Title: "Alert 1", TriggeredAt: time.Now()},
		{TenantID: 1, ServiceID: &service.ID, Type: "high_response_time", Severity: "warning", Status: "active", Title: "Alert 2", TriggeredAt: time.Now()},
		{TenantID: 1, ServiceID: &service.ID, Type: "custom", Severity: "info", Status: "resolved", Title: "Alert 3", TriggeredAt: time.Now()},
	}

	for _, alert := range alerts {
		err := monitoringService.CreateAlert(&alert)
		require.NoError(t, err)
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/alerts", handler.GetAlerts)

	// Test
	req, _ := http.NewRequest("GET", "/alerts?tenant_id=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	alertsList, ok := response["alerts"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(alertsList))
}

func TestMonitoringHandler_AcknowledgeAlert(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	monitoringService := services.NewMonitoringService(db, logger)
	handler := handlers.NewMonitoringHandler(monitoringService, logger)

	// Create a service first
	service := models.MonitoredService{
		Name:          "Test Service",
		Description:   "Test service description",
		Type:          "http",
		URL:           "https://example.com",
		Status:        "unknown",
		IsActive:      true,
		CheckInterval: 60,
		Timeout:       30,
		Retries:       3,
		TenantID:      1,
	}
	err := monitoringService.CreateService(&service)
	require.NoError(t, err)

	// Create an alert
	alert := models.Alert{
		TenantID:    service.TenantID,
		ServiceID:   &service.ID,
		Type:        "service_down",
		Severity:    "critical",
		Status:      "active",
		Title:       "Service Down Alert",
		Description: "The service is down",
		Message:     "Service has been down for more than 5 minutes",
		TriggeredAt: time.Now(),
	}
	err = monitoringService.CreateAlert(&alert)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.PUT("/alerts/:id/acknowledge", handler.AcknowledgeAlert)

	// Test
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/alerts/%d/acknowledge", alert.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseWrapper struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, "Alert acknowledged successfully", responseWrapper.Message)
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.MonitoredService{},
		&models.HealthCheck{},
		&models.HealthCheckResult{},
		&models.Alert{},
		&models.UptimeCheck{},
		&models.UptimeResult{},
		&models.PerformanceMetric{},
		&models.PerformanceDataPoint{},
		&models.LogEntry{},
	)
	require.NoError(t, err)

	return db
}
