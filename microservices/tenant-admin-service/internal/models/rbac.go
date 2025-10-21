// Package models defines RBAC (Role-Based Access Control) data models for the Tenant Admin Service.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role represents a role in the RBAC system
type Role struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	TenantID    uuid.UUID `json:"tenant_id" gorm:"type:uuid;index"`
	Name        string    `json:"name" gorm:"not null;size:100"`
	DisplayName string    `json:"display_name" gorm:"not null;size:255"`
	Description string    `json:"description" gorm:"type:text"`
	IsDefault   bool      `json:"is_default" gorm:"default:false"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
	UserRoles   []UserRole   `json:"user_roles" gorm:"foreignKey:RoleID"`

	// Computed fields
	PermissionCount int64 `json:"permission_count" gorm:"-"`
	UserCount       int64 `json:"user_count" gorm:"-"`
}

// Permission represents a permission in the RBAC system
type Permission struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string    `json:"name" gorm:"not null;size:100;uniqueIndex"`
	DisplayName string    `json:"display_name" gorm:"not null;size:255"`
	Description string    `json:"description" gorm:"type:text"`
	Resource    string    `json:"resource" gorm:"not null;size:100"`
	Action      string    `json:"action" gorm:"not null;size:50"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Roles []Role `json:"roles" gorm:"many2many:role_permissions;"`
}

// UserRole represents the relationship between users and roles
type UserRole struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID     uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	RoleID     uuid.UUID  `json:"role_id" gorm:"type:uuid;not null;index"`
	TenantID   uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	AssignedBy uuid.UUID  `json:"assigned_by" gorm:"type:uuid;not null"`
	AssignedAt time.Time  `json:"assigned_at" gorm:"default:CURRENT_TIMESTAMP"`
	ExpiresAt  *time.Time `json:"expires_at"`
	IsActive   bool       `json:"is_active" gorm:"default:true"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// Relationships
	Role Role `json:"role" gorm:"foreignKey:RoleID"`
}

// Team represents a team within a tenant for grouping users
type Team struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	TenantID    uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	Name        string    `json:"name" gorm:"not null;size:100"`
	Description string    `json:"description" gorm:"type:text"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedBy   uuid.UUID `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	TeamMembers []TeamMember `json:"team_members" gorm:"foreignKey:TeamID"`
	TeamRoles   []TeamRole   `json:"team_roles" gorm:"foreignKey:TeamID"`

	// Computed fields
	MemberCount int64 `json:"member_count" gorm:"-"`
}

// TeamMember represents the relationship between teams and users
type TeamMember struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	TeamID    uuid.UUID `json:"team_id" gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	TenantID  uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	Role      string    `json:"role" gorm:"not null;size:50;default:'member'"` // member, admin, lead
	AddedBy   uuid.UUID `json:"added_by" gorm:"type:uuid;not null"`
	AddedAt   time.Time `json:"added_at" gorm:"default:CURRENT_TIMESTAMP"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Team Team `json:"team" gorm:"foreignKey:TeamID"`
}

// TeamRole represents roles assigned to entire teams
type TeamRole struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	TeamID     uuid.UUID  `json:"team_id" gorm:"type:uuid;not null;index"`
	RoleID     uuid.UUID  `json:"role_id" gorm:"type:uuid;not null;index"`
	TenantID   uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	AssignedBy uuid.UUID  `json:"assigned_by" gorm:"type:uuid;not null"`
	AssignedAt time.Time  `json:"assigned_at" gorm:"default:CURRENT_TIMESTAMP"`
	ExpiresAt  *time.Time `json:"expires_at"`
	IsActive   bool       `json:"is_active" gorm:"default:true"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// Relationships
	Team Team `json:"team" gorm:"foreignKey:TeamID"`
	Role Role `json:"role" gorm:"foreignKey:RoleID"`
}

// AuditLog represents audit trail for RBAC actions
type AuditLog struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	TenantID     uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	UserID       uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Action       string     `json:"action" gorm:"not null;size:100"`
	Resource     string     `json:"resource" gorm:"not null;size:100"`
	ResourceID   *uuid.UUID `json:"resource_id" gorm:"type:uuid;index"`
	Details      string     `json:"details" gorm:"type:text"`
	IPAddress    string     `json:"ip_address" gorm:"size:45"`
	UserAgent    string     `json:"user_agent" gorm:"type:text"`
	Success      bool       `json:"success" gorm:"default:true"`
	ErrorMessage string     `json:"error_message" gorm:"type:text"`
	CreatedAt    time.Time  `json:"created_at"`
}

// Session represents user sessions for tracking active users
type Session struct {
	ID         string    `json:"id" gorm:"primaryKey;size:128"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	TenantID   uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	IPAddress  string    `json:"ip_address" gorm:"size:45"`
	UserAgent  string    `json:"user_agent" gorm:"type:text"`
	LastSeenAt time.Time `json:"last_seen_at" gorm:"column:last_seen;default:CURRENT_TIMESTAMP"`
	ExpiresAt  time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// UserSession represents a refresh token session for long-lived authentication
// Used for JWT refresh token flow (OAuth 2.0 pattern)
type UserSession struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	TenantID  uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Token     string    `gorm:"not null;uniqueIndex;size:255" json:"token"` // Refresh token (base64)
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	IPAddress string    `gorm:"type:text" json:"ip_address"`
	UserAgent string    `gorm:"type:text" json:"user_agent"`
	IsActive  bool      `gorm:"default:true;index" json:"is_active"`
}

// IsValid checks if the user session (refresh token) is still valid
func (us *UserSession) IsValid() bool {
	return us.IsActive && time.Now().Before(us.ExpiresAt)
}

// IsExpired checks if the user session (refresh token) has expired
func (us *UserSession) IsExpired() bool {
	return time.Now().After(us.ExpiresAt)
}

// TableName methods for custom table names
func (Role) TableName() string {
	return "roles"
}

func (Permission) TableName() string {
	return "permissions"
}

func (Session) TableName() string {
	return "sessions"
}

func (UserSession) TableName() string {
	return "user_sessions"
}

func (UserRole) TableName() string {
	return "user_roles"
}

func (Team) TableName() string {
	return "teams"
}

func (TeamMember) TableName() string {
	return "team_members"
}

func (TeamRole) TableName() string {
	return "team_roles"
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

// Default roles and permissions
var DefaultRoles = []Role{
	{
		Name:        "admin",
		DisplayName: "Administrator",
		Description: "Full access to all resources and settings",
		IsDefault:   true,
	},
	{
		Name:        "editor",
		DisplayName: "Editor",
		Description: "Can create and modify content, manage incidents and components",
		IsDefault:   true,
	},
	{
		Name:        "viewer",
		DisplayName: "Viewer",
		Description: "Read-only access to status page and reports",
		IsDefault:   true,
	},
	{
		Name:        "responder",
		DisplayName: "Incident Responder",
		Description: "Can respond to incidents and update component status",
		IsDefault:   true,
	},
}

var DefaultPermissions = []Permission{
	// Component permissions
	{Name: "components.read", DisplayName: "View Components", Resource: "components", Action: "read"},
	{Name: "components.create", DisplayName: "Create Components", Resource: "components", Action: "create"},
	{Name: "components.update", DisplayName: "Update Components", Resource: "components", Action: "update"},
	{Name: "components.delete", DisplayName: "Delete Components", Resource: "components", Action: "delete"},

	// Incident permissions
	{Name: "incidents.read", DisplayName: "View Incidents", Resource: "incidents", Action: "read"},
	{Name: "incidents.create", DisplayName: "Create Incidents", Resource: "incidents", Action: "create"},
	{Name: "incidents.update", DisplayName: "Update Incidents", Resource: "incidents", Action: "update"},
	{Name: "incidents.delete", DisplayName: "Delete Incidents", Resource: "incidents", Action: "delete"},

	// Maintenance permissions
	{Name: "maintenance.read", DisplayName: "View Maintenance", Resource: "maintenance", Action: "read"},
	{Name: "maintenance.create", DisplayName: "Create Maintenance", Resource: "maintenance", Action: "create"},
	{Name: "maintenance.update", DisplayName: "Update Maintenance", Resource: "maintenance", Action: "update"},
	{Name: "maintenance.delete", DisplayName: "Delete Maintenance", Resource: "maintenance", Action: "delete"},

	// User management permissions
	{Name: "users.read", DisplayName: "View Users", Resource: "users", Action: "read"},
	{Name: "users.create", DisplayName: "Create Users", Resource: "users", Action: "create"},
	{Name: "users.update", DisplayName: "Update Users", Resource: "users", Action: "update"},
	{Name: "users.delete", DisplayName: "Delete Users", Resource: "users", Action: "delete"},

	// Settings permissions
	{Name: "settings.read", DisplayName: "View Settings", Resource: "settings", Action: "read"},
	{Name: "settings.update", DisplayName: "Update Settings", Resource: "settings", Action: "update"},

	// Analytics permissions
	{Name: "analytics.read", DisplayName: "View Analytics", Resource: "analytics", Action: "read"},

	// Webhook permissions
	{Name: "webhooks.read", DisplayName: "View Webhooks", Resource: "webhooks", Action: "read"},
	{Name: "webhooks.create", DisplayName: "Create Webhooks", Resource: "webhooks", Action: "create"},
	{Name: "webhooks.update", DisplayName: "Update Webhooks", Resource: "webhooks", Action: "update"},
	{Name: "webhooks.delete", DisplayName: "Delete Webhooks", Resource: "webhooks", Action: "delete"},

	// Integration permissions
	{Name: "integrations.read", DisplayName: "View Integrations", Resource: "integrations", Action: "read"},
	{Name: "integrations.create", DisplayName: "Create Integrations", Resource: "integrations", Action: "create"},
	{Name: "integrations.update", DisplayName: "Update Integrations", Resource: "integrations", Action: "update"},
	{Name: "integrations.delete", DisplayName: "Delete Integrations", Resource: "integrations", Action: "delete"},
}

// Permission sets for default roles
var DefaultRolePermissions = map[string][]string{
	"admin": {
		"components.read", "components.create", "components.update", "components.delete",
		"incidents.read", "incidents.create", "incidents.update", "incidents.delete",
		"maintenance.read", "maintenance.create", "maintenance.update", "maintenance.delete",
		"users.read", "users.create", "users.update", "users.delete",
		"settings.read", "settings.update",
		"analytics.read",
		"webhooks.read", "webhooks.create", "webhooks.update", "webhooks.delete",
		"integrations.read", "integrations.create", "integrations.update", "integrations.delete",
	},
	"editor": {
		"components.read", "components.create", "components.update",
		"incidents.read", "incidents.create", "incidents.update",
		"maintenance.read", "maintenance.create", "maintenance.update",
		"analytics.read",
	},
	"viewer": {
		"components.read",
		"incidents.read",
		"maintenance.read",
		"analytics.read",
	},
	"responder": {
		"components.read", "components.update",
		"incidents.read", "incidents.create", "incidents.update",
		"maintenance.read",
	},
}

// Validation methods
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	return r.validate()
}

func (r *Role) BeforeUpdate(tx *gorm.DB) error {
	return r.validate()
}

func (r *Role) validate() error {
	if r.Name == "" {
		return gorm.ErrInvalidValue
	}
	if r.DisplayName == "" {
		return gorm.ErrInvalidValue
	}
	return nil
}

func (t *Team) BeforeCreate(tx *gorm.DB) error {
	return t.validate()
}

func (t *Team) BeforeUpdate(tx *gorm.DB) error {
	return t.validate()
}

func (t *Team) validate() error {
	if t.Name == "" {
		return gorm.ErrInvalidValue
	}
	return nil
}