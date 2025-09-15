package integration

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

	"github.com/enterprise-status/statuspage-branding-service/internal/config"
	"github.com/enterprise-status/statuspage-branding-service/internal/handlers"
	"github.com/enterprise-status/statuspage-branding-service/internal/models"
	"github.com/enterprise-status/statuspage-branding-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	db              *gorm.DB
	brandingService *services.BrandingService
	handler         *handlers.BrandingHandler
	router          *gin.Engine
)

func setupIntegrationTest() error {
	// Use in-memory SQLite database for integration tests
	var err error
	db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate
	err = db.AutoMigrate(&models.Brand{}, &models.Asset{}, &models.Theme{}, &models.CustomCSS{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Initialize services
	cfg := &config.Config{}
	logger, _ := zap.NewDevelopment()
	brandingService, _ = services.NewBrandingService(cfg, logger)
	brandingService.SetDB(db) // Set the test database
	handler = handlers.NewBrandingHandler(brandingService, logger)

	// Setup router
	gin.SetMode(gin.TestMode)
	router = gin.New()
	setupRoutes()

	return nil
}

func teardownIntegrationTest() {
	if db != nil {
		// Clean up test data
		db.Exec("DELETE FROM assets")
		db.Exec("DELETE FROM brands")
	}
}

func setupRoutes() {
	// Health check
	router.GET("/health", handler.HealthCheck)

	// Brand routes
	router.POST("/brands", handler.CreateBrand)
	router.GET("/brands/:id", handler.GetBrand)
	router.PUT("/brands/:id", handler.UpdateBrand)
	router.DELETE("/brands/:id", handler.DeleteBrand)
	router.GET("/brands", handler.ListBrands)

	// Asset routes
	router.POST("/assets/upload", handler.UploadAsset)
	router.GET("/assets", handler.ListAssets)
	router.DELETE("/assets/:id", handler.DeleteAsset)

	// Stats routes
	router.GET("/stats", handler.GetStats)
}

func TestBrandingServiceHealth(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Test health check
	req, _ := http.NewRequest("GET", "/health", nil)
	w := performRequest(req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "branding-service", response["service"])
}

func TestBrandingCRUDIntegration(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Test Create Brand
	brand := models.Brand{
		TenantID:    1,
		Name:        "Integration Test Brand",
		Slug:        "integration-test-brand",
		Description: "Integration test brand description",
		Status:      "active",
		IsDefault:   false,
		Metadata: `{
			"logo": "https://example.com/logo.png",
			"favicon": "https://example.com/favicon.ico",
			"primary_color": "#007bff",
			"secondary_color": "#6c757d",
			"font_family": "Inter",
			"custom_css": ".custom { color: red; }",
			"show_powered_by": false,
			"custom_domain": "status.example.com"
		}`,
	}

	jsonData, _ := json.Marshal(brand)
	req, _ := http.NewRequest("POST", "/brands", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := performRequest(req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var responseWrapper struct {
		Brand   models.Brand `json:"brand"`
		Message string       `json:"message"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	createdBrand := responseWrapper.Brand
	assert.Equal(t, brand.TenantID, createdBrand.TenantID)
	assert.Equal(t, brand.Name, createdBrand.Name)
	assert.Equal(t, brand.Slug, createdBrand.Slug)
	assert.NotEmpty(t, createdBrand.ID)

	// Test Get Brand
	req, _ = http.NewRequest("GET", fmt.Sprintf("/brands/%d", createdBrand.ID), nil)
	w = performRequest(req)

	assert.Equal(t, http.StatusOK, w.Code)

	var getResponseWrapper struct {
		Brand models.Brand `json:"brand"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &getResponseWrapper)
	require.NoError(t, err)

	retrievedBrand := getResponseWrapper.Brand

	assert.Equal(t, createdBrand.ID, retrievedBrand.ID)
	assert.Equal(t, createdBrand.Name, retrievedBrand.Name)

	// Test Update Brand
	updateData := models.Brand{
		Name:        "Updated Integration Test Brand",
		Slug:        "updated-integration-test-brand",
		Description: "Updated integration test brand description",
		Status:      "active",
		IsDefault:   false,
		Metadata: `{
			"logo": "https://example.com/new-logo.png",
			"primary_color": "#28a745",
			"secondary_color": "#dc3545",
			"custom_css": ".updated { color: blue; }"
		}`,
	}

	jsonData, _ = json.Marshal(updateData)
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/brands/%d", createdBrand.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = performRequest(req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updateResponseWrapper struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &updateResponseWrapper)
	require.NoError(t, err)

	assert.Equal(t, "Brand updated successfully", updateResponseWrapper.Message)

	// Test List Brands
	req, _ = http.NewRequest("GET", "/brands?tenant_id=test-tenant-id", nil)
	w = performRequest(req)

	assert.Equal(t, http.StatusOK, w.Code)

	var listResponse struct {
		Brands []models.Brand `json:"brands"`
		Count  int            `json:"count"`
		Limit  int            `json:"limit"`
		Offset int            `json:"offset"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &listResponse)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(listResponse.Brands), 1)
	assert.GreaterOrEqual(t, listResponse.Count, 1)

	// Test Delete Brand
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/brands/%d", createdBrand.ID), nil)
	w = performRequest(req)

	assert.Equal(t, http.StatusOK, w.Code)

	var deleteResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &deleteResponse)
	require.NoError(t, err)

	assert.Equal(t, "Brand deleted successfully", deleteResponse["message"])

	// Verify deletion
	req, _ = http.NewRequest("GET", fmt.Sprintf("/brands/%d", createdBrand.ID), nil)
	w = performRequest(req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAssetManagementIntegration(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Test Create Asset
	asset := models.Asset{
		BrandID:      1,
		Type:         "logo",
		Category:     "primary",
		Name:         "Integration Test Logo",
		Filename:     "logo.png",
		OriginalName: "logo.png",
		URL:          "https://example.com/logo.png",
		Description:  "Test logo for integration testing",
		Size:         1024,
		MimeType:     "image/png",
	}

	jsonData, _ := json.Marshal(asset)
	req, _ := http.NewRequest("POST", "/assets", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := performRequest(req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createdAsset models.Asset
	err = json.Unmarshal(w.Body.Bytes(), &createdAsset)
	require.NoError(t, err)

	assert.Equal(t, asset.BrandID, createdAsset.BrandID)
	assert.Equal(t, asset.Type, createdAsset.Type)
	assert.Equal(t, asset.Name, createdAsset.Name)
	assert.NotEmpty(t, createdAsset.ID)

	// Test Get Assets
	req, _ = http.NewRequest("GET", "/assets?tenant_id=test-tenant-id", nil)
	w = performRequest(req)

	assert.Equal(t, http.StatusOK, w.Code)

	var assetsResponse struct {
		Assets []models.Asset `json:"assets"`
		Total  int            `json:"total"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &assetsResponse)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(assetsResponse.Assets), 1)
	assert.GreaterOrEqual(t, assetsResponse.Total, 1)

	// Test Delete Asset
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/assets/%d", createdAsset.ID), nil)
	w = performRequest(req)

	assert.Equal(t, http.StatusOK, w.Code)

	var deleteResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &deleteResponse)
	require.NoError(t, err)

	assert.Equal(t, "Asset deleted successfully", deleteResponse["message"])
}

func TestBrandingSettingsIntegration(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Create a brand first
	brand := models.Brand{
		TenantID: 1,
		Name:     "Settings Test Brand",
		// These fields are now in Metadata
	}
	err = brandingService.CreateBrand(context.Background(), &brand)
	require.NoError(t, err)

	// Test Get Branding Settings
	req, _ := http.NewRequest("GET", "/branding/settings?tenant_id=test-tenant-id", nil)
	w := performRequest(req)

	assert.Equal(t, http.StatusOK, w.Code)

	var settingsResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &settingsResponse)
	require.NoError(t, err)

	assert.Contains(t, settingsResponse, "brand")
	assert.Contains(t, settingsResponse, "assets")
	assert.Contains(t, settingsResponse, "settings")

	// Test Update Branding Settings
	settings := map[string]interface{}{
		"show_powered_by":   false,
		"custom_domain":     "status.example.com",
		"analytics_enabled": true,
		"social_links": map[string]interface{}{
			"twitter":  "https://twitter.com/example",
			"linkedin": "https://linkedin.com/company/example",
		},
		"theme": map[string]interface{}{
			"primary_color":   "#28a745",
			"secondary_color": "#dc3545",
			"font_family":     "Roboto",
		},
	}

	jsonData, _ := json.Marshal(settings)
	req, _ = http.NewRequest("PUT", "/branding/settings?tenant_id=test-tenant-id", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = performRequest(req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updateResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &updateResponse)
	require.NoError(t, err)

	assert.Equal(t, "Branding settings updated successfully", updateResponse["message"])
}

func TestBrandingServiceConcurrency(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Test concurrent brand creation
	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer func() { done <- true }()

			brand := models.Brand{
				TenantID:    uint(index + 1),
				Name:        fmt.Sprintf("Concurrent Brand %d", index),
				Slug:        fmt.Sprintf("concurrent-brand-%d", index),
				Description: fmt.Sprintf("Concurrent brand %d for testing", index),
			}

			jsonData, _ := json.Marshal(brand)
			req, _ := http.NewRequest("POST", "/brands", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := performRequest(req)

			assert.Equal(t, http.StatusCreated, w.Code)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all brands were created
	req, _ := http.NewRequest("GET", "/brands", nil)
	w := performRequest(req)

	assert.Equal(t, http.StatusOK, w.Code)

	var listResponse struct {
		Brands []models.Brand `json:"brands"`
		Total  int            `json:"total"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &listResponse)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, listResponse.Total, numGoroutines)
}

func TestBrandingServiceErrorHandling(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Test invalid brand creation
	invalidBrand := map[string]interface{}{
		"name": "", // Invalid: empty name
		"logo": "invalid-url",
	}

	jsonData, _ := json.Marshal(invalidBrand)
	req, _ := http.NewRequest("POST", "/brands", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := performRequest(req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test non-existent brand retrieval
	req, _ = http.NewRequest("GET", "/brands/non-existent-id", nil)
	w = performRequest(req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test non-existent brand update
	updateData := map[string]interface{}{
		"name": "Updated Name",
	}

	jsonData, _ = json.Marshal(updateData)
	req, _ = http.NewRequest("PUT", "/brands/non-existent-id", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = performRequest(req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test non-existent brand deletion
	req, _ = http.NewRequest("DELETE", "/brands/non-existent-id", nil)
	w = performRequest(req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func performRequest(req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestMain(m *testing.M) {
	// Wait for database to be ready
	time.Sleep(5 * time.Second)

	// Run tests
	code := m.Run()

	// Cleanup
	teardownIntegrationTest()

	// Exit with test result code
	os.Exit(code)
}
