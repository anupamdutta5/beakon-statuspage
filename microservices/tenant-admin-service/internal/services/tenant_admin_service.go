// Package services provides business logic for the Tenant Admin Service.
package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/anupamdutta5/statuspage-tenant-admin-service/internal/config"
	"github.com/anupamdutta5/statuspage-tenant-admin-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TenantAdminService handles tenant admin-related business logic.
type TenantAdminService struct {
	config *config.Config
	logger *zap.Logger
	db     *gorm.DB
}

// NewTenantAdminService creates a new tenant admin service.
func NewTenantAdminService(cfg *config.Config, logger *zap.Logger) (*TenantAdminService, error) {
	// Initialize database connection
	db, err := initDatabase(cfg.Database)
	if err != nil {
		logger.Warn("Failed to initialize database, running without database", zap.Error(err))
		db = nil // Allow service to run without database for testing
	}

	return &TenantAdminService{
		config: cfg,
		logger: logger,
		db:     db,
	}, nil
}

// SetDB allows injecting a test database for testing purposes
func (s *TenantAdminService) SetDB(db *gorm.DB) {
	s.db = db
}

// GetDB returns the database connection for use by other services
func (s *TenantAdminService) GetDB() *gorm.DB {
	return s.db
}

// Tenant Admin Management

// CreateTenantAdmin creates a new tenant admin.
func (s *TenantAdminService) CreateTenantAdmin(ctx context.Context, admin *models.TenantAdmin) error {
	s.logger.Info("Creating tenant admin",
		zap.Uint("tenant_id", admin.TenantID),
		zap.Uint("user_id", admin.UserID),
		zap.String("role", admin.Role))

	if err := s.db.Create(admin).Error; err != nil {
		s.logger.Error("Failed to create tenant admin", zap.Error(err))
		return fmt.Errorf("failed to create tenant admin: %w", err)
	}

	s.logger.Info("Tenant admin created successfully",
		zap.Uint("admin_id", admin.ID),
		zap.Uint("tenant_id", admin.TenantID))

	return nil
}

// GetTenantAdmin retrieves a tenant admin by ID.
func (s *TenantAdminService) GetTenantAdmin(ctx context.Context, adminID uint) (*models.TenantAdmin, error) {
	s.logger.Info("Getting tenant admin", zap.Uint("admin_id", adminID))

	var admin models.TenantAdmin
	if err := s.db.First(&admin, adminID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant admin not found")
		}
		s.logger.Error("Failed to get tenant admin", zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant admin: %w", err)
	}

	return &admin, nil
}

// ListTenantAdmins lists all admins for a tenant.
func (s *TenantAdminService) ListTenantAdmins(ctx context.Context, tenantID uint, limit, offset int) ([]*models.TenantAdmin, error) {
	s.logger.Info("Listing tenant admins", zap.Uint("tenant_id", tenantID), zap.Int("limit", limit), zap.Int("offset", offset))

	var admins []*models.TenantAdmin
	query := s.db.Where("tenant_id = ?", tenantID)

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&admins).Error; err != nil {
		s.logger.Error("Failed to list tenant admins", zap.Error(err))
		return nil, fmt.Errorf("failed to list tenant admins: %w", err)
	}

	return admins, nil
}

// UpdateTenantAdmin updates a tenant admin.
func (s *TenantAdminService) UpdateTenantAdmin(ctx context.Context, adminID uint, updates *models.TenantAdmin) error {
	s.logger.Info("Updating tenant admin", zap.Uint("admin_id", adminID))

	if err := s.db.Model(&models.TenantAdmin{}).Where("id = ?", adminID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update tenant admin", zap.Error(err))
		return fmt.Errorf("failed to update tenant admin: %w", err)
	}

	s.logger.Info("Tenant admin updated successfully", zap.Uint("admin_id", adminID))
	return nil
}

// DeleteTenantAdmin deletes a tenant admin.
func (s *TenantAdminService) DeleteTenantAdmin(ctx context.Context, adminID uint) error {
	s.logger.Info("Deleting tenant admin", zap.Uint("admin_id", adminID))

	if err := s.db.Delete(&models.TenantAdmin{}, adminID).Error; err != nil {
		s.logger.Error("Failed to delete tenant admin", zap.Error(err))
		return fmt.Errorf("failed to delete tenant admin: %w", err)
	}

	s.logger.Info("Tenant admin deleted successfully", zap.Uint("admin_id", adminID))
	return nil
}

// Tenant Settings Management

// GetTenantSettings retrieves tenant settings.
func (s *TenantAdminService) GetTenantSettings(ctx context.Context, tenantID uint) (*models.TenantSettings, error) {
	s.logger.Info("Getting tenant settings", zap.Uint("tenant_id", tenantID))

	if s.db == nil {
		s.logger.Debug("Database not available, returning default tenant settings")
		return &models.TenantSettings{
			TenantID: tenantID,
			Settings: `{
				"notifications": {
					"email_enabled": true,
					"sms_enabled": false,
					"webhook_url": "https://example.com/webhook"
				},
				"branding": {
					"logo_url": "https://example.com/logo.png",
					"primary_color": "#007bff",
					"secondary_color": "#6c757d"
				},
				"security": {
					"two_factor_enabled": true,
					"session_timeout": 3600
				}
			}`,
			Version: "1.0.0",
			Status:  "active",
		}, nil
	}

	var settings models.TenantSettings
	if err := s.db.Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create default settings if not exists
			settings = models.TenantSettings{
				TenantID: tenantID,
				Settings: "{}",
				Version:  "1.0.0",
				Status:   "active",
			}
			if err := s.db.Create(&settings).Error; err != nil {
				s.logger.Error("Failed to create default tenant settings", zap.Error(err))
				return nil, fmt.Errorf("failed to create default tenant settings: %w", err)
			}
		} else {
			s.logger.Error("Failed to get tenant settings", zap.Error(err))
			return nil, fmt.Errorf("failed to get tenant settings: %w", err)
		}
	}

	return &settings, nil
}

// UpdateTenantSettings updates tenant settings.
func (s *TenantAdminService) UpdateTenantSettings(ctx context.Context, tenantID uint, updates *models.TenantSettings) error {
	s.logger.Info("Updating tenant settings", zap.Uint("tenant_id", tenantID))

	if err := s.db.Model(&models.TenantSettings{}).Where("tenant_id = ?", tenantID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update tenant settings", zap.Error(err))
		return fmt.Errorf("failed to update tenant settings: %w", err)
	}

	s.logger.Info("Tenant settings updated successfully", zap.Uint("tenant_id", tenantID))
	return nil
}

// Feature Flag Management

// CreateTenantFeatureFlag creates a new tenant feature flag.
func (s *TenantAdminService) CreateTenantFeatureFlag(ctx context.Context, flag *models.TenantFeatureFlag) error {
	s.logger.Info("Creating tenant feature flag",
		zap.Uint("tenant_id", flag.TenantID),
		zap.String("flag_name", flag.Name))

	if err := s.db.Create(flag).Error; err != nil {
		s.logger.Error("Failed to create tenant feature flag", zap.Error(err))
		return fmt.Errorf("failed to create tenant feature flag: %w", err)
	}

	s.logger.Info("Tenant feature flag created successfully",
		zap.Uint("flag_id", flag.ID),
		zap.String("flag_name", flag.Name))

	return nil
}

// GetTenantFeatureFlag retrieves a tenant feature flag by ID.
func (s *TenantAdminService) GetTenantFeatureFlag(ctx context.Context, flagID uint) (*models.TenantFeatureFlag, error) {
	s.logger.Info("Getting tenant feature flag", zap.Uint("flag_id", flagID))

	var flag models.TenantFeatureFlag
	if err := s.db.First(&flag, flagID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant feature flag not found")
		}
		s.logger.Error("Failed to get tenant feature flag", zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant feature flag: %w", err)
	}

	return &flag, nil
}

// ListTenantFeatureFlags lists all feature flags for a tenant.
func (s *TenantAdminService) ListTenantFeatureFlags(ctx context.Context, tenantID uint, limit, offset int) ([]*models.TenantFeatureFlag, error) {
	s.logger.Info("Listing tenant feature flags", zap.Uint("tenant_id", tenantID), zap.Int("limit", limit), zap.Int("offset", offset))

	var flags []*models.TenantFeatureFlag
	query := s.db.Where("tenant_id = ?", tenantID)

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&flags).Error; err != nil {
		s.logger.Error("Failed to list tenant feature flags", zap.Error(err))
		return nil, fmt.Errorf("failed to list tenant feature flags: %w", err)
	}

	return flags, nil
}

// UpdateTenantFeatureFlag updates a tenant feature flag.
func (s *TenantAdminService) UpdateTenantFeatureFlag(ctx context.Context, flagID uint, updates *models.TenantFeatureFlag) error {
	s.logger.Info("Updating tenant feature flag", zap.Uint("flag_id", flagID))

	if err := s.db.Model(&models.TenantFeatureFlag{}).Where("id = ?", flagID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update tenant feature flag", zap.Error(err))
		return fmt.Errorf("failed to update tenant feature flag: %w", err)
	}

	s.logger.Info("Tenant feature flag updated successfully", zap.Uint("flag_id", flagID))
	return nil
}

// DeleteTenantFeatureFlag deletes a tenant feature flag.
func (s *TenantAdminService) DeleteTenantFeatureFlag(ctx context.Context, flagID uint) error {
	s.logger.Info("Deleting tenant feature flag", zap.Uint("flag_id", flagID))

	if err := s.db.Delete(&models.TenantFeatureFlag{}, flagID).Error; err != nil {
		s.logger.Error("Failed to delete tenant feature flag", zap.Error(err))
		return fmt.Errorf("failed to delete tenant feature flag: %w", err)
	}

	s.logger.Info("Tenant feature flag deleted successfully", zap.Uint("flag_id", flagID))
	return nil
}

// Usage Management

// GetTenantUsage retrieves tenant usage metrics.
func (s *TenantAdminService) GetTenantUsage(ctx context.Context, tenantID uint, startDate, endDate time.Time) ([]*models.TenantUsage, error) {
	s.logger.Info("Getting tenant usage",
		zap.Uint("tenant_id", tenantID),
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate))

	var usage []*models.TenantUsage
	query := s.db.Where("tenant_id = ? AND date >= ? AND date <= ?", tenantID, startDate, endDate)

	if err := query.Find(&usage).Error; err != nil {
		s.logger.Error("Failed to get tenant usage", zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant usage: %w", err)
	}

	return usage, nil
}

// RecordTenantUsage records tenant usage metrics.
func (s *TenantAdminService) RecordTenantUsage(ctx context.Context, usage *models.TenantUsage) error {
	s.logger.Info("Recording tenant usage",
		zap.Uint("tenant_id", usage.TenantID),
		zap.Time("date", usage.Date))

	if err := s.db.Create(usage).Error; err != nil {
		s.logger.Error("Failed to record tenant usage", zap.Error(err))
		return fmt.Errorf("failed to record tenant usage: %w", err)
	}

	s.logger.Info("Tenant usage recorded successfully",
		zap.Uint("usage_id", usage.ID),
		zap.Uint("tenant_id", usage.TenantID))

	return nil
}

// Statistics

// GetTenantStats returns tenant statistics.
func (s *TenantAdminService) GetTenantStats(ctx context.Context, tenantID uint) (*models.TenantStats, error) {
	s.logger.Info("Getting tenant statistics", zap.Uint("tenant_id", tenantID))

	// For now, we'll return simulated statistics
	// In production, you would query actual statistics from the database

	stats := &models.TenantStats{
		TenantID:          tenantID,
		TotalUsers:        15,
		ActiveUsers:       12,
		TotalServices:     8,
		ActiveServices:    7,
		TotalMonitors:     25,
		ActiveMonitors:    23,
		TotalSubscribers:  150,
		ActiveSubscribers: 140,
		TotalIncidents:    5,
		OpenIncidents:     1,
		TotalMaintenance:  3,
		ActiveMaintenance: 0,
		TotalAPIRequests:  1000,
		TotalPageViews:    5000,
		StorageUsed:       1024000, // 1MB
		BandwidthUsed:     5120000, // 5MB
		LastUpdated:       time.Now(),
	}

	return stats, nil
}

// Health checks the health of the tenant admin service.
func (s *TenantAdminService) Health(ctx context.Context) error {
	s.logger.Debug("Checking tenant admin service health")

	// Check database connection if available
	if s.db != nil {
		if err := s.db.Exec("SELECT 1").Error; err != nil {
			s.logger.Error("Tenant admin service health check failed", zap.Error(err))
			return fmt.Errorf("tenant admin service health check failed: %w", err)
		}
	} else {
		s.logger.Debug("Database not available, skipping database health check")
	}

	return nil
}

// initDatabase initializes the database connection.
func initDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.Tenant{},
		&models.TenantBranding{},
		&models.TenantAdmin{},
		&models.TenantSettings{},
		&models.TenantFeatureFlag{},
		&models.TenantUsage{},
		&models.TenantBilling{},
		&models.TenantNotification{},
		&models.TenantActivity{},
		&models.TenantBackup{},
		&models.TenantStats{},
		// Status page management models
		&models.StatusPage{},
		&models.StatusPageConfig{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// ============================================================================
// CORE TENANT MANAGEMENT METHODS (migrated from tenant-service)
// ============================================================================

// CreateTenant creates a new tenant with default settings, billing, and branding.
func (s *TenantAdminService) CreateTenant(tenant *models.Tenant) error {
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
		TenantID: tenant.ID,
		Settings: `{
			"timezone": "UTC",
			"language": "en",
			"date_format": "YYYY-MM-DD",
			"time_format": "24h",
			"email_notifications": true,
			"sms_notifications": false,
			"show_incident_history": true,
			"show_maintenance_mode": true,
			"require_auth": false,
			"allow_public_access": true,
			"session_timeout": 30,
			"api_rate_limit": 1000,
			"api_key_required": false
		}`,
		Version: "1.0.0",
		Status:  "active",
	}

	if err := s.db.Create(settings).Error; err != nil {
		s.logger.Error("Failed to create tenant settings", zap.Error(err))
		// Don't fail tenant creation if settings creation fails
	}

	// Create default billing
	billing := &models.TenantBilling{
		TenantID:           tenant.ID,
		PlanName:           tenant.Plan,
		BillingCycle:       "monthly",
		Amount:             0.0,
		Currency:           "USD",
		NextBillingDate:    &[]time.Time{time.Now().AddDate(0, 1, 0)}[0],
		Status:             "active",
		IsTrialActive:      true,
		TrialStart:         &[]time.Time{time.Now()}[0],
		TrialEnd:           &[]time.Time{time.Now().AddDate(0, 0, 14)}[0], // 14-day trial
		CurrentPeriodStart: &[]time.Time{time.Now()}[0],
		CurrentPeriodEnd:   &[]time.Time{time.Now().AddDate(0, 1, 0)}[0],
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
func (s *TenantAdminService) GetTenant(id uint) (*models.Tenant, error) {
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
func (s *TenantAdminService) GetTenantBySlug(slug string) (*models.Tenant, error) {
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
func (s *TenantAdminService) GetTenantByDomain(domain string) (*models.Tenant, error) {
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
func (s *TenantAdminService) GetTenants(limit, offset int) ([]*models.Tenant, int64, error) {
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
func (s *TenantAdminService) UpdateTenant(tenant *models.Tenant) error {
	if err := s.db.Save(tenant).Error; err != nil {
		s.logger.Error("Failed to update tenant", zap.Error(err))
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	s.logger.Info("Tenant updated successfully", zap.Uint("tenant_id", tenant.ID))
	return nil
}

// DeleteTenant soft deletes a tenant.
func (s *TenantAdminService) DeleteTenant(id uint) error {
	if err := s.db.Delete(&models.Tenant{}, id).Error; err != nil {
		s.logger.Error("Failed to delete tenant", zap.Error(err))
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	s.logger.Info("Tenant deleted successfully", zap.Uint("tenant_id", id))
	return nil
}

// GetTenantBilling retrieves tenant billing information.
func (s *TenantAdminService) GetTenantBilling(tenantID uint) (*models.TenantBilling, error) {
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
func (s *TenantAdminService) UpdateTenantBilling(billing *models.TenantBilling) error {
	if err := s.db.Save(billing).Error; err != nil {
		s.logger.Error("Failed to update tenant billing", zap.Error(err))
		return fmt.Errorf("failed to update tenant billing: %w", err)
	}

	s.logger.Info("Tenant billing updated successfully", zap.Uint("tenant_id", billing.TenantID))
	return nil
}

// GetTenantBranding retrieves tenant branding information.
func (s *TenantAdminService) GetTenantBranding(tenantID uint) (*models.TenantBranding, error) {
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
func (s *TenantAdminService) UpdateTenantBranding(branding *models.TenantBranding) error {
	if err := s.db.Save(branding).Error; err != nil {
		s.logger.Error("Failed to update tenant branding", zap.Error(err))
		return fmt.Errorf("failed to update tenant branding: %w", err)
	}

	s.logger.Info("Tenant branding updated successfully", zap.Uint("tenant_id", branding.TenantID))
	return nil
}

// LogTenantActivity logs tenant activity.
func (s *TenantAdminService) LogTenantActivity(activity *models.TenantActivity) error {
	if err := s.db.Create(activity).Error; err != nil {
		s.logger.Error("Failed to log tenant activity", zap.Error(err))
		return fmt.Errorf("failed to log tenant activity: %w", err)
	}

	return nil
}

// generateSlug generates a URL-friendly slug from a string.
func (s *TenantAdminService) generateSlug(input string) string {
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

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	// Ensure minimum length
	if len(slug) == 0 {
		slug = "tenant"
	}

	return slug
}
