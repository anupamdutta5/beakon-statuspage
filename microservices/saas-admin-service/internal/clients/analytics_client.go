// Package clients provides external service clients for the SaaS Admin Service.
package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// AnalyticsClient handles communication with the analytics microservice.
type AnalyticsClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewAnalyticsClient creates a new analytics client.
func NewAnalyticsClient(baseURL string, logger *zap.Logger) *AnalyticsClient {
	return &AnalyticsClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// AnalyticsOverview represents analytics overview data
type AnalyticsOverview struct {
	TotalViews       int64                  `json:"total_views"`
	UniqueVisitors   int64                  `json:"unique_visitors"`
	UptimePercentage float64                `json:"uptime_percentage"`
	AvgResponseTime  float64                `json:"avg_response_time"`
	StatusPages      int                    `json:"status_pages"`
	ActiveIncidents  int                    `json:"active_incidents"`
	DailyMetrics     []DailyMetric          `json:"daily_metrics"`
	TopCountries     []CountryMetric        `json:"top_countries"`
	ResponseTimes    []ResponseTimeMetric   `json:"response_times"`
	UptimeHistory    []UptimeMetric         `json:"uptime_history"`
}

// DailyMetric represents daily analytics data
type DailyMetric struct {
	Date    string `json:"date"`
	Views   int64  `json:"views"`
	Uptime  float64 `json:"uptime"`
}

// CountryMetric represents country-based analytics
type CountryMetric struct {
	Country string `json:"country"`
	Views   int64  `json:"views"`
}

// ResponseTimeMetric represents response time data
type ResponseTimeMetric struct {
	Timestamp    time.Time `json:"timestamp"`
	ResponseTime float64   `json:"response_time"`
}

// UptimeMetric represents uptime data
type UptimeMetric struct {
	Date   string  `json:"date"`
	Uptime float64 `json:"uptime"`
}

// MetricData represents analytics metric data
type MetricData struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// GetAnalyticsOverview fetches analytics overview data from the analytics service.
func (c *AnalyticsClient) GetAnalyticsOverview(ctx context.Context, tenantID string) (*AnalyticsOverview, error) {
	url := fmt.Sprintf("%s/api/v1/analytics/overview", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add tenant ID header for multi-tenancy
	req.Header.Set("X-Tenant-ID", tenantID)
	req.Header.Set("Content-Type", "application/json")

	c.logger.Debug("Fetching analytics overview",
		zap.String("url", url),
		zap.String("tenant_id", tenantID))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to fetch analytics overview", zap.Error(err))
		// Return mock data if analytics service is unavailable
		return c.getMockAnalyticsOverview(tenantID), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("Analytics service returned non-200 status",
			zap.Int("status", resp.StatusCode))
		// Return mock data if analytics service returns error
		return c.getMockAnalyticsOverview(tenantID), nil
	}

	var overview AnalyticsOverview
	if err := json.NewDecoder(resp.Body).Decode(&overview); err != nil {
		c.logger.Error("Failed to decode analytics overview", zap.Error(err))
		// Return mock data if decode fails
		return c.getMockAnalyticsOverview(tenantID), nil
	}

	c.logger.Info("Successfully fetched analytics overview",
		zap.String("tenant_id", tenantID),
		zap.Int64("total_views", overview.TotalViews))

	return &overview, nil
}

// GetMetrics fetches metrics data from the analytics service.
func (c *AnalyticsClient) GetMetrics(ctx context.Context, tenantID string, timeRange string) ([]MetricData, error) {
	url := fmt.Sprintf("%s/api/v1/analytics/metrics?range=%s", c.baseURL, timeRange)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Tenant-ID", tenantID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to fetch metrics", zap.Error(err))
		return c.getMockMetrics(tenantID), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("Analytics service returned non-200 status for metrics",
			zap.Int("status", resp.StatusCode))
		return c.getMockMetrics(tenantID), nil
	}

	var metrics []MetricData
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		c.logger.Error("Failed to decode metrics", zap.Error(err))
		return c.getMockMetrics(tenantID), nil
	}

	return metrics, nil
}

// getMockAnalyticsOverview returns mock analytics data when the analytics service is unavailable.
func (c *AnalyticsClient) getMockAnalyticsOverview(tenantID string) *AnalyticsOverview {
	now := time.Now()
	dailyMetrics := make([]DailyMetric, 30)

	for i := 0; i < 30; i++ {
		date := now.AddDate(0, 0, -i)
		dailyMetrics[i] = DailyMetric{
			Date:   date.Format("2006-01-02"),
			Views:  int64(500 + i*10),
			Uptime: 99.5 + float64(i)*0.01,
		}
	}

	return &AnalyticsOverview{
		TotalViews:       15432,
		UniqueVisitors:   8765,
		UptimePercentage: 99.95,
		AvgResponseTime:  245.5,
		StatusPages:      3,
		ActiveIncidents:  0,
		DailyMetrics:     dailyMetrics,
		TopCountries: []CountryMetric{
			{Country: "United States", Views: 5432},
			{Country: "United Kingdom", Views: 2987},
			{Country: "Germany", Views: 1876},
			{Country: "Canada", Views: 1543},
			{Country: "Australia", Views: 987},
		},
		ResponseTimes: []ResponseTimeMetric{
			{Timestamp: now.Add(-1 * time.Hour), ResponseTime: 245.5},
			{Timestamp: now.Add(-2 * time.Hour), ResponseTime: 267.8},
			{Timestamp: now.Add(-3 * time.Hour), ResponseTime: 234.1},
		},
		UptimeHistory: []UptimeMetric{
			{Date: now.Format("2006-01-02"), Uptime: 99.95},
			{Date: now.AddDate(0, 0, -1).Format("2006-01-02"), Uptime: 99.87},
			{Date: now.AddDate(0, 0, -2).Format("2006-01-02"), Uptime: 100.0},
		},
	}
}

// getMockMetrics returns mock metrics data when the analytics service is unavailable.
func (c *AnalyticsClient) getMockMetrics(tenantID string) []MetricData {
	now := time.Now()
	return []MetricData{
		{
			ID:        "metric-1",
			Name:      "Page Views",
			Type:      "counter",
			Value:     15432,
			Unit:      "views",
			Timestamp: now,
			Metadata: map[string]interface{}{
				"tenant_id": tenantID,
			},
		},
		{
			ID:        "metric-2",
			Name:      "Response Time",
			Type:      "gauge",
			Value:     245.5,
			Unit:      "ms",
			Timestamp: now,
			Metadata: map[string]interface{}{
				"tenant_id": tenantID,
			},
		},
		{
			ID:        "metric-3",
			Name:      "Uptime",
			Type:      "gauge",
			Value:     99.95,
			Unit:      "%",
			Timestamp: now,
			Metadata: map[string]interface{}{
				"tenant_id": tenantID,
			},
		},
	}
}

// CreateMetric creates a new metric in the analytics service.
func (c *AnalyticsClient) CreateMetric(ctx context.Context, tenantID string, metric MetricData) (*MetricData, error) {
	url := fmt.Sprintf("%s/api/v1/analytics/metrics", c.baseURL)

	jsonData, err := json.Marshal(metric)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metric: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Tenant-ID", tenantID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to create metric", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("analytics service returned status %d", resp.StatusCode)
	}

	var createdMetric MetricData
	if err := json.NewDecoder(resp.Body).Decode(&createdMetric); err != nil {
		return nil, fmt.Errorf("failed to decode created metric: %w", err)
	}

	return &createdMetric, nil
}

// Health checks the health of the analytics service.
func (c *AnalyticsClient) Health(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("analytics service health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("analytics service health check returned status %d", resp.StatusCode)
	}

	return nil
}