# Analytics Service

## Overview
The Analytics Service provides comprehensive analytics, metrics tracking, SLA monitoring, and reporting capabilities for the Beakon status page platform. It processes analytical data, generates reports, manages dashboards, and tracks Service Level Agreements (SLAs) across all tenants.

## Key Features

### Analytics & Metrics Management
- **Metrics Collection**: Track custom metrics across tenants
- **Metric Data Points**: Time-series metric data storage and retrieval
- **Analytics Overview**: Aggregated analytics across all tenant metrics
- **Data Visualization**: Support for dashboards and widgets

### Reporting System
- **Report Generation**: Create scheduled or on-demand reports
- **Report Templates**: Customizable report formats
- **Multi-format Export**: CSV, JSON, and Excel export support
- **Public Reports**: Share reports externally with authentication bypass

### Dashboard Management
- **Custom Dashboards**: Create tenant-specific dashboards
- **Widget System**: Modular dashboard widgets
- **Real-time Updates**: Live data refresh for dashboards
- **Dashboard Sharing**: Share dashboards across teams

### SLA Monitoring
- **SLA Definition**: Define SLA targets and thresholds
- **SLA Calculation**: Automatic SLA measurement and calculation
- **SLA Breaches**: Track and report SLA violations
- **Uptime Tracking**: Component uptime calculations
- **Response Time Monitoring**: Track and analyze response times
- **SLA Reports**: Generate comprehensive SLA compliance reports
- **SLA Statistics**: Aggregated SLA performance metrics

### Advanced Features
- **Circuit Breaker**: Database and external service protection
- **Connection Pooling**: Optimized database connections
- **Caching**: Redis/in-memory caching for performance
- **Rate Limiting**: API rate limiting for fair usage
- **Health Monitoring**: Database and service health checks

## API Endpoints

### Health & Monitoring
- `GET /health` - Health check

### Public Analytics (No Authentication)
- `GET /api/v1/public/metrics` - Public metrics data
- `GET /api/v1/public/reports` - Public reports

### Protected Analytics Routes

#### Analytics Overview
- `GET /api/v1/analytics/overview` - Get analytics overview

#### Metrics Management
- `GET /api/v1/analytics/metrics` - List all metrics
- `POST /api/v1/analytics/metrics` - Create new metric
- `GET /api/v1/analytics/metrics/:id` - Get metric details
- `PUT /api/v1/analytics/metrics/:id` - Update metric
- `DELETE /api/v1/analytics/metrics/:id` - Delete metric
- `GET /api/v1/analytics/metrics/:id/data` - Get metric data points
- `POST /api/v1/analytics/metrics/:id/data` - Add metric data point

#### Reports Management
- `GET /api/v1/analytics/reports` - List all reports
- `POST /api/v1/analytics/reports` - Create new report
- `GET /api/v1/analytics/reports/:id` - Get report details
- `PUT /api/v1/analytics/reports/:id` - Update report
- `DELETE /api/v1/analytics/reports/:id` - Delete report
- `POST /api/v1/analytics/reports/:id/generate` - Generate report
- `GET /api/v1/analytics/reports/:id/download` - Download report

#### Dashboard Management
- `GET /api/v1/analytics/dashboards` - List dashboards
- `POST /api/v1/analytics/dashboards` - Create dashboard
- `GET /api/v1/analytics/dashboards/:id` - Get dashboard
- `PUT /api/v1/analytics/dashboards/:id` - Update dashboard
- `DELETE /api/v1/analytics/dashboards/:id` - Delete dashboard
- `GET /api/v1/analytics/dashboards/:id/widgets` - Get dashboard widgets
- `POST /api/v1/analytics/dashboards/:id/widgets` - Add widget to dashboard
- `PUT /api/v1/analytics/dashboards/:id/widgets/:widget_id` - Update widget
- `DELETE /api/v1/analytics/dashboards/:id/widgets/:widget_id` - Remove widget

#### Data Export
- `GET /api/v1/analytics/exports/csv` - Export data as CSV
- `GET /api/v1/analytics/exports/json` - Export data as JSON
- `GET /api/v1/analytics/exports/excel` - Export data as Excel

### SLA Management

#### SLA Definitions
- `POST /api/v1/sla` - Create SLA
- `GET /api/v1/sla` - List SLAs
- `GET /api/v1/sla/:id` - Get SLA details
- `PUT /api/v1/sla/:id` - Update SLA
- `DELETE /api/v1/sla/:id` - Delete SLA

#### SLA Measurements
- `POST /api/v1/sla/:id/calculate` - Calculate SLA measurement
- `GET /api/v1/sla/:id/measurements` - Get SLA measurements

#### SLA Monitoring
- `GET /api/v1/sla/breaches` - List SLA breaches
- `POST /api/v1/sla/reports` - Generate SLA report
- `GET /api/v1/sla/reports` - List SLA reports
- `POST /api/v1/sla/statistics` - Get SLA statistics

#### SLA Targets & Tracking
- `POST /api/v1/sla/targets` - Create SLA target template
- `GET /api/v1/sla/targets` - List SLA targets
- `POST /api/v1/sla/uptime/calculate` - Calculate uptime
- `POST /api/v1/sla/response-time` - Record response time

## Dependencies

### Internal Services
- **analytics-consumer**: Provides raw analytics events
- **component-service**: For component uptime calculations
- **incident-service**: For incident impact on SLAs

### External Dependencies
- **shared-resilience**: Common patterns, database, middleware, circuit breakers
- **PostgreSQL**: Analytics and SLA data storage
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

### Feature Flags
- `CIRCUIT_BREAKER_ENABLED`: Enable circuit breaker protection
- `CACHE_ENABLED`: Enable caching layer
- `RATE_LIMIT_ENABLED`: Enable API rate limiting
- `MONITORING_ENABLED`: Enable health monitoring

## Development

### Running the Service
```bash
cd microservices/analytics-service
go run cmd/main.go
```

### Building
```bash
go build -o analytics-service cmd/main.go
```

### Testing
```bash
go test ./...
```

### Project Structure
```
analytics-service/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── handlers/            # HTTP request handlers
│   │   ├── analytics_handler.go
│   │   └── sla_handler.go
│   ├── services/            # Business logic layer
│   │   ├── analytics_service.go
│   │   └── sla_service.go
│   └── models/              # Data models
└── go.mod                   # Go module definition
```

## Architecture

### Multi-Tenancy
- All data isolated by tenant ID
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

### Resilience Features
- **Circuit Breakers**: Protect database and external services
- **Connection Pooling**: Efficient database connections
- **Health Checks**: Continuous database health monitoring
- **Graceful Shutdown**: Clean service termination
- **Structured Logging**: Comprehensive operational visibility

### Performance Optimizations
- **Caching**: Frequently accessed data cached in Redis/memory
- **Database Indexing**: Optimized queries for analytics workloads
- **Batch Processing**: Bulk operations for efficiency
- **Query Optimization**: Efficient time-series data retrieval

## SLA Calculation

### Uptime SLA
```
Uptime % = (Total Time - Downtime) / Total Time * 100
```

### Response Time SLA
```
Response Time SLA = Percentage of requests under threshold
```

### Availability Targets
- 99.9% (Three Nines): ~43 minutes downtime/month
- 99.95%: ~21 minutes downtime/month
- 99.99% (Four Nines): ~4 minutes downtime/month

## Monitoring

### Key Metrics
- Request throughput (requests/sec)
- Response latency (p50, p95, p99)
- Error rate
- Database query performance
- Cache hit rate
- SLA compliance rate

### Health Checks
- Database connectivity
- Redis connectivity (if enabled)
- Circuit breaker status
- Service readiness

This service is essential for providing data-driven insights, performance monitoring, and SLA compliance tracking across the entire platform.
