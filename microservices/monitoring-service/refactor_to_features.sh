#!/bin/bash

# Monitoring Service Refactoring Script
# Transforms layer-based architecture to feature-based modular architecture
# Date: 2025-10-25

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "🚀 Starting Monitoring Service Refactoring..."
echo "Current directory: $SCRIPT_DIR"

# Step 1: Create new feature-based directory structure
echo ""
echo "📁 Step 1: Creating feature-based directory structure..."

mkdir -p internal/features/monitors/{http,tcp,ping,ssl,dns}
mkdir -p internal/features/alerts/{core,routing,auto_resolution,deduplication}
mkdir -p internal/features/maintenance/{windows,automation,scheduling}
mkdir -p internal/features/integrations/{slack,pagerduty,discord,telegram,teams,webhook,email}
mkdir -p internal/features/anomaly/{detection,baselines,ml_models}
mkdir -p internal/features/locations/{multi_region,failover}
mkdir -p internal/features/sla/{reporting,calculations}
mkdir -p internal/features/heartbeat
mkdir -p internal/features/escalation
mkdir -p internal/features/performance
mkdir -p internal/features/status_automation
mkdir -p internal/features/docker
mkdir -p internal/features/kubernetes
mkdir -p internal/features/external_monitoring

# Step 2: Create core utilities directory
echo ""
echo "🔧 Step 2: Creating core utilities directory..."

mkdir -p internal/core/database
mkdir -p internal/core/events
mkdir -p internal/core/middleware
mkdir -p internal/core/errors
mkdir -p internal/core/validation
mkdir -p internal/core/config

echo ""
echo "✅ Directory structure created successfully!"
echo ""
echo "📊 New structure:"
tree -L 3 -d internal/features/ internal/core/

echo ""
echo "⚠️  MANUAL STEPS REQUIRED:"
echo "1. Move service files to appropriate feature directories"
echo "2. Move handlers to feature directories"
echo "3. Move models to feature directories"
echo "4. Extract routes from main.go to feature route files"
echo "5. Move shared utilities to internal/core/"
echo "6. Update import paths"
echo "7. Run tests and validate"
echo ""
echo "📝 See REFACTORING_PROGRESS_TRACKER.md for detailed task list"
