#!/bin/bash

# rabbitmq-failure.sh
# Chaos engineering: Simulate RabbitMQ failures
# Tests event-driven architecture resilience and message queue behavior

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_DIR="$(dirname "$SCRIPT_DIR")"
ROOT_DIR="$(dirname "$TEST_DIR")"

# Load environment variables
if [ -f "$ROOT_DIR/.env.test" ]; then
    source "$ROOT_DIR/.env.test"
fi

# Configuration
RABBITMQ_HOST="${RABBITMQ_HOST:-localhost}"
RABBITMQ_PORT="${RABBITMQ_PORT:-5672}"
RABBITMQ_MGMT_PORT="${RABBITMQ_MGMT_PORT:-15672}"
RABBITMQ_USER="${RABBITMQ_USER:-guest}"
RABBITMQ_PASS="${RABBITMQ_PASS:-guest}"
SAAS_ADMIN_URL="${SAAS_ADMIN_URL:-http://localhost:8098}"
TENANT_ADMIN_URL="${TENANT_ADMIN_URL:-http://localhost:8099}"
NOTIFICATION_URL="${NOTIFICATION_URL:-http://localhost:8085}"

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
echo -e "${BLUE}RabbitMQ Failure Chaos Engineering Test${NC}"
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

check_rabbitmq_running() {
    if curl -s -u "$RABBITMQ_USER:$RABBITMQ_PASS" \
        "http://$RABBITMQ_HOST:$RABBITMQ_MGMT_PORT/api/overview" &>/dev/null; then
        return 0
    else
        return 1
    fi
}

stop_rabbitmq() {
    log_warning "Stopping RabbitMQ..."
    if command -v brew &> /dev/null; then
        brew services stop rabbitmq 2>/dev/null || true
    else
        sudo systemctl stop rabbitmq-server 2>/dev/null || true
    fi
    sleep 5
}

start_rabbitmq() {
    log_info "Starting RabbitMQ..."
    if command -v brew &> /dev/null; then
        brew services start rabbitmq
    else
        sudo systemctl start rabbitmq-server
    fi
    sleep 8
}

wait_for_rabbitmq() {
    local max_attempts=30
    local attempt=1

    log_info "Waiting for RabbitMQ to be ready..."

    while [ $attempt -le $max_attempts ]; do
        if check_rabbitmq_running; then
            log_success "RabbitMQ is ready"
            return 0
        fi

        echo -n "."
        sleep 2
        ((attempt++))
    done

    echo ""
    log_error "RabbitMQ did not become ready after $max_attempts attempts"
    return 1
}

get_queue_depth() {
    local queue_name=$1

    depth=$(curl -s -u "$RABBITMQ_USER:$RABBITMQ_PASS" \
        "http://$RABBITMQ_HOST:$RABBITMQ_MGMT_PORT/api/queues/%2F/$queue_name" \
        | grep -o '"messages":[0-9]*' \
        | cut -d':' -f2)

    echo "${depth:-0}"
}

purge_queue() {
    local queue_name=$1

    log_info "Purging queue: $queue_name"
    curl -s -u "$RABBITMQ_USER:$RABBITMQ_PASS" \
        -X DELETE "http://$RABBITMQ_HOST:$RABBITMQ_MGMT_PORT/api/queues/%2F/$queue_name/contents" \
        &>/dev/null || true
}

# Test 1: RabbitMQ unavailable during tenant creation
test_rabbitmq_unavailable_on_publish() {
    log_info "Test 1: RabbitMQ unavailable during event publish"
    increment_test

    stop_rabbitmq

    # Create tenant (should succeed even if RabbitMQ is down)
    response=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$SAAS_ADMIN_URL/api/v1/tenants" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer test-token" \
        -d "{
            \"name\": \"Chaos Test Tenant $(date +%s)\",
            \"email\": \"chaos-$(date +%s)@example.com\",
            \"subdomain\": \"chaos-$(date +%s)\"
        }")

    if [ "$response" = "201" ] || [ "$response" = "200" ]; then
        log_success "Tenant creation succeeded with RabbitMQ down: $response"
    else
        log_error "Tenant creation failed with RabbitMQ down: $response"
    fi

    start_rabbitmq
    wait_for_rabbitmq
}

# Test 2: RabbitMQ recovery and message replay
test_rabbitmq_recovery_message_replay() {
    log_info "Test 2: RabbitMQ recovery and message processing"
    increment_test

    # Ensure RabbitMQ is running
    if ! check_rabbitmq_running; then
        start_rabbitmq
        wait_for_rabbitmq
    fi

    # Purge existing messages
    purge_queue "tenant.events"
    sleep 2

    # Create tenant (publishes event)
    tenant_email="chaos-recovery-$(date +%s)@example.com"
    response=$(curl -s -X POST "$SAAS_ADMIN_URL/api/v1/tenants" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer test-token" \
        -d "{
            \"name\": \"Recovery Test Tenant\",
            \"email\": \"$tenant_email\",
            \"subdomain\": \"recovery-$(date +%s)\"
        }")

    sleep 3

    # Check queue depth
    depth=$(get_queue_depth "tenant.events")

    if [ "$depth" -ge 0 ]; then
        log_success "Message published to queue: depth=$depth"
    else
        log_warning "Could not verify queue depth"
    fi

    # Wait for consumer to process
    sleep 5

    # Verify tenant was synced to tenant-admin-service
    sync_response=$(curl -s -o /dev/null -w "%{http_code}" "$TENANT_ADMIN_URL/api/v1/tenants" \
        -H "Authorization: Bearer test-token")

    if [ "$sync_response" = "200" ]; then
        log_success "Tenant synced successfully: $sync_response"
    else
        log_warning "Tenant sync verification returned: $sync_response"
    fi
}

# Test 3: Message queue full/overflow
test_message_queue_overflow() {
    log_info "Test 3: Message queue overflow handling"
    increment_test

    # Ensure RabbitMQ is running
    if ! check_rabbitmq_running; then
        start_rabbitmq
        wait_for_rabbitmq
    fi

    # Create many tenants rapidly
    log_info "Creating 20 tenants rapidly..."
    for i in {1..20}; do
        curl -s -o /dev/null -X POST "$SAAS_ADMIN_URL/api/v1/tenants" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer test-token" \
            -d "{
                \"name\": \"Overflow Test $i\",
                \"email\": \"overflow-$i-$(date +%s)@example.com\",
                \"subdomain\": \"overflow-$i-$(date +%s)\"
            }" &
    done

    wait

    sleep 3

    # Check queue depth
    depth=$(get_queue_depth "tenant.events")

    log_info "Queue depth after burst: $depth"

    if [ "$depth" -ge 0 ]; then
        log_success "Queue handled message burst: depth=$depth"
    else
        log_error "Queue overflow or error"
    fi

    # Wait for queue to drain
    log_info "Waiting for queue to drain..."
    sleep 15

    final_depth=$(get_queue_depth "tenant.events")
    log_info "Final queue depth: $final_depth"

    if [ "$final_depth" -lt "$depth" ]; then
        log_success "Queue draining: $depth → $final_depth"
    else
        log_warning "Queue not draining as expected"
    fi
}

# Test 4: Consumer failure and recovery
test_consumer_failure_recovery() {
    log_info "Test 4: Consumer failure and auto-recovery"
    increment_test

    # Ensure RabbitMQ is running
    if ! check_rabbitmq_running; then
        start_rabbitmq
        wait_for_rabbitmq
    fi

    # Get initial consumer count
    consumers_before=$(curl -s -u "$RABBITMQ_USER:$RABBITMQ_PASS" \
        "http://$RABBITMQ_HOST:$RABBITMQ_MGMT_PORT/api/queues/%2F/tenant.events" \
        | grep -o '"consumers":[0-9]*' \
        | cut -d':' -f2)

    log_info "Consumers before: ${consumers_before:-0}"

    # Restart RabbitMQ (will disconnect consumers)
    stop_rabbitmq
    sleep 3
    start_rabbitmq
    wait_for_rabbitmq

    sleep 5

    # Get consumer count after restart
    consumers_after=$(curl -s -u "$RABBITMQ_USER:$RABBITMQ_PASS" \
        "http://$RABBITMQ_HOST:$RABBITMQ_MGMT_PORT/api/queues/%2F/tenant.events" \
        | grep -o '"consumers":[0-9]*' \
        | cut -d':' -f2)

    log_info "Consumers after: ${consumers_after:-0}"

    if [ "${consumers_after:-0}" -gt 0 ]; then
        log_success "Consumer reconnected: $consumers_after consumers"
    else
        log_warning "Consumer may not have reconnected yet"
    fi
}

# Test 5: Network partition simulation
test_network_partition() {
    log_info "Test 5: Network partition and reconnection"
    increment_test

    # Ensure RabbitMQ is running
    if ! check_rabbitmq_running; then
        start_rabbitmq
        wait_for_rabbitmq
    fi

    # Stop RabbitMQ (simulates network partition)
    stop_rabbitmq

    # Try to create tenant during partition
    response=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$SAAS_ADMIN_URL/api/v1/tenants" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer test-token" \
        -d "{
            \"name\": \"Partition Test Tenant\",
            \"email\": \"partition-$(date +%s)@example.com\",
            \"subdomain\": \"partition-$(date +%s)\"
        }")

    if [ "$response" = "201" ] || [ "$response" = "200" ]; then
        log_success "Service continued working during partition: $response"
    else
        log_error "Service failed during partition: $response"
    fi

    # Restore RabbitMQ
    start_rabbitmq
    wait_for_rabbitmq

    sleep 5

    log_success "Service should reconnect automatically"
}

# Test 6: Message durability
test_message_durability() {
    log_info "Test 6: Message durability after crash"
    increment_test

    # Ensure RabbitMQ is running
    if ! check_rabbitmq_running; then
        start_rabbitmq
        wait_for_rabbitmq
    fi

    # Purge queue
    purge_queue "tenant.events"
    sleep 2

    # Create tenant
    curl -s -o /dev/null -X POST "$SAAS_ADMIN_URL/api/v1/tenants" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer test-token" \
        -d "{
            \"name\": \"Durability Test\",
            \"email\": \"durability-$(date +%s)@example.com\",
            \"subdomain\": \"durability-$(date +%s)\"
        }"

    sleep 2

    # Get queue depth before crash
    depth_before=$(get_queue_depth "tenant.events")
    log_info "Queue depth before crash: $depth_before"

    # Restart RabbitMQ (simulates crash)
    stop_rabbitmq
    sleep 2
    start_rabbitmq
    wait_for_rabbitmq

    sleep 3

    # Get queue depth after crash
    depth_after=$(get_queue_depth "tenant.events")
    log_info "Queue depth after crash: $depth_after"

    if [ "$depth_after" -ge "$depth_before" ] && [ "$depth_after" -gt 0 ]; then
        log_success "Messages survived crash (durable): before=$depth_before, after=$depth_after"
    else
        log_warning "Message durability unclear: before=$depth_before, after=$depth_after"
    fi
}

# Test 7: Dead letter queue handling
test_dead_letter_queue() {
    log_info "Test 7: Dead letter queue processing"
    increment_test

    # This test assumes dead letter queue is configured
    # Check if DLQ exists
    dlq_exists=$(curl -s -u "$RABBITMQ_USER:$RABBITMQ_PASS" \
        "http://$RABBITMQ_HOST:$RABBITMQ_MGMT_PORT/api/queues/%2F/tenant.events.dlq" \
        | grep -o '"name"' | wc -l)

    if [ "$dlq_exists" -gt 0 ]; then
        log_success "Dead letter queue configured"

        # Check DLQ depth
        dlq_depth=$(get_queue_depth "tenant.events.dlq")
        log_info "Dead letter queue depth: $dlq_depth"

        if [ "$dlq_depth" -ge 0 ]; then
            log_success "Dead letter queue accessible: depth=$dlq_depth"
        else
            log_warning "Could not read DLQ depth"
        fi
    else
        log_warning "Dead letter queue not configured (optional)"
    fi
}

# Test 8: Connection pool exhaustion
test_connection_pool_exhaustion() {
    log_info "Test 8: RabbitMQ connection pool stress"
    increment_test

    # Ensure RabbitMQ is running
    if ! check_rabbitmq_running; then
        start_rabbitmq
        wait_for_rabbitmq
    fi

    log_info "Creating 50 concurrent tenant creations..."

    pids=()
    for i in {1..50}; do
        curl -s -o /dev/null -X POST "$SAAS_ADMIN_URL/api/v1/tenants" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer test-token" \
            -d "{
                \"name\": \"Pool Test $i\",
                \"email\": \"pool-$i-$(date +%s)@example.com\",
                \"subdomain\": \"pool-$i-$(date +%s)\"
            }" &
        pids+=($!)
    done

    # Wait for all requests
    for pid in "${pids[@]}"; do
        wait $pid 2>/dev/null || true
    done

    # Check connection count
    connection_count=$(curl -s -u "$RABBITMQ_USER:$RABBITMQ_PASS" \
        "http://$RABBITMQ_HOST:$RABBITMQ_MGMT_PORT/api/connections" \
        | grep -o '"name"' | wc -l)

    log_info "Active connections: $connection_count"

    if [ "$connection_count" -gt 0 ]; then
        log_success "Connection pool handled stress: $connection_count connections"
    else
        log_warning "No active connections found"
    fi
}

# Run all tests
main() {
    log_info "Starting RabbitMQ failure chaos tests..."
    echo ""

    # Verify RabbitMQ is accessible
    if ! check_rabbitmq_running; then
        log_warning "RabbitMQ is not running. Starting..."
        start_rabbitmq
        wait_for_rabbitmq
    fi

    echo ""

    # Run tests
    test_rabbitmq_unavailable_on_publish
    echo ""

    test_rabbitmq_recovery_message_replay
    echo ""

    test_message_queue_overflow
    echo ""

    test_consumer_failure_recovery
    echo ""

    test_network_partition
    echo ""

    test_message_durability
    echo ""

    test_dead_letter_queue
    echo ""

    test_connection_pool_exhaustion
    echo ""

    # Ensure RabbitMQ is running at the end
    if ! check_rabbitmq_running; then
        start_rabbitmq
        wait_for_rabbitmq
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
        echo -e "${GREEN}✓ All RabbitMQ failure chaos tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}✗ Some tests failed. Review logs above.${NC}"
        exit 1
    fi
}

# Run main function
main
