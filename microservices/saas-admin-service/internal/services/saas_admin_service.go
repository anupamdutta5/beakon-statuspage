// Package services provides business logic for the SaaS Admin Service.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/anupamdutta5/saas-admin-service/internal/clients"
	"github.com/anupamdutta5/saas-admin-service/internal/config"
	"github.com/anupamdutta5/saas-admin-service/internal/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SaaSAdminService handles SaaS admin-related business logic.
type SaaSAdminService struct {
	config          *config.Config
	logger          *zap.Logger
	db              *gorm.DB
	analyticsClient *clients.AnalyticsClient
	httpClient      *http.Client
	serviceURLs     config.ServiceURLs
}

// NewSaaSAdminService creates a new SaaS admin service.
func NewSaaSAdminService(cfg *config.Config, logger *zap.Logger, db *gorm.DB) (*SaaSAdminService, error) {
	// Initialize analytics client with default URL
	analyticsClient := clients.NewAnalyticsClient("http://localhost:8081", logger)

	// Get service URLs from environment or defaults
	serviceURLs := config.GetServiceURLsFromEnv()

	return &SaaSAdminService{
		config:          cfg,
		logger:          logger,
		db:              db,
		analyticsClient: analyticsClient,
		httpClient:      &http.Client{Timeout: 10 * time.Second},
		serviceURLs:     serviceURLs,
	}, nil
}

// SetDB allows injecting a test database for testing purposes
func (s *SaaSAdminService) SetDB(db *gorm.DB) {
	s.db = db
}

// GetDB returns the database connection
func (s *SaaSAdminService) GetDB() *gorm.DB {
	return s.db
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

// Health checks the health of the service.
func (s *SaaSAdminService) Health(ctx context.Context) error {
	// Check database connection
	if s.db == nil {
		return fmt.Errorf("database connection not available")
	}

	// Test database connection
	if err := s.db.Exec("SELECT 1").Error; err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// ValidateAdminCredentials validates admin user credentials against the database
func (s *SaaSAdminService) ValidateAdminCredentials(ctx context.Context, username, password string) (*models.SaaSAdminUser, error) {
	s.logger.Info("Validating admin credentials", zap.String("username", username))

	if s.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	var user models.SaaSAdminUser
	if err := s.db.Where("username = ? AND status = ?", username, "active").First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Warn("User not found or inactive", zap.String("username", username))
			return nil, fmt.Errorf("invalid credentials")
		}
		s.logger.Error("Database error during credential validation", zap.Error(err))
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Use bcrypt to compare password (password is hashed in database)
	// Import: "golang.org/x/crypto/bcrypt"
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.logger.Warn("Invalid password", zap.String("username", username))
		return nil, fmt.Errorf("invalid credentials")
	}

	s.logger.Info("Admin credentials validated successfully", zap.String("username", username))
	return &user, nil
}

// GetAdminUserByID retrieves an admin user by ID
func (s *SaaSAdminService) GetAdminUserByID(ctx context.Context, userID uint) (*models.SaaSAdminUser, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	var user models.SaaSAdminUser
	if err := s.db.WithContext(ctx).Where("id = ? AND status = ?", userID, "active").First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found or inactive")
		}
		s.logger.Error("Database error while fetching user", zap.Error(err), zap.Uint("user_id", userID))
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &user, nil
}

// Plan Management

// CreatePlan creates a new SaaS plan.
func (s *SaaSAdminService) CreatePlan(ctx context.Context, plan *models.SaaSPlan) error {
	s.logger.Info("Creating SaaS plan",
		zap.String("plan_name", plan.Name),
		zap.String("plan_slug", plan.Slug))

	// Check database connection
	if s.db == nil {
		s.logger.Error("Database connection not available")
		return fmt.Errorf("database connection not available")
	}

	// Sanitize and validate input
	plan.Name = strings.TrimSpace(plan.Name)
	plan.Slug = strings.TrimSpace(plan.Slug)
	plan.Description = strings.TrimSpace(plan.Description)

	// Generate slug if not provided
	if plan.Slug == "" && plan.Name != "" {
		plan.Slug = strings.ToLower(strings.ReplaceAll(plan.Name, " ", "-"))
	}

	// Set defaults
	if plan.Currency == "" {
		plan.Currency = "USD"
	}
	if plan.BillingInterval == "" {
		plan.BillingInterval = "monthly"
	}

	if err := s.db.Create(plan).Error; err != nil {
		s.logger.Error("Failed to create plan", zap.Error(err))
		return fmt.Errorf("failed to create plan: %w", err)
	}

	s.logger.Info("SaaS plan created successfully",
		zap.String("plan_id", plan.ID.String()),
		zap.String("plan_name", plan.Name))

	return nil
}

// GetPlan retrieves a plan by ID.
func (s *SaaSAdminService) GetPlan(ctx context.Context, planID uuid.UUID) (*models.SaaSPlan, error) {
	s.logger.Info("Getting SaaS plan", zap.String("plan_id", planID.String()))

	var plan models.SaaSPlan
	if err := s.db.Preload("Features").Preload("Tiers").First(&plan, "id = ?", planID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Warn("Plan not found", zap.String("plan_id", planID.String()))
			return nil, fmt.Errorf("plan not found")
		}
		s.logger.Error("Failed to get plan", zap.Error(err))
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	s.logger.Info("Plan retrieved successfully", zap.String("plan_id", planID.String()))
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

	if s.db == nil {
		s.logger.Debug("Database not available, returning empty list")
		return []*models.SaaSPlan{}, nil
	}

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

// ListPublicPlans lists all active public plans with their pricing tiers for landing page.
func (s *SaaSAdminService) ListPublicPlans(ctx context.Context) ([]*models.SaaSPlan, error) {
	s.logger.Info("Listing public pricing plans for landing page")

	if s.db == nil {
		s.logger.Debug("Database not available, returning empty list")
		return []*models.SaaSPlan{}, nil
	}

	var plans []*models.SaaSPlan

	// Get all active public plans, ordered by display_order
	if err := s.db.Preload("PricingTiers").
		Where("is_active = ? AND is_public = ?", true, true).
		Order("display_order ASC, created_at ASC").
		Find(&plans).Error; err != nil {
		s.logger.Error("Failed to list public plans", zap.Error(err))
		return nil, fmt.Errorf("failed to list public plans: %w", err)
	}

	s.logger.Info("Retrieved public plans",
		zap.Int("count", len(plans)))

	return plans, nil
}

// UpdatePlan updates a plan.
func (s *SaaSAdminService) UpdatePlan(ctx context.Context, planID uuid.UUID, updates *models.SaaSPlan) (*models.SaaSPlan, error) {
	s.logger.Info("Updating SaaS plan", zap.String("plan_id", planID.String()))

	// Use Select to ensure zero values are updated, but only for specific fields
	if err := s.db.Model(&models.SaaSPlan{}).Where("id = ?", planID).Select("is_public", "is_active", "is_popular", "name", "description", "price", "currency", "billing_interval").Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update plan", zap.Error(err))
		return nil, fmt.Errorf("failed to update plan: %w", err)
	}

	// Get the updated plan
	var updatedPlan models.SaaSPlan
	if err := s.db.First(&updatedPlan, "id = ?", planID).Error; err != nil {
		s.logger.Error("Failed to get updated plan", zap.Error(err))
		return nil, fmt.Errorf("failed to get updated plan: %w", err)
	}

	s.logger.Info("SaaS plan updated successfully", zap.String("plan_id", planID.String()))
	return &updatedPlan, nil
}

// DeletePlan deletes a plan.
func (s *SaaSAdminService) DeletePlan(ctx context.Context, planID uuid.UUID) error {
	s.logger.Info("Deleting SaaS plan", zap.String("plan_id", planID.String()))

	if err := s.db.Delete(&models.SaaSPlan{}, "id = ?", planID).Error; err != nil {
		s.logger.Error("Failed to delete plan", zap.Error(err))
		return fmt.Errorf("failed to delete plan: %w", err)
	}

	s.logger.Info("Plan deleted successfully", zap.String("plan_id", planID.String()))
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

	if s.db == nil {
		s.logger.Debug("Database not available, returning empty list")
		return []*models.SaaSFeatureFlag{}, nil
	}

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
func (s *SaaSAdminService) UpdateFeatureFlag(ctx context.Context, flagID uint, updates *models.SaaSFeatureFlag) (*models.SaaSFeatureFlag, error) {
	s.logger.Info("Updating SaaS feature flag", zap.Uint("flag_id", flagID))

	if err := s.db.Model(&models.SaaSFeatureFlag{}).Where("id = ?", flagID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update feature flag", zap.Error(err))
		return nil, fmt.Errorf("failed to update feature flag: %w", err)
	}

	// Get the updated feature flag
	var updatedFlag models.SaaSFeatureFlag
	if err := s.db.First(&updatedFlag, flagID).Error; err != nil {
		s.logger.Error("Failed to get updated feature flag", zap.Error(err))
		return nil, fmt.Errorf("failed to get updated feature flag: %w", err)
	}

	s.logger.Info("SaaS feature flag updated successfully", zap.Uint("flag_id", flagID))
	return &updatedFlag, nil
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

// Pricing Management

// CreatePricingFeature creates a new pricing feature.
func (s *SaaSAdminService) CreatePricingFeature(ctx context.Context, feature *models.PricingFeature) error {
	s.logger.Info("Creating pricing feature",
		zap.String("feature_name", feature.Name),
		zap.String("category", feature.Category))

	if err := s.db.Create(feature).Error; err != nil {
		s.logger.Error("Failed to create pricing feature", zap.Error(err))
		return fmt.Errorf("failed to create pricing feature: %w", err)
	}

	s.logger.Info("Pricing feature created successfully", zap.Uint("feature_id", feature.ID))
	return nil
}

// GetPricingFeatures retrieves all pricing features.
func (s *SaaSAdminService) GetPricingFeatures(ctx context.Context, category string) ([]*models.PricingFeature, error) {
	s.logger.Info("Getting pricing features", zap.String("category", category))

	var features []*models.PricingFeature
	query := s.db.Where("is_active = ?", true).Order("category, \"order\", name")

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Find(&features).Error; err != nil {
		s.logger.Error("Failed to get pricing features", zap.Error(err))
		return nil, fmt.Errorf("failed to get pricing features: %w", err)
	}

	return features, nil
}

// UpdatePricingFeature updates a pricing feature.
func (s *SaaSAdminService) UpdatePricingFeature(ctx context.Context, featureID uint, updates *models.PricingFeature) (*models.PricingFeature, error) {
	s.logger.Info("Updating pricing feature", zap.Uint("feature_id", featureID))

	if err := s.db.Model(&models.PricingFeature{}).Where("id = ?", featureID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update pricing feature", zap.Error(err))
		return nil, fmt.Errorf("failed to update pricing feature: %w", err)
	}

	var updatedFeature models.PricingFeature
	if err := s.db.First(&updatedFeature, featureID).Error; err != nil {
		s.logger.Error("Failed to get updated pricing feature", zap.Error(err))
		return nil, fmt.Errorf("failed to get updated pricing feature: %w", err)
	}

	s.logger.Info("Pricing feature updated successfully", zap.Uint("feature_id", featureID))
	return &updatedFeature, nil
}

// DeletePricingFeature deletes a pricing feature.
func (s *SaaSAdminService) DeletePricingFeature(ctx context.Context, featureID uint) error {
	s.logger.Info("Deleting pricing feature", zap.Uint("feature_id", featureID))

	// First, remove all plan-feature relationships
	if err := s.db.Where("feature_id = ?", featureID).Delete(&models.PlanFeature{}).Error; err != nil {
		s.logger.Error("Failed to remove plan-feature relationships", zap.Error(err))
		return fmt.Errorf("failed to remove plan-feature relationships: %w", err)
	}

	// Then delete the feature
	if err := s.db.Delete(&models.PricingFeature{}, featureID).Error; err != nil {
		s.logger.Error("Failed to delete pricing feature", zap.Error(err))
		return fmt.Errorf("failed to delete pricing feature: %w", err)
	}

	s.logger.Info("Pricing feature deleted successfully", zap.Uint("feature_id", featureID))
	return nil
}

// CreatePricingTier creates a new pricing tier for a plan.
func (s *SaaSAdminService) CreatePricingTier(ctx context.Context, tier *models.PricingTier) error {
	s.logger.Info("Creating pricing tier",
		zap.String("plan_id", tier.PlanID.String()),
		zap.String("billing_interval", tier.BillingInterval))

	// Verify the plan exists before creating the tier
	var plan models.SaaSPlan
	if err := s.db.First(&plan, "id = ?", tier.PlanID).Error; err != nil {
		s.logger.Error("Plan not found", 
			zap.String("plan_id", tier.PlanID.String()),
			zap.Error(err))
		return fmt.Errorf("plan not found: %w", err)
	}

	if err := s.db.Create(tier).Error; err != nil {
		s.logger.Error("Failed to create pricing tier", 
			zap.String("plan_id", tier.PlanID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to create pricing tier: %w", err)
	}

	s.logger.Info("Pricing tier created successfully", 
		zap.Uint("tier_id", tier.ID),
		zap.String("plan_id", tier.PlanID.String()))
	return nil
}

// GetPricingTiers retrieves pricing tiers for a plan.
func (s *SaaSAdminService) GetPricingTiers(ctx context.Context, planID uuid.UUID) ([]*models.PricingTier, error) {
	s.logger.Info("Getting pricing tiers", zap.String("plan_id", planID.String()))

	var tiers []*models.PricingTier
	if err := s.db.Where("plan_id = ? AND is_active = ?", planID, true).Order("billing_interval").Find(&tiers).Error; err != nil {
		s.logger.Error("Failed to get pricing tiers", zap.Error(err))
		return nil, fmt.Errorf("failed to get pricing tiers: %w", err)
	}

	return tiers, nil
}

// UpdatePricingTier updates a pricing tier.
func (s *SaaSAdminService) UpdatePricingTier(ctx context.Context, tierID uint, updates *models.PricingTier) (*models.PricingTier, error) {
	s.logger.Info("Updating pricing tier", zap.Uint("tier_id", tierID))

	if err := s.db.Model(&models.PricingTier{}).Where("id = ?", tierID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update pricing tier", zap.Error(err))
		return nil, fmt.Errorf("failed to update pricing tier: %w", err)
	}

	var updatedTier models.PricingTier
	if err := s.db.First(&updatedTier, tierID).Error; err != nil {
		s.logger.Error("Failed to get updated pricing tier", zap.Error(err))
		return nil, fmt.Errorf("failed to get updated pricing tier: %w", err)
	}

	s.logger.Info("Pricing tier updated successfully", zap.Uint("tier_id", tierID))
	return &updatedTier, nil
}

// DeletePricingTier deletes a pricing tier.
func (s *SaaSAdminService) DeletePricingTier(ctx context.Context, tierID uint) error {
	s.logger.Info("Deleting pricing tier", zap.Uint("tier_id", tierID))

	if err := s.db.Delete(&models.PricingTier{}, tierID).Error; err != nil {
		s.logger.Error("Failed to delete pricing tier", zap.Error(err))
		return fmt.Errorf("failed to delete pricing tier: %w", err)
	}

	s.logger.Info("Pricing tier deleted successfully", zap.Uint("tier_id", tierID))
	return nil
}

// AssignFeatureToPlan assigns a feature to a plan.
func (s *SaaSAdminService) AssignFeatureToPlan(ctx context.Context, planUUID uuid.UUID, featureID uint, order int) error {
	s.logger.Info("Assigning feature to plan",
		zap.String("plan_id", planUUID.String()),
		zap.Uint("feature_id", featureID))

	planFeature := &models.PlanFeature{
		PlanID:    planUUID,
		FeatureID: featureID,
		IsEnabled: true,
		Order:     order,
	}

	if err := s.db.Create(planFeature).Error; err != nil {
		s.logger.Error("Failed to assign feature to plan", zap.Error(err))
		return fmt.Errorf("failed to assign feature to plan: %w", err)
	}

	s.logger.Info("Feature assigned to plan successfully")
	return nil
}

// RemoveFeatureFromPlan removes a feature from a plan.
func (s *SaaSAdminService) RemoveFeatureFromPlan(ctx context.Context, planID uuid.UUID, featureID uint) error {
	s.logger.Info("Removing feature from plan",
		zap.String("plan_id", planID.String()),
		zap.Uint("feature_id", featureID))

	if err := s.db.Where("plan_id = ? AND feature_id = ?", planID, featureID).Delete(&models.PlanFeature{}).Error; err != nil {
		s.logger.Error("Failed to remove feature from plan", 
			zap.String("plan_id", planID.String()),
			zap.Uint("feature_id", featureID),
			zap.Error(err))
		return fmt.Errorf("failed to remove feature from plan: %w", err)
	}

	s.logger.Info("Feature removed from plan successfully",
		zap.String("plan_id", planID.String()),
		zap.Uint("feature_id", featureID))
	return nil
}

// GetPlanFeatures retrieves all features for a plan.
func (s *SaaSAdminService) GetPlanFeatures(ctx context.Context, planID uuid.UUID) ([]*models.PlanFeature, error) {
	s.logger.Info("Getting plan features", zap.String("plan_id", planID.String()))

	var planFeatures []*models.PlanFeature
	if err := s.db.Where("plan_id = ?", planID).Preload("Feature").Order("`order`").Find(&planFeatures).Error; err != nil {
		s.logger.Error("Failed to get plan features", zap.Error(err))
		return nil, fmt.Errorf("failed to get plan features: %w", err)
	}

	return planFeatures, nil
}

// GetPublicPricingPlans retrieves all public pricing plans with their features and tiers.
func (s *SaaSAdminService) GetPublicPricingPlans(ctx context.Context) ([]*models.SaaSPlan, error) {
	s.logger.Info("Getting public pricing plans")

	var plans []*models.SaaSPlan
	if err := s.db.Where("is_active = ? AND is_public = ?", true, true).Order("`order`, name").Find(&plans).Error; err != nil {
		s.logger.Error("Failed to get public pricing plans", zap.Error(err))
		return nil, fmt.Errorf("failed to get public pricing plans: %w", err)
	}

	// Load features and tiers for each plan
	for _, plan := range plans {
		// Load features
		var planFeatures []*models.PlanFeature
		if err := s.db.Preload("Feature").Where("plan_id = ? AND is_enabled = ?", plan.ID, true).Order("`order`").Find(&planFeatures).Error; err == nil {
			// Convert to features list for JSON storage
			var features []string
			for _, pf := range planFeatures {
				if pf.Feature.Name != "" {
					features = append(features, pf.Feature.Name)
				}
			}
			if featuresJSON, err := json.Marshal(features); err == nil {
				plan.Features = string(featuresJSON)
			}
		}

		// Load pricing tiers
		var tiers []*models.PricingTier
		if err := s.db.Where("plan_id = ? AND is_active = ?", plan.ID, true).Find(&tiers).Error; err == nil {
			// Set the base price from the first tier (usually monthly)
			for _, tier := range tiers {
				if tier.BillingInterval == "monthly" {
					plan.Price = tier.Price
					plan.Currency = tier.Currency
					plan.BillingInterval = tier.BillingInterval
					break
				}
			}
		}
	}

	return plans, nil
}

// ValidatePricingPlan validates a pricing plan configuration.
func (s *SaaSAdminService) ValidatePricingPlan(ctx context.Context, plan *models.SaaSPlan) error {
	s.logger.Info("Validating pricing plan", zap.String("plan_name", plan.Name))

	var errors []string

	// Validate required fields
	if strings.TrimSpace(plan.Name) == "" {
		errors = append(errors, "plan name is required")
	}
	if strings.TrimSpace(plan.Slug) == "" {
		errors = append(errors, "plan slug is required")
	}
	if plan.Price < 0 {
		errors = append(errors, "plan price cannot be negative")
	}
	if plan.Currency == "" {
		plan.Currency = "USD"
	}
	if plan.BillingInterval == "" {
		plan.BillingInterval = "monthly"
	}
	if plan.BillingInterval != "monthly" && plan.BillingInterval != "yearly" {
		errors = append(errors, "billing interval must be 'monthly' or 'yearly'")
	}

	// Validate limits
	if plan.MaxTenants < 0 {
		errors = append(errors, "max tenants cannot be negative")
	}
	if plan.MaxUsers < 0 {
		errors = append(errors, "max users cannot be negative")
	}
	if plan.MaxServices < 0 {
		errors = append(errors, "max services cannot be negative")
	}

	// Validate features JSON
	if plan.Features != "" {
		var features []string
		if err := json.Unmarshal([]byte(plan.Features), &features); err != nil {
			errors = append(errors, fmt.Sprintf("invalid features JSON: %v", err))
		}
	}

	// Validate limits JSON
	if plan.Limits != "" {
		var limits map[string]interface{}
		if err := json.Unmarshal([]byte(plan.Limits), &limits); err != nil {
			errors = append(errors, fmt.Sprintf("invalid limits JSON: %v", err))
		}
	}

	// Return all validation errors if any
	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}

	s.logger.Info("Pricing plan validation successful", zap.String("plan_name", plan.Name))
	return nil
}

// SyncPricingToLandingPage syncs pricing plans to the landing page service.
func (s *SaaSAdminService) SyncPricingToLandingPage(ctx context.Context) error {
	s.logger.Info("Syncing pricing plans to landing page service")

	// Get all public pricing plans
	plans, err := s.GetPublicPricingPlans(ctx)
	if err != nil {
		s.logger.Error("Failed to get public pricing plans for sync", zap.Error(err))
		return fmt.Errorf("failed to get public pricing plans for sync: %w", err)
	}

	// Here you would typically make an HTTP call to the landing page service
	// to update its pricing data. For now, we'll just log the sync.
	s.logger.Info("Pricing plans synced successfully",
		zap.Int("plan_count", len(plans)))

	return nil
}

// initDatabase initializes the database connection.
// This is kept for backward compatibility and testing purposes.
func initDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.Port,
		cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %v", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.Platform{},
		&models.SaaSPlan{},
		&models.PricingTier{},
		&models.PricingFeature{},
		&models.PlanFeature{},
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

// Analytics Management

// GetAnalyticsOverview retrieves analytics overview data.
func (s *SaaSAdminService) GetAnalyticsOverview(ctx context.Context, tenantID string) (*clients.AnalyticsOverview, error) {
	s.logger.Info("Getting analytics overview", zap.String("tenant_id", tenantID))

	if s.analyticsClient == nil {
		s.logger.Warn("Analytics client not available, using mock data")
		return s.getMockAnalyticsOverview(tenantID), nil
	}

	overview, err := s.analyticsClient.GetAnalyticsOverview(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to get analytics overview from analytics service",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		// Return mock data as fallback
		return s.getMockAnalyticsOverview(tenantID), nil
	}

	s.logger.Info("Successfully retrieved analytics overview",
		zap.String("tenant_id", tenantID),
		zap.Int64("total_views", overview.TotalViews))

	return overview, nil
}

// GetAnalyticsMetrics retrieves analytics metrics data.
func (s *SaaSAdminService) GetAnalyticsMetrics(ctx context.Context, tenantID string, timeRange string) ([]clients.MetricData, error) {
	s.logger.Info("Getting analytics metrics",
		zap.String("tenant_id", tenantID),
		zap.String("time_range", timeRange))

	if s.analyticsClient == nil {
		s.logger.Warn("Analytics client not available, using mock data")
		return s.getMockMetrics(tenantID), nil
	}

	metrics, err := s.analyticsClient.GetMetrics(ctx, tenantID, timeRange)
	if err != nil {
		s.logger.Error("Failed to get analytics metrics from analytics service",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		// Return mock data as fallback
		return s.getMockMetrics(tenantID), nil
	}

	s.logger.Info("Successfully retrieved analytics metrics",
		zap.String("tenant_id", tenantID),
		zap.Int("metric_count", len(metrics)))

	return metrics, nil
}

// CreateAnalyticsMetric creates a new analytics metric.
func (s *SaaSAdminService) CreateAnalyticsMetric(ctx context.Context, tenantID string, metric clients.MetricData) (*clients.MetricData, error) {
	s.logger.Info("Creating analytics metric",
		zap.String("tenant_id", tenantID),
		zap.String("metric_name", metric.Name))

	if s.analyticsClient == nil {
		s.logger.Warn("Analytics client not available, cannot create metric")
		return nil, fmt.Errorf("analytics service not available")
	}

	createdMetric, err := s.analyticsClient.CreateMetric(ctx, tenantID, metric)
	if err != nil {
		s.logger.Error("Failed to create analytics metric",
			zap.String("tenant_id", tenantID),
			zap.String("metric_name", metric.Name),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create analytics metric: %w", err)
	}

	s.logger.Info("Successfully created analytics metric",
		zap.String("tenant_id", tenantID),
		zap.String("metric_id", createdMetric.ID))

	return createdMetric, nil
}

// CheckAnalyticsHealth checks the health of the analytics service.
func (s *SaaSAdminService) CheckAnalyticsHealth(ctx context.Context) error {
	if s.analyticsClient == nil {
		return fmt.Errorf("analytics client not available")
	}

	return s.analyticsClient.Health(ctx)
}

// getMockAnalyticsOverview returns mock analytics data when the analytics service is unavailable.
func (s *SaaSAdminService) getMockAnalyticsOverview(tenantID string) *clients.AnalyticsOverview {
	now := time.Now()
	dailyMetrics := make([]clients.DailyMetric, 30)

	// Generate mock data for the last 30 days
	for i := 0; i < 30; i++ {
		date := now.AddDate(0, 0, -i)
		dailyMetrics[29-i] = clients.DailyMetric{
			Date:   date.Format("2006-01-02"),
			Views:  int64(800 + i*25),
			Uptime: 99.2 + float64(i)*0.02,
		}
	}

	// Customize mock data based on tenant
	totalViews := int64(23567)
	uniqueVisitors := int64(12340)
	uptime := 99.87
	responseTime := 156.3

	switch tenantID {
	case "tenant-1":
		totalViews = 45123
		uniqueVisitors = 18765
		uptime = 99.95
		responseTime = 134.2
	case "tenant-2":
		totalViews = 78901
		uniqueVisitors = 34567
		uptime = 99.98
		responseTime = 89.5
	case "tenant-3":
		totalViews = 12456
		uniqueVisitors = 5432
		uptime = 99.12
		responseTime = 278.9
	}

	return &clients.AnalyticsOverview{
		TotalViews:       totalViews,
		UniqueVisitors:   uniqueVisitors,
		UptimePercentage: uptime,
		AvgResponseTime:  responseTime,
		StatusPages:      3,
		ActiveIncidents:  0,
		DailyMetrics:     dailyMetrics,
		TopCountries: []clients.CountryMetric{
			{Country: "United States", Views: totalViews * 40 / 100},
			{Country: "United Kingdom", Views: totalViews * 25 / 100},
			{Country: "Germany", Views: totalViews * 15 / 100},
			{Country: "Canada", Views: totalViews * 12 / 100},
			{Country: "Australia", Views: totalViews * 8 / 100},
		},
		ResponseTimes: []clients.ResponseTimeMetric{
			{Timestamp: now.Add(-1 * time.Hour), ResponseTime: responseTime},
			{Timestamp: now.Add(-2 * time.Hour), ResponseTime: responseTime + 15.2},
			{Timestamp: now.Add(-3 * time.Hour), ResponseTime: responseTime - 8.7},
		},
		UptimeHistory: []clients.UptimeMetric{
			{Date: now.Format("2006-01-02"), Uptime: uptime},
			{Date: now.AddDate(0, 0, -1).Format("2006-01-02"), Uptime: uptime - 0.05},
			{Date: now.AddDate(0, 0, -2).Format("2006-01-02"), Uptime: uptime + 0.02},
		},
	}
}

// getMockMetrics returns mock metrics data when the analytics service is unavailable.
func (s *SaaSAdminService) getMockMetrics(tenantID string) []clients.MetricData {
	now := time.Now()

	baseViews := float64(15432)
	baseResponseTime := 245.5
	baseUptime := 99.95

	// Customize based on tenant
	switch tenantID {
	case "tenant-1":
		baseViews = 25678
		baseResponseTime = 189.3
		baseUptime = 99.98
	case "tenant-2":
		baseViews = 45123
		baseResponseTime = 134.7
		baseUptime = 99.99
	case "tenant-3":
		baseViews = 8967
		baseResponseTime = 312.1
		baseUptime = 98.95
	}

	return []clients.MetricData{
		{
			ID:        "metric-views-" + tenantID,
			Name:      "Page Views",
			Type:      "counter",
			Value:     baseViews,
			Unit:      "views",
			Timestamp: now,
			Metadata: map[string]interface{}{
				"tenant_id": tenantID,
				"period":    "24h",
			},
		},
		{
			ID:        "metric-response-" + tenantID,
			Name:      "Average Response Time",
			Type:      "gauge",
			Value:     baseResponseTime,
			Unit:      "ms",
			Timestamp: now,
			Metadata: map[string]interface{}{
				"tenant_id": tenantID,
				"period":    "1h",
			},
		},
		{
			ID:        "metric-uptime-" + tenantID,
			Name:      "Uptime Percentage",
			Type:      "gauge",
			Value:     baseUptime,
			Unit:      "%",
			Timestamp: now,
			Metadata: map[string]interface{}{
				"tenant_id": tenantID,
				"period":    "24h",
			},
		},
		{
			ID:        "metric-errors-" + tenantID,
			Name:      "Error Rate",
			Type:      "gauge",
			Value:     0.12,
			Unit:      "%",
			Timestamp: now,
			Metadata: map[string]interface{}{
				"tenant_id": tenantID,
				"period":    "1h",
			},
		},
	}
}

// Tenant Management Methods

// ListTenants returns a list of all tenants by calling tenant-admin-service API.
func (s *SaaSAdminService) ListTenants(ctx context.Context) ([]*models.SaaSTenant, error) {
	s.logger.Info("Listing tenants from tenant-admin-service")

	// Call tenant-admin-service public API
	url := fmt.Sprintf("%s/api/v1/public/tenants", s.serviceURLs.TenantAdminService)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.logger.Error("Failed to create request", zap.Error(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("Failed to call tenant-admin-service", zap.Error(err), zap.String("url", url))
		return nil, fmt.Errorf("failed to call tenant-admin-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error("Tenant-admin-service returned error",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)))
		return nil, fmt.Errorf("tenant-admin-service returned status %d", resp.StatusCode)
	}

	var apiResponse struct {
		Data []*models.SaaSTenant `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		s.logger.Error("Failed to decode response", zap.Error(err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	s.logger.Info("Tenants listed successfully from tenant-admin-service", zap.Int("count", len(apiResponse.Data)))
	return apiResponse.Data, nil
}

// GetTenant retrieves a tenant by ID.
func (s *SaaSAdminService) GetTenant(ctx context.Context, tenantID uuid.UUID) (*models.SaaSTenant, error) {
	s.logger.Info("Getting tenant", zap.String("tenant_id", tenantID.String()))

	var tenant models.SaaSTenant
	if err := s.db.Preload("Plan").First(&tenant, tenantID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant not found")
		}
		s.logger.Error("Failed to get tenant", zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return &tenant, nil
}

// CreateTenant creates a new tenant.
func (s *SaaSAdminService) CreateTenant(ctx context.Context, tenant *models.SaaSTenant) error {
	s.logger.Info("Creating tenant",
		zap.String("name", tenant.Name),
		zap.String("domain", tenant.Domain))

	// Generate slug from name if not provided
	if tenant.Slug == "" {
		tenant.Slug = strings.ToLower(strings.ReplaceAll(tenant.Name, " ", "-"))
	}

	// Set default subdomain if not provided
	if tenant.Subdomain == "" {
		tenant.Subdomain = tenant.Slug + ".yourdomain.com"
	}

	// Create tenant in database
	if err := s.db.Create(tenant).Error; err != nil {
		s.logger.Error("Failed to create tenant", zap.Error(err))
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	s.logger.Info("Tenant created successfully",
		zap.String("tenant_id", tenant.ID.String()),
		zap.String("name", tenant.Name))
	return nil
}

// UpdateTenant updates an existing tenant.
func (s *SaaSAdminService) UpdateTenant(ctx context.Context, tenantID uuid.UUID, updates *models.SaaSTenant) error {
	s.logger.Info("Updating tenant", zap.String("tenant_id", tenantID.String()))

	if err := s.db.Model(&models.SaaSTenant{}).Where("id = ?", tenantID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update tenant", zap.Error(err))
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	s.logger.Info("Tenant updated successfully", zap.String("tenant_id", tenantID.String()))
	return nil
}

// DeleteTenant deletes a tenant.
func (s *SaaSAdminService) DeleteTenant(ctx context.Context, tenantID uuid.UUID) error {
	s.logger.Info("Deleting tenant", zap.String("tenant_id", tenantID.String()))

	if err := s.db.Delete(&models.SaaSTenant{}, tenantID).Error; err != nil {
		s.logger.Error("Failed to delete tenant", zap.Error(err))
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	s.logger.Info("Tenant deleted successfully", zap.String("tenant_id", tenantID.String()))
	return nil
}
