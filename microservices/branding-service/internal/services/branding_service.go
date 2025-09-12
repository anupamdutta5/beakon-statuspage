// Package services provides business logic for the Branding Service.
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage-branding-service/internal/config"
	"github.com/enterprise-status/statuspage-branding-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// BrandingService handles branding-related business logic.
type BrandingService struct {
	config *config.Config
	logger *zap.Logger
	db     *gorm.DB
}

// NewBrandingService creates a new branding service.
func NewBrandingService(cfg *config.Config, logger *zap.Logger) (*BrandingService, error) {
	// Initialize database connection
	db, err := initDatabase(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &BrandingService{
		config: cfg,
		logger: logger,
		db:     db,
	}, nil
}

// Brand Management

// CreateBrand creates a new brand.
func (s *BrandingService) CreateBrand(ctx context.Context, brand *models.Brand) error {
	s.logger.Info("Creating brand",
		zap.String("brand_name", brand.Name),
		zap.String("brand_slug", brand.Slug))

	if err := s.db.Create(brand).Error; err != nil {
		s.logger.Error("Failed to create brand", zap.Error(err))
		return fmt.Errorf("failed to create brand: %w", err)
	}

	s.logger.Info("Brand created successfully",
		zap.Uint("brand_id", brand.ID),
		zap.String("brand_name", brand.Name))

	return nil
}

// GetBrand retrieves a brand by ID.
func (s *BrandingService) GetBrand(ctx context.Context, brandID uint) (*models.Brand, error) {
	s.logger.Info("Getting brand", zap.Uint("brand_id", brandID))

	var brand models.Brand
	if err := s.db.Preload("Themes").First(&brand, brandID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("brand not found")
		}
		s.logger.Error("Failed to get brand", zap.Error(err))
		return nil, fmt.Errorf("failed to get brand: %w", err)
	}

	return &brand, nil
}

// GetBrandBySlug retrieves a brand by slug.
func (s *BrandingService) GetBrandBySlug(ctx context.Context, slug string) (*models.Brand, error) {
	s.logger.Info("Getting brand by slug", zap.String("slug", slug))

	var brand models.Brand
	if err := s.db.Where("slug = ?", slug).Preload("Themes").First(&brand).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("brand not found")
		}
		s.logger.Error("Failed to get brand by slug", zap.Error(err))
		return nil, fmt.Errorf("failed to get brand by slug: %w", err)
	}

	return &brand, nil
}

// ListBrands lists all brands.
func (s *BrandingService) ListBrands(ctx context.Context, limit, offset int) ([]*models.Brand, error) {
	s.logger.Info("Listing brands", zap.Int("limit", limit), zap.Int("offset", offset))

	var brands []*models.Brand
	query := s.db.Preload("Themes")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&brands).Error; err != nil {
		s.logger.Error("Failed to list brands", zap.Error(err))
		return nil, fmt.Errorf("failed to list brands: %w", err)
	}

	return brands, nil
}

// UpdateBrand updates a brand.
func (s *BrandingService) UpdateBrand(ctx context.Context, brandID uint, updates *models.Brand) error {
	s.logger.Info("Updating brand", zap.Uint("brand_id", brandID))

	if err := s.db.Model(&models.Brand{}).Where("id = ?", brandID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update brand", zap.Error(err))
		return fmt.Errorf("failed to update brand: %w", err)
	}

	s.logger.Info("Brand updated successfully", zap.Uint("brand_id", brandID))
	return nil
}

// DeleteBrand deletes a brand.
func (s *BrandingService) DeleteBrand(ctx context.Context, brandID uint) error {
	s.logger.Info("Deleting brand", zap.Uint("brand_id", brandID))

	if err := s.db.Delete(&models.Brand{}, brandID).Error; err != nil {
		s.logger.Error("Failed to delete brand", zap.Error(err))
		return fmt.Errorf("failed to delete brand: %w", err)
	}

	s.logger.Info("Brand deleted successfully", zap.Uint("brand_id", brandID))
	return nil
}

// Theme Management

// CreateTheme creates a new theme.
func (s *BrandingService) CreateTheme(ctx context.Context, theme *models.Theme) error {
	s.logger.Info("Creating theme",
		zap.String("theme_name", theme.Name),
		zap.Uint("brand_id", theme.BrandID))

	if err := s.db.Create(theme).Error; err != nil {
		s.logger.Error("Failed to create theme", zap.Error(err))
		return fmt.Errorf("failed to create theme: %w", err)
	}

	s.logger.Info("Theme created successfully",
		zap.Uint("theme_id", theme.ID),
		zap.String("theme_name", theme.Name))

	return nil
}

// GetTheme retrieves a theme by ID.
func (s *BrandingService) GetTheme(ctx context.Context, themeID uint) (*models.Theme, error) {
	s.logger.Info("Getting theme", zap.Uint("theme_id", themeID))

	var theme models.Theme
	if err := s.db.Preload("Brand").Preload("ColorSchemes").Preload("Typographies").Preload("Layouts").Preload("Components").First(&theme, themeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("theme not found")
		}
		s.logger.Error("Failed to get theme", zap.Error(err))
		return nil, fmt.Errorf("failed to get theme: %w", err)
	}

	return &theme, nil
}

// ListThemes lists all themes for a brand.
func (s *BrandingService) ListThemes(ctx context.Context, brandID uint, limit, offset int) ([]*models.Theme, error) {
	s.logger.Info("Listing themes", zap.Uint("brand_id", brandID), zap.Int("limit", limit), zap.Int("offset", offset))

	var themes []*models.Theme
	query := s.db.Where("brand_id = ?", brandID).Preload("Brand")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&themes).Error; err != nil {
		s.logger.Error("Failed to list themes", zap.Error(err))
		return nil, fmt.Errorf("failed to list themes: %w", err)
	}

	return themes, nil
}

// UpdateTheme updates a theme.
func (s *BrandingService) UpdateTheme(ctx context.Context, themeID uint, updates *models.Theme) error {
	s.logger.Info("Updating theme", zap.Uint("theme_id", themeID))

	if err := s.db.Model(&models.Theme{}).Where("id = ?", themeID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update theme", zap.Error(err))
		return fmt.Errorf("failed to update theme: %w", err)
	}

	s.logger.Info("Theme updated successfully", zap.Uint("theme_id", themeID))
	return nil
}

// DeleteTheme deletes a theme.
func (s *BrandingService) DeleteTheme(ctx context.Context, themeID uint) error {
	s.logger.Info("Deleting theme", zap.Uint("theme_id", themeID))

	if err := s.db.Delete(&models.Theme{}, themeID).Error; err != nil {
		s.logger.Error("Failed to delete theme", zap.Error(err))
		return fmt.Errorf("failed to delete theme: %w", err)
	}

	s.logger.Info("Theme deleted successfully", zap.Uint("theme_id", themeID))
	return nil
}

// Asset Management

// CreateAsset creates a new asset.
func (s *BrandingService) CreateAsset(ctx context.Context, asset *models.Asset) error {
	s.logger.Info("Creating asset",
		zap.String("asset_name", asset.Name),
		zap.String("asset_type", asset.Type),
		zap.Uint("brand_id", asset.BrandID))

	if err := s.db.Create(asset).Error; err != nil {
		s.logger.Error("Failed to create asset", zap.Error(err))
		return fmt.Errorf("failed to create asset: %w", err)
	}

	s.logger.Info("Asset created successfully",
		zap.Uint("asset_id", asset.ID),
		zap.String("asset_name", asset.Name))

	return nil
}

// GetAsset retrieves an asset by ID.
func (s *BrandingService) GetAsset(ctx context.Context, assetID uint) (*models.Asset, error) {
	s.logger.Info("Getting asset", zap.Uint("asset_id", assetID))

	var asset models.Asset
	if err := s.db.Preload("Brand").First(&asset, assetID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("asset not found")
		}
		s.logger.Error("Failed to get asset", zap.Error(err))
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	return &asset, nil
}

// ListAssets lists all assets for a brand.
func (s *BrandingService) ListAssets(ctx context.Context, brandID uint, assetType string, limit, offset int) ([]*models.Asset, error) {
	s.logger.Info("Listing assets",
		zap.Uint("brand_id", brandID),
		zap.String("asset_type", assetType),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	var assets []*models.Asset
	query := s.db.Where("brand_id = ?", brandID).Preload("Brand")

	if assetType != "" {
		query = query.Where("type = ?", assetType)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&assets).Error; err != nil {
		s.logger.Error("Failed to list assets", zap.Error(err))
		return nil, fmt.Errorf("failed to list assets: %w", err)
	}

	return assets, nil
}

// UpdateAsset updates an asset.
func (s *BrandingService) UpdateAsset(ctx context.Context, assetID uint, updates *models.Asset) error {
	s.logger.Info("Updating asset", zap.Uint("asset_id", assetID))

	if err := s.db.Model(&models.Asset{}).Where("id = ?", assetID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update asset", zap.Error(err))
		return fmt.Errorf("failed to update asset: %w", err)
	}

	s.logger.Info("Asset updated successfully", zap.Uint("asset_id", assetID))
	return nil
}

// DeleteAsset deletes an asset.
func (s *BrandingService) DeleteAsset(ctx context.Context, assetID uint) error {
	s.logger.Info("Deleting asset", zap.Uint("asset_id", assetID))

	if err := s.db.Delete(&models.Asset{}, assetID).Error; err != nil {
		s.logger.Error("Failed to delete asset", zap.Error(err))
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	s.logger.Info("Asset deleted successfully", zap.Uint("asset_id", assetID))
	return nil
}

// Custom CSS Management

// CreateCustomCSS creates custom CSS.
func (s *BrandingService) CreateCustomCSS(ctx context.Context, customCSS *models.CustomCSS) error {
	s.logger.Info("Creating custom CSS",
		zap.String("css_name", customCSS.Name),
		zap.Uint("brand_id", customCSS.BrandID))

	if err := s.db.Create(customCSS).Error; err != nil {
		s.logger.Error("Failed to create custom CSS", zap.Error(err))
		return fmt.Errorf("failed to create custom CSS: %w", err)
	}

	s.logger.Info("Custom CSS created successfully",
		zap.Uint("css_id", customCSS.ID),
		zap.String("css_name", customCSS.Name))

	return nil
}

// GetCustomCSS retrieves custom CSS by ID.
func (s *BrandingService) GetCustomCSS(ctx context.Context, cssID uint) (*models.CustomCSS, error) {
	s.logger.Info("Getting custom CSS", zap.Uint("css_id", cssID))

	var customCSS models.CustomCSS
	if err := s.db.Preload("Brand").First(&customCSS, cssID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("custom CSS not found")
		}
		s.logger.Error("Failed to get custom CSS", zap.Error(err))
		return nil, fmt.Errorf("failed to get custom CSS: %w", err)
	}

	return &customCSS, nil
}

// ListCustomCSS lists all custom CSS for a brand.
func (s *BrandingService) ListCustomCSS(ctx context.Context, brandID uint, limit, offset int) ([]*models.CustomCSS, error) {
	s.logger.Info("Listing custom CSS", zap.Uint("brand_id", brandID), zap.Int("limit", limit), zap.Int("offset", offset))

	var customCSS []*models.CustomCSS
	query := s.db.Where("brand_id = ?", brandID).Preload("Brand")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&customCSS).Error; err != nil {
		s.logger.Error("Failed to list custom CSS", zap.Error(err))
		return nil, fmt.Errorf("failed to list custom CSS: %w", err)
	}

	return customCSS, nil
}

// UpdateCustomCSS updates custom CSS.
func (s *BrandingService) UpdateCustomCSS(ctx context.Context, cssID uint, updates *models.CustomCSS) error {
	s.logger.Info("Updating custom CSS", zap.Uint("css_id", cssID))

	if err := s.db.Model(&models.CustomCSS{}).Where("id = ?", cssID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update custom CSS", zap.Error(err))
		return fmt.Errorf("failed to update custom CSS: %w", err)
	}

	s.logger.Info("Custom CSS updated successfully", zap.Uint("css_id", cssID))
	return nil
}

// DeleteCustomCSS deletes custom CSS.
func (s *BrandingService) DeleteCustomCSS(ctx context.Context, cssID uint) error {
	s.logger.Info("Deleting custom CSS", zap.Uint("css_id", cssID))

	if err := s.db.Delete(&models.CustomCSS{}, cssID).Error; err != nil {
		s.logger.Error("Failed to delete custom CSS", zap.Error(err))
		return fmt.Errorf("failed to delete custom CSS: %w", err)
	}

	s.logger.Info("Custom CSS deleted successfully", zap.Uint("css_id", cssID))
	return nil
}

// Statistics

// GetBrandingStats returns branding statistics.
func (s *BrandingService) GetBrandingStats(ctx context.Context) (*models.BrandingStats, error) {
	s.logger.Info("Getting branding statistics")

	// For now, we'll return simulated statistics
	// In production, you would query actual statistics from the database

	stats := &models.BrandingStats{
		TotalBrands:     50,
		TotalThemes:     200,
		TotalAssets:     1000,
		TotalCustomCSS:  150,
		TotalCustomJS:   100,
		TotalLayouts:    75,
		TotalComponents: 300,
		LastUpdated:     time.Now(),
	}

	return stats, nil
}

// Health checks the health of the branding service.
func (s *BrandingService) Health(ctx context.Context) error {
	s.logger.Debug("Checking branding service health")

	// Check database connection
	if err := s.db.Exec("SELECT 1").Error; err != nil {
		s.logger.Error("Branding service health check failed", zap.Error(err))
		return fmt.Errorf("branding service health check failed: %w", err)
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
		&models.Brand{},
		&models.Theme{},
		&models.ColorScheme{},
		&models.Typography{},
		&models.Asset{},
		&models.CustomCSS{},
		&models.CustomJS{},
		&models.Layout{},
		&models.Component{},
		&models.BrandingStats{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

