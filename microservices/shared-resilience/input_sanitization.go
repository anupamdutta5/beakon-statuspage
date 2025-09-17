// Package resilience provides input sanitization and validation middleware
package resilience

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SanitizationConfig represents configuration for input sanitization
type SanitizationConfig struct {
	MaxStringLength int      `yaml:"max_string_length" default:"1000"`
	AllowedTags     []string `yaml:"allowed_tags"`
	BlockedPatterns []string `yaml:"blocked_patterns"`
	StrictMode      bool     `yaml:"strict_mode" default:"false"`
}

// InputSanitizer provides input sanitization functionality
type InputSanitizer struct {
	config SanitizationConfig
	logger *zap.Logger
	// Compiled regex patterns for performance
	blockedPatterns []*regexp.Regexp
}

// NewInputSanitizer creates a new input sanitizer
func NewInputSanitizer(config SanitizationConfig, logger *zap.Logger) *InputSanitizer {
	// Compile blocked patterns for performance
	var compiledPatterns []*regexp.Regexp
	for _, pattern := range config.BlockedPatterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			compiledPatterns = append(compiledPatterns, compiled)
		} else {
			logger.Warn("Failed to compile blocked pattern",
				zap.String("pattern", pattern),
				zap.Error(err))
		}
	}

	return &InputSanitizer{
		config:          config,
		logger:          logger,
		blockedPatterns: compiledPatterns,
	}
}

// SanitizeString sanitizes a string input
func (s *InputSanitizer) SanitizeString(input string) string {
	if input == "" {
		return input
	}

	// Trim whitespace
	input = strings.TrimSpace(input)

	// Check length
	if len(input) > s.config.MaxStringLength {
		input = input[:s.config.MaxStringLength]
		s.logger.Warn("Input truncated due to length limit",
			zap.Int("max_length", s.config.MaxStringLength))
	}

	// Check for blocked patterns
	for _, pattern := range s.blockedPatterns {
		if pattern.MatchString(input) {
			s.logger.Warn("Input blocked due to pattern match",
				zap.String("pattern", pattern.String()))
			return ""
		}
	}

	// HTML escape
	input = html.EscapeString(input)

	// Remove control characters
	input = s.removeControlCharacters(input)

	return input
}

// SanitizeURL sanitizes a URL input
func (s *InputSanitizer) SanitizeURL(input string) (string, error) {
	if input == "" {
		return input, nil
	}

	// Parse URL
	parsedURL, err := url.Parse(input)
	if err != nil {
		return "", err
	}

	// Only allow http and https schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", fmt.Errorf("invalid URL scheme: %s", parsedURL.Scheme)
	}

	// Sanitize path and query
	parsedURL.Path = s.SanitizeString(parsedURL.Path)
	parsedURL.RawQuery = s.SanitizeString(parsedURL.RawQuery)

	return parsedURL.String(), nil
}

// SanitizeEmail sanitizes an email input
func (s *InputSanitizer) SanitizeEmail(input string) (string, error) {
	if input == "" {
		return input, nil
	}

	// Basic email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	sanitized := s.SanitizeString(input)
	if !emailRegex.MatchString(sanitized) {
		return "", fmt.Errorf("invalid email format")
	}

	return sanitized, nil
}

// removeControlCharacters removes control characters from string
func (s *InputSanitizer) removeControlCharacters(input string) string {
	var result strings.Builder
	for _, r := range input {
		if unicode.IsPrint(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// GinMiddleware returns a Gin middleware for input sanitization
func (s *InputSanitizer) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		query := c.Request.URL.Query()
		for _, values := range query {
			for i, value := range values {
				values[i] = s.SanitizeString(value)
			}
		}
		c.Request.URL.RawQuery = query.Encode()

		// Sanitize form data
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if err := c.Request.ParseForm(); err == nil {
				form := c.Request.PostForm
				for _, values := range form {
					for i, value := range values {
						values[i] = s.SanitizeString(value)
					}
				}
			}
		}

		// Sanitize headers (basic sanitization for security headers)
		for key, values := range c.Request.Header {
			if strings.HasPrefix(strings.ToLower(key), "x-") {
				for i, value := range values {
					values[i] = s.SanitizeString(value)
				}
			}
		}

		c.Next()
	}
}

// InputValidationError represents a validation error
type InputValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// Error implements the error interface
func (e InputValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid     bool                    `json:"valid"`
	Errors    []InputValidationError  `json:"errors,omitempty"`
	Sanitized map[string]string       `json:"sanitized,omitempty"`
}

// ValidateAndSanitize validates and sanitizes input data
func (s *InputSanitizer) ValidateAndSanitize(data map[string]interface{}) ValidationResult {
	result := ValidationResult{
		Valid:     true,
		Errors:    []InputValidationError{},
		Sanitized: make(map[string]string),
	}

	for key, value := range data {
		if str, ok := value.(string); ok {
			sanitized := s.SanitizeString(str)
			result.Sanitized[key] = sanitized

			// Additional validation based on field name
			if err := s.validateField(key, sanitized); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, InputValidationError{
					Field:   key,
					Message: err.Error(),
					Value:   sanitized,
				})
			}
		}
	}

	return result
}

// validateField performs field-specific validation
func (s *InputSanitizer) validateField(fieldName, value string) error {
	fieldName = strings.ToLower(fieldName)

	// Email validation
	if strings.Contains(fieldName, "email") {
		if _, err := s.SanitizeEmail(value); err != nil {
			return err
		}
	}

	// URL validation
	if strings.Contains(fieldName, "url") || strings.Contains(fieldName, "link") {
		if _, err := s.SanitizeURL(value); err != nil {
			return err
		}
	}

	// Required field validation
	if strings.Contains(fieldName, "required") && value == "" {
		return fmt.Errorf("field is required")
	}

	// Length validation
	if len(value) > s.config.MaxStringLength {
		return fmt.Errorf("field exceeds maximum length of %d", s.config.MaxStringLength)
	}

	return nil
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// RateLimitMiddleware provides basic rate limiting
func RateLimitMiddleware(maxRequests int, window time.Duration) gin.HandlerFunc {
	// Simple in-memory rate limiter (for production, use Redis)
	clients := make(map[string][]time.Time)
	var mu sync.RWMutex

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		mu.Lock()
		defer mu.Unlock()

		// Clean old requests
		if requests, exists := clients[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < window {
					validRequests = append(validRequests, reqTime)
				}
			}
			clients[clientIP] = validRequests
		}

		// Check rate limit
		if len(clients[clientIP]) >= maxRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		// Add current request
		clients[clientIP] = append(clients[clientIP], now)
		c.Next()
	}
}
