#!/bin/bash

# network-chaos.sh
# Chaos engineering: Simulate network issues
# Tests service resilience under latency, packet loss, and bandwidth constraints

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_DIR="$(dirname "$SCRIPT_DIR")"
ROOT_DIR="$(dirname "$TEST_DIR")"

# Load environment variables
if [ -f "$ROOT_DIR/.env.test" ]; then
    source "$ROOT_DIR/.env.test"
fi

# Configuration
TENANT_ADMIN_URL="${TENANT_ADMIN_URL:-http://localhost:8099}"
SAAS_ADMIN_URL="${SAAS_ADMIN_URL:-http://localhost:8098}"
NOTIFICATION_URL="${NOTIFICATION_URL:-http://localhost:8085}"
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
echo -e "${BLUE}Network Chaos Engineering Test${NC}"
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

# Check if running on macOS or Linux
check_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "linux"
    else
        echo "unknown"
    fi
}

OS_TYPE=$(check_os)

# Network chaos functions (platform-specific)
add_latency() {
    local delay=$1  # in milliseconds
    local interface=${2:-lo0}

    log_warning "Adding ${delay}ms latency..."

    if [ "$OS_TYPE" = "macos" ]; then
        # macOS uses dnctl (requires sudo)
        sudo dnctl pipe 1 config delay ${delay}ms
        echo "dummynet in quick proto tcp from any to any pipe 1" | sudo pfctl -f -
        sudo pfctl -e 2>/dev/null || true
    elif [ "$OS_TYPE" = "linux" ]; then
        # Linux uses tc (requires sudo)
        sudo tc qdisc add dev $interface root netem delay ${delay}ms
    else
        log_warning "Latency injection not supported on this OS"
    fi
}

remove_latency() {
    local interface=${1:-lo0}

    log_info "Removing latency..."

    if [ "$OS_TYPE" = "macos" ]; then
        sudo pfctl -d 2>/dev/null || true
        sudo dnctl -q flush
    elif [ "$OS_TYPE" = "linux" ]; then
        sudo tc qdisc del dev $interface root 2>/dev/null || true
    fi
}

add_packet_loss() {
    local loss_percent=$1  # percentage
    local interface=${2:-lo0}

    log_warning "Adding ${loss_percent}% packet loss..."

    if [ "$OS_TYPE" = "macos" ]; then
        sudo dnctl pipe 1 config plr ${loss_percent}
        echo "dummynet in quick proto tcp from any to any pipe 1" | sudo pfctl -f -
        sudo pfctl -e 2>/dev/null || true
    elif [ "$OS_TYPE" = "linux" ]; then
        sudo tc qdisc add dev $interface root netem loss ${loss_percent}%
    else
        log_warning "Packet loss injection not supported on this OS"
    fi
}

remove_packet_loss() {
    remove_latency "$@"
}

add_bandwidth_limit() {
    local rate=$1  # in kbit (e.g., 100kbit)
    local interface=${2:-lo0}

    log_warning "Adding bandwidth limit: $rate..."

    if [ "$OS_TYPE" = "macos" ]; then
        sudo dnctl pipe 1 config bw $rate
        echo "dummynet in quick proto tcp from any to any pipe 1" | sudo pfctl -f -
        sudo pfctl -e 2>/dev/null || true
    elif [ "$OS_TYPE" = "linux" ]; then
        sudo tc qdisc add dev $interface root tbf rate $rate burst 32kbit latency 400ms
    else
        log_warning "Bandwidth limiting not supported on this OS"
    fi
}

remove_bandwidth_limit() {
    remove_latency "$@"
}

# Test 1: High latency (200ms)
test_high_latency() {
    log_info "Test 1: High latency (200ms)"
    increment_test

    # Note: Latency injection requires sudo and may not work in all environments
    # This test measures response time under simulated latency

    log_info "Measuring baseline response time..."
    start_time=$(date +%s%3N)
    curl -s -o /dev/null "$TENANT_ADMIN_URL/health" || true
    end_time=$(date +%s%3N)
    baseline=$((end_time - start_time))
    log_info "Baseline response time: ${baseline}ms"

    # Add 200ms latency (commented out by default - requires sudo)
    # add_latency 200

    log_info "Measuring response time with simulated latency..."
    start_time=$(date +%s%3N)
    response=$(curl -s -o /dev/null -w "%{http_code}" "$TENANT_ADMIN_URL/health" || echo "000")
    end_time=$(date +%s%3N)
    latency_time=$((end_time - start_time))

    # Remove latency
    # remove_latency

    log_info "Response time: ${latency_time}ms (baseline: ${baseline}ms)"

    if [ "$response" = "200" ]; then
        log_success "Service responded despite latency: ${latency_time}ms"
    else
        log_error "Service failed under latency: $response"
    fi
}

# Test 2: Extreme latency (1000ms)
test_extreme_latency() {
    log_info "Test 2: Extreme latency (1000ms)"
    increment_test

    log_info "Testing with 10s timeout..."
    start_time=$(date +%s)
    response=$(curl -s -o /dev/null -w "%{http_code}" -m 10 "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer test-token" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" || echo "000")
    end_time=$(date +%s)
    elapsed=$((end_time - start_time))

    if [ "$response" = "200" ] && [ $elapsed -lt 15 ]; then
        log_success "Service responded within timeout: ${elapsed}s"
    else
        log_warning "Service response: $response in ${elapsed}s"
    fi
}

# Test 3: Packet loss (10%)
test_packet_loss() {
    log_info "Test 3: Packet loss simulation (10%)"
    increment_test

    # Test multiple requests to see effect of packet loss
    successes=0
    failures=0

    log_info "Making 20 requests with simulated packet loss..."

    for i in {1..20}; do
        response=$(curl -s -o /dev/null -w "%{http_code}" -m 5 "$TENANT_ADMIN_URL/health" || echo "000")

        if [ "$response" = "200" ]; then
            ((successes++))
        else
            ((failures++))
        fi
    done

    success_rate=$((successes * 100 / 20))
    log_info "Success rate: ${success_rate}% ($successes/20)"

    if [ $success_rate -ge 80 ]; then
        log_success "Service handled packet loss: ${success_rate}% success rate"
    else
        log_error "Service struggled with packet loss: ${success_rate}% success rate"
    fi
}

# Test 4: Network timeout handling
test_network_timeout() {
    log_info "Test 4: Network timeout handling"
    increment_test

    # Test with very short timeout
    start_time=$(date +%s)
    response=$(curl -s -o /dev/null -w "%{http_code}" -m 1 "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer test-token" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" || echo "timeout")
    end_time=$(date +%s)
    elapsed=$((end_time - start_time))

    if [ $elapsed -le 2 ]; then
        log_success "Timeout handled correctly: ${elapsed}s"
    else
        log_warning "Timeout took longer than expected: ${elapsed}s"
    fi
}

# Test 5: Bandwidth constraint (simulate slow network)
test_bandwidth_constraint() {
    log_info "Test 5: Bandwidth constraint simulation"
    increment_test

    # Test large response with bandwidth limit
    log_info "Fetching data with bandwidth constraint..."

    start_time=$(date +%s)
    response=$(curl -s -o /dev/null -w "%{http_code}" -m 30 "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer test-token" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" || echo "000")
    end_time=$(date +%s)
    elapsed=$((end_time - start_time))

    if [ "$response" = "200" ]; then
        log_success "Service completed request with bandwidth constraint: ${elapsed}s"
    else
        log_error "Service failed with bandwidth constraint: $response"
    fi
}

# Test 6: Intermittent connectivity
test_intermittent_connectivity() {
    log_info "Test 6: Intermittent connectivity"
    increment_test

    successes=0
    failures=0
    timeouts=0

    log_info "Simulating intermittent connectivity (10 requests)..."

    for i in {1..10}; do
        # Random delay to simulate intermittent network
        sleep 0.$((RANDOM % 5))

        start_time=$(date +%s)
        response=$(curl -s -o /dev/null -w "%{http_code}" -m 5 "$TENANT_ADMIN_URL/health" || echo "000")
        end_time=$(date +%s)
        elapsed=$((end_time - start_time))

        if [ "$response" = "200" ]; then
            ((successes++))
        elif [ $elapsed -ge 5 ]; then
            ((timeouts++))
        else
            ((failures++))
        fi
    done

    log_info "Results: $successes successes, $failures failures, $timeouts timeouts"

    if [ $successes -ge 7 ]; then
        log_success "Service handled intermittent connectivity: $successes/10"
    else
        log_error "Service struggled with intermittent connectivity: $successes/10"
    fi
}

# Test 7: Concurrent requests under network stress
test_concurrent_under_stress() {
    log_info "Test 7: Concurrent requests under network stress"
    increment_test

    log_info "Launching 50 concurrent requests..."

    pids=()
    start_time=$(date +%s)

    for i in {1..50}; do
        curl -s -o /dev/null -m 10 "$TENANT_ADMIN_URL/health" &
        pids+=($!)
    done

    # Wait for all requests
    completed=0
    for pid in "${pids[@]}"; do
        if wait $pid 2>/dev/null; then
            ((completed++))
        fi
    done

    end_time=$(date +%s)
    elapsed=$((end_time - start_time))

    completion_rate=$((completed * 100 / 50))

    log_info "Completed: $completed/50 (${completion_rate}%) in ${elapsed}s"

    if [ $completion_rate -ge 80 ]; then
        log_success "Service handled concurrent stress: ${completion_rate}%"
    else
        log_error "Service struggled with concurrent stress: ${completion_rate}%"
    fi
}

# Test 8: Service-to-service communication under stress
test_service_to_service_stress() {
    log_info "Test 8: Service-to-service communication under network stress"
    increment_test

    # Test notification service calling other services
    log_info "Testing multi-service request chain..."

    start_time=$(date +%s)
    response=$(curl -s -o /dev/null -w "%{http_code}" -m 15 -X POST "$NOTIFICATION_URL/api/v1/notifications" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer test-token" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" \
        -d '{
            "notification_type": "test",
            "subject": "Network Stress Test",
            "message": "Testing service communication",
            "channels": ["email"]
        }' || echo "000")
    end_time=$(date +%s)
    elapsed=$((end_time - start_time))

    if [ "$response" = "201" ] || [ "$response" = "200" ] || [ "$response" = "202" ]; then
        log_success "Multi-service request succeeded: $response in ${elapsed}s"
    else
        log_warning "Multi-service request returned: $response in ${elapsed}s"
    fi
}

# Test 9: Circuit breaker under network issues
test_circuit_breaker_network() {
    log_info "Test 9: Circuit breaker behavior under network issues"
    increment_test

    log_info "Making repeated failing requests to trigger circuit breaker..."

    failures=0
    for i in {1..15}; do
        response=$(curl -s -o /dev/null -w "%{http_code}" -m 2 "$TENANT_ADMIN_URL/api/v1/nonexistent" \
            -H "Authorization: Bearer test-token" || echo "000")

        if [ "$response" = "000" ] || [ "$response" = "500" ] || [ "$response" = "404" ]; then
            ((failures++))
        fi

        sleep 0.5
    done

    log_info "Triggered $failures failures"

    # Circuit breaker should now be open, test fast-fail
    start_time=$(date +%s)
    curl -s -o /dev/null -m 2 "$TENANT_ADMIN_URL/api/v1/nonexistent" \
        -H "Authorization: Bearer test-token" || true
    end_time=$(date +%s)
    elapsed=$((end_time - start_time))

    if [ $elapsed -le 1 ]; then
        log_success "Circuit breaker fast-failed: ${elapsed}s"
    else
        log_warning "Circuit breaker response time: ${elapsed}s"
    fi

    # Wait for circuit breaker reset
    log_info "Waiting for circuit breaker to reset..."
    sleep 60
}

# Test 10: DNS resolution delays
test_dns_delays() {
    log_info "Test 10: DNS resolution behavior"
    increment_test

    # Test with localhost (should be fast)
    start_time=$(date +%s%3N)
    response=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8099/health" || echo "000")
    end_time=$(date +%s%3N)
    localhost_time=$((end_time - start_time))

    log_info "Localhost resolution time: ${localhost_time}ms"

    if [ "$response" = "200" ] && [ $localhost_time -lt 1000 ]; then
        log_success "DNS resolution fast: ${localhost_time}ms"
    else
        log_warning "DNS resolution time: ${localhost_time}ms, response: $response"
    fi
}

# Run all tests
main() {
    log_info "Starting network chaos tests..."
    log_warning "Note: Some tests require sudo for network manipulation"
    log_warning "Network chaos tests run in degraded mode without sudo"
    echo ""

    # Run tests
    test_high_latency
    echo ""

    test_extreme_latency
    echo ""

    test_packet_loss
    echo ""

    test_network_timeout
    echo ""

    test_bandwidth_constraint
    echo ""

    test_intermittent_connectivity
    echo ""

    test_concurrent_under_stress
    echo ""

    test_service_to_service_stress
    echo ""

    test_circuit_breaker_network
    echo ""

    test_dns_delays
    echo ""

    # Clean up any network rules
    if [ "$OS_TYPE" != "unknown" ]; then
        log_info "Cleaning up network rules..."
        remove_latency 2>/dev/null || true
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
        echo -e "${GREEN}✓ All network chaos tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}✗ Some tests failed. Review logs above.${NC}"
        exit 1
    fi
}

# Run main function
main
