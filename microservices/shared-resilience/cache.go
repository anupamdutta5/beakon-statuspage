// Package resilience provides caching strategies for microservices
package resilience

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CacheConfig represents configuration for caching
type CacheConfig struct {
	DefaultTTL      time.Duration `yaml:"default_ttl" default:"5m"`
	MaxSize         int           `yaml:"max_size" default:"1000"`
	CleanupInterval time.Duration `yaml:"cleanup_interval" default:"10m"`
}

// CacheItem represents a cached item
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
	CreatedAt time.Time
}

// IsExpired checks if the cache item has expired
func (item *CacheItem) IsExpired() bool {
	return time.Now().After(item.ExpiresAt)
}

// InMemoryCache provides an in-memory cache implementation
type InMemoryCache struct {
	config      CacheConfig
	logger      *zap.Logger
	items       map[string]*CacheItem
	mutex       sync.RWMutex
	stop        chan struct{}
	totalHits   int64
	totalMisses int64
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache(config CacheConfig, logger *zap.Logger) *InMemoryCache {
	cache := &InMemoryCache{
		config: config,
		logger: logger,
		items:  make(map[string]*CacheItem),
		stop:   make(chan struct{}),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a value from the cache
func (c *InMemoryCache) Get(ctx context.Context, key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists || item.IsExpired() {
		c.totalMisses++
		return nil, false
	}

	c.totalHits++
	return item.Value, true
}

// Set stores a value in the cache
func (c *InMemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = c.config.DefaultTTL
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if we need to evict items
	if len(c.items) >= c.config.MaxSize {
		c.evictOldest()
	}

	c.items[key] = &CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}

	return nil
}

// Delete removes a value from the cache
func (c *InMemoryCache) Delete(ctx context.Context, key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
	return nil
}

// Clear removes all items from the cache
func (c *InMemoryCache) Clear(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*CacheItem)
	return nil
}

// GetOrSet retrieves a value from cache or sets it using the provided function
func (c *InMemoryCache) GetOrSet(ctx context.Context, key string, fn func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	// Try to get from cache first
	if value, exists := c.Get(ctx, key); exists {
		return value, nil
	}

	// Generate value using function
	value, err := fn()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := c.Set(ctx, key, value, ttl); err != nil {
		c.logger.Warn("Failed to cache value", zap.String("key", key), zap.Error(err))
	}

	return value, nil
}

// evictOldest removes the oldest item from the cache
func (c *InMemoryCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.items {
		if oldestKey == "" || item.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.CreatedAt
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}

// cleanup removes expired items periodically
func (c *InMemoryCache) cleanup() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mutex.Lock()
			for key, item := range c.items {
				if item.IsExpired() {
					delete(c.items, key)
				}
			}
			c.mutex.Unlock()
		case <-c.stop:
			return
		}
	}
}

// Stop stops the cache cleanup goroutine
func (c *InMemoryCache) Stop() {
	close(c.stop)
}

// CacheStats represents cache statistics
type CacheStats struct {
	Size        int     `json:"size"`
	MaxSize     int     `json:"max_size"`
	HitRate     float64 `json:"hit_rate"`
	MissRate    float64 `json:"miss_rate"`
	TotalHits   int64   `json:"total_hits"`
	TotalMisses int64   `json:"total_misses"`
}

// GetStats returns cache statistics
func (c *InMemoryCache) GetStats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	total := c.totalHits + c.totalMisses
	var hitRate, missRate float64
	if total > 0 {
		hitRate = float64(c.totalHits) / float64(total)
		missRate = float64(c.totalMisses) / float64(total)
	}

	return CacheStats{
		Size:        len(c.items),
		MaxSize:     c.config.MaxSize,
		HitRate:     hitRate,
		MissRate:    missRate,
		TotalHits:   c.totalHits,
		TotalMisses: c.totalMisses,
	}
}

// Cache interface defines the contract for cache implementations
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, bool)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	GetOrSet(ctx context.Context, key string, fn func() (interface{}, error), ttl time.Duration) (interface{}, error)
}

// CacheMiddleware provides caching middleware for HTTP handlers
type CacheMiddleware struct {
	cache Cache
	ttl   time.Duration
}

// NewCacheMiddleware creates a new cache middleware
func NewCacheMiddleware(cache Cache, ttl time.Duration) *CacheMiddleware {
	return &CacheMiddleware{
		cache: cache,
		ttl:   ttl,
	}
}

// CacheKey generates a cache key from request
func (m *CacheMiddleware) CacheKey(c *gin.Context) string {
	return fmt.Sprintf("%s:%s:%s", c.Request.Method, c.Request.URL.Path, c.Request.URL.RawQuery)
}

// Handler returns a Gin middleware function
func (m *CacheMiddleware) Handler() gin.HandlerFunc {
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

// responseWriter captures the response for caching
type responseWriter struct {
	gin.ResponseWriter
	body       []byte
	statusCode int
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// JSONCache provides JSON-specific caching utilities
type JSONCache struct {
	cache Cache
}

// NewJSONCache creates a new JSON cache
func NewJSONCache(cache Cache) *JSONCache {
	return &JSONCache{cache: cache}
}

// Get retrieves and unmarshals a JSON value from cache
func (j *JSONCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
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

// Set marshals and stores a JSON value in cache
func (j *JSONCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value to JSON: %w", err)
	}

	return j.cache.Set(ctx, key, data, ttl)
}
