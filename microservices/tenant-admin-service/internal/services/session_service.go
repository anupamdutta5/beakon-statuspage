package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/anupamdutta5/tenant-admin-service/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionExpired      = errors.New("session has expired")
	ErrSessionInactive     = errors.New("session is not active")
	ErrRefreshTokenInvalid = errors.New("invalid refresh token")
)

// SessionService handles session and refresh token operations for tenant admin users
type SessionService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewSessionService creates a new session service
func NewSessionService(db *gorm.DB, logger *zap.Logger) *SessionService {
	return &SessionService{
		db:     db,
		logger: logger,
	}
}

// CreateRefreshToken creates a new refresh token for long-lived authentication
// Returns a secure random token that expires in 7 days
func (s *SessionService) CreateRefreshToken(ctx context.Context, userID, tenantID uuid.UUID, ipAddress, userAgent string) (string, error) {
	// Generate secure random token (32 bytes = 256 bits)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		s.logger.Error("Failed to generate random token", zap.Error(err))
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	userSession := &models.UserSession{
		UserID:    userID,
		TenantID:  tenantID,
		Token:     token,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // Refresh token lasts 7 days
		IPAddress: ipAddress,
		UserAgent: userAgent,
		IsActive:  true,
	}

	if err := s.db.WithContext(ctx).Create(userSession).Error; err != nil {
		s.logger.Error("Failed to create refresh token",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("tenant_id", tenantID.String()))
		return "", err
	}

	s.logger.Info("Refresh token created",
		zap.String("user_id", userID.String()),
		zap.String("tenant_id", tenantID.String()),
		zap.Time("expires_at", userSession.ExpiresAt))

	return token, nil
}

// ValidateRefreshToken validates a refresh token and returns the user ID and tenant ID
func (s *SessionService) ValidateRefreshToken(ctx context.Context, token string) (userID, tenantID uuid.UUID, err error) {
	var userSession models.UserSession
	result := s.db.WithContext(ctx).
		Where("token = ?", token).
		First(&userSession)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return uuid.Nil, uuid.Nil, ErrRefreshTokenInvalid
		}
		s.logger.Error("Failed to validate refresh token", zap.Error(result.Error))
		return uuid.Nil, uuid.Nil, result.Error
	}

	// Check if token is valid
	if !userSession.IsValid() {
		s.logger.Warn("Invalid refresh token used",
			zap.String("user_id", userSession.UserID.String()),
			zap.Bool("is_active", userSession.IsActive),
			zap.Bool("is_expired", userSession.IsExpired()))
		return uuid.Nil, uuid.Nil, ErrRefreshTokenInvalid
	}

	return userSession.UserID, userSession.TenantID, nil
}

// RevokeRefreshToken revokes a specific refresh token (e.g., on logout)
func (s *SessionService) RevokeRefreshToken(ctx context.Context, token string) error {
	result := s.db.WithContext(ctx).
		Model(&models.UserSession{}).
		Where("token = ?", token).
		Update("is_active", false)

	if result.Error != nil {
		s.logger.Error("Failed to revoke refresh token", zap.Error(result.Error))
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrRefreshTokenInvalid
	}

	s.logger.Info("Refresh token revoked", zap.String("token_prefix", token[:10]+"..."))
	return nil
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a user (e.g., on password change)
func (s *SessionService) RevokeAllUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	result := s.db.WithContext(ctx).
		Model(&models.UserSession{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Update("is_active", false)

	if result.Error != nil {
		s.logger.Error("Failed to revoke all user refresh tokens",
			zap.Error(result.Error),
			zap.String("user_id", userID.String()))
		return result.Error
	}

	s.logger.Info("All user refresh tokens revoked",
		zap.String("user_id", userID.String()),
		zap.Int64("tokens_revoked", result.RowsAffected))

	return nil
}

// RotateRefreshToken revokes the old token and creates a new one (security best practice)
// This should be called every time a refresh token is used
func (s *SessionService) RotateRefreshToken(ctx context.Context, oldToken string, userID, tenantID uuid.UUID, ipAddress, userAgent string) (string, error) {
	// Start a transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Revoke old token
	if err := tx.Model(&models.UserSession{}).
		Where("token = ?", oldToken).
		Update("is_active", false).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to revoke old refresh token during rotation", zap.Error(err))
		return "", err
	}

	// Generate new token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		tx.Rollback()
		s.logger.Error("Failed to generate new token during rotation", zap.Error(err))
		return "", err
	}
	newToken := base64.URLEncoding.EncodeToString(tokenBytes)

	// Create new session
	userSession := &models.UserSession{
		UserID:    userID,
		TenantID:  tenantID,
		Token:     newToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		IsActive:  true,
	}

	if err := tx.Create(userSession).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create new refresh token during rotation", zap.Error(err))
		return "", err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		s.logger.Error("Failed to commit token rotation transaction", zap.Error(err))
		return "", err
	}

	s.logger.Info("Refresh token rotated",
		zap.String("user_id", userID.String()),
		zap.String("tenant_id", tenantID.String()))

	return newToken, nil
}

// CleanupExpiredRefreshTokens removes expired refresh tokens from the database
// Should be run periodically (e.g., daily cron job)
func (s *SessionService) CleanupExpiredRefreshTokens(ctx context.Context) error {
	result := s.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.UserSession{})

	if result.Error != nil {
		s.logger.Error("Failed to cleanup expired refresh tokens", zap.Error(result.Error))
		return result.Error
	}

	s.logger.Info("Cleaned up expired refresh tokens",
		zap.Int64("tokens_deleted", result.RowsAffected))

	return nil
}

// GetActiveUserSessions returns all active sessions for a user (for UI display)
func (s *SessionService) GetActiveUserSessions(ctx context.Context, userID uuid.UUID) ([]models.UserSession, error) {
	var sessions []models.UserSession
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ? AND expires_at > ?", userID, true, time.Now()).
		Order("created_at DESC").
		Find(&sessions).Error

	if err != nil {
		s.logger.Error("Failed to get active user sessions",
			zap.Error(err),
			zap.String("user_id", userID.String()))
		return nil, err
	}

	return sessions, nil
}
