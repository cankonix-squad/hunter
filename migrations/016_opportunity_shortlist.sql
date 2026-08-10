BEGIN;

CREATE TABLE opportunity_shortlist (
    opportunity_id uuid PRIMARY KEY REFERENCES opportunities(id) ON DELETE CASCADE,
    shortlisted_by uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX opportunity_shortlist_updated_idx ON opportunity_shortlist(updated_at DESC);

INSERT INTO schema_migrations(version) VALUES (16);
COMMIT;
