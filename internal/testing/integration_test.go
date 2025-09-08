package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"testing"
	"time"
)

// IntegrationTestSuite represents an integration test suite
type IntegrationTestSuite struct {
	Name     string
	Services []ServiceConfig
	Setup    func() error
	Teardown func() error
	Tests    []IntegrationTestCase
	Timeout  time.Duration
	Parallel bool
}

// ServiceConfig represents a service configuration for testing
type ServiceConfig struct {
	Name        string
	Image       string
	Port        int
	Environment map[string]string
	Volumes     map[string]string
	DependsOn   []string
	HealthCheck HealthCheck
}

// HealthCheck represents a health check configuration
type HealthCheck struct {
	Path     string
	Interval time.Duration
	Timeout  time.Duration
	Retries  int
}

// IntegrationTestCase represents an integration test case
type IntegrationTestCase struct {
	Name        string
	Description string
	Setup       func() error
	Teardown    func() error
	Test        func(t *testing.T, services map[string]*Service) error
	Timeout     time.Duration
	Skip        bool
	Tags        []string
}

// Service represents a running service in tests
type Service struct {
	Config ServiceConfig
	URL    string
	Client *http.Client
	Health bool
}

// IntegrationTestRunner manages integration test execution
type IntegrationTestRunner struct {
	suites []IntegrationTestSuite
	config IntegrationTestConfig
}

// IntegrationTestConfig represents integration test configuration
type IntegrationTestConfig struct {
	Timeout     time.Duration
	Parallel    bool
	Verbose     bool
	Tags        []string
	Cleanup     bool
	Network     string
	Registry    string
	Environment string
}

// NewIntegrationTestRunner creates a new integration test runner
func NewIntegrationTestRunner(config IntegrationTestConfig) *IntegrationTestRunner {
	return &IntegrationTestRunner{
		config: config,
	}
}

// AddSuite adds an integration test suite to the runner
func (itr *IntegrationTestRunner) AddSuite(suite IntegrationTestSuite) {
	itr.suites = append(itr.suites, suite)
}

// Run executes all integration test suites
func (itr *IntegrationTestRunner) Run(t *testing.T) {
	for _, suite := range itr.suites {
		itr.runSuite(t, suite)
	}
}

// runSuite executes a single integration test suite
func (itr *IntegrationTestRunner) runSuite(t *testing.T, suite IntegrationTestSuite) {
	t.Run(suite.Name, func(t *testing.T) {
		// Setup suite
		if suite.Setup != nil {
			if err := suite.Setup(); err != nil {
				t.Fatalf("Suite setup failed: %v", err)
			}
		}

		// Teardown suite
		defer func() {
			if suite.Teardown != nil {
				if err := suite.Teardown(); err != nil {
					t.Errorf("Suite teardown failed: %v", err)
				}
			}
		}()

		// Start services
		services, err := itr.startServices(suite.Services)
		if err != nil {
			t.Fatalf("Failed to start services: %v", err)
		}

		// Stop services
		defer func() {
			if err := itr.stopServices(services); err != nil {
				t.Errorf("Failed to stop services: %v", err)
			}
		}()

		// Wait for services to be healthy
		if err := itr.waitForHealthyServices(services); err != nil {
			t.Fatalf("Services not healthy: %v", err)
		}

		// Run tests
		for _, testCase := range suite.Tests {
			itr.runIntegrationTestCase(t, testCase, services)
		}
	})
}

// startServices starts all services in the suite
func (itr *IntegrationTestRunner) startServices(configs []ServiceConfig) (map[string]*Service, error) {
	services := make(map[string]*Service)

	for _, config := range configs {
		service, err := itr.startService(config)
		if err != nil {
			// Cleanup started services
			if stopErr := itr.stopServices(services); stopErr != nil {
				log.Printf("Warning: failed to stop services during cleanup: %v", stopErr)
			}
			return nil, fmt.Errorf("failed to start service %s: %w", config.Name, err)
		}
		services[config.Name] = service
	}

	return services, nil
}

// startService starts a single service
func (itr *IntegrationTestRunner) startService(config ServiceConfig) (*Service, error) {
	// In a real implementation, this would:
	// 1. Pull the Docker image
	// 2. Start the container
	// 3. Wait for the service to be ready
	// 4. Return the service instance

	service := &Service{
		Config: config,
		URL:    fmt.Sprintf("http://localhost:%d", config.Port),
		Client: &http.Client{Timeout: 30 * time.Second},
		Health: false,
	}

	// Simulate service startup
	time.Sleep(100 * time.Millisecond)
	service.Health = true

	return service, nil
}

// stopServices stops all services
func (itr *IntegrationTestRunner) stopServices(services map[string]*Service) error {
	for name, service := range services {
		if err := itr.stopService(service); err != nil {
			return fmt.Errorf("failed to stop service %s: %w", name, err)
		}
	}
	return nil
}

// stopService stops a single service
func (itr *IntegrationTestRunner) stopService(service *Service) error {
	// In a real implementation, this would stop the Docker container
	service.Health = false
	return nil
}

// waitForHealthyServices waits for all services to be healthy
func (itr *IntegrationTestRunner) waitForHealthyServices(services map[string]*Service) error {
	timeout := 30 * time.Second
	interval := 1 * time.Second

	start := time.Now()
	for time.Since(start) < timeout {
		allHealthy := true
		for _, service := range services {
			if !service.Health {
				allHealthy = false
				break
			}

			// Check health endpoint
			if service.Config.HealthCheck.Path != "" {
				resp, err := service.Client.Get(service.URL + service.Config.HealthCheck.Path)
				if err != nil || resp.StatusCode != http.StatusOK {
					allHealthy = false
					break
				}
				resp.Body.Close()
			}
		}

		if allHealthy {
			return nil
		}

		time.Sleep(interval)
	}

	return fmt.Errorf("services not healthy after %v", timeout)
}

// runIntegrationTestCase executes a single integration test case
func (itr *IntegrationTestRunner) runIntegrationTestCase(t *testing.T, testCase IntegrationTestCase, services map[string]*Service) {
	if testCase.Skip {
		t.Skipf("Skipping test: %s", testCase.Name)
		return
	}

	// Check tags
	if !itr.shouldRunTest(testCase.Tags) {
		t.Skipf("Skipping test due to tags: %s", testCase.Name)
		return
	}

	t.Run(testCase.Name, func(t *testing.T) {
		// Setup test
		if testCase.Setup != nil {
			if err := testCase.Setup(); err != nil {
				t.Fatalf("Test setup failed: %v", err)
			}
		}

		// Teardown test
		defer func() {
			if testCase.Teardown != nil {
				if err := testCase.Teardown(); err != nil {
					t.Errorf("Test teardown failed: %v", err)
				}
			}
		}()

		// Run test with timeout
		timeout := testCase.Timeout
		if timeout == 0 {
			timeout = itr.config.Timeout
		}
		if timeout == 0 {
			timeout = 5 * time.Minute
		}

		done := make(chan error, 1)
		go func() {
			done <- testCase.Test(t, services)
		}()

		select {
		case err := <-done:
			if err != nil {
				t.Errorf("Test failed: %v", err)
			}
		case <-time.After(timeout):
			t.Errorf("Test timed out after %v", timeout)
		}
	})
}

// shouldRunTest checks if a test should run based on tags
func (itr *IntegrationTestRunner) shouldRunTest(testTags []string) bool {
	if len(itr.config.Tags) == 0 {
		return true
	}

	for _, configTag := range itr.config.Tags {
		for _, testTag := range testTags {
			if configTag == testTag {
				return true
			}
		}
	}
	return false
}

// ServiceClient provides utilities for testing services
type ServiceClient struct {
	service *Service
}

// NewServiceClient creates a new service client
func NewServiceClient(service *Service) *ServiceClient {
	return &ServiceClient{
		service: service,
	}
}

// Get makes a GET request to the service
func (sc *ServiceClient) Get(path string) (*http.Response, error) {
	return sc.service.Client.Get(sc.service.URL + path)
}

// Post makes a POST request to the service
func (sc *ServiceClient) Post(path string, body []byte) (*http.Response, error) {
	return sc.service.Client.Post(sc.service.URL+path, "application/json", bytes.NewReader(body))
}

// Put makes a PUT request to the service
func (sc *ServiceClient) Put(path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest("PUT", sc.service.URL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return sc.service.Client.Do(req)
}

// Delete makes a DELETE request to the service
func (sc *ServiceClient) Delete(path string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", sc.service.URL+path, nil)
	if err != nil {
		return nil, err
	}
	return sc.service.Client.Do(req)
}

// HealthCheck checks if the service is healthy
func (sc *ServiceClient) HealthCheck() error {
	if !sc.service.Health {
		return fmt.Errorf("service %s is not healthy", sc.service.Config.Name)
	}

	if sc.service.Config.HealthCheck.Path != "" {
		resp, err := sc.Get(sc.service.Config.HealthCheck.Path)
		if err != nil {
			return fmt.Errorf("health check failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("health check returned status %d", resp.StatusCode)
		}
	}

	return nil
}

// WaitForReady waits for the service to be ready
func (sc *ServiceClient) WaitForReady(timeout time.Duration) error {
	start := time.Now()
	interval := 1 * time.Second

	for time.Since(start) < timeout {
		if err := sc.HealthCheck(); err == nil {
			return nil
		}
		time.Sleep(interval)
	}

	return fmt.Errorf("service not ready after %v", timeout)
}

// IntegrationTestHelper provides common integration test utilities
type IntegrationTestHelper struct {
	Services map[string]*ServiceClient
}

// NewIntegrationTestHelper creates a new integration test helper
func NewIntegrationTestHelper(services map[string]*Service) *IntegrationTestHelper {
	clients := make(map[string]*ServiceClient)
	for name, service := range services {
		clients[name] = NewServiceClient(service)
	}

	return &IntegrationTestHelper{
		Services: clients,
	}
}

// GetService returns a service client by name
func (ith *IntegrationTestHelper) GetService(name string) *ServiceClient {
	return ith.Services[name]
}

// WaitForAllServices waits for all services to be ready
func (ith *IntegrationTestHelper) WaitForAllServices(timeout time.Duration) error {
	for name, client := range ith.Services {
		if err := client.WaitForReady(timeout); err != nil {
			return fmt.Errorf("service %s not ready: %w", name, err)
		}
	}
	return nil
}

// HealthCheckAllServices checks health of all services
func (ith *IntegrationTestHelper) HealthCheckAllServices() error {
	for name, client := range ith.Services {
		if err := client.HealthCheck(); err != nil {
			return fmt.Errorf("service %s health check failed: %w", name, err)
		}
	}
	return nil
}

// AssertServiceResponse asserts a service response
func AssertServiceResponse(t *testing.T, resp *http.Response, expectedStatus int, expectedBody string) {
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	if expectedBody != "" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read response body: %v", err)
			return
		}

		if string(body) != expectedBody {
			t.Errorf("Expected body %s, got %s", expectedBody, string(body))
		}
	}
}

// AssertServiceJSONResponse asserts a service JSON response
func AssertServiceJSONResponse(t *testing.T, resp *http.Response, expectedStatus int, expectedData interface{}) {
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	if expectedData != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read response body: %v", err)
			return
		}

		var actualData interface{}
		if err := json.Unmarshal(body, &actualData); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
			return
		}

		if !reflect.DeepEqual(actualData, expectedData) {
			t.Errorf("Expected data %v, got %v", expectedData, actualData)
		}
	}
}
