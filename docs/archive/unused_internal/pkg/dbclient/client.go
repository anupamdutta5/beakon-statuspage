// Package dbclient provides a client for interacting with the database-service
package dbclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// DatabaseType represents the type of database
const (
	DatabaseTypePostgreSQL = "postgresql"
	DatabaseTypeMySQL      = "mysql"
)

// Database represents a database in the database-service
type Database struct {
	ID          string    `json:"id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Type        string    `json:"type"` // postgresql, mysql, etc.
	Host        string    `json:"host"`
	Port        int       `json:"port"`
	Username    string    `json:"username"`
	Password    string    `json:"password,omitempty"`
	SSLMode     string    `json:"ssl_mode,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

// QueryRequest represents a database query request
type QueryRequest struct {
	DatabaseID string        `json:"database_id"`
	Query      string        `json:"query"`
	Parameters []interface{} `json:"parameters,omitempty"`
}

// QueryResult represents the result of a database query
type QueryResult struct {
	Columns     []string        `json:"columns"`
	Rows        [][]interface{} `json:"rows"`
	RowCount    int64           `json:"row_count"`
	TimeElapsed string          `json:"time_elapsed"`
}

// Migration represents a database migration
type Migration struct {
	ID          string    `json:"id,omitempty"`
	DatabaseID  string    `json:"database_id"`
	Version     int64     `json:"version"`
	Name        string    `json:"name"`
	SQL         string    `json:"sql"`
	AppliedAt   time.Time `json:"applied_at,omitempty"`
	Status      string    `json:"status,omitempty"`
	Error       string    `json:"error,omitempty"`
	TimeElapsed string    `json:"time_elapsed,omitempty"`
}

// Client is a client for interacting with the database-service
type Client struct {
	baseURL    string
	httpClient *http.Client
	authToken  string
}

// NewClient creates a new database client
func NewClient(baseURL, authToken string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		authToken:  authToken,
	}
}

// CreateDatabase creates a new database
func (c *Client) CreateDatabase(ctx context.Context, db *Database) (*Database, error) {
	url := fmt.Sprintf("%s/api/v1/databases", c.baseURL)
	return c.sendDatabaseRequest(ctx, http.MethodPost, url, db)
}

// GetDatabase retrieves a database by ID
func (c *Client) GetDatabase(ctx context.Context, id string) (*Database, error) {
	url := fmt.Sprintf("%s/api/v1/databases/%s", c.baseURL, id)
	return c.sendDatabaseRequest(ctx, http.MethodGet, url, nil)
}

// ListDatabases lists all databases
func (c *Client) ListDatabases(ctx context.Context) ([]Database, error) {
	url := fmt.Sprintf("%s/api/v1/databases", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var dbs []Database
	if err := json.NewDecoder(resp.Body).Decode(&dbs); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return dbs, nil
}

// UpdateDatabase updates an existing database
func (c *Client) UpdateDatabase(ctx context.Context, id string, db *Database) (*Database, error) {
	url := fmt.Sprintf("%s/api/v1/databases/%s", c.baseURL, id)
	return c.sendDatabaseRequest(ctx, http.MethodPut, url, db)
}

// DeleteDatabase deletes a database
func (c *Client) DeleteDatabase(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/api/v1/databases/%s", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return c.parseError(resp)
	}

	return nil
}

// ExecuteQuery executes a query on a database
func (c *Client) ExecuteQuery(ctx context.Context, req *QueryRequest) (*QueryResult, error) {
	url := fmt.Sprintf("%s/api/v1/query", c.baseURL)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return nil, fmt.Errorf("error encoding request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result QueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &result, nil
}

// ApplyMigration applies a database migration
func (c *Client) ApplyMigration(ctx context.Context, migration *Migration) (*Migration, error) {
	url := fmt.Sprintf("%s/api/v1/migrations/apply", c.baseURL)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(migration); err != nil {
		return nil, fmt.Errorf("error encoding request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result Migration
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &result, nil
}

// GetMigrationStatus gets the status of a migration
func (c *Client) GetMigrationStatus(ctx context.Context, migrationID string) (*Migration, error) {
	url := fmt.Sprintf("%s/api/v1/migrations/%s/status", c.baseURL, migrationID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var migration Migration
	if err := json.NewDecoder(resp.Body).Decode(&migration); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &migration, nil
}

// HealthCheck checks the health of the database-service
func (c *Client) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %s", resp.Status)
	}

	return nil
}

// sendDatabaseRequest sends a request to the database-service and decodes the response
func (c *Client) sendDatabaseRequest(ctx context.Context, method, url string, body interface{}) (*Database, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, fmt.Errorf("error encoding request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var db Database
	if err := json.NewDecoder(resp.Body).Decode(&db); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &db, nil
}

// setHeaders sets the common headers for all requests
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}
}

// parseError parses an error response from the API
type apiError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (c *Client) parseError(resp *http.Response) error {
	var apiErr apiError
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading error response: %w", err)
	}

	if err := json.Unmarshal(body, &apiErr); err != nil {
		return fmt.Errorf("error decoding error response: %s: %w", string(body), err)
	}

	if apiErr.Message != "" {
		return fmt.Errorf("database service error: %s (status: %d)", apiErr.Message, resp.StatusCode)
	}

	return fmt.Errorf("database service error: %s (status: %d)", apiErr.Error, resp.StatusCode)
}

// GenerateDatabaseName generates a unique database name for a service
func GenerateDatabaseName(serviceName string) string {
	return fmt.Sprintf("%s_%s", serviceName, uuid.New().String()[:8])
}

// GenerateDatabaseUser generates a unique database username for a service
func GenerateDatabaseUser(serviceName string) string {
	return fmt.Sprintf("%s_%s", serviceName, uuid.New().String()[:8])
}

// GenerateDatabasePassword generates a secure random password
func GenerateDatabasePassword() string {
	return uuid.New().String() + "!@#$"
}
