#!/bin/bash

# Script to create configuration files for all remaining Beakon microservices
# This creates configs/ directories with complete config.yml, .env, and .env.example files

cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices

# Service configurations: SERVICE_NAME:PORT:DATABASE_NAME:METRICS_PORT
declare -A SERVICES=(
    ["user-service"]="8081:users:9081"
    ["component-service"]="8084:components:9084"
    ["notification-service"]="8085:notifications:9085"
    ["incident-service"]="8086:incidents:9086"
    ["payment-service"]="8088:payments:9088"
    ["analytics-service"]="8090:analytics:9090"
    ["monitoring-service"]="8092:monitoring:9092"
    ["status-ui-service"]="8093:status_ui:9093"
    ["event-store-service"]="8096:events:9096"
    ["branding-service"]="8097:branding:9097"
    ["landing-page-service"]="8100:landing:9100"
)

echo "========================================="
echo "Creating configuration files for backend services"
echo "========================================="
echo ""

for service in "${!SERVICES[@]}"; do
    IFS=':' read -r port dbname metrics_port <<< "${SERVICES[$service]}"
    
    echo "Processing $service (port $port, db: $dbname)..."
    
    mkdir -p "$service/configs"
    
    # Create config.yml
    cat > "$service/configs/config.yml" << EOF
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
    
    # Create config.production.yml
    cat > "$service/configs/config.production.yml" << EOF
# Production Configuration Overrides for ${service^}
# These values override the base config.yml in production environment

database:
  ssl_mode: require  # SSL required in production
  max_conns: 200
  min_conns: 20

monitoring:
  log_level: info

security:
  strict_mode: true  # Enable strict validation in production
EOF
    
    # Create .env.example
    cat > "$service/configs/.env.example" << EOF
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
    
    # Create .env
    cat > "$service/configs/.env" << EOF
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
    
    echo "✓ Created config files for $service"
    echo ""
done

echo "========================================="
echo "Configuration files created successfully!"
echo "========================================="
