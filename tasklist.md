# CANKONIX Opportunity Intelligence — Task List

Dokumen ini adalah sumber jejak kerja utama proyek. Perbarui status, bukti
verifikasi, dan commit terkait setiap kali satu perubahan selesai.

## Konvensi status

- `[x]` selesai dan sudah diverifikasi
- `[~]` sudah tersedia sebagian, masih membutuhkan pekerjaan lanjutan
- `[ ]` belum dikerjakan
- `[!]` terblokir; alasan wajib ditulis di item tersebut

## Kondisi terakhir

- Terakhir diperbarui: 9 Agustus 2026
- Branch: `main`
- Stack lokal: React + TypeScript, Go, PostgreSQL, Docker Compose, Nginx
- Aplikasi lokal: `http://localhost:3000`
- Migrasi database aktif: versi 5
- Data acceptance: 2 opportunity canonical dan 2 version history
- Classification acceptance: 1 relevant dengan evidence dan 1 needs-review
- Worktree bersih setelah commit `44c8139`

## Selesai

### Foundation dan environment

- [x] Struktur repository dan dokumentasi produk/arsitektur
- [x] Toolchain lokal Git, Go, Node.js, dan Docker Desktop
- [x] Docker Compose untuk PostgreSQL, API, collector, web, dan edge
- [x] React + TypeScript production build
- [x] Go API dengan structured logging, request ID, recovery, dan graceful shutdown
- [x] Endpoint health dan readiness
- [x] PostgreSQL tidak dipublikasikan langsung ke host
- [x] Nginx reverse proxy, security headers, dan request-size limit
- [x] DNS upstream Nginx dinamis setelah API/web container direcreate
- [x] Migrasi forward-only dan tabel `schema_migrations`

Bukti: commit `f98c0b5`, `44c8139`; Docker smoke test web/health HTTP 200.

### Identity dan keamanan sesi

- [x] Bootstrap administrator lokal
- [x] JWT access token
- [x] Rotating refresh token
- [x] Deteksi reuse dan revocation satu token family
- [x] Logout dan revocation refresh session
- [x] Role-based access control untuk administrator/operator
- [x] Halaman login responsif dan safe error state
- [x] Protected dashboard
- [x] Persistensi sesi browser di `sessionStorage`
- [x] Refresh access token otomatis dan retry satu kali setelah HTTP 401
- [x] Password administrator development diperbarui sesuai permintaan
- [x] Sesi lama dicabut setelah perubahan password

Bukti: commit `50c73a2`, `1e152e4`; frontend auth tests dan acceptance login lulus.

### Source registry dan collector

- [x] Source registry dan administrator source creation/update
- [x] Source policy dan status management
- [x] Durable source-run queue dengan idempotency key
- [x] Worker lease dan expired-lease recovery
- [x] Bounded exponential retry dan dead-letter state
- [x] Circuit breaker dan source health score
- [x] Versioned adapter registry
- [x] Fixture-backed SPSE adapter
- [x] Stable metadata fingerprint
- [x] Idempotent source snapshots
- [x] Source-run metrics dan safe error persistence

Bukti: commit `f98c0b5`; collector acceptance memproses fixture tanpa duplikasi versi.

### Canonical opportunity feed

- [x] Migrasi v4 untuk `opportunities` dan `opportunity_versions`
- [x] Normalisasi title, agency, procurement type, status, currency, budget, dan deadline
- [x] Validasi lifecycle, currency, nilai integer, dan RFC3339 deadline
- [x] Canonical identity unik per source dan package identity
- [x] First seen, last seen, source timestamp, fingerprint, dan canonical URL
- [x] Version history yang merujuk source snapshot
- [x] Active feed hanya memuat `announced` dan `registration_open`
- [x] Search index dan active-feed index PostgreSQL
- [x] Authenticated list API dengan search, status, source, dan pagination
- [x] Authenticated opportunity detail API
- [x] UUID validation dan safe not-found response
- [x] OpenAPI contract untuk feed dan detail
- [x] Dashboard feed, metrics, search, lifecycle filter, dan pagination
- [x] Detail drawer dengan nilai, deadline, source, version count, dan official URL
- [x] Empty, loading, retry, dan degraded states
- [x] Responsive desktop/mobile layout

Bukti: commit `44c8139`; API acceptance menemukan feed dan detail fixture canonical.

### Verifikasi terakhir

- [x] `go test -race ./...`
- [x] `go vet ./...`
- [x] Frontend tests: 6/6 lulus
- [x] Frontend production build
- [x] NPM audit: 0 vulnerability
- [x] Nginx configuration test
- [x] Docker images API, collector, dan web berhasil dibangun
- [x] PostgreSQL migration v4 diterapkan
- [x] Collector end-to-end: snapshot → normalize → opportunity → version
- [x] Authenticated API search dan detail acceptance
- [x] API/web recreate tanpa edge restart tetap HTTP 200

## Task berikutnya — Rules, classification, dan scoring

Kerjakan bagian ini lebih dahulu. Satu increment dianggap selesai hanya jika
schema, domain logic, API/UI, auditability, dan test ikut selesai.

### 1. Deterministic classification foundation

- [x] Definisikan taxonomy layanan/produk IT CANKONIX yang versioned
- [x] Tambahkan migration taxonomy, opportunity classification, dan evidence
- [x] Buat rules engine murni yang berjalan sebelum AI
- [x] Buat include/exclude keyword rules dengan alasan yang dapat dibaca manusia
- [x] Pastikan rule tidak melakukan auto-reject kecuali deterministic rule disetujui
- [x] Simpan rule version, matched evidence, confidence, dan evaluated timestamp
- [x] Tambahkan fixture dan unit test untuk positive, negative, dan ambiguous case

Bukti: commit `1384f3b`; migration v5; `go test -race ./...` dan `go vet ./...` lulus; collector
acceptance mengklasifikasikan “Pengadaan Layanan Integrasi Data” sebagai
`relevant/data-ai` dengan satu evidence dan mempertahankan paket jaringan yang
ambigu sebagai `needs_review`.

### 2. Explainable scoring

- [ ] Definisikan score dimensions: capability fit, value, location, deadline, dan risk
- [ ] Gunakan versioned integer/decimal weights; jangan gunakan floating point untuk uang
- [ ] Simpan score total, dimension breakdown, evidence, dan scoring version
- [ ] Tambahkan reason codes dan explanation yang reviewable
- [ ] Tandai low-confidence opportunity sebagai `needs_review`
- [ ] Tambahkan regression tests agar perubahan weight terdeteksi

### 3. Classification/scoring API dan UI

- [ ] Tambahkan classification dan score ke feed/detail API
- [ ] Tambahkan filter relevance, minimum score, dan needs-review
- [ ] Tampilkan score breakdown dan evidence di detail drawer/page
- [ ] Tampilkan deterministic exclusion secara terpisah dari AI recommendation
- [ ] Tambahkan review queue untuk hasil ambiguous/low confidence
- [ ] Perbarui OpenAPI dan component tests

### 4. AI adapter setelah deterministic rules stabil

- [ ] Provider-agnostic AI interface
- [ ] Structured JSON schema validation
- [ ] Prompt registry dan prompt version
- [ ] Cache berdasarkan input fingerprint
- [ ] Per-run dan daily cost budget
- [ ] Provider/model/confidence/evidence persistence
- [ ] Golden dataset dan evaluation runner
- [ ] Safe fallback saat provider tidak tersedia atau output invalid

## Backlog MVP setelah classification/scoring

### Opportunity workspace dan decision workflow

- [ ] Saved views dan combined filters
- [ ] Full change timeline dari version history
- [ ] Ownership dan assignment
- [ ] Internal comments
- [ ] Shortlist
- [ ] Pursue/pass decision dengan reason wajib
- [ ] Next action dan deadline
- [ ] Readiness checklist
- [ ] Human override untuk recommendation/score
- [ ] Audit log seluruh decision dan override
- [ ] Initial pipeline analytics

### Public SPSE adapter

- [ ] Policy review untuk satu target SPSE publik
- [ ] Rate limiter dan robots/terms documentation
- [ ] HTTP client timeout, backoff, dan response-size limit
- [ ] Parser fixture dari response yang sudah direduksi/redacted
- [ ] Adapter regression tests
- [ ] Incremental discovery dan unchanged-detail avoidance
- [ ] Source-specific lag/error metrics
- [ ] Validasi tidak ada login, CAPTCHA bypass, atau supplier transaction
- [ ] Perluas bertahap menuju 20 priority sources setelah satu adapter stabil

### Notifications dan operations

- [ ] Alert rules untuk opportunity/material changes
- [ ] Quiet hours
- [ ] Notification deduplication
- [ ] Email adapter
- [ ] Webhook adapter
- [ ] Source health console
- [ ] Queue/lag/error dashboard
- [ ] Dead-letter replay workflow
- [ ] Failure recovery runbook

### Hardening dan release

- [~] Structured logs dan health/readiness sudah tersedia; metrics/traces belum lengkap
- [ ] CI pipeline untuk format, test, vet, build, migration, dan audit
- [ ] API authorization/contract integration suite
- [ ] Fresh-database migration test dari v1 sampai versi terbaru
- [ ] End-to-end discover → normalize → classify → score → review test
- [ ] Security review dan dependency/container scanning
- [ ] Performance dan pagination load test
- [ ] Backup dan restore drill
- [ ] Production secret management
- [ ] TLS/staging deployment configuration
- [ ] UAT dan deployment runbook
- [ ] Tujuh hari stability observation

## MVP completion gate

- [ ] Minimal 20 priority SPSE sources stabil selama tujuh hari
- [ ] Active feed tidak memuat paket completed/cancelled/evaluation/award/contract
- [ ] Classification precision minimal 80% pada golden dataset yang disetujui
- [ ] Recommendation dapat dijelaskan dan di-override manusia
- [ ] Alert terdeduplikasi
- [ ] Security checks lulus
- [ ] Backup/restore checks lulus

## Riwayat commit utama

| Commit | Hasil |
|---|---|
| `f98c0b5` | Bootstrap platform, API, database, collector, dan Docker stack |
| `50c73a2` | Authenticated React login dan session lifecycle |
| `1e152e4` | Perubahan password administrator development |
| `44c8139` | Canonical opportunity normalization, API, feed UI, dan history |
| `1384f3b` | Taxonomy dan deterministic classification yang auditable |

## Aturan pembaruan task list

1. Pindahkan status hanya setelah acceptance criteria dan test relevan lulus.
2. Tambahkan commit/bukti pada bagian yang selesai.
3. Catat blocker langsung pada item menggunakan `[!]`.
4. Jangan menyimpan password, token, atau secret di dokumen ini.
5. Jika scope berubah, perbarui task list dan roadmap pada commit yang sama.
