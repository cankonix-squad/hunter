#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
test_dir=$(mktemp -d)
rendered_file="$test_dir/rendered.yaml"
trap 'rm -r "$test_dir"' EXIT INT TERM

printf '%s\n' 'test-only-postgres-secret' >"$test_dir/postgres_password"
printf '%s\n' 'postgres://test:test-only-postgres-secret@postgres/test' >"$test_dir/database_url"
printf '%s\n' 'test-only-JWT-abcdefghijklmnopqrstuvwxyz-ABCDEFGHIJKLMNOPQRSTUVWXYZ-0123456789' >"$test_dir/jwt_secret"

CANKONIX_PUBLIC_HOST=hunter.cankonix.com \
CANKONIX_TRAEFIK_NETWORK=traefik-public \
CANKONIX_TRAEFIK_CERT_RESOLVER=letsencrypt \
CANKONIX_POSTGRES_PASSWORD_SECRET_FILE="$test_dir/postgres_password" \
CANKONIX_DATABASE_URL_SECRET_FILE="$test_dir/database_url" \
CANKONIX_JWT_SECRET_FILE="$test_dir/jwt_secret" \
docker compose -f "$repo_dir/compose.yaml" -f "$repo_dir/compose.production.yaml" \
  -f "$repo_dir/compose.traefik.yaml" config >"$rendered_file"

grep -q 'WEB_ORIGIN: https://hunter.cankonix.com' "$rendered_file"
grep -q 'traefik.http.routers.cankonix.rule: Host(`hunter.cankonix.com`)' "$rendered_file"
grep -q 'traefik.http.routers.cankonix.tls.certresolver: letsencrypt' "$rendered_file"
grep -q 'traefik.http.routers.cankonix-login.priority: "100"' "$rendered_file"
grep -q 'traefik.http.middlewares.cankonix-api-rate.ratelimit.average: "20"' "$rendered_file"
grep -q 'source: .*nginx/traefik.conf' "$rendered_file"
grep -q 'name: traefik-public' "$rendered_file"
if grep -q 'published: "3000"' "$rendered_file"; then
  echo "Traefik overlay must not publish the local edge port" >&2
  exit 1
fi
if grep -Eq 'change-me-for-local-development|local-development-secret-change-me-now' "$rendered_file"; then
  echo "Traefik Compose output contains a development secret" >&2
  exit 1
fi

docker run --rm \
  -v "$repo_dir/nginx/traefik.conf:/etc/nginx/conf.d/default.conf:ro" \
  -v "$repo_dir/nginx/proxy_params.conf:/etc/nginx/proxy_params.conf:ro" \
  nginx:1.29-alpine nginx -t

echo "Traefik production configuration: ok"
