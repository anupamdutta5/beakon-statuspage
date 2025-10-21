package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantAdminUser represents a user in the tenant_admin_db database.
// This model is used by SaaS Admin Service to directly update tenant owner credentials
// in the tenant_admin_db.users table, bypassing HTTP API calls for atomic updates.
//
// Note: This is an exception to the database-per-service pattern for critical credential sync.
type TenantAdminUser struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Email        string         `gorm:"not null;uniqueIndex" json:"email"`
	PasswordHash string         `gorm:"column:password_hash;not null" json:"-"`
	FirstName    string         `gorm:"column:first_name" json:"first_name,omitempty"`
	LastName     string         `gorm:"column:last_name" json:"last_name,omitempty"`
	TenantID     uuid.UUID      `gorm:"type:uuid;index" json:"tenant_id"`
	Role         string         `gorm:"size:50;not null;default:'admin'" json:"role"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	LastLoginAt  *time.Time     `gorm:"column:last_login_at" json:"last_login_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for GORM
func (TenantAdminUser) TableName() string {
	return "users"
}
