BEGIN;

CREATE TABLE opportunity_overrides (
    opportunity_id uuid PRIMARY KEY REFERENCES opportunities(id) ON DELETE CASCADE,
    recommendation text CHECK (recommendation IS NULL OR recommendation IN ('relevant','excluded','needs_review')),
    score_bp integer CHECK (score_bp IS NULL OR score_bp BETWEEN 0 AND 10000),
    reason text NOT NULL CHECK (char_length(reason) BETWEEN 3 AND 2000),
    overridden_by uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (recommendation IS NOT NULL OR score_bp IS NOT NULL)
);

INSERT INTO schema_migrations(version) VALUES (20);
COMMIT;
