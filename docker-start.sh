#!/bin/bash

# Beakon Microservices - Docker Startup Script
# This script starts all 19 microservices using Docker Compose

set -e

echo "🚀 Starting Beakon Microservices Platform"
echo "=========================================="

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

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose >/dev/null 2>&1; then
    print_error "Docker Compose is not installed. Please install Docker Compose and try again."
    exit 1
fi

print_status "Docker is running ✓"
print_status "Docker Compose is available ✓"

# Stop any existing containers
print_status "Stopping any existing containers..."
docker-compose down --remove-orphans 2>/dev/null || true

# Clean up previous builds (optional)
if [ "$1" = "--clean" ]; then
    print_warning "Cleaning up previous builds..."
    docker-compose down --volumes --remove-orphans
    docker system prune -f
fi

print_status "Starting all microservices..."

# Start the infrastructure services first
print_status "Step 1: Starting infrastructure services (PostgreSQL, Redis, Kafka)..."
docker-compose up -d postgres redis zookeeper kafka

# Wait for infrastructure to be ready
print_status "Waiting for infrastructure services to be ready..."
sleep 10

# Check if PostgreSQL is ready
print_status "Checking PostgreSQL connection..."
timeout=30
while ! docker-compose exec -T postgres pg_isready -U statuspage_user >/dev/null 2>&1; do
    if [ $timeout -le 0 ]; then
        print_error "PostgreSQL failed to start within timeout"
        exit 1
    fi
    sleep 1
    ((timeout--))
done
print_success "PostgreSQL is ready"

# Check if Redis is ready
print_status "Checking Redis connection..."
timeout=30
while ! docker-compose exec -T redis redis-cli ping >/dev/null 2>&1; do
    if [ $timeout -le 0 ]; then
        print_error "Redis failed to start within timeout"
        exit 1
    fi
    sleep 1
    ((timeout--))
done
print_success "Redis is ready"

# Start core services
print_status "Step 2: Starting core services..."
docker-compose up -d user-service database-service

# Wait for core services
sleep 5

# Start remaining services
print_status "Step 3: Starting remaining microservices..."
docker-compose up -d

print_success "All microservices are starting up!"

echo ""
echo "🎉 Beakon Platform Started Successfully!"
echo "========================================"
echo ""
echo "📋 Service Status:"
echo "   • Total Services: 19 microservices + 4 infrastructure"
echo "   • PostgreSQL:     http://localhost:5432"
echo "   • Redis:          http://localhost:6379"
echo "   • Kafka:          http://localhost:9092"
echo ""
echo "🌐 Main Services:"
echo "   • User Service:        http://localhost:8080"
echo "   • API Gateway:         http://localhost:8081"
echo "   • Component Service:   http://localhost:8082"
echo "   • Analytics Service:   http://localhost:8083"
echo "   • Incident Service:    http://localhost:8084"
echo "   • Notification Service: http://localhost:8086"
echo "   • Payment Service:     http://localhost:8087"
echo "   • SaaS Admin:          http://localhost:8089"
echo "   • Tenant Admin:        http://localhost:8090"
echo "   • Monitoring:          http://localhost:8091"
echo "   • Branding:            http://localhost:8092"
echo "   • Event Store:         http://localhost:8093"
echo "   • Status UI:           http://localhost:8094"
echo "   • Database Service:    http://localhost:8095"
echo "   • Landing Page:        http://localhost:8096"
echo "   • Analytics Consumer:  http://localhost:8097"
echo "   • Audit Consumer:      http://localhost:8098"
echo "   • Notification Consumer: http://localhost:8099"
echo "   • Billing Consumer:    http://localhost:8100"
echo ""
echo "📊 Monitoring:"
echo "   • Docker logs: docker-compose logs -f [service-name]"
echo "   • All logs:    docker-compose logs -f"
echo "   • Status:      docker-compose ps"
echo ""
echo "🛑 To stop all services: docker-compose down"
echo "🧹 To stop and clean:    docker-compose down --volumes"
echo ""
print_success "Happy coding! 🚀"