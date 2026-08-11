BEGIN;

CREATE TABLE alert_rules (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 80),
    enabled boolean NOT NULL DEFAULT true,
    event_types text[] NOT NULL CHECK (cardinality(event_types) BETWEEN 1 AND 2 AND event_types <@ ARRAY['opportunity.new','opportunity.material_change']::text[]),
    filters jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX alert_rules_user_name_unique ON alert_rules(user_id, lower(name));
CREATE INDEX alert_rules_enabled_idx ON alert_rules(enabled) WHERE enabled;

CREATE TABLE alert_events (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    rule_id uuid NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    opportunity_id uuid NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    opportunity_version_id bigint NOT NULL REFERENCES opportunity_versions(id) ON DELETE CASCADE,
    event_type text NOT NULL CHECK (event_type IN ('opportunity.new','opportunity.material_change')),
    changed_fields text[] NOT NULL DEFAULT '{}',
    delivery_status text NOT NULL DEFAULT 'pending' CHECK (delivery_status IN ('pending','suppressed','delivered','failed')),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(rule_id, opportunity_version_id, event_type)
);

CREATE INDEX alert_events_user_created_idx ON alert_events(user_id, created_at DESC);
CREATE INDEX alert_events_pending_idx ON alert_events(created_at) WHERE delivery_status='pending';

INSERT INTO schema_migrations(version) VALUES (24);
COMMIT;
