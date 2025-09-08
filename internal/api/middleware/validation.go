package middleware

import (
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/enterprise-status/statuspage/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// ValidationMiddleware provides comprehensive input validation
type ValidationMiddleware struct {
	validator *validator.Validate
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware() *ValidationMiddleware {
	validate := validator.New()

	// Register custom validators
	if err := validate.RegisterValidation("alphanumeric", validateAlphanumeric); err != nil {
		logger.Log.Error("Failed to register alphanumeric validator", zap.Error(err))
	}
	if err := validate.RegisterValidation("slug", validateSlug); err != nil {
		logger.Log.Error("Failed to register slug validator", zap.Error(err))
	}
	if err := validate.RegisterValidation("email", validateEmail); err != nil {
		logger.Log.Error("Failed to register email validator", zap.Error(err))
	}
	if err := validate.RegisterValidation("password", validatePassword); err != nil {
		logger.Log.Error("Failed to register password validator", zap.Error(err))
	}
	if err := validate.RegisterValidation("url", validateURL); err != nil {
		logger.Log.Error("Failed to register url validator", zap.Error(err))
	}
	if err := validate.RegisterValidation("phone", validatePhone); err != nil {
		logger.Log.Error("Failed to register phone validator", zap.Error(err))
	}

	return &ValidationMiddleware{
		validator: validate,
	}
}

// ValidateRequest validates request data based on struct tags
func (vm *ValidationMiddleware) ValidateRequest(data interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindJSON(data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request data",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		if err := vm.validator.Struct(data); err != nil {
			validationErrors := make(map[string]string)
			for _, err := range err.(validator.ValidationErrors) {
				validationErrors[err.Field()] = getValidationErrorMessage(err)
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"details": validationErrors,
			})
			c.Abort()
			return
		}

		c.Set("validated_data", data)
		c.Next()
	}
}

// SanitizeInput sanitizes input data to prevent XSS and injection attacks
func (vm *ValidationMiddleware) SanitizeInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		for key, values := range c.Request.URL.Query() {
			for i, value := range values {
				values[i] = vm.sanitizeString(value)
			}
			c.Request.URL.RawQuery = strings.ReplaceAll(c.Request.URL.RawQuery, key+"="+values[0], key+"="+vm.sanitizeString(values[0]))
		}

		// Sanitize form data
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			if err := c.Request.ParseForm(); err == nil {
				for key, values := range c.Request.PostForm {
					for i, value := range values {
						c.Request.PostForm[key][i] = vm.sanitizeString(value)
					}
				}
			}
		}

		c.Next()
	}
}

// ValidateContentType validates the content type of the request
func (vm *ValidationMiddleware) ValidateContentType(allowedTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		contentType := c.GetHeader("Content-Type")

		if contentType == "" && (c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Content-Type header is required",
			})
			c.Abort()
			return
		}

		if len(allowedTypes) > 0 {
			valid := false
			for _, allowedType := range allowedTypes {
				if strings.Contains(contentType, allowedType) {
					valid = true
					break
				}
			}

			if !valid {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"error":    "Unsupported content type",
					"expected": allowedTypes,
					"received": contentType,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// ValidateFileUpload validates file uploads
func (vm *ValidationMiddleware) ValidateFileUpload(maxSize int64, allowedTypes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "File upload required",
			})
			c.Abort()
			return
		}
		defer file.Close()

		// Check file size
		if header.Size > maxSize {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":     "File too large",
				"max_size":  maxSize,
				"file_size": header.Size,
			})
			c.Abort()
			return
		}

		// Check file type
		if len(allowedTypes) > 0 {
			valid := false
			for _, allowedType := range allowedTypes {
				if strings.Contains(header.Header.Get("Content-Type"), allowedType) {
					valid = true
					break
				}
			}

			if !valid {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":         "Invalid file type",
					"allowed_types": allowedTypes,
					"file_type":     header.Header.Get("Content-Type"),
				})
				c.Abort()
				return
			}
		}

		c.Set("uploaded_file", file)
		c.Set("file_header", header)
		c.Next()
	}
}

// Custom validation functions
func validateAlphanumeric(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	for _, char := range value {
		if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
			return false
		}
	}
	return true
}

func validateSlug(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	matched, _ := regexp.MatchString(`^[a-z0-9]+(?:-[a-z0-9]+)*$`, value)
	return matched
}

func validateEmail(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, value)
	return matched
}

func validatePassword(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if len(value) < 8 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range value {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func validateURL(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	matched, _ := regexp.MatchString(`^https?://[^\s/$.?#].[^\s]*$`, value)
	return matched
}

func validatePhone(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	matched, _ := regexp.MatchString(`^\+?[1-9]\d{1,14}$`, value)
	return matched
}

// Helper functions
func (vm *ValidationMiddleware) sanitizeString(input string) string {
	// Remove potentially dangerous characters
	input = strings.ReplaceAll(input, "<script", "")
	input = strings.ReplaceAll(input, "</script>", "")
	input = strings.ReplaceAll(input, "javascript:", "")
	input = strings.ReplaceAll(input, "onload=", "")
	input = strings.ReplaceAll(input, "onerror=", "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	return input
}

func getValidationErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Must be at least " + err.Param() + " characters long"
	case "max":
		return "Must be at most " + err.Param() + " characters long"
	case "alphanumeric":
		return "Must contain only letters and numbers"
	case "slug":
		return "Must be a valid slug (lowercase letters, numbers, and hyphens only)"
	case "password":
		return "Password must be at least 8 characters with uppercase, lowercase, number, and special character"
	case "url":
		return "Must be a valid URL"
	case "phone":
		return "Must be a valid phone number"
	default:
		return "Invalid value"
	}
}
