# Enterprise Status Page Application

A modern, enterprise-grade status page application built with Go and React, featuring 100% feature parity with Status.io and Atlassian Statuspage.

## 🚀 Features

### Core Status Page Features
- **Real-time Status Updates**: Live status monitoring with automatic updates
- **Incident Management**: Create, update, and resolve incidents with detailed timelines
- **Maintenance Windows**: Schedule and manage planned maintenance
- **Service Monitoring**: Monitor individual services and components
- **Uptime Tracking**: Track and display service uptime statistics
- **Public Status Page**: Beautiful, responsive public status page
- **RSS/Atom Feeds**: Subscribe to status updates via RSS or Atom feeds

### Enterprise & SaaS Features
- **Multi-tenant Architecture**: Support for multiple organizations with data isolation
- **Subscription Management**: Different pricing tiers (Free, Pro, Enterprise)
- **Billing Integration**: Stripe/PayPal integration for payment processing
- **Feature Toggles**: Dynamically enable/disable features based on subscription
- **Custom Branding**: Custom domains, logos, themes, and white-labeling
- **Advanced Analytics**: Uptime reports, incident analytics, and performance metrics
- **API Access**: Comprehensive REST API for automation and integrations

### Advanced Monitoring
- **Health Checks**: Automated health monitoring with configurable intervals
- **Alerting**: Custom alerts with multiple notification channels
- **Prometheus Integration**: Export metrics for external monitoring
- **Custom Metrics**: Track custom business metrics
- **Uptime Statistics**: Detailed uptime reports and analytics

### Integrations
- **Slack**: Send incident notifications to Slack channels
- **PagerDuty**: Integrate with PagerDuty for incident management
- **Opsgenie**: Connect with Opsgenie for alerting
- **Datadog**: Export metrics to Datadog
- **New Relic**: Integrate with New Relic monitoring
- **Webhooks**: Custom webhook integrations

### Advanced Notifications
- **Email Templates**: Customizable email templates for notifications
- **SMS Notifications**: Send SMS alerts via Twilio
- **Push Notifications**: Mobile push notifications
- **Webhook Notifications**: Send notifications to external services
- **Subscriber Management**: Manage notification subscribers

### Security & Compliance
- **Two-Factor Authentication (2FA)**: Enhanced security for admin accounts
- **Single Sign-On (SSO)**: SAML/OAuth integration
- **IP Whitelisting**: Restrict access by IP address
- **Audit Logs**: Comprehensive audit trail
- **Role-Based Access Control**: Granular permissions system

### Mobile App
- **iOS/Android Apps**: Native mobile applications
- **Push Notifications**: Real-time mobile notifications
- **Offline Support**: View status when offline
- **Incident Management**: Manage incidents from mobile

## 🏗️ Architecture

### Backend (Go)
- **Gin Gonic**: High-performance web framework
- **GORM**: Object-relational mapping
- **PostgreSQL**: Primary database
- **Redis**: Caching and session storage
- **JWT**: Authentication and authorization
- **Docker**: Containerization

### Frontend (React)
- **React 18**: Modern React with hooks
- **TypeScript**: Type-safe development
- **Tailwind CSS**: Utility-first CSS framework
- **React Query**: Data fetching and caching
- **React Router**: Client-side routing
- **Vite**: Fast build tool

### Infrastructure
- **Docker Compose**: Local development and deployment
- **PostgreSQL**: Relational database
- **Redis**: In-memory data store
- **Nginx**: Reverse proxy and load balancer
- **Prometheus**: Metrics collection
- **Grafana**: Metrics visualization

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.21+ (for local development)
- Node.js 18+ (for local development)

### 1. Clone and Start
```bash
    git clone <repository-url>
    cd statuspage
    docker-compose up -d
    ```

### 2. Access the Application
- **Public Status Page**: http://localhost:8080
- **Admin Dashboard**: http://localhost:8080/admin
- **API Documentation**: http://localhost:8080/api/v1/docs

### 3. Default Admin Credentials
- **Email**: admin@example.com
- **Password**: admin123

## 🛠️ Development

### Local Development Setup

1. **Start dependencies**
   ```bash
   make dev
   ```

2. **Run database migrations**
   ```bash
   make db-migrate
   ```

3. **Seed the database**
   ```bash
   make db-seed
   ```

4. **Start the backend**
   ```bash
   go run cmd/server/main.go
   ```

5. **Start the frontend** (in another terminal)
   ```bash
   cd frontend
   npm install
   npm start
   ```

### Available Make Commands

```bash
# Development
make dev              # Start development environment
make dev-build        # Build and start development environment
make dev-logs         # Show development logs
make dev-stop         # Stop development environment

# Testing
make test             # Run all tests
make test-unit        # Run unit tests only
make test-integration # Run integration tests only
make test-load        # Run load tests only
make test-security    # Run security tests only

# Database
make db-migrate       # Run database migrations
make db-seed          # Seed database with sample data
make db-reset         # Reset database (drop, create, migrate, seed)
make db-backup        # Create database backup

# Deployment
make deploy           # Deploy the application
make deploy-staging   # Deploy to staging environment
make deploy-production # Deploy to production environment
make rollback         # Rollback to previous version

# Monitoring
make monitor          # Run comprehensive health check
make monitor-app      # Check application health
make monitor-db       # Check database health
make logs             # Show application logs
make logs-follow      # Follow application logs

# Utilities
make clean            # Clean up build artifacts and containers
make format           # Format Go code
make lint             # Run linter
make vet              # Run go vet
```

## 📚 API Documentation

### Public Endpoints

#### Status
- `GET /api/v1/status` - Get current status of all services
- `GET /api/v1/services` - Get all services
- `GET /api/v1/services/:id` - Get specific service

#### Incidents
- `GET /api/v1/incidents` - Get recent incidents
- `GET /api/v1/incidents/:id` - Get specific incident

#### Maintenance
- `GET /api/v1/maintenance` - Get upcoming maintenance windows
- `GET /api/v1/maintenance/:id` - Get specific maintenance window

#### Subscribers
- `POST /api/v1/subscribers` - Subscribe to notifications
- `DELETE /api/v1/subscribers/:id` - Unsubscribe from notifications

### Admin Endpoints

#### Authentication
- `POST /api/v1/admin/login` - Admin login
- `POST /api/v1/admin/logout` - Admin logout
- `GET /api/v1/admin/me` - Get current admin user

#### Services
- `GET /api/v1/admin/services` - Get all services
- `POST /api/v1/admin/services` - Create a new service
- `GET /api/v1/admin/services/:id` - Get specific service
- `PUT /api/v1/admin/services/:id` - Update a service
- `DELETE /api/v1/admin/services/:id` - Delete a service

#### Incidents
- `GET /api/v1/admin/incidents` - Get all incidents
- `POST /api/v1/admin/incidents` - Create a new incident
- `GET /api/v1/admin/incidents/:id` - Get specific incident
- `PUT /api/v1/admin/incidents/:id` - Update an incident
- `DELETE /api/v1/admin/incidents/:id` - Delete an incident

#### Maintenance
- `GET /api/v1/admin/maintenance` - Get all maintenance windows
- `POST /api/v1/admin/maintenance` - Create a new maintenance window
- `GET /api/v1/admin/maintenance/:id` - Get specific maintenance window
- `PUT /api/v1/admin/maintenance/:id` - Update a maintenance window
- `DELETE /api/v1/admin/maintenance/:id` - Delete a maintenance window

#### Subscribers
- `GET /api/v1/admin/subscribers` - Get all subscribers
- `POST /api/v1/admin/subscribers` - Create a new subscriber
- `GET /api/v1/admin/subscribers/:id` - Get specific subscriber
- `PUT /api/v1/admin/subscribers/:id` - Update a subscriber
- `DELETE /api/v1/admin/subscribers/:id` - Delete a subscriber

#### Analytics
- `GET /api/v1/admin/analytics` - Get analytics data
- `GET /api/v1/admin/analytics/uptime` - Get uptime analytics
- `GET /api/v1/admin/analytics/incidents` - Get incident analytics
- `GET /api/v1/admin/analytics/performance` - Get performance analytics

### SaaS Admin Endpoints

#### Plans
- `GET /api/v1/admin/saas/plans` - Get all subscription plans
- `POST /api/v1/admin/saas/plans` - Create a new subscription plan
- `GET /api/v1/admin/saas/plans/:id` - Get specific subscription plan
- `PUT /api/v1/admin/saas/plans/:id` - Update a subscription plan
- `DELETE /api/v1/admin/saas/plans/:id` - Delete a subscription plan

#### Tenants
- `GET /api/v1/admin/saas/tenants` - Get all tenants
- `POST /api/v1/admin/saas/tenants` - Create a new tenant
- `GET /api/v1/admin/saas/tenants/:id` - Get specific tenant
- `PUT /api/v1/admin/saas/tenants/:id` - Update a tenant
- `DELETE /api/v1/admin/saas/tenants/:id` - Delete a tenant

#### Subscriptions
- `GET /api/v1/admin/saas/tenants/:id/subscription` - Get tenant subscription
- `POST /api/v1/admin/saas/tenants/:id/subscription` - Create tenant subscription
- `PUT /api/v1/admin/saas/tenants/:id/subscription/upgrade` - Upgrade tenant subscription
- `PUT /api/v1/admin/saas/tenants/:id/subscription/downgrade` - Downgrade tenant subscription
- `POST /api/v1/admin/saas/tenants/:id/subscription/cancel` - Cancel tenant subscription
- `GET /api/v1/admin/saas/tenants/:id/usage` - Get tenant usage metrics

#### Billing
- `GET /api/v1/admin/saas/billing/stats` - Get billing statistics
- `GET /api/v1/admin/saas/billing/customers` - Get all billing customers
- `GET /api/v1/admin/saas/billing/subscriptions` - Get all billing subscriptions
- `GET /api/v1/admin/saas/billing/invoices` - Get all invoices
- `GET /api/v1/admin/saas/billing/payments` - Get all payments

## ⚙️ Configuration

### Environment Variables

#### Database
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string

#### Authentication
- `JWT_SECRET` - JWT secret for authentication
- `ADMIN_EMAIL` - Admin user email
- `ADMIN_PASSWORD` - Admin user password

#### Server
- `PORT` - Server port (default: 8080)
- `ENV` - Environment (development, staging, production)

#### Email
- `SMTP_HOST` - SMTP server host
- `SMTP_PORT` - SMTP server port
- `SMTP_USERNAME` - SMTP username
- `SMTP_PASSWORD` - SMTP password
- `SMTP_FROM` - From email address

#### SMS
- `TWILIO_ACCOUNT_SID` - Twilio account SID
- `TWILIO_AUTH_TOKEN` - Twilio auth token
- `TWILIO_PHONE_NUMBER` - Twilio phone number

#### Billing
- `STRIPE_SECRET_KEY` - Stripe secret key
- `STRIPE_PUBLISHABLE_KEY` - Stripe publishable key
- `STRIPE_WEBHOOK_SECRET` - Stripe webhook secret
- `PAYPAL_CLIENT_ID` - PayPal client ID
- `PAYPAL_CLIENT_SECRET` - PayPal client secret

#### Monitoring
- `PROMETHEUS_URL` - Prometheus server URL
- `GRAFANA_URL` - Grafana server URL

### Configuration Files

#### Docker Compose
- `docker-compose.yml` - Main application services
- `docker-compose.test.yml` - Test environment services
- `docker-compose.prod.yml` - Production environment services

#### Environment Files
- `.env` - Local development environment
- `.env.staging` - Staging environment
- `.env.production` - Production environment

## 🚀 Deployment

### Docker Deployment

1. **Build the image**
   ```bash
   make docker-build
   ```

2. **Run the container**
   ```bash
   docker run -p 8080:8080 \
     -e DATABASE_URL="postgres://user:pass@host:port/db" \
     -e REDIS_URL="redis://host:port" \
     -e JWT_SECRET="your-secret" \
     statuspage:latest
   ```

### Docker Compose Deployment

1. **Copy environment file**
   ```bash
   cp .env.example .env
   ```

2. **Update configuration**
   Edit `.env` with your settings

3. **Deploy the application**
   ```bash
   make deploy
   ```

### Production Deployment

1. **Prepare production environment**
   ```bash
   cp .env.example .env.production
   # Edit .env.production with production settings
   ```

2. **Deploy to production**
   ```bash
   make deploy-production
   ```

### Kubernetes Deployment

1. **Create namespace**
   ```bash
   kubectl create namespace statuspage
   ```

2. **Apply configurations**
   ```bash
   kubectl apply -f k8s/ -n statuspage
   ```

3. **Check deployment**
   ```bash
   kubectl get pods -n statuspage
   ```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
make test

# Run specific test types
make test-unit        # Unit tests
make test-integration # Integration tests
make test-load        # Load tests
make test-security    # Security tests
```

### Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Load Testing

```bash
# Install k6
curl https://github.com/grafana/k6/releases/download/v0.47.0/k6-v0.47.0-linux-amd64.tar.gz -L | tar xvz --strip-components 1

# Run load tests
k6 run tests/load/status_page_load_test.js
```

## 📊 Monitoring

### Health Checks

```bash
# Check application health
make monitor

# Check specific components
make monitor-app      # Application health
make monitor-db       # Database health
make monitor-redis    # Redis health
make monitor-resources # System resources
```

### Logs

```bash
# View logs
make logs             # Application logs
make logs-follow      # Follow logs in real-time
make logs-db          # Database logs
make logs-redis       # Redis logs
```

### Metrics

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000
- **Application Metrics**: http://localhost:8080/metrics

## 🔒 Security

### Security Features

- **Two-Factor Authentication (2FA)**: Enhanced security for admin accounts
- **Single Sign-On (SSO)**: SAML/OAuth integration
- **IP Whitelisting**: Restrict access by IP address
- **Audit Logs**: Comprehensive audit trail
- **Role-Based Access Control**: Granular permissions system
- **JWT Authentication**: Secure token-based authentication
- **HTTPS Support**: SSL/TLS encryption
- **Rate Limiting**: API rate limiting
- **Input Validation**: Comprehensive input validation
- **SQL Injection Protection**: Parameterized queries

### Security Best Practices

1. **Environment Variables**: Never commit secrets to version control
2. **HTTPS**: Always use HTTPS in production
3. **Regular Updates**: Keep dependencies updated
4. **Access Control**: Implement proper access controls
5. **Monitoring**: Monitor for security events
6. **Backups**: Regular database backups
7. **Audit Logs**: Enable comprehensive audit logging

## 📱 Mobile App

### iOS App
- **App Store**: Available on the App Store
- **Features**: Status viewing, incident management, push notifications
- **Offline Support**: View status when offline

### Android App
- **Google Play**: Available on Google Play Store
- **Features**: Status viewing, incident management, push notifications
- **Offline Support**: View status when offline

## 🤝 Contributing

### Development Workflow

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```
3. **Make your changes**
4. **Add tests**
5. **Run tests**
   ```bash
   make test
   ```
6. **Commit your changes**
   ```bash
   git commit -m "Add amazing feature"
   ```
7. **Push to the branch**
   ```bash
   git push origin feature/amazing-feature
   ```
8. **Submit a pull request**

### Code Style

- **Go**: Follow Go conventions and use `gofmt`
- **React**: Follow React best practices and use ESLint
- **Tests**: Write comprehensive tests for new features
- **Documentation**: Update documentation for new features

### Pull Request Process

1. **Update documentation** for any new features
2. **Add tests** for new functionality
3. **Ensure all tests pass**
4. **Update version numbers** if applicable
5. **Request review** from maintainers

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

### Documentation
- **User Guide**: [docs/README.md](docs/README.md)
- **API Documentation**: [docs/API.md](docs/API.md)
- **Developer Guide**: [docs/DEVELOPER.md](docs/DEVELOPER.md)

### Community
- **GitHub Issues**: Report bugs and request features
- **Discussions**: Ask questions and share ideas
- **Discord**: Join our Discord community

### Professional Support
- **Email**: support@statuspage.com
- **Slack**: Join our Slack workspace
- **Enterprise Support**: Available for enterprise customers

## 🗺️ Roadmap

### Upcoming Features
- [ ] **Advanced Analytics**: More detailed analytics and reporting
- [ ] **Custom Integrations**: Plugin system for custom integrations
- [ ] **Multi-language Support**: Internationalization support
- [ ] **Advanced Theming**: More customization options
- [ ] **API Rate Limiting**: Advanced rate limiting features
- [ ] **Webhook Management**: Webhook management interface
- [ ] **Incident Templates**: Predefined incident templates
- [ ] **SLA Management**: Service level agreement tracking
- [ ] **Advanced Notifications**: More notification channels
- [ ] **Performance Optimization**: Performance improvements

### Long-term Goals
- [ ] **Microservices Architecture**: Break down into microservices
- [ ] **Kubernetes Native**: Full Kubernetes support
- [ ] **Cloud Native**: Cloud-native deployment options
- [ ] **AI-Powered Insights**: AI-powered incident analysis
- [ ] **Advanced Monitoring**: More monitoring integrations
- [ ] **Enterprise Features**: Advanced enterprise features
- [ ] **Global CDN**: Global content delivery network
- [ ] **Advanced Security**: Enhanced security features

## 🙏 Acknowledgments

- **Status.io** - Inspiration for the status page design
- **Atlassian Statuspage** - Feature reference and inspiration
- **Go Community** - Excellent Go libraries and tools
- **React Community** - Amazing React ecosystem
- **Docker Community** - Containerization tools and best practices
- **Open Source Contributors** - All the amazing open source projects we use

---

**Built with ❤️ by the Status Page Team**
