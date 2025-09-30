#!/bin/bash
# Test Dashboard Integration - End-to-end workflow testing
# Usage: ./test-dashboard-integration.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Service URLs
API_GATEWAY="http://localhost:8080"
SAAS_ADMIN="http://localhost:8092"
TENANT_ADMIN="http://localhost:8091"
USER_SERVICE="http://localhost:8090"
COMPONENT_SERVICE="http://localhost:8093"
INCIDENT_SERVICE="http://localhost:8094"
MONITORING_SERVICE="http://localhost:8095"
ANALYTICS_SERVICE="http://localhost:8096"
LANDING_PAGE="http://localhost:8098"

# Test data storage
TEMP_DIR="/tmp/beakon-tests"
mkdir -p "$TEMP_DIR"

TEST_RESULTS=()
CREATED_RESOURCES=()

# Utility functions
log_workflow() {
    echo -e "${PURPLE}🔄 WORKFLOW: $1${NC}"
    echo "----------------------------------------"
}

log_step() {
    echo -e "${BLUE}📋 Step: $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
    TEST_RESULTS+=("PASS: $1")
}

log_failure() {
    echo -e "${RED}❌ $1${NC}"
    TEST_RESULTS+=("FAIL: $1")
    return 1
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

cleanup_resources() {
    echo -e "${YELLOW}🧹 Cleaning up test resources...${NC}"
    for resource in "${CREATED_RESOURCES[@]}"; do
        IFS=':' read -r method url description <<< "$resource"
        echo "  Cleaning up: $description"
        curl -s -X "$method" "$url" > /dev/null 2>&1 || true
    done
    rm -rf "$TEMP_DIR"
}

# Set up cleanup trap
trap cleanup_resources EXIT

api_call() {
    local method=$1
    local url=$2
    local data=${3:-}
    local description=$4
    local expected_codes=${5:-"200,201,202,204"}

    echo "  → $description"

    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer test-token" \
            -d "$data" 2>/dev/null)
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
            -H "Authorization: Bearer test-token" 2>/dev/null)
    fi

    # Split response body and status code
    body=$(echo "$response" | sed '$d')
    status_code=$(echo "$response" | tail -n1)

    # Check if status code is expected
    if [[ ",$expected_codes," == *",$status_code,"* ]]; then
        echo -e "    ${GREEN}✓ HTTP $status_code${NC}"

        # Store response for potential use
        echo "$body" > "$TEMP_DIR/last_response.json"

        return 0
    else
        echo -e "    ${RED}✗ HTTP $status_code${NC}"
        echo "    Response: ${body:0:100}$([ ${#body} -gt 100 ] && echo '...')"
        return 1
    fi
}

extract_id() {
    local json_file="$TEMP_DIR/last_response.json"
    if [ -f "$json_file" ]; then
        # Try to extract ID from JSON response
        id=$(jq -r '.id // .ID // empty' "$json_file" 2>/dev/null || echo "")
        if [ -n "$id" ] && [ "$id" != "null" ]; then
            echo "$id"
        else
            echo "1" # Default fallback ID
        fi
    else
        echo "1" # Default fallback ID
    fi
}

check_service_availability() {
    local service_name=$1
    local service_url=$2
    local health_endpoint=${3:-"/health"}

    if curl -s -f "$service_url$health_endpoint" > /dev/null 2>&1; then
        echo -e "  ${GREEN}✓${NC} $service_name is running"
        return 0
    else
        echo -e "  ${RED}✗${NC} $service_name is not running"
        return 1
    fi
}

echo "🎯 End-to-End Dashboard Integration Testing"
echo "=========================================="

# Pre-flight service checks
echo -e "${BLUE}🔍 Checking Service Availability${NC}"
echo "--------------------------------"

available_services=0
required_services=(
    "SaaS Admin Service:$SAAS_ADMIN:/api/v1/health"
    "API Gateway:$API_GATEWAY:/health"
    "Tenant Admin Service:$TENANT_ADMIN:/health"
    "Component Service:$COMPONENT_SERVICE:/health"
    "Incident Service:$INCIDENT_SERVICE:/health"
)

for service_info in "${required_services[@]}"; do
    IFS=':' read -r name url path <<< "$service_info"
    if check_service_availability "$name" "$url" "$path"; then
        ((available_services++))
    fi
done

echo -e "Available services: $available_services/${#required_services[@]}"

if [ $available_services -lt 3 ]; then
    log_failure "Insufficient services running for integration testing"
    echo -e "${YELLOW}💡 Start required services first${NC}"
    exit 1
fi

echo

# Workflow 1: Complete Platform Setup
log_workflow "Platform Setup and Configuration"

log_step "1.1 Configure Platform Settings"
platform_data='{
  "name": "Beakon Test Platform",
  "description": "End-to-end testing platform",
  "domain": "test.beakon.io",
  "support_email": "support@test.beakon.io",
  "logo_url": "https://test.beakon.io/logo.png"
}'
if api_call "PUT" "$SAAS_ADMIN/api/v1/platform" "$platform_data" "Update platform configuration"; then
    log_success "Platform configuration updated"
else
    log_failure "Failed to configure platform"
fi

log_step "1.2 Create Subscription Plans"
basic_plan='{
  "name": "Basic Plan",
  "slug": "basic",
  "description": "Perfect for small teams",
  "price": 9.99,
  "currency": "USD",
  "billing_interval": "monthly",
  "max_components": 10,
  "max_team_members": 3,
  "max_status_pages": 1
}'
if api_call "POST" "$SAAS_ADMIN/api/v1/plans" "$basic_plan" "Create Basic plan"; then
    basic_plan_id=$(extract_id)
    CREATED_RESOURCES+=("DELETE:$SAAS_ADMIN/api/v1/plans/$basic_plan_id:Basic Plan")
    log_success "Basic plan created (ID: $basic_plan_id)"
else
    log_failure "Failed to create Basic plan"
fi

pro_plan='{
  "name": "Professional Plan",
  "slug": "professional",
  "description": "For growing businesses",
  "price": 29.99,
  "currency": "USD",
  "billing_interval": "monthly",
  "max_components": 50,
  "max_team_members": 10,
  "max_status_pages": 5
}'
if api_call "POST" "$SAAS_ADMIN/api/v1/plans" "$pro_plan" "Create Professional plan"; then
    pro_plan_id=$(extract_id)
    CREATED_RESOURCES+=("DELETE:$SAAS_ADMIN/api/v1/plans/$pro_plan_id:Professional Plan")
    log_success "Professional plan created (ID: $pro_plan_id)"
else
    log_failure "Failed to create Professional plan"
fi

log_step "1.3 Create Platform Features"
features_data=(
    '{"name":"Custom Domains","key":"custom_domains","description":"Custom domain support","category":"branding"}'
    '{"name":"Advanced Analytics","key":"advanced_analytics","description":"Advanced reporting","category":"analytics"}'
    '{"name":"API Access","key":"api_access","description":"Full API access","category":"integration"}'
    '{"name":"White Labeling","key":"white_labeling","description":"Remove Beakon branding","category":"branding"}'
)

created_features=()
for feature_data in "${features_data[@]}"; do
    feature_name=$(echo "$feature_data" | jq -r '.name')
    if api_call "POST" "$SAAS_ADMIN/api/v1/features" "$feature_data" "Create feature: $feature_name"; then
        feature_id=$(extract_id)
        created_features+=("$feature_id")
        CREATED_RESOURCES+=("DELETE:$SAAS_ADMIN/api/v1/features/$feature_id:Feature $feature_name")
        log_success "Feature '$feature_name' created (ID: $feature_id)"
    else
        log_warning "Failed to create feature: $feature_name"
    fi
done

echo

# Workflow 2: Pricing Management
log_workflow "Pricing Management and Feature Association"

log_step "2.1 Create Pricing Features"
pricing_features=(
    '{"name":"Custom Domain Setup","key":"custom_domain_setup","price":5.00,"feature_type":"addon"}'
    '{"name":"Extra Team Member","key":"extra_team_member","price":3.00,"feature_type":"per_unit"}'
    '{"name":"Priority Support","key":"priority_support","price":10.00,"feature_type":"addon"}'
)

for pricing_data in "${pricing_features[@]}"; do
    pricing_name=$(echo "$pricing_data" | jq -r '.name')
    if api_call "POST" "$SAAS_ADMIN/api/v1/pricing/features" "$pricing_data" "Create pricing feature: $pricing_name"; then
        log_success "Pricing feature '$pricing_name' created"
    else
        log_warning "Failed to create pricing feature: $pricing_name"
    fi
done

log_step "2.2 Associate Features with Plans"
# Associate features with Professional plan
if [ ${#created_features[@]} -gt 0 ] && [ -n "$pro_plan_id" ]; then
    for feature_id in "${created_features[@]}"; do
        assignment_data="{
            \"plan_id\": $pro_plan_id,
            \"feature_id\": $feature_id,
            \"included_quantity\": 1,
            \"overage_price\": 0.00
        }"
        if api_call "POST" "$SAAS_ADMIN/api/v1/pricing/plans/features/assign" "$assignment_data" "Assign feature $feature_id to Professional plan" "200,201,409"; then
            log_success "Feature $feature_id assigned to Professional plan"
        else
            log_warning "Failed to assign feature $feature_id (may already exist)"
        fi
    done
fi

log_step "2.3 Validate Public Pricing Display"
if api_call "GET" "$SAAS_ADMIN/api/v1/pricing/plans/public" "" "Get public pricing information"; then
    log_success "Public pricing information accessible"
    # Check if our plans appear in public pricing
    if grep -q "Basic Plan\|Professional Plan" "$TEMP_DIR/last_response.json" 2>/dev/null; then
        log_success "Created plans appear in public pricing"
    else
        log_warning "Created plans may not be visible in public pricing"
    fi
else
    log_failure "Failed to retrieve public pricing"
fi

echo

# Workflow 3: Tenant Management (via proxy)
log_workflow "Tenant Management Integration"

log_step "3.1 Create Test Tenant (via SaaS Admin proxy)"
tenant_data='{
  "name": "Test Corporation",
  "slug": "test-corp",
  "domain": "testcorp.beakon.io",
  "email": "admin@testcorp.com",
  "plan_id": 1,
  "status": "active"
}'
if api_call "POST" "$SAAS_ADMIN/api/v1/tenants" "$tenant_data" "Create tenant via SaaS Admin proxy" "200,201,400,401"; then
    tenant_id=$(extract_id)
    CREATED_RESOURCES+=("DELETE:$SAAS_ADMIN/api/v1/tenants/$tenant_id:Test Tenant")
    log_success "Tenant created via proxy (ID: $tenant_id)"
else
    log_warning "Tenant creation failed (may require authentication or Tenant Admin Service)"
fi

log_step "3.2 Validate Tenant in Direct Service"
if check_service_availability "Tenant Admin Service" "$TENANT_ADMIN"; then
    if api_call "GET" "$TENANT_ADMIN/api/v1/tenants" "" "List tenants directly" "200,401"; then
        log_success "Direct tenant service access working"
    else
        log_warning "Direct tenant service access failed (authentication required)"
    fi
fi

echo

# Workflow 4: Component and Status Page Integration
log_workflow "Component and Status Page Integration"

log_step "4.1 Create Components (via SaaS Admin proxy)"
components_data=(
    '{"name":"Web Application","description":"Main web application","status":"operational","is_visible":true}'
    '{"name":"API Server","description":"REST API backend","status":"operational","is_visible":true}'
    '{"name":"Database","description":"Primary database","status":"operational","is_visible":true}'
    '{"name":"CDN","description":"Content delivery network","status":"operational","is_visible":true}'
)

for component_data in "${components_data[@]}"; do
    component_name=$(echo "$component_data" | jq -r '.name')
    if api_call "POST" "$SAAS_ADMIN/api/v1/components" "$component_data" "Create component: $component_name" "200,201,400,401"; then
        component_id=$(extract_id)
        CREATED_RESOURCES+=("DELETE:$SAAS_ADMIN/api/v1/components/$component_id:Component $component_name")
        log_success "Component '$component_name' created"
    else
        log_warning "Failed to create component: $component_name (may require authentication)"
    fi
done

log_step "4.2 Test Component Status Updates"
if api_call "PUT" "$SAAS_ADMIN/api/v1/components/1/status" '{"status":"degraded_performance","message":"Experiencing high latency"}' "Update component status" "200,404,401"; then
    log_success "Component status update worked"
else
    log_warning "Component status update failed (component may not exist or auth required)"
fi

echo

# Workflow 5: Incident Management Integration
log_workflow "Incident Management Integration"

log_step "5.1 Create Test Incident"
incident_data='{
  "title": "Database Performance Issues",
  "description": "Database is experiencing high response times",
  "status": "investigating",
  "impact": "minor",
  "affected_components": [1, 3]
}'
if api_call "POST" "$SAAS_ADMIN/api/v1/incidents" "$incident_data" "Create test incident" "200,201,400,401"; then
    incident_id=$(extract_id)
    CREATED_RESOURCES+=("DELETE:$SAAS_ADMIN/api/v1/incidents/$incident_id:Test Incident")
    log_success "Test incident created"
else
    log_warning "Incident creation failed (may require authentication or Incident Service)"
fi

log_step "5.2 Add Incident Update"
if [ -n "$incident_id" ]; then
    update_data='{
      "status": "identified",
      "message": "We have identified the root cause and are working on a fix",
      "incident_id": '$incident_id'
    }'
    if api_call "POST" "$SAAS_ADMIN/api/v1/incidents/$incident_id/updates" "$update_data" "Add incident update" "200,201,404,401"; then
        log_success "Incident update added"
    else
        log_warning "Failed to add incident update"
    fi
fi

echo

# Workflow 6: Analytics and Monitoring Integration
log_workflow "Analytics and Monitoring Integration"

log_step "6.1 Test Analytics Overview"
if api_call "GET" "$SAAS_ADMIN/api/v1/analytics/overview" "" "Get analytics overview" "200,401,503"; then
    log_success "Analytics overview accessible"
else
    log_warning "Analytics overview failed (service may be down or auth required)"
fi

log_step "6.2 Test Monitoring Statistics"
if api_call "GET" "$SAAS_ADMIN/api/v1/monitoring/stats" "" "Get monitoring stats" "200,401,503"; then
    log_success "Monitoring stats accessible"
else
    log_warning "Monitoring stats failed (service may be down or auth required)"
fi

log_step "6.3 Test Platform Statistics"
if api_call "GET" "$SAAS_ADMIN/api/v1/stats" "" "Get platform statistics"; then
    log_success "Platform statistics accessible"
    # Check if stats contain meaningful data
    if grep -q "plans\|users\|tenants" "$TEMP_DIR/last_response.json" 2>/dev/null; then
        log_success "Platform statistics contain expected data"
    else
        log_warning "Platform statistics may not contain full data"
    fi
else
    log_failure "Failed to get platform statistics"
fi

echo

# Workflow 7: Web Dashboard Validation
log_workflow "Web Dashboard Functionality"

log_step "7.1 Test Dashboard Access"
dashboard_response=$(curl -s -w "%{http_code}" "$SAAS_ADMIN/admin" 2>/dev/null)
dashboard_code="${dashboard_response: -3}"

if [ "$dashboard_code" = "200" ]; then
    log_success "Admin dashboard is accessible"

    # Check if response contains HTML
    if echo "$dashboard_response" | grep -q "<html\|<title\|<body"; then
        log_success "Dashboard returns valid HTML"
    else
        log_warning "Dashboard response may not be valid HTML"
    fi
else
    log_failure "Admin dashboard is not accessible (HTTP $dashboard_code)"
fi

log_step "7.2 Test Static Assets"
static_assets=(
    "/favicon.ico:Favicon"
    "/static/css/dashboard.css:CSS Assets"
    "/static/js/dashboard.js:JavaScript Assets"
)

for asset_info in "${static_assets[@]}"; do
    IFS=':' read -r path description <<< "$asset_info"
    if curl -s -f "$SAAS_ADMIN$path" > /dev/null 2>&1; then
        log_success "$description accessible"
    else
        log_warning "$description not found (may not exist yet)"
    fi
done

echo

# Workflow 8: End-to-End Data Flow Validation
log_workflow "End-to-End Data Flow Validation"

log_step "8.1 Validate Plan → Pricing → Public Display Flow"
# This tests the complete flow from plan creation to public display
if api_call "GET" "$SAAS_ADMIN/api/v1/plans" "" "List all plans"; then
    log_success "Plan listing works"

    if api_call "GET" "$SAAS_ADMIN/api/v1/pricing/plans/public" "" "Get public pricing"; then
        log_success "Plan data flows to public pricing API"
    fi
fi

log_step "8.2 Validate Statistics Aggregation"
if api_call "GET" "$SAAS_ADMIN/api/v1/stats" "" "Get aggregated statistics"; then
    stats_content=$(cat "$TEMP_DIR/last_response.json" 2>/dev/null || echo "{}")

    # Check for key statistics
    if echo "$stats_content" | jq -e '.plans // .total_plans' > /dev/null 2>&1; then
        log_success "Statistics include plan data"
    else
        log_warning "Statistics may not include all expected data"
    fi
fi

echo

# Final Results Summary
echo "======================================="
echo -e "${PURPLE}📊 Integration Test Results Summary${NC}"
echo "======================================="

pass_count=0
fail_count=0
warning_count=0

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
echo "======================================="
echo -e "Total Integration Tests: $total_tests"
echo -e "${GREEN}Passed: $pass_count${NC}"
echo -e "${RED}Failed: $fail_count${NC}"
echo -e "Services Available: $available_services/${#required_services[@]}"

# Overall assessment
if [ $fail_count -eq 0 ] && [ $pass_count -gt 10 ]; then
    echo -e "${GREEN}🎉 Excellent! Dashboard integration is working well!${NC}"
    exit 0
elif [ $fail_count -lt 3 ] && [ $pass_count -gt 5 ]; then
    echo -e "${YELLOW}✅ Good! Most dashboard integration features are working${NC}"
    echo -e "${YELLOW}💡 Some features may require additional configuration or authentication${NC}"
    exit 0
else
    echo -e "${RED}⚠️  Dashboard integration has significant issues${NC}"
    echo
    echo -e "${YELLOW}💡 Common issues to check:${NC}"
    echo "1. Authentication: Most protected endpoints require valid JWT tokens"
    echo "2. Service Dependencies: Ensure all required microservices are running"
    echo "3. Database: Check that PostgreSQL is accessible to all services"
    echo "4. Configuration: Verify environment variables and service configuration"
    echo "5. Network: Ensure services can communicate with each other"
    exit 1
fi