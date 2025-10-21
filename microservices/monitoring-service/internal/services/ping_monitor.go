package services

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PingMonitor represents an ICMP ping monitoring configuration
type PingMonitor struct {
	ID               uint      `gorm:"primaryKey"`
	TenantID         uuid.UUID `gorm:"type:uuid;not null;index:idx_ping_monitor_tenant"`
	Name             string    `gorm:"type:varchar(255);not null"`
	Host             string    `gorm:"type:varchar(255);not null"` // IP address or hostname
	PacketCount      int       `gorm:"default:4"`                   // Number of ping packets
	Timeout          int       `gorm:"default:10"`                  // Timeout in seconds
	PacketSize       int       `gorm:"default:64"`                  // Packet size in bytes
	CheckInterval    int       `gorm:"default:60"`                  // Check interval in seconds
	FailureThreshold int       `gorm:"default:3"`                   // Consecutive failures before alert
	SuccessThreshold int       `gorm:"default:75"`                  // Min % packets received for success
	IsActive         bool      `gorm:"default:true;index:idx_ping_monitor_active"`

	// Current status
	Status              string     `gorm:"type:varchar(20);default:'unknown'"` // operational, down, degraded
	LastChecked         *time.Time
	LastOperational     *time.Time
	LastDown            *time.Time
	ConsecutiveFailures int `gorm:"default:0"`
	ConsecutiveSuccesses int `gorm:"default:0"`
	AvgLatency          float64 `gorm:"default:0"` // Average latency in ms

	// Metadata
	Description string    `gorm:"type:text"`
	Tags        string    `gorm:"type:text"` // JSON array of tags
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PingCheckResult represents a single ping check result
type PingCheckResult struct {
	ID              uint      `gorm:"primaryKey"`
	MonitorID       uint      `gorm:"not null;index:idx_ping_result_monitor"`
	CheckedAt       time.Time `gorm:"not null;index:idx_ping_result_time"`
	Status          string    `gorm:"type:varchar(20);not null"` // operational, down, degraded
	PacketsSent     int       `gorm:"not null"`
	PacketsReceived int       `gorm:"not null"`
	PacketLoss      float64   `gorm:"type:decimal(5,2)"` // Packet loss percentage
	MinLatencyMS    float64   `gorm:"column:min_latency_ms"`
	AvgLatencyMS    float64   `gorm:"column:avg_latency_ms"`
	MaxLatencyMS    float64   `gorm:"column:max_latency_ms"`
	StdDevLatencyMS float64   `gorm:"column:stddev_latency_ms"`
	ErrorMessage    string    `gorm:"type:text"`
	CreatedAt       time.Time
}

// PingMonitoringService provides ICMP ping monitoring capabilities
type PingMonitoringService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPingMonitoringService creates a new ping monitoring service
func NewPingMonitoringService(db *gorm.DB, logger *zap.Logger) *PingMonitoringService {
	return &PingMonitoringService{
		db:     db,
		logger: logger,
	}
}

// CreateMonitor creates a new ping monitor
func (s *PingMonitoringService) CreateMonitor(monitor *PingMonitor) error {
	// Validate inputs
	if monitor.Host == "" {
		return fmt.Errorf("host is required")
	}

	// Set defaults
	if monitor.PacketCount == 0 {
		monitor.PacketCount = 4
	}
	if monitor.Timeout == 0 {
		monitor.Timeout = 10
	}
	if monitor.PacketSize == 0 {
		monitor.PacketSize = 64
	}
	if monitor.CheckInterval == 0 {
		monitor.CheckInterval = 60
	}
	if monitor.FailureThreshold == 0 {
		monitor.FailureThreshold = 3
	}
	if monitor.SuccessThreshold == 0 {
		monitor.SuccessThreshold = 75
	}
	if monitor.Status == "" {
		monitor.Status = "unknown"
	}

	err := s.db.Create(monitor).Error
	if err != nil {
		s.logger.Error("Failed to create ping monitor", zap.Error(err))
		return fmt.Errorf("failed to create monitor: %w", err)
	}

	s.logger.Info("Ping monitor created",
		zap.Uint("monitor_id", monitor.ID),
		zap.String("host", monitor.Host))

	return nil
}

// GetMonitor retrieves a ping monitor by ID
func (s *PingMonitoringService) GetMonitor(id uint, tenantID uuid.UUID) (*PingMonitor, error) {
	var monitor PingMonitor
	err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&monitor).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("monitor not found")
		}
		return nil, err
	}
	return &monitor, nil
}

// GetMonitors retrieves all ping monitors for a tenant
func (s *PingMonitoringService) GetMonitors(tenantID uuid.UUID, activeOnly bool) ([]PingMonitor, error) {
	var monitors []PingMonitor
	query := s.db.Where("tenant_id = ?", tenantID)

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("created_at DESC").Find(&monitors).Error
	if err != nil {
		s.logger.Error("Failed to fetch ping monitors", zap.Error(err))
		return nil, err
	}

	return monitors, nil
}

// UpdateMonitor updates a ping monitor
func (s *PingMonitoringService) UpdateMonitor(monitor *PingMonitor) error {
	err := s.db.Save(monitor).Error
	if err != nil {
		s.logger.Error("Failed to update ping monitor", zap.Error(err))
		return err
	}
	return nil
}

// DeleteMonitor deletes a ping monitor
func (s *PingMonitoringService) DeleteMonitor(id uint, tenantID uuid.UUID) error {
	err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&PingMonitor{}).Error
	if err != nil {
		s.logger.Error("Failed to delete ping monitor", zap.Error(err))
		return err
	}
	return nil
}

// Ping performs an ICMP ping using system ping command
func (s *PingMonitoringService) Ping(monitor *PingMonitor) (*PingCheckResult, error) {
	result := &PingCheckResult{
		MonitorID:   monitor.ID,
		CheckedAt:   time.Now(),
		PacketsSent: monitor.PacketCount,
	}

	// Use system ping command (cross-platform)
	output, err := s.executePing(monitor.Host, monitor.PacketCount, monitor.Timeout, monitor.PacketSize)

	if err != nil {
		result.Status = "down"
		result.PacketsReceived = 0
		result.PacketLoss = 100.0
		result.ErrorMessage = err.Error()

		s.logger.Warn("Ping check failed",
			zap.String("host", monitor.Host),
			zap.Error(err))

		return result, nil
	}

	// Parse ping output
	stats, parseErr := s.parsePingOutput(output, monitor.PacketCount)
	if parseErr != nil {
		result.Status = "down"
		result.ErrorMessage = parseErr.Error()
		s.logger.Error("Failed to parse ping output", zap.Error(parseErr))
		return result, nil
	}

	// Populate result
	result.PacketsReceived = stats.PacketsReceived
	result.PacketLoss = stats.PacketLoss
	result.MinLatencyMS = stats.MinLatency
	result.AvgLatencyMS = stats.AvgLatency
	result.MaxLatencyMS = stats.MaxLatency
	result.StdDevLatencyMS = stats.StdDevLatency

	// Determine status based on packet loss
	successRate := float64(result.PacketsReceived) / float64(result.PacketsSent) * 100

	if successRate >= float64(monitor.SuccessThreshold) {
		result.Status = "operational"
	} else if successRate > 0 {
		result.Status = "degraded"
	} else {
		result.Status = "down"
	}

	s.logger.Info("Ping check completed",
		zap.String("host", monitor.Host),
		zap.String("status", result.Status),
		zap.Float64("avg_latency_ms", result.AvgLatencyMS),
		zap.Float64("packet_loss", result.PacketLoss))

	return result, nil
}

// PingStats represents parsed ping statistics
type PingStats struct {
	PacketsReceived int
	PacketLoss      float64
	MinLatency      float64
	AvgLatency      float64
	MaxLatency      float64
	StdDevLatency   float64
}

// executePing executes system ping command
func (s *PingMonitoringService) executePing(host string, count, timeout, packetSize int) (string, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows ping command: ping -n count -w timeout*1000 -l packetSize host
		cmd = exec.Command("ping", "-n", strconv.Itoa(count), "-w", strconv.Itoa(timeout*1000), "-l", strconv.Itoa(packetSize), host)
	case "darwin", "linux":
		// macOS/Linux ping command: ping -c count -W timeout -s packetSize host
		cmd = exec.Command("ping", "-c", strconv.Itoa(count), "-W", strconv.Itoa(timeout), "-s", strconv.Itoa(packetSize), host)
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ping command failed: %w (output: %s)", err, string(output))
	}

	return string(output), nil
}

// parsePingOutput parses ping command output
func (s *PingMonitoringService) parsePingOutput(output string, expectedCount int) (*PingStats, error) {
	stats := &PingStats{}

	lines := strings.Split(output, "\n")

	// Parse statistics line (different format for Windows vs Unix)
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse packet statistics
		if strings.Contains(line, "packets transmitted") || strings.Contains(line, "Packets: Sent") {
			stats.PacketsReceived = s.extractPacketsReceived(line)
			stats.PacketLoss = s.extractPacketLoss(line)
		}

		// Parse latency statistics
		if strings.Contains(line, "rtt") || strings.Contains(line, "round-trip") || strings.Contains(line, "Minimum") {
			stats.MinLatency, stats.AvgLatency, stats.MaxLatency, stats.StdDevLatency = s.extractLatencyStats(line)
		}
	}

	return stats, nil
}

// extractPacketsReceived extracts packets received from ping output
func (s *PingMonitoringService) extractPacketsReceived(line string) int {
	// Unix format: "4 packets transmitted, 4 received, 0% packet loss"
	if strings.Contains(line, "packets transmitted") {
		parts := strings.Fields(line)
		for i, part := range parts {
			if part == "received," && i > 0 {
				if count, err := strconv.Atoi(parts[i-1]); err == nil {
					return count
				}
			}
		}
	}

	// Windows format: "Packets: Sent = 4, Received = 4, Lost = 0 (0% loss)"
	if strings.Contains(line, "Received =") {
		parts := strings.Split(line, "Received =")
		if len(parts) > 1 {
			fields := strings.Fields(parts[1])
			if len(fields) > 0 {
				numStr := strings.TrimSuffix(fields[0], ",")
				if count, err := strconv.Atoi(numStr); err == nil {
					return count
				}
			}
		}
	}

	return 0
}

// extractPacketLoss extracts packet loss percentage from ping output
func (s *PingMonitoringService) extractPacketLoss(line string) float64 {
	// Look for "X% packet loss" or "X% loss"
	parts := strings.Fields(line)
	for _, part := range parts {
		if strings.HasSuffix(part, "%") {
			lossStr := strings.TrimSuffix(part, "%")
			if loss, err := strconv.ParseFloat(lossStr, 64); err == nil {
				return loss
			}
		}
	}
	return 0.0
}

// extractLatencyStats extracts latency statistics from ping output
func (s *PingMonitoringService) extractLatencyStats(line string) (min, avg, max, stddev float64) {
	// Unix format: "rtt min/avg/max/mdev = 12.345/23.456/34.567/5.678 ms"
	if strings.Contains(line, "min/avg/max") {
		parts := strings.Split(line, "=")
		if len(parts) > 1 {
			stats := strings.TrimSpace(parts[1])
			stats = strings.TrimSuffix(stats, " ms")
			values := strings.Split(stats, "/")
			if len(values) >= 3 {
				min, _ = strconv.ParseFloat(values[0], 64)
				avg, _ = strconv.ParseFloat(values[1], 64)
				max, _ = strconv.ParseFloat(values[2], 64)
				if len(values) >= 4 {
					stddev, _ = strconv.ParseFloat(values[3], 64)
				}
			}
		}
	}

	// Windows format: "Minimum = 12ms, Maximum = 34ms, Average = 23ms"
	if strings.Contains(line, "Minimum =") {
		parts := strings.Fields(line)
		for i, part := range parts {
			if part == "Minimum" && i+2 < len(parts) {
				minStr := strings.TrimSuffix(strings.TrimSuffix(parts[i+2], "ms"), ",")
				min, _ = strconv.ParseFloat(minStr, 64)
			}
			if part == "Maximum" && i+2 < len(parts) {
				maxStr := strings.TrimSuffix(strings.TrimSuffix(parts[i+2], "ms"), ",")
				max, _ = strconv.ParseFloat(maxStr, 64)
			}
			if part == "Average" && i+2 < len(parts) {
				avgStr := strings.TrimSuffix(strings.TrimSuffix(parts[i+2], "ms"), ",")
				avg, _ = strconv.ParseFloat(avgStr, 64)
			}
		}
	}

	return
}

// PerformCheck performs a ping check and updates monitor status
func (s *PingMonitoringService) PerformCheck(monitor *PingMonitor) error {
	// Perform the ping
	result, err := s.Ping(monitor)
	if err != nil {
		s.logger.Error("Failed to perform ping check", zap.Error(err))
		return err
	}

	// Save result to database
	if err := s.db.Create(result).Error; err != nil {
		s.logger.Error("Failed to save ping check result", zap.Error(err))
		return err
	}

	// Update monitor status
	now := time.Now()
	monitor.LastChecked = &now
	monitor.AvgLatency = result.AvgLatencyMS

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
			s.logger.Warn("Ping monitor threshold reached",
				zap.Uint("monitor_id", monitor.ID),
				zap.Int("consecutive_failures", monitor.ConsecutiveFailures))
		} else {
			monitor.Status = result.Status
		}
	}

	if err := s.db.Save(monitor).Error; err != nil {
		s.logger.Error("Failed to update monitor status", zap.Error(err))
		return err
	}

	return nil
}

// GetCheckResults retrieves recent check results for a monitor
func (s *PingMonitoringService) GetCheckResults(monitorID uint, limit int) ([]PingCheckResult, error) {
	var results []PingCheckResult
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

// TestPing tests if a host is reachable via ping (one-time check)
func (s *PingMonitoringService) TestPing(host string, count, timeout int) (bool, float64, error) {
	output, err := s.executePing(host, count, timeout, 64)
	if err != nil {
		return false, 0, err
	}

	stats, err := s.parsePingOutput(output, count)
	if err != nil {
		return false, 0, err
	}

	isReachable := stats.PacketsReceived > 0
	return isReachable, stats.AvgLatency, nil
}

// ResolveHost resolves a hostname to IP address
func (s *PingMonitoringService) ResolveHost(host string) ([]string, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve host: %w", err)
	}

	ipStrings := make([]string, len(ips))
	for i, ip := range ips {
		ipStrings[i] = ip.String()
	}

	return ipStrings, nil
}
