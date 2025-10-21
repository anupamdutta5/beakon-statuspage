# Database Service

> **⚠️ EXPERIMENTAL / DEPRECATED**
>
> This service is **NOT recommended for production use** and is **NOT currently used** by core platform services (SaaS Admin, Tenant Admin, etc.).
>
> **Why this exists:** Originally designed as a centralized database proxy layer for microservices.
>
> **Why it's not used:**
> - **Single Point of Failure (SPOF):** All services depending on this would fail if it goes down, even if PostgreSQL is healthy
> - **Added Latency:** Extra network hop reduces performance
> - **Scaling Bottleneck:** Proxy layer can become a bottleneck under load
> - **Unnecessary Complexity:** The `shared-resilience.DatabaseManager` already provides connection pooling, health checks, and resilience patterns
>
> **Recommended Approach:** Services should connect directly to PostgreSQL using `shared-resilience.DatabaseManager`, as demonstrated in saas-admin-service and tenant-admin-service.
>
> **Potential Use Cases (if any):**
> - Centralized query auditing for compliance (not currently implemented)
> - Cross-service query result caching (questionable value vs SPOF risk)
> - Database migration coordination (better done via separate migration tool)

## Overview
The Database Service provides centralized database connection management, query execution, and data access abstraction for the Beakon status page platform. It acts as a database proxy layer, handling connection pooling, query optimization, and providing a unified interface for database operations across all microservices.

## Key Features

### Database Connection Management
- **Connection Pooling**: Centralized connection pool management
- **Multi-tenant Database Routing**: Route queries to appropriate tenant databases
- **Read/Write Splitting**: Separate read replicas from write masters
- **Connection Health Monitoring**: Automatic health checks and reconnection

### Query Management
- **Query Execution**: Execute SQL queries with proper parameterization
- **Prepared Statements**: Cached prepared statements for performance
- **Transaction Management**: Support for multi-statement transactions
- **Query Timeout**: Configurable timeouts for all database operations

### Data Access Patterns
- **CRUD Operations**: Standardized create, read, update, delete interfaces
- **Batch Operations**: Efficient bulk inserts and updates
- **Streaming Results**: Memory-efficient result streaming for large datasets
- **Query Builder**: Programmatic query construction

### Resilience Features
- **Circuit Breaker**: Database circuit breaker protection
- **Retry Logic**: Automatic retry with exponential backoff
- **Graceful Degradation**: Fallback to read-only mode on issues
- **Connection Recovery**: Automatic reconnection on failures

### Monitoring & Observability
- **Query Logging**: Log slow queries and errors
- **Metrics Collection**: Connection pool metrics, query latency
- **Health Checks**: Database connectivity and performance checks
- **Audit Logging**: Track all database access for security

## API Endpoints

### Health & Monitoring
- `GET /health` - Service health check
- `GET /health/db` - Database connectivity check
- `GET /health/pool` - Connection pool status

### Query Execution (Internal API)
- `POST /api/v1/query` - Execute SELECT query
- `POST /api/v1/exec` - Execute INSERT/UPDATE/DELETE
- `POST /api/v1/transaction` - Execute transaction

### Connection Management
- `GET /api/v1/pool/stats` - Get connection pool statistics
- `POST /api/v1/pool/reset` - Reset connection pool
- `GET /api/v1/connections` - List active connections

### Database Administration
- `POST /api/v1/migrations/run` - Run database migrations
- `GET /api/v1/migrations/status` - Get migration status
- `POST /api/v1/backup` - Trigger database backup
- `GET /api/v1/backup/status` - Get backup status

## Dependencies

### External Dependencies
- **shared-resilience**: Common patterns and utilities
- **PostgreSQL**: Primary database system
- **Redis**: Caching for query results (optional)
- **Zap Logger**: Structured logging

## Configuration

### Environment Variables
- `SERVER_PORT`: HTTP server port (default: 8080)
- `DB_HOST`: Database host
- `DB_PORT`: Database port (default: 5432)
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `DB_SSLMODE`: SSL mode (disable/require/verify-full)
- `DB_MAX_OPEN_CONNS`: Maximum open connections (default: 25)
- `DB_MAX_IDLE_CONNS`: Maximum idle connections (default: 5)
- `DB_CONN_MAX_LIFETIME`: Connection max lifetime (default: 5m)
- `DB_CONN_MAX_IDLE_TIME`: Connection max idle time (default: 5m)
- `QUERY_TIMEOUT`: Default query timeout (default: 30s)
- `ENVIRONMENT`: Runtime environment (development/production)

### Read Replica Configuration (Optional)
- `DB_READ_HOST`: Read replica host
- `DB_READ_PORT`: Read replica port
- `DB_READ_USER`: Read replica username
- `DB_READ_PASSWORD`: Read replica password

## Development

### Running the Service
```bash
cd microservices/database-service
go run cmd/main.go
```

### Building
```bash
go build -o database-service cmd/main.go
```

### Testing
```bash
go test ./...
```

### Project Structure
```
database-service/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── server/              # HTTP server
│   ├── database/            # Database connection management
│   ├── query/               # Query execution
│   └── migration/           # Database migration tools
├── pkg/
│   └── logger/              # Logging utilities
└── go.mod                   # Go module definition
```

## Architecture

### Connection Pool Management
```
Service Request → Database Service → Connection Pool → PostgreSQL
                                   ↓
                               Read Replica (optional)
```

### Multi-Tenancy Support
- **Shared Database**: All tenants in one database with tenant_id filtering
- **Database Per Tenant**: Separate database per tenant (future)
- **Schema Per Tenant**: Separate schema per tenant (future)

### Query Execution Flow
1. Receive query request with tenant context
2. Validate and sanitize query
3. Get connection from pool
4. Execute query with timeout
5. Stream results back to caller
6. Return connection to pool
7. Log query metrics

## Database Schema Management

### Migration System
- **Up/Down Migrations**: Reversible database changes
- **Version Control**: Track applied migrations
- **Rollback Support**: Undo migrations if needed
- **Seed Data**: Initialize databases with default data

### Schema Versioning
```
migrations/
├── 001_initial_schema.up.sql
├── 001_initial_schema.down.sql
├── 002_add_tenants.up.sql
├── 002_add_tenants.down.sql
```

## Performance Optimization

### Connection Pooling
- Optimal pool size based on workload
- Connection recycling to prevent leaks
- Idle connection cleanup

### Query Optimization
- Prepared statement caching
- Query result caching (Redis)
- Batch operations for efficiency
- Index recommendations

### Read/Write Splitting
- Route SELECT queries to read replicas
- Route INSERT/UPDATE/DELETE to master
- Automatic failover to master if replica unavailable

## Monitoring

### Key Metrics
- **Connection Pool**: Active, idle, max connections
- **Query Performance**: Latency (p50, p95, p99), throughput
- **Error Rate**: Failed queries, connection errors
- **Slow Queries**: Queries exceeding threshold
- **Transaction Duration**: Long-running transactions

### Health Checks
- Database connectivity
- Connection pool health
- Read replica health (if configured)
- Query execution test

### Alerting
- Connection pool exhaustion
- Slow queries exceeding threshold
- Database connection failures
- High error rate

## Security

### Access Control
- Service-to-service authentication
- API key validation
- IP whitelisting for admin endpoints
- Rate limiting per service

### SQL Injection Prevention
- Parameterized queries only
- Input validation and sanitization
- Query allowlist for known patterns
- No dynamic SQL construction

### Data Protection
- TLS/SSL for database connections
- Encrypted credentials storage
- Audit logging for all access
- Data masking for sensitive fields

## Error Handling

### Connection Errors
- Automatic retry with exponential backoff
- Circuit breaker to prevent cascading failures
- Graceful degradation to read-only mode
- Alert on sustained connection failures

### Query Errors
- Detailed error logging
- Query error categorization
- Automatic rollback on transaction errors
- Dead letter queue for failed queries

## Scaling

### Horizontal Scaling
- Multiple database service instances
- Stateless design for easy scaling
- Load balancing across instances

### Database Scaling
- Connection pooling for efficiency
- Read replicas for read scalability
- Database sharding (future)
- Caching layer to reduce database load

## Best Practices

### Service Integration
- Use database service instead of direct database access
- Implement retry logic in calling services
- Set appropriate query timeouts
- Use batch operations for multiple queries

### Query Guidelines
- Always use parameterized queries
- Avoid N+1 query patterns
- Use transactions for multi-step operations
- Index frequently queried columns

This service provides a robust, scalable, and secure database access layer for the entire platform, ensuring efficient and reliable data operations across all microservices.
