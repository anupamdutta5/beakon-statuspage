// Package resilience provides resilience patterns for microservices
package resilience

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

// CircuitBreakerConfig represents configuration for circuit breaker
type CircuitBreakerConfig struct {
	Name         string        `yaml:"name" json:"name"`
	MaxRequests  uint32        `yaml:"max_requests" json:"max_requests" default:"3"`
	Interval     time.Duration `yaml:"interval" json:"interval" default:"10s"`
	Timeout      time.Duration `yaml:"timeout" json:"timeout" default:"60s"`
	FailureRatio float64       `yaml:"failure_ratio" json:"failure_ratio" default:"0.6"`
	MinRequests  uint32        `yaml:"min_requests" json:"min_requests" default:"5"`
	Enabled      bool          `yaml:"enabled" json:"enabled"`
}

// CircuitBreaker wraps gobreaker.CircuitBreaker with additional functionality
type CircuitBreaker struct {
	breaker *gobreaker.CircuitBreaker
	logger  *zap.Logger
	config  CircuitBreakerConfig
	metrics *CircuitBreakerMetrics
}

// CircuitBreakerMetrics tracks circuit breaker metrics
type CircuitBreakerMetrics struct {
	mu                 sync.RWMutex
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	CircuitOpenCount   int64
	LastFailureTime    time.Time
}

// NewCircuitBreaker creates a new circuit breaker instance
func NewCircuitBreaker(name string, config CircuitBreakerConfig, logger *zap.Logger) *CircuitBreaker {
	// Use the provided name or fall back to the config name
	if name != "" {
		config.Name = name
	}
	settings := gobreaker.Settings{
		Name:        config.Name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= config.MinRequests && failureRatio >= config.FailureRatio
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Info("Circuit breaker state changed",
				zap.String("name", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()))
		},
	}

	breaker := gobreaker.NewCircuitBreaker(settings)

	return &CircuitBreaker{
		breaker: breaker,
		logger:  logger,
		config:  config,
		metrics: &CircuitBreakerMetrics{},
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	cb.metrics.mu.Lock()
	cb.metrics.TotalRequests++
	cb.metrics.mu.Unlock()

	// Add timeout to context if not already set
	if _, hasTimeout := ctx.Deadline(); !hasTimeout {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cb.config.Timeout)
		defer cancel()
	}

	result, err := cb.breaker.Execute(func() (interface{}, error) {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		return fn()
	})

	cb.metrics.mu.Lock()
	if err != nil {
		cb.metrics.FailedRequests++
		cb.metrics.LastFailureTime = time.Now()

		// Check if circuit is open
		if cb.breaker.State() == gobreaker.StateOpen {
			cb.metrics.CircuitOpenCount++
		}
	} else {
		cb.metrics.SuccessfulRequests++
	}
	cb.metrics.mu.Unlock()

	if err != nil {
		cb.logger.Error("Circuit breaker execution failed",
			zap.String("circuit_breaker", cb.config.Name),
			zap.Error(err),
			zap.String("state", cb.breaker.State().String()))
		return nil, fmt.Errorf("circuit breaker %s: %w", cb.config.Name, err)
	}

	return result, nil
}

// ExecuteWithFallback executes a function with circuit breaker protection and fallback
func (cb *CircuitBreaker) ExecuteWithFallback(ctx context.Context, fn func() (interface{}, error), fallback func() (interface{}, error)) (interface{}, error) {
	result, err := cb.Execute(ctx, fn)
	if err != nil {
		cb.logger.Warn("Executing fallback function",
			zap.String("circuit_breaker", cb.config.Name),
			zap.Error(err))

		fallbackResult, fallbackErr := fallback()
		if fallbackErr != nil {
			return nil, fmt.Errorf("both primary and fallback failed: primary=%w, fallback=%w", err, fallbackErr)
		}
		return fallbackResult, nil
	}

	return result, nil
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() gobreaker.State {
	return cb.breaker.State()
}

// GetMetrics returns circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() CircuitBreakerMetrics {
	cb.metrics.mu.RLock()
	defer cb.metrics.mu.RUnlock()

	// Return a copy to avoid race conditions
	return CircuitBreakerMetrics{
		TotalRequests:      cb.metrics.TotalRequests,
		SuccessfulRequests: cb.metrics.SuccessfulRequests,
		FailedRequests:     cb.metrics.FailedRequests,
		CircuitOpenCount:   cb.metrics.CircuitOpenCount,
		LastFailureTime:    cb.metrics.LastFailureTime,
	}
}

// Reset resets the circuit breaker state
func (cb *CircuitBreaker) Reset() {
	// Note: gobreaker doesn't have a Reset method
	// The circuit breaker will automatically reset based on its configuration
	cb.logger.Info("Circuit breaker reset requested", zap.String("name", cb.config.Name))
}

// HTTPClientCircuitBreaker wraps HTTP client with circuit breaker
type HTTPClientCircuitBreaker struct {
	client  *http.Client
	breaker *CircuitBreaker
}

// NewHTTPClientCircuitBreaker creates a new HTTP client with circuit breaker
func NewHTTPClientCircuitBreaker(config CircuitBreakerConfig, logger *zap.Logger) *HTTPClientCircuitBreaker {
	return &HTTPClientCircuitBreaker{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		breaker: NewCircuitBreaker("http-client", config, logger),
	}
}

// Do executes HTTP request with circuit breaker protection
func (c *HTTPClientCircuitBreaker) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	result, err := c.breaker.Execute(ctx, func() (interface{}, error) {
		return c.client.Do(req)
	})

	if err != nil {
		return nil, err
	}

	return result.(*http.Response), nil
}

// GetState returns the current state of the circuit breaker
func (c *HTTPClientCircuitBreaker) GetState() gobreaker.State {
	return c.breaker.GetState()
}

// GetMetrics returns circuit breaker metrics
func (c *HTTPClientCircuitBreaker) GetMetrics() CircuitBreakerMetrics {
	return c.breaker.GetMetrics()
}
