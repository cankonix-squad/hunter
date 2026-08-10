BEGIN;

CREATE TABLE saved_opportunity_views (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 80),
    filters jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX saved_opportunity_views_user_name_unique
    ON saved_opportunity_views(user_id, lower(name));
CREATE INDEX saved_opportunity_views_user_idx
    ON saved_opportunity_views(user_id, created_at DESC);

INSERT INTO schema_migrations(version) VALUES (13);
COMMIT;
