# Direct Service Communication Guide

## Overview
As of October 26, 2025, the Beakon platform has deprecated the API Gateway in favor of direct service-to-service communication with built-in resilience patterns.

## Architecture Evolution

### Old Pattern (with API Gateway)
```
Client → API Gateway → Service A
                    → Service B
Service A → API Gateway → Service C
```

### New Pattern (Direct Communication)
```
Client → Service A (with auth middleware)
      → Service B (with auth middleware)
Service A → Service C (with circuit breaker)
```

## Key Components

### 1. ServiceClient (from shared-resilience)
Central HTTP client with automatic retries, circuit breakers, and health tracking.

```go
import resilience "github.com/anupamdutta5/shared-resilience"

// Initialize ServiceClient
client := resilience.NewServiceClient("configs/service-endpoints.yml", logger)

// Make service call
response, err := client.Call(ctx, resilience.ServiceRequest{
    ServiceName: "tenant-admin-service",
    Method:      "POST",
    Path:        "/api/v1/tenants",
    Headers:     headers,
    Body:        requestBody,
})
```

### 2. Service Endpoints Configuration
Each service defines its downstream dependencies in `configs/service-endpoints.yml`:

```yaml
endpoints:
  user-service:
    url: ${USER_SERVICE_URL:-http://user-service:8081}
    timeout: 30s
    retries: 3
    circuit_breaker:
      enabled: true
      threshold: 5
      timeout: 60s

  notification-service:
    url: ${NOTIFICATION_SERVICE_URL:-http://notification-service:8085}
    timeout: 30s
    retries: 3
    circuit_breaker:
      enabled: true
      threshold: 5
      timeout: 60s
```

### 3. Authentication Middleware
Each service implements JWT validation directly:

```go
// JWT middleware from shared-resilience
authMiddleware := resilience.NewJWTMiddleware(jwtSecret, logger)
router.Use(authMiddleware.Validate())

// Protected routes
protected := router.Group("/api/v1")
protected.Use(authMiddleware.Validate())
{
    protected.GET("/users", handlers.GetUsers)
    protected.POST("/users", handlers.CreateUser)
}
```

## Implementation Examples

### Example 1: Service A calling Service B

```go
func (s *ServiceA) CreateUserWithNotification(ctx context.Context, user User) error {
    // Step 1: Create user in user-service
    userResp, err := s.serviceClient.Call(ctx, resilience.ServiceRequest{
        ServiceName: "user-service",
        Method:      "POST",
        Path:        "/api/v1/users",
        Body:        user,
    })
    if err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }

    // Step 2: Send notification
    notifResp, err := s.serviceClient.Call(ctx, resilience.ServiceRequest{
        ServiceName: "notification-service",
        Method:      "POST",
        Path:        "/api/v1/notifications/welcome",
        Body:        map[string]interface{}{
            "user_id": userResp.Data["id"],
            "email":   user.Email,
        },
    })
    if err != nil {
        // Non-critical failure, log but don't fail
        s.logger.Warn("Failed to send welcome notification", zap.Error(err))
    }

    return nil
}
```

### Example 2: Handling Circuit Breaker States

```go
func (s *Service) CallExternalService(ctx context.Context) (*Response, error) {
    resp, err := s.serviceClient.Call(ctx, resilience.ServiceRequest{
        ServiceName: "external-api",
        Method:      "GET",
        Path:        "/api/data",
    })

    if err != nil {
        // Check if circuit breaker is open
        if errors.Is(err, resilience.ErrCircuitOpen) {
            // Return cached data or degraded response
            return s.getCachedResponse(), nil
        }
        return nil, err
    }

    // Cache successful response
    s.cacheResponse(resp)
    return resp, nil
}
```

### Example 3: Tenant Context Propagation

```go
func (s *Service) PropagateContext(ctx context.Context, tenantID string) {
    // Extract tenant context from JWT
    claims := ctx.Value("claims").(jwt.MapClaims)
    tenantID := claims["tenant_id"].(string)

    // Include in downstream calls
    headers := map[string]string{
        "X-Tenant-ID": tenantID,
        "Authorization": ctx.Value("authorization").(string),
    }

    resp, err := s.serviceClient.Call(ctx, resilience.ServiceRequest{
        ServiceName: "downstream-service",
        Method:      "GET",
        Path:        "/api/v1/resources",
        Headers:     headers,
    })
}
```

## Resilience Patterns

### 1. Retry Logic
Automatic retries with exponential backoff:
```yaml
retries: 3
retry_delay: 100ms
retry_max_delay: 5s
retry_multiplier: 2
```

### 2. Circuit Breaker
Protects against cascade failures:
```yaml
circuit_breaker:
  enabled: true
  threshold: 5        # Open after 5 failures
  timeout: 60s        # Try again after 60s
  max_requests: 100   # Max concurrent requests
```

### 3. Timeout Management
Per-service timeout configuration:
```yaml
timeout: 30s          # Overall request timeout
connect_timeout: 5s   # Connection establishment timeout
```

### 4. Load Balancing
When multiple instances available:
```yaml
endpoints:
  user-service:
    urls:  # Multiple URLs for load balancing
      - http://user-service-1:8081
      - http://user-service-2:8081
    load_balance: round_robin  # or random, least_conn
```

## Migration from API Gateway

### Step 1: Add Authentication Middleware
```go
// In main.go
authMiddleware := resilience.NewJWTMiddleware(cfg.JWT.Secret, logger)
router.Use(authMiddleware.Validate())
```

### Step 2: Configure Service Endpoints
Create `configs/service-endpoints.yml` with downstream services.

### Step 3: Replace HTTP Calls
```go
// Old (via API Gateway)
resp, err := http.Post("http://api-gateway:8080/user-service/api/v1/users", ...)

// New (direct)
resp, err := serviceClient.Call(ctx, resilience.ServiceRequest{
    ServiceName: "user-service",
    Method: "POST",
    Path: "/api/v1/users",
    Body: userData,
})
```

### Step 4: Update Docker Networking
Ensure services can reach each other directly:
```yaml
networks:
  beakon-network:
    driver: bridge
```

## Service Discovery

### Development Environment
Services use Docker service names:
- `http://user-service:8081`
- `http://notification-service:8085`

### Production Environment
Services use environment variables:
```bash
USER_SERVICE_URL=https://user-api.beakon.internal
NOTIFICATION_SERVICE_URL=https://notify-api.beakon.internal
```

### Kubernetes Environment
Services use Kubernetes DNS:
- `http://user-service.default.svc.cluster.local:8081`
- `http://notification-service.default.svc.cluster.local:8085`

## Monitoring & Observability

### Health Checks
Each service exposes health status of its dependencies:
```json
{
  "status": "healthy",
  "services": {
    "user-service": "healthy",
    "notification-service": "degraded",
    "database": "healthy"
  }
}
```

### Metrics
ServiceClient automatically tracks:
- Request count by service
- Response time percentiles
- Circuit breaker state changes
- Retry attempts

### Distributed Tracing
Include correlation IDs in all requests:
```go
headers := map[string]string{
    "X-Correlation-ID": uuid.New().String(),
}
```

## Troubleshooting

### Issue: Service unreachable
**Check:**
1. Service is running: `docker ps | grep service-name`
2. Network connectivity: `docker exec service-a ping service-b`
3. Port accessibility: `curl http://service-name:port/health`

### Issue: Circuit breaker always open
**Check:**
1. Downstream service health
2. Threshold settings in service-endpoints.yml
3. Reset circuit manually if needed

### Issue: High latency
**Check:**
1. Connection pool settings
2. Timeout configurations
3. Number of retries

### Issue: Authentication failures
**Check:**
1. JWT secret consistency across services
2. Token expiration
3. Clock synchronization

## Best Practices

1. **Always use ServiceClient** - Don't create raw HTTP clients
2. **Configure appropriate timeouts** - Prevent hanging requests
3. **Implement fallback strategies** - Handle circuit breaker opens gracefully
4. **Log correlation IDs** - Enable request tracing
5. **Monitor circuit breaker states** - Alert on persistent opens
6. **Cache responses** - Reduce downstream load
7. **Use health checks** - Verify dependencies on startup

## Security Considerations

1. **Internal network only** - Services should not be publicly accessible
2. **mTLS in production** - Mutual TLS for service-to-service
3. **Rotate JWT secrets** - Regular secret rotation
4. **Audit service calls** - Log all inter-service communication
5. **Rate limiting** - Prevent abuse even internally

## Performance Tuning

### Connection Pools
```yaml
connection_pool:
  max_idle: 100
  max_open: 200
  max_lifetime: 5m
```

### Circuit Breaker Tuning
```yaml
circuit_breaker:
  threshold: 10      # Higher for less critical services
  timeout: 30s       # Lower for fast recovery
  half_open_max: 5   # Requests in half-open state
```

### Retry Strategy
```yaml
retries: 3
retry_on: [500, 502, 503, 504]  # Only retry on server errors
retry_delay: 100ms
retry_max_delay: 2s  # Cap maximum delay
```