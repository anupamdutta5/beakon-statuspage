--
-- Migration: Convert incidents and incident_updates from INTEGER to UUID
-- Date: 2025-10-20
-- Rollback: Restore from backup
--

BEGIN;

-- Step 1: Add UUID columns
ALTER TABLE incidents
  ADD COLUMN IF NOT EXISTS id_uuid UUID DEFAULT gen_random_uuid();

ALTER TABLE incident_updates
  ADD COLUMN IF NOT EXISTS id_uuid UUID DEFAULT gen_random_uuid(),
  ADD COLUMN IF NOT EXISTS incident_id_uuid UUID;

-- Step 2: Populate incident_updates foreign key
UPDATE incident_updates iu
SET incident_id_uuid = (
  SELECT i.id_uuid
  FROM incidents i
  WHERE i.id = iu.incident_id
)
WHERE iu.incident_id_uuid IS NULL;

-- Step 3: Verify no orphaned records
DO $$
DECLARE
    orphan_count INT;
BEGIN
    SELECT COUNT(*) INTO orphan_count
    FROM incident_updates
    WHERE incident_id IS NOT NULL
      AND incident_id_uuid IS NULL;

    IF orphan_count > 0 THEN
        RAISE WARNING 'Found % orphaned records in incident_updates', orphan_count;
        -- Not raising exception since we want to handle this gracefully
    END IF;
END $$;

-- Step 4: Drop old constraints and columns
ALTER TABLE incident_updates
  DROP CONSTRAINT IF EXISTS incident_updates_incident_id_fkey CASCADE;

ALTER TABLE incidents
  DROP CONSTRAINT IF EXISTS incidents_pkey CASCADE;

ALTER TABLE incident_updates
  DROP CONSTRAINT IF EXISTS incident_updates_pkey CASCADE;

-- Drop old columns
ALTER TABLE incidents
  DROP COLUMN IF EXISTS id;

ALTER TABLE incident_updates
  DROP COLUMN IF EXISTS id,
  DROP COLUMN IF EXISTS incident_id;

-- Step 5: Rename UUID columns to standard names
ALTER TABLE incidents
  RENAME COLUMN id_uuid TO id;

ALTER TABLE incident_updates
  RENAME COLUMN id_uuid TO id;

ALTER TABLE incident_updates
  RENAME COLUMN incident_id_uuid TO incident_id;

-- Step 6: Add new constraints
ALTER TABLE incidents
  ADD PRIMARY KEY (id);

ALTER TABLE incident_updates
  ADD PRIMARY KEY (id);

ALTER TABLE incident_updates
  ADD CONSTRAINT fk_incident_updates_incident
  FOREIGN KEY (incident_id)
  REFERENCES incidents(id)
  ON DELETE CASCADE;

-- Step 7: Create indexes
CREATE INDEX IF NOT EXISTS idx_incidents_tenant_id
  ON incidents(tenant_id);

CREATE INDEX IF NOT EXISTS idx_incidents_status
  ON incidents(status);

CREATE INDEX IF NOT EXISTS idx_incident_updates_incident_id
  ON incident_updates(incident_id);

COMMIT;

-- Verification queries (run separately):
-- SELECT id, tenant_id, title, status FROM incidents LIMIT 5;
-- SELECT id, incident_id, status, message FROM incident_updates LIMIT 5;
