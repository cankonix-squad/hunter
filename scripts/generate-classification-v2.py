#!/usr/bin/env python3
"""Generate the review-only v2 taxonomy and synthetic regression set."""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "docs" / "classification-v2"
OUT.mkdir(parents=True, exist_ok=True)

# ID, name, business description, priority, positive terms, strong phrases,
# aliases, context, negative context, exclusion rule, four positive titles,
# one ambiguous title, one clearly outside-scope title.
SPECS = [
    ("software-engineering", "Rekayasa Perangkat Lunak", "Rancang bangun, pengembangan, modernisasi, dan pemeliharaan aplikasi khusus.", "P1",
     ["aplikasi", "perangkat lunak", "software"], ["pengembangan aplikasi", "pembangunan sistem informasi", "modernisasi aplikasi", "pemeliharaan aplikasi"],
     {"aplikasi": ["app", "apliksi"], "sistem informasi": ["sisfo", "sistem infrormasi"]}, ["pengembangan", "pembangunan", "implementasi", "kustomisasi", "pemeliharaan"],
     ["aplikasi perkantoran berlisensi tanpa jasa"], "jual putus aplikasi umum tanpa layanan",
     ["Pengembangan aplikasi web pelayanan publik", "Pembangunan sistem informasi kepegawaian", "Modernisasi aplikasi perizinan daerah", "Pemeliharaan aplikasi khusus retribusi"],
     "Pengadaan aplikasi untuk kantor", "Pengadaan meja kantor untuk tim pengembang aplikasi"),
    ("cybersecurity", "Keamanan Siber", "Asesmen, pengujian, respons, dan peningkatan keamanan sistem digital.", "P1",
     ["keamanan siber", "keamanan informasi", "penetration test", "vulnerability assessment"], ["penetration test aplikasi", "audit keamanan siber", "implementasi keamanan siber", "pengujian keamanan aplikasi", "pelatihan keamanan informasi"],
     {"penetration test": ["pentest", "pen test"], "keamanan siber": ["cyber security", "siber sekuriti"]}, ["audit", "asesmen", "pengujian", "implementasi", "monitoring", "aplikasi", "pelatihan"],
     ["keamanan fisik gedung"], "satpam dan keamanan fisik",
     ["Penetration test aplikasi layanan publik", "Audit keamanan siber pusat data", "Implementasi keamanan siber instansi", "Pengujian keamanan aplikasi kependudukan"],
     "Pengadaan perangkat keamanan", "Pengadaan seragam petugas keamanan siber"),
    ("it-grc", "Tata Kelola, Risiko, dan Kepatuhan TI", "Kebijakan, audit, manajemen risiko, SPBE, dan kepatuhan tata kelola TI.", "P1",
     ["tata kelola ti", "spbe", "risiko teknologi informasi"], ["audit tata kelola ti", "penyusunan tata kelola ti", "asesmen spbe", "manajemen risiko ti", "pelatihan tata kelola ti"],
     {"spbe": ["s p b e", "sistem pemerintahan berbasis elektronik"], "tata kelola ti": ["it governance", "tatakelola ti"]}, ["audit", "asesmen", "kebijakan", "risiko", "kepatuhan", "penyusunan", "pelatihan"],
     ["tata kelola organisasi non digital"], "audit keuangan tanpa lingkup TI",
     ["Audit tata kelola TI pemerintah daerah", "Penyusunan tata kelola TI kementerian", "Asesmen SPBE instansi", "Manajemen risiko TI layanan publik"],
     "Kajian tata kelola organisasi", "Pengadaan katering rapat tata kelola TI"),
    ("data-ai", "Data dan Kecerdasan Buatan", "Integrasi data, platform analitik, pembelajaran mesin, dan solusi AI terukur.", "P1",
     ["data warehouse", "analitik data", "integrasi data", "machine learning", "kecerdasan buatan"], ["pembangunan data warehouse", "pengembangan analitik data", "pembangunan analitik data", "layanan integrasi data", "implementasi machine learning", "pengembangan kecerdasan buatan"],
     {"machine learning": ["ml", "machne learning"], "kecerdasan buatan": ["ai", "artificial intelligence"], "data warehouse": ["dwh", "dataware house"]}, ["pembangunan", "pengembangan", "implementasi", "model", "analitik"],
     ["entri data manual tanpa sistem"], "jasa input data manual tanpa pengembangan",
     ["Pembangunan data warehouse kesehatan", "Pengembangan analitik data sektoral", "Implementasi machine learning untuk prediksi", "Pengembangan kecerdasan buatan layanan publik"],
     "Pengadaan data untuk kajian", "Pengadaan kertas printer untuk pusat data AI"),
    ("cloud-infrastructure", "Infrastruktur Cloud", "Arsitektur, migrasi, operasi, dan keamanan platform cloud atau container.", "P1",
     ["cloud", "komputasi awan", "container platform"], ["migrasi cloud", "implementasi cloud", "pembangunan platform cloud", "implementasi container platform"],
     {"cloud": ["komputasi awan", "clod"], "container platform": ["kubernetes", "k8s"]}, ["migrasi", "implementasi", "arsitektur", "operasi", "platform"],
     ["penyimpanan fisik tanpa integrasi"], "sewa gudang penyimpanan barang fisik",
     ["Migrasi cloud aplikasi pemerintah", "Implementasi cloud privat instansi", "Pembangunan platform cloud daerah", "Implementasi container platform layanan digital"],
     "Pengadaan kapasitas cloud", "Pengadaan kursi kantor untuk unit cloud"),
    ("system-integration", "Integrasi Sistem", "Integrasi aplikasi, API, interoperabilitas, dan pertukaran data antar sistem.", "P1",
     ["integrasi sistem", "integrasi api", "interoperabilitas"], ["integrasi sistem informasi", "pengembangan integrasi api", "implementasi interoperabilitas", "integrasi data antar aplikasi"],
     {"integrasi api": ["api integration", "integrsi api"], "interoperabilitas": ["interoperability", "interoperabiltas"]}, ["pengembangan", "implementasi", "antar aplikasi", "pertukaran data", "sistem informasi"],
     ["integrasi kegiatan non digital"], "koordinasi antarlembaga tanpa sistem",
     ["Integrasi sistem informasi kesehatan", "Pengembangan integrasi API layanan kependudukan", "Implementasi interoperabilitas antar instansi", "Integrasi data antar aplikasi daerah"],
     "Kajian integrasi organisasi", "Pengadaan alat tulis kantor tim integrasi sistem"),
    ("pmo-digital-transformation", "PMO Transformasi Digital", "Manajemen program dan pendampingan transformasi digital berbasis sistem informasi.", "P1",
     ["pmo digital", "transformasi digital", "spbe"], ["pmo transformasi digital", "pendampingan transformasi digital", "manajemen program spbe", "pmo sistem informasi"],
     {"pmo": ["project management office", "program management office"], "transformasi digital": ["digital transformation", "tranformasi digital"]}, ["pmo", "pendampingan", "program", "sistem informasi", "spbe"],
     ["transformasi organisasi tanpa teknologi"], "konsultansi organisasi tanpa sistem informasi",
     ["PMO transformasi digital provinsi", "Pendampingan transformasi digital rumah sakit", "Manajemen program SPBE kementerian", "PMO sistem informasi lintas dinas"],
     "Transformasi organisasi perangkat daerah", "Pengadaan seragam tim PMO digital"),
    ("managed-it-services", "Layanan TI Terkelola", "Operasi, pemantauan, dukungan, dan SLA untuk layanan TI.", "P2",
     ["managed service ti", "operasional ti", "helpdesk aplikasi"], ["managed service ti", "operasional layanan ti", "pengelolaan helpdesk aplikasi", "pemantauan infrastruktur ti"],
     {"managed service": ["managed services", "layanan terkelola"], "helpdesk": ["service desk", "help desk"]}, ["operasional", "pengelolaan", "pemantauan", "sla", "dukungan"],
     ["operasional gedung", "katering"], "pengelolaan fasilitas non TI",
     ["Managed service TI kantor pusat", "Operasional layanan TI 24 jam", "Pengelolaan helpdesk aplikasi publik", "Pemantauan infrastruktur TI pemerintah"],
     "Layanan bantuan kantor", "Pengadaan katering operator managed service TI"),
    ("enterprise-systems", "Sistem Enterprise", "Implementasi dan integrasi ERP, HRIS, CRM, serta sistem enterprise sejenis.", "P1",
     ["erp", "hris", "crm", "sistem enterprise"], ["implementasi erp", "pengembangan hris", "integrasi crm", "implementasi sistem enterprise", "pelatihan aplikasi erp"],
     {"erp": ["enterprise resource planning", "e r p"], "hris": ["human resource information system", "h r i s"]}, ["implementasi", "pengembangan", "integrasi", "migrasi", "sistem"],
     ["pelatihan umum tanpa aplikasi khusus"], "lisensi umum tanpa implementasi",
     ["Implementasi ERP badan usaha daerah", "Pengembangan HRIS instansi", "Integrasi CRM layanan warga", "Implementasi sistem enterprise pemerintah"],
     "Pengadaan lisensi ERP", "Pengadaan meja kantor tim ERP"),
    ("network-infrastructure", "Infrastruktur Jaringan", "Desain, implementasi, konfigurasi, integrasi, dan keamanan jaringan TI.", "P2",
     ["jaringan", "network", "sd wan"], ["implementasi jaringan", "konfigurasi jaringan", "integrasi jaringan", "keamanan jaringan"],
     {"sd wan": ["sdwan", "software defined wan"], "jaringan": ["networking", "jarngan"]}, ["implementasi", "konfigurasi", "integrasi", "keamanan", "arsitektur"],
     ["jaringan listrik"], "pengadaan kabel listrik tanpa jaringan TI",
     ["Implementasi jaringan kantor cabang", "Konfigurasi jaringan pusat data", "Integrasi jaringan antar gedung", "Keamanan jaringan instansi"],
     "Belanja perangkat jaringan kantor", "Pengadaan kabel listrik jaringan gedung"),
    ("it-consulting", "Konsultansi TI", "Perencanaan, arsitektur, asesmen, dan strategi teknologi informasi.", "P2",
     ["konsultansi ti", "arsitektur ti", "strategi teknologi informasi"], ["konsultansi arsitektur ti", "penyusunan strategi ti", "asesmen teknologi informasi", "perencanaan sistem informasi"],
     {"teknologi informasi": ["ti", "it"], "arsitektur ti": ["enterprise architecture", "arsitektur teknologi informasi"]}, ["konsultansi", "perencanaan", "asesmen", "strategi", "arsitektur"],
     ["konsultansi manajemen umum"], "nasihat organisasi tanpa teknologi",
     ["Konsultansi arsitektur TI instansi", "Penyusunan strategi TI daerah", "Asesmen teknologi informasi kementerian", "Perencanaan sistem informasi pelayanan"],
     "Konsultansi transformasi organisasi", "Pengadaan alat tulis kantor konsultan TI"),
    ("specialized-digital-solutions", "Solusi Digital Khusus", "Solusi digital spesifik sektor seperti GIS, IoT, identitas digital, dan otomasi.", "P2",
     ["gis", "iot", "identitas digital", "otomasi digital"], ["pengembangan gis", "implementasi iot", "pengembangan identitas digital", "implementasi otomasi digital"],
     {"gis": ["sistem informasi geografis", "geographic information system"], "iot": ["internet of things", "i o t"]}, ["pengembangan", "implementasi", "integrasi", "platform"],
     ["survei geografis manual tanpa sistem"], "pengadaan sensor tanpa integrasi digital",
     ["Pengembangan GIS tata ruang", "Implementasi IoT pemantauan air", "Pengembangan identitas digital warga", "Implementasi otomasi digital perizinan"],
     "Pengadaan sensor IoT", "Pengadaan bibit tanaman untuk proyek GIS"),
]

# Owner-provided master terms. These are discovery candidates, not automatic
# relevance decisions. Generic terms are deliberately kept out of strong rules.
MASTER_KEYWORDS = {
    "software-engineering": "pengembangan aplikasi, pembangunan aplikasi, pembuatan aplikasi, sistem informasi, pengembangan website, pemeliharaan aplikasi, penyempurnaan aplikasi, upgrade aplikasi, rekayasa perangkat lunak, aplikasi berbasis web, mobile application",
    "cybersecurity": "keamanan informasi, keamanan siber, penetration test, pentest, vulnerability assessment, security assessment, SOC, SIEM, vulnerability scanning, incident response, digital forensic, security monitoring",
    "it-grc": "ISO 27001, SMKI, audit keamanan informasi, IT governance, IT risk assessment, manajemen risiko TI, compliance, tata kelola SPBE, evaluasi SPBE, indeks KAMI, IT audit, BCP, DRP",
    "data-ai": "integrasi data, analitik data, data warehouse, data lake, business intelligence, dashboard analitik, dashboard eksekutif, kecerdasan buatan, machine learning, artificial intelligence, big data, data governance",
    "cloud-infrastructure": "cloud computing, cloud server, cloud hosting, virtual private server, VPS, data center, disaster recovery center, virtualisasi server, containerization, Kubernetes, DevOps, infrastruktur TI",
    "system-integration": "integrasi sistem, integrasi aplikasi, API gateway, interoperabilitas, pertukaran data, bridging data, web service, middleware, single sign on, SSO, sinkronisasi data",
    "pmo-digital-transformation": "project management office, PMO, dashboard monitoring, monitoring evaluasi, monev digital, sistem pemantauan program, pengendalian proyek, transformasi digital, manajemen portofolio, digital transformation",
    "managed-it-services": "managed service, managed IT, pemeliharaan sistem, pemeliharaan server, dukungan teknis TI, IT support, operasional aplikasi, pemeliharaan database, monitoring server, technical support",
    "enterprise-systems": "enterprise resource planning, ERP, HRIS, sistem keuangan, sistem akuntansi, financial management system, CRM, procurement management system, e-office, e-document, manajemen aset, e-arsip",
    "network-infrastructure": "konfigurasi firewall, implementasi firewall, network security, implementasi jaringan, instalasi LAN, konfigurasi VPN, network monitoring, network assessment, pengembangan infrastruktur jaringan",
    "it-consulting": "IT master plan, rencana induk TI, blueprint TI, enterprise architecture, audit sistem informasi, kajian teknologi informasi, IT assessment, feasibility study TI, konsultansi transformasi digital",
    "specialized-digital-solutions": "e-government, e-learning, LMS, sistem monitoring, logbook elektronik, sistem perizinan, sistem pelayanan publik, sistem informasi kesehatan, IoT monitoring, dashboard operasional",
}

STRONG_ADDITIONS = {
    "software-engineering": ["pembangunan aplikasi", "pembuatan aplikasi", "pengembangan website", "penyempurnaan aplikasi", "upgrade aplikasi", "rekayasa perangkat lunak"],
    "cybersecurity": ["security assessment", "vulnerability scanning", "incident response", "digital forensic", "security monitoring"],
    "it-grc": ["audit keamanan informasi", "it risk assessment", "evaluasi spbe", "tata kelola spbe", "it audit"],
    "data-ai": ["pembangunan data lake", "pengembangan business intelligence", "pengembangan dashboard analitik", "pengembangan dashboard eksekutif"],
    "cloud-infrastructure": ["implementasi cloud server", "migrasi data center", "virtualisasi server", "implementasi kubernetes"],
    "system-integration": ["integrasi aplikasi", "implementasi api gateway", "bridging data", "sinkronisasi data"],
    "pmo-digital-transformation": ["pmo transformasi digital", "monev digital berbasis sistem", "pengembangan sistem pemantauan program"],
    "managed-it-services": ["pemeliharaan sistem informasi", "pemeliharaan server", "dukungan teknis ti", "operasional aplikasi", "pemeliharaan database", "monitoring server"],
    "enterprise-systems": ["implementasi sistem keuangan", "pengembangan sistem akuntansi", "pengembangan e-office", "implementasi e-arsip"],
    "network-infrastructure": ["konfigurasi firewall", "implementasi firewall", "instalasi lan", "konfigurasi vpn", "network monitoring", "network assessment", "pengembangan infrastruktur jaringan"],
    "it-consulting": ["it master plan", "rencana induk ti", "blueprint ti", "audit sistem informasi", "kajian teknologi informasi", "it assessment", "feasibility study ti", "konsultansi transformasi digital"],
    "specialized-digital-solutions": ["pengembangan e-government", "implementasi e-learning", "pengembangan logbook elektronik", "implementasi sistem perizinan", "pengembangan sistem informasi kesehatan"],
}

POSITIVE_ADDITIONS = {
    "cybersecurity": ["soc", "siem"],
    "it-grc": ["iso 27001", "smki", "it governance", "compliance", "indeks kami", "bcp", "drp"],
    "data-ai": ["data lake", "business intelligence", "dashboard analitik", "dashboard eksekutif", "big data", "data governance"],
    "cloud-infrastructure": ["cloud computing", "cloud server", "cloud hosting", "virtual private server", "vps", "data center", "disaster recovery center", "virtualisasi server", "containerization", "kubernetes", "infrastruktur ti"],
    "system-integration": ["api gateway", "pertukaran data", "web service", "middleware", "single sign on", "sso"],
    "pmo-digital-transformation": ["dashboard monitoring", "monitoring evaluasi", "monev digital", "sistem pemantauan program", "pengendalian proyek", "manajemen portofolio"],
    "managed-it-services": ["managed it", "pemeliharaan sistem", "it support", "technical support"],
    "enterprise-systems": ["sistem keuangan", "sistem akuntansi", "financial management system", "procurement management system", "e-office", "e-document", "manajemen aset", "e-arsip"],
    "network-infrastructure": ["network security"],
    "specialized-digital-solutions": ["e-government", "e-learning", "lms", "sistem monitoring", "logbook elektronik", "sistem perizinan", "sistem pelayanan publik", "sistem informasi kesehatan", "iot monitoring", "dashboard operasional"],
}

CONTEXT_ADDITIONS = {
    "cybersecurity": ["implementasi", "operasional", "pengelolaan"],
    "it-grc": ["penyusunan", "implementasi", "evaluasi"],
    "data-ai": ["pengembangan", "pembangunan", "analitik"],
    "cloud-infrastructure": ["konfigurasi", "pengelolaan", "pemeliharaan"],
    "system-integration": ["integrasi", "implementasi", "pengembangan"],
    "pmo-digital-transformation": ["pengembangan", "berbasis sistem", "pendampingan"],
    "managed-it-services": ["pemeliharaan", "operasional", "dukungan", "pengelolaan"],
    "enterprise-systems": ["implementasi", "pengembangan", "integrasi"],
    "network-infrastructure": ["konfigurasi", "instalasi", "implementasi"],
    "specialized-digital-solutions": ["pengembangan", "implementasi", "integrasi"],
}

GLOBAL_EXCLUSIONS = [
    "alat tulis kantor", "seragam", "katering", "toner printer", "kertas printer", "tinta printer",
    "kendaraan dinas", "bahan makanan", "meja kantor", "kursi kantor", "konsumsi rapat",
    "jasa kebersihan", "pemeliharaan gedung", "renovasi gedung", "alat kesehatan", "obat obatan",
    "bahan bangunan", "bibit tanaman", "pupuk pertanian", "keamanan fisik gedung",
    "perjalanan dinas", "jasa parkir", "fotokopi dokumen", "baju dinas", "jasa laundry",
    "peralatan dapur", "sewa tenda", "pencetakan brosur", "pengadaan air minum", "perbaikan jalan",
    "pembangunan jembatan", "cat tembok", "perabot ruang rapat", "alat olahraga", "jasa taman",
    "pengadaan beras", "sewa kendaraan", "jasa kebersihan toilet", "pembangunan drainase", "pengadaan vaksin", "kabel listrik",
]

categories = []
cases = []
def unique_terms(items):
    return list(dict((item.casefold(), item) for item in items).values())

for spec in SPECS:
    cid, name, desc, priority, positive, strong, synonyms, context, negative, exclusion, titles, ambiguous, outside = spec
    master = [term.strip() for term in MASTER_KEYWORDS[cid].split(",")]
    categories.append({"category_id": cid, "name": name, "business_description": desc,
                       "priority": priority, "discovery_keywords": master,
                       "positive_keywords": unique_terms(positive + POSITIVE_ADDITIONS.get(cid, [])),
                       "strong_phrases": unique_terms(strong + STRONG_ADDITIONS.get(cid, [])),
                       "synonyms": synonyms, "context_keywords": unique_terms(context + CONTEXT_ADDITIONS.get(cid, [])), "negative_keywords": negative,
                       "exclusion_rules": [{"phrase": exclusion, "reason": "Lingkup hanya di luar layanan digital kategori ini."}]})
    for i, title in enumerate(titles, 1):
        cases.append({"id": f"rel-{cid}-{i}", "title": title,
                      "description": "Pekerjaan mencakup implementasi dan serah terima sistem digital.",
                      "procurement_type": ["Jasa Lainnya", "Jasa Konsultansi", "Barang", "Jasa Lainnya"][i-1],
                      "expected_disposition": "relevant", "expected_primary_category": cid,
                      "tags": ["category_positive", "mixed_procurement"] if i == 3 else ["category_positive"]})
    cases.append({"id": f"review-{cid}", "title": ambiguous, "procurement_type": "Barang" if cid in ("network-infrastructure", "enterprise-systems") else "Jasa Konsultansi",
                  "expected_disposition": "needs_review", "category_under_test": cid, "tags": ["category_ambiguous"]})

cases.extend([
    {"id":"rel-multi-security-grc", "title":"Audit tata kelola TI dan keamanan siber", "description":"Asesmen SPBE dan pengujian keamanan aplikasi", "procurement_type":"Jasa Konsultansi", "expected_disposition":"relevant", "expected_primary_category":"it-grc", "expected_secondary_categories":["cybersecurity"], "tags":["multi_category"]},
    {"id":"rel-multi-software-integration", "title":"Pengembangan aplikasi dan integrasi sistem informasi", "description":"Implementasi API antar aplikasi", "procurement_type":"Barang", "expected_disposition":"relevant", "expected_primary_category":"software-engineering", "expected_secondary_categories":["system-integration"], "tags":["multi_category", "mixed_procurement"]},
])

for i, phrase in enumerate(GLOBAL_EXCLUSIONS[:40], 1):
    cases.append({"id": f"excluded-{i:02d}", "title": f"Pengadaan {phrase} untuk kantor daerah", "procurement_type": "Barang" if i % 2 else "Jasa Lainnya",
                  "expected_disposition":"excluded", "tags":["category_negative", "negative_context"]})

review_extra = [
    ("organization", "Jasa Konsultansi Transformasi Organisasi", "Jasa Konsultansi", "category_ambiguous"),
    ("network", "Belanja Perangkat Jaringan Kantor", "Barang", "category_ambiguous"),
    ("training", "Pelatihan Literasi Digital Pegawai", "Jasa Lainnya", "category_ambiguous"),
    ("conflict", "Pengembangan aplikasi dan pengadaan alat tulis kantor", "Jasa Lainnya", "keyword_collision"),
    ("mixed-network", "Implementasi jaringan dan pengadaan seragam teknisi", "Barang", "mixed_procurement"),
    ("mixed-cloud", "Migrasi cloud dan pengadaan kursi kantor", "Jasa Lainnya", "mixed_procurement"),
    ("mixed-security", "Penetration test aplikasi dan pengadaan katering", "Jasa Konsultansi", "keyword_collision"),
    ("mixed-data", "Pengembangan analitik data dan pengadaan toner printer", "Barang", "keyword_collision"),
    ("generic-digital", "Pengadaan solusi digital", "Barang", "insufficient_context"),
    ("generic-platform", "Penyediaan platform teknologi", "Jasa Lainnya", "insufficient_context"),
    ("license-only", "Pengadaan lisensi aplikasi umum", "Barang", "license_scope"),
    ("hardware-only", "Pengadaan server dan perangkat jaringan", "Barang", "hardware_scope"),
    ("consulting-general", "Jasa konsultansi manajemen organisasi", "Jasa Konsultansi", "insufficient_context"),
    ("training-general", "Bimbingan teknis literasi digital", "Jasa Lainnya", "training_scope"),
    ("security-device", "Pengadaan perangkat keamanan", "Barang", "hardware_scope"),
    ("data-entry", "Jasa entri data dokumen", "Jasa Lainnya", "scope_unclear"),
    ("cloud-capacity", "Pengadaan kapasitas cloud", "Barang", "scope_unclear"),
    ("iot-device", "Pengadaan sensor IoT", "Barang", "hardware_scope"),
]
for key, title, procurement, tag in review_extra:
    cases.append({"id": f"review-{key}", "title": title, "procurement_type": procurement,
                  "expected_disposition": "needs_review", "tags": [tag]})

# Preserve the owner's fifteen v1 decisions verbatim as a subset of v2.
baseline = {
    "rel-software-engineering-1": ("Pengembangan Aplikasi Web Pelayanan Publik", "Jasa Lainnya"),
    "rel-software-engineering-2": ("Pembangunan Sistem Informasi Kepegawaian", "Jasa Konsultansi"),
    "rel-data-ai-1": ("Pengadaan Layanan Integrasi Data", "Jasa Lainnya"),
    "rel-data-ai-2": ("Pembangunan Analitik Data Sektoral", "Jasa Konsultansi"),
    "rel-cloud-infrastructure-1": ("Layanan Cloud untuk Platform Pemerintah", "Jasa Lainnya"),
    "rel-cybersecurity-1": ("Jasa Penetration Test Aplikasi", "Jasa Konsultansi"),
    "excluded-01": ("Pengadaan Alat Tulis Kantor Tahun 2026", "Barang"),
    "excluded-02": ("Belanja Seragam Pegawai Lapangan", "Barang"),
    "excluded-03": ("Pengadaan Katering Rapat Koordinasi", "Jasa Lainnya"),
    "excluded-04": ("Pengadaan Toner Printer", "Barang"),
    "excluded-05": ("Belanja Kertas Printer Tahunan", "Barang"),
    "review-organization": ("Jasa Konsultansi Transformasi Organisasi", "Jasa Konsultansi"),
    "review-network": ("Belanja Perangkat Jaringan Kantor", "Barang"),
    "review-training": ("Pelatihan Literasi Digital Pegawai", "Jasa Lainnya"),
    "review-conflict": ("Pengembangan Aplikasi dan Pengadaan Alat Tulis Kantor", "Jasa Lainnya"),
}
for case in cases:
    if case["id"] in baseline:
        case["title"], case["procurement_type"] = baseline[case["id"]]
        case["tags"].append("v1_baseline")

for index, spec in enumerate(SPECS, 6):
    case = next(c for c in cases if c["id"] == f"excluded-{index:02d}")
    case["title"] = spec[12]
    case["category_under_test"] = spec[0]
    case["tags"].append("category_negative")

variants = {
    "rel-software-engineering-3": "Pengembangan apliksi perizinan daerah",
    "rel-cybersecurity-2": "Audit cyber security pusat data",
    "rel-it-grc-2": "Penyusunan tatakelola TI kementerian",
    "rel-data-ai-3": "Implementasi machne learning untuk prediksi",
    "rel-cloud-infrastructure-2": "Implementasi clod privat instansi",
    "rel-system-integration-2": "Pengembangan integrsi API layanan kependudukan",
    "rel-pmo-digital-transformation-2": "Pendampingan tranformasi digital rumah sakit",
    "rel-network-infrastructure-2": "Konfigurasi jarngan pusat data",
    "rel-specialized-digital-solutions-2": "Implementasi internet of things pemantauan air",
}
for case in cases:
    if case["id"] in variants:
        case["title"] = variants[case["id"]]
        case["tags"].append("abbreviation_typo")
    if case["id"] == "rel-network-infrastructure-3":
        case["title"] = "Pengadaan perangkat jaringan antar gedung"
        case["tor"] = "Implementasi jaringan dan integrasi jaringan antar gedung"
        case["tags"].append("tor_context")
    if case["id"] == "rel-cybersecurity-4":
        case["title"] = "Pelatihan keamanan informasi bagi operator aplikasi"
        case["tags"].append("specialist_training")
    if case["id"] == "rel-it-grc-4":
        case["title"] = "Pelatihan tata kelola TI instansi"
        case["tags"].append("specialist_training")
    if case["id"] == "rel-enterprise-systems-4":
        case["title"] = "Pelatihan aplikasi ERP khusus"
        case["tags"].append("specialist_training")
    if case["id"] == "rel-system-integration-1":
        case["expected_secondary_categories"] = ["specialized-digital-solutions"]
        case["tags"].append("multi_category")

assert len(cases) == 120
assert {label: sum(c["expected_disposition"] == label for c in cases) for label in ("relevant", "excluded", "needs_review")} == {"relevant":50, "excluded":40, "needs_review":30}

config = {"version":"cankonix-business-v2-candidate", "status":"candidate", "categories": categories,
          "global_exclusion_phrases": GLOBAL_EXCLUSIONS,
          "ambiguous_phrases": ["transformasi organisasi", "perangkat jaringan", "literasi digital", "lisensi aplikasi umum"]}
# status is review metadata, not a runtime engine field.
config.pop("status")
dataset = {"version":"classification-testcases-v2", "taxonomy_version":config["version"],
           "review_status":"candidate", "provenance":"synthetic, authored for regression; not independently labeled procurement samples",
           "cases": cases}
(OUT / "classification-keywords-v2.json").write_text(json.dumps(config, ensure_ascii=False, indent=2) + "\n")
(OUT / "classification-testcases-v2.json").write_text(json.dumps(dataset, ensure_ascii=False, indent=2) + "\n")

taxonomy = [
    "# CANKONIX Hunter — business taxonomy v2 (candidate)", "",
    "This is a review candidate. The live collector remains on `cankonix-it-v1` / `deterministic-v1`.", "",
    "## Decision policy", "",
    "1. Normalize case and punctuation, preserving word boundaries.",
    "2. Find exact multiword phrases and synonym/typo variants in title, description, KAK/TOR, and requirements.",
    "3. Keep discovery terms separate from decision rules. Require either a strong work phrase or a positive phrase plus a distinct work-context phrase in the same field. A bare domain keyword is insufficient.",
    "4. Apply category negative phrases and explicit exclusion rules. Mixed digital and non-digital scope enters `needs_review`.",
    "5. Use procurement type as supporting evidence only. `Barang` never excludes a case by itself.",
    "6. If several categories qualify, the earliest strong phrase in the title selects the primary; then evidence strength and category ID break ties. Other qualified categories are secondary.",
    "7. `relevant` means clear business fit; `excluded` means clearly outside scope; `needs_review` means ambiguity, insufficient detail, or conflicting scope. Review status is independent of business score.",
    "8. Preserve matched field, phrase, category, and reason code as evidence. No v2 exclusion auto-rejects; a human reviews exclusions before any production policy change.",
    "", "The editable JSON uses stable category IDs and arrays of discovery, positive, strong, context, and negative phrases. The owner's master terms are recorded in `discovery_keywords`; discovery alone never makes a package relevant. Generic words such as dashboard, monitoring, pemeliharaan, aplikasi, and konsultansi always need a verified IT work context. An Admin Dashboard can later edit a draft, validate schema, run regression, and publish an immutable version with actor/time/audit trail. Published versions must never be changed in place.",
    "", "## Categories", "",
]
for spec in SPECS:
    cid, name, desc, priority, positive, strong, synonyms, context, negative, exclusion, titles, ambiguous, outside = spec
    configured = next(category for category in categories if category["category_id"] == cid)
    aliases = "; ".join(f"{key}: {', '.join(values)}" for key, values in synonyms.items())
    taxonomy.extend([
        f"### {name} (`{cid}`, {priority})", "", desc, "",
        f"- Discovery keywords: {', '.join(configured['discovery_keywords'])}.",
        f"- Positive keywords: {', '.join(configured['positive_keywords'])}.",
        f"- Strong work phrases: {', '.join(configured['strong_phrases'])}.",
        f"- Synonyms, abbreviations, typos: {aliases}.",
        f"- Work context: {', '.join(configured['context_keywords'])}.",
        f"- Negative keywords: {', '.join(negative)}.",
        f"- Exclusion rule: `{exclusion}` means scope is outside this category.",
        f"- Relevant example: {titles[0]}.",
        f"- Not relevant example: {outside}.",
        f"- Ambiguous example: {ambiguous} (review the actual scope).", "",
    ])
taxonomy.extend([
    "## Classification and opportunity score", "",
    "Classification is decided before opportunity scoring. The candidate score uses capability fit 35%, scope fit 25%, supplier qualification 20%, value/margin 10%, and delivery readiness 10%. Each component is either a 0–100 value with explicit evidence or `unknown`. The score is normalized over known weights; `data_completeness` is reported separately and is not a calibrated confidence probability. Any incomplete score has `provisional=true`. A case with no known components has no numeric score. Bands are high 80–100, medium 60–79, low 0–59. Mandatory failure flags require evidence; unproven flags remain unknown. Missing data never becomes zero.", "",
    "## Boundaries", "",
    "Only public-source data needed for opportunity monitoring may be collected. KAK/TOR text must be safely extracted and treated as untrusted input. A future provider-agnostic AI adapter may propose structured evidence for `needs_review` cases after deterministic rules, but cannot silently reject a package; a human retains the pursue/pass decision. v2 is not wired to the collector, database, or Admin Dashboard; those require a reviewed migration and release gate after independent validation.", "",
    "## Reproduce the candidate", "",
    "From the repository root run `python3 scripts/generate-classification-v2.py`. Then from `services/api` run `go test ./internal/classificationv2 ./cmd/evaluatev2` and `go run ./cmd/evaluatev2 > ../../docs/classification-v2/classification-evaluation-v2.md`. Review generated changes before accepting rule or label edits.", "",
])
(OUT / "classification-taxonomy-v2.md").write_text("\n".join(taxonomy))

review = [
    "# CANKONIX Hunter — classification review v2", "",
    "Status: **candidate for domain review**. The owner approved the original 15 v1 labels as a temporary baseline, not the expanded v2 scope or production engine.", "",
    "The 120 cases below are synthetic regression fixtures: 50 relevant, 40 excluded, 30 needs_review. Fifteen rows tagged `v1_baseline` preserve the owner's exact v1 titles, procurement types, and labels. The remaining rows require independent review, especially mixed scope, Barang procurement, abbreviations, and negative contexts.", "",
    "For each row, mark APPROVE or CHANGE; record corrected label, primary/secondary categories, and reason. Reviewers should also add real sampled LPSE packages with canonical URLs and separately assigned labels. Do not put credentials or private data in this packet.", "",
    "| ID | Title | Procurement | Expected label | Primary | Secondary | Category under test | Tags | Decision / correction |",
    "|---|---|---|---|---|---|---|---|---|",
]
for c in cases:
    review.append(f"| {c['id']} | {c['title']} | {c['procurement_type']} | {c['expected_disposition']} | {c.get('expected_primary_category','—')} | {', '.join(c.get('expected_secondary_categories',[])) or '—'} | {c.get('category_under_test','—')} | {', '.join(c['tags'])} | — |")
review.extend(["", "## Review sign-off", "", "Reviewer name / role: —", "Review UTC time: —", "Controlled approval reference: —", "Corrections and coverage gaps: —", "Decision on v2 dataset and rule scope: PENDING", ""])
(OUT / "classification-review-v2.md").write_text("\n".join(review))
