package unit

import (
	"context"
	"testing"
	"time"

	"github.com/enterprise-status/statuspage-landing-service/internal/config"
	"github.com/enterprise-status/statuspage-landing-service/internal/handlers"
	"github.com/enterprise-status/statuspage-landing-service/internal/models"
	"github.com/enterprise-status/statuspage-landing-service/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		// In a real test, you might want to use t.Fatal() instead of panic
		// For now, we'll use panic as this is a test utility function
		panic("Failed to connect to test database")
	}

	// Auto-migrate models
	err = db.AutoMigrate(
		&models.LandingPage{},
		&models.HeroSection{},
		&models.FeatureSection{},
		&models.PricingPlan{},
		&models.Testimonial{},
		&models.Article{},
		&models.FAQ{},
		&models.ContactForm{},
		&models.Newsletter{},
		&models.LandingPageStats{},
	)
	if err != nil {
		// In a real test, you might want to use t.Fatal() instead of panic
		// For now, we'll use panic as this is a test utility function
		panic("Failed to migrate test database")
	}

	return db
}

func TestLandingService_GetLandingPageData(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Create test config
	testConfig := &config.Config{
		Environment: "test",
		Service: config.ServiceConfig{
			Name:        "landing-page-service",
			Version:     "1.0.0",
			Description: "Landing Page Service for Testing",
		},
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8097,
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     "test_db",
		},
		Landing: config.LandingConfig{
			SiteName:        "Test StatusPage",
			SiteURL:         "http://localhost:8097",
			SiteDescription: "Test landing page for status page service",
			SiteKeywords:    []string{"test", "status", "page"},
			ContactEmail:    "test@example.com",
			SupportEmail:    "support@example.com",
			AnalyticsID:     "test-analytics",
			OGImage:         "/static/images/og-image.png",
			Favicon:         "/static/images/favicon.ico",
			Theme:           "modern",
		},
		Logging: config.LoggingConfig{
			Level:  "debug",
			Format: "console",
		},
	}

	// Create service using constructor
	service, err := services.NewLandingService(testConfig, logger)
	require.NoError(t, err)

	// Test getting landing page data
	data, err := service.GetLandingPageData(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.NotEmpty(t, data.SiteName)

	// Test that we get default data when no database is available
	assert.NotEmpty(t, data.Hero)
	assert.NotEmpty(t, data.Features)
	assert.NotEmpty(t, data.PricingPlans)
	assert.NotEmpty(t, data.Testimonials)
	assert.NotEmpty(t, data.FAQs)
}

func TestLandingService_HeroSection(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Create test config
	testConfig := &config.Config{
		Environment: "test",
		Service: config.ServiceConfig{
			Name:        "landing-page-service",
			Version:     "1.0.0",
			Description: "Landing Page Service for Testing",
		},
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8097,
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     "test_db",
		},
		Landing: config.LandingConfig{
			SiteName:        "Test StatusPage",
			SiteURL:         "http://localhost:8097",
			SiteDescription: "Test landing page for status page service",
			SiteKeywords:    []string{"test", "status", "page"},
			ContactEmail:    "test@example.com",
			SupportEmail:    "support@example.com",
			AnalyticsID:     "test-analytics",
			OGImage:         "/static/images/og-image.png",
			Favicon:         "/static/images/favicon.ico",
			Theme:           "modern",
		},
		Logging: config.LoggingConfig{
			Level:  "debug",
			Format: "console",
		},
	}

	// Create service using constructor
	service, err := services.NewLandingService(testConfig, logger)
	require.NoError(t, err)

	// Test getting hero section (should return default data when no database)
	hero, err := service.GetHeroSection(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, hero)
	assert.NotEmpty(t, hero.Title)
	assert.NotEmpty(t, hero.Subtitle)
	assert.NotEmpty(t, hero.Description)
	assert.NotEmpty(t, hero.ButtonText)
	assert.NotEmpty(t, hero.ButtonURL)
}

func TestLandingService_Article(t *testing.T) {
	db := setupTestDB()
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	logger, _ := zap.NewDevelopment()

	// Create test config
	testConfig := &config.Config{
		Environment: "test",
		Service: config.ServiceConfig{
			Name:        "landing-page-service",
			Version:     "1.0.0",
			Description: "Landing Page Service for Testing",
		},
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8097,
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     "test_db",
		},
		Logging: config.LoggingConfig{
			Level:  "debug",
			Format: "console",
		},
	}

	// Create service using constructor
	service, err := services.NewLandingService(testConfig, logger)
	require.NoError(t, err)
	// Override the database connection with our test database
	service.SetDB(db)

	// Test creating article
	publishedAt := time.Now()
	article := &models.Article{
		Title:       "Test Article",
		Slug:        "test-article",
		Excerpt:     "Test article excerpt",
		Content:     "Test article content",
		Author:      "Test Author",
		AuthorEmail: "test@example.com",
		Category:    "Technology",
		Tags:        `["test", "article"]`,
		Status:      "published",
		IsFeatured:  true,
		PublishedAt: &publishedAt,
	}

	err = service.CreateArticle(context.Background(), article)
	require.NoError(t, err)
	assert.NotZero(t, article.ID)

	// Test getting article
	retrievedArticle, err := service.GetArticle(context.Background(), "test-article")
	require.NoError(t, err)
	assert.Equal(t, article.Title, retrievedArticle.Title)
	assert.Equal(t, article.Slug, retrievedArticle.Slug)
}

func TestLandingService_Testimonial(t *testing.T) {
	db := setupTestDB()
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	logger, _ := zap.NewDevelopment()

	// Create test config
	testConfig := &config.Config{
		Environment: "test",
		Service: config.ServiceConfig{
			Name:        "landing-page-service",
			Version:     "1.0.0",
			Description: "Landing Page Service for Testing",
		},
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8097,
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     "test_db",
		},
		Logging: config.LoggingConfig{
			Level:  "debug",
			Format: "console",
		},
	}

	// Create service using constructor
	service, err := services.NewLandingService(testConfig, logger)
	require.NoError(t, err)
	// Override the database connection with our test database
	service.SetDB(db)

	// Test creating testimonial
	testimonial := &models.Testimonial{
		Name:     "John Doe",
		Company:  "Test Company",
		Position: "CEO",
		Avatar:   "https://example.com/avatar.jpg",
		Content:  "This is a great service!",
		Rating:   5,
		Status:   "active",
		Order:    1,
	}

	err = service.CreateTestimonial(context.Background(), testimonial)
	require.NoError(t, err)
	assert.NotZero(t, testimonial.ID)

	// Test getting testimonials
	testimonials, err := service.GetTestimonials(context.Background())
	require.NoError(t, err)
	assert.Len(t, testimonials, 1)
	assert.Equal(t, testimonial.Name, testimonials[0].Name)
}

func TestLandingService_FAQ(t *testing.T) {
	db := setupTestDB()
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	logger, _ := zap.NewDevelopment()

	// Create test config
	testConfig := &config.Config{
		Environment: "test",
		Service: config.ServiceConfig{
			Name:        "landing-page-service",
			Version:     "1.0.0",
			Description: "Landing Page Service for Testing",
		},
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8097,
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     "test_db",
		},
		Logging: config.LoggingConfig{
			Level:  "debug",
			Format: "console",
		},
	}

	// Create service using constructor
	service, err := services.NewLandingService(testConfig, logger)
	require.NoError(t, err)
	// Override the database connection with our test database
	service.SetDB(db)

	// Test creating FAQ
	faq := &models.FAQ{
		Question: "What is this service?",
		Answer:   "This is a test service for landing pages.",
		Category: "General",
		Status:   "active",
		Order:    1,
	}

	err = service.CreateFAQ(context.Background(), faq)
	require.NoError(t, err)
	assert.NotZero(t, faq.ID)

	// Test getting FAQs
	faqs, err := service.GetFAQs(context.Background())
	require.NoError(t, err)
	assert.Len(t, faqs, 1)
	assert.Equal(t, faq.Question, faqs[0].Question)
}

func TestLandingService_ContactForm(t *testing.T) {
	db := setupTestDB()
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	logger, _ := zap.NewDevelopment()

	// Create test config
	testConfig := &config.Config{
		Environment: "test",
		Service: config.ServiceConfig{
			Name:        "landing-page-service",
			Version:     "1.0.0",
			Description: "Landing Page Service for Testing",
		},
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8097,
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     "test_db",
		},
		Logging: config.LoggingConfig{
			Level:  "debug",
			Format: "console",
		},
	}

	// Create service using constructor
	service, err := services.NewLandingService(testConfig, logger)
	require.NoError(t, err)
	// Override the database connection with our test database
	service.SetDB(db)

	// Test submitting contact form
	contactForm := &models.ContactForm{
		Name:      "John Doe",
		Email:     "john@example.com",
		Company:   "Test Company",
		Subject:   "Test Subject",
		Message:   "This is a test message",
		Status:    "pending",
		IPAddress: "192.168.1.1",
		UserAgent: "Test User Agent",
	}

	err = service.SubmitContactForm(context.Background(), contactForm)
	require.NoError(t, err)
	assert.NotZero(t, contactForm.ID)
}

func TestLandingService_Newsletter(t *testing.T) {
	db := setupTestDB()
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	logger, _ := zap.NewDevelopment()

	// Create test config
	testConfig := &config.Config{
		Environment: "test",
		Service: config.ServiceConfig{
			Name:        "landing-page-service",
			Version:     "1.0.0",
			Description: "Landing Page Service for Testing",
		},
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8097,
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     "test_db",
		},
		Logging: config.LoggingConfig{
			Level:  "debug",
			Format: "console",
		},
	}

	// Create service using constructor
	service, err := services.NewLandingService(testConfig, logger)
	require.NoError(t, err)
	// Override the database connection with our test database
	service.SetDB(db)

	// Test newsletter subscription
	newsletter := &models.Newsletter{
		Email:     "test@example.com",
		Name:      "Test User",
		Status:    "active",
		Source:    "landing page",
		IPAddress: "192.168.1.1",
		UserAgent: "Test User Agent",
	}

	err = service.SubscribeNewsletter(context.Background(), newsletter)
	require.NoError(t, err)
	assert.NotZero(t, newsletter.ID)
}

func TestLandingHandler_HealthCheck(t *testing.T) {
	db := setupTestDB()
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	logger, _ := zap.NewDevelopment()

	// Create test config
	testConfig := &config.Config{
		Environment: "test",
		Service: config.ServiceConfig{
			Name:        "landing-page-service",
			Version:     "1.0.0",
			Description: "Landing Page Service for Testing",
		},
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8097,
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     "test_db",
		},
		Logging: config.LoggingConfig{
			Level:  "debug",
			Format: "console",
		},
	}

	// Create service using constructor
	service, err := services.NewLandingService(testConfig, logger)
	require.NoError(t, err)

	// Create handler
	handler := handlers.NewLandingHandler(service, logger)

	// Test health check
	err = service.Health(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, handler)
}
