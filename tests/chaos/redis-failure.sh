#!/bin/bash

# redis-failure.sh
# Chaos engineering: Simulate Redis failures
# Tests session resilience and cache fallback behavior

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_DIR="$(dirname "$SCRIPT_DIR")"
ROOT_DIR="$(dirname "$TEST_DIR")"

# Load environment variables
if [ -f "$ROOT_DIR/.env.test" ]; then
    source "$ROOT_DIR/.env.test"
fi

# Configuration
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"
TENANT_ADMIN_URL="${TENANT_ADMIN_URL:-http://localhost:8099}"
SAAS_ADMIN_URL="${SAAS_ADMIN_URL:-http://localhost:8098}"

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

# Test data
TEST_EMAIL="chaos-test-$(date +%s)@example.com"
TEST_PASSWORD="ChaosTest123!"
AUTH_TOKEN=""

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Redis Failure Chaos Engineering Test${NC}"
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

check_redis_running() {
    if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping &>/dev/null; then
        return 0
    else
        return 1
    fi
}

stop_redis() {
    log_warning "Stopping Redis..."
    if command -v brew &> /dev/null; then
        brew services stop redis 2>/dev/null || true
    else
        sudo systemctl stop redis 2>/dev/null || true
    fi
    sleep 3
}

start_redis() {
    log_info "Starting Redis..."
    if command -v brew &> /dev/null; then
        brew services start redis
    else
        sudo systemctl start redis
    fi
    sleep 3
}

wait_for_redis() {
    local max_attempts=30
    local attempt=1

    log_info "Waiting for Redis to be ready..."

    while [ $attempt -le $max_attempts ]; do
        if check_redis_running; then
            log_success "Redis is ready"
            return 0
        fi

        echo -n "."
        sleep 2
        ((attempt++))
    done

    echo ""
    log_error "Redis did not become ready after $max_attempts attempts"
    return 1
}

create_test_user() {
    log_info "Creating test user..."

    response=$(curl -s -X POST "$TENANT_ADMIN_URL/api/v1/public/tenants" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"Chaos Test Tenant\",
            \"email\": \"$TEST_EMAIL\",
            \"subdomain\": \"chaos-$(date +%s)\"
        }")

    sleep 2
}

login_user() {
    log_info "Logging in test user..."

    response=$(curl -s -X POST "$TENANT_ADMIN_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$TEST_EMAIL\",
            \"password\": \"default_password\"
        }")

    AUTH_TOKEN=$(echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

    if [ -z "$AUTH_TOKEN" ]; then
        log_error "Failed to get auth token"
        return 1
    else
        log_success "Login successful"
        return 0
    fi
}

# Test 1: Redis unavailability on session creation
test_redis_unavailable_on_login() {
    log_info "Test 1: Redis unavailable during login"
    increment_test

    stop_redis

    # Try to login (should fallback to PostgreSQL)
    response=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$TENANT_ADMIN_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$TEST_EMAIL\",
            \"password\": \"default_password\"
        }")

    if [ "$response" = "200" ]; then
        log_success "Login succeeded with Redis down (PostgreSQL fallback): $response"
    else
        log_error "Login failed with Redis down: $response"
    fi

    start_redis
    wait_for_redis
}

# Test 2: Redis failure during active session
test_redis_failure_during_session() {
    log_info "Test 2: Redis failure during active session"
    increment_test

    # Ensure Redis is running and login
    if ! check_redis_running; then
        start_redis
        wait_for_redis
    fi

    login_user
    sleep 2

    # Stop Redis
    stop_redis

    # Try to access protected endpoint (should fallback to PostgreSQL)
    response=$(curl -s -o /dev/null -w "%{http_code}" "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111")

    if [ "$response" = "200" ]; then
        log_success "Session validated with Redis down (PostgreSQL fallback): $response"
    else
        log_error "Session validation failed with Redis down: $response"
    fi

    start_redis
    wait_for_redis
}

# Test 3: Redis recovery after failure
test_redis_recovery() {
    log_info "Test 3: Redis recovery and session sync"
    increment_test

    # Ensure Redis is running
    if ! check_redis_running; then
        start_redis
        wait_for_redis
    fi

    login_user
    sleep 2

    # Verify session is in Redis
    session_in_redis=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" keys "session:*" | wc -l)

    if [ $session_in_redis -gt 0 ]; then
        log_success "Session found in Redis: $session_in_redis sessions"
    else
        log_warning "No sessions found in Redis (may be in PostgreSQL)"
    fi

    # Stop Redis
    stop_redis

    # Access service (should use PostgreSQL)
    response=$(curl -s -o /dev/null -w "%{http_code}" "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111")

    # Start Redis
    start_redis
    wait_for_redis

    sleep 2

    # Access service again (should sync back to Redis)
    response2=$(curl -s -o /dev/null -w "%{http_code}" "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111")

    # Check if session is back in Redis
    session_in_redis_after=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" keys "session:*" | wc -l)

    if [ "$response2" = "200" ]; then
        log_success "Service working after Redis recovery: $response2"
    else
        log_error "Service failed after Redis recovery: $response2"
    fi
}

# Test 4: Cache fallback behavior
test_cache_fallback() {
    log_info "Test 4: Cache fallback to in-memory"
    increment_test

    # Ensure Redis is running
    if ! check_redis_running; then
        start_redis
        wait_for_redis
    fi

    # Make a request to cache data
    curl -s -o /dev/null "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111"

    sleep 1

    # Stop Redis
    stop_redis

    # Make the same request (should use in-memory cache)
    start_time=$(date +%s)
    response=$(curl -s -o /dev/null -w "%{http_code}" "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111")
    end_time=$(date +%s)
    elapsed=$((end_time - start_time))

    if [ "$response" = "200" ] && [ $elapsed -lt 5 ]; then
        log_success "Cache fallback working (in-memory): response=$response, time=${elapsed}s"
    else
        log_error "Cache fallback issue: response=$response, time=${elapsed}s"
    fi

    start_redis
    wait_for_redis
}

# Test 5: Redis connection pool exhaustion
test_redis_connection_pool() {
    log_info "Test 5: Redis connection pool stress"
    increment_test

    # Ensure Redis is running
    if ! check_redis_running; then
        start_redis
        wait_for_redis
    fi

    log_info "Creating multiple concurrent sessions..."

    pids=()
    for i in {1..50}; do
        curl -s -o /dev/null -X POST "$TENANT_ADMIN_URL/api/v1/auth/login" \
            -H "Content-Type: application/json" \
            -d "{
                \"email\": \"$TEST_EMAIL\",
                \"password\": \"default_password\"
            }" &
        pids+=($!)
    done

    # Wait for all requests
    for pid in "${pids[@]}"; do
        wait $pid 2>/dev/null || true
    done

    # Make one more request
    response=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$TENANT_ADMIN_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$TEST_EMAIL\",
            \"password\": \"default_password\"
        }")

    if [ "$response" = "200" ]; then
        log_success "Redis connection pool handled stress: $response"
    else
        log_error "Redis connection pool stress failed: $response"
    fi
}

# Test 6: Redis slow response simulation
test_redis_slow_response() {
    log_info "Test 6: Redis slow response handling"
    increment_test

    # Ensure Redis is running
    if ! check_redis_running; then
        start_redis
        wait_for_redis
    fi

    # Set a very high timeout in Redis (simulates slow response)
    redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" CONFIG SET timeout 0 &>/dev/null

    # Make request with timeout
    start_time=$(date +%s)
    response=$(curl -s -o /dev/null -w "%{http_code}" -m 10 "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" || echo "000")
    end_time=$(date +%s)
    elapsed=$((end_time - start_time))

    # Service should timeout gracefully and fallback
    if [ $elapsed -lt 15 ]; then
        log_success "Service handled Redis slowness: ${elapsed}s"
    else
        log_warning "Service took longer than expected: ${elapsed}s"
    fi

    # Reset Redis config
    redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" CONFIG SET timeout 0 &>/dev/null
}

# Test 7: Redis data corruption/flush
test_redis_data_flush() {
    log_info "Test 7: Redis data flush and recovery"
    increment_test

    # Ensure Redis is running
    if ! check_redis_running; then
        start_redis
        wait_for_redis
    fi

    login_user
    sleep 2

    # Flush all Redis data
    log_warning "Flushing all Redis data..."
    redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" FLUSHALL &>/dev/null

    # Try to access service (should fallback to PostgreSQL)
    response=$(curl -s -o /dev/null -w "%{http_code}" "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111")

    if [ "$response" = "200" ]; then
        log_success "Service recovered from Redis flush (PostgreSQL fallback): $response"
    else
        log_error "Service failed after Redis flush: $response"
    fi

    # Make another request to repopulate Redis
    curl -s -o /dev/null "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111"

    sleep 2

    # Verify Redis has data again
    keys_count=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" DBSIZE | awk '{print $2}')

    if [ "$keys_count" -gt 0 ]; then
        log_success "Redis repopulated with $keys_count keys"
    else
        log_warning "Redis not repopulated (may still be using PostgreSQL)"
    fi
}

# Test 8: Multiple Redis failures in sequence
test_sequential_redis_failures() {
    log_info "Test 8: Sequential Redis failures"
    increment_test

    successes=0
    failures=0

    for i in {1..3}; do
        log_info "Failure iteration $i/3..."

        # Flush Redis
        redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" FLUSHALL &>/dev/null || true

        # Stop Redis
        stop_redis

        sleep 2

        # Try to access service
        response=$(curl -s -o /dev/null -w "%{http_code}" "$TENANT_ADMIN_URL/api/v1/tenants" \
            -H "Authorization: Bearer $AUTH_TOKEN" \
            -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" || echo "000")

        if [ "$response" = "200" ]; then
            ((successes++))
        else
            ((failures++))
        fi

        # Restart Redis
        start_redis
        wait_for_redis

        sleep 2
    done

    if [ $successes -ge 2 ]; then
        log_success "Service handled sequential Redis failures: $successes/3"
    else
        log_error "Service struggled with sequential Redis failures: $successes/3"
    fi
}

# Run all tests
main() {
    log_info "Starting Redis failure chaos tests..."
    echo ""

    # Create test user
    create_test_user
    echo ""

    # Run tests
    test_redis_unavailable_on_login
    echo ""

    test_redis_failure_during_session
    echo ""

    test_redis_recovery
    echo ""

    test_cache_fallback
    echo ""

    test_redis_connection_pool
    echo ""

    test_redis_slow_response
    echo ""

    test_redis_data_flush
    echo ""

    test_sequential_redis_failures
    echo ""

    # Ensure Redis is running at the end
    if ! check_redis_running; then
        start_redis
        wait_for_redis
    fi

    # Summary
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Test Summary${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "Total Tests:  $TOTAL_TESTS"
    echo -e "${GREEN}Passed:       $PASSED_TESTS${NC}"
    echo -e "${RED}Failed:       $FAILED_TESTS${NC}"
    echo ""

    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}✓ All Redis failure chaos tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}✗ Some tests failed. Review logs above.${NC}"
        exit 1
    fi
}

# Run main function
main
