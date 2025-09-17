#!/bin/bash

# Beakon Status Page - Start All Microservices
# Simple orchestration script to start all microservices

echo "🚀 Starting Beakon Status Page Microservices"
echo "============================================="

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose is not installed"
    exit 1
fi

# Start all services using Docker Compose
echo "📦 Starting all microservices..."
docker-compose -f docker-compose.microservices.yml up -d

echo "✅ All services started!"
echo ""
echo "To view logs: docker-compose -f docker-compose.microservices.yml logs -f"
echo "To stop all:  docker-compose -f docker-compose.microservices.yml down"