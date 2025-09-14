package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enterprise-status/statuspage-component-service/internal/handlers"
	"github.com/enterprise-status/statuspage-component-service/internal/models"
	"github.com/enterprise-status/statuspage-component-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestComponentHandler_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	componentService := services.NewComponentService(db, logger)
	handler := handlers.NewComponentHandler(componentService, logger)

	router := gin.New()
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
	assert.Equal(t, "component-service", response["service"])
}

func TestComponentHandler_CreateComponent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	componentService := services.NewComponentService(db, logger)
	handler := handlers.NewComponentHandler(componentService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Next()
	})
	router.POST("/components", handler.CreateComponent)

	component := models.Component{
		Name:        "API Server",
		Description: "Main API server component",
		Status:      "operational",
		TenantID:    1,
		Metadata:    `{"version":"1.0.0","region":"us-east-1"}`,
	}

	jsonData, _ := json.Marshal(component)
	req, _ := http.NewRequest("POST", "/components", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var responseWrapper struct {
		Message   string           `json:"message"`
		Component models.Component `json:"component"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, "Component created successfully", responseWrapper.Message)
	assert.Equal(t, component.Name, responseWrapper.Component.Name)
	assert.Equal(t, component.Description, responseWrapper.Component.Description)
	assert.Equal(t, component.Status, responseWrapper.Component.Status)
	assert.NotEmpty(t, responseWrapper.Component.ID)
}

func TestComponentHandler_GetComponent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	componentService := services.NewComponentService(db, nil)
	handler := handlers.NewComponentHandler(componentService, nil)

	// Create a component first
	component := models.Component{
		Name:        "Test Component",
		Description: "Test component description",
		Status:      "operational",
		TenantID:    1,
	}
	err := componentService.CreateComponent(&component)
	require.NoError(t, err)

	router := gin.New()
	router.GET("/components/:id", handler.GetComponent)

	// Test
	req, _ := http.NewRequest("GET", fmt.Sprintf("/components/%d", component.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Component
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, component.ID, response.ID)
	assert.Equal(t, component.Name, response.Name)
}

func TestComponentHandler_UpdateComponent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	componentService := services.NewComponentService(db, nil)
	handler := handlers.NewComponentHandler(componentService, nil)

	// Create a component first
	component := models.Component{
		Name:        "Test Component",
		Description: "Test component description",
		Status:      "operational",
		TenantID:    1,
	}
	err := componentService.CreateComponent(&component)
	require.NoError(t, err)

	router := gin.New()
	router.PUT("/components/:id", handler.UpdateComponent)

	// Update data
	updateData := models.Component{
		Name:        "Updated Component",
		Description: "Updated component description",
		Status:      "degraded_performance",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/components/%d", component.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Component
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Updated Component", response.Name)
	assert.Equal(t, "degraded_performance", response.Status)

	// Verify in database
	updatedComponent, err := componentService.GetComponent(component.ID)
	require.NoError(t, err)

	assert.Equal(t, "Updated Component", updatedComponent.Name)
	assert.Equal(t, "degraded_performance", updatedComponent.Status)
}

func TestComponentHandler_ListComponents(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	componentService := services.NewComponentService(db, nil)
	handler := handlers.NewComponentHandler(componentService, nil)

	// Create multiple components
	components := []models.Component{
		{Name: "Component 1", Description: "Description 1", Status: "operational", TenantID: 1},
		{Name: "Component 2", Description: "Description 2", Status: "degraded_performance", TenantID: 1},
		{Name: "Component 3", Description: "Description 3", Status: "partial_outage", TenantID: 1},
	}

	for _, component := range components {
		err := componentService.CreateComponent(&component)
		require.NoError(t, err)
	}

	router := gin.New()
	router.GET("/components", handler.GetComponents)

	// Test
	req, _ := http.NewRequest("GET", "/components?tenant_id=test-tenant-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Components []models.Component `json:"components"`
		Total      int                `json:"total"`
		Limit      int                `json:"limit"`
		Offset     int                `json:"offset"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Components), 3)
	assert.GreaterOrEqual(t, response.Total, 3)
}

func TestComponentHandler_DeleteComponent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	componentService := services.NewComponentService(db, nil)
	handler := handlers.NewComponentHandler(componentService, nil)

	// Create a component first
	component := models.Component{
		Name:        "Test Component",
		Description: "Test component description",
		Status:      "operational",
		TenantID:    1,
	}
	err := componentService.CreateComponent(&component)
	require.NoError(t, err)

	router := gin.New()
	router.DELETE("/components/:id", handler.DeleteComponent)

	// Test
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/components/%d", component.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Component deleted successfully", response["message"])

	// Verify deletion
	_, err = componentService.GetComponent(component.ID)
	assert.Error(t, err)
}

func TestComponentHandler_UpdateComponentStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	componentService := services.NewComponentService(db, nil)
	handler := handlers.NewComponentHandler(componentService, nil)

	// Create a component first
	component := models.Component{
		Name:        "Test Component",
		Description: "Test component description",
		Status:      "operational",
		TenantID:    1,
	}
	err := componentService.CreateComponent(&component)
	require.NoError(t, err)

	router := gin.New()
	router.PUT("/components/:id/status", handler.UpdateComponentStatus)

	// Update status
	statusUpdate := map[string]interface{}{
		"status":  "degraded_performance",
		"message": "Experiencing slow response times",
	}

	jsonData, _ := json.Marshal(statusUpdate)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/components/%d", component.ID)+"/status", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Component
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "degraded_performance", response.Status)
}

func TestComponentHandler_GetComponentsByCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	componentService := services.NewComponentService(db, nil)
	handler := handlers.NewComponentHandler(componentService, nil)

	// Create components with different categories
	components := []models.Component{
		{Name: "API Component 1", Status: "operational", TenantID: 1},
		{Name: "API Component 2", Status: "operational", TenantID: 1},
		{Name: "Database Component", Status: "operational", TenantID: 1},
	}

	for _, component := range components {
		err := componentService.CreateComponent(&component)
		require.NoError(t, err)
	}

	router := gin.New()
	router.GET("/components/category/:category", handler.GetComponents)

	// Test
	req, _ := http.NewRequest("GET", "/components/category/API?tenant_id=test-tenant-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Components []models.Component `json:"components"`
		Total      int                `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, len(response.Components))
	assert.Equal(t, 2, response.Total)

	// Verify all components are returned
	for _, comp := range response.Components {
		assert.NotEmpty(t, comp.Name)
	}
}

func TestComponentHandler_GetComponentMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	componentService := services.NewComponentService(db, nil)
	handler := handlers.NewComponentHandler(componentService, nil)

	// Create a component first
	component := models.Component{
		Name:        "Test Component",
		Description: "Test component description",
		Status:      "operational",
		TenantID:    1,
	}
	err := componentService.CreateComponent(&component)
	require.NoError(t, err)

	router := gin.New()
	router.GET("/components/:id/metrics", handler.GetComponent)

	// Test
	req, _ := http.NewRequest("GET", fmt.Sprintf("/components/%d/metrics", component.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "uptime")
	assert.Contains(t, response, "response_time")
	assert.Contains(t, response, "error_rate")
}

func TestComponentHandler_BulkUpdateStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	componentService := services.NewComponentService(db, nil)
	handler := handlers.NewComponentHandler(componentService, nil)

	// Create multiple components
	components := []models.Component{
		{Name: "Component 1", Status: "operational", TenantID: 1},
		{Name: "Component 2", Status: "operational", TenantID: 1},
		{Name: "Component 3", Status: "operational", TenantID: 1},
	}

	var componentIDs []uint
	for _, component := range components {
		err := componentService.CreateComponent(&component)
		require.NoError(t, err)
		componentIDs = append(componentIDs, component.ID)
	}

	router := gin.New()
	router.PUT("/components/bulk/status", handler.UpdateComponent)

	// Bulk update status
	bulkUpdate := map[string]interface{}{
		"component_ids": componentIDs,
		"status":        "maintenance",
		"message":       "Scheduled maintenance",
	}

	jsonData, _ := json.Marshal(bulkUpdate)
	req, _ := http.NewRequest("PUT", "/components/bulk/status", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Bulk status update completed", response["message"])
	assert.Equal(t, 3, int(response["updated_count"].(float64)))

	// Verify all components are updated
	for _, id := range componentIDs {
		comp, err := componentService.GetComponent(id)
		require.NoError(t, err)
		assert.Equal(t, "maintenance", comp.Status)
	}
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate all models
	err = db.AutoMigrate(
		&models.Component{},
		&models.ComponentGroup{},
		&models.ComponentStatus{},
		&models.ComponentHistory{},
		&models.ComponentMetric{},
		&models.ComponentAlert{},
		&models.ComponentWebhook{},
	)
	require.NoError(t, err)

	return db
}
