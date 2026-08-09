BEGIN;

UPDATE sources
SET adapter_version = 'v2',
    policy_version = 'malukuprov-public-v2',
    policy = policy - 'search' || '{
      "keywords": ["sistem informasi", "aplikasi", "integrasi data", "jaringan", "software", "cloud", "keamanan siber"],
      "page_size": 20,
      "max_pages_per_keyword": 3,
      "min_interval_ms": 3000
    }'::jsonb,
    updated_at = now()
WHERE adapter_name = 'spse-public-list' AND base_url = 'https://spse.inaproc.id';

INSERT INTO schema_migrations(version) VALUES (8);
COMMIT;
