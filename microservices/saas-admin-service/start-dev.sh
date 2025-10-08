#!/bin/bash

# Load environment variables from project root
if [ -f "../../.env" ]; then
    set -a
    source ../../.env
    set +a
fi

# Set development environment
export ENVIRONMENT=development
export SERVER_PORT=8098
export DB_NAME=saas_admin_db
export DB_PASSWORD=${DB_PASSWORD:-postgres}
export JWT_SECRET=${JWT_SECRET:-development-secret-key-statuspage-2024}
export TENANT_ADMIN_SERVICE_URL=http://localhost:8099

# Run the service
go run cmd/main.go
