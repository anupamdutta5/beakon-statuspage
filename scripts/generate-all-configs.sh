#!/bin/bash
#
# Generate config.yml and service-endpoints.yml for ALL 20 microservices
#

set -e

cd /Users/anuoamdutta/Desktop/statuspage/Beakon

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}Generating Config Files for All Services${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# Service configuration
# Format: service_name:port:database_name:type (backend/frontend/consumer)
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

generate_service_endpoints_yml() {
    cat > "$1" << 'EOF'
# Service endpoint configuration
# These are static URLs that services use to communicate with each other
# Override via environment variables if needed (e.g., SERVICE_USER_URL)

services:
  api_gateway:
    url: "http://api-gateway:8080"
    health: "/health"

  user_service:
    url: "http://user-service:8081"
    health: "/health"

  tenant_admin_service:
    url: "http://tenant-admin-service:8099"
    health: "/health"

  saas_admin_service:
    url: "http://saas-admin-service:8098"
    health: "/api/v1/health"

  component_service:
    url: "http://component-service:8084"
    health: "/health"

  notification_service:
    url: "http://notification-service:8085"
    health: "/health"

  incident_service:
    url: "http://incident-service:8086"
    health: "/health"

  payment_service:
    url: "http://payment-service:8088"
    health: "/health"

  analytics_service:
    url: "http://analytics-service:8090"
    health: "/health"

  monitoring_service:
    url: "http://monitoring-service:8092"
    health: "/health"

  status_ui_service:
    url: "http://status-ui-service:8093"
    health: "/health"

  event_store_service:
    url: "http://event-store-service:8096"
    health: "/health"

  branding_service:
    url: "http://branding-service:8097"
    health: "/health"

  landing_page_service:
    url: "http://landing-page-service:8100"
    health: "/health"

infrastructure:
  postgres:
    host: "statuspage-postgres"
    port: 5432

  redis:
    host: "redis"
    port: 6379

  rabbitmq:
    host: "rabbitmq"
    port: 5672
    management_port: 15672
EOF
}

generate_backend_config_yml() {
    local service=$1
    local port=$2
    local database=$3

    cat > "docker-deployment/$service/configs/config.yml" << EOF
# Configuration for $service
# Precedence: config.yml < .env < environment variables

environment: development

server:
  port: ${port}
  read_timeout: 30s
  write_timeout: 30s
  shutdown_timeout: 30s
  host: "0.0.0.0"

database:
  host: statuspage-postgres
  port: 5432
  user: postgres
  password: postgres
  name: ${database}
  sslmode: disable
  max_open_conns: 80
  max_idle_conns: 32
  conn_max_lifetime: 1h
  conn_max_idle_time: 10m

jwt:
  secret: development-secret-key-min-32-chars-for-testing
  expiration: 24h

redis:
  host: redis
  port: 6379
  password: ""
  enabled: false
  db: 0
  pool_size: 10
  min_idle_conns: 5

rabbitmq:
  host: rabbitmq
  port: 5672
  user: guest
  password: guest
  vhost: /
  exchange: beakon_events
  queue_prefix: ${service}

circuit_breaker:
  database_enabled: true
  database_failure_ratio: 0.6
  database_timeout: 5s
  database_max_requests: 3

rate_limiting:
  enabled: true
  requests_per_minute: 100
  burst: 50

logging:
  level: info
  format: json
  output: stdout

monitoring:
  enabled: true
  check_interval: 60
  timeout: 30

metrics:
  enabled: true
  port: $((port + 1010))
  path: /metrics
EOF
}

generate_consumer_config_yml() {
    local service=$1
    local database=$2

    cat > "docker-deployment/$service/configs/config.yml" << EOF
# Configuration for $service (Consumer)
# Precedence: config.yml < .env < environment variables

environment: development

database:
  host: statuspage-postgres
  port: 5432
  user: postgres
  password: postgres
  name: ${database}
  sslmode: disable
  max_open_conns: 80
  max_idle_conns: 32
  conn_max_lifetime: 1h
  conn_max_idle_time: 10m

rabbitmq:
  host: rabbitmq
  port: 5672
  user: guest
  password: guest
  vhost: /
  exchange: beakon_events
  queue: ${service}_queue
  consumer_tag: ${service}
  prefetch_count: 10
  auto_ack: false

worker:
  concurrency: 5
  prefetch_count: 10
  retry_attempts: 3
  retry_delay: 5s

circuit_breaker:
  database_enabled: true
  database_failure_ratio: 0.6
  database_timeout: 5s

logging:
  level: info
  format: json
  output: stdout

metrics:
  enabled: true
  port: 9100
  path: /metrics
EOF
}

generate_frontend_config_yml() {
    local service=$1
    local port=$2

    local api_url
    if [[ "$service" == "saas-admin-frontend" ]]; then
        api_url="http://localhost:8098"
    else
        api_url="http://localhost:8099"
    fi

    cat > "docker-deployment/$service/configs/config.yml" << EOF
# Configuration for $service (Frontend)
# Precedence: config.yml < .env < environment variables

environment: development

server:
  port: ${port}
  host: "0.0.0.0"

api:
  url: ${api_url}
  timeout: 30s

logging:
  level: info

features:
  analytics_enabled: true
  debug_mode: false
EOF
}

CREATED=0

for config_line in "${SERVICE_CONFIG[@]}"; do
    IFS=':' read -r service port database type <<< "$config_line"

    echo -e "${BLUE}Generating configs for $service ($type)...${NC}"

    # Generate service-endpoints.yml (same for all services)
    generate_service_endpoints_yml "docker-deployment/$service/configs/service-endpoints.yml"

    # Generate service-specific config.yml
    case $type in
        backend)
            generate_backend_config_yml "$service" "$port" "$database"
            ;;
        consumer)
            generate_consumer_config_yml "$service" "$database"
            ;;
        frontend)
            generate_frontend_config_yml "$service" "$port"
            ;;
    esac

    CREATED=$((CREATED + 1))
    echo -e "  ${GREEN}✓${NC} config.yml"
    echo -e "  ${GREEN}✓${NC} service-endpoints.yml"
    echo ""
done

echo ""
echo -e "${BLUE}=========================================${NC}"
echo -e "${GREEN}Summary${NC}"
echo -e "${BLUE}=========================================${NC}"
echo "Services configured: $CREATED"
echo "Files generated: $((CREATED * 2))"
echo ""
echo -e "${GREEN}✅ All config files generated successfully!${NC}"
echo ""
echo "Next: Generate docker-compose.yml and .env files"
