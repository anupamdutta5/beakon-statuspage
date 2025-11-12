#!/bin/bash

# seed-test-data.sh
# Populate test databases with fixtures for testing
# Usage: ./scripts/test/seed-test-data.sh [--reset]

set -e

echo "=== Beakon Test Data Seeding ==="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Check if --reset flag is provided
RESET=false
if [[ "$1" == "--reset" ]]; then
    RESET=true
    echo -e "${YELLOW}Reset mode: Will drop and recreate all test data${NC}"
fi

# Database connection details
export PGHOST=localhost
export PGPORT=5432
export PGUSER=postgres
export PGPASSWORD=testpass

# Function to execute SQL file
execute_sql() {
    local db=$1
    local file=$2

    if [ -f "$file" ]; then
        echo -n "  Loading $(basename $file)..."
        psql -d "$db" -f "$file" > /dev/null 2>&1
        echo -e " ${GREEN}✓${NC}"
    else
        echo -e " ${YELLOW}⚠ File not found: $file${NC}"
    fi
}

# Function to seed database
seed_database() {
    local db=$1
    local fixture_dir="test-data/sql/$db"

    echo -e "${YELLOW}Seeding $db...${NC}"

    if [ "$RESET" = true ]; then
        echo -n "  Resetting database..."
        psql -d "$db" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;" > /dev/null 2>&1
        echo -e " ${GREEN}✓${NC}"
    fi

    # Load schema if exists
    if [ -f "$fixture_dir/schema.sql" ]; then
        execute_sql "$db" "$fixture_dir/schema.sql"
    fi

    # Load seed data if exists
    if [ -f "$fixture_dir/seed.sql" ]; then
        execute_sql "$db" "$fixture_dir/seed.sql"
    fi

    # Load test fixtures if exists
    if [ -f "$fixture_dir/fixtures.sql" ]; then
        execute_sql "$db" "$fixture_dir/fixtures.sql"
    fi
}

echo -e "${YELLOW}Checking PostgreSQL connection...${NC}"
if ! psql -d postgres -c "SELECT 1" > /dev/null 2>&1; then
    echo -e "${RED}Error: Cannot connect to PostgreSQL${NC}"
    echo "Make sure test environment is running:"
    echo "  ./scripts/test/setup-test-env.sh"
    exit 1
fi
echo -e "${GREEN}✓ Connected to PostgreSQL${NC}"
echo ""

# Seed all databases
echo -e "${YELLOW}Seeding databases...${NC}"
echo ""

# Core databases
seed_database "saas_admin"
seed_database "tenant_admin_db"
seed_database "statuspage_user"
seed_database "statuspage_component"
seed_database "statuspage_incident"
seed_database "statuspage_notification"
seed_database "statuspage_payment"
seed_database "statuspage_analytics"
seed_database "statuspage_monitoring"
seed_database "statuspage_eventstore"
seed_database "statuspage_branding"
seed_database "statuspage_ui"
seed_database "statuspage_landing"
seed_database "statuspage_audit"

echo ""
echo -e "${GREEN}=== Test Data Seeding Complete! ===${NC}"
echo ""
echo "Test data loaded for 14 databases"
echo ""
echo "To reset all data:"
echo "  ./scripts/test/seed-test-data.sh --reset"
echo ""
