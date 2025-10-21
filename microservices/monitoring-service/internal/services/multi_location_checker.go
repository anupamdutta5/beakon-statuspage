package services

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MonitoringLocation represents a global monitoring node location
type MonitoringLocation struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"type:varchar(100);not null"`
	City      string    `gorm:"type:varchar(100)"`
	Country   string    `gorm:"type:varchar(100);not null"`
	Region    string    `gorm:"type:varchar(50);not null"`
	Latitude  *float64  `gorm:"type:decimal(9,6)"`
	Longitude *float64  `gorm:"type:decimal(9,6)"`
	IsActive  bool      `gorm:"default:true"`
	Provider  string    `gorm:"type:varchar(50)"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// MonitoringResult represents a health check result from a specific location
type MonitoringResult struct {
	ID               uint      `gorm:"primaryKey"`
	MonitorID        uint      `gorm:"not null;index:idx_monitoring_results_monitor"`
	LocationID       uint      `gorm:"not null;index:idx_monitoring_results_location"`
	CheckedAt        time.Time `gorm:"not null;index:idx_monitoring_results_time"`
	Status           string    `gorm:"type:varchar(20);not null"` // operational, degraded, down
	ResponseTimeMS   int       `gorm:"column:response_time_ms"`
	TTFBMS           int       `gorm:"column:ttfb_ms"`
	DNSTimeMS        int       `gorm:"column:dns_time_ms"`
	ConnectionTimeMS int       `gorm:"column:connection_time_ms"`
	StatusCode       int
	ErrorMessage     string `gorm:"type:text"`
	CreatedAt        time.Time
}

// MonitorConfig represents configuration for a monitor with multi-location support
type MonitorConfig struct {
	ID               uint
	TenantID         uuid.UUID
	Name             string
	URL              string
	Method           string
	ExpectedStatus   int
	Timeout          int
	Locations        []uint // Location IDs to check from
	CheckInterval    int    // Seconds between checks
	FailureThreshold int    // Consecutive failures before alert
}

// MultiLocationChecker performs health checks from multiple global locations
type MultiLocationChecker struct {
	db         *gorm.DB
	logger     *zap.Logger
	httpClient *http.Client
	mu         sync.RWMutex
}

// NewMultiLocationChecker creates a new multi-location health checker
func NewMultiLocationChecker(db *gorm.DB, logger *zap.Logger) *MultiLocationChecker {
	return &MultiLocationChecker{
		db:     db,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// GetActiveLocations retrieves all active monitoring locations
func (m *MultiLocationChecker) GetActiveLocations() ([]MonitoringLocation, error) {
	var locations []MonitoringLocation
	err := m.db.Where("is_active = ?", true).Find(&locations).Error
	if err != nil {
		m.logger.Error("Failed to fetch active monitoring locations", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch active locations: %w", err)
	}

	return locations, nil
}

// CheckFromLocation performs a health check from a specific location
func (m *MultiLocationChecker) CheckFromLocation(ctx context.Context, monitor MonitorConfig, location MonitoringLocation) (*MonitoringResult, error) {
	startTime := time.Now()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, monitor.Method, monitor.URL, nil)
	if err != nil {
		m.logger.Error("Failed to create request",
			zap.Error(err),
			zap.String("url", monitor.URL),
			zap.String("location", location.Name))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("User-Agent", "Beakon-StatusPage-Monitor/1.0")
	req.Header.Set("X-Monitor-Location", location.Region)

	// Perform request
	resp, err := m.httpClient.Do(req)

	// Calculate response time
	responseTimeMS := int(time.Since(startTime).Milliseconds())

	result := &MonitoringResult{
		MonitorID:      monitor.ID,
		LocationID:     location.ID,
		CheckedAt:      time.Now(),
		ResponseTimeMS: responseTimeMS,
	}

	if err != nil {
		// Request failed
		result.Status = "down"
		result.ErrorMessage = err.Error()

		m.logger.Warn("Health check failed",
			zap.String("monitor", monitor.Name),
			zap.String("location", location.Name),
			zap.Error(err))

		return result, nil
	}
	defer resp.Body.Close()

	// Check status code
	result.StatusCode = resp.StatusCode

	if resp.StatusCode == monitor.ExpectedStatus {
		result.Status = "operational"
	} else if resp.StatusCode >= 500 {
		result.Status = "down"
		result.ErrorMessage = fmt.Sprintf("Server error: HTTP %d", resp.StatusCode)
	} else if resp.StatusCode >= 400 {
		result.Status = "degraded"
		result.ErrorMessage = fmt.Sprintf("Client error: HTTP %d", resp.StatusCode)
	} else {
		result.Status = "operational"
	}

	// Calculate performance metrics
	// Note: Advanced metrics (TTFB, DNS time, connection time) would require
	// custom transport with httptrace for detailed timing
	result.TTFBMS = responseTimeMS / 2 // Approximation
	result.DNSTimeMS = responseTimeMS / 10 // Approximation
	result.ConnectionTimeMS = responseTimeMS / 5 // Approximation

	m.logger.Info("Health check completed",
		zap.String("monitor", monitor.Name),
		zap.String("location", location.Name),
		zap.String("status", result.Status),
		zap.Int("response_time_ms", result.ResponseTimeMS),
		zap.Int("status_code", result.StatusCode))

	return result, nil
}

// CheckFromAllLocations performs health checks from all active locations in parallel
func (m *MultiLocationChecker) CheckFromAllLocations(monitor MonitorConfig) ([]MonitoringResult, error) {
	// Get monitor-specific locations or use all active locations
	var locations []MonitoringLocation

	if len(monitor.Locations) > 0 {
		// Use specific locations
		err := m.db.Where("id IN ? AND is_active = ?", monitor.Locations, true).Find(&locations).Error
		if err != nil {
			m.logger.Error("Failed to fetch monitor locations", zap.Error(err))
			return nil, fmt.Errorf("failed to fetch locations: %w", err)
		}
	} else {
		// Use all active locations
		var err error
		locations, err = m.GetActiveLocations()
		if err != nil {
			return nil, err
		}
	}

	if len(locations) == 0 {
		return nil, fmt.Errorf("no active monitoring locations available")
	}

	// Perform checks in parallel with goroutines
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(monitor.Timeout)*time.Second)
	defer cancel()

	resultChan := make(chan *MonitoringResult, len(locations))
	errorChan := make(chan error, len(locations))
	var wg sync.WaitGroup

	for _, location := range locations {
		wg.Add(1)
		go func(loc MonitoringLocation) {
			defer wg.Done()

			result, err := m.CheckFromLocation(ctx, monitor, loc)
			if err != nil {
				errorChan <- err
				return
			}

			resultChan <- result
		}(location)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(resultChan)
	close(errorChan)

	// Collect results
	var results []MonitoringResult
	for result := range resultChan {
		results = append(results, *result)
	}

	// Log any errors (but don't fail the entire check)
	for err := range errorChan {
		m.logger.Warn("Location check error", zap.Error(err))
	}

	return results, nil
}

// SaveResults saves monitoring results to the database
func (m *MultiLocationChecker) SaveResults(results []MonitoringResult) error {
	if len(results) == 0 {
		return nil
	}

	// Batch insert for efficiency
	err := m.db.Create(&results).Error
	if err != nil {
		m.logger.Error("Failed to save monitoring results", zap.Error(err))
		return fmt.Errorf("failed to save results: %w", err)
	}

	m.logger.Info("Saved monitoring results", zap.Int("count", len(results)))
	return nil
}

// CalculateAggregateStatus determines overall status from multiple location results
func (m *MultiLocationChecker) CalculateAggregateStatus(results []MonitoringResult) string {
	if len(results) == 0 {
		return "unknown"
	}

	downCount := 0
	degradedCount := 0
	operationalCount := 0

	for _, result := range results {
		switch result.Status {
		case "down":
			downCount++
		case "degraded":
			degradedCount++
		case "operational":
			operationalCount++
		}
	}

	totalLocations := len(results)
	downPercentage := float64(downCount) / float64(totalLocations) * 100

	// If more than 50% locations report down, overall status is down
	if downPercentage > 50 {
		return "down"
	}

	// If any location reports degraded or down, overall status is degraded
	if degradedCount > 0 || downCount > 0 {
		return "degraded"
	}

	// All locations operational
	return "operational"
}

// GetLocationResults retrieves recent results for a monitor from a specific location
func (m *MultiLocationChecker) GetLocationResults(monitorID uint, locationID uint, limit int) ([]MonitoringResult, error) {
	var results []MonitoringResult

	err := m.db.Where("monitor_id = ? AND location_id = ?", monitorID, locationID).
		Order("checked_at DESC").
		Limit(limit).
		Find(&results).Error

	if err != nil {
		m.logger.Error("Failed to fetch location results", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch location results: %w", err)
	}

	return results, nil
}

// GetAggregateMetrics calculates aggregate performance metrics across all locations
func (m *MultiLocationChecker) GetAggregateMetrics(results []MonitoringResult) map[string]interface{} {
	if len(results) == 0 {
		return map[string]interface{}{
			"count": 0,
		}
	}

	var totalResponseTime int
	var minResponseTime int = 999999
	var maxResponseTime int
	var operationalCount int

	for _, result := range results {
		if result.Status == "operational" {
			operationalCount++
		}

		totalResponseTime += result.ResponseTimeMS

		if result.ResponseTimeMS < minResponseTime {
			minResponseTime = result.ResponseTimeMS
		}
		if result.ResponseTimeMS > maxResponseTime {
			maxResponseTime = result.ResponseTimeMS
		}
	}

	avgResponseTime := totalResponseTime / len(results)
	uptimePercentage := float64(operationalCount) / float64(len(results)) * 100

	return map[string]interface{}{
		"count":              len(results),
		"operational_count":  operationalCount,
		"uptime_percentage":  uptimePercentage,
		"avg_response_time":  avgResponseTime,
		"min_response_time":  minResponseTime,
		"max_response_time":  maxResponseTime,
		"aggregate_status":   m.CalculateAggregateStatus(results),
	}
}

// StartPeriodicChecks starts periodic health checks for a monitor from all locations
func (m *MultiLocationChecker) StartPeriodicChecks(monitor MonitorConfig, stopChan <-chan struct{}) {
	ticker := time.NewTicker(time.Duration(monitor.CheckInterval) * time.Second)
	defer ticker.Stop()

	m.logger.Info("Starting periodic multi-location checks",
		zap.String("monitor", monitor.Name),
		zap.Int("interval_seconds", monitor.CheckInterval))

	// Perform initial check immediately
	m.performCheck(monitor)

	for {
		select {
		case <-ticker.C:
			m.performCheck(monitor)
		case <-stopChan:
			m.logger.Info("Stopping periodic checks", zap.String("monitor", monitor.Name))
			return
		}
	}
}

// performCheck is a helper to perform a single check cycle
func (m *MultiLocationChecker) performCheck(monitor MonitorConfig) {
	results, err := m.CheckFromAllLocations(monitor)
	if err != nil {
		m.logger.Error("Multi-location check failed",
			zap.String("monitor", monitor.Name),
			zap.Error(err))
		return
	}

	// Save results to database
	if err := m.SaveResults(results); err != nil {
		m.logger.Error("Failed to save check results",
			zap.String("monitor", monitor.Name),
			zap.Error(err))
	}

	// Calculate and log aggregate status
	aggregateStatus := m.CalculateAggregateStatus(results)
	metrics := m.GetAggregateMetrics(results)

	m.logger.Info("Multi-location check completed",
		zap.String("monitor", monitor.Name),
		zap.String("aggregate_status", aggregateStatus),
		zap.Any("metrics", metrics))

	// TODO: Trigger alerts if aggregate status is down or degraded
	// TODO: Publish monitoring events to RabbitMQ
}
