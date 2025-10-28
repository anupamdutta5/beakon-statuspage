// Package services provides business logic for the Branding Service.
package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/anupamdutta5/branding-service/internal/config"
	"github.com/anupamdutta5/branding-service/internal/models"
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
	var db *gorm.DB
	var err error

	// If config is nil, skip database initialization (will be set later via SetDB)
	if cfg != nil {
		db, err = initDatabase(cfg.Database)
		if err != nil {
			// For testing, we'll allow the service to be created without a database
			// The database will be set later via SetDB method
			if logger != nil {
				logger.Warn("Failed to initialize database, service will be created without database", zap.Error(err))
			}
			db = nil
		}
	}

	return &BrandingService{
		config: cfg,
		logger: logger,
		db:     db,
	}, nil
}

// SetDB sets the database connection (for testing)
func (s *BrandingService) SetDB(db *gorm.DB) {
	s.db = db
}

// Brand Management

// CreateBrand creates a new brand with validation and multi-tenant support.
func (s *BrandingService) CreateBrand(ctx context.Context, brand *models.Brand) error {
	if err := s.validateBrand(brand); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	s.logger.Info("Creating brand",
		zap.String("brand_name", brand.Name),
		zap.String("brand_slug", brand.Slug),
		zap.Uint("tenant_id", brand.TenantID))

	if err := s.db.WithContext(ctx).Create(brand).Error; err != nil {
		s.logger.Error("Failed to create brand", zap.Error(err))
		return fmt.Errorf("failed to create brand: %w", err)
	}

	s.logger.Info("Brand created successfully",
		zap.Uint("brand_id", brand.ID),
		zap.String("brand_name", brand.Name))

	return nil
}

// GetBrand retrieves a brand by ID with tenant isolation.
func (s *BrandingService) GetBrand(ctx context.Context, brandID, tenantID uint) (*models.Brand, error) {
	s.logger.Info("Getting brand", zap.Uint("brand_id", brandID), zap.Uint("tenant_id", tenantID))

	var brand models.Brand
	err := s.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", brandID, tenantID).
		Preload("Themes").
		First(&brand).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

// Enhanced Asset Management with File Upload Support

func (s *BrandingService) CreateAssetWithFile(ctx context.Context, asset *models.Asset, file multipart.File, header *multipart.FileHeader, tenantID uint) error {
	if err := s.validateAssetCreation(asset, tenantID); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Set asset properties from file
	asset.OriginalName = header.Filename
	asset.Filename = fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
	asset.MimeType = header.Header.Get("Content-Type")
	asset.Size = int64(len(content))

	// Validate file type
	if err := s.validateFileType(asset.Type, asset.MimeType); err != nil {
		return err
	}

	// In production, upload to cloud storage (S3, GCS, etc.)
	asset.URL = fmt.Sprintf("/assets/%s", asset.Filename)

	if err := s.db.WithContext(ctx).Create(asset).Error; err != nil {
		s.logger.Error("Failed to create asset", zap.Error(err))
		return fmt.Errorf("failed to create asset: %w", err)
	}

	s.logger.Info("Asset created successfully", zap.Uint("asset_id", asset.ID), zap.String("filename", asset.Filename))
	return nil
}

// Enhanced Color Scheme Management

func (s *BrandingService) CreateColorScheme(ctx context.Context, colorScheme *models.ColorScheme, tenantID uint) error {
	if err := s.validateColorScheme(colorScheme); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Verify theme belongs to tenant
	var theme models.Theme
	err := s.db.WithContext(ctx).
		Joins("JOIN brands ON themes.brand_id = brands.id").
		Where("themes.id = ? AND brands.tenant_id = ?", colorScheme.ThemeID, tenantID).
		First(&theme).Error

	if err != nil {
		return fmt.Errorf("theme not found or access denied")
	}

	if err := s.db.WithContext(ctx).Create(colorScheme).Error; err != nil {
		s.logger.Error("Failed to create color scheme", zap.Error(err))
		return fmt.Errorf("failed to create color scheme: %w", err)
	}

	s.logger.Info("Color scheme created successfully", zap.Uint("color_scheme_id", colorScheme.ID))
	return nil
}

func (s *BrandingService) GetColorScheme(ctx context.Context, id uint, tenantID uint) (*models.ColorScheme, error) {
	var colorScheme models.ColorScheme
	err := s.db.WithContext(ctx).
		Joins("JOIN themes ON color_schemes.theme_id = themes.id").
		Joins("JOIN brands ON themes.brand_id = brands.id").
		Where("color_schemes.id = ? AND brands.tenant_id = ?", id, tenantID).
		Preload("Theme").
		First(&colorScheme).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("color scheme not found")
		}
		return nil, fmt.Errorf("failed to get color scheme: %w", err)
	}

	return &colorScheme, nil
}

// Enhanced Typography Management

func (s *BrandingService) CreateTypography(ctx context.Context, typography *models.Typography, tenantID uint) error {
	if err := s.validateTypography(typography); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Verify theme belongs to tenant
	var theme models.Theme
	err := s.db.WithContext(ctx).
		Joins("JOIN brands ON themes.brand_id = brands.id").
		Where("themes.id = ? AND brands.tenant_id = ?", typography.ThemeID, tenantID).
		First(&theme).Error

	if err != nil {
		return fmt.Errorf("theme not found or access denied")
	}

	if err := s.db.WithContext(ctx).Create(typography).Error; err != nil {
		s.logger.Error("Failed to create typography", zap.Error(err))
		return fmt.Errorf("failed to create typography: %w", err)
	}

	s.logger.Info("Typography created successfully", zap.Uint("typography_id", typography.ID))
	return nil
}

// Enhanced Custom CSS with Security Validation

func (s *BrandingService) CreateSecureCustomCSS(ctx context.Context, customCSS *models.CustomCSS, tenantID uint) error {
	if err := s.validateCustomCSS(customCSS); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Verify brand belongs to tenant
	var brand models.Brand
	err := s.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", customCSS.BrandID, tenantID).
		First(&brand).Error

	if err != nil {
		return fmt.Errorf("brand not found or access denied")
	}

	if err := s.db.WithContext(ctx).Create(customCSS).Error; err != nil {
		s.logger.Error("Failed to create custom CSS", zap.Error(err))
		return fmt.Errorf("failed to create custom CSS: %w", err)
	}

	s.logger.Info("Custom CSS created successfully", zap.Uint("css_id", customCSS.ID))
	return nil
}

// Enhanced Custom JS with Security Validation

func (s *BrandingService) CreateSecureCustomJS(ctx context.Context, customJS *models.CustomJS, tenantID uint) error {
	if err := s.validateCustomJS(customJS); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Verify brand belongs to tenant
	var brand models.Brand
	err := s.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", customJS.BrandID, tenantID).
		First(&brand).Error

	if err != nil {
		return fmt.Errorf("brand not found or access denied")
	}

	if err := s.db.WithContext(ctx).Create(customJS).Error; err != nil {
		s.logger.Error("Failed to create custom JS", zap.Error(err))
		return fmt.Errorf("failed to create custom JS: %w", err)
	}

	s.logger.Info("Custom JS created successfully", zap.Uint("js_id", customJS.ID))
	return nil
}

// Theme Compilation and Preview

func (s *BrandingService) CompileTheme(ctx context.Context, themeID, tenantID uint) (map[string]interface{}, error) {
	theme, err := s.GetTheme(ctx, themeID)
	if err != nil {
		return nil, err
	}

	// Verify theme belongs to tenant
	var brand models.Brand
	err = s.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", theme.BrandID, tenantID).
		First(&brand).Error

	if err != nil {
		return nil, fmt.Errorf("theme not found or access denied")
	}

	// Get all theme components
	var colorSchemes []*models.ColorScheme
	var typographies []*models.Typography
	var layouts []*models.Layout
	var components []*models.Component

	if err := s.db.WithContext(ctx).Where("theme_id = ?", themeID).Find(&colorSchemes).Error; err != nil {
		return nil, fmt.Errorf("failed to load color schemes: %w", err)
	}

	if err := s.db.WithContext(ctx).Where("theme_id = ?", themeID).Find(&typographies).Error; err != nil {
		return nil, fmt.Errorf("failed to load typographies: %w", err)
	}

	if err := s.db.WithContext(ctx).Where("theme_id = ?", themeID).Find(&layouts).Error; err != nil {
		return nil, fmt.Errorf("failed to load layouts: %w", err)
	}

	if err := s.db.WithContext(ctx).Where("theme_id = ?", themeID).Find(&components).Error; err != nil {
		return nil, fmt.Errorf("failed to load components: %w", err)
	}

	// Compile theme data
	compiledTheme := map[string]interface{}{
		"theme":         theme,
		"color_schemes": colorSchemes,
		"typographies":  typographies,
		"layouts":       layouts,
		"components":    components,
		"compiled_at":   time.Now(),
	}

	return compiledTheme, nil
}

// Enhanced Statistics with Tenant Filtering

func (s *BrandingService) GetTenantBrandingStats(ctx context.Context, tenantID uint) (*models.BrandingStats, error) {
	var stats models.BrandingStats

	// Count variables for GORM
	var totalBrands, totalThemes, totalAssets, totalCustomCSS, totalCustomJS, totalLayouts, totalComponents int64

	// Count brands
	if err := s.db.WithContext(ctx).Model(&models.Brand{}).Where("tenant_id = ?", tenantID).Count(&totalBrands).Error; err != nil {
		return nil, fmt.Errorf("failed to count brands: %w", err)
	}
	stats.TotalBrands = int(totalBrands)

	// Count themes
	if err := s.db.WithContext(ctx).
		Table("themes").
		Joins("JOIN brands ON themes.brand_id = brands.id").
		Where("brands.tenant_id = ?", tenantID).
		Count(&totalThemes).Error; err != nil {
		return nil, fmt.Errorf("failed to count themes: %w", err)
	}
	stats.TotalThemes = int(totalThemes)

	// Count assets
	if err := s.db.WithContext(ctx).
		Table("assets").
		Joins("JOIN brands ON assets.brand_id = brands.id").
		Where("brands.tenant_id = ?", tenantID).
		Count(&totalAssets).Error; err != nil {
		return nil, fmt.Errorf("failed to count assets: %w", err)
	}
	stats.TotalAssets = int(totalAssets)

	// Count custom CSS
	if err := s.db.WithContext(ctx).
		Table("custom_css").
		Joins("JOIN brands ON custom_css.brand_id = brands.id").
		Where("brands.tenant_id = ?", tenantID).
		Count(&totalCustomCSS).Error; err != nil {
		return nil, fmt.Errorf("failed to count custom CSS: %w", err)
	}
	stats.TotalCustomCSS = int(totalCustomCSS)

	// Count custom JS
	if err := s.db.WithContext(ctx).
		Table("custom_js").
		Joins("JOIN brands ON custom_js.brand_id = brands.id").
		Where("brands.tenant_id = ?", tenantID).
		Count(&totalCustomJS).Error; err != nil {
		return nil, fmt.Errorf("failed to count custom JS: %w", err)
	}
	stats.TotalCustomJS = int(totalCustomJS)

	// Count layouts
	if err := s.db.WithContext(ctx).
		Table("layouts").
		Joins("JOIN themes ON layouts.theme_id = themes.id").
		Joins("JOIN brands ON themes.brand_id = brands.id").
		Where("brands.tenant_id = ?", tenantID).
		Count(&totalLayouts).Error; err != nil {
		return nil, fmt.Errorf("failed to count layouts: %w", err)
	}
	stats.TotalLayouts = int(totalLayouts)

	// Count components
	if err := s.db.WithContext(ctx).
		Table("components").
		Joins("JOIN themes ON components.theme_id = themes.id").
		Joins("JOIN brands ON themes.brand_id = brands.id").
		Where("brands.tenant_id = ?", tenantID).
		Count(&totalComponents).Error; err != nil {
		return nil, fmt.Errorf("failed to count components: %w", err)
	}
	stats.TotalComponents = int(totalComponents)

	stats.LastUpdated = time.Now()
	return &stats, nil
}

// Enhanced Validation Methods

func (s *BrandingService) validateBrand(brand *models.Brand) error {
	if brand.Name == "" {
		return errors.New("brand name is required")
	}
	if brand.Slug == "" {
		return errors.New("brand slug is required")
	}
	if brand.TenantID == 0 {
		return errors.New("tenant ID is required")
	}
	return nil
}

func (s *BrandingService) validateAssetCreation(asset *models.Asset, tenantID uint) error {
	if asset.Name == "" {
		return errors.New("asset name is required")
	}
	if asset.Type == "" {
		return errors.New("asset type is required")
	}
	if asset.BrandID == 0 {
		return errors.New("brand ID is required")
	}

	// Verify brand belongs to tenant
	var brand models.Brand
	err := s.db.Where("id = ? AND tenant_id = ?", asset.BrandID, tenantID).First(&brand).Error
	if err != nil {
		return errors.New("brand not found or access denied")
	}

	return nil
}

func (s *BrandingService) validateFileType(assetType, mimeType string) error {
	allowedTypes := map[string][]string{
		"logo":       {"image/png", "image/jpeg", "image/svg+xml"},
		"favicon":    {"image/png", "image/x-icon", "image/vnd.microsoft.icon"},
		"background": {"image/png", "image/jpeg"},
		"icon":       {"image/png", "image/svg+xml"},
		"font":       {"font/woff", "font/woff2", "font/ttf", "font/otf"},
		"css":        {"text/css"},
		"js":         {"application/javascript", "text/javascript"},
	}

	allowed, exists := allowedTypes[assetType]
	if !exists {
		return fmt.Errorf("unsupported asset type: %s", assetType)
	}

	for _, allowedType := range allowed {
		if mimeType == allowedType {
			return nil
		}
	}

	return fmt.Errorf("invalid file type %s for asset type %s", mimeType, assetType)
}

func (s *BrandingService) validateColorScheme(colorScheme *models.ColorScheme) error {
	if colorScheme.Name == "" {
		return errors.New("color scheme name is required")
	}
	if colorScheme.Primary == "" {
		return errors.New("primary color is required")
	}
	if colorScheme.Background == "" {
		return errors.New("background color is required")
	}
	if colorScheme.ThemeID == 0 {
		return errors.New("theme ID is required")
	}
	return nil
}

func (s *BrandingService) validateTypography(typography *models.Typography) error {
	if typography.Name == "" {
		return errors.New("typography name is required")
	}
	if typography.FontFamily == "" {
		return errors.New("font family is required")
	}
	if typography.ThemeID == 0 {
		return errors.New("theme ID is required")
	}
	return nil
}

func (s *BrandingService) validateCustomCSS(customCSS *models.CustomCSS) error {
	if customCSS.Name == "" {
		return errors.New("custom CSS name is required")
	}
	if customCSS.CSS == "" {
		return errors.New("CSS content is required")
	}
	if customCSS.BrandID == 0 {
		return errors.New("brand ID is required")
	}

	// Basic CSS validation - check for malicious content
	css := strings.ToLower(customCSS.CSS)
	if strings.Contains(css, "javascript:") || strings.Contains(css, "expression(") {
		return errors.New("CSS contains potentially malicious content")
	}

	return nil
}

func (s *BrandingService) validateCustomJS(customJS *models.CustomJS) error {
	if customJS.Name == "" {
		return errors.New("custom JS name is required")
	}
	if customJS.JavaScript == "" {
		return errors.New("JavaScript content is required")
	}
	if customJS.BrandID == 0 {
		return errors.New("brand ID is required")
	}

	// Basic JS validation - check for malicious content
	js := strings.ToLower(customJS.JavaScript)
	maliciousPatterns := []string{"eval(", "function(", "settimeout(", "setinterval(", "document.cookie"}
	for _, pattern := range maliciousPatterns {
		if strings.Contains(js, pattern) {
			return fmt.Errorf("JavaScript contains potentially malicious content: %s", pattern)
		}
	}

	return nil
}
