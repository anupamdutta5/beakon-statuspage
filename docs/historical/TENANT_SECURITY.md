# Tenant Security & Data Isolation System

## Overview

The Tenant Security & Data Isolation System provides comprehensive security measures to ensure complete data isolation between tenants, audit logging, encryption, and compliance monitoring. This system implements enterprise-grade security practices for multi-tenant SaaS applications.

## Features

### 🔒 Data Isolation
- **Row-Level Security**: Automatic tenant isolation for all database queries
- **Tenant Context**: Secure tenant context propagation throughout the application
- **Access Control**: Comprehensive access validation for all tenant operations
- **Data Encryption**: Tenant-specific encryption keys for sensitive data

### 📊 Audit & Compliance
- **Comprehensive Audit Logging**: All user actions and data access logged
- **Security Violation Tracking**: Real-time monitoring of security violations
- **Compliance Reporting**: Detailed compliance metrics and reports
- **Data Access Logging**: Complete audit trail for regulatory compliance

### 🛡️ Security Controls
- **IP Whitelisting**: Restrict access to specific IP addresses
- **Rate Limiting**: Prevent abuse with configurable rate limits
- **Session Management**: Secure session handling with configurable timeouts
- **API Key Management**: Secure API key generation and validation

### 🔐 Encryption & Privacy
- **Data Encryption**: AES-256 encryption for sensitive data
- **Tenant-Specific Keys**: Unique encryption keys per tenant
- **Secure Key Management**: Secure key generation and storage
- **Privacy Controls**: Configurable privacy and data protection settings

## Architecture

### Security Middleware Stack

The security system implements a comprehensive middleware stack:

1. **Security Headers Middleware**: Adds security headers to all responses
2. **Audit Logging Middleware**: Logs all requests for audit purposes
3. **Data Access Logging Middleware**: Logs data access for compliance
4. **Rate Limiting Middleware**: Implements rate limiting per tenant
5. **Tenant Isolation Middleware**: Ensures tenant data isolation
6. **IP Whitelist Middleware**: Validates IP addresses against whitelist

### Data Flow

```
Request → Security Headers → Audit Logging → Data Access Logging → 
Rate Limiting → Tenant Isolation → IP Whitelist → Application Logic
```

## API Endpoints

### Security Configuration

#### Get Security Configuration
```http
GET /api/v1/tenant/security/config
```

Response:
```json
{
  "config": {
    "tenant_id": 1,
    "data_encryption": true,
    "audit_logging": true,
    "ip_whitelist": ["192.168.1.0/24"],
    "session_timeout": 60,
    "password_policy": "strong",
    "two_factor_auth": false,
    "api_key_rotation": 90,
    "last_security_update": "2024-01-15T10:30:00Z"
  }
}
```

#### Update Security Configuration
```http
PUT /api/v1/tenant/security/config
Content-Type: application/json

{
  "data_encryption": true,
  "audit_logging": true,
  "two_factor_auth": true,
  "session_timeout": 120,
  "password_policy": "strong",
  "api_key_rotation": 60
}
```

### Security Metrics

#### Get Security Metrics
```http
GET /api/v1/tenant/security/metrics?days=30
```

Response:
```json
{
  "metrics": {
    "tenant_id": 1,
    "period_days": 30,
    "audit_events": {
      "total": 1250,
      "successful": 1180,
      "failed": 70,
      "by_action": {
        "login": 450,
        "data_access": 320,
        "configuration_change": 180,
        "user_management": 150,
        "other": 150
      }
    },
    "security_violations": {
      "total": 12,
      "by_severity": {
        "low": 8,
        "medium": 3,
        "high": 1,
        "critical": 0
      },
      "by_type": {
        "unauthorized_access": 5,
        "suspicious_activity": 4,
        "data_breach": 2,
        "other": 1
      }
    },
    "data_access": {
      "total_requests": 3450,
      "unique_users": 25,
      "by_table": {
        "incidents": 1200,
        "services": 800,
        "users": 600,
        "settings": 450,
        "other": 400
      }
    },
    "compliance": {
      "data_encryption_enabled": true,
      "audit_logging_enabled": true,
      "two_factor_auth_enabled": false,
      "ip_whitelist_configured": true,
      "last_security_audit": "2024-01-08T10:30:00Z"
    }
  }
}
```

### Audit Logs

#### Get Audit Logs
```http
GET /api/v1/tenant/security/audit-logs?page=1&limit=50&action=GET&success=true
```

Response:
```json
{
  "logs": [
    {
      "id": 1,
      "tenant_id": 1,
      "user_id": 1,
      "action": "GET",
      "resource": "/api/v1/tenant/incidents",
      "details": "Status: 200, Duration: 45ms",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
      "timestamp": "2024-01-15T10:30:00Z",
      "success": true
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 1250,
    "total_pages": 25
  }
}
```

### Security Violations

#### Get Security Violations
```http
GET /api/v1/tenant/security/violations?severity=high&resolved=false
```

Response:
```json
{
  "violations": [
    {
      "id": 1,
      "tenant_id": 1,
      "type": "unauthorized_access",
      "severity": "high",
      "description": "Unauthorized access attempt to admin panel",
      "ip_address": "192.168.1.200",
      "user_agent": "curl/7.68.0",
      "details": "Multiple failed login attempts",
      "timestamp": "2024-01-15T09:15:00Z",
      "resolved": false
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 12,
    "total_pages": 1
  }
}
```

#### Resolve Security Violation
```http
POST /api/v1/tenant/security/violations/1/resolve
```

### API Key Management

#### Generate API Key
```http
POST /api/v1/tenant/security/api-keys
Content-Type: application/json

{
  "name": "Production API Key"
}
```

Response:
```json
{
  "message": "API key generated successfully",
  "api_key": "ak_live_1234567890abcdef",
  "name": "Production API Key",
  "warning": "Store this API key securely. It will not be shown again."
}
```

#### Validate API Key
```http
POST /api/v1/tenant/security/api-keys/validate
Content-Type: application/json

{
  "api_key": "ak_live_1234567890abcdef"
}
```

### Data Encryption

#### Encrypt Data
```http
POST /api/v1/tenant/security/encrypt
Content-Type: application/json

{
  "data": "sensitive information"
}
```

Response:
```json
{
  "encrypted_data": "base64_encoded_encrypted_data"
}
```

#### Decrypt Data
```http
POST /api/v1/tenant/security/decrypt
Content-Type: application/json

{
  "encrypted_data": "base64_encoded_encrypted_data"
}
```

Response:
```json
{
  "data": "sensitive information"
}
```

## Frontend Interface

### Security Dashboard

The security dashboard provides comprehensive security management:

1. **Security Overview**
   - Real-time security metrics
   - Compliance status indicators
   - Security violation alerts
   - Data access statistics

2. **Security Configuration**
   - Data encryption settings
   - Audit logging configuration
   - Two-factor authentication
   - IP whitelist management
   - Session timeout settings
   - Password policy configuration

3. **Audit Logs**
   - Comprehensive audit trail
   - Filterable by action, user, and status
   - Real-time log monitoring
   - Export capabilities

4. **Security Violations**
   - Real-time violation monitoring
   - Severity-based filtering
   - Resolution tracking
   - Alert notifications

5. **API Key Management**
   - Secure API key generation
   - Key rotation management
   - Usage monitoring
   - Revocation capabilities

6. **Data Encryption Tools**
   - Encrypt sensitive data
   - Decrypt encrypted data
   - Copy to clipboard functionality
   - Secure data handling

## Security Features

### Data Isolation

#### Row-Level Security
All database queries are automatically scoped to the current tenant:

```go
// Automatic tenant isolation
func (s *TenantSecurityService) GetTenantDataScope(tenantID uint) func(*gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("tenant_id = ?", tenantID)
    }
}
```

#### Tenant Context
Secure tenant context propagation:

```go
// Set tenant in context
func SetTenantInContext(c *gin.Context, tenant *models.Tenant) {
    c.Set("tenant", tenant)
}

// Get tenant from context
func GetTenantFromContext(c *gin.Context) (*models.Tenant, bool) {
    tenant, exists := c.Get("tenant")
    if !exists {
        return nil, false
    }
    tenantObj, ok := tenant.(*models.Tenant)
    return tenantObj, ok
}
```

### Encryption

#### Tenant-Specific Encryption
Each tenant has unique encryption keys:

```go
// Generate tenant-specific encryption key
func (s *TenantSecurityService) generateTenantKey(tenantID uint) []byte {
    hash := sha256.Sum256([]byte(fmt.Sprintf("tenant_key_%d", tenantID)))
    return hash[:]
}

// Encrypt data with tenant key
func (s *TenantSecurityService) EncryptData(ctx context.Context, tenantID uint, data string) (string, error) {
    key := s.generateTenantKey(tenantID)
    // AES-256-GCM encryption implementation
}
```

### Audit Logging

#### Comprehensive Logging
All user actions are logged:

```go
// Audit log structure
type SecurityAuditLog struct {
    ID        uint      `json:"id"`
    TenantID  uint      `json:"tenant_id"`
    UserID    uint      `json:"user_id,omitempty"`
    Action    string    `json:"action"`
    Resource  string    `json:"resource"`
    Details   string    `json:"details"`
    IPAddress string    `json:"ip_address"`
    UserAgent string    `json:"user_agent"`
    Timestamp time.Time `json:"timestamp"`
    Success   bool      `json:"success"`
}
```

### Rate Limiting

#### Per-Tenant Rate Limiting
Configurable rate limits per tenant:

```go
// Rate limiting implementation
func (m *SecurityMiddleware) RateLimitingMiddleware() gin.HandlerFunc {
    rateLimiter := make(map[string][]time.Time)
    
    return func(c *gin.Context) {
        tenant, _ := GetTenantFromContext(c)
        key := fmt.Sprintf("tenant_%d_%s", tenant.ID, c.ClientIP())
        
        // Check rate limit (100 requests per minute)
        if len(rateLimiter[key]) >= 100 {
            // Log violation and return error
        }
    }
}
```

## Security Best Practices

### Configuration

1. **Enable Data Encryption**: Always enable data encryption for sensitive data
2. **Configure Audit Logging**: Enable comprehensive audit logging
3. **Set Strong Password Policies**: Use strong password requirements
4. **Enable Two-Factor Authentication**: Require 2FA for admin users
5. **Configure IP Whitelisting**: Restrict access to known IP addresses
6. **Set Appropriate Session Timeouts**: Balance security and usability

### Monitoring

1. **Regular Security Reviews**: Review security metrics regularly
2. **Monitor Violations**: Set up alerts for security violations
3. **Audit Log Analysis**: Regularly analyze audit logs for anomalies
4. **API Key Rotation**: Rotate API keys regularly
5. **Compliance Monitoring**: Monitor compliance status continuously

### Data Protection

1. **Encrypt Sensitive Data**: Always encrypt sensitive information
2. **Secure Key Management**: Use secure key storage and rotation
3. **Data Access Controls**: Implement strict data access controls
4. **Privacy Controls**: Configure appropriate privacy settings
5. **Data Retention**: Implement appropriate data retention policies

## Compliance

### Regulatory Compliance

The security system supports compliance with:

- **GDPR**: Data protection and privacy controls
- **SOC 2**: Security and availability controls
- **HIPAA**: Healthcare data protection
- **PCI DSS**: Payment card data security
- **ISO 27001**: Information security management

### Audit Trail

Complete audit trail includes:

- User authentication and authorization
- Data access and modifications
- Configuration changes
- Security violations and responses
- API usage and access patterns
- System events and errors

### Reporting

Comprehensive reporting capabilities:

- Security metrics and trends
- Compliance status reports
- Violation analysis and resolution
- Data access patterns
- User activity summaries
- System security posture

## Troubleshooting

### Common Issues

#### Access Denied Errors
- Check tenant context is properly set
- Verify user belongs to the tenant
- Confirm IP address is whitelisted
- Check rate limiting status

#### Encryption/Decryption Failures
- Verify tenant ID is correct
- Check encryption key generation
- Confirm data format is valid
- Validate base64 encoding

#### Audit Log Issues
- Check audit logging is enabled
- Verify database connectivity
- Confirm log retention settings
- Check log format and structure

### Security Violations

#### Unauthorized Access
- Review IP whitelist configuration
- Check user permissions
- Verify tenant isolation
- Monitor for brute force attempts

#### Rate Limiting
- Adjust rate limit thresholds
- Check for legitimate high usage
- Monitor for abuse patterns
- Configure appropriate limits

#### Data Access Violations
- Review data access patterns
- Check user permissions
- Verify tenant isolation
- Monitor for data exfiltration

## Integration

### Database Integration

The security system integrates with:

- **PostgreSQL**: Primary database with row-level security
- **Redis**: Rate limiting and session storage
- **Elasticsearch**: Audit log storage and search
- **Prometheus**: Security metrics collection

### External Services

Integration with external security services:

- **SIEM Systems**: Security information and event management
- **Identity Providers**: SSO and identity management
- **Certificate Authorities**: SSL/TLS certificate management
- **Security Scanners**: Vulnerability assessment tools

## Performance Considerations

### Optimization

1. **Database Indexing**: Proper indexing for audit logs and violations
2. **Caching**: Cache security configurations and metrics
3. **Rate Limiting**: Efficient in-memory rate limiting
4. **Log Rotation**: Automated log rotation and archival
5. **Metrics Collection**: Efficient metrics collection and storage

### Scalability

1. **Horizontal Scaling**: Support for multiple application instances
2. **Database Sharding**: Tenant-based database sharding
3. **Load Balancing**: Secure load balancing with tenant awareness
4. **Caching Strategy**: Distributed caching for security data
5. **Message Queues**: Asynchronous processing for audit logs

## Future Enhancements

Planned security enhancements:

- **Advanced Threat Detection**: Machine learning-based threat detection
- **Behavioral Analytics**: User behavior analysis and anomaly detection
- **Zero Trust Architecture**: Implement zero trust security model
- **Advanced Encryption**: Post-quantum cryptography support
- **Security Orchestration**: Automated security response and remediation
- **Compliance Automation**: Automated compliance monitoring and reporting
