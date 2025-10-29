# Tenant Admin Service

## Overview
The Tenant Admin Service is the core multi-tenancy management service for the Beakon status page platform. It handles tenant lifecycle management, comprehensive Role-Based Access Control (RBAC), status page management, domain management, and provides administrative interfaces for tenant-specific operations.

## Key Features

### Multi-Tenant Management
- **Tenant CRUD Operations**: Complete tenant lifecycle management
- **Tenant Branding**: Custom branding and theming per tenant
- **Tenant Settings**: Configurable tenant-specific settings
- **Tenant Statistics**: Analytics and usage tracking per tenant

### Role-Based Access Control (RBAC)
- **Role Management**: Dynamic role creation and assignment
- **Permission System**: Granular permission management
- **User Role Assignment**: Flexible user-role associations
- **Team Management**: Hierarchical team structures
- **Team Member Management**: Team membership and role assignments
- **Session Management**: Secure session handling and validation
- **Audit Logging**: Complete audit trail of RBAC operations

### Status Page Management
- **Status Page Creation**: Multi-tenant status page provisioning
- **Status Page Configuration**: Per-tenant status page customization
- **Status Page Data Management**: Content and configuration management

### Domain Management
- **Custom Domain Support**: Tenant-specific custom domains
- **Domain Verification**: DNS-based domain verification
- **Domain Lifecycle**: Domain addition, verification, and removal

### Administrative Functions
- **Admin User Management**: Tenant administrator accounts
- **Feature Flag Management**: Tenant-specific feature toggles
- **Usage Tracking**: Resource usage monitoring and billing integration
- **Notification Management**: Tenant notification preferences
- **Backup Management**: Tenant data backup and restoration

### Security Features
- **JWT Authentication**: Token-based authentication system
- **Session Validation**: Secure session management
- **Audit Logging**: Comprehensive security audit trails
- **Multi-tenant Isolation**: Secure data separation between tenants

### Enterprise Authentication (SAML/SSO)
- **SAML 2.0 Authentication**: Enterprise single sign-on support
- **Multi-Provider Support**: Multiple SSO providers per tenant
- **SP-Initiated Flow**: Service Provider initiated authentication
- **IdP-Initiated Flow**: Identity Provider initiated authentication
- **JIT Provisioning**: Just-In-Time user creation from SAML assertions
- **Attribute Mapping**: Flexible mapping of IdP attributes to user fields
- **SSO Provider Management**: CRUD operations for SSO configurations
- **SSO Audit Logging**: Complete audit trail for SSO events

## API Endpoints

### Health & Authentication
- `GET /health` - Health check
- `GET /health/ready` - Readiness probe
- `GET /health/live` - Liveness probe

### Public Authentication
- `POST /api/v1/public/login` - User login

### Protected Authentication Routes
- `POST /api/v1/auth/logout` - User logout
- `POST /api/v1/auth/verify` - Token verification
- `POST /api/v1/auth/refresh` - Refresh access token

### SAML/SSO Authentication (Public)
- `POST /api/v1/saml/login` - Initiate SAML login (SP-initiated)
- `POST /api/v1/saml/acs` - Assertion Consumer Service (IdP callback)
- `GET /api/v1/saml/metadata` - Get Service Provider metadata XML
- `POST /api/v1/saml/logout` - SAML Single Logout

### SSO Provider Management (Protected)
- `GET /api/v1/sso/providers` - List SSO providers for tenant
- `GET /api/v1/sso/providers/:id` - Get SSO provider details
- `POST /api/v1/sso/providers` - Create SSO provider
- `PUT /api/v1/sso/providers/:id` - Update SSO provider
- `DELETE /api/v1/sso/providers/:id` - Delete SSO provider

### Tenant Management
- `GET /api/v1/tenants` - List tenants
- `POST /api/v1/tenants` - Create tenant
- `GET /api/v1/tenants/:id` - Get tenant details
- `PUT /api/v1/tenants/:id` - Update tenant
- `DELETE /api/v1/tenants/:id` - Delete tenant
- `GET /api/v1/tenants/slug/:slug` - Get tenant by slug
- `GET /api/v1/tenants/domain/:domain` - Get tenant by domain

### Tenant Branding
- `GET /api/v1/branding/:tenant_id` - Get tenant branding
- `PUT /api/v1/branding/:tenant_id` - Update tenant branding

### Tenant Admin Management
- `GET /api/v1/admins` - List tenant admins
- `POST /api/v1/admins` - Create tenant admin
- `GET /api/v1/admins/:id` - Get tenant admin details
- `PUT /api/v1/admins/:id` - Update tenant admin
- `DELETE /api/v1/admins/:id` - Delete tenant admin

### Status Page Management
- `GET /api/v1/status-pages` - List status pages
- `POST /api/v1/status-pages` - Create status page
- `GET /api/v1/status-pages/:id` - Get status page data
- `PUT /api/v1/status-pages/:id` - Update status page
- `DELETE /api/v1/status-pages/:id` - Delete status page
- `PUT /api/v1/status-pages/:id/config` - Update status page config

### Domain Management
- `POST /api/v1/domains` - Add domain
- `GET /api/v1/domains/:domain` - Get domain details
- `DELETE /api/v1/domains/:domain` - Delete domain
- `POST /api/v1/domains/:domain/verify` - Verify domain

### Settings & Configuration
- `GET /api/v1/settings` - Get tenant settings
- `PUT /api/v1/settings` - Update tenant settings

### Feature Flags
- `GET /api/v1/feature-flags` - List tenant feature flags
- `POST /api/v1/feature-flags` - Create tenant feature flag
- `GET /api/v1/feature-flags/:id` - Get tenant feature flag details
- `PUT /api/v1/feature-flags/:id` - Update tenant feature flag
- `DELETE /api/v1/feature-flags/:id` - Delete tenant feature flag

### Usage & Billing
- `GET /api/v1/usage` - Get tenant usage data
- `POST /api/v1/usage` - Record tenant usage

- `GET /api/v1/billing` - Get tenant billing information
- `PUT /api/v1/billing` - Update tenant billing information

### Notifications
- `GET /api/v1/notifications` - List tenant notifications
- `POST /api/v1/notifications` - Create tenant notification
- `GET /api/v1/notifications/:id` - Get notification details
- `PUT /api/v1/notifications/:id` - Update notification
- `DELETE /api/v1/notifications/:id` - Delete notification

### Activity & Backup Management
- `GET /api/v1/activities` - List tenant activities
- `GET /api/v1/stats` - Get tenant statistics

- `GET /api/v1/backups` - List tenant backups
- `POST /api/v1/backups` - Create tenant backup
- `GET /api/v1/backups/:id` - Get backup details
- `DELETE /api/v1/backups/:id` - Delete backup

### RBAC Management

#### Role Management
- `GET /api/v1/rbac/roles` - Get roles (requires `users.read` permission)
- `POST /api/v1/rbac/roles` - Create role (requires `users.create` permission)
- `GET /api/v1/rbac/roles/:id` - Get role details
- `PUT /api/v1/rbac/roles/:id` - Update role (requires `users.update` permission)
- `DELETE /api/v1/rbac/roles/:id` - Delete role (requires `users.delete` permission)
- `POST /api/v1/rbac/roles/:id/permissions` - Assign permissions to role (requires `users.update`)

#### Permission Management
- `GET /api/v1/rbac/permissions` - Get permissions (requires `users.read` permission)

#### User Role Management
- `POST /api/v1/rbac/user-roles` - Assign role to user (requires `users.update` permission)
- `DELETE /api/v1/rbac/user-roles/:user_id/:role_id` - Remove role from user (requires `users.update`)
- `GET /api/v1/rbac/user-roles/:user_id` - Get user roles
- `GET /api/v1/rbac/user-roles/:user_id/permissions` - Get user permissions
- `GET /api/v1/rbac/user-roles/:user_id/check` - Check user permission

#### Team Management
- `GET /api/v1/rbac/teams` - Get teams (requires `users.read` permission)
- `POST /api/v1/rbac/teams` - Create team (requires `users.create` permission)
- `GET /api/v1/rbac/teams/:id` - Get team details
- `PUT /api/v1/rbac/teams/:id` - Update team (requires `users.update` permission)
- `DELETE /api/v1/rbac/teams/:id` - Delete team (requires `users.delete` permission)

#### Team Member Management
- `POST /api/v1/rbac/team-members/:team_id` - Add user to team (requires `users.update` permission)
- `DELETE /api/v1/rbac/team-members/:team_id/:user_id` - Remove user from team (requires `users.update`)

#### Administrative Functions
- `POST /api/v1/rbac/admin/initialize-roles` - Initialize tenant roles (requires `admin` role)
- `GET /api/v1/rbac/admin/audit-logs` - Get audit logs (requires `admin` role)
- `GET /api/v1/rbac/admin/sessions` - Get active sessions (requires `admin` role)

### Web Interface
- `GET /` - Login page
- `GET /login` - Login page
- `GET /admin` - Admin dashboard
- `GET /static/*` - Static files

## Dependencies

### Internal Services
The service integrates with these other microservices:
- **component-service**: For status page component management
- **incident-service**: For status page incident management
- **monitoring-service**: For status page monitoring integration
- **notification-service**: For tenant notifications
- **branding-service**: For tenant branding and theming

### External Dependencies
- **shared-resilience**: Common resilience patterns, database management, middleware
- **PostgreSQL**: Primary database for tenant and RBAC data
- **Redis**: Caching and session management (optional)

### Database Models
The service manages these core models:
- Tenants, tenant branding, tenant settings
- Tenant admins and tenant feature flags
- Usage tracking and billing information
- Status pages and configurations
- Domains and domain verification records
- RBAC: Roles, permissions, user roles
- Teams and team memberships
- Sessions and audit logs

## Configuration

### Environment Variables

**Server Configuration**:
- `SERVER_PORT`: HTTP server port (default from shared-resilience config)
- `SERVER_HOST`: HTTP server host
- `ENVIRONMENT`: Runtime environment (development/production)

**Database Configuration**:
- `DB_HOST`: Database host
- `DB_PORT`: Database port
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name (tenant_admin_db)

**Authentication Configuration**:
- `JWT_SECRET`: JWT signing secret (min 32 characters)
- `REDIS_HOST`: Redis host (optional)
- `REDIS_PORT`: Redis port (optional)
- `BASE_DOMAIN`: Base domain for subdomain tenant routing (default: localhost)

**SAML/SSO Configuration** (NEW):
- `SAML_BASE_URL`: Base URL for SAML Service Provider (default: http://localhost:port)
  - Example: `https://statuspage.acme.com`
  - Used for constructing SAML metadata URLs and ACS endpoints

### Service Integration URLs
- `COMPONENT_SERVICE_URL`: Component service endpoint (default: http://localhost:8001)
- `INCIDENT_SERVICE_URL`: Incident service endpoint (default: http://localhost:8002)
- `MONITORING_SERVICE_URL`: Monitoring service endpoint (default: http://localhost:8003)
- `NOTIFICATION_SERVICE_URL`: Notification service endpoint (default: http://localhost:8004)
- `BRANDING_SERVICE_URL`: Branding service endpoint (default: http://localhost:8005)

## Development

### Running the Service
```bash
cd microservices/tenant-admin-service
go run cmd/main.go
```

### Building
```bash
go build -o tenant-admin-service cmd/main.go
```

### Testing
```bash
go test ./...
```

### Project Structure
```
tenant-admin-service/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── handlers/            # HTTP handlers
│   ├── middleware/          # RBAC and authentication middleware
│   ├── models/              # Data models and database schemas
│   └── services/            # Business service layer
├── web/
│   └── static/              # Static assets for admin pages
└── go.mod                   # Go module definition
```

## Architecture Notes

### Multi-Tenancy Design
- Complete tenant isolation at the database level
- Tenant identification via JWT tokens and middleware
- Tenant-specific resource scoping for all operations

### RBAC System Architecture
- **Hierarchical Permissions**: Fine-grained permission system
- **Role Inheritance**: Roles can inherit permissions from other roles
- **Team-based Access**: Team membership affects access rights
- **Session-based Authentication**: Secure session management
- **Audit Trail**: Complete logging of all RBAC operations

### Security Implementation
- JWT-based authentication with configurable expiration
- Middleware-based permission checking
- Session validation and management
- Multi-tenant data isolation
- Comprehensive audit logging

### SAML/SSO Architecture (NEW)
- **SAML 2.0 Protocol**: Full support for enterprise SSO
- **Multi-Provider Support**: Multiple SSO providers per tenant (Okta, Azure AD, Google Workspace, etc.)
- **Tenant-Scoped Providers**: Each tenant manages their own SSO configurations
- **SP-Initiated Flow**: Users initiate login from status page → redirect to IdP → SAML assertion → JWT token
- **IdP-Initiated Flow**: Users login at IdP → SAML assertion sent to status page → JWT token
- **JIT Provisioning**: Automatic user creation from SAML assertions with configurable attribute mapping
- **Session Management**: SAML sessions tracked for Single Logout (SLO) support
- **Audit Trail**: Complete logging of all SSO events (login success/failure, config changes, etc.)
- **Database Tables**:
  - `sso_providers` - SSO configurations (SAML/OAuth/OIDC)
  - `sso_user_identities` - Links users to IdP identities
  - `saml_requests` - Tracks pending SAML auth requests
  - `sso_audit_logs` - Complete audit trail

### Status Page Integration
- Direct integration with component, incident, and monitoring services
- Tenant-specific status page provisioning
- Custom domain support with DNS verification
- Branded status pages per tenant

### Middleware Stack
1. **Resilience Middleware**: From shared-resilience (logging, recovery, security)
2. **JWT Authentication**: Token validation and user identification
3. **Tenant Middleware**: Multi-tenant context establishment
4. **RBAC Middleware**: Permission-based access control
5. **Session Validation**: Session state verification
6. **Audit Logging**: Operation audit trail

### Database Design
- Multi-tenant aware schema design
- Foreign key relationships respect tenant boundaries
- Optimized queries with tenant scoping
- Automatic database migration system

This service is fundamental to the platform's multi-tenant architecture, providing the foundation for secure, isolated tenant operations while maintaining comprehensive administrative control and audit capabilities.