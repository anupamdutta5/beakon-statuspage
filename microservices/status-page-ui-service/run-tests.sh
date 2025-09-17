#!/bin/bash

# Status Page UI Service Test Runner
# This script runs tests for the status page UI service

set -e

echo "🧪 Starting Status Page UI Service Tests..."

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
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if test type is provided
TEST_TYPE=${1:-"unit"}

print_status "Running $TEST_TYPE tests..."

# Start test dependencies (PostgreSQL)
print_status "Starting test dependencies (PostgreSQL)..."
docker-compose -f docker-compose.test.yml up -d postgres

# Wait for services to be ready
print_status "Waiting for services to be ready..."
sleep 10

# Check if PostgreSQL is ready
print_status "Checking PostgreSQL connection..."
until docker-compose -f docker-compose.test.yml exec postgres pg_isready -U test_user -d status_page_ui_test; do
    print_warning "Waiting for PostgreSQL..."
    sleep 2
done

print_status "PostgreSQL is ready!"

# Run tests based on type
case $TEST_TYPE in
    "unit")
        print_status "Running unit tests..."
        go test -v ./internal/... -coverprofile=coverage.out
        ;;
    "integration")
        print_status "Running integration tests..."
        docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit status-page-ui-test
        ;;
    "all")
        print_status "Running all tests..."
        go test -v ./... -coverprofile=coverage.out
        ;;
    *)
        print_error "Unknown test type: $TEST_TYPE"
        print_status "Available test types: unit, integration, all"
        exit 1
        ;;
esac

# Generate coverage report
if [ -f coverage.out ]; then
    print_status "Generating coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    print_status "Coverage report generated: coverage.html"
fi

# Cleanup
print_status "Cleaning up test dependencies..."
docker-compose -f docker-compose.test.yml down

print_status "Tests completed successfully! 🎉"




