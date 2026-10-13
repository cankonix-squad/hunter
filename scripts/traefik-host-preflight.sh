#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)

require_value() {
  key=$1
  eval "value=\${$key:-}"
  test -n "$value" || { echo "FAIL: $key is required" >&2; exit 1; }
}

for key in CANKONIX_PUBLIC_HOST CANKONIX_TRAEFIK_NETWORK CANKONIX_TRAEFIK_CERT_RESOLVER \
  CANKONIX_POSTGRES_PASSWORD_SECRET_FILE CANKONIX_DATABASE_URL_SECRET_FILE CANKONIX_JWT_SECRET_FILE \
  API_IMAGE_REF COLLECTOR_IMAGE_REF NOTIFIER_IMAGE_REF WEB_IMAGE_REF; do
  require_value "$key"
done

for key in API_IMAGE_REF COLLECTOR_IMAGE_REF NOTIFIER_IMAGE_REF WEB_IMAGE_REF; do
  eval "image_ref=\${$key}"
  if ! printf '%s\n' "$image_ref" | grep -Eq '^[^[:space:]@]+@sha256:[0-9a-f]{64}$'; then
    echo "FAIL: $key must be an immutable image@sha256 reference" >&2
    exit 1
  fi
done

test "$CANKONIX_PUBLIC_HOST" = hunter.cankonix.com || {
  echo "FAIL: CANKONIX_PUBLIC_HOST must be hunter.cankonix.com" >&2
  exit 1
}

for key in CANKONIX_POSTGRES_PASSWORD_SECRET_FILE CANKONIX_DATABASE_URL_SECRET_FILE CANKONIX_JWT_SECRET_FILE; do
  eval "secret_file=\${$key}"
  test -f "$secret_file" && test -s "$secret_file" || {
    echo "FAIL: $key must reference a non-empty regular file" >&2
    exit 1
  }
  mode=$(stat -c '%a' "$secret_file" 2>/dev/null || stat -f '%Lp' "$secret_file")
  case "$mode" in
    400|440|600|640) ;;
    *) echo "FAIL: $key permissions must be owner/group read-only (found $mode)" >&2; exit 1 ;;
  esac
done

docker network inspect "$CANKONIX_TRAEFIK_NETWORK" >/dev/null 2>&1 || {
  echo "FAIL: external Docker network $CANKONIX_TRAEFIK_NETWORK does not exist" >&2
  exit 1
}

rendered_file=$(mktemp)
trap 'rm -f "$rendered_file"' EXIT INT TERM
docker compose -f "$repo_dir/compose.yaml" -f "$repo_dir/compose.production.yaml" \
  -f "$repo_dir/compose.traefik.yaml" -f "$repo_dir/compose.release.yaml" config >"$rendered_file"

grep -q "WEB_ORIGIN: https://hunter.cankonix.com" "$rendered_file"
grep -q "name: $CANKONIX_TRAEFIK_NETWORK" "$rendered_file"
if grep -q 'published:' "$rendered_file"; then
  echo "FAIL: production Traefik configuration must not publish host ports" >&2
  exit 1
fi
if grep -q 'build:' "$rendered_file"; then
  echo "FAIL: production release configuration must not build images on the host" >&2
  exit 1
fi
if grep -Eq 'change-me-for-local-development|local-development-secret-change-me-now' "$rendered_file"; then
  echo "FAIL: rendered configuration contains development credentials" >&2
  exit 1
fi

echo "Traefik host preflight: PASS"
echo "host=$CANKONIX_PUBLIC_HOST"
echo "network=$CANKONIX_TRAEFIK_NETWORK"
echo "entrypoint=${CANKONIX_TRAEFIK_ENTRYPOINT:-websecure}"
echo "certificate_resolver=$CANKONIX_TRAEFIK_CERT_RESOLVER"
