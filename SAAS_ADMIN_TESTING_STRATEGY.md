# SaaS Admin Dashboard - Comprehensive Testing Strategy

## Overview
This document outlines a systematic approach to testing the SaaS Admin Dashboard functionality, ensuring each feature is implemented in the correct microservice and that inter-service communication works properly.

## Feature Mapping to Microservices

### 🏠 Native SaaS Admin Service Features
*Features implemented directly in `saas-admin-service`*

#### Platform Management
- **Platform Configuration** (`/api/v1/platform`)
  - GET: Retrieve platform settings
  - PUT: Update platform configuration
- **Statistics Dashboard** (`/api/v1/stats`)
  - Platform-wide usage statistics
  - Performance metrics overview

#### Subscription & Plan Management
- **Plan Management** (`/api/v1/plans`)
  - CRUD operations for subscription plans
  - Plan by slug retrieval
- **Feature Management** (`/api/v1/features`)
  - Platform feature definitions
  - Feature lifecycle management
- **Feature Flag Management** (`/api/v1/feature-flags`)
  - Dynamic feature toggling
  - Feature rollout control

#### Pricing Management
- **Pricing Features** (`/api/v1/pricing/features`)
  - Pricing component management
- **Pricing Tiers** (`/api/v1/pricing/tiers`)
  - Tiered pricing structure
- **Plan-Feature Association** (`/api/v1/pricing/plans/features`)
  - Feature assignment to plans
- **Public Pricing API** (`/api/v1/pricing/plans/public`)
  - External pricing information

#### Administrative Functions
- **Admin User Management** (`/api/v1/admin-users`)
  - Platform administrator accounts
- **Notification Management** (`/api/v1/notifications`)
  - System-wide notifications
- **Activity Logging** (`/api/v1/activities`)
  - Platform audit trails
- **Backup Management** (`/api/v1/backups`)
  - System backup operations

#### Web Dashboard
- **HTML Interface** (`/admin`)
  - Server-side rendered dashboard
- **Static Assets** (`/static/*`)
  - CSS, JavaScript, images

### 🔗 Proxied Features
*Features implemented in other microservices, accessed via SaaS Admin*

#### Tenant Management → `tenant-admin-service`
- **Tenant CRUD** (`/api/v1/tenants`)
  - Proxied to tenant-admin-service:8091
- **Dependencies**: Tenant Admin Service must be running and accessible

#### Analytics Management → `analytics-service`
- **Analytics Overview** (`/api/v1/analytics/overview`)
- **Analytics Metrics** (`/api/v1/analytics/metrics`)
- **Analytics Health** (`/api/v1/analytics/health`)
- **Dependencies**: Analytics Service must be running

#### Real-time Monitoring → `monitoring-service`
- **Monitor Management** (`/api/v1/monitoring/monitors`)
- **Real-time Statistics** (`/api/v1/monitoring/stats`)
- **WebSocket Monitoring** (`/api/v1/ws/monitoring`)
- **Dependencies**: Monitoring Service must be running

#### Incident Management → `incident-service`
- **Incident Operations** (`/api/v1/incidents`)
- **Incident Statistics** (`/api/v1/incidents/stats`)
- **Dependencies**: Incident Service must be running

#### Component Management → `component-service`
- **Component Operations** (`/api/v1/components`)
- **Component Groups** (`/api/v1/component-groups`)
- **Dependencies**: Component Service must be running

#### Integration Management → Multiple Services
- **Webhook Management** (`/api/v1/webhooks`)
- **Integration Operations** (`/api/v1/integrations`)
- **Subscriber Management** (`/api/v1/subscribers`)

## Testing Strategy

### Phase 1: Individual Microservice Testing

#### 1.1 SaaS Admin Service Core Features

```bash
# Test Platform Configuration
curl -X GET http://localhost:8092/api/v1/platform
curl -X PUT http://localhost:8092/api/v1/platform -H "Content-Type: application/json" -d '{
  "name": "Test Platform",
  "description": "Test Description"
}'

# Test Plan Management
curl -X POST http://localhost:8092/api/v1/plans -H "Content-Type: application/json" -d '{
  "name": "Basic Plan",
  "slug": "basic",
  "price": 9.99,
  "currency": "USD"
}'
curl -X GET http://localhost:8092/api/v1/plans
curl -X GET http://localhost:8092/api/v1/plans/slug/basic

# Test Feature Management
curl -X POST http://localhost:8092/api/v1/features -H "Content-Type: application/json" -d '{
  "name": "Custom Domains",
  "key": "custom_domains",
  "description": "Allow custom domain configuration"
}'
curl -X GET http://localhost:8092/api/v1/features

# Test Feature Flags
curl -X POST http://localhost:8092/api/v1/feature-flags -H "Content-Type: application/json" -d '{
  "name": "beta_features",
  "enabled": true,
  "description": "Beta features toggle"
}'

# Test Admin User Management
curl -X POST http://localhost:8092/api/v1/admin-users -H "Content-Type: application/json" -d '{
  "email": "admin@test.com",
  "name": "Test Admin",
  "role": "admin"
}'

# Test Statistics
curl -X GET http://localhost:8092/api/v1/stats
```

#### 1.2 Dependency Service Testing

```bash
# Test Tenant Admin Service (Port 8091)
curl -X GET http://localhost:8091/health
curl -X GET http://localhost:8091/api/v1/tenants

# Test Analytics Service (Port 8096)
curl -X GET http://localhost:8096/health
curl -X GET http://localhost:8096/api/v1/analytics/overview

# Test Monitoring Service (Port 8095)
curl -X GET http://localhost:8095/health
curl -X GET http://localhost:8095/api/v1/monitoring/services

# Test Component Service (Port 8093)
curl -X GET http://localhost:8093/health
curl -X GET http://localhost:8093/api/v1/components

# Test Incident Service (Port 8094)
curl -X GET http://localhost:8094/health
curl -X GET http://localhost:8094/api/v1/incidents
```

### Phase 2: Inter-Service Communication Testing

#### 2.1 Proxy Connection Tests

```bash
# Test SaaS Admin → Tenant Admin Proxy
curl -X GET http://localhost:8092/api/v1/tenants
# Should proxy to http://localhost:8091/api/v1/tenants

# Test SaaS Admin → Analytics Proxy
curl -X GET http://localhost:8092/api/v1/analytics/overview
# Should proxy to analytics service

# Test SaaS Admin → Monitoring Proxy
curl -X GET http://localhost:8092/api/v1/monitoring/monitors
# Should proxy to monitoring service

# Test SaaS Admin → Component Proxy
curl -X GET http://localhost:8092/api/v1/components
# Should proxy to component service
```

#### 2.2 Error Handling Tests

```bash
# Test behavior when dependency services are down
# Stop tenant-admin-service
curl -X GET http://localhost:8092/api/v1/tenants
# Should return appropriate error/timeout

# Test circuit breaker functionality
# Make multiple requests to unavailable service
for i in {1..10}; do
  curl -X GET http://localhost:8092/api/v1/tenants
done
```

### Phase 3: End-to-End Dashboard Integration

#### 3.1 Web Dashboard Tests

```bash
# Test Dashboard Access
curl -X GET http://localhost:8092/admin
# Should return HTML dashboard

# Test Static Assets
curl -X GET http://localhost:8092/static/css/dashboard.css
curl -X GET http://localhost:8092/static/js/dashboard.js

# Test Favicon
curl -X GET http://localhost:8092/favicon.ico
```

#### 3.2 WebSocket Functionality

```bash
# Test WebSocket Connection
curl --include \
     --no-buffer \
     --header "Connection: Upgrade" \
     --header "Upgrade: websocket" \
     --header "Sec-WebSocket-Key: SGVsbG8sIHdvcmxkIQ==" \
     --header "Sec-WebSocket-Version: 13" \
     http://localhost:8092/api/v1/ws/monitoring
```

### Phase 4: Data Flow Validation

#### 4.1 Complete Feature Workflows

**Plan Creation to Public Display Workflow**:
```bash
# 1. Create plan in SaaS Admin
curl -X POST http://localhost:8092/api/v1/plans -d '{...}'

# 2. Create pricing features
curl -X POST http://localhost:8092/api/v1/pricing/features -d '{...}'

# 3. Associate features with plan
curl -X POST http://localhost:8092/api/v1/pricing/plans/features/assign -d '{...}'

# 4. Sync to landing page
curl -X POST http://localhost:8092/api/v1/pricing/sync

# 5. Verify public display
curl -X GET http://localhost:8092/api/v1/pricing/plans/public
```

**Tenant Management Workflow**:
```bash
# 1. Create tenant via SaaS Admin (proxied)
curl -X POST http://localhost:8092/api/v1/tenants -d '{...}'

# 2. Verify tenant exists in Tenant Admin Service
curl -X GET http://localhost:8091/api/v1/tenants

# 3. Configure tenant features
curl -X PUT http://localhost:8092/api/v1/tenants/1 -d '{...}'

# 4. Verify tenant statistics
curl -X GET http://localhost:8092/api/v1/stats
```

## Automated Test Scripts

### Service Health Check Script

```bash
#!/bin/bash
# test-service-health.sh

services=(
    "saas-admin-service:8092"
    "tenant-admin-service:8091"
    "analytics-service:8096"
    "monitoring-service:8095"
    "component-service:8093"
    "incident-service:8094"
)

echo "Testing service health..."
for service in "${services[@]}"; do
    name=$(echo $service | cut -d: -f1)
    port=$(echo $service | cut -d: -f2)

    echo -n "Testing $name on port $port... "
    if curl -s -f http://localhost:$port/health > /dev/null; then
        echo "✅ Healthy"
    else
        echo "❌ Unhealthy"
    fi
done
```

### Feature Testing Script

```bash
#!/bin/bash
# test-saas-admin-features.sh

BASE_URL="http://localhost:8092"

echo "Testing SaaS Admin native features..."

# Test Platform Configuration
echo "Testing platform configuration..."
curl -s -X GET $BASE_URL/api/v1/platform | jq .

# Test Plan Management
echo "Testing plan management..."
curl -s -X POST $BASE_URL/api/v1/plans \
    -H "Content-Type: application/json" \
    -d '{"name":"Test Plan","slug":"test","price":9.99}' | jq .

# Test Proxied Features
echo "Testing proxied tenant management..."
curl -s -X GET $BASE_URL/api/v1/tenants | jq .

echo "Testing proxied analytics..."
curl -s -X GET $BASE_URL/api/v1/analytics/overview | jq .
```

### Integration Test Script

```bash
#!/bin/bash
# test-dashboard-integration.sh

BASE_URL="http://localhost:8092"

echo "Testing SaaS Admin Dashboard Integration..."

# Test Web Dashboard
echo "Testing web dashboard access..."
response=$(curl -s -w "%{http_code}" $BASE_URL/admin)
if [[ $response == *"200"* ]]; then
    echo "✅ Dashboard accessible"
else
    echo "❌ Dashboard failed: $response"
fi

# Test API Gateway Integration
echo "Testing API Gateway routing..."
response=$(curl -s -w "%{http_code}" http://localhost:8080/api/v1/platform)
if [[ $response == *"200"* ]]; then
    echo "✅ API Gateway routing works"
else
    echo "❌ API Gateway routing failed: $response"
fi

# Test Database Connectivity
echo "Testing database operations..."
curl -s -X GET $BASE_URL/api/v1/stats | jq .
```

## Validation Checklist

### ✅ Core Features Validation
- [ ] Platform configuration CRUD operations
- [ ] Plan management (create, read, update, delete, slug lookup)
- [ ] Feature management lifecycle
- [ ] Feature flag toggling
- [ ] Pricing management and public API
- [ ] Admin user management
- [ ] Notification system
- [ ] Activity logging
- [ ] Backup operations
- [ ] Statistics dashboard

### ✅ Inter-Service Communication
- [ ] Tenant management proxy to tenant-admin-service
- [ ] Analytics proxy to analytics-service
- [ ] Monitoring proxy to monitoring-service
- [ ] Component management proxy to component-service
- [ ] Incident management proxy to incident-service
- [ ] Error handling for unavailable services
- [ ] Circuit breaker functionality
- [ ] Timeout handling

### ✅ Web Dashboard Integration
- [ ] HTML dashboard rendering
- [ ] Static asset serving (CSS, JS, images)
- [ ] Template loading and rendering
- [ ] Favicon serving
- [ ] CSP header configuration for CDN resources
- [ ] WebSocket connectivity for real-time monitoring

### ✅ Security & Authentication
- [ ] JWT authentication for protected endpoints
- [ ] Multi-tenant data isolation
- [ ] CORS configuration
- [ ] Rate limiting (if enabled)
- [ ] Input validation
- [ ] Session management

### ✅ Data Consistency
- [ ] Plan-feature associations
- [ ] Pricing synchronization to landing page
- [ ] Tenant statistics accuracy
- [ ] Real-time monitoring data flow
- [ ] Activity log completeness
- [ ] Backup data integrity

## Testing Environment Setup

```bash
# Start all required services
docker-compose -f docker-compose.microservices.yml up -d

# Wait for services to be ready
sleep 30

# Run health checks
./test-service-health.sh

# Run feature tests
./test-saas-admin-features.sh

# Run integration tests
./test-dashboard-integration.sh

# Run end-to-end validation
./validate-dashboard-workflows.sh
```

This comprehensive testing strategy ensures that each feature is properly implemented in its correct microservice and that the SaaS Admin Dashboard can orchestrate all services effectively.