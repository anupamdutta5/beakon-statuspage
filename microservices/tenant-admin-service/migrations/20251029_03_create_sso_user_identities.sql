-- Migration: Create sso_user_identities table
-- Service: tenant-admin-service
-- Date: 2025-10-29
-- Part: 3/5 - SSO Migration

CREATE TABLE IF NOT EXISTS sso_user_identities (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Associations
    user_id UUID NOT NULL,
    sso_provider_id UUID NOT NULL,
    tenant_id UUID NOT NULL,

    -- IdP-specific identity
    idp_user_id VARCHAR(500) NOT NULL, -- NameID for SAML
    idp_username VARCHAR(255),
    idp_email VARCHAR(255),

    -- SAML session tracking
    session_index VARCHAR(500),
    name_id_format VARCHAR(500),

    -- Attributes from IdP (JSONB)
    attributes JSONB DEFAULT '{}' NOT NULL,

    -- Login tracking
    last_login_at TIMESTAMP,
    last_login_ip VARCHAR(45),
    login_count INTEGER DEFAULT 0 NOT NULL,

    -- Foreign Keys
    CONSTRAINT fk_sso_identities_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_sso_identities_provider
        FOREIGN KEY (sso_provider_id)
        REFERENCES sso_providers(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_sso_identities_tenant
        FOREIGN KEY (tenant_id)
        REFERENCES tenants(id)
        ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_sso_identities_user ON sso_user_identities(user_id);
CREATE INDEX idx_sso_identities_provider ON sso_user_identities(sso_provider_id);
CREATE INDEX idx_sso_identities_tenant ON sso_user_identities(tenant_id);
CREATE INDEX idx_sso_identities_idp_email ON sso_user_identities(idp_email);
CREATE UNIQUE INDEX idx_sso_identities_unique ON sso_user_identities(sso_provider_id, idp_user_id, tenant_id);

-- Comments
COMMENT ON TABLE sso_user_identities IS 'Links users to their SSO identities from IdPs';
COMMENT ON COLUMN sso_user_identities.idp_user_id IS 'Unique identifier from IdP (SAML NameID)';
COMMENT ON COLUMN sso_user_identities.session_index IS 'SAML SessionIndex for Single Logout';
COMMENT ON COLUMN sso_user_identities.attributes IS 'Raw attributes from IdP assertion (JSON)';
