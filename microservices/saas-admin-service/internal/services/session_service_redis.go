package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/anupamdutta5/saas-admin-service/internal/cache"
	"github.com/anupamdutta5/saas-admin-service/internal/models"
)

// SessionServiceRedis handles session and refresh token operations with Redis + PostgreSQL fallback
type SessionServiceRedis struct {
	db          *gorm.DB
	redis       *cache.RedisClient
	logger      *zap.Logger
	useFallback bool
}

// NewSessionServiceRedis creates a new session service with Redis support
func NewSessionServiceRedis(db *gorm.DB, redis *cache.RedisClient, logger *zap.Logger) *SessionServiceRedis {
	return &SessionServiceRedis{
		db:          db,
		redis:       redis,
		logger:      logger,
		useFallback: false,
	}
}

// CreateRefreshToken creates a new refresh token with Redis primary, PostgreSQL fallback
func (s *SessionServiceRedis) CreateRefreshToken(ctx context.Context, userID uint, ipAddress, userAgent string) (string, error) {
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
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		IPAddress: ipAddress,
		UserAgent: userAgent,
		IsActive:  true,
	}

	// Try Redis first
	if s.redis != nil && s.redis.IsHealthy(ctx) {
		sessionData, err := json.Marshal(userSession)
		if err == nil {
			redisKey := fmt.Sprintf("refresh:%s", token)
			if err := s.redis.Set(ctx, redisKey, sessionData, 7*24*time.Hour); err == nil {
				s.logger.Info("Refresh token created in Redis",
					zap.Uint("user_id", userID),
					zap.String("storage", "redis"))
				return token, nil
			} else {
				s.logger.Warn("Redis SET failed, falling back to PostgreSQL",
					zap.Error(err))
			}
		}
	}

	// Fallback to PostgreSQL
	if err := s.db.WithContext(ctx).Create(userSession).Error; err != nil {
		s.logger.Error("Failed to create refresh token in PostgreSQL", zap.Error(err), zap.Uint("user_id", userID))
		return "", err
	}

	s.logger.Info("Refresh token created in PostgreSQL",
		zap.Uint("user_id", userID),
		zap.String("storage", "postgresql"))
	return token, nil
}

// ValidateRefreshToken validates a refresh token with Redis primary, PostgreSQL fallback
func (s *SessionServiceRedis) ValidateRefreshToken(ctx context.Context, token string) (uint, error) {
	// Try Redis first
	if s.redis != nil && s.redis.IsHealthy(ctx) {
		redisKey := fmt.Sprintf("refresh:%s", token)
		var userSession models.UserSession
		err := s.redis.Get(ctx, redisKey, &userSession)
		if err == nil {
			if userSession.IsValid() {
				s.logger.Debug("Refresh token validated from Redis",
					zap.Uint("user_id", userSession.UserID))
				return userSession.UserID, nil
			}
			// Token expired, delete from Redis
			s.redis.Del(ctx, redisKey)
			return 0, ErrSessionExpired
		}
	}

	// Fallback to PostgreSQL
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

	s.logger.Debug("Refresh token validated from PostgreSQL",
		zap.Uint("user_id", userSession.UserID))
	return userSession.UserID, nil
}

// RevokeRefreshToken revokes a refresh token from both Redis and PostgreSQL
func (s *SessionServiceRedis) RevokeRefreshToken(ctx context.Context, token string) error {
	// Delete from Redis
	if s.redis != nil && s.redis.IsHealthy(ctx) {
		redisKey := fmt.Sprintf("refresh:%s", token)
		if err := s.redis.Del(ctx, redisKey); err != nil {
			s.logger.Warn("Failed to delete from Redis", zap.Error(err))
		}
	}

	// Delete from PostgreSQL
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
func (s *SessionServiceRedis) RevokeAllUserRefreshTokens(ctx context.Context, userID uint) error {
	// Get all user tokens from PostgreSQL
	var userSessions []models.UserSession
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Find(&userSessions).Error; err != nil {
		s.logger.Error("Failed to get user sessions", zap.Error(err))
		return err
	}

	// Delete from Redis
	if s.redis != nil && s.redis.IsHealthy(ctx) {
		for _, session := range userSessions {
			redisKey := fmt.Sprintf("refresh:%s", session.Token)
			if err := s.redis.Del(ctx, redisKey); err != nil {
				s.logger.Warn("Failed to delete token from Redis",
					zap.String("token", session.Token[:10]+"..."),
					zap.Error(err))
			}
		}
	}

	// Revoke in PostgreSQL
	result := s.db.WithContext(ctx).
		Model(&models.UserSession{}).
		Where("user_id = ?", userID).
		Update("is_active", false)

	if result.Error != nil {
		s.logger.Error("Failed to revoke all user refresh tokens", zap.Error(result.Error), zap.Uint("user_id", userID))
		return result.Error
	}

	s.logger.Info("All user refresh tokens revoked",
		zap.Uint("user_id", userID),
		zap.Int64("count", result.RowsAffected))
	return nil
}

// CleanupExpiredRefreshTokens removes expired refresh tokens from PostgreSQL
// Redis handles expiration automatically via TTL
func (s *SessionServiceRedis) CleanupExpiredRefreshTokens(ctx context.Context) error {
	result := s.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.UserSession{})

	if result.Error != nil {
		s.logger.Error("Failed to cleanup expired refresh tokens", zap.Error(result.Error))
		return result.Error
	}

	if result.RowsAffected > 0 {
		s.logger.Info("Expired refresh tokens cleaned up from PostgreSQL",
			zap.Int64("count", result.RowsAffected))
	}

	return nil
}
