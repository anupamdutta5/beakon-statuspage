package events

import (
	"context"
	"fmt"

	"github.com/anupamdutta5/tenant-admin-service/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TenantServiceInterface defines the methods needed from TenantAdminService
// This interface allows proper dependency injection and testing
type TenantServiceInterface interface {
	CreateAdminUser(ctx context.Context, tenantID uuid.UUID, email, password string) error
	CreateTenant(tenant *models.Tenant) error
}

// RabbitMQTenantEventHandler handles tenant lifecycle events
type RabbitMQTenantEventHandler struct {
	db            *gorm.DB
	tenantService TenantServiceInterface
	logger        *zap.Logger
}

// NewRabbitMQTenantEventHandler creates a new tenant event handler
func NewRabbitMQTenantEventHandler(db *gorm.DB, tenantService TenantServiceInterface, logger *zap.Logger) *RabbitMQTenantEventHandler {
	return &RabbitMQTenantEventHandler{
		db:            db,
		tenantService: tenantService,
		logger:        logger,
	}
}

// HandleTenantCreated handles tenant.created events
func (h *RabbitMQTenantEventHandler) HandleTenantCreated(ctx context.Context, event RabbitMQTenantEvent) error {
	h.logger.Info("Handling tenant.created event",
		zap.String("event_id", event.EventID),
		zap.String("tenant_id", event.TenantID.String()),
		zap.String("tenant_name", event.Data.Name))

	// Check if tenant already exists (idempotency)
	var existingTenant models.Tenant
	result := h.db.Where("id = ?", event.Data.ID).First(&existingTenant)
	if result.Error == nil {
		h.logger.Info("Tenant already exists, skipping creation (idempotent)",
			zap.String("tenant_id", event.TenantID.String()))
		return nil
	} else if result.Error != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing tenant: %w", result.Error)
	}

	// Convert event data to tenant model
	tenant := h.eventDataToTenant(event.Data)

	// Create tenant in database
	if err := h.db.Create(&tenant).Error; err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	h.logger.Info("Successfully created tenant from event",
		zap.String("tenant_id", tenant.ID.String()),
		zap.String("tenant_name", tenant.Name))

	// Create admin user if credentials are provided in the event
	if event.Data.AdminEmail != nil && *event.Data.AdminEmail != "" {
		// Use provided password or generate a secure one
		password := ""
		if event.Data.AdminPassword != nil && *event.Data.AdminPassword != "" {
			password = *event.Data.AdminPassword
		} else {
			// Generate a secure random password (UUID-based)
			password = uuid.New().String()
			h.logger.Info("No password provided, auto-generating secure password for admin user",
				zap.String("tenant_id", tenant.ID.String()),
				zap.String("email", *event.Data.AdminEmail))
		}

		// Use service layer to create admin user (proper business logic separation)
		if err := h.tenantService.CreateAdminUser(ctx, tenant.ID, *event.Data.AdminEmail, password); err != nil {
			h.logger.Error("Failed to create admin user for tenant",
				zap.String("tenant_id", tenant.ID.String()),
				zap.String("admin_email", *event.Data.AdminEmail),
				zap.Error(err))
			// Return error so message gets requeued
			return fmt.Errorf("failed to create admin user: %w", err)
		}
		h.logger.Info("Successfully created admin user from event",
			zap.String("tenant_id", tenant.ID.String()),
			zap.String("admin_email", *event.Data.AdminEmail))
	} else {
		h.logger.Warn("No admin email provided in event, skipping admin user creation",
			zap.String("tenant_id", tenant.ID.String()))
	}

	return nil
}

// HandleTenantUpdated handles tenant.updated events
func (h *RabbitMQTenantEventHandler) HandleTenantUpdated(ctx context.Context, event RabbitMQTenantEvent) error {
	h.logger.Info("Handling tenant.updated event",
		zap.String("event_id", event.EventID),
		zap.String("tenant_id", event.TenantID.String()),
		zap.String("tenant_name", event.Data.Name))

	// Find existing tenant
	var existingTenant models.Tenant
	result := h.db.Where("id = ?", event.Data.ID).First(&existingTenant)
	if result.Error == gorm.ErrRecordNotFound {
		// Tenant doesn't exist, create it instead
		h.logger.Warn("Tenant not found for update, creating instead",
			zap.String("tenant_id", event.TenantID.String()))
		return h.HandleTenantCreated(ctx, event)
	} else if result.Error != nil {
		return fmt.Errorf("failed to find tenant: %w", result.Error)
	}

	// Convert event data to tenant model
	updatedTenant := h.eventDataToTenant(event.Data)

	// Update tenant
	if err := h.db.Model(&existingTenant).Updates(updatedTenant).Error; err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	h.logger.Info("Successfully updated tenant from event",
		zap.String("tenant_id", updatedTenant.ID.String()),
		zap.String("tenant_name", updatedTenant.Name))

	return nil
}

// HandleTenantDeleted handles tenant.deleted events
func (h *RabbitMQTenantEventHandler) HandleTenantDeleted(ctx context.Context, event RabbitMQTenantEvent) error {
	h.logger.Info("Handling tenant.deleted event",
		zap.String("event_id", event.EventID),
		zap.String("tenant_id", event.TenantID.String()))

	// Soft delete the tenant
	result := h.db.Where("id = ?", event.Data.ID).Delete(&models.Tenant{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete tenant: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		h.logger.Warn("Tenant not found for deletion (idempotent)",
			zap.String("tenant_id", event.TenantID.String()))
		return nil
	}

	h.logger.Info("Successfully deleted tenant from event",
		zap.String("tenant_id", event.TenantID.String()),
		zap.Int64("rows_affected", result.RowsAffected))

	return nil
}

// HandleTenantRestored handles tenant.restored events
func (h *RabbitMQTenantEventHandler) HandleTenantRestored(ctx context.Context, event RabbitMQTenantEvent) error {
	h.logger.Info("Handling tenant.restored event",
		zap.String("event_id", event.EventID),
		zap.String("tenant_id", event.TenantID.String()))

	// Restore the soft-deleted tenant
	result := h.db.Unscoped().Model(&models.Tenant{}).
		Where("id = ?", event.Data.ID).
		Update("deleted_at", nil)

	if result.Error != nil {
		return fmt.Errorf("failed to restore tenant: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		h.logger.Warn("Tenant not found for restoration",
			zap.String("tenant_id", event.TenantID.String()))
		// Create the tenant if it doesn't exist
		return h.HandleTenantCreated(ctx, event)
	}

	h.logger.Info("Successfully restored tenant from event",
		zap.String("tenant_id", event.TenantID.String()),
		zap.Int64("rows_affected", result.RowsAffected))

	return nil
}

// eventDataToTenant converts event data to tenant model
func (h *RabbitMQTenantEventHandler) eventDataToTenant(data RabbitMQTenantData) models.Tenant {
	// Convert MaxUsers from int to *int
	var maxUsers *int
	if data.MaxUsers > 0 {
		maxUsers = &data.MaxUsers
	}

	tenant := models.Tenant{
		ID:           data.ID,
		Name:         data.Name,
		Slug:         data.Slug,
		Email:        data.ContactEmail,  // Map contact_email to email (required field)
		ContactEmail: data.ContactEmail,
		Status:       data.Status,
		MaxUsers:     maxUsers,
		IsActive:     data.IsActive,
		CreatedAt:    data.CreatedAt,
		UpdatedAt:    data.UpdatedAt,
	}

	// Handle optional pointer fields - convert to string
	if data.Domain != nil {
		tenant.Domain = *data.Domain
	}
	if data.Subdomain != nil {
		tenant.Subdomain = *data.Subdomain
	}
	if data.BillingEmail != nil {
		tenant.BillingEmail = *data.BillingEmail
	}
	if data.Settings != nil {
		tenant.Settings = *data.Settings
	}
	if data.Branding != nil {
		tenant.Branding = *data.Branding
	}
	if data.Features != nil {
		tenant.Features = *data.Features
	}

	return tenant
}
