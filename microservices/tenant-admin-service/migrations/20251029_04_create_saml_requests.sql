-- Migration: Create saml_requests table
-- Service: tenant-admin-service
-- Date: 2025-10-29
-- Part: 4/5 - SSO Migration

CREATE TABLE IF NOT EXISTS saml_requests (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Timestamp
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Request tracking
    request_id VARCHAR(500) UNIQUE NOT NULL,
    sso_provider_id UUID NOT NULL,
    tenant_id UUID NOT NULL,

    -- Request details
    relay_state TEXT,
    acs_url TEXT NOT NULL,

    -- Request metadata
    ip_address VARCHAR(45),
    user_agent TEXT,

    -- Expiration and completion
    expires_at TIMESTAMP NOT NULL,
    is_completed BOOLEAN DEFAULT FALSE NOT NULL,
    completed_at TIMESTAMP,

    -- Foreign Keys
    CONSTRAINT fk_saml_requests_provider
        FOREIGN KEY (sso_provider_id)
        REFERENCES sso_providers(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_saml_requests_tenant
        FOREIGN KEY (tenant_id)
        REFERENCES tenants(id)
        ON DELETE CASCADE
);

-- Indexes
CREATE UNIQUE INDEX idx_saml_requests_request_id ON saml_requests(request_id);
CREATE INDEX idx_saml_requests_tenant ON saml_requests(tenant_id);
CREATE INDEX idx_saml_requests_expires_at ON saml_requests(expires_at);
CREATE INDEX idx_saml_requests_is_completed ON saml_requests(is_completed);

-- Comments
COMMENT ON TABLE saml_requests IS 'Tracks pending SAML authentication requests (SP-initiated flow)';
COMMENT ON COLUMN saml_requests.request_id IS 'Unique SAML AuthnRequest ID';
COMMENT ON COLUMN saml_requests.relay_state IS 'Return URL after successful authentication';
COMMENT ON COLUMN saml_requests.acs_url IS 'Assertion Consumer Service URL';
COMMENT ON COLUMN saml_requests.expires_at IS 'Request expires after 5 minutes (security)';
