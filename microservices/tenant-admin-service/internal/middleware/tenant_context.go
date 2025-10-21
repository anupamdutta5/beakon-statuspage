package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TenantContextMiddleware extracts tenant information from subdomain or query parameter
// and sets it in the Gin context for downstream handlers.
//
// Priority order:
// 1. Subdomain (e.g., anupam.localhost or anupam.yourdomain.com)
// 2. Query parameter ?tenant=anupam
// 3. Query parameter ?tenant_id=123
//
// Usage:
//   - Production: https://anupam.yourdomain.com → tenant "anupam"
//   - Development: http://anupam.localhost:8099 → tenant "anupam"
//   - Fallback: http://localhost:8099?tenant=anupam → tenant "anupam"
//   - Fallback: http://192.168.1.10:8099?tenant_id=123 → tenant ID 123
func TenantContextMiddleware(db *gorm.DB, logger *zap.Logger, baseDomain string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tenantSlug string
		var tenantID uint

		// 1. Try to extract subdomain from Host header
		// Check X-Forwarded-Host first (for proxied requests from Next.js)
		host := c.Request.Header.Get("X-Forwarded-Host")
		if host == "" {
			host = c.Request.Host
		}
		subdomain := extractSubdomain(host, baseDomain)

		if subdomain != "" && subdomain != "www" {
			tenantSlug = subdomain
			logger.Debug("Extracted tenant from subdomain",
				zap.String("host", host),
				zap.String("subdomain", subdomain),
				zap.String("tenant", tenantSlug))
		}

		// 2. Fallback to query parameter ?tenant=X
		if tenantSlug == "" {
			if tenant := c.Query("tenant"); tenant != "" {
				tenantSlug = tenant
				logger.Debug("Extracted tenant from query parameter",
					zap.String("tenant", tenantSlug))
			}
		}

		// 3. Fallback to query parameter ?tenant_id=X
		if tenantSlug == "" {
			if tidStr := c.Query("tenant_id"); tidStr != "" {
				// Try to parse as uint
				var tid uint64
				if _, err := fmt.Sscanf(tidStr, "%d", &tid); err == nil {
					tenantID = uint(tid)
					logger.Debug("Extracted tenant ID from query parameter",
						zap.Uint("tenant_id", tenantID))
				}
			}
		}

		// Look up tenant in database if we have a slug
		if tenantSlug != "" {
			var tenant struct {
				ID   string `gorm:"column:id;type:uuid"`
				Slug string `gorm:"column:slug"`
				Name string `gorm:"column:name"`
			}

			err := db.Table("tenants").
				Where("subdomain = ?", tenantSlug).
				First(&tenant).Error

			if err != nil {
				if err == gorm.ErrRecordNotFound {
					logger.Warn("Tenant not found",
						zap.String("tenant_slug", tenantSlug),
						zap.String("host", host))
				} else {
					logger.Error("Failed to lookup tenant",
						zap.String("tenant_slug", tenantSlug),
						zap.Error(err))
				}
				// Mark tenant as not found for RequireTenant middleware
				c.Set("tenant_not_found", true)
			} else {
				// Set tenant context
				c.Set("tenant_id", tenant.ID)
				c.Set("tenant_slug", tenant.Slug)
				c.Set("tenant_name", tenant.Name)

				logger.Info("Tenant context set",
					zap.String("tenant_id", tenant.ID),
					zap.String("tenant_slug", tenant.Slug),
					zap.String("tenant_name", tenant.Name))
			}
		} else if tenantID > 0 {
			// Direct tenant ID provided (as integer)
			c.Set("tenant_id", tenantID)
			logger.Info("Tenant ID set from parameter",
				zap.Uint("tenant_id", tenantID))
		}

		c.Next()
	}
}

// extractSubdomain extracts the subdomain from a host string
// Examples:
//   - "anupam.localhost:8099" with baseDomain "localhost" → "anupam"
//   - "anupam.yourdomain.com" with baseDomain "yourdomain.com" → "anupam"
//   - "www.yourdomain.com" with baseDomain "yourdomain.com" → "www"
//   - "localhost:8099" with baseDomain "localhost" → ""
//   - "192.168.1.10:8099" with baseDomain "localhost" → ""
func extractSubdomain(host string, baseDomain string) string {
	// Remove port if present
	if colonIdx := strings.LastIndex(host, ":"); colonIdx != -1 {
		host = host[:colonIdx]
	}

	// If host is an IP address or just the base domain, no subdomain
	if host == baseDomain || isIPAddress(host) {
		return ""
	}

	// Check if host ends with baseDomain
	if !strings.HasSuffix(host, "."+baseDomain) {
		return ""
	}

	// Extract subdomain
	subdomain := strings.TrimSuffix(host, "."+baseDomain)

	// Handle nested subdomains - only take the first part
	// e.g., "app.anupam.localhost" → "app"
	if dotIdx := strings.Index(subdomain, "."); dotIdx != -1 {
		subdomain = subdomain[:dotIdx]
	}

	return subdomain
}

// isIPAddress checks if a string is an IP address (simple check)
func isIPAddress(s string) bool {
	// Simple check: if it starts with a digit and contains only digits and dots, it's likely an IP
	if len(s) == 0 || (s[0] < '0' || s[0] > '9') {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && c != '.' {
			return false
		}
	}
	return true
}

// GetTenantID retrieves the tenant ID from the Gin context
// Returns tenant ID as a string (can be UUID or integer)
func GetTenantID(c *gin.Context) (string, bool) {
	if id, exists := c.Get("tenant_id"); exists {
		// Handle string IDs (UUID)
		if tenantID, ok := id.(string); ok {
			return tenantID, true
		}
		// Handle uint IDs (legacy)
		if tenantID, ok := id.(uint); ok {
			return fmt.Sprintf("%d", tenantID), true
		}
	}
	return "", false
}

// GetTenantSlug retrieves the tenant slug from the Gin context
func GetTenantSlug(c *gin.Context) (string, bool) {
	if slug, exists := c.Get("tenant_slug"); exists {
		if tenantSlug, ok := slug.(string); ok {
			return tenantSlug, true
		}
	}
	return "", false
}

// RequireTenant is middleware that ensures a tenant context exists
func RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if tenant was explicitly not found (invalid subdomain provided)
		if notFound, exists := c.Get("tenant_not_found"); exists && notFound.(bool) {
			// Return simple HTML error page
			c.Data(404, "text/html; charset=utf-8", []byte(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Tenant Not Found</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: #f5f5f5; }
        .error-box { background: white; padding: 40px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); text-align: center; max-width: 500px; }
        h1 { color: #e74c3c; margin: 0 0 20px 0; }
        p { color: #555; margin: 10px 0; }
        .icon { font-size: 64px; margin-bottom: 20px; }
    </style>
</head>
<body>
    <div class="error-box">
        <div class="icon">⚠️</div>
        <h1>Tenant Not Found</h1>
        <p>The tenant you're trying to access does not exist.</p>
        <p>Please check the subdomain and try again.</p>
    </div>
</body>
</html>`))
			c.Abort()
			return
		}

		// Check if tenant context exists
		if _, exists := GetTenantID(c); !exists {
			// Return simple HTML error page
			c.Data(400, "text/html; charset=utf-8", []byte(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Tenant Required</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: #f5f5f5; }
        .error-box { background: white; padding: 40px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); text-align: center; max-width: 500px; }
        h1 { color: #3498db; margin: 0 0 20px 0; }
        p { color: #555; margin: 10px 0; }
        .icon { font-size: 64px; margin-bottom: 20px; }
        code { background: #f8f8f8; padding: 2px 6px; border-radius: 3px; font-family: monospace; }
    </style>
</head>
<body>
    <div class="error-box">
        <div class="icon">🔒</div>
        <h1>Tenant Required</h1>
        <p>Please access this service via a tenant subdomain.</p>
        <p>Example: <code>yourtenant.localhost:8099</code></p>
    </div>
</body>
</html>`))
			c.Abort()
			return
		}
		c.Next()
	}
}
