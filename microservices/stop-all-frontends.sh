#!/bin/bash

# Stop All Frontend Services
# Stops both saas-admin-frontend (3001) and tenant-admin-frontend (3002)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================"
echo "Stopping All Frontend Services"
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

# Stop both frontend services
stop_service "saas-admin-frontend" 3001
stop_service "tenant-admin-frontend" 3002

echo ""
echo -e "${GREEN}All frontend services stopped${NC}"
echo ""
