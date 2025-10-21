// Package services provides on-call rotation schedule management.
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

// OnCallParticipant represents a participant in an on-call rotation.
type OnCallParticipant struct {
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name,omitempty"`
	Email       string    `json:"email,omitempty"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	Order       int       `json:"order"` // Position in rotation
}

// CurrentOnCallInfo represents the current on-call person information.
type CurrentOnCallInfo struct {
	ScheduleID      uint              `json:"schedule_id"`
	ScheduleName    string            `json:"schedule_name"`
	Participant     OnCallParticipant `json:"participant"`
	RotationStart   time.Time         `json:"rotation_start"`
	RotationEnd     time.Time         `json:"rotation_end"`
	TimeRemaining   string            `json:"time_remaining"`
	NextParticipant *OnCallParticipant `json:"next_participant,omitempty"`
}

// OnCallService handles on-call rotation schedule operations.
type OnCallService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewOnCallService creates a new on-call service.
func NewOnCallService(db *gorm.DB, logger *zap.Logger) *OnCallService {
	// Auto-migrate on-call schedules table
	if err := db.AutoMigrate(&models.OnCallSchedule{}); err != nil {
		logger.Error("Failed to migrate on_call_schedules table", zap.Error(err))
	}

	return &OnCallService{
		db:     db,
		logger: logger,
	}
}

// CreateSchedule creates a new on-call rotation schedule.
func (s *OnCallService) CreateSchedule(tenantID uuid.UUID, schedule *models.OnCallSchedule) error {
	// Validate rotation type
	if schedule.RotationType != "daily" && schedule.RotationType != "weekly" && schedule.RotationType != "custom" {
		return fmt.Errorf("invalid rotation type: %s (must be daily, weekly, or custom)", schedule.RotationType)
	}

	// Set default interval based on rotation type
	if schedule.RotationIntervalHours == 0 {
		switch schedule.RotationType {
		case "daily":
			schedule.RotationIntervalHours = 24
		case "weekly":
			schedule.RotationIntervalHours = 168
		default:
			return fmt.Errorf("rotation_interval_hours required for custom rotation type")
		}
	}

	// Validate participants
	participants, err := s.parseParticipants(schedule.Participants)
	if err != nil {
		return fmt.Errorf("invalid participants format: %w", err)
	}

	if len(participants) == 0 {
		return fmt.Errorf("at least one participant is required")
	}

	schedule.TenantID = tenantID

	if err := s.db.Create(schedule).Error; err != nil {
		s.logger.Error("Failed to create on-call schedule", zap.Error(err))
		return fmt.Errorf("failed to create on-call schedule: %w", err)
	}

	s.logger.Info("On-call schedule created",
		zap.Uint("schedule_id", schedule.ID),
		zap.String("name", schedule.Name),
		zap.String("rotation_type", schedule.RotationType),
		zap.Int("participants", len(participants)),
	)

	return nil
}

// GetSchedules retrieves all on-call schedules for a tenant.
func (s *OnCallService) GetSchedules(tenantID uuid.UUID, includeInactive bool) ([]models.OnCallSchedule, error) {
	var schedules []models.OnCallSchedule
	query := s.db.Where("tenant_id = ?", tenantID)

	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Order("created_at DESC").Find(&schedules).Error; err != nil {
		s.logger.Error("Failed to get on-call schedules", zap.Error(err))
		return nil, fmt.Errorf("failed to get on-call schedules: %w", err)
	}

	return schedules, nil
}

// GetSchedule retrieves a specific on-call schedule by ID.
func (s *OnCallService) GetSchedule(id uint, tenantID uuid.UUID) (*models.OnCallSchedule, error) {
	var schedule models.OnCallSchedule
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&schedule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("on-call schedule not found")
		}
		s.logger.Error("Failed to get on-call schedule", zap.Error(err))
		return nil, fmt.Errorf("failed to get on-call schedule: %w", err)
	}

	return &schedule, nil
}

// UpdateSchedule updates an on-call schedule.
func (s *OnCallService) UpdateSchedule(id uint, tenantID uuid.UUID, updates map[string]interface{}) error {
	// Validate participants if being updated
	if participantsJSON, ok := updates["participants"].(string); ok {
		_, err := s.parseParticipants(participantsJSON)
		if err != nil {
			return fmt.Errorf("invalid participants format: %w", err)
		}
	}

	result := s.db.Model(&models.OnCallSchedule{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(updates)

	if result.Error != nil {
		s.logger.Error("Failed to update on-call schedule", zap.Error(result.Error))
		return fmt.Errorf("failed to update on-call schedule: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("on-call schedule not found")
	}

	s.logger.Info("On-call schedule updated", zap.Uint("schedule_id", id))
	return nil
}

// DeleteSchedule deletes an on-call schedule.
func (s *OnCallService) DeleteSchedule(id uint, tenantID uuid.UUID) error {
	result := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.OnCallSchedule{})

	if result.Error != nil {
		s.logger.Error("Failed to delete on-call schedule", zap.Error(result.Error))
		return fmt.Errorf("failed to delete on-call schedule: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("on-call schedule not found")
	}

	s.logger.Info("On-call schedule deleted", zap.Uint("schedule_id", id))
	return nil
}

// GetCurrentOnCall determines who is currently on-call for a given schedule.
func (s *OnCallService) GetCurrentOnCall(scheduleID uint, tenantID uuid.UUID) (*CurrentOnCallInfo, error) {
	schedule, err := s.GetSchedule(scheduleID, tenantID)
	if err != nil {
		return nil, err
	}

	if !schedule.IsActive {
		return nil, fmt.Errorf("schedule is not active")
	}

	participants, err := s.parseParticipants(schedule.Participants)
	if err != nil {
		return nil, fmt.Errorf("failed to parse participants: %w", err)
	}

	if len(participants) == 0 {
		return nil, fmt.Errorf("no participants in schedule")
	}

	// Calculate current rotation
	now := time.Now()
	rotationDuration := time.Duration(schedule.RotationIntervalHours) * time.Hour

	// Calculate how many rotations have passed since the start
	timeSinceStart := now.Sub(schedule.RotationStart)
	if timeSinceStart < 0 {
		// Schedule hasn't started yet
		return nil, fmt.Errorf("schedule has not started yet (starts at %s)", schedule.RotationStart.Format(time.RFC3339))
	}

	rotationsPassed := int(timeSinceStart / rotationDuration)
	currentParticipantIndex := rotationsPassed % len(participants)

	// Get current and next participants
	currentParticipant := participants[currentParticipantIndex]

	var nextParticipant *OnCallParticipant
	if len(participants) > 1 {
		nextIndex := (currentParticipantIndex + 1) % len(participants)
		nextParticipant = &participants[nextIndex]
	}

	// Calculate rotation start and end times
	rotationStart := schedule.RotationStart.Add(time.Duration(rotationsPassed) * rotationDuration)
	rotationEnd := rotationStart.Add(rotationDuration)

	timeRemaining := time.Until(rotationEnd)

	info := &CurrentOnCallInfo{
		ScheduleID:      schedule.ID,
		ScheduleName:    schedule.Name,
		Participant:     currentParticipant,
		RotationStart:   rotationStart,
		RotationEnd:     rotationEnd,
		TimeRemaining:   formatDuration(timeRemaining),
		NextParticipant: nextParticipant,
	}

	s.logger.Info("Current on-call retrieved",
		zap.Uint("schedule_id", scheduleID),
		zap.String("current_user", currentParticipant.UserID.String()),
		zap.String("time_remaining", info.TimeRemaining),
	)

	return info, nil
}

// GetAllCurrentOnCall retrieves current on-call information for all active schedules.
func (s *OnCallService) GetAllCurrentOnCall(tenantID uuid.UUID) ([]CurrentOnCallInfo, error) {
	schedules, err := s.GetSchedules(tenantID, false) // Only active schedules
	if err != nil {
		return nil, err
	}

	var currentOnCallList []CurrentOnCallInfo

	for _, schedule := range schedules {
		info, err := s.GetCurrentOnCall(schedule.ID, tenantID)
		if err != nil {
			s.logger.Warn("Failed to get current on-call for schedule",
				zap.Uint("schedule_id", schedule.ID),
				zap.Error(err),
			)
			continue
		}

		currentOnCallList = append(currentOnCallList, *info)
	}

	return currentOnCallList, nil
}

// AddParticipant adds a participant to an on-call schedule.
func (s *OnCallService) AddParticipant(scheduleID uint, tenantID uuid.UUID, participant OnCallParticipant) error {
	schedule, err := s.GetSchedule(scheduleID, tenantID)
	if err != nil {
		return err
	}

	participants, err := s.parseParticipants(schedule.Participants)
	if err != nil {
		return fmt.Errorf("failed to parse existing participants: %w", err)
	}

	// Check if user already exists
	for _, p := range participants {
		if p.UserID == participant.UserID {
			return fmt.Errorf("participant already exists in schedule")
		}
	}

	// Set order to next position
	participant.Order = len(participants) + 1

	// Add new participant
	participants = append(participants, participant)

	// Serialize back to JSON
	participantsJSON, err := json.Marshal(participants)
	if err != nil {
		return fmt.Errorf("failed to marshal participants: %w", err)
	}

	// Update schedule
	return s.UpdateSchedule(scheduleID, tenantID, map[string]interface{}{
		"participants": string(participantsJSON),
	})
}

// RemoveParticipant removes a participant from an on-call schedule.
func (s *OnCallService) RemoveParticipant(scheduleID uint, tenantID uuid.UUID, userID uuid.UUID) error {
	schedule, err := s.GetSchedule(scheduleID, tenantID)
	if err != nil {
		return err
	}

	participants, err := s.parseParticipants(schedule.Participants)
	if err != nil {
		return fmt.Errorf("failed to parse existing participants: %w", err)
	}

	// Remove participant
	var updatedParticipants []OnCallParticipant
	found := false
	for _, p := range participants {
		if p.UserID == userID {
			found = true
			continue
		}
		updatedParticipants = append(updatedParticipants, p)
	}

	if !found {
		return fmt.Errorf("participant not found in schedule")
	}

	if len(updatedParticipants) == 0 {
		return fmt.Errorf("cannot remove last participant from schedule")
	}

	// Reorder participants
	for i := range updatedParticipants {
		updatedParticipants[i].Order = i + 1
	}

	// Serialize back to JSON
	participantsJSON, err := json.Marshal(updatedParticipants)
	if err != nil {
		return fmt.Errorf("failed to marshal participants: %w", err)
	}

	// Update schedule
	return s.UpdateSchedule(scheduleID, tenantID, map[string]interface{}{
		"participants": string(participantsJSON),
	})
}

// ReorderParticipants updates the order of participants in a schedule.
func (s *OnCallService) ReorderParticipants(scheduleID uint, tenantID uuid.UUID, orderedUserIDs []uuid.UUID) error {
	schedule, err := s.GetSchedule(scheduleID, tenantID)
	if err != nil {
		return err
	}

	participants, err := s.parseParticipants(schedule.Participants)
	if err != nil {
		return fmt.Errorf("failed to parse existing participants: %w", err)
	}

	if len(orderedUserIDs) != len(participants) {
		return fmt.Errorf("ordered list must contain all participants")
	}

	// Create map of existing participants
	participantMap := make(map[uuid.UUID]OnCallParticipant)
	for _, p := range participants {
		participantMap[p.UserID] = p
	}

	// Reorder participants
	var reorderedParticipants []OnCallParticipant
	for i, userID := range orderedUserIDs {
		p, exists := participantMap[userID]
		if !exists {
			return fmt.Errorf("user %s not found in schedule", userID)
		}
		p.Order = i + 1
		reorderedParticipants = append(reorderedParticipants, p)
	}

	// Serialize back to JSON
	participantsJSON, err := json.Marshal(reorderedParticipants)
	if err != nil {
		return fmt.Errorf("failed to marshal participants: %w", err)
	}

	// Update schedule
	return s.UpdateSchedule(scheduleID, tenantID, map[string]interface{}{
		"participants": string(participantsJSON),
	})
}

// GetScheduleByName retrieves an on-call schedule by name.
func (s *OnCallService) GetScheduleByName(tenantID uuid.UUID, name string) (*models.OnCallSchedule, error) {
	var schedule models.OnCallSchedule
	if err := s.db.Where("tenant_id = ? AND name = ?", tenantID, name).First(&schedule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("on-call schedule not found")
		}
		return nil, err
	}
	return &schedule, nil
}

// parseParticipants parses the JSON participants string.
func (s *OnCallService) parseParticipants(participantsJSON string) ([]OnCallParticipant, error) {
	var participants []OnCallParticipant
	if err := json.Unmarshal([]byte(participantsJSON), &participants); err != nil {
		return nil, err
	}
	return participants, nil
}

// formatDuration formats a duration in a human-readable format.
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "0s"
	}

	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
