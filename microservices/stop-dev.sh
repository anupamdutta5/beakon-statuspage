#!/bin/bash

echo "🛑 Stopping Beakon Services"
echo "==========================="

# Kill all Go processes running our services
pids=$(pgrep -f "go run.*cmd" 2>/dev/null || true)

if [ -n "$pids" ]; then
    echo "Stopping services with PIDs: $pids"

    # Send TERM signal first
    for pid in $pids; do
        kill -TERM "$pid" 2>/dev/null || true
    done

    sleep 3

    # Force kill any remaining
    for pid in $pids; do
        if kill -0 "$pid" 2>/dev/null; then
            kill -9 "$pid" 2>/dev/null || true
        fi
    done

    echo "✅ Services stopped"
else
    echo "ℹ️  No running services found"
fi

# Clean up ports
ports="8080 8081 8082 8083 8084 8094 8095"
for port in $ports; do
    pid=$(lsof -ti:$port 2>/dev/null || true)
    if [ -n "$pid" ]; then
        echo "Cleaning up port $port (PID: $pid)"
        kill -9 "$pid" 2>/dev/null || true
    fi
done

echo "🎉 All services stopped!"