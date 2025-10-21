-- Create the status_page_domains table
CREATE TABLE IF NOT EXISTS status_page_domains (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    tenant_id BIGINT NOT NULL,
    status_page_id BIGINT NOT NULL,
    domain TEXT NOT NULL,
    status INTEGER NOT NULL DEFAULT 1, -- 1 = pending, 2 = verified, 3 = failed, 4 = disabled
    verification_method TEXT NOT NULL,
    verification_token TEXT NOT NULL,
    verified_at TIMESTAMP WITH TIME ZONE,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    ssl_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ssl_cert_issued_at TIMESTAMP WITH TIME ZONE,
    ssl_cert_expires_at TIMESTAMP WITH TIME ZONE,
    ssl_cert_issuer TEXT,
    ssl_cert_common_name TEXT,
    ssl_cert_raw TEXT,
    metadata TEXT,
    
    -- Add indexes
    CONSTRAINT fk_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_status_page FOREIGN KEY (status_page_id) REFERENCES status_pages(id) ON DELETE CASCADE,
    CONSTRAINT unique_domain UNIQUE (domain) DEFERRABLE INITIALLY DEFERRED
);

-- Add indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_status_page_domains_tenant_id ON status_page_domains(tenant_id);
CREATE INDEX IF NOT EXISTS idx_status_page_domains_status_page_id ON status_page_domains(status_page_id);
CREATE INDEX IF NOT EXISTS idx_status_page_domains_status ON status_page_domains(status);
CREATE INDEX IF NOT EXISTS idx_status_page_domains_is_primary ON status_page_domains(is_primary);
CREATE INDEX IF NOT EXISTS idx_status_page_domains_verification_token ON status_page_domains(verification_token);

-- Add a trigger to update the updated_at column
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_status_page_domains_modtime
    BEFORE UPDATE ON status_page_domains
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_column();

-- Create a function to ensure only one primary domain per status page
CREATE OR REPLACE FUNCTION ensure_single_primary_domain()
RETURNS TRIGGER AS $$
BEGIN
    -- If this is an update and is_primary is being set to false, we don't need to do anything
    IF TG_OP = 'UPDATE' AND NEW.is_primary = FALSE THEN
        RETURN NEW;
    END IF;
    
    -- If this is an update and is_primary is being set to true, unset any existing primary domain
    IF TG_OP = 'UPDATE' AND NEW.is_primary = TRUE AND OLD.is_primary = FALSE THEN
        UPDATE status_page_domains
        SET is_primary = FALSE, updated_at = NOW()
        WHERE status_page_id = NEW.status_page_id
        AND id != NEW.id
        AND is_primary = TRUE;
        RETURN NEW;
    END IF;
    
    -- For inserts, unset any existing primary domain
    IF TG_OP = 'INSERT' AND NEW.is_primary = TRUE THEN
        UPDATE status_page_domains
        SET is_primary = FALSE, updated_at = NOW()
        WHERE status_page_id = NEW.status_page_id
        AND is_primary = TRUE;
        RETURN NEW;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply the trigger
CREATE TRIGGER ensure_single_primary_domain_trigger
    BEFORE INSERT OR UPDATE ON status_page_domains
    FOR EACH ROW
    EXECUTE FUNCTION ensure_single_primary_domain();

-- Add a function to get the primary domain for a status page
CREATE OR REPLACE FUNCTION get_primary_domain(status_page_id_param BIGINT)
RETURNS TEXT AS $$
DECLARE
    primary_domain TEXT;
BEGIN
    SELECT domain INTO primary_domain
    FROM status_page_domains
    WHERE status_page_id = status_page_id_param
    AND is_primary = TRUE
    AND deleted_at IS NULL
    LIMIT 1;
    
    RETURN primary_domain;
END;
$$ LANGUAGE plpgsql STABLE;
