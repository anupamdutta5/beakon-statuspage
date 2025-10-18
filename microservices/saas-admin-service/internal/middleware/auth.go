package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/anupamdutta5/saas-admin-service/internal/auth"
	"github.com/anupamdutta5/saas-admin-service/internal/services"
)

// JWTAuth is middleware that validates JWT access tokens
func JWTAuth(jwtManager *auth.JWTManager, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get token from Authorization header first
		authHeader := c.GetHeader("Authorization")
		var tokenString string
		var err error

		if authHeader != "" {
			tokenString, err = auth.ExtractTokenFromHeader(authHeader)
			if err != nil {
				logger.Warn("Invalid authorization header", zap.Error(err))
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
				c.Abort()
				return
			}
		} else {
			// Fallback to cookie for web UI
			tokenString, err = c.Cookie("access_token")
			if err != nil {
				logger.Warn("No access token found in header or cookie")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
				c.Abort()
				return
			}
		}

		// Validate the token
		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			logger.Warn("Invalid access token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// SessionAuth is middleware that validates session cookies (legacy support)
func SessionAuth(sessionService *services.SessionService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			logger.Warn("No session found in cookie")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No session found"})
			c.Abort()
			return
		}

		session, err := sessionService.ValidateSession(c.Request.Context(), sessionID)
		if err != nil {
			logger.Warn("Invalid session", zap.Error(err), zap.String("session_id", sessionID))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", session.UserID)
		c.Set("session_id", session.ID)

		c.Next()
	}
}

// OptionalAuth is middleware that allows both authenticated and unauthenticated requests
// If a valid token is present, it sets user context; otherwise, it continues without authentication
func OptionalAuth(jwtManager *auth.JWTManager, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get token from Authorization header first
		authHeader := c.GetHeader("Authorization")
		var tokenString string
		var err error

		if authHeader != "" {
			tokenString, err = auth.ExtractTokenFromHeader(authHeader)
			if err != nil {
				// Invalid header format, but continue without auth
				c.Next()
				return
			}
		} else {
			// Fallback to cookie
			tokenString, err = c.Cookie("access_token")
			if err != nil {
				// No token, continue without auth
				c.Next()
				return
			}
		}

		// Validate the token
		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			// Invalid token, but continue without auth
			c.Next()
			return
		}

		// Set user context if valid token found
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("authenticated", true)

		c.Next()
	}
}

// GetUserID extracts the user ID from the Gin context
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := userID.(uint)
	return id, ok
}

// GetUsername extracts the username from the Gin context
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	name, ok := username.(string)
	return name, ok
}

// GetEmail extracts the email from the Gin context
func GetEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("email")
	if !exists {
		return "", false
	}
	emailStr, ok := email.(string)
	return emailStr, ok
}

// IsAuthenticated checks if the current request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	authenticated, exists := c.Get("authenticated")
	if !exists {
		return false
	}
	isAuth, ok := authenticated.(bool)
	return ok && isAuth
}

// RequireRole is middleware that checks if the user has a specific role
// This is a placeholder for future RBAC implementation
func RequireRole(role string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement role checking when RBAC is added
		// For now, just check if user is authenticated
		userID, exists := GetUserID(c)
		if !exists || userID == 0 {
			logger.Warn("Unauthorized access attempt", zap.String("required_role", role))
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORS middleware for handling cross-origin requests
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Allow specific origins or use wildcard for development
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"http://localhost:8098",
		}

		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimitByIP is a simple IP-based rate limiting middleware
// For production, consider using Redis-based rate limiting
func RateLimitByIP(logger *zap.Logger) gin.HandlerFunc {
	// TODO: Implement proper rate limiting using Redis
	// For now, this is a placeholder
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		logger.Debug("Request from IP", zap.String("ip", clientIP))
		c.Next()
	}
}

// SecurityHeaders adds security-related HTTP headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Remove server header for security
		c.Writer.Header().Del("Server")

		c.Next()
	}
}

// sanitizeInput removes potentially dangerous characters from user input
func sanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	return input
}

// SanitizeInputs is middleware that sanitizes common input fields
func SanitizeInputs() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		for key, values := range c.Request.URL.Query() {
			for i, value := range values {
				c.Request.URL.Query()[key][i] = sanitizeInput(value)
			}
		}

		c.Next()
	}
}
