// Package models defines incident template data models for the Incident Service.
package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// AdvancedIncidentTemplate represents a reusable template for creating incidents with automation
type AdvancedIncidentTemplate struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	TenantID    uint      `json:"tenant_id" gorm:"not null;index"`
	Name        string    `json:"name" gorm:"not null;size:255"`
	Description string    `json:"description" gorm:"type:text"`
	Title       string    `json:"title" gorm:"not null;size:500"`
	Body        string    `json:"body" gorm:"type:text"`
	Severity    string    `json:"severity" gorm:"not null;size:50;default:'medium'"` // low, medium, high, critical
	Status      string    `json:"status" gorm:"not null;size:50;default:'investigating'"` // investigating, identified, monitoring, resolved
	IsPublic    bool      `json:"is_public" gorm:"default:true"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedBy   uint      `json:"created_by" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Component associations
	ComponentIDs string `json:"component_ids" gorm:"type:text"` // JSON array of component IDs
	Components   []uint `json:"components" gorm:"-"`            // Computed field

	// Notification settings
	NotificationSettings string                 `json:"notification_settings" gorm:"type:text"` // JSON
	NotificationConfig   *NotificationConfig    `json:"notification_config" gorm:"-"`           // Computed field

	// Automation settings
	AutomationSettings string              `json:"automation_settings" gorm:"type:text"` // JSON
	AutomationConfig   *AutomationConfig   `json:"automation_config" gorm:"-"`           // Computed field

	// Template usage statistics
	UsageCount   int       `json:"usage_count" gorm:"default:0"`
	LastUsedAt   *time.Time `json:"last_used_at"`
	LastUsedBy   *uint     `json:"last_used_by"`

	// Relationships
	WorkflowSteps []WorkflowStep `json:"workflow_steps" gorm:"foreignKey:TemplateID"`
}

// WorkflowStep represents an automated step in an incident response workflow
type WorkflowStep struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	TemplateID     uint      `json:"template_id" gorm:"not null;index"`
	Name           string    `json:"name" gorm:"not null;size:255"`
	Description    string    `json:"description" gorm:"type:text"`
	StepType       string    `json:"step_type" gorm:"not null;size:50"` // notification, status_update, component_update, webhook, delay, manual
	Order          int       `json:"order" gorm:"not null"`
	TriggerType    string    `json:"trigger_type" gorm:"not null;size:50;default:'manual'"` // manual, automatic, time_based, status_based
	TriggerDelay   int       `json:"trigger_delay" gorm:"default:0"` // seconds
	TriggerStatus  string    `json:"trigger_status" gorm:"size:50"`  // Required status to trigger this step
	IsActive       bool      `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Step configuration (JSON)
	Configuration string        `json:"configuration" gorm:"type:text"`
	Config        *StepConfig   `json:"config" gorm:"-"` // Computed field

	// Execution tracking
	ExecutionLogs []StepExecution `json:"execution_logs" gorm:"foreignKey:StepID"`
}

// StepExecution tracks the execution of workflow steps
type StepExecution struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	StepID        uint      `json:"step_id" gorm:"not null;index"`
	IncidentID    uint      `json:"incident_id" gorm:"not null;index"`
	TenantID      uint      `json:"tenant_id" gorm:"not null;index"`
	Status        string    `json:"status" gorm:"not null;size:50"` // pending, running, completed, failed, skipped
	StartedAt     *time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	ExecutedBy    *uint     `json:"executed_by"` // User ID if manually executed
	ExecutionType string    `json:"execution_type" gorm:"not null;size:50"` // manual, automatic
	Result        string    `json:"result" gorm:"type:text"`        // Execution result/output
	ErrorMessage  string    `json:"error_message" gorm:"type:text"` // Error details if failed
	Metadata      string    `json:"metadata" gorm:"type:text"`      // Additional execution metadata (JSON)
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships
	Step     WorkflowStep `json:"step" gorm:"foreignKey:StepID"`
	Incident Incident     `json:"incident" gorm:"foreignKey:IncidentID"`
}

// NotificationConfig defines notification settings for templates
type NotificationConfig struct {
	Enabled           bool     `json:"enabled"`
	Channels          []string `json:"channels"`          // email, sms, slack, teams, webhook
	Recipients        []string `json:"recipients"`        // Email addresses, phone numbers, etc.
	NotifyOnCreation  bool     `json:"notify_on_creation"`
	NotifyOnUpdate    bool     `json:"notify_on_update"`
	NotifyOnResolved  bool     `json:"notify_on_resolved"`
	CustomMessage     string   `json:"custom_message"`
	EscalationRules   []EscalationRule `json:"escalation_rules"`
}

// EscalationRule defines escalation behavior
type EscalationRule struct {
	DelayMinutes int      `json:"delay_minutes"`
	Recipients   []string `json:"recipients"`
	Channels     []string `json:"channels"`
	Condition    string   `json:"condition"` // e.g., "not_resolved", "no_update"
}

// AutomationConfig defines automation settings for templates
type AutomationConfig struct {
	Enabled                bool                   `json:"enabled"`
	AutoAssignee           *uint                  `json:"auto_assignee"`           // User ID to auto-assign
	AutoComponentUpdate    bool                   `json:"auto_component_update"`   // Auto-update affected components
	AutoStatusProgression  bool                   `json:"auto_status_progression"` // Auto-progress through statuses
	StatusProgressionRules []StatusProgressionRule `json:"status_progression_rules"`
	AutoResolution         *AutoResolutionConfig  `json:"auto_resolution"`
	CustomWebhooks         []WebhookConfig        `json:"custom_webhooks"`
}

// StatusProgressionRule defines automatic status progression
type StatusProgressionRule struct {
	FromStatus   string `json:"from_status"`
	ToStatus     string `json:"to_status"`
	DelayMinutes int    `json:"delay_minutes"`
	Condition    string `json:"condition"` // e.g., "no_updates", "time_elapsed"
}

// AutoResolutionConfig defines automatic resolution settings
type AutoResolutionConfig struct {
	Enabled           bool   `json:"enabled"`
	DelayHours        int    `json:"delay_hours"`        // Hours after last update
	RequireConfirmation bool   `json:"require_confirmation"`
	ConfirmationUsers []uint `json:"confirmation_users"` // User IDs who can confirm
}

// WebhookConfig defines webhook automation
type WebhookConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Payload string            `json:"payload"` // Template with variables
	Trigger string            `json:"trigger"` // created, updated, resolved, etc.
}

// StepConfig defines configuration for different step types
type StepConfig struct {
	// Notification step config
	NotificationChannels []string `json:"notification_channels,omitempty"`
	NotificationMessage  string   `json:"notification_message,omitempty"`
	Recipients          []string `json:"recipients,omitempty"`

	// Status update step config
	NewStatus     string `json:"new_status,omitempty"`
	StatusMessage string `json:"status_message,omitempty"`

	// Component update step config
	ComponentIDs   []uint `json:"component_ids,omitempty"`
	ComponentStatus string `json:"component_status,omitempty"` // operational, degraded_performance, partial_outage, major_outage

	// Webhook step config
	WebhookURL     string            `json:"webhook_url,omitempty"`
	WebhookMethod  string            `json:"webhook_method,omitempty"`
	WebhookHeaders map[string]string `json:"webhook_headers,omitempty"`
	WebhookPayload string            `json:"webhook_payload,omitempty"`

	// Delay step config
	DelaySeconds int `json:"delay_seconds,omitempty"`

	// Manual step config
	Instructions     string   `json:"instructions,omitempty"`
	AssignedUsers    []uint   `json:"assigned_users,omitempty"`
	RequiredApproval bool     `json:"required_approval,omitempty"`
}

// Validation methods
func (t *IncidentTemplate) BeforeCreate(tx *gorm.DB) error {
	return t.validate()
}

func (t *IncidentTemplate) BeforeUpdate(tx *gorm.DB) error {
	return t.validate()
}

func (t *IncidentTemplate) validate() error {
	if t.Name == "" {
		return gorm.ErrInvalidValue
	}
	if t.Title == "" {
		return gorm.ErrInvalidValue
	}
	return nil
}

// Helper methods for JSON fields
func (t *AdvancedIncidentTemplate) AfterFind(tx *gorm.DB) error {
	// Parse component IDs
	if t.ComponentIDs != "" {
		json.Unmarshal([]byte(t.ComponentIDs), &t.Components)
	}

	// Parse notification settings
	if t.NotificationSettings != "" {
		json.Unmarshal([]byte(t.NotificationSettings), &t.NotificationConfig)
	}

	// Parse automation settings
	if t.AutomationSettings != "" {
		json.Unmarshal([]byte(t.AutomationSettings), &t.AutomationConfig)
	}

	return nil
}

func (t *AdvancedIncidentTemplate) BeforeSave(tx *gorm.DB) error {
	// Serialize component IDs
	if len(t.Components) > 0 {
		data, _ := json.Marshal(t.Components)
		t.ComponentIDs = string(data)
	}

	// Serialize notification config
	if t.NotificationConfig != nil {
		data, _ := json.Marshal(t.NotificationConfig)
		t.NotificationSettings = string(data)
	}

	// Serialize automation config
	if t.AutomationConfig != nil {
		data, _ := json.Marshal(t.AutomationConfig)
		t.AutomationSettings = string(data)
	}

	return nil
}

func (s *WorkflowStep) AfterFind(tx *gorm.DB) error {
	// Parse step configuration
	if s.Configuration != "" {
		json.Unmarshal([]byte(s.Configuration), &s.Config)
	}
	return nil
}

func (s *WorkflowStep) BeforeSave(tx *gorm.DB) error {
	// Serialize step config
	if s.Config != nil {
		data, _ := json.Marshal(s.Config)
		s.Configuration = string(data)
	}
	return nil
}

// TableName methods
func (AdvancedIncidentTemplate) TableName() string {
	return "advanced_incident_templates"
}

func (WorkflowStep) TableName() string {
	return "workflow_steps"
}

func (StepExecution) TableName() string {
	return "step_executions"
}

// Default templates
var DefaultAdvancedIncidentTemplates = []AdvancedIncidentTemplate{
	{
		Name:        "service-outage",
		Description: "Complete service outage affecting all users",
		Title:       "Service Outage - {{.ServiceName}}",
		Body:        "We are currently experiencing a complete outage of {{.ServiceName}}. Our team is investigating the issue and working to restore service as quickly as possible. We will provide updates as more information becomes available.",
		Severity:    "critical",
		Status:      "investigating",
		IsPublic:    true,
		NotificationConfig: &NotificationConfig{
			Enabled:          true,
			Channels:         []string{"email", "sms", "slack"},
			NotifyOnCreation: true,
			NotifyOnUpdate:   true,
			NotifyOnResolved: true,
		},
		AutomationConfig: &AutomationConfig{
			Enabled:             true,
			AutoComponentUpdate: true,
			StatusProgressionRules: []StatusProgressionRule{
				{FromStatus: "investigating", ToStatus: "identified", DelayMinutes: 15},
				{FromStatus: "identified", ToStatus: "monitoring", DelayMinutes: 30},
			},
		},
	},
	{
		Name:        "degraded-performance",
		Description: "Service experiencing degraded performance",
		Title:       "Degraded Performance - {{.ServiceName}}",
		Body:        "We are experiencing degraded performance with {{.ServiceName}}. Some users may notice slower response times. Our team is working to resolve this issue.",
		Severity:    "medium",
		Status:      "investigating",
		IsPublic:    true,
		NotificationConfig: &NotificationConfig{
			Enabled:          true,
			Channels:         []string{"email", "slack"},
			NotifyOnCreation: true,
			NotifyOnResolved: true,
		},
	},
	{
		Name:        "security-incident",
		Description: "Security-related incident template",
		Title:       "Security Incident - Investigation in Progress",
		Body:        "We are investigating a potential security incident. As a precautionary measure, we have implemented additional security measures. We will provide updates as our investigation progresses.",
		Severity:    "high",
		Status:      "investigating",
		IsPublic:    false, // Security incidents often start as private
		NotificationConfig: &NotificationConfig{
			Enabled:          true,
			Channels:         []string{"email"},
			NotifyOnCreation: false, // Manual control for security incidents
			NotifyOnUpdate:   false,
			NotifyOnResolved: true,
		},
	},
	{
		Name:        "maintenance-overrun",
		Description: "Scheduled maintenance taking longer than expected",
		Title:       "Maintenance Extended - {{.ServiceName}}",
		Body:        "Our scheduled maintenance for {{.ServiceName}} is taking longer than expected. We apologize for the inconvenience and will provide updates on our progress.",
		Severity:    "medium",
		Status:      "identified",
		IsPublic:    true,
		NotificationConfig: &NotificationConfig{
			Enabled:          true,
			Channels:         []string{"email", "slack"},
			NotifyOnCreation: true,
			NotifyOnUpdate:   true,
			NotifyOnResolved: true,
		},
	},
}

// Severity levels
const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

// Step types
const (
	StepTypeNotification    = "notification"
	StepTypeStatusUpdate    = "status_update"
	StepTypeComponentUpdate = "component_update"
	StepTypeWebhook         = "webhook"
	StepTypeDelay           = "delay"
	StepTypeManual          = "manual"
)

// Trigger types
const (
	TriggerTypeManual      = "manual"
	TriggerTypeAutomatic   = "automatic"
	TriggerTypeTimeBased   = "time_based"
	TriggerTypeStatusBased = "status_based"
)

// Execution statuses
const (
	ExecutionStatusPending   = "pending"
	ExecutionStatusRunning   = "running"
	ExecutionStatusCompleted = "completed"
	ExecutionStatusFailed    = "failed"
	ExecutionStatusSkipped   = "skipped"
)