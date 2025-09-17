# Microservice Testing Guide

This document describes the comprehensive testing strategies for the Status Page microservices architecture, including unit testing, integration testing, and contract testing.

## Overview

The microservice testing system provides comprehensive testing capabilities for all aspects of the microservices architecture. It includes unit testing for individual components, integration testing for service interactions, and contract testing for API compatibility.

## Testing Strategy

### Testing Pyramid

```
                    /\
                   /  \
                  / E2E \     (End-to-End Tests)
                 /______\
                /        \
               /Integration\  (Integration Tests)
              /____________\
             /              \
            /   Unit Tests   \  (Unit Tests)
           /__________________\
```

### Testing Levels

1. **Unit Tests** - Test individual components in isolation
2. **Integration Tests** - Test service interactions and dependencies
3. **Contract Tests** - Test API contracts between services
4. **End-to-End Tests** - Test complete user workflows

## Unit Testing

### Test Suite Structure

```go
// Create a test suite
suite := TestSuite{
    Name: "User Service Tests",
    Setup: func() error {
        // Setup test environment
        return nil
    },
    Teardown: func() error {
        // Cleanup test environment
        return nil
    },
    Tests: []TestCase{
        {
            Name: "Create User",
            Test: func(t *testing.T) error {
                // Test implementation
                return nil
            },
        },
    },
}

// Run the test suite
runner := NewTestRunner(TestConfig{
    Timeout: 30 * time.Second,
    Verbose: true,
})
runner.AddSuite(suite)
runner.Run(t)
```

### Mock Services

```go
// Create a mock service
mockUserService := NewMockService("user-service")

// Set expectations
mockUserService.Expect("GetUser", []interface{}{"123"}, []interface{}{user}, nil)
mockUserService.Expect("CreateUser", []interface{}{user}, []interface{}{"456"}, nil)

// Use mock in tests
user, err := mockUserService.Call("GetUser", []interface{}{"123"})
if err != nil {
    t.Errorf("Unexpected error: %v", err)
}

// Verify expectations
if err := mockUserService.Verify(); err != nil {
    t.Errorf("Mock verification failed: %v", err)
}
```

### Test Helpers

```go
// Create test helper
helper := NewTestHelper()

// Create mocks
mockDB := helper.CreateMock("database")
mockCache := helper.CreateMock("cache")

// Run tests
// ...

// Verify all mocks
if err := helper.VerifyAllMocks(); err != nil {
    t.Errorf("Mock verification failed: %v", err)
}
```

### Assertions

```go
// Basic assertions
AssertEqual(t, expected, actual, "Values should be equal")
AssertNotEqual(t, expected, actual, "Values should not be equal")
AssertNil(t, value, "Value should be nil")
AssertNotNil(t, value, "Value should not be nil")
AssertTrue(t, condition, "Condition should be true")
AssertFalse(t, condition, "Condition should be false")

// Error assertions
AssertError(t, err, "Expected error")
AssertNoError(t, err, "Unexpected error")
```

## Integration Testing

### Service Configuration

```go
// Define service configurations
services := []ServiceConfig{
    {
        Name: "user-service",
        Image: "statuspage/user-service:latest",
        Port: 8081,
        Environment: map[string]string{
            "DATABASE_URL": "postgres://localhost:5432/test_users",
            "LOG_LEVEL": "debug",
        },
        HealthCheck: HealthCheck{
            Path:     "/health",
            Interval: 1 * time.Second,
            Timeout:  5 * time.Second,
            Retries:  3,
        },
    },
    {
        Name: "api-gateway",
        Image: "statuspage/api-gateway:latest",
        Port: 8080,
        DependsOn: []string{"user-service"},
        HealthCheck: HealthCheck{
            Path:     "/health",
            Interval: 1 * time.Second,
            Timeout:  5 * time.Second,
            Retries:  3,
        },
    },
}
```

### Integration Test Suite

```go
// Create integration test suite
suite := IntegrationTestSuite{
    Name: "User Service Integration Tests",
    Services: services,
    Setup: func() error {
        // Setup test environment
        return nil
    },
    Teardown: func() error {
        // Cleanup test environment
        return nil
    },
    Tests: []IntegrationTestCase{
        {
            Name: "Create and Get User",
            Test: func(t *testing.T, services map[string]*Service) error {
                // Get service clients
                userService := NewServiceClient(services["user-service"])
                apiGateway := NewServiceClient(services["api-gateway"])

                // Create user
                userData := map[string]interface{}{
                    "name":  "John Doe",
                    "email": "john@example.com",
                }
                resp, err := apiGateway.Post("/users", userData)
                if err != nil {
                    return err
                }
                defer resp.Body.Close()

                if resp.StatusCode != 201 {
                    return fmt.Errorf("expected status 201, got %d", resp.StatusCode)
                }

                // Get user
                resp, err = apiGateway.Get("/users/123")
                if err != nil {
                    return err
                }
                defer resp.Body.Close()

                if resp.StatusCode != 200 {
                    return fmt.Errorf("expected status 200, got %d", resp.StatusCode)
                }

                return nil
            },
        },
    },
}

// Run integration tests
runner := NewIntegrationTestRunner(IntegrationTestConfig{
    Timeout: 5 * time.Minute,
    Cleanup: true,
})
runner.AddSuite(suite)
runner.Run(t)
```

### Service Client

```go
// Create service client
client := NewServiceClient(service)

// Make requests
resp, err := client.Get("/users")
resp, err := client.Post("/users", userData)
resp, err := client.Put("/users/123", userData)
resp, err := client.Delete("/users/123")

// Health checks
if err := client.HealthCheck(); err != nil {
    t.Errorf("Service not healthy: %v", err)
}

// Wait for service to be ready
if err := client.WaitForReady(30 * time.Second); err != nil {
    t.Errorf("Service not ready: %v", err)
}
```

### Integration Test Helper

```go
// Create integration test helper
helper := NewIntegrationTestHelper(services)

// Get service clients
userService := helper.GetService("user-service")
apiGateway := helper.GetService("api-gateway")

// Wait for all services to be ready
if err := helper.WaitForAllServices(30 * time.Second); err != nil {
    t.Errorf("Services not ready: %v", err)
}

// Health check all services
if err := helper.HealthCheckAllServices(); err != nil {
    t.Errorf("Services not healthy: %v", err)
}
```

## Contract Testing

### Contract Definition

```go
// Create a contract
contract := NewContractBuilder("user-service", "user-service", "api-gateway").
    WithDescription("User service contract").
    AddInteraction(
        NewInteractionBuilder("Get user by ID").
            WithState("user exists").
            WithRequest(
                NewRequestBuilder("GET", "/users/123").
                    WithHeader("Authorization", "Bearer token").
                    Build(),
            ).
            WithResponse(
                NewResponseBuilder(200).
                    WithHeader("Content-Type", "application/json").
                    WithBody(map[string]interface{}{
                        "id":    "123",
                        "name":  "John Doe",
                        "email": "john@example.com",
                    }).
                    Build(),
            ).
            Build(),
    ).
    Build()
```

### Contract Test Suite

```go
// Create contract test suite
suite := ContractTestSuite{
    Name:     "User Service Contract Tests",
    Provider: "user-service",
    Consumer: "api-gateway",
    Contracts: []Contract{contract},
    Setup: func() error {
        // Setup provider state
        return nil
    },
    Teardown: func() error {
        // Cleanup provider state
        return nil
    },
}

// Run contract tests
runner := NewContractTestRunner(ContractTestConfig{
    Timeout: 30 * time.Second,
    Verify:  true,
})
runner.AddSuite(suite)
runner.Run(t)
```

### Contract Test Helper

```go
// Create contract test helper
helper := NewContractTestHelper(ContractTestConfig{
    Timeout: 30 * time.Second,
    Verbose: true,
})

// Create contract
contract := helper.CreateContract("user-service", "user-service", "api-gateway").
    WithDescription("User service contract").
    AddInteraction(
        helper.CreateInteraction("Get user by ID").
            WithState("user exists").
            WithRequest(
                helper.CreateRequest("GET", "/users/123").
                    WithHeader("Authorization", "Bearer token").
                    Build(),
            ).
            WithResponse(
                helper.CreateResponse(200).
                    WithHeader("Content-Type", "application/json").
                    WithBody(map[string]interface{}{
                        "id":    "123",
                        "name":  "John Doe",
                        "email": "john@example.com",
                    }).
                    Build(),
            ).
            Build(),
    ).
    Build()

// Add to test suite
suite := ContractTestSuite{
    Name:      "User Service Contract Tests",
    Provider:  "user-service",
    Consumer:  "api-gateway",
    Contracts: []Contract{contract},
}

helper.AddSuite(suite)
helper.Run(t)
```

## Test Configuration

### Unit Test Configuration

```go
config := TestConfig{
    Timeout:  30 * time.Second,
    Parallel: true,
    Verbose:  true,
    Tags:     []string{"unit", "fast"},
    Coverage: true,
    Benchmark: false,
    Race:     false,
}
```

### Integration Test Configuration

```go
config := IntegrationTestConfig{
    Timeout:     5 * time.Minute,
    Parallel:    false,
    Verbose:     true,
    Tags:        []string{"integration", "slow"},
    Cleanup:     true,
    Network:     "test-network",
    Registry:    "localhost:5000",
    Environment: "test",
}
```

### Contract Test Configuration

```go
config := ContractTestConfig{
    Timeout:  30 * time.Second,
    Verbose:  true,
    Broker:   "http://localhost:9292",
    Publish:  true,
    Verify:   true,
    Provider: "user-service",
    Consumer: "api-gateway",
}
```

## Test Execution

### Running Tests

```bash
# Run unit tests
go test ./internal/testing/... -v

# Run integration tests
go test ./internal/testing/... -tags=integration -v

# Run contract tests
go test ./internal/testing/... -tags=contract -v

# Run all tests
go test ./internal/testing/... -v

# Run tests with coverage
go test ./internal/testing/... -cover -v

# Run tests with race detection
go test ./internal/testing/... -race -v

# Run benchmarks
go test ./internal/testing/... -bench=. -v
```

### Test Tags

```go
// Tag tests for different execution contexts
func TestUserService(t *testing.T) {
    // Unit test
}

func TestUserServiceIntegration(t *testing.T) {
    // Integration test
}

func TestUserServiceContract(t *testing.T) {
    // Contract test
}
```

### Test Environment

```bash
# Set test environment variables
export TEST_DATABASE_URL="postgres://localhost:5432/test"
export TEST_REDIS_URL="redis://localhost:6379/1"
export TEST_LOG_LEVEL="debug"

# Run tests
go test ./internal/testing/... -v
```

## Best Practices

### Unit Testing

1. **Test Isolation** - Each test should be independent
2. **Mock Dependencies** - Mock external dependencies
3. **Fast Execution** - Unit tests should run quickly
4. **Clear Assertions** - Use descriptive assertion messages
5. **Test Coverage** - Aim for high test coverage

### Integration Testing

1. **Real Dependencies** - Use real databases and services
2. **Test Data** - Use consistent test data
3. **Cleanup** - Always cleanup test data
4. **Parallel Execution** - Run tests in parallel when possible
5. **Health Checks** - Wait for services to be healthy

### Contract Testing

1. **Provider State** - Set up provider state correctly
2. **Request Validation** - Validate all request parameters
3. **Response Validation** - Validate all response fields
4. **Error Cases** - Test error scenarios
5. **Versioning** - Handle API versioning

### General Testing

1. **Test Naming** - Use descriptive test names
2. **Test Organization** - Organize tests logically
3. **Test Documentation** - Document complex tests
4. **Test Maintenance** - Keep tests up to date
5. **Test Performance** - Monitor test execution time

## Test Data Management

### Test Data Setup

```go
// Setup test data
func setupTestData() error {
    // Create test users
    users := []User{
        {ID: "1", Name: "John Doe", Email: "john@example.com"},
        {ID: "2", Name: "Jane Smith", Email: "jane@example.com"},
    }

    for _, user := range users {
        if err := createUser(user); err != nil {
            return err
        }
    }

    return nil
}

// Cleanup test data
func cleanupTestData() error {
    // Delete test users
    if err := deleteAllUsers(); err != nil {
        return err
    }

    return nil
}
```

### Test Data Factories

```go
// User factory
func NewTestUser(overrides ...func(*User)) *User {
    user := &User{
        ID:    generateID(),
        Name:  "Test User",
        Email: "test@example.com",
    }

    for _, override := range overrides {
        override(user)
    }

    return user
}

// Use factory
user := NewTestUser(func(u *User) {
    u.Name = "John Doe"
    u.Email = "john@example.com"
})
```

## Continuous Integration

### Test Pipeline

```yaml
# .github/workflows/test.yml
name: Test Pipeline

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: go test ./internal/testing/... -v -race -cover

  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: go test ./internal/testing/... -tags=integration -v

  contract-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: go test ./internal/testing/... -tags=contract -v
```

## Troubleshooting

### Common Issues

#### 1. Test Timeouts

```go
// Increase test timeout
config := TestConfig{
    Timeout: 60 * time.Second,
}
```

#### 2. Mock Verification Failures

```go
// Check mock expectations
if err := mock.Verify(); err != nil {
    t.Errorf("Mock verification failed: %v", err)
}

// Check mock calls
calls := mock.GetCalls("method")
if len(calls) != expectedCalls {
    t.Errorf("Expected %d calls, got %d", expectedCalls, len(calls))
}
```

#### 3. Integration Test Failures

```go
// Wait for services to be ready
if err := client.WaitForReady(60 * time.Second); err != nil {
    t.Errorf("Service not ready: %v", err)
}

// Check service health
if err := client.HealthCheck(); err != nil {
    t.Errorf("Service not healthy: %v", err)
}
```

#### 4. Contract Test Failures

```go
// Check provider state
if err := setupState("user exists"); err != nil {
    t.Errorf("Failed to setup state: %v", err)
}

// Verify request/response
if err := verifyResponse(resp, expected); err != nil {
    t.Errorf("Response verification failed: %v", err)
}
```

## Conclusion

The microservice testing system provides comprehensive testing capabilities including:

- **Unit Testing** - Fast, isolated component testing
- **Integration Testing** - Service interaction testing
- **Contract Testing** - API compatibility testing
- **Test Utilities** - Mock services, helpers, and assertions
- **Test Configuration** - Flexible test configuration
- **CI/CD Integration** - Automated test execution

Key benefits:

- **Comprehensive Coverage** - Tests all aspects of microservices
- **Fast Feedback** - Quick test execution and results
- **Reliable Tests** - Stable and maintainable test suite
- **Easy Maintenance** - Well-organized and documented tests
- **CI/CD Ready** - Automated test execution in pipelines

This testing system provides a solid foundation for ensuring quality and reliability in microservices architectures, following testing best practices and industry standards.
