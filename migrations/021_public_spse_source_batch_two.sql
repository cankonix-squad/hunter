BEGIN;

INSERT INTO sources(name, base_url, timezone, adapter_name, adapter_version, status, policy_version, policy)
VALUES
('SPSE INAPROC Jawa Tengah', 'https://spse.inaproc.id', 'Asia/Jakarta',
 'spse-public-list', 'v2', 'active', 'jatengprov-public-v1',
 '{"public_collection_approved":true,"tenant":"jatengprov","year":2026,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-10","allowed_paths":["/jatengprov/lelang","/jatengprov/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1,"schedule_enabled":true,"schedule_interval_minutes":360}'::jsonb),
('SPSE INAPROC DI Yogyakarta', 'https://spse.inaproc.id', 'Asia/Jakarta',
 'spse-public-list', 'v2', 'active', 'jogjaprov-public-v1',
 '{"public_collection_approved":true,"tenant":"jogjaprov","year":2026,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-10","allowed_paths":["/jogjaprov/lelang","/jogjaprov/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1,"schedule_enabled":true,"schedule_interval_minutes":360}'::jsonb),
('SPSE INAPROC Banten', 'https://spse.inaproc.id', 'Asia/Jakarta',
 'spse-public-list', 'v2', 'active', 'bantenprov-public-v1',
 '{"public_collection_approved":true,"tenant":"bantenprov","year":2026,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-10","allowed_paths":["/bantenprov/lelang","/bantenprov/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1,"schedule_enabled":true,"schedule_interval_minutes":360}'::jsonb),
('SPSE INAPROC Sumatera Utara', 'https://spse.inaproc.id', 'Asia/Jakarta',
 'spse-public-list', 'v2', 'active', 'sumutprov-public-v1',
 '{"public_collection_approved":true,"tenant":"sumutprov","year":2026,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-10","allowed_paths":["/sumutprov/lelang","/sumutprov/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1,"schedule_enabled":true,"schedule_interval_minutes":360}'::jsonb),
('SPSE INAPROC Sulawesi Selatan', 'https://spse.inaproc.id', 'Asia/Makassar',
 'spse-public-list', 'v2', 'active', 'sulselprov-public-v1',
 '{"public_collection_approved":true,"tenant":"sulselprov","year":2026,"keywords":["sistem informasi","pengembangan aplikasi","pemeliharaan aplikasi","integrasi data","jaringan","data center","keamanan siber","software"],"page_size":20,"max_pages_per_keyword":2,"min_interval_ms":3000,"robots_reviewed_at":"2026-08-10","allowed_paths":["/sulselprov/lelang","/sulselprov/dt/lelang"],"max_response_bytes":1048576,"request_timeout_seconds":20,"max_concurrency":1,"schedule_enabled":true,"schedule_interval_minutes":360}'::jsonb);

INSERT INTO schema_migrations(version) VALUES (21);
COMMIT;
