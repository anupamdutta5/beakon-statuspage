package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database connection
var DB *gorm.DB

// Connect initializes the database connection
func Connect(cfg *config.DatabaseConfig) error {
	var err error

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	for i := 0; i < 5; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: newLogger,
		})
		if err == nil {
			break
		}
		log.Printf("failed to connect to database (attempt %d): %s", i+1, err)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database after multiple attempts: %w", err)
	}

	log.Println("Database connection established")

	log.Println("Running database migrations")
	if err := AutoMigrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Seeding database")
	if err := Seed(); err != nil {
		return fmt.Errorf("failed to seed database: %w", err)
	}

	return nil
}

// AutoMigrate runs the database migrations
func AutoMigrate() error {
	return DB.AutoMigrate(
		&models.Service{}, 
		&models.Incident{}, 
		&models.StatusUpdate{}, 
		&models.User{}, 
		&models.Maintenance{}, 
		&models.Subscriber{}, 
		&models.Component{}, 
		&models.Monitor{}, 
		&models.Heartbeat{},
		&models.IncidentTemplate{},
		&models.MaintenanceTemplate{},
		&models.AuditLog{},
		&models.Branding{},
		&models.Integration{},
		&models.SystemMetric{},
		&models.ThirdPartyService{},
		&models.PrivatePage{},
	)
}

// Seed populates the database with initial data
func Seed() error {
	services := []models.Service{
		{Name: "API", Description: "The main API for the status page", Status: "operational"},
		{Name: "Website", Description: "The main website for the status page", Status: "operational"},
	}

	for _, service := range services {
		// Check if the service already exists
		var existingService models.Service
		if err := DB.Where("name = ?", service.Name).First(&existingService).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create the service if it doesn't exist
				if err := DB.Create(&service).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	// Seed initial admin user
	var user models.User
	if err := DB.Where("username = ?", "admin").First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			adminUser := models.User{
				Username: "admin",
				Password: "password", // This will be hashed by the BeforeSave hook
				Role:     "admin",
			}
			if err := DB.Create(&adminUser).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

// GetDB returns the global database connection
func GetDB() *gorm.DB {
	return DB
}
