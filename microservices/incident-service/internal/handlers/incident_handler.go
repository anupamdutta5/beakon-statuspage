// Package handlers provides HTTP handlers for the Incident Service.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/enterprise-status/statuspage-incident-service/internal/models"
	"github.com/enterprise-status/statuspage-incident-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// IncidentHandler handles incident-related HTTP requests.
type IncidentHandler struct {
	incidentService *services.IncidentService
	logger          *zap.Logger
}

// NewIncidentHandler creates a new incident handler.
func NewIncidentHandler(incidentService *services.IncidentService, logger *zap.Logger) *IncidentHandler {
	return &IncidentHandler{
		incidentService: incidentService,
		logger:          logger,
	}
}

// Health returns the health status of the Incident Service.
func (h *IncidentHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "incident-service",
		"version":   "1.0.0",
		"timestamp": "2024-01-01T00:00:00Z",
	})
}

// GetPublicIncidents returns public incidents for a tenant.
func (h *IncidentHandler) GetPublicIncidents(c *gin.Context) {
	// For now, use tenant ID 1 as default
	// In production, this would be determined from the request context
	tenantID := uint(1)

	incidents, err := h.incidentService.GetPublicIncidents(tenantID)
	if err != nil {
		h.logger.Error("Failed to get public incidents", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get incidents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"incidents": incidents})
}

// GetPublicIncident returns a specific public incident.
func (h *IncidentHandler) GetPublicIncident(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	incident, err := h.incidentService.GetIncident(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
		return
	}

	// Check if incident is visible
	if !incident.IsVisible {
		c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"incident": incident})
}

// GetIncidents handles getting a list of incidents.
func (h *IncidentHandler) GetIncidents(c *gin.Context) {
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

	incidents, total, err := h.incidentService.GetIncidents(tenantID.(uint), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get incidents", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get incidents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"incidents": incidents,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetIncident handles getting a specific incident.
func (h *IncidentHandler) GetIncident(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	incident, err := h.incidentService.GetIncident(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"incident": incident})
}

// CreateIncident handles creating a new incident.
func (h *IncidentHandler) CreateIncident(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var req struct {
		Title       string    `json:"title" binding:"required"`
		Description string    `json:"description"`
		Status      string    `json:"status"`
		Impact      string    `json:"impact"`
		Severity    string    `json:"severity"`
		IsVisible   *bool     `json:"is_visible"`
		StartedAt   time.Time `json:"started_at"`
		Metadata    string    `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create incident request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	incident := &models.Incident{
		TenantID:    tenantID.(uint),
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Impact:      req.Impact,
		Severity:    req.Severity,
		StartedAt:   req.StartedAt,
		CreatedBy:   userID.(uint),
		UpdatedBy:   userID.(uint),
		Metadata:    req.Metadata,
	}

	if req.IsVisible != nil {
		incident.IsVisible = *req.IsVisible
	}

	if err := h.incidentService.CreateIncident(incident); err != nil {
		h.logger.Error("Failed to create incident", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create incident"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Incident created successfully",
		"incident": incident,
	})
}

// UpdateIncident handles updating an incident.
func (h *IncidentHandler) UpdateIncident(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	incident, err := h.incidentService.GetIncident(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var req struct {
		Title       string     `json:"title"`
		Description string     `json:"description"`
		Status      string     `json:"status"`
		Impact      string     `json:"impact"`
		Severity    string     `json:"severity"`
		IsVisible   *bool      `json:"is_visible"`
		StartedAt   *time.Time `json:"started_at"`
		ResolvedAt  *time.Time `json:"resolved_at"`
		Metadata    string     `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update incident request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update fields
	if req.Title != "" {
		incident.Title = req.Title
	}
	if req.Description != "" {
		incident.Description = req.Description
	}
	if req.Status != "" {
		incident.Status = req.Status
	}
	if req.Impact != "" {
		incident.Impact = req.Impact
	}
	if req.Severity != "" {
		incident.Severity = req.Severity
	}
	if req.IsVisible != nil {
		incident.IsVisible = *req.IsVisible
	}
	if req.StartedAt != nil {
		incident.StartedAt = *req.StartedAt
	}
	if req.ResolvedAt != nil {
		incident.ResolvedAt = req.ResolvedAt
	}
	if req.Metadata != "" {
		incident.Metadata = req.Metadata
	}

	incident.UpdatedBy = userID.(uint)

	if err := h.incidentService.UpdateIncident(incident); err != nil {
		h.logger.Error("Failed to update incident", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update incident"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Incident updated successfully",
		"incident": incident,
	})
}

// UpdateIncidentStatus handles updating an incident's status.
func (h *IncidentHandler) UpdateIncidentStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update status request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	if err := h.incidentService.UpdateIncidentStatus(uint(id), req.Status, userID.(uint)); err != nil {
		h.logger.Error("Failed to update incident status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update incident status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incident status updated successfully"})
}

// DeleteIncident handles deleting an incident.
func (h *IncidentHandler) DeleteIncident(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	if err := h.incidentService.DeleteIncident(uint(id)); err != nil {
		h.logger.Error("Failed to delete incident", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete incident"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incident deleted successfully"})
}

// AddIncidentUpdate handles adding an update to an incident.
func (h *IncidentHandler) AddIncidentUpdate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var req struct {
		Status    string `json:"status" binding:"required"`
		Message   string `json:"message" binding:"required"`
		IsVisible *bool  `json:"is_visible"`
		Metadata  string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid add update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	update := &models.IncidentUpdate{
		IncidentID: uint(id),
		Status:     req.Status,
		Message:    req.Message,
		CreatedBy:  userID.(uint),
		Metadata:   req.Metadata,
	}

	if req.IsVisible != nil {
		update.IsVisible = *req.IsVisible
	}

	if err := h.incidentService.AddIncidentUpdate(update); err != nil {
		h.logger.Error("Failed to add incident update", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add incident update"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Incident update added successfully",
		"update":  update,
	})
}

// UpdateIncidentUpdate handles updating an incident update.
func (h *IncidentHandler) UpdateIncidentUpdate(c *gin.Context) {
	_, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	updateID, err := strconv.ParseUint(c.Param("update_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update ID"})
		return
	}

	var req struct {
		Status    string `json:"status"`
		Message   string `json:"message"`
		IsVisible *bool  `json:"is_visible"`
		Metadata  string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Get the update - we need to add this method to the service
	// For now, we'll create a simple update object
	update := &models.IncidentUpdate{
		ID: uint(updateID),
	}

	// Update fields
	if req.Status != "" {
		update.Status = req.Status
	}
	if req.Message != "" {
		update.Message = req.Message
	}
	if req.IsVisible != nil {
		update.IsVisible = *req.IsVisible
	}
	if req.Metadata != "" {
		update.Metadata = req.Metadata
	}

	if err := h.incidentService.UpdateIncidentUpdate(update); err != nil {
		h.logger.Error("Failed to update incident update", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update incident update"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Incident update updated successfully",
		"update":  update,
	})
}

// DeleteIncidentUpdate handles deleting an incident update.
func (h *IncidentHandler) DeleteIncidentUpdate(c *gin.Context) {
	updateID, err := strconv.ParseUint(c.Param("update_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update ID"})
		return
	}

	if err := h.incidentService.DeleteIncidentUpdate(uint(updateID)); err != nil {
		h.logger.Error("Failed to delete incident update", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete incident update"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incident update deleted successfully"})
}

// Incident Template Handlers

// GetIncidentTemplates handles getting a list of incident templates.
func (h *IncidentHandler) GetIncidentTemplates(c *gin.Context) {
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

	templates, total, err := h.incidentService.GetIncidentTemplates(tenantID.(uint), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get incident templates", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get incident templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetIncidentTemplate handles getting a specific incident template.
func (h *IncidentHandler) GetIncidentTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	template, err := h.incidentService.GetIncidentTemplate(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Incident template not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"template": template})
}

// CreateIncidentTemplate handles creating a new incident template.
func (h *IncidentHandler) CreateIncidentTemplate(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Title       string `json:"title" binding:"required"`
		Message     string `json:"message"`
		Impact      string `json:"impact"`
		Severity    string `json:"severity"`
		IsActive    *bool  `json:"is_active"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create template request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	template := &models.IncidentTemplate{
		TenantID:    tenantID.(uint),
		Name:        req.Name,
		Description: req.Description,
		Title:       req.Title,
		Message:     req.Message,
		Impact:      req.Impact,
		Severity:    req.Severity,
		CreatedBy:   userID.(uint),
		UpdatedBy:   userID.(uint),
		Metadata:    req.Metadata,
	}

	if req.IsActive != nil {
		template.IsActive = *req.IsActive
	}

	if err := h.incidentService.CreateIncidentTemplate(template); err != nil {
		h.logger.Error("Failed to create incident template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create incident template"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Incident template created successfully",
		"template": template,
	})
}

// UpdateIncidentTemplate handles updating an incident template.
func (h *IncidentHandler) UpdateIncidentTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	template, err := h.incidentService.GetIncidentTemplate(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Incident template not found"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Title       string `json:"title"`
		Message     string `json:"message"`
		Impact      string `json:"impact"`
		Severity    string `json:"severity"`
		IsActive    *bool  `json:"is_active"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update template request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update fields
	if req.Name != "" {
		template.Name = req.Name
	}
	if req.Description != "" {
		template.Description = req.Description
	}
	if req.Title != "" {
		template.Title = req.Title
	}
	if req.Message != "" {
		template.Message = req.Message
	}
	if req.Impact != "" {
		template.Impact = req.Impact
	}
	if req.Severity != "" {
		template.Severity = req.Severity
	}
	if req.IsActive != nil {
		template.IsActive = *req.IsActive
	}
	if req.Metadata != "" {
		template.Metadata = req.Metadata
	}

	template.UpdatedBy = userID.(uint)

	if err := h.incidentService.UpdateIncidentTemplate(template); err != nil {
		h.logger.Error("Failed to update incident template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update incident template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Incident template updated successfully",
		"template": template,
	})
}

// DeleteIncidentTemplate handles deleting an incident template.
func (h *IncidentHandler) DeleteIncidentTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	if err := h.incidentService.DeleteIncidentTemplate(uint(id)); err != nil {
		h.logger.Error("Failed to delete incident template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete incident template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incident template deleted successfully"})
}
