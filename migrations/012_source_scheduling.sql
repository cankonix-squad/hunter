BEGIN;

UPDATE sources
SET policy = policy || '{"schedule_enabled":true,"schedule_interval_minutes":360}'::jsonb,
    updated_at = now()
WHERE adapter_name = 'spse-public-list'
  AND status = 'active';

INSERT INTO schema_migrations(version) VALUES (12);
COMMIT;
