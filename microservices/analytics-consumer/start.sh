#!/bin/bash

# Analytics Consumer Startup Script
# This script sets up environment variables and starts the Analytics Consumer

echo "Starting Analytics Consumer..."

# Set required environment variables
export ENVIRONMENT="development"
export SERVICE_NAME="analytics-consumer"
export SERVICE_VERSION="1.0.0"
export LOG_LEVEL="info"
export LOG_FORMAT="json"

# Queue configuration
export QUEUE_PROVIDER="redis"
export QUEUE_HOST="localhost"
export QUEUE_PORT=6379
export QUEUE_USERNAME=""
export QUEUE_PASSWORD=""
export QUEUE_NAME="analytics"
export QUEUE_MAX_RETRIES=3
export QUEUE_RETRY_DELAY=60
export QUEUE_BATCH_SIZE=10
export QUEUE_POLL_TIMEOUT=30

# Database configuration
export DB_HOST="localhost"
export DB_PORT=5432
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="statuspage_analytics"
export DB_SSL_MODE="disable"
export DB_MAX_CONNS=100
export DB_MIN_CONNS=10

# Analytics configuration
export PROCESSING_ENABLED=true
export AGGREGATION_ENABLED=true
export REPORTING_ENABLED=true
export MAX_CONCURRENCY=10
export PROCESSING_DELAY=100
export RETRY_BACKOFF=5
export DEAD_LETTER_QUEUE="analytics-dlq"
export BATCH_SIZE=100
export FLUSH_INTERVAL=60

echo "Environment variables set:"
echo "  ENVIRONMENT: $ENVIRONMENT"
echo "  SERVICE_NAME: $SERVICE_NAME"
echo "  QUEUE_PROVIDER: $QUEUE_PROVIDER"
echo "  QUEUE_HOST: $QUEUE_HOST"
echo "  QUEUE_NAME: $QUEUE_NAME"
echo "  PROCESSING_ENABLED: $PROCESSING_ENABLED"
echo "  AGGREGATION_ENABLED: $AGGREGATION_ENABLED"
echo "  REPORTING_ENABLED: $REPORTING_ENABLED"
echo "  MAX_CONCURRENCY: $MAX_CONCURRENCY"

# Start the Analytics Consumer
echo "Starting Analytics Consumer..."
go run ./cmd