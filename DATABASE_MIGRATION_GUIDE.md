# Database Migration Guide

**Last Updated**: 2025-10-26
**Status**: ✅ Production-Ready Holistic Approach

## Overview

This guide documents the **holistic database initialization and migration strategy** for the Beakon Status Page Platform. All services follow a standardized, Atlas-based approach that ensures safety, consistency, and auditability.

## Table of Contents

1. [Principles](#principles)
2. [Architecture](#architecture)
3. [Safety Features](#safety-features)
4. [Atlas Configuration](#atlas-configuration)
5. [Migration Workflow](#migration-workflow)
6. [Seed Data Management](#seed-data-management)
7. [Common Tasks](#common-tasks)
8. [Troubleshooting](#troubleshooting)

---

## Principles

### Core Principles

1. **Single Source of Truth**: GORM models in `internal/models` are the source of truth for schema
2. **No AutoMigrate**: Services NEVER run `db.AutoMigrate()` - schema is managed externally
3. **Migration Tracking**: Atlas tracks all applied migrations in `atlas_schema_revisions` table
4. **Safety First**: Never drop databases, always validate before applying migrations
5. **Separation of Concerns**: Schema migrations separate from seed data

### What We DON'T Do

❌ **NO AutoMigrate()** - Removed from all services
❌ **NO manual SQL dumps** - No large pg_dump files in repository
❌ **NO untracked migrations** - Every migration is versioned and tracked
❌ **NO data loss risks** - Safety checks prevent accidental drops

### What We DO

✅ **Use Atlas CLI** - Industry-standard migration tool
✅ **Generate from Go models** - Schema derived from code
✅ **Track migration state** - Atlas knows what's been applied
✅ **Holistic init scripts** - One script does everything correctly
✅ **Separate seed data** - Development data kept separate from schema

---

## Architecture

### Directory Structure (Per Service)

```
microservices/<service-name>/
├── atlas.hcl              # Atlas configuration
├── init-db.sh             # Holistic initialization script
├── migrations/            # Generated Atlas migrations
│   ├── 20251026_initial_schema.sql
│   ├── 20251027_add_indexes.sql
│   └── atlas.sum          # Migration checksums
├── seeds/                 # Seed data (optional)
│   ├── 01_admin_users.sql
│   ├── 02_default_settings.sql
│   └── .applied           # Marker file (tracks if seeds applied)
├── internal/
│   └── models/            # GORM models (source of truth)
│       ├── user.go
│       ├── tenant.go
│       └── ...
├── cmd/
│   └── main.go            # NO db.AutoMigrate() calls here
└── init-db.log            # Log file (generated on init)
```

### Atlas Workflow

```
┌─────────────────┐
│  GORM Models    │  (internal/models/*.go)
│  Source of      │
│  Truth           │
└────────┬────────┘
         │
         │ atlas migrate diff
         ▼
┌─────────────────┐
│  Atlas Inspects │
│  Go Models       │
└────────┬────────┘
         │
         │ generates SQL
         ▼
┌─────────────────┐
│  Migration SQL  │  (migrations/*.sql)
│  Versioned &    │
│  Tracked         │
└────────┬────────┘
         │
         │ atlas migrate apply
         ▼
┌─────────────────┐
│  PostgreSQL     │
│  Database        │
│  + Tracking      │
│    Table         │
└─────────────────┘
```

---

## Safety Features

### Database Protection

1. **Never Drop Existing Databases**
   ```bash
   # init-db.sh checks if database exists
   # If exists: SKIPS creation (safe)
   # If not exists: Creates new database
   ```

2. **Migration Tracking**
   ```sql
   -- Atlas creates this table automatically
   atlas_schema_revisions (
       version VARCHAR PRIMARY KEY,
       description TEXT,
       applied_at TIMESTAMP
   )
   ```
   - Prevents re-applying migrations
   - Tracks which migrations have been applied
   - Ensures linear migration history

3. **Destructive Change Protection**
   ```hcl
   # In atlas.hcl (production environment)
   diff {
     skip {
       drop_schema = true   # Never drop schemas
       drop_table  = true   # Never drop tables
       drop_column = true   # Never drop columns
     }
   }
   ```

4. **Dry Run Mode**
   ```bash
   DRY_RUN=true ./init-db.sh
   # Shows what WOULD happen, doesn't execute
   ```

5. **Seed Data Protection**
   ```bash
   # Seed data applied only once
   # Re-application requires explicit flag
   FORCE_SEED=true ./init-db.sh
   ```

### Logging and Audit Trail

Every init-db.sh run creates a log file:
```bash
microservices/user-service/init-db.log
```

Contains:
- Timestamp of initialization
- PostgreSQL connection details
- Migration steps
- Seed data application
- Final database status

---

## Atlas Configuration

### Standard atlas.hcl Structure

Every service has an `atlas.hcl` file with three environments:

#### 1. Development Environment (localhost)

```hcl
env "dev" {
  src = data.external_schema.gorm.url
  url = "postgres://postgres:postgres@localhost:5432/<db_name>?sslmode=disable"
  dev = "docker://postgres/14/dev?search_path=public"

  migration {
    dir = "file://migrations"
    revisions_schema = "atlas_schema_revisions"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }

  # Safety: Require manual approval for drops
  diff {
    skip {
      drop_schema = true
      drop_table  = true
    }
  }
}
```

**Usage**: Local development on macOS/Linux with PostgreSQL running on localhost

#### 2. Docker Environment (containerized)

```hcl
env "docker" {
  src = data.external_schema.gorm.url
  url = "postgres://postgres:postgres@statuspage-postgres:5432/<db_name>?sslmode=disable"
  dev = "docker://postgres/14/dev?search_path=public"

  migration {
    dir = "file://migrations"
    revisions_schema = "atlas_schema_revisions"
  }
}
```

**Usage**: Running inside Docker containers (uses service name `statuspage-postgres`)

#### 3. Production Environment (RDS/managed)

```hcl
env "prod" {
  src = data.external_schema.gorm.url
  url = getenv("ATLAS_URL")  # MUST be set via environment variable

  migration {
    dir = "file://migrations"
    revisions_schema = "atlas_schema_revisions"
  }

  # CRITICAL: Prevent destructive changes
  diff {
    skip {
      drop_schema = true
      drop_table  = true
      drop_column = true
    }
  }

  migrate {
    baseline = getenv("ATLAS_BASELINE")
  }
}
```

**Usage**: Production deployments (AWS RDS, Google Cloud SQL, etc.)

---

## Migration Workflow

### Generating Atlas Config for a New Service

```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon

# Generate atlas.hcl
./scripts/generate-atlas-config.sh user-service statuspage_user

# Generate init-db.sh
./scripts/generate-init-db.sh user-service statuspage_user
```

This creates:
- `microservices/user-service/atlas.hcl`
- `microservices/user-service/init-db.sh`

### Creating a New Migration

**Prerequisite**: Define GORM models in `internal/models/`

```bash
cd microservices/user-service

# Generate migration from current Go models
atlas migrate diff add_user_tables --env dev

# This creates: migrations/20251026_add_user_tables.sql
```

Atlas compares:
1. Current database schema
2. Desired schema (from GORM models)
3. Generates SQL to transform (1) into (2)

### Applying Migrations

**Option 1: Use init-db.sh (Recommended)**

```bash
cd microservices/user-service
./init-db.sh
```

This automatically:
1. Creates database if needed
2. Applies all pending migrations
3. Tracks migration state
4. Applies seed data (if exists)
5. Shows final status

**Option 2: Use Atlas CLI Directly**

```bash
cd microservices/user-service

# Check migration status
atlas migrate status --env dev

# Apply pending migrations
atlas migrate apply --env dev

# Apply specific number of migrations
atlas migrate apply --env dev --amount 1
```

### Viewing Migration Status

```bash
cd microservices/user-service

atlas migrate status --env dev
```

Output:
```
Migration Status:
  Pending: 2
  Applied: 5

Applied Migrations:
  ✓ 20251020_initial_schema.sql       (2025-10-20 10:30:00)
  ✓ 20251021_add_indexes.sql          (2025-10-21 14:15:30)
  ✓ 20251022_add_tenant_id.sql        (2025-10-22 09:45:12)

Pending Migrations:
  ○ 20251026_add_rbac_tables.sql
  ○ 20251027_add_audit_logs.sql
```

### Rolling Back Migrations

```bash
cd microservices/user-service

# Rollback last migration
atlas migrate down --env dev

# Rollback specific version
atlas migrate down --env dev --to-version 20251020
```

---

## Seed Data Management

### Why Separate Seed Data?

- **Schema migrations**: Structural changes (CREATE TABLE, ADD COLUMN, etc.)
- **Seed data**: Initial data (admin users, default settings, etc.)

Mixing them creates problems:
- Can't re-apply migrations without duplicating data
- Hard to manage different data for dev/staging/prod
- Security risks (production data in migrations)

### Creating Seed Data

```bash
cd microservices/user-service

# Create seeds directory
mkdir -p seeds

# Create seed files (numbered for order)
cat > seeds/01_admin_users.sql << 'EOF'
-- Default admin user for development
INSERT INTO users (email, username, password, role, status, created_at, updated_at)
VALUES (
  'admin@beaconstatus.com',
  'admin',
  '$2a$10$rN5Z9smD8.L7LMTFq.vZqO1WxQBm9JJ0V6nOp9ZL.XjKXNbW9F3qG',  -- bcrypt: admin123
  'super_admin',
  'active',
  NOW(),
  NOW()
)
ON CONFLICT (email) DO NOTHING;
EOF

cat > seeds/02_default_settings.sql << 'EOF'
-- Default system settings
INSERT INTO system_settings (key, value, created_at)
VALUES
  ('maintenance_mode', 'false', NOW()),
  ('max_upload_size_mb', '10', NOW()),
  ('session_timeout_hours', '24', NOW())
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;
EOF
```

### Applying Seed Data

```bash
cd microservices/user-service

# First time: seeds are applied automatically
./init-db.sh

# Force re-application (CAUTION: may create duplicates)
FORCE_SEED=true ./init-db.sh
```

**Safety**: Seed data is applied only once. A marker file `seeds/.applied` prevents re-application.

### Best Practices for Seed Data

1. **Use ON CONFLICT** to handle re-runs:
   ```sql
   INSERT INTO users (email, ...) VALUES (...)
   ON CONFLICT (email) DO NOTHING;
   ```

2. **Use idempotent operations**:
   ```sql
   -- Good: Upsert style
   ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;

   -- Bad: Will fail on re-run
   INSERT INTO settings VALUES (...);
   ```

3. **Separate dev and prod seeds**:
   ```
   seeds/
   ├── dev/
   │   ├── 01_test_users.sql
   │   └── 02_sample_data.sql
   └── prod/
       └── 01_admin_only.sql
   ```

---

## Common Tasks

### Initialize a Brand New Service Database

```bash
cd microservices/user-service

# Run holistic init script
./init-db.sh
```

Output:
```
========================================
user-service - Database Initialization
========================================
Database: statuspage_user
Host: localhost:5432
Atlas Environment: dev
Timestamp: 2025-10-26 10:30:15

[INFO] Checking PostgreSQL connectivity...
[SUCCESS] PostgreSQL connection successful
[INFO] Checking Atlas CLI installation...
[SUCCESS] Atlas CLI installed: atlas version v0.37.1

[INFO] Checking if database exists...
[INFO] Creating database 'statuspage_user'...
[SUCCESS] Database 'statuspage_user' created

[INFO] Applying schema migrations using Atlas...
[INFO] Running: atlas migrate apply --env dev
[SUCCESS] Migrations applied successfully

[INFO] Found 2 seed data file(s)
[INFO] Applying seed data...
  [INFO] Applying: 01_admin_users.sql
  [SUCCESS] ✓ 01_admin_users.sql applied
  [INFO] Applying: 02_default_settings.sql
  [SUCCESS] ✓ 02_default_settings.sql applied
[SUCCESS] Seed data applied successfully

========================================
Database Status
========================================
Database: statuspage_user
Tables: 12

Table List:
  - atlas_schema_revisions
  - users
  - user_profiles
  - user_sessions
  - user_activities
  - roles
  - permissions
  - role_permissions
  - teams
  - team_members
  - audit_logs
  - system_settings

Applied Migrations: 3
(tracked in atlas_schema_revisions table)

[SUCCESS] user-service database initialization complete
Log file: ./init-db.log
```

### Add a New GORM Model and Generate Migration

**Step 1**: Create the model

```go
// internal/models/subscription.go
package models

import (
    "time"
    "github.com/google/uuid"
)

type Subscription struct {
    ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    TenantID  uuid.UUID  `gorm:"type:uuid;not null;index"`
    PlanID    uuid.UUID  `gorm:"type:uuid;not null"`
    Status    string     `gorm:"type:varchar(50);not null;default:'active'"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**Step 2**: Generate migration

```bash
cd microservices/user-service

atlas migrate diff add_subscriptions --env dev
```

**Step 3**: Review generated migration

```bash
cat migrations/20251026_add_subscriptions.sql
```

**Step 4**: Apply migration

```bash
atlas migrate apply --env dev
```

### Update an Existing Model

**Step 1**: Modify the model

```go
// internal/models/user.go
type User struct {
    // ... existing fields ...
    PhoneNumber string `gorm:"type:varchar(20);index"`  // NEW FIELD
}
```

**Step 2**: Generate migration

```bash
atlas migrate diff add_phone_number --env dev
```

Atlas generates:
```sql
-- add_phone_number migration
ALTER TABLE users ADD COLUMN phone_number VARCHAR(20);
CREATE INDEX idx_users_phone_number ON users(phone_number);
```

**Step 3**: Apply

```bash
atlas migrate apply --env dev
```

### Remove AutoMigrate from Service Code

**Before** (DON'T DO THIS):

```go
// cmd/main.go
func main() {
    db := setupDatabase()

    // ❌ BAD: AutoMigrate runs on every service start
    db.AutoMigrate(
        &models.User{},
        &models.Tenant{},
        &models.Role{},
    )

    startServer(db)
}
```

**After** (CORRECT):

```go
// cmd/main.go
func main() {
    db := setupDatabase()

    // ✅ GOOD: No AutoMigrate, schema managed by Atlas
    // Migrations are applied externally via init-db.sh

    startServer(db)
}
```

### Regenerate All Configs and Scripts

```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon

# For all services
./scripts/generate-atlas-config.sh user-service statuspage_user
./scripts/generate-init-db.sh user-service statuspage_user

./scripts/generate-atlas-config.sh tenant-admin-service tenant_admin_db
./scripts/generate-init-db.sh tenant-admin-service tenant_admin_db

# ... repeat for all services
```

Or use the batch script:

```bash
./scripts/setup-all-services.sh
```

---

## Troubleshooting

### "atlas: command not found"

**Problem**: Atlas CLI not installed

**Solution**:
```bash
# macOS
brew install ariga/tap/atlas

# Linux
curl -sSf https://atlasgo.sh | sh

# Verify
atlas version
```

### "cannot connect to PostgreSQL"

**Problem**: PostgreSQL not running or wrong connection params

**Solution**:
```bash
# Check if PostgreSQL is running
pg_isready -h localhost -p 5432

# Start PostgreSQL (macOS)
brew services start postgresql@16

# Check connection manually
psql -U postgres -h localhost -d postgres -c '\q'
```

### "migration already applied"

**Problem**: Trying to re-run a migration that's already in `atlas_schema_revisions`

**Solution**: This is SAFE - Atlas prevents re-application automatically

```bash
atlas migrate status --env dev
# Shows which migrations are applied
```

### "seed data creating duplicates"

**Problem**: Seed SQL doesn't handle re-runs

**Solution**: Use `ON CONFLICT` clauses:

```sql
-- Instead of:
INSERT INTO users VALUES (...);

-- Use:
INSERT INTO users VALUES (...)
ON CONFLICT (email) DO NOTHING;
```

### "atlas migrate diff generates empty migration"

**Problem**: Database schema already matches GORM models

**Solution**: No action needed - this is expected when schema is up-to-date

```bash
# Check status
atlas migrate status --env dev
# Should show "No pending migrations"
```

### "production migration failed"

**Problem**: Trying to drop columns/tables in production (blocked by safety config)

**Solution**: This is intentional! Destructive changes are blocked in prod.

Options:
1. **Recommended**: Deprecate columns instead of dropping
2. **If absolutely necessary**: Temporarily modify `atlas.hcl` (with extreme caution)

---

## Summary: The Holistic Approach

### What Makes It Holistic?

1. **Standardized Tooling**: Every service uses Atlas
2. **Single Workflow**: Same commands work for all services
3. **Complete Automation**: `init-db.sh` handles everything
4. **Safety by Default**: Multiple layers of protection
5. **Auditability**: Full logs and tracking
6. **Separation of Concerns**: Schema vs. seed data
7. **Environment-Aware**: Dev, Docker, and Prod configs

### The One Command That Does Everything

```bash
cd microservices/<service-name>
./init-db.sh
```

This single command:
1. ✅ Validates prerequisites (PostgreSQL, Atlas)
2. ✅ Creates database (if needed, never drops existing)
3. ✅ Applies all pending migrations (tracked by Atlas)
4. ✅ Applies seed data (only once, unless forced)
5. ✅ Shows final status
6. ✅ Logs everything for audit

### Security and Safety

- ❌ **Never** stores large SQL dumps in repository
- ❌ **Never** uses AutoMigrate() in service code
- ❌ **Never** drops databases or tables accidentally
- ❌ **Never** re-applies migrations (Atlas tracking)
- ✅ **Always** validates before executing
- ✅ **Always** logs operations
- ✅ **Always** uses versioned migrations

---

## Related Documentation

- [DATABASE_ARCHITECTURE.md](DATABASE_ARCHITECTURE.md) - Database schemas and design
- [CLAUDE.md](CLAUDE.md) - Developer workflow guide
- [SERVICE_CATALOG.md](SERVICE_CATALOG.md) - Service reference
- [Atlas Documentation](https://atlasgo.io/docs) - Official Atlas docs

---

**Last Updated**: 2025-10-26
**Maintained By**: Platform Engineering Team
