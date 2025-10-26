#!/bin/bash
#
# Generate init-db.sh for a microservice
# Usage: ./generate-init-db.sh <service-name> <database-name>
#

set -e

if [ $# -lt 2 ]; then
    echo "Usage: $0 <service-name> <database-name>"
    echo "Example: $0 user-service statuspage_user"
    exit 1
fi

SERVICE_NAME=$1
DB_NAME=$2
SERVICE_DIR="./microservices/$SERVICE_NAME"
TEMPLATE="./scripts/init-db-template.sh"

# Validate service directory exists
if [ ! -d "$SERVICE_DIR" ]; then
    echo "Error: Service directory not found: $SERVICE_DIR"
    exit 1
fi

# Validate template exists
if [ ! -f "$TEMPLATE" ]; then
    echo "Error: Template not found: $TEMPLATE"
    exit 1
fi

# Generate init-db.sh from template
cat "$TEMPLATE" | \
    sed "s/__SERVICE_NAME__/$SERVICE_NAME/g" | \
    sed "s/__DB_NAME__/$DB_NAME/g" \
    > "$SERVICE_DIR/init-db.sh"

# Make executable
chmod +x "$SERVICE_DIR/init-db.sh"

echo "✓ Generated init-db.sh for $SERVICE_NAME"
echo "  Database: $DB_NAME"
echo "  Location: $SERVICE_DIR/init-db.sh"
