#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
secret_dir=$(mktemp -d)
rendered_file="$secret_dir/rendered.yaml"
trap 'rm -rf "$secret_dir"' EXIT INT TERM

touch "$secret_dir/postgres_password" "$secret_dir/database_url" "$secret_dir/jwt_secret"

DATABASE_URL=must-not-appear-database \
JWT_SECRET=must-not-appear-jwt \
POSTGRES_PASSWORD=must-not-appear-postgres \
CANKONIX_PUBLIC_HOST=hunter.cankonix.com \
CANKONIX_POSTGRES_PASSWORD_SECRET_FILE="$secret_dir/postgres_password" \
CANKONIX_DATABASE_URL_SECRET_FILE="$secret_dir/database_url" \
CANKONIX_JWT_SECRET_FILE="$secret_dir/jwt_secret" \
docker compose -f "$repo_dir/compose.yaml" -f "$repo_dir/compose.production.yaml" config >"$rendered_file"

if grep -Eq 'must-not-appear|change-me-for-local-development|local-development-secret-change-me-now' "$rendered_file"; then
  echo "production Compose output contains a direct or development secret" >&2
  exit 1
fi

grep -q 'DATABASE_URL_FILE: /run/secrets/database_url' "$rendered_file"
grep -q 'JWT_SECRET_FILE: /run/secrets/jwt_secret' "$rendered_file"
grep -q 'POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password' "$rendered_file"
test "$(grep -c 'WEB_ORIGIN: https://hunter.cankonix.com' "$rendered_file")" -eq 3
echo "production secret boundary: ok"
