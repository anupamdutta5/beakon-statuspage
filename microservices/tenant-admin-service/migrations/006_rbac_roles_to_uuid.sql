--
-- Migration: Convert roles table from BIGINT to UUID (id and tenant_id)
-- Date: 2025-10-21
-- Rollback: Restore from backup
--

BEGIN;

-- Step 1: Add UUID columns
ALTER TABLE roles
  ADD COLUMN IF NOT EXISTS id_uuid UUID DEFAULT gen_random_uuid(),
  ADD COLUMN IF NOT EXISTS tenant_id_uuid UUID;

-- Step 2: For tenant_id, since roles table is empty, we can just set it to NULL initially
-- In production, you would populate this from a tenant mapping table

-- Step 3: Update foreign key references in dependent tables
ALTER TABLE role_permissions
  ADD COLUMN IF NOT EXISTS role_id_uuid UUID;

ALTER TABLE user_roles
  ADD COLUMN IF NOT EXISTS role_id_uuid UUID;

ALTER TABLE team_roles
  ADD COLUMN IF NOT EXISTS role_id_uuid UUID;

-- Populate foreign keys
UPDATE role_permissions rp
SET role_id_uuid = (
  SELECT r.id_uuid
  FROM roles r
  WHERE r.id = rp.role_id
)
WHERE rp.role_id_uuid IS NULL;

UPDATE user_roles ur
SET role_id_uuid = (
  SELECT r.id_uuid
  FROM roles r
  WHERE r.id = ur.role_id
)
WHERE ur.role_id_uuid IS NULL;

UPDATE team_roles tr
SET role_id_uuid = (
  SELECT r.id_uuid
  FROM roles r
  WHERE r.id = tr.role_id
)
WHERE tr.role_id_uuid IS NULL;

-- Step 4: Drop old constraints
ALTER TABLE role_permissions
  DROP CONSTRAINT IF EXISTS fk_role_permissions_role CASCADE;

ALTER TABLE user_roles
  DROP CONSTRAINT IF EXISTS fk_roles_user_roles CASCADE;

ALTER TABLE team_roles
  DROP CONSTRAINT IF EXISTS fk_team_roles_role CASCADE;

ALTER TABLE roles
  DROP CONSTRAINT IF EXISTS roles_pkey CASCADE;

-- Step 5: Drop old columns
ALTER TABLE roles
  DROP COLUMN IF EXISTS id,
  DROP COLUMN IF EXISTS tenant_id;

ALTER TABLE role_permissions
  DROP COLUMN IF EXISTS role_id;

ALTER TABLE user_roles
  DROP COLUMN IF EXISTS role_id;

ALTER TABLE team_roles
  DROP COLUMN IF EXISTS role_id;

-- Step 6: Rename UUID columns
ALTER TABLE roles
  RENAME COLUMN id_uuid TO id;

ALTER TABLE roles
  RENAME COLUMN tenant_id_uuid TO tenant_id;

ALTER TABLE role_permissions
  RENAME COLUMN role_id_uuid TO role_id;

ALTER TABLE user_roles
  RENAME COLUMN role_id_uuid TO role_id;

ALTER TABLE team_roles
  RENAME COLUMN role_id_uuid TO role_id;

-- Step 7: Add NOT NULL constraints (tenant_id can be NULL for now if empty)
ALTER TABLE roles
  ALTER COLUMN id SET NOT NULL;

-- Only set NOT NULL if we have a way to populate tenant_id
-- ALTER TABLE roles ALTER COLUMN tenant_id SET NOT NULL;

-- Step 8: Add new constraints
ALTER TABLE roles
  ADD PRIMARY KEY (id);

ALTER TABLE role_permissions
  ADD CONSTRAINT fk_role_permissions_role
  FOREIGN KEY (role_id)
  REFERENCES roles(id)
  ON DELETE CASCADE;

ALTER TABLE user_roles
  ADD CONSTRAINT fk_roles_user_roles
  FOREIGN KEY (role_id)
  REFERENCES roles(id)
  ON DELETE CASCADE;

ALTER TABLE team_roles
  ADD CONSTRAINT fk_team_roles_role
  FOREIGN KEY (role_id)
  REFERENCES roles(id)
  ON DELETE CASCADE;

-- Step 9: Recreate indexes
CREATE INDEX IF NOT EXISTS idx_roles_tenant_id
  ON roles(tenant_id);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id
  ON role_permissions(role_id);

CREATE INDEX IF NOT EXISTS idx_user_roles_role_id
  ON user_roles(role_id);

CREATE INDEX IF NOT EXISTS idx_team_roles_role_id
  ON team_roles(role_id);

COMMIT;

-- Verification (run separately):
-- SELECT id, tenant_id, name, display_name FROM roles LIMIT 5;
