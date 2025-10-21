// Package services provides HTTP endpoint health check business logic.
package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// HealthCheckService handles HTTP endpoint health checks.
type HealthCheckService struct {
	httpClient *http.Client
	logger     *zap.Logger
}

// NewHealthCheckService creates a new health check service.
func NewHealthCheckService(logger *zap.Logger) *HealthCheckService {
	return &HealthCheckService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// HealthCheckRequest represents a health check request.
type HealthCheckRequest struct {
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers"`
	Body           string            `json:"body"`
	ExpectedStatus int               `json:"expected_status"`
	ExpectedBody   string            `json:"expected_body"`
	Timeout        int               `json:"timeout"` // in seconds
}

// HealthCheckResponse represents a health check response.
type HealthCheckResponse struct {
	Status       string            `json:"status"` // success, failure, timeout
	StatusCode   int               `json:"status_code"`
	ResponseTime float64           `json:"response_time"` // in milliseconds
	ResponseBody string            `json:"response_body"`
	ErrorMessage string            `json:"error_message"`
	Headers      map[string]string `json:"headers"`
	Timestamp    time.Time         `json:"timestamp"`
}

// PerformHealthCheck performs an HTTP health check.
func (s *HealthCheckService) PerformHealthCheck(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error) {
	startTime := time.Now()

	// Set default values
	if req.Method == "" {
		req.Method = "GET"
	}
	if req.ExpectedStatus == 0 {
		req.ExpectedStatus = 200
	}
	if req.Timeout == 0 {
		req.Timeout = 30
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, strings.NewReader(req.Body))
	if err != nil {
		return &HealthCheckResponse{
			Status:       "failure",
			ErrorMessage: fmt.Sprintf("Failed to create request: %v", err),
			Timestamp:    time.Now(),
		}, nil
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(req.Timeout) * time.Second,
	}

	// Perform request
	resp, err := client.Do(httpReq)
	responseTime := float64(time.Since(startTime).Nanoseconds()) / 1e6 // Convert to milliseconds

	if err != nil {
		return &HealthCheckResponse{
			Status:       "timeout",
			ErrorMessage: fmt.Sprintf("Request failed: %v", err),
			ResponseTime: responseTime,
			Timestamp:    time.Now(),
		}, nil
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Warn("Failed to read response body", zap.Error(err))
	}

	responseBody := string(bodyBytes)

	// Check status code
	status := "success"
	errorMessage := ""

	if resp.StatusCode != req.ExpectedStatus {
		status = "failure"
		errorMessage = fmt.Sprintf("Expected status %d, got %d", req.ExpectedStatus, resp.StatusCode)
	}

	// Check response body if expected
	if req.ExpectedBody != "" && !strings.Contains(responseBody, req.ExpectedBody) {
		status = "failure"
		if errorMessage != "" {
			errorMessage += "; "
		}
		errorMessage += fmt.Sprintf("Expected body to contain: %s", req.ExpectedBody)
	}

	// Extract response headers
	responseHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			responseHeaders[key] = values[0]
		}
	}

	return &HealthCheckResponse{
		Status:       status,
		StatusCode:   resp.StatusCode,
		ResponseTime: responseTime,
		ResponseBody: responseBody,
		ErrorMessage: errorMessage,
		Headers:      responseHeaders,
		Timestamp:    time.Now(),
	}, nil
}

// PerformBulkHealthCheck performs multiple health checks concurrently.
func (s *HealthCheckService) PerformBulkHealthCheck(ctx context.Context, requests []*HealthCheckRequest) ([]*HealthCheckResponse, error) {
	results := make([]*HealthCheckResponse, len(requests))

	// Use a channel to collect results
	type result struct {
		index  int
		result *HealthCheckResponse
		err    error
	}

	resultChan := make(chan result, len(requests))

	// Start goroutines for each health check
	for i, req := range requests {
		go func(index int, request *HealthCheckRequest) {
			resp, err := s.PerformHealthCheck(ctx, request)
			resultChan <- result{index: index, result: resp, err: err}
		}(i, req)
	}

	// Collect results
	for i := 0; i < len(requests); i++ {
		res := <-resultChan
		if res.err != nil {
			results[res.index] = &HealthCheckResponse{
				Status:       "failure",
				ErrorMessage: fmt.Sprintf("Health check failed: %v", res.err),
				Timestamp:    time.Now(),
			}
		} else {
			results[res.index] = res.result
		}
	}

	return results, nil
}

// ValidateHealthCheckRequest validates a health check request.
func (s *HealthCheckService) ValidateHealthCheckRequest(req *HealthCheckRequest) error {
	if req.URL == "" {
		return fmt.Errorf("URL is required")
	}

	// Validate HTTP method
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH"}
	methodValid := false
	for _, method := range validMethods {
		if strings.ToUpper(req.Method) == method {
			methodValid = true
			break
		}
	}
	if !methodValid {
		return fmt.Errorf("invalid HTTP method: %s", req.Method)
	}

	if req.Timeout < 1 || req.Timeout > 300 {
		return fmt.Errorf("timeout must be between 1 and 300 seconds")
	}

	if req.ExpectedStatus < 100 || req.ExpectedStatus > 599 {
		return fmt.Errorf("expected status must be between 100 and 599")
	}

	return nil
}

// GetHealthCheckSummary returns a summary of health check results.
func (s *HealthCheckService) GetHealthCheckSummary(responses []*HealthCheckResponse) map[string]interface{} {
	total := len(responses)
	successful := 0
	failed := 0
	timeouts := 0
	var totalResponseTime float64
	var maxResponseTime float64
	var minResponseTime float64 = 999999

	for _, resp := range responses {
		switch resp.Status {
		case "success":
			successful++
		case "failure":
			failed++
		case "timeout":
			timeouts++
		}

		if resp.ResponseTime > 0 {
			totalResponseTime += resp.ResponseTime
			if resp.ResponseTime > maxResponseTime {
				maxResponseTime = resp.ResponseTime
			}
			if resp.ResponseTime < minResponseTime {
				minResponseTime = resp.ResponseTime
			}
		}
	}

	avgResponseTime := float64(0)
	if successful > 0 {
		avgResponseTime = totalResponseTime / float64(successful)
	}

	successRate := float64(0)
	if total > 0 {
		successRate = float64(successful) / float64(total) * 100
	}

	return map[string]interface{}{
		"total_checks":          total,
		"successful_checks":     successful,
		"failed_checks":         failed,
		"timeout_checks":        timeouts,
		"success_rate":          successRate,
		"average_response_time": avgResponseTime,
		"max_response_time":     maxResponseTime,
		"min_response_time":     minResponseTime,
		"timestamp":             time.Now(),
	}
}

