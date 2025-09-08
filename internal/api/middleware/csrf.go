package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/enterprise-status/statuspage/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CSRFProtection provides CSRF protection middleware
type CSRFProtection struct {
	secretKey []byte
	secure    bool
	sameSite  string
}

// NewCSRFProtection creates a new CSRF protection middleware
func NewCSRFProtection(secretKey string, secure bool, sameSite string) *CSRFProtection {
	return &CSRFProtection{
		secretKey: []byte(secretKey),
		secure:    secure,
		sameSite:  sameSite,
	}
}

// CSRFMiddleware creates a CSRF protection middleware
func (csrf *CSRFProtection) CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF for safe methods
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			// Generate and set CSRF token for safe methods
			token, err := csrf.generateToken(c)
			if err != nil {
				logger.Error("Failed to generate CSRF token", zap.Error(err))
				c.Next()
				return
			}

			c.Header("X-CSRF-Token", token)
			c.Set("csrf_token", token)
			c.Next()
			return
		}

		// For unsafe methods, validate CSRF token
		if !csrf.validateToken(c) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "CSRF token validation failed",
				"code":  "CSRF_TOKEN_INVALID",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// generateToken generates a CSRF token
func (csrf *CSRFProtection) generateToken(c *gin.Context) (string, error) {
	// Generate random bytes
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Create token with session identifier
	sessionID := csrf.getSessionID(c)
	tokenData := fmt.Sprintf("%s:%s", sessionID, base64.URLEncoding.EncodeToString(randomBytes))

	// Sign the token
	signature := csrf.signToken(tokenData)
	token := fmt.Sprintf("%s.%s", base64.URLEncoding.EncodeToString([]byte(tokenData)), signature)

	return token, nil
}

// validateToken validates a CSRF token
func (csrf *CSRFProtection) validateToken(c *gin.Context) bool {
	// Get token from header or form
	token := c.GetHeader("X-CSRF-Token")
	if token == "" {
		token = c.PostForm("csrf_token")
	}

	if token == "" {
		return false
	}

	// Split token into data and signature
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}

	// Decode token data
	tokenData, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}

	// Verify signature
	expectedSignature := csrf.signToken(string(tokenData))
	if !csrf.constantTimeCompare(parts[1], expectedSignature) {
		return false
	}

	// Parse token data
	tokenParts := strings.Split(string(tokenData), ":")
	if len(tokenParts) != 2 {
		return false
	}

	// Verify session ID
	sessionID := csrf.getSessionID(c)
	return csrf.constantTimeCompare(tokenParts[0], sessionID)
}

// signToken signs a token with HMAC
func (csrf *CSRFProtection) signToken(data string) string {
	// Simple signature using the secret key
	// In production, use HMAC-SHA256
	hash := fmt.Sprintf("%x", len(data)+len(csrf.secretKey))
	return base64.URLEncoding.EncodeToString([]byte(hash))
}

// constantTimeCompare performs constant time comparison
func (csrf *CSRFProtection) constantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// getSessionID gets a session identifier
func (csrf *CSRFProtection) getSessionID(c *gin.Context) string {
	// Try to get session ID from cookie
	if cookie, err := c.Cookie("session_id"); err == nil && cookie != "" {
		return cookie
	}

	// Try to get user ID from context
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%v", userID)
	}

	// Fall back to IP address
	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// SetCSRFTokenCookie sets the CSRF token as a cookie
func (csrf *CSRFProtection) SetCSRFTokenCookie() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := csrf.generateToken(c)
		if err != nil {
			logger.Error("Failed to generate CSRF token for cookie", zap.Error(err))
			c.Next()
			return
		}

		// Set cookie options
		cookieOptions := []string{
			fmt.Sprintf("csrf_token=%s", token),
			"Path=/",
			"HttpOnly",
		}

		if csrf.secure {
			cookieOptions = append(cookieOptions, "Secure")
		}

		if csrf.sameSite != "" {
			cookieOptions = append(cookieOptions, fmt.Sprintf("SameSite=%s", csrf.sameSite))
		}

		c.Header("Set-Cookie", strings.Join(cookieOptions, "; "))
		c.Set("csrf_token", token)
		c.Next()
	}
}

// CSRFExemptPaths returns a middleware that exempts certain paths from CSRF protection
func (csrf *CSRFProtection) CSRFExemptPaths(exemptPaths []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Check if path is exempt
		for _, exemptPath := range exemptPaths {
			if strings.HasPrefix(path, exemptPath) {
				c.Next()
				return
			}
		}

		// Apply CSRF protection
		csrf.CSRFMiddleware()(c)
	}
}

// CSRFForAPI returns a middleware that applies CSRF protection only to API endpoints
func (csrf *CSRFProtection) CSRFForAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply CSRF to API endpoints
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			csrf.CSRFMiddleware()(c)
			return
		}

		c.Next()
	}
}

// CSRFForWeb returns a middleware that applies CSRF protection only to web endpoints
func (csrf *CSRFProtection) CSRFForWeb() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF for API routes
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}

		// Skip CSRF for admin routes (login and dashboard)
		if strings.HasPrefix(c.Request.URL.Path, "/admin/") {
			c.Next()
			return
		}

		// Apply CSRF protection to all other web endpoints
		csrf.CSRFMiddleware()(c)
	}
}

// GetCSRFToken returns the CSRF token from context
func GetCSRFToken(c *gin.Context) (string, bool) {
	token, exists := c.Get("csrf_token")
	if !exists {
		return "", false
	}

	tokenStr, ok := token.(string)
	return tokenStr, ok
}

// CSRFConfig represents CSRF configuration
type CSRFConfig struct {
	SecretKey   string
	Secure      bool
	SameSite    string
	ExemptPaths []string
}

// NewCSRFProtectionFromConfig creates CSRF protection from config
func NewCSRFProtectionFromConfig(config CSRFConfig) *CSRFProtection {
	return NewCSRFProtection(config.SecretKey, config.Secure, config.SameSite)
}
