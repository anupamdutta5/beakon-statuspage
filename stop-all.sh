#!/bin/bash

# Beakon Status Page - Stop All Microservices

echo "🛑 Stopping Beakon Status Page Microservices"
echo "============================================="

if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose is not installed"
    exit 1
fi

echo "📦 Stopping all microservices..."
docker-compose -f docker-compose.microservices.yml down

echo "✅ All services stopped!"