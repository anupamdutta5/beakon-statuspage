#!/bin/bash

# Master script to start all microservices
# This script starts all microservices in the correct order with proper dependencies

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

# Function to check if a port is available
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 1
    else
        return 0
    fi
}

# Function to wait for a service to be ready
wait_for_service() {
    local service_name=$1
    local port=$2
    local max_attempts=30
    local attempt=1

    print_status "Waiting for $service_name to be ready on port $port..."
    
    while [ $attempt -le $max_attempts ]; do
        if check_port $port; then
            sleep 1
            attempt=$((attempt + 1))
        else
            print_success "$service_name is ready on port $port"
            return 0
        fi
    done
    
    print_error "$service_name failed to start on port $port after $max_attempts attempts"
    return 1
}

# Function to start a service
start_service() {
    local service_name=$1
    local service_path=$2
    local port=$3
    local start_script=$4

    print_status "Starting $service_name..."

    if [ ! -d "$service_path" ]; then
        print_error "Service directory not found: $service_path"
        return 1
    fi

    if [ ! -f "$service_path/$start_script" ]; then
        print_error "Start script not found: $service_path/$start_script"
        return 1
    fi

    # Check if port is already in use
    if ! check_port $port; then
        print_warning "Port $port is already in use. Skipping $service_name"
        return 0
    fi

    # Start the service in background
    cd "$service_path"
    chmod +x "$start_script"
    nohup ./"$start_script" > "../logs/${service_name}.log" 2>&1 &
    local pid=$!
    echo $pid > "../logs/${service_name}.pid"
    
    print_success "$service_name started with PID $pid"
    return 0
}

# Create logs directory
mkdir -p logs

print_status "Starting Status Page Microservices Architecture"
print_status "================================================"

# Check if required services are available
print_status "Checking service availability..."

# Start services in dependency order
print_status "Starting services in dependency order..."

# 1. Start API Gateway (Port 8080)
start_service "api-gateway" "microservices/api-gateway" 8080 "start.sh"
wait_for_service "api-gateway" 8080

# 2. Start User Service (Port 8081)
start_service "user-service" "microservices/user-service" 8081 "start.sh"
wait_for_service "user-service" 8081

# 3. Start Tenant Service (Port 8082)
start_service "tenant-service" "microservices/tenant-service" 8082 "start.sh"
wait_for_service "tenant-service" 8082

# 4. Start Component Service (Port 8084)
start_service "component-service" "microservices/component-service" 8084 "start.sh"
wait_for_service "component-service" 8084

# 5. Start Incident Service (Port 8086)
start_service "incident-service" "microservices/incident-service" 8086 "start.sh"
wait_for_service "incident-service" 8086

# 6. Start Payment Service (Port 8088)
start_service "payment-service" "microservices/payment-service" 8088 "start.sh"
wait_for_service "payment-service" 8088

# 7. Start Analytics Service (Port 8090)
start_service "analytics-service" "microservices/analytics-service" 8090 "start.sh"
wait_for_service "analytics-service" 8090

# 8. Start Monitoring Service (Port 8092)
start_service "monitoring-service" "microservices/monitoring-service" 8092 "start.sh"
wait_for_service "monitoring-service" 8092

# 9. Start Database Service (Port 8095)
start_service "database-service" "microservices/database-service" 8095 "start.sh"
wait_for_service "database-service" 8095

# 10. Start Event Store Service (Port 8096)
start_service "event-store-service" "microservices/event-store-service" 8096 "start.sh"
wait_for_service "event-store-service" 8096

# 11. Start Branding Service (Port 8097)
start_service "branding-service" "microservices/branding-service" 8097 "start.sh"
wait_for_service "branding-service" 8097

# 12. Start SaaS Admin Service (Port 8098)
start_service "saas-admin-service" "microservices/saas-admin-service" 8098 "start.sh"
wait_for_service "saas-admin-service" 8098

# 13. Start Tenant Admin Service (Port 8099)
start_service "tenant-admin-service" "microservices/tenant-admin-service" 8099 "start.sh"
wait_for_service "tenant-admin-service" 8099

# 14. Start Landing Page Service (Port 8100)
start_service "landing-page-service" "microservices/landing-page-service" 8100 "start.sh"
wait_for_service "landing-page-service" 8100

# 15. Start Event Consumers
print_status "Starting event consumers..."

start_service "notification-consumer" "microservices/notification-consumer" 8089 "start.sh"
start_service "analytics-consumer" "microservices/analytics-consumer" 8091 "start.sh"
start_service "audit-consumer" "microservices/audit-consumer" 8093 "start.sh"
start_service "billing-consumer" "microservices/billing-consumer" 8094 "start.sh"

print_success "All microservices started successfully!"
print_status "================================================"
print_status "Service Status:"
print_status "API Gateway:        http://localhost:8080"
print_status "User Service:       http://localhost:8081"
print_status "Tenant Service:     http://localhost:8082"
print_status "Component Service:  http://localhost:8084"
print_status "Incident Service:   http://localhost:8086"
print_status "Payment Service:    http://localhost:8088"
print_status "Analytics Service:  http://localhost:8090"
print_status "Monitoring Service: http://localhost:8092"
print_status "Database Service:   http://localhost:8095"
print_status "Event Store Service: http://localhost:8096"
print_status "Branding Service:   http://localhost:8097"
print_status "SaaS Admin Service: http://localhost:8098"
print_status "Tenant Admin Service: http://localhost:8099"
print_status "Landing Page Service: http://localhost:8100"
print_status "================================================"
print_status "Logs are available in the 'logs' directory"
print_status "To stop all services, run: ./stop-microservices.sh"
print_status "================================================"

# Keep script running to show status
print_status "Press Ctrl+C to stop all services..."

# Function to cleanup on exit
cleanup() {
    print_status "Stopping all services..."
    
    # Kill all services
    for service in api-gateway user-service tenant-service component-service incident-service notification-service payment-service analytics-service monitoring-service database-service event-store-service branding-service saas-admin-service tenant-admin-service landing-page-service notification-consumer analytics-consumer audit-consumer billing-consumer; do
        if [ -f "logs/${service}.pid" ]; then
            local pid=$(cat "logs/${service}.pid")
            if kill -0 $pid 2>/dev/null; then
                print_status "Stopping $service (PID: $pid)..."
                kill $pid
            fi
            rm -f "logs/${service}.pid"
        fi
    done
    
    print_success "All services stopped"
    exit 0
}

# Set trap for cleanup
trap cleanup SIGINT SIGTERM

# Wait for user interrupt
while true; do
    sleep 1
done
