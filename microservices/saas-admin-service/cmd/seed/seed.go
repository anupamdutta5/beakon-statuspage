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

	// NOTE: Database migrations are managed by Atlas (see ../../migrations/ and atlas.hcl)
	// Before running this seed script, ensure migrations are applied:
	//   cd microservices/saas-admin-service
	//   atlas migrate apply --env dev
	//
	// This seed script only populates initial data, not schema.

	// Convert to GORM DB for seeding
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to create GORM instance: %v", err)
	}

	// Seed the database with initial data
	log.Println("Seeding database with initial data...")
	if err := seed.SeedPricingPlans(gormDB); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	log.Println("Successfully seeded the database with sample data")
}