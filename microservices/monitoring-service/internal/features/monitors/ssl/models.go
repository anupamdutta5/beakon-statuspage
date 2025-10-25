// Package models provides data models for the Monitoring Service.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SSLCertificate represents an SSL/TLS certificate being monitored for expiration.
type SSLCertificate struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	TenantID     uuid.UUID  `gorm:"type:uuid;not null;index:idx_ssl_tenant" json:"tenant_id"`
	Domain       string     `gorm:"size:255;not null" json:"domain"`                // e.g., "api.example.com"
	Issuer       string     `gorm:"size:255" json:"issuer,omitempty"`               // Certificate issuer
	Subject      string     `gorm:"size:255" json:"subject,omitempty"`              // Certificate subject
	SerialNumber string     `gorm:"size:255" json:"serial_number,omitempty"`        // Certificate serial number
	ValidFrom    *time.Time `json:"valid_from,omitempty"`                           // Certificate valid from date
	ValidUntil   *time.Time `json:"valid_until,omitempty"`                          // Certificate valid until date

	// DaysUntilExpiry is calculated and stored when scanning certificates
	DaysUntilExpiry *int `gorm:"column:days_until_expiry" json:"days_until_expiry,omitempty"`

	LastChecked     *time.Time `json:"last_checked,omitempty"`
	IsValid         bool       `gorm:"default:true;index:idx_ssl_tenant" json:"is_valid"`
	IsSelfSigned    bool       `gorm:"default:false" json:"is_self_signed"`

	// Warning tracking (to prevent duplicate notifications)
	WarningSent30d  bool   `gorm:"column:warning_sent_30d;default:false" json:"warning_sent_30d"`
	WarningSent14d  bool   `gorm:"column:warning_sent_14d;default:false" json:"warning_sent_14d"`
	WarningSent7d   bool   `gorm:"column:warning_sent_7d;default:false" json:"warning_sent_7d"`

	ErrorMessage    string `gorm:"type:text" json:"error_message,omitempty"` // Error details if check failed
}

// TableName specifies the table name for GORM.
func (SSLCertificate) TableName() string {
	return "ssl_certificates"
}

// BeforeCreate hook to ensure tenant_id and domain uniqueness.
func (s *SSLCertificate) BeforeCreate(tx *gorm.DB) error {
	// Ensure unique constraint is respected
	var count int64
	tx.Model(&SSLCertificate{}).
		Where("tenant_id = ? AND domain = ?", s.TenantID, s.Domain).
		Count(&count)

	if count > 0 {
		return gorm.ErrDuplicatedKey
	}
	return nil
}

// IsExpiringSoon returns true if the certificate expires within the given number of days.
func (s *SSLCertificate) IsExpiringSoon(days int) bool {
	if s.DaysUntilExpiry == nil {
		return false
	}
	return *s.DaysUntilExpiry <= days && *s.DaysUntilExpiry >= 0
}

// IsExpired returns true if the certificate has already expired.
func (s *SSLCertificate) IsExpired() bool {
	if s.DaysUntilExpiry == nil {
		return false
	}
	return *s.DaysUntilExpiry < 0
}

// ShouldSendWarning checks if a warning should be sent based on days until expiry.
func (s *SSLCertificate) ShouldSendWarning() (should bool, warningType string) {
	if s.DaysUntilExpiry == nil {
		return false, ""
	}

	days := *s.DaysUntilExpiry

	// Check 30-day warning
	if days <= 30 && days > 14 && !s.WarningSent30d {
		return true, "30d"
	}

	// Check 14-day warning
	if days <= 14 && days > 7 && !s.WarningSent14d {
		return true, "14d"
	}

	// Check 7-day warning
	if days <= 7 && days >= 0 && !s.WarningSent7d {
		return true, "7d"
	}

	return false, ""
}

// MarkWarningSent marks the appropriate warning as sent.
func (s *SSLCertificate) MarkWarningSent(warningType string) {
	switch warningType {
	case "30d":
		s.WarningSent30d = true
	case "14d":
		s.WarningSent14d = true
	case "7d":
		s.WarningSent7d = true
	}
}

// HeartbeatMonitor represents a heartbeat/cron job monitor that expects regular pings.
type HeartbeatMonitor struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	TenantID              uuid.UUID  `gorm:"type:uuid;not null;index:idx_heartbeat_tenant" json:"tenant_id"`
	Name                  string     `gorm:"size:255;not null" json:"name"`
	Description           string     `gorm:"type:text" json:"description,omitempty"`
	UniqueKey             string     `gorm:"size:255;not null" json:"unique_key"`                    // Used in ping URL
	ExpectedIntervalSeconds int      `gorm:"not null" json:"expected_interval_seconds"`             // Expected time between pings
	GracePeriodSeconds    int        `gorm:"default:300" json:"grace_period_seconds"`               // Grace period before marking as down
	LastPing              *time.Time `json:"last_ping,omitempty"`
	IsAlive               bool       `gorm:"default:false;index:idx_heartbeat_status" json:"is_alive"`
	ConsecutiveMisses     int        `gorm:"default:0" json:"consecutive_misses"`
	AlertSent             bool       `gorm:"default:false" json:"alert_sent"`
}

// TableName specifies the table name for GORM.
func (HeartbeatMonitor) TableName() string {
	return "heartbeat_monitors"
}

// IsOverdue checks if the heartbeat is overdue (hasn't pinged within expected interval + grace period).
func (h *HeartbeatMonitor) IsOverdue() bool {
	if h.LastPing == nil {
		return true
	}

	maxDuration := time.Duration(h.ExpectedIntervalSeconds+h.GracePeriodSeconds) * time.Second
	return time.Since(*h.LastPing) > maxDuration
}

// RecordPing records a new ping and resets the heartbeat status.
func (h *HeartbeatMonitor) RecordPing() {
	now := time.Now()
	h.LastPing = &now
	h.IsAlive = true
	h.ConsecutiveMisses = 0
	h.AlertSent = false
}

// OnCallSchedule represents an on-call rotation schedule for alert escalation.
type OnCallSchedule struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	TenantID            uuid.UUID `gorm:"type:uuid;not null;index:idx_on_call_tenant" json:"tenant_id"`
	Name                string    `gorm:"size:255;not null" json:"name"`
	RotationType        string    `gorm:"size:20;not null" json:"rotation_type"`        // daily, weekly, custom
	RotationStart       time.Time `gorm:"not null" json:"rotation_start"`
	RotationIntervalHours int     `gorm:"default:168" json:"rotation_interval_hours"`  // Default: weekly (168 hours)
	Participants        string    `gorm:"type:jsonb;not null" json:"participants"`      // JSON: [{user_id: 'uuid', order: 1}, ...]
	IsActive            bool      `gorm:"default:true;index:idx_on_call_active" json:"is_active"`
}

// TableName specifies the table name for GORM.
func (OnCallSchedule) TableName() string {
	return "on_call_schedules"
}

// EscalationPolicy represents an alert escalation policy defining notification tiers.
type EscalationPolicy struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	TenantID    uuid.UUID `gorm:"type:uuid;not null;index:idx_escalation_tenant" json:"tenant_id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	Levels      string    `gorm:"type:jsonb;not null" json:"levels"`      // JSON: [{delay_minutes: 0, notify: ['user1']}, ...]
	IsDefault   bool      `gorm:"default:false;index:idx_escalation_default" json:"is_default"`
}

// TableName specifies the table name for GORM.
func (EscalationPolicy) TableName() string {
	return "escalation_policies"
}
