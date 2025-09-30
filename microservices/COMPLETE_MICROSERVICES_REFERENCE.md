# Complete Beakon Platform Microservices Reference

*Exhaustive documentation of all 20+ microservices with detailed technical analysis*

---

## 🏗️ **Platform Overview**

**Total Services**: 23 total components
- **16 HTTP Microservices** (REST APIs)
- **4 Consumer Services** (Event processors)
- **1 API Gateway** (Traffic management)
- **1 Shared Library** (Common utilities)
- **1 Utility Service** (Shared resilience)

**Architecture**: Multi-tenant SaaS platform with event-driven microservices
**Technology Stack**: Go + Gin + PostgreSQL + Redis + Kafka
**Multi-tenancy**: Row-level security with tenant ID isolation

---

## 📋 **Complete Service Inventory**

### **HTTP Microservices (16 Services)**

| # | Service Name | Port | Health Port | Metrics Port | Database | Purpose |
|---|--------------|------|-------------|--------------|----------|---------|
| 1 | **api-gateway** | 8080 | 8081 | 9090 | - | Central routing & auth |
| 2 | **user-service** | 8081 | 8082 | 9091 | user_service | User auth & profiles |
| 3 | **component-service** | 8084 | 8085 | 9094 | component_service | Status components |
| 4 | **notification-service** | 8085 | - | - | notification_service | Multi-channel alerts |
| 5 | **incident-service** | 8086 | 8087 | 9096 | incident_service | Incident management |
| 6 | **payment-service** | 8088 | 8089 | 9098 | payment_service | Billing & subscriptions |
| 7 | **analytics-service** | 8090 | 8091 | 9100 | analytics_service | Metrics & reporting |
| 8 | **monitoring-service** | 8092 | 8093 | 9102 | monitoring_service | Uptime monitoring |
| 9 | **status-ui-service** | 8093 | - | - | - | Public status pages |
| 10 | **database-service** | 8095 | - | - | database_service | DB operations |
| 11 | **event-store-service** | 8096 | - | - | event_store_service | Event sourcing |
| 12 | **branding-service** | 8097 | - | - | branding_service | UI customization |
| 13 | **saas-admin-service** | 8098 | - | - | saas_admin | Platform admin |
| 14 | **tenant-admin-service** | 8099 | - | - | tenant_admin | Tenant dashboards |
| 15 | **landing-page-service** | 8100 | - | - | landing_page_service | Marketing pages |
| 16 | **shared-resilience** | - | - | - | - | Common utilities |

### **Consumer Services (4 Services)**

| # | Service Name | Type | Purpose |
|---|--------------|------|---------|
| 17 | **analytics-consumer** | Kafka Consumer | Process analytics events |
| 18 | **audit-consumer** | Kafka Consumer | Audit log processing |
| 19 | **billing-consumer** | Kafka Consumer | Billing event processing |
| 20 | **notification-consumer** | Kafka Consumer | Notification delivery |

---

## 🔍 **Detailed Service Analysis**

### **1. API Gateway (Port 8080)**
**Purpose**: Central entry point for all client requests
**Technology**: Go + Gin with advanced routing
**Type**: Infrastructure service

**Key Features**:
- **Request Routing**: Intelligent routing to 15+ downstream services
- **Authentication**: JWT middleware with tenant context extraction
- **Rate Limiting**: Configurable per-service and per-tenant limits
- **Circuit Breaker**: Fault tolerance with automatic fallback
- **Load Balancing**: Round-robin and weighted routing
- **Metrics Collection**: Prometheus metrics for all traffic
- **CORS Handling**: Cross-origin request management
- **Request/Response Transformation**: Header injection and data transformation

**Service Routing Configuration**:
```go
// Critical Issue: Port mismatches found
UserService:         "http://localhost:8081" ✅
TenantService:       "http://localhost:8082" ❌ (actual: 8099)
ComponentService:    "http://localhost:8083" ❌ (actual: 8084)
IncidentService:     "http://localhost:8084" ❌ (actual: 8086)
PaymentService:      "http://localhost:8086" ❌ (actual: 8088)
AnalyticsService:    "http://localhost:8087" ❌ (actual: 8090)
NotificationService: "http://localhost:8085" ✅
MonitoringService:   "http://localhost:8088" ❌ (actual: 8092)
```

**API Routes**:
```
# Authentication & Users
POST /api/v1/auth/login
POST /api/v1/auth/register
GET  /api/v1/auth/me
POST /api/v1/auth/refresh

# Multi-Service Routing
/api/v1/users/*          → user-service
/api/v1/tenants/*        → tenant-admin-service
/api/v1/components/*     → component-service
/api/v1/incidents/*      → incident-service
/api/v1/payments/*       → payment-service
/api/v1/analytics/*      → analytics-service
/api/v1/monitoring/*     → monitoring-service
/api/v1/notifications/*  → notification-service
```

**Middleware Stack**:
1. Request logging
2. Rate limiting (tenant-scoped)
3. Authentication & JWT validation
4. Tenant context injection
5. Circuit breaker
6. Request metrics
7. Error recovery

---

### **2. User Service (Port 8081)**
**Purpose**: User authentication, authorization, and profile management
**Database**: `user_service`
**Multi-tenancy**: Users scoped by `tenant_id`

**Core Models**:
```go
type User struct {
    ID              uint      `json:"id" gorm:"primarykey"`
    TenantID        uint      `json:"tenant_id" gorm:"index"`
    Email           string    `json:"email" gorm:"uniqueIndex"`
    PasswordHash    string    `json:"-"`
    FirstName       string    `json:"first_name"`
    LastName        string    `json:"last_name"`
    Role            UserRole  `json:"role"`
    IsActive        bool      `json:"is_active" gorm:"default:true"`
    EmailVerified   bool      `json:"email_verified" gorm:"default:false"`
    LastLoginAt     *time.Time `json:"last_login_at"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

type UserSession struct {
    ID        uint      `json:"id" gorm:"primarykey"`
    UserID    uint      `json:"user_id" gorm:"index"`
    Token     string    `json:"token" gorm:"uniqueIndex"`
    ExpiresAt time.Time `json:"expires_at"`
    CreatedAt time.Time `json:"created_at"`
}

type PasswordReset struct {
    ID        uint      `json:"id" gorm:"primarykey"`
    UserID    uint      `json:"user_id" gorm:"index"`
    Token     string    `json:"token" gorm:"uniqueIndex"`
    ExpiresAt time.Time `json:"expires_at"`
    UsedAt    *time.Time `json:"used_at"`
    CreatedAt time.Time `json:"created_at"`
}
```

**API Endpoints**:
```
# Public Authentication
POST   /api/v1/public/register     - User registration
POST   /api/v1/public/login        - User authentication
POST   /api/v1/public/forgot-password - Password reset request
POST   /api/v1/public/reset-password  - Password reset confirmation
POST   /api/v1/public/verify-email    - Email verification

# Authenticated Operations
GET    /api/v1/auth/me             - Current user profile
POST   /api/v1/auth/logout         - User logout
POST   /api/v1/auth/refresh        - Token refresh
PUT    /api/v1/auth/password       - Password change

# User Management (Admin)
GET    /api/v1/users               - List tenant users
POST   /api/v1/users               - Create user
GET    /api/v1/users/{id}          - Get user details
PUT    /api/v1/users/{id}          - Update user
DELETE /api/v1/users/{id}          - Delete user
POST   /api/v1/users/{id}/activate - Activate user
POST   /api/v1/users/{id}/deactivate - Deactivate user

# Profile Management
GET    /api/v1/profile             - User profile
PUT    /api/v1/profile             - Update profile
POST   /api/v1/profile/avatar      - Upload avatar
DELETE /api/v1/profile/avatar      - Remove avatar
```

**Authentication Flow**:
1. User provides credentials via `/login`
2. Service validates against database
3. JWT token generated with user + tenant claims
4. Refresh token stored in database
5. Subsequent requests validated via JWT middleware

**Security Features**:
- Password hashing with bcrypt
- JWT with configurable expiration
- Session management with cleanup
- Email verification workflow
- Password reset with secure tokens
- Account lockout protection
- Activity audit logging

---

### **3. Component Service (Port 8084)**
**Purpose**: Status page component management and status tracking
**Database**: `component_service`
**Multi-tenancy**: All components scoped by `tenant_id`

**Core Models**:
```go
type Component struct {
    ID          uint           `json:"id" gorm:"primarykey"`
    TenantID    uint           `json:"tenant_id" gorm:"index"`
    Name        string         `json:"name"`
    Description string         `json:"description"`
    Status      ComponentStatus `json:"status" gorm:"default:'operational'"`
    GroupID     *uint          `json:"group_id" gorm:"index"`
    Position    int            `json:"position" gorm:"default:0"`
    ShowUptime  bool           `json:"show_uptime" gorm:"default:true"`
    IsVisible   bool           `json:"is_visible" gorm:"default:true"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

    // Relationships
    Group       *ComponentGroup `json:"group,omitempty" gorm:"foreignKey:GroupID"`
    StatusHistory []ComponentStatusHistory `json:"status_history,omitempty"`
}

type ComponentGroup struct {
    ID          uint      `json:"id" gorm:"primarykey"`
    TenantID    uint      `json:"tenant_id" gorm:"index"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Position    int       `json:"position" gorm:"default:0"`
    IsCollapsed bool      `json:"is_collapsed" gorm:"default:false"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`

    // Relationships
    Components []Component `json:"components,omitempty"`
}

type ComponentStatus string
const (
    StatusOperational   ComponentStatus = "operational"
    StatusDegraded      ComponentStatus = "degraded_performance"
    StatusPartialOutage ComponentStatus = "partial_outage"
    StatusMajorOutage   ComponentStatus = "major_outage"
    StatusMaintenance   ComponentStatus = "under_maintenance"
)

type ComponentStatusHistory struct {
    ID          uint            `json:"id" gorm:"primarykey"`
    ComponentID uint            `json:"component_id" gorm:"index"`
    TenantID    uint            `json:"tenant_id" gorm:"index"`
    Status      ComponentStatus `json:"status"`
    Message     string          `json:"message"`
    CreatedBy   uint            `json:"created_by"`
    CreatedAt   time.Time       `json:"created_at"`
}
```

**API Endpoints**:
```
# Component Management
GET    /api/v1/components          - List tenant components
POST   /api/v1/components          - Create component
GET    /api/v1/components/{id}     - Get component details
PUT    /api/v1/components/{id}     - Update component
DELETE /api/v1/components/{id}     - Delete component
POST   /api/v1/components/{id}/status - Update component status
GET    /api/v1/components/{id}/history - Component status history
POST   /api/v1/components/{id}/duplicate - Duplicate component

# Component Groups
GET    /api/v1/component-groups    - List component groups
POST   /api/v1/component-groups    - Create component group
GET    /api/v1/component-groups/{id} - Get group details
PUT    /api/v1/component-groups/{id} - Update group
DELETE /api/v1/component-groups/{id} - Delete group
POST   /api/v1/component-groups/{id}/reorder - Reorder components

# Status Operations
GET    /api/v1/status              - Overall status page
GET    /api/v1/status/summary      - Status summary
POST   /api/v1/status/bulk-update  - Bulk status update
GET    /api/v1/status/history      - Global status history

# Public API (No auth required)
GET    /api/v1/public/status       - Public status page
GET    /api/v1/public/components   - Public components list
GET    /api/v1/public/history      - Public status history
```

**Status Calculation Logic**:
```go
func CalculateOverallStatus(components []Component) ComponentStatus {
    if hasStatus(components, StatusMajorOutage) {
        return StatusMajorOutage
    }
    if hasStatus(components, StatusPartialOutage) {
        return StatusPartialOutage
    }
    if hasStatus(components, StatusDegraded) {
        return StatusDegraded
    }
    if hasStatus(components, StatusMaintenance) {
        return StatusMaintenance
    }
    return StatusOperational
}
```

**Integration Points**:
- **Monitoring Service**: Receives status updates from health checks
- **Incident Service**: Component status affects incident severity
- **Status UI Service**: Renders public component display
- **Notification Service**: Status changes trigger notifications

---

### **4. Incident Service (Port 8086)**
**Purpose**: Incident management with automated workflow processing
**Database**: `incident_service`
**Multi-tenancy**: All incidents scoped by `tenant_id`

**Core Models**:
```go
type Incident struct {
    ID            uint             `json:"id" gorm:"primarykey"`
    TenantID      uint             `json:"tenant_id" gorm:"index"`
    Title         string           `json:"title"`
    Description   string           `json:"description"`
    Status        IncidentStatus   `json:"status" gorm:"default:'investigating'"`
    Severity      IncidentSeverity `json:"severity" gorm:"default:'medium'"`
    ComponentIDs  datatypes.JSON   `json:"component_ids" gorm:"type:jsonb"`
    CreatedBy     uint             `json:"created_by"`
    AssignedTo    *uint            `json:"assigned_to"`
    StartedAt     time.Time        `json:"started_at"`
    ResolvedAt    *time.Time       `json:"resolved_at"`
    PostmortemURL string           `json:"postmortem_url"`
    IsPublic      bool             `json:"is_public" gorm:"default:true"`
    CreatedAt     time.Time        `json:"created_at"`
    UpdatedAt     time.Time        `json:"updated_at"`

    // Relationships
    Updates       []IncidentUpdate `json:"updates,omitempty"`
    Components    []Component      `json:"components,omitempty" gorm:"many2many:incident_components;"`
}

type IncidentStatus string
const (
    StatusInvestigating IncidentStatus = "investigating"
    StatusIdentified    IncidentStatus = "identified"
    StatusMonitoring    IncidentStatus = "monitoring"
    StatusResolved      IncidentStatus = "resolved"
    StatusPostmortem    IncidentStatus = "postmortem"
)

type IncidentSeverity string
const (
    SeverityLow      IncidentSeverity = "low"
    SeverityMedium   IncidentSeverity = "medium"
    SeverityHigh     IncidentSeverity = "high"
    SeverityCritical IncidentSeverity = "critical"
)

type IncidentUpdate struct {
    ID         uint           `json:"id" gorm:"primarykey"`
    IncidentID uint           `json:"incident_id" gorm:"index"`
    TenantID   uint           `json:"tenant_id" gorm:"index"`
    Status     IncidentStatus `json:"status"`
    Message    string         `json:"message"`
    CreatedBy  uint           `json:"created_by"`
    IsPublic   bool           `json:"is_public" gorm:"default:true"`
    CreatedAt  time.Time      `json:"created_at"`
}

type IncidentTemplate struct {
    ID          uint             `json:"id" gorm:"primarykey"`
    TenantID    uint             `json:"tenant_id" gorm:"index"`
    Name        string           `json:"name"`
    Title       string           `json:"title"`
    Description string           `json:"description"`
    Severity    IncidentSeverity `json:"severity"`
    ComponentIDs datatypes.JSON  `json:"component_ids" gorm:"type:jsonb"`
    CreatedAt   time.Time        `json:"created_at"`
    UpdatedAt   time.Time        `json:"updated_at"`
}
```

**API Endpoints**:
```
# Incident Management
GET    /api/v1/incidents           - List tenant incidents
POST   /api/v1/incidents           - Create incident
GET    /api/v1/incidents/{id}      - Get incident details
PUT    /api/v1/incidents/{id}      - Update incident
DELETE /api/v1/incidents/{id}      - Delete incident
POST   /api/v1/incidents/{id}/resolve - Resolve incident
POST   /api/v1/incidents/{id}/reopen  - Reopen incident

# Incident Updates
GET    /api/v1/incidents/{id}/updates - List incident updates
POST   /api/v1/incidents/{id}/updates - Create incident update
PUT    /api/v1/incidents/{id}/updates/{updateId} - Update incident update
DELETE /api/v1/incidents/{id}/updates/{updateId} - Delete incident update

# Templates
GET    /api/v1/templates           - List templates
POST   /api/v1/templates           - Create template
GET    /api/v1/templates/{id}      - Get template
PUT    /api/v1/templates/{id}      - Update template
DELETE /api/v1/templates/{id}      - Delete template
POST   /api/v1/templates/{id}/create-incident - Create incident from template

# Workflow Management
POST   /api/v1/incidents/{id}/workflow/{action} - Execute workflow action
GET    /api/v1/workflows           - List available workflows
POST   /api/v1/workflows           - Create custom workflow

# Analytics
GET    /api/v1/incidents/stats     - Incident statistics
GET    /api/v1/incidents/metrics   - MTTR, MTBF metrics
GET    /api/v1/incidents/reports   - Incident reports

# Public API
GET    /api/v1/public/incidents    - Public incidents list
GET    /api/v1/public/incidents/{id} - Public incident details
```

**Workflow Engine**:
```go
type WorkflowStep struct {
    ID          uint              `json:"id"`
    Name        string            `json:"name"`
    Type        WorkflowStepType  `json:"type"`
    Config      datatypes.JSON    `json:"config"`
    Conditions  []WorkflowCondition `json:"conditions"`
    Actions     []WorkflowAction  `json:"actions"`
}

// Automated workflow examples:
// 1. Critical incident → Auto-notify on-call team
// 2. Monitoring failure → Auto-create incident
// 3. Incident resolved → Update component status
// 4. Major outage → Send customer communications
```

**Metrics Tracking**:
- **MTTR** (Mean Time To Recovery)
- **MTBF** (Mean Time Between Failures)
- **Incident frequency by severity**
- **Component impact analysis**
- **Resolution time trends**

---

### **5. Monitoring Service (Port 8092)**
**Purpose**: Uptime monitoring, health checks, and performance tracking
**Database**: `monitoring_service`
**Multi-tenancy**: All monitors scoped by `tenant_id`

**Core Models**:
```go
type Monitor struct {
    ID              uint        `json:"id" gorm:"primarykey"`
    TenantID        uint        `json:"tenant_id" gorm:"index"`
    Name            string      `json:"name"`
    Type            MonitorType `json:"type"`
    URL             string      `json:"url"`
    Method          HTTPMethod  `json:"method" gorm:"default:'GET'"`
    Headers         datatypes.JSON `json:"headers" gorm:"type:jsonb"`
    Body            string      `json:"body"`
    ExpectedStatus  int         `json:"expected_status" gorm:"default:200"`
    ExpectedContent string      `json:"expected_content"`
    Timeout         int         `json:"timeout" gorm:"default:30"`
    Interval        int         `json:"interval" gorm:"default:300"`
    Retries         int         `json:"retries" gorm:"default:3"`
    ComponentID     *uint       `json:"component_id"`
    IsActive        bool        `json:"is_active" gorm:"default:true"`
    CreatedAt       time.Time   `json:"created_at"`
    UpdatedAt       time.Time   `json:"updated_at"`

    // Relationships
    Results         []MonitorResult `json:"results,omitempty"`
    Component       *Component      `json:"component,omitempty" gorm:"foreignKey:ComponentID"`
}

type MonitorType string
const (
    TypeHTTP    MonitorType = "http"
    TypeHTTPS   MonitorType = "https"
    TypePing    MonitorType = "ping"
    TypeTCP     MonitorType = "tcp"
    TypeUDP     MonitorType = "udp"
    TypeDNS     MonitorType = "dns"
    TypeSSL     MonitorType = "ssl"
)

type MonitorResult struct {
    ID           uint      `json:"id" gorm:"primarykey"`
    MonitorID    uint      `json:"monitor_id" gorm:"index"`
    TenantID     uint      `json:"tenant_id" gorm:"index"`
    Status       ResultStatus `json:"status"`
    ResponseTime int       `json:"response_time"` // milliseconds
    StatusCode   int       `json:"status_code"`
    ResponseSize int       `json:"response_size"`
    ErrorMessage string    `json:"error_message"`
    CheckedAt    time.Time `json:"checked_at" gorm:"index"`
    CreatedAt    time.Time `json:"created_at"`
}

type ResultStatus string
const (
    StatusUp      ResultStatus = "up"
    StatusDown    ResultStatus = "down"
    StatusTimeout ResultStatus = "timeout"
    StatusError   ResultStatus = "error"
)

type UptimeStats struct {
    MonitorID       uint    `json:"monitor_id"`
    Period          string  `json:"period"` // "24h", "7d", "30d", "90d"
    UptimePercent   float64 `json:"uptime_percent"`
    TotalChecks     int     `json:"total_checks"`
    SuccessfulChecks int     `json:"successful_checks"`
    FailedChecks    int     `json:"failed_checks"`
    AverageResponse float64 `json:"average_response"`
    DowntimeMinutes int     `json:"downtime_minutes"`
}
```

**API Endpoints**:
```
# Monitor Management
GET    /api/v1/monitors            - List tenant monitors
POST   /api/v1/monitors            - Create monitor
GET    /api/v1/monitors/{id}       - Get monitor details
PUT    /api/v1/monitors/{id}       - Update monitor
DELETE /api/v1/monitors/{id}       - Delete monitor
POST   /api/v1/monitors/{id}/test  - Test monitor now
POST   /api/v1/monitors/{id}/pause - Pause monitor
POST   /api/v1/monitors/{id}/resume - Resume monitor

# Monitor Results
GET    /api/v1/monitors/{id}/results - Monitor check results
GET    /api/v1/monitors/{id}/status  - Current monitor status
GET    /api/v1/monitors/{id}/uptime  - Uptime statistics
GET    /api/v1/monitors/{id}/response-times - Response time graph data

# Bulk Operations
POST   /api/v1/monitors/bulk-pause  - Pause multiple monitors
POST   /api/v1/monitors/bulk-resume - Resume multiple monitors
POST   /api/v1/monitors/bulk-test   - Test multiple monitors
DELETE /api/v1/monitors/bulk-delete - Delete multiple monitors

# Uptime & Analytics
GET    /api/v1/uptime              - Overall uptime stats
GET    /api/v1/uptime/components   - Component uptime stats
GET    /api/v1/uptime/history      - Historical uptime data
GET    /api/v1/performance         - Performance metrics
GET    /api/v1/performance/trends  - Performance trends

# Alerts & Notifications
GET    /api/v1/monitors/{id}/alerts - Monitor alert rules
POST   /api/v1/monitors/{id}/alerts - Create alert rule
PUT    /api/v1/monitors/{id}/alerts/{alertId} - Update alert rule
DELETE /api/v1/monitors/{id}/alerts/{alertId} - Delete alert rule

# Public Status API
GET    /api/v1/public/monitors/{id}/status - Public monitor status
GET    /api/v1/public/monitors/{id}/uptime - Public uptime stats
```

**Monitoring Engine**:
```go
type MonitoringEngine struct {
    monitors      map[uint]*Monitor
    scheduler     *cron.Cron
    resultChannel chan MonitorResult
    workers       []*MonitorWorker
}

type MonitorWorker struct {
    id       int
    client   *http.Client
    timeout  time.Duration
    retries  int
}

func (w *MonitorWorker) ExecuteCheck(monitor *Monitor) *MonitorResult {
    start := time.Now()

    // Execute check based on monitor type
    switch monitor.Type {
    case TypeHTTP, TypeHTTPS:
        return w.checkHTTP(monitor)
    case TypePing:
        return w.checkPing(monitor)
    case TypeTCP:
        return w.checkTCP(monitor)
    case TypeDNS:
        return w.checkDNS(monitor)
    case TypeSSL:
        return w.checkSSL(monitor)
    }

    return &MonitorResult{
        Status: StatusError,
        ErrorMessage: "Unknown monitor type",
        CheckedAt: start,
    }
}
```

**Integration Points**:
- **Component Service**: Updates component status based on monitor results
- **Incident Service**: Creates incidents for critical monitor failures
- **Notification Service**: Sends alerts for status changes
- **Analytics Service**: Records performance and uptime metrics

---

### **6. Analytics Service (Port 8090)**
**Purpose**: Advanced analytics, metrics collection, SLA monitoring and business intelligence
**Database**: `analytics_service`
**Multi-tenancy**: All analytics scoped by `tenant_id`

**Core Models**:
```go
type Metric struct {
    ID          uint       `json:"id" gorm:"primarykey"`
    TenantID    uint       `json:"tenant_id" gorm:"index"`
    Name        string     `json:"name" gorm:"index"`
    Type        MetricType `json:"type"`
    Description string     `json:"description"`
    Unit        string     `json:"unit"`
    Tags        datatypes.JSON `json:"tags" gorm:"type:jsonb"`
    IsActive    bool       `json:"is_active" gorm:"default:true"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`

    // Relationships
    DataPoints  []MetricDataPoint `json:"data_points,omitempty"`
}

type MetricType string
const (
    TypeCounter   MetricType = "counter"   // Incrementing values
    TypeGauge     MetricType = "gauge"     // Point-in-time values
    TypeHistogram MetricType = "histogram" // Distribution of values
    TypeTimer     MetricType = "timer"     // Duration measurements
)

type MetricDataPoint struct {
    ID        uint      `json:"id" gorm:"primarykey"`
    MetricID  uint      `json:"metric_id" gorm:"index"`
    TenantID  uint      `json:"tenant_id" gorm:"index"`
    Value     float64   `json:"value"`
    Tags      datatypes.JSON `json:"tags" gorm:"type:jsonb"`
    Timestamp time.Time `json:"timestamp" gorm:"index"`
    CreatedAt time.Time `json:"created_at"`
}

type SLA struct {
    ID            uint      `json:"id" gorm:"primarykey"`
    TenantID      uint      `json:"tenant_id" gorm:"index"`
    Name          string    `json:"name"`
    Description   string    `json:"description"`
    TargetPercent float64   `json:"target_percent"` // 99.9, 99.95, 99.99
    ComponentIDs  datatypes.JSON `json:"component_ids" gorm:"type:jsonb"`
    MonitorIDs    datatypes.JSON `json:"monitor_ids" gorm:"type:jsonb"`
    Period        SLAPeriod `json:"period" gorm:"default:'monthly'"`
    IsActive      bool      `json:"is_active" gorm:"default:true"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`

    // Relationships
    Measurements []SLAMeasurement `json:"measurements,omitempty"`
}

type SLAPeriod string
const (
    PeriodDaily    SLAPeriod = "daily"
    PeriodWeekly   SLAPeriod = "weekly"
    PeriodMonthly  SLAPeriod = "monthly"
    PeriodQuarterly SLAPeriod = "quarterly"
    PeriodYearly   SLAPeriod = "yearly"
)

type SLAMeasurement struct {
    ID               uint      `json:"id" gorm:"primarykey"`
    SLAID            uint      `json:"sla_id" gorm:"index"`
    TenantID         uint      `json:"tenant_id" gorm:"index"`
    PeriodStart      time.Time `json:"period_start"`
    PeriodEnd        time.Time `json:"period_end"`
    ActualPercent    float64   `json:"actual_percent"`
    TotalChecks      int       `json:"total_checks"`
    SuccessfulChecks int       `json:"successful_checks"`
    DowntimeMinutes  int       `json:"downtime_minutes"`
    IsBreached       bool      `json:"is_breached"`
    CalculatedAt     time.Time `json:"calculated_at"`
    CreatedAt        time.Time `json:"created_at"`
}

type Dashboard struct {
    ID          uint      `json:"id" gorm:"primarykey"`
    TenantID    uint      `json:"tenant_id" gorm:"index"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Layout      datatypes.JSON `json:"layout" gorm:"type:jsonb"`
    IsPublic    bool      `json:"is_public" gorm:"default:false"`
    CreatedBy   uint      `json:"created_by"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`

    // Relationships
    Widgets     []DashboardWidget `json:"widgets,omitempty"`
}

type DashboardWidget struct {
    ID          uint           `json:"id" gorm:"primarykey"`
    DashboardID uint           `json:"dashboard_id" gorm:"index"`
    TenantID    uint           `json:"tenant_id" gorm:"index"`
    Type        WidgetType     `json:"type"`
    Title       string         `json:"title"`
    Config      datatypes.JSON `json:"config" gorm:"type:jsonb"`
    Position    datatypes.JSON `json:"position" gorm:"type:jsonb"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
}

type WidgetType string
const (
    WidgetMetric      WidgetType = "metric"
    WidgetChart       WidgetType = "chart"
    WidgetTable       WidgetType = "table"
    WidgetUptime      WidgetType = "uptime"
    WidgetIncidents   WidgetType = "incidents"
    WidgetComponents  WidgetType = "components"
    WidgetSLA         WidgetType = "sla"
)
```

**API Endpoints**:
```
# Analytics Overview
GET    /api/v1/analytics/overview   - Analytics dashboard overview
GET    /api/v1/analytics/summary    - Key metrics summary
GET    /api/v1/analytics/trends     - Trend analysis
GET    /api/v1/analytics/health     - Service health check

# Metrics Management
GET    /api/v1/metrics              - List metrics
POST   /api/v1/metrics              - Create metric
GET    /api/v1/metrics/{id}         - Get metric details
PUT    /api/v1/metrics/{id}         - Update metric
DELETE /api/v1/metrics/{id}         - Delete metric
POST   /api/v1/metrics/{id}/data    - Submit data points
GET    /api/v1/metrics/{id}/data    - Query data points

# SLA Management
GET    /api/v1/sla                  - List SLAs
POST   /api/v1/sla                  - Create SLA
GET    /api/v1/sla/{id}             - Get SLA details
PUT    /api/v1/sla/{id}             - Update SLA
DELETE /api/v1/sla/{id}             - Delete SLA
GET    /api/v1/sla/{id}/measurements - SLA measurements
POST   /api/v1/sla/{id}/calculate   - Calculate SLA
GET    /api/v1/sla/{id}/breaches    - SLA breaches
GET    /api/v1/sla/reports          - SLA reports

# Dashboard Management
GET    /api/v1/dashboards           - List dashboards
POST   /api/v1/dashboards           - Create dashboard
GET    /api/v1/dashboards/{id}      - Get dashboard
PUT    /api/v1/dashboards/{id}      - Update dashboard
DELETE /api/v1/dashboards/{id}      - Delete dashboard
GET    /api/v1/dashboards/{id}/data - Dashboard data
POST   /api/v1/dashboards/{id}/widgets - Add widget
PUT    /api/v1/dashboards/{id}/widgets/{widgetId} - Update widget
DELETE /api/v1/dashboards/{id}/widgets/{widgetId} - Remove widget

# Reports & Exports
GET    /api/v1/reports              - List reports
POST   /api/v1/reports              - Create report
GET    /api/v1/reports/{id}         - Get report
POST   /api/v1/reports/{id}/generate - Generate report
GET    /api/v1/reports/{id}/download - Download report
POST   /api/v1/exports              - Create data export
GET    /api/v1/exports/{id}         - Download export

# Public Analytics API
GET    /api/v1/public/uptime        - Public uptime stats
GET    /api/v1/public/incidents/summary - Public incident summary
GET    /api/v1/public/performance   - Public performance metrics
```

**SLA Calculation Engine**:
```go
type SLACalculator struct {
    db     *gorm.DB
    cache  cache.Cache
}

func (c *SLACalculator) CalculateSLA(sla *SLA, period TimePeriod) (*SLAMeasurement, error) {
    // Get all monitor results for the period
    results := c.getMonitorResults(sla.MonitorIDs, period)

    // Calculate uptime percentage
    totalChecks := len(results)
    successfulChecks := c.countSuccessfulChecks(results)
    actualPercent := float64(successfulChecks) / float64(totalChecks) * 100

    // Calculate downtime
    downtimeMinutes := c.calculateDowntime(results)

    // Check if SLA is breached
    isBreached := actualPercent < sla.TargetPercent

    return &SLAMeasurement{
        SLAID:            sla.ID,
        PeriodStart:      period.Start,
        PeriodEnd:        period.End,
        ActualPercent:    actualPercent,
        TotalChecks:      totalChecks,
        SuccessfulChecks: successfulChecks,
        DowntimeMinutes:  downtimeMinutes,
        IsBreached:       isBreached,
        CalculatedAt:     time.Now(),
    }
}
```

---

### **7. Notification Service (Port 8085)**
**Purpose**: Multi-channel notification delivery with subscriber management
**Database**: `notification_service`
**Multi-tenancy**: All notifications scoped by `tenant_id`

**Core Models**:
```go
type NotificationChannel struct {
    ID          uint           `json:"id" gorm:"primarykey"`
    TenantID    uint           `json:"tenant_id" gorm:"index"`
    Name        string         `json:"name"`
    Type        ChannelType    `json:"type"`
    Config      datatypes.JSON `json:"config" gorm:"type:jsonb"`
    IsActive    bool           `json:"is_active" gorm:"default:true"`
    IsDefault   bool           `json:"is_default" gorm:"default:false"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
}

type ChannelType string
const (
    ChannelEmail    ChannelType = "email"
    ChannelSMS      ChannelType = "sms"
    ChannelSlack    ChannelType = "slack"
    ChannelDiscord  ChannelType = "discord"
    ChannelWebhook  ChannelType = "webhook"
    ChannelPagerDuty ChannelType = "pagerduty"
    ChannelTeams    ChannelType = "teams"
)

type Subscriber struct {
    ID             uint      `json:"id" gorm:"primarykey"`
    TenantID       uint      `json:"tenant_id" gorm:"index"`
    Email          string    `json:"email" gorm:"index"`
    Phone          string    `json:"phone"`
    Name           string    `json:"name"`
    ComponentIDs   datatypes.JSON `json:"component_ids" gorm:"type:jsonb"`
    NotificationTypes datatypes.JSON `json:"notification_types" gorm:"type:jsonb"`
    IsActive       bool      `json:"is_active" gorm:"default:true"`
    IsVerified     bool      `json:"is_verified" gorm:"default:false"`
    VerifiedAt     *time.Time `json:"verified_at"`
    UnsubscribeToken string  `json:"unsubscribe_token" gorm:"index"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}

type Notification struct {
    ID          uint             `json:"id" gorm:"primarykey"`
    TenantID    uint             `json:"tenant_id" gorm:"index"`
    Type        NotificationType `json:"type"`
    Title       string           `json:"title"`
    Message     string           `json:"message"`
    ComponentID *uint            `json:"component_id"`
    IncidentID  *uint            `json:"incident_id"`
    Priority    Priority         `json:"priority" gorm:"default:'normal'"`
    Channels    datatypes.JSON   `json:"channels" gorm:"type:jsonb"`
    Recipients  datatypes.JSON   `json:"recipients" gorm:"type:jsonb"`
    Status      NotificationStatus `json:"status" gorm:"default:'pending'"`
    SentAt      *time.Time       `json:"sent_at"`
    FailureReason string         `json:"failure_reason"`
    RetryCount  int              `json:"retry_count" gorm:"default:0"`
    CreatedAt   time.Time        `json:"created_at"`
    UpdatedAt   time.Time        `json:"updated_at"`
}

type NotificationType string
const (
    TypeIncidentCreated   NotificationType = "incident_created"
    TypeIncidentUpdated   NotificationType = "incident_updated"
    TypeIncidentResolved  NotificationType = "incident_resolved"
    TypeComponentDown     NotificationType = "component_down"
    TypeComponentUp       NotificationType = "component_up"
    TypeMaintenanceStart  NotificationType = "maintenance_start"
    TypeMaintenanceEnd    NotificationType = "maintenance_end"
    TypeSLABreach         NotificationType = "sla_breach"
)

type NotificationStatus string
const (
    StatusPending   NotificationStatus = "pending"
    StatusSending   NotificationStatus = "sending"
    StatusSent      NotificationStatus = "sent"
    StatusFailed    NotificationStatus = "failed"
    StatusCancelled NotificationStatus = "cancelled"
)

type Priority string
const (
    PriorityLow      Priority = "low"
    PriorityNormal   Priority = "normal"
    PriorityHigh     Priority = "high"
    PriorityCritical Priority = "critical"
)
```

**API Endpoints**:
```
# Notification Channels
GET    /api/v1/channels             - List notification channels
POST   /api/v1/channels             - Create channel
GET    /api/v1/channels/{id}        - Get channel details
PUT    /api/v1/channels/{id}        - Update channel
DELETE /api/v1/channels/{id}        - Delete channel
POST   /api/v1/channels/{id}/test   - Test channel

# Subscriber Management
GET    /api/v1/subscribers          - List subscribers
POST   /api/v1/subscribers          - Create subscriber
GET    /api/v1/subscribers/{id}     - Get subscriber
PUT    /api/v1/subscribers/{id}     - Update subscriber
DELETE /api/v1/subscribers/{id}     - Delete subscriber
POST   /api/v1/subscribers/{id}/verify - Verify subscriber
POST   /api/v1/subscribers/import   - Bulk import subscribers
GET    /api/v1/subscribers/export   - Export subscribers

# Notification Management
GET    /api/v1/notifications        - List notifications
POST   /api/v1/notifications        - Send notification
GET    /api/v1/notifications/{id}   - Get notification details
POST   /api/v1/notifications/{id}/retry - Retry failed notification
DELETE /api/v1/notifications/{id}   - Cancel notification
GET    /api/v1/notifications/stats  - Notification statistics

# Templates
GET    /api/v1/templates            - List notification templates
POST   /api/v1/templates            - Create template
GET    /api/v1/templates/{id}       - Get template
PUT    /api/v1/templates/{id}       - Update template
DELETE /api/v1/templates/{id}       - Delete template

# Public Subscription API
POST   /api/v1/public/subscribe     - Public subscription
GET    /api/v1/public/unsubscribe/{token} - Unsubscribe
POST   /api/v1/public/verify/{token} - Verify subscription
```

**Notification Engine**:
```go
type NotificationEngine struct {
    channels map[ChannelType]NotificationChannel
    queue    chan *Notification
    workers  []*NotificationWorker
}

type NotificationWorker struct {
    id       int
    engine   *NotificationEngine
    channels map[ChannelType]ChannelProvider
}

type ChannelProvider interface {
    Send(notification *Notification, recipient string) error
    Validate(config map[string]interface{}) error
    GetDeliveryStatus(messageID string) (DeliveryStatus, error)
}

// Example channel implementations
type EmailProvider struct {
    smtpConfig SMTPConfig
    templates  *template.Template
}

type SlackProvider struct {
    webhookURL string
    client     *http.Client
}

type SMSProvider struct {
    provider   string // "twilio", "aws-sns"
    apiKey     string
    client     interface{}
}
```

---

### **8. Payment Service (Port 8088)**
**Purpose**: Billing, subscription management, and payment processing
**Database**: `payment_service`
**Multi-tenancy**: All billing scoped by `tenant_id`

**Core Models**:
```go
type Subscription struct {
    ID                 uint             `json:"id" gorm:"primarykey"`
    TenantID           uint             `json:"tenant_id" gorm:"index"`
    PlanID             uint             `json:"plan_id" gorm:"index"`
    Status             SubscriptionStatus `json:"status" gorm:"default:'active'"`
    StripeSubscriptionID string         `json:"stripe_subscription_id" gorm:"index"`
    CurrentPeriodStart time.Time        `json:"current_period_start"`
    CurrentPeriodEnd   time.Time        `json:"current_period_end"`
    TrialStart         *time.Time       `json:"trial_start"`
    TrialEnd           *time.Time       `json:"trial_end"`
    CancelledAt        *time.Time       `json:"cancelled_at"`
    CreatedAt          time.Time        `json:"created_at"`
    UpdatedAt          time.Time        `json:"updated_at"`

    // Relationships
    Plan               *Plan            `json:"plan,omitempty" gorm:"foreignKey:PlanID"`
    Invoices           []Invoice        `json:"invoices,omitempty"`
}

type SubscriptionStatus string
const (
    StatusActive             SubscriptionStatus = "active"
    StatusPastDue            SubscriptionStatus = "past_due"
    StatusCancelled          SubscriptionStatus = "cancelled"
    StatusUnpaid             SubscriptionStatus = "unpaid"
    StatusIncompleteExpired  SubscriptionStatus = "incomplete_expired"
    StatusTrialing           SubscriptionStatus = "trialing"
)

type Plan struct {
    ID          uint      `json:"id" gorm:"primarykey"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Price       float64   `json:"price"`
    Currency    string    `json:"currency" gorm:"default:'usd'"`
    Interval    string    `json:"interval"` // "month", "year"
    IntervalCount int     `json:"interval_count" gorm:"default:1"`
    Features    datatypes.JSON `json:"features" gorm:"type:jsonb"`
    Limits      datatypes.JSON `json:"limits" gorm:"type:jsonb"`
    IsActive    bool      `json:"is_active" gorm:"default:true"`
    IsPublic    bool      `json:"is_public" gorm:"default:true"`
    StripePriceID string  `json:"stripe_price_id"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type Invoice struct {
    ID               uint          `json:"id" gorm:"primarykey"`
    TenantID         uint          `json:"tenant_id" gorm:"index"`
    SubscriptionID   uint          `json:"subscription_id" gorm:"index"`
    StripeInvoiceID  string        `json:"stripe_invoice_id" gorm:"index"`
    Number           string        `json:"number" gorm:"index"`
    Status           InvoiceStatus `json:"status" gorm:"default:'draft'"`
    Amount           float64       `json:"amount"`
    Tax              float64       `json:"tax" gorm:"default:0"`
    Total            float64       `json:"total"`
    Currency         string        `json:"currency" gorm:"default:'usd'"`
    PeriodStart      time.Time     `json:"period_start"`
    PeriodEnd        time.Time     `json:"period_end"`
    DueDate          time.Time     `json:"due_date"`
    PaidAt           *time.Time    `json:"paid_at"`
    CreatedAt        time.Time     `json:"created_at"`
    UpdatedAt        time.Time     `json:"updated_at"`

    // Relationships
    LineItems        []InvoiceLineItem `json:"line_items,omitempty"`
}

type InvoiceStatus string
const (
    InvoiceDraft                InvoiceStatus = "draft"
    InvoiceOpen                 InvoiceStatus = "open"
    InvoicePaid                 InvoiceStatus = "paid"
    InvoiceVoid                 InvoiceStatus = "void"
    InvoiceUncollectible        InvoiceStatus = "uncollectible"
)

type PaymentMethod struct {
    ID                 uint      `json:"id" gorm:"primarykey"`
    TenantID           uint      `json:"tenant_id" gorm:"index"`
    StripePaymentMethodID string `json:"stripe_payment_method_id" gorm:"index"`
    Type               string    `json:"type"` // "card", "bank_account"
    Last4              string    `json:"last4"`
    Brand              string    `json:"brand"`
    ExpiryMonth        int       `json:"expiry_month"`
    ExpiryYear         int       `json:"expiry_year"`
    IsDefault          bool      `json:"is_default" gorm:"default:false"`
    CreatedAt          time.Time `json:"created_at"`
    UpdatedAt          time.Time `json:"updated_at"`
}

type Usage struct {
    ID           uint      `json:"id" gorm:"primarykey"`
    TenantID     uint      `json:"tenant_id" gorm:"index"`
    MetricName   string    `json:"metric_name" gorm:"index"`
    Value        float64   `json:"value"`
    Period       string    `json:"period"` // "2024-01"
    RecordedAt   time.Time `json:"recorded_at"`
    CreatedAt    time.Time `json:"created_at"`
}
```

**API Endpoints**:
```
# Subscription Management
GET    /api/v1/subscriptions        - Get tenant subscription
PUT    /api/v1/subscriptions        - Update subscription
POST   /api/v1/subscriptions/cancel - Cancel subscription
POST   /api/v1/subscriptions/reactivate - Reactivate subscription
GET    /api/v1/subscriptions/usage  - Usage statistics
POST   /api/v1/subscriptions/upgrade - Upgrade plan
POST   /api/v1/subscriptions/downgrade - Downgrade plan

# Plan Management
GET    /api/v1/plans                - List available plans
GET    /api/v1/plans/{id}           - Get plan details
GET    /api/v1/plans/compare        - Compare plans
POST   /api/v1/plans/{id}/subscribe - Subscribe to plan

# Invoice Management
GET    /api/v1/invoices             - List invoices
GET    /api/v1/invoices/{id}        - Get invoice details
GET    /api/v1/invoices/{id}/download - Download invoice PDF
POST   /api/v1/invoices/{id}/pay    - Pay invoice
GET    /api/v1/invoices/upcoming    - Preview upcoming invoice

# Payment Methods
GET    /api/v1/payment-methods      - List payment methods
POST   /api/v1/payment-methods      - Add payment method
PUT    /api/v1/payment-methods/{id} - Update payment method
DELETE /api/v1/payment-methods/{id} - Remove payment method
POST   /api/v1/payment-methods/{id}/default - Set as default

# Billing & Usage
GET    /api/v1/billing/summary      - Billing summary
GET    /api/v1/billing/history      - Billing history
POST   /api/v1/usage                - Record usage
GET    /api/v1/usage                - Get usage data
GET    /api/v1/usage/limits         - Check usage limits

# Webhooks (Stripe)
POST   /api/v1/webhooks/stripe      - Stripe webhook endpoint
```

**Stripe Integration**:
```go
type StripeService struct {
    client *stripe.Client
    config StripeConfig
}

func (s *StripeService) CreateSubscription(tenantID uint, planID string, paymentMethodID string) (*Subscription, error) {
    // Create Stripe customer
    customer, err := s.createOrUpdateCustomer(tenantID)
    if err != nil {
        return nil, err
    }

    // Attach payment method
    if err := s.attachPaymentMethod(paymentMethodID, customer.ID); err != nil {
        return nil, err
    }

    // Create subscription
    params := &stripe.SubscriptionParams{
        Customer: stripe.String(customer.ID),
        Items: []*stripe.SubscriptionItemsParams{
            {
                Price: stripe.String(planID),
            },
        },
        DefaultPaymentMethod: stripe.String(paymentMethodID),
    }

    stripeSubscription, err := subscription.New(params)
    if err != nil {
        return nil, err
    }

    // Save to database
    return s.saveSubscription(tenantID, stripeSubscription)
}
```

---

### **9. Tenant Admin Service (Port 8099)**
**Purpose**: Individual tenant administration dashboards and management
**Database**: Multi-database access with tenant context
**Multi-tenancy**: Primary service for tenant-scoped operations

**Core Models**:
```go
type TenantSettings struct {
    ID               uint      `json:"id" gorm:"primarykey"`
    TenantID         uint      `json:"tenant_id" gorm:"index"`
    CompanyName      string    `json:"company_name"`
    CompanyWebsite   string    `json:"company_website"`
    CompanyLogo      string    `json:"company_logo"`
    StatusPageTitle  string    `json:"status_page_title"`
    StatusPageURL    string    `json:"status_page_url"`
    CustomDomain     string    `json:"custom_domain"`
    Timezone         string    `json:"timezone" gorm:"default:'UTC'"`
    DateFormat       string    `json:"date_format" gorm:"default:'YYYY-MM-DD'"`
    TimeFormat       string    `json:"time_format" gorm:"default:'24h'"`
    Language         string    `json:"language" gorm:"default:'en'"`
    Theme            datatypes.JSON `json:"theme" gorm:"type:jsonb"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}

type TenantUser struct {
    ID          uint      `json:"id" gorm:"primarykey"`
    TenantID    uint      `json:"tenant_id" gorm:"index"`
    UserID      uint      `json:"user_id" gorm:"index"`
    Role        TenantRole `json:"role" gorm:"default:'member'"`
    Permissions datatypes.JSON `json:"permissions" gorm:"type:jsonb"`
    InvitedBy   uint      `json:"invited_by"`
    InvitedAt   time.Time `json:"invited_at"`
    JoinedAt    *time.Time `json:"joined_at"`
    IsActive    bool      `json:"is_active" gorm:"default:true"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`

    // Relationships
    User        User      `json:"user" gorm:"foreignKey:UserID"`
}

type TenantRole string
const (
    RoleOwner      TenantRole = "owner"       // Full access
    RoleAdmin      TenantRole = "admin"       // Most permissions
    RoleMember     TenantRole = "member"      // Limited permissions
    RoleViewer     TenantRole = "viewer"      // Read-only access
)

type TenantInvitation struct {
    ID        uint       `json:"id" gorm:"primarykey"`
    TenantID  uint       `json:"tenant_id" gorm:"index"`
    Email     string     `json:"email" gorm:"index"`
    Role      TenantRole `json:"role"`
    Token     string     `json:"token" gorm:"uniqueIndex"`
    InvitedBy uint       `json:"invited_by"`
    ExpiresAt time.Time  `json:"expires_at"`
    AcceptedAt *time.Time `json:"accepted_at"`
    CreatedAt time.Time  `json:"created_at"`
}

type TenantStats struct {
    TenantID            uint    `json:"tenant_id"`
    ComponentsCount     int     `json:"components_count"`
    MonitorsCount       int     `json:"monitors_count"`
    IncidentsCount      int     `json:"incidents_count"`
    SubscribersCount    int     `json:"subscribers_count"`
    UptimePercentage    float64 `json:"uptime_percentage"`
    AverageResponseTime float64 `json:"average_response_time"`
    LastIncidentDate    *time.Time `json:"last_incident_date"`
}
```

**API Endpoints**:
```
# Dashboard
GET    /api/v1/dashboard            - Tenant dashboard overview
GET    /api/v1/dashboard/stats      - Dashboard statistics
GET    /api/v1/dashboard/activity   - Recent activity feed
GET    /api/v1/dashboard/notifications - Dashboard notifications

# Tenant Settings
GET    /api/v1/settings             - Get tenant settings
PUT    /api/v1/settings             - Update tenant settings
POST   /api/v1/settings/logo        - Upload company logo
DELETE /api/v1/settings/logo        - Remove company logo
GET    /api/v1/settings/theme       - Get theme settings
PUT    /api/v1/settings/theme       - Update theme settings

# Team Management
GET    /api/v1/team/members         - List team members
POST   /api/v1/team/invite          - Invite team member
PUT    /api/v1/team/members/{id}    - Update member role
DELETE /api/v1/team/members/{id}    - Remove team member
GET    /api/v1/team/invitations     - List pending invitations
POST   /api/v1/team/invitations/{id}/resend - Resend invitation
DELETE /api/v1/team/invitations/{id} - Cancel invitation

# Role & Permission Management
GET    /api/v1/roles                - List available roles
GET    /api/v1/roles/{role}/permissions - Get role permissions
PUT    /api/v1/members/{id}/permissions - Update member permissions
GET    /api/v1/permissions          - List all permissions

# Tenant Analytics
GET    /api/v1/analytics            - Tenant analytics overview
GET    /api/v1/analytics/components - Component analytics
GET    /api/v1/analytics/incidents  - Incident analytics
GET    /api/v1/analytics/uptime     - Uptime analytics
GET    /api/v1/analytics/performance - Performance analytics
GET    /api/v1/analytics/subscribers - Subscriber analytics

# Billing & Subscription (Proxy to Payment Service)
GET    /api/v1/billing              - Billing information
GET    /api/v1/billing/subscription - Subscription details
GET    /api/v1/billing/invoices     - Invoice history
GET    /api/v1/billing/usage        - Usage metrics
POST   /api/v1/billing/upgrade      - Upgrade subscription

# Public API for Invitations
POST   /api/v1/public/invitations/accept/{token} - Accept invitation
GET    /api/v1/public/invitations/{token} - Get invitation details
```

**Role-Based Access Control (RBAC)**:
```go
type Permission string
const (
    // Component permissions
    PermissionComponentsView   Permission = "components:view"
    PermissionComponentsCreate Permission = "components:create"
    PermissionComponentsUpdate Permission = "components:update"
    PermissionComponentsDelete Permission = "components:delete"

    // Incident permissions
    PermissionIncidentsView    Permission = "incidents:view"
    PermissionIncidentsCreate  Permission = "incidents:create"
    PermissionIncidentsUpdate  Permission = "incidents:update"
    PermissionIncidentsDelete  Permission = "incidents:delete"

    // Monitor permissions
    PermissionMonitorsView     Permission = "monitors:view"
    PermissionMonitorsCreate   Permission = "monitors:create"
    PermissionMonitorsUpdate   Permission = "monitors:update"
    PermissionMonitorsDelete   Permission = "monitors:delete"

    // Team permissions
    PermissionTeamView         Permission = "team:view"
    PermissionTeamInvite       Permission = "team:invite"
    PermissionTeamUpdate       Permission = "team:update"
    PermissionTeamDelete       Permission = "team:delete"

    // Settings permissions
    PermissionSettingsView     Permission = "settings:view"
    PermissionSettingsUpdate   Permission = "settings:update"

    // Billing permissions
    PermissionBillingView      Permission = "billing:view"
    PermissionBillingUpdate    Permission = "billing:update"
)

var RolePermissions = map[TenantRole][]Permission{
    RoleOwner: {
        // All permissions
        PermissionComponentsView, PermissionComponentsCreate, PermissionComponentsUpdate, PermissionComponentsDelete,
        PermissionIncidentsView, PermissionIncidentsCreate, PermissionIncidentsUpdate, PermissionIncidentsDelete,
        PermissionMonitorsView, PermissionMonitorsCreate, PermissionMonitorsUpdate, PermissionMonitorsDelete,
        PermissionTeamView, PermissionTeamInvite, PermissionTeamUpdate, PermissionTeamDelete,
        PermissionSettingsView, PermissionSettingsUpdate,
        PermissionBillingView, PermissionBillingUpdate,
    },
    RoleAdmin: {
        // Most permissions except billing and team deletion
        PermissionComponentsView, PermissionComponentsCreate, PermissionComponentsUpdate, PermissionComponentsDelete,
        PermissionIncidentsView, PermissionIncidentsCreate, PermissionIncidentsUpdate, PermissionIncidentsDelete,
        PermissionMonitorsView, PermissionMonitorsCreate, PermissionMonitorsUpdate, PermissionMonitorsDelete,
        PermissionTeamView, PermissionTeamInvite, PermissionTeamUpdate,
        PermissionSettingsView, PermissionSettingsUpdate,
        PermissionBillingView,
    },
    RoleMember: {
        // Basic operational permissions
        PermissionComponentsView, PermissionComponentsUpdate,
        PermissionIncidentsView, PermissionIncidentsCreate, PermissionIncidentsUpdate,
        PermissionMonitorsView, PermissionMonitorsCreate, PermissionMonitorsUpdate,
        PermissionTeamView,
        PermissionSettingsView,
    },
    RoleViewer: {
        // Read-only permissions
        PermissionComponentsView,
        PermissionIncidentsView,
        PermissionMonitorsView,
        PermissionTeamView,
        PermissionSettingsView,
    },
}
```

**Tenant Context Middleware**:
```go
func TenantContextMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract tenant ID from JWT claims
        claims := jwt.GetClaims(c)
        tenantID := claims.TenantID

        // Set tenant context
        c.Set("tenant_id", tenantID)
        c.Set("user_id", claims.UserID)
        c.Set("user_role", claims.Role)

        // Add tenant ID to database queries
        if db, exists := c.Get("db"); exists {
            db.(*gorm.DB).Scopes(TenantScope(tenantID))
        }

        c.Next()
    }
}

func TenantScope(tenantID uint) func(*gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("tenant_id = ?", tenantID)
    }
}
```

---

### **10. SaaS Admin Service (Port 8098)**
**Purpose**: Platform-wide super administration for Beakon operators
**Database**: `saas_admin` - Platform-level data
**Multi-tenancy**: Manages multiple tenants from platform perspective

**Core Models**:
```go
type Tenant struct {
    ID               uint           `json:"id" gorm:"primarykey"`
    Name             string         `json:"name"`
    Email            string         `json:"email" gorm:"index"`
    Domain           string         `json:"domain" gorm:"uniqueIndex"`
    Subdomain        string         `json:"subdomain" gorm:"uniqueIndex"`
    Status           TenantStatus   `json:"status" gorm:"default:'active'"`
    PlanID           uint           `json:"plan_id" gorm:"index"`
    TrialEndsAt      *time.Time     `json:"trial_ends_at"`
    Settings         datatypes.JSON `json:"settings" gorm:"type:jsonb"`
    Metadata         datatypes.JSON `json:"metadata" gorm:"type:jsonb"`
    CreatedAt        time.Time      `json:"created_at"`
    UpdatedAt        time.Time      `json:"updated_at"`

    // Relationships
    Plan             *Plan          `json:"plan,omitempty" gorm:"foreignKey:PlanID"`
    Subscription     *Subscription  `json:"subscription,omitempty"`
}

type TenantStatus string
const (
    TenantActive    TenantStatus = "active"
    TenantSuspended TenantStatus = "suspended"
    TenantCancelled TenantStatus = "cancelled"
    TenantTrial     TenantStatus = "trial"
)

type PlatformPlan struct {
    ID                uint           `json:"id" gorm:"primarykey"`
    Name              string         `json:"name"`
    DisplayName       string         `json:"display_name"`
    Description       string         `json:"description"`
    Price             float64        `json:"price"`
    Currency          string         `json:"currency" gorm:"default:'usd'"`
    BillingInterval   string         `json:"billing_interval"` // monthly, yearly
    Features          datatypes.JSON `json:"features" gorm:"type:jsonb"`
    Limits            datatypes.JSON `json:"limits" gorm:"type:jsonb"`
    IsActive          bool           `json:"is_active" gorm:"default:true"`
    IsPublic          bool           `json:"is_public" gorm:"default:true"`
    SortOrder         int            `json:"sort_order" gorm:"default:0"`
    CreatedAt         time.Time      `json:"created_at"`
    UpdatedAt         time.Time      `json:"updated_at"`
}

type PlatformFeature struct {
    ID          uint      `json:"id" gorm:"primarykey"`
    Name        string    `json:"name" gorm:"uniqueIndex"`
    DisplayName string    `json:"display_name"`
    Description string    `json:"description"`
    Type        string    `json:"type"` // boolean, numeric, string
    DefaultValue interface{} `json:"default_value"`
    IsActive    bool      `json:"is_active" gorm:"default:true"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type FeatureFlag struct {
    ID          uint           `json:"id" gorm:"primarykey"`
    Name        string         `json:"name" gorm:"uniqueIndex"`
    DisplayName string         `json:"display_name"`
    Description string         `json:"description"`
    IsEnabled   bool           `json:"is_enabled" gorm:"default:false"`
    Rules       datatypes.JSON `json:"rules" gorm:"type:jsonb"`
    Rollout     datatypes.JSON `json:"rollout" gorm:"type:jsonb"`
    CreatedBy   uint           `json:"created_by"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
}

type PlatformAdmin struct {
    ID          uint      `json:"id" gorm:"primarykey"`
    Email       string    `json:"email" gorm:"uniqueIndex"`
    Name        string    `json:"name"`
    Role        AdminRole `json:"role" gorm:"default:'admin'"`
    IsActive    bool      `json:"is_active" gorm:"default:true"`
    LastLoginAt *time.Time `json:"last_login_at"`
    CreatedBy   uint      `json:"created_by"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type AdminRole string
const (
    AdminRoleSuperAdmin AdminRole = "super_admin"
    AdminRoleAdmin      AdminRole = "admin"
    AdminRoleSupport    AdminRole = "support"
    AdminRoleAnalyst    AdminRole = "analyst"
)

type PlatformStats struct {
    TotalTenants        int     `json:"total_tenants"`
    ActiveTenants       int     `json:"active_tenants"`
    TrialTenants        int     `json:"trial_tenants"`
    SuspendedTenants    int     `json:"suspended_tenants"`
    TotalRevenue        float64 `json:"total_revenue"`
    MonthlyRevenue      float64 `json:"monthly_revenue"`
    TotalComponents     int     `json:"total_components"`
    TotalMonitors       int     `json:"total_monitors"`
    TotalIncidents      int     `json:"total_incidents"`
    AverageUptime       float64 `json:"average_uptime"`
    NewSignupsToday     int     `json:"new_signups_today"`
    ChurnRate           float64 `json:"churn_rate"`
}
```

**API Endpoints**:
```
# Platform Overview
GET    /api/v1/platform             - Platform overview dashboard
GET    /api/v1/platform/stats       - Platform statistics
GET    /api/v1/platform/health      - Platform health status
GET    /api/v1/platform/metrics     - Real-time platform metrics

# Tenant Management
GET    /api/v1/tenants              - List all tenants
POST   /api/v1/tenants              - Create new tenant
GET    /api/v1/tenants/{id}         - Get tenant details
PUT    /api/v1/tenants/{id}         - Update tenant
DELETE /api/v1/tenants/{id}         - Delete tenant
POST   /api/v1/tenants/{id}/suspend - Suspend tenant
POST   /api/v1/tenants/{id}/activate - Activate tenant
GET    /api/v1/tenants/{id}/stats   - Tenant statistics
GET    /api/v1/tenants/{id}/activity - Tenant activity log

# Plan Management
GET    /api/v1/plans                - List platform plans
POST   /api/v1/plans                - Create plan
GET    /api/v1/plans/{id}           - Get plan details
PUT    /api/v1/plans/{id}           - Update plan
DELETE /api/v1/plans/{id}           - Delete plan
POST   /api/v1/plans/{id}/duplicate - Duplicate plan
GET    /api/v1/plans/usage          - Plan usage statistics

# Feature Management
GET    /api/v1/features             - List platform features
POST   /api/v1/features             - Create feature
GET    /api/v1/features/{id}        - Get feature details
PUT    /api/v1/features/{id}        - Update feature
DELETE /api/v1/features/{id}        - Delete feature
POST   /api/v1/features/{id}/toggle - Toggle feature

# Feature Flags
GET    /api/v1/feature-flags        - List feature flags
POST   /api/v1/feature-flags        - Create feature flag
GET    /api/v1/feature-flags/{id}   - Get feature flag
PUT    /api/v1/feature-flags/{id}   - Update feature flag
DELETE /api/v1/feature-flags/{id}   - Delete feature flag
POST   /api/v1/feature-flags/{id}/toggle - Toggle flag
POST   /api/v1/feature-flags/{id}/rollout - Update rollout

# Admin Management
GET    /api/v1/admin-users          - List platform admins
POST   /api/v1/admin-users          - Create admin user
GET    /api/v1/admin-users/{id}     - Get admin details
PUT    /api/v1/admin-users/{id}     - Update admin
DELETE /api/v1/admin-users/{id}     - Delete admin
POST   /api/v1/admin-users/{id}/activate - Activate admin
POST   /api/v1/admin-users/{id}/deactivate - Deactivate admin

# Analytics & Reporting
GET    /api/v1/analytics/overview   - Platform analytics overview
GET    /api/v1/analytics/revenue    - Revenue analytics
GET    /api/v1/analytics/usage      - Usage analytics
GET    /api/v1/analytics/churn      - Churn analysis
GET    /api/v1/analytics/growth     - Growth metrics
POST   /api/v1/reports              - Generate custom report
GET    /api/v1/reports/{id}         - Download report

# Notifications & Communications
GET    /api/v1/notifications        - Platform notifications
POST   /api/v1/notifications        - Send platform notification
POST   /api/v1/notifications/broadcast - Broadcast to all tenants
GET    /api/v1/notifications/templates - Notification templates

# Activity & Audit
GET    /api/v1/activities           - Platform activity feed
GET    /api/v1/activities/audit     - Audit log
POST   /api/v1/activities/search    - Search activities
GET    /api/v1/activities/export    - Export activity log

# Backup & Maintenance
GET    /api/v1/backups              - List platform backups
POST   /api/v1/backups              - Create backup
GET    /api/v1/backups/{id}         - Get backup details
DELETE /api/v1/backups/{id}         - Delete backup
POST   /api/v1/backups/{id}/restore - Restore backup
GET    /api/v1/maintenance          - Maintenance status
POST   /api/v1/maintenance          - Schedule maintenance
```

**Tenant Provisioning Workflow**:
```go
type TenantProvisioner struct {
    db             *gorm.DB
    userService    *UserService
    billingService *BillingService
    emailService   *EmailService
}

func (p *TenantProvisioner) CreateTenant(request CreateTenantRequest) (*Tenant, error) {
    // 1. Create tenant record
    tenant := &Tenant{
        Name:      request.CompanyName,
        Email:     request.Email,
        Domain:    request.Domain,
        Subdomain: request.Subdomain,
        Status:    TenantTrial,
        PlanID:    request.PlanID,
        TrialEndsAt: time.Now().Add(14 * 24 * time.Hour),
    }

    if err := p.db.Create(tenant).Error; err != nil {
        return nil, err
    }

    // 2. Create tenant admin user
    adminUser, err := p.userService.CreateUser(&CreateUserRequest{
        TenantID:  tenant.ID,
        Email:     request.Email,
        FirstName: request.FirstName,
        LastName:  request.LastName,
        Role:      "owner",
    })
    if err != nil {
        return nil, err
    }

    // 3. Initialize tenant resources
    if err := p.initializeTenantResources(tenant.ID); err != nil {
        return nil, err
    }

    // 4. Set up billing
    if err := p.billingService.CreateSubscription(tenant.ID, request.PlanID); err != nil {
        return nil, err
    }

    // 5. Send welcome email
    if err := p.emailService.SendWelcomeEmail(adminUser); err != nil {
        log.Printf("Failed to send welcome email: %v", err)
    }

    // 6. Trigger tenant created events
    p.publishTenantCreatedEvent(tenant)

    return tenant, nil
}
```

---

### **11. Status UI Service (Port 8093)**
**Purpose**: Public-facing status page rendering and customization
**Database**: Read-only access to tenant data
**Multi-tenancy**: Renders status pages for specific tenants

**Core Models**:
```go
type StatusPageConfig struct {
    TenantID        uint           `json:"tenant_id"`
    Title           string         `json:"title"`
    Description     string         `json:"description"`
    CustomDomain    string         `json:"custom_domain"`
    Theme           StatusPageTheme `json:"theme"`
    ShowUptime      bool           `json:"show_uptime"`
    ShowIncidents   bool           `json:"show_incidents"`
    ShowSubscribe   bool           `json:"show_subscribe"`
    ContactInfo     ContactInfo    `json:"contact_info"`
}

type StatusPageTheme struct {
    PrimaryColor     string `json:"primary_color"`
    BackgroundColor  string `json:"background_color"`
    TextColor        string `json:"text_color"`
    LinkColor        string `json:"link_color"`
    Logo             string `json:"logo"`
    Favicon          string `json:"favicon"`
    CustomCSS        string `json:"custom_css"`
}

type PublicStatusData struct {
    StatusPage   StatusPageConfig `json:"status_page"`
    OverallStatus string          `json:"overall_status"`
    Components   []PublicComponent `json:"components"`
    Incidents    []PublicIncident  `json:"incidents"`
    Maintenance  []PublicMaintenance `json:"maintenance"`
    Uptime       UptimeData       `json:"uptime"`
    LastUpdated  time.Time        `json:"last_updated"`
}
```

**API Endpoints**:
```
# Public Status API (No authentication)
GET    /                            - Status page root (HTML)
GET    /{subdomain}                 - Tenant-specific status page
GET    /api/v1/status               - Status API (JSON)
GET    /api/v1/status/summary       - Status summary
GET    /api/v1/status/components    - Components status
GET    /api/v1/status/incidents     - Current incidents
GET    /api/v1/status/history       - Status history
GET    /api/v1/status/uptime        - Uptime statistics
GET    /api/v1/status/uptime/{days} - Uptime for specific period

# Subscription API (Public)
POST   /api/v1/subscribe            - Subscribe to updates
GET    /api/v1/unsubscribe/{token}  - Unsubscribe from updates
POST   /api/v1/subscribe/verify     - Verify email subscription

# RSS/Atom Feeds
GET    /rss                         - RSS feed
GET    /rss.xml                     - RSS XML feed
GET    /atom.xml                    - Atom feed
GET    /incidents.rss               - Incidents RSS feed

# Embeddable Widgets
GET    /embed/status                - Embeddable status widget
GET    /embed/uptime                - Embeddable uptime widget
GET    /embed/incidents             - Embeddable incidents widget
GET    /api/v1/embed/status.js      - JavaScript widget
```

**Template Rendering System**:
```go
type StatusPageRenderer struct {
    templates *template.Template
    cache     cache.Cache
    db        *gorm.DB
}

func (r *StatusPageRenderer) RenderStatusPage(tenantID uint, domain string) ([]byte, error) {
    // Get cached data if available
    cacheKey := fmt.Sprintf("status_page:%d", tenantID)
    if cached := r.cache.Get(cacheKey); cached != nil {
        return cached.([]byte), nil
    }

    // Gather status data
    data, err := r.gatherStatusData(tenantID)
    if err != nil {
        return nil, err
    }

    // Render template
    var buf bytes.Buffer
    if err := r.templates.ExecuteTemplate(&buf, "status_page", data); err != nil {
        return nil, err
    }

    // Cache rendered page
    rendered := buf.Bytes()
    r.cache.Set(cacheKey, rendered, 5*time.Minute)

    return rendered, nil
}

func (r *StatusPageRenderer) gatherStatusData(tenantID uint) (*PublicStatusData, error) {
    var data PublicStatusData

    // Get tenant configuration
    config, err := r.getTenantConfig(tenantID)
    if err != nil {
        return nil, err
    }
    data.StatusPage = *config

    // Get components and status
    components, err := r.getPublicComponents(tenantID)
    if err != nil {
        return nil, err
    }
    data.Components = components
    data.OverallStatus = calculateOverallStatus(components)

    // Get active incidents
    incidents, err := r.getPublicIncidents(tenantID)
    if err != nil {
        return nil, err
    }
    data.Incidents = incidents

    // Get uptime data
    uptime, err := r.getUptimeData(tenantID)
    if err != nil {
        return nil, err
    }
    data.Uptime = *uptime

    data.LastUpdated = time.Now()
    return &data, nil
}
```

---

### **12. Consumer Services (4 Services)**

**Common Architecture Pattern**: All consumer services follow similar patterns for reliable event processing.

#### **Analytics Consumer**
**Purpose**: Process analytics events for business intelligence
**Technology**: Kafka consumer with batch processing
**Database**: Writes to `analytics_service` database

**Event Types**:
```go
type AnalyticsEvent struct {
    Type      string                 `json:"type"`
    TenantID  uint                   `json:"tenant_id"`
    UserID    uint                   `json:"user_id,omitempty"`
    Timestamp time.Time              `json:"timestamp"`
    Data      map[string]interface{} `json:"data"`
}

// Event types processed:
const (
    EventPageView          = "page_view"
    EventComponentView     = "component_view"
    EventIncidentView      = "incident_view"
    EventSubscription      = "subscription"
    EventStatusPageLoad    = "status_page_load"
    EventAPICall           = "api_call"
    EventUserAction        = "user_action"
)
```

**Processing Logic**:
```go
func (c *AnalyticsConsumer) ProcessEvent(event *AnalyticsEvent) error {
    switch event.Type {
    case EventPageView:
        return c.processPageView(event)
    case EventStatusPageLoad:
        return c.processStatusPageLoad(event)
    case EventAPICall:
        return c.processAPICall(event)
    default:
        return c.processGenericEvent(event)
    }
}

func (c *AnalyticsConsumer) processPageView(event *AnalyticsEvent) error {
    metric := &MetricDataPoint{
        MetricID:  c.getMetricID("page_views", event.TenantID),
        TenantID:  event.TenantID,
        Value:     1,
        Tags:      event.Data,
        Timestamp: event.Timestamp,
    }

    return c.db.Create(metric).Error
}
```

#### **Audit Consumer**
**Purpose**: Process audit logs for compliance and security
**Database**: Writes to `audit_logs` table

**Event Types**:
```go
type AuditEvent struct {
    Type        string    `json:"type"`
    TenantID    uint      `json:"tenant_id"`
    UserID      uint      `json:"user_id"`
    Action      string    `json:"action"`
    Resource    string    `json:"resource"`
    ResourceID  uint      `json:"resource_id"`
    IPAddress   string    `json:"ip_address"`
    UserAgent   string    `json:"user_agent"`
    Success     bool      `json:"success"`
    ErrorMessage string   `json:"error_message,omitempty"`
    Metadata    map[string]interface{} `json:"metadata"`
    Timestamp   time.Time `json:"timestamp"`
}

// Audit actions:
const (
    ActionCreate   = "create"
    ActionUpdate   = "update"
    ActionDelete   = "delete"
    ActionView     = "view"
    ActionLogin    = "login"
    ActionLogout   = "logout"
    ActionExport   = "export"
    ActionImport   = "import"
)
```

#### **Billing Consumer**
**Purpose**: Process billing events and usage tracking
**Database**: Writes to `payment_service` and `usage` tables

**Event Types**:
```go
type BillingEvent struct {
    Type       string    `json:"type"`
    TenantID   uint      `json:"tenant_id"`
    MetricName string    `json:"metric_name"`
    Value      float64   `json:"value"`
    Period     string    `json:"period"`
    Timestamp  time.Time `json:"timestamp"`
}

// Usage metrics tracked:
const (
    MetricMonitorChecks    = "monitor_checks"
    MetricAPIRequests      = "api_requests"
    MetricDataStorage      = "data_storage"
    MetricNotifications    = "notifications_sent"
    MetricTeamMembers      = "team_members"
    MetricComponents       = "components"
)
```

#### **Notification Consumer**
**Purpose**: Deliver notifications via multiple channels
**Database**: Updates `notification_service` delivery status

**Processing Logic**:
```go
type NotificationDelivery struct {
    NotificationID uint                   `json:"notification_id"`
    TenantID       uint                   `json:"tenant_id"`
    Channel        string                 `json:"channel"`
    Recipient      string                 `json:"recipient"`
    Message        map[string]interface{} `json:"message"`
    Priority       string                 `json:"priority"`
    MaxRetries     int                    `json:"max_retries"`
}

func (c *NotificationConsumer) ProcessDelivery(delivery *NotificationDelivery) error {
    provider := c.getChannelProvider(delivery.Channel)

    err := provider.Send(delivery)
    if err != nil {
        if delivery.MaxRetries > 0 {
            // Retry with exponential backoff
            return c.scheduleRetry(delivery, err)
        } else {
            // Mark as failed
            return c.markAsFailed(delivery, err)
        }
    }

    // Mark as delivered
    return c.markAsDelivered(delivery)
}
```

---

## 🔧 **Critical Issues Identified & Solutions**

### **1. API Gateway Port Mismatches**
**Status**: 🚨 Critical - Service communication broken

| Service | Expected | Actual | Fix Required |
|---------|----------|--------|-------------|
| tenant-admin-service | 8082 | 8099 | Update gateway config |
| component-service | 8083 | 8084 | Update gateway config |
| incident-service | 8084 | 8086 | Update gateway config |
| payment-service | 8086 | 8088 | Update gateway config |
| analytics-service | 8087 | 8090 | Update gateway config |
| monitoring-service | 8088 | 8092 | Update gateway config |
| branding-service | 8089 | 8097 | Update gateway config |
| database-service | 8090 | 8095 | Update gateway config |
| event-store-service | 8091 | 8096 | Update gateway config |
| landing-page-service | 8092 | 8100 | Update gateway config |

**Solution**:
```go
// Update api-gateway/internal/config/config.go
UserService:         "http://localhost:8081", // ✅ Correct
TenantService:       "http://localhost:8099", // Fixed
ComponentService:    "http://localhost:8084", // Fixed
IncidentService:     "http://localhost:8086", // Fixed
PaymentService:      "http://localhost:8088", // Fixed
AnalyticsService:    "http://localhost:8090", // Fixed
NotificationService: "http://localhost:8085", // ✅ Correct
MonitoringService:   "http://localhost:8092", // Fixed
```

### **2. Database Schema Inconsistencies**
**Status**: ⚠️ High - Data integrity concerns

**Issues Found**:
- Inconsistent `tenant_id` indexing across services
- Missing foreign key constraints in some relationships
- Inconsistent soft delete implementation
- Missing database migrations for some models

**Solution**:
```sql
-- Standardize tenant_id indexing
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tablename_tenant_id ON table_name(tenant_id);

-- Add missing foreign key constraints
ALTER TABLE incidents ADD CONSTRAINT fk_incidents_tenant
    FOREIGN KEY (tenant_id) REFERENCES tenants(id);

-- Standardize soft delete
ALTER TABLE components ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_components_deleted_at ON components(deleted_at);
```

### **3. Authentication & Authorization Gaps**
**Status**: ⚠️ High - Security concerns

**Issues Found**:
- Inconsistent JWT claim validation
- Missing rate limiting on authentication endpoints
- No session invalidation on role changes
- Missing API key authentication for service-to-service calls

### **4. Missing Real-time Capabilities**
**Status**: 🟡 Medium - Feature gaps

**Missing Features**:
- WebSocket connections for live status updates
- Server-sent events for incident updates
- Real-time dashboard data
- Live notification delivery status

---

## 📊 **Performance & Scalability Considerations**

### **Database Optimization**
```sql
-- Time-series data partitioning for monitor results
CREATE TABLE monitor_results_y2024m01 PARTITION OF monitor_results
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- Optimized indexes for tenant queries
CREATE INDEX CONCURRENTLY idx_components_tenant_status
    ON components(tenant_id, status) WHERE deleted_at IS NULL;

-- Analytics query optimization
CREATE MATERIALIZED VIEW tenant_stats_hourly AS
SELECT
    tenant_id,
    date_trunc('hour', created_at) as hour,
    count(*) as event_count,
    avg(response_time) as avg_response_time
FROM monitor_results
GROUP BY tenant_id, date_trunc('hour', created_at);
```

### **Caching Strategy**
```go
// Redis caching patterns
type CacheManager struct {
    client *redis.Client
    ttl    map[string]time.Duration
}

// Cache keys by data type
const (
    CacheKeyStatusPage     = "status_page:%d"          // 5 minutes
    CacheKeyComponents     = "components:%d"           // 2 minutes
    CacheKeyIncidents      = "incidents:%d"            // 1 minute
    CacheKeyUptimeStats    = "uptime:%d:%s"           // 15 minutes
    CacheKeyTenantSettings = "tenant_settings:%d"     // 30 minutes
)
```

### **Rate Limiting**
```go
// Per-tenant rate limiting
type RateLimiter struct {
    redis  *redis.Client
    limits map[string]int // endpoint -> requests per minute
}

// Rate limit by tenant and endpoint
func (r *RateLimiter) Allow(tenantID uint, endpoint string) bool {
    key := fmt.Sprintf("rate_limit:%d:%s", tenantID, endpoint)
    limit := r.limits[endpoint]

    current, err := r.redis.Incr(key).Result()
    if err != nil {
        return true // Allow on Redis errors
    }

    if current == 1 {
        r.redis.Expire(key, time.Minute)
    }

    return current <= int64(limit)
}
```

---

## 🧪 **Testing Strategy**

### **Service Testing Matrix**

| Service | Unit Tests | Integration Tests | E2E Tests | Load Tests |
|---------|------------|-------------------|-----------|------------|
| api-gateway | ✅ | ✅ | ✅ | ✅ |
| user-service | ✅ | ✅ | ✅ | ❌ |
| component-service | ✅ | ✅ | ✅ | ❌ |
| incident-service | ✅ | ✅ | ✅ | ❌ |
| monitoring-service | ✅ | ✅ | ✅ | ✅ |
| analytics-service | ✅ | ✅ | ❌ | ❌ |
| notification-service | ✅ | ✅ | ❌ | ❌ |
| payment-service | ✅ | ✅ | ✅ | ❌ |
| tenant-admin-service | ✅ | ✅ | ❌ | ❌ |
| saas-admin-service | ✅ | ✅ | ❌ | ❌ |

---

## 🛠️ **Development Guidelines**

### **Adding New Microservices**

1. **Use Service Template**: Copy from `SERVICE_TEMPLATE_main.go`
2. **Follow Naming Convention**: `service-name-service`
3. **Implement Standard Structure**:
   ```
   service-name/
   ├── cmd/main.go
   ├── internal/
   │   ├── models/
   │   ├── handlers/
   │   ├── services/
   │   └── config/
   ├── api/service-api.yaml
   └── .env.example
   ```
4. **Add to API Gateway**: Update routing configuration
5. **Add Database Migration**: Include schema migration
6. **Update Documentation**: Add to this reference document

### **Multi-tenancy Checklist**

- [ ] All models include `TenantID uint` with proper indexing
- [ ] Tenant middleware enforces data isolation
- [ ] Database queries scoped by tenant
- [ ] API endpoints validate tenant access
- [ ] Audit logging includes tenant context
- [ ] Rate limiting applied per tenant
- [ ] Cache keys include tenant ID

### **Security Checklist**

- [ ] JWT authentication implemented
- [ ] Input validation and sanitization
- [ ] SQL injection protection via ORM
- [ ] XSS protection in templates
- [ ] CORS properly configured
- [ ] Rate limiting implemented
- [ ] Audit logging for sensitive operations
- [ ] Secrets managed via environment variables

---

## 📈 **Monitoring & Observability**

### **Health Check Implementation**
```go
type HealthChecker struct {
    db    *gorm.DB
    redis *redis.Client
    deps  []HealthCheckable
}

type HealthStatus struct {
    Service     string            `json:"service"`
    Status      string            `json:"status"`
    Version     string            `json:"version"`
    Checks      map[string]string `json:"checks"`
    Timestamp   time.Time         `json:"timestamp"`
    ResponseTime time.Duration    `json:"response_time"`
}

func (h *HealthChecker) Check() *HealthStatus {
    start := time.Now()
    status := &HealthStatus{
        Service:   h.serviceName,
        Version:   h.version,
        Timestamp: start,
        Checks:    make(map[string]string),
    }

    // Check database
    if err := h.db.DB(); err != nil {
        status.Checks["database"] = "unhealthy: " + err.Error()
    } else {
        status.Checks["database"] = "healthy"
    }

    // Check Redis
    if err := h.redis.Ping().Err(); err != nil {
        status.Checks["redis"] = "unhealthy: " + err.Error()
    } else {
        status.Checks["redis"] = "healthy"
    }

    // Determine overall status
    status.Status = "healthy"
    for _, check := range status.Checks {
        if strings.Contains(check, "unhealthy") {
            status.Status = "unhealthy"
            break
        }
    }

    status.ResponseTime = time.Since(start)
    return status
}
```

### **Metrics Collection**
```go
// Prometheus metrics
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"service", "method", "endpoint", "status"},
    )

    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration",
        },
        []string{"service", "method", "endpoint"},
    )

    tenantOperationsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "tenant_operations_total",
            Help: "Total tenant operations",
        },
        []string{"tenant_id", "operation", "service"},
    )
)
```

---

## 🎯 **Next Steps & Roadmap**

### **Immediate Priorities (Week 1-2)**
1. **Fix API Gateway routing** - Update port configurations
2. **Start core services** - Get monitoring and component services running
3. **Fix database schemas** - Resolve inconsistencies
4. **Update documentation** - Ensure accuracy

### **Short-term Goals (Month 1)**
1. **Real-time capabilities** - WebSocket implementation
2. **Service discovery** - Dynamic service registration
3. **Advanced monitoring** - Multi-region health checks
4. **Enhanced security** - API keys and improved auth

### **Medium-term Goals (Months 2-3)**
1. **Advanced analytics** - Business intelligence dashboards
2. **Integration marketplace** - Third-party service integrations
3. **Multi-region deployment** - Global monitoring locations
4. **Performance optimization** - Caching and database tuning

### **Long-term Vision (Months 4-6)**
1. **AI-powered insights** - Predictive analytics and anomaly detection
2. **Advanced workflow automation** - Intelligent incident response
3. **Enterprise features** - SSO, SCIM, advanced RBAC
4. **Mobile applications** - iOS/Android apps for status management

---

*This comprehensive reference document covers all 20+ microservices in the Beakon platform with detailed technical analysis, architectural patterns, and implementation guidance. Last updated: 2025-09-25*