--
-- Migration: Convert users table from BIGINT to UUID (id only, tenant_id already UUID)
-- Date: 2025-10-21
-- Rollback: Restore from backup
--

BEGIN;

-- Step 1: Add UUID column for id
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS id_uuid UUID DEFAULT gen_random_uuid();

-- Step 2: Update foreign key references in dependent tables
ALTER TABLE user_roles
  ADD COLUMN IF NOT EXISTS user_id_uuid UUID,
  ADD COLUMN IF NOT EXISTS assigned_by_uuid UUID,
  ADD COLUMN IF NOT EXISTS tenant_id_uuid UUID;

-- Populate user_id
UPDATE user_roles ur
SET user_id_uuid = (
  SELECT u.id_uuid
  FROM users u
  WHERE u.id = ur.user_id
)
WHERE ur.user_id_uuid IS NULL;

-- Populate assigned_by
UPDATE user_roles ur
SET assigned_by_uuid = (
  SELECT u.id_uuid
  FROM users u
  WHERE u.id = ur.assigned_by
)
WHERE ur.assigned_by_uuid IS NULL AND ur.assigned_by IS NOT NULL;

-- For tenant_id in user_roles, convert BIGINT to UUID
-- Since tenant_id is BIGINT and we need UUID, and the table is empty, we can just prepare the column
-- In production with data, you'd need a mapping table

-- Step 3: Update references in other tables (incidents, components, etc.)
-- These will be handled when we migrate those tables

-- Step 4: Drop old constraints
ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_pkey CASCADE;

-- Step 5: Drop old columns
ALTER TABLE users
  DROP COLUMN IF EXISTS id;

ALTER TABLE user_roles
  DROP COLUMN IF EXISTS user_id,
  DROP COLUMN IF EXISTS assigned_by,
  DROP COLUMN IF EXISTS tenant_id;

-- Step 6: Rename UUID columns
ALTER TABLE users
  RENAME COLUMN id_uuid TO id;

ALTER TABLE user_roles
  RENAME COLUMN user_id_uuid TO user_id;

ALTER TABLE user_roles
  RENAME COLUMN assigned_by_uuid TO assigned_by;

ALTER TABLE user_roles
  RENAME COLUMN tenant_id_uuid TO tenant_id;

-- Step 7: Add NOT NULL constraints
ALTER TABLE users
  ALTER COLUMN id SET NOT NULL;

ALTER TABLE user_roles
  ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE user_roles
  ALTER COLUMN assigned_by SET NOT NULL;

-- Step 8: Add new constraints
ALTER TABLE users
  ADD PRIMARY KEY (id);

-- Step 9: Recreate indexes
CREATE INDEX IF NOT EXISTS idx_users_tenant_id
  ON users(tenant_id);

CREATE INDEX IF NOT EXISTS idx_users_email
  ON users(email);

CREATE INDEX IF NOT EXISTS idx_user_roles_user_id
  ON user_roles(user_id);

CREATE INDEX IF NOT EXISTS idx_user_roles_tenant_id
  ON user_roles(tenant_id);

-- Recreate unique constraint for tenant+email
DROP INDEX IF EXISTS idx_tenant_email;
CREATE UNIQUE INDEX idx_tenant_email
  ON users(tenant_id, email)
  WHERE deleted_at IS NULL;

COMMIT;

-- Verification (run separately):
-- SELECT id, tenant_id, email, first_name, last_name FROM users LIMIT 5;
-- SELECT id, user_id, role_id, tenant_id FROM user_roles LIMIT 5;
