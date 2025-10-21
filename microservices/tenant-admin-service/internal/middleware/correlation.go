package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// CorrelationIDHeader is the header name for correlation IDs
	CorrelationIDHeader = "X-Correlation-ID"
	
	// CorrelationIDKey is the context key for storing correlation IDs
	CorrelationIDKey = "correlation_id"
)

// CorrelationIDMiddleware adds a correlation ID to each request for distributed tracing
// If the client sends an X-Correlation-ID header, it will be used; otherwise a new UUID is generated
func CorrelationIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.GetHeader(CorrelationIDHeader)
		
		// Generate a new correlation ID if not provided
		if correlationID == "" {
			correlationID = uuid.New().String()
		}
		
		// Store in context for use by handlers
		c.Set(CorrelationIDKey, correlationID)
		
		// Add to response header so client can track the request
		c.Header(CorrelationIDHeader, correlationID)
		
		c.Next()
	}
}

// GetCorrelationID retrieves the correlation ID from the Gin context
func GetCorrelationID(c *gin.Context) string {
	if correlationID, exists := c.Get(CorrelationIDKey); exists {
		if id, ok := correlationID.(string); ok {
			return id
		}
	}
	return ""
}
