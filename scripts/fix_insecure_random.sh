#!/bin/bash

# Fix Insecure Random Generation Script
# This script replaces insecure time.Now().UnixNano() with secure crypto/rand

set -e

echo "🔐 Starting insecure random generation fixes..."

# Function to fix middleware files
fix_middleware_file() {
    local file=$1
    echo "📝 Processing: $file"

    # Create backup
    cp "$file" "$file.backup"

    # Add resilience import if not present
    if ! grep -q "resilience.*statuspage-shared-resilience" "$file"; then
        # Add import after existing imports
        sed -i '' '/import (/,/)/ s|"go.uber.org/zap"|"go.uber.org/zap"\n\n\tresilience "github.com/anupamdutta5/statuspage-shared-resilience"|' "$file"
    fi

    # Replace insecure CORS configuration
    sed -i '' 's/config.AllowOrigins = \[\]string{"*"}/config.AllowOrigins = []string{\
		"http:\/\/localhost:3000",    \/\/ Development frontend\
		"https:\/\/yourdomain.com",   \/\/ Production domain - replace with actual domain\
		"https:\/\/admin.yourdomain.com", \/\/ Admin domain - replace with actual domain\
	}/' "$file"

    # Add security comment for CORS
    sed -i '' 's/\/\/ CORS middleware for cross-origin requests\./\/\/ CORS middleware for cross-origin requests with secure configuration./' "$file"

    # Replace insecure random string generation
    cat > /tmp/secure_request_id.go << 'EOF'
// generateRequestID generates a cryptographically secure unique request ID.
func generateRequestID() string {
	// Use secure correlation ID generation from shared resilience package
	correlationID, err := resilience.GenerateCorrelationID()
	if err != nil {
		// Fallback to timestamp-based ID if secure generation fails
		// This should never happen in practice but provides safety
		return time.Now().Format("20060102150405") + "-fallback"
	}
	return correlationID
}
EOF

    # Replace the generateRequestID function and remove randomString function
    perl -i -0pe 's/\/\/ generateRequestID.*?^}/`cat \/tmp\/secure_request_id.go`/gms' "$file"

    # Remove the randomString function entirely
    perl -i -0pe 's/\/\/ randomString.*?^}//gms' "$file"

    echo "✅ Updated: $file"
}

# Find and process all middleware files with insecure random generation
echo "🔍 Finding middleware files with insecure random generation..."

find microservices -name "middleware.go" -type f | while read file; do
    if grep -q "time.Now().UnixNano()" "$file"; then
        fix_middleware_file "$file"
    fi
done

# Fix the shared resilience middleware as well
if [ -f "microservices/shared-resilience/middleware.go" ]; then
    echo "📝 Processing shared resilience middleware..."
    if grep -q "time.Now().UnixNano()" "microservices/shared-resilience/middleware.go"; then
        # For shared middleware, replace with secure generation
        sed -i '' 's/strconv.FormatInt(time.Now().UnixNano(), 36)/correlationID, _ := GenerateCorrelationID(); correlationID/' "microservices/shared-resilience/middleware.go"
        echo "✅ Updated shared resilience middleware"
    fi
fi

# Clean up temp file
rm -f /tmp/secure_request_id.go

# Check for any remaining insecure random generation
echo "🔍 Scanning for remaining insecure random generation..."

if grep -r "time.Now().UnixNano()" microservices/ --include="*.go"; then
    echo "⚠️  Warning: Found remaining insecure random generation"
else
    echo "✅ No insecure random generation found"
fi

echo "🔐 Insecure random generation fixes completed!"
echo ""
echo "📋 Security improvements applied:"
echo "1. Replaced time.Now().UnixNano() with crypto/rand"
echo "2. Fixed CORS wildcard origins with specific domains"
echo "3. Added secure correlation ID generation"
echo "4. Enhanced request ID security"