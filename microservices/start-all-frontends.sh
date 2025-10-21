#!/bin/bash

# Start All Frontend Services
# Starts both saas-admin-frontend (3001) and tenant-admin-frontend (3002)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;36m'
NC='\033[0m' # No Color

echo "========================================"
echo "Starting All Frontend Services"
echo "========================================"
echo ""

# Check Node.js version
echo -n "Checking Node.js version... "
NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    echo -e "${RED}FAILED${NC}"
    echo "Node.js 18+ is required. Current version: $(node -v)"
    exit 1
fi
echo -e "${GREEN}OK${NC} ($(node -v))"
echo ""

# Function to start a frontend service
start_frontend() {
    local service_name=$1
    local service_dir=$2
    local port=$3

    echo -e "${BLUE}Starting $service_name...${NC}"

    # Check if directory exists
    if [ ! -d "$service_dir" ]; then
        echo -e "${RED}Error: Directory $service_dir not found${NC}"
        return 1
    fi

    cd "$service_dir"

    # Install dependencies if needed
    if [ ! -d "node_modules" ]; then
        echo "Installing dependencies for $service_name..."
        npm install
    fi

    # Kill existing process on port if any
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo "Killing existing process on port $port..."
        lsof -ti :$port | xargs kill -9 2>/dev/null || true
        sleep 1
    fi

    # Start in background
    echo "Starting $service_name on port $port..."
    npm run dev > /tmp/$service_name.log 2>&1 &
    local pid=$!

    # Wait for service to start
    sleep 3

    # Check if process is still running
    if ps -p $pid > /dev/null; then
        echo -e "${GREEN}✓ $service_name started successfully (PID: $pid, Port: $port)${NC}"
    else
        echo -e "${RED}✗ $service_name failed to start${NC}"
        echo "Check logs: tail -f /tmp/$service_name.log"
        return 1
    fi

    cd - > /dev/null
    echo ""
}

# Start SaaS Admin Frontend
start_frontend "saas-admin-frontend" "saas-admin-frontend" 3001

# Start Tenant Admin Frontend
start_frontend "tenant-admin-frontend" "tenant-admin-frontend" 3002

echo "========================================"
echo "All Frontend Services Started"
echo "========================================"
echo ""
echo "Services:"
echo "  - SaaS Admin Frontend:    http://localhost:3001"
echo "  - Tenant Admin Frontend:  http://{subdomain}.localhost:3002"
echo ""
echo "Logs:"
echo "  - SaaS Admin:    tail -f /tmp/saas-admin-frontend.log"
echo "  - Tenant Admin:  tail -f /tmp/tenant-admin-frontend.log"
echo ""
echo "To stop all services: ./stop-all-frontends.sh"
echo ""
