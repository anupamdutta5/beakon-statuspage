#!/bin/bash

# Stop All Services - Complete Stack
# Stops both backend and frontend services

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BOLD='\033[1m'
NC='\033[0m' # No Color

echo ""
echo -e "${BOLD}========================================"
echo "Stopping All Services"
echo "========================================${NC}"
echo ""

# Stop frontends
echo "Stopping frontend services..."
./stop-all-frontends.sh

echo ""

# Stop backends
echo "Stopping backend services..."
./stop-all-backends.sh

echo ""
echo -e "${GREEN}All services stopped${NC}"
echo ""
