#!/bin/bash
#
# Run init-db.sh for ALL services
# This initializes all databases using the holistic Atlas-based approach
#

set -e

cd /Users/anuoamdutta/Desktop/statuspage/Beakon

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo ""
echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}Holistic Database Initialization${NC}"
echo -e "${BLUE}Running init-db.sh for ALL Services${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# Services that have databases (in order)
SERVICES=(
    "user-service"
    "component-service"
    "notification-service"
    "incident-service"
    "payment-service"
    "analytics-service"
    "monitoring-service"
    "status-ui-service"
    "event-store-service"
    "branding-service"
    "landing-page-service"
    "audit-consumer"
    "analytics-consumer"
    "notification-consumer"
    "billing-consumer"
)

SUCCESS=0
FAILED=0
FAILED_SERVICES=()

for service in "${SERVICES[@]}"; do
    echo -e "${BLUE}=========================================${NC}"
    echo -e "${BLUE}[$((SUCCESS + FAILED + 1))/${#SERVICES[@]}] $service${NC}"
    echo -e "${BLUE}=========================================${NC}"

    if [ -f "microservices/$service/init-db.sh" ]; then
        cd "microservices/$service"

        if ./init-db.sh; then
            SUCCESS=$((SUCCESS + 1))
            echo -e "${GREEN}✅ $service database initialized${NC}"
        else
            FAILED=$((FAILED + 1))
            FAILED_SERVICES+=("$service")
            echo -e "${RED}❌ $service failed${NC}"
        fi

        cd ../..
    else
        echo -e "${RED}⚠️  init-db.sh not found for $service${NC}"
        FAILED=$((FAILED + 1))
        FAILED_SERVICES+=("$service")
    fi

    echo ""
done

echo ""
echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}Summary${NC}"
echo -e "${BLUE}=========================================${NC}"
echo -e "Total Services:    ${#SERVICES[@]}"
echo -e "${GREEN}Successful:        $SUCCESS${NC}"
echo -e "${RED}Failed:            $FAILED${NC}"

if [ $FAILED -gt 0 ]; then
    echo ""
    echo -e "${RED}Failed Services:${NC}"
    for failed_service in "${FAILED_SERVICES[@]}"; do
        echo "  - $failed_service"
    done
fi

echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ ALL databases initialized successfully!${NC}"
    exit 0
else
    echo -e "${RED}⚠️  Some databases failed to initialize${NC}"
    exit 1
fi
