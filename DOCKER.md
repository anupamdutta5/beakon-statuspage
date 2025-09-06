# Enterprise Status Page - Docker Setup

This document provides comprehensive instructions for running the Enterprise Status Page using Docker and Docker Compose.

## 🚀 Quick Start

### Development Environment

1. **Clone and navigate to the project:**
   ```bash
   git clone <repository-url>
   cd statuspage
   ```

2. **Start development environment:**
   ```bash
   docker-compose -f docker-compose.dev.yml up -d
   ```

3. **Access the application:**
   - Status Page: http://localhost:8080
   - Admin Login: http://localhost:8080/admin/login
   - Default credentials: `admin` / `password`

### Production Environment

1. **Create production environment file:**
   ```bash
   cp env.production.example .env
   # Edit .env with your production values
   ```

2. **Generate SSL certificates (for HTTPS):**
   ```bash
   mkdir -p nginx/ssl
   # Add your SSL certificates to nginx/ssl/
   ```

3. **Start production environment:**
   ```bash
   docker-compose up -d
   ```

4. **Start with monitoring (optional):**
   ```bash
   docker-compose --profile monitoring up -d
   ```

## 📁 Docker Configuration Files

### Core Files
- `Dockerfile` - Production-optimized multi-stage build
- `Dockerfile.dev` - Development environment
- `docker-compose.yml` - Production services
- `docker-compose.dev.yml` - Development services

### Configuration Files
- `nginx/nginx.conf` - Nginx reverse proxy configuration
- `monitoring/prometheus.yml` - Prometheus monitoring config
- `monitoring/grafana/provisioning/` - Grafana dashboards and datasources
- `env.production.example` - Production environment template

## 🏗️ Architecture

### Services Overview

#### Core Services
- **statuspage** - Main application (Go)
- **postgres** - PostgreSQL database
- **redis** - Redis cache (optional)

#### Production Services
- **nginx** - Reverse proxy with SSL termination
- **prometheus** - Metrics collection
- **grafana** - Monitoring dashboards

### Network Architecture
```
Internet → Nginx (SSL) → Status Page App → PostgreSQL
                    ↓
                Redis Cache
                    ↓
                Prometheus → Grafana
```

## 🔧 Configuration

### Environment Variables

#### Required for Production
```bash
# Database
DB_PASSWORD=your_secure_password
JWT_SECRET=your_jwt_secret_key

# Email (if using SMTP)
SMTP_USERNAME=your-email@domain.com
SMTP_PASSWORD=your-app-password
```

#### Optional Configuration
```bash
# Server
SERVER_PORT=8080
ENVIRONMENT=production

# Monitoring
GRAFANA_PASSWORD=admin_password
```

### SSL Configuration

For production HTTPS, place your SSL certificates in:
- `nginx/ssl/cert.pem` - SSL certificate
- `nginx/ssl/key.pem` - SSL private key

## 📊 Monitoring & Observability

### Prometheus Metrics
- Application metrics: `http://localhost:9090`
- Database metrics: Available via postgres_exporter
- System metrics: Available via node_exporter

### Grafana Dashboards
- Access: `http://localhost:3000`
- Default login: `admin` / `admin` (change in .env)
- Pre-configured datasources and dashboards

### Health Checks
All services include health checks:
- Application: `/health` endpoint
- Database: `pg_isready` check
- Redis: `redis-cli ping` check

## 🚀 Deployment Commands

### Development
```bash
# Start development environment
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose -f docker-compose.dev.yml logs -f

# Stop development environment
docker-compose -f docker-compose.dev.yml down
```

### Production
```bash
# Start production environment
docker-compose up -d

# Start with monitoring
docker-compose --profile monitoring up -d

# Scale application (if needed)
docker-compose up -d --scale statuspage=3

# Update application
docker-compose build statuspage
docker-compose up -d statuspage
```

### Maintenance
```bash
# Backup database
docker-compose exec postgres pg_dump -U postgres statuspage > backup.sql

# Restore database
docker-compose exec -T postgres psql -U postgres statuspage < backup.sql

# View application logs
docker-compose logs -f statuspage

# Access application shell
docker-compose exec statuspage sh
```

## 🔒 Security Considerations

### Production Security Checklist
- [ ] Change default passwords
- [ ] Use strong JWT secrets
- [ ] Enable SSL/TLS
- [ ] Configure firewall rules
- [ ] Regular security updates
- [ ] Monitor access logs
- [ ] Use secrets management

### Network Security
- Nginx provides SSL termination
- Rate limiting on API endpoints
- Security headers configured
- Non-root user in containers

## 📈 Performance Optimization

### Resource Limits
Add to docker-compose.yml:
```yaml
services:
  statuspage:
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
```

### Caching
- Redis for session storage
- Nginx static file caching
- Database connection pooling

### Scaling
- Horizontal scaling: `docker-compose up -d --scale statuspage=3`
- Load balancing via Nginx
- Database read replicas (advanced)

## 🐛 Troubleshooting

### Common Issues

#### Application won't start
```bash
# Check logs
docker-compose logs statuspage

# Check database connection
docker-compose exec postgres pg_isready -U postgres
```

#### Database connection issues
```bash
# Verify database is running
docker-compose ps postgres

# Check database logs
docker-compose logs postgres
```

#### SSL certificate issues
```bash
# Verify certificate files exist
ls -la nginx/ssl/

# Check Nginx configuration
docker-compose exec nginx nginx -t
```

### Debug Commands
```bash
# Access application container
docker-compose exec statuspage sh

# Access database
docker-compose exec postgres psql -U postgres statuspage

# Check service health
docker-compose ps
```

## 📚 Additional Resources

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Nginx Configuration Guide](https://nginx.org/en/docs/)
- [Prometheus Monitoring](https://prometheus.io/docs/)
- [Grafana Dashboards](https://grafana.com/docs/)

## 🤝 Support

For issues and questions:
1. Check the logs: `docker-compose logs -f`
2. Verify configuration: `docker-compose config`
3. Check service health: `docker-compose ps`
4. Review this documentation

---

**Note**: This setup is production-ready but requires proper configuration of environment variables, SSL certificates, and security settings before deployment.
