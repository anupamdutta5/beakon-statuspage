# Beakon Status Page

A comprehensive microservices-based status page application with monitoring capabilities.

## 🚀 Quick Start

### Prerequisites
- Go 1.25.0+
- Docker & Docker Compose
- PostgreSQL

### Running Tests
```bash
# Make the test script executable
chmod +x run-all-tests-fixed.sh

# Run all tests (unit + integration)
./run-all-tests-fixed.sh --all

# Run only unit tests
./run-all-tests-fixed.sh --unit-only

# Run only integration tests
./run-all-tests-fixed.sh --integration

# Generate detailed test report
./run-all-tests-fixed.sh --report
```

## 📁 Project Structure

### Core Files
- `run-all-tests-fixed.sh` - **Main test runner** (handles everything)
- `docker-compose.microservices.yml` - Docker Compose for all services
- `env.template` - Environment variables template

### Microservices
- `microservices/` - All microservices and consumers
- `k8s/` - Kubernetes deployment configurations
- `docs/` - Technical documentation

## 🧪 Testing

The `run-all-tests-fixed.sh` script:
1. ✅ Checks prerequisites (Go, Docker, Docker Compose)
2. 🚀 Sets up Kafka infrastructure automatically
3. 🔧 Fixes consumer services for external Kafka
4. 🧪 Runs unit tests for all services
5. 🔗 Runs integration tests for consumer services
6. 📊 Generates detailed test reports
7. 🧹 Cleans up resources

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