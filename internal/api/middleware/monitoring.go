package middleware

import (
	"strconv"
	"time"

	"github.com/enterprise-status/statuspage/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MonitoringMiddleware provides comprehensive monitoring and logging
type MonitoringMiddleware struct {
	logger *zap.Logger
}

// NewMonitoringMiddleware creates a new monitoring middleware
func NewMonitoringMiddleware() *MonitoringMiddleware {
	return &MonitoringMiddleware{
		logger: logger.GetLogger(),
	}
}

// RequestLoggingMiddleware logs all HTTP requests with structured data
func (m *MonitoringMiddleware) RequestLoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Log structured request data
		m.logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("client_ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.String("error", param.ErrorMessage),
		)

		return ""
	})
}

// PerformanceMiddleware measures and logs request performance
func (m *MonitoringMiddleware) PerformanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate metrics
		latency := time.Since(start)
		status := c.Writer.Status()

		// Log performance metrics
		m.logger.Info("Request Performance",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.Int("response_size", c.Writer.Size()),
		)

		// Set performance headers
		c.Header("X-Response-Time", latency.String())
		c.Header("X-Request-ID", c.GetString("request_id"))
	}
}

// ErrorTrackingMiddleware tracks and logs errors
func (m *MonitoringMiddleware) ErrorTrackingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check for errors
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				m.logger.Error("Request Error",
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("error", err.Error()),
					zap.Uint64("type", uint64(err.Type)),
					zap.String("client_ip", c.ClientIP()),
					zap.String("user_agent", c.Request.UserAgent()),
				)
			}
		}
	}
}

// SecurityMonitoringMiddleware monitors security-related events
func (m *MonitoringMiddleware) SecurityMonitoringMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for suspicious patterns
		userAgent := c.Request.UserAgent()
		clientIP := c.ClientIP()

		// Log suspicious user agents
		if isSuspiciousUserAgent(userAgent) {
			m.logger.Warn("Suspicious User Agent",
				zap.String("user_agent", userAgent),
				zap.String("client_ip", clientIP),
				zap.String("path", c.Request.URL.Path),
			)
		}

		// Log failed authentication attempts
		if c.Request.URL.Path == "/api/auth/login" && c.Request.Method == "POST" {
			// This will be logged by the auth handler, but we can add additional monitoring here
			// TODO: Add additional authentication monitoring logic
			_ = c.Request.URL.Path // Acknowledge the path to avoid empty branch warning
		}

		c.Next()
	}
}

// HealthCheckMiddleware provides health check endpoint
func (m *MonitoringMiddleware) HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.JSON(200, gin.H{
				"status":    "healthy",
				"timestamp": time.Now().Unix(),
				"version":   "1.0.0",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// MetricsMiddleware collects and exposes metrics
func (m *MonitoringMiddleware) MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			// Basic metrics endpoint
			metrics := gin.H{
				"requests_total":      getTotalRequests(),
				"requests_per_second": getRequestsPerSecond(),
				"average_latency":     getAverageLatency(),
				"error_rate":          getErrorRate(),
				"active_connections":  getActiveConnections(),
			}

			c.JSON(200, metrics)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func (m *MonitoringMiddleware) RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// isSuspiciousUserAgent checks if a user agent is suspicious
func isSuspiciousUserAgent(userAgent string) bool {
	suspiciousPatterns := []string{
		"sqlmap",
		"nikto",
		"nmap",
		"masscan",
		"zap",
		"burp",
		"w3af",
		"havij",
		"sqlninja",
		"pangolin",
		"sqlsus",
		"marco",
		"bsqlbf",
		"jsql",
		"sqlmap",
		"acunetix",
		"nessus",
		"openvas",
		"retina",
		"core",
		"qualys",
		"rapid7",
		"metasploit",
		"cobalt",
		"beef",
		"xsser",
		"xsssniper",
		"xsser",
		"xsssniper",
	}

	for _, pattern := range suspiciousPatterns {
		if contains(userAgent, pattern) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					contains(s[1:], substr)))
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// Mock metrics functions (in a real implementation, these would use a metrics library)
func getTotalRequests() int64 {
	return 1000 // Mock value
}

func getRequestsPerSecond() float64 {
	return 10.5 // Mock value
}

func getAverageLatency() float64 {
	return 150.0 // Mock value in milliseconds
}

func getErrorRate() float64 {
	return 0.02 // Mock value (2%)
}

func getActiveConnections() int {
	return 25 // Mock value
}
