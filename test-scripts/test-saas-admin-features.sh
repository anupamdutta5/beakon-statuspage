#!/bin/bash
# Test SaaS Admin Service Features - Comprehensive feature testing
# Usage: ./test-saas-admin-features.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8092"
TEST_RESULTS=()

# Utility functions
log_test() {
    echo -e "${BLUE}🧪 $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
    TEST_RESULTS+=("PASS: $1")
}

log_failure() {
    echo -e "${RED}❌ $1${NC}"
    TEST_RESULTS+=("FAIL: $1")
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4

    log_test "Testing $description"

    if [ -n "$data" ]; then
        response=$(curl -s -w "%{http_code}" -X "$method" "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer fake-test-token" \
            -d "$data" 2>/dev/null)
    else
        response=$(curl -s -w "%{http_code}" -X "$method" "$BASE_URL$endpoint" \
            -H "Authorization: Bearer fake-test-token" 2>/dev/null)
    fi

    http_code="${response: -3}"
    body="${response%???}"

    case "$http_code" in
        200|201|204)
            log_success "$description - HTTP $http_code"
            if [ -n "$body" ] && [ "$body" != "null" ]; then
                echo "   Response: ${body:0:100}$([ ${#body} -gt 100 ] && echo '...')"
            fi
            return 0
            ;;
        400|401|403|404|422)
            log_warning "$description - HTTP $http_code (Expected for some tests)"
            echo "   Response: ${body:0:100}$([ ${#body} -gt 100 ] && echo '...')"
            return 1
            ;;
        500|502|503|504)
            log_failure "$description - HTTP $http_code (Server Error)"
            echo "   Response: ${body:0:100}$([ ${#body} -gt 100 ] && echo '...')"
            return 1
            ;;
        000)
            log_failure "$description - Connection failed"
            return 1
            ;;
        *)
            log_failure "$description - HTTP $http_code (Unexpected)"
            return 1
            ;;
    esac
}

echo "🔧 Testing SaaS Admin Service Features"
echo "======================================"

# Check if service is running
if ! curl -s -f "$BASE_URL/api/v1/health" > /dev/null 2>&1; then
    log_failure "SaaS Admin Service is not running on port 8092"
    echo -e "${YELLOW}💡 Start the service first: cd microservices/saas-admin-service && go run cmd/main.go${NC}"
    exit 1
fi

log_success "SaaS Admin Service is running"
echo

# Test 1: Platform Configuration
echo -e "${BLUE}📊 Testing Platform Management${NC}"
echo "------------------------------"

test_endpoint "GET" "/api/v1/platform" "" "Get platform configuration"

platform_data='{
  "name": "Test Platform",
  "description": "Test platform for validation",
  "domain": "test.statuspage.com",
  "logo_url": "https://example.com/logo.png"
}'
test_endpoint "PUT" "/api/v1/platform" "$platform_data" "Update platform configuration"

test_endpoint "GET" "/api/v1/stats" "" "Get platform statistics"
echo

# Test 2: Plan Management
echo -e "${BLUE}📋 Testing Plan Management${NC}"
echo "----------------------------"

test_endpoint "GET" "/api/v1/plans" "" "List all plans"

plan_data='{
  "name": "Test Basic Plan",
  "slug": "test-basic",
  "description": "Basic plan for testing",
  "price": 9.99,
  "currency": "USD",
  "billing_interval": "monthly",
  "max_components": 10,
  "max_team_members": 5
}'
test_endpoint "POST" "/api/v1/plans" "$plan_data" "Create new plan"

test_endpoint "GET" "/api/v1/plans/slug/test-basic" "" "Get plan by slug"

updated_plan_data='{
  "name": "Updated Test Basic Plan",
  "description": "Updated description",
  "price": 12.99
}'
test_endpoint "PUT" "/api/v1/plans/1" "$updated_plan_data" "Update plan"
echo

# Test 3: Feature Management
echo -e "${BLUE}🎛️  Testing Feature Management${NC}"
echo "------------------------------"

test_endpoint "GET" "/api/v1/features" "" "List all features"

feature_data='{
  "name": "Custom Domains",
  "key": "custom_domains",
  "description": "Allow customers to use custom domains",
  "category": "branding",
  "is_active": true
}'
test_endpoint "POST" "/api/v1/features" "$feature_data" "Create new feature"

test_endpoint "GET" "/api/v1/features/1" "" "Get feature details"
echo

# Test 4: Feature Flag Management
echo -e "${BLUE}🎚️  Testing Feature Flag Management${NC}"
echo "----------------------------------"

test_endpoint "GET" "/api/v1/feature-flags" "" "List all feature flags"

flag_data='{
  "name": "beta_monitoring",
  "key": "beta_monitoring",
  "description": "Enable beta monitoring features",
  "enabled": true,
  "rollout_percentage": 50
}'
test_endpoint "POST" "/api/v1/feature-flags" "$flag_data" "Create new feature flag"

toggle_data='{
  "enabled": false,
  "rollout_percentage": 0
}'
test_endpoint "PUT" "/api/v1/feature-flags/1" "$toggle_data" "Toggle feature flag"
echo

# Test 5: Pricing Management
echo -e "${BLUE}💰 Testing Pricing Management${NC}"
echo "-----------------------------"

test_endpoint "GET" "/api/v1/pricing/features" "" "Get pricing features"

pricing_feature_data='{
  "name": "Advanced Analytics",
  "key": "advanced_analytics",
  "description": "Advanced analytics and reporting",
  "feature_type": "addon",
  "price": 5.00
}'
test_endpoint "POST" "/api/v1/pricing/features" "$pricing_feature_data" "Create pricing feature"

test_endpoint "GET" "/api/v1/pricing/plans/public" "" "Get public pricing plans"

# Test feature assignment
assignment_data='{
  "plan_id": 1,
  "feature_id": 1,
  "included_quantity": 1,
  "overage_price": 1.00
}'
test_endpoint "POST" "/api/v1/pricing/plans/features/assign" "$assignment_data" "Assign feature to plan"
echo

# Test 6: Admin User Management
echo -e "${BLUE}👥 Testing Admin User Management${NC}"
echo "--------------------------------"

test_endpoint "GET" "/api/v1/admin-users" "" "List admin users"

admin_user_data='{
  "email": "testadmin@example.com",
  "name": "Test Administrator",
  "role": "admin",
  "permissions": ["platform.read", "platform.write", "users.manage"]
}'
test_endpoint "POST" "/api/v1/admin-users" "$admin_user_data" "Create admin user"

test_endpoint "GET" "/api/v1/admin-users/1" "" "Get admin user details"
echo

# Test 7: Notification Management
echo -e "${BLUE}🔔 Testing Notification Management${NC}"
echo "---------------------------------"

test_endpoint "GET" "/api/v1/notifications" "" "List notifications"

notification_data='{
  "title": "System Maintenance",
  "message": "Scheduled maintenance window tonight",
  "type": "maintenance",
  "priority": "medium",
  "target_audience": "all_admins"
}'
test_endpoint "POST" "/api/v1/notifications" "$notification_data" "Create notification"
echo

# Test 8: Activity Logging
echo -e "${BLUE}📝 Testing Activity Logging${NC}"
echo "---------------------------"

test_endpoint "GET" "/api/v1/activities" "" "Get activity logs"
echo

# Test 9: Backup Management
echo -e "${BLUE}💾 Testing Backup Management${NC}"
echo "----------------------------"

test_endpoint "GET" "/api/v1/backups" "" "List backups"

backup_data='{
  "name": "Test Backup",
  "description": "Automated test backup",
  "include_data": true,
  "include_config": true
}'
test_endpoint "POST" "/api/v1/backups" "$backup_data" "Create backup"
echo

# Test 10: Web Dashboard
echo -e "${BLUE}🌐 Testing Web Dashboard${NC}"
echo "-------------------------"

log_test "Testing web dashboard access"
if curl -s -f "$BASE_URL/admin" > /dev/null 2>&1; then
    log_success "Admin dashboard accessible"
else
    log_failure "Admin dashboard not accessible"
fi

log_test "Testing static assets"
if curl -s -f "$BASE_URL/favicon.ico" > /dev/null 2>&1; then
    log_success "Static assets accessible"
else
    log_failure "Static assets not accessible"
fi
echo

# Test 11: Proxied Features (Basic connectivity)
echo -e "${BLUE}🔗 Testing Proxied Features Connectivity${NC}"
echo "----------------------------------------"

test_endpoint "GET" "/api/v1/tenants" "" "Tenant management proxy (to tenant-admin-service)"
test_endpoint "GET" "/api/v1/analytics/overview" "" "Analytics proxy (to analytics-service)"
test_endpoint "GET" "/api/v1/monitoring/stats" "" "Monitoring proxy (to monitoring-service)"
test_endpoint "GET" "/api/v1/components" "" "Component management proxy (to component-service)"
test_endpoint "GET" "/api/v1/incidents" "" "Incident management proxy (to incident-service)"
echo

# Test Summary
echo "======================================"
echo -e "${BLUE}📊 Test Results Summary${NC}"
echo "======================================"

pass_count=0
fail_count=0

for result in "${TEST_RESULTS[@]}"; do
    if [[ $result == PASS:* ]]; then
        ((pass_count++))
        echo -e "${GREEN}$result${NC}"
    else
        ((fail_count++))
        echo -e "${RED}$result${NC}"
    fi
done

total_tests=$((pass_count + fail_count))

echo
echo "======================================"
echo -e "Total Tests: $total_tests"
echo -e "${GREEN}Passed: $pass_count${NC}"
echo -e "${RED}Failed: $fail_count${NC}"

if [ $fail_count -eq 0 ]; then
    echo -e "${GREEN}🎉 All SaaS Admin features are working correctly!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  Some tests failed. Check the details above.${NC}"
    echo
    echo -e "${YELLOW}💡 Common issues:${NC}"
    echo "1. Authentication: Most endpoints require valid JWT tokens"
    echo "2. Database: Ensure PostgreSQL is running and accessible"
    echo "3. Dependencies: Some features require other microservices to be running"
    echo "4. Configuration: Check environment variables and service configuration"
    exit 1
fi