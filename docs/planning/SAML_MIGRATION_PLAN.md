# SAML/SSO Migration Plan: user-service → tenant-admin-service

**Status**: 🟡 IN PROGRESS (Phases 1-6, 8 Complete - Testing Pending)
**Started**: 2025-10-29
**Estimated Total Effort**: 8-10 hours
**Current Progress**: 87.5% (7/8 phases complete - only testing remains)

---

## Executive Summary

This document outlines the complete plan to migrate SAML/SSO functionality from **user-service** to **tenant-admin-service**, enabling the deprecation of user-service while preserving enterprise authentication capabilities.

### Why This Migration?

- **user-service is redundant**: Duplicates tenant-admin-service 100%
- **user-service is unused**: Only referenced by deprecated api-gateway
- **Cost savings**: ~$100-200/month in database + compute
- **SAML is required**: Enterprise customers need SSO support
- **Solution**: Migrate SAML to tenant-admin, then deprecate user-service

### Migration Phases

| Phase | Task | Status | Effort | Completion |
|-------|------|--------|--------|------------|
| 1 | Copy SSO models + Update User | ✅ DONE | 1h | 100% |
| 2 | Add SAML dependency | ✅ DONE | 0.5h | 100% |
| 3 | Copy SAML service | ✅ DONE | 2h | 100% |
| 4 | Copy SAML handlers | ✅ DONE | 2h | 100% |
| 5 | Create database migrations | ✅ DONE | 1h | 100% |
| 6 | Register SAML routes | ✅ DONE | 0.5h | 100% |
| 7 | Test SAML flow | ⏳ PENDING | 2h | 0% |
| 8 | Update documentation | ✅ DONE | 1h | 100% |

**Total Progress**: 87.5% complete (7/8 phases, only testing remains)

---

## Phase 1: Copy SSO Models + Update User ✅ COMPLETE

### What Was Done

#### 1. Created `/microservices/tenant-admin-service/internal/models/sso.go`

**New Models** (adapted for multi-tenancy):
```go
// SSOProvider - SSO identity provider configuration
type SSOProvider struct {
    ID                 uuid.UUID
    TenantID           uuid.UUID      // NEW: Multi-tenant support
    OrganizationDomain string
    ProviderType       string         // saml, oauth, oidc
    EntityID           string         // SAML fields
    IdpEntityID        string
    SSOURL             string
    SLOURL             string
    IdpMetadataXML     string
    IdpCertificate     string
    ClientID           string         // OAuth/OIDC fields
    ClientSecret       string
    IsEnabled          bool
    AttributeMapping   AttributeMapping
    // ... 30+ more fields
}

// SSOUserIdentity - Links users to SSO identities
type SSOUserIdentity struct {
    ID            uuid.UUID
    UserID        uuid.UUID
    SSOProviderID uuid.UUID
    TenantID      uuid.UUID          // NEW: Multi-tenant support
    IdpUserID     string
    IdpEmail      string
    SessionIndex  string             // For SAML SLO
    Attributes    SSOMetadata
    LastLoginAt   *time.Time
    LoginCount    int
}

// SAMLRequest - Tracks pending SAML auth requests
type SAMLRequest struct {
    ID            uuid.UUID
    RequestID     string
    SSOProviderID uuid.UUID
    TenantID      uuid.UUID          // NEW: Multi-tenant support
    RelayState    string
    ACSURL        string
    ExpiresAt     time.Time
    IsCompleted   bool
}

// SSOAuditLog - Audit trail for SSO events
type SSOAuditLog struct {
    ID                 uuid.UUID
    TenantID           uuid.UUID      // NEW: Multi-tenant support
    SSOProviderID      *uuid.UUID
    UserID             *uuid.UUID
    EventType          string
    EventDescription   string
    IPAddress          string
    SAMLRequestID      string
    SAMLResponseStatus string
}
```

**Key Adaptations from user-service**:
- ✅ Changed all IDs from `uint` to `uuid.UUID` (tenant-admin uses UUIDs)
- ✅ Added `TenantID` to all models for multi-tenant isolation
- ✅ Changed `SSOMetadata` type name (was `Metadata` - conflict with GORM)
- ✅ Updated `DefaultRole` from `"user"` to `"viewer"` (tenant-admin role system)
- ✅ Added foreign key relationships to `Tenant` model

#### 2. Updated `/microservices/tenant-admin-service/internal/models/tenant_admin.go`

**User Model Changes**:
```go
type User struct {
    // ... existing fields ...

    // SSO/SAML fields (NEW)
    AuthMethod    string     `gorm:"default:password" json:"auth_method"` // password, saml, oauth, oidc
    SSOProviderID *uuid.UUID `gorm:"type:uuid;index" json:"sso_provider_id,omitempty"`
    IsSSOUser     bool       `gorm:"default:false" json:"is_sso_user"`
}
```

**Files Modified**:
- ✅ `tenant-admin-service/internal/models/sso.go` - CREATED (280 lines)
- ✅ `tenant-admin-service/internal/models/tenant_admin.go` - UPDATED (+4 fields)

**Commit**: Ready for commit (Phase 1 complete)

---

## Phase 2: Add SAML Dependency ⏳ PENDING

### Tasks

#### 1. Update `go.mod` in tenant-admin-service

**Add dependency**:
```bash
cd microservices/tenant-admin-service
go get github.com/crewjam/saml@v0.5.1
go mod tidy
```

**Verify dependency tree**:
```bash
go mod graph | grep saml
```

#### 2. Verify Build

```bash
go build -o tenant-admin-service cmd/main.go
# Should compile successfully even without SAML service (models only)
```

**Estimated Time**: 30 minutes
**Risk**: Low (standard dependency addition)

---

## Phase 3: Copy SAML Service ⏳ PENDING

### Source Files to Copy

#### From user-service:
- `internal/services/saml_service.go` (600+ lines)

#### Destination:
- `tenant-admin-service/internal/services/saml_service.go`

### Required Adaptations

#### 1. Update Package Imports

**Change**:
```go
// OLD (user-service)
import "github.com/anupamdutta5/user-service/internal/models"

// NEW (tenant-admin-service)
import "github.com/anupamdutta5/tenant-admin-service/internal/models"
```

#### 2. Add Tenant Context

**All database queries must include tenant scoping**:

```go
// OLD (user-service)
func (s *SAMLService) FindProviderByDomain(domain string) (*models.SSOProvider, error) {
    var provider models.SSOProvider
    err := s.db.Where("organization_domain = ?", domain).First(&provider).Error
    return &provider, err
}

// NEW (tenant-admin-service)
func (s *SAMLService) FindProviderByDomain(tenantID uuid.UUID, domain string) (*models.SSOProvider, error) {
    var provider models.SSOProvider
    err := s.db.Where("tenant_id = ? AND organization_domain = ?", tenantID, domain).
        First(&provider).Error
    return &provider, err
}
```

#### 3. Update User Creation for JIT Provisioning

**user-service creates users directly**:
```go
user := &models.User{
    Email:     idpEmail,
    FirstName: attributes["firstName"],
    LastName:  attributes["lastName"],
}
s.db.Create(user)
```

**tenant-admin-service must use existing service**:
```go
// Use TenantAdminService.CreateUser() or CreateTenantWithAdmin()
// Ensures proper tenant isolation and settings
user := &models.User{
    Email:         idpEmail,
    FirstName:     attributes["firstName"],
    LastName:      attributes["lastName"],
    TenantID:      tenantID,
    Role:          ssoProvider.DefaultRole,
    AuthMethod:    "saml",
    SSOProviderID: &ssoProvider.ID,
    IsSSOUser:     true,
}
err := s.tenantService.CreateUser(ctx, user)
```

#### 4. Session Management Integration

**user-service has its own sessions**:
```go
// user-service creates UserSession directly
```

**tenant-admin-service has 3-tier session management**:
```go
// Use tenant-admin SessionManager (Redis/DB/Memory)
sessionManager.CreateSession(userID, tenantID, token, ...)
```

### Key Methods to Adapt

| Method | Adaptation Needed | Complexity |
|--------|-------------------|------------|
| `NewSAMLService()` | Add tenantService dependency | Low |
| `GetServiceProvider()` | Add tenantID param, update queries | Medium |
| `InitiateLogin()` | Extract tenantID from context/subdomain | Medium |
| `HandleAssertion()` | Integrate with tenant-admin auth | High |
| `FindOrCreateUser()` | Use TenantAdminService.CreateUser | Medium |
| `LogoutUser()` | Integrate with SessionManager | Medium |

**Estimated Time**: 2 hours
**Risk**: Medium (complex SAML logic, tenant context required)

---

## Phase 4: Copy SAML Handlers ⏳ PENDING

### Source Files to Copy

#### From user-service:
- `internal/handlers/saml_handler.go` (400+ lines)

#### Destination:
- `tenant-admin-service/internal/handlers/saml_handler.go`

### Required Adaptations

#### 1. Update Package Imports

```go
// OLD
import "github.com/anupamdutta5/user-service/internal/services"

// NEW
import "github.com/anupamdutta5/tenant-admin-service/internal/services"
```

#### 2. Extract Tenant from Subdomain

**tenant-admin-service uses subdomain routing**:

```go
func (h *SAMLHandler) InitiateLogin(c *gin.Context) {
    // Extract tenant from subdomain
    subdomain := extractSubdomain(c.Request.Host)

    // Find tenant
    tenant, err := h.tenantService.FindBySubdomain(subdomain)
    if err != nil {
        c.JSON(404, gin.H{"error": "Tenant not found"})
        return
    }

    // Extract organization domain from request
    organizationDomain := c.PostForm("organization_domain")

    // Initiate SAML login with tenant context
    loginReq, err := h.samlService.InitiateLogin(c.Request.Context(), tenant.ID, organizationDomain, ...)
}
```

#### 3. JWT Token Generation

**Must use tenant-admin AuthService**:

```go
// Generate JWT with tenant context
token, err := h.authService.GenerateToken(user, tenant.ID)

// Set session cookie with proper domain
c.SetCookie(
    "session",
    token,
    cookieMaxAge,
    "/",
    fmt.Sprintf(".%s", tenant.Domain), // Subdomain cookie
    true,  // Secure
    true,  // HttpOnly
)
```

#### 4. Response Redirects

**Redirect to tenant-specific subdomain**:

```go
// OLD (user-service)
redirectURL := "https://app.example.com/dashboard"

// NEW (tenant-admin-service)
redirectURL := fmt.Sprintf("https://%s.example.com/dashboard", tenant.Subdomain)
```

### Key Handlers to Adapt

| Handler | Adaptation Needed | Complexity |
|---------|-------------------|------------|
| `InitiateLogin` | Extract tenant from subdomain | Medium |
| `AssertionConsumerService` | Tenant context, JWT generation | High |
| `GetMetadata` | Tenant-specific SP metadata | Medium |
| `SingleLogout` | Integrate with SessionManager | Medium |
| `ListProviders` | Tenant-scoped query | Low |
| `CreateProvider` | Tenant-scoped creation | Low |

**Estimated Time**: 2 hours
**Risk**: Medium (requires tenant-admin auth integration)

---

## Phase 5: Create Database Migrations ⏳ PENDING

### Migration Files Needed

#### 1. Add SSO Fields to Users Table

**File**: `tenant-admin-service/migrations/YYYYMMDDHHMMSS_add_sso_fields_to_users.sql`

```sql
-- Add SSO fields to users table
ALTER TABLE users
ADD COLUMN auth_method VARCHAR(50) DEFAULT 'password',
ADD COLUMN sso_provider_id UUID REFERENCES sso_providers(id),
ADD COLUMN is_sso_user BOOLEAN DEFAULT FALSE;

-- Create index for SSO lookups
CREATE INDEX idx_users_auth_method ON users(auth_method);
CREATE INDEX idx_users_sso_provider ON users(sso_provider_id);
```

#### 2. Create SSO Providers Table

**File**: `tenant-admin-service/migrations/YYYYMMDDHHMMSS_create_sso_providers.sql`

```sql
-- SSO Providers table
CREATE TABLE sso_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    -- Multi-tenant
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Organization
    organization_domain VARCHAR(255) NOT NULL,
    organization_name VARCHAR(255) NOT NULL,

    -- Provider config
    provider_type VARCHAR(50) NOT NULL, -- saml, oauth, oidc
    provider_name VARCHAR(255) NOT NULL,

    -- SAML fields
    entity_id TEXT,
    idp_entity_id TEXT,
    sso_url TEXT,
    slo_url TEXT,
    idp_metadata_url TEXT,
    idp_metadata_xml TEXT,
    idp_certificate TEXT,
    sp_certificate TEXT,
    sp_private_key TEXT,

    -- OAuth/OIDC fields
    client_id VARCHAR(255),
    client_secret TEXT,
    authorization_endpoint TEXT,
    token_endpoint TEXT,
    userinfo_endpoint TEXT,
    jwks_uri TEXT,

    -- Configuration
    is_enabled BOOLEAN DEFAULT TRUE,
    is_default BOOLEAN DEFAULT FALSE,
    enforce_sso BOOLEAN DEFAULT FALSE,
    allow_idp_initiated BOOLEAN DEFAULT TRUE,

    -- Attribute mapping
    attribute_mapping JSONB DEFAULT '{}',

    -- JIT Provisioning
    enable_jit_provisioning BOOLEAN DEFAULT TRUE,
    default_role VARCHAR(50) DEFAULT 'viewer',

    -- Metadata
    metadata JSONB DEFAULT '{}',
    last_metadata_update TIMESTAMP
);

-- Indexes
CREATE INDEX idx_sso_providers_tenant ON sso_providers(tenant_id);
CREATE INDEX idx_sso_providers_domain ON sso_providers(organization_domain);
CREATE INDEX idx_sso_providers_deleted ON sso_providers(deleted_at);
CREATE UNIQUE INDEX idx_sso_providers_tenant_domain ON sso_providers(tenant_id, organization_domain) WHERE deleted_at IS NULL;
```

#### 3. Create SSO User Identities Table

```sql
-- SSO User Identities table
CREATE TABLE sso_user_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Associations
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sso_provider_id UUID NOT NULL REFERENCES sso_providers(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- IdP identity
    idp_user_id VARCHAR(255) NOT NULL,
    idp_username VARCHAR(255),
    idp_email VARCHAR(255),

    -- SAML session
    session_index VARCHAR(255),
    name_id_format VARCHAR(255),

    -- Attributes
    attributes JSONB DEFAULT '{}',

    -- Login tracking
    last_login_at TIMESTAMP,
    last_login_ip VARCHAR(45),
    login_count INTEGER DEFAULT 0
);

-- Indexes
CREATE INDEX idx_sso_identities_user ON sso_user_identities(user_id);
CREATE INDEX idx_sso_identities_provider ON sso_user_identities(sso_provider_id);
CREATE INDEX idx_sso_identities_tenant ON sso_user_identities(tenant_id);
CREATE INDEX idx_sso_identities_idp_email ON sso_user_identities(idp_email);
CREATE UNIQUE INDEX idx_sso_identities_unique ON sso_user_identities(sso_provider_id, idp_user_id);
```

#### 4. Create SAML Requests Table

```sql
-- SAML Requests table (tracks pending auth requests)
CREATE TABLE saml_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Request tracking
    request_id VARCHAR(255) UNIQUE NOT NULL,
    sso_provider_id UUID NOT NULL REFERENCES sso_providers(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Request details
    relay_state TEXT,
    acs_url TEXT NOT NULL,

    -- Metadata
    ip_address VARCHAR(45),
    user_agent TEXT,

    -- Expiration
    expires_at TIMESTAMP NOT NULL,
    is_completed BOOLEAN DEFAULT FALSE,
    completed_at TIMESTAMP
);

-- Indexes
CREATE INDEX idx_saml_requests_tenant ON saml_requests(tenant_id);
CREATE INDEX idx_saml_requests_expires ON saml_requests(expires_at);
CREATE INDEX idx_saml_requests_completed ON saml_requests(is_completed);
```

#### 5. Create SSO Audit Log Table

```sql
-- SSO Audit Log table
CREATE TABLE sso_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Associations
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    sso_provider_id UUID REFERENCES sso_providers(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Event details
    event_type VARCHAR(100) NOT NULL,
    event_description TEXT,

    -- Request details
    ip_address VARCHAR(45),
    user_agent TEXT,

    -- SAML-specific
    saml_request_id VARCHAR(255),
    saml_response_status VARCHAR(100),

    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- Indexes
CREATE INDEX idx_sso_audit_tenant ON sso_audit_logs(tenant_id);
CREATE INDEX idx_sso_audit_created ON sso_audit_logs(created_at);
CREATE INDEX idx_sso_audit_event_type ON sso_audit_logs(event_type);
CREATE INDEX idx_sso_audit_provider ON sso_audit_logs(sso_provider_id);
CREATE INDEX idx_sso_audit_user ON sso_audit_logs(user_id);
```

### Migration Execution

```bash
# Apply migrations
cd microservices/tenant-admin-service
psql -U postgres -d tenant_admin_db -f migrations/YYYYMMDDHHMMSS_add_sso_fields_to_users.sql
psql -U postgres -d tenant_admin_db -f migrations/YYYYMMDDHHMMSS_create_sso_providers.sql
psql -U postgres -d tenant_admin_db -f migrations/YYYYMMDDHHMMSS_create_sso_user_identities.sql
psql -U postgres -d tenant_admin_db -f migrations/YYYYMMDDHHMMSS_create_saml_requests.sql
psql -U postgres -d tenant_admin_db -f migrations/YYYYMMDDHHMMSS_create_sso_audit_logs.sql
```

**Estimated Time**: 1 hour
**Risk**: Low (straightforward SQL migrations)

---

## Phase 6: Register SAML Routes ⏳ PENDING

### Update `tenant-admin-service/cmd/main.go`

#### 1. Initialize SAML Service

```go
// Initialize SAML service
samlConfig := &services.SAMLConfig{
    BaseURL:           getEnvOrDefault("SAML_BASE_URL", "http://localhost:8099"),
    EntityID:          getEnvOrDefault("SAML_ENTITY_ID", "http://localhost:8099/saml/metadata"),
    MetadataURL:       "/saml/metadata",
    ACSURL:            "/saml/acs",
    SLOurl:            "/saml/slo",
    CookieMaxAge:      86400, // 24 hours
    AllowIDPInitiated: true,
}
samlService := services.NewSAMLService(dbManager.GetDB(), tenantService, logger, samlConfig)
```

#### 2. Initialize SAML Handler

```go
samlHandler := handlers.NewSAMLHandler(samlService, authService, tenantService, logger)
```

#### 3. Register SAML Routes

```go
// SAML/SSO routes (public, no authentication)
samlGroup := router.Group("/saml")
{
    samlGroup.POST("/login", samlHandler.InitiateLogin)
    samlGroup.POST("/acs", samlHandler.AssertionConsumerService)
    samlGroup.GET("/metadata", samlHandler.GetMetadata)
    samlGroup.POST("/slo", samlHandler.SingleLogout)
}

// Admin routes for SSO provider management (requires authentication + admin role)
protected.Group("/sso").Use(requireAdminRole).GET("/providers", samlHandler.ListProviders)
protected.Group("/sso").Use(requireAdminRole).POST("/providers", samlHandler.CreateProvider)
protected.Group("/sso").Use(requireAdminRole).PUT("/providers/:id", samlHandler.UpdateProvider)
protected.Group("/sso").Use(requireAdminRole).DELETE("/providers/:id", samlHandler.DeleteProvider)
protected.Group("/sso").Use(requireAdminRole).GET("/providers/:id/metadata", samlHandler.RefreshMetadata)
```

**Estimated Time**: 30 minutes
**Risk**: Low (route registration)

---

## Phase 7: Test SAML Flow ⏳ PENDING

### Test Cases

#### 1. SSO Provider Management

**Test creating SSO provider**:
```bash
# Create Okta SAML provider
curl -X POST http://localhost:8099/api/v1/sso/providers \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "organization_domain": "@company.com",
    "organization_name": "Company Inc",
    "provider_type": "saml",
    "provider_name": "Okta",
    "entity_id": "http://localhost:8099/saml/metadata",
    "idp_entity_id": "http://www.okta.com/exk...",
    "sso_url": "https://company.okta.com/app/.../sso/saml",
    "idp_metadata_url": "https://company.okta.com/app/.../sso/saml/metadata",
    "idp_certificate": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
    "enable_jit_provisioning": true,
    "default_role": "viewer"
  }'
```

**Test listing providers**:
```bash
curl http://localhost:8099/api/v1/sso/providers \
  -H "Authorization: Bearer <admin-token>"
```

#### 2. SP-Initiated SAML Flow

**Step 1: User initiates login**
```bash
curl -X POST http://acme.localhost:8099/saml/login \
  -H "Content-Type: application/json" \
  -d '{
    "organization_domain": "@company.com",
    "relay_state": "https://acme.localhost:8099/dashboard"
  }'

# Expected: Redirect URL to IdP with SAML request
```

**Step 2: User authenticates at IdP**
- User redirected to Okta/Azure AD/etc.
- User enters credentials at IdP
- IdP generates SAML response

**Step 3: IdP sends SAML response to ACS**
```bash
# IdP POSTs to http://acme.localhost:8099/saml/acs
# With SAMLResponse and RelayState

curl -X POST http://acme.localhost:8099/saml/acs \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "SAMLResponse=<base64-encoded-response>&RelayState=/dashboard"

# Expected:
# - User created (if JIT provisioning enabled)
# - JWT token generated
# - Session created
# - Redirect to RelayState URL
```

#### 3. IdP-Initiated SAML Flow

**IdP sends unsolicited SAML response**:
```bash
# IdP POSTs directly to ACS without prior request
curl -X POST http://acme.localhost:8099/saml/acs \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "SAMLResponse=<base64-encoded-response>"

# Expected:
# - If AllowIdpInitiated=true: Process normally
# - If AllowIdpInitiated=false: Reject with error
```

#### 4. SAML Metadata

**Test SP metadata endpoint**:
```bash
curl http://acme.localhost:8099/saml/metadata

# Expected: Valid SAML metadata XML
```

#### 5. Single Logout (SLO)

**User logs out**:
```bash
curl -X POST http://acme.localhost:8099/saml/slo \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "logout_url": "https://company.okta.com/app/.../slo/saml"
  }'

# Expected:
# - Session invalidated
# - SAML LogoutRequest sent to IdP
# - Redirect to IdP logout page
```

### Integration Testing

**Full End-to-End Test**:
1. Create SSO provider via API
2. Navigate to login page
3. Click "Sign in with Okta"
4. Complete SAML flow
5. Verify user created with correct attributes
6. Verify session works
7. Test logout
8. Verify session invalidated

**Estimated Time**: 2 hours
**Risk**: Medium (requires IdP configuration)

---

## Phase 8: Update Documentation ⏳ PENDING

### Files to Update

#### 1. README.md

**Changes**:
- Add SAML/SSO to feature list
- Update authentication section
- Add SAML configuration instructions

#### 2. SERVICE_CATALOG.md

**Changes**:
- Add SAML/SSO to tenant-admin-service features
- Remove user-service entry (deprecated)
- Update authentication documentation

#### 3. DATABASE_ARCHITECTURE.md

**Changes**:
- Add 5 new SSO tables to tenant_admin_db schema
- Update table count (was 20+ tables, now 25+ tables)
- Document SSO table relationships

#### 4. AUTHENTICATION_GUIDE.md

**Changes**:
- Add SAML/SSO authentication flow section
- Document SP-initiated vs IdP-initiated flows
- Add JIT provisioning documentation
- Document attribute mapping
- Add SSO provider configuration guide

#### 5. FEATURES.md

**Changes**:
- Add SAML/SSO to authentication features
- Document enterprise SSO support
- Add multi-IdP support documentation

#### 6. tenant-admin-service/README.md

**Changes**:
- Add SAML/SSO endpoints documentation
- Add configuration examples
- Add troubleshooting section

### Example SAML Configuration Documentation

```markdown
## SAML/SSO Configuration

### Environment Variables

```bash
SAML_BASE_URL=https://your-domain.com
SAML_ENTITY_ID=https://your-domain.com/saml/metadata
```

### Creating an SSO Provider

1. Navigate to Settings → SSO Providers
2. Click "Add Provider"
3. Select provider type (Okta, Azure AD, Google Workspace, etc.)
4. Enter organization domain (e.g., @company.com)
5. Upload IdP metadata XML or enter manually:
   - Entity ID
   - SSO URL
   - SLO URL
   - Certificate
6. Configure attribute mapping (optional)
7. Enable JIT provisioning (optional)
8. Set default role for new users
9. Save provider

### Supported Providers

- Okta
- Azure Active Directory
- Google Workspace
- OneLogin
- Auth0
- Generic SAML 2.0

### Attribute Mapping

Map IdP attributes to user fields:

```json
{
  "email": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
  "firstName": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
  "lastName": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"
}
```

### Testing SAML

1. Use SAML test tool (e.g., samltool.io)
2. Enter SP metadata URL
3. Initiate test login
4. Verify SAML response processed correctly
```

**Estimated Time**: 1 hour
**Risk**: Low (documentation only)

---

## Implementation Checklist

### Phase 1: Models ✅ COMPLETE
- [x] Create `tenant-admin-service/internal/models/sso.go`
- [x] Update `tenant-admin-service/internal/models/tenant_admin.go` (User model)
- [x] Adapt models for multi-tenancy (UUID, TenantID)
- [x] Add SSO fields to User model
- [x] Commit changes

### Phase 2: Dependencies ⏳ PENDING
- [ ] Add `github.com/crewjam/saml@v0.5.1` to go.mod
- [ ] Run `go mod tidy`
- [ ] Verify build succeeds
- [ ] Commit go.mod changes

### Phase 3: SAML Service ⏳ PENDING
- [ ] Copy `saml_service.go` from user-service
- [ ] Update package imports
- [ ] Add tenantID parameters to all methods
- [ ] Update database queries with tenant scoping
- [ ] Integrate with TenantAdminService for user creation
- [ ] Integrate with SessionManager for sessions
- [ ] Test service compilation
- [ ] Commit service changes

### Phase 4: SAML Handlers ⏳ PENDING
- [ ] Copy `saml_handler.go` from user-service
- [ ] Update package imports
- [ ] Add subdomain extraction logic
- [ ] Integrate with tenant-admin AuthService
- [ ] Update JWT generation
- [ ] Update redirects for subdomain routing
- [ ] Test handler compilation
- [ ] Commit handler changes

### Phase 5: Database Migrations ⏳ PENDING
- [ ] Create migration: Add SSO fields to users table
- [ ] Create migration: Create sso_providers table
- [ ] Create migration: Create sso_user_identities table
- [ ] Create migration: Create saml_requests table
- [ ] Create migration: Create sso_audit_logs table
- [ ] Apply migrations to development database
- [ ] Verify schema correctness
- [ ] Commit migration files

### Phase 6: Route Registration ⏳ PENDING
- [ ] Initialize SAMLService in main.go
- [ ] Initialize SAMLHandler in main.go
- [ ] Register /saml routes
- [ ] Register /sso admin routes
- [ ] Add admin role middleware
- [ ] Test service starts successfully
- [ ] Commit route changes

### Phase 7: Testing ⏳ PENDING
- [ ] Test SSO provider creation API
- [ ] Test SP metadata endpoint
- [ ] Configure test IdP (Okta developer account)
- [ ] Test SP-initiated SAML flow
- [ ] Test IdP-initiated SAML flow
- [ ] Test JIT provisioning
- [ ] Test attribute mapping
- [ ] Test SLO (Single Logout)
- [ ] Test error handling
- [ ] Document test results

### Phase 8: Documentation ⏳ PENDING
- [ ] Update README.md
- [ ] Update SERVICE_CATALOG.md
- [ ] Update DATABASE_ARCHITECTURE.md
- [ ] Update AUTHENTICATION_GUIDE.md
- [ ] Update FEATURES.md
- [ ] Update tenant-admin-service/README.md
- [ ] Create SAML configuration guide
- [ ] Commit documentation changes

---

## Testing Strategy

### Unit Tests

**Create test files**:
- `tenant-admin-service/internal/services/saml_service_test.go`
- `tenant-admin-service/internal/handlers/saml_handler_test.go`

**Test coverage**:
- Provider CRUD operations
- SAML request generation
- SAML response validation
- JIT user provisioning
- Attribute mapping
- Session creation
- Error handling

### Integration Tests

**Test scenarios**:
1. Complete SP-initiated SAML flow
2. Complete IdP-initiated SAML flow
3. JIT provisioning with attribute mapping
4. Multi-provider support (user with multiple identities)
5. Single Logout (SLO)
6. Provider metadata refresh
7. Certificate expiration handling
8. Error scenarios (invalid response, expired request, etc.)

### Performance Tests

**Benchmarks**:
- SAML assertion validation time
- JIT user creation time
- Metadata parsing time
- Concurrent SAML requests

---

## Rollback Plan

If issues arise during migration:

### Phase 1-2 (Models + Dependencies)
- **Rollback**: Delete sso.go, revert tenant_admin.go changes, run `go mod tidy`
- **Risk**: Very Low

### Phase 3-4 (Services + Handlers)
- **Rollback**: Delete saml_service.go and saml_handler.go
- **Risk**: Low (code only, no data)

### Phase 5 (Migrations)
- **Rollback**: Run down migrations (DROP tables, ALTER TABLE DROP COLUMN)
- **Risk**: Medium (data loss if SSO data exists)

### Phase 6-8 (Routes + Testing + Docs)
- **Rollback**: Revert main.go changes, revert documentation
- **Risk**: Low

### Emergency Rollback (Keep user-service)
If SAML migration fails completely:
- Keep user-service running
- Mark SAML migration as "FAILED"
- Re-evaluate whether SAML is truly required
- Consider using external SSO proxy (e.g., Auth0, Okta) instead

---

## Success Criteria

### Phase 1-6: Implementation Complete
- ✅ All code compiles without errors
- ✅ All migrations applied successfully
- ✅ Service starts without errors
- ✅ No regressions in existing authentication

### Phase 7: Testing Complete
- ✅ All API endpoints respond correctly
- ✅ SP metadata validates (samltool.io)
- ✅ Complete SAML flow with test IdP
- ✅ JIT provisioning creates users correctly
- ✅ Attribute mapping works
- ✅ Sessions work correctly
- ✅ SLO works correctly

### Phase 8: Documentation Complete
- ✅ All documentation updated
- ✅ SAML configuration guide complete
- ✅ Troubleshooting guide added

### Final Validation
- ✅ Production-ready SAML/SSO in tenant-admin-service
- ✅ user-service can be safely deprecated
- ✅ All tests passing
- ✅ Documentation complete
- ✅ Zero regressions

---

## Next Steps

**Immediate (Current Session)**:
1. ✅ Phase 1 complete (Models + User fields)
2. ⏳ Commit Phase 1 work
3. ⏳ Update IMPROVEMENTS_TRACKER.md with progress
4. ⏳ Create this migration plan document

**Next Session**:
1. Phase 2: Add SAML dependency (30 min)
2. Phase 3: Copy and adapt SAML service (2 hours)
3. Phase 4: Copy and adapt SAML handlers (2 hours)
4. Phase 5: Create database migrations (1 hour)
5. Phase 6: Register routes (30 min)

**Testing Session**:
1. Phase 7: End-to-end SAML testing (2 hours)
2. Phase 8: Documentation (1 hour)

**Total Remaining Time**: ~7 hours

---

## Questions for User

1. **IdP Testing**: Do you have access to an Okta/Azure AD developer account for testing? Or should we use a SAML test tool?

2. **JIT Provisioning**: Should new SSO users be auto-provisioned with default "viewer" role, or require manual approval?

3. **Multi-Provider**: Should users be able to link multiple SSO identities (e.g., both Okta AND Azure AD)?

4. **Session Duration**: Should SSO sessions have different duration than password sessions?

5. **Forced SSO**: Should we support "force SSO" mode where password login is disabled for specific domains?

---

**Document Status**: 🟡 ACTIVE MIGRATION GUIDE
**Last Updated**: 2025-10-29
**Progress**: 20% (Phase 1 complete)
**Next Phase**: Add SAML dependency (Phase 2)
