# Docker Deployment - Quick Start Guide

## Prerequisites

```bash
# Create Docker network (required for all services)
docker network create beakon-network

# Ensure infrastructure is running
docker ps | grep -E "statuspage-postgres|redis|rabbitmq"
```

## Deploy a Single Service

```bash
# Example: Deploy user-service
cd docker-deployment/user-service

# Start the service
docker-compose up -d

# Check logs
docker-compose logs -f

# Check health
curl http://localhost:8081/health

# Stop the service
docker-compose down
```

## Deploy Multiple Services

```bash
# Start infrastructure first
docker-compose up -d statuspage-postgres redis rabbitmq

# Deploy core services
cd docker-deployment/user-service && docker-compose up -d && cd ../..
cd docker-deployment/tenant-admin-service && docker-compose up -d && cd ../..
cd docker-deployment/saas-admin-service && docker-compose up -d && cd ../..
```

## Configuration Override

### Quick Override (.env file)

```bash
cd docker-deployment/user-service
vim .env

# Edit values
DB_HOST=production-postgres
JWT_SECRET=my-production-secret
REDIS_ENABLED=true

# Restart service
docker-compose restart
```

### Runtime Override (Environment Variables)

```bash
cd docker-deployment/user-service

# Set environment variables
DB_HOST=staging-db.example.com \
JWT_SECRET=$STAGING_SECRET \
docker-compose up -d
```

## Configuration Precedence

```
Environment Variables  (Highest Priority)
         ↓
      .env file
         ↓
    config.yml         (Lowest Priority)
```

## Service Ports

**Backend Services**:
- api-gateway: 8080
- user-service: 8081
- tenant-admin-service: 8099
- saas-admin-service: 8098
- component-service: 8084
- notification-service: 8085
- incident-service: 8086
- payment-service: 8088
- analytics-service: 8090
- monitoring-service: 8092
- status-ui-service: 8093
- event-store-service: 8096
- branding-service: 8097
- landing-page-service: 8100

**Frontend Services**:
- saas-admin-frontend: 3001
- tenant-admin-frontend: 3002

**Metrics Ports**: Service Port + 1010
- Example: user-service (8081) → metrics on 9091

## Common Commands

```bash
# Start service
docker-compose up -d

# View logs
docker-compose logs -f

# Restart service
docker-compose restart

# Stop service
docker-compose down

# Rebuild and restart
docker-compose up -d --build

# Check service status
docker-compose ps

# View environment variables
docker-compose config
```

## Health Checks

All backend services expose:
- `/health` - Overall health
- `/health/live` - Liveness probe
- `/health/ready` - Readiness probe

```bash
# Check health
curl http://localhost:8081/health

# Check all services
for port in 8080 8081 8084 8085 8086 8088 8090 8092 8093 8096 8097 8098 8099 8100; do
    echo "Port $port: $(curl -s http://localhost:$port/health | jq -r '.status' 2>/dev/null || echo 'N/A')"
done
```

## Troubleshooting

### Service won't start

```bash
# Check logs
docker-compose logs <service>

# Check if network exists
docker network ls | grep beakon-network

# Check if ports are available
lsof -i :<port>
```

### Can't connect to database

```bash
# Verify PostgreSQL is running
docker ps | grep postgres

# Test connection
docker-compose exec <service> ping statuspage-postgres
```

### Configuration not loading

```bash
# Verify configs are mounted
docker-compose exec <service> ls -la /app/configs/

# Check environment variables
docker-compose exec <service> env | grep DB_
```

## Directory Structure

```
docker-deployment/
├── <service-name>/
│   ├── docker-compose.yml    # Standalone deployment
│   ├── .env                  # Environment overrides
│   └── configs/
│       ├── config.yml        # YAML configuration
│       └── service-endpoints.yml  # Service URLs
```

## Important Files

- **[README.md](README.md)** - Complete documentation
- **[IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md)** - Implementation details
- **[QUICK_START.md](QUICK_START.md)** - This file

## Need Help?

- Check [README.md](README.md) for detailed documentation
- Check service logs: `docker-compose logs -f`
- Verify configuration: `docker-compose config`
- Test connectivity: `docker-compose exec <service> ping <target>`

---

**Pro Tips**:
- Always create `beakon-network` first
- Use `.env` for development, environment variables for production
- Never commit `.env` files with real credentials
- Health checks are your friend - use them!
