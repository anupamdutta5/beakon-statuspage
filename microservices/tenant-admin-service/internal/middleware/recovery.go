package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/anupamdutta5/tenant-admin-service/internal/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryMiddleware recovers from panics and returns a proper error response
// It logs the panic with stack trace and correlation ID for debugging
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get correlation ID for tracking
				correlationID := GetCorrelationID(c)
				
				// Capture stack trace
				stackTrace := string(debug.Stack())
				
				// Log the panic with full context
				logger.Error("Panic recovered",
					zap.String("correlation_id", correlationID),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()),
					zap.Any("panic", err),
					zap.String("stack_trace", stackTrace),
				)
				
				// Create error response
				appErr := errors.NewInternalError(
					"An unexpected error occurred",
					fmt.Errorf("panic: %v", err),
				)
				
				errorResponse := errors.ToErrorResponse(appErr, correlationID)
				
				// Return 500 Internal Server Error
				c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse)
			}
		}()
		
		c.Next()
	}
}
