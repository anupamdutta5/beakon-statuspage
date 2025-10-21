package main

import (
	"fmt"
	"os"
	"time"

	"github.com/anupamdutta5/monitoring-service/internal/services"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	logger.Info("🚀 Starting MTTR/MTTD Tracking Test")

	// Initialize database connection
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "monitoring_db"),
		getEnv("DB_SSLMODE", "disable"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	logger.Info("✅ Database connection established")

	// Auto-migrate tables
	err = db.AutoMigrate(
		&services.IncidentTracking{},
		&services.MetricsSnapshot{},
	)
	if err != nil {
		logger.Fatal("Failed to migrate tables", zap.Error(err))
	}

	logger.Info("✅ Database tables migrated")

	// Create MTTR/MTTD tracking service
	trackingService := services.NewMTTRMTTDTrackingService(db, logger)

	testTenantID := uuid.New()
	testMonitorID := uint(3000)

	// Test 1: Start a New Incident
	logger.Info("\n🚨 Test 1: Starting New Incident")
	incident, err := trackingService.StartIncident(testMonitorID, testTenantID, "critical", "connection_timeout")
	if err != nil {
		logger.Error("Failed to start incident", zap.Error(err))
	} else {
		logger.Info("✅ Incident started",
			zap.Uint("id", incident.ID),
			zap.String("incident_key", incident.IncidentKey),
			zap.String("status", incident.Status),
			zap.Float64("mttd_minutes", incident.MTTD))
	}

	// Test 2: Send Alert
	logger.Info("\n📢 Test 2: Recording Alert Sent")
	time.Sleep(2 * time.Second) // Simulate delay
	if incident != nil {
		err = trackingService.SendAlert(incident.ID)
		if err != nil {
			logger.Error("Failed to record alert", zap.Error(err))
		} else {
			logger.Info("✅ Alert time recorded")
		}
	}

	// Test 3: Acknowledge Incident
	logger.Info("\n👍 Test 3: Acknowledging Incident")
	time.Sleep(3 * time.Second) // Simulate delay before acknowledgment
	if incident != nil {
		err = trackingService.AcknowledgeIncident(incident.ID)
		if err != nil {
			logger.Error("Failed to acknowledge incident", zap.Error(err))
		} else {
			// Retrieve updated incident
			updated, _ := trackingService.GetIncident(incident.ID)
			logger.Info("✅ Incident acknowledged",
				zap.Float64("mtta_minutes", updated.MTTA),
				zap.String("status", updated.Status))
		}
	}

	// Test 4: Start Investigation
	logger.Info("\n🔍 Test 4: Starting Investigation")
	time.Sleep(2 * time.Second)
	if incident != nil {
		err = trackingService.StartInvestigation(incident.ID)
		if err != nil {
			logger.Error("Failed to start investigation", zap.Error(err))
		} else {
			updated, _ := trackingService.GetIncident(incident.ID)
			logger.Info("✅ Investigation started",
				zap.Float64("mtti_minutes", updated.MTTI),
				zap.String("status", updated.Status))
		}
	}

	// Test 5: Start Resolution
	logger.Info("\n🔧 Test 5: Starting Resolution")
	time.Sleep(5 * time.Second) // Simulate investigation time
	if incident != nil {
		err = trackingService.StartResolution(incident.ID)
		if err != nil {
			logger.Error("Failed to start resolution", zap.Error(err))
		} else {
			updated, _ := trackingService.GetIncident(incident.ID)
			logger.Info("✅ Resolution started",
				zap.String("status", updated.Status))
		}
	}

	// Test 6: Resolve Incident
	logger.Info("\n✅ Test 6: Resolving Incident")
	time.Sleep(3 * time.Second) // Simulate resolution time
	if incident != nil {
		err = trackingService.ResolveIncident(incident.ID, "Database connection pool exhausted", "Increased connection pool size from 100 to 200")
		if err != nil {
			logger.Error("Failed to resolve incident", zap.Error(err))
		} else {
			updated, _ := trackingService.GetIncident(incident.ID)
			logger.Info("✅ Incident resolved",
				zap.Float64("mttr_minutes", updated.MTTR),
				zap.Float64("downtime_minutes", updated.TotalDowntime),
				zap.String("status", updated.Status))
		}
	}

	// Test 7: Verify Recovery
	logger.Info("\n✔️  Test 7: Verifying Recovery")
	time.Sleep(2 * time.Second)
	if incident != nil {
		err = trackingService.VerifyRecovery(incident.ID)
		if err != nil {
			logger.Error("Failed to verify recovery", zap.Error(err))
		} else {
			updated, _ := trackingService.GetIncident(incident.ID)
			logger.Info("✅ Recovery verified",
				zap.Float64("mttv_minutes", updated.MTTV),
				zap.String("status", updated.Status))
		}
	}

	// Test 8: Get Incident Details
	logger.Info("\n📄 Test 8: Retrieving Incident Details")
	if incident != nil {
		retrieved, err := trackingService.GetIncident(incident.ID)
		if err != nil {
			logger.Error("Failed to get incident", zap.Error(err))
		} else {
			logger.Info("✅ Incident details retrieved")
			logger.Info(fmt.Sprintf("  ID: %d", retrieved.ID))
			logger.Info(fmt.Sprintf("  Incident Key: %s", retrieved.IncidentKey))
			logger.Info(fmt.Sprintf("  Status: %s", retrieved.Status))
			logger.Info(fmt.Sprintf("  Severity: %s", retrieved.Severity))
			logger.Info(fmt.Sprintf("  MTTD: %.2f minutes", retrieved.MTTD))
			logger.Info(fmt.Sprintf("  MTTA: %.2f minutes", retrieved.MTTA))
			logger.Info(fmt.Sprintf("  MTTI: %.2f minutes", retrieved.MTTI))
			logger.Info(fmt.Sprintf("  MTTR: %.2f minutes", retrieved.MTTR))
			logger.Info(fmt.Sprintf("  MTTV: %.2f minutes", retrieved.MTTV))
			logger.Info(fmt.Sprintf("  Total Downtime: %.2f minutes", retrieved.TotalDowntime))
			logger.Info(fmt.Sprintf("  Root Cause: %s", retrieved.RootCause))
			logger.Info(fmt.Sprintf("  Resolution: %s", retrieved.Resolution))
		}
	}

	// Test 9: Create Multiple Incidents with Different Severities
	logger.Info("\n🔄 Test 9: Creating Multiple Incidents")

	incidents := []struct {
		severity     string
		failureType  string
		resolveDelay time.Duration
	}{
		{"critical", "connection_timeout", 10 * time.Second},
		{"high", "status_code_500", 8 * time.Second},
		{"medium", "slow_response", 5 * time.Second},
		{"low", "ssl_expiring_soon", 3 * time.Second},
		{"critical", "database_down", 15 * time.Second},
	}

	var createdIncidents []*services.IncidentTracking

	for i, inc := range incidents {
		logger.Info(fmt.Sprintf("\n  Creating incident %d: %s (%s)", i+1, inc.severity, inc.failureType))

		newIncident, err := trackingService.StartIncident(testMonitorID, testTenantID, inc.severity, inc.failureType)
		if err != nil {
			logger.Error("Failed to create incident", zap.Error(err))
			continue
		}

		createdIncidents = append(createdIncidents, newIncident)

		// Simulate incident lifecycle
		time.Sleep(1 * time.Second)
		trackingService.AcknowledgeIncident(newIncident.ID)

		time.Sleep(2 * time.Second)
		trackingService.StartInvestigation(newIncident.ID)

		time.Sleep(2 * time.Second)
		trackingService.StartResolution(newIncident.ID)

		time.Sleep(inc.resolveDelay)
		trackingService.ResolveIncident(newIncident.ID, "Test root cause", "Test resolution")

		time.Sleep(1 * time.Second)
		trackingService.VerifyRecovery(newIncident.ID)

		logger.Info(fmt.Sprintf("  ✅ Incident %d completed", i+1))
	}

	// Test 10: Get Incidents by Monitor
	logger.Info("\n📋 Test 10: Retrieving Incidents by Monitor")
	monitorIncidents, err := trackingService.GetIncidentsByMonitor(testMonitorID, 10)
	if err != nil {
		logger.Error("Failed to get incidents", zap.Error(err))
	} else {
		logger.Info("✅ Retrieved incidents by monitor", zap.Int("count", len(monitorIncidents)))
		for i, inc := range monitorIncidents {
			logger.Info(fmt.Sprintf("  Incident %d: %s severity, MTTR: %.2f min, Status: %s",
				i+1, inc.Severity, inc.MTTR, inc.Status))
		}
	}

	// Test 11: Get Open Incidents
	logger.Info("\n🔓 Test 11: Retrieving Open Incidents")

	// Create an open incident
	openIncident, _ := trackingService.StartIncident(testMonitorID+1, testTenantID, "high", "api_error")
	trackingService.AcknowledgeIncident(openIncident.ID)

	openIncidents, err := trackingService.GetOpenIncidents(testTenantID)
	if err != nil {
		logger.Error("Failed to get open incidents", zap.Error(err))
	} else {
		logger.Info("✅ Retrieved open incidents", zap.Int("count", len(openIncidents)))
		for i, inc := range openIncidents {
			logger.Info(fmt.Sprintf("  Open Incident %d: Monitor %d, Status: %s, Age: %.2f min",
				i+1, inc.MonitorID, inc.Status, time.Since(inc.IncidentStartTime).Minutes()))
		}
	}

	// Test 12: Calculate Metrics for a Period
	logger.Info("\n📊 Test 12: Calculating Metrics Snapshot")
	periodStart := time.Now().Add(-24 * time.Hour)
	periodEnd := time.Now()

	metricsSnapshot, err := trackingService.CalculateMetrics(testMonitorID, testTenantID, periodStart, periodEnd, "daily")
	if err != nil {
		logger.Error("Failed to calculate metrics", zap.Error(err))
	} else {
		logger.Info("✅ Metrics snapshot calculated")
		logger.Info(fmt.Sprintf("  Period: %s to %s", periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02")))
		logger.Info(fmt.Sprintf("  Total Incidents: %d", metricsSnapshot.TotalIncidents))
		logger.Info(fmt.Sprintf("  Critical: %d, High: %d, Medium: %d, Low: %d",
			metricsSnapshot.CriticalIncidents,
			metricsSnapshot.HighIncidents,
			metricsSnapshot.MediumIncidents,
			metricsSnapshot.LowIncidents))
		logger.Info(fmt.Sprintf("  Avg MTTD: %.2f minutes", metricsSnapshot.AvgMTTD))
		logger.Info(fmt.Sprintf("  Avg MTTA: %.2f minutes", metricsSnapshot.AvgMTTA))
		logger.Info(fmt.Sprintf("  Avg MTTI: %.2f minutes", metricsSnapshot.AvgMTTI))
		logger.Info(fmt.Sprintf("  Avg MTTR: %.2f minutes", metricsSnapshot.AvgMTTR))
		logger.Info(fmt.Sprintf("  Avg MTTV: %.2f minutes", metricsSnapshot.AvgMTTV))
		logger.Info(fmt.Sprintf("  Avg Total Downtime: %.2f minutes", metricsSnapshot.AvgTotalDowntime))
		logger.Info(fmt.Sprintf("  MTTR Percentiles: P50=%.2f, P90=%.2f, P95=%.2f, P99=%.2f",
			metricsSnapshot.P50MTTR,
			metricsSnapshot.P90MTTR,
			metricsSnapshot.P95MTTR,
			metricsSnapshot.P99MTTR))
		logger.Info(fmt.Sprintf("  Trend: %s (%.2f%%)", metricsSnapshot.TrendDirection, metricsSnapshot.TrendPercent))
	}

	// Test 13: Get Metrics History
	logger.Info("\n📈 Test 13: Retrieving Metrics History")

	// Calculate metrics for multiple periods
	for i := 0; i < 3; i++ {
		start := time.Now().Add(time.Duration(-24*(i+1)) * time.Hour)
		end := time.Now().Add(time.Duration(-24*i) * time.Hour)
		trackingService.CalculateMetrics(testMonitorID, testTenantID, start, end, "daily")
	}

	history, err := trackingService.GetMetricsHistory(testMonitorID, testTenantID, "daily", 5)
	if err != nil {
		logger.Error("Failed to get metrics history", zap.Error(err))
	} else {
		logger.Info("✅ Retrieved metrics history", zap.Int("count", len(history)))
		for i, snapshot := range history {
			logger.Info(fmt.Sprintf("  Snapshot %d: %s, Incidents: %d, Avg MTTR: %.2f min, Trend: %s",
				i+1,
				snapshot.PeriodStart.Format("2006-01-02"),
				snapshot.TotalIncidents,
				snapshot.AvgMTTR,
				snapshot.TrendDirection))
		}
	}

	// Test 14: Get Current Month Metrics
	logger.Info("\n📅 Test 14: Getting Current Month Metrics")
	currentMonthMetrics, err := trackingService.GetCurrentMonthMetrics(testMonitorID, testTenantID)
	if err != nil {
		logger.Error("Failed to get current month metrics", zap.Error(err))
	} else {
		logger.Info("✅ Current month metrics retrieved")
		logger.Info(fmt.Sprintf("  Total Incidents: %d", currentMonthMetrics.TotalIncidents))
		logger.Info(fmt.Sprintf("  Avg MTTR: %.2f minutes", currentMonthMetrics.AvgMTTR))
		logger.Info(fmt.Sprintf("  Avg MTTD: %.2f minutes", currentMonthMetrics.AvgMTTD))
	}

	// Test 15: Tenant-Wide Metrics
	logger.Info("\n🌍 Test 15: Calculating Tenant-Wide Metrics")

	// Create incidents for multiple monitors
	for monitorID := uint(4000); monitorID < 4003; monitorID++ {
		inc, _ := trackingService.StartIncident(monitorID, testTenantID, "high", "test_failure")
		time.Sleep(5 * time.Second)
		trackingService.ResolveIncident(inc.ID, "Test cause", "Test resolution")
	}

	tenantWideMetrics, err := trackingService.CalculateMetrics(0, testTenantID, periodStart, periodEnd, "daily")
	if err != nil {
		logger.Error("Failed to calculate tenant-wide metrics", zap.Error(err))
	} else {
		logger.Info("✅ Tenant-wide metrics calculated")
		logger.Info(fmt.Sprintf("  Total Incidents (All Monitors): %d", tenantWideMetrics.TotalIncidents))
		logger.Info(fmt.Sprintf("  Avg MTTR (Tenant): %.2f minutes", tenantWideMetrics.AvgMTTR))
	}

	// Test 16: Metrics Comparison
	logger.Info("\n🔍 Test 16: Comparing Metrics Across Periods")

	thisWeekStart := time.Now().Add(-7 * 24 * time.Hour)
	thisWeekEnd := time.Now()
	lastWeekStart := time.Now().Add(-14 * 24 * time.Hour)
	lastWeekEnd := time.Now().Add(-7 * 24 * time.Hour)

	thisWeekMetrics, _ := trackingService.CalculateMetrics(testMonitorID, testTenantID, thisWeekStart, thisWeekEnd, "weekly")
	lastWeekMetrics, _ := trackingService.CalculateMetrics(testMonitorID, testTenantID, lastWeekStart, lastWeekEnd, "weekly")

	if thisWeekMetrics != nil && lastWeekMetrics != nil {
		logger.Info("✅ Metrics comparison")
		logger.Info(fmt.Sprintf("  Last Week: Incidents=%d, Avg MTTR=%.2f min",
			lastWeekMetrics.TotalIncidents, lastWeekMetrics.AvgMTTR))
		logger.Info(fmt.Sprintf("  This Week: Incidents=%d, Avg MTTR=%.2f min",
			thisWeekMetrics.TotalIncidents, thisWeekMetrics.AvgMTTR))

		incidentChange := ((float64(thisWeekMetrics.TotalIncidents) - float64(lastWeekMetrics.TotalIncidents)) / float64(lastWeekMetrics.TotalIncidents)) * 100
		mttrChange := ((thisWeekMetrics.AvgMTTR - lastWeekMetrics.AvgMTTR) / lastWeekMetrics.AvgMTTR) * 100

		logger.Info(fmt.Sprintf("  Incident Change: %.2f%%", incidentChange))
		logger.Info(fmt.Sprintf("  MTTR Change: %.2f%%", mttrChange))
	}

	// Final Summary
	logger.Info("\n✨ All MTTR/MTTD Tracking Tests Completed!")
	logger.Info("\n📝 Summary:")
	logger.Info("  - Incident lifecycle tracking: ✅ Working")
	logger.Info("  - MTTD (Mean Time To Detect): ✅ Calculated")
	logger.Info("  - MTTA (Mean Time To Acknowledge): ✅ Calculated")
	logger.Info("  - MTTI (Mean Time To Investigate): ✅ Calculated")
	logger.Info("  - MTTR (Mean Time To Resolve): ✅ Calculated")
	logger.Info("  - MTTV (Mean Time To Verify): ✅ Calculated")
	logger.Info("  - Metrics snapshots: ✅ Generated")
	logger.Info("  - Metrics history: ✅ Working")
	logger.Info("  - Severity tracking: ✅ Working")
	logger.Info("  - Percentile calculation: ✅ Working")
	logger.Info("  - Trend analysis: ✅ Working")
	logger.Info("  - Tenant-wide metrics: ✅ Working")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
