#!/bin/bash

echo "Checking all microservices for build errors..."

services_with_errors=()

for service_dir in microservices/*/; do
    service_name=$(basename "$service_dir")

    # Skip if not a directory or if it's shared-resilience
    if [[ ! -d "$service_dir" ]] || [[ "$service_name" == "shared-resilience" ]]; then
        continue
    fi

    echo "=== Checking $service_name ==="

    cd "$service_dir"

    # Check if go.mod exists
    if [[ ! -f "go.mod" ]]; then
        echo "❌ No go.mod found in $service_name"
        services_with_errors+=("$service_name:no-go-mod")
        cd - > /dev/null
        continue
    fi

    # Check for build errors
    build_output=$(go build ./... 2>&1)
    if [[ $? -ne 0 ]]; then
        echo "❌ Build errors in $service_name:"
        echo "$build_output"
        services_with_errors+=("$service_name:build-error")
    else
        echo "✅ $service_name builds successfully"
    fi

    # Check for vet errors
    vet_output=$(go vet ./... 2>&1)
    if [[ $? -ne 0 ]]; then
        echo "⚠️ Vet warnings in $service_name:"
        echo "$vet_output"
        services_with_errors+=("$service_name:vet-warning")
    fi

    cd - > /dev/null
    echo ""
done

echo "=== SUMMARY ==="
if [[ ${#services_with_errors[@]} -eq 0 ]]; then
    echo "✅ All services are building successfully!"
else
    echo "❌ Services with issues:"
    for error in "${services_with_errors[@]}"; do
        echo "  - $error"
    done
fi