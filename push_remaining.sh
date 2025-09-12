#!/bin/bash

# Simple script to push remaining microservices

services=(
    "component-service:statuspage-component-service-private"
    "incident-service:statuspage-incident-service-private"
    "payment-service:statuspage-payment-service-private"
    "analytics-service:statuspage-analytics-service-private"
    "monitoring-service:statuspage-monitoring-service-private"
    "notification-service:statuspage-notification-service-private"
    "branding-service:statuspage-branding-service-private"
    "saas-admin-service:statuspage-saas-admin-service-private"
    "tenant-admin-service:statuspage-tenant-admin-service-private"
    "database-service:statuspage-database-service-private"
    "event-store-service:statuspage-event-store-service-private"
    "landing-page-service:statuspage-landing-page-service-private"
    "analytics-consumer:statuspage-analytics-consumer-private"
    "audit-consumer:statuspage-audit-consumer-private"
    "billing-consumer:statuspage-billing-consumer-private"
)

for service_info in "${services[@]}"; do
    service_name=$(echo $service_info | cut -d: -f1)
    repo_name=$(echo $service_info | cut -d: -f2)
    
    echo "🚀 Processing $service_name..."
    
    cd "/Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/$service_name"
    
    git init
    git add .
    git commit -m "Initial commit: $service_name microservice"
    git remote add origin "https://github.com/anupamdutta5/$repo_name.git"
    git push -u origin main
    
    echo "✅ Completed $service_name"
    echo ""
done

echo "🎉 All remaining microservices pushed to GitHub!"
