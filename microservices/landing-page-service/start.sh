#!/bin/bash

# Landing Page Service Startup Script
# This script sets up environment variables and starts the Landing Page Service

echo "Starting Landing Page Service..."

# Set required environment variables
export ENVIRONMENT="development"
export SERVICE_NAME="landing-page-service"
export SERVICE_VERSION="1.0.0"
export SERVICE_DESCRIPTION="Beautiful Landing Page Service with Content Management"
export LOG_LEVEL="info"
export LOG_FORMAT="json"

# Server configuration
export SERVER_HOST="0.0.0.0"
export SERVER_PORT=8100
export SERVER_READ_TIMEOUT=30
export SERVER_WRITE_TIMEOUT=30
export SERVER_IDLE_TIMEOUT=120

# Database configuration
export DB_HOST="localhost"
export DB_PORT=5432
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="statuspage_landing"
export DB_SSL_MODE="disable"
export DB_MAX_CONNS=100
export DB_MIN_CONNS=10
export DB_MAX_IDLE=10
export DB_MAX_LIFETIME=3600

# Landing page configuration
export LANDING_SITE_NAME="StatusPage Pro"
export LANDING_SITE_URL="https://statuspage.pro"
export LANDING_SITE_DESCRIPTION="Professional status page platform for modern teams"
export LANDING_SITE_KEYWORDS="status page,uptime monitoring,incident management,team communication"
export LANDING_CONTACT_EMAIL="hello@statuspage.pro"
export LANDING_SUPPORT_EMAIL="support@statuspage.pro"
export LANDING_SOCIAL_LINKS='{"twitter":"https://twitter.com/statuspagepro","linkedin":"https://linkedin.com/company/statuspagepro","github":"https://github.com/statuspagepro"}'
export LANDING_ANALYTICS_ID=""
export LANDING_CDN_ENABLED=false
export LANDING_CDN_URL=""
export LANDING_CACHE_ENABLED=true
export LANDING_CACHE_TTL=3600
export LANDING_SEO_ENABLED=true
export LANDING_OG_IMAGE="/static/images/og-image.png"
export LANDING_FAVICON="/static/images/favicon.ico"
export LANDING_THEME="modern"
export LANDING_CUSTOM_CSS=""
export LANDING_CUSTOM_JS=""

# Optional: Load from config file
# export CONFIG_FILE="/path/to/config.json"

echo "Environment variables set:"
echo "  SERVICE_NAME: $SERVICE_NAME"
echo "  SERVICE_VERSION: $SERVICE_VERSION"
echo "  SERVER_PORT: $SERVER_PORT"
echo "  DB_HOST: $DB_HOST"
echo "  DB_PORT: $DB_PORT"
echo "  DB_NAME: $DB_NAME"
echo "  LANDING_SITE_NAME: $LANDING_SITE_NAME"
echo "  LANDING_SITE_URL: $LANDING_SITE_URL"
echo "  LANDING_SITE_DESCRIPTION: $LANDING_SITE_DESCRIPTION"
echo "  LANDING_CONTACT_EMAIL: $LANDING_CONTACT_EMAIL"
echo "  LANDING_SUPPORT_EMAIL: $LANDING_SUPPORT_EMAIL"
echo "  LANDING_ANALYTICS_ID: $LANDING_ANALYTICS_ID"
echo "  LANDING_CDN_ENABLED: $LANDING_CDN_ENABLED"
echo "  LANDING_CACHE_ENABLED: $LANDING_CACHE_ENABLED"
echo "  LANDING_SEO_ENABLED: $LANDING_SEO_ENABLED"
echo "  LANDING_THEME: $LANDING_THEME"

# Start the service
echo "Starting Landing Page Service on port $SERVER_PORT..."
go run cmd/main.go

