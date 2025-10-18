package sessions

import (
	"context"
	"fmt"
	"time"

	"github.com/anupamdutta5/tenant-admin-service/internal/models"
	"go.uber.org/zap"
)

// SessionStore defines the interface for session storage
type SessionStore interface {
	Create(ctx context.Context, session *models.Session) error
	Get(ctx context.Context, sessionID string) (*models.Session, error)
	Update(ctx context.Context, sessionID string) error
	Delete(ctx context.Context, sessionID string) error
	GetUserSessions(ctx context.Context, userID uint) ([]*models.Session, error)
	DeleteUserSessions(ctx context.Context, userID uint) error
	CleanupExpiredSessions(ctx context.Context) error
	HealthCheck(ctx context.Context) error
}

// SessionManager manages sessions with multiple storage backends
type SessionManager struct {
	primaryStore   SessionStore
	fallbackStore  SessionStore
	logger         *zap.Logger
	useFallback    bool
	healthCheckTTL time.Duration
	lastHealthCheck time.Time
}

// NewSessionManager creates a new session manager with primary and optional fallback stores
func NewSessionManager(primaryStore, fallbackStore SessionStore, logger *zap.Logger) *SessionManager {
	return &SessionManager{
		primaryStore:    primaryStore,
		fallbackStore:   fallbackStore,
		logger:          logger,
		useFallback:     false,
		healthCheckTTL:  30 * time.Second,
		lastHealthCheck: time.Time{},
	}
}

// Create creates a new session
func (m *SessionManager) Create(ctx context.Context, session *models.Session) error {
	store := m.getActiveStore()

	if err := store.Create(ctx, session); err != nil {
		// If primary fails, try fallback
		if store == m.primaryStore && m.fallbackStore != nil {
			m.logger.Warn("Primary store failed, using fallback",
				zap.Error(err))
			m.useFallback = true
			return m.fallbackStore.Create(ctx, session)
		}
		return err
	}

	// If we're using fallback but primary is back, migrate
	if m.useFallback && store == m.fallbackStore {
		go m.checkAndMigrateBack(ctx)
	}

	return nil
}

// Get retrieves a session
func (m *SessionManager) Get(ctx context.Context, sessionID string) (*models.Session, error) {
	store := m.getActiveStore()

	session, err := store.Get(ctx, sessionID)
	if err != nil {
		// If primary fails, try fallback
		if store == m.primaryStore && m.fallbackStore != nil {
			m.logger.Warn("Primary store failed, using fallback",
				zap.Error(err))
			m.useFallback = true
			return m.fallbackStore.Get(ctx, sessionID)
		}
		return nil, err
	}

	// If session not found in primary, check fallback (migration scenario)
	if session == nil && store == m.primaryStore && m.fallbackStore != nil {
		session, err = m.fallbackStore.Get(ctx, sessionID)
		if err != nil {
			return nil, err
		}
		if session != nil {
			// Migrate session to primary
			go func() {
				if err := m.primaryStore.Create(context.Background(), session); err != nil {
					m.logger.Warn("Failed to migrate session to primary store",
						zap.String("session_id", sessionID),
						zap.Error(err))
				}
			}()
		}
	}

	return session, nil
}

// Update updates a session (extends TTL)
func (m *SessionManager) Update(ctx context.Context, sessionID string) error {
	store := m.getActiveStore()

	if err := store.Update(ctx, sessionID); err != nil {
		// If primary fails, try fallback
		if store == m.primaryStore && m.fallbackStore != nil {
			m.logger.Warn("Primary store failed, using fallback",
				zap.Error(err))
			m.useFallback = true
			return m.fallbackStore.Update(ctx, sessionID)
		}
		return err
	}

	return nil
}

// Delete removes a session
func (m *SessionManager) Delete(ctx context.Context, sessionID string) error {
	// Delete from both stores to ensure consistency
	var primaryErr, fallbackErr error

	if m.primaryStore != nil {
		primaryErr = m.primaryStore.Delete(ctx, sessionID)
		if primaryErr != nil {
			m.logger.Warn("Failed to delete from primary store",
				zap.String("session_id", sessionID),
				zap.Error(primaryErr))
		}
	}

	if m.fallbackStore != nil {
		fallbackErr = m.fallbackStore.Delete(ctx, sessionID)
		if fallbackErr != nil {
			m.logger.Warn("Failed to delete from fallback store",
				zap.String("session_id", sessionID),
				zap.Error(fallbackErr))
		}
	}

	// Return error only if both failed
	if primaryErr != nil && fallbackErr != nil {
		return fmt.Errorf("failed to delete from both stores: primary=%v, fallback=%v",
			primaryErr, fallbackErr)
	}

	return nil
}

// GetUserSessions gets all sessions for a user
func (m *SessionManager) GetUserSessions(ctx context.Context, userID uint) ([]*models.Session, error) {
	store := m.getActiveStore()

	sessions, err := store.GetUserSessions(ctx, userID)
	if err != nil {
		// If primary fails, try fallback
		if store == m.primaryStore && m.fallbackStore != nil {
			m.logger.Warn("Primary store failed, using fallback",
				zap.Error(err))
			m.useFallback = true
			return m.fallbackStore.GetUserSessions(ctx, userID)
		}
		return nil, err
	}

	return sessions, nil
}

// DeleteUserSessions deletes all sessions for a user
func (m *SessionManager) DeleteUserSessions(ctx context.Context, userID uint) error {
	// Delete from both stores to ensure consistency
	var primaryErr, fallbackErr error

	if m.primaryStore != nil {
		primaryErr = m.primaryStore.DeleteUserSessions(ctx, userID)
		if primaryErr != nil {
			m.logger.Warn("Failed to delete user sessions from primary store",
				zap.Uint("user_id", userID),
				zap.Error(primaryErr))
		}
	}

	if m.fallbackStore != nil {
		fallbackErr = m.fallbackStore.DeleteUserSessions(ctx, userID)
		if fallbackErr != nil {
			m.logger.Warn("Failed to delete user sessions from fallback store",
				zap.Uint("user_id", userID),
				zap.Error(fallbackErr))
		}
	}

	// Return error only if both failed
	if primaryErr != nil && fallbackErr != nil {
		return fmt.Errorf("failed to delete from both stores: primary=%v, fallback=%v",
			primaryErr, fallbackErr)
	}

	return nil
}

// getActiveStore returns the currently active store
func (m *SessionManager) getActiveStore() SessionStore {
	// Periodically check if primary is back
	if m.useFallback && time.Since(m.lastHealthCheck) > m.healthCheckTTL {
		m.checkPrimaryHealth()
	}

	if m.useFallback && m.fallbackStore != nil {
		return m.fallbackStore
	}
	return m.primaryStore
}

// checkPrimaryHealth checks if the primary store is healthy again
func (m *SessionManager) checkPrimaryHealth() {
	m.lastHealthCheck = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.primaryStore.HealthCheck(ctx); err == nil {
		m.logger.Info("Primary store is healthy again, switching back")
		m.useFallback = false
	}
}

// checkAndMigrateBack checks if primary is back and starts migration
func (m *SessionManager) checkAndMigrateBack(ctx context.Context) {
	if !m.useFallback {
		return
	}

	// Check primary health
	if err := m.primaryStore.HealthCheck(ctx); err != nil {
		return // Primary still unhealthy
	}

	m.logger.Info("Primary store recovered, switching back")
	m.useFallback = false

	// TODO: Implement full migration of sessions from fallback to primary
	// This would involve iterating through all sessions in fallback and copying them
	// For now, sessions will be migrated on-demand as they're accessed
}

// HealthCheck performs health check on active store
func (m *SessionManager) HealthCheck(ctx context.Context) error {
	store := m.getActiveStore()
	return store.HealthCheck(ctx)
}

// CleanupExpiredSessions cleans up expired sessions
func (m *SessionManager) CleanupExpiredSessions(ctx context.Context) error {
	// Clean up both stores
	var primaryErr, fallbackErr error

	if m.primaryStore != nil {
		primaryErr = m.primaryStore.CleanupExpiredSessions(ctx)
		if primaryErr != nil {
			m.logger.Warn("Failed to cleanup primary store",
				zap.Error(primaryErr))
		}
	}

	if m.fallbackStore != nil {
		fallbackErr = m.fallbackStore.CleanupExpiredSessions(ctx)
		if fallbackErr != nil {
			m.logger.Warn("Failed to cleanup fallback store",
				zap.Error(fallbackErr))
		}
	}

	if primaryErr != nil && fallbackErr != nil {
		return fmt.Errorf("cleanup failed in both stores: primary=%v, fallback=%v",
			primaryErr, fallbackErr)
	}

	return nil
}