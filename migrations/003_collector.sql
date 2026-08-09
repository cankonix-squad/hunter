BEGIN;

ALTER TABLE sources
    ADD COLUMN policy_version text NOT NULL DEFAULT 'v1',
    ADD COLUMN consecutive_failures integer NOT NULL DEFAULT 0,
    ADD COLUMN circuit_open_until timestamptz,
    ADD COLUMN last_success_at timestamptz,
    ADD COLUMN last_failure_at timestamptz;

ALTER TABLE source_runs
    ADD COLUMN lease_owner text,
    ADD COLUMN lease_expires_at timestamptz;

CREATE TABLE source_snapshots (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id uuid NOT NULL REFERENCES sources(id),
    source_run_id uuid NOT NULL REFERENCES source_runs(id),
    package_identity text NOT NULL,
    fingerprint text NOT NULL,
    canonical_url text NOT NULL,
    source_timestamp timestamptz,
    observed_at timestamptz NOT NULL DEFAULT now(),
    safe_metadata jsonb NOT NULL,
    UNIQUE (source_id, package_identity, fingerprint)
);

CREATE INDEX source_snapshots_identity_idx
    ON source_snapshots (source_id, package_identity, observed_at DESC);

CREATE INDEX source_runs_lease_idx
    ON source_runs (lease_expires_at)
    WHERE status = 'running';

INSERT INTO schema_migrations(version) VALUES (3);

COMMIT;
