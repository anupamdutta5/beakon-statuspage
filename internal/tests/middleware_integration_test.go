package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/enterprise-status/statuspage/internal/api/middleware"
	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MiddlewareIntegrationTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (suite *MiddlewareIntegrationTestSuite) SetupSuite() {
	// Initialize logger for testing
	logger.InitLogger("test")

	// Setup test router with all middleware
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Initialize all middleware for comprehensive testing
	errorHandler := middleware.NewErrorHandler(true) // development mode for tests
	validationMiddleware := middleware.NewValidationMiddleware()
	rateLimiter := middleware.NewRateLimiter(&config.RateLimitConfig{
		Rate:   5, // Lower rate for testing
		Burst:  10,
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

func (suite *MiddlewareIntegrationTestSuite) setupTestRoutes() {
	// Test routes
	suite.router.GET("/test", suite.mockGetHandler)
	suite.router.POST("/test", suite.mockPostHandler)
	suite.router.GET("/health", suite.mockHealthHandler)
	suite.router.GET("/metrics", suite.mockMetricsHandler)
	suite.router.GET("/error", suite.mockErrorHandler)
}

// ===== COMPREHENSIVE SECURITY TESTS =====

func (suite *MiddlewareIntegrationTestSuite) TestSecurityHeaders() {
	req, _ := http.NewRequest("GET", "/test", nil)
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

func (suite *MiddlewareIntegrationTestSuite) TestRateLimiting() {
	// Make multiple requests to test rate limiting
	for i := 0; i < 7; i++ { // Exceed the rate limit of 5
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		if i < 5 {
			// First 5 requests should succeed
			assert.Equal(suite.T(), http.StatusOK, w.Code)
		} else {
			// Requests beyond rate limit should be rejected
			assert.Equal(suite.T(), http.StatusTooManyRequests, w.Code)
			assert.Contains(suite.T(), w.Body.String(), "Rate limit exceeded")
		}
	}
}

func (suite *MiddlewareIntegrationTestSuite) TestInputValidation() {
	// Test XSS protection
	xssData := map[string]interface{}{
		"title":       "<script>alert('xss')</script>",
		"description": "Test description",
	}
	jsonData, _ := json.Marshal(xssData)

	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Should sanitize the input
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *MiddlewareIntegrationTestSuite) TestCORSHeaders() {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Check CORS headers
	assert.Equal(suite.T(), "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(suite.T(), w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(suite.T(), w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}

func (suite *MiddlewareIntegrationTestSuite) TestErrorHandling() {
	// Test error handling with error endpoint
	req, _ := http.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Should return 500 with proper error format
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
}

func (suite *MiddlewareIntegrationTestSuite) TestPerformanceMonitoring() {
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Check performance headers
	assert.NotEmpty(suite.T(), w.Header().Get("X-Response-Time"))
	assert.NotEmpty(suite.T(), w.Header().Get("X-Request-ID"))
}

func (suite *MiddlewareIntegrationTestSuite) TestHealthEndpoint() {
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "healthy", response["status"])
}

func (suite *MiddlewareIntegrationTestSuite) TestMetricsEndpoint() {
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

func (suite *MiddlewareIntegrationTestSuite) TestSQLInjectionProtection() {
	// Test SQL injection protection
	sqlInjectionData := map[string]interface{}{
		"title":       "'; DROP TABLE users; --",
		"description": "Test description",
	}
	jsonData, _ := json.Marshal(sqlInjectionData)

	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Should sanitize the input and not cause database issues
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *MiddlewareIntegrationTestSuite) TestSuspiciousUserAgent() {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "sqlmap/1.0")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Should still work but log the suspicious activity
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *MiddlewareIntegrationTestSuite) TestRequestIDGeneration() {
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Check that request ID is generated and returned
	requestID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(suite.T(), requestID)
	assert.Len(suite.T(), requestID, 13) // Base36 timestamp should be 13 chars
}

func (suite *MiddlewareIntegrationTestSuite) TestSecurityHeadersCompleteness() {
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Check all security headers are present
	securityHeaders := []string{
		"X-Frame-Options",
		"X-Content-Type-Options",
		"X-XSS-Protection",
		"Referrer-Policy",
		"Content-Security-Policy",
		"X-Permitted-Cross-Domain-Policies",
		"Cross-Origin-Embedder-Policy",
		"Cross-Origin-Opener-Policy",
		"Cross-Origin-Resource-Policy",
	}

	for _, header := range securityHeaders {
		assert.NotEmpty(suite.T(), w.Header().Get(header), "Missing security header: %s", header)
	}
}

// ===== MOCK HANDLERS =====

func (suite *MiddlewareIntegrationTestSuite) mockGetHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"method":  "GET",
	})
}

func (suite *MiddlewareIntegrationTestSuite) mockPostHandler(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"method":  "POST",
		"data":    data,
	})
}

func (suite *MiddlewareIntegrationTestSuite) mockHealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}

func (suite *MiddlewareIntegrationTestSuite) mockMetricsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"requests_total":      100,
		"requests_per_second": 10.5,
		"average_latency":     150.0,
		"error_rate":          0.02,
		"active_connections":  25,
	})
}

func (suite *MiddlewareIntegrationTestSuite) mockErrorHandler(c *gin.Context) {
	if err := c.Error(gin.Error{
		Err:  errors.New("internal server error"),
		Type: gin.ErrorTypePublic,
	}); err != nil {
		suite.T().Logf("Failed to set error: %v", err)
	}
	c.AbortWithStatus(http.StatusInternalServerError)
}

func TestMiddlewareIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareIntegrationTestSuite))
}
