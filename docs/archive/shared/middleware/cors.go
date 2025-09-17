package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig contains CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
	ExposeHeaders    []string
}

// DefaultCORSConfig returns secure default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:8080"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"X-Correlation-ID",
			"Accept",
			"Origin",
			"Cache-Control",
			"X-File-Name",
		},
		AllowCredentials: false,
		MaxAge:           3600, // 1 hour
		ExposeHeaders: []string{
			"X-Correlation-ID",
			"X-Total-Count",
			"X-Page-Count",
		},
	}
}

// ProductionCORSConfig returns production-ready CORS configuration
func ProductionCORSConfig(allowedOrigins []string) CORSConfig {
	config := DefaultCORSConfig()
	config.AllowedOrigins = allowedOrigins
	config.AllowCredentials = true
	config.MaxAge = 86400 // 24 hours

	return config
}

// DevelopmentCORSConfig returns development-friendly CORS configuration
func DevelopmentCORSConfig() CORSConfig {
	config := DefaultCORSConfig()
	config.AllowedOrigins = []string{"*"}
	config.AllowCredentials = false

	return config
}

// CORS creates a CORS middleware with the given configuration
func CORS(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		if !isOriginAllowed(origin, config.AllowedOrigins) {
			c.Next()
			return
		}

		// Set allowed origin
		if len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		// Set allowed methods
		if len(config.AllowedMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		}

		// Set allowed headers
		if len(config.AllowedHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
		}

		// Set exposed headers
		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}

		// Set credentials
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Set max age
		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
		}

		// Handle preflight request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// isOriginAllowed checks if the origin is in the allowed origins list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return true // Allow requests without origin (e.g., mobile apps, curl)
	}

	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" {
			return true
		}

		if allowedOrigin == origin {
			return true
		}

		// Support wildcard subdomains (e.g., *.example.com)
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := strings.TrimPrefix(allowedOrigin, "*.")
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true
			}
		}
	}

	return false
}

// SecureCORS creates a security-focused CORS middleware
func SecureCORS(allowedOrigins []string, environment string) gin.HandlerFunc {
	var config CORSConfig

	switch environment {
	case "production":
		if len(allowedOrigins) == 0 {
			// Default production origins - should be overridden
			allowedOrigins = []string{"https://yourdomain.com"}
		}
		config = ProductionCORSConfig(allowedOrigins)
	case "staging":
		config = ProductionCORSConfig(allowedOrigins)
		if len(allowedOrigins) == 0 {
			config.AllowedOrigins = []string{"https://staging.yourdomain.com"}
		}
	default:
		// Development environment
		config = DevelopmentCORSConfig()
		if len(allowedOrigins) > 0 {
			config.AllowedOrigins = allowedOrigins
		}
	}

	return CORS(config)
}

// CORSWithConfig creates a CORS middleware with custom configuration
func CORSWithConfig(customConfig CORSConfig) gin.HandlerFunc {
	// Merge with default config to ensure all fields are set
	config := DefaultCORSConfig()

	if len(customConfig.AllowedOrigins) > 0 {
		config.AllowedOrigins = customConfig.AllowedOrigins
	}

	if len(customConfig.AllowedMethods) > 0 {
		config.AllowedMethods = customConfig.AllowedMethods
	}

	if len(customConfig.AllowedHeaders) > 0 {
		config.AllowedHeaders = customConfig.AllowedHeaders
	}

	if len(customConfig.ExposeHeaders) > 0 {
		config.ExposeHeaders = customConfig.ExposeHeaders
	}

	config.AllowCredentials = customConfig.AllowCredentials

	if customConfig.MaxAge > 0 {
		config.MaxAge = customConfig.MaxAge
	}

	return CORS(config)
}

// ValidateCORSConfig validates CORS configuration for security issues
func ValidateCORSConfig(config CORSConfig, environment string) []string {
	var warnings []string

	// Check for wildcard origins in production
	if environment == "production" {
		for _, origin := range config.AllowedOrigins {
			if origin == "*" {
				warnings = append(warnings, "wildcard origin (*) should not be used in production")
			}
		}

		// Warn about credentials with wildcard
		if config.AllowCredentials {
			for _, origin := range config.AllowedOrigins {
				if origin == "*" {
					warnings = append(warnings, "allowing credentials with wildcard origin is insecure")
				}
			}
		}
	}

	// Check for potentially dangerous headers
	dangerousHeaders := []string{"*", "Authorization", "Cookie"}
	for _, header := range config.AllowedHeaders {
		for _, dangerous := range dangerousHeaders {
			if header == dangerous && environment == "production" {
				warnings = append(warnings, "allowing '"+header+"' header may pose security risks")
			}
		}
	}

	// Check for overly permissive methods
	if environment == "production" {
		dangerousMethods := []string{"TRACE", "CONNECT"}
		for _, method := range config.AllowedMethods {
			for _, dangerous := range dangerousMethods {
				if method == dangerous {
					warnings = append(warnings, "method '"+method+"' is not recommended for production")
				}
			}
		}
	}

	return warnings
}