BEGIN;

CREATE TABLE notification_preferences (
    user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    quiet_hours_enabled boolean NOT NULL DEFAULT false,
    quiet_hours_start time NOT NULL DEFAULT '22:00',
    quiet_hours_end time NOT NULL DEFAULT '07:00',
    timezone text NOT NULL DEFAULT 'Asia/Jakarta' CHECK (char_length(timezone) BETWEEN 1 AND 64),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (quiet_hours_start <> quiet_hours_end)
);

ALTER TABLE alert_events
    ADD COLUMN available_after timestamptz NOT NULL DEFAULT now();

CREATE INDEX alert_events_delivery_ready_idx
    ON alert_events(available_after, created_at)
    WHERE delivery_status='pending';

INSERT INTO schema_migrations(version) VALUES (26);
COMMIT;
