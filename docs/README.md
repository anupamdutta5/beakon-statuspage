# Status Page Platform - Enterprise SaaS Solution

A comprehensive, enterprise-grade status page platform built with Go and React, offering 100% feature parity with Status.io and Atlassian Statuspage.

## 🚀 Features

### Core Status Page Features
- **Real-time Status Monitoring**: Monitor services, APIs, and infrastructure components
- **Incident Management**: Create, update, and resolve incidents with detailed timelines
- **Maintenance Scheduling**: Plan and communicate scheduled maintenance windows
- **Uptime Tracking**: Comprehensive uptime statistics and historical data
- **Public Status Pages**: Beautiful, customizable public status pages
- **Private Pages**: Access-controlled status pages for internal use

### SaaS & Multi-Tenancy
- **Multi-Tenant Architecture**: Complete tenant isolation with dedicated resources
- **Subscription Management**: Flexible pricing tiers (Free, Pro, Enterprise)
- **Billing Integration**: Stripe/PayPal payment processing with automated billing
- **Feature Flags**: Granular control over tenant features and capabilities
- **Usage Tracking**: Monitor tenant resource usage and enforce limits

### Advanced Monitoring
- **Prometheus Integration**: Native Prometheus metrics collection and alerting
- **Custom Metrics**: Define and track custom business metrics
- **Health Checks**: Automated health monitoring with configurable intervals
- **Alert Management**: Create and manage alert rules with multiple conditions
- **Uptime Reports**: Detailed uptime analytics and SLA reporting

### Third-Party Integrations
- **Slack**: Real-time incident notifications and status updates
- **PagerDuty**: Critical incident escalation and on-call management
- **Opsgenie**: Advanced alerting and incident response
- **Datadog**: Infrastructure monitoring and APM integration
- **New Relic**: Application performance monitoring
- **Webhooks**: Custom webhook integrations for any system

### Advanced Notifications
- **Email Templates**: Customizable HTML email templates for all notifications
- **SMS Notifications**: SMS alerts via Twilio and AWS SNS
- **Push Notifications**: Mobile push notifications via FCM and APNS
- **Webhook Notifications**: Real-time webhook delivery for external systems
- **Notification Preferences**: Granular user notification preferences

### Custom Branding & White-Labeling
- **Custom Domains**: Dedicated domains with SSL certificate management
- **Theme Customization**: Complete control over colors, fonts, and styling
- **Logo Management**: Upload and manage custom logos and favicons
- **White-Labeling**: Remove all platform branding for enterprise customers
- **Custom CSS**: Advanced styling capabilities for complete customization

### REST API & Automation
- **Comprehensive API**: Full REST API for all platform operations
- **API Key Management**: Secure API key generation and management
- **Rate Limiting**: Configurable rate limiting and usage tracking
- **Webhook Management**: Create and manage webhook endpoints
- **API Documentation**: Interactive API documentation with examples

### Analytics & Reporting
- **Page Analytics**: Detailed visitor analytics and engagement metrics
- **Uptime Reports**: Comprehensive uptime statistics and SLA tracking
- **Incident Analytics**: Incident frequency, duration, and impact analysis
- **Performance Metrics**: Response time, throughput, and error rate tracking
- **Custom Dashboards**: Build custom analytics dashboards

### Security & Compliance
- **Two-Factor Authentication**: TOTP-based 2FA for enhanced security
- **Single Sign-On**: SAML, OAuth2, and OIDC integration
- **IP Whitelisting**: Restrict access to specific IP addresses or ranges
- **Audit Logging**: Comprehensive audit trail for all platform activities
- **Security Policies**: Configurable password and session policies

### Mobile App
- **Native Mobile Apps**: iOS and Android apps for status monitoring
- **Push Notifications**: Real-time mobile notifications for incidents
- **Offline Support**: View cached status information when offline
- **Biometric Authentication**: Secure access with fingerprint/face ID
- **Mobile Analytics**: Track mobile app usage and engagement

## 🏗️ Architecture

### Technology Stack
- **Backend**: Go with Gin framework
- **Database**: PostgreSQL with GORM ORM
- **Cache**: Redis for session management and caching
- **Frontend**: HTML/CSS/JavaScript with Chart.js
- **Containerization**: Docker and Docker Compose
- **Monitoring**: Prometheus integration
- **Notifications**: Email, SMS, and push notification services

### Database Schema
The platform uses a comprehensive multi-tenant database schema with the following key entities:

- **Tenants**: Customer organizations with isolated data
- **Users**: Platform users with role-based access control
- **Services**: Monitored services and components
- **Incidents**: Service incidents with detailed timelines
- **Subscriptions**: Tenant subscription plans and billing
- **Integrations**: Third-party service integrations
- **Analytics**: Usage and performance analytics data

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.21+ (for development)
- PostgreSQL 13+ (for development)
- Redis 6+ (for development)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/your-org/statuspage-platform.git
   cd statuspage-platform
   ```

2. **Start the application**
   ```bash
   docker-compose up -d
   ```

3. **Access the application**
   - Public Status Page: http://localhost:8080
   - Admin Dashboard: http://localhost:8080/admin/login
   - SaaS Admin: http://localhost:8080/admin/saas

### Default Credentials
- **Username**: admin
- **Password**: password

## 📚 API Documentation

### Authentication
The API uses JWT-based authentication. Obtain a token by logging in:

```bash
curl -X POST http://localhost:8080/api/v1/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

### Core Endpoints

#### Services
- `GET /api/v1/status` - Get all services status
- `POST /api/v1/admin/status` - Create a new service
- `PUT /api/v1/admin/status/:id` - Update a service
- `DELETE /api/v1/admin/status/:id` - Delete a service

#### Incidents
- `GET /api/v1/incidents` - Get all incidents
- `POST /api/v1/admin/incidents` - Create a new incident
- `PUT /api/v1/admin/incidents/:id` - Update an incident
- `DELETE /api/v1/admin/incidents/:id` - Delete an incident

#### SaaS Management
- `GET /api/v1/admin/saas/tenants` - Get all tenants
- `POST /api/v1/admin/saas/tenants` - Create a new tenant
- `GET /api/v1/admin/saas/plans` - Get subscription plans
- `POST /api/v1/admin/saas/plans` - Create a subscription plan

### Webhook Events
The platform supports webhooks for the following events:
- `incident.created` - New incident created
- `incident.updated` - Incident status updated
- `incident.resolved` - Incident resolved
- `maintenance.scheduled` - Maintenance scheduled
- `maintenance.started` - Maintenance started
- `maintenance.completed` - Maintenance completed

## 🔧 Configuration

### Environment Variables
```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=statuspage
DB_PASSWORD=password
DB_NAME=statuspage

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT
JWT_SECRET=your-secret-key

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-password

# Stripe (for billing)
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
```

### Docker Configuration
The application is fully containerized with Docker Compose:

```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=db
      - REDIS_HOST=redis
    depends_on:
      - db
      - redis

  db:
    image: postgres:13
    environment:
      - POSTGRES_DB=statuspage
      - POSTGRES_USER=statuspage
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:6-alpine
    volumes:
      - redis_data:/data
```

## 🧪 Testing

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./...
```

### Test Database
Tests use a separate test database. Set the following environment variable:
```bash
TEST_DB_URL=postgres://user:password@localhost/statuspage_test
```

## 🚀 Deployment

### Production Deployment
1. **Set up production environment variables**
2. **Configure SSL certificates**
3. **Set up monitoring and logging**
4. **Configure backup strategies**
5. **Deploy using Docker or Kubernetes**

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: statuspage-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: statuspage-app
  template:
    metadata:
      labels:
        app: statuspage-app
    spec:
      containers:
      - name: app
        image: statuspage:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: "postgres-service"
        - name: REDIS_HOST
          value: "redis-service"
```

## 📊 Monitoring

### Health Checks
- `GET /health` - Application health check
- `GET /metrics` - Prometheus metrics endpoint

### Logging
The application uses structured logging with Zap:
- **Development**: Human-readable console output
- **Production**: JSON format for log aggregation

### Metrics
Key metrics tracked:
- Request rate and latency
- Database connection pool status
- Cache hit/miss ratios
- Error rates by endpoint
- Business metrics (incidents, uptime, etc.)

## 🔒 Security

### Security Features
- JWT-based authentication
- Role-based access control (RBAC)
- Two-factor authentication (2FA)
- Single sign-on (SSO) support
- IP whitelisting
- Audit logging
- Rate limiting
- Input validation and sanitization

### Security Best Practices
- Use HTTPS in production
- Regularly update dependencies
- Implement proper secret management
- Monitor for security vulnerabilities
- Follow OWASP guidelines

## 🤝 Contributing

### Development Setup
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

### Code Style
- Follow Go conventions
- Use `gofmt` for formatting
- Add comments for public functions
- Write comprehensive tests

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

### Documentation
- [API Documentation](docs/api.md)
- [Deployment Guide](docs/deployment.md)
- [Configuration Reference](docs/configuration.md)
- [Troubleshooting Guide](docs/troubleshooting.md)

### Community
- [GitHub Issues](https://github.com/your-org/statuspage-platform/issues)
- [Discord Community](https://discord.gg/your-community)
- [Email Support](mailto:support@your-domain.com)

### Enterprise Support
For enterprise customers, we offer:
- Priority support
- Custom integrations
- On-premise deployment
- Training and consulting
- SLA guarantees

Contact us at [enterprise@your-domain.com](mailto:enterprise@your-domain.com) for more information.
