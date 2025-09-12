package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enterprise-status/statuspage-tenant-admin-service/internal/config"
	"github.com/enterprise-status/statuspage-tenant-admin-service/internal/handlers"
	"github.com/enterprise-status/statuspage-tenant-admin-service/internal/models"
	"github.com/enterprise-status/statuspage-tenant-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTenantAdminHandler_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, nil)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil)

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
	assert.Equal(t, "tenant-admin-service", response["service"])
}

func TestTenantAdminHandler_GetTenantSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, nil)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil)

	// Create tenant settings first
	settings := models.TenantSettings{
		TenantID: 1,
		Settings: `{
			"notifications": {
				"email_enabled": true,
				"sms_enabled":   false,
				"webhook_url":   "https://example.com/webhook"
			},
			"branding": {
				"logo_url":        "https://example.com/logo.png",
				"primary_color":   "#007bff",
				"secondary_color": "#6c757d"
			},
			"security": {
				"two_factor_enabled": true,
				"session_timeout":    3600
			}
		}`,
	}
	_, err := tenantAdminService.GetTenantSettings(context.Background(), settings.TenantID)
	require.NoError(t, err)

	router := gin.New()
	router.GET("/tenants/:id/settings", handler.GetTenantSettings)

	// Test
	req, _ := http.NewRequest("GET", "/tenants/test-tenant-id/settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.TenantSettings
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test-tenant-id", response.TenantID)
	assert.Contains(t, response.Settings, "notifications")
	assert.Contains(t, response.Settings, "branding")
	assert.Contains(t, response.Settings, "security")
}

func TestTenantAdminHandler_UpdateTenantSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, nil)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil)

	// Create tenant settings first
	settings := models.TenantSettings{
		TenantID: 1,
		Settings: `{
			"notifications": {
				"email_enabled": true,
				"sms_enabled":   false
			}
		}`,
	}
	_, err := tenantAdminService.GetTenantSettings(context.Background(), settings.TenantID)
	require.NoError(t, err)

	router := gin.New()
	router.PUT("/tenants/:id/settings", handler.UpdateTenantSettings)

	// Update data
	updateData := map[string]interface{}{
		"notifications": map[string]interface{}{
			"email_enabled": false,
			"sms_enabled":   true,
			"webhook_url":   "https://example.com/new-webhook",
		},
		"branding": map[string]interface{}{
			"logo_url":        "https://example.com/new-logo.png",
			"primary_color":   "#28a745",
			"secondary_color": "#dc3545",
		},
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/tenants/test-tenant-id/settings", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.TenantSettings
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test-tenant-id", response.TenantID)

	// Settings is a JSON string, so we need to parse it
	var settingsData map[string]interface{}
	err = json.Unmarshal([]byte(response.Settings), &settingsData)
	require.NoError(t, err)

	// Settings would be parsed from JSON string
	assert.Equal(t, "test-tenant-id", response.TenantID)
}

func TestTenantAdminHandler_GetTenantUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, nil)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil)

	// Create some tenant users
	users := []models.TenantAdmin{
		{
			TenantID: 1,
			UserID:   1,
			Role:     "admin",
			Status:   "active",
		},
		{
			TenantID: 1,
			UserID:   2,
			Role:     "user",
			Status:   "active",
		},
		{
			TenantID: 1,
			UserID:   3,
			Role:     "viewer",
			Status:   "inactive",
		},
	}

	for _, user := range users {
		err := tenantAdminService.CreateTenantAdmin(context.Background(), &user)
		require.NoError(t, err)
	}

	router := gin.New()
	router.GET("/tenants/:id/users", handler.ListTenantAdmins)

	// Test
	req, _ := http.NewRequest("GET", "/tenants/test-tenant-id/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Users []models.TenantAdmin `json:"users"`
		Total int                  `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Users), 3)
	assert.GreaterOrEqual(t, response.Total, 3)
}

func TestTenantAdminHandler_AddTenantUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, nil)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil)

	router := gin.New()
	router.POST("/tenants/:id/users", handler.CreateTenantAdmin)

	user := models.TenantAdmin{
		UserID: 1,
		Role:   "user",
		Status: "active",
	}

	jsonData, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/tenants/test-tenant-id/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.TenantAdmin
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test-tenant-id", response.TenantID)
	assert.Equal(t, "new-user-id", response.UserID)
	assert.Equal(t, "user", response.Role)
	assert.Equal(t, "active", response.Status)
	assert.NotEmpty(t, response.ID)
}

func TestTenantAdminHandler_UpdateTenantUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, nil)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil)

	// Create a tenant user first
	user := models.TenantAdmin{
		TenantID: 1,
		UserID:   1,
		Role:     "user",
		Status:   "active",
	}
	err := tenantAdminService.CreateTenantAdmin(context.Background(), &user)
	require.NoError(t, err)

	router := gin.New()
	router.PUT("/tenants/:id/users/:userId", handler.UpdateTenantAdmin)

	// Update data
	updateData := models.TenantAdmin{
		Role:   "admin",
		Status: "active",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/tenants/test-tenant-id/users/user-1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.TenantAdmin
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, user.ID, response.ID)
	assert.Equal(t, "admin", response.Role)
	assert.Equal(t, "active", response.Status)
}

func TestTenantAdminHandler_RemoveTenantUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, nil)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil)

	// Create a tenant user first
	user := models.TenantAdmin{
		TenantID: 1,
		UserID:   1,
		Role:     "user",
		Status:   "active",
	}
	err := tenantAdminService.CreateTenantAdmin(context.Background(), &user)
	require.NoError(t, err)

	router := gin.New()
	router.DELETE("/tenants/:id/users/:userId", handler.DeleteTenantAdmin)

	// Test
	req, _ := http.NewRequest("DELETE", "/tenants/test-tenant-id/users/user-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "User removed from tenant successfully", response["message"])

	// Verify deletion
	_, err = tenantAdminService.GetTenantAdmin(context.Background(), user.ID)
	assert.Error(t, err)
}

func TestTenantAdminHandler_GetTenantUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, nil)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil)

	router := gin.New()
	router.GET("/tenants/:id/usage", handler.GetTenantUsage)

	// Test
	req, _ := http.NewRequest("GET", "/tenants/test-tenant-id/usage", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.TenantUsage
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test-tenant-id", response.TenantID)
	assert.GreaterOrEqual(t, response.IncidentsCount, 0)
	assert.GreaterOrEqual(t, response.ServicesCount, 0)
	assert.GreaterOrEqual(t, response.UsersCount, 0)
	assert.GreaterOrEqual(t, response.APIRequests, 0)
}

func TestTenantAdminHandler_GetTenantBilling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, nil)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil)

	router := gin.New()
	router.GET("/tenants/:id/billing", handler.GetTenantBilling)

	// Test
	req, _ := http.NewRequest("GET", "/tenants/test-tenant-id/billing", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.TenantBilling
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test-tenant-id", response.TenantID)
	assert.Contains(t, response, "subscription")
	assert.Contains(t, response, "invoices")
	assert.Contains(t, response, "payment_methods")
}

func TestTenantAdminHandler_GetTenantFeatureFlags(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, nil)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil)

	// Create some tenant feature flags
	flags := []models.TenantFeatureFlag{
		{
			TenantID:  1,
			Name:      "feature1",
			IsEnabled: true,
		},
		{
			TenantID:  1,
			Name:      "feature2",
			IsEnabled: false,
		},
	}

	for _, flag := range flags {
		err := tenantAdminService.CreateTenantFeatureFlag(context.Background(), &flag)
		require.NoError(t, err)
	}

	router := gin.New()
	router.GET("/tenants/:id/feature-flags", handler.ListTenantFeatureFlags)

	// Test
	req, _ := http.NewRequest("GET", "/tenants/test-tenant-id/feature-flags", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		FeatureFlags []models.TenantFeatureFlag `json:"feature_flags"`
		Total        int                        `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.FeatureFlags), 2)
	assert.GreaterOrEqual(t, response.Total, 2)
}

func TestTenantAdminHandler_UpdateTenantFeatureFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, nil)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil)

	// Create a tenant feature flag first
	flag := models.TenantFeatureFlag{
		TenantID:  1,
		Name:      "test_feature",
		IsEnabled: false,
	}
	err := tenantAdminService.CreateTenantFeatureFlag(context.Background(), &flag)
	require.NoError(t, err)

	router := gin.New()
	router.PUT("/tenants/:id/feature-flags/:name", handler.UpdateTenantFeatureFlag)

	// Update data
	updateData := models.TenantFeatureFlag{
		IsEnabled: true,
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/tenants/test-tenant-id/feature-flags/test_feature", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.TenantFeatureFlag
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, flag.ID, response.ID)
	assert.Equal(t, true, response.IsEnabled)
	assert.Equal(t, "test_feature", response.Name)
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate
	err = db.AutoMigrate(&models.TenantSettings{}, &models.TenantAdmin{}, &models.TenantUsage{}, &models.TenantBilling{}, &models.TenantFeatureFlag{})
	require.NoError(t, err)

	return db
}
