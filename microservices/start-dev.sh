#!/bin/bash

echo "🚀 Starting Beakon Microservices"
echo "================================"

# Check Go
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed"
    exit 1
fi

echo "✅ Go is available"

# Create logs directory
mkdir -p logs

# Export development environment variables
export ENVIRONMENT=development
export LOG_LEVEL=debug

# Required environment variables for all services
export JWT_SECRET=dev-secret-for-testing-only-change-in-production
export JWT_EXPIRATION=24h
export JWT_ISSUER=statuspage-beakon

# Database configuration (use single database for development)
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=postgres
export DB_SSL_MODE=disable
export DB_MAX_CONNS=100
export DB_MIN_CONNS=10

# Redis configuration (disabled for now)
export REDIS_ENABLED=false
export REDIS_HOST=localhost
export REDIS_PORT=6379

# Security and CORS
export CORS_ENABLED=true
export CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8094
export CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
export CORS_ALLOWED_HEADERS=Authorization,Content-Type
export CORS_ALLOW_CREDENTIALS=true

# Rate limiting
export RATE_LIMIT_ENABLED=true
export RATE_LIMIT_REQUESTS_PER_SECOND=100
export RATE_LIMIT_BURST=200

# Timeouts
export READ_TIMEOUT=30s
export WRITE_TIMEOUT=30s
export IDLE_TIMEOUT=120s
export GRACEFUL_SHUTDOWN_TIMEOUT=30s

echo ""
echo "Starting services..."

# Skip database service (has compilation issues) and start core services only

# Start user service
if [ -d "user-service" ]; then
    echo "Starting user-service on port 8080..."
    cd user-service
    export PORT=8080
    nohup go run ./cmd > ../logs/user-service.log 2>&1 &
    cd ..
    sleep 3
fi

# Start API gateway
if [ -d "api-gateway" ]; then
    echo "Starting api-gateway on port 8081..."
    cd api-gateway
    export PORT=8081
    nohup go run ./cmd > ../logs/api-gateway.log 2>&1 &
    cd ..
    sleep 3
fi

# Start component service
if [ -d "component-service" ]; then
    echo "Starting component-service on port 8082..."
    cd component-service
    export PORT=8082
    nohup go run ./cmd > ../logs/component-service.log 2>&1 &
    cd ..
    sleep 3
fi

# Start analytics service
if [ -d "analytics-service" ]; then
    echo "Starting analytics-service on port 8083..."
    cd analytics-service
    export PORT=8083
    nohup go run ./cmd > ../logs/analytics-service.log 2>&1 &
    cd ..
    sleep 3
fi

# Start incident service
if [ -d "incident-service" ]; then
    echo "Starting incident-service on port 8084..."
    cd incident-service
    export PORT=8084
    nohup go run ./cmd > ../logs/incident-service.log 2>&1 &
    cd ..
    sleep 3
fi

# Start status UI
if [ -d "status-ui-service" ]; then
    echo "Starting status-ui-service on port 8094..."
    cd status-ui-service
    export PORT=8094
    nohup go run ./cmd > ../logs/status-ui-service.log 2>&1 &
    cd ..
    sleep 3
fi

echo ""
echo "🎉 Core services started!"
echo ""
echo "📋 Available URLs:"
echo "• API Gateway:  http://localhost:8081"
echo "• User Service: http://localhost:8080"
echo "• Status Page:  http://localhost:8094"
echo ""
echo "📝 View logs: tail -f logs/<service>.log"
echo "🛑 Stop: ./simple-stop.sh"
echo ""
echo "Services are running in background."