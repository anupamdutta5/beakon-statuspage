# Database Migration Scripts

**Purpose**: Holistic database initialization and migration management using Atlas

## Overview

This directory contains scripts to generate standardized Atlas configurations and database initialization scripts for all microservices.

## Files

- **generate-atlas-config.sh** - Generates `atlas.hcl` for a service
- **generate-init-db.sh** - Generates `init-db.sh` for a service
- **init-db-template.sh** - Template used by generate-init-db.sh

## Quick Start

### Generate Config for a New Service

```bash
./generate-atlas-config.sh <service-name> <database-name>
./generate-init-db.sh <service-name> <database-name>
```

**Example**:
```bash
./generate-atlas-config.sh user-service statuspage_user
./generate-init-db.sh user-service statuspage_user
```

This creates:
- `microservices/user-service/atlas.hcl`
- `microservices/user-service/init-db.sh`

### Initialize Database

```bash
cd microservices/user-service
./init-db.sh
```

## Service/Database Mapping

| Service | Database Name |
|---------|---------------|
| user-service | statuspage_user |
| tenant-admin-service | tenant_admin_db |
| saas-admin-service | saas_admin |
| component-service | statuspage_component |
| notification-service | statuspage_notification |
| incident-service | statuspage_incident |
| payment-service | statuspage_payment |
| analytics-service | statuspage_analytics |
| monitoring-service | statuspage_monitoring |
| status-ui-service | statuspage_status |
| event-store-service | statuspage_events |
| branding-service | statuspage_branding |
| landing-page-service | statuspage_landing |
| audit-consumer | statuspage_audit |
| analytics-consumer | statuspage_analytics |
| notification-consumer | statuspage_notification |
| billing-consumer | statuspage_billing |

## Documentation

See [DATABASE_MIGRATION_GUIDE.md](../DATABASE_MIGRATION_GUIDE.md) for complete documentation.

## Safety Features

All generated scripts include:
- ✅ Never drop existing databases
- ✅ Migration tracking (prevents re-application)
- ✅ Dry-run mode support
- ✅ Comprehensive logging
- ✅ Separate seed data management

## Related

- [DATABASE_MIGRATION_GUIDE.md](../DATABASE_MIGRATION_GUIDE.md) - Complete migration guide
- [DATABASE_ARCHITECTURE.md](../DATABASE_ARCHITECTURE.md) - Database schemas
- [CLAUDE.md](../CLAUDE.md) - Developer guide
