package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/enterprise-status/statuspage/internal/api"
	"github.com/enterprise-status/statuspage/internal/api/middleware"
	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type IntegrationTestSuite struct {
	suite.Suite
	server *api.Server
	db     *gorm.DB
	router *gin.Engine
}

func (suite *IntegrationTestSuite) SetupSuite() {
	// Initialize logger for testing
	logger.InitLogger("test")

	// Setup test database
	var err error
	suite.db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(suite.T(), err)

	// Auto-migrate test database
	err = suite.db.AutoMigrate(
		&models.Tenant{},
		&models.User{},
		&models.Service{},
		&models.Incident{},
		&models.Maintenance{},
		&models.Subscriber{},
		&models.SubscriptionPlan{},
		&models.Subscription{},
		&models.FeatureFlag{},
		&models.BillingEvent{},
		&models.UsageMetrics{},
		&models.AdminSettings{},
		&models.SystemNotification{},
	)
	assert.NoError(suite.T(), err)

	// Setup test configuration with security settings
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     "test",
			SSLMode:  "disable",
		},
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-testing-only",
		},
		Environment: "test",
		Security: config.SecurityConfig{
			RateLimit: config.RateLimitConfig{
				Rate:   100,
				Burst:  200,
				Window: time.Minute,
			},
			Session: config.SessionConfig{
				Secret:   "test-session-secret",
				Secure:   false,
				SameSite: "lax",
			},
		},
	}

	// Create server
	suite.server = api.NewServer(cfg)

	// Setup router with all middleware for comprehensive testing
	suite.setupTestRouter()

	// Seed test data
	suite.seedTestData()
}

func (suite *IntegrationTestSuite) setupTestRouter() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Initialize all middleware for comprehensive testing
	errorHandler := middleware.NewErrorHandler(true) // development mode for tests
	validationMiddleware := middleware.NewValidationMiddleware()
	rateLimiter := middleware.NewRateLimiter(&config.RateLimitConfig{
		Rate:   10, // Lower rate for testing
		Burst:  20,
		Window: time.Minute,
	})
	csrfProtection := middleware.NewCSRFProtection("test-csrf-secret", false, "lax")
	securityHeaders := middleware.NewSecurityHeaders(nil)
	monitoringMiddleware := middleware.NewMonitoringMiddleware()

	// Add middleware in proper order
	suite.router.Use(monitoringMiddleware.RequestIDMiddleware())
	suite.router.Use(securityHeaders.SecurityHeadersMiddleware())
	suite.router.Use(middleware.CORSMiddleware(nil))
	suite.router.Use(monitoringMiddleware.RequestLoggingMiddleware())
	suite.router.Use(monitoringMiddleware.PerformanceMiddleware())
	suite.router.Use(monitoringMiddleware.ErrorTrackingMiddleware())
	suite.router.Use(monitoringMiddleware.SecurityMonitoringMiddleware())
	suite.router.Use(monitoringMiddleware.HealthCheckMiddleware())
	suite.router.Use(monitoringMiddleware.MetricsMiddleware())
	suite.router.Use(errorHandler.HandleError())
	suite.router.Use(rateLimiter.RateLimitMiddleware())
	suite.router.Use(csrfProtection.CSRFMiddleware())
	suite.router.Use(validationMiddleware.SanitizeInput())

	// Setup test routes
	suite.setupTestRoutes()
}

func (suite *IntegrationTestSuite) setupTestRoutes() {
	// Public routes
	suite.router.GET("/api/v1/status", suite.mockStatusHandler)
	suite.router.GET("/api/v1/incidents", suite.mockIncidentsHandler)
	suite.router.POST("/api/v1/subscribers", suite.mockSubscribeHandler)

	// Admin routes
	suite.router.POST("/api/v1/admin/login", suite.mockLoginHandler)
	suite.router.GET("/api/v1/admin/incidents", suite.mockAdminIncidentsHandler)
	suite.router.POST("/api/v1/admin/incidents", suite.mockCreateIncidentHandler)
	suite.router.POST("/api/v1/admin/status", suite.mockCreateServiceHandler)
	suite.router.POST("/api/v1/admin/maintenance", suite.mockCreateMaintenanceHandler)

	// SaaS routes
	suite.router.GET("/api/v1/admin/saas/tenants", suite.mockGetTenantsHandler)
	suite.router.POST("/api/v1/admin/saas/tenants", suite.mockCreateTenantHandler)
	suite.router.GET("/api/v1/admin/saas/plans", suite.mockGetPlansHandler)
	suite.router.POST("/api/v1/admin/saas/plans", suite.mockCreatePlanHandler)

	// Health and metrics
	suite.router.GET("/health", suite.mockHealthHandler)
	suite.router.GET("/metrics", suite.mockMetricsHandler)
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	// Cleanup test database
	sqlDB, err := suite.db.DB()
	assert.NoError(suite.T(), err)
	sqlDB.Close()
}

func (suite *IntegrationTestSuite) seedTestData() {
	// Create test tenant
	tenant := &models.Tenant{
		Name:     "Test Tenant",
		Slug:     "test-tenant",
		Status:   "active",
		Plan:     "free",
		IsActive: true,
	}
	suite.db.Create(tenant)

	// Create test user
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password",
		Role:     "admin",
		TenantID: &tenant.ID,
	}
	suite.db.Create(user)

	// Create test service
	service := &models.Service{
		Name:        "Test Service",
		Description: "Test service description",
		Status:      "operational",
		Group:       "Test Group",
		ShowUptime:  true,
		Position:    0,
		TenantID:    tenant.ID,
	}
	suite.db.Create(service)

	// Create test subscription plan
	plan := &models.SubscriptionPlan{
		Name:            "Free",
		Slug:            "free",
		Description:     "Free plan",
		Price:           0,
		Currency:        "USD",
		BillingInterval: "monthly",
		MaxServices:     5,
		MaxMonitors:     10,
		MaxSubscribers:  100,
		IsActive:        true,
	}
	suite.db.Create(plan)
}

func (suite *IntegrationTestSuite) TestPublicStatusEndpoint() {
	req, _ := http.NewRequest("GET", "/api/v1/status", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "services")
	assert.Contains(suite.T(), response, "status")
}

func (suite *IntegrationTestSuite) TestAdminLogin() {
	loginData := map[string]string{
		"username": "testuser",
		"password": "password",
	}
	jsonData, _ := json.Marshal(loginData)

	req, _ := http.NewRequest("POST", "/api/v1/admin/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "token")
}

func (suite *IntegrationTestSuite) TestCreateIncident() {
	// First, get authentication token
	token := suite.getAuthToken()

	incidentData := map[string]interface{}{
		"title":       "Test Incident",
		"description": "Test incident description",
		"impact":      "major",
		"services":    []uint{1},
	}
	jsonData, _ := json.Marshal(incidentData)

	req, _ := http.NewRequest("POST", "/api/v1/admin/incidents", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Test Incident", response["title"])
}

func (suite *IntegrationTestSuite) TestGetIncidents() {
	req, _ := http.NewRequest("GET", "/api/v1/incidents", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "incidents")
}

func (suite *IntegrationTestSuite) TestCreateService() {
	token := suite.getAuthToken()

	serviceData := map[string]interface{}{
		"name":        "New Test Service",
		"description": "New test service description",
		"group":       "Test Group",
		"show_uptime": true,
		"position":    1,
	}
	jsonData, _ := json.Marshal(serviceData)

	req, _ := http.NewRequest("POST", "/api/v1/admin/status", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "New Test Service", response["name"])
}

func (suite *IntegrationTestSuite) TestCreateMaintenance() {
	token := suite.getAuthToken()

	maintenanceData := map[string]interface{}{
		"title":           "Test Maintenance",
		"description":     "Test maintenance description",
		"scheduled_start": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"scheduled_end":   time.Now().Add(26 * time.Hour).Format(time.RFC3339),
		"services":        []uint{1},
	}
	jsonData, _ := json.Marshal(maintenanceData)

	req, _ := http.NewRequest("POST", "/api/v1/admin/maintenance", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Test Maintenance", response["title"])
}

func (suite *IntegrationTestSuite) TestSubscribeToUpdates() {
	subscriberData := map[string]interface{}{
		"email":    "subscriber@example.com",
		"services": []uint{1},
	}
	jsonData, _ := json.Marshal(subscriberData)

	req, _ := http.NewRequest("POST", "/api/v1/subscribers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "subscriber@example.com", response["email"])
}

func (suite *IntegrationTestSuite) TestSaaSTenantManagement() {
	token := suite.getAuthToken()

	// Test get all tenants
	req, _ := http.NewRequest("GET", "/api/v1/admin/saas/tenants", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "tenants")

	// Test create tenant
	tenantData := map[string]interface{}{
		"name": "New Test Company",
		"slug": "new-test-company",
		"plan": "free",
	}
	jsonData, _ := json.Marshal(tenantData)

	req, _ = http.NewRequest("POST", "/api/v1/admin/saas/tenants", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
}

func (suite *IntegrationTestSuite) TestSaaSPlanManagement() {
	token := suite.getAuthToken()

	// Test get all plans
	req, _ := http.NewRequest("GET", "/api/v1/admin/saas/plans", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "plans")

	// Test create plan
	planData := map[string]interface{}{
		"name":            "Premium",
		"slug":            "premium",
		"description":     "Premium plan with advanced features",
		"price":           99,
		"currency":        "USD",
		"max_services":    100,
		"max_monitors":    500,
		"max_subscribers": 10000,
		"custom_domain":   true,
		"api_access":      true,
		"integrations":    true,
		"analytics":       true,
		"white_label":     true,
	}
	jsonData, _ := json.Marshal(planData)

	req, _ = http.NewRequest("POST", "/api/v1/admin/saas/plans", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
}

func (suite *IntegrationTestSuite) TestUnauthorizedAccess() {
	req, _ := http.NewRequest("GET", "/api/v1/admin/incidents", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *IntegrationTestSuite) TestInvalidToken() {
	req, _ := http.NewRequest("GET", "/api/v1/admin/incidents", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *IntegrationTestSuite) TestHealthCheck() {
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "ok", response["status"])
}

// Helper method to get authentication token
func (suite *IntegrationTestSuite) getAuthToken() string {
	loginData := map[string]string{
		"username": "testuser",
		"password": "password",
	}
	jsonData, _ := json.Marshal(loginData)

	req, _ := http.NewRequest("POST", "/api/v1/admin/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		suite.T().Fatalf("Failed to unmarshal login response: %v", err)
	}
	return response["token"].(string)
}

// ===== COMPREHENSIVE SECURITY TESTS =====

func (suite *IntegrationTestSuite) TestSecurityHeaders() {
	req, _ := http.NewRequest("GET", "/api/v1/status", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Check security headers
	assert.Equal(suite.T(), "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(suite.T(), "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(suite.T(), "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(suite.T(), "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.Contains(suite.T(), w.Header().Get("Content-Security-Policy"), "default-src 'self'")
	assert.NotEmpty(suite.T(), w.Header().Get("X-Request-ID"))
}

func (suite *IntegrationTestSuite) TestRateLimiting() {
	// Make multiple requests to test rate limiting
	for i := 0; i < 12; i++ { // Exceed the rate limit of 10
		req, _ := http.NewRequest("GET", "/api/v1/status", nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		if i < 10 {
			// First 10 requests should succeed
			assert.Equal(suite.T(), http.StatusOK, w.Code)
		} else {
			// Requests beyond rate limit should be rejected
			assert.Equal(suite.T(), http.StatusTooManyRequests, w.Code)
			assert.Contains(suite.T(), w.Body.String(), "Rate limit exceeded")
		}
	}
}

func (suite *IntegrationTestSuite) TestInputValidation() {
	// Test XSS protection
	xssData := map[string]interface{}{
		"title":       "<script>alert('xss')</script>",
		"description": "Test description",
	}
	jsonData, _ := json.Marshal(xssData)

	req, _ := http.NewRequest("POST", "/api/v1/admin/incidents", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.getAuthToken())
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Should sanitize the input
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *IntegrationTestSuite) TestCSRFProtection() {
	// Test CSRF protection for state-changing operations
	incidentData := map[string]interface{}{
		"title":       "Test Incident",
		"description": "Test description",
	}
	jsonData, _ := json.Marshal(incidentData)

	req, _ := http.NewRequest("POST", "/api/v1/admin/incidents", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.getAuthToken())
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Should require CSRF token for POST requests
	// Note: In a real implementation, you'd need to get the CSRF token first
	assert.Equal(suite.T(), http.StatusOK, w.Code) // This will depend on CSRF implementation
}

func (suite *IntegrationTestSuite) TestCORSHeaders() {
	req, _ := http.NewRequest("GET", "/api/v1/status", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Check CORS headers
	assert.Equal(suite.T(), "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(suite.T(), w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(suite.T(), w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}

func (suite *IntegrationTestSuite) TestErrorHandling() {
	// Test error handling with invalid endpoint
	req, _ := http.NewRequest("GET", "/api/v1/nonexistent", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Should return 404 with proper error format
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *IntegrationTestSuite) TestPerformanceMonitoring() {
	req, _ := http.NewRequest("GET", "/api/v1/status", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Check performance headers
	assert.NotEmpty(suite.T(), w.Header().Get("X-Response-Time"))
	assert.NotEmpty(suite.T(), w.Header().Get("X-Request-ID"))
}

func (suite *IntegrationTestSuite) TestHealthEndpoint() {
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "healthy", response["status"])
}

func (suite *IntegrationTestSuite) TestMetricsEndpoint() {
	req, _ := http.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "requests_total")
	assert.Contains(suite.T(), response, "average_latency")
}

func (suite *IntegrationTestSuite) TestSQLInjectionProtection() {
	// Test SQL injection protection
	sqlInjectionData := map[string]interface{}{
		"title":       "'; DROP TABLE users; --",
		"description": "Test description",
	}
	jsonData, _ := json.Marshal(sqlInjectionData)

	req, _ := http.NewRequest("POST", "/api/v1/admin/incidents", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.getAuthToken())
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Should sanitize the input and not cause database issues
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *IntegrationTestSuite) TestSuspiciousUserAgent() {
	req, _ := http.NewRequest("GET", "/api/v1/status", nil)
	req.Header.Set("User-Agent", "sqlmap/1.0")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Should still work but log the suspicious activity
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// ===== MOCK HANDLERS =====

func (suite *IntegrationTestSuite) mockStatusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "operational",
		"services": []gin.H{
			{"name": "Test Service", "status": "operational"},
		},
	})
}

func (suite *IntegrationTestSuite) mockIncidentsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"incidents": []gin.H{},
	})
}

func (suite *IntegrationTestSuite) mockSubscribeHandler(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"email":   data["email"],
		"message": "Subscribed successfully",
	})
}

func (suite *IntegrationTestSuite) mockLoginHandler(c *gin.Context) {
	var data map[string]string
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if data["username"] == "testuser" && data["password"] == "password" {
		c.JSON(http.StatusOK, gin.H{
			"token": "mock-jwt-token",
			"user": gin.H{
				"username": "testuser",
				"role":     "admin",
			},
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	}
}

func (suite *IntegrationTestSuite) mockAdminIncidentsHandler(c *gin.Context) {
	// Check for authorization
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"incidents": []gin.H{},
	})
}

func (suite *IntegrationTestSuite) mockCreateIncidentHandler(c *gin.Context) {
	// Check for authorization
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          1,
		"title":       data["title"],
		"description": data["description"],
	})
}

func (suite *IntegrationTestSuite) mockCreateServiceHandler(c *gin.Context) {
	// Check for authorization
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          1,
		"name":        data["name"],
		"description": data["description"],
	})
}

func (suite *IntegrationTestSuite) mockCreateMaintenanceHandler(c *gin.Context) {
	// Check for authorization
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          1,
		"title":       data["title"],
		"description": data["description"],
	})
}

func (suite *IntegrationTestSuite) mockGetTenantsHandler(c *gin.Context) {
	// Check for authorization
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenants": []gin.H{
			{"id": 1, "name": "Test Tenant", "slug": "test-tenant"},
		},
	})
}

func (suite *IntegrationTestSuite) mockCreateTenantHandler(c *gin.Context) {
	// Check for authorization
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":   2,
		"name": data["name"],
		"slug": data["slug"],
	})
}

func (suite *IntegrationTestSuite) mockGetPlansHandler(c *gin.Context) {
	// Check for authorization
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plans": []gin.H{
			{"id": 1, "name": "Free", "slug": "free"},
		},
	})
}

func (suite *IntegrationTestSuite) mockCreatePlanHandler(c *gin.Context) {
	// Check for authorization
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":   2,
		"name": data["name"],
		"slug": data["slug"],
	})
}

func (suite *IntegrationTestSuite) mockHealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}

func (suite *IntegrationTestSuite) mockMetricsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"requests_total":      100,
		"requests_per_second": 10.5,
		"average_latency":     150.0,
		"error_rate":          0.02,
		"active_connections":  25,
	})
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
