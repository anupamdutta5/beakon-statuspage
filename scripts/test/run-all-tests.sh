#!/bin/bash

# run-all-tests.sh
# Run complete test suite: unit + integration + E2E
# Usage: ./scripts/test/run-all-tests.sh [--skip-e2e] [--coverage]

set -e

echo "=========================================="
echo "  Beakon Complete Test Suite"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Flags
SKIP_E2E=false
COVERAGE=false

# Parse arguments
for arg in "$@"; do
    case $arg in
        --skip-e2e)
            SKIP_E2E=true
            shift
            ;;
        --coverage)
            COVERAGE=true
            shift
            ;;
    esac
done

# Track start time
START_TIME=$(date +%s)

# Track results
UNIT_RESULT=0
INTEGRATION_RESULT=0
E2E_RESULT=0

echo -e "${BLUE}Test Suite Configuration:${NC}"
echo "  Coverage:  $COVERAGE"
echo "  Skip E2E:  $SKIP_E2E"
echo ""

# Phase 1: Unit Tests
echo "=========================================="
echo -e "${BLUE}Phase 1: Unit Tests${NC}"
echo "=========================================="
echo ""

if [ "$COVERAGE" = true ]; then
    ./scripts/test/run-unit-tests.sh --coverage || UNIT_RESULT=$?
else
    ./scripts/test/run-unit-tests.sh || UNIT_RESULT=$?
fi

echo ""

# Phase 2: Integration Tests
echo "=========================================="
echo -e "${BLUE}Phase 2: Integration Tests${NC}"
echo "=========================================="
echo ""

echo -e "${YELLOW}Ensuring test environment is running...${NC}"
./scripts/test/setup-test-env.sh

echo ""
./scripts/test/run-integration-tests.sh || INTEGRATION_RESULT=$?

echo ""

# Phase 3: E2E Tests (optional)
if [ "$SKIP_E2E" = false ]; then
    echo "=========================================="
    echo -e "${BLUE}Phase 3: End-to-End Tests${NC}"
    echo "=========================================="
    echo ""

    echo -e "${YELLOW}Note: E2E tests require all services to be running${NC}"
    echo -e "${YELLOW}If services are not running, E2E tests will be skipped${NC}"
    echo ""

    ./scripts/test/run-e2e-tests.sh || E2E_RESULT=$?

    echo ""
fi

# Calculate duration
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))
MINUTES=$((DURATION / 60))
SECONDS=$((DURATION % 60))

# Final Summary
echo "=========================================="
echo -e "${BLUE}Complete Test Suite Summary${NC}"
echo "=========================================="
echo ""

echo "Test Results:"
if [ $UNIT_RESULT -eq 0 ]; then
    echo -e "${GREEN}  ✓ Unit Tests:        PASSED${NC}"
else
    echo -e "${RED}  ✗ Unit Tests:        FAILED${NC}"
fi

if [ $INTEGRATION_RESULT -eq 0 ]; then
    echo -e "${GREEN}  ✓ Integration Tests: PASSED${NC}"
else
    echo -e "${RED}  ✗ Integration Tests: FAILED${NC}"
fi

if [ "$SKIP_E2E" = false ]; then
    if [ $E2E_RESULT -eq 0 ]; then
        echo -e "${GREEN}  ✓ E2E Tests:         PASSED${NC}"
    else
        echo -e "${RED}  ✗ E2E Tests:         FAILED${NC}"
    fi
else
    echo -e "${YELLOW}  ⊘ E2E Tests:         SKIPPED${NC}"
fi

echo ""
echo "Duration: ${MINUTES}m ${SECONDS}s"
echo ""

# Generate combined report (if coverage enabled)
if [ "$COVERAGE" = true ]; then
    echo -e "${BLUE}Generating combined coverage report...${NC}"

    # Create coverage directory
    mkdir -p coverage

    # Combine all coverage files
    echo "mode: atomic" > coverage/combined.out
    find microservices -name "coverage.out" -exec tail -n +2 {} \; >> coverage/combined.out

    # Generate HTML report
    go tool cover -html=coverage/combined.out -o coverage/combined.html

    # Calculate total coverage
    TOTAL_COVERAGE=$(go tool cover -func=coverage/combined.out | grep total | awk '{print $3}')

    echo -e "${GREEN}  Combined coverage: $TOTAL_COVERAGE${NC}"
    echo -e "${BLUE}  Report: coverage/combined.html${NC}"
    echo ""
fi

# Exit with appropriate code
if [ $UNIT_RESULT -ne 0 ] || [ $INTEGRATION_RESULT -ne 0 ] || [ $E2E_RESULT -ne 0 ]; then
    echo "=========================================="
    echo -e "${RED}❌ Test Suite FAILED${NC}"
    echo "=========================================="
    exit 1
else
    echo "=========================================="
    echo -e "${GREEN}✅ Test Suite PASSED${NC}"
    echo "=========================================="
    exit 0
fi
