// Package handlers provides HTTP request handlers for the Tenant Admin Service.
package handlers

import (
	"github.com/google/uuid"

	"net/http"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/anupamdutta5/tenant-admin-service/internal/models"
	"github.com/anupamdutta5/tenant-admin-service/internal/services"
)

// IncidentHandler handles incident management HTTP requests.
type IncidentHandler struct {
	service *services.IncidentService
	logger  *zap.Logger
}

// NewIncidentHandler creates a new incident handler.
func NewIncidentHandler(service *services.IncidentService, logger *zap.Logger) *IncidentHandler {
	return &IncidentHandler{
		service: service,
		logger:  logger,
	}
}

// getTenantIDFromContext extracts and validates tenant ID from context.
func (h *IncidentHandler) getTenantIDFromContext(c *gin.Context) (uuid.UUID, error) {
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

// CreateIncidentRequest represents the request body for creating an incident.
type CreateIncidentRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	Status      string     `json:"status"`   // investigating, identified, monitoring, resolved
	Impact      string     `json:"impact"`   // none, minor, major, critical
	Severity    string     `json:"severity"` // low, medium, high, critical
	IsVisible   bool       `json:"is_visible"`
	StartedAt   *time.Time `json:"started_at"`
	Metadata    string     `json:"metadata"`
}

// UpdateIncidentRequest represents the request body for updating an incident.
type UpdateIncidentRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Impact      *string `json:"impact"`
	Severity    *string `json:"severity"`
	IsVisible   *bool   `json:"is_visible"`
	Metadata    *string `json:"metadata"`
}

// ResolveIncidentRequest represents the request body for resolving an incident.
type ResolveIncidentRequest struct {
	ResolvedBy *uuid.UUID `json:"resolved_by"` // Optional: User ID who resolved the incident
}

// GetIncidents retrieves all incidents for a tenant.
// GET /api/v1/incidents
func (h *IncidentHandler) GetIncidents(c *gin.Context) {
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
		incidents, err := h.service.GetIncidentsByStatus(c.Request.Context(), tenantID, statusFilter)
		if err != nil {
			h.logger.Error("Failed to get incidents by status",
				zap.String("tenant_id", tenantID.String()),
				zap.String("status", statusFilter),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch incidents"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": incidents, "total": len(incidents)})
		return
	}

	// Check for impact filter
	impactFilter := c.Query("impact")
	if impactFilter != "" {
		incidents, err := h.service.GetIncidentsByImpact(c.Request.Context(), tenantID, impactFilter)
		if err != nil {
			h.logger.Error("Failed to get incidents by impact",
				zap.String("tenant_id", tenantID.String()),
				zap.String("impact", impactFilter),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch incidents"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": incidents, "total": len(incidents)})
		return
	}

	// Check for active-only filter
	activeOnly := c.Query("active_only") == "true"
	if activeOnly {
		incidents, err := h.service.GetActiveIncidents(c.Request.Context(), tenantID)
		if err != nil {
			h.logger.Error("Failed to get active incidents",
				zap.String("tenant_id", tenantID.String()),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch incidents"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": incidents, "total": len(incidents)})
		return
	}

	// Check for visible-only filter
	visibleOnly := c.Query("visible_only") == "true"
	if visibleOnly {
		incidents, err := h.service.GetVisibleIncidents(c.Request.Context(), tenantID)
		if err != nil {
			h.logger.Error("Failed to get visible incidents",
				zap.String("tenant_id", tenantID.String()),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch incidents"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": incidents, "total": len(incidents)})
		return
	}

	// Check for recent filter
	if recentStr := c.Query("recent"); recentStr != "" {
		recentLimit, _ := strconv.Atoi(recentStr)
		if recentLimit < 1 {
			recentLimit = 10
		}
		if recentLimit > 50 {
			recentLimit = 50
		}
		incidents, err := h.service.GetRecentIncidents(c.Request.Context(), tenantID, recentLimit)
		if err != nil {
			h.logger.Error("Failed to get recent incidents",
				zap.String("tenant_id", tenantID.String()),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch incidents"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": incidents, "total": len(incidents)})
		return
	}

	// Fetch all incidents with pagination
	incidents, total, err := h.service.GetIncidents(c.Request.Context(), tenantID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get incidents",
			zap.String("tenant_id", tenantID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch incidents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   incidents,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetIncident retrieves a single incident by ID.
// GET /api/v1/incidents/:id
func (h *IncidentHandler) GetIncident(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	incidentIDStr := c.Param("id")
	incidentID, err := uuid.Parse(incidentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID format"})
		return
	}

	// Fetch incident
	incident, err := h.service.GetIncidentByID(c.Request.Context(), tenantID, incidentID)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
			return
		}
		h.logger.Error("Failed to get incident",
			zap.String("tenant_id", tenantID.String()),
			zap.String("incident_id", incidentID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch incident"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": incident})
}

// CreateIncident creates a new incident for a tenant.
// POST /api/v1/incidents
func (h *IncidentHandler) CreateIncident(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	var req CreateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create incident model
	incident := &models.Incident{
		TenantID:    tenantID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Impact:      req.Impact,
		Severity:    req.Severity,
		IsVisible:   req.IsVisible,
		Metadata:    req.Metadata,
	}

	// Set started_at from request or default to now
	if req.StartedAt != nil {
		incident.StartedAt = *req.StartedAt
	} else {
		incident.StartedAt = time.Now()
	}

	// Set default values
	if incident.Status == "" {
		incident.Status = "investigating"
	}
	if incident.Impact == "" {
		incident.Impact = "minor"
	}
	if incident.Severity == "" {
		incident.Severity = "low"
	}

	// Get user ID from context if available for created_by
	if userID, exists := c.Get("user_id"); exists {
		if userIDUUID, ok := userID.(uuid.UUID); ok {
			incident.CreatedBy = &userIDUUID
		}
	}

	// Create incident
	if err := h.service.CreateIncident(c.Request.Context(), tenantID, incident); err != nil {
		h.logger.Error("Failed to create incident",
			zap.String("tenant_id", tenantID.String()),
			zap.String("title", req.Title),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create incident"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Incident created successfully",
		"data":    incident,
	})
}

// UpdateIncident updates incident information.
// PUT /api/v1/incidents/:id
func (h *IncidentHandler) UpdateIncident(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	incidentIDStr := c.Param("id")
	incidentID, err := uuid.Parse(incidentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID format"})
		return
	}

	var req UpdateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Impact != nil {
		updates["impact"] = *req.Impact
	}
	if req.Severity != nil {
		updates["severity"] = *req.Severity
	}
	if req.IsVisible != nil {
		updates["is_visible"] = *req.IsVisible
	}
	if req.Metadata != nil {
		updates["metadata"] = *req.Metadata
	}

	// Get user ID from context if available for updated_by
	if userID, exists := c.Get("user_id"); exists {
		if userIDUUID, ok := userID.(uuid.UUID); ok {
			updates["updated_by"] = userIDUUID
		}
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Update incident
	if err := h.service.UpdateIncident(c.Request.Context(), tenantID, incidentID, updates); err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
			return
		}
		h.logger.Error("Failed to update incident",
			zap.String("tenant_id", tenantID.String()),
			zap.String("incident_id", incidentID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update incident"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incident updated successfully"})
}

// ResolveIncident marks an incident as resolved.
// POST /api/v1/incidents/:id/resolve
func (h *IncidentHandler) ResolveIncident(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	incidentIDStr := c.Param("id")
	incidentID, err := uuid.Parse(incidentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID format"})
		return
	}

	var req ResolveIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body for resolve
		req.ResolvedBy = nil
	}

	// Get user ID from context if not provided in request
	if req.ResolvedBy == nil {
		if userID, exists := c.Get("user_id"); exists {
			if userIDUUID, ok := userID.(uuid.UUID); ok {
				req.ResolvedBy = &userIDUUID
			}
		}
	}

	// Resolve incident
	if err := h.service.ResolveIncident(c.Request.Context(), tenantID, incidentID, req.ResolvedBy); err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
			return
		}
		h.logger.Error("Failed to resolve incident",
			zap.String("tenant_id", tenantID.String()),
			zap.String("incident_id", incidentID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incident resolved successfully"})
}

// DeleteIncident soft-deletes an incident.
// DELETE /api/v1/incidents/:id
func (h *IncidentHandler) DeleteIncident(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	incidentIDStr := c.Param("id")
	incidentID, err := uuid.Parse(incidentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID format"})
		return
	}

	// Delete incident
	if err := h.service.DeleteIncident(c.Request.Context(), tenantID, incidentID); err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
			return
		}
		h.logger.Error("Failed to delete incident",
			zap.String("tenant_id", tenantID.String()),
			zap.String("incident_id", incidentID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete incident"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incident deleted successfully"})
}

// GetIncidentStats retrieves incident statistics for a tenant.
// GET /api/v1/incidents/stats
func (h *IncidentHandler) GetIncidentStats(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	// Get incident stats
	stats, err := h.service.GetIncidentStats(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get incident stats",
			zap.String("tenant_id", tenantID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch incident statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}
