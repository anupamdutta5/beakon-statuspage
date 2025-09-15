#!/bin/bash

# Billing Consumer Test Runner
# This script runs tests for the billing consumer service

set -e

echo "🧪 Starting Billing Consumer Tests..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Build the test Docker image
print_status "Building test Docker image..."
docker build -f Dockerfile.test -t billing-consumer-test .

if [ $? -ne 0 ]; then
    print_error "Failed to build test Docker image"
    exit 1
fi

# Start test dependencies (only PostgreSQL, use external Kafka)
print_status "Starting test dependencies (PostgreSQL)..."
docker-compose -f docker-compose.test.yml up -d postgres

# Wait for services to be ready
print_status "Waiting for services to be ready..."
sleep 10

# Check if PostgreSQL is ready
print_status "Checking PostgreSQL connection..."
until docker-compose -f docker-compose.test.yml exec postgres pg_isready -U test_user -d billing_consumer_test; do
    print_warning "Waiting for PostgreSQL..."
    sleep 2
done

# Check if external Kafka is ready
print_status "Checking external Kafka connection..."
if docker exec kafka-test kafka-topics --bootstrap-server localhost:9092 --list >/dev/null 2>&1; then
    print_status "External Kafka is ready"
else
    print_error "External Kafka is not available. Please run: ./quick-kafka-fix.sh"
    docker-compose -f docker-compose.test.yml down
    exit 1
fi

# Run unit tests
print_status "Running unit tests..."
docker-compose -f docker-compose.test.yml run --rm billing-consumer-test go test -v ./tests/unit/... -coverprofile=unit-coverage.out

if [ $? -ne 0 ]; then
    print_error "Unit tests failed"
    docker-compose -f docker-compose.test.yml down
    exit 1
fi

# Run integration tests
print_status "Running integration tests..."
docker-compose -f docker-compose.test.yml run --rm billing-consumer-test go test -v ./tests/integration/... -coverprofile=integration-coverage.out

if [ $? -ne 0 ]; then
    print_error "Integration tests failed"
    docker-compose -f docker-compose.test.yml down
    exit 1
fi

# Generate coverage report
print_status "Generating coverage report..."
docker-compose -f docker-compose.test.yml run --rm billing-consumer-test go tool cover -html=unit-coverage.out -o unit-coverage.html
docker-compose -f docker-compose.test.yml run --rm billing-consumer-test go tool cover -html=integration-coverage.out -o integration-coverage.html

# Display coverage summary
print_status "Coverage Summary:"
docker-compose -f docker-compose.test.yml run --rm billing-consumer-test go tool cover -func=unit-coverage.out
docker-compose -f docker-compose.test.yml run --rm billing-consumer-test go tool cover -func=integration-coverage.out

# Clean up
print_status "Cleaning up test environment..."
docker-compose -f docker-compose.test.yml down

print_status "✅ All tests completed successfully!"
print_status "Coverage reports generated:"
print_status "  - Unit tests: unit-coverage.html"
print_status "  - Integration tests: integration-coverage.html"

