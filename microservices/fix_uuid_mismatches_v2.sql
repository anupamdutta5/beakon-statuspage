-- Fix UUID Mismatches Between Databases (Version 2)
-- This script ensures UUID consistency across saas_admin_db and tenant_admin_db

-- Step 1: Clean up saas_admin_db and recreate with correct UUIDs from tenant_admin_db
\c saas_admin_db

-- First, soft-delete all existing tenants to avoid conflicts
UPDATE tenants SET deleted_at = NOW() WHERE deleted_at IS NULL;

-- Now get all active tenants from tenant_admin_db and insert them with correct UUIDs
\echo 'Getting tenant data from tenant_admin_db...'

-- Insert tenants with matching UUIDs from tenant_admin_db
INSERT INTO tenants (id, name, slug, domain, subdomain, contact_email, status, is_active, created_at, updated_at)
SELECT
    t.id,
    t.name,
    t.slug,
    COALESCE(t.domain, t.slug || '.localhost'),
    t.slug, -- Use slug as subdomain
    t.contact_email,
    t.status,
    t.is_active,
    t.created_at,
    t.updated_at
FROM dblink(
    'dbname=tenant_admin_db host=localhost user=postgres',
    'SELECT id, name, slug, domain, subdomain, contact_email, status, is_active, created_at, updated_at FROM tenants WHERE deleted_at IS NULL'
) AS t(
    id uuid,
    name text,
    slug text,
    domain text,
    subdomain text,
    contact_email text,
    status text,
    is_active boolean,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    slug = EXCLUDED.slug,
    domain = EXCLUDED.domain,
    subdomain = EXCLUDED.subdomain,
    contact_email = EXCLUDED.contact_email,
    deleted_at = NULL;

\echo 'SaaS Admin DB tenants after synchronization:'
SELECT id, name, slug, subdomain FROM tenants WHERE deleted_at IS NULL ORDER BY created_at DESC;

\echo 'UUID synchronization completed successfully!'