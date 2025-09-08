package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid user registration",
			method:         "POST",
			path:           "/api/users",
			body:           map[string]interface{}{"username": "testuser", "email": "test@example.com", "password": "SecurePass123!"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid email format",
			method:         "POST",
			path:           "/api/users",
			body:           map[string]interface{}{"username": "testuser", "email": "invalid-email", "password": "SecurePass123!"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation failed",
		},
		{
			name:           "Weak password",
			method:         "POST",
			path:           "/api/users",
			body:           map[string]interface{}{"username": "testuser", "email": "test@example.com", "password": "123"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation failed",
		},
		{
			name:           "XSS attempt in username",
			method:         "POST",
			path:           "/api/users",
			body:           map[string]interface{}{"username": "<script>alert('xss')</script>", "email": "test@example.com", "password": "SecurePass123!"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation failed",
		},
		{
			name:           "SQL injection attempt",
			method:         "POST",
			path:           "/api/users",
			body:           map[string]interface{}{"username": "admin'; DROP TABLE users; --", "email": "test@example.com", "password": "SecurePass123!"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup router
			router := gin.New()
			validationMiddleware := NewValidationMiddleware()

			// Add validation middleware
			router.Use(validationMiddleware.SanitizeInput())

			// Add test endpoint
			router.POST("/api/users", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Prepare request
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			// Perform request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			}
		})
	}
}

func TestCustomValidators(t *testing.T) {
	validator := NewValidationMiddleware()

	tests := []struct {
		name      string
		validator string
		value     string
		expected  bool
	}{
		{"Valid alphanumeric", "alphanumeric", "test123", true},
		{"Invalid alphanumeric with special chars", "alphanumeric", "test@123", false},
		{"Valid slug", "slug", "test-slug", true},
		{"Invalid slug with spaces", "slug", "test slug", false},
		{"Valid email", "email", "test@example.com", true},
		{"Invalid email", "email", "invalid-email", false},
		{"Strong password", "password", "SecurePass123!", true},
		{"Weak password", "password", "123", false},
		{"Valid URL", "url", "https://example.com", true},
		{"Invalid URL", "url", "not-a-url", false},
		{"Valid phone", "phone", "+1234567890", true},
		{"Invalid phone", "phone", "abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validator.Var(tt.value, tt.validator)
			if tt.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestSanitization(t *testing.T) {
	validationMiddleware := NewValidationMiddleware()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"HTML tags", "<script>alert('xss')</script>", "alert('xss')"},
		{"SQL injection", "admin'; DROP TABLE users; --", "admin DROP TABLE users"},
		{"Normal text", "hello world", "hello world"},
		{"Mixed content", "Hello <b>world</b>!", "Hello world!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validationMiddleware.sanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
