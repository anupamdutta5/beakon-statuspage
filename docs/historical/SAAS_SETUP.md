# SaaS StatusPage Setup Guide

This document provides a comprehensive guide for setting up the full-fledged SaaS StatusPage platform.

## Features Implemented

### ✅ Payment Integration
- **Stripe Integration**: Complete payment gateway integration with checkout sessions
- **Subscription Management**: Automated subscription lifecycle management
- **Billing Portal**: Customer portal for subscription management
- **Webhook Handling**: Real-time payment event processing

### ✅ Tenant Management
- **Multi-tenant Architecture**: Complete tenant isolation and management
- **Domain Routing**: Subdomain and custom domain support
- **Tenant Admin Dashboard**: Dedicated admin interface for each tenant
- **Public Status Pages**: Tenant-specific public status pages

### ✅ User Experience
- **Landing Page**: Professional marketing landing page
- **Responsive Design**: Mobile-first responsive design
- **Real-time Updates**: Live status updates and notifications
- **Custom Branding**: Tenant-specific branding and customization

## Environment Setup

### Required Environment Variables

Create a `.env` file in the project root with the following variables:

```bash
# Environment Configuration
ENVIRONMENT=development
SERVER_PORT=8080

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=statuspage
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=change-this-to-a-secure-secret-key
JWT_EXPIRES_IN=24

# Email Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@yourcompany.com
SMTP_USE_TLS=true
USE_SMTP=false

# Stripe Configuration
STRIPE_SECRET_KEY=sk_test_your_stripe_secret_key
STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret

# Domain Configuration
BASE_DOMAIN=localhost:8080
ADMIN_DOMAIN=admin.localhost:8080
```

### Stripe Setup

1. **Create Stripe Account**: Sign up at [stripe.com](https://stripe.com)
2. **Get API Keys**: 
   - Go to Stripe Dashboard → Developers → API Keys
   - Copy your Publishable Key and Secret Key
3. **Configure Webhooks**:
   - Go to Stripe Dashboard → Developers → Webhooks
   - Add endpoint: `https://yourdomain.com/webhooks/stripe`
   - Select events: `checkout.session.completed`, `customer.subscription.*`, `invoice.payment_*`
   - Copy the webhook secret

## Domain Architecture

### Main Platform
- **Landing Page**: `yourdomain.com` - Marketing and signup page
- **SaaS Admin**: `admin.yourdomain.com` - Platform administration

### Tenant Access
- **Tenant Status Page**: `tenant-slug.yourdomain.com` - Public status page
- **Tenant Admin**: `tenant-slug.yourdomain.com/admin` - Tenant administration
- **Custom Domains**: `status.tenantdomain.com` - Custom domain support (Pro/Enterprise)

## Database Schema

### Key Models
- **Tenant**: Multi-tenant organization data
- **Subscription**: Plan and billing information
- **Service**: Monitored services per tenant
- **Incident**: Service incidents and updates
- **Subscriber**: Email notification subscribers
- **Branding**: Tenant-specific customization

### Migration
The application automatically creates and migrates the database schema on startup.

## API Endpoints

### Tenant API (`/api/v1/tenant/`)
- `GET /info` - Tenant information
- `GET /overview` - Dashboard statistics
- `GET /services` - List services
- `POST /services` - Create service
- `GET /incidents` - List incidents
- `POST /incidents` - Create incident
- `GET /branding` - Get branding settings
- `PUT /branding` - Update branding
- `GET /billing` - Billing information
- `POST /checkout` - Create checkout session

### Payment API
- `POST /webhooks/stripe` - Stripe webhook handler
- `POST /customer-portal` - Customer portal session

## Subscription Plans

### Free Plan ($0/month)
- Up to 5 services
- Basic monitoring
- Email notifications
- Public status page

### Pro Plan ($29/month)
- Up to 25 services
- Advanced monitoring
- Custom domains
- API access
- Analytics
- Priority support

### Enterprise Plan ($99/month)
- Unlimited services
- White-label solution
- SSO integration
- Custom integrations
- SLA guarantee
- Phone support

## Getting Started

### 1. Start the Application
```bash
docker compose up --build -d
```

### 2. Access the Platform
- **Landing Page**: http://localhost:8080
- **SaaS Admin**: http://localhost:8080/admin/saas

### 3. Create Your First Tenant
1. Go to SaaS Admin Dashboard
2. Navigate to Tenants section
3. Click "Create Tenant"
4. Fill in tenant details
5. Access tenant admin at `tenant-slug.localhost:8080/admin`

### 4. Configure Services
1. Go to tenant admin dashboard
2. Navigate to Services section
3. Add your services
4. Configure monitoring

### 5. Customize Branding
1. Go to Branding section
2. Upload logo and set colors
3. Configure custom domain (Pro/Enterprise)

## Development

### Adding New Features
1. **Models**: Add to `internal/models/models.go`
2. **Services**: Create in `internal/services/`
3. **API Handlers**: Add to `internal/api/`
4. **Frontend**: Update templates and static files

### Testing
```bash
# Run tests
go test ./...

# Test specific package
go test ./internal/services
```

## Production Deployment

### Docker Deployment
```bash
# Build production image
docker build -t statuspage:latest .

# Run with production config
docker run -d \
  --name statuspage \
  -p 80:8080 \
  -e ENVIRONMENT=production \
  -e DB_HOST=your-db-host \
  -e STRIPE_SECRET_KEY=sk_live_... \
  statuspage:latest
```

### Environment Variables for Production
- Set `ENVIRONMENT=production`
- Use production database credentials
- Use live Stripe keys
- Configure proper domain settings
- Set up SSL certificates

## Security Considerations

### Authentication
- JWT-based authentication
- Secure session management
- Role-based access control

### Data Isolation
- Tenant data isolation
- Secure API endpoints
- Input validation and sanitization

### Payment Security
- Stripe handles payment processing
- No sensitive payment data stored locally
- Webhook signature verification

## Monitoring and Analytics

### Built-in Analytics
- API usage tracking
- Page view analytics
- Uptime monitoring
- Performance metrics

### Integration Options
- Google Analytics
- Custom webhook endpoints
- Export capabilities

## Support

For issues and questions:
1. Check the logs: `docker compose logs -f`
2. Review the API documentation
3. Check Stripe dashboard for payment issues
4. Verify environment configuration

## Next Steps

### Planned Features
- [ ] Advanced monitoring integrations
- [ ] Mobile app
- [ ] Advanced analytics
- [ ] Multi-language support
- [ ] Advanced notification channels

### Customization
- Custom themes and templates
- Advanced branding options
- Custom integrations
- White-label solutions
