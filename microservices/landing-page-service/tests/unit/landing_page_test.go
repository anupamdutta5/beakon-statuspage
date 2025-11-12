package unit

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/anupamdutta5/landing-page-service/internal/config"
	"github.com/anupamdutta5/landing-page-service/internal/handlers"
	"github.com/anupamdutta5/landing-page-service/internal/models"
	"github.com/anupamdutta5/landing-page-service/internal/services"
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

	// Load test config
	testConfig := loadTestConfig(t)

	// Create service using constructor
	db := setupTestDB()
	service := services.NewLandingService(db, testConfig, logger)

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

	// Load test config
	testConfig := loadTestConfig(t)

	// Create service using constructor
	db := setupTestDB()
	service := services.NewLandingService(db, testConfig, logger)

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

	// Load test config
	testConfig := loadTestConfig(t)

	// Create service using constructor
	service := services.NewLandingService(db, testConfig, logger)

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

	err := service.CreateArticle(context.Background(), article)
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

	// Load test config
	testConfig := loadTestConfig(t)

	// Create service using constructor
	service := services.NewLandingService(db, testConfig, logger)

	// Test creating testimonial
	testimonial := &models.Testimonial{
		Name:     "John Doe",
		Company:  "Test Company",
		Position: "CEO",
		Avatar:   "https://example.com/avatar.jpg",
		Content:  "This is a great service!",
		Rating:   5,
		Status:   "active",
	}

	err := service.CreateTestimonial(context.Background(), testimonial)
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

	// Load test config
	testConfig := loadTestConfig(t)

	// Create service using constructor
	service := services.NewLandingService(db, testConfig, logger)

	// Test creating FAQ
	faq := &models.FAQ{
		Question: "What is this service?",
		Answer:   "This is a test service for landing pages.",
		Category: "General",
		Status:   "active",
		SortOrder: 1,
	}

	err := service.CreateFAQ(context.Background(), faq)
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

	// Load test config
	testConfig := loadTestConfig(t)

	// Create service using constructor
	service := services.NewLandingService(db, testConfig, logger)

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

	err := service.SubmitContactForm(context.Background(), contactForm)
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

	// Load test config
	testConfig := loadTestConfig(t)

	// Create service using constructor
	service := services.NewLandingService(db, testConfig, logger)

	// Test newsletter subscription
	newsletter := &models.Newsletter{
		Email:     "test@example.com",
		Name:      "Test User",
		Status:    "active",
		Source:    "landing page",
		IPAddress: "192.168.1.1",
		UserAgent: "Test User Agent",
	}

	err := service.SubscribeNewsletter(context.Background(), newsletter)
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

	// Load test config
	testConfig := loadTestConfig(t)

	// Create service using constructor
	service := services.NewLandingService(db, testConfig, logger)

	// Create handler
	handler := handlers.NewLandingHandler(service, logger)

	// Test health check
	err := service.Health(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, handler)
}

// Helper function to load test configuration
func loadTestConfig(t *testing.T) *config.Config {
	// Use development environment for tests
	os.Setenv("ENVIRONMENT", "development")
	defer os.Unsetenv("ENVIRONMENT")

	// Load configuration using the service's standard config loader
	// The shared-resilience library will automatically find the correct config path
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load test configuration: %v", err)
	}

	// Override database settings for testing (SQLite will be provided directly)
	cfg.Database.Host = "localhost"
	cfg.Database.Name = ":memory:"

	return &cfg
}
