# All Services v2.0 Migration Plan

**Created**: 2025-11-04
**Status**: Ready to Execute
**Estimated Time**: 4-5 hours for remaining 18 services

---

## ✅ Completed: Monitoring-Service (Template)

The monitoring-service has been fully migrated and serves as our proven template. It now:
- Loads all config from YAML (no hardcoded defaults)
- Uses readable config struct with `SharedConfig` field
- Builds successfully
- Ready for deployment testing

---

## 📋 Remaining Services (18 Total)

### Priority Order (Complexity-Based)

**Tier 1: Complex Services (Similar to monitoring-service)**
1. **notification-service** - Has providers (Discord, Slack, Telegram, etc.) like monitoring-service
2. **tenant-admin-service** - Core service with RBAC, sessions, multi-tenant logic
3. **user-service** - Authentication, JWT, user management

**Tier 2: Medium Complexity Services**
4. **incident-service** - Incident management, notifications
5. **component-service** - Component status tracking
6. **payment-service** - Stripe integration, billing
7. **saas-admin-service** - Platform admin API
8. **analytics-service** - Metrics, data aggregation

**Tier 3: Simpler Services**
9. **branding-service** - Branding/theming
10. **landing-page-service** - Marketing pages
11. **status-ui-service** - Status page rendering
12. **event-store-service** - Event sourcing
13. **api-gateway** - Request routing (deprecated gateway pattern)
14. **database-service** - Database utilities (deprecated)

**Tier 4: Consumer Services (No HTTP server)**
15. **notification-consumer** - RabbitMQ consumer
16. **analytics-consumer** - RabbitMQ consumer
17. **audit-consumer** - RabbitMQ consumer
18. **billing-consumer** - RabbitMQ consumer

---

## 🎯 Migration Strategy

### Step-by-Step Process (Per Service)

#### Phase 1: Prepare (5 min)
1. Navigate to service directory
2. Read current `cmd/main.go` to understand structure
3. Identify service-specific config needs
4. Check for external integrations (webhooks, APIs)

#### Phase 2: Create Config Struct (5 min)
```go
// Standard config pattern for all services
type ServiceInfo struct {
    Name        string `yaml:"name"`
    Version     string `yaml:"version"`
    Environment string `yaml:"environment"`
}

type ServiceClientConfig struct {
    Timeout        time.Duration                   `yaml:"timeout"`
    CircuitBreaker resilience.CircuitBreakerConfig `yaml:"circuit_breaker"`
}

type [ServiceName]Config struct {
    SharedConfig  resilience.Config      `yaml:",inline"`
    Service       ServiceInfo            `yaml:"service"`
    ServiceClient ServiceClientConfig    `yaml:"service_client"`
    Retry         resilience.RetryConfig `yaml:"retry"`
}
```

#### Phase 3: Update main.go (10-15 min)

**Key changes:**
1. Config loading
2. Logger initialization
3. Prometheus registry
4. Metrics, Database, ServiceClient, RetryManager
5. Middleware and Server setup

#### Phase 4: Fix Service-Specific Issues (5-15 min)
- Constructor signature updates
- ServiceClient vs http.Client fixes
- Import errors
- Variable scoping

#### Phase 5: Build & Test (5 min)
- `go build ./cmd/main.go`
- Fix build errors
- Verify success

---

## 🔧 Common Patterns

### External Webhooks → Use http.Client
```go
client := &http.Client{Timeout: 10 * time.Second}
resp, err := client.Do(req)
```

### Internal Services → Use ServiceClient
```go
resp, err := serviceClient.Call(ctx, resilience.ServiceRequest{
    ServiceName: "tenant-admin-service",
    Method:      "POST",
    Path:        "/api/v1/tenants",
    Body:        requestData,
})
```

### Always Use cfg.SharedConfig
```go
cfg.SharedConfig.Database.Host      // ✅ Correct
cfg.SharedConfig.Server.Port        // ✅ Correct
cfg.Database.Host                   // ❌ Wrong
```

---

## 📊 Time Estimates

| Tier | Services | Time Each | Total |
|------|----------|-----------|-------|
| Tier 1 (Complex) | 3 | 30 min | 1.5 hrs |
| Tier 2 (Medium) | 5 | 20 min | 1.7 hrs |
| Tier 3 (Simple) | 6 | 15 min | 1.5 hrs |
| Tier 4 (Consumer) | 4 | 15 min | 1.0 hrs |
| **Total** | **18** | - | **5.7 hrs** |

**With testing**: ~6-7 hours

---

## 🚀 Execution Plan

### Start with: notification-service (Most similar to monitoring-service)

**Why notification-service first?**
- Has similar provider pattern (Discord, Slack, Telegram, etc.)
- Can reuse lessons from monitoring-service
- Complex enough to catch edge cases
- Important core service

---

## ✅ Success Criteria Per Service

1. Builds without errors
2. Config loaded from YAML
3. Readable config struct with SharedConfig
4. Injected Prometheus registry
5. Proper validation on startup
6. All tests pass (if exist)

---

---

## ✅ Completed Migrations

### 1. monitoring-service (100% Complete)
- Config struct with SharedConfig pattern
- All 40+ errors fixed holistically
- Builds successfully
- **Time taken**: ~2 hours

### 2. notification-service (100% Complete)
- Config struct with SharedConfig pattern
- Provider system updated (Manager, Slack, Teams, Webhook, Twilio SMS)
- All providers use http.Client for external APIs
- Builds successfully (39MB binary)
- **Time taken**: ~1 hour
- **Key learnings**: External webhooks must use http.Client, not ServiceClient

### 3. tenant-admin-service (100% Complete)
- Most complex service (795 lines)
- RBAC, sessions (Redis → DB → Memory), RabbitMQ events, SAML/SSO
- Used existing config package, updated middleware to use shared-resilience CORS/Security
- Builds successfully (37MB binary)
- **Time taken**: ~45 minutes
- **Key learnings**: Check middleware exports, use resilience.CORSMiddleware/SecurityHeadersMiddleware

### 4. user-service (100% Complete)
- Authentication and user management service
- JWT token handling, password reset/change
- Created new main.go using proven pattern
- Builds successfully (39MB binary)
- **Time taken**: ~15 minutes
- **Key learnings**: Simpler services migrate much faster with established pattern

---

**Progress**: 4/19 services complete (21.1%)

**Next Action**: Continue with incident-service and component-service
