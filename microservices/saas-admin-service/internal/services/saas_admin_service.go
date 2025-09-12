// Package services provides business logic for the SaaS Admin Service.
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage-saas-admin-service/internal/config"
	"github.com/enterprise-status/statuspage-saas-admin-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SaaSAdminService handles SaaS admin-related business logic.
type SaaSAdminService struct {
	config *config.Config
	logger *zap.Logger
	db     *gorm.DB
}

// NewSaaSAdminService creates a new SaaS admin service.
func NewSaaSAdminService(cfg *config.Config, logger *zap.Logger) (*SaaSAdminService, error) {
	// Initialize database connection
	db, err := initDatabase(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &SaaSAdminService{
		config: cfg,
		logger: logger,
		db:     db,
	}, nil
}

// Platform Management

// GetPlatform retrieves platform configuration.
func (s *SaaSAdminService) GetPlatform(ctx context.Context) (*models.Platform, error) {
	s.logger.Info("Getting platform configuration")

	var platform models.Platform
	if err := s.db.First(&platform).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create default platform if not exists
			platform = models.Platform{
				Name:         s.config.SaaS.PlatformName,
				URL:          s.config.SaaS.PlatformURL,
				AdminEmail:   s.config.SaaS.AdminEmail,
				SupportEmail: s.config.SaaS.SupportEmail,
				Version:      s.config.Service.Version,
				Status:       "active",
			}
			if err := s.db.Create(&platform).Error; err != nil {
				s.logger.Error("Failed to create default platform", zap.Error(err))
				return nil, fmt.Errorf("failed to create default platform: %w", err)
			}
		} else {
			s.logger.Error("Failed to get platform", zap.Error(err))
			return nil, fmt.Errorf("failed to get platform: %w", err)
		}
	}

	return &platform, nil
}

// UpdatePlatform updates platform configuration.
func (s *SaaSAdminService) UpdatePlatform(ctx context.Context, updates *models.Platform) error {
	s.logger.Info("Updating platform configuration")

	if err := s.db.Model(&models.Platform{}).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update platform", zap.Error(err))
		return fmt.Errorf("failed to update platform: %w", err)
	}

	s.logger.Info("Platform updated successfully")
	return nil
}

// Plan Management

// CreatePlan creates a new SaaS plan.
func (s *SaaSAdminService) CreatePlan(ctx context.Context, plan *models.SaaSPlan) error {
	s.logger.Info("Creating SaaS plan",
		zap.String("plan_name", plan.Name),
		zap.String("plan_slug", plan.Slug))

	if err := s.db.Create(plan).Error; err != nil {
		s.logger.Error("Failed to create plan", zap.Error(err))
		return fmt.Errorf("failed to create plan: %w", err)
	}

	s.logger.Info("SaaS plan created successfully",
		zap.Uint("plan_id", plan.ID),
		zap.String("plan_name", plan.Name))

	return nil
}

// GetPlan retrieves a plan by ID.
func (s *SaaSAdminService) GetPlan(ctx context.Context, planID uint) (*models.SaaSPlan, error) {
	s.logger.Info("Getting SaaS plan", zap.Uint("plan_id", planID))

	var plan models.SaaSPlan
	if err := s.db.First(&plan, planID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("plan not found")
		}
		s.logger.Error("Failed to get plan", zap.Error(err))
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	return &plan, nil
}

// GetPlanBySlug retrieves a plan by slug.
func (s *SaaSAdminService) GetPlanBySlug(ctx context.Context, slug string) (*models.SaaSPlan, error) {
	s.logger.Info("Getting SaaS plan by slug", zap.String("slug", slug))

	var plan models.SaaSPlan
	if err := s.db.Where("slug = ?", slug).First(&plan).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("plan not found")
		}
		s.logger.Error("Failed to get plan by slug", zap.Error(err))
		return nil, fmt.Errorf("failed to get plan by slug: %w", err)
	}

	return &plan, nil
}

// ListPlans lists all SaaS plans.
func (s *SaaSAdminService) ListPlans(ctx context.Context, limit, offset int) ([]*models.SaaSPlan, error) {
	s.logger.Info("Listing SaaS plans", zap.Int("limit", limit), zap.Int("offset", offset))

	var plans []*models.SaaSPlan
	query := s.db

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&plans).Error; err != nil {
		s.logger.Error("Failed to list plans", zap.Error(err))
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}

	return plans, nil
}

// UpdatePlan updates a plan.
func (s *SaaSAdminService) UpdatePlan(ctx context.Context, planID uint, updates *models.SaaSPlan) error {
	s.logger.Info("Updating SaaS plan", zap.Uint("plan_id", planID))

	if err := s.db.Model(&models.SaaSPlan{}).Where("id = ?", planID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update plan", zap.Error(err))
		return fmt.Errorf("failed to update plan: %w", err)
	}

	s.logger.Info("SaaS plan updated successfully", zap.Uint("plan_id", planID))
	return nil
}

// DeletePlan deletes a plan.
func (s *SaaSAdminService) DeletePlan(ctx context.Context, planID uint) error {
	s.logger.Info("Deleting SaaS plan", zap.Uint("plan_id", planID))

	if err := s.db.Delete(&models.SaaSPlan{}, planID).Error; err != nil {
		s.logger.Error("Failed to delete plan", zap.Error(err))
		return fmt.Errorf("failed to delete plan: %w", err)
	}

	s.logger.Info("SaaS plan deleted successfully", zap.Uint("plan_id", planID))
	return nil
}

// Feature Management

// CreateFeature creates a new SaaS feature.
func (s *SaaSAdminService) CreateFeature(ctx context.Context, feature *models.SaaSFeature) error {
	s.logger.Info("Creating SaaS feature",
		zap.String("feature_name", feature.Name),
		zap.String("feature_slug", feature.Slug))

	if err := s.db.Create(feature).Error; err != nil {
		s.logger.Error("Failed to create feature", zap.Error(err))
		return fmt.Errorf("failed to create feature: %w", err)
	}

	s.logger.Info("SaaS feature created successfully",
		zap.Uint("feature_id", feature.ID),
		zap.String("feature_name", feature.Name))

	return nil
}

// GetFeature retrieves a feature by ID.
func (s *SaaSAdminService) GetFeature(ctx context.Context, featureID uint) (*models.SaaSFeature, error) {
	s.logger.Info("Getting SaaS feature", zap.Uint("feature_id", featureID))

	var feature models.SaaSFeature
	if err := s.db.First(&feature, featureID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("feature not found")
		}
		s.logger.Error("Failed to get feature", zap.Error(err))
		return nil, fmt.Errorf("failed to get feature: %w", err)
	}

	return &feature, nil
}

// ListFeatures lists all SaaS features.
func (s *SaaSAdminService) ListFeatures(ctx context.Context, category string, limit, offset int) ([]*models.SaaSFeature, error) {
	s.logger.Info("Listing SaaS features",
		zap.String("category", category),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	var features []*models.SaaSFeature
	query := s.db

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&features).Error; err != nil {
		s.logger.Error("Failed to list features", zap.Error(err))
		return nil, fmt.Errorf("failed to list features: %w", err)
	}

	return features, nil
}

// UpdateFeature updates a feature.
func (s *SaaSAdminService) UpdateFeature(ctx context.Context, featureID uint, updates *models.SaaSFeature) error {
	s.logger.Info("Updating SaaS feature", zap.Uint("feature_id", featureID))

	if err := s.db.Model(&models.SaaSFeature{}).Where("id = ?", featureID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update feature", zap.Error(err))
		return fmt.Errorf("failed to update feature: %w", err)
	}

	s.logger.Info("SaaS feature updated successfully", zap.Uint("feature_id", featureID))
	return nil
}

// DeleteFeature deletes a feature.
func (s *SaaSAdminService) DeleteFeature(ctx context.Context, featureID uint) error {
	s.logger.Info("Deleting SaaS feature", zap.Uint("feature_id", featureID))

	if err := s.db.Delete(&models.SaaSFeature{}, featureID).Error; err != nil {
		s.logger.Error("Failed to delete feature", zap.Error(err))
		return fmt.Errorf("failed to delete feature: %w", err)
	}

	s.logger.Info("SaaS feature deleted successfully", zap.Uint("feature_id", featureID))
	return nil
}

// Feature Flag Management

// CreateFeatureFlag creates a new feature flag.
func (s *SaaSAdminService) CreateFeatureFlag(ctx context.Context, flag *models.SaaSFeatureFlag) error {
	s.logger.Info("Creating SaaS feature flag",
		zap.String("flag_name", flag.Name))

	if err := s.db.Create(flag).Error; err != nil {
		s.logger.Error("Failed to create feature flag", zap.Error(err))
		return fmt.Errorf("failed to create feature flag: %w", err)
	}

	s.logger.Info("SaaS feature flag created successfully",
		zap.Uint("flag_id", flag.ID),
		zap.String("flag_name", flag.Name))

	return nil
}

// GetFeatureFlag retrieves a feature flag by ID.
func (s *SaaSAdminService) GetFeatureFlag(ctx context.Context, flagID uint) (*models.SaaSFeatureFlag, error) {
	s.logger.Info("Getting SaaS feature flag", zap.Uint("flag_id", flagID))

	var flag models.SaaSFeatureFlag
	if err := s.db.First(&flag, flagID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("feature flag not found")
		}
		s.logger.Error("Failed to get feature flag", zap.Error(err))
		return nil, fmt.Errorf("failed to get feature flag: %w", err)
	}

	return &flag, nil
}

// ListFeatureFlags lists all feature flags.
func (s *SaaSAdminService) ListFeatureFlags(ctx context.Context, limit, offset int) ([]*models.SaaSFeatureFlag, error) {
	s.logger.Info("Listing SaaS feature flags", zap.Int("limit", limit), zap.Int("offset", offset))

	var flags []*models.SaaSFeatureFlag
	query := s.db

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&flags).Error; err != nil {
		s.logger.Error("Failed to list feature flags", zap.Error(err))
		return nil, fmt.Errorf("failed to list feature flags: %w", err)
	}

	return flags, nil
}

// UpdateFeatureFlag updates a feature flag.
func (s *SaaSAdminService) UpdateFeatureFlag(ctx context.Context, flagID uint, updates *models.SaaSFeatureFlag) error {
	s.logger.Info("Updating SaaS feature flag", zap.Uint("flag_id", flagID))

	if err := s.db.Model(&models.SaaSFeatureFlag{}).Where("id = ?", flagID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update feature flag", zap.Error(err))
		return fmt.Errorf("failed to update feature flag: %w", err)
	}

	s.logger.Info("SaaS feature flag updated successfully", zap.Uint("flag_id", flagID))
	return nil
}

// DeleteFeatureFlag deletes a feature flag.
func (s *SaaSAdminService) DeleteFeatureFlag(ctx context.Context, flagID uint) error {
	s.logger.Info("Deleting SaaS feature flag", zap.Uint("flag_id", flagID))

	if err := s.db.Delete(&models.SaaSFeatureFlag{}, flagID).Error; err != nil {
		s.logger.Error("Failed to delete feature flag", zap.Error(err))
		return fmt.Errorf("failed to delete feature flag: %w", err)
	}

	s.logger.Info("SaaS feature flag deleted successfully", zap.Uint("flag_id", flagID))
	return nil
}

// Statistics

// GetSaaSStats returns SaaS platform statistics.
func (s *SaaSAdminService) GetSaaSStats(ctx context.Context) (*models.SaaSStats, error) {
	s.logger.Info("Getting SaaS platform statistics")

	// For now, we'll return simulated statistics
	// In production, you would query actual statistics from the database

	stats := &models.SaaSStats{
		TotalTenants:   150,
		ActiveTenants:  120,
		TotalUsers:     2500,
		ActiveUsers:    2100,
		TotalRevenue:   125000.00,
		MonthlyRevenue: 15000.00,
		TotalPlans:     5,
		ActivePlans:    4,
		TotalFeatures:  25,
		ActiveFeatures: 22,
		LastUpdated:    time.Now(),
	}

	return stats, nil
}

// Health checks the health of the SaaS admin service.
func (s *SaaSAdminService) Health(ctx context.Context) error {
	s.logger.Debug("Checking SaaS admin service health")

	// Check database connection
	if err := s.db.Exec("SELECT 1").Error; err != nil {
		s.logger.Error("SaaS admin service health check failed", zap.Error(err))
		return fmt.Errorf("saas admin service health check failed: %w", err)
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
		&models.Platform{},
		&models.SaaSPlan{},
		&models.SaaSFeature{},
		&models.SaaSFeatureFlag{},
		&models.SaaSAdminUser{},
		&models.SaaSNotification{},
		&models.SaaSActivity{},
		&models.SaaSBackup{},
		&models.SaaSStats{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

