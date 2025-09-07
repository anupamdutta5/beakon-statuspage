package middleware

import (
	"time"

	"github.com/enterprise-status/statuspage/internal/services"
	"github.com/gin-gonic/gin"
)

// AnalyticsMiddleware tracks API usage for analytics
func AnalyticsMiddleware(analyticsService *services.AnalyticsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process the request
		c.Next()

		// Calculate response time
		responseTime := float64(time.Since(start).Nanoseconds()) / 1000000 // Convert to milliseconds

		// Get tenant ID from context (if available)
		tenantID := uint(0)
		if tenantIDValue, exists := c.Get("tenant_id"); exists {
			if id, ok := tenantIDValue.(uint); ok {
				tenantID = id
			}
		}

		// Get client information
		userAgent := c.GetHeader("User-Agent")
		ipAddress := c.ClientIP()

		// Record API usage (async to avoid slowing down the request)
		go func() {
			_ = analyticsService.RecordAPIUsage(
				tenantID,
				c.Request.URL.Path,
				c.Request.Method,
				responseTime,
				c.Writer.Status(),
				userAgent,
				ipAddress,
			)
		}()
	}
}

// PageViewMiddleware tracks page views for analytics
func PageViewMiddleware(analyticsService *services.AnalyticsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only track GET requests for HTML pages
		if c.Request.Method == "GET" && c.GetHeader("Accept") != "" {
			accept := c.GetHeader("Accept")
			if accept == "text/html" || accept == "*/*" {
				// Get tenant ID from context (if available)
				tenantID := uint(0)
				if tenantIDValue, exists := c.Get("tenant_id"); exists {
					if id, ok := tenantIDValue.(uint); ok {
						tenantID = id
					}
				}

				// Get client information
				userAgent := c.GetHeader("User-Agent")
				ipAddress := c.ClientIP()
				referrer := c.GetHeader("Referer")

				// Generate a simple session ID (in a real app, you'd use proper session management)
				sessionID := c.GetHeader("X-Session-ID")
				if sessionID == "" {
					sessionID = ipAddress + "-" + userAgent // Simple session identifier
				}

				// Record page view (async to avoid slowing down the request)
				go func() {
					_ = analyticsService.RecordPageView(
						tenantID,
						c.Request.URL.Path,
						userAgent,
						ipAddress,
						referrer,
						sessionID,
					)
				}()
			}
		}

		c.Next()
	}
}
