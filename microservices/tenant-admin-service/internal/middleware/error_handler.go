package middleware

import (
	"github.com/anupamdutta5/tenant-admin-service/internal/errors"
	"github.com/gin-gonic/gin"
)

// HandleError is a utility function to handle errors consistently across handlers
// It converts AppError to proper HTTP responses with correlation ID
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	
	// Convert to AppError
	appErr := errors.AsAppError(err)
	
	// Get correlation ID from context
	correlationID := GetCorrelationID(c)
	
	// Create error response
	errorResponse := errors.ToErrorResponse(appErr, correlationID)
	
	// Return JSON response with appropriate status code
	c.AbortWithStatusJSON(appErr.StatusCode, errorResponse)
}

// RespondSuccess is a utility function for successful responses
func RespondSuccess(c *gin.Context, statusCode int, data interface{}) {
	correlationID := GetCorrelationID(c)
	
	response := gin.H{
		"status":         "success",
		"data":           data,
		"correlation_id": correlationID,
	}
	
	c.JSON(statusCode, response)
}

// RespondCreated is a utility function for resource creation responses (201)
func RespondCreated(c *gin.Context, data interface{}) {
	RespondSuccess(c, 201, data)
}

// RespondOK is a utility function for successful responses (200)
func RespondOK(c *gin.Context, data interface{}) {
	RespondSuccess(c, 200, data)
}

// RespondNoContent is a utility function for no content responses (204)
func RespondNoContent(c *gin.Context) {
	c.Status(204)
}
