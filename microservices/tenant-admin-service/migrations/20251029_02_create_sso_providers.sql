-- Migration: Create sso_providers table
-- Service: tenant-admin-service
-- Date: 2025-10-29
-- Part: 2/5 - SSO Migration

CREATE TABLE IF NOT EXISTS sso_providers (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    -- Multi-tenant association
    tenant_id UUID NOT NULL,

    -- Organization identification
    organization_domain VARCHAR(255) NOT NULL,
    organization_name VARCHAR(255) NOT NULL,

    -- Provider configuration
    provider_type VARCHAR(50) NOT NULL, -- 'saml', 'oauth', 'oidc'
    provider_name VARCHAR(255) NOT NULL, -- 'Okta', 'Azure AD', 'Google Workspace', etc.

    -- SAML-specific fields
    entity_id TEXT,
    idp_entity_id TEXT,
    sso_url TEXT,
    slo_url TEXT,
    idp_metadata_url TEXT,
    idp_metadata_xml TEXT,
    idp_certificate TEXT,
    sp_certificate TEXT,
    sp_private_key TEXT,

    -- OAuth/OIDC fields
    client_id TEXT,
    client_secret TEXT,
    authorization_endpoint TEXT,
    token_endpoint TEXT,
    userinfo_endpoint TEXT,
    jwks_uri TEXT,

    -- Configuration flags
    is_enabled BOOLEAN DEFAULT TRUE NOT NULL,
    is_default BOOLEAN DEFAULT FALSE NOT NULL,
    enforce_sso BOOLEAN DEFAULT FALSE NOT NULL,
    allow_idp_initiated BOOLEAN DEFAULT TRUE NOT NULL,

    -- JIT Provisioning
    enable_jit_provisioning BOOLEAN DEFAULT TRUE NOT NULL,
    default_role VARCHAR(50) DEFAULT 'viewer' NOT NULL,

    -- Attribute mapping (JSONB)
    attribute_mapping JSONB DEFAULT '{}' NOT NULL,

    -- Metadata
    metadata JSONB DEFAULT '{}' NOT NULL,
    last_metadata_update TIMESTAMP,

    -- Foreign Key
    CONSTRAINT fk_sso_providers_tenant
        FOREIGN KEY (tenant_id)
        REFERENCES tenants(id)
        ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_sso_providers_tenant ON sso_providers(tenant_id);
CREATE INDEX idx_sso_providers_domain ON sso_providers(organization_domain);
CREATE INDEX idx_sso_providers_tenant_domain ON sso_providers(tenant_id, organization_domain);
CREATE INDEX idx_sso_providers_deleted_at ON sso_providers(deleted_at);

-- Comments
COMMENT ON TABLE sso_providers IS 'SSO/SAML identity provider configurations (multi-tenant)';
COMMENT ON COLUMN sso_providers.provider_type IS 'Type of SSO: saml, oauth, oidc';
COMMENT ON COLUMN sso_providers.organization_domain IS 'Email domain for this organization (e.g., @company.com)';
COMMENT ON COLUMN sso_providers.attribute_mapping IS 'Maps IdP attributes to user fields (JSON)';
COMMENT ON COLUMN sso_providers.enforce_sso IS 'If true, only SSO login allowed for this domain';
COMMENT ON COLUMN sso_providers.allow_idp_initiated IS 'Allow IdP-initiated SAML login';
