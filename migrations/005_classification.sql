BEGIN;

CREATE TABLE taxonomies (
    version text PRIMARY KEY,
    name text NOT NULL,
    description text NOT NULL,
    active boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE taxonomy_categories (
    taxonomy_version text NOT NULL REFERENCES taxonomies(version),
    code text NOT NULL,
    name text NOT NULL,
    description text NOT NULL,
    PRIMARY KEY (taxonomy_version, code)
);

CREATE TABLE classification_rules (
    taxonomy_version text NOT NULL REFERENCES taxonomies(version),
    rule_version text NOT NULL,
    code text NOT NULL,
    category_code text,
    kind text NOT NULL CHECK (kind IN ('include', 'exclude')),
    keywords jsonb NOT NULL CHECK (jsonb_typeof(keywords) = 'array'),
    human_reason text NOT NULL,
    approved_auto_reject boolean NOT NULL DEFAULT false,
    PRIMARY KEY (taxonomy_version, rule_version, code),
    FOREIGN KEY (taxonomy_version, category_code) REFERENCES taxonomy_categories(taxonomy_version, code),
    CHECK (kind = 'include' OR category_code IS NULL),
    CHECK (kind = 'exclude' OR approved_auto_reject = false)
);

CREATE TABLE opportunity_classifications (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    opportunity_id uuid NOT NULL REFERENCES opportunities(id),
    opportunity_fingerprint text NOT NULL,
    taxonomy_version text NOT NULL REFERENCES taxonomies(version),
    rule_version text NOT NULL,
    disposition text NOT NULL CHECK (disposition IN ('relevant', 'excluded', 'needs_review')),
    category_code text,
    confidence_bp smallint NOT NULL CHECK (confidence_bp BETWEEN 0 AND 10000),
    reason_code text NOT NULL,
    explanation text NOT NULL,
    auto_rejected boolean NOT NULL DEFAULT false,
    evaluated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (opportunity_id, opportunity_fingerprint, taxonomy_version, rule_version),
    FOREIGN KEY (taxonomy_version, category_code) REFERENCES taxonomy_categories(taxonomy_version, code),
    CHECK (disposition = 'excluded' OR auto_rejected = false)
);

CREATE TABLE classification_evidence (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    classification_id bigint NOT NULL REFERENCES opportunity_classifications(id) ON DELETE CASCADE,
    rule_code text NOT NULL,
    kind text NOT NULL CHECK (kind IN ('include', 'exclude')),
    category_code text,
    keyword text NOT NULL,
    source_field text NOT NULL CHECK (source_field IN ('title', 'procurement_type')),
    excerpt text NOT NULL
);

CREATE INDEX opportunity_classifications_latest_idx ON opportunity_classifications(opportunity_id, evaluated_at DESC);
CREATE INDEX opportunity_classifications_review_idx ON opportunity_classifications(evaluated_at DESC) WHERE disposition = 'needs_review';

INSERT INTO taxonomies(version, name, description, active) VALUES
('cankonix-it-v1', 'CANKONIX IT Services and Products', 'Taxonomy layanan dan produk IT yang menjadi kapabilitas inti CANKONIX.', true);

INSERT INTO taxonomy_categories(taxonomy_version, code, name, description) VALUES
('cankonix-it-v1', 'software-engineering', 'Software Engineering', 'Aplikasi, sistem informasi, dan integrasi sistem.'),
('cankonix-it-v1', 'data-ai', 'Data and AI', 'Data platform, analytics, machine learning, dan AI.'),
('cankonix-it-v1', 'cloud-infrastructure', 'Cloud and Infrastructure', 'Cloud platform, DevOps, dan container platform.'),
('cankonix-it-v1', 'cybersecurity', 'Cybersecurity', 'Assessment dan layanan keamanan siber.');

INSERT INTO classification_rules(taxonomy_version, rule_version, code, category_code, kind, keywords, human_reason, approved_auto_reject) VALUES
('cankonix-it-v1','deterministic-v1','include_software','software-engineering','include','["pengembangan aplikasi","aplikasi web","sistem informasi","integrasi sistem","software"]','Paket menyebut pekerjaan pengembangan atau integrasi perangkat lunak.',false),
('cankonix-it-v1','deterministic-v1','include_data','data-ai','include','["integrasi data","data warehouse","analitik data","artificial intelligence","machine learning","kecerdasan buatan"]','Paket menyebut kapabilitas data atau AI.',false),
('cankonix-it-v1','deterministic-v1','include_cloud','cloud-infrastructure','include','["cloud","komputasi awan","devops","container platform"]','Paket menyebut kapabilitas cloud atau platform.',false),
('cankonix-it-v1','deterministic-v1','include_security','cybersecurity','include','["keamanan siber","cyber security","penetration test","security assessment"]','Paket menyebut layanan keamanan siber.',false),
('cankonix-it-v1','deterministic-v1','exclude_non_it_supply',NULL,'exclude','["alat tulis kantor","seragam","katering","kendaraan dinas","bahan makanan"]','Paket secara deterministik merupakan pengadaan non-IT.',true),
('cankonix-it-v1','deterministic-v1','exclude_it_consumable',NULL,'exclude','["toner printer","tinta printer","kertas printer"]','Paket hanya menyebut barang habis pakai IT dan bukan layanan/produk inti CANKONIX.',true);

INSERT INTO schema_migrations(version) VALUES (5);
COMMIT;
