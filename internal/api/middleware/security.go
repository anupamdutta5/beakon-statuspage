package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/internal/services"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SecurityMiddleware provides comprehensive security for tenant isolation
type SecurityMiddleware struct {
	securityService *services.TenantSecurityService
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(securityService *services.TenantSecurityService) *SecurityMiddleware {
	return &SecurityMiddleware{
		securityService: securityService,
	}
}

// TenantIsolationMiddleware ensures tenant data isolation
func (m *SecurityMiddleware) TenantIsolationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get tenant from context
		tenant, exists := GetTenantFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant not found"})
			c.Abort()
			return
		}

		// Get user from context (if available)
		userID := uint(0)
		if user, exists := GetUserFromContext(c); exists {
			userID = user.ID
		}

		// Validate tenant access
		resource := c.Request.URL.Path
		action := c.Request.Method

		hasAccess, err := m.securityService.ValidateTenantAccess(c.Request.Context(), tenant.ID, userID, resource, action)
		if err != nil {
			logger.Error("Failed to validate tenant access",
				zap.Uint("tenant_id", tenant.ID),
				zap.Uint("user_id", userID),
				zap.String("resource", resource),
				zap.String("action", action),
				zap.Error(err))

			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		if !hasAccess {
			// Log security violation
			violation := &services.SecurityViolation{
				Type:        "unauthorized_access",
				Severity:    "high",
				Description: fmt.Sprintf("Unauthorized access attempt to %s %s", action, resource),
				IPAddress:   c.ClientIP(),
				UserAgent:   c.GetHeader("User-Agent"),
				Details:     fmt.Sprintf("Tenant: %d, User: %d", tenant.ID, userID),
			}

			if err := m.securityService.RecordSecurityViolation(c.Request.Context(), tenant.ID, violation); err != nil {
				logger.Log.Error("Failed to record security violation", zap.Error(err))
			}

			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		// Add tenant context to request
		c.Set("tenant_id", tenant.ID)
		c.Set("tenant_scope", m.securityService.GetTenantDataScope(tenant.ID))

		c.Next()
	}
}

// IPWhitelistMiddleware checks IP whitelist for tenant
func (m *SecurityMiddleware) IPWhitelistMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get tenant from context
		tenant, exists := GetTenantFromContext(c)
		if !exists {
			c.Next()
			return
		}

		// Check IP whitelist
		ipAddress := c.ClientIP()
		allowed, err := m.securityService.CheckIPWhitelist(c.Request.Context(), tenant.ID, ipAddress)
		if err != nil {
			logger.Error("Failed to check IP whitelist",
				zap.Uint("tenant_id", tenant.ID),
				zap.String("ip_address", ipAddress),
				zap.Error(err))

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Security check failed"})
			c.Abort()
			return
		}

		if !allowed {
			// Log security violation
			violation := &services.SecurityViolation{
				Type:        "unauthorized_access",
				Severity:    "medium",
				Description: fmt.Sprintf("Access from non-whitelisted IP: %s", ipAddress),
				IPAddress:   ipAddress,
				UserAgent:   c.GetHeader("User-Agent"),
				Details:     fmt.Sprintf("Tenant: %d", tenant.ID),
			}

			if err := m.securityService.RecordSecurityViolation(c.Request.Context(), tenant.ID, violation); err != nil {
				logger.Log.Error("Failed to record security violation", zap.Error(err))
			}

			c.JSON(http.StatusForbidden, gin.H{"error": "IP address not allowed"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuditLoggingMiddleware logs all requests for audit purposes
func (m *SecurityMiddleware) AuditLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Get tenant from context
		tenant, exists := GetTenantFromContext(c)
		if !exists {
			return
		}

		// Get user from context (if available)
		userID := uint(0)
		if user, exists := GetUserFromContext(c); exists {
			userID = user.ID
		}

		// Create audit log entry
		auditLog := &services.SecurityAuditLog{
			UserID:    userID,
			Action:    c.Request.Method,
			Resource:  c.Request.URL.Path,
			Details:   fmt.Sprintf("Status: %d, Duration: %v", c.Writer.Status(), time.Since(start)),
			IPAddress: c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			Success:   c.Writer.Status() < 400,
		}

		// Log audit event
		if err := m.securityService.LogAuditEvent(c.Request.Context(), tenant.ID, auditLog); err != nil {
			logger.Error("Failed to log audit event",
				zap.Uint("tenant_id", tenant.ID),
				zap.Error(err))
		}
	}
}

// DataAccessLoggingMiddleware logs data access for compliance
func (m *SecurityMiddleware) DataAccessLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get tenant from context
		tenant, exists := GetTenantFromContext(c)
		if !exists {
			c.Next()
			return
		}

		// Get user from context (if available)
		userID := uint(0)
		if user, exists := GetUserFromContext(c); exists {
			userID = user.ID
		}

		// Log data access for specific operations
		if m.isDataAccessOperation(c.Request.Method, c.Request.URL.Path) {
			accessLog := &services.DataAccessLog{
				UserID:    userID,
				Table:     m.extractTableFromPath(c.Request.URL.Path),
				Operation: c.Request.Method,
				RecordID:  c.Param("id"),
				IPAddress: c.ClientIP(),
			}

			if err := m.securityService.LogDataAccess(c.Request.Context(), tenant.ID, accessLog); err != nil {
				logger.Error("Failed to log data access",
					zap.Uint("tenant_id", tenant.ID),
					zap.Error(err))
			}
		}

		c.Next()
	}
}

// RateLimitingMiddleware implements rate limiting per tenant
func (m *SecurityMiddleware) RateLimitingMiddleware() gin.HandlerFunc {
	// Simple in-memory rate limiter
	// In production, use Redis or similar
	rateLimiter := make(map[string][]time.Time)

	return func(c *gin.Context) {
		// Get tenant from context
		tenant, exists := GetTenantFromContext(c)
		if !exists {
			c.Next()
			return
		}

		// Create rate limit key
		key := fmt.Sprintf("tenant_%d_%s", tenant.ID, c.ClientIP())

		// Clean old entries
		now := time.Now()
		cutoff := now.Add(-time.Minute) // 1 minute window

		if timestamps, exists := rateLimiter[key]; exists {
			var validTimestamps []time.Time
			for _, ts := range timestamps {
				if ts.After(cutoff) {
					validTimestamps = append(validTimestamps, ts)
				}
			}
			rateLimiter[key] = validTimestamps
		}

		// Check rate limit (100 requests per minute)
		if timestamps, exists := rateLimiter[key]; exists && len(timestamps) >= 100 {
			// Log security violation
			violation := &services.SecurityViolation{
				Type:        "suspicious_activity",
				Severity:    "medium",
				Description: "Rate limit exceeded",
				IPAddress:   c.ClientIP(),
				UserAgent:   c.GetHeader("User-Agent"),
				Details:     fmt.Sprintf("Tenant: %d, Requests: %d", tenant.ID, len(timestamps)),
			}

			if err := m.securityService.RecordSecurityViolation(c.Request.Context(), tenant.ID, violation); err != nil {
				logger.Log.Error("Failed to record security violation", zap.Error(err))
			}

			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}

		// Add current request
		rateLimiter[key] = append(rateLimiter[key], now)

		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers
func (m *SecurityMiddleware) SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// Temporarily disable CSP for development
		// c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdnjs.cloudflare.com; font-src 'self' https://fonts.gstatic.com https://cdnjs.cloudflare.com; img-src 'self' data: https:; connect-src 'self'; frame-src 'none'; object-src 'none'; base-uri 'self'; form-action 'self'")

		c.Next()
	}
}

// Helper functions

// isDataAccessOperation checks if the request is a data access operation
func (m *SecurityMiddleware) isDataAccessOperation(method, path string) bool {
	// Check for data access operations
	dataAccessPaths := []string{
		"/api/v1/tenant/incidents",
		"/api/v1/tenant/services",
		"/api/v1/tenant/users",
		"/api/v1/tenant/settings",
		"/api/v1/tenant/billing",
	}

	for _, dataPath := range dataAccessPaths {
		if strings.HasPrefix(path, dataPath) {
			return true
		}
	}

	return false
}

// extractTableFromPath extracts the table name from the request path
func (m *SecurityMiddleware) extractTableFromPath(path string) string {
	// Extract table name from path
	parts := strings.Split(path, "/")
	if len(parts) >= 4 {
		return parts[3] // e.g., /api/v1/tenant/incidents -> incidents
	}
	return "unknown"
}

// GetUserFromContext retrieves user from context
func GetUserFromContext(c *gin.Context) (*models.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	userObj, ok := user.(*models.User)
	return userObj, ok
}

// SetTenantInContext sets tenant in context
func SetTenantInContext(c *gin.Context, tenant *models.Tenant) {
	c.Set("tenant", tenant)
}

// SetUserInContext sets user in context
func SetUserInContext(c *gin.Context, user *models.User) {
	c.Set("user", user)
}
