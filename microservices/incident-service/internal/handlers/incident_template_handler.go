// Package handlers provides incident template HTTP handlers for the Incident Service.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/anupamdutta5/incident-service/internal/models"
	"github.com/anupamdutta5/incident-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// IncidentTemplateHandler handles incident template-related HTTP requests
type IncidentTemplateHandler struct {
	templateService *services.IncidentTemplateService
	logger          *zap.Logger
}

// NewIncidentTemplateHandler creates a new incident template handler
func NewIncidentTemplateHandler(templateService *services.IncidentTemplateService, logger *zap.Logger) *IncidentTemplateHandler {
	return &IncidentTemplateHandler{
		templateService: templateService,
		logger:          logger,
	}
}

// Template Management Handlers

// CreateTemplate handles creating a new incident template
func (h *IncidentTemplateHandler) CreateTemplate(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var req struct {
		Name                 string                      `json:"name" binding:"required"`
		Description          string                      `json:"description"`
		Title                string                      `json:"title" binding:"required"`
		Body                 string                      `json:"body" binding:"required"`
		Severity             string                      `json:"severity" binding:"required"`
		Status               string                      `json:"status" binding:"required"`
		IsPublic             *bool                       `json:"is_public"`
		IsActive             *bool                       `json:"is_active"`
		Components           []uint                      `json:"components"`
		NotificationConfig   *models.NotificationConfig  `json:"notification_config"`
		AutomationConfig     *models.AutomationConfig    `json:"automation_config"`
		WorkflowSteps        []WorkflowStepRequest       `json:"workflow_steps"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create template request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	template := &models.AdvancedIncidentTemplate{
		TenantID:           tenantID.(uint),
		Name:               req.Name,
		Description:        req.Description,
		Title:              req.Title,
		Body:               req.Body,
		Severity:           req.Severity,
		Status:             req.Status,
		IsPublic:           isPublic,
		IsActive:           isActive,
		CreatedBy:          userID.(uint),
		Components:         req.Components,
		NotificationConfig: req.NotificationConfig,
		AutomationConfig:   req.AutomationConfig,
	}

	if err := h.templateService.CreateTemplate(c.Request.Context(), template); err != nil {
		h.logger.Error("Failed to create template", zap.Error(err))
		if err.Error() == "template with this name already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "Template with this name already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		}
		return
	}

	// Create workflow steps
	for i, stepReq := range req.WorkflowSteps {
		step := &models.WorkflowStep{
			TemplateID:    template.ID,
			Name:          stepReq.Name,
			Description:   stepReq.Description,
			StepType:      stepReq.StepType,
			Order:         i + 1,
			TriggerType:   stepReq.TriggerType,
			TriggerDelay:  stepReq.TriggerDelay,
			TriggerStatus: stepReq.TriggerStatus,
			IsActive:      stepReq.IsActive,
			Config:        stepReq.Config,
		}

		if err := h.templateService.CreateWorkflowStep(c.Request.Context(), step); err != nil {
			h.logger.Warn("Failed to create workflow step", zap.Error(err))
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Template created successfully",
		"template": template,
	})
}

// GetTemplates handles retrieving incident templates
func (h *IncidentTemplateHandler) GetTemplates(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	includeInactive := c.Query("include_inactive") == "true"

	templates, err := h.templateService.GetTemplates(c.Request.Context(), tenantID.(uint), includeInactive)
	if err != nil {
		h.logger.Error("Failed to get templates", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"count":     len(templates),
	})
}

// GetTemplate handles retrieving a specific incident template
func (h *IncidentTemplateHandler) GetTemplate(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	templateIDStr := c.Param("id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	template, err := h.templateService.GetTemplate(c.Request.Context(), uint(templateID), tenantID.(uint))
	if err != nil {
		h.logger.Error("Failed to get template", zap.Error(err))
		if err.Error() == "template not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve template"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"template": template})
}

// UpdateTemplate handles updating an incident template
func (h *IncidentTemplateHandler) UpdateTemplate(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	templateIDStr := c.Param("id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var req struct {
		Name               *string                     `json:"name"`
		Description        *string                     `json:"description"`
		Title              *string                     `json:"title"`
		Body               *string                     `json:"body"`
		Severity           *string                     `json:"severity"`
		Status             *string                     `json:"status"`
		IsPublic           *bool                       `json:"is_public"`
		IsActive           *bool                       `json:"is_active"`
		Components         *[]uint                     `json:"components"`
		NotificationConfig *models.NotificationConfig  `json:"notification_config"`
		AutomationConfig   *models.AutomationConfig    `json:"automation_config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update template request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Body != nil {
		updates["body"] = *req.Body
	}
	if req.Severity != nil {
		updates["severity"] = *req.Severity
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.Components != nil {
		updates["components"] = *req.Components
	}
	if req.NotificationConfig != nil {
		updates["notification_config"] = *req.NotificationConfig
	}
	if req.AutomationConfig != nil {
		updates["automation_config"] = *req.AutomationConfig
	}

	if err := h.templateService.UpdateTemplate(c.Request.Context(), uint(templateID), tenantID.(uint), updates); err != nil {
		h.logger.Error("Failed to update template", zap.Error(err))
		if err.Error() == "template not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template updated successfully"})
}

// DeleteTemplate handles deleting an incident template
func (h *IncidentTemplateHandler) DeleteTemplate(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	templateIDStr := c.Param("id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	if err := h.templateService.DeleteTemplate(c.Request.Context(), uint(templateID), tenantID.(uint)); err != nil {
		h.logger.Error("Failed to delete template", zap.Error(err))
		if err.Error() == "template not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully"})
}

// CloneTemplate handles cloning an incident template
func (h *IncidentTemplateHandler) CloneTemplate(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	templateIDStr := c.Param("id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid clone template request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	clonedTemplate, err := h.templateService.CloneTemplate(c.Request.Context(), uint(templateID), tenantID.(uint), req.Name)
	if err != nil {
		h.logger.Error("Failed to clone template", zap.Error(err))
		if err.Error() == "template not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clone template"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Template cloned successfully",
		"template": clonedTemplate,
	})
}

// Incident Creation Handlers

// CreateIncidentFromTemplate handles creating an incident from a template
func (h *IncidentTemplateHandler) CreateIncidentFromTemplate(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	templateIDStr := c.Param("id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var req struct {
		Variables map[string]string `json:"variables"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create incident request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if req.Variables == nil {
		req.Variables = make(map[string]string)
	}

	incident, err := h.templateService.CreateIncidentFromTemplate(
		c.Request.Context(),
		uint(templateID),
		tenantID.(uint),
		userID.(uint),
		req.Variables,
	)

	if err != nil {
		h.logger.Error("Failed to create incident from template", zap.Error(err))
		if err.Error() == "template not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create incident"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Incident created successfully",
		"incident": incident,
	})
}

// Workflow Management Handlers

// CreateWorkflowStep handles creating a workflow step
func (h *IncidentTemplateHandler) CreateWorkflowStep(c *gin.Context) {
	templateIDStr := c.Param("template_id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var req WorkflowStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create workflow step request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	step := &models.WorkflowStep{
		TemplateID:    uint(templateID),
		Name:          req.Name,
		Description:   req.Description,
		StepType:      req.StepType,
		Order:         req.Order,
		TriggerType:   req.TriggerType,
		TriggerDelay:  req.TriggerDelay,
		TriggerStatus: req.TriggerStatus,
		IsActive:      req.IsActive,
		Config:        req.Config,
	}

	if err := h.templateService.CreateWorkflowStep(c.Request.Context(), step); err != nil {
		h.logger.Error("Failed to create workflow step", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workflow step"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Workflow step created successfully",
		"step":    step,
	})
}

// UpdateWorkflowStep handles updating a workflow step
func (h *IncidentTemplateHandler) UpdateWorkflowStep(c *gin.Context) {
	stepIDStr := c.Param("step_id")
	stepID, err := strconv.ParseUint(stepIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid step ID"})
		return
	}

	var req struct {
		Name          *string             `json:"name"`
		Description   *string             `json:"description"`
		StepType      *string             `json:"step_type"`
		Order         *int                `json:"order"`
		TriggerType   *string             `json:"trigger_type"`
		TriggerDelay  *int                `json:"trigger_delay"`
		TriggerStatus *string             `json:"trigger_status"`
		IsActive      *bool               `json:"is_active"`
		Config        *models.StepConfig  `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update workflow step request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.StepType != nil {
		updates["step_type"] = *req.StepType
	}
	if req.Order != nil {
		updates["order"] = *req.Order
	}
	if req.TriggerType != nil {
		updates["trigger_type"] = *req.TriggerType
	}
	if req.TriggerDelay != nil {
		updates["trigger_delay"] = *req.TriggerDelay
	}
	if req.TriggerStatus != nil {
		updates["trigger_status"] = *req.TriggerStatus
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.Config != nil {
		updates["config"] = *req.Config
	}

	if err := h.templateService.UpdateWorkflowStep(c.Request.Context(), uint(stepID), updates); err != nil {
		h.logger.Error("Failed to update workflow step", zap.Error(err))
		if err.Error() == "workflow step not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workflow step not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update workflow step"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workflow step updated successfully"})
}

// DeleteWorkflowStep handles deleting a workflow step
func (h *IncidentTemplateHandler) DeleteWorkflowStep(c *gin.Context) {
	stepIDStr := c.Param("step_id")
	stepID, err := strconv.ParseUint(stepIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid step ID"})
		return
	}

	if err := h.templateService.DeleteWorkflowStep(c.Request.Context(), uint(stepID)); err != nil {
		h.logger.Error("Failed to delete workflow step", zap.Error(err))
		if err.Error() == "workflow step not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workflow step not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete workflow step"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workflow step deleted successfully"})
}

// ReorderWorkflowSteps handles reordering workflow steps
func (h *IncidentTemplateHandler) ReorderWorkflowSteps(c *gin.Context) {
	templateIDStr := c.Param("template_id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var req struct {
		StepOrder []uint `json:"step_order" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid reorder steps request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err := h.templateService.ReorderWorkflowSteps(c.Request.Context(), uint(templateID), req.StepOrder); err != nil {
		h.logger.Error("Failed to reorder workflow steps", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reorder workflow steps"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workflow steps reordered successfully"})
}

// Workflow Execution Handlers

// GetStepExecutions handles retrieving workflow execution history
func (h *IncidentTemplateHandler) GetStepExecutions(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	incidentIDStr := c.Param("incident_id")
	incidentID, err := strconv.ParseUint(incidentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	executions, err := h.templateService.GetStepExecutions(c.Request.Context(), uint(incidentID), tenantID.(uint))
	if err != nil {
		h.logger.Error("Failed to get step executions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve step executions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"executions": executions,
		"count":      len(executions),
	})
}

// ExecuteManualStep handles manual execution of workflow steps
func (h *IncidentTemplateHandler) ExecuteManualStep(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	executionIDStr := c.Param("execution_id")
	executionID, err := strconv.ParseUint(executionIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
		return
	}

	if err := h.templateService.ExecuteManualStep(c.Request.Context(), uint(executionID), userID.(uint)); err != nil {
		h.logger.Error("Failed to execute manual step", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute manual step"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Manual step executed successfully"})
}

// InitializeDefaultTemplates handles initializing default templates for a tenant
func (h *IncidentTemplateHandler) InitializeDefaultTemplates(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	if err := h.templateService.InitializeDefaultTemplates(c.Request.Context(), tenantID.(uint), userID.(uint)); err != nil {
		h.logger.Error("Failed to initialize default templates", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize default templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default templates initialized successfully"})
}

// Request structs
type WorkflowStepRequest struct {
	Name          string             `json:"name" binding:"required"`
	Description   string             `json:"description"`
	StepType      string             `json:"step_type" binding:"required"`
	Order         int                `json:"order"`
	TriggerType   string             `json:"trigger_type"`
	TriggerDelay  int                `json:"trigger_delay"`
	TriggerStatus string             `json:"trigger_status"`
	IsActive      bool               `json:"is_active"`
	Config        *models.StepConfig `json:"config"`
}