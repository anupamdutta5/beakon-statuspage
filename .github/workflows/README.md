# GitHub Actions CI/CD Workflows

Automated testing and security workflows for the Beakon Status Page Platform.

## Overview

This directory contains GitHub Actions workflows that automate testing, security scanning, and performance validation across all 19 microservices.

## Workflows

### 1. Unit Tests (`unit-tests.yml`)

**Trigger:**
- On pull requests to `develop` or `main`
- On push to `develop`
- Only when Go files or dependencies change

**What it does:**
- Runs unit tests for all 16 services in parallel
- Generates code coverage reports
- Uploads coverage to Codecov
- Enforces minimum 60% code coverage threshold
- Uses Go 1.21 with dependency caching

**Matrix Strategy:**
```yaml
services:
  - tenant-admin-service
  - saas-admin-service
  - landing-page-service
  - branding-service
  - component-service
  - incident-service
  - notification-service
  - monitoring-service
  - event-store-service
  - payment-service
  - analytics-service
  - database-service
  - audit-consumer
  - analytics-consumer
  - notification-consumer
  - billing-consumer
```

**Expected Duration:** 5-10 minutes (parallel execution)

**Commands:**
```bash
# Run locally
cd microservices/tenant-admin-service
go test ./... -v -race -coverprofile=coverage.out
go tool cover -func=coverage.out
```

---

### 2. Integration Tests (`integration-tests.yml`)

**Trigger:**
- On push to `develop`
- On pull requests to `main`
- When integration test files change

**What it does:**
- Spins up PostgreSQL, Redis, RabbitMQ in GitHub Actions services
- Initializes all test databases
- Runs integration tests for key services
- Tests database migrations
- Tests message queue connectivity
- Tests cache integration

**Services Used:**
- PostgreSQL 16
- Redis 7
- RabbitMQ 3.12

**Test Categories:**
1. **Service Integration** - Tests service-level integrations
2. **Database Integration** - Tests migrations and connections
3. **Message Queue Integration** - Tests RabbitMQ connectivity
4. **Cache Integration** - Tests Redis connectivity

**Expected Duration:** 10-15 minutes

**Commands:**
```bash
# Run locally (requires Docker)
docker-compose -f tests/docker-compose.test.yml up -d
cd tests/helpers
./test-db-connections.sh
./test-rabbitmq.sh
./test-redis.sh
```

---

### 3. E2E Tests (`e2e-tests.yml`)

**Trigger:**
- On push to `main`
- On pull requests to `main`
- Manual trigger via `workflow_dispatch`

**What it does:**
- Builds all required services
- Starts services in background
- Waits for health checks
- Runs 4 E2E test scenarios:
  1. Tenant Onboarding Flow (10 steps)
  2. Monitoring & Alerting Flow (10 steps)
  3. Status Page Viewing Flow (12 steps)
  4. Session Resilience Flow (14 steps)
- Uploads service logs on failure

**Services Started:**
- tenant-admin-service (8099)
- saas-admin-service (8098)
- component-service (8084)
- notification-service (8085)
- monitoring-service (8092)

**Expected Duration:** 15-20 minutes

**Timeout:** 30 minutes

**Commands:**
```bash
# Run locally
cd microservices
./start-all-services.sh
sleep 15

cd ../tests/e2e/tenant-onboarding
go test -v ./01_tenant_onboarding_test.go
```

---

### 4. Security Scan (`security-scan.yml`)

**Trigger:**
- Daily at 2 AM UTC (scheduled)
- On push to `main` or `develop`
- On pull requests to `main`
- Manual trigger

**What it does:**

#### A. **gosec** - Go Security Scanner
- Scans for common security issues in Go code
- Detects hardcoded credentials, SQL injection, XSS vulnerabilities
- Uploads results to GitHub Security tab (SARIF format)

#### B. **govulncheck** - Dependency Vulnerabilities
- Checks for known vulnerabilities in dependencies
- Uses official Go vulnerability database
- Reports CVEs and affected packages

#### C. **Gitleaks** - Secret Scanning
- Scans commit history for exposed secrets
- Detects API keys, passwords, tokens
- Prevents accidental secret commits

#### D. **Trivy** - Docker Image Scanning
- Scans Docker images for vulnerabilities
- Checks OS packages and application dependencies
- Reports CRITICAL and HIGH severity issues

#### E. **CodeQL** - Code Analysis
- Semantic code analysis by GitHub
- Detects security vulnerabilities and code quality issues
- Runs security-and-quality query suite

#### F. **License Check** - Open Source Compliance
- Checks licenses of all dependencies
- Ensures compliance with acceptable licenses
- Generates license report

**Expected Duration:** 20-30 minutes

**Severity Levels:**
- CRITICAL - Immediate action required
- HIGH - Fix within 7 days
- MEDIUM - Fix within 30 days
- LOW - Informational

**Commands:**
```bash
# Run gosec locally
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...

# Run govulncheck locally
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Run Gitleaks locally
docker run -v $(pwd):/path zricethezav/gitleaks:latest detect --source /path
```

---

### 5. Performance Tests (`performance-tests.yml`)

**Trigger:**
- Weekly on Sunday at 3 AM UTC (scheduled)
- Manual trigger with test selection

**What it does:**
- Runs k6 load tests in isolated environments
- Tests system performance under various load patterns
- Generates performance metrics and reports
- Uploads results as artifacts (30-day retention)

**Test Scenarios:**

#### A. **Status Page Load Test** (27 minutes)
- Gradual ramp: 10 → 1000 concurrent users
- Sustained load: 1000 users for 10 minutes
- Thresholds: p95 < 2s, p99 < 5s
- Tests: Page loads, components, incidents, uptime, branding

#### B. **Notification Burst Test** (5 minutes)
- Constant rate: 100 notifications/second
- Channels: Email, Slack, Webhook, SMS
- Thresholds: p95 < 3s, 95%+ delivery success
- Tests notification service throughput

#### C. **Incident Spike Test** (5 minutes)
- Normal: 10 incidents/second
- Spike: 50 incidents/second
- Sustained spike: 2 minutes
- Tests incident service resilience

#### D. **Database Pool Stress Test** (11 minutes)
- Gradual: 10 → 300 VUs
- Spike: 500 VUs for 30s
- Read/write mix: 60% reads, 25% complex reads, 15% writes
- Tests connection pool behavior

**Expected Duration:** 30-60 minutes (parallel execution)

**Commands:**
```bash
# Install k6
brew install k6  # macOS
# OR
sudo apt-get install k6  # Linux

# Run tests locally
cd tests/load
k6 run status-page-load.js
k6 run notification-burst.js
k6 run incident-spike.js
k6 run database-pool-stress.js
```

---

## Workflow Status Badges

Add these to your README.md:

```markdown
![Unit Tests](https://github.com/anupamdutta5/Beakon/actions/workflows/unit-tests.yml/badge.svg)
![Integration Tests](https://github.com/anupamdutta5/Beakon/actions/workflows/integration-tests.yml/badge.svg)
![E2E Tests](https://github.com/anupamdutta5/Beakon/actions/workflows/e2e-tests.yml/badge.svg)
![Security Scan](https://github.com/anupamdutta5/Beakon/actions/workflows/security-scan.yml/badge.svg)
![Performance Tests](https://github.com/anupamdutta5/Beakon/actions/workflows/performance-tests.yml/badge.svg)
```

---

## Required GitHub Secrets

Configure these in **Settings → Secrets and variables → Actions**:

### Optional (for enhanced features)
- `CODECOV_TOKEN` - For Codecov integration (coverage upload)
- `GITLEAKS_LICENSE` - For Gitleaks Pro features (optional)

### Not Required
All workflows run without additional secrets. Services like GitHub CodeQL and container scanning are built-in.

---

## Workflow Dependencies

### All Workflows
- Go 1.21+
- Git (with submodule support)

### Integration Tests
- PostgreSQL client (`postgresql-client`)
- Docker services (automatically provided by GitHub Actions)

### E2E Tests
- PostgreSQL client
- Service build dependencies

### Security Scans
- gosec, govulncheck (installed automatically)
- Trivy (via GitHub Action)
- CodeQL (GitHub-native)
- Gitleaks (via GitHub Action)

### Performance Tests
- k6 load testing tool
- PostgreSQL client

---

## Local Testing

### Run Unit Tests Locally
```bash
cd microservices/tenant-admin-service
go test ./... -v -race -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run Integration Tests Locally
```bash
# Start dependencies
docker-compose -f tests/docker-compose.test.yml up -d

# Initialize databases
cd microservices
./init-all-databases.sh

# Run tests
cd tests/helpers
./test-db-connections.sh
./test-rabbitmq.sh
./test-redis.sh
```

### Run E2E Tests Locally
```bash
# Start all services
cd microservices
./start-all-services.sh

# Wait for services
sleep 15

# Run E2E tests
cd ../tests/e2e/tenant-onboarding
go test -v ./01_tenant_onboarding_test.go
```

### Run Security Scans Locally
```bash
# gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...

# govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Gitleaks
docker run -v $(pwd):/path zricethezav/gitleaks:latest detect --source /path
```

### Run Performance Tests Locally
```bash
# Install k6
brew install k6  # macOS
# OR
sudo apt-get install k6  # Linux

# Start services
cd microservices
./start-all-services.sh

# Run k6 tests
cd ../tests/load
k6 run status-page-load.js
```

---

## Troubleshooting

### Unit Tests Fail
**Problem:** Tests fail with import errors

**Solution:**
```bash
cd microservices/service-name
go mod download
go mod verify
go test ./...
```

### Integration Tests Fail - Database Connection
**Problem:** Cannot connect to PostgreSQL

**Solution:**
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check logs
docker-compose -f tests/docker-compose.test.yml logs postgres

# Restart services
docker-compose -f tests/docker-compose.test.yml down
docker-compose -f tests/docker-compose.test.yml up -d
```

### E2E Tests Fail - Service Not Ready
**Problem:** Services not starting in time

**Solution:**
```bash
# Increase wait time
sleep 30

# Check service logs
tail -f microservices/*/logs/*.log

# Manually test health
curl http://localhost:8099/health
```

### Security Scan False Positives
**Problem:** gosec reports false positives

**Solution:**
```go
// Add nosec comment with justification
password := os.Getenv("DB_PASSWORD") // #nosec G101 - Environment variable, not hardcoded
```

### Performance Tests Timeout
**Problem:** k6 tests timeout in CI

**Solution:**
- Reduce VU count for CI environment
- Adjust thresholds for CI (less strict)
- Use smaller datasets

---

## Best Practices

### 1. **Keep Workflows Fast**
- Use matrix strategy for parallel execution
- Cache dependencies (Go modules, npm packages)
- Only run tests when relevant files change

### 2. **Fail Fast**
- Use `fail-fast: false` for matrix jobs to see all failures
- Set appropriate timeouts
- Run quick tests before slow tests

### 3. **Clear Reporting**
- Use GitHub step summaries for results
- Upload artifacts for detailed analysis
- Add comments to PRs with test results

### 4. **Security First**
- Never commit secrets
- Use GitHub Secrets for sensitive data
- Review security scan results regularly

### 5. **Performance Monitoring**
- Run performance tests weekly
- Compare results over time
- Set up alerts for regressions

---

## CI/CD Pipeline Flow

```
┌─────────────────────────────────────────────────────────────┐
│                     Pull Request Created                     │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
          ┌───────────────────────────────┐
          │      Unit Tests (5-10 min)    │
          │  ✓ All 16 services in parallel│
          │  ✓ Code coverage > 60%        │
          └───────────────┬───────────────┘
                          │
                          ▼
          ┌───────────────────────────────┐
          │  Security Scan (20-30 min)    │
          │  ✓ gosec, govulncheck         │
          │  ✓ Gitleaks, CodeQL           │
          └───────────────┬───────────────┘
                          │
                          ▼
            ┌─────────────────────────┐
            │   Merge to develop      │
            └─────────┬───────────────┘
                      │
                      ▼
    ┌─────────────────────────────────────┐
    │  Integration Tests (10-15 min)      │
    │  ✓ Database, Redis, RabbitMQ        │
    └─────────────┬───────────────────────┘
                  │
                  ▼
        ┌─────────────────────┐
        │   Merge to main     │
        └─────────┬───────────┘
                  │
                  ▼
    ┌─────────────────────────────────┐
    │    E2E Tests (15-20 min)        │
    │  ✓ 4 complete user flows        │
    └─────────────┬───────────────────┘
                  │
                  ▼
    ┌─────────────────────────────────┐
    │         Deploy                  │
    └─────────────────────────────────┘

Weekly (Sunday 3 AM):
    ┌─────────────────────────────────┐
    │  Performance Tests (30-60 min)  │
    │  ✓ Load, burst, spike, stress   │
    └─────────────────────────────────┘

Daily (2 AM):
    ┌─────────────────────────────────┐
    │   Security Scan (full suite)    │
    │  ✓ All security checks          │
    └─────────────────────────────────┘
```

---

## Metrics and Reporting

### Test Coverage
- **Target:** 70% overall coverage
- **Minimum:** 60% per service
- **Reported:** Codecov dashboard

### Performance Benchmarks
- **Status Page Load:** p95 < 2s, p99 < 5s
- **Notification Burst:** 100/sec with 95%+ delivery
- **Incident Spike:** Handle 50/sec sustained
- **Database Pool:** Support 500 concurrent connections

### Security Posture
- **Zero** CRITICAL vulnerabilities
- **Zero** exposed secrets
- **All** dependencies up-to-date
- **100%** license compliance

---

## Related Documentation

- [Testing Infrastructure Status](../../TESTING_INFRASTRUCTURE_STATUS.md)
- [Load Tests](../../tests/load/README.md)
- [E2E Tests](../../tests/e2e/)
- [Chaos Tests](../../tests/chaos/README.md)
- [OPERATIONAL_RUNBOOK.md](../../OPERATIONAL_RUNBOOK.md)

---

**Last Updated:** 2025-11-04
**Maintained By:** Beakon DevOps Team
