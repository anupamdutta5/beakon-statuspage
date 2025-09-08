package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthCheck represents a health check
type HealthCheck struct {
	Name        string                 `json:"name"`
	Status      HealthStatus           `json:"status"`
	Message     string                 `json:"message,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
	LastChecked time.Time              `json:"last_checked"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// HealthChecker defines the interface for health checks
type HealthChecker interface {
	Check(ctx context.Context) HealthCheck
}

// HealthRegistry manages health checks
type HealthRegistry struct {
	checks map[string]HealthChecker
	mu     sync.RWMutex
}

// NewHealthRegistry creates a new health registry
func NewHealthRegistry() *HealthRegistry {
	return &HealthRegistry{
		checks: make(map[string]HealthChecker),
	}
}

// Register adds a health check to the registry
func (hr *HealthRegistry) Register(name string, checker HealthChecker) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	hr.checks[name] = checker
}

// Unregister removes a health check from the registry
func (hr *HealthRegistry) Unregister(name string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	delete(hr.checks, name)
}

// GetHealth returns the health status of all registered checks
func (hr *HealthRegistry) GetHealth(ctx context.Context) map[string]HealthCheck {
	hr.mu.RLock()
	checks := make(map[string]HealthChecker, len(hr.checks))
	for name, checker := range hr.checks {
		checks[name] = checker
	}
	hr.mu.RUnlock()

	results := make(map[string]HealthCheck, len(checks))

	// Run checks concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, checker := range checks {
		wg.Add(1)
		go func(name string, checker HealthChecker) {
			defer wg.Done()

			checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			start := time.Now()
			healthCheck := checker.Check(checkCtx)
			healthCheck.Duration = time.Since(start)
			healthCheck.LastChecked = time.Now()

			mu.Lock()
			results[name] = healthCheck
			mu.Unlock()
		}(name, checker)
	}

	wg.Wait()
	return results
}

// GetOverallHealth returns the overall health status
func (hr *HealthRegistry) GetOverallHealth(ctx context.Context) HealthStatus {
	checks := hr.GetHealth(ctx)

	if len(checks) == 0 {
		return HealthStatusUnknown
	}

	healthyCount := 0
	unhealthyCount := 0
	degradedCount := 0

	for _, check := range checks {
		switch check.Status {
		case HealthStatusHealthy:
			healthyCount++
		case HealthStatusUnhealthy:
			unhealthyCount++
		case HealthStatusDegraded:
			degradedCount++
		}
	}

	// If any check is unhealthy, overall status is unhealthy
	if unhealthyCount > 0 {
		return HealthStatusUnhealthy
	}

	// If any check is degraded, overall status is degraded
	if degradedCount > 0 {
		return HealthStatusDegraded
	}

	// If all checks are healthy, overall status is healthy
	return HealthStatusHealthy
}

// HTTPHealthHandler provides HTTP endpoints for health checks
type HTTPHealthHandler struct {
	registry *HealthRegistry
}

// NewHTTPHealthHandler creates a new HTTP health handler
func NewHTTPHealthHandler(registry *HealthRegistry) *HTTPHealthHandler {
	return &HTTPHealthHandler{
		registry: registry,
	}
}

// HealthHandler handles health check requests
func (hh *HTTPHealthHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get overall health status
	overallHealth := hh.registry.GetOverallHealth(ctx)

	// Set appropriate HTTP status code
	var statusCode int
	switch overallHealth {
	case HealthStatusHealthy:
		statusCode = http.StatusOK
	case HealthStatusDegraded:
		statusCode = http.StatusOK // Still OK, but degraded
	case HealthStatusUnhealthy:
		statusCode = http.StatusServiceUnavailable
	default:
		statusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"status":    overallHealth,
		"timestamp": time.Now().UTC(),
		"service":   "statuspage",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DetailedHealthHandler handles detailed health check requests
func (hh *HTTPHealthHandler) DetailedHealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get detailed health status
	checks := hh.registry.GetHealth(ctx)
	overallHealth := hh.registry.GetOverallHealth(ctx)

	// Set appropriate HTTP status code
	var statusCode int
	switch overallHealth {
	case HealthStatusHealthy:
		statusCode = http.StatusOK
	case HealthStatusDegraded:
		statusCode = http.StatusOK // Still OK, but degraded
	case HealthStatusUnhealthy:
		statusCode = http.StatusServiceUnavailable
	default:
		statusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"status":    overallHealth,
		"timestamp": time.Now().UTC(),
		"service":   "statuspage",
		"checks":    checks,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ReadinessHandler handles readiness check requests
func (hh *HTTPHealthHandler) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check if service is ready to accept traffic
	overallHealth := hh.registry.GetOverallHealth(ctx)

	var statusCode int
	var ready bool

	switch overallHealth {
	case HealthStatusHealthy, HealthStatusDegraded:
		statusCode = http.StatusOK
		ready = true
	default:
		statusCode = http.StatusServiceUnavailable
		ready = false
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"ready":     ready,
		"timestamp": time.Now().UTC(),
		"service":   "statuspage",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// LivenessHandler handles liveness check requests
func (hh *HTTPHealthHandler) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	// Liveness check is simple - if the service is running, it's alive
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"alive":     true,
		"timestamp": time.Now().UTC(),
		"service":   "statuspage",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DatabaseHealthCheck checks database connectivity
type DatabaseHealthCheck struct {
	name    string
	checkFn func(ctx context.Context) error
	timeout time.Duration
}

// NewDatabaseHealthCheck creates a new database health check
func NewDatabaseHealthCheck(name string, checkFn func(ctx context.Context) error) *DatabaseHealthCheck {
	return &DatabaseHealthCheck{
		name:    name,
		checkFn: checkFn,
		timeout: 5 * time.Second,
	}
}

// Check performs the database health check
func (dhc *DatabaseHealthCheck) Check(ctx context.Context) HealthCheck {
	checkCtx, cancel := context.WithTimeout(ctx, dhc.timeout)
	defer cancel()

	err := dhc.checkFn(checkCtx)

	healthCheck := HealthCheck{
		Name:        dhc.name,
		LastChecked: time.Now(),
	}

	if err != nil {
		healthCheck.Status = HealthStatusUnhealthy
		healthCheck.Message = err.Error()
	} else {
		healthCheck.Status = HealthStatusHealthy
		healthCheck.Message = "Database connection is healthy"
	}

	return healthCheck
}

// ExternalServiceHealthCheck checks external service connectivity
type ExternalServiceHealthCheck struct {
	name    string
	url     string
	timeout time.Duration
	client  *http.Client
}

// NewExternalServiceHealthCheck creates a new external service health check
func NewExternalServiceHealthCheck(name, url string) *ExternalServiceHealthCheck {
	return &ExternalServiceHealthCheck{
		name:    name,
		url:     url,
		timeout: 5 * time.Second,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Check performs the external service health check
func (eshc *ExternalServiceHealthCheck) Check(ctx context.Context) HealthCheck {
	checkCtx, cancel := context.WithTimeout(ctx, eshc.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(checkCtx, "GET", eshc.url, nil)
	if err != nil {
		return HealthCheck{
			Name:        eshc.name,
			Status:      HealthStatusUnhealthy,
			Message:     fmt.Sprintf("Failed to create request: %v", err),
			LastChecked: time.Now(),
		}
	}

	resp, err := eshc.client.Do(req)
	if err != nil {
		return HealthCheck{
			Name:        eshc.name,
			Status:      HealthStatusUnhealthy,
			Message:     fmt.Sprintf("Request failed: %v", err),
			LastChecked: time.Now(),
		}
	}
	defer resp.Body.Close()

	healthCheck := HealthCheck{
		Name:        eshc.name,
		LastChecked: time.Now(),
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		healthCheck.Status = HealthStatusHealthy
		healthCheck.Message = "External service is healthy"
	} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		healthCheck.Status = HealthStatusDegraded
		healthCheck.Message = fmt.Sprintf("External service returned status %d", resp.StatusCode)
	} else {
		healthCheck.Status = HealthStatusUnhealthy
		healthCheck.Message = fmt.Sprintf("External service returned status %d", resp.StatusCode)
	}

	return healthCheck
}

// CustomHealthCheck allows for custom health check implementations
type CustomHealthCheck struct {
	name    string
	checkFn func(ctx context.Context) (HealthStatus, string, map[string]interface{})
	timeout time.Duration
}

// NewCustomHealthCheck creates a new custom health check
func NewCustomHealthCheck(name string, checkFn func(ctx context.Context) (HealthStatus, string, map[string]interface{})) *CustomHealthCheck {
	return &CustomHealthCheck{
		name:    name,
		checkFn: checkFn,
		timeout: 5 * time.Second,
	}
}

// Check performs the custom health check
func (chc *CustomHealthCheck) Check(ctx context.Context) HealthCheck {
	checkCtx, cancel := context.WithTimeout(ctx, chc.timeout)
	defer cancel()

	status, message, metadata := chc.checkFn(checkCtx)

	return HealthCheck{
		Name:        chc.name,
		Status:      status,
		Message:     message,
		LastChecked: time.Now(),
		Metadata:    metadata,
	}
}

// Global health registry instance
var globalHealthRegistry *HealthRegistry

// InitGlobalHealth initializes the global health registry
func InitGlobalHealth() {
	globalHealthRegistry = NewHealthRegistry()
}

// GetGlobalHealth returns the global health registry
func GetGlobalHealth() *HealthRegistry {
	return globalHealthRegistry
}

// RegisterHealthCheck is a convenience function for registering health checks
func RegisterHealthCheck(name string, checker HealthChecker) {
	if globalHealthRegistry != nil {
		globalHealthRegistry.Register(name, checker)
	}
}

// GetHealth is a convenience function for getting health status
func GetHealth(ctx context.Context) map[string]HealthCheck {
	if globalHealthRegistry != nil {
		return globalHealthRegistry.GetHealth(ctx)
	}
	return make(map[string]HealthCheck)
}

// GetOverallHealth is a convenience function for getting overall health status
func GetOverallHealth(ctx context.Context) HealthStatus {
	if globalHealthRegistry != nil {
		return globalHealthRegistry.GetOverallHealth(ctx)
	}
	return HealthStatusUnknown
}
