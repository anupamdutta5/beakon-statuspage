package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/anupamdutta5/saas-admin-service/internal/cache"
	"github.com/anupamdutta5/saas-admin-service/internal/config"
	"github.com/anupamdutta5/saas-admin-service/internal/events"
	"github.com/anupamdutta5/saas-admin-service/internal/handlers"
	"github.com/anupamdutta5/saas-admin-service/internal/models"
	"github.com/anupamdutta5/saas-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSaaSAdminHandler_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := loadTestConfig(t)
	cache := setupMockCache()
	saasAdminService, _ := services.NewSaaSAdminService(cfg, logger, db, cache)
	tenantAdminDB := setupTestDB(t) // Mock tenant admin DB
	eventPublisher := setupMockEventPublisher(logger)
	serviceURLs := setupMockServiceURLs()
	handler := handlers.NewSaaSAdminHandler(saasAdminService, tenantAdminDB, eventPublisher, serviceURLs, logger)

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
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := loadTestConfig(t)
	cache := setupMockCache()
	saasAdminService, _ := services.NewSaaSAdminService(cfg, logger, db, cache)
	tenantAdminDB := setupTestDB(t) // Mock tenant admin DB
	eventPublisher := setupMockEventPublisher(logger)
	serviceURLs := setupMockServiceURLs()
	handler := handlers.NewSaaSAdminHandler(saasAdminService, tenantAdminDB, eventPublisher, serviceURLs, logger)

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

	var response struct {
		Plan    models.SaaSPlan `json:"plan"`
		Message string          `json:"message"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, plan.Name, response.Plan.Name)
	assert.Equal(t, plan.Description, response.Plan.Description)
	assert.Equal(t, plan.Price, response.Plan.Price)
	assert.Equal(t, plan.Currency, response.Plan.Currency)
	assert.Equal(t, plan.BillingInterval, response.Plan.BillingInterval)
	assert.Equal(t, plan.IsActive, response.Plan.IsActive)
	assert.NotEmpty(t, response.Plan.ID)
}

func TestSaaSAdminHandler_GetPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := loadTestConfig(t)
	cache := setupMockCache()
	saasAdminService, _ := services.NewSaaSAdminService(cfg, logger, db, cache)
	tenantAdminDB := setupTestDB(t) // Mock tenant admin DB
	eventPublisher := setupMockEventPublisher(logger)
	serviceURLs := setupMockServiceURLs()
	handler := handlers.NewSaaSAdminHandler(saasAdminService, tenantAdminDB, eventPublisher, serviceURLs, logger)

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

	var response struct {
		Plan models.SaaSPlan `json:"plan"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, createdPlan.ID, response.Plan.ID)
	assert.Equal(t, createdPlan.Name, response.Plan.Name)
}

func TestSaaSAdminHandler_UpdatePlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := loadTestConfig(t)
	cache := setupMockCache()
	saasAdminService, _ := services.NewSaaSAdminService(cfg, logger, db, cache)
	tenantAdminDB := setupTestDB(t) // Mock tenant admin DB
	eventPublisher := setupMockEventPublisher(logger)
	serviceURLs := setupMockServiceURLs()
	handler := handlers.NewSaaSAdminHandler(saasAdminService, tenantAdminDB, eventPublisher, serviceURLs, logger)

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

	var response struct {
		Plan models.SaaSPlan `json:"plan"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Updated Plan", response.Plan.Name)
	assert.Equal(t, "Updated description", response.Plan.Description)
	assert.Equal(t, 2999.0, response.Plan.Price)
	assert.Equal(t, "annual", response.Plan.BillingInterval)
}

func TestSaaSAdminHandler_DeletePlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := loadTestConfig(t)
	cache := setupMockCache()
	saasAdminService, _ := services.NewSaaSAdminService(cfg, logger, db, cache)
	tenantAdminDB := setupTestDB(t) // Mock tenant admin DB
	eventPublisher := setupMockEventPublisher(logger)
	serviceURLs := setupMockServiceURLs()
	handler := handlers.NewSaaSAdminHandler(saasAdminService, tenantAdminDB, eventPublisher, serviceURLs, logger)

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

	assert.Equal(t, "SaaS plan deleted successfully", response["message"])

	// Verify deletion
	_, err = saasAdminService.GetPlan(context.Background(), createdPlan.ID)
	assert.Error(t, err)
}

func TestSaaSAdminHandler_ListPlans(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := loadTestConfig(t)
	cache := setupMockCache()
	saasAdminService, _ := services.NewSaaSAdminService(cfg, logger, db, cache)
	tenantAdminDB := setupTestDB(t) // Mock tenant admin DB
	eventPublisher := setupMockEventPublisher(logger)
	serviceURLs := setupMockServiceURLs()
	handler := handlers.NewSaaSAdminHandler(saasAdminService, tenantAdminDB, eventPublisher, serviceURLs, logger)

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
		Count  int               `json:"count"`
		Limit  int               `json:"limit"`
		Offset int               `json:"offset"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Plans), 3)
	assert.GreaterOrEqual(t, response.Count, 3)
}

func TestSaaSAdminHandler_CreateFeatureFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := loadTestConfig(t)
	cache := setupMockCache()
	saasAdminService, _ := services.NewSaaSAdminService(cfg, logger, db, cache)
	tenantAdminDB := setupTestDB(t) // Mock tenant admin DB
	eventPublisher := setupMockEventPublisher(logger)
	serviceURLs := setupMockServiceURLs()
	handler := handlers.NewSaaSAdminHandler(saasAdminService, tenantAdminDB, eventPublisher, serviceURLs, logger)

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

	var response struct {
		FeatureFlag models.SaaSFeatureFlag `json:"feature_flag"`
		Message     string                 `json:"message"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, featureFlag.Name, response.FeatureFlag.Name)
	assert.Equal(t, featureFlag.Description, response.FeatureFlag.Description)
	assert.Equal(t, featureFlag.IsEnabled, response.FeatureFlag.IsEnabled)
	assert.NotEmpty(t, response.FeatureFlag.ID)
}

func TestSaaSAdminHandler_GetFeatureFlags(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := loadTestConfig(t)
	cache := setupMockCache()
	saasAdminService, _ := services.NewSaaSAdminService(cfg, logger, db, cache)
	tenantAdminDB := setupTestDB(t) // Mock tenant admin DB
	eventPublisher := setupMockEventPublisher(logger)
	serviceURLs := setupMockServiceURLs()
	handler := handlers.NewSaaSAdminHandler(saasAdminService, tenantAdminDB, eventPublisher, serviceURLs, logger)

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
		Count        int                      `json:"count"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.FeatureFlags), 3)
	assert.GreaterOrEqual(t, response.Count, 3)
}

func TestSaaSAdminHandler_GetPlatformStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := loadTestConfig(t)
	cache := setupMockCache()
	saasAdminService, _ := services.NewSaaSAdminService(cfg, logger, db, cache)
	tenantAdminDB := setupTestDB(t) // Mock tenant admin DB
	eventPublisher := setupMockEventPublisher(logger)
	serviceURLs := setupMockServiceURLs()
	handler := handlers.NewSaaSAdminHandler(saasAdminService, tenantAdminDB, eventPublisher, serviceURLs, logger)

	router := gin.New()
	router.GET("/stats", handler.GetStats)

	// Test
	req, _ := http.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Stats map[string]interface{} `json:"stats"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response.Stats, "total_tenants")
	assert.Contains(t, response.Stats, "total_users")
	assert.Contains(t, response.Stats, "total_features")
	assert.Contains(t, response.Stats, "total_plans")
	assert.Contains(t, response.Stats, "total_revenue")
	assert.Contains(t, response.Stats, "active_tenants")
}

func TestSaaSAdminHandler_UpdateFeatureFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := loadTestConfig(t)
	cache := setupMockCache()
	saasAdminService, _ := services.NewSaaSAdminService(cfg, logger, db, cache)
	tenantAdminDB := setupTestDB(t) // Mock tenant admin DB
	eventPublisher := setupMockEventPublisher(logger)
	serviceURLs := setupMockServiceURLs()
	handler := handlers.NewSaaSAdminHandler(saasAdminService, tenantAdminDB, eventPublisher, serviceURLs, logger)

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

	var response struct {
		FeatureFlag models.SaaSFeatureFlag `json:"feature_flag"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, true, response.FeatureFlag.IsEnabled)
}

func TestSaaSAdminHandler_DeleteFeatureFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := loadTestConfig(t)
	cache := setupMockCache()
	saasAdminService, _ := services.NewSaaSAdminService(cfg, logger, db, cache)
	tenantAdminDB := setupTestDB(t) // Mock tenant admin DB
	eventPublisher := setupMockEventPublisher(logger)
	serviceURLs := setupMockServiceURLs()
	handler := handlers.NewSaaSAdminHandler(saasAdminService, tenantAdminDB, eventPublisher, serviceURLs, logger)

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

	assert.Equal(t, "SaaS feature flag deleted successfully", response["message"])

	// Verify deletion
	_, err = saasAdminService.GetFeatureFlag(context.Background(), createdFlag.ID)
	assert.Error(t, err)
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)

	// Manually execute migrations without PostgreSQL-specific default values
	// Use AUTOINCREMENT for IDs instead of UUID for SQLite compatibility
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS saas_plans (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			name TEXT NOT NULL UNIQUE,
			slug TEXT NOT NULL UNIQUE,
			description TEXT,
			price REAL NOT NULL,
			currency TEXT DEFAULT 'USD',
			billing_interval TEXT DEFAULT 'monthly',
			max_tenants INTEGER DEFAULT 1,
			max_users INTEGER DEFAULT 5,
			max_services INTEGER DEFAULT 10,
			max_monitors INTEGER DEFAULT 50,
			max_subscribers INTEGER DEFAULT 1000,
			max_incidents INTEGER DEFAULT 100,
			max_maintenance INTEGER DEFAULT 50,
			custom_domain BOOLEAN DEFAULT FALSE,
			white_label BOOLEAN DEFAULT FALSE,
			api BOOLEAN DEFAULT FALSE,
			integrations BOOLEAN DEFAULT FALSE,
			analytics BOOLEAN DEFAULT FALSE,
			support TEXT DEFAULT 'email',
			is_active BOOLEAN DEFAULT TRUE,
			is_public BOOLEAN DEFAULT TRUE,
			is_popular BOOLEAN DEFAULT FALSE,
			button_text TEXT DEFAULT 'Get Started',
			button_url TEXT DEFAULT '/signup',
			display_order INTEGER DEFAULT 0,
			features TEXT,
			limits TEXT,
			metadata TEXT
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS saas_feature_flags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			is_enabled BOOLEAN DEFAULT FALSE,
			config TEXT,
			metadata TEXT
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS saas_stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			key TEXT NOT NULL UNIQUE,
			value TEXT,
			metadata TEXT
		)
	`).Error
	require.NoError(t, err)

	return db
}

// MockCache implements the cache.Cache interface for unit tests
type MockCache struct{}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}

func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	return nil
}

func (m *MockCache) Del(ctx context.Context, keys ...string) error {
	return nil
}

func (m *MockCache) IsHealthy(ctx context.Context) bool {
	return true
}

// Helper function to setup mock cache for tests
func setupMockCache() cache.Cache {
	return &MockCache{}
}

// Helper function to setup mock event publisher for tests
func setupMockEventPublisher(logger *zap.Logger) *events.Publisher {
	// Create a mock RabbitMQ connection (won't actually connect in tests)
	// Return nil for unit tests as we don't need actual event publishing
	return nil
}

// Helper function to setup mock service URLs for tests
func setupMockServiceURLs() config.ServiceURLs {
	return config.ServiceURLs{
		TenantAdminService: "http://localhost:8099",
	}
}

// Helper function to load test configuration
func loadTestConfig(t *testing.T) *config.Config {
	// Use development environment for tests
	os.Setenv("ENVIRONMENT", "development")
	defer os.Unsetenv("ENVIRONMENT")

	// Load configuration using the service's standard config loader
	// The shared-resilience library will automatically find the correct config path
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load test configuration: %v", err)
	}

	// Override database settings for testing (SQLite will be provided directly)
	cfg.Database.Host = "localhost"
	cfg.Database.Name = ":memory:"

	return &cfg
}
