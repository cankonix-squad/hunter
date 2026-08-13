BEGIN;

UPDATE sources
SET policy = policy || '{"priority_source":true}'::jsonb
WHERE status = 'active'
  AND coalesce((policy->>'schedule_enabled')::boolean, false);

WITH ministry(name, tenant) AS (
  VALUES
    ('SPSE INAPROC Kementerian Keuangan', 'kemenkeu'),
    ('SPSE INAPROC Kementerian Pendidikan Dasar dan Menengah', 'kemendikdasmen'),
    ('SPSE INAPROC Kementerian Kesehatan', 'kemkes'),
    ('SPSE INAPROC Kementerian Koordinator Politik dan Keamanan', 'polkam'),
    ('SPSE INAPROC Nasional (Kementerian Bersama)', 'nasional'),
    ('SPSE INAPROC Kementerian Pekerjaan Umum', 'pu'),
    ('SPSE INAPROC Kementerian ATR/BPN', 'atrbpn'),
    ('SPSE INAPROC Kementerian UMKM', 'kemenkopukm'),
    ('SPSE INAPROC Kementerian Pelindungan Pekerja Migran', 'kp2mi'),
    ('SPSE INAPROC Kementerian Kehutanan', 'kehutanan'),
    ('SPSE INAPROC Kementerian Investasi dan Hilirisasi/BKPM', 'bkpm'),
    ('SPSE INAPROC Kementerian Dalam Negeri', 'kemendagri'),
    ('SPSE INAPROC Kementerian ESDM', 'esdm'),
    ('SPSE INAPROC Kementerian Perhubungan', 'dephub'),
    ('SPSE INAPROC LKPP (Kementerian Bersama)', 'lkpp'),
    ('SPSE INAPROC Kementerian Perindustrian', 'kemenperin'),
    ('SPSE INAPROC Kementerian Kependudukan dan Pembangunan Keluarga', 'kemendukbangga'),
    ('SPSE INAPROC Kementerian Luar Negeri', 'kemlu'),
    ('SPSE INAPROC Kementerian Agama', 'kemenag'),
    ('SPSE INAPROC Kementerian Ketenagakerjaan', 'kemnaker')
)
INSERT INTO sources(name, base_url, timezone, adapter_name, adapter_version, status, policy_version, policy)
SELECT name, 'https://spse.inaproc.id', 'Asia/Jakarta', 'spse-public-list', 'v2', 'active',
       tenant || '-public-v1',
       jsonb_build_object(
         'public_collection_approved', true,
         'tenant', tenant,
         'year', 2026,
         'keywords', jsonb_build_array('sistem informasi','pengembangan aplikasi','pemeliharaan aplikasi','integrasi data','jaringan','data center','keamanan siber','software'),
         'page_size', 20,
         'max_pages_per_keyword', 2,
         'min_interval_ms', 3000,
         'robots_reviewed_at', '2026-08-13',
         'allowed_paths', jsonb_build_array('/' || tenant || '/lelang','/' || tenant || '/dt/lelang'),
         'max_response_bytes', 1048576,
         'request_timeout_seconds', 20,
         'max_concurrency', 1,
         'schedule_enabled', true,
         'schedule_interval_minutes', 360,
         'priority_source', false
       )
FROM ministry;

INSERT INTO schema_migrations(version) VALUES (30);
COMMIT;
