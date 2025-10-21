# Database Setup and Migration Guide

This guide explains how to initialize and manage databases for all Beakon microservices in both development and production environments.

## Overview

All three core services now use **standardized database initialization**:
- **Landing Page Service**: `statuspage_landing`
- **SaaS Admin Service**: `saas_admin`
- **Tenant Admin Service**: `tenant_admin_db`

## Features

✅ **Automatic Database Creation**: Services create their databases if they don't exist
✅ **Declarative Migrations**: SQL migration files in `migrations/` directories
✅ **Production Ready**: Works with fresh RDS instances
✅ **Version Controlled**: All schema changes tracked in Git

## Quick Start

### Development (Local PostgreSQL)

```bash
# 1. Ensure PostgreSQL is running
brew services start postgresql@16  # macOS
# or
sudo systemctl start postgresql    # Linux

# 2. Run the initialization script
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices
./init-databases.sh

# 3. Start services
cd landing-page-service && ./landing-page-service &
cd saas-admin-service && ./saas-admin-service &
cd tenant-admin-service && ./tenant-admin-service &
```

### Production (RDS)

```bash
# 1. Set environment variables for your RDS instance
export DB_HOST=your-rds-endpoint.region.rds.amazonaws.com
export DB_PORT=5432
export DB_USER=admin
export DB_PASSWORD=your-secure-password
export DB_SSLMODE=require

# 2. Run the initialization script
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices
./init-databases.sh

# 3. Deploy services with connection details
# (via Kubernetes secrets, ECS task definitions, etc.)
```

## Service-Level Database Auto-Creation

Each service will automatically create its database on startup if it doesn't exist:

### Landing Page Service
- Database: `statuspage_landing`
- Auto-creation: ✅ [main.go:40-43](landing-page-service/cmd/main.go#L40-L43)
- Migrations: `landing-page-service/migrations/`

### SaaS Admin Service
- Database: `saas_admin`
- Auto-creation: ✅ (Already implemented)
- Migrations: `saas-admin-service/migrations/`

### Tenant Admin Service
- Database: `tenant_admin_db`
- Auto-creation: ❌ (Needs implementation - TODO)
- Migrations: `tenant-admin-service/migrations/`

## Migration Management

### Directory Structure

```
microservices/
├── landing-page-service/
│   ├── atlas.hcl
│   └── migrations/
│       └── 20251018203311_initial.sql
├── saas-admin-service/
│   ├── atlas.hcl
│   └── migrations/
│       └── 20251019000001_initial_schema.sql
├── tenant-admin-service/
│   ├── atlas.hcl
│   └── migrations/
│       └── 20251019000001_initial_schema.sql
└── init-databases.sh
```

### Adding New Migrations

#### Option 1: Manual SQL Migration (Recommended for now)

```bash
# 1. Create a new migration file
cd microservices/landing-page-service/migrations
touch 20251019120000_add_new_feature.sql

# 2. Write your SQL
cat > 20251019120000_add_new_feature.sql <<'EOF'
-- Add new column to pricing_plans table
ALTER TABLE pricing_plans ADD COLUMN discount_percentage NUMERIC DEFAULT 0;

-- Create index
CREATE INDEX idx_pricing_plans_discount ON pricing_plans(discount_percentage);
EOF

# 3. Apply migration
psql -U postgres -d statuspage_landing -f 20251019120000_add_new_feature.sql
```

#### Option 2: Atlas-Generated Migrations (Future)

```bash
# When Atlas GORM provider is fully configured:
cd microservices/landing-page-service
atlas migrate diff add_new_feature --env dev
atlas migrate apply --env dev
```

## Initialization Script Reference

### Usage

```bash
./init-databases.sh [OPTIONS]
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_SSLMODE` | `disable` | SSL mode (`disable`, `require`, `verify-full`) |

### What It Does

1. **Verifies Connection**: Tests PostgreSQL connectivity
2. **Creates Databases**: Creates all three databases if they don't exist
3. **Runs Migrations**: Applies all SQL files from `migrations/` directories in sorted order
4. **Shows Status**: Displays table counts and names for each database

### Example Output

```
=== Beakon Database Initialization ===
Host: localhost
Port: 5432
User: postgres

Step 1: Verifying connection
✓ Connection successful

Step 2: Creating databases
Checking database: statuspage_landing
✓ Database statuspage_landing created

Step 3: Running migrations
Running migrations for: statuspage_landing
Found 1 migration file(s)
  Applying: 20251018203311_initial.sql
  ✓ Applied successfully
✓ All migrations applied for statuspage_landing

=== Database Status ===
Database: statuspage_landing
  Tables: 20
  Table names:
    - articles
    - faqs
    - feature_sections
    - hero_sections
    - pricing_plans
    ...

=== Initialization Complete ===
All databases have been created and migrations applied successfully
```

## Testing with Fresh Databases

To test the complete initialization flow:

```bash
# 1. Drop all databases (CAUTION: This deletes all data!)
psql -U postgres -c "DROP DATABASE IF EXISTS statuspage_landing;"
psql -U postgres -c "DROP DATABASE IF EXISTS saas_admin;"
psql -U postgres -c "DROP DATABASE IF EXISTS tenant_admin_db;"

# 2. Run initialization script
./init-databases.sh

# 3. Verify databases exist
psql -U postgres -l | grep -E 'statuspage_landing|saas_admin|tenant_admin_db'

# 4. Start services and test
cd landing-page-service && ./landing-page-service &
curl http://localhost:8100/health
```

## Production Deployment Checklist

- [ ] Set `DB_SSLMODE=require` for RDS
- [ ] Use strong passwords (not `postgres`)
- [ ] Store credentials in secrets manager (AWS Secrets Manager, Kubernetes Secrets, etc.)
- [ ] Run `init-databases.sh` before first deployment
- [ ] Verify all migrations applied successfully
- [ ] Test service connectivity
- [ ] Set up automated backups (RDS snapshots)
- [ ] Configure connection pooling limits

## Troubleshooting

### Error: "failed to connect to postgres database"

**Cause**: PostgreSQL not running or wrong credentials
**Fix**:
```bash
# Check if PostgreSQL is running
pg_isready -h localhost -p 5432

# Verify credentials
psql -h localhost -U postgres -d postgres -c '\q'
```

### Error: "database already exists"

**Cause**: Database from previous run
**Fix**: This is not an error - script will skip creation and proceed to migrations

### Error: "relation already exists"

**Cause**: Migrations have been run before
**Fix**: Either:
1. Use idempotent migrations (`CREATE TABLE IF NOT EXISTS`)
2. Drop the database and re-run
3. Create incremental migrations only

### Services can't connect to database

**Cause**: Wrong connection string or credentials
**Fix**:
```bash
# Check environment variables in service
echo $DB_NAME
echo $DB_HOST
echo $DB_USER

# Test connection manually
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c '\q'
```

## Architecture Benefits

### Microservices Best Practices

✅ **Independent Databases**: Each service owns its data
✅ **Fault Isolation**: One database failure doesn't affect others
✅ **Independent Scaling**: Scale databases based on service load
✅ **Clear Boundaries**: No cross-database queries

### Production Ready

✅ **Version Control**: All schema changes in Git
✅ **Auditable**: Migration history shows what changed when
✅ **Rollback Support**: Can revert to previous schemas
✅ **Environment Parity**: Same migrations run in dev and prod

## Future Enhancements

- [ ] Implement database auto-creation for Tenant Admin Service
- [ ] Add rollback/down migrations
- [ ] Integrate with CI/CD for automatic migration testing
- [ ] Add migration linting (sqlfluff)
- [ ] Create database migration monitoring/alerting
- [ ] Add schema drift detection

## Support

For issues or questions:
- Check service logs in `/tmp/<service>.log`
- Review migration files in `migrations/` directories
- Verify database state with `psql -U postgres -d <database> -c '\dt'`
