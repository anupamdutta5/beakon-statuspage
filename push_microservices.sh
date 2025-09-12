#!/bin/bash

# Script to push all microservices to their respective GitHub repositories

# Array of microservices and their corresponding repository names
declare -A services=(
    ["tenant-service"]="statuspage-tenant-service-private"
    ["component-service"]="statuspage-component-service-private"
    ["incident-service"]="statuspage-incident-service-private"
    ["payment-service"]="statuspage-payment-service-private"
    ["analytics-service"]="statuspage-analytics-service-private"
    ["monitoring-service"]="statuspage-monitoring-service-private"
    ["notification-service"]="statuspage-notification-service-private"
    ["branding-service"]="statuspage-branding-service-private"
    ["saas-admin-service"]="statuspage-saas-admin-service-private"
    ["tenant-admin-service"]="statuspage-tenant-admin-service-private"
    ["database-service"]="statuspage-database-service-private"
    ["event-store-service"]="statuspage-event-store-service-private"
    ["landing-page-service"]="statuspage-landing-page-service-private"
    ["analytics-consumer"]="statuspage-analytics-consumer-private"
    ["audit-consumer"]="statuspage-audit-consumer-private"
    ["billing-consumer"]="statuspage-billing-consumer-private"
    ["notification-consumer"]="statuspage-notification-consumer-private"
)

# Function to push a microservice
push_service() {
    local service_name=$1
    local repo_name=$2
    
    echo "🚀 Processing $service_name..."
    
    cd "/Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/$service_name"
    
    # Initialize git if not already done
    if [ ! -d ".git" ]; then
        git init
    fi
    
    # Add all files
    git add .
    
    # Commit
    git commit -m "Initial commit: $service_name microservice" || echo "No changes to commit for $service_name"
    
    # Add remote if not exists
    git remote remove origin 2>/dev/null || true
    git remote add origin "https://github.com/anupamdutta5/$repo_name.git"
    
    # Push to GitHub
    git push -u origin main || echo "Failed to push $service_name"
    
    echo "✅ Completed $service_name"
    echo ""
}

# Push all services
for service in "${!services[@]}"; do
    push_service "$service" "${services[$service]}"
done

echo "🎉 All microservices pushed to GitHub!"
