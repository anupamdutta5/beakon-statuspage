package services

import (
	"fmt"
	"net/http"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
)

type MonitoringService struct{}

func NewMonitoringService() *MonitoringService {
	return &MonitoringService{}
}

// Start kicks off the monitoring background worker
func (s *MonitoringService) Start() {
	logger.Info("Starting monitoring service...")
	go s.runChecks()
}

// runChecks is the main loop for the monitoring worker
func (s *MonitoringService) runChecks() {
	// Run checks immediately on start, then on a ticker
	s.performAllChecks()

	ticker := time.NewTicker(1 * time.Minute) // Check every minute for which monitors to run
	defer ticker.Stop()

	for range ticker.C {
		s.performAllChecks()
	}
}

// performAllChecks fetches all monitors and checks them if it's time
func (s *MonitoringService) performAllChecks() {
	var monitors []models.Monitor
	if err := database.DB.Find(&monitors).Error; err != nil {
		logger.Error("Failed to fetch monitors", zap.Error(err))
		return
	}

	for _, monitor := range monitors {
		// Check if it's time to run the check for this monitor
		if time.Since(monitor.LastCheckAt) >= (time.Duration(monitor.Interval) * time.Second) {
			go s.checkMonitor(monitor)
		}
	}
}

// checkMonitor performs a single check for a given monitor
func (s *MonitoringService) checkMonitor(monitor models.Monitor) {
	var heartbeat models.Heartbeat
	heartbeat.MonitorID = monitor.ID
	heartbeat.Timestamp = time.Now()

	startTime := time.Now()

	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(monitor.URL)

	latency := time.Since(startTime).Milliseconds()
	heartbeat.Latency = latency

	if err != nil {
		heartbeat.Status = "down"
		heartbeat.Message = fmt.Sprintf("Request failed: %v", err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == monitor.ExpectedStatus {
			heartbeat.Status = "up"
			heartbeat.Message = fmt.Sprintf("OK (%d)", resp.StatusCode)
		} else {
			heartbeat.Status = "down"
			heartbeat.Message = fmt.Sprintf("Unexpected status code: %d", resp.StatusCode)
		}
	}

	// Save the heartbeat
	if err := database.DB.Create(&heartbeat).Error; err != nil {
		logger.Error("Failed to save heartbeat", zap.Error(err), zap.Uint("monitor_id", monitor.ID))
	}

	// Update the monitor's last check time and result
	monitor.LastCheckAt = heartbeat.Timestamp
	monitor.LastResult = heartbeat.Status
	if err := database.DB.Save(&monitor).Error; err != nil {
		logger.Error("Failed to update monitor", zap.Error(err), zap.Uint("monitor_id", monitor.ID))
	}

	// Optional: Update the associated service's status
	if monitor.LastResult != heartbeat.Status { // If status changed
		var service models.Service
		if err := database.DB.First(&service, monitor.ServiceID).Error; err == nil {
			service.Status = heartbeat.Status
			if err := database.DB.Save(&service).Error; err != nil {
				logger.Error("Failed to update service status", zap.Error(err), zap.Uint("service_id", monitor.ServiceID))
			}
		}
	}

	logger.Info("Checked monitor", zap.String("url", monitor.URL), zap.String("status", heartbeat.Status), zap.Int64("latency_ms", heartbeat.Latency))
}
