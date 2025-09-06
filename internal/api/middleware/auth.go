package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware creates a Gin middleware for JWT authentication.
// It checks for a token in the "auth_token" cookie for web requests
// and the "Authorization" header for API requests.
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// Try to get the token from the cookie first
		cookie, err := c.Cookie("auth_token")
		if err == nil {
			tokenString = cookie
		} else {
			// Fallback to Authorization header if cookie is not present
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				// Differentiate response for web vs. API
				if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
					c.Redirect(http.StatusFound, "/admin/login")
					c.Abort()
					return
				}
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
				return
			}
			tokenString = parts[1]
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil {
			if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.Redirect(http.StatusFound, "/admin/login")
				c.Abort()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("userID", claims["sub"])
			c.Set("userRole", claims["role"])
			c.Next()
		} else {
			if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.Redirect(http.StatusFound, "/admin/login")
				c.Abort()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		}
	}
}
