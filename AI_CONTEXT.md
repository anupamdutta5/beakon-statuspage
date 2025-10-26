# AI Context - Quick Reference Guide

**Purpose**: Fast AI onboarding with critical system knowledge and common gotchas
**Read this FIRST** when starting new sessions or working without recent context
**Token Budget**: ~100 lines for 80% of critical knowledge

---

## 🎯 System Overview (30 seconds)

**Platform**: Multi-tenant SaaS status page (like Statuspage.io)
**Architecture**: 19 microservices + Shared resilience library (Direct service communication, no API Gateway)
**Configuration**: YAML-first with .env secrets (October 2025 standardization)
**Language**: Go (Gin framework)
**Databases**: 14 PostgreSQL databases (database-per-service pattern)
**Multi-tenancy**: Row-level isolation with UUID tenant_ids

---

## 🚨 Critical Port Allocation (MEMORIZE THIS!)

| Service | Port | Database | Status |
|---------|------|----------|--------|
| **api-gateway** | 8080 | None | ⚠️ **DEPRECATED** |
| **user-service** | 8081 | statuspage_user | ✅ Auth |
| **component-service** | 8084 | statuspage_component | ✅ Active |
| **notification-service** | 8085 | statuspage_notification | ✅ Active |
| **incident-service** | 8086 | statuspage_incident | ✅ Active |
| **payment-service** | 8088 | statuspage_payment | ✅ Active |
| **analytics-service** | 8090 | statuspage_analytics | ✅ Active |
| **monitoring-service** | 8092 | statuspage_monitoring | ✅ Active |
| **status-ui-service** | 8093 | None | ✅ Active |
| **database-service** | 8095 | statuspage_database | ⚠️ **DEPRECATED** |
| **event-store-service** | 8096 | statuspage_event_store | ✅ Active |
| **branding-service** | 8097 | statuspage_branding | ✅ Active |
| **saas-admin-service** | 8098 | saas_admin | ✅ **Platform Admin** |
| **tenant-admin-service** | 8099 | tenant_admin_db | ✅ **Tenant Mgmt** |
| **landing-page-service** | 8100 | statuspage_landing | ✅ Active |

**Consumers** (no HTTP ports): analytics-consumer, audit-consumer, billing-consumer, notification-consumer

---

## ⚠️ Common Gotchas (DON'T MAKE THESE MISTAKES!)

### 1. Service Consolidation
- ❌ **OLD**: `tenant-service` (port 8082)
- ✅ **NEW**: Merged into `tenant-admin-service` (port 8099)
- **Why**: Consolidated tenant management into single service
- **Impact**: Port 8082 is FREE, do not reuse without checking

### 2. Deprecated Services
- ❌ `database-service` (port 8095) - **DO NOT USE**
- **Status**: Active but unused, scheduled for removal
- **Reason**: Functionality moved to individual services

### 3. Admin Service Confusion (CRITICAL!)
- **SaaS Admin (8098)**: Platform-wide super admin (manages ALL tenants)
  - Database: `saas_admin` (independent database)
  - Users: Beakon platform administrators
  - Features: Subscription plans, features, pricing, tenant creation via API

- **Tenant Admin (8099)**: Individual tenant admin dashboard (ONE tenant)
  - Database: `tenant_admin_db` (separate database)
  - Users: Customer's own admins
  - Features: User mgmt, status page config, RBAC, branding, teams

- **Key Difference**: Scope (platform vs tenant), databases are SEPARATE
- **Communication**: SaaS Admin creates tenants by calling Tenant Admin API

### 4. Database Schema Issues
- ❌ **WRONG**: `last_seen_at` column
- ✅ **CORRECT**: `last_seen` column (sessions table)
- **Why**: Historical schema uses `last_seen`, GORM model must match
- **Location**: `tenant-admin-service/internal/models/rbac.go:140`

### 5. Tenant ID Data Types
- ❌ **WRONG**: `uint` or `int` for tenant_id
- ✅ **CORRECT**: `string` (UUID format: "72ffbda4-fd0d-4448-854a-935e36d32b90")
- **Why**: Multi-tenant architecture uses UUIDs for tenant isolation
- **Impact**: Type mismatches cause database errors

### 6. Redis Configuration (tenant-admin-service ONLY)
- ❌ **WRONG**: Assume Redis is required for sessions
- ✅ **CORRECT**: Redis is DISABLED, uses database-only sessions
- **Why**: Simplified deployment, removed Redis dependency
- **Location**: `cmd/main.go:209` - hardcoded `if false`
- **Impact**: Sessions stored in PostgreSQL, not Redis

### 7. Session Auto-Creation (tenant-admin-service)
- ✅ **NEW FEATURE**: Login automatically creates RBAC session
- **When**: After successful JWT generation (non-critical)
- **Response**: Includes `session_id` field in login response
- **Impact**: No separate session creation API call needed

### 8. Database Naming Patterns
- **Pattern 1**: `statuspage_<service_name>` (most services)
- **Pattern 2**: `saas_admin` (saas-admin-service only)
- **Pattern 3**: `tenant_admin_db` (tenant-admin-service only)
- **Example**: `statuspage_analytics`, `statuspage_user`, etc.
- **Architecture**: Database-per-service (pure microservices pattern, NO shared databases)

### 9. Authentication & Session Management (NEW!)
- ✅ **Pattern**: Hybrid JWT + Refresh Tokens (OAuth 2.0)
- **Access Token**: JWT, 15 minutes TTL (stateless)
- **Refresh Token**: UUID, 7 days TTL (stored in PostgreSQL)
- **SaaS Admin**: FULLY IMPLEMENTED, NO Tenant Admin dependency
- **Tenant Admin**: Partial (models ready, handler pending)
- **Guide**: See [AUTHENTICATION_GUIDE.md](AUTHENTICATION_GUIDE.md)

**Critical**: SaaS Admin login works WITHOUT Tenant Admin running! ✨

---

## 🏗️ Service Architecture

### Request Flow
```
Client Request
    ↓
API Gateway (8080) [JWT auth, routing, rate limiting]
    ↓
    ├→ User Service (8081) [Authentication, user profiles]
    ├→ Tenant Admin (8099) [Tenant management, RBAC]
    ├→ SaaS Admin (8098) [Platform administration]
    ├→ Component Service (8084) [Status components]
    ├→ Incident Service (8086) [Incident management]
    ├→ Notification Service (8085) [Alerts, emails]
    ├→ Monitoring Service (8092) [Uptime monitoring]
    └→ [Other services...]
```

### Authentication Flow
```
1. User → API Gateway: POST /api/v1/auth/login
2. API Gateway → User Service (8081): Validate credentials
3. User Service → API Gateway: JWT token + user data
4. API Gateway → Tenant Admin (8099): Auto-create RBAC session
5. Tenant Admin → API Gateway: Session ID
6. API Gateway → User: JWT + Session ID
```

### Multi-Tenant Isolation
- **Method**: Row-level security with `tenant_id` column
- **Type**: UUID strings (not integers!)
- **Enforcement**: Middleware extracts tenant context from JWT or query param
- **Database**: Separate tables with tenant_id foreign keys

---

## 📚 Where to Find Detailed Info

### Essential Reading (Read These First!)
1. **[README.md](README.md)** - Project overview
2. **[CLAUDE.md](CLAUDE.md)** - Developer onboarding guide (750 lines)
3. **[microservices/QUICK_START.md](microservices/QUICK_START.md)** - 15-minute setup
4. **[docs/INDEX.md](docs/INDEX.md)** - Complete documentation map

### Quick Lookups
- **All service details**: [SERVICE_CATALOG.md](SERVICE_CATALOG.md) (1261 lines - comprehensive)
- **Database schemas**: [DATABASE_ARCHITECTURE.md](DATABASE_ARCHITECTURE.md)
- **Authentication**: [AUTHENTICATION_GUIDE.md](AUTHENTICATION_GUIDE.md)
- **Deployment**: [DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md)
- **Operations**: [OPERATIONAL_RUNBOOK.md](OPERATIONAL_RUNBOOK.md)

### Service-Specific
- **Service README**: `microservices/<service-name>/README.md`
- **Complex services**: `microservices/<service-name>/ARCHITECTURE.md`
  - tenant-admin-service (RBAC, multi-tenant)
  - saas-admin-service (platform admin)

### Development Guides
- **Quick Setup**: [microservices/QUICK_START.md](microservices/QUICK_START.md) - 15-minute setup
- **Frontend Guide**: [microservices/FRONTEND_GUIDE.md](microservices/FRONTEND_GUIDE.md) - Complete frontend development
- **Testing**: [microservices/docs/testing/TESTING_GUIDE.md](microservices/docs/testing/TESTING_GUIDE.md) - 500+ lines
- **Security**: [microservices/docs/testing/SECURITY_ROADMAP.md](microservices/docs/testing/SECURITY_ROADMAP.md)
- **Inter-service comms**: [microservices/docs/architecture/API_GATEWAY_COMMUNICATION_GUIDE.md](microservices/docs/architecture/API_GATEWAY_COMMUNICATION_GUIDE.md)

---

## 🔄 Recent Major Changes (October 2025)

### 1. Frontend-Backend Separation (Oct 21, 2025)
- **Split Frontends**: SaaS Admin (3001) & Tenant Admin (3002) → Separate Next.js apps
- **Why**: Independent scaling, proper SSR, fast refresh, production-ready Docker builds
- **Middleware Auth**: Server-side authentication prevents auth bypasses
- **Repositories**:
  - https://github.com/anupamdutta5/saas-admin-frontend
  - https://github.com/anupamdutta5/tenant-admin-frontend

### 2. Documentation Consolidation (Oct 21, 2025)
- **Before**: 120+ markdown files (duplicates, outdated, confusing)
- **After**: 25 essential files (clear hierarchy, no duplicates)
- **Archived**: 50+ historical docs (preserved, not deleted)
- **New Files**:
  - [microservices/QUICK_START.md](microservices/QUICK_START.md) - 15-minute setup
  - [microservices/FRONTEND_GUIDE.md](microservices/FRONTEND_GUIDE.md) - Complete frontend guide
  - [docs/INDEX.md](docs/INDEX.md) - Documentation map
- **Structure**: Root → Microservices → Service (clear hierarchy)

### 3. Service Consolidation (Sept 2025)
- Merged `tenant-service` → `tenant-admin-service`
- Reason: Unified tenant management, reduced complexity

### 4. Session Management Fixes (Sept 2025)
- Fixed UUID type mismatches (uint → string)
- Fixed column name (`last_seen_at` → `last_seen`)
- Disabled Redis, using database-only sessions
- Added auto-session creation on login

### 5. RBAC Implementation (Sept 2025)
- Full role-based access control in tenant-admin-service
- Session validation, role management, permission system
- Audit logging for all RBAC operations

---

## 🎓 Key Architectural Patterns

### 1. Clean Architecture
- **Layers**: Handlers → Services → Models/DB
- **Pattern**: Dependency injection via constructors
- **Example**: `tenant-admin-service/internal/`

### 2. Shared Resilience Library
- **Location**: `/microservices/shared-resilience`
- **Adoption**: 100% (all services use it)
- **Features**: Circuit breakers, rate limiting, caching, database utils

### 3. API Gateway Pattern
- **Port**: 8080 (all requests go here first)
- **Functions**: Routing, auth, rate limiting, CORS
- **Downstream**: Routes to all 19 HTTP services

### 4. Event-Driven Consumers
- **Consumers**: 4 services (analytics, audit, billing, notification)
- **Pattern**: Kafka consumers (no HTTP endpoints)
- **Purpose**: Async processing, event sourcing

---

## 🚀 Quick Start Checklist

When working on a service:
1. ✅ Check port number (is it correct?)
2. ✅ Check if service is deprecated (database-service!)
3. ✅ Verify tenant_id is UUID string (not int!)
4. ✅ Check database column names (especially sessions table)
5. ✅ Understand service boundary (saas-admin vs tenant-admin)
6. ✅ Check Redis configuration (tenant-admin has it disabled)
7. ✅ Read service README for API endpoints
8. ✅ Check SERVICE_CATALOG.md for dependencies

---

## 💡 Pro Tips for AI Assistants

### Before Making Changes:
1. **Read this file first** (you're doing it! ✅)
2. Check SERVICE_CATALOG.md for service details
3. Read service README for API specifics
4. Check ARCHITECTURE.md if service is complex

### Common Tasks:
- **Port changes**: Update SERVICE_CATALOG.md + service README
- **API changes**: Update SERVICE_CATALOG.md + service README
- **Database changes**: Update DATABASE_ARCHITECTURE.md
- **New features**: Document in service ARCHITECTURE.md (if complex)

### Avoid:
- ❌ Creating duplicate service catalogs
- ❌ Creating session-specific test reports
- ❌ Creating feature-specific docs (put in ARCHITECTURE.md)
- ❌ Modifying database-service (deprecated!)

---

**Last Updated**: October 21, 2025
**Total Services**: 21 active (19 backend + 2 frontend + 1 deprecated + 4 consumers + 1 shared lib)
**Documentation Files**: 25 essential files (50+ archived, see [docs/INDEX.md](docs/INDEX.md))
