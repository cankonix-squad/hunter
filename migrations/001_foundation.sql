BEGIN;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE schema_migrations (
    version bigint PRIMARY KEY,
    applied_at timestamptz NOT NULL DEFAULT now()
);

CREATE TYPE source_status AS ENUM ('active', 'paused', 'degraded', 'disabled');
CREATE TYPE run_status AS ENUM ('queued', 'running', 'succeeded', 'failed', 'dead_letter');

CREATE TABLE sources (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    base_url text NOT NULL UNIQUE,
    timezone text NOT NULL DEFAULT 'Asia/Jakarta',
    adapter_name text NOT NULL,
    adapter_version text NOT NULL,
    status source_status NOT NULL DEFAULT 'disabled',
    policy jsonb NOT NULL DEFAULT '{}'::jsonb,
    health_score smallint NOT NULL DEFAULT 100 CHECK (health_score BETWEEN 0 AND 100),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE source_runs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id uuid NOT NULL REFERENCES sources(id),
    idempotency_key text NOT NULL UNIQUE,
    policy_version text NOT NULL,
    status run_status NOT NULL DEFAULT 'queued',
    attempts integer NOT NULL DEFAULT 0,
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    started_at timestamptz,
    finished_at timestamptz,
    error_code text,
    safe_error text,
    metrics jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE audit_logs (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    actor_id uuid,
    action text NOT NULL,
    object_type text NOT NULL,
    object_id text NOT NULL,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    occurred_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX source_runs_claim_idx ON source_runs (next_attempt_at, created_at)
    WHERE status IN ('queued', 'failed');
CREATE INDEX audit_logs_object_idx ON audit_logs (object_type, object_id, occurred_at DESC);

INSERT INTO schema_migrations(version) VALUES (1);

COMMIT;
