#!/bin/bash
#
# SaaS Admin Service - Database Initialization
# Creates database and runs migrations
#
# Usage:
#   ./init-db.sh                    # Use defaults (localhost)
#   DB_HOST=rds-host ./init-db.sh   # Custom host
#

set -e

# Service-specific configuration
SERVICE_NAME="saas-admin-service"
DB_NAME="${DB_NAME:-saas_admin}"
MIGRATIONS_DIR="./migrations"

# Database connection (use env vars or defaults)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_SSLMODE="${DB_SSLMODE:-disable}"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}=== $SERVICE_NAME Database Initialization ===${NC}"
echo "Database: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"
echo ""

# Function to check if PostgreSQL is accessible
check_postgres() {
    if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c '\q' 2>/dev/null; then
        echo -e "${RED}✗ Cannot connect to PostgreSQL${NC}"
        echo "Please check your connection parameters"
        exit 1
    fi
    echo -e "${GREEN}✓ PostgreSQL connection successful${NC}"
}

# Function to create database
create_database() {
    local exists=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'")

    if [ "$exists" = "1" ]; then
        echo -e "${YELLOW}Database '$DB_NAME' already exists${NC}"
    else
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;"
        echo -e "${GREEN}✓ Database '$DB_NAME' created${NC}"
    fi
}

# Function to run migrations
run_migrations() {
    if [ ! -d "$MIGRATIONS_DIR" ]; then
        echo -e "${RED}✗ Migrations directory not found: $MIGRATIONS_DIR${NC}"
        exit 1
    fi

    local migration_count=$(find "$MIGRATIONS_DIR" -name "*.sql" -type f 2>/dev/null | wc -l | tr -d ' ')

    if [ "$migration_count" -eq 0 ]; then
        echo -e "${YELLOW}⚠ No migration files found${NC}"
        return 0
    fi

    echo -e "${YELLOW}Running $migration_count migration(s)...${NC}"

    for migration in $(ls -1 $MIGRATIONS_DIR/*.sql 2>/dev/null | sort); do
        echo -e "${YELLOW}  Applying: $(basename $migration)${NC}"

        if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$migration" -v ON_ERROR_STOP=0 2>&1 | grep -v "ERROR.*already exists" > /tmp/migration_output.log; then
            echo -e "${GREEN}  ✓ Applied successfully${NC}"
        else
            # Check if errors are only "already exists" errors
            if grep -q "ERROR" /tmp/migration_output.log && ! grep -q "already exists" /tmp/migration_output.log; then
                echo -e "${RED}  ✗ Migration failed${NC}"
                cat /tmp/migration_output.log
                exit 1
            else
                echo -e "${GREEN}  ✓ Already applied (skipped)${NC}"
            fi
        fi
    done

    echo -e "${GREEN}✓ All migrations completed${NC}"
}

# Function to show database status
show_status() {
    local table_count=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'" 2>/dev/null || echo "0")

    echo ""
    echo -e "${GREEN}=== Database Status ===${NC}"
    echo "Tables: $table_count"

    if [ "$table_count" -gt 0 ]; then
        echo "Table list:"
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT '  - ' || tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename"
    fi
}

# Main execution
main() {
    check_postgres
    create_database
    run_migrations
    show_status

    echo ""
    echo -e "${GREEN}✓ $SERVICE_NAME database initialization complete${NC}"
}

main
