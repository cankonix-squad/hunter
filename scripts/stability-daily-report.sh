#!/bin/sh
set -eu

if [ -n "${DATABASE_URL:-}" ] && [ -n "${DATABASE_URL_FILE:-}" ]; then
  echo "DATABASE_URL and DATABASE_URL_FILE cannot both be set" >&2
  exit 1
fi
if [ -z "${DATABASE_URL:-}" ] && [ -n "${DATABASE_URL_FILE:-}" ]; then
  test -f "$DATABASE_URL_FILE" || { echo "DATABASE_URL_FILE must reference a regular file" >&2; exit 1; }
  DATABASE_URL=$(sed -e 's/\r$//' "$DATABASE_URL_FILE")
fi
test -n "${DATABASE_URL:-}" || { echo "DATABASE_URL or DATABASE_URL_FILE is required" >&2; exit 1; }
command -v psql >/dev/null 2>&1 || { echo "psql is required" >&2; exit 1; }

observation_date=${1:-$(date -u +%F)}
case "$observation_date" in ????-??-??) ;; *) echo "date must use YYYY-MM-DD" >&2; exit 1 ;; esac

result=$(psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -v observation_date="$observation_date" -AtF '|' <<'SQL'
WITH priority AS (
  SELECT id FROM sources WHERE status='active'
    AND coalesce((policy->>'schedule_enabled')::boolean,false)
), daily AS (
  SELECT p.id,
    count(r.id) AS runs,
    count(*) FILTER (WHERE r.status='succeeded') AS succeeded,
    count(*) FILTER (WHERE r.status='dead_letter') AS dead_letters
  FROM priority p LEFT JOIN source_runs r ON r.source_id=p.id
    AND r.created_at >= (:'observation_date')::timestamp AT TIME ZONE 'UTC'
    AND r.created_at < ((:'observation_date')::date+1)::timestamp AT TIME ZONE 'UTC'
  GROUP BY p.id
)
SELECT (SELECT count(*) FROM priority),
  count(*) FILTER (WHERE succeeded>0),
  coalesce(sum(runs),0),coalesce(sum(succeeded),0),coalesce(sum(dead_letters),0),
  CASE WHEN (:'observation_date')::date < (now() AT TIME ZONE 'UTC')::date THEN 1 ELSE 0 END
FROM daily;
SQL
)
IFS='|' read -r priority_sources successful_sources runs succeeded dead_letters day_complete <<EOF
$result
EOF

cat <<EOF
CANKONIX daily stability checkpoint
date_utc=$observation_date
priority_sources=$priority_sources
successful_sources=$successful_sources
runs=$runs
succeeded=$succeeded
dead_letters=$dead_letters
day_complete=$day_complete
EOF

if [ "$priority_sources" -ne 20 ]; then
  echo "status=FAILED reason=priority_source_count"
  exit 1
fi
if [ "$dead_letters" -ne 0 ]; then
  echo "status=FAILED reason=dead_letter"
  exit 1
fi
if [ "$day_complete" -ne 1 ] || [ "$successful_sources" -ne 20 ]; then
  echo "status=PENDING reason=utc_day_or_source_runs_incomplete"
  exit 2
fi
echo "status=PASSED"
