#!/bin/bash

# Database Service Startup Script
# This script sets up environment variables and starts the Database Service

echo "Starting Database Service..."

# Set required environment variables
export ENVIRONMENT="development"
export SERVICE_NAME="database-service"
export SERVICE_VERSION="1.0.0"
export LOG_LEVEL="info"
export LOG_FORMAT="json"

# Server configuration
export SERVER_HOST="0.0.0.0"
export SERVER_PORT=8095
export SERVER_READ_TIMEOUT=30
export SERVER_WRITE_TIMEOUT=30
export SERVER_IDLE_TIMEOUT=120

# Database configuration
export DB_HOST="localhost"
export DB_PORT=5432
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="statuspage_database"
export DB_SSL_MODE="disable"
export DB_MAX_CONNS=100
export DB_MIN_CONNS=10
export DB_MAX_IDLE=10
export DB_MAX_LIFETIME=3600

# Cache configuration
export CACHE_PROVIDER="redis"
export CACHE_HOST="localhost"
export CACHE_PORT=6379
export CACHE_PASSWORD=""
export CACHE_DB=0
export CACHE_TTL=3600

echo "Environment variables set:"
echo "  ENVIRONMENT: $ENVIRONMENT"
echo "  SERVICE_NAME: $SERVICE_NAME"
echo "  SERVER_HOST: $SERVER_HOST"
echo "  SERVER_PORT: $SERVER_PORT"
echo "  DB_HOST: $DB_HOST"
echo "  DB_PORT: $DB_PORT"
echo "  DB_NAME: $DB_NAME"
echo "  CACHE_PROVIDER: $CACHE_PROVIDER"
echo "  CACHE_HOST: $CACHE_HOST"

# Start the Database Service
echo "Starting Database Service..."
go run ./cmd

