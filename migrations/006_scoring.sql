BEGIN;

CREATE TABLE scoring_profiles (
    version text PRIMARY KEY,
    description text NOT NULL,
    weights jsonb NOT NULL CHECK (jsonb_typeof(weights) = 'object'),
    active boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE opportunity_scores (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    opportunity_id uuid NOT NULL REFERENCES opportunities(id),
    opportunity_fingerprint text NOT NULL,
    classification_id bigint NOT NULL REFERENCES opportunity_classifications(id),
    scoring_version text NOT NULL REFERENCES scoring_profiles(version),
    total_score_bp smallint NOT NULL CHECK (total_score_bp BETWEEN 0 AND 10000),
    confidence_bp smallint NOT NULL CHECK (confidence_bp BETWEEN 0 AND 10000),
    needs_review boolean NOT NULL,
    reason_code text NOT NULL,
    explanation text NOT NULL,
    evaluated_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (opportunity_id, opportunity_fingerprint, scoring_version)
);

CREATE TABLE score_dimensions (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    score_id bigint NOT NULL REFERENCES opportunity_scores(id) ON DELETE CASCADE,
    dimension_code text NOT NULL CHECK (dimension_code IN ('capability_fit', 'value', 'location', 'deadline', 'risk')),
    weight_bp smallint NOT NULL CHECK (weight_bp BETWEEN 0 AND 10000),
    raw_score_bp smallint NOT NULL CHECK (raw_score_bp BETWEEN 0 AND 10000),
    contribution smallint NOT NULL CHECK (contribution BETWEEN 0 AND 10000),
    reason_code text NOT NULL,
    explanation text NOT NULL,
    evidence text NOT NULL,
    UNIQUE (score_id, dimension_code)
);

CREATE INDEX opportunity_scores_latest_idx ON opportunity_scores(opportunity_id, evaluated_at DESC);
CREATE INDEX opportunity_scores_review_idx ON opportunity_scores(total_score_bp DESC, evaluated_at DESC) WHERE needs_review;

INSERT INTO scoring_profiles(version, description, weights, active) VALUES
('explainable-v1', 'Deterministic five-dimension opportunity scoring.',
 '{"capability_fit":4000,"value":2000,"location":1000,"deadline":1500,"risk":1500}', true);

INSERT INTO schema_migrations(version) VALUES (6);
COMMIT;
