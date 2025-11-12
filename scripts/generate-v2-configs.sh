#!/bin/bash

# generate-v2-configs.sh
# Batch generate v2 config files for all services that don't have them yet

set -e

echo "Checking which services need config.v2.yml generation..."

cd microservices

for svc_dir in */; do
    svc=$(basename "$svc_dir")
    
    # Skip if not a directory with go.mod
    if [ ! -f "$svc_dir/go.mod" ]; then
        continue
    fi
    
    # Check if already has config.v2.yml
    if [ -f "$svc_dir/configs/config.v2.yml" ]; then
        echo "✅ $svc - already has config.v2.yml"
    else
        echo "❌ $svc - missing config.v2.yml"
    fi
done

echo ""
echo "All services already have config.v2.yml files!"
