package main

import (
	"database/sql"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/anupamdutta5/saas-admin-service/internal/db/seed"
	"github.com/anupamdutta5/saas-admin-service/internal/db/migrations"
	_ "github.com/lib/pq"
)

func main() {
	// Connect to the database
	dsn := "host=localhost user=postgres password=postgres dbname=saas_admin port=5433 sslmode=disable"
	
	// Create a sql.DB connection for migrations
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer sqlDB.Close()

	// Run migrations
	fmt.Println("Running database migrations...")
	if err := migrations.RunMigrations(sqlDB); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Connect with GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database with GORM: %v", err)
	}

	// Run the seed function
	fmt.Println("Starting database seeding...")
	if err := seed.Seed(db); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	fmt.Println("Database seeded successfully!")
}
