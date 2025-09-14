#!/bin/bash

# Fix all test functions to include logger initialization
# Add logger declaration after "// Setup" in each test function

# Find all test functions and add logger initialization
sed -i '' '/func Test.*(t \*testing\.T) {/,/\/\/ Setup$/{s/\/\/ Setup$/\/\/ Setup\n\tlogger, _ := zap.NewDevelopment()/;}' tests/unit/analytics_test.go

echo "Fixed analytics service tests"
