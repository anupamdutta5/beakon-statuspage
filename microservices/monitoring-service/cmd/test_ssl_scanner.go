// Test program for SSL certificate scanner
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/anupamdutta5/monitoring-service/internal/services"
)

func main() {
	// Connect to database
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

	log.Println("✅ Connected to database")

	// Create SSL scanner service
	sslService := services.NewSSLScannerService(db)

	// Test tenant ID
	tenantID := uuid.New()

	// Test domains
	testDomains := []string{
		"google.com",
		"github.com",
		"api.github.com",
	}

	log.Println("\n🔍 Testing SSL Certificate Scanner...")
	log.Println("=====================================")

	for _, domain := range testDomains {
		log.Printf("\n📡 Scanning %s...", domain)

		cert, err := sslService.ScanDomain(tenantID, domain)
		if err != nil {
			log.Printf("  ❌ Error: %v", err)
			continue
		}

		log.Printf("  ✅ Certificate found!")
		log.Printf("     Domain: %s", cert.Domain)
		log.Printf("     Issuer: %s", cert.Issuer)
		log.Printf("     Valid From: %s", cert.ValidFrom.Format("2006-01-02"))
		log.Printf("     Valid Until: %s", cert.ValidUntil.Format("2006-01-02"))
		if cert.DaysUntilExpiry != nil {
			log.Printf("     Days Until Expiry: %d", *cert.DaysUntilExpiry)
		}
		log.Printf("     Is Valid: %v", cert.IsValid)
		log.Printf("     Is Self-Signed: %v", cert.IsSelfSigned)

		// Check if warning needed
		shouldSend, warningType := cert.ShouldSendWarning()
		if shouldSend {
			log.Printf("     ⚠️  Warning needed: %s", warningType)
		}
	}

	// Get all scanned certificates
	log.Println("\n📋 All Scanned Certificates:")
	log.Println("=====================================")

	certs, err := sslService.GetTenantCertificates(tenantID)
	if err != nil {
		log.Fatalf("Failed to fetch certificates: %v", err)
	}

	log.Printf("Total certificates: %d\n", len(certs))
	for _, cert := range certs {
		log.Printf("  • %s (expires in %d days)", cert.Domain, *cert.DaysUntilExpiry)
	}

	log.Println("\n✅ SSL Scanner Test Complete!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
