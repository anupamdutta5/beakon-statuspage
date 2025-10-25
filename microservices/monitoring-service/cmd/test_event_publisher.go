// Test program for RabbitMQ event publisher
package main

import (
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/anupamdutta5/monitoring-service/internal/core/events"
)

func main() {
	// Get RabbitMQ URL from environment or use default
	rabbitmqURL := getEnv("RABBITMQ_URL", "amqp://admin:SecureP@ssw0rd2024!@localhost:5672/")

	log.Println("🔧 Connecting to RabbitMQ...")
	publisher, err := events.NewEventPublisher(rabbitmqURL)
	if err != nil {
		log.Fatalf("Failed to create event publisher: %v", err)
	}
	defer publisher.Close()

	log.Println("✅ Connected to RabbitMQ successfully")

	// Test SSL Expiring Event
	tenantID := uuid.New()
	log.Println("\n📤 Publishing SSL expiring event...")

	err = publisher.PublishSSLExpiring(events.SSLExpiringEvent{
		TenantID:        tenantID,
		CertificateID:   1,
		Domain:          "test-domain.com",
		DaysUntilExpiry: 7,
		WarningType:     "7d",
	})

	if err != nil {
		log.Fatalf("Failed to publish SSL expiring event: %v", err)
	}

	log.Println("✅ SSL expiring event published successfully")

	// Test Health Check
	log.Println("\n🏥 Testing RabbitMQ health check...")
	if err := publisher.HealthCheck(); err != nil {
		log.Fatalf("Health check failed: %v", err)
	}

	log.Println("✅ RabbitMQ connection is healthy")
	log.Println("\n✅ Event Publisher Test Complete!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
