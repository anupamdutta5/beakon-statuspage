package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enterprise-status/statuspage-tenant-service/internal/handlers"
	"github.com/enterprise-status/statuspage-tenant-service/internal/models"
	"github.com/enterprise-status/statuspage-tenant-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTenantHandler_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	tenantService := services.NewTenantService(db, nil)
	handler := handlers.NewTenantHandler(tenantService, nil)

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
	assert.Equal(t, "tenant-service", response["service"])
}

func TestTenantHandler_CreateTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	tenantService := services.NewTenantService(db, logger)
	handler := handlers.NewTenantHandler(tenantService, logger)

	router := gin.New()
	router.POST("/tenants", func(c *gin.Context) {
		// Set required context values for authentication
		c.Set("user_id", uint(1))
		handler.CreateTenant(c)
	})

	tenantRequest := map[string]interface{}{
		"name":          "Test Company",
		"slug":          "test-company",
		"domain":        "testcompany.com",
		"subdomain":     "testcompany",
		"contact_email": "admin@testcompany.com",
		"billing_email": "billing@testcompany.com",
		"plan":          "free",
	}

	jsonData, _ := json.Marshal(tenantRequest)
	req, _ := http.NewRequest("POST", "/tenants", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response struct {
		Message string        `json:"message"`
		Tenant  models.Tenant `json:"tenant"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Tenant created successfully", response.Message)
	assert.Equal(t, "Test Company", response.Tenant.Name)
	assert.Equal(t, "testcompany", response.Tenant.Subdomain)
	assert.Equal(t, "testcompany.com", response.Tenant.Domain)
	assert.NotZero(t, response.Tenant.ID)
}

func TestTenantHandler_GetTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	tenantService := services.NewTenantService(db, nil)
	handler := handlers.NewTenantHandler(tenantService, nil)

	// Create a tenant first
	tenant := models.Tenant{
		Name:      "Test Company",
		Subdomain: "testcompany",
		Domain:    "testcompany.com",
	}
	err := tenantService.CreateTenant(&tenant)
	require.NoError(t, err)

	router := gin.New()
	router.GET("/tenants/:id", handler.GetTenant)

	// Test
	req, _ := http.NewRequest("GET", fmt.Sprintf("/tenants/%d", tenant.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Tenant
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, tenant.ID, response.ID)
	assert.Equal(t, tenant.Name, response.Name)
}

func TestTenantHandler_UpdateTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	tenantService := services.NewTenantService(db, nil)
	handler := handlers.NewTenantHandler(tenantService, nil)

	// Create a tenant first
	tenant := models.Tenant{
		Name:      "Test Company",
		Subdomain: "testcompany",
		Domain:    "testcompany.com",
	}
	err := tenantService.CreateTenant(&tenant)
	require.NoError(t, err)

	router := gin.New()
	router.PUT("/tenants/:id", handler.UpdateTenant)

	// Update data
	updateData := models.Tenant{
		Name:      "Updated Company",
		Subdomain: "updatedcompany",
		Domain:    "updatedcompany.com",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/tenants/%d", tenant.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Tenant
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Updated Company", response.Name)
	assert.Equal(t, "updatedcompany", response.Subdomain)

	// Verify in database
	updatedTenant, err := tenantService.GetTenant(tenant.ID)
	require.NoError(t, err)

	assert.Equal(t, "Updated Company", updatedTenant.Name)
	assert.Equal(t, "updatedcompany", updatedTenant.Subdomain)
}

func TestTenantHandler_ListTenants(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	tenantService := services.NewTenantService(db, nil)
	handler := handlers.NewTenantHandler(tenantService, nil)

	// Create multiple tenants
	tenants := []models.Tenant{
		{Name: "Company 1", Subdomain: "company1", Domain: "company1.com"},
		{Name: "Company 2", Subdomain: "company2", Domain: "company2.com"},
		{Name: "Company 3", Subdomain: "company3", Domain: "company3.com"},
	}

	for _, tenant := range tenants {
		err := tenantService.CreateTenant(&tenant)
		require.NoError(t, err)
	}

	router := gin.New()
	router.GET("/tenants", handler.GetTenants)

	// Test
	req, _ := http.NewRequest("GET", "/tenants", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Tenants []models.Tenant `json:"tenants"`
		Total   int             `json:"total"`
		Limit   int             `json:"limit"`
		Offset  int             `json:"offset"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Tenants), 3)
	assert.GreaterOrEqual(t, response.Total, 3)
}

func TestTenantHandler_DeleteTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	tenantService := services.NewTenantService(db, nil)
	handler := handlers.NewTenantHandler(tenantService, nil)

	// Create a tenant first
	tenant := models.Tenant{
		Name:      "Test Company",
		Subdomain: "testcompany",
		Domain:    "testcompany.com",
	}
	err := tenantService.CreateTenant(&tenant)
	require.NoError(t, err)

	router := gin.New()
	router.DELETE("/tenants/:id", handler.DeleteTenant)

	// Test
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/tenants/%d", tenant.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Tenant deleted successfully", response["message"])

	// Verify deletion
	_, err = tenantService.GetTenant(tenant.ID)
	assert.Error(t, err)
}

func TestTenantHandler_GetTenantBySubdomain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	tenantService := services.NewTenantService(db, nil)
	handler := handlers.NewTenantHandler(tenantService, nil)

	// Create a tenant first
	tenant := models.Tenant{
		Name:      "Test Company",
		Subdomain: "testcompany",
		Domain:    "testcompany.com",
	}
	err := tenantService.CreateTenant(&tenant)
	require.NoError(t, err)

	router := gin.New()
	router.GET("/tenants/subdomain/:subdomain", handler.GetTenant)

	// Test
	req, _ := http.NewRequest("GET", "/tenants/subdomain/testcompany", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Tenant
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, tenant.ID, response.ID)
	assert.Equal(t, "testcompany", response.Subdomain)
}

func TestTenantHandler_UpdateTenantSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	tenantService := services.NewTenantService(db, nil)
	handler := handlers.NewTenantHandler(tenantService, nil)

	// Create a tenant first
	tenant := models.Tenant{
		Name:      "Test Company",
		Subdomain: "testcompany",
		Domain:    "testcompany.com",
		Settings:  `{"theme": "light"}`,
	}
	err := tenantService.CreateTenant(&tenant)
	require.NoError(t, err)

	router := gin.New()
	router.PUT("/tenants/:id/settings", handler.UpdateTenantSettings)

	// Update settings
	updateSettings := map[string]interface{}{
		"theme": "dark",
		"logo":  "https://example.com/new-logo.png",
	}

	jsonData, _ := json.Marshal(updateSettings)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/tenants/%d", tenant.ID)+"/settings", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Tenant
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Settings is a JSON string, so we need to parse it to check values
	var settings map[string]interface{}
	err = json.Unmarshal([]byte(response.Settings), &settings)
	require.NoError(t, err)
	assert.Equal(t, "dark", settings["theme"])
	assert.Equal(t, "https://example.com/new-logo.png", settings["logo"])
}

func TestTenantHandler_GetTenantStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	tenantService := services.NewTenantService(db, nil)
	handler := handlers.NewTenantHandler(tenantService, nil)

	// Create a tenant first
	tenant := models.Tenant{
		Name:      "Test Company",
		Subdomain: "testcompany",
		Domain:    "testcompany.com",
	}
	err := tenantService.CreateTenant(&tenant)
	require.NoError(t, err)

	router := gin.New()
	router.GET("/tenants/:id/stats", handler.GetTenant)

	// Test
	req, _ := http.NewRequest("GET", fmt.Sprintf("/tenants/%d", tenant.ID)+"/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "total_users")
	assert.Contains(t, response, "total_components")
	assert.Contains(t, response, "total_incidents")
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate
	err = db.AutoMigrate(&models.Tenant{}, &models.TenantActivity{}, &models.TenantSettings{}, &models.TenantBilling{}, &models.TenantBranding{})
	require.NoError(t, err)

	return db
}
