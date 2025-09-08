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
	StatusPartialOutage       = "partial_outage"
	StatusMajorOutage         = "major_outage"
	StatusUnderMaintenance    = "under_maintenance"
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
	MaintenanceStatusScheduled  = "scheduled"
	MaintenanceStatusInProgress = "in_progress"
	MaintenanceStatusCompleted  = "completed"
	MaintenanceStatusCancelled  = "cancelled"
)

// User roles
const (
	RoleAdmin  = "admin"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

// Service represents a monitored service or component.
type Service struct {
	gorm.Model
	Name           string `gorm:"not null"`
	Description    string
	Status         string `gorm:"default:'operational'"`
	Group          string // Component group for organization
	ShowUptime     bool   `gorm:"default:true"` // Whether to show uptime for this component
	Position       int    `gorm:"default:0"`    // Display order
	HealthCheckURL string // URL for health checks
	TenantID       uint   `gorm:"not null"`
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
	Severity    string `gorm:"default:'low'"`   // low, medium, high, critical
	ResolvedAt  *time.Time
	Updates     []StatusUpdate    `gorm:"foreignKey:IncidentID"`
	Services    []*Service        `gorm:"many2many:incident_services;"`
	TemplateID  *uint             // Reference to incident template
	Template    *IncidentTemplate `gorm:"foreignKey:TemplateID"`
	TenantID    uint              `gorm:"not null"`
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
	Username    string  `gorm:"not null"`
	Password    string  `gorm:"not null"`
	Role        string  `gorm:"default:'admin'"`
	Email       string  `gorm:"not null"`
	TenantID    *uint   // Nullable for super admin users
	Tenant      *Tenant `gorm:"foreignKey:TenantID"`
	IsActive    bool    `gorm:"default:true"`
	LastLoginAt *time.Time
}

// Maintenance represents a scheduled maintenance event.
type Maintenance struct {
	gorm.Model
	Title       string `gorm:"not null"`
	Description string
	Status      string     `gorm:"not null;default:'scheduled'"` // e.g., scheduled, in_progress, completed
	StartAt     time.Time  `gorm:"not null"`
	EndAt       time.Time  `gorm:"not null"`
	Services    []*Service `gorm:"many2many:maintenance_services;"`
	TenantID    uint       `gorm:"not null"`
}

// Subscriber represents a user who has subscribed to notifications.
type Subscriber struct {
	gorm.Model
	Email    string     `gorm:"not null"`
	Phone    string     // Phone number for SMS notifications
	Services []*Service `gorm:"many2many:subscriber_services;"`
	TenantID uint       `gorm:"not null"`
	Tenant   Tenant     `gorm:"foreignKey:TenantID"`
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
	TenantID       uint   `gorm:"not null"`
	Tenant         Tenant `gorm:"foreignKey:TenantID"`
}

// Heartbeat represents a single check result for a monitor.
type Heartbeat struct {
	gorm.Model
	MonitorID uint `gorm:"not null"`
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
	Impact      string     `gorm:"default:'minor'"`
	Services    []*Service `gorm:"many2many:incident_template_services;"`
}

// MaintenanceTemplate represents a template for creating maintenance events
type MaintenanceTemplate struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Title       string `gorm:"not null"`
	Description string
	Duration    uint       `gorm:"default:60"` // Duration in minutes
	Services    []*Service `gorm:"many2many:maintenance_template_services;"`
}

// AuditLog represents audit trail for admin actions
type AuditLog struct {
	gorm.Model
	UserID     uint `gorm:"not null"`
	User       User
	Action     string `gorm:"not null"` // create, update, delete, etc.
	Resource   string `gorm:"not null"` // incident, maintenance, service, etc.
	ResourceID uint   `gorm:"not null"`
	Details    string // JSON string with change details
	IPAddress  string
	UserAgent  string
	Timestamp  time.Time `gorm:"not null"`
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
	TenantID       uint   `gorm:"not null"`
	Tenant         Tenant `gorm:"foreignKey:TenantID"`
}

// Integration represents third-party integrations
type Integration struct {
	gorm.Model
	Name       string `gorm:"not null"`
	Type       string `gorm:"not null"` // slack, pagerduty, opsgenie, prometheus
	Config     string // JSON configuration
	IsActive   bool   `gorm:"default:false"`
	LastSyncAt *time.Time
	TenantID   uint `gorm:"not null"`
}

// SystemMetric represents system performance metrics
type SystemMetric struct {
	gorm.Model
	ServiceID uint `gorm:"not null"`
	Service   Service
	Name      string    `gorm:"not null"` // response_time, uptime, etc.
	Value     float64   `gorm:"not null"`
	Unit      string    // ms, %, etc.
	Timestamp time.Time `gorm:"not null"`
}

// ThirdPartyService represents external services to monitor
type ThirdPartyService struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string
	StatusURL   string `gorm:"not null"` // URL to fetch status from
	Status      string `gorm:"default:'operational'"`
	LastCheckAt time.Time
	IsActive    bool `gorm:"default:true"`
}

// PrivatePage represents access-controlled status pages
type PrivatePage struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string
	AccessKey   string     `gorm:"uniqueIndex;not null"`
	IsActive    bool       `gorm:"default:true"`
	Services    []*Service `gorm:"many2many:private_page_services;"`
	TenantID    uint       `gorm:"not null"`
}

// SaaS Models for Multi-Tenant Architecture

// Tenant represents a customer organization
type Tenant struct {
	gorm.Model
	Name           string `gorm:"not null"`
	Slug           string `gorm:"uniqueIndex;not null"` // URL-friendly identifier
	Domain         string `gorm:"uniqueIndex"`          // Custom domain
	Subdomain      string `gorm:"uniqueIndex"`          // Subdomain for status page
	Status         string `gorm:"default:'active'"`     // active, suspended, cancelled
	Plan           string `gorm:"default:'free'"`       // free, pro, enterprise
	BillingEmail   string
	ContactEmail   string
	LogoURL        string
	PrimaryColor   string `gorm:"default:'#0052cc'"`
	SecondaryColor string `gorm:"default:'#f4f5f7'"`
	CustomCSS      string
	FooterText     string
	IsActive       bool   `gorm:"default:true"`
	SubscriptionID string // External billing system ID
	CreatedBy      uint   // User who created this tenant
}

// SubscriptionPlan represents available subscription plans
type SubscriptionPlan struct {
	gorm.Model
	Name            string `gorm:"not null"`
	Slug            string `gorm:"uniqueIndex;not null"`
	Description     string
	Price           float64 `gorm:"not null"` // Monthly price in USD
	Currency        string  `gorm:"default:'USD'"`
	BillingInterval string  `gorm:"default:'monthly'"` // monthly, yearly
	MaxServices     int     `gorm:"default:5"`
	MaxMonitors     int     `gorm:"default:10"`
	MaxSubscribers  int     `gorm:"default:100"`
	MaxIncidents    int     `gorm:"default:50"` // Per month
	MaxMaintenance  int     `gorm:"default:20"` // Per month
	CustomDomain    bool    `gorm:"default:false"`
	WhiteLabel      bool    `gorm:"default:false"`
	API             bool    `gorm:"default:false"`
	Integrations    bool    `gorm:"default:false"`
	Analytics       bool    `gorm:"default:false"`
	Support         string  `gorm:"default:'email'"` // email, chat, phone
	IsActive        bool    `gorm:"default:true"`
	Features        string  // JSON array of feature flags
}

// Subscription represents a tenant's subscription
type Subscription struct {
	gorm.Model
	TenantID           uint             `gorm:"not null"`
	Tenant             Tenant           `gorm:"foreignKey:TenantID"`
	PlanID             uint             `gorm:"not null"`
	Plan               SubscriptionPlan `gorm:"foreignKey:PlanID"`
	Status             string           `gorm:"default:'active'"` // active, cancelled, past_due, trialing
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CancelAtPeriodEnd  bool `gorm:"default:false"`
	CancelledAt        *time.Time
	TrialStart         *time.Time
	TrialEnd           *time.Time
	ExternalID         string // Stripe subscription ID
	ExternalCustomerID string // Stripe customer ID
	Metadata           string // JSON metadata
}

// FeatureFlag represents feature toggles for tenants
type FeatureFlag struct {
	gorm.Model
	TenantID  uint   `gorm:"not null"`
	Feature   string `gorm:"not null"` // feature name
	IsEnabled bool   `gorm:"default:false"`
	Config    string // JSON configuration for the feature
}

// SaaSFeatureAvailability controls which features are available to tenants
type SaaSFeatureAvailability struct {
	gorm.Model
	Feature     string `gorm:"not null;unique"` // feature name
	IsAvailable bool   `gorm:"default:false"`   // whether this feature is available to any tenant
	Description string // human-readable description
	Config      string // JSON configuration for the feature
}

// Feature flag constants
const (
	FeaturePerServiceGraphs  = "per_service_graphs"
	FeatureCustomDomains     = "custom_domains"
	FeatureAdvancedAnalytics = "advanced_analytics"
	FeatureSSO               = "sso"
	FeatureAPI               = "api_access"
)

// BillingEvent represents billing-related events
type BillingEvent struct {
	gorm.Model
	TenantID    uint   `gorm:"not null"`
	Tenant      Tenant `gorm:"foreignKey:TenantID"`
	EventType   string `gorm:"not null"` // subscription_created, payment_succeeded, etc.
	Amount      float64
	Currency    string
	ExternalID  string // External system event ID
	Metadata    string // JSON metadata
	ProcessedAt time.Time
}

// UsageMetrics represents usage tracking for billing
type UsageMetrics struct {
	gorm.Model
	TenantID         uint      `gorm:"not null"`
	Tenant           Tenant    `gorm:"foreignKey:TenantID"`
	Date             time.Time `gorm:"not null"`
	ServicesCount    int       `gorm:"default:0"`
	MonitorsCount    int       `gorm:"default:0"`
	SubscribersCount int       `gorm:"default:0"`
	IncidentsCount   int       `gorm:"default:0"`
	MaintenanceCount int       `gorm:"default:0"`
	APIRequests      int       `gorm:"default:0"`
	PageViews        int       `gorm:"default:0"`
}

// AdminSettings represents global admin settings
type AdminSettings struct {
	gorm.Model
	Key         string `gorm:"uniqueIndex;not null"`
	Value       string
	Description string
	Type        string `gorm:"default:'string'"` // string, boolean, number, json
	IsPublic    bool   `gorm:"default:false"`    // Can be accessed by tenants
}

// SystemNotification represents system-wide notifications
type SystemNotification struct {
	gorm.Model
	Title         string `gorm:"not null"`
	Message       string `gorm:"not null"`
	Type          string `gorm:"default:'info'"` // info, warning, error, success
	IsActive      bool   `gorm:"default:true"`
	StartAt       time.Time
	EndAt         *time.Time
	TargetTenants string // JSON array of tenant IDs, empty means all
}

// APIUsage represents API usage tracking for analytics
type APIUsage struct {
	ID           uint    `gorm:"primaryKey"`
	TenantID     uint    `gorm:"not null"`
	Endpoint     string  `gorm:"not null"`
	Method       string  `gorm:"not null"`
	ResponseTime float64 `gorm:"not null"` // in milliseconds
	StatusCode   int     `gorm:"not null"`
	UserAgent    string
	IPAddress    string
	Timestamp    time.Time `gorm:"not null"`
	CreatedAt    time.Time
}

// Billing Models

// BillingCustomer represents a customer in the billing system
type BillingCustomer struct {
	ID        uint   `gorm:"primaryKey"`
	TenantID  uint   `gorm:"not null"`
	Email     string `gorm:"not null"`
	Name      string `gorm:"not null"`
	Status    string `gorm:"default:'active'"` // active, inactive, suspended
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BillingSubscription represents a subscription in the billing system
type BillingSubscription struct {
	ID                 uint   `gorm:"primaryKey"`
	TenantID           uint   `gorm:"not null"`
	PlanID             uint   `gorm:"not null"`
	CustomerID         string `gorm:"not null"`         // External customer ID
	Status             string `gorm:"default:'active'"` // active, cancelled, past_due, trialing
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	Price              float64 `gorm:"not null"`
	Currency           string  `gorm:"default:'USD'"`
	CancelledAt        *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// BillingInvoice represents an invoice in the billing system
type BillingInvoice struct {
	ID             uint    `gorm:"primaryKey"`
	TenantID       uint    `gorm:"not null"`
	SubscriptionID uint    `gorm:"not null"`
	Amount         float64 `gorm:"not null"`
	Currency       string  `gorm:"default:'USD'"`
	Status         string  `gorm:"default:'pending'"` // pending, paid, overdue, cancelled
	ExternalID     string  // External invoice ID (e.g., Stripe invoice ID)
	DueDate        time.Time
	PaidAt         *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// BillingPayment represents a payment in the billing system
type BillingPayment struct {
	ID            uint    `gorm:"primaryKey"`
	TenantID      uint    `gorm:"not null"`
	InvoiceID     uint    `gorm:"not null"`
	Amount        float64 `gorm:"not null"`
	PaymentMethod string  `gorm:"not null"`          // stripe, paypal, bank_transfer
	ExternalID    string  `gorm:"not null"`          // External payment ID
	Status        string  `gorm:"default:'pending'"` // pending, completed, failed, refunded
	ProcessedAt   time.Time
	CreatedAt     time.Time
}

// BillingRefund represents a refund in the billing system
type BillingRefund struct {
	ID          uint    `gorm:"primaryKey"`
	PaymentID   uint    `gorm:"not null"`
	Amount      float64 `gorm:"not null"`
	Status      string  `gorm:"default:'pending'"` // pending, processed, failed
	ExternalID  string  `gorm:"not null"`          // External refund ID
	ProcessedAt time.Time
	CreatedAt   time.Time
}

// PaymentEvent represents a payment event for analytics
type PaymentEvent struct {
	ID        uint      `gorm:"primaryKey"`
	TenantID  uint      `gorm:"not null"`
	EventType string    `gorm:"not null"` // created, completed, failed, refunded
	PaymentID string    `gorm:"not null"`
	Amount    float64   `gorm:"not null"`
	Currency  string    `gorm:"not null"`
	Method    string    `gorm:"not null"` // card, upi, netbanking, wallet
	Gateway   string    `gorm:"not null"` // stripe, razorpay, payu, paypal
	Timestamp time.Time `gorm:"not null"`
	Metadata  string    // JSON metadata
	CreatedAt time.Time
}

// PaymentRetry represents a payment retry attempt
type PaymentRetry struct {
	ID           uint      `gorm:"primaryKey"`
	PaymentID    string    `gorm:"not null"`
	ScheduledAt  time.Time `gorm:"not null"`
	Status       string    `gorm:"default:'scheduled'"` // scheduled, processing, completed, failed, cancelled
	Attempts     int       `gorm:"default:0"`
	ErrorMessage string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Notification represents a notification in the system
type Notification struct {
	ID            uint   `gorm:"primaryKey"`
	Type          string `gorm:"not null"` // payment_success, payment_failure, refund_processed, subscription_created
	Title         string `gorm:"not null"`
	Message       string `gorm:"not null"`
	RecipientID   string `gorm:"not null"`
	RecipientType string `gorm:"not null"` // tenant, user, admin
	Data          string // JSON data
	Status        string `gorm:"default:'pending'"` // pending, sent, failed
	SentAt        *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// EmailNotification represents an email notification
type EmailNotification struct {
	ID             uint   `gorm:"primaryKey"`
	NotificationID uint   `gorm:"not null"`
	RecipientEmail string `gorm:"not null"`
	Subject        string `gorm:"not null"`
	Body           string `gorm:"not null"`
	Status         string `gorm:"default:'pending'"` // pending, sent, failed
	SentAt         *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// InvoiceItem represents an item on an invoice
type InvoiceItem struct {
	ID          uint    `gorm:"primaryKey"`
	InvoiceID   uint    `gorm:"not null"`
	Description string  `gorm:"not null"`
	Quantity    int     `gorm:"not null"`
	UnitPrice   float64 `gorm:"not null"`
	TotalPrice  float64 `gorm:"not null"`
	TaxRate     float64 `gorm:"default:0"`
	TaxAmount   float64 `gorm:"default:0"`
	CreatedAt   time.Time
}

// Advanced Monitoring Models

// HealthCheck represents a health check result for a service
type HealthCheck struct {
	ID           uint      `gorm:"primaryKey"`
	ServiceID    uint      `gorm:"not null"`
	TenantID     uint      `gorm:"not null"`
	Status       string    `gorm:"not null"` // up, down, degraded
	StatusCode   int       // HTTP status code
	ResponseTime int64     // Response time in milliseconds
	Error        string    // Error message if any
	CheckedAt    time.Time `gorm:"not null"`
	CreatedAt    time.Time
}

// Alert represents a monitoring alert rule
type Alert struct {
	ID             uint   `gorm:"primaryKey"`
	TenantID       uint   `gorm:"not null"`
	Name           string `gorm:"not null"`
	Description    string
	Query          string  `gorm:"not null"` // Prometheus query
	Condition      string  `gorm:"not null"` // greater_than, less_than, equal_to, not_equal_to
	Threshold      float64 `gorm:"not null"`
	Severity       string  `gorm:"default:'medium'"` // low, medium, high, critical
	IsActive       bool    `gorm:"default:true"`
	IsFiring       bool    `gorm:"default:false"`
	CreateIncident bool    `gorm:"default:false"`
	LastFiredAt    *time.Time
	LastResolvedAt *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// UptimeStats represents uptime statistics for a service
type UptimeStats struct {
	ID        uint      `gorm:"primaryKey"`
	ServiceID uint      `gorm:"not null"`
	TenantID  uint      `gorm:"not null"`
	Period    int       `gorm:"not null"` // Number of days
	Uptime    float64   `gorm:"not null"` // Uptime percentage
	Downtime  float64   `gorm:"not null"` // Downtime percentage
	Checks    int       `gorm:"not null"` // Total number of checks
	StartDate time.Time `gorm:"not null"`
	EndDate   time.Time `gorm:"not null"`
	CreatedAt time.Time
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
