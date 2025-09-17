#!/bin/bash

# Script to fix CORS wildcard configurations and update middleware files

echo "Fixing CORS wildcard configurations..."

# Find all middleware files with wildcard CORS origins
find microservices -name "middleware.go" -type f | while read -r file; do
    if grep -q 'AllowOrigins.*\[\]string{".*\*.*"}' "$file"; then
        echo "Processing: $file"

        # Backup original file
        cp "$file" "${file}.backup"

        # Check if file already has resilience import
        if ! grep -q "resilience.*github.com/anupamdutta5/statuspage-shared-resilience" "$file"; then
            # Add resilience import if not present
            sed -i '' '
                /import (/ {
                    :loop
                    N
                    /^)$/ {
                        i\
\	resilience "github.com/anupamdutta5/statuspage-shared-resilience"
                        b end
                    }
                    b loop
                    :end
                }
            ' "$file"
        fi

        # Check if file needs net/http import
        if ! grep -q '"net/http"' "$file" && grep -q 'http\.Status' "$file"; then
            sed -i '' '
                /import (/ {
                    a\
\	"net/http"
                }
            ' "$file"
        fi

        # Replace wildcard CORS configuration with secure version
        sed -i '' '
            /\/\/ CORS middleware for cross-origin requests\./,/^}/ {
                /\/\/ CORS middleware for cross-origin requests\./ c\
// CORS middleware for cross-origin requests with secure configuration.
                /func CORS() gin\.HandlerFunc {/ {
                    :cors_loop
                    N
                    /^}$/ {
                        c\
func CORS() gin.HandlerFunc {\
	config := cors.DefaultConfig()\
	config.AllowOrigins = []string{\
		"http://localhost:3000",         // Development frontend\
		"https://yourdomain.com",        // Production domain\
		"https://admin.yourdomain.com",  // Admin domain\
		"https://api.yourdomain.com",    // API domain\
	}\
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}\
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}\
	config.ExposeHeaders = []string{"Content-Length"}\
	config.AllowCredentials = true\
	config.MaxAge = 12 * time.Hour\
\
	return cors.New(config)\
}
                        b cors_end
                    }
                    b cors_loop
                    :cors_end
                }
            }
        ' "$file"

        # Replace UUID generation with secure generation if present
        if grep -q 'uuid\.New()\.String()' "$file"; then
            sed -i '' 's/uuid\.New()\.String()/generateRequestID()/g' "$file"

            # Add generateRequestID function if not present
            if ! grep -q 'func generateRequestID()' "$file"; then
                cat >> "$file" << 'EOF'

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
            fi
        fi

        # Replace hard-coded status codes with constants
        sed -i '' 's/c\.AbortWithStatus(500)/c.AbortWithStatus(http.StatusInternalServerError)/g' "$file"

        # Remove uuid import if no longer needed
        if ! grep -q 'uuid\.' "$file"; then
            sed -i '' '/github\.com\/google\/uuid/d' "$file"
        fi

        echo "Fixed: $file"
    else
        echo "Skipping: $file (no wildcard CORS found)"
    fi
done

echo "CORS wildcard fix complete!"