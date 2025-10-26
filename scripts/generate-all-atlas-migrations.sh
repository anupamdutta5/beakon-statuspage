#!/bin/bash
#
# Generate Atlas migrations from GORM models for ALL services
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
echo -e "${BLUE}Generate Atlas Migrations from GORM Models${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

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
    "saas-admin-service"
    "tenant-admin-service"
    "audit-consumer"
    "analytics-consumer"
)

SUCCESS=0
FAILED=0

for service in "${SERVICES[@]}"; do
    echo -e "${BLUE}=========================================${NC}"
    echo -e "${BLUE}$service${NC}"
    echo -e "${BLUE}=========================================${NC}"

    if [ -d "microservices/$service/internal/models" ]; then
        cd "microservices/$service"

        echo "Generating migration from GORM models..."
        if atlas migrate diff initial_schema --env dev 2>&1 | tee /tmp/${service}_atlas.log; then
            SUCCESS=$((SUCCESS + 1))
            echo -e "${GREEN}✅ Migration generated for $service${NC}"
        else
            FAILED=$((FAILED + 1))
            echo -e "${RED}❌ Failed to generate migration for $service${NC}"
            echo "Error log:"
            cat /tmp/${service}_atlas.log
        fi

        cd ../..
    else
        echo -e "${YELLOW}⚠️  No models directory found for $service${NC}"
        FAILED=$((FAILED + 1))
    fi

    echo ""
done

echo ""
echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}Summary${NC}"
echo -e "${BLUE}=========================================${NC}"
echo "Total Services:    ${#SERVICES[@]}"
echo -e "${GREEN}Successful:        $SUCCESS${NC}"
echo -e "${RED}Failed:            $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ ALL migrations generated successfully!${NC}"
    exit 0
else
    echo -e "${RED}⚠️  Some migrations failed to generate${NC}"
    exit 1
fi
