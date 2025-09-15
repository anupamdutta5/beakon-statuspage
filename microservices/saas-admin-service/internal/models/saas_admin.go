// Package models provides data models for the SaaS Admin Service.
package models

import (
	"time"

	"gorm.io/gorm"
)

// Platform represents the SaaS platform configuration.
type Platform struct {
	ID           uint           `gorm:"primarykey" json:"id"`
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
	ID              uint           `gorm:"primarykey" json:"id"`
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
	Order           int            `gorm:"default:0" json:"order"`    // Display order
	Features        string         `gorm:"type:text" json:"features"` // JSON array of features
	Limits          string         `gorm:"type:text" json:"limits"`   // JSON object of limits
	Metadata        string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// PricingTier represents a pricing tier with different billing intervals.
type PricingTier struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	PlanID          uint           `gorm:"not null;index" json:"plan_id"`
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
	PlanID    uint           `gorm:"not null;index" json:"plan_id"`
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

// TableName returns the table name for SaaSStats.
func (SaaSStats) TableName() string {
	return "saas_stats"
}
