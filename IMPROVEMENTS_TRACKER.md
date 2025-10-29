# Beakon Platform - Improvements Tracker

**Purpose**: Track all identified issues and their implementation status
**Created**: 2025-10-29
**Status**: In Progress

---

## Summary Dashboard

| Category | Total Issues | P0 (Critical) | P1 (High) | P2 (Medium) | Fixed | In Progress | Pending |
|----------|--------------|---------------|-----------|-------------|-------|-------------|---------|
| **Circuit Breakers & Resilience** | 8 | 3 | 3 | 2 | 2 | 0 | 6 |
| **Event-Driven Architecture** | 4 | 2 | 1 | 1 | 1 | 0 | 3 |
| **Authentication & Security** | 3 | 0 | 2 | 1 | 1 | 0 | 2 |
| **Monitoring & Observability** | 5 | 0 | 3 | 2 | 2 | 0 | 3 |
| **Database & Consistency** | 3 | 1 | 1 | 1 | 2 | 0 | 1 |
| **Architecture / Technical Debt** | 3 | 1 | 1 | 1 | 0 | 0 | 3 |
| **Documentation** | 3 | 0 | 1 | 2 | 0 | 0 | 3 |
| **TOTAL** | **29** | **7** | **12** | **10** | **8** | **0** | **21** |

---

## Service Groups for Implementation

### Group 1: Admin Services (CURRENT FOCUS)
- saas-admin-service (8098)
- tenant-admin-service (8099)

### Group 2: Authentication
- user-service (8081)

### Group 3: Core Monitoring
- monitoring-service (8092)
- incident-service (8086)

### Group 4: Notification & Components
- component-service (8084)
- notification-service (8085)
- notification-consumer

### Group 5: Remaining Services
- analytics-service (8090)
- analytics-consumer
- payment-service (8088)
- billing-consumer
- branding-service (8097)
- event-store-service (8096)
- landing-page-service (8100)
- status-ui-service (8093)
- audit-consumer

---

## GROUP 1: Admin Services (saas-admin + tenant-admin)

### Issues

#### 1. RabbitMQ Path Doesn't Create Admin User

**Priority**: 🔴 P0 (Critical)
**Service**: saas-admin-service + tenant-admin-service
**Status**: ✅ COMPLETE (2025-01-29)
**Commit**: 17a05c1

**Problem**:
- When SaaS Admin creates tenant, RabbitMQ event is published
- Event does NOT contain admin_email/admin_password
- tenant-admin-service consumer creates tenant but NOT admin user
- Result: Tenant exists but admin cannot login

**Current Code**:
```go
// File: saas-admin-service/internal/handlers/saas_admin_handler.go:1325
event := events.NewTenantCreatedEvent(tenant, metadata)
// Event only contains: id, name, slug, contact_email, domain, subdomain, status, max_users
// Missing: admin_email, admin_password
```

```go
// File: tenant-admin-service/internal/events/handler.go:27
func (h *RabbitMQTenantEventHandler) HandleTenantCreated(event) error {
    tenant := h.eventDataToTenant(event.Data)
    h.db.Create(&tenant).Error  // ❌ Only creates tenant, NOT admin user
    return nil
}
```

**Solution Options**:
1. ✅ **Option A (RECOMMENDED)**: Add admin credentials to event (encrypted)
2. Option B: Separate event: tenant.admin.created
3. Option C: Always use HTTP for initial creation, RabbitMQ for updates only

**Implementation Plan**:
- [ ] Add admin_email and admin_password to TenantEvent schema
- [ ] Encrypt password in event (or use temporary token)
- [ ] Update RabbitMQTenantEventHandler.HandleTenantCreated() to create admin user
- [ ] Update event publisher to include admin credentials
- [ ] Test: Create tenant via RabbitMQ path, verify admin can login

**Files to Change**:
- `saas-admin-service/internal/events/types.go` - Add admin fields to event
- `saas-admin-service/internal/handlers/saas_admin_handler.go:1325` - Include admin in event
- `tenant-admin-service/internal/events/types.go` - Add admin fields to event
- `tenant-admin-service/internal/events/handler.go:27` - Create admin user
- `tenant-admin-service/internal/services/tenant_admin_service.go` - Add CreateAdminUser method

**Estimated Effort**: 4 hours
**Risk**: Low

---

#### 2. Missing Circuit Breakers on HTTP Calls

**Priority**: 🔴 P0 (Critical)
**Service**: saas-admin-service
**Status**: ✅ COMPLETE (2025-10-29)
**Commit**: ba3761c

**Problem**:
- saas-admin-service makes HTTP calls to tenant-admin-service
- Uses raw `http.Client.Do()` without circuit breaker
- No protection against cascade failures
- Violates architecture pattern (should use shared-resilience ServiceClient)

**Current Code**:
```go
// File: saas-admin-service/internal/handlers/saas_admin_handler.go:2077
resp, err := h.httpClient.Do(req)  // ❌ NO CIRCUIT BREAKER
```

**Expected Pattern**:
```go
// Should use shared-resilience ServiceClient
client := resilience.NewServiceClient("configs/service-endpoints.yml", logger)
response, err := client.Call(ctx, resilience.ServiceRequest{
    ServiceName: "tenant-admin-service",
    Method:      "POST",
    Path:        "/api/v1/public/tenants",
    Body:        requestData,
})
```

**Implementation Plan**:
- [ ] Create `service-endpoints.yml` configuration file
- [ ] Initialize ServiceClient in main.go
- [ ] Replace all `h.httpClient.Do()` calls with `serviceClient.Call()`
- [ ] Add circuit breaker configuration for tenant-admin-service
- [ ] Test circuit breaker triggers after N failures

**Files to Change**:
- `saas-admin-service/configs/service-endpoints.yml` - NEW FILE
- `saas-admin-service/cmd/main.go` - Initialize ServiceClient
- `saas-admin-service/internal/handlers/saas_admin_handler.go` - Replace all HTTP calls

**Affected Functions**:
- `createTenantInTenantAdminService()` (line 2051)
- `deleteTenantFromTenantAdminService()` (line 2096)
- `updateTenantCredentialsDirect()` - NO CHANGE (direct DB access)

**Estimated Effort**: 3 hours
**Risk**: Low

---

#### 3. No HTTP Retry Logic

**Priority**: 🔴 P0 (Critical)
**Service**: saas-admin-service
**Status**: ✅ COMPLETE (2025-10-29)
**Commit**: ba3761c (included in Issue #2)

**Problem**:
- HTTP fallback to tenant-admin-service has no retry logic
- Single attempt, permanent failure on network hiccup
- Should implement exponential backoff retries

**Implementation Plan**:
- [ ] Use shared-resilience ServiceClient (includes retry logic)
- [ ] Configure retry strategy: max 3 attempts, exponential backoff (100ms, 200ms, 400ms)
- [ ] Only retry on 5xx errors and network errors (not 4xx)
- [ ] Log each retry attempt

**Files to Change**:
- `saas-admin-service/configs/service-endpoints.yml` - Configure retries
- Already fixed by Issue #2 (using ServiceClient)

**Estimated Effort**: Included in Issue #2
**Risk**: Low

---

#### 4. Incomplete Tenant Admin Authentication

**Priority**: 🟡 P1 (High)
**Service**: tenant-admin-service
**Status**: ✅ COMPLETE (2025-10-29)
**Commit**: 20154b6

**Problem**:
- tenant-admin-service uses 24-hour JWTs (OLD pattern)
- Should use 15-minute JWTs + 7-day refresh tokens (like saas-admin-service)
- No refresh token implementation

**Current State**:
```go
// File: tenant-admin-service/internal/auth/jwt.go
// JWT expires in 24 hours (should be 15 minutes)
```

**Expected State** (from saas-admin-service):
```go
// Access Token: 15 minutes
// Refresh Token: 7 days, stored in user_sessions table
```

**Implementation Plan**:
- [ ] Copy JWT + Refresh token implementation from saas-admin-service
- [ ] Create `user_sessions` table in tenant_admin_db
- [ ] Update Login handler to return both tokens
- [ ] Create Refresh handler: POST /api/v1/auth/refresh
- [ ] Update frontend to handle token refresh

**Files to Change**:
- `tenant-admin-service/internal/auth/jwt.go` - Update JWT TTL
- `tenant-admin-service/internal/models/session.go` - Add UserSession model
- `tenant-admin-service/internal/handlers/tenant_admin_handler.go` - Update Login/Refresh
- `tenant-admin-service/migrations/` - Add user_sessions table

**Estimated Effort**: 6 hours
**Risk**: Medium (affects all logged-in users)

---

#### 5. No Sync Verification Endpoint

**Priority**: 🟡 P1 (High)
**Service**: saas-admin-service
**Status**: ✅ COMPLETE (2025-10-29)
**Commit**: b255c44

**Problem**:
- Cannot detect tenants in saas_admin but not in tenant_admin_db
- No way to verify sync status
- No automated reconciliation

**Implementation**:
- ✅ Created endpoint: GET /api/v1/tenants/verify-sync
- ✅ Queries tenants from saas_admin database
- ✅ Checks each tenant existence in tenant-admin-service via HTTP
- ✅ Returns structured response with in_sync, out_of_sync, and errors arrays
- ✅ Added reconciliation endpoint: POST /api/v1/tenants/reconcile
- ✅ Re-publishes tenant.created events for out-of-sync tenants
- ✅ Fallback to HTTP if RabbitMQ unavailable

**Files Changed**:
- `saas-admin-service/internal/handlers/saas_admin_handler.go` - Added VerifySync and ReconcileTenants handlers
- `saas-admin-service/cmd/main.go` - Added routes

**Actual Effort**: 3 hours
**Risk**: Low

---

#### 6. Missing Monitoring Metrics

**Priority**: 🟡 P1 (High)
**Service**: saas-admin-service, tenant-admin-service
**Status**: ✅ COMPLETE (2025-10-29)
**Commit**: b255c44

**Problem**:
- No metrics for tenant sync failures
- Cannot monitor RabbitMQ vs HTTP fallback usage
- No alerting on sync issues

**Implementation**:
- ✅ Added 5 Prometheus metrics:
  * `tenant_sync_total{method="rabbitmq|http", status="success|failure"}` - Counter for sync operations
  * `tenant_sync_duration_seconds{method="rabbitmq|http"}` - Histogram for timing
  * `tenant_creation_total` - Counter for tenant creations
  * `tenant_verification_total` - Counter for sync verifications
  * `tenant_reconciliation_total{status="success|failure"}` - Counter for reconciliation results
- ✅ Instrumented CreateTenant handler with timing and status tracking
- ✅ Instrumented VerifySync handler
- ✅ Instrumented ReconcileTenants handler
- ✅ Metrics exposed on existing /metrics endpoint

**Files Changed**:
- `saas-admin-service/internal/handlers/saas_admin_handler.go` - Added 5 metrics and instrumentation

**Actual Effort**: 2 hours
**Risk**: Low

---

#### 7. RabbitMQ Confirmation Timeout Too Long

**Priority**: 🟢 P2 (Medium)
**Service**: saas-admin-service
**Status**: ✅ COMPLETE (2025-10-29)
**Commit**: 9b2ce47

**Problem**:
- RabbitMQ publisher waits 5 seconds for confirmation
- Should be 2-3 seconds for faster failure detection

**Current Code**:
```go
// File: saas-admin-service/internal/events/publisher.go:130
case <-time.After(5 * time.Second):  // ⚠️ TOO LONG
```

**Implementation**:
- ✅ Changed default timeout from 5s to 2s
- ✅ Made timeout configurable via PublisherConfig.ConfirmationTimeout field
- ✅ Added environment variable support: RABBITMQ_CONFIRMATION_TIMEOUT
- ✅ Timeout value logged at startup
- ✅ Error messages include timeout value for debugging

**Files Changed**:
- `saas-admin-service/internal/events/publisher.go` - Added timeout configuration
- `saas-admin-service/cmd/main.go` - Environment variable parsing

**Actual Effort**: 30 minutes
**Risk**: Low

---

#### 8. Admin User Creation Not Atomic

**Priority**: 🟢 P2 (Medium)
**Service**: tenant-admin-service
**Status**: ✅ COMPLETE (2025-10-29)
**Commit**: 9b2ce47

**Problem**:
- Tenant created, then admin user created separately
- If admin user creation fails, tenant exists without admin
- Should use database transaction for atomicity

**Current Code**:
```go
// File: tenant-admin-service/internal/handlers/tenant_admin_handler.go:746-762
h.service.CreateTenant(tenant)  // ✅ Success
h.service.CreateAdminUser(...)  // ❌ Fails
// Result: Tenant exists, no admin user
```

**Implementation**:
- ✅ Created new CreateTenantWithAdmin method wrapping all operations in single transaction
- ✅ Transaction includes: tenant creation, admin user creation, tenant_admin relationship, default settings, billing record
- ✅ Password hashing with bcrypt within transaction
- ✅ Automatic rollback on any failure
- ✅ Panic recovery with rollback
- ✅ Updated event handler to use atomic method when credentials provided
- ✅ Added method to TenantServiceInterface for proper dependency injection
- ✅ Comprehensive logging for success and failure cases

**Files Changed**:
- `tenant-admin-service/internal/services/tenant_admin_service.go` - Added CreateTenantWithAdmin method
- `tenant-admin-service/internal/events/handler.go` - Updated to use atomic method
- `tenant-admin-service/internal/events/handler.go` - Added to TenantServiceInterface

**Actual Effort**: 2 hours
**Risk**: Low

---

## GROUP 2: Authentication (user-service)

### Status: ✅ Audit Complete (2025-10-29)

**Services**:
- user-service (8081)

**Audit Results**:
- ⚠️ **REDUNDANT** - Duplicates tenant-admin-service functionality
- ❌ **UNUSED** - Only referenced by deprecated api-gateway
- 💰 **COST** - Wastes ~$100-200/month (database + compute)
- 📋 **MAINTENANCE** - Doubles auth codebase maintenance burden

**Recommendation**: **DEPRECATE user-service**

**Detailed Report**: See [GROUP2_USER_SERVICE_AUDIT.md](GROUP2_USER_SERVICE_AUDIT.md)

### Issues Identified

#### 9. Duplicate User Authentication Services

**Priority**: 🔴 P0 (Critical)
**Service**: user-service + tenant-admin-service
**Status**: ⏳ Pending

**Problem**:
- user-service and tenant-admin-service implement identical authentication
- user-service is **NOT USED** by any active service
- Two databases with overlapping schemas (statuspage_user + tenant_admin_db)
- Doubles maintenance, backup costs, operational complexity

**Solution**: Deprecate user-service, migrate SAML (if needed) to tenant-admin-service

**Implementation Plan**:
- [x] Verify no active services use user-service (DONE - only api-gateway references it)
- [x] Compare feature sets (DONE - tenant-admin wins 17-5)
- [ ] Determine if SAML/SSO feature is required (DECISION NEEDED)
- [ ] If SAML needed: Migrate SAML code to tenant-admin-service (8 hours)
- [ ] Mark user-service as DEPRECATED in docs
- [ ] Stop user-service in all environments
- [ ] Archive statuspage_user database

**Files to Change**:
- README.md - Remove user-service
- SERVICE_CATALOG.md - Remove user-service
- DATABASE_ARCHITECTURE.md - Remove statuspage_user
- docker-compose files - Remove user-service
- If SAML needed: tenant-admin-service (models, services, handlers)

**Estimated Effort**: 15 hours (with SAML), 5 hours (without SAML)
**Risk**: Low (service is unused)
**Cost Savings**: ~$100-200/month

---

#### 10. User Service JWT Tokens Too Long

**Priority**: 🟡 P1 (High)
**Service**: user-service
**Status**: ⏳ Pending (⚠️ SKIP IF DEPRECATING)

**Problem**:
- JWT tokens expire in 24 hours (should be 15 minutes)
- No refresh token implementation
- Larger attack window if token compromised

**Solution**: Change to 15-minute JWT + 7-day refresh tokens (like tenant-admin-service)

**Note**: ⚠️ **SKIP THIS IF DEPRECATING SERVICE (Issue #9)**

**Estimated Effort**: 6 hours
**Risk**: Medium

---

#### 11. No Service-to-Service Circuit Breakers

**Priority**: 🟢 P2 (Medium)
**Service**: user-service
**Status**: ⏳ Pending (⚠️ NOT NEEDED)

**Problem**:
- user-service doesn't call other services currently
- If it did, no circuit breakers configured

**Note**: ⚠️ **NOT NEEDED IF DEPRECATING SERVICE (Issue #9)**

**Estimated Effort**: 1 hour (documentation only)
**Risk**: Low

---

## GROUP 3: Core Monitoring

### Status: 🔍 Not Yet Audited

**Services**:
- monitoring-service (8092)
- incident-service (8086)

**Critical Flows to Audit**:
1. Monitor health check execution
2. Auto-incident creation on component failure
3. Component status updates
4. Notification triggering

**Plan**:
- [ ] Audit monitoring-service business logic
- [ ] Audit incident-service business logic
- [ ] Verify circuit breakers on service-to-service calls
- [ ] Document monitor types (HTTP, TCP, Ping, SSL, etc.)
- [ ] Test auto-incident creation flow

---

## GROUP 4: Notification & Components

### Status: 🔍 Not Yet Audited

**Services**:
- component-service (8084)
- notification-service (8085)
- notification-consumer

**Critical Flows to Audit**:
1. Component CRUD operations
2. Component dependency mapping
3. Notification channel implementation (email, SMS, Slack, webhook)
4. Notification consumer async delivery

**Plan**:
- [ ] Audit component-service business logic
- [ ] Audit notification-service business logic
- [ ] Audit notification-consumer event processing
- [ ] Verify RabbitMQ notification queue setup
- [ ] Test multi-channel notification delivery

---

## GROUP 5: Remaining Services

### Status: 🔍 Not Yet Audited

**Services**:
- analytics-service (8090) + analytics-consumer
- payment-service (8088) + billing-consumer
- branding-service (8097)
- event-store-service (8096)
- landing-page-service (8100)
- status-ui-service (8093)
- audit-consumer

**Plan**:
- [ ] Audit each service individually
- [ ] Document business logic
- [ ] Identify common patterns/issues
- [ ] Create service-specific improvement lists

---

## Documentation Issues

### 1. README.md Still Shows API Gateway

**Priority**: 🟡 P1 (High)
**Status**: ⏳ Pending

**Problem**:
- README.md architecture diagram shows API Gateway as entry point
- API Gateway was deprecated on October 26, 2025
- Misleads new developers

**Implementation Plan**:
- [ ] Update README.md architecture diagram
- [ ] Show direct service communication pattern
- [ ] Document circuit breaker usage

**Files to Change**:
- `README.md:134-145`

**Estimated Effort**: 1 hour
**Risk**: None

---

### 2. SERVICE_CATALOG.md Recommends API Gateway

**Priority**: 🟡 P1 (High)
**Status**: ⏳ Pending

**Problem**:
- SERVICE_CATALOG.md states services should communicate via API Gateway
- Contradicts actual architecture

**Implementation Plan**:
- [ ] Update SERVICE_CATALOG.md communication pattern section
- [ ] Document direct HTTP + circuit breaker pattern

**Files to Change**:
- `SERVICE_CATALOG.md:1015-1046`

**Estimated Effort**: 1 hour
**Risk**: None

---

### 3. Add RabbitMQ Event Flow Diagram

**Priority**: 🟢 P2 (Medium)
**Status**: ⏳ Pending

**Problem**:
- No visual diagram of RabbitMQ event flows
- Hard to understand event-driven patterns

**Implementation Plan**:
- [ ] Create EVENT_FLOWS.md document
- [ ] Add diagrams for each event type
- [ ] Document queues, exchanges, routing keys

**Files to Change**:
- `EVENT_FLOWS.md` - NEW FILE

**Estimated Effort**: 2 hours
**Risk**: None

---

## Implementation Progress Log

### 2025-10-29 - Initial Audit Complete

**Completed**:
- ✅ Phase 1: Documentation analysis
- ✅ Phase 2.1: saas-admin-service business logic audit
- ✅ Phase 2.2: tenant-admin-service business logic audit
- ✅ Complete application flow documentation created
- ✅ Improvements tracker created

**Discovered**:
- 26 total issues across all categories
- 6 P0 (Critical) issues
- 11 P1 (High) issues
- 9 P2 (Medium) issues

**Next Steps**:
- Start Group 1 implementations
- Fix P0 issues first: RabbitMQ admin user creation, circuit breakers, retry logic

---

## Implementation Schedule

### Week 1: Group 1 (Admin Services)
- Day 1-2: Fix P0 issues (RabbitMQ admin user, circuit breakers, retries)
- Day 3-4: Fix P1 issues (sync verification, monitoring metrics, tenant-admin auth)
- Day 5: Fix P2 issues, testing, documentation

### Week 2: Group 2 (Authentication)
- Day 1: Audit user-service
- Day 2-3: Implement improvements
- Day 4: Testing
- Day 5: Documentation

### Week 3: Group 3 (Core Monitoring)
- Day 1-2: Audit monitoring-service + incident-service
- Day 3-4: Implement improvements
- Day 5: Testing, documentation

### Week 4: Group 4 (Notification & Components)
- Day 1-2: Audit component-service + notification-service
- Day 3-4: Implement improvements
- Day 5: Testing, documentation

### Week 5: Group 5 (Remaining Services)
- Day 1-3: Audit remaining 9 services
- Day 4-5: Implement critical improvements

### Week 6: Final Review & Polish
- Day 1-2: Cross-service pattern review
- Day 3-4: Documentation updates
- Day 5: Final testing and sign-off

---

## Testing Checklist

### Per-Service Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Health endpoints work
- [ ] Circuit breakers trigger correctly
- [ ] Metrics exposed on /metrics
- [ ] Logs are structured and meaningful

### Cross-Service Testing
- [ ] Tenant provisioning end-to-end (RabbitMQ path)
- [ ] Tenant provisioning end-to-end (HTTP fallback path)
- [ ] User creation respects max_users limit
- [ ] RBAC permissions enforced correctly
- [ ] Component status updates trigger incidents
- [ ] Incidents trigger notifications
- [ ] Multi-tenant isolation verified

### Load Testing
- [ ] 100 concurrent tenant creations
- [ ] 1000 concurrent user logins
- [ ] Circuit breakers protect under load
- [ ] RabbitMQ handles event backlog
- [ ] Database connection pooling works

---

**Document Status**: Living Document - Updated as implementation progresses
**Last Updated**: 2025-10-29
**Next Review**: After Group 1 completion
