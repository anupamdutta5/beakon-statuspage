#!/bin/bash

# Script to generate configuration files for all Beakon microservices
# This creates configs/ directories with config.yml, .env.example, and .env files

# Service configurations: SERVICE_NAME:PORT:DATABASE_NAME:NEEDS_RABBITMQ:NEEDS_REDIS
SERVICES=(
    "user-service:8081:users:0:0"
    "component-service:8084:components:0:0"
    "notification-service:8085:notifications:1:1"
    "incident-service:8086:incidents:0:0"
    "payment-service:8088:payments:0:0"
    "analytics-service:8090:analytics:0:1"
    "monitoring-service:8092:monitoring:0:1"
    "status-ui-service:8093:status_ui:0:0"
    "event-store-service:8096:events:1:1"
    "branding-service:8097:branding:0:0"
    "landing-page-service:8100:landing:0:0"
)

# Consumer services don't need database (they use event sourcing)
CONSUMERS=(
    "analytics-consumer:notifications"
    "notification-consumer:notifications"
    "audit-consumer:audit"
    "billing-consumer:payments"
)

# Function to create config.yml for backend services
create_backend_config() {
    local service=$1
    local port=$2
    local dbname=$3
    local needs_rabbitmq=$4
    local needs_redis=$5
    local config_file="$service/configs/config.yml"

    mkdir -p "$service/configs"

    cat > "$config_file" << EOF
# ${service^} Configuration
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
  metrics_port: $((port + 1000))
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

    # Add Redis if needed
    if [ "$needs_redis" = "1" ]; then
        cat >> "$config_file" << EOF

redis:
  enabled: \${REDIS_ENABLED:-true}
  host: \${REDIS_HOST:-localhost}
  port: 6379
  password: \${REDIS_PASSWORD}  # SECRET - from .env
  db: 0
  max_retries: 3
  pool_size: 10
  min_idle_conns: 5
EOF
    fi

    # Add RabbitMQ if needed
    if [ "$needs_rabbitmq" = "1" ]; then
        cat >> "$config_file" << EOF

rabbitmq:
  host: \${RABBITMQ_HOST:-localhost}
  port: 5672
  user: \${RABBITMQ_USER:-admin}
  password: \${RABBITMQ_PASSWORD}  # SECRET - from .env
  vhost: /
EOF
    fi

    echo "✓ Created $config_file"
}

# Function to create .env.example
create_env_example() {
    local service=$1
    local needs_rabbitmq=$2
    local needs_redis=$3
    local env_file="$service/configs/.env.example"

    cat > "$env_file" << EOF
# ${service^} Environment Variables Template
# Copy this file to .env and fill in your actual secrets

# Environment
ENVIRONMENT=development

# Database secrets
DB_HOST=localhost
DB_PASSWORD=your-secure-database-password
DB_SSL_MODE=disable

# JWT secret (MUST be at least 32 characters)
JWT_SECRET=your-very-long-secure-jwt-secret-minimum-32-characters-required

# Logging
LOG_LEVEL=debug
EOF

    if [ "$needs_redis" = "1" ]; then
        cat >> "$env_file" << EOF

# Redis configuration
REDIS_ENABLED=true
REDIS_HOST=localhost
REDIS_PASSWORD=your-redis-password
EOF
    fi

    if [ "$needs_rabbitmq" = "1" ]; then
        cat >> "$env_file" << EOF

# RabbitMQ secrets
RABBITMQ_HOST=localhost
RABBITMQ_USER=admin
RABBITMQ_PASSWORD=your-rabbitmq-password
EOF
    fi

    echo "✓ Created $env_file"
}

# Function to create .env with development values
create_env() {
    local service=$1
    local needs_rabbitmq=$2
    local needs_redis=$3
    local env_file="$service/configs/.env"

    cat > "$env_file" << EOF
# ${service^} - Local Development Secrets
# DO NOT COMMIT THIS FILE TO GIT

# Environment
ENVIRONMENT=development

# Database
DB_HOST=localhost
DB_PASSWORD=postgres
DB_SSL_MODE=disable

# JWT (min 32 characters)
JWT_SECRET=dev-jwt-secret-for-local-testing-only-change-in-production-min-32chars

# Logging
LOG_LEVEL=debug
EOF

    if [ "$needs_redis" = "1" ]; then
        cat >> "$env_file" << EOF

# Redis
REDIS_ENABLED=true
REDIS_HOST=localhost
REDIS_PASSWORD=
EOF
    fi

    if [ "$needs_rabbitmq" = "1" ]; then
        cat >> "$env_file" << EOF

# RabbitMQ
RABBITMQ_HOST=localhost
RABBITMQ_USER=admin
RABBITMQ_PASSWORD=SecureP@ssw0rd2024!
EOF
    fi

    echo "✓ Created $env_file"
}

# Main execution
echo "========================================="
echo "Generating configuration files for Beakon microservices"
echo "========================================="
echo ""

# Generate configs for backend services
for service_spec in "${SERVICES[@]}"; do
    IFS=':' read -r service port dbname rabbitmq redis <<< "$service_spec"

    if [ -d "$service" ]; then
        echo "Processing $service (port $port, db: $dbname)..."
        create_backend_config "$service" "$port" "$dbname" "$rabbitmq" "$redis"
        create_env_example "$service" "$rabbitmq" "$redis"
        create_env "$service" "$rabbitmq" "$redis"
        echo ""
    else
        echo "⚠ Skipping $service (directory not found)"
        echo ""
    fi
done

echo "========================================="
echo "Configuration generation complete!"
echo "========================================="
