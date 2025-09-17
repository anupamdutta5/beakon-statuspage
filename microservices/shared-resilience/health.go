package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// HealthCheck represents a health check function
type HealthCheck func(ctx context.Context) error

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Name      string        `json:"name"`
	Status    string        `json:"status"`
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// HealthManager manages health checks for services
type HealthManager struct {
	checks    map[string]HealthCheck
	cache     map[string]HealthCheckResult
	cacheTTL  time.Duration
	timeout   time.Duration
	logger    *zap.Logger
	mu        sync.RWMutex
	cacheMu   sync.RWMutex
}

// HealthConfig represents health check configuration
type HealthConfig struct {
	Timeout  time.Duration `yaml:"timeout" env:"HEALTH_CHECK_TIMEOUT" default:"5s"`
	CacheTTL time.Duration `yaml:"cache_ttl" env:"HEALTH_CHECK_CACHE_TTL" default:"10s"`
}

// NewHealthManager creates a new health check manager
func NewHealthManager(config HealthConfig, logger *zap.Logger) *HealthManager {
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}
	if config.CacheTTL == 0 {
		config.CacheTTL = 10 * time.Second
	}

	return &HealthManager{
		checks:   make(map[string]HealthCheck),
		cache:    make(map[string]HealthCheckResult),
		cacheTTL: config.CacheTTL,
		timeout:  config.Timeout,
		logger:   logger,
	}
}

// RegisterHealthCheck registers a health check
func (hm *HealthManager) RegisterHealthCheck(name string, check HealthCheck) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.checks[name] = check
	hm.logger.Info("Registered health check", zap.String("name", name))
}

// RunHealthCheck executes a single health check
func (hm *HealthManager) RunHealthCheck(ctx context.Context, name string) HealthCheckResult {
	// Check cache first
	if cachedResult := hm.getCachedResult(name); cachedResult != nil {
		return *cachedResult
	}

	hm.mu.RLock()
	check, exists := hm.checks[name]
	hm.mu.RUnlock()

	if !exists {
		return HealthCheckResult{
			Name:      name,
			Status:    "error",
			Error:     "health check not found",
			Timestamp: time.Now(),
		}
	}

	start := time.Now()

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, hm.timeout)
	defer cancel()

	err := check(timeoutCtx)
	duration := time.Since(start)

	result := HealthCheckResult{
		Name:      name,
		Duration:  duration,
		Timestamp: time.Now(),
	}

	if err != nil {
		result.Status = "unhealthy"
		result.Error = err.Error()
		hm.logger.Warn("Health check failed",
			zap.String("name", name),
			zap.Duration("duration", duration),
			zap.Error(err))
	} else {
		result.Status = "healthy"
		hm.logger.Debug("Health check passed",
			zap.String("name", name),
			zap.Duration("duration", duration))
	}

	// Cache the result
	hm.cacheResult(name, result)

	return result
}

// RunAllHealthChecks executes all registered health checks
func (hm *HealthManager) RunAllHealthChecks(ctx context.Context) map[string]HealthCheckResult {
	hm.mu.RLock()
	checkNames := make([]string, 0, len(hm.checks))
	for name := range hm.checks {
		checkNames = append(checkNames, name)
	}
	hm.mu.RUnlock()

	results := make(map[string]HealthCheckResult)
	var wg sync.WaitGroup
	resultCh := make(chan HealthCheckResult, len(checkNames))

	// Run all checks concurrently
	for _, name := range checkNames {
		wg.Add(1)
		go func(checkName string) {
			defer wg.Done()
			result := hm.RunHealthCheck(ctx, checkName)
			resultCh <- result
		}(name)
	}

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	for result := range resultCh {
		results[result.Name] = result
	}

	return results
}

// GetOverallHealth returns overall health status
func (hm *HealthManager) GetOverallHealth(ctx context.Context) map[string]interface{} {
	results := hm.RunAllHealthChecks(ctx)

	overallStatus := "healthy"
	var unhealthyChecks []string

	for name, result := range results {
		if result.Status != "healthy" {
			overallStatus = "unhealthy"
			unhealthyChecks = append(unhealthyChecks, name)
		}
	}

	health := map[string]interface{}{
		"status":    overallStatus,
		"timestamp": time.Now(),
		"checks":    results,
	}

	if len(unhealthyChecks) > 0 {
		health["unhealthy_checks"] = unhealthyChecks
	}

	return health
}

// getCachedResult retrieves a cached health check result
func (hm *HealthManager) getCachedResult(name string) *HealthCheckResult {
	hm.cacheMu.RLock()
	defer hm.cacheMu.RUnlock()

	if result, exists := hm.cache[name]; exists {
		if time.Since(result.Timestamp) < hm.cacheTTL {
			return &result
		}
	}

	return nil
}

// cacheResult caches a health check result
func (hm *HealthManager) cacheResult(name string, result HealthCheckResult) {
	hm.cacheMu.Lock()
	defer hm.cacheMu.Unlock()

	hm.cache[name] = result
}

// Common health check implementations

// DatabaseHealthCheck creates a health check for database connectivity
func DatabaseHealthCheck(db interface {
	PingContext(ctx context.Context) error
}) HealthCheck {
	return func(ctx context.Context) error {
		return db.PingContext(ctx)
	}
}

// HTTPHealthCheck creates a health check for HTTP endpoints
func HTTPHealthCheck(url string) HealthCheck {
	return func(ctx context.Context) error {
		// This is a simplified implementation
		// In practice, you'd use http.Client with context
		return fmt.Errorf("HTTP health check not implemented")
	}
}

// CacheHealthCheck creates a health check for cache connectivity
func CacheHealthCheck(cache interface {
	Ping(ctx context.Context) error
}) HealthCheck {
	return func(ctx context.Context) error {
		return cache.Ping(ctx)
	}
}

// CustomHealthCheck creates a health check from a custom function
func CustomHealthCheck(checkFn func(ctx context.Context) error) HealthCheck {
	return checkFn
}

// FileSystemHealthCheck checks if a file system path is accessible
func FileSystemHealthCheck(path string) HealthCheck {
	return func(ctx context.Context) error {
		// Check if path exists and is accessible
		// This is a placeholder implementation
		return nil
	}
}

// MemoryHealthCheck checks if memory usage is within acceptable limits
func MemoryHealthCheck(maxMemoryPercent float64) HealthCheck {
	return func(ctx context.Context) error {
		// Check memory usage
		// This is a placeholder implementation
		return nil
	}
}