#!/bin/bash

# Monitoring Service Startup Script
# This script sets up environment variables and starts the Monitoring Service

echo "Starting Monitoring Service..."

# Set required environment variables
export PORT=8092
export HOST="0.0.0.0"
export ENVIRONMENT="development"
export SERVICE_NAME="monitoring-service"
export SERVICE_VERSION="1.0.0"
export LOG_LEVEL="info"

# Database configuration
export DB_HOST="localhost"
export DB_PORT=5432
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="statuspage_monitoring"
export DB_SSL_MODE="disable"
export DB_MAX_CONNS=100
export DB_MIN_CONNS=10

# JWT configuration
export JWT_SECRET="development-secret-key-change-in-production"
export JWT_EXPIRATION=24
export JWT_ISSUER="statuspage-monitoring-service"

# Monitoring configuration
export MONITORING_ENABLED=true
export METRICS_PORT=9102
export HEALTH_PORT=8093
export CHECK_INTERVAL=60
export ALERT_COOLDOWN=15
export RETENTION_DAYS=30
export MAX_CONCURRENT_CHECKS=100
export TIMEOUT_SECONDS=30

# Prometheus configuration
export PROMETHEUS_ENABLED=true
export PROMETHEUS_PORT=9102
export PROMETHEUS_PATH="/metrics"
export PROMETHEUS_NAMESPACE="statuspage"
export PROMETHEUS_SUBSYSTEM="monitoring"

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
echo "  MONITORING_ENABLED: $MONITORING_ENABLED"
echo "  CHECK_INTERVAL: $CHECK_INTERVAL"
echo "  PROMETHEUS_ENABLED: $PROMETHEUS_ENABLED"

# Start the Monitoring Service
echo "Starting Monitoring Service on port $PORT..."
go run ./cmd