// Package handlers provides HTTP handlers for component management.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/anupamdutta5/saas-admin-service/internal/models"
	"github.com/anupamdutta5/saas-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ComponentHandler handles component-related HTTP requests.
type ComponentHandler struct {
	service *services.SaaSAdminService
	logger  *zap.Logger
}

// ComponentStatusUpdate represents a status update for a component
type ComponentStatusUpdate struct {
	Status      string `json:"status" binding:"required"`
	Message     string `json:"message"`
	UpdatedBy   string `json:"updated_by"`
	Scheduled   bool   `json:"scheduled"`
	AffectedBy  string `json:"affected_by"` // incident or maintenance ID
}

// NewComponentHandler creates a new component handler.
func NewComponentHandler(service *services.SaaSAdminService, logger *zap.Logger) *ComponentHandler {
	return &ComponentHandler{
		service: service,
		logger:  logger,
	}
}

// GetComponents handles getting all components
func (h *ComponentHandler) GetComponents(c *gin.Context) {
	h.logger.Info("Getting components")

	// Parse query parameters
	groupName := c.Query("group")
	visible := c.Query("visible")

	// Mock component data representing different service components
	components := []models.SaaSComponent{
		{
			ID:          uuid.New(),
			TenantID:    "default-tenant",
			Name:        "Web Application",
			Description: "Main web application frontend",
			Status:      "operational",
			GroupName:   "Frontend",
			Order:       1,
			Visible:     true,
			ShowUptime:  true,
			Uptime:      99.95,
			Link:        "https://app.example.com",
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		},
		{
			ID:          uuid.New(),
			TenantID:    "default-tenant",
			Name:        "API Gateway",
			Description: "Main API gateway for all services",
			Status:      "operational",
			GroupName:   "Backend",
			Order:       1,
			Visible:     true,
			ShowUptime:  true,
			Uptime:      99.87,
			Link:        "https://api.example.com",
			CreatedAt:   time.Now().Add(-25 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-30 * time.Minute),
		},
		{
			ID:          uuid.New(),
			TenantID:    "default-tenant",
			Name:        "User Service",
			Description: "User authentication and management service",
			Status:      "degraded_performance",
			GroupName:   "Backend",
			Order:       2,
			Visible:     true,
			ShowUptime:  true,
			Uptime:      98.2,
			CreatedAt:   time.Now().Add(-20 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-15 * time.Minute),
		},
		{
			ID:          uuid.New(),
			TenantID:    "default-tenant",
			Name:        "Database Primary",
			Description: "Primary PostgreSQL database cluster",
			Status:      "operational",
			GroupName:   "Database",
			Order:       1,
			Visible:     true,
			ShowUptime:  true,
			Uptime:      99.99,
			CreatedAt:   time.Now().Add(-45 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-2 * time.Hour),
		},
		{
			ID:          uuid.New(),
			TenantID:    "default-tenant",
			Name:        "Database Replica",
			Description: "Read replica database for reporting",
			Status:      "under_maintenance",
			GroupName:   "Database",
			Order:       2,
			Visible:     true,
			ShowUptime:  false,
			Uptime:      97.5,
			CreatedAt:   time.Now().Add(-40 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-10 * time.Minute),
		},
		{
			ID:          uuid.New(),
			TenantID:    "default-tenant",
			Name:        "CDN",
			Description: "Content delivery network for static assets",
			Status:      "operational",
			GroupName:   "Infrastructure",
			Order:       1,
			Visible:     true,
			ShowUptime:  true,
			Uptime:      100.0,
			CreatedAt:   time.Now().Add(-60 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-3 * time.Hour),
		},
		{
			ID:          uuid.New(),
			TenantID:    "default-tenant",
			Name:        "Email Service",
			Description: "Email notification and delivery service",
			Status:      "partial_outage",
			GroupName:   "Infrastructure",
			Order:       2,
			Visible:     true,
			ShowUptime:  true,
			Uptime:      95.1,
			CreatedAt:   time.Now().Add(-35 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-5 * time.Minute),
		},
		{
			ID:          uuid.New(),
			TenantID:    "default-tenant",
			Name:        "Internal Monitoring",
			Description: "Internal monitoring and alerting system",
			Status:      "operational",
			GroupName:   "Internal",
			Order:       1,
			Visible:     false,
			ShowUptime:  false,
			Uptime:      99.8,
			CreatedAt:   time.Now().Add(-50 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		},
	}

	// Filter by group if provided
	if groupName != "" {
		filtered := []models.SaaSComponent{}
		for _, comp := range components {
			if comp.GroupName == groupName {
				filtered = append(filtered, comp)
			}
		}
		components = filtered
	}

	// Filter by visibility if provided
	if visible != "" {
		showVisible := visible == "true"
		filtered := []models.SaaSComponent{}
		for _, comp := range components {
			if comp.Visible == showVisible {
				filtered = append(filtered, comp)
			}
		}
		components = filtered
	}

	// Group components by GroupName for better organization
	groupedComponents := make(map[string][]models.SaaSComponent)
	for _, comp := range components {
		groupedComponents[comp.GroupName] = append(groupedComponents[comp.GroupName], comp)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"components":        components,
		"grouped":           groupedComponents,
		"total":             len(components),
		"operational":       countByStatus(components, "operational"),
		"degraded":          countByStatus(components, "degraded_performance"),
		"partial_outage":    countByStatus(components, "partial_outage"),
		"major_outage":      countByStatus(components, "major_outage"),
		"under_maintenance": countByStatus(components, "under_maintenance"),
	})
}

// CreateComponent handles creating a new component
func (h *ComponentHandler) CreateComponent(c *gin.Context) {
	h.logger.Info("Creating component")

	var component models.SaaSComponent
	if err := c.ShouldBindJSON(&component); err != nil {
		h.logger.Error("Failed to bind component data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid component data",
		})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"operational":         true,
		"degraded_performance": true,
		"partial_outage":      true,
		"major_outage":        true,
		"under_maintenance":   true,
	}

	if !validStatuses[component.Status] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid component status",
		})
		return
	}

	// TODO: Implement actual component creation in database
	component.ID = uuid.New()
	component.TenantID = "default-tenant"
	component.CreatedAt = time.Now()
	component.UpdatedAt = time.Now()
	if component.Status == "" {
		component.Status = "operational"
	}
	if component.Uptime == 0 {
		component.Uptime = 100.0
	}
	component.Visible = true
	component.ShowUptime = true

	h.logger.Info("Component created successfully", zap.String("component_id", component.ID.String()))

	c.JSON(http.StatusCreated, gin.H{
		"success":   true,
		"component": component,
	})
}

// GetComponent handles getting a specific component
func (h *ComponentHandler) GetComponent(c *gin.Context) {
	h.logger.Info("Getting component")

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid component ID",
		})
		return
	}

	// Mock component data
	component := models.SaaSComponent{
		ID:          id,
		TenantID:    "default-tenant",
		Name:        "Web Application",
		Description: "Main web application frontend with user interface",
		Status:      "operational",
		GroupName:   "Frontend",
		Order:       1,
		Visible:     true,
		ShowUptime:  true,
		Uptime:      99.95,
		Link:        "https://app.example.com",
		CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt:   time.Now().Add(-1 * time.Hour),
	}

	// Mock recent status history
	statusHistory := []gin.H{
		{
			"timestamp": time.Now().Add(-2 * time.Hour),
			"status":    "operational",
			"message":   "All systems operational",
			"uptime":    99.95,
		},
		{
			"timestamp": time.Now().Add(-6 * time.Hour),
			"status":    "degraded_performance",
			"message":   "Experiencing slow response times",
			"uptime":    99.87,
		},
		{
			"timestamp": time.Now().Add(-12 * time.Hour),
			"status":    "operational",
			"message":   "Performance issues resolved",
			"uptime":    99.92,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"component":      component,
		"status_history": statusHistory,
	})
}

// UpdateComponent handles updating an existing component
func (h *ComponentHandler) UpdateComponent(c *gin.Context) {
	h.logger.Info("Updating component")

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid component ID",
		})
		return
	}

	var component models.SaaSComponent
	if err := c.ShouldBindJSON(&component); err != nil {
		h.logger.Error("Failed to bind component data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid component data",
		})
		return
	}

	component.ID = id
	component.UpdatedAt = time.Now()
	// TODO: Implement actual component update in database

	h.logger.Info("Component updated successfully", zap.String("component_id", id.String()))

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"component": component,
	})
}

// UpdateComponentStatus handles updating component status
func (h *ComponentHandler) UpdateComponentStatus(c *gin.Context) {
	h.logger.Info("Updating component status")

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid component ID",
		})
		return
	}

	var statusUpdate ComponentStatusUpdate
	if err := c.ShouldBindJSON(&statusUpdate); err != nil {
		h.logger.Error("Failed to bind status update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status update data",
		})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"operational":         true,
		"degraded_performance": true,
		"partial_outage":      true,
		"major_outage":        true,
		"under_maintenance":   true,
	}

	if !validStatuses[statusUpdate.Status] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid component status",
		})
		return
	}

	// TODO: Implement actual status update in database
	// TODO: Trigger notifications to subscribers
	// TODO: Update incident/maintenance if AffectedBy is provided

	result := gin.H{
		"component_id":     id,
		"previous_status":  "operational",
		"new_status":       statusUpdate.Status,
		"message":          statusUpdate.Message,
		"updated_by":       statusUpdate.UpdatedBy,
		"timestamp":        time.Now(),
		"notifications_sent": true,
		"subscribers_notified": 150,
	}

	h.logger.Info("Component status updated successfully",
		zap.String("component_id", id.String()),
		zap.String("new_status", statusUpdate.Status))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"result":  result,
	})
}

// DeleteComponent handles deleting a component
func (h *ComponentHandler) DeleteComponent(c *gin.Context) {
	h.logger.Info("Deleting component")

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid component ID",
		})
		return
	}

	// TODO: Implement actual component deletion in database
	// TODO: Check if component is referenced by incidents/maintenance

	h.logger.Info("Component deleted successfully", zap.String("component_id", id.String()))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Component deleted successfully",
	})
}

// GetComponentUptimeStats handles getting uptime statistics for a component
func (h *ComponentHandler) GetComponentUptimeStats(c *gin.Context) {
	h.logger.Info("Getting component uptime stats")

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid component ID",
		})
		return
	}

	// Parse time range
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days > 365 {
		days = 365
	}

	// Mock uptime statistics
	uptimeStats := gin.H{
		"component_id": id,
		"period_days":  days,
		"overall_uptime": 99.85,
		"daily_stats": []gin.H{
			{"date": "2024-01-15", "uptime": 100.0, "incidents": 0},
			{"date": "2024-01-14", "uptime": 99.2, "incidents": 1},
			{"date": "2024-01-13", "uptime": 100.0, "incidents": 0},
			{"date": "2024-01-12", "uptime": 98.5, "incidents": 2},
			{"date": "2024-01-11", "uptime": 100.0, "incidents": 0},
		},
		"incidents_count": 3,
		"maintenance_windows": 2,
		"avg_response_time": "245ms",
		"status_distribution": gin.H{
			"operational":         85.2,
			"degraded_performance": 12.1,
			"partial_outage":      2.5,
			"major_outage":        0.2,
			"under_maintenance":   0.0,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   uptimeStats,
	})
}

// GetComponentGroups handles getting all component groups
func (h *ComponentHandler) GetComponentGroups(c *gin.Context) {
	h.logger.Info("Getting component groups")

	// Mock component groups
	groups := []gin.H{
		{
			"name":        "Frontend",
			"description": "Frontend applications and user interfaces",
			"components":  2,
			"status":      "operational",
		},
		{
			"name":        "Backend",
			"description": "Backend services and APIs",
			"components":  3,
			"status":      "degraded_performance",
		},
		{
			"name":        "Database",
			"description": "Database systems and storage",
			"components":  2,
			"status":      "operational",
		},
		{
			"name":        "Infrastructure",
			"description": "Infrastructure and supporting services",
			"components":  3,
			"status":      "partial_outage",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"groups":  groups,
		"total":   len(groups),
	})
}

// Helper function to count components by status
func countByStatus(components []models.SaaSComponent, status string) int {
	count := 0
	for _, comp := range components {
		if comp.Status == status {
			count++
		}
	}
	return count
}