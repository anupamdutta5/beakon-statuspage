// Package testing provides shared testing utilities for Beakon microservices.
package testing

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestConfig provides configuration for tests
type TestConfig struct {
	DatabaseURL   string
	RedisURL      string
	Port          int
	LogLevel      string
	EnableLogging bool
}

// DefaultTestConfig returns a default test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		DatabaseURL:   ":memory:",
		RedisURL:      "redis://localhost:6379",
		Port:          8080,
		LogLevel:      "error",
		EnableLogging: false,
	}
}

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "Failed to connect to test database")
	return db
}

// SetupTestRouter creates a test Gin router with error mode disabled
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// HTTPTestRequest represents a test HTTP request
type HTTPTestRequest struct {
	Method  string
	URL     string
	Body    interface{}
	Headers map[string]string
	Params  map[string]string
}

// HTTPTestResponse represents an expected HTTP response
type HTTPTestResponse struct {
	StatusCode int
	Body       interface{}
	Headers    map[string]string
}

// ExecuteRequest executes an HTTP request against a handler and returns the response
func ExecuteRequest(t *testing.T, router *gin.Engine, req HTTPTestRequest) *httptest.ResponseRecorder {
	var body io.Reader
	if req.Body != nil {
		if bodyStr, ok := req.Body.(string); ok {
			body = strings.NewReader(bodyStr)
		} else {
			jsonBody, err := json.Marshal(req.Body)
			require.NoError(t, err, "Failed to marshal request body")
			body = strings.NewReader(string(jsonBody))
		}
	}

	request := httptest.NewRequest(req.Method, req.URL, body)

	// Set headers
	if req.Headers != nil {
		for key, value := range req.Headers {
			request.Header.Set(key, value)
		}
	}

	// Set default content type for POST/PUT requests with body
	if body != nil && request.Header.Get("Content-Type") == "" {
		request.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	return recorder
}

// AssertResponse asserts that the HTTP response matches expectations
func AssertResponse(t *testing.T, recorder *httptest.ResponseRecorder, expected HTTPTestResponse) {
	assert.Equal(t, expected.StatusCode, recorder.Code, "Status code mismatch")

	if expected.Headers != nil {
		for key, expectedValue := range expected.Headers {
			assert.Equal(t, expectedValue, recorder.Header().Get(key), fmt.Sprintf("Header %s mismatch", key))
		}
	}

	if expected.Body != nil {
		if expectedStr, ok := expected.Body.(string); ok {
			assert.JSONEq(t, expectedStr, recorder.Body.String(), "Response body mismatch")
		} else {
			expectedJSON, err := json.Marshal(expected.Body)
			require.NoError(t, err, "Failed to marshal expected body")
			assert.JSONEq(t, string(expectedJSON), recorder.Body.String(), "Response body mismatch")
		}
	}
}

// TestCase represents a complete test case
type TestCase struct {
	Name        string
	Setup       func(t *testing.T)
	Request     HTTPTestRequest
	Expected    HTTPTestResponse
	Cleanup     func(t *testing.T)
	Description string
}

// RunHTTPTest runs a complete HTTP test case
func RunHTTPTest(t *testing.T, router *gin.Engine, testCase TestCase) {
	t.Run(testCase.Name, func(t *testing.T) {
		if testCase.Setup != nil {
			testCase.Setup(t)
		}

		recorder := ExecuteRequest(t, router, testCase.Request)
		AssertResponse(t, recorder, testCase.Expected)

		if testCase.Cleanup != nil {
			testCase.Cleanup(t)
		}
	})
}

// RunHTTPTests runs multiple HTTP test cases
func RunHTTPTests(t *testing.T, router *gin.Engine, testCases []TestCase) {
	for _, testCase := range testCases {
		RunHTTPTest(t, router, testCase)
	}
}

// WaitForService waits for a service to be available at the given URL
func WaitForService(url string, timeout time.Duration) error {
	client := &http.Client{
		Timeout: time.Second * 2,
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("service at %s not available after %v", url, timeout)
}

// MockService creates a mock HTTP service for integration testing
func MockService(port int, handler http.Handler) *httptest.Server {
	server := httptest.NewServer(handler)
	return server
}

// SetTestEnv sets environment variables for testing
func SetTestEnv(envVars map[string]string) func() {
	originalVars := make(map[string]string)

	for key, value := range envVars {
		if original := os.Getenv(key); original != "" {
			originalVars[key] = original
		}
		os.Setenv(key, value)
	}

	// Return cleanup function
	return func() {
		for key := range envVars {
			if original, exists := originalVars[key]; exists {
				os.Setenv(key, original)
			} else {
				os.Unsetenv(key)
			}
		}
	}
}

// BenchmarkHTTPRequest benchmarks an HTTP request
func BenchmarkHTTPRequest(b *testing.B, router *gin.Engine, req HTTPTestRequest) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()

		var body io.Reader
		if req.Body != nil {
			if bodyStr, ok := req.Body.(string); ok {
				body = strings.NewReader(bodyStr)
			} else {
				jsonBody, _ := json.Marshal(req.Body)
				body = strings.NewReader(string(jsonBody))
			}
		}

		request := httptest.NewRequest(req.Method, req.URL, body)
		if req.Headers != nil {
			for key, value := range req.Headers {
				request.Header.Set(key, value)
			}
		}

		router.ServeHTTP(recorder, request)
	}
}