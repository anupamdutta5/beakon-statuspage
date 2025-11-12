#!/bin/bash

# Script to fix V2 constructor issues in all unit tests
# This fixes the common pattern where tests use old v1.x constructors

cd /Users/anuoamdutta/Desktop/statuspage/Beakon

echo "=== Fixing Unit Tests for V2 Migration ==="

# Event Store Service
echo "Fixing event-store-service..."
if [ -f "microservices/event-store-service/tests/unit/event_store_test.go" ]; then
  sed -i '' 's/eventStoreService, _ := services.NewEventStoreService(cfg, logger)/eventStoreService := services.NewEventStoreService(db, logger)/g' microservices/event-store-service/tests/unit/event_store_test.go
  # Remove SetDB calls
  sed -i '' '/eventStoreService.SetDB(db)/d' microservices/event-store-service/tests/unit/event_store_test.go
  # Remove cfg := &config.Config{} if followed by NewEventStoreService
  sed -i '' '/cfg := &config.Config{}$/,/eventStoreService :=/{/cfg := &config.Config{}/d;}' microservices/event-store-service/tests/unit/event_store_test.go
  echo "✅ event-store-service fixed"
fi

# Landing Page Service
echo "Fixing landing-page-service..."
if [ -f "microservices/landing-page-service/tests/unit/landing_page_test.go" ]; then
  sed -i '' 's/landingService, _ := services.NewLandingService([^)]*/landingService := services.NewLandingService(db, logger)/g' microservices/landing-page-service/tests/unit/landing_page_test.go
  sed -i '' '/landingService.SetDB(db)/d' microservices/landing-page-service/tests/unit/landing_page_test.go
  echo "✅ landing-page-service fixed"
fi

echo ""
echo "=== Done! ==="
echo "Now run: cd microservices && for svc in event-store-service landing-page-service; do cd \$svc && go test -c ./tests/unit && cd ..; done"
