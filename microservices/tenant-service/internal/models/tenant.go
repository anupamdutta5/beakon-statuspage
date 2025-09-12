// Package models provides data models for the Tenant Service.
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

// TenantSettings represents tenant-specific settings.
type TenantSettings struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	TenantID  uint      `gorm:"uniqueIndex;not null" json:"tenant_id"`
	Tenant    Tenant    `gorm:"foreignKey:TenantID" json:"tenant"`
	
	// General Settings
	Timezone        string `json:"timezone"`
	Language        string `json:"language"`
	DateFormat      string `json:"date_format"`
	TimeFormat      string `json:"time_format"`
	
	// Notification Settings
	EmailNotifications bool `json:"email_notifications"`
	SMSNotifications   bool `json:"sms_notifications"`
	WebhookURL         string `json:"webhook_url"`
	
	// Display Settings
	ShowIncidentHistory bool `json:"show_incident_history"`
	ShowMaintenanceMode bool `json:"show_maintenance_mode"`
	CustomCSS           string `gorm:"type:text" json:"custom_css"`
	CustomJS            string `gorm:"type:text" json:"custom_js"`
	
	// Security Settings
	RequireAuth         bool `json:"require_auth"`
	AllowPublicAccess   bool `json:"allow_public_access"`
	SessionTimeout      int  `json:"session_timeout"`
	
	// API Settings
	APIRateLimit        int    `json:"api_rate_limit"`
	APIKeyRequired      bool   `json:"api_key_required"`
	AllowedOrigins      string `gorm:"type:text" json:"allowed_origins"` // JSON array
}

// TenantBilling represents tenant billing information.
type TenantBilling struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	TenantID  uint      `gorm:"uniqueIndex;not null" json:"tenant_id"`
	Tenant    Tenant    `gorm:"foreignKey:TenantID" json:"tenant"`
	
	// Billing Information
	Plan              string    `json:"plan"`
	BillingCycle      string    `json:"billing_cycle"` // monthly, yearly
	Amount            float64   `json:"amount"`
	Currency          string    `json:"currency"`
	NextBillingDate   time.Time `json:"next_billing_date"`
	LastPaymentDate   *time.Time `json:"last_payment_date"`
	
	// Payment Information
	PaymentMethod     string `json:"payment_method"`
	StripeCustomerID  string `json:"stripe_customer_id"`
	StripeSubscriptionID string `json:"stripe_subscription_id"`
	
	// Usage Tracking
	UsersCount        int `json:"users_count"`
	ComponentsCount   int `json:"components_count"`
	IncidentsCount    int `json:"incidents_count"`
	APIRequestsCount  int `json:"api_requests_count"`
	
	// Limits
	MaxUsers          int `json:"max_users"`
	MaxComponents     int `json:"max_components"`
	MaxIncidents      int `json:"max_incidents"`
	MaxAPIRequests    int `json:"max_api_requests"`
	
	// Status
	Status            string `json:"status"` // active, suspended, cancelled
	IsTrial           bool   `json:"is_trial"`
	TrialEndsAt       *time.Time `json:"trial_ends_at"`
}

// TenantBranding represents tenant branding information.
type TenantBranding struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	TenantID  uint      `gorm:"uniqueIndex;not null" json:"tenant_id"`
	Tenant    Tenant    `gorm:"foreignKey:TenantID" json:"tenant"`
	
	// Branding Elements
	LogoURL           string `json:"logo_url"`
	FaviconURL        string `json:"favicon_url"`
	PrimaryColor      string `json:"primary_color"`
	SecondaryColor    string `json:"secondary_color"`
	AccentColor       string `json:"accent_color"`
	BackgroundColor   string `json:"background_color"`
	TextColor         string `json:"text_color"`
	
	// Typography
	FontFamily        string `json:"font_family"`
	FontSize          string `json:"font_size"`
	FontWeight        string `json:"font_weight"`
	
	// Layout
	Layout            string `json:"layout"` // default, minimal, custom
	ShowLogo          bool   `json:"show_logo"`
	ShowFooter        bool   `json:"show_footer"`
	FooterText        string `json:"footer_text"`
	
	// Custom Content
	CustomHeader      string `gorm:"type:text" json:"custom_header"`
	CustomFooter      string `gorm:"type:text" json:"custom_footer"`
	MetaTitle         string `json:"meta_title"`
	MetaDescription   string `json:"meta_description"`
	MetaKeywords      string `json:"meta_keywords"`
}

// TenantActivity represents tenant activity logs.
type TenantActivity struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	TenantID  uint      `gorm:"not null;index" json:"tenant_id"`
	Tenant    Tenant    `gorm:"foreignKey:TenantID" json:"tenant"`
	UserID    uint      `json:"user_id"`
	Action    string    `gorm:"not null" json:"action"`
	Resource  string    `json:"resource"`
	ResourceID string   `json:"resource_id"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Metadata  string    `gorm:"type:text" json:"metadata"` // JSON string
}

// TableName returns the table name for Tenant.
func (Tenant) TableName() string {
	return "tenants"
}

// TableName returns the table name for TenantSettings.
func (TenantSettings) TableName() string {
	return "tenant_settings"
}

// TableName returns the table name for TenantBilling.
func (TenantBilling) TableName() string {
	return "tenant_billing"
}

// TableName returns the table name for TenantBranding.
func (TenantBranding) TableName() string {
	return "tenant_branding"
}

// TableName returns the table name for TenantActivity.
func (TenantActivity) TableName() string {
	return "tenant_activities"
}
