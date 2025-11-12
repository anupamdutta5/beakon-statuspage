#!/bin/bash

cd /Users/anuoamdutta/Desktop/statuspage/Beakon

for svc in microservices/*/; do
  if [ -d "$svc/tests/unit" ]; then
    svc_name=$(basename "$svc")
    echo "=== Testing $svc_name ==="
    cd "$svc"
    if go test -c ./tests/unit 2>&1 | grep -q "ok"; then
      echo "✅ $svc_name: PASS"
    else
      echo "❌ $svc_name: FAIL"
      go test -c ./tests/unit 2>&1 | head -3
    fi
    cd ../..
  fi
done
