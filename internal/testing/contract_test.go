package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"
)

// ContractTestSuite represents a contract test suite
type ContractTestSuite struct {
	Name      string
	Provider  string
	Consumer  string
	Contracts []Contract
	Setup     func() error
	Teardown  func() error
	Timeout   time.Duration
}

// Contract represents a service contract
type Contract struct {
	Name         string
	Description  string
	Provider     string
	Consumer     string
	Interactions []Interaction
	Metadata     map[string]interface{}
}

// Interaction represents a service interaction
type Interaction struct {
	Description string
	Request     Request
	Response    Response
	State       string
}

// Request represents a contract request
type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    interface{}
	Query   map[string]string
}

// Response represents a contract response
type Response struct {
	Status  int
	Headers map[string]string
	Body    interface{}
}

// ContractTestRunner manages contract test execution
type ContractTestRunner struct {
	suites []ContractTestSuite
	config ContractTestConfig
}

// ContractTestConfig represents contract test configuration
type ContractTestConfig struct {
	Timeout  time.Duration
	Verbose  bool
	Broker   string
	Publish  bool
	Verify   bool
	Provider string
	Consumer string
}

// NewContractTestRunner creates a new contract test runner
func NewContractTestRunner(config ContractTestConfig) *ContractTestRunner {
	return &ContractTestRunner{
		config: config,
	}
}

// AddSuite adds a contract test suite to the runner
func (ctr *ContractTestRunner) AddSuite(suite ContractTestSuite) {
	ctr.suites = append(ctr.suites, suite)
}

// Run executes all contract test suites
func (ctr *ContractTestRunner) Run(t *testing.T) {
	for _, suite := range ctr.suites {
		ctr.runSuite(t, suite)
	}
}

// runSuite executes a single contract test suite
func (ctr *ContractTestRunner) runSuite(t *testing.T, suite ContractTestSuite) {
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

		// Run contract tests
		for _, contract := range suite.Contracts {
			ctr.runContract(t, contract)
		}
	})
}

// runContract executes a single contract test
func (ctr *ContractTestRunner) runContract(t *testing.T, contract Contract) {
	t.Run(contract.Name, func(t *testing.T) {
		for _, interaction := range contract.Interactions {
			ctr.runInteraction(t, interaction)
		}
	})
}

// runInteraction executes a single interaction test
func (ctr *ContractTestRunner) runInteraction(t *testing.T, interaction Interaction) {
	t.Run(interaction.Description, func(t *testing.T) {
		// Setup state if specified
		if interaction.State != "" {
			if err := ctr.setupState(interaction.State); err != nil {
				t.Fatalf("Failed to setup state %s: %v", interaction.State, err)
			}
		}

		// Make request
		resp, err := ctr.makeRequest(interaction.Request)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		// Verify response
		if err := ctr.verifyResponse(resp, interaction.Response); err != nil {
			t.Errorf("Response verification failed: %v", err)
		}
	})
}

// setupState sets up the provider state
func (ctr *ContractTestRunner) setupState(state string) error {
	// In a real implementation, this would:
	// 1. Call the provider's state setup endpoint
	// 2. Set up the required data state
	// 3. Verify the state was set correctly

	// For now, just simulate state setup
	time.Sleep(10 * time.Millisecond)
	return nil
}

// makeRequest makes a request to the provider
func (ctr *ContractTestRunner) makeRequest(req Request) (*http.Response, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	// Build URL
	url := fmt.Sprintf("http://localhost:8080%s", req.Path)
	if len(req.Query) > 0 {
		url += "?"
		first := true
		for key, value := range req.Query {
			if !first {
				url += "&"
			}
			url += fmt.Sprintf("%s=%s", key, value)
			first = false
		}
	}

	// Create request
	var httpReq *http.Request
	var err error

	if req.Body != nil {
		body, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		httpReq, err = http.NewRequest(req.Method, url, bytes.NewReader(body))
	} else {
		httpReq, err = http.NewRequest(req.Method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Make request
	return client.Do(httpReq)
}

// verifyResponse verifies the response matches the contract
func (ctr *ContractTestRunner) verifyResponse(resp *http.Response, expected Response) error {
	// Verify status code
	if resp.StatusCode != expected.Status {
		return fmt.Errorf("expected status %d, got %d", expected.Status, resp.StatusCode)
	}

	// Verify headers
	for key, expectedValue := range expected.Headers {
		actualValue := resp.Header.Get(key)
		if actualValue != expectedValue {
			return fmt.Errorf("expected header %s: %s, got %s", key, expectedValue, actualValue)
		}
	}

	// Verify body
	if expected.Body != nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		var actualBody interface{}
		if err := json.Unmarshal(body, &actualBody); err != nil {
			return fmt.Errorf("failed to unmarshal response body: %w", err)
		}

		if !reflect.DeepEqual(actualBody, expected.Body) {
			return fmt.Errorf("expected body %v, got %v", expected.Body, actualBody)
		}
	}

	return nil
}

// ContractBuilder helps build contracts
type ContractBuilder struct {
	contract Contract
}

// NewContractBuilder creates a new contract builder
func NewContractBuilder(name, provider, consumer string) *ContractBuilder {
	return &ContractBuilder{
		contract: Contract{
			Name:         name,
			Provider:     provider,
			Consumer:     consumer,
			Interactions: make([]Interaction, 0),
			Metadata:     make(map[string]interface{}),
		},
	}
}

// WithDescription sets the contract description
func (cb *ContractBuilder) WithDescription(description string) *ContractBuilder {
	cb.contract.Description = description
	return cb
}

// WithMetadata sets contract metadata
func (cb *ContractBuilder) WithMetadata(key string, value interface{}) *ContractBuilder {
	cb.contract.Metadata[key] = value
	return cb
}

// AddInteraction adds an interaction to the contract
func (cb *ContractBuilder) AddInteraction(interaction Interaction) *ContractBuilder {
	cb.contract.Interactions = append(cb.contract.Interactions, interaction)
	return cb
}

// Build builds the contract
func (cb *ContractBuilder) Build() Contract {
	return cb.contract
}

// InteractionBuilder helps build interactions
type InteractionBuilder struct {
	interaction Interaction
}

// NewInteractionBuilder creates a new interaction builder
func NewInteractionBuilder(description string) *InteractionBuilder {
	return &InteractionBuilder{
		interaction: Interaction{
			Description: description,
			Request:     Request{},
			Response:    Response{},
		},
	}
}

// WithState sets the provider state
func (ib *InteractionBuilder) WithState(state string) *InteractionBuilder {
	ib.interaction.State = state
	return ib
}

// WithRequest sets the request
func (ib *InteractionBuilder) WithRequest(request Request) *InteractionBuilder {
	ib.interaction.Request = request
	return ib
}

// WithResponse sets the response
func (ib *InteractionBuilder) WithResponse(response Response) *InteractionBuilder {
	ib.interaction.Response = response
	return ib
}

// Build builds the interaction
func (ib *InteractionBuilder) Build() Interaction {
	return ib.interaction
}

// RequestBuilder helps build requests
type RequestBuilder struct {
	request Request
}

// NewRequestBuilder creates a new request builder
func NewRequestBuilder(method, path string) *RequestBuilder {
	return &RequestBuilder{
		request: Request{
			Method:  method,
			Path:    path,
			Headers: make(map[string]string),
			Query:   make(map[string]string),
		},
	}
}

// WithHeader adds a header to the request
func (rb *RequestBuilder) WithHeader(key, value string) *RequestBuilder {
	rb.request.Headers[key] = value
	return rb
}

// WithBody sets the request body
func (rb *RequestBuilder) WithBody(body interface{}) *RequestBuilder {
	rb.request.Body = body
	return rb
}

// WithQuery adds a query parameter
func (rb *RequestBuilder) WithQuery(key, value string) *RequestBuilder {
	rb.request.Query[key] = value
	return rb
}

// Build builds the request
func (rb *RequestBuilder) Build() Request {
	return rb.request
}

// ResponseBuilder helps build responses
type ResponseBuilder struct {
	response Response
}

// NewResponseBuilder creates a new response builder
func NewResponseBuilder(status int) *ResponseBuilder {
	return &ResponseBuilder{
		response: Response{
			Status:  status,
			Headers: make(map[string]string),
		},
	}
}

// WithHeader adds a header to the response
func (rb *ResponseBuilder) WithHeader(key, value string) *ResponseBuilder {
	rb.response.Headers[key] = value
	return rb
}

// WithBody sets the response body
func (rb *ResponseBuilder) WithBody(body interface{}) *ResponseBuilder {
	rb.response.Body = body
	return rb
}

// Build builds the response
func (rb *ResponseBuilder) Build() Response {
	return rb.response
}

// ContractTestHelper provides utilities for contract testing
type ContractTestHelper struct {
	runner *ContractTestRunner
}

// NewContractTestHelper creates a new contract test helper
func NewContractTestHelper(config ContractTestConfig) *ContractTestHelper {
	return &ContractTestHelper{
		runner: NewContractTestRunner(config),
	}
}

// CreateContract creates a new contract
func (cth *ContractTestHelper) CreateContract(name, provider, consumer string) *ContractBuilder {
	return NewContractBuilder(name, provider, consumer)
}

// CreateInteraction creates a new interaction
func (cth *ContractTestHelper) CreateInteraction(description string) *InteractionBuilder {
	return NewInteractionBuilder(description)
}

// CreateRequest creates a new request
func (cth *ContractTestHelper) CreateRequest(method, path string) *RequestBuilder {
	return NewRequestBuilder(method, path)
}

// CreateResponse creates a new response
func (cth *ContractTestHelper) CreateResponse(status int) *ResponseBuilder {
	return NewResponseBuilder(status)
}

// AddSuite adds a test suite to the helper
func (cth *ContractTestHelper) AddSuite(suite ContractTestSuite) {
	cth.runner.AddSuite(suite)
}

// Run runs all test suites
func (cth *ContractTestHelper) Run(t *testing.T) {
	cth.runner.Run(t)
}

// Example contract test
func TestExampleContract(t *testing.T) {
	helper := NewContractTestHelper(ContractTestConfig{
		Timeout: 30 * time.Second,
		Verbose: true,
	})

	// Create a contract
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

	// Create test suite
	suite := ContractTestSuite{
		Name:      "User Service Contract Tests",
		Provider:  "user-service",
		Consumer:  "api-gateway",
		Contracts: []Contract{contract},
	}

	helper.AddSuite(suite)
	helper.Run(t)
}
