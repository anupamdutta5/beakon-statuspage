#!/bin/bash

# Script to apply SSO migration files to tenant_admin_db
# Service: tenant-admin-service
# Date: 2025-10-29

set -e  # Exit on error

# Database configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="tenant_admin_db"

echo "========================================="
echo "SSO Migrations for tenant-admin-service"
echo "========================================="
echo "Database: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"
echo "User: $DB_USER"
echo ""

# Check if database exists
export PGPASSWORD="$DB_PASSWORD"
if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
    echo "❌ ERROR: Database '$DB_NAME' does not exist"
    echo "Please run init-db.sh first"
    exit 1
fi

# Array of migration files in order
MIGRATIONS=(
    "20251029_01_add_sso_fields_to_users.sql"
    "20251029_02_create_sso_providers.sql"
    "20251029_03_create_sso_user_identities.sql"
    "20251029_04_create_saml_requests.sql"
    "20251029_05_create_sso_audit_logs.sql"
)

# Apply each migration
for MIGRATION in "${MIGRATIONS[@]}"; do
    echo "Applying: $MIGRATION"

    if [ ! -f "$MIGRATION" ]; then
        echo "❌ ERROR: Migration file not found: $MIGRATION"
        exit 1
    fi

    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$MIGRATION" -v ON_ERROR_STOP=1; then
        echo "✅ $MIGRATION applied successfully"
    else
        echo "❌ ERROR: Failed to apply $MIGRATION"
        exit 1
    fi
    echo ""
done

echo "========================================="
echo "✅ All SSO migrations applied successfully!"
echo "========================================="
echo ""

# Verify tables were created
echo "Verifying SSO tables:"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "\dt sso_*" -c "\dt saml_*"

echo ""
echo "SSO Migration Complete!"
echo ""
echo "New tables created:"
echo "  - sso_providers (SSO configurations)"
echo "  - sso_user_identities (user-IdP links)"
echo "  - saml_requests (pending SAML requests)"
echo "  - sso_audit_logs (audit trail)"
echo ""
echo "Existing table updated:"
echo "  - users (added auth_method, sso_provider_id, is_sso_user)"
echo ""
