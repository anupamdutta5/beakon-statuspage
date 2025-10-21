# PostgreSQL for Beakon Status Page Platform

Dockerized PostgreSQL 16 setup for all Beakon microservices.

## Quick Start

```bash
# Start PostgreSQL
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f

# Stop PostgreSQL
docker-compose down

# Stop and remove data (⚠️ destroys all data)
docker-compose down -v
```

## Databases Created

This setup automatically creates 14 databases:

1. **saas_admin** - SaaS Admin Service
2. **tenant_admin_db** - Tenant Admin Service (shared with saas-admin)
3. **statuspage_user** - User Service
4. **statuspage_component** - Component Service
5. **statuspage_incident** - Incident Service
6. **statuspage_payment** - Payment Service
7. **statuspage_landing** - Landing Page Service
8. **statuspage_notification** - Notification Service
9. **statuspage_analytics** - Analytics Service
10. **statuspage_monitoring** - Monitoring Service
11. **statuspage_eventstore** - Event Store Service
12. **statuspage_branding** - Branding Service
13. **statuspage_ui** - Status UI Service
14. **statuspage_audit** - Audit Service

## Connection Details

- **Host**: localhost
- **Port**: 5432
- **User**: postgres
- **Password**: postgres
- **Default DB**: postgres

## Environment Variables

Services should use these environment variables:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=<database_name>  # e.g., saas_admin
export DB_SSLMODE=disable       # Use 'require' in production
```

## Connecting to PostgreSQL

```bash
# Using psql command line
docker exec -it statuspage-postgres psql -U postgres

# List all databases
\l

# Connect to specific database
\c saas_admin

# List tables
\dt

# Exit
\q
```

## Backup & Restore

### Backup single database
```bash
docker exec -it statuspage-postgres pg_dump -U postgres saas_admin > saas_admin_backup.sql
```

### Restore single database
```bash
docker exec -i statuspage-postgres psql -U postgres saas_admin < saas_admin_backup.sql
```

### Backup all databases
```bash
docker exec -it statuspage-postgres pg_dumpall -U postgres > all_databases_backup.sql
```

### Restore all databases
```bash
docker exec -i statuspage-postgres psql -U postgres < all_databases_backup.sql
```

## Performance Tuning

The `postgresql.conf` file includes optimizations for:
- Connection pooling (200 max connections)
- Memory allocation (256MB shared buffers)
- Query planning
- Write-ahead logging
- Autovacuum

## Logs

View PostgreSQL logs:
```bash
docker-compose logs -f postgres
```

Logs are also available inside the container at `/var/lib/postgresql/data/pgdata/pg_log/`

## Health Check

PostgreSQL includes a health check that runs every 10 seconds:
```bash
docker-compose ps
# STATUS column should show "(healthy)"
```

## Troubleshooting

### Port already in use
If port 5432 is already in use:
```bash
# Find process using port
lsof -i :5432

# Kill local postgres (if needed)
brew services stop postgresql@16
```

### Database doesn't exist
If database wasn't created automatically:
```bash
docker exec -it statuspage-postgres psql -U postgres -c "CREATE DATABASE saas_admin;"
```

### Reset everything
```bash
docker-compose down -v
docker-compose up -d
```

## Production Configuration

For production, update `docker-compose.yml`:
1. Change POSTGRES_PASSWORD to a strong password
2. Mount SSL certificates
3. Enable SSL in postgresql.conf
4. Restrict listen_addresses
5. Use managed PostgreSQL service (AWS RDS, GCP Cloud SQL, etc.)

## Architecture

- **Image**: postgres:16-alpine (lightweight)
- **Data**: Persisted in Docker volume `postgres_data`
- **Network**: Bridge network `statuspage-network`
- **Init Script**: `init-databases.sql` runs once on first start
- **Config**: Custom `postgresql.conf` for optimization
