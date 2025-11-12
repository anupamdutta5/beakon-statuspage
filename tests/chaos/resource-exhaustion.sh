#!/bin/bash

# resource-exhaustion.sh
# Chaos engineering: Simulate resource exhaustion
# Tests service behavior under CPU, memory, and disk stress

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
echo -e "${BLUE}Resource Exhaustion Chaos Engineering Test${NC}"
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

# Get system metrics
get_cpu_usage() {
    if [ "$OS_TYPE" = "macos" ]; then
        top -l 1 | grep "CPU usage" | awk '{print $3}' | sed 's/%//'
    elif [ "$OS_TYPE" = "linux" ]; then
        top -bn1 | grep "Cpu(s)" | awk '{print $2}' | sed 's/%us,//'
    else
        echo "0"
    fi
}

get_memory_usage() {
    if [ "$OS_TYPE" = "macos" ]; then
        vm_stat | grep "Pages active" | awk '{print $3}' | sed 's/\.//'
    elif [ "$OS_TYPE" = "linux" ]; then
        free | grep Mem | awk '{print ($3/$2) * 100.0}'
    else
        echo "0"
    fi
}

get_service_pid() {
    local service_port=$1
    lsof -ti :$service_port 2>/dev/null || echo "0"
}

# CPU stress functions
stress_cpu() {
    local duration=$1  # in seconds
    local cores=${2:-2}

    log_warning "Starting CPU stress ($cores cores for ${duration}s)..."

    # Use different stress tools based on availability
    if command -v stress &> /dev/null; then
        stress --cpu $cores --timeout ${duration}s &>/dev/null &
        echo $!
    elif command -v yes &> /dev/null; then
        # Fallback: use 'yes' command
        pids=""
        for i in $(seq 1 $cores); do
            yes > /dev/null &
            pids="$pids $!"
        done
        echo "$pids"
    else
        log_warning "No CPU stress tool available"
        echo "0"
    fi
}

stop_stress() {
    local pids=$1

    if [ "$pids" != "0" ]; then
        for pid in $pids; do
            kill -9 $pid 2>/dev/null || true
        done
    fi
}

# Memory stress functions
stress_memory() {
    local size_mb=$1  # in megabytes
    local duration=$2  # in seconds

    log_warning "Starting memory stress (${size_mb}MB for ${duration}s)..."

    if command -v stress &> /dev/null; then
        stress --vm 1 --vm-bytes ${size_mb}M --timeout ${duration}s &>/dev/null &
        echo $!
    else
        # Fallback: create large file in /tmp
        dd if=/dev/zero of=/tmp/chaos-mem-stress bs=1M count=$size_mb &>/dev/null &
        echo $!
    fi
}

cleanup_memory_stress() {
    rm -f /tmp/chaos-mem-stress 2>/dev/null || true
}

# Disk I/O stress functions
stress_disk_io() {
    local duration=$1  # in seconds

    log_warning "Starting disk I/O stress (${duration}s)..."

    if command -v stress &> /dev/null; then
        stress --io 4 --timeout ${duration}s &>/dev/null &
        echo $!
    else
        # Fallback: continuous disk writes
        for i in {1..4}; do
            dd if=/dev/zero of=/tmp/chaos-io-$i bs=1M count=100 oflag=direct &>/dev/null &
        done
        pgrep -f "dd if=/dev/zero" | tail -1
    fi
}

cleanup_disk_stress() {
    rm -f /tmp/chaos-io-* 2>/dev/null || true
}

# Test 1: Service behavior under high CPU load
test_high_cpu_load() {
    log_info "Test 1: Service response under high CPU load"
    increment_test

    # Get baseline
    log_info "Measuring baseline response time..."
    start_time=$(date +%s%3N)
    curl -s -o /dev/null "$TENANT_ADMIN_URL/health" || true
    end_time=$(date +%s%3N)
    baseline=$((end_time - start_time))
    log_info "Baseline: ${baseline}ms"

    # Start CPU stress
    stress_pids=$(stress_cpu 30 4)
    sleep 2

    # Test service under CPU load
    log_info "Testing under CPU stress..."
    start_time=$(date +%s%3N)
    response=$(curl -s -o /dev/null -w "%{http_code}" -m 10 "$TENANT_ADMIN_URL/health" || echo "000")
    end_time=$(date +%s%3N)
    stress_time=$((end_time - start_time))

    # Stop stress
    stop_stress "$stress_pids"

    log_info "Response under stress: ${stress_time}ms (baseline: ${baseline}ms)"

    if [ "$response" = "200" ] && [ $stress_time -lt 10000 ]; then
        log_success "Service responded under CPU stress: ${stress_time}ms"
    else
        log_error "Service struggled under CPU stress: $response in ${stress_time}ms"
    fi
}

# Test 2: Service behavior under memory pressure
test_memory_pressure() {
    log_info "Test 2: Service response under memory pressure"
    increment_test

    # Start memory stress (1GB)
    stress_pid=$(stress_memory 1024 30)
    sleep 3

    # Test service under memory pressure
    response=$(curl -s -o /dev/null -w "%{http_code}" -m 10 "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer test-token" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" || echo "000")

    # Stop stress
    stop_stress "$stress_pid"
    cleanup_memory_stress

    if [ "$response" = "200" ]; then
        log_success "Service responded under memory pressure: $response"
    else
        log_error "Service failed under memory pressure: $response"
    fi
}

# Test 3: Service behavior under disk I/O stress
test_disk_io_stress() {
    log_info "Test 3: Service response under disk I/O stress"
    increment_test

    # Start disk I/O stress
    stress_pid=$(stress_disk_io 20)
    sleep 2

    # Test database-heavy operation
    start_time=$(date +%s)
    response=$(curl -s -o /dev/null -w "%{http_code}" -m 15 -X POST "$TENANT_ADMIN_URL/api/v1/public/tenants" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"Disk Stress Test\",
            \"email\": \"disk-stress-$(date +%s)@example.com\",
            \"subdomain\": \"disk-$(date +%s)\"
        }" || echo "000")
    end_time=$(date +%s)
    elapsed=$((end_time - start_time))

    # Stop stress
    stop_stress "$stress_pid"
    cleanup_disk_stress

    if [ "$response" = "201" ] || [ "$response" = "200" ]; then
        log_success "Service handled disk I/O stress: $response in ${elapsed}s"
    else
        log_error "Service failed under disk I/O stress: $response"
    fi
}

# Test 4: Multiple concurrent large requests (memory exhaustion)
test_concurrent_large_requests() {
    log_info "Test 4: Concurrent large requests (memory test)"
    increment_test

    log_info "Launching 20 concurrent large requests..."

    pids=()
    start_time=$(date +%s)

    for i in {1..20}; do
        curl -s -o /dev/null -m 30 "$TENANT_ADMIN_URL/api/v1/tenants" \
            -H "Authorization: Bearer test-token" \
            -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" &
        pids+=($!)
    done

    # Wait for all
    completed=0
    for pid in "${pids[@]}"; do
        if wait $pid 2>/dev/null; then
            ((completed++))
        fi
    done

    end_time=$(date +%s)
    elapsed=$((end_time - start_time))

    completion_rate=$((completed * 100 / 20))

    if [ $completion_rate -ge 90 ]; then
        log_success "Handled concurrent large requests: ${completion_rate}% in ${elapsed}s"
    else
        log_error "Failed concurrent large requests: ${completion_rate}% in ${elapsed}s"
    fi
}

# Test 5: File descriptor exhaustion
test_file_descriptor_exhaustion() {
    log_info "Test 5: File descriptor exhaustion simulation"
    increment_test

    # Get current file descriptor limit
    if [ "$OS_TYPE" = "macos" ] || [ "$OS_TYPE" = "linux" ]; then
        fd_limit=$(ulimit -n)
        log_info "File descriptor limit: $fd_limit"
    fi

    # Make many concurrent connections
    log_info "Opening 100 concurrent connections..."

    pids=()
    for i in {1..100}; do
        curl -s -o /dev/null -m 10 "$TENANT_ADMIN_URL/health" &
        pids+=($!)
    done

    sleep 2

    # Make one more request
    response=$(curl -s -o /dev/null -w "%{http_code}" -m 5 "$TENANT_ADMIN_URL/health" || echo "000")

    # Wait for all
    for pid in "${pids[@]}"; do
        wait $pid 2>/dev/null || true
    done

    if [ "$response" = "200" ]; then
        log_success "Service handled file descriptor stress: $response"
    else
        log_error "Service failed file descriptor stress: $response"
    fi
}

# Test 6: Connection pool exhaustion
test_connection_pool_exhaustion() {
    log_info "Test 6: Database connection pool exhaustion"
    increment_test

    log_info "Creating 50 concurrent database operations..."

    pids=()
    for i in {1..50}; do
        curl -s -o /dev/null -m 10 "$TENANT_ADMIN_URL/api/v1/tenants" \
            -H "Authorization: Bearer test-token" \
            -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" &
        pids+=($!)
    done

    sleep 2

    # Make one more request
    response=$(curl -s -o /dev/null -w "%{http_code}" -m 10 "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer test-token" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" || echo "000")

    # Wait for all
    for pid in "${pids[@]}"; do
        wait $pid 2>/dev/null || true
    done

    if [ "$response" = "200" ] || [ "$response" = "503" ]; then
        log_success "Connection pool handled correctly: $response"
    else
        log_error "Unexpected connection pool response: $response"
    fi
}

# Test 7: Goroutine/thread exhaustion
test_goroutine_exhaustion() {
    log_info "Test 7: Goroutine/thread exhaustion simulation"
    increment_test

    # Launch many slow requests to exhaust goroutines
    log_info "Launching 200 slow requests..."

    pids=()
    for i in {1..200}; do
        curl -s -o /dev/null -m 60 "$TENANT_ADMIN_URL/api/v1/tenants" \
            -H "Authorization: Bearer test-token" \
            -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" &
        pids+=($!)

        # Small delay to prevent overwhelming client
        if [ $((i % 20)) -eq 0 ]; then
            sleep 1
        fi
    done

    sleep 3

    # Test if service still responds
    response=$(curl -s -o /dev/null -w "%{http_code}" -m 5 "$TENANT_ADMIN_URL/health" || echo "000")

    # Kill background requests
    for pid in "${pids[@]}"; do
        kill $pid 2>/dev/null || true
    done

    if [ "$response" = "200" ]; then
        log_success "Service responded despite goroutine stress: $response"
    else
        log_error "Service failed under goroutine stress: $response"
    fi

    # Wait for cleanup
    sleep 5
}

# Test 8: Combined resource stress
test_combined_stress() {
    log_info "Test 8: Combined CPU + Memory + Disk stress"
    increment_test

    # Start all stress tests
    cpu_pids=$(stress_cpu 20 2)
    mem_pid=$(stress_memory 512 20)
    disk_pid=$(stress_disk_io 20)

    sleep 3

    # Test service under combined stress
    start_time=$(date +%s)
    response=$(curl -s -o /dev/null -w "%{http_code}" -m 15 "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer test-token" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" || echo "000")
    end_time=$(date +%s)
    elapsed=$((end_time - start_time))

    # Stop all stress
    stop_stress "$cpu_pids"
    stop_stress "$mem_pid"
    stop_stress "$disk_pid"
    cleanup_memory_stress
    cleanup_disk_stress

    if [ "$response" = "200" ] && [ $elapsed -lt 20 ]; then
        log_success "Service survived combined stress: $response in ${elapsed}s"
    else
        log_error "Service struggled with combined stress: $response in ${elapsed}s"
    fi
}

# Test 9: Recovery after resource exhaustion
test_recovery_after_exhaustion() {
    log_info "Test 9: Recovery after resource exhaustion"
    increment_test

    # Cause heavy stress
    log_info "Creating heavy resource stress..."
    cpu_pids=$(stress_cpu 15 4)
    mem_pid=$(stress_memory 1024 15)

    sleep 5

    # Try request during stress (may fail)
    response1=$(curl -s -o /dev/null -w "%{http_code}" -m 5 "$TENANT_ADMIN_URL/health" || echo "000")
    log_info "Response during stress: $response1"

    # Stop stress
    stop_stress "$cpu_pids"
    stop_stress "$mem_pid"
    cleanup_memory_stress

    # Wait for recovery
    log_info "Waiting for recovery (10s)..."
    sleep 10

    # Test again
    response2=$(curl -s -o /dev/null -w "%{http_code}" -m 5 "$TENANT_ADMIN_URL/health" || echo "000")

    if [ "$response2" = "200" ]; then
        log_success "Service recovered after exhaustion: $response2"
    else
        log_error "Service did not recover: $response2"
    fi
}

# Test 10: Graceful degradation
test_graceful_degradation() {
    log_info "Test 10: Graceful degradation under load"
    increment_test

    # Start moderate stress
    cpu_pids=$(stress_cpu 30 2)

    sleep 2

    # Test health endpoint (should always work)
    health_response=$(curl -s -o /dev/null -w "%{http_code}" -m 5 "$TENANT_ADMIN_URL/health" || echo "000")

    # Test liveness (should always work)
    live_response=$(curl -s -o /dev/null -w "%{http_code}" -m 5 "$TENANT_ADMIN_URL/health/live" || echo "000")

    # Test complex operation (may degrade)
    complex_response=$(curl -s -o /dev/null -w "%{http_code}" -m 10 "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer test-token" \
        -H "X-Tenant-ID: tenant-1111-1111-1111-111111111111" || echo "000")

    stop_stress "$cpu_pids"

    log_info "Health: $health_response, Live: $live_response, Complex: $complex_response"

    if [ "$health_response" = "200" ] && [ "$live_response" = "200" ]; then
        log_success "Service maintained core health endpoints: health=$health_response, live=$live_response"
    else
        log_error "Service failed health checks under load"
    fi
}

# Run all tests
main() {
    log_info "Starting resource exhaustion chaos tests..."
    log_warning "These tests may temporarily slow down your system"
    echo ""

    # Check for stress tool
    if ! command -v stress &> /dev/null; then
        log_warning "stress tool not found. Install with: brew install stress (macOS) or apt-get install stress (Linux)"
        log_warning "Tests will use fallback methods (less accurate)"
    fi

    echo ""

    # Run tests
    test_high_cpu_load
    echo ""

    test_memory_pressure
    echo ""

    test_disk_io_stress
    echo ""

    test_concurrent_large_requests
    echo ""

    test_file_descriptor_exhaustion
    echo ""

    test_connection_pool_exhaustion
    echo ""

    test_goroutine_exhaustion
    echo ""

    test_combined_stress
    echo ""

    test_recovery_after_exhaustion
    echo ""

    test_graceful_degradation
    echo ""

    # Cleanup
    log_info "Cleaning up..."
    cleanup_memory_stress
    cleanup_disk_stress

    # Summary
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Test Summary${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "Total Tests:  $TOTAL_TESTS"
    echo -e "${GREEN}Passed:       $PASSED_TESTS${NC}"
    echo -e "${RED}Failed:       $FAILED_TESTS${NC}"
    echo ""

    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}✓ All resource exhaustion chaos tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}✗ Some tests failed. Review logs above.${NC}"
        exit 1
    fi
}

# Run main function
main
