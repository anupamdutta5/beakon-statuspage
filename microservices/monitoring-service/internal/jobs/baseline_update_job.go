// Package jobs provides background job implementations for anomaly detection.
package jobs

import (
	"fmt"
	"time"

	"github.com/anupamdutta5/monitoring-service/internal/models"
	"github.com/anupamdutta5/monitoring-service/internal/services"
	"gorm.io/gorm"
)

// BaselineUpdateJob handles periodic recalculation of anomaly baselines.
type BaselineUpdateJob struct {
	db                 *gorm.DB
	baselineCalculator *services.BaselineCalculator
	interval           time.Duration
}

// NewBaselineUpdateJob creates a new BaselineUpdateJob instance.
func NewBaselineUpdateJob(db *gorm.DB, baselineCalculator *services.BaselineCalculator, interval time.Duration) *BaselineUpdateJob {
	return &BaselineUpdateJob{
		db:                 db,
		baselineCalculator: baselineCalculator,
		interval:           interval,
	}
}

// Start begins the baseline update job loop.
func (buj *BaselineUpdateJob) Start() {
	ticker := time.NewTicker(buj.interval)
	defer ticker.Stop()

	fmt.Printf("[BaselineUpdateJob] Started with interval: %v\n", buj.interval)

	// Run immediately on start
	buj.updateBaselines()

	for range ticker.C {
		buj.updateBaselines()
	}
}

// updateBaselines recalculates baselines for all tenants.
func (buj *BaselineUpdateJob) updateBaselines() {
	fmt.Printf("[BaselineUpdateJob] Starting baseline updates at %v\n", time.Now())

	// Get all unique tenant IDs from uptime checks
	var tenantIDs []uint
	if err := buj.db.Model(&models.UptimeCheck{}).
		Distinct("tenant_id").
		Pluck("tenant_id", &tenantIDs).Error; err != nil {
		fmt.Printf("[BaselineUpdateJob] Error fetching tenant IDs: %v\n", err)
		return
	}

	fmt.Printf("[BaselineUpdateJob] Found %d tenants to process\n", len(tenantIDs))

	successCount := 0
	errorCount := 0

	for _, tenantID := range tenantIDs {
		if err := buj.updateTenantBaselines(tenantID); err != nil {
			fmt.Printf("[BaselineUpdateJob] Error updating baselines for tenant %s: %v\n", tenantID, err)
			errorCount++
		} else {
			successCount++
		}
	}

	fmt.Printf("[BaselineUpdateJob] Completed: %d tenants succeeded, %d failed\n", successCount, errorCount)
}

// updateTenantBaselines recalculates baselines for a specific tenant.
func (buj *BaselineUpdateJob) updateTenantBaselines(tenantID uint) error {
	// Get all uptime checks for this tenant
	var monitors []models.UptimeCheck
	if err := buj.db.Where("tenant_id = ? AND is_active = ?", tenantID, true).Find(&monitors).Error; err != nil {
		return fmt.Errorf("failed to fetch uptime checks: %w", err)
	}

	if len(monitors) == 0 {
		return nil // No monitors, skip
	}

	// Calculate baselines for each monitor
	for _, monitor := range monitors {
		if err := buj.updateMonitorBaselines(tenantID, monitor.ID); err != nil {
			// Log but continue with other monitors
			fmt.Printf("[BaselineUpdateJob] Error updating monitor %d: %v\n", monitor.ID, err)
		}
	}

	return nil
}

// updateMonitorBaselines recalculates baselines for a specific monitor.
func (buj *BaselineUpdateJob) updateMonitorBaselines(tenantID uint, monitorID uint) error {
	// Check if we have enough data
	var metricCount int64
	if err := buj.db.Model(&models.MetricSnapshot{}).
		Where("tenant_id = ? AND monitor_id = ?", tenantID, monitorID).
		Count(&metricCount).Error; err != nil {
		return fmt.Errorf("failed to count metrics: %w", err)
	}

	if metricCount < 50 {
		// Not enough data yet
		return nil
	}

	// Calculate baselines using the baseline calculator
	if err := buj.baselineCalculator.CalculateMonitorBaselines(tenantID, monitorID); err != nil {
		return fmt.Errorf("failed to calculate baselines: %w", err)
	}

	return nil
}

// CleanupExpiredBaselines removes expired baseline records.
func (buj *BaselineUpdateJob) CleanupExpiredBaselines() error {
	return buj.baselineCalculator.CleanupExpiredBaselines()
}
