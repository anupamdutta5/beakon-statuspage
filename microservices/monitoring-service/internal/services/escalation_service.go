// Package services provides escalation policy management for alert routing.
package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/anupamdutta5/monitoring-service/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// EscalationLevel represents a single level in an escalation policy.
type EscalationLevel struct {
	Level         int         `json:"level"`                    // 1, 2, 3, etc.
	DelayMinutes  int         `json:"delay_minutes"`            // Delay before escalating to this level
	NotifyUsers   []uuid.UUID `json:"notify_users,omitempty"`   // User IDs to notify
	NotifySchedule *uint      `json:"notify_schedule,omitempty"` // On-call schedule ID
	NotifyChannels []string   `json:"notify_channels"`          // email, sms, webhook, slack
}

// EscalationTracker tracks the current state of an escalation.
type EscalationTracker struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	TenantID      uuid.UUID `gorm:"type:uuid;not null;index:idx_escalation_tracker_tenant" json:"tenant_id"`
	IncidentID    uuid.UUID `gorm:"type:uuid;not null;index:idx_escalation_tracker_incident" json:"incident_id"`
	PolicyID      uint      `gorm:"not null;index:idx_escalation_tracker_policy" json:"policy_id"`
	CurrentLevel  int       `gorm:"default:1" json:"current_level"`
	IsResolved    bool      `gorm:"default:false;index:idx_escalation_tracker_resolved" json:"is_resolved"`
	LastEscalated *time.Time `json:"last_escalated,omitempty"`
	NotifiedUsers string    `gorm:"type:text" json:"notified_users"` // JSON: [uuid, ...]
}

// TableName specifies the table name for GORM.
func (EscalationTracker) TableName() string {
	return "escalation_trackers"
}

// EscalationService handles escalation policy operations.
type EscalationService struct {
	db            *gorm.DB
	logger        *zap.Logger
	smsService    *SMSService
	onCallService *OnCallService
}

// NewEscalationService creates a new escalation service.
func NewEscalationService(db *gorm.DB, logger *zap.Logger, smsService *SMSService, onCallService *OnCallService) *EscalationService {
	// Auto-migrate escalation tables
	if err := db.AutoMigrate(&models.EscalationPolicy{}, &EscalationTracker{}); err != nil {
		logger.Error("Failed to migrate escalation tables", zap.Error(err))
	}

	return &EscalationService{
		db:            db,
		logger:        logger,
		smsService:    smsService,
		onCallService: onCallService,
	}
}

// CreatePolicy creates a new escalation policy.
func (s *EscalationService) CreatePolicy(tenantID uuid.UUID, policy *models.EscalationPolicy) error {
	// Validate levels
	levels, err := s.parseLevels(policy.Levels)
	if err != nil {
		return fmt.Errorf("invalid levels format: %w", err)
	}

	if len(levels) == 0 {
		return fmt.Errorf("at least one escalation level is required")
	}

	// Validate level ordering
	for i, level := range levels {
		if level.Level != i+1 {
			return fmt.Errorf("levels must be sequential starting from 1")
		}
		if i > 0 && level.DelayMinutes <= levels[i-1].DelayMinutes {
			return fmt.Errorf("level %d delay must be greater than level %d delay", level.Level, level.Level-1)
		}
	}

	policy.TenantID = tenantID

	// If this is set as default, unset other defaults
	if policy.IsDefault {
		s.db.Model(&models.EscalationPolicy{}).
			Where("tenant_id = ? AND is_default = ?", tenantID, true).
			Update("is_default", false)
	}

	if err := s.db.Create(policy).Error; err != nil {
		s.logger.Error("Failed to create escalation policy", zap.Error(err))
		return fmt.Errorf("failed to create escalation policy: %w", err)
	}

	s.logger.Info("Escalation policy created",
		zap.Uint("policy_id", policy.ID),
		zap.String("name", policy.Name),
		zap.Int("levels", len(levels)),
	)

	return nil
}

// GetPolicies retrieves all escalation policies for a tenant.
func (s *EscalationService) GetPolicies(tenantID uuid.UUID) ([]models.EscalationPolicy, error) {
	var policies []models.EscalationPolicy
	if err := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&policies).Error; err != nil {
		s.logger.Error("Failed to get escalation policies", zap.Error(err))
		return nil, fmt.Errorf("failed to get escalation policies: %w", err)
	}

	return policies, nil
}

// GetPolicy retrieves a specific escalation policy by ID.
func (s *EscalationService) GetPolicy(id uint, tenantID uuid.UUID) (*models.EscalationPolicy, error) {
	var policy models.EscalationPolicy
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&policy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("escalation policy not found")
		}
		s.logger.Error("Failed to get escalation policy", zap.Error(err))
		return nil, fmt.Errorf("failed to get escalation policy: %w", err)
	}

	return &policy, nil
}

// GetDefaultPolicy retrieves the default escalation policy for a tenant.
func (s *EscalationService) GetDefaultPolicy(tenantID uuid.UUID) (*models.EscalationPolicy, error) {
	var policy models.EscalationPolicy
	if err := s.db.Where("tenant_id = ? AND is_default = ?", tenantID, true).First(&policy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no default escalation policy found")
		}
		return nil, err
	}
	return &policy, nil
}

// UpdatePolicy updates an escalation policy.
func (s *EscalationService) UpdatePolicy(id uint, tenantID uuid.UUID, updates map[string]interface{}) error {
	// Validate levels if being updated
	if levelsJSON, ok := updates["levels"].(string); ok {
		_, err := s.parseLevels(levelsJSON)
		if err != nil {
			return fmt.Errorf("invalid levels format: %w", err)
		}
	}

	// If setting as default, unset other defaults
	if isDefault, ok := updates["is_default"].(bool); ok && isDefault {
		s.db.Model(&models.EscalationPolicy{}).
			Where("tenant_id = ? AND is_default = ? AND id != ?", tenantID, true, id).
			Update("is_default", false)
	}

	result := s.db.Model(&models.EscalationPolicy{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(updates)

	if result.Error != nil {
		s.logger.Error("Failed to update escalation policy", zap.Error(result.Error))
		return fmt.Errorf("failed to update escalation policy: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("escalation policy not found")
	}

	s.logger.Info("Escalation policy updated", zap.Uint("policy_id", id))
	return nil
}

// DeletePolicy deletes an escalation policy.
func (s *EscalationService) DeletePolicy(id uint, tenantID uuid.UUID) error {
	result := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.EscalationPolicy{})

	if result.Error != nil {
		s.logger.Error("Failed to delete escalation policy", zap.Error(result.Error))
		return fmt.Errorf("failed to delete escalation policy: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("escalation policy not found")
	}

	s.logger.Info("Escalation policy deleted", zap.Uint("policy_id", id))
	return nil
}

// StartEscalation begins an escalation for an incident.
func (s *EscalationService) StartEscalation(tenantID, incidentID uuid.UUID, policyID uint) (*EscalationTracker, error) {
	// Check if escalation already exists for this incident
	var existingTracker EscalationTracker
	err := s.db.Where("incident_id = ? AND is_resolved = ?", incidentID, false).First(&existingTracker).Error
	if err == nil {
		return &existingTracker, nil // Already escalating
	}

	// Get policy
	policy, err := s.GetPolicy(policyID, tenantID)
	if err != nil {
		return nil, err
	}

	// Create tracker
	tracker := &EscalationTracker{
		TenantID:     tenantID,
		IncidentID:   incidentID,
		PolicyID:     policyID,
		CurrentLevel: 1,
		IsResolved:   false,
	}

	if err := s.db.Create(tracker).Error; err != nil {
		s.logger.Error("Failed to create escalation tracker", zap.Error(err))
		return nil, fmt.Errorf("failed to create escalation tracker: %w", err)
	}

	s.logger.Info("Escalation started",
		zap.Uint("tracker_id", tracker.ID),
		zap.String("incident_id", incidentID.String()),
		zap.String("policy", policy.Name),
	)

	// Send initial notification (level 1)
	if err := s.notifyLevel(tracker, policy, 1); err != nil {
		s.logger.Error("Failed to send initial notification", zap.Error(err))
	}

	return tracker, nil
}

// ProcessEscalations checks all active escalations and escalates if needed.
// This should be called periodically (e.g., every minute) by a background job.
func (s *EscalationService) ProcessEscalations() error {
	var trackers []EscalationTracker
	if err := s.db.Where("is_resolved = ?", false).Find(&trackers).Error; err != nil {
		return err
	}

	s.logger.Info("Processing escalations", zap.Int("count", len(trackers)))

	for _, tracker := range trackers {
		if err := s.processEscalation(&tracker); err != nil {
			s.logger.Error("Failed to process escalation",
				zap.Uint("tracker_id", tracker.ID),
				zap.Error(err),
			)
		}
	}

	return nil
}

// processEscalation processes a single escalation tracker.
func (s *EscalationService) processEscalation(tracker *EscalationTracker) error {
	// Get policy
	policy, err := s.GetPolicy(tracker.PolicyID, tracker.TenantID)
	if err != nil {
		return err
	}

	levels, err := s.parseLevels(policy.Levels)
	if err != nil {
		return err
	}

	// Check if we're at the last level
	if tracker.CurrentLevel >= len(levels) {
		return nil // Already at max escalation
	}

	nextLevel := tracker.CurrentLevel + 1
	nextLevelConfig := levels[nextLevel-1]

	// Check if enough time has passed to escalate
	timeSinceCreated := time.Since(tracker.CreatedAt)
	if int(timeSinceCreated.Minutes()) >= nextLevelConfig.DelayMinutes {
		// Escalate to next level
		tracker.CurrentLevel = nextLevel
		now := time.Now()
		tracker.LastEscalated = &now

		if err := s.db.Save(tracker).Error; err != nil {
			return err
		}

		s.logger.Info("Escalating to next level",
			zap.Uint("tracker_id", tracker.ID),
			zap.Int("level", nextLevel),
		)

		// Send notification for this level
		return s.notifyLevel(tracker, policy, nextLevel)
	}

	return nil
}

// notifyLevel sends notifications for a specific escalation level.
func (s *EscalationService) notifyLevel(tracker *EscalationTracker, policy *models.EscalationPolicy, level int) error {
	levels, err := s.parseLevels(policy.Levels)
	if err != nil {
		return err
	}

	if level > len(levels) {
		return fmt.Errorf("invalid escalation level: %d", level)
	}

	levelConfig := levels[level-1]

	// Collect all recipients
	var recipients []string
	var notifiedUserIDs []uuid.UUID

	// Add users from level config
	for _, userID := range levelConfig.NotifyUsers {
		// TODO: Fetch user phone number from user service
		// For now, we'll just track the user ID
		notifiedUserIDs = append(notifiedUserIDs, userID)
	}

	// Add on-call person if schedule specified
	if levelConfig.NotifySchedule != nil {
		onCallInfo, err := s.onCallService.GetCurrentOnCall(*levelConfig.NotifySchedule, tracker.TenantID)
		if err != nil {
			s.logger.Error("Failed to get current on-call", zap.Error(err))
		} else {
			notifiedUserIDs = append(notifiedUserIDs, onCallInfo.Participant.UserID)
			if onCallInfo.Participant.PhoneNumber != "" {
				recipients = append(recipients, onCallInfo.Participant.PhoneNumber)
			}
		}
	}

	// Update tracker with notified users
	notifiedJSON, _ := json.Marshal(notifiedUserIDs)
	tracker.NotifiedUsers = string(notifiedJSON)
	s.db.Save(tracker)

	// Send notifications via enabled channels
	for _, channel := range levelConfig.NotifyChannels {
		switch channel {
		case "sms":
			if s.smsService != nil && s.smsService.IsEnabled() && len(recipients) > 0 {
				message := fmt.Sprintf("🚨 ESCALATION Level %d: Incident %s requires attention", level, tracker.IncidentID)
				err := s.smsService.SendCustomSMS(tracker.TenantID, message, recipients)
				if err != nil {
					s.logger.Error("Failed to send SMS escalation", zap.Error(err))
				}
			}
		case "email":
			// TODO: Implement email notification
			s.logger.Info("Email notification would be sent", zap.Int("level", level))
		case "webhook":
			// TODO: Implement webhook notification
			s.logger.Info("Webhook notification would be sent", zap.Int("level", level))
		case "slack":
			// TODO: Implement Slack notification
			s.logger.Info("Slack notification would be sent", zap.Int("level", level))
		}
	}

	s.logger.Info("Level notification sent",
		zap.Uint("tracker_id", tracker.ID),
		zap.Int("level", level),
		zap.Int("recipients", len(notifiedUserIDs)),
	)

	return nil
}

// ResolveEscalation marks an escalation as resolved.
func (s *EscalationService) ResolveEscalation(incidentID uuid.UUID) error {
	result := s.db.Model(&EscalationTracker{}).
		Where("incident_id = ? AND is_resolved = ?", incidentID, false).
		Update("is_resolved", true)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		s.logger.Info("Escalation resolved",
			zap.String("incident_id", incidentID.String()),
		)
	}

	return nil
}

// GetActiveEscalations retrieves all active escalations for a tenant.
func (s *EscalationService) GetActiveEscalations(tenantID uuid.UUID) ([]EscalationTracker, error) {
	var trackers []EscalationTracker
	if err := s.db.Where("tenant_id = ? AND is_resolved = ?", tenantID, false).
		Order("created_at DESC").
		Find(&trackers).Error; err != nil {
		return nil, err
	}

	return trackers, nil
}

// parseLevels parses the JSON levels string.
func (s *EscalationService) parseLevels(levelsJSON string) ([]EscalationLevel, error) {
	var levels []EscalationLevel
	if err := json.Unmarshal([]byte(levelsJSON), &levels); err != nil {
		return nil, err
	}
	return levels, nil
}
