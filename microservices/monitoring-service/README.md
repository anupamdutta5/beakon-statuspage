# Monitoring Service

## Overview

The Monitoring Service is a comprehensive system health and performance monitoring platform that provides active health checking, alert management, maintenance scheduling, and third-party integrations. It's the central nervous system of the Beakon platform, continuously monitoring services, collecting metrics, and triggering alerts when issues are detected.

**Service Port**: 8091

## Core Responsibilities

### 1. Active Health Monitoring
- Perform active health checks on registered services and components
- Track uptime statistics and availability metrics
- Monitor service response times and performance indicators
- Detect and report service degradation or outages

### 2. Alert Management
- Create and manage alert rules based on monitoring data
- Trigger alerts when thresholds are exceeded
- Support alert acknowledgment and resolution workflows
- Provide alert history and analytics

### 3. Maintenance Window Management
- Schedule and manage planned maintenance windows
- Automatically suppress alerts during maintenance periods
- Track maintenance updates and communications
- Support maintenance templates for recurring activities
- Associate components with maintenance windows

### 4. Webhook & Integration Management
- Send webhook notifications for monitoring events
- Integrate with third-party monitoring tools (Datadog, New Relic, etc.)
- Sync component status with external monitoring systems
- Track webhook delivery status and retry failed deliveries

### 5. Performance Metrics Collection
- Collect and store performance metrics
- Provide historical performance data
- Support custom metric definitions
- Aggregate metrics for reporting

### 6. Uptime Tracking
- Monitor endpoint availability
- Calculate uptime percentages
- Provide uptime SLA reports
- Track incident impact on uptime

## Architecture

### Data Models

#### Service Monitoring
```go
type Service struct {
    ID          uint      `json:"id"`
    Name        string    `json:"name"`
    URL         string    `json:"url"`
    Type        string    `json:"type"`        // http, tcp, ping
    Status      string    `json:"status"`      // up, down, degraded
    LastChecked time.Time `json:"last_checked"`
}

type HealthCheck struct {
    ID              uint   `json:"id"`
    ServiceID       uint   `json:"service_id"`
    CheckType       string `json:"check_type"`   // http, tcp, ping, custom
    Endpoint        string `json:"endpoint"`
    Interval        int    `json:"interval"`     // seconds
    Timeout         int    `json:"timeout"`      // seconds
    ExpectedStatus  int    `json:"expected_status"`
}
```

#### Alert Management
```go
type Alert struct {
    ID          uint      `json:"id"`
    ServiceID   uint      `json:"service_id"`
    Severity    string    `json:"severity"`    // critical, warning, info
    Status      string    `json:"status"`      // open, acknowledged, resolved
    Message     string    `json:"message"`
    CreatedAt   time.Time `json:"created_at"`
    AcknowledgedAt *time.Time `json:"acknowledged_at"`
    ResolvedAt     *time.Time `json:"resolved_at"`
}
```

#### Maintenance Management
```go
type MaintenanceWindow struct {
    ID          uint      `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    StartTime   time.Time `json:"start_time"`
    EndTime     time.Time `json:"end_time"`
    Status      string    `json:"status"`      // scheduled, in_progress, completed, cancelled
    Components  []uint    `json:"components"`  // Component IDs affected
}

type MaintenanceUpdate struct {
    ID                 uint      `json:"id"`
    MaintenanceWindowID uint     `json:"maintenance_window_id"`
    Message            string    `json:"message"`
    Status             string    `json:"status"`
    CreatedAt          time.Time `json:"created_at"`
}
```

#### Webhook Management
```go
type WebhookEndpoint struct {
    ID          uint      `json:"id"`
    URL         string    `json:"url"`
    EventTypes  []string  `json:"event_types"`  // alert.created, service.down, etc.
    IsActive    bool      `json:"is_active"`
    Secret      string    `json:"secret"`
}

type WebhookDelivery struct {
    ID              uint      `json:"id"`
    WebhookEndpointID uint    `json:"webhook_endpoint_id"`
    EventType       string    `json:"event_type"`
    Payload         string    `json:"payload"`
    Status          string    `json:"status"`      // pending, success, failed
    ResponseCode    int       `json:"response_code"`
    AttemptCount    int       `json:"attempt_count"`
}
```

#### Third-Party Integrations
```go
type Integration struct {
    ID          uint      `json:"id"`
    Type        string    `json:"type"`        // datadog, newrelic, prometheus, etc.
    Name        string    `json:"name"`
    Config      string    `json:"config"`      // JSON configuration
    IsActive    bool      `json:"is_active"`
    LastSyncAt  *time.Time `json:"last_sync_at"`
}

type ComponentMapping struct {
    ID            uint   `json:"id"`
    IntegrationID uint   `json:"integration_id"`
    ComponentID   uint   `json:"component_id"`
    ExternalID    string `json:"external_id"`
}
```

## API Endpoints

### Health & Status
- `GET /health` - Service health check
- `GET /metrics` - Service metrics (Prometheus format)
- `GET /api/v1/public/status` - Public service status
- `GET /api/v1/public/health` - Public health status

### Monitoring Overview
- `GET /api/v1/monitoring/overview` - Complete monitoring dashboard data

### Service Management
- `GET /api/v1/services` - List all monitored services
- `POST /api/v1/services` - Register new service for monitoring
- `GET /api/v1/services/:id` - Get service details
- `PUT /api/v1/services/:id` - Update service configuration
- `DELETE /api/v1/services/:id` - Remove service from monitoring
- `GET /api/v1/services/:id/health` - Get service health status
- `GET /api/v1/services/:id/metrics` - Get service metrics

### Health Check Management
- `POST /api/v1/services/:id/health-checks` - Create health check
- `PUT /api/v1/services/health-checks/:check_id` - Update health check
- `DELETE /api/v1/services/health-checks/:check_id` - Delete health check

### Alert Management
- `GET /api/v1/alerts` - List alerts
- `POST /api/v1/alerts` - Create alert rule
- `GET /api/v1/alerts/:id` - Get alert details
- `PUT /api/v1/alerts/:id` - Update alert
- `DELETE /api/v1/alerts/:id` - Delete alert
- `POST /api/v1/alerts/:id/acknowledge` - Acknowledge alert
- `POST /api/v1/alerts/:id/resolve` - Resolve alert

### Maintenance Window Management
- `GET /api/v1/maintenance/windows` - List maintenance windows
- `POST /api/v1/maintenance/windows` - Create maintenance window
- `GET /api/v1/maintenance/windows/:id` - Get maintenance window details
- `PUT /api/v1/maintenance/windows/:id` - Update maintenance window
- `DELETE /api/v1/maintenance/windows/:id` - Delete maintenance window
- `POST /api/v1/maintenance/windows/:id/start` - Start maintenance
- `POST /api/v1/maintenance/windows/:id/complete` - Complete maintenance
- `POST /api/v1/maintenance/windows/:id/cancel` - Cancel maintenance

### Maintenance Updates
- `POST /api/v1/maintenance/windows/:id/updates` - Post maintenance update
- `GET /api/v1/maintenance/windows/:id/updates` - Get maintenance updates

### Maintenance Components
- `POST /api/v1/maintenance/windows/:id/components` - Add component to maintenance
- `DELETE /api/v1/maintenance/windows/:id/components/:component_id` - Remove component

### Maintenance Queries
- `GET /api/v1/maintenance/upcoming` - Get upcoming maintenance
- `GET /api/v1/maintenance/active` - Get active maintenance
- `GET /api/v1/maintenance/statistics` - Get maintenance statistics

### Maintenance Templates
- `GET /api/v1/maintenance/templates` - List maintenance templates
- `POST /api/v1/maintenance/templates` - Create maintenance template
- `POST /api/v1/maintenance/templates/:template_id/create-maintenance` - Create from template

### Uptime Monitoring
- `GET /api/v1/uptime/checks` - List uptime checks
- `POST /api/v1/uptime/checks` - Create uptime check
- `GET /api/v1/uptime/checks/:id` - Get uptime check
- `PUT /api/v1/uptime/checks/:id` - Update uptime check
- `DELETE /api/v1/uptime/checks/:id` - Delete uptime check
- `GET /api/v1/uptime/results` - Get uptime results
- `GET /api/v1/uptime/statistics` - Get uptime statistics

### Performance Metrics
- `GET /api/v1/performance/metrics` - List performance metrics
- `POST /api/v1/performance/metrics` - Create metric definition
- `GET /api/v1/performance/metrics/:id` - Get metric details
- `PUT /api/v1/performance/metrics/:id` - Update metric
- `DELETE /api/v1/performance/metrics/:id` - Delete metric
- `GET /api/v1/performance/data` - Query performance data
- `POST /api/v1/performance/data` - Add performance data point

### Logs
- `GET /api/v1/logs` - Get logs
- `GET /api/v1/logs/search` - Search logs
- `GET /api/v1/logs/aggregate` - Aggregate logs
- `GET /api/v1/logs/stream` - Stream logs (WebSocket)

### Webhooks
- `GET /api/v1/webhooks` - List webhook endpoints
- `POST /api/v1/webhooks` - Create webhook endpoint
- `GET /api/v1/webhooks/:id` - Get webhook details
- `PUT /api/v1/webhooks/:id` - Update webhook
- `DELETE /api/v1/webhooks/:id` - Delete webhook
- `POST /api/v1/webhooks/:id/test` - Test webhook endpoint
- `GET /api/v1/webhooks/deliveries` - List webhook deliveries
- `POST /api/v1/webhooks/deliveries/:id/retry` - Retry failed delivery
- `GET /api/v1/webhooks/statistics` - Webhook statistics
- `GET /api/v1/webhooks/event-types` - List supported event types

### Third-Party Integrations
- `GET /api/v1/integrations` - List integrations
- `POST /api/v1/integrations` - Create integration
- `GET /api/v1/integrations/:id` - Get integration details
- `PUT /api/v1/integrations/:id` - Update integration
- `DELETE /api/v1/integrations/:id` - Delete integration
- `POST /api/v1/integrations/:id/sync` - Trigger integration sync
- `POST /api/v1/integrations/:id/test` - Test integration
- `POST /api/v1/integrations/:id/mappings` - Create component mapping
- `GET /api/v1/integrations/:id/mappings` - Get component mappings
- `GET /api/v1/integrations/:id/logs` - Get integration sync logs
- `GET /api/v1/integrations/supported` - List supported integration types

## Dependencies

### Internal Services
- **Component Service**: Monitors components registered in component service
- **Event Store Service**: Publishes monitoring events (alerts, status changes)
- **Notification Service**: Sends notifications when alerts are triggered

### External Dependencies
- PostgreSQL database for storing monitoring data
- Third-party monitoring platforms (optional):
  - Datadog
  - New Relic
  - Prometheus
  - PagerDuty
  - Grafana

## Environment Variables

```bash
# Server Configuration
SERVER_PORT=8091
SERVER_HOST=0.0.0.0
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=15s
SERVER_IDLE_TIMEOUT=60s
SERVER_GRACEFUL_STOP=10s

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=monitoring_db
DB_USER=postgres
DB_PASSWORD=your_password
DB_SSL_MODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# Monitoring Configuration
MONITORING_ENABLED=true
HEALTH_CHECK_INTERVAL=30s

# Environment
ENVIRONMENT=development  # development, staging, production
```

## How to Run

### Prerequisites
- Go 1.21 or higher
- PostgreSQL 14 or higher

### Local Development

1. **Set up database**:
```bash
createdb monitoring_db
```

2. **Set environment variables**:
```bash
export DB_PASSWORD="postgres"
export DB_NAME="monitoring_db"
export SERVER_PORT="8091"
```

3. **Build the service**:
```bash
go build -o monitoring-service cmd/main.go
```

4. **Run the service**:
```bash
./monitoring-service
```

### Using Docker

```bash
docker build -t monitoring-service .
docker run -p 8091:8091 \
  -e DB_HOST=postgres \
  -e DB_PASSWORD=postgres \
  -e DB_NAME=monitoring_db \
  monitoring-service
```

## Testing

### Unit Tests
```bash
go test ./tests/unit/...
```

### Integration Tests
```bash
go test ./tests/integration/...
```

## Key Features

### 1. Active Health Checking
- Continuously monitors registered services
- Supports multiple check types: HTTP, TCP, ICMP ping
- Configurable check intervals and timeouts
- Automatic retry logic with exponential backoff

### 2. Intelligent Alerting
- Rule-based alert generation
- Multiple severity levels (critical, warning, info)
- Alert deduplication to prevent notification spam
- Alert acknowledgment workflow
- Automatic alert resolution

### 3. Maintenance Window Support
- Schedule planned maintenance in advance
- Automatic alert suppression during maintenance
- Post updates to keep stakeholders informed
- Template support for recurring maintenance
- Component association for targeted suppression

### 4. Webhook Integration
- Real-time event notifications via webhooks
- Configurable event types per endpoint
- Automatic retry with exponential backoff
- Webhook signature verification
- Delivery tracking and monitoring

### 5. Third-Party Platform Integration
- Bidirectional sync with external monitoring tools
- Component mapping for unified monitoring
- Automatic status propagation
- Integration health monitoring

## Use Cases

### 1. Service Health Monitoring
Monitor all microservices in the Beakon platform:
```json
POST /api/v1/services
{
  "name": "User Service",
  "url": "http://user-service:8080",
  "type": "http"
}

POST /api/v1/services/1/health-checks
{
  "check_type": "http",
  "endpoint": "/health",
  "interval": 30,
  "timeout": 5,
  "expected_status": 200
}
```

### 2. Alert on Service Degradation
Automatically create alerts when issues are detected:
```json
POST /api/v1/alerts
{
  "service_id": 1,
  "severity": "critical",
  "condition": "response_time > 1000ms",
  "message": "User Service response time exceeded threshold"
}
```

### 3. Schedule Maintenance Window
Plan maintenance with automatic alert suppression:
```json
POST /api/v1/maintenance/windows
{
  "title": "Database Upgrade",
  "description": "Upgrading PostgreSQL to version 15",
  "start_time": "2025-10-10T02:00:00Z",
  "end_time": "2025-10-10T04:00:00Z",
  "components": [1, 2, 3]
}
```

### 4. Webhook Notifications
Send real-time notifications to external systems:
```json
POST /api/v1/webhooks
{
  "url": "https://your-app.com/webhooks/monitoring",
  "event_types": ["alert.created", "service.down", "maintenance.started"],
  "secret": "webhook_secret_key"
}
```

### 5. Datadog Integration
Sync monitoring data with Datadog:
```json
POST /api/v1/integrations
{
  "type": "datadog",
  "name": "Production Datadog",
  "config": {
    "api_key": "datadog_api_key",
    "app_key": "datadog_app_key",
    "site": "datadoghq.com"
  },
  "is_active": true
}
```

## Monitoring vs Component Service

| Feature | Monitoring Service | Component Service |
|---------|-------------------|-------------------|
| **Purpose** | Active monitoring & alerting | Passive status storage |
| **Health Checks** | Performs active checks | Stores status updates |
| **Metrics** | Collects performance data | Stores component metadata |
| **Alerts** | Generates and manages alerts | None |
| **Maintenance** | Schedules maintenance windows | None |
| **Integrations** | Syncs with external tools | None |

**Analogy**: Component Service is like a filing cabinet (stores information), while Monitoring Service is like a security guard (actively watches and responds to issues).

## Performance Considerations

- Health checks run asynchronously in background workers
- Database queries use connection pooling (25 max connections)
- Alert deduplication prevents notification storms
- Webhook deliveries use exponential backoff for retries
- Metrics are aggregated to reduce storage requirements

## Security

- All API endpoints require authentication (JWT tokens)
- Public endpoints are rate-limited
- Webhook payloads are signed with HMAC-SHA256
- Integration credentials are encrypted at rest
- Database connections use SSL in production

## Troubleshooting

### Service Not Starting
- Check database connection parameters
- Verify PostgreSQL is running and accessible
- Check logs for configuration validation errors

### Health Checks Not Running
- Verify `MONITORING_ENABLED=true`
- Check service registration in database
- Review health check configuration (interval, timeout)

### Webhooks Not Delivering
- Verify webhook endpoint URL is accessible
- Check webhook delivery logs
- Review retry attempts and error messages

### Integration Sync Failures
- Validate integration credentials
- Check external service availability
- Review integration sync logs

## Future Enhancements

- Machine learning-based anomaly detection
- Predictive alerting based on trends
- Custom dashboard builder
- Multi-region monitoring
- SLA tracking and reporting
- Incident timeline reconstruction
- Automated remediation actions
