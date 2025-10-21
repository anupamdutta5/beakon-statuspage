#!/bin/bash

# Billing Consumer Startup Script
# This script sets up environment variables and starts the Billing Consumer

echo "Starting Billing Consumer..."

# Set required environment variables
export ENVIRONMENT="development"
export SERVICE_NAME="billing-consumer"
export SERVICE_VERSION="1.0.0"
export LOG_LEVEL="info"
export LOG_FORMAT="json"

# Queue configuration
export QUEUE_PROVIDER="redis"
export QUEUE_HOST="localhost"
export QUEUE_PORT=6379
export QUEUE_USERNAME=""
export QUEUE_PASSWORD=""
export QUEUE_NAME="billing"
export QUEUE_MAX_RETRIES=3
export QUEUE_RETRY_DELAY=60
export QUEUE_BATCH_SIZE=10
export QUEUE_POLL_TIMEOUT=30

# Database configuration
export DB_HOST="localhost"
export DB_PORT=5432
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="statuspage_billing"
export DB_SSL_MODE="disable"
export DB_MAX_CONNS=100
export DB_MIN_CONNS=10

# Billing configuration
export PROCESSING_ENABLED=true
export INVOICE_ENABLED=true
export PAYMENT_ENABLED=true
export TAX_ENABLED=true
export MAX_CONCURRENCY=10
export PROCESSING_DELAY=100
export RETRY_BACKOFF=5
export DEAD_LETTER_QUEUE="billing-dlq"
export CURRENCY="USD"
export TAX_RATE=0.0

echo "Environment variables set:"
echo "  ENVIRONMENT: $ENVIRONMENT"
echo "  SERVICE_NAME: $SERVICE_NAME"
echo "  QUEUE_PROVIDER: $QUEUE_PROVIDER"
echo "  QUEUE_HOST: $QUEUE_HOST"
echo "  QUEUE_NAME: $QUEUE_NAME"
echo "  PROCESSING_ENABLED: $PROCESSING_ENABLED"
echo "  INVOICE_ENABLED: $INVOICE_ENABLED"
echo "  PAYMENT_ENABLED: $PAYMENT_ENABLED"
echo "  TAX_ENABLED: $TAX_ENABLED"
echo "  MAX_CONCURRENCY: $MAX_CONCURRENCY"
echo "  CURRENCY: $CURRENCY"
echo "  TAX_RATE: $TAX_RATE"

# Start the Billing Consumer
echo "Starting Billing Consumer..."
go run ./cmd