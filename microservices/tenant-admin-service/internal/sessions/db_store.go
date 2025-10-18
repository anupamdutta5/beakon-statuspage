package sessions

import (
	"context"
	"fmt"
	"time"

	"github.com/anupamdutta5/tenant-admin-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DBSessionStore implements session storage using database
type DBSessionStore struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewDBSessionStore creates a new database-based session store
func NewDBSessionStore(db *gorm.DB, logger *zap.Logger) *DBSessionStore {
	return &DBSessionStore{
		db:     db,
		logger: logger,
	}
}

// Create creates a new session in database
func (s *DBSessionStore) Create(ctx context.Context, session *models.Session) error {
	if err := s.db.WithContext(ctx).Create(session).Error; err != nil {
		s.logger.Error("Failed to create session in database",
			zap.String("session_id", session.ID),
			zap.Error(err))
		return fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.Debug("Session created in database",
		zap.String("session_id", session.ID),
		zap.Uint("user_id", session.UserID))

	return nil
}

// Get retrieves a session from database
func (s *DBSessionStore) Get(ctx context.Context, sessionID string) (*models.Session, error) {
	var session models.Session

	err := s.db.WithContext(ctx).
		Where("id = ? AND expires_at > ?", sessionID, time.Now()).
		First(&session).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Session doesn't exist or expired
		}
		s.logger.Error("Failed to get session from database",
			zap.String("session_id", sessionID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// Update updates a session in database (extends expiry)
func (s *DBSessionStore) Update(ctx context.Context, sessionID string) error {
	result := s.db.WithContext(ctx).Model(&models.Session{}).
		Where("id = ? AND expires_at > ?", sessionID, time.Now()).
		Updates(map[string]interface{}{
			"last_seen": time.Now(),
			"expires_at": time.Now().Add(24 * time.Hour), // Extend by 24 hours
		})

	if result.Error != nil {
		s.logger.Error("Failed to update session in database",
			zap.String("session_id", sessionID),
			zap.Error(result.Error))
		return fmt.Errorf("failed to update session: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("session not found or expired")
	}

	s.logger.Debug("Session updated in database",
		zap.String("session_id", sessionID))

	return nil
}

// Delete removes a session from database
func (s *DBSessionStore) Delete(ctx context.Context, sessionID string) error {
	result := s.db.WithContext(ctx).
		Where("id = ?", sessionID).
		Delete(&models.Session{})

	if result.Error != nil {
		s.logger.Error("Failed to delete session from database",
			zap.String("session_id", sessionID),
			zap.Error(result.Error))
		return fmt.Errorf("failed to delete session: %w", result.Error)
	}

	s.logger.Debug("Session deleted from database",
		zap.String("session_id", sessionID),
		zap.Int64("rows_affected", result.RowsAffected))

	return nil
}

// GetUserSessions gets all active sessions for a user
func (s *DBSessionStore) GetUserSessions(ctx context.Context, userID uint) ([]*models.Session, error) {
	var sessions []*models.Session

	err := s.db.WithContext(ctx).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		Find(&sessions).Error

	if err != nil {
		s.logger.Error("Failed to get user sessions from database",
			zap.Uint("user_id", userID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	s.logger.Debug("Retrieved user sessions from database",
		zap.Uint("user_id", userID),
		zap.Int("count", len(sessions)))

	return sessions, nil
}

// DeleteUserSessions deletes all sessions for a user
func (s *DBSessionStore) DeleteUserSessions(ctx context.Context, userID uint) error {
	result := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.Session{})

	if result.Error != nil {
		s.logger.Error("Failed to delete user sessions from database",
			zap.Uint("user_id", userID),
			zap.Error(result.Error))
		return fmt.Errorf("failed to delete user sessions: %w", result.Error)
	}

	s.logger.Info("Deleted user sessions from database",
		zap.Uint("user_id", userID),
		zap.Int64("count", result.RowsAffected))

	return nil
}

// CleanupExpiredSessions removes expired sessions from database
func (s *DBSessionStore) CleanupExpiredSessions(ctx context.Context) error {
	result := s.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.Session{})

	if result.Error != nil {
		s.logger.Error("Failed to cleanup expired sessions",
			zap.Error(result.Error))
		return fmt.Errorf("failed to cleanup expired sessions: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		s.logger.Info("Cleaned up expired sessions",
			zap.Int64("count", result.RowsAffected))
	}

	return nil
}

// HealthCheck checks if database is accessible
func (s *DBSessionStore) HealthCheck(ctx context.Context) error {
	// Try a simple query
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.Session{}).Count(&count).Error; err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}