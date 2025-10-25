#!/bin/bash

# Update import paths in refactored services
set -e

MICROSERVICES_DIR="/Users/anuoamdutta/Desktop/statuspage/Beakon/microservices"

echo "========================================="
echo "Updating Import Paths in All Services"
echo "========================================="
echo ""

# Function to update imports in a service
update_service_imports() {
    local service_dir=$1
    local service_name=$2
    local old_module="github.com/anupamdutta5/monitoring-service"
    local new_module="github.com/anupamdutta5/$service_name"

    echo "Updating imports in $service_name..."

    # Find all Go files and update imports
    find "$service_dir/internal/features" -name "*.go" 2>/dev/null | while read -r file; do
        if grep -q "$old_module" "$file" 2>/dev/null; then
            sed -i '' "s|$old_module|$new_module|g" "$file"
        fi
    done

    # Update core imports
    find "$service_dir/internal/core" -name "*.go" 2>/dev/null | while read -r file; do
        if grep -q "$old_module" "$file" 2>/dev/null; then
            sed -i '' "s|$old_module|$new_module|g" "$file"
        fi
    done

    echo "  ✓ Import paths updated in $service_name"
}

# Update each service
update_service_imports "$MICROSERVICES_DIR/incident-service" "incident-service"
update_service_imports "$MICROSERVICES_DIR/notification-service" "notification-service"
update_service_imports "$MICROSERVICES_DIR/analytics-service" "analytics-service"

echo ""
echo "========================================="
echo "Import Path Updates Complete!"
echo "========================================="
