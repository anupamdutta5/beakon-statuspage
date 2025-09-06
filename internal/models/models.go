package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Status constants for components/services
const (
	StatusOperational         = "operational"
	StatusDegradedPerformance = "degraded_performance"
	StatusPartialOutage      = "partial_outage"
	StatusMajorOutage        = "major_outage"
	StatusUnderMaintenance   = "under_maintenance"
)

// Incident status constants
const (
	IncidentStatusInvestigating = "investigating"
	IncidentStatusIdentified    = "identified"
	IncidentStatusMonitoring    = "monitoring"
	IncidentStatusResolved      = "resolved"
)

// Maintenance status constants
const (
	MaintenanceStatusScheduled   = "scheduled"
	MaintenanceStatusInProgress  = "in_progress"
	MaintenanceStatusCompleted  = "completed"
	MaintenanceStatusCancelled   = "cancelled"
)

// User roles
const (
	RoleAdmin   = "admin"
	RoleEditor  = "editor"
	RoleViewer  = "viewer"
)

// Service represents a monitored service or component.
type Service struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	Status      string `gorm:"default:'operational'"`
	Group       string // Component group for organization
	ShowUptime  bool   `gorm:"default:true"` // Whether to show uptime for this component
	Position    int    `gorm:"default:0"`    // Display order
}

// Component is an alias for Service, used for clarity in the context of Atlassian-style status pages.
type Component Service

// Incident represents a service incident.
type Incident struct {
	gorm.Model
	Title       string `gorm:"not null"`
	Description string
	Status      string `gorm:"default:'investigating'"`
	Impact      string `gorm:"default:'minor'"` // minor, major, critical
	ResolvedAt  *time.Time
	Updates     []StatusUpdate `gorm:"foreignKey:IncidentID"`
	Services    []*Service     `gorm:"many2many:incident_services;"`
	TemplateID  *uint          // Reference to incident template
	Template    *IncidentTemplate `gorm:"foreignKey:TemplateID"`
}

// StatusUpdate represents an update to an incident.
type StatusUpdate struct {
	gorm.Model
	IncidentID  uint   `gorm:"not null"`
	Description string `gorm:"not null"`
	Status      string `gorm:"not null"`
}

// User represents an admin user.
type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	Role     string `gorm:"default:'admin'"`
}

// Maintenance represents a scheduled maintenance event.
type Maintenance struct {
	gorm.Model
	Title       string    `gorm:"not null"`
	Description string
	Status      string    `gorm:"not null;default:'scheduled'"` // e.g., scheduled, in_progress, completed
	StartAt     time.Time `gorm:"not null"`
	EndAt       time.Time `gorm:"not null"`
	Services    []*Service  `gorm:"many2many:maintenance_services;"`
}

// Subscriber represents a user who has subscribed to notifications.
type Subscriber struct {
	gorm.Model
	Email    string     `gorm:"uniqueIndex;not null"`
	Services []*Service `gorm:"many2many:subscriber_services;"`
}

// Monitor represents a check to be performed on a service.
type Monitor struct {
	gorm.Model
	Name           string `gorm:"not null"`
	ServiceID      uint   `gorm:"not null"`
	Service        Service
	Type           string `gorm:"not null;default:'http'"` // e.g., http, ping
	URL            string `gorm:"not null"`
	Interval       uint   `gorm:"default:60"` // in seconds
	ExpectedStatus int    `gorm:"default:200"`
	Timeout        int    `gorm:"default:10"` // in seconds
	LastCheckAt    time.Time
	LastResult     string // e.g., up, down
}

// Heartbeat represents a single check result for a monitor.
type Heartbeat struct {
	gorm.Model
	MonitorID uint      `gorm:"not null"`
	Monitor   Monitor
	Timestamp time.Time `gorm:"not null"`
	Status    string    `gorm:"not null"` // e.g., up, down
	Latency   int64     // in milliseconds
	Message   string
}

// IncidentTemplate represents a template for creating incidents
type IncidentTemplate struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Title       string `gorm:"not null"`
	Description string
	Impact      string `gorm:"default:'minor'"`
	Services    []*Service `gorm:"many2many:incident_template_services;"`
}

// MaintenanceTemplate represents a template for creating maintenance events
type MaintenanceTemplate struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Title       string `gorm:"not null"`
	Description string
	Duration    uint   `gorm:"default:60"` // Duration in minutes
	Services    []*Service `gorm:"many2many:maintenance_template_services;"`
}

// AuditLog represents audit trail for admin actions
type AuditLog struct {
	gorm.Model
	UserID      uint      `gorm:"not null"`
	User        User
	Action      string    `gorm:"not null"` // create, update, delete, etc.
	Resource    string    `gorm:"not null"` // incident, maintenance, service, etc.
	ResourceID  uint      `gorm:"not null"`
	Details     string    // JSON string with change details
	IPAddress   string
	UserAgent   string
	Timestamp   time.Time `gorm:"not null"`
}

// Branding represents customizable branding settings
type Branding struct {
	gorm.Model
	CompanyName    string `gorm:"default:'Your Company'"`
	LogoURL        string
	FaviconURL     string
	PrimaryColor   string `gorm:"default:'#0052cc'"`
	SecondaryColor string `gorm:"default:'#f4f5f7'"`
	CustomCSS      string
	CustomDomain   string
	FooterText     string
}

// Integration represents third-party integrations
type Integration struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Type        string `gorm:"not null"` // slack, pagerduty, opsgenie, prometheus
	Config      string // JSON configuration
	IsActive    bool   `gorm:"default:false"`
	LastSyncAt  *time.Time
}

// SystemMetric represents system performance metrics
type SystemMetric struct {
	gorm.Model
	ServiceID   uint      `gorm:"not null"`
	Service     Service
	Name        string    `gorm:"not null"` // response_time, uptime, etc.
	Value       float64   `gorm:"not null"`
	Unit        string    // ms, %, etc.
	Timestamp   time.Time `gorm:"not null"`
}

// ThirdPartyService represents external services to monitor
type ThirdPartyService struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string
	StatusURL   string `gorm:"not null"` // URL to fetch status from
	Status      string `gorm:"default:'operational'"`
	LastCheckAt time.Time
	IsActive    bool   `gorm:"default:true"`
}

// PrivatePage represents access-controlled status pages
type PrivatePage struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string
	AccessKey   string `gorm:"uniqueIndex;not null"`
	IsActive    bool   `gorm:"default:true"`
	Services    []*Service `gorm:"many2many:private_page_services;"`
}

// BeforeSave hashes the user's password before saving to the database.
func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	if u.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	return nil
}
