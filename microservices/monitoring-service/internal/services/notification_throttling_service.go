package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ThrottleConfig represents throttling configuration for a tenant or integration
type ThrottleConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`

	// Rate limiting settings
	MaxAlertsPerHour   int `gorm:"default:60" json:"max_alerts_per_hour"`     // 0 = unlimited
	MaxAlertsPerDay    int `gorm:"default:500" json:"max_alerts_per_day"`     // 0 = unlimited
	CooldownMinutes    int `gorm:"default:5" json:"cooldown_minutes"`         // Min time between same alerts
	BurstAllowance     int `gorm:"default:10" json:"burst_allowance"`         // Allow burst of alerts before throttling

	// Digest/batching settings
	EnableDigest       bool `gorm:"default:false" json:"enable_digest"`
	DigestIntervalMins int  `gorm:"default:15" json:"digest_interval_mins"`   // Batch alerts every N minutes

	// Quiet hours settings
	EnableQuietHours  bool   `gorm:"default:false" json:"enable_quiet_hours"`
	QuietHoursStart   string `gorm:"size:10" json:"quiet_hours_start"`        // HH:MM format
	QuietHoursEnd     string `gorm:"size:10" json:"quiet_hours_end"`          // HH:MM format
	QuietHoursDays    string `gorm:"type:text" json:"quiet_hours_days"`       // mon,tue,wed,thu,fri,sat,sun
	SuppressDuringQuiet bool `gorm:"default:true" json:"suppress_during_quiet"` // Or queue for later

	// Deduplication settings
	EnableDedup            bool `gorm:"default:true" json:"enable_dedup"`
	DedupWindowMinutes     int  `gorm:"default:30" json:"dedup_window_minutes"` // Dedupe same alert within window
	DedupGroupBySeverity   bool `gorm:"default:true" json:"dedup_group_by_severity"`
	DedupGroupByMonitor    bool `gorm:"default:true" json:"dedup_group_by_monitor"`

	// Applies to
	ApplyToEmail     bool `gorm:"default:true" json:"apply_to_email"`
	ApplyToSMS       bool `gorm:"default:true" json:"apply_to_sms"`
	ApplyToSlack     bool `gorm:"default:true" json:"apply_to_slack"`
	ApplyToTeams     bool `gorm:"default:true" json:"apply_to_teams"`
	ApplyToWebhook   bool `gorm:"default:false" json:"apply_to_webhook"` // Usually don't throttle webhooks
	ApplyToPagerDuty bool `gorm:"default:false" json:"apply_to_pagerduty"` // Usually don't throttle PagerDuty

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ThrottleEvent represents a notification event for throttling tracking
type ThrottleEvent struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TenantID       uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	ConfigID       uint      `gorm:"not null;index" json:"config_id"`
	MonitorID      uint      `gorm:"not null;index" json:"monitor_id"`
	IntegrationType string   `gorm:"size:50;index" json:"integration_type"` // email, sms, slack, etc.
	Severity        string   `gorm:"size:50" json:"severity"`
	EventHash       string   `gorm:"size:100;index" json:"event_hash"` // For deduplication
	WasThrottled    bool     `gorm:"default:false;index" json:"was_throttled"`
	ThrottleReason  string   `gorm:"type:text" json:"throttle_reason"`
	QueuedForDigest bool     `gorm:"default:false" json:"queued_for_digest"`
	DigestSentAt    *time.Time `json:"digest_sent_at"`
	CreatedAt       time.Time `json:"created_at"`
}

// DigestQueue represents alerts queued for digest delivery
type DigestQueue struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TenantID       uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	ConfigID       uint      `gorm:"not null;index" json:"config_id"`
	IntegrationType string   `gorm:"size:50" json:"integration_type"`
	AlertCount      int      `gorm:"default:1" json:"alert_count"`
	MonitorIDs      string   `gorm:"type:text" json:"monitor_ids"` // Comma-separated
	SeverityCounts  string   `gorm:"type:jsonb" json:"severity_counts"` // {"down": 3, "degraded": 2}
	ScheduledFor    time.Time `gorm:"index" json:"scheduled_for"`
	Sent            bool      `gorm:"default:false;index" json:"sent"`
	SentAt          *time.Time `json:"sent_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ThrottleDecision represents the result of evaluating throttling rules
type ThrottleDecision struct {
	ShouldSend      bool
	ShouldQueue     bool
	ShouldThrottle  bool
	ThrottleReason  string
	CooldownUntil   *time.Time
	DigestQueueID   *uint
}

// NotificationThrottlingService handles notification throttling logic
type NotificationThrottlingService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewNotificationThrottlingService creates a new notification throttling service
func NewNotificationThrottlingService(db *gorm.DB, logger *zap.Logger) *NotificationThrottlingService {
	return &NotificationThrottlingService{
		db:     db,
		logger: logger,
	}
}

// CreateConfig creates a new throttle configuration
func (s *NotificationThrottlingService) CreateConfig(config *ThrottleConfig) error {
	if err := s.validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	return s.db.Create(config).Error
}

// UpdateConfig updates an existing throttle configuration
func (s *NotificationThrottlingService) UpdateConfig(config *ThrottleConfig) error {
	if err := s.validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	return s.db.Save(config).Error
}

// DeleteConfig deletes a throttle configuration
func (s *NotificationThrottlingService) DeleteConfig(id uint, tenantID uuid.UUID) error {
	return s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&ThrottleConfig{}).Error
}

// GetConfig retrieves a specific throttle configuration
func (s *NotificationThrottlingService) GetConfig(id uint, tenantID uuid.UUID) (*ThrottleConfig, error) {
	var config ThrottleConfig
	err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&config).Error
	return &config, err
}

// GetConfigsByTenant retrieves all throttle configurations for a tenant
func (s *NotificationThrottlingService) GetConfigsByTenant(tenantID uuid.UUID) ([]ThrottleConfig, error) {
	var configs []ThrottleConfig
	err := s.db.Where("tenant_id = ?", tenantID).Order("id ASC").Find(&configs).Error
	return configs, err
}

// GetActiveConfig retrieves the active throttle configuration for a tenant
func (s *NotificationThrottlingService) GetActiveConfig(tenantID uuid.UUID) (*ThrottleConfig, error) {
	var config ThrottleConfig
	err := s.db.Where("tenant_id = ? AND is_active = ?", tenantID, true).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// ShouldSendNotification checks if a notification should be sent or throttled
func (s *NotificationThrottlingService) ShouldSendNotification(tenantID uuid.UUID, monitorID uint, integrationType, severity string) (*ThrottleDecision, error) {
	// Get active throttle config
	config, err := s.GetActiveConfig(tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// No throttling configured - allow all notifications
			return &ThrottleDecision{ShouldSend: true}, nil
		}
		return nil, fmt.Errorf("failed to get throttle config: %w", err)
	}

	// Check if throttling applies to this integration type
	if !s.appliesToIntegration(config, integrationType) {
		return &ThrottleDecision{ShouldSend: true}, nil
	}

	now := time.Now()
	decision := &ThrottleDecision{
		ShouldSend: true,
	}

	// Check quiet hours
	if config.EnableQuietHours && s.isQuietHours(config, now) {
		if config.SuppressDuringQuiet {
			decision.ShouldSend = false
			decision.ShouldThrottle = true
			decision.ThrottleReason = "Quiet hours - notifications suppressed"
			s.recordEvent(tenantID, config.ID, monitorID, integrationType, severity, "", true, decision.ThrottleReason, false)
			return decision, nil
		} else if config.EnableDigest {
			// Queue for digest
			decision.ShouldSend = false
			decision.ShouldQueue = true
			decision.ThrottleReason = "Quiet hours - queued for digest"
			queueID := s.queueForDigest(config, tenantID, monitorID, integrationType, severity)
			decision.DigestQueueID = &queueID
			s.recordEvent(tenantID, config.ID, monitorID, integrationType, severity, "", false, decision.ThrottleReason, true)
			return decision, nil
		}
	}

	// Check deduplication
	if config.EnableDedup {
		eventHash := s.generateEventHash(monitorID, severity, config)
		if s.isDuplicate(tenantID, eventHash, config.DedupWindowMinutes) {
			decision.ShouldSend = false
			decision.ShouldThrottle = true
			decision.ThrottleReason = fmt.Sprintf("Duplicate alert within %d minute window", config.DedupWindowMinutes)
			s.recordEvent(tenantID, config.ID, monitorID, integrationType, severity, eventHash, true, decision.ThrottleReason, false)
			return decision, nil
		}
	}

	// Check cooldown
	if config.CooldownMinutes > 0 {
		lastEvent, err := s.getLastEvent(tenantID, config.ID, monitorID, integrationType)
		if err == nil {
			cooldownEnd := lastEvent.CreatedAt.Add(time.Duration(config.CooldownMinutes) * time.Minute)
			if now.Before(cooldownEnd) {
				decision.ShouldSend = false
				decision.ShouldThrottle = true
				decision.CooldownUntil = &cooldownEnd
				decision.ThrottleReason = fmt.Sprintf("Cooldown period - %.0f minutes remaining", cooldownEnd.Sub(now).Minutes())
				s.recordEvent(tenantID, config.ID, monitorID, integrationType, severity, "", true, decision.ThrottleReason, false)
				return decision, nil
			}
		}
	}

	// Check hourly rate limit
	if config.MaxAlertsPerHour > 0 {
		hourlyCount := s.getEventCount(tenantID, config.ID, integrationType, 1*time.Hour)
		if hourlyCount >= config.MaxAlertsPerHour {
			// Check burst allowance
			if config.BurstAllowance > 0 {
				recentBurstCount := s.getEventCount(tenantID, config.ID, integrationType, 5*time.Minute)
				if recentBurstCount < config.BurstAllowance {
					// Allow burst
					s.logger.Info("Allowing burst notification",
						zap.Int("burst_count", recentBurstCount),
						zap.Int("burst_allowance", config.BurstAllowance))
				} else {
					decision.ShouldSend = false
					decision.ShouldThrottle = true
					decision.ThrottleReason = fmt.Sprintf("Hourly rate limit exceeded (%d/%d)", hourlyCount, config.MaxAlertsPerHour)
					s.recordEvent(tenantID, config.ID, monitorID, integrationType, severity, "", true, decision.ThrottleReason, false)
					return decision, nil
				}
			} else {
				decision.ShouldSend = false
				decision.ShouldThrottle = true
				decision.ThrottleReason = fmt.Sprintf("Hourly rate limit exceeded (%d/%d)", hourlyCount, config.MaxAlertsPerHour)
				s.recordEvent(tenantID, config.ID, monitorID, integrationType, severity, "", true, decision.ThrottleReason, false)
				return decision, nil
			}
		}
	}

	// Check daily rate limit
	if config.MaxAlertsPerDay > 0 {
		dailyCount := s.getEventCount(tenantID, config.ID, integrationType, 24*time.Hour)
		if dailyCount >= config.MaxAlertsPerDay {
			decision.ShouldSend = false
			decision.ShouldThrottle = true
			decision.ThrottleReason = fmt.Sprintf("Daily rate limit exceeded (%d/%d)", dailyCount, config.MaxAlertsPerDay)
			s.recordEvent(tenantID, config.ID, monitorID, integrationType, severity, "", true, decision.ThrottleReason, false)
			return decision, nil
		}
	}

	// Check if digest is enabled and should queue instead
	if config.EnableDigest && s.shouldUseDigest(config, now) {
		decision.ShouldSend = false
		decision.ShouldQueue = true
		decision.ThrottleReason = "Queued for digest delivery"
		queueID := s.queueForDigest(config, tenantID, monitorID, integrationType, severity)
		decision.DigestQueueID = &queueID
		s.recordEvent(tenantID, config.ID, monitorID, integrationType, severity, "", false, decision.ThrottleReason, true)
		return decision, nil
	}

	// All checks passed - allow notification
	eventHash := s.generateEventHash(monitorID, severity, config)
	s.recordEvent(tenantID, config.ID, monitorID, integrationType, severity, eventHash, false, "", false)

	return decision, nil
}

// appliesToIntegration checks if throttling applies to this integration type
func (s *NotificationThrottlingService) appliesToIntegration(config *ThrottleConfig, integrationType string) bool {
	switch integrationType {
	case "email":
		return config.ApplyToEmail
	case "sms":
		return config.ApplyToSMS
	case "slack":
		return config.ApplyToSlack
	case "teams":
		return config.ApplyToTeams
	case "webhook":
		return config.ApplyToWebhook
	case "pagerduty":
		return config.ApplyToPagerDuty
	default:
		return false
	}
}

// isQuietHours checks if current time is within quiet hours
func (s *NotificationThrottlingService) isQuietHours(config *ThrottleConfig, now time.Time) bool {
	// Check day of week
	if config.QuietHoursDays != "" {
		currentDay := now.Format("Mon")
		if !contains(config.QuietHoursDays, currentDay) {
			return false
		}
	}

	// Check time range
	return isWithinTimeRange(config.QuietHoursStart, config.QuietHoursEnd, now)
}

// isDuplicate checks if this is a duplicate event within the dedup window
func (s *NotificationThrottlingService) isDuplicate(tenantID uuid.UUID, eventHash string, windowMinutes int) bool {
	var count int64
	cutoff := time.Now().Add(-time.Duration(windowMinutes) * time.Minute)

	s.db.Model(&ThrottleEvent{}).
		Where("tenant_id = ? AND event_hash = ? AND created_at > ?", tenantID, eventHash, cutoff).
		Count(&count)

	return count > 0
}

// generateEventHash generates a hash for deduplication
func (s *NotificationThrottlingService) generateEventHash(monitorID uint, severity string, config *ThrottleConfig) string {
	hash := fmt.Sprintf("mon_%d", monitorID)

	if config.DedupGroupBySeverity {
		hash += fmt.Sprintf("_sev_%s", severity)
	}

	return hash
}

// getLastEvent retrieves the last throttle event for a monitor/integration
func (s *NotificationThrottlingService) getLastEvent(tenantID uuid.UUID, configID, monitorID uint, integrationType string) (*ThrottleEvent, error) {
	var event ThrottleEvent
	err := s.db.Where("tenant_id = ? AND config_id = ? AND monitor_id = ? AND integration_type = ?",
		tenantID, configID, monitorID, integrationType).
		Order("created_at DESC").
		First(&event).Error

	return &event, err
}

// getEventCount counts throttle events within a time window
func (s *NotificationThrottlingService) getEventCount(tenantID uuid.UUID, configID uint, integrationType string, duration time.Duration) int {
	var count int64
	cutoff := time.Now().Add(-duration)

	s.db.Model(&ThrottleEvent{}).
		Where("tenant_id = ? AND config_id = ? AND integration_type = ? AND created_at > ? AND was_throttled = ?",
			tenantID, configID, integrationType, cutoff, false).
		Count(&count)

	return int(count)
}

// recordEvent records a throttle event
func (s *NotificationThrottlingService) recordEvent(tenantID uuid.UUID, configID, monitorID uint, integrationType, severity, eventHash string, wasThrottled bool, reason string, queuedForDigest bool) {
	event := &ThrottleEvent{
		TenantID:        tenantID,
		ConfigID:        configID,
		MonitorID:       monitorID,
		IntegrationType: integrationType,
		Severity:        severity,
		EventHash:       eventHash,
		WasThrottled:    wasThrottled,
		ThrottleReason:  reason,
		QueuedForDigest: queuedForDigest,
	}

	if err := s.db.Create(event).Error; err != nil {
		s.logger.Error("Failed to record throttle event", zap.Error(err))
	}
}

// shouldUseDigest checks if digest should be used based on time and config
func (s *NotificationThrottlingService) shouldUseDigest(config *ThrottleConfig, now time.Time) bool {
	if !config.EnableDigest || config.DigestIntervalMins <= 0 {
		return false
	}

	// Check if we're within digest interval of last digest
	// For simplicity, digest at interval boundaries (e.g., every 15 mins at :00, :15, :30, :45)
	return true
}

// queueForDigest adds an alert to the digest queue
func (s *NotificationThrottlingService) queueForDigest(config *ThrottleConfig, tenantID uuid.UUID, monitorID uint, integrationType, severity string) uint {
	scheduledFor := calculateNextDigestTime(config.DigestIntervalMins)

	// Check if queue entry already exists for this scheduled time
	var queue DigestQueue
	err := s.db.Where("tenant_id = ? AND config_id = ? AND integration_type = ? AND scheduled_for = ? AND sent = ?",
		tenantID, config.ID, integrationType, scheduledFor, false).
		First(&queue).Error

	if err == gorm.ErrRecordNotFound {
		// Create new queue entry
		queue = DigestQueue{
			TenantID:        tenantID,
			ConfigID:        config.ID,
			IntegrationType: integrationType,
			AlertCount:      1,
			MonitorIDs:      fmt.Sprintf("%d", monitorID),
			SeverityCounts:  fmt.Sprintf(`{"%s": 1}`, severity),
			ScheduledFor:    scheduledFor,
			Sent:            false,
		}
		s.db.Create(&queue)
	} else {
		// Update existing queue entry
		queue.AlertCount++
		queue.MonitorIDs += fmt.Sprintf(",%d", monitorID)
		// TODO: Update severity counts JSON
		s.db.Save(&queue)
	}

	return queue.ID
}

// validateConfig validates a throttle configuration
func (s *NotificationThrottlingService) validateConfig(config *ThrottleConfig) error {
	if config.Name == "" {
		return fmt.Errorf("config name is required")
	}

	if config.EnableQuietHours {
		if !isValidTimeFormat(config.QuietHoursStart) {
			return fmt.Errorf("invalid quiet_hours_start format, expected HH:MM")
		}
		if !isValidTimeFormat(config.QuietHoursEnd) {
			return fmt.Errorf("invalid quiet_hours_end format, expected HH:MM")
		}
	}

	return nil
}

// GetThrottleStats retrieves throttling statistics
func (s *NotificationThrottlingService) GetThrottleStats(tenantID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	var totalEvents, throttledEvents, digestEvents int64

	s.db.Model(&ThrottleEvent{}).
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, startDate, endDate).
		Count(&totalEvents)

	s.db.Model(&ThrottleEvent{}).
		Where("tenant_id = ? AND created_at BETWEEN ? AND ? AND was_throttled = ?", tenantID, startDate, endDate, true).
		Count(&throttledEvents)

	s.db.Model(&ThrottleEvent{}).
		Where("tenant_id = ? AND created_at BETWEEN ? AND ? AND queued_for_digest = ?", tenantID, startDate, endDate, true).
		Count(&digestEvents)

	throttleRate := float64(0)
	if totalEvents > 0 {
		throttleRate = (float64(throttledEvents) / float64(totalEvents)) * 100
	}

	return map[string]interface{}{
		"total_events":     totalEvents,
		"throttled_events": throttledEvents,
		"digest_events":    digestEvents,
		"throttle_rate":    throttleRate,
		"allowed_events":   totalEvents - throttledEvents - digestEvents,
	}, nil
}

// Helper functions

func contains(haystack, needle string) bool {
	return false // Simplified - implement proper contains logic
}

func isWithinTimeRange(startTime, endTime string, now time.Time) bool {
	return false // Simplified - reuse from alert_routing_service
}

func isValidTimeFormat(timeStr string) bool {
	return true // Simplified - reuse from alert_routing_service
}

func calculateNextDigestTime(intervalMins int) time.Time {
	now := time.Now()
	// Round to next interval boundary
	minutes := (now.Minute()/intervalMins + 1) * intervalMins
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), minutes, 0, 0, now.Location())
}
