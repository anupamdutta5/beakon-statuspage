#!/bin/bash
# Generate config.test.yml for all services that need it

set -e

cd "$(dirname "$0")/.."

# Function to get port for a service
get_port() {
    case "$1" in
        "analytics-service") echo "8090" ;;
        "api-gateway") echo "8080" ;;
        "branding-service") echo "8097" ;;
        "component-service") echo "8084" ;;
        "database-service") echo "8095" ;;
        "event-store-service") echo "8096" ;;
        "incident-service") echo "8086" ;;
        "landing-page-service") echo "8100" ;;
        "monitoring-service") echo "8092" ;;
        "notification-service") echo "8085" ;;
        "payment-service") echo "8088" ;;
        "saas-admin-service") echo "8098" ;;
        "status-ui-service") echo "8093" ;;
        "tenant-admin-service") echo "8099" ;;
        "user-service") echo "8081" ;;
        *) echo "8000" ;;  # Default port for consumers
    esac
}

echo "Generating config.test.yml files for all services..."

for svc_dir in microservices/*/; do
    svc_name=$(basename "$svc_dir")

    # Skip if no configs directory
    [ ! -d "$svc_dir/configs" ] && continue

    # Skip if config.test.yml already exists
    if [ -f "$svc_dir/configs/config.test.yml" ]; then
        echo "  ✓ $svc_name (already exists)"
        continue
    fi

    # Get port for this service
    port=$(get_port "$svc_name")

    # Generate config from template
    sed -e "s/{{SERVICE_NAME}}/$svc_name/g" \
        -e "s/{{SERVICE_PORT}}/$port/g" \
        config.test.template.yml > "$svc_dir/configs/config.test.yml"

    echo "  ✅ Generated config.test.yml for $svc_name"
done

echo ""
echo "✅ Done! Generated test configs for all services."
