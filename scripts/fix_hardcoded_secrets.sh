#!/bin/bash

# Fix Hardcoded Secrets Script
# This script removes hardcoded secrets from all microservice configuration files

set -e

echo "🔒 Starting hardcoded secrets removal..."

# Function to update YAML files with environment variable references
update_yaml_config() {
    local file=$1
    echo "📝 Processing: $file"

    # Create backup
    cp "$file" "$file.backup"

    # Replace hardcoded database passwords
    sed -i '' 's/password: test_password/password: ${DB_PASSWORD}/g' "$file"

    # Replace hardcoded JWT secrets (but preserve production env var refs)
    sed -i '' 's/secret: dev-secret-key-change-in-production/secret: ${JWT_SECRET}/g' "$file"

    # Replace database config with env vars
    sed -i '' 's/host: localhost/host: ${DB_HOST:localhost}/g' "$file"
    sed -i '' 's/port: 5432/port: ${DB_PORT:5432}/g' "$file"
    sed -i '' 's/user: test_user/user: ${DB_USER:statuspage_user}/g' "$file"
    sed -i '' 's/name: statuspage_/name: ${DB_NAME:statuspage_/g' "$file"

    # Replace Redis config with env vars
    sed -i '' 's/redis_host: localhost/redis_host: ${REDIS_HOST:localhost}/g' "$file"
    sed -i '' 's/redis_port: 6379/redis_port: ${REDIS_PORT:6379}/g' "$file"
    sed -i '' 's/redis_password: ""/redis_password: ${REDIS_PASSWORD}/g' "$file"

    # Replace logging level
    sed -i '' 's/level: debug/level: ${LOG_LEVEL:debug}/g' "$file"
    sed -i '' 's/level: info/level: ${LOG_LEVEL:info}/g' "$file"

    echo "✅ Updated: $file"
}

# Find and process all development YAML config files
echo "🔍 Finding configuration files with hardcoded secrets..."

# Process development configs
find microservices -name "development.yaml" -type f | while read file; do
    if grep -q "test_password\|dev-secret-key" "$file"; then
        update_yaml_config "$file"
    fi
done

# Check for any remaining hardcoded secrets
echo "🔍 Scanning for remaining hardcoded secrets..."

# Check for test passwords
if grep -r "test_password" microservices/ --include="*.yaml" --include="*.yml"; then
    echo "⚠️  Warning: Found remaining test_password references"
else
    echo "✅ No test_password references found"
fi

# Check for hardcoded JWT secrets (but ignore env var references)
if grep -r "dev-secret-key-change-in-production" microservices/ --include="*.yaml" --include="*.yml"; then
    echo "⚠️  Warning: Found remaining dev JWT secrets"
else
    echo "✅ No hardcoded JWT secrets found"
fi

# Create gitignore entries for sensitive files
echo "📝 Updating .gitignore for security..."

cat >> .gitignore << 'EOF'

# Security - Never commit these files
.env
.env.local
.env.development
.env.production
*.pem
*.key
secrets/
*.backup
*_backup
EOF

echo "🔒 Hardcoded secrets removal completed!"
echo ""
echo "📋 Next steps:"
echo "1. Copy .env.template to .env and fill in actual values"
echo "2. Set environment variables in your deployment environment"
echo "3. Remove .backup files after verifying changes"
echo "4. Test all services with new configuration"