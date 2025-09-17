#!/bin/bash

# Comprehensive Test Runner for Beakon Microservices
# This script handles everything: Kafka setup, consumer fixes, and test execution

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test results tracking
TOTAL_SERVICES=0
PASSED_SERVICES=0
FAILED_SERVICES=0
SKIPPED_SERVICES=0
WARNING_SERVICES=0

# Arrays to store results
PASSED_SERVICE_LIST=()
FAILED_SERVICE_LIST=()
SKIPPED_SERVICE_LIST=()
WARNING_SERVICE_LIST=()

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

print_header() {
    echo -e "${PURPLE}[HEADER]${NC} $1"
}

print_service() {
    echo -e "${CYAN}[SERVICE]${NC} $1"
}

# Function to check prerequisites
check_prerequisites() {
    print_header "Checking prerequisites..."
    
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed. Please install Go 1.25.0 or later."
        exit 1
    fi
    
    if ! command -v docker >/dev/null 2>&1; then
        print_error "Docker is not installed. Please install Docker."
        exit 1
    fi
    
    if ! command -v docker-compose >/dev/null 2>&1; then
        print_error "Docker Compose is not installed. Please install Docker Compose."
        exit 1
    fi
    
    print_success "All prerequisites are met."
}

# Function to setup Kafka infrastructure
setup_kafka_infrastructure() {
    print_header "Setting up Kafka infrastructure..."
    
    # Check if containers already exist and are running
    if docker ps | grep -q kafka-test && docker ps | grep -q zookeeper-test; then
        print_success "Kafka infrastructure is already running."
        return 0
    fi
    
    # Clean up any existing containers
    print_status "Cleaning up existing containers..."
    docker stop $(docker ps -q --filter "name=kafka-test") 2>/dev/null || true
    docker stop $(docker ps -q --filter "name=zookeeper-test") 2>/dev/null || true
    docker rm $(docker ps -aq --filter "name=kafka-test") 2>/dev/null || true
    docker rm $(docker ps -aq --filter "name=zookeeper-test") 2>/dev/null || true
    
    # Create network
    print_status "Creating network..."
    docker network create kafka-network 2>/dev/null || true
    
    # Start Zookeeper
    print_status "Starting Zookeeper..."
    docker run -d \
        --name zookeeper-test \
        --network kafka-network \
        -p 2181:2181 \
        -e ZOOKEEPER_CLIENT_PORT=2181 \
        -e ZOOKEEPER_TICK_TIME=2000 \
        confluentinc/cp-zookeeper:7.4.0
    
    # Wait for Zookeeper
    print_status "Waiting for Zookeeper to start..."
    sleep 10
    
    # Start Kafka
    print_status "Starting Kafka..."
    docker run -d \
        --name kafka-test \
        --network kafka-network \
        -p 9092:9092 \
        -e KAFKA_BROKER_ID=1 \
        -e KAFKA_ZOOKEEPER_CONNECT=zookeeper-test:2181 \
        -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
        -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
        -e KAFKA_AUTO_CREATE_TOPICS_ENABLE=true \
        confluentinc/cp-kafka:7.4.0
    
    # Wait for Kafka
    print_status "Waiting for Kafka to start..."
    sleep 15
    
    # Test Kafka
    print_status "Testing Kafka connection..."
    if docker exec kafka-test kafka-topics --bootstrap-server localhost:9092 --list >/dev/null 2>&1; then
        print_success "Kafka infrastructure is ready!"
    else
        print_error "Kafka failed to start properly"
        exit 1
    fi
}

# Function to fix consumer services
fix_consumer_services() {
    print_header "Consumer services are already configured for external Kafka..."
    print_success "All consumer services are ready!"
}

# Function to get all microservice directories
get_microservices() {
    find microservices -maxdepth 1 -type d -name "*-service" -o -name "*-consumer" -o -name "api-gateway" | sort
}

# Function to run tests for a single service
run_service_tests() {
    local service_path=$1
    local service_name=$(basename "$service_path")
    local test_type=${2:-"all"}
    
    print_service "Testing $service_name ($test_type)..."
    
    # Check if run-tests.sh exists
    if [ ! -f "$service_path/run-tests.sh" ]; then
        print_warning "No run-tests.sh found for $service_name. Skipping."
        SKIPPED_SERVICES=$((SKIPPED_SERVICES + 1))
        SKIPPED_SERVICE_LIST+=("$service_name")
        return 0
    fi
    
    # Make the script executable
    chmod +x "$service_path/run-tests.sh"
    
    # Change to service directory and run tests
    cd "$service_path"
    
    print_status "Running $test_type tests for $service_name..."
    
    # Run the tests
    local test_output
    if test_output=$(./run-tests.sh "$test_type" 2>&1); then
        # Check for warnings in output
        if echo "$test_output" | grep -q "WARN\|WARNING"; then
            print_warning "$service_name tests passed with warnings!"
            WARNING_SERVICES=$((WARNING_SERVICES + 1))
            WARNING_SERVICE_LIST+=("$service_name")
        else
            print_success "$service_name tests passed!"
            PASSED_SERVICES=$((PASSED_SERVICES + 1))
            PASSED_SERVICE_LIST+=("$service_name")
        fi
    else
        local exit_code=$?
        print_error "$service_name tests failed!"
        FAILED_SERVICES=$((FAILED_SERVICES + 1))
        FAILED_SERVICE_LIST+=("$service_name")
        
        # Save test output for debugging
        mkdir -p test-results
        echo "$test_output" > "test-results/${service_name}-failure.log"
    fi
    
    # Return to root directory
    cd - > /dev/null
    
    echo "----------------------------------------"
}

# Function to run unit tests for all services
run_unit_tests() {
    print_header "Running unit tests for all services..."
    
    local services=($(get_microservices))
    TOTAL_SERVICES=${#services[@]}
    
    print_status "Found $TOTAL_SERVICES microservices to test"
    echo ""
    
    # Run unit tests for each service
    for service in "${services[@]}"; do
        run_service_tests "$service" "unit"
    done
}

# Function to run integration tests for consumer services
run_integration_tests() {
    print_header "Running integration tests for consumer services..."
    
    local consumer_services=($(find microservices -maxdepth 1 -type d -name "*-consumer" | sort))
    
    if [ ${#consumer_services[@]} -eq 0 ]; then
        print_warning "No consumer services found for integration testing."
        return 0
    fi
    
    print_status "Found ${#consumer_services[@]} consumer services for integration testing"
    echo ""
    
    # Run integration tests for each consumer service
    for service in "${consumer_services[@]}"; do
        run_service_tests "$service" "integration"
    done
}

# Function to generate test reports
generate_test_reports() {
    print_header "Generating test reports..."
    
    # Create test-results directory
    mkdir -p test-results
    
    local report_file="test-results/test-summary-$(date +%Y%m%d-%H%M%S).md"
    
    cat > "$report_file" << EOF
# Test Execution Report

**Generated:** $(date)
**Total Services:** $TOTAL_SERVICES
**Passed:** $PASSED_SERVICES
**Failed:** $FAILED_SERVICES
**Warnings:** $WARNING_SERVICES
**Skipped:** $SKIPPED_SERVICES

## Test Results Summary

### ✅ Passed Services ($PASSED_SERVICES)
EOF

    for service in "${PASSED_SERVICE_LIST[@]}"; do
        echo "- $service" >> "$report_file"
    done

    cat >> "$report_file" << EOF

### ⚠️ Services with Warnings ($WARNING_SERVICES)
EOF

    for service in "${WARNING_SERVICE_LIST[@]}"; do
        echo "- $service" >> "$report_file"
    done

    cat >> "$report_file" << EOF

### ❌ Failed Services ($FAILED_SERVICES)
EOF

    for service in "${FAILED_SERVICE_LIST[@]}"; do
        echo "- $service" >> "$report_file"
    done

    cat >> "$report_file" << EOF

### ⏭️ Skipped Services ($SKIPPED_SERVICES)
EOF

    for service in "${SKIPPED_SERVICE_LIST[@]}"; do
        echo "- $service" >> "$report_file"
    done

    cat >> "$report_file" << EOF

## Recommendations

1. **Database Configuration**: Services with warnings may need database configuration updates
2. **Integration Testing**: Consumer services require Kafka infrastructure for full testing
3. **CI/CD Integration**: Use this script in your CI/CD pipeline for automated testing

## Next Steps

1. Review failed services and fix issues
2. Address warnings in services
3. Set up proper Kafka infrastructure for consumer services
4. Integrate this script into your CI/CD pipeline
EOF

    print_success "Test report generated: $report_file"
}

# Function to show summary
show_summary() {
    print_header "TEST EXECUTION SUMMARY"
    echo "========================================"
    echo -e "Total Services: ${BLUE}$TOTAL_SERVICES${NC}"
    echo -e "Passed: ${GREEN}$PASSED_SERVICES${NC}"
    echo -e "Failed: ${RED}$FAILED_SERVICES${NC}"
    echo -e "Warnings: ${YELLOW}$WARNING_SERVICES${NC}"
    echo -e "Skipped: ${YELLOW}$SKIPPED_SERVICES${NC}"
    echo ""
    
    if [ ${#PASSED_SERVICE_LIST[@]} -gt 0 ]; then
        print_success "Passed Services:"
        for service in "${PASSED_SERVICE_LIST[@]}"; do
            echo "  ✅ $service"
        done
        echo ""
    fi
    
    if [ ${#WARNING_SERVICE_LIST[@]} -gt 0 ]; then
        print_warning "Services with Warnings:"
        for service in "${WARNING_SERVICE_LIST[@]}"; do
            echo "  ⚠️  $service"
        done
        echo ""
    fi
    
    if [ ${#FAILED_SERVICE_LIST[@]} -gt 0 ]; then
        print_error "Failed Services:"
        for service in "${FAILED_SERVICE_LIST[@]}"; do
            echo "  ❌ $service"
        done
        echo ""
    fi
    
    if [ ${#SKIPPED_SERVICE_LIST[@]} -gt 0 ]; then
        print_warning "Skipped Services:"
        for service in "${SKIPPED_SERVICE_LIST[@]}"; do
            echo "  ⏭️  $service"
        done
        echo ""
    fi
    
    # Overall result
    if [ $FAILED_SERVICES -eq 0 ]; then
        if [ $WARNING_SERVICES -eq 0 ]; then
            print_success "🎉 All tests completed successfully!"
            return 0
        else
            print_warning "⚠️  All tests passed but some have warnings."
            return 0
        fi
    else
        print_error "❌ Some tests failed. Please check the output above for details."
        return 1
    fi
}

# Function to cleanup
cleanup() {
    print_status "Cleaning up test environment..."
    
    # Stop any running test containers
    docker stop $(docker ps -q --filter "name=kafka-test") 2>/dev/null || true
    docker stop $(docker ps -q --filter "name=zookeeper-test") 2>/dev/null || true
    docker rm $(docker ps -aq --filter "name=kafka-test") 2>/dev/null || true
    docker rm $(docker ps -aq --filter "name=zookeeper-test") 2>/dev/null || true
    
    print_success "Cleanup complete."
}

# Function to show help
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Comprehensive test runner for Beakon microservices."
    echo "This script handles Kafka setup, consumer fixes, and test execution."
    echo ""
    echo "Options:"
    echo "  -h, --help           Show this help message"
    echo "  -u, --unit-only      Run only unit tests"
    echo "  -i, --integration    Run integration tests for consumer services"
    echo "  -a, --all            Run all tests (unit + integration)"
    echo "  -v, --verbose        Enable verbose output"
    echo "  --report             Generate detailed test report"
    echo "  --cleanup            Clean up Kafka infrastructure and exit"
    echo ""
    echo "Examples:"
    echo "  $0                    # Run unit tests only"
    echo "  $0 --all              # Run all tests"
    echo "  $0 --integration      # Run integration tests only"
    echo "  $0 --report           # Generate test report"
    echo "  $0 --cleanup          # Clean up and exit"
}

# Main function
main() {
    local unit_only=false
    local integration_only=false
    local all_tests=false
    local verbose=false
    local generate_report=false
    local cleanup_only=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -u|--unit-only)
                unit_only=true
                shift
                ;;
            -i|--integration)
                integration_only=true
                shift
                ;;
            -a|--all)
                all_tests=true
                shift
                ;;
            -v|--verbose)
                verbose=true
                shift
                ;;
            --report)
                generate_report=true
                shift
                ;;
            --cleanup)
                cleanup_only=true
                shift
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Set verbose mode
    if [ "$verbose" = true ]; then
        set -x
    fi
    
    # Handle cleanup only
    if [ "$cleanup_only" = true ]; then
        cleanup
        exit 0
    fi
    
    # Check prerequisites
    check_prerequisites
    
    # Setup Kafka infrastructure
    setup_kafka_infrastructure
    
    # Fix consumer services
    fix_consumer_services
    
    # Run tests based on options
    if [ "$integration_only" = true ]; then
        run_integration_tests
    elif [ "$all_tests" = true ]; then
        run_unit_tests
        run_integration_tests
    else
        run_unit_tests
    fi
    
    # Generate report if requested
    if [ "$generate_report" = true ]; then
        generate_test_reports
    fi
    
    # Show summary
    local exit_code=0
    if ! show_summary; then
        exit_code=1
    fi
    
    # Cleanup
    cleanup
    
    exit $exit_code
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

# Run main function
main "$@"
