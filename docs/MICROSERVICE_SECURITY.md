# Microservice Security Guide

This document describes the comprehensive security implementation for the Status Page microservices architecture, including JWT authentication, OAuth2 integration, mTLS, and security middleware.

## Overview

The microservice security system provides comprehensive security features including authentication, authorization, encryption, and protection against common security vulnerabilities. It implements industry-standard security patterns and best practices for microservices architectures.

## Architecture

### Security Components

1. **JWT Authentication** - Token-based authentication with RSA signing
2. **OAuth2 Integration** - Support for multiple OAuth2 providers
3. **mTLS (Mutual TLS)** - Certificate-based authentication between services
4. **Security Middleware** - Comprehensive security middleware stack
5. **Rate Limiting** - Protection against abuse and DoS attacks
6. **CORS** - Cross-origin resource sharing configuration
7. **Security Headers** - HTTP security headers for protection

### Security Stack

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client        │    │   API Gateway   │    │   Microservices │
│   (Browser/App) │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   JWT/OAuth2    │    │   Security      │    │   mTLS          │
│   Authentication│    │   Middleware    │    │   Authentication│
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Rate Limiting │    │   CORS          │    │   Authorization │
│   & Headers     │    │   & Headers     │    │   & Validation  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## JWT Authentication

### Configuration

```go
config := security.JWTConfig{
    SecretKey:     "your-secret-key",
    PublicKey:     "-----BEGIN PUBLIC KEY-----\n...",
    PrivateKey:    "-----BEGIN RSA PRIVATE KEY-----\n...",
    Expiration:    15 * time.Minute,
    RefreshExpiry: 7 * 24 * time.Hour,
    Issuer:        "statuspage",
    Audience:      "statuspage-api",
    Algorithm:     "RS256",
}

jwtManager, err := security.NewJWTManager(config)
```

### Token Generation

```go
// Generate access token
claims := security.JWTClaims{
    UserID:    "user123",
    TenantID:  "tenant456",
    Email:     "user@example.com",
    Roles:     []string{"admin", "user"},
    Scopes:    []string{"read", "write"},
    SessionID: "session789",
}

token, err := jwtManager.GenerateToken(claims)
```

### Token Validation

```go
// Validate token
claims, err := jwtManager.ValidateToken(tokenString)
if err != nil {
    // Handle invalid token
}

// Access claims
userID := claims.UserID
tenantID := claims.TenantID
roles := claims.Roles
```

### Refresh Token

```go
// Generate refresh token
refreshToken, err := jwtManager.GenerateRefreshToken("user123", "session789")

// Refresh access token
newClaims := security.JWTClaims{
    Email:  "user@example.com",
    Roles:  []string{"admin", "user"},
    Scopes: []string{"read", "write"},
}

newToken, err := jwtManager.RefreshToken(refreshToken, newClaims)
```

## OAuth2 Integration

### Provider Configuration

```go
// Google OAuth2
googleConfig := security.OAuth2Config{
    ClientID:     "your-google-client-id",
    ClientSecret: "your-google-client-secret",
    RedirectURL:  "https://your-app.com/auth/google/callback",
    Scopes:       []string{"openid", "email", "profile"},
    AuthURL:      "https://accounts.google.com/o/oauth2/auth",
    TokenURL:     "https://oauth2.googleapis.com/token",
    UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
    Provider:     "google",
}

// GitHub OAuth2
githubConfig := security.OAuth2Config{
    ClientID:     "your-github-client-id",
    ClientSecret: "your-github-client-secret",
    RedirectURL:  "https://your-app.com/auth/github/callback",
    Scopes:       []string{"user:email"},
    AuthURL:      "https://github.com/login/oauth/authorize",
    TokenURL:     "https://github.com/login/oauth/access_token",
    UserInfoURL:  "https://api.github.com/user",
    Provider:     "github",
}
```

### OAuth2 Flow

```go
// Initialize OAuth2 manager
oauth2Manager := security.NewOAuth2Manager()

// Register providers
oauth2Manager.RegisterProvider("google", googleConfig)
oauth2Manager.RegisterProvider("github", githubConfig)

// Generate authorization URL
authURL, state, err := oauth2Manager.GenerateAuthURL("google", "/dashboard")

// Handle callback
token, userInfo, err := oauth2Manager.HandleCallback("google", code, state)
```

### OAuth2 Middleware

```go
// OAuth2 authentication handler
authHandler := oauth2Middleware.AuthHandler("google")

// OAuth2 callback handler
callbackHandler := oauth2Middleware.CallbackHandler("google")

// Register routes
http.HandleFunc("/auth/google", authHandler)
http.HandleFunc("/auth/google/callback", callbackHandler)
```

## mTLS (Mutual TLS)

### Configuration

```go
config := security.mTLSConfig{
    CAKeyPath:     "/certs/ca-key.pem",
    CACertPath:    "/certs/ca-cert.pem",
    ServerKeyPath: "/certs/server-key.pem",
    ServerCertPath: "/certs/server-cert.pem",
    ClientKeyPath: "/certs/client-key.pem",
    ClientCertPath: "/certs/client-cert.pem",
    CertValidity:  365 * 24 * time.Hour,
    KeySize:       2048,
}

mTLSManager, err := security.NewmTLSManager(config)
```

### Server TLS Configuration

```go
// Get server TLS configuration
tlsConfig, err := mTLSManager.GetServerTLSConfig("api.statuspage.com")

// Start HTTPS server
server := &http.Server{
    Addr:      ":443",
    TLSConfig: tlsConfig,
    Handler:   router,
}

server.ListenAndServeTLS("", "")
```

### Client TLS Configuration

```go
// Get client TLS configuration
tlsConfig, err := mTLSManager.GetClientTLSConfig("client123")

// Create HTTP client with mTLS
client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: tlsConfig,
    },
}
```

### mTLS Middleware

```go
// mTLS middleware
mTLSMiddleware := security.NewmTLSMiddleware(mTLSManager)

// Apply mTLS middleware
router.Use(mTLSMiddleware.ClientCertHandler)
```

## Security Middleware

### Configuration

```go
config := security.SecurityConfig{
    JWTConfig: jwtConfig,
    OAuth2Configs: map[string]security.OAuth2Config{
        "google": googleConfig,
        "github": githubConfig,
    },
    mTLSConfig: mTLSConfig,
    RateLimit: security.RateLimitConfig{
        Enabled:     true,
        Requests:    100,
        Window:      1 * time.Minute,
        Burst:       10,
        SkipSuccess: false,
    },
    CORS: security.CORSConfig{
        Enabled:          true,
        AllowedOrigins:   []string{"https://app.statuspage.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders:   []string{"Authorization", "Content-Type"},
        ExposedHeaders:   []string{"X-Total-Count"},
        AllowCredentials: true,
        MaxAge:           86400,
    },
    Headers: security.HeadersConfig{
        Enabled:                true,
        ContentTypeNosniff:     true,
        FrameOptions:           "DENY",
        XSSProtection:          "1; mode=block",
        ContentSecurityPolicy:  "default-src 'self'",
        ReferrerPolicy:         "strict-origin-when-cross-origin",
        PermissionsPolicy:      "geolocation=(), microphone=(), camera=()",
        StrictTransportSecurity: "max-age=31536000; includeSubDomains",
    },
}

securityMiddleware, err := security.NewSecurityMiddleware(config)
```

### Middleware Stack

```go
// Apply security middleware
router.Use(securityMiddleware.SecurityHeadersMiddleware())
router.Use(securityMiddleware.CORSMiddleware())
router.Use(securityMiddleware.RateLimitMiddleware())
router.Use(securityMiddleware.JWTAuthMiddleware())
router.Use(securityMiddleware.TenantAuthMiddleware())
```

### Role-Based Authorization

```go
// Admin only endpoint
router.GET("/admin/users", 
    securityMiddleware.RoleAuthMiddleware("admin"),
    getUsersHandler,
)

// Multiple roles
router.GET("/reports", 
    securityMiddleware.RoleAuthMiddleware("admin", "manager"),
    getReportsHandler,
)
```

### Scope-Based Authorization

```go
// Read scope required
router.GET("/users", 
    securityMiddleware.ScopeAuthMiddleware("read:users"),
    getUsersHandler,
)

// Write scope required
router.POST("/users", 
    securityMiddleware.ScopeAuthMiddleware("write:users"),
    createUserHandler,
)
```

### Tenant-Based Authorization

```go
// Tenant isolation
router.GET("/tenants/:tenantId/users", 
    securityMiddleware.TenantAuthMiddleware(),
    getTenantUsersHandler,
)
```

## Security Best Practices

### JWT Security

1. **Use Strong Keys** - Use RSA 2048-bit or higher keys
2. **Short Expiration** - Use short expiration times (15 minutes)
3. **Refresh Tokens** - Use refresh tokens for long-term access
4. **Secure Storage** - Store tokens securely (httpOnly cookies)
5. **Token Revocation** - Implement token revocation for security

### OAuth2 Security

1. **State Parameter** - Always use state parameter for CSRF protection
2. **Secure Redirects** - Validate redirect URLs
3. **Scope Limitation** - Request only necessary scopes
4. **Token Validation** - Validate tokens with provider
5. **Secure Storage** - Store tokens securely

### mTLS Security

1. **Strong Certificates** - Use strong certificate authorities
2. **Certificate Validation** - Validate all certificates
3. **Key Rotation** - Implement regular key rotation
4. **Secure Storage** - Store certificates securely
5. **Certificate Pinning** - Use certificate pinning for clients

### General Security

1. **HTTPS Everywhere** - Use HTTPS for all communications
2. **Security Headers** - Implement comprehensive security headers
3. **Rate Limiting** - Implement rate limiting for all endpoints
4. **Input Validation** - Validate all input data
5. **Error Handling** - Don't expose sensitive information in errors

## Security Monitoring

### Security Events

```go
// Log security events
logger.LogSecurityEvent("failed_login", "high", map[string]interface{}{
    "ip_address": "192.168.1.1",
    "user_agent": "Mozilla/5.0...",
    "reason":     "invalid_password",
})

logger.LogSecurityEvent("token_validation_failed", "medium", map[string]interface{}{
    "token_type": "jwt",
    "reason":     "expired",
})
```

### Security Metrics

```go
// Record security metrics
metrics.RecordSecurityEvent("failed_login", "high")
metrics.RecordSecurityEvent("token_validation_failed", "medium")
metrics.RecordSecurityEvent("rate_limit_exceeded", "low")
```

### Security Alerts

```yaml
# Prometheus alert rules
groups:
- name: security.rules
  rules:
  - alert: HighFailedLoginRate
    expr: rate(security_events_total{event_type="failed_login"}[5m]) > 10
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "High failed login rate detected"
      description: "Failed login rate is {{ $value }} per second"

  - alert: TokenValidationFailures
    expr: rate(security_events_total{event_type="token_validation_failed"}[5m]) > 5
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High token validation failure rate"
      description: "Token validation failure rate is {{ $value }} per second"
```

## Security Testing

### Unit Tests

```go
func TestJWTManager(t *testing.T) {
    config := security.JWTConfig{
        SecretKey:  "test-secret",
        Expiration: 15 * time.Minute,
        Issuer:     "test",
        Audience:   "test-api",
    }

    manager, err := security.NewJWTManager(config)
    assert.NoError(t, err)

    claims := security.JWTClaims{
        UserID:   "user123",
        TenantID: "tenant456",
        Email:    "user@example.com",
        Roles:    []string{"admin"},
    }

    token, err := manager.GenerateToken(claims)
    assert.NoError(t, err)

    validatedClaims, err := manager.ValidateToken(token)
    assert.NoError(t, err)
    assert.Equal(t, claims.UserID, validatedClaims.UserID)
}
```

### Integration Tests

```go
func TestSecurityMiddleware(t *testing.T) {
    config := security.SecurityConfig{
        JWTConfig: jwtConfig,
        RateLimit: security.RateLimitConfig{
            Enabled:  true,
            Requests: 10,
            Window:   1 * time.Minute,
        },
    }

    middleware, err := security.NewSecurityMiddleware(config)
    assert.NoError(t, err)

    // Test JWT authentication
    // Test rate limiting
    // Test CORS
    // Test security headers
}
```

### Security Scanning

```bash
# OWASP ZAP security scan
docker run -t owasp/zap2docker-stable zap-baseline.py -t https://api.statuspage.com

# SSL/TLS testing
testssl.sh https://api.statuspage.com

# Security headers check
curl -I https://api.statuspage.com
```

## Troubleshooting

### Common Issues

#### 1. JWT Token Validation Failed

```bash
# Check token format
echo "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." | base64 -d

# Check token expiration
jwt decode "your-token-here"

# Check issuer and audience
jwt decode "your-token-here" | jq '.iss, .aud'
```

#### 2. OAuth2 Callback Issues

```bash
# Check OAuth2 configuration
curl -X POST https://oauth2.googleapis.com/token \
  -d "client_id=your-client-id" \
  -d "client_secret=your-client-secret" \
  -d "code=authorization-code" \
  -d "grant_type=authorization_code" \
  -d "redirect_uri=your-redirect-uri"
```

#### 3. mTLS Certificate Issues

```bash
# Check certificate validity
openssl x509 -in server-cert.pem -text -noout

# Check certificate chain
openssl verify -CAfile ca-cert.pem server-cert.pem

# Test mTLS connection
openssl s_client -connect api.statuspage.com:443 -cert client-cert.pem -key client-key.pem -CAfile ca-cert.pem
```

#### 4. Rate Limiting Issues

```bash
# Check rate limit headers
curl -I https://api.statuspage.com/users

# Test rate limiting
for i in {1..20}; do curl https://api.statuspage.com/users; done
```

### Debugging Commands

```bash
# Check JWT token
jwt decode "your-token-here"

# Check OAuth2 token
curl -H "Authorization: Bearer your-token" https://www.googleapis.com/oauth2/v2/userinfo

# Check mTLS connection
openssl s_client -connect api.statuspage.com:443 -cert client-cert.pem -key client-key.pem

# Check security headers
curl -I https://api.statuspage.com

# Check CORS
curl -H "Origin: https://app.statuspage.com" -H "Access-Control-Request-Method: GET" -H "Access-Control-Request-Headers: Authorization" -X OPTIONS https://api.statuspage.com/users
```

## Conclusion

The microservice security system provides comprehensive security features including:

- **JWT Authentication** - Secure token-based authentication
- **OAuth2 Integration** - Support for multiple OAuth2 providers
- **mTLS** - Certificate-based authentication between services
- **Security Middleware** - Comprehensive security middleware stack
- **Rate Limiting** - Protection against abuse and DoS attacks
- **CORS** - Cross-origin resource sharing configuration
- **Security Headers** - HTTP security headers for protection

Key benefits:

- **Industry Standards** - Implements industry-standard security patterns
- **Comprehensive Protection** - Protects against common security vulnerabilities
- **Scalable** - Designed for microservices architectures
- **Configurable** - Highly configurable security policies
- **Monitoring** - Comprehensive security monitoring and alerting
- **Testing** - Built-in security testing capabilities

This security system provides a solid foundation for securing microservices in production environments, following security best practices and industry standards.
