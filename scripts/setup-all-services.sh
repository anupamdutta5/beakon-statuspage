#!/bin/bash
#
# Generate atlas.hcl and init-db.sh for ALL services
#

set -e

cd /Users/anuoamdutta/Desktop/statuspage/Beakon

GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}Generating Atlas Configs for All Services${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# Service -> Database mapping
generate_config() {
    local service=$1
    local database=$2

    echo "Generating configs for $service ($database)..."

    # Generate atlas.hcl
    ./scripts/generate-atlas-config.sh "$service" "$database"

    # Generate init-db.sh
    ./scripts/generate-init-db.sh "$service" "$database"

    echo -e "${GREEN}✓${NC} $service configured"
    echo ""
}

# Backend services with databases
generate_config "user-service" "statuspage_user"
generate_config "component-service" "statuspage_component"
generate_config "notification-service" "statuspage_notification"
generate_config "incident-service" "statuspage_incident"
generate_config "payment-service" "statuspage_payment"
generate_config "analytics-service" "statuspage_analytics"
generate_config "monitoring-service" "statuspage_monitoring"
generate_config "status-ui-service" "statuspage_status"
generate_config "event-store-service" "statuspage_events"
generate_config "branding-service" "statuspage_branding"
generate_config "landing-page-service" "statuspage_landing"

# Consumer services
generate_config "audit-consumer" "statuspage_audit"
generate_config "analytics-consumer" "statuspage_analytics"
generate_config "notification-consumer" "statuspage_notification"
generate_config "billing-consumer" "statuspage_billing"

echo ""
echo -e "${BLUE}=========================================${NC}"
echo -e "${GREEN}✅ All services configured!${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""
echo "Generated:"
echo "  - atlas.hcl for 15 services"
echo "  - init-db.sh for 15 services"
echo ""
echo "Next steps:"
echo "  1. Review generated configs"
echo "  2. Run init-db.sh for each service to initialize databases"
