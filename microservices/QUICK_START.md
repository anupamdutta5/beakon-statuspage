# Beakon Microservices - Quick Start Guide

**Last Updated:** 2025-10-21
**Time to Setup:** 15-30 minutes
**Prerequisites:** Docker, PostgreSQL, Go 1.21+, Node.js 18+ (for frontends)

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Setup (5 Minutes)](#quick-setup-5-minutes)
3. [Start All Services](#start-all-services)
4. [Verify Everything Works](#verify-everything-works)
5. [Access the Applications](#access-the-applications)
6. [Common Commands](#common-commands)
7. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Software

```bash
# macOS
brew install postgresql@16 go redis node
brew services start postgresql@16
brew services start redis

# Verify versions
go version    # Should be 1.21+
node --version # Should be 18+
psql --version # Should be 13+
redis-cli --version
```

### Environment Variables

```bash
# Add to ~/.zshrc or ~/.bashrc
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_SSLMODE=disable
export JWT_SECRET=dev-secret-for-testing-only-change-in-production-min-32-chars-long
export RABBITMQ_URL=amqp://admin:SecureP@ssw0rd2024!@localhost:5672/
```

---

## Quick Setup (5 Minutes)

### 1. Initialize All Databases

```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices
./init-all-databases.sh
```

**Expected Output:**
```
✓ Created database: saas_admin
✓ Created database: tenant_admin_db
✓ Created database: statuspage_user
... (14 databases total)
```

### 2. Start RabbitMQ (Event Bus)

```bash
cd rabbitmq
./manage.sh up
```

**Verify:**
- RabbitMQ Management UI: http://localhost:15672
- Username: `admin` / Password: `SecureP@ssw0rd2024!`

### 3. Start Redis (Caching)

```bash
brew services start redis
# OR
redis-server
```

---

## Start All Services

### Option 1: Start Everything (Recommended for Testing)

```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices
./start-all-services.sh
```

This starts:
- ✅ 2 Backend Services (SaaS Admin, Tenant Admin)
- ✅ 2 Frontend Services (SaaS Admin Frontend, Tenant Admin Frontend)
- ✅ Automatic health checks
- ✅ Log files in `/tmp/`

### Option 2: Start Individual Services

#### Backend Services

```bash
# SaaS Admin Backend (Port 8098)
cd saas-admin-service
./saas-admin-service

# Tenant Admin Backend (Port 8099)
cd tenant-admin-service
./tenant-admin-service
```

#### Frontend Services

```bash
# SaaS Admin Frontend (Port 3001)
cd saas-admin-frontend
npm install
npm run dev

# Tenant Admin Frontend (Port 3002)
cd tenant-admin-frontend
npm install
npm run dev
```

---

## Verify Everything Works

### 1. Check Backend Services

```bash
# SaaS Admin Backend
curl http://localhost:8098/api/v1/health
# Expected: {"service":"saas-admin-service","status":"healthy",...}

# Tenant Admin Backend
curl http://localhost:8099/health
# Expected: {"service":"beakon-service","status":"healthy",...}
```

### 2. Check Frontend Services

```bash
# SaaS Admin Frontend
curl -I http://localhost:3001
# Expected: HTTP/1.1 307 Temporary Redirect (middleware working)

# Tenant Admin Frontend
curl -I http://localhost:3002
# Expected: HTTP/1.1 307 Temporary Redirect (middleware working)
```

### 3. Check Database Connections

```bash
# List all databases
psql -U postgres -l | grep statuspage

# Check tenant data
docker exec statuspage-postgres psql -U postgres -d tenant_admin_db -c "SELECT COUNT(*) FROM tenants;"
```

### 4. Check RabbitMQ

```bash
# Via CLI
curl -u admin:SecureP@ssw0rd2024! http://localhost:15672/api/queues

# Via Browser
# Open http://localhost:15672
# Login: admin / SecureP@ssw0rd2024!
```

---

## Access the Applications

### SaaS Admin Portal (Platform Administration)

**URL:** http://localhost:3001

**Default Login:**
- Username: `admin`
- Password: `admin`

**Features:**
- Manage all tenants
- Create/edit subscription plans
- View platform statistics
- Launch individual tenant admin panels

### Tenant Admin Portal (Multi-Tenant Management)

**URL:** http://{subdomain}.localhost:3002

**Example Tenants:**
- http://anupam.localhost:3002
- http://monday.localhost:3002

**Setup `/etc/hosts`:**
```bash
sudo nano /etc/hosts

# Add these lines:
127.0.0.1 anupam.localhost
127.0.0.1 monday.localhost
```

**Default Login:** (varies per tenant)
- Email: `admin@example.com`
- Password: `password`

---

## Common Commands

### Start/Stop Services

```bash
# Start all services
./start-all-services.sh

# Start only backends
./start-all-backends.sh

# Start only frontends
./start-all-frontends.sh

# Stop all services
./stop-all-services.sh
```

### Database Operations

```bash
# Reinitialize all databases (DESTRUCTIVE)
./init-all-databases.sh

# Initialize specific service database
cd user-service
./init-db.sh

# Connect to database
psql -U postgres -d tenant_admin_db
```

### View Logs

```bash
# SaaS Admin Backend
tail -f /tmp/saas-admin-backend.log

# Tenant Admin Backend
tail -f /tmp/tenant-admin-backend.log

# SaaS Admin Frontend
tail -f /tmp/saas-admin-frontend.log

# Tenant Admin Frontend
tail -f /tmp/tenant-admin-frontend.log
```

### Health Checks

```bash
# Quick health check all services
for port in 3001 3002 8098 8099; do
  echo "Port $port: $(curl -s -o /dev/null -w '%{http_code}' http://localhost:$port/health 2>/dev/null || echo 'DOWN')"
done
```

### Build Services

```bash
# Build backend service
cd saas-admin-service
go build -o saas-admin-service cmd/main.go

# Build frontend
cd saas-admin-frontend
npm run build
```

---

## Troubleshooting

### Port Already in Use

```bash
# Find process using port
lsof -i :8098

# Kill process
kill -9 <PID>

# Or kill all by name
pkill -f saas-admin-service
```

### Database Connection Failed

```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Restart PostgreSQL (macOS)
brew services restart postgresql@16

# Check if database exists
psql -U postgres -l | grep saas_admin

# Create database if missing
psql -U postgres -c "CREATE DATABASE saas_admin;"
```

### Frontend Won't Start

```bash
# Clear Next.js cache
cd saas-admin-frontend
rm -rf .next node_modules
npm install
npm run dev
```

### RabbitMQ Connection Failed

```bash
# Check RabbitMQ status
cd rabbitmq
./manage.sh health

# Restart RabbitMQ
./manage.sh restart

# View RabbitMQ logs
docker logs beakon-rabbitmq
```

### Redis Connection Failed

```bash
# Check if Redis is running
redis-cli ping
# Expected: PONG

# Start Redis
brew services start redis

# Check Redis status
brew services list | grep redis
```

### Authentication Issues

```bash
# Clear browser cookies and localStorage
# Open DevTools → Application → Clear storage

# Reset database sessions
psql -U postgres -d tenant_admin_db -c "TRUNCATE sessions;"

# Check JWT secret is set
echo $JWT_SECRET
```

### Service Won't Build

```bash
# Update dependencies
cd saas-admin-service
go mod tidy
go mod download

# Update shared-resilience library
go get github.com/anupamdutta5/shared-resilience@latest
go mod tidy
```

---

## Testing Endpoints

### SaaS Admin API

```bash
# Get all tenants
curl http://localhost:8098/api/v1/tenants

# Get all plans
curl http://localhost:8098/api/v1/plans

# Create tenant
curl -X POST http://localhost:8098/api/v1/tenants \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Test Company",
    "slug": "test-company",
    "contact_email": "test@example.com",
    "billing_email": "billing@test.com",
    "admin_email": "admin@test.com",
    "admin_password": "SecurePass123",
    "plan_id": "uuid-here",
    "max_users": 50
  }'
```

### Tenant Admin API

```bash
# Login (get JWT token)
curl -X POST http://localhost:8099/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "admin@example.com",
    "password": "password"
  }'

# Get components (with auth)
curl -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Host: anupam.localhost:8099" \
  http://localhost:8099/api/v1/components
```

---

## Development Workflow

### 1. Make Code Changes

```bash
# Edit files in your IDE
code microservices/saas-admin-service
```

### 2. Test Locally

```bash
# Run tests
cd saas-admin-service
go test ./...

# Run with hot reload (using air)
air
```

### 3. Verify Changes

```bash
# Restart service
pkill -f saas-admin-service
./saas-admin-service

# Test endpoint
curl http://localhost:8098/api/v1/health
```

### 4. Check Logs

```bash
# View service logs
tail -f /tmp/saas-admin-backend.log

# View with filtering
tail -f /tmp/saas-admin-backend.log | grep ERROR
```

---

## Next Steps

After completing this quick start:

1. **Read Core Documentation:**
   - [SERVICE_CATALOG.md](../SERVICE_CATALOG.md) - All services reference
   - [DATABASE_ARCHITECTURE.md](../DATABASE_ARCHITECTURE.md) - Database schemas
   - [AUTHENTICATION_GUIDE.md](../AUTHENTICATION_GUIDE.md) - Auth & sessions

2. **Explore Specific Features:**
   - [FRONTEND_GUIDE.md](./FRONTEND_GUIDE.md) - Frontend architecture
   - [docs/testing/TESTING_GUIDE.md](./docs/testing/TESTING_GUIDE.md) - Complete testing guide
   - [docs/architecture/API_GATEWAY_COMMUNICATION_GUIDE.md](./docs/architecture/API_GATEWAY_COMMUNICATION_GUIDE.md) - Service communication

3. **Production Deployment:**
   - [DEPLOYMENT_GUIDE.md](../DEPLOYMENT_GUIDE.md) - Kubernetes/Docker deployment
   - [docs/testing/SECURITY_ROADMAP.md](./docs/testing/SECURITY_ROADMAP.md) - Security hardening

---

## Quick Reference

### Port Allocation

| Service | Port | Type | Description |
|---------|------|------|-------------|
| SaaS Admin Frontend | 3001 | HTTP | Platform admin UI |
| Tenant Admin Frontend | 3002 | HTTP | Tenant management UI |
| SaaS Admin Backend | 8098 | HTTP | Platform admin API |
| Tenant Admin Backend | 8099 | HTTP | Tenant management API |
| PostgreSQL | 5432 | TCP | Database |
| Redis | 6379 | TCP | Cache |
| RabbitMQ | 5672 | AMQP | Message queue |
| RabbitMQ Management | 15672 | HTTP | Admin UI |

### Service Dependencies

```
SaaS Admin Frontend (3001)
  → SaaS Admin Backend (8098)
      → PostgreSQL (saas_admin)
      → RabbitMQ (tenant.created events)

Tenant Admin Frontend (3002)
  → Tenant Admin Backend (8099)
      → PostgreSQL (tenant_admin_db)
      → Redis (sessions)
      → RabbitMQ (consumer)
```

### Default Credentials

**SaaS Admin:**
- Username: `admin`
- Password: `admin`

**RabbitMQ:**
- Username: `admin`
- Password: `SecureP@ssw0rd2024!`

**PostgreSQL:**
- Username: `postgres`
- Password: `postgres`

---

## Getting Help

- **Service Logs:** `/tmp/*-backend.log`, `/tmp/*-frontend.log`
- **Documentation:** See [docs/INDEX.md](../docs/INDEX.md)
- **Service Catalog:** [SERVICE_CATALOG.md](../SERVICE_CATALOG.md)
- **CLAUDE.md:** [CLAUDE.md](../CLAUDE.md) - Developer guide for AI

---

**Quick Start Complete!** 🎉

Your Beakon development environment is now ready. Happy coding!
