#!/bin/bash

# database-failure.sh
# Chaos engineering: Simulate database failures
# Tests database connection resilience and circuit breaker behavior

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_DIR="$(dirname "$SCRIPT_DIR")"
ROOT_DIR="$(dirname "$TEST_DIR")"

# Load environment variables
if [ -f "$ROOT_DIR/.env.test" ]; then
    source "$ROOT_DIR/.env.test"
fi

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
TENANT_ADMIN_URL="${TENANT_ADMIN_URL:-http://localhost:8099}"
SAAS_ADMIN_URL="${SAAS_ADMIN_URL:-http://localhost:8098}"
MONITORING_URL="${MONITORING_URL:-http://localhost:8092}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Database Failure Chaos Engineering Test${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    ((PASSED_TESTS++))
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    ((FAILED_TESTS++))
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

increment_test() {
    ((TOTAL_TESTS++))
}

check_service_health() {
    local service_name=$1
    local service_url=$2

    response=$(curl -s -o /dev/null -w "%{http_code}" "$service_url/health" || echo "000")

    if [ "$response" = "200" ]; then
        return 0
    else
        return 1
    fi
}

wait_for_service() {
    local service_name=$1
    local service_url=$2
    local max_attempts=30
    local attempt=1

    log_info "Waiting for $service_name to be healthy..."

    while [ $attempt -le $max_attempts ]; do
        if check_service_health "$service_name" "$service_url"; then
            log_success "$service_name is healthy"
            return 0
        fi

        echo -n "."
        sleep 2
        ((attempt++))
    done

    echo ""
    log_error "$service_name did not become healthy after $max_attempts attempts"
    return 1
}

# Test 1: Database connection loss simulation
test_database_connection_loss() {
    log_info "Test 1: Simulate database connection loss"
    increment_test

    # Stop PostgreSQL
    log_warning "Stopping PostgreSQL..."
    if command -v brew &> /dev/null; then
        brew services stop postgresql@16 2>/dev/null || true
    else
        sudo systemctl stop postgresql 2>/dev/null || true
    fi

    sleep 3

    # Try to access tenant-admin-service (should fail gracefully)
    log_info "Testing tenant-admin-service response with database down..."
    response=$(curl -s -o /dev/null -w "%{http_code}" "$TENANT_ADMIN_URL/health" || echo "000")

    if [ "$response" = "503" ] || [ "$response" = "500" ]; then
        log_success "Service responded with expected error code: $response"
    else
        log_error "Service responded with unexpected code: $response (expected 500/503)"
    fi

    # Restart PostgreSQL
    log_info "Restarting PostgreSQL..."
    if command -v brew &> /dev/null; then
        brew services start postgresql@16
    else
        sudo systemctl start postgresql
    fi

    sleep 5

    # Verify service recovers
    if wait_for_service "tenant-admin-service" "$TENANT_ADMIN_URL"; then
        log_success "Service recovered after database restoration"
    else
        log_error "Service did not recover after database restoration"
    fi
}

# Test 2: Database slow query simulation
test_database_slow_queries() {
    log_info "Test 2: Simulate slow database queries"
    increment_test

    # Create a slow query using pg_sleep
    log_info "Executing slow query in background..."
    psql -U "$DB_USER" -h "$DB_HOST" -d tenant_admin_db -c "SELECT pg_sleep(30);" &
    SLOW_QUERY_PID=$!

    sleep 2

    # Make API request that should timeout
    log_info "Making API request during slow query..."
    start_time=$(date +%s)
    response=$(curl -s -o /dev/null -w "%{http_code}" -m 10 "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer test-token" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" || echo "000")
    end_time=$(date +%s)
    elapsed=$((end_time - start_time))

    # Kill the slow query
    kill $SLOW_QUERY_PID 2>/dev/null || true

    if [ $elapsed -lt 15 ]; then
        log_success "Request completed/timed out within acceptable time: ${elapsed}s"
    else
        log_error "Request took too long: ${elapsed}s"
    fi
}

# Test 3: Database connection pool exhaustion
test_connection_pool_exhaustion() {
    log_info "Test 3: Simulate connection pool exhaustion"
    increment_test

    log_info "Creating multiple concurrent connections..."

    # Launch 100 concurrent requests
    pids=()
    for i in {1..100}; do
        curl -s -o /dev/null "$TENANT_ADMIN_URL/api/v1/tenants" \
            -H "Authorization: Bearer test-token" \
            -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" &
        pids+=($!)
    done

    sleep 2

    # Make one more request during high load
    response=$(curl -s -o /dev/null -w "%{http_code}" "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer test-token" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" || echo "000")

    # Wait for all background requests to complete
    for pid in "${pids[@]}"; do
        wait $pid 2>/dev/null || true
    done

    if [ "$response" = "200" ] || [ "$response" = "503" ]; then
        log_success "Service handled connection pool stress: $response"
    else
        log_error "Unexpected response during pool stress: $response"
    fi
}

# Test 4: Database crash and recovery
test_database_crash_recovery() {
    log_info "Test 4: Simulate database crash and recovery"
    increment_test

    # Kill all PostgreSQL connections
    log_warning "Terminating all database connections..."
    psql -U "$DB_USER" -h "$DB_HOST" -d postgres -c \
        "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'tenant_admin_db' AND pid <> pg_backend_pid();" \
        2>/dev/null || true

    sleep 2

    # Try to access service
    response=$(curl -s -o /dev/null -w "%{http_code}" "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer test-token" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" || echo "000")

    # Service should recover by reconnecting
    sleep 3

    response2=$(curl -s -o /dev/null -w "%{http_code}" "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer test-token" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" || echo "000")

    if [ "$response2" = "200" ]; then
        log_success "Service recovered from connection termination"
    else
        log_error "Service did not recover: $response2"
    fi
}

# Test 5: Multiple database failures in sequence
test_sequential_database_failures() {
    log_info "Test 5: Sequential database failures"
    increment_test

    failures=0
    recoveries=0

    for i in {1..3}; do
        log_info "Failure iteration $i/3..."

        # Terminate connections
        psql -U "$DB_USER" -h "$DB_HOST" -d postgres -c \
            "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'tenant_admin_db' AND pid <> pg_backend_pid();" \
            2>/dev/null || true

        sleep 2

        # Check if service recovers
        response=$(curl -s -o /dev/null -w "%{http_code}" "$TENANT_ADMIN_URL/health" || echo "000")

        if [ "$response" = "200" ]; then
            ((recoveries++))
        else
            ((failures++))
        fi

        sleep 3
    done

    if [ $recoveries -ge 2 ]; then
        log_success "Service recovered from sequential failures: $recoveries/3"
    else
        log_error "Service failed to recover adequately: $recoveries/3"
    fi
}

# Test 6: Database read-only mode
test_database_readonly_mode() {
    log_info "Test 6: Database read-only mode simulation"
    increment_test

    # Set database to read-only
    log_warning "Setting database to read-only mode..."
    psql -U "$DB_USER" -h "$DB_HOST" -d tenant_admin_db -c \
        "ALTER DATABASE tenant_admin_db SET default_transaction_read_only = on;" \
        2>/dev/null || true

    # Reconnect to apply setting
    psql -U "$DB_USER" -h "$DB_HOST" -d postgres -c \
        "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'tenant_admin_db' AND pid <> pg_backend_pid();" \
        2>/dev/null || true

    sleep 3

    # Try to create a tenant (should fail)
    response=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$TENANT_ADMIN_URL/api/v1/public/tenants" \
        -H "Content-Type: application/json" \
        -d '{"name":"Test Tenant","email":"test@example.com"}' || echo "000")

    if [ "$response" = "500" ] || [ "$response" = "400" ]; then
        log_success "Write operation correctly failed in read-only mode: $response"
    else
        log_warning "Unexpected response in read-only mode: $response"
    fi

    # Restore read-write mode
    log_info "Restoring read-write mode..."
    psql -U "$DB_USER" -h "$DB_HOST" -d tenant_admin_db -c \
        "ALTER DATABASE tenant_admin_db SET default_transaction_read_only = off;" \
        2>/dev/null || true

    sleep 2
}

# Test 7: Network partition simulation (if applicable)
test_network_partition() {
    log_info "Test 7: Network partition simulation"
    increment_test

    log_info "Testing circuit breaker behavior..."

    # Make multiple failed requests to trigger circuit breaker
    for i in {1..10}; do
        psql -U "$DB_USER" -h "$DB_HOST" -d postgres -c \
            "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'tenant_admin_db' AND pid <> pg_backend_pid();" \
            2>/dev/null || true

        curl -s -o /dev/null "$TENANT_ADMIN_URL/api/v1/tenants" \
            -H "Authorization: Bearer test-token" \
            -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" || true

        sleep 1
    done

    # Circuit breaker should be open now
    log_info "Circuit breaker should be open, testing fast-fail..."
    start_time=$(date +%s)
    curl -s -o /dev/null "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer test-token" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" || true
    end_time=$(date +%s)
    elapsed=$((end_time - start_time))

    if [ $elapsed -lt 3 ]; then
        log_success "Circuit breaker fast-failed in ${elapsed}s"
    else
        log_warning "Circuit breaker response took longer than expected: ${elapsed}s"
    fi

    # Wait for circuit breaker to reset
    log_info "Waiting for circuit breaker to reset..."
    sleep 60
}

# Run all tests
main() {
    log_info "Starting database failure chaos tests..."
    echo ""

    # Verify services are running
    log_info "Verifying services are running..."
    if ! wait_for_service "tenant-admin-service" "$TENANT_ADMIN_URL"; then
        log_error "tenant-admin-service is not running. Please start services first."
        exit 1
    fi

    echo ""

    # Run tests
    test_database_connection_loss
    echo ""

    test_database_slow_queries
    echo ""

    test_connection_pool_exhaustion
    echo ""

    test_database_crash_recovery
    echo ""

    test_sequential_database_failures
    echo ""

    test_database_readonly_mode
    echo ""

    test_network_partition
    echo ""

    # Summary
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Test Summary${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "Total Tests:  $TOTAL_TESTS"
    echo -e "${GREEN}Passed:       $PASSED_TESTS${NC}"
    echo -e "${RED}Failed:       $FAILED_TESTS${NC}"
    echo ""

    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}✓ All database failure chaos tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}✗ Some tests failed. Review logs above.${NC}"
        exit 1
    fi
}

# Run main function
main
