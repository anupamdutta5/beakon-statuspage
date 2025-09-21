#!/bin/bash

# Analytics Service Startup Script
# Get service name from directory
SERVICE_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVICE_NAME="$(basename "$SERVICE_DIR")"

# Load environment variables from .env file if it exists
if [ -f .env ]; then
    echo "Loading environment from .env file..."
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "Warning: No .env file found. Using default development settings."
    echo "For production, create .env file from .env.example"

    # Development-only defaults (DO NOT use in production)
    export PORT=${PORT:-8090}
    export HOST=${HOST:-"0.0.0.0"}
    export ENVIRONMENT=${ENVIRONMENT:-"development"}

    # Security check for JWT_SECRET
    if [ -z "$JWT_SECRET" ]; then
        if [ "$ENVIRONMENT" = "production" ]; then
            echo "ERROR: JWT_SECRET must be set in production environment"
            exit 1
        else
            # Generate a random secret for development only
            export JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || echo "dev-only-insecure-secret")
            echo "Generated development JWT_SECRET (not for production use)"
        fi
    fi

    export JWT_EXPIRATION=${JWT_EXPIRATION:-24}
    export JWT_ISSUER=${JWT_ISSUER:-"statuspage-$SERVICE_NAME"}

    # Database configuration with security checks
    export DB_HOST=${DB_HOST:-"localhost"}
    export DB_PORT=${DB_PORT:-5432}
    export DB_USER=${DB_USER:-"postgres"}

    if [ -z "$DB_PASSWORD" ]; then
        if [ "$ENVIRONMENT" = "production" ]; then
            echo "ERROR: DB_PASSWORD must be set in production environment"
            exit 1
        else
            export DB_PASSWORD="postgres"
            echo "Warning: Using default DB_PASSWORD (development only)"
        fi
    fi

    export DB_NAME=${DB_NAME:-"statuspage_analytics"}
    export DB_SSL_MODE=${DB_SSL_MODE:-"disable"}
    export DB_MAX_CONNS=${DB_MAX_CONNS:-100}
    export DB_MIN_CONNS=${DB_MIN_CONNS:-10}
fi

# Service configuration
export SERVICE_NAME="analytics-service"
export SERVICE_VERSION=${SERVICE_VERSION:-"1.0.0"}
export LOG_LEVEL=${LOG_LEVEL:-"info"}

# Analytics configuration
export DATA_RETENTION_DAYS=365
export AGGREGATION_INTERVAL="daily"
export CACHE_ENABLED=true
export CACHE_TTL=300
export EXPORT_FORMATS="csv,json,excel"
export MAX_DATA_POINTS=10000
export REAL_TIME_ENABLED=true

# Monitoring configuration
export MONITORING_ENABLED=true
export METRICS_PORT=9100
export HEALTH_PORT=8091

# Server timeouts
export READ_TIMEOUT=30
export WRITE_TIMEOUT=30
export IDLE_TIMEOUT=120

echo "Environment variables set:"
echo "  PORT: $PORT"
echo "  HOST: $HOST"
echo "  ENVIRONMENT: $ENVIRONMENT"
echo "  SERVICE_NAME: $SERVICE_NAME"
echo "  DB_HOST: $DB_HOST"
echo "  DB_NAME: $DB_NAME"
echo "  DATA_RETENTION_DAYS: $DATA_RETENTION_DAYS"
echo "  CACHE_ENABLED: $CACHE_ENABLED"
echo "  REAL_TIME_ENABLED: $REAL_TIME_ENABLED"

# Start the Analytics Service
echo "Starting Analytics Service on port $PORT..."
go run ./cmd