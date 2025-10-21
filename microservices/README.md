# Beakon Microservices Development Scripts

## 🚀 Quick Start

Start all 19 microservices with one command:

```bash
./start-dev.sh
```

## 📋 Available Scripts

### Development Scripts (Recommended)

| Script | Purpose | Usage |
|--------|---------|-------|
| `start-dev.sh` | Start all services | `./start-dev.sh` |
| `stop-dev.sh` | Stop all services | `./stop-dev.sh` |
| `status-dev.sh` | Check service status | `./status-dev.sh` |

### Advanced Scripts

| Script | Purpose | Usage |
|--------|---------|-------|
| `start-all-services.sh` | Advanced startup with options | `./start-all-services.sh [options]` |
| `stop-all-services.sh` | Advanced stop with cleanup | `./stop-all-services.sh` |
| `check-services.sh` | Detailed service monitoring | `./check-services.sh [command]` |

## 🏗️ Service Architecture

### 19 Microservices by Category

#### Core Services (7)
- `user-service` (8080) - User management & authentication
- `api-gateway` (8081) - API routing & gateway
- `component-service` (8082) - Component monitoring
- `analytics-service` (8083) - Analytics & reporting
- `incident-service` (8084) - Incident management
- `notification-service` (8086) - Multi-channel notifications
- `payment-service` (8087) - Billing & payments

#### Infrastructure (4)
- `saas-admin-service` (8089) - SaaS administration
- `tenant-admin-service` (8090) - Multi-tenant management
- `monitoring-service` (8091) - System monitoring
- `event-store-service` (8093) - Event sourcing

#### UI Services (4)
- `branding-service` (8092) - Theme & branding
- `status-ui-service` (8094) - Public status page
- `database-service` (8095) - Database management
- `landing-page-service` (8096) - Marketing pages

#### Consumers (4)
- `analytics-consumer` (8097) - Analytics processing
- `audit-consumer` (8098) - Audit log processing
- `notification-consumer` (8099) - Notification delivery
- `billing-consumer` (8100) - Billing calculations

## 🔧 Development Workflow

### 1. Start Development Environment
```bash
# Start all services
./start-dev.sh

# Services will start automatically in dependency order
# Logs are written to logs/ directory
```

### 2. Check Service Status
```bash
# Quick status check
./status-dev.sh

# Detailed monitoring
./check-services.sh
```

### 3. Access Services
- **API Gateway**: http://localhost:8081
- **Status Page**: http://localhost:8094
- **User API**: http://localhost:8080/api/v1
- **Admin Dashboard**: http://localhost:8089

### 4. Monitor Logs
```bash
# View all logs
ls logs/

# Tail specific service
tail -f logs/user-service.log

# Search for errors
grep -i error logs/*.log
```

### 5. Stop Services
```bash
# Stop all services
./stop-dev.sh

# Clean shutdown with resource cleanup
```

## 🧪 Testing

### Basic Health Check
```bash
# Check if API Gateway is responding
curl http://localhost:8081/health

# Check user service
curl http://localhost:8080/health
```

### API Testing
```bash
# Register a user
curl -X POST http://localhost:8080/api/v1/public/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:8080/api/v1/public/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

## 📁 File Structure

```
microservices/
├── start-dev.sh              # Simple startup script
├── stop-dev.sh               # Simple stop script
├── status-dev.sh             # Simple status check
├── start-all-services.sh     # Advanced startup (with options)
├── stop-all-services.sh      # Advanced stop (with cleanup)
├── check-services.sh         # Advanced monitoring
├── logs/                     # Service logs (auto-created)
├── user-service/             # User management service
├── api-gateway/              # API gateway service
├── component-service/        # Component monitoring
├── ... (19 services total)
└── README.md                 # This file
```

## 🔧 Configuration

### Environment Variables
Services automatically use development defaults:

```bash
ENVIRONMENT=development
LOG_LEVEL=debug
DB_HOST=localhost
DB_PASSWORD=postgres
JWT_SECRET=dev-jwt-secret-do-not-use-in-production
```

### Custom Configuration
Create `.env` files in service directories to override defaults:

```bash
cd user-service
cp .env.example .env
# Edit .env with your settings
```

## 🐛 Troubleshooting

### Common Issues

#### Services Won't Start
```bash
# Check Go installation
go version

# Check for port conflicts
./status-dev.sh

# View service logs
tail -f logs/user-service.log
```

#### Permission Issues
```bash
# Fix script permissions
chmod +x *.sh

# Fix Go module permissions
go mod tidy
```

#### Database Issues
```bash
# Check PostgreSQL
pg_isready -h localhost -p 5432

# Start PostgreSQL (macOS)
brew services start postgresql

# Start PostgreSQL (Linux)
sudo systemctl start postgresql
```

### Getting Help

#### Script Help
```bash
# Basic script help
./start-dev.sh -h

# Advanced script help
./start-all-services.sh --help
./check-services.sh help
```

#### Service Status
```bash
# Quick overview
./status-dev.sh

# Detailed analysis
./check-services.sh status
./check-services.sh performance
```

## 🎯 Performance Tips

### Resource Usage
- Each service uses ~20-50MB RAM
- Total memory: ~500MB-1GB for all services
- CPU usage is minimal during development

### Optimization
```bash
# Start fewer services for development
# Edit start-dev.sh to comment out unneeded services

# Monitor resource usage
./check-services.sh performance

# Use external databases for better performance
# Edit .env files to point to dedicated PostgreSQL/Redis
```

## 🔐 Security Notes

### Development Mode
- Auto-generated JWT secrets (insecure)
- Default database passwords
- Debug logging enabled
- CORS allows all origins

### Production Mode
```bash
# Set production environment
export ENVIRONMENT=production
export JWT_SECRET=$(openssl rand -base64 32)
export DB_PASSWORD="your-secure-password"

# Use production scripts with proper configuration
```

## 📚 Additional Resources

- [Full Development Guide](../DEVELOPMENT_GUIDE.md)
- [Competitive Analysis](../COMPETITIVE_FEATURE_ANALYSIS.md)
- [Production Deployment](../FINAL_FIX_STATUS.md)

---

**Happy coding! 🚀**

Need help? Check the logs in `logs/` directory or run `./status-dev.sh` for current status.