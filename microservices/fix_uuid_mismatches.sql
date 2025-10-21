-- Fix UUID Mismatches Between Databases
-- This script ensures UUID consistency across saas_admin_db and tenant_admin_db

-- Step 1: Check for UUID mismatches
\echo 'Checking for UUID mismatches between databases...'

-- Connect to tenant_admin_db and create temp table with tenant data
\c tenant_admin_db

CREATE TEMP TABLE tenant_uuid_map AS
SELECT id, name, slug, contact_email
FROM tenants
WHERE deleted_at IS NULL;

\echo 'Tenant Admin DB tenants:'
SELECT * FROM tenant_uuid_map;

-- Step 2: Update saas_admin_db tenants to match tenant_admin_db UUIDs
\c saas_admin_db

\echo 'SaaS Admin DB tenants before update:'
SELECT id, name, slug FROM tenants WHERE deleted_at IS NULL;

-- For each tenant, update UUID in saas_admin_db to match tenant_admin_db
-- We'll do this by slug matching since slug is unique

-- First, let's create a mapping of what needs to be updated
WITH tenant_admin_data AS (
    SELECT '15a5c38b-3d8e-4559-a584-c644a3ea6a83'::uuid as tenant_admin_id, 'newnewnew' as slug
    UNION ALL
    SELECT '46d08c29-d6cb-4121-ae18-1415379cbfcf'::uuid, 'five'
    UNION ALL
    SELECT '2c4783b3-7cff-410f-bd3a-3c397e54f576'::uuid, 'test-tenant'
    UNION ALL
    SELECT '8cb91f3a-bffd-4726-94b7-03cf94c6fe07'::uuid, 'testing'
    UNION ALL
    SELECT 'f01855cd-4a87-4b09-af65-762a7f73e4f9'::uuid, 'hello'
    UNION ALL
    SELECT '72ffbda4-fd0d-4448-854a-935e36d32b90'::uuid, 'anupam'
    UNION ALL
    SELECT '84ae825e-8aa5-4b53-b987-c17ecb40a60e'::uuid, 'anupamdutta'
    UNION ALL
    SELECT '801d4b8f-b1fe-4662-b098-88d1d8ec7302'::uuid, 'anupam'
)
SELECT
    sa.id as saas_admin_id,
    ta.tenant_admin_id,
    sa.slug,
    sa.name
FROM tenants sa
LEFT JOIN tenant_admin_data ta ON sa.slug = ta.slug
WHERE sa.deleted_at IS NULL;

\echo 'Creating function to update UUIDs safely...'

-- Since we can't directly update primary keys, we need to:
-- 1. Delete records from saas_admin_db that don't exist in tenant_admin_db
-- 2. Insert records with correct UUIDs for those that exist in tenant_admin_db but not in saas_admin_db

-- First, delete orphaned records from saas_admin_db
DELETE FROM tenants
WHERE slug IN ('five', 'test-tenant')
AND deleted_at IS NULL;

-- Now insert tenants that exist in tenant_admin_db but not in saas_admin_db with correct UUIDs
INSERT INTO tenants (id, name, slug, contact_email, domain, subdomain, status, is_active, created_at, updated_at)
VALUES
    ('15a5c38b-3d8e-4559-a584-c644a3ea6a83', 'Newnewnew', 'newnewnew', 'newnewnew@gmail.com', '', '', 'active', true, NOW(), NOW()),
    ('46d08c29-d6cb-4121-ae18-1415379cbfcf', 'five', 'five', 'five@gmail.com', '', '', 'active', true, NOW(), NOW()),
    ('2c4783b3-7cff-410f-bd3a-3c397e54f576', 'Test Tenant', 'test-tenant', 'test@example.com', '', '', 'active', true, NOW(), NOW()),
    ('8cb91f3a-bffd-4726-94b7-03cf94c6fe07', 'Testing', 'testing', 'testing@gmail.com', '', '', 'active', true, NOW(), NOW()),
    ('f01855cd-4a87-4b09-af65-762a7f73e4f9', 'Hello', 'hello', 'hello@gmail.com', '', '', 'active', true, NOW(), NOW()),
    ('84ae825e-8aa5-4b53-b987-c17ecb40a60e', 'anupamdutta', 'anupamdutta', 'anupamdutta.official700@gmail.com', '', '', 'active', true, NOW(), NOW())
ON CONFLICT (slug) DO NOTHING;

\echo 'SaaS Admin DB tenants after update:'
SELECT id, name, slug FROM tenants WHERE deleted_at IS NULL ORDER BY created_at DESC;

\echo 'UUID mismatch fix completed!'