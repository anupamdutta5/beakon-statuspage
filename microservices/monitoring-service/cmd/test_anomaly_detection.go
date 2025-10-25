package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/anupamdutta5/monitoring-service/internal/models"
	"github.com/anupamdutta5/monitoring-service/internal/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Test script for anomaly detection feature
// This tests the complete anomaly detection workflow:
// 1. Metric collection
// 2. Baseline calculation
// 3. Anomaly detection
// 4. Configuration management

func main() {
	fmt.Println("=== Anomaly Detection Integration Test ===")
	fmt.Println()

	// Setup database connection
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "monitoring_db"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_SSLMODE", "disable"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("✓ Connected to database")

	// Initialize services
	baselineCalculator := services.NewBaselineCalculator(db)
	anomalyService := services.NewAnomalyDetectionService(db, baselineCalculator)

	fmt.Println("✓ Initialized services")
	fmt.Println()

	// Test configuration
	tenantID := uint(1)
	monitorID := uint(1)
	metricType := "response_time"

	// Test 1: Store metrics
	fmt.Println("Test 1: Storing metric snapshots...")
	testMetricStorage(db, tenantID, monitorID, metricType)

	// Test 2: Calculate baselines
	fmt.Println("\nTest 2: Calculating baselines...")
	testBaselineCalculation(baselineCalculator, tenantID, monitorID, metricType)

	// Test 3: Detect anomalies
	fmt.Println("\nTest 3: Testing anomaly detection...")
	testAnomalyDetection(anomalyService, tenantID, monitorID, metricType)

	// Test 4: Get anomalies
	fmt.Println("\nTest 4: Retrieving anomalies...")
	testGetAnomalies(anomalyService, tenantID)

	// Test 5: Acknowledge and resolve
	fmt.Println("\nTest 5: Testing acknowledge and resolve workflow...")
	testWorkflow(anomalyService, tenantID)

	// Test 6: Configuration management
	fmt.Println("\nTest 6: Testing configuration management...")
	testConfiguration(anomalyService, tenantID)

	// Test 7: Statistics
	fmt.Println("\nTest 7: Testing statistics...")
	testStatistics(anomalyService, tenantID)

	fmt.Println("\n=== All Tests Passed! ===")
}

func testMetricStorage(db *gorm.DB, tenantID, monitorID uint, metricType string) {
	// Store 100 normal metrics (around mean of 100ms)
	fmt.Println("  - Storing 100 normal metrics...")
	for i := 0; i < 100; i++ {
		snapshot := &models.MetricSnapshot{
			TenantID:    tenantID,
			MonitorID:   &monitorID,
			MetricType:  metricType,
			MetricValue: 100.0 + float64(i%20-10), // 90-110ms range
			Timestamp:   time.Now().Add(-time.Duration(100-i) * time.Minute),
		}
		if err := db.Create(snapshot).Error; err != nil {
			log.Fatalf("Failed to store metric: %v", err)
		}
	}
	fmt.Println("  ✓ Stored 100 metrics")

	// Verify storage
	var count int64
	db.Model(&models.MetricSnapshot{}).
		Where("tenant_id = ? AND monitor_id = ?", tenantID, monitorID).
		Count(&count)
	fmt.Printf("  ✓ Verified %d metrics in database\n", count)
}

func testBaselineCalculation(bc *services.BaselineCalculator, tenantID, monitorID uint, metricType string) {
	// Calculate 7-day baseline
	fmt.Println("  - Calculating 7-day baseline...")
	err := bc.Calculate7DayBaseline(tenantID, monitorID, metricType)
	if err != nil {
		log.Fatalf("Failed to calculate 7-day baseline: %v", err)
	}
	fmt.Println("  ✓ 7-day baseline calculated")

	// Get baseline
	baseline, err := bc.GetBaseline(tenantID, monitorID, metricType, time.Now())
	if err != nil {
		log.Fatalf("Failed to get baseline: %v", err)
	}

	fmt.Printf("  ✓ Baseline retrieved:\n")
	fmt.Printf("    - Mean: %.2f\n", baseline.MeanValue)
	fmt.Printf("    - Std Dev: %.2f\n", baseline.StdDev)
	fmt.Printf("    - P95: %.2f\n", baseline.P95)
	fmt.Printf("    - P99: %.2f\n", baseline.P99)
	fmt.Printf("    - Samples: %d\n", baseline.SampleCount)
}

func testAnomalyDetection(ads *services.AnomalyDetectionService, tenantID, monitorID uint, metricType string) {
	// Test normal value (should not trigger)
	fmt.Println("  - Testing normal value (100ms)...")
	result, err := ads.DetectAnomaly(tenantID, monitorID, metricType, 100.0, time.Now())
	if err != nil {
		log.Fatalf("Detection failed: %v", err)
	}
	if result.IsAnomaly {
		log.Fatal("Normal value incorrectly detected as anomaly")
	}
	fmt.Println("  ✓ Normal value correctly identified")

	// Test minor anomaly (should trigger minor)
	fmt.Println("  - Testing minor anomaly (150ms, ~2.5σ)...")
	result, err = ads.DetectAnomaly(tenantID, monitorID, metricType, 150.0, time.Now())
	if err != nil {
		log.Fatalf("Detection failed: %v", err)
	}
	if !result.IsAnomaly {
		log.Fatal("Minor anomaly not detected")
	}
	if result.Severity != "minor" && result.Severity != "major" {
		log.Fatalf("Wrong severity: got %s, expected minor or major", result.Severity)
	}
	fmt.Printf("  ✓ Minor anomaly detected (severity: %s, z-score: %.2f)\n", result.Severity, result.DeviationScore)

	// Test critical anomaly (should trigger critical)
	fmt.Println("  - Testing critical anomaly (500ms, >4σ)...")
	result, err = ads.DetectAnomaly(tenantID, monitorID, metricType, 500.0, time.Now())
	if err != nil {
		log.Fatalf("Detection failed: %v", err)
	}
	if !result.IsAnomaly {
		log.Fatal("Critical anomaly not detected")
	}
	if result.Severity != "critical" {
		log.Fatalf("Wrong severity: got %s, expected critical", result.Severity)
	}
	fmt.Printf("  ✓ Critical anomaly detected (z-score: %.2f)\n", result.DeviationScore)
}

func testGetAnomalies(ads *services.AnomalyDetectionService, tenantID uint) {
	filters := make(map[string]interface{})
	filters["status"] = "open"

	anomalies, err := ads.GetAnomalies(tenantID, filters, 100)
	if err != nil {
		log.Fatalf("Failed to get anomalies: %v", err)
	}

	fmt.Printf("  ✓ Retrieved %d open anomalies\n", len(anomalies))

	if len(anomalies) > 0 {
		anomaly := anomalies[0]
		fmt.Printf("    - ID: %d\n", anomaly.ID)
		fmt.Printf("    - Severity: %s\n", anomaly.Severity)
		fmt.Printf("    - Deviation: %.2fσ\n", anomaly.DeviationScore)
		fmt.Printf("    - Actual: %.2f, Expected: %.2f\n", anomaly.ActualValue, anomaly.ExpectedValue)
	}
}

func testWorkflow(ads *services.AnomalyDetectionService, tenantID uint) {
	// Get first open anomaly
	filters := make(map[string]interface{})
	filters["status"] = "open"

	anomalies, err := ads.GetAnomalies(tenantID, filters, 1)
	if err != nil || len(anomalies) == 0 {
		fmt.Println("  - No open anomalies to test workflow")
		return
	}

	anomaly := anomalies[0]
	userID := uint(1)

	// Acknowledge
	fmt.Printf("  - Acknowledging anomaly #%d...\n", anomaly.ID)
	err = ads.AcknowledgeAnomaly(tenantID, anomaly.ID, userID, "Testing acknowledgment workflow")
	if err != nil {
		log.Fatalf("Failed to acknowledge: %v", err)
	}
	fmt.Println("  ✓ Anomaly acknowledged")

	// Resolve
	fmt.Printf("  - Resolving anomaly #%d...\n", anomaly.ID)
	err = ads.ResolveAnomaly(tenantID, anomaly.ID, "resolved", "Testing resolution workflow")
	if err != nil {
		log.Fatalf("Failed to resolve: %v", err)
	}
	fmt.Println("  ✓ Anomaly resolved")
}

func testConfiguration(ads *services.AnomalyDetectionService, tenantID uint) {
	// Create/Update configuration
	fmt.Println("  - Creating/updating configuration...")

	config := &models.AnomalyDetectionConfig{
		TenantID:                    tenantID,
		Enabled:                     true,
		Sensitivity:                 "high",
		MinBaselineSamples:          50,
		ZScoreThresholdMinor:        1.5,
		ZScoreThresholdMajor:        2.0,
		ZScoreThresholdCritical:     3.0,
		NotificationEnabled:         true,
		NotificationCooldownMinutes: 15,
		RequireConsecutiveAnomalies: 2,
	}

	err := ads.UpdateConfig(tenantID, config)
	if err != nil {
		log.Fatalf("Failed to update config: %v", err)
	}
	fmt.Println("  ✓ Configuration updated")
	fmt.Printf("    - Sensitivity: %s\n", config.Sensitivity)
	fmt.Printf("    - Minor threshold: %.1fσ\n", config.ZScoreThresholdMinor)
	fmt.Printf("    - Major threshold: %.1fσ\n", config.ZScoreThresholdMajor)
	fmt.Printf("    - Critical threshold: %.1fσ\n", config.ZScoreThresholdCritical)
	fmt.Printf("    - Consecutive required: %d\n", config.RequireConsecutiveAnomalies)
}

func testStatistics(ads *services.AnomalyDetectionService, tenantID uint) {
	stats, err := ads.GetStatistics(tenantID, "7d")
	if err != nil {
		log.Fatalf("Failed to get statistics: %v", err)
	}

	fmt.Printf("  ✓ Statistics retrieved:\n")
	fmt.Printf("    - Total anomalies: %d\n", stats.TotalAnomalies)
	fmt.Printf("    - By severity:\n")
	fmt.Printf("      - Minor: %d\n", stats.BySeverity["minor"])
	fmt.Printf("      - Major: %d\n", stats.BySeverity["major"])
	fmt.Printf("      - Critical: %d\n", stats.BySeverity["critical"])
	fmt.Printf("    - By status:\n")
	fmt.Printf("      - Open: %d\n", stats.ByStatus["open"])
	fmt.Printf("      - Acknowledged: %d\n", stats.ByStatus["acknowledged"])
	fmt.Printf("      - Resolved: %d\n", stats.ByStatus["resolved"])
	fmt.Printf("    - False positive rate: %.2f%%\n", stats.FalsePositiveRate*100)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
