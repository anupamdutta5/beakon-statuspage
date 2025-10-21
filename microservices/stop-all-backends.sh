#!/bin/bash

# Stop All Backend Services
# Stops both saas-admin-service (8098) and tenant-admin-service (8099)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================"
echo "Stopping All Backend Services"
echo "========================================"
echo ""

# Function to stop service on a port
stop_service() {
    local service_name=$1
    local port=$2

    echo -n "Stopping $service_name (port $port)... "

    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        lsof -ti :$port | xargs kill -9 2>/dev/null || true
        sleep 1

        # Verify it's stopped
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            echo -e "${RED}FAILED${NC}"
        else
            echo -e "${GREEN}OK${NC}"
        fi
    else
        echo -e "${YELLOW}Not running${NC}"
    fi
}

# Alternative: Kill by process name
stop_by_name() {
    local service_name=$1

    echo -n "Stopping $service_name by process name... "

    if pkill -f "$service_name" > /dev/null 2>&1; then
        sleep 1
        echo -e "${GREEN}OK${NC}"
    else
        echo -e "${YELLOW}Not running${NC}"
    fi
}

# Stop by port (preferred)
stop_service "saas-admin-service" 8098
stop_service "tenant-admin-service" 8099

echo ""

# Also try to kill by process name (in case processes are on different ports)
echo "Ensuring all backend processes are stopped..."
stop_by_name "saas-admin-service"
stop_by_name "tenant-admin-service"

echo ""
echo -e "${GREEN}All backend services stopped${NC}"
echo ""
