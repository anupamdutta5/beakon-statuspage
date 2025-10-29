-- Migration: Add SSO/SAML fields to users table
-- Service: tenant-admin-service
-- Date: 2025-10-29
-- Part: 1/5 - SSO Migration

-- Add SSO authentication fields to existing users table
ALTER TABLE users
ADD COLUMN IF NOT EXISTS auth_method VARCHAR(50) DEFAULT 'password' NOT NULL,
ADD COLUMN IF NOT EXISTS sso_provider_id UUID,
ADD COLUMN IF NOT EXISTS is_sso_user BOOLEAN DEFAULT FALSE NOT NULL;

-- Create index for SSO provider lookups
CREATE INDEX IF NOT EXISTS idx_users_sso_provider_id ON users(sso_provider_id);

-- Create index for SSO user queries
CREATE INDEX IF NOT EXISTS idx_users_is_sso_user ON users(is_sso_user);

-- Create index for auth method queries
CREATE INDEX IF NOT EXISTS idx_users_auth_method ON users(auth_method);

-- Add comment
COMMENT ON COLUMN users.auth_method IS 'Authentication method: password, saml, oauth, oidc';
COMMENT ON COLUMN users.sso_provider_id IS 'Foreign key to sso_providers table (if SSO user)';
COMMENT ON COLUMN users.is_sso_user IS 'Flag indicating user authenticates via SSO';
