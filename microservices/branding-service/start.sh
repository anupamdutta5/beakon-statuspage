#!/bin/bash

# Branding Service Startup Script
# This script sets up environment variables and starts the Branding Service

echo "Starting Branding Service..."

# Set required environment variables
export ENVIRONMENT="development"
export SERVICE_NAME="branding-service"
export SERVICE_VERSION="1.0.0"
export SERVICE_DESCRIPTION="Professional Branding Service for Status Page"
export LOG_LEVEL="info"
export LOG_FORMAT="json"

# Server configuration
export SERVER_HOST="0.0.0.0"
export SERVER_PORT=8097
export SERVER_READ_TIMEOUT=30
export SERVER_WRITE_TIMEOUT=30
export SERVER_IDLE_TIMEOUT=120

# Database configuration
export DB_HOST="localhost"
export DB_PORT=5432
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="statuspage_branding"
export DB_SSL_MODE="disable"
export DB_MAX_CONNS=100
export DB_MIN_CONNS=10
export DB_MAX_IDLE=10
export DB_MAX_LIFETIME=3600

# Storage configuration
export STORAGE_TYPE="local"
export STORAGE_LOCAL_PATH="./uploads"
export STORAGE_MAX_FILE_SIZE=10485760
export STORAGE_ALLOWED_TYPES="image/jpeg,image/png,image/gif,image/svg+xml,image/webp,font/woff,font/woff2,text/css,application/javascript"

# Branding configuration
export BRANDING_DEFAULT_THEME="modern"
export BRANDING_AVAILABLE_THEMES="modern,classic,minimal,dark,corporate"
export BRANDING_CUSTOM_CSS_ENABLED=true
export BRANDING_CUSTOM_JS_ENABLED=true
export BRANDING_LOGO_MAX_SIZE=2097152
export BRANDING_FAVICON_MAX_SIZE=1048576
export BRANDING_ALLOWED_IMAGE_FORMATS="jpg,jpeg,png,gif,svg,webp"
export BRANDING_ALLOWED_FONT_FORMATS="woff,woff2,ttf,otf"
export BRANDING_CDN_ENABLED=false
export BRANDING_CDN_URL=""
export BRANDING_CACHE_ENABLED=true
export BRANDING_CACHE_TTL=3600

# Optional: Load from config file
# export CONFIG_FILE="/path/to/config.json"

echo "Environment variables set:"
echo "  SERVICE_NAME: $SERVICE_NAME"
echo "  SERVICE_VERSION: $SERVICE_VERSION"
echo "  SERVER_PORT: $SERVER_PORT"
echo "  DB_HOST: $DB_HOST"
echo "  DB_PORT: $DB_PORT"
echo "  DB_NAME: $DB_NAME"
echo "  STORAGE_TYPE: $STORAGE_TYPE"
echo "  STORAGE_LOCAL_PATH: $STORAGE_LOCAL_PATH"
echo "  STORAGE_MAX_FILE_SIZE: $STORAGE_MAX_FILE_SIZE"
echo "  BRANDING_DEFAULT_THEME: $BRANDING_DEFAULT_THEME"
echo "  BRANDING_AVAILABLE_THEMES: $BRANDING_AVAILABLE_THEMES"
echo "  BRANDING_CUSTOM_CSS_ENABLED: $BRANDING_CUSTOM_CSS_ENABLED"
echo "  BRANDING_CUSTOM_JS_ENABLED: $BRANDING_CUSTOM_JS_ENABLED"
echo "  BRANDING_LOGO_MAX_SIZE: $BRANDING_LOGO_MAX_SIZE"
echo "  BRANDING_FAVICON_MAX_SIZE: $BRANDING_FAVICON_MAX_SIZE"
echo "  BRANDING_ALLOWED_IMAGE_FORMATS: $BRANDING_ALLOWED_IMAGE_FORMATS"
echo "  BRANDING_ALLOWED_FONT_FORMATS: $BRANDING_ALLOWED_FONT_FORMATS"
echo "  BRANDING_CDN_ENABLED: $BRANDING_CDN_ENABLED"
echo "  BRANDING_CACHE_ENABLED: $BRANDING_CACHE_ENABLED"
echo "  BRANDING_CACHE_TTL: $BRANDING_CACHE_TTL"

# Start the service
echo "Starting Branding Service on port $SERVER_PORT..."
go run cmd/main.go

