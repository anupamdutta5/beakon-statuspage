// Package services provides business logic for the Tenant Service.
package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/enterprise-status/statuspage-tenant-service/internal/config"
	"github.com/enterprise-status/statuspage-tenant-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TenantService handles tenant-related business logic.
type TenantService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewTenantService creates a new tenant service.
func NewTenantService(db *gorm.DB, logger *zap.Logger) *TenantService {
	return &TenantService{
		db:     db,
		logger: logger,
	}
}

// CreateTenant creates a new tenant.
func (s *TenantService) CreateTenant(tenant *models.Tenant) error {
	// Generate slug if not provided
	if tenant.Slug == "" {
		tenant.Slug = s.generateSlug(tenant.Name)
	}

	// Ensure slug is unique
	originalSlug := tenant.Slug
	counter := 1
	for {
		var existing models.Tenant
		if err := s.db.Where("slug = ?", tenant.Slug).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				break // Slug is unique
			}
			s.logger.Error("Failed to check slug uniqueness", zap.Error(err))
			return fmt.Errorf("failed to check slug uniqueness: %w", err)
		}
		tenant.Slug = fmt.Sprintf("%s-%d", originalSlug, counter)
		counter++
	}

	// Set default values
	if tenant.Plan == "" {
		tenant.Plan = "free"
	}
	if tenant.Status == "" {
		tenant.Status = "active"
	}

	// Create tenant
	if err := s.db.Create(tenant).Error; err != nil {
		s.logger.Error("Failed to create tenant", zap.Error(err))
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create default settings
	settings := &models.TenantSettings{
		TenantID:            tenant.ID,
		Timezone:            "UTC",
		Language:            "en",
		DateFormat:          "YYYY-MM-DD",
		TimeFormat:          "24h",
		EmailNotifications:  true,
		SMSNotifications:    false,
		ShowIncidentHistory: true,
		ShowMaintenanceMode: true,
		RequireAuth:         false,
		AllowPublicAccess:   true,
		SessionTimeout:      30,
		APIRateLimit:        1000,
		APIKeyRequired:      false,
	}

	if err := s.db.Create(settings).Error; err != nil {
		s.logger.Error("Failed to create tenant settings", zap.Error(err))
		// Don't fail tenant creation if settings creation fails
	}

	// Create default billing
	billing := &models.TenantBilling{
		TenantID:         tenant.ID,
		Plan:             tenant.Plan,
		BillingCycle:     "monthly",
		Amount:           0.0,
		Currency:         "USD",
		NextBillingDate:  time.Now().AddDate(0, 1, 0),
		MaxUsers:         5,
		MaxComponents:    10,
		MaxIncidents:     50,
		MaxAPIRequests:   1000,
		Status:           "active",
		IsTrial:          true,
		TrialEndsAt:      &[]time.Time{time.Now().AddDate(0, 0, 14)}[0], // 14-day trial
	}

	if err := s.db.Create(billing).Error; err != nil {
		s.logger.Error("Failed to create tenant billing", zap.Error(err))
		// Don't fail tenant creation if billing creation fails
	}

	// Create default branding
	branding := &models.TenantBranding{
		TenantID:        tenant.ID,
		PrimaryColor:    "#007bff",
		SecondaryColor:  "#6c757d",
		AccentColor:     "#28a745",
		BackgroundColor: "#ffffff",
		TextColor:       "#333333",
		FontFamily:      "system-ui, -apple-system, sans-serif",
		FontSize:        "14px",
		FontWeight:      "400",
		Layout:          "default",
		ShowLogo:        true,
		ShowFooter:      true,
		FooterText:      "Powered by Status Page",
	}

	if err := s.db.Create(branding).Error; err != nil {
		s.logger.Error("Failed to create tenant branding", zap.Error(err))
		// Don't fail tenant creation if branding creation fails
	}

	s.logger.Info("Tenant created successfully", zap.Uint("tenant_id", tenant.ID))
	return nil
}

// GetTenant retrieves a tenant by ID.
func (s *TenantService) GetTenant(id uint) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := s.db.First(&tenant, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant not found")
		}
		s.logger.Error("Failed to get tenant", zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return &tenant, nil
}

// GetTenantBySlug retrieves a tenant by slug.
func (s *TenantService) GetTenantBySlug(slug string) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := s.db.Where("slug = ?", slug).First(&tenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant not found")
		}
		s.logger.Error("Failed to get tenant by slug", zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant by slug: %w", err)
	}

	return &tenant, nil
}

// GetTenantByDomain retrieves a tenant by domain.
func (s *TenantService) GetTenantByDomain(domain string) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := s.db.Where("domain = ? OR subdomain = ?", domain, domain).First(&tenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant not found")
		}
		s.logger.Error("Failed to get tenant by domain", zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant by domain: %w", err)
	}

	return &tenant, nil
}

// GetTenants retrieves a list of tenants with pagination.
func (s *TenantService) GetTenants(limit, offset int) ([]*models.Tenant, int64, error) {
	var tenants []*models.Tenant
	var total int64

	// Get total count
	if err := s.db.Model(&models.Tenant{}).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count tenants", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	// Get tenants with pagination
	if err := s.db.Limit(limit).Offset(offset).Find(&tenants).Error; err != nil {
		s.logger.Error("Failed to get tenants", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get tenants: %w", err)
	}

	return tenants, total, nil
}

// UpdateTenant updates a tenant.
func (s *TenantService) UpdateTenant(tenant *models.Tenant) error {
	if err := s.db.Save(tenant).Error; err != nil {
		s.logger.Error("Failed to update tenant", zap.Error(err))
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	s.logger.Info("Tenant updated successfully", zap.Uint("tenant_id", tenant.ID))
	return nil
}

// DeleteTenant soft deletes a tenant.
func (s *TenantService) DeleteTenant(id uint) error {
	if err := s.db.Delete(&models.Tenant{}, id).Error; err != nil {
		s.logger.Error("Failed to delete tenant", zap.Error(err))
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	s.logger.Info("Tenant deleted successfully", zap.Uint("tenant_id", id))
	return nil
}

// GetTenantSettings retrieves tenant settings.
func (s *TenantService) GetTenantSettings(tenantID uint) (*models.TenantSettings, error) {
	var settings models.TenantSettings
	if err := s.db.Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant settings not found")
		}
		s.logger.Error("Failed to get tenant settings", zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant settings: %w", err)
	}

	return &settings, nil
}

// UpdateTenantSettings updates tenant settings.
func (s *TenantService) UpdateTenantSettings(settings *models.TenantSettings) error {
	if err := s.db.Save(settings).Error; err != nil {
		s.logger.Error("Failed to update tenant settings", zap.Error(err))
		return fmt.Errorf("failed to update tenant settings: %w", err)
	}

	s.logger.Info("Tenant settings updated successfully", zap.Uint("tenant_id", settings.TenantID))
	return nil
}

// GetTenantBilling retrieves tenant billing information.
func (s *TenantService) GetTenantBilling(tenantID uint) (*models.TenantBilling, error) {
	var billing models.TenantBilling
	if err := s.db.Where("tenant_id = ?", tenantID).First(&billing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant billing not found")
		}
		s.logger.Error("Failed to get tenant billing", zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant billing: %w", err)
	}

	return &billing, nil
}

// UpdateTenantBilling updates tenant billing information.
func (s *TenantService) UpdateTenantBilling(billing *models.TenantBilling) error {
	if err := s.db.Save(billing).Error; err != nil {
		s.logger.Error("Failed to update tenant billing", zap.Error(err))
		return fmt.Errorf("failed to update tenant billing: %w", err)
	}

	s.logger.Info("Tenant billing updated successfully", zap.Uint("tenant_id", billing.TenantID))
	return nil
}

// GetTenantBranding retrieves tenant branding information.
func (s *TenantService) GetTenantBranding(tenantID uint) (*models.TenantBranding, error) {
	var branding models.TenantBranding
	if err := s.db.Where("tenant_id = ?", tenantID).First(&branding).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant branding not found")
		}
		s.logger.Error("Failed to get tenant branding", zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant branding: %w", err)
	}

	return &branding, nil
}

// UpdateTenantBranding updates tenant branding information.
func (s *TenantService) UpdateTenantBranding(branding *models.TenantBranding) error {
	if err := s.db.Save(branding).Error; err != nil {
		s.logger.Error("Failed to update tenant branding", zap.Error(err))
		return fmt.Errorf("failed to update tenant branding: %w", err)
	}

	s.logger.Info("Tenant branding updated successfully", zap.Uint("tenant_id", branding.TenantID))
	return nil
}

// LogTenantActivity logs tenant activity.
func (s *TenantService) LogTenantActivity(activity *models.TenantActivity) error {
	if err := s.db.Create(activity).Error; err != nil {
		s.logger.Error("Failed to log tenant activity", zap.Error(err))
		return fmt.Errorf("failed to log tenant activity: %w", err)
	}

	return nil
}

// generateSlug generates a URL-friendly slug from a string.
func (s *TenantService) generateSlug(input string) string {
	// Convert to lowercase
	slug := strings.ToLower(input)
	
	// Replace spaces and special characters with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	slug = strings.ReplaceAll(slug, ".", "-")
	slug = strings.ReplaceAll(slug, ",", "-")
	slug = strings.ReplaceAll(slug, ":", "-")
	slug = strings.ReplaceAll(slug, ";", "-")
	slug = strings.ReplaceAll(slug, "!", "-")
	slug = strings.ReplaceAll(slug, "?", "-")
	slug = strings.ReplaceAll(slug, "(", "-")
	slug = strings.ReplaceAll(slug, ")", "-")
	slug = strings.ReplaceAll(slug, "[", "-")
	slug = strings.ReplaceAll(slug, "]", "-")
	slug = strings.ReplaceAll(slug, "{", "-")
	slug = strings.ReplaceAll(slug, "}", "-")
	slug = strings.ReplaceAll(slug, "@", "-")
	slug = strings.ReplaceAll(slug, "#", "-")
	slug = strings.ReplaceAll(slug, "$", "-")
	slug = strings.ReplaceAll(slug, "%", "-")
	slug = strings.ReplaceAll(slug, "^", "-")
	slug = strings.ReplaceAll(slug, "&", "-")
	slug = strings.ReplaceAll(slug, "*", "-")
	slug = strings.ReplaceAll(slug, "+", "-")
	slug = strings.ReplaceAll(slug, "=", "-")
	slug = strings.ReplaceAll(slug, "|", "-")
	slug = strings.ReplaceAll(slug, "\\", "-")
	slug = strings.ReplaceAll(slug, "/", "-")
	slug = strings.ReplaceAll(slug, "<", "-")
	slug = strings.ReplaceAll(slug, ">", "-")
	slug = strings.ReplaceAll(slug, "\"", "-")
	slug = strings.ReplaceAll(slug, "'", "-")
	slug = strings.ReplaceAll(slug, "`", "-")
	slug = strings.ReplaceAll(slug, "~", "-")
	
	// Remove multiple consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	
	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")
	
	// Ensure slug is not empty
	if slug == "" {
		slug = "tenant"
	}
	
	return slug
}

// InitDatabase initializes the database connection and runs migrations.
func InitDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.Tenant{},
		&models.TenantSettings{},
		&models.TenantBilling{},
		&models.TenantBranding{},
		&models.TenantActivity{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
