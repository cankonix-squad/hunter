#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
test_dir=$(mktemp -d)
rendered_file="$test_dir/rendered.yaml"
trap 'rm -r "$test_dir"' EXIT INT TERM

printf '%s\n' 'test-only-postgres-secret' >"$test_dir/postgres_password"
printf '%s\n' 'postgres://test:test-only-postgres-secret@postgres/test' >"$test_dir/database_url"
printf '%s\n' 'test-only-JWT-abcdefghijklmnopqrstuvwxyz-ABCDEFGHIJKLMNOPQRSTUVWXYZ-0123456789' >"$test_dir/jwt_secret"
openssl req -x509 -newkey rsa:2048 -sha256 -nodes -days 1 \
  -subj '/CN=staging.example.test' \
  -addext 'subjectAltName=DNS:staging.example.test' \
  -keyout "$test_dir/tls.key" -out "$test_dir/tls.crt" >/dev/null 2>&1

CANKONIX_PUBLIC_HOST=staging.example.test \
CANKONIX_POSTGRES_PASSWORD_SECRET_FILE="$test_dir/postgres_password" \
CANKONIX_DATABASE_URL_SECRET_FILE="$test_dir/database_url" \
CANKONIX_JWT_SECRET_FILE="$test_dir/jwt_secret" \
CANKONIX_TLS_CERTIFICATE_FILE="$test_dir/tls.crt" \
CANKONIX_TLS_PRIVATE_KEY_FILE="$test_dir/tls.key" \
docker compose -f "$repo_dir/compose.yaml" -f "$repo_dir/compose.production.yaml" \
  -f "$repo_dir/compose.staging.yaml" config >"$rendered_file"

if grep -Eq 'change-me-for-local-development|local-development-secret-change-me-now' "$rendered_file"; then
  echo "staging Compose output contains a development secret" >&2
  exit 1
fi
grep -q 'WEB_ORIGIN: https://staging.example.test' "$rendered_file"
grep -q 'published: "443"' "$rendered_file"

docker run --rm \
  -v "$repo_dir/nginx/staging.conf:/etc/nginx/conf.d/default.conf:ro" \
  -v "$repo_dir/nginx/proxy_params.conf:/etc/nginx/proxy_params.conf:ro" \
  -v "$test_dir/tls.crt:/run/secrets/tls_certificate:ro" \
  -v "$test_dir/tls.key:/run/secrets/tls_private_key:ro" \
  nginx:1.29-alpine nginx -t

echo "staging TLS configuration: ok"
