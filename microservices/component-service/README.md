# Component Service

## Overview
The Component Service is a core service of the Beakon status page platform that manages system components, their status, and organizational structure. It provides the foundation for status page displays by tracking individual services, their health status, grouping, and historical status information.

## Key Features

### Component Management
- **Component CRUD Operations**: Complete component lifecycle management
- **Component Status Tracking**: Real-time status management with history
- **Component Grouping**: Hierarchical organization using component groups
- **Visibility Control**: Public/private component display management
- **Position Management**: Ordering and arrangement of components

### Status Management
- **Status Types**: Support for operational, degraded, partial outage, major outage, and maintenance statuses
- **Status History**: Complete historical tracking of status changes
- **Status Transitions**: Automated status change logging and auditing
- **Public Status Calculation**: Automatic overall status determination

### Component Organization
- **Component Groups**: Logical grouping of related components
- **Group Management**: CRUD operations for component groups
- **Hierarchical Display**: Nested component organization
- **Group Visibility**: Public/private group display control

### Monitoring Integration
- **Component Metrics**: Performance and availability metrics collection
- **Component Alerts**: Status-based alerting system
- **Component Webhooks**: Integration with external monitoring systems
- **Automated Status Updates**: Integration-driven status management

### Public API
- **Public Component Display**: Filtered component information for status pages
- **Public Status Calculation**: Overall system status determination
- **Historical Data**: Public access to component history and trends

## API Endpoints

### Health & Monitoring
- `GET /health` - Service health check

### Component Management
- `GET /api/v1/components` - List components (with pagination)
- `POST /api/v1/components` - Create new component
- `GET /api/v1/components/:id` - Get component details
- `PUT /api/v1/components/:id` - Update component
- `DELETE /api/v1/components/:id` - Delete component
- `GET /api/v1/components/public` - Get public components (for status page display)

### Component Status Management
- `PUT /api/v1/components/:id/status` - Update component status
- `GET /api/v1/components/:id/history` - Get component status history
- `GET /api/v1/components/:id/uptime` - Get component uptime statistics

### Component Group Management
- `GET /api/v1/component-groups` - List component groups
- `POST /api/v1/component-groups` - Create component group
- `GET /api/v1/component-groups/:id` - Get group details
- `PUT /api/v1/component-groups/:id` - Update component group
- `DELETE /api/v1/component-groups/:id` - Delete component group

### Public Status API
- `GET /api/v1/public/status` - Get overall public status
- `GET /api/v1/public/components` - Get public component list
- `GET /api/v1/public/components/:id` - Get public component details

### Metrics & Analytics
- `GET /api/v1/components/:id/metrics` - Get component metrics
- `POST /api/v1/components/:id/metrics` - Record component metric
- `GET /api/v1/components/:id/alerts` - Get component alerts

### Integration Endpoints
- `POST /api/v1/webhooks/status-update` - Webhook for external status updates
- `GET /api/v1/components/:id/webhooks` - Get component webhooks
- `POST /api/v1/components/:id/webhooks` - Create component webhook

## Dependencies

### Internal Dependencies
- **tenant-admin-service**: Multi-tenant context and permissions
- **monitoring-service**: Health check and metrics integration
- **incident-service**: Incident-related component status updates

### External Dependencies
- **PostgreSQL**: Primary database for component data
- **shared-resilience**: Common patterns for database management, middleware, and observability

### Database Models
The service manages these core models:
- **Component**: Main component entity with status, visibility, and metadata
- **ComponentGroup**: Organization structure for components
- **ComponentStatus**: Status tracking with timestamps and responsible users
- **ComponentHistory**: Historical status change records
- **ComponentMetric**: Performance and availability metrics
- **ComponentAlert**: Status-based alerting configuration
- **ComponentWebhook**: External integration endpoints

## Configuration

### Environment Variables
- `SERVER_PORT`: HTTP server port
- `SERVER_HOST`: HTTP server host
- `ENVIRONMENT`: Runtime environment (development/production)
- `DB_HOST`: Database host
- `DB_PORT`: Database port
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name

### Component Status Values
- **operational**: Component is functioning normally
- **degraded_performance**: Component is operational but experiencing performance issues
- **partial_outage**: Component is partially unavailable
- **major_outage**: Component is completely unavailable
- **maintenance**: Component is under scheduled maintenance

## Development

### Running the Service
```bash
cd microservices/component-service
go run cmd/main.go
```

### Building
```bash
go build -o component-service cmd/main.go
```

### Testing
```bash
go test ./...
```

### Project Structure
```
component-service/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── handlers/            # HTTP handlers
│   ├── models/              # Data models and database schemas
│   └── services/            # Business service layer
└── go.mod                   # Go module definition
```

## Architecture Notes

### Component Lifecycle
1. **Creation**: Component created with initial operational status
2. **Status Updates**: Regular status monitoring and updates
3. **History Tracking**: All status changes logged with timestamps
4. **Grouping**: Assignment to logical component groups
5. **Visibility**: Public/private display configuration
6. **Integration**: Connection with monitoring and incident systems

### Status Management Flow
1. **Status Change Request**: Via API or webhook
2. **Validation**: Status value and permission validation
3. **History Recording**: Previous status archived to history
4. **Status Update**: Component status updated in database
5. **Notification**: Status change notifications (if configured)
6. **Public Status Recalculation**: Overall system status updated

### Public Status Calculation
The service calculates overall system status based on component statuses:
- **Operational**: All visible components are operational
- **Degraded Performance**: One or more components have degraded performance
- **Partial Outage**: One or more components have partial outages
- **Major Outage**: One or more components have major outages
- **Maintenance**: One or more components are under maintenance

### Multi-tenant Architecture
- **Tenant Isolation**: All component data scoped to specific tenants
- **Tenant Context**: Automatic tenant identification via middleware
- **Data Separation**: Complete isolation of tenant data
- **Public Access**: Tenant-specific public component access

### Integration Architecture
- **Webhook Support**: External system integration for status updates
- **Monitoring Integration**: Direct integration with monitoring services
- **Incident Integration**: Automatic component status updates during incidents
- **Metrics Collection**: Component performance and availability tracking

### Database Design
- **Optimized Queries**: Efficient retrieval of component and status data
- **Historical Data**: Efficient storage and retrieval of status history
- **Indexing**: Performance-optimized database indexes
- **Data Integrity**: Foreign key relationships and constraints

### Performance Considerations
- **Caching**: Component status and configuration caching
- **Pagination**: Efficient handling of large component lists
- **Query Optimization**: Minimized database queries for status calculations
- **Background Processing**: Asynchronous status update processing

### Security & Access Control
- **Multi-tenant Security**: Strict tenant data isolation
- **Permission-based Access**: Integration with RBAC system
- **Public API Security**: Safe exposure of public component data
- **Input Validation**: Comprehensive request validation

This service is essential for the status page functionality, providing the data foundation for displaying system health and component status to both internal administrators and external users.