# Tenant Admin Service - Architecture & Implementation Guide

## Table of Contents
1. [System Overview](#system-overview)
2. [Service Architecture](#service-architecture)
3. [Database Schema](#database-schema)
4. [User Management System](#user-management-system)
5. [Max Users Feature](#max-users-feature)
6. [API Endpoints](#api-endpoints)
7. [Authentication & Authorization](#authentication--authorization)
8. [Service Interactions](#service-interactions)
9. [Best Practices](#best-practices)
10. [Troubleshooting](#troubleshooting)

---

## System Overview

The Tenant Admin Service is a microservice responsible for managing tenant-specific administration within a multi-tenant SaaS platform. It handles tenant lifecycle management, user management within tenants, settings, branding, and feature flags.

### Key Responsibilities
- **Tenant Management**: CRUD operations for tenants
- **User Management**: Managing users within each tenant with max_users enforcement
- **Settings & Branding**: Tenant-specific customization
- **Feature Flags**: Tenant-level feature toggling
- **Usage Tracking**: Monitoring tenant resource usage
- **Billing Integration**: Managing tenant billing and subscription plans

### Service Ports
- **Development**: `localhost:8099`
- **Production**: Configured via `SERVER_PORT` environment variable

---

## Service Architecture

### Architectural Pattern
The service follows a **Clean Architecture** pattern with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                        HTTP Layer                            │
│  (Gin Router, Middleware, Request/Response Handling)        │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                     Handlers Layer                           │
│  - tenant_admin_handler.go (Tenant CRUD)                    │
│  - auth_handler.go (Authentication)                         │
│  - dashboard_handler.go (Dashboard & Stats)                 │
│  - helpers.go (Shared utilities)                            │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                    Services Layer                            │
│  - tenant_admin_service.go (Business Logic)                 │
│  - errors.go (Custom error types)                           │
│                                                              │
│  Key Methods:                                                │
│  • CreateTenant() - Creates tenant with admin user          │
│  • CanCreateUser() - Validates max_users limit              │
│  • CreateAdminUser() - Creates user with enforcement        │
│  • GetTenantUserStats() - Returns user count statistics    │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                    Models/Database Layer                     │
│  - tenant_admin.go (GORM models)                            │
│  - Database: PostgreSQL (tenant_admin_db)                   │
└─────────────────────────────────────────────────────────────┘
```

### Core Components

#### 1. **Handlers** (`internal/handlers/`)
- **Purpose**: HTTP request/response handling, input validation
- **Pattern**: Dependency injection of services via constructor
- **Responsibilities**:
  - Request parsing and validation
  - Calling service layer methods
  - Response formatting
  - Error handling and HTTP status code mapping

#### 2. **Services** (`internal/services/`)
- **Purpose**: Business logic implementation
- **Pattern**: Single database connection via constructor injection
- **Responsibilities**:
  - Business rule enforcement (e.g., max_users validation)
  - Data transformation
  - Transaction management
  - Error wrapping with context

#### 3. **Models** (`internal/models/`)
- **Purpose**: Data structure definitions and ORM mappings
- **Pattern**: GORM models with hooks for validation
- **Responsibilities**:
  - Database schema representation
  - Model validation via GORM hooks
  - Relationships between entities

#### 4. **Middleware** (`internal/middleware/`)
- **Purpose**: Request preprocessing and authorization
- **Components**:
  - `jwt_middleware.go` - JWT token validation
  - `rbac_middleware.go` - Role-based access control
- **Pattern**: Gin middleware chain

---

## Database Schema

### Core Tables

#### **tenants**
Primary table for storing tenant information.

```sql
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    domain VARCHAR(255) UNIQUE,
    subdomain VARCHAR(255) UNIQUE,
    contact_email VARCHAR(255) NOT NULL,
    billing_email VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active',
    is_active BOOLEAN DEFAULT true,
    max_users INTEGER NULL CHECK (max_users IS NULL OR max_users > 0),
    settings TEXT,
    branding TEXT,
    features TEXT
);

CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at);
CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_status ON tenants(status);
```

**Key Fields**:
- `max_users`: Maximum allowed users (NULL = unlimited, >0 = limited)
- `slug`: URL-safe unique identifier
- `status`: Tenant lifecycle state (active, inactive, suspended)

#### **users**
Stores user authentication and profile information.

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMP NULL
);

CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE UNIQUE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
```

**Security Notes**:
- Passwords are hashed using bcrypt (cost factor 10)
- Soft deletes preserve email uniqueness via partial index

#### **tenant_admins**
Junction table linking users to tenants with roles.

```sql
CREATE TABLE tenant_admins (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    role VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    permissions TEXT,
    last_login_at TIMESTAMP NULL,
    metadata TEXT
);

CREATE INDEX idx_tenant_admins_tenant_id ON tenant_admins(tenant_id);
CREATE INDEX idx_tenant_admins_user_id ON tenant_admins(user_id);
CREATE INDEX idx_tenant_admins_deleted_at ON tenant_admins(deleted_at);
CREATE UNIQUE INDEX idx_tenant_user ON tenant_admins(tenant_id, user_id) WHERE deleted_at IS NULL;
```

**Roles**:
- `owner` - Full administrative access (created during tenant provisioning)
- `admin` - Administrative access with some restrictions
- `manager` - Limited administrative access
- `viewer` - Read-only access

#### **Additional Tables**

**tenant_branding**: Customization (logos, colors, typography)
**tenant_settings**: Tenant-specific configuration
**tenant_feature_flags**: Feature toggles per tenant
**tenant_usage**: Usage metrics tracking
**tenant_billing**: Billing and subscription data
**sessions**: User session management

---

## User Management System

### User Creation Flow

```
┌───────────┐      ┌────────────────┐      ┌──────────────────┐
│  Handler  │─────▶│ CanCreateUser()│─────▶│ Count Users in   │
│           │      │ (Service)      │      │ tenant_admins    │
└───────────┘      └────────┬───────┘      └──────────────────┘
                            │
                ┌───────────▼───────────┐
                │ Check max_users limit │
                │ NULL = unlimited      │
                │ Number = enforce limit│
                └───────────┬───────────┘
                            │
                ┌───────────▼───────────┐
                │ If limit OK:          │
                │ 1. Hash password      │
                │ 2. Create user        │
                │ 3. Link to tenant     │
                └───────────────────────┘
```

### Key Service Methods

#### `CanCreateUser(ctx context.Context, tenantID uuid.UUID) error`
**Location**: `internal/services/tenant_admin_service.go:744`

**Purpose**: Validates whether a new user can be created for a tenant

**Algorithm**:
1. Fetch tenant from database
2. If `max_users` is NULL → return nil (unlimited)
3. Count active users via JOIN query:
   ```sql
   SELECT COUNT(*) FROM users u
   INNER JOIN tenant_admins ta ON u.id = ta.user_id
   WHERE ta.tenant_id = ?
   AND u.deleted_at IS NULL
   AND ta.deleted_at IS NULL
   ```
4. If `current_users >= max_users` → return error
5. Otherwise → return nil (allow creation)

**Error Response**:
```json
{
  "error": "tenant has reached maximum user limit of 10 users"
}
```

#### `CreateAdminUser(ctx context.Context, tenantID uuid.UUID, email, password string) error`
**Location**: `internal/services/tenant_admin_service.go:812`

**Flow**:
```go
1. Call CanCreateUser(ctx, tenantID) // Validate limit
2. Hash password with bcrypt
3. Begin transaction
4. Create user record
5. Create tenant_admin link (role: "owner")
6. Commit transaction
7. Log success
```

**Features**:
- Automatic max_users validation
- Atomic user + tenant_admin creation
- Proper error wrapping
- Comprehensive logging

#### `GetTenantUserStats(ctx context.Context, tenantID uuid.UUID) (*TenantUserStats, error)`
**Location**: `internal/services/tenant_admin_service.go:724`

**Returns**:
```go
type TenantUserStats struct {
    CurrentUsers   int64  `json:"current_users"`
    MaxUsers       *int   `json:"max_users"`
    IsUnlimited    bool   `json:"is_unlimited"`
    RemainingSlots *int64 `json:"remaining_slots"`
}
```

**Use Cases**:
- Dashboard user count widgets
- Capacity planning alerts
- Billing calculations
- Admin UI user slot indicators

---

## Max Users Feature

### Overview
The max_users feature provides configurable user limits per tenant, enabling tiered pricing models and resource management.

### Database Design

**Field**: `tenants.max_users` (INTEGER NULL)

**Constraints**:
```sql
CHECK (max_users IS NULL OR max_users > 0)
```

**Semantics**:
- `NULL` → Unlimited users
- `> 0` → Specific user limit

### Validation Layers

The feature implements **defense-in-depth** with three validation layers:

#### 1. **Request Layer** (Handlers)
```go
if req.MaxUsers != nil && *req.MaxUsers <= 0 {
    return c.JSON(400, gin.H{"error": "max_users must be greater than 0"})
}
```

#### 2. **Model Layer** (GORM Hooks)
```go
func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
    if t.MaxUsers != nil && *t.MaxUsers <= 0 {
        return fmt.Errorf("max_users must be greater than 0")
    }
    return nil
}
```

#### 3. **Database Layer** (Constraints)
```sql
CHECK (max_users IS NULL OR max_users > 0)
```

### Enforcement Flow

```
User Creation Request
        │
        ▼
┌───────────────────┐
│ CanCreateUser()   │
│ Service Method    │
└────────┬──────────┘
         │
         ▼
┌────────────────────┐      NO
│ max_users NULL?    │─────────┐
└────────┬───────────┘         │
         │ YES                 │
         │                     │
    Allow Creation             │
         │                     ▼
         │            ┌─────────────────────┐
         │            │ Count Active Users  │
         │            │ for Tenant          │
         │            └──────────┬──────────┘
         │                       │
         │                       ▼
         │            ┌─────────────────────┐
         │            │ current >= max?     │
         │            └──────────┬──────────┘
         │                       │
         │           YES         │         NO
         │          ┌────────────┴──────────┐
         │          │                       │
         │          ▼                       ▼
         │    ┌─────────┐          ┌──────────────┐
         │    │ Reject  │          │ Allow        │
         │    │ (HTTP   │          │ Creation     │
         │    │  400)   │          └──────────────┘
         │    └─────────┘
         │
         ▼
    Create User
```

### API Integration

#### Tenant Creation with max_users
**Endpoint**: `POST /api/v1/public/tenants`

**Request**:
```json
{
  "name": "Acme Corp",
  "slug": "acme-corp",
  "contact_email": "contact@acme.com",
  "admin_email": "admin@acme.com",
  "admin_password": "SecurePassword123!",
  "max_users": 10
}
```

**Response**:
```json
{
  "message": "Tenant created successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Acme Corp",
    "slug": "acme-corp",
    "max_users": 10,
    "created_at": "2025-10-02T10:00:00Z"
  }
}
```

#### Get User Statistics
**Endpoint**: `GET /api/v1/tenants/:tenant_id/user-stats`

**Response**:
```json
{
  "data": {
    "current_users": 7,
    "max_users": 10,
    "is_unlimited": false,
    "remaining_slots": 3
  }
}
```

### Logging & Monitoring

The feature includes comprehensive logging at multiple levels:

**INFO**: Normal operations
```
Creating tenant with user limit {"tenant": "acme-corp", "max_users": 10}
Admin user created successfully {"tenant_id": "...", "email": "admin@acme.com"}
```

**WARN**: Limit enforcement
```
Tenant user limit exceeded {"tenant_id": "...", "current": 10, "max": 10}
Cannot create user due to limit restriction {"tenant_id": "...", "email": "user@acme.com"}
```

**DEBUG**: Detailed validation
```
User limit validation {"tenant_id": "...", "current_users": 7, "max_users": 10, "remaining": 3}
```

**ERROR**: Failures
```
Failed to count existing users {"tenant_id": "...", "error": "database connection lost"}
```

---

## API Endpoints

### Public Endpoints (No Authentication)

#### `POST /api/v1/public/tenants`
Create a new tenant with admin user.

**Authentication**: None
**Rate Limit**: 10 requests/hour per IP

**Request Body**:
```json
{
  "name": "string (required)",
  "slug": "string (required, unique, URL-safe)",
  "contact_email": "email (required)",
  "admin_email": "email (required)",
  "admin_password": "string (required, min 8 chars)",
  "max_users": "integer (optional, >0 or null)"
}
```

**Responses**:
- `201 Created` - Tenant created successfully
- `400 Bad Request` - Validation error
- `409 Conflict` - Slug or email already exists
- `500 Internal Server Error` - Server error

---

### Protected Endpoints (Authentication Required)

#### `GET /api/v1/tenants`
List all tenants (admin only).

**Headers**: `Authorization: Bearer <jwt_token>`

**Query Parameters**:
- `limit` (integer, default: 20, max: 100)
- `offset` (integer, default: 0)
- `status` (string: active|inactive|suspended)

**Response**:
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "string",
      "slug": "string",
      "max_users": integer|null,
      "status": "string",
      "created_at": "timestamp"
    }
  ],
  "total": integer,
  "limit": integer,
  "offset": integer
}
```

#### `GET /api/v1/tenants/:id`
Get tenant details by ID.

**Headers**: `Authorization: Bearer <jwt_token>`

**Response**:
```json
{
  "data": {
    "id": "uuid",
    "name": "string",
    "slug": "string",
    "domain": "string",
    "subdomain": "string",
    "contact_email": "email",
    "status": "string",
    "max_users": integer|null,
    "is_active": boolean,
    "created_at": "timestamp",
    "updated_at": "timestamp"
  }
}
```

#### `GET /api/v1/tenants/:tenant_id/user-stats`
Get user count and limit statistics for a tenant.

**Headers**: `Authorization: Bearer <jwt_token>`

**Response**: See [User Statistics](#get-user-statistics) above

---

## Authentication & Authorization

### JWT Authentication

**Token Structure**:
```json
{
  "user_id": 123,
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "admin@acme.com",
  "role": "owner",
  "exp": 1735689600,
  "iat": 1735603200
}
```

**Token Lifecycle**:
1. User logs in → JWT generated with 24h expiration
2. Token stored in HttpOnly cookie (`auth_token`)
3. Middleware validates token on each request
4. Token refresh available via `/api/v1/auth/refresh`

**Cookie Security**:
```go
http.SetCookie(w, &http.Cookie{
    Name:     "auth_token",
    Value:    token,
    Path:     "/",
    HttpOnly: true,           // Prevent XSS
    Secure:   isProduction,   // HTTPS only in prod
    SameSite: http.SameSiteStrictMode, // CSRF protection
    MaxAge:   86400,          // 24 hours
})
```

### Role-Based Access Control (RBAC)

**Hierarchy**:
```
owner > admin > manager > viewer
```

**Permissions Matrix**:

| Action | Owner | Admin | Manager | Viewer |
|--------|-------|-------|---------|--------|
| Create User | ✅ | ✅ | ❌ | ❌ |
| Delete User | ✅ | ✅ | ❌ | ❌ |
| Update Settings | ✅ | ✅ | ✅ | ❌ |
| View Data | ✅ | ✅ | ✅ | ✅ |
| Manage Billing | ✅ | ❌ | ❌ | ❌ |
| Delete Tenant | ✅ | ❌ | ❌ | ❌ |

---

## Service Interactions

### Microservice Communication

```
┌──────────────┐      HTTP/JSON      ┌───────────────────┐
│  SaaS Admin  │────────────────────▶│ Tenant Admin      │
│  Service     │                     │ Service           │
│  (Port 8098) │                     │ (Port 8099)       │
└──────────────┘                     └─────────┬─────────┘
                                               │
                                               │ Manages
                                               ▼
                                     ┌───────────────────┐
                                     │ PostgreSQL DB     │
                                     │ tenant_admin_db   │
                                     └───────────────────┘
```

**SaaS Admin → Tenant Admin**:
- `POST /api/v1/public/tenants` - Create tenant (bypasses auth)
- `POST /api/v1/sessions` - Create user session for SSO
- `GET /api/v1/tenants` - List tenants for admin dashboard

**Resilience Patterns**:
- Circuit breaker on all HTTP calls (3 failures = open circuit)
- Retry logic with exponential backoff (3 attempts max)
- Timeouts: 30s for mutations, 10s for reads
- Graceful degradation (fallback responses)

---

## Best Practices

### 1. **Database Operations**
```go
// ✅ GOOD: Use context for cancellation
err := s.db.WithContext(ctx).Create(&user).Error

// ❌ BAD: No context
err := s.db.Create(&user).Error
```

### 2. **Error Handling**
```go
// ✅ GOOD: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// ❌ BAD: Lose error context
if err != nil {
    return err
}
```

### 3. **Logging**
```go
// ✅ GOOD: Structured logging
s.logger.Info("User created",
    zap.String("tenant_id", tenantID.String()),
    zap.String("email", email))

// ❌ BAD: String concatenation
log.Println("User created: " + email)
```

### 4. **Validation**
```go
// ✅ GOOD: Multiple validation layers
// 1. Request validation (handler)
// 2. Business rules (service)
// 3. Database constraints

// ❌ BAD: Only database validation
```

### 5. **Transactions**
```go
// ✅ GOOD: Use transactions for multi-step operations
tx := s.db.Begin()
if err := tx.Create(&user).Error; err != nil {
    tx.Rollback()
    return err
}
if err := tx.Create(&tenantAdmin).Error; err != nil {
    tx.Rollback()
    return err
}
tx.Commit()
```

---

## Troubleshooting

### Common Issues

#### 1. "tenant has reached maximum user limit"
**Cause**: Attempting to create user when limit is reached

**Debug Steps**:
1. Check current user count:
   ```sql
   SELECT COUNT(*) FROM users u
   INNER JOIN tenant_admins ta ON u.id = ta.user_id
   WHERE ta.tenant_id = '<tenant_id>'
   AND u.deleted_at IS NULL
   AND ta.deleted_at IS NULL;
   ```
2. Check tenant's max_users:
   ```sql
   SELECT max_users FROM tenants WHERE id = '<tenant_id>';
   ```
3. Verify soft-deleted users aren't miscounted

**Solutions**:
- Increase max_users for the tenant
- Remove inactive users
- Permanently delete soft-deleted users

#### 2. "Session ID required"
**Cause**: Missing or invalid JWT token

**Debug Steps**:
1. Check browser cookies for `auth_token`
2. Verify token hasn't expired (exp claim)
3. Check JWT_SECRET matches between services

**Solutions**:
- Re-login to get fresh token
- Verify JWT_SECRET consistency
- Check cookie domain settings

#### 3. Database connection errors
**Cause**: PostgreSQL connection issues

**Debug Steps**:
1. Verify database is running:
   ```bash
   psql -h localhost -U postgres -d tenant_admin_db
   ```
2. Check connection parameters in config
3. Review logs for connection pool exhaustion

**Solutions**:
- Restart PostgreSQL service
- Verify DB_PASSWORD in environment
- Increase connection pool size

---

## Configuration Reference

### Environment Variables

```bash
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8099

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=tenant_admin_db
DB_SSL_MODE=disable
DB_MAX_CONNS=100

# JWT Configuration
JWT_SECRET=your-secret-key-min-32-chars
JWT_EXPIRATION_HOURS=24
JWT_ISSUER=tenant-admin-service

# Tenant Configuration
TENANT_MAX_USERS_PER_TENANT=10
TENANT_DEFAULT_PLAN=free
TENANT_FEATURE_FLAGS_ENABLED=true

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2025-10-02 | Initial architecture documentation |
| 1.1.0 | 2025-10-02 | Added max_users feature documentation |

---

## Contributing

When modifying this service:
1. Update this documentation
2. Add tests for new features
3. Follow existing code patterns
4. Update API documentation
5. Add migration scripts if schema changes

---

**Last Updated**: 2025-10-02
**Maintained By**: Platform Engineering Team
