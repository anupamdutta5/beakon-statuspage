// Package unit provides unit tests for the Branding Service.
package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enterprise-status/statuspage-branding-service/internal/config"
	"github.com/enterprise-status/statuspage-branding-service/internal/handlers"
	"github.com/enterprise-status/statuspage-branding-service/internal/models"
	"github.com/enterprise-status/statuspage-branding-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBrandingHandler_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	brandingService, _ := services.NewBrandingService(cfg, nil)
	handler := handlers.NewBrandingHandler(brandingService, nil)

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
	assert.Equal(t, "branding-service", response["service"])
}

func TestBrandingHandler_CreateBrand(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	brandingService, _ := services.NewBrandingService(cfg, nil)
	handler := handlers.NewBrandingHandler(brandingService, nil)

	router := gin.New()
	router.POST("/brands", handler.CreateBrand)

	// Test data
	brand := models.Brand{
		TenantID:    1,
		Name:        "Test Brand",
		Slug:        "test-brand",
		Description: "Test brand description",
		Status:      "active",
		IsDefault:   false,
		Metadata:    `{"source": "test"}`,
	}

	jsonData, _ := json.Marshal(brand)
	req, _ := http.NewRequest("POST", "/brands", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Brand
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, brand.Name, response.Name)
	assert.Equal(t, brand.Slug, response.Slug)
	assert.Equal(t, brand.Description, response.Description)
	assert.Equal(t, brand.Status, response.Status)
	assert.Equal(t, brand.TenantID, response.TenantID)
}

func TestBrandingHandler_GetBrand(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	brandingService, _ := services.NewBrandingService(cfg, nil)
	handler := handlers.NewBrandingHandler(brandingService, nil)

	// Create a brand first
	brand := models.Brand{
		TenantID:    1,
		Name:        "Test Brand",
		Slug:        "test-brand",
		Description: "Test brand description",
		Status:      "active",
		IsDefault:   false,
		Metadata:    `{"source": "test"}`,
	}
	err := brandingService.CreateBrand(context.Background(), &brand)
	require.NoError(t, err)

	router := gin.New()
	router.GET("/brands/:id", handler.GetBrand)

	// Test
	req, _ := http.NewRequest("GET", fmt.Sprintf("/brands/%d", brand.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Brand
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, brand.Name, response.Name)
	assert.Equal(t, brand.Slug, response.Slug)
	assert.Equal(t, brand.Description, response.Description)
	assert.Equal(t, brand.Status, response.Status)
}

func TestBrandingHandler_UpdateBrand(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	brandingService, _ := services.NewBrandingService(cfg, nil)
	handler := handlers.NewBrandingHandler(brandingService, nil)

	// Create a brand first
	brand := models.Brand{
		TenantID:    1,
		Name:        "Original Brand",
		Slug:        "original-brand",
		Description: "Original description",
		Status:      "active",
		IsDefault:   false,
		Metadata:    `{"source": "test"}`,
	}
	err := brandingService.CreateBrand(context.Background(), &brand)
	require.NoError(t, err)

	router := gin.New()
	router.PUT("/brands/:id", handler.UpdateBrand)

	// Update data
	updateData := models.Brand{
		Name:        "Updated Brand",
		Slug:        "updated-brand",
		Description: "Updated description",
		Status:      "inactive",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/brands/%d", brand.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Brand
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Updated Brand", response.Name)
	assert.Equal(t, "updated-brand", response.Slug)
	assert.Equal(t, "Updated description", response.Description)
	assert.Equal(t, "inactive", response.Status)
}

func TestBrandingHandler_DeleteBrand(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	brandingService, _ := services.NewBrandingService(cfg, nil)
	handler := handlers.NewBrandingHandler(brandingService, nil)

	// Create a brand first
	brand := models.Brand{
		TenantID:    1,
		Name:        "Test Brand",
		Slug:        "test-brand",
		Description: "Test brand description",
		Status:      "active",
		IsDefault:   false,
		Metadata:    `{"source": "test"}`,
	}
	err := brandingService.CreateBrand(context.Background(), &brand)
	require.NoError(t, err)

	router := gin.New()
	router.DELETE("/brands/:id", handler.DeleteBrand)

	// Test
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/brands/%d", brand.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify deletion
	_, err = brandingService.GetBrand(context.Background(), brand.ID)
	assert.Error(t, err)
}

func TestBrandingHandler_ListBrands(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	brandingService, _ := services.NewBrandingService(cfg, nil)
	handler := handlers.NewBrandingHandler(brandingService, nil)

	// Create multiple brands
	brands := []models.Brand{
		{TenantID: 1, Name: "Brand 1", Slug: "brand-1", Description: "Description 1", Status: "active"},
		{TenantID: 1, Name: "Brand 2", Slug: "brand-2", Description: "Description 2", Status: "active"},
		{TenantID: 1, Name: "Brand 3", Slug: "brand-3", Description: "Description 3", Status: "inactive"},
	}

	for _, brand := range brands {
		err := brandingService.CreateBrand(context.Background(), &brand)
		require.NoError(t, err)
	}

	router := gin.New()
	router.GET("/brands", handler.ListBrands)

	// Test
	req, _ := http.NewRequest("GET", "/brands?tenant_id=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	brandsList, ok := response["brands"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(brandsList))
}

func TestBrandingHandler_CreateTheme(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	brandingService, _ := services.NewBrandingService(cfg, nil)
	handler := handlers.NewBrandingHandler(brandingService, nil)

	// Create a brand first
	brand := models.Brand{
		TenantID:    1,
		Name:        "Test Brand",
		Slug:        "test-brand",
		Description: "Test brand description",
		Status:      "active",
		IsDefault:   false,
		Metadata:    `{"source": "test"}`,
	}
	err := brandingService.CreateBrand(context.Background(), &brand)
	require.NoError(t, err)

	router := gin.New()
	router.POST("/themes", handler.CreateTheme)

	// Test data
	theme := models.Theme{
		BrandID:     brand.ID,
		Name:        "Test Theme",
		Slug:        "test-theme",
		Description: "Test theme description",
		Version:     "1.0.0",
		Status:      "active",
		IsDefault:   false,
		IsPublic:    true,
		PreviewURL:  "https://example.com/preview",
		Metadata:    `{"source": "test"}`,
	}

	jsonData, _ := json.Marshal(theme)
	req, _ := http.NewRequest("POST", "/themes", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Theme
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, theme.Name, response.Name)
	assert.Equal(t, theme.Slug, response.Slug)
	assert.Equal(t, theme.Description, response.Description)
	assert.Equal(t, theme.Version, response.Version)
	assert.Equal(t, theme.Status, response.Status)
	assert.Equal(t, theme.BrandID, response.BrandID)
}

func TestBrandingHandler_CreateAsset(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	brandingService, _ := services.NewBrandingService(cfg, nil)
	handler := handlers.NewBrandingHandler(brandingService, nil)

	// Create a brand first
	brand := models.Brand{
		TenantID:    1,
		Name:        "Test Brand",
		Slug:        "test-brand",
		Description: "Test brand description",
		Status:      "active",
		IsDefault:   false,
		Metadata:    `{"source": "test"}`,
	}
	err := brandingService.CreateBrand(context.Background(), &brand)
	require.NoError(t, err)

	router := gin.New()
	router.POST("/assets", handler.CreateAsset)

	// Test data
	asset := models.Asset{
		BrandID:      brand.ID,
		Name:         "Test Logo",
		Type:         "logo",
		Category:     "primary",
		Filename:     "logo.png",
		OriginalName: "test-logo.png",
		MimeType:     "image/png",
		Size:         1024,
		Width:        200,
		Height:       100,
		URL:          "https://example.com/logo.png",
		AltText:      "Test Logo",
		Description:  "Test logo asset",
		Status:       "active",
		IsDefault:    false,
		Metadata:     `{"source": "test"}`,
	}

	jsonData, _ := json.Marshal(asset)
	req, _ := http.NewRequest("POST", "/assets", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Asset
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, asset.Name, response.Name)
	assert.Equal(t, asset.Type, response.Type)
	assert.Equal(t, asset.Category, response.Category)
	assert.Equal(t, asset.Filename, response.Filename)
	assert.Equal(t, asset.OriginalName, response.OriginalName)
	assert.Equal(t, asset.MimeType, response.MimeType)
	assert.Equal(t, asset.Size, response.Size)
	assert.Equal(t, asset.BrandID, response.BrandID)
}

func TestBrandingHandler_ListAssets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	brandingService, _ := services.NewBrandingService(cfg, nil)
	handler := handlers.NewBrandingHandler(brandingService, nil)

	// Create a brand first
	brand := models.Brand{
		TenantID:    1,
		Name:        "Test Brand",
		Slug:        "test-brand",
		Description: "Test brand description",
		Status:      "active",
		IsDefault:   false,
		Metadata:    `{"source": "test"}`,
	}
	err := brandingService.CreateBrand(context.Background(), &brand)
	require.NoError(t, err)

	// Create multiple assets
	assets := []models.Asset{
		{BrandID: brand.ID, Name: "Logo 1", Type: "logo", Category: "primary", Filename: "logo1.png", OriginalName: "logo1.png", MimeType: "image/png", Size: 1024, URL: "https://example.com/logo1.png"},
		{BrandID: brand.ID, Name: "Favicon 1", Type: "favicon", Category: "utility", Filename: "favicon1.ico", OriginalName: "favicon1.ico", MimeType: "image/x-icon", Size: 512, URL: "https://example.com/favicon1.ico"},
		{BrandID: brand.ID, Name: "Background 1", Type: "background", Category: "secondary", Filename: "bg1.jpg", OriginalName: "bg1.jpg", MimeType: "image/jpeg", Size: 2048, URL: "https://example.com/bg1.jpg"},
	}

	for _, asset := range assets {
		err := brandingService.CreateAsset(context.Background(), &asset)
		require.NoError(t, err)
	}

	router := gin.New()
	router.GET("/assets", handler.ListAssets)

	// Test
	req, _ := http.NewRequest("GET", fmt.Sprintf("/assets?brand_id=%d", brand.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assetsList, ok := response["assets"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(assetsList))
}

func TestBrandingHandler_CreateCustomCSS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	brandingService, _ := services.NewBrandingService(cfg, nil)
	handler := handlers.NewBrandingHandler(brandingService, nil)

	// Create a brand first
	brand := models.Brand{
		TenantID:    1,
		Name:        "Test Brand",
		Slug:        "test-brand",
		Description: "Test brand description",
		Status:      "active",
		IsDefault:   false,
		Metadata:    `{"source": "test"}`,
	}
	err := brandingService.CreateBrand(context.Background(), &brand)
	require.NoError(t, err)

	router := gin.New()
	router.POST("/custom-css", handler.CreateCustomCSS)

	// Test data
	customCSS := models.CustomCSS{
		BrandID:     brand.ID,
		Name:        "Test CSS",
		Description: "Test CSS description",
		CSS:         ".test { color: red; }",
		Version:     "1.0.0",
		Status:      "active",
		IsMinified:  false,
		Metadata:    `{"source": "test"}`,
	}

	jsonData, _ := json.Marshal(customCSS)
	req, _ := http.NewRequest("POST", "/custom-css", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.CustomCSS
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, customCSS.Name, response.Name)
	assert.Equal(t, customCSS.Description, response.Description)
	assert.Equal(t, customCSS.CSS, response.CSS)
	assert.Equal(t, customCSS.Version, response.Version)
	assert.Equal(t, customCSS.Status, response.Status)
	assert.Equal(t, customCSS.BrandID, response.BrandID)
}

func TestBrandingHandler_GetStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	brandingService, _ := services.NewBrandingService(cfg, nil)
	handler := handlers.NewBrandingHandler(brandingService, nil)

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

	assert.Contains(t, response, "total_brands")
	assert.Contains(t, response, "total_themes")
	assert.Contains(t, response, "total_assets")
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.Brand{},
		&models.Theme{},
		&models.ColorScheme{},
		&models.Typography{},
		&models.Asset{},
		&models.CustomCSS{},
		&models.CustomJS{},
		&models.Layout{},
		&models.Component{},
		&models.BrandingStats{},
	)
	require.NoError(t, err)

	return db
}
