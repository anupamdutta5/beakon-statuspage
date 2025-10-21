--
-- Migration: Convert saas_subscribers.tenant_id from TEXT to UUID
-- Date: 2025-10-20
-- Rollback: Restore from backup
--

BEGIN;

-- Add new UUID column
ALTER TABLE saas_subscribers
  ADD COLUMN IF NOT EXISTS tenant_id_uuid UUID;

-- Populate from TEXT (which contains UUID strings)
UPDATE saas_subscribers
SET tenant_id_uuid = tenant_id::uuid
WHERE tenant_id_uuid IS NULL AND tenant_id IS NOT NULL;

-- Make NOT NULL
ALTER TABLE saas_subscribers
  ALTER COLUMN tenant_id_uuid SET NOT NULL;

-- Drop old column
ALTER TABLE saas_subscribers
  DROP COLUMN IF EXISTS tenant_id;

-- Rename UUID column
ALTER TABLE saas_subscribers
  RENAME COLUMN tenant_id_uuid TO tenant_id;

-- Create index
CREATE INDEX IF NOT EXISTS idx_saas_subscribers_tenant_id
  ON saas_subscribers(tenant_id);

COMMIT;

-- Verification query (run separately):
-- SELECT id, tenant_id, email, subscribed_at FROM saas_subscribers LIMIT 5;
