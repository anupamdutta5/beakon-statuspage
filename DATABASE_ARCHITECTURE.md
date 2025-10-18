# Beakon - Database Architecture

**Last Updated**: 2025-10-14
**Database System**: PostgreSQL 13+
**Total Databases**: 14 active (+ 1 deprecated)
**Pattern**: Database-per-Service (with one documented exception)

---

## Executive Summary

The Beakon platform follows microservices best practices with a **database-per-service** architecture. Each microservice owns and manages its own database schema, ensuring loose coupling and independent scalability. The only exception is `tenant_admin_db`, which is intentionally shared between two closely related administrative services.

### Key Principles

1. **Data Isolation**: Each service owns its data
2. **Schema Independence**: Services can evolve schemas independently
3. **No Direct Database Access**: Services communicate via APIs, not direct DB queries
4. **Multi-Tenant Data Segregation**: All tables include tenant_id for data isolation

---

## Database Inventory

### Active Databases

| Database Name | Service | Tables | Size Estimate | Backup Priority |
|---------------|---------|--------|---------------|-----------------|
| tenant_admin_db | saas-admin-service, tenant-admin-service | 25+ | Large | **CRITICAL** |
| statuspage_user | user-service | 5 | Medium | **CRITICAL** |
| statuspage_component | component-service | 6 | Medium | HIGH |
| statuspage_notification | notification-service | 7 | Large | HIGH |
| statuspage_incident | incident-service | 6 | Large | **CRITICAL** |
| statuspage_payment | payment-service | 8 | Medium | **CRITICAL** |
| statuspage_analytics | analytics-service | 7 | Very Large | MEDIUM |
| statuspage_monitoring | monitoring-service | 6 | Large | HIGH |
| statuspage_event_store | event-store-service | 3 | Very Large | HIGH |
| statuspage_branding | branding-service | 4 | Small | LOW |
| statuspage_landing | landing-page-service | 5 | Small | LOW |
| statuspage_analytics_consumer | analytics-consumer | 3 | Large | MEDIUM |
| statuspage_audit_consumer | audit-consumer | 3 | Very Large | **CRITICAL** |
| statuspage_billing_consumer | billing-consumer | 4 | Medium | HIGH |

### Deprecated Databases

| Database Name | Service | Status | Action Required |
|---------------|---------|--------|-----------------|
| statuspage_database | database-service | ⚠️ **DEPRECATED** | Archive and remove |

---

## Database Details

### 1. tenant_admin_db

**Services**: saas-admin-service, tenant-admin-service
**Status**: **SHARED** (intentionally)
**Size**: Large
**Backup**: **CRITICAL** - Daily full + continuous WAL archiving

#### Schema Ownership

##### SaaS Admin Service Tables:
```sql
-- Platform Administration
saas_admins               -- Platform admin users
subscription_plans        -- Subscription plan definitions
features                  -- Feature catalog
pricing                   -- Pricing rules and tiers
platform_settings         -- System-wide configuration
```

##### Tenant Admin Service Tables:
```sql
-- Tenant Management
tenants                   -- Tenant records (with max_users)
tenant_settings           -- Tenant-specific settings
domains                   -- Custom domain management

-- User Management
users                     -- Tenant users
sessions                  -- Session tracking

-- RBAC System
roles                     -- Role definitions
permissions               -- Permission catalog
user_roles                -- User-role assignments
role_permissions          -- Role-permission mappings

-- Teams
teams                     -- Team definitions
team_members              -- Team membership

-- Status Pages
status_pages              -- Status page configurations
status_page_config        -- Status page settings

-- Audit & Tracking
audit_logs                -- Comprehensive audit trail
tenant_activities         -- Activity tracking
user_activities           -- User action log

-- Advanced Features
feature_flags             -- Tenant-level feature flags
tenant_usage              -- Usage metrics
tenant_billing            -- Billing records
tenant_notifications      -- Notification preferences
tenant_backups            -- Backup metadata
```

#### Why Sharing Is Acceptable

1. **Same Domain**: Both services are part of the administrative control plane
2. **Clear Boundaries**: No table ownership conflicts
3. **Transactional Integrity**: Simplified for cross-cutting admin operations
4. **Tight Coupling by Design**: Admin services are intentionally coupled
5. **Simplified Deployment**: Reduces operational complexity

#### Migration Strategy

If services need to be split in the future:

```sql
-- Create separate databases
CREATE DATABASE saas_admin_db;
CREATE DATABASE tenant_management_db;

-- Migrate SaaS Admin tables
-- Use pg_dump with --table flag for selective export

-- Migrate Tenant Admin tables
-- Update foreign key relationships
-- Implement compensating transactions for cross-database operations
```

**Estimated Effort**: 2-3 days
**Risk**: Medium (requires service downtime or careful migration)
**Recommendation**: Keep shared unless scaling demands separation

---

### 2. statuspage_user

**Service**: user-service
**Size**: Medium
**Backup**: **CRITICAL** - Daily full + continuous WAL

#### Schema

```sql
-- User Authentication
users                     -- User accounts with credentials
  id BIGSERIAL PRIMARY KEY
  tenant_id UUID NOT NULL
  email TEXT NOT NULL UNIQUE
  password_hash TEXT NOT NULL
  first_name TEXT
  last_name TEXT
  is_active BOOLEAN DEFAULT true
  last_login_at TIMESTAMP
  created_at TIMESTAMP
  updated_at TIMESTAMP
  deleted_at TIMESTAMP

-- Token Management
refresh_tokens            -- JWT refresh token storage
  id BIGSERIAL PRIMARY KEY
  user_id BIGINT REFERENCES users(id)
  token TEXT NOT NULL UNIQUE
  expires_at TIMESTAMP NOT NULL
  created_at TIMESTAMP

password_reset_tokens     -- Password reset tokens
  id BIGSERIAL PRIMARY KEY
  user_id BIGINT REFERENCES users(id)
  token TEXT NOT NULL UNIQUE
  expires_at TIMESTAMP NOT NULL
  used BOOLEAN DEFAULT false
  created_at TIMESTAMP

-- User Profiles
user_profiles             -- Extended user information
  id BIGSERIAL PRIMARY KEY
  user_id BIGINT REFERENCES users(id) UNIQUE
  avatar_url TEXT
  bio TEXT
  timezone TEXT
  locale TEXT
  preferences JSONB
  created_at TIMESTAMP
  updated_at TIMESTAMP

-- OAuth Integration
oauth_providers           -- OAuth provider connections
  id BIGSERIAL PRIMARY KEY
  user_id BIGINT REFERENCES users(id)
  provider TEXT NOT NULL  -- 'google', 'github', etc.
  provider_user_id TEXT NOT NULL
  access_token TEXT
  refresh_token TEXT
  expires_at TIMESTAMP
  created_at TIMESTAMP

-- Indexes
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
```

#### Data Volume Estimates

- **Users**: ~10,000-100,000 per tenant (varies widely)
- **Refresh Tokens**: ~10x user count (multiple devices)
- **Growth Rate**: ~5% monthly

#### Retention Policy

- **Users**: Indefinite (soft delete with deleted_at)
- **Refresh Tokens**: 30 days after expiration
- **Password Reset Tokens**: 7 days after use/expiration

---

### 3. statuspage_component

**Service**: component-service
**Size**: Medium
**Backup**: HIGH - Daily full

#### Schema

```sql
-- Component Definitions
components                -- System components
  id BIGSERIAL PRIMARY KEY
  tenant_id UUID NOT NULL
  name TEXT NOT NULL
  description TEXT
  status TEXT NOT NULL  -- 'operational', 'degraded', 'down', 'maintenance'
  component_group_id BIGINT
  display_order INT DEFAULT 0
  is_visible BOOLEAN DEFAULT true
  created_at TIMESTAMP
  updated_at TIMESTAMP
  deleted_at TIMESTAMP

component_groups          -- Component grouping
  id BIGSERIAL PRIMARY KEY
  tenant_id UUID NOT NULL
  name TEXT NOT NULL
  description TEXT
  display_order INT DEFAULT 0
  created_at TIMESTAMP
  updated_at TIMESTAMP

component_status_history  -- Status change log
  id BIGSERIAL PRIMARY KEY
  component_id BIGINT REFERENCES components(id)
  old_status TEXT
  new_status TEXT NOT NULL
  reason TEXT
  changed_by BIGINT  -- user_id
  created_at TIMESTAMP

component_metrics         -- Performance metrics
  id BIGSERIAL PRIMARY KEY
  component_id BIGINT REFERENCES components(id)
  metric_name TEXT NOT NULL
  metric_value NUMERIC
  unit TEXT
  recorded_at TIMESTAMP
  created_at TIMESTAMP

component_dependencies    -- Component dependencies
  id BIGSERIAL PRIMARY KEY
  component_id BIGINT REFERENCES components(id)
  depends_on_component_id BIGINT REFERENCES components(id)
  dependency_type TEXT  -- 'hard', 'soft'
  created_at TIMESTAMP

component_tags            -- Component tagging
  id BIGSERIAL PRIMARY KEY
  component_id BIGINT REFERENCES components(id)
  tag TEXT NOT NULL
  created_at TIMESTAMP

-- Indexes
CREATE INDEX idx_components_tenant_id ON components(tenant_id);
CREATE INDEX idx_components_status ON components(status);
CREATE INDEX idx_component_history_component_id ON component_status_history(component_id);
CREATE INDEX idx_component_history_created_at ON component_status_history(created_at DESC);
```

#### Partitioning Strategy

For high-volume tenants, consider partitioning `component_status_history` by tenant_id or time range:

```sql
-- Example: Partition by time (monthly)
CREATE TABLE component_status_history_2025_01 PARTITION OF component_status_history
  FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
```

---

### 4. statuspage_notification

**Service**: notification-service
**Size**: Large
**Backup**: HIGH - Daily full

#### Schema

```sql
notifications             -- Notification log
  id BIGSERIAL PRIMARY KEY
  tenant_id UUID NOT NULL
  type TEXT NOT NULL  -- 'email', 'sms', 'push', 'webhook'
  recipient TEXT NOT NULL
  subject TEXT
  body TEXT
  status TEXT  -- 'pending', 'sent', 'failed'
  sent_at TIMESTAMP
  failed_reason TEXT
  retry_count INT DEFAULT 0
  created_at TIMESTAMP

subscribers               -- Subscriber management
  id BIGSERIAL PRIMARY KEY
  tenant_id UUID NOT NULL
  email TEXT
  phone TEXT
  push_token TEXT
  subscribed_at TIMESTAMP
  unsubscribed_at TIMESTAMP
  preferences JSONB
  created_at TIMESTAMP

notification_templates    -- Message templates
  id BIGSERIAL PRIMARY KEY
  tenant_id UUID NOT NULL
  name TEXT NOT NULL
  type TEXT NOT NULL
  subject_template TEXT
  body_template TEXT
  variables JSONB
  created_at TIMESTAMP
  updated_at TIMESTAMP

webhooks                  -- Webhook configurations
  id BIGSERIAL PRIMARY KEY
  tenant_id UUID NOT NULL
  url TEXT NOT NULL
  secret TEXT
  events TEXT[]  -- Array of event types
  is_active BOOLEAN DEFAULT true
  last_triggered_at TIMESTAMP
  created_at TIMESTAMP
  updated_at TIMESTAMP

webhook_deliveries        -- Webhook delivery log
  id BIGSERIAL PRIMARY KEY
  webhook_id BIGINT REFERENCES webhooks(id)
  event_type TEXT
  payload JSONB
  response_code INT
  response_body TEXT
  delivered_at TIMESTAMP
  created_at TIMESTAMP

notification_preferences  -- User notification preferences
  id BIGSERIAL PRIMARY KEY
  user_id BIGINT NOT NULL
  tenant_id UUID NOT NULL
  email_enabled BOOLEAN DEFAULT true
  sms_enabled BOOLEAN DEFAULT false
  push_enabled BOOLEAN DEFAULT false
  quiet_hours_start TIME
  quiet_hours_end TIME
  created_at TIMESTAMP
  updated_at TIMESTAMP

notification_channels     -- Channel configurations
  id BIGSERIAL PRIMARY KEY
  tenant_id UUID NOT NULL
  channel_type TEXT NOT NULL
  config JSONB
  is_active BOOLEAN DEFAULT true
  created_at TIMESTAMP
  updated_at TIMESTAMP
```

#### Data Retention

- **Notifications**: 90 days (configurable per tenant)
- **Webhook Deliveries**: 30 days
- **Subscriber History**: Indefinite

#### Cleanup Job

```sql
-- Delete old notification records
DELETE FROM notifications
WHERE created_at < NOW() - INTERVAL '90 days'
  AND status IN ('sent', 'failed');

-- Delete old webhook delivery logs
DELETE FROM webhook_deliveries
WHERE created_at < NOW() - INTERVAL '30 days';
```

**Run**: Daily at 2 AM UTC

---

### 5. statuspage_incident

**Service**: incident-service
**Size**: Large
**Backup**: **CRITICAL** - Daily full + hourly incrementals

#### Schema

```sql
incidents                 -- Incident records
  id BIGSERIAL PRIMARY KEY
  tenant_id UUID NOT NULL
  title TEXT NOT NULL
  description TEXT
  status TEXT NOT NULL  -- 'investigating', 'identified', 'monitoring', 'resolved'
  severity TEXT  -- 'minor', 'major', 'critical'
  started_at TIMESTAMP
  resolved_at TIMESTAMP
  created_by BIGINT  -- user_id
  created_at TIMESTAMP
  updated_at TIMESTAMP

incident_updates          -- Status updates
  id BIGSERIAL PRIMARY KEY
  incident_id BIGINT REFERENCES incidents(id)
  status TEXT NOT NULL
  message TEXT NOT NULL
  created_by BIGINT
  created_at TIMESTAMP

incident_components       -- Component associations
  id BIGSERIAL PRIMARY KEY
  incident_id BIGINT REFERENCES incidents(id)
  component_id BIGINT NOT NULL
  impact TEXT  -- 'none', 'minor', 'major', 'critical'
  created_at TIMESTAMP

maintenance_windows       -- Scheduled maintenance
  id BIGSERIAL PRIMARY KEY
  tenant_id UUID NOT NULL
  title TEXT NOT NULL
  description TEXT
  scheduled_start TIMESTAMP NOT NULL
  scheduled_end TIMESTAMP NOT NULL
  status TEXT  -- 'scheduled', 'in_progress', 'completed', 'cancelled'
  notify_subscribers BOOLEAN DEFAULT true
  created_at TIMESTAMP
  updated_at TIMESTAMP

incident_templates        -- Reusable templates
  id BIGSERIAL PRIMARY KEY
  tenant_id UUID NOT NULL
  name TEXT NOT NULL
  title_template TEXT
  description_template TEXT
  severity TEXT
  affected_components BIGINT[]
  created_at TIMESTAMP
  updated_at TIMESTAMP

incident_postmortems      -- Post-incident analysis
  id BIGSERIAL PRIMARY KEY
  incident_id BIGINT REFERENCES incidents(id) UNIQUE
  summary TEXT
  root_cause TEXT
  action_items JSONB
  lessons_learned TEXT
  created_by BIGINT
  created_at TIMESTAMP
  updated_at TIMESTAMP
```

#### Performance Optimization

- **Index**: `incidents(tenant_id, status, created_at DESC)` for dashboard queries
- **Materialized View**: For incident statistics

```sql
CREATE MATERIALIZED VIEW incident_stats AS
SELECT
  tenant_id,
  COUNT(*) as total_incidents,
  COUNT(*) FILTER (WHERE status = 'resolved') as resolved_incidents,
  AVG(EXTRACT(EPOCH FROM (resolved_at - started_at))) as avg_resolution_time,
  DATE_TRUNC('month', created_at) as month
FROM incidents
WHERE deleted_at IS NULL
GROUP BY tenant_id, DATE_TRUNC('month', created_at);

-- Refresh daily
REFRESH MATERIALIZED VIEW incident_stats;
```

---

### 6. statuspage_payment

**Service**: payment-service
**Size**: Medium
**Backup**: **CRITICAL** - Hourly full (financial data)

#### Schema

```sql
payments                  -- Payment transactions
  id BIGSERIAL PRIMARY KEY
  tenant_id UUID NOT NULL
  amount NUMERIC(10,2) NOT NULL
  currency TEXT DEFAULT 'USD'
  status TEXT  -- 'pending', 'completed', 'failed', 'refunded'
  payment_method TEXT  -- 'card', 'ach', 'paypal'
  transaction_id TEXT UNIQUE
  provider TEXT  -- 'stripe', 'paypal'
  provider_transaction_id TEXT
  metadata JSONB
  processed_at TIMESTAMP
  created_at TIMESTAMP

subscriptions             -- Subscription records
  id BIGSERIAL PRIMARY KEY
  tenant_id UUID NOT NULL
  plan_id BIGINT NOT NULL
  status TEXT  -- 'active', 'cancelled', 'past_due', 'unpaid'
  current_period_start TIMESTAMP
  current_period_end TIMESTAMP
  cancel_at_period_end BOOLEAN DEFAULT false
  cancelled_at TIMESTAMP
  created_at TIMESTAMP
  updated_at TIMESTAMP

invoices                  -- Invoice tracking
  id BIGSERIAL PRIMARY KEY
  tenant_id UUID NOT NULL
  subscription_id BIGINT REFERENCES subscriptions(id)
  invoice_number TEXT UNIQUE
  amount NUMERIC(10,2) NOT NULL
  tax_amount NUMERIC(10,2) DEFAULT 0
  total_amount NUMERIC(10,2) NOT NULL
  status TEXT  -- 'draft', 'open', 'paid', 'void', 'uncollectible'
  due_date DATE
  paid_at TIMESTAMP
  invoice_pdf_url TEXT
  created_at TIMESTAMP

payment_methods           -- Stored payment methods
  id BIGSERIAL PRIMARY KEY
  tenant_id UUID NOT NULL
  type TEXT  -- 'card', 'bank_account'
  last_four TEXT
  brand TEXT  -- 'visa', 'mastercard', etc.
  expiry_month INT
  expiry_year INT
  is_default BOOLEAN DEFAULT false
  provider TEXT
  provider_payment_method_id TEXT
  created_at TIMESTAMP
  updated_at TIMESTAMP

billing_cycles            -- Billing period tracking
  id BIGSERIAL PRIMARY KEY
  subscription_id BIGINT REFERENCES subscriptions(id)
  start_date DATE NOT NULL
  end_date DATE NOT NULL
  amount NUMERIC(10,2) NOT NULL
  status TEXT  -- 'upcoming', 'current', 'past'
  created_at TIMESTAMP

refunds                   -- Refund tracking
  id BIGSERIAL PRIMARY KEY
  payment_id BIGINT REFERENCES payments(id)
  amount NUMERIC(10,2) NOT NULL
  reason TEXT
  status TEXT  -- 'pending', 'succeeded', 'failed'
  provider_refund_id TEXT
  processed_at TIMESTAMP
  created_at TIMESTAMP

payment_disputes          -- Chargeback tracking
  id BIGSERIAL PRIMARY KEY
  payment_id BIGINT REFERENCES payments(id)
  reason TEXT
  status TEXT  -- 'warning_needs_response', 'won', 'lost'
  evidence JSONB
  created_at TIMESTAMP
  updated_at TIMESTAMP

tax_rates                 -- Tax configuration
  id BIGSERIAL PRIMARY KEY
  tenant_id UUID
  country TEXT NOT NULL
  state TEXT
  rate NUMERIC(5,4) NOT NULL
  description TEXT
  is_active BOOLEAN DEFAULT true
  created_at TIMESTAMP
```

#### Security Considerations

1. **PCI Compliance**: Never store full card numbers
2. **Encryption**: Encrypt sensitive data at rest
3. **Audit**: Log all payment operations
4. **Access Control**: Strict RBAC for financial data

#### Backup Strategy

- **Frequency**: Hourly
- **Retention**: 7 years (regulatory requirement)
- **Encryption**: AES-256 for backups
- **Testing**: Monthly restore tests

---

### 7-14. Other Databases

*(Detailed schemas available upon request)*

**Summary**:
- **statuspage_analytics**: Time-series data, large volume
- **statuspage_monitoring**: Check results, alerts
- **statuspage_event_store**: Event log, append-only
- **statuspage_branding**: Themes, assets
- **statuspage_landing**: Marketing content
- **statuspage_analytics_consumer**: Processed events
- **statuspage_audit_consumer**: Audit logs
- **statuspage_billing_consumer**: Billing events

---

## Database Configuration

### Connection Pool Settings

All services use shared-resilience with standardized connection pools:

```yaml
database:
  max_open_conns: 80      # Development: CPU × 10
  max_open_conns: 200     # Production: CPU × 25
  max_idle_conns: 32      # 40% of max_open
  max_lifetime: 3600      # 1 hour
  max_idle_time: 600      # 10 minutes
  connect_timeout: 10     # 10 seconds
```

### Connection String Format

```
postgresql://<user>:<password>@<host>:<port>/<database>?sslmode=require&connect_timeout=10
```

---

## Backup Strategy

### Backup Tiers

**CRITICAL** (RTO: 1 hour, RPO: 15 minutes):
- tenant_admin_db
- statuspage_user
- statuspage_incident
- statuspage_payment
- statuspage_audit_consumer

**HIGH** (RTO: 4 hours, RPO: 1 hour):
- statuspage_component
- statuspage_notification
- statuspage_monitoring
- statuspage_event_store
- statuspage_billing_consumer

**MEDIUM** (RTO: 24 hours, RPO: 4 hours):
- statuspage_analytics
- statuspage_analytics_consumer

**LOW** (RTO: 72 hours, RPO: 24 hours):
- statuspage_branding
- statuspage_landing

### Backup Methods

1. **pg_dump** (Full Logical Backup):
   ```bash
   pg_dump -U postgres -F c -b -v -f backup_$(date +%Y%m%d).dump database_name
   ```

2. **WAL Archiving** (Point-in-Time Recovery):
   ```
   postgresql.conf:
     wal_level = replica
     archive_mode = on
     archive_command = 'cp %p /mnt/backups/wal/%f'
   ```

3. **pg_basebackup** (Physical Backup):
   ```bash
   pg_basebackup -U replication -D /mnt/backups/base -Fp -Xs -P
   ```

### Backup Schedule

- **Full Backup**: Daily at 2 AM UTC
- **Incremental Backup**: Every 4 hours (CRITICAL databases)
- **WAL Archiving**: Continuous (CRITICAL databases)
- **Backup Retention**: 30 days (configurable)

### Restore Testing

- **Frequency**: Monthly for all databases
- **Process**: Restore to isolated environment, verify data integrity
- **Documentation**: Restore procedures maintained in runbook

---

## Monitoring

### Key Metrics

1. **Connection Pool**:
   - Active connections
   - Idle connections
   - Wait count
   - Wait duration

2. **Query Performance**:
   - Slow query log (queries > 1 second)
   - Query execution plans
   - Index usage statistics

3. **Database Size**:
   - Total database size
   - Table sizes
   - Index sizes
   - Growth rate

4. **Replication**:
   - Replication lag (if using replicas)
   - WAL position
   - Replica health

### Monitoring Tools

- **Prometheus**: Metrics collection
- **Grafana**: Visualization
- **pg_stat_statements**: Query analysis
- **pgBadger**: Log analysis

### Alerts

1. **Connection Pool Exhaustion**: > 90% utilization
2. **Slow Queries**: Queries > 5 seconds
3. **Disk Space**: < 20% free
4. **Replication Lag**: > 10 seconds
5. **Long-Running Transactions**: > 5 minutes

---

## Scaling Strategy

### Vertical Scaling

**Current Limits**:
- PostgreSQL max_connections: 200-400 per database
- Shared buffers: 25% of RAM
- Effective cache size: 75% of RAM

**Scaling Up**:
1. Increase CPU cores (improves connection pool capacity)
2. Increase RAM (improves caching)
3. Use faster SSD storage (improves I/O)

### Horizontal Scaling

**Read Replicas**:
```
Primary (Read/Write)
  ├─ Replica 1 (Read-Only) - Analytics queries
  ├─ Replica 2 (Read-Only) - Reporting
  └─ Replica 3 (Read-Only) - Backup source
```

**Sharding** (for very high scale):
- Shard by tenant_id
- Each shard handles subset of tenants
- Requires application-level routing

**Connection Pooling**:
- Use PgBouncer for connection pooling
- Reduces connection overhead
- Enables more application instances

---

## Maintenance

### Routine Maintenance

**Daily**:
- Monitor slow query log
- Check backup completion
- Verify replication health

**Weekly**:
- Analyze query performance
- Review index usage
- Check table/index bloat

**Monthly**:
- Run VACUUM ANALYZE on all tables
- Update table statistics
- Review and optimize slow queries
- Test backup restore

**Quarterly**:
- Review partitioning strategy
- Evaluate read replica needs
- Capacity planning
- Security audit

### Schema Migration Process

1. **Development**: Test migration locally
2. **Staging**: Run migration on staging environment
3. **Backup**: Take full backup before production migration
4. **Production**: Run migration during maintenance window
5. **Verification**: Validate data integrity post-migration
6. **Rollback Plan**: Have rollback script ready

### VACUUM and ANALYZE

```sql
-- Full VACUUM (during maintenance window)
VACUUM FULL ANALYZE;

-- Regular VACUUM (can run anytime)
VACUUM ANALYZE;

-- Auto-vacuum configuration
ALTER TABLE table_name SET (
  autovacuum_vacuum_scale_factor = 0.1,
  autovacuum_analyze_scale_factor = 0.05
);
```

---

## Security

### Access Control

1. **Service Accounts**: Each service has dedicated PostgreSQL user
2. **Least Privilege**: Users only have access to their database
3. **No Superuser Access**: Production services run without superuser privileges

### Encryption

1. **At Rest**: PostgreSQL data encryption (pgcrypto)
2. **In Transit**: SSL/TLS for all connections
3. **Sensitive Fields**: Additional encryption for PII/PCI data

### Audit Logging

```
postgresql.conf:
  log_statement = 'ddl'  # Log all DDL statements
  log_min_duration_statement = 1000  # Log queries > 1 second
  log_connections = on
  log_disconnections = on
  log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
```

### Compliance

- **GDPR**: Data retention policies, right to deletion
- **SOC 2**: Audit logging, backup retention
- **PCI DSS**: Payment data encryption, access controls

---

## Troubleshooting

### Common Issues

1. **Connection Pool Exhausted**:
   ```sql
   -- Check current connections
   SELECT count(*) FROM pg_stat_activity;

   -- Find long-running queries
   SELECT pid, now() - pg_stat_activity.query_start AS duration, query
   FROM pg_stat_activity
   WHERE state != 'idle'
   ORDER BY duration DESC;

   -- Kill long-running query
   SELECT pg_terminate_backend(pid);
   ```

2. **Slow Queries**:
   ```sql
   -- Explain analyze
   EXPLAIN ANALYZE SELECT * FROM table WHERE condition;

   -- Check missing indexes
   SELECT schemaname, tablename, attname
   FROM pg_stats
   WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
     AND n_distinct > 0
     AND correlation < 0.9;
   ```

3. **Table Bloat**:
   ```sql
   -- Check bloat
   SELECT schemaname, tablename,
          pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
          pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) -
                         pg_relation_size(schemaname||'.'||tablename)) AS bloat
   FROM pg_tables
   ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

   -- Fix bloat
   VACUUM FULL table_name;
   ```

---

## Migration from Other Databases

### From MySQL

Use `pgloader` for automated migration:

```bash
pgloader mysql://user:pass@localhost/sourcedb \
          postgresql://user:pass@localhost/targetdb
```

### From MongoDB

1. Export MongoDB data to JSON
2. Transform to relational schema
3. Import using COPY command

---

## Future Enhancements

### Short Term (Next Quarter)

1. **Read Replicas**: Deploy for analytics-service and monitoring-service
2. **PgBouncer**: Implement connection pooling
3. **Partitioning**: Partition high-volume tables by time

### Long Term (Next Year)

1. **TimescaleDB**: For time-series data (analytics, monitoring)
2. **Citus**: For horizontal sharding of large tenants
3. **PostGIS**: For geographic data (if needed)

---

## Appendix

### Useful Queries

**Database Size**:
```sql
SELECT pg_size_pretty(pg_database_size('database_name'));
```

**Table Sizes**:
```sql
SELECT tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename))
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

**Index Usage**:
```sql
SELECT schemaname, tablename, indexname, idx_scan
FROM pg_stat_user_indexes
ORDER BY idx_scan ASC;
```

**Active Queries**:
```sql
SELECT pid, age(clock_timestamp(), query_start), usename, query
FROM pg_stat_activity
WHERE query != '<IDLE>' AND query NOT ILIKE '%pg_stat_activity%'
ORDER BY query_start DESC;
```

---

**Document Version**: 1.0
**Last Updated**: 2025-10-14
**Maintained By**: Database Team
**Review Frequency**: Quarterly

This is a **living document** - update whenever schema changes occur.
