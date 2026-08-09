BEGIN;

ALTER TABLE sources DROP CONSTRAINT IF EXISTS sources_base_url_key;
CREATE UNIQUE INDEX sources_adapter_tenant_unique
    ON sources(base_url, adapter_name, (policy->>'tenant'));

INSERT INTO sources(name, base_url, timezone, adapter_name, adapter_version, status, policy_version, policy)
VALUES
('SPSE INAPROC Komdigi', 'https://spse.inaproc.id', 'Asia/Jakarta',
 'spse-public-list', 'v2', 'active', 'komdigi-public-v1',
 '{"public_collection_approved":true,"tenant":"komdigi","year":2027,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-09","allowed_paths":["/komdigi/lelang","/komdigi/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1}'::jsonb),
('SPSE INAPROC Jakarta', 'https://spse.inaproc.id', 'Asia/Jakarta',
 'spse-public-list', 'v2', 'active', 'jakarta-public-v1',
 '{"public_collection_approved":true,"tenant":"jakarta","year":2027,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-09","allowed_paths":["/jakarta/lelang","/jakarta/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1}'::jsonb),
('SPSE INAPROC Jawa Barat', 'https://spse.inaproc.id', 'Asia/Jakarta',
 'spse-public-list', 'v2', 'active', 'jabarprov-public-v1',
 '{"public_collection_approved":true,"tenant":"jabarprov","year":2026,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-09","allowed_paths":["/jabarprov/lelang","/jabarprov/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1}'::jsonb),
('SPSE INAPROC Jawa Timur', 'https://spse.inaproc.id', 'Asia/Jakarta',
 'spse-public-list', 'v2', 'active', 'jatimprov-public-v1',
 '{"public_collection_approved":true,"tenant":"jatimprov","year":2026,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-09","allowed_paths":["/jatimprov/lelang","/jatimprov/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1}'::jsonb);

INSERT INTO schema_migrations(version) VALUES (10);
COMMIT;
