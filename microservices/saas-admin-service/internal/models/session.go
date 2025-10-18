package models

import (
	"time"

	"gorm.io/gorm"
)

// Session represents an active user session in the database
// Maps to the existing 'sessions' table in saas_admin database
type Session struct {
	ID        string         `gorm:"primaryKey;size:128" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	TenantID  uint           `gorm:"not null;index" json:"tenant_id"`
	IPAddress string         `gorm:"size:45" json:"ip_address"`
	UserAgent string         `gorm:"type:text" json:"user_agent"`
	IsActive  bool           `gorm:"default:true;index" json:"is_active"`
	LastSeen  time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"last_seen"`
	ExpiresAt time.Time      `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Session model
func (Session) TableName() string {
	return "sessions"
}

// UserSession represents a refresh token session
// Maps to the existing 'user_sessions' table in saas_admin database
// Used for long-lived refresh tokens (7 days) to issue new short-lived JWTs
type UserSession struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Token     string         `gorm:"not null;uniqueIndex;size:255" json:"token"`
	ExpiresAt time.Time      `gorm:"not null;index" json:"expires_at"`
	IPAddress string         `gorm:"type:text" json:"ip_address"`
	UserAgent string         `gorm:"type:text" json:"user_agent"`
	IsActive  bool           `gorm:"default:true;index" json:"is_active"`
}

// TableName specifies the table name for UserSession model
func (UserSession) TableName() string {
	return "user_sessions"
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsExpired checks if the user session (refresh token) has expired
func (us *UserSession) IsExpired() bool {
	return time.Now().After(us.ExpiresAt)
}

// IsValid checks if session is active and not expired
func (s *Session) IsValid() bool {
	return s.IsActive && !s.IsExpired()
}

// IsValid checks if user session is active and not expired
func (us *UserSession) IsValid() bool {
	return us.IsActive && !us.IsExpired()
}
