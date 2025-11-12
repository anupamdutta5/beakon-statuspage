#!/bin/bash

# run-e2e-tests.sh
# Run end-to-end tests for critical user flows
# Requires all services to be running
# Usage: ./scripts/test/run-e2e-tests.sh [flow-name]

set -e

echo "=== Beakon End-to-End Tests ==="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Specific flow to test (optional)
SPECIFIC_FLOW="$1"

# Test environment configuration
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=testpass
export REDIS_HOST=localhost
export REDIS_PORT=6379
export RABBITMQ_HOST=localhost
export RABBITMQ_PORT=5672

# Service URLs
export SAAS_ADMIN_URL=http://localhost:8098
export TENANT_ADMIN_URL=http://localhost:8099
export USER_SERVICE_URL=http://localhost:8081
export COMPONENT_SERVICE_URL=http://localhost:8084
export INCIDENT_SERVICE_URL=http://localhost:8086
export NOTIFICATION_SERVICE_URL=http://localhost:8085
export MONITORING_SERVICE_URL=http://localhost:8092
export STATUS_UI_URL=http://localhost:8093
export LANDING_PAGE_URL=http://localhost:8100

# E2E test flows
E2E_FLOWS=(
    "tenant_onboarding"
    "monitoring_alerting"
    "status_page_viewing"
    "session_resilience"
)

# If specific flow provided
if [ -n "$SPECIFIC_FLOW" ]; then
    E2E_FLOWS=("$SPECIFIC_FLOW")
    echo -e "${BLUE}Testing single flow: $SPECIFIC_FLOW${NC}"
fi

# Check if required services are running
echo -e "${YELLOW}Checking required services...${NC}"

check_service() {
    local name=$1
    local url=$2

    if curl -s -f "$url/health" > /dev/null 2>&1 || curl -s -f "$url" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ $name${NC}"
        return 0
    else
        echo -e "${RED}✗ $name (not responding at $url)${NC}"
        return 1
    fi
}

SERVICES_OK=true

check_service "saas-admin-service" "$SAAS_ADMIN_URL" || SERVICES_OK=false
check_service "tenant-admin-service" "$TENANT_ADMIN_URL" || SERVICES_OK=false
check_service "component-service" "$COMPONENT_SERVICE_URL" || SERVICES_OK=false
check_service "incident-service" "$INCIDENT_SERVICE_URL" || SERVICES_OK=false
check_service "notification-service" "$NOTIFICATION_SERVICE_URL" || SERVICES_OK=false
check_service "monitoring-service" "$MONITORING_SERVICE_URL" || SERVICES_OK=false
check_service "status-ui-service" "$STATUS_UI_URL" || SERVICES_OK=false

if [ "$SERVICES_OK" = false ]; then
    echo ""
    echo -e "${RED}Error: Not all required services are running${NC}"
    echo "Please start services first:"
    echo "  cd microservices && ./start-all-services.sh"
    exit 1
fi

echo ""

# Test results tracking
TOTAL=0
PASSED=0
FAILED=0
FAILED_FLOWS=()

# Function to run E2E test flow
run_e2e_flow() {
    local flow=$1
    local test_file="tests/e2e/${flow}_test.go"

    echo -e "${YELLOW}Running E2E flow: $flow...${NC}"

    # Check if test file exists
    if [ ! -f "$test_file" ]; then
        echo -e "${YELLOW}  ⚠ Test file not found: $test_file (skipping)${NC}"
        echo -e "${BLUE}  📝 Create test file at: $test_file${NC}"
        return 2
    fi

    # Run the E2E test
    if go test "./$test_file" -v -timeout 30m > test.log 2>&1; then
        echo -e "${GREEN}  ✓ E2E flow passed${NC}"
        return 0
    else
        echo -e "${RED}  ✗ E2E flow failed${NC}"
        echo -e "${YELLOW}  Last 30 lines of output:${NC}"
        tail -30 test.log
        return 1
    fi
}

# Run E2E tests
echo -e "${BLUE}Running ${#E2E_FLOWS[@]} E2E flow(s)...${NC}"
echo ""

for flow in "${E2E_FLOWS[@]}"; do
    TOTAL=$((TOTAL + 1))

    result=$(run_e2e_flow "$flow"; echo $?)

    if [ "$result" -eq 0 ]; then
        PASSED=$((PASSED + 1))
    elif [ "$result" -eq 2 ]; then
        # Skipped (test not implemented)
        continue
    else
        FAILED=$((FAILED + 1))
        FAILED_FLOWS+=("$flow")
    fi

    echo ""
done

# Summary
echo "======================================"
echo -e "${BLUE}E2E Test Summary${NC}"
echo "======================================"
echo "Total flows:     $TOTAL"
echo -e "${GREEN}Passed:         $PASSED${NC}"
if [ $FAILED -gt 0 ]; then
    echo -e "${RED}Failed:         $FAILED${NC}"
    echo ""
    echo "Failed flows:"
    for flow in "${FAILED_FLOWS[@]}"; do
        echo -e "${RED}  - $flow${NC}"
    done
else
    echo -e "${GREEN}Failed:         0${NC}"
fi
echo "======================================"
echo ""

# Exit with error if any flows failed
if [ $FAILED -gt 0 ]; then
    echo -e "${RED}❌ Some E2E flows failed${NC}"
    exit 1
else
    echo -e "${GREEN}✅ All E2E flows passed!${NC}"
    exit 0
fi
