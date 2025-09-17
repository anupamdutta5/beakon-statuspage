# Beakon Status Page

A comprehensive microservices-based status page application. Each microservice is an independent Git repository.

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose

### Start All Services
```bash
# Start all microservices
./start-all.sh

# Stop all microservices
./stop-all.sh
```

## 📁 Project Structure

### Root Files (Coordination Only)
- `start-all.sh` - Start all microservices
- `stop-all.sh` - Stop all microservices
- `docker-compose.microservices.yml` - Docker orchestration
- `.env.template` - Environment configuration template

### Directories
- `microservices/` - **Independent Git repositories** for each service (20 services)
- `docs/` - Documentation and archived development files

### Recent Changes
- ✅ **Tenant services consolidated**: `tenant-service` merged into `tenant-admin-service` for unified tenant management

## 🛠️ Individual Service Development

Each microservice in `microservices/` is its own Git repository with:
- Own build system (`Makefile` or build scripts)
- Own testing framework
- Own dependencies (`go.mod`)
- Own documentation

Navigate to any service directory to work on that specific service:
```bash
cd microservices/[service-name]
# Follow that service's README for development
```

## 📊 Monitoring System

Comprehensive monitoring capabilities including:
- Component & Container Monitoring
- External Service Monitoring
- Custom Metrics Monitoring
- Status Automation
- HTTP Endpoint Health Checks
- Kubernetes Monitoring
- Docker Monitoring

## 🛠️ Development

Each microservice has its own:
- `run-tests.sh` - Service-specific test runner
- `docker-compose.test.yml` - Test environment
- `Dockerfile.test` - Test container
- `tests/` - Unit and integration tests

## 📚 Documentation

See `docs/` directory for detailed technical documentation.

## 🎯 Next Steps

1. Run tests: `./run-all-tests-fixed.sh --all`
2. Review test results and fix any issues
3. Implement monitoring system features
4. Deploy to Kubernetes using `k8s/` configurations