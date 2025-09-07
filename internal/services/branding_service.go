package services

import (
	"fmt"
	"time"

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
		db: database.DB,
	}
}

// Custom Domain Management

type CustomDomain struct {
	ID         uint   `gorm:"primaryKey"`
	TenantID   uint   `gorm:"not null"`
	Domain     string `gorm:"uniqueIndex;not null"`
	SSLEnabled bool   `gorm:"default:true"`
	SSLStatus  string `gorm:"default:'pending'"` // pending, active, failed, expired
	Verified   bool   `gorm:"default:false"`
	VerifiedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (s *BrandingService) CreateCustomDomain(tenantID uint, domain string) (*CustomDomain, error) {
	// Check if domain already exists
	var existing CustomDomain
	if err := s.db.Where("domain = ?", domain).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("domain already exists")
	}

	customDomain := &CustomDomain{
		TenantID:   tenantID,
		Domain:     domain,
		SSLEnabled: true,
		SSLStatus:  "pending",
		Verified:   false,
	}

	if err := s.db.Create(customDomain).Error; err != nil {
		logger.Error("Failed to create custom domain", zap.Error(err))
		return nil, err
	}

	// Generate SSL certificate (this would integrate with Let's Encrypt or similar)
	go s.generateSSLCertificate(customDomain)

	logger.Info("Custom domain created successfully",
		zap.Uint("tenant_id", tenantID),
		zap.String("domain", domain))

	return customDomain, nil
}

func (s *BrandingService) VerifyDomainOwnership(domainID uint, verificationCode string) error {
	var domain CustomDomain
	if err := s.db.First(&domain, domainID).Error; err != nil {
		return err
	}

	// Verify domain ownership (this would check DNS records)
	verified, err := s.checkDomainVerification(domain.Domain, verificationCode)
	if err != nil {
		return err
	}

	if verified {
		domain.Verified = true
		now := time.Now()
		domain.VerifiedAt = &now
		domain.SSLStatus = "active"

		if err := s.db.Save(&domain).Error; err != nil {
			return err
		}

		logger.Info("Domain ownership verified", zap.String("domain", domain.Domain))
	}

	return nil
}

// Theme Management

type Theme struct {
	ID              uint   `gorm:"primaryKey"`
	TenantID        uint   `gorm:"not null"`
	Name            string `gorm:"not null"`
	IsDefault       bool   `gorm:"default:false"`
	PrimaryColor    string `gorm:"default:'#0052cc'"`
	SecondaryColor  string `gorm:"default:'#f4f5f7'"`
	AccentColor     string `gorm:"default:'#36b37e'"`
	BackgroundColor string `gorm:"default:'#ffffff'"`
	TextColor       string `gorm:"default:'#172b4d'"`
	FontFamily      string `gorm:"default:'Inter, sans-serif'"`
	CustomCSS       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (s *BrandingService) CreateTheme(tenantID uint, theme *Theme) error {
	theme.TenantID = tenantID

	if err := s.db.Create(theme).Error; err != nil {
		logger.Error("Failed to create theme", zap.Error(err))
		return err
	}

	// If this is set as default, unset other defaults
	if theme.IsDefault {
		if err := s.setDefaultTheme(tenantID, theme.ID); err != nil {
			return err
		}
	}

	logger.Info("Theme created successfully",
		zap.Uint("tenant_id", tenantID),
		zap.String("theme_name", theme.Name))

	return nil
}

func (s *BrandingService) GetThemes(tenantID uint) ([]Theme, error) {
	var themes []Theme
	err := s.db.Where("tenant_id = ?", tenantID).Find(&themes).Error
	return themes, err
}

func (s *BrandingService) SetDefaultTheme(tenantID uint, themeID uint) error {
	return s.setDefaultTheme(tenantID, themeID)
}

func (s *BrandingService) setDefaultTheme(tenantID uint, themeID uint) error {
	// Unset all other defaults
	if err := s.db.Model(&Theme{}).Where("tenant_id = ?", tenantID).Update("is_default", false).Error; err != nil {
		return err
	}

	// Set new default
	if err := s.db.Model(&Theme{}).Where("id = ? AND tenant_id = ?", themeID, tenantID).Update("is_default", true).Error; err != nil {
		return err
	}

	return nil
}

// Logo Management

type Logo struct {
	ID        uint   `gorm:"primaryKey"`
	TenantID  uint   `gorm:"not null"`
	Name      string `gorm:"not null"`
	URL       string `gorm:"not null"`
	AltText   string
	Width     int
	Height    int
	IsActive  bool `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *BrandingService) UploadLogo(tenantID uint, logo *Logo) error {
	logo.TenantID = tenantID

	// Deactivate other logos
	if err := s.db.Model(&Logo{}).Where("tenant_id = ?", tenantID).Update("is_active", false).Error; err != nil {
		return err
	}

	// Activate new logo
	logo.IsActive = true

	if err := s.db.Create(logo).Error; err != nil {
		logger.Error("Failed to upload logo", zap.Error(err))
		return err
	}

	logger.Info("Logo uploaded successfully",
		zap.Uint("tenant_id", tenantID),
		zap.String("logo_name", logo.Name))

	return nil
}

func (s *BrandingService) GetActiveLogo(tenantID uint) (*Logo, error) {
	var logo Logo
	err := s.db.Where("tenant_id = ? AND is_active = ?", tenantID, true).First(&logo).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // No logo found
	}
	return &logo, err
}

// White-labeling

type WhiteLabelConfig struct {
	ID                uint   `gorm:"primaryKey"`
	TenantID          uint   `gorm:"not null"`
	CompanyName       string `gorm:"not null"`
	CompanyURL        string
	SupportEmail      string
	SupportPhone      string
	PrivacyPolicyURL  string
	TermsOfServiceURL string
	FooterText        string
	HidePoweredBy     bool `gorm:"default:false"`
	CustomFavicon     string
	CustomMetaTags    string // JSON string for custom meta tags
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (s *BrandingService) UpdateWhiteLabelConfig(tenantID uint, config *WhiteLabelConfig) error {
	config.TenantID = tenantID

	// Use upsert to create or update
	if err := s.db.Where("tenant_id = ?", tenantID).
		Assign(*config).
		FirstOrCreate(config).Error; err != nil {
		logger.Error("Failed to update white-label config", zap.Error(err))
		return err
	}

	logger.Info("White-label config updated successfully", zap.Uint("tenant_id", tenantID))
	return nil
}

func (s *BrandingService) GetWhiteLabelConfig(tenantID uint) (*WhiteLabelConfig, error) {
	var config WhiteLabelConfig
	err := s.db.Where("tenant_id = ?", tenantID).First(&config).Error
	if err == gorm.ErrRecordNotFound {
		// Return default config
		return &WhiteLabelConfig{
			TenantID:      tenantID,
			CompanyName:   "Status Page",
			HidePoweredBy: false,
		}, nil
	}
	return &config, err
}

// Branding Package (combines all branding elements)

type BrandingPackage struct {
	TenantID     uint
	CustomDomain *CustomDomain
	Theme        *Theme
	Logo         *Logo
	WhiteLabel   *WhiteLabelConfig
	Branding     *models.Branding
}

func (s *BrandingService) GetBrandingPackage(tenantID uint) (*BrandingPackage, error) {
	brandingPackage := &BrandingPackage{
		TenantID: tenantID,
	}

	// Get custom domain
	var domain CustomDomain
	if err := s.db.Where("tenant_id = ? AND verified = ?", tenantID, true).First(&domain).Error; err == nil {
		brandingPackage.CustomDomain = &domain
	}

	// Get active theme
	var theme Theme
	if err := s.db.Where("tenant_id = ? AND is_default = ?", tenantID, true).First(&theme).Error; err == nil {
		brandingPackage.Theme = &theme
	}

	// Get active logo
	logo, err := s.GetActiveLogo(tenantID)
	if err == nil && logo != nil {
		brandingPackage.Logo = logo
	}

	// Get white-label config
	whiteLabel, err := s.GetWhiteLabelConfig(tenantID)
	if err == nil {
		brandingPackage.WhiteLabel = whiteLabel
	}

	// Get branding
	var branding models.Branding
	if err := s.db.Where("tenant_id = ?", tenantID).First(&branding).Error; err == nil {
		brandingPackage.Branding = &branding
	}

	return brandingPackage, nil
}

// CSS Generation

func (s *BrandingService) GenerateCustomCSS(tenantID uint) (string, error) {
	brandingPackage, err := s.GetBrandingPackage(tenantID)
	if err != nil {
		return "", err
	}

	css := ":root {\n"

	if brandingPackage.Theme != nil {
		css += fmt.Sprintf("  --primary-color: %s;\n", brandingPackage.Theme.PrimaryColor)
		css += fmt.Sprintf("  --secondary-color: %s;\n", brandingPackage.Theme.SecondaryColor)
		css += fmt.Sprintf("  --accent-color: %s;\n", brandingPackage.Theme.AccentColor)
		css += fmt.Sprintf("  --background-color: %s;\n", brandingPackage.Theme.BackgroundColor)
		css += fmt.Sprintf("  --text-color: %s;\n", brandingPackage.Theme.TextColor)
		css += fmt.Sprintf("  --font-family: %s;\n", brandingPackage.Theme.FontFamily)
	}

	css += "}\n\n"

	// Add custom CSS from theme
	if brandingPackage.Theme != nil && brandingPackage.Theme.CustomCSS != "" {
		css += brandingPackage.Theme.CustomCSS + "\n"
	}

	// Add custom CSS from branding
	if brandingPackage.Branding != nil && brandingPackage.Branding.CustomCSS != "" {
		css += brandingPackage.Branding.CustomCSS + "\n"
	}

	return css, nil
}

// Private helper methods

func (s *BrandingService) generateSSLCertificate(domain *CustomDomain) {
	// This would integrate with Let's Encrypt or similar SSL certificate provider
	// For now, we'll just simulate the process

	time.Sleep(5 * time.Second) // Simulate certificate generation time

	domain.SSLStatus = "active"
	s.db.Save(domain)

	logger.Info("SSL certificate generated", zap.String("domain", domain.Domain))
}

func (s *BrandingService) checkDomainVerification(domain, verificationCode string) (bool, error) {
	// This would check DNS records or file-based verification
	// For now, we'll simulate verification
	return true, nil
}

// Branding Analytics

type BrandingAnalytics struct {
	ID                 uint      `gorm:"primaryKey"`
	TenantID           uint      `gorm:"not null"`
	PageViews          int64     `gorm:"default:0"`
	UniqueVisitors     int64     `gorm:"default:0"`
	BounceRate         float64   `gorm:"default:0"`
	AvgSessionDuration float64   `gorm:"default:0"`
	Date               time.Time `gorm:"not null"`
	CreatedAt          time.Time
}

func (s *BrandingService) RecordBrandingAnalytics(tenantID uint, analytics *BrandingAnalytics) error {
	analytics.TenantID = tenantID

	// Use upsert to create or update daily analytics
	if err := s.db.Where("tenant_id = ? AND date = ?", tenantID, analytics.Date).
		Assign(analytics).
		FirstOrCreate(analytics).Error; err != nil {
		logger.Error("Failed to record branding analytics", zap.Error(err))
		return err
	}

	return nil
}

func (s *BrandingService) GetBrandingAnalytics(tenantID uint, days int) ([]BrandingAnalytics, error) {
	var analytics []BrandingAnalytics
	startDate := time.Now().AddDate(0, 0, -days)

	err := s.db.Where("tenant_id = ? AND date >= ?", tenantID, startDate).
		Order("date DESC").
		Find(&analytics).Error

	return analytics, err
}

// GetBrandingByTenantID retrieves branding information for a specific tenant
func (s *BrandingService) GetBrandingByTenantID(tenantID uint) (*models.Branding, error) {
	var branding models.Branding
	if err := s.db.Where("tenant_id = ?", tenantID).First(&branding).Error; err != nil {
		return nil, err
	}
	return &branding, nil
}

// UpdateBranding updates branding information for a tenant
func (s *BrandingService) UpdateBranding(branding *models.Branding) error {
	if err := s.db.Save(branding).Error; err != nil {
		logger.Error("Failed to update branding", zap.Error(err))
		return err
	}
	return nil
}
