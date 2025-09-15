# Service-Level Redis Configuration for Kubernetes Deployment

## 🎯 **Problem: Services Need Redis Config for Kubernetes**

You're absolutely right! The Docker Compose file is only for local development. For **Kubernetes deployment**, each service needs to have its own Redis configuration with proper key prefixes.

## 📊 **Current Status**

### **✅ What's Already Done:**
- **Docker Compose**: Updated with key prefixes for local development
- **Kubernetes ConfigMap**: Updated with service-specific key prefixes
- **User Service**: Updated with Redis configuration and key prefix

### **🔄 What Needs to be Done:**
- **19 services** need Redis configuration updates
- **Service config files** need Redis key prefixes
- **Kubernetes deployments** need Redis environment variables

## 🔧 **Implementation Plan**

### **1. Update Service Configuration Files**

Each service needs Redis configuration in their config files:

#### **Development Config (`configs/development.yaml`):**
```yaml
cache:
  provider: redis
  host: localhost
  port: 6379
  password: ""
  db: 0
  ttl: 300
  key_prefix: "statuspage:SERVICE_NAME"
```

#### **Production Config (`configs/production.yaml`):**
```yaml
cache:
  provider: redis
  host: ${REDIS_HOST}
  port: ${REDIS_PORT}
  password: ${REDIS_PASSWORD}
  db: 0
  ttl: 300
  key_prefix: "statuspage:SERVICE_NAME"
```

### **2. Update Service Code**

Each service needs to load Redis configuration:

```go
// internal/config/config.go
type CacheConfig struct {
    Provider  string `yaml:"provider"`
    Host      string `yaml:"host"`
    Port      int    `yaml:"port"`
    Password  string `yaml:"password"`
    DB        int    `yaml:"db"`
    TTL       int    `yaml:"ttl"`
    KeyPrefix string `yaml:"key_prefix"`
}

func Load() (*Config, error) {
    config := &Config{
        // ... other config
        Cache: CacheConfig{
            Provider:  getEnv("CACHE_PROVIDER", "redis"),
            Host:      getEnv("CACHE_HOST", "localhost"),
            Port:      getEnvInt("CACHE_PORT", 6379),
            Password:  getEnv("CACHE_PASSWORD", ""),
            DB:        getEnvInt("CACHE_DB", 0),
            TTL:       getEnvInt("CACHE_TTL", 300),
            KeyPrefix: getEnv("REDIS_KEY_PREFIX", "statuspage"),
        },
    }
    return config, nil
}
```

### **3. Update Kubernetes Deployments**

Each service deployment needs Redis environment variables:

```yaml
# k8s/deployments/SERVICE_NAME.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: SERVICE_NAME
spec:
  template:
    spec:
      containers:
      - name: SERVICE_NAME
        env:
        - name: REDIS_HOST
          valueFrom:
            configMapKeyRef:
              name: statuspage-config
              key: REDIS_HOST
        - name: REDIS_PORT
          valueFrom:
            configMapKeyRef:
              name: statuspage-config
              key: REDIS_PORT
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: statuspage-secrets
              key: REDIS_PASSWORD
        - name: REDIS_KEY_PREFIX
          valueFrom:
            configMapKeyRef:
              name: statuspage-config
              key: REDIS_KEY_PREFIX_SERVICE_NAME
```

## 📋 **Service-Specific Key Prefixes**

| Service | Key Prefix | Config File | Kubernetes Key |
|---------|------------|-------------|----------------|
| user-service | `statuspage:user-service` | ✅ Updated | `REDIS_KEY_PREFIX_USER_SERVICE` |
| tenant-service | `statuspage:tenant-service` | 🔄 Needs Update | `REDIS_KEY_PREFIX_TENANT_SERVICE` |
| incident-service | `statuspage:incident-service` | 🔄 Needs Update | `REDIS_KEY_PREFIX_INCIDENT_SERVICE` |
| component-service | `statuspage:component-service` | 🔄 Needs Update | `REDIS_KEY_PREFIX_COMPONENT_SERVICE` |
| monitoring-service | `statuspage:monitoring-service` | 🔄 Needs Update | `REDIS_KEY_PREFIX_MONITORING_SERVICE` |
| analytics-service | `statuspage:analytics-service` | 🔄 Needs Update | `REDIS_KEY_PREFIX_ANALYTICS_SERVICE` |
| notification-service | `statuspage:notification-service` | 🔄 Needs Update | `REDIS_KEY_PREFIX_NOTIFICATION_SERVICE` |
| payment-service | `statuspage:payment-service` | 🔄 Needs Update | `REDIS_KEY_PREFIX_PAYMENT_SERVICE` |
| branding-service | `statuspage:branding-service` | 🔄 Needs Update | `REDIS_KEY_PREFIX_BRANDING_SERVICE` |
| api-gateway | `statuspage:api-gateway` | 🔄 Needs Update | `REDIS_KEY_PREFIX_API_GATEWAY` |

## 🚀 **Quick Update Script**

Here's a script to update all services:

```bash
#!/bin/bash

# List of services to update
services=(
    "tenant-service"
    "incident-service"
    "component-service"
    "monitoring-service"
    "analytics-service"
    "notification-service"
    "payment-service"
    "branding-service"
    "api-gateway"
)

# Update each service
for service in "${services[@]}"; do
    echo "Updating $service..."
    
    # Update development config
    sed -i 's/port: 5432/port: 6379/g' "microservices/$service/configs/development.yaml"
    sed -i '/cache:/a\  key_prefix: "statuspage:'$service'"' "microservices/$service/configs/development.yaml"
    
    # Update production config
    sed -i 's/port: 5432/port: ${REDIS_PORT}/g' "microservices/$service/configs/production.yaml"
    sed -i 's/host: localhost/host: ${REDIS_HOST}/g' "microservices/$service/configs/production.yaml"
    sed -i 's/password: test_password/password: ${REDIS_PASSWORD}/g' "microservices/$service/configs/production.yaml"
    sed -i '/cache:/a\  key_prefix: "statuspage:'$service'"' "microservices/$service/configs/production.yaml"
    
    echo "✅ $service updated"
done

echo "🎉 All services updated!"
```

## 📊 **Benefits of Service-Level Configuration**

### **✅ Kubernetes Ready:**
- **Environment Variables**: Services can read Redis config from environment
- **ConfigMaps**: Centralized configuration management
- **Secrets**: Secure password management
- **Service Independence**: Each service manages its own Redis connection

### **✅ Development Ready:**
- **Local Development**: Services can run independently
- **Test Environments**: Each service has its own Redis config
- **Configuration Files**: Easy to modify for different environments

### **✅ Production Ready:**
- **Scalability**: Services can scale independently
- **Security**: Redis passwords in Kubernetes secrets
- **Monitoring**: Each service can be monitored separately
- **Maintenance**: Easy to update individual service configurations

## 🎯 **Next Steps**

1. **Update Service Configs**: Add Redis configuration to all 19 services
2. **Update Kubernetes Deployments**: Add Redis environment variables
3. **Test Locally**: Verify services can connect to Redis with key prefixes
4. **Deploy to Kubernetes**: Test Redis connectivity in production

This approach ensures that each service has its own Redis configuration that works both locally and in Kubernetes! 😊

