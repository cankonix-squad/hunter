BEGIN;

CREATE TABLE opportunity_assignments (
    opportunity_id uuid PRIMARY KEY REFERENCES opportunities(id) ON DELETE CASCADE,
    assignee_user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    assigned_by uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    assigned_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX opportunity_assignments_assignee_idx
    ON opportunity_assignments(assignee_user_id, updated_at DESC);

INSERT INTO schema_migrations(version) VALUES (14);
COMMIT;
