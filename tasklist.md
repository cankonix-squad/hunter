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
- Migrasi database aktif: versi 8
- Data acceptance: 2 opportunity canonical dan 2 version history
- Classification acceptance: 1 relevant dengan evidence dan 1 needs-review
- Scoring acceptance: 2 score dengan masing-masing 5 dimension breakdown
- Public crawl acceptance: SPSE Maluku v2, 7 keyword/7 halaman diproses tanpa error;
  tidak ada paket IT yang cocok pada run terbaru
- Runtime lokal aktif: PostgreSQL, API, collector, web, dan edge
- Smoke test lokal: web, health, dan readiness HTTP 200
- Worktree bersih setelah commit `a7f5b24`

## Selesai

### Foundation dan environment

- [x] Struktur repository dan dokumentasi produk/arsitektur
- [x] Toolchain lokal Git, Go, Node.js, dan Docker Desktop
- [x] Docker Compose untuk PostgreSQL, API, collector, web, dan edge
- [x] React + TypeScript production build
- [x] Logo CANKONIX resmi pada halaman login dan sidebar aplikasi
- [x] Go API dengan structured logging, request ID, recovery, dan graceful shutdown
- [x] Endpoint health dan readiness
- [x] PostgreSQL tidak dipublikasikan langsung ke host
- [x] Nginx reverse proxy, security headers, dan request-size limit
- [x] DNS upstream Nginx dinamis setelah API/web container direcreate
- [x] Migrasi forward-only dan tabel `schema_migrations`

Bukti: commit `f98c0b5`, `44c8139`; Docker smoke test web, health, dan readiness HTTP 200 pada 9 Agustus 2026.

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

- [x] Definisikan score dimensions: capability fit, value, location, deadline, dan risk
- [x] Gunakan versioned integer/decimal weights; jangan gunakan floating point untuk uang
- [x] Simpan score total, dimension breakdown, evidence, dan scoring version
- [x] Tambahkan reason codes dan explanation yang reviewable
- [x] Tandai low-confidence opportunity sebagai `needs_review`
- [x] Tambahkan regression tests agar perubahan weight terdeteksi

Bukti: commit `9c3fce7`; migration v6; `go test -race ./...` dan `go vet ./...` lulus. Collector
acceptance menyimpan lima dimensi per opportunity: Integrasi Data mendapat score
8000 dan confidence 9000 (`score_ready`), sedangkan paket jaringan mendapat score
5625 dan confidence 6600 (`needs_review`).

### 3. Classification/scoring API dan UI

- [x] Tambahkan classification dan score ke feed/detail API
- [x] Tambahkan filter relevance, minimum score, dan needs-review
- [x] Tampilkan score breakdown dan evidence di detail drawer/page
- [x] Tampilkan deterministic exclusion secara terpisah dari AI recommendation
- [x] Tambahkan review queue untuk hasil ambiguous/low confidence
- [x] Perbarui OpenAPI dan component tests

Bukti: commit `343ce4d`; API acceptance memfilter paket relevant dengan minimum score 7000,
review queue mengembalikan paket low-confidence, dan detail Integrasi Data memuat
1 classification evidence serta 5 score dimensions. Frontend tests 7/7,
production build, `go test -race ./...`, dan `go vet ./...` lulus.

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

- [x] Policy review public listing SPSE INAPROC Maluku; robots.txt mengizinkan `/malukuprov`
- [x] Rate limiter dan robots/terms documentation
- [x] HTTP timeout, response-size limit, dan bounded retry/backoff untuk kegagalan transien
- [x] Parser fixture dari response yang sudah direduksi/redacted
- [x] Adapter regression tests
- [x] Incremental discovery via fingerprint; detail HTTP 403 tidak diambil
- [x] Server-side keyword search, pagination terbatas, dan deduplikasi lintas keyword
- [ ] Source-specific lag/error metrics
- [x] Validasi tidak ada login, CAPTCHA bypass, atau supplier transaction
- [ ] Perluas bertahap menuju 20 priority sources setelah satu adapter stabil

Target pertama dibatasi ke listing dan DataTables route yang dideklarasikan halaman
publik Maluku. Detail HTTP 403 tidak diakses ulang atau diakali; canonical URL tetap
disimpan untuk verifikasi manual.

Bukti: commit `0cb78f3`, `a7f5b24`; migration v7; live run `malukuprov-public-v1-acceptance` sukses
pada satu attempt dan menyimpan satu paket berstatus `completed`, sehingga tidak muncul
pada active-registration feed. Increment v2 memakai migration v8; live run
`malukuprov-public-v2-acceptance` sukses pada satu attempt dengan metrics 7 keyword,
7 halaman, 0 raw row, dan 0 kandidat. Hasil nol mencerminkan tidak adanya kecocokan
keyword IT saat run, bukan kegagalan transport atau parser.

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
| `9c3fce7` | Explainable scoring lima dimensi dan review confidence |
| `343ce4d` | Classification/scoring API, filter, review queue, dan UI evidence |
| `0cb78f3` | Public SPSE Maluku adapter dan acceptance crawl tanpa login |
| `57f2720` | Logo resmi CANKONIX pada login dan sidebar web |
| `a7f5b24` | Incremental keyword discovery, retry, pagination, dan crawl metrics SPSE v2 |

## Aturan pembaruan task list

1. Pindahkan status hanya setelah acceptance criteria dan test relevan lulus.
2. Tambahkan commit/bukti pada bagian yang selesai.
3. Catat blocker langsung pada item menggunakan `[!]`.
4. Jangan menyimpan password, token, atau secret di dokumen ini.
5. Jika scope berubah, perbarui task list dan roadmap pada commit yang sama.
