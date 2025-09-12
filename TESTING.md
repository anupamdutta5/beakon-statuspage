# Testing Framework Documentation

This document provides comprehensive information about the testing framework for the StatusPage Pro microservices architecture.

## Overview

The testing framework is designed to provide comprehensive testing capabilities for all microservices and event consumers in the system. Each microservice has its own testing framework that can be run independently, ensuring true microservice independence.

## Architecture

### Testing Structure

```
Beakon/
├── microservices/
│   ├── api-gateway/
│   │   ├── tests/
│   │   │   ├── unit/
│   │   │   ├── integration/
│   │   │   ├── contract/
│   │   │   ├── performance/
│   │   │   └── e2e/
│   │   ├── Dockerfile.test
│   │   ├── docker-compose.test.yml
│   │   └── run-tests.sh
│   ├── user-service/
│   │   ├── tests/
│   │   │   ├── unit/
│   │   │   ├── integration/
│   │   │   ├── contract/
│   │   │   ├── performance/
│   │   │   └── e2e/
│   │   ├── Dockerfile.test
│   │   ├── docker-compose.test.yml
│   │   └── run-tests.sh
│   └── ... (other services)
├── tests/
│   └── config/
│       └── system_test_config.json
├── docker-compose.test.yml
└── run-all-tests.sh
```

## Test Types

### 1. Unit Tests

**Purpose**: Test individual functions, methods, and components in isolation.

**Location**: `microservices/{service}/tests/unit/`

**Features**:
- Fast execution
- No external dependencies
- High coverage
- Mocked dependencies
- In-memory databases for data tests

**Example**:
```bash
cd microservices/api-gateway
./run-tests.sh unit
```

### 2. Integration Tests

**Purpose**: Test the interaction between different components within a service.

**Location**: `microservices/{service}/tests/integration/`

**Features**:
- Real database connections
- External service interactions
- End-to-end workflows within a service
- Configuration validation

**Example**:
```bash
cd microservices/user-service
./run-tests.sh integration
```

### 3. Contract Tests

**Purpose**: Test the API contracts between services.

**Location**: `microservices/{service}/tests/contract/`

**Features**:
- API schema validation
- Request/response format testing
- Version compatibility testing
- Consumer-driven contract testing

**Example**:
```bash
cd microservices/api-gateway
./run-tests.sh contract
```

### 4. Performance Tests

**Purpose**: Test the performance characteristics of services.

**Location**: `microservices/{service}/tests/performance/`

**Features**:
- Load testing
- Stress testing
- Response time validation
- Throughput measurement
- Resource utilization monitoring

**Example**:
```bash
cd microservices/user-service
./run-tests.sh performance
```

### 5. End-to-End Tests

**Purpose**: Test complete user workflows across multiple services.

**Location**: `microservices/{service}/tests/e2e/`

**Features**:
- Full system integration
- User journey testing
- Cross-service communication
- Real-world scenarios

**Example**:
```bash
cd microservices/api-gateway
./run-tests.sh e2e
```

## Running Tests

### Individual Service Testing

Each microservice can be tested independently:

```bash
# Navigate to service directory
cd microservices/api-gateway

# Run all tests
./run-tests.sh

# Run specific test types
./run-tests.sh unit
./run-tests.sh integration
./run-tests.sh contract
./run-tests.sh performance
./run-tests.sh e2e

# Run with cleanup
./run-tests.sh --clean

# Run with verbose output
./run-tests.sh --verbose
```

### System-Wide Testing

Test all services and consumers:

```bash
# Run all tests
./run-all-tests.sh

# Run specific test types
./run-all-tests.sh services
./run-all-tests.sh consumers
./run-all-tests.sh e2e

# Run with coverage report
./run-all-tests.sh --coverage

# Run with cleanup
./run-all-tests.sh --clean
```

### Docker-Based Testing

Run tests in isolated containers:

```bash
# Run tests for a specific service
cd microservices/api-gateway
docker-compose -f docker-compose.test.yml up --build

# Run system-wide tests
docker-compose -f docker-compose.test.yml up --build
```

## Test Configuration

### Service-Level Configuration

Each service has its own test configuration:

```json
{
  "environment": "testing",
  "log_level": "debug",
  "timeout": "5m",
  "retry_attempts": 5,
  "retry_delay": "2s",
  "cleanup_on_exit": true,
  "parallel_tests": false,
  "coverage_report": true,
  "service": {
    "name": "api-gateway",
    "port": "8080",
    "health_path": "/health",
    "base_url": "http://localhost:8080"
  }
}
```

### System-Level Configuration

Global test configuration for the entire system:

```json
{
  "environment": "testing",
  "log_level": "debug",
  "timeout": "10m",
  "services": {
    "api-gateway": {
      "port": "8080",
      "health_path": "/health",
      "base_url": "http://localhost:8080"
    }
  },
  "test_scenarios": {
    "unit_tests": {
      "enabled": true,
      "timeout": "30s",
      "parallel": true,
      "coverage_threshold": 80
    }
  }
}
```

## Test Data Management

### Test Data Directory

Each service has a `testdata` directory for test fixtures:

```
microservices/user-service/
├── tests/
│   └── testdata/
│       ├── users.json
│       ├── tenants.json
│       └── scenarios.json
```

### Test Data Loading

```go
// Load test data
func (ts *TestSuite) LoadTestData(filename string, v interface{}) error {
    path := filepath.Join(ts.Config.TestDataDir, filename)
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("failed to read test data file %s: %w", path, err)
    }
    
    if err := json.Unmarshal(data, v); err != nil {
        return fmt.Errorf("failed to unmarshal test data: %w", err)
    }
    
    return nil
}
```

## Coverage Reporting

### Coverage Thresholds

- **Unit Tests**: 80% minimum coverage
- **Integration Tests**: 60% minimum coverage
- **Overall Coverage**: 80% minimum coverage

### Coverage Reports

Coverage reports are generated in multiple formats:

- **HTML**: Interactive coverage reports
- **XML**: For CI/CD integration
- **JSON**: For programmatic analysis

### Coverage Exclusions

The following patterns are excluded from coverage:

- `**/vendor/**`
- `**/testdata/**`
- `**/mocks/**`
- `**/cmd/**`
- `**/tests/**`

## Performance Testing

### Load Testing

- **Concurrent Users**: 100-1000 users
- **Duration**: 30 seconds to 10 minutes
- **Response Time Threshold**: 100ms
- **Throughput**: Measured in requests per second

### Stress Testing

- **Concurrent Users**: 1000+ users
- **Duration**: 10+ minutes
- **Ramp-up Time**: 2 minutes
- **Failure Threshold**: 1% error rate

### Performance Metrics

- Response time (p50, p95, p99)
- Throughput (requests/second)
- Error rate
- Resource utilization (CPU, memory, disk)

## Continuous Integration

### GitHub Actions

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Tests
        run: ./run-all-tests.sh --coverage
      - name: Upload Coverage
        uses: codecov/codecov-action@v3
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any
    stages {
        stage('Test') {
            steps {
                sh './run-all-tests.sh --coverage'
            }
        }
        stage('Coverage') {
            steps {
                publishHTML([
                    allowMissing: false,
                    alwaysLinkToLastBuild: true,
                    keepAll: true,
                    reportDir: 'test-results',
                    reportFiles: 'overall-coverage.html',
                    reportName: 'Coverage Report'
                ])
            }
        }
    }
}
```

## Best Practices

### Test Organization

1. **One test file per source file**
2. **Descriptive test names**
3. **Arrange-Act-Assert pattern**
4. **Independent tests**
5. **Clean up after tests**

### Test Data

1. **Use realistic test data**
2. **Keep test data minimal**
3. **Use factories for complex objects**
4. **Clean up test data**

### Performance

1. **Run tests in parallel when possible**
2. **Use in-memory databases for unit tests**
3. **Mock external dependencies**
4. **Keep tests fast**

### Maintenance

1. **Update tests when code changes**
2. **Remove obsolete tests**
3. **Keep test code clean**
4. **Document complex test scenarios**

## Troubleshooting

### Common Issues

1. **Port Conflicts**: Ensure test ports don't conflict with running services
2. **Database Connections**: Check database connectivity and credentials
3. **Timeout Issues**: Increase timeout values for slow tests
4. **Memory Issues**: Use smaller test datasets or increase memory limits

### Debug Mode

Enable debug mode for detailed logging:

```bash
./run-tests.sh --verbose
```

### Test Isolation

Ensure tests don't interfere with each other:

```bash
./run-tests.sh --clean
```

## Monitoring and Observability

### Test Metrics

- Test execution time
- Test success rate
- Coverage percentage
- Performance metrics

### Test Reporting

- JUnit XML reports
- HTML coverage reports
- Performance test reports
- Test execution logs

### Integration with Monitoring

- Prometheus metrics
- Grafana dashboards
- Jaeger tracing
- ELK stack logging

## Conclusion

This testing framework provides comprehensive testing capabilities for the StatusPage Pro microservices architecture. Each service can be tested independently, ensuring true microservice independence while maintaining system-wide testing capabilities.

For more information, refer to the individual service test documentation or contact the development team.

