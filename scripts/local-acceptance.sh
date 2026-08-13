#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_dir"

compose() {
  docker compose "$@"
}

for service in postgres api collector notifier web edge; do
  container_id=$(compose ps -q "$service")
  test -n "$container_id" || { echo "FAIL: $service container is missing" >&2; exit 1; }
  running=$(docker inspect -f '{{.State.Running}}' "$container_id")
  test "$running" = true || { echo "FAIL: $service is not running" >&2; exit 1; }
done

for service in postgres api collector; do
  container_id=$(compose ps -q "$service")
  health=$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}missing{{end}}' "$container_id")
  test "$health" = healthy || { echo "FAIL: $service health=$health" >&2; exit 1; }
done

root_code=$(curl -fsS -o /dev/null -w '%{http_code}' http://localhost:3000/)
health_code=$(curl -fsS -o /dev/null -w '%{http_code}' http://localhost:3000/healthz)
ready_code=$(curl -fsS -o /dev/null -w '%{http_code}' http://localhost:3000/readyz)
auth_code=$(curl -sS -o /dev/null -w '%{http_code}' http://localhost:3000/api/v1/opportunities)
test "$root_code:$health_code:$ready_code:$auth_code" = "200:200:200:401" || {
  echo "FAIL: HTTP root/health/ready/auth=$root_code/$health_code/$ready_code/$auth_code" >&2
  exit 1
}

compose exec -T api wget -q -O - http://127.0.0.1:8080/metrics | grep -q 'cankonix_operational_metrics_scrape_success 1'
compose exec -T collector wget -q -O - http://127.0.0.1:8081/metrics | grep -q 'cankonix_collector_scheduler_enabled 1'

database_result=$(compose exec -T postgres sh -lc 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -X -AtF "|" -v ON_ERROR_STOP=1 -c "SELECT (SELECT max(version) FROM schema_migrations),(SELECT count(*) FROM sources WHERE status='"'"'active'"'"' AND coalesce((policy->>'"'"'schedule_enabled'"'"')::boolean,false)),(SELECT count(*) FROM opportunities WHERE lifecycle_status NOT IN ('"'"'announced'"'"','"'"'registration_open'"'"','"'"'registration_closed'"'"','"'"'evaluation'"'"','"'"'awarded'"'"','"'"'contracted'"'"','"'"'completed'"'"','"'"'cancelled'"'"')),(SELECT count(*) FROM source_runs WHERE status='"'"'dead_letter'"'"' AND finished_at >= now()-interval '"'"'1 hour'"'"')"')
IFS='|' read -r migration priority_sources invalid_lifecycle_values recent_dead_letters <<EOF
$database_result
EOF
test "$migration" -eq 31 || { echo "FAIL: migration=$migration, want 31" >&2; exit 1; }
test "$priority_sources" -ge 20 || { echo "FAIL: active scheduled sources=$priority_sources, want at least 20" >&2; exit 1; }
test "$invalid_lifecycle_values" -eq 0 || { echo "FAIL: invalid opportunity lifecycle values=$invalid_lifecycle_values" >&2; exit 1; }
test "$recent_dead_letters" -eq 0 || { echo "FAIL: recent source dead letters=$recent_dead_letters" >&2; exit 1; }

cat <<EOF
CANKONIX local acceptance: PASS
services=6/6 running
healthchecks=postgres,api,collector healthy
http=root 200, health 200, ready 200, protected API 401
migration=$migration
active_scheduled_sources=$priority_sources
recent_source_dead_letters=$recent_dead_letters
EOF
