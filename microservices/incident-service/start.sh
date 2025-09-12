#!/bin/bash

# Incident Service Startup Script
# This script sets up environment variables and starts the Incident Service

echo "Starting Incident Service..."

# Set required environment variables
export PORT=8086
export HOST="0.0.0.0"
export ENVIRONMENT="development"
export SERVICE_NAME="incident-service"
export SERVICE_VERSION="1.0.0"
export LOG_LEVEL="info"

# Database configuration
export DB_HOST="localhost"
export DB_PORT=5432
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="statuspage_incidents"
export DB_SSL_MODE="disable"
export DB_MAX_CONNS=100
export DB_MIN_CONNS=10

# JWT configuration
export JWT_SECRET="development-secret-key-change-in-production"
export JWT_EXPIRATION=24
export JWT_ISSUER="statuspage-incident-service"

# Monitoring configuration
export MONITORING_ENABLED=true
export METRICS_PORT=9096
export HEALTH_PORT=8087

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

# Start the Incident Service
echo "Starting Incident Service on port $PORT..."
go run ./cmd