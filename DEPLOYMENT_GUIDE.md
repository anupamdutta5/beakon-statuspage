# Beakon - Production Deployment Guide

**Last Updated**: 2025-10-14
**Target**: Kubernetes (recommended) or Docker Compose
**Prerequisites**: PostgreSQL 13+, Redis 6+, Go 1.21+

---

## Quick Start

### Local Development

```bash
# 1. Start PostgreSQL and Redis
brew services start postgresql
brew services start redis

# 2. Create databases
psql -U postgres -f scripts/create_databases.sql

# 3. Set environment variables
export ENVIRONMENT=development
export DB_HOST=localhost
export DB_USER=postgres
export DB_PASSWORD=postgres
export JWT_SECRET="your-secret-key-minimum-32-characters-long"

# 4. Start services (example)
cd microservices/tenant-admin-service
go run cmd/main.go
```

---

## Production Deployment

### Option 1: Kubernetes (Recommended)

#### Prerequisites

```bash
# Install kubectl
brew install kubectl

# Install helm
brew install helm

# Connect to your Kubernetes cluster
kubectl config use-context production
```

#### 1. Create Namespace

```bash
kubectl create namespace beakon-prod
kubectl config set-context --current --namespace=beakon-prod
```

#### 2. Create Secrets

```bash
# Create database secrets
kubectl create secret generic postgres-credentials \
  --from-literal=username=beakon_user \
  --from-literal=password='your-secure-password'

# Create JWT secret
kubectl create secret generic jwt-secret \
  --from-literal=secret='your-jwt-secret-minimum-32-characters'

# Create Redis password
kubectl create secret generic redis-credentials \
  --from-literal=password='your-redis-password'
```

#### 3. Deploy PostgreSQL (if not using managed service)

```yaml
# postgres-statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
spec:
  serviceName: postgres
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:13
        ports:
        - containerPort: 5432
        env:
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: postgres-credentials
              key: username
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-credentials
              key: password
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 100Gi
---
apiVersion: v1
kind: Service
metadata:
  name: postgres
spec:
  ports:
  - port: 5432
  clusterIP: None
  selector:
    app: postgres
```

```bash
kubectl apply -f postgres-statefulset.yaml
```

#### 4. Deploy Redis

```yaml
# redis-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:6-alpine
        ports:
        - containerPort: 6379
        args: ["--requirepass", "$(REDIS_PASSWORD)"]
        env:
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-credentials
              key: password
---
apiVersion: v1
kind: Service
metadata:
  name: redis
spec:
  ports:
  - port: 6379
  selector:
    app: redis
```

```bash
kubectl apply -f redis-deployment.yaml
```

#### 5. Deploy Microservices

Example for tenant-admin-service:

```yaml
# tenant-admin-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tenant-admin-service
  labels:
    app: tenant-admin-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: tenant-admin-service
  template:
    metadata:
      labels:
        app: tenant-admin-service
    spec:
      containers:
      - name: tenant-admin-service
        image: beakon/tenant-admin-service:latest
        ports:
        - containerPort: 8099
          name: http
        - containerPort: 9109
          name: metrics
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: SERVER_PORT
          value: "8099"
        - name: DB_HOST
          value: "postgres"
        - name: DB_PORT
          value: "5432"
        - name: DB_NAME
          value: "tenant_admin_db"
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: postgres-credentials
              key: username
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-credentials
              key: password
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: jwt-secret
              key: secret
        - name: REDIS_HOST
          value: "redis"
        - name: REDIS_PORT
          value: "6379"
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-credentials
              key: password
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8099
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8099
          initialDelaySeconds: 10
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: tenant-admin-service
spec:
  type: ClusterIP
  ports:
  - port: 8099
    targetPort: 8099
    name: http
  - port: 9109
    targetPort: 9109
    name: metrics
  selector:
    app: tenant-admin-service
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: tenant-admin-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: tenant-admin-service
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

```bash
kubectl apply -f tenant-admin-deployment.yaml
```

#### 6. Deploy API Gateway with Ingress

```yaml
# api-gateway-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-gateway
  template:
    metadata:
      labels:
        app: api-gateway
    spec:
      containers:
      - name: api-gateway
        image: beakon/api-gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: "production"
        # Add all service URLs
        - name: TENANT_ADMIN_SERVICE_URL
          value: "http://tenant-admin-service:8099"
        - name: USER_SERVICE_URL
          value: "http://user-service:8081"
        # ... other services
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: api-gateway
spec:
  type: ClusterIP
  ports:
  - port: 8080
    targetPort: 8080
  selector:
    app: api-gateway
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: beakon-ingress
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
spec:
  tls:
  - hosts:
    - api.beakon.com
    secretName: beakon-tls
  rules:
  - host: api.beakon.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-gateway
            port:
              number: 8080
```

```bash
kubectl apply -f api-gateway-deployment.yaml
```

#### 7. Deploy Monitoring Stack

```bash
# Add Prometheus Helm repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install Prometheus
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false

# Create ServiceMonitor for services
kubectl apply -f monitoring/service-monitors.yaml
```

---

### Option 2: Docker Compose (Development/Staging)

#### docker-compose.yml

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:13
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/create_databases.sql:/docker-entrypoint-initdb.d/create_databases.sql
    ports:
      - "5432:5432"

  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"

  api-gateway:
    build:
      context: ./microservices/api-gateway
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - ENVIRONMENT=staging
      - TENANT_ADMIN_SERVICE_URL=http://tenant-admin-service:8099
      - USER_SERVICE_URL=http://user-service:8081
    depends_on:
      - postgres
      - redis

  tenant-admin-service:
    build:
      context: ./microservices/tenant-admin-service
      dockerfile: Dockerfile
    ports:
      - "8099:8099"
    environment:
      - ENVIRONMENT=staging
      - DB_HOST=postgres
      - DB_USER=postgres
      - DB_PASSWORD=postgres
      - DB_NAME=tenant_admin_db
      - JWT_SECRET=development-secret-key-statuspage-2024
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    depends_on:
      - postgres
      - redis

  # Add other services...

volumes:
  postgres_data:
```

```bash
docker-compose up -d
```

---

## Environment Variables Reference

### Required (All Services)

```bash
ENVIRONMENT=production          # development, staging, production
DB_HOST=postgres
DB_PORT=5432
DB_NAME=service_database
DB_USER=beakon_user
DB_PASSWORD=secure_password
JWT_SECRET=minimum-32-character-secret-key
```

### Optional (With Defaults)

```bash
SERVER_PORT=8099                     # Service-specific
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=100
CB_DATABASE_ENABLED=true
CB_DATABASE_FAILURE_RATIO=0.6
DB_MAX_OPEN_CONNS=200               # CPU × 25 for production
LOG_LEVEL=info                       # debug, info, warn, error
```

---

## Database Initialization

### Create Databases Script

```sql
-- scripts/create_databases.sql
CREATE DATABASE tenant_admin_db;
CREATE DATABASE statuspage_user;
CREATE DATABASE statuspage_component;
CREATE DATABASE statuspage_notification;
CREATE DATABASE statuspage_incident;
CREATE DATABASE statuspage_payment;
CREATE DATABASE statuspage_analytics;
CREATE DATABASE statuspage_monitoring;
CREATE DATABASE statuspage_event_store;
CREATE DATABASE statuspage_branding;
CREATE DATABASE statuspage_landing;
CREATE DATABASE statuspage_analytics_consumer;
CREATE DATABASE statuspage_audit_consumer;
CREATE DATABASE statuspage_billing_consumer;

-- Create service users
CREATE USER tenant_admin WITH PASSWORD 'change_me_prod';
CREATE USER user_service WITH PASSWORD 'change_me_prod';
-- ... create users for each service

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE tenant_admin_db TO tenant_admin;
GRANT ALL PRIVILEGES ON DATABASE statuspage_user TO user_service;
-- ... grant permissions for each service
```

### Run Database Migrations

```bash
# Each service will auto-migrate on startup, or run manually:
cd microservices/tenant-admin-service
go run cmd/main.go --migrate-only
```

---

## Security Checklist

- [ ] Change all default passwords
- [ ] Generate strong JWT secrets (min 32 characters)
- [ ] Enable TLS/SSL for database connections
- [ ] Enable TLS/SSL for Redis connections
- [ ] Configure firewall rules (only allow necessary ports)
- [ ] Enable database encryption at rest
- [ ] Configure secure headers (CSP, HSTS)
- [ ] Set up rate limiting
- [ ] Enable audit logging
- [ ] Configure backup encryption
- [ ] Review and restrict IAM permissions
- [ ] Enable monitoring and alerting
- [ ] Set up intrusion detection
- [ ] Configure CORS properly
- [ ] Implement secrets rotation

---

## Monitoring Setup

### Prometheus Metrics

All services expose metrics on `/metrics` endpoint (port + 1010):

```yaml
# prometheus-config.yaml
scrape_configs:
  - job_name: 'tenant-admin-service'
    static_configs:
      - targets: ['tenant-admin-service:9109']
  - job_name: 'user-service'
    static_configs:
      - targets: ['user-service:9091']
  # ... add all services
```

### Grafana Dashboards

Import pre-built dashboards:
- Service Overview Dashboard
- Database Performance Dashboard
- API Gateway Dashboard
- Business Metrics Dashboard

---

## Rolling Updates

### Zero-Downtime Deployment

```bash
# Update image
kubectl set image deployment/tenant-admin-service \
  tenant-admin-service=beakon/tenant-admin-service:v1.2.0

# Monitor rollout
kubectl rollout status deployment/tenant-admin-service

# Rollback if needed
kubectl rollout undo deployment/tenant-admin-service
```

---

## Scaling

### Horizontal Scaling

```bash
# Manual scaling
kubectl scale deployment tenant-admin-service --replicas=5

# Auto-scaling (already configured in HPA)
kubectl get hpa
```

### Vertical Scaling

Update resource limits in deployment:

```yaml
resources:
  requests:
    memory: "512Mi"
    cpu: "500m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

---

## Backup Procedures

### Database Backup

```bash
# Full backup
kubectl exec -it postgres-0 -- pg_dump -U postgres tenant_admin_db > backup.sql

# Automated backup with CronJob
kubectl apply -f backup-cronjob.yaml
```

### Configuration Backup

```bash
# Backup all Kubernetes resources
kubectl get all --all-namespaces -o yaml > cluster-backup.yaml

# Backup secrets (encrypted)
kubectl get secrets --all-namespaces -o yaml > secrets-backup.yaml.enc
```

---

## Troubleshooting

### Check Service Logs

```bash
# Kubernetes
kubectl logs -f deployment/tenant-admin-service

# Docker Compose
docker-compose logs -f tenant-admin-service
```

### Check Health

```bash
# Individual service
curl http://tenant-admin-service:8099/health

# Via API Gateway
curl http://api-gateway:8080/health
```

### Common Issues

1. **Database Connection Failed**
   - Check credentials
   - Verify network connectivity
   - Check database is running

2. **Out of Memory**
   - Increase resource limits
   - Check for memory leaks
   - Scale horizontally

3. **High CPU Usage**
   - Check for infinite loops
   - Optimize slow queries
   - Scale horizontally

---

## Disaster Recovery

### Recovery Time Objectives (RTO)

- **Critical Services**: < 1 hour
- **High Priority**: < 4 hours
- **Medium Priority**: < 24 hours

### Recovery Point Objectives (RPO)

- **Critical Data**: < 15 minutes
- **High Priority**: < 1 hour
- **Medium Priority**: < 4 hours

### Recovery Procedure

1. Restore database from latest backup
2. Deploy services from last known good version
3. Verify health of all services
4. Run smoke tests
5. Gradually restore traffic
6. Monitor for issues

---

## Maintenance Windows

**Recommended Schedule**:
- Database maintenance: Sundays 2-4 AM UTC
- Service updates: Tuesday/Thursday 10 PM - 12 AM UTC
- Security patches: As needed (emergency window)

**Procedure**:
1. Notify users 48 hours in advance
2. Create maintenance incident
3. Take pre-maintenance backup
4. Perform maintenance
5. Run post-maintenance tests
6. Resolve maintenance incident
7. Post-maintenance report

---

## Cost Optimization

### Kubernetes

1. **Right-size pods**: Don't over-provision resources
2. **Use HPA**: Auto-scale based on actual usage
3. **Spot instances**: Use for non-critical workloads
4. **PVC cleanup**: Remove unused persistent volumes
5. **Image optimization**: Use multi-stage builds

### Database

1. **Connection pooling**: Use PgBouncer
2. **Read replicas**: Only when needed
3. **Archive old data**: Move to cheaper storage
4. **Optimize queries**: Reduce compute time

---

**Document Version**: 1.0
**Last Updated**: 2025-10-14
**Maintained By**: DevOps Team
**Review Frequency**: Quarterly
