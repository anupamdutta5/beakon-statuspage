package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anupamdutta5/tenant-admin-service/internal/config"
	"github.com/anupamdutta5/tenant-admin-service/internal/handlers"
	"github.com/anupamdutta5/tenant-admin-service/internal/models"
	"github.com/anupamdutta5/tenant-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTenantAdminHandler_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, logger)
	tenantAdminService.SetDB(db)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
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
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, logger)
	tenantAdminService.SetDB(db)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil, logger)

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
		Version: "1.0.0",
		Status:  "active",
	}
	err := db.Create(&settings).Error
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/tenants/:tenant_id/settings", handler.GetTenantSettings)

	// Test
	req, _ := http.NewRequest("GET", "/tenants/1/settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseWrapper struct {
		Settings *models.TenantSettings `json:"settings"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)
	require.NotNil(t, responseWrapper.Settings)

	assert.Equal(t, uint(1), responseWrapper.Settings.TenantID)
	assert.Contains(t, responseWrapper.Settings.Settings, "notifications")
	assert.Contains(t, responseWrapper.Settings.Settings, "branding")
	assert.Contains(t, responseWrapper.Settings.Settings, "security")
}

func TestTenantAdminHandler_UpdateTenantSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, logger)
	tenantAdminService.SetDB(db)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil, logger)

	// Create tenant settings first
	settings := models.TenantSettings{
		TenantID: 1,
		Settings: `{
			"notifications": {
				"email_enabled": true,
				"sms_enabled":   false
			}
		}`,
		Version: "1.0.0",
		Status:  "active",
	}
	err := db.Create(&settings).Error
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.PUT("/tenants/:tenant_id/settings", handler.UpdateTenantSettings)

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
	req, _ := http.NewRequest("PUT", "/tenants/1/settings", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Tenant settings updated successfully", response["message"])
}

func TestTenantAdminHandler_GetTenantUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, logger)
	tenantAdminService.SetDB(db)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil, logger)

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
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/tenants/:tenant_id/users", handler.ListTenantAdmins)

	// Test
	req, _ := http.NewRequest("GET", "/tenants/1/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Admins []models.TenantAdmin `json:"admins"`
		Count  int                  `json:"count"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Admins), 3)
	assert.GreaterOrEqual(t, response.Count, 3)
}

func TestTenantAdminHandler_AddTenantUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, logger)
	tenantAdminService.SetDB(db)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.POST("/tenants/:tenant_id/users", handler.CreateTenantAdmin)

	user := models.TenantAdmin{
		UserID: 1,
		Role:   "user",
		Status: "active",
	}

	jsonData, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/tenants/1/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var responseWrapper struct {
		Admin   *models.TenantAdmin `json:"admin"`
		Message string              `json:"message"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)
	require.NotNil(t, responseWrapper.Admin)

	assert.Equal(t, uint(1), responseWrapper.Admin.TenantID)
	assert.Equal(t, uint(1), responseWrapper.Admin.UserID)
	assert.Equal(t, "user", responseWrapper.Admin.Role)
	assert.Equal(t, "active", responseWrapper.Admin.Status)
	assert.NotEmpty(t, responseWrapper.Admin.ID)
	assert.Equal(t, "Tenant admin created successfully", responseWrapper.Message)
}

func TestTenantAdminHandler_UpdateTenantUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, logger)
	tenantAdminService.SetDB(db)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil, logger)

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
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.PUT("/tenants/:tenant_id/users/:id", handler.UpdateTenantAdmin)

	// Update data
	updateData := models.TenantAdmin{
		Role:   "admin",
		Status: "active",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/tenants/1/users/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Tenant admin updated successfully", response["message"])
}

func TestTenantAdminHandler_RemoveTenantUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, logger)
	tenantAdminService.SetDB(db)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil, logger)

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
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.DELETE("/tenants/:tenant_id/users/:id", handler.DeleteTenantAdmin)

	// Test
	req, _ := http.NewRequest("DELETE", "/tenants/1/users/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Tenant admin deleted successfully", response["message"])

	// Verify deletion
	_, err = tenantAdminService.GetTenantAdmin(context.Background(), user.ID)
	assert.Error(t, err)
}

func TestTenantAdminHandler_GetTenantUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, logger)
	tenantAdminService.SetDB(db)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/tenants/:tenant_id/usage", handler.GetTenantUsage)

	// Test
	req, _ := http.NewRequest("GET", "/tenants/1/usage", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseWrapper struct {
		Usage []models.TenantUsage `json:"usage"`
		Count int                  `json:"count"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, responseWrapper.Count, 0)
	assert.GreaterOrEqual(t, len(responseWrapper.Usage), 0)
}

func TestTenantAdminHandler_GetTenantBilling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, logger)
	tenantAdminService.SetDB(db)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/tenants/:tenant_id/billing", handler.GetTenantBilling)

	// Test
	req, _ := http.NewRequest("GET", "/tenants/1/billing", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions - this endpoint is not implemented yet
	assert.Equal(t, http.StatusNotImplemented, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Not implemented", response["error"])
}

func TestTenantAdminHandler_GetTenantFeatureFlags(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, logger)
	tenantAdminService.SetDB(db)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil, logger)

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
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/tenants/:tenant_id/feature-flags", handler.ListTenantFeatureFlags)

	// Test
	req, _ := http.NewRequest("GET", "/tenants/1/feature-flags", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		FeatureFlags []models.TenantFeatureFlag `json:"feature_flags"`
		Count        int                        `json:"count"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.FeatureFlags), 2)
	assert.GreaterOrEqual(t, response.Count, 2)
}

func TestTenantAdminHandler_UpdateTenantFeatureFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	tenantAdminService, _ := services.NewTenantAdminService(cfg, logger)
	tenantAdminService.SetDB(db)
	handler := handlers.NewTenantAdminHandler(tenantAdminService, nil, logger)

	// Create a tenant feature flag first
	flag := models.TenantFeatureFlag{
		TenantID:  1,
		Name:      "test_feature",
		IsEnabled: false,
	}
	err := tenantAdminService.CreateTenantFeatureFlag(context.Background(), &flag)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.PUT("/tenants/:tenant_id/feature-flags/:id", handler.UpdateTenantFeatureFlag)

	// Update data
	updateData := models.TenantFeatureFlag{
		IsEnabled: true,
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/tenants/1/feature-flags/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Tenant feature flag updated successfully", response["message"])
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
