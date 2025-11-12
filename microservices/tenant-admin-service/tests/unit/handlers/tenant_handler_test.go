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

	"tenant-admin-service/internal/handlers"
	"tenant-admin-service/internal/models"
)

// MockTenantService is a mock implementation of the tenant service
type MockTenantService struct {
	mock.Mock
}

func (m *MockTenantService) GetAllTenants() ([]models.Tenant, error) {
	args := m.Called()
	return args.Get(0).([]models.Tenant), args.Error(1)
}

func (m *MockTenantService) GetTenantByID(id string) (*models.Tenant, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantService) GetTenantBySubdomain(subdomain string) (*models.Tenant, error) {
	args := m.Called(subdomain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantService) CreateTenant(tenant *models.Tenant) error {
	args := m.Called(tenant)
	return args.Error(0)
}

func (m *MockTenantService) UpdateTenant(id string, tenant *models.Tenant) error {
	args := m.Called(id, tenant)
	return args.Error(0)
}

func (m *MockTenantService) DeleteTenant(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTenantService) CheckMaxUsers(tenantID string) (bool, error) {
	args := m.Called(tenantID)
	return args.Bool(0), args.Error(1)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestGetAllTenants_Success(t *testing.T) {
	// Arrange
	mockService := new(MockTenantService)
	handler := handlers.NewTenantHandler(mockService)
	router := setupTestRouter()
	router.GET("/api/v1/tenants", handler.GetAllTenants)

	expectedTenants := []models.Tenant{
		{
			ID:        "tenant-1",
			Name:      "Test Tenant 1",
			Subdomain: "test1",
			Status:    "active",
			MaxUsers:  5,
		},
		{
			ID:        "tenant-2",
			Name:      "Test Tenant 2",
			Subdomain: "test2",
			Status:    "active",
			MaxUsers:  10,
		},
	}

	mockService.On("GetAllTenants").Return(expectedTenants, nil)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Tenant
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "Test Tenant 1", response[0].Name)
	assert.Equal(t, "test1", response[0].Subdomain)

	mockService.AssertExpectations(t)
}

func TestGetTenantByID_Success(t *testing.T) {
	// Arrange
	mockService := new(MockTenantService)
	handler := handlers.NewTenantHandler(mockService)
	router := setupTestRouter()
	router.GET("/api/v1/tenants/:id", handler.GetTenantByID)

	expectedTenant := &models.Tenant{
		ID:        "tenant-1",
		Name:      "Test Tenant 1",
		Subdomain: "test1",
		PlanID:    "plan-1",
		Status:    "active",
		MaxUsers:  5,
	}

	mockService.On("GetTenantByID", "tenant-1").Return(expectedTenant, nil)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants/tenant-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Tenant
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "tenant-1", response.ID)
	assert.Equal(t, "Test Tenant 1", response.Name)
	assert.Equal(t, 5, response.MaxUsers)

	mockService.AssertExpectations(t)
}

func TestGetTenantByID_NotFound(t *testing.T) {
	// Arrange
	mockService := new(MockTenantService)
	handler := handlers.NewTenantHandler(mockService)
	router := setupTestRouter()
	router.GET("/api/v1/tenants/:id", handler.GetTenantByID)

	mockService.On("GetTenantByID", "nonexistent").Return(nil, assert.AnError)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetTenantBySubdomain_Success(t *testing.T) {
	// Arrange
	mockService := new(MockTenantService)
	handler := handlers.NewTenantHandler(mockService)
	router := setupTestRouter()
	router.GET("/api/v1/tenants/subdomain/:subdomain", handler.GetTenantBySubdomain)

	expectedTenant := &models.Tenant{
		ID:        "tenant-1",
		Name:      "Test Tenant 1",
		Subdomain: "test1",
		Status:    "active",
	}

	mockService.On("GetTenantBySubdomain", "test1").Return(expectedTenant, nil)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants/subdomain/test1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Tenant
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test1", response.Subdomain)

	mockService.AssertExpectations(t)
}

func TestCreateTenant_Success(t *testing.T) {
	// Arrange
	mockService := new(MockTenantService)
	handler := handlers.NewTenantHandler(mockService)
	router := setupTestRouter()
	router.POST("/api/v1/public/tenants", handler.CreateTenant)

	newTenant := models.Tenant{
		Name:      "New Tenant",
		Subdomain: "newtenant",
		Email:     "admin@newtenant.com",
	}

	mockService.On("CreateTenant", mock.AnythingOfType("*models.Tenant")).Return(nil)

	// Act
	body, _ := json.Marshal(newTenant)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/tenants", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreateTenant_InvalidJSON(t *testing.T) {
	// Arrange
	mockService := new(MockTenantService)
	handler := handlers.NewTenantHandler(mockService)
	router := setupTestRouter()
	router.POST("/api/v1/public/tenants", handler.CreateTenant)

	// Act
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/tenants", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateTenant_ValidationError(t *testing.T) {
	tests := []struct {
		name   string
		tenant models.Tenant
	}{
		{
			name: "Empty name",
			tenant: models.Tenant{
				Name:      "",
				Subdomain: "test",
				Email:     "test@example.com",
			},
		},
		{
			name: "Empty subdomain",
			tenant: models.Tenant{
				Name:      "Test",
				Subdomain: "",
				Email:     "test@example.com",
			},
		},
		{
			name: "Invalid email",
			tenant: models.Tenant{
				Name:      "Test",
				Subdomain: "test",
				Email:     "invalid-email",
			},
		},
		{
			name: "Invalid subdomain format",
			tenant: models.Tenant{
				Name:      "Test",
				Subdomain: "Test@123",
				Email:     "test@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := new(MockTenantService)
			handler := handlers.NewTenantHandler(mockService)
			router := setupTestRouter()
			router.POST("/api/v1/public/tenants", handler.CreateTenant)

			// Act
			body, _ := json.Marshal(tt.tenant)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/public/tenants", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestUpdateTenant_Success(t *testing.T) {
	// Arrange
	mockService := new(MockTenantService)
	handler := handlers.NewTenantHandler(mockService)
	router := setupTestRouter()
	router.PUT("/api/v1/tenants/:id", handler.UpdateTenant)

	updatedTenant := models.Tenant{
		Name:     "Updated Name",
		Status:   "active",
		MaxUsers: 10,
	}

	mockService.On("UpdateTenant", "tenant-1", mock.AnythingOfType("*models.Tenant")).Return(nil)

	// Act
	body, _ := json.Marshal(updatedTenant)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/tenants/tenant-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteTenant_Success(t *testing.T) {
	// Arrange
	mockService := new(MockTenantService)
	handler := handlers.NewTenantHandler(mockService)
	router := setupTestRouter()
	router.DELETE("/api/v1/tenants/:id", handler.DeleteTenant)

	mockService.On("DeleteTenant", "tenant-1").Return(nil)

	// Act
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tenants/tenant-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestCheckMaxUsers_UnderLimit(t *testing.T) {
	// Arrange
	mockService := new(MockTenantService)
	handler := handlers.NewTenantHandler(mockService)
	router := setupTestRouter()
	router.GET("/api/v1/tenants/:id/check-max-users", handler.CheckMaxUsers)

	mockService.On("CheckMaxUsers", "tenant-1").Return(false, nil) // false = under limit

	// Act
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants/tenant-1/check-max-users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["limit_exceeded"].(bool))

	mockService.AssertExpectations(t)
}

func TestCheckMaxUsers_LimitExceeded(t *testing.T) {
	// Arrange
	mockService := new(MockTenantService)
	handler := handlers.NewTenantHandler(mockService)
	router := setupTestRouter()
	router.GET("/api/v1/tenants/:id/check-max-users", handler.CheckMaxUsers)

	mockService.On("CheckMaxUsers", "tenant-1").Return(true, nil) // true = limit exceeded

	// Act
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants/tenant-1/check-max-users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["limit_exceeded"].(bool))

	mockService.AssertExpectations(t)
}
