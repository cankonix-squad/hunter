BEGIN;

CREATE TYPE opportunity_decision AS ENUM ('pursue', 'pass');

CREATE TABLE opportunity_decisions (
    opportunity_id uuid PRIMARY KEY REFERENCES opportunities(id) ON DELETE CASCADE,
    decision opportunity_decision NOT NULL,
    reason text NOT NULL CHECK (char_length(reason) BETWEEN 3 AND 2000),
    decided_by uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    decided_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX opportunity_decisions_pipeline_idx ON opportunity_decisions(decision, updated_at DESC);

INSERT INTO schema_migrations(version) VALUES (17);
COMMIT;
