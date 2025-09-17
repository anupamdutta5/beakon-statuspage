# Shared Resilience Package

## 🎯 **Purpose**

This package provides common resilience patterns that can be shared across microservices to ensure consistency, maintainability, and best practices.

## 📦 **Package Structure**

```
shared-resilience/
├── circuit_breaker.go      # Circuit breaker pattern
├── input_sanitization.go   # Security middleware
├── cache.go               # Caching strategies
├── rate_limiter.go        # Rate limiting implementation
├── retry.go              # Retry mechanisms with exponential backoff
├── shutdown.go           # Graceful shutdown management
├── validation.go         # Input validation utilities
├── health.go            # Health check framework
├── database.go          # Database connection management
├── security.go          # Security utilities
├── middleware.go        # HTTP middleware collection
├── config.go           # Configuration management
├── errors.go           # Error handling utilities
├── helpers.go          # Common helper functions
└── README.md           # This file
```

## 🏗️ **Architecture Decision**

### **Shared Package Benefits:**
- ✅ **Consistency** across all services
- ✅ **Single source of truth** for patterns
- ✅ **Easier maintenance** and updates
- ✅ **Standardized configuration**
- ✅ **Reduced code duplication**

### **Implementation Strategy:**

#### **Option 1: Shared Library (Recommended)**
```go
// Each service imports the shared package
import "github.com/anupamdutta5/statuspage-shared-resilience"

// Usage in service
circuitBreaker := resilience.NewCircuitBreaker(config, logger)
sanitizer := resilience.NewInputSanitizer(config, logger)
cache := resilience.NewInMemoryCache(config, logger)
```

#### **Option 2: Service-Specific Implementation**
```go
// Each service has its own resilience package
// internal/resilience/circuit_breaker.go
// internal/resilience/cache.go
```

#### **Option 3: Hybrid Approach (Best Practice)**
- **Core patterns** in shared package (circuit breaker, sanitization)
- **Service-specific** implementations for specialized needs
- **Configuration** driven by environment variables

## 🚀 **Usage Examples**

### **Circuit Breaker**
```go
import "github.com/anupamdutta5/statuspage-shared-resilience"

// Create circuit breaker
cb := resilience.NewCircuitBreaker(resilience.CircuitBreakerConfig{
    Name:         "external-api",
    MaxRequests:  3,
    Interval:     10 * time.Second,
    Timeout:      60 * time.Second,
    FailureRatio: 0.6,
}, logger)

// Use in service
result, err := cb.Execute(ctx, func() (interface{}, error) {
    return externalAPICall()
})
```

### **Input Sanitization**
```go
// Add middleware to Gin router
sanitizer := resilience.NewInputSanitizer(config, logger)
router.Use(sanitizer.GinMiddleware())
router.Use(resilience.SecurityHeadersMiddleware())
```

### **Caching**
```go
// Create cache
cache := resilience.NewInMemoryCache(resilience.CacheConfig{
    DefaultTTL:     5 * time.Minute,
    MaxSize:        1000,
    CleanupInterval: 10 * time.Minute,
}, logger)

// Use in service
value, err := cache.GetOrSet(ctx, "key", func() (interface{}, error) {
    return expensiveOperation()
}, 5*time.Minute)
```

## 🔧 **Configuration**

### **Environment Variables**
```bash
# Circuit Breaker
CIRCUIT_BREAKER_MAX_REQUESTS=3
CIRCUIT_BREAKER_INTERVAL=10s
CIRCUIT_BREAKER_TIMEOUT=60s
CIRCUIT_BREAKER_FAILURE_RATIO=0.6

# Caching
CACHE_DEFAULT_TTL=5m
CACHE_MAX_SIZE=1000
CACHE_CLEANUP_INTERVAL=10m

# Security
SANITIZATION_MAX_STRING_LENGTH=1000
SANITIZATION_STRICT_MODE=true
RATE_LIMIT_MAX_REQUESTS=100
RATE_LIMIT_WINDOW=1m
```

### **Service-Specific Configuration**
```yaml
# config.yaml
resilience:
  circuit_breaker:
    external_api:
      max_requests: 3
      interval: 10s
      timeout: 60s
  cache:
    default_ttl: 5m
    max_size: 1000
  security:
    rate_limit: 100
    sanitization:
      max_length: 1000
```

## 📊 **Monitoring & Metrics**

### **Circuit Breaker Metrics**
- Total requests
- Success/failure rates
- Circuit state changes
- Response times

### **Cache Metrics**
- Hit/miss rates
- Cache size
- Memory usage
- Eviction counts

### **Security Metrics**
- Rate limit violations
- Sanitization events
- Security header compliance

## 🚀 **Deployment Strategy**

### **Development**
```bash
# Use local shared package
go mod replace github.com/anupamdutta5/statuspage-shared-resilience => ./shared-resilience
```

### **Production**
```bash
# Use versioned package
go get github.com/anupamdutta5/statuspage-shared-resilience@v1.0.0
```

### **Docker**
```dockerfile
# Copy shared package
COPY shared-resilience /app/shared-resilience
WORKDIR /app
```

## 🔄 **Versioning Strategy**

### **Semantic Versioning**
- **Major**: Breaking changes
- **Minor**: New features, backward compatible
- **Patch**: Bug fixes, backward compatible

### **Migration Path**
1. **v1.0.0**: Initial implementation
2. **v1.1.0**: Add Redis caching
3. **v2.0.0**: Breaking changes (if needed)

## 🎯 **Best Practices**

### **1. Service Integration**
```go
// In each service's main.go
func setupResilience() {
    // Initialize shared resilience components
    circuitBreaker := resilience.NewCircuitBreaker(config, logger)
    sanitizer := resilience.NewInputSanitizer(config, logger)
    cache := resilience.NewInMemoryCache(config, logger)
    
    // Add middleware
    router.Use(sanitizer.GinMiddleware())
    router.Use(resilience.SecurityHeadersMiddleware())
    router.Use(resilience.RateLimitMiddleware(100, time.Minute))
}
```

### **2. Configuration Management**
```go
// Load configuration from environment
config := resilience.LoadConfigFromEnv()

// Override with service-specific settings
config.CircuitBreaker.Name = "user-service-external-api"
config.Cache.DefaultTTL = 10 * time.Minute
```

### **3. Error Handling**
```go
// Consistent error handling across services
if err != nil {
    logger.Error("Operation failed", 
        zap.String("service", "user-service"),
        zap.Error(err))
    return resilience.HandleError(err)
}
```

## 🔮 **Future Enhancements**

### **Implemented Features**
- [x] Circuit breaker pattern
- [x] Input sanitization and validation
- [x] In-memory and Redis caching
- [x] Rate limiting with memory/Redis backends
- [x] Retry mechanisms with exponential backoff
- [x] Graceful shutdown management
- [x] Health check framework
- [x] Database connection management
- [x] Security utilities and middleware
- [x] Configuration management

### **Planned Features**
- [ ] Bulkhead pattern implementation
- [ ] Distributed tracing integration
- [ ] Configuration hot-reloading
- [ ] Metrics collection and export
- [ ] Service mesh integration

### **Integration Points**
- [ ] Prometheus metrics
- [ ] Jaeger tracing
- [ ] Consul service discovery
- [ ] Vault secret management

## 📚 **Documentation**

- [Circuit Breaker Pattern](circuit_breaker.go)
- [Input Sanitization](input_sanitization.go)
- [Caching Strategies](cache.go)
- [Configuration Guide](docs/configuration.md)
- [Migration Guide](docs/migration.md)

## 🤝 **Contributing**

1. Follow Go best practices
2. Add comprehensive tests
3. Update documentation
4. Ensure backward compatibility
5. Add monitoring/metrics

## 📄 **License**

This package is part of the Status Page microservices architecture and follows the same licensing terms.




