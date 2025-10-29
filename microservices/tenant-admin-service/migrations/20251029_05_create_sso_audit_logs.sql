-- Migration: Create sso_audit_logs table
-- Service: tenant-admin-service
-- Date: 2025-10-29
-- Part: 5/5 - SSO Migration

CREATE TABLE IF NOT EXISTS sso_audit_logs (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Timestamp
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Multi-tenant association
    tenant_id UUID NOT NULL,

    -- Associations (nullable - event may not have user/provider yet)
    sso_provider_id UUID,
    user_id UUID,

    -- Event details
    event_type VARCHAR(100) NOT NULL,
    event_description TEXT,

    -- Request context
    ip_address VARCHAR(45),
    user_agent TEXT,

    -- SAML-specific
    saml_request_id VARCHAR(500),
    saml_response_status VARCHAR(100),

    -- Additional metadata (JSONB)
    metadata JSONB DEFAULT '{}' NOT NULL,

    -- Foreign Keys (nullable)
    CONSTRAINT fk_sso_audit_tenant
        FOREIGN KEY (tenant_id)
        REFERENCES tenants(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_sso_audit_provider
        FOREIGN KEY (sso_provider_id)
        REFERENCES sso_providers(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_sso_audit_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE SET NULL
);

-- Indexes
CREATE INDEX idx_sso_audit_tenant ON sso_audit_logs(tenant_id);
CREATE INDEX idx_sso_audit_provider ON sso_audit_logs(sso_provider_id);
CREATE INDEX idx_sso_audit_user ON sso_audit_logs(user_id);
CREATE INDEX idx_sso_audit_created_at ON sso_audit_logs(created_at DESC);
CREATE INDEX idx_sso_audit_event_type ON sso_audit_logs(event_type);

-- Comments
COMMENT ON TABLE sso_audit_logs IS 'Complete audit trail for all SSO-related events';
COMMENT ON COLUMN sso_audit_logs.event_type IS 'Event types: login_success, login_failure, logout, config_change, jit_provision, metadata_refresh, certificate_expiring';
COMMENT ON COLUMN sso_audit_logs.metadata IS 'Additional event-specific data (JSON)';
COMMENT ON COLUMN sso_audit_logs.saml_response_status IS 'SAML StatusCode from assertion';
