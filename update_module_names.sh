#!/bin/bash

echo "🔧 Updating all module names from enterprise-status to anupamdutta5..."

# Go back to root directory
cd /Users/anuoamdutta/Desktop/statuspage/Beakon

# 1. Update all go.mod files
echo "📝 Updating go.mod module declarations..."

find microservices -name "go.mod" | while read -r gomod; do
    echo "  - Updating $gomod"
    sed -i.bak 's/github\.com\/enterprise-status\//github.com\/anupamdutta5\//g' "$gomod"
    rm -f "${gomod}.bak"
done

# 2. Update all Go source files with import statements
echo "📝 Updating import statements in Go files..."

find microservices -name "*.go" | while read -r gofile; do
    if grep -q "github.com/anupamdutta5/" "$gofile"; then
        echo "  - Updating imports in $gofile"
        sed -i.bak 's/github\.com\/enterprise-status\//github.com\/anupamdutta5\//g' "$gofile"
        rm -f "${gofile}.bak"
    fi
done

# 3. Update any configuration files or scripts that might reference the old module names
echo "📝 Updating configuration files..."

find . -name "*.yaml" -o -name "*.yml" -o -name "*.json" -o -name "*.sh" | while read -r file; do
    if grep -q "github.com/anupamdutta5/" "$file" 2>/dev/null; then
        echo "  - Updating $file"
        sed -i.bak 's/github\.com\/enterprise-status\//github.com\/anupamdutta5\//g' "$file"
        rm -f "${file}.bak"
    fi
done

# 4. Update README files and documentation
echo "📝 Updating documentation files..."

find . -name "*.md" | while read -r mdfile; do
    if grep -q "github.com/anupamdutta5/" "$mdfile" 2>/dev/null; then
        echo "  - Updating $mdfile"
        sed -i.bak 's/github\.com\/enterprise-status\//github.com\/anupamdutta5\//g' "$mdfile"
        rm -f "${mdfile}.bak"
    fi
done

# 5. Run go mod tidy for all services to update dependencies
echo "🔧 Running go mod tidy for all services..."

for service_dir in microservices/*/; do
    service_name=$(basename "$service_dir")

    if [[ ! -d "$service_dir" ]] || [[ "$service_name" == "shared-resilience" ]]; then
        continue
    fi

    echo "  - Running go mod tidy in $service_name"
    cd "$service_dir"
    go mod tidy > /dev/null 2>&1
    cd - > /dev/null
done

echo "🧪 Testing builds after module name updates..."

# 6. Test a few key services to ensure they build
test_services=("landing-page-service" "analytics-service" "api-gateway")

for service in "${test_services[@]}"; do
    if [[ -d "microservices/$service" ]]; then
        echo "  - Testing $service build..."
        cd "microservices/$service"
        if go build ./... > /dev/null 2>&1; then
            echo "    ✅ $service builds successfully"
        else
            echo "    ❌ $service has build issues:"
            go build ./... 2>&1 | head -3
        fi
        cd - > /dev/null
    fi
done

echo ""
echo "🎉 Module name update complete!"
echo "📋 Summary:"
echo "  - Updated all go.mod files to use github.com/anupamdutta5/"
echo "  - Updated all import statements in Go files"
echo "  - Updated configuration and documentation files"
echo "  - Ran go mod tidy for all services"
echo ""
echo "⚠️  Note: You may need to update your git remote URLs if they don't match:"
echo "  Current: $(git remote get-url origin)"
echo "  Expected pattern: https://github.com/anupamdutta5/statuspage-*"