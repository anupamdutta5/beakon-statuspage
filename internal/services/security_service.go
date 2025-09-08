package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"

	// "github.com/pquerna/otp/totp" // Mocked for now
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SecurityService struct {
	db *gorm.DB
}

// MockTOTPKey represents a mock TOTP key for development
type MockTOTPKey struct {
	Secret string
	URL    string
}

func NewSecurityService() *SecurityService {
	return &SecurityService{
		db: database.DB,
	}
}

// mockValidateTOTP provides a mock TOTP validation for development
func (s *SecurityService) mockValidateTOTP(code, secret string) bool {
	// For development, accept any 6-digit code
	return len(code) == 6 && code >= "000000" && code <= "999999"
}

// Two-Factor Authentication

type TwoFactorAuth struct {
	ID          uint     `gorm:"primaryKey"`
	UserID      uint     `gorm:"not null"`
	Secret      string   `gorm:"not null"`
	IsEnabled   bool     `gorm:"default:false"`
	BackupCodes []string `gorm:"type:text"` // JSON array of backup codes
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *SecurityService) Generate2FASecret(userID uint) (string, string, error) {
	// Generate TOTP secret
	// Mock TOTP generation for development
	key := &MockTOTPKey{
		Secret: "MOCK_SECRET_KEY_12345678901234567890",
		URL:    fmt.Sprintf("otpauth://totp/Status%%20Page:user_%d?secret=MOCK_SECRET_KEY_12345678901234567890&issuer=Status%%20Page", userID),
	}

	// Generate backup codes
	backupCodes := s.generateBackupCodes()

	// Save 2FA record
	twoFA := &TwoFactorAuth{
		UserID:      userID,
		Secret:      key.Secret,
		IsEnabled:   false,
		BackupCodes: backupCodes,
	}

	// Use upsert to create or update
	if err := s.db.Where("user_id = ?", userID).
		Assign(*twoFA).
		FirstOrCreate(twoFA).Error; err != nil {
		logger.Error("Failed to save 2FA secret", zap.Error(err))
		return "", "", err
	}

	return key.URL, key.Secret, nil
}

func (s *SecurityService) Verify2FACode(userID uint, code string) (bool, error) {
	var twoFA TwoFactorAuth
	if err := s.db.Where("user_id = ? AND is_enabled = ?", userID, true).First(&twoFA).Error; err != nil {
		return false, err
	}

	// Verify TOTP code
	// Mock TOTP validation for development
	valid := s.mockValidateTOTP(code, twoFA.Secret)
	if valid {
		return true, nil
	}

	// Check backup codes
	for i, backupCode := range twoFA.BackupCodes {
		if backupCode == code {
			// Remove used backup code
			twoFA.BackupCodes = append(twoFA.BackupCodes[:i], twoFA.BackupCodes[i+1:]...)
			s.db.Save(&twoFA)
			return true, nil
		}
	}

	return false, nil
}

func (s *SecurityService) Enable2FA(userID uint, code string) error {
	var twoFA TwoFactorAuth
	if err := s.db.Where("user_id = ?", userID).First(&twoFA).Error; err != nil {
		return err
	}

	// Verify the code before enabling
	valid, err := s.Verify2FACode(userID, code)
	if err != nil || !valid {
		return fmt.Errorf("invalid 2FA code")
	}

	twoFA.IsEnabled = true
	return s.db.Save(&twoFA).Error
}

func (s *SecurityService) Disable2FA(userID uint) error {
	return s.db.Model(&TwoFactorAuth{}).Where("user_id = ?", userID).Update("is_enabled", false).Error
}

func (s *SecurityService) Is2FAEnabled(userID uint) bool {
	var twoFA TwoFactorAuth
	if err := s.db.Where("user_id = ? AND is_enabled = ?", userID, true).First(&twoFA).Error; err != nil {
		return false
	}
	return true
}

// Single Sign-On (SSO)

type SSOProvider struct {
	ID           uint   `gorm:"primaryKey"`
	TenantID     uint   `gorm:"not null"`
	Name         string `gorm:"not null"`
	Type         string `gorm:"not null"` // saml, oauth2, oidc
	ClientID     string `gorm:"not null"`
	ClientSecret string `gorm:"not null"`
	IssuerURL    string `gorm:"not null"`
	RedirectURL  string `gorm:"not null"`
	IsActive     bool   `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type SSOSession struct {
	ID           uint   `gorm:"primaryKey"`
	UserID       uint   `gorm:"not null"`
	ProviderID   uint   `gorm:"not null"`
	ExternalID   string `gorm:"not null"`
	AccessToken  string `gorm:"not null"`
	RefreshToken string
	ExpiresAt    time.Time `gorm:"not null"`
	CreatedAt    time.Time
}

func (s *SecurityService) CreateSSOProvider(tenantID uint, provider *SSOProvider) error {
	provider.TenantID = tenantID

	if err := s.db.Create(provider).Error; err != nil {
		logger.Error("Failed to create SSO provider", zap.Error(err))
		return err
	}

	logger.Info("SSO provider created successfully",
		zap.Uint("tenant_id", tenantID),
		zap.String("name", provider.Name))

	return nil
}

func (s *SecurityService) GetSSOProviders(tenantID uint) ([]SSOProvider, error) {
	var providers []SSOProvider
	err := s.db.Where("tenant_id = ? AND is_active = ?", tenantID, true).Find(&providers).Error
	return providers, err
}

func (s *SecurityService) CreateSSOSession(userID, providerID uint, externalID, accessToken, refreshToken string, expiresAt time.Time) error {
	session := &SSOSession{
		UserID:       userID,
		ProviderID:   providerID,
		ExternalID:   externalID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}

	return s.db.Create(session).Error
}

// IP Whitelisting

type IPWhitelist struct {
	ID          uint   `gorm:"primaryKey"`
	TenantID    uint   `gorm:"not null"`
	IPAddress   string `gorm:"not null"`
	CIDR        string // For IP ranges
	Description string
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *SecurityService) AddIPToWhitelist(tenantID uint, ipAddress, cidr, description string) error {
	whitelist := &IPWhitelist{
		TenantID:    tenantID,
		IPAddress:   ipAddress,
		CIDR:        cidr,
		Description: description,
		IsActive:    true,
	}

	if err := s.db.Create(whitelist).Error; err != nil {
		logger.Error("Failed to add IP to whitelist", zap.Error(err))
		return err
	}

	logger.Info("IP added to whitelist",
		zap.Uint("tenant_id", tenantID),
		zap.String("ip_address", ipAddress))

	return nil
}

func (s *SecurityService) IsIPWhitelisted(tenantID uint, ipAddress string) bool {
	var whitelist []IPWhitelist
	if err := s.db.Where("tenant_id = ? AND is_active = ?", tenantID, true).Find(&whitelist).Error; err != nil {
		return false
	}

	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return false
	}

	for _, entry := range whitelist {
		// Check exact IP match
		if entry.IPAddress == ipAddress {
			return true
		}

		// Check CIDR range
		if entry.CIDR != "" {
			_, network, err := net.ParseCIDR(entry.CIDR)
			if err == nil && network.Contains(ip) {
				return true
			}
		}
	}

	return false
}

func (s *SecurityService) GetIPWhitelist(tenantID uint) ([]IPWhitelist, error) {
	var whitelist []IPWhitelist
	err := s.db.Where("tenant_id = ?", tenantID).Find(&whitelist).Error
	return whitelist, err
}

func (s *SecurityService) RemoveIPFromWhitelist(whitelistID uint) error {
	return s.db.Delete(&IPWhitelist{}, whitelistID).Error
}

// Audit Logging

type AuditLog struct {
	ID         uint   `gorm:"primaryKey"`
	TenantID   uint   `gorm:"not null"`
	UserID     uint   `gorm:"not null"`
	Action     string `gorm:"not null"` // login, logout, create, update, delete
	Resource   string `gorm:"not null"` // user, service, incident, etc.
	ResourceID uint   // ID of the affected resource
	Details    string // JSON details of the action
	IPAddress  string
	UserAgent  string
	Success    bool `gorm:"default:true"`
	CreatedAt  time.Time
}

func (s *SecurityService) LogAuditEvent(tenantID, userID uint, action, resource string, resourceID uint, details map[string]interface{}, ipAddress, userAgent string, success bool) error {
	detailsJSON := "{}"
	if details != nil {
		// Convert details to JSON (simplified for now)
		detailsJSON = fmt.Sprintf(`{"action": "%s", "resource": "%s"}`, action, resource)
	}

	auditLog := &AuditLog{
		TenantID:   tenantID,
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    detailsJSON,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Success:    success,
	}

	return s.db.Create(auditLog).Error
}

func (s *SecurityService) GetAuditLogs(tenantID uint, days int) ([]AuditLog, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	var logs []AuditLog
	err := s.db.Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Order("created_at DESC").
		Find(&logs).Error

	return logs, err
}

func (s *SecurityService) GetAuditLogsByUser(tenantID, userID uint, days int) ([]AuditLog, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	var logs []AuditLog
	err := s.db.Where("tenant_id = ? AND user_id = ? AND created_at >= ?", tenantID, userID, startDate).
		Order("created_at DESC").
		Find(&logs).Error

	return logs, err
}

// Security Policies

type SecurityPolicy struct {
	ID                       uint `gorm:"primaryKey"`
	TenantID                 uint `gorm:"not null"`
	PasswordMinLength        int  `gorm:"default:8"`
	PasswordRequireUppercase bool `gorm:"default:true"`
	PasswordRequireLowercase bool `gorm:"default:true"`
	PasswordRequireNumbers   bool `gorm:"default:true"`
	PasswordRequireSymbols   bool `gorm:"default:true"`
	PasswordExpiryDays       int  `gorm:"default:90"`
	SessionTimeoutMinutes    int  `gorm:"default:60"`
	MaxLoginAttempts         int  `gorm:"default:5"`
	LockoutDurationMinutes   int  `gorm:"default:15"`
	Require2FA               bool `gorm:"default:false"`
	RequireSSO               bool `gorm:"default:false"`
	IPWhitelistEnabled       bool `gorm:"default:false"`
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

func (s *SecurityService) GetSecurityPolicy(tenantID uint) (*SecurityPolicy, error) {
	var policy SecurityPolicy
	err := s.db.Where("tenant_id = ?", tenantID).First(&policy).Error
	if err == gorm.ErrRecordNotFound {
		// Return default policy
		return &SecurityPolicy{
			TenantID:                 tenantID,
			PasswordMinLength:        8,
			PasswordRequireUppercase: true,
			PasswordRequireLowercase: true,
			PasswordRequireNumbers:   true,
			PasswordRequireSymbols:   true,
			PasswordExpiryDays:       90,
			SessionTimeoutMinutes:    60,
			MaxLoginAttempts:         5,
			LockoutDurationMinutes:   15,
			Require2FA:               false,
			RequireSSO:               false,
			IPWhitelistEnabled:       false,
		}, nil
	}
	return &policy, err
}

func (s *SecurityService) UpdateSecurityPolicy(tenantID uint, policy *SecurityPolicy) error {
	policy.TenantID = tenantID

	// Use upsert to create or update
	if err := s.db.Where("tenant_id = ?", tenantID).
		Assign(*policy).
		FirstOrCreate(policy).Error; err != nil {
		logger.Error("Failed to update security policy", zap.Error(err))
		return err
	}

	logger.Info("Security policy updated successfully", zap.Uint("tenant_id", tenantID))
	return nil
}

// Login Attempt Tracking

type LoginAttempt struct {
	ID        uint   `gorm:"primaryKey"`
	TenantID  uint   `gorm:"not null"`
	UserID    uint   `gorm:"not null"`
	IPAddress string `gorm:"not null"`
	UserAgent string
	Success   bool `gorm:"default:false"`
	CreatedAt time.Time
}

func (s *SecurityService) RecordLoginAttempt(tenantID, userID uint, ipAddress, userAgent string, success bool) error {
	attempt := &LoginAttempt{
		TenantID:  tenantID,
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   success,
	}

	return s.db.Create(attempt).Error
}

func (s *SecurityService) IsUserLockedOut(tenantID, userID uint, ipAddress string) bool {
	policy, err := s.GetSecurityPolicy(tenantID)
	if err != nil {
		return false
	}

	// Check recent failed attempts
	cutoffTime := time.Now().Add(-time.Duration(policy.LockoutDurationMinutes) * time.Minute)

	var failedAttempts int64
	if err := s.db.Model(&LoginAttempt{}).
		Where("tenant_id = ? AND user_id = ? AND ip_address = ? AND success = ? AND created_at >= ?",
			tenantID, userID, ipAddress, false, cutoffTime).
		Count(&failedAttempts).Error; err != nil {
		return false
	}

	return failedAttempts >= int64(policy.MaxLoginAttempts)
}

// Private helper methods

func (s *SecurityService) generateBackupCodes() []string {
	codes := make([]string, 10)
	for i := 0; i < 10; i++ {
		bytes := make([]byte, 4)
		if _, err := rand.Read(bytes); err != nil {
			logger.Log.Error("Failed to generate random bytes for backup code", zap.Error(err))
			codes[i] = fmt.Sprintf("BACKUP%d", i)
		} else {
			codes[i] = strings.ToUpper(hex.EncodeToString(bytes))
		}
	}
	return codes
}
