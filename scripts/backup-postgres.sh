#!/bin/sh
set -eu

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is required" >&2
  exit 1
fi
if [ "${BACKUP_WRITERS_QUIESCED:-}" != "true" ]; then
  echo "BACKUP_WRITERS_QUIESCED=true is required so the dump and count manifest share a stable dataset" >&2
  exit 1
fi
if [ "$#" -ne 1 ]; then
  echo "usage: $0 BACKUP_DIRECTORY" >&2
  exit 1
fi
for command_name in pg_dump psql; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "$command_name is required" >&2
    exit 1
  fi
done

backup_dir=$1
mkdir -p "$backup_dir"
timestamp=$(date -u +%Y%m%dT%H%M%SZ)
prefix="$backup_dir/cankonix-oi-$timestamp"
dump_tmp="$prefix.dump.tmp"
dump_file="$prefix.dump"
manifest_tmp="$prefix.manifest.tmp"
manifest_file="$prefix.manifest"
checksum_tmp="$prefix.sha256.tmp"
checksum_file="$prefix.sha256"

cleanup() {
  rm -f "$dump_tmp" "$manifest_tmp" "$checksum_tmp"
}
trap cleanup EXIT HUP INT TERM

pg_dump "$DATABASE_URL" --format=custom --compress=9 --no-owner --no-acl --file="$dump_tmp"

schema_state=$(psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -Atc \
  "SELECT count(*)::text || ':' || max(version)::text FROM schema_migrations")
table_counts=$(psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -Atc \
  "SELECT coalesce(jsonb_object_agg(tablename,row_count ORDER BY tablename),'{}'::jsonb)::text
   FROM (SELECT tablename,(xpath('/row/count/text()',query_to_xml(format('SELECT count(*) AS count FROM %I.%I',schemaname,tablename),false,true,'')))[1]::text::bigint AS row_count
         FROM pg_tables WHERE schemaname='public') counts")

{
  printf 'created_at_utc=%s\n' "$timestamp"
  printf 'schema_state=%s\n' "$schema_state"
  printf 'table_counts=%s\n' "$table_counts"
} >"$manifest_tmp"

if command -v sha256sum >/dev/null 2>&1; then
  checksum=$(sha256sum "$dump_tmp" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
  checksum=$(shasum -a 256 "$dump_tmp" | awk '{print $1}')
else
  echo "sha256sum or shasum is required" >&2
  exit 1
fi
printf '%s  %s\n' "$checksum" "$(basename "$dump_file")" >"$checksum_tmp"

mv "$manifest_tmp" "$manifest_file"
mv "$checksum_tmp" "$checksum_file"
mv "$dump_tmp" "$dump_file"
trap - EXIT HUP INT TERM

echo "$dump_file"
