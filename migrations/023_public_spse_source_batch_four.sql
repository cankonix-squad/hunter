BEGIN;

INSERT INTO sources(name, base_url, timezone, adapter_name, adapter_version, status, policy_version, policy)
VALUES
('SPSE INAPROC Nusa Tenggara Barat', 'https://spse.inaproc.id', 'Asia/Makassar',
 'spse-public-list', 'v2', 'active', 'ntbprov-public-v1',
 '{"public_collection_approved":true,"tenant":"ntbprov","year":2026,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-11","allowed_paths":["/ntbprov/lelang","/ntbprov/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1,"schedule_enabled":true,"schedule_interval_minutes":360}'::jsonb),
('SPSE INAPROC Nusa Tenggara Timur', 'https://spse.inaproc.id', 'Asia/Makassar',
 'spse-public-list', 'v2', 'active', 'nttprov-public-v1',
 '{"public_collection_approved":true,"tenant":"nttprov","year":2026,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-11","allowed_paths":["/nttprov/lelang","/nttprov/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1,"schedule_enabled":true,"schedule_interval_minutes":360}'::jsonb),
('SPSE INAPROC Sulawesi Utara', 'https://spse.inaproc.id', 'Asia/Makassar',
 'spse-public-list', 'v2', 'active', 'sulutprov-public-v1',
 '{"public_collection_approved":true,"tenant":"sulutprov","year":2026,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-11","allowed_paths":["/sulutprov/lelang","/sulutprov/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1,"schedule_enabled":true,"schedule_interval_minutes":360}'::jsonb),
('SPSE INAPROC Sulawesi Tengah', 'https://spse.inaproc.id', 'Asia/Makassar',
 'spse-public-list', 'v2', 'active', 'sultengprov-public-v1',
 '{"public_collection_approved":true,"tenant":"sultengprov","year":2026,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-11","allowed_paths":["/sultengprov/lelang","/sultengprov/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1,"schedule_enabled":true,"schedule_interval_minutes":360}'::jsonb),
('SPSE INAPROC Sulawesi Tenggara', 'https://spse.inaproc.id', 'Asia/Makassar',
 'spse-public-list', 'v2', 'active', 'sultraprov-public-v1',
 '{"public_collection_approved":true,"tenant":"sultraprov","year":2026,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-11","allowed_paths":["/sultraprov/lelang","/sultraprov/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1,"schedule_enabled":true,"schedule_interval_minutes":360}'::jsonb);

INSERT INTO schema_migrations(version) VALUES (23);
COMMIT;
