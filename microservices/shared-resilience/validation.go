package resilience

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"go.uber.org/zap"
)

// ValidationManager provides centralized validation utilities
type ValidationManager struct {
	config ValidationConfig
	logger *zap.Logger
}

// ValidationConfig represents validation configuration
type ValidationConfig struct {
	MaxStringLength int      `yaml:"max_string_length" env:"VALIDATION_MAX_STRING_LENGTH" default:"1000"`
	MaxArrayLength  int      `yaml:"max_array_length" env:"VALIDATION_MAX_ARRAY_LENGTH" default:"100"`
	AllowedDomains  []string `yaml:"allowed_domains" env:"VALIDATION_ALLOWED_DOMAINS"`
	BlockedPatterns []string `yaml:"blocked_patterns"`
	StrictMode      bool     `yaml:"strict_mode" env:"VALIDATION_STRICT_MODE" default:"true"`
}

// Use existing types from other files
// ValidationError and ValidationResult are already defined in input_sanitization.go

// NewValidationManager creates a new validation manager
func NewValidationManager(config ValidationConfig, logger *zap.Logger) *ValidationManager {
	if config.MaxStringLength == 0 {
		config.MaxStringLength = 1000
	}
	if config.MaxArrayLength == 0 {
		config.MaxArrayLength = 100
	}

	return &ValidationManager{
		config: config,
		logger: logger,
	}
}

// ValidateString validates and sanitizes a string
func (vm *ValidationManager) ValidateString(input string, fieldName string, required bool) (string, error) {
	if required && strings.TrimSpace(input) == "" {
		return "", InputValidationError{
			Field:   fieldName,
			Message: "is required",
		}
	}

	if len(input) > vm.config.MaxStringLength {
		return "", InputValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("exceeds maximum length of %d characters", vm.config.MaxStringLength),
			Value:   input[:50] + "...", // Show first 50 chars
		}
	}

	if !utf8.ValidString(input) {
		return "", InputValidationError{
			Field:   fieldName,
			Message: "contains invalid UTF-8 characters",
		}
	}

	// Basic sanitization
	sanitized := strings.TrimSpace(input)

	// Check blocked patterns
	for _, pattern := range vm.config.BlockedPatterns {
		if matched, err := regexp.MatchString(pattern, sanitized); err == nil && matched {
			return "", InputValidationError{
				Field:   fieldName,
				Message: "contains blocked content",
			}
		}
	}

	return sanitized, nil
}

// ValidateEmail validates email format
func (vm *ValidationManager) ValidateEmail(email string) (string, error) {
	if email == "" {
		return "", InputValidationError{
			Field:   "email",
			Message: "is required",
		}
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	sanitized, err := vm.ValidateString(email, "email", true)
	if err != nil {
		return "", err
	}

	if !emailRegex.MatchString(sanitized) {
		return "", InputValidationError{
			Field:   "email",
			Message: "invalid email format",
			Value:   sanitized,
		}
	}

	return sanitized, nil
}

// ValidateEnum validates enum values
func (vm *ValidationManager) ValidateEnum(value string, allowedValues []string, fieldName string) (string, error) {
	sanitized, err := vm.ValidateString(value, fieldName, true)
	if err != nil {
		return "", err
	}

	for _, allowed := range allowedValues {
		if strings.EqualFold(sanitized, allowed) {
			return allowed, nil // Return canonical form
		}
	}

	return "", InputValidationError{
		Field:   fieldName,
		Message: fmt.Sprintf("must be one of: %s", strings.Join(allowedValues, ", ")),
		Value:   sanitized,
	}
}

// ValidateRange validates numeric values within a range
func (vm *ValidationManager) ValidateRange(value interface{}, min, max float64, fieldName string) error {
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
		return InputValidationError{
			Field:   fieldName,
			Message: "must be a valid number",
		}
	}

	if numValue < min || numValue > max {
		return InputValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("must be between %v and %v", min, max),
		}
	}

	return nil
}

// ValidateID validates that an ID is positive
func (vm *ValidationManager) ValidateID(id interface{}, fieldName string) (uint, error) {
	switch val := id.(type) {
	case uint:
		if val == 0 {
			return 0, InputValidationError{
				Field:   fieldName,
				Message: "must be greater than 0",
			}
		}
		return val, nil
	case int:
		if val <= 0 {
			return 0, InputValidationError{
				Field:   fieldName,
				Message: "must be greater than 0",
			}
		}
		return uint(val), nil
	case float64:
		if val <= 0 {
			return 0, InputValidationError{
				Field:   fieldName,
				Message: "must be greater than 0",
			}
		}
		return uint(val), nil
	default:
		return 0, InputValidationError{
			Field:   fieldName,
			Message: "must be a valid positive integer",
		}
	}
}

// ValidateArray validates array length
func (vm *ValidationManager) ValidateArray(array interface{}, fieldName string) error {
	var length int

	switch arr := array.(type) {
	case []interface{}:
		length = len(arr)
	case []string:
		length = len(arr)
	case []int:
		length = len(arr)
	default:
		return InputValidationError{
			Field:   fieldName,
			Message: "must be a valid array",
		}
	}

	if length > vm.config.MaxArrayLength {
		return InputValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("cannot contain more than %d items", vm.config.MaxArrayLength),
		}
	}

	return nil
}

// ValidateAndSanitizeMap validates and sanitizes all string values in a map
func (vm *ValidationManager) ValidateAndSanitizeMap(data map[string]interface{}) ValidationResult {
	result := ValidationResult{
		Valid:     true,
		Errors:    []InputValidationError{},
		Sanitized: make(map[string]string),
	}

	for key, value := range data {
		if str, ok := value.(string); ok {
			sanitized, err := vm.ValidateString(str, key, false)
			if err != nil {
				result.Valid = false
				if valErr, ok := err.(InputValidationError); ok {
					result.Errors = append(result.Errors, valErr)
				} else {
					result.Errors = append(result.Errors, InputValidationError{
						Field:   key,
						Message: err.Error(),
					})
				}
			} else {
				result.Sanitized[key] = sanitized
			}
		}
	}

	return result
}

// Common validation patterns
var (
	CommonValidationConfig = ValidationConfig{
		MaxStringLength: 1000,
		MaxArrayLength:  100,
		BlockedPatterns: []string{
			`<script[^>]*>.*?</script>`,           // Script tags
			`javascript:`,                         // JavaScript URLs
			`on\w+\s*=`,                          // Event handlers
			`\bexec\b|\beval\b|\balert\b`,        // Dangerous JS functions
			`<iframe[^>]*>.*?</iframe>`,          // iframes
			`data:text/html`,                     // Data URLs with HTML
		},
		StrictMode: true,
	}
)

// NewCommonValidationManager creates a validation manager with common settings
func NewCommonValidationManager(logger *zap.Logger) *ValidationManager {
	return NewValidationManager(CommonValidationConfig, logger)
}