BEGIN;

CREATE TABLE opportunity_comments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    opportunity_id uuid NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    author_user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    body text NOT NULL CHECK (char_length(body) BETWEEN 1 AND 4000),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX opportunity_comments_timeline_idx
    ON opportunity_comments(opportunity_id, created_at, id);

INSERT INTO schema_migrations(version) VALUES (15);
COMMIT;
