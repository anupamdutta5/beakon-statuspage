#!/bin/bash

# setup-test-env.sh
# Start all test infrastructure (PostgreSQL, Redis, RabbitMQ, MockServer)
# Usage: ./scripts/test/setup-test-env.sh

set -e

echo "=== Beakon Test Environment Setup ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}Error: Docker is not running. Please start Docker and try again.${NC}"
    exit 1
fi

echo -e "${YELLOW}Step 1: Stopping any existing test containers...${NC}"
docker-compose -f docker-compose.test.yml down -v 2>/dev/null || true

echo -e "${YELLOW}Step 2: Starting test infrastructure...${NC}"
docker-compose -f docker-compose.test.yml up -d

echo -e "${YELLOW}Step 3: Waiting for services to be healthy...${NC}"

# Wait for PostgreSQL
echo -n "Waiting for PostgreSQL..."
until docker exec beakon-test-postgres pg_isready -U postgres > /dev/null 2>&1; do
    echo -n "."
    sleep 1
done
echo -e " ${GREEN}✓${NC}"

# Wait for Redis
echo -n "Waiting for Redis..."
until docker exec beakon-test-redis redis-cli ping > /dev/null 2>&1; do
    echo -n "."
    sleep 1
done
echo -e " ${GREEN}✓${NC}"

# Wait for RabbitMQ
echo -n "Waiting for RabbitMQ..."
until docker exec beakon-test-rabbitmq rabbitmq-diagnostics ping > /dev/null 2>&1; do
    echo -n "."
    sleep 1
done
echo -e " ${GREEN}✓${NC}"

# Wait for MockServer
echo -n "Waiting for MockServer..."
until curl -s http://localhost:1080/mockserver/status > /dev/null 2>&1; do
    echo -n "."
    sleep 1
done
echo -e " ${GREEN}✓${NC}"

echo ""
echo -e "${GREEN}=== Test Environment Ready! ===${NC}"
echo ""
echo "Services available:"
echo "  PostgreSQL:   localhost:5432 (user: postgres, pass: testpass)"
echo "  Redis:        localhost:6379"
echo "  RabbitMQ:     localhost:5672 (user: test, pass: testpass)"
echo "  RabbitMQ UI:  http://localhost:15672"
echo "  MockServer:   http://localhost:1080"
echo "  Prometheus:   http://localhost:9090"
echo "  Grafana:      http://localhost:3000 (user: admin, pass: admin)"
echo ""
echo "To stop the environment:"
echo "  ./scripts/test/teardown-test-env.sh"
echo ""
