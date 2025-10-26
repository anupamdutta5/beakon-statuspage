#!/bin/bash
#
# Verify all config files are self-contained in docker-deployment folders
#

cd /Users/anuoamdutta/Desktop/statuspage/Beakon/docker-deployment

GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}Docker Deployment Config Verification${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

ALL_SERVICES=(
    # Infrastructure (3)
    "postgres"
    "redis"
    "rabbitmq"
    # Backend Services (14)
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
    # Consumers (4)
    "analytics-consumer"
    "notification-consumer"
    "audit-consumer"
    "billing-consumer"
    # Frontends (2)
    "saas-admin-frontend"
    "tenant-admin-frontend"
)

VERIFIED=0
MISSING=0
FAILED_SERVICES=()

for service in "${ALL_SERVICES[@]}"; do
    echo -n "Checking $service... "

    if [ ! -d "$service" ]; then
        echo -e "${RED}✗ Directory missing${NC}"
        MISSING=$((MISSING + 1))
        FAILED_SERVICES+=("$service (no directory)")
        continue
    fi

    if [ ! -d "$service/configs" ]; then
        echo -e "${RED}✗ No configs folder${NC}"
        MISSING=$((MISSING + 1))
        FAILED_SERVICES+=("$service (no configs)")
        continue
    fi

    CONFIG_COUNT=$(ls -1 "$service/configs/" 2>/dev/null | wc -l | tr -d ' ')
    if [ "$CONFIG_COUNT" -eq 0 ]; then
        echo -e "${RED}✗ Empty configs folder${NC}"
        MISSING=$((MISSING + 1))
        FAILED_SERVICES+=("$service (empty configs)")
        continue
    fi

    echo -e "${GREEN}✓ $CONFIG_COUNT config files${NC}"
    VERIFIED=$((VERIFIED + 1))
done

echo ""
echo -e "${BLUE}=========================================${NC}"
echo -e "${GREEN}Verification Summary${NC}"
echo -e "${BLUE}=========================================${NC}"
echo "Total services: ${#ALL_SERVICES[@]}"
echo -e "${GREEN}Verified: $VERIFIED${NC}"
echo -e "${RED}Missing/Failed: $MISSING${NC}"

if [ $MISSING -gt 0 ]; then
    echo ""
    echo -e "${RED}Failed services:${NC}"
    for svc in "${FAILED_SERVICES[@]}"; do
        echo "  - $svc"
    done
    echo ""
    echo -e "${RED}❌ Verification FAILED${NC}"
    exit 1
else
    echo ""
    echo -e "${GREEN}✅ All services have self-contained configs!${NC}"
    echo ""
    echo "Config files are properly located in docker-deployment folders."
    echo "Services will read configs from mounted volumes."
    echo "No dependency on service git repositories."
    exit 0
fi
