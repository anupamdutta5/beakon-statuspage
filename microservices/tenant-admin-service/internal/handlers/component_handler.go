// Package handlers provides HTTP request handlers for the Tenant Admin Service.
package handlers

import (
	"github.com/google/uuid"

	"net/http"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/anupamdutta5/tenant-admin-service/internal/models"
	"github.com/anupamdutta5/tenant-admin-service/internal/services"
)

// ComponentHandler handles component management HTTP requests.
type ComponentHandler struct {
	service *services.ComponentService
	logger  *zap.Logger
}

// NewComponentHandler creates a new component handler.
func NewComponentHandler(service *services.ComponentService, logger *zap.Logger) *ComponentHandler {
	return &ComponentHandler{
		service: service,
		logger:  logger,
	}
}

// getTenantIDFromContext extracts and validates tenant ID from context.
func (h *ComponentHandler) getTenantIDFromContext(c *gin.Context) (uuid.UUID, error) {
	tenantIDVal, exists := c.Get("tenant_id")
	if !exists {
		return uuid.Nil, http.ErrNoCookie // Using as a sentinel error
	}

	tenantID, ok := tenantIDVal.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("tenant_id is not a UUID")
	}

	return tenantID, nil
}

// CreateComponentRequest represents the request body for creating a component.
type CreateComponentRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Status      string `json:"status"` // operational, degraded_performance, partial_outage, major_outage, maintenance
	Position    int    `json:"position"`
	IsVisible   bool   `json:"is_visible"`
	GroupID     *uint  `json:"group_id"`
	Metadata    string `json:"metadata"`
}

// UpdateComponentRequest represents the request body for updating a component.
type UpdateComponentRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Position    *int    `json:"position"`
	IsVisible   *bool   `json:"is_visible"`
	GroupID     *uint   `json:"group_id"`
	Metadata    *string `json:"metadata"`
}

// UpdateComponentStatusRequest represents the request body for updating component status.
type UpdateComponentStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// ReorderComponentsRequest represents the request body for reordering components.
type ReorderComponentsRequest struct {
	Positions map[string]int `json:"positions" binding:"required"` // component_id: position
}

// GetComponents retrieves all components for a tenant.
// GET /api/v1/components
func (h *ComponentHandler) GetComponents(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	// Parse pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Validate limits
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Check for status filter
	statusFilter := c.Query("status")
	if statusFilter != "" {
		components, err := h.service.GetComponentsByStatus(c.Request.Context(), tenantID, statusFilter)
		if err != nil {
			h.logger.Error("Failed to get components by status",
				zap.String("tenant_id", tenantID.String()),
				zap.String("status", statusFilter),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch components"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": components, "total": len(components)})
		return
	}

	// Check for visible-only filter
	visibleOnly := c.Query("visible_only") == "true"
	if visibleOnly {
		components, err := h.service.GetVisibleComponents(c.Request.Context(), tenantID)
		if err != nil {
			h.logger.Error("Failed to get visible components",
				zap.String("tenant_id", tenantID.String()),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch components"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": components, "total": len(components)})
		return
	}

	// Fetch all components with pagination
	components, total, err := h.service.GetComponents(c.Request.Context(), tenantID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get components",
			zap.String("tenant_id", tenantID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch components"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   components,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetComponent retrieves a single component by ID.
// GET /api/v1/components/:id
func (h *ComponentHandler) GetComponent(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	componentIDStr := c.Param("id")
	componentID, err := uuid.Parse(componentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID format"})
		return
	}

	// Fetch component
	component, err := h.service.GetComponentByID(c.Request.Context(), tenantID, componentID)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Component not found"})
			return
		}
		h.logger.Error("Failed to get component",
			zap.String("tenant_id", tenantID.String()),
			zap.String("component_id", componentID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch component"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": component})
}

// CreateComponent creates a new component for a tenant.
// POST /api/v1/components
func (h *ComponentHandler) CreateComponent(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	var req CreateComponentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create component model
	component := &models.Component{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Visible:     req.IsVisible, // Map is_visible to visible
		// Note: position, group_id, metadata not in migration schema - ignoring
	}

	// Set default values
	if component.Status == "" {
		component.Status = "operational"
	}

	// Create component
	if err := h.service.CreateComponent(c.Request.Context(), tenantID, component); err != nil {
		h.logger.Error("Failed to create component",
			zap.String("tenant_id", tenantID.String()),
			zap.String("name", req.Name),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create component"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Component created successfully",
		"data":    component,
	})
}

// UpdateComponent updates component information.
// PUT /api/v1/components/:id
func (h *ComponentHandler) UpdateComponent(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	componentIDStr := c.Param("id")
	componentID, err := uuid.Parse(componentIDStr)
	if err != nil {
	}

	var req UpdateComponentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Position != nil {
		updates["position"] = *req.Position
	}
	if req.IsVisible != nil {
		updates["is_visible"] = *req.IsVisible
	}
	if req.GroupID != nil {
		updates["group_id"] = *req.GroupID
	}
	if req.Metadata != nil {
		updates["metadata"] = *req.Metadata
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Update component
	if err := h.service.UpdateComponent(c.Request.Context(), tenantID, componentID, updates); err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Component not found"})
			return
		}
		h.logger.Error("Failed to update component",
			zap.String("tenant_id", tenantID.String()),
			zap.String("component_id", componentID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update component"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Component updated successfully"})
}

// UpdateComponentStatus updates only the status of a component.
// PUT /api/v1/components/:id/status
func (h *ComponentHandler) UpdateComponentStatus(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	componentIDStr := c.Param("id")
	componentID, err := uuid.Parse(componentIDStr)
	if err != nil {
	}

	var req UpdateComponentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update component status
	if err := h.service.UpdateComponentStatus(c.Request.Context(), tenantID, componentID, req.Status); err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Component not found"})
			return
		}
		h.logger.Error("Failed to update component status",
			zap.String("tenant_id", tenantID.String()),
			zap.String("component_id", componentID.String()),
			zap.String("status", req.Status),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Component status updated successfully"})
}

// DeleteComponent soft-deletes a component.
// DELETE /api/v1/components/:id
func (h *ComponentHandler) DeleteComponent(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	componentIDStr := c.Param("id")
	componentID, err := uuid.Parse(componentIDStr)
	if err != nil {
	}

	// Delete component
	if err := h.service.DeleteComponent(c.Request.Context(), tenantID, componentID); err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Component not found"})
			return
		}
		h.logger.Error("Failed to delete component",
			zap.String("tenant_id", tenantID.String()),
			zap.String("component_id", componentID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete component"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Component deleted successfully"})
}

// GetComponentStats retrieves component statistics for a tenant.
// GET /api/v1/components/stats
func (h *ComponentHandler) GetComponentStats(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	// Get component stats
	stats, err := h.service.GetComponentStats(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get component stats",
			zap.String("tenant_id", tenantID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch component statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// ReorderComponents updates the position of components for custom ordering.
// POST /api/v1/components/reorder
func (h *ComponentHandler) ReorderComponents(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	var req ReorderComponentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert string keys to uint
	componentPositions := make(map[string]int)
	for componentIDStr, position := range req.Positions {
		componentPositions[componentIDStr] = position
	}

	// Reorder components
	if err := h.service.ReorderComponents(c.Request.Context(), tenantID, componentPositions); err != nil {
		h.logger.Error("Failed to reorder components",
			zap.String("tenant_id", tenantID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reorder components"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Components reordered successfully"})
}
