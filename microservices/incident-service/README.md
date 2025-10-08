# Incident Service

## Overview
The Incident Service manages incident lifecycle, incident updates, and incident workflow automation for the Beakon status page platform. It handles creation, tracking, and resolution of incidents affecting status pages, including automated workflows and communication templates.

## Key Features

### Incident Management
- **Incident CRUD Operations**: Create, read, update, delete incidents
- **Incident Lifecycle**: Investigate, identified, monitoring, resolved
- **Impact Levels**: Critical, major, minor, maintenance
- **Incident Updates**: Timeline of incident status updates
- **Component Affectation**: Track which components are affected

### Incident Templates
- **Template Management**: Reusable incident templates for common scenarios
- **Template Cloning**: Duplicate templates for customization
- **Create from Template**: Initialize incidents from templates
- **Category Management**: Organize templates by category

### Workflow Automation
- **Workflow Steps**: Define automated incident response workflows
- **Step Types**: Manual, notification, API call, delay
- **Step Execution**: Automatic and manual step execution
- **Step Reordering**: Flexible workflow step ordering
- **Execution Tracking**: Monitor workflow progress

### Public API
- **Public Incidents**: Publicly visible incident information
- **Status Page Integration**: Feed incidents to public status pages
- **Historical Data**: Access to past incidents

### Advanced Features
- **Circuit Breaker**: Database and external service protection
- **Connection Pooling**: Optimized database connections
- **Caching**: Redis/in-memory caching for performance
- **Rate Limiting**: API rate limiting for fair usage
- **Health Monitoring**: Database and service health checks

## API Endpoints

### Health & Monitoring
- `GET /health` - Health check

### Public Incident API (No Authentication)
- `GET /api/v1/public/incidents` - List public incidents
- `GET /api/v1/public/incidents/:id` - Get public incident details

### Protected Incident Routes

#### Incident Management
- `GET /api/v1/incidents` - List all incidents
- `POST /api/v1/incidents` - Create new incident
- `GET /api/v1/incidents/:id` - Get incident details
- `PUT /api/v1/incidents/:id` - Update incident
- `DELETE /api/v1/incidents/:id` - Delete incident

#### Incident Updates
- `POST /api/v1/incidents/:id/updates` - Add incident update
- `PUT /api/v1/incidents/:id/updates/:update_id` - Update incident update
- `DELETE /api/v1/incidents/:id/updates/:update_id` - Delete incident update

#### Incident Templates
- `GET /api/v1/templates` - List templates
- `POST /api/v1/templates` - Create template
- `GET /api/v1/templates/:id` - Get template details
- `PUT /api/v1/templates/:id` - Update template
- `DELETE /api/v1/templates/:id` - Delete template
- `POST /api/v1/templates/:id/clone` - Clone template
- `POST /api/v1/templates/:id/create-incident` - Create incident from template

#### Workflow Steps
- `POST /api/v1/templates/:template_id/steps` - Create workflow step
- `PUT /api/v1/templates/steps/:step_id` - Update workflow step
- `DELETE /api/v1/templates/steps/:step_id` - Delete workflow step
- `PUT /api/v1/templates/:template_id/steps/reorder` - Reorder workflow steps

#### Workflow Execution
- `GET /api/v1/workflows/incidents/:incident_id/executions` - Get step executions
- `POST /api/v1/workflows/executions/:execution_id/execute` - Execute manual step

#### Administrative Functions
- `POST /api/v1/admin/initialize-templates` - Initialize default templates

## Dependencies

### Internal Services
- **component-service**: Component status integration
- **notification-service**: Incident notification delivery
- **tenant-admin-service**: Tenant and user information

### External Dependencies
- **shared-resilience**: Common patterns, database, middleware
- **PostgreSQL**: Incident data storage
- **Redis**: Caching and rate limiting (optional)
- **Gin**: HTTP web framework

## Configuration

### Environment Variables
- `SERVER_PORT`: HTTP server port (default: 8080)
- `SERVER_HOST`: HTTP server host
- `ENVIRONMENT`: Runtime environment (development/production)
- `DB_HOST`: Database host
- `DB_PORT`: Database port
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `JWT_SECRET`: JWT signing secret
- `REDIS_HOST`: Redis host (optional)
- `REDIS_PORT`: Redis port (optional)
- `COMPONENT_SERVICE_URL`: Component service endpoint
- `NOTIFICATION_SERVICE_URL`: Notification service endpoint

### Feature Flags
- `CIRCUIT_BREAKER_ENABLED`: Enable circuit breaker protection
- `CACHE_ENABLED`: Enable caching layer
- `RATE_LIMIT_ENABLED`: Enable API rate limiting
- `MONITORING_ENABLED`: Enable health monitoring

## Development

### Running the Service
```bash
cd microservices/incident-service
go run cmd/main.go
```

### Building
```bash
go build -o incident-service cmd/main.go
```

### Testing
```bash
go test ./...
```

### Project Structure
```
incident-service/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── handlers/            # HTTP request handlers
│   ├── services/            # Business logic layer
│   │   ├── incident_service.go
│   │   └── template_service.go
│   ├── models/              # Data models
│   └── workflow/            # Workflow execution engine
└── go.mod                   # Go module definition
```

## Incident Schema

### Incident Model
```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "title": "Database Connection Issues",
  "description": "Users experiencing connection timeouts",
  "status": "investigating|identified|monitoring|resolved",
  "impact": "critical|major|minor|maintenance",
  "affected_components": ["component_uuid"],
  "started_at": "ISO8601",
  "resolved_at": "ISO8601",
  "created_by": "user_uuid",
  "updates": [
    {
      "id": "uuid",
      "status": "investigating",
      "message": "We are investigating the issue",
      "created_at": "ISO8601",
      "created_by": "user_uuid"
    }
  ]
}
```

### Template Model
```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "name": "Database Outage Template",
  "category": "database",
  "title_template": "Database {{severity}} - {{location}}",
  "description_template": "Database experiencing {{issue_type}}",
  "impact": "critical",
  "workflow_steps": [
    {
      "id": "uuid",
      "step_number": 1,
      "step_type": "notification",
      "name": "Notify on-call engineer",
      "config": {"channel": "pagerduty"},
      "is_automated": true
    }
  ]
}
```

## Incident Lifecycle

### Status Flow
1. **Investigating**: Initial incident creation, investigating root cause
2. **Identified**: Root cause identified, working on fix
3. **Monitoring**: Fix deployed, monitoring for stability
4. **Resolved**: Incident fully resolved, post-mortem complete

### Impact Levels
- **Critical**: Complete service outage
- **Major**: Significant degradation affecting many users
- **Minor**: Minor issues affecting few users
- **Maintenance**: Planned maintenance window

## Workflow Automation

### Step Types
- **Manual**: Requires human action to complete
- **Notification**: Automatically send notification
- **API Call**: Execute API request
- **Delay**: Wait for specified duration

### Workflow Execution
1. Incident created from template
2. Workflow steps initiated
3. Automated steps execute immediately
4. Manual steps wait for user action
5. Track execution status and completion

### Example Workflows
- **Database Outage**: Notify team → Failover → Monitor → Update status
- **DDoS Attack**: Enable rate limiting → Notify security → Analyze traffic
- **Deployment Issue**: Rollback → Test → Notify stakeholders

## Architecture

### Multi-Tenancy
- All incidents isolated by tenant ID
- JWT token contains tenant context
- Automatic tenant filtering on all queries

### Middleware Stack
1. **Recovery Middleware**: Panic recovery
2. **Logger Middleware**: Request/response logging
3. **CORS Middleware**: Cross-origin resource sharing
4. **Security Headers**: Standard security headers
5. **Rate Limiting**: Tenant-based rate limiting
6. **JWT Authentication**: Token validation
7. **Tenant Middleware**: Multi-tenant context

### Integration Points
- **Component Service**: Update component status
- **Notification Service**: Send incident notifications
- **Analytics Service**: Track incident metrics
- **Event Store**: Publish incident events

## Monitoring

### Key Metrics
- Active incidents count
- Incident resolution time (MTTR)
- Incident creation rate
- Workflow execution success rate
- API throughput and latency

### Health Checks
- Database connectivity
- Redis connectivity (if enabled)
- External service connectivity
- Circuit breaker status

## Best Practices

### Incident Creation
- Use templates for common incident types
- Include clear, concise descriptions
- Set appropriate impact levels
- Link affected components

### Incident Updates
- Post regular updates during incidents
- Use consistent status transitions
- Document resolution steps
- Conduct post-mortems for major incidents

### Template Design
- Create templates for recurring incidents
- Define clear workflow steps
- Balance automation with manual oversight
- Test templates before production use

## Notifications

### Incident Notifications
- **Created**: Notify subscribers when incident created
- **Updated**: Notify on status changes
- **Resolved**: Notify when incident resolved

### Notification Channels
- Email
- SMS
- Webhook
- Slack/Teams integration

This service is critical for transparent incident communication, efficient incident response, and automated incident workflows across the platform.
