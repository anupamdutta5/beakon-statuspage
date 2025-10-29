-- Migration: Create user_sessions table for refresh token management
-- Date: 2025-01-29
-- Description: Implements refresh token storage for JWT authentication
--              Access Token: 15 minutes (short-lived)
--              Refresh Token: 7 days (long-lived, stored in this table)

-- Create user_sessions table
-- Note: Foreign key constraints are omitted as they will be added by Atlas migrations
--       when the full schema is generated from GORM models
CREATE TABLE IF NOT EXISTS user_sessions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    token VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_tenant_id ON user_sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(token);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_sessions_token_unique ON user_sessions(token) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_is_active ON user_sessions(is_active);

-- Updated_at trigger
CREATE OR REPLACE FUNCTION update_user_sessions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_user_sessions_updated_at
    BEFORE UPDATE ON user_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_user_sessions_updated_at();

-- Add comment
COMMENT ON TABLE user_sessions IS 'Stores refresh tokens for JWT authentication (7-day expiry)';
COMMENT ON COLUMN user_sessions.token IS 'Base64-encoded random token (32 bytes)';
COMMENT ON COLUMN user_sessions.expires_at IS 'Refresh token expiration (7 days from creation)';
COMMENT ON COLUMN user_sessions.is_active IS 'False when token is revoked (logout)';
