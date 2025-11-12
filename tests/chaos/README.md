# Chaos Engineering Tests

Comprehensive chaos engineering test suite for validating system resilience under failure conditions.

## Overview

This directory contains chaos engineering tests that simulate various failure scenarios to validate the Beakon Status Page Platform's resilience, fault tolerance, and recovery capabilities.

## Test Scripts

### 1. Database Failure (`database-failure.sh`)

Tests database resilience and circuit breaker behavior.

**Scenarios:**
- Database connection loss (PostgreSQL stop/start)
- Slow query simulation (pg_sleep)
- Connection pool exhaustion (100 concurrent connections)
- Database crash and recovery (connection termination)
- Sequential database failures (3 iterations)
- Read-only mode simulation
- Circuit breaker behavior

**Expected Behavior:**
- Services should handle database failures gracefully
- Circuit breakers should prevent cascade failures
- Services should auto-recover when database is restored
- Connection pools should handle stress without crashing

**Run:**
```bash
cd tests/chaos
./database-failure.sh
```

---

### 2. Redis Failure (`redis-failure.sh`)

Tests three-tier session management resilience.

**Scenarios:**
- Redis unavailability during login (fallback to PostgreSQL)
- Redis failure during active session
- Redis recovery and session sync
- Cache fallback to in-memory
- Redis connection pool stress (50 concurrent sessions)
- Redis slow response handling
- Redis data flush and recovery
- Sequential Redis failures (3 iterations)

**Expected Behavior:**
- Sessions should fallback to PostgreSQL when Redis is down
- Sessions should sync back to Redis after recovery
- Cache should fallback to in-memory when Redis unavailable
- No session loss during Redis failures

**Run:**
```bash
cd tests/chaos
./redis-failure.sh
```

---

### 3. RabbitMQ Failure (`rabbitmq-failure.sh`)

Tests event-driven architecture resilience.

**Scenarios:**
- RabbitMQ unavailable during event publish
- RabbitMQ recovery and message processing
- Message queue overflow (20 rapid messages)
- Consumer failure and auto-recovery
- Network partition simulation
- Message durability after crash
- Dead letter queue handling
- Connection pool exhaustion (50 concurrent publishers)

**Expected Behavior:**
- Services should continue operating when RabbitMQ is down
- Messages should be durable and survive RabbitMQ restarts
- Consumers should auto-reconnect after RabbitMQ recovery
- Dead letter queue should capture failed messages

**Run:**
```bash
cd tests/chaos
./rabbitmq-failure.sh
```

---

### 4. Network Chaos (`network-chaos.sh`)

Tests service behavior under network degradation.

**Scenarios:**
- High latency (200ms)
- Extreme latency (1000ms)
- Packet loss simulation (10%)
- Network timeout handling
- Bandwidth constraints (slow network)
- Intermittent connectivity
- Concurrent requests under network stress (50 requests)
- Service-to-service communication under stress
- Circuit breaker under network issues
- DNS resolution delays

**Expected Behavior:**
- Services should timeout gracefully under extreme latency
- Circuit breakers should prevent hanging requests
- Services should handle packet loss with retries
- Multi-service communication should remain functional

**Note:** Some network chaos tests require `sudo` for network manipulation (tc on Linux, dnctl/pfctl on macOS). Tests run in degraded mode without sudo.

**Run:**
```bash
cd tests/chaos
./network-chaos.sh
```

---

### 5. Resource Exhaustion (`resource-exhaustion.sh`)

Tests service behavior under CPU, memory, and disk stress.

**Scenarios:**
- High CPU load (4 cores stressed)
- Memory pressure (1GB allocation)
- Disk I/O stress
- Concurrent large requests (20 simultaneous)
- File descriptor exhaustion (100 concurrent connections)
- Database connection pool exhaustion (50 concurrent operations)
- Goroutine/thread exhaustion (200 slow requests)
- Combined CPU + Memory + Disk stress
- Recovery after resource exhaustion
- Graceful degradation under load

**Expected Behavior:**
- Services should respond under resource pressure
- Core health endpoints should remain available
- Services should recover after stress ends
- Connection pools should prevent resource exhaustion

**Note:** Tests work best with the `stress` tool installed:
- macOS: `brew install stress`
- Linux: `apt-get install stress`

**Run:**
```bash
cd tests/chaos
./resource-exhaustion.sh
```

---

## Prerequisites

### Required Services
All tests require the following services to be running:
- PostgreSQL (port 5432)
- Redis (port 6379)
- RabbitMQ (ports 5672, 15672)
- tenant-admin-service (port 8099)
- saas-admin-service (port 8098)
- notification-service (port 8085)
- monitoring-service (port 8092)

**Start services:**
```bash
cd microservices
./start-all-services.sh
```

### Optional Tools
For enhanced chaos testing:
- `stress` - CPU/memory/disk stress tool
  - macOS: `brew install stress`
  - Linux: `apt-get install stress`
- `tc` (Linux) or `dnctl/pfctl` (macOS) - Network traffic control (requires sudo)

### Environment Variables
Set in `.env.test` or export manually:
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
REDIS_HOST=localhost
REDIS_PORT=6379
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
TENANT_ADMIN_URL=http://localhost:8099
SAAS_ADMIN_URL=http://localhost:8098
```

---

## Running Tests

### Individual Test
```bash
cd tests/chaos
./database-failure.sh
```

### All Chaos Tests
```bash
cd tests/chaos
for script in database-failure.sh redis-failure.sh rabbitmq-failure.sh network-chaos.sh resource-exhaustion.sh; do
    echo "Running $script..."
    ./$script
    echo ""
done
```

### Make Scripts Executable
```bash
cd tests/chaos
chmod +x *.sh
```

---

## Test Output

Each test script provides color-coded output:
- **BLUE** - Informational messages
- **GREEN** - Test passed
- **RED** - Test failed
- **YELLOW** - Warnings

**Example Output:**
```
========================================
Database Failure Chaos Engineering Test
========================================

[INFO] Test 1: Simulate database connection loss
[WARNING] Stopping PostgreSQL...
[INFO] Testing tenant-admin-service response with database down...
[SUCCESS] Service responded with expected error code: 503
[INFO] Restarting PostgreSQL...
[SUCCESS] Service recovered after database restoration

========================================
Test Summary
========================================
Total Tests:  7
Passed:       7
Failed:       0

✓ All database failure chaos tests passed!
```

---

## Test Coverage Matrix

| Failure Type | Database | Redis | RabbitMQ | Network | Resources |
|--------------|----------|-------|----------|---------|-----------|
| Service Down | ✅ | ✅ | ✅ | ✅ | - |
| Slow Response | ✅ | ✅ | - | ✅ | - |
| Connection Pool | ✅ | ✅ | ✅ | - | ✅ |
| Crash Recovery | ✅ | ✅ | ✅ | - | ✅ |
| Data Loss | - | ✅ | ✅ | ✅ | - |
| Circuit Breaker | ✅ | - | - | ✅ | - |
| Resource Stress | - | - | - | - | ✅ |

---

## Architecture Validation

### Resilience Patterns Tested

1. **Circuit Breakers**
   - Tested in: database-failure.sh, network-chaos.sh
   - Validates: Fast-fail behavior, auto-recovery

2. **Retry with Exponential Backoff**
   - Tested in: All tests
   - Validates: Automatic retry on transient failures

3. **Bulkhead Pattern**
   - Tested in: resource-exhaustion.sh
   - Validates: Resource isolation, connection pooling

4. **Fallback Mechanisms**
   - Tested in: redis-failure.sh
   - Validates: Redis → PostgreSQL → In-memory fallback

5. **Graceful Degradation**
   - Tested in: resource-exhaustion.sh
   - Validates: Core endpoints remain available under stress

6. **Timeout Handling**
   - Tested in: network-chaos.sh
   - Validates: Requests timeout gracefully

---

## Integration with CI/CD

### GitHub Actions Integration
```yaml
name: Chaos Tests

on:
  schedule:
    - cron: '0 2 * * 0'  # Weekly on Sunday 2am

jobs:
  chaos-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Start Services
        run: |
          cd microservices
          docker-compose up -d
      - name: Run Chaos Tests
        run: |
          cd tests/chaos
          ./database-failure.sh
          ./redis-failure.sh
          ./rabbitmq-failure.sh
          ./network-chaos.sh
          ./resource-exhaustion.sh
```

---

## Troubleshooting

### Tests Fail Due to Service Not Running
**Problem:** Services not started before running tests

**Solution:**
```bash
cd microservices
./start-all-services.sh
sleep 10  # Wait for services to be ready
cd ../tests/chaos
./database-failure.sh
```

### Permission Denied on Scripts
**Problem:** Scripts not executable

**Solution:**
```bash
chmod +x tests/chaos/*.sh
```

### Network Chaos Tests Require Sudo
**Problem:** Network manipulation requires root privileges

**Solution:**
- Run with sudo: `sudo ./network-chaos.sh`
- Or run in degraded mode (tests still validate resilience)

### Stress Tool Not Found
**Problem:** `stress` command not available

**Solution:**
```bash
# macOS
brew install stress

# Linux
sudo apt-get install stress
```

### PostgreSQL/Redis/RabbitMQ Commands Not Found
**Problem:** Services not installed or not in PATH

**Solution:**
```bash
# macOS
brew install postgresql@16 redis rabbitmq

# Linux
sudo apt-get install postgresql redis-server rabbitmq-server
```

---

## Best Practices

1. **Run in Staging First**
   - Never run chaos tests in production
   - Use staging environment identical to production

2. **Monitor During Tests**
   - Watch logs: `tail -f microservices/*/logs/*.log`
   - Monitor metrics: Check Prometheus dashboards

3. **Run During Off-Peak Hours**
   - Schedule via cron for weekends/nights
   - Avoid running during business hours

4. **Review Failures**
   - All failures indicate potential production issues
   - Fix issues before deploying to production

5. **Document Findings**
   - Keep track of which tests fail
   - Create tickets for resilience improvements

---

## Chaos Testing Philosophy

> "Chaos Engineering is the discipline of experimenting on a system in order to build confidence in the system's capability to withstand turbulent conditions in production."
> — Principles of Chaos Engineering

**Goals:**
- Identify weaknesses before they cause outages
- Verify circuit breakers and fallback mechanisms
- Build confidence in system resilience
- Validate monitoring and alerting

**Principles:**
1. Build a hypothesis around steady-state behavior
2. Vary real-world events
3. Run experiments in production (or production-like environments)
4. Automate experiments to run continuously
5. Minimize blast radius

---

## Related Documentation

- [TESTING_INFRASTRUCTURE_STATUS.md](../../TESTING_INFRASTRUCTURE_STATUS.md) - Overall testing status
- [Load Tests](../load/README.md) - Performance testing
- [E2E Tests](../e2e/README.md) - End-to-end scenarios
- [ARCHITECTURE.md](../../ARCHITECTURE.md) - System architecture
- [OPERATIONAL_RUNBOOK.md](../../OPERATIONAL_RUNBOOK.md) - Production operations

---

## Contributing

When adding new chaos tests:
1. Follow existing script structure
2. Use color-coded logging (log_info, log_success, log_error, log_warning)
3. Increment test counters (TOTAL_TESTS, PASSED_TESTS, FAILED_TESTS)
4. Clean up resources after tests
5. Document expected behavior
6. Update this README with new test coverage

---

**Last Updated:** 2025-11-04
**Maintained By:** Beakon DevOps Team
