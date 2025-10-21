#!/bin/bash
# Landing Page Service Startup Script

# Set environment variables
export ENVIRONMENT=development
export DB_HOST=localhost
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=statuspage_landing
export SERVER_PORT=8100

# Build the service
echo "Building Landing Page Service..."
go build -o landing-page-service cmd/main.go

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

# Start the service
echo "Starting Landing Page Service on port $SERVER_PORT..."
./landing-page-service
