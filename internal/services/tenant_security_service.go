package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TenantSecurityService handles tenant data isolation and security
type TenantSecurityService struct {
	db *gorm.DB
}

// NewTenantSecurityService creates a new tenant security service
func NewTenantSecurityService(db *gorm.DB) *TenantSecurityService {
	return &TenantSecurityService{
		db: db,
	}
}

// SecurityConfig represents tenant-specific security configuration
type SecurityConfig struct {
	TenantID           uint      `json:"tenant_id"`
	DataEncryption     bool      `json:"data_encryption"`
	AuditLogging       bool      `json:"audit_logging"`
	IPWhitelist        []string  `json:"ip_whitelist"`
	SessionTimeout     int       `json:"session_timeout"` // minutes
	PasswordPolicy     string    `json:"password_policy"` // weak, medium, strong
	TwoFactorAuth      bool      `json:"two_factor_auth"`
	APIKeyRotation     int       `json:"api_key_rotation"` // days
	LastSecurityUpdate time.Time `json:"last_security_update"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// SecurityAuditLog represents an audit log entry
type SecurityAuditLog struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	UserID    uint      `json:"user_id,omitempty"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Details   string    `json:"details"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
}

// DataAccessLog represents data access logging
type DataAccessLog struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	UserID    uint      `json:"user_id,omitempty"`
	Table     string    `json:"table"`
	Operation string    `json:"operation"` // SELECT, INSERT, UPDATE, DELETE
	RecordID  string    `json:"record_id"`
	IPAddress string    `json:"ip_address"`
	Timestamp time.Time `json:"timestamp"`
}

// SecurityViolation represents a security violation
type SecurityViolation struct {
	ID          uint      `json:"id"`
	TenantID    uint      `json:"tenant_id"`
	Type        string    `json:"type"`     // unauthorized_access, data_breach, suspicious_activity
	Severity    string    `json:"severity"` // low, medium, high, critical
	Description string    `json:"description"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	Details     string    `json:"details"`
	Timestamp   time.Time `json:"timestamp"`
	Resolved    bool      `json:"resolved"`
}

// GetSecurityConfig retrieves security configuration for a tenant
func (s *TenantSecurityService) GetSecurityConfig(ctx context.Context, tenantID uint) (*SecurityConfig, error) {
	// For now, return default configuration
	// In a real implementation, you'd store this in the database
	config := SecurityConfig{
		TenantID:           tenantID,
		DataEncryption:     true,
		AuditLogging:       true,
		IPWhitelist:        []string{},
		SessionTimeout:     60, // 1 hour
		PasswordPolicy:     "strong",
		TwoFactorAuth:      false,
		APIKeyRotation:     90, // 90 days
		LastSecurityUpdate: time.Now(),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	return &config, nil
}

// UpdateSecurityConfig updates security configuration for a tenant
func (s *TenantSecurityService) UpdateSecurityConfig(ctx context.Context, tenantID uint, config *SecurityConfig) error {
	config.TenantID = tenantID
	config.UpdatedAt = time.Now()
	config.LastSecurityUpdate = time.Now()

	// In a real implementation, you'd save this to the database
	// For now, we'll just log the update
	logger.Info("Security configuration updated",
		zap.Uint("tenant_id", tenantID),
		zap.Bool("data_encryption", config.DataEncryption),
		zap.Bool("audit_logging", config.AuditLogging),
		zap.Bool("two_factor_auth", config.TwoFactorAuth))

	return nil
}

// LogAuditEvent logs an audit event
func (s *TenantSecurityService) LogAuditEvent(ctx context.Context, tenantID uint, event *SecurityAuditLog) error {
	event.TenantID = tenantID
	event.Timestamp = time.Now()

	// In a real implementation, you'd save this to the database
	logger.Info("Audit event logged",
		zap.Uint("tenant_id", tenantID),
		zap.String("action", event.Action),
		zap.String("resource", event.Resource),
		zap.String("ip_address", event.IPAddress),
		zap.Bool("success", event.Success))

	return nil
}

// LogDataAccess logs data access for compliance
func (s *TenantSecurityService) LogDataAccess(ctx context.Context, tenantID uint, access *DataAccessLog) error {
	access.TenantID = tenantID
	access.Timestamp = time.Now()

	// In a real implementation, you'd save this to the database
	logger.Info("Data access logged",
		zap.Uint("tenant_id", tenantID),
		zap.String("table", access.Table),
		zap.String("operation", access.Operation),
		zap.String("record_id", access.RecordID),
		zap.String("ip_address", access.IPAddress))

	return nil
}

// RecordSecurityViolation records a security violation
func (s *TenantSecurityService) RecordSecurityViolation(ctx context.Context, tenantID uint, violation *SecurityViolation) error {
	violation.TenantID = tenantID
	violation.Timestamp = time.Now()

	// In a real implementation, you'd save this to the database
	logger.Warn("Security violation recorded",
		zap.Uint("tenant_id", tenantID),
		zap.String("type", violation.Type),
		zap.String("severity", violation.Severity),
		zap.String("description", violation.Description),
		zap.String("ip_address", violation.IPAddress))

	return nil
}

// EncryptData encrypts sensitive data for a tenant
func (s *TenantSecurityService) EncryptData(ctx context.Context, tenantID uint, data string) (string, error) {
	// Generate tenant-specific encryption key
	key := s.generateTenantKey(tenantID)

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data
	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	return encoded, nil
}

// DecryptData decrypts sensitive data for a tenant
func (s *TenantSecurityService) DecryptData(ctx context.Context, tenantID uint, encryptedData string) (string, error) {
	// Generate tenant-specific encryption key
	key := s.generateTenantKey(tenantID)

	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode data: %w", err)
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %w", err)
	}

	return string(plaintext), nil
}

// ValidateTenantAccess validates if a request has access to tenant data
func (s *TenantSecurityService) ValidateTenantAccess(ctx context.Context, tenantID uint, userID uint, resource string, action string) (bool, error) {
	// Check if user belongs to tenant
	var user models.User
	if err := s.db.Where("id = ? AND tenant_id = ?", userID, tenantID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, fmt.Errorf("user not found in tenant")
		}
		return false, fmt.Errorf("failed to validate user: %w", err)
	}

	// Check user permissions
	// In a real implementation, you'd check RBAC permissions
	// For now, we'll allow access if user belongs to tenant
	return true, nil
}

// CheckIPWhitelist checks if an IP address is whitelisted for a tenant
func (s *TenantSecurityService) CheckIPWhitelist(ctx context.Context, tenantID uint, ipAddress string) (bool, error) {
	config, err := s.GetSecurityConfig(ctx, tenantID)
	if err != nil {
		return false, err
	}

	// If no whitelist configured, allow all IPs
	if len(config.IPWhitelist) == 0 {
		return true, nil
	}

	// Check if IP is in whitelist
	for _, allowedIP := range config.IPWhitelist {
		if ipAddress == allowedIP {
			return true, nil
		}
	}

	return false, nil
}

// GetSecurityMetrics returns security metrics for a tenant
func (s *TenantSecurityService) GetSecurityMetrics(ctx context.Context, tenantID uint, days int) (map[string]interface{}, error) {
	// In a real implementation, you'd query the database for actual metrics
	// For now, we'll return mock data
	metrics := map[string]interface{}{
		"tenant_id":   tenantID,
		"period_days": days,
		"audit_events": map[string]interface{}{
			"total":      1250,
			"successful": 1180,
			"failed":     70,
			"by_action": map[string]int{
				"login":                450,
				"data_access":          320,
				"configuration_change": 180,
				"user_management":      150,
				"other":                150,
			},
		},
		"security_violations": map[string]interface{}{
			"total": 12,
			"by_severity": map[string]int{
				"low":      8,
				"medium":   3,
				"high":     1,
				"critical": 0,
			},
			"by_type": map[string]int{
				"unauthorized_access": 5,
				"suspicious_activity": 4,
				"data_breach":         2,
				"other":               1,
			},
		},
		"data_access": map[string]interface{}{
			"total_requests": 3450,
			"unique_users":   25,
			"by_table": map[string]int{
				"incidents": 1200,
				"services":  800,
				"users":     600,
				"settings":  450,
				"other":     400,
			},
		},
		"compliance": map[string]interface{}{
			"data_encryption_enabled": true,
			"audit_logging_enabled":   true,
			"two_factor_auth_enabled": false,
			"ip_whitelist_configured": false,
			"last_security_audit":     time.Now().AddDate(0, 0, -7),
		},
	}

	return metrics, nil
}

// GenerateAPIKey generates a new API key for a tenant
func (s *TenantSecurityService) GenerateAPIKey(ctx context.Context, tenantID uint, name string) (string, error) {
	// Generate random API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("failed to generate API key: %w", err)
	}

	// Encode to base64
	apiKey := base64.URLEncoding.EncodeToString(keyBytes)

	// In a real implementation, you'd save this to the database with expiration
	logger.Info("API key generated",
		zap.Uint("tenant_id", tenantID),
		zap.String("name", name),
		zap.String("key_prefix", apiKey[:8]))

	return apiKey, nil
}

// ValidateAPIKey validates an API key for a tenant
func (s *TenantSecurityService) ValidateAPIKey(ctx context.Context, apiKey string) (uint, error) {
	// In a real implementation, you'd look up the API key in the database
	// For now, we'll return a mock tenant ID
	return 1, nil
}

// Helper functions

// generateTenantKey generates a tenant-specific encryption key
func (s *TenantSecurityService) generateTenantKey(tenantID uint) []byte {
	// In a real implementation, you'd use a more secure key derivation
	// For now, we'll use a simple hash of tenant ID
	hash := sha256.Sum256([]byte(fmt.Sprintf("tenant_key_%d", tenantID)))
	return hash[:]
}

// GetTenantDataScope returns a GORM scope for tenant data isolation
func (s *TenantSecurityService) GetTenantDataScope(tenantID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", tenantID)
	}
}

// ApplyTenantScope applies tenant isolation to a GORM query
func (s *TenantSecurityService) ApplyTenantScope(db *gorm.DB, tenantID uint) *gorm.DB {
	return db.Where("tenant_id = ?", tenantID)
}
