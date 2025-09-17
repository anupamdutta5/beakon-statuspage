// Package services provides incident template and workflow automation business logic for the Incident Service.
package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/anupamdutta5/statuspage-incident-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IncidentTemplateService provides incident template and workflow management functionality
type IncidentTemplateService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewIncidentTemplateService creates a new incident template service instance
func NewIncidentTemplateService(db *gorm.DB, logger *zap.Logger) *IncidentTemplateService {
	return &IncidentTemplateService{
		db:     db,
		logger: logger,
	}
}

// Template Management

// CreateTemplate creates a new incident template
func (s *IncidentTemplateService) CreateTemplate(ctx context.Context, template *models.AdvancedIncidentTemplate) error {
	// Check if template name already exists for this tenant
	var existing models.AdvancedIncidentTemplate
	err := s.db.WithContext(ctx).Where("tenant_id = ? AND name = ?", template.TenantID, template.Name).First(&existing).Error
	if err == nil {
		return errors.New("template with this name already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check existing template: %w", err)
	}

	if err := s.db.WithContext(ctx).Create(template).Error; err != nil {
		s.logger.Error("Failed to create incident template", zap.Error(err))
		return fmt.Errorf("failed to create template: %w", err)
	}

	s.logger.Info("Incident template created successfully",
		zap.Uint("template_id", template.ID),
		zap.String("template_name", template.Name),
		zap.Uint("tenant_id", template.TenantID))

	return nil
}

// GetTemplates retrieves incident templates for a tenant
func (s *IncidentTemplateService) GetTemplates(ctx context.Context, tenantID uint, includeInactive bool) ([]models.AdvancedIncidentTemplate, error) {
	var templates []models.AdvancedIncidentTemplate
	query := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	err := query.Preload("WorkflowSteps", func(db *gorm.DB) *gorm.DB {
		return db.Order("workflow_steps.order ASC")
	}).Order("name ASC").Find(&templates).Error

	if err != nil {
		s.logger.Error("Failed to get incident templates", zap.Error(err))
		return nil, fmt.Errorf("failed to get templates: %w", err)
	}

	return templates, nil
}

// GetTemplate retrieves a specific incident template
func (s *IncidentTemplateService) GetTemplate(ctx context.Context, templateID, tenantID uint) (*models.IncidentTemplate, error) {
	var template models.IncidentTemplate
	err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", templateID, tenantID).
		Preload("WorkflowSteps", func(db *gorm.DB) *gorm.DB {
			return db.Order("workflow_steps.order ASC")
		}).First(&template).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("template not found")
		}
		s.logger.Error("Failed to get incident template", zap.Error(err))
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return &template, nil
}

// UpdateTemplate updates an existing incident template
func (s *IncidentTemplateService) UpdateTemplate(ctx context.Context, templateID, tenantID uint, updates map[string]interface{}) error {
	result := s.db.WithContext(ctx).Model(&models.AdvancedIncidentTemplate{}).
		Where("id = ? AND tenant_id = ?", templateID, tenantID).Updates(updates)

	if result.Error != nil {
		s.logger.Error("Failed to update incident template", zap.Error(result.Error))
		return fmt.Errorf("failed to update template: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("template not found")
	}

	s.logger.Info("Incident template updated successfully",
		zap.Uint("template_id", templateID),
		zap.Uint("tenant_id", tenantID))

	return nil
}

// DeleteTemplate soft deletes an incident template
func (s *IncidentTemplateService) DeleteTemplate(ctx context.Context, templateID, tenantID uint) error {
	result := s.db.WithContext(ctx).Model(&models.IncidentTemplate{}).
		Where("id = ? AND tenant_id = ?", templateID, tenantID).Update("is_active", false)

	if result.Error != nil {
		s.logger.Error("Failed to delete incident template", zap.Error(result.Error))
		return fmt.Errorf("failed to delete template: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("template not found")
	}

	s.logger.Info("Incident template deleted successfully",
		zap.Uint("template_id", templateID),
		zap.Uint("tenant_id", tenantID))

	return nil
}

// CloneTemplate creates a copy of an existing template
func (s *IncidentTemplateService) CloneTemplate(ctx context.Context, templateID, tenantID uint, newName string) (*models.IncidentTemplate, error) {
	// Get the original template
	original, err := s.GetTemplate(ctx, templateID, tenantID)
	if err != nil {
		return nil, err
	}

	// Create a clone
	clone := *original
	clone.ID = 0 // Reset ID for new record
	clone.Name = newName
	// clone.UsageCount // TODO: Add UsageCount field to model = 0
	// clone.LastUsedAt // TODO: Add LastUsedAt field to model = nil
	clone.LastUsedBy = nil
	clone.CreatedAt = time.Time{}
	clone.UpdatedAt = time.Time{}

	// Clone workflow steps
	originalSteps := clone.WorkflowSteps
	clone.WorkflowSteps = nil

	// Create the cloned template
	if err := s.CreateTemplate(ctx, &clone); err != nil {
		return nil, err
	}

	// Clone workflow steps
	for _, step := range originalSteps {
		stepClone := step
		stepClone.ID = 0
		stepClone.TemplateID = clone.ID
		stepClone.CreatedAt = time.Time{}
		stepClone.UpdatedAt = time.Time{}

		if err := s.CreateWorkflowStep(ctx, &stepClone); err != nil {
			s.logger.Warn("Failed to clone workflow step", zap.Error(err))
		}
	}

	return &clone, nil
}

// Workflow Step Management

// CreateWorkflowStep creates a new workflow step
func (s *IncidentTemplateService) CreateWorkflowStep(ctx context.Context, step *models.WorkflowStep) error {
	if err := s.db.WithContext(ctx).Create(step).Error; err != nil {
		s.logger.Error("Failed to create workflow step", zap.Error(err))
		return fmt.Errorf("failed to create workflow step: %w", err)
	}

	s.logger.Info("Workflow step created successfully",
		zap.Uint("step_id", step.ID),
		zap.Uint("template_id", step.TemplateID))

	return nil
}

// UpdateWorkflowStep updates an existing workflow step
func (s *IncidentTemplateService) UpdateWorkflowStep(ctx context.Context, stepID uint, updates map[string]interface{}) error {
	result := s.db.WithContext(ctx).Model(&models.WorkflowStep{}).
		Where("id = ?", stepID).Updates(updates)

	if result.Error != nil {
		s.logger.Error("Failed to update workflow step", zap.Error(result.Error))
		return fmt.Errorf("failed to update workflow step: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("workflow step not found")
	}

	return nil
}

// DeleteWorkflowStep deletes a workflow step
func (s *IncidentTemplateService) DeleteWorkflowStep(ctx context.Context, stepID uint) error {
	result := s.db.WithContext(ctx).Delete(&models.WorkflowStep{}, stepID)

	if result.Error != nil {
		s.logger.Error("Failed to delete workflow step", zap.Error(result.Error))
		return fmt.Errorf("failed to delete workflow step: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("workflow step not found")
	}

	return nil
}

// ReorderWorkflowSteps updates the order of workflow steps
func (s *IncidentTemplateService) ReorderWorkflowSteps(ctx context.Context, templateID uint, stepOrder []uint) error {
	tx := s.db.WithContext(ctx).Begin()

	for i, stepID := range stepOrder {
		if err := tx.Model(&models.WorkflowStep{}).
			Where("id = ? AND template_id = ?", stepID, templateID).
			Update("order", i+1).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to reorder step %d: %w", stepID, err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit reorder transaction: %w", err)
	}

	return nil
}

// Incident Creation from Template

// CreateIncidentFromTemplate creates a new incident using a template
func (s *IncidentTemplateService) CreateIncidentFromTemplate(ctx context.Context, templateID, tenantID, createdBy uint, variables map[string]string) (*models.Incident, error) {
	// Get the template
	template, err := s.GetTemplate(ctx, templateID, tenantID)
	if err != nil {
		return nil, err
	}

	// Replace variables in title and body
	title := s.replaceVariables(template.Title, variables)
	body := s.replaceVariables(template.Body, variables)

	// Create incident
	incident := &models.Incident{
		TenantID:    tenantID,
		Title:       title,
		Body:        body,
		Status:      template.Status,
		Severity:    template.Severity,
		IsPublic:    template.IsPublic,
		CreatedBy:   createdBy,
		ResolvedAt:  nil,
	}

	// Set component associations if specified
	if len(template.Components) > 0 {
		componentsJSON, _ := json.Marshal(template.Components)
		incident.ComponentIDs = string(componentsJSON)
	}

	// Create the incident
	if err := s.db.WithContext(ctx).Create(incident).Error; err != nil {
		s.logger.Error("Failed to create incident from template", zap.Error(err))
		return nil, fmt.Errorf("failed to create incident: %w", err)
	}

	// Update template usage statistics
	now := time.Now()
	s.db.WithContext(ctx).Model(template).Updates(map[string]interface{}{
		"usage_count":  template.UsageCount + 1,
		"last_used_at": &now,
		"last_used_by": &createdBy,
	})

	// Initialize workflow execution if template has workflow steps
	if len(template.WorkflowSteps) > 0 {
		go s.initializeWorkflowExecution(context.Background(), incident.ID, template.WorkflowSteps)
	}

	s.logger.Info("Incident created from template",
		zap.Uint("incident_id", incident.ID),
		zap.Uint("template_id", templateID),
		zap.String("title", incident.Title))

	return incident, nil
}

// Workflow Execution

// initializeWorkflowExecution sets up workflow step executions for an incident
func (s *IncidentTemplateService) initializeWorkflowExecution(ctx context.Context, incidentID uint, steps []models.WorkflowStep) {
	for _, step := range steps {
		execution := &models.StepExecution{
			StepID:        step.ID,
			IncidentID:    incidentID,
			TenantID:      0, // Will be set from incident
			Status:        models.ExecutionStatusPending,
			ExecutionType: models.TriggerTypeAutomatic,
		}

		if step.TriggerType == models.TriggerTypeAutomatic {
			execution.Status = models.ExecutionStatusRunning
			now := time.Now()
			execution.StartedAt = &now

			// Execute the step
			go s.executeWorkflowStep(context.Background(), execution, &step)
		}

		if err := s.db.WithContext(ctx).Create(execution).Error; err != nil {
			s.logger.Error("Failed to create step execution", zap.Error(err))
		}
	}
}

// executeWorkflowStep executes a specific workflow step
func (s *IncidentTemplateService) executeWorkflowStep(ctx context.Context, execution *models.StepExecution, step *models.WorkflowStep) {
	s.logger.Info("Executing workflow step",
		zap.Uint("step_id", step.ID),
		zap.String("step_type", step.StepType),
		zap.Uint("incident_id", execution.IncidentID))

	var result string
	var err error

	// Add delay if specified
	if step.TriggerDelay > 0 {
		time.Sleep(time.Duration(step.TriggerDelay) * time.Second)
	}

	switch step.StepType {
	case models.StepTypeNotification:
		result, err = s.executeNotificationStep(ctx, step, execution.IncidentID)
	case models.StepTypeStatusUpdate:
		result, err = s.executeStatusUpdateStep(ctx, step, execution.IncidentID)
	case models.StepTypeComponentUpdate:
		result, err = s.executeComponentUpdateStep(ctx, step, execution.IncidentID)
	case models.StepTypeWebhook:
		result, err = s.executeWebhookStep(ctx, step, execution.IncidentID)
	case models.StepTypeDelay:
		result, err = s.executeDelayStep(ctx, step)
	case models.StepTypeManual:
		result = "Manual step pending user action"
		// Manual steps require user intervention
		s.db.WithContext(ctx).Model(execution).Updates(map[string]interface{}{
			"status": models.ExecutionStatusPending,
			"result": result,
		})
		return
	default:
		err = fmt.Errorf("unknown step type: %s", step.StepType)
	}

	// Update execution status
	now := time.Now()
	updates := map[string]interface{}{
		"completed_at": &now,
		"result":       result,
	}

	if err != nil {
		updates["status"] = models.ExecutionStatusFailed
		updates["error_message"] = err.Error()
		s.logger.Error("Workflow step execution failed",
			zap.Error(err),
			zap.Uint("step_id", step.ID))
	} else {
		updates["status"] = models.ExecutionStatusCompleted
		s.logger.Info("Workflow step execution completed",
			zap.Uint("step_id", step.ID),
			zap.String("result", result))
	}

	s.db.WithContext(ctx).Model(execution).Updates(updates)
}

// Step execution methods
func (s *IncidentTemplateService) executeNotificationStep(ctx context.Context, step *models.WorkflowStep, incidentID uint) (string, error) {
	// This would integrate with the notification service
	// For now, return a placeholder result
	return "Notification sent to configured channels", nil
}

func (s *IncidentTemplateService) executeStatusUpdateStep(ctx context.Context, step *models.WorkflowStep, incidentID uint) (string, error) {
	if step.Config == nil || step.Config.NewStatus == "" {
		return "", errors.New("new status not specified")
	}

	// Update incident status
	result := s.db.WithContext(ctx).Model(&models.Incident{}).
		Where("id = ?", incidentID).
		Updates(map[string]interface{}{
			"status": step.Config.NewStatus,
		})

	if result.Error != nil {
		return "", result.Error
	}

	return fmt.Sprintf("Status updated to %s", step.Config.NewStatus), nil
}

func (s *IncidentTemplateService) executeComponentUpdateStep(ctx context.Context, step *models.WorkflowStep, incidentID uint) (string, error) {
	if step.Config == nil || len(step.Config.ComponentIDs) == 0 {
		return "", errors.New("component IDs not specified")
	}

	// This would integrate with the component service
	// For now, return a placeholder result
	return fmt.Sprintf("Updated %d components to status %s", len(step.Config.ComponentIDs), step.Config.ComponentStatus), nil
}

func (s *IncidentTemplateService) executeWebhookStep(ctx context.Context, step *models.WorkflowStep, incidentID uint) (string, error) {
	if step.Config == nil || step.Config.WebhookURL == "" {
		return "", errors.New("webhook URL not specified")
	}

	// This would make an HTTP request to the webhook URL
	// For now, return a placeholder result
	return fmt.Sprintf("Webhook called: %s", step.Config.WebhookURL), nil
}

func (s *IncidentTemplateService) executeDelayStep(ctx context.Context, step *models.WorkflowStep) (string, error) {
	if step.Config == nil || step.Config.DelaySeconds == 0 {
		return "", errors.New("delay duration not specified")
	}

	time.Sleep(time.Duration(step.Config.DelaySeconds) * time.Second)
	return fmt.Sprintf("Delayed execution for %d seconds", step.Config.DelaySeconds), nil
}

// GetStepExecutions retrieves execution history for workflow steps
func (s *IncidentTemplateService) GetStepExecutions(ctx context.Context, incidentID uint, tenantID uint) ([]models.StepExecution, error) {
	var executions []models.StepExecution
	err := s.db.WithContext(ctx).Where("incident_id = ? AND tenant_id = ?", incidentID, tenantID).
		Preload("Step").Order("created_at ASC").Find(&executions).Error

	if err != nil {
		s.logger.Error("Failed to get step executions", zap.Error(err))
		return nil, fmt.Errorf("failed to get step executions: %w", err)
	}

	return executions, nil
}

// ExecuteManualStep manually executes a pending manual workflow step
func (s *IncidentTemplateService) ExecuteManualStep(ctx context.Context, executionID uint, executedBy uint) error {
	var execution models.StepExecution
	err := s.db.WithContext(ctx).Preload("Step").First(&execution, executionID).Error
	if err != nil {
		return fmt.Errorf("execution not found: %w", err)
	}

	if execution.Status != models.ExecutionStatusPending {
		return errors.New("step is not in pending status")
	}

	// Update execution
	now := time.Now()
	updates := map[string]interface{}{
		"status":         models.ExecutionStatusCompleted,
		"executed_by":    &executedBy,
		"execution_type": "manual",
		"started_at":     &now,
		"completed_at":   &now,
		"result":         "Manual step completed by user",
	}

	return s.db.WithContext(ctx).Model(&execution).Updates(updates).Error
}

// Initialize default templates for a tenant
func (s *IncidentTemplateService) InitializeDefaultTemplates(ctx context.Context, tenantID, createdBy uint) error {
	for _, templateData := range models.DefaultIncidentTemplates {
		template := templateData
		template.TenantID = tenantID
		template.CreatedBy = createdBy

		var existing models.IncidentTemplate
		err := s.db.WithContext(ctx).Where("tenant_id = ? AND name = ?", tenantID, template.Name).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := s.CreateTemplate(ctx, &template); err != nil {
				s.logger.Error("Failed to create default template", zap.Error(err), zap.String("template", template.Name))
			}
		}
	}

	s.logger.Info("Default incident templates initialized for tenant", zap.Uint("tenant_id", tenantID))
	return nil
}

// Helper methods
func (s *IncidentTemplateService) replaceVariables(text string, variables map[string]string) string {
	result := text
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}