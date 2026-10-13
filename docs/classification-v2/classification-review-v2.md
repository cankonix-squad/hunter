# CANKONIX Hunter — classification review v2

Status: **candidate for domain review**. The owner approved the original 15 v1 labels as a temporary baseline, not the expanded v2 scope or production engine.

The 120 cases below are synthetic regression fixtures: 50 relevant, 40 excluded, 30 needs_review. Fifteen rows tagged `v1_baseline` preserve the owner's exact v1 titles, procurement types, and labels. The remaining rows require independent review, especially mixed scope, Barang procurement, abbreviations, and negative contexts.

For each row, mark APPROVE or CHANGE; record corrected label, primary/secondary categories, and reason. Reviewers should also add real sampled LPSE packages with canonical URLs and separately assigned labels. Do not put credentials or private data in this packet.

| ID | Title | Procurement | Expected label | Primary | Secondary | Category under test | Tags | Decision / correction |
|---|---|---|---|---|---|---|---|---|
| rel-software-engineering-1 | Pengembangan Aplikasi Web Pelayanan Publik | Jasa Lainnya | relevant | software-engineering | — | — | category_positive, v1_baseline | — |
| rel-software-engineering-2 | Pembangunan Sistem Informasi Kepegawaian | Jasa Konsultansi | relevant | software-engineering | — | — | category_positive, v1_baseline | — |
| rel-software-engineering-3 | Pengembangan apliksi perizinan daerah | Barang | relevant | software-engineering | — | — | category_positive, mixed_procurement, abbreviation_typo | — |
| rel-software-engineering-4 | Pemeliharaan aplikasi khusus retribusi | Jasa Lainnya | relevant | software-engineering | — | — | category_positive | — |
| review-software-engineering | Pengadaan aplikasi untuk kantor | Jasa Konsultansi | needs_review | — | — | software-engineering | category_ambiguous | — |
| rel-cybersecurity-1 | Jasa Penetration Test Aplikasi | Jasa Konsultansi | relevant | cybersecurity | — | — | category_positive, v1_baseline | — |
| rel-cybersecurity-2 | Audit cyber security pusat data | Jasa Konsultansi | relevant | cybersecurity | — | — | category_positive, abbreviation_typo | — |
| rel-cybersecurity-3 | Implementasi keamanan siber instansi | Barang | relevant | cybersecurity | — | — | category_positive, mixed_procurement | — |
| rel-cybersecurity-4 | Pelatihan keamanan informasi bagi operator aplikasi | Jasa Lainnya | relevant | cybersecurity | — | — | category_positive, specialist_training | — |
| review-cybersecurity | Pengadaan perangkat keamanan | Jasa Konsultansi | needs_review | — | — | cybersecurity | category_ambiguous | — |
| rel-it-grc-1 | Audit tata kelola TI pemerintah daerah | Jasa Lainnya | relevant | it-grc | — | — | category_positive | — |
| rel-it-grc-2 | Penyusunan tatakelola TI kementerian | Jasa Konsultansi | relevant | it-grc | — | — | category_positive, abbreviation_typo | — |
| rel-it-grc-3 | Asesmen SPBE instansi | Barang | relevant | it-grc | — | — | category_positive, mixed_procurement | — |
| rel-it-grc-4 | Pelatihan tata kelola TI instansi | Jasa Lainnya | relevant | it-grc | — | — | category_positive, specialist_training | — |
| review-it-grc | Kajian tata kelola organisasi | Jasa Konsultansi | needs_review | — | — | it-grc | category_ambiguous | — |
| rel-data-ai-1 | Pengadaan Layanan Integrasi Data | Jasa Lainnya | relevant | data-ai | — | — | category_positive, v1_baseline | — |
| rel-data-ai-2 | Pembangunan Analitik Data Sektoral | Jasa Konsultansi | relevant | data-ai | — | — | category_positive, v1_baseline | — |
| rel-data-ai-3 | Implementasi machne learning untuk prediksi | Barang | relevant | data-ai | — | — | category_positive, mixed_procurement, abbreviation_typo | — |
| rel-data-ai-4 | Pengembangan kecerdasan buatan layanan publik | Jasa Lainnya | relevant | data-ai | — | — | category_positive | — |
| review-data-ai | Pengadaan data untuk kajian | Jasa Konsultansi | needs_review | — | — | data-ai | category_ambiguous | — |
| rel-cloud-infrastructure-1 | Layanan Cloud untuk Platform Pemerintah | Jasa Lainnya | relevant | cloud-infrastructure | — | — | category_positive, v1_baseline | — |
| rel-cloud-infrastructure-2 | Implementasi clod privat instansi | Jasa Konsultansi | relevant | cloud-infrastructure | — | — | category_positive, abbreviation_typo | — |
| rel-cloud-infrastructure-3 | Pembangunan platform cloud daerah | Barang | relevant | cloud-infrastructure | — | — | category_positive, mixed_procurement | — |
| rel-cloud-infrastructure-4 | Implementasi container platform layanan digital | Jasa Lainnya | relevant | cloud-infrastructure | — | — | category_positive | — |
| review-cloud-infrastructure | Pengadaan kapasitas cloud | Jasa Konsultansi | needs_review | — | — | cloud-infrastructure | category_ambiguous | — |
| rel-system-integration-1 | Integrasi sistem informasi kesehatan | Jasa Lainnya | relevant | system-integration | specialized-digital-solutions | — | category_positive, multi_category | — |
| rel-system-integration-2 | Pengembangan integrsi API layanan kependudukan | Jasa Konsultansi | relevant | system-integration | — | — | category_positive, abbreviation_typo | — |
| rel-system-integration-3 | Implementasi interoperabilitas antar instansi | Barang | relevant | system-integration | — | — | category_positive, mixed_procurement | — |
| rel-system-integration-4 | Integrasi data antar aplikasi daerah | Jasa Lainnya | relevant | system-integration | — | — | category_positive | — |
| review-system-integration | Kajian integrasi organisasi | Jasa Konsultansi | needs_review | — | — | system-integration | category_ambiguous | — |
| rel-pmo-digital-transformation-1 | PMO transformasi digital provinsi | Jasa Lainnya | relevant | pmo-digital-transformation | — | — | category_positive | — |
| rel-pmo-digital-transformation-2 | Pendampingan tranformasi digital rumah sakit | Jasa Konsultansi | relevant | pmo-digital-transformation | — | — | category_positive, abbreviation_typo | — |
| rel-pmo-digital-transformation-3 | Manajemen program SPBE kementerian | Barang | relevant | pmo-digital-transformation | — | — | category_positive, mixed_procurement | — |
| rel-pmo-digital-transformation-4 | PMO sistem informasi lintas dinas | Jasa Lainnya | relevant | pmo-digital-transformation | — | — | category_positive | — |
| review-pmo-digital-transformation | Transformasi organisasi perangkat daerah | Jasa Konsultansi | needs_review | — | — | pmo-digital-transformation | category_ambiguous | — |
| rel-managed-it-services-1 | Managed service TI kantor pusat | Jasa Lainnya | relevant | managed-it-services | — | — | category_positive | — |
| rel-managed-it-services-2 | Operasional layanan TI 24 jam | Jasa Konsultansi | relevant | managed-it-services | — | — | category_positive | — |
| rel-managed-it-services-3 | Pengelolaan helpdesk aplikasi publik | Barang | relevant | managed-it-services | — | — | category_positive, mixed_procurement | — |
| rel-managed-it-services-4 | Pemantauan infrastruktur TI pemerintah | Jasa Lainnya | relevant | managed-it-services | — | — | category_positive | — |
| review-managed-it-services | Layanan bantuan kantor | Jasa Konsultansi | needs_review | — | — | managed-it-services | category_ambiguous | — |
| rel-enterprise-systems-1 | Implementasi ERP badan usaha daerah | Jasa Lainnya | relevant | enterprise-systems | — | — | category_positive | — |
| rel-enterprise-systems-2 | Pengembangan HRIS instansi | Jasa Konsultansi | relevant | enterprise-systems | — | — | category_positive | — |
| rel-enterprise-systems-3 | Integrasi CRM layanan warga | Barang | relevant | enterprise-systems | — | — | category_positive, mixed_procurement | — |
| rel-enterprise-systems-4 | Pelatihan aplikasi ERP khusus | Jasa Lainnya | relevant | enterprise-systems | — | — | category_positive, specialist_training | — |
| review-enterprise-systems | Pengadaan lisensi ERP | Barang | needs_review | — | — | enterprise-systems | category_ambiguous | — |
| rel-network-infrastructure-1 | Implementasi jaringan kantor cabang | Jasa Lainnya | relevant | network-infrastructure | — | — | category_positive | — |
| rel-network-infrastructure-2 | Konfigurasi jarngan pusat data | Jasa Konsultansi | relevant | network-infrastructure | — | — | category_positive, abbreviation_typo | — |
| rel-network-infrastructure-3 | Pengadaan perangkat jaringan antar gedung | Barang | relevant | network-infrastructure | — | — | category_positive, mixed_procurement, tor_context | — |
| rel-network-infrastructure-4 | Keamanan jaringan instansi | Jasa Lainnya | relevant | network-infrastructure | — | — | category_positive | — |
| review-network-infrastructure | Belanja perangkat jaringan kantor | Barang | needs_review | — | — | network-infrastructure | category_ambiguous | — |
| rel-it-consulting-1 | Konsultansi arsitektur TI instansi | Jasa Lainnya | relevant | it-consulting | — | — | category_positive | — |
| rel-it-consulting-2 | Penyusunan strategi TI daerah | Jasa Konsultansi | relevant | it-consulting | — | — | category_positive | — |
| rel-it-consulting-3 | Asesmen teknologi informasi kementerian | Barang | relevant | it-consulting | — | — | category_positive, mixed_procurement | — |
| rel-it-consulting-4 | Perencanaan sistem informasi pelayanan | Jasa Lainnya | relevant | it-consulting | — | — | category_positive | — |
| review-it-consulting | Konsultansi transformasi organisasi | Jasa Konsultansi | needs_review | — | — | it-consulting | category_ambiguous | — |
| rel-specialized-digital-solutions-1 | Pengembangan GIS tata ruang | Jasa Lainnya | relevant | specialized-digital-solutions | — | — | category_positive | — |
| rel-specialized-digital-solutions-2 | Implementasi internet of things pemantauan air | Jasa Konsultansi | relevant | specialized-digital-solutions | — | — | category_positive, abbreviation_typo | — |
| rel-specialized-digital-solutions-3 | Pengembangan identitas digital warga | Barang | relevant | specialized-digital-solutions | — | — | category_positive, mixed_procurement | — |
| rel-specialized-digital-solutions-4 | Implementasi otomasi digital perizinan | Jasa Lainnya | relevant | specialized-digital-solutions | — | — | category_positive | — |
| review-specialized-digital-solutions | Pengadaan sensor IoT | Jasa Konsultansi | needs_review | — | — | specialized-digital-solutions | category_ambiguous | — |
| rel-multi-security-grc | Audit tata kelola TI dan keamanan siber | Jasa Konsultansi | relevant | it-grc | cybersecurity | — | multi_category | — |
| rel-multi-software-integration | Pengembangan aplikasi dan integrasi sistem informasi | Barang | relevant | software-engineering | system-integration | — | multi_category, mixed_procurement | — |
| excluded-01 | Pengadaan Alat Tulis Kantor Tahun 2026 | Barang | excluded | — | — | — | category_negative, negative_context, v1_baseline | — |
| excluded-02 | Belanja Seragam Pegawai Lapangan | Barang | excluded | — | — | — | category_negative, negative_context, v1_baseline | — |
| excluded-03 | Pengadaan Katering Rapat Koordinasi | Jasa Lainnya | excluded | — | — | — | category_negative, negative_context, v1_baseline | — |
| excluded-04 | Pengadaan Toner Printer | Barang | excluded | — | — | — | category_negative, negative_context, v1_baseline | — |
| excluded-05 | Belanja Kertas Printer Tahunan | Barang | excluded | — | — | — | category_negative, negative_context, v1_baseline | — |
| excluded-06 | Pengadaan meja kantor untuk tim pengembang aplikasi | Jasa Lainnya | excluded | — | — | software-engineering | category_negative, negative_context, category_negative | — |
| excluded-07 | Pengadaan seragam petugas keamanan siber | Barang | excluded | — | — | cybersecurity | category_negative, negative_context, category_negative | — |
| excluded-08 | Pengadaan katering rapat tata kelola TI | Jasa Lainnya | excluded | — | — | it-grc | category_negative, negative_context, category_negative | — |
| excluded-09 | Pengadaan kertas printer untuk pusat data AI | Barang | excluded | — | — | data-ai | category_negative, negative_context, category_negative | — |
| excluded-10 | Pengadaan kursi kantor untuk unit cloud | Jasa Lainnya | excluded | — | — | cloud-infrastructure | category_negative, negative_context, category_negative | — |
| excluded-11 | Pengadaan alat tulis kantor tim integrasi sistem | Barang | excluded | — | — | system-integration | category_negative, negative_context, category_negative | — |
| excluded-12 | Pengadaan seragam tim PMO digital | Jasa Lainnya | excluded | — | — | pmo-digital-transformation | category_negative, negative_context, category_negative | — |
| excluded-13 | Pengadaan katering operator managed service TI | Barang | excluded | — | — | managed-it-services | category_negative, negative_context, category_negative | — |
| excluded-14 | Pengadaan meja kantor tim ERP | Jasa Lainnya | excluded | — | — | enterprise-systems | category_negative, negative_context, category_negative | — |
| excluded-15 | Pengadaan kabel listrik jaringan gedung | Barang | excluded | — | — | network-infrastructure | category_negative, negative_context, category_negative | — |
| excluded-16 | Pengadaan alat tulis kantor konsultan TI | Jasa Lainnya | excluded | — | — | it-consulting | category_negative, negative_context, category_negative | — |
| excluded-17 | Pengadaan bibit tanaman untuk proyek GIS | Barang | excluded | — | — | specialized-digital-solutions | category_negative, negative_context, category_negative | — |
| excluded-18 | Pengadaan bibit tanaman untuk kantor daerah | Jasa Lainnya | excluded | — | — | — | category_negative, negative_context | — |
| excluded-19 | Pengadaan pupuk pertanian untuk kantor daerah | Barang | excluded | — | — | — | category_negative, negative_context | — |
| excluded-20 | Pengadaan keamanan fisik gedung untuk kantor daerah | Jasa Lainnya | excluded | — | — | — | category_negative, negative_context | — |
| excluded-21 | Pengadaan perjalanan dinas untuk kantor daerah | Barang | excluded | — | — | — | category_negative, negative_context | — |
| excluded-22 | Pengadaan jasa parkir untuk kantor daerah | Jasa Lainnya | excluded | — | — | — | category_negative, negative_context | — |
| excluded-23 | Pengadaan fotokopi dokumen untuk kantor daerah | Barang | excluded | — | — | — | category_negative, negative_context | — |
| excluded-24 | Pengadaan baju dinas untuk kantor daerah | Jasa Lainnya | excluded | — | — | — | category_negative, negative_context | — |
| excluded-25 | Pengadaan jasa laundry untuk kantor daerah | Barang | excluded | — | — | — | category_negative, negative_context | — |
| excluded-26 | Pengadaan peralatan dapur untuk kantor daerah | Jasa Lainnya | excluded | — | — | — | category_negative, negative_context | — |
| excluded-27 | Pengadaan sewa tenda untuk kantor daerah | Barang | excluded | — | — | — | category_negative, negative_context | — |
| excluded-28 | Pengadaan pencetakan brosur untuk kantor daerah | Jasa Lainnya | excluded | — | — | — | category_negative, negative_context | — |
| excluded-29 | Pengadaan pengadaan air minum untuk kantor daerah | Barang | excluded | — | — | — | category_negative, negative_context | — |
| excluded-30 | Pengadaan perbaikan jalan untuk kantor daerah | Jasa Lainnya | excluded | — | — | — | category_negative, negative_context | — |
| excluded-31 | Pengadaan pembangunan jembatan untuk kantor daerah | Barang | excluded | — | — | — | category_negative, negative_context | — |
| excluded-32 | Pengadaan cat tembok untuk kantor daerah | Jasa Lainnya | excluded | — | — | — | category_negative, negative_context | — |
| excluded-33 | Pengadaan perabot ruang rapat untuk kantor daerah | Barang | excluded | — | — | — | category_negative, negative_context | — |
| excluded-34 | Pengadaan alat olahraga untuk kantor daerah | Jasa Lainnya | excluded | — | — | — | category_negative, negative_context | — |
| excluded-35 | Pengadaan jasa taman untuk kantor daerah | Barang | excluded | — | — | — | category_negative, negative_context | — |
| excluded-36 | Pengadaan pengadaan beras untuk kantor daerah | Jasa Lainnya | excluded | — | — | — | category_negative, negative_context | — |
| excluded-37 | Pengadaan sewa kendaraan untuk kantor daerah | Barang | excluded | — | — | — | category_negative, negative_context | — |
| excluded-38 | Pengadaan jasa kebersihan toilet untuk kantor daerah | Jasa Lainnya | excluded | — | — | — | category_negative, negative_context | — |
| excluded-39 | Pengadaan pembangunan drainase untuk kantor daerah | Barang | excluded | — | — | — | category_negative, negative_context | — |
| excluded-40 | Pengadaan pengadaan vaksin untuk kantor daerah | Jasa Lainnya | excluded | — | — | — | category_negative, negative_context | — |
| review-organization | Jasa Konsultansi Transformasi Organisasi | Jasa Konsultansi | needs_review | — | — | — | category_ambiguous, v1_baseline | — |
| review-network | Belanja Perangkat Jaringan Kantor | Barang | needs_review | — | — | — | category_ambiguous, v1_baseline | — |
| review-training | Pelatihan Literasi Digital Pegawai | Jasa Lainnya | needs_review | — | — | — | category_ambiguous, v1_baseline | — |
| review-conflict | Pengembangan Aplikasi dan Pengadaan Alat Tulis Kantor | Jasa Lainnya | needs_review | — | — | — | keyword_collision, v1_baseline | — |
| review-mixed-network | Implementasi jaringan dan pengadaan seragam teknisi | Barang | needs_review | — | — | — | mixed_procurement | — |
| review-mixed-cloud | Migrasi cloud dan pengadaan kursi kantor | Jasa Lainnya | needs_review | — | — | — | mixed_procurement | — |
| review-mixed-security | Penetration test aplikasi dan pengadaan katering | Jasa Konsultansi | needs_review | — | — | — | keyword_collision | — |
| review-mixed-data | Pengembangan analitik data dan pengadaan toner printer | Barang | needs_review | — | — | — | keyword_collision | — |
| review-generic-digital | Pengadaan solusi digital | Barang | needs_review | — | — | — | insufficient_context | — |
| review-generic-platform | Penyediaan platform teknologi | Jasa Lainnya | needs_review | — | — | — | insufficient_context | — |
| review-license-only | Pengadaan lisensi aplikasi umum | Barang | needs_review | — | — | — | license_scope | — |
| review-hardware-only | Pengadaan server dan perangkat jaringan | Barang | needs_review | — | — | — | hardware_scope | — |
| review-consulting-general | Jasa konsultansi manajemen organisasi | Jasa Konsultansi | needs_review | — | — | — | insufficient_context | — |
| review-training-general | Bimbingan teknis literasi digital | Jasa Lainnya | needs_review | — | — | — | training_scope | — |
| review-security-device | Pengadaan perangkat keamanan | Barang | needs_review | — | — | — | hardware_scope | — |
| review-data-entry | Jasa entri data dokumen | Jasa Lainnya | needs_review | — | — | — | scope_unclear | — |
| review-cloud-capacity | Pengadaan kapasitas cloud | Barang | needs_review | — | — | — | scope_unclear | — |
| review-iot-device | Pengadaan sensor IoT | Barang | needs_review | — | — | — | hardware_scope | — |

## Review sign-off

Reviewer name / role: —
Review UTC time: —
Controlled approval reference: —
Corrections and coverage gaps: —
Decision on v2 dataset and rule scope: PENDING
