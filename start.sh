#!/bin/bash

# Set required environment variables
export JWT_SECRET="development-secret-key-change-in-production"
export STRIPE_SECRET_KEY="sk_test_development_key"
export STRIPE_WEBHOOK_SECRET="whsec_development_webhook"
export STRIPE_PUBLISHABLE_KEY="pk_test_development_key"
export DB_PASSWORD="postgres"

# Start the application
go run ./cmd/api
