#!/bin/bash

# Start the Fixed Landing Page Service
echo "🚀 Starting the Fixed Landing Page Service..."

# Navigate to the landing page service directory
cd microservices/landing-page-service

# Set environment variables
export SERVER_PORT=8100
export ENVIRONMENT=development
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=statuspage_landing
export DB_SSL_MODE=disable

echo "✅ Environment variables set"
echo "📁 Working directory: $(pwd)"
echo "🌐 Server will start on port: $SERVER_PORT"

# Start the service
echo "🎯 Starting the service..."
go run cmd/main.go
