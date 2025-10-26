#!/bin/bash
#
# Add Viper dependency to all Go microservices
#

set -e

cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}Adding Viper to All Go Services${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# All Go services (17 total: 14 backends + 1 gateway + 2 consumers that don't have separate repos)
GO_SERVICES=(
    "api-gateway"
    "user-service"
    "tenant-admin-service"
    "saas-admin-service"
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
    "analytics-consumer"
    "notification-consumer"
    "audit-consumer"
    "billing-consumer"
)

SUCCESS=0
FAILED=0
SKIPPED=0

for service in "${GO_SERVICES[@]}"; do
    if [ ! -d "$service" ]; then
        echo -e "${RED}✗${NC} $service - Directory not found, skipping"
        SKIPPED=$((SKIPPED + 1))
        continue
    fi

    echo -e "${BLUE}Processing $service...${NC}"

    cd "$service"

    # Check if go.mod exists
    if [ ! -f "go.mod" ]; then
        echo -e "${RED}  ✗ No go.mod found, skipping${NC}"
        SKIPPED=$((SKIPPED + 1))
        cd ..
        continue
    fi

    # Add Viper dependency
    if go get github.com/spf13/viper@latest 2>&1 | tail -1; then
        echo -e "${GREEN}  ✓ Added Viper dependency${NC}"

        # Run go mod tidy
        if go mod tidy 2>&1 | tail -1; then
            echo -e "${GREEN}  ✓ Tidied dependencies${NC}"
            SUCCESS=$((SUCCESS + 1))
        else
            echo -e "${RED}  ✗ Failed to tidy dependencies${NC}"
            FAILED=$((FAILED + 1))
        fi
    else
        echo -e "${RED}  ✗ Failed to add Viper${NC}"
        FAILED=$((FAILED + 1))
    fi

    cd ..
    echo ""
done

echo ""
echo -e "${BLUE}=========================================${NC}"
echo -e "${GREEN}Summary${NC}"
echo -e "${BLUE}=========================================${NC}"
echo "Total services: ${#GO_SERVICES[@]}"
echo -e "${GREEN}Success: $SUCCESS${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
if [ $SKIPPED -gt 0 ]; then
    echo "Skipped: $SKIPPED"
fi
echo ""

if [ $SUCCESS -eq ${#GO_SERVICES[@]} ]; then
    echo -e "${GREEN}✅ All services updated successfully!${NC}"
else
    echo -e "${RED}⚠️  Some services failed to update${NC}"
fi
