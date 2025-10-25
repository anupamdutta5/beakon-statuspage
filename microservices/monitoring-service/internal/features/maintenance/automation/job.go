// Package automation provides automated maintenance window management.
package automation

import (
	"time"

	"github.com/anupamdutta5/monitoring-service/internal/services"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MaintenanceWindowJob struct {
	db                  *gorm.DB
	maintenanceService  *services.MaintenanceManagementService
	logger              *zap.Logger
	interval            time.Duration
	stopChan            chan struct{}
}

// NewMaintenanceWindowJob creates a new maintenance window job.
func NewMaintenanceWindowJob(db *gorm.DB, logger *zap.Logger, checkInterval time.Duration) *MaintenanceWindowJob {
	return &MaintenanceWindowJob{
		db:                 db,
		maintenanceService: services.NewMaintenanceManagementService(db, logger),
		logger:             logger,
		interval:           checkInterval,
		stopChan:           make(chan struct{}),
	}
}

// Start begins the maintenance window job.
func (j *MaintenanceWindowJob) Start() {
	j.logger.Info("Starting Maintenance Window Job", zap.Duration("interval", j.interval))

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			j.processMaintenanceWindows()
		case <-j.stopChan:
			j.logger.Info("Stopping Maintenance Window Job")
			return
		}
	}
}

// Stop stops the maintenance window job.
func (j *MaintenanceWindowJob) Stop() {
	close(j.stopChan)
}

// processMaintenanceWindows checks and auto-starts/completes maintenance windows.
func (j *MaintenanceWindowJob) processMaintenanceWindows() {
	j.logger.Debug("Processing maintenance windows")

	// Send 60-minute reminders for upcoming maintenance
	err := j.maintenanceService.SendMaintenanceReminders()
	if err != nil {
		j.logger.Error("Failed to send maintenance reminders", zap.Error(err))
	}

	// Auto-start scheduled maintenance windows
	err = j.maintenanceService.AutoStartMaintenanceWindows()
	if err != nil {
		j.logger.Error("Failed to auto-start maintenance windows", zap.Error(err))
	}

	// Auto-complete expired maintenance windows
	err = j.maintenanceService.AutoCompleteMaintenanceWindows()
	if err != nil {
		j.logger.Error("Failed to auto-complete maintenance windows", zap.Error(err))
	}
}
