package resilience

import (
	"context"
	"fmt"
	"time"
)

// TimeoutConfig represents the configuration for timeout
type TimeoutConfig struct {
	// DefaultTimeout is the default timeout duration
	DefaultTimeout time.Duration
	// MaxTimeout is the maximum allowed timeout
	MaxTimeout time.Duration
	// MinTimeout is the minimum allowed timeout
	MinTimeout time.Duration
}

// DefaultTimeoutConfig returns a default timeout configuration
func DefaultTimeoutConfig() *TimeoutConfig {
	return &TimeoutConfig{
		DefaultTimeout: 30 * time.Second,
		MaxTimeout:     5 * time.Minute,
		MinTimeout:     1 * time.Second,
	}
}

// TimeoutManager manages timeouts for different operations
type TimeoutManager struct {
	config *TimeoutConfig
}

// NewTimeoutManager creates a new timeout manager
func NewTimeoutManager(config *TimeoutConfig) *TimeoutManager {
	if config == nil {
		config = DefaultTimeoutConfig()
	}

	return &TimeoutManager{
		config: config,
	}
}

// WithTimeout executes a function with a timeout
func (tm *TimeoutManager) WithTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	// Validate timeout
	if timeout < tm.config.MinTimeout {
		timeout = tm.config.MinTimeout
	}
	if timeout > tm.config.MaxTimeout {
		timeout = tm.config.MaxTimeout
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute function
	return fn(ctx)
}

// WithDefaultTimeout executes a function with the default timeout
func (tm *TimeoutManager) WithDefaultTimeout(ctx context.Context, fn func(context.Context) error) error {
	return tm.WithTimeout(ctx, tm.config.DefaultTimeout, fn)
}

// WithDeadline executes a function with a deadline
func (tm *TimeoutManager) WithDeadline(ctx context.Context, deadline time.Time, fn func(context.Context) error) error {
	// Create context with deadline
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	// Execute function
	return fn(ctx)
}

// TimeoutError represents a timeout error
type TimeoutError struct {
	Duration  time.Duration
	Operation string
}

// Error implements the error interface
func (e *TimeoutError) Error() string {
	return fmt.Sprintf("operation '%s' timed out after %v", e.Operation, e.Duration)
}

// IsTimeoutError checks if an error is a timeout error
func IsTimeoutError(err error) bool {
	_, ok := err.(*TimeoutError)
	return ok
}

// WithTimeoutAndError executes a function with timeout and custom error
func (tm *TimeoutManager) WithTimeoutAndError(ctx context.Context, timeout time.Duration, operation string, fn func(context.Context) error) error {
	// Validate timeout
	if timeout < tm.config.MinTimeout {
		timeout = tm.config.MinTimeout
	}
	if timeout > tm.config.MaxTimeout {
		timeout = tm.config.MaxTimeout
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create channel for result
	result := make(chan error, 1)

	// Execute function in goroutine
	go func() {
		result <- fn(ctx)
	}()

	// Wait for result or timeout
	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return &TimeoutError{
			Duration:  timeout,
			Operation: operation,
		}
	}
}

// WithTimeoutAndFallback executes a function with timeout and fallback
func (tm *TimeoutManager) WithTimeoutAndFallback(ctx context.Context, timeout time.Duration, fn func(context.Context) error, fallback func() error) error {
	// Validate timeout
	if timeout < tm.config.MinTimeout {
		timeout = tm.config.MinTimeout
	}
	if timeout > tm.config.MaxTimeout {
		timeout = tm.config.MaxTimeout
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create channel for result
	result := make(chan error, 1)

	// Execute function in goroutine
	go func() {
		result <- fn(ctx)
	}()

	// Wait for result or timeout
	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		// Execute fallback if timeout occurs
		if fallback != nil {
			return fallback()
		}
		return &TimeoutError{
			Duration:  timeout,
			Operation: "unknown",
		}
	}
}

// WithTimeoutAndRetry executes a function with timeout and retry
func (tm *TimeoutManager) WithTimeoutAndRetry(ctx context.Context, timeout time.Duration, retryConfig *RetryConfig, fn func(context.Context) error) error {
	// Validate timeout
	if timeout < tm.config.MinTimeout {
		timeout = tm.config.MinTimeout
	}
	if timeout > tm.config.MaxTimeout {
		timeout = tm.config.MaxTimeout
	}

	// Create retry function with timeout
	retryFn := func() error {
		return tm.WithTimeout(ctx, timeout, fn)
	}

	// Execute with retry
	return Retry(ctx, retryConfig, retryFn)
}

// WithTimeoutAndCircuitBreaker executes a function with timeout and circuit breaker
func (tm *TimeoutManager) WithTimeoutAndCircuitBreaker(ctx context.Context, timeout time.Duration, circuitBreaker *CircuitBreaker, fn func(context.Context) error) error {
	// Validate timeout
	if timeout < tm.config.MinTimeout {
		timeout = tm.config.MinTimeout
	}
	if timeout > tm.config.MaxTimeout {
		timeout = tm.config.MaxTimeout
	}

	// Create function with timeout
	timeoutFn := func() error {
		return tm.WithTimeout(ctx, timeout, fn)
	}

	// Execute with circuit breaker
	return circuitBreaker.Execute(ctx, timeoutFn)
}

// WithTimeoutAndAll executes a function with timeout, retry, and circuit breaker
func (tm *TimeoutManager) WithTimeoutAndAll(ctx context.Context, timeout time.Duration, retryConfig *RetryConfig, circuitBreaker *CircuitBreaker, fn func(context.Context) error) error {
	// Validate timeout
	if timeout < tm.config.MinTimeout {
		timeout = tm.config.MinTimeout
	}
	if timeout > tm.config.MaxTimeout {
		timeout = tm.config.MaxTimeout
	}

	// Create function with timeout
	timeoutFn := func() error {
		return tm.WithTimeout(ctx, timeout, fn)
	}

	// Create retry function
	retryFn := func() error {
		return Retry(ctx, retryConfig, timeoutFn)
	}

	// Execute with circuit breaker
	return circuitBreaker.Execute(ctx, retryFn)
}

// GetTimeoutForOperation returns the appropriate timeout for an operation
func (tm *TimeoutManager) GetTimeoutForOperation(operation string) time.Duration {
	// Define operation-specific timeouts
	timeouts := map[string]time.Duration{
		"database": 10 * time.Second,
		"http":     30 * time.Second,
		"grpc":     15 * time.Second,
		"redis":    5 * time.Second,
		"file":     60 * time.Second,
		"external": 45 * time.Second,
		"default":  tm.config.DefaultTimeout,
	}

	if timeout, exists := timeouts[operation]; exists {
		// Validate timeout
		if timeout < tm.config.MinTimeout {
			timeout = tm.config.MinTimeout
		}
		if timeout > tm.config.MaxTimeout {
			timeout = tm.config.MaxTimeout
		}
		return timeout
	}

	return tm.config.DefaultTimeout
}

// WithOperationTimeout executes a function with operation-specific timeout
func (tm *TimeoutManager) WithOperationTimeout(ctx context.Context, operation string, fn func(context.Context) error) error {
	timeout := tm.GetTimeoutForOperation(operation)
	return tm.WithTimeout(ctx, timeout, fn)
}

// TimeoutMiddleware creates a middleware for HTTP timeouts
func (tm *TimeoutManager) TimeoutMiddleware(timeout time.Duration) func(next func(context.Context) error) func(context.Context) error {
	return func(next func(context.Context) error) func(context.Context) error {
		return func(ctx context.Context) error {
			return tm.WithTimeout(ctx, timeout, next)
		}
	}
}

// TimeoutForService returns the appropriate timeout for a service
func (tm *TimeoutManager) TimeoutForService(serviceName string) time.Duration {
	// Define service-specific timeouts
	timeouts := map[string]time.Duration{
		"user-service":         10 * time.Second,
		"tenant-service":       10 * time.Second,
		"component-service":    15 * time.Second,
		"incident-service":     20 * time.Second,
		"notification-service": 30 * time.Second,
		"payment-service":      45 * time.Second,
		"analytics-service":    60 * time.Second,
		"monitoring-service":   15 * time.Second,
		"default":              tm.config.DefaultTimeout,
	}

	if timeout, exists := timeouts[serviceName]; exists {
		// Validate timeout
		if timeout < tm.config.MinTimeout {
			timeout = tm.config.MinTimeout
		}
		if timeout > tm.config.MaxTimeout {
			timeout = tm.config.MaxTimeout
		}
		return timeout
	}

	return tm.config.DefaultTimeout
}

// WithServiceTimeout executes a function with service-specific timeout
func (tm *TimeoutManager) WithServiceTimeout(ctx context.Context, serviceName string, fn func(context.Context) error) error {
	timeout := tm.TimeoutForService(serviceName)
	return tm.WithTimeout(ctx, timeout, fn)
}
