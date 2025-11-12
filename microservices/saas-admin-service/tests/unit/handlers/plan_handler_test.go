package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"saas-admin-service/internal/handlers"
	"saas-admin-service/internal/models"
)

// MockPlanService is a mock implementation of the plan service
type MockPlanService struct {
	mock.Mock
}

func (m *MockPlanService) GetAllPlans() ([]models.Plan, error) {
	args := m.Called()
	return args.Get(0).([]models.Plan), args.Error(1)
}

func (m *MockPlanService) GetPlanByID(id string) (*models.Plan, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Plan), args.Error(1)
}

func (m *MockPlanService) CreatePlan(plan *models.Plan) error {
	args := m.Called(plan)
	return args.Error(0)
}

func (m *MockPlanService) UpdatePlan(id string, plan *models.Plan) error {
	args := m.Called(id, plan)
	return args.Error(0)
}

func (m *MockPlanService) DeletePlan(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestGetAllPlans_Success(t *testing.T) {
	// Arrange
	mockService := new(MockPlanService)
	handler := handlers.NewPlanHandler(mockService)
	router := setupTestRouter()
	router.GET("/api/v1/plans", handler.GetAllPlans)

	now := time.Now()
	expectedPlans := []models.Plan{
		{
			ID:          "plan-1",
			Name:        "Free Plan",
			PlanKey:     "free",
			Description: "Free forever",
			Status:      "active",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "plan-2",
			Name:        "Pro Plan",
			PlanKey:     "pro",
			Description: "Professional features",
			Status:      "active",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	mockService.On("GetAllPlans").Return(expectedPlans, nil)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/api/v1/plans", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Plan
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "Free Plan", response[0].Name)
	assert.Equal(t, "Pro Plan", response[1].Name)

	mockService.AssertExpectations(t)
}

func TestGetAllPlans_ServiceError(t *testing.T) {
	// Arrange
	mockService := new(MockPlanService)
	handler := handlers.NewPlanHandler(mockService)
	router := setupTestRouter()
	router.GET("/api/v1/plans", handler.GetAllPlans)

	mockService.On("GetAllPlans").Return([]models.Plan{}, assert.AnError)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/api/v1/plans", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetPlanByID_Success(t *testing.T) {
	// Arrange
	mockService := new(MockPlanService)
	handler := handlers.NewPlanHandler(mockService)
	router := setupTestRouter()
	router.GET("/api/v1/plans/:id", handler.GetPlanByID)

	now := time.Now()
	expectedPlan := &models.Plan{
		ID:          "plan-1",
		Name:        "Free Plan",
		PlanKey:     "free",
		Description: "Free forever",
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mockService.On("GetPlanByID", "plan-1").Return(expectedPlan, nil)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/api/v1/plans/plan-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Plan
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "plan-1", response.ID)
	assert.Equal(t, "Free Plan", response.Name)

	mockService.AssertExpectations(t)
}

func TestGetPlanByID_NotFound(t *testing.T) {
	// Arrange
	mockService := new(MockPlanService)
	handler := handlers.NewPlanHandler(mockService)
	router := setupTestRouter()
	router.GET("/api/v1/plans/:id", handler.GetPlanByID)

	mockService.On("GetPlanByID", "nonexistent").Return(nil, assert.AnError)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/api/v1/plans/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreatePlan_Success(t *testing.T) {
	// Arrange
	mockService := new(MockPlanService)
	handler := handlers.NewPlanHandler(mockService)
	router := setupTestRouter()
	router.POST("/api/v1/plans", handler.CreatePlan)

	newPlan := models.Plan{
		Name:        "Enterprise Plan",
		PlanKey:     "enterprise",
		Description: "For large teams",
		Status:      "active",
	}

	mockService.On("CreatePlan", mock.AnythingOfType("*models.Plan")).Return(nil)

	// Act
	body, _ := json.Marshal(newPlan)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/plans", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreatePlan_InvalidJSON(t *testing.T) {
	// Arrange
	mockService := new(MockPlanService)
	handler := handlers.NewPlanHandler(mockService)
	router := setupTestRouter()
	router.POST("/api/v1/plans", handler.CreatePlan)

	// Act
	req := httptest.NewRequest(http.MethodPost, "/api/v1/plans", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePlan_ValidationError(t *testing.T) {
	// Arrange
	mockService := new(MockPlanService)
	handler := handlers.NewPlanHandler(mockService)
	router := setupTestRouter()
	router.POST("/api/v1/plans", handler.CreatePlan)

	invalidPlan := models.Plan{
		Name:    "", // Empty name should fail validation
		PlanKey: "test",
	}

	// Act
	body, _ := json.Marshal(invalidPlan)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/plans", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdatePlan_Success(t *testing.T) {
	// Arrange
	mockService := new(MockPlanService)
	handler := handlers.NewPlanHandler(mockService)
	router := setupTestRouter()
	router.PUT("/api/v1/plans/:id", handler.UpdatePlan)

	updatedPlan := models.Plan{
		Name:        "Updated Plan",
		PlanKey:     "updated",
		Description: "Updated description",
		Status:      "active",
	}

	mockService.On("UpdatePlan", "plan-1", mock.AnythingOfType("*models.Plan")).Return(nil)

	// Act
	body, _ := json.Marshal(updatedPlan)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/plans/plan-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeletePlan_Success(t *testing.T) {
	// Arrange
	mockService := new(MockPlanService)
	handler := handlers.NewPlanHandler(mockService)
	router := setupTestRouter()
	router.DELETE("/api/v1/plans/:id", handler.DeletePlan)

	mockService.On("DeletePlan", "plan-1").Return(nil)

	// Act
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/plans/plan-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeletePlan_NotFound(t *testing.T) {
	// Arrange
	mockService := new(MockPlanService)
	handler := handlers.NewPlanHandler(mockService)
	router := setupTestRouter()
	router.DELETE("/api/v1/plans/:id", handler.DeletePlan)

	mockService.On("DeletePlan", "nonexistent").Return(assert.AnError)

	// Act
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/plans/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}
