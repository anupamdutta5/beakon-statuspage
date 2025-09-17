// Package models provides SLA reporting data models for the Analytics Service.
package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// SLA represents a Service Level Agreement definition.
type SLA struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID       uint           `gorm:"not null;index" json:"tenant_id"`
	ComponentID    *uint          `gorm:"index" json:"component_id"`
	Name           string         `gorm:"not null" json:"name"`
	Description    string         `json:"description"`
	Type           string         `gorm:"not null" json:"type"`           // uptime, response_time, error_rate
	TargetValue    float64        `gorm:"not null" json:"target_value"`   // 99.9 for uptime, 200ms for response time
	Unit           string         `gorm:"not null" json:"unit"`           // percentage, milliseconds, requests_per_second
	PeriodType     string         `gorm:"not null" json:"period_type"`    // monthly, quarterly, yearly, rolling_30d
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	AlertThreshold float64        `json:"alert_threshold"`                // threshold for SLA breach alerts
	Settings       string         `gorm:"type:text" json:"settings"`     // JSON string for additional settings
	Metadata       string         `gorm:"type:text" json:"metadata"`     // JSON string for additional data

	// Related entities
	Measurements []SLAMeasurement `gorm:"foreignKey:SLAID" json:"measurements,omitempty"`
	Breaches     []SLABreach      `gorm:"foreignKey:SLAID" json:"breaches,omitempty"`
	Reports      []SLAReport      `gorm:"foreignKey:SLAID" json:"reports,omitempty"`
}

// SLAMeasurement represents a measurement of SLA compliance.
type SLAMeasurement struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	SLAID         uint           `gorm:"not null;index" json:"sla_id"`
	SLA           SLA            `gorm:"foreignKey:SLAID" json:"sla"`
	PeriodStart   time.Time      `gorm:"not null;index" json:"period_start"`
	PeriodEnd     time.Time      `gorm:"not null;index" json:"period_end"`
	ActualValue   float64        `gorm:"not null" json:"actual_value"`
	TargetValue   float64        `gorm:"not null" json:"target_value"`
	ComplianceRate float64       `gorm:"not null" json:"compliance_rate"` // percentage of compliance
	IsCompliant   bool           `gorm:"not null" json:"is_compliant"`
	TotalSamples  int64          `json:"total_samples"`
	ValidSamples  int64          `json:"valid_samples"`
	FailedSamples int64          `json:"failed_samples"`
	Status        string         `gorm:"default:active" json:"status"` // active, breached, warning
	Details       string         `gorm:"type:text" json:"details"`     // JSON string for detailed measurements
	Metadata      string         `gorm:"type:text" json:"metadata"`    // JSON string for additional data
}

// SLABreach represents an SLA breach incident.
type SLABreach struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	SLAID         uint           `gorm:"not null;index" json:"sla_id"`
	SLA           SLA            `gorm:"foreignKey:SLAID" json:"sla"`
	MeasurementID uint           `gorm:"not null;index" json:"measurement_id"`
	Measurement   SLAMeasurement `gorm:"foreignKey:MeasurementID" json:"measurement"`
	Severity      string         `gorm:"not null" json:"severity"`      // warning, minor, major, critical
	Status        string         `gorm:"default:open" json:"status"`    // open, acknowledged, resolved
	TriggeredAt   time.Time      `gorm:"not null" json:"triggered_at"`
	DetectedAt    time.Time      `gorm:"not null" json:"detected_at"`
	ResolvedAt    *time.Time     `json:"resolved_at"`
	Duration      int64          `json:"duration"`                      // duration in seconds
	ImpactValue   float64        `json:"impact_value"`                  // how much the SLA was missed by
	Description   string         `json:"description"`
	RootCause     string         `json:"root_cause"`
	Resolution    string         `json:"resolution"`
	NotifiedUsers string         `gorm:"type:text" json:"notified_users"` // JSON array of notified user IDs
	Metadata      string         `gorm:"type:text" json:"metadata"`       // JSON string for additional data
}

// SLAReport represents a generated SLA report.
type SLAReport struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID       uint           `gorm:"not null;index" json:"tenant_id"`
	SLAID          *uint          `gorm:"index" json:"sla_id"`          // null for multi-SLA reports
	SLA            *SLA           `gorm:"foreignKey:SLAID" json:"sla"`
	Name           string         `gorm:"not null" json:"name"`
	Type           string         `gorm:"not null" json:"type"`         // summary, detailed, trending, compliance
	Period         string         `gorm:"not null" json:"period"`       // monthly, quarterly, yearly, custom
	PeriodStart    time.Time      `gorm:"not null" json:"period_start"`
	PeriodEnd      time.Time      `gorm:"not null" json:"period_end"`
	Status         string         `gorm:"default:generating" json:"status"` // generating, completed, failed
	Format         string         `gorm:"default:pdf" json:"format"`        // pdf, html, json, csv
	GeneratedAt    *time.Time     `json:"generated_at"`
	FilePath       string         `json:"file_path"`
	FileSize       int64          `json:"file_size"`
	DownloadURL    string         `json:"download_url"`
	ExpiresAt      *time.Time     `json:"expires_at"`
	Summary        string         `gorm:"type:text" json:"summary"`    // JSON string for report summary
	Error          string         `json:"error"`
	Metadata       string         `gorm:"type:text" json:"metadata"`  // JSON string for additional data
}

// UptimeCalculation represents uptime calculation data.
type UptimeCalculation struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID        uint           `gorm:"not null;index" json:"tenant_id"`
	ComponentID     uint           `gorm:"not null;index" json:"component_id"`
	PeriodStart     time.Time      `gorm:"not null;index" json:"period_start"`
	PeriodEnd       time.Time      `gorm:"not null;index" json:"period_end"`
	TotalMinutes    int64          `gorm:"not null" json:"total_minutes"`
	UptimeMinutes   int64          `gorm:"not null" json:"uptime_minutes"`
	DowntimeMinutes int64          `gorm:"not null" json:"downtime_minutes"`
	UptimePercent   float64        `gorm:"not null" json:"uptime_percent"`
	IncidentCount   int            `json:"incident_count"`
	MaintenanceMinutes int64       `json:"maintenance_minutes"`
	CalculationType string         `gorm:"not null" json:"calculation_type"` // daily, weekly, monthly
	Status          string         `gorm:"default:active" json:"status"`     // active, archived
	Details         string         `gorm:"type:text" json:"details"`         // JSON string for detailed breakdown
	Metadata        string         `gorm:"type:text" json:"metadata"`        // JSON string for additional data
}

// ResponseTimeMetric represents response time measurement data.
type ResponseTimeMetric struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID      uint           `gorm:"not null;index" json:"tenant_id"`
	ComponentID   uint           `gorm:"not null;index" json:"component_id"`
	Timestamp     time.Time      `gorm:"not null;index" json:"timestamp"`
	ResponseTime  float64        `gorm:"not null" json:"response_time"` // in milliseconds
	StatusCode    int            `json:"status_code"`
	IsSuccessful  bool           `gorm:"not null" json:"is_successful"`
	Endpoint      string         `json:"endpoint"`
	Method        string         `json:"method"`
	Location      string         `json:"location"` // monitoring location
	ErrorMessage  string         `json:"error_message"`
	Metadata      string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// SLATarget represents an SLA target configuration.
type SLATarget struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID     uint           `gorm:"not null;index" json:"tenant_id"`
	Name         string         `gorm:"not null" json:"name"`
	Description  string         `json:"description"`
	Type         string         `gorm:"not null" json:"type"`         // uptime, response_time, availability
	ServiceTier  string         `gorm:"not null" json:"service_tier"` // basic, standard, premium, enterprise
	TargetValue  float64        `gorm:"not null" json:"target_value"`
	Unit         string         `gorm:"not null" json:"unit"`
	IsDefault    bool           `gorm:"default:false" json:"is_default"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	Settings     string         `gorm:"type:text" json:"settings"`   // JSON string for target settings
	Metadata     string         `gorm:"type:text" json:"metadata"`   // JSON string for additional data
}

// Table name functions
func (SLA) TableName() string {
	return "slas"
}

func (SLAMeasurement) TableName() string {
	return "sla_measurements"
}

func (SLABreach) TableName() string {
	return "sla_breaches"
}

func (SLAReport) TableName() string {
	return "sla_reports"
}

func (UptimeCalculation) TableName() string {
	return "uptime_calculations"
}

func (ResponseTimeMetric) TableName() string {
	return "response_time_metrics"
}

func (SLATarget) TableName() string {
	return "sla_targets"
}

// Validation methods
func (s *SLA) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("SLA name is required")
	}
	if s.TenantID == 0 {
		return fmt.Errorf("tenant ID is required")
	}
	if s.Type == "" {
		return fmt.Errorf("SLA type is required")
	}
	if s.TargetValue <= 0 {
		return fmt.Errorf("target value must be greater than 0")
	}
	if s.Unit == "" {
		return fmt.Errorf("unit is required")
	}
	if s.PeriodType == "" {
		return fmt.Errorf("period type is required")
	}

	// Validate SLA type
	validTypes := []string{"uptime", "response_time", "error_rate", "availability"}
	isValidType := false
	for _, validType := range validTypes {
		if s.Type == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("invalid SLA type: %s", s.Type)
	}

	// Validate period type
	validPeriods := []string{"monthly", "quarterly", "yearly", "rolling_30d", "rolling_7d"}
	isValidPeriod := false
	for _, validPeriod := range validPeriods {
		if s.PeriodType == validPeriod {
			isValidPeriod = true
			break
		}
	}
	if !isValidPeriod {
		return fmt.Errorf("invalid period type: %s", s.PeriodType)
	}

	return nil
}

func (u *UptimeCalculation) CalculateUptimePercent() {
	if u.TotalMinutes > 0 {
		u.UptimePercent = (float64(u.UptimeMinutes) / float64(u.TotalMinutes)) * 100
	} else {
		u.UptimePercent = 0
	}
}

func (s *SLAMeasurement) CalculateCompliance() {
	switch s.SLA.Type {
	case "uptime":
		if s.TargetValue > 0 {
			s.ComplianceRate = (s.ActualValue / s.TargetValue) * 100
		}
	case "response_time":
		// For response time, lower is better
		if s.ActualValue <= s.TargetValue {
			s.ComplianceRate = 100.0
		} else {
			s.ComplianceRate = (s.TargetValue / s.ActualValue) * 100
		}
	case "error_rate":
		// For error rate, lower is better
		if s.ActualValue <= s.TargetValue {
			s.ComplianceRate = 100.0
		} else {
			s.ComplianceRate = (s.TargetValue / s.ActualValue) * 100
		}
	default:
		if s.TargetValue > 0 {
			s.ComplianceRate = (s.ActualValue / s.TargetValue) * 100
		}
	}

	// Determine compliance
	s.IsCompliant = s.ActualValue >= s.TargetValue

	// For metrics where lower is better
	if s.SLA.Type == "response_time" || s.SLA.Type == "error_rate" {
		s.IsCompliant = s.ActualValue <= s.TargetValue
	}
}