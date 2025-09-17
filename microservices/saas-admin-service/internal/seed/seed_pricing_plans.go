package seed

import (
	"log"

	"gorm.io/gorm"

	"github.com/anupamdutta5/statuspage-saas-admin-service/internal/models"
)

// SeedPricingPlans populates the database with sample pricing plans
func SeedPricingPlans(db *gorm.DB) error {
	// First create some pricing features
	features := []models.PricingFeature{
		{Name: "Basic Support", Description: "Email support with 48h response time", Category: "core", Order: 1, IsActive: true},
		{Name: "Project Management", Description: "Create and manage projects", Category: "core", Order: 2, IsActive: true},
		{Name: "Team Collaboration", Description: "Invite team members", Category: "core", Order: 3, IsActive: true},
		{Name: "API Access", Description: "Full access to REST API", Category: "advanced", Order: 4, IsActive: true},
		{Name: "Custom Integrations", Description: "Custom integration support", Category: "enterprise", Order: 5, IsActive: true},
		{Name: "24/7 Support", Description: "Dedicated support with SLA", Category: "enterprise", Order: 6, IsActive: true},
	}

	// Create pricing features
	for i := range features {
		if err := db.FirstOrCreate(&features[i], models.PricingFeature{Name: features[i].Name}).Error; err != nil {
			return err
		}
	}

	plans := []models.SaaSPlan{
		{
			Name:         "Starter",
			Slug:         "starter",
			Description:  "Perfect for small teams getting started",
			Price:        9.99,
			Currency:     "USD",
			IsActive:     true,
			IsPopular:    false,
			IsPublic:     true,
			DisplayOrder: 1,
		},
		{
			Name:         "Professional",
			Slug:         "professional",
			Description:  "For growing teams with advanced needs",
			Price:        29.99,
			Currency:     "USD",
			IsActive:     true,
			IsPopular:    true,
			IsPublic:     true,
			DisplayOrder: 2,
		},
		{
			Name:         "Enterprise",
			Slug:         "enterprise",
			Description:  "For large organizations with custom needs",
			Price:        99.99,
			Currency:     "USD",
			IsActive:     true,
			IsPopular:    false,
			IsPublic:     true,
			DisplayOrder: 3,
		},
	}

	// Use a transaction to ensure data consistency
	err := db.Transaction(func(tx *gorm.DB) error {
		// Create plans
		for i := range plans {
			if err := tx.FirstOrCreate(&plans[i], models.SaaSPlan{Slug: plans[i].Slug}).Error; err != nil {
				return err
			}

			log.Printf("Created plan: %s (ID: %s)", plans[i].Name, plans[i].ID)
		}

		return nil
	})

	if err != nil {
		log.Printf("Error seeding pricing plans: %v", err)
		return err
	}

	log.Println("✅ Successfully seeded pricing plans")
	return nil
}