#!/bin/bash

echo "🔧 Comprehensive fix for all microservices..."

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

    # 1. Backup and clean go.mod
    cp go.mod go.mod.backup

    # Remove malformed lines that were added by previous script
    echo "🔧 Cleaning malformed go.mod entries"
    grep -v "^github.com/lib/pq" go.mod | \
    grep -v "^github.com/anupamdutta5/statuspage-shared-resilience v0.0.0$" > go.mod.tmp
    mv go.mod.tmp go.mod

    # 2. Fix package declarations in database directory
    if [[ -f "internal/database/logger.go" ]]; then
        echo "🔧 Fixing package declaration in database/logger.go"
        sed -i.bak 's/^package logger$/package database/' internal/database/logger.go
        rm -f internal/database/logger.go.bak
    fi

    # 3. Properly add dependencies to go.mod
    echo "🔧 Adding dependencies using go get"

    # Add shared-resilience dependency if not present
    if ! grep -q "statuspage-shared-resilience" go.mod; then
        echo "  - Adding shared-resilience dependency"
        go mod edit -require=github.com/anupamdutta5/statuspage-shared-resilience@v0.0.0
        go mod edit -replace=github.com/anupamdutta5/statuspage-shared-resilience=../shared-resilience
    fi

    # Add lib/pq dependency if not present
    if ! grep -q "github.com/lib/pq" go.mod; then
        echo "  - Adding lib/pq dependency"
        go get github.com/lib/pq@v1.10.9
    fi

    # 4. Run go mod tidy to fix dependencies
    echo "🔧 Running go mod tidy"
    go mod tidy

    # 5. Fix specific code issues
    echo "🔧 Fixing code issues"

    # Fix shutdown manager errors import
    if [[ -f "internal/shutdown/manager.go" ]]; then
        # Remove unused errors import
        sed -i.bak '/^import (/,/)/ { /^\s*"errors"$/d; }' internal/shutdown/manager.go
        # Fix errors.New usage
        sed -i.bak 's/errors\.New(/fmt.Errorf(/g' internal/shutdown/manager.go
        rm -f internal/shutdown/manager.go.bak
    fi

    # Fix validation unused import
    if [[ -f "internal/validation/validator.go" ]]; then
        sed -i.bak '/^import (/,/)/ { /^\s*"regexp"$/d; }' internal/validation/validator.go
        rm -f internal/validation/validator.go.bak
    fi

    # Remove duplicate logger declarations
    if [[ -f "internal/database/logger.go" ]] && [[ -f "internal/database/gorm_logger.go" ]]; then
        echo "  - Removing duplicate logger file"
        rm -f internal/database/logger.go
    fi

    # Remove unused resilience import
    if [[ -f "internal/database/manager.go" ]]; then
        sed -i.bak '/resilience "github.com\/enterprise-status\/statuspage-shared-resilience"/d' internal/database/manager.go
        rm -f internal/database/manager.go.bak
    fi

    # 6. Quick build test
    echo "🧪 Testing build"
    if go build ./... > /dev/null 2>&1; then
        echo "✅ $service_name fixed successfully"
        rm -f go.mod.backup
    else
        echo "⚠️ $service_name still has issues - showing first few errors:"
        go build ./... 2>&1 | head -10
        # Restore backup if fix failed
        if [[ -f "go.mod.backup" ]]; then
            cp go.mod.backup go.mod
            rm -f go.mod.backup
        fi
    fi

    cd - > /dev/null
    echo ""
done

echo "🏁 Comprehensive fix complete!"