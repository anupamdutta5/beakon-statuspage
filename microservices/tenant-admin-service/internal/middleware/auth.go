// Package middleware provides authentication middleware for the Tenant Admin Service.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// Claims represents JWT claims.
type Claims struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	TenantID uint   `json:"tenant_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware provides JWT authentication middleware.
type AuthMiddleware struct {
	jwtSecret string
	logger    *zap.Logger
}

// NewAuthMiddleware creates a new authentication middleware.
func NewAuthMiddleware(jwtSecret string, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

// RequireAuth middleware that requires authentication.
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := m.extractToken(c)
		if tokenString == "" {
			m.logger.Warn("No token provided")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		claims, err := m.validateToken(tokenString)
		if err != nil {
			m.logger.Warn("Invalid token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("username", claims.Username)
		c.Set("tenant_id", claims.TenantID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRole middleware that requires a specific role.
func (m *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check authentication
		tokenString := m.extractToken(c)
		if tokenString == "" {
			m.logger.Warn("No token provided")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		claims, err := m.validateToken(tokenString)
		if err != nil {
			m.logger.Warn("Invalid token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Check role
		if claims.Role != requiredRole {
			m.logger.Warn("Insufficient permissions",
				zap.String("required_role", requiredRole),
				zap.String("user_role", claims.Role))
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("username", claims.Username)
		c.Set("tenant_id", claims.TenantID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireAdmin middleware that requires admin role.
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole("admin")
}

// RequireOwner middleware that requires owner role.
func (m *AuthMiddleware) RequireOwner() gin.HandlerFunc {
	return m.RequireRole("owner")
}

// OptionalAuth middleware that validates token if present but doesn't require it.
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := m.extractToken(c)
		if tokenString != "" {
			claims, err := m.validateToken(tokenString)
			if err == nil {
				// Set user information in context if token is valid
				c.Set("user_id", claims.UserID)
				c.Set("email", claims.Email)
				c.Set("username", claims.Username)
				c.Set("tenant_id", claims.TenantID)
				c.Set("role", claims.Role)
			}
		}
		c.Next()
	}
}

// extractToken extracts the JWT token from the request.
func (m *AuthMiddleware) extractToken(c *gin.Context) string {
	// Try to get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// Check if it's a Bearer token
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// Try to get token from query parameter
	token := c.Query("token")
	if token != "" {
		return token
	}

	// Try to get token from cookie
	cookie, err := c.Cookie("auth_token")
	if err == nil && cookie != "" {
		return cookie
	}

	return ""
}

// validateToken validates the JWT token and returns the claims.
func (m *AuthMiddleware) validateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(m.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenMalformed
	}

	return claims, nil
}

// GetUserID extracts user ID from context.
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := userID.(uint)
	return id, ok
}

// GetTenantID extracts tenant ID from context.
func GetTenantID(c *gin.Context) (uint, bool) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		return 0, false
	}
	id, ok := tenantID.(uint)
	return id, ok
}

// GetUserRole extracts user role from context.
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}
	r, ok := role.(string)
	return r, ok
}





