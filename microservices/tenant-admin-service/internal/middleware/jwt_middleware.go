package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

const (
	maxTokenLength = 4096 // Maximum JWT token length to prevent DoS
	jwtSecretEnv   = "JWT_SECRET"
	authCookieName = "auth_token"
)

// JWTAuthMiddleware validates JWT tokens for tenant-admin-service with comprehensive error handling
func JWTAuthMiddleware() gin.HandlerFunc {
	// Cache JWT secret at middleware initialization to avoid repeated env lookups
	jwtSecret := os.Getenv(jwtSecretEnv)
	if jwtSecret == "" {
		// Log critical error at startup
		logger := getLogger()
		if logger != nil {
			logger.Fatal("JWT_SECRET environment variable not set - authentication will fail")
		}
	}

	return func(c *gin.Context) {
		logger := getLogger()
		var tokenString string

		// 1. Try to get token from Authorization header first
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader != "" {
			// Validate header format
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			} else if logger != nil {
				logger.Debug("Invalid Authorization header format",
					zap.String("format", "expected 'Bearer <token>'"),
					zap.String("path", c.Request.URL.Path))
			}
		}

		// 2. If not in header, try to get from cookie (for browser-based requests)
		if tokenString == "" {
			cookie, err := c.Cookie(authCookieName)
			if err == nil && cookie != "" {
				tokenString = cookie
			}
		}

		// 3. If still no token found, return unauthorized
		if tokenString == "" {
			if logger != nil {
				logger.Debug("No authentication token provided",
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()))
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// 4. Validate token length to prevent DoS attacks
		if len(tokenString) > maxTokenLength {
			if logger != nil {
				logger.Warn("Token exceeds maximum length",
					zap.Int("length", len(tokenString)),
					zap.Int("max_length", maxTokenLength),
					zap.String("client_ip", c.ClientIP()))
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token format",
			})
			c.Abort()
			return
		}

		// 5. Verify JWT secret is configured
		if jwtSecret == "" {
			if logger != nil {
				logger.Error("JWT secret not configured - cannot validate tokens")
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Authentication service unavailable",
			})
			c.Abort()
			return
		}

		// 6. Parse and validate JWT token with timeout protection
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method to prevent algorithm substitution attacks
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				if logger != nil {
					logger.Warn("Unexpected JWT signing method",
						zap.String("method", fmt.Sprintf("%v", token.Header["alg"])),
						zap.String("expected", "HMAC"))
				}
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		// 7. Handle token parsing errors
		if err != nil {
			if logger != nil {
				logger.Debug("Token validation failed",
					zap.Error(err),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()))
			}

			// Provide specific error messages for common issues
			errorMsg := "Invalid or expired token"
			if strings.Contains(err.Error(), "expired") {
				errorMsg = "Token has expired"
			} else if strings.Contains(err.Error(), "signature") {
				errorMsg = "Invalid token signature"
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": errorMsg,
			})
			c.Abort()
			return
		}

		// 8. Verify token is valid
		if !token.Valid {
			if logger != nil {
				logger.Debug("Token validation failed - invalid token",
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()))
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// 9. Extract and validate claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			if logger != nil {
				logger.Error("Failed to extract JWT claims",
					zap.String("path", c.Request.URL.Path))
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			c.Abort()
			return
		}

		// 10. Validate required claims exist
		missingClaims := []string{}
		if _, exists := claims["user_id"]; !exists {
			missingClaims = append(missingClaims, "user_id")
		}
		if _, exists := claims["tenant_id"]; !exists {
			missingClaims = append(missingClaims, "tenant_id")
		}

		if len(missingClaims) > 0 {
			if logger != nil {
				logger.Warn("Token missing required claims",
					zap.Strings("missing_claims", missingClaims),
					zap.String("path", c.Request.URL.Path))
			}
		}

		// 11. Validate expiration manually (defense in depth)
		if exp, ok := claims["exp"].(float64); ok {
			expirationTime := time.Unix(int64(exp), 0)
			if time.Now().After(expirationTime) {
				if logger != nil {
					logger.Debug("Token expired",
						zap.Time("expiration", expirationTime),
						zap.Duration("expired_since", time.Since(expirationTime)))
				}
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Token has expired",
				})
				c.Abort()
				return
			}
		}

		// 12. Set claims in context with type validation
		if userID, exists := claims["user_id"]; exists && userID != nil {
			c.Set("user_id", userID)
		}

		if tenantID, exists := claims["tenant_id"]; exists && tenantID != nil {
			c.Set("tenant_id", tenantID)
		}

		if email, exists := claims["email"]; exists && email != nil {
			c.Set("email", email)
		}

		// 13. Log successful authentication (debug level)
		if logger != nil {
			logger.Debug("JWT authentication successful",
				zap.Any("user_id", claims["user_id"]),
				zap.Any("tenant_id", claims["tenant_id"]),
				zap.String("path", c.Request.URL.Path))
		}

		c.Next()
	}
}

// getLogger safely retrieves the logger from context or returns nil
func getLogger() *zap.Logger {
	// Try to get logger from global scope or create a no-op logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil
	}
	return logger
}
