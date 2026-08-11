BEGIN;

CREATE TABLE notification_deliveries (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    alert_event_id bigint NOT NULL REFERENCES alert_events(id) ON DELETE CASCADE,
    channel text NOT NULL CHECK (channel IN ('email','webhook')),
    destination text NOT NULL,
    status text NOT NULL DEFAULT 'queued' CHECK (status IN ('queued','running','delivered','failed','dead_letter')),
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    lease_owner text,
    lease_expires_at timestamptz,
    safe_error text,
    delivered_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(alert_event_id, channel)
);

CREATE INDEX notification_deliveries_ready_idx
    ON notification_deliveries(next_attempt_at, created_at)
    WHERE status IN ('queued','failed');

INSERT INTO schema_migrations(version) VALUES (27);
COMMIT;
