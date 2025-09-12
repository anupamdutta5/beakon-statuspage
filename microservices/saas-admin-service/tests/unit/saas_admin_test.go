package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enterprise-status/statuspage-saas-admin-service/internal/config"
	"github.com/enterprise-status/statuspage-saas-admin-service/internal/handlers"
	"github.com/enterprise-status/statuspage-saas-admin-service/internal/models"
	"github.com/enterprise-status/statuspage-saas-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSaaSAdminHandler_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	saasAdminService, _ := services.NewSaaSAdminService(cfg, nil)
	handler := handlers.NewSaaSAdminHandler(saasAdminService, nil)

	router := gin.New()
	router.GET("/health", handler.HealthCheck)

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
	assert.Equal(t, "saas-admin-service", response["service"])
}

func TestSaaSAdminHandler_CreatePlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	saasAdminService, _ := services.NewSaaSAdminService(cfg, nil)
	handler := handlers.NewSaaSAdminHandler(saasAdminService, nil)

	router := gin.New()
	router.POST("/plans", handler.CreatePlan)

	plan := models.SaaSPlan{
		Name:            "Pro Plan",
		Slug:            "pro-plan",
		Description:     "Professional plan with advanced features",
		Price:           2999.0,
		Currency:        "USD",
		BillingInterval: "monthly",
		Features:        `{"max_incidents": 100, "max_components": 50, "custom_domain": true, "api_access": true, "priority_support": true}`,
		Limits:          `{"api_calls_per_month": 10000, "storage_gb": 10, "team_members": 5}`,
		IsActive:        true,
	}

	jsonData, _ := json.Marshal(plan)
	req, _ := http.NewRequest("POST", "/plans", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.SaaSPlan
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, plan.Name, response.Name)
	assert.Equal(t, plan.Description, response.Description)
	assert.Equal(t, plan.Price, response.Price)
	assert.Equal(t, plan.Currency, response.Currency)
	assert.Equal(t, plan.BillingInterval, response.BillingInterval)
	assert.Equal(t, plan.IsActive, response.IsActive)
	assert.NotEmpty(t, response.ID)
}

func TestSaaSAdminHandler_GetPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	saasAdminService, _ := services.NewSaaSAdminService(cfg, nil)
	handler := handlers.NewSaaSAdminHandler(saasAdminService, nil)

	// Create a plan first
	plan := models.SaaSPlan{
		Name:            "Test Plan",
		Slug:            "test-plan",
		Description:     "Test plan description",
		Price:           1999.0,
		Currency:        "USD",
		BillingInterval: "monthly",
		IsActive:        true,
	}
	err := saasAdminService.CreatePlan(context.Background(), &plan)
	createdPlan := &plan
	require.NoError(t, err)

	router := gin.New()
	router.GET("/plans/:id", handler.GetPlan)

	// Test
	req, _ := http.NewRequest("GET", "/plans/"+fmt.Sprintf("%d", createdPlan.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.SaaSPlan
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, createdPlan.ID, response.ID)
	assert.Equal(t, createdPlan.Name, response.Name)
}

func TestSaaSAdminHandler_UpdatePlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	saasAdminService, _ := services.NewSaaSAdminService(cfg, nil)
	handler := handlers.NewSaaSAdminHandler(saasAdminService, nil)

	// Create a plan first
	plan := models.SaaSPlan{
		Name:            "Original Plan",
		Slug:            "original-plan",
		Description:     "Original description",
		Price:           1999.0,
		Currency:        "USD",
		BillingInterval: "monthly",
		IsActive:        true,
	}
	err := saasAdminService.CreatePlan(context.Background(), &plan)
	createdPlan := &plan
	require.NoError(t, err)

	router := gin.New()
	router.PUT("/plans/:id", handler.UpdatePlan)

	// Update data
	updateData := models.SaaSPlan{
		Name:            "Updated Plan",
		Slug:            "updated-plan",
		Description:     "Updated description",
		Price:           2999.0,
		Currency:        "USD",
		BillingInterval: "annual",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/plans/"+fmt.Sprintf("%d", createdPlan.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.SaaSPlan
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Updated Plan", response.Name)
	assert.Equal(t, "Updated description", response.Description)
	assert.Equal(t, 2999.0, response.Price)
	assert.Equal(t, "annual", response.BillingInterval)
}

func TestSaaSAdminHandler_DeletePlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	saasAdminService, _ := services.NewSaaSAdminService(cfg, nil)
	handler := handlers.NewSaaSAdminHandler(saasAdminService, nil)

	// Create a plan first
	plan := models.SaaSPlan{
		Name:            "Test Plan",
		Slug:            "test-plan-2",
		Description:     "Test plan description",
		Price:           1999.0,
		Currency:        "USD",
		BillingInterval: "monthly",
		IsActive:        true,
	}
	err := saasAdminService.CreatePlan(context.Background(), &plan)
	createdPlan := &plan
	require.NoError(t, err)

	router := gin.New()
	router.DELETE("/plans/:id", handler.DeletePlan)

	// Test
	req, _ := http.NewRequest("DELETE", "/plans/"+fmt.Sprintf("%d", createdPlan.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Plan deleted successfully", response["message"])

	// Verify deletion
	_, err = saasAdminService.GetPlan(context.Background(), createdPlan.ID)
	assert.Error(t, err)
}

func TestSaaSAdminHandler_ListPlans(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	saasAdminService, _ := services.NewSaaSAdminService(cfg, nil)
	handler := handlers.NewSaaSAdminHandler(saasAdminService, nil)

	// Create multiple plans
	plans := []models.SaaSPlan{
		{Name: "Plan 1", Slug: "plan-1", Description: "Plan 1 description", Price: 999.0, Currency: "USD", BillingInterval: "monthly", IsActive: true},
		{Name: "Plan 2", Slug: "plan-2", Description: "Plan 2 description", Price: 1999.0, Currency: "USD", BillingInterval: "monthly", IsActive: true},
		{Name: "Plan 3", Slug: "plan-3", Description: "Plan 3 description", Price: 2999.0, Currency: "USD", BillingInterval: "annual", IsActive: false},
	}

	for _, plan := range plans {
		err := saasAdminService.CreatePlan(context.Background(), &plan)
		require.NoError(t, err)
	}

	router := gin.New()
	router.GET("/plans", handler.ListPlans)

	// Test
	req, _ := http.NewRequest("GET", "/plans", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Plans  []models.SaaSPlan `json:"plans"`
		Total  int               `json:"total"`
		Limit  int               `json:"limit"`
		Offset int               `json:"offset"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Plans), 3)
	assert.GreaterOrEqual(t, response.Total, 3)
}

func TestSaaSAdminHandler_CreateFeatureFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	saasAdminService, _ := services.NewSaaSAdminService(cfg, nil)
	handler := handlers.NewSaaSAdminHandler(saasAdminService, nil)

	router := gin.New()
	router.POST("/feature-flags", handler.CreateFeatureFlag)

	featureFlag := models.SaaSFeatureFlag{
		Name:        "new_dashboard",
		Description: "New dashboard feature",
		IsEnabled:   true,
		Config:      `{"type": "boolean", "value": "true", "target_audience": "all"}`,
		Metadata:    `{"rollout_percentage": 100, "release_date": "2024-01-01"}`,
	}

	jsonData, _ := json.Marshal(featureFlag)
	req, _ := http.NewRequest("POST", "/feature-flags", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.SaaSFeatureFlag
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, featureFlag.Name, response.Name)
	assert.Equal(t, featureFlag.Description, response.Description)
	assert.Equal(t, featureFlag.IsEnabled, response.IsEnabled)
	assert.NotEmpty(t, response.ID)
}

func TestSaaSAdminHandler_GetFeatureFlags(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	saasAdminService, _ := services.NewSaaSAdminService(cfg, nil)
	handler := handlers.NewSaaSAdminHandler(saasAdminService, nil)

	// Create some feature flags
	featureFlags := []models.SaaSFeatureFlag{
		{Name: "feature1", Description: "Feature 1", IsEnabled: true, Config: `{"type": "boolean", "value": "true"}`},
		{Name: "feature2", Description: "Feature 2", IsEnabled: false, Config: `{"type": "boolean", "value": "false"}`},
		{Name: "feature3", Description: "Feature 3", IsEnabled: true, Config: `{"type": "string", "value": "test"}`},
	}

	for _, flag := range featureFlags {
		err := saasAdminService.CreateFeatureFlag(context.Background(), &flag)
		require.NoError(t, err)
	}

	router := gin.New()
	router.GET("/feature-flags", handler.ListFeatureFlags)

	// Test
	req, _ := http.NewRequest("GET", "/feature-flags", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		FeatureFlags []models.SaaSFeatureFlag `json:"feature_flags"`
		Total        int                      `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.FeatureFlags), 3)
	assert.GreaterOrEqual(t, response.Total, 3)
}

func TestSaaSAdminHandler_GetPlatformStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	saasAdminService, _ := services.NewSaaSAdminService(cfg, nil)
	handler := handlers.NewSaaSAdminHandler(saasAdminService, nil)

	router := gin.New()
	router.GET("/stats", handler.GetStats)

	// Test
	req, _ := http.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "total_tenants")
	assert.Contains(t, response, "total_users")
	assert.Contains(t, response, "total_incidents")
	assert.Contains(t, response, "total_components")
	assert.Contains(t, response, "revenue")
	assert.Contains(t, response, "active_subscriptions")
}

func TestSaaSAdminHandler_UpdateFeatureFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	saasAdminService, _ := services.NewSaaSAdminService(cfg, nil)
	handler := handlers.NewSaaSAdminHandler(saasAdminService, nil)

	// Create a feature flag first
	featureFlag := models.SaaSFeatureFlag{
		Name:        "test_feature",
		Description: "Test feature",
		IsEnabled:   false,
		Config:      `{"type": "boolean", "value": "false"}`,
	}
	err := saasAdminService.CreateFeatureFlag(context.Background(), &featureFlag)
	createdFlag := &featureFlag
	require.NoError(t, err)

	router := gin.New()
	router.PUT("/feature-flags/:id", handler.UpdateFeatureFlag)

	// Update data
	updateData := models.SaaSFeatureFlag{
		IsEnabled: true,
		Config:    `{"value": "true"}`,
		Metadata:  `{"rollout_percentage": 50}`,
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/feature-flags/"+fmt.Sprintf("%d", createdFlag.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.SaaSFeatureFlag
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, true, response.IsEnabled)
}

func TestSaaSAdminHandler_DeleteFeatureFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	saasAdminService, _ := services.NewSaaSAdminService(cfg, nil)
	handler := handlers.NewSaaSAdminHandler(saasAdminService, nil)

	// Create a feature flag first
	featureFlag := models.SaaSFeatureFlag{
		Name:        "test_feature",
		Description: "Test feature",
		IsEnabled:   true,
		Config:      `{"type": "boolean", "value": "true"}`,
	}
	err := saasAdminService.CreateFeatureFlag(context.Background(), &featureFlag)
	createdFlag := &featureFlag
	require.NoError(t, err)

	router := gin.New()
	router.DELETE("/feature-flags/:id", handler.DeleteFeatureFlag)

	// Test
	req, _ := http.NewRequest("DELETE", "/feature-flags/"+fmt.Sprintf("%d", createdFlag.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Feature flag deleted successfully", response["message"])

	// Verify deletion
	_, err = saasAdminService.GetFeatureFlag(context.Background(), createdFlag.ID)
	assert.Error(t, err)
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate
	err = db.AutoMigrate(&models.SaaSPlan{}, &models.SaaSFeatureFlag{}, &models.SaaSStats{})
	require.NoError(t, err)

	return db
}
