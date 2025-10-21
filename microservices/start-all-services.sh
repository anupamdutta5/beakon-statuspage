#!/bin/bash

# Start All Services - Complete Stack
# Starts both backend and frontend services

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

echo ""
echo -e "${BOLD}========================================"
echo "Beakon Status Page - Complete Stack"
echo "========================================${NC}"
echo ""
echo "This script will start:"
echo "  1. Backend services (saas-admin-service, tenant-admin-service)"
echo "  2. Frontend services (saas-admin-frontend, tenant-admin-frontend)"
echo ""

# Check prerequisites
echo -e "${BLUE}Checking prerequisites...${NC}"
echo ""

# Check Node.js
echo -n "Node.js version: "
if command -v node > /dev/null 2>&1; then
    NODE_VERSION=$(node -v)
    echo -e "${GREEN}$NODE_VERSION${NC}"

    NODE_MAJOR=$(echo $NODE_VERSION | cut -d'v' -f2 | cut -d'.' -f1)
    if [ "$NODE_MAJOR" -lt 18 ]; then
        echo -e "${RED}Error: Node.js 18+ is required${NC}"
        exit 1
    fi
else
    echo -e "${RED}Not installed${NC}"
    exit 1
fi

# Check Go
echo -n "Go version: "
if command -v go > /dev/null 2>&1; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}$GO_VERSION${NC}"
else
    echo -e "${RED}Not installed${NC}"
    exit 1
fi

# Check PostgreSQL
echo -n "PostgreSQL: "
if pg_isready -h localhost -p 5432 > /dev/null 2>&1; then
    echo -e "${GREEN}Running${NC}"
else
    echo -e "${RED}Not running${NC}"
    echo ""
    echo "Start PostgreSQL with:"
    echo "  brew services start postgresql@16"
    echo ""
    exit 1
fi

# Check RabbitMQ (optional)
echo -n "RabbitMQ: "
if curl -s http://localhost:15672 > /dev/null 2>&1; then
    echo -e "${GREEN}Running${NC}"
else
    echo -e "${YELLOW}Not running (will fall back to HTTP sync)${NC}"
fi

echo ""
echo -e "${BLUE}Starting backend services...${NC}"
echo ""

# Start backends
./start-all-backends.sh

echo ""
echo -e "${BLUE}Starting frontend services...${NC}"
echo ""

# Start frontends
./start-all-frontends.sh

echo ""
echo -e "${BOLD}========================================"
echo "All Services Running"
echo "========================================${NC}"
echo ""
echo -e "${GREEN}Backend Services:${NC}"
echo "  - SaaS Admin Backend:     http://localhost:8098/api/v1/health"
echo "  - Tenant Admin Backend:   http://localhost:8099/health"
echo ""
echo -e "${GREEN}Frontend Applications:${NC}"
echo "  - SaaS Admin Dashboard:   http://localhost:3001"
echo "  - Tenant Admin Dashboard: http://{subdomain}.localhost:3002"
echo "    (e.g., http://anupam.localhost:3002)"
echo ""
echo -e "${YELLOW}Logs:${NC}"
echo "  Backend logs:"
echo "    tail -f /tmp/saas-admin-service.log"
echo "    tail -f /tmp/tenant-admin-service.log"
echo ""
echo "  Frontend logs:"
echo "    tail -f /tmp/saas-admin-frontend.log"
echo "    tail -f /tmp/tenant-admin-frontend.log"
echo ""
echo -e "${YELLOW}To stop all services:${NC}"
echo "  ./stop-all-services.sh"
echo ""
