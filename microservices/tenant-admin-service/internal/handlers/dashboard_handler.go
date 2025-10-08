package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/anupamdutta5/tenant-admin-service/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetAdminDashboard renders the admin dashboard with tenant-specific information.
func (h *TenantAdminHandler) GetAdminDashboard(c *gin.Context) {
	// Use middleware helpers to get tenant context
	tenantIDStr, hasID := middleware.GetTenantID(c)
	tenantSlugStr, hasSlug := middleware.GetTenantSlug(c)

	// Require valid tenant context - reject if tenant doesn't exist in database
	if !hasID || !hasSlug {
		h.logger.Warn("Tenant not found - dashboard access denied",
			zap.String("host", c.Request.Host),
			zap.String("path", c.Request.URL.Path))

		c.HTML(http.StatusNotFound, "login.html", gin.H{
			"title": "Tenant Not Found",
			"error": "Tenant not found",
		})
		return
	}

	// Get tenant name from context
	tenantName, _ := c.Get("tenant_name")
	tenantNameStr := ""
	if name, ok := tenantName.(string); ok {
		tenantNameStr = name
	}

	// Parse tenant ID from string to UUID
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.logger.Error("Invalid tenant ID format",
			zap.String("tenant_id", tenantIDStr),
			zap.Error(err))
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"title": "Error",
			"error": "Invalid tenant configuration",
		})
		return
	}

	h.logger.Info("Rendering tenant admin dashboard",
		zap.String("tenant_id", tenantIDStr),
		zap.String("tenant_slug", tenantSlugStr),
		zap.String("tenant_name", tenantNameStr))

	// Generate tenant-specific data
	dashboardData := h.generateTenantDashboardData(tenantID, tenantNameStr)

	c.HTML(http.StatusOK, "admin_dashboard.html", dashboardData)
}

// generateTenantDashboardData creates tenant-specific dashboard data.
//
// ARCHITECTURE NOTE: This function only queries tenant-admin-service's own database tables.
// Data from other microservices (components, incidents, subscribers) should be fetched via
// HTTP API calls to maintain proper service boundaries and avoid cross-database queries.
//
// Current implementation: Returns placeholder zeros for external service data to prevent
// database errors. These should be replaced with API calls to respective services:
// - components -> GET /api/v1/components/count (component-service)
// - incidents -> GET /api/v1/incidents/count (incident-service)
// - subscribers -> GET /api/v1/subscribers/count (notification-service)
func (h *TenantAdminHandler) generateTenantDashboardData(tenantID uuid.UUID, tenantName string) gin.H {
	// Base dashboard data
	dashboardData := gin.H{
		"title":           fmt.Sprintf("%s - Admin Dashboard", tenantName),
		"tenant_id":       tenantID.String(),
		"tenant_name":     tenantName,
		"dashboard_title": fmt.Sprintf("%s Dashboard", tenantName),
		"last_updated":    time.Now().Format("2006-01-02 15:04:05"),
	}

	// Count status pages for this tenant (owned by tenant-admin-service)
	var statusPagesCount int64
	err := h.service.GetDB().Table("status_pages").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Count(&statusPagesCount).Error
	if err != nil {
		h.logger.Warn("Failed to count status pages", zap.Error(err))
		statusPagesCount = 0
	}
	dashboardData["status_pages_count"] = statusPagesCount
	dashboardData["status_pages"] = statusPagesCount

	// MICROSERVICE BOUNDARY: Components are owned by component-service
	// TODO: Replace with HTTP call to component-service: GET /api/v1/components/count?tenant_id={tenantID}
	// For now, return 0 to prevent database errors (table doesn't exist in this service's DB)
	componentsCount := int64(0)
	dashboardData["components_count"] = componentsCount
	dashboardData["components"] = componentsCount

	// MICROSERVICE BOUNDARY: Incidents are owned by incident-service
	// TODO: Replace with HTTP call to incident-service: GET /api/v1/incidents/count?tenant_id={tenantID}&status=active
	// For now, return 0 to prevent database errors (table doesn't exist in this service's DB)
	activeIncidentsCount := int64(0)
	dashboardData["active_incidents"] = activeIncidentsCount

	// MICROSERVICE BOUNDARY: Subscribers are owned by notification-service
	// TODO: Replace with HTTP call to notification-service: GET /api/v1/subscribers/count?tenant_id={tenantID}
	// For now, return 0 to prevent database errors (table doesn't exist in this service's DB)
	subscribersCount := int64(0)
	dashboardData["subscribers_count"] = subscribersCount
	dashboardData["subscribers"] = subscribersCount

	// Determine overall status
	// Since we don't have incident data, default to operational
	// TODO: Fetch actual incident status from incident-service API
	dashboardData["overall_status"] = "All Systems Operational"
	dashboardData["overall_status_class"] = "operational"
	dashboardData["status_icon"] = "fas fa-check-circle"
	dashboardData["status_color"] = "text-success"

	h.logger.Info("Generated dashboard data",
		zap.String("tenant_id", tenantID.String()),
		zap.Int64("status_pages", statusPagesCount),
		zap.Int64("components", componentsCount),
		zap.Int64("incidents", activeIncidentsCount),
		zap.Int64("subscribers", subscribersCount),
		zap.String("note", "External service data pending API integration"))

	return dashboardData
}
