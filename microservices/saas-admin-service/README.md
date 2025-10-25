# SaaS Admin Service

## Overview
The SaaS Admin Service is the central platform administration service for the Beakon status page platform. It provides comprehensive platform management capabilities including subscription plans, features, pricing, statistics, and platform-wide configuration. This service also includes a web-based administrative dashboard.

## Key Features

### Platform Management
- **Platform Configuration**: Global platform settings and configuration
- **Plan Management**: Subscription plan creation, modification, and lifecycle management
- **Feature Management**: Platform feature definition and management
- **Feature Flag Management**: Dynamic feature flagging system
- **Statistics & Analytics**: Platform-wide statistics and metrics collection

### Administrative Functions
- **Admin User Management**: Platform administrator accounts and permissions
- **Notification Management**: System-wide notification handling
- **Activity Logging**: Comprehensive audit trail of administrative actions
- **Backup Management**: System backup creation and management

### Pricing & Billing Integration
- **Pricing Management**: Pricing structure configuration and management
- **Pricing Features**: Feature-based pricing component management
- **Pricing Tiers**: Tiered pricing structure management
- **Plan Feature Assignment**: Association of features with subscription plans
- **Public Pricing API**: External pricing information exposure

### Web Dashboard
- **HTML Template System**: Server-side rendered administrative interface
- **Static File Serving**: CSS, JavaScript, and asset management
- **Admin Dashboard**: Comprehensive administrative control panel

### Tenant Management Proxy
- **Tenant Operations**: Proxied tenant management operations
- **Tenant Analytics**: Tenant-specific analytics and reporting

### Real-time Monitoring
- **WebSocket Support**: Real-time monitoring updates
- **Monitor Management**: System monitor configuration and control
- **Real-time Statistics**: Live system statistics and metrics

## API Endpoints

### Health & Platform
- `GET /api/v1/health` - Health check
- `GET /api/v1/platform` - Get platform configuration
- `PUT /api/v1/platform` - Update platform configuration
- `GET /api/v1/stats` - Get platform statistics

### Web Interface
- `GET /` - Redirect to health endpoint
- `GET /admin` - Administrative dashboard
- `GET /static/*` - Static files (CSS, JS, images)
- `GET /favicon.ico` - Favicon

### Plan Management
- `GET /api/v1/plans` - List all plans
- `POST /api/v1/plans` - Create new plan
- `GET /api/v1/plans/:id` - Get plan details
- `PUT /api/v1/plans/:id` - Update plan
- `DELETE /api/v1/plans/:id` - Delete plan
- `GET /api/v1/plans/slug/:slug` - Get plan by slug

### Feature Management
- `GET /api/v1/features` - List all features
- `POST /api/v1/features` - Create new feature
- `GET /api/v1/features/:id` - Get feature details
- `PUT /api/v1/features/:id` - Update feature
- `DELETE /api/v1/features/:id` - Delete feature

### Feature Flag Management
- `GET /api/v1/feature-flags` - List all feature flags
- `POST /api/v1/feature-flags` - Create new feature flag
- `GET /api/v1/feature-flags/:id` - Get feature flag details
- `PUT /api/v1/feature-flags/:id` - Update feature flag
- `DELETE /api/v1/feature-flags/:id` - Delete feature flag

### Admin User Management
- `GET /api/v1/admin-users` - List admin users
- `POST /api/v1/admin-users` - Create admin user
- `GET /api/v1/admin-users/:id` - Get admin user details
- `PUT /api/v1/admin-users/:id` - Update admin user
- `DELETE /api/v1/admin-users/:id` - Delete admin user

### Notification Management
- `GET /api/v1/notifications` - List notifications
- `POST /api/v1/notifications` - Create notification
- `GET /api/v1/notifications/:id` - Get notification details
- `PUT /api/v1/notifications/:id` - Update notification
- `DELETE /api/v1/notifications/:id` - Delete notification

### Activity & Audit
- `GET /api/v1/activities` - List activity logs

### Backup Management
- `GET /api/v1/backups` - List backups
- `POST /api/v1/backups` - Create backup
- `GET /api/v1/backups/:id` - Get backup details
- `DELETE /api/v1/backups/:id` - Delete backup

### Pricing Management
- `GET /api/v1/pricing/features` - Get pricing features
- `POST /api/v1/pricing/features` - Create pricing feature
- `PUT /api/v1/pricing/features/:id` - Update pricing feature
- `DELETE /api/v1/pricing/features/:id` - Delete pricing feature

- `GET /api/v1/pricing/plans/:planId/tiers` - Get pricing tiers for plan
- `POST /api/v1/pricing/tiers` - Create pricing tier
- `PUT /api/v1/pricing/tiers/:id` - Update pricing tier
- `DELETE /api/v1/pricing/tiers/:id` - Delete pricing tier

- `GET /api/v1/pricing/plans/:planId/features` - Get plan features
- `POST /api/v1/pricing/plans/features/assign` - Assign feature to plan
- `POST /api/v1/pricing/plans/features/remove` - Remove feature from plan

- `GET /api/v1/pricing/plans/public` - Get public pricing plans
- `POST /api/v1/pricing/sync` - Sync pricing to landing page

### Tenant Management (Proxied)
- `GET /api/v1/tenants` - List tenants
- `POST /api/v1/tenants` - Create tenant
- `GET /api/v1/tenants/:id` - Get tenant details
- `PUT /api/v1/tenants/:id` - Update tenant
- `DELETE /api/v1/tenants/:id` - Delete tenant

### Analytics Management
- `GET /api/v1/analytics/overview` - Analytics overview
- `GET /api/v1/analytics/metrics` - Get analytics metrics
- `POST /api/v1/analytics/metrics` - Create analytics metric
- `GET /api/v1/analytics/health` - Check analytics health

### Real-time Monitoring
- `GET /api/v1/monitoring/monitors` - Get monitors
- `POST /api/v1/monitoring/monitors` - Create monitor
- `PUT /api/v1/monitoring/monitors/:id` - Update monitor
- `DELETE /api/v1/monitoring/monitors/:id` - Delete monitor
- `POST /api/v1/monitoring/monitors/:id/pause` - Pause monitor
- `POST /api/v1/monitoring/monitors/:id/resume` - Resume monitor
- `POST /api/v1/monitoring/monitors/pause-all` - Pause all monitors
- `POST /api/v1/monitoring/monitors/resume-all` - Resume all monitors
- `GET /api/v1/monitoring/stats` - Get real-time stats

### WebSocket Endpoints
- `GET /api/v1/ws/monitoring` - WebSocket for real-time monitoring updates

### Incident Management (Proxied)
- `GET /api/v1/incidents` - List incidents
- `POST /api/v1/incidents` - Create incident
- `GET /api/v1/incidents/:id` - Get incident details
- `PUT /api/v1/incidents/:id` - Update incident
- `DELETE /api/v1/incidents/:id` - Delete incident
- `GET /api/v1/incidents/stats` - Get incident statistics

### Integration Management (Proxied)
- `GET /api/v1/integrations` - List integrations
- `POST /api/v1/integrations` - Create integration
- `PUT /api/v1/integrations/:id` - Update integration
- `DELETE /api/v1/integrations/:id` - Delete integration
- `POST /api/v1/integrations/:id/test` - Test integration

### Component Management (Proxied)
- `GET /api/v1/components` - List components
- `POST /api/v1/components` - Create component
- `GET /api/v1/components/:id` - Get component details
- `PUT /api/v1/components/:id` - Update component
- `DELETE /api/v1/components/:id` - Delete component
- `GET /api/v1/component-groups` - Get component groups

## Dependencies

### Internal Services
- **tenant-admin-service**: Tenant management operations (proxied)
- **analytics-service**: Analytics data collection and reporting (proxied)
- **monitoring-service**: Real-time monitoring functionality (proxied)
- **incident-service**: Incident management operations (proxied)
- **component-service**: Component management operations (proxied)
- **integration-service**: Third-party integration management (proxied)

### External Dependencies
- **shared-resilience**: Common resilience patterns, database management, middleware
- **PostgreSQL**: Primary database for platform data
- **Redis**: Caching and real-time features (optional)

### Database Models
The service manages these core models:
- Platform configuration
- SaaS plans, features, and feature flags
- Pricing tiers and feature associations
- Admin users and notifications
- Activity logs and backup records
- Statistics and metrics

## Configuration

### Environment Variables
- `SERVER_PORT`: HTTP server port (default from shared-resilience config)
- `SERVER_HOST`: HTTP server host
- `ENVIRONMENT`: Runtime environment (development/production)
- `DB_HOST`: Database host
- `DB_PORT`: Database port
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `REDIS_HOST`: Redis host (optional)
- `REDIS_PORT`: Redis port (optional)
- `JWT_SECRET`: JWT signing secret

### Template Configuration
- Templates are loaded from `../web/templates/` (both root and partials)
- Static files served from `../web/static/`
- Automatic template dependency resolution

## Development

### Running the Service
```bash
cd microservices/saas-admin-service
go run cmd/main.go
```

### Building
```bash
go build -o saas-admin-service cmd/main.go
```

### Testing
```bash
go test ./...
```

### Project Structure
```
saas-admin-service/
├── cmd/
│   ├── main.go              # Application entry point
│   └── seed/                # Database seeding utilities
├── internal/
│   ├── clients/             # External service clients
│   ├── database/            # Database management
│   ├── handlers/            # HTTP handlers and business logic
│   ├── models/              # Data models and database schemas
│   └── services/            # Business service layer
├── web/
│   ├── static/              # Static assets (CSS, JS, images)
│   └── templates/           # HTML templates
│       ├── partials/        # Template partials
│       └── *.html           # Main templates
└── configs/                 # Configuration files
```

## Architecture Notes

### Template System
- Server-side rendered HTML templates using Go's html/template package
- Support for template inheritance and partials
- Automatic template loading from multiple directories
- CDN integration for external resources (Bootstrap, jQuery, etc.)

### Database Management
- Automatic database creation if not exists
- Model migration system with graceful error handling
- Connection pooling and health checks via shared-resilience

### Middleware Stack
- Comprehensive middleware via shared-resilience
- Custom CSP headers for admin dashboard with CDN support
- Authentication, logging, recovery, and monitoring

### Real-time Features
- WebSocket support for real-time monitoring updates
- Live statistics and metrics streaming
- Monitor pause/resume functionality

### Pricing Integration
- Dynamic pricing feature management
- Plan-feature association system
- Public pricing API for external consumption
- Automatic sync to landing page service

This service serves as the administrative backbone of the platform, providing both API endpoints for programmatic access and a web-based dashboard for human administrators. It integrates with multiple other services while maintaining its own core platform management responsibilities.# Status Update - 2025-10-25 10:15:29

Repository synchronized and verified.
