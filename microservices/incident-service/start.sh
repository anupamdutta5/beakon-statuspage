#!/bin/bash

# User Service Startup Script

# Load environment variables from .env file if it exists
if [ -f .env ]; then
    echo "Loading environment from .env file..."
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "Warning: No .env file found. Using default development settings."
    echo "For production, create .env file from .env.template"

    # Development-only defaults (DO NOT use in production)
    export PORT=${PORT:-}
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
    export JWT_ISSUER=${JWT_ISSUER:-"statuspage-incident-service"}

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

    export DB_NAME=${DB_NAME:-"statuspage_incident_service"}
    export DB_SSL_MODE=${DB_SSL_MODE:-"disable"}
    export DB_MAX_CONNS=${DB_MAX_CONNS:-100}
    export DB_MIN_CONNS=${DB_MIN_CONNS:-10}
fi

# Service configuration
export SERVICE_NAME="incident-service"
export SERVICE_VERSION="1.0.0"
export SERVICE_DESCRIPTION="User Management Service for Status Page"

# Monitoring
export MONITORING_ENABLED=true
export METRICS_PORT=9091
export HEALTH_PORT=8082
export LOG_LEVEL="info"

# Server timeouts
export READ_TIMEOUT=30
export WRITE_TIMEOUT=30
export IDLE_TIMEOUT=120

echo "Starting User Service..."
echo "Port: $PORT"
echo "Environment: $ENVIRONMENT"
echo "Service: $SERVICE_NAME v$SERVICE_VERSION"
echo "Database: $DB_HOST:$DB_PORT/$DB_NAME"

# Start the service
go run ./cmd
