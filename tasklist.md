# CANKONIX Opportunity Intelligence — Task List

Dokumen ini adalah sumber jejak kerja utama proyek. Perbarui status, bukti
verifikasi, dan commit terkait setiap kali satu perubahan selesai.

## Konvensi status

- `[x]` selesai dan sudah diverifikasi
- `[~]` sudah tersedia sebagian, masih membutuhkan pekerjaan lanjutan
- `[ ]` belum dikerjakan
- `[!]` terblokir; alasan wajib ditulis di item tersebut

## Kondisi terakhir

- Terakhir diperbarui: 10 Agustus 2026
- Branch: `main`
- Stack lokal: React + TypeScript, Go, PostgreSQL, Docker Compose, Nginx
- Aplikasi lokal: `http://localhost:3000`
- Migrasi database aktif: versi 17
- Data acceptance: 173 opportunity canonical dan 173 version history
- Classification acceptance: 1 relevant dengan evidence dan 1 needs-review
- Scoring acceptance: 2 score dengan masing-masing 5 dimension breakdown
- Public crawl acceptance: 5 source aktif; batch Komdigi, Jakarta, Jawa Barat, dan
  Jawa Timur menghasilkan 171 paket publik nyata, 70 terklasifikasi relevant
- Runtime lokal aktif: PostgreSQL, API, collector, web, dan edge
- Smoke test lokal: web, health, dan readiness HTTP 200
- Worktree bersih setelah commit `45a8568`

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

- [x] Provider-agnostic AI interface
- [x] Structured JSON schema validation
- [x] Prompt registry dan prompt version
- [x] Cache berdasarkan input fingerprint
- [x] Per-run dan daily cost budget
- [x] Provider/model/confidence/evidence persistence
- [~] Evaluation runner dan candidate golden dataset tersedia; menunggu domain-owner approval
- [x] Safe fallback saat provider tidak tersedia atau output invalid

Bukti: commit `28f2972`; migration v9; package `internal/ai` menyediakan provider, prompt, store, dan budget
contracts; output invalid diperbaiki paling banyak satu kali; evidence wajib berasal dari
field input; provider/budget failure menghasilkan fallback `needs_review`. `go test -race
./...` dan `go vet ./...` lulus. Provider eksternal dan AI worker belum diaktifkan, sehingga
belum ada data atau biaya yang dikirim ke pihak ketiga.

Candidate golden dataset v1 (commit `bca56f7`) berisi 15 kasus versioned untuk tiga disposition dan empat
kategori relevan. Runner memvalidasi schema/taxonomy, menghasilkan confusion matrix,
precision/recall per label, accuracy, macro precision, serta exit gate untuk CI. Baseline
saat ini 15/15 dan macro precision 10000 bp; status tetap `candidate`, sehingga MVP gate
belum dianggap selesai sebelum review domain owner.

## Backlog MVP setelah classification/scoring

### Opportunity workspace dan decision workflow

- [x] Saved views dan combined filters
- [x] Full change timeline dari version history
- [x] Ownership dan assignment
- [x] Internal comments
- [x] Shortlist
- [x] Pursue/pass decision dengan reason wajib
- [ ] Next action dan deadline
- [ ] Readiness checklist
- [ ] Human override untuk recommendation/score
- [ ] Audit log seluruh decision dan override
- [ ] Initial pipeline analytics

Saved views (commit `1d9d929`) memakai migration v13 dan menyimpan kombinasi search, lifecycle, relevance,
minimum score, source, dan review-only per user. API create/list/delete membatasi ownership
di query database dan mengaudit create/delete; dashboard dapat menyimpan, menerapkan, dan
menghapus view. Acceptance API create/list/delete lulus dengan HTTP 204 saat delete dan dua
audit event tersimpan. `go test -race ./...`, `go vet ./...`, 8 frontend tests, production
build, OpenAPI YAML, container build, migration v13, serta smoke test HTTP lulus.

Change timeline (commit `598501f`) memuat seluruh normalized version pada detail API dalam urutan terbaru,
menyertakan observed time, lifecycle, snapshot/fingerprint trace, dan daftar field yang
berubah terhadap versi sebelumnya. Drawer membedakan initial observation, repeated
observation, dan material field changes. Acceptance detail API memastikan history count
sama dengan version count; perhitungan perubahan lifecycle, budget, dan deadline tercakup
unit test. Go race/vet, 8 frontend tests, build, OpenAPI, dan container acceptance lulus.

Ownership (commit `4b65a28`) memakai migration v14 dengan satu assignee aktif per opportunity, active-team
selector, serta RBAC administrator/operator/business-development/bid-manager untuk assign
dan unassign. Setiap perubahan menyimpan actor dan audit event; detail drawer menampilkan
owner saat ini. Acceptance API assign lalu unassign berhasil, assignment kembali null,
dan event `opportunity.assigned` serta `opportunity.unassigned` tersimpan. Go race/vet,
8 frontend tests, build, OpenAPI, container build, dan migration acceptance lulus.

Internal comments (commit `dd94328`) memakai migration v15 dan menyimpan komentar immutable dengan author
serta timestamp dalam timeline kronologis. Semua user terautentikasi dapat membaca dan
menambahkan komentar tervalidasi maksimal 4000 karakter; setiap create menghasilkan audit
event yang merujuk comment ID. Detail drawer memuat diskusi dan safe error state. Acceptance
API create/list dan audit lulus; record komentar acceptance kemudian dibersihkan dari database
lokal. Go race/vet, 8 frontend tests, build, OpenAPI, dan container acceptance lulus.

Shortlist (commit `baefc21`) memakai migration v16 sebagai status bersama per opportunity dengan actor dan
timestamp. API mendukung get/toggle, feed mendukung filter `shortlisted`, dan saved views
menyimpan filter tersebut. Detail drawer menyediakan toggle yang auditabel. Acceptance
add → filtered feed → remove berhasil; feed shortlist memuat tepat opportunity acceptance
dan audit `opportunity.shortlisted`/`opportunity.unshortlisted` tersimpan. Go race/vet,
8 frontend tests, build, OpenAPI, migration, dan container acceptance lulus.

Human pursue/pass decision memakai migration v17 dan sepenuhnya terpisah dari hasil rules,
score, maupun AI. Workflow role dapat menetapkan `pursue` atau `pass` hanya dengan alasan
3–2000 karakter; current state menyimpan decider dan timestamp. Setiap perubahan diaudit
dengan previous/new decision dan reason. Acceptance membuktikan alasan kosong ditolak HTTP
422, transisi undecided → pursue → pass menghasilkan metadata audit yang tepat, lalu state
acceptance dibersihkan. Go race/vet, 8 frontend tests, build, OpenAPI, dan container lulus.

### Public SPSE adapter

- [x] Policy review public listing SPSE INAPROC Maluku; robots.txt mengizinkan `/malukuprov`
- [x] Rate limiter dan robots/terms documentation
- [x] HTTP timeout, response-size limit, dan bounded retry/backoff untuk kegagalan transien
- [x] Parser fixture dari response yang sudah direduksi/redacted
- [x] Adapter regression tests
- [x] Incremental discovery via fingerprint; detail HTTP 403 tidak diambil
- [x] Server-side keyword search, pagination terbatas, dan deduplikasi lintas keyword
- [x] Source-specific lag/error metrics
- [x] Validasi tidak ada login, CAPTCHA bypass, atau supplier transaction
- [~] Lima source aktif (Maluku, Komdigi, Jakarta, Jawa Barat, Jawa Timur);
  perluas bertahap menuju 20 priority sources setelah stability observation

Target pertama dibatasi ke listing dan DataTables route yang dideklarasikan halaman
publik Maluku. Detail HTTP 403 tidak diakses ulang atau diakali; canonical URL tetap
disimpan untuk verifikasi manual.

Bukti: commit `0cb78f3`, `a7f5b24`; migration v7; live run `malukuprov-public-v1-acceptance` sukses
pada satu attempt dan menyimpan satu paket berstatus `completed`, sehingga tidak muncul
pada active-registration feed. Increment v2 memakai migration v8; live run
`malukuprov-public-v2-acceptance` sukses pada satu attempt dengan metrics 7 keyword,
7 halaman, 0 raw row, dan 0 kandidat. Hasil nol mencerminkan tidak adanya kecocokan
keyword IT saat run, bukan kegagalan transport atau parser.

Batch source pertama (commit `45a8568`) memakai migration v10-v11. Robots.txt secara eksplisit mengizinkan
seluruh tenant terpilih. Acceptance `source-batch-v3-*` sukses satu attempt per source:
Komdigi 38 kandidat, Jakarta 130, Jawa Barat 2, dan Jawa Timur 0. Parser mempertahankan
nilai kontrak yang belum tersedia sebagai unknown, membersihkan badge HTML dari judul,
dan memfilter keyword berdasarkan batas kata. Total database menjadi 171 paket publik
nyata; 70 terklasifikasi relevant. Paket IT nyata pada batch ini mayoritas sudah selesai
dan hanya tampil bila lifecycle `All` dipilih.

Increment operasi (commit `62946b7`) pada 10 Agustus 2026 menambahkan migration v12, scheduler idempoten enam
jam untuk lima source publik aktif, serta source health console dengan health score,
status dan metrik run terakhir, crawl lag, queue depth, dan safe error. Acceptance
scheduler menghasilkan tepat satu run sukses per source pada satu attempt tanpa snapshot
duplikat; fixture tetap tidak dijadwalkan. Go test/vet, 7 frontend tests, production
build, container build, dan smoke test HTTP lulus.

### Notifications dan operations

- [ ] Alert rules untuk opportunity/material changes
- [ ] Quiet hours
- [ ] Notification deduplication
- [ ] Email adapter
- [ ] Webhook adapter
- [x] Source health console
- [x] Queue/lag/error dashboard
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
| `28f2972` | Fondasi AI classification yang provider-neutral, tervalidasi, cached, dan budgeted |
| `62946b7` | Scheduler source publik dan source health/lag/queue console |
| `bca56f7` | Candidate golden dataset dan classification evaluation regression gate |
| `1d9d929` | Personal saved views dengan combined filters dan audit create/delete |
| `598501f` | Full opportunity change timeline dengan material field diff |
| `4b65a28` | Opportunity ownership assignment dengan team selector, RBAC, dan audit |
| `dd94328` | Internal opportunity comments dengan author, timeline, dan audit |
| `baefc21` | Shared opportunity shortlist dengan feed filter dan audit toggle |
| `45a8568` | Batch lima source SPSE dan parser nilai/keyword yang lebih tangguh |

## Aturan pembaruan task list

1. Pindahkan status hanya setelah acceptance criteria dan test relevan lulus.
2. Tambahkan commit/bukti pada bagian yang selesai.
3. Catat blocker langsung pada item menggunakan `[!]`.
4. Jangan menyimpan password, token, atau secret di dokumen ini.
5. Jika scope berubah, perbarui task list dan roadmap pada commit yang sama.
