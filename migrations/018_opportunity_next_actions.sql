BEGIN;

CREATE TABLE opportunity_next_actions (
    opportunity_id uuid PRIMARY KEY REFERENCES opportunities(id) ON DELETE CASCADE,
    action text NOT NULL CHECK (char_length(action) BETWEEN 3 AND 500),
    due_at timestamptz NOT NULL,
    updated_by uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX opportunity_next_actions_due_idx ON opportunity_next_actions(due_at, updated_at DESC);

INSERT INTO schema_migrations(version) VALUES (18);
COMMIT;
