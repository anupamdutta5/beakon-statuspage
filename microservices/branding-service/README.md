# Branding Service

## Overview
The Branding Service manages tenant-specific branding, theming, and visual customization for the Beakon status page platform. It provides APIs for customizing logos, colors, fonts, and overall visual identity for each tenant's status pages and administrative interfaces.

## Key Features

### Branding Management
- **Tenant Branding**: Manage logos, colors, and brand assets per tenant
- **Theme Customization**: Custom color schemes and styling
- **Logo Management**: Upload and manage tenant logos (primary, secondary, favicon)
- **Font Configuration**: Custom font selections for status pages
- **CSS Customization**: Advanced custom CSS for power users

### Visual Customization
- **Color Palettes**: Primary, secondary, accent colors
- **Dark Mode Support**: Separate dark mode color schemes
- **Layout Options**: Customizable page layouts
- **Brand Guidelines**: Store brand asset usage guidelines

### Asset Management
- **Image Upload**: Secure image upload and storage
- **Asset CDN**: Fast delivery of brand assets
- **Asset Versioning**: Track branding changes over time
- **Multi-format Support**: PNG, SVG, JPEG for various assets

### Advanced Features
- **Circuit Breaker**: Database and external service protection
- **Connection Pooling**: Optimized database connections
- **Caching**: Redis/in-memory caching for brand assets
- **Rate Limiting**: API rate limiting for fair usage
- **Health Monitoring**: Database and service health checks

## API Endpoints

### Health & Monitoring
- `GET /health` - Health check

### Branding Management (Protected)
- `GET /api/v1/brands` - List tenant brands
- `POST /api/v1/brands` - Create/update tenant brand
- `GET /api/v1/brands/:tenant_id` - Get tenant brand details
- `PUT /api/v1/brands/:tenant_id` - Update tenant brand
- `DELETE /api/v1/brands/:tenant_id` - Delete tenant brand

### Brand Assets
- `POST /api/v1/brands/:tenant_id/logo` - Upload tenant logo
- `POST /api/v1/brands/:tenant_id/favicon` - Upload favicon
- `GET /api/v1/brands/:tenant_id/assets` - List brand assets
- `DELETE /api/v1/brands/:tenant_id/assets/:asset_id` - Delete brand asset

### Theme Management
- `GET /api/v1/brands/:tenant_id/theme` - Get theme configuration
- `PUT /api/v1/brands/:tenant_id/theme` - Update theme configuration
- `POST /api/v1/brands/:tenant_id/theme/preview` - Preview theme changes

### Public Endpoints
- `GET /public/brands/:tenant_id/css` - Public CSS for status pages
- `GET /public/brands/:tenant_id/logo` - Public logo URL
- `GET /public/brands/:tenant_id/theme` - Public theme data

## Dependencies

### Internal Services
- **tenant-admin-service**: Tenant information and validation
- **status-ui-service**: Applies branding to status pages

### External Dependencies
- **shared-resilience**: Common patterns, database, middleware
- **PostgreSQL**: Branding data storage
- **Redis**: Caching for brand assets (optional)
- **S3/Object Storage**: Brand asset storage (optional)
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
- `S3_BUCKET`: S3 bucket for brand assets (optional)
- `S3_REGION`: AWS region (optional)
- `CDN_URL`: CDN URL for asset delivery (optional)

### Feature Flags
- `CIRCUIT_BREAKER_ENABLED`: Enable circuit breaker protection
- `CACHE_ENABLED`: Enable caching layer
- `RATE_LIMIT_ENABLED`: Enable API rate limiting
- `MONITORING_ENABLED`: Enable health monitoring

## Development

### Running the Service
```bash
cd microservices/branding-service
go run cmd/main.go
```

### Building
```bash
go build -o branding-service cmd/main.go
```

### Testing
```bash
go test ./...
```

### Project Structure
```
branding-service/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── handlers/            # HTTP request handlers
│   ├── services/            # Business logic layer
│   ├── models/              # Data models
│   └── storage/             # Asset storage layer
└── go.mod                   # Go module definition
```

## Branding Schema

### Brand Configuration
```json
{
  "tenant_id": "uuid",
  "company_name": "Acme Corp",
  "tagline": "Building the Future",
  "primary_color": "#0066CC",
  "secondary_color": "#00AA00",
  "accent_color": "#FF6600",
  "text_color": "#333333",
  "background_color": "#FFFFFF",
  "logo_url": "https://cdn.example.com/logos/acme.png",
  "favicon_url": "https://cdn.example.com/favicons/acme.ico",
  "font_family": "Inter, sans-serif",
  "custom_css": ".header { border-radius: 8px; }",
  "dark_mode": {
    "enabled": true,
    "primary_color": "#3399FF",
    "background_color": "#1A1A1A",
    "text_color": "#EEEEEE"
  }
}
```

## Architecture

### Multi-Tenancy
- All branding isolated by tenant ID
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

### Asset Storage
- **Local Storage**: Development environment
- **S3/Object Storage**: Production environment
- **CDN Integration**: Fast asset delivery
- **Image Optimization**: Automatic resizing and compression

### Caching Strategy
- **Brand Data**: Cache for 5 minutes
- **Assets**: Cache for 1 hour with CDN
- **Theme CSS**: Cache for 30 minutes
- **Cache Invalidation**: On brand updates

## Branding Application

### Status Pages
1. Fetch tenant brand configuration
2. Generate custom CSS
3. Inject branding into HTML templates
4. Serve assets from CDN

### Admin Interfaces
1. Apply tenant colors to admin UI
2. Display tenant logo in header
3. Use custom fonts throughout interface

### Email Templates
1. Use tenant colors in email design
2. Include tenant logo in email header
3. Apply brand styling to buttons and links

## Monitoring

### Key Metrics
- Request throughput (requests/sec)
- Response latency (p50, p95, p99)
- Cache hit rate
- Asset upload size and frequency
- Theme update frequency

### Health Checks
- Database connectivity
- Redis connectivity (if enabled)
- S3 connectivity (if enabled)
- Circuit breaker status

## Security

### Asset Upload Security
- File type validation (PNG, SVG, JPEG only)
- File size limits (5MB max)
- Virus scanning (optional)
- Secure storage with access controls

### CSS Security
- CSS sanitization to prevent XSS
- Content Security Policy headers
- Subresource Integrity for CDN assets

## Performance

### Optimization Strategies
- Lazy loading of brand assets
- Image compression and optimization
- CSS minification
- CDN distribution for global reach
- Aggressive caching policies

### Scaling
- Horizontal scaling for API requests
- CDN for asset delivery
- Redis for distributed caching
- Database read replicas for queries

This service enables complete visual customization for each tenant, ensuring a white-label experience where tenants can maintain their brand identity across all customer-facing interfaces.
# Status Update - 2025-10-25 10:15:38

Repository synchronized and verified.
