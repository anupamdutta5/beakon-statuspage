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

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode)

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
		&models.Tenant{},
		&models.SubscriptionPlan{},
		&models.Subscription{},
		&models.FeatureFlag{},
		&models.BillingEvent{},
		&models.UsageMetrics{},
		&models.AdminSettings{},
		&models.SystemNotification{},
		&models.BillingCustomer{},
		&models.BillingSubscription{},
		&models.BillingInvoice{},
		&models.BillingPayment{},
		&models.HealthCheck{},
		&models.Alert{},
		&models.UptimeStats{},
		&models.User{},
		&models.Service{},
		&models.Incident{},
		&models.StatusUpdate{},
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
		&models.APIUsage{},
	)
}

// Seed populates the database with initial data
func Seed() error {
	// Create default tenant first
	var tenant models.Tenant
	if err := DB.Where("slug = ?", "default").First(&tenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			defaultTenant := models.Tenant{
				Name:         "Default Tenant",
				Slug:         "default",
				ContactEmail: "admin@example.com",
				BillingEmail: "admin@example.com",
				Plan:         "free",
				Status:       "active",
				IsActive:     true,
			}
			if err := DB.Create(&defaultTenant).Error; err != nil {
				return err
			}
			tenant = defaultTenant
		} else {
			return err
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
				Email:    "admin@example.com",
				TenantID: &tenant.ID,
				IsActive: true,
			}
			if err := DB.Create(&adminUser).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Seed default services
	services := []models.Service{
		{Name: "API", Description: "The main API for the status page", Status: "operational", TenantID: tenant.ID},
		{Name: "Website", Description: "The main website for the status page", Status: "operational", TenantID: tenant.ID},
	}

	for _, service := range services {
		// Check if the service already exists
		var existingService models.Service
		if err := DB.Where("name = ? AND tenant_id = ?", service.Name, tenant.ID).First(&existingService).Error; err != nil {
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

	// Initialize default subscription plans
	plans := []models.SubscriptionPlan{
		{
			Name:            "Free",
			Slug:            "free",
			Description:     "Perfect for small teams and personal projects",
			Price:           0,
			Currency:        "USD",
			BillingInterval: "monthly",
			MaxServices:     5,
			MaxMonitors:     10,
			MaxSubscribers:  100,
			MaxIncidents:    10,
			MaxMaintenance:  5,
			CustomDomain:    false,
			WhiteLabel:      false,
			API:             false,
			Integrations:    false,
			Analytics:       false,
			Support:         "email",
			IsActive:        true,
			Features:        `["basic_monitoring", "email_notifications"]`,
		},
		{
			Name:            "Pro",
			Slug:            "pro",
			Description:     "Advanced features for growing businesses",
			Price:           29,
			Currency:        "USD",
			BillingInterval: "monthly",
			MaxServices:     25,
			MaxMonitors:     100,
			MaxSubscribers:  1000,
			MaxIncidents:    100,
			MaxMaintenance:  50,
			CustomDomain:    true,
			WhiteLabel:      false,
			API:             true,
			Integrations:    true,
			Analytics:       true,
			Support:         "chat",
			IsActive:        true,
			Features:        `["advanced_monitoring", "custom_domains", "api_access", "integrations", "analytics", "priority_support"]`,
		},
		{
			Name:            "Enterprise",
			Slug:            "enterprise",
			Description:     "Full-featured solution for large organizations",
			Price:           99,
			Currency:        "USD",
			BillingInterval: "monthly",
			MaxServices:     -1, // Unlimited
			MaxMonitors:     -1, // Unlimited
			MaxSubscribers:  -1, // Unlimited
			MaxIncidents:    -1, // Unlimited
			MaxMaintenance:  -1, // Unlimited
			CustomDomain:    true,
			WhiteLabel:      true,
			API:             true,
			Integrations:    true,
			Analytics:       true,
			Support:         "phone",
			IsActive:        true,
			Features:        `["unlimited_monitoring", "white_labeling", "custom_domains", "api_access", "all_integrations", "advanced_analytics", "phone_support", "sla_guarantee"]`,
		},
	}

	for _, plan := range plans {
		var existingPlan models.SubscriptionPlan
		if err := DB.Where("slug = ?", plan.Slug).First(&existingPlan).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := DB.Create(&plan).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	return nil
}

// GetDB returns the global database connection
func GetDB() *gorm.DB {
	return DB
}
