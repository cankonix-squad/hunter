BEGIN;

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL UNIQUE CHECK (email = lower(email)),
    display_name text NOT NULL,
    password_hash text NOT NULL,
    active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE roles (
    id smallint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text NOT NULL UNIQUE
);

CREATE TABLE user_roles (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id smallint NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE refresh_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    family_id uuid NOT NULL,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    rotated_at timestamptz,
    revoked_at timestamptz,
    replaced_by uuid REFERENCES refresh_sessions(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (expires_at > created_at)
);

CREATE INDEX refresh_sessions_user_idx ON refresh_sessions (user_id, expires_at DESC);
CREATE INDEX refresh_sessions_family_idx ON refresh_sessions (family_id);

INSERT INTO roles(name) VALUES
    ('administrator'),
    ('operator'),
    ('business_development'),
    ('bid_manager'),
    ('technical_lead'),
    ('executive');

INSERT INTO schema_migrations(version) VALUES (2);

COMMIT;
