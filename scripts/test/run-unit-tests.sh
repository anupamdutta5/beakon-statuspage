#!/bin/bash

# run-unit-tests.sh
# Run unit tests for all 17 active microservices
# Usage: ./scripts/test/run-unit-tests.sh [service-name] [--coverage]

set -e

echo "=== Beakon Unit Tests ==="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Flags
COVERAGE=false
SPECIFIC_SERVICE=""

# Parse arguments
for arg in "$@"; do
    case $arg in
        --coverage)
            COVERAGE=true
            shift
            ;;
        *)
            SPECIFIC_SERVICE=$arg
            shift
            ;;
    esac
done

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

# If specific service provided, test only that one
if [ -n "$SPECIFIC_SERVICE" ]; then
    SERVICES=("$SPECIFIC_SERVICE")
    echo -e "${BLUE}Testing single service: $SPECIFIC_SERVICE${NC}"
    echo ""
fi

# Test results tracking
TOTAL=0
PASSED=0
FAILED=0
FAILED_SERVICES=()

# Function to run unit tests for a service
run_unit_tests() {
    local service=$1
    local service_dir="microservices/$service"

    echo -e "${YELLOW}Testing $service...${NC}"

    # Check if service directory exists
    if [ ! -d "$service_dir" ]; then
        echo -e "${RED}  ✗ Service directory not found: $service_dir${NC}"
        return 1
    fi

    # Check if tests exist
    if [ ! -d "$service_dir/tests/unit" ] && ! ls "$service_dir"/*_test.go > /dev/null 2>&1; then
        echo -e "${YELLOW}  ⚠ No unit tests found (skipping)${NC}"
        return 0
    fi

    cd "$service_dir"

    # Run tests
    if [ "$COVERAGE" = true ]; then
        # Run with coverage
        if go test ./tests/unit/... -v -race -coverprofile=coverage.out -covermode=atomic > test.log 2>&1; then
            # Extract coverage percentage
            COVERAGE_PCT=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
            echo -e "${GREEN}  ✓ Tests passed (Coverage: $COVERAGE_PCT)${NC}"

            # Generate HTML coverage report
            go tool cover -html=coverage.out -o coverage.html
            echo -e "${BLUE}    Coverage report: $service_dir/coverage.html${NC}"

            cd - > /dev/null
            return 0
        else
            echo -e "${RED}  ✗ Tests failed${NC}"
            cat test.log
            cd - > /dev/null
            return 1
        fi
    else
        # Run without coverage
        if go test ./tests/unit/... -v -race > test.log 2>&1; then
            # Count test results
            TESTS_RUN=$(grep -c "^=== RUN" test.log || echo "0")
            echo -e "${GREEN}  ✓ Tests passed ($TESTS_RUN tests)${NC}"
            cd - > /dev/null
            return 0
        else
            echo -e "${RED}  ✗ Tests failed${NC}"
            cat test.log
            cd - > /dev/null
            return 1
        fi
    fi
}

# Run tests for all services
echo -e "${BLUE}Running unit tests for ${#SERVICES[@]} service(s)...${NC}"
echo ""

for service in "${SERVICES[@]}"; do
    TOTAL=$((TOTAL + 1))

    if run_unit_tests "$service"; then
        PASSED=$((PASSED + 1))
    else
        FAILED=$((FAILED + 1))
        FAILED_SERVICES+=("$service")
    fi

    echo ""
done

# Summary
echo "======================================"
echo -e "${BLUE}Unit Test Summary${NC}"
echo "======================================"
echo "Total services:  $TOTAL"
echo -e "${GREEN}Passed:         $PASSED${NC}"
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
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
else
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
fi
