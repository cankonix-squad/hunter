BEGIN;

ALTER TABLE notification_deliveries
    ADD COLUMN response_code integer CHECK (response_code IS NULL OR response_code BETWEEN 100 AND 599);

INSERT INTO schema_migrations(version) VALUES (28);
COMMIT;
