// Package handlers provides HTTP handlers for the Component Service.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/enterprise-status/statuspage-component-service/internal/models"
	"github.com/enterprise-status/statuspage-component-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ComponentHandler handles component-related HTTP requests.
type ComponentHandler struct {
	componentService *services.ComponentService
	logger           *zap.Logger
}

// NewComponentHandler creates a new component handler.
func NewComponentHandler(componentService *services.ComponentService, logger *zap.Logger) *ComponentHandler {
	return &ComponentHandler{
		componentService: componentService,
		logger:           logger,
	}
}

// Health returns the health status of the Component Service.
func (h *ComponentHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "component-service",
		"version":   "1.0.0",
		"timestamp": "2024-01-01T00:00:00Z",
	})
}

// GetPublicStatus returns the overall status for a tenant.
func (h *ComponentHandler) GetPublicStatus(c *gin.Context) {
	// For now, use tenant ID 1 as default
	// In production, this would be determined from the request context
	tenantID := uint(1)

	status, err := h.componentService.GetPublicStatus(tenantID)
	if err != nil {
		h.logger.Error("Failed to get public status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get status"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetPublicComponents returns public components for a tenant.
func (h *ComponentHandler) GetPublicComponents(c *gin.Context) {
	// For now, use tenant ID 1 as default
	// In production, this would be determined from the request context
	tenantID := uint(1)

	components, err := h.componentService.GetPublicComponents(tenantID)
	if err != nil {
		h.logger.Error("Failed to get public components", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get components"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"components": components})
}

// GetPublicComponent returns a specific public component.
func (h *ComponentHandler) GetPublicComponent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID"})
		return
	}

	component, err := h.componentService.GetComponent(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Component not found"})
		return
	}

	// Check if component is visible
	if !component.IsVisible {
		c.JSON(http.StatusNotFound, gin.H{"error": "Component not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"component": component})
}

// GetComponents handles getting a list of components.
func (h *ComponentHandler) GetComponents(c *gin.Context) {
	// Get tenant ID from context (set by auth middleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	components, total, err := h.componentService.GetComponents(tenantID.(uint), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get components", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get components"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"components": components,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}

// GetComponent handles getting a specific component.
func (h *ComponentHandler) GetComponent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID"})
		return
	}

	component, err := h.componentService.GetComponent(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Component not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"component": component})
}

// CreateComponent handles creating a new component.
func (h *ComponentHandler) CreateComponent(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Status      string `json:"status"`
		Position    int    `json:"position"`
		IsVisible   *bool  `json:"is_visible"`
		GroupID     *uint  `json:"group_id"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create component request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	component := &models.Component{
		TenantID:    tenantID.(uint),
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Position:    req.Position,
		GroupID:     req.GroupID,
		Metadata:    req.Metadata,
	}

	if req.IsVisible != nil {
		component.IsVisible = *req.IsVisible
	}

	if err := h.componentService.CreateComponent(component); err != nil {
		h.logger.Error("Failed to create component", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create component"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Component created successfully",
		"component": component,
	})
}

// UpdateComponent handles updating a component.
func (h *ComponentHandler) UpdateComponent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID"})
		return
	}

	component, err := h.componentService.GetComponent(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Component not found"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
		Position    *int   `json:"position"`
		IsVisible   *bool  `json:"is_visible"`
		GroupID     *uint  `json:"group_id"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update component request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update fields
	if req.Name != "" {
		component.Name = req.Name
	}
	if req.Description != "" {
		component.Description = req.Description
	}
	if req.Status != "" {
		component.Status = req.Status
	}
	if req.Position != nil {
		component.Position = *req.Position
	}
	if req.IsVisible != nil {
		component.IsVisible = *req.IsVisible
	}
	if req.GroupID != nil {
		component.GroupID = req.GroupID
	}
	if req.Metadata != "" {
		component.Metadata = req.Metadata
	}

	if err := h.componentService.UpdateComponent(component); err != nil {
		h.logger.Error("Failed to update component", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update component"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Component updated successfully",
		"component": component,
	})
}

// UpdateComponentStatus handles updating a component's status.
func (h *ComponentHandler) UpdateComponentStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID"})
		return
	}

	var req struct {
		Status    string `json:"status" binding:"required"`
		Message   string `json:"message"`
		UpdatedBy uint   `json:"updated_by"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update status request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Get user ID from context if not provided
	updatedBy := req.UpdatedBy
	if updatedBy == 0 {
		if userID, exists := c.Get("user_id"); exists {
			updatedBy = userID.(uint)
		}
	}

	if err := h.componentService.UpdateComponentStatus(uint(id), req.Status, req.Message, updatedBy); err != nil {
		h.logger.Error("Failed to update component status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update component status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Component status updated successfully"})
}

// DeleteComponent handles deleting a component.
func (h *ComponentHandler) DeleteComponent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID"})
		return
	}

	if err := h.componentService.DeleteComponent(uint(id)); err != nil {
		h.logger.Error("Failed to delete component", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete component"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Component deleted successfully"})
}

// GetComponentHistory handles getting component status history.
func (h *ComponentHandler) GetComponentHistory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	history, total, err := h.componentService.GetComponentHistory(uint(id), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get component history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get component history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// Component Group Handlers

// GetComponentGroups handles getting a list of component groups.
func (h *ComponentHandler) GetComponentGroups(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	groups, total, err := h.componentService.GetComponentGroups(tenantID.(uint), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get component groups", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get component groups"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetComponentGroup handles getting a specific component group.
func (h *ComponentHandler) GetComponentGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	group, err := h.componentService.GetComponentGroup(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Component group not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"group": group})
}

// CreateComponentGroup handles creating a new component group.
func (h *ComponentHandler) CreateComponentGroup(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Position    int    `json:"position"`
		IsVisible   *bool  `json:"is_visible"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create group request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	group := &models.ComponentGroup{
		TenantID:    tenantID.(uint),
		Name:        req.Name,
		Description: req.Description,
		Position:    req.Position,
	}

	if req.IsVisible != nil {
		group.IsVisible = *req.IsVisible
	}

	if err := h.componentService.CreateComponentGroup(group); err != nil {
		h.logger.Error("Failed to create component group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create component group"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Component group created successfully",
		"group":   group,
	})
}

// UpdateComponentGroup handles updating a component group.
func (h *ComponentHandler) UpdateComponentGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	group, err := h.componentService.GetComponentGroup(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Component group not found"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Position    *int   `json:"position"`
		IsVisible   *bool  `json:"is_visible"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update group request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update fields
	if req.Name != "" {
		group.Name = req.Name
	}
	if req.Description != "" {
		group.Description = req.Description
	}
	if req.Position != nil {
		group.Position = *req.Position
	}
	if req.IsVisible != nil {
		group.IsVisible = *req.IsVisible
	}

	if err := h.componentService.UpdateComponentGroup(group); err != nil {
		h.logger.Error("Failed to update component group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update component group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Component group updated successfully",
		"group":   group,
	})
}

// DeleteComponentGroup handles deleting a component group.
func (h *ComponentHandler) DeleteComponentGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	if err := h.componentService.DeleteComponentGroup(uint(id)); err != nil {
		h.logger.Error("Failed to delete component group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete component group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Component group deleted successfully"})
}
