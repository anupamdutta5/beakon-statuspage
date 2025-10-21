# SaaS Admin Service - Architecture & Implementation Guide

## Table of Contents
1. [System Overview](#system-overview)
2. [Service Architecture](#service-architecture)
3. [Database Schema](#database-schema)
4. [Tenant Management](#tenant-management)
5. [Max Users Configuration](#max-users-configuration)
6. [API Endpoints](#api-endpoints)
7. [Authentication System](#authentication-system)
8. [Service Interactions](#service-interactions)
9. [UI/UX Architecture](#uiux-architecture)
10. [Best Practices](#best-practices)

---

## System Overview

The SaaS Admin Service is the **control plane** for the entire multi-tenant platform. It provides the administrative interface for platform operators to manage tenants, billing, plans, and platform-wide settings.

### Key Responsibilities
- **Platform Administration**: Manage the entire SaaS platform
- **Tenant Provisioning**: Create and configure new tenants
- **Billing & Subscriptions**: Manage pricing plans and tenant subscriptions
- **Analytics & Reporting**: Platform-wide metrics and insights
- **System Configuration**: Global settings and feature flags

### Service Ports
- **Development**: `localhost:8098`
- **Production**: Configured via `SERVER_PORT` environment variable

---

## Service Architecture

### Architectural Pattern
The service follows a **Monolithic Admin Panel** pattern optimized for administrative operations:

```
┌─────────────────────────────────────────────────────────────┐
│                      Web Interface                           │
│  (HTML Templates, CSS, JavaScript)                          │
│  - Dashboard (analytics, metrics)                           │
│  - Tenant Management                                        │
│  - Billing & Plans                                          │
│  - System Settings                                          │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                     HTTP Layer                               │
│  (Gin Router, Session Middleware, Authentication)          │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                    Handlers Layer                            │
│  - saas_admin_handler.go (Main handler)                     │
│    • GetDashboard() - Render admin dashboard                │
│    • CreateTenant() - Provision new tenant                  │
│    • GetTenants() - List all tenants                        │
│    • UpdateTenant() - Modify tenant settings                │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                    Services Layer                            │
│  - saas_admin_service.go (Business Logic)                   │
│    • TenantProvisioning - Creates tenant via API call       │
│    • SessionManagement - Delegates to Tenant Admin          │
│    • BillingCalculation - Plan pricing logic                │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│               External Service Clients                       │
│  - Tenant Admin Service (Port 8099)                         │
│  - Analytics Service (Future)                               │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                    Database Layer                            │
│  - PostgreSQL (saas_admin database)                         │
│  - Stores: admin users, audit logs                          │
└─────────────────────────────────────────────────────────────┘
```

### Core Components

#### 1. **Web Templates** (`web/templates/`)
- **Purpose**: Server-rendered HTML for admin interface
- **Technology**: Go templates with Bootstrap 5
- **Structure**:
  - `dashboard.html` - Main dashboard layout
  - `partials/header.html` - Navigation and header
  - `partials/sidebar.html` - Navigation sidebar
  - `partials/dashboard.html` - Dashboard content
  - `partials/scripts.html` - JavaScript logic
  - `partials/styles.html` - Custom CSS styles

#### 2. **Handlers** (`internal/handlers/`)
- **Purpose**: HTTP request routing and response rendering
- **Pattern**: Server-side rendering with AJAX for dynamic updates
- **Responsibilities**:
  - Template rendering
  - API proxy to Tenant Admin Service
  - Session validation
  - Error handling

#### 3. **Services** (`internal/services/`)
- **Purpose**: Business logic and external service orchestration
- **Pattern**: HTTP client-based service communication
- **Responsibilities**:
  - Tenant provisioning workflow
  - Session management delegation
  - Analytics aggregation
  - Billing calculations

#### 4. **Models** (`internal/models/`)
- **Purpose**: Data structures for admin entities
- **Components**:
  - `SaaSAdminUser` - Platform administrator
  - `AuditLog` - Activity tracking
  - `PlatformSettings` - Global configuration

---

## Database Schema

### Core Tables

#### **saas_admin_users**
Platform administrators who can manage the SaaS platform.

```sql
CREATE TABLE saas_admin_users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'admin',
    status VARCHAR(50) DEFAULT 'active',
    last_login_at TIMESTAMP NULL,
    permissions TEXT
);

CREATE INDEX idx_saas_admin_users_email ON saas_admin_users(email);
CREATE INDEX idx_saas_admin_users_username ON saas_admin_users(username);
CREATE INDEX idx_saas_admin_users_deleted_at ON saas_admin_users(deleted_at);
```

**Roles**:
- `super_admin` - Full platform access, can manage other admins
- `admin` - Standard admin access, tenant management only
- `billing_admin` - Billing and subscription management only
- `support_admin` - Read-only access for support operations

#### **audit_logs**
Comprehensive activity tracking for compliance and security.

```sql
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    user_id INTEGER REFERENCES saas_admin_users(id),
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(255),
    resource_id VARCHAR(255),
    details TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
```

**Tracked Actions**:
- `tenant.create` - New tenant provisioned
- `tenant.update` - Tenant settings modified
- `tenant.delete` - Tenant deleted/suspended
- `user.login` - Admin login
- `user.logout` - Admin logout
- `settings.update` - Platform settings changed

---

## Tenant Management

### Tenant Creation Flow

```
┌─────────────────┐
│ SaaS Admin UI   │
│ Create Tenant   │
│ Form            │
└────────┬────────┘
         │
         │ User fills:
         │ - Organization Name
         │ - Domain/Slug
         │ - Contact Email
         │ - Admin Email/Password
         │ - Subscription Plan
         │ - Max Users ← NEW FIELD
         │
         ▼
┌─────────────────────────┐
│ JavaScript Validation   │
│ - Required fields       │
│ - Email format          │
│ - Password strength     │
│ - Max users > 0         │
└────────┬────────────────┘
         │
         │ Valid
         ▼
┌──────────────────────────────┐
│ AJAX POST Request            │
│ /api/v1/tenants              │
│ (SaaS Admin Service)         │
└────────┬─────────────────────┘
         │
         │ Proxies to
         ▼
┌────────────────────────────────────┐
│ Tenant Admin Service               │
│ POST /api/v1/public/tenants        │
│                                    │
│ Creates:                           │
│ 1. Tenant record (with max_users) │
│ 2. Admin user                      │
│ 3. Database schema                 │
│ 4. Initial settings                │
└────────┬───────────────────────────┘
         │
         │ Success
         ▼
┌────────────────────────┐
│ SaaS Admin Dashboard   │
│ - Show success message │
│ - Refresh tenant list  │
│ - Display tenant card  │
└────────────────────────┘
```

### Tenant Data Flow

**From UI to Backend**:
```javascript
const tenantData = {
    name: "Acme Corp",
    domain: "acme-corp",
    subdomain: "acme-corp",
    contact_email: "contact@acme.com",
    admin_email: "admin@acme.com",
    admin_password: "SecurePass123!",
    plan: "pro",
    max_users: 25  // NEW: User limit
};

fetch('/api/v1/tenants', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(tenantData)
});
```

**SaaS Admin → Tenant Admin**:
```http
POST http://localhost:8099/api/v1/public/tenants
Content-Type: application/json

{
  "name": "Acme Corp",
  "slug": "acme-corp",
  "contact_email": "contact@acme.com",
  "admin_email": "admin@acme.com",
  "admin_password": "SecurePass123!",
  "max_users": 25
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
    "max_users": 25,
    "status": "active",
    "created_at": "2025-10-02T10:00:00Z"
  }
}
```

---

## Max Users Configuration

### UI Implementation

#### Form Field HTML
**Location**: `web/templates/partials/scripts.html` (lines 681-699)

```html
<div style="margin-bottom: 24px;">
    <label style="display: block; color: #2d3748; font-weight: 600; margin-bottom: 8px; font-size: 14px;">
        <i class="fas fa-users" style="margin-right: 8px; color: #667eea;"></i>Maximum Users
    </label>
    <input type="number" id="tenantMaxUsers" min="1" style="
        width: 100%;
        padding: 12px 16px;
        border: 2px solid #e2e8f0;
        border-radius: 8px;
        font-size: 16px;
        transition: all 0.2s;
        box-sizing: border-box;
    " placeholder="e.g., 10 (leave empty for unlimited)"
    onfocus="this.style.borderColor='#667eea'; this.style.boxShadow='0 0 0 3px rgba(102, 126, 234, 0.1)';"
    onblur="this.style.borderColor='#e2e8f0'; this.style.boxShadow='none';">
    <small style="color: #718096; font-size: 12px; margin-top: 4px; display: block;">
        Leave empty for unlimited users. Set a number to limit users for this tenant.
    </small>
</div>
```

#### JavaScript Validation
**Location**: `web/templates/partials/scripts.html` (lines 770-804)

```javascript
function submitTenantForm(event) {
    event.preventDefault();

    // Get form values
    const maxUsersInput = document.getElementById('tenantMaxUsers').value.trim();

    // Validate max_users if provided
    let maxUsers = null;
    if (maxUsersInput && maxUsersInput !== '') {
        maxUsers = parseInt(maxUsersInput);
        if (isNaN(maxUsers) || maxUsers < 1) {
            showNotification('Maximum users must be a positive number.', 'warning', 4000);
            return;
        }
    }

    // Include in request
    const tenantData = {
        // ... other fields ...
        max_users: maxUsers  // null for unlimited, number for limit
    };

    // Send to backend
    fetch('/api/v1/tenants', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(tenantData),
        credentials: 'include'
    })
    .then(response => response.json())
    .then(data => {
        if (data.error) {
            showNotification(data.error, 'error', 5000);
        } else {
            showNotification('Tenant created successfully!', 'success', 4000);
            closeCreateTenantModal();
            refreshTenantList();
        }
    });
}
```

### Best Practices for max_users

#### 1. **Unlimited by Default**
```javascript
// Empty field = null = unlimited users
// This is the safest default for new tenants
let maxUsers = null;  // Unlimited
```

#### 2. **Validation Rules**
```javascript
// Must be positive integer if set
if (maxUsers !== null && (maxUsers < 1 || !Number.isInteger(maxUsers))) {
    return error("max_users must be a positive integer");
}
```

#### 3. **Plan-Based Defaults** (Recommended)
```javascript
// Auto-populate based on selected plan
document.getElementById('tenantPlan').addEventListener('change', (e) => {
    const planDefaults = {
        'basic': 5,
        'pro': 25,
        'enterprise': null  // unlimited
    };
    document.getElementById('tenantMaxUsers').value = planDefaults[e.target.value] || '';
});
```

---

## API Endpoints

### Admin Dashboard Endpoints

#### `GET /admin`
Render the main admin dashboard.

**Authentication**: Session-based (cookie: `auth_token`)

**Response**: HTML template with dashboard

**Template Variables**:
```go
{
    "User": {
        "Email": "admin@saas.com",
        "Role": "super_admin"
    },
    "Stats": {
        "TotalTenants": 47,
        "ActiveTenants": 42,
        "MonthlyRevenue": 24580
    }
}
```

---

### Tenant Management Endpoints

#### `POST /api/v1/tenants`
Create a new tenant (proxied to Tenant Admin Service).

**Authentication**: Required (JWT or session)

**Request**:
```json
{
  "name": "Acme Corp",
  "domain": "acme-corp",
  "contact_email": "contact@acme.com",
  "admin_email": "admin@acme.com",
  "admin_password": "SecurePass123!",
  "plan": "pro",
  "max_users": 25
}
```

**Responses**:
- `201 Created` - Tenant provisioned successfully
- `400 Bad Request` - Invalid input (e.g., max_users = 0)
- `409 Conflict` - Slug/email already exists
- `500 Internal Server Error` - Provisioning failed

**Implementation**:
```go
func (h *Handler) CreateTenant(c *gin.Context) {
    var req CreateTenantRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Validate max_users
    if req.MaxUsers != nil && *req.MaxUsers <= 0 {
        c.JSON(400, gin.H{"error": "max_users must be greater than 0"})
        return
    }

    // Proxy to Tenant Admin Service
    resp, err := h.tenantAdminClient.CreateTenant(c.Request.Context(), req)
    if err != nil {
        h.logger.Error("Failed to create tenant", zap.Error(err))
        c.JSON(500, gin.H{"error": "Failed to provision tenant"})
        return
    }

    c.JSON(201, resp)
}
```

#### `GET /api/v1/tenants`
List all tenants with filtering and pagination.

**Authentication**: Required

**Query Parameters**:
- `limit` (integer, default: 20, max: 100)
- `offset` (integer, default: 0)
- `status` (string: active|inactive|suspended)
- `plan` (string: basic|pro|enterprise)

**Response**:
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Acme Corp",
      "slug": "acme-corp",
      "status": "active",
      "plan": "pro",
      "max_users": 25,
      "current_users": 12,
      "created_at": "2025-10-01T10:00:00Z"
    }
  ],
  "total": 47,
  "limit": 20,
  "offset": 0
}
```

---

## Authentication System

### Session-Based Authentication

**Flow**:
```
┌──────────────┐
│ Login Form   │
│ (Email/Pass) │
└──────┬───────┘
       │
       ▼
┌──────────────────────────┐
│ POST /api/v1/auth/login  │
│ Credentials validated    │
└──────┬───────────────────┘
       │
       │ Valid
       ▼
┌────────────────────────────────┐
│ Create Session                 │
│ - Generate session token       │
│ - Store in tenant-admin DB     │
│ - Return auth_token cookie     │
└──────┬─────────────────────────┘
       │
       ▼
┌────────────────────────┐
│ Set HttpOnly Cookie    │
│ Name: auth_token       │
│ Secure: true (prod)    │
│ SameSite: Strict       │
│ MaxAge: 24h            │
└────────────────────────┘
```

**Session Validation Middleware**:
```go
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        cookie, err := c.Cookie("auth_token")
        if err != nil || cookie == "" {
            c.Redirect(302, "/login")
            c.Abort()
            return
        }

        // Validate session with Tenant Admin Service
        valid, user := h.validateSession(cookie)
        if !valid {
            c.Redirect(302, "/login")
            c.Abort()
            return
        }

        c.Set("user", user)
        c.Next()
    }
}
```

### Security Features

**1. HttpOnly Cookies** - Prevents XSS attacks
```go
c.SetCookie("auth_token", token, 86400, "/", "", true, true)
//                                              ↑     ↑
//                                           Secure HttpOnly
```

**2. CSRF Protection** - SameSite attribute
```go
c.SetSameSite(http.SameSiteStrictMode)
```

**3. Password Hashing** - Bcrypt with cost 10
```go
hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
```

**4. Session Delegation** - Tenant Admin manages sessions
```
SaaS Admin doesn't store sessions locally.
All session validation is delegated to Tenant Admin Service.
This ensures single source of truth for authentication.
```

---

## Service Interactions

### Microservice Communication Pattern

```
┌────────────────────┐
│  Browser/Client    │
│  (Admin User)      │
└─────────┬──────────┘
          │ HTTP
          ▼
┌────────────────────────────┐
│  SaaS Admin Service        │
│  (Port 8098)               │
│                            │
│  Responsibilities:         │
│  • Render UI               │
│  • Validate input          │
│  • Proxy API calls         │
│  • Aggregate data          │
└─────────┬──────────────────┘
          │
          │ HTTP/JSON
          ▼
┌────────────────────────────┐
│  Tenant Admin Service      │
│  (Port 8099)               │
│                            │
│  Responsibilities:         │
│  • Tenant provisioning     │
│  • User management         │
│  • Session management      │
│  • max_users enforcement   │
└─────────┬──────────────────┘
          │
          │ SQL
          ▼
┌────────────────────────────┐
│  PostgreSQL                │
│  tenant_admin_db           │
└────────────────────────────┘
```

### API Client Pattern

**Resilient HTTP Client**:
```go
type TenantAdminClient struct {
    baseURL    string
    httpClient *http.Client
    logger     *zap.Logger
}

func (c *TenantAdminClient) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
    url := fmt.Sprintf("%s/api/v1/public/tenants", c.baseURL)

    jsonData, _ := json.Marshal(req)
    httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("failed to call tenant admin service: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 201 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("tenant creation failed: %s", string(body))
    }

    var result struct {
        Data *Tenant `json:"data"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    return result.Data, nil
}
```

---

## UI/UX Architecture

### Component Structure

```
web/templates/
├── dashboard.html              # Main layout
├── login.html                  # Login page
└── partials/
    ├── header.html             # Top navigation bar
    ├── sidebar.html            # Left navigation menu
    ├── dashboard.html          # Dashboard content
    ├── scripts.html            # JavaScript logic
    └── styles.html             # Custom CSS
```

### Dashboard Layout

```
┌─────────────────────────────────────────────────────────┐
│                      Header                              │
│  [Logo] Enterprise Dashboard    [Refresh] [Logout]      │
└─────────────────────────────────────────────────────────┘
│         │                                                │
│ Sidebar │             Main Content Area                 │
│         │                                                │
│ ▸ Overview│  ┌──────────────────────────────────────┐  │
│ ▸ Tenants │  │ Stats Cards                          │  │
│ ▸ Billing │  │ [Total Tenants] [Revenue] [Churn]   │  │
│ ▸ Analytics│  └──────────────────────────────────────┘  │
│ ▸ Settings│                                             │
│         │  ┌──────────────────────────────────────┐  │
│         │  │ Tenant List Table                    │  │
│         │  │ [Search] [Filter]                    │  │
│         │  │ ┌────┬────────┬────────┬─────────┐  │  │
│         │  │ │Name│Plan    │Users   │Actions  │  │  │
│         │  │ └────┴────────┴────────┴─────────┘  │  │
│         │  └──────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Key UI Components

#### 1. **Tenant Creation Modal**
- Modal overlay with form
- Real-time validation
- Plan selection dropdown
- **Max users input field** (new)
- AJAX submission

#### 2. **Tenant List Table**
- Searchable and filterable
- Pagination support
- Action buttons (Edit, Delete, View)
- Status indicators

#### 3. **Dashboard Widgets**
- Revenue chart (Chart.js)
- Tenant count metrics
- User growth graph
- Churn rate indicator

---

## Best Practices

### 1. **Input Validation**
```javascript
// Client-side
if (!email.match(/^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$/)) {
    showNotification('Invalid email format', 'warning');
    return;
}

// Server-side
if req.MaxUsers != nil && *req.MaxUsers <= 0 {
    return c.JSON(400, gin.H{"error": "max_users must be > 0"})
}
```

### 2. **Error Handling**
```javascript
fetch('/api/v1/tenants', {...})
    .then(response => {
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }
        return response.json();
    })
    .catch(error => {
        showNotification(`Failed to create tenant: ${error.message}`, 'error');
    });
```

### 3. **Logging**
```go
h.logger.Info("Tenant created",
    zap.String("tenant_id", tenant.ID),
    zap.String("created_by", user.Email),
    zap.Int("max_users", *tenant.MaxUsers))
```

### 4. **Session Security**
```go
// Always use HTTPS in production
if config.Environment == "production" {
    c.SetCookie("auth_token", token, maxAge, "/", "", true, true)
    //                                              ↑
    //                                         Secure=true
}
```

---

## Configuration Reference

### Environment Variables

```bash
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8098
BASE_DOMAIN=localhost

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=saas_admin
DB_SSL_MODE=disable

# JWT Configuration
JWT_SECRET=your-secret-key-min-32-chars
JWT_EXPIRATION_HOURS=24

# Tenant Admin Service
TENANT_ADMIN_SERVICE_URL=http://localhost:8099

# Session Management
SESSION_TIMEOUT=86400  # 24 hours
SESSION_CLEANUP_INTERVAL=3600  # 1 hour

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2025-10-02 | Initial architecture documentation |
| 1.1.0 | 2025-10-02 | Added max_users feature to tenant creation |

---

**Last Updated**: 2025-10-02
**Maintained By**: Platform Engineering Team
