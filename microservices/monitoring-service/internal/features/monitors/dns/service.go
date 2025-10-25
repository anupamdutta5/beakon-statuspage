package dns

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DNSMonitor represents a DNS monitoring configuration
type DNSMonitor struct {
	ID               uint      `gorm:"primaryKey"`
	TenantID         uuid.UUID `gorm:"type:uuid;not null;index:idx_dns_monitor_tenant"`
	Name             string    `gorm:"type:varchar(255);not null"`
	Hostname         string    `gorm:"type:varchar(255);not null"` // Domain to resolve
	RecordType       string    `gorm:"type:varchar(10);not null;default:'A'"` // A, AAAA, CNAME, MX, TXT, NS, SOA
	ExpectedValues   string    `gorm:"type:text"` // JSON array of expected IPs/values
	DNSServer        string    `gorm:"type:varchar(255)"` // Custom DNS server (optional)
	Timeout          int       `gorm:"default:10"` // Timeout in seconds
	CheckInterval    int       `gorm:"default:60"` // Check interval in seconds
	FailureThreshold int       `gorm:"default:3"`  // Consecutive failures before alert
	IsActive         bool      `gorm:"default:true;index:idx_dns_monitor_active"`

	// Current status
	Status              string     `gorm:"type:varchar(20);default:'unknown'"` // operational, down, degraded
	LastChecked         *time.Time
	LastOperational     *time.Time
	LastDown            *time.Time
	ConsecutiveFailures int `gorm:"default:0"`
	ConsecutiveSuccesses int `gorm:"default:0"`
	AvgResolutionTime   float64 `gorm:"default:0"` // Average DNS resolution time in ms

	// Metadata
	Description string    `gorm:"type:text"`
	Tags        string    `gorm:"type:text"` // JSON array of tags
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// DNSCheckResult represents a single DNS check result
type DNSCheckResult struct {
	ID               uint      `gorm:"primaryKey"`
	MonitorID        uint      `gorm:"not null;index:idx_dns_result_monitor"`
	CheckedAt        time.Time `gorm:"not null;index:idx_dns_result_time"`
	Status           string    `gorm:"type:varchar(20);not null"` // operational, down, degraded
	RecordType       string    `gorm:"type:varchar(10);not null"`
	ResolvedValues   string    `gorm:"type:text"` // JSON array of resolved IPs/values
	ResolutionTimeMS int       `gorm:"column:resolution_time_ms"`
	MatchesExpected  bool      `gorm:"default:false"`
	ErrorMessage     string    `gorm:"type:text"`
	CreatedAt        time.Time
}

// DNSMonitoringService provides DNS monitoring capabilities
type DNSMonitoringService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewDNSMonitoringService creates a new DNS monitoring service
func NewDNSMonitoringService(db *gorm.DB, logger *zap.Logger) *DNSMonitoringService {
	return &DNSMonitoringService{
		db:     db,
		logger: logger,
	}
}

// CreateMonitor creates a new DNS monitor
func (s *DNSMonitoringService) CreateMonitor(monitor *DNSMonitor) error {
	// Validate inputs
	if monitor.Hostname == "" {
		return fmt.Errorf("hostname is required")
	}

	// Validate record type
	validTypes := []string{"A", "AAAA", "CNAME", "MX", "TXT", "NS", "SOA", "PTR"}
	validType := false
	for _, t := range validTypes {
		if monitor.RecordType == t {
			validType = true
			break
		}
	}
	if !validType {
		return fmt.Errorf("invalid record type: must be one of %v", validTypes)
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
		s.logger.Error("Failed to create DNS monitor", zap.Error(err))
		return fmt.Errorf("failed to create monitor: %w", err)
	}

	s.logger.Info("DNS monitor created",
		zap.Uint("monitor_id", monitor.ID),
		zap.String("hostname", monitor.Hostname),
		zap.String("record_type", monitor.RecordType))

	return nil
}

// GetMonitor retrieves a DNS monitor by ID
func (s *DNSMonitoringService) GetMonitor(id uint, tenantID uuid.UUID) (*DNSMonitor, error) {
	var monitor DNSMonitor
	err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&monitor).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("monitor not found")
		}
		return nil, err
	}
	return &monitor, nil
}

// GetMonitors retrieves all DNS monitors for a tenant
func (s *DNSMonitoringService) GetMonitors(tenantID uuid.UUID, activeOnly bool) ([]DNSMonitor, error) {
	var monitors []DNSMonitor
	query := s.db.Where("tenant_id = ?", tenantID)

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("created_at DESC").Find(&monitors).Error
	if err != nil {
		s.logger.Error("Failed to fetch DNS monitors", zap.Error(err))
		return nil, err
	}

	return monitors, nil
}

// UpdateMonitor updates a DNS monitor
func (s *DNSMonitoringService) UpdateMonitor(monitor *DNSMonitor) error {
	err := s.db.Save(monitor).Error
	if err != nil {
		s.logger.Error("Failed to update DNS monitor", zap.Error(err))
		return err
	}
	return nil
}

// DeleteMonitor deletes a DNS monitor
func (s *DNSMonitoringService) DeleteMonitor(id uint, tenantID uuid.UUID) error {
	err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&DNSMonitor{}).Error
	if err != nil {
		s.logger.Error("Failed to delete DNS monitor", zap.Error(err))
		return err
	}
	return nil
}

// ResolveDNS performs DNS resolution
func (s *DNSMonitoringService) ResolveDNS(monitor *DNSMonitor) (*DNSCheckResult, error) {
	startTime := time.Now()

	result := &DNSCheckResult{
		MonitorID:  monitor.ID,
		CheckedAt:  time.Now(),
		RecordType: monitor.RecordType,
	}

	// Create resolver with timeout
	resolver := &net.Resolver{}
	if monitor.DNSServer != "" {
		// Use custom DNS server
		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Duration(monitor.Timeout) * time.Second,
				}
				return d.DialContext(ctx, "udp", monitor.DNSServer+":53")
			},
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(monitor.Timeout)*time.Second)
	defer cancel()

	var resolvedValues []string
	var err error

	// Resolve based on record type
	switch monitor.RecordType {
	case "A":
		resolvedValues, err = s.resolveA(ctx, resolver, monitor.Hostname)
	case "AAAA":
		resolvedValues, err = s.resolveAAAA(ctx, resolver, monitor.Hostname)
	case "CNAME":
		resolvedValues, err = s.resolveCNAME(ctx, resolver, monitor.Hostname)
	case "MX":
		resolvedValues, err = s.resolveMX(ctx, resolver, monitor.Hostname)
	case "TXT":
		resolvedValues, err = s.resolveTXT(ctx, resolver, monitor.Hostname)
	case "NS":
		resolvedValues, err = s.resolveNS(ctx, resolver, monitor.Hostname)
	default:
		err = fmt.Errorf("unsupported record type: %s", monitor.RecordType)
	}

	// Calculate resolution time
	resolutionTime := int(time.Since(startTime).Milliseconds())
	result.ResolutionTimeMS = resolutionTime

	if err != nil {
		result.Status = "down"
		result.ErrorMessage = err.Error()

		s.logger.Warn("DNS resolution failed",
			zap.String("hostname", monitor.Hostname),
			zap.String("record_type", monitor.RecordType),
			zap.Error(err))

		return result, nil
	}

	// Store resolved values as JSON string
	result.ResolvedValues = strings.Join(resolvedValues, ",")

	// Check if resolved values match expected values
	if monitor.ExpectedValues != "" {
		expectedList := strings.Split(monitor.ExpectedValues, ",")
		result.MatchesExpected = s.compareValues(resolvedValues, expectedList)

		if result.MatchesExpected {
			result.Status = "operational"
		} else {
			result.Status = "degraded"
		}
	} else {
		// No expected values specified, just check if we got any results
		if len(resolvedValues) > 0 {
			result.Status = "operational"
		} else {
			result.Status = "down"
		}
	}

	s.logger.Info("DNS check completed",
		zap.String("hostname", monitor.Hostname),
		zap.String("record_type", monitor.RecordType),
		zap.String("status", result.Status),
		zap.Int("resolution_time_ms", resolutionTime),
		zap.Int("record_count", len(resolvedValues)))

	return result, nil
}

// resolveA resolves A records (IPv4)
func (s *DNSMonitoringService) resolveA(ctx context.Context, resolver *net.Resolver, hostname string) ([]string, error) {
	ips, err := resolver.LookupIP(ctx, "ip4", hostname)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(ips))
	for _, ip := range ips {
		result = append(result, ip.String())
	}
	return result, nil
}

// resolveAAAA resolves AAAA records (IPv6)
func (s *DNSMonitoringService) resolveAAAA(ctx context.Context, resolver *net.Resolver, hostname string) ([]string, error) {
	ips, err := resolver.LookupIP(ctx, "ip6", hostname)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(ips))
	for _, ip := range ips {
		result = append(result, ip.String())
	}
	return result, nil
}

// resolveCNAME resolves CNAME records
func (s *DNSMonitoringService) resolveCNAME(ctx context.Context, resolver *net.Resolver, hostname string) ([]string, error) {
	cname, err := resolver.LookupCNAME(ctx, hostname)
	if err != nil {
		return nil, err
	}
	return []string{cname}, nil
}

// resolveMX resolves MX records
func (s *DNSMonitoringService) resolveMX(ctx context.Context, resolver *net.Resolver, hostname string) ([]string, error) {
	mxRecords, err := resolver.LookupMX(ctx, hostname)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(mxRecords))
	for _, mx := range mxRecords {
		// Format: "priority:host"
		result = append(result, fmt.Sprintf("%d:%s", mx.Pref, mx.Host))
	}
	return result, nil
}

// resolveTXT resolves TXT records
func (s *DNSMonitoringService) resolveTXT(ctx context.Context, resolver *net.Resolver, hostname string) ([]string, error) {
	txtRecords, err := resolver.LookupTXT(ctx, hostname)
	if err != nil {
		return nil, err
	}
	return txtRecords, nil
}

// resolveNS resolves NS records
func (s *DNSMonitoringService) resolveNS(ctx context.Context, resolver *net.Resolver, hostname string) ([]string, error) {
	nsRecords, err := resolver.LookupNS(ctx, hostname)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(nsRecords))
	for _, ns := range nsRecords {
		result = append(result, ns.Host)
	}
	return result, nil
}

// compareValues checks if resolved values match expected values
func (s *DNSMonitoringService) compareValues(resolved, expected []string) bool {
	// Trim whitespace from expected values
	for i := range expected {
		expected[i] = strings.TrimSpace(expected[i])
	}

	// Simple comparison: at least one expected value must be in resolved values
	for _, exp := range expected {
		for _, res := range resolved {
			if strings.EqualFold(exp, res) {
				return true
			}
		}
	}

	return false
}

// PerformCheck performs a DNS check and updates monitor status
func (s *DNSMonitoringService) PerformCheck(monitor *DNSMonitor) error {
	// Perform the DNS resolution
	result, err := s.ResolveDNS(monitor)
	if err != nil {
		s.logger.Error("Failed to perform DNS check", zap.Error(err))
		return err
	}

	// Save result to database
	if err := s.db.Create(result).Error; err != nil {
		s.logger.Error("Failed to save DNS check result", zap.Error(err))
		return err
	}

	// Update monitor status
	now := time.Now()
	monitor.LastChecked = &now
	monitor.AvgResolutionTime = float64(result.ResolutionTimeMS)

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
			s.logger.Warn("DNS monitor threshold reached",
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
func (s *DNSMonitoringService) GetCheckResults(monitorID uint, limit int) ([]DNSCheckResult, error) {
	var results []DNSCheckResult
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

// GetMonitorStats calculates statistics for a DNS monitor
func (s *DNSMonitoringService) GetMonitorStats(monitorID uint, timeRange string) (map[string]interface{}, error) {
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
	var results []DNSCheckResult
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
	matchingChecks := 0
	var totalResolutionTime int64

	for _, result := range results {
		if result.Status == "operational" {
			successfulChecks++
		}
		if result.MatchesExpected {
			matchingChecks++
		}
		totalResolutionTime += int64(result.ResolutionTimeMS)
	}

	uptimePercentage := float64(successfulChecks) / float64(totalChecks) * 100
	accuracyPercentage := float64(matchingChecks) / float64(totalChecks) * 100
	avgResolutionTime := float64(totalResolutionTime) / float64(totalChecks)

	stats := map[string]interface{}{
		"total_checks":          totalChecks,
		"successful_checks":     successfulChecks,
		"failed_checks":         totalChecks - successfulChecks,
		"matching_checks":       matchingChecks,
		"uptime_percentage":     uptimePercentage,
		"accuracy_percentage":   accuracyPercentage,
		"avg_resolution_time_ms": avgResolutionTime,
		"time_range":            timeRange,
	}

	return stats, nil
}

// TestDNS tests DNS resolution (one-time check)
func (s *DNSMonitoringService) TestDNS(hostname, recordType string, dnsServer string, timeout int) ([]string, int, error) {
	monitor := &DNSMonitor{
		Hostname:   hostname,
		RecordType: recordType,
		DNSServer:  dnsServer,
		Timeout:    timeout,
	}

	result, err := s.ResolveDNS(monitor)
	if err != nil {
		return nil, 0, err
	}

	values := strings.Split(result.ResolvedValues, ",")
	return values, result.ResolutionTimeMS, nil
}

// GetDNSPropagation checks DNS propagation across multiple DNS servers
func (s *DNSMonitoringService) GetDNSPropagation(hostname, recordType string) (map[string]interface{}, error) {
	// List of public DNS servers to check
	dnsServers := map[string]string{
		"Google DNS":        "8.8.8.8",
		"Cloudflare DNS":    "1.1.1.1",
		"OpenDNS":           "208.67.222.222",
		"Quad9":             "9.9.9.9",
		"Verisign":          "64.6.64.6",
	}

	results := make(map[string]interface{})
	allMatch := true
	var firstResult []string

	for name, server := range dnsServers {
		values, resolutionTime, err := s.TestDNS(hostname, recordType, server, 5)

		serverResult := map[string]interface{}{
			"values":           values,
			"resolution_time":  resolutionTime,
			"error":            nil,
		}

		if err != nil {
			serverResult["error"] = err.Error()
			allMatch = false
		} else {
			if firstResult == nil {
				firstResult = values
			} else {
				// Check if this result matches the first one
				if !s.slicesEqual(values, firstResult) {
					allMatch = false
				}
			}
		}

		results[name] = serverResult
	}

	results["propagation_complete"] = allMatch
	results["checked_at"] = time.Now()

	return results, nil
}

// slicesEqual checks if two string slices contain the same elements
func (s *DNSMonitoringService) slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for _, valA := range a {
		found := false
		for _, valB := range b {
			if valA == valB {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
