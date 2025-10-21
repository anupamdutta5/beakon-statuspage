#!/bin/bash

# Load environment variables from project root
if [ -f "../../.env" ]; then
    set -a
    source ../../.env
    set +a
fi

# Set development environment
export ENVIRONMENT=development

# Run the service
go run cmd/main.go
