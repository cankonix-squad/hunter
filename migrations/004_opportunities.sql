BEGIN;

CREATE TABLE opportunities (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id uuid NOT NULL REFERENCES sources(id),
    package_identity text NOT NULL,
    title text NOT NULL,
    agency text NOT NULL,
    procurement_type text NOT NULL,
    lifecycle_status text NOT NULL CHECK (lifecycle_status IN (
        'announced', 'registration_open', 'registration_closed', 'evaluation',
        'awarded', 'contracted', 'completed', 'cancelled'
    )),
    budget_minor bigint CHECK (budget_minor IS NULL OR budget_minor >= 0),
    currency text NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    registration_deadline timestamptz,
    canonical_url text NOT NULL,
    source_timestamp timestamptz,
    first_seen_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    current_fingerprint text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (source_id, package_identity)
);

CREATE TABLE opportunity_versions (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    opportunity_id uuid NOT NULL REFERENCES opportunities(id),
    snapshot_id uuid NOT NULL REFERENCES source_snapshots(id),
    fingerprint text NOT NULL,
    lifecycle_status text NOT NULL,
    normalized_data jsonb NOT NULL,
    observed_at timestamptz NOT NULL,
    UNIQUE (opportunity_id, snapshot_id)
);

CREATE INDEX opportunities_active_feed_idx
    ON opportunities (registration_deadline, last_seen_at DESC)
    WHERE lifecycle_status IN ('announced', 'registration_open');
CREATE INDEX opportunities_search_idx
    ON opportunities USING gin ((title || ' ' || agency || ' ' || package_identity) gin_trgm_ops);
CREATE INDEX opportunity_versions_history_idx
    ON opportunity_versions (opportunity_id, observed_at DESC);

INSERT INTO schema_migrations(version) VALUES (4);

COMMIT;
