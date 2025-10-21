package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// RateLimitStrategy defines different rate limiting strategies
type RateLimitStrategy string

const (
	// PerIP rate limits per IP address
	PerIP RateLimitStrategy = "per_ip"
	// PerUser rate limits per authenticated user
	PerUser RateLimitStrategy = "per_user"
	// PerTenant rate limits per tenant
	PerTenant RateLimitStrategy = "per_tenant"
	// Global applies a global rate limit
	Global RateLimitStrategy = "global"
)

// RateLimitConfig configures rate limiting
type RateLimitConfig struct {
	Strategy        RateLimitStrategy
	RequestsPerMin  int
	BurstSize       int
	CleanupInterval time.Duration
	WhitelistedIPs  []string
	ExcludedPaths   []string
}

// RateLimiter implements various rate limiting strategies
type RateLimiter struct {
	config    RateLimitConfig
	limiters  map[string]*rateLimiterEntry
	mu        sync.RWMutex
	logger    *zap.Logger
	stopClean chan struct{}
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimitConfig, logger *zap.Logger) *RateLimiter {
	rl := &RateLimiter{
		config:    config,
		limiters:  make(map[string]*rateLimiterEntry),
		logger:    logger,
		stopClean: make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Middleware returns the rate limiting middleware
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if path is excluded
		if rl.isPathExcluded(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Check if IP is whitelisted
		clientIP := c.ClientIP()
		if rl.isIPWhitelisted(clientIP) {
			c.Next()
			return
		}

		// Get rate limit key based on strategy
		key := rl.getKey(c)
		if key == "" {
			c.Next()
			return
		}

		// Get or create limiter for this key
		limiter := rl.getLimiter(key)

		// Check if request is allowed
		if !limiter.Allow() {
			rl.logger.Warn("Rate limit exceeded",
				zap.String("key", key),
				zap.String("strategy", string(rl.config.Strategy)),
				zap.String("ip", clientIP))

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.config.RequestsPerMin))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
			c.Header("Retry-After", "60")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"message": "too many requests, please try again later",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.config.RequestsPerMin))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", int(limiter.Tokens())))

		c.Next()
	}
}

// AdvancedMiddleware provides more sophisticated rate limiting with different limits per endpoint
func (rl *RateLimiter) AdvancedMiddleware(limits map[string]RateLimitConfig) gin.HandlerFunc {
	limiters := make(map[string]*RateLimiter)

	// Create separate rate limiters for each endpoint pattern
	for pattern, config := range limits {
		limiters[pattern] = NewRateLimiter(config, rl.logger)
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Find matching pattern
		for pattern, limiter := range limiters {
			if matched := matchPath(pattern, path); matched {
				limiter.Middleware()(c)
				return
			}
		}

		// Use default limiter if no pattern matches
		rl.Middleware()(c)
	}
}

// getKey returns the rate limit key based on strategy
func (rl *RateLimiter) getKey(c *gin.Context) string {
	switch rl.config.Strategy {
	case PerIP:
		return "ip:" + c.ClientIP()
	case PerUser:
		userID := c.GetString("user_id")
		if userID == "" {
			// Fall back to IP if user not authenticated
			return "ip:" + c.ClientIP()
		}
		return "user:" + userID
	case PerTenant:
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			return ""
		}
		return "tenant:" + tenantID
	case Global:
		return "global"
	default:
		return "ip:" + c.ClientIP()
	}
}

// getLimiter returns the rate limiter for a key
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.RLock()
	entry, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if exists {
		// Update last seen
		rl.mu.Lock()
		entry.lastSeen = time.Now()
		rl.mu.Unlock()
		return entry.limiter
	}

	// Create new limiter
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if entry, exists := rl.limiters[key]; exists {
		entry.lastSeen = time.Now()
		return entry.limiter
	}

	// Create new limiter
	limiter := rate.NewLimiter(
		rate.Every(time.Minute/time.Duration(rl.config.RequestsPerMin)),
		rl.config.BurstSize,
	)

	rl.limiters[key] = &rateLimiterEntry{
		limiter:  limiter,
		lastSeen: time.Now(),
	}

	return limiter
}

// cleanup removes old rate limiters
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanupOldEntries()
		case <-rl.stopClean:
			return
		}
	}
}

// cleanupOldEntries removes entries not seen for cleanup interval
func (rl *RateLimiter) cleanupOldEntries() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, entry := range rl.limiters {
		if now.Sub(entry.lastSeen) > rl.config.CleanupInterval {
			delete(rl.limiters, key)
		}
	}

	rl.logger.Debug("Cleaned up rate limiters",
		zap.Int("remaining", len(rl.limiters)))
}

// isPathExcluded checks if path should be excluded from rate limiting
func (rl *RateLimiter) isPathExcluded(path string) bool {
	for _, excluded := range rl.config.ExcludedPaths {
		if strings.HasPrefix(path, excluded) {
			return true
		}
	}
	return false
}

// isIPWhitelisted checks if IP is whitelisted
func (rl *RateLimiter) isIPWhitelisted(ip string) bool {
	for _, whitelisted := range rl.config.WhitelistedIPs {
		if ip == whitelisted {
			return true
		}
	}
	return false
}

// Stop stops the cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stopClean)
}

// matchPath checks if a path matches a pattern (simple glob matching)
func matchPath(pattern, path string) bool {
	if pattern == path {
		return true
	}

	// Simple wildcard matching
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}

	return false
}

// DynamicRateLimiter adjusts rate limits based on system load
type DynamicRateLimiter struct {
	*RateLimiter
	minRate     int
	maxRate     int
	loadChecker func() float64 // Returns system load 0.0-1.0
}

// NewDynamicRateLimiter creates a rate limiter that adjusts based on system load
func NewDynamicRateLimiter(config RateLimitConfig, logger *zap.Logger, loadChecker func() float64) *DynamicRateLimiter {
	drl := &DynamicRateLimiter{
		RateLimiter: NewRateLimiter(config, logger),
		minRate:     config.RequestsPerMin / 4,
		maxRate:     config.RequestsPerMin,
		loadChecker: loadChecker,
	}

	// Start load adjustment goroutine
	go drl.adjustRateLimits()

	return drl
}

// adjustRateLimits periodically adjusts rate limits based on load
func (drl *DynamicRateLimiter) adjustRateLimits() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		load := drl.loadChecker()

		// Calculate new rate based on load (inverse relationship)
		// High load = lower rate limit
		newRate := drl.maxRate - int(float64(drl.maxRate-drl.minRate)*load)

		if newRate != drl.config.RequestsPerMin {
			drl.logger.Info("Adjusting rate limit based on load",
				zap.Float64("load", load),
				zap.Int("old_rate", drl.config.RequestsPerMin),
				zap.Int("new_rate", newRate))

			drl.config.RequestsPerMin = newRate
		}
	}
}

// CreateDefaultRateLimitConfigs creates default rate limit configurations
func CreateDefaultRateLimitConfigs() map[string]RateLimitConfig {
	return map[string]RateLimitConfig{
		// Public endpoints - more restrictive
		"/api/v1/public/*": {
			Strategy:        PerIP,
			RequestsPerMin:  30,
			BurstSize:       5,
			CleanupInterval: 5 * time.Minute,
		},
		// Auth endpoints - very restrictive
		"/api/v1/auth/login": {
			Strategy:        PerIP,
			RequestsPerMin:  5,
			BurstSize:       2,
			CleanupInterval: 5 * time.Minute,
		},
		// Admin endpoints - per user
		"/api/v1/admin/*": {
			Strategy:        PerUser,
			RequestsPerMin:  100,
			BurstSize:       20,
			CleanupInterval: 10 * time.Minute,
		},
		// Tenant operations - per tenant
		"/api/v1/tenants/*": {
			Strategy:        PerTenant,
			RequestsPerMin:  200,
			BurstSize:       50,
			CleanupInterval: 10 * time.Minute,
		},
	}
}