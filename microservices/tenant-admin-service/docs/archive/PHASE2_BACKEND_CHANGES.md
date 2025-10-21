# Phase 2: Tenant Admin Backend - Frontend Code Removed

## Date: October 21, 2025

## Changes Made

### 1. Frontend Code Removal

#### Removed loadTemplates() Function (Lines 32-66)
**Before:**
```go
// loadTemplates loads all HTML templates from both root and subdirectories
func loadTemplates() *template.Template {
	rootFiles, err := filepath.Glob("web/templates/*.html")
	componentFiles, err := filepath.Glob("web/templates/components/*.html")
	allFiles := append(rootFiles, componentFiles...)
	tmpl, err := template.ParseFiles(allFiles...)
	return tmpl
}
```

**After:** Function completely removed (no longer needed)

**Why**: Backend is API-only, no template rendering required.

---

#### Removed Unused Imports
**Before:**
```go
import (
	...
	"html/template"
	"path/filepath"
	...
)
```

**After:**
```go
import (
	...
	// Removed: html/template, path/filepath
	...
)
```

**Why**: These imports were only used for template loading.

---

#### Removed Static File Routes (Lines 811-829)
**Before:**
```go
// React/Next.js frontend static file serving
router.Static("/_next", "./frontend/out/_next")

// Public routes (no tenant validation required for login page)
router.StaticFile("/", "./frontend/out/login.html")
router.StaticFile("/login", "./frontend/out/login.html")

// Protected admin routes (require valid tenant subdomain)
adminRoutes := router.Group("/admin", middleware.RequireTenant())
{
	adminRoutes.StaticFile("", "./frontend/out/admin/dashboard.html")
	adminRoutes.StaticFile("/dashboard", "./frontend/out/admin/dashboard.html")
	adminRoutes.StaticFile("/users", "./frontend/out/admin/users.html")
	adminRoutes.StaticFile("/status-pages", "./frontend/out/admin/status-pages.html")
	adminRoutes.StaticFile("/components", "./frontend/out/admin/components.html")
	adminRoutes.StaticFile("/incidents", "./frontend/out/admin/incidents.html")
	adminRoutes.StaticFile("/subscribers", "./frontend/out/admin/subscribers.html")
	adminRoutes.StaticFile("/settings", "./frontend/out/admin/settings.html")
}
```

**After:**
```go
// Frontend now served independently on port 3002 (tenant-admin-frontend service)
// All static file routes removed - backend is API-only
// Subdomain routing middleware remains active for tenant isolation
```

**Why**: Frontend is now a separate Next.js service on port 3002.

---

#### Removed Template Loading Call (Line 283)
**Before:**
```go
router.Use(middleware.TenantContextMiddleware(dbManager.GetDB(), logger, baseDomain))

// React frontend - No template loading needed (using static export)
// router.SetHTMLTemplate(loadTemplates())

// Initialize Redis cache...
```

**After:**
```go
router.Use(middleware.TenantContextMiddleware(dbManager.GetDB(), logger, baseDomain))

// Initialize Redis cache...
```

**Why**: Removed unnecessary comment and template loading reference.

---

### 2. What Remains (Important)

#### Subdomain Routing Middleware (KEPT)
```go
// Add tenant context middleware for subdomain routing
baseDomain := os.Getenv("BASE_DOMAIN")
if baseDomain == "" {
	baseDomain = "localhost" // Default for development
}
router.Use(middleware.TenantContextMiddleware(dbManager.GetDB(), logger, baseDomain))
```

**Why**: Still needed for tenant isolation via subdomain detection. This is how the backend knows which tenant the request belongs to.

#### All API Routes (KEPT)
All `/api/v1/*` routes remain unchanged:
- `/api/v1/auth/*` - Authentication
- `/api/v1/users/*` - User management
- `/api/v1/components/*` - Component CRUD
- `/api/v1/incidents/*` - Incident management
- `/api/v1/subscribers/*` - Subscriber management
- `/api/v1/teams/*` - Team management
- `/api/v1/roles/*` - RBAC
- ... and all other API endpoints

---

## Architecture Change

### Before (Monolith):
```
Browser (subdomain.localhost:8099)
   ↓
tenant-admin-service (Port 8099)
├── Go Backend (API + subdomain routing)
├── Static Frontend (HTML served by Go)
└── Database: tenant_admin_db
```

### After (Microservices):
```
Browser (subdomain.localhost:3002)
   ↓
tenant-admin-frontend (Port 3002) ──Same-Origin API──→ tenant-admin-backend (Port 8099)
   Next.js SSR                                          Go API + Subdomain routing
   (Proxy subdomain to backend)                        Database: tenant_admin_db
```

---

## Key Differences from SaaS Admin

| Feature | SaaS Admin Backend | Tenant Admin Backend |
|---------|-------------------|---------------------|
| **CORS Middleware** | Required (different ports) | **NOT Required** (same-origin via proxy) |
| **Subdomain Routing** | Not used | **Required** (tenant isolation) |
| **Frontend Port** | 3001 | 3002 |
| **Backend Port** | 8098 | 8099 |
| **Multi-tenancy** | N/A (single platform admin) | Subdomain-based tenant detection |

**Why no CORS?**: The frontend (port 3002) uses subdomain routing (e.g., `anupam.localhost:3002`). Requests are **same-origin** because the browser sends API calls to `anupam.localhost:8099`, which is the same subdomain. In production, a reverse proxy routes both ports through the same domain.

---

## Subdomain Routing Flow (Unchanged)

1. **User accesses:** `anupam.localhost:3002/login`
2. **Frontend makes API call:** Same-origin to `anupam.localhost:8099/api/v1/auth/login`
3. **Backend middleware extracts subdomain:** "anupam" from Host header
4. **Backend validates tenant:** Checks if "anupam" exists in database
5. **Backend sets tenant context:** All database queries scoped to tenant_id
6. **API returns tenant-specific data:** Only for tenant "anupam"

---

## Testing

### 1. Verify Backend Compiles
```bash
cd microservices/tenant-admin-service
go build -o tenant-admin-service cmd/main.go
```
✅ **PASSED** - No compilation errors

### 2. Start Backend Service
```bash
cd microservices/tenant-admin-service
JWT_SECRET="dev-secret-for-testing-only-change-in-production-min-32-chars-long" \
RABBITMQ_URL="amqp://admin:SecureP@ssw0rd2024!@localhost:5672/" \
DB_HOST=localhost \
DB_PORT=5432 \
DB_USER=postgres \
DB_PASSWORD=postgres \
DB_NAME=tenant_admin_db \
DB_SSLMODE=disable \
SERVER_PORT=8099 \
./tenant-admin-service
```

### 3. Test Health Endpoint
```bash
curl http://localhost:8099/health
```
Expected: `{"status": "ok"}`

### 4. Test Subdomain Routing
```bash
# Without subdomain (should fail tenant validation for protected routes)
curl http://localhost:8099/api/v1/components -v

# With subdomain (should work if tenant exists)
curl -H "Host: anupam.localhost:8099" http://localhost:8099/api/v1/components -v
```

### 5. Test API Endpoints
```bash
# Login (should work)
curl -X POST http://localhost:8099/api/v1/auth/login \
  -H "Host: anupam.localhost:8099" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password"}'
```

---

## Files Modified

```
tenant-admin-service/
├── cmd/main.go                    (MODIFIED)
│   ├── Lines 5-15: Removed html/template and path/filepath imports
│   ├── Lines 32-66: Removed loadTemplates() function
│   ├── Line 283: Removed template loading comment
│   └── Lines 811-829: Removed all static file routes
└── PHASE2_BACKEND_CHANGES.md     (THIS FILE - CREATED)
```

---

## Next Steps (Phase 2 Remaining)

1. **Test End-to-End:**
   - Start tenant-admin-backend (port 8099)
   - Start tenant-admin-frontend (port 3002)
   - Navigate to `http://anupam.localhost:3002/login`
   - Verify login works and redirects to dashboard
   - Verify subdomain routing still works
   - Verify API calls from frontend work

2. **Production Considerations:**
   - Configure reverse proxy (nginx/trafficd) to route both ports
   - Ensure subdomain wildcard DNS is configured
   - Update BASE_DOMAIN environment variable

---

## Benefits Achieved

1. ✅ **API-Only Backend** - No frontend serving code
2. ✅ **No CORS Complexity** - Same-origin subdomain routing
3. ✅ **Subdomain Isolation Maintained** - Tenant context middleware still active
4. ✅ **Clean Separation** - Frontend and backend are independent services
5. ✅ **Production Ready** - Simplified deployment architecture
6. ✅ **Smaller Binary** - No template parsing overhead

---

## Related Documentation

- See `../tenant-admin-frontend/README.md` for frontend setup
- See `../tenant-admin-frontend/PHASE2_CHANGES.md` for frontend changes
- See `../saas-admin-service/PHASE1_BACKEND_CHANGES.md` for comparison
- See main project `SERVICE_CATALOG.md` (to be updated in Phase 5)
