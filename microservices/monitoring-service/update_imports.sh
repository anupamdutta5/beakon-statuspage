#!/bin/bash

# Update Import Paths After Refactoring
# This script updates all import paths from old layer-based to new feature-based structure

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "🔄 Updating import paths across all Go files..."
echo ""

# Phase 1: Update Core Utilities import paths
echo "Phase 1: Updating Core Utilities imports..."

# Database imports
find . -name "*.go" -type f -exec sed -i '' \
  's|github.com/anupamdutta5/monitoring-service/internal/database|github.com/anupamdutta5/monitoring-service/internal/core/database|g' {} \;
echo "  ✅ Database imports updated"

# Events imports
find . -name "*.go" -type f -exec sed -i '' \
  's|github.com/anupamdutta5/monitoring-service/internal/events|github.com/anupamdutta5/monitoring-service/internal/core/events|g' {} \;
echo "  ✅ Events imports updated"

# Middleware imports
find . -name "*.go" -type f -exec sed -i '' \
  's|github.com/anupamdutta5/monitoring-service/internal/middleware|github.com/anupamdutta5/monitoring-service/internal/core/middleware|g' {} \;
echo "  ✅ Middleware imports updated"

# Config imports
find . -name "*.go" -type f -exec sed -i '' \
  's|github.com/anupamdutta5/monitoring-service/internal/config|github.com/anupamdutta5/monitoring-service/internal/core/config|g' {} \;
echo "  ✅ Config imports updated"

# Utils/Validation imports
find . -name "*.go" -type f -exec sed -i '' \
  's|github.com/anupamdutta5/monitoring-service/internal/utils|github.com/anupamdutta5/monitoring-service/internal/core/validation|g' {} \;
echo "  ✅ Validation imports updated"

# Shutdown imports
find . -name "*.go" -type f -exec sed -i '' \
  's|github.com/anupamdutta5/monitoring-service/internal/shutdown|github.com/anupamdutta5/monitoring-service/internal/core/shutdown|g' {} \;
echo "  ✅ Shutdown imports updated"

echo ""
echo "✅ All import paths updated!"
echo ""
echo "🔍 Verifying changes..."

# Count updated files
UPDATED_FILES=$(find . -name "*.go" -type f -exec grep -l "internal/core/" {} \; | wc -l)
echo "📊 $UPDATED_FILES files now reference internal/core/"

echo ""
echo "Next step: Build the service to validate changes"
echo "Run: go build -o monitoring-service cmd/main.go"
