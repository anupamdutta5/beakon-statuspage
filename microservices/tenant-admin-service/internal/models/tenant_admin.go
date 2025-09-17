// Package models provides data models for the Tenant Admin Service.
package models

import (
	"time"

	"gorm.io/gorm"
)

// Tenant represents a tenant in the system.
type Tenant struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name         string         `gorm:"not null" json:"name"`
	Slug         string         `gorm:"uniqueIndex;not null" json:"slug"`
	Domain       string         `gorm:"uniqueIndex" json:"domain"`
	Subdomain    string         `gorm:"uniqueIndex" json:"subdomain"`
	ContactEmail string         `gorm:"not null" json:"contact_email"`
	BillingEmail string         `json:"billing_email"`
	Plan         string         `gorm:"default:free" json:"plan"`
	Status       string         `gorm:"default:active" json:"status"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	Settings     string         `gorm:"type:text" json:"settings"` // JSON string
	Branding     string         `gorm:"type:text" json:"branding"` // JSON string
	Features     string         `gorm:"type:text" json:"features"` // JSON string
}

// TenantBranding represents tenant branding information.
type TenantBranding struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID  uint           `gorm:"uniqueIndex;not null" json:"tenant_id"`
	Tenant    Tenant         `gorm:"foreignKey:TenantID" json:"tenant"`

	// Branding Elements
	LogoURL         string `json:"logo_url"`
	FaviconURL      string `json:"favicon_url"`
	PrimaryColor    string `json:"primary_color"`
	SecondaryColor  string `json:"secondary_color"`
	AccentColor     string `json:"accent_color"`
	BackgroundColor string `json:"background_color"`
	TextColor       string `json:"text_color"`

	// Typography
	FontFamily string `json:"font_family"`
	FontSize   string `json:"font_size"`
	FontWeight string `json:"font_weight"`

	// Layout
	Layout     string `json:"layout"` // default, minimal, custom
	ShowLogo   bool   `json:"show_logo"`
	ShowFooter bool   `json:"show_footer"`
	FooterText string `json:"footer_text"`

	// Custom Content
	CustomHeader    string `gorm:"type:text" json:"custom_header"`
	CustomFooter    string `gorm:"type:text" json:"custom_footer"`
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
	MetaKeywords    string `json:"meta_keywords"`
}

// TenantAdmin represents a tenant administrator.
type TenantAdmin struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	Role        string         `gorm:"not null;index" json:"role"`   // owner, admin, manager, viewer
	Status      string         `gorm:"default:active" json:"status"` // active, inactive, suspended
	Permissions string         `gorm:"type:text" json:"permissions"` // JSON array of permissions
	LastLoginAt *time.Time     `json:"last_login_at"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TenantSettings represents tenant-specific settings.
type TenantSettings struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID  uint           `gorm:"not null;uniqueIndex" json:"tenant_id"`
	Settings  string         `gorm:"type:text;not null" json:"settings"` // JSON object of settings
	Version   string         `gorm:"default:1.0.0" json:"version"`
	Status    string         `gorm:"default:active" json:"status"` // active, inactive, draft
	Metadata  string         `gorm:"type:text" json:"metadata"`    // JSON string for additional data
}

// TenantFeatureFlag represents tenant-specific feature flags.
type TenantFeatureFlag struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	Name        string         `gorm:"not null;index" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	IsEnabled   bool           `gorm:"default:false" json:"is_enabled"`
	Config      string         `gorm:"type:text" json:"config"`   // JSON configuration
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TenantUsage represents tenant usage metrics.
type TenantUsage struct {
	ID               uint           `gorm:"primarykey" json:"id"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID         uint           `gorm:"not null;index" json:"tenant_id"`
	Date             time.Time      `gorm:"not null;index" json:"date"`
	UsersCount       int            `gorm:"default:0" json:"users_count"`
	ServicesCount    int            `gorm:"default:0" json:"services_count"`
	MonitorsCount    int            `gorm:"default:0" json:"monitors_count"`
	SubscribersCount int            `gorm:"default:0" json:"subscribers_count"`
	IncidentsCount   int            `gorm:"default:0" json:"incidents_count"`
	MaintenanceCount int            `gorm:"default:0" json:"maintenance_count"`
	APIRequests      int            `gorm:"default:0" json:"api_requests"`
	PageViews        int            `gorm:"default:0" json:"page_views"`
	StorageUsed      int64          `gorm:"default:0" json:"storage_used"`   // in bytes
	BandwidthUsed    int64          `gorm:"default:0" json:"bandwidth_used"` // in bytes
	Metadata         string         `gorm:"type:text" json:"metadata"`       // JSON string for additional data
}

// TenantBilling represents tenant billing information.
type TenantBilling struct {
	ID                 uint           `gorm:"primarykey" json:"id"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID           uint           `gorm:"not null;uniqueIndex" json:"tenant_id"`
	PlanID             uint           `gorm:"not null;index" json:"plan_id"`
	PlanName           string         `gorm:"not null" json:"plan_name"`
	BillingCycle       string         `gorm:"default:monthly" json:"billing_cycle"` // monthly, yearly
	Status             string         `gorm:"default:active" json:"status"`         // active, suspended, cancelled
	CurrentPeriodStart *time.Time     `json:"current_period_start"`
	CurrentPeriodEnd   *time.Time     `json:"current_period_end"`
	TrialStart         *time.Time     `json:"trial_start"`
	TrialEnd           *time.Time     `json:"trial_end"`
	IsTrialActive      bool           `gorm:"default:false" json:"is_trial_active"`
	NextBillingDate    *time.Time     `json:"next_billing_date"`
	Amount             float64        `gorm:"default:0" json:"amount"`
	Currency           string         `gorm:"default:USD" json:"currency"`
	PaymentMethod      string         `json:"payment_method"`
	Metadata           string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TenantNotification represents tenant-specific notifications.
type TenantNotification struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID  uint           `gorm:"not null;index" json:"tenant_id"`
	Title     string         `gorm:"not null" json:"title"`
	Message   string         `gorm:"type:text;not null" json:"message"`
	Type      string         `gorm:"not null;index" json:"type"`     // info, warning, error, success
	Priority  string         `gorm:"default:normal" json:"priority"` // low, normal, high, critical
	Status    string         `gorm:"default:active" json:"status"`   // active, inactive, dismissed
	IsRead    bool           `gorm:"default:false" json:"is_read"`
	ReadAt    *time.Time     `json:"read_at"`
	ExpiresAt *time.Time     `json:"expires_at"`
	Metadata  string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TenantActivity represents tenant activity logs.
type TenantActivity struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	Action      string         `gorm:"not null;index" json:"action"`
	Resource    string         `gorm:"not null;index" json:"resource"`
	ResourceID  string         `gorm:"index" json:"resource_id"`
	Description string         `gorm:"type:text" json:"description"`
	IPAddress   string         `json:"ip_address"`
	UserAgent   string         `gorm:"type:text" json:"user_agent"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TenantBackup represents tenant backups.
type TenantBackup struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID  uint           `gorm:"not null;index" json:"tenant_id"`
	Name      string         `gorm:"not null" json:"name"`
	Type      string         `gorm:"not null;index" json:"type"`    // full, incremental, differential
	Status    string         `gorm:"default:pending" json:"status"` // pending, in_progress, completed, failed
	Size      int64          `json:"size"`
	Location  string         `gorm:"not null" json:"location"`
	Checksum  string         `json:"checksum"`
	ExpiresAt *time.Time     `json:"expires_at"`
	Metadata  string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TenantStats represents tenant statistics.
type TenantStats struct {
	ID                uint      `gorm:"primarykey" json:"id"`
	CreatedAt         time.Time `json:"created_at"`
	TenantID          uint      `gorm:"not null;uniqueIndex" json:"tenant_id"`
	TotalUsers        int       `json:"total_users"`
	ActiveUsers       int       `json:"active_users"`
	TotalServices     int       `json:"total_services"`
	ActiveServices    int       `json:"active_services"`
	TotalMonitors     int       `json:"total_monitors"`
	ActiveMonitors    int       `json:"active_monitors"`
	TotalSubscribers  int       `json:"total_subscribers"`
	ActiveSubscribers int       `json:"active_subscribers"`
	TotalIncidents    int       `json:"total_incidents"`
	OpenIncidents     int       `json:"open_incidents"`
	TotalMaintenance  int       `json:"total_maintenance"`
	ActiveMaintenance int       `json:"active_maintenance"`
	TotalAPIRequests  int       `json:"total_api_requests"`
	TotalPageViews    int       `json:"total_page_views"`
	StorageUsed       int64     `json:"storage_used"`
	BandwidthUsed     int64     `json:"bandwidth_used"`
	LastUpdated       time.Time `json:"last_updated"`
}

// TableName returns the table name for Tenant.
func (Tenant) TableName() string {
	return "tenants"
}

// TableName returns the table name for TenantBranding.
func (TenantBranding) TableName() string {
	return "tenant_branding"
}

// TableName returns the table name for TenantAdmin.
func (TenantAdmin) TableName() string {
	return "tenant_admins"
}

// TableName returns the table name for TenantSettings.
func (TenantSettings) TableName() string {
	return "tenant_settings"
}

// TableName returns the table name for TenantFeatureFlag.
func (TenantFeatureFlag) TableName() string {
	return "tenant_feature_flags"
}

// TableName returns the table name for TenantUsage.
func (TenantUsage) TableName() string {
	return "tenant_usage"
}

// TableName returns the table name for TenantBilling.
func (TenantBilling) TableName() string {
	return "tenant_billing"
}

// TableName returns the table name for TenantNotification.
func (TenantNotification) TableName() string {
	return "tenant_notifications"
}

// TableName returns the table name for TenantActivity.
func (TenantActivity) TableName() string {
	return "tenant_activities"
}

// TableName returns the table name for TenantBackup.
func (TenantBackup) TableName() string {
	return "tenant_backups"
}

// TableName returns the table name for TenantStats.
func (TenantStats) TableName() string {
	return "tenant_stats"
}

