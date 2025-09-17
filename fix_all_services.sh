#!/bin/bash

echo "🔧 Fixing all microservices systematically..."

for service_dir in microservices/*/; do
    service_name=$(basename "$service_dir")

    # Skip if not a directory or if it's shared-resilience
    if [[ ! -d "$service_dir" ]] || [[ "$service_name" == "shared-resilience" ]]; then
        continue
    fi

    echo "=== Fixing $service_name ==="

    cd "$service_dir"

    # Check if go.mod exists
    if [[ ! -f "go.mod" ]]; then
        echo "❌ No go.mod found in $service_name - skipping"
        cd - > /dev/null
        continue
    fi

    # 1. Fix package declaration conflicts in database directory
    if [[ -f "internal/database/logger.go" ]]; then
        echo "🔧 Fixing package declaration in database/logger.go"
        sed -i.bak 's/^package logger$/package database/' internal/database/logger.go
        rm -f internal/database/logger.go.bak
    fi

    # 2. Add missing dependencies to go.mod
    echo "🔧 Adding missing dependencies to go.mod"

    # Check if shared-resilience is already in go.mod
    if ! grep -q "statuspage-shared-resilience" go.mod; then
        # Add shared-resilience dependency
        if grep -q "^require (" go.mod; then
            # Insert after require block starts
            sed -i.bak '/^require (/a\
\	github.com/anupamdutta5/statuspage-shared-resilience v0.0.0' go.mod
        else
            # Add require block
            echo -e "\nrequire (\n\tgithub.com/anupamdutta5/statuspage-shared-resilience v0.0.0\n)" >> go.mod
        fi
    fi

    # Check if lib/pq is already in go.mod
    if ! grep -q "github.com/lib/pq" go.mod; then
        # Add lib/pq dependency
        if grep -q "^require (" go.mod; then
            sed -i.bak '/github.com\/enterprise-status\/statuspage-shared-resilience/a\
\	github.com/lib/pq v1.10.9' go.mod
        fi
    fi

    # Check if replace directive exists
    if ! grep -q "replace.*statuspage-shared-resilience" go.mod; then
        echo -e "\nreplace github.com/anupamdutta5/statuspage-shared-resilience => ../shared-resilience" >> go.mod
    fi

    # Clean up backup files
    rm -f go.mod.bak

    # 3. Run go mod tidy to fix dependencies
    echo "🔧 Running go mod tidy"
    go mod tidy > /dev/null 2>&1

    # 4. Quick build test
    echo "🧪 Testing build"
    if go build ./... > /dev/null 2>&1; then
        echo "✅ $service_name fixed successfully"
    else
        echo "⚠️ $service_name still has issues (will fix in next iteration)"
    fi

    cd - > /dev/null
    echo ""
done

echo "🎉 First pass complete! Running second pass for remaining issues..."

# Second pass to handle remaining issues
for service_dir in microservices/*/; do
    service_name=$(basename "$service_dir")

    if [[ ! -d "$service_dir" ]] || [[ "$service_name" == "shared-resilience" ]]; then
        continue
    fi

    cd "$service_dir"

    if [[ ! -f "go.mod" ]]; then
        cd - > /dev/null
        continue
    fi

    # Check for remaining build issues
    build_output=$(go build ./... 2>&1)
    if [[ $? -ne 0 ]]; then
        echo "🔧 Second pass fixing $service_name"

        # Fix go.mod tidy issues
        if echo "$build_output" | grep -q "go mod tidy"; then
            echo "  - Running go mod tidy again"
            go mod tidy
        fi

        # Fix missing dependencies
        if echo "$build_output" | grep -q "github.com/gin-gonic/gin"; then
            echo "  - Adding gin dependency"
            go get github.com/gin-gonic/gin
        fi

        # Final build test
        if go build ./... > /dev/null 2>&1; then
            echo "✅ $service_name fixed in second pass"
        else
            echo "❌ $service_name needs manual intervention"
        fi
    fi

    cd - > /dev/null
done

echo "🏁 All services processing complete!"