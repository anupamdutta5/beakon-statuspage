#!/bin/bash

# run-integration-tests.sh
# Run integration tests for all microservices
# Requires test environment to be running
# Usage: ./scripts/test/run-integration-tests.sh [service-name]

set -e

echo "=== Beakon Integration Tests ==="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Specific service to test (optional)
SPECIFIC_SERVICE="$1"

# All 17 active services
SERVICES=(
    "monitoring-service"
    "notification-service"
    "tenant-admin-service"
    "user-service"
    "incident-service"
    "component-service"
    "payment-service"
    "analytics-service"
    "event-store-service"
    "branding-service"
    "status-ui-service"
    "audit-consumer"
    "billing-consumer"
    "notification-consumer"
    "analytics-consumer"
    "landing-page-service"
    "saas-admin-service"
)

# If specific service provided
if [ -n "$SPECIFIC_SERVICE" ]; then
    SERVICES=("$SPECIFIC_SERVICE")
    echo -e "${BLUE}Testing single service: $SPECIFIC_SERVICE${NC}"
fi

# Test environment configuration
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=testpass
export DB_SSLMODE=disable
export REDIS_HOST=localhost
export REDIS_PORT=6379
export RABBITMQ_HOST=localhost
export RABBITMQ_PORT=5672
export RABBITMQ_USER=test
export RABBITMQ_PASSWORD=testpass
export MOCKSERVER_URL=http://localhost:1080

# Check if test environment is running
echo -e "${YELLOW}Checking test environment...${NC}"

# Check PostgreSQL
if ! psql -h localhost -U postgres -d postgres -c "SELECT 1" > /dev/null 2>&1; then
    echo -e "${RED}Error: PostgreSQL not running${NC}"
    echo "Start test environment first:"
    echo "  ./scripts/test/setup-test-env.sh"
    exit 1
fi
echo -e "${GREEN}✓ PostgreSQL${NC}"

# Check Redis
if ! redis-cli -h localhost ping > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠ Redis not running (tests may skip Redis features)${NC}"
else
    echo -e "${GREEN}✓ Redis${NC}"
fi

# Check RabbitMQ
if ! curl -s -u test:testpass http://localhost:15672/api/overview > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠ RabbitMQ not running (consumer tests may fail)${NC}"
else
    echo -e "${GREEN}✓ RabbitMQ${NC}"
fi

echo ""

# Test results tracking
TOTAL=0
PASSED=0
FAILED=0
SKIPPED=0
FAILED_SERVICES=()

# Function to run integration tests
run_integration_tests() {
    local service=$1
    local service_dir="microservices/$service"

    echo -e "${YELLOW}Testing $service...${NC}"

    # Check if service exists
    if [ ! -d "$service_dir" ]; then
        echo -e "${RED}  ✗ Service directory not found${NC}"
        return 1
    fi

    # Check if integration tests exist
    if [ ! -d "$service_dir/tests/integration" ]; then
        echo -e "${YELLOW}  ⚠ No integration tests found (skipping)${NC}"
        SKIPPED=$((SKIPPED + 1))
        return 2
    fi

    cd "$service_dir"

    # Set service-specific DB name
    case $service in
        "saas-admin-service")
            export DB_NAME=saas_admin
            ;;
        "tenant-admin-service")
            export DB_NAME=tenant_admin_db
            ;;
        "user-service")
            export DB_NAME=statuspage_user
            ;;
        "component-service")
            export DB_NAME=statuspage_component
            ;;
        "incident-service")
            export DB_NAME=statuspage_incident
            ;;
        "notification-service")
            export DB_NAME=statuspage_notification
            ;;
        "payment-service")
            export DB_NAME=statuspage_payment
            ;;
        "analytics-service")
            export DB_NAME=statuspage_analytics
            ;;
        "monitoring-service")
            export DB_NAME=statuspage_monitoring
            ;;
        "event-store-service")
            export DB_NAME=statuspage_eventstore
            ;;
        "branding-service")
            export DB_NAME=statuspage_branding
            ;;
        "status-ui-service")
            export DB_NAME=statuspage_ui
            ;;
        "landing-page-service")
            export DB_NAME=statuspage_landing
            ;;
        "audit-consumer")
            export DB_NAME=statuspage_audit
            ;;
        *)
            export DB_NAME=postgres
            ;;
    esac

    # Run integration tests
    if go test ./tests/integration/... -v -race -timeout 30m > test.log 2>&1; then
        TESTS_RUN=$(grep -c "^=== RUN" test.log || echo "0")
        echo -e "${GREEN}  ✓ Integration tests passed ($TESTS_RUN tests)${NC}"
        cd - > /dev/null
        return 0
    else
        echo -e "${RED}  ✗ Integration tests failed${NC}"
        echo -e "${YELLOW}  Last 20 lines of output:${NC}"
        tail -20 test.log
        cd - > /dev/null
        return 1
    fi
}

# Run integration tests for all services
echo -e "${BLUE}Running integration tests for ${#SERVICES[@]} service(s)...${NC}"
echo ""

for service in "${SERVICES[@]}"; do
    TOTAL=$((TOTAL + 1))

    result=$(run_integration_tests "$service"; echo $?)

    if [ "$result" -eq 0 ]; then
        PASSED=$((PASSED + 1))
    elif [ "$result" -eq 2 ]; then
        # Skipped (no tests)
        continue
    else
        FAILED=$((FAILED + 1))
        FAILED_SERVICES+=("$service")
    fi

    echo ""
done

# Summary
echo "======================================"
echo -e "${BLUE}Integration Test Summary${NC}"
echo "======================================"
echo "Total services:  $TOTAL"
echo -e "${GREEN}Passed:         $PASSED${NC}"
echo -e "${YELLOW}Skipped:        $SKIPPED${NC}"
if [ $FAILED -gt 0 ]; then
    echo -e "${RED}Failed:         $FAILED${NC}"
    echo ""
    echo "Failed services:"
    for service in "${FAILED_SERVICES[@]}"; do
        echo -e "${RED}  - $service${NC}"
    done
else
    echo -e "${GREEN}Failed:         0${NC}"
fi
echo "======================================"
echo ""

# Exit with error if any tests failed
if [ $FAILED -gt 0 ]; then
    echo -e "${RED}❌ Some integration tests failed${NC}"
    exit 1
else
    echo -e "${GREEN}✅ All integration tests passed!${NC}"
    exit 0
fi
