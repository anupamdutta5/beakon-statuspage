package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisClient wraps redis client with health checking and graceful fallback
type RedisClient struct {
	client    *redis.Client
	logger    *zap.Logger
	enabled   bool
	healthTTL time.Duration
	lastCheck time.Time
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	Enabled  bool
}

// NewRedisClient creates a new Redis client with health checking
func NewRedisClient(config RedisConfig, logger *zap.Logger) *RedisClient {
	if !config.Enabled {
		logger.Info("Redis disabled - using database-only mode")
		return &RedisClient{
			client:  nil,
			logger:  logger,
			enabled: false,
		}
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     config.Password,
		DB:           config.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	rc := &RedisClient{
		client:    client,
		logger:    logger,
		enabled:   true,
		healthTTL: 30 * time.Second,
		lastCheck: time.Time{},
	}

	// Initial health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rc.Ping(ctx); err != nil {
		logger.Warn("Redis initial connection failed - will retry",
			zap.String("addr", addr),
			zap.Error(err))
		rc.enabled = false
	} else {
		logger.Info("Redis connected successfully", zap.String("addr", addr))
	}

	return rc
}

// Ping checks Redis connectivity
func (r *RedisClient) Ping(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	return r.client.Ping(ctx).Err()
}

// IsHealthy returns true if Redis is currently accessible
func (r *RedisClient) IsHealthy(ctx context.Context) bool {
	if !r.enabled || r.client == nil {
		return false
	}

	// Use cached health status if recent
	if time.Since(r.lastCheck) < r.healthTTL {
		return r.enabled
	}

	// Perform health check
	if err := r.Ping(ctx); err != nil {
		r.logger.Warn("Redis health check failed", zap.Error(err))
		r.enabled = false
		r.lastCheck = time.Now()
		return false
	}

	r.enabled = true
	r.lastCheck = time.Now()
	return true
}

// Set stores a key-value pair with expiration
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if !r.IsHealthy(ctx) {
		return fmt.Errorf("redis not available")
	}

	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	if !r.IsHealthy(ctx) {
		return "", fmt.Errorf("redis not available")
	}

	return r.client.Get(ctx, key).Result()
}

// Del deletes a key
func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	if !r.IsHealthy(ctx) {
		return fmt.Errorf("redis not available")
	}

	return r.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func (r *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	if !r.IsHealthy(ctx) {
		return 0, fmt.Errorf("redis not available")
	}

	return r.client.Exists(ctx, keys...).Result()
}

// Expire sets expiration on a key
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if !r.IsHealthy(ctx) {
		return fmt.Errorf("redis not available")
	}

	return r.client.Expire(ctx, key, expiration).Err()
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	if r.client == nil {
		return nil
	}

	return r.client.Close()
}

// GetClient returns the underlying Redis client (use with caution)
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}
