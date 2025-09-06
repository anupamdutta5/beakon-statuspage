package services

import (
	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BrandingService struct {
	db *gorm.DB
}

func NewBrandingService() *BrandingService {
	return &BrandingService{
		db: database.GetDB(),
	}
}

// GetBranding retrieves the current branding configuration
func (s *BrandingService) GetBranding() (*models.Branding, error) {
	var branding models.Branding
	
	// Try to get existing branding, if not found, create default
	if err := s.db.First(&branding).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create default branding
			defaultBranding := &models.Branding{
				CompanyName:    "Your Company",
				PrimaryColor:   "#0052cc",
				SecondaryColor: "#f4f5f7",
				FooterText:     "© 2025 Your Company. All rights reserved.",
			}
			
			if err := s.db.Create(defaultBranding).Error; err != nil {
				logger.Error("Failed to create default branding", zap.Error(err))
				return nil, err
			}
			
			return defaultBranding, nil
		}
		return nil, err
	}
	
	return &branding, nil
}

// UpdateBranding updates the branding configuration
func (s *BrandingService) UpdateBranding(branding *models.Branding) error {
	// Ensure we have an ID to update
	if branding.ID == 0 {
		// Get existing branding to preserve ID
		existing, err := s.GetBranding()
		if err != nil {
			return err
		}
		branding.ID = existing.ID
	}
	
	if err := s.db.Save(branding).Error; err != nil {
		logger.Error("Failed to update branding", zap.Error(err))
		return err
	}
	
	logger.Info("Branding updated successfully")
	return nil
}

// ResetBranding resets branding to default values
func (s *BrandingService) ResetBranding() error {
	defaultBranding := &models.Branding{
		CompanyName:    "Your Company",
		PrimaryColor:   "#0052cc",
		SecondaryColor: "#f4f5f7",
		FooterText:     "© 2025 Your Company. All rights reserved.",
	}
	
	// Delete existing branding
	if err := s.db.Where("1 = 1").Delete(&models.Branding{}).Error; err != nil {
		logger.Error("Failed to delete existing branding", zap.Error(err))
		return err
	}
	
	// Create new default branding
	if err := s.db.Create(defaultBranding).Error; err != nil {
		logger.Error("Failed to create default branding", zap.Error(err))
		return err
	}
	
	logger.Info("Branding reset to default values")
	return nil
}

// ValidateBranding validates branding configuration
func (s *BrandingService) ValidateBranding(branding *models.Branding) error {
	if branding.CompanyName == "" {
		return gorm.ErrInvalidData
	}
	
	// Validate color format (basic hex color validation)
	if branding.PrimaryColor != "" && !isValidHexColor(branding.PrimaryColor) {
		return gorm.ErrInvalidData
	}
	
	if branding.SecondaryColor != "" && !isValidHexColor(branding.SecondaryColor) {
		return gorm.ErrInvalidData
	}
	
	return nil
}

// isValidHexColor checks if a string is a valid hex color
func isValidHexColor(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}
	
	for i := 1; i < 7; i++ {
		c := color[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	
	return true
}
