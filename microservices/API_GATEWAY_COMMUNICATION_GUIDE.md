# API Gateway Communication Guide

## Overview

This document describes the **proper microservices communication pattern** for the Beakon platform. All service-to-service communication **MUST** go through the API Gateway to ensure consistent authentication, rate limiting, logging, and circuit breaking.

## Architecture

```
┌──────────────┐
│   Frontend   │
└──────┬───────┘
       │
       ▼
┌──────────────────────────────────────────────────────┐
│              API Gateway (Port 8080)                 │
│  ✓ Authentication    ✓ Rate Limiting                │
│  ✓ Circuit Breakers  ✓ Metrics Collection           │
│  ✓ Request Logging   ✓ CORS                         │
└───────┬──────────────────────────────────────────────┘
        │
        ├─────────────────┬─────────────────┬──────────────┐
        ▼                 ▼                 ▼              ▼
┌───────────────┐  ┌──────────────┐  ┌─────────────┐  ┌──────────────┐
│ User Service  │  │ Tenant Admin │  │  Component  │  │   Incident   │
│   (8081)      │  │   (8099)     │  │   (8084)    │  │   (8086)     │
└───────────────┘  └──────────────┘  └─────────────┘  └──────────────┘
```

## Port Reference

See `microservices/MICROSERVICES_PORT_REFERENCE.md` for complete port assignments.

**Key Services:**
- **API Gateway**: 8080 (Entry point for ALL clients)
- **User Service**: 8081
- **Component Service**: 8084
- **Notification Service**: 8085
- **Incident Service**: 8086
- **Payment Service**: 8088
- **Analytics Service**: 8090
- **Monitoring Service**: 8092
- **Status UI Service**: 8093
- **Database Service**: 8095 (⚠️ DEPRECATED - see Database Service README)
- **Event Store Service**: 8096
- **Branding Service**: 8097
- **SaaS Admin Service**: 8098
- **Tenant Admin Service**: 8099
- **Landing Page Service**: 8100

## Communication Rules

### ✅ CORRECT: Via API Gateway

**All services should call other services through the API Gateway:**

```go
// Example: SaaS Admin Service calling Tenant Admin Service
serviceURLs := config.GetServiceURLsFromEnv()

// Use the API Gateway URL
apiGatewayURL := serviceURLs.APIGatewayURL // http://localhost:8080

// Make request to /api/v1/tenants through gateway
req, _ := http.NewRequest("GET",
    fmt.Sprintf("%s/api/v1/tenants/%d", apiGatewayURL, tenantID),
    nil)

// Gateway routes to tenant-admin-service:8099 internally
resp, _ := httpClient.Do(req)
```

### ❌ INCORRECT: Direct Service-to-Service Calls

**DO NOT make direct calls to service ports:**

```go
// ❌ BAD: Direct call to tenant-admin-service
req, _ := http.NewRequest("GET",
    "http://localhost:8099/api/v1/tenants/123", // WRONG!
    nil)
```

## Configuration

### Environment Variables

All services should define `API_GATEWAY_URL`:

```bash
# .env
API_GATEWAY_URL=http://localhost:8080

# Optional: Direct service URLs (for fallback or special cases only)
TENANT_ADMIN_SERVICE_URL=http://localhost:8099
COMPONENT_SERVICE_URL=http://localhost:8084
INCIDENT_SERVICE_URL=http://localhost:8086
```

### Service Configuration Structs

All services have been updated with an `APIGatewayURL` field:

```go
type ServicesConfig struct {
    APIGatewayURL       string // Centralized gateway (RECOMMENDED)
    TenantAdminService  string // Direct URL (fallback only)
    ComponentService    string
    IncidentService     string
    // ... other services
}
```

## API Gateway Routes

The API Gateway routes requests based on URL patterns:

```
Frontend/Service Request          API Gateway Routes To
────────────────────────────────  ─────────────────────────────────
GET  /api/v1/users          →     user-service:8081
GET  /api/v1/tenants        →     tenant-admin-service:8099
GET  /api/v1/components     →     component-service:8084
POST /api/v1/incidents      →     incident-service:8086
GET  /api/v1/notifications  →     notification-service:8085
POST /api/v1/payments       →     payment-service:8088
GET  /api/v1/analytics      →     analytics-service:8090
GET  /api/v1/monitoring     →     monitoring-service:8092
```

## Benefits of API Gateway Pattern

### 1. **Centralized Authentication**
- JWT validation happens once at gateway
- Services don't need to verify tokens individually
- Consistent authentication across all services

### 2. **Rate Limiting**
- Prevent API abuse at the gateway level
- Per-tenant or per-user rate limiting
- Protects all downstream services

### 3. **Circuit Breakers**
- Gateway detects failing services
- Stops sending requests to unhealthy services
- Prevents cascading failures

### 4. **Observability**
- Centralized request/response logging
- Metrics collection (Prometheus)
- Distributed tracing
- Request ID tracking across services

### 5. **Security**
- Single entry point for security policies
- CORS configuration in one place
- Request validation
- SSL/TLS termination

### 6. **Simplified Client Code**
- Clients only need to know one URL
- No need to manage multiple service endpoints
- Easier service discovery

## Migration Guide

### For Services Making HTTP Calls

If your service currently makes direct HTTP calls to other services:

1. **Update configuration** to include `APIGatewayURL`:
   ```go
   Services: ServicesConfig{
       APIGatewayURL: getEnv("API_GATEWAY_URL", "http://localhost:8080"),
       // ... other services
   }
   ```

2. **Update HTTP clients** to use gateway URL:
   ```go
   // Before
   url := fmt.Sprintf("%s/api/v1/tenants/%d",
       config.TenantAdminService, tenantID)

   // After
   url := fmt.Sprintf("%s/api/v1/tenants/%d",
       config.APIGatewayURL, tenantID)
   ```

3. **Set environment variable**:
   ```bash
   export API_GATEWAY_URL=http://localhost:8080
   ```

### Services Already Updated

The following services have been updated to support API Gateway communication:

- ✅ **saas-admin-service** - Added `APIGatewayURL` field
- ✅ **tenant-admin-service** - Added `APIGatewayURL` field
- ✅ **landing-page-service** - Added `APIGatewayURL` field
- ✅ **status-ui-service** - Added `APIGatewayURL` field
- ✅ **api-gateway** - Port configurations corrected

## Production Deployment

### Load Balancer Configuration

In production, deploy multiple API Gateway instances behind a load balancer:

```
                    ┌─────────────────┐
                    │  Load Balancer  │
                    └────────┬────────┘
                             │
            ┌────────────────┼────────────────┐
            ▼                ▼                ▼
    ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
    │ API Gateway  │  │ API Gateway  │  │ API Gateway  │
    │  Instance 1  │  │  Instance 2  │  │  Instance 3  │
    └──────────────┘  └──────────────┘  └──────────────┘
```

### Environment Variables

```bash
# Production
API_GATEWAY_URL=https://api.beakon.io

# Staging
API_GATEWAY_URL=https://api.staging.beakon.io

# Development
API_GATEWAY_URL=http://localhost:8080
```

## Special Cases

### When Direct Calls Are Acceptable

Direct service-to-service calls may be acceptable ONLY in these scenarios:

1. **Internal Admin Operations**
   - SaaS Admin Service performing low-level operations
   - Database migration scripts
   - Health check systems

2. **High-Performance Requirements**
   - Real-time monitoring systems with <10ms latency requirements
   - Must be explicitly documented and approved

3. **Service Mesh / K8s Service Discovery**
   - When using Istio, Linkerd, or similar service mesh
   - Service discovery handles routing, auth, and observability

## Troubleshooting

### Connection Refused

```
Error: dial tcp 127.0.0.1:8080: connect: connection refused
```

**Solution**: Ensure API Gateway is running on port 8080

```bash
cd microservices/api-gateway
go run cmd/main.go
```

### Wrong Port Errors

```
Error: service returned 404
```

**Solution**: Check `MICROSERVICES_PORT_REFERENCE.md` for correct ports. API Gateway configuration may be out of sync.

### Authentication Failures

```
Error: 401 Unauthorized
```

**Solution**: Ensure JWT token is included in request headers:

```go
req.Header.Set("Authorization", "Bearer " + jwtToken)
```

## Monitoring

### API Gateway Metrics

Monitor these key metrics:

- **Request throughput** (requests/sec per service)
- **Response latency** (p50, p95, p99)
- **Error rates** (4xx, 5xx by service)
- **Circuit breaker state** (open/closed per service)
- **Rate limit hits** (throttled requests)

### Health Checks

```bash
# Check API Gateway health
curl http://localhost:8080/health

# Check circuit breaker status
curl http://localhost:8080/circuit-breakers
```

## References

- **Port Reference**: `microservices/MICROSERVICES_PORT_REFERENCE.md`
- **API Gateway README**: `microservices/api-gateway/README.md`
- **Database Service Status**: `microservices/database-service/README.md` (⚠️ DEPRECATED)

---

**Last Updated**: 2025-10-03
**Maintained By**: Platform Team
