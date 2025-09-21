// Package middleware provides RBAC middleware for the Tenant Admin Service.
package middleware

import (
	"net/http"
	"strings"

	"github.com/anupamdutta5/tenant-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RBACMiddleware provides role-based access control middleware
type RBACMiddleware struct {
	rbacService *services.RBACService
	logger      *zap.Logger
}

// NewRBACMiddleware creates a new RBAC middleware instance
func NewRBACMiddleware(rbacService *services.RBACService, logger *zap.Logger) *RBACMiddleware {
	return &RBACMiddleware{
		rbacService: rbacService,
		logger:      logger,
	}
}

// RequirePermission creates middleware that requires a specific permission
func (m *RBACMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant not found"})
			c.Abort()
			return
		}

		hasPermission, err := m.rbacService.CheckPermission(c.Request.Context(), userID.(uint), tenantID.(uint), permission)
		if err != nil {
			m.logger.Error("Failed to check permission", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
			c.Abort()
			return
		}

		if !hasPermission {
			m.logger.Warn("Permission denied",
				zap.Uint("user_id", userID.(uint)),
				zap.String("permission", permission),
				zap.String("path", c.Request.URL.Path))

			// Log audit event for denied access
			m.rbacService.LogAuditEvent(c.Request.Context(), tenantID.(uint), userID.(uint),
				"access_denied", "permission", nil, "Permission denied: "+permission,
				c.ClientIP(), c.GetHeader("User-Agent"), false, "Insufficient permissions")

			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		// Log successful access for sensitive operations
		if m.isSensitiveOperation(permission) {
			m.rbacService.LogAuditEvent(c.Request.Context(), tenantID.(uint), userID.(uint),
				"access_granted", "permission", nil, "Permission granted: "+permission,
				c.ClientIP(), c.GetHeader("User-Agent"), true, "")
		}

		c.Next()
	}
}

// RequireAnyPermission creates middleware that requires any of the specified permissions
func (m *RBACMiddleware) RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant not found"})
			c.Abort()
			return
		}

		for _, permission := range permissions {
			hasPermission, err := m.rbacService.CheckPermission(c.Request.Context(), userID.(uint), tenantID.(uint), permission)
			if err != nil {
				m.logger.Error("Failed to check permission", zap.Error(err))
				continue
			}

			if hasPermission {
				c.Next()
				return
			}
		}

		m.logger.Warn("Permission denied - no matching permissions",
			zap.Uint("user_id", userID.(uint)),
			zap.Strings("permissions", permissions),
			zap.String("path", c.Request.URL.Path))

		// Log audit event for denied access
		m.rbacService.LogAuditEvent(c.Request.Context(), tenantID.(uint), userID.(uint),
			"access_denied", "permission", nil, "Permission denied - required one of: "+strings.Join(permissions, ", "),
			c.ClientIP(), c.GetHeader("User-Agent"), false, "Insufficient permissions")

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}

// RequireAllPermissions creates middleware that requires all of the specified permissions
func (m *RBACMiddleware) RequireAllPermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant not found"})
			c.Abort()
			return
		}

		for _, permission := range permissions {
			hasPermission, err := m.rbacService.CheckPermission(c.Request.Context(), userID.(uint), tenantID.(uint), permission)
			if err != nil {
				m.logger.Error("Failed to check permission", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
				c.Abort()
				return
			}

			if !hasPermission {
				m.logger.Warn("Permission denied - missing required permission",
					zap.Uint("user_id", userID.(uint)),
					zap.String("missing_permission", permission),
					zap.String("path", c.Request.URL.Path))

				// Log audit event for denied access
				m.rbacService.LogAuditEvent(c.Request.Context(), tenantID.(uint), userID.(uint),
					"access_denied", "permission", nil, "Permission denied - missing: "+permission,
					c.ClientIP(), c.GetHeader("User-Agent"), false, "Insufficient permissions")

				c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RequireRole creates middleware that requires a specific role
func (m *RBACMiddleware) RequireRole(roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant not found"})
			c.Abort()
			return
		}

		userRoles, err := m.rbacService.GetUserRoles(c.Request.Context(), userID.(uint), tenantID.(uint))
		if err != nil {
			m.logger.Error("Failed to get user roles", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user roles"})
			c.Abort()
			return
		}

		hasRole := false
		for _, userRole := range userRoles {
			if userRole.Role.Name == roleName {
				hasRole = true
				break
			}
		}

		if !hasRole {
			m.logger.Warn("Role required",
				zap.Uint("user_id", userID.(uint)),
				zap.String("required_role", roleName),
				zap.String("path", c.Request.URL.Path))

			// Log audit event for denied access
			m.rbacService.LogAuditEvent(c.Request.Context(), tenantID.(uint), userID.(uint),
				"access_denied", "role", nil, "Role required: "+roleName,
				c.ClientIP(), c.GetHeader("User-Agent"), false, "Insufficient role")

			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient role"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SessionValidation validates user session and updates last seen
func (m *RBACMiddleware) SessionValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.GetHeader("Session-ID")
		if sessionID == "" {
			// Try to get from query parameter as fallback
			sessionID = c.Query("session_id")
		}

		if sessionID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session ID required"})
			c.Abort()
			return
		}

		session, err := m.rbacService.ValidateSession(c.Request.Context(), sessionID)
		if err != nil {
			m.logger.Warn("Invalid session", zap.String("session_id", sessionID), zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			c.Abort()
			return
		}

		// Set user and tenant context
		c.Set("user_id", session.UserID)
		c.Set("tenant_id", session.TenantID)
		c.Set("session_id", session.ID)

		c.Next()
	}
}

// AuditLogging logs all requests for audit purposes
func (m *RBACMiddleware) AuditLogging() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			// Log to structured logger
			if param.StatusCode >= 400 {
				m.logger.Warn("HTTP Request",
					zap.String("method", param.Method),
					zap.String("path", param.Path),
					zap.Int("status", param.StatusCode),
					zap.Duration("latency", param.Latency),
					zap.String("client_ip", param.ClientIP),
					zap.String("user_agent", param.Request.UserAgent()),
				)
			} else {
				m.logger.Info("HTTP Request",
					zap.String("method", param.Method),
					zap.String("path", param.Path),
					zap.Int("status", param.StatusCode),
					zap.Duration("latency", param.Latency),
					zap.String("client_ip", param.ClientIP),
				)
			}

			// Return empty string to avoid duplicate logging
			return ""
		},
	})
}

// RateLimitByUser creates user-specific rate limiting
func (m *RBACMiddleware) RateLimitByUser(requestsPerMinute int) gin.HandlerFunc {
	// This is a placeholder for user-specific rate limiting
	// In a real implementation, you would use a rate limiter like go-redis/redis_rate
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		// TODO: Implement actual rate limiting logic based on user ID
		// For now, just log the rate limit check
		m.logger.Debug("Rate limit check",
			zap.Uint("user_id", userID.(uint)),
			zap.Int("limit", requestsPerMinute))

		c.Next()
	}
}

// Helper methods

// isSensitiveOperation determines if an operation is sensitive and should be audited
func (m *RBACMiddleware) isSensitiveOperation(permission string) bool {
	sensitiveOps := []string{
		"users.delete",
		"users.create",
		"settings.update",
		"roles.delete",
		"roles.create",
		"webhooks.delete",
		"integrations.delete",
	}

	for _, op := range sensitiveOps {
		if permission == op {
			return true
		}
	}

	return false
}

// ResourcePermissionMap maps HTTP methods and paths to required permissions
type ResourcePermissionMap struct {
	Method     string
	Path       string
	Permission string
}

// GetResourcePermissions returns a map of API endpoints to required permissions
func GetResourcePermissions() []ResourcePermissionMap {
	return []ResourcePermissionMap{
		// Component management
		{"GET", "/api/v1/components", "components.read"},
		{"POST", "/api/v1/components", "components.create"},
		{"PUT", "/api/v1/components/*", "components.update"},
		{"DELETE", "/api/v1/components/*", "components.delete"},

		// Incident management
		{"GET", "/api/v1/incidents", "incidents.read"},
		{"POST", "/api/v1/incidents", "incidents.create"},
		{"PUT", "/api/v1/incidents/*", "incidents.update"},
		{"DELETE", "/api/v1/incidents/*", "incidents.delete"},

		// Maintenance management
		{"GET", "/api/v1/maintenance/*", "maintenance.read"},
		{"POST", "/api/v1/maintenance/*", "maintenance.create"},
		{"PUT", "/api/v1/maintenance/*", "maintenance.update"},
		{"DELETE", "/api/v1/maintenance/*", "maintenance.delete"},

		// User management
		{"GET", "/api/v1/users", "users.read"},
		{"POST", "/api/v1/users", "users.create"},
		{"PUT", "/api/v1/users/*", "users.update"},
		{"DELETE", "/api/v1/users/*", "users.delete"},

		// Settings
		{"GET", "/api/v1/settings", "settings.read"},
		{"PUT", "/api/v1/settings", "settings.update"},

		// Analytics
		{"GET", "/api/v1/analytics/*", "analytics.read"},

		// Webhooks
		{"GET", "/api/v1/webhooks", "webhooks.read"},
		{"POST", "/api/v1/webhooks", "webhooks.create"},
		{"PUT", "/api/v1/webhooks/*", "webhooks.update"},
		{"DELETE", "/api/v1/webhooks/*", "webhooks.delete"},

		// Integrations
		{"GET", "/api/v1/integrations", "integrations.read"},
		{"POST", "/api/v1/integrations", "integrations.create"},
		{"PUT", "/api/v1/integrations/*", "integrations.update"},
		{"DELETE", "/api/v1/integrations/*", "integrations.delete"},
	}
}

// AutoPermissionMiddleware automatically applies permission checks based on the endpoint
func (m *RBACMiddleware) AutoPermissionMiddleware() gin.HandlerFunc {
	permissionMap := make(map[string]string)
	for _, mapping := range GetResourcePermissions() {
		key := mapping.Method + ":" + mapping.Path
		permissionMap[key] = mapping.Permission
	}

	return func(c *gin.Context) {
		// Skip permission check for public endpoints
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/public/") ||
			c.Request.URL.Path == "/health" ||
			c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		// Look up required permission
		key := c.Request.Method + ":" + c.Request.URL.Path
		permission, exists := permissionMap[key]
		if !exists {
			// Try wildcard matching
			for mapKey, mapPermission := range permissionMap {
				if strings.HasSuffix(mapKey, "/*") {
					prefix := strings.TrimSuffix(mapKey, "/*")
					if strings.HasPrefix(key, prefix) {
						permission = mapPermission
						exists = true
						break
					}
				}
			}
		}

		if !exists {
			// No permission requirement found, allow access
			c.Next()
			return
		}

		// Apply permission check
		m.RequirePermission(permission)(c)
	}
}