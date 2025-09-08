package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test config
	rateLimitConfig := &config.RateLimitConfig{
		Rate:   5,
		Burst:  10,
		Window: time.Minute,
	}

	// Create rate limiter for testing
	rateLimiter := NewRateLimiter(rateLimitConfig)

	tests := []struct {
		name           string
		requests       int
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Within rate limit",
			requests:       3,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Exceed rate limit",
			requests:       6,
			expectedStatus: http.StatusTooManyRequests,
			expectedError:  "Rate limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup router
			router := gin.New()
			router.Use(rateLimiter.RateLimitMiddleware())

			// Add test endpoint
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Make multiple requests
			for i := 0; i < tt.requests; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if i < tt.requests-1 {
					// All requests except the last should succeed
					assert.Equal(t, http.StatusOK, w.Code)
				} else {
					// Last request should match expected status
					assert.Equal(t, tt.expectedStatus, w.Code)

					if tt.expectedError != "" {
						assert.Contains(t, w.Body.String(), tt.expectedError)
					}
				}
			}
		})
	}
}

func TestInMemoryRateLimitStore(t *testing.T) {
	// Skip this test for now as we need to implement the store interface
	t.Skip("InMemoryRateLimitStore test needs to be implemented")
}
