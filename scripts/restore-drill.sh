#!/bin/sh
set -eu

if [ -z "${RESTORE_DATABASE_URL:-}" ]; then
  echo "RESTORE_DATABASE_URL is required" >&2
  exit 1
fi
if [ "$#" -ne 1 ]; then
  echo "usage: $0 BACKUP_DUMP" >&2
  exit 1
fi
for command_name in pg_restore psql; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "$command_name is required" >&2
    exit 1
  fi
done

dump_file=$1
prefix=${dump_file%.dump}
manifest_file="$prefix.manifest"
checksum_file="$prefix.sha256"
for required_file in "$dump_file" "$manifest_file" "$checksum_file"; do
  if [ ! -f "$required_file" ]; then
    echo "required backup artifact is missing: $required_file" >&2
    exit 1
  fi
done

expected_checksum=$(awk 'NR==1 {print $1}' "$checksum_file")
if command -v sha256sum >/dev/null 2>&1; then
  actual_checksum=$(sha256sum "$dump_file" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
  actual_checksum=$(shasum -a 256 "$dump_file" | awk '{print $1}')
else
  echo "sha256sum or shasum is required" >&2
  exit 1
fi
if [ "$actual_checksum" != "$expected_checksum" ]; then
  echo "backup checksum mismatch" >&2
  exit 1
fi
pg_restore --list "$dump_file" >/dev/null

existing_tables=$(psql "$RESTORE_DATABASE_URL" -X -v ON_ERROR_STOP=1 -Atc \
  "SELECT count(*) FROM pg_tables WHERE schemaname='public'")
if [ "$existing_tables" != "0" ]; then
  echo "restore target must be an empty database; found $existing_tables public tables" >&2
  exit 1
fi

pg_restore --dbname="$RESTORE_DATABASE_URL" --exit-on-error --no-owner --no-acl "$dump_file"

expected_schema=$(sed -n 's/^schema_state=//p' "$manifest_file")
actual_schema=$(psql "$RESTORE_DATABASE_URL" -X -v ON_ERROR_STOP=1 -Atc \
  "SELECT count(*)::text || ':' || max(version)::text FROM schema_migrations")
if [ "$actual_schema" != "$expected_schema" ]; then
  echo "schema state mismatch: expected $expected_schema, found $actual_schema" >&2
  exit 1
fi

expected_counts=$(sed -n 's/^table_counts=//p' "$manifest_file")
actual_counts=$(psql "$RESTORE_DATABASE_URL" -X -v ON_ERROR_STOP=1 -Atc \
  "SELECT coalesce(jsonb_object_agg(tablename,row_count ORDER BY tablename),'{}'::jsonb)::text
   FROM (SELECT tablename,(xpath('/row/count/text()',query_to_xml(format('SELECT count(*) AS count FROM %I.%I',schemaname,tablename),false,true,'')))[1]::text::bigint AS row_count
         FROM pg_tables WHERE schemaname='public') counts")
if [ "$actual_counts" != "$expected_counts" ]; then
  echo "restored table counts do not match backup manifest" >&2
  exit 1
fi

psql "$RESTORE_DATABASE_URL" -X -v ON_ERROR_STOP=1 >/dev/null <<'SQL'
SELECT 1 / CASE WHEN NOT EXISTS (
  SELECT 1 FROM opportunities
  WHERE lifecycle_status NOT IN ('announced','registration_open','registration_closed','evaluation','awarded','contracted','completed','cancelled')
) THEN 1 ELSE 0 END;
SELECT 1 / CASE WHEN NOT EXISTS (
  SELECT 1 FROM opportunity_versions v LEFT JOIN opportunities o ON o.id=v.opportunity_id WHERE o.id IS NULL
) THEN 1 ELSE 0 END;
SELECT 1 / CASE WHEN NOT EXISTS (
  SELECT 1 FROM opportunity_scores s LEFT JOIN opportunities o ON o.id=s.opportunity_id WHERE o.id IS NULL
) THEN 1 ELSE 0 END;
SQL

echo "restore drill verified: schema $actual_schema and all public table counts match"
