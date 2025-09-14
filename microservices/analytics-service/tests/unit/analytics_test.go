// Package unit provides unit tests for the Analytics Service.
package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/enterprise-status/statuspage-analytics-service/internal/handlers"
	"github.com/enterprise-status/statuspage-analytics-service/internal/models"
	"github.com/enterprise-status/statuspage-analytics-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAnalyticsHandler_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	analyticsService := services.NewAnalyticsService(db, logger)
	handler := handlers.NewAnalyticsHandler(analyticsService, logger)

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
	assert.Equal(t, "analytics-service", response["service"])
}

func TestAnalyticsHandler_GetPublicMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	analyticsService := services.NewAnalyticsService(db, logger)
	handler := handlers.NewAnalyticsHandler(analyticsService, logger)

	// Create a public metric
	metric := models.Metric{
		Name:        "Public Metric",
		Description: "A public metric for testing",
		Type:        "counter",
		Unit:        "requests",
		Category:    "performance",
		IsActive:    true,
		IsPublic:    true,
		TenantID:    1,
	}
	err := analyticsService.CreateMetric(&metric)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/public/metrics", handler.GetPublicMetrics)

	// Test
	req, _ := http.NewRequest("GET", "/public/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	metrics, ok := response["metrics"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(metrics), 1)
}

func TestAnalyticsHandler_CreateMetric(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	analyticsService := services.NewAnalyticsService(db, logger)
	handler := handlers.NewAnalyticsHandler(analyticsService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.POST("/metrics", handler.CreateMetric)

	// Test data
	metricData := models.Metric{
		Name:        "Test Metric",
		Description: "Test metric description",
		Type:        "gauge",
		Unit:        "seconds",
		Category:    "performance",
		IsActive:    true,
		IsPublic:    false,
		TenantID:    1,
	}

	jsonData, _ := json.Marshal(metricData)
	req, _ := http.NewRequest("POST", "/metrics", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var responseWrapper struct {
		Metric models.Metric `json:"metric"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, metricData.Name, responseWrapper.Metric.Name)
	assert.Equal(t, metricData.Description, responseWrapper.Metric.Description)
	assert.Equal(t, metricData.Type, responseWrapper.Metric.Type)
	assert.Equal(t, metricData.Unit, responseWrapper.Metric.Unit)
	assert.Equal(t, metricData.Category, responseWrapper.Metric.Category)
	assert.Equal(t, metricData.TenantID, responseWrapper.Metric.TenantID)
}

func TestAnalyticsHandler_GetMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	analyticsService := services.NewAnalyticsService(db, logger)
	handler := handlers.NewAnalyticsHandler(analyticsService, logger)

	// Create multiple metrics
	metrics := []models.Metric{
		{Name: "Metric 1", Description: "Description 1", Type: "counter", Unit: "requests", Category: "performance", TenantID: 1},
		{Name: "Metric 2", Description: "Description 2", Type: "gauge", Unit: "seconds", Category: "system", TenantID: 1},
		{Name: "Metric 3", Description: "Description 3", Type: "histogram", Unit: "bytes", Category: "business", TenantID: 1},
	}

	for _, metric := range metrics {
		err := analyticsService.CreateMetric(&metric)
		require.NoError(t, err)
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/metrics", handler.GetMetrics)

	// Test
	req, _ := http.NewRequest("GET", "/metrics?tenant_id=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	metricsList, ok := response["metrics"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(metricsList))
}

func TestAnalyticsHandler_GetMetric(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	analyticsService := services.NewAnalyticsService(db, logger)
	handler := handlers.NewAnalyticsHandler(analyticsService, logger)

	// Create a metric first
	metric := models.Metric{
		Name:        "Test Metric",
		Description: "Test metric description",
		Type:        "counter",
		Unit:        "requests",
		Category:    "performance",
		TenantID:    1,
	}
	err := analyticsService.CreateMetric(&metric)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/metrics/:id", handler.GetMetric)

	// Test
	req, _ := http.NewRequest("GET", fmt.Sprintf("/metrics/%d", metric.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseWrapper struct {
		Metric models.Metric `json:"metric"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, metric.Name, responseWrapper.Metric.Name)
	assert.Equal(t, metric.Description, responseWrapper.Metric.Description)
	assert.Equal(t, metric.Type, responseWrapper.Metric.Type)
	assert.Equal(t, metric.Unit, responseWrapper.Metric.Unit)
	assert.Equal(t, metric.Category, responseWrapper.Metric.Category)
}

func TestAnalyticsHandler_UpdateMetric(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	analyticsService := services.NewAnalyticsService(db, logger)
	handler := handlers.NewAnalyticsHandler(analyticsService, logger)

	// Create a metric first
	metric := models.Metric{
		Name:        "Original Metric",
		Description: "Original description",
		Type:        "counter",
		Unit:        "requests",
		Category:    "performance",
		TenantID:    1,
	}
	err := analyticsService.CreateMetric(&metric)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.PUT("/metrics/:id", handler.UpdateMetric)

	// Update data
	updateData := models.Metric{
		Name:        "Updated Metric",
		Description: "Updated description",
		Type:        "gauge",
		Unit:        "seconds",
		Category:    "system",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/metrics/%d", metric.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseWrapper struct {
		Metric models.Metric `json:"metric"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, "Updated Metric", responseWrapper.Metric.Name)
	assert.Equal(t, "Updated description", responseWrapper.Metric.Description)
	assert.Equal(t, "gauge", responseWrapper.Metric.Type)
	assert.Equal(t, "seconds", responseWrapper.Metric.Unit)
	assert.Equal(t, "system", responseWrapper.Metric.Category)
}

func TestAnalyticsHandler_DeleteMetric(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	analyticsService := services.NewAnalyticsService(db, logger)
	handler := handlers.NewAnalyticsHandler(analyticsService, logger)

	// Create a metric first
	metric := models.Metric{
		Name:        "Test Metric",
		Description: "Test metric description",
		Type:        "counter",
		Unit:        "requests",
		Category:    "performance",
		TenantID:    1,
	}
	err := analyticsService.CreateMetric(&metric)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.DELETE("/metrics/:id", handler.DeleteMetric)

	// Test
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/metrics/%d", metric.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify deletion
	_, err = analyticsService.GetMetric(metric.ID)
	assert.Error(t, err)
}

func TestAnalyticsHandler_AddMetricData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	analyticsService := services.NewAnalyticsService(db, logger)
	handler := handlers.NewAnalyticsHandler(analyticsService, logger)

	// Create a metric first
	metric := models.Metric{
		Name:        "Test Metric",
		Description: "Test metric description",
		Type:        "counter",
		Unit:        "requests",
		Category:    "performance",
		TenantID:    1,
	}
	err := analyticsService.CreateMetric(&metric)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.POST("/metrics/:id/data", handler.AddMetricData)

	// Test data
	dataPoint := models.MetricDataPoint{
		Value:     100.5,
		Timestamp: time.Now(),
		Source:    "api",
		Labels:    `{"endpoint": "/api/v1/users"}`,
		Metadata:  `{"user_id": "123"}`,
	}

	jsonData, _ := json.Marshal(dataPoint)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/metrics/%d/data", metric.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var responseWrapper struct {
		DataPoint models.MetricDataPoint `json:"data_point"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, dataPoint.Value, responseWrapper.DataPoint.Value)
	assert.Equal(t, dataPoint.Source, responseWrapper.DataPoint.Source)
	assert.Equal(t, dataPoint.Labels, responseWrapper.DataPoint.Labels)
	assert.Equal(t, dataPoint.Metadata, responseWrapper.DataPoint.Metadata)
}

func TestAnalyticsHandler_GetMetricData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	analyticsService := services.NewAnalyticsService(db, logger)
	handler := handlers.NewAnalyticsHandler(analyticsService, logger)

	// Create a metric first
	metric := models.Metric{
		Name:        "Test Metric",
		Description: "Test metric description",
		Type:        "counter",
		Unit:        "requests",
		Category:    "performance",
		TenantID:    1,
	}
	err := analyticsService.CreateMetric(&metric)
	require.NoError(t, err)

	// Add some data points
	dataPoints := []models.MetricDataPoint{
		{Value: 100.0, Timestamp: time.Now().Add(-2 * time.Hour), Source: "api"},
		{Value: 150.0, Timestamp: time.Now().Add(-1 * time.Hour), Source: "api"},
		{Value: 200.0, Timestamp: time.Now(), Source: "api"},
	}

	for _, dp := range dataPoints {
		dp.MetricID = metric.ID
		err := analyticsService.AddMetricData(&dp)
		require.NoError(t, err)
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/metrics/:id/data", handler.GetMetricData)

	// Test
	req, _ := http.NewRequest("GET", fmt.Sprintf("/metrics/%d/data", metric.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataPointsList, ok := response["data_points"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(dataPointsList))
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.Metric{},
		&models.MetricDataPoint{},
		&models.Report{},
		&models.ReportGeneration{},
		&models.Dashboard{},
	)
	require.NoError(t, err)

	return db
}
