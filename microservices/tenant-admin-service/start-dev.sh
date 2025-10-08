#!/bin/bash

# Load environment variables from project root
if [ -f "../../.env" ]; then
    set -a
    source ../../.env
    set +a
fi

# Set development environment
export ENVIRONMENT=development
export SERVER_PORT=8099

# Run the service
go run cmd/main.go
