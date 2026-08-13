#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
rendered_file=$(mktemp)
secret_dir=$(mktemp -d)
trap 'rm -f "$rendered_file"; rm -r "$secret_dir"' EXIT INT TERM

digest=sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
for name in postgres_password database_url jwt_secret; do
  printf '%s\n' "test-only-$name" >"$secret_dir/$name"
done

API_IMAGE_REF="ghcr.io/cankonix-squad/hunter-api@$digest" \
COLLECTOR_IMAGE_REF="ghcr.io/cankonix-squad/hunter-collector@$digest" \
NOTIFIER_IMAGE_REF="ghcr.io/cankonix-squad/hunter-notifier@$digest" \
WEB_IMAGE_REF="ghcr.io/cankonix-squad/hunter-web@$digest" \
CANKONIX_PUBLIC_HOST=hunter.cankonix.com \
CANKONIX_TRAEFIK_NETWORK=traefik-public \
CANKONIX_TRAEFIK_CERT_RESOLVER=letsencrypt \
CANKONIX_POSTGRES_PASSWORD_SECRET_FILE="$secret_dir/postgres_password" \
CANKONIX_DATABASE_URL_SECRET_FILE="$secret_dir/database_url" \
CANKONIX_JWT_SECRET_FILE="$secret_dir/jwt_secret" \
docker compose -f "$repo_dir/compose.yaml" -f "$repo_dir/compose.production.yaml" \
  -f "$repo_dir/compose.traefik.yaml" -f "$repo_dir/compose.release.yaml" config >"$rendered_file"

for service in api collector notifier web; do
  grep -q "image: ghcr.io/cankonix-squad/hunter-$service@$digest" "$rendered_file"
done
if grep -q 'build:' "$rendered_file"; then
  echo "release configuration must not contain build directives" >&2
  exit 1
fi
test "$(grep -c 'pull_policy: always' "$rendered_file")" -eq 4
echo "immutable release image configuration: ok"
