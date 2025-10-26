#!/bin/bash
#
# Start all microservices from docker-deployment folder
#

set -e

cd /Users/anuoamdutta/Desktop/statuspage/Beakon

GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}Starting All Services from docker-deployment${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# Define service groups
BACKEND_SERVICES=(
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
)

CONSUMER_SERVICES=(
    "analytics-consumer"
    "notification-consumer"
    "audit-consumer"
    "billing-consumer"
)

FRONTEND_SERVICES=(
    "saas-admin-frontend"
    "tenant-admin-frontend"
)

SUCCESS=0
FAILED=0
FAILED_SERVICES=()

start_service() {
    local service=$1
    local service_dir="docker-deployment/$service"

    if [ ! -d "$service_dir" ]; then
        echo -e "${RED}✗ $service - directory not found${NC}"
        FAILED=$((FAILED + 1))
        FAILED_SERVICES+=("$service (no directory)")
        return 1
    fi

    cd "$service_dir"

    # Try to start without building first (use existing images)
    if docker-compose up -d --no-build 2>&1 | grep -q "Error\|failed"; then
        echo -e "${YELLOW}  Retrying with build...${NC}"
        # If failed, try building from root docker-compose
        cd /Users/anuoamdutta/Desktop/statuspage/Beakon
        if docker-compose build "$service" >/dev/null 2>&1; then
            cd "$service_dir"
            if docker-compose up -d --no-build 2>&1; then
                echo -e "${GREEN}✓ $service started (after build)${NC}"
                SUCCESS=$((SUCCESS + 1))
            else
                echo -e "${RED}✗ $service failed to start${NC}"
                FAILED=$((FAILED + 1))
                FAILED_SERVICES+=("$service")
            fi
        else
            echo -e "${RED}✗ $service build failed${NC}"
            FAILED=$((FAILED + 1))
            FAILED_SERVICES+=("$service (build failed)")
        fi
    else
        echo -e "${GREEN}✓ $service started${NC}"
        SUCCESS=$((SUCCESS + 1))
    fi

    cd /Users/anuoamdutta/Desktop/statuspage/Beakon
}

echo -e "${BLUE}Starting Backend Services (14)...${NC}"
for service in "${BACKEND_SERVICES[@]}"; do
    echo -n "Starting $service... "
    start_service "$service"
done

echo ""
echo -e "${BLUE}Starting Consumer Services (4)...${NC}"
for service in "${CONSUMER_SERVICES[@]}"; do
    echo -n "Starting $service... "
    start_service "$service"
done

echo ""
echo -e "${BLUE}Starting Frontend Services (2)...${NC}"
for service in "${FRONTEND_SERVICES[@]}"; do
    echo -n "Starting $service... "
    start_service "$service"
done

echo ""
echo -e "${BLUE}=========================================${NC}"
echo -e "${GREEN}Startup Summary${NC}"
echo -e "${BLUE}=========================================${NC}"
echo "Total services: $((SUCCESS + FAILED))"
echo -e "${GREEN}Started successfully: $SUCCESS${NC}"
echo -e "${RED}Failed: $FAILED${NC}"

if [ $FAILED -gt 0 ]; then
    echo ""
    echo -e "${RED}Failed services:${NC}"
    for svc in "${FAILED_SERVICES[@]}"; do
        echo "  - $svc"
    done
fi

echo ""
echo -e "${BLUE}Waiting 10 seconds for services to start...${NC}"
sleep 10

echo ""
echo -e "${BLUE}Current Service Status:${NC}"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "beakon|statuspage|NAMES"

echo ""
if [ $SUCCESS -eq $((${#BACKEND_SERVICES[@]} + ${#CONSUMER_SERVICES[@]} + ${#FRONTEND_SERVICES[@]})) ]; then
    echo -e "${GREEN}✅ All services started successfully!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  Some services failed to start${NC}"
    exit 1
fi
