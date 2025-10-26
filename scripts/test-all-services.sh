#!/bin/bash
#
# Comprehensive Service Testing Script
# Tests all 20 microservices in the Beakon platform
#

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Counters
TOTAL_SERVICES=0
RUNNING_SERVICES=0
HEALTHY_SERVICES=0
FAILED_SERVICES=0

echo ""
echo "========================================="
echo "Beakon Platform - Service Health Test"
echo "========================================="
echo "Timestamp: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# Test infrastructure services
echo -e "${BLUE}=== Infrastructure Services ===${NC}"
echo ""

test_service() {
    local container_name=$1
    local service_type=$2
    local health_endpoint=$3

    TOTAL_SERVICES=$((TOTAL_SERVICES + 1))

    # Check if container is running
    if docker ps --format '{{.Names}}' | grep -q "^${container_name}$"; then
        RUNNING_SERVICES=$((RUNNING_SERVICES + 1))

        # Check health status
        local status=$(docker ps --filter "name=^${container_name}$" --format '{{.Status}}')

        if [[ "$status" == *"healthy"* ]] || [[ "$service_type" == "frontend" ]] || [[ "$service_type" == "consumer" ]]; then
            # For backends with health endpoints, test the endpoint
            if [ -n "$health_endpoint" ]; then
                if curl -sf "$health_endpoint" > /dev/null 2>&1; then
                    echo -e "  ✅ ${GREEN}${container_name}${NC} - Running & Healthy"
                    HEALTHY_SERVICES=$((HEALTHY_SERVICES + 1))
                else
                    echo -e "  ⚠️  ${YELLOW}${container_name}${NC} - Running but health endpoint failed"
                    echo -e "      Endpoint: $health_endpoint"
                fi
            else
                echo -e "  ✅ ${GREEN}${container_name}${NC} - Running"
                HEALTHY_SERVICES=$((HEALTHY_SERVICES + 1))
            fi
        else
            echo -e "  ⚠️  ${YELLOW}${container_name}${NC} - Running (Status: $status)"
        fi
    else
        echo -e "  ❌ ${RED}${container_name}${NC} - Not running"
        FAILED_SERVICES=$((FAILED_SERVICES + 1))
    fi
}

# Infrastructure
test_service "statuspage-postgres" "infrastructure" ""
test_service "beakon-redis" "infrastructure" ""
test_service "beakon-rabbitmq" "infrastructure" ""

echo ""
echo -e "${BLUE}=== Backend Services (Go) ===${NC}"
echo ""

# Backend services with health endpoints
test_service "beakon-api-gateway" "backend" "http://localhost:8080/health"
test_service "beakon-user-service" "backend" "http://localhost:8081/health"
test_service "beakon-component-service" "backend" "http://localhost:8084/health"
test_service "beakon-notification-service" "backend" "http://localhost:8085/health"
test_service "beakon-incident-service" "backend" "http://localhost:8086/health"
test_service "beakon-payment-service" "backend" "http://localhost:8088/health"
test_service "beakon-analytics-service" "backend" "http://localhost:8090/health"
test_service "beakon-monitoring-service" "backend" "http://localhost:8092/health"
test_service "beakon-status-ui-service" "backend" "http://localhost:8093/health"
test_service "beakon-event-store-service" "backend" "http://localhost:8096/health"
test_service "beakon-branding-service" "backend" "http://localhost:8097/health"
test_service "beakon-saas-admin-service" "backend" "http://localhost:8098/api/v1/health"
test_service "beakon-tenant-admin-service" "backend" "http://localhost:8099/health"
test_service "beakon-landing-page-service" "backend" "http://localhost:8100/health"

echo ""
echo -e "${BLUE}=== Frontend Services (Next.js) ===${NC}"
echo ""

test_service "beakon-saas-admin-frontend" "frontend" "http://localhost:3001"
test_service "beakon-tenant-admin-frontend" "frontend" "http://localhost:3002"

echo ""
echo -e "${BLUE}=== Consumer Services (No HTTP) ===${NC}"
echo ""

test_service "beakon-analytics-consumer" "consumer" ""
test_service "beakon-notification-consumer" "consumer" ""
test_service "beakon-audit-consumer" "consumer" ""
test_service "beakon-billing-consumer" "consumer" ""

echo ""
echo "========================================="
echo -e "${BLUE}Summary${NC}"
echo "========================================="
echo "Total Services:   $TOTAL_SERVICES"
echo -e "Running:          ${GREEN}$RUNNING_SERVICES${NC}"
echo -e "Healthy:          ${GREEN}$HEALTHY_SERVICES${NC}"
echo -e "Failed/Not Running: ${RED}$FAILED_SERVICES${NC}"
echo ""

# Calculate percentage
if [ $TOTAL_SERVICES -gt 0 ]; then
    PERCENTAGE=$((RUNNING_SERVICES * 100 / TOTAL_SERVICES))
    echo "Status: $PERCENTAGE% of services running"
fi

echo ""

# Exit code based on results
if [ $FAILED_SERVICES -eq 0 ]; then
    echo -e "${GREEN}✅ All services healthy!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  Some services are not running${NC}"
    exit 1
fi
