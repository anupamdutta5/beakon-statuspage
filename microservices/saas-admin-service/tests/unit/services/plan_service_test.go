package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"saas-admin-service/internal/models"
	"saas-admin-service/internal/services"
)

// MockPlanRepository is a mock implementation of the plan repository
type MockPlanRepository struct {
	mock.Mock
}

func (m *MockPlanRepository) FindAll() ([]models.Plan, error) {
	args := m.Called()
	return args.Get(0).([]models.Plan), args.Error(1)
}

func (m *MockPlanRepository) FindByID(id string) (*models.Plan, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Plan), args.Error(1)
}

func (m *MockPlanRepository) FindByKey(key string) (*models.Plan, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Plan), args.Error(1)
}

func (m *MockPlanRepository) Create(plan *models.Plan) error {
	args := m.Called(plan)
	return args.Error(0)
}

func (m *MockPlanRepository) Update(plan *models.Plan) error {
	args := m.Called(plan)
	return args.Error(0)
}

func (m *MockPlanRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestGetAllPlans_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockPlanRepository)
	service := services.NewPlanService(mockRepo)

	now := time.Now()
	expectedPlans := []models.Plan{
		{
			ID:        "plan-1",
			Name:      "Free Plan",
			PlanKey:   "free",
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "plan-2",
			Name:      "Pro Plan",
			PlanKey:   "pro",
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	mockRepo.On("FindAll").Return(expectedPlans, nil)

	// Act
	plans, err := service.GetAllPlans()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, plans, 2)
	assert.Equal(t, "Free Plan", plans[0].Name)
	assert.Equal(t, "Pro Plan", plans[1].Name)
	mockRepo.AssertExpectations(t)
}

func TestGetAllPlans_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockPlanRepository)
	service := services.NewPlanService(mockRepo)

	mockRepo.On("FindAll").Return([]models.Plan{}, errors.New("database error"))

	// Act
	plans, err := service.GetAllPlans()

	// Assert
	assert.Error(t, err)
	assert.Empty(t, plans)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestGetPlanByID_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockPlanRepository)
	service := services.NewPlanService(mockRepo)

	now := time.Now()
	expectedPlan := &models.Plan{
		ID:        "plan-1",
		Name:      "Free Plan",
		PlanKey:   "free",
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo.On("FindByID", "plan-1").Return(expectedPlan, nil)

	// Act
	plan, err := service.GetPlanByID("plan-1")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, "plan-1", plan.ID)
	assert.Equal(t, "Free Plan", plan.Name)
	mockRepo.AssertExpectations(t)
}

func TestGetPlanByID_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockPlanRepository)
	service := services.NewPlanService(mockRepo)

	mockRepo.On("FindByID", "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	// Act
	plan, err := service.GetPlanByID("nonexistent")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, plan)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestCreatePlan_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockPlanRepository)
	service := services.NewPlanService(mockRepo)

	newPlan := &models.Plan{
		Name:        "Enterprise Plan",
		PlanKey:     "enterprise",
		Description: "For large teams",
		Status:      "active",
	}

	mockRepo.On("FindByKey", "enterprise").Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("Create", newPlan).Return(nil)

	// Act
	err := service.CreatePlan(newPlan)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, newPlan.ID)
	assert.False(t, newPlan.CreatedAt.IsZero())
	assert.False(t, newPlan.UpdatedAt.IsZero())
	mockRepo.AssertExpectations(t)
}

func TestCreatePlan_DuplicateKey(t *testing.T) {
	// Arrange
	mockRepo := new(MockPlanRepository)
	service := services.NewPlanService(mockRepo)

	existingPlan := &models.Plan{
		ID:      "plan-1",
		Name:    "Existing Plan",
		PlanKey: "existing",
	}

	newPlan := &models.Plan{
		Name:    "New Plan",
		PlanKey: "existing", // Duplicate key
	}

	mockRepo.On("FindByKey", "existing").Return(existingPlan, nil)

	// Act
	err := service.CreatePlan(newPlan)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plan with key 'existing' already exists")
	mockRepo.AssertExpectations(t)
}

func TestCreatePlan_ValidationError(t *testing.T) {
	// Arrange
	mockRepo := new(MockPlanRepository)
	service := services.NewPlanService(mockRepo)

	tests := []struct {
		name     string
		plan     *models.Plan
		errorMsg string
	}{
		{
			name: "Empty name",
			plan: &models.Plan{
				Name:    "",
				PlanKey: "test",
			},
			errorMsg: "name is required",
		},
		{
			name: "Empty plan key",
			plan: &models.Plan{
				Name:    "Test Plan",
				PlanKey: "",
			},
			errorMsg: "plan_key is required",
		},
		{
			name: "Invalid status",
			plan: &models.Plan{
				Name:    "Test Plan",
				PlanKey: "test",
				Status:  "invalid",
			},
			errorMsg: "status must be 'active', 'inactive', or 'archived'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := service.CreatePlan(tt.plan)

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestUpdatePlan_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockPlanRepository)
	service := services.NewPlanService(mockRepo)

	existingPlan := &models.Plan{
		ID:        "plan-1",
		Name:      "Old Name",
		PlanKey:   "test",
		Status:    "active",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now().Add(-24 * time.Hour),
	}

	updatedPlan := &models.Plan{
		Name:        "New Name",
		PlanKey:     "test",
		Description: "Updated description",
		Status:      "active",
	}

	mockRepo.On("FindByID", "plan-1").Return(existingPlan, nil)
	mockRepo.On("Update", mock.AnythingOfType("*models.Plan")).Return(nil)

	// Act
	err := service.UpdatePlan("plan-1", updatedPlan)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdatePlan_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockPlanRepository)
	service := services.NewPlanService(mockRepo)

	updatedPlan := &models.Plan{
		Name:    "New Name",
		PlanKey: "test",
	}

	mockRepo.On("FindByID", "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	// Act
	err := service.UpdatePlan("nonexistent", updatedPlan)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestDeletePlan_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockPlanRepository)
	service := services.NewPlanService(mockRepo)

	existingPlan := &models.Plan{
		ID:      "plan-1",
		Name:    "Test Plan",
		PlanKey: "test",
		Status:  "active",
	}

	mockRepo.On("FindByID", "plan-1").Return(existingPlan, nil)
	mockRepo.On("Delete", "plan-1").Return(nil)

	// Act
	err := service.DeletePlan("plan-1")

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeletePlan_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockPlanRepository)
	service := services.NewPlanService(mockRepo)

	mockRepo.On("FindByID", "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	// Act
	err := service.DeletePlan("nonexistent")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestDeletePlan_WithActiveSubscriptions(t *testing.T) {
	// Arrange
	mockRepo := new(MockPlanRepository)
	service := services.NewPlanService(mockRepo)

	existingPlan := &models.Plan{
		ID:                   "plan-1",
		Name:                 "Test Plan",
		PlanKey:              "test",
		Status:               "active",
		ActiveSubscriptions:  5, // Has active subscriptions
	}

	mockRepo.On("FindByID", "plan-1").Return(existingPlan, nil)

	// Act
	err := service.DeletePlan("plan-1")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete plan with active subscriptions")
	mockRepo.AssertExpectations(t)
}

func TestArchivePlan_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockPlanRepository)
	service := services.NewPlanService(mockRepo)

	existingPlan := &models.Plan{
		ID:      "plan-1",
		Name:    "Test Plan",
		PlanKey: "test",
		Status:  "active",
	}

	mockRepo.On("FindByID", "plan-1").Return(existingPlan, nil)
	mockRepo.On("Update", mock.AnythingOfType("*models.Plan")).Return(nil)

	// Act
	err := service.ArchivePlan("plan-1")

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
