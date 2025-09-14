#!/bin/bash

echo "🔧 Fixing response parsing issues across all services..."

# Function to fix response parsing in a service
fix_service_response_parsing() {
    local service_name=$1
    local test_file="microservices/$service_name/tests/unit/*_test.go"
    
    echo "Fixing response parsing in $service_name..."
    
    # Fix common response parsing patterns
    find microservices/$service_name/tests/unit/ -name "*_test.go" -exec sed -i '' 's/var response models\.\([A-Za-z]*\)/var responseWrapper struct {\
		\1 models.\1 `json:"\L\1"`\
	}/g' {} \;
    
    find microservices/$service_name/tests/unit/ -name "*_test.go" -exec sed -i '' 's/err = json\.Unmarshal(w\.Body\.Bytes(), &response)/err = json.Unmarshal(w.Body.Bytes(), \&responseWrapper)/g' {} \;
    
    find microservices/$service_name/tests/unit/ -name "*_test.go" -exec sed -i '' 's/assert\.Equal(t, [^,]*\.\([A-Za-z]*\), response\.\([A-Za-z]*\))/assert.Equal(t, \1.\2, responseWrapper.\1.\2)/g' {} \;
    
    echo "✅ $service_name response parsing fixed"
}

# Fix services with response parsing issues
fix_service_response_parsing "incident-service"
fix_service_response_parsing "monitoring-service"
fix_service_response_parsing "payment-service"
fix_service_response_parsing "notification-service"
fix_service_response_parsing "tenant-service"

echo "🎉 All response parsing issues fixed!"
