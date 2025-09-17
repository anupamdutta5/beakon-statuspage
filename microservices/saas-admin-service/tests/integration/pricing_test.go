package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anupamdutta5/statuspage-saas-admin-service/internal/config"
	"github.com/anupamdutta5/statuspage-saas-admin-service/internal/handlers"
	"github.com/anupamdutta5/statuspage-saas-admin-service/internal/models"
	"github.com/anupamdutta5/statuspage-saas-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPricingTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate models
	err = db.AutoMigrate(
		&models.Platform{},
		&models.SaaSPlan{},
		&models.PricingTier{},
		&models.PricingFeature{},
		&models.PlanFeature{},
		&models.SaaSFeature{},
		&models.SaaSFeatureFlag{},
		&models.SaaSAdminUser{},
		&models.SaaSNotification{},
		&models.SaaSActivity{},
		&models.SaaSBackup{},
		&models.SaaSStats{},
	)
	require.NoError(t, err)

	return db
}

func setupPricingTestService(t *testing.T) (*services.SaaSAdminService, *gorm.DB) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		SaaS: config.SaaSConfig{
			PlatformName:   "Test Platform",
			PlatformURL:    "https://test.com",
			AdminEmail:     "admin@test.com",
			SupportEmail:   "support@test.com",
			DefaultPlan:    "free",
			AvailablePlans: []string{"free", "pro", "enterprise"},
		},
	}

	db := setupPricingTestDB(t)

	service, err := services.NewSaaSAdminService(cfg, logger, db)
	require.NoError(t, err)

	return service, db
}

func TestPricingFeatureCRUD(t *testing.T) {
	service, _ := setupPricingTestService(t)
	ctx := context.Background()

	// Create a pricing feature
	feature := &models.PricingFeature{
		Name:        "Custom Domain",
		Description: "Use your own domain for your status page",
		Category:    "advanced",
		Icon:        "fas fa-globe",
		IsActive:    true,
		Order:       1,
	}

	err := service.CreatePricingFeature(ctx, feature)
	require.NoError(t, err)
	assert.NotZero(t, feature.ID)

	// Get pricing features
	features, err := service.GetPricingFeatures(ctx, "")
	require.NoError(t, err)
	assert.Len(t, features, 1)
	assert.Equal(t, "Custom Domain", features[0].Name)

	// Update pricing feature
	feature.Description = "Updated description"
	updatedFeature, err := service.UpdatePricingFeature(ctx, feature.ID, feature)
	require.NoError(t, err)
	assert.Equal(t, "Updated description", updatedFeature.Description)

	// Delete pricing feature
	err = service.DeletePricingFeature(ctx, feature.ID)
	require.NoError(t, err)

	// Verify deletion
	features, err = service.GetPricingFeatures(ctx, "")
	require.NoError(t, err)
	assert.Len(t, features, 0)
}

func TestPricingPlanCRUD(t *testing.T) {
	service, _ := setupPricingTestService(t)
	ctx := context.Background()

	// Create a pricing plan
	plan := &models.SaaSPlan{
		Name:            "Pro Plan",
		Slug:            "pro",
		Description:     "Advanced features for growing businesses",
		Price:           29.99,
		Currency:        "USD",
		BillingInterval: "monthly",
		ButtonText:      "Start Free Trial",
		ButtonURL:       "/signup?plan=pro",
		IsPopular:       true,
		IsActive:        true,
		IsPublic:        true,
		DisplayOrder:    1,
		Features:        `["Up to 25 services", "Custom domains", "Advanced analytics"]`,
	}

	err := service.CreatePlan(ctx, plan)
	require.NoError(t, err)
	assert.NotZero(t, plan.ID)

	// Get plan by slug
	retrievedPlan, err := service.GetPlanBySlug(ctx, "pro")
	require.NoError(t, err)
	assert.Equal(t, "Pro Plan", retrievedPlan.Name)
	assert.Equal(t, 29.99, retrievedPlan.Price)

	// Update plan
	plan.Price = 39.99
	_, err = service.UpdatePlan(ctx, plan.ID, plan)
	require.NoError(t, err)

	// Verify update
	updatedPlan, err := service.GetPlan(ctx, plan.ID)
	require.NoError(t, err)
	assert.Equal(t, 39.99, updatedPlan.Price)

	// Delete plan
	err = service.DeletePlan(ctx, plan.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = service.GetPlan(ctx, plan.ID)
	assert.Error(t, err)
}

func TestPricingTierManagement(t *testing.T) {
	service, _ := setupPricingTestService(t)
	ctx := context.Background()

	// Create a plan first
	plan := &models.SaaSPlan{
		Name:            "Test Plan",
		Slug:            "test",
		Description:     "Test plan",
		Price:           10.00,
		Currency:        "USD",
		BillingInterval: "monthly",
		IsActive:        true,
		IsPublic:        true,
	}
	err := service.CreatePlan(ctx, plan)
	require.NoError(t, err)

	// Create monthly tier
	monthlyTier := &models.PricingTier{
		PlanID:          plan.ID,
		BillingInterval: "monthly",
		Price:           10.00,
		Currency:        "USD",
		DiscountPercent: 0,
		IsActive:        true,
	}
	err = service.CreatePricingTier(ctx, monthlyTier)
	require.NoError(t, err)

	// Create yearly tier with discount
	yearlyTier := &models.PricingTier{
		PlanID:          plan.ID,
		BillingInterval: "yearly",
		Price:           100.00,
		Currency:        "USD",
		DiscountPercent: 20,
		IsActive:        true,
	}
	err = service.CreatePricingTier(ctx, yearlyTier)
	require.NoError(t, err)

	// Get pricing tiers
	tiers, err := service.GetPricingTiers(ctx, plan.ID)
	require.NoError(t, err)
	assert.Len(t, tiers, 2)

	// Verify monthly tier
	monthlyTierFound := false
	yearlyTierFound := false
	for _, tier := range tiers {
		if tier.BillingInterval == "monthly" {
			monthlyTierFound = true
			assert.Equal(t, 10.00, tier.Price)
			assert.Equal(t, 0.0, tier.DiscountPercent)
		} else if tier.BillingInterval == "yearly" {
			yearlyTierFound = true
			assert.Equal(t, 100.00, tier.Price)
			assert.Equal(t, 20.0, tier.DiscountPercent)
		}
	}
	assert.True(t, monthlyTierFound)
	assert.True(t, yearlyTierFound)
}

func TestPlanFeatureAssignment(t *testing.T) {
	service, _ := setupPricingTestService(t)
	ctx := context.Background()

	// Create a plan
	plan := &models.SaaSPlan{
		Name:            "Test Plan",
		Slug:            "test",
		Description:     "Test plan",
		Price:           10.00,
		Currency:        "USD",
		BillingInterval: "monthly",
		IsActive:        true,
		IsPublic:        true,
	}
	err := service.CreatePlan(ctx, plan)
	require.NoError(t, err)

	// Create features
	feature1 := &models.PricingFeature{
		Name:        "Feature 1",
		Description: "First feature",
		Category:    "core",
		IsActive:    true,
		Order:       1,
	}
	err = service.CreatePricingFeature(ctx, feature1)
	require.NoError(t, err)

	feature2 := &models.PricingFeature{
		Name:        "Feature 2",
		Description: "Second feature",
		Category:    "advanced",
		IsActive:    true,
		Order:       2,
	}
	err = service.CreatePricingFeature(ctx, feature2)
	require.NoError(t, err)

	// Assign features to plan
	err = service.AssignFeatureToPlan(ctx, plan.ID, feature1.ID, 1)
	require.NoError(t, err)

	err = service.AssignFeatureToPlan(ctx, plan.ID, feature2.ID, 2)
	require.NoError(t, err)

	// Get plan features
	planFeatures, err := service.GetPlanFeatures(ctx, plan.ID)
	require.NoError(t, err)
	assert.Len(t, planFeatures, 2)

	// Verify feature order
	assert.Equal(t, 1, planFeatures[0].Order)
	assert.Equal(t, "Feature 1", planFeatures[0].Feature.Name)
	assert.Equal(t, 2, planFeatures[1].Order)
	assert.Equal(t, "Feature 2", planFeatures[1].Feature.Name)

	// Remove a feature
	err = service.RemoveFeatureFromPlan(ctx, plan.ID, feature1.ID)
	require.NoError(t, err)

	// Verify removal
	planFeatures, err = service.GetPlanFeatures(ctx, plan.ID)
	require.NoError(t, err)
	assert.Len(t, planFeatures, 1)
	assert.Equal(t, "Feature 2", planFeatures[0].Feature.Name)
}

func TestPublicPricingPlans(t *testing.T) {
	service, _ := setupPricingTestService(t)
	ctx := context.Background()

	// Create multiple plans
	plans := []*models.SaaSPlan{
		{
			Name:            "Free Plan",
			Slug:            "free",
			Description:     "Free plan",
			Price:           0.00,
			Currency:        "USD",
			BillingInterval: "monthly",
			IsActive:        true,
			IsPublic:        true,
			DisplayOrder:    0,
		},
		{
			Name:            "Pro Plan",
			Slug:            "pro",
			Description:     "Pro plan",
			Price:           29.99,
			Currency:        "USD",
			BillingInterval: "monthly",
			IsPopular:       true,
			IsActive:        true,
			IsPublic:        true,
			DisplayOrder:    1,
		},
		{
			Name:            "Enterprise Plan",
			Slug:            "enterprise",
			Description:     "Enterprise plan",
			Price:           99.99,
			Currency:        "USD",
			BillingInterval: "monthly",
			IsActive:        true,
			IsPublic:        false, // Not public
			DisplayOrder:    2,
		},
	}

	for _, plan := range plans {
		err := service.CreatePlan(ctx, plan)
		require.NoError(t, err)

		// Explicitly update IsPublic field to ensure it's set correctly
		if plan.Name == "Enterprise Plan" {
			updateData := &models.SaaSPlan{IsPublic: false}
			_, err = service.UpdatePlan(ctx, plan.ID, updateData)
			require.NoError(t, err)
		}
	}

	// Get public pricing plans
	publicPlans, err := service.GetPublicPricingPlans(ctx)
	require.NoError(t, err)
	assert.Len(t, publicPlans, 2) // Only free and pro plans

	// Verify order
	assert.Equal(t, "Free Plan", publicPlans[0].Name)
	assert.Equal(t, "Pro Plan", publicPlans[1].Name)
	assert.True(t, publicPlans[1].IsPopular)
}

func TestPricingPlanValidation(t *testing.T) {
	service, _ := setupPricingTestService(t)
	ctx := context.Background()

	// Test invalid plan
	invalidPlan := &models.SaaSPlan{
		Name:            "", // Empty name
		Slug:            "test",
		Description:     "Test plan",
		Price:           -10.00, // Negative price
		Currency:        "USD",
		BillingInterval: "invalid", // Invalid billing interval
	}

	err := service.ValidatePricingPlan(ctx, invalidPlan)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plan name is required")
	assert.Contains(t, err.Error(), "plan price cannot be negative")
	assert.Contains(t, err.Error(), "billing interval must be 'monthly' or 'yearly'")

	// Test valid plan
	validPlan := &models.SaaSPlan{
		Name:            "Valid Plan",
		Slug:            "valid",
		Description:     "Valid plan",
		Price:           10.00,
		Currency:        "USD",
		BillingInterval: "monthly",
		Features:        `["Feature 1", "Feature 2"]`,
	}

	err = service.ValidatePricingPlan(ctx, validPlan)
	assert.NoError(t, err)
}

func TestPricingHTTPHandlers(t *testing.T) {
	service, _ := setupPricingTestService(t)
	logger, _ := zap.NewDevelopment()
	handler := handlers.NewSaaSAdminHandler(service, logger)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/pricing/features", handler.CreatePricingFeature)
	router.GET("/api/v1/pricing/features", handler.GetPricingFeatures)
	router.GET("/api/v1/pricing/plans/public", handler.GetPublicPricingPlans)

	// Test create pricing feature
	featureData := map[string]interface{}{
		"name":        "Test Feature",
		"description": "Test feature description",
		"category":    "core",
		"icon":        "fas fa-check",
		"is_active":   true,
	}

	jsonData, _ := json.Marshal(featureData)
	req := httptest.NewRequest("POST", "/api/v1/pricing/features", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Test get pricing features
	req = httptest.NewRequest("GET", "/api/v1/pricing/features", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "features")
	assert.Contains(t, response, "count")
}

func TestPricingSystemIntegration(t *testing.T) {
	service, _ := setupPricingTestService(t)
	ctx := context.Background()

	// Create features
	features := []*models.PricingFeature{
		{
			Name:        "Basic Support",
			Description: "Email support",
			Category:    "core",
			Icon:        "fas fa-envelope",
			IsActive:    true,
			Order:       1,
		},
		{
			Name:        "Custom Domain",
			Description: "Use your own domain",
			Category:    "advanced",
			Icon:        "fas fa-globe",
			IsActive:    true,
			Order:       2,
		},
		{
			Name:        "White Label",
			Description: "Remove branding",
			Category:    "enterprise",
			Icon:        "fas fa-paint-brush",
			IsActive:    true,
			Order:       3,
		},
	}

	for _, feature := range features {
		err := service.CreatePricingFeature(ctx, feature)
		require.NoError(t, err)
	}

	// Create plans
	plans := []*models.SaaSPlan{
		{
			Name:            "Free",
			Slug:            "free",
			Description:     "Perfect for small teams",
			Price:           0.00,
			Currency:        "USD",
			BillingInterval: "monthly",
			ButtonText:      "Get Started",
			ButtonURL:       "/signup?plan=free",
			IsActive:        true,
			IsPublic:        true,
			DisplayOrder:    0,
		},
		{
			Name:            "Pro",
			Slug:            "pro",
			Description:     "Advanced features for growing businesses",
			Price:           29.99,
			Currency:        "USD",
			BillingInterval: "monthly",
			ButtonText:      "Start Free Trial",
			ButtonURL:       "/signup?plan=pro",
			IsPopular:       true,
			IsActive:        true,
			IsPublic:        true,
			DisplayOrder:    1,
		},
		{
			Name:            "Enterprise",
			Slug:            "enterprise",
			Description:     "Complete solution for large organizations",
			Price:           99.99,
			Currency:        "USD",
			BillingInterval: "monthly",
			ButtonText:      "Contact Sales",
			ButtonURL:       "/contact?plan=enterprise",
			IsActive:        true,
			IsPublic:        true,
			DisplayOrder:    2,
		},
	}

	for _, plan := range plans {
		err := service.CreatePlan(ctx, plan)
		require.NoError(t, err)
	}

	// Assign features to plans
	// Free plan gets basic support
	err := service.AssignFeatureToPlan(ctx, plans[0].ID, features[0].ID, 1)
	require.NoError(t, err)

	// Pro plan gets basic support and custom domain
	err = service.AssignFeatureToPlan(ctx, plans[1].ID, features[0].ID, 1)
	require.NoError(t, err)
	err = service.AssignFeatureToPlan(ctx, plans[1].ID, features[1].ID, 2)
	require.NoError(t, err)

	// Enterprise plan gets all features
	for i, feature := range features {
		err = service.AssignFeatureToPlan(ctx, plans[2].ID, feature.ID, i+1)
		require.NoError(t, err)
	}

	// Create pricing tiers for Pro plan
	tiers := []*models.PricingTier{
		{
			PlanID:          plans[1].ID,
			BillingInterval: "monthly",
			Price:           29.99,
			Currency:        "USD",
			DiscountPercent: 0,
			IsActive:        true,
		},
		{
			PlanID:          plans[1].ID,
			BillingInterval: "yearly",
			Price:           299.99,
			Currency:        "USD",
			DiscountPercent: 20,
			IsActive:        true,
		},
	}

	for _, tier := range tiers {
		err = service.CreatePricingTier(ctx, tier)
		require.NoError(t, err)
	}

	// Test complete system
	publicPlans, err := service.GetPublicPricingPlans(ctx)
	require.NoError(t, err)
	assert.Len(t, publicPlans, 3)

	// Verify Pro plan features
	proPlanFeatures, err := service.GetPlanFeatures(ctx, plans[1].ID)
	require.NoError(t, err)
	assert.Len(t, proPlanFeatures, 2)

	// Verify Pro plan tiers
	proPlanTiers, err := service.GetPricingTiers(ctx, plans[1].ID)
	require.NoError(t, err)
	assert.Len(t, proPlanTiers, 2)

	// Test sync functionality
	err = service.SyncPricingToLandingPage(ctx)
	require.NoError(t, err)

	// Verify all features are available
	allFeatures, err := service.GetPricingFeatures(ctx, "")
	require.NoError(t, err)
	assert.Len(t, allFeatures, 3)

	// Test feature filtering by category
	coreFeatures, err := service.GetPricingFeatures(ctx, "core")
	require.NoError(t, err)
	assert.Len(t, coreFeatures, 1)
	assert.Equal(t, "Basic Support", coreFeatures[0].Name)

	advancedFeatures, err := service.GetPricingFeatures(ctx, "advanced")
	require.NoError(t, err)
	assert.Len(t, advancedFeatures, 1)
	assert.Equal(t, "Custom Domain", advancedFeatures[0].Name)

	enterpriseFeatures, err := service.GetPricingFeatures(ctx, "enterprise")
	require.NoError(t, err)
	assert.Len(t, enterpriseFeatures, 1)
	assert.Equal(t, "White Label", enterpriseFeatures[0].Name)
}

func TestPricingSystemPerformance(t *testing.T) {
	service, _ := setupPricingTestService(t)
	ctx := context.Background()

	// Create many features and plans to test performance
	start := time.Now()

	// Create 100 features
	for i := 0; i < 100; i++ {
		feature := &models.PricingFeature{
			Name:        fmt.Sprintf("Feature %d", i),
			Description: fmt.Sprintf("Description for feature %d", i),
			Category:    []string{"core", "advanced", "enterprise"}[i%3],
			Icon:        "fas fa-check",
			IsActive:    true,
			Order:       i,
		}
		err := service.CreatePricingFeature(ctx, feature)
		require.NoError(t, err)
	}

	// Create 10 plans
	for i := 0; i < 10; i++ {
		plan := &models.SaaSPlan{
			Name:            fmt.Sprintf("Plan %d", i),
			Slug:            fmt.Sprintf("plan-%d", i),
			Description:     fmt.Sprintf("Description for plan %d", i),
			Price:           float64(i * 10),
			Currency:        "USD",
			BillingInterval: "monthly",
			IsActive:        true,
			IsPublic:        true,
			DisplayOrder:    i,
		}
		err := service.CreatePlan(ctx, plan)
		require.NoError(t, err)
	}

	creationTime := time.Since(start)
	t.Logf("Created 100 features and 10 plans in %v", creationTime)

	// Test retrieval performance
	start = time.Now()
	features, err := service.GetPricingFeatures(ctx, "")
	require.NoError(t, err)
	assert.Len(t, features, 100)

	plans, err := service.GetPublicPricingPlans(ctx)
	require.NoError(t, err)
	assert.Len(t, plans, 10)

	retrievalTime := time.Since(start)
	t.Logf("Retrieved all features and plans in %v", retrievalTime)

	// Performance should be reasonable (less than 1 second for this test)
	assert.Less(t, creationTime, 5*time.Second)
	assert.Less(t, retrievalTime, 1*time.Second)
}
