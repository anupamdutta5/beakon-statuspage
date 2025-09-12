#!/bin/bash

# Simple Landing Page Service Startup Script (No Database Required)

echo "Starting Simple Landing Page Service..."

# Set environment variables
export SERVER_PORT=8100
export ENVIRONMENT="development"

echo "Environment variables set:"
echo "  SERVER_PORT: $SERVER_PORT"
echo "  ENVIRONMENT: $ENVIRONMENT"

# Start the simple server
echo "Starting Simple Landing Page Service on port $SERVER_PORT..."
go run simple-server.go
