# Kubernetes Deployment Guide

This document describes the Kubernetes deployment and orchestration for the Status Page microservices architecture.

## Overview

The Kubernetes deployment provides a production-ready, scalable, and resilient platform for running the Status Page microservices. It includes all necessary components for service orchestration, monitoring, logging, and security.

## Architecture

### Components

1. **Microservices** - Individual service deployments with auto-scaling
2. **Databases** - StatefulSets for persistent data storage
3. **Ingress** - External access and load balancing
4. **Monitoring** - Prometheus, Grafana, and Jaeger
5. **Logging** - Elasticsearch and Kibana
6. **Security** - RBAC, Network Policies, and Secrets

### Deployment Structure

```
k8s/
├── namespace/           # Namespace and RBAC configuration
├── deployments/         # Service deployments
├── services/           # Service definitions
├── ingress/            # Ingress configuration
├── configmaps/         # Configuration management
├── secrets/            # Secret management
├── monitoring/         # Monitoring stack
├── logging/            # Logging stack
└── scripts/            # Deployment scripts
```

## Prerequisites

### Required Tools

- **kubectl** - Kubernetes command-line tool
- **Docker** - Container runtime
- **Helm** (optional) - Package manager for Kubernetes

### Cluster Requirements

- **Kubernetes Version**: 1.24+
- **Nodes**: Minimum 3 nodes
- **CPU**: 8 cores total
- **Memory**: 16GB total
- **Storage**: 100GB for databases and logs

### Required Components

- **NGINX Ingress Controller**
- **Cert-Manager** (for SSL certificates)
- **Metrics Server** (for HPA)

## Deployment

### Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd statuspage

# Deploy to Kubernetes
./scripts/deploy-k8s.sh

# Check deployment status
kubectl get pods -n statuspage
```

### Manual Deployment

#### 1. Create Namespace and RBAC

```bash
kubectl apply -f k8s/namespace/statuspage-namespace.yaml
```

#### 2. Create Secrets

```bash
kubectl apply -f k8s/secrets/statuspage-secrets.yaml
```

#### 3. Create ConfigMaps

```bash
kubectl apply -f k8s/configmaps/statuspage-config.yaml
```

#### 4. Deploy Databases

```bash
kubectl apply -f k8s/deployments/postgresql.yaml
```

#### 5. Deploy Microservices

```bash
kubectl apply -f k8s/deployments/api-gateway.yaml
kubectl apply -f k8s/deployments/user-service.yaml
# ... other services
```

#### 6. Deploy Ingress

```bash
kubectl apply -f k8s/ingress/statuspage-ingress.yaml
```

## Configuration

### Environment Variables

```yaml
# Production configuration
ENVIRONMENT: "production"
SERVICE_NAME: "api-gateway"
SERVICE_VERSION: "1.0.0"
LOG_LEVEL: "info"
CONFIG_PATH: "/app/configs/production.json"
```

### Resource Limits

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### Health Checks

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  failureThreshold: 3
```

## Auto-Scaling

### Horizontal Pod Autoscaler (HPA)

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-gateway
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-gateway
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### Vertical Pod Autoscaler (VPA)

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: api-gateway
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-gateway
  updatePolicy:
    updateMode: "Auto"
```

## Security

### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: statuspage-network-policy
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: statuspage
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: statuspage
```

### RBAC

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: api-gateway
rules:
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "list", "watch"]
```

### Pod Security

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 2000
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  capabilities:
    drop:
    - ALL
```

## Monitoring

### Prometheus Configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
    scrape_configs:
    - job_name: 'statuspage-services'
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - statuspage
```

### Grafana Dashboards

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboards
data:
  statuspage-dashboard.json: |
    {
      "dashboard": {
        "title": "Status Page Services",
        "panels": [
          {
            "title": "Request Rate",
            "type": "graph",
            "targets": [
              {
                "expr": "rate(http_requests_total[5m])"
              }
            ]
          }
        ]
      }
    }
```

## Logging

### Elasticsearch Configuration

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearch
spec:
  serviceName: elasticsearch
  replicas: 3
  template:
    spec:
      containers:
      - name: elasticsearch
        image: elasticsearch:8.5.0
        env:
        - name: discovery.type
          value: "single-node"
        - name: xpack.security.enabled
          value: "false"
```

### Kibana Configuration

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kibana
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: kibana
        image: kibana:8.5.0
        env:
        - name: ELASTICSEARCH_HOSTS
          value: "http://elasticsearch:9200"
```

## Troubleshooting

### Common Issues

#### 1. Pods Not Starting

```bash
# Check pod status
kubectl get pods -n statuspage

# Check pod logs
kubectl logs -f deployment/api-gateway -n statuspage

# Check pod events
kubectl describe pod <pod-name> -n statuspage
```

#### 2. Services Not Accessible

```bash
# Check service status
kubectl get services -n statuspage

# Check service endpoints
kubectl get endpoints -n statuspage

# Test service connectivity
kubectl run test-pod --image=curlimages/curl --rm -i --restart=Never -- \
  curl -f http://api-gateway.statuspage.svc.cluster.local/health
```

#### 3. Ingress Issues

```bash
# Check ingress status
kubectl get ingress -n statuspage

# Check ingress controller
kubectl get pods -n ingress-nginx

# Check ingress logs
kubectl logs -f deployment/ingress-nginx-controller -n ingress-nginx
```

#### 4. Database Connection Issues

```bash
# Check database status
kubectl get pods -l app=postgresql -n statuspage

# Check database logs
kubectl logs -f statefulset/postgresql -n statuspage

# Test database connectivity
kubectl run test-db --image=postgres:15-alpine --rm -i --restart=Never -- \
  psql -h postgresql.statuspage.svc.cluster.local -U postgres -d statuspage
```

### Debugging Commands

```bash
# Get all resources
kubectl get all -n statuspage

# Check resource usage
kubectl top pods -n statuspage
kubectl top nodes

# Check events
kubectl get events -n statuspage --sort-by='.lastTimestamp'

# Port forward for local testing
kubectl port-forward service/api-gateway 8080:80 -n statuspage

# Execute commands in pod
kubectl exec -it deployment/api-gateway -n statuspage -- /bin/sh
```

## Performance Tuning

### Resource Optimization

```yaml
# Optimize resource requests and limits
resources:
  requests:
    memory: "128Mi"
    cpu: "100m"
  limits:
    memory: "256Mi"
    cpu: "200m"
```

### Node Affinity

```yaml
affinity:
  nodeAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 1
      preference:
        matchExpressions:
        - key: node-type
          operator: In
          values:
          - compute
```

### Pod Anti-Affinity

```yaml
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: app
            operator: In
            values:
            - api-gateway
        topologyKey: kubernetes.io/hostname
```

## Backup and Recovery

### Database Backup

```bash
# Create backup
kubectl exec -it statefulset/postgresql -n statuspage -- \
  pg_dump -U postgres statuspage > backup.sql

# Restore backup
kubectl exec -i statefulset/postgresql -n statuspage -- \
  psql -U postgres statuspage < backup.sql
```

### Configuration Backup

```bash
# Backup ConfigMaps
kubectl get configmaps -n statuspage -o yaml > configmaps-backup.yaml

# Backup Secrets
kubectl get secrets -n statuspage -o yaml > secrets-backup.yaml
```

## Maintenance

### Rolling Updates

```bash
# Update deployment
kubectl set image deployment/api-gateway api-gateway=statuspage/api-gateway:v2.0.0 -n statuspage

# Check rollout status
kubectl rollout status deployment/api-gateway -n statuspage

# Rollback if needed
kubectl rollout undo deployment/api-gateway -n statuspage
```

### Scaling

```bash
# Scale deployment
kubectl scale deployment api-gateway --replicas=5 -n statuspage

# Scale HPA
kubectl patch hpa api-gateway -n statuspage -p '{"spec":{"maxReplicas":20}}'
```

## Best Practices

### Security

1. **Use non-root users** - Run containers as non-root users
2. **Read-only filesystems** - Use read-only root filesystems
3. **Network policies** - Implement network segmentation
4. **RBAC** - Use least privilege access
5. **Secrets management** - Use Kubernetes secrets or external secret managers

### Performance

1. **Resource limits** - Set appropriate resource requests and limits
2. **Health checks** - Implement proper liveness and readiness probes
3. **Auto-scaling** - Use HPA and VPA for automatic scaling
4. **Node affinity** - Use node affinity for optimal placement
5. **Pod anti-affinity** - Distribute pods across nodes

### Reliability

1. **Multiple replicas** - Run multiple replicas for high availability
2. **Pod disruption budgets** - Use PDBs to maintain availability
3. **Graceful shutdown** - Implement proper shutdown handling
4. **Circuit breakers** - Use circuit breakers for resilience
5. **Monitoring** - Implement comprehensive monitoring

## Conclusion

The Kubernetes deployment provides a robust, scalable, and secure platform for running the Status Page microservices. It includes all necessary components for production deployment, monitoring, logging, and maintenance.

Key benefits:

- **High Availability** - Multiple replicas and auto-scaling
- **Security** - Network policies, RBAC, and pod security
- **Monitoring** - Comprehensive observability stack
- **Scalability** - Horizontal and vertical auto-scaling
- **Maintainability** - Easy updates and rollbacks
- **Reliability** - Health checks and graceful shutdowns

This deployment follows Kubernetes best practices and provides a solid foundation for running microservices in production.
