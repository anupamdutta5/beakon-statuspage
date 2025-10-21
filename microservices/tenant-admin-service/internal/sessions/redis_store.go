package sessions

import (
	"github.com/google/uuid"

	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anupamdutta5/tenant-admin-service/internal/models"
	"github.com/anupamdutta5/shared-resilience"
	"go.uber.org/zap"
)

// RedisSessionStore implements session storage using Redis
type RedisSessionStore struct {
	cache  resilience.Cache
	logger *zap.Logger
	ttl    time.Duration
}

// NewRedisSessionStore creates a new Redis-based session store
func NewRedisSessionStore(redisConfig resilience.RedisConfig, logger *zap.Logger) (*RedisSessionStore, error) {
	// Test Redis connection first before creating store
	testClient := resilience.NewRedisCache(resilience.RedisConfig{
		Host: redisConfig.Host,
		Port: redisConfig.Port,
		Password: redisConfig.Password,
		DB: redisConfig.DB,
	}, resilience.CacheConfig{
		Enabled:         true,
		Type:            "redis",
		DefaultTTL:      24 * time.Hour,
		CleanupInterval: 5 * time.Minute,
	}, logger)

	// Test if Redis is actually accessible
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	testKey := "redis:health:check"
	if err := testClient.Set(ctx, testKey, "ok", time.Second); err != nil {
		logger.Warn("Redis connection test failed", zap.Error(err))
		return nil, fmt.Errorf("redis not accessible: %w", err)
	}
	_ = testClient.Delete(ctx, testKey)

	// Create a cache config for the Redis cache
	cacheConfig := resilience.CacheConfig{
		Enabled:         true,
		Type:            "redis",
		DefaultTTL:      24 * time.Hour,
		CleanupInterval: 5 * time.Minute,
	}

	cache := resilience.NewRedisCache(redisConfig, cacheConfig, logger)

	return &RedisSessionStore{
		cache:  cache,
		logger: logger,
		ttl:    24 * time.Hour, // Default session TTL
	}, nil
}

// Create creates a new session in Redis
func (s *RedisSessionStore) Create(ctx context.Context, session *models.Session) error {
	key := s.getSessionKey(session.ID)

	// Serialize session to JSON
	data, err := json.Marshal(session)
	if err != nil {
		s.logger.Error("Failed to marshal session", zap.Error(err))
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Store in Redis with TTL
	if err := s.cache.Set(ctx, key, string(data), s.ttl); err != nil {
		s.logger.Error("Failed to store session in Redis",
			zap.String("session_id", session.ID),
			zap.Error(err))
		return fmt.Errorf("failed to store session: %w", err)
	}

	// Also maintain a user-to-sessions index for easy lookup
	// Store user sessions as a JSON list since SetAdd is not available
	userKey := s.getUserSessionsKey(session.UserID)
	existingSessions := s.getUserSessionIDs(ctx, session.UserID)

	// Add new session ID if not already present
	found := false
	for _, id := range existingSessions {
		if id == session.ID {
			found = true
			break
		}
	}
	if !found {
		existingSessions = append(existingSessions, session.ID)
		if err := s.cache.Set(ctx, userKey, existingSessions, s.ttl); err != nil {
			s.logger.Warn("Failed to update user session index",
				zap.String("session_id", session.ID),
				zap.String("user_id", session.UserID.String()),
				zap.Error(err))
		}
	}

	s.logger.Debug("Session created in Redis",
		zap.String("session_id", session.ID),
		zap.String("user_id", session.UserID.String()))

	return nil
}

// Get retrieves a session from Redis
func (s *RedisSessionStore) Get(ctx context.Context, sessionID string) (*models.Session, error) {
	key := s.getSessionKey(sessionID)

	// Get from Redis
	data, found := s.cache.Get(ctx, key)
	if !found {
		return nil, nil // Session doesn't exist
	}

	// Convert data to string first
	var jsonStr string
	switch v := data.(type) {
	case string:
		jsonStr = v
	case []byte:
		jsonStr = string(v)
	default:
		// Try to marshal and unmarshal to handle other types
		bytes, err := json.Marshal(data)
		if err != nil {
			s.logger.Error("Failed to convert cached data",
				zap.String("session_id", sessionID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to convert cached data: %w", err)
		}
		jsonStr = string(bytes)
	}

	// Deserialize session
	var session models.Session
	if err := json.Unmarshal([]byte(jsonStr), &session); err != nil {
		s.logger.Error("Failed to unmarshal session",
			zap.String("session_id", sessionID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// Update updates a session in Redis (extends TTL)
func (s *RedisSessionStore) Update(ctx context.Context, sessionID string) error {
	key := s.getSessionKey(sessionID)

	// Get existing session
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session not found")
	}

	// Update last seen time
	session.LastSeenAt = time.Now()

	// Re-save with extended TTL
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := s.cache.Set(ctx, key, string(data), s.ttl); err != nil {
		s.logger.Error("Failed to update session in Redis",
			zap.String("session_id", sessionID),
			zap.Error(err))
		return fmt.Errorf("failed to update session: %w", err)
	}

	s.logger.Debug("Session updated in Redis",
		zap.String("session_id", sessionID))

	return nil
}

// Delete removes a session from Redis
func (s *RedisSessionStore) Delete(ctx context.Context, sessionID string) error {
	// Get session first to find user ID for index cleanup
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		s.logger.Warn("Failed to get session for deletion",
			zap.String("session_id", sessionID),
			zap.Error(err))
	}

	// Delete session
	key := s.getSessionKey(sessionID)
	if err := s.cache.Delete(ctx, key); err != nil {
		s.logger.Error("Failed to delete session from Redis",
			zap.String("session_id", sessionID),
			zap.Error(err))
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Remove from user index if we have the session data
	if session != nil {
		userKey := s.getUserSessionsKey(session.UserID)
		existingSessions := s.getUserSessionIDs(ctx, session.UserID)

		// Remove session ID from the list
		newSessions := make([]string, 0, len(existingSessions))
		for _, id := range existingSessions {
			if id != sessionID {
				newSessions = append(newSessions, id)
			}
		}

		if len(newSessions) > 0 {
			if err := s.cache.Set(ctx, userKey, newSessions, s.ttl); err != nil {
				s.logger.Warn("Failed to update user session index",
					zap.String("session_id", sessionID),
					zap.String("user_id", session.UserID.String()),
					zap.Error(err))
			}
		} else {
			// Delete the key if no sessions remain
			if err := s.cache.Delete(ctx, userKey); err != nil {
				s.logger.Warn("Failed to delete user session index",
					zap.String("user_id", session.UserID.String()),
					zap.Error(err))
			}
		}
	}

	s.logger.Debug("Session deleted from Redis",
		zap.String("session_id", sessionID))

	return nil
}

// GetUserSessions gets all sessions for a user
func (s *RedisSessionStore) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	// Get all session IDs for user
	sessionIDs := s.getUserSessionIDs(ctx, userID)

	// Get each session
	sessions := make([]*models.Session, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		session, err := s.Get(ctx, sessionID)
		if err != nil {
			s.logger.Warn("Failed to get session",
				zap.String("session_id", sessionID),
				zap.Error(err))
			continue
		}
		if session != nil {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

// DeleteUserSessions deletes all sessions for a user
func (s *RedisSessionStore) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	sessions, err := s.GetUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	for _, session := range sessions {
		if err := s.Delete(ctx, session.ID); err != nil {
			s.logger.Warn("Failed to delete user session",
				zap.String("session_id", session.ID),
				zap.String("user_id", userID.String()),
				zap.Error(err))
		}
	}

	// Clean up user index
	userKey := s.getUserSessionsKey(userID)
	if err := s.cache.Delete(ctx, userKey); err != nil {
		s.logger.Warn("Failed to delete user sessions index",
			zap.String("user_id", userID.String()),
			zap.Error(err))
	}

	s.logger.Info("Deleted all user sessions",
		zap.String("user_id", userID.String()),
		zap.Int("count", len(sessions)))

	return nil
}

// CleanupExpiredSessions removes expired sessions (called periodically)
func (s *RedisSessionStore) CleanupExpiredSessions(ctx context.Context) error {
	// Redis automatically expires keys with TTL, so this is mainly for logging
	s.logger.Debug("Redis handles session expiry automatically via TTL")
	return nil
}

// getSessionKey returns the Redis key for a session
func (s *RedisSessionStore) getSessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

// getUserSessionsKey returns the Redis key for user's sessions set
func (s *RedisSessionStore) getUserSessionsKey(userID uuid.UUID) string {
	return fmt.Sprintf("user_sessions:%s", userID.String())
}

// SetTTL updates the default TTL for new sessions
func (s *RedisSessionStore) SetTTL(ttl time.Duration) {
	s.ttl = ttl
}

// HealthCheck checks if Redis is accessible
func (s *RedisSessionStore) HealthCheck(ctx context.Context) error {
	// Try to ping Redis
	testKey := "health:check"
	if err := s.cache.Set(ctx, testKey, "ok", time.Second); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}
	_ = s.cache.Delete(ctx, testKey)
	return nil
}

// getUserSessionIDs retrieves the list of session IDs for a user
func (s *RedisSessionStore) getUserSessionIDs(ctx context.Context, userID uuid.UUID) []string {
	userKey := s.getUserSessionsKey(userID)

	// Get session IDs from cache
	data, found := s.cache.Get(ctx, userKey)
	if !found {
		return []string{}
	}

	// Convert to string array
	var sessionIDs []string
	switch v := data.(type) {
	case []string:
		sessionIDs = v
	case []interface{}:
		sessionIDs = make([]string, 0, len(v))
		for _, id := range v {
			if strID, ok := id.(string); ok {
				sessionIDs = append(sessionIDs, strID)
			}
		}
	default:
		// Try to unmarshal from JSON
		if jsonStr, ok := data.(string); ok {
			_ = json.Unmarshal([]byte(jsonStr), &sessionIDs)
		}
	}

	// Clean up expired sessions
	validSessionIDs := make([]string, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		// Check if session still exists
		key := s.getSessionKey(sessionID)
		if _, found := s.cache.Get(ctx, key); found {
			validSessionIDs = append(validSessionIDs, sessionID)
		}
	}

	// Update the index if we removed any expired sessions
	if len(validSessionIDs) != len(sessionIDs) {
		if len(validSessionIDs) > 0 {
			_ = s.cache.Set(ctx, userKey, validSessionIDs, s.ttl)
		} else {
			_ = s.cache.Delete(ctx, userKey)
		}
	}

	return validSessionIDs
}