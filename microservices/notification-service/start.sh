#!/bin/bash

# Notification Service Startup Script
# This script sets up environment variables and starts the Notification Service

echo "Starting Notification Service..."

# Set required environment variables
export PORT=8085
export HOST="0.0.0.0"
export ENVIRONMENT="development"
export SERVICE_NAME="notification-service"
export SERVICE_VERSION="1.0.0"
export LOG_LEVEL="info"

# Database configuration
export DB_HOST="localhost"
export DB_PORT=5432
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="statuspage_notifications"
export DB_SSL_MODE="disable"
export DB_MAX_CONNS=100
export DB_MIN_CONNS=10

# JWT configuration
export JWT_SECRET="development-secret-key-change-in-production"
export JWT_EXPIRATION=24
export JWT_ISSUER="statuspage-notification-service"

# Email configuration
export EMAIL_ENABLED=true
export SMTP_HOST="localhost"
export SMTP_PORT=587
export SMTP_USER=""
export SMTP_PASS=""
export FROM_EMAIL="noreply@statuspage.com"
export FROM_NAME="Status Page"
export SMTP_USE_TLS=true
export SMTP_USE_SSL=false
export EMAIL_MAX_RETRIES=3
export EMAIL_RETRY_DELAY=60

# SMS configuration
export SMS_ENABLED=false
export SMS_PROVIDER="twilio"
export SMS_ACCOUNT_SID=""
export SMS_AUTH_TOKEN=""
export SMS_FROM_NUMBER=""
export SMS_MAX_RETRIES=3
export SMS_RETRY_DELAY=60

# Webhook configuration
export WEBHOOK_ENABLED=true
export WEBHOOK_MAX_RETRIES=3
export WEBHOOK_RETRY_DELAY=60
export WEBHOOK_TIMEOUT=30
export WEBHOOK_SECRET=""

# Queue configuration
export QUEUE_ENABLED=false
export QUEUE_PROVIDER="redis"
export QUEUE_HOST="localhost"
export QUEUE_PORT=6379
export QUEUE_USERNAME=""
export QUEUE_PASSWORD=""
export QUEUE_NAME="notifications"
export QUEUE_MAX_RETRIES=3
export QUEUE_RETRY_DELAY=60

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
echo "  EMAIL_ENABLED: $EMAIL_ENABLED"
echo "  SMS_ENABLED: $SMS_ENABLED"
echo "  WEBHOOK_ENABLED: $WEBHOOK_ENABLED"
echo "  QUEUE_ENABLED: $QUEUE_ENABLED"

# Start the Notification Service
echo "Starting Notification Service on port $PORT..."
go run ./cmd