package cache

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Cache defines the interface for cache operations with generic value support
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Del(ctx context.Context, keys ...string) error
	IsHealthy(ctx context.Context) bool
}

// RedisClient wraps redis client with health checking, graceful fallback, and circuit breaker
// Uses atomic operations to prevent race conditions
type RedisClient struct {
	client    *redis.Client
	logger    *zap.Logger
	enabled   int32         // atomic: 0 = disabled, 1 = enabled
	healthTTL time.Duration
	lastCheck int64         // atomic: Unix nanoseconds timestamp
	mu        sync.RWMutex  // protects health check operations

	// Circuit breaker fields
	failureCount    int32 // atomic: consecutive failure count
	circuitOpen     int32 // atomic: 0 = closed, 1 = open (tripped)
	lastFailureTime int64 // atomic: Unix nanoseconds of last failure

	// Metrics (atomic counters)
	cacheHits   int64 // atomic: total cache hits
	cacheMisses int64 // atomic: total cache misses
	cacheErrors int64 // atomic: total cache errors
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	Enabled  bool
}

// Circuit breaker constants
const (
	circuitBreakerThreshold = 5             // Open circuit after 5 consecutive failures
	circuitBreakerTimeout   = 30 * time.Second // Try closing circuit after 30 seconds
)

// Tracer for distributed tracing
var tracer = otel.Tracer("saas-admin-service/cache")

// NewRedisClient creates a new Redis client with health checking and circuit breaker
func NewRedisClient(config RedisConfig, logger *zap.Logger) *RedisClient {
	if !config.Enabled {
		logger.Info("Redis disabled - using database-only mode")
		return &RedisClient{
			client:  nil,
			logger:  logger,
			enabled: 0, // atomic: 0 = disabled
		}
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// CPU-based connection pool sizing
	// Development: 10 connections per CPU core
	// Production: 25 connections per CPU core (adjust via environment)
	numCPU := runtime.NumCPU()
	poolSize := numCPU * 10
	minIdleConns := poolSize / 2

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     config.Password,
		DB:           config.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		MaxIdleConns: minIdleConns,
		ConnMaxIdleTime: 10 * time.Minute,
		ConnMaxLifetime: 1 * time.Hour,
	})

	rc := &RedisClient{
		client:    client,
		logger:    logger,
		enabled:   1, // atomic: 1 = enabled
		healthTTL: 30 * time.Second,
		lastCheck: 0,
		mu:        sync.RWMutex{},
	}

	// Initial health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rc.Ping(ctx); err != nil {
		logger.Warn("Redis initial connection failed - will retry",
			zap.String("addr", addr),
			zap.Int("cpus", numCPU),
			zap.Int("pool_size", poolSize),
			zap.Error(err))
		atomic.StoreInt32(&rc.enabled, 0)
	} else {
		logger.Info("Redis connected successfully",
			zap.String("addr", addr),
			zap.Int("cpus", numCPU),
			zap.Int("pool_size", poolSize),
			zap.Int("min_idle", minIdleConns))
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
// Uses atomic operations and double-checked locking to prevent race conditions
func (r *RedisClient) IsHealthy(ctx context.Context) bool {
	// Quick check without lock (atomic read)
	if atomic.LoadInt32(&r.enabled) == 0 {
		return false
	}

	if r.client == nil {
		return false
	}

	// Check if health check is recent (atomic read)
	lastCheck := atomic.LoadInt64(&r.lastCheck)
	if time.Since(time.Unix(0, lastCheck)) < r.healthTTL {
		return atomic.LoadInt32(&r.enabled) == 1
	}

	// Need to perform health check - acquire lock
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring lock (another goroutine may have checked)
	lastCheck = atomic.LoadInt64(&r.lastCheck)
	if time.Since(time.Unix(0, lastCheck)) < r.healthTTL {
		return atomic.LoadInt32(&r.enabled) == 1
	}

	// Perform health check
	if err := r.Ping(ctx); err != nil {
		atomic.StoreInt32(&r.enabled, 0)
		atomic.StoreInt64(&r.lastCheck, time.Now().UnixNano())
		r.logger.Warn("Redis health check failed", zap.Error(err))
		return false
	}

	atomic.StoreInt32(&r.enabled, 1)
	atomic.StoreInt64(&r.lastCheck, time.Now().UnixNano())
	return true
}

// Set stores a key-value pair with expiration
// Uses a 3-second timeout to prevent goroutine leaks
// Serializes value to JSON before storing
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// Start distributed tracing span
	ctx, span := tracer.Start(ctx, "cache.Set",
		trace.WithAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.operation", "set"),
			attribute.Int64("cache.ttl_seconds", int64(expiration.Seconds())),
		))
	defer span.End()

	if !r.IsHealthy(ctx) {
		atomic.AddInt64(&r.cacheErrors, 1)
		span.SetAttributes(attribute.Bool("cache.healthy", false))
		span.RecordError(fmt.Errorf("redis not available"))
		return fmt.Errorf("redis not available")
	}

	// Check circuit breaker
	if !r.shouldAllowRequest() {
		atomic.AddInt64(&r.cacheErrors, 1)
		span.SetAttributes(attribute.Bool("cache.circuit_open", true))
		span.RecordError(fmt.Errorf("circuit breaker open"))
		return fmt.Errorf("circuit breaker open")
	}

	// Serialize value to JSON
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		atomic.AddInt64(&r.cacheErrors, 1)
		span.SetAttributes(attribute.Bool("cache.success", false))
		span.RecordError(fmt.Errorf("failed to marshal value: %w", err))
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Add timeout to prevent goroutine leaks
	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err = r.client.Set(timeoutCtx, key, jsonBytes, expiration).Err()

	// Track metrics and circuit breaker
	if err != nil {
		atomic.AddInt64(&r.cacheErrors, 1)
		r.recordFailure()
		span.SetAttributes(attribute.Bool("cache.success", false))
		span.RecordError(err)
	} else {
		r.recordSuccess()
		span.SetAttributes(attribute.Bool("cache.success", true))
	}

	return err
}

// Get retrieves a value by key and deserializes into dest
// Uses a 3-second timeout to prevent goroutine leaks
func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	// Start distributed tracing span
	ctx, span := tracer.Start(ctx, "cache.Get",
		trace.WithAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.operation", "get"),
		))
	defer span.End()

	if !r.IsHealthy(ctx) {
		atomic.AddInt64(&r.cacheErrors, 1)
		span.SetAttributes(attribute.Bool("cache.healthy", false))
		span.RecordError(fmt.Errorf("redis not available"))
		return fmt.Errorf("redis not available")
	}

	// Check circuit breaker
	if !r.shouldAllowRequest() {
		atomic.AddInt64(&r.cacheErrors, 1)
		span.SetAttributes(attribute.Bool("cache.circuit_open", true))
		span.RecordError(fmt.Errorf("circuit breaker open"))
		return fmt.Errorf("circuit breaker open")
	}

	// Add timeout to prevent goroutine leaks
	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	result, err := r.client.Get(timeoutCtx, key).Result()

	// Track metrics and circuit breaker
	if err != nil {
		if err != redis.Nil { // Nil = cache miss, not a failure
			atomic.AddInt64(&r.cacheErrors, 1)
			r.recordFailure()
			span.SetAttributes(attribute.Bool("cache.hit", false), attribute.Bool("cache.error", true))
			span.RecordError(err)
		} else {
			atomic.AddInt64(&r.cacheMisses, 1)
			span.SetAttributes(attribute.Bool("cache.hit", false), attribute.Bool("cache.miss", true))
		}
		return err
	}

	// Deserialize JSON into dest
	if err := json.Unmarshal([]byte(result), dest); err != nil {
		atomic.AddInt64(&r.cacheErrors, 1)
		span.SetAttributes(attribute.Bool("cache.hit", true), attribute.Bool("cache.unmarshal_error", true))
		span.RecordError(fmt.Errorf("failed to unmarshal value: %w", err))
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	atomic.AddInt64(&r.cacheHits, 1)
	r.recordSuccess()
	span.SetAttributes(
		attribute.Bool("cache.hit", true),
		attribute.Int("cache.value_size", len(result)),
	)
	return nil
}

// Del deletes a key
// Uses a 3-second timeout to prevent goroutine leaks
func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	if !r.IsHealthy(ctx) {
		return fmt.Errorf("redis not available")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return r.client.Del(timeoutCtx, keys...).Err()
}

// Exists checks if a key exists
// Uses a 3-second timeout to prevent goroutine leaks
func (r *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	if !r.IsHealthy(ctx) {
		return 0, fmt.Errorf("redis not available")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return r.client.Exists(timeoutCtx, keys...).Result()
}

// Expire sets expiration on a key
// Uses a 3-second timeout to prevent goroutine leaks
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if !r.IsHealthy(ctx) {
		return fmt.Errorf("redis not available")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return r.client.Expire(timeoutCtx, key, expiration).Err()
}

// Close closes the Redis connection immediately
func (r *RedisClient) Close() error {
	if r.client == nil {
		return nil
	}

	return r.client.Close()
}

// Shutdown gracefully shuts down the Redis connection
// Waits for in-flight operations to complete within the given context timeout
func (r *RedisClient) Shutdown(ctx context.Context) error {
	if r.client == nil {
		return nil
	}

	r.logger.Info("Gracefully shutting down Redis client")

	// Mark as disabled to reject new operations
	atomic.StoreInt32(&r.enabled, 0)

	// Wait a bit for in-flight operations to complete
	select {
	case <-time.After(100 * time.Millisecond):
		// Brief pause for in-flight ops
	case <-ctx.Done():
		r.logger.Warn("Context cancelled during Redis shutdown")
		return ctx.Err()
	}

	// Close the connection
	if err := r.client.Close(); err != nil {
		r.logger.Error("Error closing Redis client", zap.Error(err))
		return err
	}

	r.logger.Info("Redis client shut down successfully")
	return nil
}

// recordSuccess resets failure count when operation succeeds
func (r *RedisClient) recordSuccess() {
	atomic.StoreInt32(&r.failureCount, 0)
	atomic.StoreInt32(&r.circuitOpen, 0)
}

// recordFailure increments failure count and opens circuit if threshold exceeded
func (r *RedisClient) recordFailure() {
	count := atomic.AddInt32(&r.failureCount, 1)
	atomic.StoreInt64(&r.lastFailureTime, time.Now().UnixNano())

	if count >= circuitBreakerThreshold {
		atomic.StoreInt32(&r.circuitOpen, 1)
		r.logger.Warn("Circuit breaker opened",
			zap.Int32("failure_count", count))
	}
}

// shouldAllowRequest checks if circuit breaker allows request
func (r *RedisClient) shouldAllowRequest() bool {
	if atomic.LoadInt32(&r.circuitOpen) == 0 {
		return true
	}

	lastFailure := atomic.LoadInt64(&r.lastFailureTime)
	if time.Since(time.Unix(0, lastFailure)) > circuitBreakerTimeout {
		r.logger.Info("Circuit breaker attempting recovery")
		return true
	}

	return false
}

// CacheMetrics represents cache performance metrics
type CacheMetrics struct {
	Hits          int64   `json:"hits"`
	Misses        int64   `json:"misses"`
	Errors        int64   `json:"errors"`
	HitRate       float64 `json:"hit_rate"`
	CircuitOpen   bool    `json:"circuit_open"`
	FailureCount  int32   `json:"failure_count"`
	Enabled       bool    `json:"enabled"`
}

// GetMetrics returns current cache metrics
func (r *RedisClient) GetMetrics() CacheMetrics {
	hits := atomic.LoadInt64(&r.cacheHits)
	misses := atomic.LoadInt64(&r.cacheMisses)
	errors := atomic.LoadInt64(&r.cacheErrors)

	total := hits + misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	return CacheMetrics{
		Hits:         hits,
		Misses:       misses,
		Errors:       errors,
		HitRate:      hitRate,
		CircuitOpen:  atomic.LoadInt32(&r.circuitOpen) == 1,
		FailureCount: atomic.LoadInt32(&r.failureCount),
		Enabled:      atomic.LoadInt32(&r.enabled) == 1,
	}
}

// compress compresses data using gzip (for large payloads > 1KB)
func compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, fmt.Errorf("gzip write failed: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("gzip close failed: %w", err)
	}

	return buf.Bytes(), nil
}

// decompress decompresses gzip data
func decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("gzip reader failed: %w", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("gzip read failed: %w", err)
	}

	return decompressed, nil
}

// SetCompressed stores compressed data in cache (use for large payloads)
func (r *RedisClient) SetCompressed(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if !r.IsHealthy(ctx) {
		atomic.AddInt64(&r.cacheErrors, 1)
		return fmt.Errorf("redis not available")
	}

	// Only compress if data is larger than 1KB
	var dataToStore []byte
	if len(value) > 1024 {
		compressed, err := compress(value)
		if err != nil {
			r.logger.Warn("Compression failed, storing uncompressed",
				zap.Error(err),
				zap.Int("original_size", len(value)))
			dataToStore = value
		} else {
			dataToStore = compressed
			r.logger.Debug("Compressed cache data",
				zap.String("key", key),
				zap.Int("original_size", len(value)),
				zap.Int("compressed_size", len(compressed)),
				zap.Float64("compression_ratio", float64(len(value))/float64(len(compressed))))
		}
	} else {
		dataToStore = value
	}

	// Check circuit breaker
	if !r.shouldAllowRequest() {
		atomic.AddInt64(&r.cacheErrors, 1)
		return fmt.Errorf("circuit breaker open")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := r.client.Set(timeoutCtx, key, dataToStore, expiration).Err()
	if err != nil {
		atomic.AddInt64(&r.cacheErrors, 1)
		r.recordFailure()
	} else {
		r.recordSuccess()
	}

	return err
}

// GetCompressed retrieves and decompresses data from cache
func (r *RedisClient) GetCompressed(ctx context.Context, key string) ([]byte, error) {
	if !r.IsHealthy(ctx) {
		atomic.AddInt64(&r.cacheErrors, 1)
		return nil, fmt.Errorf("redis not available")
	}

	// Check circuit breaker
	if !r.shouldAllowRequest() {
		atomic.AddInt64(&r.cacheErrors, 1)
		return nil, fmt.Errorf("circuit breaker open")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	data, err := r.client.Get(timeoutCtx, key).Bytes()
	if err != nil {
		if err != redis.Nil {
			atomic.AddInt64(&r.cacheErrors, 1)
			r.recordFailure()
		} else {
			atomic.AddInt64(&r.cacheMisses, 1)
		}
		return nil, err
	}

	// Try to decompress - if it fails, data might not be compressed
	decompressed, err := decompress(data)
	if err != nil {
		r.logger.Debug("Data not compressed or decompression failed, returning raw data",
			zap.String("key", key))
		atomic.AddInt64(&r.cacheHits, 1)
		r.recordSuccess()
		return data, nil
	}

	atomic.AddInt64(&r.cacheHits, 1)
	r.recordSuccess()
	return decompressed, nil
}

// GetClient returns the underlying Redis client (use with caution)
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}
