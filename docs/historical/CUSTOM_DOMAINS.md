# Custom Domain Management System

## Overview

The Custom Domain Management System allows tenants to use their own domains for their status pages and admin dashboards, with automatic SSL certificate management and DNS verification.

## Features

### 🌐 Domain Management
- **Custom Domain Setup**: Configure custom domains for status pages and admin dashboards
- **DNS Verification**: Automatic verification of domain ownership through DNS records
- **Domain Validation**: Comprehensive domain format validation
- **Status Monitoring**: Real-time domain and SSL status monitoring

### 🔒 SSL Certificate Management
- **Automatic SSL**: Let's Encrypt integration with automatic certificate provisioning
- **Custom Certificates**: Support for uploading custom SSL certificates
- **Certificate Monitoring**: Automatic monitoring of certificate expiration
- **Auto-Renewal**: Automatic renewal of Let's Encrypt certificates

### 📊 Status Monitoring
- **Real-time Status**: Live monitoring of domain and SSL status
- **Error Reporting**: Detailed error messages for troubleshooting
- **Health Checks**: Regular health checks for domain resolution and SSL validity

## API Endpoints

### Domain Management

#### Set Custom Domain
```http
POST /api/v1/tenant/domains
Content-Type: application/json

{
  "domain": "example.com",
  "verification_type": "dns",
  "admin_domain": "admin.example.com"
}
```

#### Get Domain Status
```http
GET /api/v1/tenant/domains/status
```

Response:
```json
{
  "status": {
    "domain": "example.com",
    "is_verified": true,
    "verification_type": "dns",
    "last_checked": "2024-01-15T10:30:00Z",
    "ssl_status": {
      "is_valid": true,
      "expires_at": "2024-04-15T10:30:00Z",
      "issuer": "Let's Encrypt",
      "common_name": "example.com",
      "last_checked": "2024-01-15T10:30:00Z"
    }
  },
  "instructions": {
    "dns_instructions": {
      "domain": "example.com",
      "instructions": [
        {
          "type": "A",
          "name": "@",
          "value": "YOUR_SERVER_IP",
          "ttl": "300"
        }
      ]
    }
  }
}
```

#### Verify Domain
```http
POST /api/v1/tenant/domains/verify
```

#### Remove Custom Domain
```http
DELETE /api/v1/tenant/domains
```

### DNS Instructions

#### Get DNS Setup Instructions
```http
GET /api/v1/tenant/domains/dns-instructions
```

Response:
```json
{
  "domain": "example.com",
  "instructions": [
    {
      "type": "A",
      "name": "@",
      "value": "YOUR_SERVER_IP",
      "ttl": "300"
    },
    {
      "type": "CNAME",
      "name": "www",
      "value": "example.com",
      "ttl": "300"
    }
  ],
  "verification": {
    "type": "TXT",
    "name": "_statuspage-verification",
    "value": "statuspage-verification=verify-1234567890"
  }
}
```

### SSL Management

#### Get SSL Setup Instructions
```http
GET /api/v1/tenant/domains/ssl-instructions
```

#### Enable Automatic SSL
```http
POST /api/v1/tenant/domains/ssl/auto
```

#### Upload Custom SSL Certificate
```http
POST /api/v1/tenant/domains/ssl/upload
Content-Type: multipart/form-data

certificate: [certificate file]
private_key: [private key file]
```

## Frontend Interface

### Domain Status Dashboard

The custom domain management interface provides:

1. **Current Domain Status**
   - Domain verification status
   - SSL certificate status
   - Last checked timestamp
   - Error messages (if any)

2. **Domain Setup Form**
   - Custom domain input
   - Verification method selection
   - Optional admin domain configuration

3. **Setup Instructions**
   - DNS record configuration
   - Verification record details
   - Step-by-step setup guide

4. **SSL Management**
   - Automatic SSL option (Let's Encrypt)
   - Custom certificate upload
   - SSL status monitoring

### Status Indicators

- ✅ **Verified**: Domain is properly configured and verified
- ⏳ **Pending**: Domain setup in progress
- ❌ **Error**: Configuration issue requiring attention
- 🔒 **SSL Active**: SSL certificate is valid and active
- ⚠️ **SSL Pending**: SSL certificate setup required

## DNS Configuration

### Required DNS Records

For a domain `example.com`, you need to configure:

1. **A Record** (Root Domain)
   ```
   Type: A
   Name: @
   Value: YOUR_SERVER_IP
   TTL: 300
   ```

2. **CNAME Record** (WWW Subdomain)
   ```
   Type: CNAME
   Name: www
   Value: example.com
   TTL: 300
   ```

3. **Verification Record** (TXT)
   ```
   Type: TXT
   Name: _statuspage-verification
   Value: statuspage-verification=verify-1234567890
   ```

### DNS Propagation

- DNS changes typically take 5-30 minutes to propagate
- Use the "Refresh" button to check status updates
- Some DNS providers may take up to 24 hours for full propagation

## SSL Certificate Management

### Automatic SSL (Let's Encrypt)

**Benefits:**
- Free SSL certificates
- Automatic renewal
- No manual configuration required
- Industry-standard security

**Requirements:**
- Domain must be verified
- Domain must point to our servers
- Port 80 and 443 must be accessible

### Custom SSL Certificates

**Supported Formats:**
- PEM format (.crt, .pem)
- Private key files (.key)

**Upload Process:**
1. Generate or obtain SSL certificate and private key
2. Use the upload form in the admin dashboard
3. System validates and configures the certificate
4. SSL status updates automatically

## Security Considerations

### Domain Verification
- DNS-based verification ensures domain ownership
- Prevents unauthorized domain usage
- Regular verification checks maintain security

### SSL Certificate Security
- Automatic certificate validation
- Expiration monitoring and alerts
- Secure certificate storage
- Regular security updates

### Access Control
- Tenant-specific domain management
- Authentication required for all operations
- Audit logging for all domain changes

## Troubleshooting

### Common Issues

#### Domain Not Resolving
- Check DNS records are correctly configured
- Verify DNS propagation (use `dig` or `nslookup`)
- Ensure domain points to correct server IP

#### SSL Certificate Issues
- Verify domain is properly verified
- Check certificate format and validity
- Ensure private key matches certificate

#### Verification Failures
- Confirm verification record is correctly added
- Check DNS propagation for verification record
- Verify record format and value

### Error Messages

- **"Invalid domain format"**: Check domain syntax and format
- **"Domain already in use"**: Domain is configured for another tenant
- **"DNS resolution failed"**: Domain doesn't resolve to our servers
- **"SSL certificate invalid"**: Certificate format or validity issue

## Best Practices

### Domain Setup
1. Use a dedicated subdomain for status pages (e.g., `status.example.com`)
2. Keep admin dashboard on separate subdomain (e.g., `admin.example.com`)
3. Use HTTPS for all custom domains
4. Set appropriate DNS TTL values (300-3600 seconds)

### SSL Management
1. Prefer automatic SSL for simplicity
2. Monitor certificate expiration dates
3. Keep backup certificates for critical domains
4. Use strong certificate authorities

### Monitoring
1. Regularly check domain status
2. Monitor SSL certificate expiration
3. Set up alerts for domain issues
4. Keep DNS records updated

## Integration Examples

### Nginx Configuration

For custom domains, you'll need to configure your web server:

```nginx
server {
    listen 80;
    server_name example.com www.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl;
    server_name example.com www.example.com;
    
    ssl_certificate /path/to/certificate.crt;
    ssl_certificate_key /path/to/private.key;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Docker Configuration

For containerized deployments:

```yaml
version: '3.8'
services:
  statuspage:
    image: statuspage:latest
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./ssl:/etc/ssl/certs
      - ./nginx.conf:/etc/nginx/nginx.conf
    environment:
      - DOMAIN=example.com
      - SSL_CERT_PATH=/etc/ssl/certs/certificate.crt
      - SSL_KEY_PATH=/etc/ssl/certs/private.key
```

## Support

For additional support with custom domain setup:

1. Check the troubleshooting section above
2. Verify DNS and SSL configuration
3. Contact support with domain and error details
4. Provide DNS and SSL status information

## Future Enhancements

Planned features for future releases:

- **Wildcard SSL Support**: Support for wildcard certificates
- **Multiple Domain Support**: Support for multiple domains per tenant
- **Advanced DNS Management**: Integration with DNS providers
- **Domain Transfer**: Easy domain transfer between tenants
- **Custom CDN Integration**: Integration with CDN providers
- **Advanced Monitoring**: Enhanced domain and SSL monitoring
