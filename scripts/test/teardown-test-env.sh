#!/bin/bash

# teardown-test-env.sh
# Stop all test infrastructure and clean up volumes
# Usage: ./scripts/test/teardown-test-env.sh [--keep-volumes]

set -e

echo "=== Beakon Test Environment Teardown ==="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check for --keep-volumes flag
KEEP_VOLUMES=false
if [[ "$1" == "--keep-volumes" ]]; then
    KEEP_VOLUMES=true
fi

echo -e "${YELLOW}Stopping test containers...${NC}"

if [ "$KEEP_VOLUMES" = true ]; then
    docker-compose -f docker-compose.test.yml down
    echo -e "${GREEN}✓ Containers stopped (volumes preserved)${NC}"
else
    docker-compose -f docker-compose.test.yml down -v
    echo -e "${GREEN}✓ Containers stopped and volumes removed${NC}"
fi

echo ""
echo -e "${GREEN}=== Test Environment Cleaned Up ===${NC}"
echo ""

if [ "$KEEP_VOLUMES" = true ]; then
    echo "Volumes were preserved. To remove them:"
    echo "  docker-compose -f docker-compose.test.yml down -v"
fi
