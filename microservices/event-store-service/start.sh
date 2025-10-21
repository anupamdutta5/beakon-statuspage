#!/bin/bash

# Event Store Service Startup Script
# This script sets up environment variables and starts the Event Store Service

echo "Starting Event Store Service..."

# Set required environment variables
export ENVIRONMENT="development"
export SERVICE_NAME="event-store-service"
export SERVICE_VERSION="1.0.0"
export SERVICE_DESCRIPTION="Event Store Service for Status Page"
export LOG_LEVEL="info"
export LOG_FORMAT="json"

# Server configuration
export SERVER_HOST="0.0.0.0"
export SERVER_PORT=8096
export SERVER_READ_TIMEOUT=30
export SERVER_WRITE_TIMEOUT=30
export SERVER_IDLE_TIMEOUT=120

# Database configuration
export DB_HOST="localhost"
export DB_PORT=5432
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="statuspage_eventstore"
export DB_SSL_MODE="disable"
export DB_MAX_CONNS=100
export DB_MIN_CONNS=10
export DB_MAX_IDLE=10
export DB_MAX_LIFETIME=3600

# Event Store configuration
export MAX_EVENTS_PER_STREAM=10000
export SNAPSHOT_INTERVAL=100
export COMPRESSION_ENABLED=true
export ENCRYPTION_ENABLED=true
export ENCRYPTION_KEY=""
export RETENTION_DAYS=2555
export MAX_EVENT_SIZE=1048576
export BATCH_SIZE=100
export FLUSH_INTERVAL=1000
export REPLICATION_ENABLED=false
export REPLICATION_FACTOR=3

# Optional: Load from config file
# export CONFIG_FILE="/path/to/config.json"

echo "Environment variables set:"
echo "  SERVICE_NAME: $SERVICE_NAME"
echo "  SERVICE_VERSION: $SERVICE_VERSION"
echo "  SERVER_PORT: $SERVER_PORT"
echo "  DB_HOST: $DB_HOST"
echo "  DB_PORT: $DB_PORT"
echo "  DB_NAME: $DB_NAME"
echo "  MAX_EVENTS_PER_STREAM: $MAX_EVENTS_PER_STREAM"
echo "  SNAPSHOT_INTERVAL: $SNAPSHOT_INTERVAL"
echo "  COMPRESSION_ENABLED: $COMPRESSION_ENABLED"
echo "  ENCRYPTION_ENABLED: $ENCRYPTION_ENABLED"
echo "  RETENTION_DAYS: $RETENTION_DAYS"
echo "  BATCH_SIZE: $BATCH_SIZE"
echo "  FLUSH_INTERVAL: $FLUSH_INTERVAL"
echo "  REPLICATION_ENABLED: $REPLICATION_ENABLED"
echo "  REPLICATION_FACTOR: $REPLICATION_FACTOR"

# Start the service
echo "Starting Event Store Service on port $SERVER_PORT..."
go run cmd/main.go

