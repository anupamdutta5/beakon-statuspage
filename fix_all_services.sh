#!/bin/bash

echo "🔧 Fixing all microservices tests..."

# Function to fix a service
fix_service() {
    local service_name=$1
    local test_file="microservices/$service_name/tests/unit/*_test.go"
    
    echo "Fixing $service_name..."
    
    # Add authentication middleware to all router setups
    find microservices/$service_name/tests/unit/ -name "*_test.go" -exec sed -i '' 's/router := gin\.New()/router := gin.New()\
	router.Use(func(c *gin.Context) {\
		c.Set("tenant_id", uint(1))\
		c.Set("user_id", uint(1))\
		c.Next()\
	})/g' {} \;
    
    echo "✅ $service_name fixed"
}

# Fix all services with test issues
fix_service "component-service"
fix_service "notification-service"
fix_service "payment-service"
fix_service "incident-service"
fix_service "monitoring-service"
fix_service "tenant-service"
fix_service "analytics-service"
fix_service "branding-service"
fix_service "database-service"
fix_service "event-store-service"
fix_service "user-service"

echo "🎉 All services fixed!"
