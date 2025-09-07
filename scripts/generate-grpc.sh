#!/bin/bash

# Generate gRPC code from proto files
# This script generates Go code from Protocol Buffer definitions

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    print_error "protoc is not installed. Please install Protocol Buffers compiler."
    print_status "Installation instructions:"
    print_status "  macOS: brew install protobuf"
    print_status "  Ubuntu: apt-get install protobuf-compiler"
    print_status "  CentOS: yum install protobuf-compiler"
    exit 1
fi

# Check if protoc-gen-go is installed
if ! command -v protoc-gen-go &> /dev/null; then
    print_error "protoc-gen-go is not installed. Please install the Go protobuf plugin."
    print_status "Installation: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    exit 1
fi

# Check if protoc-gen-go-grpc is installed
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    print_error "protoc-gen-go-grpc is not installed. Please install the Go gRPC plugin."
    print_status "Installation: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    exit 1
fi

# Create output directories
print_status "Creating output directories..."
mkdir -p proto/generated/user
mkdir -p proto/generated/tenant
mkdir -p proto/generated/component

# Generate Go code for user service
print_status "Generating Go code for user service..."
protoc \
    --go_out=proto/generated/user \
    --go_opt=paths=source_relative \
    --go-grpc_out=proto/generated/user \
    --go-grpc_opt=paths=source_relative \
    proto/user.proto

# Generate Go code for tenant service
print_status "Generating Go code for tenant service..."
protoc \
    --go_out=proto/generated/tenant \
    --go_opt=paths=source_relative \
    --go-grpc_out=proto/generated/tenant \
    --go-grpc_opt=paths=source_relative \
    proto/tenant.proto

# Generate Go code for component service
print_status "Generating Go code for component service..."
protoc \
    --go_out=proto/generated/component \
    --go_opt=paths=source_relative \
    --go-grpc_out=proto/generated/component \
    --go-grpc_opt=paths=source_relative \
    proto/component.proto

# Generate additional services if they exist
if [ -f "proto/incident.proto" ]; then
    print_status "Generating Go code for incident service..."
    mkdir -p proto/generated/incident
    protoc \
        --go_out=proto/generated/incident \
        --go_opt=paths=source_relative \
        --go-grpc_out=proto/generated/incident \
        --go-grpc_opt=paths=source_relative \
        proto/incident.proto
fi

if [ -f "proto/notification.proto" ]; then
    print_status "Generating Go code for notification service..."
    mkdir -p proto/generated/notification
    protoc \
        --go_out=proto/generated/notification \
        --go_opt=paths=source_relative \
        --go-grpc_out=proto/generated/notification \
        --go-grpc_opt=paths=source_relative \
        proto/notification.proto
fi

if [ -f "proto/payment.proto" ]; then
    print_status "Generating Go code for payment service..."
    mkdir -p proto/generated/payment
    protoc \
        --go_out=proto/generated/payment \
        --go_opt=paths=source_relative \
        --go-grpc_out=proto/generated/payment \
        --go-grpc_opt=paths=source_relative \
        proto/payment.proto
fi

if [ -f "proto/analytics.proto" ]; then
    print_status "Generating Go code for analytics service..."
    mkdir -p proto/generated/analytics
    protoc \
        --go_out=proto/generated/analytics \
        --go_opt=paths=source_relative \
        --go-grpc_out=proto/generated/analytics \
        --go-grpc_opt=paths=source_relative \
        proto/analytics.proto
fi

if [ -f "proto/monitoring.proto" ]; then
    print_status "Generating Go code for monitoring service..."
    mkdir -p proto/generated/monitoring
    protoc \
        --go_out=proto/generated/monitoring \
        --go_opt=paths=source_relative \
        --go-grpc_out=proto/generated/monitoring \
        --go-grpc_opt=paths=source_relative \
        proto/monitoring.proto
fi

# Create a summary of generated files
print_status "Generated files:"
find proto/generated -name "*.pb.go" | while read file; do
    print_status "  $file"
done

print_status "gRPC code generation completed successfully!"

# Optional: Format the generated code
if command -v gofmt &> /dev/null; then
    print_status "Formatting generated code..."
    find proto/generated -name "*.pb.go" -exec gofmt -w {} \;
    print_status "Code formatting completed!"
fi

# Optional: Run go mod tidy
if [ -f "go.mod" ]; then
    print_status "Running go mod tidy..."
    go mod tidy
    print_status "Dependencies updated!"
fi
