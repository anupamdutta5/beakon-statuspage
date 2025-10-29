# SAML/SSO Configuration Guide

**Version**: 1.0
**Date**: October 29, 2025
**Service**: tenant-admin-service
**Supported Providers**: Okta, Azure AD, Google Workspace, OneLogin, Auth0, and any SAML 2.0 compliant IdP

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Quick Start](#quick-start)
4. [Configuration Steps](#configuration-steps)
   - [Okta Setup](#okta-setup)
   - [Azure AD Setup](#azure-ad-setup)
   - [Google Workspace Setup](#google-workspace-setup)
5. [Testing](#testing)
6. [Troubleshooting](#troubleshooting)
7. [Advanced Configuration](#advanced-configuration)
8. [Security Best Practices](#security-best-practices)

---

## Overview

The Beakon Status Page platform supports SAML 2.0 Single Sign-On (SSO) for enterprise authentication. This allows tenants to:

- **Centralized Authentication**: Users authenticate via their organization's Identity Provider (IdP)
- **Just-In-Time Provisioning**: Automatic user creation from SAML assertions
- **Attribute Mapping**: Map IdP attributes to user fields (email, name, role)
- **Multi-Provider Support**: Each tenant can configure multiple SSO providers
- **Secure Integration**: SAML 2.0 compliant with signature validation

### Key Concepts

**Service Provider (SP)**: Beakon Status Page platform (your application)
**Identity Provider (IdP)**: Your organization's SSO system (Okta, Azure AD, etc.)
**SP-Initiated Flow**: User starts at status page → redirected to IdP → logs in → redirected back
**IdP-Initiated Flow**: User starts at IdP portal → clicks status page app → logged in automatically
**JIT Provisioning**: Users created automatically on first SSO login

---

## Prerequisites

### 1. Database Migration

Apply SSO database migrations:

```bash
cd microservices/tenant-admin-service/migrations
./apply_sso_migrations.sh
```

This creates:
- `sso_providers` - SSO configurations
- `sso_user_identities` - User-IdP links
- `saml_requests` - Pending auth requests
- `sso_audit_logs` - Audit trail

### 2. Environment Configuration

Set the SAML base URL:

```bash
export SAML_BASE_URL=https://statuspage.yourdomain.com
```

Or for development:
```bash
export SAML_BASE_URL=https://acme.localhost:8099
```

### 3. HTTPS Requirement

**IMPORTANT**: SAML requires HTTPS in production. For development:

```bash
# Option 1: Use ngrok
ngrok http 8099

# Option 2: Use local SSL certificates
# Generate self-signed cert
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Run with HTTPS
./tenant-admin-service --tls --cert cert.pem --key key.pem
```

### 4. Admin Access

You need admin access to:
- Your IdP (Okta, Azure AD, etc.)
- Beakon tenant-admin-service API

---

## Quick Start

### 5-Minute Setup (Okta Example)

1. **Create SAML App in Okta**:
   - Go to Applications → Create App Integration
   - Choose SAML 2.0
   - App name: "Beakon Status Page"

2. **Configure SAML Settings**:
   ```
   Single sign on URL: https://acme.statuspage.com/api/v1/saml/acs
   Audience URI (SP Entity ID): https://acme.statuspage.com/saml
   ```

3. **Get IdP Metadata**:
   - Download metadata XML or copy certificate

4. **Create SSO Provider in Beakon**:
   ```bash
   curl -X POST https://acme.statuspage.com/api/v1/sso/providers \
     -H "Authorization: Bearer $JWT_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "organization_domain": "acme.com",
       "organization_name": "Acme Corp",
       "provider_type": "saml",
       "provider_name": "Okta",
       "idp_entity_id": "http://www.okta.com/xyz",
       "sso_url": "https://acme.okta.com/app/xyz/sso/saml",
       "idp_certificate": "-----BEGIN CERTIFICATE-----\n...",
       "is_enabled": true,
       "attribute_mapping": {
         "email": "email",
         "first_name": "firstName",
         "last_name": "lastName"
       }
     }'
   ```

5. **Test**:
   ```bash
   # Initiate login
   curl -X POST https://acme.statuspage.com/api/v1/saml/login \
     -H "Content-Type: application/json" \
     -d '{"organization_domain": "acme.com"}'
   ```

---

## Configuration Steps

### Okta Setup

#### Step 1: Create SAML Application

1. Log in to Okta Admin Console
2. Navigate to **Applications** → **Applications**
3. Click **Create App Integration**
4. Select **SAML 2.0**
5. Click **Next**

#### Step 2: Configure General Settings

- **App name**: Beakon Status Page
- **App logo**: (optional) Upload your logo
- **App visibility**: Check "Do not display application icon to users"

#### Step 3: Configure SAML Settings

**General**:
```
Single sign on URL: https://<tenant-subdomain>.statuspage.com/api/v1/saml/acs
  ☑ Use this for Recipient URL and Destination URL

Audience URI (SP Entity ID): https://<tenant-subdomain>.statuspage.com/saml

Default RelayState: (leave empty)

Name ID format: EmailAddress

Application username: Email
```

**Attribute Statements** (optional, for attribute mapping):
| Name | Name format | Value |
|------|-------------|-------|
| email | Unspecified | user.email |
| firstName | Unspecified | user.firstName |
| lastName | Unspecified | user.lastName |
| role | Unspecified | user.role |

**Group Attribute Statements** (optional):
| Name | Name format | Filter | Value |
|------|-------------|--------|-------|
| groups | Unspecified | Matches regex | .* |

#### Step 4: Get IdP Metadata

1. After creating the app, go to **Sign On** tab
2. Scroll to **SAML 2.0** section
3. Right-click **Identity Provider metadata** link → Copy Link Address
4. **OR** Click **View Setup Instructions** to see:
   - Identity Provider Single Sign-On URL
   - Identity Provider Issuer
   - X.509 Certificate

#### Step 5: Create SSO Provider in Beakon

```bash
curl -X POST https://acme.statuspage.com/api/v1/sso/providers \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "organization_domain": "acme.com",
    "organization_name": "Acme Corp",
    "provider_type": "saml",
    "provider_name": "Okta",
    "idp_entity_id": "http://www.okta.com/exk...",
    "sso_url": "https://acme.okta.com/app/acme_beakon_1/exk.../sso/saml",
    "idp_certificate": "-----BEGIN CERTIFICATE-----\nMIIDpDCCAoygAwIBAgIGAX...\n-----END CERTIFICATE-----",
    "is_enabled": true,
    "enable_jit_provisioning": true,
    "default_role": "viewer",
    "attribute_mapping": {
      "email": "email",
      "first_name": "firstName",
      "last_name": "lastName"
    }
  }'
```

#### Step 6: Assign Users in Okta

1. Go to **Assignments** tab
2. Click **Assign** → **Assign to People** or **Assign to Groups**
3. Select users/groups and click **Assign**

---

### Azure AD Setup

#### Step 1: Create Enterprise Application

1. Log in to Azure Portal
2. Navigate to **Azure Active Directory** → **Enterprise applications**
3. Click **+ New application**
4. Click **+ Create your own application**
5. Name: "Beakon Status Page"
6. Select **Integrate any other application you don't find in the gallery (Non-gallery)**
7. Click **Create**

#### Step 2: Configure Single Sign-On

1. Go to **Single sign-on** in left menu
2. Select **SAML**
3. Click **Edit** on **Basic SAML Configuration**

**Basic SAML Configuration**:
```
Identifier (Entity ID): https://<tenant-subdomain>.statuspage.com/saml

Reply URL (Assertion Consumer Service URL):
  https://<tenant-subdomain>.statuspage.com/api/v1/saml/acs

Sign on URL (optional): https://<tenant-subdomain>.statuspage.com

Relay State (optional): (leave empty)

Logout URL (optional): https://<tenant-subdomain>.statuspage.com/api/v1/saml/logout
```

#### Step 3: Configure User Attributes & Claims

Default claims:
- **Unique User Identifier**: user.userprincipalname → Change to user.mail
- **emailaddress**: user.mail
- **givenname**: user.givenname
- **surname**: user.surname

Add custom claims if needed:
- **Name**: role, **Value**: user.jobtitle

#### Step 4: Get SAML Metadata

1. Scroll to **SAML Signing Certificate** section
2. Download **Certificate (Base64)**
3. Copy **Login URL** from **Set up Beakon Status Page** section
4. Copy **Azure AD Identifier**

#### Step 5: Create SSO Provider in Beakon

```bash
curl -X POST https://acme.statuspage.com/api/v1/sso/providers \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "organization_domain": "acme.com",
    "organization_name": "Acme Corp",
    "provider_type": "saml",
    "provider_name": "Azure AD",
    "idp_entity_id": "https://sts.windows.net/...",
    "sso_url": "https://login.microsoftonline.com/.../saml2",
    "idp_certificate": "-----BEGIN CERTIFICATE-----\nMIIC8DCCAdigAwIBAgIQ...\n-----END CERTIFICATE-----",
    "is_enabled": true,
    "enable_jit_provisioning": true,
    "default_role": "viewer",
    "attribute_mapping": {
      "email": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
      "first_name": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
      "last_name": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"
    }
  }'
```

#### Step 6: Assign Users and Groups

1. Go to **Users and groups** in left menu
2. Click **+ Add user/group**
3. Select users or groups
4. Click **Assign**

---

### Google Workspace Setup

#### Step 1: Create SAML App

1. Log in to Google Admin Console
2. Go to **Apps** → **Web and mobile apps**
3. Click **Add App** → **Add custom SAML app**
4. Enter app name: "Beakon Status Page"
5. Click **Continue**

#### Step 2: Download IdP Metadata

1. On **Google Identity Provider details** page:
   - Download **Metadata** file (XML)
   - **OR** copy:
     - SSO URL
     - Entity ID
     - Certificate

2. Click **Continue**

#### Step 3: Configure Service Provider Details

```
ACS URL: https://<tenant-subdomain>.statuspage.com/api/v1/saml/acs

Entity ID: https://<tenant-subdomain>.statuspage.com/saml

Start URL (optional): https://<tenant-subdomain>.statuspage.com

Signed Response: ☑ (checked)

Name ID format: EMAIL

Name ID: Basic Information > Primary email
```

Click **Continue**

#### Step 4: Configure Attribute Mapping

| Google Directory attributes | App attributes |
|-----------------------------|----------------|
| Primary email | email |
| First name | firstName |
| Last name | lastName |

Click **Finish**

#### Step 5: Turn On the App

1. Click on the app you just created
2. Click **User access**
3. Select **ON for everyone** or specific organizational units
4. Click **Save**

#### Step 6: Create SSO Provider in Beakon

```bash
curl -X POST https://acme.statuspage.com/api/v1/sso/providers \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "organization_domain": "acme.com",
    "organization_name": "Acme Corp",
    "provider_type": "saml",
    "provider_name": "Google Workspace",
    "idp_entity_id": "https://accounts.google.com/o/saml2?idpid=...",
    "sso_url": "https://accounts.google.com/o/saml2/idp?idpid=...",
    "idp_certificate": "-----BEGIN CERTIFICATE-----\nMIIDdDCCAlygAwIBAgIGA...\n-----END CERTIFICATE-----",
    "is_enabled": true,
    "enable_jit_provisioning": true,
    "default_role": "viewer",
    "attribute_mapping": {
      "email": "email",
      "first_name": "firstName",
      "last_name": "lastName"
    }
  }'
```

---

## Testing

### 1. Test SP Metadata Endpoint

Verify your Service Provider metadata is accessible:

```bash
curl https://acme.statuspage.com/api/v1/saml/metadata?organization_domain=acme.com
```

Should return XML like:
```xml
<?xml version="1.0"?>
<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://acme.statuspage.com/saml">
  <SPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://acme.statuspage.com/api/v1/saml/acs" index="1"/>
  </SPSSODescriptor>
</EntityDescriptor>
```

### 2. Test SP-Initiated Login

```bash
# Step 1: Initiate login
curl -X POST https://acme.statuspage.com/api/v1/saml/login \
  -H "Content-Type: application/json" \
  -d '{"organization_domain": "acme.com", "relay_state": "/dashboard"}' | jq

# Response:
# {
#   "request_id": "id_abc123",
#   "redirect_url": "https://idp.example.com/saml?SAMLRequest=..."
# }

# Step 2: Open redirect_url in browser
# Step 3: Login at IdP
# Step 4: You'll be redirected back to /dashboard with JWT cookie set
```

### 3. Test IdP-Initiated Login

1. Log in to your IdP (Okta/Azure AD/Google)
2. Click on "Beakon Status Page" app icon
3. Should be logged in automatically

### 4. Verify User Creation (JIT Provisioning)

```bash
# Check if user was created
curl https://acme.statuspage.com/api/v1/users \
  -H "Authorization: Bearer $JWT_TOKEN" | jq '.[] | select(.email == "user@acme.com")'
```

### 5. Check SSO Audit Logs

```bash
# Get recent SSO events
psql -d tenant_admin_db -c "
  SELECT event_type, created_at, saml_response_status
  FROM sso_audit_logs
  WHERE tenant_id = '<tenant-uuid>'
  ORDER BY created_at DESC
  LIMIT 10;
"
```

---

## Troubleshooting

### Common Issues

#### 1. "Invalid SAML Response"

**Symptoms**: Error after IdP redirects back to ACS

**Possible Causes**:
- IdP certificate mismatch
- Clock skew between SP and IdP
- Audience/Entity ID mismatch

**Solutions**:
```bash
# Check certificate
openssl x509 -in idp_cert.pem -text -noout

# Verify entity ID matches
curl https://acme.statuspage.com/api/v1/saml/metadata?organization_domain=acme.com | grep entityID

# Check system time
timedatectl status
```

#### 2. "SSO Provider Not Found"

**Symptoms**: Error when initiating login

**Possible Causes**:
- organization_domain doesn't match
- SSO provider is disabled
- Tenant ID mismatch

**Solutions**:
```sql
-- Check provider configuration
SELECT id, organization_domain, is_enabled, tenant_id
FROM sso_providers
WHERE organization_domain = 'acme.com';

-- Enable provider
UPDATE sso_providers SET is_enabled = true WHERE id = '<provider-uuid>';
```

#### 3. "User Not Provisioned"

**Symptoms**: Authentication succeeds but user not created

**Possible Causes**:
- JIT provisioning disabled
- Email attribute not mapped correctly
- Max users limit reached

**Solutions**:
```sql
-- Enable JIT provisioning
UPDATE sso_providers
SET enable_jit_provisioning = true
WHERE organization_domain = 'acme.com';

-- Check attribute mapping
SELECT attribute_mapping FROM sso_providers WHERE organization_domain = 'acme.com';

-- Check tenant user limit
SELECT max_users, (SELECT COUNT(*) FROM users WHERE tenant_id = t.id) as current_users
FROM tenants t WHERE subdomain = 'acme';
```

#### 4. "SAML Request Expired"

**Symptoms**: "Request not found or expired" error

**Possible Causes**:
- More than 5 minutes elapsed since login initiation
- Request ID mismatch

**Solutions**:
```sql
-- Check recent requests
SELECT request_id, expires_at, is_completed, created_at
FROM saml_requests
WHERE tenant_id = '<tenant-uuid>'
ORDER BY created_at DESC
LIMIT 5;

-- Cleanup expired requests
DELETE FROM saml_requests WHERE expires_at < NOW();
```

### Debug Mode

Enable detailed SAML logging:

```bash
export LOG_LEVEL=debug
export SAML_DEBUG=true
./tenant-admin-service
```

Check logs for:
- SAML request generation
- SAML response validation
- Attribute extraction
- User provisioning

---

## Advanced Configuration

### Multiple SSO Providers per Tenant

Support different providers for different domains:

```bash
# Okta for @acme.com
curl -X POST https://acme.statuspage.com/api/v1/sso/providers \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "organization_domain": "acme.com",
    "provider_type": "saml",
    "provider_name": "Okta",
    ...
  }'

# Azure AD for @partner.acme.com
curl -X POST https://acme.statuspage.com/api/v1/sso/providers \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "organization_domain": "partner.acme.com",
    "provider_type": "saml",
    "provider_name": "Azure AD",
    ...
  }'
```

### Custom Attribute Mapping

Map custom IdP attributes to user fields:

```json
{
  "attribute_mapping": {
    "email": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
    "first_name": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
    "last_name": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
    "role": "customRole",
    "department": "department",
    "title": "jobTitle"
  }
}
```

### Dynamic Role Assignment

Assign roles based on IdP groups:

```json
{
  "attribute_mapping": {
    "email": "email",
    "first_name": "firstName",
    "last_name": "lastName",
    "role": "groups"
  },
  "role_mapping": {
    "Admins": "admin",
    "Managers": "manager",
    "Developers": "viewer"
  }
}
```

### Enforce SSO Only

Disable password login for SSO users:

```sql
UPDATE sso_providers
SET enforce_sso = true
WHERE organization_domain = 'acme.com';
```

---

## Security Best Practices

### 1. Certificate Management

- **Rotate certificates annually**
- **Use strong encryption** (RSA 2048-bit minimum)
- **Monitor expiration**: Set alerts 30 days before expiry

```bash
# Check certificate expiration
openssl x509 -in idp_cert.pem -noout -enddate
```

### 2. HTTPS Only

- **Never use HTTP** for SAML in production
- **Use TLS 1.2+**
- **Enable HSTS** headers

```nginx
# Nginx configuration
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
```

### 3. Audit Logging

Monitor SSO events:

```sql
-- Failed login attempts
SELECT COUNT(*), ip_address
FROM sso_audit_logs
WHERE event_type = 'login_failure'
  AND created_at > NOW() - INTERVAL '1 hour'
GROUP BY ip_address
HAVING COUNT(*) > 5;

-- Unusual login times
SELECT user_id, created_at
FROM sso_audit_logs
WHERE event_type = 'login_success'
  AND EXTRACT(HOUR FROM created_at) NOT BETWEEN 6 AND 22;
```

### 4. Session Management

- **Short token lifetime**: 15-minute access tokens
- **Refresh tokens**: 7-day refresh tokens
- **Idle timeout**: 30 minutes

### 5. Input Validation

All SAML assertions are validated:
- ✅ Signature verification
- ✅ Timestamp checks (5-minute tolerance)
- ✅ Audience validation
- ✅ Issuer validation
- ✅ Attribute sanitization

---

## API Reference

### List SSO Providers

```bash
GET /api/v1/sso/providers
Authorization: Bearer <jwt_token>

Response:
{
  "providers": [
    {
      "id": "uuid",
      "organization_domain": "acme.com",
      "organization_name": "Acme Corp",
      "provider_type": "saml",
      "provider_name": "Okta",
      "is_enabled": true,
      "is_default": false,
      "created_at": "2025-10-29T10:00:00Z"
    }
  ],
  "count": 1
}
```

### Create SSO Provider

```bash
POST /api/v1/sso/providers
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "organization_domain": "acme.com",
  "organization_name": "Acme Corp",
  "provider_type": "saml",
  "provider_name": "Okta",
  "idp_entity_id": "http://www.okta.com/xyz",
  "sso_url": "https://acme.okta.com/app/xyz/sso/saml",
  "idp_certificate": "-----BEGIN CERTIFICATE-----\n...",
  "is_enabled": true,
  "attribute_mapping": {
    "email": "email"
  }
}
```

### Update SSO Provider

```bash
PUT /api/v1/sso/providers/:id
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "is_enabled": false
}
```

### Delete SSO Provider

```bash
DELETE /api/v1/sso/providers/:id
Authorization: Bearer <jwt_token>
```

### Initiate SAML Login

```bash
POST /api/v1/saml/login
Content-Type: application/json

{
  "organization_domain": "acme.com",
  "relay_state": "/dashboard"
}

Response:
{
  "request_id": "id_xyz",
  "redirect_url": "https://idp.example.com/saml?SAMLRequest=..."
}
```

### Get SP Metadata

```bash
GET /api/v1/saml/metadata?organization_domain=acme.com

Response: XML (Service Provider metadata)
```

---

## Support

For issues or questions:

1. **Check logs**: `tail -f logs/tenant-admin-service.log`
2. **Check audit trail**: Query `sso_audit_logs` table
3. **Review documentation**: [README.md](microservices/tenant-admin-service/README.md)
4. **Test connectivity**: Verify IdP and SP can communicate

---

**Document Version**: 1.0
**Last Updated**: October 29, 2025
**Maintained By**: Beakon Engineering Team
