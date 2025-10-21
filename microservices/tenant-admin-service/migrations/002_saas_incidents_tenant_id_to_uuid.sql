--
-- Migration: Convert saas_incidents.tenant_id from TEXT to UUID
-- Date: 2025-10-20
-- Rollback: Restore from backup
--

BEGIN;

-- Add new UUID column
ALTER TABLE saas_incidents
  ADD COLUMN IF NOT EXISTS tenant_id_uuid UUID;

-- Populate from TEXT (which contains UUID strings)
UPDATE saas_incidents
SET tenant_id_uuid = tenant_id::uuid
WHERE tenant_id_uuid IS NULL AND tenant_id IS NOT NULL;

-- Make NOT NULL
ALTER TABLE saas_incidents
  ALTER COLUMN tenant_id_uuid SET NOT NULL;

-- Drop old column
ALTER TABLE saas_incidents
  DROP COLUMN IF EXISTS tenant_id;

-- Rename UUID column
ALTER TABLE saas_incidents
  RENAME COLUMN tenant_id_uuid TO tenant_id;

-- Create index
CREATE INDEX IF NOT EXISTS idx_saas_incidents_tenant_id
  ON saas_incidents(tenant_id);

COMMIT;

-- Verification query (run separately):
-- SELECT id, tenant_id, title, status FROM saas_incidents LIMIT 5;
