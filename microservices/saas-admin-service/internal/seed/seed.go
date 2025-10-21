package seed

import (
	"log"

	"gorm.io/gorm"
)

// Seed runs all database seeders
func Seed(db *gorm.DB) error {
	// Add your seeders here
	if err := seedPricingPlans(db); err != nil {
		return err
	}

	return nil
}

func seedPricingPlans(db *gorm.DB) error {
	log.Println("Seeding pricing plans...")

	return SeedPricingPlans(db)
}
