// Package models provides data models for incident management in Tenant Admin Service.
package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Incident represents a service incident reported on the status page.
type Incident struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Title       string         `gorm:"not null;size:255" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Status      string         `gorm:"default:investigating" json:"status"`   // investigating, identified, monitoring, resolved
	Impact      string         `gorm:"default:minor" json:"impact"`            // none, minor, major, critical
	Severity    string         `gorm:"default:low" json:"severity"`            // low, medium, high, critical
	IsVisible   bool           `gorm:"default:true" json:"is_visible"`
	StartedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"started_at"`
	ResolvedAt  *time.Time     `json:"resolved_at,omitempty"`
	CreatedBy   *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`  // User ID who created incident
	UpdatedBy   *uuid.UUID     `gorm:"type:uuid" json:"updated_by,omitempty"`  // User ID who last updated
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for Incident.
func (Incident) TableName() string {
	return "saas_incidents"
}

// Validate performs validation on Incident.
func (i *Incident) Validate() error {
	if i.Title == "" {
		return fmt.Errorf("incident title is required")
	}
	if i.TenantID == uuid.Nil {
		return fmt.Errorf("tenant ID is required")
	}

	// Validate status
	validStatuses := map[string]bool{
		"investigating": true,
		"identified":    true,
		"monitoring":    true,
		"resolved":      true,
	}
	if i.Status != "" && !validStatuses[i.Status] {
		return fmt.Errorf("invalid incident status: %s", i.Status)
	}

	// Validate impact
	validImpacts := map[string]bool{
		"none":     true,
		"minor":    true,
		"major":    true,
		"critical": true,
	}
	if i.Impact != "" && !validImpacts[i.Impact] {
		return fmt.Errorf("invalid incident impact: %s", i.Impact)
	}

	// Validate severity
	validSeverities := map[string]bool{
		"low":      true,
		"medium":   true,
		"high":     true,
		"critical": true,
	}
	if i.Severity != "" && !validSeverities[i.Severity] {
		return fmt.Errorf("invalid incident severity: %s", i.Severity)
	}

	return nil
}

// GetStatusText returns human-readable status text.
func (i *Incident) GetStatusText() string {
	statusMap := map[string]string{
		"investigating": "Investigating",
		"identified":    "Identified",
		"monitoring":    "Monitoring",
		"resolved":      "Resolved",
	}

	if text, exists := statusMap[i.Status]; exists {
		return text
	}
	return i.Status
}

// GetImpactText returns human-readable impact text.
func (i *Incident) GetImpactText() string {
	impactMap := map[string]string{
		"none":     "No Impact",
		"minor":    "Minor Impact",
		"major":    "Major Impact",
		"critical": "Critical Impact",
	}

	if text, exists := impactMap[i.Impact]; exists {
		return text
	}
	return i.Impact
}

// IsResolved returns true if the incident is resolved.
func (i *Incident) IsResolved() bool {
	return i.Status == "resolved" || i.ResolvedAt != nil
}

// GetDuration returns the incident duration.
// If not resolved, returns duration from start to now.
func (i *Incident) GetDuration() time.Duration {
	if i.ResolvedAt != nil {
		return i.ResolvedAt.Sub(i.StartedAt)
	}
	return time.Since(i.StartedAt)
}
