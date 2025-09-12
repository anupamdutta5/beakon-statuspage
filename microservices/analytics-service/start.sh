#!/bin/bash

# Analytics Service Startup Script
# This script sets up environment variables and starts the Analytics Service

echo "Starting Analytics Service..."

# Set required environment variables
export PORT=8090
export HOST="0.0.0.0"
export ENVIRONMENT="development"
export SERVICE_NAME="analytics-service"
export SERVICE_VERSION="1.0.0"
export LOG_LEVEL="info"

# Database configuration
export DB_HOST="localhost"
export DB_PORT=5432
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="statuspage_analytics"
export DB_SSL_MODE="disable"
export DB_MAX_CONNS=100
export DB_MIN_CONNS=10

# JWT configuration
export JWT_SECRET="development-secret-key-change-in-production"
export JWT_EXPIRATION=24
export JWT_ISSUER="statuspage-analytics-service"

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