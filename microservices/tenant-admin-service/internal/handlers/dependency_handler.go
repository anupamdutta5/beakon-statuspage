// Package handlers provides HTTP handlers for dependency management.
package handlers

import (
	"net/http"

	"github.com/anupamdutta5/tenant-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DependencyHandler handles HTTP requests for dependency operations.
type DependencyHandler struct {
	dependencyService *services.DependencyService
}

// NewDependencyHandler creates a new DependencyHandler instance.
func NewDependencyHandler(dependencyService *services.DependencyService) *DependencyHandler {
	return &DependencyHandler{
		dependencyService: dependencyService,
	}
}

// GetDependencyGraph retrieves the complete dependency graph for the tenant.
// GET /api/v1/dependencies/graph
func (h *DependencyHandler) GetDependencyGraph(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	graph, err := h.dependencyService.GetDependencyGraph(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch dependency graph",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, graph)
}

// AddDependency creates a new dependency edge between components.
// POST /api/v1/dependencies
// Request body:
// {
//   "from_component_id": "uuid",
//   "to_component_id": "uuid",
//   "dependency_type": "hard" | "soft"
// }
func (h *DependencyHandler) AddDependency(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var req struct {
		FromComponentID string `json:"from_component_id" binding:"required"`
		ToComponentID   string `json:"to_component_id" binding:"required"`
		DependencyType  string `json:"dependency_type" binding:"required,oneof=hard soft"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	fromComponentID, err := uuid.Parse(req.FromComponentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from_component_id"})
		return
	}

	toComponentID, err := uuid.Parse(req.ToComponentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to_component_id"})
		return
	}

	if err := h.dependencyService.AddDependency(tenantID, fromComponentID, toComponentID, req.DependencyType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to add dependency",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Dependency created successfully",
	})
}

// RemoveDependency deletes a dependency edge.
// DELETE /api/v1/dependencies/:from_id/:to_id
func (h *DependencyHandler) RemoveDependency(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	fromComponentID, err := uuid.Parse(c.Param("from_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from_component_id"})
		return
	}

	toComponentID, err := uuid.Parse(c.Param("to_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to_component_id"})
		return
	}

	if err := h.dependencyService.RemoveDependency(tenantID, fromComponentID, toComponentID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Failed to remove dependency",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Dependency removed successfully",
	})
}

// AnalyzeImpact performs impact analysis for a component failure.
// GET /api/v1/dependencies/impact/:component_id
func (h *DependencyHandler) AnalyzeImpact(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	componentID, err := uuid.Parse(c.Param("component_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID"})
		return
	}

	impact, err := h.dependencyService.AnalyzeImpact(tenantID, componentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Failed to analyze impact",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, impact)
}

// GetDependencyHealth retrieves the health status of a component's dependencies.
// GET /api/v1/dependencies/health/:component_id
func (h *DependencyHandler) GetDependencyHealth(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	componentID, err := uuid.Parse(c.Param("component_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID"})
		return
	}

	health, err := h.dependencyService.GetDependencyHealth(tenantID, componentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Failed to get dependency health",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, health)
}

// ValidateDependency checks if adding a dependency would create a cycle.
// POST /api/v1/dependencies/validate
// Request body:
// {
//   "from_component_id": "uuid",
//   "to_component_id": "uuid"
// }
func (h *DependencyHandler) ValidateDependency(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var req struct {
		FromComponentID string `json:"from_component_id" binding:"required"`
		ToComponentID   string `json:"to_component_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	fromComponentID, err := uuid.Parse(req.FromComponentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from_component_id"})
		return
	}

	toComponentID, err := uuid.Parse(req.ToComponentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to_component_id"})
		return
	}

	// Use the validate method from service (we'll need to expose it)
	// For now, try to add and rollback if it fails
	err = h.dependencyService.AddDependency(tenantID, fromComponentID, toComponentID, "hard")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"valid":         false,
			"error_message": err.Error(),
		})
		return
	}

	// Rollback - remove the dependency we just added
	_ = h.dependencyService.RemoveDependency(tenantID, fromComponentID, toComponentID)

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
	})
}
