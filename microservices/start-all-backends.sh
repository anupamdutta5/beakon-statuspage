#!/bin/bash

# Start All Backend Services
# Starts both saas-admin-service (8098) and tenant-admin-service (8099)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;36m'
NC='\033[0m' # No Color

echo "========================================"
echo "Starting All Backend Services"
echo "========================================"
echo ""

# Common environment variables
export JWT_SECRET="dev-secret-for-testing-only-change-in-production-min-32-chars-long"
export RABBITMQ_URL="amqp://admin:SecureP@ssw0rd2024!@localhost:5672/"
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_SSLMODE=disable

# Check PostgreSQL
echo -n "Checking PostgreSQL... "
if pg_isready -h localhost -p 5432 > /dev/null 2>&1; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}FAILED${NC}"
    echo "PostgreSQL is not running on localhost:5432"
    echo "Start it with: brew services start postgresql@16"
    exit 1
fi

# Check RabbitMQ
echo -n "Checking RabbitMQ... "
if curl -s http://localhost:15672 > /dev/null 2>&1; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${YELLOW}WARNING${NC}"
    echo "RabbitMQ does not appear to be running. Services will fall back to HTTP sync."
fi
echo ""

# Function to start a backend service
start_backend() {
    local service_name=$1
    local service_dir=$2
    local port=$3
    local db_name=$4

    echo -e "${BLUE}Starting $service_name...${NC}"

    # Check if directory exists
    if [ ! -d "$service_dir" ]; then
        echo -e "${RED}Error: Directory $service_dir not found${NC}"
        return 1
    fi

    cd "$service_dir"

    # Check if binary exists
    if [ ! -f "$service_name" ]; then
        echo "Building $service_name..."
        go build -o "$service_name" cmd/main.go
    fi

    # Kill existing process on port if any
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo "Killing existing process on port $port..."
        lsof -ti :$port | xargs kill -9 2>/dev/null || true
        sleep 1
    fi

    # Start in background
    echo "Starting $service_name on port $port..."
    DB_NAME=$db_name SERVER_PORT=$port ./$service_name > /tmp/$service_name.log 2>&1 &
    local pid=$!

    # Wait for service to start
    sleep 3

    # Check if process is still running
    if ps -p $pid > /dev/null; then
        echo -e "${GREEN}✓ $service_name started successfully (PID: $pid, Port: $port)${NC}"

        # Test health endpoint
        if curl -s http://localhost:$port/health > /dev/null 2>&1 || curl -s http://localhost:$port/api/v1/health > /dev/null 2>&1; then
            echo -e "${GREEN}  Health check passed${NC}"
        else
            echo -e "${YELLOW}  Health check failed (service may still be initializing)${NC}"
        fi
    else
        echo -e "${RED}✗ $service_name failed to start${NC}"
        echo "Check logs: tail -f /tmp/$service_name.log"
        return 1
    fi

    cd - > /dev/null
    echo ""
}

# Start SaaS Admin Service
start_backend "saas-admin-service" "saas-admin-service" 8098 "saas_admin"

# Start Tenant Admin Service
start_backend "tenant-admin-service" "tenant-admin-service" 8099 "tenant_admin_db"

echo "========================================"
echo "All Backend Services Started"
echo "========================================"
echo ""
echo "Services:"
echo "  - SaaS Admin API:     http://localhost:8098/api/v1/health"
echo "  - Tenant Admin API:   http://localhost:8099/health"
echo ""
echo "Logs:"
echo "  - SaaS Admin:    tail -f /tmp/saas-admin-service.log"
echo "  - Tenant Admin:  tail -f /tmp/tenant-admin-service.log"
echo ""
echo "To stop all services: ./stop-all-backends.sh"
echo ""
