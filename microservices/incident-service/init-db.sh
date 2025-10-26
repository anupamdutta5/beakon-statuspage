#!/bin/bash
#
# HOLISTIC Database Initialization Script Template
# This is a template - use scripts/generate-init-db.sh to create service-specific versions
#
# SAFETY FEATURES:
# - Never drops existing databases
# - Uses Atlas migration tracking (prevents re-application)
# - Validates PostgreSQL connectivity before proceeding
# - Separates schema migrations from seed data
# - Logs all operations for audit trail
#
# Usage:
#   ./init-db.sh                    # Use defaults (localhost)
#   DB_HOST=rds-host ./init-db.sh   # Custom host
#   FORCE_SEED=true ./init-db.sh    # Re-apply seed data (CAUTION)
#

set -e

# ============================================
# CONFIGURATION (will be replaced per service)
# ============================================
SERVICE_NAME="incident-service"
DB_NAME="statuspage_incident"
MIGRATIONS_DIR="./migrations"
SEEDS_DIR="./seeds"

# ============================================
# DATABASE CONNECTION (use env vars or defaults)
# ============================================
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_SSLMODE="${DB_SSLMODE:-disable}"
ATLAS_ENV="${ATLAS_ENV:-dev}"

# ============================================
# SAFETY FLAGS
# ============================================
FORCE_SEED="${FORCE_SEED:-false}"
DRY_RUN="${DRY_RUN:-false}"

# ============================================
# COLORS FOR OUTPUT
# ============================================
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# ============================================
# LOGGING
# ============================================
LOG_FILE="./init-db.log"

log() {
    echo -e "$1" | tee -a "$LOG_FILE"
}

log_info() {
    log "${BLUE}[INFO]${NC} $1"
}

log_success() {
    log "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    log "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    log "${RED}[ERROR]${NC} $1"
}

# ============================================
# HEADER
# ============================================
echo ""
log "${GREEN}========================================${NC}"
log "${GREEN}$SERVICE_NAME - Database Initialization${NC}"
log "${GREEN}========================================${NC}"
log "Database: $DB_NAME"
log "Host: $DB_HOST:$DB_PORT"
log "Atlas Environment: $ATLAS_ENV"
log "Timestamp: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# ============================================
# VALIDATION FUNCTIONS
# ============================================

# Check if PostgreSQL is accessible
check_postgres() {
    log_info "Checking PostgreSQL connectivity..."

    if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c '\q' 2>/dev/null; then
        log_error "Cannot connect to PostgreSQL"
        log_error "Please check your connection parameters:"
        log_error "  Host: $DB_HOST:$DB_PORT"
        log_error "  User: $DB_USER"
        exit 1
    fi

    log_success "PostgreSQL connection successful"
}

# Check if Atlas CLI is installed
check_atlas() {
    log_info "Checking Atlas CLI installation..."

    if ! command -v atlas &> /dev/null; then
        log_error "Atlas CLI not found"
        log_error "Install Atlas: https://atlasgo.io/getting-started"
        log_error "  macOS: brew install ariga/tap/atlas"
        log_error "  Linux: curl -sSf https://atlasgo.sh | sh"
        exit 1
    fi

    local atlas_version=$(atlas version | head -1)
    log_success "Atlas CLI installed: $atlas_version"
}

# ============================================
# DATABASE CREATION
# ============================================

create_database() {
    log_info "Checking if database exists..."

    local exists=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'")

    if [ "$exists" = "1" ]; then
        log_warning "Database '$DB_NAME' already exists (SAFE: Not dropping)"

        # Check if database has tables
        local table_count=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'" 2>/dev/null || echo "0")

        if [ "$table_count" -gt 0 ]; then
            log_warning "Database has $table_count existing table(s)"
            log_warning "Atlas will use migration tracking to avoid re-applying migrations"
        fi
    else
        log_info "Creating database '$DB_NAME'..."
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;"
        log_success "Database '$DB_NAME' created"
    fi
}

# ============================================
# SCHEMA MIGRATIONS (using Atlas)
# ============================================

apply_migrations() {
    log_info "Applying schema migrations using Atlas..."

    if [ ! -f "atlas.hcl" ]; then
        log_error "atlas.hcl not found"
        log_error "Generate it using: scripts/generate-atlas-config.sh $SERVICE_NAME $DB_NAME"
        exit 1
    fi

    if [ "$DRY_RUN" = "true" ]; then
        log_warning "DRY RUN MODE: Showing migration status only"
        atlas migrate status --env $ATLAS_ENV
        return 0
    fi

    # Apply migrations (Atlas tracks state automatically)
    log_info "Running: atlas migrate apply --env $ATLAS_ENV"

    if atlas migrate apply --env $ATLAS_ENV 2>&1 | tee -a "$LOG_FILE"; then
        log_success "Migrations applied successfully"
    else
        log_error "Migration failed"
        log_error "Check $LOG_FILE for details"
        exit 1
    fi

    # Show current migration status
    log_info "Current migration status:"
    atlas migrate status --env $ATLAS_ENV | tee -a "$LOG_FILE"
}

# ============================================
# SEED DATA (separate from migrations)
# ============================================

apply_seed_data() {
    if [ ! -d "$SEEDS_DIR" ]; then
        log_warning "Seeds directory not found: $SEEDS_DIR (skipping seed data)"
        return 0
    fi

    local seed_count=$(find "$SEEDS_DIR" -name "*.sql" -type f 2>/dev/null | wc -l | tr -d ' ')

    if [ "$seed_count" -eq 0 ]; then
        log_info "No seed data files found (skipping)"
        return 0
    fi

    log_info "Found $seed_count seed data file(s)"

    # Check if seeds have been applied before
    local seed_marker="$SEEDS_DIR/.applied"

    if [ -f "$seed_marker" ] && [ "$FORCE_SEED" != "true" ]; then
        log_warning "Seed data already applied (use FORCE_SEED=true to re-apply)"
        return 0
    fi

    if [ "$FORCE_SEED" = "true" ]; then
        log_warning "FORCE_SEED=true: Re-applying seed data (may cause duplicates)"
    fi

    log_info "Applying seed data..."

    for seed_file in $(ls -1 $SEEDS_DIR/*.sql 2>/dev/null | sort); do
        local seed_name=$(basename $seed_file)
        log_info "  Applying: $seed_name"

        if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$seed_file" >> "$LOG_FILE" 2>&1; then
            log_success "    ✓ $seed_name applied"
        else
            log_error "    ✗ $seed_name failed"
            log_error "Check $LOG_FILE for details"
            # Continue with other seeds instead of failing
        fi
    done

    # Mark seeds as applied
    date > "$seed_marker"
    log_success "Seed data applied successfully"
}

# ============================================
# STATUS REPORTING
# ============================================

show_status() {
    local table_count=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'" 2>/dev/null || echo "0")

    echo ""
    log "${GREEN}========================================${NC}"
    log "${GREEN}Database Status${NC}"
    log "${GREEN}========================================${NC}"
    log "Database: $DB_NAME"
    log "Tables: $table_count"

    if [ "$table_count" -gt 0 ]; then
        log ""
        log "Table List:"
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT '  - ' || tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename" | tee -a "$LOG_FILE"
    fi

    # Show migration tracking table
    local migration_count=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT COUNT(*) FROM atlas_schema_revisions" 2>/dev/null || echo "0")

    if [ "$migration_count" -gt 0 ]; then
        log ""
        log "Applied Migrations: $migration_count"
        log "(tracked in atlas_schema_revisions table)"
    fi

    echo ""
}

# ============================================
# MAIN EXECUTION
# ============================================

main() {
    # Step 1: Validate prerequisites
    check_postgres
    check_atlas

    # Step 2: Create database (SAFE: never drops)
    create_database

    # Step 3: Apply schema migrations (Atlas handles tracking)
    apply_migrations

    # Step 4: Apply seed data (optional, tracked separately)
    apply_seed_data

    # Step 5: Show final status
    show_status

    # Success
    echo ""
    log_success "$SERVICE_NAME database initialization complete"
    log "Log file: $LOG_FILE"
    echo ""
}

# Run main function
main
