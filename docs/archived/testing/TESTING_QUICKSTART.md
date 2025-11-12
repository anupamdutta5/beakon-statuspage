# Testing Infrastructure - Quick Start Guide

**5-Minute Setup Guide for Developers**

---

## Prerequisites

```bash
# Check prerequisites
go version              # Requires 1.21+
docker --version        # Requires Docker
docker-compose --version
psql --version         # PostgreSQL client
```

---

## Setup (One-Time)

### 1. Install Test Tools

```bash
# Install k6 (load testing)
# macOS
brew install k6

# Linux
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg \
  --keyserver hkp://keyserver.ubuntu.com:80 \
  --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | \
  sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Install stress (optional, for chaos tests)
# macOS
brew install stress

# Linux
sudo apt-get install stress
```

### 2. Start Test Environment

```bash
# Start Docker services (PostgreSQL, Redis, RabbitMQ)
cd tests
docker-compose -f docker-compose.test.yml up -d

# Verify services are healthy
docker-compose -f docker-compose.test.yml ps
```

### 3. Initialize Databases

```bash
# Initialize all 14 test databases
cd microservices
./init-all-databases.sh

# Verify database connections
cd ../tests/helpers
./test-db-connections.sh
```

---

## Running Tests

### Unit Tests

```bash
# Run unit tests for a specific service
cd microservices/tenant-admin-service
go test ./tests/unit/... -v -race -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out

# Run all unit tests with helper script
cd tests/helpers
./test-all-services.sh unit
```

**Expected Output:**
```
✓ All unit tests passed!
Total: 92 tests
Coverage: 75%
```

---

### Integration Tests

```bash
# Test database connections
cd tests/helpers
./test-db-connections.sh

# Test Redis
./test-redis.sh

# Test RabbitMQ
./test-rabbitmq.sh

# Run all integration tests
./test-all-services.sh integration
```

---

### E2E Tests

```bash
# 1. Start all services
cd microservices
./start-all-services.sh

# 2. Wait for services to be ready (15 seconds)
sleep 15

# 3. Run E2E tests
cd ../tests/e2e/tenant-onboarding
go test -v ./01_tenant_onboarding_test.go

cd ../monitoring-alerting
go test -v ./02_monitoring_alerting_test.go

cd ../status-page-viewing
go test -v ./03_status_page_viewing_test.go

cd ../session-resilience
go test -v ./04_session_resilience_test.go
```

**Duration:** ~14 minutes for all 4 scenarios

---

### Load Tests

```bash
# Make sure services are running
cd microservices
./start-all-services.sh

# Run individual load tests
cd ../tests/load

# Test 1: Status page load (27 minutes)
k6 run status-page-load.js

# Test 2: Notification burst (5 minutes)
k6 run notification-burst.js

# Test 3: Incident spike (5 minutes)
k6 run incident-spike.js

# Test 4: Database pool stress (11 minutes)
k6 run database-pool-stress.js
```

---

### Chaos Tests

```bash
# Make sure services are running
cd microservices
./start-all-services.sh

# Run chaos tests (requires services to be running)
cd ../tests/chaos

# Test database resilience
./database-failure.sh

# Test Redis failover
./redis-failure.sh

# Test RabbitMQ resilience
./rabbitmq-failure.sh

# Test network issues (requires sudo for full features)
./network-chaos.sh

# Test resource exhaustion
./resource-exhaustion.sh
```

**Note:** Chaos tests may temporarily impact system performance

---

## CI/CD Workflows

### Local Workflow Testing

```bash
# Install act (GitHub Actions local runner)
# macOS
brew install act

# Linux
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Run unit tests workflow locally
act pull_request -j unit-tests

# Run E2E tests workflow locally
act push -j e2e-tests
```

### GitHub Actions

Workflows run automatically:
- **Unit Tests**: Every PR to develop/main
- **Integration Tests**: Push to develop, PR to main
- **E2E Tests**: Push/PR to main
- **Security Scan**: Daily at 2 AM UTC
- **Performance Tests**: Weekly on Sunday 3 AM UTC

View results in GitHub Actions tab.

---

## Common Commands

### Test Environment

```bash
# Start test environment
cd tests
docker-compose -f docker-compose.test.yml up -d

# Stop test environment (keep data)
docker-compose -f docker-compose.test.yml stop

# Stop and remove (clean slate)
docker-compose -f docker-compose.test.yml down

# View logs
docker-compose -f docker-compose.test.yml logs -f postgres
docker-compose -f docker-compose.test.yml logs -f redis
docker-compose -f docker-compose.test.yml logs -f rabbitmq
```

### Services

```bash
# Start all services
cd microservices
./start-all-services.sh

# Check service status
./status-dev.sh

# Stop all services
./stop-all-services.sh

# View service logs
tail -f microservices/tenant-admin-service/logs/tenant-admin.log
```

### Quick Health Check

```bash
# Check all services are healthy
curl http://localhost:8099/health  # tenant-admin
curl http://localhost:8098/health  # saas-admin
curl http://localhost:8084/health  # component
curl http://localhost:8085/health  # notification
curl http://localhost:8092/health  # monitoring
```

---

## Troubleshooting

### Tests Fail - Database Connection

**Problem:** Cannot connect to PostgreSQL

**Solution:**
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Restart PostgreSQL
cd tests
docker-compose -f docker-compose.test.yml restart postgres

# Re-initialize databases
cd ../microservices
./init-all-databases.sh
```

---

### Tests Fail - Service Not Running

**Problem:** E2E tests fail because service isn't running

**Solution:**
```bash
# Check which services are running
lsof -i :8099  # tenant-admin
lsof -i :8098  # saas-admin

# Start missing services
cd microservices
./start-all-services.sh

# Wait for health checks
sleep 15

# Verify health
curl http://localhost:8099/health
```

---

### Tests Fail - Port Already in Use

**Problem:** Service can't start because port is in use

**Solution:**
```bash
# Find process using port
lsof -ti :8099

# Kill process
kill -9 $(lsof -ti :8099)

# Restart service
cd microservices/tenant-admin-service
./start-dev.sh
```

---

### Load Tests Timeout

**Problem:** k6 tests timeout or fail

**Solution:**
```bash
# Reduce VU count for local testing
# Edit tests/load/*.js and reduce target VUs

# Example: Change from 1000 to 100
# stages: [
#   { duration: '2m', target: 100 },  // Was 1000
# ]

# Or increase timeout
k6 run --timeout 30m status-page-load.js
```

---

### Chaos Tests Require Sudo

**Problem:** Network chaos tests don't work without sudo

**Solution:**
```bash
# Run with sudo
sudo ./network-chaos.sh

# Or run in degraded mode (still validates resilience)
./network-chaos.sh
# Tests will skip network manipulation but still test resilience
```

---

## Test File Locations

```
Beakon/
├── tests/
│   ├── docker-compose.test.yml    # Test environment
│   ├── fixtures/                  # Test data (14 files)
│   ├── helpers/                   # Helper scripts (7 files)
│   ├── e2e/                       # E2E tests (4 scenarios)
│   ├── load/                      # k6 load tests (4 scenarios)
│   └── chaos/                     # Chaos tests (5 scripts)
│
├── microservices/
│   ├── tenant-admin-service/tests/unit/
│   ├── saas-admin-service/tests/unit/
│   ├── landing-page-service/tests/unit/
│   └── branding-service/tests/unit/
│
└── .github/workflows/             # CI/CD (5 workflows)
```

---

## Performance Expectations

### Unit Tests
- **Duration**: 5-10 minutes (all services)
- **Coverage**: 60%+ minimum, 70%+ target

### Integration Tests
- **Duration**: 10-15 minutes
- **Pass Rate**: >95%

### E2E Tests
- **Duration**: 14 minutes (all 4 scenarios)
- **Pass Rate**: 100% (all must pass)

### Load Tests
- **Status Page Load**: p95 < 2s, p99 < 5s
- **Notification Burst**: 100/sec with 95%+ delivery
- **Incident Spike**: 50/sec sustained
- **Database Pool**: Support 500 concurrent connections

### Chaos Tests
- **Recovery Rate**: 100% (all services must recover)
- **Failover Time**: < 100ms for Redis → PostgreSQL

---

## Documentation

- **[TESTING_INFRASTRUCTURE_STATUS.md](./TESTING_INFRASTRUCTURE_STATUS.md)** - Complete status report
- **[TESTING_COMPLETE_SUMMARY.md](./TESTING_COMPLETE_SUMMARY.md)** - Implementation summary
- **[tests/load/README.md](./tests/load/README.md)** - Load testing guide
- **[tests/chaos/README.md](./tests/chaos/README.md)** - Chaos engineering guide
- **[.github/workflows/README.md](./.github/workflows/README.md)** - CI/CD workflows

---

## Getting Help

### Check Logs
```bash
# Service logs
tail -f microservices/*/logs/*.log

# Docker logs
docker-compose -f tests/docker-compose.test.yml logs -f

# Test output
go test -v ./... 2>&1 | tee test-output.log
```

### Verify Environment
```bash
# Check all prerequisites
cd tests/helpers
./test-service-health.sh

# Verify test data
ls -la tests/fixtures/
```

### Clean Slate
```bash
# Stop everything
cd microservices
./stop-all-services.sh

cd ../tests
docker-compose -f docker-compose.test.yml down -v

# Start fresh
docker-compose -f docker-compose.test.yml up -d
cd ../microservices
./init-all-databases.sh
./start-all-services.sh
```

---

## Next Steps

1. **Run Your First Test**
   ```bash
   cd microservices/tenant-admin-service
   go test ./tests/unit/... -v
   ```

2. **Add Tests for Your Service**
   - Copy test structure from existing services
   - Follow patterns in `tenant-admin-service/tests/unit/`

3. **Run Load Tests**
   ```bash
   cd tests/load
   k6 run status-page-load.js
   ```

4. **Validate Resilience**
   ```bash
   cd tests/chaos
   ./database-failure.sh
   ```

---

**Quick Reference Card:**

```bash
# Setup
docker-compose -f tests/docker-compose.test.yml up -d
cd microservices && ./init-all-databases.sh

# Unit Tests
go test ./tests/unit/... -v -race -coverprofile=coverage.out

# E2E Tests
./start-all-services.sh && sleep 15
cd tests/e2e/tenant-onboarding && go test -v

# Load Tests
cd tests/load && k6 run status-page-load.js

# Chaos Tests
cd tests/chaos && ./database-failure.sh

# Cleanup
./stop-all-services.sh
docker-compose -f tests/docker-compose.test.yml down
```

---

**Last Updated:** November 4, 2025
**Version:** 1.0.0
