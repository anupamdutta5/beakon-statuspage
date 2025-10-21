#!/bin/bash
#
# Database Initialization Template for Beakon Services
# Copy this file to each service directory and customize the configuration section
#
# Usage:
#   ./init-db.sh                    # Use defaults (localhost)
#   DB_HOST=rds-host ./init-db.sh   # Custom host
#

set -e

# ============================================================================
# SERVICE-SPECIFIC CONFIGURATION - CUSTOMIZE FOR EACH SERVICE
# ============================================================================
SERVICE_NAME="REPLACE_WITH_SERVICE_NAME"        # e.g., "component-service"
DB_NAME="REPLACE_WITH_DATABASE_NAME"            # e.g., "statuspage_component"
MIGRATIONS_DIR="./migrations"

# ============================================================================
# DATABASE CONNECTION (use env vars or defaults)
# ============================================================================
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_SSLMODE="${DB_SSLMODE:-disable}"

# ============================================================================
# COLORS FOR OUTPUT
# ============================================================================
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# ============================================================================
# FUNCTIONS
# ============================================================================

echo -e "${GREEN}=== $SERVICE_NAME Database Initialization ===${NC}"
echo "Database: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"
echo ""

# Check if PostgreSQL is accessible
check_postgres() {
    if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c '\q' 2>/dev/null; then
        echo -e "${RED}✗ Cannot connect to PostgreSQL${NC}"
        echo "Please check your connection parameters"
        echo "  Host: $DB_HOST"
        echo "  Port: $DB_PORT"
        echo "  User: $DB_USER"
        exit 1
    fi
    echo -e "${GREEN}✓ PostgreSQL connection successful${NC}"
}

# Create database if not exists
create_database() {
    local exists=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'")

    if [ "$exists" = "1" ]; then
        echo -e "${YELLOW}Database '$DB_NAME' already exists${NC}"
    else
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;"
        echo -e "${GREEN}✓ Database '$DB_NAME' created${NC}"
    fi
}

# Enable required extensions
enable_extensions() {
    echo -e "${BLUE}Enabling PostgreSQL extensions...${NC}"

    # UUID generation extension
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";" 2>&1 | grep -v "already exists" || true

    echo -e "${GREEN}✓ Extensions enabled${NC}"
}

# Run migrations in order
run_migrations() {
    if [ ! -d "$MIGRATIONS_DIR" ]; then
        echo -e "${YELLOW}⚠ No migrations directory found, skipping migrations${NC}"
        echo -e "${YELLOW}  Expected: $MIGRATIONS_DIR${NC}"
        return 0
    fi

    local migration_count=$(find "$MIGRATIONS_DIR" -name "*.sql" -type f 2>/dev/null | wc -l | tr -d ' ')

    if [ "$migration_count" -eq 0 ]; then
        echo -e "${YELLOW}⚠ No migration files found in $MIGRATIONS_DIR${NC}"
        return 0
    fi

    echo -e "${BLUE}Running $migration_count migration(s)...${NC}"

    for migration in $(ls -1 $MIGRATIONS_DIR/*.sql 2>/dev/null | sort); do
        local migration_name=$(basename $migration)
        echo -e "${YELLOW}  Applying: $migration_name${NC}"

        # Run migration with error handling
        if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME \
            -f "$migration" \
            -v ON_ERROR_STOP=0 \
            --single-transaction \
            2>&1 | tee /tmp/migration_${migration_name}.log | grep -E "ERROR|FATAL" > /tmp/migration_errors.log || true; then

            # Check if there were any real errors (not "already exists" errors)
            if [ -s /tmp/migration_errors.log ]; then
                if grep -v "already exists" /tmp/migration_errors.log | grep -q "ERROR\|FATAL"; then
                    echo -e "${RED}  ✗ Migration failed with errors${NC}"
                    cat /tmp/migration_errors.log
                    exit 1
                else
                    echo -e "${GREEN}  ✓ Already applied (skipped duplicate objects)${NC}"
                fi
            else
                echo -e "${GREEN}  ✓ Applied successfully${NC}"
            fi
        else
            echo -e "${GREEN}  ✓ Applied successfully${NC}"
        fi
    done

    # Cleanup temp files
    rm -f /tmp/migration_*.log /tmp/migration_errors.log

    echo -e "${GREEN}✓ All migrations completed${NC}"
}

# Show database status
show_status() {
    local table_count=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'" 2>/dev/null || echo "0")

    echo ""
    echo -e "${GREEN}=== Database Status ===${NC}"
    echo "Database: $DB_NAME"
    echo "Tables: $table_count"

    if [ "$table_count" -gt 0 ]; then
        echo ""
        echo "Table list:"
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT '  - ' || tablename || ' (' || n_live_tup || ' rows)' FROM pg_stat_user_tables ORDER BY tablename" 2>/dev/null || \
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT '  - ' || tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename"
    fi

    echo ""
    echo "Extensions:"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT '  - ' || extname || ' ' || extversion FROM pg_extension WHERE extname != 'plpgsql' ORDER BY extname"
}

# ============================================================================
# MAIN EXECUTION
# ============================================================================

main() {
    check_postgres
    create_database
    enable_extensions
    run_migrations
    show_status

    echo ""
    echo -e "${GREEN}✓ $SERVICE_NAME database initialization complete${NC}"
    echo ""
}

main
