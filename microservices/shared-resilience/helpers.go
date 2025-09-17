package resilience

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetCorrelationID extracts the correlation ID from the Gin context
func GetCorrelationID(c *gin.Context) string {
	if correlationID, exists := c.Get("correlation_id"); exists {
		return correlationID.(string)
	}
	return ""
}

// GetUserID extracts the user ID from the Gin context
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(string)
	}
	return ""
}

// GetTenantID extracts the tenant ID from the Gin context
func GetTenantID(c *gin.Context) string {
	if tenantID, exists := c.Get("tenant_id"); exists {
		return tenantID.(string)
	}
	return ""
}

// GetIntQuery extracts an integer query parameter with a default value
func GetIntQuery(c *gin.Context, key string, defaultValue int) int {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// GetStringQuery extracts a string query parameter with a default value
func GetStringQuery(c *gin.Context, key string, defaultValue string) string {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetBoolQuery extracts a boolean query parameter with a default value
func GetBoolQuery(c *gin.Context, key string, defaultValue bool) bool {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// SetUserContext sets user information in the Gin context
func SetUserContext(c *gin.Context, userID, tenantID string) {
	c.Set("user_id", userID)
	if tenantID != "" {
		c.Set("tenant_id", tenantID)
	}
}

// SetCorrelationID sets the correlation ID in the Gin context
func SetCorrelationID(c *gin.Context, correlationID string) {
	c.Set("correlation_id", correlationID)
}

// GetClientIP returns the real client IP address
func GetClientIP(c *gin.Context) string {
	// Check for IP in various headers due to proxies
	clientIP := c.GetHeader("X-Forwarded-For")
	if clientIP != "" {
		return clientIP
	}

	clientIP = c.GetHeader("X-Real-IP")
	if clientIP != "" {
		return clientIP
	}

	return c.ClientIP()
}

// GetUserAgent returns the user agent from the request
func GetUserAgent(c *gin.Context) string {
	return c.GetHeader("User-Agent")
}

// IsAPIRequest checks if the request is an API request based on Accept header
func IsAPIRequest(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	return accept == "application/json" ||
		   c.GetHeader("Content-Type") == "application/json" ||
		   c.Request.URL.Path[:4] == "/api"
}

// RespondWithError sends a standardized error response
func RespondWithError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"error":   message,
		"code":    statusCode,
	})
}

// RespondWithSuccess sends a standardized success response
func RespondWithSuccess(c *gin.Context, statusCode int, data interface{}, message string) {
	response := gin.H{
		"success": true,
		"data":    data,
	}

	if message != "" {
		response["message"] = message
	}

	c.JSON(statusCode, response)
}

// RespondWithPagination sends a response with pagination metadata
func RespondWithPagination(c *gin.Context, data interface{}, page, limit, total int) {
	c.JSON(200, gin.H{
		"success": true,
		"data":    data,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}