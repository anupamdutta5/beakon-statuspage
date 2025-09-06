package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/enterprise-status/statuspage/internal/api"
	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/internal/models"
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

	// Setup test configuration
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "test",
			Password: "test",
			DBName:   "test",
			SSLMode:  "disable",
		},
		JWT: config.JWTConfig{
			Secret: "test-secret",
		},
		Environment: "test",
	}

	// Create server
	suite.server = api.NewServer(cfg)
	suite.router = suite.server.GetRouter()

	// Seed test data
	suite.seedTestData()
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
		TenantID: tenant.ID,
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
	json.Unmarshal(w.Body.Bytes(), &response)
	return response["token"].(string)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
