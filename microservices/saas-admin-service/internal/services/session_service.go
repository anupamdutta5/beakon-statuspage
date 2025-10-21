package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/anupamdutta5/saas-admin-service/internal/models"
)

var (
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionExpired      = errors.New("session has expired")
	ErrSessionInactive     = errors.New("session is not active")
	ErrRefreshTokenInvalid = errors.New("invalid refresh token")
)

// SessionService handles session and refresh token operations
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

// CreateSession creates a new session in the database
func (s *SessionService) CreateSession(ctx context.Context, userID uint, ipAddress, userAgent string) (*models.Session, error) {
	session := &models.Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		TenantID:  0, // SaaS admin doesn't belong to a tenant
		IPAddress: ipAddress,
		UserAgent: userAgent,
		IsActive:  true,
		LastSeen:  time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // Session lasts 24 hours
	}

	if err := s.db.WithContext(ctx).Create(session).Error; err != nil {
		s.logger.Error("Failed to create session", zap.Error(err), zap.Uint("user_id", userID))
		return nil, err
	}

	s.logger.Info("Session created", zap.String("session_id", session.ID), zap.Uint("user_id", userID))
	return session, nil
}

// ValidateSession validates a session by ID and updates last_seen
func (s *SessionService) ValidateSession(ctx context.Context, sessionID string) (*models.Session, error) {
	var session models.Session
	err := s.db.WithContext(ctx).
		Where("id = ?", sessionID).
		First(&session).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		s.logger.Error("Failed to validate session", zap.Error(err), zap.String("session_id", sessionID))
		return nil, err
	}

	// Check if session is valid
	if !session.IsActive {
		return nil, ErrSessionInactive
	}

	if session.IsExpired() {
		return nil, ErrSessionExpired
	}

	// Update last_seen timestamp
	if err := s.db.WithContext(ctx).Model(&session).Update("last_seen", time.Now()).Error; err != nil {
		s.logger.Warn("Failed to update last_seen", zap.Error(err), zap.String("session_id", sessionID))
		// Don't fail validation if we can't update last_seen
	}

	return &session, nil
}

// DeleteSession marks a session as inactive (soft logout)
func (s *SessionService) DeleteSession(ctx context.Context, sessionID string) error {
	result := s.db.WithContext(ctx).
		Model(&models.Session{}).
		Where("id = ?", sessionID).
		Update("is_active", false)

	if result.Error != nil {
		s.logger.Error("Failed to delete session", zap.Error(result.Error), zap.String("session_id", sessionID))
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}

	s.logger.Info("Session deleted", zap.String("session_id", sessionID))
	return nil
}

// DeleteAllUserSessions deletes all sessions for a user (logout all devices)
func (s *SessionService) DeleteAllUserSessions(ctx context.Context, userID uint) error {
	result := s.db.WithContext(ctx).
		Model(&models.Session{}).
		Where("user_id = ?", userID).
		Update("is_active", false)

	if result.Error != nil {
		s.logger.Error("Failed to delete all user sessions", zap.Error(result.Error), zap.Uint("user_id", userID))
		return result.Error
	}

	s.logger.Info("All user sessions deleted", zap.Uint("user_id", userID), zap.Int64("count", result.RowsAffected))
	return nil
}

// CreateRefreshToken creates a new refresh token for long-lived authentication
func (s *SessionService) CreateRefreshToken(ctx context.Context, userID uint, ipAddress, userAgent string) (string, error) {
	// Generate secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		s.logger.Error("Failed to generate random token", zap.Error(err))
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	userSession := &models.UserSession{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // Refresh token lasts 7 days
		IPAddress: ipAddress,
		UserAgent: userAgent,
		IsActive:  true,
	}

	if err := s.db.WithContext(ctx).Create(userSession).Error; err != nil {
		s.logger.Error("Failed to create refresh token", zap.Error(err), zap.Uint("user_id", userID))
		return "", err
	}

	s.logger.Info("Refresh token created", zap.Uint("user_id", userID))
	return token, nil
}

// ValidateRefreshToken validates a refresh token and returns the user ID
func (s *SessionService) ValidateRefreshToken(ctx context.Context, token string) (uint, error) {
	var userSession models.UserSession
	err := s.db.WithContext(ctx).
		Where("token = ?", token).
		First(&userSession).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrRefreshTokenInvalid
		}
		s.logger.Error("Failed to validate refresh token", zap.Error(err))
		return 0, err
	}

	// Check if refresh token is valid
	if !userSession.IsActive {
		return 0, ErrRefreshTokenInvalid
	}

	if userSession.IsExpired() {
		return 0, ErrSessionExpired
	}

	return userSession.UserID, nil
}

// RevokeRefreshToken revokes a refresh token
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

	s.logger.Info("Refresh token revoked")
	return nil
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a user
func (s *SessionService) RevokeAllUserRefreshTokens(ctx context.Context, userID uint) error {
	result := s.db.WithContext(ctx).
		Model(&models.UserSession{}).
		Where("user_id = ?", userID).
		Update("is_active", false)

	if result.Error != nil {
		s.logger.Error("Failed to revoke all user refresh tokens", zap.Error(result.Error), zap.Uint("user_id", userID))
		return result.Error
	}

	s.logger.Info("All user refresh tokens revoked", zap.Uint("user_id", userID), zap.Int64("count", result.RowsAffected))
	return nil
}

// CleanupExpiredSessions removes expired sessions from the database
func (s *SessionService) CleanupExpiredSessions(ctx context.Context) error {
	result := s.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.Session{})

	if result.Error != nil {
		s.logger.Error("Failed to cleanup expired sessions", zap.Error(result.Error))
		return result.Error
	}

	s.logger.Info("Expired sessions cleaned up", zap.Int64("count", result.RowsAffected))
	return nil
}

// CleanupExpiredRefreshTokens removes expired refresh tokens from the database
func (s *SessionService) CleanupExpiredRefreshTokens(ctx context.Context) error {
	result := s.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.UserSession{})

	if result.Error != nil {
		s.logger.Error("Failed to cleanup expired refresh tokens", zap.Error(result.Error))
		return result.Error
	}

	s.logger.Info("Expired refresh tokens cleaned up", zap.Int64("count", result.RowsAffected))
	return nil
}
