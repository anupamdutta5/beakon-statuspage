# Frontend Access Guide

## Overview

This guide explains how to access both frontend applications in the Beakon Status Page Platform after the frontend-backend split.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         User Browser                             │
└──────────────────┬─────────────────────┬────────────────────────┘
                   │                     │
                   │                     │
         ┌─────────▼─────────┐  ┌────────▼──────────┐
         │  SaaS Admin       │  │  Tenant Admin     │
         │  Frontend         │  │  Frontend         │
         │  Port 3001        │  │  Port 3002        │
         │                   │  │  (subdomain)      │
         └─────────┬─────────┘  └────────┬──────────┘
                   │                     │
                   │ CORS                │ Same-Origin
                   │                     │
         ┌─────────▼─────────┐  ┌────────▼──────────┐
         │  SaaS Admin       │  │  Tenant Admin     │
         │  Backend          │  │  Backend          │
         │  Port 8098        │  │  Port 8099        │
         └───────────────────┘  └───────────────────┘
```

## Prerequisites

### 1. Install Node.js Dependencies

**SaaS Admin Frontend:**
```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/saas-admin-frontend
npm install
```

**Tenant Admin Frontend:**
```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/tenant-admin-frontend
npm install
```

### 2. Configure Environment Variables

**SaaS Admin Frontend** (`.env.local`):
```bash
# Already configured - no changes needed
NEXT_PUBLIC_API_URL=http://localhost:8098/api/v1
```

**Tenant Admin Frontend** (`.env.local`):
```bash
# Already configured - no changes needed
NEXT_PUBLIC_API_URL=http://localhost:8099/api/v1
```

### 3. DNS Configuration for Subdomain Routing

**macOS/Linux:**

Edit `/etc/hosts` to add subdomain mappings:
```bash
sudo nano /etc/hosts
```

Add these lines:
```
127.0.0.1 anupam.localhost
127.0.0.1 monday.localhost
127.0.0.1 testsync.localhost
127.0.0.1 e2e-test.localhost
127.0.0.1 rabbitmq-success.localhost
# Add more tenant subdomains as needed
```

**Windows:**

Edit `C:\Windows\System32\drivers\etc\hosts` (as Administrator):
```
127.0.0.1 anupam.localhost
127.0.0.1 monday.localhost
127.0.0.1 testsync.localhost
127.0.0.1 e2e-test.localhost
127.0.0.1 rabbitmq-success.localhost
```

**Note**: You need to add a line for each tenant subdomain you want to access.

## Starting the Applications

### Option 1: Start Complete Stack (Recommended)

**Start everything (backends + frontends):**
```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices
./start-all-services.sh
```

This will:
1. Check prerequisites (Node.js, Go, PostgreSQL, RabbitMQ)
2. Start both backend services (ports 8098, 8099)
3. Start both frontend services (ports 3001, 3002)
4. Display access URLs and log locations

**Stop everything:**
```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices
./stop-all-services.sh
```

### Option 2: Start Frontends Only

If backends are already running:

```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices
./start-all-frontends.sh
```

**Stop frontends:**
```bash
./stop-all-frontends.sh
```

### Option 3: Start Individual Services

**Start SaaS Admin Frontend only:**
```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/saas-admin-frontend
npm run dev
# OR
./start-dev.sh
```

**Start Tenant Admin Frontend only:**
```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/tenant-admin-frontend
npm run dev
# OR
./start-dev.sh
```

## Accessing the Frontends

### 1. SaaS Admin Dashboard

**URL:** http://localhost:3001

**Default Login Credentials:**
```
Email: admin@example.com
Password: admin123
```

**Features Available:**
- Platform overview with stats
- Tenant management (view, create, edit, delete, restore)
- Subscription plans management
- Platform features management
- Pricing tiers configuration
- **Tenant Launch Button** - Opens Tenant Admin panel for specific tenant

**Navigation:**
1. Open http://localhost:3001
2. Login with credentials above
3. Navigate to "Tenants" page
4. See list of all tenants with actions

### 2. Tenant Admin Dashboard

**URL Format:** http://{subdomain}.localhost:3002

**Examples:**
- http://anupam.localhost:3002
- http://monday.localhost:3002
- http://testsync.localhost:3002
- http://e2e-test.localhost:3002

**Default Login Credentials (per tenant):**
```
Email: admin@{tenant-domain}.com
Password: (varies by tenant)
```

**For "monday" tenant:**
```
Email: monday@gmail.com
Password: test
```

**Features Available:**
- Tenant-specific dashboard
- Component management
- Incident management
- Subscriber management
- User management (RBAC)
- Tenant settings

**Navigation:**
1. Open http://{subdomain}.localhost:3002
2. Login with tenant-specific credentials
3. Access tenant admin features

### 3. Tenant Launch Flow (SaaS Admin → Tenant Admin)

This is the **primary integration point** between the two frontends:

**Step-by-step:**

1. **Login to SaaS Admin:**
   - Open http://localhost:3001
   - Login with admin credentials

2. **Navigate to Tenants:**
   - Click "Tenants" in sidebar
   - See list of all tenants

3. **Launch Tenant Admin Panel:**
   - Find a tenant (e.g., "anupam", "monday")
   - Click the **"Launch Tenant Admin Panel"** button (rocket icon)
   - A new tab opens: http://{subdomain}.localhost:3002

4. **Login to Tenant Admin:**
   - Enter tenant-specific credentials
   - Access tenant dashboard

**Button Functionality:**
```typescript
// When clicked, opens:
const subdomain = tenant.subdomain || tenant.name.toLowerCase().replace(/\s+/g, '-');
const url = `http://${subdomain}.localhost:3002`; // Frontend on port 3002
window.open(url, '_blank');
```

**✓ VERIFIED:** The launch button now correctly points to port 3002 (frontend) instead of port 8099 (backend API).

## Port Allocation Summary

| Service | Port | URL |
|---------|------|-----|
| **SaaS Admin Frontend** | 3001 | http://localhost:3001 |
| **Tenant Admin Frontend** | 3002 | http://{subdomain}.localhost:3002 |
| **SaaS Admin Backend** | 8098 | http://localhost:8098/api/v1 |
| **Tenant Admin Backend** | 8099 | http://localhost:8099/api/v1 |

## Testing the Complete Flow

### End-to-End Test

1. **Start all services:**
   ```bash
   cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices
   ./start-all-services.sh
   ```

2. **Verify backends are running:**
   ```bash
   curl http://localhost:8098/api/v1/health
   curl http://localhost:8099/health
   ```

3. **Access SaaS Admin:**
   - Open http://localhost:3001
   - Login: `admin@example.com` / `admin123`

4. **Create a new tenant (optional):**
   - Click "Tenants" → "Create Tenant"
   - Fill in details:
     - Name: Test Tenant
     - Slug: test-tenant
     - Domain: test-tenant.example.com
     - Subdomain: test-tenant
     - Contact Email: test@example.com
   - Click "Create Tenant"

5. **Add subdomain to hosts file:**
   ```bash
   sudo nano /etc/hosts
   # Add: 127.0.0.1 test-tenant.localhost
   ```

6. **Launch tenant admin:**
   - In tenant list, click "Launch Tenant Admin Panel" button
   - New tab opens: http://test-tenant.localhost:3002
   - Login with tenant credentials

7. **Verify tenant isolation:**
   - Each tenant sees only their own data
   - Subdomain determines which tenant's data is displayed

## Troubleshooting

### Frontend Not Loading

**Check if frontend is running:**
```bash
lsof -i :3001  # SaaS Admin
lsof -i :3002  # Tenant Admin
```

**Check logs:**
```bash
tail -f /tmp/saas-admin-frontend.log
tail -f /tmp/tenant-admin-frontend.log
```

**Restart frontend:**
```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices
./stop-all-frontends.sh
./start-all-frontends.sh
```

### Backend API Errors

**Check if backends are running:**
```bash
lsof -i :8098  # SaaS Admin Backend
lsof -i :8099  # Tenant Admin Backend
```

**Check backend logs:**
```bash
tail -f /tmp/saas-admin-service.log
tail -f /tmp/tenant-admin-service.log
```

**Restart backends:**
```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices
./stop-all-backends.sh
./start-all-backends.sh
```

### CORS Errors (SaaS Admin)

If you see CORS errors in browser console:

1. **Verify backend CORS config** in `saas-admin-service/cmd/main.go`:
   ```go
   AllowedOrigins: []string{"http://localhost:3001"},
   AllowCredentials: true,
   ```

2. **Restart backend:**
   ```bash
   cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices
   ./stop-all-backends.sh
   ./start-all-backends.sh
   ```

### Subdomain Not Working (Tenant Admin)

**Issue:** `http://anupam.localhost:3002` shows error

**Fix 1:** Add to hosts file:
```bash
sudo nano /etc/hosts
# Add: 127.0.0.1 anupam.localhost
```

**Fix 2:** Clear DNS cache (macOS):
```bash
sudo dscacheutil -flushcache
sudo killall -HUP mDNSResponder
```

**Fix 3:** Use wildcard DNS (advanced):
- Install `dnsmasq` for automatic wildcard subdomain routing
- Configure `*.localhost` to resolve to `127.0.0.1`

### Tenant Launch Button Opens Wrong URL

**Symptom:** Clicking launch button opens API endpoint instead of frontend

**Fix:** Already resolved - button now points to port 3002 (frontend)

**Verify in code** (`saas-admin-frontend/app/admin/tenants/page.tsx`):
```typescript
const url = `http://${subdomain}.localhost:3002`; // Should be 3002, not 8099
```

### Port Already in Use

**Kill process on specific port:**
```bash
# SaaS Admin Frontend
lsof -ti :3001 | xargs kill -9

# Tenant Admin Frontend
lsof -ti :3002 | xargs kill -9
```

## Development Workflow

### Making Frontend Changes

1. **Edit files** in `saas-admin-frontend/` or `tenant-admin-frontend/`
2. **Next.js hot reload** automatically updates the browser
3. **No restart needed** for most changes

### Making Backend Changes

1. **Edit files** in `saas-admin-service/` or `tenant-admin-service/`
2. **Restart backend:**
   ```bash
   cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices
   ./stop-all-backends.sh
   ./start-all-backends.sh
   ```

### Building for Production

**SaaS Admin Frontend:**
```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/saas-admin-frontend
npm run build
npm start
```

**Tenant Admin Frontend:**
```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/tenant-admin-frontend
npm run build
npm start
```

## Summary of Changes Made

### ✓ Tenant Launch Button Fixed

**File:** `/microservices/saas-admin-frontend/app/admin/tenants/page.tsx`

**Changes:**
1. **Line 102-107:** Updated `launchTenantAdmin` function to use port 3002
2. **Line 612:** Updated tenant details modal URL display to use port 3002

**Before:**
```typescript
const url = `http://${subdomain}.localhost:8099`; // Backend API
```

**After:**
```typescript
const url = `http://${subdomain}.localhost:3002`; // Frontend
```

**Result:** Clicking "Launch Tenant Admin Panel" now correctly opens the Tenant Admin frontend instead of the backend API.

## Quick Reference

### Start Everything
```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices
./start-all-services.sh
```

### Access URLs
- **SaaS Admin:** http://localhost:3001
- **Tenant Admin:** http://{subdomain}.localhost:3002

### Stop Everything
```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices
./stop-all-services.sh
```

### View Logs
```bash
# Frontend logs
tail -f /tmp/saas-admin-frontend.log
tail -f /tmp/tenant-admin-frontend.log

# Backend logs
tail -f /tmp/saas-admin-service.log
tail -f /tmp/tenant-admin-service.log
```

## Additional Resources

- **SaaS Admin Frontend README:** `/microservices/saas-admin-frontend/README.md`
- **Tenant Admin Frontend README:** `/microservices/tenant-admin-frontend/README.md`
- **Service Catalog:** `/SERVICE_CATALOG.md`
- **Architecture:** `/ARCHITECTURE.md`
- **Claude Instructions:** `/CLAUDE.md`
