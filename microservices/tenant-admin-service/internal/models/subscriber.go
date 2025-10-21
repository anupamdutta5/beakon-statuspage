// Package models provides data models for subscriber management in Tenant Admin Service.
package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Subscriber represents an email subscriber for status page updates.
type Subscriber struct {
	ID               uuid.UUID      `gorm:"type:uuid;primarykey;default:uuid_generate_v4()" json:"id"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Email            string         `gorm:"not null;index" json:"email"`
	Phone            string         `json:"phone,omitempty"`
	Status           string         `gorm:"default:active" json:"status"` // active, unsubscribed, bounced
	EventTypes       string         `json:"event_types,omitempty"`        // JSON array of event types
	Components       string         `json:"components,omitempty"`         // JSON array of component IDs
	VerifiedAt       *time.Time     `json:"verified_at,omitempty"`
	UnsubscribeToken string         `gorm:"unique" json:"unsubscribe_token,omitempty"`
	Preferences      string         `json:"preferences,omitempty"` // JSON object for preferences
}

// TableName returns the table name for Subscriber.
func (Subscriber) TableName() string {
	return "saas_subscribers"
}

// Validate performs validation on Subscriber.
func (s *Subscriber) Validate() error {
	if s.Email == "" {
		return fmt.Errorf("subscriber email is required")
	}
	if s.TenantID == uuid.Nil {
		return fmt.Errorf("tenant ID is required")
	}

	// Validate status
	validStatuses := map[string]bool{
		"active":       true,
		"unsubscribed": true,
		"bounced":      true,
	}
	if s.Status != "" && !validStatuses[s.Status] {
		return fmt.Errorf("invalid subscriber status: %s", s.Status)
	}

	return nil
}

// IsVerified returns true if the subscriber has verified their email.
func (s *Subscriber) IsVerified() bool {
	return s.VerifiedAt != nil
}

// IsActive returns true if the subscriber is active and verified.
func (s *Subscriber) IsActive() bool {
	return s.Status == "active" && s.IsVerified()
}
