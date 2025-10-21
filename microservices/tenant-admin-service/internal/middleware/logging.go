package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggingMiddleware logs all HTTP requests with structured logging and context
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		// Get correlation ID
		correlationID := GetCorrelationID(c)
		
		// Get tenant ID if available
		tenantID := ""
		if tid, exists := c.Get("tenant_id"); exists {
			if id, ok := tid.(string); ok {
				tenantID = id
			}
		}
		
		// Process request
		c.Next()
		
		// Calculate latency
		latency := time.Since(start)
		
		// Get response status
		statusCode := c.Writer.Status()
		
		// Build log fields
		fields := []zap.Field{
			zap.String("correlation_id", correlationID),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", raw),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("latency_human", latency.String()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}
		
		// Add tenant ID if present
		if tenantID != "" {
			fields = append(fields, zap.String("tenant_id", tenantID))
		}
		
		// Add user ID if available
		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(string); ok {
				fields = append(fields, zap.String("user_id", uid))
			}
		}
		
		// Log with appropriate level based on status code
		switch {
		case statusCode >= 500:
			logger.Error("HTTP Request - Server Error", fields...)
		case statusCode >= 400:
			logger.Warn("HTTP Request - Client Error", fields...)
		case statusCode >= 300:
			logger.Info("HTTP Request - Redirect", fields...)
		default:
			logger.Info("HTTP Request - Success", fields...)
		}
	}
}

// ContextLogger creates a logger with context-specific fields
func ContextLogger(logger *zap.Logger, c *gin.Context) *zap.Logger {
	fields := []zap.Field{}
	
	// Add correlation ID
	if correlationID := GetCorrelationID(c); correlationID != "" {
		fields = append(fields, zap.String("correlation_id", correlationID))
	}
	
	// Add tenant ID
	if tenantID, exists := c.Get("tenant_id"); exists {
		if tid, ok := tenantID.(string); ok {
			fields = append(fields, zap.String("tenant_id", tid))
		}
	}
	
	// Add user ID
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			fields = append(fields, zap.String("user_id", uid))
		}
	}
	
	// Add request path
	fields = append(fields, zap.String("path", c.Request.URL.Path))
	
	return logger.With(fields...)
}
