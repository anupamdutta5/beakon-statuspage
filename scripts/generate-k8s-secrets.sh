#!/bin/bash

# Generate Kubernetes secrets from environment variables
# This script helps convert environment variables to Kubernetes secrets

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to base64 encode a value
encode_secret() {
    local value="$1"
    if [ -z "$value" ]; then
        print_error "Value is empty for encoding"
        return 1
    fi
    echo -n "$value" | base64 -w 0
}

# Function to check if environment variable exists
check_env_var() {
    local var_name="$1"
    if [ -z "${!var_name}" ]; then
        print_warning "Environment variable $var_name is not set"
        return 1
    fi
    return 0
}

# Function to generate secrets YAML
generate_secrets_yaml() {
    local output_file="$1"
    local namespace="${2:-statuspage}"
    
    print_status "Generating Kubernetes secrets YAML..."
    
    cat > "$output_file" << EOF
apiVersion: v1
kind: Secret
metadata:
  name: statuspage-secrets
  namespace: $namespace
type: Opaque
data:
EOF

    # Database secrets
    if check_env_var "DB_PASSWORD"; then
        echo "  db-password: $(encode_secret "$DB_PASSWORD")" >> "$output_file"
    fi
    
    if check_env_var "DB_USER"; then
        echo "  db-user: $(encode_secret "$DB_USER")" >> "$output_file"
    fi

    # JWT secrets
    if check_env_var "JWT_SECRET"; then
        echo "  jwt-secret: $(encode_secret "$JWT_SECRET")" >> "$output_file"
    fi

    # Email secrets
    if check_env_var "SMTP_PASSWORD"; then
        echo "  smtp-password: $(encode_secret "$SMTP_PASSWORD")" >> "$output_file"
    fi
    
    if check_env_var "SMTP_USERNAME"; then
        echo "  smtp-username: $(encode_secret "$SMTP_USERNAME")" >> "$output_file"
    fi
    
    if check_env_var "SENDGRID_API_KEY"; then
        echo "  sendgrid-api-key: $(encode_secret "$SENDGRID_API_KEY")" >> "$output_file"
    fi

    # Payment secrets
    if check_env_var "STRIPE_SECRET_KEY"; then
        echo "  stripe-secret-key: $(encode_secret "$STRIPE_SECRET_KEY")" >> "$output_file"
    fi
    
    if check_env_var "STRIPE_WEBHOOK_SECRET"; then
        echo "  stripe-webhook-secret: $(encode_secret "$STRIPE_WEBHOOK_SECRET")" >> "$output_file"
    fi
    
    if check_env_var "STRIPE_PUBLISHABLE_KEY"; then
        echo "  stripe-publishable-key: $(encode_secret "$STRIPE_PUBLISHABLE_KEY")" >> "$output_file"
    fi
    
    if check_env_var "PAYPAL_CLIENT_ID"; then
        echo "  paypal-client-id: $(encode_secret "$PAYPAL_CLIENT_ID")" >> "$output_file"
    fi
    
    if check_env_var "PAYPAL_CLIENT_SECRET"; then
        echo "  paypal-client-secret: $(encode_secret "$PAYPAL_CLIENT_SECRET")" >> "$output_file"
    fi

    # Security secrets
    if check_env_var "ENCRYPTION_KEY"; then
        echo "  encryption-key: $(encode_secret "$ENCRYPTION_KEY")" >> "$output_file"
    fi
    
    if check_env_var "SESSION_SECRET"; then
        echo "  session-secret: $(encode_secret "$SESSION_SECRET")" >> "$output_file"
    fi

    # Monitoring secrets
    if check_env_var "INFLUX_TOKEN"; then
        echo "  influx-token: $(encode_secret "$INFLUX_TOKEN")" >> "$output_file"
    fi

    # Service discovery secrets
    if check_env_var "CONSUL_TOKEN"; then
        echo "  consul-token: $(encode_secret "$CONSUL_TOKEN")" >> "$output_file"
    fi

    # Redis secrets
    if check_env_var "REDIS_PASSWORD"; then
        echo "  redis-password: $(encode_secret "$REDIS_PASSWORD")" >> "$output_file"
    fi

    # TLS secrets (if provided)
    if [ -n "$TLS_CERT" ] && [ -n "$TLS_KEY" ]; then
        cat >> "$output_file" << EOF

---
apiVersion: v1
kind: Secret
metadata:
  name: statuspage-tls
  namespace: $namespace
type: kubernetes.io/tls
data:
  tls.crt: $(encode_secret "$TLS_CERT")
  tls.key: $(encode_secret "$TLS_KEY")
EOF
    fi

    print_status "Secrets YAML generated: $output_file"
}

# Function to generate environment file template
generate_env_template() {
    local output_file="$1"
    
    print_status "Generating environment variables template..."
    
    cat > "$output_file" << 'EOF'
# Status Page Environment Variables Template
# Copy this file to .env and fill in your actual values

# Database Configuration
DB_PASSWORD=your_database_password_here
DB_USER=your_database_user_here

# JWT Configuration
JWT_SECRET=your_jwt_secret_key_here

# Email Configuration
SMTP_PASSWORD=your_smtp_password_here
SMTP_USERNAME=your_smtp_username_here
SENDGRID_API_KEY=your_sendgrid_api_key_here

# Payment Configuration
STRIPE_SECRET_KEY=sk_live_your_stripe_secret_key_here
STRIPE_WEBHOOK_SECRET=whsec_your_stripe_webhook_secret_here
STRIPE_PUBLISHABLE_KEY=pk_live_your_stripe_publishable_key_here
PAYPAL_CLIENT_ID=your_paypal_client_id_here
PAYPAL_CLIENT_SECRET=your_paypal_client_secret_here

# Security Configuration
ENCRYPTION_KEY=your_32_character_encryption_key_here
SESSION_SECRET=your_session_secret_here

# Monitoring Configuration
INFLUX_TOKEN=your_influxdb_token_here

# Service Discovery Configuration
CONSUL_TOKEN=your_consul_token_here

# Redis Configuration
REDIS_PASSWORD=your_redis_password_here

# TLS Configuration (optional)
# TLS_CERT=your_tls_certificate_here
# TLS_KEY=your_tls_private_key_here
EOF

    print_status "Environment template generated: $output_file"
}

# Function to validate secrets
validate_secrets() {
    local secrets_file="$1"
    
    print_status "Validating secrets file..."
    
    if [ ! -f "$secrets_file" ]; then
        print_error "Secrets file not found: $secrets_file"
        return 1
    fi
    
    # Check if file contains any secrets
    if ! grep -q "data:" "$secrets_file"; then
        print_warning "No secrets found in file"
        return 1
    fi
    
    # Count secrets
    local secret_count=$(grep -c ":" "$secrets_file" | grep -v "kind:" | grep -v "metadata:" | grep -v "name:" | grep -v "namespace:" | grep -v "type:" || echo "0")
    
    print_status "Found $secret_count secrets in file"
    
    # Validate base64 encoding
    local invalid_secrets=0
    while IFS= read -r line; do
        if [[ $line =~ ^[[:space:]]*[a-zA-Z0-9_-]+:[[:space:]]*[A-Za-z0-9+/=]+$ ]]; then
            local encoded_value=$(echo "$line" | cut -d':' -f2 | tr -d ' ')
            if ! echo "$encoded_value" | base64 -d >/dev/null 2>&1; then
                print_error "Invalid base64 encoding in line: $line"
                ((invalid_secrets++))
            fi
        fi
    done < "$secrets_file"
    
    if [ $invalid_secrets -eq 0 ]; then
        print_status "All secrets are properly base64 encoded"
        return 0
    else
        print_error "Found $invalid_secrets invalid secrets"
        return 1
    fi
}

# Function to apply secrets to Kubernetes
apply_secrets() {
    local secrets_file="$1"
    local namespace="${2:-statuspage}"
    
    print_status "Applying secrets to Kubernetes..."
    
    # Check if kubectl is available
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed or not in PATH"
        return 1
    fi
    
    # Check if namespace exists
    if ! kubectl get namespace "$namespace" &> /dev/null; then
        print_status "Creating namespace: $namespace"
        kubectl create namespace "$namespace"
    fi
    
    # Apply secrets
    if kubectl apply -f "$secrets_file"; then
        print_status "Secrets applied successfully to namespace: $namespace"
        return 0
    else
        print_error "Failed to apply secrets"
        return 1
    fi
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -g, --generate FILE     Generate secrets YAML from environment variables"
    echo "  -t, --template FILE     Generate environment variables template"
    echo "  -v, --validate FILE     Validate secrets YAML file"
    echo "  -a, --apply FILE        Apply secrets to Kubernetes cluster"
    echo "  -n, --namespace NS      Kubernetes namespace (default: statuspage)"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --generate k8s/secrets/generated-secrets.yaml"
    echo "  $0 --template .env.template"
    echo "  $0 --validate k8s/secrets/statuspage-secrets.yaml"
    echo "  $0 --apply k8s/secrets/statuspage-secrets.yaml --namespace production"
}

# Main script logic
main() {
    local generate_file=""
    local template_file=""
    local validate_file=""
    local apply_file=""
    local namespace="statuspage"
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -g|--generate)
                generate_file="$2"
                shift 2
                ;;
            -t|--template)
                template_file="$2"
                shift 2
                ;;
            -v|--validate)
                validate_file="$2"
                shift 2
                ;;
            -a|--apply)
                apply_file="$2"
                shift 2
                ;;
            -n|--namespace)
                namespace="$2"
                shift 2
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # Execute requested operations
    if [ -n "$generate_file" ]; then
        generate_secrets_yaml "$generate_file" "$namespace"
    fi
    
    if [ -n "$template_file" ]; then
        generate_env_template "$template_file"
    fi
    
    if [ -n "$validate_file" ]; then
        validate_secrets "$validate_file"
    fi
    
    if [ -n "$apply_file" ]; then
        apply_secrets "$apply_file" "$namespace"
    fi
    
    # If no options provided, show usage
    if [ -z "$generate_file" ] && [ -z "$template_file" ] && [ -z "$validate_file" ] && [ -z "$apply_file" ]; then
        show_usage
    fi
}

# Run main function
main "$@"
