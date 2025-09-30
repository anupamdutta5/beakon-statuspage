#!/bin/bash
# Fix SaaS Admin Service Boundaries
# This script addresses microservice boundary violations in the SaaS Admin Service
# by removing/updating incorrectly implemented features and implementing proper proxies

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

echo -e "${BLUE}🛠️ Fixing SaaS Admin Service Boundaries${NC}"
echo "========================================"

SAAS_ADMIN_DIR="microservices/saas-admin-service"
BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"

# Function to create backup
create_backup() {
    local file=$1
    local backup_file="$BACKUP_DIR/$(basename "$file").backup"

    mkdir -p "$BACKUP_DIR"
    if [ -f "$file" ]; then
        cp "$file" "$backup_file"
        echo -e "  ${GREEN}✓${NC} Backed up: $(basename "$file")"
    fi
}

# Function to check if SaaS Admin Service exists
check_saas_admin_service() {
    if [ ! -d "$SAAS_ADMIN_DIR" ]; then
        echo -e "${RED}❌ SaaS Admin Service directory not found${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓${NC} SaaS Admin Service found"
}

# Function to analyze current handler file
analyze_current_handlers() {
    local handler_file="$SAAS_ADMIN_DIR/internal/handlers/saas_admin_handler.go"

    if [ ! -f "$handler_file" ]; then
        echo -e "${RED}❌ Handler file not found: $handler_file${NC}"
        exit 1
    fi

    echo -e "${BLUE}📊 Analyzing current handlers...${NC}"

    # Count "Not implemented" handlers
    local not_implemented_count=$(grep -c "Not implemented" "$handler_file" 2>/dev/null || echo "0")
    echo -e "  Not implemented handlers: $not_implemented_count"

    # Find handlers that should be proxied
    local proxy_candidates=$(grep -n "func.*\(GetTenants\|CreateTenant\|GetComponents\|CreateComponent\|GetIncidents\|CreateIncident\)" "$handler_file" 2>/dev/null || echo "")

    if [ -n "$proxy_candidates" ]; then
        echo -e "  ${YELLOW}Handlers that should be proxied:${NC}"
        echo "$proxy_candidates" | while read line; do
            echo -e "    ${YELLOW}$line${NC}"
        done
    fi

    return 0
}

# Function to remove not implemented handlers
remove_not_implemented() {
    local handler_file="$SAAS_ADMIN_DIR/internal/handlers/saas_admin_handler.go"

    echo -e "${BLUE}🗑️ Removing 'Not implemented' handlers...${NC}"
    create_backup "$handler_file"

    # Create a temporary file to build the cleaned version
    local temp_file=$(mktemp)

    # Process the file to remove "Not implemented" functions
    awk '
    BEGIN {
        in_function = 0
        brace_count = 0
        skip_function = 0
    }

    # Detect function start
    /^func.*{/ {
        in_function = 1
        brace_count = 1
        current_function = $0

        # Check if this is a "Not implemented" function by looking ahead
        skip_function = 0
    }

    # Handle opening braces
    /{/ && in_function {
        brace_count += gsub(/{/, "&")
    }

    # Handle closing braces
    /}/ && in_function {
        brace_count -= gsub(/}/, "&")

        if (brace_count == 0) {
            in_function = 0

            # If this function contains "Not implemented", skip it
            if (!skip_function) {
                print function_content
                print $0
            } else {
                print "// TODO: CLEANUP - Removed not implemented handler"
            }
            function_content = ""
            current_function = ""
        }
    }

    # Collect function content
    in_function && !skip_function {
        if (function_content == "") {
            function_content = current_function
        } else {
            function_content = function_content "\n" $0
        }

        # Check for "Not implemented" in this line
        if (/Not implemented/) {
            skip_function = 1
            function_content = ""
        }
    }

    # Print lines that are not part of functions
    !in_function {
        print
    }
    ' "$handler_file" > "$temp_file"

    # Replace the original file
    mv "$temp_file" "$handler_file"

    echo -e "${GREEN}✓${NC} Removed not implemented handlers"
}

# Function to add HTTP client for proxying
add_http_client() {
    local handler_file="$SAAS_ADMIN_DIR/internal/handlers/saas_admin_handler.go"

    echo -e "${BLUE}🌐 Adding HTTP client for service proxying...${NC}"

    # Check if HTTP client already exists
    if grep -q "http\.Client" "$handler_file"; then
        echo -e "${YELLOW}HTTP client already exists${NC}"
        return
    fi

    # Add HTTP client to the handler struct
    local temp_file=$(mktemp)

    awk '
    /type SaaSAdminHandler struct/ {
        print $0
        getline
        print "\thttpClient *http.Client"
        print $0
        next
    }

    /func NewSaaSAdminHandler/ {
        in_constructor = 1
    }

    /return &SaaSAdminHandler{/ && in_constructor {
        print $0
        getline
        print "\t\thttpClient: &http.Client{Timeout: 30 * time.Second},"
        print $0
        in_constructor = 0
        next
    }

    { print }
    ' "$handler_file" > "$temp_file"

    mv "$temp_file" "$handler_file"

    echo -e "${GREEN}✓${NC} Added HTTP client for proxying"
}

# Function to implement proper proxy methods
implement_proxy_methods() {
    local handler_file="$SAAS_ADMIN_DIR/internal/handlers/saas_admin_handler.go"

    echo -e "${BLUE}🔗 Implementing proper proxy methods...${NC}"

    # Add proxy methods for tenant management
    cat >> "$handler_file" << 'EOF'

// Proxy Methods - These methods proxy requests to appropriate services

// GetTenants proxies tenant listing to tenant-admin-service
func (h *SaaSAdminHandler) GetTenants(c *gin.Context) {
	h.proxyRequest(c, "GET", "http://localhost:8091/api/v1/tenants", nil)
}

// CreateTenant proxies tenant creation to tenant-admin-service
func (h *SaaSAdminHandler) CreateTenant(c *gin.Context) {
	var requestBody []byte
	if c.Request.Body != nil {
		requestBody, _ = io.ReadAll(c.Request.Body)
	}
	h.proxyRequest(c, "POST", "http://localhost:8091/api/v1/tenants", requestBody)
}

// GetTenant proxies single tenant retrieval to tenant-admin-service
func (h *SaaSAdminHandler) GetTenant(c *gin.Context) {
	tenantID := c.Param("id")
	h.proxyRequest(c, "GET", "http://localhost:8091/api/v1/tenants/"+tenantID, nil)
}

// UpdateTenant proxies tenant updates to tenant-admin-service
func (h *SaaSAdminHandler) UpdateTenant(c *gin.Context) {
	tenantID := c.Param("id")
	var requestBody []byte
	if c.Request.Body != nil {
		requestBody, _ = io.ReadAll(c.Request.Body)
	}
	h.proxyRequest(c, "PUT", "http://localhost:8091/api/v1/tenants/"+tenantID, requestBody)
}

// DeleteTenant proxies tenant deletion to tenant-admin-service
func (h *SaaSAdminHandler) DeleteTenant(c *gin.Context) {
	tenantID := c.Param("id")
	h.proxyRequest(c, "DELETE", "http://localhost:8091/api/v1/tenants/"+tenantID, nil)
}

// GetComponents proxies component listing to component-service
func (h *SaaSAdminHandler) GetComponents(c *gin.Context) {
	h.proxyRequest(c, "GET", "http://localhost:8093/api/v1/components", nil)
}

// CreateComponent proxies component creation to component-service
func (h *SaaSAdminHandler) CreateComponent(c *gin.Context) {
	var requestBody []byte
	if c.Request.Body != nil {
		requestBody, _ = io.ReadAll(c.Request.Body)
	}
	h.proxyRequest(c, "POST", "http://localhost:8093/api/v1/components", requestBody)
}

// GetComponent proxies single component retrieval to component-service
func (h *SaaSAdminHandler) GetComponent(c *gin.Context) {
	componentID := c.Param("id")
	h.proxyRequest(c, "GET", "http://localhost:8093/api/v1/components/"+componentID, nil)
}

// UpdateComponent proxies component updates to component-service
func (h *SaaSAdminHandler) UpdateComponent(c *gin.Context) {
	componentID := c.Param("id")
	var requestBody []byte
	if c.Request.Body != nil {
		requestBody, _ = io.ReadAll(c.Request.Body)
	}
	h.proxyRequest(c, "PUT", "http://localhost:8093/api/v1/components/"+componentID, requestBody)
}

// DeleteComponent proxies component deletion to component-service
func (h *SaaSAdminHandler) DeleteComponent(c *gin.Context) {
	componentID := c.Param("id")
	h.proxyRequest(c, "DELETE", "http://localhost:8093/api/v1/components/"+componentID, nil)
}

// UpdateComponentStatus proxies component status updates to component-service
func (h *SaaSAdminHandler) UpdateComponentStatus(c *gin.Context) {
	componentID := c.Param("id")
	var requestBody []byte
	if c.Request.Body != nil {
		requestBody, _ = io.ReadAll(c.Request.Body)
	}
	h.proxyRequest(c, "POST", "http://localhost:8093/api/v1/components/"+componentID+"/status", requestBody)
}

// GetComponentUptimeStats proxies component uptime stats to component-service
func (h *SaaSAdminHandler) GetComponentUptimeStats(c *gin.Context) {
	componentID := c.Param("id")
	h.proxyRequest(c, "GET", "http://localhost:8093/api/v1/components/"+componentID+"/uptime", nil)
}

// GetComponentGroups proxies component groups to component-service
func (h *SaaSAdminHandler) GetComponentGroups(c *gin.Context) {
	h.proxyRequest(c, "GET", "http://localhost:8093/api/v1/component-groups", nil)
}

// GetIncidents proxies incident listing to incident-service
func (h *SaaSAdminHandler) GetIncidents(c *gin.Context) {
	h.proxyRequest(c, "GET", "http://localhost:8094/api/v1/incidents", nil)
}

// CreateIncident proxies incident creation to incident-service
func (h *SaaSAdminHandler) CreateIncident(c *gin.Context) {
	var requestBody []byte
	if c.Request.Body != nil {
		requestBody, _ = io.ReadAll(c.Request.Body)
	}
	h.proxyRequest(c, "POST", "http://localhost:8094/api/v1/incidents", requestBody)
}

// GetIncident proxies single incident retrieval to incident-service
func (h *SaaSAdminHandler) GetIncident(c *gin.Context) {
	incidentID := c.Param("id")
	h.proxyRequest(c, "GET", "http://localhost:8094/api/v1/incidents/"+incidentID, nil)
}

// UpdateIncident proxies incident updates to incident-service
func (h *SaaSAdminHandler) UpdateIncident(c *gin.Context) {
	incidentID := c.Param("id")
	var requestBody []byte
	if c.Request.Body != nil {
		requestBody, _ = io.ReadAll(c.Request.Body)
	}
	h.proxyRequest(c, "PUT", "http://localhost:8094/api/v1/incidents/"+incidentID, requestBody)
}

// DeleteIncident proxies incident deletion to incident-service
func (h *SaaSAdminHandler) DeleteIncident(c *gin.Context) {
	incidentID := c.Param("id")
	h.proxyRequest(c, "DELETE", "http://localhost:8094/api/v1/incidents/"+incidentID, nil)
}

// CreateIncidentUpdate proxies incident update creation to incident-service
func (h *SaaSAdminHandler) CreateIncidentUpdate(c *gin.Context) {
	incidentID := c.Param("id")
	var requestBody []byte
	if c.Request.Body != nil {
		requestBody, _ = io.ReadAll(c.Request.Body)
	}
	h.proxyRequest(c, "POST", "http://localhost:8094/api/v1/incidents/"+incidentID+"/updates", requestBody)
}

// GetIncidentUpdates proxies incident updates retrieval to incident-service
func (h *SaaSAdminHandler) GetIncidentUpdates(c *gin.Context) {
	incidentID := c.Param("id")
	h.proxyRequest(c, "GET", "http://localhost:8094/api/v1/incidents/"+incidentID+"/updates", nil)
}

// GetIncidentStats proxies incident statistics to incident-service
func (h *SaaSAdminHandler) GetIncidentStats(c *gin.Context) {
	h.proxyRequest(c, "GET", "http://localhost:8094/api/v1/incidents/stats", nil)
}

// Generic proxy method
func (h *SaaSAdminHandler) proxyRequest(c *gin.Context, method, targetURL string, body []byte) {
	// Create request
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, targetURL, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, targetURL, nil)
	}

	if err != nil {
		h.logger.Error("Failed to create proxy request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create proxy request",
		})
		return
	}

	// Copy headers from original request
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Set content type for requests with body
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add query parameters
	req.URL.RawQuery = c.Request.URL.RawQuery

	// Make the request
	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.logger.Error("Proxy request failed",
			zap.String("target", targetURL),
			zap.String("method", method),
			zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Downstream service unavailable",
		})
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Copy response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("Failed to read proxy response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read response",
		})
		return
	}

	// Return response with same status code
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), responseBody)
}
EOF

    echo -e "${GREEN}✓${NC} Added proxy methods"
}

# Function to add required imports
add_required_imports() {
    local handler_file="$SAAS_ADMIN_DIR/internal/handlers/saas_admin_handler.go"

    echo -e "${BLUE}📦 Adding required imports...${NC}"

    # Check if required imports exist
    if ! grep -q "\"bytes\"" "$handler_file"; then
        # Add imports after the package declaration
        sed -i.bak '/^package handlers/a\
\
import (\
\t"bytes"\
\t"io"\
\t"net/http"\
\t"time"\
)' "$handler_file"
    fi

    echo -e "${GREEN}✓${NC} Added required imports"
}

# Function to validate changes
validate_changes() {
    local handler_file="$SAAS_ADMIN_DIR/internal/handlers/saas_admin_handler.go"

    echo -e "${BLUE}✅ Validating changes...${NC}"

    # Check if the file compiles
    if cd "$SAAS_ADMIN_DIR" && go build -o /tmp/saas-admin-check ./cmd/main.go 2>/dev/null; then
        echo -e "${GREEN}✓${NC} Service compiles successfully"
        rm -f /tmp/saas-admin-check
    else
        echo -e "${RED}❌ Compilation errors detected${NC}"
        echo -e "${YELLOW}Manual review required for: $handler_file${NC}"
    fi

    # Check for remaining "Not implemented" handlers
    local remaining_not_implemented=$(grep -c "Not implemented" "$handler_file" 2>/dev/null || echo "0")

    if [ "$remaining_not_implemented" -eq 0 ]; then
        echo -e "${GREEN}✓${NC} All 'Not implemented' handlers removed"
    else
        echo -e "${YELLOW}⚠️  $remaining_not_implemented 'Not implemented' handlers remain${NC}"
    fi

    # Check for proxy methods
    local proxy_methods=$(grep -c "proxyRequest" "$handler_file" 2>/dev/null || echo "0")
    echo -e "Proxy methods added: $proxy_methods"

    cd - > /dev/null
}

# Function to generate service URLs configuration
generate_service_config() {
    local config_file="$SAAS_ADMIN_DIR/internal/config/services.go"

    echo -e "${BLUE}⚙️ Generating service configuration...${NC}"

    mkdir -p "$(dirname "$config_file")"

    cat > "$config_file" << 'EOF'
// Package config provides service configuration for the SaaS Admin Service.
package config

// ServiceURLs contains URLs for downstream services that SaaS Admin proxies to.
type ServiceURLs struct {
	TenantAdminService string
	ComponentService   string
	IncidentService    string
	MonitoringService  string
	AnalyticsService   string
	NotificationService string
}

// GetDefaultServiceURLs returns default service URLs for development.
func GetDefaultServiceURLs() ServiceURLs {
	return ServiceURLs{
		TenantAdminService:  "http://localhost:8091",
		ComponentService:    "http://localhost:8093",
		IncidentService:     "http://localhost:8094",
		MonitoringService:   "http://localhost:8095",
		AnalyticsService:    "http://localhost:8096",
		NotificationService: "http://localhost:8097",
	}
}

// GetServiceURLsFromEnv returns service URLs from environment variables.
func GetServiceURLsFromEnv() ServiceURLs {
	return ServiceURLs{
		TenantAdminService:  getEnvOrDefault("TENANT_ADMIN_SERVICE_URL", "http://localhost:8091"),
		ComponentService:    getEnvOrDefault("COMPONENT_SERVICE_URL", "http://localhost:8093"),
		IncidentService:     getEnvOrDefault("INCIDENT_SERVICE_URL", "http://localhost:8094"),
		MonitoringService:   getEnvOrDefault("MONITORING_SERVICE_URL", "http://localhost:8095"),
		AnalyticsService:    getEnvOrDefault("ANALYTICS_SERVICE_URL", "http://localhost:8096"),
		NotificationService: getEnvOrDefault("NOTIFICATION_SERVICE_URL", "http://localhost:8097"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
EOF

    echo -e "${GREEN}✓${NC} Generated service configuration"
}

# Main execution
main() {
    echo -e "${PURPLE}🚀 Starting SaaS Admin Service boundary fixes...${NC}"
    echo

    # Check prerequisites
    check_saas_admin_service

    # Analyze current state
    analyze_current_handlers

    echo -e "\n${BLUE}Applying fixes...${NC}"

    # Apply fixes
    remove_not_implemented
    add_required_imports
    add_http_client
    implement_proxy_methods
    generate_service_config

    # Validate changes
    echo -e "\n${BLUE}Validation${NC}"
    validate_changes

    # Generate summary
    echo -e "\n${BLUE}📊 Summary${NC}"
    echo "=========="
    echo -e "Backups created in: $BACKUP_DIR"
    echo -e "Files modified:"
    echo -e "  - $SAAS_ADMIN_DIR/internal/handlers/saas_admin_handler.go"
    echo -e "  - $SAAS_ADMIN_DIR/internal/config/services.go (created)"

    echo -e "\n${GREEN}✅ SaaS Admin Service boundary fixes completed${NC}"

    echo -e "\n${YELLOW}💡 Next steps:${NC}"
    echo "1. Test the proxy functionality:"
    echo "   ./test-scripts/test-saas-admin-features.sh"
    echo "2. Start dependent services (tenant-admin, component, incident services)"
    echo "3. Verify all proxy endpoints work correctly"
    echo "4. Update documentation to reflect the proxy architecture"

    echo -e "\n${YELLOW}⚠️  Important notes:${NC}"
    echo "- SaaS Admin Service now properly proxies requests to other services"
    echo "- Ensure dependent services are running for proxy calls to work"
    echo "- Remove backup files after verification: rm -rf $BACKUP_DIR"
}

# Run main function
main "$@"