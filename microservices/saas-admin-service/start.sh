#!/bin/bash

# SaaS Admin Service Startup Script
# This script sets up environment variables and starts the SaaS Admin Service

echo "Starting SaaS Admin Service..."

# Set required environment variables
export ENVIRONMENT="development"
export SERVICE_NAME="saas-admin-service"
export SERVICE_VERSION="1.0.0"
export SERVICE_DESCRIPTION="SaaS Admin Service for Platform Administration"
export LOG_LEVEL="info"
export LOG_FORMAT="json"

# Server configuration
export SERVER_HOST="0.0.0.0"
export SERVER_PORT=8098
export SERVER_READ_TIMEOUT=30
export SERVER_WRITE_TIMEOUT=30
export SERVER_IDLE_TIMEOUT=120

# Database configuration
export DB_HOST="localhost"
export DB_PORT=5432
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="statuspage_saas_admin"
export DB_SSL_MODE="disable"
export DB_MAX_CONNS=100
export DB_MIN_CONNS=10
export DB_MAX_IDLE=10
export DB_MAX_LIFETIME=3600

# SaaS configuration
export SAAS_PLATFORM_NAME="StatusPage Pro"
export SAAS_PLATFORM_URL="https://statuspage.pro"
export SAAS_ADMIN_EMAIL="admin@statuspage.pro"
export SAAS_SUPPORT_EMAIL="support@statuspage.pro"
export SAAS_DEFAULT_PLAN="free"
export SAAS_AVAILABLE_PLANS="free,pro,enterprise"
export SAAS_MAX_TENANTS_PER_PLAN=1000
export SAAS_DEFAULT_TRIAL_DAYS=14
export SAAS_BILLING_ENABLED=true
export SAAS_ANALYTICS_ENABLED=true
export SAAS_MONITORING_ENABLED=true
export SAAS_FEATURE_FLAGS_ENABLED=true
export SAAS_AUDIT_LOGGING_ENABLED=true
export SAAS_BACKUP_ENABLED=true
export SAAS_BACKUP_INTERVAL=24
export SAAS_RETENTION_DAYS=2555

# Optional: Load from config file
# export CONFIG_FILE="/path/to/config.json"

echo "Environment variables set:"
echo "  SERVICE_NAME: $SERVICE_NAME"
echo "  SERVICE_VERSION: $SERVICE_VERSION"
echo "  SERVER_PORT: $SERVER_PORT"
echo "  DB_HOST: $DB_HOST"
echo "  DB_PORT: $DB_PORT"
echo "  DB_NAME: $DB_NAME"
echo "  SAAS_PLATFORM_NAME: $SAAS_PLATFORM_NAME"
echo "  SAAS_PLATFORM_URL: $SAAS_PLATFORM_URL"
echo "  SAAS_ADMIN_EMAIL: $SAAS_ADMIN_EMAIL"
echo "  SAAS_SUPPORT_EMAIL: $SAAS_SUPPORT_EMAIL"
echo "  SAAS_DEFAULT_PLAN: $SAAS_DEFAULT_PLAN"
echo "  SAAS_AVAILABLE_PLANS: $SAAS_AVAILABLE_PLANS"
echo "  SAAS_MAX_TENANTS_PER_PLAN: $SAAS_MAX_TENANTS_PER_PLAN"
echo "  SAAS_DEFAULT_TRIAL_DAYS: $SAAS_DEFAULT_TRIAL_DAYS"
echo "  SAAS_BILLING_ENABLED: $SAAS_BILLING_ENABLED"
echo "  SAAS_ANALYTICS_ENABLED: $SAAS_ANALYTICS_ENABLED"
echo "  SAAS_MONITORING_ENABLED: $SAAS_MONITORING_ENABLED"
echo "  SAAS_FEATURE_FLAGS_ENABLED: $SAAS_FEATURE_FLAGS_ENABLED"
echo "  SAAS_AUDIT_LOGGING_ENABLED: $SAAS_AUDIT_LOGGING_ENABLED"
echo "  SAAS_BACKUP_ENABLED: $SAAS_BACKUP_ENABLED"
echo "  SAAS_BACKUP_INTERVAL: $SAAS_BACKUP_INTERVAL"
echo "  SAAS_RETENTION_DAYS: $SAAS_RETENTION_DAYS"

# Start the service
echo "Starting SaaS Admin Service on port $SERVER_PORT..."
go run cmd/main.go

