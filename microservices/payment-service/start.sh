#!/bin/bash

# Payment Service Startup Script
# This script sets up environment variables and starts the Payment Service

echo "Starting Payment Service..."

# Set required environment variables
export PORT=8088
export HOST="0.0.0.0"
export ENVIRONMENT="development"
export SERVICE_NAME="payment-service"
export SERVICE_VERSION="1.0.0"
export LOG_LEVEL="info"

# Database configuration
export DB_HOST="localhost"
export DB_PORT=5432
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="statuspage_payments"
export DB_SSL_MODE="disable"
export DB_MAX_CONNS=100
export DB_MIN_CONNS=10

# JWT configuration
export JWT_SECRET="development-secret-key-change-in-production"
export JWT_EXPIRATION=24
export JWT_ISSUER="statuspage-payment-service"

# Payment gateway configuration
export STRIPE_SECRET_KEY="sk_test_51234567890abcdefghijklmnopqrstuvwxyz"
export STRIPE_PUBLISHABLE_KEY="pk_test_51234567890abcdefghijklmnopqrstuvwxyz"
export STRIPE_WEBHOOK_SECRET="whsec_1234567890abcdefghijklmnopqrstuvwxyz"
export STRIPE_ENABLED=true

export PAYPAL_CLIENT_ID=""
export PAYPAL_CLIENT_SECRET=""
export PAYPAL_WEBHOOK_ID=""
export PAYPAL_ENABLED=false
export PAYPAL_SANDBOX=true

export RAZORPAY_KEY_ID=""
export RAZORPAY_KEY_SECRET=""
export RAZORPAY_WEBHOOK_SECRET=""
export RAZORPAY_ENABLED=false

export DEFAULT_CURRENCY="USD"
export WEBHOOK_SECRET="development-webhook-secret"

# Monitoring configuration
export MONITORING_ENABLED=true
export METRICS_PORT=9098
export HEALTH_PORT=8089

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
echo "  STRIPE_ENABLED: $STRIPE_ENABLED"
echo "  DEFAULT_CURRENCY: $DEFAULT_CURRENCY"

# Start the Payment Service
echo "Starting Payment Service on port $PORT..."
go run ./cmd