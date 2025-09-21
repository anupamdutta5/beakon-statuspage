package main

import (
	"log"

	"github.com/anupamdutta5/saas-admin-service/internal/config"
	"github.com/anupamdutta5/saas-admin-service/internal/database"
	"github.com/anupamdutta5/saas-admin-service/internal/models"
	"github.com/anupamdutta5/saas-admin-service/internal/seed"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/driver/postgres"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize database
	db, err := database.InitDatabase(cfg, logger)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDatabase(db)

	// Auto-migrate models
	log.Println("Running database migrations...")

	// Convert to GORM DB for migration
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to create GORM instance: %v", err)
	}

	err = gormDB.AutoMigrate(
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
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed the database
	if err := seed.SeedPricingPlans(gormDB); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	log.Println("Successfully seeded the database with sample data")
}