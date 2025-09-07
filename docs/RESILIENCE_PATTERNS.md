# Resilience Patterns Implementation

This document describes the resilience patterns implementation for the Status Page microservices architecture, including circuit breakers, retry logic, and timeout management.

## Overview

Resilience patterns help microservices handle failures gracefully and maintain system stability. This implementation provides circuit breakers, retry mechanisms, and timeout management to ensure robust service communication.

## Architecture

### Components

1. **Circuit Breaker** - Prevents cascade failures by opening the circuit when failures exceed threshold
2. **Retry Logic** - Automatically retries failed operations with configurable backoff strategies
3. **Timeout Management** - Ensures operations don't hang indefinitely
4. **Resilience Manager** - Coordinates all resilience patterns

### Resilience Flow

```
Request → Circuit Breaker → Retry Logic → Timeout → Service
   ↓           ↓              ↓           ↓         ↓
 Success    Open/Closed    Retry/Stop   Timeout   Response
```

## Implementation

### Circuit Breaker

The circuit breaker is implemented in `internal/resilience/circuit_breaker.go`:

#### States

- **Closed** - Normal operation, requests are allowed
- **Open** - Circuit is open, requests are blocked
- **Half-Open** - Limited requests are allowed to test if service is healthy

#### Configuration

```go
type CircuitBreakerConfig struct {
    FailureThreshold int           // Number of failures before opening
    SuccessThreshold int           // Successes needed to close from half-open
    Timeout          time.Duration // Time to wait before half-open
    MaxRequests      int           // Max requests in half-open state
}
```

#### Usage Example

```go
// Create circuit breaker
config := &CircuitBreakerConfig{
    FailureThreshold: 5,
    SuccessThreshold: 3,
    Timeout:          30 * time.Second,
    MaxRequests:      3,
}
breaker := NewCircuitBreaker(config)

// Execute with circuit breaker protection
err := breaker.Execute(ctx, func() error {
    return callExternalService()
})

// Execute with fallback
err := breaker.ExecuteWithFallback(ctx, 
    func() error { return callExternalService() },
    func() error { return callFallbackService() },
)
```

#### Circuit Breaker Manager

```go
// Create manager
manager := NewCircuitBreakerManager()

// Get circuit breaker for service
breaker := manager.GetCircuitBreakerWithDefaults("user-service")

// Get all statistics
stats := manager.GetAllStats()
```

### Retry Logic

The retry logic is implemented in `internal/resilience/retry.go`:

#### Configuration

```go
type RetryConfig struct {
    MaxAttempts  int           // Maximum retry attempts
    InitialDelay time.Duration // Initial delay between retries
    MaxDelay     time.Duration // Maximum delay between retries
    Multiplier   float64       // Exponential backoff multiplier
    Jitter       bool          // Add randomness to delay
}
```

#### Retry Strategies

1. **Exponential Backoff** - Delay increases exponentially
2. **Linear Backoff** - Delay increases linearly
3. **Fixed Delay** - Constant delay between retries
4. **Custom Delay** - Custom delay logic

#### Usage Examples

```go
// Basic retry with exponential backoff
config := &RetryConfig{
    MaxAttempts:  3,
    InitialDelay: 100 * time.Millisecond,
    MaxDelay:     5 * time.Second,
    Multiplier:   2.0,
    Jitter:       true,
}

err := Retry(ctx, config, func() error {
    return callExternalService()
})

// Retry with linear backoff
err := RetryWithLinearBackoff(ctx, config, func() error {
    return callExternalService()
})

// Retry with fixed delay
err := RetryWithFixedDelay(ctx, config, func() error {
    return callExternalService()
})

// Retry with custom condition
err := RetryWithCondition(ctx, config, 
    func(err error, attempt int) bool {
        return err != nil && attempt < 3
    },
    func() error {
        return callExternalService()
    },
)

// Retry with timeout
err := RetryWithTimeout(ctx, 30*time.Second, config, func() error {
    return callExternalService()
})

// Retry forever until success or context cancellation
err := RetryForever(ctx, config, func() error {
    return callExternalService()
})
```

#### Retryable Errors

```go
// Create retryable error with custom retry after
err := RetryAfterError(fmt.Errorf("service unavailable"), 5*time.Second)

// Retry with custom delay
err := RetryWithCustomDelay(ctx, config, func() error {
    return callExternalService()
})
```

### Timeout Management

The timeout management is implemented in `internal/resilience/timeout.go`:

#### Configuration

```go
type TimeoutConfig struct {
    DefaultTimeout time.Duration // Default timeout duration
    MaxTimeout     time.Duration // Maximum allowed timeout
    MinTimeout     time.Duration // Minimum allowed timeout
}
```

#### Usage Examples

```go
// Create timeout manager
config := &TimeoutConfig{
    DefaultTimeout: 30 * time.Second,
    MaxTimeout:     5 * time.Minute,
    MinTimeout:     1 * time.Second,
}
manager := NewTimeoutManager(config)

// Execute with timeout
err := manager.WithTimeout(ctx, 10*time.Second, func(ctx context.Context) error {
    return callExternalService(ctx)
})

// Execute with default timeout
err := manager.WithDefaultTimeout(ctx, func(ctx context.Context) error {
    return callExternalService(ctx)
})

// Execute with deadline
deadline := time.Now().Add(30 * time.Second)
err := manager.WithDeadline(ctx, deadline, func(ctx context.Context) error {
    return callExternalService(ctx)
})

// Execute with timeout and fallback
err := manager.WithTimeoutAndFallback(ctx, 10*time.Second,
    func(ctx context.Context) error {
        return callExternalService(ctx)
    },
    func() error {
        return callFallbackService()
    },
)

// Execute with operation-specific timeout
err := manager.WithOperationTimeout(ctx, "database", func(ctx context.Context) error {
    return callDatabase(ctx)
})

// Execute with service-specific timeout
err := manager.WithServiceTimeout(ctx, "user-service", func(ctx context.Context) error {
    return callUserService(ctx)
})
```

#### Timeout Error Handling

```go
// Check if error is timeout error
if IsTimeoutError(err) {
    // Handle timeout
    log.Printf("Operation timed out: %v", err)
}

// Custom timeout error
timeoutErr := &TimeoutError{
    Duration:  10 * time.Second,
    Operation: "user.create",
}
```

## Combined Patterns

### Circuit Breaker + Retry + Timeout

```go
// Create all components
breaker := NewCircuitBreaker(DefaultCircuitBreakerConfig())
retryConfig := DefaultRetryConfig()
timeoutManager := NewTimeoutManager(DefaultTimeoutConfig())

// Execute with all resilience patterns
err := timeoutManager.WithTimeoutAndAll(ctx, 30*time.Second, retryConfig, breaker,
    func(ctx context.Context) error {
        return callExternalService(ctx)
    },
)
```

### HTTP Client with Resilience

```go
// Create resilient HTTP client
client := &http.Client{
    Timeout: 30 * time.Second,
}

// Create circuit breaker for HTTP calls
breaker := NewCircuitBreaker(DefaultCircuitBreakerConfig())

// Make HTTP request with resilience
err := breaker.Execute(ctx, func() error {
    resp, err := client.Get("http://api.example.com/users")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 500 {
        return fmt.Errorf("server error: %d", resp.StatusCode)
    }
    
    return nil
})
```

### gRPC Client with Resilience

```go
// Create gRPC connection
conn, err := grpc.Dial("user-service:8080", grpc.WithInsecure())
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

client := pb.NewUserServiceClient(conn)

// Create circuit breaker
breaker := NewCircuitBreaker(DefaultCircuitBreakerConfig())

// Make gRPC call with resilience
err := breaker.Execute(ctx, func() error {
    resp, err := client.GetUser(ctx, &pb.GetUserRequest{Id: "123"})
    if err != nil {
        return err
    }
    
    // Process response
    log.Printf("User: %v", resp.User)
    return nil
})
```

## Configuration

### Environment Variables

```bash
# Circuit Breaker Configuration
CIRCUIT_BREAKER_FAILURE_THRESHOLD=5
CIRCUIT_BREAKER_SUCCESS_THRESHOLD=3
CIRCUIT_BREAKER_TIMEOUT=30s
CIRCUIT_BREAKER_MAX_REQUESTS=3

# Retry Configuration
RETRY_MAX_ATTEMPTS=3
RETRY_INITIAL_DELAY=100ms
RETRY_MAX_DELAY=5s
RETRY_MULTIPLIER=2.0
RETRY_JITTER=true

# Timeout Configuration
TIMEOUT_DEFAULT=30s
TIMEOUT_MAX=5m
TIMEOUT_MIN=1s
```

### Service-Specific Configuration

```go
// User Service Configuration
userServiceConfig := &CircuitBreakerConfig{
    FailureThreshold: 3,
    SuccessThreshold: 2,
    Timeout:          20 * time.Second,
    MaxRequests:      2,
}

// Payment Service Configuration (more lenient)
paymentServiceConfig := &CircuitBreakerConfig{
    FailureThreshold: 10,
    SuccessThreshold: 5,
    Timeout:          60 * time.Second,
    MaxRequests:      5,
}

// Database Configuration (strict)
databaseConfig := &CircuitBreakerConfig{
    FailureThreshold: 2,
    SuccessThreshold: 1,
    Timeout:          10 * time.Second,
    MaxRequests:      1,
}
```

## Monitoring and Observability

### Metrics

- **Circuit Breaker State** - Current state of circuit breakers
- **Failure Rate** - Percentage of failed requests
- **Retry Count** - Number of retry attempts
- **Timeout Count** - Number of timeout occurrences
- **Response Time** - Average response time with resilience patterns

### Logging

```go
// Circuit breaker state change logging
breaker.SetStateChangeCallback(func(from, to CircuitState) {
    log.Printf("Circuit breaker state changed from %s to %s", from, to)
})

// Retry logging
err := RetryWithCondition(ctx, config, 
    func(err error, attempt int) bool {
        log.Printf("Retry attempt %d, error: %v", attempt, err)
        return err != nil && attempt < 3
    },
    func() error {
        return callExternalService()
    },
)
```

### Health Checks

```go
// Circuit breaker health check
func (cb *CircuitBreaker) IsHealthy() bool {
    return cb.GetState() != StateOpen
}

// Get circuit breaker statistics
stats := breaker.GetStats()
log.Printf("Circuit breaker stats: %+v", stats)

// Get all circuit breaker statistics
allStats := manager.GetAllStats()
for service, stats := range allStats {
    log.Printf("Service %s: %+v", service, stats)
}
```

## Best Practices

### Circuit Breaker

1. **Appropriate Thresholds** - Set failure thresholds based on service characteristics
2. **Timeout Configuration** - Configure timeouts based on service response times
3. **Fallback Strategies** - Always provide fallback mechanisms
4. **Monitoring** - Monitor circuit breaker states and transitions

### Retry Logic

1. **Exponential Backoff** - Use exponential backoff to avoid overwhelming services
2. **Jitter** - Add jitter to prevent thundering herd problems
3. **Max Attempts** - Set reasonable maximum retry attempts
4. **Idempotent Operations** - Only retry idempotent operations

### Timeout Management

1. **Service-Specific Timeouts** - Configure timeouts based on service characteristics
2. **Context Propagation** - Always propagate context for cancellation
3. **Graceful Degradation** - Provide fallback when timeouts occur
4. **Monitoring** - Monitor timeout occurrences and adjust accordingly

### Combined Patterns

1. **Order of Application** - Apply patterns in the correct order
2. **Configuration Tuning** - Tune configurations based on monitoring data
3. **Testing** - Test resilience patterns under failure conditions
4. **Documentation** - Document resilience patterns and configurations

## Testing

### Unit Tests

```go
func TestCircuitBreaker(t *testing.T) {
    config := &CircuitBreakerConfig{
        FailureThreshold: 2,
        SuccessThreshold: 1,
        Timeout:          100 * time.Millisecond,
        MaxRequests:      1,
    }
    
    breaker := NewCircuitBreaker(config)
    
    // Test successful execution
    err := breaker.Execute(context.Background(), func() error {
        return nil
    })
    assert.NoError(t, err)
    assert.Equal(t, StateClosed, breaker.GetState())
    
    // Test failure threshold
    for i := 0; i < 2; i++ {
        err := breaker.Execute(context.Background(), func() error {
            return fmt.Errorf("service error")
        })
        assert.Error(t, err)
    }
    
    assert.Equal(t, StateOpen, breaker.GetState())
}
```

### Integration Tests

```go
func TestResiliencePatterns(t *testing.T) {
    // Create test server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusInternalServerError)
    }))
    defer server.Close()
    
    // Create resilient client
    breaker := NewCircuitBreaker(DefaultCircuitBreakerConfig())
    retryConfig := DefaultRetryConfig()
    timeoutManager := NewTimeoutManager(DefaultTimeoutConfig())
    
    // Test with all patterns
    err := timeoutManager.WithTimeoutAndAll(context.Background(), 30*time.Second, retryConfig, breaker,
        func(ctx context.Context) error {
            resp, err := http.Get(server.URL)
            if err != nil {
                return err
            }
            defer resp.Body.Close()
            
            if resp.StatusCode >= 500 {
                return fmt.Errorf("server error: %d", resp.StatusCode)
            }
            
            return nil
        },
    )
    
    assert.Error(t, err)
}
```

## Deployment

### Docker

```dockerfile
# Add resilience configuration
ENV CIRCUIT_BREAKER_FAILURE_THRESHOLD=5
ENV CIRCUIT_BREAKER_SUCCESS_THRESHOLD=3
ENV CIRCUIT_BREAKER_TIMEOUT=30s
ENV RETRY_MAX_ATTEMPTS=3
ENV RETRY_INITIAL_DELAY=100ms
ENV TIMEOUT_DEFAULT=30s
```

### Kubernetes

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
spec:
  template:
    spec:
      containers:
      - name: user-service
        env:
        - name: CIRCUIT_BREAKER_FAILURE_THRESHOLD
          value: "5"
        - name: CIRCUIT_BREAKER_SUCCESS_THRESHOLD
          value: "3"
        - name: CIRCUIT_BREAKER_TIMEOUT
          value: "30s"
        - name: RETRY_MAX_ATTEMPTS
          value: "3"
        - name: RETRY_INITIAL_DELAY
          value: "100ms"
        - name: TIMEOUT_DEFAULT
          value: "30s"
```

## Conclusion

The resilience patterns implementation provides comprehensive failure handling for microservices:

- **Circuit Breakers** - Prevent cascade failures and provide fast failure
- **Retry Logic** - Handle transient failures with intelligent backoff
- **Timeout Management** - Ensure operations don't hang indefinitely
- **Combined Patterns** - Work together to provide robust service communication

This implementation follows industry best practices and provides a solid foundation for building resilient microservices that can handle failures gracefully and maintain system stability.
