#!/bin/bash

# Script to stop all microservices

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_status "Stopping Status Page Microservices"
print_status "=================================="

# List of services to stop
services=(
    "api-gateway"
    "user-service"
    "tenant-service"
    "component-service"
    "incident-service"
    "notification-service"
    "payment-service"
    "analytics-service"
    "monitoring-service"
    "notification-consumer"
    "analytics-consumer"
    "audit-consumer"
    "billing-consumer"
)

# Stop each service
for service in "${services[@]}"; do
    if [ -f "logs/${service}.pid" ]; then
        local pid=$(cat "logs/${service}.pid")
        if kill -0 $pid 2>/dev/null; then
            print_status "Stopping $service (PID: $pid)..."
            kill $pid
            
            # Wait for graceful shutdown
            local count=0
            while kill -0 $pid 2>/dev/null && [ $count -lt 10 ]; do
                sleep 1
                count=$((count + 1))
            done
            
            # Force kill if still running
            if kill -0 $pid 2>/dev/null; then
                print_warning "Force killing $service..."
                kill -9 $pid
            fi
            
            print_success "$service stopped"
        else
            print_warning "$service was not running"
        fi
        rm -f "logs/${service}.pid"
    else
        print_warning "PID file not found for $service"
    fi
done

# Kill any remaining processes on our ports
print_status "Cleaning up any remaining processes on microservice ports..."

ports=(8080 8081 8082 8083 8084 8085 8086 8087 8088 8089 8090 8091 8092)

for port in "${ports[@]}"; do
    local pid=$(lsof -ti:$port 2>/dev/null || true)
    if [ ! -z "$pid" ]; then
        print_status "Killing process on port $port (PID: $pid)..."
        kill -9 $pid 2>/dev/null || true
    fi
done

print_success "All microservices stopped successfully!"
print_status "=================================="
