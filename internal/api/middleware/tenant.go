package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/internal/services"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const TenantContextKey = "tenant"

// TenantMiddleware extracts tenant information from the request
func TenantMiddleware(saasService *services.SaaSService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip tenant resolution for API routes
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}

		// Skip tenant resolution for all admin routes
		if strings.HasPrefix(c.Request.URL.Path, "/admin") {
			c.Next()
			return
		}

		// Skip tenant resolution for all admin API routes
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/admin") {
			c.Next()
			return
		}

		// Skip tenant resolution for static files
		if strings.HasPrefix(c.Request.URL.Path, "/static") {
			c.Next()
			return
		}

		// Skip tenant resolution for webhooks
		if strings.HasPrefix(c.Request.URL.Path, "/webhooks") {
			c.Next()
			return
		}

		// Skip tenant resolution for health check
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		// Get tenant from subdomain or custom domain
		tenant, err := resolveTenant(c, saasService)
		if err != nil {
			logger.Error("Failed to resolve tenant", zap.Error(err))
			// If no tenant found, show default landing page
			c.HTML(http.StatusOK, "index.html", gin.H{})
			c.Abort()
			return
		}

		// Add tenant to context
		c.Set(TenantContextKey, tenant)
		c.Next()
	}
}

// resolveTenant resolves tenant from request host
func resolveTenant(c *gin.Context, saasService *services.SaaSService) (*models.Tenant, error) {
	host := c.Request.Host

	// Remove port if present
	if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	// Check for custom domain first
	tenant, err := saasService.GetTenantByDomain(host)
	if err == nil {
		return tenant, nil
	}

	// Check for subdomain
	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		subdomain := parts[0]

		// Skip common subdomains
		if subdomain == "www" || subdomain == "admin" || subdomain == "api" {
			return nil, fmt.Errorf("invalid subdomain")
		}

		// Try to get tenant by subdomain
		tenant, err := saasService.GetTenantBySubdomain(subdomain)
		if err == nil {
			return tenant, nil
		}
	}

	return nil, fmt.Errorf("tenant not found")
}

// GetTenantFromContext retrieves tenant from gin context
func GetTenantFromContext(c *gin.Context) (*models.Tenant, bool) {
	tenant, exists := c.Get(TenantContextKey)
	if !exists {
		return nil, false
	}

	tenantObj, ok := tenant.(*models.Tenant)
	return tenantObj, ok
}

// RequireTenant middleware ensures tenant is present in context
func RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenant, exists := GetTenantFromContext(c)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			c.Abort()
			return
		}

		// Check if tenant is active
		if !tenant.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "Tenant is inactive"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TenantAdminMiddleware ensures user has admin access to the tenant
func TenantAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenant, exists := GetTenantFromContext(c)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			c.Abort()
			return
		}

		// Get user from auth context
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		userObj, ok := user.(*models.User)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
			c.Abort()
			return
		}

		// Check if user is admin of this tenant
		// For now, we'll check if user created the tenant
		// In a more complex system, you'd have tenant-specific roles
		if userObj.ID != tenant.CreatedBy {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}
