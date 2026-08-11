BEGIN;

ALTER TABLE notification_deliveries
    ADD COLUMN replay_count integer NOT NULL DEFAULT 0 CHECK (replay_count >= 0),
    ADD COLUMN last_replayed_at timestamptz,
    ADD COLUMN last_replayed_by uuid REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX notification_deliveries_dead_letter_idx
    ON notification_deliveries(updated_at DESC)
    WHERE status='dead_letter';

INSERT INTO schema_migrations(version) VALUES (29);
COMMIT;
