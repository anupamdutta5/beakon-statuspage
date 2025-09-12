# Status Page Microservices Architecture

This document describes the complete microservices architecture for the Status Page application, following best practices for independent, scalable, and maintainable services.

## 🏗️ Architecture Overview

The Status Page application has been refactored from a monolithic architecture into a microservices architecture with the following principles:

- **Business Feature Separation**: Each microservice represents a distinct business capability
- **Independent Deployment**: Each service can be deployed, updated, and scaled independently
- **Event-Driven Communication**: Services communicate through events and APIs
- **Shared Library**: Common code is shared through a separate Go module
- **Repository Independence**: Each microservice can be in its own repository

## 🎯 Microservices

### Core Business Services

| Service | Port | Description | Repository |
|---------|------|-------------|------------|
| **API Gateway** | 8080 | Central entry point, routing, authentication | `statuspage-api-gateway` |
| **User Service** | 8081 | User management, authentication, authorization | `statuspage-user-service` |
| **Tenant Service** | 8082 | Multi-tenant SaaS management | `statuspage-tenant-service` |
| **Component Service** | 8083 | Service/component status management | `statuspage-component-service` |
| **Incident Service** | 8084 | Incident management, status updates | `statuspage-incident-service` |
| **Notification Service** | 8085 | Email, SMS, webhook notifications | `statuspage-notification-service` |
| **Payment Service** | 8086 | Billing, subscriptions, payment processing | `statuspage-payment-service` |
| **Analytics Service** | 8087 | Metrics, reporting, uptime statistics | `statuspage-analytics-service` |
| **Monitoring Service** | 8088 | Health checks, monitoring, alerts | `statuspage-monitoring-service` |

### Event Consumers

| Consumer | Port | Description | Repository |
|----------|------|-------------|------------|
| **Notification Consumer** | 8089 | Processes notification events | `statuspage-notification-consumer` |
| **Analytics Consumer** | 8090 | Processes analytics events | `statuspage-analytics-consumer` |
| **Audit Consumer** | 8091 | Processes audit events | `statuspage-audit-consumer` |
| **Billing Consumer** | 8092 | Processes billing events | `statuspage-billing-consumer` |

## 📦 Shared Library

The `shared-lib` package provides common functionality across all microservices:

- **Models**: Database models and data structures
- **Config**: Configuration management utilities
- **Events**: Event handling interfaces and types
- **Database**: Database connection and migration utilities
- **Logger**: Structured logging utilities
- **Utils**: Common utility functions

### Usage in Microservices

Each microservice imports the shared library as a Go module:

```go
// Each service now has its own local types and dependencies
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose (for infrastructure)
- PostgreSQL (for data storage)
- Redis (for caching and sessions)

### Starting All Services

1. **Start Infrastructure**:
   ```bash
   docker-compose up -d postgres redis
   ```

2. **Start All Microservices**:
   ```bash
   ./start-microservices.sh
   ```

3. **Stop All Services**:
   ```bash
   ./stop-microservices.sh
   ```

### Individual Service Development

Each service can be developed and run independently:

```bash
cd microservices/api-gateway
./start.sh
```

## 🔧 Service Configuration

Each service is configured through environment variables with sensible defaults:

### Common Configuration

- `PORT`: Service port
- `ENVIRONMENT`: development/production
- `LOG_LEVEL`: Logging level
- `JWT_SECRET`: JWT secret key

### Service-Specific Configuration

Each service has its own configuration file and environment variables. See individual service READMEs for details.

## 📡 API Endpoints

### API Gateway (Port 8080)

The API Gateway provides a unified API for all client requests:

#### Public Endpoints
- `GET /health` - Health check
- `GET /api/v1/public/status` - Public status
- `GET /api/v1/public/components` - Public components
- `GET /api/v1/public/incidents` - Public incidents

#### Protected Endpoints
- `GET /api/v1/users` - User management
- `GET /api/v1/tenants` - Tenant management
- `GET /api/v1/components` - Component management
- `GET /api/v1/incidents` - Incident management
- `GET /api/v1/payments` - Payment management
- `GET /api/v1/analytics/*` - Analytics data

## 🔐 Authentication & Authorization

- **JWT Tokens**: All protected endpoints require valid JWT tokens
- **Role-Based Access**: Different user roles with appropriate permissions
- **Tenant Isolation**: Multi-tenant data isolation
- **API Gateway**: Centralized authentication and authorization

## 📊 Event-Driven Architecture

### Event Types

- **User Events**: UserCreated, UserUpdated, UserDeleted
- **Tenant Events**: TenantCreated, TenantUpdated, TenantDeleted
- **Component Events**: ComponentCreated, ComponentStatusChanged
- **Incident Events**: IncidentCreated, IncidentResolved
- **Payment Events**: PaymentProcessed, SubscriptionCreated
- **Notification Events**: NotificationSent, NotificationFailed

### Event Flow

```
Service → Event Publisher → Event Store → Event Consumers
```

## 🗄️ Data Management

### Database Strategy

- **Service Databases**: Each service can have its own database
- **Shared Models**: Common data structures in shared library
- **Event Sourcing**: Event store for audit and replay capabilities
- **Migrations**: Automated database migrations

### Data Consistency

- **Eventual Consistency**: Through event-driven architecture
- **Saga Pattern**: For distributed transactions
- **Compensation**: For handling failures

## 🔍 Monitoring & Observability

### Health Checks

Each service exposes health endpoints:
- `GET /health` - Service health status
- `GET /metrics` - Prometheus metrics (if enabled)

### Logging

- **Structured Logging**: JSON format with correlation IDs
- **Centralized Logging**: All logs in `logs/` directory
- **Log Levels**: Configurable per service

### Metrics

- **Service Metrics**: Request counts, response times, error rates
- **Business Metrics**: User registrations, payment success rates
- **Infrastructure Metrics**: CPU, memory, disk usage

## 🧪 Testing Strategy

### Unit Tests

Each service has comprehensive unit tests:
```bash
cd microservices/api-gateway
go test ./...
```

### Integration Tests

Test service interactions:
```bash
go test ./tests/integration/...
```

### Contract Tests

Ensure API compatibility:
```bash
go test ./tests/contract/...
```

## 🚢 Deployment

### Development

```bash
./start-microservices.sh
```

### Production

Each service can be deployed independently:

1. **Docker Containers**:
   ```bash
   docker build -t statuspage-api-gateway .
   docker run -p 8080:8080 statuspage-api-gateway
   ```

2. **Kubernetes**:
   ```bash
   kubectl apply -f k8s/
   ```

3. **Cloud Platforms**: AWS ECS, Google Cloud Run, Azure Container Instances

## 🔧 Development Workflow

### Adding a New Service

1. Create service directory: `microservices/new-service`
2. Initialize Go module: `go mod init github.com/enterprise-status/statuspage-new-service`
3. Each service is now independent with its own local types
4. Implement service following the established patterns
5. Add to startup scripts
6. Update documentation

### Modifying Shared Library

1. Make changes in `shared-lib/`
2. Update version in `go.mod`
3. Each service maintains its own independent dependencies

## 📚 Service Documentation

Each service has its own README with:
- Service-specific configuration
- API documentation
- Development setup
- Deployment instructions

## 🤝 Contributing

1. Follow Go best practices
2. Write comprehensive tests
3. Update documentation
4. Follow the established patterns
5. Ensure backward compatibility

## 📋 TODO

- [ ] Implement remaining microservices
- [ ] Add comprehensive testing
- [ ] Set up CI/CD pipelines
- [ ] Add monitoring dashboards
- [ ] Implement service mesh
- [ ] Add distributed tracing
- [ ] Set up automated deployment

## 🆘 Troubleshooting

### Common Issues

1. **Port Conflicts**: Check if ports are already in use
2. **Service Dependencies**: Ensure dependent services are running
3. **Configuration**: Verify environment variables
4. **Database**: Check database connectivity

### Logs

All service logs are available in the `logs/` directory:
```bash
tail -f logs/api-gateway.log
```

### Health Checks

Check service health:
```bash
curl http://localhost:8080/health
```

## 📞 Support

For issues and questions:
1. Check service-specific documentation
2. Review logs in `logs/` directory
3. Check health endpoints
4. Verify configuration

---

**Note**: This architecture follows microservices best practices and is designed for scalability, maintainability, and independent deployment. Each service can be developed, tested, and deployed independently while maintaining system coherence through the shared library and event-driven communication.
