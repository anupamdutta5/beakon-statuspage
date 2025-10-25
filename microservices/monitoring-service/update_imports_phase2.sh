#!/bin/bash

# Phase 2: Update Import Paths for Monitor Features
# Updates all files that reference monitor-related services/handlers/models

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "🔄 Phase 2: Updating import paths for Monitor features..."
echo ""

# Update imports in all Go files to point to new feature locations
echo "Updating imports across all files..."

# Monitor service imports
find . -name "*.go" -type f -exec sed -i '' \
  's|"github.com/anupamdutta5/monitoring-service/internal/services"|"github.com/anupamdutta5/monitoring-service/internal/features/monitors/http"|g' {} \;

# Keep core services imports separate
find . -name "*.go" -type f -exec sed -i '' \
  's|"github.com/anupamdutta5/monitoring-service/internal/features/monitors/http"  // For specific monitor services|"github.com/anupamdutta5/monitoring-service/internal/services"|g' {} \;

echo "✅ Import paths updated"

echo ""
echo "📝 Note: You may need to manually adjust some imports as we're keeping"
echo "   non-monitor services in internal/services/ for now"
echo ""
echo "Next: Validate with 'go build -o monitoring-service cmd/main.go'"
