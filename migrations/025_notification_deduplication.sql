BEGIN;

DELETE FROM alert_events duplicate
USING alert_events retained
WHERE duplicate.user_id = retained.user_id
  AND duplicate.opportunity_version_id = retained.opportunity_version_id
  AND duplicate.event_type = retained.event_type
  AND duplicate.id > retained.id;

ALTER TABLE alert_events
    ADD CONSTRAINT alert_events_user_version_event_unique
    UNIQUE(user_id, opportunity_version_id, event_type);

INSERT INTO schema_migrations(version) VALUES (25);
COMMIT;
