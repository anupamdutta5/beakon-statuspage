#!/bin/bash

# Array of all microservices to push
services=(
    "api-gateway"
    "component-service" 
    "incident-service"
    "notification-service"
    "payment-service"
    "database-service"
    "monitoring-service"
    "tenant-admin-service"
    "saas-admin-service"
    "event-store-service"
    "branding-service"
    "landing-page-service"
    "analytics-consumer"
    "audit-consumer"
    "notification-consumer"
    "billing-consumer"
    "status-ui-service"
)

commit_message="feat: enterprise improvements with security enhancements

🔐 Security Enhancements:
- Fixed hardcoded JWT secrets in start.sh scripts
- Added production validation for sensitive environment variables
- Enhanced input sanitization and validation
- Implemented secure CORS configuration

🛠️ Architecture Improvements:
- Added comprehensive shared-resilience module integration
- Implemented graceful shutdown patterns
- Enhanced error handling and logging
- Service-specific improvements and optimizations

📋 Documentation & Development:
- Created comprehensive .env.example file
- Enhanced service documentation
- Improved development workflow

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# Function to push a single service
push_service() {
    local service=$1
    echo "🚀 Pushing $service..."
    
    if [ ! -d "$service" ]; then
        echo "❌ Directory $service not found"
        return 1
    fi
    
    cd "$service" || return 1
    
    if [ ! -d ".git" ]; then
        echo "❌ $service is not a git repository"
        cd ..
        return 1
    fi
    
    # Add all changes
    git add .
    
    # Check if there are changes to commit
    if git diff --cached --quiet; then
        echo "ℹ️  No changes to commit in $service"
        cd ..
        return 0
    fi
    
    # Commit changes
    git commit -m "$commit_message"
    
    # Push to remote
    if git push origin develop; then
        echo "✅ Successfully pushed $service"
    else
        echo "❌ Failed to push $service"
    fi
    
    cd ..
}

# Push all services
for service in "${services[@]}"; do
    push_service "$service"
    echo "---"
done

echo "🎉 All services processed!"
