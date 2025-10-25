# Phase 3 Week 11-12: Deployment Guide
## Advanced Analytics Dashboard + Service Dependency Mapping

**Features**: Advanced Analytics Dashboard, Service Dependency Mapping
**Deploy Time**: 30-40 minutes total
**Downtime Required**: None (zero-downtime deployment)
**Risk Level**: Low

---

## 📋 Pre-Deployment Checklist

### Environment Requirements

**Backend (tenant-admin-service)**:
- [x] PostgreSQL 14+ running
- [x] Go 1.21+ installed
- [x] Database `tenant_admin_db` exists
- [x] Environment variables configured
- [x] Port 8099 available

**Frontend (tenant-admin-frontend)**:
- [ ] Node.js 18+ installed
- [ ] npm 8+ installed
- [ ] Port 3002 available
- [ ] Environment variables configured

**Optional**:
- [ ] Redis running (for caching)
- [ ] Reverse proxy (nginx/Traefik) configured

---

## 🚀 Deployment Steps

### Part 1: Advanced Analytics Dashboard (10 minutes)

**Status**: Backend already exists, only frontend deployment needed

#### Step 1.1: Verify Backend (monitoring-service)

```bash
# Check if monitoring-service is running
curl http://localhost:8092/health

# Expected response:
# {"status":"healthy"}

# Verify MTTR/MTTD endpoints exist
curl http://localhost:8092/api/v1/analytics/dashboard \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Expected: 200 OK with dashboard data
```

**If monitoring-service is not running**:
```bash
cd microservices/monitoring-service
go build -o monitoring-service cmd/main.go
./monitoring-service
```

#### Step 1.2: Frontend Build (tenant-admin-frontend)

```bash
cd microservices/tenant-admin-frontend

# Install dependencies (if not already installed)
npm install

# Verify analytics page exists
ls -la app/admin/analytics/page.tsx
# Expected: File exists

# Verify API client exists
ls -la lib/api/analytics.ts
# Expected: File exists

# Build frontend
npm run build

# Expected output:
# ✓ Compiled successfully
# Route (app)                                 Size     First Load JS
# ...
# ○ /admin/analytics                          XX kB    XXX kB
```

#### Step 1.3: Start/Restart Frontend

**Development**:
```bash
npm run dev
# Opens on http://localhost:3002
```

**Production**:
```bash
# Using PM2 (recommended)
pm2 start npm --name "tenant-admin-frontend" -- start
pm2 save
pm2 startup

# Or using Docker
docker build -t tenant-admin-frontend .
docker run -d -p 3002:3002 tenant-admin-frontend
```

#### Step 1.4: Verify Analytics Dashboard

```bash
# 1. Login to tenant admin
open http://anupam.localhost:3002/auth/login

# 2. Navigate to Analytics
open http://anupam.localhost:3002/admin/analytics

# 3. Verify elements load:
# - Summary cards (Total Incidents, Avg MTTR, Avg MTTD, Uptime)
# - Charts render
# - Period selector works
# - No console errors
```

**Expected Result**: ✅ Analytics dashboard loads with data

---

### Part 2: Service Dependency Mapping (20-30 minutes)

**Status**: Frontend and backend both ready, database migration needed

#### Step 2.1: Database Migration

```bash
cd microservices/tenant-admin-service

# Verify migration file exists
ls -la migrations/009_add_component_dependencies.sql
# Expected: File exists

# Apply migration
PGPASSWORD=postgres psql -h localhost -p 5432 -U postgres -d tenant_admin_db \
  -f migrations/009_add_component_dependencies.sql

# Expected output:
# CREATE TABLE
# CREATE INDEX (multiple)
# CREATE FUNCTION
# CREATE TRIGGER
# CREATE VIEW
# COMMENT (multiple)
```

#### Step 2.2: Verify Database Schema

```bash
# Check dependency_edges table
psql -U postgres -d tenant_admin_db -c "\d dependency_edges"

# Expected output:
# Table "public.dependency_edges"
# Column              | Type      | ...
# --------------------+-----------+-----
# id                  | bigint    | ...
# tenant_id           | uuid      | ...
# from_component_id   | uuid      | ...
# to_component_id     | uuid      | ...
# dependency_type     | varchar(20) | ...
# ...

# Check indexes
psql -U postgres -d tenant_admin_db -c "\di dependency_edges*"

# Expected: 8 indexes listed

# Check view
psql -U postgres -d tenant_admin_db -c "\dv component_dependency_graph"

# Expected: View exists
```

#### Step 2.3: Build Backend (tenant-admin-service)

```bash
cd microservices/tenant-admin-service

# Clean build
rm -f tenant-admin-service

# Build
go build -o tenant-admin-service cmd/main.go

# Expected: No errors, binary created
ls -lh tenant-admin-service
# Expected: File ~40-50MB
```

#### Step 2.4: Stop Old Service (if running)

```bash
# Find process
lsof -i :8099

# Kill gracefully
kill $(lsof -t -i:8099)

# Or using pkill
pkill -f tenant-admin-service

# Verify stopped
curl http://localhost:8099/health
# Expected: Connection refused
```

#### Step 2.5: Start New Service

**Development**:
```bash
cd microservices/tenant-admin-service
./tenant-admin-service
```

**Production (with PM2)**:
```bash
cd microservices/tenant-admin-service

# Create PM2 ecosystem file
cat > ecosystem.config.js <<EOF
module.exports = {
  apps: [{
    name: 'tenant-admin-service',
    script: './tenant-admin-service',
    instances: 2,
    exec_mode: 'cluster',
    env: {
      SERVER_PORT: 8099,
      DB_HOST: 'localhost',
      DB_PORT: 5432,
      DB_NAME: 'tenant_admin_db',
      DB_USER: 'postgres',
      DB_PASSWORD: 'postgres',
      JWT_SECRET: process.env.JWT_SECRET,
      ENVIRONMENT: 'production'
    }
  }]
}
EOF

# Start with PM2
pm2 start ecosystem.config.js
pm2 save
```

**Production (with systemd)**:
```bash
sudo tee /etc/systemd/system/tenant-admin-service.service <<EOF
[Unit]
Description=Tenant Admin Service
After=network.target postgresql.service

[Service]
Type=simple
User=beakon
WorkingDirectory=/opt/beakon/tenant-admin-service
ExecStart=/opt/beakon/tenant-admin-service/tenant-admin-service
Restart=always
RestartSec=5
Environment=SERVER_PORT=8099
Environment=DB_HOST=localhost
Environment=DB_PORT=5432
Environment=DB_NAME=tenant_admin_db

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable tenant-admin-service
sudo systemctl start tenant-admin-service
sudo systemctl status tenant-admin-service
```

#### Step 2.6: Verify Backend Health

```bash
# Health check
curl http://localhost:8099/health

# Expected:
# {"status":"healthy","timestamp":"2025-01-25T..."}

# Check dependency endpoints (requires auth)
TOKEN="your_jwt_token_here"

curl http://localhost:8099/api/v1/dependencies/graph \
  -H "Authorization: Bearer $TOKEN"

# Expected: 200 OK with graph data (may be empty initially)
```

#### Step 2.7: Frontend Deployment (if not already done)

```bash
cd microservices/tenant-admin-frontend

# Verify dependencies page exists
ls -la app/admin/dependencies/page.tsx
# Expected: File exists

# Verify API client exists
ls -la lib/api/dependencies.ts
# Expected: File exists

# Verify D3 is installed
npm list d3
# Expected: d3@7.x.x

# Rebuild (if needed)
npm run build
```

#### Step 2.8: Restart Frontend (if needed)

```bash
# Development
npm run dev

# Production (PM2)
pm2 restart tenant-admin-frontend

# Production (systemd)
sudo systemctl restart tenant-admin-frontend
```

#### Step 2.9: Verify Dependency Mapping UI

```bash
# 1. Login to tenant admin
open http://anupam.localhost:3002/auth/login

# 2. Navigate to Dependencies
open http://anupam.localhost:3002/admin/dependencies

# 3. Verify elements load:
# - Dependency graph renders (D3.js SVG)
# - Overview stats cards display
# - "Add Dependency" button works
# - No console errors
```

**Expected Result**: ✅ Dependency graph page loads (empty graph is OK if no dependencies yet)

---

## ✅ Post-Deployment Verification

### Functional Tests

#### Test 1: Create Test Components (if needed)

```bash
TOKEN="your_jwt_token_here"
API_URL="http://localhost:8099/api/v1"

# Create test components
API_UUID=$(curl -X POST $API_URL/components \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"API Gateway","status":"operational","group_name":"Core"}' \
  | jq -r '.id')

DB_UUID=$(curl -X POST $API_URL/components \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Database","status":"operational","group_name":"Core"}' \
  | jq -r '.id')

CACHE_UUID=$(curl -X POST $API_URL/components \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Cache","status":"operational","group_name":"Core"}' \
  | jq -r '.id')

echo "Created components:"
echo "API Gateway: $API_UUID"
echo "Database: $DB_UUID"
echo "Cache: $CACHE_UUID"
```

#### Test 2: Add Dependencies

```bash
# API depends on Database (hard dependency)
curl -X POST $API_URL/dependencies \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"from_component_id\": \"$API_UUID\",
    \"to_component_id\": \"$DB_UUID\",
    \"dependency_type\": \"hard\"
  }"

# Expected: {"message":"Dependency created successfully"}

# API depends on Cache (soft dependency)
curl -X POST $API_URL/dependencies \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"from_component_id\": \"$API_UUID\",
    \"to_component_id\": \"$CACHE_UUID\",
    \"dependency_type\": \"soft\"
  }"

# Expected: {"message":"Dependency created successfully"}
```

#### Test 3: Verify Dependency Graph

```bash
# Get dependency graph
curl -X GET $API_URL/dependencies/graph \
  -H "Authorization: Bearer $TOKEN" | jq

# Expected: Graph with 3 nodes and 2 edges
```

#### Test 4: Test Impact Analysis

```bash
# Analyze impact of Database failure
curl -X GET $API_URL/dependencies/impact/$DB_UUID \
  -H "Authorization: Bearer $TOKEN" | jq

# Expected:
# {
#   "component_id": "...",
#   "component_name": "Database",
#   "direct_impact_count": 1,
#   "total_impact_count": 1,
#   "affected_components": [
#     {
#       "component_name": "API Gateway",
#       "impact_level": "critical"
#     }
#   ],
#   "mitigation_suggestions": [...]
# }
```

#### Test 5: Test Circular Dependency Prevention

```bash
# Try to create cycle: Database -> API (should fail)
curl -X POST $API_URL/dependencies \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"from_component_id\": \"$DB_UUID\",
    \"to_component_id\": \"$API_UUID\",
    \"dependency_type\": \"hard\"
  }"

# Expected: {"error":"Failed to add dependency","message":"circular dependency detected..."}
```

#### Test 6: Verify Analytics Dashboard

```bash
# Access analytics dashboard in browser
open http://anupam.localhost:3002/admin/analytics

# Verify:
# - [x] Summary cards display
# - [x] Period selector works (week/month/quarter/year)
# - [x] Charts render (may be empty if no incidents yet)
# - [x] Refresh button works
# - [x] Tab navigation works (Overview/Incidents/Trends/Monitors)
```

#### Test 7: Verify Dependency UI

```bash
# Access dependencies page in browser
open http://anupam.localhost:3002/admin/dependencies

# Verify:
# - [x] D3.js graph renders with 3 nodes
# - [x] Nodes are draggable
# - [x] Zoom/pan works
# - [x] Click node shows impact analysis
# - [x] Overview stats display (3 components, 2 dependencies)
# - [x] "Add Dependency" dialog works
```

---

## 🔧 Troubleshooting

### Issue 1: Backend Build Fails

**Error**: `undefined: models.DependencyEdge`

**Solution**:
```bash
cd microservices/tenant-admin-service

# Verify dependency files exist
ls -la internal/models/dependency.go
ls -la internal/services/dependency_service.go
ls -la internal/handlers/dependency_handler.go

# If missing, files need to be created
# Clean Go cache
go clean -modcache
go mod tidy
go build -o tenant-admin-service cmd/main.go
```

### Issue 2: Frontend Build Fails

**Error**: `Module not found: Can't resolve 'd3'`

**Solution**:
```bash
cd microservices/tenant-admin-frontend

# Install D3
npm install d3 @types/d3

# Clear cache and rebuild
rm -rf .next
npm run build
```

### Issue 3: Database Migration Fails

**Error**: `relation "dependency_edges" already exists`

**Solution**:
```bash
# Migration already applied, skip this step
# Or if you need to reapply:
psql -U postgres -d tenant_admin_db -c "DROP TABLE IF EXISTS dependency_edges CASCADE;"
psql -U postgres -d tenant_admin_db -f migrations/009_add_component_dependencies.sql
```

### Issue 4: API Returns 404 for /dependencies/*

**Error**: `404 Not Found`

**Solution**:
```bash
# Check if service restarted after code changes
ps aux | grep tenant-admin-service

# If old process, kill and restart
pkill -f tenant-admin-service
./tenant-admin-service

# Verify routes registered
curl http://localhost:8099/health
```

### Issue 5: D3.js Graph Not Rendering

**Symptoms**: Blank page or console errors

**Solution**:
```bash
# Open browser console (F12)
# Check for errors

# Common fixes:
# 1. Verify D3 imported correctly
# 2. Check SVG element exists in DOM
# 3. Verify API returns valid graph data
# 4. Check browser console for CORS errors

# Test API directly
curl http://localhost:8099/api/v1/dependencies/graph \
  -H "Authorization: Bearer $TOKEN"
```

### Issue 6: "Circular dependency" Error on Valid Dependency

**Symptoms**: False positive cycle detection

**Solution**:
```bash
# Check existing dependencies
curl http://localhost:8099/api/v1/dependencies/graph \
  -H "Authorization: Bearer $TOKEN" | jq '.edges'

# Look for unexpected edges
# If found, remove incorrect edge:
curl -X DELETE http://localhost:8099/api/v1/dependencies/$FROM_UUID/$TO_UUID \
  -H "Authorization: Bearer $TOKEN"
```

---

## 🔒 Security Checklist

### Pre-Production

- [ ] JWT_SECRET is strong (32+ characters)
- [ ] Database password changed from default
- [ ] CORS configured for production domain
- [ ] Rate limiting enabled
- [ ] HTTPS enforced (if deploying to public internet)
- [ ] Database SSL enabled (if production)
- [ ] API endpoints require authentication
- [ ] Tenant isolation verified (users can't access other tenants' data)

### Production Hardening

```bash
# Generate strong JWT secret
openssl rand -base64 64

# Update environment
export JWT_SECRET="your_generated_secret_here"

# Enable rate limiting (already in shared-resilience)
# Check config:
grep -r "RateLimit" microservices/tenant-admin-service/cmd/main.go

# Enable database SSL (production only)
export DB_SSLMODE=require

# Use environment-specific config
export ENVIRONMENT=production
```

---

## 📊 Monitoring & Alerts

### Metrics to Watch

**Backend (tenant-admin-service)**:
- Response time: `/api/v1/dependencies/graph` should be <100ms
- Error rate: Should be <1%
- Active connections: Monitor database pool usage
- Memory usage: Should be stable ~50-100MB

**Frontend (tenant-admin-frontend)**:
- Page load time: `/admin/dependencies` should load in <2 seconds
- D3.js render time: Graph should render in <500ms
- API call latency: Check Network tab in browser

**Database**:
- `dependency_edges` table size
- Query performance on indexes
- Connection pool saturation

### Health Check Endpoints

```bash
# Backend
curl http://localhost:8099/health
curl http://localhost:8099/health/ready
curl http://localhost:8099/health/live

# Frontend (if configured)
curl http://localhost:3002/api/health
```

### Log Monitoring

```bash
# Backend logs (if using PM2)
pm2 logs tenant-admin-service

# Backend logs (if using systemd)
sudo journalctl -u tenant-admin-service -f

# Check for errors
grep -i "error" /var/log/tenant-admin-service.log

# Check dependency operations
grep -i "dependency" /var/log/tenant-admin-service.log
```

---

## 🔄 Rollback Procedure

### If Deployment Fails

**Step 1: Rollback Backend**
```bash
# Stop new service
pm2 stop tenant-admin-service

# Or systemd
sudo systemctl stop tenant-admin-service

# Restore old binary (if backed up)
cp tenant-admin-service.backup tenant-admin-service

# Restart
pm2 start tenant-admin-service
# Or
sudo systemctl start tenant-admin-service
```

**Step 2: Rollback Database (if needed)**
```bash
# IMPORTANT: Only if migration caused issues

# Drop new table
psql -U postgres -d tenant_admin_db -c "DROP TABLE IF EXISTS dependency_edges CASCADE;"

# Drop view
psql -U postgres -d tenant_admin_db -c "DROP VIEW IF EXISTS component_dependency_graph;"

# Service will continue to work without dependency features
```

**Step 3: Rollback Frontend**
```bash
cd microservices/tenant-admin-frontend

# Checkout previous version
git checkout HEAD~1 app/admin/dependencies/
git checkout HEAD~1 app/admin/analytics/
git checkout HEAD~1 lib/api/dependencies.ts
git checkout HEAD~1 lib/api/analytics.ts

# Rebuild
npm run build

# Restart
pm2 restart tenant-admin-frontend
```

---

## ✅ Deployment Checklist

### Pre-Deployment

- [x] Code reviewed and tested locally
- [x] Database migration prepared
- [x] Backup of current deployment
- [x] Maintenance window scheduled (if needed)
- [ ] Team notified of deployment
- [ ] Rollback plan documented

### During Deployment

**Advanced Analytics** (10 min):
- [ ] Verify monitoring-service running
- [ ] Frontend build successful
- [ ] Frontend restarted
- [ ] Analytics page loads

**Dependency Mapping** (20-30 min):
- [ ] Database migration applied successfully
- [ ] Database schema verified (table + indexes + view)
- [ ] Backend build successful
- [ ] Backend restarted
- [ ] Health checks passing
- [ ] Frontend build successful (if needed)
- [ ] Frontend restarted (if needed)
- [ ] Dependencies page loads
- [ ] D3.js graph renders

### Post-Deployment

- [ ] All functional tests pass
- [ ] No errors in logs
- [ ] Metrics look normal
- [ ] Users can access new features
- [ ] Documentation updated
- [ ] Team notified of completion

---

## 📞 Support

### If Issues Occur

**1. Check Logs**:
```bash
# Backend
pm2 logs tenant-admin-service --lines 100

# Frontend
pm2 logs tenant-admin-frontend --lines 100

# Database
psql -U postgres -d tenant_admin_db -c "SELECT * FROM dependency_edges LIMIT 10;"
```

**2. Verify Services**:
```bash
# Backend
curl http://localhost:8099/health

# Frontend
curl http://localhost:3002
```

**3. Contact**:
- Check GitHub Issues: https://github.com/anupamdutta5/statuspage
- Review documentation in this repo

---

## 🎉 Success Criteria

Deployment is successful when:

- ✅ Backend health check returns healthy
- ✅ Frontend loads without errors
- ✅ Analytics dashboard displays (even if empty)
- ✅ Dependency graph renders (even if empty)
- ✅ Can create dependencies via UI
- ✅ Can view impact analysis
- ✅ No errors in console or logs
- ✅ All API endpoints return 200 or expected status
- ✅ Circular dependency prevention works

---

**Estimated Total Time**: 30-40 minutes
**Risk Level**: Low (zero-downtime deployment)
**Rollback Time**: <5 minutes

**Document Version**: 1.0
**Last Updated**: January 2025
**Status**: Ready for Production Deployment ✅

