#!/bin/bash
set -e

SERVICE_NAME="monitoring-service"
DB_NAME="${DB_NAME:-monitoring}"
MIGRATIONS_DIR="./migrations"

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}=== $SERVICE_NAME Database Initialization ===${NC}"

check_postgres() {
    if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c '\q' 2>/dev/null; then
        echo -e "${RED}✗ Cannot connect to PostgreSQL${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ PostgreSQL connected${NC}"
}

create_database() {
    local exists=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'")
    if [ "$exists" = "1" ]; then
        echo -e "${YELLOW}Database exists${NC}"
    else
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;"
        echo -e "${GREEN}✓ Database created${NC}"
    fi
}

enable_extensions() {
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";" 2>&1 | grep -v "already exists" || true
}

run_migrations() {
    if [ ! -d "$MIGRATIONS_DIR" ]; then
        return 0
    fi
    for migration in $(ls -1 $MIGRATIONS_DIR/*.sql 2>/dev/null | sort); do
        echo "  Applying: $(basename $migration)"
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$migration" -v ON_ERROR_STOP=0 2>&1 | grep -v "already exists" || true
    done
}

main() {
    check_postgres
    create_database
    enable_extensions
    run_migrations
    echo -e "${GREEN}✓ Complete${NC}"
}

main
