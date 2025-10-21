#!/bin/bash

# Tenant Admin Service Test Runner
# This script runs all tests for the Tenant Admin Service

set -e

echo "🧪 Starting Tenant Admin Service Tests..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="tenant-admin-service"
TEST_TIMEOUT="10m"
COVERAGE_THRESHOLD="80"

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

# Function to cleanup
cleanup() {
    print_status "Cleaning up test environment..."
    
    # Stop any running containers
    docker-compose -f docker-compose.test.yml down -v 2>/dev/null || true
    
    # Remove test images
    docker rmi $(docker images -q "${SERVICE_NAME}_test" 2>/dev/null) 2>/dev/null || true
    
    print_success "Cleanup completed"
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if Docker is running
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    
    # Check if docker-compose is available
    if ! command -v docker-compose >/dev/null 2>&1; then
        print_error "docker-compose is not installed. Please install docker-compose and try again."
        exit 1
    fi
    
    # Check if Go is available
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed. Please install Go and try again."
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Function to run unit tests
run_unit_tests() {
    print_status "Running unit tests..."
    
    # Set test environment variables
    export ENVIRONMENT=testing
    export DB_NAME=statuspage_tenant_admin_test
    
    # Run unit tests
    if go test -v -timeout $TEST_TIMEOUT ./tests/unit/... -coverprofile=coverage/unit_coverage.out; then
        print_success "Unit tests passed"
        return 0
    else
        print_error "Unit tests failed"
        return 1
    fi
}

# Function to run integration tests
run_integration_tests() {
    print_status "Running integration tests..."
    
    # Start test environment
    print_status "Starting test environment..."
    docker-compose -f docker-compose.test.yml up -d postgres redis
    
    # Wait for services to be ready
    print_status "Waiting for services to be ready..."
    sleep 30
    
    # Run integration tests
    if go test -v -timeout $TEST_TIMEOUT ./tests/integration/... -coverprofile=coverage/integration_coverage.out; then
        print_success "Integration tests passed"
        return 0
    else
        print_error "Integration tests failed"
        return 1
    fi
}

# Function to run all tests with Docker
run_docker_tests() {
    print_status "Running tests with Docker..."
    
    # Build and run tests
    if docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit test-runner; then
        print_success "Docker tests passed"
        return 0
    else
        print_error "Docker tests failed"
        return 1
    fi
}

# Function to generate coverage report
generate_coverage_report() {
    print_status "Generating coverage report..."
    
    # Create coverage directory if it doesn't exist
    mkdir -p coverage
    
    # Merge coverage files if they exist
    if [ -f coverage/unit_coverage.out ] && [ -f coverage/integration_coverage.out ]; then
        echo "mode: atomic" > coverage/combined_coverage.out
        tail -n +2 coverage/unit_coverage.out >> coverage/combined_coverage.out
        tail -n +2 coverage/integration_coverage.out >> coverage/combined_coverage.out
    elif [ -f coverage/unit_coverage.out ]; then
        cp coverage/unit_coverage.out coverage/combined_coverage.out
    elif [ -f coverage/integration_coverage.out ]; then
        cp coverage/integration_coverage.out coverage/combined_coverage.out
    fi
    
    # Generate HTML coverage report
    if [ -f coverage/combined_coverage.out ]; then
        go tool cover -html=coverage/combined_coverage.out -o coverage/coverage.html
        print_success "Coverage report generated: coverage/coverage.html"
        
        # Check coverage threshold
        coverage_percent=$(go tool cover -func=coverage/combined_coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        if (( $(echo "$coverage_percent >= $COVERAGE_THRESHOLD" | bc -l) )); then
            print_success "Coverage threshold met: ${coverage_percent}% >= ${COVERAGE_THRESHOLD}%"
        else
            print_warning "Coverage threshold not met: ${coverage_percent}% < ${COVERAGE_THRESHOLD}%"
        fi
    else
        print_warning "No coverage files found"
    fi
}

# Function to run performance tests
run_performance_tests() {
    print_status "Running performance tests..."
    
    # Set test environment variables
    export ENVIRONMENT=testing
    export DB_NAME=statuspage_tenant_admin_test
    
    # Run performance tests
    if go test -v -timeout $TEST_TIMEOUT ./tests/performance/... -bench=. -benchmem; then
        print_success "Performance tests passed"
        return 0
    else
        print_warning "Performance tests failed or not implemented"
        return 1
    fi
}

# Function to run linting
run_linting() {
    print_status "Running linting..."
    
    # Check if golangci-lint is available
    if command -v golangci-lint >/dev/null 2>&1; then
        if golangci-lint run; then
            print_success "Linting passed"
            return 0
        else
            print_error "Linting failed"
            return 1
        fi
    else
        print_warning "golangci-lint not found, skipping linting"
        return 0
    fi
}

# Function to run security scan
run_security_scan() {
    print_status "Running security scan..."
    
    # Check if gosec is available
    if command -v gosec >/dev/null 2>&1; then
        if gosec ./...; then
            print_success "Security scan passed"
            return 0
        else
            print_warning "Security scan found issues"
            return 1
        fi
    else
        print_warning "gosec not found, skipping security scan"
        return 0
    fi
}

# Main function
main() {
    print_status "Starting Tenant Admin Service Test Suite..."
    
    # Parse command line arguments
    case "${1:-all}" in
        "unit")
            run_unit_tests
            ;;
        "integration")
            run_integration_tests
            ;;
        "docker")
            run_docker_tests
            ;;
        "performance")
            run_performance_tests
            ;;
        "lint")
            run_linting
            ;;
        "security")
            run_security_scan
            ;;
        "coverage")
            generate_coverage_report
            ;;
        "all")
            check_prerequisites
            
            # Run all tests
            unit_result=0
            integration_result=0
            docker_result=0
            lint_result=0
            security_result=0
            
            # Run unit tests
            if ! run_unit_tests; then
                unit_result=1
            fi
            
            # Run integration tests
            if ! run_integration_tests; then
                integration_result=1
            fi
            
            # Run Docker tests
            if ! run_docker_tests; then
                docker_result=1
            fi
            
            # Run linting
            if ! run_linting; then
                lint_result=1
            fi
            
            # Run security scan
            if ! run_security_scan; then
                security_result=1
            fi
            
            # Generate coverage report
            generate_coverage_report
            
            # Summary
            print_status "Test Summary:"
            print_status "  Unit Tests: $([ $unit_result -eq 0 ] && echo "PASSED" || echo "FAILED")"
            print_status "  Integration Tests: $([ $integration_result -eq 0 ] && echo "PASSED" || echo "FAILED")"
            print_status "  Docker Tests: $([ $docker_result -eq 0 ] && echo "PASSED" || echo "FAILED")"
            print_status "  Linting: $([ $lint_result -eq 0 ] && echo "PASSED" || echo "FAILED")"
            print_status "  Security Scan: $([ $security_result -eq 0 ] && echo "PASSED" || echo "FAILED")"
            
            # Exit with error if any test failed
            if [ $unit_result -ne 0 ] || [ $integration_result -ne 0 ] || [ $docker_result -ne 0 ] || [ $lint_result -ne 0 ] || [ $security_result -ne 0 ]; then
                print_error "Some tests failed"
                exit 1
            fi
            
            print_success "All tests passed!"
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [command]"
            echo ""
            echo "Commands:"
            echo "  unit        Run unit tests only"
            echo "  integration Run integration tests only"
            echo "  docker      Run tests with Docker"
            echo "  performance Run performance tests"
            echo "  lint        Run linting"
            echo "  security    Run security scan"
            echo "  coverage    Generate coverage report"
            echo "  all         Run all tests (default)"
            echo "  help        Show this help message"
            ;;
        *)
            print_error "Unknown command: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Set up trap for cleanup
trap cleanup EXIT

# Run main function
main "$@"

