// Package services provides business logic for the Tenant Admin Service.
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage-tenant-admin-service/internal/config"
	"github.com/enterprise-status/statuspage-tenant-admin-service/internal/models"
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
