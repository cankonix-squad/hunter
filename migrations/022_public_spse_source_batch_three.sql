BEGIN;

INSERT INTO sources(name, base_url, timezone, adapter_name, adapter_version, status, policy_version, policy)
VALUES
('SPSE INAPROC Sumatera Barat', 'https://spse.inaproc.id', 'Asia/Jakarta',
 'spse-public-list', 'v2', 'active', 'sumbarprov-public-v1',
 '{"public_collection_approved":true,"tenant":"sumbarprov","year":2026,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-11","allowed_paths":["/sumbarprov/lelang","/sumbarprov/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1,"schedule_enabled":true,"schedule_interval_minutes":360}'::jsonb),
('SPSE INAPROC Sumatera Selatan', 'https://spse.inaproc.id', 'Asia/Jakarta',
 'spse-public-list', 'v2', 'active', 'sumselprov-public-v1',
 '{"public_collection_approved":true,"tenant":"sumselprov","year":2026,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-11","allowed_paths":["/sumselprov/lelang","/sumselprov/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1,"schedule_enabled":true,"schedule_interval_minutes":360}'::jsonb),
('SPSE INAPROC Kalimantan Barat', 'https://spse.inaproc.id', 'Asia/Jakarta',
 'spse-public-list', 'v2', 'active', 'kalbarprov-public-v1',
 '{"public_collection_approved":true,"tenant":"kalbarprov","year":2026,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-11","allowed_paths":["/kalbarprov/lelang","/kalbarprov/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1,"schedule_enabled":true,"schedule_interval_minutes":360}'::jsonb),
('SPSE INAPROC Kalimantan Timur', 'https://spse.inaproc.id', 'Asia/Makassar',
 'spse-public-list', 'v2', 'active', 'kaltimprov-public-v1',
 '{"public_collection_approved":true,"tenant":"kaltimprov","year":2026,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-11","allowed_paths":["/kaltimprov/lelang","/kaltimprov/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1,"schedule_enabled":true,"schedule_interval_minutes":360}'::jsonb),
('SPSE INAPROC Bali', 'https://spse.inaproc.id', 'Asia/Makassar',
 'spse-public-list', 'v2', 'active', 'baliprov-public-v1',
 '{"public_collection_approved":true,"tenant":"baliprov","year":2026,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-11","allowed_paths":["/baliprov/lelang","/baliprov/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1,"schedule_enabled":true,"schedule_interval_minutes":360}'::jsonb);

INSERT INTO schema_migrations(version) VALUES (22);
COMMIT;
