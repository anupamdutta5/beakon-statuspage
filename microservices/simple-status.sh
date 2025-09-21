#!/bin/bash

echo "🔍 Beakon Service Status"
echo "========================"

# Check Go
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed"
    exit 1
fi

echo "✅ Go: $(go version | cut -d' ' -f3)"

echo ""
echo "Service Status:"

# Check each service port
services="database-service:8095 user-service:8080 api-gateway:8081 component-service:8082 analytics-service:8083 incident-service:8084 status-ui-service:8094"

running=0
total=0

for service_port in $services; do
    service=$(echo $service_port | cut -d: -f1)
    port=$(echo $service_port | cut -d: -f2)
    total=$((total + 1))

    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo "✅ $service (port $port)"
        running=$((running + 1))
    else
        echo "❌ $service (port $port)"
    fi
done

echo ""
echo "Summary: $running/$total services running"

if [ $running -gt 0 ]; then
    echo ""
    echo "🌐 Available URLs:"
    if lsof -Pi :8081 -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo "• API Gateway:  http://localhost:8081"
    fi
    if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo "• User Service: http://localhost:8080/api/v1"
    fi
    if lsof -Pi :8094 -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo "• Status Page:  http://localhost:8094"
    fi
fi

echo ""
echo "📝 Commands:"
echo "• Start:      ./simple-start.sh"
echo "• Stop:       ./simple-stop.sh"
echo "• View logs:  tail -f logs/<service>.log"