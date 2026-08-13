BEGIN;

CREATE TEMP TABLE fixture_sources ON COMMIT DROP AS
SELECT id
FROM sources
WHERE adapter_name = 'spse-fixture'
   OR lower(name) LIKE '%fixture%';

CREATE TEMP TABLE fixture_opportunities ON COMMIT DROP AS
SELECT id
FROM opportunities
WHERE source_id IN (SELECT id FROM fixture_sources);

DELETE FROM opportunity_scores
WHERE opportunity_id IN (SELECT id FROM fixture_opportunities);

DELETE FROM opportunity_classifications
WHERE opportunity_id IN (SELECT id FROM fixture_opportunities);

DELETE FROM opportunity_versions
WHERE opportunity_id IN (SELECT id FROM fixture_opportunities);

DELETE FROM opportunities
WHERE id IN (SELECT id FROM fixture_opportunities);

DELETE FROM source_snapshots
WHERE source_id IN (SELECT id FROM fixture_sources);

DELETE FROM source_runs
WHERE source_id IN (SELECT id FROM fixture_sources);

DELETE FROM sources
WHERE id IN (SELECT id FROM fixture_sources);

INSERT INTO schema_migrations(version) VALUES (31);
COMMIT;
