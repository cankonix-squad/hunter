BEGIN;
INSERT INTO sources(name, base_url, timezone, adapter_name, adapter_version, status, policy_version, policy)
VALUES ('SPSE INAPROC Maluku', 'https://spse.inaproc.id', 'Asia/Jayapura',
'spse-public-list', 'v1', 'active', 'malukuprov-public-v1',
'{"public_collection_approved":true,"tenant":"malukuprov","year":2026,"search":"sistem informasi","page_size":20,"robots_reviewed_at":"2026-08-09","allowed_paths":["/malukuprov/lelang","/malukuprov/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1}'::jsonb);
INSERT INTO schema_migrations(version) VALUES (7);
COMMIT;
