#!/bin/bash

# Beakon Status Page - Development Setup Script
# This script sets up the development environment for the microservices

set -e

echo "🚀 Beakon Status Page - Development Setup"
echo "=========================================="

# Check if required tools are installed
check_tool() {
    if ! command -v "$1" &> /dev/null; then
        echo "❌ $1 is not installed. Please install it first."
        exit 1
    fi
    echo "✅ $1 is available"
}

echo "📋 Checking required tools..."
check_tool "go"
check_tool "docker"
check_tool "docker-compose"

echo ""
echo "📦 Setting up Go workspace..."
go work sync

echo ""
echo "🔧 Installing dependencies for all services..."
for service in microservices/*/; do
    if [ -f "$service/go.mod" ]; then
        echo "   📁 Setting up $(basename $service)..."
        cd "$service" && go mod tidy && cd - > /dev/null
    fi
done

echo ""
echo "🧪 Running a quick build test..."
if make build > /dev/null 2>&1; then
    echo "✅ Build test passed"
else
    echo "❌ Build test failed - please check the services"
    exit 1
fi

echo ""
echo "📋 Creating necessary directories..."
mkdir -p logs
mkdir -p data

echo ""
echo "📄 Environment setup..."
if [ ! -f .env ]; then
    echo "   📝 Copying .env.template to .env"
    cp .env.template .env
    echo "   ⚠️  Please edit .env file with your actual configuration values"
else
    echo "   ✅ .env file already exists"
fi

echo ""
echo "🎉 Development environment setup complete!"
echo ""
echo "Next steps:"
echo "1. Edit .env file with your configuration"
echo "2. Run 'make run' to start the services"
echo "3. Run 'make test' to run all tests"
echo "4. Run 'make help' to see all available commands"
echo ""