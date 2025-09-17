package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/anupamdutta5/statuspage-landing-service/internal/config"
	"github.com/anupamdutta5/statuspage-landing-service/internal/handlers"
	"github.com/anupamdutta5/statuspage-landing-service/internal/models"
	"github.com/anupamdutta5/statuspage-landing-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	baseURL = "http://localhost:8097"
	timeout = 30 * time.Second
)

var (
	landingService *services.LandingService
	landingHandler *handlers.LandingHandler
	db             *gorm.DB
)

func TestMain(m *testing.M) {
	// Setup test environment
	setupIntegrationTest()

	// Run tests
	code := m.Run()

	// Cleanup
	teardownIntegrationTest()

	os.Exit(code)
}

func setupIntegrationTest() error {
	// Initialize in-memory SQLite database for testing
	var err error
	db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return err
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
		return err
	}

	// Initialize logger
	logger, _ := zap.NewDevelopment()

	// Create a test config
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

	// Initialize service
	landingService, err = services.NewLandingService(testConfig, logger)
	if err != nil {
		return err
	}

	// Set the test database
	landingService.SetDB(db)

	// Initialize handler
	landingHandler = handlers.NewLandingHandler(landingService, logger)

	return nil
}

func teardownIntegrationTest() {
	if db != nil {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

func TestLandingPageServiceHealth(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Test health check
	err = landingService.Health(context.Background())
	assert.NoError(t, err)
}

func TestLandingPageService_HeroSection(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Test getting hero section
	hero, err := landingService.GetHeroSection(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, hero)
	assert.NotEmpty(t, hero.Title)

	// Test creating hero section
	newHero := &models.HeroSection{
		Title:           "Integration Test Hero",
		Subtitle:        "Integration Test Subtitle",
		Description:     "Integration test description",
		ButtonText:      "Get Started",
		ButtonURL:       "/signup",
		BackgroundColor: "#ffffff",
		TextColor:       "#000000",
		Status:          "active",
		Order:           1,
	}

	err = landingService.CreateHeroSection(context.Background(), newHero)
	require.NoError(t, err)
	assert.NotZero(t, newHero.ID)
}

func TestLandingPageService_Features(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Test getting features
	features, err := landingService.GetFeatures(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, features)

	// Test creating feature
	feature := &models.FeatureSection{
		Title:       "Integration Test Feature",
		Description: "Integration test feature description",
		Icon:        "star",
		Status:      "active",
		Order:       1,
	}

	err = landingService.CreateFeature(context.Background(), feature)
	require.NoError(t, err)
	assert.NotZero(t, feature.ID)
}

func TestLandingPageService_PricingPlans(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Test getting pricing plans
	plans, err := landingService.GetPricingPlans(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, plans)

	// Test syncing pricing plans
	pricingPlans := []*models.PricingPlan{
		{
			PlanID:          1,
			Name:            "Basic Plan",
			Slug:            "basic",
			Description:     "Basic plan for small teams",
			Price:           29.99,
			Currency:        "USD",
			BillingInterval: "monthly",
			Features:        `["Feature 1", "Feature 2"]`,
			IsPopular:       false,
			IsActive:        true,
			ButtonText:      "Get Started",
			ButtonURL:       "/signup/basic",
			Status:          "active",
			Order:           1,
		},
		{
			PlanID:          2,
			Name:            "Pro Plan",
			Slug:            "pro",
			Description:     "Pro plan for growing teams",
			Price:           99.99,
			Currency:        "USD",
			BillingInterval: "monthly",
			Features:        `["Feature 1", "Feature 2", "Feature 3"]`,
			IsPopular:       true,
			IsActive:        true,
			ButtonText:      "Get Started",
			ButtonURL:       "/signup/pro",
			Status:          "active",
			Order:           2,
		},
	}

	err = landingService.SyncPricingPlans(context.Background(), pricingPlans)
	require.NoError(t, err)
}

func TestLandingPageService_Testimonials(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Test getting testimonials
	testimonials, err := landingService.GetTestimonials(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, testimonials)

	// Test creating testimonial
	testimonial := &models.Testimonial{
		Name:     "Integration Test User",
		Company:  "Test Company",
		Position: "CEO",
		Avatar:   "https://example.com/avatar.jpg",
		Content:  "This is a great service for integration testing!",
		Rating:   5,
		Status:   "active",
		Order:    1,
	}

	err = landingService.CreateTestimonial(context.Background(), testimonial)
	require.NoError(t, err)
	assert.NotZero(t, testimonial.ID)
}

func TestLandingPageService_Articles(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Test getting articles
	articles, err := landingService.GetArticles(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.NotNil(t, articles)

	// Test creating article
	publishedAt := time.Now()
	article := &models.Article{
		Title:       "Integration Test Article",
		Slug:        "integration-test-article",
		Excerpt:     "Integration test article excerpt",
		Content:     "Integration test article content",
		Author:      "Integration Test Author",
		AuthorEmail: "test@example.com",
		Category:    "Technology",
		Tags:        `["integration", "test", "article"]`,
		Status:      "published",
		IsFeatured:  true,
		PublishedAt: &publishedAt,
	}

	err = landingService.CreateArticle(context.Background(), article)
	require.NoError(t, err)
	assert.NotZero(t, article.ID)

	// Test getting article by slug
	retrievedArticle, err := landingService.GetArticle(context.Background(), "integration-test-article")
	require.NoError(t, err)
	assert.Equal(t, article.Title, retrievedArticle.Title)
}

func TestLandingPageService_FAQs(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Test getting FAQs
	faqs, err := landingService.GetFAQs(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, faqs)

	// Test creating FAQ
	faq := &models.FAQ{
		Question: "What is integration testing?",
		Answer:   "Integration testing is a type of testing where individual units are combined and tested as a group.",
		Category: "Testing",
		Status:   "active",
		Order:    1,
	}

	err = landingService.CreateFAQ(context.Background(), faq)
	require.NoError(t, err)
	assert.NotZero(t, faq.ID)
}

func TestLandingPageService_ContactForm(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Test submitting contact form
	contactForm := &models.ContactForm{
		Name:      "Integration Test User",
		Email:     "integration@example.com",
		Company:   "Integration Test Company",
		Subject:   "Integration Test Subject",
		Message:   "This is an integration test message",
		Status:    "pending",
		IPAddress: "192.168.1.1",
		UserAgent: "Integration Test User Agent",
	}

	err = landingService.SubmitContactForm(context.Background(), contactForm)
	require.NoError(t, err)
	assert.NotZero(t, contactForm.ID)
}

func TestLandingPageService_Newsletter(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Test newsletter subscription
	newsletter := &models.Newsletter{
		Email:     "integration@example.com",
		Name:      "Integration Test User",
		Status:    "active",
		Source:    "integration test",
		IPAddress: "192.168.1.1",
		UserAgent: "Integration Test User Agent",
	}

	err = landingService.SubscribeNewsletter(context.Background(), newsletter)
	require.NoError(t, err)
	assert.NotZero(t, newsletter.ID)
}

func TestLandingPageHandler_HTTP(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup routes
	router.GET("/health", landingHandler.HealthCheck)
	router.GET("/", landingHandler.LandingPage)
	router.GET("/blog", landingHandler.BlogPage)
	router.GET("/contact", landingHandler.ContactPage)
	router.POST("/contact", landingHandler.SubmitContactForm)
	router.POST("/newsletter", landingHandler.SubscribeNewsletter)

	// Test health check endpoint
	w := performRequest(router, "GET", "/health")
	assert.Equal(t, http.StatusOK, w.Code)

	// Test landing page endpoint
	w = performRequest(router, "GET", "/")
	assert.Equal(t, http.StatusOK, w.Code)

	// Test blog page endpoint
	w = performRequest(router, "GET", "/blog")
	assert.Equal(t, http.StatusOK, w.Code)

	// Test contact page endpoint
	w = performRequest(router, "GET", "/contact")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLandingPageHandler_ContactForm(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/contact", landingHandler.SubmitContactForm)

	// Test contact form submission
	contactData := map[string]string{
		"name":    "Test User",
		"email":   "test@example.com",
		"company": "Test Company",
		"subject": "Test Subject",
		"message": "Test message",
	}

	jsonData, _ := json.Marshal(contactData)
	w := performRequestWithBody(router, "POST", "/contact", bytes.NewBuffer(jsonData))
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestLandingPageHandler_Newsletter(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/newsletter", landingHandler.SubscribeNewsletter)

	// Test newsletter subscription
	newsletterData := map[string]string{
		"email": "test@example.com",
		"name":  "Test User",
	}

	jsonData, _ := json.Marshal(newsletterData)
	w := performRequestWithBody(router, "POST", "/newsletter", bytes.NewBuffer(jsonData))
	assert.Equal(t, http.StatusCreated, w.Code)
}

// Helper functions
func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func performRequestWithBody(r http.Handler, method, path string, body *bytes.Buffer) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
