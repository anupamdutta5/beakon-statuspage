#!/bin/bash
# Test Service Health - Comprehensive health check for all microservices
# Usage: ./test-service-health.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Service definitions: name:port:path
services=(
    "API Gateway:8080:/health"
    "User Service:8090:/health"
    "Tenant Admin Service:8091:/health"
    "SaaS Admin Service:8092:/api/v1/health"
    "Component Service:8093:/health"
    "Incident Service:8094:/health"
    "Monitoring Service:8095:/health"
    "Analytics Service:8096:/health"
    "Notification Service:8097:/health"
    "Landing Page Service:8098:/health"
    "Payment Service:8099:/health"
    "Branding Service:8100:/health"
    "Database Service:8101:/health"
    "Event Store Service:8102:/health"
    "Status UI Service:8103:/health"
)

echo "🔍 Testing Microservice Health Status..."
echo "============================================="

total_services=${#services[@]}
healthy_services=0
failed_services=()

for service in "${services[@]}"; do
    IFS=':' read -r name port path <<< "$service"

    echo -n "Testing $name (port $port)... "

    # Test with timeout
    if timeout 10s curl -s -f "http://localhost:$port$path" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Healthy${NC}"
        ((healthy_services++))
    else
        echo -e "${RED}❌ Unhealthy${NC}"
        failed_services+=("$name")
    fi
done

echo
echo "============================================="
echo "Health Check Summary:"
echo -e "Total Services: $total_services"
echo -e "${GREEN}Healthy: $healthy_services${NC}"
echo -e "${RED}Failed: $((total_services - healthy_services))${NC}"

if [ ${#failed_services[@]} -gt 0 ]; then
    echo
    echo -e "${RED}Failed Services:${NC}"
    for service in "${failed_services[@]}"; do
        echo -e "${RED}  - $service${NC}"
    done

    echo
    echo -e "${YELLOW}💡 Troubleshooting Tips:${NC}"
    echo "1. Check if services are running: docker ps"
    echo "2. Check service logs: docker logs <container_name>"
    echo "3. Verify port availability: netstat -tlnp | grep <port>"
    echo "4. Check environment variables and configuration"
    exit 1
else
    echo
    echo -e "${GREEN}🎉 All services are healthy!${NC}"
fi