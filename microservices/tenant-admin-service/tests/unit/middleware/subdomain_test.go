package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"tenant-admin-service/internal/middleware"
	"tenant-admin-service/internal/models"
)

// MockTenantService for middleware testing
type MockTenantService struct {
	mock.Mock
}

func (m *MockTenantService) GetTenantBySubdomain(subdomain string) (*models.Tenant, error) {
	args := m.Called(subdomain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestSubdomainMiddleware_ValidSubdomain(t *testing.T) {
	// Arrange
	mockService := new(MockTenantService)
	router := setupTestRouter()

	expectedTenant := &models.Tenant{
		ID:        "tenant-1",
		Name:      "Test Tenant",
		Subdomain: "test",
		Status:    "active",
	}

	mockService.On("GetTenantBySubdomain", "test").Return(expectedTenant, nil)

	router.Use(middleware.SubdomainMiddleware(mockService))
	router.GET("/test", func(c *gin.Context) {
		tenant, exists := c.Get("tenant")
		assert.True(t, exists)
		assert.Equal(t, expectedTenant, tenant)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "test.localhost:3002"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestSubdomainMiddleware_NoSubdomain(t *testing.T) {
	// Arrange
	mockService := new(MockTenantService)
	router := setupTestRouter()

	router.Use(middleware.SubdomainMiddleware(mockService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "localhost:3002" // No subdomain
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "subdomain is required")
}

func TestSubdomainMiddleware_TenantNotFound(t *testing.T) {
	// Arrange
	mockService := new(MockTenantService)
	router := setupTestRouter()

	mockService.On("GetTenantBySubdomain", "nonexistent").Return(nil, assert.AnError)

	router.Use(middleware.SubdomainMiddleware(mockService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "nonexistent.localhost:3002"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "tenant not found")
	mockService.AssertExpectations(t)
}

func TestSubdomainMiddleware_InactiveTenant(t *testing.T) {
	// Arrange
	mockService := new(MockTenantService)
	router := setupTestRouter()

	inactiveTenant := &models.Tenant{
		ID:        "tenant-1",
		Name:      "Inactive Tenant",
		Subdomain: "inactive",
		Status:    "inactive",
	}

	mockService.On("GetTenantBySubdomain", "inactive").Return(inactiveTenant, nil)

	router.Use(middleware.SubdomainMiddleware(mockService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "inactive.localhost:3002"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "tenant is not active")
	mockService.AssertExpectations(t)
}

func TestSubdomainMiddleware_SuspendedTenant(t *testing.T) {
	// Arrange
	mockService := new(MockTenantService)
	router := setupTestRouter()

	suspendedTenant := &models.Tenant{
		ID:        "tenant-1",
		Name:      "Suspended Tenant",
		Subdomain: "suspended",
		Status:    "suspended",
	}

	mockService.On("GetTenantBySubdomain", "suspended").Return(suspendedTenant, nil)

	router.Use(middleware.SubdomainMiddleware(mockService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "suspended.localhost:3002"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "tenant is suspended")
	mockService.AssertExpectations(t)
}

func TestExtractSubdomain(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		expected string
	}{
		{"Subdomain with port", "test.localhost:3002", "test"},
		{"Subdomain without port", "test.example.com", "test"},
		{"No subdomain with port", "localhost:3002", ""},
		{"No subdomain without port", "example.com", ""},
		{"Multiple subdomains", "api.test.example.com", "api"},
		{"Localhost only", "localhost", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			subdomain := middleware.ExtractSubdomain(tt.host)

			// Assert
			assert.Equal(t, tt.expected, subdomain)
		})
	}
}

func TestSubdomainMiddleware_SkipPublicRoutes(t *testing.T) {
	// Arrange
	mockService := new(MockTenantService)
	router := setupTestRouter()

	router.Use(middleware.SubdomainMiddleware(mockService))
	router.POST("/api/v1/public/tenants", func(c *gin.Context) {
		// Public route should not require subdomain
		_, exists := c.Get("tenant")
		assert.False(t, exists) // No tenant context for public routes
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/tenants", nil)
	req.Host = "localhost:3002" // No subdomain
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	// No mock assertions as service should not be called for public routes
}

func TestSubdomainMiddleware_TenantContextAvailable(t *testing.T) {
	// Arrange
	mockService := new(MockTenantService)
	router := setupTestRouter()

	expectedTenant := &models.Tenant{
		ID:        "tenant-1",
		Name:      "Test Tenant",
		Subdomain: "test",
		Status:    "active",
		MaxUsers:  10,
	}

	mockService.On("GetTenantBySubdomain", "test").Return(expectedTenant, nil)

	router.Use(middleware.SubdomainMiddleware(mockService))
	router.GET("/api/v1/users", func(c *gin.Context) {
		// Verify tenant context is available
		tenant, exists := c.Get("tenant")
		assert.True(t, exists)

		tenantModel := tenant.(*models.Tenant)
		assert.Equal(t, "tenant-1", tenantModel.ID)
		assert.Equal(t, "test", tenantModel.Subdomain)
		assert.Equal(t, 10, tenantModel.MaxUsers)

		// Verify tenant_id is also set for convenience
		tenantID, exists := c.Get("tenant_id")
		assert.True(t, exists)
		assert.Equal(t, "tenant-1", tenantID)

		c.JSON(http.StatusOK, gin.H{"tenant_id": tenantID})
	})

	// Act
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Host = "test.localhost:3002"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "tenant-1")
	mockService.AssertExpectations(t)
}
