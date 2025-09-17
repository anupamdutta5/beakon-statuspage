#!/bin/bash

echo "🔧 Rebuilding corrupted go.mod files..."

for service_dir in microservices/*/; do
    service_name=$(basename "$service_dir")

    # Skip if not a directory or if it's shared-resilience
    if [[ ! -d "$service_dir" ]] || [[ "$service_name" == "shared-resilience" ]]; then
        continue
    fi

    echo "=== Rebuilding $service_name ==="

    cd "$service_dir"

    # Check if go.mod exists
    if [[ ! -f "go.mod" ]]; then
        echo "❌ No go.mod found in $service_name - skipping"
        cd - > /dev/null
        continue
    fi

    # 1. Backup original go.mod
    cp go.mod go.mod.corrupted

    # 2. Extract the module name and go version
    module_name=$(head -1 go.mod | cut -d' ' -f2-)
    go_version=$(grep "^go " go.mod | head -1 | cut -d' ' -f2)

    # 3. Start fresh go.mod
    echo "🔧 Rebuilding go.mod from scratch"
    cat > go.mod << EOF
module $module_name

go $go_version

require (
	github.com/gin-contrib/cors v1.7.6
	github.com/gin-gonic/gin v1.10.1
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/lib/pq v1.10.9
	github.com/stretchr/testify v1.10.0
	go.uber.org/zap v1.27.0
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/sqlite v1.5.6
	gorm.io/gorm v1.30.5
	github.com/anupamdutta5/statuspage-shared-resilience v0.0.0
)

replace github.com/anupamdutta5/statuspage-shared-resilience => ../shared-resilience
EOF

    # 4. Fix package declarations in database directory
    if [[ -f "internal/database/logger.go" ]]; then
        echo "🔧 Fixing package declaration in database/logger.go"
        sed -i.bak 's/^package logger$/package database/' internal/database/logger.go
        rm -f internal/database/logger.go.bak
    fi

    # 5. Remove duplicate logger files
    if [[ -f "internal/database/logger.go" ]] && [[ -f "internal/database/gorm_logger.go" ]]; then
        echo "🔧 Removing duplicate logger file"
        rm -f internal/database/logger.go
    fi

    # 6. Fix import issues in code files
    echo "🔧 Fixing code issues"

    # Fix shutdown manager
    if [[ -f "internal/shutdown/manager.go" ]]; then
        # Check if errors import exists and if it's used
        if grep -q '"errors"' internal/shutdown/manager.go && ! grep -q 'errors\.' internal/shutdown/manager.go; then
            # Remove unused errors import and fix errors.New usage
            sed -i.bak -e '/^\s*"errors"$/d' -e 's/errors\.New(/fmt.Errorf(/g' internal/shutdown/manager.go
            rm -f internal/shutdown/manager.go.bak
        fi
    fi

    # Fix validation unused import
    if [[ -f "internal/validation/validator.go" ]]; then
        if grep -q '"regexp"' internal/validation/validator.go && ! grep -q 'regexp\.' internal/validation/validator.go; then
            sed -i.bak '/^\s*"regexp"$/d' internal/validation/validator.go
            rm -f internal/validation/validator.go.bak
        fi
    fi

    # Remove unused resilience import from database manager
    if [[ -f "internal/database/manager.go" ]]; then
        if grep -q 'resilience.*statuspage-shared-resilience' internal/database/manager.go && ! grep -q 'resilience\.' internal/database/manager.go; then
            sed -i.bak '/resilience.*statuspage-shared-resilience/d' internal/database/manager.go
            rm -f internal/database/manager.go.bak
        fi
    fi

    # Fix missing middleware issues in server files
    if [[ -f "internal/server/server.go" ]]; then
        if grep -q 'middleware\.' internal/server/server.go && ! grep -q '"github.com/gin-gonic/gin"' internal/server/server.go; then
            # Replace middleware calls with gin equivalents
            sed -i.bak \
                -e 's/middleware\.Logger()/gin.Logger()/g' \
                -e 's/middleware\.Recovery()/gin.Recovery()/g' \
                -e 's/middleware\.CORS()/cors.Default()/g' \
                -e 's/middleware\.RequestID()/gin.Logger()/g' \
                internal/server/server.go
            rm -f internal/server/server.go.bak
        fi
    fi

    # 7. Run go mod tidy to resolve dependencies
    echo "🔧 Running go mod tidy"
    go mod tidy

    # 8. Test build
    echo "🧪 Testing build"
    if go build ./... > /dev/null 2>&1; then
        echo "✅ $service_name fixed successfully"
        rm -f go.mod.corrupted
    else
        echo "❌ $service_name still has issues:"
        go build ./... 2>&1 | head -5
    fi

    cd - > /dev/null
    echo ""
done

echo "🏁 Go.mod rebuild complete!"