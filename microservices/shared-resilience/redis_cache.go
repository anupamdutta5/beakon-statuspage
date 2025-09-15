// Package resilience provides Redis-based caching for microservices
package resilience

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisCacheConfig represents configuration for Redis cache
type RedisCacheConfig struct {
	Host         string        `yaml:"host" default:"localhost"`
	Port         int           `yaml:"port" default:"6379"`
	Password     string        `yaml:"password"`
	DB           int           `yaml:"db" default:"0"`
	PoolSize     int           `yaml:"pool_size" default:"10"`
	MinIdleConns int           `yaml:"min_idle_conns" default:"5"`
	MaxRetries   int           `yaml:"max_retries" default:"3"`
	DialTimeout  time.Duration `yaml:"dial_timeout" default:"5s"`
	ReadTimeout  time.Duration `yaml:"read_timeout" default:"3s"`
	WriteTimeout time.Duration `yaml:"write_timeout" default:"3s"`
	KeyPrefix    string        `yaml:"key_prefix" default:"statuspage"`
}

// RedisCache provides Redis-based cache implementation
type RedisCache struct {
	config RedisCacheConfig
	logger *zap.Logger
	client *redis.Client
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(config RedisCacheConfig, logger *zap.Logger) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Connected to Redis cache",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.Int("db", config.DB))

	return &RedisCache{
		config: config,
		logger: logger,
		client: rdb,
	}, nil
}

// Get retrieves a value from Redis cache
func (r *RedisCache) Get(ctx context.Context, key string) (interface{}, bool) {
	fullKey := r.getFullKey(key)

	val, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false
		}
		r.logger.Error("Failed to get value from Redis",
			zap.String("key", key),
			zap.Error(err))
		return nil, false
	}

	// Try to unmarshal as JSON first
	var jsonValue interface{}
	if err := json.Unmarshal([]byte(val), &jsonValue); err == nil {
		return jsonValue, true
	}

	// Return as string if not JSON
	return val, true
}

// Set stores a value in Redis cache
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := r.getFullKey(key)

	var val string
	switch v := value.(type) {
	case string:
		val = v
	case []byte:
		val = string(v)
	default:
		// Marshal to JSON
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		val = string(data)
	}

	if err := r.client.Set(ctx, fullKey, val, ttl).Err(); err != nil {
		r.logger.Error("Failed to set value in Redis",
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("failed to set value in Redis: %w", err)
	}

	return nil
}

// Delete removes a value from Redis cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := r.getFullKey(key)

	if err := r.client.Del(ctx, fullKey).Err(); err != nil {
		r.logger.Error("Failed to delete value from Redis",
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("failed to delete value from Redis: %w", err)
	}

	return nil
}

// Clear removes all keys with the prefix from Redis cache
func (r *RedisCache) Clear(ctx context.Context) error {
	pattern := r.getFullKey("*")

	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		r.logger.Error("Failed to get keys from Redis",
			zap.String("pattern", pattern),
			zap.Error(err))
		return fmt.Errorf("failed to get keys from Redis: %w", err)
	}

	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			r.logger.Error("Failed to delete keys from Redis",
				zap.Strings("keys", keys),
				zap.Error(err))
			return fmt.Errorf("failed to delete keys from Redis: %w", err)
		}
	}

	return nil
}

// GetOrSet retrieves a value from cache or sets it using the provided function
func (r *RedisCache) GetOrSet(ctx context.Context, key string, fn func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	// Try to get from cache first
	if value, exists := r.Get(ctx, key); exists {
		return value, nil
	}

	// Generate value using function
	value, err := fn()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := r.Set(ctx, key, value, ttl); err != nil {
		r.logger.Warn("Failed to cache value",
			zap.String("key", key),
			zap.Error(err))
	}

	return value, nil
}

// Exists checks if a key exists in Redis cache
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := r.getFullKey(key)

	count, err := r.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	return count > 0, nil
}

// Expire sets expiration time for a key
func (r *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := r.getFullKey(key)

	if err := r.client.Expire(ctx, fullKey, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set expiration: %w", err)
	}

	return nil
}

// TTL returns the time to live for a key
func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := r.getFullKey(key)

	ttl, err := r.client.TTL(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}

	return ttl, nil
}

// Increment increments a numeric value in Redis
func (r *RedisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	fullKey := r.getFullKey(key)

	val, err := r.client.IncrBy(ctx, fullKey, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment value: %w", err)
	}

	return val, nil
}

// SetNX sets a key only if it doesn't exist (atomic operation)
func (r *RedisCache) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	fullKey := r.getFullKey(key)

	var val string
	switch v := value.(type) {
	case string:
		val = v
	case []byte:
		val = string(v)
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return false, fmt.Errorf("failed to marshal value: %w", err)
		}
		val = string(data)
	}

	result, err := r.client.SetNX(ctx, fullKey, val, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set NX: %w", err)
	}

	return result, nil
}

// GetStats returns Redis cache statistics
func (r *RedisCache) GetStats(ctx context.Context) (map[string]interface{}, error) {
	_, err := r.client.Info(ctx, "memory", "stats", "clients").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// Parse Redis info (simplified)
	stats := map[string]interface{}{
		"connected_clients": 0,
		"used_memory":       0,
		"keyspace_hits":     0,
		"keyspace_misses":   0,
	}

	// In a real implementation, you'd parse the info string
	// For now, return basic stats
	return stats, nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// getFullKey returns the full key with prefix
func (r *RedisCache) getFullKey(key string) string {
	if r.config.KeyPrefix != "" {
		return fmt.Sprintf("%s:%s", r.config.KeyPrefix, key)
	}
	return key
}

// RedisCacheMiddleware provides Redis-based caching middleware for HTTP handlers
type RedisCacheMiddleware struct {
	cache *RedisCache
	ttl   time.Duration
}

// NewRedisCacheMiddleware creates a new Redis cache middleware
func NewRedisCacheMiddleware(cache *RedisCache, ttl time.Duration) *RedisCacheMiddleware {
	return &RedisCacheMiddleware{
		cache: cache,
		ttl:   ttl,
	}
}

// CacheKey generates a cache key from request
func (m *RedisCacheMiddleware) CacheKey(c *gin.Context) string {
	return fmt.Sprintf("http:%s:%s:%s", c.Request.Method, c.Request.URL.Path, c.Request.URL.RawQuery)
}

// Handler returns a Gin middleware function
func (m *RedisCacheMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		key := m.CacheKey(c)

		// Try to get from cache
		if cached, exists := m.cache.Get(c.Request.Context(), key); exists {
			if data, ok := cached.([]byte); ok {
				c.Data(http.StatusOK, "application/json", data)
				c.Abort()
				return
			}
		}

		// Capture response
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           make([]byte, 0),
		}
		c.Writer = writer

		c.Next()

		// Cache successful responses
		if writer.statusCode == http.StatusOK && len(writer.body) > 0 {
			m.cache.Set(c.Request.Context(), key, writer.body, m.ttl)
		}
	}
}

// RedisJSONCache provides Redis-specific JSON caching utilities
type RedisJSONCache struct {
	cache *RedisCache
}

// NewRedisJSONCache creates a new Redis JSON cache
func NewRedisJSONCache(cache *RedisCache) *RedisJSONCache {
	return &RedisJSONCache{cache: cache}
}

// Get retrieves and unmarshals a JSON value from Redis cache
func (j *RedisJSONCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	cached, exists := j.cache.Get(ctx, key)
	if !exists {
		return false, nil
	}

	data, ok := cached.([]byte)
	if !ok {
		return false, fmt.Errorf("cached value is not []byte")
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return false, fmt.Errorf("failed to unmarshal cached JSON: %w", err)
	}

	return true, nil
}

// Set marshals and stores a JSON value in Redis cache
func (j *RedisJSONCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value to JSON: %w", err)
	}

	return j.cache.Set(ctx, key, data, ttl)
}

// GetOrSet retrieves JSON from cache or sets it using the provided function
func (j *RedisJSONCache) GetOrSet(ctx context.Context, key string, dest interface{}, fn func() (interface{}, error), ttl time.Duration) error {
	// Try to get from cache first
	if found, err := j.Get(ctx, key, dest); err != nil {
		return err
	} else if found {
		return nil
	}

	// Generate value using function
	value, err := fn()
	if err != nil {
		return err
	}

	// Store in cache
	if err := j.Set(ctx, key, value, ttl); err != nil {
		// Log warning but don't fail
		return err
	}

	// Set the destination
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}
