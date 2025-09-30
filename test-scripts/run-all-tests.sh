#!/bin/bash
# Master Test Runner - Execute all SaaS Admin Dashboard tests
# Usage: ./run-all-tests.sh [--quick|--full|--services-only]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="$SCRIPT_DIR/logs"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
FULL_LOG="$LOG_DIR/test_run_$TIMESTAMP.log"

# Test mode
TEST_MODE=${1:-"--full"}

# Create log directory
mkdir -p "$LOG_DIR"

# Utility functions
log_header() {
    echo -e "${CYAN}$1${NC}" | tee -a "$FULL_LOG"
    echo "$(printf '=%.0s' {1..80})" | tee -a "$FULL_LOG"
}

log_section() {
    echo | tee -a "$FULL_LOG"
    echo -e "${PURPLE}$1${NC}" | tee -a "$FULL_LOG"
    echo "$(printf '-%.0s' {1..60})" | tee -a "$FULL_LOG"
}

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}" | tee -a "$FULL_LOG"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}" | tee -a "$FULL_LOG"
}

log_failure() {
    echo -e "${RED}❌ $1${NC}" | tee -a "$FULL_LOG"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}" | tee -a "$FULL_LOG"
}

run_test_script() {
    local script_name=$1
    local description=$2
    local optional=${3:-false}
    local script_path="$SCRIPT_DIR/$script_name"

    log_section "Running: $description"

    if [ ! -f "$script_path" ]; then
        if [ "$optional" = "true" ]; then
            log_warning "Test script not found: $script_name (optional)"
            return 0
        else
            log_failure "Test script not found: $script_name"
            return 1
        fi
    fi

    if [ ! -x "$script_path" ]; then
        log_warning "Making script executable: $script_name"
        chmod +x "$script_path"
    fi

    echo "Starting test: $description" >> "$FULL_LOG"
    echo "Script: $script_name" >> "$FULL_LOG"
    echo "Time: $(date)" >> "$FULL_LOG"
    echo "----------------------------------------" >> "$FULL_LOG"

    # Run the test and capture both output and exit code
    if "$script_path" 2>&1 | tee -a "$FULL_LOG"; then
        log_success "$description completed successfully"
        echo "Test result: SUCCESS" >> "$FULL_LOG"
        return 0
    else
        local exit_code=$?
        log_failure "$description failed (exit code: $exit_code)"
        echo "Test result: FAILED (exit code: $exit_code)" >> "$FULL_LOG"
        return $exit_code
    fi
}

check_prerequisites() {
    log_section "Checking Prerequisites"

    # Check required commands
    required_commands=("curl" "jq")
    missing_commands=()

    for cmd in "${required_commands[@]}"; do
        if ! command -v "$cmd" > /dev/null 2>&1; then
            missing_commands+=("$cmd")
        else
            echo -e "  ${GREEN}✓${NC} $cmd is available"
        fi
    done

    if [ ${#missing_commands[@]} -gt 0 ]; then
        log_failure "Missing required commands: ${missing_commands[*]}"
        echo "Please install missing dependencies:"
        for cmd in "${missing_commands[@]}"; do
            case "$cmd" in
                "curl")
                    echo "  - curl: sudo apt-get install curl (Ubuntu) or brew install curl (macOS)"
                    ;;
                "jq")
                    echo "  - jq: sudo apt-get install jq (Ubuntu) or brew install jq (macOS)"
                    ;;
            esac
        done
        return 1
    fi

    # Check if we're in the right directory
    if [ ! -d "microservices" ]; then
        log_warning "Not in project root directory. Tests may fail."
        echo "Current directory: $(pwd)"
        echo "Expected to see: microservices/ directory"
    else
        echo -e "  ${GREEN}✓${NC} In project root directory"
    fi

    # Check if any services are running
    log_info "Checking if any services are already running..."
    running_services=0

    test_ports=(8080 8090 8091 8092 8093 8094 8095 8096)
    for port in "${test_ports[@]}"; do
        if curl -s -f "http://localhost:$port/health" > /dev/null 2>&1 || \
           curl -s -f "http://localhost:$port/api/v1/health" > /dev/null 2>&1; then
            ((running_services++))
        fi
    done

    if [ $running_services -gt 0 ]; then
        echo -e "  ${GREEN}✓${NC} $running_services services detected running"
    else
        log_warning "No services detected. You may need to start services first."
        echo
        echo "To start services:"
        echo "  docker-compose -f docker-compose.microservices.yml up -d"
        echo "  OR"
        echo "  ./microservices/simple-start.sh"
        echo
        echo "Continue anyway? Services may be running on different ports."
        read -p "Continue? (y/N): " -r
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Test run cancelled by user"
            exit 0
        fi
    fi

    return 0
}

show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "OPTIONS:"
    echo "  --quick         Run only essential tests (health + basic functionality)"
    echo "  --full          Run all tests including integration (default)"
    echo "  --services-only Run only service health checks"
    echo "  --help          Show this help message"
    echo
    echo "Test Scripts Executed:"
    echo "  1. test-service-health.sh           - Check all service health"
    echo "  2. test-saas-admin-features.sh      - Test SaaS Admin native features"
    echo "  3. test-inter-service-communication.sh - Test service integration"
    echo "  4. test-dashboard-integration.sh    - End-to-end workflow testing"
    echo
    echo "Logs are saved to: test-scripts/logs/"
}

# Parse command line arguments
case "$TEST_MODE" in
    "--help")
        show_usage
        exit 0
        ;;
    "--quick")
        log_info "Running in QUICK mode (essential tests only)"
        ;;
    "--full")
        log_info "Running in FULL mode (all tests)"
        ;;
    "--services-only")
        log_info "Running SERVICES-ONLY mode (health checks only)"
        ;;
    *)
        log_warning "Unknown option: $TEST_MODE"
        show_usage
        exit 1
        ;;
esac

# Start test execution
log_header "🚀 Beakon SaaS Admin Dashboard - Comprehensive Test Suite"

echo "Test Mode: $TEST_MODE" | tee -a "$FULL_LOG"
echo "Start Time: $(date)" | tee -a "$FULL_LOG"
echo "Log File: $FULL_LOG" | tee -a "$FULL_LOG"
echo | tee -a "$FULL_LOG"

# Check prerequisites
if ! check_prerequisites; then
    log_failure "Prerequisites check failed"
    exit 1
fi

# Initialize test results
declare -A test_results
total_tests=0
passed_tests=0
failed_tests=0

# Test Phase 1: Service Health Checks
log_section "🏥 Phase 1: Service Health Checks"
if run_test_script "test-service-health.sh" "Service Health Checks"; then
    test_results["health"]="PASS"
    ((passed_tests++))
else
    test_results["health"]="FAIL"
    ((failed_tests++))
fi
((total_tests++))

# Exit early for services-only mode
if [ "$TEST_MODE" = "--services-only" ]; then
    log_section "📊 Services-Only Test Summary"
    echo "Health Check: ${test_results["health"]}"
    log_info "Services-only mode completed"
    exit 0
fi

# Test Phase 2: SaaS Admin Features
log_section "🛠️  Phase 2: SaaS Admin Feature Testing"
if run_test_script "test-saas-admin-features.sh" "SaaS Admin Features"; then
    test_results["features"]="PASS"
    ((passed_tests++))
else
    test_results["features"]="FAIL"
    ((failed_tests++))
fi
((total_tests++))

# Exit early for quick mode
if [ "$TEST_MODE" = "--quick" ]; then
    log_section "📊 Quick Test Summary"
    echo "Health Check: ${test_results["health"]}"
    echo "Feature Test: ${test_results["features"]}"

    if [ $failed_tests -eq 0 ]; then
        log_success "Quick tests completed successfully!"
    else
        log_failure "Some quick tests failed"
    fi
    exit $failed_tests
fi

# Test Phase 3: Inter-Service Communication (Full mode only)
log_section "🔗 Phase 3: Inter-Service Communication"
if run_test_script "test-inter-service-communication.sh" "Inter-Service Communication"; then
    test_results["communication"]="PASS"
    ((passed_tests++))
else
    test_results["communication"]="FAIL"
    ((failed_tests++))
fi
((total_tests++))

# Test Phase 4: Dashboard Integration (Full mode only)
log_section "🎯 Phase 4: Dashboard Integration Testing"
if run_test_script "test-dashboard-integration.sh" "Dashboard Integration"; then
    test_results["integration"]="PASS"
    ((passed_tests++))
else
    test_results["integration"]="FAIL"
    ((failed_tests++))
fi
((total_tests++))

# Generate comprehensive report
log_header "📊 COMPREHENSIVE TEST REPORT"

echo "Test Execution Summary:" | tee -a "$FULL_LOG"
echo "  Mode: $TEST_MODE" | tee -a "$FULL_LOG"
echo "  Start Time: $(date)" | tee -a "$FULL_LOG"
echo "  Duration: $SECONDS seconds" | tee -a "$FULL_LOG"
echo "  Log File: $FULL_LOG" | tee -a "$FULL_LOG"
echo | tee -a "$FULL_LOG"

echo "Test Phase Results:" | tee -a "$FULL_LOG"
for phase in health features communication integration; do
    if [[ -n "${test_results[$phase]}" ]]; then
        result="${test_results[$phase]}"
        if [ "$result" = "PASS" ]; then
            echo -e "  ${GREEN}✅ $(printf '%-20s' "${phase^}:")${NC} $result" | tee -a "$FULL_LOG"
        else
            echo -e "  ${RED}❌ $(printf '%-20s' "${phase^}:")${NC} $result" | tee -a "$FULL_LOG"
        fi
    fi
done

echo | tee -a "$FULL_LOG"
echo "Overall Statistics:" | tee -a "$FULL_LOG"
echo -e "  Total Test Phases: $total_tests" | tee -a "$FULL_LOG"
echo -e "  ${GREEN}Passed: $passed_tests${NC}" | tee -a "$FULL_LOG"
echo -e "  ${RED}Failed: $failed_tests${NC}" | tee -a "$FULL_LOG"

success_rate=$((passed_tests * 100 / total_tests))
echo -e "  Success Rate: $success_rate%" | tee -a "$FULL_LOG"

echo | tee -a "$FULL_LOG"

# Final assessment and recommendations
log_section "🎯 Assessment and Recommendations"

if [ $failed_tests -eq 0 ]; then
    log_success "🎉 ALL TESTS PASSED! SaaS Admin Dashboard is working excellently!"
    echo | tee -a "$FULL_LOG"
    echo "✨ Your Beakon platform is ready for:" | tee -a "$FULL_LOG"
    echo "  • Production deployment" | tee -a "$FULL_LOG"
    echo "  • Customer onboarding" | tee -a "$FULL_LOG"
    echo "  • Full feature utilization" | tee -a "$FULL_LOG"

elif [ $success_rate -ge 75 ]; then
    log_success "✅ GOOD RESULTS! Most features are working correctly"
    echo | tee -a "$FULL_LOG"
    echo "🔧 Minor issues to address:" | tee -a "$FULL_LOG"
    for phase in "${!test_results[@]}"; do
        if [ "${test_results[$phase]}" = "FAIL" ]; then
            echo "  • Review $phase test failures" | tee -a "$FULL_LOG"
        fi
    done

elif [ $success_rate -ge 50 ]; then
    log_warning "⚠️  MODERATE RESULTS - Several areas need attention"
    echo | tee -a "$FULL_LOG"
    echo "🔨 Priority fixes needed:" | tee -a "$FULL_LOG"
    echo "  • Check service connectivity and authentication" | tee -a "$FULL_LOG"
    echo "  • Verify database connections" | tee -a "$FULL_LOG"
    echo "  • Review configuration settings" | tee -a "$FULL_LOG"

else
    log_failure "❌ SIGNIFICANT ISSUES - Major problems detected"
    echo | tee -a "$FULL_LOG"
    echo "🚨 Critical fixes required:" | tee -a "$FULL_LOG"
    echo "  • Many services may not be running or configured correctly" | tee -a "$FULL_LOG"
    echo "  • Check docker-compose or service startup scripts" | tee -a "$FULL_LOG"
    echo "  • Verify environment variables and database connections" | tee -a "$FULL_LOG"
fi

echo | tee -a "$FULL_LOG"
log_info "📋 Next Steps:"
echo "1. Review detailed logs: $FULL_LOG" | tee -a "$FULL_LOG"
echo "2. Check individual test script outputs above" | tee -a "$FULL_LOG"
echo "3. Fix any failing services or configurations" | tee -a "$FULL_LOG"
echo "4. Re-run tests: $0 $TEST_MODE" | tee -a "$FULL_LOG"

if [ $failed_tests -gt 0 ]; then
    echo | tee -a "$FULL_LOG"
    log_info "🛠️  Common troubleshooting:"
    echo "• Start services: docker-compose -f docker-compose.microservices.yml up -d" | tee -a "$FULL_LOG"
    echo "• Check logs: docker-compose logs <service-name>" | tee -a "$FULL_LOG"
    echo "• Verify ports: netstat -tlnp | grep <port>" | tee -a "$FULL_LOG"
    echo "• Test individual service: curl http://localhost:<port>/health" | tee -a "$FULL_LOG"
fi

echo | tee -a "$FULL_LOG"
log_header "🏁 Test Run Complete - $(date)"

# Exit with appropriate code
exit $failed_tests