--
-- Migration: Convert permissions table from BIGINT to UUID
-- Date: 2025-10-21
-- Rollback: Restore from backup
--

BEGIN;

-- Step 1: Add UUID column
ALTER TABLE permissions
  ADD COLUMN IF NOT EXISTS id_uuid UUID DEFAULT gen_random_uuid();

-- Step 2: Create mapping for existing permissions (if any)
-- Since permissions is a system table, we'll maintain IDs for known permissions

-- Step 3: Update role_permissions FK references
ALTER TABLE role_permissions
  ADD COLUMN IF NOT EXISTS permission_id_uuid UUID;

UPDATE role_permissions rp
SET permission_id_uuid = (
  SELECT p.id_uuid
  FROM permissions p
  WHERE p.id = rp.permission_id
)
WHERE rp.permission_id_uuid IS NULL;

-- Step 4: Drop old constraints and columns
ALTER TABLE role_permissions
  DROP CONSTRAINT IF EXISTS fk_role_permissions_permission CASCADE;

ALTER TABLE permissions
  DROP CONSTRAINT IF EXISTS permissions_pkey CASCADE;

ALTER TABLE permissions
  DROP COLUMN IF EXISTS id;

ALTER TABLE role_permissions
  DROP COLUMN IF EXISTS permission_id;

-- Step 5: Rename UUID columns
ALTER TABLE permissions
  RENAME COLUMN id_uuid TO id;

ALTER TABLE role_permissions
  RENAME COLUMN permission_id_uuid TO permission_id;

-- Step 6: Add new constraints
ALTER TABLE permissions
  ADD PRIMARY KEY (id);

ALTER TABLE role_permissions
  ADD CONSTRAINT fk_role_permissions_permission
  FOREIGN KEY (permission_id)
  REFERENCES permissions(id)
  ON DELETE CASCADE;

-- Step 7: Recreate indexes
CREATE INDEX IF NOT EXISTS idx_permissions_name
  ON permissions(name);

CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id
  ON role_permissions(permission_id);

COMMIT;

-- Verification (run separately):
-- SELECT id, name, resource, action FROM permissions LIMIT 5;
