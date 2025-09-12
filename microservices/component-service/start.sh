#!/bin/bash

# Component Service Startup Script
# This script sets up environment variables and starts the Component Service

echo "Starting Component Service..."

# Set required environment variables
export PORT=8084
export HOST="0.0.0.0"
export ENVIRONMENT="development"
export SERVICE_NAME="component-service"
export SERVICE_VERSION="1.0.0"
export LOG_LEVEL="info"

# Database configuration
export DB_HOST="localhost"
export DB_PORT=5432
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="statuspage_components"
export DB_SSL_MODE="disable"
export DB_MAX_CONNS=100
export DB_MIN_CONNS=10

# JWT configuration
export JWT_SECRET="development-secret-key-change-in-production"
export JWT_EXPIRATION=24
export JWT_ISSUER="statuspage-component-service"

# Monitoring configuration
export MONITORING_ENABLED=true
export METRICS_PORT=9094
export HEALTH_PORT=8085

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

# Start the Component Service
echo "Starting Component Service on port $PORT..."
go run ./cmd