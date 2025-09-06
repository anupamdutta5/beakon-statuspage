#!/bin/bash

# Test runner script for the status page application

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

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if Docker is running
check_docker() {
    if ! command_exists docker; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    if ! docker info >/dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker first."
        exit 1
    fi
}

# Function to check if Docker Compose is available
check_docker_compose() {
    if ! command_exists docker-compose && ! docker compose version >/dev/null 2>&1; then
        print_error "Docker Compose is not available. Please install Docker Compose first."
        exit 1
    fi
}

# Function to start test database
start_test_db() {
    print_status "Starting test database..."
    
    if command_exists docker-compose; then
        docker-compose -f docker-compose.test.yml up -d
    else
        docker compose -f docker-compose.test.yml up -d
    fi
    
    # Wait for database to be ready
    print_status "Waiting for test database to be ready..."
    sleep 10
    
    print_success "Test database started successfully"
}

# Function to stop test database
stop_test_db() {
    print_status "Stopping test database..."
    
    if command_exists docker-compose; then
        docker-compose -f docker-compose.test.yml down
    else
        docker compose -f docker-compose.test.yml down
    fi
    
    print_success "Test database stopped successfully"
}

# Function to run unit tests
run_unit_tests() {
    print_status "Running unit tests..."
    
    if ! command_exists go; then
        print_error "Go is not installed. Please install Go first."
        exit 1
    fi
    
    # Run unit tests with coverage
    go test -v -race -coverprofile=coverage.out ./internal/tests/...
    
    # Generate coverage report
    if command_exists go; then
        go tool cover -html=coverage.out -o coverage.html
        print_success "Coverage report generated: coverage.html"
    fi
    
    print_success "Unit tests completed"
}

# Function to run integration tests
run_integration_tests() {
    print_status "Running integration tests..."
    
    # Set test environment variables
    export ENV=test
    export DATABASE_URL="postgres://postgres:password@localhost:5433/statuspage_test?sslmode=disable"
    export JWT_SECRET="test-secret-key"
    export ADMIN_EMAIL="admin@test.com"
    export ADMIN_PASSWORD="testpassword"
    
    # Run integration tests
    go test -v -race -tags=integration ./internal/tests/...
    
    print_success "Integration tests completed"
}

# Function to run load tests
run_load_tests() {
    print_status "Running load tests..."
    
    if ! command_exists k6; then
        print_warning "k6 is not installed. Skipping load tests."
        print_warning "To install k6: https://k6.io/docs/getting-started/installation/"
        return
    fi
    
    # Run load tests
    k6 run tests/load/status_page_load_test.js
    
    print_success "Load tests completed"
}

# Function to run security tests
run_security_tests() {
    print_status "Running security tests..."
    
    if ! command_exists gosec; then
        print_warning "gosec is not installed. Skipping security tests."
        print_warning "To install gosec: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
        return
    fi
    
    # Run security tests
    gosec ./...
    
    print_success "Security tests completed"
}

# Function to run all tests
run_all_tests() {
    print_status "Running all tests..."
    
    # Start test database
    start_test_db
    
    # Run unit tests
    run_unit_tests
    
    # Run integration tests
    run_integration_tests
    
    # Run load tests
    run_load_tests
    
    # Run security tests
    run_security_tests
    
    # Stop test database
    stop_test_db
    
    print_success "All tests completed successfully!"
}

# Function to show help
show_help() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  unit        Run unit tests only"
    echo "  integration Run integration tests only"
    echo "  load        Run load tests only"
    echo "  security    Run security tests only"
    echo "  all         Run all tests (default)"
    echo "  help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 unit"
    echo "  $0 integration"
    echo "  $0 all"
}

# Main script logic
main() {
    # Check prerequisites
    check_docker
    check_docker_compose
    
    # Parse command line arguments
    case "${1:-all}" in
        "unit")
            run_unit_tests
            ;;
        "integration")
            start_test_db
            run_integration_tests
            stop_test_db
            ;;
        "load")
            start_test_db
            run_load_tests
            stop_test_db
            ;;
        "security")
            run_security_tests
            ;;
        "all")
            run_all_tests
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
