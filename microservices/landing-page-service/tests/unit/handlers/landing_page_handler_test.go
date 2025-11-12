package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"landing-page-service/internal/handlers"
	"landing-page-service/internal/models"
)

// MockLandingPageService is a mock implementation
type MockLandingPageService struct {
	mock.Mock
}

func (m *MockLandingPageService) GetPageBySlug(tenantID, slug string) (*models.LandingPage, error) {
	args := m.Called(tenantID, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LandingPage), args.Error(1)
}

func (m *MockLandingPageService) GetAllPages(tenantID string) ([]models.LandingPage, error) {
	args := m.Called(tenantID)
	return args.Get(0).([]models.LandingPage), args.Error(1)
}

func (m *MockLandingPageService) CreatePage(page *models.LandingPage) error {
	args := m.Called(page)
	return args.Error(0)
}

func (m *MockLandingPageService) UpdatePage(id string, page *models.LandingPage) error {
	args := m.Called(id, page)
	return args.Error(0)
}

func (m *MockLandingPageService) PublishPage(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockLandingPageService) UnpublishPage(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestGetLandingPage_Success(t *testing.T) {
	// Arrange
	mockService := new(MockLandingPageService)
	handler := handlers.NewLandingPageHandler(mockService)
	router := setupTestRouter()
	router.GET("/landing/:slug", handler.GetLandingPage)

	expectedPage := &models.LandingPage{
		ID:          "page-1",
		TenantID:    "tenant-1",
		PageName:    "Homepage",
		PageSlug:    "home",
		Title:       "Welcome to Our Platform",
		Subtitle:    "The best solution",
		IsPublished: true,
	}

	mockService.On("GetPageBySlug", "tenant-1", "home").Return(expectedPage, nil)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/landing/home", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.LandingPage
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Homepage", response.PageName)
	assert.Equal(t, "home", response.PageSlug)
	assert.True(t, response.IsPublished)

	mockService.AssertExpectations(t)
}

func TestGetLandingPage_NotPublished(t *testing.T) {
	// Arrange
	mockService := new(MockLandingPageService)
	handler := handlers.NewLandingPageHandler(mockService)
	router := setupTestRouter()
	router.GET("/landing/:slug", handler.GetLandingPage)

	unpublishedPage := &models.LandingPage{
		ID:          "page-1",
		TenantID:    "tenant-1",
		PageName:    "Draft Page",
		PageSlug:    "draft",
		IsPublished: false,
	}

	mockService.On("GetPageBySlug", "tenant-1", "draft").Return(unpublishedPage, nil)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/landing/draft", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "page is not published")
	mockService.AssertExpectations(t)
}

func TestGetAllLandingPages_Success(t *testing.T) {
	// Arrange
	mockService := new(MockLandingPageService)
	handler := handlers.NewLandingPageHandler(mockService)
	router := setupTestRouter()
	router.GET("/api/v1/landing-pages", handler.GetAllLandingPages)

	expectedPages := []models.LandingPage{
		{
			ID:          "page-1",
			TenantID:    "tenant-1",
			PageName:    "Homepage",
			PageSlug:    "home",
			IsPublished: true,
		},
		{
			ID:          "page-2",
			TenantID:    "tenant-1",
			PageName:    "Pricing",
			PageSlug:    "pricing",
			IsPublished: true,
		},
	}

	mockService.On("GetAllPages", "tenant-1").Return(expectedPages, nil)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/api/v1/landing-pages", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.LandingPage
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "Homepage", response[0].PageName)
	assert.Equal(t, "Pricing", response[1].PageName)

	mockService.AssertExpectations(t)
}

func TestCreateLandingPage_Success(t *testing.T) {
	// Arrange
	mockService := new(MockLandingPageService)
	handler := handlers.NewLandingPageHandler(mockService)
	router := setupTestRouter()
	router.POST("/api/v1/landing-pages", handler.CreateLandingPage)

	newPage := models.LandingPage{
		TenantID:     "tenant-1",
		PageName:     "New Page",
		PageSlug:     "new-page",
		Title:        "New Page Title",
		Subtitle:     "New Page Subtitle",
		TemplateName: "modern",
	}

	mockService.On("CreatePage", mock.AnythingOfType("*models.LandingPage")).Return(nil)

	// Act
	body, _ := json.Marshal(newPage)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/landing-pages", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "tenant-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreateLandingPage_ValidationError(t *testing.T) {
	tests := []struct {
		name string
		page models.LandingPage
	}{
		{
			name: "Empty page name",
			page: models.LandingPage{
				TenantID: "tenant-1",
				PageName: "",
				PageSlug: "test",
			},
		},
		{
			name: "Empty page slug",
			page: models.LandingPage{
				TenantID: "tenant-1",
				PageName: "Test",
				PageSlug: "",
			},
		},
		{
			name: "Invalid slug format",
			page: models.LandingPage{
				TenantID: "tenant-1",
				PageName: "Test",
				PageSlug: "Test Page!",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := new(MockLandingPageService)
			handler := handlers.NewLandingPageHandler(mockService)
			router := setupTestRouter()
			router.POST("/api/v1/landing-pages", handler.CreateLandingPage)

			// Act
			body, _ := json.Marshal(tt.page)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/landing-pages", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Tenant-ID", "tenant-1")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestUpdateLandingPage_Success(t *testing.T) {
	// Arrange
	mockService := new(MockLandingPageService)
	handler := handlers.NewLandingPageHandler(mockService)
	router := setupTestRouter()
	router.PUT("/api/v1/landing-pages/:id", handler.UpdateLandingPage)

	updatedPage := models.LandingPage{
		PageName: "Updated Page",
		Title:    "Updated Title",
		Subtitle: "Updated Subtitle",
	}

	mockService.On("UpdatePage", "page-1", mock.AnythingOfType("*models.LandingPage")).Return(nil)

	// Act
	body, _ := json.Marshal(updatedPage)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/landing-pages/page-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "tenant-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPublishLandingPage_Success(t *testing.T) {
	// Arrange
	mockService := new(MockLandingPageService)
	handler := handlers.NewLandingPageHandler(mockService)
	router := setupTestRouter()
	router.POST("/api/v1/landing-pages/:id/publish", handler.PublishLandingPage)

	mockService.On("PublishPage", "page-1").Return(nil)

	// Act
	req := httptest.NewRequest(http.MethodPost, "/api/v1/landing-pages/page-1/publish", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUnpublishLandingPage_Success(t *testing.T) {
	// Arrange
	mockService := new(MockLandingPageService)
	handler := handlers.NewLandingPageHandler(mockService)
	router := setupTestRouter()
	router.POST("/api/v1/landing-pages/:id/unpublish", handler.UnpublishLandingPage)

	mockService.On("UnpublishPage", "page-1").Return(nil)

	// Act
	req := httptest.NewRequest(http.MethodPost, "/api/v1/landing-pages/page-1/unpublish", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetLandingPage_WithSections(t *testing.T) {
	// Arrange
	mockService := new(MockLandingPageService)
	handler := handlers.NewLandingPageHandler(mockService)
	router := setupTestRouter()
	router.GET("/landing/:slug", handler.GetLandingPage)

	pageWithSections := &models.LandingPage{
		ID:          "page-1",
		TenantID:    "tenant-1",
		PageName:    "Homepage",
		PageSlug:    "home",
		IsPublished: true,
		Sections: []models.Section{
			{
				ID:           "section-1",
				SectionType:  "hero",
				SectionName:  "Hero Banner",
				DisplayOrder: 1,
				IsVisible:    true,
			},
			{
				ID:           "section-2",
				SectionType:  "features",
				SectionName:  "Key Features",
				DisplayOrder: 2,
				IsVisible:    true,
			},
		},
	}

	mockService.On("GetPageBySlug", "tenant-1", "home").Return(pageWithSections, nil)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/landing/home", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.LandingPage
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Sections, 2)
	assert.Equal(t, "hero", response.Sections[0].SectionType)
	assert.Equal(t, "features", response.Sections[1].SectionType)

	mockService.AssertExpectations(t)
}

func TestGetLandingPage_WithForms(t *testing.T) {
	// Arrange
	mockService := new(MockLandingPageService)
	handler := handlers.NewLandingPageHandler(mockService)
	router := setupTestRouter()
	router.GET("/landing/:slug", handler.GetLandingPage)

	pageWithForms := &models.LandingPage{
		ID:          "page-1",
		TenantID:    "tenant-1",
		PageName:    "Homepage",
		PageSlug:    "home",
		IsPublished: true,
		Forms: []models.Form{
			{
				ID:       "form-1",
				FormName: "Contact Us",
				FormType: "contact",
				IsActive: true,
			},
			{
				ID:       "form-2",
				FormName: "Newsletter",
				FormType: "newsletter",
				IsActive: true,
			},
		},
	}

	mockService.On("GetPageBySlug", "tenant-1", "home").Return(pageWithForms, nil)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/landing/home", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.LandingPage
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Forms, 2)
	assert.Equal(t, "contact", response.Forms[0].FormType)
	assert.Equal(t, "newsletter", response.Forms[1].FormType)

	mockService.AssertExpectations(t)
}
