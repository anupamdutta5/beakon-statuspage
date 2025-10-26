#!/bin/bash
#
# Generate docker-compose.yml and .env for ALL 20 microservices
#

set -e

cd /Users/anuoamdutta/Desktop/statuspage/Beakon

GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}Generating Docker Deployment Files${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# Service configuration
SERVICE_CONFIG=(
    "api-gateway:8080::backend"
    "user-service:8081:statuspage_user:backend"
    "tenant-admin-service:8099:tenant_admin_db:backend"
    "saas-admin-service:8098:saas_admin:backend"
    "component-service:8084:statuspage_component:backend"
    "notification-service:8085:statuspage_notification:backend"
    "incident-service:8086:statuspage_incident:backend"
    "payment-service:8088:statuspage_payment:backend"
    "analytics-service:8090:statuspage_analytics:backend"
    "monitoring-service:8092:statuspage_monitoring:backend"
    "status-ui-service:8093:statuspage_status:backend"
    "event-store-service:8096:statuspage_events:backend"
    "branding-service:8097:statuspage_branding:backend"
    "landing-page-service:8100:statuspage_landing:backend"
    "analytics-consumer:0:statuspage_analytics:consumer"
    "notification-consumer:0:statuspage_notification:consumer"
    "audit-consumer:0:statuspage_audit:consumer"
    "billing-consumer:0:statuspage_billing:consumer"
    "saas-admin-frontend:3001::frontend"
    "tenant-admin-frontend:3002::frontend"
)

generate_backend_docker_compose() {
    local service=$1
    local port=$2
    local database=$3

    cat > "docker-deployment/$service/docker-compose.yml" << EOF
version: '3.8'

services:
  $service:
    image: beakon-${service}:latest
    container_name: beakon-${service}
    build:
      context: ../../microservices/${service}
      dockerfile: Dockerfile
    ports:
      - "${port}:${port}"
      - "$((port + 1010)):$((port + 1010))"
    volumes:
      - ./configs:/app/configs:ro
    environment:
      - ENVIRONMENT=\${ENVIRONMENT:-development}
      - SERVER_PORT=${port}
      - DB_HOST=\${DB_HOST:-statuspage-postgres}
      - DB_PORT=\${DB_PORT:-5432}
      - DB_USER=\${DB_USER:-postgres}
      - DB_PASSWORD=\${DB_PASSWORD:-postgres}
      - DB_NAME=${database}
      - DB_SSLMODE=\${DB_SSLMODE:-disable}
      - JWT_SECRET=\${JWT_SECRET}
      - REDIS_HOST=\${REDIS_HOST:-redis}
      - REDIS_PORT=\${REDIS_PORT:-6379}
      - REDIS_ENABLED=\${REDIS_ENABLED:-false}
      - RABBITMQ_HOST=\${RABBITMQ_HOST:-rabbitmq}
      - RABBITMQ_PORT=\${RABBITMQ_PORT:-5672}
      - RABBITMQ_USER=\${RABBITMQ_USER:-guest}
      - RABBITMQ_PASSWORD=\${RABBITMQ_PASSWORD:-guest}
      - LOG_LEVEL=\${LOG_LEVEL:-info}
    networks:
      - beakon-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:${port}/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

networks:
  beakon-network:
    name: beakon-network
    external: true
EOF
}

generate_consumer_docker_compose() {
    local service=$1
    local database=$2

    cat > "docker-deployment/$service/docker-compose.yml" << EOF
version: '3.8'

services:
  $service:
    image: beakon-${service}:latest
    container_name: beakon-${service}
    build:
      context: ../../microservices/${service}
      dockerfile: Dockerfile
    volumes:
      - ./configs:/app/configs:ro
    environment:
      - ENVIRONMENT=\${ENVIRONMENT:-development}
      - DB_HOST=\${DB_HOST:-statuspage-postgres}
      - DB_PORT=\${DB_PORT:-5432}
      - DB_USER=\${DB_USER:-postgres}
      - DB_PASSWORD=\${DB_PASSWORD:-postgres}
      - DB_NAME=${database}
      - DB_SSLMODE=\${DB_SSLMODE:-disable}
      - RABBITMQ_HOST=\${RABBITMQ_HOST:-rabbitmq}
      - RABBITMQ_PORT=\${RABBITMQ_PORT:-5672}
      - RABBITMQ_USER=\${RABBITMQ_USER:-guest}
      - RABBITMQ_PASSWORD=\${RABBITMQ_PASSWORD:-guest}
      - LOG_LEVEL=\${LOG_LEVEL:-info}
      - WORKER_CONCURRENCY=\${WORKER_CONCURRENCY:-5}
    networks:
      - beakon-network
    restart: unless-stopped

networks:
  beakon-network:
    name: beakon-network
    external: true
EOF
}

generate_frontend_docker_compose() {
    local service=$1
    local port=$2

    local api_url_var
    if [[ "$service" == "saas-admin-frontend" ]]; then
        api_url_var="NEXT_PUBLIC_API_URL:-http://localhost:8098"
    else
        api_url_var="NEXT_PUBLIC_API_URL:-http://localhost:8099"
    fi

    cat > "docker-deployment/$service/docker-compose.yml" << EOF
version: '3.8'

services:
  $service:
    image: beakon-${service}:latest
    container_name: beakon-${service}
    build:
      context: ../../microservices/${service}
      dockerfile: Dockerfile
    ports:
      - "${port}:${port}"
    volumes:
      - ./configs:/app/configs:ro
    environment:
      - NEXT_PUBLIC_API_URL=\${${api_url_var}}
      - NODE_ENV=\${NODE_ENV:-development}
      - PORT=${port}
    networks:
      - beakon-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:${port}"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

networks:
  beakon-network:
    name: beakon-network
    external: true
EOF
}

generate_env_file() {
    local service=$1
    local type=$2

    cat > "docker-deployment/$service/.env" << 'EOF'
# Environment configuration for $service
# Override values from configs/config.yml

ENVIRONMENT=development

# Database
DB_HOST=statuspage-postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable

# JWT (Backend services only)
JWT_SECRET=development-secret-key-min-32-chars-for-testing

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_ENABLED=false

# RabbitMQ
RABBITMQ_HOST=rabbitmq
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest

# Logging
LOG_LEVEL=info

# Worker (Consumers only)
WORKER_CONCURRENCY=5
EOF
}

CREATED=0

for config_line in "${SERVICE_CONFIG[@]}"; do
    IFS=':' read -r service port database type <<< "$config_line"

    echo -e "${BLUE}Generating deployment files for $service ($type)...${NC}"

    # Generate docker-compose.yml
    case $type in
        backend)
            generate_backend_docker_compose "$service" "$port" "$database"
            ;;
        consumer)
            generate_consumer_docker_compose "$service" "$database"
            ;;
        frontend)
            generate_frontend_docker_compose "$service" "$port"
            ;;
    esac

    # Generate .env
    generate_env_file "$service" "$type"

    CREATED=$((CREATED + 1))
    echo -e "  ${GREEN}✓${NC} docker-compose.yml"
    echo -e "  ${GREEN}✓${NC} .env"
    echo ""
done

echo ""
echo -e "${BLUE}=========================================${NC}"
echo -e "${GREEN}Summary${NC}"
echo -e "${BLUE}=========================================${NC}"
echo "Services configured: $CREATED"
echo "Files generated: $((CREATED * 2))"
echo ""
echo -e "${GREEN}✅ All deployment files generated successfully!${NC}"
echo ""
echo "Structure created:"
echo "  docker-deployment/"
echo "    ├── <service>/"
echo "    │   ├── docker-compose.yml (standalone)"
echo "    │   ├── .env (environment overrides)"
echo "    │   └── configs/"
echo "    │       ├── config.yml (YAML config)"
echo "    │       └── service-endpoints.yml (service URLs)"
echo ""
echo "Next steps:"
echo "  1. Update Go services to use Viper for config loading"
echo "  2. Update Dockerfiles to copy configs directory"
echo "  3. Test individual service deployments"
