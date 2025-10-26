# Docker Deployment Structure

This directory contains standalone deployment configurations for all 20 Beakon microservices.

## Directory Structure

```
docker-deployment/
├── api-gateway/
│   ├── docker-compose.yml      # Standalone deployment for this service
│   ├── .env                    # Environment variable overrides
│   └── configs/
│       ├── config.yml          # YAML configuration (lowest priority)
│       └── service-endpoints.yml  # Static service endpoint URLs
├── user-service/
│   ├── docker-compose.yml
│   ├── .env
│   └── configs/
│       ├── config.yml
│       └── service-endpoints.yml
... (repeated for all 20 services)
```

## Services Included (20 total)

### Backend Services (14)
- api-gateway (Port 8080)
- user-service (Port 8081)
- tenant-admin-service (Port 8099)
- saas-admin-service (Port 8098)
- component-service (Port 8084)
- notification-service (Port 8085)
- incident-service (Port 8086)
- payment-service (Port 8088)
- analytics-service (Port 8090)
- monitoring-service (Port 8092)
- status-ui-service (Port 8093)
- event-store-service (Port 8096)
- branding-service (Port 8097)
- landing-page-service (Port 8100)

### Consumer Services (4)
- analytics-consumer
- notification-consumer
- audit-consumer
- billing-consumer

### Frontend Services (2)
- saas-admin-frontend (Port 3001)
- tenant-admin-frontend (Port 3002)

## Configuration Precedence

The configuration system follows this precedence order (lowest to highest):

**config.yml < .env < environment variables**

1. **config.yml** (Lowest priority)
   - YAML-based configuration
   - Located in `configs/config.yml`
   - Provides default values for all settings
   - Can be customized per deployment environment

2. **.env file** (Medium priority)
   - Environment variable definitions
   - Located in service root directory
   - Overrides values from config.yml
   - Not committed to git (use .env.example as template)

3. **Environment Variables** (Highest priority)
   - Set in docker-compose.yml or shell
   - Final override for any configuration value
   - Used for sensitive data (secrets, passwords)

## Configuration Files

### config.yml

YAML-based configuration with all service settings:

```yaml
environment: development

server:
  port: 8081
  host: "0.0.0.0"
  read_timeout: 30s
  write_timeout: 30s

database:
  host: statuspage-postgres
  port: 5432
  user: postgres
  password: postgres
  name: statuspage_user
  sslmode: disable

jwt:
  secret: development-secret-key-min-32-chars-for-testing
  expiration: 24h

redis:
  host: redis
  port: 6379
  enabled: false

rabbitmq:
  host: rabbitmq
  port: 5672
  user: guest
  password: guest
```

### service-endpoints.yml

Static service endpoint URLs for inter-service communication:

```yaml
services:
  user_service:
    url: "http://user-service:8081"
    health: "/health"

  tenant_admin_service:
    url: "http://tenant-admin-service:8099"
    health: "/health"

infrastructure:
  postgres:
    host: "statuspage-postgres"
    port: 5432

  redis:
    host: "redis"
    port: 6379
```

## Usage

### Deploy a Single Service

```bash
cd docker-deployment/user-service
docker-compose up -d
```

### Stop a Service

```bash
cd docker-deployment/user-service
docker-compose down
```

### View Logs

```bash
cd docker-deployment/user-service
docker-compose logs -f
```

### Override Configuration

#### Method 1: Edit .env file

```bash
cd docker-deployment/user-service
vim .env

# Change values
DB_HOST=custom-postgres-host
JWT_SECRET=my-production-secret
```

#### Method 2: Set environment variables

```bash
cd docker-deployment/user-service
DB_HOST=custom-host docker-compose up -d
```

#### Method 3: Edit config.yml (not recommended for secrets)

```bash
cd docker-deployment/user-service/configs
vim config.yml

# Edit YAML values
database:
  host: custom-postgres-host
```

### Deploy Multiple Services

To deploy related services together:

```bash
# Create a custom docker-compose.yml that references multiple services
docker-compose -f docker-deployment/user-service/docker-compose.yml \
               -f docker-deployment/tenant-admin-service/docker-compose.yml \
               up -d
```

Or use the root docker-compose.yml for full stack deployment.

## Best Practices

### 1. Configuration Management

- ✅ **DO**: Use config.yml for non-sensitive default values
- ✅ **DO**: Use .env for environment-specific overrides
- ✅ **DO**: Use environment variables for secrets and passwords
- ❌ **DON'T**: Commit .env files with real credentials to git
- ❌ **DON'T**: Store secrets in config.yml files

### 2. Service Isolation

- Each service has its own docker-compose.yml for standalone deployment
- Services share a common `beakon-network` for inter-service communication
- Create the network first: `docker network create beakon-network`

### 3. Database Configuration

- All backend and consumer services connect to `statuspage-postgres`
- Each service has its own database (database-per-service pattern)
- Connection pooling is configured in config.yml

### 4. Secrets Management

For production deployments:

```bash
# Use Docker secrets
echo "my-jwt-secret" | docker secret create jwt_secret -

# Reference in docker-compose.yml
secrets:
  jwt_secret:
    external: true
```

Or use environment variables from a secure vault:

```bash
# From AWS Secrets Manager, HashiCorp Vault, etc.
export JWT_SECRET=$(aws secretsmanager get-secret-value ...)
docker-compose up -d
```

## Network Requirements

All services require the `beakon-network` Docker network:

```bash
docker network create beakon-network
```

This network allows services to communicate using their service names as hostnames.

## Infrastructure Dependencies

Most services depend on:

- **PostgreSQL**: statuspage-postgres (Port 5432)
- **Redis**: redis (Port 6379)
- **RabbitMQ**: rabbitmq (Ports 5672, 15672)

Ensure these are running before starting services:

```bash
# Check infrastructure
docker ps | grep -E "postgres|redis|rabbitmq"
```

## Health Checks

All backend services expose health endpoints:

- `GET /health` - Overall health status
- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness probe

Example:

```bash
curl http://localhost:8081/health
```

## Metrics

Backend services expose Prometheus metrics:

- Metrics port: Service port + 1010
- Example: user-service (8081) → metrics on 9091
- Path: `/metrics`

```bash
curl http://localhost:9091/metrics
```

## Environment-Specific Deployment

### Development

```bash
cd docker-deployment/user-service
ENVIRONMENT=development docker-compose up -d
```

### Staging

```bash
cd docker-deployment/user-service
ENVIRONMENT=staging \
DB_HOST=staging-postgres.example.com \
JWT_SECRET=staging-secret \
docker-compose up -d
```

### Production

```bash
cd docker-deployment/user-service
ENVIRONMENT=production \
DB_HOST=prod-postgres.example.com \
DB_SSLMODE=require \
JWT_SECRET=$PROD_JWT_SECRET \
REDIS_ENABLED=true \
docker-compose up -d
```

## Troubleshooting

### Service won't start

```bash
# Check logs
docker-compose logs <service-name>

# Verify configuration
docker-compose config

# Check environment variables
docker-compose exec <service> env
```

### Configuration not loading

```bash
# Verify config file is mounted
docker-compose exec <service> ls -la /app/configs/

# Check Viper is loading config
docker-compose logs <service> | grep "Loaded config"
```

### Network connectivity issues

```bash
# Verify network exists
docker network ls | grep beakon-network

# Check service can reach postgres
docker-compose exec <service> ping statuspage-postgres

# Check DNS resolution
docker-compose exec <service> nslookup statuspage-postgres
```

## Migration from Root docker-compose.yml

To migrate from the root docker-compose.yml to this structure:

1. **Stop all services**:
   ```bash
   docker-compose down
   ```

2. **Create network** (if not exists):
   ```bash
   docker network create beakon-network
   ```

3. **Deploy individual services**:
   ```bash
   cd docker-deployment/user-service
   docker-compose up -d
   ```

4. **Verify health**:
   ```bash
   curl http://localhost:8081/health
   ```

## Adding a New Service

To add a new service to this structure:

1. Create service directory:
   ```bash
   mkdir -p docker-deployment/new-service/configs
   ```

2. Generate config files using the script:
   ```bash
   cd scripts
   ./generate-all-configs.sh
   ```

3. Generate docker-compose.yml:
   ```bash
   ./generate-docker-deployment-files.sh
   ```

4. Test deployment:
   ```bash
   cd docker-deployment/new-service
   docker-compose up -d
   ```

## Support

For issues or questions:
- Check service-specific README: `microservices/<service>/README.md`
- Review logs: `docker-compose logs -f`
- Verify environment: `docker-compose config`
- Check infrastructure: `docker ps | grep -E "postgres|redis|rabbitmq"`

---

**Last Updated**: October 2025
**Maintainer**: Beakon Platform Team
