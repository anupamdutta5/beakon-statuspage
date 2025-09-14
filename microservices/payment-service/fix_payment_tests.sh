#!/bin/bash

# Fix all test functions to include logger initialization
# Add logger declaration after "// Setup" in each test function

# Find all test functions and add logger initialization
sed -i '' '/func Test.*(t \*testing\.T) {/,/\/\/ Setup$/{s/\/\/ Setup$/\/\/ Setup\n\tlogger, _ := zap.NewDevelopment()/;}' tests/unit/payment_test.go

echo "Fixed payment service tests"
