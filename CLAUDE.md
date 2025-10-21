# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Quick Reference

**Project**: Beakon Status Page Platform
**Architecture**: 20 Go microservices (19 active + 1 deprecated)
**Language**: Go 1.21+
**Database**: PostgreSQL (14 databases, database-per-service pattern)
**Shared Library**: `shared-resilience` (100% adoption across all services)

## Essential Reading

Start with these documents in order:
1. **[AI_CONTEXT.md](AI_CONTEXT.md)** - Quick start for new developers/AI
2. **[SERVICE_CATALOG.md](SERVICE_CATALOG.md)** - Complete service reference
3. **[ARCHITECTURE.md](ARCHITECTURE.md)** - System architecture and service interactions
4. **[DATABASE_ARCHITECTURE.md](DATABASE_ARCHITECTURE.md)** - Database schemas and management

## Repository Structure

```
Beakon/
├── microservices/              # 20 independent services (some are git submodules)
│   ├── api-gateway/           # Port 8080 - Request routing & auth
│   ├── user-service/          # Port 8081 - User auth & management
│   ├── tenant-admin-service/  # Port 8099 - Multi-tenant mgmt & RBAC
│   ├── saas-admin-service/    # Port 8098 - Platform admin
│   ├── landing-page-service/  # Port 8100 - Marketing website
│   ├── shared-resilience/     # Shared library (circuit breakers, DB, middleware)
│   └── ...                    # 15 other services (see SERVICE_CATALOG.md)
├── scripts/                   # Utility scripts
└── docs/                      # Additional documentation
```

**Important**: Each service in `microservices/` may be its own Git repository (submodule). Navigate to specific services to work on them independently.

## Common Development Commands

### Database Setup

**Initialize all databases from scratch:**
```bash
cd microservices
./init-all-databases.sh
```

**Initialize a specific service database:**
```bash
cd microservices
./init-all-databases.sh landing-page-service
```

**Environment variables for database connection:**
```bash
export DB_HOST=localhost        # Default: localhost
export DB_PORT=5432             # Default: 5432
export DB_USER=postgres         # Default: postgres
export DB_PASSWORD=postgres     # Default: postgres
export DB_SSLMODE=disable       # Default: disable (use 'require' in production)
```

**Individual service database initialization:**
```bash
cd microservices/landing-page-service
./init-db.sh
```

### Running Services

**Start core services in development:**
```bash
cd microservices
./start-dev.sh
```

**Start a specific service:**
```bash
cd microservices/tenant-admin-service
./start-dev.sh
# OR
go run cmd/main.go
```

**Check service status:**
```bash
cd microservices
./status-dev.sh
```

**Stop all services:**
```bash
cd microservices
./stop-dev.sh
```

### Building Services

**Build a service:**
```bash
cd microservices/<service-name>
go build -o <service-name> cmd/main.go
```

**Run the compiled binary:**
```bash
./<service-name>
```

### Testing

**Run tests for a service:**
```bash
cd microservices/<service-name>
go test ./...
```

**Run tests with coverage:**
```bash
go test ./... -cover
```

**Run specific test:**
```bash
go test ./internal/handlers -run TestHandlerName
```

### Dependencies

**Update dependencies:**
```bash
cd microservices/<service-name>
go mod tidy
go mod download
```

**Update shared-resilience library:**
```bash
cd microservices/<service-name>
go get github.com/anupamdutta5/shared-resilience@latest
go mod tidy
```

### Database Migrations

**Create a new migration (manual approach):**
```bash
cd microservices/<service-name>/migrations
touch $(date +%Y%m%d%H%M%S)_description.sql
# Edit the SQL file
```

**Apply migrations manually:**
```bash
psql -U postgres -d <database_name> -f migrations/<migration_file>.sql
```

**Using Atlas (if configured):**
```bash
cd microservices/<service-name>
atlas migrate diff <description> --env dev
atlas migrate apply --env dev
```

## Architecture Fundamentals

### Microservices Pattern

**Key Principles:**
- **Database-per-service**: Each service owns its database (pure microservices pattern, no exceptions)
- **API-first**: Services communicate via HTTP/REST APIs
- **Multi-tenant**: All services support multi-tenancy with tenant isolation
- **Stateless**: Services designed for horizontal scaling

### Service Communication

**Standard Pattern:**
```
Client → API Gateway (8080) → Downstream Service
```

**Inter-service calls should also go through API Gateway:**
```
Service A → API Gateway (8080) → Service B
```

**Known Exception:**
- SaaS Admin → Tenant Admin (direct call for session validation)

### Authentication Flow

1. User calls `POST /api/v1/auth/login` → User Service (via API Gateway)
2. User Service validates credentials, generates JWT token
3. Client includes JWT in `Authorization: Bearer <token>` header
4. API Gateway validates JWT and routes to downstream services
5. Downstream services receive validated tenant context from API Gateway

### Multi-Tenancy

**All database tables include `tenant_id` for data isolation:**
```sql
CREATE TABLE components (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,  -- Always present for multi-tenant tables
    name TEXT NOT NULL,
    ...
);
CREATE INDEX idx_components_tenant_id ON components(tenant_id);
```

**Tenant context is injected by middleware from JWT token.**

## Shared Resilience Library

**All services MUST use `shared-resilience` for:**

1. **Database Connections** - CPU-based connection pooling
2. **Circuit Breakers** - Protect against cascade failures
3. **Health Checks** - `/health`, `/health/live`, `/health/ready` endpoints
4. **Middleware** - JWT auth, CORS, rate limiting, security headers
5. **Rate Limiting** - Per-IP, per-user, per-tenant
6. **Caching** - In-memory and Redis with fallback
7. **Error Handling** - Standardized error responses
8. **Graceful Shutdown** - Signal handling and cleanup

**Example usage in a new service:**
```go
import (
    "github.com/anupamdutta5/shared-resilience"
)

// Database connection
db, err := resilience.NewDatabaseConnection(config)

// Circuit breaker
breaker := resilience.NewCircuitBreaker("service-name")

// Health checks
healthChecker := resilience.NewHealthChecker()
healthChecker.AddCheck("database", dbHealthCheck)
```

## Database Architecture

### Active Databases (14)

| Database | Service | Notes |
|----------|---------|-------|
| `saas_admin` | saas-admin-service | Platform admin, plans, features, pricing |
| `tenant_admin_db` | tenant-admin-service | Tenant management, RBAC |
| `statuspage_user` | user-service | User authentication |
| `statuspage_component` | component-service | Component status |
| `statuspage_incident` | incident-service | Incident management |
| `statuspage_payment` | payment-service | Billing & payments |
| `statuspage_landing` | landing-page-service | Marketing content |
| ... | ... | See DATABASE_ARCHITECTURE.md for full list |

### Database Connection Configuration

**Standard environment variables for each service:**
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=<service_specific_database>
DB_SSLMODE=disable  # Use 'require' in production
```

**Connection pool settings (from shared-resilience):**
- Development: 80 max connections (CPU × 10)
- Production: 200 max connections (CPU × 25)
- Max idle: 40% of max connections
- Connection lifetime: 1 hour
- Idle timeout: 10 minutes

## Port Allocation

**Frontend Services (Next.js SSR):**
- 3001: SaaS Admin Frontend ⭐ (platform admin UI)
- 3002: Tenant Admin Frontend ⭐ (multi-tenant admin UI)

**Backend HTTP Services:**
- 8080: API Gateway ⭐ (entry point)
- 8081: User Service
- 8084: Component Service
- 8085: Notification Service
- 8086: Incident Service
- 8088: Payment Service
- 8090: Analytics Service
- 8092: Monitoring Service
- 8093: Status UI Service
- 8095: Database Service ⚠️ DEPRECATED
- 8096: Event Store Service
- 8097: Branding Service
- 8098: SaaS Admin Service (API only - UI separated to port 3001)
- 8099: Tenant Admin Service (API only - UI separated to port 3002)
- 8100: Landing Page Service

**Consumer Services (no HTTP port):**
- Analytics Consumer
- Notification Consumer
- Audit Consumer
- Billing Consumer

**Prometheus Metrics:** Backend service port + 1010 (e.g., 8080 → 9090)

## Development Workflow

### Starting from Scratch

1. **Install prerequisites:**
   ```bash
   # macOS
   brew install postgresql@16 go redis nodejs
   brew services start postgresql@16

   # Verify versions
   go version       # Should be 1.21+
   node -v          # Should be 18.0+
   npm -v           # Should be 8.0+
   ```

2. **Initialize databases:**
   ```bash
   cd microservices
   ./init-all-databases.sh
   ```

3. **Start ALL services (backends + frontends):**
   ```bash
   cd microservices
   ./start-all-services.sh
   ```
   This starts:
   - Backend services (saas-admin-service on 8098, tenant-admin-service on 8099)
   - Frontend services (saas-admin-frontend on 3001, tenant-admin-frontend on 3002)

   **OR start individually:**
   ```bash
   # Start only backend services
   ./start-all-backends.sh

   # Start only frontend services
   ./start-all-frontends.sh
   ```

4. **Verify health:**
   ```bash
   # Backend health
   curl http://localhost:8098/api/v1/health
   curl http://localhost:8099/health

   # Frontend access
   open http://localhost:3001                 # SaaS Admin UI
   open http://anupam.localhost:3002          # Tenant Admin UI (subdomain required)
   ```

5. **Stop all services:**
   ```bash
   cd microservices
   ./stop-all-services.sh

   # Or stop individually:
   ./stop-all-backends.sh
   ./stop-all-frontends.sh
   ```

### Working on a Single Service

1. **Navigate to service:**
   ```bash
   cd microservices/tenant-admin-service
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Run service:**
   ```bash
   ./start-dev.sh
   # OR with custom environment
   export SERVER_PORT=8099
   export DB_NAME=tenant_admin_db
   go run cmd/main.go
   ```

4. **Test changes:**
   ```bash
   go test ./...
   ```

### Adding a New Service

1. **Create service directory:**
   ```bash
   cd microservices
   mkdir my-new-service
   cd my-new-service
   ```

2. **Initialize Go module:**
   ```bash
   go mod init github.com/anupamdutta5/my-new-service
   ```

3. **Add shared-resilience dependency:**
   ```bash
   go get github.com/anupamdutta5/shared-resilience@latest
   ```

4. **Create standard structure:**
   ```
   my-new-service/
   ├── cmd/
   │   └── main.go              # Entry point
   ├── internal/
   │   ├── handlers/            # HTTP handlers
   │   ├── services/            # Business logic
   │   ├── models/              # Data models
   │   └── middleware/          # Service-specific middleware
   ├── migrations/              # Database migrations
   ├── atlas.hcl                # Atlas migration config
   ├── init-db.sh               # Database initialization script
   ├── start-dev.sh             # Development startup script
   └── README.md
   ```

5. **Implement health checks:**
   ```go
   // Required endpoints: /health, /health/live, /health/ready
   // Use shared-resilience health checker
   ```

6. **Update documentation:**
   - Add to SERVICE_CATALOG.md
   - Update ARCHITECTURE.md
   - Update DATABASE_ARCHITECTURE.md (if has database)
   - Update this file (CLAUDE.md)

## Critical Patterns & Best Practices

### Session Management (Tenant Admin Service)

Three-tier session system:
1. **Primary**: Redis (fastest)
2. **Fallback**: PostgreSQL database
3. **Emergency**: In-memory cache

Always validate sessions using the three-tier approach for resilience.

### RBAC (Role-Based Access Control)

Implemented in `tenant-admin-service`:
- Roles: owner, admin, manager, viewer
- Permissions are granular and customizable
- User-role assignments support multiple roles per user
- Team-based access control available

### Max Users Enforcement

`tenant-admin-service` enforces max_users limit:
- Validation occurs before user creation
- Returns 400 error if limit exceeded
- Limit stored in `tenants.max_users` column

### Error Handling

Use shared-resilience error types:
```go
if err != nil {
    return resilience.NewAPIError(
        http.StatusBadRequest,
        "VALIDATION_ERROR",
        "Invalid input",
        err,
    )
}
```

### Logging

All services use structured logging (zap):
```go
logger.Info("Processing request",
    zap.String("tenant_id", tenantID),
    zap.String("user_id", userID),
    zap.String("correlation_id", correlationID),
)
```

## Troubleshooting

### Port Already in Use
```bash
# Find process using port
lsof -i :8099
# Kill process
kill -9 <PID>
```

### Database Connection Failed
```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# List databases
psql -U postgres -l

# Create missing database
psql -U postgres -c "CREATE DATABASE tenant_admin_db;"
```

### Service Won't Start
```bash
# Check logs
cat logs/<service>.log

# Check environment variables
env | grep DB_
env | grep JWT_

# Verify dependencies
cd microservices/<service>
go mod verify
```

### Migration Issues
```bash
# Check current schema
psql -U postgres -d <database> -c '\dt'

# Manually apply migration
psql -U postgres -d <database> -f migrations/<file>.sql
```

### Redis Connection Issues
Services gracefully fall back to in-memory cache if Redis is unavailable. To fix:
```bash
brew services start redis
# OR
redis-server
```

## Testing Endpoints

### Health Checks
```bash
curl http://localhost:8099/health
curl http://localhost:8099/health/live
curl http://localhost:8099/health/ready
```

### Authentication
```bash
# Login (Tenant Admin)
curl -X POST http://localhost:8099/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password"}'

# Create tenant
curl -X POST http://localhost:8099/api/v1/public/tenants \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Tenant","email":"tenant@example.com"}'
```

### With Authentication
```bash
# Get tenants (authenticated)
curl http://localhost:8099/api/v1/tenants \
  -H "Authorization: Bearer <token>"
```

## Special Considerations

### Service Separation

**Important**: As of 2025-10-21, SaaS Admin and Tenant Admin services have been split into separate frontend and backend microservices.

**Frontend Services (Next.js SSR):**

**saas-admin-frontend** (port 3001):
- User interface for platform administration
- React-based UI with TypeScript
- Communicates with saas-admin-service (port 8098) via CORS
- Authentication: JWT tokens in localStorage
- Deployment: Docker-ready with Next.js standalone mode
- Repository: https://github.com/anupamdutta5/saas-admin-frontend

**tenant-admin-frontend** (port 3002):
- Multi-tenant admin interface
- React-based UI with TypeScript
- Communicates with tenant-admin-service (port 8099) via same-origin (subdomain routing)
- Authentication: JWT tokens + session cookies
- Deployment: Docker-ready with Next.js standalone mode
- Repository: https://github.com/anupamdutta5/tenant-admin-frontend
- **Requires subdomain DNS** for multi-tenant routing (e.g., `anupam.localhost:3002`)

**Backend Services (Go APIs):**

**SaaS Admin Service** (port 8098) uses `saas_admin` database:
- **API-only** (frontend separated to port 3001)
- Platform administration API
- Subscription plans, features, pricing API
- SaaS admin user management API
- Creates tenants by calling Tenant Admin Service API
- RabbitMQ event publishing for tenant sync
- CORS enabled for frontend (port 3001)

**Tenant Admin Service** (port 8099) uses `tenant_admin_db` database:
- **API-only** (frontend separated to port 3002)
- Multi-tenant management API
- Tenant users, roles, permissions, teams API
- Component, incident, subscriber management API
- Session management and RBAC
- Subdomain-based tenant isolation middleware
- RabbitMQ event consumer for tenant sync
- Same-origin requests (no CORS needed due to subdomain routing)

**Communication**:
- Frontend ↔ Backend: HTTP/HTTPS with CORS (SaaS Admin) or same-origin (Tenant Admin)
- Backend ↔ Backend: HTTP APIs (pure microservices, database-per-service)
- Event-driven: RabbitMQ for tenant synchronization between saas-admin and tenant-admin services

### Deprecated Services

**database-service** (port 8095) is deprecated:
- Not actively used by any service
- Functionality moved to shared-resilience library
- Can be removed from active deployment

### Git Submodules

Some services (api-gateway, user-service, status-ui-service) are Git submodules:
```bash
# Update submodules
git submodule update --init --recursive

# Update specific submodule
cd microservices/api-gateway
git pull origin main
```

## Environment-Specific Configuration

### Development
```bash
export ENVIRONMENT=development
export JWT_SECRET=dev-secret-for-testing-only-change-in-production
export DB_SSLMODE=disable
export REDIS_ENABLED=false
export LOG_LEVEL=debug
```

### Production
```bash
export ENVIRONMENT=production
export JWT_SECRET=<strong-random-secret-min-32-chars>
export DB_SSLMODE=require
export DB_HOST=<rds-endpoint>
export REDIS_ENABLED=true
export REDIS_HOST=<redis-endpoint>
export LOG_LEVEL=info
```

## Metrics & Observability

### Health Endpoints
All services expose:
- `GET /health` - Overall health with component details
- `GET /health/live` - Liveness probe (Kubernetes)
- `GET /health/ready` - Readiness probe (Kubernetes)

### Prometheus Metrics
Services expose metrics on port + 1010:
```bash
curl http://localhost:9090/metrics  # API Gateway metrics
curl http://localhost:9099/metrics  # Tenant Admin metrics
```

### Logging
Structured JSON logs with:
- Timestamp
- Service name
- Log level
- Correlation ID
- Tenant ID (if applicable)
- User ID (if applicable)

## Security

### JWT Tokens
- Generated by user-service
- Validated by API Gateway
- Include tenant_id and user_id claims
- Expiration configurable (default 24h)
- Secret MUST be ≥32 characters in production

### Password Hashing
- bcrypt with cost factor 10
- Never store plaintext passwords
- Password reset tokens expire after use

### Rate Limiting
Configured in shared-resilience:
- Per-IP limiting
- Per-user limiting
- Per-tenant limiting
- Configurable burst sizes

### Security Headers
All services include (via shared-resilience middleware):
- Content-Security-Policy
- X-Frame-Options: DENY
- X-Content-Type-Options: nosniff
- Strict-Transport-Security (HSTS)

## Common Pitfalls

1. **Forgetting tenant_id**: Always include tenant scoping in queries
2. **Direct database access**: Services should only access their own database
3. **Bypassing API Gateway**: Client requests should go through gateway
4. **Missing health checks**: All new services must implement health endpoints
5. **Ignoring shared-resilience**: Always use library for DB, auth, and middleware
6. **Hardcoded secrets**: Use environment variables for all sensitive data
7. **Missing migrations**: Track all schema changes in migrations/ directory

## Resources

- **Go Documentation**: https://go.dev/doc/
- **GORM Documentation**: https://gorm.io/docs/
- **PostgreSQL Documentation**: https://www.postgresql.org/docs/
- **Atlas Migrations**: https://atlasgo.io/
- **Prometheus Metrics**: https://prometheus.io/docs/

## Getting Help

For issues or questions:
1. Check service-specific README.md
2. Review logs: `tail -f logs/<service>.log`
3. Verify environment variables are set correctly
4. Check database connectivity: `psql -U postgres -d <database> -c '\q'`
5. Test individual service health: `curl http://localhost:<port>/health`


Read all the .md files in the root /Bekon directory and all microservice specififc documents and .md files inside all the directories under /microservices directory. Ensure to have a proper understanding of the project before making cnages and taking up tasks as services are interdependant and careless changes can break the other parts

Never run any git restore, reset, delete commands without permission