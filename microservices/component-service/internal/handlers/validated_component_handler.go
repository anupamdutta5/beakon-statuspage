// Package handlers provides HTTP handlers for the Component Service with comprehensive validation.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/anupamdutta5/component-service/internal/models"
	"github.com/anupamdutta5/component-service/internal/services"
	"github.com/anupamdutta5/component-service/internal/validation"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ValidatedComponentHandler handles component-related HTTP requests with comprehensive validation.
type ValidatedComponentHandler struct {
	componentService *services.ComponentService
	validator        *validation.ServiceValidator
	logger           *zap.Logger
}

// NewValidatedComponentHandler creates a new validated component handler.
func NewValidatedComponentHandler(componentService *services.ComponentService, logger *zap.Logger) *ValidatedComponentHandler {
	return &ValidatedComponentHandler{
		componentService: componentService,
		validator:        validation.NewServiceValidator(logger),
		logger:           logger,
	}
}

// CreateComponentRequest represents a validated component creation request
type CreateComponentRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Position    int    `json:"position"`
	IsVisible   *bool  `json:"is_visible"`
	GroupID     *uint  `json:"group_id"`
	Metadata    string `json:"metadata"`
}

// Validate validates and sanitizes the create component request
func (r *CreateComponentRequest) Validate(validator *validation.ServiceValidator) error {
	var err error

	// Validate and sanitize name
	r.Name, err = validator.ValidateAndSanitizeString(r.Name, "name", true)
	if err != nil {
		return err
	}

	// Validate name length
	if err := validator.ValidateStringLength(r.Name, 1, 255, "name"); err != nil {
		return err
	}

	// Validate and sanitize description
	r.Description, err = validator.ValidateAndSanitizeString(r.Description, "description", false)
	if err != nil {
		return err
	}

	// Validate description length
	if r.Description != "" {
		if err := validator.ValidateStringLength(r.Description, 0, 1000, "description"); err != nil {
			return err
		}
	}

	// Validate status enum
	allowedStatuses := []string{"operational", "degraded", "partial_outage", "major_outage", "under_maintenance"}
	if r.Status != "" {
		r.Status, err = validator.ValidateEnum(r.Status, allowedStatuses, "status")
		if err != nil {
			return err
		}
	} else {
		r.Status = "operational" // Default status
	}

	// Validate position range
	if err := validator.ValidateRange(r.Position, 0, 9999, "position"); err != nil {
		return err
	}

	// Validate group ID if provided
	if r.GroupID != nil {
		_, err = validator.ValidateID(*r.GroupID, "group_id")
		if err != nil {
			return err
		}
	}

	// Validate and sanitize metadata
	r.Metadata, err = validator.ValidateAndSanitizeString(r.Metadata, "metadata", false)
	if err != nil {
		return err
	}

	// Validate metadata length
	if r.Metadata != "" {
		if err := validator.ValidateStringLength(r.Metadata, 0, 2000, "metadata"); err != nil {
			return err
		}
	}

	return nil
}

// UpdateComponentRequest represents a validated component update request
type UpdateComponentRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Position    *int   `json:"position"`
	IsVisible   *bool  `json:"is_visible"`
	GroupID     *uint  `json:"group_id"`
	Metadata    string `json:"metadata"`
}

// Validate validates and sanitizes the update component request
func (r *UpdateComponentRequest) Validate(validator *validation.ServiceValidator) error {
	var err error

	// Validate and sanitize name if provided
	if r.Name != "" {
		r.Name, err = validator.ValidateAndSanitizeString(r.Name, "name", false)
		if err != nil {
			return err
		}

		if err := validator.ValidateStringLength(r.Name, 1, 255, "name"); err != nil {
			return err
		}
	}

	// Validate and sanitize description if provided
	if r.Description != "" {
		r.Description, err = validator.ValidateAndSanitizeString(r.Description, "description", false)
		if err != nil {
			return err
		}

		if err := validator.ValidateStringLength(r.Description, 0, 1000, "description"); err != nil {
			return err
		}
	}

	// Validate status enum if provided
	if r.Status != "" {
		allowedStatuses := []string{"operational", "degraded", "partial_outage", "major_outage", "under_maintenance"}
		r.Status, err = validator.ValidateEnum(r.Status, allowedStatuses, "status")
		if err != nil {
			return err
		}
	}

	// Validate position range if provided
	if r.Position != nil {
		if err := validator.ValidateRange(*r.Position, 0, 9999, "position"); err != nil {
			return err
		}
	}

	// Validate group ID if provided
	if r.GroupID != nil {
		_, err = validator.ValidateID(*r.GroupID, "group_id")
		if err != nil {
			return err
		}
	}

	// Validate and sanitize metadata if provided
	if r.Metadata != "" {
		r.Metadata, err = validator.ValidateAndSanitizeString(r.Metadata, "metadata", false)
		if err != nil {
			return err
		}

		if err := validator.ValidateStringLength(r.Metadata, 0, 2000, "metadata"); err != nil {
			return err
		}
	}

	return nil
}

// UpdateComponentStatusRequest represents a validated component status update request
type UpdateComponentStatusRequest struct {
	Status    string `json:"status" binding:"required"`
	Message   string `json:"message"`
	UpdatedBy uint   `json:"updated_by"`
}

// Validate validates and sanitizes the status update request
func (r *UpdateComponentStatusRequest) Validate(validator *validation.ServiceValidator) error {
	var err error

	// Validate status enum
	allowedStatuses := []string{"operational", "degraded", "partial_outage", "major_outage", "under_maintenance"}
	r.Status, err = validator.ValidateEnum(r.Status, allowedStatuses, "status")
	if err != nil {
		return err
	}

	// Validate and sanitize message
	r.Message, err = validator.ValidateAndSanitizeString(r.Message, "message", false)
	if err != nil {
		return err
	}

	// Validate message length
	if r.Message != "" {
		if err := validator.ValidateStringLength(r.Message, 0, 500, "message"); err != nil {
			return err
		}
	}

	// Validate updated_by if provided
	if r.UpdatedBy != 0 {
		_, err = validator.ValidateID(r.UpdatedBy, "updated_by")
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateComponent handles creating a new component with comprehensive validation.
func (h *ValidatedComponentHandler) CreateComponent(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		h.logger.Warn("Tenant ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req CreateComponentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid JSON in create component request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate and sanitize the request
	if err := req.Validate(h.validator); err != nil {
		h.logger.Warn("Component creation validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate tenant ID
	validatedTenantID, err := h.validator.ValidateID(tenantID, "tenant_id")
	if err != nil {
		h.logger.Error("Invalid tenant ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	component := &models.Component{
		TenantID:    validatedTenantID,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Position:    req.Position,
		GroupID:     req.GroupID,
		Metadata:    req.Metadata,
	}

	if req.IsVisible != nil {
		component.IsVisible = *req.IsVisible
	} else {
		component.IsVisible = true // Default to visible
	}

	if err := h.componentService.CreateComponent(component); err != nil {
		h.logger.Error("Failed to create component", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create component"})
		return
	}

	h.logger.Info("Component created successfully",
		zap.String("component_name", component.Name),
		zap.Uint("tenant_id", validatedTenantID),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Component created successfully",
		"component": component,
	})
}

// UpdateComponent handles updating a component with comprehensive validation.
func (h *ValidatedComponentHandler) UpdateComponent(c *gin.Context) {
	// Validate component ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Warn("Invalid component ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID format"})
		return
	}

	validatedID, err := h.validator.ValidateID(uint(id), "component_id")
	if err != nil {
		h.logger.Warn("Invalid component ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	component, err := h.componentService.GetComponent(validatedID)
	if err != nil {
		h.logger.Error("Component not found", zap.Uint("component_id", validatedID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Component not found"})
		return
	}

	var req UpdateComponentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid JSON in update component request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate and sanitize the request
	if err := req.Validate(h.validator); err != nil {
		h.logger.Warn("Component update validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields with validated data
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

	h.logger.Info("Component updated successfully",
		zap.Uint("component_id", validatedID),
		zap.String("component_name", component.Name),
	)

	c.JSON(http.StatusOK, gin.H{
		"message":   "Component updated successfully",
		"component": component,
	})
}

// UpdateComponentStatus handles updating a component's status with comprehensive validation.
func (h *ValidatedComponentHandler) UpdateComponentStatus(c *gin.Context) {
	// Validate component ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Warn("Invalid component ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID format"})
		return
	}

	validatedID, err := h.validator.ValidateID(uint(id), "component_id")
	if err != nil {
		h.logger.Warn("Invalid component ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req UpdateComponentStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid JSON in status update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate and sanitize the request
	if err := req.Validate(h.validator); err != nil {
		h.logger.Warn("Status update validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context if not provided
	updatedBy := req.UpdatedBy
	if updatedBy == 0 {
		if userID, exists := c.Get("user_id"); exists {
			if validatedUserID, err := h.validator.ValidateID(userID, "user_id"); err == nil {
				updatedBy = validatedUserID
			}
		}
	}

	if err := h.componentService.UpdateComponentStatus(validatedID, req.Status, req.Message, updatedBy); err != nil {
		h.logger.Error("Failed to update component status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update component status"})
		return
	}

	h.logger.Info("Component status updated successfully",
		zap.Uint("component_id", validatedID),
		zap.String("new_status", req.Status),
		zap.Uint("updated_by", updatedBy),
	)

	c.JSON(http.StatusOK, gin.H{"message": "Component status updated successfully"})
}

// GetComponents handles getting a list of components with pagination validation.
func (h *ValidatedComponentHandler) GetComponents(c *gin.Context) {
	// Get tenant ID from context (set by auth middleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		h.logger.Warn("Tenant ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	// Validate tenant ID
	validatedTenantID, err := h.validator.ValidateID(tenantID, "tenant_id")
	if err != nil {
		h.logger.Error("Invalid tenant ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// Parse and validate pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Validate pagination ranges
	if err := h.validator.ValidateRange(limit, 1, 100, "limit"); err != nil {
		h.logger.Warn("Invalid limit parameter", zap.Int("limit", limit))
		limit = 10 // Use default
	}

	if err := h.validator.ValidateRange(offset, 0, 10000, "offset"); err != nil {
		h.logger.Warn("Invalid offset parameter", zap.Int("offset", offset))
		offset = 0 // Use default
	}

	components, total, err := h.componentService.GetComponents(validatedTenantID, limit, offset)
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

// Health returns the health status of the Component Service (no validation needed).
func (h *ValidatedComponentHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "component-service",
		"version":   "1.0.0",
		"timestamp": "2024-01-01T00:00:00Z",
	})
}

// Delegate remaining methods to original handler for consistency
// These can be enhanced with validation as needed

func (h *ValidatedComponentHandler) GetPublicStatus(c *gin.Context) {
	// Delegate to original handler since this is a public endpoint with minimal validation needs
	tenantID := uint(1) // Default tenant for public view

	status, err := h.componentService.GetPublicStatus(tenantID)
	if err != nil {
		h.logger.Error("Failed to get public status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get status"})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *ValidatedComponentHandler) GetPublicComponents(c *gin.Context) {
	tenantID := uint(1) // Default tenant for public view

	components, err := h.componentService.GetPublicComponents(tenantID)
	if err != nil {
		h.logger.Error("Failed to get public components", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get components"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"components": components})
}

func (h *ValidatedComponentHandler) GetPublicComponent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID"})
		return
	}

	validatedID, err := h.validator.ValidateID(uint(id), "component_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	component, err := h.componentService.GetComponent(validatedID)
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

// Add more validated methods as needed...
// For brevity, I'm showing the pattern for the most critical endpoints