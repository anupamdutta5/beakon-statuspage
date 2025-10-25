// Test program for maintenance automation features
package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type MaintenanceWindow struct {
	ID               uint      `gorm:"primarykey"`
	TenantID         string    `gorm:"type:uuid"`
	Name             string
	Description      string
	StartsAt         time.Time
	EndsAt           time.Time
	Status           string
	ReminderSent     bool
	AutoStarted      bool
	AutoCompleted    bool
	ActualStartTime  *time.Time
	ActualEndTime    *time.Time
	IsActive         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func main() {
	// Connect to database
	dsn := "host=localhost user=postgres password=postgres dbname=monitoring_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("===========================================")
	fmt.Println("Maintenance Automation Test Suite")
	fmt.Println("===========================================\n")

	// Clean up any existing test data
	db.Exec("DELETE FROM maintenance_windows WHERE name LIKE 'TEST_%'")

	testTenantID := "123e4567-e89b-12d3-a456-426614174000"

	// Test 1: Auto-reminder (60 minutes before)
	fmt.Println("Test 1: Auto-Reminder (60 minutes before start)")
	fmt.Println("-------------------------------------------------")

	reminderWindow := MaintenanceWindow{
		TenantID:    testTenantID,
		Name:        "TEST_REMINDER_WINDOW",
		Description: "Testing 60-minute reminder automation",
		StartsAt:    time.Now().Add(55 * time.Minute), // Starts in 55 minutes (within 60-min window)
		EndsAt:      time.Now().Add(115 * time.Minute),
		Status:      "scheduled",
		ReminderSent: false,
		IsActive:    true,
	}

	if err := db.Create(&reminderWindow).Error; err != nil {
		log.Fatalf("Failed to create reminder test window: %v", err)
	}
	fmt.Printf("✅ Created maintenance window (ID: %d) starting in 55 minutes\n", reminderWindow.ID)
	fmt.Printf("   Status: %s, ReminderSent: %v\n", reminderWindow.Status, reminderWindow.ReminderSent)

	// Wait for background job to process (runs every minute)
	fmt.Println("\n⏳ Waiting 90 seconds for background job to send reminder...")
	time.Sleep(90 * time.Second)

	// Check if reminder was sent
	db.First(&reminderWindow, reminderWindow.ID)
	if reminderWindow.ReminderSent {
		fmt.Println("✅ PASS: Reminder was sent successfully!")
	} else {
		fmt.Println("❌ FAIL: Reminder was NOT sent")
	}
	fmt.Printf("   Updated ReminderSent: %v\n\n", reminderWindow.ReminderSent)

	// Test 2: Auto-start (when start time arrives)
	fmt.Println("Test 2: Auto-Start (transition to 'in_progress')")
	fmt.Println("-------------------------------------------------")

	autoStartWindow := MaintenanceWindow{
		TenantID:    testTenantID,
		Name:        "TEST_AUTO_START_WINDOW",
		Description: "Testing auto-start automation",
		StartsAt:    time.Now().Add(-2 * time.Minute), // Started 2 minutes ago
		EndsAt:      time.Now().Add(58 * time.Minute),
		Status:      "scheduled",
		AutoStarted: false,
		IsActive:    true,
	}

	if err := db.Create(&autoStartWindow).Error; err != nil {
		log.Fatalf("Failed to create auto-start test window: %v", err)
	}
	fmt.Printf("✅ Created maintenance window (ID: %d) that started 2 minutes ago\n", autoStartWindow.ID)
	fmt.Printf("   Status: %s, AutoStarted: %v\n", autoStartWindow.Status, autoStartWindow.AutoStarted)

	// Wait for background job to process
	fmt.Println("\n⏳ Waiting 90 seconds for background job to auto-start...")
	time.Sleep(90 * time.Second)

	// Check if window was auto-started
	db.First(&autoStartWindow, autoStartWindow.ID)
	if autoStartWindow.Status == "in_progress" {
		fmt.Println("✅ PASS: Window was auto-started successfully!")
	} else {
		fmt.Println("❌ FAIL: Window was NOT auto-started")
	}
	fmt.Printf("   Updated Status: %s\n\n", autoStartWindow.Status)

	// Test 3: Auto-complete (when end time arrives)
	fmt.Println("Test 3: Auto-Complete (transition to 'completed')")
	fmt.Println("-------------------------------------------------")

	autoCompleteWindow := MaintenanceWindow{
		TenantID:      testTenantID,
		Name:          "TEST_AUTO_COMPLETE_WINDOW",
		Description:   "Testing auto-complete automation",
		StartsAt:      time.Now().Add(-62 * time.Minute), // Started 62 minutes ago
		EndsAt:        time.Now().Add(-2 * time.Minute),   // Ended 2 minutes ago
		Status:        "in_progress",
		AutoCompleted: false,
		IsActive:      true,
	}

	if err := db.Create(&autoCompleteWindow).Error; err != nil {
		log.Fatalf("Failed to create auto-complete test window: %v", err)
	}
	fmt.Printf("✅ Created maintenance window (ID: %d) that ended 2 minutes ago\n", autoCompleteWindow.ID)
	fmt.Printf("   Status: %s, AutoCompleted: %v\n", autoCompleteWindow.Status, autoCompleteWindow.AutoCompleted)

	// Wait for background job to process
	fmt.Println("\n⏳ Waiting 90 seconds for background job to auto-complete...")
	time.Sleep(90 * time.Second)

	// Check if window was auto-completed
	db.First(&autoCompleteWindow, autoCompleteWindow.ID)
	if autoCompleteWindow.Status == "completed" {
		fmt.Println("✅ PASS: Window was auto-completed successfully!")
	} else {
		fmt.Println("❌ FAIL: Window was NOT auto-completed")
	}
	fmt.Printf("   Updated Status: %s\n\n", autoCompleteWindow.Status)

	// Summary
	fmt.Println("===========================================")
	fmt.Println("Test Summary")
	fmt.Println("===========================================")

	passCount := 0
	if reminderWindow.ReminderSent {
		passCount++
	}
	if autoStartWindow.Status == "in_progress" {
		passCount++
	}
	if autoCompleteWindow.Status == "completed" {
		passCount++
	}

	fmt.Printf("Tests Passed: %d/3\n", passCount)

	if passCount == 3 {
		fmt.Println("\n🎉 ALL TESTS PASSED! Maintenance automation is working correctly.")
	} else {
		fmt.Println("\n⚠️  Some tests failed. Check the logs at /tmp/monitoring-service.log")
	}

	// Clean up test data
	fmt.Println("\n🧹 Cleaning up test data...")
	db.Exec("DELETE FROM maintenance_windows WHERE name LIKE 'TEST_%'")
	fmt.Println("✅ Cleanup complete")
}
