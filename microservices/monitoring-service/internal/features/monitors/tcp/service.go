package tcp

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TCPPortMonitor represents a TCP port monitoring configuration
type TCPPortMonitor struct {
	ID               uint      `gorm:"primaryKey"`
	TenantID         uuid.UUID `gorm:"type:uuid;not null;index:idx_tcp_monitor_tenant"`
	Name             string    `gorm:"type:varchar(255);not null"`
	Host             string    `gorm:"type:varchar(255);not null"` // IP address or hostname
	Port             int       `gorm:"not null"`
	Timeout          int       `gorm:"default:10"` // Timeout in seconds
	CheckInterval    int       `gorm:"default:60"` // Check interval in seconds
	FailureThreshold int       `gorm:"default:3"`  // Consecutive failures before alert
	IsActive         bool      `gorm:"default:true;index:idx_tcp_monitor_active"`

	// Current status
	Status              string     `gorm:"type:varchar(20);default:'unknown'"` // operational, down, degraded
	LastChecked         *time.Time
	LastOperational     *time.Time
	LastDown            *time.Time
	ConsecutiveFailures int `gorm:"default:0"`
	ConsecutiveSuccesses int `gorm:"default:0"`

	// Metadata
	Description string    `gorm:"type:text"`
	Tags        string    `gorm:"type:text"` // JSON array of tags
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TCPPortCheckResult represents a single TCP port check result
type TCPPortCheckResult struct {
	ID            uint      `gorm:"primaryKey"`
	MonitorID     uint      `gorm:"not null;index:idx_tcp_result_monitor"`
	CheckedAt     time.Time `gorm:"not null;index:idx_tcp_result_time"`
	Status        string    `gorm:"type:varchar(20);not null"` // operational, down
	ConnectionTimeMS int    `gorm:"column:connection_time_ms"`
	IsOpen        bool      `gorm:"not null"`
	ErrorMessage  string    `gorm:"type:text"`
	CreatedAt     time.Time
}

// TCPPortMonitoringService provides TCP port monitoring capabilities
type TCPPortMonitoringService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewTCPPortMonitoringService creates a new TCP port monitoring service
func NewTCPPortMonitoringService(db *gorm.DB, logger *zap.Logger) *TCPPortMonitoringService {
	return &TCPPortMonitoringService{
		db:     db,
		logger: logger,
	}
}

// CreateMonitor creates a new TCP port monitor
func (s *TCPPortMonitoringService) CreateMonitor(monitor *TCPPortMonitor) error {
	// Validate inputs
	if monitor.Host == "" {
		return fmt.Errorf("host is required")
	}
	if monitor.Port <= 0 || monitor.Port > 65535 {
		return fmt.Errorf("invalid port number: must be between 1 and 65535")
	}

	// Set defaults
	if monitor.Timeout == 0 {
		monitor.Timeout = 10
	}
	if monitor.CheckInterval == 0 {
		monitor.CheckInterval = 60
	}
	if monitor.FailureThreshold == 0 {
		monitor.FailureThreshold = 3
	}
	if monitor.Status == "" {
		monitor.Status = "unknown"
	}

	err := s.db.Create(monitor).Error
	if err != nil {
		s.logger.Error("Failed to create TCP port monitor", zap.Error(err))
		return fmt.Errorf("failed to create monitor: %w", err)
	}

	s.logger.Info("TCP port monitor created",
		zap.Uint("monitor_id", monitor.ID),
		zap.String("host", monitor.Host),
		zap.Int("port", monitor.Port))

	return nil
}

// GetMonitor retrieves a TCP port monitor by ID
func (s *TCPPortMonitoringService) GetMonitor(id uint, tenantID uuid.UUID) (*TCPPortMonitor, error) {
	var monitor TCPPortMonitor
	err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&monitor).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("monitor not found")
		}
		return nil, err
	}
	return &monitor, nil
}

// GetMonitors retrieves all TCP port monitors for a tenant
func (s *TCPPortMonitoringService) GetMonitors(tenantID uuid.UUID, activeOnly bool) ([]TCPPortMonitor, error) {
	var monitors []TCPPortMonitor
	query := s.db.Where("tenant_id = ?", tenantID)

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("created_at DESC").Find(&monitors).Error
	if err != nil {
		s.logger.Error("Failed to fetch TCP monitors", zap.Error(err))
		return nil, err
	}

	return monitors, nil
}

// UpdateMonitor updates a TCP port monitor
func (s *TCPPortMonitoringService) UpdateMonitor(monitor *TCPPortMonitor) error {
	err := s.db.Save(monitor).Error
	if err != nil {
		s.logger.Error("Failed to update TCP monitor", zap.Error(err))
		return err
	}
	return nil
}

// DeleteMonitor deletes a TCP port monitor
func (s *TCPPortMonitoringService) DeleteMonitor(id uint, tenantID uuid.UUID) error {
	err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&TCPPortMonitor{}).Error
	if err != nil {
		s.logger.Error("Failed to delete TCP monitor", zap.Error(err))
		return err
	}
	return nil
}

// CheckPort performs a TCP port check
func (s *TCPPortMonitoringService) CheckPort(monitor *TCPPortMonitor) (*TCPPortCheckResult, error) {
	startTime := time.Now()

	result := &TCPPortCheckResult{
		MonitorID: monitor.ID,
		CheckedAt: time.Now(),
	}

	// Create connection with timeout
	address := fmt.Sprintf("%s:%d", monitor.Host, monitor.Port)
	timeout := time.Duration(monitor.Timeout) * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	dialer := &net.Dialer{
		Timeout: timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", address)

	// Calculate connection time
	connectionTime := int(time.Since(startTime).Milliseconds())
	result.ConnectionTimeMS = connectionTime

	if err != nil {
		// Connection failed
		result.Status = "down"
		result.IsOpen = false
		result.ErrorMessage = err.Error()

		s.logger.Warn("TCP port check failed",
			zap.String("host", monitor.Host),
			zap.Int("port", monitor.Port),
			zap.Error(err))

		return result, nil
	}

	// Connection successful
	defer conn.Close()

	result.Status = "operational"
	result.IsOpen = true

	s.logger.Info("TCP port check succeeded",
		zap.String("host", monitor.Host),
		zap.Int("port", monitor.Port),
		zap.Int("connection_time_ms", connectionTime))

	return result, nil
}

// PerformCheck performs a check and updates monitor status
func (s *TCPPortMonitoringService) PerformCheck(monitor *TCPPortMonitor) error {
	// Perform the check
	result, err := s.CheckPort(monitor)
	if err != nil {
		s.logger.Error("Failed to perform TCP check", zap.Error(err))
		return err
	}

	// Save result to database
	if err := s.db.Create(result).Error; err != nil {
		s.logger.Error("Failed to save TCP check result", zap.Error(err))
		return err
	}

	// Update monitor status
	now := time.Now()
	monitor.LastChecked = &now

	if result.Status == "operational" {
		monitor.Status = "operational"
		monitor.LastOperational = &now
		monitor.ConsecutiveSuccesses++
		monitor.ConsecutiveFailures = 0
	} else {
		monitor.ConsecutiveFailures++
		monitor.ConsecutiveSuccesses = 0

		if monitor.ConsecutiveFailures >= monitor.FailureThreshold {
			monitor.Status = "down"
			monitor.LastDown = &now

			// TODO: Trigger alert
			s.logger.Warn("TCP port monitor threshold reached",
				zap.Uint("monitor_id", monitor.ID),
				zap.Int("consecutive_failures", monitor.ConsecutiveFailures))
		}
	}

	if err := s.db.Save(monitor).Error; err != nil {
		s.logger.Error("Failed to update monitor status", zap.Error(err))
		return err
	}

	return nil
}

// CheckAllActiveMonitors checks all active TCP port monitors for a tenant
func (s *TCPPortMonitoringService) CheckAllActiveMonitors(tenantID uuid.UUID) error {
	monitors, err := s.GetMonitors(tenantID, true)
	if err != nil {
		return err
	}

	for _, monitor := range monitors {
		if err := s.PerformCheck(&monitor); err != nil {
			s.logger.Error("Failed to check monitor",
				zap.Uint("monitor_id", monitor.ID),
				zap.Error(err))
		}
	}

	return nil
}

// GetCheckResults retrieves recent check results for a monitor
func (s *TCPPortMonitoringService) GetCheckResults(monitorID uint, limit int) ([]TCPPortCheckResult, error) {
	var results []TCPPortCheckResult
	err := s.db.Where("monitor_id = ?", monitorID).
		Order("checked_at DESC").
		Limit(limit).
		Find(&results).Error

	if err != nil {
		s.logger.Error("Failed to fetch check results", zap.Error(err))
		return nil, err
	}

	return results, nil
}

// GetMonitorStats calculates statistics for a TCP port monitor
func (s *TCPPortMonitoringService) GetMonitorStats(monitorID uint, timeRange string) (map[string]interface{}, error) {
	// Parse time range
	var startTime time.Time
	now := time.Now()

	switch timeRange {
	case "1h":
		startTime = now.Add(-1 * time.Hour)
	case "24h":
		startTime = now.Add(-24 * time.Hour)
	case "7d":
		startTime = now.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = now.Add(-30 * 24 * time.Hour)
	default:
		startTime = now.Add(-24 * time.Hour)
	}

	// Fetch results
	var results []TCPPortCheckResult
	err := s.db.Where("monitor_id = ? AND checked_at >= ?", monitorID, startTime).
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return map[string]interface{}{
			"total_checks": 0,
		}, nil
	}

	// Calculate statistics
	totalChecks := len(results)
	successfulChecks := 0
	var totalConnectionTime int64
	minConnectionTime := 999999
	maxConnectionTime := 0

	for _, result := range results {
		if result.IsOpen {
			successfulChecks++
		}

		totalConnectionTime += int64(result.ConnectionTimeMS)

		if result.ConnectionTimeMS < minConnectionTime {
			minConnectionTime = result.ConnectionTimeMS
		}
		if result.ConnectionTimeMS > maxConnectionTime {
			maxConnectionTime = result.ConnectionTimeMS
		}
	}

	uptimePercentage := float64(successfulChecks) / float64(totalChecks) * 100
	avgConnectionTime := float64(totalConnectionTime) / float64(totalChecks)

	stats := map[string]interface{}{
		"total_checks":          totalChecks,
		"successful_checks":     successfulChecks,
		"failed_checks":         totalChecks - successfulChecks,
		"uptime_percentage":     uptimePercentage,
		"avg_connection_time_ms": avgConnectionTime,
		"min_connection_time_ms": minConnectionTime,
		"max_connection_time_ms": maxConnectionTime,
		"time_range":            timeRange,
	}

	return stats, nil
}

// StartPeriodicChecks starts periodic checking for a monitor
func (s *TCPPortMonitoringService) StartPeriodicChecks(monitor *TCPPortMonitor, stopChan <-chan struct{}) {
	ticker := time.NewTicker(time.Duration(monitor.CheckInterval) * time.Second)
	defer ticker.Stop()

	s.logger.Info("Starting periodic TCP port checks",
		zap.Uint("monitor_id", monitor.ID),
		zap.String("host", monitor.Host),
		zap.Int("port", monitor.Port),
		zap.Int("interval_seconds", monitor.CheckInterval))

	// Perform initial check immediately
	s.PerformCheck(monitor)

	for {
		select {
		case <-ticker.C:
			s.PerformCheck(monitor)
		case <-stopChan:
			s.logger.Info("Stopping periodic TCP port checks",
				zap.Uint("monitor_id", monitor.ID))
			return
		}
	}
}

// TestConnection tests if a TCP port is accessible (one-time check)
func (s *TCPPortMonitoringService) TestConnection(host string, port int, timeout int) (bool, int, error) {
	startTime := time.Now()

	address := fmt.Sprintf("%s:%d", host, port)
	timeoutDuration := time.Duration(timeout) * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	dialer := &net.Dialer{
		Timeout: timeoutDuration,
	}

	conn, err := dialer.DialContext(ctx, "tcp", address)

	connectionTime := int(time.Since(startTime).Milliseconds())

	if err != nil {
		return false, connectionTime, err
	}

	defer conn.Close()
	return true, connectionTime, nil
}
