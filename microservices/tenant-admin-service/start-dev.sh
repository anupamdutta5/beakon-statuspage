#!/bin/bash

# Load environment variables from project root
if [ -f "../../.env" ]; then
    set -a
    source ../../.env
    set +a
fi

# Set development environment
export ENVIRONMENT=development
export DB_HOST=localhost
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=tenant_admin_db
export JWT_SECRET=development-secret-key-statuspage-2024
export SERVER_PORT=8099

# Run the service
go run cmd/main.go
