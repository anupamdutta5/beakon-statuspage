# Configuration Management Guide

This guide explains how to configure all services and consumers in the StatusPage microservices architecture.

## Overview

Each service and consumer has its own configuration files located in their respective `configs/` directories. The configuration system supports:

- **Environment-based configuration**: Different configs for development, testing, and production
- **Environment variables**: For sensitive data and deployment flexibility
- **YAML format**: Human-readable and easy to modify
- **Validation**: Built-in configuration validation
- **Security**: Sensitive data handling through environment variables

## Configuration Structure

### Services Configuration

Each service has the following configuration structure:

```yaml
environment: development

service:
  name: service-name
  version: 1.0.0
  description: Service description
  tags:
    - tag1
    - tag2
  metadata:
    owner: team-name
    maintainer: email@company.com

server:
  host: 0.0.0.0
  port: 8090
  read_timeout: 30
  write_timeout: 30
  idle_timeout: 120

database:
  host: localhost
  port: 5432
  user: postgres
  password: password
  name: database_name
  ssl_mode: disable
  max_conns: 25
  min_conns: 5
  max_idle: 10
  max_lifetime: 3600

cache:
  provider: redis
  host: localhost
  port: 6379
  password: ""
  db: 0
  ttl: 300

logging:
  level: debug
  format: json
  output: stdout

jwt:
  secret: secret-key
  expiration: 24
  issuer: statuspage

monitoring:
  enabled: true
  metrics_port: 9090
  health_port: 10090
  log_level: debug
```

### Consumers Configuration

Each consumer has the following configuration structure:

```yaml
environment: development

consumer:
  name: consumer-name
  version: 1.0.0
  description: Consumer description
  tags:
    - consumer
    - event-driven
  metadata:
    owner: team-name
    maintainer: email@company.com

kafka:
  brokers:
    - localhost:9092
  group_id: consumer-group
  auto_offset_reset: earliest
  enable_auto_commit: true

database:
  host: localhost
  port: 5432
  user: postgres
  password: password
  name: database_name
  ssl_mode: disable
  max_conns: 25
  min_conns: 5
  max_idle: 10
  max_lifetime: 3600

cache:
  provider: redis
  host: localhost
  port: 6379
  password: ""
  db: 0
  ttl: 300

logging:
  level: debug
  format: json
  output: stdout

processing:
  batch_size: 100
  max_retries: 3
  retry_delay: 5s
  timeout: 30s
```

## Service Ports

Here are the default ports for all services:

| Service | Port | Description |
|---------|------|-------------|
| API Gateway | 8080 | Main entry point |
| Database Service | 8090 | Database management |
| Event Store Service | 8091 | Event storage |
| SaaS Admin Service | 8092 | SaaS administration |
| Tenant Service | 8093 | Tenant management |
| Incident Service | 8094 | Incident management |
| Notification Service | 8095 | Notifications |
| Analytics Service | 8096 | Analytics |
| Monitoring Service | 8097 | System monitoring |
| Branding Service | 8098 | Branding management |
| Tenant Admin Service | 8099 | Tenant administration |
| Landing Page Service | 8100 | Marketing landing page |
| User Service | 8001 | User management |
| Component Service | 8002 | Component management |
| Payment Service | 8003 | Payment processing |

## Database Configuration

Each service has its own database. The database names follow the pattern: `statuspage_{service_name}`

### Database Names

- `statuspage_tenant`
- `statuspage_incident`
- `statuspage_notification`
- `statuspage_analytics`
- `statuspage_monitoring`
- `statuspage_branding`
- `statuspage_database`
- `statuspage_event_store`
- `statuspage_saas_admin`
- `statuspage_tenant_admin`
- `statuspage_user`
- `statuspage_component`
- `statuspage_payment`
- `statuspage_landing`

### Consumer Database Names

- `statuspage_analytics_consumer`
- `statuspage_audit_consumer`
- `statuspage_billing_consumer`
- `statuspage_notification_consumer`

## Environment Variables

For production deployments, use environment variables for sensitive data. Copy `env.template` to `.env` and update the values:

```bash
cp env.template .env
```

### Key Environment Variables

- `DB_HOST` - Database host
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name (service-specific)
- `REDIS_HOST` - Redis host
- `REDIS_PASSWORD` - Redis password
- `JWT_SECRET` - JWT signing secret
- `KAFKA_BROKERS` - Kafka broker addresses

## Configuration Files

### Development Configuration

Development configs are located at:
```
microservices/{service-name}/configs/development.yaml
```

These configs:
- Use localhost for all services
- Have debug logging enabled
- Use development secrets
- Have relaxed security settings

### Production Configuration

Production configs are located at:
```
microservices/{service-name}/configs/production.yaml
```

These configs:
- Use environment variables for sensitive data
- Have info-level logging
- Use production secrets
- Have strict security settings

## Service-Specific Configuration

### Landing Page Service

Additional configuration for the landing page service:

```yaml
templates:
  path: web/templates
  cache: false

static:
  path: web/static
  cache_control: "public, max-age=3600"

features:
  enable_analytics: true
  enable_contact_form: true
  enable_newsletter: true
  enable_blog: true
```

### Notification Service

Additional configuration for the notification service:

```yaml
email:
  provider: smtp
  smtp_host: smtp.gmail.com
  smtp_port: 587
  smtp_username: user@example.com
  smtp_password: password
  from_email: noreply@statuspage.com

sms:
  provider: twilio
  account_sid: account-sid
  auth_token: auth-token
  from_number: "+1234567890"
```

## Loading Configuration

Services load configuration using the following priority:

1. Environment variables (highest priority)
2. Configuration file
3. Default values (lowest priority)

Example Go code for loading configuration:

```go
func LoadConfig() (*Config, error) {
    config := &Config{}
    
    // Load from file
    if err := loadFromFile(config); err != nil {
        return nil, err
    }
    
    // Override with environment variables
    if err := loadFromEnv(config); err != nil {
        return nil, err
    }
    
    return config, nil
}
```

## Validation

Configuration validation ensures that:

- Required fields are present
- Port numbers are valid
- Database connection parameters are correct
- JWT secrets are not empty
- Timeout values are reasonable

## Security Best Practices

1. **Never commit secrets to version control**
2. **Use environment variables for production**
3. **Rotate secrets regularly**
4. **Use different secrets for different environments**
5. **Encrypt sensitive data at rest**
6. **Use TLS/SSL for all connections in production**

## Troubleshooting

### Common Issues

1. **Database connection failed**
   - Check database host and port
   - Verify database exists
   - Check credentials

2. **Port already in use**
   - Check if another service is running on the same port
   - Use a different port in the configuration

3. **Configuration file not found**
   - Ensure the config file exists in the correct location
   - Check file permissions

4. **Environment variables not loaded**
   - Verify the .env file exists
   - Check environment variable names
   - Ensure the application is loading the .env file

### Debug Mode

Enable debug logging to troubleshoot configuration issues:

```yaml
logging:
  level: debug
  format: json
  output: stdout
```

## Migration to Kubernetes

When deploying to Kubernetes, convert configuration files to ConfigMaps and Secrets:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: service-config
data:
  development.yaml: |
    environment: development
    server:
      port: 8090
---
apiVersion: v1
kind: Secret
metadata:
  name: service-secrets
type: Opaque
data:
  db-password: <base64-encoded-password>
  jwt-secret: <base64-encoded-secret>
```

This configuration system provides a robust foundation for managing all aspects of the StatusPage microservices architecture.
