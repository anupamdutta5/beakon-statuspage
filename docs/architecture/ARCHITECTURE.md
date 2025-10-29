# Beakon Status Page - System Architecture

## Overview
Beakon is a comprehensive, enterprise-grade status page platform built using microservices architecture. The platform provides multi-tenant SaaS capabilities for organizations to create and manage status pages for their services, with advanced features like incident management, monitoring integration, analytics, and custom branding.

## Architecture Philosophy

### Microservices Design
- **Service Independence**: Each microservice is an independent unit with its own database, business logic, and deployment lifecycle
- **Domain-Driven Design**: Services are organized around business domains (authentication, tenant management, monitoring, etc.)
- **API-First**: All services expose RESTful APIs and communicate via HTTP
- **Multi-Tenant**: All services support multi-tenancy with complete data isolation

### Technology Stack
- **Runtime**: Go 1.21+ for all microservices
- **Database**: PostgreSQL for data persistence
- **Caching**: Redis for caching and session management (optional)
- **Messaging**: HTTP/REST for direct service-to-service communication
- **Monitoring**: Prometheus for metrics collection
- **Authentication**: JWT-based authentication and authorization
- **Resilience**: Circuit breakers, rate limiting, and retry mechanisms

### Communication Pattern (Post-October 2025)
⚠️ **API Gateway Deprecated**: As of October 26, 2025, the API Gateway is no longer used. All services communicate directly with each other using HTTP/REST with built-in circuit breakers and service discovery.

## Service Interaction Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Frontend Applications                              │
│         saas-admin-frontend (3001) + tenant-admin-frontend (3002)           │
└────────────────────────────┬────────────────────────────────────────────────┘
                             │
                             ▼
      ┌──────────────────────────────────────────────────────────┐
      │         Direct HTTP Calls (with Circuit Breakers)         │
      └──────────────────────────────────────────────────────────┘
                             │
      ┌──────────────────────┼──────────────────┬──────────────────┬──────────────┐
      │                      │                   │                  │              │
      ▼                      ▼                   ▼                  ▼              ▼
┌───────────┐          ┌───────────┐      ┌───────────┐     ┌──────────┐   ┌──────────┐
│   User    │          │  Tenant   │      │   SaaS    │     │Component │   │Incident  │
│  Service  │          │   Admin   │      │  Admin    │     │ Service  │   │ Service  │
│  (8081)   │          │  (8099)   │      │  (8098)   │     │  (8084)  │   │  (8086)  │
│           │◄────────►│           │◄────►│           │◄───►│          │◄─►│          │
│ Auth &    │          │ Tenant    │      │ Platform  │     │ Status   │   │ Incident │
│ Users     │          │ Mgmt RBAC │      │ Admin     │     │ Tracking │   │ Mgmt     │
└─────┬─────┘          └─────┬─────┘      └─────┬─────┘     └────┬─────┘   └────┬─────┘
      │                      │                  │                  │              │
      │                  └──────────────────┴──────────────────┘              │
      │                           │                                           │
      │                           ▼                                           │
      │                  ┌────────────────┐                                   │
      │                  │  tenant_admin  │                                   │
      │                  │      _db       │                                   │
      │                  │                │                                   │
      │                  │ (Shared by 2   │                                   │
      │                  │   services)    │                                   │
      │                  └────────────────┘                                   │
      │                                                                        │
      ▼                                                                        ▼
┌───────────────┐                                                    ┌──────────────┐
│ statuspage_   │                                                    │statuspage_   │
│    user       │                                                    │  incident    │
└───────────────┘                                                    └──────────────┘

                           Other HTTP Services
      ┌─────────────┬─────────────┬──────────────┬──────────────┐
      │             │             │              │              │
      ▼             ▼             ▼              ▼              ▼
┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
│ Notif    │  │ Payment  │  │Analytics │  │Monitoring│  │ Branding │
│ Service  │  │ Service  │  │ Service  │  │ Service  │  │ Service  │
│ (8085)   │  │ (8088)   │  │ (8090)   │  │ (8092)   │  │ (8097)   │
└──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘

                         Event Consumers (Kafka)
              ┌──────────────────────────────────────┐
              │                                       │
         ┌────▼────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐
         │Analytics│  │   Audit   │  │  Billing  │  │   Notif   │
         │Consumer │  │  Consumer │  │  Consumer │  │  Consumer │
         └─────────┘  └───────────┘  └───────────┘  └───────────┘
```

**Legend**:
- **Solid lines (│─)**: HTTP/REST API calls
- **Arrows (▼)**: Request/data flow direction
- **Numbers in ()**: Service port numbers

**Key Relationships**:
1. **API Gateway** is the single entry point for all client requests
2. **User Service** handles authentication, other services validate JWT tokens
3. **Tenant Admin & SaaS Admin** share the same database but serve different roles
4. Each HTTP service has its own dedicated PostgreSQL database
5. **Event Consumers** process asynchronous events from Kafka (no HTTP ports)

---

## System Components

### Core Services

#### 1. API Gateway (`api-gateway`)
**Role**: Central entry point and request router
- **Port**: 8080 (default)
- **Purpose**: Single entry point for all client requests
- **Responsibilities**:
  - Request routing to appropriate microservices
  - JWT authentication and authorization
  - Rate limiting and request validation
  - CORS handling and security headers
  - Circuit breaker implementation
  - Prometheus metrics collection

#### 2. User Service (`user-service`)
**Role**: Authentication and user management
- **Purpose**: Core authentication service for the platform
- **Responsibilities**:
  - User registration and login
  - JWT token management (generation, validation, refresh)
  - Password management (reset, change)
  - User profile management
  - Multi-tenant user isolation

#### 3. Tenant Admin Service (`tenant-admin-service`)
**Role**: Multi-tenant management and RBAC
- **Purpose**: Comprehensive tenant lifecycle and access control management
- **Responsibilities**:
  - Tenant CRUD operations and branding
  - Role-Based Access Control (RBAC) system
  - Status page management and configuration
  - Domain management and verification
  - Team and user role management
  - Audit logging and session management

#### 4. SaaS Admin Service (`saas-admin-service`)
**Role**: Platform administration and business logic
- **Purpose**: Platform-wide administration and SaaS operations
- **Responsibilities**:
  - Subscription plan and feature management
  - Pricing structure and billing integration
  - Platform statistics and analytics
  - Administrative dashboard (web interface)
  - Feature flag management
  - Backup and notification management

### Status Page Core Services

#### 5. Component Service (`component-service`)
**Role**: System component and status management
- **Purpose**: Manages the components displayed on status pages
- **Responsibilities**:
  - Component CRUD operations and organization
  - Component status tracking and history
  - Component groups and visibility management
  - Public status calculation for status pages
  - Component metrics and alerting

#### 6. Incident Service (`incident-service`)
**Role**: Incident lifecycle management
- **Purpose**: Complete incident management for status pages
- **Responsibilities**:
  - Incident creation, updates, and resolution
  - Incident templates and workflow management
  - Incident-component associations
  - Public incident display for status pages
  - Incident metrics and reporting (MTTR, etc.)

#### 7. Monitoring Service (`monitoring-service`)
**Role**: Comprehensive system monitoring
- **Purpose**: Multi-faceted monitoring and alerting system
- **Responsibilities**:
  - Health checks for HTTP endpoints and services
  - External service monitoring
  - Component and container monitoring
  - Kubernetes and Docker integration
  - Custom metrics collection
  - Alert management and automation
  - Maintenance window management

### Supporting Services

#### 8. Analytics Service (`analytics-service`)
**Role**: Data analytics and reporting
- **Purpose**: Platform analytics and SLA management
- **Responsibilities**:
  - Analytics data collection and processing
  - SLA definition and measurement
  - Uptime calculations and reporting
  - Performance metrics analysis
  - Business intelligence and reporting

#### 9. Notification Service (`notification-service`)
**Role**: Communication and alerting
- **Purpose**: Multi-channel notification system
- **Responsibilities**:
  - Email, SMS, and push notifications
  - Webhook delivery and management
  - Notification preferences and subscriptions
  - Integration with external communication services

#### 10. Landing Page Service (`landing-page-service`)
**Role**: Public-facing website and marketing
- **Purpose**: Public website and customer acquisition
- **Responsibilities**:
  - Marketing website and landing pages
  - Pricing information display
  - Lead generation and customer onboarding
  - SEO optimization and public content

#### 11. Payment Service (`payment-service`)
**Role**: Billing and payment processing
- **Purpose**: Financial transactions and subscription management
- **Responsibilities**:
  - Payment processing and gateway integration
  - Subscription billing and invoicing
  - Payment method management
  - Financial reporting and reconciliation

#### 12. Branding Service (`branding-service`)
**Role**: Custom theming and white-labeling
- **Purpose**: Tenant-specific customization
- **Responsibilities**:
  - Custom themes and branding management
  - Logo and color scheme customization
  - CSS and template management
  - White-label configuration

### Consumer Services (Event Processing)

#### 13. Analytics Consumer (`analytics-consumer`)
**Role**: Analytics event processing
- **Purpose**: Asynchronous analytics data processing

#### 14. Notification Consumer (`notification-consumer`)
**Role**: Notification event processing
- **Purpose**: Asynchronous notification delivery processing

#### 15. Audit Consumer (`audit-consumer`)
**Role**: Audit event processing
- **Purpose**: Asynchronous audit log processing

#### 16. Billing Consumer (`billing-consumer`)
**Role**: Billing event processing
- **Purpose**: Asynchronous billing and payment processing

### Specialized Services

#### 17. Database Service (`database-service`) ⚠️ **DEPRECATED**
**Role**: Database management and utilities
- **Port**: 8095
- **Status**: **DEPRECATED** - Not actively used by any service
- **Purpose**: Originally intended for centralized database operations
- **Note**: Each service now manages its own database connections via shared-resilience

#### 18. Event Store Service (`event-store-service`)
**Role**: Event sourcing and CQRS
- **Purpose**: Event storage and replay capabilities

#### 19. Status UI Service (`status-ui-service`)
**Role**: Status page frontend
- **Purpose**: Public status page user interface

#### 20. Shared Resilience (`shared-resilience`)
**Role**: Common patterns and utilities
- **Purpose**: Shared libraries for resilience patterns, middleware, database management, and observability

## Data Flow and Service Interactions

### Authentication Flow
1. **User Registration/Login** → User Service
2. **JWT Token Generation** → User Service
3. **Token Validation** → API Gateway → All Protected Services
4. **Multi-tenant Context** → Tenant Admin Service

### Status Page Display Flow
1. **Public Request** → API Gateway
2. **Component Data** → Component Service
3. **Incident Data** → Incident Service
4. **Status Calculation** → Component/Monitoring Services
5. **Branding Information** → Branding Service
6. **Rendered Page** → Status UI Service

### Incident Management Flow
1. **Incident Creation** → Incident Service
2. **Component Status Update** → Component Service
3. **Notification Triggering** → Notification Service
4. **Public Communication** → Incident Service → Status UI
5. **Analytics Recording** → Analytics Service

### Monitoring and Alerting Flow
1. **Health Checks** → Monitoring Service
2. **Status Updates** → Component Service
3. **Alert Generation** → Monitoring Service
4. **Notification Delivery** → Notification Consumer
5. **Incident Creation** → Incident Service (if configured)

## Deployment Architecture

### Service Distribution
- **Core Services**: High availability, horizontal scaling
- **Consumer Services**: Queue-based processing, auto-scaling
- **Database**: PostgreSQL with read replicas
- **Cache**: Redis cluster for session and data caching
- **Load Balancing**: API Gateway with circuit breakers

### Port Allocation

**HTTP Services:**
- **API Gateway**: 8080
- **User Service**: 8081
- **Component Service**: 8084
- **Notification Service**: 8085
- **Incident Service**: 8086
- **Payment Service**: 8088
- **Analytics Service**: 8090
- **Monitoring Service**: 8092
- **Status UI Service**: 8093
- **Database Service**: 8095 ⚠️ **DEPRECATED** - Not actively used
- **Event Store Service**: 8096
- **Branding Service**: 8097
- **SaaS Admin Service**: 8098
- **Tenant Admin Service**: 8099
- **Landing Page Service**: 8100

**Consumer Services (Event-Driven, no HTTP port):**
- **Analytics Consumer**: Background processor
- **Notification Consumer**: Background processor
- **Audit Consumer**: Background processor
- **Billing Consumer**: Background processor

**Metrics Ports (Prometheus):**
- Services expose metrics on ports 9090-9102 (service port + 1010)

**Shared Libraries:**
- **Shared Resilience**: Go module library (no dedicated port)

### Data Storage Strategy
- **Multi-tenant Database**: Single database with tenant-scoped data
- **Service Databases**: Each service manages its own schema
- **Data Isolation**: Complete separation of tenant data
- **Backup Strategy**: Tenant-aware backup and restoration

## Security Architecture

### Authentication & Authorization
- **JWT-based Authentication**: Stateless token authentication
- **Multi-tenant Authorization**: Tenant-scoped access control
- **Role-Based Access Control**: Granular permission system
- **API Gateway Security**: Centralized security enforcement

### Data Security
- **Tenant Isolation**: Complete data separation at database level
- **Encryption**: Data encryption at rest and in transit
- **Secure Communication**: HTTPS/TLS for all service communication
- **Input Validation**: Comprehensive request validation

### Monitoring & Audit
- **Access Logging**: Complete request and access logging
- **Audit Trails**: Comprehensive audit logging via audit consumer
- **Security Monitoring**: Security event monitoring and alerting
- **Compliance**: GDPR and SOC 2 compliance capabilities

## Scalability and Performance

### Horizontal Scaling
- **Stateless Services**: All services designed for horizontal scaling
- **Load Distribution**: API Gateway distributes load across service instances
- **Database Scaling**: Read replicas and connection pooling
- **Cache Layer**: Redis for performance optimization

### Performance Optimization
- **Circuit Breakers**: Prevent cascade failures
- **Rate Limiting**: Protect against abuse and overload
- **Caching Strategy**: Multi-layer caching for frequently accessed data
- **Database Optimization**: Query optimization and indexing

### Resilience Patterns
- **Circuit Breakers**: Automatic failure detection and recovery
- **Retry Mechanisms**: Intelligent retry with exponential backoff
- **Bulkhead Pattern**: Failure isolation between components
- **Health Checks**: Continuous health monitoring and reporting

## Development and Operations

### Development Workflow
- **Independent Services**: Each service developed and deployed independently
- **API-First Development**: API contracts defined before implementation
- **Testing Strategy**: Unit, integration, and end-to-end testing
- **Documentation**: Comprehensive service documentation

### Operations and Monitoring
- **Health Checks**: Comprehensive health monitoring
- **Metrics Collection**: Prometheus-based metrics
- **Logging**: Structured logging with correlation IDs
- **Alerting**: Multi-level alerting and escalation

### Deployment Strategy
- **Docker Containers**: Containerized deployment for all services
- **Kubernetes**: Container orchestration and management
- **CI/CD Pipeline**: Automated testing, building, and deployment
- **Blue-Green Deployment**: Zero-downtime deployment strategy

This architecture provides a robust, scalable, and maintainable platform for status page management with enterprise-grade features and multi-tenant capabilities.