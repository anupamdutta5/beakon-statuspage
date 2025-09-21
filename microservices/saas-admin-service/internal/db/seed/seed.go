package seed

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/anupamdutta5/saas-admin-service/internal/models"
	"gorm.io/gorm"
)

// Seed initializes the database with sample data
func Seed(db *gorm.DB) error {
	// Create sample plans
	plans := createSamplePlans()

	// Create pricing features first in a separate transaction
	features := []models.PricingFeature{
		{Name: "Team Members", Description: "Number of team members"},
		{Name: "Projects", Description: "Number of projects"},
		{Name: "Support", Description: "Type of support"},
		{Name: "API Access", Description: "Access to REST API"},
		{Name: "Custom Integrations", Description: "Custom integration support"},
		{Name: "Account Manager", Description: "Dedicated account manager"},
	}

	// Create a map of feature names to their IDs
	featureIDs := make(map[string]uint)

	// Create features if they don't exist in a separate transaction
	txFeatures := db.Begin()
	for i := range features {
		var existingFeature models.PricingFeature
		if err := txFeatures.Where("name = ?", features[i].Name).First(&existingFeature).Error; err != nil {
			if err := txFeatures.Create(&features[i]).Error; err != nil {
				txFeatures.Rollback()
				return fmt.Errorf("failed to create feature %s: %w", features[i].Name, err)
			}
			log.Printf("Created feature: %s (ID: %d)\n", features[i].Name, features[i].ID)
			featureIDs[features[i].Name] = features[i].ID
		} else {
			log.Printf("Feature already exists: %s (ID: %d)\n", existingFeature.Name, existingFeature.ID)
			featureIDs[existingFeature.Name] = existingFeature.ID
		}
	}

	// Commit the features transaction
	if err := txFeatures.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit features transaction: %w", err)
	}

	// Start a new transaction for plans
	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	// Ensure we rollback in case of error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // Re-throw panic after rollback
		}
	}()

	// Create plans
	for i, plan := range plans {
		// Check if plan already exists
		var existingPlan models.SaaSPlan
		if err := tx.Where("id = ?", plan.ID).First(&existingPlan).Error; err == nil {
			log.Printf("Plan %s already exists, skipping...\n", plan.Name)
			continue
		}

		// Set feature IDs based on the plan type
		if i == 0 { // Starter plan
			plans[i].PlanFeatures = []models.PlanFeature{
				{FeatureID: featureIDs["Team Members"], IsEnabled: true, Order: 1, PlanID: plan.ID},
				{FeatureID: featureIDs["Projects"], IsEnabled: true, Order: 2, PlanID: plan.ID},
				{FeatureID: featureIDs["Support"], IsEnabled: true, Order: 3, PlanID: plan.ID},
			}
		} else if i == 1 { // Professional plan
			plans[i].PlanFeatures = []models.PlanFeature{
				{FeatureID: featureIDs["Team Members"], IsEnabled: true, Order: 1, PlanID: plan.ID},
				{FeatureID: featureIDs["Projects"], IsEnabled: true, Order: 2, PlanID: plan.ID},
				{FeatureID: featureIDs["Support"], IsEnabled: true, Order: 3, PlanID: plan.ID},
				{FeatureID: featureIDs["API Access"], IsEnabled: true, Order: 4, PlanID: plan.ID},
			}
		} else if i == 2 { // Enterprise plan
			plans[i].PlanFeatures = []models.PlanFeature{
				{FeatureID: featureIDs["Team Members"], IsEnabled: true, Order: 1, PlanID: plan.ID},
				{FeatureID: featureIDs["Projects"], IsEnabled: true, Order: 2, PlanID: plan.ID},
				{FeatureID: featureIDs["Support"], IsEnabled: true, Order: 3, PlanID: plan.ID},
				{FeatureID: featureIDs["API Access"], IsEnabled: true, Order: 4, PlanID: plan.ID},
				{FeatureID: featureIDs["Custom Integrations"], IsEnabled: true, Order: 5, PlanID: plan.ID},
				{FeatureID: featureIDs["Account Manager"], IsEnabled: true, Order: 6, PlanID: plan.ID},
			}
		}
		
		// Log the feature IDs being used
		log.Printf("Creating plan %s with features: %+v\n", plan.Name, plans[i].PlanFeatures)

		// Create the plan without associations first
		if err := tx.Omit("PricingTiers", "PlanFeatures").Create(&plan).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create plan %s: %w", plan.Name, err)
		}

		// Log the created plan
		log.Printf("Created plan: %s (ID: %s)\n", plan.Name, plan.ID.String())

		// Create pricing tiers
		for _, tier := range plan.PricingTiers {
			tier.PlanID = plan.ID
			if err := tx.Create(&tier).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create pricing tier for plan %s: %w", plan.Name, err)
			}
			log.Printf("  - Created %s tier: $%.2f/%s\n", tier.BillingInterval, tier.Price, tier.Currency)
		}

		// Create plan features
		for j := range plans[i].PlanFeatures {
			plans[i].PlanFeatures[j].PlanID = plan.ID
			log.Printf("Creating plan feature: PlanID=%s, FeatureID=%d, Order=%d\n", 
				plans[i].PlanFeatures[j].PlanID, 
				plans[i].PlanFeatures[j].FeatureID,
				plans[i].PlanFeatures[j].Order)
			
			if err := tx.Create(&plans[i].PlanFeatures[j]).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create plan feature for plan %s: %w", plan.Name, err)
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Println("Successfully seeded the database with sample pricing plans and features")
	return nil
}

// createSamplePlans generates sample plans with pricing tiers and features
// Note: Feature IDs will be set in the Seed function
func createSamplePlans() []models.SaaSPlan {
	// Create plan IDs
	starterID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	professionalID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	enterpriseID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	return []models.SaaSPlan{
		{
			ID:            starterID,
			Name:          "Starter",
			Slug:          "starter",
			Description:   "Perfect for small teams getting started",
			Price:         9.99,
			Currency:      "USD",
			BillingInterval: "monthly",
			MaxUsers:      3,
			MaxServices:   5,
			IsActive:      true,
			IsPublic:      true,
			IsPopular:     false,
			DisplayOrder:  1,
			Features:      "Team Members: 3\nProjects: 5\nSupport: Email (48h response)",
			PricingTiers: []models.PricingTier{
				{
					BillingInterval: "monthly",
					Price:          9.99,
					Currency:       "USD",
					IsActive:       true,
					PlanID:         starterID,
				},
			},
			PlanFeatures: []models.PlanFeature{
				{
					// FeatureID will be set based on the created features
					IsEnabled: true,
					Order:     1,
					PlanID:    starterID,
				},
				{
					IsEnabled: true,
					Order:     2,
					PlanID:    starterID,
				},
				{
					IsEnabled: true,
					Order:     3,
					PlanID:    starterID,
				},
			},
		},
		{
			ID:            professionalID,
			Name:          "Professional",
			Slug:          "professional",
			Description:   "For growing teams with advanced needs",
			Price:         29.99,
			Currency:      "USD",
			BillingInterval: "monthly",
			MaxUsers:      10,
			MaxServices:   50,
			IsActive:      true,
			IsPublic:      true,
			IsPopular:     true,
			DisplayOrder:  2,
			Features:      "Team Members: 10\nUnlimited Projects\nPriority Support (24h response)\nAPI Access",
			PricingTiers: []models.PricingTier{
				{
					BillingInterval: "monthly",
					Price:          99.99,
					Currency:       "USD",
					IsActive:       true,
					PlanID:         enterpriseID,
				},
				{
					BillingInterval: "yearly",
					Price:          999.99,
					Currency:       "USD",
					DiscountPercent: 16.7, // ~2 months free
					IsActive:       true,
					PlanID:         enterpriseID,
				},
			},
			PlanFeatures: []models.PlanFeature{
				{
					IsEnabled: true,
					Order:     1,
					PlanID:    professionalID,
				},
				{
					IsEnabled: true,
					Order:     2,
					PlanID:    professionalID,
				},
				{
					IsEnabled: true,
					Order:     3,
					PlanID:    professionalID,
				},
				{
					IsEnabled: true,
					Order:     4,
					PlanID:    professionalID,
				},
			},
		},
		{
			ID:            enterpriseID,
			Name:          "Enterprise",
			Slug:          "enterprise",
			Description:   "For large organizations with custom needs",
			Price:         0, // Custom pricing
			Currency:      "USD",
			BillingInterval: "custom",
			MaxUsers:      0, // Unlimited
			MaxServices:   0, // Unlimited
			IsActive:      true,
			IsPublic:      true,
			IsPopular:     false,
			DisplayOrder:  3,
			Features:      "Unlimited Team Members\nUnlimited Projects\n24/7 Dedicated Support\nAPI Access\nCustom Integrations\nDedicated Account Manager",
			PricingTiers: []models.PricingTier{
				{
					BillingInterval: "custom",
					Price:          0, // Custom pricing
					Currency:       "USD",
					IsActive:       true,
					PlanID:         enterpriseID,
				},
			},
			PlanFeatures: []models.PlanFeature{
				{
					IsEnabled: true,
					Order:     1,
					PlanID:    enterpriseID,
				},
				{
					IsEnabled: true,
					Order:     2,
					PlanID:    enterpriseID,
				},
				{
					IsEnabled: true,
					Order:     3,
					PlanID:    enterpriseID,
				},
				{
					IsEnabled: true,
					Order:     4,
					PlanID:    enterpriseID,
				},
				{
					IsEnabled: true,
					Order:     5,
					PlanID:    enterpriseID,
				},
				{
					IsEnabled: true,
					Order:     6,
					PlanID:    enterpriseID,
				},
			},
		},
	}
}
