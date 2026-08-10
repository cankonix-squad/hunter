BEGIN;

CREATE TABLE opportunity_readiness_items (
    opportunity_id uuid NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    checklist_version text NOT NULL,
    item_code text NOT NULL,
    completed boolean NOT NULL DEFAULT false,
    note text NOT NULL DEFAULT '' CHECK (char_length(note) <= 1000),
    updated_by uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (opportunity_id, checklist_version, item_code)
);

CREATE INDEX opportunity_readiness_incomplete_idx
    ON opportunity_readiness_items(opportunity_id, updated_at DESC) WHERE NOT completed;

INSERT INTO schema_migrations(version) VALUES (19);
COMMIT;
