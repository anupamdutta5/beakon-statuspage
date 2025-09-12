#!/bin/bash
export PORT=8082
export HOST="0.0.0.0"
export ENVIRONMENT="development"
export SERVICE_NAME="tenant-service"
export SERVICE_VERSION="1.0.0"
export DB_HOST="localhost"
export DB_PORT=5432
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="statuspage_tenants"
export DB_SSL_MODE="disable"
export LOG_LEVEL="info"
echo "Starting Tenant Service on port $PORT..."
go run ./cmd
