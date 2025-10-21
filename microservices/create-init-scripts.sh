#!/bin/bash
# Script to create init-db.sh for all services that don't have one

# Define services and their database names
declare -A SERVICES
SERVICES[component-service]="statuspage_component"
SERVICES[incident-service]="statuspage_incident"
SERVICES[notification-service]="statuspage_notification"
SERVICES[monitoring-service]="statuspage_monitoring"
SERVICES[payment-service]="statuspage_payment"
SERVICES[branding-service]="statuspage_branding"
SERVICES[event-store-service]="statuspage_event_store"
SERVICES[analytics-service]="statuspage_analytics"
SERVICES[analytics-consumer]="statuspage_analytics_consumer"
SERVICES[audit-consumer]="statuspage_audit_consumer"
SERVICES[billing-consumer]="statuspage_billing_consumer"
SERVICES[notification-consumer]="statuspage_notification_consumer"

for service in "${!SERVICES[@]}"; do
    db_name="${SERVICES[$service]}"
    echo "Creating init-db.sh for $service (database: $db_name)"
    
    cat > "$service/init-db.sh" << INNER_EOF
#!/bin/bash
#
# $service - Database Initialization
# Creates database and runs migrations
#

set -e

SERVICE_NAME="$service"
DB_NAME="$db_name"
MIGRATIONS_DIR="./migrations"

DB_HOST="\${DB_HOST:-localhost}"
DB_PORT="\${DB_PORT:-5432}"
DB_USER="\${DB_USER:-postgres}"
DB_PASSWORD="\${DB_PASSWORD:-postgres}"
DB_SSLMODE="\${DB_SSLMODE:-disable}"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "\${GREEN}=== \$SERVICE_NAME Database Initialization ===\${NC}"
echo "Database: \$DB_NAME"
echo "Host: \$DB_HOST:\$DB_PORT"
echo ""

check_postgres() {
    if ! PGPASSWORD=\$DB_PASSWORD psql -h \$DB_HOST -p \$DB_PORT -U \$DB_USER -d postgres -c '\q' 2>/dev/null; then
        echo -e "\${RED}✗ Cannot connect to PostgreSQL\${NC}"
        exit 1
    fi
    echo -e "\${GREEN}✓ PostgreSQL connection successful\${NC}"
}

create_database() {
    local exists=\$(PGPASSWORD=\$DB_PASSWORD psql -h \$DB_HOST -p \$DB_PORT -U \$DB_USER -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='\$DB_NAME'")
    
    if [ "\$exists" = "1" ]; then
        echo -e "\${YELLOW}Database '\$DB_NAME' already exists\${NC}"
    else
        PGPASSWORD=\$DB_PASSWORD psql -h \$DB_HOST -p \$DB_PORT -U \$DB_USER -d postgres -c "CREATE DATABASE \$DB_NAME;"
        echo -e "\${GREEN}✓ Database '\$DB_NAME' created\${NC}"
    fi
}

enable_extensions() {
    echo -e "\${BLUE}Enabling PostgreSQL extensions...\${NC}"
    PGPASSWORD=\$DB_PASSWORD psql -h \$DB_HOST -p \$DB_PORT -U \$DB_USER -d \$DB_NAME -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";" 2>&1 | grep -v "already exists" || true
    echo -e "\${GREEN}✓ Extensions enabled\${NC}"
}

run_migrations() {
    if [ ! -d "\$MIGRATIONS_DIR" ]; then
        echo -e "\${YELLOW}⚠ No migrations directory found\${NC}"
        return 0
    fi
    
    local migration_count=\$(find "\$MIGRATIONS_DIR" -name "*.sql" -type f 2>/dev/null | wc -l | tr -d ' ')
    
    if [ "\$migration_count" -eq 0 ]; then
        echo -e "\${YELLOW}⚠ No migration files found\${NC}"
        return 0
    fi
    
    echo -e "\${BLUE}Running \$migration_count migration(s)...\${NC}"
    
    for migration in \$(ls -1 \$MIGRATIONS_DIR/*.sql 2>/dev/null | sort); do
        echo -e "\${YELLOW}  Applying: \$(basename \$migration)\${NC}"
        PGPASSWORD=\$DB_PASSWORD psql -h \$DB_HOST -p \$DB_PORT -U \$DB_USER -d \$DB_NAME -f "\$migration" -v ON_ERROR_STOP=0 --single-transaction 2>&1 | grep -v "already exists" || echo -e "\${GREEN}  ✓ Applied\${NC}"
    done
    
    echo -e "\${GREEN}✓ All migrations completed\${NC}"
}

show_status() {
    local table_count=\$(PGPASSWORD=\$DB_PASSWORD psql -h \$DB_HOST -p \$DB_PORT -U \$DB_USER -d \$DB_NAME -tAc "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'" 2>/dev/null || echo "0")
    
    echo ""
    echo -e "\${GREEN}=== Database Status ===\${NC}"
    echo "Database: \$DB_NAME"
    echo "Tables: \$table_count"
    
    if [ "\$table_count" -gt 0 ]; then
        echo "Table list:"
        PGPASSWORD=\$DB_PASSWORD psql -h \$DB_HOST -p \$DB_PORT -U \$DB_USER -d \$DB_NAME -tAc "SELECT '  - ' || tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename"
    fi
}

main() {
    check_postgres
    create_database
    enable_extensions
    run_migrations
    show_status
    
    echo ""
    echo -e "\${GREEN}✓ \$SERVICE_NAME database initialization complete\${NC}"
}

main
INNER_EOF
    
    chmod +x "$service/init-db.sh"
    echo "✓ Created and made executable: $service/init-db.sh"
done

echo ""
echo "✓ All init-db.sh scripts created"
