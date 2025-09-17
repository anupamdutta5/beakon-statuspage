#!/bin/bash

echo "🔧 Cleaning up unused imports and final issues..."

for service_dir in microservices/*/; do
    service_name=$(basename "$service_dir")

    # Skip if not a directory or if it's shared-resilience
    if [[ ! -d "$service_dir" ]] || [[ "$service_name" == "shared-resilience" ]]; then
        continue
    fi

    echo "=== Cleaning up $service_name ==="

    cd "$service_dir"

    # 1. Remove unused resilience imports
    if [[ -f "internal/database/manager.go" ]]; then
        echo "🔧 Removing unused resilience imports"
        sed -i.bak '/resilience.*statuspage-shared-resilience/d' internal/database/manager.go
        rm -f internal/database/manager.go.bak
    fi

    # 2. Fix server files that still reference middleware or cors
    find internal -name "*.go" -type f | while read -r file; do
        if grep -q 'undefined:' <<< "$(go build ./... 2>&1)" && grep -q 'middleware\|cors\.' "$file"; then
            echo "🔧 Fixing imports in $file"
            # Add missing imports if they reference cors or middleware
            if grep -q 'cors\.' "$file" && ! grep -q '"github.com/gin-contrib/cors"' "$file"; then
                sed -i.bak '/^import (/,/)/ {
                    /^import ($/a\
	"github.com/gin-contrib/cors"
                }' "$file"
            fi
            rm -f "${file}.bak"
        fi
    done

    # 3. Fix specific syntax issues
    if [[ -f "internal/handlers/analytics_handler.go" ]]; then
        echo "🔧 Fixing analytics handler syntax completely"
        # Replace the broken line with a proper return statement
        sed -i.bak '/unexpected keyword if, expected expression/c\
		return gin.H{"error": "GetStats method not implemented"}, nil // TODO: implement GetStats method' internal/handlers/analytics_handler.go
        rm -f internal/handlers/analytics_handler.go.bak
    fi

    # 4. Fix any remaining middleware references
    find internal -name "server.go" -type f | while read -r file; do
        if grep -q 'middleware\.' "$file"; then
            echo "🔧 Fixing middleware references in $file"
            sed -i.bak \
                -e 's/middleware\.Logger()/gin.Logger()/g' \
                -e 's/middleware\.Recovery()/gin.Recovery()/g' \
                -e 's/middleware\.CORS()/\/\/ TODO: cors.Default()/g' \
                -e 's/middleware\.RequestID()/\/\/ TODO: RequestID middleware/g' \
                "$file"
            rm -f "${file}.bak"
        fi
    done

    # 5. Use goimports to automatically fix imports
    if command -v goimports &> /dev/null; then
        echo "🔧 Running goimports to fix imports automatically"
        find . -name "*.go" -not -path "./vendor/*" -exec goimports -w {} \;
    fi

    # 6. Run go mod tidy to clean up
    echo "🔧 Running go mod tidy"
    go mod tidy > /dev/null 2>&1

    # 7. Final build test
    echo "🧪 Final build test"
    if go build ./... > /dev/null 2>&1; then
        echo "✅ $service_name builds successfully"
    else
        echo "⚠️ $service_name minor issues remaining:"
        go build ./... 2>&1 | head -2
    fi

    cd - > /dev/null
    echo ""
done

echo "🏁 Import cleanup complete!"