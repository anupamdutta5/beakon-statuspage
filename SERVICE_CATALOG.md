# Beakon Microservices - Service Catalog

**Last Updated**: 2025-10-14
**Total Services**: 19 active (+ 1 deprecated)
**Architecture**: Microservices with API Gateway
**Shared Library**: shared-resilience (100% adoption)

---

## Quick Reference Table

| Service | Port | Database | Status | Resilience | Primary Function |
|---------|------|----------|--------|------------|------------------|
| api-gateway | 8080 | None | ✅ Active | ✅ Full | Request routing & auth |
| user-service | 8081 | statuspage_user | ✅ Active | ✅ Full | User auth & management |
| component-service | 8084 | statuspage_component | ✅ Active | ✅ Full | Component monitoring |
| notification-service | 8085 | statuspage_notification | ✅ Active | ✅ Full | Multi-channel alerts |
| incident-service | 8086 | statuspage_incident | ✅ Active | ✅ Full | Incident management |
| payment-service | 8088 | statuspage_payment | ✅ Active | ✅ Full | Billing & payments |
| analytics-service | 8090 | statuspage_analytics | ✅ Active | ✅ Full | Analytics & reporting |
| monitoring-service | 8092 | statuspage_monitoring | ✅ Active | ✅ Full | System monitoring |
| status-ui-service | 8093 | None | ✅ Active | ✅ Full | Public status pages |
| database-service | 8095 | statuspage_database | ⚠️ DEPRECATED | ✅ Full | Database utils (unused) |
| event-store-service | 8096 | statuspage_event_store | ✅ Active | ✅ Full | Event sourcing |
| branding-service | 8097 | statuspage_branding | ✅ Active | ✅ Full | Theming & branding |
| saas-admin-service | 8098 | saas_admin | ✅ Active | ✅ Full | Platform admin |
| tenant-admin-service | 8099 | tenant_admin_db | ✅ Active | ✅ Full | Multi-tenant mgmt |
| landing-page-service | 8100 | statuspage_landing | ✅ Active | ✅ Full | Marketing website |
| analytics-consumer | N/A | statuspage_analytics_consumer | ✅ Active | ✅ Full | Analytics processing |
| notification-consumer | N/A | N/A | ✅ Active | ✅ Full | Notification delivery |
| audit-consumer | N/A | statuspage_audit_consumer | ✅ Active | ✅ Full | Audit log processing |
| billing-consumer | N/A | statuspage_billing_consumer | ✅ Active | ✅ Full | Billing calculations |
| shared-resilience | N/A | N/A | ✅ Active | N/A | Common patterns lib |

---

## Service Details

### 1. API Gateway
**Port**: 8080
**Database**: None (stateless routing)
**Type**: HTTP Service
**Language**: Go
**Repository**: [Submodule] microservices/api-gateway

**Purpose**: Central entry point for all client requests

**Features**:
- JWT authentication and validation
- Request routing to downstream services
- Rate limiting (per-IP, per-user, per-tenant)
- CORS handling
- Security headers (CSP, HSTS, X-Frame-Options)
- Circuit breaker implementation
- Request/response logging
- Prometheus metrics collection

**Key Dependencies**:
- shared-resilience (circuit breakers, rate limiting)
- All downstream services

**Health Endpoints**:
- `GET /health` - Overall health
- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness probe

**Configuration**:
- Environment: `ENVIRONMENT` (development/staging/production)
- JWT Secret: `JWT_SECRET`
- Rate Limits: Configurable per endpoint
- Downstream service URLs via config

---

### 2. User Service
**Port**: 8081
**Database**: statuspage_user
**Type**: HTTP Service
**Language**: Go
**Repository**: [Submodule] microservices/user-service

**Purpose**: User authentication and profile management

**Features**:
- User registration and login
- JWT token generation, validation, and refresh
- Password management (reset, change, hashing with bcrypt)
- User profile CRUD operations
- Multi-tenant user isolation
- Session management integration
- Email verification
- OAuth integration support

**API Endpoints**:
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh JWT token
- `POST /api/v1/auth/logout` - User logout
- `GET /api/v1/users/:id` - Get user profile
- `PUT /api/v1/users/:id` - Update user profile
- `POST /api/v1/users/:id/password` - Change password
- `POST /api/v1/auth/reset-password` - Request password reset

**Database Schema**:
- `users` - User accounts
- `refresh_tokens` - JWT refresh tokens
- `password_reset_tokens` - Password reset tokens

**Dependencies**:
- shared-resilience (database, auth middleware)
- Notification service (for email verification)

---

### 3. Component Service
**Port**: 8084
**Database**: statuspage_component
**Type**: HTTP Service
**Language**: Go

**Purpose**: Manages components displayed on status pages

**Features**:
- Component CRUD operations
- Component groups and hierarchy
- Component status tracking (operational, degraded, down, maintenance)
- Status history and timeline
- Component visibility management
- Public/private component designation
- Component metrics and alerts
- Integration with monitoring service

**API Endpoints**:
- `POST /api/v1/components` - Create component
- `GET /api/v1/components` - List components (tenant-scoped)
- `GET /api/v1/components/:id` - Get component details
- `PUT /api/v1/components/:id` - Update component
- `DELETE /api/v1/components/:id` - Delete component
- `GET /api/v1/components/:id/status` - Get current status
- `POST /api/v1/components/:id/status` - Update status
- `GET /api/v1/components/:id/history` - Status history
- `GET /api/v1/public/components` - Public component list

**Database Schema**:
- `components` - Component definitions
- `component_groups` - Component grouping
- `component_status_history` - Status change log
- `component_metrics` - Performance metrics

---

### 4. Notification Service
**Port**: 8085
**Database**: statuspage_notification
**Type**: HTTP Service
**Language**: Go

**Purpose**: Multi-channel notification delivery system

**Features**:
- Email notifications (SMTP, SendGrid, AWS SES)
- SMS notifications (Twilio, AWS SNS)
- Push notifications (FCM, APNS)
- Webhook delivery with retries
- Notification templates and customization
- Subscriber management
- Notification preferences and opt-out
- Delivery tracking and status
- Rate limiting to prevent spam
- Queue-based processing

**API Endpoints**:
- `POST /api/v1/notifications/send` - Send notification
- `GET /api/v1/notifications/:id` - Get notification status
- `POST /api/v1/notifications/subscribe` - Add subscriber
- `DELETE /api/v1/notifications/subscribe/:id` - Unsubscribe
- `GET /api/v1/notifications/subscribers` - List subscribers
- `POST /api/v1/notifications/test` - Test notification

**Database Schema**:
- `notifications` - Notification log
- `subscribers` - Subscription management
- `notification_templates` - Message templates
- `notification_preferences` - User preferences
- `webhooks` - Webhook configurations

**Integration Partners**:
- SendGrid / AWS SES (Email)
- Twilio / AWS SNS (SMS)
- Firebase Cloud Messaging (Push)

---

### 5. Incident Service
**Port**: 8086
**Database**: statuspage_incident
**Type**: HTTP Service
**Language**: Go

**Purpose**: Complete incident lifecycle management

**Features**:
- Incident creation, updates, and resolution
- Incident status workflow (investigating, identified, monitoring, resolved)
- Incident-component associations
- Incident updates and timeline
- Incident templates for common scenarios
- Scheduled maintenance management
- Public incident display on status pages
- Incident metrics (MTTR, MTBF)
- Automatic notification triggering
- Post-mortem management

**API Endpoints**:
- `POST /api/v1/incidents` - Create incident
- `GET /api/v1/incidents` - List incidents
- `GET /api/v1/incidents/:id` - Get incident details
- `PUT /api/v1/incidents/:id` - Update incident
- `POST /api/v1/incidents/:id/updates` - Add incident update
- `POST /api/v1/incidents/:id/resolve` - Resolve incident
- `GET /api/v1/incidents/:id/timeline` - Incident timeline
- `GET /api/v1/public/incidents` - Public incidents
- `POST /api/v1/maintenance` - Schedule maintenance

**Database Schema**:
- `incidents` - Incident records
- `incident_updates` - Status updates
- `incident_components` - Component associations
- `maintenance_windows` - Scheduled maintenance
- `incident_templates` - Reusable templates

---

### 6. Payment Service
**Port**: 8088
**Database**: statuspage_payment
**Type**: HTTP Service
**Language**: Go

**Purpose**: Billing and payment processing

**Features**:
- Payment processing (Stripe, PayPal integration)
- Subscription management
- Invoice generation and tracking
- Payment method management (credit cards, ACH)
- Billing cycle management
- Proration calculations
- Revenue reporting
- Failed payment handling and retries
- Tax calculation integration
- PCI compliance support

**API Endpoints**:
- `POST /api/v1/payments/process` - Process payment
- `POST /api/v1/subscriptions` - Create subscription
- `GET /api/v1/subscriptions/:id` - Get subscription
- `PUT /api/v1/subscriptions/:id` - Update subscription
- `DELETE /api/v1/subscriptions/:id` - Cancel subscription
- `GET /api/v1/invoices` - List invoices
- `GET /api/v1/invoices/:id` - Get invoice
- `POST /api/v1/payment-methods` - Add payment method
- `GET /api/v1/payment-methods` - List payment methods

**Database Schema**:
- `payments` - Payment transactions
- `subscriptions` - Subscription records
- `invoices` - Invoice tracking
- `payment_methods` - Stored payment methods
- `billing_cycles` - Billing period tracking

---

### 7. Analytics Service
**Port**: 8090
**Database**: statuspage_analytics
**Type**: HTTP Service
**Language**: Go

**Purpose**: Platform analytics and business intelligence

**Features**:
- Analytics data collection and aggregation
- Uptime calculations and SLA tracking
- Performance metrics analysis
- Component availability reporting
- Custom dashboards and visualizations
- Real-time analytics streaming
- Historical data analysis
- Export functionality (CSV, JSON, PDF)
- Scheduled reports
- Business intelligence queries

**API Endpoints**:
- `POST /api/v1/analytics/track` - Track event
- `GET /api/v1/analytics/uptime/:component_id` - Get uptime stats
- `GET /api/v1/analytics/dashboard` - Dashboard data
- `GET /api/v1/analytics/reports` - List reports
- `POST /api/v1/analytics/reports` - Generate report
- `GET /api/v1/analytics/sla` - SLA metrics
- `GET /api/v1/analytics/export` - Export data

**Database Schema**:
- `analytics_events` - Event tracking
- `uptime_records` - Uptime calculations
- `sla_definitions` - SLA configurations
- `reports` - Generated reports
- `dashboards` - Custom dashboards

---

### 8. Monitoring Service
**Port**: 8092
**Database**: statuspage_monitoring
**Type**: HTTP Service
**Language**: Go

**Purpose**: Comprehensive system monitoring and health checks

**Features**:
- HTTP endpoint monitoring
- TCP/UDP port monitoring
- SSL certificate monitoring
- DNS monitoring
- API health checks
- Custom script execution
- Container/pod monitoring (Docker/Kubernetes)
- Alert generation and routing
- Maintenance window management
- Monitoring schedules and intervals
- Retry logic for failed checks
- Integration with incident service

**API Endpoints**:
- `POST /api/v1/monitors` - Create monitor
- `GET /api/v1/monitors` - List monitors
- `GET /api/v1/monitors/:id` - Get monitor
- `PUT /api/v1/monitors/:id` - Update monitor
- `DELETE /api/v1/monitors/:id` - Delete monitor
- `POST /api/v1/monitors/:id/test` - Test monitor
- `GET /api/v1/monitors/:id/results` - Monitor results
- `POST /api/v1/alerts` - Create alert rule
- `GET /api/v1/alerts` - List alerts

**Database Schema**:
- `monitors` - Monitor configurations
- `monitor_results` - Check results
- `alert_rules` - Alerting configuration
- `maintenance_windows` - Maintenance schedules

---

### 9. Status UI Service
**Port**: 8093
**Database**: None (stateless frontend)
**Type**: HTTP Service (Frontend)
**Language**: Go (backend) + HTML/CSS/JS
**Repository**: [Submodule] microservices/status-ui-service

**Purpose**: Public-facing status page interface

**Features**:
- Customizable status page rendering
- Real-time status updates via WebSocket
- Mobile-responsive design
- Custom branding and theming
- Historical incident display
- Component grouping and filtering
- RSS feed generation
- Status badge generation
- Public API for status data
- SEO optimization

**Routes**:
- `GET /` - Main status page
- `GET /:tenant_slug` - Tenant-specific status page
- `GET /incidents` - Incident history
- `GET /incidents/:id` - Incident details
- `GET /subscribe` - Subscription management
- `GET /history` - Historical uptime
- `GET /api/v1/status.json` - JSON status API
- `GET /badge.svg` - Status badge

**Integration**:
- Component Service (status data)
- Incident Service (incident data)
- Branding Service (theme data)
- Notification Service (subscriptions)

---

### 10. Database Service ⚠️ DEPRECATED
**Port**: 8095
**Database**: statuspage_database
**Type**: HTTP Service
**Status**: **DEPRECATED - NOT IN USE**

**Purpose**: Originally designed for centralized database operations

**Note**: This service has been superseded by the shared-resilience library. Each microservice now manages its own database connections using standardized patterns from shared-resilience. This service is no longer actively used or maintained.

**Recommendation**: Archive this service or remove from active deployment.

---

### 11. Event Store Service
**Port**: 8096
**Database**: statuspage_event_store
**Type**: HTTP Service
**Language**: Go

**Purpose**: Event sourcing and CQRS implementation

**Features**:
- Event persistence and storage
- Event stream management
- Event replay capabilities
- Aggregate root management
- Snapshot creation and restoration
- Event versioning
- Event projection to read models
- Idempotency handling
- Event ordering guarantees

**API Endpoints**:
- `POST /api/v1/events` - Append event
- `GET /api/v1/events/:stream_id` - Get event stream
- `GET /api/v1/events/:stream_id/from/:position` - Read from position
- `POST /api/v1/snapshots` - Create snapshot
- `GET /api/v1/snapshots/:aggregate_id` - Get latest snapshot

**Database Schema**:
- `events` - Event log
- `snapshots` - Aggregate snapshots
- `event_metadata` - Event metadata

---

### 12. Branding Service
**Port**: 8097
**Database**: statuspage_branding
**Type**: HTTP Service
**Language**: Go

**Purpose**: Custom theming and white-labeling

**Features**:
- Custom theme management (colors, fonts, layouts)
- Logo upload and management
- CSS customization
- White-label configuration
- Domain-specific branding
- Theme preview and testing
- Brand asset management
- Theme versioning
- Mobile theme support

**API Endpoints**:
- `POST /api/v1/branding` - Create branding
- `GET /api/v1/branding/:tenant_id` - Get tenant branding
- `PUT /api/v1/branding/:tenant_id` - Update branding
- `POST /api/v1/branding/:tenant_id/logo` - Upload logo
- `GET /api/v1/branding/:tenant_id/theme.css` - Get theme CSS
- `POST /api/v1/branding/:tenant_id/preview` - Preview theme

**Database Schema**:
- `branding` - Branding configurations
- `themes` - Theme definitions
- `assets` - Brand assets (logos, images)

---

### 13. SaaS Admin Service
**Port**: 8098
**Database**: saas_admin
**Type**: HTTP Service + Web UI
**Language**: Go

**Purpose**: Platform-wide administration and SaaS operations

**Features**:
- Platform administration dashboard (web UI)
- Subscription plan management
- Feature management and feature flags
- Pricing structure configuration
- Platform-level statistics and analytics
- Tenant provisioning and management
- Admin user management
- Billing integration oversight
- System-wide settings
- Backup and restore management
- Notification template management

**Web Routes**:
- `GET /` - Admin login page
- `GET /dashboard` - Main admin dashboard
- `GET /tenants` - Tenant management
- `GET /plans` - Subscription plans
- `GET /features` - Feature management
- `GET /pricing` - Pricing configuration
- `GET /users` - Platform users
- `GET /settings` - System settings

**API Endpoints**:
- `POST /api/v1/auth/login` - Admin login
- `GET /api/v1/dashboard/stats` - Dashboard statistics
- `POST /api/v1/plans` - Create subscription plan
- `GET /api/v1/plans` - List plans
- `PUT /api/v1/plans/:id` - Update plan
- `POST /api/v1/features` - Create feature
- `GET /api/v1/features` - List features
- `POST /api/v1/tenants` - Create tenant
- `GET /api/v1/tenants` - List all tenants
- `GET /api/v1/tenants/:id` - Get tenant details
- `PUT /api/v1/tenants/:id` - Update tenant
- `DELETE /api/v1/tenants/:id` - Delete tenant

**Database Tables** (in saas_admin):
- `saas_admin_users` - Platform admin users
- `subscription_plans` - Plan definitions
- `features` - Feature catalog
- `pricing` - Pricing rules
- `platform_settings` - System configuration
- `sessions` - Admin session management

**Service Communication**:
- Calls tenant-admin-service API for tenant CRUD operations
- Communicates via HTTP (pure microservices pattern)

---

### 14. Tenant Admin Service
**Port**: 8099
**Database**: tenant_admin_db
**Type**: HTTP Service + Web UI
**Language**: Go

**Purpose**: Multi-tenant management and RBAC

**Features**:
- Tenant CRUD operations
- Role-Based Access Control (RBAC) system
- User management with max users enforcement
- Team management
- Session management (Redis + DB + in-memory fallback)
- Status page configuration
- Domain management and verification
- Audit logging
- Feature flag management per tenant
- Usage tracking and statistics
- Backup management
- Notification settings

**Web Routes**:
- `GET /` - Tenant admin login
- `GET /login` - Login page
- `GET /admin` - Admin dashboard

**API Endpoints**:
- `POST /api/v1/auth/login` - Tenant admin login
- `POST /api/v1/auth/logout` - Logout
- `POST /api/v1/auth/verify` - Verify token
- `POST /api/v1/public/tenants` - Create tenant
- `GET /api/v1/public/tenants` - List tenants (public)
- `GET /api/v1/tenants` - List tenants (authenticated)
- `GET /api/v1/tenants/:id` - Get tenant
- `PUT /api/v1/tenants/:id` - Update tenant
- `DELETE /api/v1/tenants/:id` - Delete tenant
- `GET /api/v1/admins` - List tenant admins
- `POST /api/v1/admins` - Create admin
- `GET /api/v1/users/:tenant_id` - List users
- `POST /api/v1/users/:tenant_id` - Create user
- `GET /api/v1/users/:tenant_id/stats` - User statistics
- `POST /api/v1/rbac/roles` - Create role
- `GET /api/v1/rbac/roles` - List roles
- `POST /api/v1/rbac/user-roles` - Assign role to user
- `GET /api/v1/rbac/teams` - List teams
- `POST /api/v1/rbac/teams` - Create team

**Database Tables** (in tenant_admin_db):
- `tenants` - Tenant records (with max_users)
- `users` - Tenant users
- `roles` - RBAC roles
- `permissions` - RBAC permissions
- `user_roles` - Role assignments
- `teams` - Team definitions
- `team_members` - Team membership
- `sessions` - Session tracking
- `domains` - Custom domain management
- `audit_logs` - Audit trail

**Key Features**:
1. **Max Users Enforcement**: Validates tenant user limit before user creation
2. **Three-Tier Session Management**:
   - Primary: Redis
   - Fallback: Database
   - Emergency: In-memory cache
3. **Complete RBAC**: Roles, permissions, teams, user assignments
4. **Audit Logging**: Comprehensive activity tracking

---

### 15. Landing Page Service
**Port**: 8100
**Database**: statuspage_landing
**Type**: HTTP Service (Marketing Website)
**Language**: Go

**Purpose**: Public marketing website and lead generation

**Features**:
- Marketing website content
- Product information pages
- Pricing page display
- Feature comparison
- Customer testimonials
- Blog/news section
- Lead capture forms
- Contact forms
- Demo request handling
- SEO optimization
- Analytics integration

**Routes**:
- `GET /` - Homepage
- `GET /pricing` - Pricing page
- `GET /features` - Features overview
- `GET /about` - About us
- `GET /contact` - Contact form
- `GET /demo` - Request demo
- `GET /blog` - Blog listing
- `GET /blog/:slug` - Blog post

**Database Schema**:
- `landing_pages` - Page content
- `leads` - Lead captures
- `blog_posts` - Blog content
- `testimonials` - Customer testimonials

---

### 16. Analytics Consumer
**Port**: N/A (Background Processor)
**Database**: statuspage_analytics_consumer
**Type**: Event Consumer
**Language**: Go

**Purpose**: Asynchronous analytics event processing

**Features**:
- Event-driven analytics data processing
- Data aggregation and rollup
- Metric calculations
- Report generation scheduling
- Data cleanup and archiving
- Event batching for performance
- Worker pool for parallel processing

**Event Sources**:
- Component status changes
- Incident lifecycle events
- User activity events
- API usage events

**Database Schema**:
- `processed_events` - Event processing log
- `aggregations` - Pre-calculated metrics
- `scheduled_reports` - Report queue

---

### 17. Notification Consumer
**Port**: N/A (Background Processor)
**Database**: None (processes from queue)
**Type**: Event Consumer
**Language**: Go

**Purpose**: Asynchronous notification delivery

**Features**:
- Queue-based notification processing
- Retry logic for failed deliveries
- Delivery status tracking
- Rate limiting per channel
- Batch processing for efficiency
- Dead letter queue for failed notifications

**Event Sources**:
- Incident updates
- Component status changes
- System alerts
- User-triggered notifications

**Processing**:
- Email delivery
- SMS delivery
- Push notification delivery
- Webhook delivery

---

### 18. Audit Consumer
**Port**: N/A (Background Processor)
**Database**: statuspage_audit_consumer
**Type**: Event Consumer
**Language**: Go

**Purpose**: Asynchronous audit log processing

**Features**:
- Audit event collection
- Log aggregation and storage
- Compliance reporting
- Retention policy enforcement
- Audit trail generation
- Security event detection

**Event Sources**:
- User actions (login, CRUD operations)
- Administrative actions
- Security events
- Configuration changes

**Database Schema**:
- `audit_events` - Audit log
- `security_events` - Security-specific events
- `compliance_reports` - Generated reports

---

### 19. Billing Consumer
**Port**: N/A (Background Processor)
**Database**: statuspage_billing_consumer
**Type**: Event Consumer
**Language**: Go

**Purpose**: Asynchronous billing calculations and processing

**Features**:
- Usage-based billing calculations
- Subscription renewal processing
- Invoice generation
- Payment retry handling
- Proration calculations
- Revenue recognition
- Billing notification triggering

**Event Sources**:
- User activity (usage metrics)
- Subscription lifecycle events
- Payment events
- Plan changes

**Database Schema**:
- `billing_events` - Billing event log
- `usage_records` - Usage tracking
- `billing_cycles` - Billing period tracking

---

### 20. Shared Resilience
**Port**: N/A (Go Module Library)
**Database**: N/A
**Type**: Shared Library
**Language**: Go
**Package**: `github.com/anupamdutta5/shared-resilience`

**Purpose**: Common resilience patterns and utilities for all services

**Adoption**: 100% (all 20 services use it)

**Features**:

1. **Circuit Breakers** (Sony's gobreaker)
   - Database connection protection
   - External service call protection
   - Configurable failure thresholds
   - Automatic recovery

2. **Database Management**
   - Connection pooling (CPU-based sizing)
   - Auto-migration with GORM
   - Health checks
   - Graceful connection handling

3. **Caching**
   - In-memory cache with TTL
   - Redis cache with fallback
   - Cache invalidation

4. **Rate Limiting**
   - Per-IP rate limiting
   - Per-user rate limiting
   - Per-tenant rate limiting
   - Global rate limiting
   - Configurable burst sizes

5. **Health Checks**
   - Liveness probes (`/health/live`)
   - Readiness probes (`/health/ready`)
   - Custom health checks
   - Component health tracking
   - Concurrent check execution

6. **Middleware**
   - JWT authentication
   - Tenant context injection
   - Request logging
   - Security headers (CSP, HSTS, X-Frame-Options)
   - CORS handling
   - Input sanitization
   - Panic recovery

7. **Retry Logic**
   - Exponential backoff
   - Configurable retry attempts
   - Max delay capping
   - Context-aware retries

8. **Configuration**
   - Environment variable loading
   - YAML configuration support
   - Environment-specific configs
   - Validation

9. **Error Handling**
   - Standardized error types
   - Error wrapping and context
   - HTTP error responses

10. **Utilities**
    - Graceful shutdown
    - Signal handling
    - Logging setup (zap)
    - Correlation ID generation

**Files**:
```
shared-resilience/
├── cache.go                # Caching interfaces
├── circuit_breaker.go      # Circuit breaker implementation
├── config.go               # Configuration management
├── database.go             # Database connection management
├── errors.go               # Error handling utilities
├── health.go               # Health check implementation
├── middleware.go           # Common middleware
├── rate_limiter.go         # Rate limiting
├── redis_cache.go          # Redis cache implementation
├── retry.go                # Retry logic
├── security.go             # Security utilities
├── validation.go           # Input validation
└── README.md               # Documentation
```

---

## Database Architecture

### Database-per-Service Pattern

Beakon follows the microservices best practice of **database-per-service** (pure microservices pattern):

| Database Name | Service | Purpose |
|---------------|---------|---------|
| saas_admin | saas-admin-service | Platform admin, plans, features |
| tenant_admin_db | tenant-admin-service | Tenants, users, RBAC |
| statuspage_user | user-service | User authentication |
| statuspage_component | component-service | Components & status |
| statuspage_notification | notification-service | Notifications |
| statuspage_incident | incident-service | Incidents |
| statuspage_payment | payment-service | Payments & billing |
| statuspage_analytics | analytics-service | Analytics |
| statuspage_monitoring | monitoring-service | Monitoring |
| statuspage_event_store | event-store-service | Event sourcing |
| statuspage_branding | branding-service | Theming |
| statuspage_landing | landing-page-service | Marketing content |
| statuspage_analytics_consumer | analytics-consumer | Analytics processing |
| statuspage_audit_consumer | audit-consumer | Audit logs |
| statuspage_billing_consumer | billing-consumer | Billing events |
| statuspage_database | database-service | ⚠️ **DEPRECATED** |

### Service Communication

Services communicate via HTTP APIs, not database sharing:
- **SaaS Admin → Tenant Admin**: Creates tenants via `POST /api/v1/tenants` API
- **All Services → API Gateway**: Standard request routing pattern
- **Consumer Services**: Event-driven processing via message queues

---

## Service Communication Patterns

### Primary Pattern: API Gateway

**All client requests** should go through the API Gateway (port 8080):

```
Client → API Gateway (8080) → Downstream Service
```

### Inter-Service Communication

**Standard Pattern** (recommended):
```
Service A → API Gateway (8080) → Service B
```

**Known Exception**:
```
SaaS Admin Service → Tenant Admin Service (direct call)
```
- Used for: Session validation
- Reason: Reduced latency for admin operations
- Status: Documented exception

### Event-Driven Communication

Consumer services process events asynchronously:
```
Service → Event Queue → Consumer Service → Database
```

---

## Monitoring & Observability

### Health Check Endpoints

All HTTP services implement:
- `GET /health` - Complete health status
- `GET /health/live` - Kubernetes liveness probe
- `GET /health/ready` - Kubernetes readiness probe

### Metrics

All services expose Prometheus metrics on port +1010:
- Service on 8080 → Metrics on 9090
- Service on 8081 → Metrics on 9091
- etc.

**Collected Metrics**:
- Request throughput (requests/sec)
- Response latency (p50, p95, p99)
- Error rates (4xx, 5xx)
- Circuit breaker state
- Database connection pool usage
- Cache hit/miss rates
- Event processing times

### Logging

**Format**: Structured JSON logging with zap

**Fields**:
- Timestamp
- Service name
- Log level
- Correlation ID
- Tenant ID (if applicable)
- User ID (if applicable)
- Message
- Stack trace (for errors)

---

## Security

### Authentication & Authorization

1. **JWT-Based Authentication**
   - Issued by user-service
   - Validated by API Gateway
   - Includes tenant context

2. **Role-Based Access Control (RBAC)**
   - Managed by tenant-admin-service
   - Roles: owner, admin, manager, viewer
   - Granular permissions

3. **Session Management**
   - Three-tier: Redis → Database → In-memory
   - 24-hour expiration
   - HttpOnly cookies
   - CSRF protection

### Security Measures

- Input sanitization (XSS prevention)
- SQL injection prevention (parameterized queries)
- Rate limiting (per-IP, per-user, per-tenant)
- Security headers (CSP, HSTS, X-Frame-Options)
- Bcrypt password hashing (cost factor 10)
- TLS/SSL for all communication

---

## Deployment

### Development Ports Summary

```
8080  API Gateway
8081  User Service
8084  Component Service
8085  Notification Service
8086  Incident Service
8088  Payment Service
8090  Analytics Service
8092  Monitoring Service
8093  Status UI Service
8095  Database Service (DEPRECATED)
8096  Event Store Service
8097  Branding Service
8098  SaaS Admin Service
8099  Tenant Admin Service
8100  Landing Page Service
```

### Container Deployment

All services are containerized with Docker and can be orchestrated with Kubernetes.

**Required Environment Variables** (per service):
- `ENVIRONMENT` - development/staging/production
- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`
- `JWT_SECRET` (min 32 chars)
- `REDIS_HOST`, `REDIS_PORT` (optional)
- Service-specific variables

---

## Quick Start Guide

### Starting All Services Locally

1. **Prerequisites**:
   ```bash
   # Install PostgreSQL
   brew install postgresql

   # Install Redis (optional)
   brew install redis

   # Install Go 1.21+
   brew install go
   ```

2. **Setup Databases**:
   ```bash
   psql -U postgres -c "CREATE DATABASE tenant_admin_db;"
   # Create other databases as needed
   ```

3. **Start Services**:
   ```bash
   cd microservices/api-gateway && go run cmd/main.go &
   cd microservices/user-service && go run cmd/main.go &
   cd microservices/tenant-admin-service && go run cmd/main.go &
   cd microservices/saas-admin-service && go run cmd/main.go &
   # ... continue for other services
   ```

4. **Verify Health**:
   ```bash
   curl http://localhost:8080/health
   ```

---

## Troubleshooting

### Common Issues

1. **Port Already in Use**:
   ```bash
   lsof -i :<port>
   kill -9 <PID>
   ```

2. **Database Connection Failed**:
   - Check PostgreSQL is running: `pg_isready`
   - Verify connection string in config
   - Check database exists: `psql -U postgres -l`

3. **Redis Connection Failed**:
   - System gracefully falls back to in-memory cache
   - To fix: Start Redis with `redis-server`

4. **Service Not Starting**:
   - Check logs for error messages
   - Verify all environment variables are set
   - Ensure no port conflicts

---

## Contributing

### Adding a New Service

1. Create service directory under `microservices/`
2. Implement using `shared-resilience` patterns
3. Add health check endpoints
4. Create configuration files (development.yaml, production.yaml)
5. Add service to this catalog
6. Update ARCHITECTURE.md
7. Document API endpoints

### Service Development Checklist

- [ ] Uses shared-resilience for database connections
- [ ] Implements circuit breakers
- [ ] Includes health check endpoints
- [ ] Exposes Prometheus metrics
- [ ] Uses structured logging (zap)
- [ ] Implements graceful shutdown
- [ ] Includes configuration management
- [ ] Multi-tenant aware (if applicable)
- [ ] Comprehensive error handling
- [ ] API documentation
- [ ] README.md with setup instructions

---

**Document Version**: 1.0
**Last Audit**: 2025-10-14
**Maintained By**: Engineering Team
**Review Frequency**: Monthly or after major changes

This is a **living document** - update it whenever services are added, modified, or removed.
