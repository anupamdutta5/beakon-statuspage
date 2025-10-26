#!/bin/bash

# Script to create standardized YAML configuration files for all backend services
# This migrates from environment-variable-based config to YAML-first approach

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=============================================="
echo "Creating Configuration Files for All Services"
echo "=============================================="
echo ""

# Function to create config.yml for backend services
create_backend_config() {
    local service=$1
    local port=$2
    local dbname=$3
    local metrics_port=$4
    
    echo "Creating config for $service..."
    
    mkdir -p "$SCRIPT_DIR/$service/configs"
    
    cat > "$SCRIPT_DIR/$service/configs/config.yml" <<EOF
# $service Configuration
# This file contains all non-secret configuration
# Secrets (passwords, tokens) are in .env file

service:
  name: $service
  version: 1.0.0
  environment: \${ENVIRONMENT:-development}

server:
  port: $port
  host: 0.0.0.0
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s

database:
  host: \${DB_HOST:-localhost}
  port: 5432
  user: postgres
  password: \${DB_PASSWORD}  # SECRET - from .env
  name: $dbname
  ssl_mode: \${DB_SSL_MODE:-disable}
  max_conns: 100
  min_conns: 10
  conn_max_lifetime: 3600s  # 1 hour
  conn_max_idle_time: 600s   # 10 minutes

jwt:
  secret: \${JWT_SECRET}  # SECRET - from .env (min 32 characters)
  expiration: 24h
  issuer: beakon-$service

monitoring:
  enabled: true
  metrics_port: $metrics_port
  health_path: /health
  metrics_path: /metrics
  log_level: \${LOG_LEVEL:-info}

rate_limiting:
  enabled: true
  requests_per_minute: 100
  burst: 10

security:
  sanitization_enabled: true
  max_string_length: 1000
  strict_mode: false
EOF

    cat > "$SCRIPT_DIR/$service/configs/.env" <<EOF
ENVIRONMENT=development
JWT_SECRET=dev-secret-key-min-32-chars-long-change-in-production
DB_PASSWORD=postgres
LOG_LEVEL=debug
EOF

    echo "✓ Created configs for $service"
}

# Backend services with databases
create_backend_config "incident-service" 8086 "incidents" 9086
create_backend_config "payment-service" 8088 "payments" 9088
create_backend_config "analytics-service" 8090 "analytics" 9090
create_backend_config "status-ui-service" 8093 "status_ui" 9093
create_backend_config "event-store-service" 8096 "event_store" 9096
create_backend_config "branding-service" 8097 "branding" 9097
create_backend_config "landing-page-service" 8100 "landing_page" 9100

# notification-service (has RabbitMQ)
echo "Creating config for notification-service..."
mkdir -p "$SCRIPT_DIR/notification-service/configs"

cat > "$SCRIPT_DIR/notification-service/configs/config.yml" <<'EOF'
# notification-service Configuration

service:
  name: notification-service
  version: 1.0.0
  environment: ${ENVIRONMENT:-development}

server:
  port: 8085
  host: 0.0.0.0
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s

database:
  host: ${DB_HOST:-localhost}
  port: 5432
  user: postgres
  password: ${DB_PASSWORD}
  name: notifications
  ssl_mode: ${DB_SSL_MODE:-disable}
  max_conns: 100
  min_conns: 10
  conn_max_lifetime: 3600s
  conn_max_idle_time: 600s

rabbitmq:
  host: ${RABBITMQ_HOST:-localhost}
  port: 5672
  user: ${RABBITMQ_USER:-admin}
  password: ${RABBITMQ_PASSWORD}
  vhost: /
  queue: notifications
  exchange: beakon_events
  routing_key: notification

jwt:
  secret: ${JWT_SECRET}
  expiration: 24h
  issuer: beakon-notification-service

monitoring:
  enabled: true
  metrics_port: 9085
  health_path: /health
  metrics_path: /metrics
  log_level: ${LOG_LEVEL:-info}
EOF

cat > "$SCRIPT_DIR/notification-service/configs/.env" <<'EOF'
ENVIRONMENT=development
JWT_SECRET=dev-secret-key-min-32-chars-long-change-in-production
DB_PASSWORD=postgres
RABBITMQ_PASSWORD=SecureP@ssw0rd2024!
LOG_LEVEL=debug
EOF

echo "✓ Created configs for notification-service"

echo ""
echo "=============================================="
echo "Configuration Files Created Successfully!"
echo "=============================================="
echo ""
echo "Next steps:"
echo "1. Update each service's internal/config/config.go to use ConfigLoader"
echo "2. Test each service builds successfully"
echo "3. Commit the changes"

