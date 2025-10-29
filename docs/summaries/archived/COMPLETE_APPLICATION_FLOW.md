# Beakon Status Page - Complete Application Flow Guide

**Document Purpose**: Comprehensive reference for understanding the complete flow of the Beakon platform
**Created**: 2025-10-29
**Status**: Living Document

---

## Table of Contents

1. [User Types & Authentication Systems](#1-user-types--authentication-systems)
2. [Tenant Provisioning Flow (Complete)](#2-tenant-provisioning-flow-complete)
3. [User Creation & RBAC Flow](#3-user-creation--rbac-flow)
4. [Service Communication Architecture](#4-service-communication-architecture)
5. [RabbitMQ Event-Driven Patterns](#5-rabbitmq-event-driven-patterns)
6. [Critical Business Logic Flows](#6-critical-business-logic-flows)
7. [Database Architecture & Multi-Tenancy](#7-database-architecture--multi-tenancy)
8. [Session Management](#8-session-management)
9. [Frontend-Backend Communication](#9-frontend-backend-communication)
10. [Deployment Flow](#10-deployment-flow)

---

## 1. User Types & Authentication Systems

### 1.1 Three Distinct User Types

The platform has **THREE separate user types** with **THREE independent authentication systems**:

| User Type | Service | Database | Purpose | Created By |
|-----------|---------|----------|---------|------------|
| **SaaS Admin** | saas-admin-service (8098) | saas_admin.saas_admins | Platform super users (Beakon staff) | Database seeding (NOT created via API) |
| **Tenant Admin** | tenant-admin-service (8099) | tenant_admin_db.users (role='owner') | Customer administrators | SaaS Admin via tenant provisioning |
| **Users** | tenant-admin-service (8099) | tenant_admin_db.users | Tenant employees/team members | Tenant Admin (limited by max_users) |

### 1.2 Authentication Systems

#### A. SaaS Admin Authentication (saas-admin-service)

**Status**: ✅ **COMPLETE** - Hybrid JWT + Refresh Tokens

**Login Flow**:
```
1. POST /api/v1/auth/login
   - email: admin@beakon.com
   - password: *******

2. saas-admin-service validates credentials
   - Check saas_admins table
   - Verify password (bcrypt)

3. Generate tokens:
   - Access Token (JWT): 15 minutes
   - Refresh Token (UUID): 7 days
   - Store refresh token in user_sessions table

4. Return both tokens + set HTTP-only cookies

5. Frontend uses JWT for API calls
```

**Database**: `saas_admin.saas_admins` table
**JWT Claims**: admin_id, email, role="admin"
**Session Storage**: user_sessions table in saas_admin database

#### B. Tenant Admin Authentication (tenant-admin-service)

**Status**: ⚠️ **PARTIAL** - JWT implemented, Refresh tokens NOT implemented

**Login Flow**:
```
1. POST /api/v1/auth/login (or /api/v1/public/login)
   - email: admin@acme.com
   - password: *******
   - subdomain: acme (from URL or body)

2. tenant-admin-service validates credentials
   - Check users table WHERE email AND tenant_id (from subdomain)
   - Verify password (bcrypt)

3. Generate JWT token (24 hours - OLD, should be 15 min)
   - Claims: user_id, tenant_id, email, role

4. Create session in sessions table (3-tier: Redis → PostgreSQL → in-memory)

5. Return JWT + set cookie

6. Frontend uses JWT for API calls
```

**Database**: `tenant_admin_db.users` table (with tenant_id scoping)
**JWT Claims**: user_id, tenant_id, email, role
**Session Storage**: Redis (primary) → PostgreSQL (fallback) → In-memory (emergency)

#### C. User Authentication (Regular Users)

**Status**: ✅ **SAME AS TENANT ADMIN** - Uses tenant-admin-service

Regular users authenticate the same way as tenant admins, but with different roles:
- Tenant Admin: role='owner'
- Users: role='admin', 'manager', 'viewer', or custom roles

---

## 2. Tenant Provisioning Flow (Complete)

This is the **MOST CRITICAL BUSINESS FLOW** in the platform.

### 2.1 Overview

**Who**: SaaS Admin (Beakon staff)
**What**: Creates a new tenant (customer) in the platform
**Where**: SaaS Admin Frontend (port 3001) → saas-admin-service (port 8098) → tenant-admin-service (port 8099)
**When**: When onboarding a new customer

### 2.2 Complete Step-by-Step Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      TENANT PROVISIONING FLOW                            │
└─────────────────────────────────────────────────────────────────────────┘

STEP 1: SaaS Admin Initiates Tenant Creation
=========================================
SaaS Admin Frontend (port 3001)
   ↓
   POST http://localhost:8098/api/v1/tenants
   {
     "name": "Acme Corp",
     "contact_email": "contact@acme.com",
     "domain": "acme.com",
     "subdomain": "acme",
     "admin_email": "admin@acme.com",       ← Tenant Admin credentials
     "admin_password": "SecurePass123!",   ← Tenant Admin credentials
     "plan_id": "uuid-of-plan",
     "max_users": 50                       ← Limit set by SaaS Admin
   }


STEP 2: saas-admin-service Processes Request
==========================================
File: microservices/saas-admin-service/internal/handlers/saas_admin_handler.go:1237

A. Validate Request
   - Check tenant name is not empty
   - Check contact_email is not empty

B. Get/Assign Subscription Plan
   - If plan_id provided, use it
   - If not, use default plan (first available)
   - If no plans exist, set plan_id = nil

C. Create Tenant in saas_admin Database
   - Table: saas_admin.saas_tenants
   - Generate UUID for tenant
   - Set status = "active", is_active = true
   - INSERT tenant record

   ✅ CRITICAL: Tenant now exists in saas_admin


STEP 3A: RabbitMQ Event Publishing (PRIMARY PATH)
==============================================
File: microservices/saas-admin-service/internal/handlers/saas_admin_handler.go:1316-1337

IF eventPublisher != nil (RabbitMQ available):

   A. Build Event Metadata
      - source: "saas-admin-service"
      - correlation_id: from request header
      - user_id: from JWT
      - ip_address: client IP
      - user_agent: from request

   B. Create tenant.created Event
      - event_id: UUID
      - event_type: "tenant.created"
      - tenant_id: tenant UUID
      - timestamp: now
      - data: {
          id, name, slug, contact_email, domain, subdomain,
          status, is_active, max_users, created_at, updated_at
        }

      ⚠️ NOTE: admin_email and admin_password are NOT in the event!

   C. Publish to RabbitMQ
      - Exchange: "tenant.events" (topic)
      - Routing Key: "tenant.created"
      - Delivery Mode: Persistent
      - Wait for confirmation (5s timeout)
      - Headers: x-event-type, x-tenant-id, x-correlation-id, x-source

   D. On Success:
      - Log event published
      - Tenant will sync to tenant-admin-service via consumer

   E. On Failure:
      - Log error (non-blocking)
      - Tenant exists in saas_admin but NOT in tenant_admin_db
      - ⚠️ RISK: Tenant admin cannot login!


STEP 3B: HTTP Fallback (IF RABBITMQ UNAVAILABLE)
=============================================
File: microservices/saas-admin-service/internal/handlers/saas_admin_handler.go:1338-1345

ELSE (RabbitMQ unavailable):

   A. Log warning: "Event publisher not configured, falling back to HTTP sync"

   B. Call createTenantInTenantAdminService()
      - Method: POST
      - URL: http://tenant-admin-service:8099/api/v1/public/tenants
      - Body: {
          "slug": tenant.Slug,
          "name": tenant.Name,
          "contact_email": tenant.ContactEmail,
          "domain": tenant.Domain,
          "subdomain": tenant.Subdomain,
          "admin_email": adminEmail,       ← ONLY IN HTTP PATH!
          "admin_password": adminPassword   ← ONLY IN HTTP PATH!
        }
      - ❌ NO CIRCUIT BREAKER (issue identified in audit)
      - ❌ NO RETRY LOGIC (issue identified in audit)

   C. On Success:
      - Log success
      - Tenant created in tenant_admin_db
      - Tenant admin user created

   D. On Failure:
      - Log error (non-blocking)
      - Tenant exists in saas_admin but NOT in tenant_admin_db
      - ⚠️ RISK: Tenant admin cannot login!


STEP 4: tenant-admin-service Receives Sync (RabbitMQ Path)
=======================================================
File: microservices/tenant-admin-service/internal/events/consumer.go

A. RabbitMQ Consumer Running in Background
   - Started at service startup (main.go:323-329)
   - Consumes from queue: "tenant.created.queue"
   - QoS: Prefetch 1 (fair dispatch)
   - Manual acknowledgment (no auto-ack)

B. Message Processing
   - Timeout: 30 seconds per message
   - Unmarshal JSON event
   - Log: event_id, event_type, tenant_id, correlation_id

C. Route to Handler
   - File: internal/events/handler.go:27
   - Function: HandleTenantCreated()

D. Handler Logic (Idempotent)
   - Check if tenant already exists (by UUID)
   - If exists: Log "skipping creation (idempotent)", return nil
   - If not exists: Create tenant in tenant_admin_db.tenants

   ⚠️ CRITICAL OBSERVATION: Handler does NOT create admin user!
   - RabbitMQ event does NOT contain admin_email/password
   - Tenant exists but NO admin user created
   - ❌ ISSUE: Tenant admin cannot login via this path!

E. Acknowledgment
   - On success: msg.Ack(false)
   - On failure: msg.Nack(false, true) - requeue
   - On repeated failures: Message goes to DLX (dead-letter exchange)


STEP 4 (Alternative): tenant-admin-service Receives Sync (HTTP Path)
=================================================================
File: microservices/tenant-admin-service/internal/handlers/tenant_admin_handler.go:712

A. Receives POST /api/v1/public/tenants
   - Parse request body
   - Extract: name, slug, contact_email, domain, subdomain, admin_email, admin_password

B. Create Tenant Model
   - If ID provided (from SaaS Admin): Use it (UUID consistency)
   - Set status = "active", is_active = true

C. Insert Tenant into tenant_admin_db.tenants
   - Table: tenant_admin_db.tenants
   - Columns: id, name, slug, contact_email, domain, subdomain, status, is_active

D. Create Admin User (CRITICAL!)
   - Function: CreateAdminUser(tenant.ID, adminEmail, adminPassword)
   - Table: tenant_admin_db.users
   - Fields:
     * tenant_id: tenant UUID
     * email: adminEmail
     * password_hash: bcrypt(adminPassword)
     * role: "owner"
     * is_active: true

   ✅ SUCCESS: Admin user created, can login

E. Error Handling
   - If admin user creation fails:
     * Tenant still exists in database
     * Return 201 with warning: "Tenant created but failed to create admin user"
     * ⚠️ ISSUE: Tenant exists but admin cannot login


STEP 5: Return Success to SaaS Admin
==================================
File: microservices/saas-admin-service/internal/handlers/saas_admin_handler.go:1347

A. Return HTTP 201 Created
   {
     "status": "success",
     "message": "Tenant created successfully",
     "data": {
       "id": tenant.ID,
       "name": tenant.Name,
       "domain": tenant.Domain,
       "plan": planID,
       "status": tenant.Status
     }
   }

B. Frontend displays success message

C. ⚠️ NOTE: Success returned EVEN IF sync failed!
   - Tenant exists in saas_admin
   - Tenant might NOT exist in tenant_admin_db (if RabbitMQ/HTTP failed)
   - Admin user might NOT exist (if HTTP failed or RabbitMQ path used)
```

### 2.3 Critical Decision Points

| Decision | Implementation | Reasoning | Risk |
|----------|---------------|-----------|------|
| **Non-blocking sync** | Tenant creation succeeds even if sync fails | SaaS Admin sees success, operations can continue | Tenant admin cannot login until manual sync |
| **Dual-mode sync** | RabbitMQ (primary) + HTTP (fallback) | Resilience when RabbitMQ unavailable | Inconsistent behavior between paths |
| **No admin user in RabbitMQ event** | admin_email/password only in HTTP path | Events should be immutable, credentials shouldn't be in events | Tenant admin cannot login via RabbitMQ path |
| **Idempotent event handlers** | Check if tenant exists before creating | Safe to replay events | Good design |
| **No retry on HTTP fallback failure** | Single attempt, log error | Simple implementation | Sync failures not retried |

### 2.4 Sync Failure Scenarios

| Scenario | Result | Impact | Detection | Recovery |
|----------|--------|--------|-----------|----------|
| **RabbitMQ publish fails** | Tenant in saas_admin only | Cannot login | Error in logs | HTTP fallback called automatically |
| **HTTP fallback fails** | Tenant in saas_admin only | Cannot login | Error in logs | Manual sync required |
| **RabbitMQ consumer down** | Event queued | Delayed sync | Queue monitoring | Automatic when consumer restarts |
| **Admin user creation fails (HTTP)** | Tenant exists, no admin | Cannot login | Warning in response | Manual user creation |
| **Network partition during sync** | Tenant in saas_admin only | Cannot login | Timeout in logs | Manual sync or retry |

### 2.5 Identified Issues & Recommendations

#### Issues

1. ❌ **RabbitMQ Path Doesn't Create Admin User**
   - Event doesn't contain admin_email/password
   - Tenant exists but admin cannot login
   - Only HTTP fallback creates admin user

2. ❌ **No Circuit Breaker on HTTP Fallback**
   - Direct http.Client.Do() call
   - No protection against cascade failures

3. ❌ **No Retry Logic**
   - Single attempt on HTTP fallback
   - Network hiccups cause permanent failure

4. ❌ **No Sync Verification Endpoint**
   - Cannot detect out-of-sync tenants
   - No automated reconciliation

5. ⚠️ **Graceful Degradation Can Hide Failures**
   - Returns 201 even if sync fails
   - SaaS Admin sees success, but tenant admin locked out

#### Recommendations

1. **P0: Fix RabbitMQ Path Admin User Creation**
   - Option A: Include encrypted admin_email/password in event
   - Option B: Separate event: tenant.admin.created
   - Option C: Always use HTTP for initial creation, RabbitMQ for updates

2. **P0: Implement Circuit Breakers**
   - Use shared-resilience ServiceClient for HTTP calls
   - Protect against cascade failures

3. **P1: Add HTTP Retry Logic**
   - Exponential backoff: 100ms, 200ms, 400ms
   - Max 3 retries for transient failures

4. **P1: Add Sync Verification Endpoint**
   ```
   GET /api/v1/tenants/verify-sync
   Returns: List of tenants in saas_admin but not in tenant_admin_db
   ```

5. **P2: Add Monitoring & Alerting**
   - Metric: tenant_sync_failures_total{method="rabbitmq|http"}
   - Alert when sync failures > threshold

---

## 3. User Creation & RBAC Flow

### 3.1 User Hierarchy

```
SaaS Admin (Beakon Staff)
   └─ Creates Tenants
         └─ Tenant Admin (role='owner')
               └─ Creates Users (limited by max_users)
                     ├─ Admin (role='admin')
                     ├─ Manager (role='manager')
                     └─ Viewer (role='viewer')
```

### 3.2 User Creation by Tenant Admin

**Flow**:
```
1. Tenant Admin logs in
   - POST /api/v1/auth/login
   - Subdomain: acme.beakon.com
   - JWT contains: user_id, tenant_id, role='owner'

2. Tenant Admin creates user
   - POST /api/v1/users
   - Headers: Authorization: Bearer <JWT>
   - Body: {
       "email": "john@acme.com",
       "name": "John Doe",
       "role": "manager",
       "team_id": "uuid" (optional)
     }

3. tenant-admin-service validates request
   - Extract tenant_id from JWT
   - Check current user count < max_users
   - If limit exceeded: Return 400 "Max users limit reached"

4. Create user in tenant_admin_db.users
   - tenant_id: from JWT (automatic tenant scoping)
   - email: from request
   - password_hash: bcrypt(auto-generated password)
   - role: from request
   - is_active: true
   - created_by: user_id from JWT

5. Assign role via RBAC
   - Table: user_roles
   - user_id: new user ID
   - role_id: role ID from request
   - tenant_id: from JWT

6. Send invitation email
   - Email contains: login link, temporary password
   - User must change password on first login

7. Return success
   {
     "user": {
       "id": "uuid",
       "email": "john@acme.com",
       "role": "manager",
       "tenant_id": "uuid"
     }
   }
```

### 3.3 RBAC System

**Tables** (tenant_admin_db):
- `roles` - Role definitions (owner, admin, manager, viewer, custom)
- `permissions` - Permission catalog (users.read, users.create, incidents.update, etc.)
- `role_permissions` - Which permissions each role has
- `user_roles` - Which roles each user has (many-to-many)
- `teams` - Team definitions
- `team_members` - Team membership
- `team_roles` - Team-level role assignments

**Permission Enforcement**:

File: `internal/middleware/rbac_middleware.go`

```go
// Middleware checks permissions before allowing access
func (m *RBACMiddleware) RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get user_id and tenant_id from JWT (set by AuthMiddleware)
        userID := c.GetString("user_id")
        tenantID := c.GetString("tenant_id")

        // Check if user has permission
        hasPermission, err := m.service.CheckUserPermission(tenantID, userID, permission)
        if err != nil || !hasPermission {
            c.JSON(403, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

**Usage in Routes** (main.go):
```go
// Example: Only users with users.create permission can create users
protected.POST("/users",
    rbacMiddleware.RequirePermission("users.create"),
    userHandler.CreateUser)
```

---

## 4. Service Communication Architecture

### 4.1 Post-October 2025 Architecture

**API Gateway DEPRECATED** on October 26, 2025.

**OLD Pattern** (Pre-Oct 26):
```
Client → API Gateway (8080) → Service
Service → API Gateway (8080) → Service
```

**NEW Pattern** (Post-Oct 26):
```
Client → Direct HTTP → Service
Service → Direct HTTP → Service (with circuit breakers)
```

### 4.2 Service-to-Service Communication Map

#### HTTP Communication (Synchronous)

| Source Service | Destination Service | Endpoint | Purpose | Circuit Breaker? |
|----------------|---------------------|----------|---------|------------------|
| saas-admin (8098) | tenant-admin (8099) | POST /api/v1/public/tenants | Create tenant | ❌ NO (identified issue) |
| saas-admin (8098) | tenant-admin (8099) | GET /api/v1/public/tenants/{id} | Get tenant | ❌ NO |
| saas-admin (8098) | tenant-admin (8099) | PUT /api/v1/public/tenants/{id} | Update tenant | ❌ NO |
| saas-admin (8098) | tenant-admin (8099) | DELETE /api/v1/public/tenants/{id} | Delete tenant | ❌ NO |
| saas-admin (8098) | tenant-admin (8099) | PUT /api/v1/admin/update-credentials | Update tenant owner password | ❌ NO |
| status-ui (8093) | component (8084) | GET /api/v1/components | Fetch component data | ✅ Expected |
| status-ui (8093) | incident (8086) | GET /api/v1/incidents | Fetch incident data | ✅ Expected |
| status-ui (8093) | branding (8097) | GET /api/v1/branding/{tenant_id} | Fetch tenant branding | ✅ Expected |
| status-ui (8093) | analytics (8090) | GET /api/v1/metrics | Fetch uptime metrics | ✅ Expected |
| monitoring (8092) | incident (8086) | POST /api/v1/incidents | Auto-create incident on failure | ✅ Expected |
| monitoring (8092) | component (8084) | PUT /api/v1/components/{id}/status | Update component status | ✅ Expected |
| incident (8086) | notification (8085) | POST /api/v1/notifications | Trigger notifications | ✅ Expected |
| incident (8086) | component (8084) | PUT /api/v1/components/{id}/status | Update component status | ✅ Expected |

**Critical Finding**: saas-admin-service → tenant-admin-service calls do NOT use circuit breakers (identified in audit).

### 4.3 Service Discovery

**Method**: Static configuration via environment variables or service-endpoints.yml

**Example** (service-endpoints.yml):
```yaml
endpoints:
  tenant-admin-service:
    url: ${TENANT_ADMIN_URL:-http://tenant-admin-service:8099}
    timeout: 30s
    retries: 3
    circuit_breaker:
      enabled: true
      threshold: 5
      timeout: 60s
```

**Expected Usage** (from shared-resilience):
```go
client := resilience.NewServiceClient("configs/service-endpoints.yml", logger)

response, err := client.Call(ctx, resilience.ServiceRequest{
    ServiceName: "tenant-admin-service",
    Method:      "POST",
    Path:        "/api/v1/public/tenants",
    Body:        requestData,
})
```

**Actual Usage in saas-admin-service**: ❌ NOT using ServiceClient - direct http.Client.Do()

---

## 5. RabbitMQ Event-Driven Patterns

### 5.1 Event Bus Architecture

**Exchange**: `tenant.events` (type: topic)
**Queues**:
- `tenant.created.queue` → routing key: `tenant.created`
- `tenant.updated.queue` → routing key: `tenant.updated`
- `tenant.deleted.queue` → routing key: `tenant.deleted`
- `tenant.restored.queue` → routing key: `tenant.restored`

**Dead-Letter Exchange**: `tenant.events.dlx` (for failed messages)

### 5.2 Event Schema

**TenantEvent**:
```json
{
  "event_id": "uuid",
  "event_type": "tenant.created",
  "tenant_id": "uuid",
  "timestamp": "2025-10-29T10:00:00Z",
  "metadata": {
    "source": "saas-admin-service",
    "correlation_id": "uuid",
    "user_id": "uuid",
    "ip_address": "192.168.1.1",
    "user_agent": "Mozilla/5.0..."
  },
  "data": {
    "id": "uuid",
    "name": "Acme Corp",
    "slug": "acme",
    "contact_email": "contact@acme.com",
    "domain": "acme.com",
    "subdomain": "acme",
    "status": "active",
    "is_active": true,
    "max_users": 50,
    "created_at": "2025-10-29T10:00:00Z",
    "updated_at": "2025-10-29T10:00:00Z"
  }
}
```

**Critical**: ❌ `admin_email` and `admin_password` are NOT in the event!

### 5.3 Publisher Implementation

**Location**: `microservices/saas-admin-service/internal/events/publisher.go`

**Features**:
- ✅ Publisher confirms enabled
- ✅ Persistent messages (survive broker restart)
- ✅ 5-second confirmation timeout
- ✅ Topic exchange for flexible routing
- ✅ Message headers (event_type, tenant_id, correlation_id, source)
- ✅ Context-aware publishing (respects cancellation)

**Code**:
```go
func (p *Publisher) PublishTenantEvent(ctx context.Context, event TenantEvent) error {
    body, _ := json.Marshal(event)

    publishing := amqp.Publishing{
        ContentType:  "application/json",
        Body:         body,
        DeliveryMode: amqp.Persistent,  // Survives restart
        Timestamp:    event.Timestamp,
        MessageId:    event.EventID,
        Type:         string(event.EventType),
        Headers: amqp.Table{
            "x-event-type":    string(event.EventType),
            "x-tenant-id":     event.TenantID.String(),
            "x-correlation-id": event.Metadata.CorrelationID,
            "x-source":        event.Metadata.Source,
        },
    }

    // Publish with mandatory flag
    err := p.channel.PublishWithContext(
        ctx,
        TenantExchange,  // "tenant.events"
        routingKey,      // "tenant.created"
        true,            // mandatory - must reach queue
        false,           // immediate
        publishing,
    )

    // Wait for confirmation
    select {
    case confirm := <-p.confirms:
        if !confirm.Ack {
            return fmt.Errorf("message not acknowledged by broker")
        }
    case <-time.After(5 * time.Second):
        return fmt.Errorf("timeout waiting for publish confirmation")
    case <-ctx.Done():
        return ctx.Err()
    }

    return nil
}
```

### 5.4 Consumer Implementation

**Location**: `microservices/tenant-admin-service/internal/events/consumer.go`

**Features**:
- ✅ QoS: Prefetch count = 1 (fair dispatch)
- ✅ Manual acknowledgment (no auto-ack)
- ✅ 30-second timeout per message
- ✅ Nack with requeue on failure
- ✅ Idempotent event handlers
- ✅ Separate goroutine per queue
- ✅ Graceful shutdown

**Code**:
```go
func (c *Consumer) consumeQueue(ctx context.Context, queueName string, eventType EventType) error {
    msgs, err := c.channel.Consume(
        queueName,
        "",    // consumer tag (auto-generated)
        false, // auto-ack (we manually ack)
        false, // exclusive
        false, // no-local
        false, // no-wait
        nil,   // args
    )

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case msg, ok := <-msgs:
            if !ok {
                return fmt.Errorf("message channel closed")
            }

            // Process with timeout
            if err := c.processMessage(ctx, msg, eventType); err != nil {
                c.logger.Error("Failed to process message", zap.Error(err))
                msg.Nack(false, true) // Requeue on failure
            } else {
                msg.Ack(false) // Acknowledge success
            }
        }
    }
}
```

### 5.5 Event Handler (Idempotent)

**Location**: `microservices/tenant-admin-service/internal/events/handler.go`

**HandleTenantCreated**:
```go
func (h *RabbitMQTenantEventHandler) HandleTenantCreated(ctx context.Context, event RabbitMQTenantEvent) error {
    // CHECK IF TENANT ALREADY EXISTS (IDEMPOTENCY)
    var existingTenant models.Tenant
    result := h.db.Where("id = ?", event.Data.ID).First(&existingTenant)
    if result.Error == nil {
        h.logger.Info("Tenant already exists, skipping creation (idempotent)")
        return nil  // Success - safe to replay event
    }

    // Convert event data to tenant model
    tenant := h.eventDataToTenant(event.Data)

    // Create tenant in database
    if err := h.db.Create(&tenant).Error; err != nil {
        return fmt.Errorf("failed to create tenant: %w", err)
    }

    // ⚠️ NOTE: Does NOT create admin user!
    // Event doesn't contain admin credentials

    return nil
}
```

**Critical Issue**: Event handler creates tenant but NOT admin user!

---

## 6. Critical Business Logic Flows

### 6.1 Login Flow (Tenant Admin)

```
1. User navigates to: acme.beakon.com:3002
   - Subdomain "acme" extracted from URL

2. Tenant Admin Frontend loads
   - Checks if logged in (JWT in localStorage)
   - If not logged in, shows login page

3. User enters credentials
   - email: admin@acme.com
   - password: *******

4. Frontend calls tenant-admin-service
   POST /api/v1/auth/login
   {
     "email": "admin@acme.com",
     "password": "*******",
     "subdomain": "acme"  // OR extracted from Host header
   }

5. tenant-admin-service validates
   File: internal/handlers/tenant_admin_handler.go

   A. Extract subdomain
      - From request body OR from Host header

   B. Get tenant by subdomain
      - Query: SELECT * FROM tenants WHERE subdomain = 'acme'

   C. Verify user credentials
      - Query: SELECT * FROM users WHERE email = 'admin@acme.com' AND tenant_id = {tenant_id}
      - Verify: bcrypt.CompareHashAndPassword(user.PasswordHash, password)

   D. Generate JWT token
      - Claims: {
          user_id: user.ID,
          tenant_id: user.TenantID,
          email: user.Email,
          role: user.Role,
          exp: now + 24 hours  // ⚠️ Should be 15 min
        }
      - Sign with JWT_SECRET

   E. Create session (3-tier storage)
      - Try Redis first
      - Fallback to PostgreSQL
      - Emergency fallback to in-memory
      - Session: {
          session_id: UUID,
          user_id: user.ID,
          tenant_id: user.TenantID,
          expires_at: now + 24 hours,
          ip_address: client IP,
          user_agent: client user agent
        }

   F. Set HTTP-only cookie
      - Name: "token"
      - Value: JWT
      - HttpOnly: true
      - Secure: true (if HTTPS)
      - SameSite: Strict
      - Domain: .beakon.com

6. Return success
   {
     "token": "eyJhbGciOiJIUzI1NiIs...",
     "user": {
       "id": "uuid",
       "email": "admin@acme.com",
       "role": "owner",
       "tenant_id": "uuid"
     }
   }

7. Frontend stores token
   - localStorage.setItem("token", jwt)
   - Future requests include: Authorization: Bearer {jwt}

8. Frontend redirects to dashboard
   - URL: /admin/dashboard
```

### 6.2 Authenticated Request Flow

```
1. Frontend makes request
   GET /api/v1/components
   Headers:
     Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
     Host: acme.beakon.com:3002

2. Request hits middleware stack (in order)
   File: main.go:222-242

   A. Logger Middleware
      - Log request: method, path, ip

   B. Recovery Middleware
      - Catch panics, return 500

   C. CORS Middleware
      - Set CORS headers for frontend

   D. Correlation ID Middleware
      - Generate X-Correlation-ID

   E. Rate Limit Middleware (if enabled)
      - Check requests per minute

   F. Tenant Context Middleware
      - Extract subdomain from Host header
      - Query tenant by subdomain
      - Set c.Set("tenant_id", tenant.ID)

   G. JWT Authentication Middleware
      - Extract Bearer token from Authorization header
      - Verify JWT signature
      - Validate expiration
      - Extract claims: user_id, tenant_id, email, role
      - Set c.Set("user_id", claims.user_id)
      - Set c.Set("tenant_id", claims.tenant_id)
      - Set c.Set("email", claims.email)
      - Set c.Set("role", claims.role)

   H. Tenant Middleware (double-check)
      - Ensure tenant_id from JWT matches tenant_id from subdomain
      - If mismatch: Return 403 Forbidden

   I. RBAC Middleware (on specific routes)
      - Example: RequirePermission("components.read")
      - Check if user has permission
      - Query user_roles → role_permissions
      - If no permission: Return 403 Insufficient permissions

3. Request reaches handler
   func (h *ComponentHandler) GetComponents(c *gin.Context) {
       // Extract tenant_id from context (set by middleware)
       tenantID := c.GetString("tenant_id")

       // Query components for this tenant only
       var components []models.Component
       h.db.Where("tenant_id = ?", tenantID).Find(&components)

       c.JSON(200, gin.H{"components": components})
   }

4. Response with tenant-scoped data
   {
     "components": [
       {
         "id": "uuid",
         "tenant_id": "uuid",  // Always scoped to tenant
         "name": "API Server",
         "status": "operational"
       }
     ]
   }
```

### 6.3 Component Status Update Flow

```
1. Monitoring service checks component health
   File: microservices/monitoring-service/internal/monitors/http_monitor.go

   - Runs health check every 60 seconds
   - Makes HTTP GET request to component URL
   - Checks response code and latency

2. If component is DOWN
   A. Update component status
      - Make HTTP PUT to component-service
      - PUT /api/v1/components/{id}/status
      - Body: { "status": "major_outage" }

   B. Create incident (if not exists)
      - Make HTTP POST to incident-service
      - POST /api/v1/incidents
      - Body: {
          "title": "API Server is down",
          "status": "investigating",
          "impact": "critical",
          "component_ids": ["uuid"]
        }

   C. Trigger notifications
      - incident-service calls notification-service
      - POST /api/v1/notifications
      - Body: {
          "incident_id": "uuid",
          "channels": ["email", "slack", "webhook"]
        }

3. notification-service processes
   A. Save notification to database
      - Table: notifications
      - Status: "pending"

   B. Publish to RabbitMQ
      - Queue: notification.queue
      - Event: notification.created

   C. notification-consumer picks up event
      - Sends email via SMTP
      - Posts to Slack via webhook
      - Calls custom webhooks
      - Updates notification status: "sent"

4. Frontend receives update
   - WebSocket connection (if implemented)
   - OR polling /api/v1/incidents every 30s
   - Updates dashboard in real-time
```

---

## 7. Database Architecture & Multi-Tenancy

### 7.1 Database-per-Service Pattern

**Compliance**: ✅ 100% (with one documented exception)

| Database | Service Owner | Tables | Multi-Tenant? |
|----------|---------------|--------|---------------|
| saas_admin | saas-admin-service | 15 | ❌ NO (platform-level data) |
| tenant_admin_db | tenant-admin-service | 25 | ✅ YES (all tables have tenant_id) |
| statuspage_user | user-service | 5 | ✅ YES |
| statuspage_component | component-service | 6 | ✅ YES |
| statuspage_incident | incident-service | 6 | ✅ YES |
| statuspage_notification | notification-service | 7 | ✅ YES |
| statuspage_monitoring | monitoring-service | 6 | ✅ YES |
| statuspage_analytics | analytics-service | 7 | ✅ YES |
| statuspage_payment | payment-service | 8 | ✅ YES |
| statuspage_event_store | event-store-service | 3 | ✅ YES |
| statuspage_branding | branding-service | 4 | ✅ YES |
| statuspage_landing | landing-page-service | 5 | ❌ NO (marketing content) |
| statuspage_analytics_consumer | analytics-consumer | 3 | ✅ YES |
| statuspage_audit_consumer | audit-consumer | 3 | ✅ YES |
| statuspage_billing_consumer | billing-consumer | 4 | ✅ YES |

### 7.2 Exception: tenant_admin_db Shared Access

**Services with Access**:
1. tenant-admin-service (primary owner)
2. saas-admin-service (credential sync only)

**Justification** (from DATABASE_ARCHITECTURE.md):
- Both services are part of "administrative control plane"
- Table ownership is clear (no overlap)
- Direct access for critical credential updates (atomic, no HTTP)

**Tables Accessed by saas-admin-service**:
- `users` table - ONLY for updating tenant owner credentials
- Function: `updateTenantCredentialsDirect()` in saas_admin_handler.go:2157

**Migration Path Documented**:
- Effort: 2-3 days
- Risk: Medium
- Would require: New credential-sync-service OR event-based sync

### 7.3 Multi-Tenancy Enforcement

**Method**: Row-level security with `tenant_id` UUID column

**Example Schema**:
```sql
CREATE TABLE components (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,  -- ✅ ALWAYS PRESENT
    name TEXT NOT NULL,
    status VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_components_tenant_id ON components(tenant_id);
```

**Query Pattern**:
```go
// Extract tenant_id from JWT (set by middleware)
tenantID := c.GetString("tenant_id")

// ALL queries must filter by tenant_id
var components []models.Component
db.Where("tenant_id = ?", tenantID).Find(&components)

// Inserts must include tenant_id
component := models.Component{
    TenantID:  tenantID,  // ✅ MANDATORY
    Name:      "API Server",
    Status:    "operational",
}
db.Create(&component)
```

**Middleware Enforcement**:

File: `internal/middleware/tenant_middleware.go`

```go
func TenantContextMiddleware(db *gorm.DB, logger *zap.Logger, baseDomain string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract subdomain from Host header
        host := c.Request.Host
        subdomain := extractSubdomain(host, baseDomain)

        // Get tenant by subdomain
        var tenant models.Tenant
        if err := db.Where("subdomain = ?", subdomain).First(&tenant).Error; err != nil {
            c.JSON(404, gin.H{"error": "Tenant not found"})
            c.Abort()
            return
        }

        // Set tenant context for downstream handlers
        c.Set("tenant_id", tenant.ID)
        c.Set("tenant", tenant)

        c.Next()
    }
}
```

**Critical**: Every request is scoped to a tenant via subdomain → tenant_id mapping.

---

## 8. Session Management

### 8.1 Three-Tier Session System

**Architecture** (tenant-admin-service):
```
1. PRIMARY: Redis (fastest, volatile)
   └─ If fails →
2. FALLBACK: PostgreSQL (persistent, slower)
   └─ If fails →
3. EMERGENCY: In-memory (process-local, non-shared)
```

**Implementation**:

File: `internal/sessions/session_manager.go`

```go
type SessionManager struct {
    primaryStore   SessionStore  // Redis
    fallbackStore  SessionStore  // PostgreSQL
    logger         *zap.Logger
}

func (sm *SessionManager) CreateSession(ctx context.Context, session *Session) error {
    // Try primary (Redis) first
    if err := sm.primaryStore.Create(ctx, session); err != nil {
        sm.logger.Warn("Primary store failed, using fallback", zap.Error(err))

        // Try fallback (PostgreSQL)
        if fallbackErr := sm.fallbackStore.Create(ctx, session); fallbackErr != nil {
            sm.logger.Error("Both stores failed",
                zap.Error(err),
                zap.Error(fallbackErr))
            return fallbackErr
        }
    }
    return nil
}
```

### 8.2 Session Lifecycle

**Creation** (on login):
```
1. User logs in successfully
2. Generate session:
   {
     session_id: UUID,
     user_id: user.ID,
     tenant_id: user.TenantID,
     token: JWT,
     expires_at: now + 24 hours,
     ip_address: client IP,
     user_agent: client user agent,
     created_at: now,
     last_accessed_at: now
   }

3. Store in Redis
   - Key: "tenant-admin:session:{session_id}"
   - TTL: 24 hours
   - Value: JSON(session)

4. Store in PostgreSQL (fallback)
   - Table: sessions
   - Row: session data
```

**Validation** (on every request):
```
1. Extract session_id from JWT OR cookie

2. Look up session:
   A. Try Redis first
      - GET tenant-admin:session:{session_id}

   B. If not in Redis, try PostgreSQL
      - SELECT * FROM sessions WHERE id = {session_id}
      - If found, repopulate Redis

   C. If not found anywhere
      - Return 401 Unauthorized
      - User must login again

3. Check expiration
   - If expired: Delete session, return 401
   - If valid: Update last_accessed_at

4. Return session data to handler
```

**Cleanup** (background job):
```
1. Run every 30 minutes
   - Ticker: time.NewTicker(30 * time.Minute)

2. Delete expired sessions
   A. Redis: Automatic TTL

   B. PostgreSQL: Manual cleanup
      DELETE FROM sessions WHERE expires_at < NOW()

3. Log cleanup results
```

---

## 9. Frontend-Backend Communication

### 9.1 SaaS Admin Frontend → Backend

**Service**: saas-admin-frontend (port 3001) → saas-admin-service (port 8098)

**Pattern**: CORS-enabled cross-origin requests

**Configuration**:
```yaml
# saas-admin-service CORS
CORS_ALLOWED_ORIGINS=http://localhost:3001,https://admin.beakon.com

# Headers set:
Access-Control-Allow-Origin: http://localhost:3001
Access-Control-Allow-Credentials: true
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS, PATCH
Access-Control-Allow-Headers: Origin, Content-Type, Accept, Authorization
```

**Request Flow**:
```
1. Frontend makes request
   fetch('http://localhost:8098/api/v1/tenants', {
     method: 'GET',
     headers: {
       'Authorization': `Bearer ${token}`,
       'Content-Type': 'application/json'
     },
     credentials: 'include'  // Send cookies
   })

2. Browser sends preflight (OPTIONS)
   - Backend responds with CORS headers

3. Backend validates JWT
   - Extract from Authorization header
   - Verify signature
   - Check expiration

4. Return data
   {
     "tenants": [...]
   }
```

### 9.2 Tenant Admin Frontend → Backend

**Service**: tenant-admin-frontend (port 3002) → tenant-admin-service (port 8099)

**Pattern**: Same-origin requests (subdomain routing)

**Configuration**:
```
Frontend: acme.beakon.com:3002
Backend:  acme.beakon.com:8099

OR (development):
Frontend: acme.localhost:3002
Backend:  acme.localhost:8099
```

**Request Flow**:
```
1. User navigates to: acme.localhost:3002

2. Frontend Next.js app loads
   - SSR on port 3002
   - Detects subdomain: "acme"

3. Frontend makes API request
   fetch('/api/v1/components', {
     method: 'GET',
     headers: {
       'Authorization': `Bearer ${token}`
     }
   })

   // Proxied to: http://acme.localhost:8099/api/v1/components

4. Next.js rewrites (next.config.js)
   rewrites: async () => [
     {
       source: '/api/:path*',
       destination: 'http://localhost:8099/api/:path*'
     }
   ]

5. Backend receives request
   - Same origin (no CORS needed)
   - Subdomain "acme" in Host header
   - Middleware extracts tenant_id

6. Return tenant-scoped data
```

**Critical**: Subdomain-based routing is REQUIRED for multi-tenancy in tenant-admin-frontend.

---

## 10. Deployment Flow

### 10.1 Service Dependencies

**Startup Order**:
```
1. Infrastructure
   - PostgreSQL (all databases)
   - Redis (optional)
   - RabbitMQ (optional)

2. Core Services
   - user-service (8081)
   - saas-admin-service (8098)
   - tenant-admin-service (8099)

3. Domain Services
   - component-service (8084)
   - incident-service (8086)
   - notification-service (8085)
   - monitoring-service (8092)
   - branding-service (8097)
   - analytics-service (8090)
   - payment-service (8088)
   - event-store-service (8096)
   - landing-page-service (8100)
   - status-ui-service (8093)

4. Consumer Services
   - analytics-consumer
   - notification-consumer
   - audit-consumer
   - billing-consumer

5. Frontend Services
   - saas-admin-frontend (3001)
   - tenant-admin-frontend (3002)
```

### 10.2 Database Initialization

**Script**: `microservices/init-all-databases.sh`

**Flow**:
```bash
#!/bin/bash
# Initialize all 14 databases

DATABASES=(
  "saas_admin"
  "tenant_admin_db"
  "statuspage_user"
  "statuspage_component"
  "statuspage_notification"
  "statuspage_incident"
  "statuspage_payment"
  "statuspage_analytics"
  "statuspage_monitoring"
  "statuspage_event_store"
  "statuspage_branding"
  "statuspage_landing"
  "statuspage_analytics_consumer"
  "statuspage_audit_consumer"
  "statuspage_billing_consumer"
)

for db in "${DATABASES[@]}"; do
  echo "Creating database: $db"
  psql -U postgres -c "CREATE DATABASE $db;" 2>/dev/null || echo "$db already exists"
done

echo "Running migrations..."
cd saas-admin-service && atlas migrate apply --env dev
cd ../tenant-admin-service && atlas migrate apply --env dev
# ... for each service with migrations
```

### 10.3 Docker Deployment

**Structure**:
```
docker-deployment/
├── postgres/
│   └── docker-compose.yml
├── redis/
│   └── docker-compose.yml
├── rabbitmq/
│   └── docker-compose.yml
├── saas-admin-service/
│   ├── docker-compose.yml
│   └── configs/
├── tenant-admin-service/
│   ├── docker-compose.yml
│   └── configs/
... (for each service)
```

**Example docker-compose.yml** (saas-admin-service):
```yaml
version: '3.8'

services:
  saas-admin-service:
    build: .
    ports:
      - "8098:8098"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=postgres
      - DB_NAME=saas_admin
      - JWT_SECRET=${JWT_SECRET}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - RABBITMQ_HOST=rabbitmq
      - RABBITMQ_PORT=5672
      - TENANT_ADMIN_SERVICE_URL=http://tenant-admin-service:8099
    depends_on:
      - postgres
      - redis
      - rabbitmq
    networks:
      - beakon-network
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "-O", "/dev/null", "http://127.0.0.1:8098/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

networks:
  beakon-network:
    external: true
```

---

## Summary: Critical Flow Takeaways

### ✅ What Works Well

1. **Database-per-Service Pattern**: 100% compliance (with one documented exception)
2. **Multi-Tenancy**: Proper tenant_id scoping in all tables
3. **RabbitMQ Event Bus**: Reliable messaging with publisher confirms and idempotent consumers
4. **Session Management**: Three-tier fallback (Redis → PostgreSQL → In-memory)
5. **RBAC System**: Comprehensive permission-based access control
6. **Graceful Degradation**: Services continue even if dependencies fail

### ❌ Critical Issues

1. **RabbitMQ Path Doesn't Create Admin User**: Tenant created but admin cannot login
2. **No Circuit Breakers**: saas-admin → tenant-admin calls lack protection
3. **No Retry Logic**: Single HTTP attempt, no resilience for transient failures
4. **Inconsistent Auth Implementation**: saas-admin has refresh tokens, tenant-admin doesn't
5. **No Sync Verification**: Cannot detect out-of-sync tenants

### 🎯 Key Business Flows

1. **Tenant Provisioning**: SaaS Admin creates tenant → dual-mode sync (RabbitMQ + HTTP fallback) → tenant-admin-service
2. **User Creation**: Tenant Admin creates users (limited by max_users) → RBAC role assignment
3. **Authentication**: Subdomain-based tenant isolation → JWT validation → session management
4. **Component Monitoring**: monitoring-service → component-service → incident-service → notification-service
5. **Event-Driven Sync**: saas-admin publishes events → RabbitMQ → tenant-admin consumes (idempotent)

---

**Document End**

This document provides a complete reference for understanding how the Beakon platform works, from tenant provisioning to user authentication to service communication patterns. Use this as the authoritative source for understanding business logic flows and architectural decisions.
