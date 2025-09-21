package validation

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	resilience "github.com/anupamdutta5/shared-resilience"
)

// ValidatorConfig represents validation configuration
type ValidatorConfig struct {
	MaxStringLength int
	MaxArrayLength  int
	StrictMode      bool
}

// ServiceValidator provides validation for the service
type ServiceValidator struct {
	config    ValidatorConfig
	sanitizer *resilience.InputSanitizer
	logger    *zap.Logger
}

// NewServiceValidator creates a new service validator
func NewServiceValidator(logger *zap.Logger) *ServiceValidator {
	config := ValidatorConfig{
		MaxStringLength: 1000,
		MaxArrayLength:  100,
		StrictMode:      true,
	}

	sanitizer := resilience.NewInputSanitizer(resilience.SanitizationConfig{
		MaxStringLength: config.MaxStringLength,
		AllowedTags:     []string{},
		BlockedPatterns: []string{
			`<script[^>]*>.*?</script>`,           // Script tags
			`javascript:`,                         // JavaScript URLs
			`on\w+\s*=`,                          // Event handlers
			`\bexec\b|\beval\b|\balert\b`,        // Dangerous JS functions
			`<iframe[^>]*>.*?</iframe>`,          // iframes
			`data:text/html`,                     // Data URLs with HTML
		},
		StrictMode: config.StrictMode,
	}, logger)

	return &ServiceValidator{
		config:    config,
		sanitizer: sanitizer,
		logger:    logger,
	}
}

// ValidateAndSanitizeString validates and sanitizes a string input
func (v *ServiceValidator) ValidateAndSanitizeString(input string, fieldName string, required bool) (string, error) {
	// Check if required field is empty
	if required && strings.TrimSpace(input) == "" {
		return "", fmt.Errorf("%s is required", fieldName)
	}

	// Sanitize the input
	sanitized := v.sanitizer.SanitizeString(input)

	// Additional validation
	if len(sanitized) > v.config.MaxStringLength {
		return "", fmt.Errorf("%s exceeds maximum length of %d characters", fieldName, v.config.MaxStringLength)
	}

	// Check for valid UTF-8
	if !utf8.ValidString(sanitized) {
		return "", fmt.Errorf("%s contains invalid UTF-8 characters", fieldName)
	}

	return sanitized, nil
}

// ValidateEmail validates and sanitizes email input
func (v *ServiceValidator) ValidateEmail(email string) (string, error) {
	if email == "" {
		return "", fmt.Errorf("email is required")
	}

	sanitized, err := v.sanitizer.SanitizeEmail(email)
	if err != nil {
		return "", fmt.Errorf("invalid email format: %w", err)
	}

	return sanitized, nil
}

// ValidateURL validates and sanitizes URL input
func (v *ServiceValidator) ValidateURL(url string) (string, error) {
	if url == "" {
		return "", nil // URL is optional in most cases
	}

	sanitized, err := v.sanitizer.SanitizeURL(url)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %w", err)
	}

	return sanitized, nil
}

// ValidateID validates numeric ID
func (v *ServiceValidator) ValidateID(id interface{}, fieldName string) (uint, error) {
	switch val := id.(type) {
	case uint:
		if val == 0 {
			return 0, fmt.Errorf("%s must be greater than 0", fieldName)
		}
		return val, nil
	case int:
		if val <= 0 {
			return 0, fmt.Errorf("%s must be greater than 0", fieldName)
		}
		return uint(val), nil
	case float64:
		if val <= 0 {
			return 0, fmt.Errorf("%s must be greater than 0", fieldName)
		}
		return uint(val), nil
	default:
		return 0, fmt.Errorf("%s must be a valid positive integer", fieldName)
	}
}

// ValidateEnum validates enum values
func (v *ServiceValidator) ValidateEnum(value string, allowedValues []string, fieldName string) (string, error) {
	sanitized := v.sanitizer.SanitizeString(value)

	for _, allowed := range allowedValues {
		if strings.EqualFold(sanitized, allowed) {
			return allowed, nil // Return canonical form
		}
	}

	return "", fmt.Errorf("%s must be one of: %s", fieldName, strings.Join(allowedValues, ", "))
}

// ValidateRange validates numeric range
func (v *ServiceValidator) ValidateRange(value interface{}, min, max float64, fieldName string) error {
	var numValue float64

	switch val := value.(type) {
	case int:
		numValue = float64(val)
	case uint:
		numValue = float64(val)
	case float64:
		numValue = val
	case float32:
		numValue = float64(val)
	default:
		return fmt.Errorf("%s must be a valid number", fieldName)
	}

	if numValue < min || numValue > max {
		return fmt.Errorf("%s must be between %v and %v", fieldName, min, max)
	}

	return nil
}

// ValidateStringLength validates string length
func (v *ServiceValidator) ValidateStringLength(value string, minLen, maxLen int, fieldName string) error {
	length := len(value)

	if length < minLen {
		return fmt.Errorf("%s must be at least %d characters", fieldName, minLen)
	}

	if maxLen > 0 && length > maxLen {
		return fmt.Errorf("%s cannot exceed %d characters", fieldName, maxLen)
	}

	return nil
}

// SanitizeMap sanitizes all string values in a map
func (v *ServiceValidator) SanitizeMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		switch val := value.(type) {
		case string:
			result[key] = v.sanitizer.SanitizeString(val)
		case map[string]interface{}:
			result[key] = v.SanitizeMap(val)
		default:
			result[key] = value
		}
	}

	return result
}

// MiddlewareFunc returns a Gin middleware function for validation
func (v *ServiceValidator) MiddlewareFunc() gin.HandlerFunc {
	return v.sanitizer.GinMiddleware()
}
