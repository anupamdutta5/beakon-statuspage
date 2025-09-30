# Beakon Statuspage Platform - Microservices Architecture Reference

*Comprehensive analysis of all microservices, their purposes, and architectural relationships*

---

## 🏗️ **Architecture Overview**

**Total Services**: 20 microservices + 4 consumers + 1 API gateway
**Architecture Pattern**: Multi-tenant SaaS with microservices
**Communication**: REST APIs via API Gateway + Event-driven consumers
**Database**: PostgreSQL with tenant isolation

---

## 🎯 **Service Categories**

### **1. Platform Administration**
| Service | Port | Purpose |
|---------|------|---------|
| **saas-admin-service** | 8098 | Platform-wide super admin management |
| **api-gateway** | 8080 | Request routing and API aggregation |

### **2. Tenant Management**
| Service | Port | Purpose |
|---------|------|---------|
| **tenant-admin-service** | 8099 | Individual tenant admin dashboards |
| **user-service** | 8081 | User authentication and management |

### **3. Core Status Page Features**
| Service | Port | Purpose |
|---------|------|---------|
| **component-service** | 8084 | Status page components management |
| **incident-service** | 8086 | Incident reporting and management |
| **monitoring-service** | 8092 | Uptime monitoring and health checks |
| **status-ui-service** | 8093 | Public status page rendering |

### **4. Business Services**
| Service | Port | Purpose |
|---------|------|---------|
| **analytics-service** | 8090 | Usage analytics and reporting |
| **notification-service** | 8085 | Alerts and subscriber notifications |
| **payment-service** | 8088 | Billing and subscription management |
| **branding-service** | 8097 | Tenant branding and customization |

### **5. Infrastructure Services**
| Service | Port | Purpose |
|---------|------|---------|
| **database-service** | 8095 | Database operations and migrations |
| **event-store-service** | 8096 | Event sourcing and audit trail |
| **landing-page-service** | 8100 | Public marketing pages |

### **6. Event Consumers**
| Service | Port | Purpose |
|---------|------|---------|
| **analytics-consumer** | - | Analytics data processing |
| **audit-consumer** | - | Audit log processing |
| **billing-consumer** | - | Billing event processing |
| **notification-consumer** | - | Notification delivery |

---

## 🏢 **Tenant Architecture Deep Dive**

### **SaaS Admin vs Tenant Admin**

#### **🔧 SaaS Admin Service (Port 8098)**
**Role**: Platform Super Administration
**Users**: Beakon platform administrators
**Scope**: Entire platform management

**Key Responsibilities**:
- **Tenant Management**: Create, suspend, delete tenant accounts
- **Plan Management**: Manage pricing plans and feature sets
- **Platform Analytics**: Cross-tenant insights and metrics
- **Feature Flags**: Global feature rollouts and A/B testing
- **Billing Management**: Platform-wide billing and revenue tracking
- **Admin User Management**: Platform admin accounts
- **System Monitoring**: Platform health and performance
- **Integration Management**: Third-party platform integrations

**Database**: `saas_admin` - Platform-level data

#### **👥 Tenant Admin Service (Port 8099)**
**Role**: Individual Customer Administration
**Users**: Customer's own administrators
**Scope**: Single tenant operations

**Key Responsibilities**:
- **User Management**: Tenant-specific user accounts and roles
- **Status Page Configuration**: Tenant's status page settings
- **Component Management**: Tenant's components and services
- **Incident Management**: Tenant's incident response
- **Monitoring Setup**: Tenant's monitoring configuration
- **Branding Customization**: Tenant's visual identity
- **Subscriber Management**: Tenant's notification lists
- **Analytics Dashboard**: Tenant-specific insights
- **Billing Overview**: Tenant's usage and billing info

**Database**: Tenant-specific tables with `tenant_id` isolation

### **Customer Journey Flow**

```
1. [SaaS Admin] → Creates new tenant account
2. [System] → Provisions tenant-specific resources
3. [Tenant Admin] → Configures their status page
4. [Tenant Users] → Manage daily operations
5. [End Users] → View public status page
```

---

## 📋 **Detailed Service Analysis**

### **🌐 API Gateway (Port 8080)**
**Purpose**: Central entry point for all client requests
**Technology**: Go + Gin framework
**Key Features**:
- Request routing to appropriate microservices
- Authentication and authorization
- Rate limiting and throttling
- Request/response transformation
- Load balancing and failover

**Critical Issues Found**:
- Port mismatches with actual services
- Service discovery hardcoded
- Needs dynamic routing configuration

**API Routes**:
```
/api/v1/auth/* → user-service:8081
/api/v1/components/* → component-service:8084 ⚠️(expects 8083)
/api/v1/incidents/* → incident-service:8086 ⚠️(expects 8084)
/api/v1/monitoring/* → monitoring-service:8092 ⚠️(expects 8088)
```

---

### **🔐 User Service (Port 8081)**
**Purpose**: Authentication and user management across platform
**Database**: `user_service` table space
**Multi-tenancy**: Users belong to tenants via `tenant_id`

**Key Models**:
- `User`: User accounts with tenant association
- `Role`: Role-based access control
- `Permission`: Granular permissions
- `UserSession`: Session management
- `APIKey`: API authentication tokens

**API Endpoints**:
- `POST /auth/login` - User authentication
- `POST /auth/register` - User registration
- `GET /users` - List users (tenant-scoped)
- `POST /users` - Create user
- `PUT /users/{id}` - Update user
- `DELETE /users/{id}` - Delete user

**Tenant Integration**: Every user operation is scoped by `tenant_id`

---

### **🏢 SaaS Admin Service (Port 8098) - DETAILED**
**Purpose**: Platform administration and tenant lifecycle management
**Database**: `saas_admin`
**Authentication**: Super admin JWT tokens

**Core Models**:
```go
type Tenant struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Domain      string    `json:"domain"`
    PlanID      string    `json:"plan_id"`
    Status      string    `json:"status"`      // active, suspended, cancelled
    CreatedAt   time.Time `json:"created_at"`
    Settings    JSONB     `json:"settings"`
}

type Plan struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Price       float64   `json:"price"`
    Features    []string  `json:"features"`
    Limits      Limits    `json:"limits"`
}

type PlatformStats struct {
    TotalTenants     int `json:"total_tenants"`
    ActiveTenants    int `json:"active_tenants"`
    TotalRevenue     float64 `json:"total_revenue"`
    TotalIncidents   int `json:"total_incidents"`
}
```

**API Endpoints**:
- `GET /api/v1/platform` - Platform overview
- `GET /api/v1/plans` - Manage pricing plans
- `GET /api/v1/stats` - Platform analytics
- `GET /api/v1/admin-users` - Platform admin management
- `GET /api/v1/features` - Feature flag management
- `GET /api/v1/activities` - Platform audit log

---

### **👤 Tenant Admin Service (Port 8099) - DETAILED**
**Purpose**: Individual tenant administration dashboard
**Database**: Tenant-scoped data access
**Authentication**: Tenant admin JWT tokens

**Core Models**:
```go
type TenantConfig struct {
    TenantID      string    `json:"tenant_id"`
    StatusPage    StatusPageConfig `json:"status_page"`
    Branding      BrandingConfig   `json:"branding"`
    Notifications NotificationConfig `json:"notifications"`
    Users         []TenantUser     `json:"users"`
}

type TenantUser struct {
    ID       string `json:"id"`
    Email    string `json:"email"`
    Role     string `json:"role"`    // owner, admin, viewer
    TenantID string `json:"tenant_id"`
}

type TenantStats struct {
    Components      int `json:"components"`
    ActiveMonitors  int `json:"active_monitors"`
    Incidents       int `json:"incidents"`
    Subscribers     int `json:"subscribers"`
    Uptime          float64 `json:"uptime"`
}
```

**API Endpoints**:
- `GET /api/v1/dashboard` - Tenant dashboard data
- `GET /api/v1/users` - Tenant user management
- `PUT /api/v1/settings` - Tenant configuration
- `GET /api/v1/stats` - Tenant analytics
- `GET /api/v1/billing` - Tenant billing info

---

### **🔧 Component Service (Port 8084)**
**Purpose**: Status page component management
**Database**: `component_service` with tenant isolation
**Multi-tenancy**: All operations scoped by `tenant_id`

**Key Models**:
```go
type Component struct {
    ID          string    `json:"id"`
    TenantID    string    `json:"tenant_id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Status      ComponentStatus `json:"status"`
    GroupID     string    `json:"group_id"`
    Position    int       `json:"position"`
    ShowUptime  bool      `json:"show_uptime"`
}

type ComponentGroup struct {
    ID       string `json:"id"`
    TenantID string `json:"tenant_id"`
    Name     string `json:"name"`
    Position int    `json:"position"`
}

type ComponentStatus string // operational, degraded, partial_outage, major_outage, maintenance
```

**API Endpoints**:
- `GET /components` - List tenant components
- `POST /components` - Create component
- `PUT /components/{id}/status` - Update component status
- `GET /component-groups` - List component groups
- `POST /component-groups` - Create component group

**Integration Points**:
- **Monitoring Service**: Receives status updates from monitors
- **Incident Service**: Component status affects incident severity
- **Status UI Service**: Renders public component status

---

### **📊 Monitoring Service (Port 8092)**
**Purpose**: Uptime monitoring and health checks
**Database**: `monitoring_service` with tenant isolation
**Multi-tenancy**: Monitors belong to tenants

**Key Models**:
```go
type Monitor struct {
    ID          string        `json:"id"`
    TenantID    string        `json:"tenant_id"`
    Name        string        `json:"name"`
    Type        MonitorType   `json:"type"`    // http, ping, tcp, dns
    URL         string        `json:"url"`
    Interval    int           `json:"interval"`
    Timeout     int           `json:"timeout"`
    ComponentID string        `json:"component_id"`
    IsActive    bool          `json:"is_active"`
}

type MonitorResult struct {
    MonitorID    string    `json:"monitor_id"`
    Status       string    `json:"status"`
    ResponseTime int       `json:"response_time"`
    StatusCode   int       `json:"status_code"`
    CheckedAt    time.Time `json:"checked_at"`
}
```

**API Endpoints**:
- `GET /monitors` - List tenant monitors
- `POST /monitors` - Create monitor
- `GET /monitors/{id}/status` - Monitor status
- `POST /monitors/{id}/pause` - Pause monitor
- `GET /uptime` - Uptime statistics

**Integration Points**:
- **Component Service**: Updates component status based on monitor results
- **Incident Service**: Creates incidents for monitor failures
- **Notification Service**: Sends alerts for status changes

---

### **🚨 Incident Service (Port 8086)**
**Purpose**: Incident management and communication
**Database**: `incident_service` with tenant isolation

**Key Models**:
```go
type Incident struct {
    ID            string           `json:"id"`
    TenantID      string           `json:"tenant_id"`
    Title         string           `json:"title"`
    Description   string           `json:"description"`
    Status        IncidentStatus   `json:"status"`
    Severity      IncidentSeverity `json:"severity"`
    ComponentIDs  []string         `json:"component_ids"`
    CreatedBy     string           `json:"created_by"`
    CreatedAt     time.Time        `json:"created_at"`
    ResolvedAt    *time.Time       `json:"resolved_at"`
}

type IncidentUpdate struct {
    ID         string    `json:"id"`
    IncidentID string    `json:"incident_id"`
    Status     string    `json:"status"`
    Message    string    `json:"message"`
    CreatedAt  time.Time `json:"created_at"`
}
```

**API Endpoints**:
- `GET /incidents` - List tenant incidents
- `POST /incidents` - Create incident
- `PUT /incidents/{id}` - Update incident
- `POST /incidents/{id}/updates` - Add incident update

---

### **📈 Analytics Service (Port 8090)**
**Purpose**: Usage analytics and insights
**Database**: `analytics_service` with tenant data

**Key Models**:
```go
type AnalyticsData struct {
    TenantID   string    `json:"tenant_id"`
    Metric     string    `json:"metric"`
    Value      float64   `json:"value"`
    Timestamp  time.Time `json:"timestamp"`
    Dimensions map[string]string `json:"dimensions"`
}

type TenantMetrics struct {
    TenantID        string  `json:"tenant_id"`
    PageViews       int     `json:"page_views"`
    UniqueVisitors  int     `json:"unique_visitors"`
    AverageUptime   float64 `json:"average_uptime"`
    IncidentCount   int     `json:"incident_count"`
}
```

---

### **🔔 Notification Service (Port 8085)**
**Purpose**: Alerts and subscriber notifications
**Database**: `notification_service` with tenant isolation

**Key Models**:
```go
type NotificationChannel struct {
    ID       string            `json:"id"`
    TenantID string            `json:"tenant_id"`
    Type     string            `json:"type"`     // email, sms, slack, webhook
    Config   map[string]string `json:"config"`
    IsActive bool              `json:"is_active"`
}

type Subscriber struct {
    ID        string   `json:"id"`
    TenantID  string   `json:"tenant_id"`
    Email     string   `json:"email"`
    Phone     string   `json:"phone"`
    Components []string `json:"components"`
}
```

---

### **💰 Payment Service (Port 8088)**
**Purpose**: Billing and subscription management
**Integration**: Stripe, PayPal, etc.

**Key Models**:
```go
type Subscription struct {
    ID         string    `json:"id"`
    TenantID   string    `json:"tenant_id"`
    PlanID     string    `json:"plan_id"`
    Status     string    `json:"status"`
    CurrentPeriodStart time.Time `json:"current_period_start"`
    CurrentPeriodEnd   time.Time `json:"current_period_end"`
}

type Invoice struct {
    ID       string  `json:"id"`
    TenantID string  `json:"tenant_id"`
    Amount   float64 `json:"amount"`
    Status   string  `json:"status"`
    DueDate  time.Time `json:"due_date"`
}
```

---

### **🎨 Branding Service (Port 8097)**
**Purpose**: Tenant visual customization
**Features**: Custom logos, colors, domains

---

### **🌐 Status UI Service (Port 8093)**
**Purpose**: Public status page rendering
**Features**: Server-side rendering of tenant status pages

---

### **🏠 Landing Page Service (Port 8100)**
**Purpose**: Public marketing and signup pages
**Features**: SEO-optimized landing pages

---

## 🔄 **Data Flow & Integration Patterns**

### **Typical Status Update Flow**:
```
1. Monitor Service → Detects downtime
2. Component Service → Updates component status
3. Incident Service → Creates incident (if critical)
4. Notification Service → Alerts subscribers
5. Analytics Service → Records metrics
6. Status UI Service → Updates public page
```

### **Tenant Onboarding Flow**:
```
1. SaaS Admin → Creates tenant account
2. Database Service → Provisions tenant data
3. User Service → Creates initial admin user
4. Tenant Admin → Configures status page
5. Component Service → Sets up components
6. Monitoring Service → Configures monitors
7. Branding Service → Applies customization
```

---

## 🗄️ **Database Architecture**

### **Multi-tenancy Strategy**: Row-Level Security (RLS)
- All tables include `tenant_id` column
- Middleware enforces tenant context
- No cross-tenant data access
- Shared PostgreSQL instances with logical isolation

### **Database Distribution**:
```
saas_admin: Platform-level data
user_service: User accounts and auth
component_service: Components and groups
monitoring_service: Monitors and results
incident_service: Incidents and updates
analytics_service: Metrics and insights
notification_service: Subscribers and channels
payment_service: Billing and subscriptions
branding_service: Customization settings
```

---

## ⚡ **Event-Driven Architecture**

### **Kafka Topics**:
- `tenant.created` → Provision resources
- `component.status.changed` → Update dependent services
- `incident.created` → Send notifications
- `monitor.failed` → Create incidents
- `user.activity` → Track analytics

### **Consumer Services**:
- **Analytics Consumer**: Process usage metrics
- **Audit Consumer**: Maintain audit trails
- **Billing Consumer**: Handle subscription events
- **Notification Consumer**: Deliver notifications

---

## 🚨 **Critical Issues Identified**

### **1. Port Configuration Mismatches**
| Service | API Gateway Expects | Actual Port | Status |
|---------|-------------------|-------------|--------|
| component-service | 8083 | 8084 | ❌ Fix needed |
| incident-service | 8084 | 8086 | ❌ Fix needed |
| monitoring-service | 8088 | 8092 | ❌ Fix needed |
| analytics-service | 8087 | 8090 | ❌ Fix needed |

### **2. Service Discovery**
- No dynamic service discovery
- Hardcoded service endpoints
- Manual configuration maintenance

### **3. Missing Real-time Capabilities**
- No WebSocket connections for live updates
- No event streaming for status changes
- Limited real-time dashboard functionality

---

## 🛠️ **Recommended Next Steps**

### **Immediate (Week 1)**:
1. Fix API Gateway port configurations
2. Start monitoring-service and component-service
3. Test tenant admin dashboard functionality

### **Short-term (Weeks 2-4)**:
1. Implement real-time WebSocket streaming
2. Add service discovery mechanism
3. Enhance monitoring capabilities

### **Medium-term (Months 1-3)**:
1. Advanced analytics and reporting
2. Multi-region monitoring
3. Integration marketplace

---

*Last Updated: 2025-09-25*
*Document Version: 1.0*
*Services Analyzed: 20 microservices + 4 consumers + 1 gateway*