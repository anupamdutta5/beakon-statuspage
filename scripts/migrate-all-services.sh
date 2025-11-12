#!/bin/bash

# migrate-all-services.sh
# Batch migrate all services that are ready for v2.0

set -e

echo "================================================"
echo "Beakon V2.0 Batch Migration Tool"
echo "================================================"
echo ""

cd microservices

# Services already migrated (have backup files)
echo "✅ ALREADY MIGRATED:"
for svc_dir in */; do
    if [ -f "$svc_dir/cmd/main.v1.backup.go" ]; then
        svc=$(basename "$svc_dir")
        lines=$(wc -l < "$svc_dir/cmd/main.go" 2>/dev/null | tr -d ' ' || echo "N/A")
        echo "  - $svc ($lines lines)"
    fi
done

echo ""
echo "🔶 NEEDS MIGRATION:"

# Services that need migration
NEED_MIGRATION=()
for svc_dir in */; do
    svc=$(basename "$svc_dir")
    
    # Skip if not a service directory
    if [ ! -f "$svc_dir/go.mod" ]; then
        continue
    fi
    
    # Skip if already migrated
    if [ -f "$svc_dir/cmd/main.v1.backup.go" ]; then
        continue
    fi
    
    # Skip deprecated services
    if [ "$svc" = "database-service" ] || [ "$svc" = "api-gateway" ]; then
        continue
    fi
    
    # Check if has config.v2.yml
    if [ ! -f "$svc_dir/configs/config.v2.yml" ]; then
        echo "  - $svc (❌ No config.v2.yml)"
        continue
    fi
    
    lines=$(wc -l < "$svc_dir/cmd/main.go" 2>/dev/null | tr -d ' ' || echo "N/A")
    echo "  - $svc ($lines lines)"
    NEED_MIGRATION+=("$svc")
done

echo ""
echo "================================================"
echo "Total services needing migration: ${#NEED_MIGRATION[@]}"
echo "================================================"

# Show service categories
echo ""
echo "📋 SERVICE CATEGORIES:"
echo ""
echo "HTTP Services (standard pattern):"
echo "  - None remaining (all require custom handling)"
echo ""
echo "Services with custom config:"
for svc in "${NEED_MIGRATION[@]}"; do
    if [[ ! "$svc" =~ -consumer$ ]]; then
        echo "  - $svc"
    fi
done
echo ""
echo "Consumer services (no HTTP server):"
for svc in "${NEED_MIGRATION[@]}"; do
    if [[ "$svc" =~ -consumer$ ]]; then
        echo "  - $svc"
    fi
done

echo ""
echo "================================================"
echo "NEXT STEPS:"
echo "================================================"
echo "1. All remaining services require manual refactoring"
echo "2. Custom config services need constructor changes"
echo "3. Consumer services need specialized template"
echo "4. See SERVICES_V2_MIGRATION_STATUS.md for details"
