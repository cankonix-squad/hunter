#!/bin/sh
set -eu

if [ -n "${DATABASE_URL:-}" ] && [ -n "${DATABASE_URL_FILE:-}" ]; then
  echo "DATABASE_URL and DATABASE_URL_FILE cannot both be set" >&2
  exit 1
fi
if [ -z "${DATABASE_URL:-}" ] && [ -n "${DATABASE_URL_FILE:-}" ]; then
  if [ ! -f "$DATABASE_URL_FILE" ]; then
    echo "DATABASE_URL_FILE must reference a regular file" >&2
    exit 1
  fi
  DATABASE_URL=$(sed -e 's/\r$//' "$DATABASE_URL_FILE")
fi
if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL or DATABASE_URL_FILE is required" >&2
  exit 1
fi
if ! command -v psql >/dev/null 2>&1; then
  echo "psql is required" >&2
  exit 1
fi

end_date=${1:-$(date -u +%F)}
case "$end_date" in
  ????-??-??) ;;
  *) echo "end date must use YYYY-MM-DD" >&2; exit 1 ;;
esac

result=$(psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -v end_date="$end_date" -AtF '|' <<'SQL'
WITH bounds AS (
  SELECT (:'end_date')::date AS end_date, (:'end_date')::date - 6 AS start_date
), priority AS (
  SELECT id, name, consecutive_failures, circuit_open_until
  FROM sources
  WHERE status = 'active'
    AND coalesce((policy->>'schedule_enabled')::boolean, false)
    AND coalesce((policy->>'priority_source')::boolean, false)
), days AS (
  SELECT generate_series(b.start_date, b.end_date, interval '1 day')::date AS day
  FROM bounds b
), daily AS (
  SELECT p.id, p.name, d.day,
    count(r.id) AS runs,
    count(*) FILTER (WHERE r.status = 'succeeded') AS succeeded,
    count(*) FILTER (WHERE r.status = 'dead_letter') AS dead_letters
  FROM priority p CROSS JOIN days d
  LEFT JOIN source_runs r ON r.source_id = p.id
    AND (r.created_at AT TIME ZONE 'UTC')::date = d.day
  GROUP BY p.id, p.name, d.day
), totals AS (
  SELECT
    (SELECT count(*) FROM priority) AS priority_sources,
    count(*) FILTER (WHERE runs = 0 OR succeeded = 0) AS missing_source_days,
    coalesce(sum(runs), 0) AS runs,
    coalesce(sum(succeeded), 0) AS succeeded,
    coalesce(sum(dead_letters), 0) AS dead_letters
  FROM daily
), state AS (
  SELECT
    count(*) FILTER (WHERE consecutive_failures > 0) AS sources_with_failures,
    count(*) FILTER (WHERE circuit_open_until IS NOT NULL AND circuit_open_until > now()) AS open_circuits
  FROM priority
), interventions AS (
  SELECT count(*) AS changes
  FROM audit_logs a, bounds b
  WHERE a.action = 'source.updated'
    AND a.occurred_at >= b.start_date::timestamp AT TIME ZONE 'UTC'
    AND a.occurred_at < (b.end_date + 1)::timestamp AT TIME ZONE 'UTC'
)
SELECT b.start_date, b.end_date, t.priority_sources, t.missing_source_days,
  t.runs, t.succeeded, t.dead_letters, s.sources_with_failures, s.open_circuits,
  i.changes,
  CASE WHEN t.runs = 0 THEN 0 ELSE round(t.succeeded * 100.0 / t.runs, 2) END
FROM bounds b CROSS JOIN totals t CROSS JOIN state s CROSS JOIN interventions i;
SQL
)

IFS='|' read -r start_date end_date_actual priority_sources missing_source_days runs succeeded \
  dead_letters sources_with_failures open_circuits interventions success_rate <<EOF
$result
EOF

cat <<EOF
CANKONIX seven-day stability observation
window_utc=$start_date..$end_date_actual
priority_sources=$priority_sources
source_days_expected=140
missing_source_days=$missing_source_days
runs=$runs
succeeded=$succeeded
success_rate_percent=$success_rate
dead_letters=$dead_letters
sources_with_consecutive_failures=$sources_with_failures
open_circuits=$open_circuits
source_configuration_interventions=$interventions
EOF

if [ "$priority_sources" -ne 20 ]; then
  echo "status=INCOMPLETE reason=priority_source_count"
  exit 2
fi
if [ "$dead_letters" -ne 0 ] || [ "$sources_with_failures" -ne 0 ] || \
   [ "$open_circuits" -ne 0 ] || [ "$interventions" -ne 0 ]; then
  echo "status=FAILED reason=operational_instability"
  exit 1
fi
if ! awk -v rate="$success_rate" 'BEGIN { exit !(rate >= 95) }'; then
  echo "status=FAILED reason=success_rate_below_95_percent"
  exit 1
fi
if [ "$missing_source_days" -ne 0 ]; then
  echo "status=INCOMPLETE reason=seven_calendar_days_not_complete"
  exit 2
fi

echo "status=PASSED"
