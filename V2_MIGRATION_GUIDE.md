# V2 Migration Guide

## Overview
This guide documents the migration from shared-resilience V1 to V2 (completed October 2025). V2 introduces a "Primitives, Not Policies" design philosophy with breaking changes.

## Migration Status
✅ **ALL 19 ACTIVE SERVICES MIGRATED** (as of October 28, 2025)

## Key Changes in V2

### 1. Configuration Loading
**Before (V1):**
```go
// V1 had multiple ways to load config
cfg := config.LoadFromFile("config.yml")
cfg := config.LoadFromEnv()
```

**After (V2):**
```go
import resilience "github.com/anupamdutta5/shared-resilience"

// Unified configuration loading
loader := resilience.NewConfigLoader("configs")
var cfg Config
if err := loader.Load(&cfg); err != nil {
    log.Fatalf("Failed to load configuration: %v", err)
}
```

### 2. Database Connection
**Before (V1):**
```go
// V1 had connection logic in each service
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
```

**After (V2):**
```go
// V2 provides connection with automatic pooling
db, err := resilience.NewDatabaseConnection(config.Database)
```

### 3. Service Client (HTTP Calls)
**Before (V1):**
```go
// Manual HTTP client creation
client := &http.Client{Timeout: 30 * time.Second}
resp, err := client.Do(req)
```

**After (V2):**
```go
// ServiceClient with built-in circuit breakers
client := resilience.NewServiceClient("configs/service-endpoints.yml", logger)
response, err := client.Call(ctx, resilience.ServiceRequest{
    ServiceName: "tenant-admin-service",
    Method:      "POST",
    Path:        "/api/v1/tenants",
    Body:        requestData,
})
```

### 4. Health Checks
**Before (V1):**
```go
// Manual health endpoint implementation
router.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ok"})
})
```

**After (V2):**
```go
// Standardized health checker
healthChecker := resilience.NewHealthChecker()
healthChecker.AddCheck("database", dbHealthCheck)
healthChecker.AddCheck("redis", redisHealthCheck)

// Auto-registers /health, /health/live, /health/ready
router.GET("/health", healthChecker.Handler())
```

### 5. Circuit Breakers
**New in V2:**
```go
// Circuit breaker for external services
breaker := resilience.NewCircuitBreaker("payment-gateway", resilience.CircuitBreakerConfig{
    Threshold:   5,
    Timeout:     60 * time.Second,
    MaxRequests: 100,
})

err := breaker.Execute(func() error {
    return callPaymentGateway()
})
```

## Migration Steps

### Step 1: Update Dependencies
```bash
cd microservices/<service-name>
go get github.com/anupamdutta5/shared-resilience@v2.0.0
go mod tidy
```

### Step 2: Update Configuration Loading
1. Move configs to `configs/` directory
2. Update main.go to use `NewConfigLoader`
3. Remove environment-specific loading logic

### Step 3: Update Database Connection
Replace manual connection with `resilience.NewDatabaseConnection()`

### Step 4: Update Service Calls
Replace HTTP clients with `resilience.NewServiceClient()`

### Step 5: Implement Health Checks
Use `resilience.NewHealthChecker()` for standardized health endpoints

### Step 6: Add Circuit Breakers
Wrap external calls with circuit breakers for resilience

## Common Issues & Solutions

### Issue: Config not found
**Solution:** Ensure configs are in `configs/` directory or `docker-deployment/<service>/configs/`

### Issue: Database connection pool exhausted
**Solution:** V2 auto-configures pools based on CPU. Check `DB_MAX_CONNECTIONS` env var.

### Issue: Circuit breaker always open
**Solution:** Check threshold settings and ensure downstream service is healthy

### Issue: Service discovery fails
**Solution:** Update `service-endpoints.yml` with correct service URLs

## Configuration Path Resolution

V2 automatically resolves configuration paths:

1. **Docker**: `/app/configs` (mounted volume)
2. **Local Dev**: `docker-deployment/<service>/configs/`
3. **Tests**: Creates temporary config with defaults
4. **Override**: Set `CONFIG_PATH` environment variable

## Service-Specific Notes

### monitoring-service
- Complex integration adapters updated
- All 12 notification integrations use ServiceClient
- SSL checker and health monitoring updated

### notification-service
- All providers (Slack, Teams, Discord, etc.) use ServiceClient
- Circuit breakers on all external integrations

### tenant-admin-service
- Session management updated with V2 patterns
- RBAC using new middleware structure

### saas-admin-service
- Tenant provisioning uses ServiceClient
- Plan management with circuit breakers

## Rollback Procedure

If rollback is needed:
1. Checkout V1 branch: `git checkout pre-v2-migration`
2. Restore V1 dependencies: `go get github.com/anupamdutta5/shared-resilience@v1.0.0`
3. Restore old config loading patterns
4. Deploy V1 version

## Monitoring Post-Migration

Key metrics to watch:
- Database connection pool usage
- Circuit breaker states
- Service call latencies
- Health check statuses

## Support

For V2 migration issues:
- Check service logs: `docker logs <service-name>`
- Verify health: `curl http://localhost:<port>/health`
- Review configuration: `cat docker-deployment/<service>/configs/config.yml`