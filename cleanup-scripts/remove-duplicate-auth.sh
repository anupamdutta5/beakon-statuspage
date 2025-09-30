#!/bin/bash
# Remove Duplicate Authentication Middleware
# This script removes duplicated authentication middleware from individual services
# and updates them to use the shared auth package

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧹 Removing Duplicate Authentication Middleware${NC}"
echo "=================================================="

# Services with their own auth middleware (identified from audit)
SERVICES_WITH_AUTH=(
    "tenant-admin-service"
    "user-service"
    "saas-admin-service"
    "component-service"
    "incident-service"
    "monitoring-service"
    "analytics-service"
    "notification-service"
    "payment-service"
    "branding-service"
    "event-store-service"
)

MICROSERVICES_DIR="microservices"

# Function to backup files before modification
backup_file() {
    local file=$1
    if [ -f "$file" ]; then
        cp "$file" "$file.backup.$(date +%Y%m%d_%H%M%S)"
        echo -e "  ${GREEN}✓${NC} Backed up: $file"
    fi
}

# Function to check if shared-resilience is available
check_shared_resilience() {
    if [ ! -d "$MICROSERVICES_DIR/shared-resilience" ]; then
        echo -e "${RED}❌ shared-resilience directory not found${NC}"
        exit 1
    fi

    if [ ! -f "$MICROSERVICES_DIR/shared-resilience/auth/middleware.go" ]; then
        echo -e "${RED}❌ Shared auth middleware not found${NC}"
        echo "Please ensure shared-resilience/auth/middleware.go exists"
        exit 1
    fi

    echo -e "${GREEN}✓${NC} Shared resilience auth middleware found"
}

# Function to remove duplicate auth files
remove_duplicate_auth() {
    local service=$1
    local service_dir="$MICROSERVICES_DIR/$service"

    if [ ! -d "$service_dir" ]; then
        echo -e "${YELLOW}⚠️  Service directory not found: $service${NC}"
        return
    fi

    echo -e "${BLUE}Processing service: $service${NC}"

    # Remove standalone auth.go files
    local auth_files=(
        "$service_dir/internal/middleware/auth.go"
        "$service_dir/internal/auth/middleware.go"
        "$service_dir/pkg/auth/middleware.go"
    )

    for auth_file in "${auth_files[@]}"; do
        if [ -f "$auth_file" ]; then
            echo -e "  ${YELLOW}Removing duplicate auth file: $auth_file${NC}"
            backup_file "$auth_file"
            rm "$auth_file"
            echo -e "  ${GREEN}✓${NC} Removed: $auth_file"
        fi
    done

    # Update middleware.go files to use shared auth
    local middleware_files=(
        "$service_dir/internal/middleware/middleware.go"
        "$service_dir/pkg/middleware/middleware.go"
    )

    for middleware_file in "${middleware_files[@]}"; do
        if [ -f "$middleware_file" ]; then
            update_middleware_imports "$middleware_file"
        fi
    done

    # Update main.go files to use shared auth
    local main_files=(
        "$service_dir/cmd/main.go"
        "$service_dir/main.go"
    )

    for main_file in "${main_files[@]}"; do
        if [ -f "$main_file" ]; then
            update_main_file_imports "$main_file"
        fi
    done
}

# Function to update middleware imports
update_middleware_imports() {
    local file=$1
    if [ ! -f "$file" ]; then
        return
    fi

    echo -e "  ${BLUE}Updating middleware imports in: $file${NC}"
    backup_file "$file"

    # Check if file contains JWT-related auth code
    if grep -q "jwt\|JWT\|Claims\|AuthMiddleware" "$file"; then
        echo -e "  ${YELLOW}Found JWT code in $file - needs manual review${NC}"
        echo -e "  ${YELLOW}Please review and update to use: github.com/anupamdutta5/shared-resilience/auth${NC}"

        # Add a comment to the file indicating it needs review
        echo -e "\n// TODO: CLEANUP - Review and update to use shared auth middleware" >> "$file"
        echo -e "// Import: github.com/anupamdutta5/shared-resilience/auth" >> "$file"
        echo -e "// Replace local auth implementations with shared.NewMiddleware()" >> "$file"
    fi
}

# Function to update main.go imports
update_main_file_imports() {
    local file=$1
    if [ ! -f "$file" ]; then
        return
    fi

    echo -e "  ${BLUE}Updating main file imports in: $file${NC}"
    backup_file "$file"

    # Check if file contains auth middleware usage
    if grep -q "AuthMiddleware\|auth\.New\|middleware\.Auth" "$file"; then
        echo -e "  ${YELLOW}Found auth middleware usage in $file - needs manual review${NC}"
        echo -e "  ${YELLOW}Please update to use shared auth middleware${NC}"

        # Add comments to guide the update
        echo -e "\n// TODO: CLEANUP - Update auth middleware usage" >> "$file"
        echo -e "// Replace local auth with: auth.NewMiddleware(authConfig, logger)" >> "$file"
        echo -e "// Import: github.com/anupamdutta5/shared-resilience/auth" >> "$file"
    fi
}

# Function to update go.mod files
update_go_mod() {
    local service=$1
    local service_dir="$MICROSERVICES_DIR/$service"
    local go_mod="$service_dir/go.mod"

    if [ ! -f "$go_mod" ]; then
        return
    fi

    echo -e "  ${BLUE}Checking go.mod for JWT dependencies: $go_mod${NC}"

    # Check if golang-jwt is used directly (might be redundant now)
    if grep -q "github.com/golang-jwt/jwt" "$go_mod"; then
        echo -e "  ${YELLOW}⚠️  Service has direct JWT dependency - might be redundant${NC}"
        echo -e "  ${YELLOW}Consider removing if using shared auth${NC}"
    fi
}

# Main execution
main() {
    # Check prerequisites
    check_shared_resilience

    echo -e "${BLUE}Starting cleanup of duplicate authentication middleware...${NC}"
    echo

    # Process each service
    for service in "${SERVICES_WITH_AUTH[@]}"; do
        remove_duplicate_auth "$service"
        update_go_mod "$service"
        echo
    done

    # Generate summary report
    echo -e "${BLUE}📊 Cleanup Summary${NC}"
    echo "=================="

    # Count backup files created
    backup_count=$(find "$MICROSERVICES_DIR" -name "*.backup.*" 2>/dev/null | wc -l)
    echo -e "Backup files created: $backup_count"

    # List services that need manual review
    echo -e "\n${YELLOW}Services requiring manual review:${NC}"

    manual_review_needed=0
    for service in "${SERVICES_WITH_AUTH[@]}"; do
        service_dir="$MICROSERVICES_DIR/$service"
        if find "$service_dir" -name "*.go" -exec grep -l "TODO: CLEANUP" {} \; 2>/dev/null | head -1 > /dev/null; then
            echo -e "  - $service"
            ((manual_review_needed++))
        fi
    done

    if [ $manual_review_needed -eq 0 ]; then
        echo -e "${GREEN}None - all services can use automated cleanup${NC}"
    else
        echo -e "\n${YELLOW}Manual steps required:${NC}"
        echo "1. Review files with 'TODO: CLEANUP' comments"
        echo "2. Update imports to use shared-resilience/auth"
        echo "3. Replace local auth implementations with shared middleware"
        echo "4. Test authentication flows in each service"
        echo "5. Remove backup files after verification"
    fi

    echo
    echo -e "${GREEN}✅ Authentication middleware cleanup completed${NC}"
    echo -e "${YELLOW}💡 Next steps:${NC}"
    echo "1. Update each service to import: github.com/anupamdutta5/shared-resilience/auth"
    echo "2. Replace local auth middleware with: auth.NewMiddleware(config, logger)"
    echo "3. Test authentication in each service"
    echo "4. Run integration tests: ./test-scripts/test-inter-service-communication.sh"
}

# Run main function
main "$@"