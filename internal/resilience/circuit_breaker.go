package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	// StateClosed - Circuit is closed, requests are allowed
	StateClosed CircuitState = iota
	// StateOpen - Circuit is open, requests are blocked
	StateOpen
	// StateHalfOpen - Circuit is half-open, limited requests are allowed
	StateHalfOpen
)

// String returns the string representation of the circuit state
func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig represents the configuration for a circuit breaker
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of failures before opening the circuit
	FailureThreshold int
	// SuccessThreshold is the number of successes needed to close the circuit from half-open
	SuccessThreshold int
	// Timeout is the duration to wait before transitioning from open to half-open
	Timeout time.Duration
	// MaxRequests is the maximum number of requests allowed in half-open state
	MaxRequests int
}

// DefaultCircuitBreakerConfig returns a default circuit breaker configuration
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          30 * time.Second,
		MaxRequests:      3,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config        *CircuitBreakerConfig
	state         CircuitState
	failures      int
	successes     int
	requests      int
	lastFailure   time.Time
	mutex         sync.RWMutex
	onStateChange func(from, to CircuitState)
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	// Check if we can execute the function
	if !cb.canExecute() {
		return fmt.Errorf("circuit breaker is %s", cb.state)
	}

	// Execute the function
	err := fn()

	// Record the result
	cb.recordResult(err)

	return err
}

// ExecuteWithFallback executes a function with circuit breaker protection and fallback
func (cb *CircuitBreaker) ExecuteWithFallback(ctx context.Context, fn func() error, fallback func() error) error {
	// Check if we can execute the function
	if !cb.canExecute() {
		// Execute fallback if circuit is open
		if cb.state == StateOpen {
			return fallback()
		}
		return fmt.Errorf("circuit breaker is %s", cb.state)
	}

	// Execute the function
	err := fn()

	// Record the result
	cb.recordResult(err)

	// If function failed and we have a fallback, execute it
	if err != nil && fallback != nil {
		return fallback()
	}

	return err
}

// ExecuteWithTimeout executes a function with circuit breaker protection and timeout
func (cb *CircuitBreaker) ExecuteWithTimeout(ctx context.Context, timeout time.Duration, fn func() error) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute with circuit breaker protection
	return cb.Execute(ctx, fn)
}

// canExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) canExecute() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if timeout has passed
		if time.Since(cb.lastFailure) >= cb.config.Timeout {
			cb.transitionTo(StateHalfOpen)
			return true
		}
		return false
	case StateHalfOpen:
		// Check if we've reached the max requests limit
		if cb.requests >= cb.config.MaxRequests {
			return false
		}
		return true
	default:
		return false
	}
}

// recordResult records the result of a function execution
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}
}

// recordFailure records a failure
func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.FailureThreshold {
			cb.transitionTo(StateOpen)
		}
	case StateHalfOpen:
		cb.transitionTo(StateOpen)
	}
}

// recordSuccess records a success
func (cb *CircuitBreaker) recordSuccess() {
	cb.successes++

	switch cb.state {
	case StateClosed:
		// Reset failure count on success
		cb.failures = 0
	case StateHalfOpen:
		if cb.successes >= cb.config.SuccessThreshold {
			cb.transitionTo(StateClosed)
		}
	}
}

// transitionTo transitions the circuit breaker to a new state
func (cb *CircuitBreaker) transitionTo(newState CircuitState) {
	oldState := cb.state
	cb.state = newState

	// Reset counters based on new state
	switch newState {
	case StateClosed:
		cb.failures = 0
		cb.successes = 0
		cb.requests = 0
	case StateOpen:
		cb.successes = 0
		cb.requests = 0
	case StateHalfOpen:
		cb.failures = 0
		cb.successes = 0
		cb.requests = 0
	}

	// Call state change callback
	if cb.onStateChange != nil {
		cb.onStateChange(oldState, newState)
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetStats returns statistics about the circuit breaker
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return map[string]interface{}{
		"state":        cb.state.String(),
		"failures":     cb.failures,
		"successes":    cb.successes,
		"requests":     cb.requests,
		"last_failure": cb.lastFailure,
	}
}

// SetStateChangeCallback sets a callback for state changes
func (cb *CircuitBreaker) SetStateChangeCallback(callback func(from, to CircuitState)) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.onStateChange = callback
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.transitionTo(StateClosed)
}

// IsOpen returns true if the circuit breaker is open
func (cb *CircuitBreaker) IsOpen() bool {
	return cb.GetState() == StateOpen
}

// IsClosed returns true if the circuit breaker is closed
func (cb *CircuitBreaker) IsClosed() bool {
	return cb.GetState() == StateClosed
}

// IsHalfOpen returns true if the circuit breaker is half-open
func (cb *CircuitBreaker) IsHalfOpen() bool {
	return cb.GetState() == StateHalfOpen
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mutex    sync.RWMutex
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// GetCircuitBreaker gets or creates a circuit breaker for a service
func (cbm *CircuitBreakerManager) GetCircuitBreaker(serviceName string, config *CircuitBreakerConfig) *CircuitBreaker {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	if breaker, exists := cbm.breakers[serviceName]; exists {
		return breaker
	}

	breaker := NewCircuitBreaker(config)
	cbm.breakers[serviceName] = breaker
	return breaker
}

// GetCircuitBreakerWithDefaults gets or creates a circuit breaker with default config
func (cbm *CircuitBreakerManager) GetCircuitBreakerWithDefaults(serviceName string) *CircuitBreaker {
	return cbm.GetCircuitBreaker(serviceName, DefaultCircuitBreakerConfig())
}

// RemoveCircuitBreaker removes a circuit breaker
func (cbm *CircuitBreakerManager) RemoveCircuitBreaker(serviceName string) {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()
	delete(cbm.breakers, serviceName)
}

// GetAllStats returns statistics for all circuit breakers
func (cbm *CircuitBreakerManager) GetAllStats() map[string]map[string]interface{} {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	stats := make(map[string]map[string]interface{})
	for name, breaker := range cbm.breakers {
		stats[name] = breaker.GetStats()
	}
	return stats
}

// ResetAll resets all circuit breakers
func (cbm *CircuitBreakerManager) ResetAll() {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	for _, breaker := range cbm.breakers {
		breaker.Reset()
	}
}
