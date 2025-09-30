// Package services provides business logic for feature flag management.
package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/anupamdutta5/saas-admin-service/internal/models"
)

// FeatureFlagService handles feature flag operations.
type FeatureFlagService struct {
	db *gorm.DB
}

// FeatureFlagConfig represents feature flag configuration.
type FeatureFlagConfig struct {
	Enabled          bool                   `json:"enabled"`
	Rules            []FeatureFlagRule      `json:"rules,omitempty"`
	Rollout          *RolloutConfig         `json:"rollout,omitempty"`
	Dependencies     []string               `json:"dependencies,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	ExpiresAt        *time.Time             `json:"expires_at,omitempty"`
	DefaultValue     interface{}            `json:"default_value,omitempty"`
	AllowedValues    []interface{}          `json:"allowed_values,omitempty"`
	ValidationRules  []ValidationRule       `json:"validation_rules,omitempty"`
	AuditLog         bool                   `json:"audit_log,omitempty"`
	NotificationHook string                 `json:"notification_hook,omitempty"`
}

// FeatureFlagRule represents a conditional rule for feature flags.
type FeatureFlagRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Conditions  []Condition            `json:"conditions"`
	Action      string                 `json:"action"` // enable, disable, set_value
	Value       interface{}            `json:"value,omitempty"`
	Priority    int                    `json:"priority"`
	Enabled     bool                   `json:"enabled"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Condition represents a condition in a feature flag rule.
type Condition struct {
	Field    string      `json:"field"`    // tenant_id, plan, user_id, ip_address, etc.
	Operator string      `json:"operator"` // equals, not_equals, in, not_in, contains, starts_with, etc.
	Value    interface{} `json:"value"`
}

// RolloutConfig represents rollout configuration for gradual feature deployment.
type RolloutConfig struct {
	Type       string  `json:"type"`       // percentage, user_list, plan_based
	Percentage float64 `json:"percentage"` // 0-100
	UserList   []string `json:"user_list,omitempty"`
	PlanList   []string `json:"plan_list,omitempty"`
	StartTime  *time.Time `json:"start_time,omitempty"`
	EndTime    *time.Time `json:"end_time,omitempty"`
}

// ValidationRule represents validation rules for feature flag values.
type ValidationRule struct {
	Type     string      `json:"type"` // required, min, max, regex, custom
	Value    interface{} `json:"value,omitempty"`
	Message  string      `json:"message,omitempty"`
	Function string      `json:"function,omitempty"` // For custom validations
}

// EvaluationContext represents context for feature flag evaluation.
type EvaluationContext struct {
	TenantID    string                 `json:"tenant_id"`
	UserID      string                 `json:"user_id,omitempty"`
	Plan        string                 `json:"plan,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Environment string                 `json:"environment,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Custom      map[string]interface{} `json:"custom,omitempty"`
}

// NewFeatureFlagService creates a new feature flag service.
func NewFeatureFlagService(db *gorm.DB) *FeatureFlagService {
	return &FeatureFlagService{
		db: db,
	}
}

// InitializeDefaultFlags creates default feature flags.
func (s *FeatureFlagService) InitializeDefaultFlags() error {
	defaultFlags := []models.SaaSFeatureFlag{
		{
			Name:        "custom_domains",
			Description: "Enable custom domain functionality for paid plans",
			IsEnabled:   true,
			Config:      s.getCustomDomainFlagConfig(),
		},
		{
			Name:        "ssl_certificates",
			Description: "Enable SSL certificate management for custom domains",
			IsEnabled:   true,
			Config:      s.getSSLCertificateFlagConfig(),
		},
		{
			Name:        "domain_verification",
			Description: "Enable domain ownership verification",
			IsEnabled:   true,
			Config:      `{"enabled": true, "verification_timeout": 3600}`,
		},
		{
			Name:        "advanced_analytics",
			Description: "Enable advanced analytics features",
			IsEnabled:   false,
			Config:      `{"enabled": false, "rollout": {"type": "percentage", "percentage": 0}}`,
		},
		{
			Name:        "api_access",
			Description: "Enable API access for tenants",
			IsEnabled:   true,
			Config:      `{"enabled": true, "rate_limits": {"requests_per_minute": 60}}`,
		},
	}

	for _, flag := range defaultFlags {
		// Check if flag already exists
		var existing models.SaaSFeatureFlag
		if err := s.db.Where("name = ?", flag.Name).First(&existing).Error; err == nil {
			continue // Skip if already exists
		}

		if err := s.db.Create(&flag).Error; err != nil {
			return fmt.Errorf("failed to create default flag %s: %w", flag.Name, err)
		}
	}

	return nil
}

// IsFeatureEnabled checks if a feature is enabled for the given context.
func (s *FeatureFlagService) IsFeatureEnabled(flagName string, context *EvaluationContext) (bool, error) {
	result, err := s.EvaluateFeatureFlag(flagName, context)
	if err != nil {
		return false, err
	}

	enabled, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("feature flag %s does not return boolean value", flagName)
	}

	return enabled, nil
}

// EvaluateFeatureFlag evaluates a feature flag and returns its value.
func (s *FeatureFlagService) EvaluateFeatureFlag(flagName string, context *EvaluationContext) (interface{}, error) {
	var flag models.SaaSFeatureFlag
	if err := s.db.Where("name = ?", flagName).First(&flag).Error; err != nil {
		return nil, fmt.Errorf("feature flag not found: %s", flagName)
	}

	// If flag is globally disabled, return false
	if !flag.IsEnabled {
		return false, nil
	}

	// Parse flag configuration
	var config FeatureFlagConfig
	if err := json.Unmarshal([]byte(flag.Config), &config); err != nil {
		return config.Enabled, nil // Fallback to simple enabled/disabled
	}

	// Check if flag has expired
	if config.ExpiresAt != nil && time.Now().After(*config.ExpiresAt) {
		return config.DefaultValue, nil
	}

	// Evaluate rules in priority order
	for _, rule := range config.Rules {
		if !rule.Enabled {
			continue
		}

		if s.evaluateRule(rule, context) {
			switch rule.Action {
			case "enable":
				return true, nil
			case "disable":
				return false, nil
			case "set_value":
				return rule.Value, nil
			}
		}
	}

	// Check rollout configuration
	if config.Rollout != nil {
		if s.evaluateRollout(*config.Rollout, context) {
			return config.Enabled, nil
		}
		return config.DefaultValue, nil
	}

	return config.Enabled, nil
}

// HasCustomDomainAccess checks if a tenant has access to custom domains.
func (s *FeatureFlagService) HasCustomDomainAccess(tenantID, plan string) (bool, error) {
	context := &EvaluationContext{
		TenantID:  tenantID,
		Plan:      plan,
		Timestamp: time.Now(),
	}

	return s.IsFeatureEnabled("custom_domains", context)
}

// GetFeatureFlags retrieves all feature flags with optional filtering.
func (s *FeatureFlagService) GetFeatureFlags(enabled *bool) ([]models.SaaSFeatureFlag, error) {
	var flags []models.SaaSFeatureFlag
	query := s.db

	if enabled != nil {
		query = query.Where("is_enabled = ?", *enabled)
	}

	if err := query.Find(&flags).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve feature flags: %w", err)
	}

	return flags, nil
}

// UpdateFeatureFlag updates an existing feature flag.
func (s *FeatureFlagService) UpdateFeatureFlag(name string, updates map[string]interface{}) error {
	var flag models.SaaSFeatureFlag
	if err := s.db.Where("name = ?", name).First(&flag).Error; err != nil {
		return fmt.Errorf("feature flag not found: %s", name)
	}

	if err := s.db.Model(&flag).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update feature flag: %w", err)
	}

	return nil
}

// CreateFeatureFlag creates a new feature flag.
func (s *FeatureFlagService) CreateFeatureFlag(flag *models.SaaSFeatureFlag) error {
	// Check if flag already exists
	var existing models.SaaSFeatureFlag
	if err := s.db.Where("name = ?", flag.Name).First(&existing).Error; err == nil {
		return fmt.Errorf("feature flag already exists: %s", flag.Name)
	}

	if err := s.db.Create(flag).Error; err != nil {
		return fmt.Errorf("failed to create feature flag: %w", err)
	}

	return nil
}

// DeleteFeatureFlag deletes a feature flag.
func (s *FeatureFlagService) DeleteFeatureFlag(name string) error {
	var flag models.SaaSFeatureFlag
	if err := s.db.Where("name = ?", name).First(&flag).Error; err != nil {
		return fmt.Errorf("feature flag not found: %s", name)
	}

	if err := s.db.Delete(&flag).Error; err != nil {
		return fmt.Errorf("failed to delete feature flag: %w", err)
	}

	return nil
}

// Internal helper methods

func (s *FeatureFlagService) evaluateRule(rule FeatureFlagRule, context *EvaluationContext) bool {
	for _, condition := range rule.Conditions {
		if !s.evaluateCondition(condition, context) {
			return false // All conditions must be true
		}
	}
	return true
}

func (s *FeatureFlagService) evaluateCondition(condition Condition, context *EvaluationContext) bool {
	var fieldValue interface{}

	// Extract field value from context
	switch condition.Field {
	case "tenant_id":
		fieldValue = context.TenantID
	case "user_id":
		fieldValue = context.UserID
	case "plan":
		fieldValue = context.Plan
	case "ip_address":
		fieldValue = context.IPAddress
	case "user_agent":
		fieldValue = context.UserAgent
	case "environment":
		fieldValue = context.Environment
	default:
		// Check custom fields
		if context.Custom != nil {
			if val, ok := context.Custom[condition.Field]; ok {
				fieldValue = val
			}
		}
	}

	// Evaluate condition based on operator
	switch condition.Operator {
	case "equals":
		return fieldValue == condition.Value
	case "not_equals":
		return fieldValue != condition.Value
	case "in":
		if values, ok := condition.Value.([]interface{}); ok {
			for _, v := range values {
				if fieldValue == v {
					return true
				}
			}
		}
		return false
	case "not_in":
		if values, ok := condition.Value.([]interface{}); ok {
			for _, v := range values {
				if fieldValue == v {
					return false
				}
			}
			return true
		}
		return true
	case "contains":
		if fieldStr, ok := fieldValue.(string); ok {
			if condStr, ok := condition.Value.(string); ok {
				return strings.Contains(fieldStr, condStr)
			}
		}
		return false
	case "starts_with":
		if fieldStr, ok := fieldValue.(string); ok {
			if condStr, ok := condition.Value.(string); ok {
				return strings.HasPrefix(fieldStr, condStr)
			}
		}
		return false
	case "regex":
		// Would need to implement regex matching
		return false
	default:
		return false
	}
}

func (s *FeatureFlagService) evaluateRollout(rollout RolloutConfig, context *EvaluationContext) bool {
	// Check time window
	now := time.Now()
	if rollout.StartTime != nil && now.Before(*rollout.StartTime) {
		return false
	}
	if rollout.EndTime != nil && now.After(*rollout.EndTime) {
		return false
	}

	switch rollout.Type {
	case "percentage":
		// Simple hash-based percentage rollout
		hash := s.hashString(context.TenantID + context.UserID)
		return (hash % 100) < int(rollout.Percentage)

	case "user_list":
		for _, userID := range rollout.UserList {
			if userID == context.UserID {
				return true
			}
		}
		return false

	case "plan_based":
		for _, plan := range rollout.PlanList {
			if plan == context.Plan {
				return true
			}
		}
		return false

	default:
		return false
	}
}

func (s *FeatureFlagService) hashString(input string) int {
	hash := 0
	for _, char := range input {
		hash = (hash * 31) + int(char)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

func (s *FeatureFlagService) getCustomDomainFlagConfig() string {
	config := FeatureFlagConfig{
		Enabled: true,
		Rules: []FeatureFlagRule{
			{
				ID:          "free_plan_restriction",
				Name:        "Disable for free plans",
				Description: "Custom domains are only available for paid plans",
				Conditions: []Condition{
					{
						Field:    "plan",
						Operator: "in",
						Value:    []interface{}{"free", "starter"},
					},
				},
				Action:   "disable",
				Priority: 1,
				Enabled:  true,
			},
			{
				ID:          "paid_plan_access",
				Name:        "Enable for paid plans",
				Description: "Enable custom domains for Pro and Enterprise plans",
				Conditions: []Condition{
					{
						Field:    "plan",
						Operator: "in",
						Value:    []interface{}{"pro", "enterprise", "business"},
					},
				},
				Action:   "enable",
				Priority: 2,
				Enabled:  true,
			},
		},
		DefaultValue: false,
		AuditLog:     true,
	}

	configJSON, _ := json.Marshal(config)
	return string(configJSON)
}

func (s *FeatureFlagService) getSSLCertificateFlagConfig() string {
	config := FeatureFlagConfig{
		Enabled: true,
		Rules: []FeatureFlagRule{
			{
				ID:          "ssl_for_custom_domains",
				Name:        "SSL for custom domains only",
				Description: "SSL certificates are only issued for verified custom domains",
				Conditions: []Condition{
					{
						Field:    "plan",
						Operator: "not_in",
						Value:    []interface{}{"free"},
					},
				},
				Action:   "enable",
				Priority: 1,
				Enabled:  true,
			},
		},
		Metadata: map[string]interface{}{
			"provider":           "lets_encrypt",
			"auto_renewal":       true,
			"renewal_days":       30,
			"challenge_type":     "http-01",
			"fallback_challenge": "dns-01",
		},
		DefaultValue: false,
		AuditLog:     true,
	}

	configJSON, _ := json.Marshal(config)
	return string(configJSON)
}