package health

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Status represents health check status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// ComponentType represents different component types
type ComponentType string

const (
	ComponentDatabase ComponentType = "database"
	ComponentRedis    ComponentType = "redis"
	ComponentAPI      ComponentType = "api"
	ComponentQueue    ComponentType = "queue"
	ComponentCache    ComponentType = "cache"
	ComponentCustom   ComponentType = "custom"
)

// Check represents a health check
type Check struct {
	Name          string                 `json:"name"`
	Type          ComponentType          `json:"type"`
	Status        Status                 `json:"status"`
	Message       string                 `json:"message,omitempty"`
	ResponseTime  time.Duration          `json:"response_time_ms"`
	LastChecked   time.Time              `json:"last_checked"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Critical      bool                   `json:"critical"`
}

// HealthChecker manages health checks
type HealthChecker struct {
	checks        map[string]CheckFunc
	criticalChecks map[string]bool
	cache         *checkCache
	logger        *zap.Logger
	mu            sync.RWMutex
	config        Config
}

// CheckFunc is a function that performs a health check
type CheckFunc func(ctx context.Context) Check

// Config configures the health checker
type Config struct {
	Timeout         time.Duration
	CacheDuration   time.Duration
	DetailedErrors  bool
	IncludeMetadata bool
}

// checkCache caches health check results
type checkCache struct {
	results map[string]*cachedResult
	mu      sync.RWMutex
}

type cachedResult struct {
	check     Check
	expiresAt time.Time
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(logger *zap.Logger, config Config) *HealthChecker {
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}
	if config.CacheDuration == 0 {
		config.CacheDuration = 10 * time.Second
	}

	return &HealthChecker{
		checks:         make(map[string]CheckFunc),
		criticalChecks: make(map[string]bool),
		cache: &checkCache{
			results: make(map[string]*cachedResult),
		},
		logger: logger,
		config: config,
	}
}

// RegisterCheck registers a health check
func (hc *HealthChecker) RegisterCheck(name string, check CheckFunc, critical bool) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.checks[name] = check
	hc.criticalChecks[name] = critical

	hc.logger.Debug("Registered health check",
		zap.String("name", name),
		zap.Bool("critical", critical))
}

// RegisterDatabaseCheck registers a database health check
func (hc *HealthChecker) RegisterDatabaseCheck(name string, db *gorm.DB, critical bool) {
	check := func(ctx context.Context) Check {
		start := time.Now()

		// Try to get underlying SQL DB
		sqlDB, err := db.DB()
		if err != nil {
			return Check{
				Name:         name,
				Type:         ComponentDatabase,
				Status:       StatusUnhealthy,
				Message:      fmt.Sprintf("failed to get database connection: %v", err),
				ResponseTime: time.Since(start),
				LastChecked:  time.Now(),
				Critical:     critical,
			}
		}

		// Check connection
		if err := sqlDB.PingContext(ctx); err != nil {
			return Check{
				Name:         name,
				Type:         ComponentDatabase,
				Status:       StatusUnhealthy,
				Message:      fmt.Sprintf("database ping failed: %v", err),
				ResponseTime: time.Since(start),
				LastChecked:  time.Now(),
				Critical:     critical,
			}
		}

		// Get connection stats
		stats := sqlDB.Stats()

		// Determine status based on connection pool usage
		status := StatusHealthy
		var message string

		if float64(stats.InUse)/float64(stats.MaxOpenConnections) > 0.9 {
			status = StatusDegraded
			message = "connection pool usage above 90%"
		}

		metadata := map[string]interface{}{
			"open_connections": stats.OpenConnections,
			"in_use":          stats.InUse,
			"idle":            stats.Idle,
			"max_open":        stats.MaxOpenConnections,
			"wait_count":      stats.WaitCount,
			"wait_duration":   stats.WaitDuration.String(),
		}

		return Check{
			Name:         name,
			Type:         ComponentDatabase,
			Status:       status,
			Message:      message,
			ResponseTime: time.Since(start),
			LastChecked:  time.Now(),
			Metadata:     metadata,
			Critical:     critical,
		}
	}

	hc.RegisterCheck(name, check, critical)
}

// RegisterHTTPCheck registers an HTTP endpoint health check
func (hc *HealthChecker) RegisterHTTPCheck(name string, url string, critical bool) {
	check := func(ctx context.Context) Check {
		start := time.Now()

		client := &http.Client{
			Timeout: hc.config.Timeout,
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return Check{
				Name:         name,
				Type:         ComponentAPI,
				Status:       StatusUnhealthy,
				Message:      fmt.Sprintf("failed to create request: %v", err),
				ResponseTime: time.Since(start),
				LastChecked:  time.Now(),
				Critical:     critical,
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			return Check{
				Name:         name,
				Type:         ComponentAPI,
				Status:       StatusUnhealthy,
				Message:      fmt.Sprintf("request failed: %v", err),
				ResponseTime: time.Since(start),
				LastChecked:  time.Now(),
				Critical:     critical,
			}
		}
		defer resp.Body.Close()

		status := StatusHealthy
		var message string

		if resp.StatusCode >= 500 {
			status = StatusUnhealthy
			message = fmt.Sprintf("HTTP %d", resp.StatusCode)
		} else if resp.StatusCode >= 400 {
			status = StatusDegraded
			message = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}

		return Check{
			Name:         name,
			Type:         ComponentAPI,
			Status:       status,
			Message:      message,
			ResponseTime: time.Since(start),
			LastChecked:  time.Now(),
			Metadata: map[string]interface{}{
				"status_code": resp.StatusCode,
				"url":         url,
			},
			Critical: critical,
		}
	}

	hc.RegisterCheck(name, check, critical)
}

// CheckAll performs all health checks
func (hc *HealthChecker) CheckAll(ctx context.Context) HealthReport {
	hc.mu.RLock()
	checks := make(map[string]CheckFunc, len(hc.checks))
	for name, check := range hc.checks {
		checks[name] = check
	}
	hc.mu.RUnlock()

	var wg sync.WaitGroup
	results := make([]Check, 0, len(checks))
	resultsChan := make(chan Check, len(checks))

	// Run checks concurrently
	for name, checkFunc := range checks {
		wg.Add(1)
		go func(n string, cf CheckFunc) {
			defer wg.Done()

			// Check cache first
			if cached := hc.cache.get(n); cached != nil {
				resultsChan <- *cached
				return
			}

			// Run check with timeout
			checkCtx, cancel := context.WithTimeout(ctx, hc.config.Timeout)
			defer cancel()

			result := cf(checkCtx)

			// Cache result
			hc.cache.set(n, result, hc.config.CacheDuration)

			resultsChan <- result
		}(name, checkFunc)
	}

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		results = append(results, result)
	}

	// Calculate overall status
	overallStatus := hc.calculateOverallStatus(results)

	return HealthReport{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Checks:    results,
		Version:   "1.0.0", // You might want to get this from configuration
		Uptime:    getUptime(),
	}
}

// CheckLiveness checks if the service is alive
func (hc *HealthChecker) CheckLiveness(ctx context.Context) HealthReport {
	// Liveness check only includes critical checks
	hc.mu.RLock()
	criticalChecks := make(map[string]CheckFunc)
	for name, check := range hc.checks {
		if hc.criticalChecks[name] {
			criticalChecks[name] = check
		}
	}
	hc.mu.RUnlock()

	results := make([]Check, 0, len(criticalChecks))
	for _, checkFunc := range criticalChecks {
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second) // Shorter timeout for liveness
		result := checkFunc(checkCtx)
		cancel()
		results = append(results, result)
	}

	overallStatus := hc.calculateOverallStatus(results)

	return HealthReport{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Checks:    results,
	}
}

// CheckReadiness checks if the service is ready to serve traffic
func (hc *HealthChecker) CheckReadiness(ctx context.Context) HealthReport {
	// Readiness includes all checks
	return hc.CheckAll(ctx)
}

// calculateOverallStatus determines overall health status
func (hc *HealthChecker) calculateOverallStatus(checks []Check) Status {
	hasUnhealthy := false
	hasDegraded := false
	hasCriticalFailure := false

	for _, check := range checks {
		if check.Status == StatusUnhealthy {
			hasUnhealthy = true
			if check.Critical {
				hasCriticalFailure = true
			}
		} else if check.Status == StatusDegraded {
			hasDegraded = true
		}
	}

	if hasCriticalFailure {
		return StatusUnhealthy
	}
	if hasUnhealthy || hasDegraded {
		return StatusDegraded
	}
	return StatusHealthy
}

// HealthReport represents the complete health status
type HealthReport struct {
	Status    Status    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Checks    []Check   `json:"checks,omitempty"`
	Version   string    `json:"version,omitempty"`
	Uptime    string    `json:"uptime,omitempty"`
}

// GinHandlers returns Gin handlers for health endpoints
func (hc *HealthChecker) GinHandlers() (health, liveness, readiness gin.HandlerFunc) {
	health = func(c *gin.Context) {
		report := hc.CheckAll(c.Request.Context())

		statusCode := http.StatusOK
		if report.Status == StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, report)
	}

	liveness = func(c *gin.Context) {
		report := hc.CheckLiveness(c.Request.Context())

		statusCode := http.StatusOK
		if report.Status == StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		}

		// Simple response for Kubernetes
		c.JSON(statusCode, gin.H{
			"status": report.Status,
		})
	}

	readiness = func(c *gin.Context) {
		report := hc.CheckReadiness(c.Request.Context())

		statusCode := http.StatusOK
		if report.Status == StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		}

		// Simple response for Kubernetes
		c.JSON(statusCode, gin.H{
			"status": report.Status,
		})
	}

	return health, liveness, readiness
}

// Cache methods
func (c *checkCache) get(name string) *Check {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if result, exists := c.results[name]; exists {
		if time.Now().Before(result.expiresAt) {
			return &result.check
		}
	}
	return nil
}

func (c *checkCache) set(name string, check Check, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.results[name] = &cachedResult{
		check:     check,
		expiresAt: time.Now().Add(duration),
	}
}

// Helper functions

var startTime = time.Now()

func getUptime() string {
	return time.Since(startTime).Round(time.Second).String()
}

// MetricsCollector collects health check metrics
type MetricsCollector struct {
	mu            sync.RWMutex
	checkResults  map[string][]Check
	checkDuration map[string][]time.Duration
	maxHistory    int
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(maxHistory int) *MetricsCollector {
	return &MetricsCollector{
		checkResults:  make(map[string][]Check),
		checkDuration: make(map[string][]time.Duration),
		maxHistory:    maxHistory,
	}
}

// Record records a health check result
func (mc *MetricsCollector) Record(check Check) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Add to history
	mc.checkResults[check.Name] = append(mc.checkResults[check.Name], check)
	mc.checkDuration[check.Name] = append(mc.checkDuration[check.Name], check.ResponseTime)

	// Trim history
	if len(mc.checkResults[check.Name]) > mc.maxHistory {
		mc.checkResults[check.Name] = mc.checkResults[check.Name][1:]
		mc.checkDuration[check.Name] = mc.checkDuration[check.Name][1:]
	}
}

// GetMetrics returns collected metrics
func (mc *MetricsCollector) GetMetrics() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics := make(map[string]interface{})

	for name, results := range mc.checkResults {
		successCount := 0
		for _, result := range results {
			if result.Status == StatusHealthy {
				successCount++
			}
		}

		var avgDuration time.Duration
		if len(mc.checkDuration[name]) > 0 {
			var total time.Duration
			for _, d := range mc.checkDuration[name] {
				total += d
			}
			avgDuration = total / time.Duration(len(mc.checkDuration[name]))
		}

		metrics[name] = map[string]interface{}{
			"success_rate":      float64(successCount) / float64(len(results)),
			"avg_response_time": avgDuration.Milliseconds(),
			"total_checks":      len(results),
		}
	}

	return metrics
}