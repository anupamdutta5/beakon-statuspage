#!/bin/bash

# migrate-service-to-v2.sh
# Migrates a single service from shared-resilience v1.x to v2.0

set -e

SERVICE_NAME=$1

if [ -z "$SERVICE_NAME" ]; then
    echo "Usage: ./migrate-service-to-v2.sh <service-name>"
    exit 1
fi

SERVICE_DIR="microservices/$SERVICE_NAME"

if [ ! -d "$SERVICE_DIR" ]; then
    echo "Error: Service directory not found: $SERVICE_DIR"
    exit 1
fi

echo "Migrating $SERVICE_NAME to v2.0..."

# Check if config.v2.yml exists
if [ ! -f "$SERVICE_DIR/configs/config.v2.yml" ]; then
    echo "Error: $SERVICE_DIR/configs/config.v2.yml not found"
    exit 1
fi

# Check if already migrated (has backup file)
if [ -f "$SERVICE_DIR/cmd/main.v1.backup.go" ]; then
    echo "✅ $SERVICE_NAME already migrated (backup file exists)"
    exit 0
fi

# Use payment-service as template since it's the cleanest v2 implementation
TEMPLATE_FILE="microservices/payment-service/cmd/main.go"

if [ ! -f "$TEMPLATE_FILE" ]; then
    echo "Error: Template file not found: $TEMPLATE_FILE"
    exit 1
fi

# Generate main_v2.go from template
echo "Generating main_v2.go from template..."
SERVICE_CLASS=$(echo $SERVICE_NAME | sed 's/-service//' | sed 's/-/ /g; s/\b\(.\)/\u\1/g; s/ //g')
SERVICE_SNAKE=$(echo $SERVICE_NAME | sed 's/-service//' | tr '-' '_')

sed "s/payment-service/${SERVICE_NAME}/g; s/Payment/${SERVICE_CLASS}/g; s/payment/${SERVICE_SNAKE}/g; s/beakon:payment/beakon:${SERVICE_SNAKE}/g" "$TEMPLATE_FILE" > "$SERVICE_DIR/cmd/main_v2.go"

echo "✅ Generated main_v2.go for $SERVICE_NAME"
echo ""
echo "Next steps:"
echo "1. cd $SERVICE_DIR"
echo "2. Update routes in cmd/main_v2.go to match your handler methods"
echo "3. go build -o ${SERVICE_NAME}-v2 ./cmd/main_v2.go"
echo "4. If successful:"
echo "   mv cmd/main.go cmd/main.v1.backup.go"
echo "   mv cmd/main_v2.go cmd/main.go"
echo "   mv configs/config.yml configs/config.v1.backup.yml"
echo "   mv configs/config.v2.yml configs/config.yml"
