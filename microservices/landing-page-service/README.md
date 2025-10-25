# Landing Page Service

## Overview
The Landing Page Service serves the marketing and sign-up pages for the Beakon status page platform. It provides public-facing pages for product information, pricing, features, sign-up flows, and tenant onboarding.

## Key Features

### Landing Pages
- **Home Page**: Product overview and value proposition
- **Features Page**: Detailed feature descriptions
- **Pricing Page**: Pricing tiers and plan comparison
- **About Page**: Company information and mission
- **Contact Page**: Contact forms and support information

### Sign-up & Onboarding
- **Tenant Registration**: New tenant sign-up flow
- **Plan Selection**: Choose subscription plan during sign-up
- **Email Verification**: Verify email during registration
- **Onboarding Wizard**: Guide new tenants through setup
- **Trial Management**: Free trial activation and tracking

### Marketing Features
- **SEO Optimization**: Meta tags, structured data, sitemap
- **Analytics Integration**: Google Analytics, tracking pixels
- **A/B Testing**: Test different page variations
- **Lead Capture**: Newsletter subscriptions, demo requests
- **Blog Integration**: Marketing blog and content

### Static Asset Management
- **Static File Serving**: Serve CSS, JavaScript, images
- **CDN Integration**: Serve assets from CDN for performance
- **Image Optimization**: Compressed and responsive images
- **Cache Control**: Proper caching headers for assets

## API Endpoints

### Public Pages
- `GET /` - Home page
- `GET /features` - Features page
- `GET /pricing` - Pricing page
- `GET /about` - About page
- `GET /contact` - Contact page
- `GET /blog` - Blog listing
- `GET /blog/:slug` - Blog post

### Sign-up & Authentication
- `GET /signup` - Sign-up page
- `POST /api/v1/signup` - Submit sign-up form
- `GET /verify-email/:token` - Email verification
- `POST /api/v1/contact` - Submit contact form
- `POST /api/v1/newsletter` - Newsletter subscription

### Static Assets
- `GET /static/*` - Static files (CSS, JS, images)
- `GET /assets/*` - Asset files
- `GET /favicon.ico` - Favicon
- `GET /robots.txt` - Search engine robots file
- `GET /sitemap.xml` - XML sitemap

### Health & Monitoring
- `GET /health` - Health check

## Dependencies

### Internal Services
- **tenant-admin-service**: Create new tenants during sign-up
- **user-service**: Create initial admin user
- **payment-service**: Initialize subscription and billing

### External Dependencies
- **Zap Logger**: Structured logging
- **Template Engine**: HTML template rendering
- **Email Service**: Send verification and welcome emails
- **Analytics**: Google Analytics, Mixpanel

## Configuration

### Environment Variables
- `SERVER_PORT`: HTTP server port (default: 8080)
- `SERVER_HOST`: HTTP server host
- `ENVIRONMENT`: Runtime environment (development/production)
- `TENANT_ADMIN_SERVICE_URL`: Tenant admin service endpoint
- `USER_SERVICE_URL`: User service endpoint
- `PAYMENT_SERVICE_URL`: Payment service endpoint
- `SMTP_HOST`: Email SMTP host
- `SMTP_PORT`: Email SMTP port
- `SMTP_USER`: Email SMTP username
- `SMTP_PASSWORD`: Email SMTP password
- `FROM_EMAIL`: Email from address
- `GOOGLE_ANALYTICS_ID`: Google Analytics tracking ID
- `RECAPTCHA_SECRET`: reCAPTCHA secret key

### Feature Flags
- `ENABLE_SIGNUP`: Enable tenant sign-up
- `ENABLE_FREE_TRIAL`: Enable free trial registration
- `ENABLE_BLOG`: Enable blog section
- `ENABLE_ANALYTICS`: Enable analytics tracking

## Development

### Running the Service
```bash
cd microservices/landing-page-service
go run cmd/main.go
```

### Building
```bash
go build -o landing-page-service cmd/main.go
```

### Testing
```bash
go test ./...
```

### Project Structure
```
landing-page-service/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── server/              # HTTP server
│   ├── handlers/            # HTTP request handlers
│   ├── services/            # Business logic
│   └── templates/           # HTML templates
├── web/
│   ├── static/              # Static assets (CSS, JS, images)
│   └── templates/           # HTML templates
├── pkg/
│   └── logger/              # Logging utilities
└── go.mod                   # Go module definition
```

## Sign-up Flow

### Registration Process
1. User visits sign-up page
2. User fills out registration form (company name, email, password)
3. User selects subscription plan
4. Form submitted via POST /api/v1/signup
5. Service creates tenant via tenant-admin-service
6. Service creates admin user via user-service
7. Service initializes subscription via payment-service
8. Verification email sent to user
9. User clicks verification link
10. User redirected to tenant admin dashboard

### Email Verification
- Verification token generated and stored
- Email sent with verification link
- Link expires after 24 hours
- User can request new verification email

## Template System

### Template Structure
```
templates/
├── layouts/
│   ├── base.html           # Base layout
│   └── simple.html         # Simple layout (no nav)
├── pages/
│   ├── home.html
│   ├── features.html
│   ├── pricing.html
│   └── signup.html
└── partials/
    ├── header.html
    ├── footer.html
    └── nav.html
```

### Template Data
- Page metadata (title, description)
- Feature flags
- Pricing information
- Analytics tracking codes

## SEO Optimization

### Meta Tags
- Page titles optimized for search
- Meta descriptions for all pages
- Open Graph tags for social sharing
- Twitter Card tags

### Structured Data
- Organization schema
- Product schema for pricing
- FAQ schema
- BreadcrumbList schema

### Performance
- Minified CSS and JavaScript
- Compressed images (WebP with fallbacks)
- Lazy loading for images
- CDN for static assets

## Forms & Validation

### Sign-up Form Validation
- Email format validation
- Password strength requirements
- Company name required
- reCAPTCHA verification
- Duplicate email check

### Contact Form
- Name, email, message required
- Spam protection via reCAPTCHA
- Email notification to support team

## Analytics & Tracking

### Event Tracking
- Page views
- Sign-up starts
- Sign-up completions
- Pricing page views
- Feature page interactions

### Conversion Tracking
- Sign-up conversion rate
- Plan selection distribution
- Trial-to-paid conversion

## Monitoring

### Key Metrics
- Page load time
- Sign-up conversion rate
- Form submission errors
- Email delivery rate
- 404 error rate

### Health Checks
- HTTP server status
- External service connectivity
- Template rendering test

## Security

### Form Protection
- CSRF tokens on all forms
- reCAPTCHA on sign-up and contact forms
- Rate limiting on form submissions
- Input sanitization and validation

### Email Security
- Email verification required
- Prevent email enumeration
- Rate limit verification attempts

## Internationalization (Future)

### Multi-language Support
- Language detection from browser
- Translation files for all text
- Currency conversion for pricing
- Date/time localization

## A/B Testing (Future)

### Test Variations
- Different pricing page layouts
- Sign-up form variations
- CTA button colors and text
- Feature descriptions

This service is the first touchpoint for potential customers, providing an attractive and functional entry point into the platform while collecting leads and facilitating tenant onboarding.
# Status Update - 2025-10-25 10:15:33

Repository synchronized and verified.
