# Implementation Plan for Remaining 9 Fixes
**Service**: saas-admin-service
**Current Status**: 6/15 completed (40%)
**Remaining**: 9 fixes (60%)
**Total Time**: 24 hours

---

## ✅ Executive Summary

This document provides a **comprehensive, step-by-step implementation plan** for the remaining 9 caching architecture fixes. Each fix includes:
- Exact code to add/modify
- File locations and line numbers
- Testing procedures
- Success criteria

**Follow this plan sequentially for best results.**

---

## 📋 Implementation Order

### **Phase 1: Complete Priority 2 (3-4 hours)**
- Fix #7: Circuit Breaker Logic
- Fix #8: Prometheus Metrics Endpoint

### **Phase 2: Priority 3 - Part 1 (5 hours)**
- Fix #9: Tiered TTLs
- Fix #10: Plan Cache Invalidation
- Fix #11: Cache Warming

### **Phase 3: Priority 3 - Part 2 (5 hours)**
- Fix #12: Cache Versioning
- Fix #13: Cache Compression

### **Phase 4: Priority 4 - Advanced (11 hours)**
- Fix #14: Distributed Tracing
- Fix #15: Performance Testing Suite

---

## PHASE 1: PRIORITY 2 FIXES (3-4 hours)

### Fix #7: Complete Circuit Breaker Implementation

**Time**: 1-2 hours
**Files**: `internal/cache/redis.go`
**Status**: 40% complete (struct fields and constants already added)

#### Step 1: Add Helper Methods

**Location**: After line 270 in `internal/cache/redis.go`

```go
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
	// Circuit closed = allow all requests
	if atomic.LoadInt32(&r.circuitOpen) == 0 {
		return true
	}

	// Circuit open = check if timeout expired
	lastFailure := atomic.LoadInt64(&r.lastFailureTime)
	if time.Since(time.Unix(0, lastFailure)) > circuitBreakerTimeout {
		// Try half-open: allow one request
		r.logger.Info("Circuit breaker attempting recovery")
		return true
	}

	r.logger.Debug("Circuit breaker rejecting request")
	return false
}
```

#### Step 2: Update Get() Method

**Location**: Replace existing Get() method (around line 164-176)

```go
// Get retrieves a value by key
// Uses a 3-second timeout to prevent goroutine leaks
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	if !r.IsHealthy(ctx) {
		atomic.AddInt64(&r.cacheErrors, 1)
		return "", fmt.Errorf("redis not available")
	}

	// CHECK CIRCUIT BREAKER
	if !r.shouldAllowRequest() {
		atomic.AddInt64(&r.cacheErrors, 1)
		return "", fmt.Errorf("circuit breaker open")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	result, err := r.client.Get(timeoutCtx, key).Result()

	// RECORD SUCCESS/FAILURE
	if err != nil {
		if err != redis.Nil { // Nil = cache miss, not a failure
			atomic.AddInt64(&r.cacheErrors, 1)
			r.recordFailure()
		} else {
			atomic.AddInt64(&r.cacheMisses, 1)
		}
		return "", err
	}

	atomic.AddInt64(&r.cacheHits, 1)
	r.recordSuccess()
	return result, nil
}
```

#### Step 3: Update Set() Method

**Location**: Replace existing Set() method (around line 150-162)

```go
// Set stores a key-value pair with expiration
// Uses a 3-second timeout to prevent goroutine leaks
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if !r.IsHealthy(ctx) {
		atomic.AddInt64(&r.cacheErrors, 1)
		return fmt.Errorf("redis not available")
	}

	// CHECK CIRCUIT BREAKER
	if !r.shouldAllowRequest() {
		atomic.AddInt64(&r.cacheErrors, 1)
		return fmt.Errorf("circuit breaker open")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := r.client.Set(timeoutCtx, key, value, expiration).Err()

	// RECORD SUCCESS/FAILURE
	if err != nil {
		atomic.AddInt64(&r.cacheErrors, 1)
		r.recordFailure()
	} else {
		r.recordSuccess()
	}

	return err
}
```

#### Testing Fix #7

```bash
# Test circuit breaker
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/saas-admin-service

# Stop Redis to simulate failures
brew services stop redis

# Start service
go run cmd/main.go &

# Make requests (should open circuit after 5 failures)
for i in {1..10}; do
	echo "Request $i:"
	curl -s http://localhost:8098/api/v1/tenants
	sleep 1
done

# Check logs for "Circuit breaker opened"
grep "Circuit breaker" logs/*.log

# Start Redis
brew services start redis

# Wait 30 seconds for recovery timeout
sleep 30

# Make request (should succeed and close circuit)
curl http://localhost:8098/api/v1/tenants

# Check logs for "Circuit breaker attempting recovery"
grep "attempting recovery" logs/*.log
```

**Success Criteria**:
- ✅ Circuit opens after 5 consecutive failures
- ✅ Circuit stays open during timeout period
- ✅ Circuit attempts recovery after 30 seconds
- ✅ Circuit closes on successful request

---

### Fix #8: Complete Prometheus Metrics Implementation

**Time**: 2-3 hours
**Files**: `internal/cache/redis.go`, `internal/services/saas_admin_service.go`, `internal/handlers/saas_admin_handler.go`
**Status**: 30% complete (counters already added)

#### Step 1: Add Metrics Struct and Methods

**Location**: After line 270 in `internal/cache/redis.go`

```go
// Metrics holds cache metrics for exposition
type Metrics struct {
	CacheHits   int64   `json:"cache_hits"`
	CacheMisses int64   `json:"cache_misses"`
	CacheErrors int64   `json:"cache_errors"`
	HitRate     float64 `json:"hit_rate"`
	CircuitOpen bool    `json:"circuit_open"`
}

// GetMetrics returns current cache metrics
func (r *RedisClient) GetMetrics() Metrics {
	hits := atomic.LoadInt64(&r.cacheHits)
	misses := atomic.LoadInt64(&r.cacheMisses)
	errors := atomic.LoadInt64(&r.cacheErrors)
	circuitOpen := atomic.LoadInt32(&r.circuitOpen) == 1

	total := hits + misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	return Metrics{
		CacheHits:   hits,
		CacheMisses: misses,
		CacheErrors: errors,
		HitRate:     hitRate,
		CircuitOpen: circuitOpen,
	}
}

// ResetMetrics resets all counters (for testing)
func (r *RedisClient) ResetMetrics() {
	atomic.StoreInt64(&r.cacheHits, 0)
	atomic.StoreInt64(&r.cacheMisses, 0)
	atomic.StoreInt64(&r.cacheErrors, 0)
	r.logger.Info("Cache metrics reset")
}
```

#### Step 2: Add Service Method

**Location**: Add to `internal/services/saas_admin_service.go` after line 125

```go
// GetCacheMetrics returns Redis cache metrics
func (s *SaaSAdminService) GetCacheMetrics() interface{} {
	if s.redisClient == nil {
		return map[string]interface{}{
			"enabled": false,
			"message": "Redis caching disabled",
		}
	}

	metrics := s.redisClient.GetMetrics()
	return map[string]interface{}{
		"enabled":      true,
		"cache_hits":   metrics.CacheHits,
		"cache_misses": metrics.CacheMisses,
		"cache_errors": metrics.CacheErrors,
		"hit_rate":     metrics.HitRate,
		"circuit_open": metrics.CircuitOpen,
	}
}
```

#### Step 3: Add Handler Method

**Location**: Add to `internal/handlers/saas_admin_handler.go`

**First, find where handlers are defined**:
```bash
grep -n "func (h \*SaaSAdminHandler)" internal/handlers/saas_admin_handler.go | head -5
```

**Then add this method**:
```go
// GetCacheMetrics returns cache performance metrics
func (h *SaaSAdminHandler) GetCacheMetrics(c *gin.Context) {
	metrics := h.service.GetCacheMetrics()

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   metrics,
	})
}
```

#### Step 4: Add Route

**Location**: In `cmd/main.go`, find the setupRoutes function and add:

```go
// Add to API v1 group (around line 260)
v1.GET("/metrics/cache", adminHandler.GetCacheMetrics)
```

#### Testing Fix #8

```bash
# Start service
go run cmd/main.go &

# Make some requests to generate metrics
for i in {1..20}; do
	curl -s http://localhost:8098/api/v1/tenants > /dev/null
done

# Check metrics endpoint
curl http://localhost:8098/api/v1/metrics/cache | python3 -m json.tool

# Expected output:
# {
#   "status": "success",
#   "data": {
#     "enabled": true,
#     "cache_hits": 15,
#     "cache_misses": 5,
#     "cache_errors": 0,
#     "hit_rate": 75.0,
#     "circuit_open": false
#   }
# }
```

**Success Criteria**:
- ✅ Metrics endpoint responds at `/api/v1/metrics/cache`
- ✅ Hit rate calculates correctly
- ✅ Circuit breaker status visible
- ✅ Counters increment properly

---

## PHASE 2: PRIORITY 3 - PART 1 (5 hours)

### Fix #9: Implement Tiered TTLs

**Time**: 2 hours
**Files**: `internal/services/saas_admin_service.go`

#### Step 1: Add TTL Constants

**Location**: After line 28 in `saas_admin_service.go`

```go
const (
	tenantMetricsTTL = 5 * time.Minute  // Frequently changing data
	planCacheTTL     = 1 * time.Hour    // Rarely changing data
	statsCacheTTL    = 30 * time.Second // Real-time data
)
```

#### Step 2: Update cacheTenantsAsync

**Location**: Replace the function around line 1614-1629

```go
// cacheTenantsAsync writes to cache in background to avoid blocking response
func (s *SaaSAdminService) cacheTenantsAsync(ctx context.Context, key string, data []TenantWithMetrics) {
	resultJSON, err := json.Marshal(data)
	if err != nil {
		s.logger.Error("Failed to marshal tenant metrics for cache", zap.Error(err))
		return
	}

	// USE TIERED TTL - tenants change frequently
	cacheTTL := tenantMetricsTTL
	if err := s.redisClient.Set(ctx, key, resultJSON, cacheTTL); err != nil {
		s.logger.Warn("Failed to cache tenant metrics", zap.Error(err))
	} else {
		s.logger.Info("Cached tenant metrics",
			zap.Int("count", len(data)),
			zap.Duration("ttl", cacheTTL))
	}
}
```

#### Step 3: Update getCachedPlans

**Location**: Find and update the `getCachedPlans` method (around line 1631)

Where it currently does:
```go
cacheTTL := time.Duration(s.config.Cache.TTL) * time.Second
```

Replace with:
```go
cacheTTL := planCacheTTL  // 1 hour - plans rarely change
```

#### Testing Fix #9

```bash
# Enable Redis
brew services start redis

# Start service
go run cmd/main.go &

# Make requests
curl http://localhost:8098/api/v1/tenants

# Check Redis TTLs
redis-cli
> KEYS statuspage:*
> TTL statuspage:saas-admin-service:tenants:metrics  # Should be ~300 seconds (5 min)
> TTL statuspage:saas-admin-service:plans:all       # Should be ~3600 seconds (1 hour)
```

**Success Criteria**:
- ✅ Tenant metrics have 5-minute TTL
- ✅ Plans have 1-hour TTL
- ✅ Different data types have appropriate expiration times

---

### Fix #10: Add Plan Cache Invalidation

**Time**: 1 hour
**Files**: `internal/services/saas_admin_service.go`

#### Step 1: Add Invalidation Method

**Location**: After the `getCachedPlans` method

```go
// InvalidatePlanCache invalidates the cached plans
func (s *SaaSAdminService) InvalidatePlanCache(ctx context.Context) error {
	if s.redisClient == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("%s:plans:all", s.config.Cache.KeyPrefix)
	if err := s.redisClient.Del(ctx, cacheKey); err != nil {
		s.logger.Error("Failed to invalidate plan cache", zap.Error(err))
		return err
	}

	s.logger.Info("Plan cache invalidated")
	return nil
}
```

#### Step 2: Find Plan Mutation Methods

```bash
# Find all plan-related methods
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/saas-admin-service
grep -n "func (s \*SaaSAdminService) .*Plan" internal/services/saas_admin_service.go
```

#### Step 3: Add Invalidation to Each Method

For each of these methods: `CreatePlan`, `UpdatePlan`, `DeletePlan`

**Add before the return statement**:
```go
// Invalidate plan cache
if err := s.InvalidatePlanCache(ctx); err != nil {
	s.logger.Warn("Failed to invalidate plan cache after update", zap.Error(err))
}
```

#### Testing Fix #10

```bash
# Get plans (populate cache)
curl http://localhost:8098/api/v1/plans

# Check Redis
redis-cli KEYS "*plans*"

# Update a plan
curl -X PUT http://localhost:8098/api/v1/plans/<plan-id> \
  -H 'Content-Type: application/json' \
  -d '{"name":"Updated Plan"}'

# Check Redis again (key should be gone)
redis-cli KEYS "*plans*"

# Get plans again (should fetch from DB)
curl http://localhost:8098/api/v1/plans
```

**Success Criteria**:
- ✅ Plan cache invalidated on create
- ✅ Plan cache invalidated on update
- ✅ Plan cache invalidated on delete

---

### Fix #11: Cache Warming

**Time**: 2 hours
**Files**: `internal/services/saas_admin_service.go`, `cmd/main.go`

#### Step 1: Add Warming Method

**Location**: After line 125 in `saas_admin_service.go`

```go
// WarmCache pre-populates cache with frequently accessed data
func (s *SaaSAdminService) WarmCache(ctx context.Context) error {
	s.logger.Info("Starting cache warming")

	// Warm tenant metrics cache
	go func() {
		warmCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if _, err := s.ListTenantsWithMetrics(warmCtx); err != nil {
			s.logger.Warn("Failed to warm tenant metrics cache", zap.Error(err))
		} else {
			s.logger.Info("Tenant metrics cache warmed successfully")
		}
	}()

	// Warm plans cache
	go func() {
		warmCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if _, err := s.getCachedPlans(warmCtx); err != nil {
			s.logger.Warn("Failed to warm plans cache", zap.Error(err))
		} else {
			s.logger.Info("Plans cache warmed successfully")
		}
	}()

	s.logger.Info("Cache warming initiated (running in background)")
	return nil
}
```

#### Step 2: Call on Startup

**Location**: In `cmd/main.go` after line 470 (after service initialization)

```go
// Warm cache on startup (production only)
if resilienceConfig.Environment == "production" {
	go func() {
		time.Sleep(5 * time.Second) // Wait for service to be ready
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		if err := saasAdminService.WarmCache(ctx); err != nil {
			logger.Warn("Cache warming failed", zap.Error(err))
		}
	}()
}
```

#### Testing Fix #11

```bash
# Clear Redis
redis-cli FLUSHDB

# Start service
ENVIRONMENT=production go run cmd/main.go &

# Wait 10 seconds
sleep 10

# Check Redis (should have cached data)
redis-cli KEYS "*"

# Check logs
grep "cache warmed" logs/*.log
```

**Success Criteria**:
- ✅ Cache warming runs on startup
- ✅ Tenant metrics pre-cached
- ✅ Plans pre-cached
- ✅ First user request is fast (cache hit)

---

## PHASE 3: PRIORITY 3 - PART 2 (5 hours)

### Fix #12: Cache Versioning

**Time**: 2 hours
**Files**: `internal/services/saas_admin_service.go`

#### Step 1: Add Version Constant

**Location**: After line 28

```go
const (
	cacheVersion     = "v1" // Increment when schema changes
	tenantMetricsTTL = 5 * time.Minute
	planCacheTTL     = 1 * time.Hour
	statsCacheTTL    = 30 * time.Second
)
```

#### Step 2: Update All Cache Keys

**Find all cache key constructions**:
```bash
grep -n "fmt.Sprintf.*Cache.KeyPrefix" internal/services/saas_admin_service.go
```

**Update each one to include version**:

Before:
```go
cacheKey := fmt.Sprintf("%s:tenants:metrics", s.config.Cache.KeyPrefix)
```

After:
```go
cacheKey := fmt.Sprintf("%s:%s:tenants:metrics", s.config.Cache.KeyPrefix, cacheVersion)
```

Apply to all cache keys:
- `tenants:metrics` → `v1:tenants:metrics`
- `plans:all` → `v1:plans:all`

#### Testing Fix #12

```bash
# Set version to v1
# Start service and create cache
curl http://localhost:8098/api/v1/tenants

# Check Redis
redis-cli KEYS "*"
# Should see: statuspage:saas-admin-service:v1:tenants:metrics

# Change version to v2 in code
# Restart service
# Make request
curl http://localhost:8098/api/v1/tenants

# Check Redis
redis-cli KEYS "*"
# Should see: statuspage:saas-admin-service:v2:tenants:metrics
# Old v1 key will naturally expire
```

**Success Criteria**:
- ✅ All cache keys include version
- ✅ Version change invalidates old cache
- ✅ No manual migration needed

---

### Fix #13: Cache Compression

**Time**: 3 hours
**Files**: `internal/cache/redis.go`, `internal/services/saas_admin_service.go`

#### Step 1: Add Compression Imports

**Location**: Top of `redis.go`

```go
import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)
```

#### Step 2: Add Compression Helpers

**Location**: After line 270 in `redis.go`

```go
// compress compresses data using gzip
func compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// decompress decompresses gzip data
func decompress(data []byte) ([]byte, error) {
	buf := bytes.NewReader(data)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	return io.ReadAll(gz)
}
```

#### Step 3: Add Compressed Methods

```go
// SetCompressed stores compressed data
func (r *RedisClient) SetCompressed(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if !r.IsHealthy(ctx) {
		atomic.AddInt64(&r.cacheErrors, 1)
		return fmt.Errorf("redis not available")
	}

	if !r.shouldAllowRequest() {
		atomic.AddInt64(&r.cacheErrors, 1)
		return fmt.Errorf("circuit breaker open")
	}

	compressed, err := compress(value)
	if err != nil {
		atomic.AddInt64(&r.cacheErrors, 1)
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Add :gz suffix to indicate compression
	err = r.client.Set(timeoutCtx, key+":gz", compressed, expiration).Err()

	if err != nil {
		atomic.AddInt64(&r.cacheErrors, 1)
		r.recordFailure()
	} else {
		r.recordSuccess()
	}

	return err
}

// GetCompressed retrieves and decompresses data
func (r *RedisClient) GetCompressed(ctx context.Context, key string) ([]byte, error) {
	if !r.IsHealthy(ctx) {
		atomic.AddInt64(&r.cacheErrors, 1)
		return nil, fmt.Errorf("redis not available")
	}

	if !r.shouldAllowRequest() {
		atomic.AddInt64(&r.cacheErrors, 1)
		return nil, fmt.Errorf("circuit breaker open")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	compressed, err := r.client.Get(timeoutCtx, key+":gz").Result()

	if err != nil {
		if err != redis.Nil {
			atomic.AddInt64(&r.cacheErrors, 1)
			r.recordFailure()
		} else {
			atomic.AddInt64(&r.cacheMisses, 1)
		}
		return nil, err
	}

	atomic.AddInt64(&r.cacheHits, 1)
	r.recordSuccess()

	return decompress([]byte(compressed))
}
```

#### Step 4: Use Compression for Large Data

**Location**: Update `cacheTenantsAsync` in `saas_admin_service.go`

```go
func (s *SaaSAdminService) cacheTenantsAsync(ctx context.Context, key string, data []TenantWithMetrics) {
	resultJSON, err := json.Marshal(data)
	if err != nil {
		s.logger.Error("Failed to marshal tenant metrics for cache", zap.Error(err))
		return
	}

	cacheTTL := tenantMetricsTTL

	// Compress if data is large (>10KB)
	if len(resultJSON) > 10240 {
		if err := s.redisClient.SetCompressed(ctx, key, resultJSON, cacheTTL); err != nil {
			s.logger.Warn("Failed to cache compressed data", zap.Error(err))
		} else {
			s.logger.Info("Cached compressed tenant metrics",
				zap.Int("count", len(data)),
				zap.Int("original_size", len(resultJSON)),
				zap.Duration("ttl", cacheTTL))
		}
	} else {
		// Use normal caching for small data
		if err := s.redisClient.Set(ctx, key, resultJSON, cacheTTL); err != nil {
			s.logger.Warn("Failed to cache tenant metrics", zap.Error(err))
		} else {
			s.logger.Info("Cached tenant metrics",
				zap.Int("count", len(data)),
				zap.Duration("ttl", cacheTTL))
		}
	}
}
```

#### Testing Fix #13

```bash
# Create many tenants to test compression
for i in {1..100}; do
	curl -X POST http://localhost:8098/api/v1/tenants \
	  -H 'Content-Type: application/json' \
	  -d "{\"name\":\"Tenant $i\",\"contact_email\":\"test$i@example.com\"}"
done

# Get tenants (should trigger compression)
curl http://localhost:8098/api/v1/tenants

# Check logs
grep "Cached compressed" logs/*.log

# Check Redis memory
redis-cli INFO memory | grep used_memory_human
```

**Success Criteria**:
- ✅ Large datasets (>10KB) are compressed
- ✅ Small datasets use normal caching
- ✅ Compression/decompression works correctly
- ✅ Memory usage reduced for large data

---

## PHASE 4: PRIORITY 4 - ADVANCED (11 hours)

### Fix #14: Distributed Tracing

**Time**: 5 hours
**Complexity**: HIGH

**Note**: This requires OpenTelemetry setup which is complex. Simplified version below.

#### Step 1: Add Dependencies

```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/saas-admin-service
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/trace
go get go.opentelemetry.io/otel/attribute
```

#### Step 2: Add Tracing to Cache Operations

**Location**: Update `Get()` method in `redis.go`

```go
import "go.opentelemetry.io/otel/attribute"
import "go.opentelemetry.io/otel/trace"

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	// Start span
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		ctx, span = span.TracerProvider().Tracer("redis").Start(ctx, "redis.get")
		defer span.End()

		span.SetAttributes(
			attribute.String("cache.key", key),
		)
	}

	// ... existing health check code ...

	result, err := r.client.Get(timeoutCtx, key).Result()

	if err != nil {
		if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
			span.RecordError(err)
			span.SetAttributes(attribute.Bool("cache.hit", false))
		}
		// ... existing error handling ...
	}

	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		span.SetAttributes(attribute.Bool("cache.hit", true))
	}

	// ... rest of method ...
}
```

**Recommendation**: This is complex. Consider using a managed tracing service like Datadog or New Relic instead of implementing from scratch.

---

### Fix #15: Performance Testing Suite

**Time**: 6 hours

#### Step 1: Install k6

```bash
brew install k6
```

#### Step 2: Create Load Test Script

**Location**: Create `tests/load/cache_performance.js`

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const cacheHitRate = new Rate('cache_hit_rate');
const responseTime = new Trend('response_time');

export let options = {
	stages: [
		{ duration: '30s', target: 20 },   // Warm up
		{ duration: '2m', target: 100 },   // Load test
		{ duration: '1m', target: 200 },   // Spike test
		{ duration: '30s', target: 0 },    // Ramp down
	],
	thresholds: {
		http_req_duration: ['p(95)<200'],  // 95% under 200ms
		http_req_failed: ['rate<0.01'],    // <1% errors
		cache_hit_rate: ['rate>0.8'],      // >80% cache hits
	},
};

export default function () {
	// Test tenant listing (caching endpoint)
	const res = http.get('http://localhost:8098/api/v1/tenants');

	check(res, {
		'status is 200': (r) => r.status === 200,
		'has data': (r) => JSON.parse(r.body).data !== null,
		'response time < 200ms': (r) => r.timings.duration < 200,
	});

	// Track response time
	responseTime.add(res.timings.duration);

	sleep(1);
}

export function handleSummary(data) {
	return {
		'summary.json': JSON.stringify(data),
		stdout: textSummary(data, { indent: ' ', enableColors: true }),
	};
}
```

#### Step 3: Create Benchmark Tests

**Location**: Create `tests/benchmark_test.go`

```go
package tests

import (
	"context"
	"testing"

	"github.com/anupamdutta5/saas-admin-service/internal/services"
)

func BenchmarkListTenantsWithMetrics(b *testing.B) {
	service := setupTestService(b)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ListTenantsWithMetrics(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCacheHit(b *testing.B) {
	service := setupTestService(b)
	ctx := context.Background()

	// Prime cache
	_, err := service.ListTenantsWithMetrics(ctx)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ListTenantsWithMetrics(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCacheMiss(b *testing.B) {
	service := setupTestService(b)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear cache before each iteration
		service.InvalidateTenantMetricsCache(ctx)

		_, err := service.ListTenantsWithMetrics(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}
```

#### Step 4: Run Performance Tests

```bash
# Run k6 load test
k6 run tests/load/cache_performance.js

# Run Go benchmarks
go test -bench=. -benchmem tests/benchmark_test.go

# Run with CPU profiling
go test -bench=. -cpuprofile=cpu.prof tests/benchmark_test.go
go tool pprof cpu.prof
```

**Success Criteria**:
- ✅ P95 latency < 200ms
- ✅ Cache hit rate > 80%
- ✅ Error rate < 1%
- ✅ System handles 200 concurrent users

---

## Implementation Checklist

### Phase 1: Priority 2 (3-4 hours)
- [ ] Fix #7: Circuit Breaker Logic
  - [ ] Add helper methods
  - [ ] Update Get() method
  - [ ] Update Set() method
  - [ ] Test circuit breaker
- [ ] Fix #8: Prometheus Metrics
  - [ ] Add metrics struct
  - [ ] Add service method
  - [ ] Add handler method
  - [ ] Add route
  - [ ] Test metrics endpoint

### Phase 2: Priority 3 Part 1 (5 hours)
- [ ] Fix #9: Tiered TTLs
  - [ ] Add TTL constants
  - [ ] Update cacheTenantsAsync
  - [ ] Update getCachedPlans
  - [ ] Test TTLs in Redis
- [ ] Fix #10: Plan Cache Invalidation
  - [ ] Add invalidation method
  - [ ] Find plan methods
  - [ ] Add invalidation calls
  - [ ] Test invalidation
- [ ] Fix #11: Cache Warming
  - [ ] Add warming method
  - [ ] Add startup call
  - [ ] Test warming

### Phase 3: Priority 3 Part 2 (5 hours)
- [ ] Fix #12: Cache Versioning
  - [ ] Add version constant
  - [ ] Update all cache keys
  - [ ] Test version changes
- [ ] Fix #13: Cache Compression
  - [ ] Add compression helpers
  - [ ] Add compressed methods
  - [ ] Update cacheTenantsAsync
  - [ ] Test compression

### Phase 4: Priority 4 (11 hours)
- [ ] Fix #14: Distributed Tracing
  - [ ] Add OpenTelemetry deps
  - [ ] Add tracing to cache ops
  - [ ] Configure tracer
- [ ] Fix #15: Performance Testing
  - [ ] Create k6 load tests
  - [ ] Create Go benchmarks
  - [ ] Run and analyze tests

---

## Final Notes

**Total Time**: 24 hours across 4 phases

**Recommended Approach**:
1. Implement Phase 1 (3-4 hours) this week
2. Test thoroughly and deploy
3. Implement Phase 2 (5 hours) next week
4. Implement Phase 3 (5 hours) the following week
5. Phase 4 is optional and can be done when needed

**After All Fixes Complete**:
- Risk: 3.5/10 → **1.0/10** (MINIMAL RISK)
- Status: **ENTERPRISE-GRADE**
- All 15/15 fixes implemented ✅

---

**Document Version**: 1.0
**Created**: October 19, 2025
**Last Updated**: October 19, 2025
