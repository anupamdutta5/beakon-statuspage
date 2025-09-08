package middleware

import (
	"net/http"
	"strings"

	"github.com/enterprise-status/statuspage/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SecurityHeaders provides comprehensive security headers middleware
type SecurityHeaders struct {
	config *SecurityConfig
}

// SecurityConfig represents security headers configuration
type SecurityConfig struct {
	// Content Security Policy
	CSPDirectives map[string][]string

	// HSTS (HTTP Strict Transport Security)
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
	HSTSPreload           bool

	// X-Frame-Options
	XFrameOptions string

	// X-Content-Type-Options
	XContentTypeOptions bool

	// X-XSS-Protection
	XXSSProtection bool

	// Referrer Policy
	ReferrerPolicy string

	// Permissions Policy
	PermissionsPolicy map[string][]string

	// Feature Policy (deprecated but still used by some browsers)
	FeaturePolicy map[string][]string
}

// NewSecurityHeaders creates a new security headers middleware
func NewSecurityHeaders(config *SecurityConfig) *SecurityHeaders {
	if config == nil {
		config = getDefaultSecurityConfig()
	}

	return &SecurityHeaders{
		config: config,
	}
}

// SecurityHeadersMiddleware creates a security headers middleware
func (sh *SecurityHeaders) SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set Content Security Policy
		if len(sh.config.CSPDirectives) > 0 {
			csp := sh.buildCSP()
			c.Header("Content-Security-Policy", csp)
			c.Header("Content-Security-Policy-Report-Only", csp) // For development
		}

		// Set HSTS
		if sh.config.HSTSMaxAge > 0 {
			hsts := sh.buildHSTS()
			c.Header("Strict-Transport-Security", hsts)
		}

		// Set X-Frame-Options
		if sh.config.XFrameOptions != "" {
			c.Header("X-Frame-Options", sh.config.XFrameOptions)
		}

		// Set X-Content-Type-Options
		if sh.config.XContentTypeOptions {
			c.Header("X-Content-Type-Options", "nosniff")
		}

		// Set X-XSS-Protection
		if sh.config.XXSSProtection {
			c.Header("X-XSS-Protection", "1; mode=block")
		}

		// Set Referrer Policy
		if sh.config.ReferrerPolicy != "" {
			c.Header("Referrer-Policy", sh.config.ReferrerPolicy)
		}

		// Set Permissions Policy
		if len(sh.config.PermissionsPolicy) > 0 {
			permissionsPolicy := sh.buildPermissionsPolicy()
			c.Header("Permissions-Policy", permissionsPolicy)
		}

		// Set Feature Policy (for older browsers)
		if len(sh.config.FeaturePolicy) > 0 {
			featurePolicy := sh.buildFeaturePolicy()
			c.Header("Feature-Policy", featurePolicy)
		}

		// Additional security headers
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		c.Header("Cross-Origin-Embedder-Policy", "require-corp")
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")

		// Remove server information
		c.Header("Server", "")
		c.Header("X-Powered-By", "")

		c.Next()
	}
}

// buildCSP builds the Content Security Policy header
func (sh *SecurityHeaders) buildCSP() string {
	var directives []string

	for directive, sources := range sh.config.CSPDirectives {
		if len(sources) > 0 {
			directives = append(directives, directive+" "+strings.Join(sources, " "))
		}
	}

	return strings.Join(directives, "; ")
}

// buildHSTS builds the HSTS header
func (sh *SecurityHeaders) buildHSTS() string {
	hsts := "max-age=" + string(rune(sh.config.HSTSMaxAge))

	if sh.config.HSTSIncludeSubdomains {
		hsts += "; includeSubDomains"
	}

	if sh.config.HSTSPreload {
		hsts += "; preload"
	}

	return hsts
}

// buildPermissionsPolicy builds the Permissions Policy header
func (sh *SecurityHeaders) buildPermissionsPolicy() string {
	var policies []string

	for feature, allowlist := range sh.config.PermissionsPolicy {
		if len(allowlist) > 0 {
			policies = append(policies, feature+"=("+strings.Join(allowlist, " ")+")")
		} else {
			policies = append(policies, feature+"=()")
		}
	}

	return strings.Join(policies, ", ")
}

// buildFeaturePolicy builds the Feature Policy header
func (sh *SecurityHeaders) buildFeaturePolicy() string {
	var policies []string

	for feature, allowlist := range sh.config.FeaturePolicy {
		if len(allowlist) > 0 {
			policies = append(policies, feature+" "+strings.Join(allowlist, " "))
		} else {
			policies = append(policies, feature+" 'none'")
		}
	}

	return strings.Join(policies, "; ")
}

// getDefaultSecurityConfig returns a default security configuration
func getDefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		CSPDirectives: map[string][]string{
			"default-src": {"'self'"},
			"script-src":  {"'self'", "'unsafe-inline'", "https://cdnjs.cloudflare.com", "https://cdn.jsdelivr.net"},
			"style-src":   {"'self'", "'unsafe-inline'", "https://fonts.googleapis.com", "https://cdnjs.cloudflare.com"},
			"font-src":    {"'self'", "https://fonts.gstatic.com", "https://cdnjs.cloudflare.com"},
			"img-src":     {"'self'", "data:", "https:"},
			"connect-src": {"'self'"},
			"frame-src":   {"'none'"},
			"object-src":  {"'none'"},
			"base-uri":    {"'self'"},
			"form-action": {"'self'"},
		},
		HSTSMaxAge:            31536000, // 1 year
		HSTSIncludeSubdomains: true,
		HSTSPreload:           true,
		XFrameOptions:         "DENY",
		XContentTypeOptions:   true,
		XXSSProtection:        true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		PermissionsPolicy: map[string][]string{
			"camera":        {},
			"microphone":    {},
			"geolocation":   {},
			"payment":       {},
			"usb":           {},
			"magnetometer":  {},
			"gyroscope":     {},
			"accelerometer": {},
		},
		FeaturePolicy: map[string][]string{
			"camera":        {"'none'"},
			"microphone":    {"'none'"},
			"geolocation":   {"'none'"},
			"payment":       {"'none'"},
			"usb":           {"'none'"},
			"magnetometer":  {"'none'"},
			"gyroscope":     {"'none'"},
			"accelerometer": {"'none'"},
		},
	}
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowedOrigins         []string
	AllowedMethods         []string
	AllowedHeaders         []string
	ExposedHeaders         []string
	AllowCredentials       bool
	MaxAge                 int
	AllowWildcard          bool
	AllowBrowserExtensions bool
	AllowWebSockets        bool
	AllowFiles             bool
}

// NewCORSConfig creates a new CORS configuration
func NewCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:8080"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-CSRF-Token",
			"X-API-Key",
		},
		ExposedHeaders:         []string{"X-Total-Count", "X-RateLimit-Limit", "X-RateLimit-Remaining"},
		AllowCredentials:       true,
		MaxAge:                 86400, // 24 hours
		AllowWildcard:          false,
		AllowBrowserExtensions: false,
		AllowWebSockets:        false,
		AllowFiles:             false,
	}
}

// CORSMiddleware creates a CORS middleware
func CORSMiddleware(config *CORSConfig) gin.HandlerFunc {
	if config == nil {
		config = NewCORSConfig()
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		if isOriginAllowed(origin, config) {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if config.AllowWildcard {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		// Set other CORS headers
		c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
		c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
		c.Header("Access-Control-Max-Age", string(rune(config.MaxAge)))

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// isOriginAllowed checks if an origin is allowed
func isOriginAllowed(origin string, config *CORSConfig) bool {
	if origin == "" {
		return false
	}

	for _, allowedOrigin := range config.AllowedOrigins {
		if allowedOrigin == origin {
			return true
		}
	}

	return false
}

// LogSecurityViolation logs security violations
func LogSecurityViolation(c *gin.Context, violationType string, details string) {
	log := logger.GetLogger()

	log.Warn("Security violation detected",
		zap.String("type", violationType),
		zap.String("ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("details", details),
	)
}
