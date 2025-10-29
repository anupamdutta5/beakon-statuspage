package events

import (
	"time"

	"github.com/anupamdutta5/saas-admin-service/internal/models"
	"github.com/google/uuid"
)

// NewTenantCreatedEvent creates a new tenant created event from SaaSTenant
// adminEmail and adminPassword are optional - if provided, tenant-admin-service will create admin user
func NewTenantCreatedEvent(tenant *models.SaaSTenant, metadata EventMetadata, adminEmail, adminPassword string) TenantEvent {
	return TenantEvent{
		EventID:   uuid.New().String(),
		EventType: TenantCreated,
		TenantID:  tenant.ID,
		Timestamp: time.Now().UTC(),
		Data:      saasTenantToDataWithAdmin(tenant, adminEmail, adminPassword),
		Metadata:  metadata,
	}
}

// NewTenantUpdatedEvent creates a new tenant updated event from SaaSTenant
func NewTenantUpdatedEvent(tenant *models.SaaSTenant, metadata EventMetadata) TenantEvent {
	return TenantEvent{
		EventID:   uuid.New().String(),
		EventType: TenantUpdated,
		TenantID:  tenant.ID,
		Timestamp: time.Now().UTC(),
		Data:      saasTenantToData(tenant),
		Metadata:  metadata,
	}
}

// NewTenantDeletedEvent creates a new tenant deleted event from SaaSTenant
func NewTenantDeletedEvent(tenant *models.SaaSTenant, metadata EventMetadata) TenantEvent {
	return TenantEvent{
		EventID:   uuid.New().String(),
		EventType: TenantDeleted,
		TenantID:  tenant.ID,
		Timestamp: time.Now().UTC(),
		Data:      saasTenantToData(tenant),
		Metadata:  metadata,
	}
}

// NewTenantRestoredEvent creates a new tenant restored event from SaaSTenant
func NewTenantRestoredEvent(tenant *models.SaaSTenant, metadata EventMetadata) TenantEvent {
	return TenantEvent{
		EventID:   uuid.New().String(),
		EventType: TenantRestored,
		TenantID:  tenant.ID,
		Timestamp: time.Now().UTC(),
		Data:      saasTenantToData(tenant),
		Metadata:  metadata,
	}
}

// saasTenantToData converts a SaaSTenant model to TenantData event data
func saasTenantToData(tenant *models.SaaSTenant) TenantData {
	return saasTenantToDataWithAdmin(tenant, "", "")
}

// saasTenantToDataWithAdmin converts a SaaSTenant model to TenantData event data with admin credentials
func saasTenantToDataWithAdmin(tenant *models.SaaSTenant, adminEmail, adminPassword string) TenantData {
	// Convert string fields to pointers for optional fields
	var domain, subdomain, billingEmail, settings, branding, features *string

	if tenant.Domain != "" {
		domain = &tenant.Domain
	}
	if tenant.Subdomain != "" {
		subdomain = &tenant.Subdomain
	}
	if tenant.BillingEmail != "" {
		billingEmail = &tenant.BillingEmail
	}
	if tenant.Settings != "" {
		settings = &tenant.Settings
	}
	if tenant.Branding != "" {
		branding = &tenant.Branding
	}
	if tenant.Features != "" {
		features = &tenant.Features
	}

	// Convert MaxUsers from *int64 to int
	maxUsers := 0
	if tenant.MaxUsers != nil {
		maxUsers = int(*tenant.MaxUsers)
	}

	// Convert DeletedAt from gorm.DeletedAt to *time.Time
	var deletedAt *time.Time
	if tenant.DeletedAt.Valid {
		deletedAt = &tenant.DeletedAt.Time
	}

	// Convert admin credentials to pointers
	var adminEmailPtr, adminPasswordPtr *string
	if adminEmail != "" {
		adminEmailPtr = &adminEmail
	}
	if adminPassword != "" {
		adminPasswordPtr = &adminPassword
	}

	return TenantData{
		ID:            tenant.ID,
		Name:          tenant.Name,
		Slug:          tenant.Slug,
		Domain:        domain,
		Subdomain:     subdomain,
		ContactEmail:  tenant.ContactEmail,
		BillingEmail:  billingEmail,
		PlanID:        tenant.PlanID,
		Status:        tenant.Status,
		MaxUsers:      maxUsers,
		Settings:      settings,
		Branding:      branding,
		Features:      features,
		IsActive:      tenant.IsActive,
		CreatedAt:     tenant.CreatedAt,
		UpdatedAt:     tenant.UpdatedAt,
		DeletedAt:     deletedAt,
		AdminEmail:    adminEmailPtr,
		AdminPassword: adminPasswordPtr,
	}
}
