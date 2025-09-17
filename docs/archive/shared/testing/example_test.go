// Package testing provides an example of how to use the shared testing utilities.
package testing

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

// Example of how to use the shared testing utilities
func ExampleHTTPTest(t *testing.T) {
	// Setup test router
	router := SetupTestRouter()

	// Add a simple route for testing
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "example-service",
		})
	})

	// Define test cases
	testCases := []TestCase{
		{
			Name: "Health check returns 200",
			Request: HTTPTestRequest{
				Method: "GET",
				URL:    "/health",
			},
			Expected: HTTPTestResponse{
				StatusCode: http.StatusOK,
				Body: gin.H{
					"status":  "healthy",
					"service": "example-service",
				},
			},
		},
		{
			Name: "Non-existent endpoint returns 404",
			Request: HTTPTestRequest{
				Method: "GET",
				URL:    "/nonexistent",
			},
			Expected: HTTPTestResponse{
				StatusCode: http.StatusNotFound,
			},
		},
	}

	// Run all test cases
	RunHTTPTests(t, router, testCases)
}

// Example of how to use database testing
func ExampleDatabaseTest(t *testing.T) {
	// Setup test database
	db := SetupTestDB(t)

	// Your database testing logic here
	// db.AutoMigrate(&YourModel{})
	// ... test your database operations

	// The database will be automatically cleaned up after the test
	_ = db
}

// Example of how to benchmark HTTP requests
func BenchmarkHealthEndpoint(b *testing.B) {
	router := SetupTestRouter()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	req := HTTPTestRequest{
		Method: "GET",
		URL:    "/health",
	}

	BenchmarkHTTPRequest(b, router, req)
}