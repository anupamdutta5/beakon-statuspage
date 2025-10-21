# Phase 5: Integration Testing & Production Readiness - Test Plan

**Date**: October 20, 2025
**Service**: tenant-admin-service
**Version**: Phase 4 Complete (70%)
**Objective**: Comprehensive testing and production readiness validation

---

## Test Environment

### Service Configuration
```bash
Service: tenant-admin-service
Port: 8099
Database: tenant_admin_db (PostgreSQL)
Redis: beakon-redis (Docker container, port 6379)
Redis Status: Enabled
Environment: Development
```

### Prerequisites
- ✅ PostgreSQL running with tenant_admin_db database
- ✅ Redis running (beakon-redis container)
- ✅ Phase 4 implementation complete (correlation IDs, error handling, logging)
- ✅ Service compiled with all Phase 1-4 features

---

## Test Categories

### 1. End-to-End API Testing (21 endpoints)
**Objective**: Verify all API endpoints function correctly with correlation ID tracking

#### 1.1 Health Check Endpoints
```bash
# Test: Basic health check
curl -s http://localhost:8099/health

# Test: Health check with custom correlation ID
curl -v -H "X-Correlation-ID: test-health-001" http://localhost:8099/health

# Expected: 200 OK, returns correlation ID in header
```

#### 1.2 User Management (6 endpoints)
```bash
# Authentication required - get JWT token first
TOKEN=$(curl -s -X POST http://localhost:8099/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password"}' \
  | jq -r '.token')

# Test: List users
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: test-users-list" \
  http://localhost:8099/api/v1/users

# Test: Get user stats
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: test-users-stats" \
  http://localhost:8099/api/v1/users/stats

# Test: Create user
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: test-users-create" \
  -d '{"email":"newuser@example.com","password":"testpass123","role":"viewer"}' \
  http://localhost:8099/api/v1/users

# Test: Get single user
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: test-users-get" \
  http://localhost:8099/api/v1/users/{user_id}

# Test: Update user
curl -s -X PUT -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: test-users-update" \
  -d '{"role":"manager"}' \
  http://localhost:8099/api/v1/users/{user_id}

# Test: Delete user
curl -s -X DELETE -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: test-users-delete" \
  http://localhost:8099/api/v1/users/{user_id}
```

#### 1.3 Component Management (8 endpoints)
```bash
# Test: List components
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: test-components-list" \
  http://localhost:8099/api/v1/components

# Test: Component stats
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: test-components-stats" \
  http://localhost:8099/api/v1/components/stats

# Test: Create component
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: test-components-create" \
  -d '{"name":"API Server","status":"operational","display_order":1}' \
  http://localhost:8099/api/v1/components

# Test: Get component
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: test-components-get" \
  http://localhost:8099/api/v1/components/{component_id}

# Test: Update component
curl -s -X PUT -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: test-components-update" \
  -d '{"name":"API Server Updated"}' \
  http://localhost:8099/api/v1/components/{component_id}

# Test: Update component status
curl -s -X PUT -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: test-components-status" \
  -d '{"status":"degraded_performance"}' \
  http://localhost:8099/api/v1/components/{component_id}/status

# Test: Reorder components
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: test-components-reorder" \
  -d '{"component_ids":["id1","id2","id3"]}' \
  http://localhost:8099/api/v1/components/reorder

# Test: Delete component
curl -s -X DELETE -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: test-components-delete" \
  http://localhost:8099/api/v1/components/{component_id}
```

#### 1.4 Incident Management (7 endpoints)
```bash
# Test: List incidents
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: test-incidents-list" \
  http://localhost:8099/api/v1/incidents

# Test: Incident stats
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: test-incidents-stats" \
  http://localhost:8099/api/v1/incidents/stats

# Test: Create incident
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: test-incidents-create" \
  -d '{"title":"API Slowness","status":"investigating","impact":"minor","component_id":"comp_id"}' \
  http://localhost:8099/api/v1/incidents

# Test: Get incident
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: test-incidents-get" \
  http://localhost:8099/api/v1/incidents/{incident_id}

# Test: Update incident
curl -s -X PUT -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: test-incidents-update" \
  -d '{"status":"identified"}' \
  http://localhost:8099/api/v1/incidents/{incident_id}

# Test: Resolve incident
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: test-incidents-resolve" \
  -d '{"resolution":"Issue resolved after restart"}' \
  http://localhost:8099/api/v1/incidents/{incident_id}/resolve

# Test: Delete incident
curl -s -X DELETE -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: test-incidents-delete" \
  http://localhost:8099/api/v1/incidents/{incident_id}
```

#### 1.5 Subscriber Management (7 endpoints)
```bash
# Test: List subscribers
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: test-subscribers-list" \
  http://localhost:8099/api/v1/subscribers

# Test: Subscriber stats
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: test-subscribers-stats" \
  http://localhost:8099/api/v1/subscribers/stats

# Test: Create subscriber
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: test-subscribers-create" \
  -d '{"email":"subscriber@example.com"}' \
  http://localhost:8099/api/v1/subscribers

# Test: Get subscriber
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: test-subscribers-get" \
  http://localhost:8099/api/v1/subscribers/{subscriber_id}

# Test: Update subscriber
curl -s -X PUT -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: test-subscribers-update" \
  -d '{"status":"active"}' \
  http://localhost:8099/api/v1/subscribers/{subscriber_id}

# Test: Verify subscriber
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: test-subscribers-verify" \
  -d '{"verification_token":"token_here"}' \
  http://localhost:8099/api/v1/subscribers/{subscriber_id}/verify

# Test: Delete subscriber
curl -s -X DELETE -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: test-subscribers-delete" \
  http://localhost:8099/api/v1/subscribers/{subscriber_id}
```

---

### 2. Error Response Testing (20+ error codes)
**Objective**: Verify all error codes return proper responses with correlation IDs

#### 2.1 Validation Errors
```bash
# Test: MISSING_FIELD error
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: test-error-missing-field" \
  -d '{"password":"test"}' \
  http://localhost:8099/api/v1/users

# Expected Response:
# {
#   "status": "error",
#   "code": "MISSING_FIELD",
#   "message": "Required field is missing: email",
#   "correlation_id": "test-error-missing-field"
# }

# Test: INVALID_FORMAT error
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: test-error-invalid-format" \
  -d '{"email":"notanemail","password":"test"}' \
  http://localhost:8099/api/v1/users

# Expected: INVALID_FORMAT error code
```

#### 2.2 Authorization Errors
```bash
# Test: UNAUTHORIZED error (no token)
curl -s -H "X-Correlation-ID: test-error-unauthorized" \
  http://localhost:8099/api/v1/users

# Expected: 401 UNAUTHORIZED

# Test: INVALID_TOKEN error (malformed token)
curl -s -H "Authorization: Bearer invalid_token_here" \
  -H "X-Correlation-ID: test-error-invalid-token" \
  http://localhost:8099/api/v1/users

# Expected: INVALID_TOKEN error code

# Test: EXPIRED_TOKEN error
# (requires generating an expired token)

# Test: FORBIDDEN error (insufficient permissions)
curl -s -H "Authorization: Bearer $VIEWER_TOKEN" \
  -H "X-Correlation-ID: test-error-forbidden" \
  -X DELETE \
  http://localhost:8099/api/v1/users/{user_id}

# Expected: 403 FORBIDDEN
```

#### 2.3 Resource Errors
```bash
# Test: NOT_FOUND error
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: test-error-not-found" \
  http://localhost:8099/api/v1/users/00000000-0000-0000-0000-000000000000

# Expected:
# {
#   "status": "error",
#   "code": "NOT_FOUND",
#   "message": "User not found: 00000000-0000-0000-0000-000000000000",
#   "correlation_id": "test-error-not-found"
# }

# Test: ALREADY_EXISTS error
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: test-error-already-exists" \
  -d '{"email":"admin@example.com","password":"test"}' \
  http://localhost:8099/api/v1/users

# Expected: ALREADY_EXISTS error code
```

#### 2.4 Business Logic Errors
```bash
# Test: MAX_USERS_EXCEEDED error
# (requires tenant with max_users limit set)
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: test-error-max-users" \
  -d '{"email":"newuser@example.com","password":"test"}' \
  http://localhost:8099/api/v1/users

# Expected (when limit reached):
# {
#   "status": "error",
#   "code": "MAX_USERS_EXCEEDED",
#   "message": "Cannot add user: tenant has reached maximum allowed users (current: 10, max: 10)",
#   "details": {"current_count": 10, "max_allowed": 10},
#   "correlation_id": "test-error-max-users"
# }

# Test: INVALID_STATUS error
curl -s -X PUT -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Correlation-ID: test-error-invalid-status" \
  -d '{"status":"invalid_status"}' \
  http://localhost:8099/api/v1/components/{component_id}/status

# Expected: INVALID_STATUS error code
```

---

### 3. Panic Recovery Testing
**Objective**: Verify service gracefully handles panics with stack traces

#### 3.1 Intentional Panic Test
```bash
# Note: This requires adding a test endpoint that triggers a panic
# For production testing, we verify logs contain panic recovery

# Check logs for panic recovery examples:
tail -100 /tmp/tenant-admin-phase5.log | grep "Panic recovered"

# Expected log format:
# {
#   "level": "error",
#   "msg": "Panic recovered",
#   "correlation_id": "...",
#   "method": "POST",
#   "path": "/api/v1/...",
#   "client_ip": "127.0.0.1",
#   "panic": "...",
#   "stack_trace": "goroutine ...\n..."
# }
```

---

### 4. Correlation ID Tracking
**Objective**: Verify correlation IDs work correctly and are unique

#### 4.1 Auto-Generation Test
```bash
# Test: Auto-generated UUID when no correlation ID provided
for i in {1..5}; do
  curl -s -v http://localhost:8099/health 2>&1 | grep "X-Correlation-Id"
done

# Expected: 5 unique UUID v4 correlation IDs
```

#### 4.2 Client-Provided Correlation ID Test
```bash
# Test: Service accepts and uses client-provided correlation ID
curl -v -H "X-Correlation-ID: client-request-12345" \
  http://localhost:8099/health 2>&1 | grep "X-Correlation-Id"

# Expected: X-Correlation-Id: client-request-12345
```

#### 4.3 Correlation ID in Logs
```bash
# Test: Correlation ID appears in structured logs
curl -s -H "X-Correlation-ID: test-logging-001" \
  http://localhost:8099/health

# Check logs:
tail -10 /tmp/tenant-admin-phase5.log | grep "test-logging-001"

# Expected: Log entry with correlation_id field
```

---

### 5. Redis Caching Tests
**Objective**: Verify Redis caching works correctly with proper invalidation

#### 5.1 Cache Hit/Miss Verification
```bash
# Step 1: Clear Redis cache
docker exec beakon-redis redis-cli FLUSHDB

# Step 2: First request (cache miss)
time curl -s -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: cache-test-001" \
  http://localhost:8099/api/v1/components

# Step 3: Second request (cache hit - should be faster)
time curl -s -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: cache-test-002" \
  http://localhost:8099/api/v1/components

# Expected: Second request significantly faster (2-10x)
```

#### 5.2 Cache Invalidation Test
```bash
# Step 1: Populate cache
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8099/api/v1/components

# Step 2: Verify cache exists
docker exec beakon-redis redis-cli KEYS "tenant-admin:components:*"

# Step 3: Create new component (should invalidate cache)
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Component","status":"operational"}' \
  http://localhost:8099/api/v1/components

# Step 4: Verify cache was invalidated
docker exec beakon-redis redis-cli KEYS "tenant-admin:components:*"

# Expected: Cache keys removed or updated
```

#### 5.3 Cache TTL Verification
```bash
# Check TTL for components cache (5 minutes = 300 seconds)
docker exec beakon-redis redis-cli TTL "tenant-admin:components:tenant_id:20:0"

# Expected: ~300 seconds (or less if cached earlier)

# Check TTL for incidents cache (3 minutes = 180 seconds)
docker exec beakon-redis redis-cli TTL "tenant-admin:incidents:tenant_id:20:0"

# Expected: ~180 seconds

# Check TTL for subscribers cache (10 minutes = 600 seconds)
docker exec beakon-redis redis-cli TTL "tenant-admin:subscribers:tenant_id:20:0"

# Expected: ~600 seconds
```

---

### 6. Redis Failover Testing
**Objective**: Verify graceful degradation when Redis unavailable

#### 6.1 Redis Unavailable Test
```bash
# Step 1: Stop Redis
docker stop beakon-redis

# Step 2: Test API still works (should fall back to database)
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: failover-test-001" \
  http://localhost:8099/api/v1/components | jq

# Expected: 200 OK, data returned from database

# Step 3: Check logs for cache errors (should be handled gracefully)
tail -20 /tmp/tenant-admin-phase5.log | grep -i "cache"

# Step 4: Restart Redis
docker start beakon-redis

# Step 5: Verify caching resumes
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: failover-test-002" \
  http://localhost:8099/api/v1/components | jq

# Expected: Caching resumes automatically
```

---

### 7. Structured Logging Verification
**Objective**: Verify all log fields are present and correct

#### 7.1 Log Field Verification
```bash
# Make request with known parameters
curl -s -H "Authorization: Bearer $TOKEN" \
  -H "X-Correlation-ID: logging-test-001" \
  -H "User-Agent: TestAgent/1.0" \
  http://localhost:8099/api/v1/users

# Check log output
tail -5 /tmp/tenant-admin-phase5.log

# Expected fields in log:
# - level (info/warn/error)
# - ts (timestamp)
# - correlation_id: logging-test-001
# - method: GET
# - path: /api/v1/users
# - status: 200
# - latency: <duration in ns>
# - latency_human: "2.5ms"
# - client_ip: 127.0.0.1
# - user_agent: TestAgent/1.0
# - tenant_id: <uuid>
# - user_id: <uuid>
```

#### 7.2 Log Level Verification
```bash
# Test: Successful request (2xx) logs at INFO level
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8099/health

# Expected: INFO level log

# Test: Client error (4xx) logs at WARN level
curl -s http://localhost:8099/api/v1/users

# Expected: WARN level log (401 unauthorized)

# Test: Server error (5xx) logs at ERROR level
# (requires triggering a server error)

# Expected: ERROR level log
```

---

### 8. Performance Benchmarking
**Objective**: Measure baseline performance and cache improvements

#### 8.1 Baseline Performance (No Cache)
```bash
# Disable cache temporarily or use fresh data
for i in {1..10}; do
  time curl -s -H "Authorization: Bearer $TOKEN" \
    -H "X-Correlation-ID: perf-test-$i" \
    http://localhost:8099/api/v1/components > /dev/null
done

# Calculate average latency
```

#### 8.2 Cache Performance
```bash
# First request (cache miss)
time curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8099/api/v1/components > /dev/null

# Subsequent requests (cache hit)
for i in {1..10}; do
  time curl -s -H "Authorization: Bearer $TOKEN" \
    http://localhost:8099/api/v1/components > /dev/null
done

# Expected: 4-10x faster with cache
```

#### 8.3 Concurrent Request Handling
```bash
# Test concurrent requests with unique correlation IDs
for i in {1..50}; do
  curl -s -H "Authorization: Bearer $TOKEN" \
    -H "X-Correlation-ID: concurrent-$i" \
    http://localhost:8099/api/v1/components > /dev/null &
done
wait

# Verify all 50 unique correlation IDs in logs
grep "concurrent-" /tmp/tenant-admin-phase5.log | wc -l

# Expected: 50 log entries with unique correlation IDs
```

---

### 9. CORS Header Verification
**Objective**: Verify CORS headers are properly configured

```bash
# Test: CORS headers present
curl -v http://localhost:8099/health 2>&1 | grep -i "access-control"

# Expected headers:
# Access-Control-Allow-Headers: Content-Type, Authorization, X-Requested-With, X-Correlation-ID
# Access-Control-Expose-Headers: X-Correlation-ID, X-Total-Count
# Access-Control-Allow-Origin: *
```

---

### 10. Production Readiness Checklist

#### 10.1 Security Checklist
- [ ] JWT secret is strong (≥32 characters)
- [ ] All endpoints require authentication (except health and public)
- [ ] RBAC permissions verified
- [ ] SQL injection protection (GORM parameterized queries)
- [ ] Password hashing with bcrypt
- [ ] Security headers configured (CORS, CSP, X-Frame-Options)
- [ ] Rate limiting enabled
- [ ] No sensitive data in logs
- [ ] No hardcoded credentials

#### 10.2 Performance Checklist
- [ ] Redis caching enabled and working
- [ ] Database connection pool configured (CPU-based)
- [ ] Query pagination implemented
- [ ] Index optimization verified
- [ ] API response times < 100ms (cached)
- [ ] API response times < 500ms (uncached)

#### 10.3 Reliability Checklist
- [ ] Panic recovery implemented
- [ ] Graceful shutdown configured
- [ ] Health check endpoints working
- [ ] Redis failover graceful degradation
- [ ] Database connection retry logic
- [ ] Circuit breakers configured
- [ ] Timeout handling implemented

#### 10.4 Observability Checklist
- [ ] Structured logging with zap
- [ ] Correlation IDs on all requests
- [ ] Log levels properly configured
- [ ] Prometheus metrics exposed
- [ ] Health check detailed component status
- [ ] Error tracking with stack traces
- [ ] Request latency tracking

#### 10.5 Operational Checklist
- [ ] Docker/Kubernetes deployment ready
- [ ] Environment variables documented
- [ ] Database migrations tracked
- [ ] Backup and restore procedures
- [ ] Monitoring and alerting configured
- [ ] Incident response procedures
- [ ] Documentation updated

---

## Test Execution Summary

### Test Results Template
```markdown
| Test Category | Total Tests | Passed | Failed | Notes |
|---------------|-------------|--------|--------|-------|
| End-to-End API | 21 | - | - | - |
| Error Responses | 20+ | - | - | - |
| Panic Recovery | 3 | - | - | - |
| Correlation IDs | 5 | - | - | - |
| Redis Caching | 6 | - | - | - |
| Redis Failover | 5 | - | - | - |
| Structured Logging | 4 | - | - | - |
| Performance | 6 | - | - | - |
| CORS Headers | 3 | - | - | - |
| Production Readiness | 35 | - | - | - |
```

### Critical Issues Log
```markdown
| Issue ID | Severity | Description | Status | Resolution |
|----------|----------|-------------|--------|------------|
| - | - | - | - | - |
```

---

## Next Steps After Testing

1. **Document Test Results**: Create PHASE5_TEST_RESULTS.md with all test outcomes
2. **Fix Critical Issues**: Address any failed tests or critical findings
3. **Update Documentation**: Production deployment guide, troubleshooting guide
4. **Performance Optimization**: If benchmarks show bottlenecks
5. **Final Review**: Code review, security audit
6. **Production Deployment**: Deploy to staging environment first

---

**Document Version**: 1.0
**Last Updated**: October 20, 2025
**Status**: Ready for Execution
