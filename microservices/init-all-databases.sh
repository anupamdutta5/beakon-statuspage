#!/bin/bash
#
# Master Database Initialization Orchestrator
# Runs individual per-service init-db.sh scripts for all microservices
#
# Usage:
#   All services:    ./init-all-databases.sh
#   Single service:  ./init-all-databases.sh landing-page-service
#   Production:      DB_HOST=rds-endpoint ./init-all-databases.sh
#

set -e  # Exit on error

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Export connection details for child scripts
export DB_HOST="${DB_HOST:-localhost}"
export DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-postgres}"
export DB_PASSWORD="${DB_PASSWORD:-postgres}"
export DB_SSLMODE="${DB_SSLMODE:-disable}"

echo -e "${GREEN}=== Beakon Master Database Initialization ===${NC}"
echo "Host: $DB_HOST"
echo "Port: $DB_PORT"
echo "User: $DB_USER"
echo ""
echo -e "${BLUE}Orchestrating per-service database initialization${NC}"
echo ""

# Service directories that have init-db.sh scripts
# Ordered by dependency: core services first, then feature services, then consumers
SERVICES=(
    # Core authentication & tenant management (no dependencies)
    "user-service"
    "tenant-admin-service"
    "saas-admin-service"

    # Core features (depend on tenant/user)
    "component-service"
    "incident-service"
    "notification-service"
    "monitoring-service"

    # Supporting services
    "payment-service"
    "branding-service"
    "event-store-service"
    "landing-page-service"

    # Analytics & consumers (depend on events)
    "analytics-service"
    "analytics-consumer"
    "audit-consumer"
    "billing-consumer"
    "notification-consumer"
)

# If specific service provided, run only that one
if [ -n "$1" ]; then
    SERVICE_TO_RUN="$1"
    echo -e "${YELLOW}Running initialization for: $SERVICE_TO_RUN${NC}"
    echo ""

    if [ -f "./$SERVICE_TO_RUN/init-db.sh" ]; then
        cd "./$SERVICE_TO_RUN"
        ./init-db.sh
        cd ..
        echo ""
        echo -e "${GREEN}✓ $SERVICE_TO_RUN initialization complete${NC}"
    else
        echo -e "${RED}✗ init-db.sh not found in $SERVICE_TO_RUN${NC}"
        exit 1
    fi
    exit 0
fi

# Run all service initializations
SUCCESS_COUNT=0
FAILED_COUNT=0
FAILED_SERVICES=()

for service in "${SERVICES[@]}"; do
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}Initializing: $service${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    if [ -f "./$service/init-db.sh" ]; then
        if (cd "./$service" && ./init-db.sh); then
            SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
            echo -e "${GREEN}✓ $service completed successfully${NC}"
        else
            FAILED_COUNT=$((FAILED_COUNT + 1))
            FAILED_SERVICES+=("$service")
            echo -e "${RED}✗ $service failed${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ init-db.sh not found in $service, skipping${NC}"
    fi

    echo ""
done

# Summary
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}=== Initialization Summary ===${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Successful: $SUCCESS_COUNT${NC}"

if [ $FAILED_COUNT -gt 0 ]; then
    echo -e "${RED}Failed: $FAILED_COUNT${NC}"
    echo -e "${RED}Failed services:${NC}"
    for failed in "${FAILED_SERVICES[@]}"; do
        echo -e "${RED}  - $failed${NC}"
    done
    exit 1
else
    echo -e "${GREEN}Failed: 0${NC}"
fi

echo ""
echo -e "${GREEN}✓ All database initializations complete${NC}"
echo ""
echo "Next steps:"
echo "  1. Start your microservices"
echo "  2. Verify database connections"
echo "  3. Run integration tests"
