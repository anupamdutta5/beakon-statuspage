#!/bin/bash
#
# Initialize ALL service databases
# This creates all databases needed by the platform
#

set -e

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}Initializing ALL Beakon Databases${NC}"
echo -e "${BLUE}=========================================${NC}"
echo "Host: $DB_HOST:$DB_PORT"
echo ""

# Database list (service -> database_name)
declare -A DATABASES=(
    ["user-service"]="statuspage_user"
    ["component-service"]="statuspage_component"
    ["notification-service"]="statuspage_notification"
    ["incident-service"]="statuspage_incident"
    ["payment-service"]="statuspage_payment"
    ["analytics-service"]="statuspage_analytics"
    ["monitoring-service"]="statuspage_monitoring"
    ["status-ui-service"]="statuspage_status"
    ["event-store-service"]="statuspage_events"
    ["branding-service"]="statuspage_branding"
    ["landing-page-service"]="statuspage_landing"
    ["audit-consumer"]="statuspage_audit"
    ["analytics-consumer"]="statuspage_analytics"
    ["notification-consumer"]="statuspage_notification"
    ["billing-consumer"]="statuspage_billing"
)

# Already initialized
echo -e "${GREEN}Already initialized:${NC}"
echo "  - saas_admin (saas-admin-service)"
echo "  - tenant_admin_db (tenant-admin-service)"
echo ""

CREATED=0
SKIPPED=0

for service in "${!DATABASES[@]}"; do
    db_name="${DATABASES[$service]}"

    # Check if database exists
    exists=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$db_name'" 2>/dev/null || echo "")

    if [ "$exists" = "1" ]; then
        echo -e "  ${YELLOW}⊘${NC} $db_name (already exists)"
        SKIPPED=$((SKIPPED + 1))
    else
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $db_name;" > /dev/null 2>&1
        echo -e "  ${GREEN}✓${NC} $db_name (created)"
        CREATED=$((CREATED + 1))
    fi
done

echo ""
echo -e "${BLUE}=========================================${NC}"
echo -e "${GREEN}Summary${NC}"
echo -e "${BLUE}=========================================${NC}"
echo "Databases created: $CREATED"
echo "Already existed:   $SKIPPED"
echo ""
echo -e "${GREEN}✅ All databases initialized${NC}"
echo ""
echo "Note: Schemas will be created by init-db.sh scripts"
echo "      or by AutoMigrate on service startup"
