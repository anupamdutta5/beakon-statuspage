#!/bin/bash

# Kubernetes deployment script for Status Page microservices
# This script deploys the entire microservices stack to Kubernetes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}[HEADER]${NC} $1"
}

# Configuration
NAMESPACE="statuspage"
ENVIRONMENT="${ENVIRONMENT:-production}"
REGISTRY="${REGISTRY:-statuspage}"
TAG="${TAG:-latest}"

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    print_error "kubectl is not installed. Please install kubectl."
    exit 1
fi

# Check if kubectl can connect to cluster
if ! kubectl cluster-info &> /dev/null; then
    print_error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
    exit 1
fi

print_header "Starting Kubernetes deployment for Status Page microservices"
print_status "Environment: $ENVIRONMENT"
print_status "Registry: $REGISTRY"
print_status "Tag: $TAG"
print_status "Namespace: $NAMESPACE"

# Create namespace and RBAC
print_header "Creating namespace and RBAC"
kubectl apply -f k8s/namespace/statuspage-namespace.yaml
print_status "Namespace and RBAC created"

# Wait for namespace to be ready
kubectl wait --for=condition=Active namespace/$NAMESPACE --timeout=60s

# Create secrets
print_header "Creating secrets"
if [ -f "k8s/secrets/statuspage-secrets.yaml" ]; then
    kubectl apply -f k8s/secrets/statuspage-secrets.yaml
    print_status "Secrets created"
else
    print_warning "Secrets file not found. Please create k8s/secrets/statuspage-secrets.yaml"
fi

# Create configmaps
print_header "Creating configmaps"
if [ -f "k8s/configmaps/statuspage-config.yaml" ]; then
    kubectl apply -f k8s/configmaps/statuspage-config.yaml
    print_status "ConfigMaps created"
else
    print_warning "ConfigMaps file not found. Please create k8s/configmaps/statuspage-config.yaml"
fi

# Deploy databases
print_header "Deploying databases"
kubectl apply -f k8s/deployments/postgresql.yaml
print_status "PostgreSQL deployed"

# Wait for databases to be ready
print_status "Waiting for databases to be ready..."
kubectl wait --for=condition=Ready pod -l app=postgresql -n $NAMESPACE --timeout=300s

# Deploy microservices
print_header "Deploying microservices"

# Deploy API Gateway
print_status "Deploying API Gateway..."
kubectl apply -f k8s/deployments/api-gateway.yaml

# Deploy User Service
print_status "Deploying User Service..."
kubectl apply -f k8s/deployments/user-service.yaml

# Deploy other services (if they exist)
for service in tenant-service component-service incident-service notification-service payment-service analytics-service monitoring-service; do
    if [ -f "k8s/deployments/$service.yaml" ]; then
        print_status "Deploying $service..."
        kubectl apply -f k8s/deployments/$service.yaml
    else
        print_warning "$service deployment file not found, skipping..."
    fi
done

# Wait for services to be ready
print_status "Waiting for services to be ready..."
kubectl wait --for=condition=Available deployment -l component=gateway -n $NAMESPACE --timeout=300s
kubectl wait --for=condition=Available deployment -l component=service -n $NAMESPACE --timeout=300s

# Deploy ingress
print_header "Deploying ingress"
kubectl apply -f k8s/ingress/statuspage-ingress.yaml
print_status "Ingress deployed"

# Deploy monitoring (if available)
print_header "Deploying monitoring"
if [ -f "k8s/monitoring/prometheus.yaml" ]; then
    kubectl apply -f k8s/monitoring/prometheus.yaml
    print_status "Prometheus deployed"
fi

if [ -f "k8s/monitoring/grafana.yaml" ]; then
    kubectl apply -f k8s/monitoring/grafana.yaml
    print_status "Grafana deployed"
fi

if [ -f "k8s/monitoring/jaeger.yaml" ]; then
    kubectl apply -f k8s/monitoring/jaeger.yaml
    print_status "Jaeger deployed"
fi

# Deploy logging (if available)
print_header "Deploying logging"
if [ -f "k8s/logging/elasticsearch.yaml" ]; then
    kubectl apply -f k8s/logging/elasticsearch.yaml
    print_status "Elasticsearch deployed"
fi

if [ -f "k8s/logging/kibana.yaml" ]; then
    kubectl apply -f k8s/logging/kibana.yaml
    print_status "Kibana deployed"
fi

# Wait for all deployments to be ready
print_status "Waiting for all deployments to be ready..."
kubectl wait --for=condition=Available deployment --all -n $NAMESPACE --timeout=600s

# Display deployment status
print_header "Deployment Status"
kubectl get pods -n $NAMESPACE
kubectl get services -n $NAMESPACE
kubectl get ingress -n $NAMESPACE

# Display useful information
print_header "Useful Commands"
echo "View pods: kubectl get pods -n $NAMESPACE"
echo "View services: kubectl get services -n $NAMESPACE"
echo "View ingress: kubectl get ingress -n $NAMESPACE"
echo "View logs: kubectl logs -f deployment/api-gateway -n $NAMESPACE"
echo "Port forward: kubectl port-forward service/api-gateway 8080:80 -n $NAMESPACE"

# Check if ingress is ready
print_status "Checking ingress status..."
if kubectl get ingress statuspage-ingress -n $NAMESPACE &> /dev/null; then
    INGRESS_IP=$(kubectl get ingress statuspage-ingress -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    if [ -n "$INGRESS_IP" ]; then
        print_status "Ingress IP: $INGRESS_IP"
        print_status "API Gateway: http://$INGRESS_IP"
        print_status "Admin Dashboard: http://admin.$INGRESS_IP"
        print_status "Status Page: http://status.$INGRESS_IP"
    else
        print_warning "Ingress IP not available yet. Please check ingress status."
    fi
fi

print_header "Deployment completed successfully!"
print_status "All microservices have been deployed to Kubernetes"
print_status "Namespace: $NAMESPACE"
print_status "Environment: $ENVIRONMENT"

# Optional: Run health checks
if [ "$1" = "--health-check" ]; then
    print_header "Running health checks..."
    
    # Check API Gateway health
    if kubectl get service api-gateway -n $NAMESPACE &> /dev/null; then
        print_status "Checking API Gateway health..."
        kubectl run health-check --image=curlimages/curl --rm -i --restart=Never -- \
            curl -f http://api-gateway.$NAMESPACE.svc.cluster.local/health || print_warning "API Gateway health check failed"
    fi
    
    # Check User Service health
    if kubectl get service user-service -n $NAMESPACE &> /dev/null; then
        print_status "Checking User Service health..."
        kubectl run health-check --image=curlimages/curl --rm -i --restart=Never -- \
            curl -f http://user-service.$NAMESPACE.svc.cluster.local/health || print_warning "User Service health check failed"
    fi
    
    print_status "Health checks completed"
fi

print_header "Deployment script completed successfully!"
