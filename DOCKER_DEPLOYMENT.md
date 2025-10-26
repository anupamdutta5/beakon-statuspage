# Beakon Platform - Docker Deployment Guide

Complete guide for deploying the Beakon status page platform using Docker and Docker Compose.

## Table of Contents
- [Quick Start](#quick-start)
- [Architecture Overview](#architecture-overview)
- [Prerequisites](#prerequisites)
- [Configuration](#configuration)
- [Deployment Options](#deployment-options)
- [Service Management](#service-management)
- [Troubleshooting](#troubleshooting)
- [Production Deployment](#production-deployment)

---

## Quick Start

**1. Clone and setup:**
```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon
cp .env.example .env
# Edit .env with your configuration
```

**2. Start infrastructure only (PostgreSQL, Redis, RabbitMQ):**
```bash
docker-compose -f docker-compose.infrastructure.yml up -d
```

**3. Start all services:**
```bash
docker-compose up -d
```

**4. Verify deployment:**
```bash
# Check all services are running
docker-compose ps

# Check API Gateway health
curl http://localhost:8080/health

# Access frontends
open http://localhost:3001  # SaaS Admin Frontend
open http://localhost:3002  # Tenant Admin Frontend
```

---

## Architecture Overview

### Service Distribution

**Infrastructure (3 services):**
- PostgreSQL (port 5432) - Primary database
- Redis (port 6379) - Caching & sessions
- RabbitMQ (ports 5672, 15672) - Message broker

**Backend Services (15 services):**
- API Gateway (8080) - Entry point
- User Service (8081)
- Component Service (8084)
- Notification Service (8085)
- Incident Service (8086)
- Payment Service (8088)
- Analytics Service (8090)
- Monitoring Service (8092)
- Status UI Service (8093)
- Event Store Service (8096)
- Branding Service (8097)
- SaaS Admin Service (8098)
- Tenant Admin Service (8099)
- Landing Page Service (8100)

**Frontend Services (2 services):**
- SaaS Admin Frontend (3001) - Next.js
- Tenant Admin Frontend (3002) - Next.js

**Consumer Services (4 background workers):**
- Analytics Consumer (no port)
- Notification Consumer (no port)
- Audit Consumer (no port)
- Billing Consumer (no port)

**Total:** 24 containerized services

### Port Allocation

| Service Type | Port Range | Metrics Ports |
|--------------|------------|---------------|
| Frontends | 3001-3002 | N/A |
| Backend HTTP | 8080-8100 | 9090-9110 |
| PostgreSQL | 5432 | N/A |
| Redis | 6379 | N/A |
| RabbitMQ | 5672, 15672 | N/A |

---

## Prerequisites

**Required:**
- Docker 20.10+ (`docker --version`)
- Docker Compose 2.0+ (`docker-compose --version`)
- At least 8GB RAM available for Docker
- 20GB free disk space

**Recommended:**
- Docker Desktop 4.0+ (for macOS/Windows)
- 16GB RAM total system memory
- SSD storage

**Install Docker (macOS):**
```bash
brew install --cask docker
# Or download from https://www.docker.com/products/docker-desktop
```

---

## Configuration

### Environment Variables

**1. Copy example configuration:**
```bash
cp .env.example .env
```

**2. Configure required variables:**

**Minimum required for development:**
```bash
# .env
ENVIRONMENT=development
JWT_SECRET=your-super-secret-jwt-key-min-32-chars-long
DB_HOST=postgres
DB_USER=postgres
DB_PASSWORD=postgres
```

**Production configuration:**
```bash
# .env
ENVIRONMENT=production
JWT_SECRET=<generate-strong-random-32+character-secret>
DB_HOST=postgres
DB_USER=beakon_user
DB_PASSWORD=<generate-strong-password>
DB_SSLMODE=require

# Email (required for notifications)
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASSWORD=<sendgrid-api-key>

# Payment (required for billing)
STRIPE_SECRET_KEY=sk_live_...
STRIPE_PUBLISHABLE_KEY=pk_live_...

# SMS (optional)
TWILIO_ACCOUNT_SID=AC...
TWILIO_AUTH_TOKEN=...
```

**3. Validate configuration:**
```bash
# Check .env file
cat .env | grep -v '^#' | grep -v '^$'
```

---

## Deployment Options

### Option 1: Infrastructure Only

Start only PostgreSQL, Redis, and RabbitMQ:

```bash
docker-compose -f docker-compose.infrastructure.yml up -d
```

**Use case:** When you want to run services manually for development

### Option 2: All Services

Start everything (infrastructure + all microservices):

```bash
docker-compose up -d
```

**Use case:** Full platform deployment

### Option 3: Specific Services

Start infrastructure + selected services:

```bash
# Start infrastructure first
docker-compose -f docker-compose.infrastructure.yml up -d

# Start specific services
docker-compose up -d api-gateway user-service tenant-admin-service
```

### Option 4: Individual Service

Build and run a single service:

```bash
cd microservices/tenant-admin-service
docker-compose -f build/docker-compose.yml up -d
```

**Note:** Ensure infrastructure is running first!

---

## Service Management

### Starting Services

**Start all services:**
```bash
docker-compose up -d
```

**Start with logs:**
```bash
docker-compose up
```

**Start specific service:**
```bash
docker-compose up -d tenant-admin-service
```

**Rebuild and start:**
```bash
docker-compose up -d --build
```

### Stopping Services

**Stop all:**
```bash
docker-compose down
```

**Stop without removing:**
```bash
docker-compose stop
```

**Stop specific service:**
```bash
docker-compose stop tenant-admin-service
```

**Stop and remove volumes (⚠️ deletes data):**
```bash
docker-compose down -v
```

### Viewing Logs

**All services:**
```bash
docker-compose logs -f
```

**Specific service:**
```bash
docker-compose logs -f tenant-admin-service
```

**Last 100 lines:**
```bash
docker-compose logs --tail=100 tenant-admin-service
```

**Since timestamp:**
```bash
docker-compose logs --since 2024-10-25T10:00:00 tenant-admin-service
```

### Health Checks

**Check all container status:**
```bash
docker-compose ps
```

**Check specific service health:**
```bash
curl http://localhost:8099/health
curl http://localhost:8080/health
```

**Check all backend services:**
```bash
for port in 8080 8081 8084 8085 8086 8088 8090 8092 8093 8096 8097 8098 8099 8100; do
  echo "Port $port:"
  curl -s http://localhost:$port/health | jq '.status' || echo "Failed"
done
```

### Scaling Services

**Scale a service:**
```bash
docker-compose up -d --scale tenant-admin-service=3
```

**Note:** Update nginx/load balancer for proper distribution

---

## Database Management

### Initialize Databases

**Option 1: Automatic (recommended):**

Each service auto-migrates its database on startup.

**Option 2: Manual initialization:**

```bash
# Run initialization script
docker exec -it beakon-postgres psql -U postgres -f /docker-entrypoint-initdb.d/init.sql

# Or initialize specific database
docker exec -it beakon-postgres createdb -U postgres tenant_admin_db
```

### Database Backup

**Backup all databases:**
```bash
docker exec beakon-postgres pg_dumpall -U postgres > backup_$(date +%Y%m%d).sql
```

**Backup specific database:**
```bash
docker exec beakon-postgres pg_dump -U postgres tenant_admin_db > tenant_admin_backup.sql
```

### Database Restore

```bash
# Restore all databases
docker exec -i beakon-postgres psql -U postgres < backup_20251025.sql

# Restore specific database
docker exec -i beakon-postgres psql -U postgres -d tenant_admin_db < tenant_admin_backup.sql
```

### Access Database

```bash
# PostgreSQL CLI
docker exec -it beakon-postgres psql -U postgres

# List databases
docker exec -it beakon-postgres psql -U postgres -c '\l'

# Connect to specific database
docker exec -it beakon-postgres psql -U postgres -d tenant_admin_db
```

---

## Troubleshooting

### Common Issues

**1. Port already in use:**
```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>

# Or change port in docker-compose.yml
ports:
  - "8081:8080"  # Map host 8081 to container 8080
```

**2. Out of memory:**
```bash
# Check Docker memory
docker system df

# Prune unused resources
docker system prune -a

# Increase Docker memory (Docker Desktop → Settings → Resources)
```

**3. Service won't start:**
```bash
# Check logs
docker-compose logs service-name

# Check container status
docker-compose ps service-name

# Inspect container
docker inspect beakon-service-name

# Restart service
docker-compose restart service-name
```

**4. Database connection failed:**
```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check PostgreSQL logs
docker-compose logs postgres

# Test connection
docker exec -it beakon-postgres pg_isready -U postgres
```

**5. Build failures:**
```bash
# Clean build cache
docker-compose build --no-cache service-name

# Remove old images
docker rmi $(docker images -f "dangling=true" -q)

# Full rebuild
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### Debugging

**Enter container shell:**
```bash
docker exec -it beakon-tenant-admin-service sh
```

**Check environment variables:**
```bash
docker exec beakon-tenant-admin-service env
```

**Check network connectivity:**
```bash
docker exec beakon-api-gateway ping tenant-admin-service
```

**Inspect logs in real-time:**
```bash
docker-compose logs -f --tail=0
```

---

## Production Deployment

### Production Checklist

- [ ] Generate strong JWT_SECRET (32+ characters)
- [ ] Set strong database passwords
- [ ] Enable SSL/TLS for PostgreSQL (DB_SSLMODE=require)
- [ ] Configure SMTP for email notifications
- [ ] Configure Stripe for payments
- [ ] Set ENVIRONMENT=production
- [ ] Configure proper backup strategy
- [ ] Set up monitoring (Prometheus/Grafana)
- [ ] Configure log aggregation
- [ ] Set resource limits for containers
- [ ] Use secrets management (Docker Secrets/Vault)
- [ ] Enable container restart policies
- [ ] Configure reverse proxy (Nginx/Traefik)
- [ ] Set up SSL certificates (Let's Encrypt)

### Resource Limits

Add to each service in docker-compose.yml:

```yaml
services:
  tenant-admin-service:
    # ... existing config ...
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

### Secrets Management

**Using Docker Secrets:**

```bash
# Create secrets
echo "my-jwt-secret" | docker secret create jwt_secret -
echo "db-password" | docker secret create db_password -

# Update docker-compose.yml
secrets:
  jwt_secret:
    external: true
  db_password:
    external: true
```

### Reverse Proxy (Nginx)

```nginx
# /etc/nginx/conf.d/beakon.conf
upstream api_gateway {
    server localhost:8080;
}

server {
    listen 80;
    server_name api.beakon.com;

    location / {
        proxy_pass http://api_gateway;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## Monitoring

### Prometheus Metrics

All services expose metrics on port + 1010:

```bash
# Scrape metrics
curl http://localhost:9099/metrics  # Tenant Admin
curl http://localhost:9090/metrics  # API Gateway
```

### Health Dashboard

Create a simple monitoring script:

```bash
#!/bin/bash
# health-check.sh

services=(8080 8081 8084 8085 8086 8088 8090 8092 8093 8096 8097 8098 8099 8100)

for port in "${services[@]}"; do
  status=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$port/health)
  if [ "$status" == "200" ]; then
    echo "✅ Port $port: Healthy"
  else
    echo "❌ Port $port: Unhealthy ($status)"
  fi
done
```

---

## Maintenance

### Update Services

```bash
# Pull latest code
git pull origin main

# Rebuild and restart
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### Clean Up

```bash
# Remove stopped containers
docker-compose down

# Remove unused images
docker image prune -a

# Remove unused volumes (⚠️ careful!)
docker volume prune

# Complete cleanup
docker system prune -a --volumes
```

---

## Quick Reference

**Common Commands:**
```bash
# Start everything
docker-compose up -d

# Stop everything
docker-compose down

# View logs
docker-compose logs -f service-name

# Restart service
docker-compose restart service-name

# Rebuild service
docker-compose build --no-cache service-name

# Scale service
docker-compose up -d --scale service-name=3

# Check status
docker-compose ps

# Execute command in container
docker exec -it container-name command
```

**Service URLs:**
- API Gateway: http://localhost:8080
- SaaS Admin Frontend: http://localhost:3001
- Tenant Admin Frontend: http://localhost:3002
- RabbitMQ Management: http://localhost:15672 (guest/guest)

---

## Support

For issues or questions:
1. Check service logs: `docker-compose logs service-name`
2. Review [OPERATIONAL_RUNBOOK.md](OPERATIONAL_RUNBOOK.md)
3. Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md)

**Document Version:** 1.0
**Last Updated:** 2025-10-25
**Maintained By:** DevOps Team
