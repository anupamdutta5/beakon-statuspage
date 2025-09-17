package resilience

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CORSMiddleware creates a CORS middleware with the given configuration
func CORSMiddleware(config CORSConfig) gin.HandlerFunc {
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
		return true // Allow requests without origin
	}

	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" {
			return true
		}

		if allowedOrigin == origin {
			return true
		}

		// Support wildcard subdomains
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := strings.TrimPrefix(allowedOrigin, "*.")
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true
			}
		}
	}

	return false
}

// RequestIDMiddleware adds a correlation ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if correlation ID already exists in headers
		correlationID := c.Request.Header.Get("X-Correlation-ID")

		if correlationID == "" {
			// Generate a new correlation ID
			var err error
			correlationID, err = GenerateCorrelationID()
			if err != nil {
				// Fallback to timestamp-based ID if generation fails
				correlationID = fmt.Sprintf("fallback-%d", time.Now().UnixNano())
			}
		}

		// Set correlation ID in context and response header
		c.Set("correlation_id", correlationID)
		c.Header("X-Correlation-ID", correlationID)

		// Add to request context for logger
		ctx := context.WithValue(c.Request.Context(), "correlation_id", correlationID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// LoggingMiddleware logs HTTP requests and responses
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		correlationID, _ := params.Keys["correlation_id"].(string)

		logger.Info("HTTP Request",
			zap.String("method", params.Method),
			zap.String("path", params.Path),
			zap.String("protocol", params.Request.Proto),
			zap.Int("status", params.StatusCode),
			zap.Duration("latency", params.Latency),
			zap.String("client_ip", params.ClientIP),
			zap.String("user_agent", params.Request.UserAgent()),
			zap.String("correlation_id", correlationID),
			zap.String("error", params.ErrorMessage),
		)
		return ""
	})
}

// RecoveryMiddleware handles panics and converts them to 500 errors
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered interface{}) {
		correlationID, _ := c.Get("correlation_id")

		logger.Error("Panic recovered",
			zap.Any("panic", recovered),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Any("correlation_id", correlationID),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":          "Internal server error",
			"correlation_id": correlationID,
			"timestamp":      time.Now().UTC(),
		})
	})
}


// TimeoutMiddleware adds a timeout to requests
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		// Channel to signal completion
		finished := make(chan struct{})

		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			// Request completed normally
		case <-ctx.Done():
			// Request timed out
			correlationID, _ := c.Get("correlation_id")
			c.JSON(http.StatusGatewayTimeout, gin.H{
				"error":          "Request timeout",
				"correlation_id": correlationID,
				"timestamp":      time.Now().UTC(),
			})
			c.Abort()
		}
	}
}

// AuthMiddleware validates JWT tokens
func AuthMiddleware(jwtConfig JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// TODO: Implement JWT validation logic here
		// For now, just check if token is not empty
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// Set user info in context (would be extracted from validated JWT)
		c.Set("user_id", "user-from-jwt")
		c.Set("tenant_id", "tenant-from-jwt")

		c.Next()
	}
}

// TenantMiddleware extracts tenant information from various sources
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tenantID string

		// Try to get tenant ID from header
		tenantID = c.Request.Header.Get("X-Tenant-ID")

		// Try to get from subdomain
		if tenantID == "" {
			host := c.Request.Host
			if strings.Contains(host, ".") {
				parts := strings.Split(host, ".")
				if len(parts) > 2 {
					tenantID = parts[0] // Extract subdomain as tenant
				}
			}
		}

		// Try to get from JWT claims (if already set by auth middleware)
		if tenantID == "" {
			if jwtTenantID, exists := c.Get("tenant_id"); exists {
				tenantID = jwtTenantID.(string)
			}
		}

		// Set tenant ID in context
		if tenantID != "" {
			c.Set("tenant_id", tenantID)
			c.Header("X-Tenant-ID", tenantID)
		}

		c.Next()
	}
}

// MetricsMiddleware collects basic request metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		// TODO: Send metrics to monitoring system
		// For now, just add to context for logging
		c.Set("request_duration", duration)
		c.Set("response_size", c.Writer.Size())
	}
}

// HealthCheckMiddleware provides a simple health check endpoint
func HealthCheckMiddleware(healthPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == healthPath {
			c.JSON(http.StatusOK, gin.H{
				"status":    "healthy",
				"timestamp": time.Now().UTC(),
				"service":   "beakon-service",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// DefaultMiddlewareStack returns a stack of commonly used middleware
func DefaultMiddlewareStack(config *Config, logger *zap.Logger) []gin.HandlerFunc {
	var middleware []gin.HandlerFunc

	// Always include request ID and logging
	middleware = append(middleware, RequestIDMiddleware())
	middleware = append(middleware, LoggingMiddleware(logger))
	middleware = append(middleware, RecoveryMiddleware(logger))

	// Add security headers if enabled
	if config.Security.SecurityHeadersEnabled {
		middleware = append(middleware, SecurityHeadersMiddleware())
	}

	// Add CORS if configured
	if len(config.CORS.AllowedOrigins) > 0 {
		middleware = append(middleware, CORSMiddleware(config.CORS))
	}

	// Add rate limiting if enabled
	if config.RateLimit.Enabled {
		middleware = append(middleware, RateLimitMiddleware(config.RateLimit.RequestsPerMinute, time.Minute))
	}

	// Add input sanitization if enabled
	if config.Security.SanitizationEnabled {
		middleware = append(middleware, InputSanitizationMiddleware())
	}

	// Add metrics collection if enabled
	if config.Monitoring.MetricsEnabled {
		middleware = append(middleware, MetricsMiddleware())
	}

	// Add health check
	middleware = append(middleware, HealthCheckMiddleware(config.Monitoring.HealthPath))

	return middleware
}

// InputSanitizationMiddleware provides basic input sanitization
func InputSanitizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Basic sanitization - in practice, you'd want more sophisticated handling
		c.Next()
	}
}

// SecureMiddlewareStack returns a security-focused middleware stack
func SecureMiddlewareStack(config *Config, logger *zap.Logger) []gin.HandlerFunc {
	middleware := DefaultMiddlewareStack(config, logger)

	// Add additional security middleware
	middleware = append(middleware, TenantMiddleware())
	middleware = append(middleware, TimeoutMiddleware(30*time.Second))

	return middleware
}