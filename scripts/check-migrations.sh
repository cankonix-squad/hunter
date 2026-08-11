#!/bin/sh
set -eu

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is required" >&2
  exit 1
fi

if ! command -v psql >/dev/null 2>&1; then
  echo "psql is required" >&2
  exit 1
fi

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_dir=$(dirname "$script_dir")
migration_dir="$repo_dir/migrations"
expected_version=0

for migration in "$migration_dir"/[0-9][0-9][0-9]_*.sql; do
  version_text=$(basename "$migration" | cut -d_ -f1)
  version=$(printf '%s' "$version_text" | sed 's/^0*//')
  version=${version:-0}
  expected_version=$((expected_version + 1))

  if [ "$version" -ne "$expected_version" ]; then
    echo "migration sequence error: expected $expected_version, found $version_text" >&2
    exit 1
  fi

  psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -f "$migration" >/dev/null
done

if [ "$expected_version" -eq 0 ]; then
  echo "no migrations found" >&2
  exit 1
fi

applied=$(psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -Atc \
  'SELECT count(*)::text || '"'"':'"'"' || max(version)::text FROM schema_migrations')
expected="$expected_version:$expected_version"

if [ "$applied" != "$expected" ]; then
  echo "migration verification failed: expected $expected, found $applied" >&2
  exit 1
fi

echo "migration chain verified through version $expected_version"
