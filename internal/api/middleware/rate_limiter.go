package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	config *config.RateLimitConfig
	store  RateLimitStore
}

// RateLimitStore defines the interface for rate limit storage
type RateLimitStore interface {
	Get(ctx context.Context, key string) (*RateLimitEntry, error)
	Set(ctx context.Context, key string, entry *RateLimitEntry, ttl time.Duration) error
	Increment(ctx context.Context, key string, window time.Duration) (*RateLimitEntry, error)
}

// RateLimitEntry represents a rate limit entry
type RateLimitEntry struct {
	Count     int       `json:"count"`
	Window    time.Time `json:"window"`
	ResetTime time.Time `json:"reset_time"`
}

// InMemoryStore provides in-memory rate limit storage
type InMemoryStore struct {
	mu    sync.RWMutex
	store map[string]*RateLimitEntry
}

// NewInMemoryStore creates a new in-memory store
func NewInMemoryStore() *InMemoryStore {
	store := &InMemoryStore{
		store: make(map[string]*RateLimitEntry),
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

// Get retrieves a rate limit entry
func (s *InMemoryStore) Get(ctx context.Context, key string) (*RateLimitEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.store[key]
	if !exists {
		return nil, nil
	}

	// Check if entry has expired
	if time.Now().After(entry.ResetTime) {
		return nil, nil
	}

	return entry, nil
}

// Set stores a rate limit entry
func (s *InMemoryStore) Set(ctx context.Context, key string, entry *RateLimitEntry, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.store[key] = entry
	return nil
}

// Increment increments the count for a key
func (s *InMemoryStore) Increment(ctx context.Context, key string, window time.Duration) (*RateLimitEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	entry, exists := s.store[key]

	if !exists || now.After(entry.ResetTime) {
		// Create new entry
		entry = &RateLimitEntry{
			Count:     1,
			Window:    now,
			ResetTime: now.Add(window),
		}
	} else {
		// Increment existing entry
		entry.Count++
	}

	s.store[key] = entry
	return entry, nil
}

// cleanup removes expired entries
func (s *InMemoryStore) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, entry := range s.store {
			if now.After(entry.ResetTime) {
				delete(s.store, key)
			}
		}
		s.mu.Unlock()
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *config.RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		config: config,
		store:  NewInMemoryStore(),
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.config.Enabled {
			c.Next()
			return
		}

		// Get client identifier
		clientID := rl.getClientID(c)

		// Create rate limit key
		key := fmt.Sprintf("rate_limit:%s:%s", c.Request.Method, clientID)

		// Check rate limit
		entry, err := rl.store.Increment(c.Request.Context(), key, rl.config.Window)
		if err != nil {
			logger.Error("Rate limit store error", zap.Error(err))
			c.Next()
			return
		}

		// Check if limit exceeded
		if entry.Count > rl.config.Rate {
			// Add rate limit headers
			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.Rate))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(entry.ResetTime.Unix(), 10))
			c.Header("Retry-After", strconv.FormatInt(int64(time.Until(entry.ResetTime).Seconds()), 10))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "TOO_MANY_REQUESTS",
				"details": map[string]interface{}{
					"limit":       rl.config.Rate,
					"remaining":   0,
					"reset_at":    entry.ResetTime,
					"retry_after": int(time.Until(entry.ResetTime).Seconds()),
				},
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		remaining := rl.config.Rate - entry.Count
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.Rate))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(entry.ResetTime.Unix(), 10))

		c.Next()
	}
}

// TenantRateLimitMiddleware creates a tenant-specific rate limiting middleware
func (rl *RateLimiter) TenantRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.config.Enabled {
			c.Next()
			return
		}

		// Get tenant from context
		tenant, exists := GetTenantFromContext(c)
		if !exists {
			c.Next()
			return
		}

		// Get client identifier
		clientID := rl.getClientID(c)

		// Create tenant-specific rate limit key
		key := fmt.Sprintf("tenant_rate_limit:%d:%s:%s", tenant.ID, c.Request.Method, clientID)

		// Check rate limit
		entry, err := rl.store.Increment(c.Request.Context(), key, rl.config.Window)
		if err != nil {
			logger.Error("Tenant rate limit store error", zap.Error(err))
			c.Next()
			return
		}

		// Check if limit exceeded
		if entry.Count > rl.config.Rate {
			// Add rate limit headers
			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.Rate))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(entry.ResetTime.Unix(), 10))
			c.Header("Retry-After", strconv.FormatInt(int64(time.Until(entry.ResetTime).Seconds()), 10))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Tenant rate limit exceeded",
				"code":  "TOO_MANY_REQUESTS",
				"details": map[string]interface{}{
					"tenant_id":   tenant.ID,
					"limit":       rl.config.Rate,
					"remaining":   0,
					"reset_at":    entry.ResetTime,
					"retry_after": int(time.Until(entry.ResetTime).Seconds()),
				},
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		remaining := rl.config.Rate - entry.Count
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.Rate))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(entry.ResetTime.Unix(), 10))

		c.Next()
	}
}

// APIKeyRateLimitMiddleware creates an API key specific rate limiting middleware
func (rl *RateLimiter) APIKeyRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.config.Enabled {
			c.Next()
			return
		}

		// Get API key from header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.Next()
			return
		}

		// Create API key specific rate limit key
		key := fmt.Sprintf("api_rate_limit:%s:%s", c.Request.Method, apiKey)

		// Check rate limit
		entry, err := rl.store.Increment(c.Request.Context(), key, rl.config.Window)
		if err != nil {
			logger.Error("API key rate limit store error", zap.Error(err))
			c.Next()
			return
		}

		// Check if limit exceeded
		if entry.Count > rl.config.Rate {
			// Add rate limit headers
			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.Rate))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(entry.ResetTime.Unix(), 10))
			c.Header("Retry-After", strconv.FormatInt(int64(time.Until(entry.ResetTime).Seconds()), 10))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "API key rate limit exceeded",
				"code":  "TOO_MANY_REQUESTS",
				"details": map[string]interface{}{
					"api_key":     apiKey[:8] + "...", // Mask API key
					"limit":       rl.config.Rate,
					"remaining":   0,
					"reset_at":    entry.ResetTime,
					"retry_after": int(time.Until(entry.ResetTime).Seconds()),
				},
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		remaining := rl.config.Rate - entry.Count
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.Rate))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(entry.ResetTime.Unix(), 10))

		c.Next()
	}
}

// getClientID gets a unique identifier for the client
func (rl *RateLimiter) getClientID(c *gin.Context) string {
	// Try to get user ID from context first
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%v", userID)
	}

	// Try to get tenant ID from context
	if tenant, exists := GetTenantFromContext(c); exists {
		return fmt.Sprintf("tenant:%d", tenant.ID)
	}

	// Fall back to IP address
	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// BurstRateLimitMiddleware provides burst protection
func (rl *RateLimiter) BurstRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.config.Enabled {
			c.Next()
			return
		}

		// Get client identifier
		clientID := rl.getClientID(c)

		// Create burst rate limit key (shorter window)
		burstWindow := time.Minute // 1 minute burst window
		key := fmt.Sprintf("burst_rate_limit:%s", clientID)

		// Check burst rate limit
		entry, err := rl.store.Increment(c.Request.Context(), key, burstWindow)
		if err != nil {
			logger.Error("Burst rate limit store error", zap.Error(err))
			c.Next()
			return
		}

		// Check if burst limit exceeded
		if entry.Count > rl.config.Burst {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Burst rate limit exceeded",
				"code":  "TOO_MANY_REQUESTS",
				"details": map[string]interface{}{
					"burst_limit": rl.config.Burst,
					"count":       entry.Count,
					"reset_at":    entry.ResetTime,
					"retry_after": int(time.Until(entry.ResetTime).Seconds()),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
