package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimiter interface for rate limiting implementations
type RateLimiter interface {
	Allow(key string) bool
	AllowN(key string, n int) bool
	Close() error
}

// InMemoryRateLimiter provides in-memory rate limiting
type InMemoryRateLimiter struct {
	config   RateLimitConfig
	logger   *zap.Logger
	buckets  map[string]*bucket
	mutex    sync.RWMutex
	ticker   *time.Ticker
	stopChan chan struct{}
}

type bucket struct {
	tokens     int
	lastUpdate time.Time
	mutex      sync.Mutex
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter
func NewInMemoryRateLimiter(config RateLimitConfig, logger *zap.Logger) *InMemoryRateLimiter {
	rl := &InMemoryRateLimiter{
		config:   config,
		logger:   logger,
		buckets:  make(map[string]*bucket),
		ticker:   time.NewTicker(config.CleanupInterval),
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

func (rl *InMemoryRateLimiter) Allow(key string) bool {
	return rl.AllowN(key, 1)
}

func (rl *InMemoryRateLimiter) AllowN(key string, n int) bool {
	rl.mutex.Lock()
	b, exists := rl.buckets[key]
	if !exists {
		b = &bucket{
			tokens:     rl.config.Burst,
			lastUpdate: time.Now(),
		}
		rl.buckets[key] = b
	}
	rl.mutex.Unlock()

	b.mutex.Lock()
	defer b.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastUpdate)
	tokensToAdd := int(elapsed.Minutes()) * rl.config.RequestsPerMinute

	// Refill tokens
	b.tokens += tokensToAdd
	if b.tokens > rl.config.Burst {
		b.tokens = rl.config.Burst
	}
	b.lastUpdate = now

	// Check if we have enough tokens
	if b.tokens >= n {
		b.tokens -= n
		return true
	}

	return false
}

func (rl *InMemoryRateLimiter) cleanup() {
	for {
		select {
		case <-rl.ticker.C:
			rl.mutex.Lock()
			now := time.Now()
			for key, bucket := range rl.buckets {
				bucket.mutex.Lock()
				if now.Sub(bucket.lastUpdate) > time.Hour {
					delete(rl.buckets, key)
				}
				bucket.mutex.Unlock()
			}
			rl.mutex.Unlock()
		case <-rl.stopChan:
			return
		}
	}
}

func (rl *InMemoryRateLimiter) Close() error {
	close(rl.stopChan)
	rl.ticker.Stop()
	return nil
}

// RedisRateLimiter provides Redis-based rate limiting
type RedisRateLimiter struct {
	config RateLimitConfig
	logger *zap.Logger
	client *redis.Client
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(redisConfig RedisConfig, rateLimitConfig RateLimitConfig, logger *zap.Logger) *RedisRateLimiter {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password:     redisConfig.Password,
		DB:           redisConfig.DB,
		MaxRetries:   redisConfig.MaxRetries,
		PoolSize:     redisConfig.PoolSize,
		MinIdleConns: redisConfig.MinIdleConns,
	})

	return &RedisRateLimiter{
		config: rateLimitConfig,
		logger: logger,
		client: client,
	}
}

func (rl *RedisRateLimiter) Allow(key string) bool {
	return rl.AllowN(key, 1)
}

func (rl *RedisRateLimiter) AllowN(key string, n int) bool {
	ctx := context.Background()
	fullKey := rl.config.RedisKeyPrefix + key

	// Use Redis sliding window approach
	now := time.Now()
	windowStart := now.Add(-time.Minute)

	pipe := rl.client.TxPipeline()

	// Remove old entries
	pipe.ZRemRangeByScore(ctx, fullKey, "0", fmt.Sprintf("%d", windowStart.UnixNano()))

	// Count current requests
	countCmd := pipe.ZCard(ctx, fullKey)

	// Add current request
	for i := 0; i < n; i++ {
		pipe.ZAdd(ctx, fullKey, redis.Z{
			Score:  float64(now.Add(time.Duration(i) * time.Nanosecond).UnixNano()),
			Member: fmt.Sprintf("%d-%d", now.UnixNano(), i),
		})
	}

	// Set expiration
	pipe.Expire(ctx, fullKey, time.Minute)

	_, err := pipe.Exec(ctx)
	if err != nil {
		rl.logger.Error("Redis rate limiter error", zap.Error(err))
		return false
	}

	// Check if we're under the limit
	count := countCmd.Val()
	return int(count) <= rl.config.RequestsPerMinute
}

func (rl *RedisRateLimiter) Close() error {
	return rl.client.Close()
}

// RateLimitingMiddleware creates Gin middleware for rate limiting
func RateLimitingMiddleware(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting for excluded paths
		for _, path := range config.ExcludedPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// For now, just continue - would need RateLimiter instance
		// In practice, this would be injected or retrieved from context
		c.Next()
	}
}