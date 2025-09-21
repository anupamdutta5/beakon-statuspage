#!/bin/bash
services=(notification-service payment-service database-service monitoring-service tenant-admin-service saas-admin-service event-store-service branding-service landing-page-service analytics-consumer audit-consumer notification-consumer billing-consumer status-ui-service)

for service in "${services[@]}"; do
  echo "🚀 Pushing $service..."
  cd "$service" && git add . && git commit -m "feat: enterprise improvements with security enhancements

🔐 Security Enhancements:
- Fixed hardcoded JWT secrets in start.sh scripts
- Added production validation for sensitive environment variables
- Enhanced input sanitization and validation
- Implemented secure CORS configuration

🛠️ Architecture Improvements:
- Added comprehensive shared-resilience module integration
- Implemented graceful shutdown patterns
- Enhanced error handling and logging

📋 Documentation & Development:
- Created comprehensive .env.example file
- Enhanced service documentation

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>" && git push origin develop && echo "✅ $service pushed" && cd ..
done
