# Configuration Management System

This document describes the comprehensive configuration management system implemented for the Status Page application, designed to eliminate hardcoded values and prepare for Kubernetes deployment.

## Overview

The configuration management system provides:
- **No hardcoded values**: All configuration is externalized
- **Environment-based configuration**: Different configs for development, testing, and production
- **Kubernetes-ready**: Easy conversion to Kubernetes ConfigMaps and Secrets
- **Flexible loading**: Support for JSON, YAML, and environment variables
- **Validation**: Built-in configuration validation
- **Security**: Sensitive data handling through secrets

## Configuration Structure

### Configuration Files

The system supports multiple configuration formats and locations:

```
configs/
├── development.json    # Development environment configuration
├── production.json     # Production environment configuration
└── testing.json        # Testing environment configuration

k8s/
├── secrets/
│   └── statuspage-secrets.yaml    # Kubernetes secrets template
└── configmaps/
    └── statuspage-config.yaml     # Kubernetes configmaps

env.template           # Environment variables template
scripts/
└── generate-k8s-secrets.sh    # Script to generate K8s secrets
```

### Configuration Categories

#### 1. Server Configuration
- Host and port settings
- Timeout configurations
- TLS/SSL settings

#### 2. Database Configuration
- Connection parameters
- Pool settings
- SSL mode configuration

#### 3. Redis Configuration
- Cache settings
- Connection pool configuration

#### 4. JWT Configuration
- Secret keys
- Token expiration settings
- Issuer and audience configuration

#### 5. Email Configuration
- SMTP settings
- SendGrid integration
- Provider selection

#### 6. Payment Configuration
- Payment provider settings
- API keys and secrets
- Webhook configurations

#### 7. Monitoring Configuration
- Prometheus settings
- Jaeger tracing configuration
- InfluxDB time-series database

#### 8. Security Configuration
- Rate limiting
- CORS settings
- Encryption configuration
- Session management

#### 9. Service Discovery Configuration
- Consul integration
- Service registration settings

#### 10. Kafka Configuration
- Message broker settings
- Topic configurations

#### 11. Tracing Configuration
- Distributed tracing settings
- Sampling rates

#### 12. Logging Configuration
- Log levels and formats
- Output destinations
- Rotation settings

## Usage

### Loading Configuration

```go
// Load configuration from file
config, err := config.LoadConfig("configs/development.json")
if err != nil {
    log.Fatal("Failed to load configuration:", err)
}

// Load configuration from environment variables only
config, err := config.LoadConfig("")
if err != nil {
    log.Fatal("Failed to load configuration:", err)
}
```

### Environment Variables

All configuration can be overridden using environment variables:

```bash
# Server configuration
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080
export SERVER_READ_TIMEOUT=10s

# Database configuration
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your_password
export DB_NAME=statuspage

# JWT configuration
export JWT_SECRET=your_jwt_secret_here
export JWT_EXPIRES_IN=24h

# And many more...
```

### Configuration Priority

Configuration is loaded in the following priority order:
1. **Environment variables** (highest priority)
2. **Configuration files** (JSON/YAML)
3. **Default values** (lowest priority)

## Kubernetes Integration

### ConfigMaps

Non-sensitive configuration is stored in Kubernetes ConfigMaps:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: statuspage-config
  namespace: statuspage
data:
  SERVER_HOST: "0.0.0.0"
  SERVER_PORT: "8080"
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  # ... more configuration
```

### Secrets

Sensitive configuration is stored in Kubernetes Secrets:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: statuspage-secrets
  namespace: statuspage
type: Opaque
data:
  db-password: <base64-encoded-password>
  jwt-secret: <base64-encoded-secret>
  stripe-secret-key: <base64-encoded-key>
  # ... more secrets
```

### Generating Kubernetes Secrets

Use the provided script to generate Kubernetes secrets from environment variables:

```bash
# Generate secrets from environment variables
./scripts/generate-k8s-secrets.sh --generate k8s/secrets/generated-secrets.yaml

# Validate generated secrets
./scripts/generate-k8s-secrets.sh --validate k8s/secrets/statuspage-secrets.yaml

# Apply secrets to Kubernetes
./scripts/generate-k8s-secrets.sh --apply k8s/secrets/statuspage-secrets.yaml --namespace production
```

## Environment-Specific Configuration

### Development Environment

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080,
    "tls": {
      "enabled": false
    }
  },
  "database": {
    "host": "localhost",
    "port": 5432,
    "ssl_mode": "disable"
  },
  "environment": "development"
}
```

### Production Environment

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080,
    "tls": {
      "enabled": true,
      "cert_file": "/etc/ssl/certs/statuspage.crt",
      "key_file": "/etc/ssl/private/statuspage.key"
    }
  },
  "database": {
    "host": "postgres-service",
    "port": 5432,
    "ssl_mode": "require"
  },
  "environment": "production"
}
```

### Testing Environment

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080,
    "tls": {
      "enabled": false
    }
  },
  "database": {
    "host": "localhost",
    "port": 5432,
    "ssl_mode": "disable"
  },
  "environment": "testing"
}
```

## Security Considerations

### Secret Management

1. **Never commit secrets to version control**
2. **Use environment variables for local development**
3. **Use Kubernetes Secrets for production**
4. **Rotate secrets regularly**
5. **Use different secrets for different environments**

### Configuration Validation

The system validates configuration on startup:

```go
func validateConfig(config *Config) error {
    // Validate required fields
    if config.JWT.Secret == "" {
        return fmt.Errorf("JWT secret is required")
    }
    if config.Database.Password == "" {
        return fmt.Errorf("database password is required")
    }
    
    // Validate ranges
    if config.Server.Port <= 0 || config.Server.Port > 65535 {
        return fmt.Errorf("invalid server port: %d", config.Server.Port)
    }
    
    return nil
}
```

### Encryption

Sensitive data is encrypted using:
- **AES-256-GCM** for data encryption
- **Base64 encoding** for Kubernetes secrets
- **TLS** for data in transit

## Microservices Configuration

### Service-Specific Configuration

Each microservice can have its own configuration:

```yaml
# User Service
user-service:
  environment:
    - DB_NAME=statuspage_users
    - SERVICE_NAME=user-service
    - CONSUL_ADDR=consul:8500

# Payment Service
payment-service:
  environment:
    - DB_NAME=statuspage_payments
    - SERVICE_NAME=payment-service
    - STRIPE_SECRET_KEY=${STRIPE_SECRET_KEY}
```

### Service Discovery

Configuration for service discovery:

```yaml
service_discovery:
  provider: consul
  consul:
    address: consul:8500
    token: ${CONSUL_TOKEN}
    dc: dc1
```

## Monitoring and Observability

### Metrics Configuration

```yaml
monitoring:
  prometheus:
    enabled: true
    port: 9090
    path: /metrics
  jaeger:
    enabled: true
    endpoint: http://jaeger:14268/api/traces
    service_name: statuspage
```

### Logging Configuration

```yaml
logging:
  level: info
  format: json
  output: stdout
  max_size: 100
  max_backups: 10
  max_age: 30
  compress: true
```

## Best Practices

### 1. Configuration Organization

- **Group related settings** together
- **Use descriptive names** for configuration keys
- **Document all configuration options**
- **Provide sensible defaults**

### 2. Environment Management

- **Use different configs** for different environments
- **Never use production configs** in development
- **Test configuration changes** in staging first
- **Use feature flags** for gradual rollouts

### 3. Security

- **Separate secrets from config** data
- **Use least privilege** for service accounts
- **Encrypt sensitive data** at rest and in transit
- **Audit configuration changes**

### 4. Deployment

- **Use infrastructure as code** for configuration
- **Automate configuration deployment**
- **Validate configuration** before deployment
- **Monitor configuration drift**

## Troubleshooting

### Common Issues

1. **Configuration not loading**
   - Check file paths and permissions
   - Verify JSON/YAML syntax
   - Check environment variable names

2. **Validation errors**
   - Ensure required fields are set
   - Check value ranges and formats
   - Verify secret values are not empty

3. **Kubernetes deployment issues**
   - Check ConfigMap and Secret names
   - Verify namespace permissions
   - Check base64 encoding of secrets

### Debugging

Enable debug logging to troubleshoot configuration issues:

```bash
export LOG_LEVEL=debug
export ENVIRONMENT=development
```

## Migration Guide

### From Hardcoded Values

1. **Identify hardcoded values** in your code
2. **Move to configuration files** or environment variables
3. **Update code** to use configuration struct
4. **Test thoroughly** in all environments
5. **Deploy gradually** with monitoring

### To Kubernetes

1. **Generate ConfigMaps** from configuration files
2. **Create Secrets** for sensitive data
3. **Update deployment manifests** to use ConfigMaps and Secrets
4. **Test in Kubernetes** environment
5. **Monitor application** behavior

## Future Enhancements

- **Configuration hot-reloading** without restart
- **Configuration versioning** and rollback
- **Configuration templates** for different deployment scenarios
- **Integration with external secret management** systems (Vault, AWS Secrets Manager)
- **Configuration drift detection** and alerting
- **Configuration backup and restore** capabilities
