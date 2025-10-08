# Max Users Feature - Visual Flow Diagrams

## 1. Tenant Creation Flow (SaaS Admin → Tenant Admin)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            SAAS ADMIN PANEL                                  │
│                              (Port 8098)                                     │
└──────────────────────────┬──────────────────────────────────────────────────┘
                           │
                           │ Admin creates tenant
                           │ Sets max_users = 10
                           │
                           ▼
                    ┌─────────────┐
                    │ HTTP POST   │
                    │ /api/admin/ │
                    │   tenants   │
                    └──────┬──────┘
                           │
                           │ {"name": "Acme", "max_users": 10}
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                      SAAS ADMIN SERVICE (Internal)                           │
│                                                                              │
│   ┌──────────────────────────────────────────────────────────────────┐     │
│   │  1. Validates request                                             │     │
│   │  2. Enriches with metadata (plan, billing)                        │     │
│   │  3. Forwards to Tenant Admin Service                              │     │
│   └──────────────────────────────────────────────────────────────────┘     │
└──────────────────────────┬──────────────────────────────────────────────────┘
                           │
                           │ Service-to-Service call
                           │
                           ▼
                    ┌─────────────┐
                    │ HTTP POST   │
                    │ /api/v1/    │
                    │ public/     │
                    │ tenants     │
                    └──────┬──────┘
                           │
                           │ {"name": "Acme", "max_users": 10}
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                     TENANT ADMIN SERVICE (Port 8099)                         │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │                    HANDLER LAYER                                    │    │
│  │  tenant_admin_handler.go:CreateTenant()                             │    │
│  │  ┌──────────────────────────────────────────────────────────────┐  │    │
│  │  │ 1. Bind JSON to CreateTenantRequest                          │  │    │
│  │  │ 2. Validate: max_users > 0 (Gin binding)                     │  │    │
│  │  │ 3. Validate: explicit check for max_users <= 0               │  │    │
│  │  │ 4. Log tenant creation with limit                            │  │    │
│  │  └──────────────────────────────────────────────────────────────┘  │    │
│  └────────────────────────────┬───────────────────────────────────────┘    │
│                                │                                             │
│                                ▼                                             │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │                    SERVICE LAYER                                    │    │
│  │  tenant_admin_service.go:CreateTenant()                             │    │
│  │  ┌──────────────────────────────────────────────────────────────┐  │    │
│  │  │ 1. Generate unique slug                                      │  │    │
│  │  │ 2. Set default values (status, is_active)                    │  │    │
│  │  │ 3. Create tenant record (triggers BeforeCreate hook)         │  │    │
│  │  │ 4. Create default settings                                   │  │    │
│  │  │ 5. Create default billing (includes max_users)               │  │    │
│  │  │ 6. Create default branding                                   │  │    │
│  │  └──────────────────────────────────────────────────────────────┘  │    │
│  └────────────────────────────┬───────────────────────────────────────┘    │
│                                │                                             │
│                                ▼                                             │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │                    MODEL LAYER                                      │    │
│  │  tenant_admin.go:Tenant.BeforeCreate()                              │    │
│  │  ┌──────────────────────────────────────────────────────────────┐  │    │
│  │  │ 1. Validate() called automatically                           │  │    │
│  │  │ 2. Check: max_users IS NULL OR max_users > 0                │  │    │
│  │  │ 3. Return error if validation fails                          │  │    │
│  │  └──────────────────────────────────────────────────────────────┘  │    │
│  └────────────────────────────┬───────────────────────────────────────┘    │
│                                │                                             │
└────────────────────────────────┼─────────────────────────────────────────────┘
                                 │
                                 ▼
                    ┌────────────────────────┐
                    │   POSTGRESQL DATABASE   │
                    │                        │
                    │  INSERT INTO tenants   │
                    │  VALUES (              │
                    │    id = UUID(),        │
                    │    name = 'Acme',      │
                    │    max_users = 10,     │  ← STORED
                    │    ...                 │
                    │  )                     │
                    │                        │
                    │  CHECK CONSTRAINT:     │
                    │  max_users IS NULL     │
                    │  OR max_users > 0      │  ← ENFORCED
                    └────────────────────────┘
                                 │
                                 │ SUCCESS
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│              CREATE ADMIN USER (First User)                                  │
│                                                                              │
│  tenant_admin_service.go:CreateAdminUser()                                   │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │ 1. CanCreateUser(tenantID) - validates limit                         │  │
│  │    ├─ Get tenant from DB                                             │  │
│  │    ├─ Count current users: 0                                         │  │
│  │    ├─ Compare: 0 < 10 ✓                                              │  │
│  │    └─ Return: nil (allowed)                                          │  │
│  │ 2. Hash password                                                     │  │
│  │ 3. Create user record                                                │  │
│  │ 4. Create tenant_admin link (role: owner)                            │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
                                 │
                                 │
                                 ▼
                    ┌────────────────────────┐
                    │      RESPONSE          │
                    │                        │
                    │  201 Created           │
                    │  {                     │
                    │    "data": {           │
                    │      "id": "UUID",     │
                    │      "name": "Acme",   │
                    │      "max_users": 10   │  ← RETURNED
                    │    }                   │
                    │  }                     │
                    └────────────────────────┘
```

---

## 2. User Creation Flow (Within Tenant Admin Portal)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        TENANT ADMIN PORTAL                                   │
│                    (acme-corp.localhost:8099)                                │
└──────────────────────────┬──────────────────────────────────────────────────┘
                           │
                           │ Admin invites new user
                           │ Current users: 7/10
                           │
                           ▼
                    ┌─────────────┐
                    │ HTTP POST   │
                    │ /api/v1/    │
                    │ admins      │
                    └──────┬──────┘
                           │
                           │ {"email": "user@acme.com", "role": "editor"}
                           │ Header: Authorization: Bearer {JWT}
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                   TENANT ADMIN SERVICE - HANDLER                             │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │ CreateTenantAdmin(c *gin.Context)                                   │    │
│  │ ┌────────────────────────────────────────────────────────────────┐ │    │
│  │ │ 1. Extract tenant_id from JWT/context                          │ │    │
│  │ │ 2. Bind JSON to TenantAdmin model                              │ │    │
│  │ │ 3. Set admin.TenantID = extracted tenant_id                    │ │    │
│  │ │ 4. Call service.CreateTenantAdmin()                            │ │    │
│  │ └────────────────────────────────────────────────────────────────┘ │    │
│  └────────────────────────────────────────────────────────────────────┘    │
└──────────────────────────┬──────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                   TENANT ADMIN SERVICE - SERVICE LAYER                       │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │ CreateAdminUser(tenantID, email, password)                          │    │
│  │                                                                      │    │
│  │  ┌───────────────────────────────────────────────────────────┐     │    │
│  │  │  STEP 1: Validate User Limit                              │     │    │
│  │  │  CanCreateUser(tenantID)                                  │     │    │
│  │  └─────────────────┬─────────────────────────────────────────┘     │    │
│  │                    │                                                │    │
│  │                    ▼                                                │    │
│  │  ┌───────────────────────────────────────────────────────────┐     │    │
│  │  │  Get tenant from DB                                       │     │    │
│  │  │  ├─ SELECT * FROM tenants WHERE id = {tenantID}          │     │    │
│  │  │  └─ Result: max_users = 10                               │     │    │
│  │  └─────────────────┬─────────────────────────────────────────┘     │    │
│  │                    │                                                │    │
│  │                    ▼                                                │    │
│  │  ┌───────────────────────────────────────────────────────────┐     │    │
│  │  │  Count current users                                      │     │    │
│  │  │  ├─ SELECT COUNT(*) FROM users u                         │     │    │
│  │  │  │  INNER JOIN tenant_admins ta                          │     │    │
│  │  │  │  ON u.id = ta.user_id                                 │     │    │
│  │  │  │  WHERE ta.tenant_id = {tenantID}                      │     │    │
│  │  │  │  AND u.deleted_at IS NULL                             │     │    │
│  │  │  │  AND ta.deleted_at IS NULL                            │     │    │
│  │  │  └─ Result: 7 users                                      │     │    │
│  │  └─────────────────┬─────────────────────────────────────────┘     │    │
│  │                    │                                                │    │
│  │                    ▼                                                │    │
│  │  ┌───────────────────────────────────────────────────────────┐     │    │
│  │  │  Compare: currentUsers < maxUsers ?                       │     │    │
│  │  │  ├─ 7 < 10 ✓                                             │     │    │
│  │  │  ├─ Log: "User creation allowed, 3 slots remaining"      │     │    │
│  │  │  └─ Return: nil (success)                                │     │    │
│  │  └─────────────────┬─────────────────────────────────────────┘     │    │
│  │                    │                                                │    │
│  │                    ▼                                                │    │
│  │  ┌───────────────────────────────────────────────────────────┐     │    │
│  │  │  STEP 2: Create User (only if validation passed)         │     │    │
│  │  │  ├─ Hash password                                         │     │    │
│  │  │  ├─ INSERT INTO users (email, password_hash)             │     │    │
│  │  │  ├─ INSERT INTO tenant_admins (tenant_id, user_id)       │     │    │
│  │  │  └─ Log: "User created successfully"                     │     │    │
│  │  └───────────────────────────────────────────────────────────┘     │    │
│  └────────────────────────────────────────────────────────────────────┘    │
└──────────────────────────┬──────────────────────────────────────────────────┘
                           │
                           ▼
                    ┌────────────────────────┐
                    │      RESPONSE          │
                    │                        │
                    │  201 Created           │
                    │  {                     │
                    │    "message": "User    │
                    │     created",          │
                    │    "data": {           │
                    │      "id": 123,        │
                    │      "email": "..."    │
                    │    }                   │
                    │  }                     │
                    └────────────────────────┘
```

---

## 3. User Creation Flow - LIMIT EXCEEDED Scenario

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        TENANT ADMIN PORTAL                                   │
│                    Current users: 10/10 (AT LIMIT)                           │
└──────────────────────────┬──────────────────────────────────────────────────┘
                           │
                           │ Admin tries to invite 11th user
                           │
                           ▼
                    ┌─────────────┐
                    │ HTTP POST   │
                    │ /api/v1/    │
                    │ admins      │
                    └──────┬──────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                   HANDLER → SERVICE LAYER                                    │
│                                                                              │
│  CreateAdminUser(tenantID, email, password)                                  │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  VALIDATION: CanCreateUser(tenantID)                               │    │
│  │                                                                      │    │
│  │  ┌──────────────────────────────────────────────────────────────┐  │    │
│  │  │ 1. Get tenant: max_users = 10                                │  │    │
│  │  │ 2. Count users: 10                                           │  │    │
│  │  │ 3. Compare: 10 >= 10 ✗ (LIMIT REACHED)                      │  │    │
│  │  │ 4. Log WARN: "Tenant user limit exceeded"                   │  │    │
│  │  │ 5. Return ERROR: "tenant has reached maximum user limit"    │  │    │
│  │  └──────────────────────────────────────────────────────────────┘  │    │
│  │                                                                      │    │
│  │  ❌ STOP - User creation blocked                                    │    │
│  │                                                                      │    │
│  └────────────────────────────────────────────────────────────────────┘    │
└──────────────────────────┬──────────────────────────────────────────────────┘
                           │
                           │ Error propagates back
                           │
                           ▼
                    ┌────────────────────────┐
                    │   ERROR RESPONSE       │
                    │                        │
                    │  400 Bad Request       │
                    │  {                     │
                    │    "error": "tenant    │
                    │     has reached        │
                    │     maximum user       │
                    │     limit of 10        │
                    │     users"             │
                    │  }                     │
                    └────────────────────────┘
                           │
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                      TENANT ADMIN PORTAL                                     │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  Error displayed to user:                                           │    │
│  │                                                                      │    │
│  │  ⚠️  Cannot create user                                             │    │
│  │  Your tenant has reached the maximum user limit of 10 users.       │    │
│  │  Please upgrade your plan to add more users.                        │    │
│  │                                                                      │    │
│  │  [Upgrade Plan]  [Contact Support]                                  │    │
│  └────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 4. User Statistics Query Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        TENANT ADMIN DASHBOARD                                │
│                     Displays: "7/10 users"                                   │
└──────────────────────────┬──────────────────────────────────────────────────┘
                           │
                           │ Periodic refresh or on-demand query
                           │
                           ▼
                    ┌─────────────┐
                    │ HTTP GET    │
                    │ /api/v1/    │
                    │ tenants/    │
                    │ {id}/       │
                    │ user-stats  │
                    └──────┬──────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                    TENANT ADMIN SERVICE                                      │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │ GetTenantUserStats(tenantID)                                        │    │
│  │                                                                      │    │
│  │  ┌──────────────────────────────────────────────────────────────┐  │    │
│  │  │ 1. SELECT * FROM tenants WHERE id = {tenantID}               │  │    │
│  │  │    Result: max_users = 10                                    │  │    │
│  │  └──────────────────────────────────────────────────────────────┘  │    │
│  │                    │                                                │    │
│  │                    ▼                                                │    │
│  │  ┌──────────────────────────────────────────────────────────────┐  │    │
│  │  │ 2. SELECT COUNT(*) FROM users u                              │  │    │
│  │  │    INNER JOIN tenant_admins ta ON u.id = ta.user_id          │  │    │
│  │  │    WHERE ta.tenant_id = {tenantID}                           │  │    │
│  │  │    AND u.deleted_at IS NULL                                  │  │    │
│  │  │    Result: current_users = 7                                 │  │    │
│  │  └──────────────────────────────────────────────────────────────┘  │    │
│  │                    │                                                │    │
│  │                    ▼                                                │    │
│  │  ┌──────────────────────────────────────────────────────────────┐  │    │
│  │  │ 3. Calculate stats:                                          │  │    │
│  │  │    - current_users: 7                                        │  │    │
│  │  │    - max_users: 10                                           │  │    │
│  │  │    - is_unlimited: false                                     │  │    │
│  │  │    - remaining_slots: 10 - 7 = 3                            │  │    │
│  │  └──────────────────────────────────────────────────────────────┘  │    │
│  └────────────────────────────────────────────────────────────────────┘    │
└──────────────────────────┬──────────────────────────────────────────────────┘
                           │
                           ▼
                    ┌────────────────────────┐
                    │      RESPONSE          │
                    │                        │
                    │  200 OK                │
                    │  {                     │
                    │    "data": {           │
                    │      "current_users": 7│
                    │      "max_users": 10,  │
                    │      "is_unlimited":   │
                    │        false,          │
                    │      "remaining_slots":│
                    │        3               │
                    │    }                   │
                    │  }                     │
                    └────────────┬───────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                      TENANT ADMIN DASHBOARD                                  │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  User Management                                                    │    │
│  │  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    │    │
│  │                                                                      │    │
│  │  Users: 7 / 10                          [70% capacity]              │    │
│  │  ████████████████████░░░░░░░░                                       │    │
│  │                                                                      │    │
│  │  Remaining slots: 3                                                 │    │
│  │                                                                      │    │
│  │  [Invite User]  [Manage Users]  [Upgrade Plan]                      │    │
│  └────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 5. Validation Layers Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          VALIDATION LAYERS                                   │
│                          (Defense in Depth)                                  │
└─────────────────────────────────────────────────────────────────────────────┘

    Request with max_users = 0 (Invalid)
              │
              ▼
    ┌───────────────────────────────────┐
    │   LAYER 1: Handler Validation     │
    │   (tenant_admin_handler.go)       │
    │                                   │
    │   Gin Binding:                    │
    │   binding:"omitempty,gt=0"        │
    │                                   │
    │   ❌ REJECTED: "max_users must    │
    │      be greater than 0"           │
    └───────────────────────────────────┘
              │ (If bypassed somehow)
              ▼
    ┌───────────────────────────────────┐
    │   LAYER 2: Model Validation       │
    │   (tenant_admin.go)               │
    │                                   │
    │   BeforeCreate() Hook:            │
    │   if max_users <= 0 {             │
    │     return ErrInvalidValue        │
    │   }                               │
    │                                   │
    │   ❌ REJECTED: GORM error         │
    └───────────────────────────────────┘
              │ (If bypassed somehow)
              ▼
    ┌───────────────────────────────────┐
    │   LAYER 3: Database Constraint    │
    │   (PostgreSQL)                    │
    │                                   │
    │   CHECK CONSTRAINT:               │
    │   max_users IS NULL OR            │
    │   max_users > 0                   │
    │                                   │
    │   ❌ REJECTED: DB constraint      │
    │      violation                    │
    └───────────────────────────────────┘
              │ (If all pass)
              ▼
    ┌───────────────────────────────────┐
    │   ✓ ACCEPTED                      │
    │   Tenant created with valid       │
    │   max_users value                 │
    └───────────────────────────────────┘


    User creation attempt when at limit
              │
              ▼
    ┌───────────────────────────────────┐
    │   LAYER 4: Service Validation     │
    │   (tenant_admin_service.go)       │
    │                                   │
    │   CanCreateUser():                │
    │   1. Get tenant max_users         │
    │   2. Count current users          │
    │   3. Compare: current >= max?     │
    │                                   │
    │   ❌ REJECTED: "tenant has        │
    │      reached maximum user limit"  │
    └───────────────────────────────────┘
```

---

## 6. Database Schema Relationships

```
┌────────────────────────────────────────┐
│          TENANTS TABLE                 │
├────────────────────────────────────────┤
│ id (UUID) PK                           │
│ name VARCHAR(255)                      │
│ slug VARCHAR(255) UNIQUE               │
│ max_users INTEGER ◄─────────┐         │  ← Stores the limit
│   CHECK (max_users IS NULL   │         │
│         OR max_users > 0)    │         │
│ status VARCHAR(50)           │         │
│ is_active BOOLEAN            │         │
│ created_at TIMESTAMP         │         │
│ updated_at TIMESTAMP         │         │
│ deleted_at TIMESTAMP         │         │
└────────────────┬───────────────────────┘
                 │
                 │ tenant_id (FK)
                 │
                 ▼
┌────────────────────────────────────────┐
│       TENANT_ADMINS TABLE              │  ← Join table
├────────────────────────────────────────┤
│ id SERIAL PK                           │
│ tenant_id UUID FK ───────┐             │
│ user_id INTEGER FK ──┐   │             │
│ role VARCHAR(50)     │   │             │
│ status VARCHAR(50)   │   │             │
│ created_at TIMESTAMP │   │             │
│ deleted_at TIMESTAMP │   │             │
└──────────────────────┼───┼─────────────┘
                       │   │
                       │   └──────────────┐
                       │                  │
                       ▼                  ▼
      ┌────────────────────────────────────────┐
      │           USERS TABLE                  │
      ├────────────────────────────────────────┤
      │ id SERIAL PK                           │
      │ email VARCHAR(255) UNIQUE              │
      │ password_hash VARCHAR(255)             │
      │ first_name VARCHAR(100)                │
      │ last_name VARCHAR(100)                 │
      │ is_active BOOLEAN                      │
      │ created_at TIMESTAMP                   │
      │ deleted_at TIMESTAMP                   │
      └────────────────────────────────────────┘

COUNT Query for validation:
────────────────────────────
SELECT COUNT(*)
FROM users u
INNER JOIN tenant_admins ta ON u.id = ta.user_id
WHERE ta.tenant_id = {tenantID}
  AND u.deleted_at IS NULL      ◄─ Exclude soft-deleted users
  AND ta.deleted_at IS NULL     ◄─ Exclude soft-deleted links
```

---

## 7. Logging Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          LOGGING TIMELINE                                    │
└─────────────────────────────────────────────────────────────────────────────┘

T+0ms   [INFO]  Creating tenant with user limit
                {"name": "Acme", "slug": "acme-corp", "max_users": 10}
                ├─ Location: tenant_admin_handler.go:1087
                └─ Level: INFO (audit trail)

T+10ms  [DEBUG] Generating slug
                {"name": "Acme Corporation", "slug": "acme-corp"}
                ├─ Location: tenant_admin_service.go:410
                └─ Level: DEBUG (detailed flow)

T+15ms  [DEBUG] Checking slug uniqueness
                {"slug": "acme-corp", "exists": false}
                ├─ Location: tenant_admin_service.go:420
                └─ Level: DEBUG

T+50ms  [INFO]  Tenant created successfully
                {"tenant_id": "550e8400-e29b-41d4-a716-446655440000"}
                ├─ Location: tenant_admin_service.go:511
                └─ Level: INFO (success event)

T+55ms  [DEBUG] User limit validation
                {"tenant_id": "550e8400-...", "current_users": 0, "max_users": 10}
                ├─ Location: tenant_admin_service.go:761
                └─ Level: DEBUG (validation check)

T+60ms  [DEBUG] User creation allowed
                {"tenant_id": "550e8400-...", "remaining_slots": 10}
                ├─ Location: tenant_admin_service.go:778
                └─ Level: DEBUG (validation passed)

T+100ms [INFO]  Admin user created successfully
                {"tenant_id": "550e8400-...", "email": "admin@acme.com", "user_id": 1}
                ├─ Location: tenant_admin_service.go:752
                └─ Level: INFO (user created)

─────────────────────────────────────────────────────────────────────────────

Later: User #8 creation attempt (at 7/10 capacity)

T+0ms   [INFO]  Creating user
                {"tenant_id": "550e8400-...", "email": "user8@acme.com"}

T+5ms   [DEBUG] User limit validation
                {"tenant_id": "550e8400-...", "current_users": 7, "max_users": 10}

T+10ms  [DEBUG] User creation allowed
                {"tenant_id": "550e8400-...", "remaining_slots": 3}

T+50ms  [INFO]  User created successfully
                {"tenant_id": "550e8400-...", "email": "user8@acme.com", "user_id": 123}

─────────────────────────────────────────────────────────────────────────────

Later: User #11 creation attempt (at 10/10 capacity - LIMIT EXCEEDED)

T+0ms   [INFO]  Creating user
                {"tenant_id": "550e8400-...", "email": "user11@acme.com"}

T+5ms   [DEBUG] User limit validation
                {"tenant_id": "550e8400-...", "current_users": 10, "max_users": 10}

T+10ms  [WARN]  Tenant user limit exceeded
                {"tenant_id": "550e8400-...", "current_users": 10, "max_users": 10}
                ├─ Location: tenant_admin_service.go:769
                └─ Level: WARN (limit reached)

T+15ms  [WARN]  Cannot create user due to limit restriction
                {"tenant_id": "550e8400-...", "email": "user11@acme.com",
                 "error": "tenant has reached maximum user limit of 10 users"}
                ├─ Location: tenant_admin_service.go:795
                └─ Level: WARN (operation blocked)

T+20ms  [ERROR] Failed to create user
                {"error": "tenant has reached maximum user limit of 10 users"}
                ├─ Location: tenant_admin_handler.go (depends on implementation)
                └─ Level: ERROR (operation failed)
```

---

## 8. Comparison: Unlimited vs Limited Tenant

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    UNLIMITED TENANT (max_users = NULL)                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Database: max_users = NULL                                                  │
│                                                                              │
│  CanCreateUser() logic:                                                      │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │ 1. Get tenant: max_users = NULL                                    │    │
│  │ 2. Check: if max_users == NULL { return nil }  ✓                  │    │
│  │ 3. ✓ ALLOWED - Skip counting, skip comparison                     │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  Result: Users can be created without limit                                  │
│                                                                              │
│  Dashboard display: "10 users (unlimited)"                                   │
└─────────────────────────────────────────────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────────────────────┐
│                    LIMITED TENANT (max_users = 10)                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Database: max_users = 10                                                    │
│                                                                              │
│  CanCreateUser() logic:                                                      │
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │ 1. Get tenant: max_users = 10                                      │    │
│  │ 2. Check: if max_users == NULL { return nil } ✗ (not null)        │    │
│  │ 3. Count current users: SELECT COUNT(*) ... = 7                    │    │
│  │ 4. Compare: 7 < 10 ✓                                               │    │
│  │ 5. ✓ ALLOWED (3 slots remaining)                                  │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  Result: Users can be created up to limit                                    │
│                                                                              │
│  Dashboard display: "7 / 10 users (70%)"                                     │
│                   ████████████████████░░░░░░░░                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

These diagrams provide visual representations of all flows in the max_users feature implementation. Each diagram focuses on a specific aspect:

1. **Tenant Creation** - Complete flow from SaaS Admin to database
2. **User Creation (Success)** - Normal flow when under limit
3. **User Creation (Blocked)** - Error flow when limit exceeded
4. **Statistics Query** - Real-time monitoring
5. **Validation Layers** - Defense in depth approach
6. **Database Schema** - Relationships and constraints
7. **Logging Timeline** - Audit trail and debugging
8. **Tenant Types** - Unlimited vs limited comparison
