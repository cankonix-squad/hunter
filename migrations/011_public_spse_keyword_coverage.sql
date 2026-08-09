BEGIN;

UPDATE sources
SET policy = jsonb_set(
        jsonb_set(policy, '{year}', '2026'::jsonb),
        '{keywords}',
        '["sistem","aplikasi","digital","data","jaringan","teknologi informasi","software","keamanan siber"]'::jsonb
    ),
    policy_version = (policy->>'tenant') || '-public-v2',
    updated_at = now()
WHERE adapter_name = 'spse-public-list'
  AND (policy->>'tenant') IN ('komdigi','jakarta','jabarprov','jatimprov');

INSERT INTO schema_migrations(version) VALUES (11);
COMMIT;
