// Package seed provides database seeding functionality
package main

import (
	"context"
	"log"
	"os"

	"github.com/anupamdutta5/statuspage-saas-admin-service/internal/config"
	"github.com/anupamdutta5/statuspage-saas-admin-service/internal/db/seed"
	"github.com/anupamdutta5/statuspage-saas-admin-service/internal/storage"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := storage.NewDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed the database
	if err := seed.Seed(db.DB); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	log.Println("✅ Database seeded successfully")
	os.Exit(0)
}
