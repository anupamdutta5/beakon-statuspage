// Package models provides data models for the SaaS Admin Service.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Platform represents the SaaS platform configuration.
type Platform struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name         string         `gorm:"not null;uniqueIndex" json:"name"`
	URL          string         `gorm:"not null" json:"url"`
	Description  string         `gorm:"type:text" json:"description"`
	Version      string         `gorm:"not null" json:"version"`
	Status       string         `gorm:"default:active" json:"status"` // active, maintenance, disabled
	AdminEmail   string         `gorm:"not null" json:"admin_email"`
	SupportEmail string         `gorm:"not null" json:"support_email"`
	Metadata     string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// SaaSPlan represents a subscription plan.
type SaaSPlan struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name            string         `gorm:"not null;uniqueIndex" json:"name"`
	Slug            string         `gorm:"not null;uniqueIndex" json:"slug"`
	Description     string         `gorm:"type:text" json:"description"`
	Price           float64        `gorm:"not null" json:"price"`
	Currency        string         `gorm:"default:USD" json:"currency"`
	BillingInterval string         `gorm:"default:monthly" json:"billing_interval"` // monthly, yearly
	MaxTenants      int            `gorm:"default:1" json:"max_tenants"`
	MaxUsers        int            `gorm:"default:5" json:"max_users"`
	MaxServices     int            `gorm:"default:10" json:"max_services"`
	MaxMonitors     int            `gorm:"default:50" json:"max_monitors"`
	MaxSubscribers  int            `gorm:"default:1000" json:"max_subscribers"`
	MaxIncidents    int            `gorm:"default:100" json:"max_incidents"`
	MaxMaintenance  int            `gorm:"default:50" json:"max_maintenance"`
	CustomDomain    bool           `gorm:"default:false" json:"custom_domain"`
	WhiteLabel      bool           `gorm:"default:false" json:"white_label"`
	API             bool           `gorm:"default:false" json:"api"`
	Integrations    bool           `gorm:"default:false" json:"integrations"`
	Analytics       bool           `gorm:"default:false" json:"analytics"`
	Support         string         `gorm:"default:email" json:"support"` // email, chat, phone
	IsActive        bool           `gorm:"default:true;index" json:"is_active"`
	IsPublic        bool           `gorm:"default:true;index" json:"is_public"`
	IsPopular       bool           `gorm:"default:false;index" json:"is_popular"` // For frontend display
	ButtonText      string         `gorm:"default:Get Started" json:"button_text"`
	ButtonURL       string         `gorm:"default:/signup" json:"button_url"`
	DisplayOrder    int            `gorm:"default:0" json:"display_order"` // Display order
	Features        string         `gorm:"type:text" json:"features"`      // JSON array of features
	Limits          string         `gorm:"type:text" json:"limits"`        // JSON object of limits
	Metadata        string         `gorm:"type:text" json:"metadata"`      // JSON string for additional data

	// Relationships
	PricingTiers []PricingTier `gorm:"foreignKey:PlanID" json:"pricing_tiers,omitempty"`
	PlanFeatures []PlanFeature `gorm:"foreignKey:PlanID" json:"plan_features,omitempty"`
}

// PricingTier represents a pricing tier with different billing intervals.
type PricingTier struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	PlanID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"plan_id"`
	Plan            SaaSPlan       `gorm:"foreignKey:PlanID" json:"plan"`
	BillingInterval string         `gorm:"not null" json:"billing_interval"` // monthly, yearly
	Price           float64        `gorm:"not null" json:"price"`
	Currency        string         `gorm:"default:USD" json:"currency"`
	DiscountPercent float64        `gorm:"default:0" json:"discount_percent"` // For yearly discounts
	IsActive        bool           `gorm:"default:true;index" json:"is_active"`
	Metadata        string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// PricingFeature represents individual features that can be assigned to plans.
type PricingFeature struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Category    string         `gorm:"not null;index" json:"category"` // core, advanced, enterprise
	Icon        string         `json:"icon"`                           // Icon class or URL
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	Order       int            `gorm:"default:0" json:"order"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// PlanFeature represents the relationship between plans and features.
type PlanFeature struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	PlanID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"plan_id"`
	Plan      SaaSPlan       `gorm:"foreignKey:PlanID" json:"plan"`
	FeatureID uint           `gorm:"not null;index" json:"feature_id"`
	Feature   PricingFeature `gorm:"foreignKey:FeatureID" json:"feature"`
	IsEnabled bool           `gorm:"default:true;index" json:"is_enabled"`
	Order     int            `gorm:"default:0" json:"order"`
	Metadata  string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// SaaSFeature represents a platform feature.
type SaaSFeature struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name        string         `gorm:"not null;uniqueIndex" json:"name"`
	Slug        string         `gorm:"not null;uniqueIndex" json:"slug"`
	Description string         `gorm:"type:text" json:"description"`
	Category    string         `gorm:"not null;index" json:"category"` // core, advanced, enterprise
	IsEnabled   bool           `gorm:"default:true" json:"is_enabled"`
	IsPublic    bool           `gorm:"default:true" json:"is_public"`
	Config      string         `gorm:"type:text" json:"config"`   // JSON configuration
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// SaaSFeatureFlag represents global feature flags.
type SaaSFeatureFlag struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name        string         `gorm:"not null;uniqueIndex" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	IsEnabled   bool           `gorm:"default:false" json:"is_enabled"`
	Config      string         `gorm:"type:text" json:"config"`   // JSON configuration
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// SaaSAdminUser represents SaaS admin users.
type SaaSAdminUser struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Email       string         `gorm:"not null;uniqueIndex" json:"email"`
	Username    string         `gorm:"not null;uniqueIndex" json:"username"`
	Password    string         `gorm:"not null" json:"-"` // Hidden from JSON
	FirstName   string         `gorm:"not null" json:"first_name"`
	LastName    string         `gorm:"not null" json:"last_name"`
	Role        string         `gorm:"default:admin" json:"role"`    // super_admin, admin, support
	Status      string         `gorm:"default:active" json:"status"` // active, inactive, suspended
	LastLoginAt *time.Time     `json:"last_login_at"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// SaaSNotification represents platform notifications.
type SaaSNotification struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Title       string         `gorm:"not null" json:"title"`
	Message     string         `gorm:"type:text;not null" json:"message"`
	Type        string         `gorm:"not null;index" json:"type"`     // info, warning, error, success
	Priority    string         `gorm:"default:normal" json:"priority"` // low, normal, high, critical
	Status      string         `gorm:"default:active" json:"status"`   // active, inactive, dismissed
	IsGlobal    bool           `gorm:"default:false" json:"is_global"`
	TargetRoles string         `gorm:"type:text" json:"target_roles"` // JSON array of roles
	ExpiresAt   *time.Time     `json:"expires_at"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// SaaSActivity represents platform activity logs.
type SaaSActivity struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	User        SaaSAdminUser  `gorm:"foreignKey:UserID" json:"user"`
	Action      string         `gorm:"not null;index" json:"action"`
	Resource    string         `gorm:"not null;index" json:"resource"`
	ResourceID  string         `gorm:"index" json:"resource_id"`
	Description string         `gorm:"type:text" json:"description"`
	IPAddress   string         `json:"ip_address"`
	UserAgent   string         `gorm:"type:text" json:"user_agent"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// SaaSBackup represents platform backups.
type SaaSBackup struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name      string         `gorm:"not null" json:"name"`
	Type      string         `gorm:"not null;index" json:"type"`    // full, incremental, differential
	Status    string         `gorm:"default:pending" json:"status"` // pending, in_progress, completed, failed
	Size      int64          `json:"size"`
	Location  string         `gorm:"not null" json:"location"`
	Checksum  string         `json:"checksum"`
	ExpiresAt *time.Time     `json:"expires_at"`
	Metadata  string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// SaaSStats represents platform statistics.
type SaaSStats struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	TotalTenants   int       `json:"total_tenants"`
	ActiveTenants  int       `json:"active_tenants"`
	TotalUsers     int       `json:"total_users"`
	ActiveUsers    int       `json:"active_users"`
	TotalRevenue   float64   `json:"total_revenue"`
	MonthlyRevenue float64   `json:"monthly_revenue"`
	TotalPlans     int       `json:"total_plans"`
	ActivePlans    int       `json:"active_plans"`
	TotalFeatures  int       `json:"total_features"`
	ActiveFeatures int       `json:"active_features"`
	LastUpdated    time.Time `json:"last_updated"`
}

// TableName returns the table name for Platform.
func (Platform) TableName() string {
	return "platforms"
}

// TableName returns the table name for SaaSPlan.
func (SaaSPlan) TableName() string {
	return "saas_plans"
}

// TableName returns the table name for PricingTier.
func (PricingTier) TableName() string {
	return "pricing_tiers"
}

// TableName returns the table name for PricingFeature.
func (PricingFeature) TableName() string {
	return "pricing_features"
}

// TableName returns the table name for PlanFeature.
func (PlanFeature) TableName() string {
	return "plan_features"
}

// TableName returns the table name for SaaSFeature.
func (SaaSFeature) TableName() string {
	return "saas_features"
}

// TableName returns the table name for SaaSFeatureFlag.
func (SaaSFeatureFlag) TableName() string {
	return "saas_feature_flags"
}

// TableName returns the table name for SaaSAdminUser.
func (SaaSAdminUser) TableName() string {
	return "saas_admin_users"
}

// TableName returns the table name for SaaSNotification.
func (SaaSNotification) TableName() string {
	return "saas_notifications"
}

// TableName returns the table name for SaaSActivity.
func (SaaSActivity) TableName() string {
	return "saas_activities"
}

// TableName returns the table name for SaaSBackup.
func (SaaSBackup) TableName() string {
	return "saas_backups"
}

// SaaSIncident represents an incident on the status page.
type SaaSIncident struct {
	ID              uuid.UUID           `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	DeletedAt       gorm.DeletedAt      `gorm:"index" json:"deleted_at,omitempty"`
	Title           string              `gorm:"not null" json:"title"`
	Description     string              `gorm:"type:text" json:"description"`
	Status          string              `gorm:"not null;index" json:"status"` // investigating, identified, monitoring, resolved
	Impact          string              `gorm:"not null;index" json:"impact"` // none, minor, major, critical
	ComponentStatus string              `gorm:"default:operational" json:"component_status"` // operational, degraded_performance, partial_outage, major_outage
	StartedAt       time.Time           `gorm:"not null" json:"started_at"`
	ResolvedAt      *time.Time          `json:"resolved_at"`
	TenantID        string              `gorm:"not null;index" json:"tenant_id"`
	CreatedBy       string              `json:"created_by"`
	UpdatedBy       string              `json:"updated_by"`
	IsVisible       bool                `gorm:"default:true" json:"is_visible"`
	ExternalID      string              `gorm:"index" json:"external_id"` // For integration with external systems
	Metadata        string              `gorm:"type:text" json:"metadata"` // JSON string for additional data

	// Relationships
	Updates []SaaSIncidentUpdate `gorm:"foreignKey:IncidentID" json:"updates,omitempty"`
}

// SaaSIncidentUpdate represents an update to an incident.
type SaaSIncidentUpdate struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	IncidentID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"incident_id"`
	Incident    SaaSIncident   `gorm:"foreignKey:IncidentID" json:"incident"`
	Status      string         `gorm:"not null" json:"status"` // investigating, identified, monitoring, resolved
	Message     string         `gorm:"type:text;not null" json:"message"`
	CreatedBy   string         `json:"created_by"`
	IsPublic    bool           `gorm:"default:true" json:"is_public"`
	UpdateType  string         `gorm:"default:status_update" json:"update_type"` // status_update, postmortem
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// SaaSMaintenanceWindow represents scheduled maintenance.
type SaaSMaintenanceWindow struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Title           string         `gorm:"not null" json:"title"`
	Description     string         `gorm:"type:text" json:"description"`
	Status          string         `gorm:"not null;index" json:"status"` // scheduled, in_progress, completed, cancelled
	Impact          string         `gorm:"not null" json:"impact"` // none, minor, major, critical
	ScheduledFor    time.Time      `gorm:"not null" json:"scheduled_for"`
	EstimatedEnd    time.Time      `json:"estimated_end"`
	ActualStart     *time.Time     `json:"actual_start"`
	ActualEnd       *time.Time     `json:"actual_end"`
	TenantID        string         `gorm:"not null;index" json:"tenant_id"`
	CreatedBy       string         `json:"created_by"`
	UpdatedBy       string         `json:"updated_by"`
	IsVisible       bool           `gorm:"default:true" json:"is_visible"`
	NotifySubscribers bool         `gorm:"default:true" json:"notify_subscribers"`
	Metadata        string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// SaaSService represents a service/component being monitored.
type SaaSService struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Status      string         `gorm:"not null;index" json:"status"` // operational, degraded_performance, partial_outage, major_outage
	Category    string         `gorm:"index" json:"category"` // core, api, web, mobile, infrastructure
	Position    int            `gorm:"default:0" json:"position"` // Display order
	TenantID    string         `gorm:"not null;index" json:"tenant_id"`
	IsVisible   bool           `gorm:"default:true" json:"is_visible"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for SaaSStats.
func (SaaSStats) TableName() string {
	return "saas_stats"
}

// TableName returns the table name for SaaSIncident.
func (SaaSIncident) TableName() string {
	return "saas_incidents"
}

// TableName returns the table name for SaaSIncidentUpdate.
func (SaaSIncidentUpdate) TableName() string {
	return "saas_incident_updates"
}

// TableName returns the table name for SaaSMaintenanceWindow.
func (SaaSMaintenanceWindow) TableName() string {
	return "saas_maintenance_windows"
}

// TableName returns the table name for SaaSService.
func (SaaSService) TableName() string {
	return "saas_services"
}

// SaaSIntegration represents a notification integration configuration
type SaaSIntegration struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    string         `gorm:"not null;index" json:"tenant_id"`
	Name        string         `gorm:"not null" json:"name"`
	Type        string         `gorm:"not null;index" json:"type"` // slack, discord, teams, email, webhook, sms
	Status      string         `gorm:"default:active" json:"status"` // active, inactive, error
	Config      string         `gorm:"type:text" json:"config"` // JSON configuration for the integration
	LastUsed    *time.Time     `json:"last_used"`
	LastError   string         `gorm:"type:text" json:"last_error"`
	EventTypes  string         `gorm:"type:text" json:"event_types"` // JSON array of subscribed event types
	Enabled     bool           `gorm:"default:true" json:"enabled"`
	Description string         `gorm:"type:text" json:"description"`
}

// SaaSWebhook represents a webhook configuration
type SaaSWebhook struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    string         `gorm:"not null;index" json:"tenant_id"`
	Name        string         `gorm:"not null" json:"name"`
	URL         string         `gorm:"not null" json:"url"`
	Secret      string         `json:"secret"` // Optional webhook secret
	EventTypes  string         `gorm:"type:text" json:"event_types"` // JSON array of subscribed event types
	Headers     string         `gorm:"type:text" json:"headers"` // JSON object of custom headers
	Method      string         `gorm:"default:POST" json:"method"` // HTTP method
	Status      string         `gorm:"default:active" json:"status"` // active, inactive, error
	LastStatus  int            `json:"last_status"` // Last HTTP response status
	LastError   string         `gorm:"type:text" json:"last_error"`
	LastSent    *time.Time     `json:"last_sent"`
	Enabled     bool           `gorm:"default:true" json:"enabled"`
	Retries     int            `gorm:"default:3" json:"retries"`
	Timeout     int            `gorm:"default:30" json:"timeout"` // seconds
}

// SaaSSubscriber represents a status page subscriber
type SaaSSubscriber struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID      string         `gorm:"not null;index" json:"tenant_id"`
	Email         string         `gorm:"not null;index" json:"email"`
	Phone         string         `json:"phone"`
	Status        string         `gorm:"default:active" json:"status"` // active, inactive, unsubscribed
	EventTypes    string         `gorm:"type:text" json:"event_types"` // JSON array of subscribed event types
	Components    string         `gorm:"type:text" json:"components"` // JSON array of subscribed component IDs
	VerifiedAt    *time.Time     `json:"verified_at"`
	UnsubscribeToken string      `gorm:"uniqueIndex" json:"unsubscribe_token"`
	Preferences   string         `gorm:"type:text" json:"preferences"` // JSON object of notification preferences
}

// SaaSComponent represents a service component that can have status updates
type SaaSComponent struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    string         `gorm:"not null;index" json:"tenant_id"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Status      string         `gorm:"default:operational" json:"status"` // operational, degraded_performance, partial_outage, major_outage, under_maintenance
	GroupName   string         `json:"group_name"` // Component group for organization
	Order       int            `gorm:"default:0" json:"order"` // Display order
	Visible     bool           `gorm:"default:true" json:"visible"`
	ShowUptime  bool           `gorm:"default:true" json:"show_uptime"`
	Uptime      float64        `gorm:"default:100.0" json:"uptime"` // Uptime percentage
	Link        string         `json:"link"` // Optional link to component details
}

// TableName returns the table name for SaaSIntegration.
func (SaaSIntegration) TableName() string {
	return "saas_integrations"
}

// TableName returns the table name for SaaSWebhook.
func (SaaSWebhook) TableName() string {
	return "saas_webhooks"
}

// TableName returns the table name for SaaSSubscriber.
func (SaaSSubscriber) TableName() string {
	return "saas_subscribers"
}

// TableName returns the table name for SaaSComponent.
func (SaaSComponent) TableName() string {
	return "saas_components"
}

// SaaSTenant represents a tenant in the SaaS platform.
type SaaSTenant struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name         string         `gorm:"not null" json:"name"`
	Slug         string         `gorm:"uniqueIndex;not null" json:"slug"`
	Domain       string         `gorm:"uniqueIndex" json:"domain"`
	Subdomain    string         `gorm:"uniqueIndex" json:"subdomain"`
	ContactEmail string         `gorm:"not null" json:"contact_email"`
	BillingEmail string         `json:"billing_email"`
	PlanID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"plan_id"`
	Plan         SaaSPlan       `gorm:"foreignKey:PlanID" json:"plan"`
	Status       string         `gorm:"default:active" json:"status"` // active, inactive, suspended, cancelled
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	Settings     string         `gorm:"type:text" json:"settings"` // JSON string
	Branding     string         `gorm:"type:text" json:"branding"` // JSON string
	Features     string         `gorm:"type:text" json:"features"` // JSON string
	Metadata     string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for SaaSTenant.
func (SaaSTenant) TableName() string {
	return "tenants"
}
