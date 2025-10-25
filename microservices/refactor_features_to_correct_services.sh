#!/bin/bash

# Refactor features from monitoring-service to correct microservices
# This script moves features to their architecturally correct services

set -e

MICROSERVICES_DIR="/Users/anuoamdutta/Desktop/statuspage/Beakon/microservices"
MONITORING_SERVICE="$MICROSERVICES_DIR/monitoring-service"
INCIDENT_SERVICE="$MICROSERVICES_DIR/incident-service"
NOTIFICATION_SERVICE="$MICROSERVICES_DIR/notification-service"
ANALYTICS_SERVICE="$MICROSERVICES_DIR/analytics-service"
COMPONENT_SERVICE="$MICROSERVICES_DIR/component-service"

echo "========================================="
echo "Feature Refactoring to Correct Services"
echo "========================================="
echo ""

# Create feature directories if they don't exist
create_feature_structure() {
    local service=$1
    echo "Creating feature structure in $service..."
    mkdir -p "$service/internal/features"
    mkdir -p "$service/internal/core/middleware"
    mkdir -p "$service/internal/core/config"
    mkdir -p "$service/internal/core/database"
    mkdir -p "$service/internal/core/events"
    mkdir -p "$service/internal/core/validation"
}

# STEP 1: Move alert features to incident-service
echo "STEP 1: Moving alerts to incident-service..."
create_feature_structure "$INCIDENT_SERVICE"

if [ -d "$MONITORING_SERVICE/internal/features/alerts" ]; then
    echo "  - Copying alerts features..."
    cp -r "$MONITORING_SERVICE/internal/features/alerts" "$INCIDENT_SERVICE/internal/features/"
fi

if [ -d "$MONITORING_SERVICE/internal/features/anomaly" ]; then
    echo "  - Copying anomaly detection features..."
    cp -r "$MONITORING_SERVICE/internal/features/anomaly" "$INCIDENT_SERVICE/internal/features/"
fi

if [ -d "$MONITORING_SERVICE/internal/features/escalation" ]; then
    echo "  - Copying escalation features..."
    cp -r "$MONITORING_SERVICE/internal/features/escalation" "$INCIDENT_SERVICE/internal/features/"
fi

if [ -d "$MONITORING_SERVICE/internal/features/status_automation" ]; then
    echo "  - Copying status automation features..."
    cp -r "$MONITORING_SERVICE/internal/features/status_automation" "$INCIDENT_SERVICE/internal/features/"
fi

# STEP 2: Move integrations to notification-service
echo ""
echo "STEP 2: Moving integrations to notification-service..."
create_feature_structure "$NOTIFICATION_SERVICE"

if [ -d "$MONITORING_SERVICE/internal/features/integrations" ]; then
    echo "  - Copying integration features..."
    cp -r "$MONITORING_SERVICE/internal/features/integrations" "$NOTIFICATION_SERVICE/internal/features/"
fi

# STEP 3: Move SLA features to analytics-service
echo ""
echo "STEP 3: Moving SLA features to analytics-service..."
create_feature_structure "$ANALYTICS_SERVICE"

if [ -d "$MONITORING_SERVICE/internal/features/sla" ]; then
    echo "  - Copying SLA features..."
    cp -r "$MONITORING_SERVICE/internal/features/sla" "$ANALYTICS_SERVICE/internal/features/"
fi

# STEP 4: Copy core infrastructure to all services
echo ""
echo "STEP 4: Copying core infrastructure to all services..."

for service in "$INCIDENT_SERVICE" "$NOTIFICATION_SERVICE" "$ANALYTICS_SERVICE"; do
    service_name=$(basename "$service")
    echo "  - Copying core to $service_name..."

    # Copy core components if they don't exist
    if [ ! -d "$service/internal/core/middleware" ] && [ -d "$MONITORING_SERVICE/internal/core/middleware" ]; then
        cp -r "$MONITORING_SERVICE/internal/core/middleware" "$service/internal/core/"
    fi

    if [ ! -d "$service/internal/core/database" ] && [ -d "$MONITORING_SERVICE/internal/core/database" ]; then
        cp -r "$MONITORING_SERVICE/internal/core/database" "$service/internal/core/"
    fi

    if [ ! -d "$service/internal/core/events" ] && [ -d "$MONITORING_SERVICE/internal/core/events" ]; then
        cp -r "$MONITORING_SERVICE/internal/core/events" "$service/internal/core/"
    fi

    if [ ! -d "$service/internal/core/validation" ] && [ -d "$MONITORING_SERVICE/internal/core/validation" ]; then
        cp -r "$MONITORING_SERVICE/internal/core/validation" "$service/internal/core/"
    fi
done

echo ""
echo "========================================="
echo "Feature Distribution Complete!"
echo "========================================="
echo ""
echo "Summary:"
echo "  ✓ incident-service: alerts, anomaly, escalation, status_automation"
echo "  ✓ notification-service: integrations (7 channels)"
echo "  ✓ analytics-service: SLA calculations and reporting"
echo "  ✓ monitoring-service: retains core monitoring features"
echo ""
echo "Next steps:"
echo "  1. Update import paths in each service"
echo "  2. Update API endpoints"
echo "  3. Wire services together"
echo "  4. Test end-to-end"
echo ""
