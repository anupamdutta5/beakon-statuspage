package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"saas-admin-service/internal/models"
)

func TestPlan_Validate_Success(t *testing.T) {
	// Arrange
	plan := &models.Plan{
		Name:        "Test Plan",
		PlanKey:     "test",
		Description: "A test plan",
		Status:      "active",
	}

	// Act
	err := plan.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestPlan_Validate_EmptyName(t *testing.T) {
	// Arrange
	plan := &models.Plan{
		Name:    "",
		PlanKey: "test",
		Status:  "active",
	}

	// Act
	err := plan.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestPlan_Validate_EmptyPlanKey(t *testing.T) {
	// Arrange
	plan := &models.Plan{
		Name:    "Test Plan",
		PlanKey: "",
		Status:  "active",
	}

	// Act
	err := plan.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plan_key is required")
}

func TestPlan_Validate_InvalidStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
	}{
		{"Empty status", ""},
		{"Invalid status", "invalid"},
		{"Uppercase status", "ACTIVE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			plan := &models.Plan{
				Name:    "Test Plan",
				PlanKey: "test",
				Status:  tt.status,
			}

			// Act
			err := plan.Validate()

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "status must be")
		})
	}
}

func TestPlan_Validate_AllValidStatuses(t *testing.T) {
	validStatuses := []string{"active", "inactive", "archived"}

	for _, status := range validStatuses {
		t.Run("Valid status: "+status, func(t *testing.T) {
			// Arrange
			plan := &models.Plan{
				Name:    "Test Plan",
				PlanKey: "test",
				Status:  status,
			}

			// Act
			err := plan.Validate()

			// Assert
			assert.NoError(t, err)
		})
	}
}

func TestPlan_Validate_PlanKeyFormat(t *testing.T) {
	tests := []struct {
		name    string
		planKey string
		valid   bool
	}{
		{"Valid lowercase", "free", true},
		{"Valid with underscore", "free_trial", true},
		{"Valid with hyphen", "free-trial", true},
		{"Invalid uppercase", "FREE", false},
		{"Invalid space", "free trial", false},
		{"Invalid special chars", "free@plan", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			plan := &models.Plan{
				Name:    "Test Plan",
				PlanKey: tt.planKey,
				Status:  "active",
			}

			// Act
			err := plan.Validate()

			// Assert
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "plan_key must contain only lowercase letters")
			}
		})
	}
}

func TestPlan_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"Active plan", "active", true},
		{"Inactive plan", "inactive", false},
		{"Archived plan", "archived", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			plan := &models.Plan{
				Name:    "Test Plan",
				PlanKey: "test",
				Status:  tt.status,
			}

			// Act
			result := plan.IsActive()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPlan_CanBeDeleted(t *testing.T) {
	tests := []struct {
		name                 string
		activeSubscriptions  int
		expected             bool
	}{
		{"No subscriptions", 0, true},
		{"Has subscriptions", 5, false},
		{"One subscription", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			plan := &models.Plan{
				Name:                "Test Plan",
				PlanKey:             "test",
				Status:              "active",
				ActiveSubscriptions: tt.activeSubscriptions,
			}

			// Act
			result := plan.CanBeDeleted()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPlan_Archive(t *testing.T) {
	// Arrange
	plan := &models.Plan{
		Name:    "Test Plan",
		PlanKey: "test",
		Status:  "active",
	}

	// Act
	plan.Archive()

	// Assert
	assert.Equal(t, "archived", plan.Status)
}

func TestPlan_Activate(t *testing.T) {
	// Arrange
	plan := &models.Plan{
		Name:    "Test Plan",
		PlanKey: "test",
		Status:  "inactive",
	}

	// Act
	plan.Activate()

	// Assert
	assert.Equal(t, "active", plan.Status)
}

func TestPlan_Deactivate(t *testing.T) {
	// Arrange
	plan := &models.Plan{
		Name:    "Test Plan",
		PlanKey: "test",
		Status:  "active",
	}

	// Act
	plan.Deactivate()

	// Assert
	assert.Equal(t, "inactive", plan.Status)
}

func TestPlan_TableName(t *testing.T) {
	// Arrange
	plan := &models.Plan{}

	// Act
	tableName := plan.TableName()

	// Assert
	assert.Equal(t, "saas_plans", tableName)
}

func TestPlan_Clone(t *testing.T) {
	// Arrange
	original := &models.Plan{
		ID:                  "plan-1",
		Name:                "Original Plan",
		PlanKey:             "original",
		Description:         "Original description",
		Status:              "active",
		ActiveSubscriptions: 10,
	}

	// Act
	clone := original.Clone()

	// Assert
	assert.NotEqual(t, original.ID, clone.ID) // ID should be different
	assert.Empty(t, clone.ID)                  // New clone has no ID yet
	assert.Equal(t, original.Name+" (Copy)", clone.Name)
	assert.Equal(t, original.PlanKey+"_copy", clone.PlanKey)
	assert.Equal(t, original.Description, clone.Description)
	assert.Equal(t, "inactive", clone.Status) // Clone starts inactive
	assert.Equal(t, 0, clone.ActiveSubscriptions) // No subscriptions
}
