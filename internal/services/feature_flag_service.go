package services

import (
	"encoding/json"
	"fmt"

	"github.com/enterprise-status/statuspage/internal/models"
	"gorm.io/gorm"
)

type FeatureFlagService struct {
	db *gorm.DB
}

func NewFeatureFlagService(db *gorm.DB) *FeatureFlagService {
	return &FeatureFlagService{
		db: db,
	}
}

// IsFeatureEnabled checks if a feature is enabled for a tenant
// This checks both SaaS-level availability AND tenant-level enablement
func (s *FeatureFlagService) IsFeatureEnabled(tenantID uint, feature string) bool {
	// First check if the feature is available at SaaS level
	var availability models.SaaSFeatureAvailability
	err := s.db.Where("feature = ?", feature).First(&availability).Error
	if err != nil || !availability.IsAvailable {
		// Feature is not available at SaaS level
		return false
	}

	// Then check if the tenant has enabled it
	var flag models.FeatureFlag
	err = s.db.Where("tenant_id = ? AND feature = ?", tenantID, feature).First(&flag).Error
	if err != nil {
		// If feature flag doesn't exist, return false (disabled by default)
		return false
	}
	return flag.IsEnabled
}

// IsFeatureAvailable checks if a feature is available at SaaS level
func (s *FeatureFlagService) IsFeatureAvailable(feature string) bool {
	var availability models.SaaSFeatureAvailability
	err := s.db.Where("feature = ?", feature).First(&availability).Error
	if err != nil {
		return false
	}
	return availability.IsAvailable
}

// SetFeatureEnabled enables or disables a feature for a tenant
func (s *FeatureFlagService) SetFeatureEnabled(tenantID uint, feature string, enabled bool) error {
	var flag models.FeatureFlag
	err := s.db.Where("tenant_id = ? AND feature = ?", tenantID, feature).First(&flag).Error

	if err == gorm.ErrRecordNotFound {
		// Create new feature flag
		flag = models.FeatureFlag{
			TenantID:  tenantID,
			Feature:   feature,
			IsEnabled: enabled,
		}
		return s.db.Create(&flag).Error
	} else if err != nil {
		return err
	}

	// Update existing feature flag
	flag.IsEnabled = enabled
	return s.db.Save(&flag).Error
}

// SetFeatureConfig sets configuration for a feature
func (s *FeatureFlagService) SetFeatureConfig(tenantID uint, feature string, config interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var flag models.FeatureFlag
	err = s.db.Where("tenant_id = ? AND feature = ?", tenantID, feature).First(&flag).Error

	if err == gorm.ErrRecordNotFound {
		// Create new feature flag
		flag = models.FeatureFlag{
			TenantID:  tenantID,
			Feature:   feature,
			IsEnabled: false,
			Config:    string(configJSON),
		}
		return s.db.Create(&flag).Error
	} else if err != nil {
		return err
	}

	// Update existing feature flag
	flag.Config = string(configJSON)
	return s.db.Save(&flag).Error
}

// GetFeatureConfig gets configuration for a feature
func (s *FeatureFlagService) GetFeatureConfig(tenantID uint, feature string, config interface{}) error {
	var flag models.FeatureFlag
	err := s.db.Where("tenant_id = ? AND feature = ?", tenantID, feature).First(&flag).Error
	if err != nil {
		return err
	}

	if flag.Config == "" {
		return nil // No config set
	}

	return json.Unmarshal([]byte(flag.Config), config)
}

// GetAllFeatureFlags returns all feature flags for a tenant
func (s *FeatureFlagService) GetAllFeatureFlags(tenantID uint) ([]models.FeatureFlag, error) {
	var flags []models.FeatureFlag
	err := s.db.Where("tenant_id = ?", tenantID).Find(&flags).Error
	return flags, err
}

// GetFeatureFlagsForAllTenants returns all feature flags for all tenants (for SaaS admin)
func (s *FeatureFlagService) GetFeatureFlagsForAllTenants() ([]models.FeatureFlag, error) {
	var flags []models.FeatureFlag
	err := s.db.Preload("Tenant").Find(&flags).Error
	return flags, err
}

// SetGlobalFeatureEnabled enables or disables a feature globally for all tenants
func (s *FeatureFlagService) SetGlobalFeatureEnabled(feature string, enabled bool) error {
	// Get all tenants
	var tenants []models.Tenant
	err := s.db.Find(&tenants).Error
	if err != nil {
		return err
	}

	// Set feature flag for each tenant
	for _, tenant := range tenants {
		err := s.SetFeatureEnabled(tenant.ID, feature, enabled)
		if err != nil {
			return fmt.Errorf("failed to set feature for tenant %d: %w", tenant.ID, err)
		}
	}

	return nil
}

// SaaS-level feature availability methods

// SetFeatureAvailability sets whether a feature is available at SaaS level
func (s *FeatureFlagService) SetFeatureAvailability(feature string, available bool, description string) error {
	var availability models.SaaSFeatureAvailability
	err := s.db.Where("feature = ?", feature).First(&availability).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new availability record
			availability = models.SaaSFeatureAvailability{
				Feature:     feature,
				IsAvailable: available,
				Description: description,
			}
			result := s.db.Create(&availability)
			if result.Error != nil {
				return fmt.Errorf("failed to create feature availability: %w", result.Error)
			}
		} else {
			return fmt.Errorf("failed to get feature availability: %w", err)
		}
	} else {
		// Update existing record
		availability.IsAvailable = available
		if description != "" {
			availability.Description = description
		}
		result := s.db.Save(&availability)
		if result.Error != nil {
			return fmt.Errorf("failed to update feature availability: %w", result.Error)
		}
	}
	return nil
}

// GetAllFeatureAvailability returns all feature availability settings
func (s *FeatureFlagService) GetAllFeatureAvailability() ([]models.SaaSFeatureAvailability, error) {
	var availabilities []models.SaaSFeatureAvailability
	result := s.db.Find(&availabilities)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get feature availability: %w", result.Error)
	}
	return availabilities, nil
}

// GetAvailableFeaturesForTenant returns features that are both available at SaaS level and enabled for tenant
func (s *FeatureFlagService) GetAvailableFeaturesForTenant(tenantID uint) ([]models.FeatureFlag, error) {
	var flags []models.FeatureFlag

	// Get all features that are available at SaaS level
	var availabilities []models.SaaSFeatureAvailability
	err := s.db.Where("is_available = ?", true).Find(&availabilities).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get available features: %w", err)
	}

	// For each available feature, check if tenant has it enabled
	for _, availability := range availabilities {
		var flag models.FeatureFlag
		err := s.db.Where("tenant_id = ? AND feature = ?", tenantID, availability.Feature).First(&flag).Error
		if err == nil {
			// Tenant has this feature configured
			flags = append(flags, flag)
		} else if err == gorm.ErrRecordNotFound {
			// Tenant doesn't have this feature configured, create default disabled entry
			flags = append(flags, models.FeatureFlag{
				TenantID:  tenantID,
				Feature:   availability.Feature,
				IsEnabled: false,
			})
		}
	}

	return flags, nil
}
