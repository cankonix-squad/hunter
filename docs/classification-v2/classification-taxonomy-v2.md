# CANKONIX Hunter — business taxonomy v2 (candidate)

This is a review candidate. The live collector remains on `cankonix-it-v1` / `deterministic-v1`.

## Decision policy

1. Normalize case and punctuation, preserving word boundaries.
2. Find exact multiword phrases and synonym/typo variants in title, description, KAK/TOR, and requirements.
3. Keep discovery terms separate from decision rules. Require either a strong work phrase or a positive phrase plus a distinct work-context phrase in the same field. A bare domain keyword is insufficient.
4. Apply category negative phrases and explicit exclusion rules. Mixed digital and non-digital scope enters `needs_review`.
5. Use procurement type as supporting evidence only. `Barang` never excludes a case by itself.
6. If several categories qualify, the earliest strong phrase in the title selects the primary; then evidence strength and category ID break ties. Other qualified categories are secondary.
7. `relevant` means clear business fit; `excluded` means clearly outside scope; `needs_review` means ambiguity, insufficient detail, or conflicting scope. Review status is independent of business score.
8. Preserve matched field, phrase, category, and reason code as evidence. No v2 exclusion auto-rejects; a human reviews exclusions before any production policy change.

The editable JSON uses stable category IDs and arrays of discovery, positive, strong, context, and negative phrases. The owner's master terms are recorded in `discovery_keywords`; discovery alone never makes a package relevant. Generic words such as dashboard, monitoring, pemeliharaan, aplikasi, and konsultansi always need a verified IT work context. An Admin Dashboard can later edit a draft, validate schema, run regression, and publish an immutable version with actor/time/audit trail. Published versions must never be changed in place.

## Categories

### Rekayasa Perangkat Lunak (`software-engineering`, P1)

Rancang bangun, pengembangan, modernisasi, dan pemeliharaan aplikasi khusus.

- Discovery keywords: pengembangan aplikasi, pembangunan aplikasi, pembuatan aplikasi, sistem informasi, pengembangan website, pemeliharaan aplikasi, penyempurnaan aplikasi, upgrade aplikasi, rekayasa perangkat lunak, aplikasi berbasis web, mobile application.
- Positive keywords: aplikasi, perangkat lunak, software.
- Strong work phrases: pengembangan aplikasi, pembangunan sistem informasi, modernisasi aplikasi, pemeliharaan aplikasi, pembangunan aplikasi, pembuatan aplikasi, pengembangan website, penyempurnaan aplikasi, upgrade aplikasi, rekayasa perangkat lunak.
- Synonyms, abbreviations, typos: aplikasi: app, apliksi; sistem informasi: sisfo, sistem infrormasi.
- Work context: pengembangan, pembangunan, implementasi, kustomisasi, pemeliharaan.
- Negative keywords: aplikasi perkantoran berlisensi tanpa jasa.
- Exclusion rule: `jual putus aplikasi umum tanpa layanan` means scope is outside this category.
- Relevant example: Pengembangan aplikasi web pelayanan publik.
- Not relevant example: Pengadaan meja kantor untuk tim pengembang aplikasi.
- Ambiguous example: Pengadaan aplikasi untuk kantor (review the actual scope).

### Keamanan Siber (`cybersecurity`, P1)

Asesmen, pengujian, respons, dan peningkatan keamanan sistem digital.

- Discovery keywords: keamanan informasi, keamanan siber, penetration test, pentest, vulnerability assessment, security assessment, SOC, SIEM, vulnerability scanning, incident response, digital forensic, security monitoring.
- Positive keywords: keamanan siber, keamanan informasi, penetration test, vulnerability assessment, soc, siem.
- Strong work phrases: penetration test aplikasi, audit keamanan siber, implementasi keamanan siber, pengujian keamanan aplikasi, pelatihan keamanan informasi, security assessment, vulnerability scanning, incident response, digital forensic, security monitoring.
- Synonyms, abbreviations, typos: penetration test: pentest, pen test; keamanan siber: cyber security, siber sekuriti.
- Work context: audit, asesmen, pengujian, implementasi, monitoring, aplikasi, pelatihan, operasional, pengelolaan.
- Negative keywords: keamanan fisik gedung.
- Exclusion rule: `satpam dan keamanan fisik` means scope is outside this category.
- Relevant example: Penetration test aplikasi layanan publik.
- Not relevant example: Pengadaan seragam petugas keamanan siber.
- Ambiguous example: Pengadaan perangkat keamanan (review the actual scope).

### Tata Kelola, Risiko, dan Kepatuhan TI (`it-grc`, P1)

Kebijakan, audit, manajemen risiko, SPBE, dan kepatuhan tata kelola TI.

- Discovery keywords: ISO 27001, SMKI, audit keamanan informasi, IT governance, IT risk assessment, manajemen risiko TI, compliance, tata kelola SPBE, evaluasi SPBE, indeks KAMI, IT audit, BCP, DRP.
- Positive keywords: tata kelola ti, spbe, risiko teknologi informasi, iso 27001, smki, it governance, compliance, indeks kami, bcp, drp.
- Strong work phrases: audit tata kelola ti, penyusunan tata kelola ti, asesmen spbe, manajemen risiko ti, pelatihan tata kelola ti, audit keamanan informasi, it risk assessment, evaluasi spbe, tata kelola spbe, it audit.
- Synonyms, abbreviations, typos: spbe: s p b e, sistem pemerintahan berbasis elektronik; tata kelola ti: it governance, tatakelola ti.
- Work context: audit, asesmen, kebijakan, risiko, kepatuhan, penyusunan, pelatihan, implementasi, evaluasi.
- Negative keywords: tata kelola organisasi non digital.
- Exclusion rule: `audit keuangan tanpa lingkup TI` means scope is outside this category.
- Relevant example: Audit tata kelola TI pemerintah daerah.
- Not relevant example: Pengadaan katering rapat tata kelola TI.
- Ambiguous example: Kajian tata kelola organisasi (review the actual scope).

### Data dan Kecerdasan Buatan (`data-ai`, P1)

Integrasi data, platform analitik, pembelajaran mesin, dan solusi AI terukur.

- Discovery keywords: integrasi data, analitik data, data warehouse, data lake, business intelligence, dashboard analitik, dashboard eksekutif, kecerdasan buatan, machine learning, artificial intelligence, big data, data governance.
- Positive keywords: data warehouse, analitik data, integrasi data, machine learning, kecerdasan buatan, data lake, business intelligence, dashboard analitik, dashboard eksekutif, big data, data governance.
- Strong work phrases: pembangunan data warehouse, pengembangan analitik data, pembangunan analitik data, layanan integrasi data, implementasi machine learning, pengembangan kecerdasan buatan, pembangunan data lake, pengembangan business intelligence, pengembangan dashboard analitik, pengembangan dashboard eksekutif.
- Synonyms, abbreviations, typos: machine learning: ml, machne learning; kecerdasan buatan: ai, artificial intelligence; data warehouse: dwh, dataware house.
- Work context: pembangunan, pengembangan, implementasi, model, analitik.
- Negative keywords: entri data manual tanpa sistem.
- Exclusion rule: `jasa input data manual tanpa pengembangan` means scope is outside this category.
- Relevant example: Pembangunan data warehouse kesehatan.
- Not relevant example: Pengadaan kertas printer untuk pusat data AI.
- Ambiguous example: Pengadaan data untuk kajian (review the actual scope).

### Infrastruktur Cloud (`cloud-infrastructure`, P1)

Arsitektur, migrasi, operasi, dan keamanan platform cloud atau container.

- Discovery keywords: cloud computing, cloud server, cloud hosting, virtual private server, VPS, data center, disaster recovery center, virtualisasi server, containerization, Kubernetes, DevOps, infrastruktur TI.
- Positive keywords: cloud, komputasi awan, container platform, cloud computing, cloud server, cloud hosting, virtual private server, vps, data center, disaster recovery center, virtualisasi server, containerization, kubernetes, infrastruktur ti.
- Strong work phrases: migrasi cloud, implementasi cloud, pembangunan platform cloud, implementasi container platform, implementasi cloud server, migrasi data center, virtualisasi server, implementasi kubernetes.
- Synonyms, abbreviations, typos: cloud: komputasi awan, clod; container platform: kubernetes, k8s.
- Work context: migrasi, implementasi, arsitektur, operasi, platform, konfigurasi, pengelolaan, pemeliharaan.
- Negative keywords: penyimpanan fisik tanpa integrasi.
- Exclusion rule: `sewa gudang penyimpanan barang fisik` means scope is outside this category.
- Relevant example: Migrasi cloud aplikasi pemerintah.
- Not relevant example: Pengadaan kursi kantor untuk unit cloud.
- Ambiguous example: Pengadaan kapasitas cloud (review the actual scope).

### Integrasi Sistem (`system-integration`, P1)

Integrasi aplikasi, API, interoperabilitas, dan pertukaran data antar sistem.

- Discovery keywords: integrasi sistem, integrasi aplikasi, API gateway, interoperabilitas, pertukaran data, bridging data, web service, middleware, single sign on, SSO, sinkronisasi data.
- Positive keywords: integrasi sistem, integrasi api, interoperabilitas, api gateway, pertukaran data, web service, middleware, single sign on, sso.
- Strong work phrases: integrasi sistem informasi, pengembangan integrasi api, implementasi interoperabilitas, integrasi data antar aplikasi, integrasi aplikasi, implementasi api gateway, bridging data, sinkronisasi data.
- Synonyms, abbreviations, typos: integrasi api: api integration, integrsi api; interoperabilitas: interoperability, interoperabiltas.
- Work context: pengembangan, implementasi, antar aplikasi, pertukaran data, sistem informasi, integrasi.
- Negative keywords: integrasi kegiatan non digital.
- Exclusion rule: `koordinasi antarlembaga tanpa sistem` means scope is outside this category.
- Relevant example: Integrasi sistem informasi kesehatan.
- Not relevant example: Pengadaan alat tulis kantor tim integrasi sistem.
- Ambiguous example: Kajian integrasi organisasi (review the actual scope).

### PMO Transformasi Digital (`pmo-digital-transformation`, P1)

Manajemen program dan pendampingan transformasi digital berbasis sistem informasi.

- Discovery keywords: project management office, PMO, dashboard monitoring, monitoring evaluasi, monev digital, sistem pemantauan program, pengendalian proyek, transformasi digital, manajemen portofolio, digital transformation.
- Positive keywords: pmo digital, transformasi digital, spbe, dashboard monitoring, monitoring evaluasi, monev digital, sistem pemantauan program, pengendalian proyek, manajemen portofolio.
- Strong work phrases: pmo transformasi digital, pendampingan transformasi digital, manajemen program spbe, pmo sistem informasi, monev digital berbasis sistem, pengembangan sistem pemantauan program.
- Synonyms, abbreviations, typos: pmo: project management office, program management office; transformasi digital: digital transformation, tranformasi digital.
- Work context: pmo, pendampingan, program, sistem informasi, spbe, pengembangan, berbasis sistem.
- Negative keywords: transformasi organisasi tanpa teknologi.
- Exclusion rule: `konsultansi organisasi tanpa sistem informasi` means scope is outside this category.
- Relevant example: PMO transformasi digital provinsi.
- Not relevant example: Pengadaan seragam tim PMO digital.
- Ambiguous example: Transformasi organisasi perangkat daerah (review the actual scope).

### Layanan TI Terkelola (`managed-it-services`, P2)

Operasi, pemantauan, dukungan, dan SLA untuk layanan TI.

- Discovery keywords: managed service, managed IT, pemeliharaan sistem, pemeliharaan server, dukungan teknis TI, IT support, operasional aplikasi, pemeliharaan database, monitoring server, technical support.
- Positive keywords: managed service ti, operasional ti, helpdesk aplikasi, managed it, pemeliharaan sistem, it support, technical support.
- Strong work phrases: managed service ti, operasional layanan ti, pengelolaan helpdesk aplikasi, pemantauan infrastruktur ti, pemeliharaan sistem informasi, pemeliharaan server, dukungan teknis ti, operasional aplikasi, pemeliharaan database, monitoring server.
- Synonyms, abbreviations, typos: managed service: managed services, layanan terkelola; helpdesk: service desk, help desk.
- Work context: operasional, pengelolaan, pemantauan, sla, dukungan, pemeliharaan.
- Negative keywords: operasional gedung, katering.
- Exclusion rule: `pengelolaan fasilitas non TI` means scope is outside this category.
- Relevant example: Managed service TI kantor pusat.
- Not relevant example: Pengadaan katering operator managed service TI.
- Ambiguous example: Layanan bantuan kantor (review the actual scope).

### Sistem Enterprise (`enterprise-systems`, P1)

Implementasi dan integrasi ERP, HRIS, CRM, serta sistem enterprise sejenis.

- Discovery keywords: enterprise resource planning, ERP, HRIS, sistem keuangan, sistem akuntansi, financial management system, CRM, procurement management system, e-office, e-document, manajemen aset, e-arsip.
- Positive keywords: erp, hris, crm, sistem enterprise, sistem keuangan, sistem akuntansi, financial management system, procurement management system, e-office, e-document, manajemen aset, e-arsip.
- Strong work phrases: implementasi erp, pengembangan hris, integrasi crm, implementasi sistem enterprise, pelatihan aplikasi erp, implementasi sistem keuangan, pengembangan sistem akuntansi, pengembangan e-office, implementasi e-arsip.
- Synonyms, abbreviations, typos: erp: enterprise resource planning, e r p; hris: human resource information system, h r i s.
- Work context: implementasi, pengembangan, integrasi, migrasi, sistem.
- Negative keywords: pelatihan umum tanpa aplikasi khusus.
- Exclusion rule: `lisensi umum tanpa implementasi` means scope is outside this category.
- Relevant example: Implementasi ERP badan usaha daerah.
- Not relevant example: Pengadaan meja kantor tim ERP.
- Ambiguous example: Pengadaan lisensi ERP (review the actual scope).

### Infrastruktur Jaringan (`network-infrastructure`, P2)

Desain, implementasi, konfigurasi, integrasi, dan keamanan jaringan TI.

- Discovery keywords: konfigurasi firewall, implementasi firewall, network security, implementasi jaringan, instalasi LAN, konfigurasi VPN, network monitoring, network assessment, pengembangan infrastruktur jaringan.
- Positive keywords: jaringan, network, sd wan, network security.
- Strong work phrases: implementasi jaringan, konfigurasi jaringan, integrasi jaringan, keamanan jaringan, konfigurasi firewall, implementasi firewall, instalasi lan, konfigurasi vpn, network monitoring, network assessment, pengembangan infrastruktur jaringan.
- Synonyms, abbreviations, typos: sd wan: sdwan, software defined wan; jaringan: networking, jarngan.
- Work context: implementasi, konfigurasi, integrasi, keamanan, arsitektur, instalasi.
- Negative keywords: jaringan listrik.
- Exclusion rule: `pengadaan kabel listrik tanpa jaringan TI` means scope is outside this category.
- Relevant example: Implementasi jaringan kantor cabang.
- Not relevant example: Pengadaan kabel listrik jaringan gedung.
- Ambiguous example: Belanja perangkat jaringan kantor (review the actual scope).

### Konsultansi TI (`it-consulting`, P2)

Perencanaan, arsitektur, asesmen, dan strategi teknologi informasi.

- Discovery keywords: IT master plan, rencana induk TI, blueprint TI, enterprise architecture, audit sistem informasi, kajian teknologi informasi, IT assessment, feasibility study TI, konsultansi transformasi digital.
- Positive keywords: konsultansi ti, arsitektur ti, strategi teknologi informasi.
- Strong work phrases: konsultansi arsitektur ti, penyusunan strategi ti, asesmen teknologi informasi, perencanaan sistem informasi, it master plan, rencana induk ti, blueprint ti, audit sistem informasi, kajian teknologi informasi, it assessment, feasibility study ti, konsultansi transformasi digital.
- Synonyms, abbreviations, typos: teknologi informasi: ti, it; arsitektur ti: enterprise architecture, arsitektur teknologi informasi.
- Work context: konsultansi, perencanaan, asesmen, strategi, arsitektur.
- Negative keywords: konsultansi manajemen umum.
- Exclusion rule: `nasihat organisasi tanpa teknologi` means scope is outside this category.
- Relevant example: Konsultansi arsitektur TI instansi.
- Not relevant example: Pengadaan alat tulis kantor konsultan TI.
- Ambiguous example: Konsultansi transformasi organisasi (review the actual scope).

### Solusi Digital Khusus (`specialized-digital-solutions`, P2)

Solusi digital spesifik sektor seperti GIS, IoT, identitas digital, dan otomasi.

- Discovery keywords: e-government, e-learning, LMS, sistem monitoring, logbook elektronik, sistem perizinan, sistem pelayanan publik, sistem informasi kesehatan, IoT monitoring, dashboard operasional.
- Positive keywords: gis, iot, identitas digital, otomasi digital, e-government, e-learning, lms, sistem monitoring, logbook elektronik, sistem perizinan, sistem pelayanan publik, sistem informasi kesehatan, iot monitoring, dashboard operasional.
- Strong work phrases: pengembangan gis, implementasi iot, pengembangan identitas digital, implementasi otomasi digital, pengembangan e-government, implementasi e-learning, pengembangan logbook elektronik, implementasi sistem perizinan, pengembangan sistem informasi kesehatan.
- Synonyms, abbreviations, typos: gis: sistem informasi geografis, geographic information system; iot: internet of things, i o t.
- Work context: pengembangan, implementasi, integrasi, platform.
- Negative keywords: survei geografis manual tanpa sistem.
- Exclusion rule: `pengadaan sensor tanpa integrasi digital` means scope is outside this category.
- Relevant example: Pengembangan GIS tata ruang.
- Not relevant example: Pengadaan bibit tanaman untuk proyek GIS.
- Ambiguous example: Pengadaan sensor IoT (review the actual scope).

## Classification and opportunity score

Classification is decided before opportunity scoring. The candidate score uses capability fit 35%, scope fit 25%, supplier qualification 20%, value/margin 10%, and delivery readiness 10%. Each component is either a 0–100 value with explicit evidence or `unknown`. The score is normalized over known weights; `data_completeness` is reported separately and is not a calibrated confidence probability. Any incomplete score has `provisional=true`. A case with no known components has no numeric score. Bands are high 80–100, medium 60–79, low 0–59. Mandatory failure flags require evidence; unproven flags remain unknown. Missing data never becomes zero.

## Boundaries

Only public-source data needed for opportunity monitoring may be collected. KAK/TOR text must be safely extracted and treated as untrusted input. A future provider-agnostic AI adapter may propose structured evidence for `needs_review` cases after deterministic rules, but cannot silently reject a package; a human retains the pursue/pass decision. v2 is not wired to the collector, database, or Admin Dashboard; those require a reviewed migration and release gate after independent validation.

## Reproduce the candidate

From the repository root run `python3 scripts/generate-classification-v2.py`. Then from `services/api` run `go test ./internal/classificationv2 ./cmd/evaluatev2` and `go run ./cmd/evaluatev2 > ../../docs/classification-v2/classification-evaluation-v2.md`. Review generated changes before accepting rule or label edits.
