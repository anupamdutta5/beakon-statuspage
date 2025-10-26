#!/bin/bash
#
# Create docker-deployment directory structure for all 20 microservices
#

set -e

cd /Users/anuoamdutta/Desktop/statuspage/Beakon

GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}Creating Docker Deployment Structure${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# Create main directory
mkdir -p docker-deployment

# All 20 services
SERVICES=(
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
    "saas-admin-frontend"
    "tenant-admin-frontend"
)

CREATED=0

for service in "${SERVICES[@]}"; do
    echo "Creating $service..."
    mkdir -p "docker-deployment/$service/configs"
    CREATED=$((CREATED + 1))
    echo -e "  ${GREEN}✓${NC} docker-deployment/$service/configs/"
done

echo ""
echo -e "${BLUE}=========================================${NC}"
echo -e "${GREEN}Summary${NC}"
echo -e "${BLUE}=========================================${NC}"
echo "Services configured: $CREATED"
echo ""
echo -e "${GREEN}✅ Directory structure created successfully!${NC}"
echo ""
echo "Next: Generate config.yml and service-endpoints.yml files"
