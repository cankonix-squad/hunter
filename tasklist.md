# CANKONIX Opportunity Intelligence — Task List

Dokumen ini adalah sumber jejak kerja utama proyek. Perbarui status, bukti
verifikasi, dan commit terkait setiap kali satu perubahan selesai.

## Konvensi status

- `[x]` selesai dan sudah diverifikasi
- `[~]` sudah tersedia sebagian, masih membutuhkan pekerjaan lanjutan
- `[ ]` belum dikerjakan
- `[!]` terblokir; alasan wajib ditulis di item tersebut

## Kondisi terakhir

- Terakhir diperbarui: 13 Agustus 2026
- Branch: `main`
- Repository: `https://github.com/cankonix-squad/hunter`
- Stack lokal: React + TypeScript, Go, PostgreSQL, Docker Compose, Nginx
- Aplikasi lokal: `http://localhost:3000`
- Target URL production: `https://hunter.cankonix.com` (belum dinyatakan live)
- Migrasi database aktif: versi 30
- Data acceptance: 192 opportunity canonical
- Classification acceptance: 1 relevant dengan evidence dan 1 needs-review
- Scoring acceptance: 2 score dengan masing-masing 5 dimension breakdown
- Public crawl acceptance: 40 tenant source aktif dan terjadwal
- Runtime lokal aktif: PostgreSQL, API, collector, notifier, web, dan edge
- Local acceptance: 6/6 service berjalan; health, readiness, auth boundary, metrics,
  migrasi, source schedule, lifecycle, dan recent dead-letter gate lulus

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
- [x] Acceptance gate lokal terpadu melalui `scripts/local-acceptance.sh`
- [x] Authenticated opportunity workflow integration: feed → detail → shortlist → decision → next action → activity
- [x] Browser workflow component test dan live audit refresh setelah mutation
- [x] Konsistensi live audit untuk assignment, komentar, readiness, dan human override
- [x] Full opportunity workflow HTTP integration dengan tujuh mutation auditable
- [x] Workflow reversal integration untuk unassign, unshortlist, dan clear state
- [x] Browser reversal acceptance dengan payload null, state reset, dan live audit
- [x] Pipeline analytics live refresh setelah workflow mutation
- [x] Pipeline analytics integration delta sebelum, sesudah, dan reversal workflow
- [x] Contextual empty-search guidance dan one-click lifecycle expansion

Bukti local acceptance 13 Agustus 2026: 6/6 service berjalan; PostgreSQL, API, dan collector
healthy; root/health/readiness HTTP 200; protected API HTTP 401 tanpa autentikasi; migrasi v29;
20 source aktif terjadwal; metrik internal sehat; lifecycle valid; nol source dead-letter dalam
satu jam terakhir. Notifier berjalan tetapi delivery tetap disabled secara sengaja di local.
Full local quality rerun juga lulus: Go formatting/module verification/race tests/vet, golden
dataset 15/15, frontend 8/8 tests, production build, dan npm audit 0 vulnerability.
Workflow integration memakai PostgreSQL disposable bermigrasi v1–v29 dan membuktikan seluruh
mutation menghasilkan state serta audit activity yang benar tanpa mengubah database aplikasi lokal.
React acceptance membuktikan payload shortlist/decision/next-action, konversi deadline Jakarta ke
UTC, state UI, dan audit activity yang langsung diperbarui; frontend 9/9 tests dan production build
lulus, image web lokal direbuild, lalu full local acceptance kembali PASS.
Seluruh mutation detail opportunity kini merefresh audit secara best-effort: kegagalan refresh tidak
menyamarkan mutation yang telah sukses. Component acceptance diperluas untuk readiness, override,
dan komentar; frontend 9/9, build, rebuild container, dan local acceptance tetap lulus.
Backend workflow acceptance pada PostgreSQL disposable v1–v29 membuktikan assignment, shortlist,
decision, next action, comment, readiness, dan override melalui production HTTP handlers, termasuk
unified audit event masing-masing; race test lulus dan database utama tidak dimutasi.
Reversal acceptance membuktikan unassign, unshortlist, clear next action, dan clear override benar-benar
menghapus state target serta mencatat empat audit event, tanpa menghapus decision, comment, atau
readiness yang tidak termasuk aksi pembatalan.
React acceptance membuktikan assign/unassign, shortlist/unshortlist, next-action save/clear, dan
override save/clear memakai payload yang tepat, mengembalikan UI ke state kosong, dan langsung
menampilkan audit reversal; full frontend 9/9 dan production build lulus.
Pipeline analytics kini direfresh best-effort setelah assignment, shortlist, decision, dan next-action
mutation termasuk reversal; kegagalan background refresh mempertahankan ringkasan terakhir dan tidak
menyamarkan mutation sukses. Component test memverifikasi seluruh delapan analytics request lifecycle.
HTTP integration acceptance membandingkan analytics terhadap baseline disposable: assigned,
shortlisted, pursue, undecided, dan due-seven-days berubah tepat setelah mutation; assignment,
shortlist, dan deadline kembali ke baseline setelah reversal sementara decision tetap persisten.
Empty feed kini menjelaskan query/filter penyebab nol hasil, menawarkan reset seluruh filter, dan
mempertahankan keyword saat memperluas dari active registration ke all lifecycle. Regression test
memakai contoh `pengembangan sistem`; frontend 10/10 dan production build lulus.

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
- [x] Next action dan deadline
- [x] Readiness checklist
- [x] Human override untuk recommendation/score
- [x] Audit log seluruh decision dan override
- [x] Initial pipeline analytics

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

Human pursue/pass decision (commit `bc7e121`) memakai migration v17 dan sepenuhnya terpisah dari hasil rules,
score, maupun AI. Workflow role dapat menetapkan `pursue` atau `pass` hanya dengan alasan
3–2000 karakter; current state menyimpan decider dan timestamp. Setiap perubahan diaudit
dengan previous/new decision dan reason. Acceptance membuktikan alasan kosong ditolak HTTP
422, transisi undecided → pursue → pass menghasilkan metadata audit yang tepat, lalu state
acceptance dibersihkan. Go race/vet, 8 frontend tests, build, OpenAPI, dan container lulus.

Next action (commit `03bca39`) memakai migration v18 dengan satu action aktif, UTC deadline, updating actor,
serta timestamps per opportunity. Workflow role dapat update atau clear; UI menerima waktu
Asia/Jakarta, mengirim UTC, dan menandai action overdue. Setiap perubahan menghasilkan audit
event dengan action/deadline atau clear event. Acceptance update → clear mempertahankan UTC
deadline yang tepat dan menyimpan kedua audit event. Go race/vet, 8 frontend tests, build,
OpenAPI, migration, dan container acceptance lulus.

Readiness checklist (commit `f2b9fbd`) memakai migration v19 dan checklist version `readiness-v1` dengan lima
item stabil: scope, eligibility, delivery capacity, commercial review, dan document plan.
Setiap item memiliki completion state, note, updater, timestamp, serta audit metadata;
API mengembalikan progress aggregate dan UI menampilkan progress/toggle/note. Acceptance
menyelesaikan `scope_validated`, menghasilkan progress 1/5 dan audit yang tepat, lalu data
acceptance dibersihkan. Go race/vet, 8 frontend tests, build, OpenAPI, migration, dan
container acceptance lulus (satu retry build diperlukan setelah DNS registry transien).

Human override (commit `0a961df`) memakai migration v20 dan menyimpan recommendation dan/atau score override
sebagai layer terpisah tanpa mengubah output deterministic/AI asli. Reason 3–2000 karakter,
actor, timestamp, RBAC, update, dan clear semuanya diaudit. Detail drawer menandai override
secara eksplisit terpisah dari model output. Acceptance menyimpan `relevant`/8200 bp lalu
clear, dan kedua audit event tersimpan. Go race/vet, 8 frontend tests, build, OpenAPI,
migration, dan container acceptance lulus.

Opportunity activity API/UI (commit `c139fec`) menyatukan audit assignment, comments, shortlist, decision,
next action, readiness, serta override dalam timeline newest-first, dengan actor, timestamp,
dan metadata aman; hasil dibatasi 200 event per detail. Acceptance memuat 12 event dari 11
jenis action workflow dengan actor yang benar. Go race/vet, 8 frontend tests, build,
OpenAPI, dan container acceptance lulus.

Initial pipeline analytics (commit `2281fd7`) merangkum total dan active opportunity, shortlist, assignment,
pursue/pass/undecided, overdue next action, serta action jatuh tempo tujuh hari melalui
endpoint aggregate terautentikasi. Dashboard menampilkan portfolio overview yang responsif
dan menandai overdue risk. Acceptance API menghasilkan 173 total dan 3 active, sama dengan
agregat langsung Postgres; Go race/vet, 8 frontend tests, production build, OpenAPI YAML,
container build, dan health check lulus.

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
- [~] Dua puluh source aktif (Maluku, Komdigi, Jakarta, Jawa Barat, Jawa Timur,
  Jawa Tengah, DI Yogyakarta, Banten, Sumatera Utara, Sulawesi Selatan,
  Sumatera Barat, Sumatera Selatan, Kalimantan Barat, Kalimantan Timur, Bali,
  NTB, NTT, Sulawesi Utara, Sulawesi Tengah, Sulawesi Tenggara); target coverage
  tercapai, menunggu stability observation tujuh hari
- [x] Dua puluh tenant LPSE kementerian/pusat unik dari direktori resmi LKPP ditambahkan
  tanpa menduplikasi tenant bersama; total registry publik menjadi 40 tenant

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

Batch source kedua (commit `ebf1362`) memakai migration v21 dan menambah Jawa Tengah, DI Yogyakarta,
Banten, Sumatera Utara, serta Sulawesi Selatan. Review 10 Agustus 2026 memastikan
robots.txt resmi mengizinkan seluruh tenant dan kelima listing publik merespons HTTP
200. Scheduler menghasilkan satu run per source; semuanya sukses pada attempt pertama:
Jawa Tengah 0 kandidat, DI Yogyakarta 2, Banten 2, Sumatera Utara 3, dan Sulawesi
Selatan 0. Seluruh tujuh paket yang ditemukan tersimpan sebagai snapshot baru; tidak
ada safe error dan paket nonaktif tetap tidak masuk active feed.

Batch source ketiga (commit `1f59e59`) memakai migration v22 dan menambah Sumatera Barat, Sumatera
Selatan, Kalimantan Barat, Kalimantan Timur, serta Bali. Review 11 Agustus 2026
memastikan robots.txt resmi mengizinkan seluruh tenant dan kelima listing merespons
HTTP 200. Kelima scheduler run sukses pada attempt pertama tanpa safe error: Bali dan
Kalimantan Timur masing-masing menghasilkan satu snapshot completed, sedangkan tiga
source lain valid dengan hasil nol. Total registry menjadi tepat 15 source SPSE aktif;
dua paket nonaktif yang ditemukan tidak masuk active feed.

Batch source keempat (commit `77d53ac`) memakai migration v23 dan menambah NTB, NTT, Sulawesi Utara,
Sulawesi Tengah, serta Sulawesi Tenggara. Review 11 Agustus 2026 memastikan robots.txt
resmi mengizinkan seluruh tenant dan kelima listing merespons HTTP 200. Semua scheduler
run sukses pada attempt pertama tanpa safe error: NTT menghasilkan satu snapshot
completed dan Sulawesi Tengah lima snapshot (satu completed, empat evaluation); tiga
source lainnya valid dengan hasil nol. Registry mencapai tepat 20 source SPSE aktif dan
seluruh enam paket nonaktif tetap tidak masuk active feed. Completion gate tetap menunggu
observasi stabilitas tujuh hari.

Pagination hardening 13 Agustus 2026 mempertahankan kandidat halaman sebelumnya ketika endpoint
DataTables mengembalikan HTTP 404 pada halaman lanjutan dan mencatat `pagination_end_404` sebagai
akhir pagination yang dapat diobservasi. HTTP 404 pada halaman pertama tetap menjadi permanent
source failure. Regression test kedua jalur, full Go race test, vet, serta collector image build
lulus; collector lokal hasil rebuild healthy. Canary Jakarta berikutnya menunggu scheduler atau
credential operator lokal untuk antrean melalui API resmi; dead-letter tidak diubah langsung.

Ekspansi kementerian migration v30 menambahkan 20 tenant unik yang tercantum pada kategori kementerian
direktori resmi LKPP. Seluruh tenant eksplisit diizinkan robots.txt INAPROC dan listing `/lelang`
merespons HTTP 200 pada 13 Agustus 2026. Direktori memuat entri warisan/duplikat dan beberapa
kementerian berbagi tenant `nasional`, `lkpp`, atau provider lain, sehingga registry menyimpan satu
source per tenant, bukan satu source duplikat per nama kementerian. Total menjadi 40 tenant publik
aktif dan terjadwal; 20 source lama tetap ditandai sebagai baseline priority untuk stability gate.
Scheduler mengantrekan 20 canary tenant baru dan seluruhnya sukses tanpa dead-letter; hasil valid
bervariasi dari nol kandidat hingga 60 kandidat dan snapshot baru tersimpan secara idempoten.

Increment operasi (commit `62946b7`) pada 10 Agustus 2026 menambahkan migration v12, scheduler idempoten enam
jam untuk lima source publik aktif, serta source health console dengan health score,
status dan metrik run terakhir, crawl lag, queue depth, dan safe error. Acceptance
scheduler menghasilkan tepat satu run sukses per source pada satu attempt tanpa snapshot
duplikat; fixture tetap tidak dijadwalkan. Go test/vet, 7 frontend tests, production
build, container build, dan smoke test HTTP lulus.

### Notifications dan operations

- [x] Alert rules untuk opportunity/material changes
- [x] Quiet hours
- [x] Notification deduplication
- [x] Email adapter
- [x] Webhook adapter
- [x] Source health console
- [x] Queue/lag/error dashboard
- [x] Dead-letter replay workflow
- [x] Failure recovery runbook

Alert rules (commit `9dd7dbd`) memakai migration v24 dan mendukung event `opportunity.new` serta
`opportunity.material_change` dengan filter source, relevance, dan minimum score.
Rule dibatasi per user, dapat diaktifkan/nonaktifkan, dan create/update/delete diaudit.
Collector mencocokkan rule secara atomik setelah classification/scoring tersimpan;
event immutable hanya dibuat untuk opportunity version baru, sementara fingerprint
duplikat dan perubahan normalized non-material tidak memicu event. Dashboard menyediakan
rule manager dan inbox 200 event terbaru; delivery tetap `pending` untuk increment
deduplication/quiet-hours/adapter berikutnya. Acceptance create/list/delete lulus dengan
HTTP 204 dan dua audit event yang benar. Go race/vet, 8 frontend tests, production build,
OpenAPI YAML, migration v24, serta container build/health lulus.

Notification deduplication (commit `b2448c6`) memakai migration v25 dan unique invariant per user,
opportunity version, serta event type. Matcher memilih tepat satu rule paling spesifik
per user dengan tie-break deterministik sebelum insert; constraint database tetap menjadi
pengaman terhadap concurrent collector dan retry. Acceptance dalam transaksi rollback
membuat dua rule overlap dan membuktikan hanya satu event tersimpan. Go race/vet lulus.

Quiet hours (commit `2f36033`) memakai migration v26 dan preferensi per user untuk enabled state, waktu
mulai/akhir, serta IANA timezone. Rentang same-day dan overnight didukung; matcher tetap
mencatat event segera tetapi mengisi `available_after` pada akhir waktu hening lokal,
sehingga delivery worker kelak hanya mengambil event pending yang sudah due. Update API
bersifat atomik dengan audit log dan dashboard menyediakan kontrol WIB/WITA/WIT.
Acceptance mengaktifkan rentang 00:00–23:59 Asia/Jakarta, membuktikan waktu defer berada
di masa depan, lalu mengembalikan setting ke disabled 22:00–07:00. Go race/vet, 8 frontend
tests, production build, OpenAPI YAML, migration v26, dan seluruh container build lulus.

Email adapter (commit `6dc92a8`) memakai migration v27 dan notifier process terpisah dengan transactional
outbox per alert event, database lease, expired-lease recovery, bounded exponential retry,
dead-letter setelah lima attempt, serta safe error maksimal 500 karakter. SMTP production
mewajibkan STARTTLS TLS 1.2+, mendukung auth opsional, dan mencegah subject header injection;
credential hanya berasal dari environment. Mode dry-run eksplisit memakai jalur outbox dan
status yang sama tanpa koneksi eksternal. Acceptance end-to-end menghasilkan delivery email
`delivered` pada attempt pertama dan memperbarui alert event secara atomik, kemudian seluruh
data uji dibersihkan dan notifier dikembalikan ke default nonaktif. Go race/vet, Compose
validation, migration v27, dan notifier container build lulus.

Webhook adapter (commit `74fe9e1`) memakai migration v28 dan shared notifier/outbox reliability path.
Payload JSON terstruktur menyertakan event, opportunity, changed fields, canonical URL,
dan timestamp; request memiliki deterministic idempotency key serta HMAC-SHA256 atas body
exact. URL production wajib HTTPS tanpa embedded credentials, secret minimal 32 karakter,
redirect ditolak, timeout dibatasi, response body dibuang secara bounded, dan status HTTP
dicatat tanpa menyimpan body. Acceptance dry-run menghasilkan delivery `webhook` HTTP 204
pada attempt pertama tanpa safe error, kemudian data uji dibersihkan dan notifier kembali
nonaktif. Go race/vet, signature/idempotency tests, Compose validation, migration v28, dan
notifier container build lulus.

Dead-letter replay (commit `4745e0b`) memakai migration v29 dan console administrator/operator untuk 200
failure terbaru. Replay hanya berlaku pada status `dead_letter`, mengantrekan ulang row
yang sama agar dedup identity terjaga, mereset attempts/lease/error/response, menaikkan
durable replay count, menyimpan actor dan timestamp, serta menulis audit event atomik.
Destination dan safe error terlihat untuk operasi, sedangkan lease/secret tidak masuk
kontrak JSON. Acceptance mengubah email dead letter attempt ke-5 menjadi `queued`, attempt
0, replay count 1, dan membuktikan audit metadata yang benar sebelum cleanup. Go race/vet,
8 frontend tests, production build, OpenAPI YAML, migration v29, dan container build lulus.

Failure recovery runbook (commit `f077593`) mendefinisikan severity/ownership, evidence capture,
triage read-only, stabilisasi writer, pemulihan API/edge/database, circuit breaker source,
canary source run, notifier/dead-letter replay, rollback forward-only, escalation criteria,
closure gate, dan template incident record. Semua recovery mutation memakai endpoint RBAC
yang sudah tersedia; dokumen melarang bypass upstream, perubahan queue langsung, rollback
migration, dan penghapusan volume. Acceptance memvalidasi referensi route terhadap router dan
OpenAPI, Compose config, query diagnosis pada migration v29, seluruh container aktif, serta
health/readiness melalui edge HTTP 200.

### Hardening dan release

- [~] Structured logs, health/readiness, HTTP metrics, dan staging alerts tersedia; traces/exporter lanjutan
- [~] CI pipeline tersedia dan lulus lokal; remote GitHub tercatat, menunggu hosted run pertama
- [x] API authorization/contract integration suite
- [x] Fresh-database migration test dari v1 sampai versi terbaru
- [x] End-to-end discover → normalize → classify → score → review test
- [x] Security review dan dependency/container scanning
- [x] Performance dan pagination load test
- [x] Backup dan restore drill
- [x] Production secret management
- [x] TLS/staging deployment configuration
- [~] UAT/deployment runbook dan automated preflight tersedia; menunggu hosted UAT + approval
- [~] Production domain routing/TLS — overlay Traefik tervalidasi; menunggu parameter dan deployment host
- [x] Traefik deployment-host preflight dan environment template non-secret
- [x] Immutable release image overlay terikat ke digest preflight
- [!] Tujuh hari stability observation — baseline 6–12 Agustus gagal; mulai ulang setelah DNS stabil

Production domain diagnosis 13 Agustus 2026: `hunter.cankonix.com` resolve dan menerima TLS, tetapi
menyajikan self-signed `TRAEFIK DEFAULT CERT`; request diagnostik tanpa verifikasi mendapat HTTP 404.
Gate baru `scripts/check-production-domain.sh` tetap fail-closed dan mewajibkan CA-valid HTTPS serta
HTTP 200 untuk root, health, dan readiness. Integrasi router/certificate resolver Traefik harus
ditentukan pada host; nama network/resolver tidak diasumsikan di repository.
Stability checkpoint 13 Agustus juga gagal setelah Jakarta menerima HTTP 404 pada page 2 keyword
`sistem`; adapter telah di-hardening, tetapi hari tersebut tetap tidak dapat dihitung sebagai hari
bersih dan window tujuh hari harus dimulai ulang setelah canary sukses.
Overlay `compose.traefik.yaml` kini menghapus publikasi port local 3000, menghubungkan hanya edge ke
external Traefik network, mengatur Host router/TLS/service port, dan mewajibkan nama network serta
certificate resolver nyata. `scripts/check-traefik-config.sh` membuktikan merged production Compose
tidak membawa development secret dan lulus dengan target `hunter.cankonix.com`.
Edge khusus Traefik mempertahankan CSP/HSTS/security headers dan HTTPS proxy semantics; router API
memakai rate limit 20/s burst 40, sedangkan login memakai priority lebih tinggi dan 5/menit burst 3.
Merged Compose contract serta `nginx -t` lulus tanpa membuka port host.
Template `.env.traefik.example` hanya menyimpan hostname, identifier infra, dan path secret. Host
preflight mewajibkan hostname production tepat, secret file non-empty dengan mode 0400/0600,
external Docker network nyata, Compose tanpa published port/development credential, dan tidak
melakukan deployment. Acceptance lokal membuktikan network inspect dan merged render secara terpisah
dengan secret dummy sementara yang kemudian dihapus.
Overlay `compose.release.yaml` menonaktifkan seluruh application `build:`, mewajibkan empat
`registry/image@sha256` reference, dan selalu menarik artifact immutable. Release preflight memastikan
setiap ref berakhir pada digest evidence pasangannya; host-preflight menolak ref/tag mutable dan build
directive. Contract render serta full host-preflight dengan dummy digest lulus.

CI release checks (commit `8d395f2`) menambahkan workflow GitHub Actions least-privilege dengan
concurrency cancellation dan empat job paralel: Go format/module verification/race test/vet,
frontend test/build/dependency audit, migration PostgreSQL 16 kosong, serta Compose validation
dan build image API/collector/notifier/web. Skrip reusable memverifikasi nama migration berurutan,
menerapkan seluruh file dengan `ON_ERROR_STOP`, lalu mewajibkan count dan versi maksimum sama.
Acceptance lokal: Go race/vet lulus, 9 frontend tests dan production build lulus, npm audit 0
vulnerability, database disposable berhasil dari v1 sampai v29 lalu dihapus, YAML/shell syntax
dan Compose config valid, serta seluruh image berhasil dibangun. Status CI tetap parsial sampai
autentikasi GitHub tersedia dan satu hosted run terbukti hijau. Remote `origin` sudah mengarah ke
repository resmi, tetapi push pertama 13 Agustus 2026 berhenti aman sebelum upload karena mesin belum
memiliki HTTPS credential, GitHub CLI, maupun SSH key.

API authorization/contract suite (commit `5159e0e`) menjalankan handler produksi melalui
`httptest` terhadap PostgreSQL 16 yang dimigrasikan penuh dan memakai credential test acak.
Coverage membuktikan health/security header/request ID, login bearer response, 401 untuk token
hilang/invalid, authenticated read bagi user tanpa privileged role, 403 viewer/operator pada
source administration, akses administrator untuk create/update source, akses operator untuk
durable source-run queue dan dead-letter console, serta safe JSON error envelope untuk unknown
field dan pagination invalid. Suite memakai build tag `integration` dan menjadi gate job migration
CI. Acceptance `go test -race -tags=integration ./internal/httpapi` lulus pada database disposable
yang kemudian dihapus; seluruh regression `go test -race ./...`, `go vet ./...`, YAML, dan diff
check tetap lulus.

Collector workflow E2E (commit `ef22ccd`) memakai embedded SPSE fixture, runner produksi, API
handler produksi, credential test acak, dan PostgreSQL 16 yang dimigrasikan penuh. Discovery
pertama menghasilkan tepat 2 snapshot, canonical opportunity, version, classification, dan score,
serta 10 score dimension. Feed relevant dengan minimum score 7000 mengembalikan Integrasi Data
sebagai `relevant/data-ai`, score 8000, satu evidence, lima dimension, dan satu history version;
paket jaringan ambigu muncul tepat sekali pada review API sebagai `needs_review`. Discovery kedua
dengan idempotency key baru sukses tetapi `new_snapshots=0`; snapshot/version/classification/score/
dimension count tidak berubah. Suite menjadi bagian gate integration CI. Acceptance race test
integration dan seluruh Go regression/vet lulus; database/container disposable dihapus otomatis.

Security release gate (commit `90b7e32`) menemukan vulnerability fixable HIGH/CRITICAL pada
image lama (Go 1.23.12, pgx 5.7.2, x/crypto 0.36.0, dan base nginx lama), lalu memperbarui Go
ke 1.25, jwt 5.3.1, pgx 5.10.0, x/crypto 0.54.0, Alpine 3.23, Node 24, Nginx 1.29, dan
PostgreSQL Compose 16.10. Build dibuat reproducible dengan `go.sum` serta `npm ci`. CI kini
menjalankan govulncheck v1.6.0 dan Trivy release terverifikasi yang dipin full commit untuk
filesystem vulnerability/misconfiguration/secret serta keempat application image; tidak ada
exception suppression. Acceptance: govulncheck 0 reachable vulnerability, npm audit 0,
Docker Scout 0 fixable HIGH/CRITICAL pada API/collector/notifier/web hasil rebuild, Go race/vet,
8 frontend tests/build, Compose validation, dan runtime hardened health/readiness HTTP 200.
Review dan residual production blocker dicatat di `docs/SECURITY_REVIEW.md`.

Performance/pagination gate (commit `f8198f0`) mengisi 10.000 canonical opportunity secara
set-based pada PostgreSQL disposable, menjalankan `ANALYZE`, dan memverifikasi first/last page,
total pages, exact search, 20% review subset, serta cap `page_size` 100. Load API terautentikasi
menjalankan 40 mixed-page request dengan 8 worker dan budget p95 750 ms, mencakup JWT, count,
classification/score joins, pagination, dan JSON serialization. Race-enabled standalone run
mencatat p50 275 ms, p95 390 ms, max 408 ms; seluruh integration job bersama E2E/RBAC mencatat
p50 309 ms, p95 581 ms, max 637 ms dan tetap lulus. Gate masuk CI; batas interpretasi dan
trigger evaluasi cursor pagination dicatat di `docs/PERFORMANCE.md`. Seluruh Go regression/vet
lulus dan database/container pengujian dihapus otomatis.

Backup/restore drill (commit `797325e`) menambahkan custom-format PostgreSQL backup dengan
writer-quiesced guard, temporary-file atomic publication, SHA-256, migration state, dan manifest
count seluruh public table. Restore menolak target yang tidak kosong, memverifikasi checksum dan
archive list sebelum `pg_restore --exit-on-error`, lalu mewajibkan schema migration, seluruh table
count, lifecycle, dan core foreign-key integrity cocok. Workflow CI menjalankan drill setelah
integration suite. Acceptance pada dua PostgreSQL 16.10 disposable membuktikan backup tanpa
acknowledgement ditolak, restore mencapai schema 29:29, semua table count cocok, serta fixture user
dan audit event pulih; artefak sementara dan kedua container dibersihkan otomatis. Kebijakan daily
encrypted external backup, retention 30 hari, cutover, dan escalation dicatat di
`docs/BACKUP_RESTORE.md`; scheduling/storage production tetap bagian deployment.

Production secret boundary (commit `40d0882`) menambahkan resolver file `*_FILE` untuk database,
JWT, bootstrap administrator, SMTP, dan webhook tanpa memasukkan nilai ke error. Mode production
mewajibkan database/JWT file-backed, menolak sumber ambigu, credential development, secret lemah,
bootstrap parsial, serta credential notifier langsung. Overlay `compose.production.yaml` menghapus
nilai langsung hasil merge dan memasang Docker secrets; gate CI merender Compose dengan sentinel
host lalu memastikan nilai/fallback tidak muncul. Runbook mencatat provisioning, permission,
bootstrap sekali pakai, dan rotasi tanpa menyimpan secret nyata. Acceptance: production boundary
script dan Compose config lulus, `go test -race ./...`, `go vet ./...`, 8 frontend tests/build,
shell/diff check, serta image API/collector/notifier berhasil dibangun.

TLS/staging configuration (commit `621cd44`) menambahkan overlay Compose berlapis untuk hostname
HTTPS dan certificate/private-key Docker secrets, hanya mempublikasikan edge 80/443, serta membuat
edge read-only dengan tmpfs terbatas. Nginx mewajibkan TLS 1.2/1.3, redirect HTTP 308, HSTS, CSP,
security headers, request/proxy timeout, API rate limit, dan login throttling; seluruh inbound
forwarding header ditimpa karena edge menjadi satu-satunya trusted proxy. Startup production juga
menolak `WEB_ORIGIN` non-HTTPS atau origin dengan path/query/credential. CI membuat certificate
sementara, merender tiga overlay tanpa development credential, lalu menjalankan `nginx -t`.
Acceptance lokal lulus untuk gate staging TLS, `go test -race ./...`, `go vet ./...`, 8 frontend
tests/build, shell/diff check, serta image API/collector/notifier. DNS/firewall, certificate CA,
renewal, dan hosted smoke evidence tetap wajib saat deployment aktual sesuai runbook staging.

UAT dan deployment release gates (commit `20c96d8`) menyediakan sembilan acceptance scenario
ber-ID untuk role/access, discover idempoten, lifecycle, review evidence, workflow manusia, alert,
operations recovery, audit, dan performance beserta evidence aman serta sign-off product,
operations, dan security. Deployment runbook mengatur ownership, immutable artifact, hosted CI,
secret/TLS/firewall preflight, verified backup, forward migration, deployment, canary, observability,
rollback threshold, dan production promotion. `staging-smoke.sh` memverifikasi CA-trusted HTTPS,
health/readiness, unauthenticated 401, serta enam security header tanpa menerima URL credential;
CI memvalidasi syntax, section contract, ID unik, dan approval fields. Acceptance repository lulus
melalui `check-uat-runbook.sh`, shell syntax, dan diff check. Item tetap parsial sampai staging nyata
menjalankan seluruh scenario dan tiga approver mencatat keputusan GO; tidak ada approval direkayasa.

Seven-day stability gate (commit `6560b56`) menghitung evidence langsung dari PostgreSQL untuk
tepat 20 source aktif-terjadwal, 140 source-day sukses, aggregate success rate minimal 95%, nol
dead-letter, nol consecutive failure/open circuit, dan nol audit intervensi konfigurasi. Report
memisahkan `INCOMPLETE` dari `FAILED`, menerima credential database via file, hanya mengeluarkan
agregat non-secret, serta divalidasi CI terhadap schema migration kosong agar durasi tidak dapat
dipalsukan. Baseline read-only untuk 6–12 Agustus 2026 menemukan 20 source, 55/140 source-day
tercakup, 108/139 run sukses (77,70%), 31 dead-letter, dan satu source update sehingga status benar
`FAILED`. Seluruh dead-letter berasal dari kegagalan resolver Docker terhadap `spse.inaproc.id`;
resolver kembali berhasil dan run sukses terakhir tercatat 12 Agustus 12:31 UTC, tetapi window
bersih tujuh hari harus dimulai ulang setelah kegagalan terakhir 12 Agustus 01:37 UTC. Gate contract,
shell syntax, SQL PostgreSQL 16, dan diff check lulus; status tidak dinaikkan tanpa tujuh hari nyata.

Transient collector recovery hardening (commit `a93c7f5`) menutup akar dead-letter massal saat
resolver Docker sempat gagal. Transport/DNS/HTTP 5xx kini memperoleh enam attempt dengan exponential
backoff 30 detik sampai 16 menit dan cap 30 menit; circuit breaker tetap membatasi outage lebih lama.
HTTP 4xx termasuk 403/429 diberi typed permanent policy error dan langsung dead-letter agar akses
yang dilarang tidak diulang. Unit test membuktikan backoff/cap dan error type tetap terdeteksi setelah
wrapping; seluruh `go test -race ./...` dan `go vet ./...` lulus. Image collector dibangun ulang dan
diterapkan; seluruh container aktif, PostgreSQL 16.10 sehat, edge health/readiness HTTP 200, resolver
collector kembali menjawab, dan log startup bersih. Jika tidak ada failure/intervensi baru, window
UTC bersih paling awal adalah 13–19 Agustus 2026 dan baru dapat dievaluasi setelah hari terakhir.
Checkpoint 13 Agustus saat hari UTC masih berjalan: 20 priority source, 20/20 source sukses,
20/20 run sukses, nol dead-letter; status benar `PENDING` hanya karena `day_complete=0`.

MVP evidence reconciliation (commit `0c85b68`) memperkuat integration E2E dengan menyisipkan enam
lifecycle nonaktif (`registration_closed`, `evaluation`, `awarded`, `contracted`, `completed`, dan
`cancelled`) lalu mewajibkan active API tetap hanya mengembalikan dua fixture `registration_open`.
Gate lifecycle, explainable human override, dan alert dedup kini ditandai selesai berdasarkan test,
API/UI acceptance, audit behavior, serta unique database invariant yang sudah ada; pemetaan bukti
dan external acceptance dicatat di `docs/MVP_GATE_EVIDENCE.md`. Integration-tag compile, seluruh Go
race test/vet, dan diff check lulus. Gate classification tetap terbuka walaupun candidate dataset
mencapai macro precision 100%, karena domain owner belum menyetujui label dan coverage-nya.

Classification approval hardening (commit `1323806`) menambahkan review packet Markdown deterministik
untuk seluruh 15 case dan prosedur koreksi/coverage review oleh domain owner. Dataset berstatus
`approved` kini wajib memuat nama/role reviewer, timestamp RFC3339 non-future, dan approval reference;
candidate dilarang membawa metadata approval prematur. Release script mewajibkan status approved dan
macro precision minimal 8000 bp, sedangkan CI terus menjaga candidate precision tanpa mengarang
approval. Acceptance menghasilkan review packet 15 baris dan precision 10000 bp, unit/full Go race
test serta vet lulus, dan release gate menghasilkan `BLOCKED` karena status masih candidate—hasil
yang benar sampai reviewer CANKONIX memberikan keputusan nyata sesuai `docs/CLASSIFICATION_APPROVAL.md`.

Fail-closed release preflight (commit `c5587f9`) menggabungkan clean immutable commit, remote origin,
hosted CI untuk commit yang sama, HTTPS staging, tiga UAT approval reference, backup/restore evidence,
on-call owner, rollback commit, dan empat immutable image digest. Setelah metadata tersedia, command
menjalankan ulang production secret, staging TLS, UAT/runbook, stability contract, approved
classification, database seven-day report, dan CA-verified staging smoke; kegagalan child gate tidak
dapat jatuh-through menjadi GO. CI memvalidasi syntax dan seluruh required field. Acceptance contract
lulus dan kondisi repository saat ini menghasilkan `NO-GO` dengan blocker eksplisit: belum ada
remote/hosted evidence, approval, staging URL, artifact digest, on-call/rollback reference, serta gate
classification dan waktu stabilitas yang masih terbuka. Tidak ada nilai secret yang diterima sebagai
evidence metadata atau dicetak oleh preflight.

Staging monitoring (commit `75b06aa`) menambahkan internal `/metrics` dengan process uptime,
counter method/route/status, dan cumulative request-duration histogram. Label route memakai pattern
terbatas, bukan raw URL/ID, dan unit test membuktikan identifier request tidak bocor serta bucket
kumulatif benar. Overlay staging menjalankan Prometheus 3.13.1 LTS distroless dengan retention 15
hari, storage persistent, read-only root, dan UI hanya bind `127.0.0.1`; edge tidak mengekspos
metrics. Tiga rule awal mendeteksi API scrape down, sustained 5xx di atas 5%, dan p95 di atas satu
detik. `promtool` memvalidasi config dan tiga rule, staging Compose/Nginx tetap valid, seluruh Go
race test/vet lulus, API image dibangun/diterapkan, metrics internal merespons, dan edge
health/readiness HTTP 200. Hosted Alertmanager routing/canary, collector/database exporter, dan
distributed tracing tetap dicatat sebagai pekerjaan observability deployment, bukan diklaim selesai.

Operational metrics extension (commit `0851627`) mengekspor health score, lag, pending runs,
consecutive failures, dan circuit state per 20 source dengan bounded source ID/name labels, aggregate
run status, notification channel/status, serta exporter scrape-success. Query database dibatasi dua
detik agar scrape tidak menggantung API. Enam alert tambahan mencakup operational scrape failure,
source lag lebih dari 12 jam, queue buildup, open circuit, source dead-letter satu jam terakhir,
dan notification dead-letter; window recent mencegah 31 dead-letter historis membuat alarm permanen.
Acceptance runtime membaca seluruh source aktual, `source_dead_letters_recent=0` dan
`operational_metrics_scrape_success=1`; `promtool` memvalidasi total sembilan rule. Integration-tag
contract compile, seluruh Go race test/vet, API rebuild/deploy, dan internal metric smoke lulus.

Collector process monitoring (commit `500d70b`) menambahkan internal health/metrics server port
8081 dengan graceful shutdown dan fatal bind-error behavior. Metrics bounded mencatat uptime,
scheduler enabled, claimed/idle/error loop, scheduler ticks, queued runs, dan scheduler errors;
Compose kini memiliki collector Docker healthcheck. Prometheus men-scrape API dan collector secara
terpisah serta menambah alert collector-down dua menit dan scheduler error, sehingga process failure
tidak perlu menunggu source lag 12 jam. Acceptance: unit race test monitor, seluruh Go race/vet,
Compose config, `promtool` dua target/sebelas rule, image rebuild/deploy, internal metric smoke, dan
collector Docker `healthy` lulus; edge health/readiness tetap HTTP 200.

Daily stability checkpoint (commit `73bca10`) memisahkan diagnosis per tanggal UTC dari rolling
seven-day release gate. Laporan harian mewajibkan tepat 20 source, nol dead-letter, tanggal sudah
selesai, dan minimal satu run sukses per source; hasil dibedakan `PENDING`, `FAILED`, dan `PASSED`.
Audit timestamp mengoreksi interpretasi sebelumnya: dead-letter ke-32 terjadi 12 Agustus 18:44 UTC
pada Kalimantan Barat setelah HTTP 403, bukan 13 Agustus. Pada 13 Agustus 00:17 UTC seluruh service
sehat, terdapat 20 source, belum ada run/dead-letter hari itu, sehingga checkpoint benar `PENDING`
(hari UTC belum selesai), bukan gagal. Kandidat window 13–19 Agustus masih dapat berjalan jika
Kalimantan Barat kembali sukses dan tidak ada dead-letter/intervensi baru; seven-day gate tetap
satu-satunya bukti final.

## MVP completion gate

- [ ] Minimal 20 priority SPSE sources stabil selama tujuh hari
- [x] Active feed tidak memuat paket completed/cancelled/evaluation/award/contract
- [ ] Classification precision minimal 80% pada golden dataset yang disetujui
- [x] Recommendation dapat dijelaskan dan di-override manusia
- [x] Alert terdeduplikasi
- [x] Security checks lulus
- [x] Backup/restore checks lulus

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
| `bc7e121` | Auditable human pursue/pass decision dengan mandatory reason |
| `03bca39` | Opportunity next action dengan UTC deadline dan overdue state |
| `f2b9fbd` | Versioned readiness checklist dengan progress, notes, RBAC, dan audit |
| `0a961df` | Human recommendation/score override terpisah dengan mandatory reason |
| `c139fec` | Unified opportunity audit activity timeline untuk seluruh workflow |
| `2281fd7` | Initial pipeline analytics untuk portfolio dan next-action risk |
| `45a8568` | Batch lima source SPSE dan parser nilai/keyword yang lebih tangguh |
| `ebf1362` | Batch kedua lima source SPSE hingga total sepuluh source aktif |
| `1f59e59` | Batch ketiga lima source SPSE hingga total lima belas source aktif |
| `77d53ac` | Batch keempat mencapai target dua puluh source SPSE aktif |
| `9dd7dbd` | Personal alert rules dan immutable new/material-change event inbox |
| `b2448c6` | Deterministic per-user alert deduplication untuk overlapping rules dan retry |
| `2f36033` | Auditable per-user quiet hours dengan timezone-aware delivery deferral |
| `6dc92a8` | Reliable STARTTLS email notifier dengan transactional outbox dan retry |
| `74fe9e1` | Signed HTTPS webhook notifier dengan idempotency key dan bounded response handling |
| `4745e0b` | RBAC dead-letter console dan auditable replay dengan preserved delivery identity |
| `f077593` | Failure recovery runbook dengan triage, canary, replay, verification, dan escalation gate |
| `8d395f2` | CI release gates dan fresh PostgreSQL migration-chain verification sampai v29 |
| `5159e0e` | PostgreSQL-backed API authorization dan response-contract integration suite |
| `ef22ccd` | Idempotent discover-to-review collector workflow E2E pada PostgreSQL kosong |
| `90b7e32` | Dependency remediation dan automated repository/container vulnerability gates |
| `f8198f0` | Authenticated 10k-row opportunity feed pagination and concurrent load gate |
| `797325e` | Atomic PostgreSQL backup, checksum/count manifest, dan empty-target restore drill |
| `40d0882` | File-backed production secrets, startup policy, Compose boundary, dan rotation runbook |
| `621cd44` | Hardened TLS 1.2/1.3 staging edge, trusted-proxy policy, dan deployment overlay |
| `20c96d8` | Executable HTTPS smoke, UAT matrix, approval gate, deployment, dan rollback runbook |
| `6560b56` | Database-backed seven-day source stability report dan non-bypassable release gate |
| `a93c7f5` | Bounded transient collector recovery dan immediate permanent access-policy failure |
| `0c85b68` | Full inactive-lifecycle integration gate untuk active opportunity feed |
| `1323806` | Auditable domain-owner classification approval metadata dan release gate |
| `c5587f9` | Unified fail-closed GO/NO-GO release preflight untuk seluruh local dan hosted evidence |
| `75b06aa` | Bounded-cardinality API metrics, Prometheus staging, dan validated alert rules |
| `0851627` | Database-backed source/queue/circuit/delivery metrics dan operational alerts |
| `500d70b` | Collector internal health/metrics server, Docker healthcheck, dan down/error alerts |
| `73bca10` | UTC daily stability checkpoint dengan explicit pending/failed/passed semantics |

## Aturan pembaruan task list

1. Pindahkan status hanya setelah acceptance criteria dan test relevan lulus.
2. Tambahkan commit/bukti pada bagian yang selesai.
3. Catat blocker langsung pada item menggunakan `[!]`.
4. Jangan menyimpan password, token, atau secret di dokumen ini.
5. Jika scope berubah, perbarui task list dan roadmap pada commit yang sama.
